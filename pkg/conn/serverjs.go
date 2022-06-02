package conn

import (
	"context"
	"fmt"
	"io"

	pb "github.com/find-in-docs/sidecar/protos/v1/messages"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) pubNATS(topic, name string, in *pb.DocUpload) error {

	js := s.Pubs.natsConn.js

	fmt.Printf("PubNATS: entered\n")
	defer fmt.Printf("pubNATS: exiting\n")

	s.Logs.logger.Log("PubNATS: %s\n", in)

	bs, err := proto.Marshal(in)
	if err != nil {
		return fmt.Errorf("Error marshalling upload document: %w", err)
	}

	fmt.Printf("PubNATS: publishing\n")
	future, err := js.PublishAsync(topic, bs)
	if err != nil {
		return fmt.Errorf("Error publishing to JetStream with topic: %s\n",
			topic)
	}
	s.Logs.logger.Log("Published to JetStream - future returned: %v\n", future)

	fmt.Printf("PubNATS: published\n")

	return nil
}

func (s *Server) DocUploadStream(stream pb.Sidecar_DocUploadStreamServer) error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	topic := viper.GetString("nats.jetstream.subject")
	name := viper.GetString("nats.jetstream.name")

	var err error
	var docUpload *pb.DocUpload

	// Based on how busy NATS server is, throttle the sender
	err = s.ThrottleGRPCSender(ctx, stream)
	if err != nil {
		return fmt.Errorf("Error throttling GRPC sender stream: %w", err)
	}

	for {

		docUpload, err = stream.Recv()
		if err == io.EOF {
			fmt.Printf("DocUploadStream: Stream ended\n")
			return nil
		}

		if err != nil {
			fmt.Printf("DocUploadStream: error: %v\n", err)
			return fmt.Errorf("Error receiving from stream on server: %w", err)
		}

		fmt.Printf("DocUploadStream: Sending to NATS server\n")
		// Send document to NATS server
		s.pubNATS(topic, name, docUpload)
	}
}

func (s *Server) AddJS(ctx context.Context, in *pb.AddJSMsg) (*pb.AddJSMsgResponse, error) {

	in.Header.MsgId = NextMsgId()
	s.Logs.logger.Log("Received AddJSMsg: %s\n", in)

	var err error
	s.Subs.natsConn.js, err = NewNATSConnJS(s.Subs.natsConn.nc)
	if err != nil {

		return nil, fmt.Errorf("Error adding Jetstream: %w\n",
			err)
	}

	m, err := s.Subs.AddJS(ctx, in)
	if err != nil {
		s.Logs.logger.Log("Error subscribing: %s\n", err.Error())
		return nil, err
	}

	m.Header.MsgId = NextMsgId()
	s.Logs.logger.Log("Sending SubJSMsgRsp: %s\n", m)

	return m, err
}

func (s *Server) UnsubJS(ctx context.Context, in *pb.UnsubJSMsg) (*pb.UnsubJSMsgResponse, error) {

	in.Header.MsgId = NextMsgId()
	s.Logs.logger.Log("Received UnsubJSMsg: %s\n", in)

	m, err := s.Subs.UnsubscribeJS(s.Logs.logger, in)
	if err != nil {
		s.Logs.logger.Log("Error unsubscribing: %s\n", err.Error())
		return nil, err
	}
	m.Header.MsgId = NextMsgId()
	s.Logs.logger.Log("Sending UnsubMsgRsp: %s\n", m)

	return m, err
}

func emptySubJSTopicResponse(header *pb.Header, topic, wq string) *pb.SubJSTopicResponse {
	return &pb.SubJSTopicResponse{

		Header: &pb.Header{
			MsgType:     pb.MsgType_MSG_TYPE_SUB_JS_TOPIC_RSP,
			SrcServType: header.DstServType,
			DstServType: header.SrcServType,
			ServId:      header.ServId,
			MsgId:       NextMsgId(),
		},

		Topic:     topic,
		WorkQueue: wq,
		Msg:       []byte(""),
	}
}

func (s *Server) RecvJS(ctx context.Context, in *pb.ReceiveJS) (*pb.SubJSTopicResponse, error) {

	in.Header.MsgId = NextMsgId()
	// Do not log message to NATS. This creates a loop.
	s.Logs.logger.PrintMsg("Received from NATS: %s\n", in)

	m, err := RecvJSFromNATS(ctx, s, in)
	if err != nil {
		s.Logs.logger.Log("Could not receive from NATS: %s\n", err.Error())
		return nil, err
	}
	if m == nil {
		fmt.Printf("RecvJSFromNATS returned nil m\n")
		return emptySubJSTopicResponse(in.Header, in.Topic, in.WorkQueue), err
	}
	m.Header.MsgId = NextMsgId()
	s.Logs.logger.PrintMsg("Converted NATS message to GRPC message\n")

	return m, err
}

func (s *Server) PubJS(ctx context.Context, in *pb.PubJSMsg) (*emptypb.Empty, error) {

	in.Header.MsgId = NextMsgId()
	s.Logs.logger.Log("Received PubMsg: %s\n", in)

	future, err := s.Pubs.natsConn.js.PublishAsync(in.Topic, in.Msg)
	if err != nil {
		return &emptypb.Empty{}, fmt.Errorf("Error publishing to JetStream with topic: %s\n", in.Topic)
	}
	s.Logs.logger.Log("Published to JetStream - future returned: %v\n", future)

	return &emptypb.Empty{}, err
}
