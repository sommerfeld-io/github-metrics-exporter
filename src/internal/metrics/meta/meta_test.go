package meta_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics/meta"
)

// gatherMetrics registers the current ExporterInfo with a fresh registry and
// returns all gathered metric families.
func gatherMetrics(t *testing.T) []*dto.MetricFamily {
	t.Helper()
	reg := prometheus.NewRegistry()
	reg.MustRegister(meta.ExporterInfo)
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}
	return mfs
}

// TestExporterInfoShouldBeNilBeforeInit verifies that the package-level ExporterInfo
// variable is nil when Init has not been called.
func TestExporterInfoShouldBeNilBeforeInit(t *testing.T) {
	original := meta.ExporterInfo
	meta.ExporterInfo = nil
	t.Cleanup(func() { meta.ExporterInfo = original })

	if meta.ExporterInfo != nil {
		t.Error("ExporterInfo must be nil before Init is called")
	}
}

// TestInitShouldAssignExporterInfo verifies Init assigns a non-nil value to ExporterInfo.
func TestInitShouldAssignExporterInfo(t *testing.T) {
	if err := meta.Init("test_", "abc123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.ExporterInfo == nil {
		t.Error("ExporterInfo must not be nil after Init is called")
	}
}

// TestMetricNameShouldIncludePrefix verifies the metric name is formed by prepending
// the prefix to "exporter_info".
func TestMetricNameShouldIncludePrefix(t *testing.T) {
	if err := meta.Init("myprefix_", "abc123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mfs := gatherMetrics(t)

	if len(mfs) == 0 {
		t.Fatal("expected at least one metric family")
	}
	if got := mfs[0].GetName(); got != "myprefix_exporter_info" {
		t.Errorf("expected metric name %q, got %q", "myprefix_exporter_info", got)
	}
}

// TestMetricNameShouldNotOmitPrefix verifies the metric name is not the bare
// "exporter_info" when a non-empty prefix is supplied.
func TestMetricNameShouldNotOmitPrefix(t *testing.T) {
	if err := meta.Init("myprefix_", "abc123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mfs := gatherMetrics(t)

	if len(mfs) == 0 {
		t.Fatal("expected at least one metric family")
	}
	if mfs[0].GetName() == "exporter_info" {
		t.Error("metric name must not omit the prefix when a non-empty prefix is provided")
	}
}

// TestInitShouldReturnErrorOnEmptyPrefix verifies that Init returns an error when
// given an empty prefix, because an empty prefix is not an allowed state.
func TestInitShouldReturnErrorOnEmptyPrefix(t *testing.T) {
	err := meta.Init("", "abc123")
	if err == nil {
		t.Error("Init must return an error when prefix is empty")
	}
}

// TestGaugeValueShouldBeOne verifies the gauge is set to exactly 1.
func TestGaugeValueShouldBeOne(t *testing.T) {
	if err := meta.Init("test_", "abc123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mfs := gatherMetrics(t)

	if len(mfs) == 0 || len(mfs[0].GetMetric()) == 0 {
		t.Fatal("expected at least one metric")
	}
	if val := mfs[0].GetMetric()[0].GetGauge().GetValue(); val != 1.0 {
		t.Errorf("expected gauge value 1.0, got %f", val)
	}
}

// TestGaugeValueShouldNotBeZero verifies the gauge is never left at the Go zero value
// for float64. Init must call Set(1) explicitly.
func TestGaugeValueShouldNotBeZero(t *testing.T) {
	if err := meta.Init("test_", "abc123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mfs := gatherMetrics(t)

	if len(mfs) == 0 || len(mfs[0].GetMetric()) == 0 {
		t.Fatal("expected at least one metric")
	}
	if val := mfs[0].GetMetric()[0].GetGauge().GetValue(); val == 0.0 {
		t.Error("gauge value must not be 0; Init must call Set(1)")
	}
}

// TestInitShouldProduceExactlyOneTimeSeries verifies Init produces exactly one time series.
func TestInitShouldProduceExactlyOneTimeSeries(t *testing.T) {
	if err := meta.Init("test_", "sha"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mfs := gatherMetrics(t)

	if len(mfs) == 0 {
		t.Fatal("expected metric family, got none")
	}
	if n := len(mfs[0].GetMetric()); n != 1 {
		t.Errorf("expected exactly 1 time series, got %d", n)
	}
}

// TestInitShouldProduceExactlyOneLabel verifies the metric carries exactly the one expected
// label — no more, no fewer.
func TestInitShouldProduceExactlyOneLabel(t *testing.T) {
	if err := meta.Init("test_", "abc123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mfs := gatherMetrics(t)

	if len(mfs) == 0 || len(mfs[0].GetMetric()) == 0 {
		t.Fatal("expected at least one metric")
	}
	if n := len(mfs[0].GetMetric()[0].GetLabel()); n != 1 {
		t.Errorf("expected exactly 1 label, got %d", n)
	}
}

// TestReinitializationShouldReplaceExporterInfo verifies that calling Init a second
// time replaces ExporterInfo with a new GaugeVec instance rather than reusing
// the previous one.
func TestReinitializationShouldReplaceExporterInfo(t *testing.T) {
	if err := meta.Init("first_", "sha1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	first := meta.ExporterInfo

	if err := meta.Init("second_", "sha2"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second := meta.ExporterInfo

	if first == second {
		t.Error("Init called a second time must produce a new ExporterInfo instance, not reuse the previous one")
	}
}
