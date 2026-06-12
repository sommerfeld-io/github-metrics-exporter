package metrics

import "github.com/prometheus/client_golang/prometheus"

// CommitSHA holds the build-time commit SHA. It defaults to "development" and
// may be overridden via a linker flag at build time.
var CommitSHA = "development"

// version is the semantic version of the exporter.
const version = "0.0.0"

var exporterInfo = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "ghme_exporter_info",
		Help: "A constant metric with a value of 1 containing build metadata labels.",
	},
	[]string{"commit_sha", "version"},
)

// Register registers all core exporter meta-metrics with the given registry and
// sets their initial values so that they are present on the /metrics endpoint
// immediately upon startup.
func Register(reg prometheus.Registerer) {
	reg.MustRegister(exporterInfo)
	exporterInfo.WithLabelValues(CommitSHA, version).Set(1)
}
