---
phase: "21"
slug: "native-emission-ownership-and-resource-discharge-m004"
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-25"
---

# Phase 21 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` / `go test` |
| **Config file** | `go.mod` |
| **Quick run command** | `go test ./internal/compiler/session ./internal/compiler/cgen -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | Measure in Phase 21; do not assume a value |

---

## Sampling Rate

- **After every task commit:** Run the task's focused Go test command.
- **After every plan wave:** Run `go test ./... -count=1`.
- **Before `$gsd-verify-work`:** Full suite must be green.
- **Max feedback latency:** Record measured focused and full-suite durations during execution.

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 21-01-01 | 01 | 1 | Phase goal | — | All modeled exits have an explicit, checked disposition; admitted contract paths require discharge evidence | unit | `go test ./internal/compiler/session -run '^TestPhase21ResourceDischargeContract' -count=1 -v` | ❌ W0 | ⬜ pending |
| 21-02-01 | 02 | 1 | Phase goal | — | Legacy foreign/by-pointer lowering bodies are absent while public emission remains single-law and fail-closed | unit | `go test ./internal/compiler/cgen -run '^TestPhase21LegacyEmitterBodiesRetired$' -count=1 -v` | ❌ W0 | ⬜ pending |
| 21-02-02 | 02 | 1 | Phase goal | — | Existing whole-program foreign and by-pointer refusal gates remain green | unit | `go test ./internal/compiler/cgen -run 'TestPublicDispatchUsesOnlyEmitProgram|TestProgramBorrowedByPointerDisposition|TestProgramBranchValidationOrder' -count=1 -v` | ✅ | ⬜ pending |
| 21-03-01 | 03 | 1 | Phase goal | — | One emitted multi-function fixture is compared at `-O3` and `-O3 -flto` using the existing comparator | integration | `go test -tags=phase21_lto_evidence ./internal/compiler/session -run '^TestPhase21EmittedMultiFunctionLTOComparison$' -count=1 -v` | ❌ W0 | ⬜ pending |
| 21-03-02 | 03 | 1 | Phase goal | — | Existing seeded semantic control still detects comparator divergence | unit | `go test ./internal/compiler/session -run '^TestPhase16EmitterPortSemanticGuardIsNotInert$' -count=1 -v` | ✅ | ⬜ pending |
| 21-04-01 | 04 | 2 | Phase goal | — | Debt records and generated claims view reflect measured/retired evidence without widening the claim | unit | `go test ./internal/compiler/session -run 'TestDebtRegistersAreWellFormed|TestUnreachableClaimsViewIsCurrent' -count=1 -v` | ✅ | ⬜ pending |

---

## Wave 0 Requirements

- Phase 21 plans create the contract artifact, structural tests, one-shot build-tagged LTO comparison, and retirement guard before their focused commands run.
- Existing Go and Clang toolchain are present; no new framework or package installation is required.

---

## Manual-Only Verifications

All phase behaviors have automated verification. Human judgment is reserved for an explicit user decision if a later emitter-admission phase proposes changing the locked refusal boundary.

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency measured and recorded
- [ ] `nyquist_compliant: true` set in frontmatter after execution

**Approval:** pending
