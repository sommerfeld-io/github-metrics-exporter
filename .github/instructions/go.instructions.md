---
applyTo: "src/**/*.go,tests/acceptance/**/*.go"
---

# Go Code Style and Project Structure

Apply these instructions to all Go code in `src/`.

## Style priorities

- Prefer clarity, simplicity, maintainability, and consistency over clever code.
- Use standard language features and the standard library before introducing extra abstraction or dependencies.
- Keep local style consistent with nearby code unless that would worsen a deviation from the Google style guide.
- Format all Go code with `gofmt`; keep imports compatible with `goimports`.

## Project structure

- Keep all Go source code in `src/`.
- Keep `src/main.go` lightweight and focused on orchestration such as loading config, wiring dependencies, and starting the app.
- Put private application logic in `src/internal/**`.
- Put non-private packages in other `src/**` packages with short, descriptive package names.
- Keep acceptance tests in `tests/acceptance`, with Gherkin feature files in `tests/acceptance/features`.

## Naming

- Use short, descriptive, lowercase package names. Avoid underscores, dashes, mixed case, generic names like `util`, and unnecessary plurals.
- Use `PascalCase` for exported identifiers and `camelCase` for unexported identifiers.
- Prefer concise names that fit the local context and avoid stutter such as `user.Service` instead of `user.UserService`.
- Use short but meaningful local variable names.
- Use short, consistent receiver names.
- Keep interface names small and idiomatic; define interfaces where they are consumed.

## Functions and methods

- Keep functions small and focused on one responsibility.
- Prefer helper functions and better names over line-by-line comments that only restate the code.
- Use comments to explain why logic exists, important edge cases, or non-obvious trade-offs.
- Prefer the simplest implementation that keeps call sites easy to understand and hard to misuse.

## Comments and documentation

- Use GoDoc-style `//` comments.
- Comment all exported packages, types, variables, constants, functions, and methods.
- Keep comments short, specific, and focused on behavior or rationale instead of narration.
- Avoid redundant comments that only describe what the code already says.
- Add a `PACKAGE.md` file to document the purpose of each package.

## Line-level comments

Avoid commenting on *what* the code does. Use comments to explain the *why* when the logic is non-obvious or complex. Prefer refactoring complex logic into well-named helper functions over adding inline comments.

```go
// Bad: comment only restates what the code already says
// check if restaurant exists in the system
if restaurant == nil {
    return err
}

// Good: extract intent into a well-named function
if err := checkRestaurantExists(restaurant); err != nil {
    return err
}
```

## Imports

Split imports into three groups separated by blank lines:

1. Standard library
2. Third-party dependencies
3. Local packages from `github.com/sommerfeld-io/github-metrics-exporter/src/...`

## Error handling

- Return `error` as the last return value from functions that can fail.
- Handle errors explicitly and early. Prefer the usual Go pattern:

```go
if err := someFunction(); err != nil {
    return err
}
```

- Avoid wrapping normal execution in `else` branches after an error check.
- Wrap lower-level errors with context by using `fmt.Errorf("context: %w", err)`.
- Define sentinel errors for expected failure cases that callers may need to detect.
- Lower-level packages should return or wrap errors, not log them.
- Boundary layers such as `main.go`, `server/`, or scraper entry points decide whether to log, convert, or terminate.
- Log errors to `stderr` and all non-error output to `stdout`, using structured logging whenever possible.
- Reserve `panic` for truly unrecoverable programmer or runtime faults, not normal control flow.

## Testing

- Follow TDD unconditionally. The test comes first - no exceptions.
- Keep unit tests next to the code they verify and use the `_test.go` suffix.
- Name tests with the `TestXxxShouldYyy` pattern.
- Cover happy paths, negative cases, and inversions so tests prove both what should and should not happen.
- Prefer table-driven tests when they improve clarity and reduce repetition.

### Acceptance tests and Gherkin feature files

- `.feature` files define GoDog acceptance tests. They are **not** used to derive unit tests.
- Step definitions live in `tests/acceptance/*_steps_test.go`.
- Wire the GoDog suite in `tests/acceptance/suite_test.go`.
- Acceptance test Go code in `tests/acceptance/` follows the same TDD discipline as production
  code. Any non-trivial helper extracted into a non-test file must have its own `*_test.go` unit
  tests.
