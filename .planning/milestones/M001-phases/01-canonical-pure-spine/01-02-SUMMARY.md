---
phase: 01-canonical-pure-spine
plan: "02"
subsystem: frontend
tags: [formatter, lossless-cst, recovery, property-testing, fuzzing]
requires: [01-01]
provides:
  - "Single stdout-only canonical S1 formatter"
  - "Lossless bounded malformed-input recovery"
  - "Generated, fuzz-seed, and concurrent read-only evidence"
affects: [frontend, diagnostics, agent-feedback, ownership-syntax]
actuals:
  tasks: 3
  commits: 3
tech-stack:
  added: ["Go testing/quick", "Go native fuzz seed corpus"]
  patterns: ["formatter-owned projection", "advance-or-boundary recovery", "deterministic generated properties"]
key-files:
  created: [internal/compiler/syntax/format.go, internal/compiler/syntax/syntax_test.go, testdata/phase1/comments.lang, testdata/phase1/malformed.lang]
  modified: [cmd/lang/main.go, internal/compiler/syntax/lexer.go, internal/compiler/syntax/parser.go, internal/compiler/session/session.go, internal/compiler/session/session_test.go]
key-decisions:
  - "Phase 1 formatting writes canonical bytes only to stdout; format --check is read-only and in-place mutation remains deferred."
  - "Syntax diagnostics are capped at 20 primary facts plus one stable truncation fact."
  - "Recovery preserves top-level declaration boundaries and asserts every skipping path advances or is already at a caller-owned boundary."
patterns-established:
  - "Canonical source is a lossless token/CST projection, not an AST pretty-printer."
  - "Ordinary tests run deterministic fuzz seeds; unbounded coverage-guided fuzzing is never in the default loop."
requirements-completed: [SYN-02, SYN-03, SYN-04]
coverage:
  - id: D1
    description: "Formatting is byte-idempotent, semantically stable, and preserves ordered comment text."
    requirement: SYN-02
    verification:
      - kind: property
        ref: "internal/compiler/syntax/syntax_test.go#TestFormatIdempotent and TestCommentPreservation"
        status: pass
    human_judgment: false
  - id: D2
    description: "Malformed bytes and syntax preserve their CST bytes, terminate with bounded deterministic diagnostics, and do not swallow the following declaration."
    requirement: SYN-03
    verification:
      - kind: property
        ref: "internal/compiler/syntax/syntax_test.go#TestRecovery, TestInvalidUTF8, and FuzzParseFormat"
        status: pass
    human_judgment: false
  - id: D3
    description: "One thousand fixed-seed generated programs reach a canonical fixed point, and concurrent format/check processes do not mutate source."
    requirement: SYN-04
    verification:
      - kind: generated
        ref: "internal/compiler/syntax/syntax_test.go#TestGeneratedRoundTrips and internal/compiler/session/session_test.go#TestConcurrentReadOnlyCommands"
        status: pass
    human_judgment: false
duration: 4min
completed: 2026-09-03
status: complete
---

# Phase 1 Plan 2: Canonical Lossless Frontend Summary

**Lang now has one calm canonical source projection that remains lossless and bounded through generated programs, malformed AI edits, and concurrent read-only commands.**

## Performance

- **Duration:** 4 min
- **Started:** 2026-09-03T19:03:48Z
- **Completed:** 2026-09-03T19:07:55Z
- **Tasks:** 3
- **Files modified:** 9

## Accomplishments

- Added `lang format FILE` and `lang format --check FILE` without granting either command permission to mutate source.
- Preserved original bytes in the CST for valid, malformed, and invalid UTF-8 inputs while bounding diagnostics and declaration-local recovery.
- Exercised 1,000 deterministic generated programs, native fuzz seeds, and concurrent CLI processes; semantic order, comments, canonical bytes, and source digests remained stable.

## Task Commits

1. **Task 1: Canonical lossless formatter** — `b866a36`
2. **Task 2: Bounded malformed recovery** — `1708af6`
3. **Task 3: Generated and concurrent properties** — `0bd077a`

## Decisions Made

- Canonical formatting is derived from lossless tokens, so comments remain first-class source data while AST source spans may legitimately change.
- Declaration keywords and closing braces are caller-owned recovery boundaries; recovery never consumes them accidentally.
- The property lane is deterministic and cheap enough for every edit, while coverage-guided fuzzing remains an opt-in bounded discovery lane.

## Deviations from Plan

None - plan executed as written.

## Issues Encountered

- Semantic round-trip comparison initially included byte spans; the test now explicitly erases incidental locations before comparing AST meaning.
- Exported `fn` entries and top-level function declarations share a prefix, so recovery uses bounded non-trivia lookahead to avoid swallowing either form.

## User Setup Required

None.

## Next Phase Readiness

The source surface is stable enough to bind deterministic schema identities and evidence manifests in Plan 01-03. Structured CLI output can now reuse the same diagnostics and semantic results without prose scraping.

## Self-Check: PASSED

- All named task tests passed, including the fuzz seed corpus and 1,000 generated cases.
- `go test ./...`, `go test -race ./...`, and `go vet ./...` passed with the explicit sandbox cache.
- Every key file exists and all three task commits are present.

---
*Phase: 01-canonical-pure-spine*
*Completed: 2026-09-03*
