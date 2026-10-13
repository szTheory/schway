---
status: complete
phase: 29-byte-views-and-the-adler-32-library
source: [29-00-SUMMARY.md, 29-01-SUMMARY.md, 29-02-SUMMARY.md, 29-03-SUMMARY.md, 29-04-SUMMARY.md, 29-05-SUMMARY.md, 29-06-SUMMARY.md, 29-07-SUMMARY.md, 29-08-SUMMARY.md, 29-09-SUMMARY.md, 29-10-SUMMARY.md, 29-11-SUMMARY.md, 29-12-SUMMARY.md]
started: 2026-10-12T02:36:08Z
updated: 2026-10-12T02:38:13.129Z
---

## Current Test

[testing complete]

## Tests

### 1. Adler-32 fixture prints the reference checksums
expected: From the repo root: `go build -o /tmp/schway ./cmd/schway`, then `/tmp/schway build examples/phase29/adler32_fixture.schway --manifest examples/phase29/fixture_bytes.bindings.json --output /tmp/adler32-fixture`. `/tmp/schway app run /tmp/adler32-fixture -- 3` prints 300286872 and `/tmp/schway app run /tmp/adler32-fixture -- 6` prints 33620352 (0x80 0xFF read as 128 and 255, not sign-extended). Both exit 0 with empty stderr. Cases 0..9 match the table in examples/phase29/README.md.
result: pass
source: automated
evidence: TestPhase29FixtureAppCLI (cmd/schway; all 10 cases via public build --manifest + app run), TestPhase29AdlerBothEngines (O0/O3/O3-LTO + interpreter). Agent CLI run 2026-10-12: cases 0..9 print the README table values, exit 0, empty stderr.

### 2. Unknown case id is refused
expected: `/tmp/schway app run /tmp/adler32-fixture -- 10` prints nothing on stdout and exits 65 (the acquire reports FixtureMissing; no owner is published).
result: pass
source: automated
evidence: TestPhase29FixtureAppCLI/unknown-case, TestPhase29FixtureBytesUnknownCaseRefused, TestPhase29InterpreterFixtureUnknownCase. Agent CLI run: case 10 -> empty stdout, exit 65.

### 3. Out-of-bounds byte read is caught, not read
expected: `/tmp/schway build examples/phase29/byte_index_oob.schway --manifest examples/phase29/fixture_bytes.bindings.json --output /tmp/byte-index-oob`, then `/tmp/schway app run /tmp/byte-index-oob -- 0` and `/tmp/schway app run /tmp/byte-index-oob -- 2` each print nothing on stdout, print exactly `schway: byte index out of bounds` on stderr, and exit 65.
result: pass
source: automated
evidence: TestPhase29OOBControlAppCLI (new, cmd/schway; public CLI, exact stderr), TestPhase29OOBApplicationLanes (O0/O3/O3-LTO/ASan + interpreter). Agent CLI run: cases 0 and 2 -> empty stdout, stderr `schway: byte index out of bounds\n`, exit 65.

### 4. Pure Adler-32 library source checks clean
expected: `/tmp/schway check testdata/phase29/adler32_library.schway` exits 0 with a `check pass` result line; `byte_len` and `byte_at` on a `ByteView` parameter type-check with no foreign owner in the file.
result: pass
source: automated
evidence: phase28NewlyAddedEntries row for testdata/phase29/adler32_library.schway (Check exit 0) in session_phase28_manifest_test.go. Agent CLI run: `check pass`, exit 0.

### 5. byte_len / byte_at names are reserved
expected: Save a file /tmp/shadow.schway containing `module tmp.shadow` / `export { fn main }` / `fn byte_len(n: U64) -> U64 { n }` / `fn main(n: U64) -> U64 { byte_len(n) }` (one per line). `/tmp/schway check /tmp/shadow.schway` exits 2 and reports `check.reserved_builtin_name` ("byte_len and byte_at are reserved byte-read builtins").
result: pass
source: automated
evidence: TestPhase29ReservedBuiltinNames (check_phase29_surface_test.go). Agent CLI run: user fn byte_len -> exit 2, check.reserved_builtin_name. (A main that calls byte_len(n) with a U64 stops earlier at the parser, syntax.expected_rbrace, which is the builtin-atom arity path, not a defect.)

### 6. Example README documents cases and sanitizer evidence
expected: examples/phase29/README.md shows the 10-row case table with expected checksums (sources: Python zlib.adler32 and Go hash/adler32), the build/run commands, the out-of-bounds control with its exit 65 and D-20 abort-like note, and a "Sanitizer evidence" table with three AddressSanitizer lanes.
result: pass
source: automated
evidence: TestPhase29ExampleReadmeCaseTable (new, cmd/schway): every README case row equals the pinned oracle table and the OOB stderr/exit-65 claims are present; armed by mutating case 3 to 300286873 (fails, then restored). Sanitizer table and commands verified by agent read.

## Summary

total: 6
passed: 6
issues: 0
pending: 0
skipped: 0
blocked: 0

## Automation Note

All six checkpoints were run by the agent and are pinned in `go test ./...` CI; no human checkpoint remained (PROJECT.md § Verification Operating Preference). Two tests were added to close the only uncovered seams: TestPhase29OOBControlAppCLI and TestPhase29ExampleReadmeCaseTable.

## Gaps

[none]
