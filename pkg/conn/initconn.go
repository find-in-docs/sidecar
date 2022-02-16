package conn

import (
	"fmt"
	"net"
	"os"

	"github.com/samirgadkari/sidecar/pkg/conn/config"
	"github.com/samirgadkari/sidecar/protos/v1/messages"
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

	go func() {
		s := grpc.NewServer()

		sidecarServiceAddr := viper.GetString("sidecarServiceAddr")

		lis, err := net.Listen("tcp", sidecarServiceAddr)
		if err != nil {
			fmt.Printf("Error starting net listener\n\terr: %v\n", err)
			os.Exit(-1)
		}

		messages.RegisterSidecarServer(s, srv)

		fmt.Printf("Server listening at %v\n", lis.Addr())
		if err = s.Serve(lis); err != nil {
			fmt.Print("Failed to serve: %v\n", err)
		}
	}()
}

func Initconns() (*Conn, *Server, error) {

	// Initialize empty server. Load it with values you need later.
	srv := &Server{}

	config.LoadConfig()

	InitGRPCconn(srv)

	natsConn, err := InitNATSconn()
	if err != nil {
		fmt.Printf("Error initializing NATS conn:\n\terr: %v\n", err)
		return nil, nil, err
	}

	return natsConn, srv, nil
}
