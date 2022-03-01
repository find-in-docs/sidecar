package conn

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/samirgadkari/sidecar/protos/v1/messages"
)

const (
	msgChSize = 10
)

type Subs struct {
	natsMsgs      map[string]chan *nats.Msg
	subscriptions map[string]*nats.Subscription
	msgId         uint32
	natsConn      *Conn
	header        *messages.Header
}

func InitSubs(natsConn *Conn, srv *Server) {

	natsMsgs := make(map[string]chan *nats.Msg)
	subscriptions := make(map[string]*nats.Subscription)

	srv.Subs = &Subs{

		natsMsgs:      natsMsgs,
		subscriptions: subscriptions,
		msgId:         1,
		natsConn:      natsConn,
	}
}

func (subs *Subs) Subscribe(in *messages.SubMsg) (*messages.SubMsgResponse, error) {

	fmt.Printf("Received SubMsg: %v\n", in)
	topic := in.GetTopic()
	chanSize := in.GetChanSize()

	topicMsgs := make(chan *nats.Msg, chanSize)
	subs.natsMsgs[topic] = topicMsgs

	subscription, err := subs.natsConn.Subscribe(topic, func(m *nats.Msg) {
		fmt.Printf("Received msg from NATS: %v\n", m)
		subs.natsMsgs[topic] <- m
	})
	if err != nil {
		return nil, err
	}
	subs.subscriptions[topic] = subscription

	srcHeader := in.GetHeader()
	header := messages.Header{
		MsgType:     messages.MsgType_MSG_TYPE_SUB_RSP,
		DstServType: srcHeader.GetSrcServType(),
		SrcServType: srcHeader.GetDstServType(),
		ServId:      srcHeader.GetServId(),
		MsgId:       0,
	}
	subs.header = &header

	rspHeader := messages.ResponseHeader{

		Status: uint32(messages.Status_OK),
	}

	subMsgRsp := messages.SubMsgResponse{

		Header:    &header,
		RspHeader: &rspHeader,
		Msg:       "OK",
	}

	return &subMsgRsp, nil
}

func RecvFromNATS(srv *Server, in *messages.Receive) (*messages.SubTopicResponse, error) {

	natsMsgs, ok := srv.Subs.natsMsgs[in.Topic]
	fmt.Printf("ok: %v\nnatsMsgs: %v\n", ok, natsMsgs)
	if !ok {
		return nil, fmt.Errorf("Warning - already unsubscribed from topic: %s\n", in.Topic)
	}

	m := <-natsMsgs
	if m == nil {
		return nil, fmt.Errorf("Warning - already unsubscribed from topic: %s\n", in.Topic)
	}

	fmt.Printf("Got msg from NATS server:\n\t%#v\n\ttopic:%s", m, in.Topic)

	header := srv.Subs.header
	header.MsgType = messages.MsgType_MSG_TYPE_SUB_TOPIC_RSP

	return &messages.SubTopicResponse{
		Header: header,
		Topic:  m.Subject,
		Msg:    m.Data,
	}, nil
}

func (subs *Subs) Unsubscribe(in *messages.UnsubMsg) (*messages.UnsubMsgResponse, error) {

	fmt.Printf("Received UnsubMsg: %v\n", in)
	topic := in.GetTopic()

	if _, ok := subs.subscriptions[topic]; !ok {
		fmt.Printf("Error - topic not found to unsubscribe:\n\ttopic: %s\n", topic)
		return nil, fmt.Errorf("Topic not found to unsubscribe: %s\n", topic)
	}

	subs.subscriptions[topic].Drain()
	subs.subscriptions[topic].Unsubscribe()
	delete(subs.subscriptions, topic)

	close(subs.natsMsgs[topic])
	delete(subs.natsMsgs, topic)

	fmt.Printf("Successfully unsubscribed from topic: %s\n", topic)

	srcHeader := in.GetHeader()
	header := messages.Header{
		MsgType:     messages.MsgType_MSG_TYPE_UNSUB_RSP,
		DstServType: srcHeader.GetSrcServType(),
		SrcServType: srcHeader.GetDstServType(),
		ServId:      srcHeader.GetServId(),
		MsgId:       0,
	}
	subs.header = &header

	rspHeader := messages.ResponseHeader{

		Status: uint32(messages.Status_OK),
	}

	unsubMsgRsp := messages.UnsubMsgResponse{

		Header:    &header,
		RspHeader: &rspHeader,
		Msg:       "OK",
	}

	return &unsubMsgRsp, nil
}
