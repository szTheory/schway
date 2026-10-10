---
phase: 27-exact-fizzbuzz-and-cross-host-evidence
reviewed: 2026-10-10T00:00:00Z
depth: standard
files_reviewed: 39
files_reviewed_list:
  - .github/workflows/ci.yml
  - README.md
  - cmd/schway/main_test.go
  - examples/fizzbuzz.schway
  - examples/phase27/README.md
  - examples/phase27/fizzbuzz-1000.out
  - examples/phase27/fizzbuzz-15.out
  - internal/compiler/ast/ast.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_phase27_evidence_test.go
  - internal/compiler/cgen/cgen_phase27_output_test.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/check/scalar.go
  - internal/compiler/core/core.go
  - internal/compiler/core/core_test.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/execution/execution.go
  - internal/compiler/executionpeer/executionpeer.go
  - internal/compiler/executionpeer/executionpeer_phase27_test.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interp_phase27_evidence_test.go
  - internal/compiler/interp/interp_phase27_output_test.go
  - internal/compiler/native/native.go
  - internal/compiler/native/native_app.go
  - internal/compiler/native/native_app_test.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_phase27_docs_test.go
  - internal/compiler/session/session_phase27_evidence_test.go
  - internal/compiler/session/session_phase27_host_test.go
  - internal/compiler/session/session_phase27_output_test.go
  - internal/compiler/session/session_phase27_remainder_test.go
  - internal/compiler/session/session_phase27_test.go
  - internal/compiler/syntax/format.go
  - internal/compiler/syntax/lexer.go
  - internal/compiler/syntax/parser.go
  - internal/compiler/syntax/token.go
  - scripts/check-phase27-receipts.sh
  - scripts/verify-phase27.sh
findings:
  critical: 0
  warning: 5
  info: 2
  total: 7
status: issues_found
---

# Phase 27: Code Review Report

**Reviewed:** 2026-10-10
**Depth:** standard
**Files Reviewed:** 39
**Status:** issues_found

## Summary

I read the remainder, equality and output-write paths in the interpreter, the C emitter, the checker and corevalidate. I also read the parser, lexer and formatter changes, the evidence and native decoding changes, and both shell scripts.

Output-limit accounting agrees across the two engines. Both the interpreter sink and the C `schway_write_bytes` use `count > limit || len > limit - count`. An exact-fit write is allowed, a crossing write is rejected whole, and no partial chunk is emitted. Dynamic-zero remainder is guarded before the C `%` in the emitter, and in the interpreter before the `%`. Both paths exit 65 with a defect.

I found no blockers. The warnings are mostly engine-divergence and robustness gaps around the output-limit failure path and the shell evidence gate. I did not run the focused Go tests. The formatter finding (WR-04) was reproduced with `schway format`.

## Warnings

### WR-01: C application keeps executing after an output-limit failure; the interpreter stops

**File:** `internal/compiler/cgen/cgen.go:735`, `internal/compiler/cgen/cgen_program.go:2398-2402,2684-2688,2736`
**Issue:** `schway_application_output_failure()` only sets a flag and prints once. Control returns to the emitted program, which keeps running to completion. The interpreter returns a terminal `function.defected` at the first failed write (`interp.go`, `OpWriteU64`/`OpWriteNewline`/`OpWriteText` arms).

This produces observable divergence after the limit is crossed:
- A later checked-add overflow or `defect` statement prints a second stderr line (`schway app: ...`). A later dynamic-zero remainder prints a different one (`schway: U64 remainder by zero`). The interpreter reports only the limit defect.
- Loop work after the failure is not bounded. A program like `while i < n { write_u64(i) ... }` with a large `n` spins up to `n` iterations with every write silently failing. The interpreter terminates immediately.

Exit code 65 and stdout bytes still agree, so the current tests will not catch this.
**Fix:** Terminate at the first failure.
```c
static void schway_application_output_failure(void) {
  fputs("schway: application output limit exceeded\n", stderr);
  fflush(stdout);
  exit(65);
}
```
If the evidence capture must still be written on this path, call the evidence writer before `exit`, or at least do not continue user code.

### WR-02: Evidence capture reports `returned/complete` when the app already failed on output limit

**File:** `internal/compiler/cgen/cgen_program.go:1053-1068`
**Issue:** After `schway_application_output_failure()` the generated `main` still writes `{"status":"complete","execution":{"outcome":{"kind":"returned",...}}}` and then returns 65. The interpreter's document for the same run has a `defect` outcome plus a `function.defected` event with reason `application output limit exceeded`.

`TestPhase27EvidenceCapacity` in `native_app_test.go` codifies that "process outcome wins". Consumers that read the execution document alone still see a successful return for a run that exited 65, and the differential evidence disagrees with the interpreter.
**Fix:** When `schway_application_output_failed` is set, write a distinct capture status instead of `complete/returned`, for example a defect execution. Alternatively, skip the evidence file and let the report say `unobserved`. Fixing WR-01 as an immediate exit with no evidence makes this moot.

### WR-03: `expectApplication` loosens U64 application validation

**File:** `internal/compiler/native/native.go:932-935,1008-1011`, `internal/compiler/native/native_app.go:573-576`
**Issue:** `decodeApplicationCapture` now validates with `expectApplication` for every application, not only Unit-returning ones. Unlike `ExpectValue`, it drops the `Outcome.Value != ""` requirement. The return event is validated in the Unit shape (`SourcePlace == ""`, `Output == ""`) whenever the value is empty.

A forged or buggy capture from a U64-returning application can therefore claim `returned` with an empty value and pass decode. The decoder does not know the entry's declared return type.
**Fix:** Thread the entry return type into the decoder (`ExpectValue` for U64, `expectApplication` for Unit). Alternatively, in `expectApplication`, require `Value == ""` and Unit-form return events to hold together, and require a non-empty value with a SourcePlace otherwise.

### WR-04: Formatter splits multi-operator `let` bindings across lines

**File:** `internal/compiler/syntax/format.go:205-212`
**Issue:** The new `%` special case only protects `identifier %`. I reproduced these with `schway format`:
```
let a = n % 3 + 1   ->  let a = n % 3
                         + 1
let b = a == n      ->  let b = a
                         == n
let c = n + 2 % 4   ->  let c = n
                         + 2 % 4
```
The output re-parses and is idempotent, so semantics are preserved. But `==` is a new operator whose canonical form is broken for the most natural use, `let x = a == b`. `+`/`<` behave similarly (`let d = n < 5` gives `n` then `<5`), so this extends an existing bug to the new operators. `format --check` then forces every user to commit the ugly shape. The fixture only uses `let x = a % k` and `if x == 0`, so it does not exercise this.
**Fix:** In the identifier and number branches, only terminate the binding line when the next token is not a binary operator (`TokenPlus`, `TokenLAngle`, `TokenEqualEqual`, `TokenPercent`). Terminate after the final operand. Add formatter golden tests for `let` with `==`, `+`, `<` and mixed precedence.

### WR-05: Receipt gate accepts forgeable plain-text logs

**File:** `scripts/check-phase27-receipts.sh:34-79`
**Issue:** The gate validates only the textual shape of the logs. Anyone can hand-write an identity row, ten case rows, three control rows and a result row with the expected revision and hashes, and it passes. The hashes and revision are recomputed locally, but nothing binds a log to an actual CI run, runner or artifact (no run ID, no attestation, no checksum of the build outputs). The `toolchain` row is not checked at all. `tree=clean` is self-reported.

Additionally, extra non-`phase27` lines are ignored, so output can be interleaved. The "first/last `phase27 ` line" checks only constrain `phase27`-prefixed lines.
**Fix:** Bind receipts to the hosted run. Require and verify a `run_id`/`run_attempt` and `runner_os` in the identity row against `gh run view`, or verify GitHub artifact attestations (`gh attestation verify`). At minimum, document that this gate checks consistency, not authenticity. Also assert a toolchain row exists and the log has no unexpected `phase27 ` rows (reject any `phase27 ` line that is not `host identity|toolchain|case|control|host result`).

## Info

### IN-01: Output-write failure conflates `fwrite` errors with limit exhaustion

**File:** `internal/compiler/cgen/cgen.go:734-735`
**Issue:** `schway_write_bytes` returns 0 both for ceiling overflow and for an `fwrite` failure such as a closed pipe or EPIPE. Both then print "application output limit exceeded" and exit 65. A downstream write error is misreported as a limit breach.
**Fix:** Distinguish the two causes with separate flags or messages. Report an I/O error with its own message and exit status.

### IN-02: Duplicated "reason" literals across layers

**File:** `internal/compiler/interp/interp.go:1043,1415`, `internal/compiler/cgen/cgen_program.go:228,2660`, `internal/compiler/session/session.go:533`, `internal/compiler/executionpeer/executionpeer.go:387`
**Issue:** The strings `"U64 remainder by zero"` and `"application output limit exceeded"` are repeated as bare literals in at least five packages. Only the output-limit one has a constant (`applicationOutputLimitDefectReason`, unexported to interp). Drift between the C stderr text (`schway:` vs `schway app:` prefix) and the others is already visible.
**Fix:** Export both reasons as constants from `internal/compiler/execution` and reference them everywhere.

---

_Reviewed: 2026-10-10_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
