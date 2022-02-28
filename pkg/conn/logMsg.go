package conn

import (
	"fmt"
	"os"

	"github.com/samirgadkari/sidecar/protos/v1/messages"
)

const (
	logChSize = 20
)

type Logs struct {
	logs     chan *messages.LogMsg
	done     chan struct{}
	msgId    uint32
	natsConn *Conn
}

func InitLogs(natsConn *Conn, srv *Server) {

	logs := make(chan *messages.LogMsg, logChSize)
	done := make(chan struct{})

	srv.Logs = &Logs{
		logs:     logs,
		done:     done,
		msgId:    1,
		natsConn: natsConn,
	}

	SendLogsToMsgQueue(srv.Logs, done)
}

func (l *Logs) ReceivedLogMsg(in *messages.LogMsg) (*messages.LogMsgResponse, error) {

	l.logs <- in

	fmt.Printf("Received LogMsg: %v\n", in)

	srcHeader := in.GetHeader()
	header := messages.Header{
		MsgType:     messages.MsgType_MSG_TYPE_LOG_RSP,
		DstServType: srcHeader.GetSrcServType(),
		SrcServType: srcHeader.GetDstServType(),
		ServId:      srcHeader.GetServId(),
		MsgId:       0,
	}

	rspHeader := messages.ResponseHeader{
		Status: uint32(messages.Status_OK),
	}

	logRsp := &messages.LogMsgResponse{
		Header:    &header,
		RspHeader: &rspHeader,
		Msg:       "OK",
	}

	return logRsp, nil
}

func SendLogsToMsgQueue(logs *Logs, done chan struct{}) {

	var l *messages.LogMsg
	var header *messages.Header
	var err error

	go func() {
		for {
			select {
			case l = <-logs.logs:
				fmt.Printf("Got log msg:\n\t%v\n", l)

				header = l.GetHeader()
				header.MsgId = logs.msgId
				logs.msgId += 1

				err = logs.natsConn.Publish("search.v1.logs", []byte(l.String()))
				if err != nil {
					fmt.Printf("Error publishing to NATS server: %v\n", err)
					os.Exit(-1)
				}
				fmt.Printf("Server sent message\n")

			case <-done:
				// drain logs channel here before breaking out of the forever loop
			}
		}
	}()
}
