package log

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	pb "github.com/samirgadkari/sidecar/protos/v1/messages"
)

type Logger struct {
	forSidecar bool
	client     *pb.SidecarClient
	natsConn   *nats.Conn
	topic      string
	header     *pb.Header
}

func NewLogger(forSidecar bool, client *pb.SidecarClient, natsConn *nats.Conn, header *pb.Header) *Logger {

	return &Logger{
		forSidecar: forSidecar,
		client:     client,
		natsConn:   natsConn,
		topic:      "search.v1.Log",
		header:     header,
	}
}

func (l *Logger) Log(s string, args ...interface{}) {

	str := fmt.Sprintf(s, args...)
	l.LogString(&str)
}

func (l *Logger) LogString(msg *string) error {

	// Print message to stdout
	fmt.Println(*msg)

	header := l.header
	header.MsgType = pb.MsgType_MSG_TYPE_LOG
	header.MsgId = 0

	logMsg := pb.LogMsg{
		Header: header,
		Msg:    *msg,
	}

	if l.forSidecar == false { // for client service

		// Send message to message queue
		logRsp, err := (*l.client).Log(context.Background(), &logMsg)
		if err != nil {
			fmt.Printf("Could not send log message:\n\tmsg: %s\n\terr: %v\n", *msg, err)
			return err
		}

		if logRsp.RspHeader.Status != uint32(pb.Status_OK) {
			fmt.Printf("Error received while logging msg:\n\tmsg: %s\n\tStatus: %d\n",
				*msg, logRsp.RspHeader.Status)
			return err
		}
	} else { // for sidecar service
		l.natsConn.Publish(l.topic, []byte(*msg))
	}

	return nil
}
