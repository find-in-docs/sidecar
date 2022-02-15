/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	conn "github.com/samirgadkari/sidecar/pkg/connection"
	"github.com/samirgadkari/sidecar/pkg/connection/config"
	"github.com/samirgadkari/sidecar/protos/v1/messages"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

const (
	thisServType = "sidecarService"
)

type server struct {
	messages.UnimplementedSidecarServer
	header *messages.Header
}

func servId() []byte {
	uuid := uuid.New()
	servId, err := uuid.MarshalText()
	if err != nil {
		fmt.Printf("Error converting UUID: %v to text\n\terr: %v\n", servId, err)
		os.Exit(-1)
	}

	return servId
}

func (s *server) Register(ctx context.Context, in *messages.RegistrationMsg) (*messages.RegistrationResponse, error) {
	fmt.Printf("Received RegistrationMsg: %v\n", in)

	rspHeader := messages.ResponseHeader{
		Status: uint32(messages.Status_OK),
	}

	header := messages.Header{
		SrcServType: "sidecarService",
		DstServType: in.Header.SrcServType,
		ServId:      servId(),
		MsgId:       0,
	}
	s.header = &header

	regRsp := &messages.RegistrationResponse{
		Header:    &header,
		RspHeader: &rspHeader,
		Msg:       "OK",
	}
	fmt.Printf("Sending regRsp: %#v\n", *regRsp)

	return regRsp, nil
}

func (s *server) Log(ctx context.Context, in *messages.LogMsg) (*messages.LogResponse, error) {

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

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a server to talk to a NATS client and listen to GRPC requests.",
	Long: `Start a server to talk to a NATS client and listen to GRPC requests. 
The server will connect, and run some tests between two sidecar instances`,
	Run: func(cmd *cobra.Command, args []string) {

		config.LoadConfig()

		c, err := conn.New(viper.GetString("natsUrl"))

		if err != nil {
			os.Exit(-1)
		}

		go func() {
			sidecarServiceAddr := viper.GetString("sidecarServiceAddr")

			lis, err := net.Listen("tcp", sidecarServiceAddr)
			if err != nil {
				fmt.Printf("Error starting net listener\n\terr: %v\n", err)
				os.Exit(-1)
			}

			s := grpc.NewServer()
			messages.RegisterSidecarServer(s, &server{})

			fmt.Printf("Server listening at %v\n", lis.Addr())
			if err = s.Serve(lis); err != nil {
				fmt.Print("Failed to serve: %v\n", err)
			}
		}()

		numMsgs := 10

		msgs := make(chan *nats.Msg, 10)

		_, err = c.Subscribe("client", func(msg *nats.Msg) {

			fmt.Printf("Server received msg: %s\n  on topic: %s\n",
				string(msg.Data), msg.Subject)

			msgs <- msg

		})
		if err != nil {
			fmt.Printf("Error subscribing to NATS server: %v\n", err)
			os.Exit(-1)
		}

		for i := 0; i < numMsgs; i++ {

			err = c.Publish("server", []byte("++++ testing ++++"))
			if err != nil {
				fmt.Printf("Error publishing to NATS server: %v\n", err)
				os.Exit(-1)
			}
			fmt.Printf("Server sent message\n")
		}

		conn.BlockForever()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
