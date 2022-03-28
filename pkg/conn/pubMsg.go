package conn

import (
	"context"
	"fmt"
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

type Effector func(context.Context, *nats.Msg) (*nats.Msg, error)

func (pubs *Pubs) Retry(ctx context.Context, msg *nats.Msg) (*nats.Msg, error) {

	return pubs.natsConn.nc.RequestMsgWithContext(ctx, msg)
}

func RetryFunc(effector Effector, retries int, delay time.Duration) Effector {

	return func(ctx context.Context, m *nats.Msg) (*nats.Msg, error) {

		for r := 0; ; r++ {

			response, err := effector(ctx, m)
			if err == nil || r >= retries {
				if r >= retries {
				}
				return response, err
			}

			fmt.Printf("Attempt %d failed; retrying in %v", r+1, delay)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
		}
	}
}

func (pubs *Pubs) Publish(in *pb.PubMsg, retryBehavior *pb.RetryBehavior) (*pb.PubMsgResponse, error) {

	topic := in.GetTopic()
	responseTopic := topic + "_" + string(in.Header.MsgId) + "_response"
	data := in.GetMsg()

	fmt.Printf(">>>>> pb.PubMsg: %s\n", in)

	// pubs.natsConn.Publish(topic, data)
	msg := nats.Msg{
		Subject: topic,
		Reply:   topic + "_response",
		Data:    data,
	}

	var retryFunc Effector
	retryFunc = RetryFunc(pubs.Retry, int(*retryBehavior.RetryNum), retryBehavior.RetryDelay.AsDuration())
	ctx, cancel := context.WithTimeout(context.Background(), retryBehavior.RetryDelay.AsDuration())
	defer cancel()

	reply, err := retryFunc(ctx, &msg)

	if err != nil {
		return &pb.PubMsgResponse{
			Header: &pb.Header{
				MsgType:     pb.MsgType_MSG_TYPE_PUB_RSP,
				SrcServType: serviceType(),
				DstServType: in.Header.SrcServType,
				ServId:      serviceId()(),
				MsgId:       NextMsgId(),
			},

			RspHeader: &pb.ResponseHeader{
				Status: uint32(pb.Status_ERR_PUBLISHING),
			},

			Msg: fmt.Sprintf("Error publishing msg: %s\n\tto topic: %s\n\tresponse topic: %s\n",
				string(data), topic, responseTopic),
		}, err
	} else {

		return &pb.PubMsgResponse{
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

			Msg: string(reply.Data),
		}, nil
	}
}
