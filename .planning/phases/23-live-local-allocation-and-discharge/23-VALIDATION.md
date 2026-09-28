---
phase: "23"
slug: "live-local-allocation-and-discharge"
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-27"
---

# Phase 23 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing package (Go 1.24); native integration uses installed Clang |
| **Config file** | `go.mod` |
| **Quick run command** | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/pathoracle -run '^TestPhase23' -count=1` |
| **Full suite command** | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./...` |
| **Estimated runtime** | Measure in Wave 0; keep the focused gate within the existing CI time budget |

## Sampling Rate

- **After every task commit:** Run the focused Phase 23 test for the changed compiler, peer, native, session, or CLI package.
- **After every plan wave:** Run the Phase 23 focused aggregate, including public native app and independent-observer controls added in Wave 0.
- **Before `$gsd-verify-work`:** Run `go test ./...` and obtain passing focused native receipts from both the existing macOS and Linux CI lanes.
- **Max feedback latency:** Measure in Wave 0; avoid repeating full or sanitizer suites when an existing CI lane already owns the same evidence question.

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 23-01-01 | 01 | 1 | FFI-03 | ASVS L1 | Each acquire/use/release operation uses its own checked signature, operand mode, failure facts, and release pairing. | unit + ABI | `go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/cgen ./internal/compiler/native -run '^TestPhase23' -count=1` | ❌ Wave 0 | ⬜ pending |
| 23-01-02 | 01 | 1 | RES-07 | ASVS L1 | Empty, oversized, unreadable, malformed, allocation-failure, and partial-acquire outcomes are bounded and typed; failed acquisition leaves no Lang owner and frees partial adapter storage. | unit + native adapter | `go test ./internal/compiler/native ./internal/compiler/session -run '^TestPhase23Acquire' -count=1` | ❌ Wave 0 | ⬜ pending |
| 23-01-03 | 01 | 1 | RES-08 | ASVS L1 | Discarding an owning acquisition is refused before C serialization or is immediately consumed exactly once. | admission + emitter | `go test ./internal/compiler/check ./internal/compiler/cgen -run '^TestPhase23Discard' -count=1` | ❌ Wave 0 | ⬜ pending |
| 23-02-01 | 02 | 2 | RES-09 | ASVS L1 | Acquisition-derived validation rejects all-release-deleted, duplicate, premature, wrong-resource, and fabricated cleanup controls, including reached paths after successful acquisition. | peer mutation | `go test ./internal/compiler/corevalidate ./internal/compiler/pathoracle -run '^TestPhase23ResourceMutation' -count=1` | ❌ Wave 0 | ⬜ pending |
| 23-02-02 | 02 | 2 | RES-04 | ASVS L1 | Public app reads the caller-selected byte through a live malloc-backed owner, borrowed use, and generated consuming release; two fixtures produce 65 and 66. | native app | `go test ./internal/compiler/session ./internal/compiler/native ./cmd/lang -run '^TestPhase23PublicFileByte' -count=1` | ❌ Wave 0 | ⬜ pending |
| 23-03-01 | 03 | 3 | RES-04, RES-07 | ASVS L1 | Actual allocation, later use, matching free, and failure cleanup are observed independently of compiler events; omitted, premature, duplicate, and wrong-resource destructor controls are reached and rejected. | native observer + public process | `sh scripts/verify-phase23.sh` | ❌ Wave 0 | ⬜ pending |
| 23-03-02 | 03 | 3 | FFI-03, RES-04, RES-07, RES-08, RES-09 | ASVS L1 | The focused source, model, native, observer, and negative-control evidence passes on macOS and Linux with host/build identity recorded. | cross-host CI | `sh scripts/verify-phase23.sh` | ❌ Wave 0 | ⬜ pending |
| 23-03-03 | 03 | 3 | RES-04 | ASVS L1 | Interpreter/model outcomes match independent answers and do not claim filesystem or physical cleanup proof. | model replay | `go test ./internal/compiler/interp ./internal/compiler/session -run '^TestPhase23Model' -count=1` | ❌ Wave 0 | ⬜ pending |

## Wave 0 Requirements

- [ ] Add source fixtures for path-token success, empty and oversized input, unreadable input, successful byte use, and the public `0x43` post-acquisition use error.
- [ ] Add operation-specific ABI/error fixtures and refusals for discarded ownership, moved-from use, duplicate release, escape, and unsupported exits.
- [ ] Add acquisition-seeded independent peer mutations, especially removal of every release from candidate core.
- [ ] Add an independent native observer and reached omitted, premature, duplicate, and wrong-resource destructor controls; establish malloc → use → matching free → exit on the positive path.
- [ ] Add the Phase 23 example, explicit binding manifest, and developer README instructions.
- [ ] Add one focused recurring command to the existing evidence-aggregate jobs on macOS and Linux without duplicating their full or sanitizer suites.
- [ ] Measure focused command runtime and set the maximum feedback latency from observed results.

## Manual-Only Verifications

All phase behaviors have automated verification. No conversational UAT or subjective human judgment is required. A host is complete only when its automated native receipt is available; an unavailable host remains incomplete.

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < measured Phase 23 focused-command latency target
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending automated evidence
