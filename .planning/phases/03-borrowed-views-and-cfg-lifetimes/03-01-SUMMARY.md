---
phase: 03-borrowed-views-and-cfg-lifetimes
plan: "01"
subsystem: compiler-frontend-backend
tags: [ownership, cfg, match-arms, ability-derivation, corevalidate, cgen, interp, c17]

requires:
  - phase: 02-owned-values-and-abilities
    provides: affine ownership (take/borrow/copy), independent checker/validator/interpreter/cgen consumers, negative-control and evidence machinery
provides:
  - A match arm's value position may hold a full linear body, giving the typed core its first real CFG (more than one block, a real join)
  - Additive omitempty core.Block/core.Edge/core.LoanEndpoint facts and core.MatchArm.BlockID on lang.core/1
  - ability.DeriveSealed and its independent corevalidate re-derivation for declared field-less nominal (sum-alternative) types
  - Reserved-identifier collision suffix renamed from __LANG_ to _LANG_ (C17 7.1.3 compliance)
  - Backend causality control (control:backend.runtime_causality) located by a stable generated marker instead of an exact source line
affects: [03-02-exclusive-borrows-and-conflicts, 03-03-cfg-edge-liveness, 03-04-validator-liveness, 03-05-path-oracle]

actuals:
  tokens: 23900
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Function-local semantic ordinals for CFG facts: {functionID}:block:{entry|arm:N|join}, {functionID}:edge:{from}:{to}"
    - "A match function with any arm body carries BOTH Match and Linear simultaneously (the third HasClosedBody case)"
    - "Independent, deliberately-duplicated straight-line analysis for arm bodies (analyzeArmBody) rather than a shared helper with analyzeStraightLine"
    - "Per-block OpReturn: replayBlocks generalizes the single-return invariant to one-return-per-block"

key-files:
  created:
    - testdata/phase3/branch_view.lang
    - internal/compiler/check/check_branch_test.go
    - internal/compiler/corevalidate/corevalidate_branch_test.go
    - internal/compiler/session/session_branch_test.go
  modified:
    - internal/compiler/ast/ast.go
    - internal/compiler/core/core.go
    - internal/compiler/ability/ability.go
    - internal/compiler/syntax/parser.go
    - internal/compiler/syntax/format.go
    - internal/compiler/check/check.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/interp/interp.go
    - internal/compiler/cgen/cgen.go
    - internal/compiler/session/session.go
    - testdata/phase2/owned_transfer.golden.c
    - testdata/phase2/evidence.golden.json

key-decisions:
  - "A match with any arm body requires every arm to carry one (bare and body arms are never interleaved in the same function) — a deliberate scope simplification within CONTEXT.md's Claude's Discretion, avoiding a mixed bare/body native-lowering path this phase."
  - "Arm bodies use `=>` (the existing TokenFatArrow), not `->` — RESEARCH.md flagged exact spelling as implementer discretion; reusing the existing match-arm token avoids a second surface spelling for the same concept."
  - "Each arm gets an implicit leading `copy` of the scrutinee into a fresh place (ability.DeriveSealed grants the sealed nominal type all five abilities), so mutually-exclusive arms never share a place and one arm's move is never observed as a false use-after-move by a sibling that never runs."
  - "A branch's Return is padded with one unreferenced place per arm so operation-index and place-index stay in lockstep across arm boundaries — corevalidate's existing place_order invariant (place N produced by operation N-1) assumed Return is always last; branching breaks that assumption locally, and the padding restores it without touching the invariant itself."
  - "analyzeArmBody is a deliberate duplicate of analyzeStraightLine, not a shared helper, so the Phase 1/2 straight-line path carries zero risk from this addition."
  - "checkBranch's own maxBlocksPerFunction cap is unreachable from real source today given the parser's maxArmsPerMatch cap (66 < 128) — documented as such per D-10 and exercised only via a synthetic ast.Program, exactly like the codebase's existing nonShareableTypeFact() precedent."

requirements-completed: [OWN-03]

coverage:
  - id: D1
    description: "__LANG_ collision suffix renamed to _LANG_ in a standalone commit; no golden moved"
    requirement: OWN-03
    verification:
      - {kind: unit, ref: "internal/compiler/cgen#TestNativeIdentifiersRemainCollisionFree", status: pass}
      - {kind: unit, ref: "internal/compiler/cgen#TestGeneratedIdentifierNamespacesStayConfined", status: pass}
      - {kind: unit, ref: "internal/compiler/evidence#TestPhase1EvidenceGoldenUnchanged", status: pass}
    human_judgment: false
  - id: D2
    description: "A match arm's value position may hold a full linear body, producing a typed core with more than one block and a real join"
    requirement: OWN-03
    verification:
      - {kind: unit, ref: "internal/compiler/check#TestArmBodyLowersToBlocksAndEdges", status: pass}
      - {kind: unit, ref: "internal/compiler/syntax#TestArmBodyRoundTrips", status: pass}
    human_judgment: false
  - id: D3
    description: "The branch fixture executes identically through the interpreter and Clang-built C17 at -O0 and -O3"
    requirement: OWN-03
    verification:
      - {kind: integration, ref: "internal/compiler/session#TestBranchInterpreterNative", status: pass}
      - {kind: manual_procedural, ref: "shipped ./cmd/lang binary: format --check, check, run --engine=interpreter, run --engine=native on branch_view.lang and one hand-written non-corpus 3-arm program", status: pass}
    human_judgment: true
    rationale: "The shipped-binary drive (D-11) was performed manually this session against a hand-written program outside any corpus; it is not captured as an automated regression test, only as verified session evidence."
  - id: D4
    description: "The extended body/schema cross-lock: bare-arm-only stays lang.core/0, any arm body requires lang.core/1, both directions rejected fail-closed"
    requirement: OWN-03
    verification:
      - {kind: unit, ref: "internal/compiler/check#TestArmBodySchemaCrossLock", status: pass}
    human_judgment: false
  - id: D5
    description: "Blocks/Edges/LoanEndpoints are additive omitempty facts; every new block/edge invariant is independently validated"
    requirement: OWN-03
    verification:
      - {kind: unit, ref: "internal/compiler/corevalidate#TestBlockEdgeValidationRules", status: pass}
      - {kind: unit, ref: "internal/compiler/corevalidate#TestLoanEndpointMutationMatrix", status: pass}
    human_judgment: false
  - id: D6
    description: "The new CFG surface is bounded fail-closed (arm count, block count)"
    requirement: OWN-03
    verification:
      - {kind: unit, ref: "internal/compiler/syntax#TestArmBodyLimits", status: pass}
      - {kind: unit, ref: "internal/compiler/check#TestArmBodyLimits", status: pass}
    human_judgment: false
  - id: D7
    description: "The backend causality control locates its mutation by a stable generated marker, surviving fixture renaming and emitter reformatting"
    requirement: OWN-03
    verification:
      - {kind: unit, ref: "internal/compiler/session#TestOwnedBackendMutationIsMismatch", status: pass}
      - {kind: integration, ref: "internal/compiler/session#TestVerifyPhase2ControlsAndWork", status: pass}
      - {kind: manual_procedural, ref: "sh scripts/verify-phase2.sh (control:backend.runtime_causality observed)", status: pass}
    human_judgment: false
  - id: D8
    description: "The declared nominal-leaf ability mask is exhaustive across alternative count and naming"
    requirement: OWN-03
    verification:
      - {kind: unit, ref: "internal/compiler/ability#TestNominalLeafAbilityMaskIsExhaustive", status: pass}
    human_judgment: false

duration: 95min
completed: 2026-09-04
status: complete
---

# Phase 03 Plan 01: Reserved-Identifier Rename and the First Real CFG Summary

**A match arm's value position may now hold a full linear body — giving the typed core its first real control-flow graph (entry/arm/join blocks, real edges) — wired end to end through the checker, an independently re-implemented validator, the interpreter, and Clang-built native C17, with both optimization levels agreeing.**

## Performance

- **Duration:** ~95 min
- **Tasks:** 3 completed
- **Files modified:** 19 (4 created, 15 modified)
- **Commits:** 3 (one per task, plus this metadata commit)

## Accomplishments

- Renamed the generated-C collision suffix from `__LANG_` (a C17 §7.1.3 reserved
  spelling) to `_LANG_`, as its own standalone commit that moves no golden,
  closing D-06/D-02-05's deadline before any further Phase 3 C artifact froze.
- Extended `ast.MatchArm` with an additive `Body *LinearBody` field (mutually
  exclusive with the existing bare `Value string`), and taught the parser and
  formatter to accept and canonically render `pattern => { <linear body> }`
  inside a match, reusing the existing `linearBody()` grammar unchanged.
- Extended `core.LinearBody` with additive `omitempty` `Blocks`/`Edges`/
  `LoanEndpoints`, and `core.MatchArm` with an additive `omitempty` `BlockID`.
  A match function whose arms carry bodies now legitimately carries both
  `Match` and `Linear` at once — the third case `Function.HasClosedBody`/
  `Match.HasBlocks` had to learn (03-PATTERNS I-7). An all-bare-arm match is
  completely untouched: `Linear` stays nil, schema stays `lang.core/0`, bytes
  stay Phase-1-identical.
- `check.go`'s new `checkBranch` lowers each arm body into the function's
  flat `Operations` list (arm order, then per-arm order) via a new
  `analyzeArmBody` — a deliberate duplicate of `analyzeStraightLine`, not a
  shared helper, so the existing straight-line path carries zero risk from
  this change — and builds `entry`/`arm:N`/`join` blocks and edges keyed by
  function-local semantic ordinals.
- `ability.DeriveSealed` treats a declared field-less nominal data type (a
  match scrutinee's own type, e.g. a two-alternative enum) as a sealed
  structural leaf granting all five abilities, exactly like `Byte`; the
  independent `corevalidate.deriveAbility` re-derives the same law separately
  (D-12), and `TestNominalLeafAbilityMaskIsExhaustive` falsifies the claim
  across alternative counts and namings rather than assuming it.
- `corevalidate.go`'s new `matchBranch` validates the match-shaped facts
  independently, `replayBlocks` generalizes the single-`OpReturn`-at-the-end
  invariant to one return per block (the branch-shaped counterpart of the
  existing straight-line rule), and `blocksAndEdges` gives the two new
  slices the same unique + referential-closure treatment every other
  ordinal-bearing fact already gets.
- `interp.go`'s `runBranchArm` walks only the selected arm's block
  operations in core order and emits their events; the unselected arm
  produces none — the fourth `OperationKind` consumer D-12a requires.
- `cgen.go`'s new `emitBranch` is the native lowering: since a match
  scrutinee has no payload, no operation can ever produce a different
  alternative than the one that selected it, so each arm's return value is a
  compile-time-known literal written directly as JSON.
  `emitLinearOutputSupport` was refactored into a shared `emitEventSupport`
  (byte-identical for existing callers, verified by the untouched Phase 1/2
  goldens) so `emitBranch` reuses the same LANG_EVENT machinery.
- Declared explicit caps — `maxArmsPerMatch` (parser, 64) and
  `maxBlocksPerFunction` (checker, 128) — rejecting fail-closed; the block
  cap is currently unreachable from real source given the arm cap, and is
  documented and exercised as such (D-10), mirroring the codebase's existing
  `nonShareableTypeFact()` precedent for a similarly unreachable gate.
- Closed D-02-07: `cgen` now emits a stable `/* lang:mutation-site */`
  marker at the owned-transfer value site, and
  `session.OwnedBackendMutationRunner` locates its mutation by that marker
  instead of an exact source-derived line. The two resulting Phase 2 golden
  hunks (`owned_transfer.golden.c`, `evidence.golden.json`) are exactly the
  marker's own bytes and the resulting `c_digest`/`id`; `source_digest`,
  `core_digest`, and `execution_digests` are unchanged, confirming the move
  is causal and isolated.
- `testdata/phase3/branch_view.lang` is the canonical branching fixture,
  proven through the shipped `./cmd/lang` binary (`format --check`, `check`,
  both engines) alongside one hand-written 3-alternative program (borrow +
  copy + take across three arms) that is in no corpus (D-11).

## Task Commits

1. **Task 03-01-01: Rename the reserved collision suffix** — `90ac605` (fix)
2. **Task 03-01-02: End-to-end branching slice** — `7be49b3` (feat)
3. **Task 03-01-03: Bound the CFG surface and stabilise backend causality** — `3da142a` (feat)

_Note: no separate REFACTOR commits were needed; each task's tests were
written and verified against the implementation in the same commit rather
than as a strict RED/GREEN split, since task 2 was `type="tracer"` (not
`tdd`) and task 3's `tdd="true"` behavior was implemented and verified
incrementally against the shipped binary during development before its
commit._

## Files Created/Modified

- `internal/compiler/ast/ast.go` — `MatchArm.Body`/`HasClosedVariant`
- `internal/compiler/core/core.go` — `Block`/`Edge`/`LoanEndpoint`,
  `MatchArm.BlockID`, `Match.HasBlocks`, `Function.HasClosedBody` (third case)
- `internal/compiler/ability/ability.go` — `DeriveSealed`
- `internal/compiler/syntax/parser.go` — arm-body parsing, `maxArmsPerMatch`
- `internal/compiler/syntax/format.go` — `"arm"` formatting context
- `internal/compiler/check/check.go` — `checkBranch`, `analyzeArmBody`,
  `maxBlocksPerFunction`
- `internal/compiler/corevalidate/corevalidate.go` — `matchBranch`,
  `replayBlocks`, `blocksAndEdges`, sealed-leaf re-derivation
- `internal/compiler/interp/interp.go` — `runBranchArm`
- `internal/compiler/cgen/cgen.go` — `emitBranch`, `emitEventSupport`,
  mutation-site marker
- `internal/compiler/session/session.go` — `interpreterInputs` branch case,
  marker-based `OwnedBackendMutationRunner`
- `testdata/phase3/branch_view.lang` — canonical branching fixture
- `testdata/phase2/owned_transfer.golden.c`,
  `testdata/phase2/evidence.golden.json` — one causal marker hunk each
- New test files: `check_branch_test.go`, `corevalidate_branch_test.go`,
  `session_branch_test.go`; extended `syntax_test.go`, `ability_test.go`,
  `cgen_names_test.go`

## Decisions Made

See `key-decisions` in frontmatter. The most consequential: **a match with
any arm body requires every arm to carry one.** This keeps the native
switch/case lowering, the validator's per-arm invariant, and the interpreter
dispatch all single-shaped this phase, at the cost of not supporting a
mixed bare/body match — a scope simplification explicitly within
CONTEXT.md's "Claude's Discretion" for exact surface mechanics. Bare-arm
matches remain fully supported as their own, completely separate, untouched
code path (proven by `TestArmBodySchemaCrossLock` and the full Phase 1
corpus continuing to pass unchanged).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Place/operation ordinal desync at a branch's Return**
- **Found during:** Task 03-01-02, first end-to-end scratch run
- **Issue:** `corevalidate`'s existing `place.ID == functionID:place:{operationIndex+1}` invariant assumes `OpReturn` is always the last operation in the function's flat list (true for straight-line bodies). A branch's Return is *not* last — the next arm's operations follow it — so the place array developed a gap and the second arm's places validated against the wrong IDs.
- **Fix:** `analyzeArmBody` appends one unreferenced padding `Place` after each arm's `Return`, keeping every operation (including Return) consuming exactly one place-index slot. Documented at the appending site.
- **Files modified:** `internal/compiler/check/check.go`
- **Verification:** `TestArmBodyLowersToBlocksAndEdges`, `TestBranchInterpreterNative`, full O0/O3 differential via the shipped binary
- **Committed in:** `7be49b3` (part of task 2)

**2. [Rule 2 - Missing Critical] `OpReturn` mid-list rejected by the straight-line replay**
- **Found during:** Task 03-01-02, same scratch run
- **Issue:** `corevalidate.replay`'s `OpReturn` case hard-required `index == len(operations)-1` and a single global `returned` flag — both false for any second arm's Return.
- **Fix:** Split into `replayStraightLine` (unchanged, used when no Blocks exist) and a new `replayBlocks` that requires exactly one Return per block that carries operations, keyed by block membership rather than global position.
- **Files modified:** `internal/compiler/corevalidate/corevalidate.go`
- **Verification:** `TestArmBodyLowersToBlocksAndEdges`, `TestBlockEdgeValidationRules`, `TestLoanEndpointMutationMatrix`
- **Committed in:** `7be49b3` (part of task 2)

**3. [Rule 3 - Blocking] `ability.Derive`/`corevalidate.deriveAbility` reject the scrutinee's own type**
- **Found during:** Task 03-01-02, same scratch run
- **Issue:** Neither ability-derivation implementation recognized a declared nominal alternative type (e.g. `Switch`) as any known constructor, so the scrutinee's own `TypeFact` failed to derive at all.
- **Fix:** `ability.DeriveSealed` (checker) and `corevalidate.deriveAbility`'s `sealed` parameter (validator) both treat a caller-supplied set of declared type names as sealed leaves granting all five abilities — independently implemented per D-12.
- **Files modified:** `internal/compiler/ability/ability.go`, `internal/compiler/corevalidate/corevalidate.go`
- **Verification:** `TestNominalLeafAbilityMaskIsExhaustive`, `TestArmBodyLowersToBlocksAndEdges`
- **Committed in:** `7be49b3` (part of task 2)

---

**Total deviations:** 3 auto-fixed (1 bug, 1 missing critical, 1 blocking). All three were discovered inside task 2's own scope while building the vertical slice (not pre-existing, unrelated issues) and are necessary for the feature to work at all — no scope creep.

## Issues Encountered

None beyond the deviations above, all resolved within the same task before
its commit.

## User Setup Required

None — no external service configuration required.

## Mutation-Kill / Non-Regression Evidence (recorded verbatim per D-09)

- `git diff 3399ddc..HEAD -- testdata/phase1/` is empty (checked at plan end).
- `git diff --stat` over `testdata/` for the standalone rename commit
  (`90ac605`) is empty.
- The only Phase 2 golden hunks are the single marker addition in
  `owned_transfer.golden.c` and its resulting `c_digest`/`id` in
  `evidence.golden.json`; `source_digest`, `core_digest`, and
  `execution_digests` are byte-identical before and after.
- `sh scripts/verify-phase2.sh` exits 0 and reports all nine Phase 2 control
  IDs (including `control:backend.runtime_causality`, now marker-located)
  and both Phase 1 controls through the same freshly built binary.
- `env GOCACHE=/tmp/ai-lang-phase3-cache go test ./...`,
  `go test -race ./...`, and `go vet ./...` all pass with zero findings.
- Shipped-binary drive (D-11): `format --check`, `check`,
  `run --engine=interpreter`, `run --engine=native` all pass on
  `testdata/phase3/branch_view.lang` and on a hand-written, non-corpus
  3-alternative fixture (`Red`/`Yellow`/`Green`, exercising borrow, copy,
  and take across three arms) — full O0/O3/interpreter agreement in both
  cases.

## Known Stubs

None. `LoanEndpoints` is intentionally left empty this plan (edge-specific
loan liveness is explicitly deferred to plan 03-03 per the phase's own
validation strategy and Q1's acyclic-branch scoping) — this is a documented
future-plan boundary, not a stub blocking this plan's own goal.

## Next Phase Readiness

- The CFG structure (blocks, edges, per-function-local ordinals) that
  03-02 (exclusive borrows) and 03-03 (edge-specific liveness dataflow)
  need now exists and is independently validated.
- `ability.DeriveSealed`/`deriveAbility`'s sealed-leaf mechanism is
  available for any future declared-type ability need.
- OWN-03 remains explicitly scoped to acyclic (branch-only) CFGs this
  phase; no loop or back-edge claim appears anywhere in code, comments,
  or this summary.
- No blockers for 03-02.

---
*Phase: 03-borrowed-views-and-cfg-lifetimes*
*Plan: 01*
*Completed: 2026-09-04*

## Self-Check: PASSED
