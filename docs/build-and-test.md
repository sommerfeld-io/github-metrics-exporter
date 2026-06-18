# Build and Test

This document describes how the GitHub Metrics Exporter is built, tested, and released - both locally during development and in the automated CI/CD pipeline.

## Local Development

All Go work is driven through [Task](https://taskfile.dev). The root `taskfile.yml` delegates all Go tasks to `src/taskfile.yml` under the `go:` namespace. Running `task go:build` from the repo root is the standard development feedback loop: it runs every quality gate in sequence and only produces the binary if every step passes.

```ditaa
+-----------------------------------------------------------------------------+
| task go:build                                                               |
|                                                                             |
|  +----------+   +-------------+   +----------+   +------------------------+ |
|  | cleanup  +-->| complexity  +-->| licenses +-->| lint (golangci-lint)   | |
|  |          |   | (gocyclo    |   | check    |   |                        | |
|  +----------+   | max: 10)    |   |          |   +----------+-------------+ |
|                 +-------------+   +----------+              |               |
|                                                             |               |
|  +----------+   +----------+   +----------+                 |               |
|  | go mod   |   |  go fmt  |   |  go vet  |<----------------+               |
|  | tidy     +-->|          +-->|          |                                 |
|  +----------+   +----------+   +----+-----+                                 |
|                                     |                                       |
|                         +-----------+                                       |
|                         v                                                   |
|  +--------------+   +-------------------+   +--------------------------+    |
|  | unit tests   +-->| acceptance tests  +-->| go build                 |    |
|  | ./internal/  |   | ./acceptance-     |   | -ldflags CommitSHA=...   |    |
|  |              |   | tests/            |   |                          |    |
|  +--------------+   +-------------------+   +--------------------------+    |
+-----------------------------------------------------------------------------+
```

### Unit tests

Unit tests live next to the code they verify in `src/internal/` and follow the standard Go `_test.go` convention. They are fast, have no I/O or network dependencies, and produce a coverage report at `src/coverage.out`. The coverage scope is limited to `./internal/...`. Run them in isolation with `task go:test`.

### Acceptance tests

Acceptance tests are GoDog (Cucumber/Gherkin) tests that live in `src/acceptance-tests/`. They are part of the main Go module but excluded from the unit-test coverage run. When `TestMain` starts, it registers the Prometheus metrics and calls `httptest.NewServer(server.New())`, which creates a fully functional HTTP server bound to a random ephemeral port. Every Gherkin step then makes real `http.Get` calls against that address. No binary, no Docker, no fixed port is involved. Run them in isolation with `task go:test:acceptance`.

The acceptance tests are instrumented with `-coverpkg=./internal/...` and write a second coverage report to `src/acceptance-coverage.out`. This is a separate signal from `coverage.out`: it shows which lines in `internal/` are reachable end-to-end through the HTTP layer, not whether individual branches are exercised by unit tests. The two reports are intentionally kept apart - merging them would inflate unit-test coverage numbers and obscure gaps in either test suite.

```ditaa
+---------------------------+          +---------------------------------+
|       TestMain            |          |  Gherkin scenario               |
|                           |          |                                 |
|  metrics.Register(        |          |  Given the exporter is running  |
|    prometheus.Default...) |          |  When a user requests "/"       |
|                           |  HTTP    |  Then status should be 200      |
|  httptest.NewServer(      +--------->|  And Content-Type is text/html  |
|    server.New())          |  calls   |                                 |
|                           |          +---------------------------------+
|  baseURL = srv.URL        |
+---------------------------+
```

Both test kinds are mandatory gates inside `task go:build`. The binary is only compiled after both pass.

## Building the Docker Image

The `Dockerfile` uses a two-stage build. The `build` stage uses the full Go toolchain image, copies the source and the `.git` directory (needed for the linker-injected commit SHA), then runs `task build`. Because `task build` calls both the unit tests and the acceptance tests internally, all quality gates run inside Docker before the binary is produced. The `run` stage starts from a minimal Alpine image and contains only the compiled binary and a dedicated non-root user (`ghme`, UID 1000).

```ditaa
  +---------------------------------------+
  | FROM golang:1.26-alpine AS build      |
  |                                       |
  |  COPY .git + src/                     |
  |  RUN task build                       |
  |    :  cleanup                         |
  |    :  complexity + licenses + lint    |
  |    :  unit tests                      |
  |    :  acceptance tests                |
  |    :  go build -ldflags CommitSHA=..  |
  |                                       |
  +----------------+----------------------+
                   | binary only
                   v
  +---------------------------------------+
  | FROM alpine AS run                    |
  |                                       |
  |  adduser ghme (UID 1000)              |
  |  COPY binary from build stage         |
  |  CMD ./github-metrics-exporter        |
  |  (port 9400)                          |
  +---------------------------------------+
```

Locally, `task docker:build` first runs all project-wide linters (YAML, Markdown, Gherkin, file names, and others via Docker Compose services), then lints the Dockerfile itself, and finally calls `docker compose build`. This is the full local equivalent of the CI pipeline.

## CI/CD Pipeline

### Commit pipeline

The commit pipeline answers one question: can this code be integrated into `main` without breaking the app? Every change that touches the application is automatically built and tested. If any quality gate fails, the change does not advance. Automated tests are the primary signal here - they exist in this pipeline specifically to catch regressions in code that has not yet been released.

When all gates pass on `main`, the pipeline also publishes a release candidate as the `:edge` image on Docker Hub. This gives integration consumers a stable, always-up-to-date image that reflects the tip of `main`, without committing to a numbered release.

### Release pipeline

The release pipeline promotes a release candidate to a real release. Its job is not to rebuild or re-test from scratch - the `:edge` image that enters this pipeline has already passed every quality gate in the commit pipeline. Instead, the release pipeline handles everything that needs to happen at the moment a version becomes official: updating Docker Hub metadata, publishing release notes, and running a final set of tests against the released image to confirm the published artifact behaves as expected.

After a release, `:latest` on Docker Hub always points to the most recent stable version, and `:edge` continues to track the tip of `main`.
