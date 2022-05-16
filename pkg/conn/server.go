package conn

import (
	"context"

	pb "github.com/find-in-docs/sidecar/protos/v1/messages"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	pb.UnimplementedSidecarServer

	GrcpServer *grpc.Server
	regParams  *pb.RegistrationParams
	Logs       *Logs
	Pubs       *Pubs
	Subs       *Subs
}

func (s *Server) Register(ctx context.Context, in *pb.RegistrationMsg) (*pb.RegistrationMsgResponse, error) {

	// Record Registration parameters for later use
	s.regParams = in.RegParams

	// Server does assignment of message IDs.
	in.Header.MsgId = NextMsgId()

	s.Logs.logger.Log("Received RegistrationMsg: %s\n", in)

	regRsp := &pb.RegistrationMsgResponse{
		Header: &pb.Header{
			MsgType:     pb.MsgType_MSG_TYPE_REG_RSP,
			SrcServType: serviceType(),
			DstServType: in.Header.SrcServType,
			ServId:      serviceId()(),
			MsgId:       NextMsgId(),
		},

		RspHeader: &pb.ResponseHeader{
			Status: uint32(pb.Status_OK),
		},

		Msg:            "OK",
		AssignedServId: createServiceId(), // assign new service ID to client
	}

	s.Logs.logger.Log("Sending regRsp: %s\n", regRsp)

	return regRsp, nil
}

func (s *Server) Log(ctx context.Context, in *pb.LogMsg) (*emptypb.Empty, error) {

	in.Header.MsgId = NextMsgId()
	s.Logs.logger.PrintMsg("Received LogMsg: %s\n", in)

	s.Logs.ReceivedLogMsg(in)

	return &emptypb.Empty{}, nil
}

func (s *Server) Sub(ctx context.Context, in *pb.SubMsg) (*pb.SubMsgResponse, error) {

	in.Header.MsgId = NextMsgId()
	s.Logs.logger.Log("Received SubMsg: %s\n", in)

	m, err := s.Subs.Subscribe(in)
	if err != nil {
		s.Logs.logger.Log("Error subscribing: %s\n", err.Error())
		return nil, err
	}
	m.Header.MsgId = NextMsgId()
	s.Logs.logger.Log("Sending SubMsgRsp: %s\n", m)

	return m, err
}

func (s *Server) Unsub(ctx context.Context, in *pb.UnsubMsg) (*pb.UnsubMsgResponse, error) {

	in.Header.MsgId = NextMsgId()
	s.Logs.logger.Log("Received UnsubMsg: %s\n", in)

	m, err := s.Subs.Unsubscribe(s.Logs.logger, in)
	if err != nil {
		s.Logs.logger.Log("Error unsubscribing: %s\n", err.Error())
		return nil, err
	}
	m.Header.MsgId = NextMsgId()
	s.Logs.logger.Log("Sending UnsubMsgRsp: %s\n", m)

	return m, err
}

func (s *Server) Recv(ctx context.Context, in *pb.Receive) (*pb.SubTopicResponse, error) {

	in.Header.MsgId = NextMsgId()
	// Do not log message to NATS. This creates a loop.
	s.Logs.logger.PrintMsg("Received from NATS: %s\n", in)

	m, err := RecvFromNATS(ctx, s, in)
	if err != nil {
		s.Logs.logger.Log("Could not receive from NATS: %s\n", err.Error())
		return nil, err
	}
	m.Header.MsgId = NextMsgId()
	s.Logs.logger.PrintMsg("Converted NATS message to GRPC message\n")

	return m, err
}

func (s *Server) Pub(ctx context.Context, in *pb.PubMsg) (*pb.PubMsgResponse, error) {

	in.Header.MsgId = NextMsgId()
	s.Logs.logger.Log("Received PubMsg: %s\n", in)

	var retryBehavior *pb.RetryBehavior
	if in.Retry != nil {
		retryBehavior = in.Retry
	} else {
		retryBehavior = s.Pubs.regParams.Retry
	}

	m, err := s.Pubs.Publish(ctx, s.Logs.logger, in, retryBehavior)
	if err == nil { // different logic - checking if err == nil instead of != nil
		s.Logs.logger.Log("Sending PubMsgResponse: %s\n", m)
	}

	return m, err
}
