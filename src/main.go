package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/config"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/server"
)

func main() {
	configPath := flag.String("config", "", "path to YAML configuration file (required)")
	flag.Parse()

	if *configPath == "" {
		log.Fatal("--config flag is required")
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := metrics.Register(prometheus.DefaultRegisterer); err != nil {
		log.Fatalf("failed to register metrics: %v", err)
	}

	srv := server.New(cfg.Port)
	log.Printf("Starting github-metrics-exporter on :%d", cfg.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), srv))
}
