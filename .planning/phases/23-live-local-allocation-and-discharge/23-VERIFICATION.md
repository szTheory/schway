---
phase: 23-live-local-allocation-and-discharge
verified: 2026-10-01T12:28:24Z
status: passed
score: 5/5 roadmap success criteria verified
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
  - cmd/schway/main.go
  - cmd/schway/main_test.go
  - examples/phase23/README.md
  - examples/phase23/adapter.c
  - examples/phase23/adapter.h
  - examples/phase23/file_byte.bindings.json
  - examples/phase23/file_byte.expected.json
  - examples/phase23/file_byte.schway
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
  - testdata/phase23/discard_owner.schway
covered_digest: "v1:sha256:0675c6277581e232550158b03d2e284406203be69c454526718953579f152722"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 5/5 roadmap truths verified
  gaps_closed:
    - "The configured hosted Ubuntu evidence-aggregate lane has a passing Phase 23 receipt."
  gaps_remaining: []
  regressions: []
---

# Phase 23: Live Local Allocation and Discharge — Verification Report

**Phase Goal:** A developer can read a caller-selected file byte through a real allocation returned live to Schway and observe its generated local cleanup.
**Verified:** 2026-10-01T12:28:24Z
**Status:** passed
**Re-verification:** Yes — freshness refresh against the hosted-tested source revision; no tests or hosted workflows were rerun.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | An opaque noncopyable resource receives bounded malloc storage, remains live through Schway-directed use, yields the independently expected byte, and is physically destroyed before exit. | ✓ VERIFIED | Source inspection: `examples/phase23/file_byte.schway`, `adapter.c`, and `adapter.h` implement separate acquisition, use, and consuming release. `cmd/schway/main_test.go#TestPhase23PublicFileByte` covers caller-selected byte files; `internal/compiler/native/phase23_observer_test.go#TestPhase23ObserverPublicLifecycleAndUseFailure` independently observes allocation, use, cleanup, and the post-acquisition use error. The hosted focused gate passed on Ubuntu and macOS at the tested source SHA. |
| 2 | Acquire, use, and infallible consuming release each use their own checked signature, operand/failure facts, and release pairing. | ✓ VERIFIED | Source inspection: operation contracts in `core.LinearOperation`, admission/replay in `check`, `corevalidate`, `originvalidate`, and `pathoracle`, emission in `cgen`, and per-symbol C prototype probes in `native/bindings.go`. `TestPhase23OperationABI` and peer/emitter refusal tests are included in the hosted focused gate. |
| 3 | Empty, maximum-size, oversized, malformed, and failed inputs obey the buffer/error contract; failed acquisition creates no owner and partial storage is cleaned. | ✓ VERIFIED | Source inspection: bounded path/file handling and initialized empty failure owner in `adapter.c`; injected failures and partial cleanup in `native_app_test.go`; public cases in `cmd/schway/main_test.go`. Named `TestPhase23AcquireFailuresInitializeAndFreePartialAllocations` and related acquisition/public tests are exercised by the focused gate. |
| 4 | Discarded owning success is rejected or immediately destroyed; independent acquisition-seeded validators reject missing, duplicate, wrong-resource, or fabricated cleanup even if every release is removed from candidate core. | ✓ VERIFIED | Independent obligation derivation is present in `corevalidate`, `originvalidate`, and `pathoracle`; source/emitter discard refusals are in `check` and `cgen`. Named peer mutation tests cover all-release-deleted, duplicate, premature, wrong-resource, and fabricated cases and are included in the hosted focused gate. |
| 5 | Normal completion and a real post-acquisition use/output failure clean each remaining owner once; independent physical observation rejects omitted or premature destruction despite plausible compiler events. | ✓ VERIFIED | `TestPhase23ObserverPublicLifecycleAndUseFailure` executes normal and use-failure paths. `TestPhase23ObserverMutationControlsPreservePlausibleEvents` exercises omitted, premature, duplicate, and wrong-resource physical destruction controls. These tests ran under the passing hosted Phase 23 gate. |

**Score:** 5/5 roadmap success criteria verified (0 behavior-unverified).

### Seven-Plan Cross-Check

All seven plan files and summaries were reviewed against their implementation paths and the phase criteria. The named Phase 23 tests exist in the source and the hosted `scripts/verify-phase23.sh` step passed, supplying current behavioral evidence for the testable claims.

| Plan | Claim checked against implementation/evidence | Result |
|---|---|---|
| 23-01 | Public retained file-byte route, distinct operation contracts, and 0x41/0x42 expected outputs | ✓ VERIFIED |
| 23-02 | Independent acquire-seeded core, origin, and path validation plus reached mutations | ✓ VERIFIED |
| 23-03 | Discard/copy/moved-from/escape/unsupported-exit refusals before C serialization | ✓ VERIFIED |
| 23-04 | Bounded input, typed acquisition failures, no owner on failure, and partial cleanup | ✓ VERIFIED |
| 23-05 | Interpreter outcomes remain model-only and successor ownership contract distinguishes release, borrow, and transfer | ✓ VERIFIED |
| 23-06 | Independent native allocation/use/free observer and reached physical mutation controls | ✓ VERIFIED |
| 23-07 | Published runnable witness, expected answers, and focused macOS/Linux CI integration | ✓ VERIFIED |

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `examples/phase23/file_byte.schway` | Public checked local-owner flow | ✓ VERIFIED | Source-constructible opaque owner and distinct acquire, borrowed use, and consuming release operations. |
| `examples/phase23/adapter.c`, `adapter.h`, `file_byte.bindings.json` | Bounded adapter and explicit pinned ABI | ✓ VERIFIED | Adapter reads caller path, returns live allocation, exposes contracted use/free operations, and is compiled against per-symbol declarations. |
| `internal/compiler/{check,corevalidate,originvalidate,pathoracle,cgen}` | Admission, independent validation, and sole emission | ✓ VERIFIED | Each layer has phase-specific validation/refusal coverage; independent peers seed obligations from acquisition operations. |
| `internal/compiler/native/phase23_observer_test.go` | Physical lifetime evidence | ✓ VERIFIED | Independently observes malloc/use/free/exit and rejects reached lifecycle mutations. |
| `scripts/verify-phase23.sh` and `.github/workflows/ci.yml` | Focused host evidence | ✓ VERIFIED | Script is wired in `current evidence aggregate` for Ubuntu and macOS; both hosted invocations passed. |
| `examples/phase23/README.md` and expected-answer fixture | Clean-checkout usage and independent expected results | ✓ VERIFIED | Contract tests are among the hosted focused gate. |

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| `cmd/schway/main.go` | native retained app runner | CLI dispatch | WIRED | Public `TestPhase23PublicFileByte` invokes the path with caller-selected files. |
| Generated native entry | checked acquire/use/release | `.schway` source plus manifest | WIRED | Public and independent observer tests exercise the emitted app and local cleanup. |
| Foreign operation contract | adapter declaration | per-symbol compiled ABI probe | WIRED | `TestPhase23OperationABI` covers acquire, use, and release prototypes. |
| Independent acquire operation | peer-derived obligation | core/origin/path replay | WIRED | Peer tests and mutation controls exercise discharged and invalid paths. |
| Ubuntu/macOS evidence aggregate | `scripts/verify-phase23.sh` | workflow step | WIRED | Hosted job metadata records the named step successful on both hosts. |

### Data-Flow Trace (Level 4)

| Artifact | Data variable | Source | Produces real data | Status |
|---|---|---|---|---|
| File-byte result | owner pointer and loaded byte | Caller file → adapter read → malloc-backed owner → Schway-directed use | Yes | ✓ FLOWING |
| Generated cleanup | owner pointer | Matching consuming release → adapter `free` | Yes; observed independently in native process | ✓ FLOWING |

### Behavioral Spot-Checks

No local project suite or project script was run during this report refresh. Behavioral evidence is the hosted CI receipt below at source revision `ed94ef7972b25deb85ab90fafbf3403dc31629f5`. Per the supplied execution record, implementation sources are unchanged since that run; the local checkout is at `ee1190031484dfb7e468db0b597d954802c06b8b` and its current uncommitted changes are planning metadata only. The hosted receipt is historical evidence, not a run performed during this refresh.

| Behavior | Command/evidence | Result | Status |
|---|---|---|---|
| Full CI checks | Hosted workflow run [36856048690](https://github.com/szTheory/schway/actions/runs/36856048690), `checks` jobs on Ubuntu and macOS | Vet, build, full tests, and race suites passed on both hosts at `ed94ef7972b25deb85ab90fafbf3403dc31629f5` | ✓ PASS (historical hosted receipt) |
| Phase 23 focused evidence aggregate | Same hosted run, `current evidence aggregate` on Ubuntu and macOS, step `scripts/verify-phase23.sh` | Step passed on both hosts at `ed94ef7972b25deb85ab90fafbf3403dc31629f5` | ✓ PASS (historical hosted receipt) |

### Probe Execution

| Probe | Command | Result | Status |
|---|---|---|---|
| `scripts/verify-phase23.sh` | Hosted workflow step on Ubuntu and macOS | Both step conclusions `success` in run `36856048690`, source SHA `ed94ef7972b25deb85ab90fafbf3403dc31629f5` | PASS (historical hosted receipt) |

The optional validation-corpus receipt job was skipped by workflow configuration. It is outside Phase 23's hosted Ubuntu focused-evidence criterion and is not treated as a missing Phase 23 receipt.

Historical/local receipts retained from the earlier validation record (not rerun during this refresh):

| Receipt | Environment and revision | Result | Scope |
|---|---|---|---|
| Focused aggregate, 2026-09-28 | macOS Darwin/arm64, revision `c50430d9fa1490b393c5d22805f094787ee99186` | Pass, 8 s | Local focused Phase 23 evidence after the close-failure fix. |
| Focused aggregate, 2026-09-28 | Local Docker Linux ARM64, Go 1.24.13, Debian Clang 14.0.6, revision `c50430d` | Pass, 14 s | Container host evidence; not a hosted Ubuntu receipt. |
| Full Go suite, 2026-09-28 | Local macOS after the close-failure fix | Pass (historical) | Distinct from the hosted run and the focused host receipts. |

### Requirements Coverage

| Requirement | Source plans | Description | Status | Evidence |
|---|---|---|---|---|
| FFI-03 | 01, 02, 03, 05, 07 | Each foreign operation has its own checked signature, operand/failure facts, and release pairing | ✓ SATISFIED | Operation contracts, independent peers, compiled ABI probes, hosted focused gate. |
| RES-04 | 01, 02, 04, 05, 06, 07 | Real bounded allocation remains live and supplies caller-selected byte through Schway use | ✓ SATISFIED | Public success tests and independent physical observer; hosted Ubuntu/macOS focused receipts. |
| RES-07 | 01, 04, 06, 07 | Failed acquisition creates no owner and adapter cleans partial allocation under published bounds | ✓ SATISFIED | Source and injected failure cases; hosted focused gate. |
| RES-08 | 03, 07 | Discarded owning acquisition is refused or immediately consumed | ✓ SATISFIED | Checker/emitter refusal tests; hosted focused gate. |
| RES-09 | 02, 03, 06, 07 | Independent validation rejects invalid or missing cleanup from acquire-derived obligation | ✓ SATISFIED | Peer mutation tests and physical observer controls; hosted focused gate. |

No additional `REQUIREMENTS.md` item is assigned to Phase 23 outside these five; no mapped requirement is orphaned from the plans.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---|---|---|---|
| — | — | No blocking stub or unreferenced debt marker found in the inspected Phase 23 implementation paths. | — | — |

### Human Verification Required

None. The phase defines observable behavior through the public command, native observer, and independent negative controls; hosted runs exercise those tests. No visual, subjective UX, or external service behavior is required.

### Gaps Summary

The prior report's only gap (a passing hosted Phase 23 evidence aggregate) is closed. Hosted run [36856048690](https://github.com/szTheory/schway/actions/runs/36856048690) passed the full vet/build/test/race checks and the Phase 23 focused script on Ubuntu and macOS at source SHA `ed94ef7972b25deb85ab90fafbf3403dc31629f5`. This is a historical hosted receipt; no tests, probes, or hosted workflows were rerun for this refresh. The current checkout is `ee1190031484dfb7e468db0b597d954802c06b8b`; per the supplied run record, implementation sources are unchanged since the tested revision, while current uncommitted changes are planning metadata only. The tested source commit object is unavailable in this checkout, so no local commit-to-commit comparison was possible. The covered digest was regenerated with `gsd_run query verification.fingerprint` over the exact existing `covered_files` list. No Phase 23 UAT was created or replayed. The optional validation-corpus receipt remains outside this phase's hosted Ubuntu/macOS focused-evidence criterion.

Evidence classes are kept distinct: source inspection establishes wiring and implementation shape; the 2026-09-30 hosted run is the current behavioral/CI receipt; earlier local macOS/Linux-container and full-suite receipts remain historical and were not rerun. The Phase 23 goal and its five roadmap success criteria are achieved.

### Decision Coverage

The decision-coverage query reports all 6 trackable CONTEXT.md decisions honored, with no unhonored decisions. This non-blocking check does not affect status.

---

_Verified: 2026-10-01T12:28:24Z_
_Verifier: the agent (gsd-verifier)_
