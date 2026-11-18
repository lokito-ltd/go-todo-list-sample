---
requirements:
  - id: AC0001
    status: active
    statement: Listing the to-do items shows every incomplete item and skips done ones.
    scenarios:
      - '[AC0001.1] Each incomplete item is shown on its own line as its number, ". [ ] " and its title.'
      - '[AC0001.2] Done items are not shown.'
      - '[AC0001.3] Items are numbered from 1 in the order they are listed.'
  - id: AC0003
    status: active
    statement: Marking an item complete takes it off the list of incomplete items.
    scenarios:
      - '[AC0003.1] A completed item is no longer shown on the list.'
      - '[AC0003.2] The other items keep their order after one is completed.'
      - '[AC0003.3] Completing a number that is not on the list is refused and leaves the list unchanged.'
      - '[AC0003.4] Completing without a single whole number is refused and leaves the list unchanged.'
---

# Mark items complete by number

## Report

```text
prediction: exact · protected files: 1 · risk: LOW
AC scenarios: 5 added, 1 reworded, 0 retired
tests: 23 added, 29 modified, 1 removed
change coverage 100% (38 of 38 lines) · mutants: 14 caught, 0 survived (by the tool's account)
follow-ups +0 −0 · code +54 −6 lines
2 setup warnings
```

## Why

`go-todo` could list and add items, but nothing could be ticked off, so an
item stayed on the list for ever. The operator asked for a way to mark an item
complete so it no longer shows among the items still to do, and chose to pick
the item by number.

## What

`go-todo list` now numbers each incomplete item from 1 (`1. [ ] Buy milk`).
The number counts only the items shown, so there are no gaps where done items
are hidden. `go-todo add` prints the new item with the number `list` will give
it.

`go-todo done <number>` marks that item done, saves the list and prints
`[x] <title>`. The item disappears from `list`, and the items after it move up
one number. A number with no item behind it prints
`go-todo: no item <n> on the list` and exits 1. Anything that is not a single
whole number of at least 1 prints the usage, now three lines, and exits 2.
Neither error changes the list.

Out of scope: undoing a completion, listing done items, removing or editing
items, and ids that stay fixed when an earlier item is completed.

AC0001.1 was reworded for the numbered format; its old test was renamed to
the new scenario text, which the report counts as one test removed and one
added.

## Departures

None.

## How

**Numbers are positions in the shown list, not stored ids.** This keeps the
data file unchanged and makes `done 2` mean exactly what `list` showed as 2.
The cost is that numbers shift after a completion, so a user must list again
before completing a second item further down. Rejected: numbering every stored
item including done ones, which leaves gaps that do not match the screen.

**`todo.Complete` is a pure function over the slice.** It counts incomplete
items and marks the nth, so the counting rule is unit-tested without files,
and the CLI only loads, calls it, and saves.

**The confirmation is `[x] <title>` without a number.** After completion the
item's old number already belongs to the next item, so printing it would
mislead.

**Dispatch stays as plain `if` blocks**, as for `add`, because Gremlins
cannot see mutants in `switch` case expressions (see `FOLLOW-UPS.md`).
