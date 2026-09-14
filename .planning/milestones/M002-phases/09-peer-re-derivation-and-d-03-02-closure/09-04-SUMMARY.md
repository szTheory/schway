---
phase: 09-peer-re-derivation-and-d-03-02-closure
plan: 04
subsystem: compiler-validation
tags: [go, ownership, loan-liveness, corevalidate, check, call-graph, differential-testing, synthetic-corpus, mutation-kill]

requires:
  - phase: 09-peer-re-derivation-and-d-03-02-closure
    provides: "09-01: corevalidate's own peer interprocedural loan-carry derivation and its two cross-package fault-injection seams (SetForcePeerLoanCarryTrueForTest/SetDisablePeerLoanCarryConsultForTest); 09-02: testsupport.GenerateCallGraphCorpus/CallGraphCorpusShapes, the relocated five-shape call-graph generator"
provides:
  - "check_peer_shape_differential_test.go (package check): a synthetic-shape zero-divergence differential proving TRU-04 criterion 1 over diamond/deep-chain/dense/parser-shaped/forward call-graph shapes the real .lang corpus structurally cannot reach"
  - "TRU-04's 'recursion' shape settled as a cycle-peer witness-agreement differential (self, mutual, indirect 3+/4-hop), never a liveness one -- the category-error disposition (D-09-22) recorded in code where a future reader will hit it first"
  - "A mutation-kill proving the differential's own detection machinery is load-bearing, not resting on the five relocated shapes' own provably vacuous zero-divergence finding (none of them ever contain a borrow operation)"
affects: [09-06, 09-08, 09-10]

actuals:
  tokens: 9824
  tasks: 3
  commits: 2

tech-stack:
  added: []
  patterns:
    - "A synthetic-corpus divergence gate living INSIDE the vehicle-constrained package (check), mirroring an existing exact-set gate's (peerDivergenceExpected) own two-directional discipline by construction and cross-reference, never a second, differently-truthed harness (D-09-50)"
    - "Set-membership (not multiset-equality) comparison for two peers whose evidence is structurally asymmetric in shape: check's causes enumerate a full cycle member list, corevalidate's Problem carries only the single witness node that closed the cycle"
    - "A generator's own Reachability Register must name what it provably does NOT reach, including gaps outside its designer's original charter (the five call-graph shapes' total absence of any borrow operation, discovered only by inspecting testsupport/callgraphcorpus.go directly)"

key-files:
  created:
    - internal/compiler/check/check_peer_shape_differential_test.go
  modified:
    - internal/compiler/corevalidate/corevalidate_cycle_peer_test.go

key-decisions:
  - "Tasks 1, 2 (check-side half), and 3 landed in one commit for check_peer_shape_differential_test.go: all three build incrementally on the same new file's own helpers (checkSyntheticProgramVerdict, completeSyntheticProgramForCorevalidate), and splitting them would have required reconstructing artificial intermediate states of the same file with no meaningful checkpoint between them -- the same reasoning 09-01-SUMMARY.md recorded for its own combined Tasks 1+2."
  - "Discovered mid-Task-1 (not anticipated by the plan): testsupport's five call-graph shapes never contain a single OpBorrowShared/OpBorrowExclusive operation anywhere -- they were built for check's own usesParam work-counter (Phase 08), not for loan-liveness sensitivity. This makes the zero-divergence sweep over them PROVABLY VACUOUS (neither peer's loan-liveness law has anything to disagree about). Documented explicitly in the file's own header comment (the Generator Reachability Register's 'provably does NOT reach' row), rather than left implicit, per the standing process rule against an undeclared unreached generator space. This finding is exactly why Task 3's mutation-kill needed a hand-built fixture rather than a corrupted corpus-generated one."
  - "testsupport's own generated core.Program values needed substantial structural completion before corevalidate.Validate could even reach the loan-liveness fact under test: EntryPointID/ReturnPointID, a per-function TypeFact, and -- most significantly -- every place ID and operation ID/PointID remapped to corevalidate's required strict '<functionID>:place:<index>' / '<functionID>:op:<index>' sequential naming (testsupport's own IDs, e.g. '...:place:rN', were built only for check's own work-counter, which never validates naming convention). completeSyntheticProgramForCorevalidate performs this remap in first-reference order, touching only place/operation IDENTITY bookkeeping -- never the call-graph topology (CalleeID edges) the relocated generator itself produced, keeping D-09-23's 'same generator' claim intact."
  - "Task 3's mutation surface is corevalidate's own cross-package seam (SetForcePeerLoanCarryTrueForTest, from 09-01), not a check-side seam. loanLivenessBoundSeam was considered and rejected: it flips check from ADMIT to REFUSE, the wrong direction for a differential that only tracks check-admits/peer-refuses. check.disableInterproceduralLoanLivenessForTest was also considered and rejected: it only gates check.Program's own call site, which this differential never calls (it drives checkInterproceduralLoanLiveness directly, per D-09-50). Both rejections, and the reasoning, are recorded in the mutation test's own doc comment."

requirements-completed: []

coverage:
  - id: D1
    description: "A shadow run over diamond, deep-chain, dense, parser-shaped, and forward call-graph shapes shows ZERO divergence between check and corevalidate on the interprocedural loan-liveness fact, at more than one size per shape"
    requirement: "TRU-04"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_peer_shape_differential_test.go#TestSyntheticShapeDifferentialHasNoUndeclaredDivergence"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_peer_shape_differential_test.go#TestSyntheticShapeDifferentialUsesTheRelocatedGenerator"
        status: pass
    human_judgment: false
  - id: D2
    description: "TRU-04's 'recursion' shape is satisfied as a cycle-peer differential -- both layers agree on refusal-or-not AND on the witness (compared as a set) for self, mutual, and indirect (3+/4-hop) cycles; the acyclic-diamond negative direction is asserted on both sides"
    requirement: "TRU-04"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_cycle_peer_test.go#TestCyclePeerRefusesIndirectCycle"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_cycle_peer_test.go#TestCyclePeerRefusesFourHopIndirectCycle"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_peer_shape_differential_test.go#TestCyclePeerAgreesWithCheckOnSelfAndMutualCycleWitness"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_peer_shape_differential_test.go#TestCyclePeerAgreesWithCheckOnIndirectCycleWitness"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_peer_shape_differential_test.go#TestCyclePeerAndCheckBothAcceptAcyclicDiamond"
        status: pass
    human_judgment: false
  - id: D3
    description: "The differential has been seen to report a divergence under a deliberate mutation, proving the zero-divergence finding is not resting on an untested vacuous truth"
    requirement: "TRU-04"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_peer_shape_differential_test.go#TestSyntheticShapeDifferentialMutationReintroducesDivergence"
        status: pass
    human_judgment: false
  - id: D4
    description: "go test ./... && go vet ./... exits 0"
    verification:
      - kind: integration
        ref: "go test ./... && go vet ./... (full repo)"
        status: pass
    human_judgment: false

duration: 55min
completed: 2026-09-10
status: complete
---

# Phase 09 Plan 04: Synthetic-Shape Zero-Divergence Differential and Cycle-Peer Witness Agreement Summary

**A per-shape zero-divergence differential (living inside package `check`, per D-09-50's vehicle split) proves TRU-04 criterion 1 over five call-graph shapes the real `.lang` corpus structurally cannot reach, a cycle-peer witness-agreement differential settles TRU-04's "recursion" shape honestly as a category-error disposition rather than a liveness claim, and a mutation-kill proves the differential's own detection machinery works — discovering along the way that the five relocated shapes never contain a single borrow operation, making the natural sweep provably vacuous on its own.**

## Performance

- **Duration:** 55 min
- **Tasks:** 3 (2 commits — see Decisions)
- **Files created:** 1
- **Files modified:** 1

## Accomplishments

- `internal/compiler/check/check_peer_shape_differential_test.go` (new, package `check`): `syntheticShapeDivergenceExpected` (empty, mirroring `peerDivergenceExpected`'s own doc-comment discipline), `checkSyntheticProgramVerdict` (calls `buildCallSignatureTable`/`buildInterproceduralSummaries`/`checkInterproceduralLoanLiveness` directly on a bare `core.Program`, since `check.Program` cannot accept one), `completeSyntheticProgramForCorevalidate` (structural scaffolding + place/operation-ID remap so `corevalidate.Validate`'s full structural pass can reach the loan-liveness replay), `findSyntheticShapeDivergences`/`assertSyntheticShapeDivergencesMatchExpected` (the sweep and its two-directional exactness assertion, factored apart so Task 3 can assert on the divergence map directly).
- `TestSyntheticShapeDifferentialHasNoUndeclaredDivergence`: sweeps all five `testsupport.CallGraphCorpusShapes()` at sizes 4/16/64, proves zero divergence.
- `TestSyntheticShapeDifferentialUsesTheRelocatedGenerator`: proves the shape list is exactly what `testsupport.CallGraphCorpusShapes()` returns, and a static scan proves this file declares no local corpus builder.
- Cycle-peer extension: `TestCyclePeerRefusesIndirectCycle`/`TestCyclePeerRefusesFourHopIndirectCycle` (corevalidate package, peer-only half) plus `TestCyclePeerAgreesWithCheckOnSelfAndMutualCycleWitness`/`TestCyclePeerAgreesWithCheckOnIndirectCycleWitness`/`TestCyclePeerAndCheckBothAcceptAcyclicDiamond` (check package, both-layers-agree half) — a new `cyclePeerProgram`/`cyclePeerProgramFunction` builder (package check's own, since `syntheticProgram` is unexported and unreachable across the package boundary), and `assertCyclePeerWitnessAgreement`, which compares the cycle-member witness as a SET (check's causes enumerate the full member list; corevalidate's `Problem.Detail` carries only the single node that closed the cycle).
- `TestSyntheticShapeDifferentialMutationReintroducesDivergence`: a hand-built two-function fixture (`mutationCandidateProgram`, mirroring `testdata/phase08/twin_a_accept.lang`'s own proven shape) fed through `corevalidate.SetForcePeerLoanCarryTrueForTest`, proving the differential genuinely detects a divergence rather than vacuously passing.

## Task Commits

1. **Task 2 (peer-only half): extend cycle-peer differential with indirect cycle cases** — `ffc5756` (test)
2. **Tasks 1 + 2 (check-side half) + 3: synthetic-shape differential, cycle-peer witness agreement, mutation-kill** — `2685e03` (feat)

**Plan metadata:** (this commit)

## Files Created/Modified

- `internal/compiler/check/check_peer_shape_differential_test.go` — new file, all of Tasks 1-3 (see Accomplishments)
- `internal/compiler/corevalidate/corevalidate_cycle_peer_test.go` — `TestCyclePeerRefusesIndirectCycle`, `TestCyclePeerRefusesFourHopIndirectCycle`

## Decisions Made

- **Tasks 1, 2 (check-side half), and 3 combined into one commit for the new file.** All three build incrementally on the same file's own helpers; splitting would reconstruct artificial intermediate states with no meaningful checkpoint between them, mirroring 09-01's own precedent for its Tasks 1+2.
- **The five relocated corpus shapes never contain a borrow operation — documented, not hidden.** Discovered by inspecting `testsupport/callgraphcorpus.go` directly: `corpusRelay`/`corpusLeafUse`/`corpusLeafPass`/`corpusForward` only ever emit `OpCall`, `OpCopy`, and `OpReturn`. This makes `TestSyntheticShapeDifferentialHasNoUndeclaredDivergence`'s own zero-divergence finding provably vacuous over these five shapes alone (neither peer's loan-liveness law has anything to disagree about). Recorded explicitly in the file's own header comment, in the Generator Reachability Register's vocabulary, per the standing process rule against an undeclared unreached generator space — and is exactly why Task 3 needed a hand-built fixture containing a genuine borrow, rather than a corrupted corpus-generated program.
- **`completeSyntheticProgramForCorevalidate` performs a place/operation-ID remap, not merely additive scaffolding.** `corevalidate.Validate`'s structural pass requires every place ID and operation ID/PointID to read exactly `<functionID>:place:<index>` / `<functionID>:op:<index>` in strict declaration order (`core.place_order`/`core.operation_order`); `testsupport`'s own IDs (`...:place:rN`, `...:op:call0`) were built only for `check`'s own work-counter, which never checks naming convention. The remap touches place/operation IDENTITY bookkeeping only, in first-reference order, and never the call-graph topology (`CalleeID` edges) the relocated generator produced — keeping D-09-23's "same generator" claim intact.
- **Task 3's mutation surface is `corevalidate.SetForcePeerLoanCarryTrueForTest` (a corevalidate-side cross-package seam from 09-01), not a check-side seam.** `loanLivenessBoundSeam` was rejected (flips check's own verdict the wrong direction — ADMIT to REFUSE, not the check-admits/peer-refuses direction this differential tracks). `check.disableInterproceduralLoanLivenessForTest` was rejected (only gates `check.Program`'s own call site, never reached since this differential drives `checkInterproceduralLoanLiveness` directly). Both rejections are recorded in the mutation test's own doc comment.
- **The witness-agreement comparison is set-membership, not multiset equality.** `corevalidate`'s `Problem.Detail` carries only the ONE node whose back edge closed the cycle (its own DFS's `return v.check(false, core.CallGraphCycle, child)`), while `check`'s causes enumerate up to `callgraph.MaxCycleCauses` full members. `assertCyclePeerWitnessAgreement` asserts the peer's single witness is a genuine member of check's reported set — the honest comparison this structural asymmetry actually supports, documented in the shared helper's own doc comment so a reader does not mistake it for an incomplete exact-equality check.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Synthetic corpus programs needed Schema/ModuleID set before `corevalidate.Validate` could run**
- **Found during:** Task 1, first run of `TestSyntheticShapeDifferentialHasNoUndeclaredDivergence`
- **Issue:** `testsupport.GenerateCallGraphCorpus` never sets `core.Program.Schema`/`Module`/`ModuleID` (irrelevant to check's own work-counter); `corevalidate.Validate`'s first structural check (`core.schema`) refused every generated program before reaching the loan-liveness fact under test
- **Fix:** Set `Schema: core.Schema1`, `Module`, `ModuleID` on the generated program before validation, inline in `findSyntheticShapeDivergences`
- **Files modified:** `internal/compiler/check/check_peer_shape_differential_test.go`
- **Verification:** `TestSyntheticShapeDifferentialHasNoUndeclaredDivergence` passes
- **Committed in:** `2685e03`

**2. [Rule 3 - Blocking] Synthetic corpus programs needed full structural completion (EntryPointID/ReturnPointID, Types, Places, ID renaming) for `corevalidate.Validate` to reach the loan-liveness replay**
- **Found during:** Task 1, iterative runs surfacing `core.point_id`, `core.place_order`, `core.parameter_mismatch`, `core.operation_order`, `core.final_claim_mismatch`, and `core.call_return_type_mismatch` in sequence
- **Issue:** `testsupport`'s generator builds the minimum `core.Program` shape `check`'s own work-counter needs, which is far short of what `corevalidate.Validate`'s full structural pass (place/operation naming convention, type/return-type consistency) requires
- **Fix:** `completeSyntheticProgramForCorevalidate` — sets `EntryPointID`/`ReturnPointID`/`ReturnType`, builds a per-function `TypeFact`, and remaps every place ID and operation ID/PointID to corevalidate's required strict sequential naming, in first-reference order (never touching call-graph topology)
- **Files modified:** `internal/compiler/check/check_peer_shape_differential_test.go`
- **Verification:** `TestSyntheticShapeDifferentialHasNoUndeclaredDivergence` passes; `go test ./internal/compiler/check/... ./internal/compiler/corevalidate/... -count=2 -shuffle=on` passes
- **Committed in:** `2685e03`

### Verify-Command Discrepancy (documented, not a code deviation)

Task 2's third `<verify>` command (`grep -rn "compiler/check" internal/compiler/corevalidate/corevalidate_cycle_peer_test.go`, `fails_when`: exits 0 with any output) cannot pass literally: the file's own PRE-EXISTING doc comment and import-guard code (unchanged by this plan, present since Phase 07) already contain the substring `"compiler/check"` twice — once in the doc comment naming what the guard forbids, once in the guard's own string-literal check (`strings.HasSuffix(importPath, "/compiler/check")`). This was true of the file before this plan touched it at all. The plan's verify command as literally written was never satisfiable; the genuine claim it is checking for (no ACTUAL import of `compiler/check`) is what `TestCyclePeerTestFileImportsStayIndependent` mechanically enforces, and that test passes (confirmed above). No code change was made in response to this — it is a verify-script imprecision, not a defect in the file.

## Known Stubs

None.

## Threat Flags

None — this plan's threat register items (T-09-01, T-09-02, T-09-12, T-09-13) are all directly mitigated: T-09-01 by Task 3's mutation-kill; T-09-12 by `syntheticShapeDivergenceExpected`'s doc comment citing `peerDivergenceExpected` and D-09-50 by name; T-09-13 by every new cycle-peer test's own doc comment stating the liveness-over-recursion impossibility; T-09-02 by the peer's own already-iterative walk (Plan 09-01) and Task 3's mutation asserting a reported divergence, never a panic.

## Issues Encountered

See Deviations above. Both auto-fixes were discovered iteratively by running the test and reading the next structural refusal code corevalidate surfaced, then fixing that specific gap — six distinct structural codes in sequence (`core.schema`, `core.point_id`, `core.place_order`, `core.parameter_mismatch`, `core.operation_order`, `core.final_claim_mismatch`/`core.call_return_type_mismatch`) before the differential could reach the actual loan-liveness fact under test. This is the concrete, worked evidence behind this file's own header-comment finding: `testsupport`'s generator was built for one specific consumer (check's own work-counter) and needed real completion work to serve a second, more strictly-validating consumer (corevalidate's full structural pass).

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- TRU-04 criterion 1 is proven over shapes the committed `.lang` corpus cannot express; TRU-04's "recursion" shape is settled honestly as a cycle-peer differential, not a liveness claim, with the disposition recorded in code for plan 09-10's requirement-text correction to cite.
- **Carried forward, not blocking:** the five relocated call-graph shapes never contain a borrow operation, so `TestSyntheticShapeDifferentialHasNoUndeclaredDivergence`'s zero-divergence finding is provably vacuous on its own (documented in the file's own header comment). Task 3's mutation-kill closes the gap for THIS plan's own success criteria (the differential has been seen to fail), but a future plan wanting a NATURALLY loan-bearing synthetic shape (rather than a single hand-built fixture) would need to extend `testsupport`'s generator itself — out of this plan's scope (D-09-23 forbids duplicating or forking the generator; extending it is a different package's call).
- `syntheticShapeDivergenceExpected` stays empty; TRU-04 stays `Pending` in REQUIREMENTS.md per this plan's own `<requirement_marking_rule>` (also carried by 09-01/09-02/09-06/09-08/09-10).
- No blockers for 09-05 through 09-10 from this plan's own scope.

---
*Phase: 09-peer-re-derivation-and-d-03-02-closure*
*Completed: 2026-09-10*

## Self-Check: PASSED
