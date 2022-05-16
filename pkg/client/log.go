package client

import (
	"context"
	"fmt"

	pb "github.com/find-in-docs/sidecar/protos/v1/messages"
)

type Logger struct {
	client *pb.SidecarClient
	topic  string
	header *pb.Header
}

func NewLogger(client *pb.SidecarClient, header *pb.Header) *Logger {

	return &Logger{
		client: client,
		topic:  "search.log.v1",
		header: header,
	}
}

func (l *Logger) Log(s string, args ...interface{}) {

	str := fmt.Sprintf(s, args...)
	l.logString(&str)
}

func (l *Logger) logString(msg *string) {

	// Print message to stdout
	fmt.Println(*msg)

	header := l.header
	header.MsgType = pb.MsgType_MSG_TYPE_LOG
	header.MsgId = 0

	logMsg := pb.LogMsg{
		Header: header,
		Msg:    *msg,
	}

	// Send message to message queue
	_, err := (*l.client).Log(context.Background(), &logMsg)
	if err != nil {
		fmt.Printf("Could not send log message:\n\tmsg: %s\n\terr: %v\n", *msg, err)
		return
	}
}
