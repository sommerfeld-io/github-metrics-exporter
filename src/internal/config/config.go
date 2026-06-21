// Package config loads and validates the exporter's YAML configuration file.
package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// GitHubTargets holds the GitHub organizations and users to discover repositories for.
type GitHubTargets struct {
	Organizations []string
	Users         []string
}

// Config holds the validated application configuration.
type Config struct {
	Port   int
	GitHub GitHubTargets
}

// rawConfig is the intermediate struct used for YAML unmarshaling.
// Port is kept as interface{} so the existing type-checked parsePort logic is preserved.
type rawConfig struct {
	Port   interface{}   `yaml:"port"`
	GitHub githubSection `yaml:"github"`
}

type githubSection struct {
	Organizations []string `yaml:"organizations"`
	Users         []string `yaml:"users"`
}

// Load reads the YAML file at path, validates all required fields and their
// types, and returns a populated Config or a descriptive error.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: cannot read %q: %w", path, err)
	}

	var raw rawConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("config: invalid YAML: %w", err)
	}

	if raw.Port == nil {
		return nil, errors.New(`config: required field "port" is missing`)
	}

	port, err := parsePort(raw.Port)
	if err != nil {
		return nil, err
	}

	return &Config{
		Port: port,
		GitHub: GitHubTargets{
			Organizations: raw.GitHub.Organizations,
			Users:         raw.GitHub.Users,
		},
	}, nil
}

// parsePort converts the raw YAML value for the port field into an int.
func parsePort(raw interface{}) (int, error) {
	switch v := raw.(type) {
	case nil:
		return 0, errors.New(`config: "port" must not be null or empty`)
	case string:
		return 0, fmt.Errorf(`config: "port" must be a raw integer, not a string (got %q)`, v)
	case float64:
		return 0, fmt.Errorf(`config: "port" must be an integer; floating-point numbers are not allowed (got %v)`, v)
	case bool:
		return 0, fmt.Errorf(`config: "port" must be an integer, not a boolean (got %v)`, v)
	case int:
		return v, nil
	case int64:
		return int(v), nil
	default:
		return 0, fmt.Errorf(`config: "port" has unexpected type %T`, v)
	}
}
