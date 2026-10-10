---
milestone: M005
milestone_name: Practical Computation
audited: 2026-10-10T00:00:00Z
audited_head: b6cbcc2
evidence_revision: 6224f8b2c9c15f54270a232fab2bce12628c7f19
status: tech_debt
scores:
  requirements: 9/9
  phases: 2/2
  integration: 12/12 cross-phase links connected (1 orphaned export, 2 pre-existing peer-coverage gaps)
  flows: 3/6 complete (3 partial, none breaking a requirement)
gaps:
  requirements: []
  integration: []
  flows: []
process_gaps:
  - phase: 27-exact-fizzbuzz-and-cross-host-evidence
    item: "27-SECURITY.md missing while workflow.security_enforcement=true (26-SECURITY.md exists: verified, 0 open)"
    action: "/gsd-secure-phase 27"
  - phase: 27-exact-fizzbuzz-and-cross-host-evidence
    item: "27-VALIDATION.md not reconciled: status 'complete' (not 'validated'), all five requirement rows still '⬜ pending', approval pending"
    action: "/gsd-validate-phase 27"
nyquist:
  compliant_phases: ["26"]
  partial_phases: []
  not_validated_phases: ["27"]
  missing_phases: []
  overall: partial
tech_debt:
  - phase: cross-phase (M004 route reused by M005)
    items:
      - "F1: session.BuildApplication (internal/compiler/session/session.go:1382) runs corevalidate only, not originvalidate; the public `schway build` artifacts for sum_to_n and fizzbuzz are not origin-peer-checked on that route. The origin peer runs on check, run-native and interpreter-app routes. Pre-existing since Phase 22 (cd90d36), not an M005 regression. Affects U64-01/02, FLOW-01/02, APP-07/08 assurance depth."
      - "F3: `schway app run --evidence=events` capture (native_app.go RunApplicationWithEvidence) is not peer-validated in-route; executionpeer is applied by `app verify` and run-native comparison, and by Phase 26 source event tests. Consistent with the documented boundary that capture never self-claims Verified. (EVD-12)"
      - "F4: checker, corevalidate and originvalidate use different fixed-point budget counting (per-op work vs sweep+transfer), all at 65536 and all fail-closed; acceptance boundaries could diverge on large programs. (FLOW-01/02)"
      - "F6: executionpeer.ValidateFullCoverage (executionpeer.go:85) has no production caller."
  - phase: 27-exact-fizzbuzz-and-cross-host-evidence
    items:
      - "F2: scripts/check-phase27-receipts.sh defaults expected revision to `git rev-parse HEAD`; at any later HEAD the guide's documented command fails closed with wrong_revision. examples/phase27/README.md does not mention PHASE27_EXPECTED_REVISION (only 27-HOST-EVIDENCE.md does). Fails closed, so no false pass. (EVD-12, DX-16)"
      - "F5: the 65,536-byte output ceiling is unreachable from accepted inputs (max FizzBuzz output 4,673 B at n=1000; sum_to_n ≤ 7 B); the ceiling is exercised only via test seams. (APP-09)"
      - "27-09 and 27-10 SUMMARY frontmatter lack requirements_completed; DX-16 appears in no SUMMARY frontmatter (manually confirmed satisfied: guide, root README link, TestPhase27GuideContract, executed commands)."
      - "Review IN-01 open: output-write failure conflates fwrite errors with limit exhaustion."
      - "Review IN-02 open: duplicated reason literals across layers."
      - "examples/fizzbuzz.schway and examples/sum_to_n.schway fail `schway format --check` (format.non_canonical); pre-existing hand-written style, not claimed by any must-have."
  - phase: 26-checked-scalar-sum
    items:
      - "26-05-PLAN wording says capture is 'Verified=true'; the report boolean intentionally stays false (peer verdict is separate). Verifier recommends a dated planning amendment."
      - "Focused Phase 26 hosted job not re-dispatched at 6224f8b; current behavioral evidence is the full/race suite run 38054848320 plus earlier focused run 37118144404."
---

# Milestone M005 — Practical Computation: Audit

**Audited HEAD:** b6cbcc2 (non-planning source identical to hosted evidence revision 6224f8b)
**Status:** tech_debt — all 9 requirements satisfied, no critical blockers; two process artifacts for Phase 27 are outstanding and several deferred items need review.

## Phases

| Phase | Verification | Score | Validation (Nyquist) | Security | Review |
|---|---|---|---|---|---|
| 26 Checked Scalar Sum | passed (re-verified after Phase 27) | 5/5 | validated, compliant | verified, 0 open | clean |
| 27 Exact FizzBuzz and Cross-Host Evidence | passed (re-verified after review fixes) | 5/5 | **not validated** (status `complete`, rows pending) | **missing** | WR-01..05 fixed; IN-01, IN-02 open |

## Requirements (3-source cross-reference)

| REQ | Phase | VERIFICATION | SUMMARY frontmatter | REQUIREMENTS.md | Final |
|---|---|---|---|---|---|
| U64-01 | 26 | passed | 26-01, 02, 05 | [x] | satisfied |
| U64-02 | 27 | passed | 27-03, 04, 06, 08 | [x] | satisfied |
| FLOW-01 | 26 | passed | 26-01, 03, 04, 05 | [x] | satisfied |
| FLOW-02 | 26 | passed | 26-01, 03, 04, 05 | [x] | satisfied |
| APP-07 | 26 | passed | 26-01, 02, 04, 05 | [x] | satisfied |
| APP-08 | 27 | passed | 27-01, 02, 05, 06, 07, 08 | [x] | satisfied |
| APP-09 | 27 | passed | 27-01, 02, 06, 07, 08 | [x] | satisfied |
| EVD-12 | 27 | passed | 27-07, 08 | [x] | satisfied |
| DX-16 | 27 | passed | none (27-09/10 frontmatter empty) | [x] | satisfied — matrix says *partial*; manually confirmed (guide exists, root-linked, `TestPhase27GuideContract`, commands executed by verifier and integration checker) |

No orphaned requirements. No unsatisfied requirements.

## Integration

Integration checker built `schway` and ran the public routes live:
- FizzBuzz 15 and 1000 byte-identical to committed fixtures; n=0 empty stdout.
- sum_to_n 0/10/1000 → 0/55/500500.
- Both programs at 1001 → exit 65, empty stdout, `schway app: input exceeds 1000`.
- Focused `TestPhase27DynamicZero`, `OutputExactLimit`, `OutputCrossingWrite`, `OutputPublicStatus` pass; build and vet clean.

Every Phase 27 op (remainder, equality, write_u64/newline/text) is handled in checker, corevalidate, originvalidate, interpreter and C emitter, reusing Phase 26's scalar CFG fixed-point, checked arithmetic and app route rather than forking them. The remainder-by-zero guard is emitted before the C `%`. The U64 auto-print counts against the shared output limit on both engines. The receipt driver covers both programs (10 cases, 3 controls) and runs in both CI host lanes.

| Flow | Status |
|---|---|
| source → check → corevalidate → C17 → build → app run | complete (live) |
| input guard 1001 | complete (live) |
| overflow / dynamic-zero controls | complete (focused tests) |
| native build → origin peer | partial (F1, pre-existing) |
| app-run evidence → executionpeer | partial (F3, by design) |
| committed receipts → gate via guide command | partial (F2, fails closed at later HEAD) |

Spot-checked by the auditor: F1 (`BuildApplication` calls only `corevalidate.Validate`) and F2 (gate default `git rev-parse HEAD`; guide omits the override) confirmed.

## Tech debt

See the `tech_debt` frontmatter: 4 cross-phase items, 6 Phase 27 items, 2 Phase 26 items.

## Nyquist coverage

| Phase | VALIDATION.md | Compliant | Action |
|---|---|---|---|
| 26 | validated | true | — |
| 27 | exists, unreconciled | not authoritative | `/gsd-validate-phase 27` |
