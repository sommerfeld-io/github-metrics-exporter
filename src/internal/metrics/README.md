# Package: `metrics`

Coordinates metric registration for the exporter. The package declares two shared variables used across all metric definitions: `MetricPrefix` (`ghme_`), which is prepended to every metric name to namespace them consistently, and `CommitSHA`, which defaults to `"development"` and is overridden at build time via a linker flag. The exported `Register` function initializes all sub-packages and registers their gauges with the provided `prometheus.Registerer`, making every metric immediately visible on the `/metrics` endpoint at startup.
