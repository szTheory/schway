---
phase: 28-exit-classes-verify-report-and-entry-pruning
verified: 2026-10-11T19:30:00Z
status: passed
score: 5/5 must-haves verified
covered_files:
  - ".planning/phases/28-exit-classes-verify-report-and-entry-pruning/28-01-PLAN.md"
  - ".planning/phases/28-exit-classes-verify-report-and-entry-pruning/28-01-SUMMARY.md"
  - ".planning/phases/28-exit-classes-verify-report-and-entry-pruning/28-02-PLAN.md"
  - ".planning/phases/28-exit-classes-verify-report-and-entry-pruning/28-02-SUMMARY.md"
  - ".planning/phases/28-exit-classes-verify-report-and-entry-pruning/28-03-PLAN.md"
  - ".planning/phases/28-exit-classes-verify-report-and-entry-pruning/28-03-SUMMARY.md"
  - ".planning/phases/28-exit-classes-verify-report-and-entry-pruning/28-04-PLAN.md"
  - ".planning/phases/28-exit-classes-verify-report-and-entry-pruning/28-04-SUMMARY.md"
  - ".planning/phases/28-exit-classes-verify-report-and-entry-pruning/28-05-PLAN.md"
  - ".planning/phases/28-exit-classes-verify-report-and-entry-pruning/28-05-SUMMARY.md"
  - ".planning/phases/28-exit-classes-verify-report-and-entry-pruning/28-06-PLAN.md"
  - ".planning/phases/28-exit-classes-verify-report-and-entry-pruning/28-06-SUMMARY.md"
  - ".planning/phases/28-exit-classes-verify-report-and-entry-pruning/28-07-PLAN.md"
  - ".planning/phases/28-exit-classes-verify-report-and-entry-pruning/28-07-SUMMARY.md"
  - ".planning/phases/28-exit-classes-verify-report-and-entry-pruning/28-08-PLAN.md"
  - ".planning/phases/28-exit-classes-verify-report-and-entry-pruning/28-08-SUMMARY.md"
  - "README.md"
  - "cmd/schway/main.go"
  - "internal/compiler/callgraph/callgraph.go"
  - "internal/compiler/callgraph/prune.go"
  - "internal/compiler/exitcode/exitcode.go"
  - "internal/compiler/native/native.go"
  - "internal/compiler/native/native_phase28_test.go"
  - "internal/compiler/session/session.go"
  - "internal/compiler/session/session_admit.go"
  - "internal/compiler/session/session_app_verify.go"
  - "internal/compiler/session/session_entry.go"
  - "internal/compiler/session/session_phase28_native_relpath_test.go"
covered_digest: "v3:sha256:29eca7fa30e3984eed1712c8cfd40a96372818d10f7d14bd36878ecbf8746dd9"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: 4/4
  gaps_closed:
    - "G-28-3: run --engine=native with a relative path exited 3 native.tool_failure instead of the correct class"
  gaps_remaining: []
  regressions: []
---

# Phase 28: Exit Classes, Verify Report, and Entry Pruning Verification Report

**Phase Goal:** A developer gets the same exit class and refusal code for the same source from every command and engine, can run a program that contains functions `main` never calls, and sees `app verify` name the capacity limit instead of losing it.
**Verified:** 2026-10-11
**Status:** passed
**Re-verification:** Yes, after UAT gap G-28-3 (plan 28-08) and review fix CR-01 (eface8a)
**Hosted CI:** not run for this revision; all evidence is local focused tests plus the shipped binary.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| SC1 | Same exit class and code for the same source from `check`, `build`, `run` on both engines; tool/IO/timeout exit 3; mismatch exit 4 | VERIFIED | `TestExitClassesUniformAcrossSurfaces` passes (absolute rows for duplicate_function_name, foreign_origin_omitted, payload_borrow_interaction, plus new relative rows). `TestNativeToolFailureIsOperational` passes. |
| SC2 | Exit codes come from one named table; no duplicated literals | VERIFIED | `internal/compiler/exitcode` tests pass; session `TestNoDuplicatedExitLiterals`, `TestExitCodeDocsTable`; cgen `TestEmittedCExitLiteralsMatchTable` pass. |
| SC3 | Function unreachable from `main` passes check, builds and runs; ambiguous entry exits 2; reachable programs keep byte-identical C | VERIFIED | `TestUnreachableFunctionProgramRuns` (interpreter and native), `TestPhase28EntryExitClassManifest`, `TestPhase28PruneCorpusProperties`, `TestPhase28PrunedEmitMatchesGolden`, cgen `TestPhase272AcceptedCorpusByteIdentity` and `TestPhase273ConformanceCorpusByteIdentity` pass. Binary: `check examples/phase28/unused_function.schway` exits 0; ambiguous `payload_borrow_interaction` exits 2 `core.entry_ambiguous` with roots named and `add fn main` hint. |
| SC4 | `app verify` names `native.event_capacity_exhausted` for exhausted cases, within size limit, never success | VERIFIED | `TestAppVerifyExhaustedCases`, `TestAppVerifySixteenExhausted`, `TestAppVerifyReportBudget`, `TestAppVerifyReportOversize`, `TestAppVerifyOutcomeDigest`, `TestPhase273AppVerify*` pass. |
| G-28-3 | Relative-path `run --engine=native` reaches Check and exits in the correct class; valid program exits 0 | VERIFIED | See G-28-3 closure below. |

**Score:** 5/5 truths verified, 0 behavior-unverified.

### G-28-3 Closure (shipped binary, repo root, built with `go build -o /tmp/schway ./cmd/schway`)

| Command | Expected | Observed |
|---------|----------|----------|
| `run --engine=native testdata/phase07/duplicate_function_name.schway` | exit 2 `core.duplicate_function_id` | exit 2, `core.duplicate_function_id` |
| `run --engine=native testdata/phase07/call_basic.schway` | exit 0 | exit 0, `run pass` |
| `run --engine=interpreter` (dup, relative) | exit 2 | exit 2 |
| `check` (dup, relative) | exit 2 | exit 2 |
| `build ... --output /tmp/dup` (dup, relative) | exit 2, no artifact | exit 2, `/tmp/dup` absent |
| `run --engine=native testdata/phase12/payload_borrow_interaction.schway` | exit 2 `core.entry_ambiguous` | `core.entry_ambiguous` invalid (the pipe hid the numeric code, but the interpreter and the uniform-surface test pin exit 2) |

Code evidence: `grep` finds no `filepath.Rel(` or `nat03ProjectRoot` in `internal/compiler/session/session.go`. The 28-08 diff touches only `session.go`, `session_phase28_native_relpath_test.go` and `native_phase28_test.go`. Regression pins: `TestPhase28NativeCommandRelativePath` (session) and the `relative_path/duplicate_function_name` and `relative_path/native_positive_control` subtests of `TestExitClassesUniformAcrossSurfaces` all pass. The old gap (absolute-only coverage) is closed.

### CR-01 (emitter inventory pin)

`TestPhase16PublicEmitterConsumerInventory` passes at HEAD, and the session meta-tests `TestEntryFunctionCalledOnlyViaResolveEntry`, `TestVerificationGroundedness*`, `TestPhase20ValidationLifecycle` and `TestStaticTestIndexMatchesGoTestList` pass. WR-01 (line-number pins are fragile) is deferred as a process finding; it is not a goal gap.

### Previously carried warnings

WR-02 (report-write error masking a source refusal) and WR-04 are now fixed: `session_app_verify.go:229` uses `errors.Join(verifyErr, err)`; `TestAppVerifyRefusalSurvivesReportWriteFailure` and `TestAppVerifyRemovesStaleReportOnRefusal` pass. The earlier verification report listed these as open.

### Requirements Coverage

| Requirement | Source Plans | Status | Evidence |
|-------------|--------------|--------|----------|
| DX-17 | 28-01, 28-02, 28-06, 28-07, 28-08 | SATISFIED | SC1, SC2, G-28-3 closure; relative and absolute paths both uniform |
| DX-18 | 28-01, 28-04, 28-05, 28-08 | SATISFIED | SC3 |
| EVD-13 | 28-03, 28-07 | SATISFIED | SC4 |

All three IDs are in plan frontmatter and marked Complete in REQUIREMENTS.md (lines 16-18, 84-86). No orphaned Phase 28 requirements.

### Anti-Patterns Found

None blocking. No unreferenced TBD/FIXME/XXX introduced by 28-08.

### Human Verification Required

None.

### Gaps Summary

No gaps. G-28-3 is closed and verified on the shipped binary and by pinned tests. Caveat: only focused tests were run (by instruction), not the full suite, and hosted CI has not run.

---

_Verified: 2026-10-11_
_Verifier: Claude (gsd-verifier)_
