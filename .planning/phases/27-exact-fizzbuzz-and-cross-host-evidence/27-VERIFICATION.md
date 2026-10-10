---
phase: 27-exact-fizzbuzz-and-cross-host-evidence
verified: 2026-10-10T00:00:00Z
status: passed
score: 5/5 must-haves verified
covered_files:
  - ".github/workflows/ci.yml"
  - ".planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/27-01-PLAN.md"
  - ".planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/27-01-SUMMARY.md"
  - ".planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/27-02-PLAN.md"
  - ".planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/27-02-SUMMARY.md"
  - ".planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/27-03-PLAN.md"
  - ".planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/27-03-SUMMARY.md"
  - ".planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/27-04-PLAN.md"
  - ".planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/27-04-SUMMARY.md"
  - ".planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/27-05-PLAN.md"
  - ".planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/27-05-SUMMARY.md"
  - ".planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/27-06-PLAN.md"
  - ".planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/27-06-SUMMARY.md"
  - ".planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/27-07-PLAN.md"
  - ".planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/27-07-SUMMARY.md"
  - ".planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/27-08-PLAN.md"
  - ".planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/27-08-SUMMARY.md"
  - ".planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/27-09-PLAN.md"
  - ".planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/27-09-SUMMARY.md"
  - ".planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/27-10-PLAN.md"
  - ".planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/27-10-SUMMARY.md"
  - "AGENTS.md"
  - "README.md"
  - "cmd/schway/main_test.go"
  - "examples/fizzbuzz.schway"
  - "examples/phase27/README.md"
  - "examples/phase27/fizzbuzz-1000.out"
  - "examples/phase27/fizzbuzz-15.out"
  - "examples/sum_to_n.schway"
  - "internal/compiler/ast/ast.go"
  - "internal/compiler/cgen/cgen.go"
  - "internal/compiler/cgen/cgen_phase27_evidence_test.go"
  - "internal/compiler/cgen/cgen_phase27_output_test.go"
  - "internal/compiler/cgen/cgen_program.go"
  - "internal/compiler/check/scalar.go"
  - "internal/compiler/core/core.go"
  - "internal/compiler/core/core_convention_absence_test.go"
  - "internal/compiler/core/core_test.go"
  - "internal/compiler/corevalidate/corevalidate.go"
  - "internal/compiler/execution/execution.go"
  - "internal/compiler/executionpeer/executionpeer.go"
  - "internal/compiler/executionpeer/executionpeer_phase27_test.go"
  - "internal/compiler/interp/interp.go"
  - "internal/compiler/interp/interp_phase27_evidence_test.go"
  - "internal/compiler/interp/interp_phase27_output_test.go"
  - "internal/compiler/native/native.go"
  - "internal/compiler/native/native_app.go"
  - "internal/compiler/native/native_app_test.go"
  - "internal/compiler/originvalidate/originvalidate.go"
  - "internal/compiler/session/session.go"
  - "internal/compiler/session/session_phase27_docs_test.go"
  - "internal/compiler/session/session_phase27_evidence_test.go"
  - "internal/compiler/session/session_phase27_host_test.go"
  - "internal/compiler/session/session_phase27_output_test.go"
  - "internal/compiler/session/session_phase27_remainder_test.go"
  - "internal/compiler/session/session_phase27_test.go"
  - "internal/compiler/syntax/format.go"
  - "internal/compiler/syntax/lexer.go"
  - "internal/compiler/syntax/parser.go"
  - "internal/compiler/syntax/token.go"
  - "scripts/check-phase27-receipts.sh"
  - "scripts/verify-phase27.sh"
  - "testdata/phase16/public-emitter-consumers.json"
covered_digest: "v3:sha256:b8b5086986fc50a7b476f01f27d9bd5ff25fb4abeb5c6154df483cfae36f758c"
behavior_unverified: 0
overrides_applied: 0
---

# Phase 27: Exact FizzBuzz and Cross-Host Evidence Verification Report

**Phase Goal:** A developer can run bounded FizzBuzz from ordinary Schway source with exact text and defined remainder behavior, while both M005 programs have reproducible interpreter/native evidence and clean-checkout commands.
**Verified:** 2026-10-10
**Status:** passed
**Re-verification:** No, initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | FizzBuzz is ordinary Schway source (`examples/fizzbuzz.schway`) producing exact text for 1..n (APP-08) | VERIFIED | Source read: `%`, `==`, `<`, if/else, `write_text`/`write_u64`/`write_newline`. Built locally with `schway build`; `app run -- 15` and `-- 1000` are byte-identical (`cmp`) to committed fixtures `fizzbuzz-15.out`/`fizzbuzz-1000.out`; `-- 0` writes 0 bytes; fixture ends with `\n`. |
| 2 | Inputs 0..1000 accepted; 1001+ fail with no app output; output bounded by 65,536 B, overflow is non-success (APP-09) | VERIFIED | Local `-- 1001` gives exit 65, stderr `schway app: input exceeds 1000`, empty stdout. Focused tests `TestPhase27*` pass in session, cmd/schway and cgen (including `TestPhase27OutputExactLimit`, `OutputCrossingWrite`, `OutputPublicStatus`). Hosted receipts show 1001 gives exit 65 on both hosts for both programs. |
| 3 | U64 comparison/equality/Bool/remainder have one meaning across check, validate, interp and C17; constant zero divisor rejected, dynamic zero checked failure (U64-02) | VERIFIED | Changes present in `check/scalar.go`, `corevalidate`, `interp`, `cgen`, `execution`, `executionpeer`; remainder/zero tests in `session_phase27_remainder_test.go`, `executionpeer_phase27_test.go`; focused session tests pass; hosted control `dynamic-zero` (`TestPhase27DynamicZero`) pass 3/3 on both hosts. Guide documents `type.scalar_remainder_zero` and runtime exit 65 `U64 remainder by zero`. |
| 4 | Interpreter and native each compared to independently pinned expectations incl. boundaries and reached controls; macOS and Linux native covered (EVD-12) | VERIFIED | Hosted run 38052601527 at `fa5a169216b6...` (confirmed via `gh run view`: conclusion success; both `checks` and both `current evidence aggregate` jobs success on macos-latest and ubuntu-latest). Receipt gate re-run by me with `PHASE27_EXPECTED_REVISION=fa5a169...`: `status=pass hosts=macOS,Linux`, source and fixture hashes match. Both receipts: 10/10 cases, 3/3 controls, `tree=clean`, distinct hosts (arm64-apple-darwin25.6.0, x86_64-pc-linux-gnu). Only `.planning/` files changed after `fa5a169` (`git diff --name-only fa5a169 HEAD` outside `.planning/` is empty), so source hashes are unchanged. Earlier failing run 38052011600 is recorded as superseded, not claimed. Phase 25/26 receipts not cited. |
| 5 | From a clean checkout, documented commands build and run both programs and locate expected results and limits (DX-16) | VERIFIED | `examples/phase27/README.md` read: build/run commands, exact answers, bounds, limits, controls, gate; root `README.md` links it; `TestPhase27GuideContract` exists in `session_phase27_docs_test.go`. I executed the guide's build/run/cmp commands for FizzBuzz successfully (scratch dir, not committed). |

**Score:** 5/5 truths verified (0 present, behavior-unverified)

### Requirements Coverage

| Requirement | Source Plans | Status | Evidence |
|-------------|--------------|--------|----------|
| U64-02 | 27-03, 27-04, 27-06, 27-08, 27-10 | SATISFIED | Truth 3 |
| APP-08 | 27-01, 27-02, 27-05, 27-06, 27-07, 27-08, 27-09, 27-10 | SATISFIED | Truth 1 |
| APP-09 | 27-01, 27-02, 27-06, 27-07, 27-08, 27-09, 27-10 | SATISFIED | Truth 2 |
| EVD-12 | 27-07, 27-08, 27-09, 27-10 | SATISFIED | Truth 4 |
| DX-16 | 27-09, 27-10 | SATISFIED | Truth 5 |

All five IDs appear in plan frontmatter and are marked Complete for Phase 27 in REQUIREMENTS.md. No orphaned requirements.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| FizzBuzz 15 / 1000 exact bytes | `schway app run ./fb -- N \| cmp - fixture` | identical | PASS |
| Input 1001 refused | `schway app run ./fb -- 1001` | exit 65, no stdout | PASS |
| Focused Phase 27 tests | `go test ./internal/compiler/session ./cmd/schway ./internal/compiler/cgen -run '^TestPhase27'` | all ok | PASS |
| Receipt gate | `check-phase27-receipts.sh` with expected revision | status=pass | PASS |

Full suites were not run locally by design. The hosted run executed the full and race lanes on both hosts.

### Anti-Patterns Found

None. No unreferenced TBD/FIXME/XXX markers in files changed by the phase.

### Human Verification Required

None.

### Gaps Summary

No gaps. Note for provenance: the hosted evidence is pinned to `fa5a169`; later commits are planning-only, so it still describes the current source. Any future source change requires fresh hosted receipts.

---

_Verified: 2026-10-10_
_Verifier: Claude (gsd-verifier)_
