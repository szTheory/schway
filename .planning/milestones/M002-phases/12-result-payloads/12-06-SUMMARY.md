---
phase: 12-result-payloads
plan: 06
subsystem: compiler-check
tags: [go, compiler, diagnostics, payload-types, cr-01]

requires:
  - phase: 12-result-payloads
    provides: "payload-carrying alternatives as a general data feature (D-12-12), the four sibling payload refusals, and 12-VERIFICATION.md's FAILED truth 2 naming CR-01 as the open gap"
provides:
  - "check.duplicate_payload_type: a fail-closed declaration-time refusal for a data declaration whose two alternatives share a non-empty payload type"
  - "A committed fixture (payload_duplicate_payload_type.lang) and a direct-code-assertion control that fails if the refusal is reverted"
  - "WR-01's construction-side-aware message on check.binder_on_nullary_alternative, with the code string unchanged"
affects: ["12-07 (core-layer shared ambiguity-detecting resolver, the engine-layer half of CR-01)", "12-08 (D-12-44 decision entry)"]

actuals:
  tokens: 62000
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Declaration-time refusal added to the SAME program.Data walk region as check.resource_payload_refused, before the existing diagnostics early-return"
    - "Diagnostic-message precision fix via a conditional branch on which arm side supplied the binder, keeping the diagnostic code stable"

key-files:
  created:
    - testdata/phase12/payload_duplicate_payload_type.lang
  modified:
    - internal/compiler/check/check.go
    - internal/compiler/check/check_payload_test.go

key-decisions:
  - "D-12-44 (CR-01 disposition = FIX): a fail-closed check.* diagnostic, not a re-derived AlternativeName IR field -- the IR-schema route was rejected on corpus-byte-stability grounds; the narrowed engine-layer resolver collapse is deferred to plan 12-07."
  - "WR-01 took the review's SECOND suggested fix shape (conditional message), not a new diagnostic code, so no existing fixture's expected code moves."

requirements-completed: [RES-02, RES-03]

coverage:
  - id: D1
    description: "A data declaration whose alternatives collide on payload type is refused at check time with check.duplicate_payload_type, spanned at the second (offending) alternative."
    requirement: "RES-02"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_payload_test.go#TestDuplicatePayloadTypeRefused/colliding_payload_types_refused"
        status: pass
      - kind: integration
        ref: "go run ./cmd/lang --json check testdata/phase12/payload_duplicate_payload_type.lang"
        status: pass
    human_judgment: false
  - id: D2
    description: "Nullary repetition and cross-data-type payload-type reuse both stay accepted (the refusal is per-declaration and skips empty PayloadType)."
    requirement: "RES-02"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_payload_test.go#TestDuplicatePayloadTypeRefused/three_nullary_alternatives_on_one_data_type_stay_clean"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_payload_test.go#TestDuplicatePayloadTypeRefused/one_Buffer_payload_in_each_of_two_different_data_types_stays_clean"
        status: pass
    human_judgment: false
  - id: D3
    description: "A control exists that goes red when the refusal is removed, demonstrated once during execution."
    requirement: "RES-02"
    verification:
      - kind: manual_procedural
        ref: "Manual revert-and-observe transcript recorded below (not committed)"
        status: pass
    human_judgment: true
    rationale: "The red-then-green transition is an execution-time procedural demonstration, not a re-runnable committed test by construction (the demonstration itself requires temporarily reverting production code) -- verified manually and recorded verbatim."
  - id: D4
    description: "The pre-Phase-12 corpus replays byte-identically with the refusal in place; no golden was re-pinned."
    requirement: "RES-03"
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_payload_replay_test.go#TestPayloadCorpusCharacterizationReplay"
        status: pass
    human_judgment: false
  - id: D5
    description: "WR-01's message names the construction side without introducing a new diagnostic code; the pattern-side case keeps its original wording."
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_payload_test.go#TestBinderOnNullaryAlternativeNamesConstructionSide"
        status: pass
      - kind: integration
        ref: "go run ./cmd/lang --json check testdata/phase12/payload_binder_on_nullary.lang"
        status: pass
    human_judgment: false

duration: 35min
completed: 2026-09-13
status: complete
---

# Phase 12 Plan 06: CR-01 Source-Layer Closure + WR-01 Summary

**`check.duplicate_payload_type` refuses a `data` declaration whose alternatives collide on payload type; a direct-code-assertion control proves the refusal is load-bearing; WR-01's nullary-binder message now names the construction side.**

## Performance

- **Duration:** ~35 min
- **Tasks:** 3
- **Files modified:** 2 (check.go, check_payload_test.go) + 1 created (fixture)

## Accomplishments

- Added `check.duplicate_payload_type`, a declaration-time refusal in `check.Program`'s existing `program.Data` walk, positioned in the same region as `check.resource_payload_refused` and before its early return. Tracks payload types seen per declaration; skips nullary alternatives; the restriction is within one data type only.
- Committed `testdata/phase12/payload_duplicate_payload_type.lang`, cloning `payload_tracer.lang`'s shape with both alternatives (`Ok`, `Failed`) carrying payload type `Buffer`.
- Added `TestDuplicatePayloadTypeRefused` with four subtests: the colliding fixture refused, the tracer fixture clean, three nullary alternatives on one data type clean, and one `Buffer` payload in each of two different data types clean. Appended one additive row to `TestPayloadPatternRefusals`.
- Manually demonstrated the red-then-green transition (see below).
- Measured `TestPayloadCorpusCharacterizationReplay` green and byte-identical across all 97 pre-Phase-12 corpus subtests (62 PASS, 35 SKIP) with the refusal in place.
- Fixed WR-01: `analyzePayloadArm`'s `check.binder_on_nullary_alternative` message now names the construction side when `arm.Binder` is empty and `arm.ConstructBinder` is non-empty, naming the offending binder identifier and both the pattern and value alternatives. The pattern-side case keeps its exact original wording. The diagnostic code is unchanged.

## Task Commits

1. **Task 1: End-to-end colliding payload-type declaration refused** - `698b7bd` (feat)
2. **Task 2: Control that fails if the refusal is reverted, plus measured corpus neutrality** - `6fe8283` (test)
3. **Task 3: WR-01 message-precision fix** - `149612c` (fix)

_Note: Task 2's `<verify>` red-then-green demonstration required temporarily commenting out Task 1's walk. That revert was never committed — the demonstration ran against a scratch copy restored to the committed state immediately after, verified via `git diff --stat` showing zero change to `check.go` from that demonstration._

## Files Created/Modified

- `internal/compiler/check/check.go` - Added the `check.duplicate_payload_type` declaration-time walk (Task 1) and WR-01's conditional message (Task 3)
- `internal/compiler/check/check_payload_test.go` - Added `TestDuplicatePayloadTypeRefused`, one additive `TestPayloadPatternRefusals` row, and `TestBinderOnNullaryAlternativeNamesConstructionSide`
- `testdata/phase12/payload_duplicate_payload_type.lang` - New fixture: two alternatives (`Ok`, `Failed`) both carrying payload type `Buffer`

## Observed Red-Then-Green Demonstration (Task 2)

Procedure: temporarily replaced `for _, declaration := range program.Data {` with `for _, declaration := range []ast.DataDecl(nil) {` (disabling the new walk without deleting it, to keep the file syntactically valid for the demonstration), ran the control, then restored the original line from a pre-edit backup copy and re-ran.

**RED** (walk disabled):
```
=== RUN   TestDuplicatePayloadTypeRefused/colliding_payload_types_refused
    check_payload_test.go:228: expected at least one diagnostic, got none
--- FAIL: TestDuplicatePayloadTypeRefused (0.00s)
    --- FAIL: TestDuplicatePayloadTypeRefused/colliding_payload_types_refused (0.00s)
    --- PASS: TestDuplicatePayloadTypeRefused/tracer_fixture_stays_clean_(distinct_payload_types) (0.00s)
    --- PASS: TestDuplicatePayloadTypeRefused/three_nullary_alternatives_on_one_data_type_stay_clean (0.00s)
    --- PASS: TestDuplicatePayloadTypeRefused/one_Buffer_payload_in_each_of_two_different_data_types_stays_clean (0.00s)
FAIL
```

**GREEN** (walk restored):
```
--- PASS: TestDuplicatePayloadTypeRefused (0.00s)
    --- PASS: TestDuplicatePayloadTypeRefused/colliding_payload_types_refused (0.00s)
    --- PASS: TestDuplicatePayloadTypeRefused/tracer_fixture_stays_clean_(distinct_payload_types) (0.00s)
    --- PASS: TestDuplicatePayloadTypeRefused/three_nullary_alternatives_on_one_data_type_stay_clean (0.00s)
    --- PASS: TestDuplicatePayloadTypeRefused/one_Buffer_payload_in_each_of_two_different_data_types_stays_clean (0.00s)
PASS
```

## Measured Corpus Replay (Task 2)

`go test ./internal/compiler/session/... -run 'TestPayloadCorpusCharacterizationReplay$' -v -count=1` exits 0. Fixture count: **97 subtests total (62 PASS, 35 SKIP)** across all 10 pre-Phase-12 corpus directories (`phase1, phase2, phase3, phase4, phase5, phase6, phase07, phase08, phase10, phase11`). No fixture's execution document moved; no corpus fixture starts failing; no golden was re-pinned. `internal/compiler/session/session_payload_replay_test.go` is byte-unchanged (`git diff --stat` does not name it).

## Diagnostic Message Text Shipped

- `check.duplicate_payload_type` (on `payload_duplicate_payload_type.lang`):
  > `alternative "Failed" shares payload type "Buffer" with alternative "Ok": a construct/destructure operation cannot be resolved to a declaring alternative unambiguously while two alternatives of one data type collide on payload type (D-12-44, see PHASE-12-DEBT.md)`

- `check.binder_on_nullary_alternative`, pattern-side branch (unchanged, verified on `payload_binder_on_nullary.lang`):
  > `binder present on an alternative declared with no payload`

- `check.binder_on_nullary_alternative`, construction-side branch (WR-01, new — verified via `TestBinderOnNullaryAlternativeNamesConstructionSide`'s `Empty => Ok(w)` shape):
  > `construction binder "w" names a payload source for alternative "Ok", but this arm's pattern alternative "Empty" declares no payload to bind it from`

## Decisions Made

- **D-12-44 confirmed as recorded in the plan's `<cr01_disposition>`**: CR-01's disposition is FIX via a fail-closed `check.*` diagnostic (this plan), with the narrowed engine-layer resolver collapse deferred to plan 12-07 as a backstop rather than the primary closure. No new decision text needed beyond what the plan already carries; entry authored in plan 12-08 per the plan's own note.
- **WR-01 took the conditional-message fix shape** (review's second suggestion) rather than a new diagnostic code, preserving `check.binder_on_nullary_alternative` as the sole code for this refusal.

## Deviations from Plan

None — plan executed exactly as written. All acceptance criteria and verification commands ran as specified; no auto-fixes, no scope changes, no architectural questions arose.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- CR-01's source-layer half is closed: `check.duplicate_payload_type` refuses the colliding shape before it reaches `cgen` or `interp`.
- Plan 12-07 can proceed: the engine-layer shared ambiguity-detecting resolver becomes an unreachable-by-construction backstop (this plan's refusal already prevents the hazardous shape from crossing the parser), consistent with the disposition's stated ordering.
- `go test ./...` is green across all 25 packages; `git diff --stat` since the last commit before this plan names exactly `internal/compiler/check/check.go`, `internal/compiler/check/check_payload_test.go`, and the new fixture — no engine file, no golden, no debt register touched.

---
*Phase: 12-result-payloads*
*Completed: 2026-09-13*

## Self-Check: PASSED

- FOUND: `.planning/phases/12-result-payloads/12-06-SUMMARY.md`
- FOUND: `testdata/phase12/payload_duplicate_payload_type.lang`
- FOUND: commit `698b7bd` (Task 1)
- FOUND: commit `6fe8283` (Task 2)
- FOUND: commit `149612c` (Task 3)
- Re-ran plan-level `<verification>`: `go build ./... && go vet ./internal/compiler/check/...` clean; `go test ./internal/compiler/check/... -count=1` green; `go test ./internal/compiler/session/... -run 'TestPayloadCorpusCharacterizationReplay' -v -count=1` green and byte-identical; `go test ./...` green across all 25 packages; `git diff --stat` since the pre-plan commit names only `internal/compiler/check/check.go`, `internal/compiler/check/check_payload_test.go`, and the new fixture.
