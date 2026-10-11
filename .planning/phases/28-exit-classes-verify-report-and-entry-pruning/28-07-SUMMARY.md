---
phase: 28-exit-classes-verify-report-and-entry-pruning
plan: 07
subsystem: app-verify-exit-classes
tags: [DX-17, EVD-13, exit-classes, docs]
requires:
  - phase: 28-06
    provides: admitProgram single admission sequence
provides:
  - SourceRefusalError and app verify exit 2 for source refusals
  - README two-domain exit-code table pinned to the constants
  - Phase 28 living-docs transition amendments
affects: []
status: complete
commits: 3
plan_head_before: 4b565c260a5480aad0491a1422d3d80b4f34fc83
plan_head_after: a5899fe1334866635ef2f02beba338958015961c
actuals:
  tokens: 14000
  tasks: 3
  commits: 3
key-files:
  created:
    - internal/compiler/session/session_phase28_exitdocs_test.go
  modified:
    - internal/compiler/session/session_app_verify.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_phase271_admission_test.go
    - cmd/schway/main.go
    - cmd/schway/main_exitcode_test.go
    - testdata/phase16/public-emitter-consumers.json
    - README.md
    - .planning/PRODUCT-ROADMAP.md
    - .planning/LANGUAGE-MATURITY.md
key-decisions:
  - "VerifyApplicationCases runs Check, admitProgram, resolveEntry and returns *SourceRefusalError; the CLI maps it to exitcode.Invalid"
  - "admitPublishedApplication, applicationAdmission and applicationRefusalLayer deleted; admitProgram is the only admission sequence"
---

# Phase 28 Plan 07: App verify exit 2 and published exit table Summary

`schway app verify` now exits 2 with `source refused: <code>` for every source refusal (checker, core, origin, path oracle, entry), still writes the fail report, and keeps exit 3 for capacity-exhausted cases. The README publishes the exit-code table, pinned to the constants by a test.

## Tasks

| Task | Name | Commit |
| ---- | ---- | ------ |
| 1 (tracer) | app verify source refusal exits 2, old admission gate deleted | 9b95892 |
| 2 | README Exit codes tables and TestExitCodeDocsTable | a711be0 |
| 3 | PRODUCT-ROADMAP and LANGUAGE-MATURITY amendments | a5899fe |

## Observed values

- `TestAppVerifySourceRefusalExitsInvalid`: corevalidate (`core.duplicate_function_id`) and entry_ambiguous (`core.entry_ambiguous`) exit 2 with stderr `schway app verify: source refused: <code>` and a written report; capacity_stays_3 exits 3 with a report.
- `TestExitCodeDocsTable` parses 11 rows, matches all constants, and its negative controls (wrong value, renamed row) are reported.
- The `EmitNative` call-site pin in `testdata/phase16/public-emitter-consumers.json` shifted 3088 -> 3041 after the deletion in session.go and was re-keyed in the same commit.

## Deviations from Plan

1. The phase271 verify rows live in `TestApplicationRoutesApplySameAdmissionLayers` and `TestPhase271VerifyFileRouteWritesOnlyFailReport`; both were rewritten to assert the typed error and its code and pass. `requireVerifyRefusal` now takes the expected code instead of error text.
2. The README Exit codes section was added at the end of the file (after Start here), per the plan.
3. The example in the amendments is described as `app run` with input 7 printing 7 (the plan said "prints 7" without naming the input).

No phase28ClassChanges entries were needed: no program's exit class moved in this plan beyond the app verify command, which the class manifest does not cover.

## Verification

- `go test ./internal/compiler/session -run 'Inventory|Groundedness|Reconciliation|ValidationLifecycle|TestPhase28|TestPhase272AcceptedCorpusByteIdentity|TestPhase273ConformanceCorpusByteIdentity|TestCLICheckAdmissionDivergenceIsExactlyKnown|TestPhase16|TestApplicationRoutesApplySameAdmissionLayers|TestSelfDescribingDocsGuardIsNotInert|TestAppVerify' -count=1` passes.
- `go test ./internal/compiler/session -run '^TestExitCodeDocsTable$' -count=1` passes.
- `go test ./cmd/schway -count=1` and `go test ./internal/compiler/exitcode -count=1` pass.
- `go test ./cmd/schway -run '^(TestAppVerifySourceRefusalExitsInvalid|TestAppRunExitCode|TestPhase273AppVerifyCapacityCLI)$' -count=1` passes.
- `go test ./internal/compiler/native -run '^(TestExitClassesUniformAcrossSurfaces|TestUnreachableFunctionProgramRuns)$' -count=1` passes.
- Hosted macOS and Linux CI has not run.

## Known Stubs

None.

## Threat Flags

None.

## Self-Check: PASSED

Commits 9b95892, a711be0 and a5899fe are ancestors of HEAD; session_phase28_exitdocs_test.go exists.
