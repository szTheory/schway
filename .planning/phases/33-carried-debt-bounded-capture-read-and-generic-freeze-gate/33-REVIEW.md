---
phase: 33-carried-debt-bounded-capture-read-and-generic-freeze-gate
reviewed: 2026-10-16T00:35:09Z
depth: standard
files_reviewed: 28
files_reviewed_list:
  - cmd/schway/main_phase321_run_input_test.go
  - cmd/schway/main_phase32_routes_test.go
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
  critical: 2
  warning: 0
  info: 0
  total: 2
status: issues_found
---

# Phase 33: Code Review Report

**Reviewed:** 2026-10-16T00:35:09Z
**Depth:** standard
**Files Reviewed:** 28 cumulative unique paths
**Status:** issues_found

## Summary

The initial standard-depth review covered the explicit 25-file scope at source head `5f1554eab21187cdb8be5fdc507c84d66e40b23b`. The dated closure pass separately reviewed four changed execution-peer files; `executionpeer.go` overlaps the initial scope, yielding 28 distinct paths across the two reviews. This is cumulative coverage, not a fresh review of all 28 files at the closure timestamp. The review covered bounded application capture reads, interpreter defect and ownership handling, execution-peer admission, generated C invocation evidence, and the generic freeze and receipt gates. No tests were run by this reviewer. The parent reports that its separate current validation passed 18 mapped rows, 49 roots, and six packages; that is parent evidence, not a test run by this review.

The historical findings remain visible here: the initial 2026-10-15 review reported CR-01 (empty manifest controls) and WR-01/WR-02 (incomplete coverage verification records and undeclared receipt case programs). Its 19:52 closure pass marked WR-01/WR-02 fixed and reported WR-03/WR-04 (empty case program and incomplete coverage entry metadata). The Phase 33 review-fix record later marked all five findings fixed. This review rechecked the current source: the receipt gate rejects missing/malformed controls and malformed case rows, checks the case-program inventory in both directions, and the coverage checker validates per-item metadata plus nonempty, unique entry IDs and descriptions. These historical findings are excluded from the current count; their dated history and closure state are retained rather than treated as an issue-free past review.

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

## Dated Closure Review — 2026-10-16T00:35:09Z

**Closure scope:** exactly `internal/compiler/executionpeer/executionpeer.go`, `internal/compiler/executionpeer/executionpeer_phase33_terminal_test.go`, `internal/compiler/executionpeer/executionpeer_test.go`, and `internal/compiler/executionpeer/executionpeer_phase35_test.go`. This pass inspected the working-tree diff against HEAD and did not run tests or background processes.

**CR-02:** Closed by independent source inspection. The validator now checks the matched `OpReturn`, `OpFail`, and `OpDefect` source/type attribution and rejects terminal targets; it binds user-defect output to `OpDefect.Reason`; it checks root terminal outcome kinds and requires an empty defect value. The root/defect condition deliberately leaves child returns independent of the eventual root outcome, and no numeric returned or failed value replay was introduced. The unconditional `closedActivations` check retains the Phase 36 repeated-terminal diagnostic precedence. Bare match-return classification still bypasses linear-operation attribution, so the new checks do not reject those existing match traces.

The new named terminal test covers positive OpReturn/OpFail/OpDefect cases, source/type/target mutations, each alternate terminal outcome, defect reason/value mutations, a post-terminal event, and acceptance of a structurally plausible nonmatching numeric value for return/failure. Parent-reported evidence says a temporary overlay reproduced the original false accepts, the named test killed the pre-fix peer source with exit 1 and 30 failing test nodes, a focused control passed, all 18 current execution-peer roots passed before the final occurrence-1 post-terminal refinement, and the final two-root terminal/call-control run passed after that refinement. The focused inherited regression was still running; no result is claimed for it. This is parent evidence and was not reproduced by this review.

The initial 25-file scope plus these four closure files is 28 unique paths: `executionpeer.go` is the sole overlap with the initial list. The frontmatter inventory is cumulative; only the four closure paths were reread in this pass. CR-03 and CR-04 are newly recorded remaining findings from the requested propagation interaction check.

---

_Reviewed: 2026-10-16T00:35:09Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_
