---
status: resolved
trigger: "Phase 16 public by-pointer emitter cuts appear to break Phase 11 native zero-attribute regression controls during the automated go test ./... gate. The phase's recorded disposition is cut-m004; determine whether the regression controls were migrated to refusal-first/frozen evidence correctly or still expect the removed public lowering path."
created: 2026-09-23T00:00:00-04:00
updated: 2026-09-26T19:47:39Z
---

## Current Focus

hypothesis: Phase 16's consumer migration misclassified the three Phase 11 gate emitter call sites as admitted dynamic schema-2 even though their N=1/N=2 inputs deliberately contain pointer-specialized functions; the gate therefore still requests live lowering from the current public emitter instead of proving refusal first and reading frozen, provenance-bound evidence.
test: run the five focused tests, trace each error to its emitter call, inspect the input shape and the public-emitter registry classification, and compare with successfully migrated Phase 4/5 refusal-first controls
expecting: all failures occur deterministically before Clang/runtime work; the Phase 11 fixture independently selects the by-pointer shape; the registry labels the calls admitted rather than refusal-frozen; migrated cut controls assert refusal and then use frozen evidence
next_action: return the root-cause-only diagnosis to the orchestrator; do not change source or tests
bug_class: bohrbug
candidate_causes:
  - code: the Phase 11 gate and direct diff-local helper still call cgen.EmitNative, and their consumer-registry rows were incorrectly recorded as admitted-dynamic-schema2
  - config: the attribute suppression profile could select the wrong lowering path, but the whole-program call happens before the profile is installed and both profiles fail identically
  - environment: Clang/host differences could reject generated C, but the refusal is returned by emitProgram before serialization or process execution
  - data: the gate corpus deliberately contains the exact exclusive-borrow chain that both Phase 11's independent knower and cgen classify as by-pointer; this shape is intentional evidence, not fixture drift
and_gate: no; the incorrect consumer disposition/migration fully accounts for the regression, while cut-m004 and the fixture shape are intentional boundary conditions

## Symptoms

expected: Phase 11 native zero-attribute controls continue to verify their admitted evidence after the Phase 16 public M004 cut, while the cut-m004 disposition remains unchanged.
actual: The automated Go regression gate fails five Phase 11 session tests because emitting their multi-function whole-program fixtures refuses the touch or touchTwo by-pointer bodies as unsupported this phase.
errors: TestPhase11ZeroAttributeGate, TestPhase11GateIsNonVacuous, TestPhase11GateCountsAdjacentWouldCarryFunctions, TestPhase11GateMutationKill, and TestPhase11SuppressionIsDiffLocal fail during current public emission with by-pointer unsupported diagnostics.
reproduction: Run GOCACHE=/tmp/ai-lang-go-cache go test ./...; the cgen package passes and the session package reports the five Phase 11 failures.
started: After the Phase 16 public cutover selected cut-m004 for all by-pointer families because Linux probe evidence was unavailable.

## Eliminated

- hypothesis: AttributeSuppressionProfile configuration accidentally selects the cut path.
  evidence: VerifyPhase11ZeroAttributeGate calls cgen.EmitNative for the whole program before SetAttributeSuppressionProfileForTest, and focused suppressed and justified cases return the same refusal.
  timestamp: 2026-09-23T21:57:29Z
- hypothesis: macOS/Clang environment differences cause the regression.
  evidence: emitProgram returns the named by-pointer refusal during supported-shape validation before C serialization; the focused tests fail in 0.00s without invoking the runner.
  timestamp: 2026-09-23T21:57:29Z
- hypothesis: the Phase 11 fixture accidentally drifted into the by-pointer family.
  evidence: testdata/phase11/multi_function_gate_corpus.lang documents touch as the exact would-have-carried-restrict shape, and phase11WouldCarryRestrict intentionally re-derives that same structural fact for the gate's non-vacuity proof.
  timestamp: 2026-09-23T21:57:29Z

## Evidence

- timestamp: 2026-09-23T00:00:00-04:00
  checked: Phase 16 UAT report
  found: Test 21 records the five named Phase 11 regression failures, while Tests 5, 7, 19, and 20 confirm cut-m004, a single public schema-2 authority, no legacy public fallback, and refusal-only M004 families.
  implication: A successful diagnosis must preserve the public refusal boundary and evaluate whether the historical Phase 11 evidence consumer is stale.
- timestamp: 2026-09-23T21:57:29Z
  checked: focused reproduction of the five named tests
  found: All five fail deterministically. Four fail in VerifyPhase11ZeroAttributeGate at the whole-program cgen.EmitNative call for touch or touchTwo; TestPhase11SuppressionIsDiffLocal fails at its direct single-function cgen.EmitNative helper for touch.
  implication: The shared failure mechanism is live emission of deliberately by-pointer Phase 11 evidence, not unrelated full-suite noise.
- timestamp: 2026-09-23T21:57:29Z
  checked: internal/compiler/cgen/cgen_program.go supported-shape validation and commit c33aebd
  found: Plan 16-06 intentionally broadened cut-m004 from single-function only to every program cardinality; emitProgram now refuses any selectsByPointerLowering or selectsByPointerLoweringSharedOnly function before serialization.
  implication: The emitter behavior is the recorded Phase 16 decision and must not be relaxed to repair these tests.
- timestamp: 2026-09-23T21:57:29Z
  checked: Phase 11 fixture and gate implementation
  found: The committed corpus deliberately makes touch the exact exclusive-borrow/reborrow-to-return shape; phase11WouldCarryRestrict identifies it as the required non-vacuous would-carry witness. VerifyPhase11ZeroAttributeGate nevertheless asks cgen.EmitNative to lower the whole program and each would-carry single-function counterfactual live.
  implication: The Phase 11 evidence claim remains meaningful, but its artifact acquisition strategy is obsolete after cut-m004.
- timestamp: 2026-09-23T21:57:29Z
  checked: testdata/phase16/public-emitter-consumers.json and its source-derived validator
  found: The three Phase 11 call sites are classified admitted-dynamic-schema2. The validator proves call-site inventory bijection and classification vocabulary, but does not derive whether each call's actual fixture is admitted or cut; representative provenance checks do not include these Phase 11 rows.
  implication: The incorrect semantic classification passed all registry mutation controls and prevented Plans 16-11 through 16-14 from routing these consumers through refusal-first frozen evidence.
- timestamp: 2026-09-23T21:57:29Z
  checked: migrated Phase 4/5 controls and Phase 16 planning contract
  found: Successfully migrated cut controls first call the public emitter and require the M004 refusal, then load digest- and canonical-program-bound historical C through a test-only loader. Phase 16 Plans 9-14 explicitly require later consumer migration without restoring a production fallback.
  implication: Phase 11 is an omission from the established migration pattern, not a reason to change cut-m004.
- timestamp: 2026-09-23T21:57:29Z
  checked: knowledge base
  found: No prior resolved entry matches this exact stale Phase 11 consumer classification; related Phase 15 entries reinforce that admission/refusal ordering and schema boundaries must be explicit.
  implication: Diagnosis rests on current direct evidence rather than an assumed known pattern.

## Resolution

root_cause: Phase 16's public-emitter consumer inventory semantically misclassified all three Phase 11 zero-attribute gate EmitNative sites as admitted-dynamic-schema2. Their fixed N=1/N=2 evidence intentionally contains by-pointer touch functions, so the gate and its diff-local helper still request the removed live lowering path and deterministically hit the correct cut-m004 refusal instead of using the project's refusal-first, provenance-bound frozen-evidence path.
fix: diagnosis only; no source or test changes
verification: Focused execution reproduced all five named failures; source tracing showed the errors arise at the three misclassified EmitNative sites before serialization, and comparison with migrated Phase 4/5 controls confirmed the missing refusal-first/frozen-evidence adaptation.
files_changed:
  - .planning/debug/phase16-m004-phase11-gates.md

## Post-diagnosis closure (2026-09-26)

Plan 16-21 migrated the Phase 11 by-pointer consumers to require the exact current M004 refusal before loading fixture-bound frozen evidence, and added provenance-bound inventory and mutation controls. Phase 16 UAT gap G-16-21-A is resolved by Plan 16-21; the Phase 16 verification confirms the refusal-first boundary and passes. M004 remains the recorded owner for any future by-pointer admission. The original diagnosis-only scope is preserved above; this section records the later closure evidence.
