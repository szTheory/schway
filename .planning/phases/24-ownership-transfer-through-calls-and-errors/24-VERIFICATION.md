---
phase: 24-ownership-transfer-through-calls-and-errors
verified: 2026-10-01T12:28:56Z
status: passed
score: 5/5 roadmap success criteria verified
covered_files:
  - .github/workflows/ci.yml
  - .planning/REQUIREMENTS.md
  - .planning/ROADMAP.md
  - .planning/phases/24-ownership-transfer-through-calls-and-errors/24-01-PLAN.md
  - .planning/phases/24-ownership-transfer-through-calls-and-errors/24-01-SUMMARY.md
  - .planning/phases/24-ownership-transfer-through-calls-and-errors/24-02-PLAN.md
  - .planning/phases/24-ownership-transfer-through-calls-and-errors/24-02-SUMMARY.md
  - .planning/phases/24-ownership-transfer-through-calls-and-errors/24-03-PLAN.md
  - .planning/phases/24-ownership-transfer-through-calls-and-errors/24-03-SUMMARY.md
  - .planning/phases/24-ownership-transfer-through-calls-and-errors/24-VALIDATION.md
  - examples/phase24/README.md
  - examples/phase24/error.schway
  - examples/phase24/transfer.schway
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/check/check.go
  - internal/compiler/check/check_test.go
  - internal/compiler/core/core.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/interp/interp.go
  - internal/compiler/native/native_app_test.go
  - internal/compiler/native/phase24_observer_test.go
  - internal/compiler/native/testdata/phase24_observer.c
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/session/session_phase24_error_test.go
  - internal/compiler/session/session_phase24_model_test.go
  - internal/compiler/session/session_phase24_transfer_test.go
  - scripts/verify-phase24.sh
covered_digest: "v1:sha256:c2c6b3fc985d2f0875ec9ab98c19b8a8de2233f0323d783cb4cf61008a57101d"
behavior_unverified: 0
overrides_applied: 0
---

# Phase 24: Ownership Transfer Through Calls and Errors — Verification Report

**Phase Goal:** A developer can transfer a live resource through Schway calls and returns, use it under its new owner, and rely on exactly-once cleanup on admitted normal and typed-error paths.
**Verified:** 2026-10-01T12:28:56Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | An allocation acquired in one frame remains usable after ownership transfer through a Schway call and return; transfer invokes no destructor, borrow preserves ownership, and the final owner invokes the declared consuming destructor exactly once. | ✓ VERIFIED | Source inspection: `checkLocalFileByteAcquireHelper` and `checkLocalFileByteTransferCaller` lower owning return, borrow, then paired release; interpreter and `emitProgram` carry the owner across frames. `TestPhase24TransferPeerValidatesAcquisitionDerivedDischarge`, `TestPhase24EmitterTransfer`, `TestPhase24PositiveTransferNativeApplication`, and `TestPhase24ObserverPublicLifecycleAndTypedError` are in the hosted full suite and focused aggregate. The independent observer matches actual allocated pointers to use and free. |
| 2 | Two activations of the same static acquisition site produce distinct semantic resource identities that survive transfer and are independently checked without using raw host addresses as portable IDs. | ✓ VERIFIED | Source inspection: interpreter activation keys combine the static acquire operation with dynamic call activation; core, origin, and path peers derive ownership from acquisition/callee facts. `TestPhase24ActivationPeerDerivesRepeatedAcquisitions` and `TestPhase24ObserverPublicLifecycleAndTypedError` cover repeated activations; the observer stores pointer values privately and reports operation/activation IDs. Hosted focused aggregate passed on both hosts. |
| 3 | Actual entry-to-success and entry-to-typed-error executions release every non-transferred caller/callee resource in reverse successful-acquisition completion order, including multiple acquisitions followed by a real later operation/output failure; failed acquisitions contribute no obligation. | ✓ VERIFIED | Source inspection: `examples/phase24/error.schway`, checker success/error edges, independent peer cleanup derivations, interpreter, and generated C encode the admitted path. `TestPhase24CleanupPeerRejectsOwnershipAndOrderMutations`, `TestPhase24EmitterErrorPreservesTypedFailureAndCleanupOrder`, `TestPhase24NativeErrorApplication`, and the physical observer assert the `0x43` typed error after three acquisitions and C,B,A release order. Observer verifies zero outstanding allocations before error exit. Hosted CI full and focused checks passed. |
| 4 | Copying ownership, using a moved-from owner, or returning an owning process-entry result without an external receiver is rejected before execution with source-attributed diagnostics. | ✓ VERIFIED | Source inspection: `TestPhase24SourceRefusal` mutates the canonical witness for owner-copy, moved-from-use, and owning-entry-result cases and asserts refusal/source spans. The focused source-refusal group passed in hosted CI on Linux and macOS. |
| 5 | An observer independent of compiler release events proves allocation, use after acquisition/return, and destruction before exit. Reached omitted, premature, duplicate, and wrong-resource destruction controls all fail even with plausible reported events. | ✓ VERIFIED | `phase24_observer.c` observes actual pointer allocation/use/free and semantic identity; `TestPhase24ObserverReachedPhysicalDestructorControls` reaches omitted, premature, duplicate, wrong-resource, and identity-collision controls while checking the reporting boundary. The hosted Phase 24 aggregate passed on Linux/x86_64 and Darwin/arm64. |

**Score:** 5/5 roadmap truths verified (0 behavior-unverified).

### Plan Cross-Check

All three plans and summaries were read. Their required source witnesses, independent validation peers, interpreter/emitter paths, physical observer, model-only replay, and dual-host validation are represented in the five roadmap truths above. No plan-specific must-have reduces the roadmap contract.

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `examples/phase24/transfer.schway`, `error.schway` | Ordinary success and post-acquisition typed-error witnesses | ✓ VERIFIED | Non-stub fixtures feed source admission, interpreter/emitter, and native-app tests. |
| `internal/compiler/check/check.go` and `core/core.go` | Bounded ownership transfer and typed-error facts | ✓ VERIFIED | Source lowering carries per-operation foreign contracts and explicit success/error edges; source refusals are tested. |
| `internal/compiler/corevalidate/corevalidate.go`, `originvalidate/originvalidate.go`, `pathoracle/pathoracle.go` | Independent acquisition-derived validation | ✓ VERIFIED | Separate peer entry points and mutation tests reject deleted, duplicated, mismatched, reordered, colliding, and frame-confused obligations. |
| `internal/compiler/interp/interp.go`, `internal/compiler/cgen/cgen_program.go` | Model execution and sole production C serializer | ✓ VERIFIED | Model tracks activation identity; C emitter preflights and emits admitted transfer/error paths. Tests keep model claims separate from physical cleanup. |
| `internal/compiler/native/phase24_observer_test.go`, `testdata/phase24_observer.c` | Independent physical lifecycle proof and reached controls | ✓ VERIFIED | Observer uses actual allocation pointers privately and checks use, free ordering, reporting boundary, and no outstanding storage. |
| `scripts/verify-phase24.sh`, `.github/workflows/ci.yml` | Focused Phase 24 evidence in existing two-host aggregate | ✓ VERIFIED | The existing evidence-aggregate matrix invokes the named script on Ubuntu and macOS. |

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| Phase 24 Schway fixtures | `check` | source parsing/admission to ownership operations | WIRED | Named source tests inspect transfer, error edges, refusal diagnostics, and paired release facts. |
| Checked core/acquisition operation | `corevalidate`, `originvalidate`, `pathoracle` | independent acquisition-derived replay | WIRED | Each peer is invoked in session tests on valid core and mutations; no candidate release is needed to seed the obligation. |
| Checked core and foreign contracts | interpreter / `emitProgram` | call-return and typed-error lowering | WIRED | Model and native application tests exercise the ordinary fixtures; unsupported shapes are refused before serialization. |
| Generated app and adapter | Phase 24 observer | pointer-private instrumentation at actual allocation/use/free | WIRED | Instrumented app records physical lifecycle independent of compiler release events. |
| `scripts/verify-phase24.sh` | CI `evidence-aggregate` matrix | Ubuntu/macOS workflow step | WIRED | Workflow has one named invocation per matrix host; final hosted run reports both successful. |

### Data-Flow Trace (Level 4)

| Artifact | Data variable | Source | Produces real data | Status |
|---|---|---|---|---|
| Success app | byte and U64 result | caller-created one-byte file → Phase 23 adapter allocation/use → helper return → caller borrow | Yes; hosted native app cases independently expect 65 and 66 | ✓ FLOWING |
| Typed-error app | `UseError.UnsupportedByte` and owner obligations | caller path → three successful helper acquisitions → real `0x43` use operation → typed-error edge | Yes; hosted native app and physical observer cases execute this input | ✓ FLOWING |
| Physical cleanup receipt | operation/activation identity and lifecycle | actual malloc/use/free calls; host pointer remains private to observer | Yes; `0x43` path checks C,B,A and zero outstanding before report/exit | ✓ FLOWING |

### Behavioral Spot-Checks

No local project commands were run, as required by the phase validation contract. The following named tests were exercised by hosted full suites and the focused aggregate; this is historical hosted execution evidence, distinct from this report's source inspection.

| Behavior | Command/evidence | Result | Status |
|---|---|---|---|
| Phase 24 focused aggregate | Hosted run [36856048690](https://github.com/szTheory/schway/actions/runs/36856048690), `scripts/verify-phase24.sh`, Linux/x86_64 | 25 seconds; passed | ✓ PASS |
| Phase 24 focused aggregate | Same run, `scripts/verify-phase24.sh`, Darwin/arm64 | 22 seconds; passed | ✓ PASS |
| Full vet/build/test/race checks | Same hosted run, Linux/x86_64 | All passed, 5m44s | ✓ PASS |
| Full vet/build/test/race checks | Same hosted run, Darwin/arm64 | All passed, 11m30s | ✓ PASS |

The receipt binds to source revision `ed94ef7972b25deb85ab90fafbf3403dc31629f5`, with clean trees on both hosts. Current HEAD `ee1190031484dfb7e468db0b597d954802c06b8b` differs from that tested revision only in planning files: `24-02-SUMMARY.md`, `24-03-SUMMARY.md`, `24-REVIEW.md`, and `24-VALIDATION.md`. The implementation, test, probe, and CI files covered here are unchanged, so the hosted behavioral evidence applies to the current implementation. The optional validation-corpus job was skipped by manual dispatch and is outside the phase acceptance contract.

### Probe Execution

| Probe | Command/evidence | Result | Status |
|---|---|---|---|
| `scripts/verify-phase24.sh` | Hosted run `36856048690`, Ubuntu and macOS `evidence-aggregate` steps | Both passed; focused groups include source refusal, independent peers, model replay, emitter admission, native apps, observer controls, and public contract | PASS |

### Requirements Coverage

| Requirement | Source plans | Description | Status | Evidence |
|---|---|---|---|---|
| RES-05 | 24-01, 24-02, 24-03 | Transfer keeps the resource live; borrow preserves obligation; exact declared destructor consumes it once | ✓ SATISFIED | Source/core/emitter inspection plus transfer peer, native app, and independent physical observer hosted checks. |
| RES-06 | 24-02, 24-03 | Normal and typed-error cleanup follows reverse successful-acquisition completion order; failed acquire adds no obligation | ✓ SATISFIED | Checker/peer/emitter evidence and actual C,B,A observer path after real `0x43` failure. |
| OWN-10 | 24-01, 24-02 | Transfer through call/return, new-owner use/release, misuse and entry escape refusals | ✓ SATISFIED | Source refusal tests, three independent peers, ordinary native transfer application. |
| OWN-11 | 24-02, 24-03 | Activation-qualified identity survives transfer and is independent of host addresses | ✓ SATISFIED | Interpreter/peer source inspection, repeated activation tests, physical observer and identity-collision control. |
| OWN-12 | 24-02, 24-03 | Typed-error propagation discharges caller/callee obligations in declared order | ✓ SATISFIED | Ordinary native `0x43` application, cleanup mutation peers, C,B,A physical observer. |
| EVD-09 | 24-03 | Independent observer proves actual lifecycle and reached hostile controls | ✓ SATISFIED | Hosted physical observer test on Linux and macOS; all five controls reach and are rejected. |

All six requirements mapped to Phase 24 are declared in the plans; no orphaned Phase 24 requirement was found.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---|---|---|---|
| — | — | No blocking TODO/FIXME/XXX debt marker or user-visible stub found in the inspected phase implementation paths. Grep matches were test fixtures, empty helper return values, or the observer's deliberate `NULL` failure handling. | — | — |

### Human Verification Required

None. The contract is bounded and objectively exercised by hosted native app cases and an independent physical observer; no visual, subjective UX, or external service claim remains.

### Decision Coverage

The decision-coverage query reports all 9 trackable CONTEXT decisions honored; it is non-blocking.

### Gaps Summary

No gaps found. All five roadmap truths have implementation witnesses and hosted behavioral evidence. Source inspection, hosted execution receipts, and earlier Phase 23 receipts are kept distinct; Phase 23 history is not used as evidence for the Phase 24 claim.

---

_Verified: 2026-10-01T12:28:56Z_
_Verifier: the agent (gsd-verifier)_
