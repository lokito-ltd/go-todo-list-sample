---
status: active
---

Adding an item puts it on the to-do list as incomplete.

## Scenarios

- [AC0002.1] An added item is shown as incomplete at the end of the list.
- [AC0002.2] The words given after the command form the title, separated by single spaces.
- [AC0002.3] Adding with no title is refused and leaves the list unchanged.

## Why

A to-do list is only useful if things can be put on it. Something just added
is by definition not done yet, so it starts incomplete, and it goes to the end
so the list keeps the order things were added in.

The title is taken from every word after the command, so `go-todo add Buy
bread` works without quotes. Extra spaces between or around words are not
kept. An empty title would show as a bare `[ ] ` line that says nothing, so it
is refused with the usage message and the list is left as it was.

Added items are kept in a data file on the user's machine, outside the
project's tracked files, so they survive between runs. Today that file is
`~/.go-todo/items.json`, or the file named by `GO_TODO_FILE`. Both are
implementation details.
