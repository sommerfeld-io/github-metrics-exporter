package acceptance_test

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/token"
)

type tokenScenarioState struct {
	inputToken string
	result     error
}

func (s *tokenScenarioState) theExporterIsStartedWithoutGITHUBTOKENSet() error {
	s.inputToken = ""
	return nil
}

func (s *tokenScenarioState) theExporterIsStartedWithGITHUBTOKENSetToAnEmptyString() error {
	s.inputToken = ""
	return nil
}

func (s *tokenScenarioState) theExporterIsStartedWithANonEmptyGITHUBTOKEN() error {
	s.inputToken = "ghp_acceptance-test-token-placeholder"
	return nil
}

func (s *tokenScenarioState) theStartupSequenceValidatesTheToken() error {
	s.result = token.Validate(s.inputToken)
	return nil
}

func (s *tokenScenarioState) theValidationReturnsAClearErrorIndicatingTheMissingToken() error {
	if s.result == nil {
		return fmt.Errorf("expected token validation to return an error, got nil")
	}
	return nil
}

func (s *tokenScenarioState) theErrorMessageMentions(expected string) error {
	if s.result == nil {
		return fmt.Errorf("expected an error, but result was nil")
	}
	if !strings.Contains(s.result.Error(), expected) {
		return fmt.Errorf("expected error message to mention %q, got %q", expected, s.result.Error())
	}
	return nil
}

func (s *tokenScenarioState) theValidationSucceedsWithoutError() error {
	if s.result != nil {
		return fmt.Errorf("expected no error from token validation, got %v", s.result)
	}
	return nil
}

// InitializeTokenScenario registers token validation step definitions with GoDog.
func InitializeTokenScenario(ctx *godog.ScenarioContext) {
	s := &tokenScenarioState{}

	ctx.Step(`^the exporter is started without GITHUB_TOKEN set$`, s.theExporterIsStartedWithoutGITHUBTOKENSet)
	ctx.Step(`^the exporter is started with GITHUB_TOKEN set to an empty string$`, s.theExporterIsStartedWithGITHUBTOKENSetToAnEmptyString)
	ctx.Step(`^the exporter is started with a non-empty GITHUB_TOKEN$`, s.theExporterIsStartedWithANonEmptyGITHUBTOKEN)
	ctx.Step(`^the startup sequence validates the token$`, s.theStartupSequenceValidatesTheToken)
	ctx.Step(`^the validation returns a clear error indicating the missing token$`, s.theValidationReturnsAClearErrorIndicatingTheMissingToken)
	ctx.Step(`^the error message mentions "([^"]*)"$`, s.theErrorMessageMentions)
	ctx.Step(`^the validation succeeds without error$`, s.theValidationSucceedsWithoutError)
}
