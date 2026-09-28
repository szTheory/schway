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
| 23-01-01 | 01 | 1 | FFI-03, RES-04 | ASVS L1 | The first public byte path crosses checked acquire/use/release operations and independent peers before native output. | public native tracer | `go test ./cmd/lang ./internal/compiler/native -run '^TestPhase23PublicFileByte' -count=1` | ❌ Wave 0 | ⬜ pending |
| 23-01-02 | 01 | 1 | FFI-03, RES-04 | ASVS L1 | Two independent files yield 65/66, and all three operation prototypes have target-specific conformance. | public native + ABI | `go test ./cmd/lang ./internal/compiler/native -run '^TestPhase23(PublicFileByte|OperationABI)' -count=1` | ❌ Wave 0 | ⬜ pending |
| 23-02-01 | 02 | 2 | FFI-03, RES-09 | ASVS L1 | Each peer derives acquire-seeded owner obligations and use/cleanup facts independently. | peer validation | `go test ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/pathoracle -run '^TestPhase23' -count=1` | ❌ Wave 0 | ⬜ pending |
| 23-02-02 | 02 | 2 | RES-08, RES-09 | ASVS L1 | All-release-deleted, duplicate, premature, wrong-resource, and fabricated controls are reached and refused; discard is refused before C emission. | peer + source mutation | `go test ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/pathoracle ./internal/compiler/check ./internal/compiler/cgen -run '^TestPhase23(ResourceMutation|SourceRefusal|Discard)' -count=1` | ❌ Wave 0 | ⬜ pending |
| 23-02-03 | 02 | 2 | RES-07 | ASVS L1 | Empty, oversized, non-regular, malformed, allocation/read/close failure, and partial cleanup are typed, bounded, and owner-free. | public acquisition | `go test ./internal/compiler/native ./cmd/lang -run '^TestPhase23(Acquire|PublicFileByte)' -count=1` | ❌ Wave 0 | ⬜ pending |
| 23-03-01 | 03 | 2 | RES-04, RES-07 | ASVS L1 | Interpreter outcomes agree with independent answers and are explicitly model only. | model replay | `go test ./internal/compiler/interp ./internal/compiler/session -run '^TestPhase23Model' -count=1` | ❌ Wave 0 | ⬜ pending |
| 23-03-02 | 03 | 2 | FFI-03 | ASVS L1 | Successor contract states release consumes, borrow preserves, and later transfer would preserve under a new owner. | contract schema | `go test ./internal/compiler/session -run '^TestPhase23Contract' -count=1` | ❌ Wave 0 | ⬜ pending |
| 23-04-01 | 04 | 3 | RES-04, RES-07 | ASVS L1 | Public success/use error show actual allocation, later use, matching free, and no outstanding pointer before exit. | native observer | `go test ./internal/compiler/native ./cmd/lang -run '^TestPhase23(Observer|PublicFileByte|PublicUseError)' -count=1` | ❌ Wave 0 | ⬜ pending |
| 23-04-02 | 04 | 3 | RES-09 | ASVS L1 | Four reached physical destructor controls are rejected despite plausible compiler events. | native mutations | `go test ./internal/compiler/native ./internal/compiler/cgen -run '^TestPhase23(ObserverMutation|PhysicalDestructorControl)' -count=1` | ❌ Wave 0 | ⬜ pending |
| 23-05-01 | 05 | 4 | RES-04 | ASVS L1 | Published clean-checkout commands and independent answers are machine checked. | documentation contract | `go test ./cmd/lang ./internal/compiler/session -run '^TestPhase23(Readme|Public|Contract)' -count=1` | ❌ Wave 0 | ⬜ pending |
| 23-05-02 | 05 | 4 | FFI-03, RES-04, RES-07, RES-08, RES-09 | ASVS L1 | Focused evidence runs on macOS/Linux with distinct host receipts and without duplicate full suites. | cross-host CI | `sh scripts/verify-phase23.sh` | ❌ Wave 0 | ⬜ pending |

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
