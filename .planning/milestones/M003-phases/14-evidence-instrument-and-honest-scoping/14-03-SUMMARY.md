---
phase: 14-evidence-instrument-and-honest-scoping
plan: 03
subsystem: testing
tags: [lang-repair, diagnostics, evidence, go]

# Dependency graph
requires:
  - phase: 13-agent-loop-for-interprocedural-defects
    provides: check.interprocedural_loan_liveness's move_after_interprocedural_loan repair (backward direction) and the forward-direction unrepairable decline this plan now classifies
provides:
  - "Outcome.DiagnosisCodes / DeclineReason / BestApplicability: additive omitempty fields on cmd/lang-repair's Outcome envelope, populated on every decline"
  - "A closed three-value decline_reason vocabulary (repair.none_offered, repair.none_eligible, repair.no_diagnostics)"
  - "TestUnrepairableAlwaysCarriesDiagnosis: a corpus-wide guard that no unrepairable outcome ever carries an empty diagnosis_code"
  - "TestUnrepairableDiagnosisGuardIsNotInert: a seeded-fault proof the guard can actually go red"
affects: [lang-repair, agent-loop-consumers-of-lang-repair-json]

# Actuals (#2632)
actuals:
  tokens: 6931
  tasks: 3
  commits: 5

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Comparable string-backed list type (diagnosisCodeList) with custom (Un)MarshalJSON, used when a slice field would break an existing struct-equality test in a file the plan must leave byte-unchanged"
    - "Production-visible fault-injection toggle var (diagnosisEmissionDisabledForTest), following the project's own corevalidate.SetDisableCyclePeerForTest / cgen.PayloadSlotSwapInjectedWriteCount precedent, for a same-package non-inertness proof"

key-files:
  created:
    - cmd/lang-repair/repair_diagnosis_test.go
  modified:
    - cmd/lang-repair/repair.go

key-decisions:
  - "selectRepair's three-value signature (jsonRepair, string, bool) is left byte-unchanged; a new sibling classifyDecline carries the decline classification instead, because antitheater_test.go and repair_test.go call selectRepair at several sites and both files (the former by an explicit acceptance criterion) must stay byte-unchanged by this plan"
  - "Outcome.DiagnosisCodes uses a new comparable diagnosisCodeList string type (JSON-array-marshaling via custom MarshalJSON/UnmarshalJSON), not a plain []string, because antitheater_test.go compares two Outcome values with plain != and a slice field would make Outcome uncomparable, breaking that check at compile time"
  - "The no_diagnostics decline sets diagnosis_code to the decline reason itself (repair.no_diagnostics) rather than leaving it empty, honoring the plan's own unqualified must-have truth ('never emits an unrepairable outcome with an empty diagnosis_code') even though there is no diagnostic to name as the first element in that one case"

patterns-established:
  - "When a JSON-protocol field must marshal as an array but the containing struct is compared with == elsewhere in the codebase, use a comparable string-backed named type with custom (Un)MarshalJSON rather than a bare slice field"

requirements-completed: [DX-09]

coverage:
  - id: D1
    description: "Outcome gains three additive omitempty fields (diagnosis_codes, decline_reason, best_applicability); no existing field's type or tag changes; a repairable-path Outcome is byte-identical before and after"
    requirement: DX-09
    verification:
      - kind: unit
        ref: "cmd/lang-repair/repair_diagnosis_test.go#TestDeclineCarriesReasonAndDiagnosis"
        status: pass
      - kind: unit
        ref: "cmd/lang-repair/repair_diagnosis_test.go#TestRepairOutcomeDeclineFieldsPopulated"
        status: pass
      - kind: e2e
        ref: "manual: lang-repair --json against testdata/phase13/derivation_interprocedural_loan_defect.lang, byte-identical output on pre-change vs post-change repair.go"
        status: pass
    human_judgment: false
  - id: D2
    description: "TestUnrepairableAlwaysCarriesDiagnosis: every checked-in cmd/lang-repair/testdata/*_capture.json and every real unrepairable fixture class carries a non-empty diagnosis_code and an in-vocabulary decline_reason when its outcome is unrepairable"
    requirement: DX-09
    verification:
      - kind: unit
        ref: "cmd/lang-repair/repair_diagnosis_test.go#TestUnrepairableAlwaysCarriesDiagnosis"
        status: pass
      - kind: unit
        ref: "cmd/lang-repair/repair_diagnosis_test.go#TestDeclineReasonOutsideVocabularyFailsTheGuard"
        status: pass
    human_judgment: false
  - id: D3
    description: "The diagnosis guard is proven non-inert: a seeded empty-diagnosis fault makes it fail, and the unfaulted control passes"
    requirement: DX-09
    verification:
      - kind: unit
        ref: "cmd/lang-repair/repair_diagnosis_test.go#TestUnrepairableDiagnosisGuardIsNotInert"
        status: pass
    human_judgment: false

duration: 9min
completed: 2026-09-18
status: complete
---

# Phase 14 Plan 03: Unrepairable Decline Carries a Diagnosis Summary

**`lang-repair --json` no longer emits a silent `unrepairable` decline: every decline now carries a non-empty `diagnosis_code`, the full ordered `diagnosis_codes` list, and a coded `decline_reason` from a closed three-value vocabulary.**

## Performance

- **Duration:** 9 min
- **Started:** 2026-09-18T01:22:52Z
- **Completed:** 2026-09-18T01:31:03Z
- **Tasks:** 3
- **Files modified:** 2 (1 created, 1 modified)

## Accomplishments

- `Outcome` gains three additive `omitempty` fields — `DiagnosisCodes`, `DeclineReason`, `BestApplicability` — populated whenever the driver declines with `unrepairable`, with the repairable path proven byte-identical before/after.
- Three new closed-vocabulary decline reasons: `repair.none_offered` (invalid, ≥1 diagnostic, zero repairs anywhere), `repair.none_eligible` (repairs present, none driver-eligible), `repair.no_diagnostics` (invalid, zero diagnostics — a fail-closed anomaly that still reads `unrepairable`, never a pass).
- A new `classifyDecline` function carries out the classification `selectRepair`'s false path used to discard, without touching `selectRepair`'s own three-value signature (multiple existing test files call it and had to stay byte-unchanged).
- `TestUnrepairableAlwaysCarriesDiagnosis` sweeps the whole checked-in capture corpus (8 files) plus the three real unrepairable fixture classes, asserting the invariant holds everywhere the driver can actually produce `unrepairable` today.
- `TestUnrepairableDiagnosisGuardIsNotInert` proves the guard is a real guard: a fault-injection toggle (`diagnosisEmissionDisabledForTest`) reproduces the bug on demand through the existing stand-in-binary harness, and the guard catches it.

## Task Commits

1. **Task 1 (RED): failing tests for decline classification** — `474b1d7` (test)
2. **Task 1 (GREEN): carry the decline reason through `selectRepair` onto `Outcome`** — `8977d0b` (feat)
3. **Task 2: guard that no `unrepairable` outcome ever carries an empty diagnosis** — `a8196ad` (test)
4. **[Rule 1 deviation] fix: never leave `diagnosis_code` empty on a zero-diagnostic decline** — `a07e3fa` (fix)
5. **Task 3: prove the diagnosis guard is not inert with a stand-in binary** — `4bcd508` (test)

**Plan metadata:** committed alongside this SUMMARY.

_Note: Task 1 is TDD (RED → GREEN); Tasks 2 and 3 are `type="auto"` and land as single `test` commits each, matching their own nature as new guard/proof tests rather than production-behavior changes._

## Files Created/Modified

- `cmd/lang-repair/repair.go` — Widened `Outcome`; added `DiagnosisCodes`/`DeclineReason`/`BestApplicability`, the `Decline*` vocabulary, the applicability ranking helpers, `classifyDecline`, the comparable `diagnosisCodeList` marshaling type, and the `diagnosisEmissionDisabledForTest` fault-injection seam.
- `cmd/lang-repair/repair_diagnosis_test.go` (new) — `TestDeclineCarriesReasonAndDiagnosis`, `TestRepairOutcomeDeclineFieldsPopulated`, `TestUnrepairableAlwaysCarriesDiagnosis`, `TestDeclineReasonOutsideVocabularyFailsTheGuard`, `TestUnrepairableDiagnosisGuardIsNotInert`.

## Reclassified Declines (Task 2 requirement)

These pre-existing unrepairable declines change their reported `decline_reason` under this plan; all three remain `OutcomeUnrepairable` — honest declines stay honest, now stating their reason instead of leaving it as silence:

| Decline | Fixture / class | Old behavior | New `decline_reason` |
|---|---|---|---|
| Forward-direction interprocedural loan liveness (D-13-09b) | `testdata/phase07/relay_escort_witness.lang`, exercised as `interprocedural_loan_forward_direction` | `diagnosis_code` empty | `repair.none_offered` (zero repairs anywhere on the diagnostic) |
| Call-graph cycle (D-13-12) | Not exercised by any checked-in `.lang` fixture today — documented, not asserted by a live test | `diagnosis_code` empty | `repair.none_offered` (same shape: a cycle has no local, mechanical edit) |
| Call-argument-type uniqueness gate | `testdata/phase13/heldout_call_argument_ambiguous.lang` (mutated via `session.CallArgumentTypeInjector`), exercised as `call_argument_type_ambiguous` | `diagnosis_code` empty | `repair.none_offered` in this corpus's fixture (emission suppressed entirely) — would be `repair.none_eligible` if a non-eligible repair were left in place instead |

## Decisions Made

- `selectRepair`'s three-value signature stays exactly as it was; the widening lives in a new sibling `classifyDecline`, called only from `Repair`'s own false-path branch. Forced by the acceptance criterion that `antitheater_test.go` (and, in practice, `repair_test.go`) must stay byte-unchanged — both call `selectRepair` directly at several sites.
- `Outcome.DiagnosisCodes` is a comparable `diagnosisCodeList` string type (custom `MarshalJSON`/`UnmarshalJSON` producing/consuming a real JSON array) rather than `[]string`. `antitheater_test.go`'s `TestProseScrambleLeavesRepairBehaviourIdentical` compares two `Outcome` values with plain `!=`; a slice field makes a struct uncomparable and fails at compile time. Discovered by actually running `go vet` after the first (naive `[]string`) attempt, not by reasoning about it in advance.
- The `no_diagnostics` decline sets `diagnosis_code` to the decline reason itself (`repair.no_diagnostics`) instead of leaving it empty. The plan's own must-have truth is unqualified ("never emits an `unrepairable` outcome with an empty `diagnosis_code`"), and there is genuinely no diagnostic to name in this one case — the decline reason is the only honest non-empty value available, and it is still a closed-vocabulary constant, never invented prose.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `diagnosisEmissionDisabledForTest = declineDecision{}` left `diagnosis_code` empty on the `no_diagnostics` path**
- **Found during:** Re-reading the plan's must-have truths against the just-committed Task 2 guard, before starting Task 3
- **Issue:** `classifyDecline`'s `len(codes) == 0` branch left `decision.diagnosisCode` empty, honestly reflecting that there is no diagnostic to name — but this contradicts the plan's own unqualified must-have truth that `lang-repair --json` never emits an `unrepairable` outcome with an empty `diagnosis_code`. The Task 2 guard (`unrepairableDiagnosisOK`) already enforces non-emptiness unconditionally, so this was a latent bug that just happened not to be exercised by any test in the current corpus (no checked-in fixture reaches `no_diagnostics`).
- **Fix:** `classifyDecline` now sets `diagnosis_code` to the decline reason (`repair.no_diagnostics`) itself in that one branch — still a closed-vocabulary constant, never prose.
- **Files modified:** `cmd/lang-repair/repair.go`, `cmd/lang-repair/repair_diagnosis_test.go` (updated the corresponding `TestDeclineCarriesReasonAndDiagnosis` subtest assertion)
- **Verification:** `go test ./cmd/lang-repair/... -count=1` green; the updated subtest asserts `diagnosisCode == DeclineNoDiagnostics` directly
- **Committed in:** `a07e3fa`

---

**Total deviations:** 1 auto-fixed (1 Rule 1 bug fix).
**Impact on plan:** Necessary for correctness against the plan's own stated invariant; caught before Task 3's non-inertness proof, so the proof exercises the corrected behavior throughout. No scope creep.

## Issues Encountered

- **`Outcome` uncomparability with a bare `[]string` field.** The plan's literal action text describes `DiagnosisCodes []string`. A first attempt using that exact type broke `go vet` on `antitheater_test.go:543` and `:691` (`invalid operation: scrambledOutcome != identityOutcome (struct containing []string cannot be compared)`), because `Outcome` is compared with `!=` in an existing, must-stay-unchanged test file. Resolved by introducing the comparable `diagnosisCodeList` string type with custom JSON marshaling (see Decisions Made). The wire shape (`"diagnosis_codes": [...]`) is unaffected.
- **`selectRepair`'s signature could not literally be "widened."** The plan's action text says to widen `selectRepair` to return a decision value alongside. Doing so directly breaks every existing call site in `antitheater_test.go` and `repair_test.go` (both call it with the original 3-value form), and the plan's own acceptance criteria require `antitheater_test.go` to stay byte-unchanged. Resolved by adding a sibling `classifyDecline` function instead, called only from `Repair`'s false-path branch — same net effect (the classification computed on the false path is carried out instead of discarded), without touching `selectRepair`'s shape.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `DX-09` is closed: `lang-repair --json` never emits an `unrepairable` outcome with an empty diagnosis, and the guard proving it is itself proven non-inert.
- `go test ./...` (full repo): every package is green except the pre-existing, out-of-scope `internal/compiler/native` `TestSourceNeverSpawnsUnboundedProcesses` failure on two files that predate this plan (`.claude/skills/spike-findings-ai-lang/sources/005-native-ffi-provenance-cleanup/lab/lab.go`, and `internal/compiler/session/verification_groundedness_test.go` from plan 14-01) — explicitly called out as out of this plan's scope in the dispatch context, and the orchestrator is handling it at the wave gate.
- `git diff --stat -- cmd/lang-repair/antitheater_test.go cmd/lang-repair/import_boundary_test.go` prints nothing, confirming both protected test files are byte-unchanged.
- No blockers for the next plan in this phase.

---
*Phase: 14-evidence-instrument-and-honest-scoping*
*Completed: 2026-09-18*

## Self-Check: PASSED

- `cmd/lang-repair/repair.go` — FOUND
- `cmd/lang-repair/repair_diagnosis_test.go` — FOUND
- `.planning/phases/14-evidence-instrument-and-honest-scoping/14-03-SUMMARY.md` — FOUND
- Commits `474b1d7`, `8977d0b`, `a8196ad`, `a07e3fa`, `4bcd508` — all FOUND in `git log --oneline --all`
