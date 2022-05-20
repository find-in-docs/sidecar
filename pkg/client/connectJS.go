package client

import (
	"context"
	"fmt"
	"os"

	"github.com/find-in-docs/sidecar/pkg/utils"
	pb "github.com/find-in-docs/sidecar/protos/v1/messages"
)

func (sc *SC) PubJS(ctx context.Context, topic string, workQueue string, data []byte) error {

	header := sc.header
	header.MsgType = pb.MsgType_MSG_TYPE_PUB_JS
	header.MsgId = 0

	pubJSMsg := pb.PubJSMsg{
		Header:    sc.header,
		Topic:     topic,
		WorkQueue: workQueue,
		Msg:       data,
	}

	_, err := sc.client.PubJS(ctx, &pubJSMsg)
	sc.Logger.Log("Pub JS message sent: %s\n", pubJSMsg.String())
	if err != nil {
		sc.Logger.Log("Could not publish to topic: %s\n\tworkQueue: %s\n\tmessage:\n\tmsg: %s %v\n",
			topic, workQueue, string(data), err)
		return err
	}

	return nil
}

func (sc *SC) SubJS(ctx context.Context, topic string, workQueue string, chanSize uint32) error {

	header := sc.header
	header.MsgType = pb.MsgType_MSG_TYPE_SUB_JS
	header.MsgId = 0

	subJSMsg := pb.SubJSMsg{
		Header:    header,
		Topic:     topic,
		WorkQueue: workQueue,
		ChanSize:  chanSize,
	}

	subJSRsp, err := sc.client.SubJS(ctx, &subJSMsg)
	sc.Logger.Log("Sub message sent:\n\t%s\n", &subJSMsg)
	if err != nil {
		sc.Logger.Log("Could not subscribe to topic: %s workQueue: %s %v\n",
			topic, workQueue, err)
		return err
	}
	sc.Logger.Log("SubJS rsp received:\n\t%s\n", subJSRsp)

	if subJSRsp.RspHeader.Status != uint32(pb.Status_OK) {
		sc.Logger.Log("Error received while publishing to topic:\n\ttopic: %s workQueue: %s %v\n",
			topic, workQueue, err)
		return err
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

	unsubJSRsp, err := sc.client.UnsubJS(ctx, &unsubJSMsg)
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
	response *pb.SubJSTopicResponse
	err      error
}

func (sc *SC) ProcessSubJSMsgs(ctx context.Context, topic, workQueue string,
	chanSize uint32, f func(*pb.SubJSTopicResponse)) error {

	responseCh := sc.RecvJS(ctx, topic, workQueue)

	goroutineName := "ProcessSubJSMsgs"
	var err Error
	err = utils.StartGoroutine(goroutineName,
		func() {
			subscribedTopic := topic

		LOOP:
			for {
				select {

				case r := <-responseCh:
					if r.err != nil {
						sc.Logger.Log("Error receiving from sidecar: %v\n", r.err)
						_ = sc.Unsub(ctx, subscribedTopic)
						break LOOP
					}

					// Do not log received message to NATS. This creates a loop.

					f(r.response)

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

	err := sc.SubJS(ctx, topic, workQueue, chanSize)
	if err != nil {
		return err
	}

	return nil
}

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
				subJSTopicRsp, err := sc.client.RecvJS(ctx, &recvJSMsg)
				if err != nil {
					sc.Logger.Log("Could not receive from sidecar - err: %v\n", err)
					break LOOP
				}

				// Do not log received message to NATS. This creates a loop.

				responseJSCh <- &ResponseJS{
					subJSTopicRsp,
					nil,
				}

				if ctx.Err() != nil {
					break LOOP
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
