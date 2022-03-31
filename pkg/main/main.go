package main

import (
	"context"
	"fmt"
	"time"

	"github.com/samirgadkari/sidecar/pkg/config"
	"github.com/samirgadkari/sidecar/pkg/conn"
	"github.com/samirgadkari/sidecar/pkg/utils"
)

const (
	thisServType = "sidecarService"
)

func main() {

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
	fmt.Println("User pressed Enter key")

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
}
