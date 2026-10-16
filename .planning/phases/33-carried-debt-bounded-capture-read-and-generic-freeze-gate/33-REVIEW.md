---
phase: 33-carried-debt-bounded-capture-read-and-generic-freeze-gate
reviewed: 2026-10-16T02:51:13Z
depth: standard
files_reviewed: 30
files_reviewed_list:
  - cmd/schway/main_phase321_run_input_test.go
  - cmd/schway/main_phase32_routes_test.go
  - internal/compiler/cgen/cgen_phase33_scalar_defect_test.go
  - internal/compiler/cgen/cgen_phase36_caller_shapes_test.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/execution/defect_reasons.go
  - internal/compiler/executionpeer/executionpeer.go
  - internal/compiler/executionpeer/executionpeer_phase33_terminal_test.go
  - internal/compiler/executionpeer/executionpeer_phase35_test.go
  - internal/compiler/executionpeer/executionpeer_test.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interp_phase321_defect_test.go
  - internal/compiler/interp/interp_phase33_defect_test.go
  - internal/compiler/native/native_app.go
  - internal/compiler/native/native_phase33_capture_test.go
  - internal/compiler/native/native_phase33_defect_parity_test.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_freeze_gate_test.go
  - internal/compiler/session/session_gate_differential_test.go
  - internal/compiler/session/session_gate_retire_test.go
  - internal/compiler/session/session_phase32_receipt_test.go
  - internal/compiler/session/session_phase37_validation_test.go
  - internal/compiler/session/session_receipt_gate_test.go
  - internal/compiler/session/summary_coverage_test.go
  - scripts/check-freeze.sh
  - scripts/check-receipts.sh
  - scripts/codedigest.sh
  - scripts/verify-phase32.sh
  - testdata/phase16/public-emitter-consumers.json
  - testdata/receipts/M006.manifest
findings:
  critical: 3
  warning: 0
  info: 0
  total: 3
status: issues_found
---

# Phase 33: Code Review Report

**Reviewed:** 2026-10-16T02:51:13Z
**Depth:** standard
**Files Reviewed:** 30 cumulative unique paths
**Status:** issues_found

## Summary

The initial standard-depth review covered the explicit 25-file scope at source head `5f1554eab21187cdb8be5fdc507c84d66e40b23b`. The dated closure pass separately reviewed four changed execution-peer files; `executionpeer.go` overlaps the initial scope, yielding 28 distinct paths across the two reviews. This is cumulative coverage, not a fresh review of all 28 files at the closure timestamp. Those review passes covered bounded application capture reads, interpreter defect and ownership handling, execution-peer admission, generated C invocation evidence, and the generic freeze and receipt gates; no tests were run during those passes. Later scoped test execution is recorded in the dated refresh sections below. The parent reports that its separate current validation passed 18 mapped rows, 49 roots, and six packages; that is parent evidence, not a test run by this review.

The historical findings remain visible here: the initial 2026-10-15 review reported CR-01 (empty manifest controls) and WR-01/WR-02 (incomplete coverage verification records and undeclared receipt case programs). Its 19:52 closure pass marked WR-01/WR-02 fixed and reported WR-03/WR-04 (empty case program and incomplete coverage entry metadata). The Phase 33 review-fix record later marked all five findings fixed. This review rechecked the current source: the receipt gate rejects missing/malformed controls and malformed case rows, checks the case-program inventory in both directions, and the coverage checker validates per-item metadata plus nonempty, unique entry IDs and descriptions. These historical findings are excluded from the current count; their dated history and closure state are retained rather than treated as an issue-free past review.

The incremental refresh at source head `639b4965d7f9c71fca8088c723ec9fac32725e6e` reviewed exactly three files: `internal/compiler/cgen/cgen_phase36_caller_shapes_test.go`, `internal/compiler/cgen/cgen_program.go`, and `testdata/phase16/public-emitter-consumers.json`. The canonical resolver returned usable `degraded` phase-range scope because there were no phase-task commits since `aad275f42d7358f49fdc8f63ab27adf2b4f63f66`. The latter two paths were already in the cumulative inventory; the new caller-shape test raises cumulative coverage from 28 to 29 unique paths. This is an incremental review of those three files, not a fresh reread of all 29. The added manifest row matches the direct `EmitNative` call and its enclosing witness. The review found a native defect-terminal attribution failure on the newly admitted scalar caller path, recorded as CR-05 below.

The subsequent closure refresh reviewed exactly three files: `internal/compiler/cgen/cgen_program.go`, `internal/compiler/cgen/cgen_phase33_scalar_defect_test.go`, and `testdata/phase16/public-emitter-consumers.json`. The latter two paths from the prior refresh overlap this scope, so the new test adds one unique path and cumulative coverage is now 30. It corrects the initial CR-05 reproducer claim: the Byte explicit-defect source was checker-refused, and the ByteView Unit test did not execute that body. The confirmed checker-admitted reproducer is a U64 caller that consumes its input and defects at entry or in a nested scalar helper. The native failure was a C compile error from an application-only capacity-capture helper, not an observed exit-65 run. CR-05 is closed for dynamic-invocation scalar callers by the final guarded emitter path. The guard preserves the legacy no-call byte baseline, leaving a separate native defect-parity gap recorded as CR-06 below.

## Narrative Findings (AI reviewer)

### CR-02: Execution peer accepts terminal evidence that contradicts the reported outcome [CLOSED]

**File:** internal/compiler/executionpeer/executionpeer.go:665-688,699-724
**Severity:** BLOCKER

**Issue:** The validator classifies ordinary return, failure, and user-defect events by operation ID and event kind, then closes the activation. It does not verify that the root terminal event agrees with Execution.Outcome.Kind, or that a matched return/failure event's source place and type agree with the core operation. For example, a valid single-function function.returned trace with Outcome.Kind set to defect or typed_failure drains the stack and reaches the final nil return. A user-defect event can also carry an Output different from the matched OpDefect.Reason. The independent peer therefore accepts contradictory or forged terminal evidence.

**Fix:** Bind the final root terminal kind to Outcome.Kind (returned, typed failure, or defect), and check terminal source place, type, and user-defect reason against the matched core operation. Keep numeric outcome-value replay outside this peer's stated responsibility.

### CR-03: Legacy child defects leave the caller activation open

**File:** internal/compiler/executionpeer/executionpeer.go:693-700
**Severity:** BLOCKER

**Issue:** A `function.defected` event from a child is a whole-execution terminal, but the validator clears the full activation stack only when `dynamicCallIdentity` is true. In the bounded legacy path it closes and pops the child alone, leaves its caller on `stack`, then rejects the valid terminal document at the final `len(stack) != 0` check. This disagrees with the interpreter's defect-unwinds-the-whole-execution behavior and makes nested defects in statically bounded call traces inadmissible.

**Fix:** On any accepted `function.defected` terminal, mark every currently active invocation closed and clear the activation stack independent of call-identity mode. Keep ordinary child returns as one-frame pops.

### CR-04: Propagated typed failure from OpReturn is unclassifiable

**File:** internal/compiler/executionpeer/executionpeer.go:721-741
**Severity:** BLOCKER

**Issue:** When a caller propagates a modeled failure through its `OpReturn`, the interpreter emits `function.failed` with ID `<return-op>:event:failed`. `classify` accepts `function.failed` only for an `OpFail` operation, and accepts an `OpReturn` only with `function.returned`, so the validator rejects this supported error-edge execution as `unknown_kind` before applying terminal checks. The new attribution check also assumes every matched `OpReturn` is a returned event, while the interpreter can attach the propagated failure type to this terminal.

**Fix:** Classify the documented `OpReturn` failure-event form and validate its source place and failure type using the modeled-failure contract; when it is the root terminal, require `OutcomeTypedFailure`. Preserve child-failure continuations that the caller handles on its error edge.

### CR-05: Dynamic native scalar callers emit application-only output for explicit defects [CLOSED]

**File:** internal/compiler/cgen/cgen_program.go:3007-3014,3027-3031
**Severity:** BLOCKER

**Issue:** Before repair, the scalar emitter selected `emitProgramApplicationScalarFailure` for `OpDefect` even when emitting the native shell. With dynamic invocation enabled, that application handler referenced `schway_write_capacity_capture`, which is emitted only for the application evidence shell. The final checker-admitted U64 entry and nested-helper reproducers consume their input and then reach an explicit defect; their generated native C therefore failed compilation in all six O0/O3/O3-LTO entry/helper cases. The application controls passed. The earlier Byte explicit-defect reproduction was refused by the checker, and the ByteView Unit case did not execute that body; neither supports a Byte/ByteView defect-admission claim. The current guard sends dynamic native callers through the structured defect terminal and leaves application behavior unchanged; the remaining no-call case is tracked separately as CR-06.

**Fix:** The emitter now uses the application stderr/exit path for `programApplicationShell` and for legacy emission without dynamic invocation support, and uses `emitProgramDefectTerminal(out, function, op)` followed by `exit(0)` for dynamic native emission. The latter matches the adjacent native defect paths and passed the pinned U64 entry/nested-helper regression at `cgen_program.go:3008-3013`.

### CR-06: Legacy no-call native scalar defects still exit as application failures

**File:** internal/compiler/cgen/cgen_program.go:3007-3013,3027-3031
**Severity:** BLOCKER

**Issue:** When `dynamicInvocationSupport` is false, the `OpDefect` arm routes emission through `emitProgramApplicationScalarFailure` even for the native shell. A checker-admitted no-call scalar program that reaches an explicit defect therefore writes `schway app: scalar caller stopped` to stderr and exits 65 without a structured `function.defected` terminal; the native runner reports `native.run_failed`, so native parity and peer attribution cannot be established. The temporary overlay that removes only the helper call from the U64 reproducer fails all three native O0/O3/O3-LTO lanes on current source with this result. The same overlay against original source `639b4965` produces the same three failures, while application controls pass. These are intentionally failing diagnostic probes, not green validation. The initial broad native branch also changed eight frozen-C no-call baseline/control roots; preserving those bytes does not satisfy the separate native defect contract.

**Fix:** Add a native structured-terminal path for no-call `OpDefect` emission while preserving and rerunning the frozen no-call byte-baseline/control evidence. The emitter maintainer and M007 auditor must resolve this scoped gap before Phase38 freeze; if the correction changes canonical C bytes, record a dated compatibility amendment and refresh the baseline under that explicit decision. Do not implicitly waive either requirement.

## Dated Closure Review — 2026-10-16T00:35:09Z

**Closure scope:** exactly `internal/compiler/executionpeer/executionpeer.go`, `internal/compiler/executionpeer/executionpeer_phase33_terminal_test.go`, `internal/compiler/executionpeer/executionpeer_test.go`, and `internal/compiler/executionpeer/executionpeer_phase35_test.go`. This pass inspected the working-tree diff against HEAD and did not run tests or background processes.

**CR-02:** Closed by independent source inspection. The validator now checks the matched `OpReturn`, `OpFail`, and `OpDefect` source/type attribution and rejects terminal targets; it binds user-defect output to `OpDefect.Reason`; it checks root terminal outcome kinds and requires an empty defect value. The root/defect condition deliberately leaves child returns independent of the eventual root outcome, and no numeric returned or failed value replay was introduced. The unconditional `closedActivations` check retains the Phase 36 repeated-terminal diagnostic precedence. Bare match-return classification still bypasses linear-operation attribution, so the new checks do not reject those existing match traces.

The new named terminal test covers positive OpReturn/OpFail/OpDefect cases, source/type/target mutations, each alternate terminal outcome, defect reason/value mutations, a post-terminal event, and acceptance of a structurally plausible nonmatching numeric value for return/failure. Parent-reported evidence says a temporary overlay reproduced the original false accepts, the named test killed the pre-fix peer source with exit 1 and 30 failing test nodes, a focused control passed, all 18 current execution-peer roots passed before the final occurrence-1 post-terminal refinement, and the final two-root terminal/call-control run passed after that refinement. The focused inherited regression was still running; no result is claimed for it. This is parent evidence and was not reproduced by this review.

The initial 25-file scope plus these four closure files is 28 unique paths: `executionpeer.go` is the sole overlap with the initial list. The frontmatter inventory is cumulative; only the four closure paths were reread in this pass. CR-03 and CR-04 are newly recorded remaining findings from the requested propagation interaction check.

---

_Reviewed: 2026-10-16T00:35:09Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_

## Dated Incremental Refresh — 2026-10-16T02:38:22Z

**Refresh scope:** exactly `internal/compiler/cgen/cgen_phase36_caller_shapes_test.go`, `internal/compiler/cgen/cgen_program.go`, and `testdata/phase16/public-emitter-consumers.json`. The resolver status was `degraded` (`phase-range`, `no-task-commits-since`) with range base `aad275f42d7358f49fdc8f63ab27adf2b4f63f66`; the review used its three-file result. Two paths overlap the prior 28-path inventory, so this adds one unique path and brings cumulative coverage to 29. No source or test files were modified by this reviewer, and no tests were run.

The direct `EmitNative` source call at `cgen_phase36_caller_shapes_test.go:41` agrees with the registry row at `public-emitter-consumers.json:5-7`; the witness names the test containing that call. The new tests cover returned-value parity and translation-unit linking, but the ByteView Unit test deliberately does not execute that body. Source inspection found that admitted dynamic scalar bodies also allow `OpDefect`, while the scalar emitter's `OpDefect` arm always selects the application-only stderr/exit handler. For the native shell this omits the defect event and terminal execution record; CR-05 remains open as a BLOCKER.

The parent supplied `/tmp/schway33-r3-validation-run.json` for source `639b4965d7f9c71fca8088c723ec9fac32725e6e`: 18 validation rows, 49 roots, six packages, and zero missing, failed, or skipped roots. This is parent evidence and was not rerun by this review. No evidence in this three-file pass resolves CR-03 or CR-04: both remain current reviewer-classified BLOCKER false rejections in execution-peer propagation, deferred to the execution-peer maintainer through the mandatory M007 audit before Phase38 freeze. They are not closed, accepted, or relabeled clean.

---

_Reviewed: 2026-10-16T02:38:22Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_

## Dated CR-05 Closure Review — 2026-10-16T02:43:54Z

**Closure scope:** exactly `internal/compiler/cgen/cgen_program.go`, `internal/compiler/cgen/cgen_phase33_scalar_defect_test.go`, and `testdata/phase16/public-emitter-consumers.json`. The emitter and registry were already in the 29-path cumulative inventory; the new test adds one unique path, bringing cumulative coverage to 30. This is an incremental closure review of those three files, not a reread of all 30. The registry row at `public-emitter-consumers.json:5-7` matches the direct `EmitNative` call at `cgen_phase33_scalar_defect_test.go:66` and names its covering test.

The initial CR-05 description is corrected above. The final regression uses checker-admitted U64 input at the direct entry and through a nested `stop` helper, then reaches the pinned reason `scalar caller stopped`. It checks a defect outcome with an empty value, verifies the final `function.defected` event and reason, and validates interpreter and native executions with `executionpeer.Validate`. Each location runs native O0/O3/O3-LTO with `ExpectDefect`; separate application controls check exit 65, empty stdout, and exact stderr. I independently ran `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -run '^TestPhase33ScalarExplicitDefectShells$' -count=1 -timeout=600s`; it passed.

The parent-supplied original-source RED log `/tmp/schway33-r3-explicit-defect-final-red.log` records all six native entry/helper optimization cases failing to compile on the undeclared application-only `schway_write_capacity_capture`; both application controls pass. The parent-supplied repaired GREEN log `/tmp/schway33-r3-explicit-defect-final-green.log` records all cases passing. The five-root cgen safety run in `/tmp/schway33-r3-fix-safety.log` and the two inventory roots in `/tmp/schway33-r3-inventory-final.log` also pass; these logs are parent evidence and were not rerun in this closure. The exploratory Byte, unused-U64, and runner drafts are not credited as RED evidence.

At this closure snapshot, the repaired branch retained application behavior and used `emitProgramDefectTerminal` plus `exit(0)` for dynamic native emission, matching the adjacent checked-add, remainder, and bounds defect paths. The generated terminal record carried the attributed reason and was accepted by the native runner under `ExpectDefect`; the zero exit was transport behavior after the record was flushed. CR-05 was closed for that dynamic caller path. CR-03 and CR-04 remained current reviewer-classified BLOCKER false rejections in execution-peer propagation, deferred to the execution-peer maintainer through the mandatory M007 audit before Phase38 freeze; neither was fixed, accepted, or relabeled clean. The current report at that closure snapshot retained two critical findings and `issues_found` status.

---

_Reviewed: 2026-10-16T02:43:54Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_

## Dated Final-Guard and Residual-Debt Review — 2026-10-16T02:51:13Z

**Refresh scope:** exactly `internal/compiler/cgen/cgen_program.go`, `internal/compiler/cgen/cgen_phase33_scalar_defect_test.go`, and `testdata/phase16/public-emitter-consumers.json`. These paths were already counted in the cumulative inventory; this is a fresh review of three files at source head `28e6cae5`, not a full rereview of the 30 unique paths. The registry row still matches the test's direct `EmitNative` call. No source, test, or registry file was changed by this reviewer.

The final guard at `cgen_program.go:3008-3013` keeps the application handler for application emission and for the legacy no-call native shape, while dynamic native emission writes the attributed defect terminal and exits through the native transport path. The previously recorded U64 entry and nested-helper defect regression remains valid; I reran `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -run '^TestPhase33ScalarExplicitDefectShells$' -count=1 -timeout=600s`, which passed. This closes CR-05 for the dynamic scalar-caller failure.

The no-call compatibility constraint required a narrower repair. The earlier broad branch at `4b9d2ff8` failed eight frozen-C no-call baseline/control roots in `/tmp/schway33-r3-regression-13.log`; this is preserved as a failed regression, not green credit. Parent-supplied final evidence in `/tmp/schway33-r3-fix-final-safety.log` reports five cgen safety roots passing, and `/tmp/schway33-r3-nocall-final.log` reports all eight no-call baseline/control roots passing. I did not rerun those parent-supplied roots.

The remaining no-call behavior is independently attributable from the guarded branch and is also captured by the parent-supplied identical temporary overlay probes: `/tmp/schway33-r3-nocall-debt.log` fails the three native optimization lanes on current source with exit 65 and the application stderr; `/tmp/schway33-r3-nocall-original-debt.log` fails the same lanes against original `639b4965`; application controls pass in both. These intentionally failing diagnostics are not included as passing validation. CR-06 records this current no-call parity/attribution blocker for the emitter maintainer and mandatory M007 audit before Phase38 freeze, with a dated compatibility amendment required if preserving the byte baseline conflicts with the terminal fix. CR-03 and CR-04 remain current execution-peer false-rejection BLOCKERs, deferred to their separate execution-peer maintainer through M007; none of these three blockers is accepted, closed, or relabeled clean.

This pass adds no source paths to the inventory: cumulative coverage remains 30 unique paths, while the current incremental scope remains exactly three files. The report therefore has three current critical findings (CR-03, CR-04, CR-06), no warnings or info findings, and `issues_found` status.

---

_Reviewed: 2026-10-16T02:51:13Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_
