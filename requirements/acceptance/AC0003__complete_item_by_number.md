---
status: active
---

Marking an item complete takes it off the list of incomplete items.

## Scenarios

- [AC0003.1] A completed item is no longer shown on the list.
- [AC0003.2] The other items keep their order after one is completed.
- [AC0003.3] Completing a number that is not on the list is refused and leaves the list unchanged.
- [AC0003.4] Completing without a single whole number is refused and leaves the list unchanged.

## Why

The list is for what is still left to do, so finishing something has to take
it off. The item is picked by the number the list shows next to it, so
`go-todo done 2` completes whatever `go-todo list` showed as number 2.

Completing one item must not reshuffle the rest: the others stay in the order
they were in, and the ones after it simply move up a number.

A number with no item behind it (past the end of the list) is an error, not a
silent no-op, so a typo is noticed. Anything that is not one whole number of at
least 1 (nothing, a word, zero, a negative number, two numbers) is a usage
mistake and gets the usage message. In both cases nothing is saved.

The command name `done`, the `[x]` confirmation it prints and the data file are
implementation details.
