---
phase: 07-calls-signatures-and-call-graph-refusal
plan: 12
subsystem: semantic-core
tags: [closure-digest, foreign-reach, lattice-join, corevalidate, originvalidate, mutation-testing, gap-closure]

requires:
  - phase: 07-calls-signatures-and-call-graph-refusal
    provides: "07-08's ClosureDigest chaining (chainOrder/peerPostorder, callee-before-caller) and 07-10/07-11's peer-consulted lang check plus call-argument consume"
provides:
  - "core.FunctionSignature.Foreign and .Fails are genuinely closure-derived: a caller of a transitively fallible, libc-reaching callee now publishes the real reach, not the zero/empty value core.go documents as meaning 'none' -- closes 07-REVIEW.md CR-03 / 07-VERIFICATION.md PVG-02"
  - "an explicit, order-independent worst-case lattice join over ForeignReach (joinForeignReach/peerJoinForeignReach), with a declared core.ForeignReachConflict sentinel for an unorderable allocator disagreement, run before ClosureDigest on both the producer and an independently-written peer"
  - "the IN-01 index-coupling fix: originvalidate.BuildInterface's second pass resolves the program function by ID (programFunctionByID) instead of an index shared with summary.Functions"
  - "two new mutation-killed controls (control:summary.foreign_reach_closure_derived, control:summary.fails_closure_derived) wired into all three phase-07 registries, each killed independently on the producer and the peer"
  - "testdata/phase07/call_fallible_foreign_reach.lang as the standing CR-03/PVG-02 witness"
  - "PHASE-07-DEBT.md D-07-53 (join granularity, disclosed) and D-07-54 (IN-02/IN-03 deliberately carried)"
affects: [08-interprocedural-loan-liveness-in-check, 09-peer-re-derivation-and-d-03-02-closure]

actuals:
  tokens: 15200
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Worst-case lattice join before the digest: both originvalidate's chainOrder loop and corevalidate's own peerPostorder walk join a caller's Foreign/Fails with every already-joined callee's PUBLISHED signature (never a body), strictly before ClosureDigest is computed, so 07-08's callee-changes-invalidates-caller property extends to these two fields automatically"
    - "Two independently-written joins, one shared schema sentinel: joinForeignReach (originvalidate) and peerJoinForeignReach (corevalidate) are separate functions with no shared helper, sharing only core.ForeignReachConflict -- a schema constant, not a derivation -- mirroring core.CallGraphCycle's precedent for a value shared verbatim across independent implementations"
    - "Fails join is deliberately NOT order-independent (first-non-empty/caller-nearest wins), disclosed as D-07-53 rather than hidden -- unlike Foreign's true commutative/associative lattice join, a single string cannot express a union of two distinct error types"

key-files:
  created:
    - testdata/phase07/call_fallible_foreign_reach.lang
    - internal/compiler/originvalidate/originvalidate_foreign_closure_test.go
    - internal/compiler/corevalidate/corevalidate_foreign_closure_test.go
  modified:
    - internal/compiler/core/core.go
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/originvalidate/export_test.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/export_test.go
    - internal/compiler/session/session_phase7.go
    - internal/compiler/session/session_phase7_test.go
    - scripts/verify-phase7.sh
    - .planning/phases/07-calls-signatures-and-call-graph-refusal/PHASE-07-DEBT.md

key-decisions:
  - "The two fault-injection seams for the producer's join (foreignClosureJoinSeam/failsClosureJoinSeam) had to be declared as production-visible functions in originvalidate.go itself, not export_test.go -- corevalidate's own bilateral-independence test (TestForeignClosureJoinPeersIndependent) imports originvalidate as an ordinary cross-package dependency, and Go excludes every _test.go file (export_test.go included) from that build. This exactly mirrors 07-11's SetCallArgumentConsumeSeamForTest precedent."
  - "joinFails/peerJoinFails deliberately break order-independence (first-non-empty/'existing' wins) where joinForeignReach/peerJoinForeignReach do not: the plan's own acceptance criteria name commutativity/associativity ONLY for the Foreign join (TestForeignJoinIsOrderIndependent); Fails' single-string schema cannot express a union of two distinct error types, so the imprecision is disclosed as D-07-53 rather than forced into a false lattice shape."
  - "TestSecondPassResolvesProgramFunctionByID proves the IN-01 fix by building the SAME three-function set in two different program.Functions declaration orders and asserting each function's published Foreign/Fails depends only on its own identity and its own callees, never on array position -- the historical index-coupling bug is dormant under today's code shape (no loop ever skips a function), so this is a regression-shaped proof of the invariant rather than a reproduction of a live bug."

requirements-completed: [SEM-05, QLT-08]

coverage:
  - id: D1
    description: "core.FunctionSignature.Foreign and .Fails are closure-derived: main (which never itself declares a ForeignContract) publishes the same non-empty ForeignReach/Fails as tracer, its transitively-foreign-reaching callee -- closing CR-03/PVG-02 on both the producer and an independent peer"
    requirement: SEM-05
    verification:
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_foreign_closure_test.go#TestForeignReachIsClosureDerived"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_foreign_closure_test.go#TestFailsIsClosureDerived"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_foreign_closure_test.go#TestPeerForeignReachIsClosureDerived"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_foreign_closure_test.go#TestPeerFailsIsClosureDerived"
        status: pass
      - kind: e2e
        ref: "go run ./cmd/lang --json interface export testdata/phase07/call_fallible_foreign_reach.lang"
        status: pass
    human_judgment: false
  - id: D2
    description: "The join is an explicit, order-independent worst-case lattice join over ForeignReach (never last-writer-wins), with a declared core.ForeignReachConflict sentinel for an unorderable allocator disagreement; runs before ClosureDigest so callee-changes-invalidates-caller extends to these fields; and the IN-01 index-coupling defect is fixed in the same pass"
    requirement: SEM-05
    verification:
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_foreign_closure_test.go#TestForeignJoinIsOrderIndependent"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_foreign_closure_test.go#TestForeignReachClosureLeafUnperturbed"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_foreign_closure_test.go#TestSecondPassResolvesProgramFunctionByID"
        status: pass
    human_judgment: false
  - id: D3
    description: "Producer and peer derive the joined Foreign/Fails through materially different traversals (originvalidate's chainOrder vs corevalidate's own checkCallGraphAcyclic-derived peerPostorder/peerAdjacency) with no shared helper and no cross-package import in either direction; each side is observed still joining with the other side seamed off"
    requirement: QLT-08
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_foreign_closure_test.go#TestForeignClosureJoinPeersIndependent"
        status: pass
      - kind: other
        ref: "grep -rn 'compiler/originvalidate' internal/compiler/corevalidate/ | grep -v _test.go"
        status: pass
      - kind: other
        ref: "grep -rn 'compiler/corevalidate' internal/compiler/originvalidate/ | grep -v _test.go"
        status: pass
    human_judgment: false
  - id: D4
    description: "Two new controls (control:summary.foreign_reach_closure_derived, control:summary.fails_closure_derived) are declared, wired into Phase7RequiredControls() (24 identifiers), scripts/verify-phase7.sh, and controlsWithRecordedMutationKill; each observed to fail under a seeded fault on both the producer and the peer; exact-set-equality gates stay green unweakened"
    requirement: QLT-08
    verification:
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_foreign_closure_test.go#TestForeignClosureJoinMutationKilled"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_foreign_closure_test.go#TestFailsClosureJoinMutationKilled"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_foreign_closure_test.go#TestForeignClosureJoinPeerMutationMatrix"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase7_test.go#TestPhase7ControlsAreMutationKilled"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase7_test.go#TestPhase7RequiredControlsMatchScript"
        status: pass
      - kind: e2e
        ref: "sh scripts/verify-phase7.sh"
        status: pass
    human_judgment: false
  - id: D5
    description: "The join's disclosed granularity limits (D-07-53) and the carried check-side INFO items IN-02/IN-03 (D-07-54) are recorded in PHASE-07-DEBT.md rather than implied"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestDebtRegistersAreWellFormed"
        status: pass
    human_judgment: false

duration: 55min
completed: 2026-09-09
status: complete
---

# Phase 07 Plan 12: Foreign/Fails closure-derived, closing CR-03/PVG-02 Summary

**`originvalidate.BuildInterface`'s existing acyclic second pass now joins each caller's published `Foreign`/`Fails` with every already-joined callee's published signature before computing `ClosureDigest`, and `corevalidate` independently re-derives the same join over its own postorder with no shared helper -- a caller of a fallible, libc-reaching callee finally publishes what it actually reaches.**

## Performance

- **Duration:** 55 min
- **Started:** 2026-09-09T19:02:00Z
- **Completed:** 2026-09-09T19:57:00Z
- **Tasks:** 3 completed
- **Files modified:** 12 (3 created, 9 modified)

## Accomplishments

- `originvalidate.BuildInterface`'s `chainOrder` loop now folds each caller's `Foreign`/`Fails` with every callee's already-committed value via `joinForeignReach`/`joinFails`, strictly before `computeClosureDigest` -- `joinForeignReach` is a true worst-case lattice join (empty is the identity; equal merges; `forbidden` beats `permitted` in either order; two different non-empty allocator names resolve to the new `core.ForeignReachConflict` sentinel, never an arbitrary pick), proven commutative and associative by `TestForeignJoinIsOrderIndependent`.
- 07-REVIEW.md's IN-01 is fixed incidentally, in the same pass: the second pass now resolves the program function from `programFunctionByID` (keyed by ID) instead of indexing `program.Functions` with an index built from `len(summary.Functions)`.
- `corevalidate`'s `chainPeerClosureDigests` grows its own, independently-written join (`peerJoinForeignReach`/`peerJoinFails`) over its own `checkCallGraphAcyclic`-derived `peerPostorder`/`peerAdjacency` -- never `callgraph`'s or `originvalidate`'s -- closing CR-03's coordinated-blindness finding where producer and peer previously reproduced the same local read verbatim.
- `TestForeignClosureJoinPeersIndependent` proves the bilateral claim directly: with the producer's join seamed off the peer still joins (and parity breaks), and with the peer's join seamed off the producer still joins (and parity breaks the other direction).
- `testdata/phase07/call_fallible_foreign_reach.lang` (module `phase07.call_fallible_foreign_reach`) is the standing witness: `main` (no local `ForeignContract`) calls `tracer`, which `try`s a foreign `probe` symbol declared `allocator: "libc_malloc"`, `unwind`/`nonlocal_exit: forbidden`, `fails: ProbeError`. Before this plan `main` published the zero `ForeignReach` and `fails=""`; after, `main` publishes the identical `libc_malloc`/`forbidden`/`forbidden`/`ProbeError` `tracer` itself does.
- Two new controls (`control:summary.foreign_reach_closure_derived`, `control:summary.fails_closure_derived`) are wired into `Phase7RequiredControls()` (24 identifiers), `scripts/verify-phase7.sh`, and `controlsWithRecordedMutationKill`, each mutation-killed on both the producer and the peer sides.
- `PHASE-07-DEBT.md` gains D-07-53 (the join's disclosed granularity: three plain strings, no structured lattice; `Fails` cannot express a union of two error types) and D-07-54 (IN-02/IN-03 deliberately carried, `check`-side paths this gap-closure run does not touch).

## Task Commits

Each task was committed atomically:

1. **Task 1: `Foreign` and `Fails` are closure-derived in the producer** - `bcaf0bb` (feat)
2. **Task 2: `corevalidate` independently re-derives both joined fields** - `175260f` (feat)
3. **Task 3: both new controls wired into all registries, debt recorded** - `1535f1a` (feat)

## Files Created/Modified

- `internal/compiler/core/core.go` - `ForeignReachConflict = "conflict"` const, with doc comment naming its role as the join's declared sentinel
- `internal/compiler/originvalidate/originvalidate.go` - `joinForeignReach`, `joinAllocatorName`, `joinReachPolicy`, `joinFails`; two seams (`foreignClosureJoinSeam`, `failsClosureJoinSeam`) with production-visible `Set...Seam` setters (cross-package visibility); `programFunctionByID` (IN-01 fix); the second pass applies both joins before `computeClosureDigest` and refuses (rather than silently continuing) when a callee ID has no published signature
- `internal/compiler/originvalidate/export_test.go` - net no-op (seam setters moved to the production file for cross-package visibility; see Decisions)
- `internal/compiler/originvalidate/originvalidate_foreign_closure_test.go` (new) - `TestForeignReachIsClosureDerived`, `TestFailsIsClosureDerived`, `TestForeignReachClosureLeafUnperturbed`, `TestForeignJoinIsOrderIndependent`, `TestSecondPassResolvesProgramFunctionByID`, `TestForeignClosureJoinMutationKilled`, `TestFailsClosureJoinMutationKilled`
- `internal/compiler/corevalidate/corevalidate.go` - `peerJoinForeignReach`, `peerJoinAllocatorName`, `peerJoinReachPolicy`, `peerJoinFails`; two peer-side seams (`disableForeignClosureJoinPeerForTest`, `disableFailsClosureJoinPeerForTest`); `chainPeerClosureDigests` applies both joins before `peerComputeClosureDigest`
- `internal/compiler/corevalidate/export_test.go` - `SetDisableForeignClosureJoinPeerForTest`, `SetDisableFailsClosureJoinPeerForTest`
- `internal/compiler/corevalidate/corevalidate_foreign_closure_test.go` (new) - `TestPeerForeignReachIsClosureDerived`, `TestPeerFailsIsClosureDerived`, `TestForeignClosureJoinPeersIndependent`, `TestForeignClosureJoinPeerMutationMatrix`
- `internal/compiler/session/session_phase7.go` - `ControlSummaryForeignReachClosureDerived`, `ControlSummaryFailsClosureDerived` consts; both appended to `Phase7RequiredControls()`
- `internal/compiler/session/session_phase7_test.go` - both controls added to `controlsWithRecordedMutationKill` with a `// 07-12:` comment
- `scripts/verify-phase7.sh` - both new control identifiers added to the phase-07 required-control loop
- `testdata/phase07/call_fallible_foreign_reach.lang` (new) - the CR-03/PVG-02 standing witness
- `.planning/phases/07-calls-signatures-and-call-graph-refusal/PHASE-07-DEBT.md` - D-07-53, D-07-54 added (items: 9 → 11)

## Decisions Made

- The producer-side seam setters (`SetForeignClosureJoinSeam`/`SetFailsClosureJoinSeam`) had to be declared in `originvalidate.go` itself (a production file), not `export_test.go`, because `corevalidate`'s own bilateral-independence test imports `originvalidate` as an ordinary cross-package dependency and Go's build model excludes every `_test.go` file (export_test.go included) from that non-test build. This exactly mirrors 07-11's `check.SetCallArgumentConsumeSeamForTest` precedent -- deliberately minimal, test-only-no-op in production, always restored via its returned closure.
- `joinFails`/`peerJoinFails` deliberately keep the EXISTING (caller-nearest, i.e. accumulator-so-far) value on a two-callee disagreement rather than being a true commutative/associative join like `joinForeignReach` -- the plan's own acceptance criteria name order-independence only for the Foreign join. `Fails` is a single string that cannot express a union of two distinct error types; that imprecision is disclosed as D-07-53, not hidden inside a false lattice shape. In production the fold always runs in sorted callee-ID order (`calleeIDsForClosureDigest`/`v.peerAdjacency` are both sorted), so the published value is still fully deterministic across runs.
- `TestSecondPassResolvesProgramFunctionByID` proves the IN-01 fix by building the SAME three-function set in two different declaration orders and asserting each function's own published Foreign/Fails depends only on its own identity, never on array position. The historical index-mis-association bug is dormant under today's code shape (no `continue` currently skips a function in the first loop), so this is a forward-looking invariant proof rather than a reproduction of a live defect -- exactly the preventive posture 07-REVIEW.md's IN-01 finding asked for.

## Deviations from Plan

None - plan executed exactly as written, aside from the seam-placement adjustment (production file vs export_test.go) documented above under Decisions, which follows an established codebase pattern (07-11) rather than introducing a new one.

## Issues Encountered

- The plan's own verify command `grep -c 'libc_malloc' /tmp/phase07_foreign_reach.json` (Task 1) expects an occurrence count of at least 2, but `lang interface export`'s file output is single-line compact JSON, so `grep -c` reports the LINE count (1), not the substring occurrence count, even though `libc_malloc` genuinely appears twice (once for `tracer`, once for the now-correctly-joined `main`). Confirmed via `grep -o 'libc_malloc' ... | wc -l` = 2 and by inspecting the decoded JSON directly (both `tracer.foreign.allocator` and `main.foreign.allocator` equal `"libc_malloc"`). This is a verify-script artifact of single-line JSON output, not an implementation defect; the substantive claim the command exists to check holds.

## Known Stubs

None.

## Threat Flags

None -- every new surface (the two joins, the four fault-injection seams, the `ForeignReachConflict` sentinel) is inside this plan's own declared `<threat_model>`.

## User Setup Required

None - no external service configuration required.

## Verification Results (re-run live)

- `go build ./...` - clean.
- `go vet ./...` - clean.
- `go test ./... -p 1` - all packages `ok`, zero `FAIL`.
- `go test -race ./...` - all packages `ok`, zero `DATA RACE`, zero `FAIL`.
- `sh scripts/verify-phase7.sh` - exits 0; `phase07.json`'s `lane:kind-exhaustive-dispatch-phase07` lists all 24 controls, all `status: pass`, including both new identifiers.
- `go run ./cmd/lang --json interface export testdata/phase07/call_fallible_foreign_reach.lang` - `main` publishes `foreign: {allocator: "libc_malloc", unwind: "forbidden", nonlocal_exit: "forbidden"}` and `fails: "ProbeError"`, identical to `tracer`'s own published values.
- `go run ./cmd/lang --json check testdata/phase07/call_fallible_foreign_reach.lang` - `status: pass`, exit 0.
- **Ripple, enumerated explicitly (Task 1 Test 8):** among every existing `testdata/phase07` fixture, only `call_fallible_foreign_reach.lang` (new) has a function whose published `Foreign`/`Fails`/`ClosureDigest` changes. `foreign_symbol_shadowing.lang` is the only OTHER fixture declaring a `foreign C` block reachable through a call, but it is a refusing (cyclic) fixture -- `check` refuses it with `core.call_graph_cycle` before `BuildInterface` is ever reached, so its published summary (which never exists) is unaffected. Every other fixture declares no `foreign C` block at all, so their leaf-identity join contribution is the empty value, byte-identical to before.
- `git status --porcelain testdata/` - empty (the fixture was already committed in Task 1's commit; no `*.golden.json` or `*.core.json` touched).
- `git diff --name-only go.mod` - empty; no `go.sum` in this repo (unaffected either way).
- `git diff internal/compiler/check/check.go` - empty; no hunk touches `computeLoanLastUses` (PVG-04 stays Phase 08 scope per D-07-49).

## Next Phase Readiness

- 07-REVIEW.md CR-03 / 07-VERIFICATION.md PVG-02 is closed: `core.FunctionSignature.Foreign`/`.Fails` are genuinely closure-derived, on both the producer and an independent peer, before `ClosureDigest` is computed.
- 07-REVIEW.md IN-01 is closed incidentally, in the same pass.
- All prior-wave gap closures (07-10's peer-consulted `lang check`, 07-11's call-argument consume) remain untouched and green.
- PVG-04 / CR-02 (`check.computeLoanLastUses` has no `"call"` case) remains explicit Phase 08 scope (D-07-49) -- confirmed unmodified.
- The `relay_escort_witness.lang` check/corevalidate divergence (D-03-02) is neither closed, widened, nor narrowed by this plan.
- D-07-53 (join granularity) and D-07-54 (IN-02/IN-03) are new, explicitly disclosed residuals for a future phase to reopen.
- This closes 07-12, the phase's last recorded post-verification gap (07-10 closed PVG-03/CR-04, 07-11 closed PVG-01/CR-01). Phase 07 is now fully verified against 07-VERIFICATION.md and 07-REVIEW.md with no open post-verification gaps. Ready for `/gsd-verify-work 07` and Phase 08 planning.

---
*Phase: 07-calls-signatures-and-call-graph-refusal*
*Completed: 2026-09-09*

## Self-Check: PASSED

- All created/modified files verified present on disk (core.go, originvalidate.go, originvalidate/export_test.go, originvalidate_foreign_closure_test.go, corevalidate.go, corevalidate/export_test.go, corevalidate_foreign_closure_test.go, session_phase7.go, session_phase7_test.go, scripts/verify-phase7.sh, testdata/phase07/call_fallible_foreign_reach.lang, PHASE-07-DEBT.md).
- All three task commits (`bcaf0bb`, `175260f`, `1535f1a`) confirmed present in `git log --oneline --all`.
- `go build ./...`, `go vet ./...`, `go test ./... -p 1`, `go test -race ./...` all green; `sh scripts/verify-phase7.sh` exits 0 with 24/24 controls passing.
- `git status --porcelain testdata/` empty; `git diff --name-only go.mod` empty.
