package conn

import (
	"fmt"

	"github.com/samirgadkari/sidecar/protos/v1/messages"
)

type Pubs struct {
	msgId    uint32
	natsConn *Conn
}

func InitPubs(natsConn *Conn, srv *Server) {

	srv.Pubs = &Pubs{
		msgId:    1,
		natsConn: natsConn,
	}
}

func (pubs *Pubs) Publish(in *messages.PubMsg) (*messages.PubMsgResponse, error) {

	fmt.Printf("Received PubMsg: %v\n", in)
	topic := in.GetTopic()
	data := in.GetMsg()

	pubs.natsConn.Publish(topic, data)

	srcHeader := in.GetHeader()
	header := messages.Header{
		DstServType: srcHeader.GetSrcServType(),
		SrcServType: srcHeader.GetDstServType(),
		ServId:      srcHeader.GetServId(),
		MsgId:       0,
	}

	rspHeader := messages.ResponseHeader{

		Status: uint32(messages.Status_OK),
	}

	pubMsgRsp := messages.PubMsgResponse{

		Header:    &header,
		RspHeader: &rspHeader,
		Msg:       "OK",
	}

	return &pubMsgRsp, nil
}
