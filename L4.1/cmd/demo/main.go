package main

import (
	"fmt"
	"time"

	"or-channel"
)

// sig creates a channel that closes after the specified duration
func sig(after time.Duration) <-chan any {
	c := make(chan any)
	go func() {
		defer close(c)
		time.Sleep(after)
	}()
	return c
}

func main() {
	start := time.Now()

	<-or.Or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)

	fmt.Printf("Завершено после %v\n", time.Since(start))
}
