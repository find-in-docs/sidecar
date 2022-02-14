/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/samirgadkari/sidecar/pkg/connection"
	"github.com/samirgadkari/sidecar/pkg/connection/config"
	"github.com/samirgadkari/sidecar/protos/v1/messages"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type server struct {
	messages.UnimplementedSidecarServer
}

func (s *server) Register(ctx context.Context, in *messages.RegistrationMsg) (*messages.RegistrationResponse, error) {
	fmt.Printf("Received RegistrationMsg: %v\n", in)

	header := messages.ResponseHeader{
		Status: uint32(messages.Status_OK),
	}
	return &messages.RegistrationResponse{
		Header: &header,
		Msg:    "OK",
	}, nil
}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a server to talk to a NATS client and listen to GRPC requests.",
	Long: `Start a server to talk to a NATS client and listen to GRPC requests. 
The server will connect, and run some tests between two sidecar instances`,
	Run: func(cmd *cobra.Command, args []string) {

		config.LoadConfig()

		c, err := connection.New(viper.GetString("natsUrl"))

		if err != nil {
			fmt.Printf("Error connecting to NATS server: %v\n", err)
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

			fmt.Printf("Server listening at %v", lis.Addr())
			if err = s.Serve(lis); err != nil {
				fmt.Print("Failed to serve: %v\n", err)
			}
		}()

		var wg sync.WaitGroup
		numMsgs := 10
		wg.Add(numMsgs)

		msgs := make(chan *nats.Msg, 10)

		_, err = c.Subscribe("client", func(msg *nats.Msg) {

			fmt.Printf("Server received msg: %s\n  on topic: %s\n",
				string(msg.Data), msg.Subject)

			msgs <- msg

			wg.Done()
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
		wg.Wait()
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
