---
phase: 03-borrowed-views-and-cfg-lifetimes
plan: "02"
subsystem: compiler-frontend-backend
tags: [ownership, borrowing, exclusive-loans, corevalidate, cgen, interp, native, c17]

requires:
  - phase: 03-borrowed-views-and-cfg-lifetimes
    provides: "03-01's first real CFG (core.Block/Edge/LoanEndpoint, checkBranch, corevalidate.matchBranch/replayBlocks, interp.runBranchArm, cgen.emitBranch) — extended, not recreated, by this plan's exclusive-loan work"
provides:
  - "`borrow mut` surface syntax and `core.OpBorrowExclusive`, wired through check, corevalidate, interp, cgen, and native (five consumers, not four — see deviations)"
  - "The five-row shared/exclusive loan conflict matrix (shared+shared accept, shared+exclusive/exclusive+exclusive/exclusive+move reject, sequential accept), decided by liveness overlap in check.go and independently re-derived in corevalidate.go"
  - "check.go's execution-admission gate closing D-02-09/D-07: a linear function whose parameter shape has no native lowering (anything but Byte/Buffer) is refused at check time with a causal diagnostic instead of dying spanless downstream"
affects: [03-03-cfg-edge-liveness, 03-04-validator-liveness, 03-05-path-oracle, 03-06-public-origins]

actuals:
  tokens: 22143
  tasks: 3
  commits: 4

tech-stack:
  added: []
  patterns:
    - "Per-owner running high-water-mark maps (ownerLiveSharedUntil/ownerLiveExclusiveUntil) as corevalidate's independent conflict-matrix derivation — materially different from check.go's per-owner active-loan-set mechanism (D-12)"
    - "Reborrow (`let x = borrow y`) as the canonical way to extend a noncopyable (Buffer) loan's last use without an implicit copy, since a borrowed place carries the same TypeID as its owner and Buffer denies copy"

key-files:
  created:
    - testdata/phase3/shared_shared_accept.lang
    - testdata/phase3/shared_exclusive_reject.lang
    - testdata/phase3/exclusive_exclusive_reject.lang
    - testdata/phase3/exclusive_move_reject.lang
    - testdata/phase3/sequential_shared_then_exclusive_accept.lang
    - internal/compiler/check/check_exclusive_test.go
    - internal/compiler/check/check_unexecutable_test.go
    - internal/compiler/corevalidate/corevalidate_exclusive_test.go
    - internal/compiler/session/session_exclusive_test.go
    - internal/compiler/session/session_borrow_conflict_test.go
  modified:
    - internal/compiler/syntax/lexer.go (no functional change — normalization lives in parser.go's normalizeOwnershipTokens; lexer keyword table untouched)
    - internal/compiler/syntax/token.go
    - internal/compiler/syntax/parser.go
    - internal/compiler/syntax/format.go
    - internal/compiler/syntax/syntax_test.go
    - internal/compiler/core/core.go
    - internal/compiler/check/check.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/interp/interp.go
    - internal/compiler/cgen/cgen.go
    - internal/compiler/native/native.go
    - internal/compiler/session/session_test.go
    - internal/compiler/testsupport/cli_test.go

key-decisions:
  - "`borrow mut` is the surface spelling for an exclusive loan; `mut` is normalized from identifier to a keyword token the same contextual way `let`/`take`/`borrow` already are, not added to the lexer's permanent keyword table."
  - "An exclusive loan is gated on the same `AbilityShare` requirement a shared loan is gated on (not a new ability), per the plan's explicit instruction — this gate is source-unreachable today (no declared type withholds share), documented as such beside both borrow cases per D-10, exactly mirroring the pre-existing shared-borrow gate's own note."
  - "The five-row conflict matrix is decided in check.go by a new `conflictingLoan`/`borrowConflictDiagnostic` pair applied identically to both `analyzeStraightLine` and `analyzeArmBody` (deliberately duplicated, not shared, matching the file's existing straight-line/arm-body duplication convention)."
  - "corevalidate's independent re-derivation uses running per-owner high-water-mark maps updated in program order, not a whole-function aggregate (which the pre-existing move-while-borrowed check safely uses only because a place can never be re-borrowed after being moved) — this avoids a later-created loan spuriously blocking an earlier one."
  - "D-02-09/D-07 closure (Task 3) lives in `checkLinear` only: match-based functions (bare-arm or branch) always execute via `emitMatch`/`emitBranch`'s enum-based lowering regardless of the scrutinee's declared shape, so the gate is scoped to the straight-line linear path where `cgen.linearInput` is the actual native-lowering bottleneck (Byte/Buffer only)."

requirements-completed: [OWN-03]

coverage:
  - id: D1
    description: "`borrow mut x` parses, checks, validates, interprets, and lowers to C as an exclusive loan, appearing in the execution trace at both -O0 and -O3"
    requirement: OWN-03
    verification:
      - {kind: unit, ref: "internal/compiler/syntax#TestExclusiveBorrowRoundTrips", status: pass}
      - {kind: unit, ref: "internal/compiler/check#TestExclusiveBorrowLowersToCore", status: pass}
      - {kind: unit, ref: "internal/compiler/corevalidate#TestExclusiveBorrowAuthorizedIndependently", status: pass}
      - {kind: integration, ref: "internal/compiler/session#TestExclusiveBorrowInterpreterNative", status: pass}
      - {kind: manual_procedural, ref: "shipped ./cmd/lang binary: format --check, check, run --engine=interpreter, run --engine=native on a hand-written non-corpus program (module owned.handwritten_exclusive)", status: pass}
    human_judgment: true
    rationale: "The shipped-binary drive (D-11) was performed manually this session against a hand-written program outside any corpus; it is verified session evidence, not captured as an automated regression test."
  - id: D2
    description: "The five-row conflict matrix is decided by overlapping liveness (not lexical order) and independently re-decided at the second admission layer"
    requirement: OWN-03
    verification:
      - {kind: unit, ref: "internal/compiler/session#TestSharedSharedOverlapAccepted", status: pass}
      - {kind: unit, ref: "internal/compiler/session#TestSequentialLoansAccepted", status: pass}
      - {kind: unit, ref: "internal/compiler/session#TestBorrowConflictMatrix", status: pass}
      - {kind: unit, ref: "internal/compiler/session#TestBorrowConflictCauseChain", status: pass}
      - {kind: unit, ref: "internal/compiler/session#TestExclusiveMoveRejected", status: pass}
      - {kind: unit, ref: "internal/compiler/corevalidate#TestCorevalidateIndependentlyRejectsBorrowConflict", status: pass}
    human_judgment: false
  - id: D3
    description: "corevalidate.replay still rejects an unknown operation kind fail-closed after the new dispatch arm is added"
    requirement: OWN-03
    verification:
      - {kind: unit, ref: "internal/compiler/corevalidate#TestUnknownOperationStillRejected", status: pass}
    human_judgment: false
  - id: D4
    description: "A Box/Pair (or any non-Byte/Buffer) linear function is rejected at check time with a causal, repair-bearing diagnostic instead of dying spanless at exit 3; ability derivation for that shape still runs and produces facts"
    requirement: OWN-03
    verification:
      - {kind: unit, ref: "internal/compiler/check#TestUnexecutableShapeRejectedWithSpan", status: pass}
      - {kind: unit, ref: "internal/compiler/check#TestAbilityFactsSurviveExecutionRejection", status: pass}
      - {kind: unit, ref: "internal/compiler/session#TestSourceBoxPairAbilityFacts", status: pass}
      - {kind: e2e, ref: "internal/compiler/testsupport#TestHumanJSONMixedDiagnosticVersionParity", status: pass}
    human_judgment: false

duration: 70min
completed: 2026-09-04
status: complete
---

# Phase 03 Plan 02: Exclusive Loans, the Conflict Matrix, and the Unexecutable-Shape Gate Summary

**`borrow mut` exclusive loans travel source to native through five independently updated consumers (check, corevalidate, interp, cgen, and — a D-12a gap this plan's own audit surfaced — native's execution-document validator), the five-row shared/exclusive conflict matrix is decided by liveness overlap and independently re-derived at the validator, and `check` now refuses Box/Pair-shaped linear functions with a causal diagnostic instead of letting every engine die spanless at exit 3.**

## Performance

- **Duration:** ~70 min
- **Started:** 2026-09-04T03:20:00Z (approx.)
- **Completed:** 2026-09-04T04:30:00Z (approx.)
- **Tasks:** 3 completed
- **Files modified:** 26 (10 created, 16 modified)

## Accomplishments

- Added `borrow mut` surface syntax: `TokenMut` normalized contextually
  (like `let`/`take`/`borrow`), the parser accepts an optional `mut` after
  `borrow` producing `RHS.Kind == "borrow_mut"`, and the formatter emits it
  canonically with the same "newline after the RHS identifier" rule the
  existing `take`/`borrow` cases use.
- Added `core.OpBorrowExclusive`, lowered in `check.go`'s `checkLinear`
  (`analyzeStraightLine`) and `checkBranch` (`analyzeArmBody`) — both
  deliberately duplicated, not shared — gated on `AbilityShare` exactly like
  a shared borrow.
- Implemented the five-row conflict matrix in `check.go` via
  `conflictingLoan`/`borrowConflictDiagnostic`: shared+shared never
  conflicts; shared+exclusive, exclusive+exclusive, and exclusive+move all
  reject (`ownership.borrow_conflict` for the first two, the pre-existing
  `ownership.move_while_borrowed` for the third, unchanged); a shared loan
  whose last use precedes an exclusive borrow is accepted, proving the rule
  is liveness-based, not lexical.
- Independently re-derived the same matrix in `corevalidate.go`'s
  `replayStraightLine`/`replayBlocks`, using running per-owner high-water-mark
  maps (`ownerLiveSharedUntil`/`ownerLiveExclusiveUntil`) built in program
  order — a materially different mechanism from the checker's active-loan-set
  approach (D-12). The pre-existing `default:` fail-closed arm still rejects
  any operation kind it has not been explicitly taught.
- Wired `interp.go` (`runLinear`/`runBranchArm`) and `cgen.go`
  (`emitLinear`/`emitBranchOperations`) to emit/lower `OpBorrowExclusive`
  identically to `OpBorrowShared`, with its own event kind
  `value.borrowed_exclusive`, so the operation appears in both engines'
  traces at `-O0` and `-O3`.
- Closed D-02-09/D-07: `checkLinear` now refuses a straight-line linear
  function whose parameter shape is not `Byte`/`Buffer` (i.e. `Box`, `Pair`,
  or any other constructor `cgen.linearInput` cannot lower) with a
  span-bearing, repair-bearing `check.unexecutable_shape` diagnostic naming
  the offending type and constructor — ability derivation for the shape
  still runs and produces its facts first, so the rejection is purely about
  execution admission. `testdata/phase2/ability_shapes.lang` (three
  functions: `Box<Byte>`, `Box<Buffer>`, `Pair<Byte,Buffer>`) is the
  fixture this closes; no engine or evidence build can reach a spanless exit
  3 for it any more because `check` now refuses it first.
- All five Q3 fixtures under `testdata/phase3/` are canonical
  (`Format(Parse(src)) == src`), comment-led (each states the law it
  distinguishes), and reachable from real source — no synthetic type fact
  was needed for any of the five rows.
- Proved the slice through the shipped `./cmd/lang` binary (`format --check`,
  `check`, `run --engine=interpreter`, `run --engine=native`) on a
  hand-written, non-corpus exclusive-borrow program (D-11), with identical
  execution traces (including `value.borrowed_exclusive`) at both
  optimization levels.
- Mutation-killed corevalidate's independent conflict re-derivation per
  D-09: temporarily neutralized both `core.borrow_conflict` guards in
  `replayStraightLine`/`replayBlocks`, confirmed
  `TestCorevalidateIndependentlyRejectsBorrowConflict` failed
  ("validator admitted an exclusive loan overlapping a shared loan on the
  same owner"), then restored the guards with zero resulting diff.

## Task Commits

1. **Task 03-02-01: End-to-end exclusive loan across all four operation-kind sites** — `2059f87` (feat) — this commit also carries Task 03-02-02's conflict-matrix production code and Task 03-02-03's unexecutable-shape gate, since both live in the same `check.go` functions the exclusive-loan work touches; see Deviations.
2. **Task 03-02-02: Five-row conflict matrix fixtures and falsifiers** — `0f60a95` (test)
3. **Additional coverage: mutation-kill corevalidate's conflict re-derivation** — `0ad2a94` (test)
4. **Task 03-02-03: Unexecutable-shape rejection tests** — `dadfbad` (test)

**Plan metadata:** (this commit)

_Note: this plan's production code (Task 1's exclusive-loan wiring, Task 2's
conflict matrix, and Task 3's unexecutable-shape gate) all landed in
`internal/compiler/check/check.go` in the same continuous editing session,
so it is committed as one unit (`2059f87`) rather than split three ways —
splitting a single file's interleaved hunks across three commits after the
fact would have required reconstructing intermediate states with no
functional benefit. Each task's own tests are committed separately and map
exactly to the task boundaries._

## Files Created/Modified

- `internal/compiler/syntax/token.go` — `TokenMut`
- `internal/compiler/syntax/parser.go` — `borrow mut` parsing, `mut` token normalization
- `internal/compiler/syntax/format.go` — canonical `mut ` formatting
- `internal/compiler/core/core.go` — `OpBorrowExclusive`
- `internal/compiler/check/check.go` — exclusive-loan lowering, conflict matrix, unexecutable-shape gate
- `internal/compiler/corevalidate/corevalidate.go` — `OpBorrowExclusive` dispatch arm, independent conflict re-derivation
- `internal/compiler/interp/interp.go` — `OpBorrowExclusive` event emission (both engines' dispatch)
- `internal/compiler/cgen/cgen.go` — `OpBorrowExclusive` C lowering (both `emitLinear` and `emitBranchOperations`)
- `internal/compiler/native/native.go` — execution-document event-kind whitelist extended (see Deviations)
- `testdata/phase3/{shared_shared_accept,shared_exclusive_reject,exclusive_exclusive_reject,exclusive_move_reject,sequential_shared_then_exclusive_accept}.lang` — the five Q3 fixtures
- New test files: `check_exclusive_test.go`, `check_unexecutable_test.go`, `corevalidate_exclusive_test.go`, `session_exclusive_test.go`, `session_borrow_conflict_test.go`
- `internal/compiler/session/session_test.go`, `internal/compiler/testsupport/cli_test.go` — updated for the D-02-09/D-07 behavior change

## Decisions Made

See `key-decisions` in frontmatter.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] `native.go`'s execution-document event-kind whitelist is a fifth D-12a consumer**
- **Found during:** Task 03-02-01, first end-to-end native run of the exclusive-borrow fixture
- **Issue:** D-12a's text names four consumers (check, corevalidate, interp, cgen). The generated C's execution-document JSON is decoded and re-validated by `native.go`'s own event-kind whitelist (`"value.copied", "value.transferred", "value.borrowed"`) before it is trusted as a real execution — an operation kind cgen correctly emits into that JSON is still rejected as `native.invalid_execution` ("unknown execution event kind") if this fifth site is not updated. Missing it would have silently turned a correct exclusive-borrow trace into an operational failure rather than a semantic one, exactly the self-confirming shape D-12a exists to prevent.
- **Fix:** Added `"value.borrowed_exclusive"` to the whitelist at `native.go`'s event-kind switch.
- **Files modified:** `internal/compiler/native/native.go`
- **Verification:** `TestExclusiveBorrowInterpreterNative` (O0/O3 differential) failed with `native.invalid_execution` before this fix, passes after.
- **Committed in:** `2059f87` (part of Task 1)

**2. [Rule 1 - Bug] Pre-existing `TestSourceBoxPairAbilityFacts` asserted the exact behavior D-02-09/D-07 closes**
- **Found during:** Task 03-02-03, first full-suite run after adding the unexecutable-shape gate
- **Issue:** This Phase 2 test asserted `testdata/phase2/ability_shapes.lang` checked with zero diagnostics and its three functions' ability facts landed in the returned `core.Program`. That is precisely the taxonomy smell D-02-09/D-07 requires this plan to close — the test's premise and the plan's required behavior are mutually exclusive.
- **Fix:** Rewrote the test to assert the new behavior (three `check.unexecutable_shape` diagnostics, no functions admitted into the checked core), and added `TestAbilityFactsSurviveExecutionRejection` (unit-testing `ability.Derive` directly for the same three shapes) to preserve the original test's actual load-bearing claim — that ability derivation for these shapes still works — decoupled from check-time admission.
- **Files modified:** `internal/compiler/session/session_test.go`
- **Verification:** Both tests pass; full suite green.
- **Committed in:** `dadfbad` (Task 3)

---

**Total deviations:** 2 auto-fixed (1 blocking, 1 bug). Both were necessary consequences of implementing this plan's own required behavior — no scope creep.

## Issues Encountered

- Several fixture drafts for the conflict-matrix accept cases initially used
  an implicit copy (`let observed = first`) to extend a shared loan's last
  use before creating the next loan. Since a borrowed place carries the same
  `TypeID` as its `Buffer` owner (no separate view type exists yet), and
  `Buffer` denies `copy`, this tripped `ownership.transfer_requires_take`
  instead of exercising the intended conflict/accept path. Resolved by using
  a reborrow (`let reviewed = borrow first`) instead, which only requires
  `AbilityShare` (which `Buffer` grants) — documented as a key-decision.

## User Setup Required

None — no external service configuration required.

## Mutation-Kill / Non-Regression Evidence (recorded verbatim per D-09)

- `git diff 3da142a..HEAD -- testdata/phase1/` is empty.
- `git diff 3da142a..HEAD -- testdata/phase2/owned_transfer.golden.c testdata/phase2/evidence.golden.json` is empty — no Phase 2 golden moved.
- `sh scripts/verify-phase2.sh` exits 0 and reports all nine Phase 2 control
  IDs and both Phase 1 controls through a freshly built binary.
- `env GOCACHE=/tmp/ai-lang-phase3-cache go test ./...`,
  `go test -race ./...`, and `go vet ./...` all pass with zero findings.
- **Mutation-kill (corevalidate's independent conflict re-derivation):**
  with both `"core.borrow_conflict"` `v.check` calls in
  `replayStraightLine`/`replayBlocks` temporarily replaced with no-ops,
  `TestCorevalidateIndependentlyRejectsBorrowConflict` failed with:
  `corevalidate_exclusive_test.go:89: validator admitted an exclusive loan
  overlapping a shared loan on the same owner`. Restoring the guards made it
  pass again with zero resulting diff against the committed file.
- Shipped-binary drive (D-11): `format --check`, `check`,
  `run --engine=interpreter`, `run --engine=native` all pass on a
  hand-written, non-corpus program (`module owned.handwritten_exclusive`,
  `borrow mut` followed by a reborrow) — full interpreter/O0/O3 agreement,
  including the `value.borrowed_exclusive` event.

## Known Stubs

None.

## Next Phase Readiness

- Exclusive loans and the full conflict matrix exist end to end, independently
  re-derived at both admission layers, ready for 03-03's edge-specific CFG
  liveness to place loan endpoints on branch edges rather than uniformly at
  the join.
- The D-02-09/D-07 taxonomy smell is closed: every program `lang check`
  accepts is now executable by every engine, an invariant 03-03 through
  03-07 can rely on without re-checking it.
- No blockers for 03-03.

---
*Phase: 03-borrowed-views-and-cfg-lifetimes*
*Plan: 02*
*Completed: 2026-09-04*

## Self-Check: PASSED
