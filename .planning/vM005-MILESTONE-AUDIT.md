---
milestone: M005
milestone_name: Practical Computation
audited: 2026-10-10T22:00:00Z
audited_head: d12200e
evidence_revision: 112eaf7bb5f7e2a60cd0dbcea5021f53d0c0cf0f
supersedes: "audit at afb2509 (status tech_debt; R1 closed since by Phase 27.2)"
status: tech_debt
scores:
  requirements: 9/9
  phases: 4/4
  integration: 15/15 cross-phase links connected (2 orphaned exports, both test-facing)
  flows: 6/7 complete (R1 straight-line flow added and complete; 1 partial by design: F3)
gaps:
  requirements: []
  integration: []
  flows: []
process_gaps:
  - phase: 27.1-address-tech-debt-build-route-origin-validation-and-receipt
    item: "27.1-SECURITY.md missing while workflow.security_enforcement=true; plans declare threats T-27.1-01..09"
    action: "/gsd-secure-phase 27.1"
  - phase: 27.1-address-tech-debt-build-route-origin-validation-and-receipt
    item: "27.1-VALIDATION.md status in-progress; row 27.1-03-xx still pending although hosted run 38068888687 at 7936dac is green. Moving to validated enrols the table in the graded corpus (needs Grade/Non-inertness columns plus a regenerated corpus record)"
    action: "/gsd-validate-phase 27.1"
  - phase: 27.1-address-tech-debt-build-route-origin-validation-and-receipt
    item: "No 27.1-REVIEW.md while workflow.code_review=true"
    action: "/gsd-code-review 27.1 (optional before completion)"
  - phase: 27.2-address-tech-debt-straight-line-scalar-let-admission-r1
    item: "27.2-SECURITY.md missing while workflow.security_enforcement=true; plans declare threats T-27.2-01..12"
    action: "/gsd-secure-phase 27.2"
  - phase: 27.2-address-tech-debt-straight-line-scalar-let-admission-r1
    item: "27.2-VALIDATION.md status in-progress, nyquist_compliant false; all five SC rows are green (SC5 = run 38080236918 at 112eaf7)"
    action: "/gsd-validate-phase 27.2"
nyquist:
  compliant_phases: ["26", "27"]
  partial_phases: []
  not_validated_phases: ["27.1", "27.2"]
  missing_phases: []
  overall: partial
tech_debt:
  - phase: cross-phase (evidence route)
    items:
      - "W1: `schway app verify` on examples/sum_to_n.schway verifies small n but aborts with native.run_signaled (abort trap) when the generated C event recorder runs out of capacity. At d12200e the integration checker observed the boundary at n >= 15 (n <= 14 verifies); the prior audit recorded n >= 10. 27.2 does not touch this path. Fails closed but reports a signal, not capacity exhaustion. FizzBuzz has a Unit entry, which app verify refuses, so it has no public peer route. EVD-12 rests on the per-engine session tests and hosted receipts. (EVD-12, APP-07, APP-08)"
      - "W2: one admission helper, two exit classes. `schway build` reports an origin refusal as invalid (exit 2) but a corevalidate refusal as operational_failure (exit 3, native.build_failed). R1 was the only known live trigger and is now closed; the mapping in runBuild is unchanged. (DX-16)"
      - "W3 (narrowed): `schway run --engine=interpreter` on testdata/phase4/foreign_origin_omitted.schway now exits 2 core.foreign_origin_omitted; `--engine=native` still exits 3 operational_failure (native.tool_failure). Pre-existing M004 behavior, out of 27.1 scope (D-27.1-02). No M005 requirement."
      - "W4: the guide's clean-checkout steps write ./phase27-fizzbuzz, ./phase27-sum and ./*.actual into the repo root (not gitignored), so a later receipt is rejected with tree_not_clean. Fails closed. (DX-16)"
  - phase: 27.2-address-tech-debt-straight-line-scalar-let-admission-r1
    items:
      - "R1b (deferred): `let a = n  let b = helper(a)  b` still refused with syntax.expected_linear_result; pinned by TestPhase272R1bScalarFirstLetThenCallCharacterization."
      - "IN-01: hard-coded `return:1` block ID in a 27.2 test; IN-02: unreachable Bool allowance in isStraightLineScalarBody (27.2-REVIEW-DISPOSITION.md). WR-01 skipped with rationale."
      - "Note (integration checker): cgen hasScalarOperations now admits straight-line bodies, while corevalidate.hasScalarCFG and originvalidate.originHasScalarCFG keep their old classifier sets by design (D-27.2-04, validators unchanged). Divergence is fail-closed (validators run first and refuse anything they do not admit); no live divergence on accepted programs, and byte identity of the accepted corpus is pinned by TestPhase272AcceptedCorpusByteIdentity. (FLOW-01, U64-01)"
      - "Phase 27 and 27.1 receipts not regenerated after 27.2 (D-27.2-05); receipt-pinned sources/fixtures verified unchanged."
  - phase: 27.1 deferred dispositions (D-27.1-10, recorded in STATE.md)
    items:
      - "F3: app run --evidence=events not peer-validated in-route (by design)"
      - "F4: fixed-point budget counting differs across checker/corevalidate/originvalidate (all 65536, all fail-closed)"
      - "F5: 65,536-byte output ceiling unreachable from accepted inputs (seam-tested only)"
      - "IN-01: fwrite failure conflated with output-limit exhaustion"
      - "IN-02: duplicated reason literals across layers"
      - "examples/fizzbuzz.schway and examples/sum_to_n.schway not canonically formatted (changes pinned hashes)"
      - "build corevalidate failure as a diagnostic (see W2)"
      - "RunInterpreterCommandFile re-read in publishedOriginProblemFile"
      - "gate ancestor-revision acceptance (rejected for now, D-27.1-06)"
  - phase: 26-checked-scalar-sum
    items:
      - "Focused Phase 26 hosted job not re-dispatched since 37118144404; current behavioral evidence is the full/race run 38080236918 at 112eaf7 (26-VERIFICATION addendum)."
closed_since_prior_audit:
  - "R1: straight-line scalar let bodies admitted again by producer-side fix (Phase 27.2; check/scalar.go return-block split, cgen isStraightLineScalarBody, conditional (void)param). Live at d12200e: `fn main(n: U64) -> U64 { let total = n  total }` and examples/phase272/straight_line_copy.schway check, build and `app run -- 41` prints 41. Validators unchanged (empty diff from 9f6100a). Hosted run 38080236918 at 112eaf7 green on macOS and Linux (full and race)."
  - "W3 interpreter path now reports the origin refusal before execution (native path still does not)."
---

# Milestone M005: Practical Computation (Re-audit after Phase 27.2)

**Audited HEAD:** d12200e. Commits after 112eaf7 change only planning docs.
**Evidence revision:** 112eaf7, hosted run [38080236918](https://github.com/szTheory/schway/actions/runs/38080236918). Full and race suites plus the Phase 27 receipt lanes passed on macOS and Ubuntu.
**Status:** tech_debt. All 9 requirements are satisfied and there are no critical blockers. Phase 27.2 closed R1. W1–W4 remain (W3 narrowed to the native path). Phases 27.1 and 27.2 have open process items (security, validation, and a 27.1 review).

## Phases

| Phase | Verification | Score | Validation (Nyquist) | Security | Review |
|---|---|---|---|---|---|
| 26 Checked Scalar Sum | passed, with 27.1 and 27.2 addenda | 5/5 | validated, compliant | verified, 0 open | clean |
| 27 Exact FizzBuzz and Cross-Host Evidence | passed, with 27.1 addendum | 5/5 | validated, compliant (5/5, 0 gaps) | verified, 26/26 closed | IN-01 and IN-02 deferred |
| 27.1 Build-route origin validation and receipt-gate docs | passed | 25/25 | **in-progress** (hosted row pending) | **missing** | **none** |
| 27.2 Straight-line scalar let admission (R1) | passed | 5/5 | **in-progress** (all rows green) | **missing** | WR-01 skipped, IN-01/IN-02 open |

## Requirements (3-source cross-reference)

| REQ | Phase | VERIFICATION | SUMMARY frontmatter | REQUIREMENTS.md | Final |
|---|---|---|---|---|---|
| U64-01 | 26 | passed | 26-01, 02, 05 | [x] | satisfied |
| U64-02 | 27 | passed | 27-03, 04, 06, 08, 10 | [x] | satisfied |
| FLOW-01 | 26 | passed | 26-01, 03, 04, 05 | [x] | satisfied (R1 closed by 27.2) |
| FLOW-02 | 26 | passed | 26-01, 03, 04, 05 | [x] | satisfied |
| APP-07 | 26 | passed | 26-01, 02, 04, 05 | [x] | satisfied |
| APP-08 | 27 | passed | 27-01, 02, 05–10 | [x] | satisfied |
| APP-09 | 27 | passed | 27-01, 02, 06–10 | [x] | satisfied |
| EVD-12 | 27 | passed | 27-07, 08, 09, 10 | [x] | satisfied (public app verify peer route still limited, W1) |
| DX-16 | 27 | passed | 27-09, 10 | [x] | satisfied |

Phases 27.1 and 27.2 map no requirement IDs (their SUMMARY frontmatter lists none). Both restore or strengthen existing behavior. No requirement is orphaned or unsatisfied.

## Integration

The integration checker built `schway` into a scratch directory and ran the public routes live at d12200e. The repo tree stayed clean.

- R1 headline (inline and `examples/phase272/straight_line_copy.schway`): check passes, build succeeds, `app run -- 41` prints 41.
- R1b is still refused with `syntax.expected_linear_result`, as deferred.
- `sum_to_n` with inputs 0, 10 and 1000 prints 0, 55 and 500500. Input 1001 exits 65 with empty stdout.
- FizzBuzz 15 and 1000 are byte-identical to `examples/phase27/fizzbuzz-{15,1000}.out`. Input 0 gives empty stdout and 1001 exits 65.
- Overflow and a dynamic zero divisor exit 65. A constant zero divisor is refused at build with `type.scalar_remainder_zero` and no artifact.
- F1: `testdata/phase3/public_view_omitted.schway` build exits 2 `core.origin_omitted` with no artifact. The phase4 fixture gives `core.foreign_origin_omitted`, also exit 2.
- F2: the guide's pinned gate command passes on the committed receipts.

All 15 links are still connected. `admitPublishedApplication` still has exactly three callers. The 27.2 cgen classifier is reached only through `hasScalarOperations` and does not fire on `sum_to_n` (it has branches).

| Flow | Status |
|---|---|
| source → check → corevalidate → origin → build → app run | complete (live) |
| straight-line scalar body (R1) | **complete, closed by 27.2** |
| input guard 1001 | complete (live) |
| overflow / dynamic-zero / constant-zero controls | complete (live) |
| native build → origin peer (F1) | complete |
| committed receipts → gate via guide command (F2) | complete |
| app-run evidence → executionpeer (F3) | partial (by design; see W1) |

## Changes since the prior audit

- **R1 closed.** See `closed_since_prior_audit`.
- **W1 boundary moved.** `app verify` on `sum_to_n` now verifies n ≤ 14 and aborts at n ≥ 15. The prior audit said n ≥ 10. 27.2 does not touch this path. It is still fail-closed and still mislabelled as a signal.
- **W3 narrowed.** The interpreter `run` path now reports the origin refusal. The native path still reports `operational_failure`.
- **New note.** The cgen scalar classifier and the validator classifiers differ by design (D-27.2-04). This is fail-closed and is recorded as tech debt, not a gap.

## Tech debt

See the `tech_debt` frontmatter:
- 4 warnings (W1–W4).
- 4 Phase 27.2 items: R1b, review info items, classifier note, receipts not regenerated.
- 9 dispositions deferred by D-27.1-10.
- 1 Phase 26 evidence note.

## Nyquist coverage

| Phase | VALIDATION.md | Compliant | Action |
|---|---|---|---|
| 26 | validated | true | — |
| 27 | validated | true | — |
| 27.1 | in-progress | not authoritative | `/gsd-validate-phase 27.1` |
| 27.2 | in-progress | not authoritative | `/gsd-validate-phase 27.2` |
