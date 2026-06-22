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
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics/workflow"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/server"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/token"
)

// discoverFunc is the signature of the repository discovery function.
// It is injected into run so that tests can replace it with a no-op.
type discoverFunc func(ctx context.Context, orgs, users []string) ([]github.Repository, error)

// fetchWorkflowsFunc is the signature of the workflow fetch function.
// It is injected into run so that tests can replace it with a no-op.
type fetchWorkflowsFunc func(ctx context.Context, owner, repo string) ([]github.RunWithJobs, error)

func run(configPath string, reg prometheus.Registerer, discover discoverFunc, fetchWorkflows fetchWorkflowsFunc, listenAndServe func(string, http.Handler) error) error {
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

	if err := applyWorkflows(context.Background(), srvRepos, fetchWorkflows); err != nil {
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

// applyWorkflows fetches workflow data for each accessible repository and records it as metrics.
// Per-repo fetch failures are logged and skipped rather than aborting startup.
func applyWorkflows(ctx context.Context, repos []server.Repository, fetch fetchWorkflowsFunc) error {
	for _, r := range repos {
		if !r.Accessible {
			continue
		}
		runsWithJobs, err := fetch(ctx, r.Owner, r.Name)
		if err != nil {
			slog.Warn("failed to fetch workflow data; skipping repo", "owner", r.Owner, "repo", r.Name, "error", err)
			continue
		}
		if err := workflow.Record(r.Owner, r.Name, runsWithJobs); err != nil {
			return fmt.Errorf("failed to record workflow metrics for %s/%s: %w", r.Owner, r.Name, err)
		}
	}
	return nil
}

func main() {
	configPath := flag.String("config", "", "path to YAML configuration file (required)")
	flag.Parse()
	ghClient := github.NewWithToken(os.Getenv("GITHUB_TOKEN"))
	if err := run(*configPath, prometheus.DefaultRegisterer, ghClient.Discover, ghClient.FetchWorkflowsWithJobs, http.ListenAndServe); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
