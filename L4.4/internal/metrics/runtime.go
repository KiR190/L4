package metrics

import (
	"runtime"
	"runtime/debug"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	memAlloc = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "go",
		Subsystem: "mem",
		Name:      "alloc_bytes",
		Help:      "Currently allocated memory",
	})

	heapAlloc = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "go",
		Subsystem: "mem",
		Name:      "heap_alloc_bytes",
		Help:      "Heap allocated memory",
	})

	totalAlloc = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "go",
		Subsystem: "mem",
		Name:      "total_alloc_bytes",
		Help:      "Total allocated memory",
	})

	mallocs = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "go",
		Subsystem: "mem",
		Name:      "mallocs_total",
		Help:      "Number of mallocs",
	})

	frees = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "go",
		Subsystem: "mem",
		Name:      "frees_total",
		Help:      "Number of frees",
	})

	gcCycles = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "go",
		Subsystem: "gc",
		Name:      "cycles_total",
		Help:      "Total number of GC cycles",
	})

	lastGCTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "go",
		Subsystem: "gc",
		Name:      "last_time_seconds",
		Help:      "Last GC time (unix timestamp)",
	})
)

func init() {
	prometheus.MustRegister(
		memAlloc,
		heapAlloc,
		totalAlloc,
		mallocs,
		frees,
		gcCycles,
		lastGCTime,
	)
}

func SetGCPercent(percent int) {
	debug.SetGCPercent(percent)
}

func StartRuntimeMetrics(interval time.Duration) {
	go func() {
		var (
			lastTotalAlloc uint64
			lastMallocs    uint64
			lastFrees      uint64
			lastNumGC      uint32
		)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			var ms runtime.MemStats
			runtime.ReadMemStats(&ms)

			memAlloc.Set(float64(ms.Alloc))
			heapAlloc.Set(float64(ms.HeapAlloc))

			// Counters should be incremented by the delta
			if ms.TotalAlloc > lastTotalAlloc {
				totalAlloc.Add(float64(ms.TotalAlloc - lastTotalAlloc))
				lastTotalAlloc = ms.TotalAlloc
			}

			if ms.Mallocs > lastMallocs {
				mallocs.Add(float64(ms.Mallocs - lastMallocs))
				lastMallocs = ms.Mallocs
			}

			if ms.Frees > lastFrees {
				frees.Add(float64(ms.Frees - lastFrees))
				lastFrees = ms.Frees
			}

			if ms.NumGC > lastNumGC {
				gcCycles.Add(float64(ms.NumGC - lastNumGC))
				lastNumGC = ms.NumGC
			}

			if ms.LastGC != 0 {
				lastGCTime.Set(float64(ms.LastGC) / 1e9)
			}
		}
	}()
}
