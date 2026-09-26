---
status: resolved
trigger: "Diagnose a Phase 15 UAT gap. Do not fix code; goal is find_root_cause_only."
created: 2026-09-22T09:16:43Z
updated: 2026-09-26T19:47:39Z
---

## Current Focus

hypothesis: Confirmed: Test 1 is a traceability/oracle gap, not an observed schema-validation implementation failure. Missing 15-01 coverage metadata forces legacy human UAT routing, and its structural negative test asserts refusal only, not actionable diagnostic identity or the complete JSON decode seam.
test: Completed classifier differential (15-01 versus 15-02 through 15-05), uncached schema tests, validator/test source trace, and CI entrypoint inspection.
expecting: Confirmed by direct observations recorded below.
next_action: Return diagnose-only root cause and recommend adding machine-readable coverage plus a wire-boundary diagnostic regression; do not modify implementation.
bug_class: bohrbug
candidate_causes:
  - "config/evidence: 15-01-SUMMARY.md lacks coverage metadata consumed by verify-work"
  - "code/test: schema rejection tests may exist but assert only non-nil errors, not actionable diagnostic content"
  - "environment/CI: CI may not invoke the relevant package/tests despite local verification"
and_gate: "Yes for the full reported automation gap: missing summary coverage metadata causes human routing, while the existing test's refusal-only/direct-call oracle prevents the complete actionable wire-seam truth from being honestly auto-certified. Neither condition demonstrates a production admission bug."

## Symptoms

expected: A multi-function execution document with canonical invocation identities is accepted, while a non-canonical invocation, duplicate activation identity, invalid event kind, or invalid caller-owned callee identity is refused with an actionable diagnostic.
actual: "integration/e2e/smoke/seam test... automate the world devops mindset shift left, goal is 0 human verificaiton / uat required (even shift to CI ... iff recurring value there)"
errors: No runtime/schema error reported; the reported defect asks for automated integration/e2e/smoke/seam and recurring CI evidence.
reproduction: Test 1 in Phase 15 UAT.
started: Observed during Phase 15 UAT on 2026-09-22; implementation regression timing not reported.

## Eliminated

- hypothesis: The Schema 2 validator currently accepts the reported malformed documents or rejects the canonical document.
  evidence: Uncached TestInvocationGrammar, TestExecutionLegacyBytesFrozen, and TestValidateExecutionSchema2 all pass; source inspection shows explicit canonical parsing, pair uniqueness, callee-field ownership, and closed event-kind branches.
  timestamp: 2026-09-22T09:27:00Z

- hypothesis: CI never executes the existing schema-admission tests.
  evidence: .github/workflows/ci.yml runs go test ./... and go test -race ./... on Ubuntu and macOS, and the tests are ordinary non-skipped Go tests; a later recorded Phase 16 validation corpus also contains successful executions of all three named tests.
  timestamp: 2026-09-22T09:27:00Z

## Evidence

- timestamp: 2026-09-22T09:16:43Z
  checked: Phase 15 UAT and project state
  found: Test 1 is marked issue solely from a request to automate integration/e2e/smoke/seam verification; Tests 2-10 are automated passes, while Tests 1, 11, and 12 share the same automation-focused report.
  implication: The reported gap is systemic evidence/automation coverage, not direct evidence that Schema 2 admission behavior is incorrect.

- timestamp: 2026-09-22T09:23:00Z
  checked: Phase 15 Plan 01, Summary 01, validation map, native validator/tests, and GitHub CI
  found: Plan 01 requires and implements TestInvocationGrammar, TestExecutionLegacyBytesFrozen, and TestValidateExecutionSchema2; 15-VALIDATION.md maps those tests as green, and CI runs go test ./... plus go test -race ./... on Linux and macOS. Unlike summaries 02-05, 15-01-SUMMARY.md contains no coverage block for the verify-work classifier.
  implication: The recurring tests are present and CI-reachable, but the UAT automation seam cannot associate Plan 01's deliverable with them, so it falls back to a human checkpoint.

- timestamp: 2026-09-22T09:23:00Z
  checked: TestValidateExecutionSchema2 assertions
  found: The positive document and all named negative mutations are automated, but each negative subtest asserts only err != nil; it does not assert the diagnostic string or a stable diagnostic code.
  implication: Schema acceptance/refusal is tested, while the expected actionable-diagnostic portion is not regression-pinned by this test; this is a test-oracle gap, not evidence that production validation admits invalid documents.

- timestamp: 2026-09-22T09:27:00Z
  checked: GSD uat.classify-coverage across all seven Phase 15 summaries
  found: 15-01, 15-06, and 15-07 return mode=legacy, total=0, all_auto_covered=false; 15-02 through 15-05 return mode=coverage and all_auto_covered=true. These exactly align with UAT issues 1, 11, and 12 versus automated passes 2-10.
  implication: The systemic UAT failures are deterministically caused by omitted machine-readable coverage blocks in three summaries, not by failing implementation tests.

- timestamp: 2026-09-22T09:27:00Z
  checked: Uncached focused schema-admission run
  found: TestInvocationGrammar, TestExecutionLegacyBytesFrozen, and TestValidateExecutionSchema2 all passed, including missing invocation/callee, extra callee, duplicate pair, unknown kind, unknown schema, and same ID across distinct invocations.
  implication: Existing structural schema behavior is green; diagnosis must not mislabel the UAT automation request as a validator logic defect.

- timestamp: 2026-09-22T09:27:00Z
  checked: Plan 01 promised seam versus implemented regression oracle
  found: The plan requires model to canonical JSON to decoder to validator and actionable diagnostics, but TestValidateExecutionSchema2 calls unexported validateExecution directly and checks only whether err is non-nil. TestExecutionPeerFailuresAreActionable pins diagnostic tokens and event index, but it belongs to the separate program-aware peer truth and does not cover native JSON decoding for Test 1.
  implication: A real seam-level regression could break JSON decode/ToolError wrapping or degrade diagnostic specificity while the Test 1 unit oracle remains green; this explains the user's request for integration/e2e/seam automation.

## Resolution

root_cause: "Phase 15 Plan 01's automated schema checks were not published in 15-01-SUMMARY.md's machine-readable coverage block, so verify-work deterministically routed the deliverable to human UAT despite CI already running the tests; additionally, TestValidateExecutionSchema2 bypasses canonical JSON decoding and asserts only non-nil errors, leaving the actionable wire-boundary diagnostic portion of the UAT truth without a faithful automated oracle. The production Schema 2 admission branches themselves are present and pass their current positive/negative tests."
fix:
verification: "Diagnose-only: classifier differential reproduced legacy routing for 15-01/06/07 and automated routing for 15-02..05; focused uncached schema tests passed; CI and recorded corpus execution were inspected."
files_changed: []

## Post-diagnosis closure (2026-09-26)

Plan 15-08 added the canonical-JSON-to-decoder-to-ToolError admission regression, distinct refusal diagnostics, and machine-readable Plan 01 coverage. The seam is included in the Phase 15 CI evidence job. Phase 15 UAT is complete (12/12, zero issues); gap G-15-1 is recorded resolved by Plan 15-08. The original diagnosis-only scope is preserved above; this section records the later closure evidence.
