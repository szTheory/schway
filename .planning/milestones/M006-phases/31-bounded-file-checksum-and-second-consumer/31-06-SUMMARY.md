---
phase: 31-bounded-file-checksum-and-second-consumer
plan: 06
subsystem: testing
tags: [checksum, case-matrix, differential, argv-boundary, expected-json]
requires: [31-01, 31-02]
provides:
  - shared ChecksumCaseMatrix run on native and interpreter against literal bytes
  - public-route oracle, oversize, argv-boundary and leading-dash rows
  - examples/checksum/checksum.expected.json (D-08)
  - root env token (D-20)
affects: [31-07, 31-08]
key-files:
  created:
    - internal/compiler/native/native_phase31_matrix_test.go
    - internal/compiler/testsupport/checksum_matrix_test.go
    - examples/checksum/checksum.expected.json
  modified:
    - internal/compiler/testsupport/checksum_oracle.go
    - cmd/schway/main_phase31_test.go
    - internal/compiler/session/session_phase31_interp_test.go
    - internal/compiler/session/session_test.go
key-decisions:
  - "4097-byte argument refusal pinned as the fixed host line 'schway app run: native.input_too_long: argument is 4097 bytes; limit is 4096' (reveals length, never the argument)"
  - "Empty argument confirmed silent: exit 65 with both streams empty"
status: complete
commits: 3
plan_head_before: 76f263894f1b89b5acc781ecf7931966f35e55a7
plan_head_after: fb1ea7cb
actuals:
  tokens: 28000
  tasks: 3
  commits: 3
---

# Phase 31 Plan 06: Checksum case matrix Summary

One shared case list (10 rows) gives byte-identical (exit, stdout, stderr) on the native engine and the interpreter, with literal expectations, the public route pins the four oracle answers plus argv boundaries, and no error ever names the path.

## Accomplishments
- `ChecksumCaseMatrix`/`MaterializeChecksumCase` under a non-ASCII `phase31-secret-ü-segment`; success stdout derived from `ChecksumOracleRows`.
- `TestPhase31CaseMatrix`: 30 passing subtests (10 rows x native/interpreter), no -short gate, FIFO under a 5 s deadline, interpreter success rows hold exactly one release event and no live resources. `TestPhase31CaseMatrixNativeBoundaries` covers NUL refusal and the FIFO deadline.
- Public route: oracle-empty/zero/wikipedia/ff65536, oversize-65537, leading-dash, path-4096, path-4097, empty-argument, path-free-stderr.
- `TestPhase31ExpectedMatrixJSON` ties the JSON to the matrix, `exitcode.AppInputRefused` and the `FileError` alternative order.
- `root` added to `debtRegisterEnvironmentalSet`; `TestNoSuppressionOutlivesItsWitness` green.

## Deviations from Plan
None of substance. Task 1 tracer committed the full 10-row matrix rather than only `ff65537`, since the harness was identical; Task 2 added the release-event, NUL and FIFO-deadline checks as planned. The NUL and FIFO-deadline subtests live in `TestPhase31CaseMatrixNativeBoundaries` (a sibling test) rather than as children of `TestPhase31CaseMatrix`.

## Known Stubs
None.

## Threat Flags
None.

## Self-Check: PASSED
