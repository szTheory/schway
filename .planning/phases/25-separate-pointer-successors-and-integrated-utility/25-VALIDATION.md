---
phase: "25"
slug: separate-pointer-successors-and-integrated-utility
status: incomplete
nyquist_compliant: true
wave_0_complete: true
created: "2026-10-01"
---

# Phase 25 — Validation Strategy

> Execution validation is complete on the local macOS host. Hosted macOS/Linux family-by-lane receipts remain required before EVD-10 can close.

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
- Before phase verification, run `go test -count=1 ./...`, relevant race/vet/build checks, and `scripts/verify-phase25.sh` on both configured hosts through existing CI ownership.
- Keep focused feedback below 60 seconds where practical. Record cold and warm distributions for the integrated build/run/observe/repair loop.
- Evidence is a matrix by foreign/shared/exclusive family × macOS/Linux × applicable baseline, optimized, and sanitizer lane. A missing row remains incomplete; replay or cross-compilation does not substitute for a native host run.

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| T-25-01 | 01 | 1 | NAT-11, NAT-13 | T-25-01 | Shared source family is checked as read/copy-only and exact-shape | checker/source positive and conflict control | `go test -count=1 -run '^TestPhase25Shared(PointerSuccessor|Source)' ./internal/compiler/check` | Created in task | ✅ local pass |
| T-25-02 | 01 | 1 | NAT-11, NAT-13 | T-25-02, T-25-03 | Actual shared pointer C, call-site representation, and manifest agree without unsupported attributes | generated C/manifest contract | `go test -count=1 -run '^TestPhase25Shared(PointerABI|PointerManifest|NativeShape)' ./internal/compiler/cgen ./internal/compiler/native` | Created in task | ✅ local pass |
| T-25-03 | 02 | 2 | NAT-11, NAT-12, NAT-13 | T-25-04, T-25-06 | Core and origin peers independently reject wrong family, conflict, and escape | independent peer mutations | `go test -count=1 -run '^TestPhase25(PointerFamily|PointerOrigin)' ./internal/compiler/corevalidate ./internal/compiler/originvalidate` | Created in task | ✅ local pass |
| T-25-04 | 02 | 2 | NAT-11, NAT-12 | T-25-05, T-25-06 | CFG endpoints and straight-line loan replay separately cover compatible, sequential, overlapping, and borrowed-result escape paths; empty/forged endpoint claims cannot steer straight-line replay | independent path-liveness controls | `TestPhase25PointerPathLiveOverlap`, `TestPhase25PointerPathBorrowedResultEscape`, `TestPhase25IndependentPeerMutations` | Created in task | ✅ local pass at 2026-10-02 follow-up |
| T-25-05 | 03 | 3 | NAT-12, NAT-13 | T-25-07 | Exclusive helper and bounded owner-transfer composition check with independent U64-copy origin semantics and 0x43 ordering | source/checker and peer integration | `go test -count=1 -run '^TestPhase25(ExclusiveSource|TransferCallerComposition)' ./internal/compiler/check && go test -count=1 -run '^TestPhase25(U64CopyOrigin|PointerFamily|PointerOrigin)' ./internal/compiler/corevalidate ./internal/compiler/originvalidate` | Created in task | ✅ local pass |
| T-25-06 | 03 | 3 | NAT-12, NAT-13, DX-15 | T-25-08, T-25-09 | Each family has distinct conflict/escape evidence; all three peer verdicts, command admission, source/cause spans, schema, function identity, and absent unsafe repairs are exercised | source and independent peer negative controls | `TestPhase25IndependentPeerMutations`, `TestPhase25CheckCommandPeerGate`, `TestPhase25CheckCommandPeerObservation`, `TestPhase25EscapeDiagnosticSourceAttribution`, `TestPhase25EscapeDiagnosticFunctionIdentity`, `TestPhase25FamilyConflict` | Created in task | ✅ local pass at 2026-10-02 follow-up |
| T-25-07 | 04 | 4 | NAT-12, NAT-13 | T-25-10 | Exclusive C/manifest agree; local wrong-result is reached; unsupported shapes fail before serialization | emitter/native negative controls | `go test -count=1 -run '^TestPhase25Exclusive(PointerABI|PointerManifest|WrongResult|Refusal|PointerRefusal)$' ./internal/compiler/cgen ./internal/compiler/native` | Created in task | ✅ local pass |
| T-25-08 | 04 | 4 | NAT-12, NAT-13, DX-15 | T-25-11, T-25-12 | Model answer and error ordering are stable; exact owner-result flow and resource discharge are independently checked; source-expressible ownership boundaries and serializer-only pointer mutations have controls | model, source refusal, peer, serializer, and diagnostic controls | `go test -count=1 -run '^(TestPeerCalleeFrameDrained|TestPhase25(OwnerTransfer|InterpreterComposition|ErrorBeforeHelpers|StructuredDiagnostic|UnsupportedPointerShape|OwnershipDiagnosticBoundary))' ./internal/compiler/corevalidate ./internal/compiler/interp ./internal/compiler/diagnostic ./internal/compiler/session && go test -count=1 -run '^TestPhase25Exclusive(PointerABI|PointerManifest|WrongResult|Refusal)' ./internal/compiler/cgen ./internal/compiler/native` | Created in task | ✅ local pass |
| T-25-09 | 05 | 5 | NAT-11, NAT-12, NAT-13 | T-25-13 | Exact shared U64-copy ABI and Phase 25 caller emit natively; path and origin peers independently accept only the same exact chain; utility returns 65/66 and typed 0x43 failure with native reached wrong-result controls | cgen/application/independent peer/native integration | `go test -count=1 -run '^TestPhase25SharedPointerCopyABI' ./internal/compiler/cgen && go test -count=1 -run '^TestPhase25UtilityOwnerTransfer' ./internal/compiler/pathoracle ./internal/compiler/originvalidate && go test -count=1 -run '^TestPhase25(Utility|NativeFamilyWrongResult)' ./internal/compiler/native` | Created in task | ✅ local pass |
| T-25-10 | 06 | 6 | EVD-10, DX-14, NAT-11, NAT-12, NAT-13 | T-25-14, T-25-15 | Every family × host × applicable lane is indexed and fail-closed; README reproduces utility results; existing family fixtures use bounded subprocesses; living claims separate source, local/hosted checks, and history | evidence script, README, bounded native controls, and living-document contract | `go test -count=1 -run '^(TestSourceNeverSpawnsUnboundedProcesses|TestPhase25(EvidenceScript|EvidenceIndex|LivingRoadmap))$' ./internal/compiler/native && sh scripts/verify-phase25.sh` | Created in task | ✅ local pass; Linux/hosted matrix pending |

## Security audit follow-up — 2026-10-02

Source inspection at code revision `b19446f` confirms that the command gate
runs core validation, origin validation, then the independent path oracle;
the path oracle derives straight-line loan ancestry, last use, access-family
overlap, and terminal borrowed-result escape from operation/place/type facts,
without consulting `LoanEndpoints`. `RecomputeEndpoints` remains limited to
CFG endpoint synthesis. Origin refusals carry structured function and return
operation identity; the command projects that identity through checked core
facts and lossless source tokens to the terminal result and creating borrow.

Fresh local checks passed with `GOCACHE=/private/tmp/schway-gocache`:

- `TestPhase25PointerPathLiveOverlap`,
  `TestPhase25PointerPathBorrowedResultEscape`,
  `TestPhase25IndependentPeerMutations`, and
  `TestPhase25CheckCommandPeerGate` and `TestPhase25CheckCommandPeerObservation`.
- `TestPhase25EscapeDiagnosticSourceAttribution` and
  `TestPhase25EscapeDiagnosticFunctionIdentity`, plus the retained conflict,
  schema, unsupported-shape, and ownership-boundary controls.
- The final combined verification reran those controls and
  `TestLanguageMaturityCountsAreCurrent`.

The direct overlap and borrowed-result escape controls reject empty and
forged endpoint declarations. Core validation reports the overlap refusal and
marks the escaped function non-callable in its independently derived
signature; origin validation reports `core.origin_omitted`; pathoracle reports
its own overlap/escape refusal. The actual `CheckCommandFile` cases refuse
both negative source witnesses. Escape diagnostics select exact nonempty
primary and borrow-origin cause spans, retain `lang.diagnostic/0` and
`core.origin_omitted`, and carry no repair. Existing Phase 25 native receipts
remain historical at their recorded revisions. No hosted Phase 25 matrix was
run by this follow-up; EVD-10 remains open.

## Wave 0 Requirements

No separate Wave 0 setup was needed. Each task created its focused tests and source fixtures alongside implementation. Local focused tests, the repository-wide suite, and the native evidence script are recorded during execution; hosted dual-host receipts remain open.

## Manual-Only Verifications

All behavioral criteria are intended to have automated checks. macOS and Linux receipts must originate on their native host lanes; a local run on one host cannot stand in for the other. Human review should confirm the clean-checkout README sequence is legible and the evidence index directly navigates to every family/host/lane record.

## Validation Sign-Off

- [x] All tasks have automated verification or a Wave 0 dependency.
- [x] Sampling continuity: no three consecutive tasks lack automated verification.
- [x] Wave 0 covers all missing references.
- [x] No watch-mode flags are used.
- [x] Focused feedback latency is below 60 seconds where practical and cold/warm distributions are recorded.
- [ ] Every required host/family/lane result is present with no silent skips.
- [x] nyquist_compliant: true set in frontmatter after execution validation.

**Approval:** local validation complete; hosted dual-host evidence remains pending.
