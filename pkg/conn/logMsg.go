package conn

import (
	"fmt"
	"os"

	"github.com/samirgadkari/sidecar/pkg/log"
	pb "github.com/samirgadkari/sidecar/protos/v1/messages"
)

const (
	logChSize = 20
)

type Logs struct {
	logger   *log.Logger
	logs     chan *pb.LogMsg
	done     chan struct{}
	natsConn *Conn
}

func InitLogs(natsConn *Conn, srv *Server) {

	logs := make(chan *pb.LogMsg, logChSize)
	done := make(chan struct{})
	header := &pb.Header{
		DstServType: "",
		SrcServType: "sidecar",
	}

	srv.Logs = &Logs{
		logger:   log.NewLogger(natsConn.nc, header),
		logs:     logs,
		done:     done,
		natsConn: natsConn,
	}

	SendLogsToMsgQueue(srv.Logs, done)
}

func (l *Logs) ReceivedLogMsg(in *pb.LogMsg) (*pb.LogMsgResponse, error) {

	l.logs <- in

	srcHeader := in.GetHeader()
	header := pb.Header{
		MsgType:     pb.MsgType_MSG_TYPE_LOG_RSP,
		DstServType: srcHeader.GetSrcServType(),
		SrcServType: srcHeader.GetDstServType(),
		ServId:      srcHeader.GetServId(),
		MsgId:       NextMsgId(),
	}

	rspHeader := pb.ResponseHeader{
		Status: uint32(pb.Status_OK),
	}

	logRsp := &pb.LogMsgResponse{
		Header:    &header,
		RspHeader: &rspHeader,
		Msg:       "OK",
	}

	return logRsp, nil
}

func SendLogsToMsgQueue(logs *Logs, done chan struct{}) {

	var l *pb.LogMsg
	var header *pb.Header
	var err error

	go func() {
		for {
			select {
			case l = <-logs.logs:
				fmt.Printf("Got log msg:\n\tHeader: %s\n\tMsg: %s\n",
					l.Header, l.Msg)

				header = l.GetHeader()
				header.MsgId = NextMsgId()

				err = logs.natsConn.Publish("search.log.v1", []byte(l.String()))
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
