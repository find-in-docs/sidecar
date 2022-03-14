package conn

import (
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
)

func createServiceId() []byte {
	uuid := uuid.New()
	servId, err := uuid.MarshalText()
	if err != nil {
		fmt.Printf("Error converting UUID: %v to text\n\terr: %v\n", servId, err)
		os.Exit(-1)
	}

	return servId
}

func serviceId() func() []byte {

	var servId []byte

	return func() []byte {
		var once sync.Once
		once.Do(func() {
			servId = createServiceId()
		})

		return servId
	}
}

func serviceType() string {

	return "sidecar"
}
