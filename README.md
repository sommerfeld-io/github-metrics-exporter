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

## Software Tags and Versioning

Learn about our tagging policy and the difference between rolling tags and immutable tags [on our documentation page⁠](https://github.com/sommerfeld-io/.github/blob/main/docs/tags-and-versions.md).

## Usage

Run the exporter with Docker Compose. Create a `ghme-config.yml` configuration file (see `src/ghme-config.yml` on GitHub for a documented reference) and mount it into the container:

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

## Architecture Decisions

All issues labeled as `ADR` [are tracked as GitHub issue](https://github.com/sommerfeld-io/github-metrics-exporter/issues?q=is%3Aissue+label%3AADR).

## Risks and Technical Debts

All issues labeled as `risk` (= some sort of risk or a technical debt) or `security` (= disclosed security issues - e.g. CVEs) [are tracked as GitHub issue](https://github.com/sommerfeld-io/github-metrics-exporter/issues?q=is%3Aissue+label%3Asecurity%2Crisk+is%3Aopen) and carry the respective label.

## Security Artifacts

Every [Release](https://github.com/sommerfeld-io/github-metrics-exporter/releases) includes a Software Bill of Materials (SBOM) in [SPDX JSON](https://spdx.dev) format, attached as a release asset named `sbom.spdx.json`.

The SBOM provides a full inventory of all software components bundled in the container image - including Alpine OS packages and Go module dependencies - with their associated license identifiers(e.g., `MIT`, `Apache-2.0`). It is generated automatically during the release pipeline.

To download the SBOM for a specific release, navigate to the [Releases page](https://github.com/sommerfeld-io/github-metrics-exporter/releases), open the target release, and download `sbom.spdx.json` from the **Assets** section.

## Contact

Feel free to contact me via <sebastian@sommerfeld.io>.
