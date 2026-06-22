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

## Software Tags and Versioning

Learn about our tagging policy and the difference between rolling tags and immutable tags [on our documentation page⁠](https://github.com/sommerfeld-io/.github/blob/main/docs/tags-and-versions.md).

## Usage

> [!NOTE]
> This usage guide targets the `:latest` container image and the corresponding most recent version. Usage, configuration options, and features may differ in older releases or `:edge` release candidates. For documentation of a specific version, refer to the corresponding branch or tag in the [GitHub repository](https://github.com/sommerfeld-io/github-metrics-exporter).

Run the exporter with Docker Compose. Create a `ghme-config.yml` configuration file (see `src/ghme-config.yml` on GitHub for a documented reference), mount it into the container, and pass the token from your host environment:

```yaml
services:
  github-metrics-exporter:
    image: sommerfeldio/github-metrics-exporter:latest
    environment:
      GITHUB_TOKEN: ${GITHUB_METRICS_EXPORTER_TOKEN}
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

### GitHub Token

The exporter requires a GitHub personal access token (PAT) to query the GitHub API. Set the `GITHUB_TOKEN` environment variable before starting the exporter. If the variable is absent or empty, the exporter exits immediately with a non-zero code and a clear error message - no silent failures at scrape time.

**Required permissions (fine-grained PAT):**

| Permission | Access     |
|------------|------------|
| Actions    | Read-only  |

**Security:** Never hardcode the token in `docker-compose.yml` or any version-controlled file. Store it as an environment variable on the host and reference it with `${VARIABLE_NAME}` substitution.

On your machine, export the token under a specific name to avoid collisions with other tools:

```bash
export GITHUB_METRICS_EXPORTER_TOKEN=github_pat_...
```

## Architecture Decisions

All issues labeled as `ADR` [are tracked as GitHub issue](https://github.com/sommerfeld-io/github-metrics-exporter/issues?q=is%3Aissue+label%3AADR).

## Risks and Technical Debts

All issues labeled as `risk` (= some sort of risk or a technical debt) or `security` (= disclosed security issues - e.g. CVEs) [are tracked as GitHub issue](https://github.com/sommerfeld-io/github-metrics-exporter/issues?q=is%3Aissue+label%3Asecurity%2Crisk+is%3Aopen) and carry the respective label.

## Security Artifacts

Starting with version `0.4.1`, every [Release](https://github.com/sommerfeld-io/github-metrics-exporter/releases) includes a Software Bill of Materials (SBOM) in [SPDX JSON](https://spdx.dev) format, attached as a release asset named `sommerfeldio-github-metrics-exporter_{version}.spdx.json` (e.g. `sommerfeldio-github-metrics-exporter_0.4.1.spdx.json`).

The SBOM provides a full inventory of all software components bundled in the container image - including Alpine OS packages and Go module dependencies - with their associated license identifiers(e.g., `MIT`, `Apache-2.0`). It is generated automatically during the release pipeline.

To download the SBOM for a specific release, navigate to the [Releases page](https://github.com/sommerfeld-io/github-metrics-exporter/releases), open the target release, and download the `.spdx.json` file from the **Assets** section.

## Contact

Feel free to contact me via <sebastian@sommerfeld.io>.
