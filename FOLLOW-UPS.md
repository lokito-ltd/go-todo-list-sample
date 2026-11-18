# Follow-ups

- **Gremlins wrapper breaks on `switch` case conditions.** Go's coverage
  instrumentation puts no counter on `case` expressions, so Gremlins marks
  mutants there "NoCoverage" even when tests run them. `tools/ktty/gremlins`
  then exits 1 ("survived or no test reached them") while its report counts
  no survivors, and verify fails with "report contradicts the command's
  survived outcome" — a broken check rather than a finding. Seen on
  `feat/add-command` (2026-11-18), worked around by writing the dispatch with
  `if` statements. Fix the wrapper so NoCoverage mutants are reported
  consistently (exit code and report agree), and decide whether case
  expressions should count as reached.
  Upstream: known Gremlins bug go-gremlins/gremlins#298 (open, reported on
  v0.6.0). Fixes proposed in #299 and #325, both unmerged as of 2026-11-18;
  v0.6.0 is still the latest release. Bump Gremlins once a release has it.
  Alternative that runs every mutant (no coverage filter, so no blind spot):
  avito-tech/go-mutesting (JSON report, file-level targeting only).
