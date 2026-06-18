package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
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

func TestRunShouldReturnErrorWhenConfigPathIsEmpty(t *testing.T) {
	err := run("", prometheus.NewRegistry(), noopListen)
	if err == nil {
		t.Error("expected error when config path is empty, got nil")
	}
}

func TestRunShouldNotReturnNilWhenConfigPathIsEmpty(t *testing.T) {
	err := run("", prometheus.NewRegistry(), noopListen)
	if err == nil {
		t.Error("error must not be nil when config path is empty")
	}
}

func TestRunShouldReturnErrorMessageMentioningConfigFlag(t *testing.T) {
	err := run("", prometheus.NewRegistry(), noopListen)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "--config") {
		t.Errorf("expected error message to mention --config flag, got %q", err.Error())
	}
}

func TestRunShouldReturnErrorWhenConfigFileDoesNotExist(t *testing.T) {
	err := run("/nonexistent/path/config.yml", prometheus.NewRegistry(), noopListen)
	if err == nil {
		t.Error("expected error when config file does not exist, got nil")
	}
}

func TestRunShouldWrapConfigLoadError(t *testing.T) {
	err := run("/nonexistent/path/config.yml", prometheus.NewRegistry(), noopListen)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to load config") {
		t.Errorf("expected error to wrap config load context, got %q", err.Error())
	}
}

func TestRunShouldReturnErrorWhenMetricsRegistrationFails(t *testing.T) {
	original := metrics.MetricPrefix
	metrics.MetricPrefix = ""
	t.Cleanup(func() { metrics.MetricPrefix = original })

	path := writeConfig(t, "port: 9400\n")
	err := run(path, prometheus.NewRegistry(), noopListen)
	if err == nil {
		t.Error("expected error when metrics registration fails, got nil")
	}
}

func TestRunShouldWrapMetricsRegistrationError(t *testing.T) {
	original := metrics.MetricPrefix
	metrics.MetricPrefix = ""
	t.Cleanup(func() { metrics.MetricPrefix = original })

	path := writeConfig(t, "port: 9400\n")
	err := run(path, prometheus.NewRegistry(), noopListen)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to register metrics") {
		t.Errorf("expected error to wrap metrics registration context, got %q", err.Error())
	}
}

func TestRunShouldSucceedWithValidConfigAndNoopListen(t *testing.T) {
	path := writeConfig(t, "port: 9400\n")
	err := run(path, prometheus.NewRegistry(), noopListen)
	if err != nil {
		t.Errorf("expected no error with valid config and noop listen, got %v", err)
	}
}

func TestRunShouldNotReturnErrorOnSuccess(t *testing.T) {
	path := writeConfig(t, "port: 9400\n")
	err := run(path, prometheus.NewRegistry(), noopListen)
	if err != nil {
		t.Errorf("error must be nil on success, got %v", err)
	}
}

func TestRunShouldCallListenOnConfiguredPort(t *testing.T) {
	var capturedAddr string
	stub := func(addr string, _ http.Handler) error {
		capturedAddr = addr
		return nil
	}

	path := writeConfig(t, "port: 9400\n")
	if err := run(path, prometheus.NewRegistry(), stub); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAddr != ":9400" {
		t.Errorf("expected listen address :9400, got %q", capturedAddr)
	}
}

func TestRunShouldNotListenOnWrongPort(t *testing.T) {
	var capturedAddr string
	stub := func(addr string, _ http.Handler) error {
		capturedAddr = addr
		return nil
	}

	path := writeConfig(t, "port: 9400\n")
	if err := run(path, prometheus.NewRegistry(), stub); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAddr == ":9999" {
		t.Error("listen address must not be :9999 when port 9400 is configured")
	}
}

func TestRunShouldPropagateListenError(t *testing.T) {
	sentinel := errors.New("listen failed")
	stub := func(_ string, _ http.Handler) error {
		return sentinel
	}

	path := writeConfig(t, "port: 9400\n")
	err := run(path, prometheus.NewRegistry(), stub)
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel listen error to be propagated, got %v", err)
	}
}

func TestRunShouldNotSwallowListenError(t *testing.T) {
	stub := func(_ string, _ http.Handler) error {
		return fmt.Errorf("listen failed")
	}

	path := writeConfig(t, "port: 9400\n")
	err := run(path, prometheus.NewRegistry(), stub)
	if err == nil {
		t.Error("run must not return nil when the listen function errors")
	}
}
