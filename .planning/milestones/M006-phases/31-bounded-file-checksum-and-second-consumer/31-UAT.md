---
status: complete
phase: 31-bounded-file-checksum-and-second-consumer
source: [31-00-SUMMARY.md, 31-01-SUMMARY.md, 31-02-SUMMARY.md, 31-03-SUMMARY.md, 31-04-SUMMARY.md, 31-05-SUMMARY.md, 31-06-SUMMARY.md, 31-07-SUMMARY.md, 31-08-SUMMARY.md, 31-09-SUMMARY.md]
started: 2026-10-12T19:30:00Z
updated: 2026-10-12T19:48:45.600Z
---

## Current Test

[testing complete]

## Tests

### 1. checksum prints the reference Adler-32 of a file
expected: From the repo root, build `examples/checksum/checksum.schway` with `--manifest examples/checksum/file_bytes.bindings.json`. `schway app run <artifact> -- PATH` prints the README value: empty → 1, 0x00 → 65537, `Wikipedia` → 300286872, 65,536×0xFF → 2006388466, 0x80 → 8454273, 0x80 0xFF → 33620352. Each exits 0 with empty stderr.
result: pass
source: automated
evidence: TestPhase31ChecksumPublicRoute (with a mutated-library control), TestPhase31ChecksumInterpreter, TestPhase31OracleRows. Agent CLI run on 2026-10-12: all six rows matched, and `python3 zlib.adler32` gave the same values.

### 2. Oversized and unreadable inputs are refused without leaking the path
expected: A 65,537-byte file prints `FileError.FileTooLarge`. A missing file, a directory, a FIFO, `/dev/null` and a mode-000 file each print `FileError.FileUnavailable`. Every case exits 65, stdout is empty, and stderr is that single fixed line, with no part of the path, even under a directory named `sëcret-ρath`. The FIFO returns at once rather than hanging. A path starting with `-` is read as a path. An empty argument is refused with exit 65 and no output; a 4,097-byte argument is refused with `native.input_too_long`, and the path is not echoed.
result: pass
source: automated
evidence: TestPhase31CaseMatrix, TestPhase31CaseMatrixNativeBoundaries, TestPhase31FileErrorStderrTable, TestPhase31BoundedReadMutants, TestPhase31ReleaseExactlyOnce, cmd/schway TestPhase31* (empty-argument is the named silent-by-design row). Agent CLI run: every refusal matched on both apps, and the FIFO returned inside the 10 s alarm.

### 3. adler_low is a second consumer of the same library
expected: `adler_low` builds from the same `adler32.schway` and prints the low 16 bits: empty → 1, 0x00 → 1, `Wikipedia` → 920, 65,536×0xFF → 3826, 0x80 → 129, 0x80 0xFF → 384. It gives the same error table as `checksum`.
result: pass
source: automated
evidence: TestPhase31AdlerLowPublicRoute, TestPhase31SingleCopyAndDualFlip. Agent CLI run: every value and refusal matched; Python `zlib.adler32(...) % 65536` agrees.

### 4. Editing the one library changes both apps
expected: In a copy of `examples/checksum`, changing the A-half modulus changes the 65,536×0xFF output of both `checksum` and `adler_low`. Changing only the B-half modulus changes `checksum` and leaves `adler_low` at 3826.
result: pass
source: automated
evidence: TestPhase31SingleCopyAndDualFlip (copied-library-is-detected, dual flip, B-only control). Agent CLI run: the A mutation gave 2558202096 / 4336; the B mutation gave 2006454002 / 3826 (the originals are 2006388466 / 3826).

### 5. Build identity and receipt record every module
expected: The `checksum` receipt has sorted `modules` rows for `adler32.schway` and `checksum.schway`, each with a sha256 digest and a relative path. A comment-only library edit changes `build_id` and the `adler32` digest but not `emitted_c_digest`. `app run` refuses a receipt whose module rows were tampered with (an absolute path, or rows reordered) with `native.receipt_invalid`, exit 125.
result: pass
source: automated
evidence: TestPhase31ModuleDigests, TestPhase31ValidateReceiptModulesTable. Agent CLI run: build_id dee31723… → 8f20202b…, library digest 317ac139… → 2b7a7860…, emitted C 9bb3babc… unchanged. Both tamper cases were refused with exit 125.

### 6. Programs without imports and earlier corpora are unchanged
expected: Every committed no-import source produces byte-identical output to the pre-Phase-30 snapshot, the accepted and conformance corpora are unchanged, and the moved library is hash-identical apart from its header.
result: pass
source: automated
evidence: TestPhase30PreChangeGolden, TestPhase272AcceptedCorpusByteIdentity, TestPhase273ConformanceCorpusByteIdentity, TestPhase31MovedLibraryIdentity. Focused run on 2026-10-12: `go test ./internal/compiler/... ./cmd/... -run 'Phase31|…golden…'` was ok in cgen, check, corevalidate, native, originvalidate, pathoracle, session, testsupport and cmd/schway.

## Summary

total: 6
passed: 6
issues: 0
pending: 0
skipped: 0
blocked: 0

## Automation Note

The Phase 31 SUMMARYs have no `coverage:` blocks (legacy mode), so the checkpoints come from the phase success criteria SC1–SC4. The agent ran each one through the public `build` + `app run` CLI on a fresh binary, and the named tests pin every checkpoint in CI. No human checkpoint remained (PROJECT.md § Verification Operating Preference). The EVD-14 controls (TestPhase31Controls, TestPhase31ControlInventory) passed in the same focused run. Hosted CI run 38299800002 (at c7256924) was still running when this UAT was recorded.

## Gaps

[none]
