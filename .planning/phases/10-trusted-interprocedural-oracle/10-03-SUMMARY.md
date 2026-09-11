---
phase: 10-trusted-interprocedural-oracle
plan: 03
subsystem: pathoracle
tags: [interprocedural, path-composition, mutation-testing, go-list-deps, ownership]

# Dependency graph
requires:
  - phase: 10-trusted-interprocedural-oracle
    provides: "plan 10-01's real []frame call-stack executor and plan 10-02's callee-contract origin walk, the two peers whose contract-hop pattern at OpCall this plan deliberately does NOT copy"
provides:
  - "pathoracle.RecomputeEndpoints composes across core.OpCall by recursing into the callee via its OWN EnumeratePaths (composeCall/composeCarriesOwnLoan, pathoracle_compose.go), re-deriving PER CALLEE PATH whether that path's own return carries a loan the callee itself created -- never a declared contract"
  - "A separate, genuinely reachable composition-depth cap (MaxCompositionDepth) with its own typed, fail-closed refusal (compositionDepthError) distinct from MaxPaths' pathoracle.path_count_exceeded, plus composition's own mutual-recursion cycle guard (compositionCycleError) distinct from EnumeratePaths' intra-function back-edge handling"
  - "TestCompositionDiscriminatesPerPathBorrow (D-10-14): a callee with two concrete paths where only one borrows its own parameter into its return splits the caller's composed endpoints into an edge-kind and a point-kind endpoint; the D-10-55 seeded contract-hop fault (pathoracle.SetForceContractHopForTest) collapses the split to one uniform kind, with a companion assertion that check's/corevalidate's own answers never move"
  - "pathoracle's import-independence guard hardened from direct-import-only to also transitive (go list -deps), mirroring originvalidate's own D-10-17 hardening, proven load-bearing by a negative control"
affects: [10-04, 10-05, 10-06, 10-07, 10-08, 10-09, 11]

# Actuals (#2632)
actuals:
  tokens: 15200
  tasks: 3
  commits: 4

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Per-callee-path re-derivation (composeCarriesOwnLoan) rather than a per-function declared-contract hop: the SAME question check.derivePlaceLoans/corevalidate.buildLoanChainIndex answer once per function, pathoracle answers once per concrete callee path, by recursing into the callee's own EnumeratePaths/linearizePath and asking whether a loan the callee itself created on that path survives to its own core.OpReturn"
    - "Path-forking forward replay (composeLinearizeCaller): the caller's own linearizePath-equivalent forks into one variant per distinct composeCall answer at each core.OpCall, still replaying each concrete continuation exactly once -- no convergence loop, fixpoint, or relation closure anywhere"
    - "Two independently-capped recursion limits reporting which one fired (D-10-12): MaxPaths (per-function path count, deliberately set above its real reachable maximum) and MaxCompositionDepth (call-frame depth, set at a genuinely reachable bound), each with its own typed error and Code()"
    - "D-10-55 seeded contract-hop fault as an unexported, package-scoped seam with NO exported Set/Force/Disable symbol in the production file itself (grep-verified): the export_test.go bridge (mirroring originvalidate's identical convention) is the only path in, making the package boundary itself the independence proof"
    - "Real-pipeline-checked-but-stitched fixture pair, used when the grammar/checker cannot express one joint calling program for the shape under test: two fixtures are each independently checked through syntax.Parse -> check.Program, and the test rewrites one core.Function's own OpCall.CalleeID onto the other's real ID before driving both through pathoracle, which never itself consults Callable"

key-files:
  created:
    - internal/compiler/pathoracle/pathoracle_compose.go
    - internal/compiler/pathoracle/export_test.go
    - testdata/phase10/compose_per_path_borrow_callee_accept.lang
    - testdata/phase10/compose_per_path_borrow_caller_accept.lang
    - .planning/phases/10-trusted-interprocedural-oracle/deferred-items.md
  modified:
    - internal/compiler/pathoracle/pathoracle.go
    - internal/compiler/pathoracle/pathoracle_test.go
    - internal/compiler/core/core_test.go
    - internal/compiler/check/check_test.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_phase7.go

key-decisions:
  - "RecomputeEndpoints widened with a new CalleeLookup parameter (func(functionID string) (core.Function, bool)), never a bare core.Program (D-10-13) -- BuildCalleeLookup (the map's builder) deliberately lives in pathoracle.go, not pathoracle_compose.go, so the composition file itself never needs to know what a core.Program is, satisfying the plan's own grep-verified boundary"
  - "composeCarriesOwnLoan deliberately does NOT seed the callee's own parameter chain with anything the caller's argument carries (D-10-11): the per-callee-path fact is purely intraprocedural (does THIS path's own core.OpBorrowShared/Exclusive survive to its own core.OpReturn), and composeLinearizeCaller uses that boolean purely as the gate for whether the CALLER's own chain propagates across the call -- exactly mirroring derivePlaceLoans'/buildLoanChainIndex's per-function gate shape, re-derived per path instead of trusted from a declared contract"
  - "The discriminating fixture's callee and caller are checked as two SEPARATE programs, not one joint calling program: a match-bodied function whose undeclared return diverges per-arm between borrow and owned triggers core.origin_omitted (SEM-06's Callable gate), and a match-bodied function cannot declare a borrowed return origin at all this phase (ownership.match_borrowed_return_unsupported) -- so no single real .lang program can express a genuinely-Callable, genuinely-divergent two-path callee. The plan's own action text anticipated exactly this fallback (\"a checked-core construction built by the same pipeline from the closest expressible source\"); the test stitches the caller's real OpCall.CalleeID onto the callee's real ID post-check, since pathoracle.RecomputeEndpoints never consults Callable at all"
  - "The composed-fork endpoint split (synthesizeComposedEndpoints) is handled as a dedicated code path in RecomputeEndpoints, bypassing the pre-existing single-answer `deaths` map bookkeeping entirely for a forked caller path -- that bookkeeping's own inconsistent_path_death guard exists to catch a genuine bug (the SAME loan disagreeing on its death position across genuinely distinct top-level paths), and treating a load-bearing composed divergence the same way would misfire that guard"

patterns-established:
  - "A composition-caused per-path split maps to the SAME edge/point vocabulary RecomputeEndpoints already uses for a local branch fan-out, without requiring the call site to be a real multi-successor block in the caller's own CFG -- the composed variant that keeps propagating the loan reaches an ordinary point at its own eventual, latest position; every earlier composed variant's position becomes a distinct, synthetic-edge-ID endpoint"

requirements-completed: [TRU-03]

coverage:
  - id: D1
    description: "pathoracle composes callee path enumerations across core.OpCall, re-deriving per callee path (never a declared contract) whether the callee's own return carries a loan it created"
    requirement: TRU-03
    verification:
      - kind: unit
        ref: "pathoracle_test.go#TestCompositionDiscriminatesPerPathBorrow"
        status: pass
      - kind: integration
        ref: "go test ./internal/compiler/pathoracle/... ./internal/compiler/check/... ./internal/compiler/corevalidate/... ./internal/compiler/session/... ./internal/compiler/core/..."
        status: pass
    human_judgment: false
  - id: D2
    description: "Composition depth is bounded by its own declared, typed, fail-closed cap, distinguishable from MaxPaths"
    requirement: TRU-03
    verification:
      - kind: unit
        ref: "pathoracle_test.go#TestCompositionDepthCapRejects"
        status: pass
      - kind: unit
        ref: "pathoracle_test.go#TestOraclePathCountCapRejects"
        status: pass
    human_judgment: false
  - id: D3
    description: "Composition carries its own guard against mutual recursion across the call edge, distinct from EnumeratePaths' intra-function back-edge handling, and fails closed within bounded time"
    verification:
      - kind: unit
        ref: "pathoracle_test.go#TestCompositionCycleGuardFailsClosed"
        status: pass
    human_judgment: false
  - id: D4
    description: "The per-path borrow split survives composition and demonstrably collapses under a stubbed contract hop, with check's and corevalidate's own answers proven unaffected"
    requirement: TRU-03
    verification:
      - kind: unit
        ref: "pathoracle_test.go#TestCompositionDiscriminatesPerPathBorrow"
        status: pass
    human_judgment: true
    rationale: "The specific edge-vs-point mapping this plan chose (which composed variant maps to which endpoint kind) is a design decision this plan made under acknowledged latitude, not dictated by the plan text; a reviewer should confirm the chosen mapping's rationale (documented inline in synthesizeComposedEndpoints) is sound, not just that both kinds appear."
  - id: D5
    description: "pathoracle transitively imports none of check, corevalidate, ast, interp, or cgen, enforced by a go list -deps check that fails CI, proven load-bearing by a negative control"
    requirement: TRU-03
    verification:
      - kind: unit
        ref: "pathoracle_test.go#TestOracleImportsStayIndependent"
        status: pass
      - kind: unit
        ref: "pathoracle_test.go#TestTransitiveImportsGuardCanFail"
        status: pass
      - kind: other
        ref: "go list -deps ./internal/compiler/pathoracle | grep -c 'compiler/check\\|compiler/corevalidate\\|compiler/ast\\|compiler/interp\\|compiler/cgen'"
        status: pass
    human_judgment: false

duration: 100min
completed: 2026-09-11
status: complete
---

# Phase 10 Plan 3: Trusted Interprocedural Oracle — Cross-Function Path Composition Summary

`pathoracle` independently re-derives the cross-function loan-chain rule by recursing into a callee's own `EnumeratePaths` and re-deriving per-path (never per-function-contract) whether a loan survives the call, proven load-bearing by a discriminating two-arm fixture that only real composition can distinguish from a stubbed contract hop.

## Performance

- **Duration:** 100 min
- **Started:** 2026-09-11 (see git commit timestamps)
- **Completed:** 2026-09-11
- **Tasks:** 3 completed (plus 1 checkpoint:decision, auto-ratified)
- **Files modified:** 10

## Accomplishments

- `pathoracle` now handles `core.OpCall` by recursing into the callee via its OWN `EnumeratePaths` (`composeCall`), re-deriving per callee path whether that path's own return carries a loan the callee itself created (`composeCarriesOwnLoan`) -- deliberately never consulting a declared return contract, the exact distinction D-10-09/D-10-11/D-10-14 require. The caller's own forward replay (`composeLinearizeCaller`) forks at each `core.OpCall` into one variant per distinct callee-path answer, still running the existing per-operation replay exactly once over each concrete continuation -- no convergence loop, fixpoint, or relation closure anywhere on the composition path.
- A separate, genuinely reachable `MaxCompositionDepth` cap with its own typed `compositionDepthError` (a `Code()` distinct from `MaxPaths`' `pathoracle.path_count_exceeded`) bounds recursion depth; `MaxPaths` itself is unchanged (`grep -c 'MaxPaths = 4096'` still returns 1). A separate `compositionCycleError` guards against mutual recursion across the call edge in a corrupted artifact, distinct from `EnumeratePaths`' own intra-function back-edge handling, proven to fail closed within bounded time (`TestCompositionCycleGuardFailsClosed`).
- `RecomputeEndpoints` widened with a narrow `CalleeLookup` capability (`func(functionID string) (core.Function, bool)`), never a bare `core.Program` (`grep -c 'core.Program' pathoracle_compose.go` returns 0) -- every existing call site (core_test.go, check_test.go, session.go, session_phase7.go, pathoracle_test.go) updated to pass `nil` or a real lookup.
- `TestCompositionDiscriminatesPerPathBorrow` (D-10-14) is the mutation-kill proof: a callee (`compose_per_path_borrow_callee_accept.lang`) whose two match arms diverge -- one borrows its own parameter and returns it, the other takes (owns) it -- causes the caller's (`compose_per_path_borrow_caller_accept.lang`) composed endpoint set to split into a genuine edge-kind and point-kind endpoint. Engaging the D-10-55 seeded contract-hop fault (`pathoracle.SetForceContractHopForTest`, an unexported seam bridged only via `export_test.go`, mirroring originvalidate's identical convention) collapses the split to one uniform kind, and a companion assertion proves `check`'s own stored endpoints and `corevalidate.Validate`'s own agreement with the same caller fixture are byte-for-byte unchanged while the seam is engaged.
- `TestOracleImportsStayIndependent` hardened from a direct `go/parser` scan to ALSO run a transitive `go list -deps` check (D-10-17), mirroring originvalidate's own hardening exactly; the five-entry forbidden list is unchanged, and `TestTransitiveImportsGuardCanFail` is the negative control proving the guard can fire.

## Task Commits

Each task was committed atomically:

1. **Decision: composition over contract-hop is a one-way commitment** — auto-ratified (see note below)
2. **Task 1: Compose callee paths across OpCall, with its own declared depth cap** - `c45d466` (feat)
3. **Task 2: The discriminating per-path-borrow fixture and its seeded contract-hop fault** - `0185622` (test)
4. **Task 3: Harden pathoracle's import guard to a transitive check** - `c040f61` (test)
5. **Rule 1 fix: bound the `go list -deps` spawn discovered by `TestSourceNeverSpawnsUnboundedProcesses`** - `9241354` (fix)

_Note: the checkpoint:decision task (adopting recursive composition over a contract hop, D-10-09) was auto-ratified under `workflow.auto_advance: true` / `mode: yolo` -- its gate was the default `"blocking"` (no `gate="blocking-human"` override), and "Proceed" is the plan's own documented default option, already locked in `10-CONTEXT.md`. No separate commit; the decision is realized by Task 1's implementation._

## Files Created/Modified

- `internal/compiler/pathoracle/pathoracle_compose.go` — new file: `CalleeLookup`, `MaxCompositionDepth`, `compositionDepthError`/`CompositionDepthError`, `compositionCycleError`/`CompositionCycleError`, `forceContractHopForTest`, `composeCall`, `composeCarriesOwnLoan`, `composedVariant`, `composeLinearizeCaller`, `synthesizeComposedEndpoints`, `hasOpCall`, clone helpers.
- `internal/compiler/pathoracle/pathoracle.go` — widened package doc (the enumerate-and-replay property composition preserves, D-10-11), `BuildCalleeLookup`, widened `RecomputeEndpoints` signature and its per-path loop to route through `composeLinearizeCaller` when a path contains `core.OpCall`, else the unchanged `linearizePath`.
- `internal/compiler/pathoracle/export_test.go` — new file: `SetForceContractHopForTest` (D-10-55 bridge).
- `internal/compiler/pathoracle/pathoracle_test.go` — `chainCallFunction`, `TestCompositionDepthCapRejects`, `TestCompositionCycleGuardFailsClosed`, `checkedProgram`, `functionNamed`, `withCalleeIDRewritten`, `endpointKinds`, `TestCompositionDiscriminatesPerPathBorrow`, transitive-import hardening (`pathOracleForbiddenImports`, `directImportViolation`, `transitiveImportViolation`, `transitiveImportsViolation`, `TestTransitiveImportsGuardCanFail`), `boundedGoListWriter`; updated all pre-existing `RecomputeEndpoints` call sites to the widened signature.
- `internal/compiler/core/core_test.go`, `internal/compiler/check/check_test.go`, `internal/compiler/session/session.go`, `internal/compiler/session/session_phase7.go` — updated their own `RecomputeEndpoints` call sites to pass `nil` (none of these fixtures contain `core.OpCall`).
- `testdata/phase10/compose_per_path_borrow_callee_accept.lang`, `testdata/phase10/compose_per_path_borrow_caller_accept.lang` — new fixtures (see Deviations for why they are two separate checked programs).
- `.planning/phases/10-trusted-interprocedural-oracle/deferred-items.md` — new file, recording an out-of-scope pre-existing lint violation discovered but not fixed (see Deviations).

## Decisions Made

See `key-decisions` in frontmatter.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `pathoracle_compose.go`'s comments literally contained the string `core.Program`, failing the plan's own grep-verified acceptance criterion**
- **Found during:** Task 1, running `grep -c 'core.Program' internal/compiler/pathoracle/pathoracle_compose.go`
- **Issue:** `BuildCalleeLookup` was initially written inside `pathoracle_compose.go` (it takes a `core.Program`), and two doc comments referenced `core.Program` by name. The plan's own acceptance criterion requires this file to contain zero occurrences of that literal string, so that `RecomputeEndpoints`'s new parameter is verifiably a callee-indexed lookup, never a bare program value.
- **Fix:** Relocated `BuildCalleeLookup` to `pathoracle.go` and reworded the two comments to avoid the literal string while preserving their meaning.
- **Files modified:** `internal/compiler/pathoracle/pathoracle.go`, `internal/compiler/pathoracle/pathoracle_compose.go`
- **Verification:** `grep -c 'core.Program' internal/compiler/pathoracle/pathoracle_compose.go` now prints `0`.
- **Committed in:** `c45d466`

**2. [Rule 1 - Bug] The new transitive-import-guard helper spawned a bare, unbounded `go list -deps` process**
- **Found during:** Task 3, `go test ./...` (discovered via `internal/compiler/native/native_test.go`'s `TestSourceNeverSpawnsUnboundedProcesses`, D-02-01)
- **Issue:** `transitiveImportsViolation`'s first draft used `exec.Command` (no context deadline) and `.Output()` (an unbounded merged-output buffer) -- both forbidden for every process this repository spawns.
- **Fix:** Switched to `exec.CommandContext` with a 30-second deadline and a small size-capped `Stdout` writer local to the test file (`boundedGoListWriter`), mirroring `testsupport`'s own `boundedWriter` shape without depending on its unexported type.
- **Files modified:** `internal/compiler/pathoracle/pathoracle_test.go`
- **Verification:** `go test ./internal/compiler/native/... -run TestSourceNeverSpawnsUnboundedProcesses` no longer names `pathoracle_test.go`.
- **Committed in:** `9241354`

---

**Total deviations:** 2 auto-fixed (both Rule 1).
**Impact on plan:** Both fixes were necessary for correctness/verifiability and directly caused by this plan's own new code. No scope creep: no new feature or architectural surface was added beyond what Tasks 1-3 specify.

### Scope boundary (not fixed, logged instead)

`internal/compiler/originvalidate/originvalidate_test.go`'s own `transitiveImportsViolation` (landed in Plan 10-02, commit `7d7553b`) has the IDENTICAL unbounded-spawn pattern this plan's Task 3 fixed for pathoracle. Per the scope-boundary rule (only auto-fix issues directly caused by the CURRENT task's changes), this pre-existing instance was left untouched and recorded in `.planning/phases/10-trusted-interprocedural-oracle/deferred-items.md`. `go test ./internal/compiler/native/... -run TestSourceNeverSpawnsUnboundedProcesses` currently fails on `main` for this one remaining reason.

## Issues Encountered

- The plan's own action text for Task 2 anticipated that the grammar/checker might not be able to express a genuinely-Callable, genuinely-divergent two-path callee as one joint calling program, with an explicit fallback instruction. That anticipation proved correct: a match-bodied function whose undeclared return diverges per-arm between borrow and owned trips `core.origin_omitted` (making it un-Callable under SEM-06), and a match-bodied function cannot declare a borrowed return origin at all this phase (`ownership.match_borrowed_return_unsupported`). Resolved per the plan's own fallback: two separately-checked fixtures, stitched post-check by rewriting one `core.Function`'s own `OpCall.CalleeID`, since `pathoracle.RecomputeEndpoints` never itself consults `Callable`.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Success Criterion 1 is now closed for BOTH halves: plan 10-02 closed the `OpCall` origin walk, and this plan closes the cross-function loan-chain composition. All three admission/re-derivation layers (`check`, `corevalidate`, `pathoracle`) now independently handle `core.OpCall`, and `pathoracle`'s own procedure is verified structurally distinct at exactly the point the independence claim is tested (D-10-14's mutation-kill proof).
- The composed enumeration (`composeCall`/`composeLinearizeCaller`) is the substrate plan 10-07's declared-depth corpus and plan 10-08's criterion-4 differential are both built on, per the plan's own `key_links`. `MaxCompositionDepth`'s current value (8) is a placeholder the plan explicitly defers tuning to plan 10-07's own bidirectional corpus gate.
- The must_haves' own flagged assumption ("EDGE (TRU-03, unclassified -- UNRESOLVED): the deterministic edge probe could not classify TRU-03's edge category... reviewer must confirm manually that cross-function path composition has no unconsidered edge beyond the composition-depth cap and the mutual-recursion guard covered here") is UNCHANGED by this plan -- still flagged, not dismissed. A human reviewer should confirm no other composition edge case exists beyond the two covered here.
- Ready for plan 10-04.

## Self-Check: PASSED

- `internal/compiler/pathoracle/pathoracle_compose.go` — FOUND
- `internal/compiler/pathoracle/export_test.go` — FOUND
- `testdata/phase10/compose_per_path_borrow_callee_accept.lang` — FOUND
- `testdata/phase10/compose_per_path_borrow_caller_accept.lang` — FOUND
- `.planning/phases/10-trusted-interprocedural-oracle/deferred-items.md` — FOUND
- Commit `c45d466` — FOUND (`git log --oneline --all`)
- Commit `0185622` — FOUND
- Commit `c040f61` — FOUND
- Commit `9241354` — FOUND
- `go test ./internal/compiler/pathoracle/... -v` — PASS (13/13 tests, all named)
- `go test ./internal/compiler/pathoracle/... -count=2 -shuffle=on` — PASS
- `go test ./internal/compiler/check/... ./internal/compiler/corevalidate/... ./internal/compiler/session/... ./internal/compiler/core/...` — PASS
- `go list -deps ./internal/compiler/pathoracle | grep -c 'compiler/check\|compiler/corevalidate\|compiler/ast\|compiler/interp\|compiler/cgen'` — prints `0`
- `grep -c 'MaxPaths = 4096' internal/compiler/pathoracle/pathoracle.go` — prints `1`
- `grep -c 'core.Program' internal/compiler/pathoracle/pathoracle_compose.go` — prints `0`
- `go build ./... && go vet ./...` — PASS

---
*Phase: 10-trusted-interprocedural-oracle*
*Completed: 2026-09-11*
