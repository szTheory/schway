---
phase: 29-byte-views-and-the-adler-32-library
reviewed: 2026-10-11T00:00:00Z
depth: standard
files_reviewed: 23
files_reviewed_list:
  - .github/workflows/ci.yml
  - examples/phase29/README.md
  - examples/phase29/adapter.c
  - examples/phase29/adapter.h
  - examples/phase29/adler32_fixture.schway
  - examples/phase29/byte_index_oob.schway
  - internal/compiler/ability/ability.go
  - internal/compiler/ast/ast.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/check/check.go
  - internal/compiler/check/scalar.go
  - internal/compiler/core/core.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/executionpeer/executionpeer.go
  - internal/compiler/interp/interp.go
  - internal/compiler/native/native_app.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/session/session.go
  - internal/compiler/syntax/parser.go
  - internal/compiler/testsupport/phase29_fixture_bytes.go
findings:
  critical: 0
  warning: 4
  info: 4
  total: 8
status: issues_found
---

# Phase 29: Code Review Report

**Reviewed:** 2026-10-11
**Depth:** standard
**Files Reviewed:** 22 source files (plus the adapter header, counted above as listed)
**Status:** issues_found

## Summary

The Phase 29 diff was read file by file, and adversarial programs were driven through the built `schway` CLI. The main soundness properties hold up.

- The bounds check precedes the load in both C shells. The octet widens through `const unsigned char *` to `uint64_t`, so it is unsigned. There is no `assert` or `restrict`, and nothing the compiler can elide.
- The three validator peers (corevalidate, originvalidate, pathoracle) and cgen each re-derive the closed fixture shape without sharing helpers.
- The interpreter's D-20 live-owner allowance is keyed to the final event resolving to an `OpBytesAt` operation. A user `defect "byte index out of bounds"` resolves to `OpDefect` and is not admitted.
- The C adapter is memory-safe. Case ids outside 0..9 return before any allocation. A zero-length case allocates 1 byte. `free(NULL)` is safe on the unknown-case path. Every `case_fill` write is within `case_length`.
- Shadowed binding names (`let case = view(owner)`) work end to end.

I found no BLOCKER. I did find several robustness and soundness-adjacent gaps, listed below.

Direct answers to the points you flagged:

- **corevalidate `fixtureViewCall`:** An ordinary program cannot reach it. The exemption requires a function-wide fixture-contract operation. Any function carrying one is forced through `fixtureOwnerLifecycleValid` at the end of `replayStraightLine`. The residual concerns are WR-03 and IN-03.
- **interp D-20:** Another defect with a live owner does not slip through. It becomes a hard error instead of a defect Execution, which diverges from native (WR-03).
- **check fixture spoofing:** A user program can spoof the symbol. See WR-01.
- **Forged core without a check operand-type law:** It cannot reach interp or cgen unchecked in any product path. Corevalidate and originvalidate both gate operand and result shapes. See IN-02.

## Warnings

### WR-01: Fixture-owner admission is advertised as test-only but reachable from any user program and manifest

**File:** `internal/compiler/check/check.go:4146-4154` (`fixtureBytesEntryCandidate`), `internal/compiler/session/session.go:1405`
**Issue:** The closed fixture branch is keyed solely on the first binding being `try schway_fixture_bytes_acquire(...)`. The decision records call the owner "test-only" (D-05, D-23), but nothing enforces that. The program text is self-declared.

- Any user source can declare `schway_fixture_bytes_{acquire,view,release}` with the exact contract facts.
- Any user can pass `--manifest` pointing at their own adapter, for example a `view` that returns `length` greater than the allocation. `schway build` accepts it. I confirmed that the shipped examples build through the plain CLI with no test seam.
- `FixtureBytes` is granted Share, Send and Escape in both `ability` and `corevalidate`.
- The byte-view bounds check is only as honest as the adapter-supplied `length`. A user adapter therefore turns `byte_at` into an out-of-bounds read, with the compiler having admitted the shape.

FFI trust is inherent to this design, so I do not rate this a BLOCKER. The "test-only" claim should not be relied on as a security boundary.

**Fix:** Either state in the README, decision records and ability comments that the ByteView length is a trusted foreign fact (same trust class as every local-C owner). Or gate fixture admission, for example on a manifest field or a build-time test seam, so production `schway build` refuses these symbols outside tests.

### WR-02: A `ByteView` entry passes `check` but cgen emits an ill-formed initializer and the interpreter cannot bind the input

**File:** `internal/compiler/cgen/cgen.go:1523-1524`, `internal/compiler/cgen/cgen_program.go:1119`, `internal/compiler/interp/interp.go:871-904`, `internal/compiler/check/scalar.go:35`
**Issue:** `fn main(b: ByteView) -> U64 { var n = byte_len(b)  n }` passes `schway check`. `linearInput` returns an empty initializer for `ByteView`, and `emitProgram` formats it into `SCHWAY_BYTE_VIEW schway_entry_input = ;` at line 1119. That is invalid C, and `schway run --engine=native` fails with the opaque `native.tool_failure`.

`inputValue` falls through to `value{payload: input}`, which has no `isBytes`. The interpreter then fails the first byte op with `byte_len operand is not a byte view`, surfaced as `tool.run_failed`. `schway build` refuses it with an unrelated "only a linear U64-to-U64 ..." message.

Neither engine returns a typed diagnostic for the unsupported entry. The type checker admits an entry that no engine can execute.

**Fix:** Reject a `ByteView` parameter on any exported or entry function in check, with a named diagnostic. Alternatively, make `linearInput` return a typed error (`cgen.byteview_entry_unsupported`) and `inputValue` return an explicit refusal.

### WR-03: Interpreter and native disagree on non-bounds defects with a live fixture owner

**File:** `internal/compiler/interp/interp.go:349-351` (`RunWithFixtureBytes`), `internal/compiler/interp/interp.go:383-392` (`terminalLiveResources`)
**Issue:** Native treats every defect as abort-like. I built a variant whose callee overflows checked addition (`x = 18446744073709551615 + len + 1`). It prints `schway: U64 addition overflow` and exits 65, with the owner leaked.

In the interpreter, the same terminal reports the whole stack's live resources, because `terminalLiveResources` walks all frames on the fixture route. `RunWithFixtureBytes` then accepts a live owner only when `isByteIndexDefect` is true. The checked-add, remainder-by-zero and `defect` terminals instead return a hard Go error, "reached a terminal outcome with live local resources". A well-formed program that natively yields exit 65 has no interpreter Execution, so the differential cannot compare it.

The conservative refusal is not a soundness hole. It is a correctness gap against the stated Stage-A and Stage-B equivalence.

**Fix:** Accept live-owner defect terminals for every defect kind raised by an abort-like operation (`OpBytesAt`, `OpAddChecked`, `OpRemainderU64`, `OpDefect`). Keep refusing live resources on normal return and on typed failure. Or document the limitation and add a test pinning the error.

### WR-04: New law-disable switches and a C-source rewrite hook live in production packages

**File:** `internal/compiler/corevalidate/corevalidate.go:202-215, 2608` (4 `disable...ForTest` vars), `internal/compiler/originvalidate/originvalidate.go:1447-1456`, `internal/compiler/check/check.go:4247` and `internal/compiler/pathoracle/pathoracle.go:1409`, `internal/compiler/native/native_app.go:117-121`
**Issue:** Phase 29 adds about eight package-level `var disable...ForTest bool` switches to non-test files. They turn off safety laws in the shipped validators. A new function hook, `applicationCSourceMutationForTest`, rewrites emitted C inside `BuildApplication` before digesting. `interp.RunByteViewForTest` is an exported non-test entry point.

- Each is read unsynchronized in production code. A parallel test that arms one races with every other test in the package under `-race`.
- `export_phase29_test.go` documents "Callers must not run in parallel", which is a convention, not enforcement.
- The `disable...ForTest` pattern already exists widely in corevalidate and check, so this is consistent with the codebase. Phase 29 adds about eight more, plus a function hook that mutates emitted C inside `BuildApplication`. If one is ever armed outside tests, the "independent peers" guarantee silently drops.

**Fix:** Move the seams behind a test-only build mechanism, or at least keep the setters in `export_test.go` and unexport the vars. A cheap alternative is a test-package-level guard that fails if any switch is non-default after the test run.

## Info

### IN-01: Fixture admission code in check is unreachable defense

**File:** `internal/compiler/check/check.go:4247-4326` (`checkFixtureOwnerPostAssembly`)
**Issue:** `checkFixtureBytesEntry` always assembles the same five-op, block-free core. The release-while-view-live law, the cycle-refusal branch and the `shaped` check therefore cannot fire on a checker-assembled core. They fire only on hand-built cores in tests. The `wantKinds` check is order-sensitive, so the seeded fault of disabling only the law is redundant with it. The comment's "decision point over the hand-built fixture core" understates that production never reaches these branches.
**Fix:** State this in the comment, or exercise the law from a parsed program shape that actually varies.

### IN-02: Check's scalar fixed point has no operand-type law for byte ops

**File:** `internal/compiler/check/scalar.go:239-249`
**Issue:** The fixed point only checks initialization of the source and index. The type law lives in the builder (`byte_len`/`byte_at` at line 658), and corevalidate and originvalidate each re-derive it. I confirmed that the builder refuses a non-view source (`type.scalar_operand`) and a view used as an index. Both validators also gate shape, so forged cores are refused before interp or cgen in the session path.

cgen itself performs no shape check on `OpBytesAt` (`cgen_program.go:2734-2748`). A forged core handed straight to `EmitApplication`, which bypasses session, would emit `.length` on a non-view local, or `if ( >= x.length)` for an empty `RightID`. That fails under `-Werror` rather than silently misbehaving.
**Fix:** Optional hardening: have cgen refuse a byte op whose source or index place is not the expected type, as it already does for the fixture shape.

### IN-03: The fixture lifecycle pin lives only in straight-line replay and flat-order laws assume program order

**File:** `internal/compiler/corevalidate/corevalidate.go:2985-2990, 2611-2661`
**Issue:** `fixtureOwnerLifecycleValid` is invoked only at the end of `replayStraightLine`. `replayBlocks` has no equivalent. The release-while-view-live laws in corevalidate, check and pathoracle use the flat `Operations` order, not CFG order.

All three peers, interp and cgen refuse fixture shapes that have blocks, so no product path reaches this today. The laws would be unsound for a future block-bearing fixture shape.
**Fix:** Add a comment at each law that they assume a straight-line body, or add an explicit `len(Blocks) == 0` precondition that refuses rather than skips.

### IN-04: The shipped `adler32_fixture.schway` check-passes but cannot be run through `run` or `evidence`

**File:** `examples/phase29/adler32_fixture.schway`
**Issue:** `schway check` reports pass. `schway run --engine=interpreter` and `--engine=native` fail with generic `tool.run_failed`/`native.tool_failure`. `schway evidence` fails with `evidence.canonical_unstable`. `build` without `--manifest` reports "local-owner application requires one explicit binding manifest". The README documents only the `build` plus `app run` route, so a user trying `run` gets opaque failures.
**Fix:** Add a sentence to the README that `run` and `evidence` do not support the fixture owner. Better, have `run` emit a typed refusal.

---

_Reviewed: 2026-10-11_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
