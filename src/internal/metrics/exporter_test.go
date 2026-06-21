package metrics_test

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics/repository"
)

// TestRegisterShouldSucceedWithValidRegistry verifies that Register returns nil
// when given a valid registry and the default MetricPrefix.
func TestRegisterShouldSucceedWithValidRegistry(t *testing.T) {
	reg := prometheus.NewRegistry()
	if err := metrics.Register(reg); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

// TestRegisterShouldExposeMetricOnRegistry verifies that after Register the
// metric is gatherable from the registry that was passed in.
func TestRegisterShouldExposeMetricOnRegistry(t *testing.T) {
	reg := prometheus.NewRegistry()
	if err := metrics.Register(reg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}
	if len(mfs) == 0 {
		t.Error("expected at least one metric family after Register")
	}
}

// TestRegisterShouldNotExposeMetricOnOtherRegistry verifies that a registry
// that was not passed to Register does not contain the exported metric.
func TestRegisterShouldNotExposeMetricOnOtherRegistry(t *testing.T) {
	if err := metrics.Register(prometheus.NewRegistry()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	other := prometheus.NewRegistry()
	mfs, err := other.Gather()
	if err != nil {
		t.Fatalf("failed to gather from other registry: %v", err)
	}
	if len(mfs) != 0 {
		t.Error("metric must not appear on a registry that was not passed to Register")
	}
}

// TestRegisterShouldUseMetricPrefix verifies that the registered metric name
// starts with the configured MetricPrefix.
func TestRegisterShouldUseMetricPrefix(t *testing.T) {
	reg := prometheus.NewRegistry()
	if err := metrics.Register(reg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}
	if len(mfs) == 0 {
		t.Fatal("expected at least one metric family")
	}
	if !strings.HasPrefix(mfs[0].GetName(), metrics.MetricPrefix) {
		t.Errorf("expected metric name to start with %q, got %q", metrics.MetricPrefix, mfs[0].GetName())
	}
}

// TestRegisterShouldNotUseEmptyPrefix verifies the metric name is not the bare
// "exporter_info" - the prefix must always be present.
func TestRegisterShouldNotUseEmptyPrefix(t *testing.T) {
	reg := prometheus.NewRegistry()
	if err := metrics.Register(reg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}
	if len(mfs) == 0 {
		t.Fatal("expected at least one metric family")
	}
	if mfs[0].GetName() == "exporter_info" {
		t.Error("metric name must include the prefix, not just \"exporter_info\"")
	}
}

// TestRegisterShouldReturnErrorWhenPrefixIsEmpty verifies that Register
// propagates the error from meta.Init when MetricPrefix is empty.
func TestRegisterShouldReturnErrorWhenPrefixIsEmpty(t *testing.T) {
	original := metrics.MetricPrefix
	metrics.MetricPrefix = ""
	t.Cleanup(func() { metrics.MetricPrefix = original })

	err := metrics.Register(prometheus.NewRegistry())
	if err == nil {
		t.Error("Register must return an error when MetricPrefix is empty")
	}
}

// TestRegisterShouldExposeRepositoryAccessibleMetric verifies that after Register
// the ghme_repository_accessible metric family is gatherable from the registry.
func TestRegisterShouldExposeRepositoryAccessibleMetric(t *testing.T) {
	reg := prometheus.NewRegistry()
	if err := metrics.Register(reg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// A GaugeVec only appears in Gather output once it has at least one time series.
	if err := repository.Set("test-owner", "test-repo", true); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}
	found := false
	for _, mf := range mfs {
		if strings.Contains(mf.GetName(), "repository_accessible") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ghme_repository_accessible metric family to be registered")
	}
}

// TestRegisterShouldNotExposeRepositoryMetricOnOtherRegistry verifies that a registry
// that was not passed to Register does not contain the repository metric.
func TestRegisterShouldNotExposeRepositoryMetricOnOtherRegistry(t *testing.T) {
	if err := metrics.Register(prometheus.NewRegistry()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	other := prometheus.NewRegistry()
	mfs, err := other.Gather()
	if err != nil {
		t.Fatalf("failed to gather from other registry: %v", err)
	}
	for _, mf := range mfs {
		if strings.Contains(mf.GetName(), "repository_accessible") {
			t.Error("repository_accessible metric must not appear on a registry that was not passed to Register")
		}
	}
}

// TestRegisterShouldWrapInitError verifies that the error returned when
// MetricPrefix is empty is wrapped with the "metrics.Register:" context.
func TestRegisterShouldWrapInitErrorContext(t *testing.T) {
	original := metrics.MetricPrefix
	metrics.MetricPrefix = ""
	t.Cleanup(func() { metrics.MetricPrefix = original })

	err := metrics.Register(prometheus.NewRegistry())
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.HasPrefix(err.Error(), "metrics.Register:") {
		t.Errorf("expected error to start with \"metrics.Register:\", got: %q", err.Error())
	}
}
