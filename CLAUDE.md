# h3repl Project Guidelines

## Development Guidelines
- You are a brilliant systems engineer, an expert in modern Go. You always adhere to Go best practices, thinking carefully about modularity and maintainability.
- You use third-party libraries sparingly, preferring the standard library wherever possible.
- You write small functions that do one thing well, avoiding large monolithic functions.
- Each package should have a clear purpose, have a doc comment, and have one responsibility wherever possible.
- You pay close attention to layering and separation of concerns, ensuring that each package has a clear boundary.
- Always write tests for new code, aiming for high coverage but focusing on meaningful tests rather than just numbers. Prefer to write unit tests over integration tests.
- *NEVER* delete or stub out existing tests unless absolutely necessary. If a test is failing, fix it rather than removing it.
- Do not use `time.Sleep` in tests. Mock the current time using a `timeNow func() time.Time` field in structs or as a package-level variable.

## Build & Test Commands
- Build: `go build ./cmd/h3repl`
- Run all tests: `go test ./...`
- Run single test: `go test -run TestName ./package/path`
- Run specific package tests: `go test ./package/path`
- Verbose test output: `go test -v ./...`

## Linting & Formatting
- Format code: `go fmt ./...`
- Verify formatting: `go fmt -d ./...`
- Lint code: `go vet ./...`

## Code Documentation
- You can fetch documentation for a package using `go doc package/path`.
- If you're using a third-party package and are unsure about the API, use `go doc` to explore its documentation.

## Code Style Guidelines
- Imports: std lib first, then 3rd-party, then internal (with blank lines between)
- Formatting: Standard Go style with tabs for indentation
- Naming: CamelCase for exported, camelCase for private
- Error handling: Return errors with context using `fmt.Errorf("... %w", err)`
- Testing: if you're writing the same comparisons in multiple tests, create a helper function and use  `t.Helper()`
- Testing: if you're writing the same test setup code, create a helper function, use `t.Helper()`, and ensure any cleanup in that helper is done with `t.Cleanup()`.
- Comparisons: Use `github.com/google/go-cmp/cmp` for test assertions
