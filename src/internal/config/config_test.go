package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sommerfeld-io/github-metrics-exporter/internal/config"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("setup: %v", err)
	}
	return path
}

func TestLoadShouldSucceedWithValidConfig(t *testing.T) {
	path := writeConfig(t, "port: 9400\n")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Port != 9400 {
		t.Errorf("expected port 9400, got %d", cfg.Port)
	}
}

func TestLoadShouldNotReturnZeroPort(t *testing.T) {
	path := writeConfig(t, "port: 9400\n")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Port == 0 {
		t.Error("port must not be 0 (zero value means the field was never set)")
	}
}

func TestLoadShouldReturnErrorWhenFileDoesNotExist(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yml")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func TestLoadShouldReturnErrorWhenPortKeyIsMissing(t *testing.T) {
	path := writeConfig(t, "other_key: value\n")
	_, err := config.Load(path)
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
			_, err := config.Load(path)
			if err == nil {
				t.Errorf("expected error for input %q, got nil", tc.content)
			}
		})
	}
}

func TestLoadShouldReturnDescriptiveErrorForStringPort(t *testing.T) {
	path := writeConfig(t, "port: \"9400\"\n")
	_, err := config.Load(path)
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
	_, err := config.Load(path)
	if err == nil {
		t.Error("expected error for malformed YAML, got nil")
	}
}
