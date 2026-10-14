---
status: resolved
trigger: "Phase 35 raw shift validators accept initialized target reuse despite planned fresh-target law"
created: 2026-10-14
updated: 2026-10-14
---

## Current Focus

bug_class: Bohrbug (deterministic raw-core admission defect)
hypothesis: Core's replay freshness predicate silently returns false without recording a Problem; origin lacks a competing-definition check entirely.
test: Add shift-only prior-static-write checks and preserve source-produced shift temporaries across mutable assignment, branch joins and loop backedges.
expecting: Both named freshness repros refuse reused initialized targets, while one static shift definition may execute repeatedly in a loop.
next_action: Return the verified repair/receipts to the parent Phase35 workflow for its remaining D-11 admission mutation harness.
reasoning_checkpoint:
  hypothesis: Shift target reuse is admitted because core's failed scalarTargetMatches does not append a Problem (Valid depends only on Problems), and origin's structural shift predicate never checks other writes to the target.
  confirming_evidence:
    - Both TestPhase35ShiftTargetFreshness tests failed with no Problems on an initialized OpConst target with its consumer rewired.
    - corevalidate scalarTargetMatches returns a bare bool; replay returns false directly; Validate sets Valid from len(v.problems)==0.
    - origin phase27ArithmeticProblem checks types/immediate/aliasing but does not count target definitions.
    - check/scalar.go allocates a separate immutable temporary for a shift, then uses OpScalarStore for assignment; loop lowering has one static operation per expression.
  falsification_test: A reused initialized shift target produces a refusal before any shift-specific definition check, or a legitimate producer loop needs two static writers to its shift temporary.
  fix_rationale: Independently require each shift target to have no prior static writer, outside fixed-point sweeps. This records a real refusal without treating loop reexecution as a new definition or imposing SSA on mutable locals. Later ordinary mutable assignments retain their existing laws.
  blind_spots: Hosted Linux/macOS and native lanes are outside this repair; branch/assignment/loop controls passed locally, including explicit backedge and completed replay checks.
  candidate_causes:
    - code: missing origin target-definition check and silent core replay failure
    - data: uninitialized consumer or source could have masked a refusal, ruled out by initialized OpConst target and rewired consumer
    - config/environment: no feature/config gates are used by these deterministic raw-core predicates
  and_gate: No; each peer independently admits the same invalid target, without a co-occurring runtime or configuration condition.

## Symptoms

expected: Raw shifts independently require initialized U64 source and fresh, valid, nonparameter U64 result place under D-09/D-11 and 35-04-PLAN.
actual: Nyquist auditor found that both peers accept a reused initialized OpConst target when the consumer is rewired accordingly.
errors: Existing inherited TestPhase35ShiftTargetFreshness reproductions expose missing target refusal; exact output pending rerun.
reproduction: Named TestPhase35ShiftTargetFreshness in each peer's corevalidate_phase35_test.go and originvalidate_phase35_test.go; includes separate uninitialized-source rows.
timeline: Newly discovered during final Phase 35 validation after all eight execution plans; no final validation completion claimed.

## Scope and Authority

- Active GSD execute Phase 35 delegates this bounded debug repair; user authorized coherent follow-up repairs.
- Owned production paths: internal/compiler/corevalidate/corevalidate.go and internal/compiler/originvalidate/originvalidate.go.
- Owned inherited repro paths: internal/compiler/corevalidate/corevalidate_phase35_test.go and internal/compiler/originvalidate/originvalidate_phase35_test.go.
- Preserve valid source assignment, CFG and loop shapes, legacy Byte admission, and all unrelated edits, including .planning/config.json.
- No blanket SSA freshness rule on mutable operations, requirement waiver, unrelated source edit, global test/package sweep, install, hosted work, branch or worktree.
- Run only anchored named tests with GOCACHE=/tmp/ai-lang-verification-gocache, -count=1 -v -timeout=600s.
- Root-pin guard /tmp/schway34-root-pin.sh before first edit and every commit; exact-path SDK query commit only.
- Do not claim D-11 fully covered from invalid-input controls; parent owns remaining actual admission mutation harness.

## Evidence

- timestamp: 2026-10-14
  checked: Anchored session selector TestPhase35PreChangeEmitterCorpus|TestPhase35PreChangeCorpusControls|TestPhase35PreChangeNumericDeltaControls|TestPhase35FocusedValueAnswers|TestPhase35PublicDocumentValues (all fully anchored as one selector)
  found: All five top-level tests RUN/PASS, no skips, exit 0; session elapsed 2.399s. git diff --check exit 0. Host darwin/arm64, go1.24.0.
  implication: Old emitter bytes and the runnable Phase35 value/public document witnesses are preserved locally; no native matrix or hosted lane claimed.

- timestamp: 2026-10-14
  checked: Serial owner-local refusal removal, then restoration (TestPhase35ShiftTargetFreshness)
  found: Core refusal bypass true || freshTarget caused all eight rows to FAIL (exit 1); restored core PASS (exit 0). Origin refusal bypass false && !freshTarget caused all eight rows to FAIL (exit 1); restored origin and full targeted peer selector PASS (exit 0). No mutation remains.
  implication: Both new refusal sites are reached and load-bearing; manual scoped mutation substitutes for absent Stryker. This is only freshness coverage, not the remaining D-11 admission harness.
- timestamp: 2026-10-14
  checked: Final targeted peer selector after restoration
  found: Core 10 top-level RUN/PASS tests; origin 9 top-level RUN/PASS tests; no skips/failures. Includes Phase35 ShiftTargetFreshness, ShiftProducerShapes, ShiftImmediateAdmission, ScalarTransferMutants, AnalysisTriggers, UnusedByteIRAdmission, LegacyByteCompatibility; peer baselines and Phase19 constants. Loop controls explicitly assert a backedge; core positive controls require a finished ClosureDigest to exclude silent replay exit.
  implication: Held-out adjacent laws and real producer semantics survive the bounded repair.

- timestamp: 2026-10-14
  checked: Expanded freshness RED then GREEN runs, plus anchored adjacent selector
  found: Before production edits all eight initialized target-reuse rows failed per peer. After independent structural checks all rows passed, as did producer assignment/branch/loop controls and selected Phase 35 admission/transfer/trigger/Byte and Phase 19 baseline tests (exit 0).
  implication: Actual behavior changes are scoped to shift target introduction; core emits core.invalid_target and origin emits core.scalar_operation_invalid.

- timestamp: 2026-10-14
  checked: TestPhase35ShiftProducerShapes in both packages before production edits
  found: Both RUN/PASS, including assignment, branch and loop subtests; exit 0.
  implication: Preserve these source-produced shapes; freshness tracks static introduction, not dynamically initialized fixed-point state.

- timestamp: 2026-10-14
  checked: Initial anchored go test on both peers, TestPhase35ShiftTargetFreshness, -count=1 -v -timeout=600s
  found: Exit 1; both named tests RUN and FAIL; core Valid=true Problems=[] and origin problems=[] on initialized reused OpConst targets.
  implication: Actual RED receipts establish the missing refusal independently of source initialization.
- timestamp: 2026-10-14
  checked: core Validate/run/replayBlocks/scalarTargetMatches and origin phase27ArithmeticProblem
  found: Core silently stops replay on target reuse but returns Valid based only on the empty Problems list; origin has no competing-definition law.
  implication: Add explicit shift-scoped structural refusal in both authorities; changing transfer alone cannot repair this.
- timestamp: 2026-10-14
  checked: check/scalar.go shift expression and var/assign/while lowering
  found: Shift expression creates addPlace("scalar", "U64", false); assignments use separate StoreTargetID; each loop expression has one static definition.
  implication: Static writer uniqueness preserves repeated dynamic loop execution and mutable stores to independent locals.
- timestamp: 2026-10-14
  checked: Knowledge-base fallback, project skills and optional SBFL
  found: No prior matching shift-freshness resolution; no project skills/configured debugger skills; no per-test coverage available for SBFL.
  implication: Direct deterministic repro and source tracing used; no coverage ranking claimed.

- timestamp: 2026-10-14
  checked: 35-04-PLAN Task 1 action
  found: Explicit fresh/nonparameter U64 result obligation; independent peers own their own admission checks.
  implication: A scoped validator repair is authorized; source producer shapes must guide the precise freshness interpretation.

## Eliminated

- hypothesis: The reused-target repro only passes because its consumer is uninitialized or its source is invalid.
  evidence: Initialized OpConst/copy/store/shift targets and rewired consumers all reproduced acceptance before the fix; source-initialization refusal remains independently covered and passing.
  timestamp: 2026-10-14
- hypothesis: Freshness requires an uninitialized target on every dynamic fixed-point transfer.
  evidence: Producer loops reuse the same expression operation on backedges; all three producer controls pass with static introduction tracking instead. Dynamic initialization would reject valid loop execution.
  timestamp: 2026-10-14

## Resolution

root_cause: Core silently returns from scalar replay when its target predicate fails without recording a Problem, so Valid remains true; origin omits a prior-definition check from its shift contract.
fix: Each peer independently tracks prior static writes (ordinary, store, error and payload targets) and explicitly refuses a shift reusing such a target. The shift-only law executes outside fixed-point sweeps; mutable assignments and loop reexecution retain existing semantics.
oracle_type: specified (D-09/D-11 and 35-04-PLAN fresh/nonparameter result contract)
verification:
  target_test: { result: pass, tests: "both peers TestPhase35ShiftTargetFreshness; 8 initialized target-reuse rows each" }
  mutation_check: { result: pass, method: "manual owner-local refusal bypass because Stryker is not configured", mutant_killed: "core true || freshTarget and origin false && !freshTarget, independently; all 8 rows each failed, then restored" }
  no_op_deletion: { result: pass, deletion_justified_by_rca: false, evidence: "27 additive production lines; explicit shift refusals; no behavior/assertions removed or weakened" }
  adjacent_tests: { result: pass, suites_run: ["core targeted selector: 10 top-level tests", "origin targeted selector: 9 top-level tests", "session targeted selector: 5 top-level tests"] }
  revert_and_reconfirm: { result: pass, bug_returned_on_revert: true, fixed_on_reapply: true, method: "serial semantic removal/restoration of the two refusal gates without git stash/reset; equivalent to reverting the admission change" }
  guardrail_verdict: accepted
limitations: Local Darwin arm64/Go1.24.0 only. Other existing scalarTargetMatches callers may still silently return false; this repair is intentionally shift-scoped. Parent owns complete D-11 refusal-removal matrix; invalid-input controls alone do not establish that claim. KB/MemPalace writes omitted because delegated ownership excludes those artifacts.
commit: 3800080e (fix(phase35): independently refuse reused shift targets)
completion_authority: Parent delegated autonomous repair, verification and archive within the already authorized Phase35 execute workflow; no separate routine human-verify checkpoint requested.
files_changed: [internal/compiler/corevalidate/corevalidate.go, internal/compiler/corevalidate/corevalidate_phase35_test.go, internal/compiler/originvalidate/originvalidate.go, internal/compiler/originvalidate/originvalidate_phase35_test.go]

## Exact Local Receipts

All Go commands used GOCACHE=/tmp/ai-lang-verification-gocache and -count=1 -v -timeout=600s.

1. `go test ./internal/compiler/corevalidate ./internal/compiler/originvalidate -run '^TestPhase35ShiftTargetFreshness$'`: initial inherited repro RED, exit 1 in both; expanded eight-row repro also RED in both.
2. `go test ./internal/compiler/corevalidate ./internal/compiler/originvalidate -run '^TestPhase35ShiftProducerShapes$'`: pre-fix controls PASS, exit 0 in both (assignment/branch/loop).
3. `go test ./internal/compiler/corevalidate -run '^TestPhase35ShiftTargetFreshness$'`: core-only refusal bypass RED, exit 1; restoration GREEN, exit 0.
4. `go test ./internal/compiler/originvalidate -run '^TestPhase35ShiftTargetFreshness$'`: origin-only refusal bypass RED, exit 1; restoration GREEN in final selector, exit 0.
5. `go test ./internal/compiler/corevalidate ./internal/compiler/originvalidate -run '^(TestPhase35ShiftTargetFreshness|TestPhase35ShiftProducerShapes|TestPhase35ShiftImmediateAdmission|TestPhase35ScalarTransferMutants|TestPhase35AnalysisTriggers|TestPhase35UnusedByteIRAdmission|TestPhase35LegacyByteCompatibility|TestPhase35CorePeerBaseline|TestPhase35OriginPeerBaseline|TestPhase19OpConstIndependentAdmission|TestPhase19ForgedOpConstFactsRefused|TestPhase19ConstantOriginStopsAtRoot)$'`: final restored GREEN, exit 0; core 10 named tests, origin 9. All actual selected top-level names ran and passed, no skip.
6. `go test ./internal/compiler/session -run '^(TestPhase35PreChangeEmitterCorpus|TestPhase35PreChangeCorpusControls|TestPhase35PreChangeNumericDeltaControls|TestPhase35FocusedValueAnswers|TestPhase35PublicDocumentValues)$'`: all five named RUN/PASS, exit 0.
7. `git diff --check`: exit 0. SDK exact-path code/test commit returned committed=true, 3800080e; no config change staged.
