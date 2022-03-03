package conn

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	pb "github.com/samirgadkari/sidecar/protos/v1/messages"
)

type Server struct {
	pb.UnimplementedSidecarServer

	Logs *Logs
	Pubs *Pubs
	Subs *Subs

	Header *pb.Header
}

func servId() []byte {
	uuid := uuid.New()
	servId, err := uuid.MarshalText()
	if err != nil {
		fmt.Printf("Error converting UUID: %v to text\n\terr: %v\n", servId, err)
		os.Exit(-1)
	}

	return servId
}

func PrintRegistrationMsg(prefix string, m *pb.RegistrationMsg) {

	fmt.Printf("%s", prefix)
	fmt.Printf("\n\tHeader: %s\n\tserviceName: %s",
		m.Header, m.ServiceName)
	fmt.Printf("\n\tcircuitFailureThreshold: %d\n\tdebounceDelay: %s",
		*m.CircuitFailureThreshold, m.DebounceDelay)
	fmt.Printf("\n\tretryNum: %d\n\tretryDelay: %s\n",
		*m.RetryNum, m.RetryDelay)
}

func PrintRegistrationMsgRsp(prefix string, m *pb.RegistrationMsgResponse) {

	fmt.Printf("%s\n\tHeader: %s\n\tRspHeader: %s\n\tMsg: %s\n",
		prefix, m.Header, m.RspHeader, m.Msg)
}

func (s *Server) Register(ctx context.Context, in *pb.RegistrationMsg) (*pb.RegistrationMsgResponse, error) {

	PrintRegistrationMsg("Received RegistrationMsg:", in)

	rspHeader := pb.ResponseHeader{
		Status: uint32(pb.Status_OK),
	}

	header := pb.Header{
		MsgType:     pb.MsgType_MSG_TYPE_REG_RSP,
		SrcServType: "sidecar",
		DstServType: in.Header.SrcServType,
		ServId:      servId(),
	}
	if header.MsgId == 0 {
		in.Header.MsgId = NextMsgId()
	}
	s.Header = &header

	regRsp := &pb.RegistrationMsgResponse{
		Header:    &header,
		RspHeader: &rspHeader,
		Msg:       "OK",
	}
	PrintRegistrationMsgRsp("Sending regRsp:", regRsp)

	return regRsp, nil
}

func (s *Server) Log(ctx context.Context, in *pb.LogMsg) (*pb.LogMsgResponse, error) {

	if in.Header.MsgId == 0 {
		in.Header.MsgId = NextMsgId()
	}
	return s.Logs.ReceivedLogMsg(in)
}

func (s *Server) Sub(ctx context.Context, in *pb.SubMsg) (*pb.SubMsgResponse, error) {

	if in.Header.MsgId == 0 {
		in.Header.MsgId = NextMsgId()
	}
	return s.Subs.Subscribe(in)
}

func (s *Server) Unsub(ctx context.Context, in *pb.UnsubMsg) (*pb.UnsubMsgResponse, error) {

	if in.Header.MsgId == 0 {
		in.Header.MsgId = NextMsgId()
	}
	return s.Subs.Unsubscribe(in)
}

func (s *Server) Recv(ctx context.Context, in *pb.Receive) (*pb.SubTopicResponse, error) {

	if in.Header.MsgId == 0 {
		in.Header.MsgId = NextMsgId()
	}
	return RecvFromNATS(s, in)
}

func (s *Server) Pub(ctx context.Context, in *pb.PubMsg) (*pb.PubMsgResponse, error) {

	if in.Header.MsgId == 0 {
		in.Header.MsgId = NextMsgId()
	}
	return s.Pubs.Publish(in)
}
