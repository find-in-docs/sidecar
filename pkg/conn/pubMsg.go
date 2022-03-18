package conn

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	pb "github.com/samirgadkari/sidecar/protos/v1/messages"
)

type Pubs struct {
	msgId     uint32
	natsConn  *Conn
	regParams *pb.RegistrationParams
}

func InitPubs(natsConn *Conn, srv *Server) {

	srv.Pubs = &Pubs{
		msgId:     1,
		natsConn:  natsConn,
		regParams: srv.regParams, // record registration parameters
	}
}

func (pubs *Pubs) Publish(in *pb.PubMsg) (*pb.PubMsgResponse, error) {

	topic := in.GetTopic()
	data := in.GetMsg()

	fmt.Printf(">>>>> pb.PubMsg: %s\n", in)

	pubs.natsConn.Publish(topic, data)

	pubMsgRsp := pb.PubMsgResponse{
		Header: &pb.Header{
			MsgType:     pb.MsgType_MSG_TYPE_PUB_RSP,
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

	return &pubMsgRsp, nil
}

type Effector func(context.Context, *nats.Msg) (*nats.Msg, error)

func (pubs *Pubs) Retry(effector Effector, retries int, delay time.Duration) Effector {

	return func(ctx context.Context, m *nats.Msg) (*nats.Msg, error) {

		for r := 0; ; r++ {

			response, err := effector(ctx, m)
			if err == nil || r >= retries {
				return response, err
			}

			log.Printf("Attempt %d failed; retrying in %v", r+1, delay)

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}
}
