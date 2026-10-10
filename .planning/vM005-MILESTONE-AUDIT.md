---
milestone: M005
milestone_name: Practical Computation
audited: 2026-10-10T23:30:00Z
audited_head: 3d4edce
evidence_revision: 112eaf7bb5f7e2a60cd0dbcea5021f53d0c0cf0f
supersedes: "audit at d12200e (91b53e0; status tech_debt, Nyquist partial, 5 process gaps; 4 closed since by secure-phase and validate-phase runs)"
status: tech_debt
scores:
  requirements: 9/9
  phases: 4/4
  integration: 15/15 cross-phase links connected (2 orphaned exports, both test-facing)
  flows: 6/7 complete (1 partial by design: F3)
gaps:
  requirements: []
  integration: []
  flows: []
process_gaps:
  - phase: 27.1-address-tech-debt-build-route-origin-validation-and-receipt
    item: "No 27.1-REVIEW.md while workflow.code_review=true"
    action: "/gsd-code-review 27.1 (optional before completion)"
nyquist:
  compliant_phases: ["26", "27", "27.1", "27.2"]
  partial_phases: []
  not_validated_phases: []
  missing_phases: []
  overall: compliant
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
  - "Process: 27.2-SECURITY.md (084c3dc, 13/13 closed), 27.1-SECURITY.md (af48608, 13/13 closed), 27.2-VALIDATION.md validated (fa20d34, 0 gaps), 27.1-VALIDATION.md validated (3d4edce, 0 gaps; row 27.1-03-xx green from run 38068888687; table kept outside the graded corpus by user decision)."
---

# Milestone M005: Practical Computation (Re-audit after 27.1/27.2 process closure)

**Audited HEAD:** 3d4edce. Commits after 112eaf7 change only planning docs (`git diff --name-only d12200e HEAD` lists only `.planning/` files).
**Evidence revision:** 112eaf7, hosted run [38080236918](https://github.com/szTheory/schway/actions/runs/38080236918). Full and race suites plus the Phase 27 receipt lanes passed on macOS and Ubuntu.
**Status:** tech_debt. All 9 requirements are satisfied and there are no critical blockers. W1–W4 remain (W3 narrowed to the native path). Security and Nyquist validation are now complete for all four phases. The only open process item is the optional 27.1 code review.

## Phases

| Phase | Verification | Score | Validation (Nyquist) | Security | Review |
|---|---|---|---|---|---|
| 26 Checked Scalar Sum | passed, with 27.1 and 27.2 addenda | 5/5 | validated, compliant | verified, 0 open | clean |
| 27 Exact FizzBuzz and Cross-Host Evidence | passed, with 27.1 addendum | 5/5 | validated, compliant (5/5, 0 gaps) | verified, 26/26 closed | IN-01 and IN-02 deferred |
| 27.1 Build-route origin validation and receipt-gate docs | passed | 25/25 | validated, compliant (5/5, 0 gaps) | verified, 13/13 closed | **none** (optional) |
| 27.2 Straight-line scalar let admission (R1) | passed | 5/5 | validated, compliant (5/5, 0 gaps) | verified, 13/13 closed | WR-01 skipped, IN-01/IN-02 open |

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

The integration checker built `schway` into a scratch directory and re-ran the public routes live at 3d4edce. The repo tree stayed clean. No link or flow status changed from the d12200e audit.

- R1 headline (inline and `examples/phase272/straight_line_copy.schway`): check passes, build succeeds, `app run -- 41` prints 41.
- R1b is still refused with `syntax.expected_linear_result`, as deferred.
- `sum_to_n` with inputs 0, 10 and 1000 prints 0, 55 and 500500. Input 1001 exits 65 with empty stdout.
- FizzBuzz 15 and 1000 are byte-identical to `examples/phase27/fizzbuzz-{15,1000}.out`. Input 0 gives empty stdout and 1001 exits 65.
- Overflow and a dynamic zero divisor exit 65. A constant zero divisor is refused at build with `type.scalar_remainder_zero` and no artifact.
- F1: `testdata/phase3/public_view_omitted.schway` build exits 2 `core.origin_omitted` with no artifact. The phase4 fixture gives `core.foreign_origin_omitted`, also exit 2.
- F2: the guide's pinned gate command, re-run at 3d4edce, prints `status=pass hosts=macOS,Linux revision=6224f8b…`.
- A dynamic zero divisor needs an input-derived divisor (`100 % n` with n=0). `var d = 0` then `n % d` is refused at check with `type.scalar_remainder_zero`, because the checker folds the constant through the initializer. Fail-closed; noted, not a gap.

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

- **Process gaps closed.** Security reviews written for 27.1 and 27.2 (13/13 threats closed each). Validation records for 27.1 and 27.2 moved to `validated`, Nyquist-compliant, with zero gaps. See `closed_since_prior_audit`.
- **Nyquist overall: compliant** (was partial).
- **No code changes.** Integration and flow results are unchanged; the checker's re-run reproduced every live claim, and F2 was re-run by the orchestrator.
- **Still open:** optional `/gsd-code-review 27.1`.

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
| 27.1 | validated | true | — |
| 27.2 | validated | true | — |
