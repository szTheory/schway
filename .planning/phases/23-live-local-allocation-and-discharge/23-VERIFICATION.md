---
phase: 23-live-local-allocation-and-discharge
verified: 2026-09-28T16:00:00Z
status: gaps_found
score: 5/5 roadmap truths verified
covered_files:
  - .github/workflows/ci.yml
  - .planning/REQUIREMENTS.md
  - .planning/ROADMAP.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-01-PLAN.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-01-SUMMARY.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-02-PLAN.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-02-SUMMARY.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-03-PLAN.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-03-SUMMARY.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-04-PLAN.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-04-SUMMARY.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-05-PLAN.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-05-SUMMARY.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-06-PLAN.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-06-SUMMARY.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-07-PLAN.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-07-SUMMARY.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-REVIEW.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-SECURITY.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-VALIDATION.md
  - cmd/lang/main.go
  - cmd/lang/main_test.go
  - examples/phase23/README.md
  - examples/phase23/adapter.c
  - examples/phase23/adapter.h
  - examples/phase23/file_byte.bindings.json
  - examples/phase23/file_byte.lang
  - internal/compiler/ability/ability.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/check/check.go
  - internal/compiler/core/core.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/interp/interp.go
  - internal/compiler/native/bindings.go
  - internal/compiler/native/native_app.go
  - internal/compiler/native/native_app_test.go
  - internal/compiler/native/phase23_observer_test.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_app_verify.go
  - scripts/verify-phase23.sh
  - testdata/phase23/discard_owner.lang
covered_digest: "v1:sha256:013c3469bcf07198d2ff9c6c82f8cf6b30bc9edb7324f37d97ce1ccaab7fdb11"
behavior_unverified: 0
overrides_applied: 0
gaps:
  - truth: "The configured hosted Ubuntu evidence-aggregate lane has a passing Phase 23 receipt."
    status: partial
    reason: "Only local host receipts exist. The focused gate is wired into the Ubuntu/macOS matrix, but no hosted workflow receipt is available; the validation contract explicitly requires it before phase completion."
    artifacts:
      - path: ".github/workflows/ci.yml"
        issue: "The Ubuntu matrix step is wired, but its hosted execution result cannot be established from this checkout."
      - path: ".planning/phases/23-live-local-allocation-and-discharge/23-VALIDATION.md"
        issue: "Hosted Ubuntu receipt is explicitly pending; validation remains in-progress and nyquist_compliant is false."
    missing:
      - "A passing hosted Ubuntu evidence-aggregate receipt for scripts/verify-phase23.sh."
---

<!-- schway-current:start -->
Current publication identity (2026-09-28): Schway uses the Go module `github.com/szTheory/schway`, the `schway` and `schway-repair` commands, `.schway` source files, `schway.*` and `schway:*` protocol identifiers, and `schway_` and `SCHWAY_` native ABI symbols. Preserve older spellings only where they document historical implementation evidence.
<!-- schway-current:end -->


# Phase 23: Live Local Allocation and Discharge — Verification Report

**Phase Goal:** A developer can read a caller-selected file byte through a real allocation returned live to Lang and observe its generated local cleanup.
**Verified:** 2026-09-28T16:00:00Z
**Status:** gaps_found
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | An opaque noncopyable resource receives bounded malloc storage, remains live through Lang-directed use, yields the independently expected byte, and is physically destroyed before exit. | ✓ VERIFIED | `examples/phase23/file_byte.lang:8-43` declares the distinct acquire/borrow/consume sequence; `adapter.c:80-147` allocates, reads and frees; `cmd/lang/main_test.go#TestPhase23PublicFileByte` checks caller files 0x41/0x42; `internal/compiler/native/phase23_observer_test.go#TestPhase23ObserverPublicLifecycleAndUseFailure` observes pointer lifecycle. Focused gate passed. |
| 2 | Acquire, use, and consuming release each carry independent checked signatures, operand/failure facts, and release pairing. | ✓ VERIFIED | `file_byte.lang:7-32` declares separate operation contracts; `adapter.h:42-48` defines separate C function types; `native/bindings.go` compiles per-symbol binding checks. `TestPhase23OperationABI` and focused peer/emitter mutation groups passed. |
| 3 | Empty, maximum-size, oversized, malformed, and failed inputs follow the published bounds and failure contract; failed acquisition creates no owner and partial storage is cleaned. | ✓ VERIFIED | `adapter.c:62-128` initializes failed owner records empty and frees partial allocations; `cmd/lang/main_test.go#TestPhase23PublicFileByte` covers path bounds, empty/two-byte, and malformed/NUL cases; `internal/compiler/native/native_app_test.go#TestPhase23AcquireFailuresInitializeAndFreePartialAllocations` covers injected failures. |
| 4 | Discarded owning success is refused or immediately destroyed; independent acquisition-seeded validators reject missing, duplicate, wrong-resource, and fabricated cleanup. | ✓ VERIFIED | `internal/compiler/check/check.go`, `corevalidate/corevalidate.go`, `originvalidate/originvalidate.go`, and `pathoracle/pathoracle.go` provide separate admission/replay checks. Source/emitter refusal tests and all-release-deleted/duplicate/wrong-resource/fabricated peer mutation controls passed in the focused gate. |
| 5 | Normal completion and a real post-acquisition use/output failure clean each remaining owner once; separate physical observation rejects omitted or premature destruction. | ✓ VERIFIED | `TestPhase23ObserverPublicLifecycleAndUseFailure` exercises normal and 0x43 use-error paths; `TestPhase23ObserverMutationControlsPreservePlausibleEvents` reaches omitted, premature, duplicate, and wrong-resource physical controls. All passed. |

**Score:** 5/5 roadmap truths verified (0 present, behavior-unverified).

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `examples/phase23/file_byte.lang` | Public checked owner flow | ✓ VERIFIED | Source declares explicit acquire, borrowed use, and consuming release, then returns the value. |
| `examples/phase23/adapter.c`, `adapter.h`, `file_byte.bindings.json` | Local allocation adapter and pinned ABI | ✓ VERIFIED | Real file descriptor/read/allocation/use/free path; C17 record layout and operation type checks. |
| `scripts/verify-phase23.sh` | Focused reproducible evidence gate | ✓ VERIFIED | Direct run passed all groups on Darwin/arm64 in 7 seconds with writable Go cache. Default cache invocation failed due sandbox permissions; rerun succeeded using the documented cache path. |
| `.github/workflows/ci.yml` | Existing host matrix wiring | ✓ VERIFIED (wiring) | Focused script runs in `evidence-aggregate` for `ubuntu-latest` and `macos-latest`; hosted Ubuntu result remains the sole gap. |
| `examples/phase23/README.md`, `file_byte.expected.json` | Clean-checkout commands and independent answers | ✓ VERIFIED | Contract test passed; constants cover 0x41/0x42 success and acquisition/use errors, with evidence scope limits. |

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| `cmd/lang/main.go` | `native.RunApplication` | CLI dispatch | WIRED | Public test executes the app with caller-selected file paths. |
| Generated entry/source | checked acquire/use/release | `file_byte.lang` plus explicit manifest | WIRED | Focused public and native observer tests prove outcome and generated cleanup. |
| Each foreign operation | C declaration | per-symbol ABI probes | WIRED | Independent mismatch controls pass. |
| `evidence-aggregate` matrix | `scripts/verify-phase23.sh` | workflow step | WIRED | Source inspection confirms Ubuntu/macOS matrix configuration; no hosted execution receipt found. |
| README | independent fixture and app outcomes | contract test | WIRED | `TestPhase23ReadmeContract` and expected-constant test passed. |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|---|---|---|---|---|
| `file_byte.lang` main result | `value` | `lang_file_byte_use(owner)` reading `owner.data` from `malloc` storage populated from caller file | Yes | ✓ FLOWING |
| Generated owner cleanup | `owner.data` | Matching `lang_file_byte_release` → `free` | Yes; independently observed by native observer | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|---|---|---|---|
| Entire focused Phase 23 evidence set | `GOCACHE=/tmp/ai-lang-verification-gocache sh scripts/verify-phase23.sh` | Exit 0; all five groups passed; Darwin/arm64, Go 1.24.0, Apple Clang 21.0.0, target `arm64-apple-darwin25.6.0`, revision `c50430d9fa1490b393c5d22805f094787ee99186`, 7 seconds. The script reports `tree=modified` due existing planning edits. | ✓ PASS |
| Local Linux ARM64 focused receipt | Historical receipt recorded in `23-VALIDATION.md` | Docker Linux ARM64, Go 1.24.13, Debian Clang 14.0.6, target `aarch64-unknown-linux-gnu`, revision c50430d, 14 seconds, passed. This is local Docker evidence, not hosted CI. | ✓ PASS (local only) |
| Full Go suite | Historical receipt recorded in `23-VALIDATION.md` | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./...` passed on macOS after the post-fix change. Not rerun in this verification. | ✓ PASS (historical receipt) |

### Probe Execution

| Probe | Command | Result | Status |
|---|---|---|---|
| `scripts/verify-phase23.sh` | `GOCACHE=/tmp/ai-lang-verification-gocache sh scripts/verify-phase23.sh` | Exit 0; all groups passed. | PASS |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|---|---|---|---|---|
| FFI-03 | 01, 02, 03, 05, 07 | Operation-specific checked foreign ABI contracts | ✓ SATISFIED | Separate source contracts, per-symbol compiled ABI probes, peer and emitter mismatch controls. |
| RES-04 | 01, 04, 05, 06, 07 | Live bounded allocation and real byte use | ✓ SATISFIED | Public 0x41/0x42 files, observer-confirmed allocation/use/release, independent expected outputs. |
| RES-07 | 04, 06, 07 | Failed acquisition and partial cleanup | ✓ SATISFIED | Typed failure controls, initialized empty owner, injected partial allocation cleanup; local macOS and Linux receipts. |
| RES-08 | 03, 07 | Refuse or immediately clean discarded ownership | ✓ SATISFIED | Source and pre-serialization refusal tests pass. |
| RES-09 | 02, 03, 06, 07 | Independent acquisition-seeded cleanup validation | ✓ SATISFIED | Independent peer mutation and physical observer controls pass. |

No requirement mapped to Phase 23 in `REQUIREMENTS.md` is orphaned from the plans.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---|---|---|---|
| — | — | None found in the implementation and integration files scanned. | — | — |

## Human Verification Required

None. The phase explicitly defines these command, resource-lifecycle, and observer behaviors as automated checks; no conversational UAT artifact is present or required.

## Gaps Summary

The five ROADMAP success criteria are objectively verified by implementation inspection and automated evidence. Phase completion still has one external evidence blocker: the hosted Ubuntu `evidence-aggregate` receipt. The workflow is wired and local Linux ARM64 evidence exists, but the validation contract explicitly says a local container does not substitute for hosted Ubuntu CI and requires that receipt before phase completion. No Git remote is configured in this checkout, and no hosted receipt is available. Keep validation in progress and do not transition or complete the phase until the hosted lane passes.

### Decision Coverage

The phase context decisions D-23-01 through D-23-06 are represented in the implementation, focused tests, README, and host validation contract. This is a source-inspection assessment; no decision-coverage query receipt was available.

---

_Verified: 2026-09-28T16:00:00Z_  
_Verifier: the agent (gsd-verifier)_
