package repository_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics/repository"
)

func initAndRegister(t *testing.T, prefix string) *prometheus.Registry {
	t.Helper()
	if err := repository.Init(prefix); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	reg := prometheus.NewRegistry()
	reg.MustRegister(repository.Accessible)
	return reg
}

func gatherValue(t *testing.T, reg *prometheus.Registry, owner, name string) float64 {
	t.Helper()
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather failed: %v", err)
	}
	for _, mf := range mfs {
		for _, m := range mf.GetMetric() {
			labels := labelsToMap(m.GetLabel())
			if labels["owner"] == owner && labels["repo"] == name {
				return m.GetGauge().GetValue()
			}
		}
	}
	t.Fatalf("metric for owner=%q repo=%q not found", owner, name)
	return 0
}

func labelsToMap(labels []*dto.LabelPair) map[string]string {
	m := make(map[string]string, len(labels))
	for _, l := range labels {
		m[l.GetName()] = l.GetValue()
	}
	return m
}

func TestInitShouldAssignAccessibleGauge(t *testing.T) {
	if err := repository.Init("test_"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repository.Accessible == nil {
		t.Error("Accessible must not be nil after Init")
	}
}

func TestInitShouldNotLeaveAccessibleNil(t *testing.T) {
	if err := repository.Init("test_"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repository.Accessible == nil {
		t.Error("Accessible must be assigned after Init (must not remain nil)")
	}
}

func TestInitShouldReturnErrorOnEmptyPrefix(t *testing.T) {
	err := repository.Init("")
	if err == nil {
		t.Error("Init must return an error when prefix is empty")
	}
}

func TestSetShouldRecordAccessibleAs1(t *testing.T) {
	reg := initAndRegister(t, "test_")
	if err := repository.Set("test-org", "repo1", true); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	val := gatherValue(t, reg, "test-org", "repo1")
	if val != 1.0 {
		t.Errorf("expected 1.0 for accessible repo, got %f", val)
	}
}

func TestSetShouldNotRecordAccessibleAs0(t *testing.T) {
	reg := initAndRegister(t, "test_")
	if err := repository.Set("test-org", "repo1", true); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	val := gatherValue(t, reg, "test-org", "repo1")
	if val == 0.0 {
		t.Error("accessible repo must not have value 0.0 (that would be the zero value, not an explicit setting)")
	}
}

func TestSetShouldRecordInaccessibleAs0(t *testing.T) {
	reg := initAndRegister(t, "test_")
	if err := repository.Set("test-org", "repo1", false); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	val := gatherValue(t, reg, "test-org", "repo1")
	if val != 0.0 {
		t.Errorf("expected 0.0 for inaccessible repo, got %f", val)
	}
}

func TestSetShouldNotRecordInaccessibleAs1(t *testing.T) {
	reg := initAndRegister(t, "test_")
	if err := repository.Set("test-org", "repo1", false); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	val := gatherValue(t, reg, "test-org", "repo1")
	if val == 1.0 {
		t.Error("inaccessible repo must not have value 1.0")
	}
}

func TestSetShouldReturnErrorWhenNotInitialized(t *testing.T) {
	original := repository.Accessible
	repository.Accessible = nil
	t.Cleanup(func() { repository.Accessible = original })

	err := repository.Set("org", "repo", true)
	if err == nil {
		t.Error("Set must return an error when Accessible is nil (Init not called)")
	}
}

func TestSetShouldProduceCorrectValueForAccessibility(t *testing.T) {
	cases := []struct {
		name       string
		accessible bool
		want       float64
	}{
		{"accessible", true, 1.0},
		{"inaccessible", false, 0.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reg := initAndRegister(t, "test_")
			if err := repository.Set("owner", "repo", tc.accessible); err != nil {
				t.Fatalf("Set failed: %v", err)
			}
			val := gatherValue(t, reg, "owner", "repo")
			if val != tc.want {
				t.Errorf("accessible=%v: expected %f, got %f", tc.accessible, tc.want, val)
			}
		})
	}
}

func TestSetShouldProduceLabelOwnerAndRepo(t *testing.T) {
	initAndRegister(t, "test_")
	if err := repository.Set("my-org", "my-repo", true); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	reg := prometheus.NewRegistry()
	reg.MustRegister(repository.Accessible)
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather failed: %v", err)
	}
	if len(mfs) == 0 || len(mfs[0].GetMetric()) == 0 {
		t.Fatal("expected at least one metric")
	}
	labels := labelsToMap(mfs[0].GetMetric()[0].GetLabel())
	if _, ok := labels["owner"]; !ok {
		t.Error("expected label 'owner' to be present")
	}
	if _, ok := labels["repo"]; !ok {
		t.Error("expected label 'repo' to be present")
	}
}

func TestSetShouldProduceExactlyTwoLabels(t *testing.T) {
	initAndRegister(t, "test_")
	if err := repository.Set("my-org", "my-repo", true); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	reg := prometheus.NewRegistry()
	reg.MustRegister(repository.Accessible)
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather failed: %v", err)
	}
	if len(mfs) == 0 || len(mfs[0].GetMetric()) == 0 {
		t.Fatal("expected at least one metric")
	}
	n := len(mfs[0].GetMetric()[0].GetLabel())
	if n != 2 {
		t.Errorf("expected exactly 2 labels (owner, repo), got %d", n)
	}
}
