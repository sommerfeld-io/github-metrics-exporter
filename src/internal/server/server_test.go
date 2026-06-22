package server_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/metrics"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/server"
)

// captureLog redirects the global slog to a buffer for the duration of the test.
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(orig) })
	return &buf
}

func collectNothing(_ context.Context) ([]server.Repository, error) {
	return nil, nil
}

func setup(t *testing.T) *http.ServeMux {
	t.Helper()
	reg := prometheus.NewRegistry()
	if err := metrics.Register(reg); err != nil {
		t.Fatalf("failed to register metrics: %v", err)
	}
	return server.New(9400, collectNothing)
}

// scrapeMetrics hits /metrics on mux, triggering the collect func and populating lastRepos.
func scrapeMetrics(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	mux.ServeHTTP(rec, req)
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

func TestRootEndpointShouldDisplayConfiguredPort(t *testing.T) {
	reg := prometheus.NewRegistry()
	if err := metrics.Register(reg); err != nil {
		t.Fatalf("failed to register metrics: %v", err)
	}
	mux := server.New(8080, collectNothing)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mux.ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), "8080") {
		t.Error("expected configured port 8080 to appear in the response body")
	}
}

func TestRootEndpointShouldNotDisplayUnconfiguredPort(t *testing.T) {
	reg := prometheus.NewRegistry()
	if err := metrics.Register(reg); err != nil {
		t.Fatalf("failed to register metrics: %v", err)
	}
	mux := server.New(8080, collectNothing)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mux.ServeHTTP(rec, req)

	if strings.Contains(rec.Body.String(), "9999") {
		t.Error("port 9999 must not appear when server is configured with 8080")
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

func TestMetricsEndpointShouldReturn500WhenCollectFails(t *testing.T) {
	reg := prometheus.NewRegistry()
	if err := metrics.Register(reg); err != nil {
		t.Fatalf("failed to register metrics: %v", err)
	}
	mux := server.New(9400, func(_ context.Context) ([]server.Repository, error) {
		return nil, errors.New("github API unavailable")
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500 when collect fails, got %d", rec.Code)
	}
}

func TestMetricsEndpointShouldNotReturn200WhenCollectFails(t *testing.T) {
	reg := prometheus.NewRegistry()
	if err := metrics.Register(reg); err != nil {
		t.Fatalf("failed to register metrics: %v", err)
	}
	mux := server.New(9400, func(_ context.Context) ([]server.Repository, error) {
		return nil, errors.New("github API unavailable")
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code == http.StatusOK {
		t.Error("status must not be 200 when collect fails")
	}
}

func TestRootEndpointShouldShowPreScrapeMessageBeforeFirstMetricsScrape(t *testing.T) {
	mux := setup(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mux.ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), "No data yet") {
		t.Error("expected pre-scrape message before first /metrics scrape")
	}
}

func TestRootEndpointShouldNotShowNoTargetsMessageBeforeFirstMetricsScrape(t *testing.T) {
	mux := setup(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mux.ServeHTTP(rec, req)

	if strings.Contains(rec.Body.String(), "No GitHub targets") {
		t.Error("must not show no-targets message before any scrape has happened")
	}
}

func TestRootEndpointShouldDisplayWarningWhenNoReposConfigured(t *testing.T) {
	mux := setup(t)
	scrapeMetrics(t, mux)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mux.ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), "No GitHub targets") {
		t.Error("expected no-targets warning when repo list is empty after scrape")
	}
}

func TestRootEndpointShouldNotDisplayRepoSectionWhenNoReposConfigured(t *testing.T) {
	mux := setup(t)
	scrapeMetrics(t, mux)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mux.ServeHTTP(rec, req)

	if strings.Contains(rec.Body.String(), "<h2>") {
		t.Error("owner heading must not appear when repo list is empty")
	}
}

func TestRootEndpointShouldDisplayOwnerHeadingWhenReposPresent(t *testing.T) {
	reg := prometheus.NewRegistry()
	if err := metrics.Register(reg); err != nil {
		t.Fatalf("failed to register metrics: %v", err)
	}
	repos := []server.Repository{
		{Owner: "test-org", Name: "repo1", Accessible: true},
	}
	mux := server.New(9400, func(_ context.Context) ([]server.Repository, error) {
		return repos, nil
	})
	scrapeMetrics(t, mux)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mux.ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), "test-org") {
		t.Error("expected owner heading 'test-org' in response body")
	}
}

func TestRootEndpointShouldDisplayRepoNameWhenPresent(t *testing.T) {
	reg := prometheus.NewRegistry()
	if err := metrics.Register(reg); err != nil {
		t.Fatalf("failed to register metrics: %v", err)
	}
	repos := []server.Repository{
		{Owner: "test-org", Name: "my-repo", Accessible: true},
	}
	mux := server.New(9400, func(_ context.Context) ([]server.Repository, error) {
		return repos, nil
	})
	scrapeMetrics(t, mux)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mux.ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), "my-repo") {
		t.Error("expected repository name 'my-repo' in response body")
	}
}

func TestRootEndpointShouldDisplayCorrectBadgeForAccessibility(t *testing.T) {
	tests := []struct {
		name        string
		accessible  bool
		wantPresent string
		wantAbsent  string
	}{
		{
			name:        "accessible repo shows accessible badge",
			accessible:  true,
			wantPresent: `class="badge accessible"`,
			wantAbsent:  `class="badge inaccessible"`,
		},
		{
			name:        "inaccessible repo shows inaccessible badge",
			accessible:  false,
			wantPresent: `class="badge inaccessible"`,
			wantAbsent:  `class="badge accessible"`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reg := prometheus.NewRegistry()
			if err := metrics.Register(reg); err != nil {
				t.Fatalf("failed to register metrics: %v", err)
			}
			repos := []server.Repository{
				{Owner: "test-org", Name: "repo", Accessible: tc.accessible},
			}
			mux := server.New(9400, func(_ context.Context) ([]server.Repository, error) {
				return repos, nil
			})
			scrapeMetrics(t, mux)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			mux.ServeHTTP(rec, req)
			body := rec.Body.String()
			if !strings.Contains(body, tc.wantPresent) {
				t.Errorf("expected %q in response body", tc.wantPresent)
			}
			if strings.Contains(body, tc.wantAbsent) {
				t.Errorf("must not contain %q in response body", tc.wantAbsent)
			}
		})
	}
}

func TestMetricsEndpointShouldLogScrapeStarted(t *testing.T) {
	buf := captureLog(t)
	mux := setup(t)
	scrapeMetrics(t, mux)

	if !strings.Contains(buf.String(), "metrics: scrape started") {
		t.Error("expected log entry 'metrics: scrape started' when /metrics is requested")
	}
}

func TestMetricsEndpointShouldNotSkipScrapeStartedLog(t *testing.T) {
	buf := captureLog(t)
	mux := setup(t)
	scrapeMetrics(t, mux)

	if strings.Contains(buf.String(), "metrics: scrape started") == false {
		t.Error("scrape start log must not be absent when /metrics is requested")
	}
}

func TestMetricsEndpointShouldLogScrapeCompleted(t *testing.T) {
	buf := captureLog(t)
	mux := setup(t)
	scrapeMetrics(t, mux)

	if !strings.Contains(buf.String(), "metrics: scrape completed") {
		t.Error("expected log entry 'metrics: scrape completed' when /metrics request finishes")
	}
}

func TestMetricsEndpointShouldNotSkipScrapeCompletedLog(t *testing.T) {
	buf := captureLog(t)
	mux := setup(t)
	scrapeMetrics(t, mux)

	if strings.Contains(buf.String(), "metrics: scrape completed") == false {
		t.Error("scrape completion log must not be absent when /metrics is requested")
	}
}

func TestNonMetricsEndpointShouldNotLogScrapeStarted(t *testing.T) {
	buf := captureLog(t)
	mux := setup(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	mux.ServeHTTP(rec, req)

	if strings.Contains(buf.String(), "metrics: scrape started") {
		t.Error("scrape start log must not be written for non-metrics endpoints")
	}
}
