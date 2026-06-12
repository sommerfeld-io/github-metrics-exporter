package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/server"
)

func main() {
	metrics.Register(prometheus.DefaultRegisterer)

	mux := server.New()
	log.Println("Starting github-metrics-exporter on :9400")
	log.Fatal(http.ListenAndServe(":9400", mux))
}
