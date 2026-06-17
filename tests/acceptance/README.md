# Acceptance Tests

GoDog (Cucumber/Gherkin) acceptance tests for the GitHub Metrics Exporter HTTP server.

## What is tested

The Gherkin scenarios in `features/` describe the observable behavior of the HTTP server from a client's perspective as defined in the `*.feature` files.

## How the tests work

The tests start a **real HTTP server** - not a mock - from the test entrypoint. The Go packages under `src/internal/` are imported directly by the test binary.

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

- **`main.go`** - the wiring in the entrypoint (flag parsing, `ListenAndServe` on `:9400`) is not exercised.
- **Linker-injected `CommitSHA`** - the `-ldflags` value is only set during `go build`. In tests the variable defaults to `"development"`.

### Outlook

For that gap, a **smoke test against the compiled binary** (or the Docker image) is the right tool. The smoke test only needs to confirm the binary starts and the endpoints respond - it does not need to re-verify the full HTTP behavior already covered here. See [Miletone 2 - Smoke Tests](https://github.com/sommerfeld-io/github-metrics-exporter/milestone/4) for details in the implementation.

## Running the tests

```bash
# From the repo root
task go:acceptance-test

# Directly
cd tests/acceptance && go test -v ./...

# Against the Docker build stage (human-controlled)
task docker:acceptance-test
```
