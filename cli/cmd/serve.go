/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/nats-io/nats.go"
	"github.com/samirgadkari/sidecar/pkg/config"
	"github.com/samirgadkari/sidecar/pkg/conn"
	"github.com/spf13/cobra"
)

const (
	thisServType = "sidecarService"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start a server to talk to a NATS client and listen to GRPC requests.",
	Long: `Start a server to talk to a NATS client and listen to GRPC requests. 
The server will connect, and run some tests between two sidecar instances`,
	Run: func(cmd *cobra.Command, args []string) {

		config.Load()

		natsConn, srv, err := conn.Initconns()
		if err != nil {
			return
		}

		conn.InitLogs(natsConn, srv)
		conn.InitPubs(natsConn, srv)
		conn.InitSubs(natsConn, srv)

		numMsgs := 10

		msgs := make(chan *nats.Msg, 10)

		_, err = natsConn.Subscribe("client", func(msg *nats.Msg) {

			fmt.Printf("Server received msg: %s\n  on topic: %s\n",
				string(msg.Data), msg.Subject)

			msgs <- msg

		})
		if err != nil {
			fmt.Printf("Error subscribing to NATS server: %v\n", err)
			os.Exit(-1)
		}

		for i := 0; i < numMsgs; i++ {

			err = natsConn.Publish("server", []byte("++++ testing ++++"))
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
