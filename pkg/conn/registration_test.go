package conn

import (
	"testing"
	"time"

	"github.com/samirgadkari/sidecar/pkg/client"
	"github.com/samirgadkari/sidecar/pkg/config"
	pb "github.com/samirgadkari/sidecar/protos/v1/messages"
	"google.golang.org/protobuf/types/known/durationpb"
)

func InitSC(t *testing.T) *client.SC {

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

	return client.InitSidecar("testing", rParams)
}

func TestRegsitration(t *testing.T) {

	sidecar := InitSC(t)

	if sidecar == nil {
		t.Errorf("Error initializing sidecar - Exiting\n")
	}
	t.Logf("Success !")
}

func TestPublish(t *testing.T) {

	sidecar := InitSC(t)

	if sidecar == nil {
		t.Errorf("Error initializing sidecar - Exiting\n")
	}

	err := sidecar.Pub("search.testdata.v1", []byte("Test data"), nil)
	if err != nil {
		t.Errorf("Error publishing data without retries.\n\terr: %v\n", err)
	}

	var retryNum uint32 = 9
	retryDelayDuration, err := time.ParseDuration("9s")
	if err != nil {
		t.Errorf("Error creating Golang time duration.\nerr: %v\n", err)
	}
	retryDelay := durationpb.New(retryDelayDuration)

	err = sidecar.Pub("search.testdata.v1", []byte("Test data with retries"),
		&pb.RetryBehavior{
			RetryNum:   &retryNum,
			RetryDelay: retryDelay,
		})
	if err != nil {
		t.Errorf("Error publishing data with retries.\n\terr: %v\n", err)
	}
}
