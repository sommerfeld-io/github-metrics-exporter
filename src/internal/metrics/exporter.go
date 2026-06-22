package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics/meta"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics/repository"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics/workflow"
)

// Register registers all core exporter metrics with the given registry and
// sets their initial values so that they are present on the /metrics endpoint
// immediately upon startup.
func Register(reg prometheus.Registerer) error {
	if err := meta.Init(MetricPrefix, CommitSHA); err != nil {
		return fmt.Errorf("metrics.Register: %w", err)
	}
	reg.MustRegister(meta.ExporterInfo)

	if err := repository.Init(MetricPrefix); err != nil {
		return fmt.Errorf("metrics.Register: %w", err)
	}
	reg.MustRegister(repository.Accessible)

	if err := workflow.Init(MetricPrefix); err != nil {
		return fmt.Errorf("metrics.Register: %w", err)
	}
	reg.MustRegister(workflow.RunConclusion, workflow.JobConclusion)
	return nil
}
