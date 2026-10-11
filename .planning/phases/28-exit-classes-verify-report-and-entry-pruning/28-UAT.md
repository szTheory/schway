---
status: testing
phase: 28-exit-classes-verify-report-and-entry-pruning
source: [28-01-SUMMARY.md, 28-02-SUMMARY.md, 28-03-SUMMARY.md, 28-04-SUMMARY.md, 28-05-SUMMARY.md, 28-06-SUMMARY.md, 28-07-SUMMARY.md]
started: 2026-10-11T03:21:51Z
updated: 2026-10-11T03:21:51Z
---

## Current Test
<!-- OVERWRITE each test - shows where we are -->

number: 1
name: Uncalled function no longer blocks check, build and run
expected: |
  From the repo root: `go build -o /tmp/schway ./cmd/schway`, then
  `/tmp/schway check examples/phase28/unused_function.schway` exits 0,
  `/tmp/schway build examples/phase28/unused_function.schway --output /tmp/uf` exits 0,
  and `/tmp/schway app run /tmp/uf -- 7` prints 7 and exits 0.
awaiting: user response

## Tests

### 1. Uncalled function no longer blocks check, build and run
expected: From the repo root, build the CLI (`go build -o /tmp/schway ./cmd/schway`). `/tmp/schway check examples/phase28/unused_function.schway` exits 0, `/tmp/schway build examples/phase28/unused_function.schway --output /tmp/uf` exits 0, and `/tmp/schway app run /tmp/uf -- 7` prints 7 and exits 0.
result: [pending]

### 2. Ambiguous entry is a source refusal
expected: `/tmp/schway check testdata/phase12/payload_borrow_interaction.schway` and `/tmp/schway run --engine=interpreter testdata/phase12/payload_borrow_interaction.schway` each exit 2 with first diagnostic `core.entry_ambiguous`, naming the roots `choose` and `identity` and suggesting adding `fn main`.
result: [pending]

### 3. Same source gets the same exit class on every surface
expected: For `testdata/phase07/duplicate_function_name.schway` (relative path, from the repo root), `check`, `build ... --output /tmp/dup`, `run --engine=interpreter` and `run --engine=native` all exit 2 with `core.duplicate_function_id`, and the build leaves no /tmp/dup file.
result: [pending]

### 4. App verify writes a schema /2 report
expected: `/tmp/schway app verify examples/phase22/identity.schway --cases examples/phase22/identity.cases.json --report /tmp/r.json` exits 0, and /tmp/r.json starts with `"schema":"lang.app-verification/2"` and `"status":"pass"`.
result: [pending]

### 5. App verify source refusal exits 2 and still writes the report
expected: `/tmp/schway app verify testdata/phase12/payload_borrow_interaction.schway --cases examples/phase22/identity.cases.json --report /tmp/r2.json` exits 2, prints `schway app verify: source refused: core.entry_ambiguous` on stderr, and /tmp/r2.json exists as a fail report.
result: [pending]

### 6. README publishes the exit-code table
expected: README.md has an "Exit codes" section with two tables. The compiler table lists Success 0, Unrepaired 1, Invalid 2, Operational 3, Mismatch 4 and Usage 64. The `schway app run` table lists AppInputRefused 65, EventCapacity 69, AppTimeout 124 and AppLaunchError 125, and the section notes that the child's exit status passes through.
result: [pending]

## Summary

total: 6
passed: 0
issues: 0
pending: 6
skipped: 0
blocked: 0

## Gaps

[none yet]
