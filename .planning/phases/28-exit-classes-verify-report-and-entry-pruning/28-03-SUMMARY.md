---
phase: 28-exit-classes-verify-report-and-entry-pruning
plan: 03
subsystem: session
tags: [app-verify, report-schema, outcome-digest, event-capacity, size-budget]
requires:
  - phase: 28-01
    provides: exit-pin inventory and class manifest baseline
provides:
  - ReplayOutcomeDigest and ReplayCaseReport.OutcomeDigest replacing embedded evidence for exhausted cases (D-13)
  - lang.app-verification/2 report schema (D-15)
  - MaxReplayCaseIDBytes = 128 with printable-ASCII check (D-16)
  - ErrApplicationVerificationReportOversize named runtime defense (D-16)
  - static 16-case report budget test and digest preimage proof
affects: [28-04, 28-05, 28-06, 28-07]
tech-stack:
  added: []
  patterns: [fixed-size digest in place of unbounded evidence, reflection-based worst-case size budget]
key-files:
  created:
    - internal/compiler/session/session_phase28_appverify_test.go
  modified:
    - internal/compiler/session/session_app_verify.go
    - internal/compiler/session/export_test.go
    - internal/compiler/session/session_app_verify_test.go
    - internal/compiler/session/session_phase273_appverify_test.go
    - cmd/schway/main_test.go
    - examples/phase22/README.md
key-decisions:
  - "D-15 schema bump to lang.app-verification/2 is costly (published report schema); pre-approved in CONTEXT, recorded, no checkpoint"
  - "Budget worst case uses closed-vocabulary values for enum-like string fields (status, kind, scope, engine, level)"
requirements-completed: [EVD-13]
status: complete
commits: 3
plan_head_before: 8009a746fac2183a6ab923932729e145445e64d9
plan_head_after: 7b7799c
actuals:
  tokens: 14000
  tasks: 3
  commits: 3
duration: 20min
completed: 2026-10-11
---

# Phase 28 Plan 03: Verify Report Digest Summary

`app verify` now reports every event-capacity-exhausted case as `unsupported` with the named `native.event_capacity_exhausted` reason at any n, using a fixed-size interpreter outcome digest under schema `lang.app-verification/2` instead of embedded evidence.

## What changed

- Exhausted cases (both-level and one-level) carry `outcome_digest {engine, outcome_kind, outcome_value, event_count, evidence_sha256, evidence_bytes, exhausted_level}` and no `engines` map. Statuses and diagnostics are byte-identical to before (D-14). Pass and fail cases with complete evidence keep their full engines map.
- `MaxReplayCaseIDBytes = 128`; `DecodeReplayCases` refuses longer IDs and any byte outside 0x20..0x7E. Existing committed cases file (`examples/phase22/identity.cases.json`) is far under the bound (IDs under 70 bytes).
- `writeApplicationVerificationReport` returns `ErrApplicationVerificationReportOversize` (size and limit in the text) before `os.CreateTemp`, so no file is written.
- The 27.3 tests were rewritten (not deleted) to assert the digest via a shared `phase273RequireDigest` helper.

## Measured values

- Sixteen worst-case exhausted cases (reflection-filled): 19,941 bytes against a 32,768 budget.
- Real sixteen-exhausted-case run (n = 15..29 and 1000, 128-byte IDs): 13,012 byte report, limit 65,536.
- Consumer re-grep: `lang.app-verification` appears only in `examples/phase22/README.md` under scripts/, .github/ and examples/; the `"engines"` key appears nowhere in them (D-15).

## Tests (all existing and passing)

- `go test ./internal/compiler/session -run '^(TestAppVerifyExhaustedCases|TestAppVerifySixteenExhausted|TestAppVerifyReportBudget|TestAppVerifyReportOversize|TestAppVerifyOutcomeDigest|TestDecodeReplayCasesCaseIDBound)$' -count=1`
- `go test ./internal/compiler/session -run '^TestPhase273AppVerify.*$' -count=1`
- `go test ./cmd/schway -run '^TestPhase273AppVerifyCapacityCLI$' -count=1` (includes subtest `n=1000`: exit 3, schema /2, unsupported with the named prefix)
- Meta-tests `Inventory|Groundedness|Reconciliation|ValidationLifecycle|TestPhase28EntryExitClassManifest` and `go test ./internal/compiler/exitcode` pass clean.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Budget worst case over budget with all-128-byte strings**
- **Found during:** Task 2
- **Issue:** The plan's "any other string 128 bytes" rule yielded 33,034 bytes, over the 32,768 budget, because closed-vocabulary fields (status, kind, evidence scope, engine, outcome kind, exhausted level) were inflated to 128 bytes though the producer can only emit fixed values.
- **Fix:** The filler uses the longest real constant for those closed-vocabulary fields; all genuinely free fields (IDs, diagnostics, digests, numeric strings, ModeledOutcome strings) stay at worst case. Result 19,941 bytes. The filler still fails if a struct gains a field with no rule.
- **Files modified:** internal/compiler/session/session_phase28_appverify_test.go
- **Commit:** be503bf

**2. [Scope] Task 3 n=14 subtest uses the file entry point and reuses the common helper**
- The n=14 subtest passes through `VerifyApplicationCasesFile`, so the written report is also exercised for the passing case.

## Known Stubs

None.

## Threat Flags

None.

## Self-Check: PASSED

Commits 89ef003, be503bf and 7b7799c exist; created and modified files verified present.
