---
id: 260926-ewj
phase: quick
plan: 260926-ewj
subsystem: compiler
tags: [cgen, schema-2, evidence, maturity]
requires: [Phase 18 computed-match terminal streaming]
provides: [Historical generated C compatibility, current source corpus count]
affects: [compiler emission, evidence verification, project maturity snapshot]
tech-stack:
  added: []
  patterns: [Conditional generated support helper emission]
key-files:
  created: []
  modified:
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_program.go
    - .planning/LANGUAGE-MATURITY.md
decisions:
  - "Emit the JSON string-content helper only for payload branch returns that use the terminal writer."
metrics:
  duration: 12min
  completed: 2026-09-26
  status: complete
actuals:
  tasks: 2
  commits: 2
plan_head_before: d8e65978d3420b2165db5b2031617fb8cf205938
---

# Quick Task 260926-ewj Summary

Conditional schema-2 string streaming restores frozen emitter bytes for entries without payload terminal values, while retaining the long-tag terminal path.

## Accomplishments

- Reserved `lang_write_json_string_content` in both generated identifier namespaces.
- Kept the pre-streaming inline JSON string writer for entries whose terminal writer does not consume the content helper.
- Passed the helper requirement from the same checked entry return shape that selects terminal value writing.
- Re-assessed the current corpus snapshot on 2026-09-26 at 142 `.lang` programs and 4,661 lines; retained the 33-line average and 193-line maximum.

## Verification

- Focused cgen/session/evidence regressions and Phase 18 long-tag controls: passed.
- Focused maturity count and self-describing documentation controls: passed.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./...`: all packages passed, including `internal/compiler/session` (131.727s).
- Scoped `git diff --check`: passed.

## Deviations

- The executor's initial staging attempt failed because `.git/index.lock` was not writable in its sandbox. The orchestrator completed both task commits through the GSD commit command with escalated filesystem access.
- Unrelated pre-existing workspace changes were left untouched.

## Commits

- `4cc5124` — `fix(260926-ewj): preserve generated C compatibility`
- `7be22ab` — `docs(260926-ewj): refresh language maturity corpus count`

## Self-Check

- All three planned files are modified and present.
- The focused, full-suite, and diff checks passed as recorded above.
- Both scoped task commits are present: `4cc5124` and `7be22ab`.
