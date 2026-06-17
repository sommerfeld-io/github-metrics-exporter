package acceptance_test

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cucumber/godog"
)

// scenarioState holds the HTTP response and body for a single GoDog scenario.
// A fresh instance is created per scenario so state does not leak between runs.
type scenarioState struct {
	response *http.Response
	body     string
}

func (s *scenarioState) theExporterApplicationIsRunningAndHealthy() error {
	resp, err := http.Get(baseURL + "/healthz")
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected health check status 200, got %d", resp.StatusCode)
	}
	return nil
}

func (s *scenarioState) aUserRequestsThePath(path string) error {
	resp, err := http.Get(baseURL + path)
	if err != nil {
		return fmt.Errorf("GET %s: %w", path, err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("reading response body for %s: %w", path, err)
	}
	s.response = resp
	s.body = string(body)
	return nil
}

func (s *scenarioState) theHTTPStatusCodeShouldBe(code int) error {
	if s.response.StatusCode != code {
		return fmt.Errorf("expected HTTP status %d, got %d", code, s.response.StatusCode)
	}
	return nil
}

func (s *scenarioState) theResponseContentTypeShouldContain(contentType string) error {
	actual := s.response.Header.Get("Content-Type")
	if !strings.Contains(actual, contentType) {
		return fmt.Errorf("expected Content-Type to contain %q, got %q", contentType, actual)
	}
	return nil
}

func (s *scenarioState) thePageShouldContainTheHeadline(headline string) error {
	if !strings.Contains(s.body, headline) {
		return fmt.Errorf("expected response body to contain headline %q", headline)
	}
	return nil
}

func (s *scenarioState) thePageShouldContainALinkTo(url string) error {
	if !strings.Contains(s.body, url) {
		return fmt.Errorf("expected response body to contain link to %q", url)
	}
	return nil
}

func (s *scenarioState) thePageShouldDisplayTheStaticBuildCommitSHA() error {
	if !strings.Contains(s.body, "Build commit SHA:") {
		return fmt.Errorf("expected response body to display build commit SHA section")
	}
	return nil
}

func (s *scenarioState) thePageSourceHTMLStyleShouldDeclareADarkBackgroundTheme() error {
	if !strings.Contains(s.body, "background-color: #121212") {
		return fmt.Errorf("expected response body CSS to declare dark background-color (#121212)")
	}
	return nil
}

func (s *scenarioState) thePageShouldDisplayTheConfiguredPort() error {
	if !strings.Contains(s.body, "9400") {
		return fmt.Errorf("expected response body to display the configured port (9400)")
	}
	return nil
}

func (s *scenarioState) theResponseBodyMustContainTheDefaultMetricWithACommitSHALabel(metric string) error {
	if !strings.Contains(s.body, metric) {
		return fmt.Errorf("expected response body to contain metric %q", metric)
	}
	if !strings.Contains(s.body, `commit_sha="`) {
		return fmt.Errorf("expected metric %q to have a commit_sha label", metric)
	}
	return nil
}

// InitializeScenario registers all step definitions with GoDog.
// A new scenarioState is created per scenario to prevent state from leaking between scenarios.
func InitializeScenario(ctx *godog.ScenarioContext) {
	s := &scenarioState{}

	ctx.Step(`^the exporter application is running and healthy$`, s.theExporterApplicationIsRunningAndHealthy)
	ctx.Step(`^a user requests the path "([^"]*)"$`, s.aUserRequestsThePath)
	ctx.Step(`^the HTTP status code should be (\d+)$`, s.theHTTPStatusCodeShouldBe)
	ctx.Step(`^the response content-type should contain "([^"]*)"$`, s.theResponseContentTypeShouldContain)
	ctx.Step(`^the page should contain the headline "([^"]*)"$`, s.thePageShouldContainTheHeadline)
	ctx.Step(`^the page should contain a link to "([^"]*)"$`, s.thePageShouldContainALinkTo)
	ctx.Step(`^the page should display the static build commit SHA$`, s.thePageShouldDisplayTheStaticBuildCommitSHA)
	ctx.Step(`^the page source HTML style should declare a dark background theme$`, s.thePageSourceHTMLStyleShouldDeclareADarkBackgroundTheme)
	ctx.Step(`^the page should display the configured port$`, s.thePageShouldDisplayTheConfiguredPort)
	ctx.Step(`^the response body must contain the default metric "([^"]*)" with a commit_sha label$`, s.theResponseBodyMustContainTheDefaultMetricWithACommitSHALabel)
}
