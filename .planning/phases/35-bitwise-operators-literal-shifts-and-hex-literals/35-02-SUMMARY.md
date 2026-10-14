---
phase: 35-bitwise-operators-literal-shifts-and-hex-literals
plan: "02"
subsystem: syntax
tags: [parser, formatter, numeric-literals, diagnostics, go]
requires:
  - phase: 35-01
    provides: frozen numeric, index, and pre-change emitter baselines
provides:
  - Grouped scalar bitwise and literal-shift grammar with stable formatting
  - A shared 64-entry scalar expression recursion cap with sibling recovery
  - Legacy uppercase hexadecimal admission and exact malformed-token ownership
affects: [phase-35-plans-03-through-08, scalar-parser, numeric-policy]
actuals:
  tokens: 23565.75
  tasks: 3
  commits: 3
commits: 3
plan_head_before: a6e04fee1899220043998e3f4ff4b83426d7a7d1
plan_head_after: f902893942284fcf92dfc6c7b225b2fff853b137
tech-stack:
  added: []
  patterns:
    - Exact-span lexer ownership for malformed numeric tokens
    - Source-hash-bound compatibility amendments preserve immutable before-state facts
key-files:
  created:
    - internal/compiler/syntax/parser_phase34_recovery_test.go
    - internal/compiler/session/session_phase34_compat_test.go
    - internal/compiler/syntax/numeric_phase35_test.go
    - internal/compiler/check/check_phase35_numeric_test.go
    - testdata/phase35/recovery-after.json
    - .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-02-COMPATIBILITY-NOTE.md
  modified:
    - internal/compiler/syntax/parser.go
    - internal/compiler/syntax/format.go
    - internal/compiler/surface/refusals.go
    - internal/compiler/syntax/parser_phase35_test.go
    - internal/compiler/check/check.go
    - internal/compiler/session/session_phase35_prechange_test.go
    - testdata/phase35/numeric-delta.json
key-decisions:
  - "Only suppress parser expectations for a malformed numeric TokenUnknown at the lexer's exact full-token span."
  - "Accept the uppercase 0X prefix on the legacy route while retaining scalar conversion and 010/0B10 policy."
  - "Keep Phase 34 and Phase 35 historical fixtures immutable; record narrow, literal live after-state amendments."
patterns-established:
  - "The top-level scalar expression starts at depth zero; groups and recursive byte_at indices share the 64-entry cap."
  - "Compatibility comparisons validate a literal amendment against the frozen source hash and before projection before applying its explicit after projection."
requirements-completed: [BIT-01, BIT-02, BIT-03]
coverage:
  - id: D1
    description: Grouped scalar bitwise operators and literal shifts parse and format to stable semantic trees.
    requirement: BIT-01
    verification:
      - kind: unit
        ref: internal/compiler/syntax/parser_phase35_test.go#TestPhase35ScalarGrammar
        status: pass
      - kind: unit
        ref: internal/compiler/syntax/parser_phase35_test.go#TestPhase35FormatterRoundTrip
        status: pass
      - kind: unit
        ref: internal/compiler/syntax/syntax_test.go#TestFormatLetBinaryExpressions
        status: pass
    human_judgment: false
  - id: D2
    description: Scalar recursion is capped at depth 64 with delimiter-aware recovery and independent sibling diagnostics.
    requirement: BIT-02
    verification:
      - kind: unit
        ref: internal/compiler/syntax/parser_phase35_test.go#TestPhase35ExpressionDepth
        status: pass
      - kind: unit
        ref: internal/compiler/syntax/parser_phase34_recovery_test.go#TestPhase34RecoveryCorpus
        status: pass
      - kind: unit
        ref: internal/compiler/syntax/parser_phase34_recovery_test.go#TestPhase34RecoveryControls
        status: pass
      - kind: unit
        ref: internal/compiler/syntax/parser_phase35_test.go#TestPhase35RecoveryDeltaControls
        status: pass
      - kind: unit
        ref: internal/compiler/session/session_phase34_compat_test.go#TestPhase34CompatibilityDeltaLedger
        status: pass
    human_judgment: false
  - id: D3
    description: Legacy uppercase hexadecimal, route-local overflow refusals, and exact full-span malformed numeric ownership are pinned.
    requirement: BIT-03
    verification:
      - kind: unit
        ref: internal/compiler/syntax/numeric_phase35_test.go#TestPhase35NumericOwnership
        status: pass
      - kind: unit
        ref: internal/compiler/check/check_phase35_numeric_test.go#TestPhase35NumericRoutePolicy
        status: pass
      - kind: unit
        ref: internal/compiler/check/check_phase35_baseline_test.go#TestPhase35NumericPolicyBaseline
        status: pass
      - kind: unit
        ref: internal/compiler/session/session_phase35_prechange_test.go#TestPhase35PreChangeEmitterCorpus
        status: pass
      - kind: unit
        ref: internal/compiler/session/session_phase35_prechange_test.go#TestPhase35PreChangeNumericDeltaControls
        status: pass
    human_judgment: false
duration: 22min
completed: 2026-10-14
status: complete
---

# Phase 35 Plan 02: Grouped Operators, Bounded Parsing, and Hex Literals Summary

**Grouped scalar operators, a shared 64-entry recursion bound, and legacy uppercase hex now have focused parser, recovery, formatting, and compatibility evidence.**

## Performance

- **Duration:** 22 minutes
- **Started:** 2026-10-14T14:35:00Z
- **Completed:** 2026-10-14T14:57:16Z
- **Tasks:** 3
- **Files modified:** 13

## Accomplishments

- Added grouped scalar `^`, `&`, `|`, `<<`, and `>>` parsing with formatter round-trip evidence and unchanged generic angle handling.
- Bounded recursive scalar expressions at 64 entries across groups and `byte_at` indices, with exact refusals and preserved recovery siblings.
- Added legacy `0X` support while preserving the existing scalar/legacy numeric policy, exact malformed candidate spans, and immutable historical corpus facts.

## Task Commits

1. **Task 1: Parse and format the locked grouped operator witness end to end** — `7958eebf` (`feat`)
2. **Task 2: Bound every scalar recursive entry and preserve recovery siblings** — `b9c9af38` (`feat`)
3. **Task 3: Complete hex admission with exact lexical ownership** — `f9028939` (`feat`)

The task commits are the three measured commits in the plan ledger. The summary/state metadata commit is recorded separately after closeout.

## Files Created/Modified

- `internal/compiler/syntax/parser.go` — grouped operators, shared recursion bound, and exact malformed-token expectation suppression.
- `internal/compiler/syntax/format.go` — stable rendering and reparsing for grouped scalar expressions.
- `internal/compiler/surface/refusals.go` — named depth and shift-amount refusal messages.
- `internal/compiler/check/check.go` — legacy `0X` parsing without changing route-local conversion policy.
- `internal/compiler/syntax/parser_phase35_test.go`, `internal/compiler/syntax/numeric_phase35_test.go`, and `internal/compiler/check/check_phase35_numeric_test.go` — grammar, depth, exact-span, route, and negative-control evidence.
- `internal/compiler/syntax/parser_phase34_recovery_test.go` and `internal/compiler/session/session_phase34_compat_test.go` — frozen recovery comparison and narrow current Phase 35 projection.
- `internal/compiler/session/session_phase35_prechange_test.go` — source-hash-bound Phase 19 diagnostic ownership amendment and mutation controls.
- `testdata/phase35/numeric-delta.json` and `testdata/phase35/recovery-after.json` — dated resource and compatibility projections; historical fixture files remain unchanged.
- `.planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-02-COMPATIBILITY-NOTE.md` — recorded the parent-execution compatibility reconciliation.

## Decisions Made

- The existing lexer already owns complete numeric candidates, so it required no code change. Parser expectations are suppressed only for `syntax.expected_*` at an exact malformed `TokenUnknown` span.
- Uppercase `0X` is admitted only by the legacy U64 conversion branch; scalar conversion remains as before, and `010`/`0B10` retain their route-specific meanings.
- Historical Phase 34 and Phase 35 fixtures remain unchanged. Current behavior is compared through exact, source-bound amendments with independently checked before and after projections.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added a narrow Phase 34 recovery compatibility projection**
- **Found during:** Task 2
- **Issue:** The required recovery corpus observed that newly admitted operators remove two historical parser diagnostics from `crc32-reconstructed`; the historical Phase 34 draft-after expected state had an unexpected-byte diagnostic at 211:212 and an expected-brace diagnostic at 184:185.
- **Fix:** Preserved all frozen Phase 34 source, before/after fixtures, and ledger facts; added a literal Phase 35 after-state with zero parser diagnostics and the `crc32` sibling, plus exact missing/wrong projection controls. No CRC numeric result or whole-program admission is implied.
- **Files modified:** `internal/compiler/syntax/parser_phase34_recovery_test.go`, `internal/compiler/session/session_phase34_compat_test.go`, `testdata/phase35/recovery-after.json`, `testdata/phase35/numeric-delta.json`, and the compatibility note.
- **Verification:** `TestPhase34RecoveryCorpus`, `TestPhase34RecoveryControls`, `TestPhase35RecoveryDeltaControls`, and `TestPhase34CompatibilityDeltaLedger` pass.
- **Committed in:** `b9c9af38`

**2. [Rule 3 - Blocking] Bound the malformed-diagnostic refinement to one historical Phase 19 source row**
- **Found during:** Task 3
- **Issue:** Exact-span suppression intentionally changes the frozen live projection for `testdata/phase19/literal_malformed.schway`, while its pre-change emitter corpus records two parser diagnostics and three refusal reasons.
- **Fix:** Kept the frozen corpus and source bytes unchanged. Added a source-hash-bound literal before/after amendment to `numeric-delta.json`; the session control validates the old projection, applies only the declared after-state, and checks missing/wrong amendment and missing-row mutations. This extension followed the existing delegated execution scope; no user approval is claimed.
- **Files modified:** `internal/compiler/syntax/parser.go`, `internal/compiler/session/session_phase35_prechange_test.go`, `internal/compiler/syntax/numeric_phase35_test.go`, `internal/compiler/check/check_phase35_numeric_test.go`, and `testdata/phase35/numeric-delta.json`.
- **Verification:** `TestPhase35NumericOwnership`, `TestPhase35NumericRoutePolicy`, `TestPhase35PreChangeEmitterCorpus`, `TestPhase35PreChangeCorpusControls`, and `TestPhase35PreChangeNumericDeltaControls` pass.
- **Committed in:** `f9028939`

**Total deviations:** 2 Rule 3 compatibility-boundary amendments. Both preserve immutable historical data and add only exact current projections.

## Issues Encountered

- A post-change frozen-corpus check correctly exposed the single malformed-literal row affected by diagnostic de-duplication. The narrow amendment preserves the old row as historical evidence and keeps all unrelated comparisons active.
- The required selectors initially found a parser expectation duplicated at the malformed numeric candidate span. Exact token-span ownership removed that duplicate while controls confirm unrelated unknown tokens, closers, sibling diagnostics, and diagnostic caps remain intact.

## User Setup Required

None — no external service configuration is required.

## Next Phase Readiness

Plan 35-02 is complete. Its grouped grammar, depth boundary, and hex ownership contracts are ready for downstream Phase 35 plans; no full package suite, native lane, hosted receipt, or later plan was run as part of this focused plan.

## Self-Check: PASSED

- The summary and all 13 plan files exist.
- Task commits `7958eebf`, `b9c9af38`, and `f9028939` are ancestors of the current HEAD.
- The persisted plan ledger measures three task commits from `a6e04fee1899220043998e3f4ff4b83426d7a7d1` through `f902893942284fcf92dfc6c7b225b2fff853b137`.

---
*Phase: 35-bitwise-operators-literal-shifts-and-hex-literals*
*Completed: 2026-10-14*
