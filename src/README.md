# github-metrics-exporter

Entry point for the GitHub Metrics Exporter binary. `main` is intentionally thin: it registers
all Prometheus metrics with the default registry, constructs the HTTP server mux, and calls
`http.ListenAndServe` on port `9400`. All application logic lives in the `internal/` packages;
`main` only wires them together and handles fatal startup errors.
