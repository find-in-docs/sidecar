package client

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/samirgadkari/sidecar/protos/v1/messages"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/durationpb"
)

type SC struct {
	client messages.SidecarClient
	header *messages.Header
}

func Connect(serverAddr string) (*grpc.ClientConn, *SC, error) {

	// var opts []grpc.DialOption

	fmt.Printf("serverAddr: %s\n", serverAddr)
	conn, err := grpc.Dial(serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("Error creating GRPC channel\n:\terr: %v\n", err)
		os.Exit(-1)
	}

	fmt.Printf("conn: %v\n", conn)

	client := messages.NewSidecarClient(conn)
	fmt.Printf("GRPC connection to sidecar created\n")

	header := messages.Header{
		SrcServType: "",
		DstServType: "",
		ServId:      []byte(""),
		MsgId:       uint32(0),
	}

	sc := &SC{client, &header}
	err = sc.Register()
	if err != nil {
		fmt.Printf("Error registering client:\n\terr: %v\n", err)
		os.Exit(-1_
	}

	return conn, sc, nil
}

func (sc *SC) Register() error {

	var circuitConsecutiveFailures uint32 = 3

	// Go Duration is in the time package: https://pkg.go.dev/time#Duration
	// Go Duration maps to protobuf Duration.
	// You can convert between them using durationpb:
	//   https://pkg.go.dev/google.golang.org/protobuf/types/known/durationpb
	debounceDelayDuration, err := time.ParseDuration("5s")
	if err != nil {
		fmt.Printf("Error creating Golang time duration.\nerr: %v\n", err)
		os.Exit(-1)
	}
	debounceDelay := durationpb.New(debounceDelayDuration)

	var retryNum uint32 = 2

	retryDelayDuration, err := time.ParseDuration("2s")
	if err != nil {
		fmt.Printf("Error creating Golang time duration.\nerr: %v\n", err)
		os.Exit(-1)
	}
	retryDelay := durationpb.New(retryDelayDuration)
	header := messages.Header{
		SrcServType: "postgresService",
		DstServType: "sidecarService",
		ServId:      []byte(""),
		MsgId:       0,
	}

	rMsg := &messages.RegistrationMsg{
		Header: &header,

		CircuitFailureThreshold: &circuitConsecutiveFailures,
		DebounceDelay:           debounceDelay,
		RetryNum:                &retryNum,
		RetryDelay:              retryDelay,
	}

	rRsp, err := sc.client.Register(context.Background(), rMsg)
	if err != nil {
		fmt.Printf("Sending Registration caused error:\n\terr: %v\n", err)
		return err
	}

	fmt.Printf("Registration message sent\n\tRegRsp: %v\n\tStatus: %d\n\tMsg: %s\n",
		*rRsp, rRsp.RspHeader.Status, rRsp.Msg)

	sc.header = &header
	sc.header.SrcServType = rRsp.Header.DstServType
	sc.header.DstServType = rRsp.Header.SrcServType
	sc.header.ServId = rRsp.Header.ServId
	sc.header.MsgId = 0

	return nil
}
