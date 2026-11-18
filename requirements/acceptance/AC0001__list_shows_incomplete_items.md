---
status: active
---

Listing the to-do items shows every incomplete item and skips done ones.

## Scenarios

- [AC0001.1] Each incomplete item is shown on its own line as its number, ". [ ] " and its title.
- [AC0001.2] Done items are not shown.
- [AC0001.3] Items are numbered from 1 in the order they are listed.

## Why

The list is for deciding what to do next, so it shows only what is still left
to do. Items appear in the order they are stored.

Each item carries its place in the list (`1. [ ] Buy milk`), so another
command can point at it by number. The number counts only the items shown, so
there are no gaps where done items are hidden. That also means numbers shift
when an earlier item is completed: the number is where the item is on the list
now, not a permanent id.

Today the list is read with `go-todo list`, and the items come from a JSON
data file on the user's machine. Until that file exists, the list is a set of
sample items built into the executable. All of these are implementation
details: a different command name or storage would still have to keep this
promise.
