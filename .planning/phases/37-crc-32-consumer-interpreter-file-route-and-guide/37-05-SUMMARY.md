---
phase: 37-crc-32-consumer-interpreter-file-route-and-guide
plan: "05"
subsystem: testing
tags: [crc32, zlib, public-cli, interpreter, native, cross-engine]

requires:
  - phase: "37-03"
    provides: "CRC-32 consumers and the public file-run route"
  - phase: "37-04"
    provides: "Bounded file reader and empty/path boundary behavior"
provides:
  - "Independent CRC-32 recipes, literals, hashes, and generator provenance"
  - "Expected-first public CLI outcome coverage for CRC-32 and checksum consumers"
affects: ["37-06", "Phase 37 acceptance"]

actuals:
  tokens: 11640 # chars/4 over the four plan-owned changed files
  tasks: 2
  commits: 3 # measured from plan_head_before at summary write
plan_head_before: 422d703ef009950b0d39ee8d5d6831818a10f4cb
plan_head_after: 19e8dc63bb4990a7dce9968292aaf3b2fdea551a

tech-stack:
  added: []
  patterns:
    - "Inert byte recipes with independent Python zlib receipts and recurring Go IEEE checks"
    - "Compare each engine to fixed tuples before checking cross-engine parity"
key-files:
  created:
    - examples/crc32/crc32.expected.json
    - internal/compiler/testsupport/crc32_oracle.go
    - internal/compiler/testsupport/crc32_oracle_test.go
    - cmd/schway/main_phase37_matrix_test.go
  modified: []
key-decisions:
  - "Pin Python zlib literals and SHA256 digests; use Go hash/crc32.ChecksumIEEE only as a recurring independent cross-check."
  - "Check both public engines against their own expected tuple before comparing parity."
  - "Test embedded NUL through explicitly labeled public dispatch because host process argv cannot transport it."
patterns-established:
  - "Finite JSON oracle rows bind recipes, digests, provenance, and consumer output tuples."
  - "Public file tests use bounded child processes, private inputs and temp directories, and exact output capture."

requirements-completed: [] # Plan witnesses recorded; parent phase acceptance owns requirement completion.
coverage:
  - id: D1
    description: "Independent CRC-32 and Adler recipes, unsigned literals, low16 projections, digests, and Python/zlib provenance are pinned and drift-tested."
    verification:
      - kind: unit
        ref: "internal/compiler/testsupport/crc32_oracle_test.go#TestPhase37CRC32OracleRows"
        status: pass
      - kind: unit
        ref: "internal/compiler/testsupport/crc32_oracle_test.go#TestPhase37CRC32OracleDriftControls"
        status: pass
    human_judgment: false
  - id: D2
    description: "All 26 OS-representable input cases run through both public engines for four consumers with expected-first outcome checks; direct NUL, manifest failures, and isolated parallel controls are covered separately."
    verification:
      - kind: integration
        ref: "cmd/schway/main_phase37_matrix_test.go#TestPhase37PublicFileMatrix"
        status: pass
      - kind: unit
        ref: "cmd/schway/main_phase37_matrix_test.go#TestPhase37PublicMatrixDriftControl"
        status: pass
    human_judgment: false

duration: 7min
completed: 2026-10-15
status: complete
---

# Phase 37 Plan 05: Independent CRC Oracle and Public File Matrix Summary

**Independent Python/zlib CRC-32 receipts and expected-first interpreter/native file outcomes across the CRC and checksum consumers.**

## Performance

- **Duration:** 7 minutes from the first durable task commit; context loading preceded it, and the executor's earlier start timestamp was not preserved.
- **Started:** 2026-10-15T12:22:24-04:00 (first task commit)
- **Completed:** 2026-10-15T12:29:41-04:00
- **Tasks:** 2
- **Files modified:** 4 plan-owned artifacts

## Accomplishments

- Recorded ten fixed byte recipes with exact lengths, SHA256 digests, unsigned CRC-32 and Adler values, low16 projections, and Python 3.14.4 / zlib 1.2.12 compile/runtime provenance. The exact generator command is pinned by SHA256 `61502ba4db0b2ee2b2208ef6dcc167cd7858e60a3fd9ad4142d8d2192606290f`.
- Cross-checked every in-range CRC literal with Go's `hash/crc32.ChecksumIEEE`, checked Adler values with the Go standard library, and added armed literal, digest, generator-receipt, and failure-tuple drift controls.
- Ran 26 OS-representable cases across four consumers and both public engines (208 CLI children), then checked eight separately labeled NUL dispatch controls, eight missing/bad-manifest cases, and an isolated two-process pair. Every engine matched its exact expected tuple before parity was assessed.
- Pinned key values: empty CRC `0`, zero-byte CRC `3523407757`, check CRC/low16 `3421780262`/`14630`, maximum 65,536-byte FF CRC/low16 `3735780942`/`32334`, and cap-plus-one refusal `(65, "", "FileError.FileTooLarge\n")`.
- The current maximum-input CRC interpreter observation was **2.027450833 seconds** for 65,536 bytes. Earlier Phase 37 evidence recorded a 2.536-second observation; these are single-run observations, not a latency distribution.

## Task Commits

1. **Task 1 test-first assertions:** `def8fb9f` — `test(37-05): add independent CRC oracle assertions`
2. **Task 1 oracle and fixture:** `7b7eee3f` — `test(37-05): bind CRC literals to independent recipes`
3. **Task 2 public matrix:** `19e8dc63` — `test(37-05): exercise expected-first public file matrix`

The summary and canonical state/roadmap updates are captured by the plan metadata commit after this file is self-checked.

## Files Created

- `examples/crc32/crc32.expected.json` — inert byte recipes, independent numeric tuples, provenance, and finite file/path cases.
- `internal/compiler/testsupport/crc32_oracle.go` — test-only recipe expansion and case materialization.
- `internal/compiler/testsupport/crc32_oracle_test.go` — Go IEEE/Adler cross-checks and drift controls.
- `cmd/schway/main_phase37_matrix_test.go` — bounded public CLI matrix and separate NUL, manifest, concurrency, and tuple-drift controls.

## Verification

- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/testsupport -run '^(TestPhase37CRC32OracleRows|TestPhase37CRC32OracleDriftControls)$' -count=1 -v -timeout=600s` — passed both named roots with no skips.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase37PublicFileMatrix|TestPhase37PublicMatrixDriftControl)$' -count=1 -v -timeout=600s` — passed both named roots with no skips. All 26 representable rows reached interpreter and native for every consumer; FIFO, `/dev/null`, unreadable mode-000, and symlink rows were available and executed.
- The 65,537-byte row refused with the fixed `FileError.FileTooLarge` tuple. The exact empty-token tuple remained `(65, "", "")`; 4,096-byte paths reached file lookup and 4,097-byte paths returned the fixed input-length refusal.

## Decisions Made

- Python zlib supplies the recorded oracle; the test suite checks those fixed literals with Go's standard library rather than generating expectations at runtime.
- Each engine is checked against its independent tuple before the matrix compares engine outputs.
- Embedded NUL is a direct-public-dispatch case only; no host argv launch is used to represent a value the OS cannot transport.

## Deviations from Plan

None — plan executed as written.

## Issues Encountered

The test-first commit intentionally failed before the planned fixture existed. After the fixture and test-only oracle were added, both task-1 roots passed; this was the expected RED/GREEN sequence, not a semantic failure.

## Known Stubs

None. The fixture's empty streams and null numeric values are explicit expected tuples for empty-path, empty-file, and oversize cases, not placeholders.

## Next Phase Readiness

The independent receipts and consumer matrix are ready for the remaining Phase 37 work. Requirement completion remains pending parent phase acceptance; this plan records its witnesses without advancing the broader requirement tracker.

---
*Phase: 37-crc-32-consumer-interpreter-file-route-and-guide*
*Completed: 2026-10-15*

## Self-Check: PASSED

- All four plan-owned artifacts exist.
- Task commits `def8fb9f`, `7b7eee3f`, and `19e8dc63` are ancestors of HEAD.
- No stub or placeholder patterns were found in the plan-owned files.
