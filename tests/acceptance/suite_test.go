// Package acceptance_test contains GoDog acceptance tests for the GitHub Metrics Exporter.
// It starts the real HTTP server in-process via httptest and runs all Gherkin scenarios
// defined in the features/ directory.
package acceptance_test

import (
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/server"
)

var (
	testSrv *httptest.Server
	baseURL string
)

// TestMain registers metrics, starts a real HTTP test server, runs all GoDog scenarios,
// then tears down the server.
func TestMain(m *testing.M) {
	if err := metrics.Register(prometheus.DefaultRegisterer); err != nil {
		log.Fatalf("failed to register metrics: %v", err)
	}
	testSrv = httptest.NewServer(server.New())
	baseURL = testSrv.URL

	opts := godog.Options{
		Format: "pretty",
		Paths:  []string{"features"},
	}

	status := godog.TestSuite{
		Name:                "acceptance",
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}.Run()

	testSrv.Close()
	os.Exit(status)
}
