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
|  +----------+   +----------+   +----------+                |                |
|  | go mod   |   |  go fmt  |   |  go vet  |<---------------+                |
|  | tidy     +-->|          +-->|          |                                 |
|  +----------+   +----------+   +----+-----+                                 |
|                                     |                                       |
|                      +--------------+                                       |
|                      v                                                      |
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
+---------------------------+          +----------------------------------+
|       TestMain            |          |  Gherkin scenario                |
|                           |          |                                  |
|  metrics.Register(        |          |  Given the exporter is running   |
|    prometheus.Default...) |          |  When a user requests "/"        |
|                           |    HTTP  |  Then status should be 200       |
|  httptest.NewServer(  +------------>|  And Content-Type is text/html   |
|    server.New())          |  calls   |                                  |
|                           |          +----------------------------------+
|  baseURL = srv.URL        |
+---------------------------+
```

Both test kinds are mandatory gates inside `task go:build`. The binary is only compiled after
both pass.

---

## Building the Docker Image

The `Dockerfile` uses a two-stage build. The `build` stage uses the full Go toolchain image,
copies the source and the `.git` directory (needed for the linker-injected commit SHA), then
runs `task build`. Because `task build` calls both the unit tests and the acceptance tests
internally, all quality gates run inside Docker before the binary is produced. The `run` stage
starts from a minimal Alpine image and contains only the compiled binary and a dedicated
non-root user (`ghme`, UID 1000).

```ditaa
  +--------------------------------------+
  | FROM golang:1.26-alpine AS build     |
  |                                      |
  |  COPY .git + src/                    |
  |  RUN task build                      |
  |    :  cleanup                        |
  |    :  complexity + licenses + lint   |
  |    :  unit tests                     |
  |    :  acceptance tests               |
  |    :  go build -ldflags CommitSHA=.. |
  |                                      |
  +----------------+---------------------+
                   | binary only
                   v
  +--------------------------------------+
  | FROM alpine AS run                   |
  |                                      |
  |  adduser ghme (UID 1000)             |
  |  COPY binary from build stage        |
  |  CMD ./github-metrics-exporter       |
  |  (port 9400)                         |
  +--------------------------------------+
```

Locally, `task docker:build` first runs all project-wide linters (YAML, Markdown, Gherkin, file
names, and others via Docker Compose services), then lints the Dockerfile itself, and finally
calls `docker compose build`. This is the full local equivalent of the CI pipeline.

---

## CI/CD Pipeline

### Commit pipeline

The commit pipeline (`pipeline.yml`) triggers on every push to `main` and on pull requests
targeting `main`. Documentation-only changes (`.md` files, `docs/`) skip the pipeline entirely.
A weekly scheduled run on Wednesdays at 04:00 UTC also fires the full pipeline regardless of
changes, to catch upstream dependency regressions.

All lint jobs run in parallel. The Docker image build only starts after every lint job passes.

```ditaa
  push to main / pull request / schedule
                    |
                    v
  +-----------------+-----------------+------------------+
  |                 |                 |                  |
  v                 v                 v                  |
+-----------+  +----------+  +----------------+         |
| shellcheck|  | lint     |  | lint-dockerfile|         |
|           |  | (matrix: |  |                |         |
|           |  | yaml,    |  |                |         |
|           |  | workflows|  |                |         |
|           |  | filenames|  |                |         |
|           |  | folders  |  |                |         |
|           |  | gherkin  |  |                |         |
|           |  | md-links)|  |                |         |
+-----------+  +----------+  +----------------+         |
        |            |                |                  |
        +------------+----------------+                  |
                     |                                   |
                     v                                   |
          +------------------------+                     |
          | build-image            |                     |
          | docker build & push    |                     |
          | tag: :<commit-sha>     |                     |
          | (includes SBOM +       |                     |
          |  provenance)           |                     |
          +---+--------+-----------+                     |
              |        |                                 |
              v        v                                 |
  +-------------+  +----------------+                   |
  | docker-scout|  | sonar-analysis |                   |
  | compare vs  |  | task go:test   |                   |
  | :latest,    |  | + SonarQube    |                   |
  | post PR     |  | scan           |                   |
  | comment     |  +----------------+                   |
  +------+------+                                       |
         |                                              |
         v  (main branch only)                          |
  +--------------------+                                |
  | publish-edge       |                                |
  | retag :sha -> :edge|                                |
  +---+----------+-----+                                |
      |          |                                      |
      v          v (main, not dependabot)               |
  +--------+  +----------------+                        |
  | cleanup|  | release-code   |                        |
  | delete |  | semantic       |                        |
  | :sha   |  | release +      |                        |
  | from   |  | git tag        |                        |
  | DockerH|  +----------------+                        |
  +--------+                                            |
```

The `build-image` step builds and immediately pushes the image tagged with the full Git commit
SHA. This SHA-tagged image is ephemeral - it is cleaned up by `cleanup-dockerhub` at the end of
every run. The `publish-edge` step re-tags that image as `:edge` on the Docker Hub registry, so
`:edge` always reflects the latest passing commit on `main`.

The `sonar-analysis` job re-runs `task go:test` to produce a fresh `coverage.out` and then
uploads the results to SonarQube alongside a full source scan. The `docker-scout` job compares
the new image against the current `:latest` for vulnerability regressions and posts the
comparison as a pull-request comment.

The `release-code` job calls a shared reusable workflow from `sommerfeld-io/.github` that
handles semantic versioning, changelog generation, and GitHub Release creation based on
Conventional Commit messages.

### Release pipeline

A GitHub Release being created triggers `release.yml`. At that point the `:edge` image - which
has already passed every gate in the commit pipeline - is promoted to the release version tag and
to `:latest`. No rebuild occurs; only the tag is added to the existing manifest.

```ditaa
  GitHub Release created (tag: v1.x.y)
                   |
       +-----------+-----------+
       |                       |
       v                       v
  +--------------------+  +----------------------+
  | update-release-    |  | publish-release      |
  | notes              |  |                      |
  |                    |  | :edge --> :1.x.y     |
  | (shared workflow)  |  | :edge --> :latest    |
  +--------------------+  | update DockerHub     |
                          | description          |
                          +----------+-----------+
                                     |
                                     v
                          +----------------------+
                          | docker-scout         |
                          | CVE scan :latest     |
                          | upload SARIF to      |
                          | GitHub Security tab  |
                          +----------------------+
```

After a release the `:latest` tag on Docker Hub always points to the most recent stable version,
and the `:edge` tag continues to track the tip of `main` for integration consumers.
