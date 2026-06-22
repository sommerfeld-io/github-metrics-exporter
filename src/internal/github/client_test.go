package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// newTestClient creates a Client pointed at a local test HTTP server.
// It replaces the BaseURL on the underlying go-github client so all API calls
// go to the mock server instead of api.github.com.
func newTestClient(t *testing.T, handler http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	c := NewWithToken("test-token")
	u, err := url.Parse(srv.URL + "/")
	if err != nil {
		t.Fatalf("setup: parse test server URL: %v", err)
	}
	c.gh.BaseURL = u
	return c, srv
}

// repoJSON returns a minimal GitHub repository JSON object for use in mock responses.
func repoJSON(owner, name string) map[string]interface{} {
	return map[string]interface{}{
		"name":      name,
		"full_name": fmt.Sprintf("%s/%s", owner, name),
		"owner":     map[string]interface{}{"login": owner},
		"html_url":  fmt.Sprintf("https://github.com/%s/%s", owner, name),
	}
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func TestDiscoverShouldReturnEmptySliceWhenBothOrgsAndUsersAreEmpty(t *testing.T) {
	c, _ := newTestClient(t, http.NewServeMux())
	repos, _ := c.Discover(context.Background(), nil, nil)
	if len(repos) != 0 {
		t.Errorf("expected empty slice, got %d repos", len(repos))
	}
}

func TestDiscoverShouldNotReturnErrorWhenBothOrgsAndUsersAreEmpty(t *testing.T) {
	c, _ := newTestClient(t, http.NewServeMux())
	_, err := c.Discover(context.Background(), nil, nil)
	if err != nil {
		t.Errorf("expected nil error for empty targets, got %v", err)
	}
}

func TestDiscoverShouldReturnRepositoriesForOrg(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/test-org/repos", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []interface{}{
			repoJSON("test-org", "repo1"),
			repoJSON("test-org", "repo2"),
		})
	})
	c, _ := newTestClient(t, mux)

	repos, err := c.Discover(context.Background(), []string{"test-org"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 2 {
		t.Errorf("expected 2 repos, got %d", len(repos))
	}
}

func TestDiscoverShouldNotReturnEmptySliceWhenOrgHasRepos(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/test-org/repos", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []interface{}{repoJSON("test-org", "repo1")})
	})
	c, _ := newTestClient(t, mux)

	repos, err := c.Discover(context.Background(), []string{"test-org"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) == 0 {
		t.Error("expected non-empty slice when org has repos")
	}
}

func TestDiscoverShouldReturnRepositoriesForUser(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/users/test-user/repos", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []interface{}{
			repoJSON("test-user", "repo1"),
		})
	})
	c, _ := newTestClient(t, mux)

	repos, err := c.Discover(context.Background(), nil, []string{"test-user"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 1 {
		t.Errorf("expected 1 repo, got %d", len(repos))
	}
}

func TestDiscoverShouldHandlePagination(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/test-org/repos", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		if page == "2" {
			writeJSON(w, http.StatusOK, []interface{}{
				repoJSON("test-org", "repo3"),
				repoJSON("test-org", "repo4"),
			})
			return
		}
		// Page 1: include Link header pointing to page 2
		serverURL := "http://" + r.Host
		w.Header().Set("Link", fmt.Sprintf(`<%s/orgs/test-org/repos?page=2>; rel="next"`, serverURL))
		writeJSON(w, http.StatusOK, []interface{}{
			repoJSON("test-org", "repo1"),
			repoJSON("test-org", "repo2"),
		})
	})
	c, _ := newTestClient(t, mux)

	repos, err := c.Discover(context.Background(), []string{"test-org"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 4 {
		t.Errorf("expected 4 repos (2 pages), got %d", len(repos))
	}
}

func TestDiscoverShouldMarkOrgInaccessibleOn403(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/forbidden-org/repos", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusForbidden, map[string]string{"message": "Forbidden"})
	})
	mux.HandleFunc("/orgs/ok-org/repos", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []interface{}{repoJSON("ok-org", "repo1")})
	})
	c, _ := newTestClient(t, mux)

	repos, err := c.Discover(context.Background(), []string{"forbidden-org", "ok-org"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// forbidden-org contributes 0 repos; ok-org contributes 1
	if len(repos) != 1 {
		t.Errorf("expected 1 repo (from ok-org), got %d", len(repos))
	}
}

func TestDiscoverShouldNotReturnErrorOn403(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/forbidden-org/repos", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusForbidden, map[string]string{"message": "Forbidden"})
	})
	c, _ := newTestClient(t, mux)

	_, err := c.Discover(context.Background(), []string{"forbidden-org"}, nil)
	if err != nil {
		t.Errorf("expected nil error on 403, got %v", err)
	}
}

func TestDiscoverShouldMarkOrgInaccessibleOn404(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/missing-org/repos", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]string{"message": "Not Found"})
	})
	mux.HandleFunc("/orgs/ok-org/repos", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []interface{}{repoJSON("ok-org", "repo1")})
	})
	c, _ := newTestClient(t, mux)

	repos, err := c.Discover(context.Background(), []string{"missing-org", "ok-org"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 1 {
		t.Errorf("expected 1 repo (from ok-org), got %d", len(repos))
	}
}

func TestDiscoverShouldNotReturnErrorOn404(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/missing-org/repos", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]string{"message": "Not Found"})
	})
	c, _ := newTestClient(t, mux)

	_, err := c.Discover(context.Background(), []string{"missing-org"}, nil)
	if err != nil {
		t.Errorf("expected nil error on 404, got %v", err)
	}
}

func TestDiscoverShouldReturnErrorOnServerError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/test-org/repos", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Internal Server Error"})
	})
	c, _ := newTestClient(t, mux)

	_, err := c.Discover(context.Background(), []string{"test-org"}, nil)
	if err == nil {
		t.Error("expected error on 500, got nil")
	}
}

func TestDiscoverShouldReturnReposSortedByOwnerThenName(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/z-org/repos", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []interface{}{
			repoJSON("z-org", "b-repo"),
			repoJSON("z-org", "a-repo"),
		})
	})
	mux.HandleFunc("/orgs/a-org/repos", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []interface{}{repoJSON("a-org", "c-repo")})
	})
	c, _ := newTestClient(t, mux)

	repos, err := c.Discover(context.Background(), []string{"z-org", "a-org"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 3 {
		t.Fatalf("expected 3 repos, got %d", len(repos))
	}
	if repos[0].Owner != "a-org" || repos[0].Name != "c-repo" {
		t.Errorf("expected first repo a-org/c-repo, got %s/%s", repos[0].Owner, repos[0].Name)
	}
	if repos[1].Owner != "z-org" || repos[1].Name != "a-repo" {
		t.Errorf("expected second repo z-org/a-repo, got %s/%s", repos[1].Owner, repos[1].Name)
	}
	if repos[2].Owner != "z-org" || repos[2].Name != "b-repo" {
		t.Errorf("expected third repo z-org/b-repo, got %s/%s", repos[2].Owner, repos[2].Name)
	}
}

func TestDiscoverShouldNotReturnUnsortedRepos(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/z-org/repos", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []interface{}{
			repoJSON("z-org", "b-repo"),
			repoJSON("z-org", "a-repo"),
		})
	})
	mux.HandleFunc("/orgs/a-org/repos", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []interface{}{repoJSON("a-org", "c-repo")})
	})
	c, _ := newTestClient(t, mux)

	repos, err := c.Discover(context.Background(), []string{"z-org", "a-org"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) >= 2 && repos[0].Owner == "z-org" {
		t.Error("repos must not be in unsorted order: z-org must not come before a-org")
	}
}

func TestDiscoverShouldMarkAllDiscoveredReposAccessible(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/test-org/repos", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []interface{}{repoJSON("test-org", "repo1")})
	})
	c, _ := newTestClient(t, mux)

	repos, err := c.Discover(context.Background(), []string{"test-org"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range repos {
		if !r.Accessible {
			t.Errorf("repo %s/%s must be accessible when successfully listed", r.Owner, r.Name)
		}
	}
}

func TestDiscoverShouldNotReturnNilForEmptyOrg(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/empty-org/repos", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []interface{}{})
	})
	c, _ := newTestClient(t, mux)

	repos, err := c.Discover(context.Background(), []string{"empty-org"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repos == nil {
		t.Error("Discover must return an empty (non-nil) slice for an org with no repos")
	}
}

// workflowRunJSON returns a minimal GitHub workflow run JSON object.
func workflowRunJSON(id int64, name, headBranch, actor, event, conclusion, createdAt, updatedAt string) map[string]interface{} {
	return map[string]interface{}{
		"id":          id,
		"name":        name,
		"head_branch": headBranch,
		"event":       event,
		"status":      "completed",
		"conclusion":  conclusion,
		"created_at":  createdAt,
		"updated_at":  updatedAt,
		"actor":       map[string]interface{}{"login": actor},
	}
}

// workflowRunsPageJSON wraps runs in the GitHub API envelope.
func workflowRunsPageJSON(runs []map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"total_count":   len(runs),
		"workflow_runs": runs,
	}
}

// jobJSON returns a minimal GitHub job JSON object.
func jobJSON(id int64, name, conclusion string) map[string]interface{} {
	return map[string]interface{}{
		"id":         id,
		"name":       name,
		"status":     "completed",
		"conclusion": conclusion,
	}
}

// jobsPageJSON wraps jobs in the GitHub API envelope.
func jobsPageJSON(jobs []map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"total_count": len(jobs),
		"jobs":        jobs,
	}
}

func TestWorkflowRunsShouldReturnRunsForRepo(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/actions/runs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, workflowRunsPageJSON([]map[string]interface{}{
			workflowRunJSON(1, "CI", "main", "alice", "push", "success", "2026-06-15T10:00:00Z", "2026-06-15T10:05:00Z"),
			workflowRunJSON(2, "CI", "main", "bob", "push", "failure", "2026-06-16T10:00:00Z", "2026-06-16T10:05:00Z"),
		}))
	})
	c, _ := newTestClient(t, mux)

	runs, err := c.WorkflowRuns(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("expected 2 runs, got %d", len(runs))
	}
}

func TestWorkflowRunsShouldNotReturnEmptySliceWhenRunsExist(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/actions/runs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, workflowRunsPageJSON([]map[string]interface{}{
			workflowRunJSON(1, "CI", "main", "alice", "push", "success", "2026-06-15T10:00:00Z", "2026-06-15T10:05:00Z"),
		}))
	})
	c, _ := newTestClient(t, mux)

	runs, err := c.WorkflowRuns(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runs) == 0 {
		t.Error("expected non-empty slice when repo has workflow runs")
	}
}

func TestWorkflowRunsShouldReturnEmptySliceWhenNoRunsExist(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/actions/runs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, workflowRunsPageJSON([]map[string]interface{}{}))
	})
	c, _ := newTestClient(t, mux)

	runs, err := c.WorkflowRuns(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runs) != 0 {
		t.Errorf("expected empty slice, got %d runs", len(runs))
	}
}

func TestWorkflowRunsShouldNotReturnNilWhenNoRunsExist(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/actions/runs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, workflowRunsPageJSON([]map[string]interface{}{}))
	})
	c, _ := newTestClient(t, mux)

	runs, err := c.WorkflowRuns(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runs == nil {
		t.Error("WorkflowRuns must return a non-nil slice even when the repo has no runs")
	}
}

func TestWorkflowRunsShouldPopulateRunScalarFields(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/actions/runs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, workflowRunsPageJSON([]map[string]interface{}{
			workflowRunJSON(42, "Build", "feature-x", "alice", "push", "success",
				"2026-06-15T10:00:00Z", "2026-06-15T10:05:00Z"),
		}))
	})
	c, _ := newTestClient(t, mux)

	runs, err := c.WorkflowRuns(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	r := runs[0]
	if r.ID != 42 {
		t.Errorf("expected ID 42, got %d", r.ID)
	}
	if r.Name != "Build" {
		t.Errorf("expected Name 'Build', got %q", r.Name)
	}
	if r.HeadBranch != "feature-x" {
		t.Errorf("expected HeadBranch 'feature-x', got %q", r.HeadBranch)
	}
	if r.Actor != "alice" {
		t.Errorf("expected Actor 'alice', got %q", r.Actor)
	}
	if r.Event != "push" {
		t.Errorf("expected Event 'push', got %q", r.Event)
	}
	if r.Conclusion != "success" {
		t.Errorf("expected Conclusion 'success', got %q", r.Conclusion)
	}
}

func TestWorkflowRunsShouldPopulateRunTimestamps(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/actions/runs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, workflowRunsPageJSON([]map[string]interface{}{
			workflowRunJSON(42, "Build", "feature-x", "alice", "push", "success",
				"2026-06-15T10:00:00Z", "2026-06-15T10:05:00Z"),
		}))
	})
	c, _ := newTestClient(t, mux)

	runs, err := c.WorkflowRuns(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	r := runs[0]
	if r.CreatedAt.IsZero() {
		t.Error("CreatedAt must not be zero")
	}
	if r.UpdatedAt.IsZero() {
		t.Error("UpdatedAt must not be zero")
	}
}

func TestWorkflowRunsShouldNotReturnZeroIDInMetadata(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/actions/runs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, workflowRunsPageJSON([]map[string]interface{}{
			workflowRunJSON(99, "CI", "main", "alice", "push", "success",
				"2026-06-15T10:00:00Z", "2026-06-15T10:05:00Z"),
		}))
	})
	c, _ := newTestClient(t, mux)

	runs, _ := c.WorkflowRuns(context.Background(), "owner", "repo")
	if len(runs) > 0 && runs[0].ID == 0 {
		t.Error("run ID must not be zero when the API returns a non-zero ID")
	}
}

func TestWorkflowRunsShouldNotSendCreatedQueryFilter(t *testing.T) {
	var receivedCreated string
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/actions/runs", func(w http.ResponseWriter, r *http.Request) {
		receivedCreated = r.URL.Query().Get("created")
		writeJSON(w, http.StatusOK, workflowRunsPageJSON([]map[string]interface{}{}))
	})
	c, _ := newTestClient(t, mux)

	_, _ = c.WorkflowRuns(context.Background(), "owner", "repo")

	if receivedCreated != "" {
		t.Errorf("WorkflowRuns must never send a created filter; got %q", receivedCreated)
	}
}

func TestWorkflowRunsShouldReturnErrorOnServerError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/actions/runs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Internal Server Error"})
	})
	c, _ := newTestClient(t, mux)

	_, err := c.WorkflowRuns(context.Background(), "owner", "repo")
	if err == nil {
		t.Error("expected error on 500, got nil")
	}
}

func TestJobsForRunShouldReturnAllJobs(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/actions/runs/123/jobs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, jobsPageJSON([]map[string]interface{}{
			jobJSON(1, "build", "success"),
			jobJSON(2, "test", "failure"),
		}))
	})
	c, _ := newTestClient(t, mux)

	jobs, err := c.JobsForRun(context.Background(), "owner", "repo", 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestJobsForRunShouldNotReturnEmptySliceWhenJobsExist(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/actions/runs/123/jobs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, jobsPageJSON([]map[string]interface{}{
			jobJSON(1, "build", "success"),
		}))
	})
	c, _ := newTestClient(t, mux)

	jobs, err := c.JobsForRun(context.Background(), "owner", "repo", 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobs) == 0 {
		t.Error("expected non-empty slice when run has jobs")
	}
}

func TestJobsForRunShouldReturnEmptySliceWhenNoJobsExist(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/actions/runs/123/jobs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, jobsPageJSON([]map[string]interface{}{}))
	})
	c, _ := newTestClient(t, mux)

	jobs, err := c.JobsForRun(context.Background(), "owner", "repo", 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobs) != 0 {
		t.Errorf("expected empty slice, got %d jobs", len(jobs))
	}
}

func TestJobsForRunShouldNotReturnNilWhenNoJobsExist(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/actions/runs/123/jobs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, jobsPageJSON([]map[string]interface{}{}))
	})
	c, _ := newTestClient(t, mux)

	jobs, err := c.JobsForRun(context.Background(), "owner", "repo", 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jobs == nil {
		t.Error("JobsForRun must return a non-nil slice even when the run has no jobs")
	}
}

func TestJobsForRunShouldPopulateJobFields(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/actions/runs/123/jobs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, jobsPageJSON([]map[string]interface{}{
			jobJSON(42, "build-and-test", "success"),
		}))
	})
	c, _ := newTestClient(t, mux)

	jobs, err := c.JobsForRun(context.Background(), "owner", "repo", 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	j := jobs[0]
	if j.ID != 42 {
		t.Errorf("expected ID 42, got %d", j.ID)
	}
	if j.Name != "build-and-test" {
		t.Errorf("expected Name 'build-and-test', got %q", j.Name)
	}
	if j.Conclusion != "success" {
		t.Errorf("expected Conclusion 'success', got %q", j.Conclusion)
	}
}

func TestJobsForRunShouldNotReturnZeroIDInJobFields(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/actions/runs/123/jobs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, jobsPageJSON([]map[string]interface{}{
			jobJSON(99, "lint", "success"),
		}))
	})
	c, _ := newTestClient(t, mux)

	jobs, _ := c.JobsForRun(context.Background(), "owner", "repo", 123)
	if len(jobs) > 0 && jobs[0].ID == 0 {
		t.Error("job ID must not be zero when the API returns a non-zero ID")
	}
}

func TestJobsForRunShouldSupportAllConclusionValues(t *testing.T) {
	conclusions := []string{
		"success", "failure", "cancelled", "skipped",
		"timed_out", "action_required", "neutral", "stale",
	}
	for _, conclusion := range conclusions {
		t.Run(conclusion, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/repos/owner/repo/actions/runs/1/jobs", func(w http.ResponseWriter, r *http.Request) {
				writeJSON(w, http.StatusOK, jobsPageJSON([]map[string]interface{}{
					jobJSON(1, "job", conclusion),
				}))
			})
			c, _ := newTestClient(t, mux)

			jobs, err := c.JobsForRun(context.Background(), "owner", "repo", 1)
			if err != nil {
				t.Fatalf("unexpected error for conclusion %q: %v", conclusion, err)
			}
			if len(jobs) != 1 {
				t.Fatalf("expected 1 job, got %d", len(jobs))
			}
			if jobs[0].Conclusion != conclusion {
				t.Errorf("expected conclusion %q, got %q", conclusion, jobs[0].Conclusion)
			}
		})
	}
}

func TestJobsForRunShouldHandlePagination(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/actions/runs/123/jobs", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		if page == "2" {
			writeJSON(w, http.StatusOK, jobsPageJSON([]map[string]interface{}{
				jobJSON(3, "deploy", "success"),
				jobJSON(4, "smoke-test", "success"),
			}))
			return
		}
		serverURL := "http://" + r.Host
		w.Header().Set("Link", fmt.Sprintf(`<%s/repos/owner/repo/actions/runs/123/jobs?page=2>; rel="next"`, serverURL))
		writeJSON(w, http.StatusOK, jobsPageJSON([]map[string]interface{}{
			jobJSON(1, "build", "success"),
			jobJSON(2, "test", "failure"),
		}))
	})
	c, _ := newTestClient(t, mux)

	jobs, err := c.JobsForRun(context.Background(), "owner", "repo", 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobs) != 4 {
		t.Errorf("expected 4 jobs across 2 pages, got %d", len(jobs))
	}
}

func TestJobsForRunShouldReturnErrorOnServerError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/actions/runs/123/jobs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Internal Server Error"})
	})
	c, _ := newTestClient(t, mux)

	_, err := c.JobsForRun(context.Background(), "owner", "repo", 123)
	if err == nil {
		t.Error("expected error on 500, got nil")
	}
}

func TestFetchWorkflowsWithJobsShouldReturnRunsWithJobs(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/actions/runs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, workflowRunsPageJSON([]map[string]interface{}{
			workflowRunJSON(1, "CI", "main", "alice", "push", "success", "2026-06-15T10:00:00Z", "2026-06-15T10:05:00Z"),
		}))
	})
	mux.HandleFunc("/repos/owner/repo/actions/runs/1/jobs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, jobsPageJSON([]map[string]interface{}{
			jobJSON(10, "build", "success"),
			jobJSON(11, "test", "success"),
		}))
	})
	c, _ := newTestClient(t, mux)

	result, err := c.FetchWorkflowsWithJobs(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 run-with-jobs, got %d", len(result))
	}
	if len(result[0].Jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(result[0].Jobs))
	}
}

func TestFetchWorkflowsWithJobsShouldContinueWhenJobFetchFails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/actions/runs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, workflowRunsPageJSON([]map[string]interface{}{
			workflowRunJSON(1, "CI", "main", "alice", "push", "success", "2026-06-15T10:00:00Z", "2026-06-15T10:05:00Z"),
			workflowRunJSON(2, "CI", "main", "bob", "push", "failure", "2026-06-16T10:00:00Z", "2026-06-16T10:05:00Z"),
		}))
	})
	// Run 1 jobs: returns error
	mux.HandleFunc("/repos/owner/repo/actions/runs/1/jobs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Internal Server Error"})
	})
	// Run 2 jobs: returns successfully
	mux.HandleFunc("/repos/owner/repo/actions/runs/2/jobs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, jobsPageJSON([]map[string]interface{}{
			jobJSON(20, "build", "failure"),
		}))
	})
	c, _ := newTestClient(t, mux)

	result, err := c.FetchWorkflowsWithJobs(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("expected no top-level error, got %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 runs (both present despite job-fetch error), got %d", len(result))
	}
}

func TestFetchWorkflowsWithJobsShouldReturnEmptyJobsForFailedRunJobFetch(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/actions/runs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, workflowRunsPageJSON([]map[string]interface{}{
			workflowRunJSON(1, "CI", "main", "alice", "push", "success", "2026-06-15T10:00:00Z", "2026-06-15T10:05:00Z"),
		}))
	})
	mux.HandleFunc("/repos/owner/repo/actions/runs/1/jobs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Internal Server Error"})
	})
	c, _ := newTestClient(t, mux)

	result, _ := c.FetchWorkflowsWithJobs(context.Background(), "owner", "repo")
	if len(result) > 0 && result[0].Jobs == nil {
		t.Error("Jobs must be a non-nil empty slice (not nil) when job fetch fails")
	}
}
