package conn

type ServiceTyp uint32
type ServiceIdTyp uint64
type RequestIdTyp uint64
type MessageIdTyp uint32

type Header struct {
	ServType ServiceTyp
	ServId   ServiceIdTyp
	ReqId    RequestIdTyp
	MsgId    MessageIdTyp
}

type Message struct {
	header Header
	data   []byte
}
