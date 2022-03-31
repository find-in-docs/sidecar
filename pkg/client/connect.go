package client

import (
	"context"
	"fmt"
	"time"

	"github.com/samirgadkari/sidecar/pkg/utils"
	pb "github.com/samirgadkari/sidecar/protos/v1/messages"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/durationpb"
)

type SC struct {
	client pb.SidecarClient
	header *pb.Header
	Logger *Logger
}

func InitSidecar(serviceName string, regParams *pb.RegistrationParams) (*SC, error) {

	sidecarServiceAddr := viper.GetString("sidecarServiceAddr")
	conn, err := Connect(serviceName, sidecarServiceAddr, regParams)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to client: %w\n", err)
	}

	client := pb.NewSidecarClient(conn)
	fmt.Printf("GRPC connection to sidecar created\n")

	header := &pb.Header{
		SrcServType: serviceName,
		DstServType: "sidecar",
		ServId:      []byte(""),
		MsgId:       0,
	}

	logger := NewLogger(&client, header)
	sc := &SC{client, header, logger}
	err = sc.Register(serviceName, regParams)
	if err != nil {
		sc.Logger.Log("Error registering client: %s", err.Error())
		return nil, fmt.Errorf("Error registering client: %w\n", err)
	}

	return sc, nil
}

func Connect(serviceName string, serverAddr string, regParams *pb.RegistrationParams) (*grpc.ClientConn, error) {

	// var opts []grpc.DialOption

	fmt.Printf("serverAddr: %s\n", serverAddr)
	conn, err := grpc.Dial(serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("Error creating GRPC channel: %w", err)
	}

	fmt.Printf("conn: %v\n", conn)
	return conn, nil
}

func DefaultRegParams(sc *SC) (*pb.RegistrationParams, error) {

	var circuitConsecutiveFailures uint32 = 3

	// Go Duration is in the time package: https://pkg.go.dev/time#Duration
	// Go Duration maps to protobuf Duration.
	// You can convert between them using durationpb:
	//   https://pkg.go.dev/google.golang.org/protobuf/types/known/durationpb
	debounceDelayDuration, err := time.ParseDuration("5s")
	if err != nil {
		return nil, fmt.Errorf("Error creating Golang time duration: %w\n", err)
	}
	debounceDelay := durationpb.New(debounceDelayDuration)

	var retryNum uint32 = 2

	retryDelayDuration, err := time.ParseDuration("2s")
	if err != nil {
		return nil, fmt.Errorf("Error creating Golang time duration: %w\n", err)
	}
	retryDelay := durationpb.New(retryDelayDuration)

	return &pb.RegistrationParams{
		CircuitFailureThreshold: &circuitConsecutiveFailures,
		DebounceDelay:           debounceDelay,

		Retry: &pb.RetryBehavior{
			RetryNum:   &retryNum,
			RetryDelay: retryDelay,
		},
	}, nil
}

func (sc *SC) Register(serviceName string, regParams *pb.RegistrationParams) error {

	var rParams *pb.RegistrationParams

	rParams = regParams
	if rParams == nil {

		var err error

		rParams, err = DefaultRegParams(sc)
		if err != nil {
			return fmt.Errorf("Error creating default registration params: %s", err.Error())
		}
	}

	header := sc.header
	header.MsgType = pb.MsgType_MSG_TYPE_REG
	header.MsgId = 0

	rMsg := &pb.RegistrationMsg{
		Header: header,

		ServiceName: serviceName,
		RegParams:   rParams,
	}

	rRsp, err := sc.client.Register(context.Background(), rMsg)
	sc.Logger.Log("Registration msg sent:\n\t%s\n", rMsg)
	if err != nil {
		sc.Logger.Log("Sending Registration caused error: %v\n", err)
		return err
	}

	sc.Logger.Log("Registration rsp received:\n\t%s\n", rRsp)

	sc.header.ServId = rRsp.Header.ServId

	return nil
}

func (sc *SC) Pub(ctx context.Context, topic string, data []byte, rb *pb.RetryBehavior) error {

	header := sc.header
	header.MsgType = pb.MsgType_MSG_TYPE_PUB
	header.MsgId = 0

	pubMsg := pb.PubMsg{
		Header: sc.header,
		Topic:  topic,
		Msg:    data,
		Retry:  rb,
	}

	pubRsp, err := sc.client.Pub(ctx, &pubMsg)
	sc.Logger.Log("Pub message sent: %s\n", pubMsg.String())
	if err != nil {
		sc.Logger.Log("Could not publish to topic: %s\n\tmessage:\n\tmsg: %s %v\n",
			topic, string(data), err)
		return err
	}
	sc.Logger.Log("Pub rsp received: %s\n", pubRsp)

	if pubRsp.RspHeader.Status != uint32(pb.Status_OK) {
		sc.Logger.Log("Error received while publishing to topic:\n\ttopic: %s\n\tmsg: %s %v\n",
			topic, data, err)
		return err
	}

	return nil
}

func (sc *SC) Sub(ctx context.Context, topic string, chanSize uint32) error {

	header := sc.header
	header.MsgType = pb.MsgType_MSG_TYPE_SUB
	header.MsgId = 0

	subMsg := pb.SubMsg{
		Header:   header,
		Topic:    topic,
		ChanSize: chanSize,
	}

	subRsp, err := sc.client.Sub(ctx, &subMsg)
	sc.Logger.Log("Sub message sent:\n\t%s\n", &subMsg)
	if err != nil {
		sc.Logger.Log("Could not subscribe to topic: %s %v\n",
			topic, err)
		return err
	}
	sc.Logger.Log("Sub rsp received:\n\t%s\n", subRsp)

	if subRsp.RspHeader.Status != uint32(pb.Status_OK) {
		sc.Logger.Log("Error received while publishing to topic:\n\ttopic: %s %v\n",
			topic, err)
		return err
	}

	return nil
}

func (sc *SC) Unsub(ctx context.Context, topic string) error {

	header := sc.header
	header.MsgType = pb.MsgType_MSG_TYPE_UNSUB
	header.MsgId = 0

	unsubMsg := pb.UnsubMsg{
		Header: header,
		Topic:  topic,
	}

	unsubRsp, err := sc.client.Unsub(ctx, &unsubMsg)
	sc.Logger.Log("Unsub message sent:\n\t%s\n", &unsubMsg)
	if err != nil {
		sc.Logger.Log("Could not unsubscribe from topic:\n\ttopic: %s %v\n", topic, err)
		return err
	}

	if unsubRsp.RspHeader.Status != uint32(pb.Status_OK) {
		sc.Logger.Log("Error received while unsubscribing to topic:\n\ttopic: %s %v", topic, err)
		return err
	}

	return nil
}

type Response struct {
	response *pb.SubTopicResponse
	err      error
}

func (sc *SC) ProcessSubMsgs(ctx context.Context, topic string,
	chanSize uint32, f func(*pb.SubTopicResponse)) error {

	err := sc.Sub(ctx, topic, chanSize)
	if err != nil {
		return err
	}

	responseCh := sc.Recv(ctx, topic)

	goroutineName := "ProcessSubMsgs"
	utils.StartGoroutine(goroutineName,
		func() {
			subscribedTopic := topic

		LOOP:
			for {
				select {

				case r := <-responseCh:
					if r.err != nil {
						sc.Logger.Log("Error receiving from sidecar: %v\n", r.err)
						_ = sc.Unsub(ctx, subscribedTopic)
						break LOOP
					}

					// Do not log received message to NATS. This creates a loop.

					f(r.response)

				case <-ctx.Done():
					_ = sc.Unsub(ctx, subscribedTopic)

					if ctx.Err() != nil {
						sc.Logger.Log("Done channel signaled: %v\n", err)
					}
					break LOOP
				}
			}
			sc.Logger.Log("GOROUTINE 2 completed in function ProcessSubMsgs\n")
			utils.GoroutineEnded(goroutineName)
		})

	return nil
}

func (sc *SC) Recv(ctx context.Context, topic string) <-chan *Response {

	header := sc.header
	header.MsgId = 0

	recvMsg := pb.Receive{
		Header: header,
		Topic:  topic,
	}

	responseCh := make(chan *Response)

	goroutineName := "Recv"
	utils.StartGoroutine(goroutineName,
		func() {
		LOOP:
			for {
				subTopicRsp, err := sc.client.Recv(ctx, &recvMsg)
				if err != nil {
					sc.Logger.Log("Could not receive from sidecar - err: %v\n", err)
					break LOOP
				}

				// Do not log received message to NATS. This creates a loop.

				responseCh <- &Response{
					subTopicRsp,
					nil,
				}

				if ctx.Err() != nil {
					break LOOP
				}
			}
			sc.Logger.Log("GOROUTINE 1 completed in function Recv\n")
			utils.GoroutineEnded(goroutineName)
		})

	return responseCh
}
