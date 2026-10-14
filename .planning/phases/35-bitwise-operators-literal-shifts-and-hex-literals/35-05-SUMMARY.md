---
phase: 35-bitwise-operators-literal-shifts-and-hex-literals
plan: "05"
subsystem: compiler-runtime
tags: [Go, U64, bitwise, literal-shifts, interpreter, pathoracle, executionpeer]

requires:
  - phase: 35-04
    provides: Independently validated five-op scalar transfers and the narrow unused-Byte source bridge
provides:
  - Interpreter execution for U64 XOR, AND, OR, left shift, and logical right shift
  - Independent decimal value vectors, per-op wrong-answer controls, and an unused-Byte runtime witness
  - Loan-free pathoracle ancestry and causal-only event-free classification
affects: [35-06, 35-07, interpreter, pathoracle, executionpeer]

actuals:
  tokens: 8098
  tasks: 3
  commits: 3
commits: 3
plan_head_before: 7910972f7c0b94fc3faa177d2ccc4ce26b3851af
plan_head_after: d31404e4b44f38165ca0a46eb0fa21c0c0db70c6

tech-stack:
  added: []
  patterns:
    - Pure scalar operation results remain event-free while existing return bookkeeping is preserved.
    - Interpreter numeric evidence, path ancestry, and executionpeer causal evidence are checked independently.

key-files:
  created:
    - examples/bitwise.schway
    - testdata/phase35/value-vectors.json
    - internal/compiler/pathoracle/pathoracle_phase35_test.go
    - internal/compiler/executionpeer/executionpeer_phase35_test.go
  modified:
    - internal/compiler/interp/interp.go
    - internal/compiler/interp/interp_phase35_test.go
    - internal/compiler/core/core_test.go
    - internal/compiler/pathoracle/pathoracle.go
    - internal/compiler/executionpeer/executionpeer.go

key-decisions:
  - "Keep executionpeer limited to event-free causal admission and fabricated-event refusal; it does not recompute numeric values."
  - "Preserve the Byte formal as Byte and exercise it only as an unused entry parameter."
  - "Pin interpreter outputs to independently authored decimal literals, including discriminating expression association answers."

patterns-established:
  - "Every pure U64 operation has separate numeric, ancestry, and causal controls at the peer that owns each fact."
  - "Literal shifts use checked counts from 0 through 63 and logical unsigned right-shift semantics."

requirements-completed: []
coverage:
  - id: D1
    description: Five U64 operators match independent vectors and each wrong-result substitution is caught.
    verification:
      - kind: unit
        ref: "internal/compiler/interp/interp_phase35_test.go#TestPhase35BitwiseShiftVectorsEventFree"
        status: pass
      - kind: unit
        ref: "internal/compiler/interp/interp_phase35_test.go#TestPhase35InterpreterWrongAnswerControls"
        status: pass
    human_judgment: false
  - id: D2
    description: The tracked unused-Byte example executes the locked expression and yields U64 value 264.
    verification:
      - kind: unit
        ref: "internal/compiler/interp/interp_phase35_test.go#TestPhase35UnusedByteRuntime"
        status: pass
    human_judgment: false
  - id: D3
    description: Each pure scalar operation ends loan ancestry in straight-line and CFG path-oracle routes.
    verification:
      - kind: unit
        ref: "internal/compiler/pathoracle/pathoracle_phase35_test.go#TestPhase35PureOperationAncestry"
        status: pass
    human_judgment: false
  - id: D4
    description: The causal peer admits event-free operations and rejects their fabricated step and defect events.
    verification:
      - kind: unit
        ref: "internal/compiler/executionpeer/executionpeer_phase35_test.go#TestPhase35PureOperationCausalControls"
        status: pass
    human_judgment: false

duration: 8min
completed: 2026-10-14
status: complete
---

# Phase 35 Plan 05: Interpreter Bitwise Execution and Causal Evidence Summary

**The interpreter now executes all five U64 bitwise and shift operations against independent literal vectors, while the ancestry and causal peers independently preserve their event-free, loan-free contracts.**

## Performance

- **Duration:** 8min
- **Started:** 2026-10-14T16:34:09Z
- **Completed:** 2026-10-14T16:42:31Z
- **Tasks:** 3
- **Files modified:** 9

## Accomplishments

- Added interpreter dispatch for XOR, AND, OR, left shift, and logical right shift, with defensive operand and shift-count checks and no new operation event or defect.
- Authored 34 independent value vectors for boundary values, hexadecimal spellings, shift counts 0/31/32/63, and two expression associations; all five wrong-answer substitutions are detected.
- Added the real `main(unused: Byte) -> U64` witness and registered it for exhaustive dispatch coverage; its locked expression returns 264 and its alternate association returns 8.
- Made pathoracle ancestry loan-free for each operation and made executionpeer explicitly reject fabricated step/defect events while admitting a clean return-only document.

## Verification Evidence

All selectors ran from the checkout root with `GOCACHE=/tmp/ai-lang-verification-gocache`, `-count=1`, verbose output, and a 600-second timeout. The final closeout rerun printed RUN and PASS for all five selected top-level tests and their per-operation controls:

- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp -run '^(TestPhase35BitwiseShiftVectorsEventFree|TestPhase35InterpreterWrongAnswerControls|TestPhase35UnusedByteRuntime)$' -count=1 -v -timeout=600s`
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/pathoracle -run '^TestPhase35PureOperationAncestry$' -count=1 -v -timeout=600s`
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/executionpeer -run '^TestPhase35PureOperationCausalControls$' -count=1 -v -timeout=600s`
- `python3 -m json.tool testdata/phase35/value-vectors.json` parsed the authored fixture successfully.

## Task Commits

1. **Task 1: Execute one checked literal shift to an event-free interpreter outcome** - `149003cf` (`feat(35-05): run pure U64 operations in the interpreter`)
2. **Task 2: Expand literal vectors and execute the unused-Byte bridge** - `e0cb6c39` (`feat(35-05): add independent bitwise vectors and runtime witness`)
3. **Task 3: Validate per-op causal admission and forged-event rejection** - `d31404e4` (`feat(35-05): reject causal events at pure U64 operations`)

**Plan metadata:** a separate SDK metadata commit follows state and roadmap updates.

## Files Created/Modified

- `internal/compiler/interp/interp.go` - validated pure U64 runtime dispatch and unsigned shift semantics.
- `internal/compiler/interp/interp_phase35_test.go` - independent vectors, wrong-answer controls, and the tracked unused-Byte witness.
- `examples/bitwise.schway` - real source witness containing all five operations.
- `internal/compiler/core/core_test.go` - additive exhaustive-dispatch fixture registration.
- `testdata/phase35/value-vectors.json` - independently authored decimal answers and source forms.
- `internal/compiler/pathoracle/pathoracle.go` - loan-free scalar ancestry classification.
- `internal/compiler/pathoracle/pathoracle_phase35_test.go` - per-operation straight-line and CFG ancestry controls.
- `internal/compiler/executionpeer/executionpeer.go` - closed event-free operation inventory and causal refusals.
- `internal/compiler/executionpeer/executionpeer_phase35_test.go` - clean return-only admission and ten fabricated-event controls.

## Decisions Made

- Kept the execution peer a causal checker: it neither receives runtime values nor replays numeric results.
- Kept the `Byte` entry formal unchanged and unused; the separate U64 runner input contract remains untouched.
- Kept expected decimal values in a fixture authored independently from production engines; the vector fixture remains available for Plan 35-07.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking issue] Used the existing structured scalar body shape for the tracked witness.**
- **Found during:** Task 2
- **Issue:** The accepted D-07 source route requires an initial local declaration and a final named-local return, so a direct operator expression as the body result was rejected by the parser.
- **Fix:** Added the required seed local and bound the locked expression to `witness`, then returned that identifier; the expected expression and value stayed unchanged.
- **Files modified:** `examples/bitwise.schway`, `internal/compiler/interp/interp_phase35_test.go`
- **Verification:** `TestPhase35UnusedByteRuntime` passed and checked 264 against the independently run alternate result 8.
- **Committed in:** `e0cb6c39`

**Total deviations:** 1 auto-fixed blocking issue. No semantic scope expansion.

## Issues Encountered

The interpreter test uses the external `interp_test` package to avoid an import cycle. Its event-free assertion permits the existing scalar-copy return bookkeeping event while refusing events attributed to any of the five pure operations and any defect.

## User Setup Required

None - no external service or package setup is required.

## Next Phase Readiness

The interpreter, pathoracle, and causal executionpeer evidence is committed and ready for Plan 35-07's public/native checks and all-site dispatch inventory after the emitter owner completes. `requirements.ready-ids` returned no ready requirements and reported BIT-01 and BIT-02 blocked, so this plan does not mark them complete; the remaining consumers must provide their evidence first.

---
*Phase: 35-bitwise-operators-literal-shifts-and-hex-literals*
*Completed: 2026-10-14*

## Self-Check: PASSED

- Summary file exists at the required plan path.
- All three task commits are ancestors of the current `HEAD`.
- The committed-source stub-pattern scan found no TODO, FIXME, placeholder, coming-soon, or not-available stub markers in the nine plan-owned source/test/fixture files.
