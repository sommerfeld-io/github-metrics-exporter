package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/config"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/github"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics/repository"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/server"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/token"
)

// discoverFunc is the signature of the repository discovery function.
// It is injected into run so that tests can replace it with a no-op.
type discoverFunc func(ctx context.Context, orgs, users []string) ([]github.Repository, error)

func run(configPath string, reg prometheus.Registerer, discover discoverFunc, listenAndServe func(string, http.Handler) error) error {
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

	if len(cfg.GitHub.Organizations) == 0 && len(cfg.GitHub.Users) == 0 {
		slog.Warn("no GitHub targets configured; repository list will be empty")
	}

	ghRepos, err := discover(context.Background(), cfg.GitHub.Organizations, cfg.GitHub.Users)
	if err != nil {
		return fmt.Errorf("failed to discover repositories: %w", err)
	}

	srvRepos, err := applyRepos(ghRepos)
	if err != nil {
		return err
	}

	srv := server.New(cfg.Port, srvRepos)
	slog.Info("Starting github-metrics-exporter", "port", cfg.Port)
	return listenAndServe(fmt.Sprintf(":%d", cfg.Port), srv)
}

// applyRepos records each repository's accessibility metric and returns
// the equivalent slice of server.Repository for page rendering.
func applyRepos(ghRepos []github.Repository) ([]server.Repository, error) {
	srvRepos := make([]server.Repository, len(ghRepos))
	for i, r := range ghRepos {
		if err := repository.Set(r.Owner, r.Name, r.Accessible); err != nil {
			return nil, fmt.Errorf("failed to set repository metric: %w", err)
		}
		srvRepos[i] = server.Repository{Owner: r.Owner, Name: r.Name, Accessible: r.Accessible}
	}
	return srvRepos, nil
}

func main() {
	configPath := flag.String("config", "", "path to YAML configuration file (required)")
	flag.Parse()
	ghClient := github.NewWithToken(os.Getenv("GITHUB_TOKEN"))
	if err := run(*configPath, prometheus.DefaultRegisterer, ghClient.Discover, http.ListenAndServe); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
