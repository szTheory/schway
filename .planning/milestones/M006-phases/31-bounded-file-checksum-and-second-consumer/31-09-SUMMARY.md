---
phase: 31-bounded-file-checksum-and-second-consumer
plan: 09
subsystem: docs/closure
tags: [closure, readme, registers, validation, transition-check]
requires: [31-03, 31-04, 31-05, 31-06, 31-07, 31-08]
provides:
  - examples/checksum/README.md documenting checksum and adler_low
  - exit-65 outcome names in README.md and the exitcode doc comment
  - LANGUAGE-MATURITY and PRODUCT-ROADMAP Phase 31 rows and amendments
  - 31-TRANSITION-CHECK.md
  - 31-VALIDATION.md finalized (status validated)
affects: [31-VERIFICATION, Phase 32]
key-files:
  created:
    - .planning/phases/31-bounded-file-checksum-and-second-consumer/31-TRANSITION-CHECK.md
  modified:
    - examples/checksum/README.md
    - README.md
    - internal/compiler/exitcode/exitcode.go
    - .planning/LANGUAGE-MATURITY.md
    - .planning/PRODUCT-ROADMAP.md
    - .planning/phases/31-bounded-file-checksum-and-second-consumer/31-VALIDATION.md
decisions:
  - "Docs name tests by identifier only; the 65 exit row names FileTooLarge, FileUnavailable, AllocFailed with no constant change"
status: complete
commits: 3
plan_head_before: 5344d769a8812d058bba756ac34b0264f270a7b0
plan_head_after: ef2b6b0a86bc62ff6b72460cc8e8dab9fe66e8c6
actuals:
  tokens: 14000
  tasks: 3
  commits: 3
---

# Phase 31 Plan 09: Closure Summary

The checksum README now tells a developer how to build and run `checksum` and `adler_low`, every pinned answer and error line is tied to a named Phase 31 test, the registers and transition check are written, and 31-VALIDATION.md is finalized as validated after a green local phase-close gate.

## Commits
- 0f71048e: docs(31-09): checksum README, outcome names on exit 65 (D-08, D-18)
- b9a8e84d: docs(31-09): living registers and the Phase 31 transition check
- ef2b6b0a: docs(31-09): finalize Phase 31 validation

## Verification (all local, focused, run in this plan)
- Exit docs: the exitcode package and TestExitCodeDocsTable, TestPhase28EntryExitClassManifest, TestVerificationGroundedness pass. The exitcode.go diff is comment-only.
- Registers: TestLanguageMaturityCountsAreCurrent, TestVerificationGroundedness and TestVerificationGroundednessScopeByIllocutionaryRole pass.
- Phase gate: the Phase31 focused set passes in check, corevalidate, originvalidate, pathoracle, cgen, native, session, testsupport and cmd/schway with no skips in the lanes run. The session meta set, TestSourceNeverSpawnsUnboundedProcesses, TestSingleLoanLivenessLaw and the exitcode package pass.
- Every test named in 31-VALIDATION.md exists (located by grep); TestPhase20ValidationLifecycle passes with status validated.

## Deviations from Plan

**1. [Plan-spec] No Phase31 tests in `internal/compiler/interp`.** The plan's first gate command lists the interp package and treats "[no tests to run]" as a failure. Plan 31-02 deliberately put the interpreter-route test in session (TestPhase31ChecksumInterpreter), and interp itself holds no Phase31-named test. The test passes; the package emits "[no tests to run]". I did not add a placeholder test. The orchestrator may want to drop interp from that command.

**2. Extra VALIDATION edits.** The status legend line contained the literal pending marker that the plan's acceptance grep counts, so it was reworded. The "names proposed in RESEARCH" sentence was updated to say the names now exist.

**3. Roadmap non-claim wording.** The Phase 31 roadmap amendment avoids repeating the phrase "separate compilation" (the plan's grep must not increase); the boundary is stated as B1/DX-06 staying open.

## Known Stubs
None.

## Threat Flags
None.

## Self-Check: PASSED
Created and modified files exist; commits 0f71048e, b9a8e84d and ef2b6b0a are in HEAD. Hosted CI has not run; no document claims it.
