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
	natsMsgs map[string]chan *nats.Msg
	done     chan struct{}
	msgId    uint32
	natsConn *Conn
	header   *messages.Header
}

func InitSubs(natsConn *Conn, srv *Server) {

	natsMsgs := make(map[string]chan *nats.Msg)
	done := make(chan struct{})

	srv.Subs = &Subs{

		natsMsgs: natsMsgs,
		done:     done,
		msgId:    1,
		natsConn: natsConn,
	}
}

func (subs *Subs) Subscribe(in *messages.SubMsg) (*messages.SubMsgResponse, error) {

	fmt.Printf("Received SubMsg: %v\n", in)
	topic := in.GetTopic()
	chanSize := in.GetChanSize()

	topicMsgs := make(chan *nats.Msg, chanSize)
	subs.natsMsgs[topic] = topicMsgs

	subs.natsConn.Subscribe(topic, func(m *nats.Msg) {
		fmt.Printf("Received msg from NATS: %v\n", m)
		subs.natsMsgs[topic] <- m
	})

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

	m := <-srv.Subs.natsMsgs[in.Topic]
	fmt.Printf("Got msg from NATS server:\n\t%#v\n", m)

	header := srv.Subs.header
	header.MsgType = messages.MsgType_MSG_TYPE_SUB_TOPIC_RSP

	return &messages.SubTopicResponse{
		Header: header,
		Topic:  m.Subject,
		Msg:    m.Data,
	}, nil
}
