# AI Instructions

This file defines the rules and conventions that AI coding assistants should follow when working in this repository. It is tool-agnostic and serves as the single source of truth for all AI-generated code and commit messages.

Go-specific coding rules live in `src/CLAUDE.md` which is a symlink to `.github/instructions/go.instructions.md`. Run `task symlinks` from the repo root to (re)create the symlink if needed.

## Commit Messages: Conventional Commits

Always use Conventional Commits for every commit message.

**Format:** `<type>(optional scope): <description>`

| Type                                                                | Effect        | When to use                      |
|---------------------------------------------------------------------|---------------|----------------------------------|
| `fix`                                                               | PATCH release | Patches a bug                    |
| `feat`                                                              | MINOR release | Introduces a new feature         |
| `BREAKING CHANGE` footer                                            | MAJOR release | Introduces a breaking API change |
| `build`, `chore`, `ci`, `docs`, `style`, `refactor`, `perf`, `test` | No release    | All other changes                |

**Rules:**

- A scope may be added in parentheses for extra context: `feat(parser): add ability to parse arrays`. A scope may **NOT** contain a slash (`/`).
- Breaking changes must include `BREAKING CHANGE:` in the footer: `feat: drop support for Node 6`
- Commit message titles must also match the project pattern: `^(fix|feat|build|chore|ci|docs|style|refactor|perf|test)/[a-z0-9._-]+$`

Write commit messages using the Conventional Commits format, ensuring the header (`type(scope): summary`) is clear and descriptive, as it will be displayed on GitHub release pages and used for changelogs. Focus the header on user-visible, meaningful change descriptions and avoid vague wording. Always document breaking changes explicitly in the footer using `BREAKING CHANGE:` (do not use the `!` notation).

## Development Commands

All Go work runs from `src/`. The root `taskfile.yml` delegates to `src/taskfile.yml` via the `go:` namespace.

```bash
# From repo root
task go:build               # lint + vet + test + compile binary (full pipeline)
task go:test-with-coverage  # run unit tests with coverage report
task go:lint                # run golangci-lint
task go:run                 # build then run the binary locally (port 9400)
task symlinks               # (re)create CLAUDE.md and src/CLAUDE.md symlinks

# From src/
go test ./...                           # run all unit tests
go test -run TestXxx ./internal/...     # run a single test by name
go build -ldflags "-X github.com/sommerfeld-io/github-metrics-exporter/internal/metrics.CommitSHA=$(git rev-parse HEAD)" -o github-metrics-exporter
```

The binary listens on `:9400`. Endpoints: `/` (HTML landing page), `/metrics` (Prometheus), `/healthz` (plain-text health check).

## Architecture

```plain
src/
  main.go                        # wires metrics registration + HTTP server, starts on :9400
  internal/
    metrics/
      vars.go                    # MetricPrefix ("ghme_") and CommitSHA (injected at build time)
      exporter.go                # Register() - registers all meta-metrics with a prometheus.Registerer
      meta/
        meta.go                  # Init() - creates the ghme_exporter_info GaugeVec with commit_sha label
    server/
      server.go                  # New() - builds the http.ServeMux with /, /metrics, /healthz routes
tests/
  acceptance/features/           # Gherkin .feature files (godog; not yet wired to a runner)
```

## Test-Driven Development

Always follow TDD. The test comes before the implementation - no exceptions.

**Red-Green-Refactor:**

1. **Red** - write a failing test that describes the behavior you want. The test must fail because the implementation does not exist yet, not because of a syntax error.
2. **Green** - write the minimum implementation needed to make the test pass. Resist the urge to write more than the test demands.
3. **Refactor** - clean up both the implementation while keeping everything green.

**Make yourself the first consumer.** Before writing any function or type, write the call site first (in a test). This forces the API to be designed from the outside in, which keeps interfaces small and call sites easy to use.

**Test behavior, not implementation.** Tests must assert on observable outcomes: return values, HTTP responses, metric values, error messages. Never assert on internal state, private fields, or how many times a helper was called. A test that breaks when you rename a private variable is a bad test.

Cover the inversion. For every "should do X" test, add a "should not do Y" counterpart where it increases confidence. Example: if you assert gauge == 1.0, also assert gauge != 0.0 so the test proves the value was set explicitly and not left at the Go zero value. In Go, prefer expressing these positive/negative cases as table-driven tests when they share the same setup and differ only in inputs or expected outcomes. This keeps the test suite concise while ensuring both the expected behavior and its inversion are verified.

## Linux Requirement

This repository uses symlinks and is intended to be worked on in a Linux environment. The devcontainer makes working from a Windows host possible - the container itself is Linux, and the actual development environment runs inside it.
