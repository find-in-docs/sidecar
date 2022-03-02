package conn

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	pb "github.com/samirgadkari/sidecar/protos/v1/messages"
)

// This value is used in other places than just within the server context.
// This means we have to keep it as a global.
type MsgID struct {
	mu    sync.Mutex
	Value uint64
}

func NextMsgIdCall() func() uint64 {

	var MsgId MsgID

	return func() uint64 {
		MsgId.mu.Lock()
		defer MsgId.mu.Unlock()

		MsgId.Value += 1
		return MsgId.Value
	}
}

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

func (s *Server) Register(ctx context.Context, in *pb.RegistrationMsg) (*pb.RegistrationMsgResponse, error) {
	fmt.Printf("Received RegistrationMsg: %v\n", in)

	rspHeader := pb.ResponseHeader{
		Status: uint32(pb.Status_OK),
	}

	header := pb.Header{
		MsgType:     pb.MsgType_MSG_TYPE_REG_RSP,
		SrcServType: "sidecarService",
		DstServType: in.Header.SrcServType,
		ServId:      servId(),
	}
	if header.MsgId == 0 {
		in.Header.MsgId = NextMsgIdCall()()
	}
	s.Header = &header

	regRsp := &pb.RegistrationMsgResponse{
		Header:    &header,
		RspHeader: &rspHeader,
		Msg:       "OK",
	}
	fmt.Printf("Sending regRsp: %v\n", regRsp)

	return regRsp, nil
}

func (s *Server) Log(ctx context.Context, in *pb.LogMsg) (*pb.LogMsgResponse, error) {

	if in.Header.MsgId == 0 {
		in.Header.MsgId = NextMsgIdCall()()
	}
	return s.Logs.ReceivedLogMsg(in)
}

func (s *Server) Sub(ctx context.Context, in *pb.SubMsg) (*pb.SubMsgResponse, error) {

	if in.Header.MsgId == 0 {
		in.Header.MsgId = NextMsgIdCall()()
	}
	return s.Subs.Subscribe(in)
}

func (s *Server) Unsub(ctx context.Context, in *pb.UnsubMsg) (*pb.UnsubMsgResponse, error) {

	if in.Header.MsgId == 0 {
		in.Header.MsgId = NextMsgIdCall()()
	}
	return s.Subs.Unsubscribe(in)
}

func (s *Server) Recv(ctx context.Context, in *pb.Receive) (*pb.SubTopicResponse, error) {

	if in.Header.MsgId == 0 {
		in.Header.MsgId = NextMsgIdCall()()
	}
	return RecvFromNATS(s, in)
}

func (s *Server) Pub(ctx context.Context, in *pb.PubMsg) (*pb.PubMsgResponse, error) {

	if in.Header.MsgId == 0 {
		in.Header.MsgId = NextMsgIdCall()()
	}
	return s.Pubs.Publish(in)
}
