package conn

import (
	"fmt"
	"time"

	pb "github.com/find-in-docs/sidecar/protos/v1/messages"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
)

func NewRegParams(circuitFailureThreshold uint32,
	debounceDelay string,
	retryNum uint32,
	retryDelay string) (*pb.RegistrationParams, error) {

	debounceDelayDuration, err := time.ParseDuration(debounceDelay)
	if err != nil {
		return nil, fmt.Errorf("Could not parse debounce delay: %w\n", err)
	}

	retryDelayDuration, err := time.ParseDuration(retryDelay)
	if err != nil {
		return nil, fmt.Errorf("Could not parse retry delay: %w\n", err)
	}

	retryBehavior := pb.RetryBehavior{
		RetryNum:   &retryNum,
		RetryDelay: durationpb.New(retryDelayDuration),
	}

	return &pb.RegistrationParams{
		CircuitFailureThreshold: &circuitFailureThreshold,
		DebounceDelay:           durationpb.New(debounceDelayDuration),
		Retry:                   &retryBehavior,
	}, nil
}
