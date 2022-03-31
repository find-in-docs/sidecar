package conn

import "sync"

// GRPC creates a different goroutine to handle each request.
// This value is used by many goroutines.
// This means we have to keep it as a global, and protect it using a mutex.
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
