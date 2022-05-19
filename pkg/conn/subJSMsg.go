package conn

import (
	"context"
	"fmt"

	"github.com/find-in-docs/sidecar/pkg/log"
	pb "github.com/find-in-docs/sidecar/protos/v1/messages"
	"github.com/nats-io/nats.go"
)

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

func (subs *Subs) UnsubscribeJS(logger *log.Logger, in *pb.UnsubJSMsg) (*pb.UnsubJSMsgResponse, error) {

	topic := in.GetTopic()
	workQueue := in.GetWorkQueue()

	if _, ok := subs.subscriptionsJS[topic]; !ok {
		return nil, fmt.Errorf("Error - topic not found to unsubscribe:\n\ttopic: %s\n\tworkQueue: %s\n",
			topic, workQueue)
	}

	subs.subscriptionsJS[topic].Drain()
	subs.subscriptionsJS[topic].Unsubscribe()
	delete(subs.subscriptionsJS, topic)

	close(subs.natsJSMsgs[topic])
	delete(subs.natsJSMsgs, topic)

	unsubJSMsgRsp := &pb.UnsubJSMsgResponse{
		Header: &pb.Header{
			MsgType:     pb.MsgType_MSG_TYPE_UNSUB_JS_RSP,
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

	logger.Log("Successfully unsubscribed from topic: %s\n\tworkQueue: %s\n", topic, workQueue)

	return unsubJSMsgRsp, nil
}
