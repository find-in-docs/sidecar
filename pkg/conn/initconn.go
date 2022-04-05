package conn

import (
	"fmt"
	"net"
	"os"

	"github.com/samirgadkari/sidecar/pkg/utils"
	pb "github.com/samirgadkari/sidecar/protos/v1/messages"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func InitNATSconn() (*Conn, error) {

	c, err := NewNATSConn(viper.GetString("natsUrl"))

	if err != nil {
		os.Exit(-1)
	}

	return c, nil
}

func InitGRPCconn(srv *Server) {

	goroutineName := "InitGRCPconn"
	err := utils.StartGoroutine(goroutineName, func() {
		s := grpc.NewServer()

		sidecarServiceAddr := viper.GetString("sidecarServiceAddr")

		lis, err := net.Listen("tcp", sidecarServiceAddr)
		if err != nil {
			fmt.Printf("Error starting net listener\n\terr: %v\n", err)
			os.Exit(-1)
		}

		pb.RegisterSidecarServer(s, srv)

		srv.GrcpServer = s

		fmt.Printf("Server listening at %v\n", lis.Addr())
		if err = s.Serve(lis); err != nil {
			fmt.Printf("Failed to serve: %v\n", err)
		}
		fmt.Printf("GOROUTINE 3 for GRCP server completed\n\n")
		utils.GoroutineEnded(goroutineName)
	})

	if err != nil {
		fmt.Printf("Error starting goroutine: %v\n", err)
		os.Exit(-1)
	}
}

func Initconns() (*Conn, *Server, error) {

	// Initialize empty server. Load it with values you need later.
	srv := &Server{}

	InitGRPCconn(srv)

	natsConn, err := InitNATSconn()
	if err != nil {
		fmt.Printf("Error initializing NATS conn:\n\terr: %v\n", err)
		return nil, nil, err
	}

	return natsConn, srv, nil
}
