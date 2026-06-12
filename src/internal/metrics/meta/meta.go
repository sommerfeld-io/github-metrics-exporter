package meta

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

// ExporterInfo is the gauge that exposes build metadata about the exporter itself.
// It is initialised by calling Init and carries a constant value of 1.
var ExporterInfo *prometheus.GaugeVec

// Init creates the ExporterInfo gauge using the provided metric prefix and commit SHA,
// and sets its value to 1 so it is present on /metrics immediately.
// It returns an error if prefix is empty.
func Init(prefix, commitSHA string) error {
	if prefix == "" {
		return errors.New("meta.Init: prefix must not be empty")
	}
	ExporterInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: prefix + "exporter_info",
			Help: "A constant metric with a value of 1 containing build metadata labels.",
		},
		[]string{"commit_sha"},
	)
	ExporterInfo.WithLabelValues(commitSHA).Set(1)
	return nil
}
