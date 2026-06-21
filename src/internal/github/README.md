# Package: `github`

Provides GitHub API-based repository discovery for the exporter.

The `Client` type wraps the GitHub REST API and exposes a single `Discover` method that returns all repositories accessible for a given set of organizations and users. Pagination is handled transparently. A 403 or 404 response for a target is logged as a warning and that target is skipped; any other API error is returned to the caller.

The discovered repository slice is sorted by owner then name to ensure deterministic output.
