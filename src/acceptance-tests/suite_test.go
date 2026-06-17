// Package acceptance_test contains GoDog acceptance tests for the GitHub Metrics Exporter.
// It starts the real HTTP server in-process via httptest and runs all Gherkin scenarios
// defined in the features/ directory.
//
// These tests are part of the main module and share its go.mod. They are intentionally
// excluded from the unit-test coverage run (go test ./internal/...) and are invoked
// explicitly via "task go:test:acceptance" or as a gate inside "task go:build".
package acceptance_test

import (
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/config"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/server"
)

var (
	testSrv *httptest.Server
	baseURL string
)

func writeTempConfig(port int) string {
	f, err := os.CreateTemp("", "ghme-acceptance-config-*.yml")
	if err != nil {
		log.Fatalf("setup: create temp config: %v", err)
	}
	if _, err := fmt.Fprintf(f, "port: %d\n", port); err != nil {
		log.Fatalf("setup: write temp config: %v", err)
	}
	f.Close()
	return f.Name()
}

// TestMain registers metrics, starts a real HTTP test server, delegates to m.Run so that
// Go's coverage instrumentation is flushed before os.Exit is called, then tears down the server.
func TestMain(m *testing.M) {
	if err := metrics.Register(prometheus.DefaultRegisterer); err != nil {
		log.Fatalf("failed to register metrics: %v", err)
	}

	cfgPath := writeTempConfig(9400)
	defer os.Remove(cfgPath)

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	testSrv = httptest.NewServer(server.New(cfg.Port))
	baseURL = testSrv.URL

	exitCode := m.Run()

	testSrv.Close()
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
		},
		Options: &opts,
	}

	if suite.Run() != 0 {
		t.Fatal("acceptance test suite returned non-zero exit code")
	}
}
