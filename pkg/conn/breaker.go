// Code taken from Matthew Titmus' book "Cloud Native Go" with minor modifications.
// The modifications are:
//   - This code runs on the sidecar service, so that it does not have to be duplicated
//     across all services. For that reason, the Circuit function type has been changed
//     to take a message and return a message.
package conn

import (
	"context"
	"errors"
	"sync"
	"time"
)

func Breaker(circuit Circuit, failureThreshold uint) Circuit {
	// Number of failures after the first
	var consecutiveFailures int = 0

	// Time of the last interaction with the downstream service
	var lastAttempt = time.Now()

	var m sync.RWMutex

	// Construct and return the Circuit closure
	return func(ctx context.Context, msg *Message) (*Message, error) {
		m.RLock()

		d := consecutiveFailures - int(failureThreshold)

		if d >= 0 {
			shouldRetryAt := lastAttempt.Add(time.Second * 2 << d)

			if !time.Now().After(shouldRetryAt) {
				m.RUnlock()
				return &Message{}, errors.New("service unreachable")
			}
		}

		m.RUnlock()

		response, err := circuit(ctx, msg) // Issue request proper

		m.Lock() // Lock around shared resources
		defer m.Unlock()

		lastAttempt = time.Now() // Record time of attempt

		if err != nil { // Circuit returned an error,
			consecutiveFailures++ // so we count the failure
			return response, err  // and return
		}

		consecutiveFailures = 0 // Reset failures counter

		return response, nil
	}
}
