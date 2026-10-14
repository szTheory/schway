---
phase: 35-bitwise-operators-literal-shifts-and-hex-literals
plan: "06"
subsystem: compiler-cgen
tags: [Go, C17, Clang, U64, bitwise, literal-shifts]

requires:
  - phase: 35-04
    provides: Independently validated scalar transfers and the narrow unused-Byte bridge
provides:
  - Checked U64 literal shifts lower to unsigned C and produce the pinned native result
  - A positive two-site C shift census with dynamic, width-sized, signed-operand, and removal controls
  - Event-free projection and preserved unused-Byte signature behavior at native emission
affects: [35-07, cgen, native-emission]

actuals:
  tokens: 13397
  tasks: 2
  commits: 4
commits: 4
plan_head_before: f6f1aa08b87f4207364ccf0fff58699eb92a186f
plan_head_after: a5ccff74043608a262abdedc6c2d61d8402b4493

tech-stack:
  added: []
  patterns:
    - Literal shifts emit through an unsigned widened operand and a final U64 cast.
    - Generated shift evidence is counted only inside scalar function regions against an independently authored positive manifest.
    - Pure scalar operation projection preserves existing return events without introducing operation events.

key-files:
  created:
    - internal/compiler/cgen/cgen_phase35_test.go
    - testdata/phase35/shift-sites.json
  modified:
    - internal/compiler/cgen/cgen_program.go

key-decisions:
  - "Use the installed C17 runner at -O0 for the native tracer and keep the expected answer literal in the test."
  - "Require exactly two generated scalar shift sites and fail closed on dynamic counts, count 64, signed operands, or a missing site."
  - "Keep BIT-01 and BIT-02 pending because the shared readiness gate reports both blocked on remaining phase consumers."

patterns-established:
  - "C shift safety is verified from emitted scalar function regions, separately from support-library shifts."
  - "Native projection checks event-free operation accounting independently from the native numeric answer."

requirements-completed: []
coverage:
  - id: D1
    description: A checked left shift compiles and runs through the native C17 route to the pinned U64 answer.
    requirement: BIT-02
    verification:
      - kind: integration
        ref: "internal/compiler/cgen/cgen_phase35_test.go#TestPhase35NativeShiftTracer"
        status: pass
    human_judgment: false
  - id: D2
    description: Exported Emit, EmitNative, and EmitApplication independently reject malformed raw shift immediates before C emission.
    requirement: BIT-02
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_phase35_test.go#TestPhase35EmitterAdmissionGates"
        status: pass
    human_judgment: false
  - id: D3
    description: Five pure scalar kinds keep event-free projection, vectors pass, and the unused Byte signature keeps its Byte input and U64 result.
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_phase35_test.go#TestPhase35UnsignedShiftEmissionAndEventProjection"
        status: pass
    human_judgment: false
  - id: D4
    description: The generated-C scan finds exactly two user shift sites and rejects unsafe or removed-site controls.
    requirement: BIT-02
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_phase35_test.go#TestPhase35ShiftSiteControls"
        status: pass
    human_judgment: false

duration: 37min
completed: 2026-10-14
status: complete
---

# Phase 35 Plan 06: Native Shift Emission and Event Projection Summary

**Checked U64 shifts now emit as defined unsigned C, run to the pinned native answer, and pass a positive two-site generated-C census with unsafe-shift controls.**

## Performance

- **Duration:** 37 min
- **Started:** 2026-10-14T16:43:34Z
- **Completed:** 2026-10-14T17:20:37Z
- **Tasks:** 2
- **Files modified:** 9 across the two task commits, the compatibility repair, and its receipt

## Accomplishments

- Compiled and ran the checked `1 << 63` tracer through the installed C17 runner at `-O0`; the test pins the result `9223372036854775808`.
- Emitted left and right shifts with unsigned widened operands and final exact-width casts, while retaining the existing event-free operation projection and return bookkeeping.
- Authored an independent manifest for two generated scalar shift sites. The scan and controls cover dynamic counts, count 64, signed operands, and a removed expected site; vector cases cover counts 0, 31, 32, and 63.
- Preserved the unused `Byte` formal with a U64 result and U64/Bool locals without adding a Byte read, conversion, or default input.

## Verification Evidence

Both commands ran from the checkout root with `GOCACHE=/tmp/ai-lang-verification-gocache`, `-count=1`, verbose output, and a 600-second timeout. Each selected top-level test printed RUN and PASS; no selected test skipped.

- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -run '^(TestPhase35NativeShiftTracer|TestPhase35UnsignedShiftEmissionAndEventProjection|TestPhase35EmitterAdmissionGates|TestPhase35ShiftSiteControls)$' -count=1 -v -timeout=600s` — PASS. This includes all four task tests, native execution, the three exported emitter routes, shift boundary vectors, and dynamic/count-64/signed-operand controls for both shift directions.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase35PreChangeEmitterCorpus|TestPhase35PreChangeCorpusControls|TestPhase35PreChangeNumericDeltaControls)$' -count=1 -v -timeout=600s` — PASS. The frozen emitter corpus, its controls, and the numeric-delta controls all ran and passed after the compatibility repair.

## Task Commits

1. **Task 1: Compile one checked literal shift to its pinned native answer** — `b64d2159` (`feat(35-06): lower checked U64 shifts to unsigned C`)
2. **Task 2: Census all scalar/event seams and reach unsafe generated-shift controls** — `ed8475b9` (`feat(35-06): census scalar shift C evidence`)

The measured plan ledger also includes the closeout repair and its verification record: `5883b2fe` (`fix(35-04): preserve legacy Byte validator contracts`) and `a5ccff74` (`docs(debug): resolve phase35-byte-compat session`). The final planning metadata commit follows this summary.

## Files Created/Modified

- `internal/compiler/cgen/cgen_phase35_test.go` — native tracer, independent emitter gates, event projection, and positive shift-site scan controls.
- `internal/compiler/cgen/cgen_program.go` — U64 shift lowering, scalar selection, event-free projection, and capacity handling.
- `testdata/phase35/shift-sites.json` — independently authored two-site shift census and expected control failures.

## Decisions Made

- Keep the native expected value literal in the test and invoke the installed runner at `-O0`; no compiler dependency or optimization matrix was added.
- Count only user-operation shifts inside generated scalar function regions and require a positive exact count.
- Leave BIT-01 and BIT-02 pending; `requirements.ready-ids` returned an empty ready set and reported both blocked on remaining phase consumers.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Restored legacy Byte validator admission after the frozen emitter corpus exposed a regression.**
- **Found during:** Phase 35-06 closeout verification of the Phase 35-01 frozen C corpus.
- **Issue:** The initial comparison found 158 byte mismatches across 34 rows because the new scalar-bridge Byte guard was applied to legacy Byte functions and blocked valid historical admission.
- **Fix:** Commit `5883b2fe` made core and origin applicability depend independently on scalar instruction families or the reserved scalar Bool/Unit facts, while retaining the unused-Byte signature and local eligibility laws. The original failure remains recorded in the compatibility receipt; it was neither accepted as a passing gate nor removed from the history. A fresh corpus and control rerun passed after the repair.
- **Files modified:** `internal/compiler/corevalidate/corevalidate.go`, `internal/compiler/corevalidate/corevalidate_phase35_test.go`, `internal/compiler/originvalidate/originvalidate.go`, `internal/compiler/originvalidate/originvalidate_phase35_test.go`.
- **Verification:** The session command above freshly passed `TestPhase35PreChangeEmitterCorpus`, `TestPhase35PreChangeCorpusControls`, and `TestPhase35PreChangeNumericDeltaControls`; the repair receipt is `.planning/debug/resolved/phase35-byte-compat.md`.
- **Committed in:** `5883b2fe`; diagnosis and final evidence recorded in `a5ccff74`.

**Total deviations:** 1 auto-fixed bug. No 35-06 source changes were needed after its two task commits.

## Issues Encountered

The initial frozen-corpus comparison failed on the compatibility regression described above. The final fresh comparison passed after the upstream validator repair; no required 35-06 verification remains unrun.

## User Setup Required

None — the plan used the installed C17 runner and added no dependency.

## Next Phase Readiness

Plan 35-06's native emission and event-projection evidence is ready for Plan 35-07's public conformance and optimization checks. Phase 35 remains in progress, and BIT-01/BIT-02 remain blocked pending their remaining engine and public-consumer evidence. No Phase 35 completion or hosted receipt is claimed.

---
*Phase: 35-bitwise-operators-literal-shifts-and-hex-literals*
*Completed: 2026-10-14*

## Self-Check: PASSED

- The required summary file exists.
- Task commits `b64d2159` and `ed8475b9`, repair commit `5883b2fe`, and repair receipt commit `a5ccff74` are ancestors of the current plan head.
- The focused stub scan found no TODO, FIXME, placeholder, coming-soon, or not-available markers in the plan-owned C emitter/test/fixture files.
