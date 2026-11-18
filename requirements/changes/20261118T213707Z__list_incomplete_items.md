---
requirements:
  - id: AC0001
    status: active
    statement: Listing the to-do items shows every incomplete item and skips done ones.
    scenarios:
      - '[AC0001.1] Each incomplete item is shown as "[ ] " followed by its title, one per line.'
      - '[AC0001.2] Done items are not shown.'
---

# List incomplete to-do items

## Report

```text
prediction: exact · protected files: 3 · risk: LOW
AC scenarios: 2 added, 0 reworded, 0 retired
tests: 10 added, 8 modified, 1 removed
change coverage 96% (25 of 26 lines) · mutants: 5 caught, 0 survived (by the tool's account)
follow-ups +0 −0 · code +62 −5 lines
2 setup warnings
```

## Why

The app printed a fixed `[ ] Hello World` placeholder. It should behave like a
to-do list: hold items, some done and some not, and show what is still left to
do. This is the first real behaviour and the project's first acceptance
criterion.

## What

`go-todo list` prints every incomplete item as `[ ] <title>`, one per line, in
stored order, and skips done items. The data is five hardcoded items: four
incomplete and one done ("Get a haircut"). Running `go-todo` with no command,
an unknown command or extra arguments prints `usage: go-todo list` to stderr
and exits with status 2. The Hello World output is retired.

Out of scope: adding, completing or removing items, runtime file storage, and
any display of done items.

The operator accepted one coverage gap: `cmd/go-todo/main.go:10`, the
`os.Exit(cli.Run(...))` line, is run only by the e2e tests in a separate
process, which in-process coverage cannot see.

## Departures

- The plan said the acceptance tests would be tagged by a comment above each
  test function. `ktty req trace` requires the test's name to be the
  scenario's text, which a Go function name cannot hold. So each scenario is a
  `t.Run` subtest named with the scenario text and its id.
- Revision 1 of the plan added the fix to the Gremlins mutation wrapper
  (`tools/ktty/gremlins/main.go`) after the first verify broke on it. The fix
  was approved before it was made.

## How

**Embedded JSON.** The items live in `internal/todo/items.json` and are
compiled into the binary with `//go:embed`, so the executable still runs on its
own on another Mac. Reading a file from disk at runtime was rejected for now:
it would mean choosing a location and handling a missing file before anything
writes one. Moving to disk later only changes where `todo.Data` comes from.

**Testable error path.** `cli.Run` delegates to an unexported `run` that takes
the JSON bytes. That lets unit tests feed broken JSON and check the
`read items` error without touching the embedded data.

**Mutation wrapper.** Gremlins exits 0 and writes no output file when a folder
has nothing to mutate, as `cmd/go-todo` does. The wrapper now reads that as
"no mutants" instead of failing. A real Gremlins failure still has a non-zero
exit and is still reported.
