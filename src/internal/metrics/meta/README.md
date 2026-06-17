# Package: `meta`

Owns the `ghme_exporter_info` gauge, a constant metric with a value of `1` that carries build metadata as labels. `Init` creates the gauge using the provided metric prefix and commit SHA and sets its value immediately so the metric is present on the first scrape without requiring any scrape-time computation. The resulting gauge is exposed via the package-level `ExporterInfo` variable, which callers register with a `prometheus.Registerer`.
