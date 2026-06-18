// Package config loads and validates the exporter's YAML configuration file.
package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the validated application configuration.
type Config struct {
	Port int
}

// Load reads the YAML file at path, validates all required fields and their
// types, and returns a populated Config or a descriptive error.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: cannot read %q: %w", path, err)
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("config: invalid YAML: %w", err)
	}

	portRaw, exists := raw["port"]
	if !exists {
		return nil, errors.New(`config: required field "port" is missing`)
	}

	port, err := parsePort(portRaw)
	if err != nil {
		return nil, err
	}
	return &Config{Port: port}, nil
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
