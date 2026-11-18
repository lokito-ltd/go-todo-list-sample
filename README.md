# go-todo

A standalone command-line to-do app for Apple Silicon and Intel Macs. Run `go-todo list` to display every incomplete item, `go-todo add <title>` to add one, and `go-todo done <number>` to mark one complete.

## Where items are kept

Items are saved in `~/.go-todo/items.json`, outside this repository. Set `GO_TODO_FILE` to keep them in a different file. Until the file exists, the list holds the sample items built into the executable from `internal/todo/items.json`; the first `add` saves them to the file along with the new item.

## Build and run

From the project root, with Go 1.25 or newer and Apple's command line tools installed:

```sh
make build
./bin/go-todo list
```

Expected output:

```text
1. [ ] Buy milk
2. [ ] Take cat to vet
3. [ ] Print forms at library
4. [ ] Wash the car
```

Done items are not shown, and the numbers count only the items shown. To add an item:

```sh
./bin/go-todo add Buy bread
```

This prints `5. [ ] Buy bread`, the line `list` will show for it. Every word after `add` becomes part of the title, so quotes are optional. New items are incomplete and go to the end of the list.

To mark an item complete, pass the number `list` shows for it:

```sh
./bin/go-todo done 2
```

This prints `[x] Take cat to vet`. The item disappears from `list`, and the items after it move up one number. A number with no item behind it prints `go-todo: no item 9 on the list` and exits with status 1.

Running `go-todo` with no command, an unknown one, `add` without a title, or `done` without a single whole number of at least 1 prints the usage to stderr and exits with status 2:

```text
usage: go-todo list
       go-todo add <title>
       go-todo done <number>
```

If the items file cannot be read or written, `go-todo` prints a `go-todo: ...` message to stderr and exits with status 1.

On a development Mac, install Go with `brew install go` and Apple's command line tools with `xcode-select --install` if needed. The build uses `lipo` to combine and verify both architectures. Copy `bin/go-todo` to a compatible Mac to run it without a Go installation.

## Development

| Command | What it does |
| --- | --- |
| `make run` | Run `go-todo list` from source. |
| `make test` | Run unit, acceptance, and end-to-end tests with Go's built-in test runner. |
| `make check` | Check formatting, run `go vet`, and run all tests. |
| `make mutate` | Use Gremlins to change CLI code and check that the tests catch the changes. |
| `make build` | Build the universal macOS executable at `bin/go-todo`. |

Unit tests live beside the command in `internal/cli`. Acceptance tests in `tests/acceptance` exercise startup in memory. End-to-end tests in `tests/e2e` compile the entry point and run it from an empty directory. Use `go test ./...` for all suites, or run one with a package path such as `go test ./tests/e2e`.

Go modules record third-party versions in `go.mod` and checksums in `go.sum`. The app uses the standard library. Gremlins v0.6.0 is a development tool pinned in the module and invoked by `make mutate`. The first mutation run needs network access to download its dependencies. The check requires at least one killed mutation and fails for surviving or uncovered mutations.
