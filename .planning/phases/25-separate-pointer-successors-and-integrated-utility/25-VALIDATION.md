---
phase: "25"
slug: separate-pointer-successors-and-integrated-utility
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-10-01"
---

# Phase 25 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution. Commands and evidence lanes below are planned, not executed.

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing standard library; Go version constrained by the project to 1.24 |
| **Config file** | go.mod |
| **Quick run command** | go test -count=1 -run '^TestPhase25' ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/pathoracle ./internal/compiler/interp ./internal/compiler/session ./internal/compiler/cgen ./internal/compiler/native |
| **Full suite command** | go test ./... |
| **Estimated runtime** | Not measured during planning; report cold and warm distributions in the Phase 25 evidence script. |

## Sampling Rate

- After each task, run the focused command in this map.
- After each plan wave, run the quick Phase 25 command above.
- Before phase verification, run `go test ./...`, relevant race/vet/build checks, and `scripts/verify-phase25.sh` on both configured hosts through existing CI ownership.
- Keep focused feedback below 60 seconds where practical. Record cold and warm distributions for the integrated build/run/observe/repair loop.
- Evidence is a matrix by foreign/shared/exclusive family × macOS/Linux × applicable baseline, optimized, and sanitizer lane. A missing row remains incomplete; replay or cross-compilation does not substitute for a native host run.

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| T-25-01 | 01 | 1 | NAT-11, NAT-13 | T-25-01 | Shared source family is checked as read/copy-only and exact-shape | checker/source positive and conflict control | `go test -count=1 -run '^TestPhase25Shared(PointerSuccessor|Source)' ./internal/compiler/check` | Created in task | ⬜ pending |
| T-25-02 | 01 | 1 | NAT-11, NAT-13 | T-25-02, T-25-03 | Actual shared pointer C, call-site representation, and manifest agree without unsupported attributes | generated C/manifest contract | `go test -count=1 -run '^TestPhase25Shared(PointerABI|PointerManifest|NativeShape)' ./internal/compiler/cgen ./internal/compiler/native` | Created in task | ⬜ pending |
| T-25-03 | 02 | 2 | NAT-11, NAT-12, NAT-13 | T-25-04, T-25-06 | Core and origin peers independently reject wrong family, conflict, and escape | independent peer mutations | `go test -count=1 -run '^TestPhase25(PointerFamily|PointerOrigin)' ./internal/compiler/corevalidate ./internal/compiler/originvalidate` | Created in task | ⬜ pending |
| T-25-04 | 02 | 2 | NAT-11, NAT-12 | T-25-05, T-25-06 | Path oracle proves compatible, sequential, overlapping, and escaping loan paths | independent path-liveness controls | `go test -count=1 -run '^TestPhase25PointerPath' ./internal/compiler/pathoracle` | Created in task | ⬜ pending |
| T-25-05 | 03 | 3 | NAT-12, NAT-13 | T-25-07 | Exclusive helper and bounded owner-transfer composition check with independent U64-copy origin semantics and 0x43 ordering | source/checker and peer integration | `go test -count=1 -run '^TestPhase25(ExclusiveSource|TransferCallerComposition)' ./internal/compiler/check && go test -count=1 -run '^TestPhase25(U64CopyOrigin|PointerFamily|PointerOrigin)' ./internal/compiler/corevalidate ./internal/compiler/originvalidate` | Created in task | ⬜ pending |
| T-25-06 | 03 | 3 | NAT-12, NAT-13, DX-15 | T-25-08, T-25-09 | Each family has distinct conflict/escape evidence; peers and structured causes are exercised | source and independent peer negative controls | `go test -count=1 -run '^TestPhase25(FamilyConflict|FamilyEscape|IndependentPeer)' ./internal/compiler/session` | Created in task | ⬜ pending |
| T-25-07 | 04 | 4 | NAT-12, NAT-13 | T-25-10 | Exclusive C/manifest agree; local wrong-result is reached; unsupported shapes fail before serialization | emitter/native negative controls | `go test -count=1 -run '^TestPhase25Exclusive(PointerABI|PointerManifest|WrongResult|Refusal)' ./internal/compiler/cgen ./internal/compiler/native` | Created in task | ⬜ pending |
| T-25-08 | 04 | 4 | NAT-12, NAT-13, DX-15 | T-25-11, T-25-12 | Model answer and error ordering are stable; resource discharge, moved-from/discarded/escaped/cleanup boundaries have source-attributed controls | model, refusal, peer, and diagnostic controls | `go test -count=1 -run '^(TestPeerCalleeFrameDrained|TestPhase25(InterpreterComposition|ErrorBeforeHelpers|StructuredDiagnostic|UnsupportedPointerShape|OwnershipDiagnosticBoundary))' ./internal/compiler/corevalidate ./internal/compiler/interp ./internal/compiler/diagnostic ./internal/compiler/session` | Created in task | ⬜ pending |
| T-25-09 | 05 | 5 | EVD-10, DX-14, NAT-11, NAT-12 | T-25-13 | Clean-checkout utility yields 65/66 and typed 0x43 failure; each family has native and reached wrong-result controls | application/native integration | `go test -count=1 -run '^TestPhase25(Utility|NativeFamilyWrongResult)' ./internal/compiler/native` | Created in task | ⬜ pending |
| T-25-10 | 05 | 5 | EVD-10, DX-14 | T-25-14, T-25-15 | Every family × host × applicable lane is indexed and fail-closed; living claims separate source, current CI, and history | evidence script and living-document contract | `go test -count=1 -run '^TestPhase25(EvidenceScript|EvidenceIndex|LivingRoadmap)' ./internal/compiler/native && sh scripts/verify-phase25.sh` | Created in task | ⬜ pending |

## Wave 0 Requirements

No separate Wave 0 setup is planned. Each task creates its focused tests and source fixtures alongside implementation. Tests and native evidence have not been executed as part of planning.

## Manual-Only Verifications

All behavioral criteria are intended to have automated checks. macOS and Linux receipts must originate on their native host lanes; a local run on one host cannot stand in for the other. Human review should confirm the clean-checkout README sequence is legible and the evidence index directly navigates to every family/host/lane record.

## Validation Sign-Off

- [ ] All tasks have automated verification or a Wave 0 dependency.
- [ ] Sampling continuity: no three consecutive tasks lack automated verification.
- [ ] Wave 0 covers all missing references.
- [ ] No watch-mode flags are used.
- [ ] Focused feedback latency is below 60 seconds where practical and cold/warm distributions are recorded.
- [ ] Every required host/family/lane result is present with no silent skips.
- [ ] nyquist_compliant: true set in frontmatter after execution validation.

**Approval:** pending
