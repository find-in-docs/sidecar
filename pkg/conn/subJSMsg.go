package conn

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/find-in-docs/sidecar/pkg/log"
	"github.com/find-in-docs/sidecar/pkg/utils"
	pb "github.com/find-in-docs/sidecar/protos/v1/messages"
	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/proto"
)

func unsubscribeJS(subs *Subs, topic string) {
	subs.subscriptionsJS[topic].Drain()
	subs.subscriptionsJS[topic].Unsubscribe()
	delete(subs.subscriptionsJS, topic)

	// close(subs.natsJSMsgs[topic])
	// delete(subs.natsJSMsgs, topic)
}

func (subs *Subs) DownloadJS(stream pb.Sidecar_DocDownloadStreamServer) error {

	ctx := stream.Context()

	var err error

	var flow pb.StreamFlow

	fmt.Printf("In DownloadJS\n")
	defer fmt.Printf("Exiting DownloadJS\n")

	err = utils.StartGoroutine("downloadDocsThrottle", func() {
	LOOP:
		for {
			select {
			case <-ctx.Done():
				if ctx.Err() != nil {
					fmt.Errorf("Done channel signaled: %v\n", err)
				}
				break LOOP
			default:
				response, err := stream.Recv()
				if err == io.EOF {
					fmt.Printf("Document download stream ended.\n")
					break LOOP
				}
				if err != nil {
					fmt.Printf("Error during receive from document download stream: %v\n", err)
					break LOOP
				}

				flow = response.Control.Flow
				if flow == pb.StreamFlow_ON {
					fmt.Printf("^")
				} else {
					fmt.Printf("v")
				}
			}
		}
	})

	if err != nil {
		return fmt.Errorf("Error starting goroutine downloadDocsThrottle: %w", err)
	}

	// Fetch messages in batches here
	numMsgsToFetch := viper.GetInt("nats.jetstream.fetch.numMsgs")
	maxWaitStr := viper.GetString("nats.jetstream.fetch.timeoutInSecs")
	flowControlTimeoutInNs, err := time.ParseDuration(viper.GetString("nats.jetstream.flowControlTimeoutInNs"))
	if err != nil {
		return fmt.Errorf("Error parsing flow control timeout: %w", err)
	}

	maxWait, err := time.ParseDuration(maxWaitStr)
	if err != nil {
		return fmt.Errorf("Error parsing NATS max wait timeout value: %w", err)
	}
	natsMaxWait := nats.MaxWait(maxWait)

	topic := subs.currentStreamTopic
	subscription := subs.subscriptionsJS[topic]
	fmt.Printf("topic: %s\n", topic)

LOOP:
	for {
		select {
		case <-ctx.Done():

			if ctx.Err() != nil {
				fmt.Printf("Done channel signaled: err: %v\n",
					ctx.Err())
			}

			unsubscribeJS(subs, subs.currentStreamTopic)
			break LOOP
		default:
			for flow == pb.StreamFlow_OFF {
				time.Sleep(flowControlTimeoutInNs)
			}
		}

		ms, err := subscription.Fetch(numMsgsToFetch, natsMaxWait)
		if err != nil {
			fmt.Printf("Error fetching from topic: %s\n\terr: %v\n",
				topic, err)
			break LOOP
		}
		if ms == nil {
			continue
		}

		for _, m := range ms {
			m.Ack()

			var docDownload pb.DocDownload

			err := proto.Unmarshal(m.Data, &docDownload)
			if err != nil {
				fmt.Printf("Error unmarshalling download document: %v\n", err)
				break LOOP
			}

			if err = stream.Send(&pb.DocDownload{
				Documents: docDownload.Documents,
				MsgNumber: docDownload.MsgNumber,
			}); err != nil {

				fmt.Printf("Error sending to document download stream: %v\n", err)
				break LOOP
			}
			fmt.Printf("<")
		}
	}

	return nil
}

func (subs *Subs) AddJS(ctx context.Context, in *pb.AddJSMsg) (*pb.AddJSMsgResponse, error) {

	topic := in.GetTopic()
	workQueue := in.GetWorkQueue()
	chanSize := viper.GetInt("nats.jetstream.goroutineChanSize")

	topicMsgs := make(chan *nats.Msg, chanSize)
	subs.natsJSMsgs[topic] = topicMsgs

	subscription, err := subs.natsConn.js.PullSubscribe(topic, workQueue,
		// nats.PullMaxWaiting(512),
		nats.ManualAck(),
		nats.DeliverAll(),
		// nats.MaxDeliver(32),
		nats.BindStream("uploadDocs"),
	)
	if err != nil {
		return nil, fmt.Errorf("Could not subscribe: topic: %s workQueue: %s: %w",
			topic, workQueue, err)
	}

	subs.subscriptionsJS[topic] = subscription
	subs.currentStreamTopic = topic

	subJSMsgRsp := &pb.AddJSMsgResponse{
		Header: &pb.Header{
			MsgType:     pb.MsgType_MSG_TYPE_SUB_JS_RSP,
			SrcServType: "sidecar",
			DstServType: in.Header.SrcServType,
			ServId:      serviceId()(),
			MsgId:       NextMsgId(),
		},

		RspHeader: &pb.ResponseHeader{
			Status: uint32(pb.Status_OK),
		},

		Msg: "OK",
	}

	return subJSMsgRsp, nil
}

func (subs *Subs) UnsubscribeJS(logger *log.Logger, in *pb.UnsubJSMsg) (*pb.UnsubJSMsgResponse, error) {

	topic := in.GetTopic()
	workQueue := in.GetWorkQueue()

	if _, ok := subs.subscriptionsJS[topic]; !ok {
		return nil, fmt.Errorf("Error - topic not found to unsubscribe:\n\ttopic: %s\n\tworkQueue: %s\n",
			topic, workQueue)
	}

	unsubscribeJS(subs, topic)

	unsubJSMsgRsp := &pb.UnsubJSMsgResponse{
		Header: &pb.Header{
			MsgType:     pb.MsgType_MSG_TYPE_UNSUB_JS_RSP,
			SrcServType: serviceType(),
			DstServType: in.Header.SrcServType,
			ServId:      serviceId()(),
			MsgId:       NextMsgId(),
		},

		RspHeader: &pb.ResponseHeader{
			Status: uint32(pb.Status_OK),
		},

		Msg: "OK",
	}

	logger.Log("Successfully unsubscribed from topic: %s\n\tworkQueue: %s\n", topic, workQueue)

	return unsubJSMsgRsp, nil
}
