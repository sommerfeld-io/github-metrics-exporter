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
