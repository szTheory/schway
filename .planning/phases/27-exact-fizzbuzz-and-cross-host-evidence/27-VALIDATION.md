---
phase: "27"
slug: "exact-fizzbuzz-and-cross-host-evidence"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-10-03"
---

# Phase 27 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go standard testing package (Go 1.24) |
| **Config file** | Go module defaults; no separate test configuration identified |
| **Quick run command** | Focused Go test selectors named in each plan for the touched packages and the Phase 27 cases |
| **Full suite command** | The hosted CI full and race lanes on Linux and macOS; no full suite is run locally |
| **Estimated runtime** | Not measured for Phase 27; keep per-task selectors focused and record observed latency during execution |

---

## Sampling Rate

- **After every task commit:** Run the focused package selector named by that task.
- **After every plan wave:** Run the Phase 27 focused regression selector covering that wave's changed semantics and its negative control.
- **Before `$gsd-verify-work`:** Require the full Go suite and hosted macOS/Linux Phase 27 receipts; include race results from hosted CI.
- **Max feedback latency:** about 10 seconds for the focused Phase 27 selectors across all seven packages (measured 2026-10-10).

---

## Requirement Verification Map

These rows name the plan/task/wave owner and the concrete test selectors that cover each requirement. Commands were run locally on macOS arm64 on 2026-10-10 (all listed packages `ok`, 47 top-level `TestPhase27*` tests passed, none skipped). Hosted full and race suites on macOS and Ubuntu passed at 7936dac in run 38068888687 (recorded in Phase 27.1's `27.1-HOST-EVIDENCE.md`).

| Requirement | Plan / Wave | Threat Ref | Secure Behavior | Test Type | Automated Command Target | Status |
|-------------|-------------|------------|-----------------|-----------|--------------------------|--------|
| U64-02 | 27-03 Tasks 1-2 / Wave 3; 27-04 Tasks 1-2 / Wave 4; 27-08 Task 2 / Wave 8 | T-27-03, T-27-04 | Static zero remainder divisor is rejected; dynamic zero fails consistently before native C remainder; independent validators reject forged operations. | Unit + native integration + negative controls | `go test ./internal/compiler/session -run 'TestPhase27StaticZero\|TestPhase27ScalarZero\|TestPhase27DynamicZero\|TestPhase27ForgedRemainder\|TestPhase27EqualityRemainder\|TestPhase27OperatorSyntax\|TestPhase27OverflowZeroControls' -count=1` | ✅ green |
| APP-08 | 27-01 Tasks 1-2 / Wave 1; 27-02 Tasks 1-2 / Wave 2; 27-05 Tasks 1-2 / Wave 5; 27-06 Tasks 1-2 / Wave 6; 27-08 Task 1 / Wave 8 | T-27-01, T-27-02, T-27-05, T-27-16, T-27-08 | FizzBuzz output is exact for 0, 1, 15, and 1,000; every emitted line ends in LF. | App integration | `go test ./internal/compiler/session -run 'TestPhase27WriteAdmission\|TestPhase27WritePeerAdmission\|TestPhase27TracerBoundaries\|TestPhase27TracerFizzBuzz1\|TestPhase27LiteralWriteAdmission\|TestPhase27LiteralTextApp\|TestPhase27LiteralEscapeFormat\|TestPhase27FizzBuzzExact\|TestPhase27ExactAnswerMatrix' -count=1`; `go test ./cmd/schway -run 'TestPhase27PublicAppCLI$' -count=1`; `go test ./internal/compiler/core -run 'TestPhase27TextWriteOperationInventory$' -count=1` | ✅ green |
| APP-09 | 27-02 Task 2 / Wave 2; 27-06 Task 2 / Wave 6; 27-07 Tasks 1-2 / Wave 7; 27-08 Task 2 / Wave 8 | T-27-02, T-27-16, T-27-06, T-27-07 | Both apps reject 1,001 before output; a write crossing the lowered test limit is non-success and does not claim truncated success. | App boundary integration | `go test ./cmd/schway -run 'TestPhase27AppBoundary\|TestPhase27OutputPublicStatus\|TestPhase27PublicAppCLI' -count=1`; `go test ./internal/compiler/cgen ./internal/compiler/interp -run 'TestPhase27Output' -count=1`; `go test ./internal/compiler/native -run 'TestPhase27OutputEvidenceExactLimit\|TestPhase27ApplicationReturnShapeCoupling' -count=1`; `go test ./internal/compiler/session -run 'TestPhase27OutputEvidenceSeparation$' -count=1` | ✅ green |
| EVD-12 | 27-08 Tasks 1-2 / Wave 8; 27-09 Tasks 1-2 / Wave 8; 27-10 Task 2 / Wave 9 | T-27-08, T-27-09, T-27-10, T-27-11, T-27-12, T-27-13, T-27-14 | Each engine matches independent literal answers; reached mutations, repeated event identity, and evidence exhaustion are independently adjudicated; hosted native receipts identify revision and inputs. | Evidence integration + host matrix | `go test ./internal/compiler/session -run 'TestPhase27ExactAnswerMatrix\|TestPhase27ReachedWrongRemainder\|TestPhase27ReachedWrongArithmetic\|TestPhase27SkippedIteration\|TestPhase27RepeatedEventIdentity\|TestPhase27LegacyDocumentCompatibility\|TestPhase27CIHostMatrix\|TestPhase27HostReceiptGate' -count=1`; `go test ./internal/compiler/interp ./internal/compiler/executionpeer ./internal/compiler/cgen ./internal/compiler/native -run 'TestPhase27RepeatedEventIdentity\|TestPhase27PeerOccurrenceOrder\|TestPhase27EvidenceCapacity' -count=1`; hosted receipts: `27-HOST-EVIDENCE.md` (run 38054848320 at 6224f8b) and current-source run 38068888687 at 7936dac | ✅ green |
| DX-16 | 27-10 Task 1 / Wave 9 | T-27-15 | One clean-checkout guide owns both programs' commands, expected results, limits, failure behavior, and evidence provenance; root README links it. | Documentation contract + clean-checkout integration | `go test ./internal/compiler/session -run '^TestPhase27GuideContract$' -count=1` (the test checks links and tracked paths and executes the pinned gate command) | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

This table keeps its `Automated Command Target` header, so it stays outside the graded evidence corpus (`primaryValidationTable` keys on an `Automated Command` header). Enrolling it would need Grade and Non-inertness columns and a regenerated hosted corpus run record; that is out of scope for this validation.

---

## Wave 0 Requirements

Existing Go package and hosted CI infrastructure covers all requirement types. Phase-specific witnesses, expected answers, negative controls, documentation, and receipt tooling are implementation tasks; no test framework installation or shared-fixture setup is required before the first slice.

---

## Manual-Only Verifications

All semantic acceptance behaviors have automated checks. The macOS and Linux native receipts must come from actual hosted execution and cannot be replaced by source inspection or an unavailable-host record; they are recorded in `27-HOST-EVIDENCE.md` and, for current source, Phase 27.1's `27.1-HOST-EVIDENCE.md`. The clean-checkout walk-through of the guide is covered by `TestPhase27GuideContract` plus the hosted receipt lanes, not by a separate manual run.

---

## Validation Sign-Off

- [x] All tasks have an `<automated>` verify command or explicit Wave 0 dependency.
- [x] Sampling continuity: no three consecutive tasks without automated verification.
- [x] Wave 0 covers all missing test infrastructure references.
- [x] No watch-mode flags.
- [x] Focused feedback latency is measured: about 10 seconds for all seven packages locally (session 8.6s).
- [x] `nyquist_compliant: true` set in frontmatter after validation.

**Approval:** validated 2026-10-10 by `/gsd-validate-phase 27`: 5/5 requirements covered, 0 gaps, 0 escalated.

## Validation Audit 2026-10-10

| Metric | Count |
|---|---|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |
