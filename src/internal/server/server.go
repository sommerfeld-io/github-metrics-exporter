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
    body {
      background-color: #121212;
      color: #e0e0e0;
      font-family: sans-serif;
      padding: 2rem 4rem 2rem 4rem;
      margin: 0;
    }
	header {
		border: 2px solid #444444;
		padding: 1rem 2rem 1rem 2rem;
	}
	main {
		border: 2px solid transparent;
		padding: 1rem 2rem 1rem 2rem;
		margin-bottom: 2rem;
	}
    a { color: #90caf9; text-decoration: none; }
    a:hover { color: #ffb74d; text-decoration: none; }
	code { background-color: #333333; padding: 0.2rem 0.4rem; border-radius: 4px; }
  </style>
</head>
<body>
  <header>
    <h1>GitHub Metrics Exporter</h1>
    <ul>
      <li>Visit us <a href="https://github.com/sommerfeld-io/github-metrics-exporter/" target="_blank" rel="noopener noreferrer">on GitHub</a></li>
      <li>Build commit SHA: <code>{{.CommitSHA}}</code> (<a href="https://github.com/sommerfeld-io/github-metrics-exporter/tree/{{.CommitSHA}}" target="_blank" rel="noopener noreferrer">browse files for this commit</a>)</li>
    </ul>
  </header>

  <main>
	<ul>
		<li><a href="/metrics">Metrics</a></li>
		<li><a href="/healthz">Health Check</a></li>
	</ul>
  </main>
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
