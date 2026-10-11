---
phase: 28-exit-classes-verify-report-and-entry-pruning
verified: 2026-10-11T03:10:00Z
status: passed
score: 4/4 must-haves verified
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
  - "README.md"
  - "cmd/schway/main.go"
  - "internal/compiler/callgraph/callgraph.go"
  - "internal/compiler/callgraph/prune.go"
  - "internal/compiler/exitcode/exitcode.go"
  - "internal/compiler/native/native.go"
  - "internal/compiler/session/session.go"
  - "internal/compiler/session/session_admit.go"
  - "internal/compiler/session/session_app_verify.go"
  - "internal/compiler/session/session_entry.go"
covered_digest: "v3:sha256:7c4c863f02b902f7ecc92b830cdd78bf172070fc4be37f3f3558f5bcb36e6b82"
behavior_unverified: 0
overrides_applied: 0
---

# Phase 28: Exit Classes, Verify Report, and Entry Pruning Verification Report

**Phase Goal:** A developer gets the same exit class and refusal code for the same source from every command and engine, can run a program that contains functions `main` never calls, and sees `app verify` name the capacity limit instead of losing it.
**Verified:** 2026-10-11
**Status:** passed (with non-blocking warnings WR-02 and WR-04 carried forward)
**Re-verification:** No, initial verification
**Hosted CI:** has not run for this revision. Nothing below claims a hosted receipt; all evidence is local focused tests.

## Goal Achievement

### Observable Truths (ROADMAP SC1..SC4)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| SC1 | One corevalidate-refusal fixture gives the same exit 2 and code from `check`, `build`, `run` on both engines; `foreign_origin_omitted` exits 2 `core.foreign_origin_omitted` on interpreter and native; tool/IO/timeout exit 3; mismatch exit 4 | VERIFIED | `TestExitClassesUniformAcrossSurfaces` runs `duplicate_function_name` (core.duplicate_function_id), `foreign_origin_omitted` (core.foreign_origin_omitted) and `payload_borrow_interaction` (core.entry_ambiguous) across surfaces (`native_phase28_test.go:116-118`); passes. `admitProgram` is called before any tool in the command-file lanes (`session_admit.go`). Pins `TestNativeToolFailureIsOperational`, `TestOwnedBackendMutationIsMismatch` cover 3 and 4 (named in 28-06 summary; not re-run by me). |
| SC2 | Every exit code read from one named table; static test finds no duplicated literals in session, cmd, native; 65 and 69 unchanged | VERIFIED | `internal/compiler/exitcode/exitcode.go` is the leaf table. `TestExitCodeTable`, `TestNoDuplicatedExitLiterals`, `TestExitScanDetectorFires` (negative control) and `TestRepairExitMirrorMatchesTable` pass. `TestEmittedCExitLiteralsMatchTable` and `TestPhase273CapacityContractLiteralsAgree` pass (cgen). `TestAppRunExitCode` passes (65/69/124/125/128+n pinned). |
| SC3 | Function unreachable from `main` passes check, builds and runs with expected output; genuinely ambiguous entry exits 2; fully reachable programs keep byte-identical C | VERIFIED | `TestUnreachableFunctionProgramRuns` passes on interpreter and native for `multi_function_unreachable`, `multi_function_zero_call`, `unused_function`, `payload_borrow_interaction`. Ambiguity maps to a diagnostic (`entryRefusalDiagnostic`, `session_entry.go:49`) so exit 2. Byte identity: `TestPhase272AcceptedCorpusByteIdentity`, `TestPhase273ConformanceCorpusByteIdentity`, `TestPhase28EntryExitClassManifest` pass. |
| SC4 | `app verify` sum_to_n n=14 passes; n=15,172,173,1000 each `unsupported` with `native.event_capacity_exhausted`; 16 exhausted cases all named, within size limit, never success | VERIFIED | `TestAppVerifyExhaustedCases` subtests n=14/15/172/173/1000 pass; `TestAppVerifySixteenExhausted`, `TestAppVerifyReportBudget`, `TestAppVerifyReportOversize`, `TestAppVerifyOutcomeDigest`, `TestDecodeReplayCasesCaseIDBound` pass. CLI: `TestPhase273AppVerifyCapacityCLI` passes. |

**Score:** 4/4 truths verified, 0 behavior-unverified.

### Requirements Coverage

| Requirement | Source Plans | Status | Evidence |
|-------------|--------------|--------|----------|
| DX-17 | 28-01, 28-02, 28-06, 28-07 | SATISFIED | SC1, SC2; `TestAppVerifySourceRefusalExitsInvalid` (D-21, app verify source refusal exits 2); README exit table pinned by `TestExitCodeDocsTable` |
| DX-18 | 28-01, 28-04, 28-05 | SATISFIED | SC3; `resolveEntry` + `callgraph.PruneToReachable`, `main` precedence, entry ambiguity is a source refusal |
| EVD-13 | 28-03, 28-07 | SATISFIED | SC4; schema `lang.app-verification/2` outcome digest |

All three IDs appear in plan frontmatter and in REQUIREMENTS.md (marked Complete, traceability rows 84-86). No orphaned Phase 28 requirements.

### Code Review Findings, Independently Checked

- **WR-02 (CONFIRMED in source, classified as non-blocking warning).** `VerifyApplicationCasesFile` (`session_app_verify.go:~219-223`) returns the report-write error ahead of `verifyErr`. So a `*SourceRefusalError` is replaced by a write error and exits 3. This violates D-21 literally. It only triggers when the report write itself fails (IO failure or oversize). A refused source produces an empty report, so oversize cannot occur there. The class for the same source is still uniform whenever the report path is writable, and an unwritable report path is a legitimate tool/IO failure (exit 3). It does not defeat the SC1/SC4 behavior, which tests confirm. Recommend a gap-closure fix (`errors.Join(verifyErr, werr)`) but it does not block the phase.
- **WR-04 (CONFIRMED in source, classified as latent hardening, not a goal gap).** At `session_app_verify.go:276-283` a plain `entryErr` is stashed and the unpruned program used. Only source-kind cases consult it (line 319), so a cases file made only of `verifier_model_only` cases would not see it. However the plain-error branches of `resolveEntry` are `callgraph.prune_entry_unknown`, which cannot happen because the entry came from the same program, and unresolved-callee/ambiguity, which are mapped to diagnostics. After `Check` and `admitProgram` succeed no reachable input produces a plain error. I found no input that triggers it, so the fail-open path is currently unreachable. Fix by returning `entryErr` immediately; low risk.
- WR-01 (duplicate `main` reports in-degree roots and wrong hint), WR-03 (stale report file on refused/oversize run), IN-01, IN-02 were not asked about; they are diagnostic-quality or hardening items. WR-03 is a fail-open concern for report-reading automation and deserves triage. All six remain `open` in `28-REVIEW-DISPOSITION.md`.

### Anti-Patterns Found

None blocking. No TBD/FIXME/XXX debt markers were investigated beyond the focused tests; the AST exit-literal scan has a negative-control self-test.

### Human Verification Required

None.

### Gaps Summary

No blocking gaps. Phase goal achieved locally. Follow-ups to triage: WR-02 and WR-04 (cheap fixes), WR-03 (stale report), WR-01 (diagnostic text). Hosted macOS/Linux CI has not run for this revision; that remains outstanding and is owned by Phase 32 receipts.

### Commands Run (all passed)

- `go test ./internal/compiler/exitcode -run '^(TestExitCodeTable|TestNoDuplicatedExitLiterals|TestExitScanDetectorFires|TestRepairExitMirrorMatchesTable)$' -count=1`
- `go test ./internal/compiler/native -run '^(TestExitClassesUniformAcrossSurfaces|TestUnreachableFunctionProgramRuns)$' -count=1`
- `go test ./internal/compiler/session -run '^(TestAppVerifyExhaustedCases|TestAppVerifySixteenExhausted|TestAppVerifyReportBudget|TestAppVerifyReportOversize|TestAppVerifyOutcomeDigest|TestDecodeReplayCasesCaseIDBound)$' -count=1`
- `go test ./internal/compiler/session -run '^(TestPhase272AcceptedCorpusByteIdentity|TestPhase273ConformanceCorpusByteIdentity|TestPhase28EntryExitClassManifest|TestExitCodeDocsTable)$' -count=1`
- `go test ./cmd/schway -run '^(TestAppVerifySourceRefusalExitsInvalid|TestAppRunExitCode|TestPhase273AppVerifyCapacityCLI)$' -count=1`
- `go test ./internal/compiler/cgen -run '^(TestEmittedCExitLiteralsMatchTable|TestPhase273CapacityContractLiteralsAgree)$' -count=1`

---

_Verified: 2026-10-11_
_Verifier: Claude (gsd-verifier)_
