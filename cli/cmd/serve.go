/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/samirgadkari/sidecar/pkg/config"
	"github.com/samirgadkari/sidecar/pkg/conn"
	"github.com/samirgadkari/sidecar/pkg/utils"
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

		natsConn, srv, err := conn.Initconns()
		if err != nil {
			return
		}

		conn.InitLogs(ctx, natsConn, srv)
		conn.InitPubs(natsConn, srv)
		conn.InitSubs(natsConn, srv)

		fmt.Println("Press the Enter key to stop")
		fmt.Scanln()

		// TODO: The grcp.GracefulStop() routine is blocking forever.
		// This is probably because some RPC is not completed.
		// Using grcp.Stop() temporarily.
		srv.GrcpServer.Stop()

		cancel()

		sleepDur, _ := time.ParseDuration("3s")
		fmt.Printf("Sleeping for %s seconds\n", sleepDur)
		time.Sleep(sleepDur)

		utils.ListGoroutinesRunning()

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
