package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"

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

	collect := func(ctx context.Context) ([]server.Repository, error) {
		return doCollect(ctx, cfg.GitHub.Organizations, cfg.GitHub.Users, discover, fetchWorkflows)
	}

	srv := server.New(cfg.Port, collect)
	slog.Info("Starting github-metrics-exporter", "port", cfg.Port)
	return listenAndServe(fmt.Sprintf(":%d", cfg.Port), srv)
}

// doCollect discovers repositories, resets and repopulates all Prometheus gauges,
// and returns the current repository list. It is called on every /metrics scrape.
// Per-repo workflow fetch failures are logged and skipped (warn-and-continue).
func doCollect(ctx context.Context, orgs, users []string, discover discoverFunc, fetchWorkflows fetchWorkflowsFunc) ([]server.Repository, error) {
	ghRepos, err := discover(ctx, orgs, users)
	if err != nil {
		return nil, fmt.Errorf("failed to discover repositories: %w", err)
	}

	repository.Accessible.Reset()
	srvRepos := make([]server.Repository, 0, len(ghRepos))
	for _, r := range ghRepos {
		if err := repository.Set(r.Owner, r.Name, r.Accessible); err != nil {
			return nil, fmt.Errorf("failed to set repository metric: %w", err)
		}
		srvRepos = append(srvRepos, server.Repository{Owner: r.Owner, Name: r.Name, Accessible: r.Accessible})
	}

	workflow.RunConclusion.Reset()
	workflow.JobConclusion.Reset()
	var wg sync.WaitGroup
	for _, r := range srvRepos {
		if !r.Accessible {
			continue
		}
		wg.Add(1)
		go func(r server.Repository) {
			defer wg.Done()
			runsWithJobs, err := fetchWorkflows(ctx, r.Owner, r.Name)
			if err != nil {
				slog.Warn("collect: failed to fetch workflow data; skipping repo", "owner", r.Owner, "repo", r.Name, "error", err)
				return
			}
			if err := workflow.Record(r.Owner, r.Name, runsWithJobs); err != nil {
				slog.Warn("collect: failed to record workflow metrics; skipping repo", "owner", r.Owner, "repo", r.Name, "error", err)
			}
		}(r)
	}
	wg.Wait()

	return srvRepos, nil
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
