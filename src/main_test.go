package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/github"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "ghme-config-*.yml")
	if err != nil {
		t.Fatalf("failed to create temp config: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	return f.Name()
}

func noopListen(addr string, handler http.Handler) error {
	return nil
}

func noopDiscover(_ context.Context, _, _ []string) ([]github.Repository, error) {
	return nil, nil
}

func noopFetchWorkflows(_ context.Context, _, _ string) ([]github.RunWithJobs, error) {
	return nil, nil
}

func setValidGitHubToken(t *testing.T) {
	t.Helper()
	t.Setenv("GITHUB_TOKEN", "test-token-for-unit-tests")
}

func TestRunShouldReturnErrorWhenGITHUBTOKENIsMissing(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	err := run("", prometheus.NewRegistry(), noopDiscover, noopFetchWorkflows, noopListen)
	if err == nil {
		t.Error("expected error when GITHUB_TOKEN is missing, got nil")
	}
}

func TestRunShouldNotReturnNilWhenGITHUBTOKENIsMissing(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	err := run("", prometheus.NewRegistry(), noopDiscover, noopFetchWorkflows, noopListen)
	if err == nil {
		t.Error("error must not be nil when GITHUB_TOKEN is missing")
	}
}

func TestRunShouldReturnErrorWhenGITHUBTOKENIsEmptyString(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	path := writeConfig(t, "port: 9400\n")
	err := run(path, prometheus.NewRegistry(), noopDiscover, noopFetchWorkflows, noopListen)
	if err == nil {
		t.Error("expected error when GITHUB_TOKEN is an empty string, got nil")
	}
}

func TestRunShouldReturnErrorMessageMentioningGITHUBTOKEN(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	err := run("", prometheus.NewRegistry(), noopDiscover, noopFetchWorkflows, noopListen)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "GITHUB_TOKEN") {
		t.Errorf("expected error message to mention GITHUB_TOKEN, got %q", err.Error())
	}
}

func TestRunShouldReturnErrorWhenConfigPathIsEmpty(t *testing.T) {
	setValidGitHubToken(t)
	err := run("", prometheus.NewRegistry(), noopDiscover, noopFetchWorkflows, noopListen)
	if err == nil {
		t.Error("expected error when config path is empty, got nil")
	}
}

func TestRunShouldNotReturnNilWhenConfigPathIsEmpty(t *testing.T) {
	setValidGitHubToken(t)
	err := run("", prometheus.NewRegistry(), noopDiscover, noopFetchWorkflows, noopListen)
	if err == nil {
		t.Error("error must not be nil when config path is empty")
	}
}

func TestRunShouldReturnErrorMessageMentioningConfigFlag(t *testing.T) {
	setValidGitHubToken(t)
	err := run("", prometheus.NewRegistry(), noopDiscover, noopFetchWorkflows, noopListen)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "--config") {
		t.Errorf("expected error message to mention --config flag, got %q", err.Error())
	}
}

func TestRunShouldReturnErrorWhenConfigFileDoesNotExist(t *testing.T) {
	setValidGitHubToken(t)
	err := run("/nonexistent/path/config.yml", prometheus.NewRegistry(), noopDiscover, noopFetchWorkflows, noopListen)
	if err == nil {
		t.Error("expected error when config file does not exist, got nil")
	}
}

func TestRunShouldWrapConfigLoadError(t *testing.T) {
	setValidGitHubToken(t)
	err := run("/nonexistent/path/config.yml", prometheus.NewRegistry(), noopDiscover, noopFetchWorkflows, noopListen)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to load config") {
		t.Errorf("expected error to wrap config load context, got %q", err.Error())
	}
}

func TestRunShouldReturnErrorWhenMetricsRegistrationFails(t *testing.T) {
	setValidGitHubToken(t)
	original := metrics.MetricPrefix
	metrics.MetricPrefix = ""
	t.Cleanup(func() { metrics.MetricPrefix = original })

	path := writeConfig(t, "port: 9400\n")
	err := run(path, prometheus.NewRegistry(), noopDiscover, noopFetchWorkflows, noopListen)
	if err == nil {
		t.Error("expected error when metrics registration fails, got nil")
	}
}

func TestRunShouldWrapMetricsRegistrationError(t *testing.T) {
	setValidGitHubToken(t)
	original := metrics.MetricPrefix
	metrics.MetricPrefix = ""
	t.Cleanup(func() { metrics.MetricPrefix = original })

	path := writeConfig(t, "port: 9400\n")
	err := run(path, prometheus.NewRegistry(), noopDiscover, noopFetchWorkflows, noopListen)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to register metrics") {
		t.Errorf("expected error to wrap metrics registration context, got %q", err.Error())
	}
}

func TestRunShouldSucceedWithValidConfigAndNoopListen(t *testing.T) {
	setValidGitHubToken(t)
	path := writeConfig(t, "port: 9400\n")
	err := run(path, prometheus.NewRegistry(), noopDiscover, noopFetchWorkflows, noopListen)
	if err != nil {
		t.Errorf("expected no error with valid config and noop listen, got %v", err)
	}
}

func TestRunShouldNotReturnErrorOnSuccess(t *testing.T) {
	setValidGitHubToken(t)
	path := writeConfig(t, "port: 9400\n")
	err := run(path, prometheus.NewRegistry(), noopDiscover, noopFetchWorkflows, noopListen)
	if err != nil {
		t.Errorf("error must be nil on success, got %v", err)
	}
}

func TestRunShouldCallListenOnConfiguredPort(t *testing.T) {
	setValidGitHubToken(t)
	var capturedAddr string
	stub := func(addr string, _ http.Handler) error {
		capturedAddr = addr
		return nil
	}

	path := writeConfig(t, "port: 9400\n")
	if err := run(path, prometheus.NewRegistry(), noopDiscover, noopFetchWorkflows, stub); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAddr != ":9400" {
		t.Errorf("expected listen address :9400, got %q", capturedAddr)
	}
}

func TestRunShouldNotListenOnWrongPort(t *testing.T) {
	setValidGitHubToken(t)
	var capturedAddr string
	stub := func(addr string, _ http.Handler) error {
		capturedAddr = addr
		return nil
	}

	path := writeConfig(t, "port: 9400\n")
	if err := run(path, prometheus.NewRegistry(), noopDiscover, noopFetchWorkflows, stub); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAddr == ":9999" {
		t.Error("listen address must not be :9999 when port 9400 is configured")
	}
}

func TestRunShouldPropagateListenError(t *testing.T) {
	setValidGitHubToken(t)
	sentinel := errors.New("listen failed")
	stub := func(_ string, _ http.Handler) error {
		return sentinel
	}

	path := writeConfig(t, "port: 9400\n")
	err := run(path, prometheus.NewRegistry(), noopDiscover, noopFetchWorkflows, stub)
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel listen error to be propagated, got %v", err)
	}
}

func TestRunShouldNotSwallowListenError(t *testing.T) {
	setValidGitHubToken(t)
	stub := func(_ string, _ http.Handler) error {
		return fmt.Errorf("listen failed")
	}

	path := writeConfig(t, "port: 9400\n")
	err := run(path, prometheus.NewRegistry(), noopDiscover, noopFetchWorkflows, stub)
	if err == nil {
		t.Error("run must not return nil when the listen function errors")
	}
}

func TestRunShouldNotReturnErrorWhenDiscoverReturnsEmptyList(t *testing.T) {
	setValidGitHubToken(t)
	emptyDiscover := func(_ context.Context, _, _ []string) ([]github.Repository, error) {
		return []github.Repository{}, nil
	}
	path := writeConfig(t, "port: 9400\n")
	err := run(path, prometheus.NewRegistry(), emptyDiscover, noopFetchWorkflows, noopListen)
	if err != nil {
		t.Errorf("expected no error when discover returns empty list, got %v", err)
	}
}

func TestDoCollectShouldReturnErrorWhenDiscoverFails(t *testing.T) {
	reg := prometheus.NewRegistry()
	if err := metrics.Register(reg); err != nil {
		t.Fatalf("failed to register metrics: %v", err)
	}
	failDiscover := func(_ context.Context, _, _ []string) ([]github.Repository, error) {
		return nil, fmt.Errorf("discovery failed")
	}
	_, err := doCollect(context.Background(), nil, nil, failDiscover, noopFetchWorkflows)
	if err == nil {
		t.Error("expected error when discover fails, got nil")
	}
}

func TestDoCollectShouldNotReturnErrorForEmptyDiscovery(t *testing.T) {
	reg := prometheus.NewRegistry()
	if err := metrics.Register(reg); err != nil {
		t.Fatalf("failed to register metrics: %v", err)
	}
	repos, err := doCollect(context.Background(), nil, nil, noopDiscover, noopFetchWorkflows)
	if err != nil {
		t.Errorf("expected no error for empty discovery, got %v", err)
	}
	if repos == nil {
		t.Error("expected non-nil repo slice for empty discovery")
	}
}

func TestDoCollectShouldSucceedWhenWorkflowFetchFails(t *testing.T) {
	reg := prometheus.NewRegistry()
	if err := metrics.Register(reg); err != nil {
		t.Fatalf("failed to register metrics: %v", err)
	}
	accessibleDiscover := func(_ context.Context, _, _ []string) ([]github.Repository, error) {
		return []github.Repository{{Owner: "org", Name: "repo", Accessible: true}}, nil
	}
	failFetch := func(_ context.Context, _, _ string) ([]github.RunWithJobs, error) {
		return nil, fmt.Errorf("github API error")
	}
	repos, err := doCollect(context.Background(), nil, nil, accessibleDiscover, failFetch)
	if err != nil {
		t.Errorf("expected no error when workflow fetch fails (warn-and-continue), got %v", err)
	}
	if len(repos) == 0 {
		t.Error("expected repos to be returned even when workflow fetch fails")
	}
}

func TestDoCollectShouldNotReturnErrorWhenWorkflowFetchReturnsEmpty(t *testing.T) {
	reg := prometheus.NewRegistry()
	if err := metrics.Register(reg); err != nil {
		t.Fatalf("failed to register metrics: %v", err)
	}
	accessibleDiscover := func(_ context.Context, _, _ []string) ([]github.Repository, error) {
		return []github.Repository{{Owner: "org", Name: "repo", Accessible: true}}, nil
	}
	emptyFetch := func(_ context.Context, _, _ string) ([]github.RunWithJobs, error) {
		return []github.RunWithJobs{}, nil
	}
	_, err := doCollect(context.Background(), nil, nil, accessibleDiscover, emptyFetch)
	if err != nil {
		t.Errorf("expected no error when workflow fetch returns empty list, got %v", err)
	}
}

func TestDoCollectShouldReturnAllReposWhenSomeFetchesFail(t *testing.T) {
	reg := prometheus.NewRegistry()
	if err := metrics.Register(reg); err != nil {
		t.Fatalf("failed to register metrics: %v", err)
	}
	multiDiscover := func(_ context.Context, _, _ []string) ([]github.Repository, error) {
		return []github.Repository{
			{Owner: "org", Name: "repo-ok", Accessible: true},
			{Owner: "org", Name: "repo-fail", Accessible: true},
			{Owner: "org", Name: "repo-ok2", Accessible: true},
		}, nil
	}
	partialFetch := func(_ context.Context, _, repo string) ([]github.RunWithJobs, error) {
		if repo == "repo-fail" {
			return nil, fmt.Errorf("github API error")
		}
		return []github.RunWithJobs{}, nil
	}
	repos, err := doCollect(context.Background(), nil, nil, multiDiscover, partialFetch)
	if err != nil {
		t.Errorf("expected no error when some workflow fetches fail, got %v", err)
	}
	if len(repos) != 3 {
		t.Errorf("expected 3 repos regardless of fetch failures, got %d", len(repos))
	}
}

func TestDoCollectShouldNotReturnFewerReposWhenSomeFetchesFail(t *testing.T) {
	reg := prometheus.NewRegistry()
	if err := metrics.Register(reg); err != nil {
		t.Fatalf("failed to register metrics: %v", err)
	}
	multiDiscover := func(_ context.Context, _, _ []string) ([]github.Repository, error) {
		return []github.Repository{
			{Owner: "org", Name: "repo-ok", Accessible: true},
			{Owner: "org", Name: "repo-fail", Accessible: true},
		}, nil
	}
	partialFetch := func(_ context.Context, _, repo string) ([]github.RunWithJobs, error) {
		if repo == "repo-fail" {
			return nil, fmt.Errorf("github API error")
		}
		return []github.RunWithJobs{}, nil
	}
	repos, err := doCollect(context.Background(), nil, nil, multiDiscover, partialFetch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) < 2 {
		t.Errorf("fetch failure must not drop repos from the returned list; got %d, want 2", len(repos))
	}
}
