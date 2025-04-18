package main

import (
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
)

// Function that leaks goroutines - it starts a goroutine that never exits
func leakyFunction() {
	go func() {
		for {
			time.Sleep(time.Second)
			// This goroutine never exits
		}
	}()
}

// Function that doesn't leak - goroutine properly exits
func nonLeakyFunction() {
	go func() {
		time.Sleep(100 * time.Millisecond)
		// This goroutine exits after sleeping
	}()
}

func TestLeakyFunction(t *testing.T) {
	defer leaktest.Check(t)() // This will detect goroutine leaks

	leakyFunction()
	// The test will fail because leakyFunction leaves a goroutine running
}

func TestNonLeakyFunction(t *testing.T) {
	defer leaktest.CheckTimeout(t, 200*time.Millisecond)() // Give time for goroutine to exit

	nonLeakyFunction()
	// This test will pass because the goroutine exits
}

// Another example with channel leak
func channelLeaker() chan int {
	ch := make(chan int)
	go func() {
		// This goroutine blocks forever waiting on the channel
		<-ch
	}()
	return ch // Return channel but nobody will send to it
}

func TestChannelLeak(t *testing.T) {
	defer leaktest.Check(t)()

	_ = channelLeaker()
	// Test will fail because of the blocked goroutine
}
