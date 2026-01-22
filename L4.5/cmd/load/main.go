package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

var (
	targetURL   = flag.String("url", "http://localhost:8080/sum", "Target URL")
	concurrency = flag.Int("c", 10, "Concurrency level")
	duration    = flag.Duration("d", 10*time.Second, "Test duration")
)

func main() {
	flag.Parse()

	log.Printf("Starting load test on %s with %d workers for %v", *targetURL, *concurrency, *duration)

	var ops int64
	var errors int64

	start := time.Now()
	timeout := time.After(*duration)
	
	var wg sync.WaitGroup
	done := make(chan struct{})

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{
				Transport: &http.Transport{
					MaxIdleConns:        100,
					MaxIdleConnsPerHost: 100,
				},
				Timeout: 2 * time.Second,
			}

			for {
				select {
				case <-done:
					return
				default:
					a := rand.Intn(1000)
					b := rand.Intn(1000)
					u := fmt.Sprintf("%s?a=%d&b=%d", *targetURL, a, b)

					resp, err := client.Get(u)
					if err != nil {
						atomic.AddInt64(&errors, 1)
						continue
					}
					io.Copy(ioutil.Discard, resp.Body)
					resp.Body.Close()
					
					if resp.StatusCode != http.StatusOK {
						atomic.AddInt64(&errors, 1)
					} else {
						atomic.AddInt64(&ops, 1)
					}
				}
			}
		}()
	}

	<-timeout
	close(done)
	wg.Wait()

	totalTime := time.Since(start)
	rps := float64(ops) / totalTime.Seconds()

	log.Printf("Done!")
	log.Printf("Total Requests: %d", ops)
	log.Printf("Total Errors: %d", errors)
	log.Printf("RPS: %.2f", rps)
}
