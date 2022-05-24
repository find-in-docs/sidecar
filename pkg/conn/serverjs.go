package conn

import (
	"context"
	"fmt"

	pb "github.com/find-in-docs/sidecar/protos/v1/messages"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) SubJS(ctx context.Context, in *pb.SubJSMsg) (*pb.SubJSMsgResponse, error) {

	in.Header.MsgId = NextMsgId()
	s.Logs.logger.Log("Received SubJSMsg: %s\n", in)

	m, err := s.Subs.SubscribeJS(ctx, in)
	if err != nil {
		s.Logs.logger.Log("Error subscribing: %s\n", err.Error())
		return nil, err
	}
	m.Header.MsgId = NextMsgId()
	s.Logs.logger.Log("Sending SubJSMsgRsp: %s\n", m)

	return m, err
}

func (s *Server) UnsubJS(ctx context.Context, in *pb.UnsubJSMsg) (*pb.UnsubJSMsgResponse, error) {

	in.Header.MsgId = NextMsgId()
	s.Logs.logger.Log("Received UnsubJSMsg: %s\n", in)

	m, err := s.Subs.UnsubscribeJS(s.Logs.logger, in)
	if err != nil {
		s.Logs.logger.Log("Error unsubscribing: %s\n", err.Error())
		return nil, err
	}
	m.Header.MsgId = NextMsgId()
	s.Logs.logger.Log("Sending UnsubMsgRsp: %s\n", m)

	return m, err
}

func emptySubJSTopicResponse(header *pb.Header, topic, wq string) *pb.SubJSTopicResponse {
	return &pb.SubJSTopicResponse{

		Header: &pb.Header{
			MsgType:     pb.MsgType_MSG_TYPE_SUB_JS_TOPIC_RSP,
			SrcServType: header.DstServType,
			DstServType: header.SrcServType,
			ServId:      header.ServId,
			MsgId:       NextMsgId(),
		},

		Topic:     topic,
		WorkQueue: wq,
		Msg:       []byte(""),
	}
}

func (s *Server) RecvJS(ctx context.Context, in *pb.ReceiveJS) (*pb.SubJSTopicResponse, error) {

	in.Header.MsgId = NextMsgId()
	// Do not log message to NATS. This creates a loop.
	s.Logs.logger.PrintMsg("Received from NATS: %s\n", in)

	m, err := RecvJSFromNATS(ctx, s, in)
	if err != nil {
		s.Logs.logger.Log("Could not receive from NATS: %s\n", err.Error())
		return nil, err
	}
	if m == nil {
		fmt.Printf("RecvJSFromNATS returned nil m\n")
		return emptySubJSTopicResponse(in.Header, in.Topic, in.WorkQueue), err
	}
	m.Header.MsgId = NextMsgId()
	s.Logs.logger.PrintMsg("Converted NATS message to GRPC message\n")

	return m, err
}

func (s *Server) PubJS(ctx context.Context, in *pb.PubJSMsg) (*emptypb.Empty, error) {

	in.Header.MsgId = NextMsgId()
	s.Logs.logger.Log("Received PubMsg: %s\n", in)

	future, err := s.Pubs.natsConn.js.PublishAsync(in.Topic, in.Msg)
	if err != nil {
		return &emptypb.Empty{}, fmt.Errorf("Error publishing to JetStream with topic: %s\n", in.Topic)
	}
	s.Logs.logger.Log("Published to JetStream - future returned: %v\n", future)

	return &emptypb.Empty{}, err
}
