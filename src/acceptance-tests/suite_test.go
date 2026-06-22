// Package acceptance_test contains GoDog acceptance tests for the GitHub Metrics Exporter.
// It starts the real HTTP server in-process via httptest and runs all Gherkin scenarios
// defined in the features/ directory.
//
// These tests are part of the main module and share its go.mod. They are intentionally
// excluded from the unit-test coverage run (go test ./internal/...) and are invoked
// explicitly via "task go:test:acceptance" or as a gate inside "task go:build".
package acceptance_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/github"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics/repository"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics/workflow"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/config"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/server"
)

var (
	testSrv      *httptest.Server
	noTargetsSrv *httptest.Server
	baseURL      string
	noTargetsURL string
)

// testRepos is the fixed set of repositories used in the acceptance test server.
// It includes one accessible and one inaccessible entry so that both states can be tested.
var testRepos = []server.Repository{
	{Owner: "test-org", Name: "repo-accessible", Accessible: true},
	{Owner: "test-org", Name: "repo-locked", Accessible: false},
}

func writeTempConfig(port int) string {
	f, err := os.CreateTemp("", "ghme-acceptance-config-*.yml")
	if err != nil {
		slog.Error("setup: create temp config", "error", err)
		os.Exit(1)
	}
	if _, err := fmt.Fprintf(f, "port: %d\n", port); err != nil {
		slog.Error("setup: write temp config", "error", err)
		os.Exit(1)
	}
	f.Close()
	return f.Name()
}

// testCollect reseeds repository and workflow metrics from the fixed test data on every call.
// It is used as the CollectFunc for the acceptance test server so that /metrics always
// reflects the expected test state regardless of how many times it is scraped.
func testCollect(_ context.Context) ([]server.Repository, error) {
	repository.Accessible.Reset()
	for _, r := range testRepos {
		if err := repository.Set(r.Owner, r.Name, r.Accessible); err != nil {
			return nil, err
		}
	}
	workflow.RunConclusion.Reset()
	workflow.JobConclusion.Reset()
	if err := workflow.Record("test-org", "repo-accessible", []github.RunWithJobs{
		{
			Run:  github.WorkflowRun{Name: "CI", HeadBranch: "main", Conclusion: "success"},
			Jobs: []github.Job{{Name: "build", Conclusion: "success"}},
		},
		{
			Run:  github.WorkflowRun{Name: "CI", HeadBranch: "main", Conclusion: "failure"},
			Jobs: []github.Job{{Name: "build", Conclusion: "failure"}},
		},
	}); err != nil {
		return nil, err
	}
	return testRepos, nil
}

// TestMain registers metrics, starts two real HTTP test servers (one with repos, one without),
// triggers an initial /metrics scrape on the main server so the index page is populated,
// delegates to m.Run so that Go's coverage instrumentation is flushed before os.Exit is called,
// then tears down both servers.
func TestMain(m *testing.M) {
	if err := metrics.Register(prometheus.DefaultRegisterer); err != nil {
		slog.Error("failed to register metrics", "error", err)
		os.Exit(1)
	}

	cfgPath := writeTempConfig(9400)
	defer os.Remove(cfgPath)

	cfg, err := config.Load(cfgPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	testSrv = httptest.NewServer(server.New(cfg.Port, testCollect))
	baseURL = testSrv.URL

	// Trigger an initial scrape so that lastRepos is populated and the index page
	// shows repository data for scenarios that navigate to "/".
	if _, err := http.Get(baseURL + "/metrics"); err != nil {
		slog.Error("setup: initial metrics scrape failed", "error", err)
		os.Exit(1)
	}

	noTargetsSrv = httptest.NewServer(server.New(cfg.Port, func(_ context.Context) ([]server.Repository, error) {
		return nil, nil
	}))
	noTargetsURL = noTargetsSrv.URL

	exitCode := m.Run()

	testSrv.Close()
	noTargetsSrv.Close()
	os.Exit(exitCode)
}

// TestAcceptanceSuite runs all GoDog Gherkin scenarios as a regular Go test so that coverage
// data collected by -coverpkg is flushed properly by m.Run before os.Exit is called.
func TestAcceptanceSuite(t *testing.T) {
	opts := godog.Options{
		Format: "pretty",
		Paths:  []string{"features"},
	}

	suite := godog.TestSuite{
		Name: "acceptance",
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			InitializeScenario(ctx)
			InitializeConfigScenario(ctx)
			InitializeTokenScenario(ctx)
			InitializeRepositoryScenario(ctx)
			InitializeWorkflowScenario(ctx)
		},
		Options: &opts,
	}

	if suite.Run() != 0 {
		t.Fatal("acceptance test suite returned non-zero exit code")
	}
}
