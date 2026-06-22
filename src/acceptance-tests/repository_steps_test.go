package acceptance_test

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cucumber/godog"
)

// repositoryState holds per-scenario state for repository-related scenarios.
type repositoryState struct {
	serverURL string
	body      string
	response  *http.Response
}

func (s *repositoryState) theExporterHasRepositoriesConfigured() error {
	s.serverURL = baseURL
	return nil
}

func (s *repositoryState) theExporterHasNoTargetsConfigured() error {
	s.serverURL = noTargetsURL
	return nil
}

func (s *repositoryState) aUserRequestsTheMetricsEndpoint() error {
	resp, err := http.Get(s.serverURL + "/metrics")
	if err != nil {
		return fmt.Errorf("GET /metrics: %w", err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("reading /metrics body: %w", err)
	}
	s.response = resp
	s.body = string(body)
	return nil
}

func (s *repositoryState) aUserNavigatesTo(path string) error {
	resp, err := http.Get(s.serverURL + path)
	if err != nil {
		return fmt.Errorf("GET %s: %w", path, err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}
	s.response = resp
	s.body = string(body)
	return nil
}

func (s *repositoryState) theMetricsBodyContainsRepositoryAccessibleWithValue(value int, owner, repo string) error {
	// Look for: ghme_repository_accessible{owner="...",repo="..."} <value>
	// Prometheus text format may have labels in any order, so check for both fields.
	ownerLabel := fmt.Sprintf(`owner="%s"`, owner)
	repoLabel := fmt.Sprintf(`repo="%s"`, repo)
	valueStr := fmt.Sprintf(" %d", value)

	for _, line := range strings.Split(s.body, "\n") {
		if strings.Contains(line, "repository_accessible") &&
			strings.Contains(line, ownerLabel) &&
			strings.Contains(line, repoLabel) &&
			strings.HasSuffix(strings.TrimSpace(line), valueStr) {
			return nil
		}
	}
	return fmt.Errorf("metric ghme_repository_accessible{owner=%q,repo=%q} = %d not found in /metrics output", owner, repo, value)
}

func (s *repositoryState) thePageShouldContainOwnerHeading(owner string) error {
	if !strings.Contains(s.body, owner) {
		return fmt.Errorf("expected owner heading %q on index page", owner)
	}
	return nil
}

func (s *repositoryState) thePageShouldListRepository(repo string) error {
	if !strings.Contains(s.body, repo) {
		return fmt.Errorf("expected repository %q to be listed on index page", repo)
	}
	return nil
}

func (s *repositoryState) thePageShouldShowBadgeFor(badge, repo string) error {
	badgeClass := fmt.Sprintf(`class="badge %s"`, badge)
	for _, line := range strings.Split(s.body, "\n") {
		if strings.Contains(line, repo) && strings.Contains(line, badgeClass) {
			return nil
		}
	}
	return fmt.Errorf("expected %q badge on the line containing repository %q", badge, repo)
}

func (s *repositoryState) thePageShouldDisplayNoTargetsWarning() error {
	if !strings.Contains(s.body, "No GitHub targets") {
		return fmt.Errorf("expected no-targets warning on index page")
	}
	return nil
}

func (s *repositoryState) theMetricsBodyContainsWorkflowRunConclusion(conclusion, owner, repo string) error {
	ownerLabel := fmt.Sprintf(`owner="%s"`, owner)
	repoLabel := fmt.Sprintf(`repo="%s"`, repo)
	conclusionLabel := fmt.Sprintf(`conclusion="%s"`, conclusion)
	for _, line := range strings.Split(s.body, "\n") {
		if strings.Contains(line, "workflow_run_conclusion") &&
			strings.Contains(line, ownerLabel) &&
			strings.Contains(line, repoLabel) &&
			strings.Contains(line, conclusionLabel) &&
			strings.HasSuffix(strings.TrimSpace(line), " 1") {
			return nil
		}
	}
	return fmt.Errorf("ghme_workflow_run_conclusion{owner=%q,repo=%q,conclusion=%q} not found in /metrics", owner, repo, conclusion)
}

func (s *repositoryState) theMetricsBodyContainsWorkflowJobConclusion(conclusion, owner, repo string) error {
	ownerLabel := fmt.Sprintf(`owner="%s"`, owner)
	repoLabel := fmt.Sprintf(`repo="%s"`, repo)
	conclusionLabel := fmt.Sprintf(`conclusion="%s"`, conclusion)
	for _, line := range strings.Split(s.body, "\n") {
		if strings.Contains(line, "workflow_job_conclusion") &&
			strings.Contains(line, ownerLabel) &&
			strings.Contains(line, repoLabel) &&
			strings.Contains(line, conclusionLabel) &&
			strings.HasSuffix(strings.TrimSpace(line), " 1") {
			return nil
		}
	}
	return fmt.Errorf("ghme_workflow_job_conclusion{owner=%q,repo=%q,conclusion=%q} not found in /metrics", owner, repo, conclusion)
}

// InitializeRepositoryScenario registers all repository-related step definitions with GoDog.
func InitializeRepositoryScenario(ctx *godog.ScenarioContext) {
	s := &repositoryState{}

	ctx.Step(`^the exporter has repositories configured$`, s.theExporterHasRepositoriesConfigured)
	ctx.Step(`^the exporter has no targets configured$`, s.theExporterHasNoTargetsConfigured)
	ctx.Step(`^a user requests the metrics endpoint$`, s.aUserRequestsTheMetricsEndpoint)
	ctx.Step(`^a user navigates to "([^"]*)"$`, s.aUserNavigatesTo)
	ctx.Step(`^the metrics body contains a repository accessible metric with value (\d+) for "([^"]*)" and "([^"]*)"$`, s.theMetricsBodyContainsRepositoryAccessibleWithValue)
	ctx.Step(`^the metrics body contains workflow run conclusion "([^"]*)" for "([^"]*)" and "([^"]*)"$`, s.theMetricsBodyContainsWorkflowRunConclusion)
	ctx.Step(`^the metrics body contains workflow job conclusion "([^"]*)" for "([^"]*)" and "([^"]*)"$`, s.theMetricsBodyContainsWorkflowJobConclusion)
	ctx.Step(`^the page should contain the owner heading "([^"]*)"$`, s.thePageShouldContainOwnerHeading)
	ctx.Step(`^the page should list repository "([^"]*)"$`, s.thePageShouldListRepository)
	ctx.Step(`^the page should show an? "([^"]*)" badge for "([^"]*)"$`, s.thePageShouldShowBadgeFor)
	ctx.Step(`^the page should display the no-targets warning$`, s.thePageShouldDisplayNoTargetsWarning)
}
