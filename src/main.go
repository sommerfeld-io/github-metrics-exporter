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

func run(configPath string, reg prometheus.Registerer, listenAndServe func(string, http.Handler) error) error {
	if configPath == "" {
		return fmt.Errorf("--config flag is required")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := metrics.Register(reg); err != nil {
		return fmt.Errorf("failed to register metrics: %w", err)
	}

	srv := server.New(cfg.Port)
	log.Printf("Starting github-metrics-exporter on :%d", cfg.Port)
	return listenAndServe(fmt.Sprintf(":%d", cfg.Port), srv)
}

func main() {
	configPath := flag.String("config", "", "path to YAML configuration file (required)")
	flag.Parse()
	if err := run(*configPath, prometheus.DefaultRegisterer, http.ListenAndServe); err != nil {
		log.Fatal(err)
	}
}
