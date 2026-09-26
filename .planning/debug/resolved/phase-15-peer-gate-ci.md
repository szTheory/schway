---
status: resolved
trigger: "Diagnose a Phase 15 UAT gap. Do not fix code; goal is find_root_cause_only."
created: 2026-09-22T09:20:55Z
updated: 2026-09-26T19:47:39Z
---

## Current Focus

hypothesis: Confirmed: Phase 15 Plan 06 omitted verify-work coverage metadata, causing deterministic human-UAT fallback despite a working, CI-reachable peer gate; the test oracle also lacks a direct /0-and-/1 preservation regression through the new wrapper.
test: Completed classifier differential, uncached focused peer/fault tests, wrapper call-site audit, and CI workflow inspection.
expecting: Confirmed by direct observations below.
next_action: Return diagnose-only root cause; recommend summary coverage metadata plus a legacy-wrapper regression and a current named phase gate, without changing implementation.
bug_class: bohrbug
candidate_causes:
  - "config/evidence: 15-06-SUMMARY.md omits the machine-readable coverage block required by verify-work"
  - "code/test: focused gate coverage may omit the /0 and /1 preservation clause"
  - "environment/CI: the tests may be absent from CI despite local execution"
and_gate: "Yes for complete zero-human-UAT closure: missing metadata alone causes the human prompt, while direct legacy-wrapper coverage is also needed to auto-certify the full expected truth; neither indicates the production peer gate is defective."

## Symptoms

expected: The program-aware four-engine Schema 2 comparison refuses a corrupted engine document before it can report agreement, while /0 and /1 comparison behavior remains unchanged.
actual: "integration/e2e/smoke/seam test... automate the world devops mindset shift left, goal is 0 human verificaiton / uat required (even shift to CI ... iff recurring value there)"
errors: No runtime error reported; UAT identifies missing automated integration/e2e/smoke/seam coverage.
reproduction: Test 11 in Phase 15 UAT.
started: Identified during Phase 15 UAT on 2026-09-22.

## Eliminated

- hypothesis: The actual Schema 2 program-aware comparison gate is missing or permits matching corrupted documents to report agreement.
  evidence: Phase5CompareProgramEngines validates every /2 document before pair comparison, and the uncached peer/fault suite passed all five focused tests, including matching forged documents and both producer/peer fault directions.
  timestamp: 2026-09-22T09:32:00Z
- hypothesis: Existing peer-gate tests have no recurring CI execution.
  evidence: .github/workflows/ci.yml runs ordinary go test ./... and go test -race ./... on Ubuntu and macOS; all focused controls are ordinary non-skipped tests in the session package.
  timestamp: 2026-09-22T09:32:00Z

## Evidence

- timestamp: 2026-09-22T09:24:30Z
  checked: Phase5CompareProgramEngines and its focused tests
  found: Every lang.execution/2 engine document is passed to executionpeer.Validate before Phase5CompareEngines; TestSchema2ComparisonRequiresPeerVerdict corrupts all matching documents and expects executionpeer.malformed_invocation. Legacy documents skip the peer and retain Phase5CompareEngines.
  implication: The reported gap is not absence of the production peer gate.
- timestamp: 2026-09-22T09:24:30Z
  checked: Phase 15 validation and summaries
  found: 15-VALIDATION records focused automated commands for comparator/peer controls and says no manual-only verification; 15-07-SUMMARY explicitly declares automated release evidence sufficient.
  implication: Phase artifacts intended this behavior to be automation-owned, so interactive UAT classification conflicts with the recorded validation contract.
- timestamp: 2026-09-22T09:24:30Z
  checked: .github/workflows/ci.yml
  found: The checks job runs go test ./... and go test -race ./... on Linux and macOS, but the separately named phase-gate job is still hard-pinned to scripts/verify-phase6.sh and describes Phase 6 as current.
  implication: The Go tests are transitively in CI, but there is no current Phase 15 named smoke/release seam; explicit phase-gate drift is systemic and obscures coverage provenance.
- timestamp: 2026-09-22T09:31:00Z
  checked: GSD uat.classify-coverage for Phase 15 summaries
  found: 15-06-SUMMARY.md and 15-07-SUMMARY.md return mode=legacy, total=0, all_auto_covered=false, while 15-05-SUMMARY.md returns mode=coverage with all three deliverables auto-covered.
  implication: The exact UAT issue is reproducible as deterministic traceability fallback: verify-work cannot associate Plan 06's prose with its passing tests.
- timestamp: 2026-09-22T09:31:00Z
  checked: all test call sites of Phase5CompareProgramEngines
  found: Schema-2 positive/corruption paths are exercised, but no test invokes the program-aware wrapper with both Schema0 and Schema1 documents to assert they bypass the peer and retain legacy comparison behavior.
  implication: Existing tests strongly prove the new /2 gate but do not faithfully cover the full Test 11 clause that /0 and /1 behavior remains unchanged.
- timestamp: 2026-09-22T09:32:00Z
  checked: Uncached focused Plan 06 peer-gate and bidirectional fault suite
  found: TestComparisonFieldRoutingIsExhaustive, TestInvocationFieldsAreCompared, TestSchema2ComparisonRequiresPeerVerdict, TestExecutionProducerFaultIsCaughtByPeer, and TestExecutionPeerAcceptanceFaultIsCaughtByControl all passed (2.940s package result).
  implication: The production /2 gate and its corruption controls are green; the UAT issue is automation traceability/completeness, not a demonstrated gate failure.

## Resolution

root_cause: "15-06-SUMMARY.md omits the machine-readable coverage block that verify-work requires, so the Schema 2 peer-gate deliverable deterministically falls back to human UAT even though its focused tests already run under CI's cross-platform go test ./... jobs. The focused tests prove /2 peer refusal and fault-seam non-inertness, but no test directly drives Phase5CompareProgramEngines with both /0 and /1 documents, so the full 'legacy behavior remains unchanged' clause is not faithfully automated. The separately named CI phase-gate is also stale at verify-phase6.sh; this obscures current phase provenance but does not prevent the ordinary Go tests from running."
fix:
verification: "Diagnose-only: uat.classify-coverage reproduced mode=legacy,total=0 for 15-06 versus all-auto coverage for 15-05; the five focused peer-gate/fault tests passed uncached; all program-aware wrapper test call sites and CI jobs were inspected."
files_changed: []

## Post-diagnosis closure (2026-09-26)

Plan 15-09 added direct Schema 0/1 wrapper-preservation coverage, machine-readable Plan 06 coverage, and the current cross-platform CI aggregate. Phase 15 UAT is complete (12/12, zero issues); gap G-15-11 is recorded resolved by Plan 15-09. The original diagnosis-only scope is preserved above; this section records the later closure evidence.
