---
phase: 29-byte-views-and-the-adler-32-library
plan: 12
subsystem: spike-evidence
tags: [spike-008, byte-views, cap, native, interpreter, timing, hosted-ci]
requires: [29-00, 29-09, 29-11]
provides: [spike 008 harness, hosted strict-budget evidence, VALIDATED verdict, cap kept at 65,536]
key-files:
  created:
    - .planning/spikes/008-byteview-cap/README.md
    - internal/compiler/native/native_phase29_spike008_test.go
    - internal/compiler/native/native_phase29_race_test.go
  modified:
    - .github/workflows/ci.yml
    - .planning/spikes/MANIFEST.md
    - internal/compiler/session/qlt01_registry.json
key-decisions:
  - "Cap kept at 65,536 (D-12/D-13): all lanes pass the strict budget on ubuntu-latest and macos-latest, plain and -race, plus local"
  - "Spike 008 QLT-01 row stays a waived row (001-005 pattern), wording updated to the VALIDATED verdict"
requirements-completed: [APP-10, BYTE-02]
status: complete
actuals:
  tokens: 14000
  tasks: 3
  commits: 6
plan_head_before: b884c250ea29c3a0b39857762aa8d5c94a581b2b
plan_head_after: f2018ee
---

# Phase 29 Plan 12: Spike 008 byte view cap Summary

A three-mode (default, record, strict) harness measures the 65,536-byte Adler-32 fixture through the real fixture path on every native lane (O0, O3, O3-LTO) and the interpreter; hosted strict run 38186068052 at 906d648 passed on both hosts, so the verdict is VALIDATED and the cap is kept at 65,536.

## Tasks
1. Harness, ci.yml dispatch input plus `spike 008 strict budget` step, local record (1327cf1).
2. Checkpoint approved: orchestrator pushed and dispatched.
3. Hosted evidence and decision recorded (f2018ee): README Hosted runs and Decision sections, MANIFEST 008 row VALIDATED, qlt01 waiver text updated.

## Hosted evidence
Run 38186068052 (workflow_dispatch, spike008_strict=true), head 906d648: checks (ubuntu-latest) and checks (macos-latest) succeeded on go test, go test -race and `spike 008 strict budget`; both aggregate jobs succeeded. Worst native max 8.0 ms (3 s budget); worst interpreter max 376.5 ms (12 s budget). No lane OVER BUDGET. Full lane tables are in the spike README.

## Deviations from Plan

### Auto-fixed Issues (orchestrator-dispatched, surfaced by hosted CI)
First dispatch run 38185331902 at 1327cf1 failed before the strict step on five repo convention tests.

**1. [Rule 3 - Blocking] TestSingleLoanLivenessLaw** - 0364898: reach the D-22 cycle refusal without a fifth loanLivenessFixpoint caller.
**2. [Rule 3 - Blocking] TestSourceNeverSpawnsUnboundedProcesses / skip citations** - 532f0ba: bound Phase 29 test spawns, cite every skip; replaced native_phase29_race_on/off_test.go build-tag files with native_phase29_race_test.go reading -race via debug.ReadBuildInfo.
**3. [Rule 3 - Blocking] TestNoUndeclaredCheckPeerDivergenceAcrossCorpus** - 614841f: pathoracle cyclic-CFG refusal treated as its declared scope limit in the peer gate.
**4. [Rule 3 - Blocking] TestQLT01RegistryCoversAllFiveSpikes** - 906d648: added spike 008 waived row (then PARTIAL wording); updated in f2018ee to match VALIDATED.

TestNoSuppressionOutlivesItsWitness was covered by the same fixes. All five convention tests and the session meta-tests pass locally.

## Known Stubs
None.

## Notes
29-11-SUMMARY and the Phase 29 closure documents say the cap was not yet measured and Stage B had not run; that is now superseded by this plan (historical record left unchanged).

## Self-Check: PASSED
