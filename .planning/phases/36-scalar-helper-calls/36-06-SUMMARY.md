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
  - Native borrowed-ByteView caller/helper composition for boundary and repeated fixture/FileBytes loops.
affects: ["36-07", "36-08", native-cgen, CALL-01, CALL-02]

actuals:
  tokens: 19794.75
  tasks: 3
  commits: 11
commits: 11
plan_head_before: b2bccdd07107a881a83bb1bc8174e30e3c9eea0c
plan_head_after: 37687a0c956b7e07121587c2cef77937487b132d

tech-stack:
  added: []
  patterns:
    - Isolated source overlays run actual interpreter/native consumers and require the named semantic assertion to kill each mutant.
    - Literal fixtures record independent arithmetic provenance separately from engine output.
    - Mixed borrowed-ByteView callers retain legacy ownership emitters while scalar helper arguments are evaluated before helper entry.
key-files:
  created:
    - examples/helper_calls/crc_step.schway
    - examples/helper_calls/importer.schway
    - examples/helper_calls/public.schway
    - testdata/phase36/helper-values.json
    - internal/compiler/session/session_phase36_byteview_test.go
    - internal/compiler/session/session_phase36_calls_test.go
    - internal/compiler/session/session_phase36_peer_mutation_test.go
    - .planning/phases/36-scalar-helper-calls/36-06-task1-red-evidence.json
    - .planning/phases/36-scalar-helper-calls/36-06-task2-red-evidence.json
    - .planning/phases/36-scalar-helper-calls/36-06-task3-red-evidence.json
  modified:
    - testdata/phase36/new-fixture-ledger.json
    - cmd/schway/main_phase36_test.go
    - internal/compiler/cgen/cgen_program.go

key-decisions:
  - "Independent expected CRC values come from Python integer arithmetic, never from Schway engine output."
  - "Checker source intent, raw core/origin contracts, causal attribution, and actual numeric engine outcomes remain separate proof owners."
  - "Native ByteView caller composition changes only checked scalar-call applicability and borrowed-view caller admission; the scalar call signature remains U64/Bool."

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
        ref: "cmd/schway/main_phase321_run_input_test.go#TestPhase321RunInputRequiredCLI"
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
    description: "Native caller byte_at bounds before scalar helper entry, plus repeated fixture and FileBytes ByteView helper loops."
    requirement: CALL-02
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_phase36_calls_test.go#TestPhase36NativeByteViewCallerArgumentBoundary"
        status: pass
      - kind: integration
        ref: "internal/compiler/session/session_phase36_byteview_test.go#TestPhase36NativeByteViewLoopHelpers"
        status: pass
      - kind: integration
        ref: "internal/compiler/session/session_phase36_byteview_test.go#TestPhase36NativeFileByteViewLoopHelpers"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase36_byteview_test.go#TestPhase36ByteViewArgumentOrderMutationControls"
        status: pass
    human_judgment: false
  - id: D5
    description: "The native argument-order oracle kills early-helper mutants at O0/O3/O3-LTO with clean and restored positives."
    requirement: CALL-02
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase36_byteview_test.go#TestPhase36ByteViewArgumentOrderMutationControls"
        status: pass
    human_judgment: false

duration: 38min
completed: 2026-10-15
status: complete
---

# Phase 36 Plan 36-06: Scalar Helper Calls Summary

**Repeated local/imported crc_step helpers and borrowed-ByteView caller, fixture, and FileBytes loops now have literal interpreter/native witnesses and O0/O3/LTO mutation controls.**

## Performance

- **Duration:** 38 minutes across the initial execution and the parent-owned debug repair.
- **Tasks:** 3/3 complete.
- **Files modified:** 13 plan source, fixture, test, and RED-evidence files.

## Accomplishments

- Added local `crc_step` and importing application examples. Independent Python integer arithmetic pins the single-step vectors `3988292384`, `1`, `152709948`, `3835865116`, and `2454158559`; four steps from `305419897` print `4046247459\n` from both examples.
- Exercised all four U64/Bool parameter/result pairs, nested computed arguments, result arithmetic, mutable assignment, and true/false branches with independent outputs `12\n` and `36\n`. Dynamic U64/Bool overflow and remainder defects preserve exact reason, empty stdout, exact stderr, and exit 65.
- Ran six successful local/imported native answer rows and six native defect rows across O0, O3, and O3-LTO. UBSan also ran the local positive and U64/Bool defect rows. Native defect traces prove entry into `add_one` or `remainder_to_bool` before the terminal comparison.
- Kept public conformance on its existing interpreter/native route: both return the independent execution value `1994146192`. The existing required-input CLI test remains intact; no U64 default or flag was added. Capture exhaustion and disabled capture remain separate from actual success/defect outcomes.
- Preserved the separate legacy ByteView callee bounds case through fixture bytes, file bytes, retained native execution, and native O0/O3/O3-LTO. The defect remains `byte index out of bounds`, exit 65, with no stdout; restored controls pass.
- Killed four interpreter mutants and twelve native mutants at their named value or terminal assertions. Every mutant had a fresh no-op positive and a restored positive; compiler errors and upstream refusal were not counted as kills.
- Resolved the mixed caller case in parent-owned debug work. The native boundary/loop consumers execute 21 O0/O3/O3-LTO lanes; OOB arguments exit 65 with zero helper calls, valid first/last reads return 90/100, and both fixture and real FileBytes loops return 928 with nine distinct helper returns.
- Killed three native argument-order mutants across O0/O3/O3-LTO at the exact zero-entry assertion, with fresh no-op and restored positives. Independent executionpeer validation accepts successful fixture and FileBytes traces.

## Native ByteView Caller Resolution

**Resolution dated 2026-10-15:** The original interpreter-only boundary evidence did not establish native caller ordering. The parent routed the refusal through GSD debug; RCA found that the scalar-call applicability predicate treated any legacy non-scalar call in the graph as a veto, disabling dynamic support for a borrowed-ByteView caller.

Commit `5d874384` selects checked scalar calls independently of legacy ownership calls and admits the existing ByteView parameter place in borrowed-view scalar caller bodies. Scalar helper signatures remain U64/Bool; legacy ownership emitters and release pairing remain in place. The earlier exact refusal was reproduced by overlaying only the pre-fix `cgen_program.go` and disappeared on reapply.

The historical RED root was `TestPhase36NativeByteViewCallerArgumentRefusal` in `session_phase36_calls_test.go`. With the Phase29 fixture binding manifest and `identity(byte_at(bytes, byte_len(bytes)))`, `session.BuildApplication` stopped before producing an executable at `internal/compiler/cgen/cgen_program.go:2934` with the exact error `operation "s1:examples.phase29.adler32_fixture:fn:adler32:op:4": scalar call emission requires dynamic invocation support`. A pre-fix `cgen_program.go` overlay reproduced the same refusal; the interpreter-only evidence did not satisfy the native consumer requirement. This was a compiler blocker, not human UAT. The replacement `TestPhase36NativeByteViewCallerArgumentBoundary` executes the native consumer after the 2026-10-15 repair.

`TestPhase36NativeByteViewCallerArgumentBoundary` runs first-byte, last-byte, at-length, past-length, and empty-view cases at O0/O3/O3-LTO. Valid reads return `90` and `100` with one helper entry. All three OOB cases exit 65 with exact `byte index out of bounds` stderr and zero helper entries. `TestPhase36NativeByteViewLoopHelpers` and `TestPhase36NativeFileByteViewLoopHelpers` each run their successful repeated helper loop at all three lanes and return `928` with nine distinct helper returns; successful evidence passes executionpeer validation.

`TestPhase36ByteViewArgumentOrderMutationControls` moves helper entry before the byte read in source overlays. Across O0/O3/O3-LTO, every mutant fails at `native helper invocation events=1 want 0`, while fresh no-op and restored positives pass. The cgen repair, consumer roots, and mutation controls are documented in [the resolved debug record](../../debug/resolved/phase36-byteview-helper.md), commits `5d874384` and `37687a0c`.

## Task Commits

1. **Task 1 RED** — `955f2daa` (`test(36-06): pin crc_step semantic red evidence`)
2. **Task 1 GREEN** — `294abbf2` (`feat(36-06): add repeated crc_step helper programs`)
3. **Task 2 RED** — `2a173c0b` (`test(36-06): pin public helper conformance red evidence`)
4. **Task 2 partial GREEN** — `e6d57b96` (`feat(36-06): pin public helper and byte-view bounds routes`)
5. **Task 3 RED** — `4e2074fb` (`test(36-06): add failing engine call mutation control test`)
6. **Task 2 blocker evidence** — `126ba940` (`test(36-06): record native ByteView caller emission refusal`)
7. **Task 3 GREEN** — `22bc7db3` (`test(36-06): kill helper call engine mutants`)
8. **Mutation harness cleanup** — `47aae994` (`refactor(36-06): run native lane child without skips`)
9. **Initial halted summary** — `57060d8e` (superseded by this completion summary)
10. **Native ByteView composition repair** — `5d874384` (`fix(36): compose native ByteView callers with scalar helpers`)
11. **Debug resolution record** — `37687a0c` (`docs(debug): resolve phase36-byteview-helper session`)

The final summary/STATE/ROADMAP metadata commit follows this fresh verification.

## Verification

- Task 1: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase36HelperAnswersAndDefects|TestPhase36HelperOptimizationMatrix|TestPhase36HelperUBSan)$' -count=1 -v -timeout=600s` — PASS.
- Task 2: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway ./internal/compiler/session -run '^(TestPhase36PublicHelperConformance|TestPhase321RunInputRequiredCLI|TestPhase36ByteViewCallerArgumentBoundary|TestPhase36NativeByteViewCallerArgumentBoundary|TestPhase36NativeByteViewLoopHelpers|TestPhase36NativeFileByteViewLoopHelpers|TestPhase36ByteViewArgumentOrderMutationControls|TestPhase36LegacyByteViewCalleeBounds|TestPhase36NativeCallCaptureOutcomeSeparation|TestPhase29InterpreterFixtureVectors|TestPhase29InterpreterDefectWithLiveOwner|TestPhase29NativeFixtureVectors)$' -count=1 -v -timeout=600s` — PASS; all 21 native consumer lanes and all three ordering mutants ran.
- Task 3: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase36EngineCallMutationControls|TestPhase36HelperAnswersAndDefects)$' -count=1 -v -timeout=600s` — PASS; 4 interpreter and 12 native mutants killed, with O0/O3/O3-LTO positives and restorations.
- RED classifiers for Task 1, Task 2, and Task 3 each returned `RED_EVIDENCE_OK`. Each target failed on its intended numeric, route-value, or mutation-control assertion before GREEN.
- Parent-manager fresh gates also passed for the four native consumer/control roots, adjacent helper/capture/legacy roots, independent peer/capacity roots, and both frozen C baseline roots after the three ledger hash/schema rows were repaired. This manager receipt is recorded in the debug report; the three task selectors above were rerun locally for this completion.
- No full suite, broad cross-program matrix, Phase37 CLI, full CRC implementation, or guide was run or added.

## Deviations from Plan

- The native mixed-call applicability and borrowed-ByteView admission repair was executed as a separate parent-owned GSD debug session, with one production revision and a narrow fixture-ledger source-hash/schema correction. Those commits are included above; the final native semantics were rerun through the exact Task 2 selector.

**Total deviations:** 1 separately routed debug correction and 1 required ledger metadata repair. No dependencies, backend, schema, file CLI, or phase37 scope were added.

## Next Phase Readiness

Plan36-06 is complete. Phase36 remains open: Plans36-07 and 36-08 are still pending and no Phase37 work was started.

## Self-Check: PASSED

- Summary, test, and debug commits are verified in the current history.
- The three focused selectors passed with current HEAD; the new native caller, fixture-loop, FileBytes-loop, and mutation roots executed rather than being skipped.
- The native caller refusal remains recorded above as historical pre-fix evidence, followed by its dated resolution and fresh semantic evidence.
- No TODO/FIXME/placeholder stubs or skipped roots were found in the owned source/test changes.

---
*Phase: 36-scalar-helper-calls*
*Status: plan complete; Phase36 remains in progress*
