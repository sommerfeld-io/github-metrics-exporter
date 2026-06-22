// Package workflow provides Prometheus gauges for workflow run and job conclusions.
package workflow

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/github"
)

// RunConclusion tracks workflow run conclusions from the most recent API page.
// Label dimensions: owner, repo, workflow, path, branch, conclusion. Value: always 1.
// It is initialized by calling Init.
var RunConclusion *prometheus.GaugeVec

// JobConclusion tracks workflow job conclusions from the most recent API page.
// Label dimensions: owner, repo, workflow, path, branch, job, conclusion. Value: always 1.
// It is initialized by calling Init.
var JobConclusion *prometheus.GaugeVec

// ErrNotInitialized is returned by Record when Init has not been called.
var ErrNotInitialized = errors.New("workflow.Record: Init must be called first")

// Init creates the RunConclusion and JobConclusion gauges using the given metric prefix.
// It returns an error if prefix is empty.
func Init(prefix string) error {
	if prefix == "" {
		return errors.New("workflow.Init: prefix must not be empty")
	}
	RunConclusion = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: prefix + "workflow_run_conclusion",
			Help: "1 if a workflow run with the given conclusion appears in the most recent API page.",
		},
		[]string{"owner", "repo", "workflow", "path", "branch", "conclusion"},
	)
	JobConclusion = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: prefix + "workflow_job_conclusion",
			Help: "1 if a workflow job with the given conclusion appears in the most recent API page.",
		},
		[]string{"owner", "repo", "workflow", "path", "branch", "job", "conclusion"},
	)
	return nil
}

// Record populates the run and job conclusion gauges for the given repository.
// Each unique (workflow, branch, conclusion) combination seen in runs is set to 1.
// It returns ErrNotInitialized if Init has not been called.
func Record(owner, repo string, runs []github.RunWithJobs) error {
	if RunConclusion == nil || JobConclusion == nil {
		return ErrNotInitialized
	}
	for _, r := range runs {
		RunConclusion.WithLabelValues(owner, repo, r.Run.Name, r.Run.Path, r.Run.HeadBranch, r.Run.Conclusion).Set(1)
		for _, j := range r.Jobs {
			JobConclusion.WithLabelValues(owner, repo, r.Run.Name, r.Run.Path, r.Run.HeadBranch, j.Name, j.Conclusion).Set(1)
		}
	}
	return nil
}
