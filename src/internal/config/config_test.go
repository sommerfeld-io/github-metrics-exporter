package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("setup: %v", err)
	}
	return path
}

func TestLoadShouldSucceedWithGitHubOrganizations(t *testing.T) {
	path := writeConfig(t, "port: 9400\ngithub:\n  organizations:\n    - org1\n    - org2\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cfg.GitHub.Organizations) != 2 {
		t.Errorf("expected 2 organizations, got %d", len(cfg.GitHub.Organizations))
	}
	if cfg.GitHub.Organizations[0] != "org1" || cfg.GitHub.Organizations[1] != "org2" {
		t.Errorf("unexpected organizations: %v", cfg.GitHub.Organizations)
	}
}

func TestLoadShouldNotReturnNilOrganizationsWhenPresent(t *testing.T) {
	path := writeConfig(t, "port: 9400\ngithub:\n  organizations:\n    - org1\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.GitHub.Organizations == nil {
		t.Error("organizations must not be nil when key is present in config")
	}
}

func TestLoadShouldSucceedWithGitHubUsers(t *testing.T) {
	path := writeConfig(t, "port: 9400\ngithub:\n  users:\n    - user1\n    - user2\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cfg.GitHub.Users) != 2 {
		t.Errorf("expected 2 users, got %d", len(cfg.GitHub.Users))
	}
	if cfg.GitHub.Users[0] != "user1" || cfg.GitHub.Users[1] != "user2" {
		t.Errorf("unexpected users: %v", cfg.GitHub.Users)
	}
}

func TestLoadShouldSucceedWithEmptyGitHubSection(t *testing.T) {
	path := writeConfig(t, "port: 9400\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error for config without github section, got %v", err)
	}
	if len(cfg.GitHub.Organizations) != 0 {
		t.Errorf("expected empty organizations, got %v", cfg.GitHub.Organizations)
	}
	if len(cfg.GitHub.Users) != 0 {
		t.Errorf("expected empty users, got %v", cfg.GitHub.Users)
	}
}

func TestLoadShouldSucceedWithBothOrgsAndUsers(t *testing.T) {
	path := writeConfig(t, "port: 9400\ngithub:\n  organizations:\n    - org1\n  users:\n    - user1\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cfg.GitHub.Organizations) != 1 || cfg.GitHub.Organizations[0] != "org1" {
		t.Errorf("unexpected organizations: %v", cfg.GitHub.Organizations)
	}
	if len(cfg.GitHub.Users) != 1 || cfg.GitHub.Users[0] != "user1" {
		t.Errorf("unexpected users: %v", cfg.GitHub.Users)
	}
}

func TestLoadShouldSucceedWithValidConfig(t *testing.T) {
	path := writeConfig(t, "port: 9400\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Port != 9400 {
		t.Errorf("expected port 9400, got %d", cfg.Port)
	}
}

func TestLoadShouldNotReturnZeroPort(t *testing.T) {
	path := writeConfig(t, "port: 9400\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Port == 0 {
		t.Error("port must not be 0 (zero value means the field was never set)")
	}
}

func TestLoadShouldReturnErrorWhenFileDoesNotExist(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yml")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func TestLoadShouldReturnErrorWhenPortKeyIsMissing(t *testing.T) {
	path := writeConfig(t, "other_key: value\n")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error when port key is missing, got nil")
	}
	if !strings.Contains(err.Error(), "port") {
		t.Errorf("expected error to mention \"port\", got %q", err.Error())
	}
}

func TestLoadShouldReturnErrorForInvalidPortTypes(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{"string integer", "port: \"9400\"\n"},
		{"string text", "port: \"localhost\"\n"},
		{"float", "port: 9400.5\n"},
		{"boolean", "port: true\n"},
		{"null", "port: ~\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := writeConfig(t, tc.content)
			_, err := Load(path)
			if err == nil {
				t.Errorf("expected error for input %q, got nil", tc.content)
			}
		})
	}
}

func TestLoadShouldReturnDescriptiveErrorForStringPort(t *testing.T) {
	path := writeConfig(t, "port: \"9400\"\n")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "string") && !strings.Contains(msg, "!!str") {
		t.Errorf("expected error to mention string type, got %q", msg)
	}
}

func TestLoadShouldReturnErrorWhenYAMLIsMalformed(t *testing.T) {
	path := writeConfig(t, "port: [\n")
	_, err := Load(path)
	if err == nil {
		t.Error("expected error for malformed YAML, got nil")
	}
}

func TestParsePortShouldReturnPortForInt(t *testing.T) {
	port, err := parsePort(9400)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if port != 9400 {
		t.Errorf("expected 9400, got %d", port)
	}
}

func TestParsePortShouldReturnPortForInt64(t *testing.T) {
	port, err := parsePort(int64(9400))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if port != 9400 {
		t.Errorf("expected 9400, got %d", port)
	}
}

func TestParsePortShouldNotReturnZeroForValidInput(t *testing.T) {
	port, err := parsePort(9400)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if port == 0 {
		t.Error("port must not be 0 (zero value means the field was never set)")
	}
}

func TestParsePortShouldReturnErrorForInvalidTypes(t *testing.T) {
	cases := []struct {
		name  string
		input interface{}
	}{
		{"nil", nil},
		{"string", "9400"},
		{"float64", 9400.5},
		{"bool", true},
		{"unexpected type", []int{9400}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parsePort(tc.input)
			if err == nil {
				t.Errorf("expected error for input %v (%T), got nil", tc.input, tc.input)
			}
		})
	}
}

func TestParsePortShouldReturnDescriptiveErrorForStringPort(t *testing.T) {
	_, err := parsePort("9400")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "string") {
		t.Errorf("expected error to mention string type, got %q", err.Error())
	}
}
