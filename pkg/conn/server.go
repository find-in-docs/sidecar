package conn

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/samirgadkari/sidecar/protos/v1/messages"
)

type Server struct {
	messages.UnimplementedSidecarServer

	Logs *Logs
	Pubs *Pubs
	Subs *Subs

	Header *messages.Header
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

func (s *Server) Register(ctx context.Context, in *messages.RegistrationMsg) (*messages.RegistrationMsgResponse, error) {
	fmt.Printf("Received RegistrationMsg: %v\n", in)

	rspHeader := messages.ResponseHeader{
		Status: uint32(messages.Status_OK),
	}

	header := messages.Header{
		MsgType:     messages.MsgType_MSG_TYPE_REG_RSP,
		SrcServType: "sidecarService",
		DstServType: in.Header.SrcServType,
		ServId:      servId(),
		MsgId:       0,
	}
	s.Header = &header

	regRsp := &messages.RegistrationMsgResponse{
		Header:    &header,
		RspHeader: &rspHeader,
		Msg:       "OK",
	}
	fmt.Printf("Sending regRsp: %v\n", regRsp)

	return regRsp, nil
}

func (s *Server) Log(ctx context.Context, in *messages.LogMsg) (*messages.LogMsgResponse, error) {

	return s.Logs.ReceivedLogMsg(in)
}

func (s *Server) Sub(ctx context.Context, in *messages.SubMsg) (*messages.SubMsgResponse, error) {
	return s.Subs.Subscribe(in)
}

func (s *Server) Unsub(ctx context.Context, in *messages.UnsubMsg) (*messages.UnsubMsgResponse, error) {
	return s.Subs.Unsubscribe(in)
}

func (s *Server) Recv(ctx context.Context, m *messages.Receive) (*messages.SubTopicResponse, error) {
	return RecvFromNATS(s, m)
}

func (s *Server) Pub(ctx context.Context, in *messages.PubMsg) (*messages.PubMsgResponse, error) {
	return s.Pubs.Publish(in)
}
