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
	"sort"

	gogithub "github.com/google/go-github/v67/github"
)

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

// isAccessError reports whether the error is a 403 Forbidden or 404 Not Found response.
func isAccessError(err error) bool {
	var ghErr *gogithub.ErrorResponse
	if errors.As(err, &ghErr) {
		code := ghErr.Response.StatusCode
		return code == http.StatusForbidden || code == http.StatusNotFound
	}
	return false
}
