---
phase: 14-evidence-instrument-and-honest-scoping
reviewed: 2026-09-18T00:00:00Z
depth: standard
files_reviewed: 39
files_reviewed_list:
  - cmd/lang-repair/repair.go
  - cmd/lang-repair/repair_diagnosis_test.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/check/diagnostic_distinctness_test.go
  - internal/compiler/native/native_lto_test.go
  - internal/compiler/native/symbols_test.go
  - internal/compiler/session/evidence_grade_test.go
  - internal/compiler/session/qlt02_budget_manifest.json
  - internal/compiler/session/self_describing_docs_test.go
  - internal/compiler/session/session_payload_replay_test.go
  - internal/compiler/session/session_phase11_differential_test.go
  - internal/compiler/session/session_phase5.go
  - internal/compiler/session/session_phase5_alias.go
  - internal/compiler/session/session_phase5_alias_test.go
  - internal/compiler/session/session_phase5_corpus_test.go
  - internal/compiler/session/session_phase5_test.go
  - internal/compiler/session/session_phase6_budget.go
  - internal/compiler/session/session_phase6_evidence_test.go
  - internal/compiler/session/session_phase6_injectors_test.go
  - internal/compiler/session/session_test.go
  - internal/compiler/session/verification_groundedness_test.go
  - internal/compiler/session/witness_registry_test.go
  - internal/compiler/syntax/parser.go
  - internal/compiler/syntax/parser_skipped_region_test.go
  - scripts/assert-reconciliation-touched.sh
  - scripts/evidence-run-record.sh
  - testdata/distinctness/README.md
  - testdata/distinctness/arithmetic.lang
  - testdata/distinctness/collision_control.json
  - testdata/distinctness/dangling_pipe.lang
  - testdata/distinctness/empty_file.lang
  - testdata/distinctness/for_range.lang
  - testdata/distinctness/lone_else.lang
  - testdata/distinctness/one_token_hard_case.lang
  - testdata/distinctness/spiral_bare.lang
  - testdata/distinctness/spiral_full.lang
  - testdata/distinctness/spiral_narrow.lang
  - testdata/distinctness/stray_semicolon.lang
  - testdata/distinctness/unclosed_brace.lang
  - testdata/phase14/blame_unreachable_admission_refusal.lang
  - testdata/phase14/multi_function_match_refusal.lang
findings:
  critical: 0
  warning: 2
  info: 2
  total: 4
status: issues_found
---

# Phase 14: Code Review Report

**Reviewed:** 2026-09-18T00:00:00Z
**Depth:** standard
**Files Reviewed:** 39
**Status:** issues_found

## Summary

This phase's actual new production-code surface is small and concentrated:
`cmd/lang-repair/repair.go`'s decline classification (`classifyDecline`,
`declineDecision`, the `Decline*` vocabulary — D-14-42), and
`internal/compiler/syntax/parser.go`'s `skipped_region` diagnostic cause
(`recoverRegion`, the widened `problem` signature — DX-08). The remaining
non-test files (`session_phase5.go`, `session_phase5_alias.go`,
`session_phase6_budget.go`) show as large diffs only because `diff_base`
predates most of their history; their content is largely unchanged prior-
phase work and was read for context, not flagged as new-phase risk. Test
files were scanned for reliability defects (flaky patterns, silent
skip-without-reason, missing assertions after fallible calls) rather than
re-litigated line-by-line; none were found — this codebase already carries
unusually heavy adversarial self-testing (frozen collision captures, fault-
injection seams, non-inertness proofs) that a normal project would lack. A
prior review round's WR-01/WR-02/WR-03 findings are already closed by plans
14-11/14-12/14-13 (visible as doc-comment call-outs in
`evidence_grade_test.go`) and are not re-raised here.

The findings below are new, and are genuine — not restatements of the prior
round.

## Warnings

### WR-01: Subprocess timeout is indistinguishable from a legitimate nonzero exit, losing diagnosability on a hang

**File:** `cmd/lang-repair/repair.go:242-265`
**Issue:** `runLangCheck` wraps every `lang --json check` invocation in a
30-second `context.WithTimeout` (`subprocessTimeout`), but the timeout path
is never distinguished from a normal nonzero-exit `StatusInvalid` run.
When `ctx` fires, Go's `exec.CommandContext` kills the process and
`cmd.Run()` typically returns an `*exec.ExitError` for the resulting
signal-killed process — which satisfies `errors.As(runErr, &exitErr)` and
is treated as "the subprocess legitimately exited nonzero" (per the
function's own doc comment). Execution then falls through to
`json.Unmarshal(stdout.bytes(), &result)` against whatever partial JSON was
captured before the kill. Best case this produces `CodeDecodeFailed` with a
generic "decoding lang --json check output" message; worst case (if the
subprocess happened to have already flushed a complete-looking but stale
JSON document before being killed) it could be misread as a genuine
`checkResult`. Either way, an operator debugging a hang has no signal that
the subprocess was killed by the driver's own deadline rather than exiting
on its own.
**Fix:** Check `ctx.Err()` after `cmd.Run()` returns and report a
dedicated code (e.g. `CodeSubprocessTimeout`) when `ctx.Err() ==
context.DeadlineExceeded`, before falling through to the generic
`exitErr`/decode-failure paths:
```go
runErr := cmd.Run()
if ctx.Err() == context.DeadlineExceeded {
    return checkResult{}, &DriverError{Code: CodeSubprocessTimeout, Err: fmt.Errorf("lang check exceeded %s (stderr: %s)", subprocessTimeout, stderr.bytes())}
}
```

### WR-02: `EvaluateBudget`'s stage-work-delta parameter is threaded through but never asserted end-to-end against a real regression

**File:** `internal/compiler/session/session_phase6_budget.go:377-402`
**Issue:** `EvaluateBudget` accepts `stageDeltas []protocol.StageTiming` and,
on a flagged (non-blocking) observation, computes `stageWorkDelta` and
attaches it to the returned `Observation.StageDelta`/`Citable`. For the
gate-eligible branch (`qlt02GateEligibleMetricSet()[row.Metric]`), however,
a blocking verdict (`value > row.ValueOrBound`) is returned as `(measure.
VerdictBlocking, nil)` — `stageDeltas` is silently dropped and no
`Observation` is ever constructed for the one case (a hard-gate regression)
where citing the responsible stage's work-count delta would matter most for
a human triaging a blocking failure. `BudgetRegressionDiagnostic` (same
file) also only ever renders `laneID`/`row`/`value`, never a stage
breakdown. This is not a defect in the sense of "wrong verdict" — the
blocking classification itself is correct — but it means the one artifact
most likely to accompany a build-breaking regression (a hard-gated
`recomputed_work` overshoot) carries strictly less diagnostic context than
a soft, non-blocking `Observation` does.
**Fix:** Either drop the unused `stageDeltas` parameter from the blocking
branch's call sites (since it's dead weight there), or extend the blocking
return to also carry a stage-delta-annotated diagnostic:
```go
if qlt02GateEligibleMetricSet()[row.Metric] {
    if value > row.ValueOrBound {
        delta, citable := stageWorkDelta(stageDeltas)
        return measure.VerdictBlocking, &Observation{
            Metric: row.Metric, Bound: row.ValueOrBound, Value: value,
            StageDelta: delta, Citable: citable,
        }
    }
    return measure.VerdictObserved, nil
}
```

## Info

### IN-01: Hand-rolled insertion sort duplicates `sort.Strings`

**File:** `internal/compiler/session/evidence_grade_test.go:1072-1078`
**Issue:** `sortStrings` reimplements an O(n²) insertion sort over
`[]string` purely to avoid importing `sort` in this file, even though
`sort.Strings` is already used for the identical purpose elsewhere in the
same package (e.g. `internal/compiler/check/diagnostic_distinctness_test.go`,
`sort.Strings(diagIDs)`). This is a maintainability smell — a second,
independently-maintained sort implementation with no test of its own,
sitting in a file already over 1800 lines.
**Fix:**
```go
import "sort"
// ... replace the sortStrings body with:
sort.Strings(matched)
```
and delete the `sortStrings` function, updating its one call site in
`resolvedPkgPatterns`.

### IN-02: `classifyDecline`'s "no diagnostics" branch reuses the decline-reason constant as the diagnosis code, collapsing two distinct signals into one string

**File:** `cmd/lang-repair/repair.go:376-385`
**Issue:** When `len(codes) == 0` (an invalid program reporting zero
diagnostics — the documented "fail-closed anomaly" case), both
`decision.declineReason` and `decision.diagnosisCode` are set to the same
literal, `DeclineNoDiagnostics` ("repair.no_diagnostics"). This is
deliberate per the inline comment ("the decline reason itself is the only
honest non-empty value available here"), and it is tested
(`TestDeclineCarriesReasonAndDiagnosis/no_diagnostics`), so it is not a
correctness bug. It is, however, a slightly awkward protocol shape: a
consumer reading `Outcome.DiagnosisCode` on this one path gets a
decline-reason-shaped string (`"repair.no_diagnostics"`) rather than
anything resembling the closed diagnostic-code vocabulary
(`"syntax.expected_rbrace"`-shaped) every other path populates it with —
a downstream consumer pattern-matching on `DiagnosisCode`'s shape (e.g.
"does it look like `<namespace>.<code>`?") would silently misclassify this
one case as a real diagnostic code from an unfamiliar namespace rather than
recognizing it as "there was no diagnostic at all."
**Fix:** Consider a distinct sentinel (e.g. `"repair.diagnosis_unavailable"`)
for `diagnosisCode` in this branch, kept different from `declineReason`, so
a consumer can tell "no diagnostic existed" apart from "the diagnostic
happened to be one named after the decline reason" — purely a protocol-
clarity improvement, not a behavior change.

---

_Reviewed: 2026-09-18T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
