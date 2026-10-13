---
phase: 31-bounded-file-checksum-and-second-consumer
fixed_at: 2026-10-12T00:00:00Z
review_path: .planning/phases/31-bounded-file-checksum-and-second-consumer/31-REVIEW.md
iteration: 1
findings_in_scope: 3
fixed: 3
skipped: 0
status: all_fixed
---

# Phase 31: Code Review Fix Report

**Fixed at:** 2026-10-12
**Source review:** .planning/phases/31-bounded-file-checksum-and-second-consumer/31-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope: 3 (critical 0, warning 3; Info excluded)
- Fixed: 3
- Skipped: 0

Verification ran in the isolated review-fix worktree (not the main checkout) with
`go test ./internal/compiler/... ./cmd/... -run 'Phase31|Phase30|Phase29|Phase23'`
and the session inventory/groundedness/lifecycle set; all green. `session.go` is untouched, so the
pinned EmitNative line is unchanged.

## Fixed Issues

### WR-01: adapter opens arbitrary paths without O_NOCTTY

**Files modified:** `examples/checksum/adapter.c`, `internal/compiler/native/testdata/phase31_observer.c`, `internal/compiler/native/native_phase31_fault_test.go`
**Commit:** 85a0fe39
**Applied fix:** The open now ORs `SCHWAY_FILE_BYTES_NOCTTY` and `SCHWAY_FILE_BYTES_CLOEXEC`, which expand to `O_NOCTTY` / `O_CLOEXEC` where defined and to 0 otherwise. The `SCHWAY_FILE_BYTES_OPEN` hook macros are unchanged and `examples/phase23/adapter.c` is untouched.
**Test:** `TestPhase31BoundedReadMutants/open-flags-mutant`. The counting observer has a new `require-open-flags` fault mode that refuses any open lacking O_NOCTTY, O_CLOEXEC (where defined) or O_NONBLOCK. The faithful adapter still returns the Wikipedia checksum; mutants that drop either flag turn the same input into `FileError.FileUnavailable` (exit 65). The existing `nonblocking-mutant` anchor (`O_RDONLY | O_NONBLOCK`) still matches exactly once.

### WR-02: status-to-FileError mapping is an unchecked positional coupling

**Files modified:** `internal/compiler/check/check.go`, `internal/compiler/cgen/cgen_program.go`, `internal/compiler/check/check_phase31_filebytes_test.go`, `internal/compiler/cgen/cgen_phase31_filebytes_test.go`
**Commit:** 949f2386
**Applied fix:** Chose refusal over a name-keyed table. The adapter's C status codes 1..3 are a closed ABI contract with a fixed order, so refusing any other `FileError` matches the closed-shape style of the rest of the file-bytes matcher and keeps the interpreter (which already keys on alternative names) in parity without touching it. A name-keyed table would have needed cgen to guess a status for names it does not know and would have silently accepted extra or missing alternatives. Both gates are independent copies of the rule:
- check: `checkFileBytesEntry` now receives `dataTypes` and refuses with `check.file_bytes_error_alternatives` unless `FileError` is exactly `FileTooLarge, FileUnavailable, AllocFailed`.
- cgen: `emitProgramFileBytesFunction` returns a `file-bytes owner failure type ... must declare exactly FileTooLarge, FileUnavailable, AllocFailed in that order` error.
- Interpreter: unchanged, because it never maps alternatives positionally; the refusal upstream guarantees native and interpreter see the same declaration.
**Test:** `TestPhase31FileErrorDeclarationIsPinned` (check; faithful control plus reversed, swapped, too-few, too-many, renamed, duplicate; 6 mutants fail with the fix removed) and `TestPhase31FileErrorStderrTable/reversed-alternatives` (cgen; rewritten from "table follows the reversed order" to "refuse" for reversed, too-few, too-many, renamed; fails without the fix). The `derived-table` subtest is unchanged and remains the faithful control.
**Status note:** logic change; recommend a human glance that refusing (rather than name-keying) is the intended policy.

### WR-03: receipt module path validation is incomplete

**Files modified:** `internal/compiler/native/native_app.go`, `internal/compiler/native/native_phase31_receipt_modules_internal_test.go` (new)
**Commit:** 7256e8dc
**Applied fix:** `validateReceiptModules` now also rejects empty and `.` path segments (`a//b`, `./x`, trailing slash), any path containing `:` (drive or scheme prefix such as `C:x`) or a NUL, and duplicate `Path` values via a `seenPaths` set. Duplicate module names were already refused by the strict-sort check; the table pins that too. Existing messages for absolute, `..` and empty paths are preserved so `TestPhase31ModuleDigests/receipt-validation` still matches.
**Test:** `TestPhase31ValidateReceiptModulesTable`, a toolchain-free internal table test with faithful rows (flat and nested) as controls. Six of the malformed rows fail with the fix removed.

---

_Fixed: 2026-10-12_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
