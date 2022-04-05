package utils

import (
	"fmt"
	"sync"
)

const (
	maxNumGoroutines = 32
)

type goroutineData struct {
	running map[string]struct{}
}

var goroutines goroutineData

func StartGoroutine(name string, f func()) error {

	var once sync.Once
	once.Do(func() {
		goroutines.running = make(map[string]struct{}, maxNumGoroutines)
	})

	if len(goroutines.running) >= maxNumGoroutines {
		return fmt.Errorf("Cannot start GOROUTINE - too many goroutines\n")
	}
	go f()

	goroutines.running[name] = struct{}{}
	fmt.Printf("GOROUTINE ADDED: %s\n", name)

	return nil
}

func GoroutineEnded(name string) {

	if goroutines.running == nil {
		fmt.Printf("ERROR - goroutine handler not initialized\n")
		return
	}

	delete(goroutines.running, name)
	fmt.Printf("GOROUTINE ENDED: %s\n", name)
}

func ListGoroutinesRunning() {

	fmt.Printf("Num goroutines running: %d\n", len(goroutines.running))
	for k := range goroutines.running {
		fmt.Printf("GOROUTINE running: %s\n", k)
	}
}
