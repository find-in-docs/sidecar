/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/samirgadkari/sidecar/pkg/conn"
	"github.com/samirgadkari/sidecar/pkg/conn/config"
	"github.com/spf13/cobra"
)

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Start a client to talk to a NATS server.",
	Long: `Start a client to talk to a NATS server.
The client will connect, and run some tests between two sidecar instances.`,
	Run: func(cmd *cobra.Command, args []string) {

		config.LoadConfig()

		c, err := conn.InitNATSConn()

		if err != nil {
			fmt.Printf("Error connecting to NATS server: %v\n", err)
			os.Exit(-1)
		}

		var wg sync.WaitGroup
		numMsgs := 10
		wg.Add(numMsgs)

		msgs := make(chan *nats.Msg, 10)

		_, err = c.Subscribe("server", func(msg *nats.Msg) {

			fmt.Printf("Client received msg: %s\n  on topic: %s\n",
				string(msg.Data), msg.Subject)

			msgs <- msg

			wg.Done()
		})
		if err != nil {
			fmt.Printf("Error subscribing to NATS server: %v\n", err)
			os.Exit(-1)
		}

		for i := 0; i < numMsgs; i++ {

			err = c.Publish("client", []byte(" testing ..."))
			if err != nil {
				fmt.Printf("Error publishing to NATS server: %v\n", err)
				os.Exit(-1)
			}
			fmt.Printf("Client sent message\n")
		}
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clientCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clientCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
