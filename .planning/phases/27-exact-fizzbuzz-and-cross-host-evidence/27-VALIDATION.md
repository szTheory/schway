---
phase: "27"
slug: "exact-fizzbuzz-and-cross-host-evidence"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
status: draft
nyquist_compliant: false
wave_0_complete: false
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
| **Quick run command** | Focused `go test` selectors in the plan for the touched packages and `TestPhase27` cases |
| **Full suite command** | `go test ./...`; hosted CI also runs `go test -race ./...` |
| **Estimated runtime** | Not measured for Phase 27; keep per-task selectors focused and record observed latency during execution |

---

## Sampling Rate

- **After every task commit:** Run the focused package selector named by that task.
- **After every plan wave:** Run the Phase 27 focused regression selector covering that wave's changed semantics and its negative control.
- **Before `$gsd-verify-work`:** Require the full Go suite and hosted macOS/Linux Phase 27 receipts; include race results from hosted CI.
- **Max feedback latency:** Not measured yet; use focused selectors per task and record observed latency during execution.

---

## Requirement Verification Map

These rows name the concrete plan/task/wave owner and its task selector. Each runnable command appears in that task's `<verify>` with an adjacent `<fails_when>`. These checks are planned, not executed during planning.

| Requirement | Plan / Wave | Threat Ref | Secure Behavior | Test Type | Automated Command Target | Status |
|-------------|-------------|------------|-----------------|-----------|--------------------------|--------|
| U64-02 | 27-02 Tasks 1-2 / Wave 2; 27-05 Task 2 / Wave 5 | T-27-03, T-27-04 | Static zero remainder divisor is rejected; dynamic zero fails consistently before native C remainder; independent validators reject forged operations. | Unit + native integration + negative controls | 27-02 Tasks 1-2 automated selectors for equality, remainder, static zero, dynamic zero, and forged remainder; 27-05 Task 2 independent peer control | ⬜ pending |
| APP-08 | 27-01 Task 1 / Wave 1; 27-03 Tasks 1-2 / Wave 3; 27-05 Task 1 / Wave 5 | T-27-01, T-27-05, T-27-08 | FizzBuzz output is exact for 0, 1, 15, and 1,000; every emitted line ends in LF. | App integration | 27-03 Tasks 1-2 automated selectors for literal writes and exact FizzBuzz; 27-05 Task 1 literal per-engine golden answers | ⬜ pending |
| APP-09 | 27-01 Task 2 / Wave 1; 27-03 Task 2 / Wave 3; 27-04 Tasks 1-2 / Wave 4 | T-27-02, T-27-16, T-27-06, T-27-07 | Both apps reject 1,001 before output; a write crossing the lowered test limit is non-success and does not claim truncated success. | App boundary integration | 27-04 Tasks 1-2 automated selectors for exact/crossing byte limits and public status; 27-03 Task 2 FizzBuzz refusal | ⬜ pending |
| EVD-12 | 27-05 Tasks 1-2 / Wave 5; 27-06 Tasks 1-2 / Wave 5; 27-07 Task 2 / Wave 6 | T-27-08, T-27-09, T-27-10, T-27-11, T-27-12, T-27-13, T-27-14 | Each engine matches independent literal answers; reached mutations, repeated event identity, and evidence exhaustion are independently adjudicated; hosted native receipts identify revision and inputs. | Evidence integration + host matrix | 27-05/06 task automated selectors; hosted sh scripts/verify-phase27.sh; 27-07 Task 2 receipt gate | ⬜ pending |
| DX-16 | 27-07 Task 1 / Wave 6 | T-27-15 | One clean-checkout guide owns both programs' commands, expected results, limits, failure behavior, and evidence provenance; root README links it. | Documentation contract + clean-checkout integration | go test ./internal/compiler/session -run '^TestPhase27GuideContract$' -count=1; follow the guide commands from a fresh checkout | ⬜ pending |

---

## Wave 0 Requirements

Existing Go package and hosted CI infrastructure covers all requirement types. Phase-specific witnesses, expected answers, negative controls, documentation, and receipt tooling are implementation tasks; no test framework installation or shared-fixture setup is required before the first slice.

---

## Manual-Only Verifications

All semantic acceptance behaviors have automated checks. The macOS and Linux native receipts must come from actual hosted execution and cannot be replaced by source inspection or an unavailable-host record.

---

## Validation Sign-Off

- [ ] All tasks have an `<automated>` verify command or explicit Wave 0 dependency.
- [ ] Sampling continuity: no three consecutive tasks without automated verification.
- [ ] Wave 0 covers all missing test infrastructure references.
- [ ] No watch-mode flags.
- [ ] Focused feedback latency is measured during execution.
- [ ] `nyquist_compliant: true` set in frontmatter after validation.

**Approval:** pending
