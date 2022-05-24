package conn

import (
	"context"
	"fmt"
	"time"

	"github.com/find-in-docs/sidecar/pkg/log"
	pb "github.com/find-in-docs/sidecar/protos/v1/messages"
	"github.com/nats-io/nats.go"
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

			ctx2, cancel := context.WithTimeout(context.Background(), delay)
			defer cancel()

			fmt.Printf("--- Publishing msg\n")
			response, err := effector(ctx2, m)
			if err == nil || r >= retries {
				return response, err
			}

			fmt.Printf("Attempt %d failed; \n\terr: %v\n\tretrying in %v", r+1, err, delay)

			select {
			case <-ctx2.Done():
				return nil, ctx2.Err()
			default:
			}
		}
	}
}

func (pubs *Pubs) Publish(ctx context.Context, logger *log.Logger, in *pb.PubMsg,
	retryBehavior *pb.RetryBehavior) (*pb.PubMsgResponse, error) {

	topic := in.GetTopic()
	responseTopic := topic + "_" + string(in.Header.MsgId) + "_response"
	data := in.GetMsg()

	// pubs.natsConn.Publish(topic, data)
	msg := nats.Msg{
		Subject: topic,
		Reply:   responseTopic,
		Data:    data,
	}

	var retryFunc Effector
	retryFunc = RetryFunc(pubs.Retry, int(*retryBehavior.RetryNum), retryBehavior.RetryDelay.AsDuration())

	reply, err := retryFunc(ctx, &msg)
	header := pb.Header{
		MsgType:     pb.MsgType_MSG_TYPE_PUB_RSP,
		SrcServType: serviceType(),
		DstServType: in.Header.SrcServType,
		ServId:      serviceId()(),
		MsgId:       NextMsgId(),
	}

	if err != nil {

		logger.Log("Error publishing msg: %s\n\tto topic: %s\n\tresponse topic: %s\n\terr: %s",
			in, topic, responseTopic, err.Error())
		return &pb.PubMsgResponse{
			Header: &header,

			RspHeader: &pb.ResponseHeader{
				Status: uint32(pb.Status_ERR_PUBLISHING),
			},

			Msg: fmt.Sprintf("Error publishing msg: %s\n\tto topic: %s\n\tresponse topic: %s\n",
				string(data), topic, responseTopic),
		}, err
	} else {

		return &pb.PubMsgResponse{
			Header: &header,

			RspHeader: &pb.ResponseHeader{
				Status: uint32(pb.Status_OK),
			},

			Msg: string(reply.Data),
		}, nil
	}
}
