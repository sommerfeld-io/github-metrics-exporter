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
	Path       string
	HeadBranch string
	Actor      string
	Event      string
	Conclusion string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Workflow represents a GitHub Actions workflow definition file.
type Workflow struct {
	ID   int64
	Name string
	Path string
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

// ListWorkflows fetches all workflow definitions for the given repository.
// 403/404 responses are logged and result in an empty list; other errors are returned.
func (c *Client) ListWorkflows(ctx context.Context, owner, repo string) ([]Workflow, error) {
	slog.Info("github: listing workflows", "owner", owner, "repo", repo)
	opts := &gogithub.ListOptions{PerPage: 100}
	var workflows []Workflow
	for {
		page, resp, err := c.gh.Actions.ListWorkflows(ctx, owner, repo, opts)
		if err != nil {
			if isAccessError(err) {
				slog.Warn("github: cannot list workflows", "owner", owner, "repo", repo, "error", err)
				return nil, nil
			}
			return nil, fmt.Errorf("github: list workflows for %s/%s: %w", owner, repo, err)
		}
		for _, w := range page.Workflows {
			workflows = append(workflows, Workflow{
				ID:   w.GetID(),
				Name: w.GetName(),
				Path: w.GetPath(),
			})
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	if workflows == nil {
		workflows = []Workflow{}
	}
	return workflows, nil
}

// latestRunByWorkflowID fetches the single most recent run for a specific workflow.
// Returns nil, nil when the workflow has no runs yet.
func (c *Client) latestRunByWorkflowID(ctx context.Context, owner, repo string, workflowID int64) (*WorkflowRun, error) {
	slog.Info("github: fetching latest run for workflow", "owner", owner, "repo", repo, "workflow_id", workflowID)
	opts := &gogithub.ListWorkflowRunsOptions{
		ListOptions: gogithub.ListOptions{PerPage: 1},
	}
	page, _, err := c.gh.Actions.ListWorkflowRunsByID(ctx, owner, repo, workflowID, opts)
	if err != nil {
		return nil, fmt.Errorf("github: list runs for workflow %d in %s/%s: %w", workflowID, owner, repo, err)
	}
	if len(page.WorkflowRuns) == 0 {
		return nil, nil
	}
	r := page.WorkflowRuns[0]
	return &WorkflowRun{
		ID:         r.GetID(),
		Name:       r.GetName(),
		Path:       r.GetPath(),
		HeadBranch: r.GetHeadBranch(),
		Actor:      r.GetActor().GetLogin(),
		Event:      r.GetEvent(),
		Conclusion: r.GetConclusion(),
		CreatedAt:  r.GetCreatedAt().Time,
		UpdatedAt:  r.GetUpdatedAt().Time,
	}, nil
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

// FetchWorkflowsWithJobs returns the latest run and its jobs for every workflow file in the
// repository. Each workflow file is queried independently so that all pipelines appear in the
// result regardless of which one ran most recently.
// Per-workflow run fetch failures are logged and that workflow is skipped.
// Per-run job fetch failures are logged and the run is included with an empty job list.
func (c *Client) FetchWorkflowsWithJobs(ctx context.Context, owner, repo string) ([]RunWithJobs, error) {
	workflows, err := c.ListWorkflows(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	result := make([]RunWithJobs, 0, len(workflows))
	for _, wf := range workflows {
		run, err := c.latestRunByWorkflowID(ctx, owner, repo, wf.ID)
		if err != nil {
			slog.Warn("github: run fetch failed for workflow; skipping", "owner", owner, "repo", repo, "workflow_id", wf.ID, "error", err)
			continue
		}
		if run == nil {
			continue
		}
		jobs, err := c.JobsForRun(ctx, owner, repo, run.ID)
		if err != nil {
			slog.Warn("github: job fetch failed; run included with empty job list", "owner", owner, "repo", repo, "run_id", run.ID, "error", err)
			result = append(result, RunWithJobs{Run: *run, Jobs: []Job{}})
			continue
		}
		result = append(result, RunWithJobs{Run: *run, Jobs: jobs})
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
