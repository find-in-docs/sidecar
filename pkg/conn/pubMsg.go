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

	topic := in.GetTopic()
	data := in.GetMsg()

	pubs.natsConn.Publish(topic, data)

	pubMsgRsp := pb.PubMsgResponse{
		Header: &pb.Header{
			MsgType:     pb.MsgType_MSG_TYPE_PUB_RSP,
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

	return &pubMsgRsp, nil
}

func PrintPubMsg(prefix string, m *pb.PubMsg) {

	fmt.Printf("%s\n\tHeader: %s\n\tTopic: %s\n\t:Msg: %s\n",
		m.Header, m.Topic, m.Msg)
}

func PrintPubMsgRsp(prefix string, m *pb.PubMsgResponse) {

	fmt.Printf("%s\n\tHeader: %s\n\tRspHeader: %s\n\t:Msg: %s\n",
		m.Header, m.RspHeader, m.Msg)
}
