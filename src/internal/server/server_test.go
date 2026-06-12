package server_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/server"
)

func setup(t *testing.T) *http.ServeMux {
	t.Helper()
	reg := prometheus.NewRegistry()
	if err := metrics.Register(reg); err != nil {
		t.Fatalf("failed to register metrics: %v", err)
	}
	return server.New()
}

func TestRootEndpointShouldReturnHTML(t *testing.T) {
	mux := setup(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("expected Content-Type to contain text/html, got %q", ct)
	}
}

func TestRootEndpointShouldContainHeadline(t *testing.T) {
	mux := setup(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mux.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "GitHub Metrics Exporter") {
		t.Error("expected headline 'GitHub Metrics Exporter' in response body")
	}
}

func TestRootEndpointShouldContainGitHubLink(t *testing.T) {
	mux := setup(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mux.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "https://github.com/sommerfeld-io/github-metrics-exporter/") {
		t.Error("expected GitHub repository link in response body")
	}
	if !strings.Contains(body, "on GitHub") {
		t.Error("expected link label 'on GitHub' in response body")
	}
	if !strings.Contains(body, `target="_blank"`) {
		t.Error("expected GitHub link to open in a new tab (target=_blank)")
	}
	if !strings.Contains(body, `rel="noopener noreferrer"`) {
		t.Error("expected GitHub link to have rel=noopener noreferrer")
	}
}

func TestRootEndpointShouldContainMetricsLink(t *testing.T) {
	mux := setup(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mux.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, `href="/metrics"`) {
		t.Error("expected link to /metrics in response body")
	}
}

func TestRootEndpointShouldContainHealthzLink(t *testing.T) {
	mux := setup(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mux.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, `href="/healthz"`) {
		t.Error("expected link to /healthz in response body")
	}
}

func TestRootEndpointShouldContainCommitSHA(t *testing.T) {
	mux := setup(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mux.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, metrics.CommitSHA) {
		t.Errorf("expected commit SHA %q in response body", metrics.CommitSHA)
	}
}

func TestRootEndpointShouldContainDarkTheme(t *testing.T) {
	mux := setup(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mux.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "background-color") {
		t.Error("expected dark theme CSS (background-color) in response body")
	}
}

func TestInvalidRouteShouldReturn404(t *testing.T) {
	mux := setup(t)
	for _, path := range []string{"/invalid-route", "/wrong/path"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("path %q: expected status 404, got %d", path, rec.Code)
		}
	}
}

func TestHealthzEndpointShouldReturnOK(t *testing.T) {
	mux := setup(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Errorf("expected Content-Type to contain text/plain, got %q", ct)
	}
}

func TestMetricsEndpointShouldReturnPlainText(t *testing.T) {
	mux := setup(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Errorf("expected Content-Type to contain text/plain, got %q", ct)
	}
}
