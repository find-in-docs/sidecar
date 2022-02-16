package conn

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/samirgadkari/sidecar/protos/v1/messages"
)

type Server struct {
	messages.UnimplementedSidecarServer

	Logs *Logs

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

func (s *Server) Register(ctx context.Context, in *messages.RegistrationMsg) (*messages.RegistrationResponse, error) {
	fmt.Printf("Received RegistrationMsg: %v\n", in)

	rspHeader := messages.ResponseHeader{
		Status: uint32(messages.Status_OK),
	}

	header := messages.Header{
		SrcServType: "sidecarService",
		DstServType: in.Header.SrcServType,
		ServId:      servId(),
		MsgId:       0,
	}
	s.Header = &header

	regRsp := &messages.RegistrationResponse{
		Header:    &header,
		RspHeader: &rspHeader,
		Msg:       "OK",
	}
	fmt.Printf("Sending regRsp: %#v\n", *regRsp)

	return regRsp, nil
}

func (s *Server) Log(ctx context.Context, in *messages.LogMsg) (*messages.LogResponse, error) {

	return s.Logs.ReceivedLogMsg(in)
}

func (s *Server) Sub(context.Context, *messages.SubMsg) (*messages.SubResponse, error) {
	return nil, errors.New("Not implemented")
}

func (s *Server) Recv(context.Context, *messages.Empty) (*messages.RecvResponse, error) {
	return nil, errors.New("Not implemented")
}

func (s *Server) Pub(context.Context, *messages.PubMsg) (*messages.PubResponse, error) {
	return nil, errors.New("Not implemented")
}
