package workflow

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/github"
)

func resetGlobals() {
	RunConclusion = nil
	JobConclusion = nil
}

func gaugeValue(t *testing.T, g *prometheus.GaugeVec, labelValues ...string) float64 {
	t.Helper()
	var m dto.Metric
	if err := g.WithLabelValues(labelValues...).Write(&m); err != nil {
		t.Fatalf("failed to read gauge: %v", err)
	}
	return m.Gauge.GetValue()
}

func TestInitShouldReturnErrorForEmptyPrefix(t *testing.T) {
	err := Init("")
	if err == nil {
		t.Error("expected error for empty prefix, got nil")
	}
}

func TestInitShouldNotReturnErrorForValidPrefix(t *testing.T) {
	t.Cleanup(resetGlobals)
	err := Init("ghme_")
	if err != nil {
		t.Errorf("expected no error for valid prefix, got %v", err)
	}
}

func TestInitShouldSetRunConclusionToNonNil(t *testing.T) {
	t.Cleanup(resetGlobals)
	if err := Init("ghme_"); err != nil {
		t.Fatalf("unexpected Init error: %v", err)
	}
	if RunConclusion == nil {
		t.Error("RunConclusion must not be nil after Init")
	}
}

func TestInitShouldSetJobConclusionToNonNil(t *testing.T) {
	t.Cleanup(resetGlobals)
	if err := Init("ghme_"); err != nil {
		t.Fatalf("unexpected Init error: %v", err)
	}
	if JobConclusion == nil {
		t.Error("JobConclusion must not be nil after Init")
	}
}

func TestInitShouldLeaveRunConclusionNilOnEmptyPrefix(t *testing.T) {
	t.Cleanup(resetGlobals)
	_ = Init("")
	if RunConclusion != nil {
		t.Error("RunConclusion must remain nil when Init fails")
	}
}

func TestRecordShouldReturnErrNotInitializedBeforeInit(t *testing.T) {
	t.Cleanup(resetGlobals)
	err := Record("owner", "repo", nil)
	if err == nil {
		t.Error("expected ErrNotInitialized before Init, got nil")
	}
}

func TestRecordShouldNotReturnErrorAfterInit(t *testing.T) {
	t.Cleanup(resetGlobals)
	if err := Init("ghme_"); err != nil {
		t.Fatalf("unexpected Init error: %v", err)
	}
	err := Record("owner", "repo", nil)
	if err != nil {
		t.Errorf("expected no error for empty run list, got %v", err)
	}
}

func TestRecordShouldSetRunConclusionGaugeToOne(t *testing.T) {
	t.Cleanup(resetGlobals)
	if err := Init("ghme_"); err != nil {
		t.Fatalf("unexpected Init error: %v", err)
	}
	runs := []github.RunWithJobs{
		{Run: github.WorkflowRun{Name: "CI", Path: ".github/workflows/ci.yml", HeadBranch: "main", Conclusion: "success"}},
	}
	if err := Record("owner", "repo", runs); err != nil {
		t.Fatalf("unexpected Record error: %v", err)
	}
	val := gaugeValue(t, RunConclusion, "owner", "repo", "CI", ".github/workflows/ci.yml", "main", "success")
	if val != 1 {
		t.Errorf("expected RunConclusion gauge to be 1, got %v", val)
	}
}

func TestRecordShouldNotLeaveRunConclusionGaugeAtZero(t *testing.T) {
	t.Cleanup(resetGlobals)
	if err := Init("ghme_"); err != nil {
		t.Fatalf("unexpected Init error: %v", err)
	}
	runs := []github.RunWithJobs{
		{Run: github.WorkflowRun{Name: "CI", Path: ".github/workflows/ci.yml", HeadBranch: "main", Conclusion: "success"}},
	}
	if err := Record("owner", "repo", runs); err != nil {
		t.Fatalf("unexpected Record error: %v", err)
	}
	val := gaugeValue(t, RunConclusion, "owner", "repo", "CI", ".github/workflows/ci.yml", "main", "success")
	if val == 0 {
		t.Error("RunConclusion gauge must not be 0 after Record (zero means the value was never set)")
	}
}

func TestRecordShouldSetJobConclusionGaugeToOne(t *testing.T) {
	t.Cleanup(resetGlobals)
	if err := Init("ghme_"); err != nil {
		t.Fatalf("unexpected Init error: %v", err)
	}
	runs := []github.RunWithJobs{
		{
			Run:  github.WorkflowRun{Name: "CI", Path: ".github/workflows/ci.yml", HeadBranch: "main", Conclusion: "success"},
			Jobs: []github.Job{{Name: "build", Conclusion: "success"}},
		},
	}
	if err := Record("owner", "repo", runs); err != nil {
		t.Fatalf("unexpected Record error: %v", err)
	}
	val := gaugeValue(t, JobConclusion, "owner", "repo", "CI", ".github/workflows/ci.yml", "main", "build", "success")
	if val != 1 {
		t.Errorf("expected JobConclusion gauge to be 1, got %v", val)
	}
}

func TestRecordShouldNotLeaveJobConclusionGaugeAtZero(t *testing.T) {
	t.Cleanup(resetGlobals)
	if err := Init("ghme_"); err != nil {
		t.Fatalf("unexpected Init error: %v", err)
	}
	runs := []github.RunWithJobs{
		{
			Run:  github.WorkflowRun{Name: "CI", Path: ".github/workflows/ci.yml", HeadBranch: "main", Conclusion: "success"},
			Jobs: []github.Job{{Name: "build", Conclusion: "success"}},
		},
	}
	if err := Record("owner", "repo", runs); err != nil {
		t.Fatalf("unexpected Record error: %v", err)
	}
	val := gaugeValue(t, JobConclusion, "owner", "repo", "CI", ".github/workflows/ci.yml", "main", "build", "success")
	if val == 0 {
		t.Error("JobConclusion gauge must not be 0 after Record (zero means the value was never set)")
	}
}

func TestRecordShouldNotPanicForRunWithEmptyJobList(t *testing.T) {
	t.Cleanup(resetGlobals)
	if err := Init("ghme_"); err != nil {
		t.Fatalf("unexpected Init error: %v", err)
	}
	runs := []github.RunWithJobs{
		{
			Run:  github.WorkflowRun{Name: "CI", HeadBranch: "main", Conclusion: "success"},
			Jobs: []github.Job{},
		},
	}
	if err := Record("owner", "repo", runs); err != nil {
		t.Errorf("expected no error for run with empty job list, got %v", err)
	}
}

func TestRecordShouldNotReturnErrorForEmptyRunList(t *testing.T) {
	t.Cleanup(resetGlobals)
	if err := Init("ghme_"); err != nil {
		t.Fatalf("unexpected Init error: %v", err)
	}
	if err := Record("owner", "repo", []github.RunWithJobs{}); err != nil {
		t.Errorf("expected no error for empty run list, got %v", err)
	}
}

func TestRecordShouldIncludePathInRunConclusionLabel(t *testing.T) {
	t.Cleanup(resetGlobals)
	if err := Init("ghme_"); err != nil {
		t.Fatalf("unexpected Init error: %v", err)
	}
	runs := []github.RunWithJobs{
		{Run: github.WorkflowRun{Name: "CI", Path: ".github/workflows/ci.yml", HeadBranch: "main", Conclusion: "success"}},
	}
	if err := Record("owner", "repo", runs); err != nil {
		t.Fatalf("unexpected Record error: %v", err)
	}
	val := gaugeValue(t, RunConclusion, "owner", "repo", "CI", ".github/workflows/ci.yml", "main", "success")
	if val != 1 {
		t.Errorf("expected RunConclusion gauge to be 1 with path label, got %v", val)
	}
}

func TestRecordShouldIncludePathInJobConclusionLabel(t *testing.T) {
	t.Cleanup(resetGlobals)
	if err := Init("ghme_"); err != nil {
		t.Fatalf("unexpected Init error: %v", err)
	}
	runs := []github.RunWithJobs{
		{
			Run:  github.WorkflowRun{Name: "CI", Path: ".github/workflows/ci.yml", HeadBranch: "main", Conclusion: "success"},
			Jobs: []github.Job{{Name: "build", Conclusion: "success"}},
		},
	}
	if err := Record("owner", "repo", runs); err != nil {
		t.Fatalf("unexpected Record error: %v", err)
	}
	val := gaugeValue(t, JobConclusion, "owner", "repo", "CI", ".github/workflows/ci.yml", "main", "build", "success")
	if val != 1 {
		t.Errorf("expected JobConclusion gauge to be 1 with path label, got %v", val)
	}
}
