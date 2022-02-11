package connection

const (
	maxInputs = 1000
)

var MsgFunc map[MessageIdTyp][]interface{}

func AddInterface(msgId MessageIdTyp, i interface{}) {
	MsgFunc[msgId] = append(MsgFunc[msgId], i)
}

/*
func InputPipeline() error {

	inputs := make(chan Message, maxInputs)
	go func() {
		for {
			select {
			case msg <- inputs:
				fns, ok := MsgFunc[msg.MsgId]

				// If the service registered requirements for
				// this message type, then enforce them.
				if ok {

					for _, fnName := range fns {
*/
