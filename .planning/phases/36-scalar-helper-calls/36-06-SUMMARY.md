---
phase: 36-scalar-helper-calls
plan: "06"
subsystem: compiler-runtime
tags: [scalar-calls, crc-step, native-cgen, mutation-controls, byteview]

requires:
  - phase: "36-04"
    provides: Independent causal attribution for dynamic helper calls.
  - phase: "36-05"
    provides: Native scalar calls and bounded invocation evidence.
provides:
  - Local and imported repeated crc_step source witnesses with independent literal answers.
  - Interpreter and native O0/O3/O3-LTO answer and defect controls for scalar helper calls.
  - Public helper route, legacy ByteView callee bounds, and engine mutation evidence.
  - Exact native ByteView caller-helper emission refusal retained as an unresolved gate.
affects: ["36-07", "36-08", native-cgen, CALL-01, CALL-02]

actuals:
  tokens: 18052.25
  tasks: 2
  commits: 8
commits: 8
plan_head_before: b2bccdd07107a881a83bb1bc8174e30e3c9eea0c
plan_head_after: 47aae9946b6be7667451e63fccacd16595820013

tech-stack:
  added: []
  patterns:
    - Isolated source overlays run actual interpreter/native consumers and require the named semantic assertion to kill each mutant.
    - Literal fixtures record independent arithmetic provenance separately from engine output.
key-files:
  created:
    - examples/helper_calls/crc_step.schway
    - examples/helper_calls/importer.schway
    - examples/helper_calls/public.schway
    - testdata/phase36/helper-values.json
    - internal/compiler/session/session_phase36_calls_test.go
    - internal/compiler/session/session_phase36_peer_mutation_test.go
    - .planning/phases/36-scalar-helper-calls/36-06-task1-red-evidence.json
    - .planning/phases/36-scalar-helper-calls/36-06-task2-red-evidence.json
    - .planning/phases/36-scalar-helper-calls/36-06-task3-red-evidence.json
  modified:
    - testdata/phase36/new-fixture-ledger.json
    - cmd/schway/main_phase36_test.go

key-decisions:
  - "Independent expected CRC values come from Python integer arithmetic, never from Schway engine output."
  - "Checker source intent, raw core/origin contracts, causal attribution, and actual numeric engine outcomes remain separate proof owners."
  - "Interpreter caller-before-entry evidence does not discharge the native ByteView caller requirement."

patterns-established:
  - "Mutant runs require a fresh no-op positive, the named semantic failure, and a restored positive."
  - "Native mutation children execute one actual O0, O3, or O3-LTO case and verify compiler flags for LTO."

requirements-completed: []
coverage:
  - id: D1
    description: "Local and imported reflected-bit helpers repeat four steps and match independent literal values."
    requirement: CALL-01
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_phase36_calls_test.go#TestPhase36HelperAnswersAndDefects"
        status: pass
      - kind: integration
        ref: "internal/compiler/session/session_phase36_calls_test.go#TestPhase36HelperOptimizationMatrix"
        status: pass
    human_judgment: false
  - id: D2
    description: "Public helper conformance, explicit-input refusal, and optional capture outcome separation remain pinned."
    verification:
      - kind: integration
        ref: "cmd/schway/main_phase36_test.go#TestPhase36PublicHelperConformance"
        status: pass
      - kind: unit
        ref: "cmd/schway/main_phase35_test.go#TestPhase321RunInputRequiredCLI"
        status: pass
    human_judgment: false
  - id: D3
    description: "Legacy ByteView callee bounds remain separate and run on interpreter, retained native app, O0, O3, and O3-LTO."
    requirement: CALL-02
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_phase36_calls_test.go#TestPhase36LegacyByteViewCalleeBounds"
        status: pass
    human_judgment: false
  - id: D4
    description: "Native caller byte_at bounds before scalar helper entry."
    requirement: CALL-02
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_phase36_calls_test.go#TestPhase36NativeByteViewCallerArgumentRefusal"
        status: fail
    human_judgment: true
    rationale: "The native build is refused before a process exists, so automation cannot establish native caller evaluation order. The parent must route the cgen repair through GSD debug and rerun this consumer case."

duration: 23min
completed: 2026-10-15
status: halted
---

# Phase 36 Plan 36-06: Scalar Helper Calls Summary

**Repeated local and imported crc_step calls now have independent interpreter/native witnesses and mutation controls; native ByteView caller ordering remains blocked by a cgen refusal.**

## Performance

- **Duration:** 23 minutes, from the first Task 1 RED commit at 2026-10-15T03:47:03Z through the final verified test commit at 2026-10-15T04:09:36Z.
- **Tasks:** 2 complete; Task 2's native caller-ordering obligation is blocked.
- **Files modified:** 11 plan-owned source, fixture, test, and RED-evidence files.

## Accomplishments

- Added local `crc_step` and importing application examples. Independent Python integer arithmetic pins the single-step vectors `3988292384`, `1`, `152709948`, `3835865116`, and `2454158559`; four steps from `305419897` print `4046247459\n` from both examples.
- Exercised all four U64/Bool parameter/result pairs, nested computed arguments, result arithmetic, mutable assignment, and true/false branches with independent outputs `12\n` and `36\n`. Dynamic U64/Bool overflow and remainder defects preserve exact reason, empty stdout, exact stderr, and exit 65.
- Ran six successful local/imported native answer rows and six native defect rows across O0, O3, and O3-LTO. UBSan also ran the local positive and U64/Bool defect rows. Native defect traces prove entry into `add_one` or `remainder_to_bool` before the terminal comparison.
- Kept public conformance on its existing interpreter/native route: both return the independent execution value `1994146192`. The existing required-input CLI test remains intact; no U64 default or flag was added. Capture exhaustion and disabled capture remain separate from actual success/defect outcomes.
- Preserved the separate legacy ByteView callee bounds case through fixture bytes, file bytes, retained native execution, and native O0/O3/O3-LTO. The defect remains `byte index out of bounds`, exit 65, with no stdout; restored controls pass.
- Killed four interpreter mutants and twelve native mutants at their named value or terminal assertions. Every mutant had a fresh no-op positive and a restored positive; compiler errors and upstream refusal were not counted as kills.

## Native ByteView Caller Blocker

**Clarification dated 2026-10-15:** The caller fixture `identity(byte_at(bytes, byte_len(bytes)))` still proves interpreter evaluation order: the byte-index defect occurs with one live owner and zero `identity` calls. That interpreter result does not establish native caller ordering.

`TestPhase36NativeByteViewCallerArgumentRefusal` attempts the same caller through `session.BuildApplication` with the Phase29 fixture binding manifest. The current native build stops before producing an executable with:

```text
operation "s1:examples.phase29.adler32_fixture:fn:adler32:op:4": scalar call emission requires dynamic invocation support
```

The refusal originates at `internal/compiler/cgen/cgen_program.go:2934`. No native caller-side OOB or helper-entry outcome was observed, so CALL-02 remains incomplete here. The parent will route the cgen repair through GSD debug; after that repair, rerun the exact native consumer at O0/O3/O3-LTO and require zero helper calls on the OOB case plus a valid-index control that enters the helper. This plan is intentionally **halted**, not complete.

## Task Commits

1. **Task 1 RED** — `955f2daa` (`test(36-06): pin crc_step semantic red evidence`)
2. **Task 1 GREEN** — `294abbf2` (`feat(36-06): add repeated crc_step helper programs`)
3. **Task 2 RED** — `2a173c0b` (`test(36-06): pin public helper conformance red evidence`)
4. **Task 2 partial GREEN** — `e6d57b96` (`feat(36-06): pin public helper and byte-view bounds routes`)
5. **Task 3 RED** — `4e2074fb` (`test(36-06): add failing engine call mutation control test`)
6. **Task 2 blocker evidence** — `126ba940` (`test(36-06): record native ByteView caller emission refusal`)
7. **Task 3 GREEN** — `22bc7db3` (`test(36-06): kill helper call engine mutants`)
8. **Mutation harness cleanup** — `47aae994` (`refactor(36-06): run native lane child without skips`)

The metadata completion commit and STATE/ROADMAP advancement are intentionally pending the parent-owned cgen repair and a complete rerun.

## Verification

- Task 1: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase36HelperAnswersAndDefects|TestPhase36HelperOptimizationMatrix|TestPhase36HelperUBSan)$' -count=1 -v -timeout=600s` — PASS.
- Task 2: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway ./internal/compiler/session -run '^(TestPhase36PublicHelperConformance|TestPhase321RunInputRequiredCLI|TestPhase36ByteViewCallerArgumentBoundary|TestPhase36NativeByteViewCallerArgumentRefusal|TestPhase36LegacyByteViewCalleeBounds|TestPhase36NativeCallCaptureOutcomeSeparation|TestPhase29InterpreterFixtureVectors|TestPhase29InterpreterDefectWithLiveOwner|TestPhase29NativeFixtureVectors)$' -count=1 -v -timeout=600s` — PASS, with the refusal explicitly logged; this is not native semantic coverage.
- Task 3: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase36EngineCallMutationControls|TestPhase36HelperAnswersAndDefects)$' -count=1 -v -timeout=600s` — PASS; 4 interpreter and 12 native mutants killed, with O0/O3/O3-LTO positives and restorations.
- RED classifiers for Task 1, Task 2, and Task 3 each returned `RED_EVIDENCE_OK`. Each target failed on its intended numeric, route-value, or mutation-control assertion before GREEN.
- No full suite, broad cross-program matrix, Phase37 CLI, full CRC implementation, or guide was run or added.

## Deviations from Plan

- Added a focused refusal test so the native ByteView caller gap remains visible as a reproducible compiler boundary instead of being inferred from interpreter evidence. No production cgen change was made; the parent owns its GSD debug route.
- Task 2 remains incomplete until native caller OOB-before-entry behavior runs successfully. Its public route, interpreter caller case, native legacy callee case, and capture controls pass.

**Total deviations:** 1 scoped blocker-evidence addition; 1 unresolved plan obligation.

## Deferred Issues

- **Native caller-side ByteView OOB before scalar helper entry:** reproduce with the inline Phase29 fixture replacing `adler32` by the function in `TestPhase36NativeByteViewCallerArgumentRefusal` (`internal/compiler/session/session_phase36_calls_test.go:655`). The exact current refusal is `operation "s1:examples.phase29.adler32_fixture:fn:adler32:op:4": scalar call emission requires dynamic invocation support`. Native execution did not begin. Parent-owned next action: route cgen through GSD debug, then implement and verify the native consumer at O0/O3/O3-LTO.

## Self-Check: PASSED

- This summary file exists at the canonical Phase36 Plan36-06 path.
- All eight recorded task/evidence commits are ancestors of HEAD.
- The owned source/test scan found no TODO/FIXME/placeholder stubs or skipped tests.
- The native caller refusal is recorded as a blocker; no completion state is claimed.

---
*Phase: 36-scalar-helper-calls*
*Status: halted on native ByteView caller support*
