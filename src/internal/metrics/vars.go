package metrics

// MetricPrefix is the shared prefix for all metrics exposed by this exporter.
var MetricPrefix = "ghme_"

// CommitSHA holds the build-time commit SHA. It defaults to "development" and
// may be overridden via a linker flag at build time.
var CommitSHA = "development"
