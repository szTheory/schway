---
phase: 31-bounded-file-checksum-and-second-consumer
reviewed: 2026-10-12T00:00:00Z
depth: standard
files_reviewed: 25
files_reviewed_list:
  - examples/checksum/adapter.c
  - examples/checksum/adapter.h
  - examples/checksum/file_bytes.bindings.json
  - examples/checksum/checksum.schway
  - examples/checksum/adler_low.schway
  - examples/checksum/adler32.schway
  - examples/checksum/adler32_cli.schway
  - examples/checksum/README.md
  - internal/compiler/ability/ability.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/check/check.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/exitcode/exitcode.go
  - internal/compiler/interp/interp.go
  - internal/compiler/native/bindings.go
  - internal/compiler/native/native_app.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_filebytes.go
  - internal/compiler/session/session_moduledigests.go
  - internal/compiler/testsupport/checksum_oracle.go
findings:
  critical: 0
  warning: 3
  info: 3
  total: 6
status: issues_found
---

# Phase 31: Code Review Report

**Reviewed:** 2026-10-12
**Depth:** standard
**Files Reviewed:** 22 production/example files (test files were not reviewed)
**Status:** issues_found

## Summary

I traced the C adapter, the cgen emitter, the interpreter reader parity, the module digest and receipt validation, and the matcher ordering across check, corevalidate, originvalidate, pathoracle and cgen. I found no blockers.

The adapter handles the cases you listed correctly:
- It opens with O_NONBLOCK and calls fstat on the descriptor.
- It rejects anything that is not a regular file.
- It retries reads on EINTR.
- It probes with limit+1, so a file of exactly 65536 bytes is accepted and one of 65537 is refused.
- Every failure path frees the buffer and closes the descriptor once. The realloc-failure path correctly frees the original buffer.
- It allocates at least one byte, so release never receives NULL on success.
- Nothing in it echoes the path.

The `FileBytes` marker routing is consistent across the five matchers, and none of them can reach the Phase 23 `schway_file_byte_` shapes. The module digests are sorted, `omitempty` keeps no-import identity unchanged, and the receipt validator checks them.

The warnings concern hardening and an implicit coupling between the C status codes and the Schway `FileError` declaration.

## Warnings

### WR-01: adapter opens arbitrary paths without O_NOCTTY

**File:** `examples/checksum/adapter.c:62`
**Issue:** The path is opened before the regular-file check. Opening a terminal device without O_NOCTTY can make it the process's controlling terminal if the process is a session leader with none. Device nodes such as a tty or pty slave are exactly the non-regular inputs the code means to reject. The open itself can have side effects before fstat rejects the file. O_CLOEXEC is also absent, though no exec follows.
**Fix:**
```c
descriptor = SCHWAY_FILE_BYTES_OPEN(path, O_RDONLY | O_NONBLOCK | O_NOCTTY | O_CLOEXEC);
```
This needs `_POSIX_C_SOURCE 200809L`, which is already defined.

### WR-02: status-to-FileError mapping is an unchecked positional coupling

**File:** `internal/compiler/cgen/cgen_program.go:3130-3132` (the `range failure.Alternatives` loop emitting `status == alternative+1`), against `examples/checksum/adapter.h:22-27`
**Issue:** The generated C maps C status N to `Alternatives[N-1]` of the user-declared `data FileError`. The adapter enum is `TOO_LARGE=1`, `UNAVAILABLE=2`, `ALLOC_FAILED=3`. Nothing in check, cgen or the bindings manifest verifies that `FileError` declares exactly `FileTooLarge`, `FileUnavailable`, `AllocFailed` in that order.
- If someone reorders the alternatives, stderr prints the wrong name, for example `FileTooLarge` for an unavailable file. The exit code stays 65, so exit-code-only tests would not notice.
- If someone declares fewer than three alternatives, the status falls through to the bare `if (status != 0) exit(65)` with no stderr line. The interpreter, which keys on names, would diverge from native.
- If someone declares more, the extra alternatives are unreachable.

**Fix:** Pin the exact alternative list in `checkFileBytesEntry`, or in `validateFileBytesOwnerFunction` where the data types are available. Emit by a fixed name-to-status table and refuse the program otherwise:
```go
want := []string{"FileTooLarge", "FileUnavailable", "AllocFailed"}
if !slices.Equal(failure.Alternatives, want) { return fmt.Errorf("function %q: FileError must declare %v in order", function.ID, want) }
```

### WR-03: receipt module path validation is incomplete

**File:** `internal/compiler/native/native_app.go:714-723` (`validateReceiptModules`)
**Issue:** The validator rejects absolute paths, backslashes and `..` segments. It does not reject empty or `.` segments (`a//b`, `./x`) or a Windows drive prefix such as `C:x`. It also does not require unique `Path` values, only unique `Module` names. A receipt can therefore carry two rows with distinct names and the same file path, or a non-canonical path. The receipt is protected by the identity hash, but this validator is the stated gate (T-31-15/16) against malformed receipts, and it admits non-canonical forms.
**Fix:**
```go
for _, part := range strings.Split(row.Path, "/") {
    if part == "" || part == "." || part == ".." { return errors.New("receipt module path must be canonical") }
}
if filepath.Clean(row.Path) != row.Path { ... }
```
Also track a `seenPaths` set.

## Info

### IN-01: close() failure after a complete read is reported as FileUnavailable

**File:** `examples/checksum/adapter.c:90-93`
**Issue:** A successful, fully-read file is discarded and reported unavailable if close() fails. On Linux an EINTR from close() still releases the descriptor, so this turns a successful read into a failure. The Go oracle mirrors the behavior, so the engines agree. The README documents it as a "close failure" cause, so this is a design choice and not a bug. Consider tolerating EINTR on close, or note the choice in a comment.

### IN-02: interpreter trusts the host reader to enforce the 65,536 limit

**File:** `internal/compiler/interp/interp.go:370-375` (`fileForeignCall`, acquire case)
**Issue:** `RunWithFileBytes` accepts any `[]byte` the reader returns and does not itself reject more than 65,536 bytes. The bound lives only in `testsupport.ChecksumReadFile`. Another caller of the public `RunWithFileBytes` with a naive `os.ReadFile` would get native/interp divergence on oversized files. The contract comment says nothing about the limit.
**Fix:** Document that the reader owns the bound. Alternatively, have `fileForeignCall` return `FileTooLarge` (when the `Fails` type declares it) for `len(data) > 65536`.

### IN-03: unstable sort and duplicated entry-prefix logic

**File:** `internal/compiler/session/session_moduledigests.go:41`; `internal/compiler/corevalidate/corevalidate.go:1593-1596`, `internal/compiler/originvalidate/originvalidate.go:1809-1812`
**Issue:**
- `sort.Slice` is not stable. If two sources ever resolve to the same module name, the row order is nondeterministic and the digest list, and so the build ID, would be unstable. The loader presumably forbids this, but `sort.SliceStable` costs nothing.
- The `entryPrefix` suffix-trimming is hand-rolled in two packages, while pathoracle and cgen use `strings.TrimSuffix`. The duplication is deliberate for matcher independence, but the hand-rolled form is easy to get wrong.

**Fix:** Use `sort.SliceStable`. Use `strings.TrimSuffix` consistently.

---

_Reviewed: 2026-10-12_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
