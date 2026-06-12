package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/server"
)

func main() {
	if err := metrics.Register(prometheus.DefaultRegisterer); err != nil {
		log.Fatalf("failed to register metrics: %v", err)
	}
	port := "9400"

	server := server.New()
	log.Printf("Starting github-metrics-exporter on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, server))
}
