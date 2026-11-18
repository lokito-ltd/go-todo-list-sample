# go-todo

A standalone command-line to-do app for Apple Silicon and Intel Macs. Run the executable with no arguments to display one unchecked item.

## Build and run

From the project root, with Go 1.25 or newer and Apple's command line tools installed:

```sh
make build
./bin/go-todo
```

Expected output:

```text
[ ] Hello World
```

On a development Mac, install Go with `brew install go` and Apple's command line tools with `xcode-select --install` if needed. The build uses `lipo` to combine and verify both architectures. Copy `bin/go-todo` to a compatible Mac to run it without a Go installation.

## Development

| Command | What it does |
| --- | --- |
| `make run` | Run the app from source. |
| `make test` | Run unit, acceptance, and end-to-end tests with Go's built-in test runner. |
| `make check` | Check formatting, run `go vet`, and run all tests. |
| `make mutate` | Use Gremlins to change CLI code and check that the tests catch the changes. |
| `make build` | Build the universal macOS executable at `bin/go-todo`. |

Unit tests live beside the command in `internal/cli`. Acceptance tests in `tests/acceptance` exercise startup in memory. End-to-end tests in `tests/e2e` compile the entry point and run it from an empty directory. Use `go test ./...` for all suites, or run one with a package path such as `go test ./tests/e2e`.

Go modules record third-party versions in `go.mod` and checksums in `go.sum`. The app uses the standard library. Gremlins v0.6.0 is a development tool pinned in the module and invoked by `make mutate`. The first mutation run needs network access to download its dependencies. The check requires at least one killed mutation and fails for surviving or uncovered mutations.
