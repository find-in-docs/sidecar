package conn

import "sync"

// This value is used in other places than just within the server context.
// This means we have to keep it as a global.
type MsgID struct {
	mu    sync.Mutex
	Value uint64
}

var msgId MsgID

func NextMsgId() uint64 {

	msgId.mu.Lock()
	defer msgId.mu.Unlock()

	msgId.Value += 1
	return msgId.Value
}
