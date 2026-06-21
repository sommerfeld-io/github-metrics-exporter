package server

import (
	"html/template"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics"
)

// Repository is the server's view of a discovered GitHub repository.
type Repository struct {
	Owner      string
	Name       string
	Accessible bool
}

// repoGroup groups repositories under a common owner name for template rendering.
type repoGroup struct {
	Owner string
	Repos []Repository
}

// groupRepos assembles a slice of repoGroup from an already-sorted Repository slice.
// The input is expected to be sorted by Owner then Name (as returned by github.Discover).
func groupRepos(repos []Repository) []repoGroup {
	var groups []repoGroup
	for _, r := range repos {
		if len(groups) == 0 || groups[len(groups)-1].Owner != r.Owner {
			groups = append(groups, repoGroup{Owner: r.Owner})
		}
		groups[len(groups)-1].Repos = append(groups[len(groups)-1].Repos, r)
	}
	return groups
}

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
	.warning { background-color: #3a2000; border: 1px solid #ff6f00; padding: 1rem; border-radius: 4px; margin-bottom: 1rem; }
	.badge { color: #ffffff; font-size: 0.75rem; padding: 0.1rem 0.4rem; border-radius: 4px; margin-left: 0.5rem; }
	.badge.accessible { background-color: #1b5e20; }
	.badge.inaccessible { background-color: #b71c1c; }
  </style>
</head>
<body>
  <header>
    <h1>GitHub Metrics Exporter</h1>
    <ul>
      <li>Visit us <a href="https://github.com/sommerfeld-io/github-metrics-exporter/" target="_blank" rel="noopener noreferrer">on GitHub</a></li>
      <li>Build commit SHA: <code>{{.CommitSHA}}</code> (<a href="https://github.com/sommerfeld-io/github-metrics-exporter/tree/{{.CommitSHA}}" target="_blank" rel="noopener noreferrer">browse files for this commit</a>)</li>
      <li>Listening on port: <code>{{.Port}}</code></li>
    </ul>
  </header>

  <main>
	<ul>
		<li><a href="/metrics">Metrics</a></li>
		<li><a href="/healthz">Health Check</a></li>
	</ul>

	{{if not .HasRepos}}
	<div class="warning">No GitHub targets are configured. Add organizations or users to the config file.</div>
	{{else}}
	{{range .RepoGroups}}
	<h2>{{.Owner}}</h2>
	<ul>
		{{range .Repos}}
		<li>{{.Name}}{{if .Accessible}}<span class="badge accessible">accessible</span>{{else}}<span class="badge inaccessible">inaccessible</span>{{end}}</li>
		{{end}}
	</ul>
	{{end}}
	{{end}}
  </main>
</body>
</html>`

var indexTmpl = template.Must(template.New("index").Funcs(template.FuncMap{
	"not": func(b bool) bool { return !b },
}).Parse(indexHTML))

func indexHandler(port int, repos []Repository) http.HandlerFunc {
	groups := groupRepos(repos)
	data := struct {
		CommitSHA  string
		Port       int
		HasRepos   bool
		RepoGroups []repoGroup
	}{
		CommitSHA:  metrics.CommitSHA,
		Port:       port,
		HasRepos:   len(repos) > 0,
		RepoGroups: groups,
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "404 page not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = indexTmpl.Execute(w, data)
	}
}

func healthzEndpointHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok"))
}

// New sets up the HTTP server with all application routes: the index page,
// the Prometheus metrics endpoint, and the health check endpoint.
// repos is the list of discovered repositories to display on the index page.
// The port is rendered on the root page so operators can confirm the active configuration.
func New(port int, repos []Repository) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler(port, repos))
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", healthzEndpointHandler)
	return mux
}
