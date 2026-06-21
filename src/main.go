package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/config"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/server"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/token"
)

func run(configPath string, reg prometheus.Registerer, listenAndServe func(string, http.Handler) error) error {
	if err := token.Validate(os.Getenv("GITHUB_TOKEN")); err != nil {
		return err
	}

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
	slog.Info("Starting github-metrics-exporter", "port", cfg.Port)
	return listenAndServe(fmt.Sprintf(":%d", cfg.Port), srv)
}

func main() {
	configPath := flag.String("config", "", "path to YAML configuration file (required)")
	flag.Parse()
	if err := run(*configPath, prometheus.DefaultRegisterer, http.ListenAndServe); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
