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

	in.Header.MsgId = NextMsgId()
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
	s.Header = &header

	regRsp := &pb.RegistrationMsgResponse{
		Header:    &header,
		RspHeader: &rspHeader,
		Msg:       "OK",
	}
	regRsp.Header.MsgId = NextMsgId()

	PrintRegistrationMsgRsp("Sending regRsp:", regRsp)

	return regRsp, nil
}

func (s *Server) Log(ctx context.Context, in *pb.LogMsg) (*pb.LogMsgResponse, error) {

	in.Header.MsgId = NextMsgId()
	PrintLogMsg("Received LogMsg:", in)

	m, err := s.Logs.ReceivedLogMsg(in)
	if err == nil {
		m.Header.MsgId = NextMsgId()
		PrintLogMsgRsp("Sending LogMsgResponse:", m)
	}

	return m, err
}

func (s *Server) Sub(ctx context.Context, in *pb.SubMsg) (*pb.SubMsgResponse, error) {

	in.Header.MsgId = NextMsgId()
	PrintSubMsg("Received SubMsg:", in)

	m, err := s.Subs.Subscribe(in)
	if err == nil {
		m.Header.MsgId = NextMsgId()
		PrintSubMsgRsp("Sending SubMsgRsp:", m)
	}

	return m, err
}

func (s *Server) Unsub(ctx context.Context, in *pb.UnsubMsg) (*pb.UnsubMsgResponse, error) {

	in.Header.MsgId = NextMsgId()
	PrintUnsubMsg("Received UnsubMsg:", in)

	m, err := s.Subs.Unsubscribe(in)
	if err == nil {
		m.Header.MsgId = NextMsgId()
		PrintUnsubMsgRsp("Sending UnsubMsgRsp:", m)
	}

	return m, err
}

func (s *Server) Recv(ctx context.Context, in *pb.Receive) (*pb.SubTopicResponse, error) {

	in.Header.MsgId = NextMsgId()
	PrintRecvMsg("Received Receive:", in)

	m, err := RecvFromNATS(s, in)
	if err == nil {
		m.Header.MsgId = NextMsgId()
		PrintSubTopicRsp("Sending SubTopicRsp:", m)
	}

	return m, err
}

func (s *Server) Pub(ctx context.Context, in *pb.PubMsg) (*pb.PubMsgResponse, error) {

	in.Header.MsgId = NextMsgId()
	PrintPubMsg("Received PubMsg:", in)

	m, err := s.Pubs.Publish(in)
	if err == nil {
		m.Header.MsgId = NextMsgId()
		PrintPubMsgRsp("Sending PubMsgResponse:", m)
	}

	return m, err
}
