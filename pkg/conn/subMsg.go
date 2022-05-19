package conn

import (
	"context"
	"fmt"

	"github.com/find-in-docs/sidecar/pkg/log"
	pb "github.com/find-in-docs/sidecar/protos/v1/messages"
	"github.com/nats-io/nats.go"
)

const (
	msgChSize = 10
)

type Subs struct {
	natsMsgs        map[string]chan *nats.Msg
	natsJSMsgs      map[string]chan *nats.Msg
	subscriptions   map[string]*nats.Subscription
	subscriptionsJS map[string]*nats.Subscription
	msgId           uint32
	natsConn        *Conn
	header          *pb.Header
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

func (subs *Subs) Subscribe(in *pb.SubMsg) (*pb.SubMsgResponse, error) {

	topic := in.GetTopic()
	chanSize := in.GetChanSize()

	topicMsgs := make(chan *nats.Msg, chanSize)
	subs.natsMsgs[topic] = topicMsgs

	subscription, err := subs.natsConn.Subscribe(topic, func(m *nats.Msg) {
		subs.natsMsgs[topic] <- m
	})
	if err != nil {
		return nil, err
	}
	subs.subscriptions[topic] = subscription

	subMsgRsp := &pb.SubMsgResponse{
		Header: &pb.Header{
			MsgType:     pb.MsgType_MSG_TYPE_SUB_RSP,
			SrcServType: "sidecar",
			DstServType: in.Header.SrcServType,
			ServId:      serviceId()(),
			MsgId:       NextMsgId(),
		},

		RspHeader: &pb.ResponseHeader{
			Status: uint32(pb.Status_OK),
		},

		Msg: "OK",
	}

	return subMsgRsp, nil
}

func (subs *Subs) SubscribeJS(in *pb.SubJSMsg) (*pb.SubJSMsgResponse, error) {

	topic := in.GetTopic()
	workQueue := in.GetWorkQueue()
	chanSize := in.GetChanSize()

	topicMsgs := make(chan *nats.Msg, chanSize)
	subs.natsJSMsgs[topic] = topicMsgs

	subscription, err := subs.natsConn.js.PullSubscribe(topic, workQueue,
		nats.PullMaxWaiting(128))
	if err != nil {
		return nil, err
	}
	subs.subscriptionsJS[topic] = subscription

	subJSMsgRsp := &pb.SubJSMsgResponse{
		Header: &pb.Header{
			MsgType:     pb.MsgType_MSG_TYPE_SUB_RSP,
			SrcServType: "sidecar",
			DstServType: in.Header.SrcServType,
			ServId:      serviceId()(),
			MsgId:       NextMsgId(),
		},

		RspHeader: &pb.ResponseHeader{
			Status: uint32(pb.Status_OK),
		},

		Msg: "OK",
	}

	return subJSMsgRsp, nil
}

func RecvFromNATS(ctx context.Context, srv *Server, in *pb.Receive) (*pb.SubTopicResponse, error) {

	natsMsgs, ok := srv.Subs.natsMsgs[in.Topic]
	if !ok {
		return nil, fmt.Errorf("Warning - already unsubscribed from topic: %s\n", in.Topic)
	}

	select {

	case m := <-natsMsgs:
		if m == nil {
			return nil, fmt.Errorf("Warning - already unsubscribed from topic: %s\n", in.Topic)
		}

		s := fmt.Sprintf("Header:%s\n\tTopic: %s\n",
			in.Header, in.Topic)
		srv.Logs.logger.PrintMsg("Got msg from NATS server: %s\n", s)

		subTopicRsp := &pb.SubTopicResponse{
			Header: &pb.Header{
				MsgType:     pb.MsgType_MSG_TYPE_SUB_TOPIC_RSP,
				SrcServType: serviceType(),
				DstServType: in.Header.SrcServType,
				ServId:      serviceId()(),
				MsgId:       NextMsgId(),
			},

			Topic: m.Subject,

			Msg: m.Data,
		}

		return subTopicRsp, nil

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func RecvJSFromNATS(ctx context.Context, srv *Server, in *pb.ReceiveJS) (*pb.SubJSTopicResponse, error) {

	natsJSMsgs, ok := srv.Subs.natsJSMsgs[in.Topic]
	if !ok {
		return nil, fmt.Errorf("Warning - already unsubscribed from topic: %s\n", in.Topic)
	}

	select {

	case m := <-natsJSMsgs:
		if m == nil {
			return nil, fmt.Errorf("Warning - already unsubscribed from topic: %s\n", in.Topic)
		}

		s := fmt.Sprintf("Header:%s\n\tTopic: %s\n\tWorkQueue: %s\n",
			in.Header, in.Topic, in.WorkQueue)
		srv.Logs.logger.PrintMsg("Got msg from NATS server: %s\n", s)

		subJSTopicRsp := &pb.SubJSTopicResponse{
			Header: &pb.Header{
				MsgType:     pb.MsgType_MSG_TYPE_SUB_JS_TOPIC_RSP,
				SrcServType: serviceType(),
				DstServType: in.Header.SrcServType,
				ServId:      serviceId()(),
				MsgId:       NextMsgId(),
			},

			Topic: m.Subject,
			Msg:   m.Data,
		}

		return subJSTopicRsp, nil

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (subs *Subs) Unsubscribe(logger *log.Logger, in *pb.UnsubMsg) (*pb.UnsubMsgResponse, error) {

	topic := in.GetTopic()

	if _, ok := subs.subscriptions[topic]; !ok {
		return nil, fmt.Errorf("Error - topic not found to unsubscribe:\n\ttopic: %s\n", topic)
	}

	subs.subscriptions[topic].Drain()
	subs.subscriptions[topic].Unsubscribe()
	delete(subs.subscriptions, topic)

	close(subs.natsMsgs[topic])
	delete(subs.natsMsgs, topic)

	unsubMsgRsp := &pb.UnsubMsgResponse{
		Header: &pb.Header{
			MsgType:     pb.MsgType_MSG_TYPE_UNSUB_RSP,
			SrcServType: serviceType(),
			DstServType: in.Header.SrcServType,
			ServId:      serviceId()(),
			MsgId:       NextMsgId(),
		},

		RspHeader: &pb.ResponseHeader{
			Status: uint32(pb.Status_OK),
		},

		Msg: "OK",
	}

	logger.Log("Successfully unsubscribed from topic: %s\n", topic)

	return unsubMsgRsp, nil
}
