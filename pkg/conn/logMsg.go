package conn

import (
	"context"
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

func InitLogs(ctx context.Context, natsConn *Conn, srv *Server) {

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

	SendLogsToMsgQueue(ctx, srv.Logs)
}

func (l *Logs) ReceivedLogMsg(in *pb.LogMsg) (*pb.LogMsgResponse, error) {

	l.logs <- in

	logRsp := &pb.LogMsgResponse{
		Header: &pb.Header{
			MsgType:     pb.MsgType_MSG_TYPE_LOG_RSP,
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

	return logRsp, nil
}

func processLogMsg(logs *Logs, l *pb.LogMsg) {

	var header *pb.Header
	var err error

	logs.logger.PrintMsg("Got log msg: %s\n", l)

	header = l.GetHeader()
	header.MsgId = NextMsgId()

	err = logs.natsConn.Publish("search.log.v1", []byte(l.String()))
	if err != nil {
		fmt.Printf("Error publishing to NATS server: %v\n", err)
		os.Exit(-1)
	}
	fmt.Printf("Server sent message\n")
}

func SendLogsToMsgQueue(ctx context.Context, logs *Logs) {

	var l *pb.LogMsg

	go func() {
	LOOP:
		for {
			select {
			case l = <-logs.logs:
				processLogMsg(logs, l)

			case <-ctx.Done():
				// drain logs channel here before breaking out of the forever loop
				for i := 0; i < len(logs.logs); i++ {
					processLogMsg(logs, <-logs.logs)
				}
				break LOOP
			}
		}

		fmt.Printf("GOROUTINE 4 completed in function SendLogsToMsgQueue\n")
	}()
}
