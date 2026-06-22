// Package github provides GitHub API discovery for the exporter.
// It lists repositories for configured organizations and users,
// handling pagination and access errors transparently.
package github

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"time"

	gogithub "github.com/google/go-github/v67/github"
)

// WorkflowRun represents a single GitHub Actions workflow run.
type WorkflowRun struct {
	ID         int64
	Name       string
	HeadBranch string
	Actor      string
	Event      string
	Conclusion string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Job represents a single job within a workflow run.
type Job struct {
	ID         int64
	Name       string
	Conclusion string
}

// RunWithJobs pairs a workflow run with its retrieved jobs.
type RunWithJobs struct {
	Run  WorkflowRun
	Jobs []Job
}

// Repository represents a GitHub repository discovered for a target org or user.
type Repository struct {
	Owner      string
	Name       string
	Accessible bool
}

// Client wraps the GitHub REST API client and exposes repository discovery.
type Client struct {
	gh *gogithub.Client
}

// NewWithToken constructs a Client authenticated with the given personal access token.
func NewWithToken(token string) *Client {
	return &Client{gh: gogithub.NewClient(nil).WithAuthToken(token)}
}

// SetBaseURL redirects all API calls to u. The URL must end with a trailing slash.
// This is intended for tests that point the client at a mock server.
func (c *Client) SetBaseURL(u *url.URL) {
	c.gh.BaseURL = u
}

// Discover fetches all repositories for the given organizations and users.
// Repositories are marked Accessible=true when successfully listed.
// A 403 or 404 response for an organization or user is logged as a warning
// and that target is skipped; discovery continues for the remaining targets.
// Any other API error is returned immediately.
// The returned slice is sorted by Owner then Name.
func (c *Client) Discover(ctx context.Context, orgs, users []string) ([]Repository, error) {
	repos := make([]Repository, 0)

	for _, org := range orgs {
		found, err := c.listOrgRepos(ctx, org)
		if err != nil {
			return nil, err
		}
		repos = append(repos, found...)
	}

	for _, user := range users {
		found, err := c.listUserRepos(ctx, user)
		if err != nil {
			return nil, err
		}
		repos = append(repos, found...)
	}

	sort.Slice(repos, func(i, j int) bool {
		if repos[i].Owner != repos[j].Owner {
			return repos[i].Owner < repos[j].Owner
		}
		return repos[i].Name < repos[j].Name
	})

	return repos, nil
}

// listOrgRepos returns all repositories for a single organization, handling pagination.
// 403/404 responses are logged and result in an empty list for that org; other errors are returned.
func (c *Client) listOrgRepos(ctx context.Context, org string) ([]Repository, error) {
	slog.Info("github: listing repositories", "org", org)
	opts := &gogithub.RepositoryListByOrgOptions{
		ListOptions: gogithub.ListOptions{PerPage: 100},
	}
	var repos []Repository
	for {
		page, resp, err := c.gh.Repositories.ListByOrg(ctx, org, opts)
		if err != nil {
			if isAccessError(err) {
				slog.Warn("github: cannot list org repos", "org", org, "error", err)
				return nil, nil
			}
			return nil, fmt.Errorf("github: list repos for org %q: %w", org, err)
		}
		for _, r := range page {
			repos = append(repos, Repository{
				Owner:      org,
				Name:       r.GetName(),
				Accessible: true,
			})
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return repos, nil
}

// listUserRepos returns all repositories for a single user, handling pagination.
// 403/404 responses are logged and result in an empty list for that user; other errors are returned.
func (c *Client) listUserRepos(ctx context.Context, user string) ([]Repository, error) {
	slog.Info("github: listing repositories", "user", user)
	opts := &gogithub.RepositoryListByUserOptions{
		ListOptions: gogithub.ListOptions{PerPage: 100},
	}
	var repos []Repository
	for {
		page, resp, err := c.gh.Repositories.ListByUser(ctx, user, opts)
		if err != nil {
			if isAccessError(err) {
				slog.Warn("github: cannot list user repos", "user", user, "error", err)
				return nil, nil
			}
			return nil, fmt.Errorf("github: list repos for user %q: %w", user, err)
		}
		for _, r := range page {
			repos = append(repos, Repository{
				Owner:      user,
				Name:       r.GetName(),
				Accessible: true,
			})
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return repos, nil
}

// WorkflowRuns fetches the most recent workflow runs for the given repository.
// It fetches a single page of results with no date filter, exposing current state
// as required by the Prometheus exporter model. Prometheus manages the time series.
func (c *Client) WorkflowRuns(ctx context.Context, owner, repo string) ([]WorkflowRun, error) {
	slog.Info("github: fetching workflow runs", "owner", owner, "repo", repo)
	opts := &gogithub.ListWorkflowRunsOptions{
		ListOptions: gogithub.ListOptions{PerPage: 100},
	}
	page, _, err := c.gh.Actions.ListRepositoryWorkflowRuns(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("github: list workflow runs for %s/%s: %w", owner, repo, err)
	}
	runs := make([]WorkflowRun, 0, len(page.WorkflowRuns))
	for _, r := range page.WorkflowRuns {
		runs = append(runs, WorkflowRun{
			ID:         r.GetID(),
			Name:       r.GetName(),
			HeadBranch: r.GetHeadBranch(),
			Actor:      r.GetActor().GetLogin(),
			Event:      r.GetEvent(),
			Conclusion: r.GetConclusion(),
			CreatedAt:  r.GetCreatedAt().Time,
			UpdatedAt:  r.GetUpdatedAt().Time,
		})
	}
	return runs, nil
}

// JobsForRun fetches all jobs for a single workflow run.
// Pagination is handled automatically.
func (c *Client) JobsForRun(ctx context.Context, owner, repo string, runID int64) ([]Job, error) {
	slog.Info("github: fetching jobs for run", "owner", owner, "repo", repo, "run_id", runID)
	opts := &gogithub.ListWorkflowJobsOptions{
		ListOptions: gogithub.ListOptions{PerPage: 100},
	}

	var jobs []Job
	for {
		page, resp, err := c.gh.Actions.ListWorkflowJobs(ctx, owner, repo, runID, opts)
		if err != nil {
			return nil, fmt.Errorf("github: list jobs for run %d in %s/%s: %w", runID, owner, repo, err)
		}
		for _, j := range page.Jobs {
			jobs = append(jobs, Job{
				ID:         j.GetID(),
				Name:       j.GetName(),
				Conclusion: j.GetConclusion(),
			})
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	if jobs == nil {
		jobs = []Job{}
	}
	return jobs, nil
}

// FetchWorkflowsWithJobs fetches the most recent workflow runs and their jobs for the given repository.
// If fetching jobs for an individual run fails, the error is logged and that run is
// included with an empty job list; fetching continues for the remaining runs.
func (c *Client) FetchWorkflowsWithJobs(ctx context.Context, owner, repo string) ([]RunWithJobs, error) {
	runs, err := c.WorkflowRuns(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	result := make([]RunWithJobs, 0, len(runs))
	for _, run := range runs {
		jobs, err := c.JobsForRun(ctx, owner, repo, run.ID)
		if err != nil {
			slog.Warn("github: job fetch failed; run included with empty job list", "owner", owner, "repo", repo, "run_id", run.ID, "error", err)
			result = append(result, RunWithJobs{Run: run, Jobs: []Job{}})
			continue
		}
		result = append(result, RunWithJobs{Run: run, Jobs: jobs})
	}
	return result, nil
}

// isAccessError reports whether the error is a 403 Forbidden or 404 Not Found response.
func isAccessError(err error) bool {
	var ghErr *gogithub.ErrorResponse
	if errors.As(err, &ghErr) {
		code := ghErr.Response.StatusCode
		return code == http.StatusForbidden || code == http.StatusNotFound
	}
	return false
}
