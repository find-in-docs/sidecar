package conn

import (
	"fmt"

	pb "github.com/samirgadkari/sidecar/protos/v1/messages"
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

func (pubs *Pubs) Publish(in *pb.PubMsg) (*pb.PubMsgResponse, error) {

	fmt.Printf("Received PubMsg: %v\n", in)
	topic := in.GetTopic()
	data := in.GetMsg()

	pubs.natsConn.Publish(topic, data)

	srcHeader := in.GetHeader()
	header := pb.Header{
		MsgType:     pb.MsgType_MSG_TYPE_PUB_RSP,
		DstServType: srcHeader.GetSrcServType(),
		SrcServType: srcHeader.GetDstServType(),
		ServId:      srcHeader.GetServId(),
		MsgId:       NextMsgId(),
	}

	rspHeader := pb.ResponseHeader{

		Status: uint32(pb.Status_OK),
	}

	pubMsgRsp := pb.PubMsgResponse{

		Header:    &header,
		RspHeader: &rspHeader,
		Msg:       "OK",
	}

	return &pubMsgRsp, nil
}
