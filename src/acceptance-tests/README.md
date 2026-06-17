# Acceptance Tests

GoDog (Cucumber/Gherkin) acceptance tests for the GitHub Metrics Exporter HTTP server.

## What is tested

The Gherkin scenarios in `features/` describe the observable behaviour of the HTTP server from a client's perspective as defined in the `*.feature` files.

## How the tests work

The tests start a **real HTTP server** - not a mock - from the test entrypoint. These files are part of the main Go module (`src/`) and import the production packages directly - no separate module, no `replace` directive.

```plain
TestMain (suite_test.go)
  │
  ├─ metrics.Register(prometheus.DefaultRegisterer)
  │     Registers ghme_exporter_info into the default Prometheus registry,
  │     mirroring what main.go does at startup.
  │
  ├─ httptest.NewServer(server.New())
  │     Starts a real net/http server on a random ephemeral port.
  │     No fixed port. No process. No Docker.
  │
  └─ godog.TestSuite{...}.Run()
        Executes every Gherkin scenario.
        Each step makes a real http.Get call to the ephemeral address.
```

After the suite finishes, `testSrv.Close()` shuts down the server.

## What is NOT tested here

These tests cover the Go packages but not the compiled binary or its entrypoint. Two things are outside their scope:

- **`main.go`** - the wiring in the entrypoint (`ListenAndServe` on `:9400`) is not exercised.
- **Linker-injected `CommitSHA`** - the `-ldflags` value is only set during `go build`. In tests the variable defaults to `"development"`.

For that gap, a **smoke test against the compiled binary** (or the Docker image) is the right tool. The smoke test only needs to confirm the binary starts and the endpoints respond - it does not need to re-verify the full HTTP behaviour already covered here.

## Running the tests

```bash
# Acceptance tests only (from repo root)
task go:test:acceptance

# Directly from src/
cd src && go test -v ./acceptance-tests/...

# Full build pipeline - acceptance tests run as a gate before go build
task go:build
```

## Relation to unit tests

Unit test coverage is measured separately with `go test ./internal/...` and written to `src/coverage.out`. Acceptance tests are excluded from that scope deliberately - they start a server and test observable HTTP behaviour, not individual package internals.

## Acceptance-test-side coverage (`acceptance-coverage.out`)

Running the acceptance tests with `-coverpkg=./internal/...` produces a second coverage report, `src/acceptance-coverage.out`. This file answers a different question from `coverage.out`:

| File                      | Question answered                                                          |
|---------------------------|----------------------------------------------------------------------------|
| `coverage.out`            | Are the individual branches in `internal/` exercised by targeted tests?    |
| `acceptance-coverage.out` | Which lines in `internal/` are reachable at all through the HTTP layer?    |

`acceptance-coverage.out` is a **diagnostic**, not a gate metric. The distinction matters:

- A line that appears covered in `acceptance-coverage.out` was merely *reachable* via an HTTP round-trip. It does not mean the surrounding logic was exercised for all inputs, error paths, or edge cases.
- Merging the two files (e.g. via `sonar.go.coverage.reportPaths`) would conflate reachability with correctness and inflate the unit-test coverage number. SonarQube gates are intentionally kept on `coverage.out` alone.
- A line that is *uncovered* in `acceptance-coverage.out` but covered in `coverage.out` is a warning sign: the unit test exercises it in isolation, but no HTTP path ever reaches it. That can indicate dead code or a missing acceptance scenario.

### Visualising the report locally

```bash
go tool cover -html=src/acceptance-coverage.out
```

This opens a browser view with line-by-line red/green coloring of the `internal/` packages as seen from the HTTP layer. Use it when auditing reachability or investigating whether a new endpoint is wired up end-to-end.
