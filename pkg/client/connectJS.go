package client

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/find-in-docs/sidecar/pkg/utils"
	pb "github.com/find-in-docs/sidecar/protos/v1/messages"
)

const (
	chunkSize             = 5
	allTopicsRecvChanSize = 32
)

func (sc *SC) UploadDocs(ctx context.Context, docsCh <-chan *pb.Doc) error {

	var flow pb.StreamFlow
	stream, err := sc.Client.DocUploadStream(ctx)
	if err != nil {
		return fmt.Errorf("Error initializing document upload stream: %w", err)
	}

	utils.StartGoroutine("uploadDocsClientRecv", func() {
	LOOP:
		for {
			select {
			case <-ctx.Done():
				if ctx.Err() != nil {
					sc.Logger.Log("Done channel signaled: %v\n", err)
				}
				break LOOP
			default:
				response, err := stream.Recv()
				if err == io.EOF {
					fmt.Printf("Document upload stream ended.")
					break LOOP
				}
				if err != nil {
					fmt.Printf("Error receiving from document upload stream: %w", err)
					break LOOP
				}

				flow = response.Control.Flow
			}
		}
	})

	documents := new(pb.Documents)
	docs := make([]*pb.Doc, chunkSize)
	documents.Doc = docs
	var count int
	var msgNumber uint64
	var numOutput int

LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case doc := <-docsCh:

			fmt.Println(count, doc)
			docs[count] = doc
			if count == chunkSize-1 {

				fmt.Printf("Sending through stream\n")

				if err = stream.Send(&pb.DocUpload{
					Documents: documents,
					MsgNumber: msgNumber,
				}); err != nil {

					return fmt.Errorf("Error sending to document upload stream: %w", err)
				}

				msgNumber++

				fmt.Printf("Sent\n")
				count = 0
				numOutput++
				if numOutput == 2 {
					break LOOP
				}
			} else {

				count++
			}
		}
	}

	return nil
}

func (sc *SC) AddJS(ctx context.Context, topic, workQueue string) error {

	header := sc.header
	header.MsgType = pb.MsgType_MSG_TYPE_ADD_JS
	header.MsgId = 0

	addJSMsg := pb.AddJSMsg{
		Header:    sc.header,
		Topic:     topic,
		WorkQueue: workQueue,
	}

	addJSResp, err := sc.Client.AddJS(ctx, &addJSMsg)
	sc.Logger.Log("Add JS message sent: %s\n", addJSMsg.String())
	if err != nil {
		sc.Logger.Log("Could not add stream with topic: %s\n\t"+
			"workQueue: %s\n\terr: %v\n",
			topic, workQueue, err)
		return err
	}

	sc.Logger.Log("AddJS rsp received:\n\t%s\n", addJSResp)

	if addJSResp.RspHeader.Status != uint32(pb.Status_OK) {
		return fmt.Errorf("Error: received while adding topic: %s workQueue: %s. err: %s",
			topic, workQueue, pb.Status_name[int32(addJSResp.RspHeader.Status)])
	}

	return nil
}

func (sc *SC) UnsubJS(ctx context.Context, topic string, workQueue string) error {

	header := sc.header
	header.MsgType = pb.MsgType_MSG_TYPE_UNSUB_JS
	header.MsgId = 0

	unsubJSMsg := pb.UnsubJSMsg{
		Header:    header,
		Topic:     topic,
		WorkQueue: workQueue,
	}

	unsubJSRsp, err := sc.Client.UnsubJS(ctx, &unsubJSMsg)
	sc.Logger.Log("Unsub message sent:\n\t%s\n", &unsubJSMsg)
	if err != nil {
		sc.Logger.Log("Could not unsubscribe from topic:\n\ttopic: %s workQueue: %s %v\n",
			topic, workQueue, err)
		return err
	}

	if unsubJSRsp.RspHeader.Status != uint32(pb.Status_OK) {
		sc.Logger.Log("Error received while unsubscribing to topic:\n\ttopic: %s workQueue: %s %v",
			topic, workQueue, err)
		return err
	}

	return nil
}

type ResponseJS struct {
	Response *pb.SubJSTopicResponse
	Err      error
}

/*
func (sc *SC) ProcessSubJSMsgs(ctx context.Context, topic, workQueue string,
	chanSize uint32, f func(*pb.SubJSTopicResponse)) error {

	responseCh := sc.RecvJS(ctx, topic, workQueue)
	if responseCh == nil {
		return fmt.Errorf("Could not receive JetStream")
	}

	goroutineName := "ProcessSubJSMsgs"
	var err error
	err = utils.StartGoroutine(goroutineName,
		func() {
			subscribedTopic := topic

		LOOP:
			for {
				select {

				case r := <-responseCh:
					if r.Err != nil {
						sc.Logger.Log("Error receiving from sidecar: %v\n", r.Err)
						_ = sc.Unsub(ctx, subscribedTopic)
						break LOOP
					}

					// Do not log received message to NATS. This creates a loop.

					f(r.Response)

				case <-ctx.Done():
					_ = sc.Unsub(ctx, subscribedTopic)

					if ctx.Err() != nil {
						sc.Logger.Log("Done channel signaled: %v\n",
							ctx.Err())
					}
					break LOOP
				}
			}
			fmt.Printf("GOROUTINE 2 completed in function ProcessSubJSMsgs\n")
			utils.GoroutineEnded(goroutineName)
		})

	if err != nil {
		fmt.Printf("Error starting goroutine: %v\n", err)
		os.Exit(-1)
	}

	err = sc.SubJS(ctx, topic, workQueue, chanSize)
	if err != nil {
		return err
	}

	return nil
}
*/

func (sc *SC) RecvJS(ctx context.Context, topic string, workQueue string) <-chan *ResponseJS {

	header := sc.header
	header.MsgId = 0

	recvJSMsg := pb.ReceiveJS{
		Header:    header,
		Topic:     topic,
		WorkQueue: workQueue,
	}

	responseJSCh := make(chan *ResponseJS)

	goroutineName := "RecvJS"
	err := utils.StartGoroutine(goroutineName,
		func() {
		LOOP:
			for {
				select {
				case <-ctx.Done():
					fmt.Printf(">>>> Context done for goroutine: %s. err: %v\n",
						goroutineName, ctx.Err())

					break LOOP
				default:
					// fmt.Printf(">>>> Calling sc.Client.RecvJS\n")
					subJSTopicRsp, err := sc.Client.RecvJS(ctx, &recvJSMsg)
					if err != nil {
						sc.Logger.Log("Could not receive from sidecar - err: %v\n", err)
						break LOOP
					}

					// Do not log received message to NATS. This creates a loop.
					// fmt.Printf(">>>> Got 1 message\n")
					responseJSCh <- &ResponseJS{
						subJSTopicRsp,
						nil,
					}
					// time.Sleep(time.Second)
				}
			}
			fmt.Printf("GOROUTINE 1 completed in function Recv\n")
			utils.GoroutineEnded(goroutineName)
		})

	if err != nil {
		fmt.Printf("Error starting goroutine: %v\n", err)
		os.Exit(-1)
	}

	return responseJSCh
}
