package server

import (
	"html/template"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics"
)

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>GitHub Metrics Exporter</title>
  <style>
    body { background-color: #121212; color: #e0e0e0; font-family: sans-serif; padding: 2rem; }
    a { color: #90caf9; }
  </style>
</head>
<body>
  <h1>GitHub Metrics Exporter</h1>
  <p>Visit us <a href="https://github.com/sommerfeld-io/github-metrics-exporter/" target="_blank" rel="noopener noreferrer">on GitHub</a></p>
  <ul>
    <li><a href="/metrics">Metrics</a></li>
    <li><a href="/healthz">Health Check</a></li>
  </ul>
  <p>Build commit SHA: {{.CommitSHA}}</p>
</body>
</html>`

var indexTmpl = template.Must(template.New("index").Parse(indexHTML))

func indexEndpointHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = indexTmpl.Execute(w, struct{ CommitSHA string }{metrics.CommitSHA})
}

func healthzEndpointHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok"))
}

// New sets up the HTTP server with all application routes: the index page,
// the Prometheus metrics endpoint, and the health check endpoint.
func New() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexEndpointHandler)
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", healthzEndpointHandler)
	return mux
}
