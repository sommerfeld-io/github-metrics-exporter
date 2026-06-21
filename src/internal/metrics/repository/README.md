# Package: `repository`

Provides the `ghme_repository_accessible` Prometheus gauge metric.

Call `Init(prefix)` once at startup to create the `GaugeVec`. Then call `Set(owner, name, accessible)` for each discovered repository to record its accessibility state: `1` for accessible, `0` for inaccessible. Both functions return an error rather than panicking on misuse.
