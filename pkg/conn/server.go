package conn

import (
	"context"

	pb "github.com/samirgadkari/sidecar/protos/v1/messages"
	"google.golang.org/grpc"
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

	s.Logs.logger.PrintMsg("Received RegistrationMsg: %s\n", in)

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

	s.Logs.logger.PrintMsg("Sending regRsp: %s\n", regRsp)

	return regRsp, nil
}

func (s *Server) Log(ctx context.Context, in *pb.LogMsg) (*pb.LogMsgResponse, error) {

	in.Header.MsgId = NextMsgId()
	s.Logs.logger.PrintMsg("Received LogMsg: %s\n", in)

	m, err := s.Logs.ReceivedLogMsg(in)
	if err == nil {
		m.Header.MsgId = NextMsgId()
		s.Logs.logger.PrintMsg("Sending LogMsgResponse: %s\n", m)
	}

	return m, err
}

func (s *Server) Sub(ctx context.Context, in *pb.SubMsg) (*pb.SubMsgResponse, error) {

	in.Header.MsgId = NextMsgId()
	s.Logs.logger.PrintMsg("Received SubMsg: %s\n", in)

	m, err := s.Subs.Subscribe(in)
	if err == nil {
		m.Header.MsgId = NextMsgId()
		s.Logs.logger.PrintMsg("Sending SubMsgRsp: %s\n", m)
	}

	return m, err
}

func (s *Server) Unsub(ctx context.Context, in *pb.UnsubMsg) (*pb.UnsubMsgResponse, error) {

	in.Header.MsgId = NextMsgId()
	s.Logs.logger.PrintMsg("Received UnsubMsg: %s\n", in)

	m, err := s.Subs.Unsubscribe(in)
	if err == nil {
		m.Header.MsgId = NextMsgId()
		s.Logs.logger.PrintMsg("Sending UnsubMsgRsp: %s\n", m)
	}

	return m, err
}

func (s *Server) Recv(ctx context.Context, in *pb.Receive) (*pb.SubTopicResponse, error) {

	in.Header.MsgId = NextMsgId()
	// Do not log message to NATS. This creates a loop.
	s.Logs.logger.PrintMsg("Received Receive: %s\n", in)

	m, err := RecvFromNATS(s, in)
	if err == nil {
		m.Header.MsgId = NextMsgId()
	}

	return m, err
}

func (s *Server) Pub(ctx context.Context, in *pb.PubMsg) (*pb.PubMsgResponse, error) {

	in.Header.MsgId = NextMsgId()
	s.Logs.logger.PrintMsg("Received PubMsg: %s\n", in)

	var retryBehavior *pb.RetryBehavior
	if in.Retry != nil {
		retryBehavior = in.Retry
	} else {
		retryBehavior = s.Pubs.regParams.Retry
	}

	m, err := s.Pubs.Publish(ctx, in, retryBehavior)
	if err == nil {
		s.Logs.logger.PrintMsg("Sending PubMsgResponse: %s\n", m)
	}

	return m, err
}
