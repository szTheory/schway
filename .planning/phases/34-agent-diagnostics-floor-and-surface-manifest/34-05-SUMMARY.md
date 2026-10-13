---
phase: 34-agent-diagnostics-floor-and-surface-manifest
plan: "05"
subsystem: compiler-diagnostics
tags: [go, parser-recovery, surface-refusals, checker]

requires:
  - phase: 34-03
    provides: Imported diagnostic presentation from retained module snapshots
  - phase: 34-04
    provides: Bounded parser recovery and exact token-span diagnostic ownership
provides:
  - Seven finite, stable, prose-only surface refusal catalog entries
  - Exact parser refusals for the six parser-owned unsupported forms
  - Binding-aware Bool-tail refusal at the final identifier span
  - Focused evidence preserving admitted identifiers, calls, byte_at, and recovery controls
affects: [34-06 surface manifest, 34-07, 34-08, parser diagnostics, checker diagnostics]

actuals:
  tokens: 83920
  tasks: 3
  commits: 6
commits: 6
plan_head_before: 96863718cfeb2972eec95efe028f53a541276c63
plan_head_after: 4ab415501f0477ba61fd4c64332556fb09350359

tech-stack:
  added: []
  patterns:
    - Parser-owned exact source claims suppress only matching generic lexer bytes
    - Final result source spans remain AST metadata and do not enter core IR or semantic identity
    - Refusal hints are literal prose and carry no structured repair authority

key-files:
  created:
    - internal/compiler/surface/refusals.go
    - internal/compiler/syntax/parser_phase34_refusal_test.go
    - internal/compiler/check/check_phase34_refusal_test.go
  modified:
    - internal/compiler/syntax/lexer.go
    - internal/compiler/syntax/parser.go
    - internal/compiler/ast/ast.go
    - internal/compiler/check/check.go
    - internal/compiler/check/scalar.go
    - testdata/phase34/draft-after.json

key-decisions:
  - "Keep true and false as identifiers; classify only an unbound final result name in a known result-type context."
  - "Preserve result-tail source spans in the AST only; core schema and semantic IDs stay unchanged."
  - "Publish exactly the seven current refusal codes with independently pinned prose hints and no repairs."

patterns-established:
  - "Contextual refusal: resolve bindings first, then classify only the exact unsupported result-tail shape."
  - "Recovery claims suppress only the exact parser-owned token span and retain near-miss diagnostics."

requirements-completed: [DX-22]

coverage:
  - id: D1
    description: The six parser-owned unsupported forms receive dedicated exact-span refusals while recovery preserves near misses, siblings, and the measured draft corpus.
    requirement: DX-22
    verification:
      - kind: unit
        ref: internal/compiler/syntax/parser_phase34_refusal_test.go#TestPhase34UnsupportedSurfaceRefusals
        status: pass
      - kind: unit
        ref: internal/compiler/syntax/parser_phase34_recovery_test.go#TestPhase34RecoveryCorpus
        status: pass
      - kind: unit
        ref: internal/compiler/syntax/parser_phase34_recovery_test.go#TestPhase34RecoveryControls
        status: pass
    human_judgment: false
  - id: D2
    description: Unbound true/false result tails receive the seventh refusal only after binding lookup, while legal names, ordinary calls, and byte_at retain their checked behavior.
    requirement: DX-22
    verification:
      - kind: unit
        ref: internal/compiler/check/check_phase34_refusal_test.go#TestPhase34UnsupportedSurfaceRefusals
        status: pass
      - kind: unit
        ref: internal/compiler/check/check_phase34_refusal_test.go#TestPhase34BoolSpellingsRemainIdentifiers
        status: pass
      - kind: unit
        ref: internal/compiler/check/check_phase34_refusal_test.go#TestPhase34CallArityBoundaries
        status: pass
    human_judgment: false
  - id: D3
    description: All seven refusal hints are independently pinned and expose no repair edits.
    requirement: DX-22
    verification:
      - kind: unit
        ref: internal/compiler/check/check_phase34_refusal_test.go#TestPhase34RefusalHintsHaveNoRepairs
        status: pass
    human_judgment: false

duration: 35min
completed: 2026-10-13
status: complete
---

# Phase 34 Plan 05: Unsupported Surface Refusals Summary

**Seven unsupported source forms now have dedicated exact-span refusals, with bound `true`/`false` identifiers and admitted call boundaries preserved.**

## Performance

- **Duration:** 35 minutes, from the first captured task command through summary creation.
- **Started:** 2026-10-13T22:54:52Z (first recorded focused test run).
- **Completed:** 2026-10-13T23:29:34Z.
- **Tasks:** 3.
- **Files modified:** 43 plan outputs and evidence files, including 34 focused run records and JUnit reports.

## Accomplishments

- Added the seven-code `surface.Refusals` catalog and parser-owned exact-span refusals for `&&`, `||`, `!`, subtraction, excessive Schway/foreign declaration arity, and an `if` without `else`. Recovery suppresses only the claimed token bytes and preserves strings, comments, spacing, arrows, ownership syntax, siblings, and caller closers.
- Added source-only `LinearBody.ResultSpan` and classified only unbound final `true`/`false` names after result binding lookup. Bound names retain their previous type and initialization behavior; unknown names elsewhere keep their existing diagnostics.
- Kept ordinary arity-two call checks, `byte_at(view, index)`, and the absent scalar helper-call grammar unchanged. Independently pinned hints remain prose with no repairs. Updated only authorized measured rows in `draft-after.json`; the immutable baseline stayed unchanged.

## Task Commits

1. **Task 1 RED fixture:** `fdb82d93` (initial fixture, later corrected after an invalid attempt was recorded).
2. **Task 1 corrected RED and catalog:** `6bc8ddae`.
3. **Task 1 implementation:** `91454f2c`.
4. **Task 2 RED tests:** `1e589460`.
5. **Task 2 implementation:** `ac9c02c1`.
6. **Task 3 implementation and tests:** `4ab41550`.

The plan commit count is measured from `96863718cfeb2972eec95efe028f53a541276c63` to `4ab415501f0477ba61fd4c64332556fb09350359`: 6 commits.

## Validation Evidence

- Task 1 valid RED: `34-05-task1-tdd-red-evidence.json`; actual `TestPhase34UnsupportedSurfaceRefusals` failed at the intended missing logical-and refusal, and `gsd_run check tdd-red-evidence` returned `RED_EVIDENCE_OK`. The corrected implementation GREEN run passed in 1,079.6 ms; the post-commit tracer passed in 194.4 ms.
- Task 2 valid RED: `34-05-task2-tdd-red-evidence.json`; the new forms retained their old generic syntax diagnostics while the existing logical-and tracer passed, and the evidence gate returned `RED_EVIDENCE_OK`. The final focused selector passed in 1,212.1 ms: `TestPhase34UnsupportedSurfaceRefusals`, `TestPhase34RecoveryCorpus`, `TestPhase34RecoveryControls`, and its bounded-delimiter and diagnostic-cap subtests. The named negative controls reached and restored `cascade_count`, `missing_sibling`, and `missing_independent_cause`.
- Task 3 valid RED: `34-05-task3-tdd-red-evidence.json`; all four tests compiled and the bound-name/call controls passed, while unbound Bool tails returned the old unknown-name diagnostics. The evidence gate returned `RED_EVIDENCE_OK`. The final focused selector passed in 636.9 ms: `TestPhase34UnsupportedSurfaceRefusals`, `TestPhase34BoolSpellingsRemainIdentifiers`, `TestPhase34CallArityBoundaries`, and `TestPhase34RefusalHintsHaveNoRepairs`.
- The actual test commands were focused `go test -json` invocations through the standard-library reporter at `/tmp/schway-go-junit.py`, with `GOCACHE=/tmp/ai-lang-verification-gocache`. Records preserve original JSON output, exit codes, durations, semantic assessments, and JUnit reports. The invalid fixture/setup attempts are retained under `34-05-task*-initial-*`; none was accepted as RED evidence.
- No whole-package or project-wide suite, native matrix, hosted dispatch, or receipt binding was run, per the plan.

## TDD Gate Compliance

All three tasks have fresh machine-accepted RED evidence followed by passing GREEN verification, with semantic inspection recorded. Task 3's test and implementation landed together in `4ab41550` rather than in separate `test(...)` and `feat(...)` commits. The RED behavior gap was genuine and the green result is verified, but that commit sequence does not meet the canonical separate RED-commit pattern; it is recorded here rather than represented as compliant.

## Decisions Made

- Keep `true` and `false` available as ordinary parameter, local, and imported function names; the source spelling alone does not make them literals.
- Keep final-result byte spans in the syntax AST for contextual diagnostics without adding a core field or changing stable semantic IDs.
- Keep refusal hints as independently tested prose and avoid bitwise, hex, or dynamic-shift promises owned by later work.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Narrowed missing-else recognition to a complete then block**
- **Found during:** Task 2.
- **Issue:** A broad `if`-without-`else` recognizer could claim malformed or incomplete then-block syntax as the dedicated refusal.
- **Fix:** Required the completed then block and a valid statement boundary before emitting `surface.unsupported_if_without_else`; malformed near misses keep their original parser diagnostics.
- **Files modified:** `internal/compiler/syntax/parser.go`, `internal/compiler/syntax/parser_phase34_refusal_test.go`.
- **Verification:** Task 2 refusal and recovery selectors passed with incomplete-block and spacing controls.
- **Committed in:** `ac9c02c1`.

**Total deviations:** 1 auto-fixed parser-classification bug. Several invalid test-fixture attempts were preserved as failed attempts and corrected before their respective accepted RED gates.
**Impact on plan:** The fix preserves the planned exact refusal boundary; no dependencies, CLI flags, or later-phase language features were added.

## Known Stubs

None found in the added or modified plan files during the stub-pattern scan.

## Self-Check: PASSED

The summary and all nine declared plan files exist. All six measured task commits are ancestors of the current `phase26-validation` HEAD.

## Next Phase Readiness

Plan 34-05 is complete. Phase 34 remains active with Plans 06–08 and phase-level verification still outstanding; this summary does not claim phase completion.
