# Package: `server`

Builds the HTTP server for the exporter.

Constructs the HTTP server mux for the exporter. `New` registers three routes: `/` serves a dark-themed HTML landing page with project links and the build commit SHA rendered from a Go template; `/metrics` delegates to the standard `promhttp.Handler` to expose the default Prometheus registry; and `/healthz` returns a plain-text `ok` response for liveness checks. Any request to an unregistered path is handled by the `/` catch-all, which responds with a `404` status.
