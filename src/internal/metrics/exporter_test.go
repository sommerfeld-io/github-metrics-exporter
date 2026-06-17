package metrics_test

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics"
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
