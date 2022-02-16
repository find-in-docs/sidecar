package conn

import (
	"fmt"

	"github.com/samirgadkari/sidecar/protos/v1/messages"
)

const (
	logChSize = 20
)

type Logs struct {
	logs chan *messages.LogMsg
	done chan struct{}
}

func InitLogs(srv *Server) {

	logs := make(chan *messages.LogMsg, logChSize)
	done := make(chan struct{})

	SendLogsToMsgQueue(logs, done)

	srv.Logs = &Logs{
		logs: logs,
		done: done,
	}
}

func (l *Logs) ReceivedLogMsg(in *messages.LogMsg) (*messages.LogResponse, error) {

	l.logs <- in

	fmt.Printf("Received LogMsg: %#v\n", in)

	rspHeader := messages.ResponseHeader{
		Status: uint32(messages.Status_OK),
	}

	logRsp := &messages.LogResponse{
		Header: &rspHeader,
		Msg:    "OK",
	}
	fmt.Printf("Sending logRsp: %#v\n", *logRsp)

	return logRsp, nil
}

func SendLogsToMsgQueue(logs chan *messages.LogMsg, done chan struct{}) {

	var l *messages.LogMsg

	go func() {
		for {
			select {
			case l = <-logs:
				fmt.Printf("Got log msg:\n\t%v\n", l)

			case <-done:
				// drain logs channel here before breaking out of the forever loop
			}
		}
	}()
}
