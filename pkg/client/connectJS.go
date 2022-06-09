package client

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/find-in-docs/sidecar/pkg/utils"
	pb "github.com/find-in-docs/sidecar/protos/v1/messages"
	"github.com/spf13/viper"
)

const (
	chunkSize             = 5
	allTopicsRecvChanSize = 32
)

func (sc *SC) ReceiveDocs(ctx context.Context, subject, durableName string) (chan *pb.DocDownload, error) {

	err := sc.AddJS(ctx, subject, durableName)
	if err != nil {
		return nil, fmt.Errorf("Error adding jetstream: %w", err)
	}

	recvChanSize := viper.GetInt("nats.jetstream.recvChanSize")
	recvDocs := make(chan *pb.DocDownload, recvChanSize)

	stream, err := sc.Client.DocDownloadStream(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error initializing document upload stream: %w", err)
	}

	utils.StartGoroutine("downloadDocsClientRecv", func() {
	LOOP:
		for {
			select {
			case <-ctx.Done():
				if ctx.Err() != nil {
					sc.Logger.Log("Done channel signaled: %v\n", err)
				}
				break LOOP
			default:
				docsDownload, err := stream.Recv()
				if err == io.EOF {
					fmt.Printf("Document download stream ended.")
					break LOOP
				}
				if err != nil {
					fmt.Printf("Error receiving from document download stream: %v\n", err)
					break LOOP
				}

				recvDocs <- docsDownload
			}
		}
	})

	utils.StartGoroutine("downloadDocsClientRecv", func() {
	LOOP2:
		for {
			select {
			case <-ctx.Done():
				if ctx.Err() != nil {
					sc.Logger.Log("Done channel signaled: %v\n", err)
				}
				break LOOP2
			case <-time.After(time.Second):

				percentChannelUsed := int((cap(recvDocs) - len(recvDocs)) / (cap(recvDocs) * 100))
				// fmt.Printf("cap: %d, l: %d, percentChannelUsed: %d\n",
				//	c, l, percentChannelUsed)

				if percentChannelUsed > 50 {

					if err = stream.Send(&pb.DocDownloadResponse{
						Control: &pb.StreamControl{
							Flow: pb.StreamFlow_OFF,
						},

						AckMsgNumber: 0,
					}); err != nil {

						break LOOP2
					}
				} else {
					if err = stream.Send(&pb.DocDownloadResponse{
						Control: &pb.StreamControl{
							Flow: pb.StreamFlow_ON,
						},

						AckMsgNumber: 0,
					}); err != nil {

						break LOOP2
					}
				}
			}
		}
	})

	return recvDocs, nil
}

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
				fmt.Printf(">>>> Flow Control: %s\n", pb.StreamFlow_name[int32(flow)])
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
