package conn

import (
	"context"
	"fmt"
	"time"

	"github.com/find-in-docs/sidecar/pkg/utils"
	pb "github.com/find-in-docs/sidecar/protos/v1/messages"
	"github.com/spf13/viper"
)

func (s *Server) StreamFlowControl(stream pb.Sidecar_DocUploadStreamServer,
	unprocessedMsgs uint64) {

	thresholdOFF := viper.GetUint64("nats.jetstream.thresholdOFF")
	thresholdON := viper.GetUint64("nats.jetstream.thresholdON")

	if unprocessedMsgs > thresholdOFF {

		fmt.Printf("Flow OFF\n")
		stream.Send(&pb.DocUploadResponse{
			Control: &pb.StreamControl{
				Flow: pb.StreamFlow_OFF,
			},
			AckMsgNumber: 0,
		})
	} else if unprocessedMsgs <= thresholdON {

		fmt.Printf("Flow ON\n")
		stream.Send(&pb.DocUploadResponse{
			Control: &pb.StreamControl{
				Flow: pb.StreamFlow_ON,
			},
			AckMsgNumber: 0,
		})
	}
}

func (s *Server) ThrottleGRPCSender(ctx context.Context,
	stream pb.Sidecar_DocUploadStreamServer) error {

	var err error

	numSeconds, err := time.ParseDuration(viper.GetString("nats.jetstream.throttleCheckPeriod"))
	if err != nil {
		return fmt.Errorf("Error Parsing Jetstream throttle check period: %w", err)
	}
	fmt.Printf("numSeconds: %v\n", numSeconds)

	jsName := viper.GetString("nats.jetstream.name")

	utils.StartGoroutine("uploadDocsClientRecv", func() {
	LOOP:
		for {
			select {
			case <-ctx.Done():
				if ctx.Err() != nil {
					s.Logs.logger.Log("Done channel signaled: %v\n", err)
				}
				break LOOP
			case <-time.After(numSeconds):

				jsInfo, err := s.Pubs.natsConn.js.StreamInfo(jsName)
				if err != nil {
					return
				}

				fmt.Printf("jsInfo.State.Msgs: %d\n", jsInfo.State.Msgs)
				s.StreamFlowControl(stream, jsInfo.State.Msgs)
			}
		}
	})

	return nil
}
