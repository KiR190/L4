package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/trace"

	"sum-service/internal/handler"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if os.Getenv("TRACE") == "1" {
		f, err := os.Create("trace.out")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		if err := trace.Start(f); err != nil {
			log.Fatal(err)
		}
		defer trace.Stop()
	}

	// Основной API
	mux := http.NewServeMux()
	mux.HandleFunc("/sum", handler.SumHandler)
	mux.HandleFunc("/sum-bad", handler.SumHandlerBad)

	// pprof на отдельном порту
	go func() {
		log.Println("pprof on :6060")
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	log.Println("server on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
