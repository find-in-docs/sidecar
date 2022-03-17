package conn

import (
	"testing"
	"time"

	pb "github.com/samirgadkari/sidecar/protos/v1/messages"
	"google.golang.org/protobuf/types/known/durationpb"
)

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
