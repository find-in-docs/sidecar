/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/samirgadkari/sidecar/pkg/connection"
	"github.com/samirgadkari/sidecar/pkg/connection/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a server to talk to a NATS client.",
	Long: `Start a server to talk to a NATS client. 
The server will connect, and run some tests between two sidecar instances`,
	Run: func(cmd *cobra.Command, args []string) {

		config.LoadConfig()

		c, err := connection.New(viper.GetString("natsUrl"))

		if err != nil {
			fmt.Printf("Error connecting to NATS server: %v\n", err)
			os.Exit(-1)
		}

		wg := new(sync.WaitGroup)
		wg.Add(2)

		msgs := make(chan *nats.Msg, 10)

		_, err = c.Subscribe("foo", func(msg *nats.Msg) {

			fmt.Printf("Server received msg: %s\n  on topic: %s\n",
				string(msg.Data), msg.Subject)

			msgs <- msg
			wg.Done()

		})
		if err != nil {
			fmt.Printf("Error subscribing to NATS server: %v\n", err)
			os.Exit(-1)
		}

		err = c.Publish("foo", []byte("++++ testing ++++"))
		if err != nil {
			fmt.Printf("Error publishing to NATS server: %v\n", err)
			os.Exit(-1)
		}
		fmt.Printf("Server sent message\n")
		// wg.Done()
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
