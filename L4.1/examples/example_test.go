package main

import (
	"fmt"
	"time"

	"or-channel"
)

// ExampleOr demonstrates the basic usage of the Or function
func ExampleOr() {
	sig := func(after time.Duration) <-chan any {
		c := make(chan any)
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	start := time.Now()
	<-or.Or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)

	fmt.Printf("done after %v", time.Since(start))
	// Output will be approximately: done after 1s
}

// ExampleOr_contextCancellation demonstrates using Or for context cancellation patterns
func ExampleOr_contextCancellation() {
	// Simulate a user cancellation signal
	userCancel := make(chan any)
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(userCancel)
	}()

	// Simulate a timeout
	timeout := func(d time.Duration) <-chan any {
		c := make(chan any)
		go func() {
			defer close(c)
			time.Sleep(d)
		}()
		return c
	}

	start := time.Now()
	<-or.Or(
		userCancel,             // User clicks "cancel" button
		timeout(5*time.Second), // Operation timeout
	)

	duration := time.Since(start)
	if duration < 200*time.Millisecond {
		fmt.Println("Operation cancelled by user")
	} else {
		fmt.Println("Operation timed out")
	}
	// Output: Operation cancelled by user
}

// ExampleOr_noChannels demonstrates Or with no input channels
func ExampleOr_noChannels() {
	start := time.Now()
	<-or.Or() // Returns immediately with a closed channel
	duration := time.Since(start)

	if duration < 10*time.Millisecond {
		fmt.Println("Completed immediately")
	}
	// Output: Completed immediately
}

// ExampleOr_singleChannel demonstrates Or with a single channel
func ExampleOr_singleChannel() {
	sig := func(after time.Duration) <-chan any {
		c := make(chan any)
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	start := time.Now()
	<-or.Or(sig(50 * time.Millisecond))
	duration := time.Since(start)

	if duration >= 40*time.Millisecond && duration <= 100*time.Millisecond {
		fmt.Println("Completed after approximately 50ms")
	}
	// Output: Completed after approximately 50ms
}

// ExampleOr_multipleServices demonstrates using Or to wait for the first service to respond
func ExampleOr_multipleServices() {
	// Simulate calling multiple redundant services and using the fastest response
	callService := func(_ string, latency time.Duration) <-chan any {
		c := make(chan any)
		go func() {
			defer close(c)
			time.Sleep(latency)
			// In a real scenario, you'd make an actual service call here
		}()
		return c
	}

	start := time.Now()
	<-or.Or(
		callService("service-us-west", 200*time.Millisecond),
		callService("service-us-east", 50*time.Millisecond), // Fastest
		callService("service-eu", 300*time.Millisecond),
		callService("service-asia", 400*time.Millisecond),
	)

	duration := time.Since(start)
	if duration >= 40*time.Millisecond && duration <= 100*time.Millisecond {
		fmt.Println("Got response from fastest service")
	}
	// Output: Got response from fastest service
}
