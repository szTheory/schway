---
phase: separate-pointer-successors-and-integrated-utility
verified: "2026-10-02T02:23:56Z"
status: gaps_found
score: 4/5 roadmap success criteria verified
covered_files:
  - .github/workflows/ci.yml
  - .planning/REQUIREMENTS.md
  - .planning/ROADMAP.md
  - .planning/PRODUCT-ROADMAP.md
  - .planning/LANGUAGE-MATURITY.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-CONTEXT.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-01-PLAN.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-01-SUMMARY.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-02-PLAN.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-02-SUMMARY.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-03-PLAN.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-03-SUMMARY.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-04-PLAN.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-04-SUMMARY.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-05-PLAN.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-05-SUMMARY.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-06-PLAN.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-06-SUMMARY.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md
  - examples/phase24/README.md
  - examples/phase24/transfer.schway
  - examples/phase24/error.schway
  - examples/phase23/file_byte.bindings.json
  - examples/phase23/adapter.c
  - examples/phase23/adapter.h
  - scripts/verify-phase25.sh
  - internal/compiler/check/check.go
  - internal/compiler/check/check_test.go
  - internal/compiler/core/core.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/cgen/cgen_pointer_successor_test.go
  - internal/compiler/interp/interp.go
  - internal/compiler/native/phase25_utility_test.go
  - internal/compiler/native/phase25_pointer_successor_test.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/originvalidate/originvalidate_test.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/pathoracle/pathoracle_test.go
  - internal/compiler/session/session_admission_divergence_test.go
  - internal/compiler/session/verification_groundedness_test.go
covered_digest: "v1:sha256:8207e637325a3b7f5f8f1792924e18684f9e0d5b80468a852ef92114f59de3b5"
behavior_unverified: 0
overrides_applied: 0
---

# Phase 25: Separate Pointer Successors and Integrated Utility — Verification Report

**Phase Goal:** A developer can use separately checked shared and exclusive read-copy pointer helpers in the documented native utility, with reproducible evidence for each admitted family.
**Verified:** 2026-10-02T02:23:56Z
**Status:** gaps_found
**Re-verification:** Initial goal-backward review after all six implementation plans.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | A source consumer invokes the bounded shared read/copy helper through a real pointer-parameter C ABI, and independent checking rejects conflicting access and escape. | ✓ VERIFIED | Source inspection of the shared admission and peer predicates. `TestPhase25SharedPointerCopyABI`, `TestPhase25UtilityOwnerTransfer`, `TestPhase25UtilityNative`, and separate conflict/escape controls pass. The local macOS/arm64 evidence script reports 65/66 in all three lanes. |
| 2 | A distinct source consumer invokes the bounded exclusive read/copy helper through its pointer-parameter C ABI, and independent checking rejects conflicting access and escape. | ✓ VERIFIED | `TestPhase25ExclusivePointerRefusal`, `TestPhase25UtilityOwnerTransfer`, `TestPhase25UtilityNative`, and separate exclusive conflict/escape controls pass. The local macOS/arm64 evidence script reports 65/66 in all three lanes. |
| 3 | Emitted C and manifests agree without unsupported alias/capture/alignment promises; unsupported mutation, forwarding, retention, callbacks, nonlocal exits, and wider forms are refused before serialization. | ✓ VERIFIED | C-generation and manifest contract tests, serializer refusal controls, and the complete local Go suite pass. No pointer alias or ownership attributes were added. |
| 4 | Foreign, shared, and exclusive families each have reproducible native receipts on macOS and Linux for every applicable optimizer/sanitizer lane. | OPEN — HOSTED EVIDENCE | The local macOS/arm64 script passes baseline `-O0`, optimized `-O2`, and ASan+UBSan. It explicitly reports all nine Linux family/lane rows as `incomplete`. No hosted Phase 25 CI receipt is available in this checkout. |
| 5 | A clean checkout can follow the README to build and run the integrated utility, observe 65/66 and the typed failure, and locate its C inputs and evidence. | ✓ VERIFIED | The documented zsh/POSIX-safe commands use the explicit binding manifest. Local clean-checkout smoke returned 65/66 and `UseError.UnsupportedByte` for 0x43 before either helper; README/evidence-index contract tests pass. |

**Score:** 4/5 roadmap truths verified; the remaining truth is the required hosted native host matrix, not an untested local behavior.

### Six-Plan Cross-Check

All six plan files and summaries were reviewed. Shared/exclusive admission, independent core/origin/path checks, exact C lowering, bounded owner transfer, model behavior, structured diagnostics, README commands, bounded subprocesses, and the CI evidence gate have named witnesses. The evidence gate itself passes locally and fails closed in its matrix records when a host is absent. The hosted requirement remains open until that gate runs on the configured native hosts.

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| Shared and exclusive source witnesses | Distinct checked read/copy families | ✓ VERIFIED | Phase 25 fixtures and integrated utility produce separate family-specific operations and controls. |
| `internal/compiler/{corevalidate,originvalidate,pathoracle}` | Independent family, origin, conflict, and escape validation | ✓ VERIFIED | Focused phase tests and the full local suite pass; unknown and malformed places remain rejected. |
| `internal/compiler/cgen` | Actual pointer C and exact manifest; fail before serialization for unsupported shapes | ✓ VERIFIED | ABI/manifest tests and negative controls pass. The exact shape carries no unsupported optimizer promise. |
| `examples/phase24/README.md` | Reproducible clean-checkout utility path and directly indexed evidence | ✓ VERIFIED | Explicit source, bindings, result/error behavior, and evidence pointers are present and contract-tested. |
| `scripts/verify-phase25.sh`, CI workflow | Fail-closed family × host × lane evidence under existing aggregate ownership | PARTIAL | One invocation is wired into the existing aggregate. Local macOS lanes pass; Linux rows are marked incomplete pending hosted execution. |
| PRODUCT-ROADMAP and LANGUAGE-MATURITY | Evidence-calibrated boundaries and three ranked next capabilities | ✓ VERIFIED | Current source facts, local evidence, historical receipts, future slices, checker changes, owners, and reprioritization conditions are distinguished. |

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| Phase 25 source helpers | checker and core operation facts | source admission | WIRED | Tests inspect each family and exact U64 copy shape. |
| checked borrow facts | origin and path peers | independent re-derivation | WIRED | Mutation and conflict/escape cases exercise both families. |
| checked pointer ABI facts | sole C serializer and manifest | preflighted lowering | WIRED | ABI and manifest derive from the same bounded facts; unsupported cases fail before serialization. |
| integrated utility | native utility test and README | caller files and explicit C binding inputs | WIRED | Native tests return 65/66 and preserve the typed 0x43 failure order. |
| evidence script | CI `evidence-aggregate` | one existing job step | WIRED | The workflow owns one invocation; missing host/lane evidence stays incomplete. |

### Behavioral Spot-Checks

| Check | Result | Scope |
|---|---|---|
| `GOCACHE=/tmp/ai-lang-verification-gocache go test -count=1 ./...` | PASS | Full local Go suite; all packages passed. |
| `GOCACHE=/tmp/ai-lang-verification-gocache go test -race -count=1 ./...` | PASS | Full local race suite; all packages passed. |
| `GOCACHE=/tmp/ai-lang-verification-gocache go vet ./...` | PASS | Local static analysis. |
| `GOCACHE=/tmp/ai-lang-verification-gocache go build ./...` | PASS | Local build. |
| `sh scripts/verify-phase25.sh` | Local host PASS; matrix INCOMPLETE | macOS/arm64, Go 1.24.0, Apple Clang 21.0.0; all families and three lanes pass. All nine Linux rows remain incomplete. |
| Groundedness, validation lifecycle, maturity census, README/evidence/CI contract tests | PASS | Focused local session/native tests after final documentation updates. |

The initial repository-wide run found a modeled typed-error path skipping `OpFail`, a stale Phase 24 operation-order assertion, a malformed path-oracle test fixture, an expected origin conflict that the old test accepted, and two newly visible admission divergences. Those were corrected at the owning seams. The normal and race suites subsequently passed. The current local tree is modified from source commit `f00cdf843bb3b7ed93948a47d3b2acd4af3085af`; no hosted workflow is claimed.

### Requirements Coverage

| Requirement | Status | Evidence |
|---|---|---|
| NAT-11 | Locally verified | Shared source, C ABI, peer checks, native results, and negative controls pass on macOS/arm64. |
| NAT-12 | Locally verified | Exclusive source, C ABI, peer checks, native results, and negative controls pass on macOS/arm64. |
| NAT-13 | Locally verified | Integrated utility and shared/exclusive successor chain pass; 0x43 error ordering is preserved. |
| DX-14 | Locally verified | README commands and direct evidence index pass their contract tests and local smoke. |
| DX-15 | Locally verified | Structured diagnostic and ownership-boundary controls pass in the full local suite. |
| EVD-10 | OPEN | Hosted native receipts for macOS and Linux across all applicable family/lane combinations are absent. |

### Gaps Summary

One acceptance gap remains: the current source revision must run through the existing hosted dual-host evidence aggregate so native macOS and Linux receipts can be inspected. The local matrix is complete only for macOS/arm64. Phase 25 must remain open and EVD-10 must remain pending until those hosted receipts pass.

---

_Verified: 2026-10-02T02:23:56Z_
_Verifier: local goal-backward review_
