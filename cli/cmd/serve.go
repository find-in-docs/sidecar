/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"fmt"

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

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		natsConn, srv, err := conn.Initconns(ctx)
		if err != nil {
			return
		}

		conn.InitLogs(ctx, natsConn, srv)
		conn.InitPubs(natsConn, srv)
		conn.InitSubs(natsConn, srv)

		fmt.Println("Press the Enter key to stop")
		fmt.Scanln()

		fmt.Printf("Gracefully stopping GRCP server\n")
		srv.GrcpServer.GracefulStop()

		fmt.Printf("Cancelling Logs GOROUTINE\n")
		cancel() // see if this cancels the goroutine

		fmt.Printf("Waiting for cancel---\n")
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
