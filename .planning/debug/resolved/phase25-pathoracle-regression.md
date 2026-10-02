---
status: resolved
trigger: "After Phase 25 security repairs, the fresh full Go suite shows the new path-oracle validator rejecting unrelated legacy interprocedural borrow fixtures and invalidating several source/evidence pins."
created: 2026-10-02
updated: 2026-10-02
---

## Current Focus

hypothesis: The straight-line pointer validator treated both OpCall results and explicitly declared shared-return results as invalid owned-result escapes; it lacks callee contract composition and must distinguish owned from declared borrowed returns. Separate line and documentation pin failures are independent.
bug_class: deterministic regression
test: Run pathoracle, session, originvalidate, and repair packages after treating OpCall targets as unknown ancestry while retaining source-loan last-use and owned-result escape behavior.
expecting: All original Phase 25, legacy admission, and repair tests pass; new mixed-call overlap and escape tests prove local checks remain active across calls.
next_action: Archive this confirmed session and commit the focused code, test, pin, debug, and knowledge-base artifacts.
oracle_type: derived
reasoning_checkpoint:
  hypothesis: "The local pointer replay rejects valid legacy borrow results because it propagates source ancestry through OpCall without callee return provenance and unconditionally rejects loan-derived returns even when an explicit borrow origin is declared."
  confirming_evidence:
    - "A deterministic focused rerun reproduces 11 undeclared command-admission divergences across phase3, phase6, phase8, phase10, phase11, and phase13 fixtures, all with core.pathoracle_refused."
    - "TestOpCallOriginWalkGateIsLoadBearing fails on twin_a_accept with pathoracle.pointer_escape at escort's call-result place; the fixture documents that borrow provenance for OpCall results is contract-dependent."
    - "validateStraightLinePointerLoans writes inherited source ancestry onto every non-copy operation target, including OpCall, and checks every OpReturn against that ancestry."
  falsification_test: "After breaking ancestry at OpCall targets and limiting escape refusal to owned returns, rerun the exact failing originvalidate/session/repair tests and mixed-call overlap/escape negative controls; if a known call witness still fails or local overlap/escape is missed, this hypothesis is incomplete."
  fix_rationale: "The Phase 25 replay remains active inside call-containing functions. A call remains a use of its source loans, while its result ancestry is delegated to the contract-aware core/origin peers; declared borrowed returns are not owned-result escapes."
  blind_spots: "Call-result ancestry is intentionally unknown to this local replay; correctness for later uses of that result depends on the existing independent peers. The mixed-call negative controls must prove direct local invariants still hold."
  candidate_causes:
    - "code: the new path validator misclassifies a borrowed result returned through OpCall as a direct local escape"
    - "code: applying the new path gate to every command admission broadens it beyond the Phase 25 pointer-successor subset"
    - "evidence: insertion shifted a registered EmitNative call and the test census still pins the old validation command"
    - "environment: the regression appears only after the Phase 25 repair commits, rather than depending on host/runtime differences"
    - "data: the failing programs exercise multiple distinct fixture families, so confirm they share the OpCall result shape rather than assuming fixture data is equivalent"
  and_gate: "No additional environment condition is required: source-only propagation through an OpCall result or treating a declared borrowed return as owned is sufficient. Inventory and evidence-pin failures are independent documentation/test-index conditions."

## Symptoms

expected: `go test -count=1 ./...` passes after Phase 25 security repairs, without changing Phase 24/legacy interprocedural borrow semantics.
actual: The fresh full suite exits 1. Repair driver borrow/interprocedural-loan cases return `reverify_failed`; legacy originvalidate and CLI admission fixtures are refused with `core.pathoracle_refused`; Phase 16 emitter inventory sees `EmitNative` at line 2978 instead of pinned 2784; verification-groundedness tests report Phase 25 validation command at line 41 as stale pinned R2b evidence.
errors: `cmd/schway-repair` has failures in `TestRepairDriverFixesEveryDefectClassSinglePass`, `TestHeldoutOutcomeSetContainsNoLaundering`, `TestRepairDriverFixesInterproceduralLoanLivenessSinglePass`, and `TestTwinPairBlame`. `internal/compiler/originvalidate` rejects `testdata/phase08/twin_a_accept.schway` because pathoracle reports function `s1:phase08.twin_a_accept:fn:escort` returns loan-derived place `...:place:2`. `internal/compiler/session` reports 11 undeclared admission divergences across Phase 3/6/8/10/11/13 fixtures, plus the stale Phase 16 emitter entry and stale Phase 25 groundedness frontier (`R2b=43`).
reproduction: `GOCACHE=/private/tmp/schway-gocache go test -count=1 ./...`
started: 2026-10-02, immediately after the five Phase 25 security-repair commits.

## Eliminated

## Evidence

- timestamp: 2026-10-02
  checked: Fresh default-parallel full Go suite after commits `b78952c`, `7472fbe`, `f56b135`, `b19446f`, and `3e6da03`.
  found: New `core.pathoracle_refused` rejects Phase 8 and many older accepted interprocedural fixtures; repair driver tests fail re-verification. Independent failures point to a shifted emitter inventory line and a stale Phase 25 validation-command evidence pin.
  implication: The new gate needs a narrower, evidence-defined Phase 25 applicability boundary; source inventory and evidence census require separate reconciliation after the code boundary is corrected.

- timestamp: 2026-10-02
  checked: Focused deterministic rerun `GOCACHE=/private/tmp/schway-gocache go test -count=1 ./internal/compiler/session ./internal/compiler/originvalidate ./cmd/schway-repair`.
  found: The run reproduced 11 `core.pathoracle_refused` divergences, the `TestOpCallOriginWalkGateIsLoadBearing` failure at `escort:place:2`, and five repair-driver reverify failures. The same run independently reproduced one stale `EmitNative` inventory location and two stale R2b groundness pins.
  implication: The OpCall applicability error is deterministic and shared across direct command admission and repair re-verification; the source/document evidence failures are independent.

- timestamp: 2026-10-02
  checked: Phase 08 fixture source and `validateStraightLinePointerLoans` implementation.
  found: The fixture's `escort` borrows `buffer`, passes that derived place to `escortee`, and returns the call result. The local replay transfers source ancestry to every non-`OpCopy` target, while the origin peer's call-composition logic treats call-result provenance as dependent on callee contract.
  implication: The new validator needs an explicit bounded applicability boundary around operations whose loan provenance it independently models.

- timestamp: 2026-10-02
  checked: `cmd/schway-repair` after the initial OpCall applicability fix.
  found: The interprocedural repair regressions are resolved, but the single-pass Phase 6 borrow repair still fails re-verification. Its held-out valid source declares `-> borrow(buffer) Buffer`; the new replay rejected its intentional returned loan because it checked every OpReturn as an owned-result escape.
  implication: Borrowed-return intent must be distinguished from an escape that contradicts an owned return contract.

- timestamp: 2026-10-02
  checked: Focused regression, negative-control, repair, and stale-pin tests after both fixes.
  found: The Phase 08 OpCall witness and Phase 6 declared shared-return witness pass the local path peer. Direct overlap and owned-result escape controls still fail closed. `TestRepairDriverFixesEveryDefectClassSinglePass/borrow` now passes. The Phase 16 emitter inventory and both verification-groundedness assertions pass after their exact source/evidence pins were reconciled.
  implication: Both pointer-replay semantic regressions and the independent stale test records are repaired; full affected-package and repository-wide checks remain.

- timestamp: 2026-10-02
  checked: Focused pathoracle tests after narrowing call handling.
  found: Regression controls for call-result deferral, declared shared return, mixed-call overlap, mixed-call owned-result escape, direct overlap, and direct borrowed-result escape all passed. `OpCall` output ancestry is broken locally while source ancestry remains live through the call operation.
  implication: Call-containing functions retain local overlap and direct escape validation instead of bypassing the oracle wholesale.

- timestamp: 2026-10-02
  checked: Full affected package matrix `GOCACHE=/private/tmp/schway-gocache go test -count=1 ./internal/compiler/pathoracle ./internal/compiler/session ./internal/compiler/originvalidate ./cmd/schway-repair`.
  found: All four packages passed. Session's full package suite completed in 63.281s; originvalidate and repair also passed after the final narrowed call semantics.
  implication: The original CLI admission, command divergence, repair re-verification, source inventory, and groundedness regressions are resolved across their owning packages.

- timestamp: 2026-10-02
  checked: Revert-and-reconfirm at the exact OpCall result ancestry condition.
  found: Temporarily changing `if operation.Kind == core.OpCopy || operation.Kind == core.OpCall` back to only `core.OpCopy` made `TestPhase25PointerPathDefersCallResultProvenance` fail with the original `pathoracle.pointer_escape` on `phase08.twin_a_accept`; reapplying the call boundary made that regression plus all direct and mixed-call overlap/escape tests pass.
  implication: The narrow call-result boundary, rather than unrelated shared-worktree changes, addresses the original reproduced defect.

- timestamp: 2026-10-02
  checked: Fresh full repository suite `GOCACHE=/private/tmp/schway-gocache go test -count=1 ./...` after the final call-boundary refinement.
  found: Exit code 0. `internal/compiler/session` completed in 279.242s; `native` in 61.582s; all reported packages, including CLI repair, syntax, testsupport, and scripts, passed.
  implication: The complete current Go test suite passes with the targeted regression guards in place.

- timestamp: 2026-10-02
  checked: Fresh uncached full repository suite after resuming the answered verification checkpoint, followed by all three quick-task focused validation commands.
  found: `GOCACHE=/private/tmp/schway-gocache go test -count=1 ./...` passed across every package; the three exact quick-task focused commands each exited 0, covering path-oracle controls, diagnostic attribution, and evidence-groundedness.
  implication: The final source and quick-task acceptance claims remain verified after the OpCall regression fix.

## Resolution

root_cause: The new local loan replay treated every non-copy operation, including `OpCall`, as preserving source ancestry and treated every loan-derived return as an invalid escape, including valid shared borrowed returns. It lacked callee contract composition and ignored declared return mode.
fix: `ValidateLocalOwnerPaths` remains active on call-containing functions, counts each call as a source-loan use, and assigns no local ancestry to the call result; it rejects a loan-derived return only when no declared `PublicOrigin` exists. Contract-aware peers own call-result and borrowed-return validation. Added regressions for Phase 08 call-result, Phase 6 shared-return, and mixed-call overlap/escape witnesses.
verification:
  target_test:
    result: pass
    evidence: "Full repository suite `GOCACHE=/private/tmp/schway-gocache go test -count=1 ./...` exited 0; all affected package tests passed."
  mutation_check:
    result: skipped
    reason_if_skipped: "No Go mutation-testing runner is configured for this repository; source-level mutated-core negative controls are present and passed."
  no_op_deletion:
    result: pass
    deletion_justified_by_rca: false
  adjacent_tests:
    result: pass
    suites_run:
      - "./internal/compiler/pathoracle"
      - "./internal/compiler/session"
      - "./internal/compiler/originvalidate"
      - "./cmd/schway-repair"
      - "./..."
  revert_and_reconfirm:
    result: pass
    bug_returned_on_revert: true
    fixed_on_reapply: true
    evidence: "At the exact OpCall result ancestry condition, the regression test failed with the original `pathoracle.pointer_escape` after temporarily restoring old behavior and passed after restoring the fix."
  guardrail_verdict: accepted
files_changed:
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/pathoracle/pathoracle_pointer_successor_test.go
  - testdata/phase16/public-emitter-consumers.json
  - internal/compiler/session/verification_groundedness_test.go

## Prevention

- **Code branch:** The local replay originally propagated all non-copy ancestry. That rule was too broad because `OpCall` carries a return value whose provenance depends on the callee's declared contract. The bounded oracle now stops local result ancestry at calls while still counting source loans as used; independent contract-aware peers retain responsibility for call-result provenance.
- **Evidence branch:** Earlier focused controls covered direct overlap and escape, but did not pair an `OpCall` result with both declared borrowed returns and owned-result escapes. The integration suite was the first broad consumer census to expose the unsupported operation shape.
- **Why not caught:** The original focused oracle tests did not exercise call-result provenance or declared borrowed-return exits, and the local package command did not include the wider origin/admission/repair consumers. The fresh uncached full suite caught the missing boundary when run after integration.
- **Recurrence guard:** `internal/compiler/pathoracle/pathoracle_pointer_successor_test.go` now includes `TestPhase25PointerPathDefersCallResultProvenance`, `TestPhase25PointerPathAllowsDeclaredBorrowReturn`, `TestPhase25PointerPathChecksOverlapAcrossCalls`, and `TestPhase25PointerPathChecksEscapeInCallFunction`; the uncached `go test ./...` suite exercises downstream command and repair flows.
