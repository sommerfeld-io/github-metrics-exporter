# GitHub Metrics Exporter

An exporter that translates GitHub API data into Prometheus metrics.

<!-- ===== START status badge ===== -->

[![Pipeline: Commit + Test](https://github.com/sommerfeld-io/github-metrics-exporter/actions/workflows/pipeline.yml/badge.svg)](https://github.com/sommerfeld-io/github-metrics-exporter/actions/workflows/pipeline.yml)
[![Pipeline: Release](https://github.com/sommerfeld-io/github-metrics-exporter/actions/workflows/release.yml/badge.svg)](https://github.com/sommerfeld-io/github-metrics-exporter/actions/workflows/release.yml)

<!-- ===== END status badge ===== -->

Provide insights into GitHub Actions, pipeline health and trends. Provide deep operational visibility into GitHub Actions, pipeline health trends.

- [Github Repository](https://github.com/sommerfeld-io/github-metrics-exporter)
- [Docker Hub Repository](https://hub.docker.com/r/sommerfeldio/github-metrics-exporter)
- [Sonarcloud Code Quality and Security Analysis](https://sonarcloud.io/project/overview?id=sommerfeld-io_github-metrics-exporter)
- [Where to file issues](https://github.com/sommerfeld-io/github-metrics-exporter/issues)
- [Project Board for Issues and Pull Requests](https://github.com/orgs/sommerfeld-io/projects/1/views/1)

## Usage

Run the exporter with Docker Compose. Create a `ghme-config.yml` configuration file (see [`src/ghme-config.yml`](src/ghme-config.yml) for a documented reference) and mount it into the container:

```yaml
services:
  github-metrics-exporter:
    image: sommerfeldio/github-metrics-exporter:latest
    ports:
      - 9400:9400
    volumes:
      - ./ghme-config.yml:/opt/ghme/ghme-config.yml:ro
```

The exporter exposes three endpoints once running:

| Endpoint    | Description                        |
|-------------|------------------------------------|
| `/`         | HTML landing page with build info  |
| `/metrics`  | Prometheus metrics                 |
| `/healthz`  | Plain-text health check            |

## AI-Assisted Development

This project is developed with the help of AI coding assistants. We use [Claude Code](https://code.claude.com) and [GitHub Copilot](https://github.com/features/copilot) in parallel.

AI assistants are guided by instruction files that define coding conventions, commit message rules, and project-specific guidelines. We maintain a single source of truth for these instructions and use symlinks so that each tool reads from the same file:

| Symlink         | Points to (master file)                   |
|-----------------|-------------------------------------------|
| `CLAUDE.md`     | `.github/copilot-instructions.md`         |
| `src/CLAUDE.md` | `.github/instructions/go.instructions.md` |

The master instruction files live under `.github/` and are committed to the repository:

- `.github/copilot-instructions.md` - project-wide conventions (commit messages, general style)
- `.github/instructions/go.instructions.md` - Go-specific coding guidelines for the `src/` subtree

The symlinks are also committed to git. Still, they are created locally by running `task symlinks` (see `taskfile.yml` for the exact `ln` commands used).

> [!NOTE]
> Because this repository uses symlinks, it is intended to be worked on in a Linux environment. The devcontainer setup makes it possible to work on a Windows host since the dev container is Linux, and the actual development environment runs inside that container.

## Architecture Decisions

All issues labeled as `ADR` [are tracked as GitHub issue](https://github.com/sommerfeld-io/github-metrics-exporter/issues?q=is%3Aissue+label%3AADR) and carry the respective label.

## Risks and Technical Debts

All issues labeled as `risk` (= some sort of risk or a technical debt) or `security` (= disclosed security issues - e.g. CVEs) [are tracked as GitHub issue](https://github.com/sommerfeld-io/github-metrics-exporter/issues?q=is%3Aissue+label%3Asecurity%2Crisk+is%3Aopen) and carry the respective label.

## Contact

Feel free to contact me via <sebastian@sommerfeld.io>.
