package main

import (
	"log"
	"net/http"

	"github.com/mirkobrombin/go-logger/pkg/logger"
)

func main() {
	lg := logger.New()
	sink := logger.NewPrometheusSink(logger.InfoLevel, "go_logger")
	lg.RegisterSink(sink)

	mux := http.NewServeMux()
	mux.Handle("/metrics", sink.Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		lg.Info("received request", logger.Field{Key: "path", Value: r.URL.Path})
		_, _ = w.Write([]byte("ok"))
	})

	log.Println("listening :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
