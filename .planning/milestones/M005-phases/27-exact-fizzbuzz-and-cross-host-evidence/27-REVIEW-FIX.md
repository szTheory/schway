---
phase: 27-exact-fizzbuzz-and-cross-host-evidence
fixed_at: 2026-10-10T00:00:00Z
review_path: .planning/milestones/M005-phases/27-exact-fizzbuzz-and-cross-host-evidence/27-REVIEW.md
iteration: 1
findings_in_scope: 5
fixed: 5
skipped: 0
status: all_fixed
---

# Phase 27: Code Review Fix Report

**Fixed at:** 2026-10-10
**Source review:** .planning/milestones/M005-phases/27-exact-fizzbuzz-and-cross-host-evidence/27-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope: 5
- Fixed: 5 (WR-03 and WR-05 partially, see notes)
- Skipped: 0

Edits and commits were made in the main checkout on `phase26-validation`, not in a temp worktree, because the caller forbade creating branches.

## Fixed Issues

### WR-01: C application keeps executing after an output-limit failure; the interpreter stops

**Files modified:** `internal/compiler/cgen/cgen.go`, `internal/compiler/cgen/cgen_phase27_output_test.go`
**Commit:** b96f1b7
**Applied fix:** `schway_application_output_failure()` now flushes stdout, prints the fixed stderr line and calls `exit(65)`. This matches the interpreter's terminal defect. New test `TestPhase27OutputFailureStopsExecution` uses a program that writes and then hits a later `defect`. Before the fix it printed a second stderr line; now only the limit line appears. Status: fixed, requires human verification (control-flow semantics).

### WR-02: Evidence capture reports `returned/complete` when the app already failed on output limit

**Files modified:** `internal/compiler/session/session_phase27_output_test.go`
**Commit:** 53cf8c1
**Applied fix:** Resolved by WR-01. The process exits before the capture is written, so the report is `incomplete` with exit 65 preserved. `TestPhase27OutputEvidenceSeparation` was updated to expect `native.evidence_incomplete`, no Execution and not Verified, in place of the old complete/returned expectation.

### WR-03: `expectApplication` loosens U64 application validation

**Files modified:** `internal/compiler/native/native_app_test.go`
**Commit:** f815e37
**Applied fix:** Partial. Threading the entry return type into the decoder would change the public `RunApplicationWithEvidence` API, and callers only hold an artifact path. The review's alternative, coupling value to return-event shape, was already enforced by `validateExecution`. New test `TestPhase27ApplicationReturnShapeCoupling` pins it: an empty value with a source-place return event, and a non-empty value with a Unit-form event, are both rejected. Residual: a forged Unit return for a U64 entry is not detectable without the type.

### WR-04: Formatter splits multi-operator `let` bindings across lines

**Files modified:** `internal/compiler/syntax/format.go`, `internal/compiler/syntax/syntax_test.go`
**Commit:** 5fbc09d
**Applied fix:** Added `isBinaryOperatorToken` (`+`, `<`, `==`, `%`). A `let` binding now ends its line only after the final operand, and `<` is spaced inside a `let`. Added `TestFormatLetBinaryExpressions`, which covers the five reviewed shapes, checks idempotence and checks the AST is unchanged. It fails without the fix. Status: fixed, requires human verification (formatter logic).

### WR-05: Receipt gate accepts forgeable plain-text logs

**Files modified:** `scripts/check-phase27-receipts.sh`, `internal/compiler/session/session_phase27_host_test.go`
**Commit:** 9214bf9
**Applied fix:** Partial. The gate now requires exactly one well-formed `toolchain go="..." clang="..."` row and rejects any `phase27 ` row of an unknown kind. The header comment states that the gate checks consistency, not authenticity. Three new mutation cases were added to `TestPhase27HostReceiptGate`. The recorded receipts still pass at revision fa5a169.
**Residual risk:** Binding logs to a hosted run (run_id, runner_os, `gh run view` or attestation) needs hosted infrastructure. I did not add a mechanism for it. A hand-written log with the right hashes and revision still passes.

## Skipped Issues

None.

## Verification

Run in the main checkout with `GOCACHE=/tmp/ai-lang-verification-gocache`:
- `go test ./internal/compiler/cgen -run 'TestPhase27|Application'`: ok
- `go test ./internal/compiler/native -run 'TestPhase27|TestPhase22Evidence'`: ok
- `go test ./internal/compiler/syntax`: ok
- `go test ./internal/compiler/session -run 'Phase27|Format'`: ok
- `go test ./cmd/schway -run 'Format|Phase27'`: ok
- `go build ./...` and `go vet` on the touched packages: ok

## Stale hosted evidence

The source changes make the recorded hosted evidence stale: `27-HOST-EVIDENCE.md`, `host-macos.receipt` and `host-linux.receipt`. It describes hosted run 38052601527 at fa5a169. It was not edited. A new hosted run is needed to re-establish cross-host evidence for the fixed revision.

`.planning/config.json` has pre-existing local modifications. It was not staged or committed.

---

_Fixed: 2026-10-10_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
