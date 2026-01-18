package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"gc-metrics/internal/metrics"
)

var ballast []byte

func main() {
	// GC percent (optional)
	if v := os.Getenv("GC_PERCENT"); v != "" {
		if percent, err := strconv.Atoi(v); err == nil {
			metrics.SetGCPercent(percent)
			log.Printf("GC percent set to %d\n", percent)
		}
	}

	// Memory ballast (optional)
	if v := os.Getenv("BALLAST_MB"); v != "" {
		if mb, err := strconv.Atoi(v); err == nil && mb > 0 {
			ballast = make([]byte, mb<<20)
			log.Printf("Memory ballast allocated: %d MB\n", mb)
		}
	}

	// Start runtime metrics collection
	metrics.StartRuntimeMetrics(5 * time.Second)

	// HTTP endpoints
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	log.Println("Server started on :8080")
	log.Println("Metrics: http://localhost:8080/metrics")
	log.Println("pprof:   http://localhost:8080/debug/pprof/")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
