---
phase: 26-checked-scalar-sum
plan: 03
subsystem: compiler-analysis
tags: [cfg, fixed-point, scalar, ownership, provenance, ci]
requires:
  - phase: 26-01
    provides: scalar loop admission foundation and the `sum_to_n` witness
provides:
  - Deterministic scalar analysis with fail-closed transfer budgets in checker and peers
  - Independent core and origin peer checks for scalar-copy loop admission and forged edges
  - Category-attributed checker refusals for owner, resource, loan, loan-derived provenance, and unsupported events
affects: [26-04, compiler-checker, corevalidate, originvalidate]
actuals:
  tokens: 8360
  tasks: 2
  commits: 16
  plan_head_before: 70c911f2fef043a89d19606f640a65f79faca428
tech-stack:
  added: []
  patterns: [finite monotone scalar worklists, test-only analysis budgets, independent typed-core peer validation]
key-files:
  created: [.planning/phases/26-checked-scalar-sum/26-03-red-evidence.json]
  modified: [.github/workflows/ci.yml, internal/compiler/check/check.go, internal/compiler/check/scalar.go, internal/compiler/check/check_test.go, internal/compiler/corevalidate/corevalidate.go, internal/compiler/corevalidate/corevalidate_cycle_peer_test.go, internal/compiler/originvalidate/originvalidate.go, internal/compiler/originvalidate/originvalidate_test.go]
key-decisions:
  - "Only initialized U64/Bool copies are admitted as cyclic scalar transfers; legacy ownership copies remain outside scalar analysis."
  - "Each validator keeps its own scalar transfer and cycle checks; pathoracle remains acyclic-only."
patterns-established:
  - "Analysis-budget overrides are package-local and restored by tests, while production retains a deterministic fixed cap."
  - "Back-edge diagnostics retain `check.cfg_back_edge`, a function source span, and a machine-readable authority/event cause."
requirements-completed: [FLOW-01, FLOW-02]
coverage:
  - id: D1
    description: "Checker, corevalidate, and originvalidate admit the checked scalar sum under reordered block listings and fail closed at test-forced analysis budgets."
    requirement: FLOW-01
    verification:
      - kind: unit
        ref: "GitHub Actions run 37085604802: TestPhase26FixedPoint, TestPhase26AnalysisExhaustion, TestPhase26PeerIndependence"
        status: pass
    human_judgment: false
  - id: D2
    description: "Scalar-copy loop cycles are admitted while forged source types and CFG edges are rejected independently; authority/event categories receive stable checker causes."
    requirement: FLOW-02
    verification:
      - kind: unit
        ref: "GitHub Actions run 37085604802: TestPhase26BackEdgeAuthority, TestPhase26BackEdgeCauseCategories, TestPhase26CFGBackEdgeCategories, TestPhase26ScalarCopyCycleBoundary"
        status: pass
      - kind: unit
        ref: "GitHub Actions run 37085604802: TestCompositionCycleGuardFailsClosed and Phase 26 session/core/interp/cgen/native selectors"
        status: pass
    human_judgment: false
duration: 31min
completed: 2026-10-03
status: complete
---

# Phase 26 Plan 03: Checked Scalar Sum Summary

**Independent finite scalar analysis now admits U64/Bool copy events in cycles while retaining fail-closed authority and event refusals.**

## Performance

- **Duration:** 31 min
- **Started:** 2026-10-03T00:49:47Z
- **Completed:** 2026-10-03T01:20:00Z
- **Tasks:** 2
- **Files modified:** 9 plan-delivery files

## Accomplishments

- Added deterministic scalar fixed-point budgets and test seams in checker, corevalidate, and originvalidate; the production cap admits `sum_to_n`, while forced exhaustion returns named failures.
- Kept scalar U64/Bool `OpCopy` transfers separate from ownership copies and validated scalar-copy cycles independently in all three validators.
- Added forged branch-edge and source-type controls, category-specific `check.cfg_back_edge` causes with stable function spans, and focused CI selectors for the new controls.
- Preserved pathoracle's acyclic-only cycle refusal as an independent regression.

## Task Commits

1. **Task 1: Make scalar fixed-point convergence and exhaustion observable** - `6b9f154` (tests), `61a001b`, `8ccc628`, `4abb770`, `f1415dd` (implementation and fixes).
2. **Task 2: Refuse loop-carried authority while admitting scalar-copy events** - `6f6e596`, `147a71f`, `7865ff5`, `0bb1c3d`, `9bc8576`, `39c48aa`, `997f5b5`, `f112a21`, `8dbb4fd`, `df2a2bc`, `d653312` (implementation, controls, and fixes).

**Plan metadata:** captured in the GSD closeout commit after state and roadmap updates.

## Files Created/Modified

- `internal/compiler/check/scalar.go` - finite scalar transfer worklist, forced budget seam, scalar copy transfer, and refusal categories.
- `internal/compiler/check/check.go` - category-attributed cyclic CFG diagnostics with source spans.
- `internal/compiler/check/check_test.go` - convergence, exhaustion, peer independence, forged edge/type, copy-cycle, and refusal-category controls.
- `internal/compiler/corevalidate/corevalidate.go` and `internal/compiler/originvalidate/originvalidate.go` - independent scalar copy transfer validation and test-only budget controls.
- `internal/compiler/corevalidate/corevalidate_cycle_peer_test.go` and `internal/compiler/originvalidate/originvalidate_test.go` - forced-exhaustion controls.
- `.github/workflows/ci.yml` - phase26_focused selectors include checker, independent peers, pathoracle, and existing cross-backend regressions.
- `.planning/phases/26-checked-scalar-sum/26-03-red-evidence.json` - hosted RED evidence for scalar copy cycle admission.

## Decisions Made

- Scalar `OpCopy` is admissible only for initialized places with matching U64 or Bool types. Ownership copies continue through their existing validator path.
- The validators rederive scalar facts independently; pathoracle does not participate in cycle admission.
- Back-edge diagnostics use the existing stable `check.cfg_back_edge` code, a function-level source span, and a cause kind for owner, resource, loan, loan-derived provenance, or unsupported event semantics.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Kept ownership copies outside scalar CFG detection**
- **Found during:** Task 2
- **Issue:** Treating every `OpCopy` as scalar routed pre-existing ownership functions into the scalar validator.
- **Fix:** Scalar CFG detection remains tied to scalar analysis operations; scalar copy transfer is admitted only after entering that scalar subset.
- **Files modified:** `internal/compiler/check/scalar.go`, `internal/compiler/corevalidate/corevalidate.go`, `internal/compiler/originvalidate/originvalidate.go`
- **Commit:** `f1415dd`

**2. [Rule 1 - Bug] Corrected synthetic scalar-copy fixtures and peer edge checks**
- **Found during:** Tasks 1 and 2
- **Issue:** Initial hosted controls used duplicate or invalid copy targets, and originvalidate did not validate the branch edge identity used by the forged-edge control.
- **Fix:** Mutations now use fresh typed result places; originvalidate resolves edge IDs and verifies their source block and branch pattern.
- **Files modified:** checker and originvalidate tests/validators.
- **Commits:** `61a001b`, `8ccc628`, `7865ff5`, `0bb1c3d`, `9bc8576`, `39c48aa`

**3. [Rule 2 - Missing critical functionality] Added category-specific cycle refusal causes**
- **Found during:** Task 2
- **Issue:** Existing cycle refusals exposed only a generic authority code or cycle block, without the planned owner/resource/loan/provenance/event classification.
- **Fix:** Checker back-edge diagnostics now include stable source spans and classified causes; tests cover each cause class.
- **Files modified:** `internal/compiler/check/check.go`, `internal/compiler/check/scalar.go`, `internal/compiler/check/check_test.go`
- **Commits:** `997f5b5`, `f112a21`, `df2a2bc`, `d653312`

## Hosted Verification

- Focused run **37085604802**, revision `d653312d`: passed on `ubuntu-latest` and `macos-latest`.
- Named controls: `TestBackEdgeRejected`, `TestPhase26FixedPoint`, `TestPhase26AnalysisExhaustion`, `TestPhase26PeerIndependence`, `TestPhase26BackEdgeAuthority`, `TestPhase26BackEdgeCauseCategories`, `TestPhase26CFGBackEdgeCategories`, `TestPhase26ScalarCopyCycleBoundary`, unchanged `TestCompositionCycleGuardFailsClosed`, plus existing Phase 26 session/core/interp/cgen/native selectors.
- Local verification: `gofmt` and `GOCACHE=/tmp/schway-phase26-build-cache go build ./...` passed. No local test suites or native executions were run.
- Earlier RED run **37084053751** failed the intended `TestPhase26ScalarCopyCycleBoundary` assertion on both hosts because scalar `OpCopy` was refused; the normalized receipt is saved in `26-03-red-evidence.json`.
- A full workflow-dispatch run **37085236824** exercised the non-focused aggregate and failed at pre-existing Phase 24 evidence and `go vet` checks. Those failures are outside this plan's focused lane and remain visible for their owning plans.

## TDD Gate Compliance

- RED: hosted run `37084053751` failed the target scalar-copy cycle assertion on both hosts.
- GREEN: hosted run `37085604802` passed the focused selectors on both hosts after implementation and category controls.
- REFACTOR: source was formatted and the repository compile-only build passed; local tests remain intentionally unrun under project constraints.

## Known Stubs

None introduced by this plan.

## Self-Check: PASSED

- SUMMARY file exists at the required phase path.
- All 16 task commits listed by the measured plan ledger exist in git history.
