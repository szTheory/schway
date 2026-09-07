---
phase: 05-native-equivalence-and-adversarial-evidence
plan: 13
subsystem: testing
tags: [go, adversarial-evidence, escape-vocabulary, corevalidate, originvalidate, ownership]

# Dependency graph
requires:
  - phase: 05-native-equivalence-and-adversarial-evidence (05-11)
    provides: "QLT-01 registry's spike-004-coordinated-frontend-summary-lie row, already citing originvalidate.KnownEscape; this plan's shipped constant is the correct target that row already points at"
provides:
  - "testdata/phase5/coordinated_lie.lang + coordinated_lie.core.json — a real adversarial artifact pair naming one module/function identity that asserts two different ownership-transfer claims"
  - "session.EscapeCoordinatedSourceToCoreFalseClaim — the named, gate-visible escape constant (D-05-30)"
  - "session.VerifyCoordinatedLieEscape — the lane proving the gate passes the pair under the named escape"
  - "LaneResult.ExpectedEscapes (qlt01.go, additive) — the field a lane uses to attribute a pass to a declared escape rather than a silent absence of checking"
affects: [05-14]

actuals:
  tokens: 58000
  tasks: 2
  commits: 2

tech-stack:
  added: []
  patterns:
    - "A hand-authored core.Program JSON fixture, decoded and run directly through corevalidate.Validate with no checker or source involved at all — the literal shape of the boundary corevalidate.KnownEscape documents, made executable rather than asserted in prose."
    - "Declared-but-never-detected escape lane (mirrors EscapeCallbackInvocationUnsubjected/session_phase5_sanitize.go): a lane whose PASS is attributed to a named escape via an explicit ExpectedEscapes field, never left as an unexplained absence of findings."

key-files:
  created:
    - testdata/phase5/coordinated_lie.lang
    - testdata/phase5/coordinated_lie.core.json
    - internal/compiler/session/session_phase5_escapes.go
    - internal/compiler/session/session_phase5_escapes_test.go
  modified:
    - internal/compiler/session/qlt01.go

key-decisions:
  - "The false claim is an ownership-transfer disagreement, not a type or schema mismatch: coordinated_lie.lang is an ordinary, checker-accepted move-then-return program (identical shape to testdata/phase2/owned_transfer.lang, chosen deliberately so the automatic Phase5MilestoneCorpus sweep's three-engine agreement differential exercises it safely); coordinated_lie.core.json hand-authors the SAME module/function identity but replaces the first operation with a shared borrow (OpBorrowShared) instead of a move (OpMove) — the returned Buffer is claimed to be merely shared, never fully transferred. corevalidate.Validate accepts this because it never checks return-origin honesty (that lives in originvalidate.ValidatePublished, which is deliberately never invoked against the hand-authored core) — precisely the coordinated-lie boundary corevalidate.KnownEscape documents."
  - "LaneResult (qlt01.go, plan 05-11) gained one additive field, ExpectedEscapes []string, so VerifyCoordinatedLieEscape's pass can be attributed to the named escape rather than reported as an unexplained zero-findings result. This is a small edit outside this plan's declared file list (session_phase5_escapes.go/_test.go, the two testdata files) — see Deviations. All existing LaneResult{} construction sites use keyed literals, so the addition is fully backward compatible and verified not to change qlt01_test.go's own passing behavior."
  - "Folding EscapeCoordinatedSourceToCoreFalseClaim into the shipped session.Phase5ExpectedEscapes() is explicitly NOT done in this plan, per the plan's own instruction — that single-line edit is plan 05-14's job, alongside the reducer/registry control-set update, so the Go control/escape sets and scripts/verify-phase5.sh's own parity test never observe a transiently divergent pair. TestBothPhase5EscapesAreVisible therefore constructs the eventual union set in-test (Phase5ExpectedEscapes() plus this plan's constant) rather than asserting the production function already returns both."

requirements-completed: [QLT-01, INT-02]

coverage:
  - id: D1
    description: "An adversarial matching-but-false source/core artifact pair exists; each half independently passes its own validator (check+originvalidate for the source, corevalidate for the hand-authored core), and the pair's semantic disagreement is asserted explicitly rather than described in prose"
    requirement: QLT-01
    verification:
      - kind: unit
        ref: "internal/compiler/session#TestCoordinatedLieArtifactsBothValidate"
        status: pass
      - kind: integration
        ref: "internal/compiler/session#TestPhase5CorpusThreeEngineAgreement/coordinated_lie.lang (automatic corpus sweep, native compile+execute, three-engine agreement)"
        status: pass
    human_judgment: false
  - id: D2
    description: "escape:coordinated-source-to-core-false-claim is declared, reuses corevalidate.KnownEscape/originvalidate.KnownEscape's wording and machinery verbatim (never redefines them), and never appears in any shipped required-control set"
    requirement: QLT-01
    verification:
      - kind: unit
        ref: "internal/compiler/session#TestCoordinatedLieEscapeIsDeclared"
        status: pass
      - kind: unit
        ref: "internal/compiler/session#TestCoordinatedLieEscapeIsNeverDetected"
        status: pass
    human_judgment: false
  - id: D3
    description: "The gate (check+originvalidate+corevalidate, run in full over the pair) PASSES the adversarial artifacts, and the pass is attributed to the named escape rather than a silent absence of checking; both Phase 5 escapes are jointly visible and neither appears in any control set"
    requirement: INT-02
    verification:
      - kind: unit
        ref: "internal/compiler/session#TestCoordinatedLiePassesTheGateUnderTheNamedEscape"
        status: pass
      - kind: unit
        ref: "internal/compiler/session#TestBothPhase5EscapesAreVisible"
        status: pass
      - kind: unit
        ref: "internal/compiler/session#TestExpectedEscapesAreVisibleNotSolved"
        status: pass
    human_judgment: false
  - id: D4
    description: "No regression: session.go and session_phase5.go remain byte-frozen for this plan; full go build/vet/test(-race) and sh scripts/verify-phase5.sh stay green"
    requirement: QLT-01
    verification:
      - kind: integration
        ref: "go build ./... && go vet ./... && go test ./... && go test -race ./..."
        status: pass
      - kind: integration
        ref: "sh scripts/verify-phase5.sh (full run, exit 0, all lanes pass)"
        status: pass
    human_judgment: true
    rationale: "Whether the chosen false claim (a shared-borrow-vs-move ownership-transfer disagreement) is a sufficiently real semantic lie, and whether the LaneResult.ExpectedEscapes field addition is the correct minimal shape for 05-14 to build on, are documentary/design judgments worth a human's confirmation alongside 05-14's own gate-wiring review."

duration: ~50min
completed: 2026-09-06
status: complete
---

# Phase 5 Plan 13: Coordinated Source-to-Core False Claim, Made Demonstrable Summary

**A hand-authored `coordinated_lie.core.json` claims the identical `phase5.coordinated_lie::relay` function only ever shares its Buffer parameter, while the paired `coordinated_lie.lang` source actually moves and returns it — both independently pass their own validator, and `VerifyCoordinatedLieEscape` proves the gate passes the pair under the newly declared `escape:coordinated-source-to-core-false-claim`.**

## Performance

- **Duration:** ~50 min
- **Started:** 2026-09-06T16:00:00Z (approx, first file read)
- **Completed:** 2026-09-06T16:50:00Z
- **Tasks:** 2 (both `type="auto"`)
- **Files modified:** 5 (4 created, 1 modified)

## Accomplishments

- `testdata/phase5/coordinated_lie.lang` — an ordinary, checker-accepted move-then-return `Buffer` program (`relay`), identical in shape to `testdata/phase2/owned_transfer.lang`. It passes `check` and `originvalidate.ValidatePublished` on its own terms, and is automatically swept into `Phase5MilestoneCorpus`'s three-engine agreement differential (native compile + execute), which passes.
- `testdata/phase5/coordinated_lie.core.json` — a HAND-AUTHORED core artifact naming the exact same `module_id`/function identity, whose first operation is `borrow_shared` (not `move`): it claims the returned Buffer is only ever a shared alias, never a full ownership transfer. It independently passes `corevalidate.Validate` on its own terms.
- `TestCoordinatedLieArtifactsBothValidate` asserts both validations pass AND explicitly compares the two artifacts' first-operation kinds, requiring them to differ (`OpMove` vs `OpBorrowShared`) despite sharing one module identity — the semantic disagreement is asserted, not described.
- `session.EscapeCoordinatedSourceToCoreFalseClaim = "escape:coordinated-source-to-core-false-claim"` is declared in a new file, `session_phase5_escapes.go`, with a doc comment reusing `corevalidate.KnownEscape`'s and `originvalidate.KnownEscape`'s wording verbatim.
- `session.VerifyCoordinatedLieEscape` runs check+originvalidate over the source and corevalidate over the hand-authored core, confirms no control fires on either half, and returns a `LaneResult` whose `ExpectedEscapes` names the constant — the pass is attributed to a declared residual, never left as a silent absence of findings.
- `LaneResult` (`qlt01.go`, plan 05-11) gained one additive `ExpectedEscapes []string` field to carry that attribution.
- Four new tests plus a re-run of the existing `TestExpectedEscapesAreVisibleNotSolved` all pass; `escape:coordinated-source-to-core-false-claim` never appears in `session.AllShippedControlIDs()` (`grep -c 'RequiredControls'` over its own occurrences returns `0`).
- `session.go` and `session_phase5.go` are untouched (`git diff --stat` confirms no changes) — folding this constant into `Phase5ExpectedEscapes()` is explicitly deferred to plan 05-14.

## Task Commits

Each task was committed atomically:

1. **Task 1: Construct the adversarial matching-but-false source/core artifact pair** - `675395e` (feat)
2. **Task 2: Declare the escape, assert the gate passes the pair under it, and assert it is never claimed as covered** - `21a4269` (feat)

**Plan metadata:** committed as part of this SUMMARY's own commit (see below).

## Files Created/Modified

- `testdata/phase5/coordinated_lie.lang` — the adversarial SOURCE half of the pair (checker-accepted, natively executable)
- `testdata/phase5/coordinated_lie.core.json` — the adversarial hand-authored CORE half of the pair (corevalidate-accepted on its own terms)
- `internal/compiler/session/session_phase5_escapes.go` — `EscapeCoordinatedSourceToCoreFalseClaim`, `LaneCoordinatedLieEscape`, `VerifyCoordinatedLieEscape`
- `internal/compiler/session/session_phase5_escapes_test.go` — `TestCoordinatedLieArtifactsBothValidate`, `TestCoordinatedLieEscapeIsDeclared`, `TestCoordinatedLieEscapeIsNeverDetected`, `TestCoordinatedLiePassesTheGateUnderTheNamedEscape`, `TestBothPhase5EscapesAreVisible`
- `internal/compiler/session/qlt01.go` — additive `ExpectedEscapes []string` field on `LaneResult`

## Decisions Made

See `key-decisions` in frontmatter: the ownership-transfer (move-vs-borrow) shape of the false claim; the additive `LaneResult.ExpectedEscapes` field; and the explicit deferral of `Phase5ExpectedEscapes()`'s update to plan 05-14.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking issue] Added an additive `ExpectedEscapes []string` field to `LaneResult` in `qlt01.go`**

- **Found during:** Task 2, designing `VerifyCoordinatedLieEscape`'s return shape
- **Issue:** The plan's acceptance criteria require `VerifyCoordinatedLieEscape` to attribute its pass "in the result's ExpectedEscapes" (Task 2's own action text), but the existing `LaneResult` type (`qlt01.go`, shipped by plan 05-11) has no such field — it was designed narrowly for the QLT-01 registry audit's own `ID`/`Status`/`Controls`/`RecomputedWork`/`Fired` shape. Task 2's declared `<files>` list only names `session_phase5_escapes.go`/`session_phase5_escapes_test.go`, not `qlt01.go`.
- **Fix:** Added one additive field, `ExpectedEscapes []string`, to `LaneResult`. Verified every existing `LaneResult{}` construction site (`qlt01.go` lines 243/280) uses either an empty literal or a fully keyed literal, so the addition is 100% backward compatible — no positional-literal call site could break. `QLT01LaneFromRows` leaves the new field nil (the registry audit attributes no escape of its own), confirmed by re-running the full `qlt01_test.go` suite green.
- **Files modified:** `internal/compiler/session/qlt01.go`
- **Verification:** `sh scripts/assert-go-tests.sh ./internal/compiler/session/... TestQLT01RegistryParses TestQLT01RegistryComplete TestQLT01LaneCountsWork` (all pass, unaffected by the new field); full `go test ./...`, `go test -race ./...`, `go vet ./...`, and `sh scripts/verify-phase5.sh` (exit 0) all green after the change.
- **Committed in:** `21a4269` (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking issue — Rule 3). The only file touched outside the plan's declared list is a five-line additive struct-field change with zero behavioral impact on any existing caller.
**Impact on plan:** Necessary for `VerifyCoordinatedLieEscape` to satisfy its own declared acceptance criteria (attributing the pass to a named escape). No scope creep — the field is unused by any pre-existing code path except by explicit, verified omission (`nil`).

## Issues Encountered

None beyond the deviation above.

## User Setup Required

None — no external service configuration required.

## Known Stubs

None. `coordinated_lie.lang`/`coordinated_lie.core.json` are real, independently-validating artifacts (not placeholders), and `VerifyCoordinatedLieEscape` performs genuine validation calls against both — no hardcoded pass, no mocked validator.

## Next Phase Readiness

- `escape:coordinated-source-to-core-false-claim` is declared, demonstrated, and gate-visible via `VerifyCoordinatedLieEscape`. `session.go` and `session_phase5.go` remain byte-frozen for this plan.
- **Plan 05-14 owns:** (1) adding `EscapeCoordinatedSourceToCoreFalseClaim` to `Phase5ExpectedEscapes()` in `session_phase5.go`, (2) wiring `VerifyCoordinatedLieEscape`'s lane into `VerifyPhase5ControlsAndWork`'s overall `Result.Lanes`/`ExpectedEscapes`, and (3) updating `scripts/verify-phase5.sh`'s own parity block alongside the reducer and QLT-01 registry controls — all as one coordinated update so the Go control/escape sets and the shell gate script never observe a transiently divergent pair (matching this plan's explicit instruction).
- The QLT-01 registry's `spike-004-coordinated-frontend-summary-lie` row (05-11) already cites the correct shipped `originvalidate.KnownEscape` constant and needs no change from this plan, as 05-11's own summary anticipated.

## Self-Check: PASSED

- `testdata/phase5/coordinated_lie.lang` — FOUND
- `testdata/phase5/coordinated_lie.core.json` — FOUND
- `internal/compiler/session/session_phase5_escapes.go` — FOUND
- `internal/compiler/session/session_phase5_escapes_test.go` — FOUND
- `.planning/phases/05-native-equivalence-and-adversarial-evidence/05-13-SUMMARY.md` — FOUND
- Commits `675395e`, `21a4269` — both present in `git log --oneline --all`
- All plan-level `<verification>` commands re-run and green: `sh scripts/assert-go-tests.sh ./internal/compiler/session/... TestCoordinatedLieArtifactsBothValidate TestCoordinatedLieEscapeIsDeclared TestCoordinatedLieEscapeIsNeverDetected TestCoordinatedLiePassesTheGateUnderTheNamedEscape TestBothPhase5EscapesAreVisible TestExpectedEscapesAreVisibleNotSolved` (pass), `go test ./... && go vet ./...` (green), `go test -race ./...` (green), `sh scripts/verify-phase5.sh` (exit 0)

---
*Phase: 05-native-equivalence-and-adversarial-evidence*
*Completed: 2026-09-06*
