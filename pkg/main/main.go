package main

import (
	"context"
	// "fmt"
  // "os"
	// "time"

	"github.com/find-in-docs/sidecar/pkg/config"
	"github.com/find-in-docs/sidecar/pkg/conn"
	// "github.com/find-in-docs/sidecar/pkg/utils"
)

const (
	thisServType = "sidecarService"
)

func main() {

  /*
  fmt.Printf("Sleeping for 10000 seconds\n")
  time.Sleep(10000 * time.Second)
  os.Exit(0)
  */

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

  /*
  This section was for testing if the GoRoutines are all
  finished before exiting. This was only meant as a debug
  mechanism. Since we're now running in minikube, this mechanism
  will not work, so it is commented out for now.
  Not sure what to replace it with, at this time.

	fmt.Println("Press the Enter key to stop")
	fmt.Scanln()
	fmt.Println("User pressed Enter key")

	// TODO: The grcp.GracefulStop() routine is blocking forever.
	// This is probably because some RPC is not completed.
	// Using grcp.Stop() temporarily.
	srv.GrcpServer.Stop()

	sleepDur, _ := time.ParseDuration("3s")
	fmt.Printf("Sleeping for %s seconds\n", sleepDur)
	time.Sleep(sleepDur)

	utils.ListGoroutinesRunning()
  */

	conn.BlockForever()
}
