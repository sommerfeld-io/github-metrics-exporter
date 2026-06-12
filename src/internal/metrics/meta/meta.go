package meta

import "github.com/prometheus/client_golang/prometheus"

// ExporterInfo is the gauge that exposes build metadata about the exporter itself.
// It is initialised by calling Init and carries a constant value of 1.
var ExporterInfo *prometheus.GaugeVec

// Init creates the ExporterInfo gauge using the provided metric prefix, commit SHA,
// and version string, and sets its value to 1 so it is present on /metrics immediately.
func Init(prefix, commitSHA, version string) {
	ExporterInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: prefix + "exporter_info",
			Help: "A constant metric with a value of 1 containing build metadata labels.",
		},
		[]string{"commit_sha", "version"},
	)
	ExporterInfo.WithLabelValues(commitSHA, version).Set(1)
}
