package conn

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/samirgadkari/sidecar/protos/v1/messages"
)

const (
	msgChSize = 10
)

type Subs struct {
	natsMsgs chan *nats.Msg
	done     chan struct{}
	msgId    uint32
	natsConn *Conn
}

func InitSubs(natsConn *Conn, srv *Server) {

	natsMsgs := make(chan *nats.Msg, msgChSize)
	done := make(chan struct{})

	srv.Subs = &Subs{

		natsMsgs: natsMsgs,
		done:     done,
		msgId:    1,
		natsConn: natsConn,
	}

	SendIncomingMsgsToClient(srv.Subs.natsMsgs, done)
}

func (subs *Subs) Subscribe(in *messages.SubMsg) (*messages.SubMsgResponse, error) {

	fmt.Printf("Received SubMsg: %v\n", in)
	topic := in.GetTopic()

	subs.natsConn.Subscribe(topic, func(m *nats.Msg) {
		fmt.Printf("Received msg from NATS: %v\n", m)
		subs.natsMsgs <- m
	})

	srcHeader := in.GetHeader()
	header := messages.Header{
		DstServType: srcHeader.GetSrcServType(),
		SrcServType: srcHeader.GetDstServType(),
		ServId:      srcHeader.GetServId(),
		MsgId:       0,
	}

	rspHeader := messages.ResponseHeader{

		Status: uint32(messages.Status_OK),
	}

	subMsgRsp := messages.SubMsgResponse{

		Header:    &header,
		RspHeader: &rspHeader,
		Msg:       "OK",
	}

	return &subMsgRsp, nil
}

func SendIncomingMsgsToClient(natsMsgs chan *nats.Msg,
	done chan struct{}) {

	var m *nats.Msg
	/*
		var header *messages.Header
		var err error
	*/

	go func() {
		for {
			select {
			case m = <-natsMsgs:
				fmt.Printf("Got msg from NATS server:\n\t%v\n", m)

				/*
					header = l.GetHeader()
					header.MsgId = logs.msgId
					logs.msgId += 1

					err = logs.natsConn.Publish("logs", []byte(l.String()))
					if err != nil {
						fmt.Printf("Error publishing to NATS server: %v\n", err)
						os.Exit(-1)
					}
					fmt.Printf("Server sent message\n")
				*/

			case <-done:
				// drain logs channel here before breaking out of the forever loop
			}
		}
	}()
}
