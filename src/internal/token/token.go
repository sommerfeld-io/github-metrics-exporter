// Package token provides startup validation for the GitHub API token.
package token

import "errors"

// Validate returns an error if value is empty, treating both an absent and an
// explicitly-empty GITHUB_TOKEN as a configuration error that must be surfaced
// before the HTTP server starts.
func Validate(value string) error {
	if value == "" {
		return errors.New("GITHUB_TOKEN is required but was not set or is empty")
	}
	return nil
}
