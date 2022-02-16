package conn

import (
	"testing"
	"time"

	"protos/go/messagespb"

	"google.golang.org/protobuf/types/known/durationpb"
)

func TestRegsitration(t *testing.T) {

	debounceDur, err := time.ParseDuration("5s")
	if err != nil {
		t.Errorf("Failed: Cannot parse debounce duration: err: %v\n", err)
	}

	retryDelay, err := time.ParseDuration("2s")
	if err != nil {
		t.Errorf("Failed: Cannot parse retry duration: err: %v\n", err)
	}

	rMsg := messagespb.RegistrationMsg{
		messagespb.RegistrationMsg.Header: {
			ServType: messagespb.IMPORT,
			ServId:   100,
			RegId:    200,
			MsgId:    300,
		},

		messagespb.Limit.Circuit: {
			FailureThreshold: 3,
		},
		messagespb.Limit.Debounce: {
			Delay: durationpb.New(debounceDur),
		},
		messagespb.Limit.Retry: {
			Retries: 2,
			Delay:   durationpb.New(retryDelay),
		},
	}

	if err := conn.RegisterMsg(&rMsg); err == nil {
		t.Errorf("Failed. RegisterMsg returned err: %v\n", err)
	}

	t.Logf("Success !")
}
