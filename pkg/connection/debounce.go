// Code taken from Matthew Titmus' book "Cloud Native Go" with minor modifications.
// The modifications are:
//   - This code runs on the sidecar service, so that it does not have to be duplicated
//     across all services. For that reason, the Circuit function type has been changed
//     to take a message and return a message.
package connection

import (
	"context"
	"sync"
	"time"
)

func (msg *Message) DebounceFirst(circuit Circuit, d time.Duration) Circuit {
	var threshold time.Time
	var result *Message
	var err error
	var m sync.Mutex

	return func(ctx context.Context, msg *Message) (*Message, error) {
		m.Lock()

		defer func() {
			threshold = time.Now().Add(d)
			m.Unlock()
		}()

		if time.Now().Before(threshold) {
			return &Message{}, err
		}

		result, err = circuit(ctx, msg)

		return result, err
	}
}
