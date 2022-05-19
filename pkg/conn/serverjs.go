package conn

import (
	"context"

	pb "github.com/find-in-docs/sidecar/protos/v1/messages"
)

func (s *Server) SubJS(ctx context.Context, in *pb.SubJSMsg) (*pb.SubJSMsgResponse, error) {

	in.Header.MsgId = NextMsgId()
	s.Logs.logger.Log("Received SubJSMsg: %s\n", in)

	m, err := s.Subs.SubscribeJS(in)
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

func (s *Server) RecvJS(ctx context.Context, in *pb.ReceiveJS) (*pb.SubJSTopicResponse, error) {

	in.Header.MsgId = NextMsgId()
	// Do not log message to NATS. This creates a loop.
	s.Logs.logger.PrintMsg("Received from NATS: %s\n", in)

	m, err := RecvJSFromNATS(ctx, s, in)
	if err != nil {
		s.Logs.logger.Log("Could not receive from NATS: %s\n", err.Error())
		return nil, err
	}
	m.Header.MsgId = NextMsgId()
	s.Logs.logger.PrintMsg("Converted NATS message to GRPC message\n")

	return m, err
}
