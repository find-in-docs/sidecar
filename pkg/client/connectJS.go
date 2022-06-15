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

func (sc *SC) ReceiveDocs(ctx context.Context, subject, durableName string) (chan *pb.DocDownload, error) {

	err := sc.AddJS(ctx, subject, durableName)
	if err != nil {
		return nil, fmt.Errorf("Error adding jetstream: %w", err)
	}

	flowControlTimeoutInNs, err := time.ParseDuration(viper.GetString("nats.jetstream.flowControlTimeoutInNs"))
	if err != nil {
		return nil, fmt.Errorf("Error parsing flow control timeout: %w", err)
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

				fmt.Printf("<")
				recvDocs <- docsDownload
			}
		}
	})

	utils.StartGoroutine("downloadDocsClientSend", func() {
	LOOP2:
		for {
			select {
			case <-ctx.Done():
				if ctx.Err() != nil {
					sc.Logger.Log("Done channel signaled: %v\n", err)
				}
				break LOOP2
			case <-time.After(flowControlTimeoutInNs):

				percentChannelUsed := int((cap(recvDocs) - len(recvDocs)) / (cap(recvDocs) * 100))
				// fmt.Printf("cap: %d, l: %d, percentChannelUsed: %d: ",
				//	cap(recvDocs), len(recvDocs), percentChannelUsed)

				if percentChannelUsed > 50 {

					fmt.Printf("v")
					if err = stream.Send(&pb.DocDownloadResponse{
						Control: &pb.StreamControl{
							Flow: pb.StreamFlow_OFF,
						},

						AckMsgNumber: 0,
					}); err != nil {

						break LOOP2
					}
				} else {
					fmt.Printf("^")
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
				if flow == pb.StreamFlow_ON {
					fmt.Printf("^")
				} else {
					fmt.Printf("v")
				}
			}
		}
	})

	documents := new(pb.Documents)
	chunkSize := viper.GetInt("nats.jetstream.msgChunkSize")
	fmt.Printf("chunkSize: %d\n", chunkSize)
	docs := make([]*pb.Doc, chunkSize)
	documents.Doc = docs
	var count int
	var msgNumber uint64
	var numOutput int

	flowControlTimeoutInNs, err := time.ParseDuration(viper.GetString("nats.jetstream.flowControlTimeoutInNs"))
	if err != nil {
		return fmt.Errorf("Error parsing flow control timeout: %w", err)
	}

LOOP2:
	for {
		select {
		case <-ctx.Done():
			break LOOP2
		case doc, ok := <-docsCh:

			// Channel closed
			if !ok {
				break LOOP2
			}

			for flow == pb.StreamFlow_OFF {
				time.Sleep(flowControlTimeoutInNs)
			}

			fmt.Println(count, doc)
			docs[count] = doc
			if count == chunkSize-1 {

				if err = stream.Send(&pb.DocUpload{
					Documents: documents,
					MsgNumber: msgNumber,
				}); err != nil {

					return fmt.Errorf("Error sending to document upload stream: %w", err)
				}
				fmt.Printf(">")

				msgNumber++

				count = 0
				numOutput++
				if numOutput == 2000 {
					break LOOP2
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
