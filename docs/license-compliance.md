# License Compliance

This document describes the licensing policy for the GitHub Metrics Exporter, explains how license compliance is enforced during the build, and clarifies what the licenses of bundled third-party components mean for the application.

## Application License

The application source code and documentation are released under the **MIT License** (see [LICENSE.md](../LICENSE.md)).

## Go Module License Policy

All Go module dependencies are validated against an allowlist as a mandatory gate inside every build. Any dependency whose license falls outside this list causes the build to fail immediately.

**Allowed licenses:**

| SPDX identifier | License family     |
|-----------------|--------------------|
| `MIT`           | MIT License        |
| `Apache-2.0`    | Apache License 2.0 |
| `BSD-2-Clause`  | BSD 2-Clause       |
| `BSD-3-Clause`  | BSD 3-Clause       |
| `ISC`           | ISC License        |
| `Unlicense`     | The Unlicense      |

The check runs via [`google/go-licenses`](https://github.com/google/go-licenses) as part of `task go:build` and again inside the Dockerfile build stage, so the gate applies both during local development and in CI. The tool also generates a full license report listing every dependency alongside its identified license.

## Alpine Base Image Licenses

The Docker image is built on Alpine Linux. The Alpine base layer includes OS utilities - `busybox`, `musl`, `apk-tools`, and others - that carry their own licenses, including `GPL-2.0-only`, `MPL-2.0`, and `Zlib`. These are independent components that retain their own licenses and **do not affect the license of the Go application binary**.

The key distinction is that GPL's copyleft clause activates when GPL-licensed code is statically linked into your software to form a single combined work. A Go binary running alongside Alpine OS utilities in a container is not linking - the two are separate, isolated processes that share a filesystem layer. The Go application remains MIT-licensed. This is the same principle as running any application on a GPL Linux kernel: the kernel is GPL, the application is not.

Distributing the container image is legal under these terms. Recipients of the image receive:

- The MIT-licensed Go binary (source: this repository)
- Alpine OS utilities under their respective licenses (source: [Alpine Linux packages](https://pkgs.alpinelinux.org))

No action is required beyond preserving the license notices already present in each component.

## Audit Flow

License compliance is verified at two complementary layers:

```plain
Layer 1 - Build time (Go modules)
  go-licenses check
    → validates every Go import against the allowed list
    → fails the build immediately on any unlisted license
    → runs inside task go:build and inside the Dockerfile

Layer 2 - Release time (full container)
  syft / anchore/sbom-action
    → scans the published Docker image
    → captures Alpine OS package licenses and Go module licenses
    → resolves Go module licenses from the public Go module proxy
    → attaches the SPDX JSON SBOM to the GitHub release as an asset
```

The SBOM attached to each release provides a full, machine-readable inventory of every component in the shipped container - both the Go module layer and the Alpine OS layer - with resolved SPDX license identifiers. Download it from the **Assets** section of any [release](https://github.com/sommerfeld-io/github-metrics-exporter/releases).
