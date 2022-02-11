// Code taken from Matthew Titmus' book "Cloud Native Go" with minor modifications.
// The modifications are:
//   - This code runs on the sidecar service, so that it does not have to be duplicated
//     across all services. For that reason, the Circuit function type has been changed
//     to take a message and return a message.
package connection

import (
	"context"
	"log"
	"time"
)

func Retry(effector Effector, retries int, delay time.Duration) Effector {

	return func(ctx context.Context) (*Message, error) {

		for r := 0; ; r++ {

			response, err := effector(ctx)
			if err == nil || r >= retries {
				return response, err
			}

			log.Printf("Attempt %d failed; retrying in %v", r+1, delay)

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return &Message{}, ctx.Err()
			}
		}
	}
}
