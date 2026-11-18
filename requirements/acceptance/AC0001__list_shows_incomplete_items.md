---
status: active
---

Listing the to-do items shows every incomplete item and skips done ones.

## Scenarios

- [AC0001.1] Each incomplete item is shown as "[ ] " followed by its title, one per line.
- [AC0001.2] Done items are not shown.

## Why

The list is for deciding what to do next, so it shows only what is still left
to do. Items appear in the order they are stored.

Today the list is read with `go-todo list`, and the items come from a JSON
data file on the user's machine. Until that file exists, the list is a set of
sample items built into the executable. All of these are implementation
details: a different command name or storage would still have to keep this
promise.
