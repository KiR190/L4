package or

import (
	"testing"
	"time"
)

// Helper function to create a channel that closes after a duration
func afterDuration(d time.Duration) <-chan any {
	c := make(chan any)
	go func() {
		defer close(c)
		time.Sleep(d)
	}()
	return c
}

// Helper function to create an immediately closed channel
func closedChannel() <-chan any {
	c := make(chan any)
	close(c)
	return c
}

// Helper function to create a channel that never closes (for timeout tests)
func neverClosingChannel() <-chan any {
	return make(chan any)
}

// TestOrNoChannels tests Or with zero channels
func TestOrNoChannels(t *testing.T) {
	start := time.Now()
	<-Or()
	duration := time.Since(start)

	// Should return immediately (already closed channel)
	if duration > 10*time.Millisecond {
		t.Errorf("Or() with no channels took too long: %v", duration)
	}
}

// TestOrSingleChannel tests Or with a single channel
func TestOrSingleChannel(t *testing.T) {
	start := time.Now()
	<-Or(afterDuration(50 * time.Millisecond))
	duration := time.Since(start)

	// Should close after approximately 50ms
	if duration < 40*time.Millisecond || duration > 100*time.Millisecond {
		t.Errorf("Or() with single channel took unexpected time: %v (expected ~50ms)", duration)
	}
}

// TestOrTwoChannels tests Or with two channels
func TestOrTwoChannels(t *testing.T) {
	start := time.Now()
	<-Or(
		afterDuration(100*time.Millisecond),
		afterDuration(50*time.Millisecond), // This one closes first
	)
	duration := time.Since(start)

	// Should close after approximately 50ms (the faster one)
	if duration < 40*time.Millisecond || duration > 100*time.Millisecond {
		t.Errorf("Or() with two channels took unexpected time: %v (expected ~50ms)", duration)
	}
}

// TestOrMultipleChannels tests Or with multiple channels
func TestOrMultipleChannels(t *testing.T) {
	start := time.Now()
	<-Or(
		afterDuration(2*time.Second),
		afterDuration(500*time.Millisecond),
		afterDuration(100*time.Millisecond), // This one closes first
		afterDuration(1*time.Second),
		afterDuration(300*time.Millisecond),
	)
	duration := time.Since(start)

	// Should close after approximately 100ms (the fastest one)
	if duration < 90*time.Millisecond || duration > 200*time.Millisecond {
		t.Errorf("Or() with multiple channels took unexpected time: %v (expected ~100ms)", duration)
	}
}

// TestOrWithAlreadyClosedChannel tests Or when one channel is already closed
func TestOrWithAlreadyClosedChannel(t *testing.T) {
	start := time.Now()
	<-Or(
		afterDuration(1*time.Second),
		closedChannel(), // Already closed
		afterDuration(2*time.Second),
	)
	duration := time.Since(start)

	// Should return almost immediately
	if duration > 50*time.Millisecond {
		t.Errorf("Or() with already-closed channel took too long: %v (expected immediate)", duration)
	}
}

// TestOrDoesNotBlockOnSlowChannels tests that Or doesn't wait for slow channels
func TestOrDoesNotBlockOnSlowChannels(t *testing.T) {
	start := time.Now()
	<-Or(
		neverClosingChannel(),              // Never closes
		afterDuration(50*time.Millisecond), // Fast channel
		neverClosingChannel(),              // Never closes
	)
	duration := time.Since(start)

	// Should close after approximately 50ms, not wait forever
	if duration < 40*time.Millisecond || duration > 100*time.Millisecond {
		t.Errorf("Or() blocked on slow channels: %v (expected ~50ms)", duration)
	}
}

// TestOrManyChannels tests Or with a large number of channels
func TestOrManyChannels(t *testing.T) {
	channels := make([]<-chan any, 100)

	// Create 99 slow channels
	for i := 0; i < 99; i++ {
		channels[i] = afterDuration(10 * time.Second)
	}

	// One fast channel
	channels[99] = afterDuration(50 * time.Millisecond)

	start := time.Now()
	<-Or(channels...)
	duration := time.Since(start)

	// Should close after approximately 50ms (the fastest one)
	if duration < 40*time.Millisecond || duration > 150*time.Millisecond {
		t.Errorf("Or() with 100 channels took unexpected time: %v (expected ~50ms)", duration)
	}
}

// TestOrReturnsClosedChannel verifies the returned channel is actually closed
func TestOrReturnsClosedChannel(t *testing.T) {
	done := Or(closedChannel())

	// Try to read from it - should not block
	select {
	case <-done:
		// Good, channel is closed
	case <-time.After(100 * time.Millisecond):
		t.Error("Or() returned channel is not closed")
	}

	// Try reading again - should still work (closed channels can be read multiple times)
	select {
	case <-done:
		// Good, channel is still readable
	case <-time.After(100 * time.Millisecond):
		t.Error("Or() returned channel became blocking on second read")
	}
}

// BenchmarkOr2Channels benchmarks Or with 2 channels
func BenchmarkOr2Channels(b *testing.B) {
	for i := 0; i < b.N; i++ {
		<-Or(
			afterDuration(1*time.Millisecond),
			afterDuration(2*time.Millisecond),
		)
	}
}

// BenchmarkOr10Channels benchmarks Or with 10 channels
func BenchmarkOr10Channels(b *testing.B) {
	for i := 0; i < b.N; i++ {
		channels := make([]<-chan any, 10)
		for j := 0; j < 10; j++ {
			channels[j] = afterDuration(time.Duration(j+1) * time.Millisecond)
		}
		<-Or(channels...)
	}
}

// BenchmarkOr100Channels benchmarks Or with 100 channels
func BenchmarkOr100Channels(b *testing.B) {
	for i := 0; i < b.N; i++ {
		channels := make([]<-chan any, 100)
		channels[0] = afterDuration(1 * time.Millisecond)
		for j := 1; j < 100; j++ {
			channels[j] = afterDuration(time.Duration(j+1) * time.Second)
		}
		<-Or(channels...)
	}
}
