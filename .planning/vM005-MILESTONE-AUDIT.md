---
milestone: M005
milestone_name: Practical Computation
audited: 2026-10-10T23:59:00Z
audited_head: ed0aeff
evidence_revision: d7ffbcd14d581b61c6b265b135a52b2191ba2428
supersedes: "audit at 3d4edce (status tech_debt; W1-W4 open, Nyquist compliant 4/4)"
status: tech_debt
scores:
  requirements: 9/9
  phases: 5/5
  integration: 18/18 cross-phase links connected (0 new orphans)
  flows: 8/9 complete (1 partial by design: F3)
gaps:
  requirements: []
  integration: []
  flows: []
process_gaps:
  - phase: 27.1-address-tech-debt-build-route-origin-validation-and-receipt
    item: "No 27.1-REVIEW.md while workflow.code_review=true"
    action: "/gsd-code-review 27.1 (optional before completion)"
nyquist:
  compliant_phases: ["26", "27", "27.1", "27.2", "27.3"]
  partial_phases: []
  not_validated_phases: []
  missing_phases: []
  overall: compliant
tech_debt:
  - phase: 27.3-address-tech-debt-app-verify-capacity-abort-and-guide-clean
    items:
      - "N1: exhausted app verify cases embed full interpreter evidence (~371 bytes per unit of n), so at n >= 173 (single case; smaller with several cases) the report exceeds MaxReplayCasesBytes and app verify exits 3 with 'application verification report exceeds 65536 bytes' and no report file, instead of the named native.event_capacity_exhausted case. Fail-closed, never passes. Cause: session_app_verify.go sets Engines={interpreter} on exhausted cases; size check in writeApplicationVerificationReport. (EVD-12, APP-07)"
      - "N2: hosted two-host run 38091738463 is at d7ffbcd; post-review code changes (WR-01 logic 66ab558, WR-02 docs-test hardening dacbf4a/d2eadb8, IN-01/IN-02 comments, Nyquist test 6529d79) have local evidence only. (EVD-12)"
      - "N3: capacity contract string and exit 69 are literals in cgen.go, kept in sync with native.EventCapacityStderr/EventCapacityExitStatus only by cgen_phase273_test.go. (EVD-12)"
      - "N4: user-facing guides (examples/phase22/README.md, examples/phase27/README.md) do not mention the n <= 14 app verify bound or the unsupported/native.event_capacity_exhausted outcome. (DX-16)"
  - phase: cross-phase (evidence route)
    items:
      - "W2: one admission helper, two exit classes. `schway build` reports an origin refusal as invalid (exit 2) but a corevalidate refusal as operational_failure (exit 3, native.build_failed). runBuild mapping unchanged at ed0aeff. (DX-16)"
      - "W3 (narrowed): `run --engine=interpreter` on testdata/phase4/foreign_origin_omitted.schway exits 2 core.foreign_origin_omitted; `--engine=native` still exits 3 native.tool_failure. Pre-existing M004 behavior. No M005 requirement."
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
  - "W1 (Phase 27.3): app verify on sum_to_n no longer aborts with native.run_signaled. Generated conformance C exits 69 with 'schway: event capacity exhausted'; native runner names native.event_capacity_exhausted; app verify reports per-case unsupported (n=15) with Verified=false and exit 3; n=14 passes on all three engines; wrong expected stays fail; one-level-only exhaustion fails as divergence (WR-01). Live at ed0aeff. Residual N1 above."
  - "W4 (Phase 27.3): guide outputs live under ignored .schway/phase27/; clean clone after the full guide run has empty porcelain and verify-phase27.sh prints tree=clean status=pass. Live at ed0aeff."
  - "Process: 27.3-REVIEW (5/5 fixed, disposition 0 open), 27.3-SECURITY.md (19/19 closed), 27.3-VALIDATION.md validated (1 gap filled: WR-01 both directions, 6529d79)."
---

# Milestone M005: Practical Computation (Re-audit after Phase 27.3)

**Audited HEAD:** ed0aeff. **Hosted evidence:** run [38091738463](https://github.com/szTheory/schway/actions/runs/38091738463) at d7ffbcd (full and race, macOS and Ubuntu, success). Post-review code changes after d7ffbcd have local evidence only (N2).
**Status:** tech_debt. 9/9 requirements satisfied, no blockers. W1 and W4 closed by 27.3. W2 and W3 unchanged. Four new non-blocking warnings (N1-N4). Security and Nyquist complete for all five phases. Only open process item: optional 27.1 code review.

## Phases

| Phase | Verification | Validation (Nyquist) | Security | Review |
|---|---|---|---|---|
| 26 Checked Scalar Sum | passed (27.1/27.2 addenda) | compliant | 0 open | clean |
| 27 Exact FizzBuzz and Cross-Host Evidence | passed (27.1 addendum) | compliant | 26/26 closed | IN-01/IN-02 deferred |
| 27.1 Build-route origin validation | passed | compliant | 13/13 closed | **none** (optional) |
| 27.2 Straight-line scalar let admission (R1) | passed | compliant | 13/13 closed | WR-01 skipped, IN-01/IN-02 open |
| 27.3 App verify capacity abort and guide clean tree (W1, W4) | passed | compliant (1 gap filled) | 19/19 closed | 5/5 fixed |

## Requirements (3-source cross-reference)

Unchanged from the 3d4edce audit: U64-01, U64-02, FLOW-01, FLOW-02, APP-07, APP-08, APP-09, EVD-12 and DX-16 are each passed in VERIFICATION, listed in SUMMARY frontmatter, and checked in REQUIREMENTS.md. Phases 27.1-27.3 map no requirement IDs. No orphaned or unsatisfied requirement.

## Integration (live at ed0aeff, scratch build, repo tree clean)

18/18 links wired: the prior 15 (validators, scripts, go.mod and go.sum diff empty from 3d4edce; admitPublishedApplication still 3 callers) plus L16 generated C -> runner (exit 69 + exact stderr -> native.event_capacity_exhausted), L17 runner -> app verify -> CLI exit 3, L18 guide -> receipt (ignored .schway/phase27/).

| Flow | Status |
|---|---|
| source -> check -> corevalidate -> origin -> build -> app run | complete (live) |
| R1 straight-line copy (app run -- 41 prints 41) | complete (live) |
| input guard 1001 | complete (live) |
| overflow / dynamic-zero / constant-zero controls | complete (live) |
| F1 native build -> origin peer | complete (live) |
| F2 committed receipts -> gate via guide command | complete (live) |
| F3 app-run evidence -> executionpeer | partial (by design) |
| F8 app verify past capacity -> named unsupported -> exit 3 | complete for 15 <= n <= 172 (N1 above) |
| F9 guide clean checkout -> receipt driver | complete (live, tree=clean) |

Byte identity: TestPhase273ConformanceCorpusByteIdentity and TestPhase272AcceptedCorpusByteIdentity pass; loop-free conformance C and application C unchanged.

## Tech debt

See the `tech_debt` frontmatter: N1-N4 (27.3), W2-W3, 4 Phase 27.2 items, 9 dispositions deferred by D-27.1-10, 1 Phase 26 evidence note.

## Nyquist coverage

| Phase | VALIDATION.md | Compliant | Action |
|---|---|---|---|
| 26 | validated | true | — |
| 27 | validated | true | — |
| 27.1 | validated | true | — |
| 27.2 | validated | true | — |
| 27.3 | validated | true | — |
