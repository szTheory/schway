---
status: diagnosed
phase: 16-branch-match-emitter-port
source: [16-01-SUMMARY.md, 16-02-SUMMARY.md, 16-03-SUMMARY.md, 16-04-SUMMARY.md, 16-05-SUMMARY.md, 16-06-SUMMARY.md, 16-07-SUMMARY.md, 16-08-SUMMARY.md, 16-09-SUMMARY.md, 16-10-SUMMARY.md, 16-11-SUMMARY.md, 16-12-SUMMARY.md, 16-13-SUMMARY.md, 16-14-SUMMARY.md, 16-15-SUMMARY.md]
started: 2026-09-23T00:00:00Z
updated: 2026-09-24T00:00:00Z
---

## Current Test

number: 21
name: Run the Go integration and regression suite
expected: |
  `go test ./...` completes successfully, including compiler emission, session evidence, and project invariant checks.
awaiting: gap-closure planning

## Tests

### 1. Confirm ordinary tracked-transfer emission
expected: Direct emitProgram lowers and executes the ordinary tracked-transfer fixture as a schema-2 document. The summary cites TestProgramOrdinaryLinearTracer.
result: pass
source: automated

### 2. Confirm branch emission behavior
expected: Branch arms execute through direct emitProgram with schema-2 events, and branch admission preserves the documented graph, shape, and preflight diagnostic ordering.
result: pass
source: automated

### 3. Confirm payload match and defect behavior
expected: Payload match construction/destruction uses checker-derived tagged fields, and defect arms write a schema-2 terminal event before abort.
result: pass
source: automated

### 4. Review the restrict probe evidence
expected: The exact one-pointer C17 probe records source-shape, O0, O3, O3-LTO, sanitizer, and host-availability results. Linux evidence is unavailable on the macOS executor and is considered in the disposition.
context: Linux evidence is explicitly unavailable on this macOS executor and must be considered in the later disposition.
source: automated
result: pass

### 5. Confirm the by-pointer disposition
expected: The by-pointer admission boundary records exact source, macOS lane, Linux availability, and refusal-fence evidence before selecting a disposition. The recorded disposition is cut-m004 because required Linux evidence is unavailable.
context: The blocking-human checkpoint selected cut-m004 because the required Linux evidence is unavailable; automated evidence cannot substitute for that disposition.
result: pass
source: previously-recorded-user-decision-and-automated-evidence

### 6. Confirm the convergence and schema-2 evidence
expected: The five-shape characterization covers both public modes and records the intentional legacy-versus-direct protocol relation; golden-C provenance distinguishes the pre-cut state; direct-program schema-2 C agrees with interpreter results across O0, O3, and O3-LTO for three admitted fixtures.
result: pass
source: automated

### 7. Confirm the public cutover evidence
expected: Public production emission has one schema-2 authority and no legacy public fallback, while the atomic cutover is preserved as immutable history and M004 families remain refusal-only.
result: pass
source: automated

### 8. Confirm public-emitter consumer registry reconciliation
expected: The public-emitter registry exactly mirrors the current AST-derived direct-call inventory, with mutation controls for duplicate, stale, missing, invalid-classification, and missing-refusal-witness rows.
result: pass
source: automated

### 9. Confirm branch admission ordering
expected: Branch admission retains graph, shape, and preflight diagnostic ordering before serialization.
result: pass
source: automated
coverage_id: D2

### 10. Confirm ordinary-linear schema-2 event sequence
expected: Direct emitProgram lowers and executes the ordinary tracked-transfer fixture as a schema-2 document.
result: pass
source: automated
coverage_id: D1

### 11. Confirm derived resource serialization
expected: Schema-2 live_resources is rendered from a derived value and a seed changes generated serialization.
result: pass
source: automated
coverage_id: D2

### 12. Confirm branch program events
expected: Branch arms execute through direct emitProgram with schema-2 events.
result: pass
source: automated
coverage_id: D1

### 13. Confirm checker-derived payload lowering
expected: Payload match selection and construction/destruction lower through the program emitter using checker-derived tagged fields.
result: pass
source: automated
coverage_id: D1

### 14. Confirm defect terminal evidence
expected: Defect arms write a schema-2 terminal event before abort, with _Noreturn restricted to lang_defect.
result: pass
source: automated
coverage_id: D2

### 15. Confirm the restrict admission fence
expected: Candidate structural fence accepts only the exact shape and rejects every D-16-07 extension with a live bypass control.
result: pass
source: automated
coverage_id: D2

### 16. Confirm five-shape convergence characterization
expected: The five-shape characterization covers both public modes and records the intentional legacy-versus-direct protocol relation.
result: pass
source: automated
coverage_id: D1

### 17. Confirm golden-C provenance
expected: Golden-C provenance distinguishes real pre-cut state from the atomic post-cut baseline.
result: pass
source: automated
coverage_id: D2

### 18. Confirm admitted-fixture native agreement
expected: Direct-program schema-2 C agrees with the interpreter across O0, O3, and O3-LTO for the three admitted fixtures.
result: pass
source: automated
coverage_id: D3

### 19. Confirm the single public emission authority
expected: Public production emission has one schema-2 authority and no legacy public fallback.
result: pass
source: automated
coverage_id: D1

### 20. Confirm refusal-only M004 boundary
expected: The atomic cutover is preserved as immutable history and M004 families remain refusal-only.
result: pass
source: automated
coverage_id: D2

### 21. Run the Go integration and regression suite
expected: `go test ./...` completes successfully, including compiler emission, session evidence, and project invariant checks.
result: issue
reported: "GOCACHE=/tmp/ai-lang-go-cache go test ./... exited 1. Failures include Phase 11 zero-attribute native gates refusing by-pointer fixtures; TestValidationRowGradesAreEarnedOverArchivedCorpus reporting a corpus digest mismatch; TestVerificationGroundednessFrontierIsPinned reporting new unowned rows; TestDebtRegistersAreWellFormed finding missing Phase 13 witness tests; ProbeMachine failures in Phase 6 budget tests; and LANGUAGE-MATURITY.md corpus counts differing from the re-derived counts."
severity: blocker

## Summary

total: 21
passed: 20
issues: 1
pending: 0
skipped: 0

## Gaps

- gap_id: G-16-21-A
  truth: Phase 11 native zero-attribute controls continue to verify their admitted evidence after the Phase 16 public M004 cut.
  status: failed
  reason: "Automated regression tests TestPhase11ZeroAttributeGate, TestPhase11GateIsNonVacuous, TestPhase11GateCountsAdjacentWouldCarryFunctions, TestPhase11GateMutationKill, and TestPhase11SuppressionIsDiffLocal fail because current by-pointer emission is refused."
  severity: blocker
  test: 21
  root_cause: ""
  root_cause: "Phase 16's public-emitter consumer inventory semantically misclassified three Phase 11 zero-attribute gate EmitNative sites as admitted-dynamic-schema2. Their N=1/N=2 fixtures intentionally include by-pointer touch functions, so they still request a live lowering path correctly refused by the cut-m004 emitter."
  artifacts: [".planning/debug/phase16-m004-phase11-gates.md", "internal/compiler/session/session_phase11_gate.go", "internal/compiler/session/session_phase11_gate_test.go", "internal/compiler/session/session_phase11_differential_test.go", "testdata/phase16/public-emitter-consumers.json"]
  missing: ["Migrate all three Phase 11 by-pointer evidence consumers to assert the current public refusal before loading digest- and canonical-program-bound frozen evidence; keep live structural/non-vacuity controls and live N=0 coverage; correct registry classifications and add provenance-bound mutation controls; preserve the recorded cut-m004 boundary and do not restore a production fallback."]
  debug_session: ".planning/debug/phase16-m004-phase11-gates.md"
- gap_id: G-16-21-B
  truth: Archived validation and groundedness evidence matches the current corpus and assigns every retained verification command.
  status: failed
  reason: "TestValidationRowGradesAreEarnedOverArchivedCorpus reports a corpus-pair digest mismatch. TestVerificationGroundednessFrontierIsPinned and TestVerificationGroundednessThreeClassesAreEmpty report newly surfaced unowned commands in Phase 14, 16, and 17 validation artifacts."
  severity: blocker
  test: 21
  root_cause: ""
  root_cause: "Committed test-index changes (the Phase 17 B1 witness replacement and five new cgen two-type tests) plus committed scanner-visible documentation changed derived inputs after snapshots were sealed. The Phase 16 corpus manifest/run record and Phase 14 pinnedFrontier/r2bLandingPhases were not refreshed atomically: the live 31-pair digest is 07e30b53013291e76df6cd4d01d06ddbf4184be0fa33c230000aa9c17d4814cb vs archived 85a6d72957ccf508d16259f370126e2d44c938526d17a0ac9f7f0e83f7e5976d, and six scanner findings are unpinned, including five R2b rows without P20 ownership."
  artifacts: [".planning/debug/phase16-validation-frontier-drift.md", "internal/compiler/session/evidence_grade_test.go", "internal/compiler/session/verification_groundedness_test.go", "testdata/phase16/validation-corpus-run-record.jsonl", "testdata/phase16/validation-corpus-run-record.manifest.json", ".planning/phases/14-evidence-instrument-and-honest-scoping/14-VALIDATION.md", ".planning/phases/16-branch-match-emitter-port/16-RESEARCH.md", ".planning/phases/16-branch-match-emitter-port/16-VALIDATION.md", ".planning/phases/17-return-type-parameter-type/17-VERIFICATION.md"]
  missing: ["Regenerate and validate the corpus manifest/run record from the live 31-pair producer, preserving its completion witness and real execution results; reconcile the six current groundedness findings with exact frontier equality and assign every surviving R2b row under the existing P20/QLT-10 ownership policy; resolve the unparseable research command and keep R1/R2/R3 empty without weakening scanner gates."]
  debug_session: ".planning/debug/phase16-validation-frontier-drift.md"
- gap_id: G-16-21-C
  truth: Debt register evidence references executable tests and machine-budget probes succeed in the supported host environment.
  status: failed
  reason: "TestDebtRegistersAreWellFormed finds missing Phase 13 witness test names. TestBudgetLaneCarriesMachineIDAndVerdict, TestBudgetAuditRefusesUndeclaredMachine, and TestQLT02InterproceduralGrowthExponent fail at ProbeMachine."
  severity: blocker
  test: 21
  root_cause: ""
  root_cause: "Two independent causes: PHASE-13-DEBT.md D-13-02b and D-13-10a still cite removed TestB1BlameIsStructurallyUnreachable after Phase 17 replaced it with TestPhase17B1RequiresUnverifiableDeclaredContract; and three budget tests call the real Darwin ProbeMachine whose sysctl CPU-model subprocess is denied by this sandbox, yielding measure.probe_failed."
  artifacts: [".planning/debug/phase16-debt-and-machine-probe.md", ".planning/milestones/M002-phases/13-explain-function-attribution/PHASE-13-DEBT.md", ".planning/UNREACHABLE-CLAIMS.md", "internal/compiler/session/session_phase6_budget_test.go", "internal/compiler/measure/machine.go"]
  missing: ["Reconcile historical Phase 13 witness references with current Phase 17 evidence in the source register and generated view; make budget tests deterministic across host policy by injecting machine facts/probe results while retaining an explicit observation path for real host probing, without claiming the restricted host was successfully probed."]
  debug_session: ".planning/debug/phase16-debt-and-machine-probe.md"
- gap_id: G-16-21-D
  truth: LANGUAGE-MATURITY.md reports corpus counts that match the current source fixtures.
  status: failed
  reason: "TestLanguageMaturityCountsAreCurrent derives 133 programs and 4478 lines, while the document states 128 and 4311."
  severity: major
  test: 21
  root_cause: ""
  root_cause: "The independently derived source-fixture corpus grew to 133 programs and 4,478 lines after LANGUAGE-MATURITY.md's last recorded snapshot of 128 programs and 4,311 lines."
  artifacts: [".planning/debug/phase16-debt-and-machine-probe.md", ".planning/LANGUAGE-MATURITY.md", "internal/compiler/session/self_describing_docs_test.go"]
  missing: ["Refresh the dated maturity snapshot to the derived 133-program/4,478-line counts and retain the machine-checked count guard so future fixture drift is detected automatically."]
  debug_session: ".planning/debug/phase16-debt-and-machine-probe.md"

- gap_id: G-16-21-E
  truth: Every scanner-visible verification command in the Phase 16 report has an owned R2b landing phase, and the groundedness suite passes against the current report.
  status: failed
  reason: "Fresh GOCACHE=/tmp/ai-lang-gocache go test ./... run fails only TestVerificationGroundednessFrontierIsPinned and TestVerificationGroundednessThreeClassesAreEmpty: the command at 16-VERIFICATION.md:143 is an unowned R2b finding."
  severity: blocker
  test: 21
  root_cause: "Groundedness classification splits the literal selector at each `|`; the grouped alternation leaves invalid standalone branches and is classified as an unowned R2b command. The report is semantically accurate, but its selector form trips this scanner."
  artifacts:
    - path: ".planning/phases/16-branch-match-emitter-port/16-VERIFICATION.md"
      issue: "Line 143 uses a grouped alternation that the groundedness scanner splits into invalid branches."
    - path: "internal/compiler/session/verification_groundedness_test.go"
      issue: "The per-branch classifier intentionally splits selectors on `|` and requires each branch to be independently grounded."
  missing:
    - "Express the same two test names as independently anchored alternation branches, then rerun the groundedness tests and complete Go suite."
  debug_session: ".planning/debug/phase16-groundedness-owner.md"

## Verification Blockers

- kind: commit_claim_mismatch
  severity: blocker
  status: reconciled_by_plan_snapshot
  note: "The base-to-current-HEAD count includes later plan commits and unrelated concurrent work. Measuring from each summary's plan_head_before to the commit that recorded that summary yields the claimed count or the single allowed summary-metadata commit for 16-01 through 16-14. Plan 16-15's three task commits are present; its later execution-state/measurement metadata commits postdate the measured task ledger. The GSD aggregate check should use the summary snapshot, not current HEAD, when reconciling historical plan claims."
