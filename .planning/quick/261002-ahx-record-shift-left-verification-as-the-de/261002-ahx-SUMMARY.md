---
phase: quick
plan: 261002-ahx
subsystem: testing
tags: [go, markdown, verification]
requires: []
provides:
  - Section-scoped Evidence index local-link integrity check with a broken-link negative control
  - Shift-left verification policy and clarified Phase 25 human review boundary
affects: [phase25-validation, verification-policy]
tech-stack:
  added: []
  patterns: [standard-library Markdown link check]
key-files:
  created: []
  modified:
    - .planning/PROJECT.md
    - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md
    - internal/compiler/native/phase25_utility_test.go
decisions:
  - Keep the check specific to the README Evidence index and ordinary inline links.
  - Reserve human review for clean-checkout comprehension; automated link resolution proves navigation only.
metrics:
  duration: 5m
  completed: 2026-10-02
status: complete
actuals:
  tokens: 2523.25
  tasks: 1
  commits: 1
commits: 1
plan_head_before: d163d9e11b268d0934c20825f9d56f473e6fbfcd
---

# Quick Task 261002-ahx Summary

**The Phase 24/25 README Evidence index now rejects missing local file links, and verification guidance assigns repeatable acceptance checks to automation while preserving subjective comprehension for human review.**

## Accomplishments

- Added a standard-library check that bounds parsing to the `Evidence index` heading, requires local file links, resolves relative destinations from the README directory, strips fragments, and requires regular files.
- Added an in-memory negative control for `missing-evidence-control.md`, which must fail with a diagnostic naming that target.
- Amended the existing `.planning/PROJECT.md` preference and Phase 25 validation record. Existing authorization and stale-report guidance and historical hosted receipts remain intact.
- Existing CI Go suites include the native package; no workflow or dependency changes were needed.

## Verification

Command:

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test -count=1 -run '^TestPhase25EvidenceIndex$' ./internal/compiler/native
```

Result: **passed**, `ok github.com/szTheory/schway/internal/compiler/native` (0.006s). The exact focused command passed before and after the commit. `git diff --check` also passed.

## Commit

- `123f8a6` — `test(261002-ahx): verify phase25 evidence index links`
- Measured plan commits: 1, from base `d163d9e11b268d0934c20825f9d56f473e6fbfcd`.

## Deviations

None. No full suite, hosted CI, or push was run.

---
*Plan: 261002-ahx*
*Completed: 2026-10-02*
