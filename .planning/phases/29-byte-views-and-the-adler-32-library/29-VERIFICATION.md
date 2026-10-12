---
phase: 29-byte-views-and-the-adler-32-library
verified: 2026-10-11T00:00:00Z
status: passed
score: 5/5 must-haves verified
covered_files:
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-00-PLAN.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-00-SUMMARY.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-01-PLAN.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-01-SUMMARY.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-02-PLAN.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-02-SUMMARY.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-03-PLAN.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-03-SUMMARY.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-04-PLAN.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-04-SUMMARY.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-05-PLAN.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-05-SUMMARY.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-06-PLAN.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-06-SUMMARY.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-07-PLAN.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-07-SUMMARY.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-08-PLAN.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-08-SUMMARY.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-09-PLAN.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-09-SUMMARY.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-10-PLAN.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-10-SUMMARY.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-11-PLAN.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-11-SUMMARY.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-12-PLAN.md
  - .planning/phases/29-byte-views-and-the-adler-32-library/29-12-SUMMARY.md
  - examples/phase29/adapter.c
  - examples/phase29/adler32_fixture.schway
  - examples/phase29/byte_index_oob.schway
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/check/check.go
  - internal/compiler/check/scalar.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/interp/interp.go
  - internal/compiler/native/native_app.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/session/session.go
  - internal/compiler/testsupport/phase29_fixture_bytes.go
covered_digest: "v3:sha256:80628a4056911c49cf740e6a4eaccca6915f51f0d7c8ac0ef2d97220e18373f3"
behavior_unverified: 0
overrides_applied: 0
---

# Phase 29: Byte Views and the Adler-32 Library Verification Report

**Phase Goal:** Ordinary Schway source can read a read-only byte view in a scalar loop, with one defined meaning on every engine and a dynamic bounds check that is never removed, and computes RFC 1950 Adler-32 over it.
**Status:** passed (with 4 open code-review warnings classified below as non-blocking)
**Re-verification:** No, initial verification

## Goal Achievement

### Observable Truths (ROADMAP SC1..SC5)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| SC1 | `adler32(bytes: ByteView) -> U64` in ordinary Schway returns the six pinned literals on interpreter and native | VERIFIED | `examples/phase29/adler32_fixture.schway` uses only `+`, `<`, `% 65521` (the final b * 65536 + a is built from sixteen `r = r + r` doublings, so no multiply or shift). Literals 300286872 and 4052720524 are pinned in `testsupport/phase29_fixture_bytes.go` and the session gate and interpreter tests. `TestPhase29NativeFixtureVectors` builds the artifact and runs cases 0..9 with exit 0 and exact stdout. `TestPhase29InterpreterFixtureVectors` covers the interpreter. Both passed in a focused local run. The `bytes` name versus `data` is an acknowledged dated note (D-18). |
| SC2 | One meaning of byte length and byte-at across check, corevalidate, originvalidate, pathoracle, interp and C17 emission, independently derived; `0x80` and `0xFF` read 128 and 255 | VERIFIED | Phase29 test files exist and pass in check, corevalidate, originvalidate, pathoracle, interp, cgen, executionpeer and session. The review confirms that each peer re-derives the shape without shared helpers, and that the octet widens unsigned through `const unsigned char *`. |
| SC3 | Index at or beyond length exits 65 with stderr exactly `schway: byte index out of bounds` and empty stdout, in the interpreter and native at -O0, -O3, -O3+LTO | VERIFIED | `cgen_phase29_oob_test.go`, `cgen_phase29_gate_test.go` and the interp OOB test pass. The review confirms the bounds check precedes the load, with no `assert` or `restrict`. Hosted run 38186068052 passed on ubuntu and macos. |
| SC4 | Scalar loop function over a byte view with U64/Bool state is admitted; each unsound neighbor refused with its stable code (D-22 amendment recorded) | VERIFIED | Check, corevalidate and pathoracle release-law and neighbor tests pass (TestPhase29ReleaseBeforeViewUseRefused, TestPhase29OwnerLoopBlockRefused, TestPhase29LoopNeighborParseRefusals). The recorded honest limit about an owner move under a foreign-borrow view is a dated, accepted note in ROADMAP. |
| SC5 / Runnable gain | `build` plus `app run` on the fixture natively, and the interpreter via `RunWithFixtureBytes`, give the same pinned checksums; the OOB control fails identically | VERIFIED | `session_phase29_native_fixture_test.go` (exit 0 for cases 0..9, exit 65 and empty stdout for unknown case), `session_phase29_interp_fixture_test.go`, and `examples/phase29/byte_index_oob.schway`. Dated note D-16 records that `run --engine=*` is unchanged. |

**Score:** 5/5 truths verified. Behavior-dependent truths were backed by named passing tests rather than presence.

### Requirements Coverage

| Requirement | Source Plans | Status | Evidence |
|-------------|--------------|--------|----------|
| BYTE-01 | 29-00, 29-01, 29-02, 29-04, 29-05, 29-06, 29-07 and others | SATISFIED | SC2 evidence. |
| BYTE-02 | 29-00, 29-01, 29-03, 29-04 and others | SATISFIED | SC3 evidence. |
| BYTE-03 | 29-00, 29-03, 29-09 | SATISFIED | SC4 evidence. |
| APP-10 | 29-00 through 29-12 | SATISFIED | SC1 evidence. |

All four IDs are declared in PLAN frontmatter and in REQUIREMENTS.md (marked Complete, mapped to Phase 29). No orphaned requirements.

### Test and CI Evidence

| Check | Result |
|-------|--------|
| Focused `go test ./internal/compiler/native ./internal/compiler/interp ./internal/compiler/executionpeer ./internal/compiler/cgen ./internal/compiler/pathoracle ./internal/compiler/syntax ./internal/compiler/originvalidate ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/session ./cmd/schway -run 'Phase29' -count=1` | All pass (local; the full suite was not run, per project policy) |
| Hosted run 38186068052 at 906d648 (spike008_strict=true) | `go test ./...`, `-race`, evidence aggregate and the strict budget passed on ubuntu and macos (from the task brief, not re-fetched) |
| `git diff --name-only 906d648 HEAD` | Docs, spike README/MANIFEST, and the text-only `session/qlt01_registry.json`. No implementation changes after the hosted run. |
| Spike 008 | VALIDATED, cap kept at 65,536 |

### Anti-Patterns and Code Review Warnings

No critical findings. Open review warnings, classified against the goal:

| ID | Classification | Reasoning |
|----|----------------|-----------|
| WR-02 (`fn main(b: ByteView)` checks but no engine runs it) | Acceptable, out of scope for the Phase 29 goal. Warning to fix or defer. | The goal and SC2 concern the meaning of `byte_len` and `byte_at`, and the views the goal admits come from the fixture owner. A ByteView entry parameter is not a surface Phase 29 promises. Every engine refuses it, so no engine gives it a divergent meaning. The defect is a missing typed diagnostic (opaque `native.tool_failure` and `tool.run_failed`). Fix with a typed refusal in check, or track it for Phase 31 when the real file-adapter entry binding lands. |
| WR-03 (native aborts exit 65 on any defect with a live fixture owner; interp errors for non-OpBytesAt defects) | Acceptable, not a goal gap, but a real narrow divergence. Defer or document. | The divergence is confined to the closed fixture route and to defects other than the bounds defect. SC3 is satisfied on both engines for the bounds defect. The shipped Adler-32 library cannot raise any other defect, because its arithmetic is bounded by the modulus. The interpreter's rejection is a deliberate conservative refusal pinned by the test "defect statement with the owner live stays rejected". It is not a soundness hole. It does mean the differential cannot compare an overflow variant. Track it for a later phase if the fixture route is widened. |
| WR-01 (fixture owner "test-only" but user-reachable) | Acceptable. FFI trust class. | Does not affect the bounds-check guarantee against honest adapters. The "test-only" wording must not be relied on as a security boundary. |
| WR-04 (`disable...ForTest` vars in production packages) | Acceptable. Hygiene. | Consistent with the existing codebase pattern. Not a goal issue. |
| IN-01..IN-04 | Informational | IN-04 (README should state that `run` and `evidence` do not support the fixture owner) is worth a one-line doc fix. |

All 8 review items remain `open` in 29-REVIEW-DISPOSITION.md. None blocks the phase.

### Human Verification Required

None.

### Gaps Summary

No gaps against the phase goal. Recommended follow-ups, not blockers: add a typed refusal for ByteView entry parameters (WR-02), reconcile or document the fixture-route defect divergence (WR-03), add the README note (IN-04), and triage the review dispositions.

---

_Verifier: Claude (gsd-verifier)_
