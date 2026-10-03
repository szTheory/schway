---
phase: 27-exact-fizzbuzz-and-cross-host-evidence
plan: 06
type: execute
subsystem: application
tags: [text-output, fizzbuzz, interpreter, c17, fixtures]
requires:
  - phase: 27-exact-fizzbuzz-and-cross-host-evidence
    provides: Admitted literal-only text operations and checked U64 remainder
provides:
  - Exact admitted literal-byte writes in interpreter and native application execution
  - Complete guarded Unit FizzBuzz source with fixed 15- and 1000-line answers
  - Exact 0, 1, 15, 1000, and pre-output 1001 application behavior in both engines
affects: [phase-27-parity, native-output, app-examples]
tech-stack:
  added: []
  patterns: [length-coded-literal-output, independent-golden-output]
key-files:
  created:
    - examples/phase27/fizzbuzz-15.out
    - examples/phase27/fizzbuzz-1000.out
  modified:
    - examples/fizzbuzz.schway
    - internal/compiler/interp/interp.go
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/session/session_phase27_test.go
key-decisions:
  - "Emit admitted text through the existing counted writer using an unsigned-byte compound literal plus the independently validated byte length, so embedded NUL and empty values remain exact."
  - "Keep FizzBuzz as a guarded Unit app: check the input limit before the loop, test divisibility by 15 before 3 and 5, and write one LF after each result."
  - "Use a separate POSIX awk reference that concatenates the Fizz and Buzz labels to author the checked-in expected outputs; do not execute the Schway program or either engine to produce expected bytes."
patterns-established:
  - "Application literal output passes exact byte slices to the bounded writer and never relies on NUL termination."
  - "A checked-in output fixture can pin the full byte stream independently from a differently structured reference implementation."
requirements-completed: [APP-08, APP-09, U64-02]
coverage:
  - id: D-27-01
    description: "Literal Fizz text, embedded NUL, UTF-8 bytes, empty text, and LF reach exact output through interpreter and retained native execution; Unit adds no automatic result line."
    requirement: APP-08
    verification:
      - kind: unit
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase27(LiteralTextApp|LiteralWriteAdmission)$' -count=1"
        status: pass
    human_judgment: false
  - id: D-27-02/D-27-03/D-27-04
    description: "The committed FizzBuzz app matches fixed output bytes at inputs 0, 1, 15, and 1000 in both engines; input 1001 defects with empty stdout, and the existing U64 sum application retains automatic decimal output."
    requirement: APP-08
    verification:
      - kind: unit
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase27(FizzBuzzExact|TracerFizzBuzz1|LiteralTextApp)$|^TestPhase26ExactSumMatrix$' -count=1"
        status: pass
    human_judgment: false

# Metrics
duration: 11min
completed: 2026-10-03
tasks: 2
files_modified: 6
status: complete
---

# Phase 27 Plan 06: Literal Byte Output and Exact FizzBuzz Summary

**Admitted literal bytes now reach the application output in both engines, and the FizzBuzz example prints exact fixed answers through 1000 while rejecting 1001 before output.**

## Performance

- **Duration:** 11 min
- **Started:** 2026-10-03T16:24:00-04:00
- **Completed:** 2026-10-03T16:35:00-04:00
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments

- Added interpreter byte-sink support for `OpWriteText` and length-coded C17 emission through the existing bounded byte writer. Embedded NUL, UTF-8, and empty literals are preserved exactly.
- Completed `examples/fizzbuzz.schway` with a pre-output `n < 1001` guard, a scalar loop, `snapshot`, guarded divisibility checks in 15/3/5 order, decimal fallback, and one newline per result.
- Added independent fixed output files for inputs through 15 and 1000, then compared interpreter and retained native bytes at inputs 0, 1, 15, and 1000.
- Verified 1001 defects before output and preserved the existing U64 sum app's automatic decimal return line.

## Task Commits

1. **Tasks 1–2: Emit literal bytes and complete exact FizzBuzz** - `78ef5c9` (feat)

**Plan metadata:** pending.

## Files Created/Modified

- `examples/fizzbuzz.schway` - Implements guarded FizzBuzz with explicit Unit output.
- `examples/phase27/fizzbuzz-15.out` - Independent expected bytes for the first 15 results.
- `examples/phase27/fizzbuzz-1000.out` - Independent expected bytes for all 1000 results.
- `internal/compiler/interp/interp.go` - Appends admitted text bytes to the interpreter application sink.
- `internal/compiler/cgen/cgen_program.go` - Sends byte arrays and explicit lengths through the bounded native writer; accounts for text writes as event-free application output.
- `internal/compiler/session/session_phase27_test.go` - Covers literal bytes, exact FizzBuzz fixtures, the output limit, Unit behavior, and the snapshot copy event.

## Decisions Made

- Used an unsigned-byte compound literal in generated C and passed the checked byte length to `schway_write_bytes`; this keeps NUL and non-ASCII bytes independent of string termination.
- Authored the golden output with a separate POSIX awk oracle that independently concatenates labels for multiples of three and five, rather than running the source program or either execution engine.
- Updated the Phase 27 tracer's event expectation to allow the new `snapshot` copy events while continuing to require a Unit return and exact output.

## Deviations from Plan

- Updated `schema2ExecutionDocumentSize` in `cgen_program.go` to classify `OpWriteText` with other event-free writes, and added the flat-function C emitter case in addition to the scalar CFG emitter. These are required to keep the operation supported across the existing C generation routes.
- The pre-existing tracer assertion expected only a return event. The new required snapshot copy now produces `value.copied` events, so the assertion now verifies those events precede the final Unit return.

**Total deviations:** 2 necessary compatibility updates.

## Issues Encountered

- The original red test failed because the interpreter attempted to read a source place for the no-operand Unit text operation. After adding the no-source path and counted C emission, both requested selectors passed.
- No full project suite was run. The required focused selectors passed with the installed Clang native build path.

## Next Plan Readiness

- Plan 27-07 can now build the independent interpreter/native parity matrix and reached mutation controls on top of the complete program and fixed answers.
- The fixed 15- and 1000-line fixtures are committed and available to later cross-host evidence plans.

---
*Phase: 27-exact-fizzbuzz-and-cross-host-evidence*
*Completed: 2026-10-03*
