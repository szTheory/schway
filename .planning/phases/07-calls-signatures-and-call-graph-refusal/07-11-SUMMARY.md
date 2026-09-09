---
phase: 07-calls-signatures-and-call-graph-refusal
plan: 11
subsystem: compiler-check-corevalidate
tags: [go, ownership, affine-types, call-admission, corevalidate, gap-closure]

requires:
  - phase: 07-01..07-10
    provides: check's call admission (types, call-graph refusal, closure-digest chaining), corevalidate's independent peer re-derivations of each, and 07-10's peer-consulted lang check that makes a peer-only refusal CLI-observable
provides:
  - "check.resolveCallBinding consumes a call's argument at admission: a copyable argument (core.AbilityCopy) is copied, a non-copyable argument is moved (all four placeState fields set), reachable through the EXISTING ownership.use_after_move gate on a second use -- closing 07-VERIFICATION.md PVG-01 / 07-REVIEW.md CR-01"
  - "corevalidate independently re-derives the same consume rule from the emitted core artifact alone (its own deriveAbility over types[operation.TypeID].Shape, never check's Abilities list), on BOTH core.OpCall replay arms, refusing with the pre-existing generic core.place_uninitialized law"
  - "two new fixtures pinning both directions: call_argument_used_twice.lang (refusing) and call_argument_used_once.lang (admitting, A/B control)"
  - "two new mutation-killed controls (control:call.argument_consumed_when_noncopyable, control:call.copyable_argument_not_consumed) wired into all three phase-07 registries, each observed to fail on both check and corevalidate sides"
  - "PHASE-07-DEBT.md D-07-07 corrected (no longer claims the single-argument case was handled) and D-07-52 recording the accepted implicit-transfer residual"
affects: [08-interprocedural-loan-liveness-in-check, 09-peer-re-derivation-and-d-03-02-closure]

actuals:
  tokens: 13500
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Ownership law promoted to apply at every place a core.Place is produced from another place (including inside resolveCallBinding), not just the binding switch's five arms -- a promote, not an add-alongside; the binding switch is demoted from 'the law' to 'one of two sites that apply it'"
    - "Two independent consume derivations reading materially different sources: check from the AST binding sequence plus placeState plus the ability package; corevalidate from the emitted core.Program's own operation stream plus its own in-package deriveAbility -- proven independent by bilateral seam-disable tests, not merely asserted"
    - "Cross-package independence test seam exposed as a production (non-test-file) exported setter (check.SetCallArgumentConsumeSeamForTest) mirroring corevalidate.go's pre-existing SetDisableCyclePeerForTest -- the established pattern in this codebase for reaching an unexported same-package seam from the OTHER package's own test binary"

key-files:
  created:
    - testdata/phase07/call_argument_used_twice.lang
    - testdata/phase07/call_argument_used_once.lang
    - internal/compiler/corevalidate/corevalidate_call_argument_consume_internal_test.go
  modified:
    - internal/compiler/check/check.go
    - internal/compiler/check/check_test.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/corevalidate_test.go
    - internal/compiler/corevalidate/export_test.go
    - internal/compiler/session/session_phase7.go
    - internal/compiler/session/session_phase7_test.go
    - scripts/verify-phase7.sh
    - .planning/phases/07-calls-signatures-and-call-graph-refusal/PHASE-07-DEBT.md

key-decisions:
  - "Checkpoint (ratify consume-on-call): auto-approved under auto_advance/yolo mode (gate not blocking-human) -- accepted option (A), the default, verbatim: consumption decided by the argument type's core.AbilityCopy, gate order arity -> unknown-name -> use-after-move -> argument-type -> return-derivation -> consume, D-07-52 implicit-transfer residual disclosed."
  - "check.SetCallArgumentConsumeSeamForTest added directly to check.go (production file, not export_test.go) so corevalidate's own test binary can reach check's seam as an ordinary cross-package import -- mirrors corevalidate.go's pre-existing SetDisableCyclePeerForTest exactly; Go excludes _test.go files (including export_test.go) from a normal cross-package import, so this was the only way to satisfy Task 2's required TestCallConsumePeerIndependentOfCheck."
  - "consumeCallArgument's fail-closed check reuses v.derive's own core.unknown_type_constructor problem rather than minting a new code string -- an absent/unknown type fact already fails the peer through the existing derivation machinery."

requirements-completed: [SEM-05, QLT-08]

coverage:
  - id: D1
    description: "A call CONSUMES a non-copyable argument: the same Buffer passed to two separate calls is refused with ownership.use_after_move (check) and independently with core.place_uninitialized (corevalidate) -- closing 07-VERIFICATION.md PVG-01 / 07-REVIEW.md CR-01"
    requirement: SEM-05
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCallConsumesNoncopyableArgument"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestPeerRefusesDoubleConsumedCallArgument"
        status: pass
      - kind: e2e
        ref: "go run ./cmd/lang --json check testdata/phase07/call_argument_used_twice.lang"
        status: pass
    human_judgment: false
  - id: D2
    description: "A call does NOT consume a copyable argument: a Byte argument passed to two calls, and a Buffer argument passed to exactly one call, both check clean -- pinning the non-refusing direction as hard as the refusing one"
    requirement: SEM-05
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCallDoesNotConsumeCopyableArgument"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestPeerAdmitsRepeatedCopyableCallArgument"
        status: pass
      - kind: e2e
        ref: "go run ./cmd/lang --json check testdata/phase07/call_argument_used_once.lang"
        status: pass
    human_judgment: false
  - id: D3
    description: "The two derivations (check, corevalidate) are provably independent: different sources, different refusal codes, no shared helper, zero real cross-package imports, each side observed refusing with the OTHER side's gate disabled"
    requirement: QLT-08
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestCallConsumePeerIndependentOfCheck"
        status: pass
      - kind: other
        ref: "grep -rn 'compiler/check' internal/compiler/corevalidate/ --include='*.go' | grep -v _test.go"
        status: pass
    human_judgment: false
  - id: D4
    description: "Both new controls are mutation-killed on both sides, in both directions, wired into Phase7RequiredControls(), scripts/verify-phase7.sh, and controlsWithRecordedMutationKill, exact-set-equality unweakened"
    requirement: QLT-08
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCallArgumentConsumeMutationKilled"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCallArgumentConsumeOverRefusalMutationKilled"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestCallConsumePeerMutationMatrix"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase7_test.go#TestPhase7ControlsAreMutationKilled"
        status: pass
      - kind: e2e
        ref: "sh scripts/verify-phase7.sh"
        status: pass
    human_judgment: false
  - id: D5
    description: "D-07-07 corrected (not supplemented) in PHASE-07-DEBT.md: no longer implies the single-argument case was handled; D-07-52 records the accepted implicit-transfer residual"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestDebtRegistersAreWellFormed"
        status: pass
    human_judgment: false

duration: 90min
completed: 2026-09-09
status: complete
---

# Phase 07 Plan 11: A call now consumes its non-copyable argument, closing CR-01/PVG-01 Summary

**`check.resolveCallBinding` and `corevalidate`'s two `core.OpCall` replay arms now apply affine ownership to the call boundary itself: a `Buffer` passed to two separate calls is refused independently by both admission layers, where before this plan it was admitted `status: pass`, exit 0, zero diagnostics.**

## Performance

- **Duration:** ~90 min
- **Tasks:** 3 (plus one auto-ratified checkpoint)
- **Files modified:** 12 (3 created, 9 modified)

## Accomplishments

- `resolveCallBinding`'s declared-function arm (and the `verifyCallInvariantsSeam` mirror) now applies the consume rule exactly once, after the operation and target place are constructed and before the success return: a copyable argument (`core.AbilityCopy`) stays initialized; a non-copyable argument has all four `placeState` fields set (`initialized=false`, `movedAt`, `moveTargetID`, `moveTargetName`), mirroring the `take` arm verbatim. This makes the EXISTING `ownership.use_after_move` gate reachable on a second use, including a second call. Both call arms (`analyzeStraightLine`, `analyzeArmBody`) stay structurally unchanged — the law lives once, in the shared resolver.
- `corevalidate`'s new `consumeCallArgument` helper, called from both `core.OpCall` replay arms (`replayStraightLine`, `replayBlocks`), independently re-derives the argument's copy ability from the emitted core artifact's own `types[operation.TypeID].Shape` via corevalidate's own `v.derive`/`deriveAbility` — never `check`'s recorded `Abilities` list, never the `ability` package, no helper shared with `check`. A non-copyable argument is consumed (`initialized[operation.SourceID] = false`), refused by the pre-existing generic `core.place_uninitialized` law on its next use.
- Two new fixtures: `call_argument_used_twice.lang` (the standing negative control for PVG-01/CR-01) and `call_argument_used_once.lang` (the A/B control pinning the non-refusing direction).
- Bilateral independence proven directly: `TestCallConsumePeerIndependentOfCheck` shows `corevalidate` still refuses with `check`'s own consume gate disabled, and `check` still refuses with the peer's own gate disabled — each side refuses alone.
- Two new controls (`control:call.argument_consumed_when_noncopyable`, `control:call.copyable_argument_not_consumed`) wired into `Phase7RequiredControls()` (22 identifiers), `scripts/verify-phase7.sh`, and `controlsWithRecordedMutationKill`, each mutation-killed on both `check` and `corevalidate` sides, both directions.
- `PHASE-07-DEBT.md`'s D-07-07 wording is corrected (not supplemented): it no longer reads as a claim the single-argument case was handled from 07-03 through 07-09 — it plainly wasn't, closed here — and D-07-52 records the accepted implicit-transfer residual (A-normal form has no `f(take y)` syntax, so the call-site transfer is invisible in source text).

## Task Commits

Each task was committed atomically:

1. **Task 1: a call consumes its non-copyable argument in `check`** - `7e3626d` (feat)
2. **Task 2: `corevalidate` independently refuses the same double-consume** - `e3a147c` (feat)
3. **Task 3: both new controls wired into all three registries, D-07-07 corrected** - `d34e745` (feat)

**Plan metadata:** commit pending (this SUMMARY + STATE.md + ROADMAP.md)

_Checkpoint (ratify consume-on-call as call-site ownership semantics): auto-approved under auto_advance/yolo mode — gate not `blocking-human`, the default/first option accepted verbatim, no separate commit._

## Files Created/Modified

- `internal/compiler/check/check.go` - consume rule added to `resolveCallBinding`'s declared-function arm and the `verifyCallInvariantsSeam` mirror; two new seams (`callArgumentConsumeSeam`, `callArgumentConsumeAlwaysSeam`); `SetCallArgumentConsumeSeamForTest` exported (cross-package test seam, mirroring corevalidate's own precedent)
- `internal/compiler/check/check_test.go` - `TestCallConsumesNoncopyableArgument`, `TestCallDoesNotConsumeCopyableArgument`, `TestCallConsumeMoveStateComplete`, `TestCallArgumentConsumptionUnchangedAcrossAcceptingCorpus`, `TestCallArgumentConsumeMutationKilled`, `TestCallArgumentConsumeOverRefusalMutationKilled`; reset-move-state fix in two pre-existing 07-09 tests (see Deviations)
- `internal/compiler/corevalidate/corevalidate.go` - `consumeCallArgument` (shared by both `core.OpCall` replay arms), `abilityGranted`, two new seams (`disableCallArgumentConsumePeerForTest`, `forceCallArgumentConsumePeerForTest`)
- `internal/compiler/corevalidate/corevalidate_test.go` - synthetic double-consume program builders (straight-line and branch-shaped), `TestPeerRefusesDoubleConsumedCallArgument`, `TestPeerAdmitsRepeatedCopyableCallArgument`, `TestCallConsumePeerMatchesBranchReplayArm`, `TestCallConsumeFailsClosedOnUnresolvableType`, `TestCallArgumentConsumeSeamSettersRestoreCleanly`, `TestCallConsumePeerIndependentOfCheck`, `TestCallConsumePeerMutationMatrix`
- `internal/compiler/corevalidate/corevalidate_call_argument_consume_internal_test.go` (new) - direct `consumeCallArgument` tests proving the decision follows Shape re-derivation, not a hand-corrupted recorded `Abilities` list
- `internal/compiler/corevalidate/export_test.go` - `SetDisableCallArgumentConsumePeerForTest`, `SetForceCallArgumentConsumePeerForTest`
- `internal/compiler/session/session_phase7.go` - `ControlCallArgumentConsumedWhenNoncopyable`, `ControlCallCopyableArgumentNotConsumed` consts; both appended to `Phase7RequiredControls()`
- `internal/compiler/session/session_phase7_test.go` - both controls added to `controlsWithRecordedMutationKill`
- `scripts/verify-phase7.sh` - both new control identifiers added to the phase-07 required-control loop
- `testdata/phase07/call_argument_used_twice.lang` (new) - the refusing direction, standing negative control
- `testdata/phase07/call_argument_used_once.lang` (new) - the non-refusing direction, A/B control
- `.planning/phases/07-calls-signatures-and-call-graph-refusal/PHASE-07-DEBT.md` - D-07-07 corrected, D-07-52 added (items: 8 → 9)

## Decisions Made

- Checkpoint auto-ratified under `auto_advance`/yolo mode: accepted consume-on-call, ability-decided, with the ratified gate order and the D-07-52 disclosure, verbatim as proposed.
- `check.SetCallArgumentConsumeSeamForTest` added to check.go's production source (not `export_test.go`) so it is visible to `corevalidate`'s own test binary as an ordinary cross-package import — Go's build model excludes `_test.go` files (including `export_test.go`) from a normal import, so a same-package-only unexported seam cannot be reached that way. This exactly mirrors `corevalidate.go`'s own pre-existing `SetDisableCyclePeerForTest`, an established pattern in this codebase for the identical problem (07-07's bilateral cycle-independence tests).
- `consumeCallArgument`'s fail-closed behavior on an unresolvable/unknown type fact reuses `v.derive`'s own `core.unknown_type_constructor` problem rather than minting a new code string — the existing derivation machinery already fails closed; no new refusal identity was needed.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug caused directly by this task's change] Two pre-existing 07-09 mutation-kill tests broke from the new consume side effect**
- **Found during:** Task 1
- **Issue:** `TestCallArgumentTypeCheckMutationKilled` and `TestCallReturnTypeDerivationMutationKilled` (check_test.go) each call `resolveCallBinding` twice against the SAME shared `places` map and the SAME argument place, first with a seam engaged then with it disengaged. Once the consume rule started mutating `placeState` as a side effect of admission, the first (seam-engaged) call's consume left the place moved, so the second (production) call hit the unrelated `ownership.use_after_move` gate before ever reaching the gate each test actually isolates.
- **Fix:** Reset the place's four move-state fields (`initialized`, `movedAt`, `moveTargetID`, `moveTargetName`) between the two calls in both tests.
- **Files modified:** internal/compiler/check/check_test.go
- **Verification:** `go test ./internal/compiler/check/...` green; both tests pass and still isolate their own gate
- **Committed in:** 7e3626d (Task 1 commit)

**2. [Rule 3 - Blocking, cross-package test seam] `check.SetCallArgumentConsumeSeamForTest` added to check.go (production file), not in Task 2's declared `<files>`**
- **Found during:** Task 2
- **Issue:** Task 2 requires `TestCallConsumePeerIndependentOfCheck` to run under `go test ./internal/compiler/corevalidate/`, proving each side refuses with the OTHER side's gate disabled. Reaching `check`'s own unexported `callArgumentConsumeSeam` from `corevalidate`'s test binary is impossible via a same-package-only var or via an `export_test.go`-only setter (Go excludes `_test.go` files from a normal cross-package import) — the plan's own `read_first`/`files` list for Task 2 did not name `check.go`, but the required test could not be written without a production-visible setter there.
- **Fix:** Added `check.SetCallArgumentConsumeSeamForTest` directly to check.go, mirroring corevalidate.go's own pre-existing `SetDisableCyclePeerForTest` (07-07's identical cross-package bilateral-test pattern) — deliberately minimal, test-only-no-op-in-production, always restored via its returned closure, never called from any production code path.
- **Files modified:** internal/compiler/check/check.go (declaration added), internal/compiler/corevalidate/corevalidate_test.go (imports `compiler/check` as an ordinary test-only dependency)
- **Verification:** `grep -rn 'compiler/check' internal/compiler/corevalidate/ --include='*.go' | grep -v _test.go` stays empty — the independence property under test (corevalidate.go's own PRODUCTION source never importing check) is untouched; `TestCallConsumePeerIndependentOfCheck` passes both directions
- **Committed in:** e3a147c (Task 2 commit)

---

**Total deviations:** 2 auto-fixed (1 Rule 1 bug caused directly by this task's own change, 1 Rule 3 blocking addition required to satisfy the plan's own required test). Neither changes what the plan delivers; the second is exactly the established cross-package-seam pattern this codebase already uses elsewhere (D-07-42).
**Impact on plan:** No scope creep — both fixes are necessary corrections to keep the plan's own required tests and gates green.

## Issues Encountered

None beyond the two deviations above (resolved inline, not left open). The first `sh scripts/verify-phase7.sh` run hit two pre-existing environment-contention flakes (`cache.input_undeclared`, cgen's `native.timeout`) unrelated to `check`/`corevalidate`/`session` — both confirmed passing in isolated re-runs, matching 07-09/07-10-SUMMARY.md's own documented precedent for this machine; a clean re-run of the full script passed outright (exit 0, both new controls `pass` in `lane:kind-exhaustive-dispatch-phase07`).

## Known Stubs

None.

## Threat Flags

None — every new surface (the consume rule at the call boundary, the two seams on each side, the cross-package independence seam) is inside this plan's own declared `<threat_model>`.

## User Setup Required

None - no external service configuration required.

## Verification Results (re-run live)

- `go build ./...` — clean.
- `go vet ./...` — clean.
- `go test ./... -p 1` — all packages `ok`, zero `FAIL` (except the two known pre-existing environment flakes noted above, both re-confirmed passing in isolation).
- `go test -race ./...` — all packages `ok`, zero `DATA RACE`, zero `FAIL` in `check`/`corevalidate`/`session` (one pre-existing `cgen` timeout flake under `-race` load, re-confirmed passing in isolation).
- `sh scripts/verify-phase7.sh` — exits 0 on the clean re-run; `phase07.json`'s `lane:kind-exhaustive-dispatch-phase07` lists all 22 controls, all `status: pass`, including both new identifiers.
- `go run ./cmd/lang --json check testdata/phase07/call_argument_used_twice.lang` — `status: invalid`, `ownership.use_after_move`, `Primary` on the second call site, `moved_here` cause on the first — exit 1.
- `go run ./cmd/lang --json check testdata/phase07/call_argument_used_once.lang` — `status: pass`, exit 0.
- Every other fixture in `testdata/phase07` (13 fixtures) swept live: each reports exactly the status/code 07-VERIFICATION.md's Behavioral Spot-Checks table records, as amended by 07-10's two declared divergences (`relay_escort_witness.lang` → `core.move_while_borrowed` via the CLI peer consult; `duplicate_function_name.lang` → `core.duplicate_function_id`).
- `git status --porcelain testdata/` — empty (both new fixtures already committed in Task 1; no golden/core artifact touched).
- `git diff internal/compiler/check/check.go` — contains no hunk touching `computeLoanLastUses`, and no added line inside either call arm's `if binding.RHS.Kind == "call"` block.
- `git diff --name-only -- go.mod` — empty (this repo has no `go.sum`; no dependency added).

## Next Phase Readiness

- 07-REVIEW.md CR-01 / 07-VERIFICATION.md PVG-01 is closed: the call boundary no longer bypasses affine ownership, independently on both admission layers, reachable through the CLI channel 07-10 built.
- PVG-04 / CR-02 remains untouched and explicit Phase 08 scope (D-07-49): `check.computeLoanLastUses` still has no `"call"` case.
- The `relay_escort_witness.lang` check/corevalidate divergence (D-03-02) is neither closed, widened, nor narrowed by this plan — still deferred to Phase 08/09.
- D-07-52 (the implicit call-site transfer) is a new, explicitly accepted residual for a future grammar phase to reopen if a call-site `take` marker is ever added.
- 07-12 (PVG-02/CR-03, closure-derived signature peer) is the phase's remaining post-verification gap.

---
*Phase: 07-calls-signatures-and-call-graph-refusal*
*Completed: 2026-09-09*

## Self-Check: PASSED

All key files confirmed present on disk (check.go, check_test.go, corevalidate.go, corevalidate_test.go, corevalidate_call_argument_consume_internal_test.go, export_test.go, session_phase7.go, session_phase7_test.go, scripts/verify-phase7.sh, testdata/phase07/call_argument_used_twice.lang, testdata/phase07/call_argument_used_once.lang, PHASE-07-DEBT.md); all three task commits (7e3626d, e3a147c, d34e745) confirmed in `git log --oneline --all`.
