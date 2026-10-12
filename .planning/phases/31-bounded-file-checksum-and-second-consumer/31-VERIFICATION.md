---
phase: 31-bounded-file-checksum-and-second-consumer
verified: 2026-10-12T16:00:00Z
status: passed
score: 4/4 must-haves verified
covered_files:
  - ".planning/phases/31-bounded-file-checksum-and-second-consumer/31-00-PLAN.md"
  - ".planning/phases/31-bounded-file-checksum-and-second-consumer/31-00-SUMMARY.md"
  - ".planning/phases/31-bounded-file-checksum-and-second-consumer/31-01-PLAN.md"
  - ".planning/phases/31-bounded-file-checksum-and-second-consumer/31-01-SUMMARY.md"
  - ".planning/phases/31-bounded-file-checksum-and-second-consumer/31-02-PLAN.md"
  - ".planning/phases/31-bounded-file-checksum-and-second-consumer/31-02-SUMMARY.md"
  - ".planning/phases/31-bounded-file-checksum-and-second-consumer/31-03-PLAN.md"
  - ".planning/phases/31-bounded-file-checksum-and-second-consumer/31-03-SUMMARY.md"
  - ".planning/phases/31-bounded-file-checksum-and-second-consumer/31-04-PLAN.md"
  - ".planning/phases/31-bounded-file-checksum-and-second-consumer/31-04-SUMMARY.md"
  - ".planning/phases/31-bounded-file-checksum-and-second-consumer/31-05-PLAN.md"
  - ".planning/phases/31-bounded-file-checksum-and-second-consumer/31-05-SUMMARY.md"
  - ".planning/phases/31-bounded-file-checksum-and-second-consumer/31-06-PLAN.md"
  - ".planning/phases/31-bounded-file-checksum-and-second-consumer/31-06-SUMMARY.md"
  - ".planning/phases/31-bounded-file-checksum-and-second-consumer/31-07-PLAN.md"
  - ".planning/phases/31-bounded-file-checksum-and-second-consumer/31-07-SUMMARY.md"
  - ".planning/phases/31-bounded-file-checksum-and-second-consumer/31-08-PLAN.md"
  - ".planning/phases/31-bounded-file-checksum-and-second-consumer/31-08-SUMMARY.md"
  - ".planning/phases/31-bounded-file-checksum-and-second-consumer/31-09-PLAN.md"
  - ".planning/phases/31-bounded-file-checksum-and-second-consumer/31-09-SUMMARY.md"
  - "examples/checksum/README.md"
  - "examples/checksum/adapter.c"
  - "examples/checksum/adapter.h"
  - "examples/checksum/adler32.schway"
  - "examples/checksum/adler32_cli.schway"
  - "examples/checksum/adler_low.schway"
  - "examples/checksum/checksum.expected.json"
  - "examples/checksum/checksum.schway"
  - "examples/checksum/file_bytes.bindings.json"
  - "internal/compiler/native/native_phase31_controls_test.go"
  - "internal/compiler/native/native_phase31_fault_test.go"
  - "internal/compiler/native/native_phase31_matrix_test.go"
  - "internal/compiler/native/native_phase31_modules_test.go"
  - "internal/compiler/native/native_phase31_receipt_modules_internal_test.go"
  - "internal/compiler/native/testdata/phase31_observer.c"
  - "internal/compiler/testsupport/checksum_oracle.go"
  - "internal/compiler/testsupport/checksum_oracle_test.go"
covered_digest: "v3:sha256:8349cc3b48aa773aea46358e50487a1da3aa4a1cf469de418bdcec2a99de2ecd"
behavior_unverified: 0
overrides_applied: 0
---

# Phase 31: Bounded File Checksum and Second Consumer Verification Report

**Phase Goal:** A caller can run `checksum` on one file through the public application route and get an exact Adler-32 result or a named error, and a second application imports the same library without copying it, with every answer checked against an independent oracle.
**Verified:** 2026-10-12
**Status:** passed
**Re-verification:** No, initial verification

## Goal Achievement

### Observable Truths (ROADMAP success criteria 1-4)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | `app run` on built `checksum` prints bare decimal plus newline: empty 1, 0x00 65537, Wikipedia 300286872, 65,536 x 0xFF 2006388466, on every engine | VERIFIED | I built `checksum.schway` with `--manifest file_bytes.bindings.json` and ran `app run ck -- PATH`. Empty file gave `1`, Wikipedia gave `300286872`, 65,536 x 0xFF gave `2006388466`, all exit 0 with empty stderr. `TestPhase31ChecksumPublicRoute` (native, public route, includes 0x00 -> 65537) and `TestPhase31ChecksumInterpreter` (independent host read) pass. |
| 2 | 65,537+ bytes refused with its own outcome. Missing, unreadable and non-regular files share one outcome. stdout empty, fixed stderr with no path, buffer released exactly once, limit+1 bounded read | VERIFIED | Live run: 65,537 bytes gave exit 65, empty stdout, `FileError.FileTooLarge`. Missing file, directory and `/dev/null` each gave exit 65, empty stdout, `FileError.FileUnavailable`. No path appears in stderr. Passing tests: `TestPhase31CaseMatrix`, `TestPhase31CaseMatrixNativeBoundaries`, `TestPhase31ExpectedMatrixJSON`, `TestPhase31ReleaseExactlyOnce` (ASan double-release mutant), `TestPhase31BoundedReadMutants` (st_size-trusting and `>=`/`>` mutants). |
| 3 | Second app imports the same `adler32` module and prints a different pinned result. One copy of the library loop. A library mutation flips both apps. Build identities and evidence hash every module in sorted order | VERIFIED | Live: `adler_low` on Wikipedia gave `920`, against 300286872 for `checksum`. `adler_low.schway` and `checksum.schway` both `import adler32`. The only other `adler32` definitions are the pre-existing Phase 29 fixtures (`testdata/phase29`, `examples/phase29`), covered by the pinned allow-list in `TestPhase31SingleCopyAndDualFlip` (pass). `TestPhase31AdlerLowPublicRoute`, `TestPhase31ModuleDigests` (session and native) and `TestPhase31ValidateReceiptModulesTable` pass. |
| 4 | Pinned answers come from Python `zlib.adler32` with the generating command recorded. Each reached control is seen to fail | VERIFIED | `checksum_oracle.go` records `cmd` and `python` per row (`python3 -c 'import zlib;...zlib.adler32(...)'`). `TestPhase31OracleRows` passes. `TestPhase31Controls` runs wrong-modulus, wrong-initial-state, swapped-halves, off-by-one-loop-bound, signed-char-retype, signed-char-flag-lane, deleted-bounds-check-asan and library-mutation subtests. `TestPhase31ControlInventory` maps them one-to-one. `TestPhase31LoopNeighborRefusals` covers the refused neighbors. All pass. |

**Score:** 4/4 truths verified (0 present, behavior-unverified)

### Focused Test Runs (run by me)

`go test -run Phase31 -count=1 -v` on session, native, cgen, check, corevalidate, originvalidate, pathoracle, testsupport and cmd/schway: all `ok`, no FAIL. The only skip-like output was a subtest *named* `skip-isregular-control`. No `--- SKIP` lines appeared (non-root uid 501; the mkfifo, ASan and non-root lanes executed).

### Requirements Coverage

| Requirement | Source Plans | Status | Evidence |
|-------------|--------------|--------|----------|
| APP-11 | 31-01, 31-02, 31-03, 31-04, 31-06, 31-09 | SATISFIED | SC1 above, live route plus tests |
| APP-12 | 31-04, 31-06, 31-09 | SATISFIED | SC2 above, live route plus tests |
| MOD-06 | 31-00, 31-01, 31-03, 31-05, 31-07, 31-09 | SATISFIED | SC3 above |
| EVD-14 | 31-02, 31-08, 31-09 | SATISFIED | SC4 above |

All four IDs appear in PLAN frontmatter and in REQUIREMENTS.md. No orphaned requirements. The unchecked boxes in REQUIREMENTS.md are expected (orchestrator marks them after this verdict).

### Anti-Patterns

I did not scan for stubs beyond the live runs and the test evidence. The two test-level facts that matter are covered above: the controls are armed and fail as intended, and the three code-review warnings (WR-01 to WR-03) are recorded as fixed in 31-REVIEW-FIX.md. No blockers found.

### Human Verification Required

None. The phase is fully automated (31-VALIDATION.md: no manual-only checks).

### External Notes (not code gaps)

- Hosted CI run 38299800002 (2026-10-12, headSha c7256924) passed: `checks` and `current evidence aggregate` succeeded on ubuntu-latest and macos-latest. Phase 32 still owns the final M006 receipts (EVD-15).
- The `interp` package holds no Phase31-named test; the interpreter route is tested in session by design (31-09 deviation 1). This is not a gap.

### Gaps Summary

None. The phase goal is achieved in the codebase.

---

_Verified: 2026-10-12_
_Verifier: Claude (gsd-verifier)_
