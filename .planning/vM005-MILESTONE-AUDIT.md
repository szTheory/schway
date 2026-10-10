---
milestone: M005
milestone_name: Practical Computation
audited: 2026-10-10T19:00:00Z
audited_head: afb2509
evidence_revision: 7936dac39a6215091d4dbd02a6acec20f2601386
supersedes: "audit at b6cbcc2 (status tech_debt; F1/F2 closed since by Phase 27.1)"
status: tech_debt
scores:
  requirements: 9/9
  phases: 3/3
  integration: 15/15 cross-phase links connected (2 orphaned exports, both test-facing)
  flows: 5/6 complete (1 partial by design: F3)
gaps:
  requirements: []
  integration: []
  flows: []
process_gaps:
  - phase: 27.1-address-tech-debt-build-route-origin-validation-and-receipt
    item: "27.1-SECURITY.md missing while workflow.security_enforcement=true; plans declare threats T-27.1-01..09"
    action: "/gsd-secure-phase 27.1"
  - phase: 27.1-address-tech-debt-build-route-origin-validation-and-receipt
    item: "27.1-VALIDATION.md status in-progress; row 27.1-03-xx still pending although hosted run 38068888687 at 7936dac is green and its gated receipts are committed. Moving to validated enrols the table in the graded corpus (needs Grade/Non-inertness columns plus a regenerated corpus record)"
    action: "/gsd-validate-phase 27.1"
  - phase: 27.1-address-tech-debt-build-route-origin-validation-and-receipt
    item: "No 27.1-REVIEW.md while workflow.code_review=true (source diff: session.go, session_app_verify.go, executionpeer.go comment, receipt-gate script, tests)"
    action: "/gsd-code-review 27.1 (optional before completion)"
nyquist:
  compliant_phases: ["26", "27"]
  partial_phases: []
  not_validated_phases: ["27.1"]
  missing_phases: []
  overall: partial
tech_debt:
  - phase: cross-phase (Phase 26 regression, found in this re-audit)
    items:
      - "R1: straight-line scalar `let` bodies are refused. `fn main(n: U64) -> U64 { let total = n  total }` fails `schway check` and `schway build` with core.release_order_indeterminate (corevalidate terminal-block-has-no-incoming-edge rule); the same body inside if/else is accepted. Bisected by the integration checker to a671308 `feat(26-05): lower source-authored scalar lets as copies`; the parent accepted and ran it. Reproduced by the auditor at afb2509. Fails closed, no committed example affected; the symptom was noted in 27-05-SUMMARY (branchless Unit witness) but not escalated. Not a FLOW-01 blocker (if/else and loops admitted), but a straight-line program that ran before M005 no longer builds. (FLOW-01, U64-01)"
  - phase: cross-phase (evidence route)
    items:
      - "W1: `schway app verify` on examples/sum_to_n.schway passes for n 0..9 but fails for n=10 and n=1000 with native.run_signaled (abort trap): the generated C aborts when the event recorder runs out of capacity (cgen_program.go ~1660). Fails closed (verified:false, exit 3) but reports a signal, not capacity exhaustion. FizzBuzz has a Unit entry, which app verify reports as unsupported. The public app verify route therefore does not peer-check the pinned M005 answers; EVD-12 rests on the per-engine session tests and the hosted receipts. Pre-existing (identical at b6cbcc2). (EVD-12, APP-07, APP-08)"
      - "W2: one admission helper, two exit classes. `schway build` reports an origin refusal as invalid (exit 2) but a corevalidate refusal as operational_failure (exit 3, native.build_failed), while `schway check` reports both as invalid (exit 2). Strings were preserved by design (D-27.1-04); the open item is the deferred 'build corevalidate failure as a diagnostic'. (DX-16)"
      - "W3: `schway run` (RunInterpreterCommandFile/RunNativeCommandFile) checks origin only after successful execution; on the origin fixture both engines return operational_failure, not core.origin_omitted. Pre-existing M004 behavior, explicitly out of 27.1 scope (D-27.1-02). No M005 requirement."
      - "W4: following the guide's clean-checkout steps writes ./phase27-fizzbuzz, ./phase27-sum and ./*.actual into the repo root (not gitignored); a subsequent `sh scripts/verify-phase27.sh` then records tree=modified and the gate rejects with tree_not_clean. Fails closed; the guide already says modified-tree receipts are development-only. (DX-16)"
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
      - "Focused Phase 26 hosted job not re-dispatched since 37118144404; current behavioral evidence is the full/race run 38068888687 at 7936dac (26-VERIFICATION addendum)."
closed_since_prior_audit:
  - "F1: build route now origin-checked (admitPublishedApplication shared by BuildApplication, RunInterpreterApplication, VerifyApplicationCases); live CLI exit 2 core.origin_omitted, no artifact/sidecar"
  - "F2: guide documents pinned-revision re-check; TestPhase27GuideContract executes it; bare command fails closed with stderr hint"
  - "F6: executionpeer.ValidateFullCoverage documented test-only"
  - "27-09/27-10 SUMMARY requirements-completed added; DX-16 now in SUMMARY frontmatter"
  - "26-05 Verified=true wording amended (26-PLANNING-AMENDMENTS.md)"
  - "27-SECURITY.md verified (26/26 closed); 27-VALIDATION.md validated (5/5, 0 gaps)"
---

# Milestone M005: Practical Computation (Re-audit)

**Audited HEAD:** afb2509. Commits after 7936dac change only planning docs.
**Evidence revision:** 7936dac, hosted run [38068888687](https://github.com/szTheory/schway/actions/runs/38068888687). Full and race suites plus the Phase 27 receipt lanes passed on macOS and Ubuntu.
**Status:** tech_debt. All 9 requirements are satisfied and there are no critical blockers. Phase 27.1 closed F1 and F2. This re-audit found one new fail-closed regression (R1) and four warnings. Phase 27.1 still has three open process items.

## Phases

| Phase | Verification | Score | Validation (Nyquist) | Security | Review |
|---|---|---|---|---|---|
| 26 Checked Scalar Sum | passed, with 27.1 addendum | 5/5 | validated, compliant | verified, 0 open | clean |
| 27 Exact FizzBuzz and Cross-Host Evidence | passed, with 27.1 addendum | 5/5 | validated, compliant (5/5, 0 gaps) | verified, 26/26 closed | IN-01 and IN-02 deferred |
| 27.1 Build-route origin validation and receipt-gate docs | passed | 25/25 | **in-progress** (hosted row pending) | **missing** | **none** |

## Requirements (3-source cross-reference)

| REQ | Phase | VERIFICATION | SUMMARY frontmatter | REQUIREMENTS.md | Final |
|---|---|---|---|---|---|
| U64-01 | 26 | passed | 26-01, 02, 05 | [x] | satisfied |
| U64-02 | 27 | passed | 27-03, 04, 06, 08, 10 | [x] | satisfied |
| FLOW-01 | 26 | passed | 26-01, 03, 04, 05 | [x] | satisfied (R1 is debt, not a gap) |
| FLOW-02 | 26 | passed | 26-01, 03, 04, 05 | [x] | satisfied |
| APP-07 | 26 | passed | 26-01, 02, 04, 05 | [x] | satisfied |
| APP-08 | 27 | passed | 27-01, 02, 05–10 | [x] | satisfied |
| APP-09 | 27 | passed | 27-01, 02, 06–10 | [x] | satisfied |
| EVD-12 | 27 | passed | 27-07, 08, 09, 10 | [x] | satisfied |
| DX-16 | 27 | passed | 27-09, 10 | [x] | satisfied (was partial in the prior audit and is now fully matched) |

Phase 27.1 maps no requirement IDs. It strengthens assurance depth for all nine requirements. None is orphaned or unsatisfied.

## Integration

The integration checker built `schway` into a scratch directory and ran the public routes live at afb2509.

- `sum_to_n` with inputs 0, 10 and 1000 prints 0, 55 and 500500.
- FizzBuzz output for 15 and 1000 is byte-identical to the fixtures. For 0, stdout is empty.
- Input 1001 exits 65 on both programs, with empty stdout.
- Overflow exits 65. A dynamic zero divisor exits 65. A constant zero divisor is refused at build with `type.scalar_remainder_zero`.

All 15 links are connected: the 12 from the prior audit plus 3 added by 27.1.
- **Shared admission helper:** one helper has exactly three callers.
- **Origin refusal to CLI:** an origin refusal reaches the CLI as exit 2, `invalid`.
- **Guide to gate:** the guide's pinned command reaches the receipt gate.

Two exports are orphaned, both acceptable:
- `RunInterpreterApplication` is a test-facing API. It is gated.
- `ValidateFullCoverage` is documented as test-only.

| Flow | Status |
|---|---|
| source → check → corevalidate → origin → build → app run | complete (live) |
| input guard 1001 | complete (live) |
| overflow / dynamic-zero / constant-zero controls | complete (live and focused tests) |
| native build → origin peer (F1) | **complete, closed by 27.1** |
| committed receipts → gate via guide command (F2) | **complete, closed by 27.1** |
| app-run evidence → executionpeer (F3) | partial (by design; see W1) |

## New findings this audit

- **R1 (regression, fails closed).** A straight-line scalar `let` body is refused with `core.release_order_indeterminate`, while the same body inside `if/else` is accepted. The integration checker bisected the cause to a671308, from Phase 26-05. I reproduced it at afb2509: `check` exits 2 and `build` exits 3. It is not a requirement gap, but it narrows what ordinary source can do.
- **W1.** `app verify` aborts on event capacity for `sum_to_n` at n ≥ 10. It reports this as a signal, not as capacity exhaustion. FizzBuzz has no public peer route.
- **W2.** `build` uses two exit classes for refusals from the same helper.
- **W3.** `schway run` checks origin only after the program has run successfully. This predates M005.
- **W4.** Following the guide's steps dirties the repository tree, so a later receipt is rejected with `tree_not_clean`.

## Tech debt

See the `tech_debt` frontmatter:
- 1 regression (R1).
- 4 warnings (W1–W4).
- 9 dispositions deferred by D-27.1-10.
- 1 Phase 26 evidence note.

## Nyquist coverage

| Phase | VALIDATION.md | Compliant | Action |
|---|---|---|---|
| 26 | validated | true | — |
| 27 | validated | true | — |
| 27.1 | in-progress | not authoritative | `/gsd-validate-phase 27.1` |
