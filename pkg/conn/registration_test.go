package conn

import (
	"testing"
	"time"

	"github.com/samirgadkari/sidecar/pkg/client"
	"github.com/samirgadkari/sidecar/pkg/config"
	pb "github.com/samirgadkari/sidecar/protos/v1/messages"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestRegsitration(t *testing.T) {

	config.Load()

	var circuitConsecutiveFailures uint32 = 3

	// Go Duration is in the time package: https://pkg.go.dev/time#Duration
	// Go Duration maps to protobuf Duration.
	// You can convert between them using durationpb:
	//   https://pkg.go.dev/google.golang.org/protobuf/types/known/durationpb
	debounceDelayDuration, err := time.ParseDuration("5s")
	if err != nil {
		t.Errorf("Error creating Golang time duration.\nerr: %v\n", err)
	}
	debounceDelay := durationpb.New(debounceDelayDuration)

	var retryNum uint32 = 2

	retryDelayDuration, err := time.ParseDuration("2s")
	if err != nil {
		t.Errorf("Error creating Golang time duration.\nerr: %v\n", err)
	}
	retryDelay := durationpb.New(retryDelayDuration)

	rParams := &pb.RegistrationParams{
		CircuitFailureThreshold: &circuitConsecutiveFailures,
		DebounceDelay:           debounceDelay,

		Retry: &pb.RetryBehavior{
			RetryNum:   &retryNum,
			RetryDelay: retryDelay,
		},
	}

	_ = client.InitSidecar("testing", rParams)

	t.Logf("Success !")
}
