package token_test

import (
	"strings"
	"testing"

	"github.com/sommerfeld-io/github-metrics-exporter/internal/token"
)

func TestValidateShouldReturnErrorWhenValueIsEmpty(t *testing.T) {
	err := token.Validate("")
	if err == nil {
		t.Error("expected error for empty token value, got nil")
	}
}

func TestValidateShouldNotReturnNilWhenValueIsEmpty(t *testing.T) {
	err := token.Validate("")
	if err == nil {
		t.Error("error must not be nil when token value is empty")
	}
}

func TestValidateShouldReturnNilWhenValueIsNonEmpty(t *testing.T) {
	err := token.Validate("ghp_somevalidtoken123")
	if err != nil {
		t.Errorf("expected no error for non-empty token, got %v", err)
	}
}

func TestValidateShouldNotReturnErrorWhenValueIsNonEmpty(t *testing.T) {
	err := token.Validate("ghp_somevalidtoken123")
	if err != nil {
		t.Errorf("error must be nil when token value is non-empty, got %v", err)
	}
}

func TestValidateShouldReturnErrorMentioningGITHUB_TOKEN(t *testing.T) {
	err := token.Validate("")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "GITHUB_TOKEN") {
		t.Errorf("expected error message to mention GITHUB_TOKEN, got %q", err.Error())
	}
}

func TestValidateShouldNotReturnErrorMessageContainingTokenValue(t *testing.T) {
	secretToken := "ghp_this_is_a_secret_value_12345"
	err := token.Validate(secretToken)
	if err != nil {
		if strings.Contains(err.Error(), secretToken) {
			t.Errorf("error message must not contain the token value, got %q", err.Error())
		}
	}
}
