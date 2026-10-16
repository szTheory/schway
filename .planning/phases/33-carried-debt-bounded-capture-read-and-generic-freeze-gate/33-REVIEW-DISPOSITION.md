---
phase: 33
review: 33-REVIEW.md
titles: json
findings:
  - id: CR-02
    severity: critical
    disposition: fixed
    title: "Execution peer accepts terminal evidence that contradicts the reported outcome [CLOSED]"
  - id: CR-03
    severity: critical
    disposition: deferred
    title: "Legacy child defects leave the caller activation open"
  - id: CR-04
    severity: critical
    disposition: deferred
    title: "Propagated typed failure from OpReturn is unclassifiable"
  - id: CR-05
    severity: critical
    disposition: fixed
    title: "Dynamic native scalar callers emit application-only output for explicit defects [CLOSED]"
  - id: CR-06
    severity: critical
    disposition: deferred
    title: "Legacy no-call native scalar defects still exit as application failures"
  - id: WR-01
    severity: warning
    disposition: fixed
    title: "Coverage checker accepts incomplete verification records"
  - id: WR-02
    severity: warning
    disposition: fixed
    title: "Receipt gate allows undeclared case programs"
  - id: WR-03
    severity: warning
    disposition: fixed
    title: "Empty case program bypasses the reverse inventory check"
  - id: WR-04
    severity: warning
    disposition: fixed
    title: "Coverage checker still accepts incomplete entry metadata"
  - id: CR-01
    severity: critical
    disposition: fixed
    title: "Empty manifest controls bypass receipt control validation"
open: 0
total: 10
recorded: 2026-10-16T02:53:36.270Z
---

# Phase 33: Code Review Disposition

| Finding | Severity | Disposition | Source |
|---------|----------|-------------|--------|
| CR-02 | critical | fixed | 33-REFRESH2-FIX.md; independent four-file closure review |
| CR-03 | critical | deferred | Execution-peer maintainer via M007 audit before Phase38 freeze; pre-existing nested legacy-defect false rejection outside the Phase33 capture/predicate/gate witnesses |
| CR-04 | critical | deferred | Execution-peer maintainer via M007 audit before Phase38 freeze; pre-existing modeled-failure propagation requires its own contract/edge evidence rather than broadening terminal classification without proof |
| CR-05 | critical | fixed | 33-REFRESH3-FIX.md; independent final three-file closure, dynamic scalar caller scope; renamed title does not undo closure |
| CR-06 | critical | deferred | Emitter maintainer through mandatory M007 audit before Phase38 freeze; original/current no-call native defect probes reach exit65; resolve against frozen-C obligation through scoped gap work and dated compatibility amendment when required |
| WR-01 | warning | fixed | 33-REVIEW-FIX.md (not in the current review) |
| WR-02 | warning | fixed | 33-REVIEW-FIX.md (not in the current review) |
| WR-03 | warning | fixed | 33-REVIEW-FIX.md (not in the current review) |
| WR-04 | warning | fixed | 33-REVIEW-FIX.md (not in the current review) |
| CR-01 | critical | fixed | 33-REVIEW-FIX.md (not in the current review) |

Dispositions: `open` (recorded, not yet triaged), `fixed`, `skipped`, `deferred`.
Set `deferred` by hand and put the reason in the Source cell; both are preserved. A `|` in the reason is kept as prose and escaped on the next run.
Re-running the gate keeps every row it can. A row the current review no longer reports is kept and its Source cell flagged, so a finding does not leave this record silently. ONE exception: when a finding id is REUSED by a different finding, the earlier decision cannot keep a row — the id is taken — and it is dropped. A RECORDED decision (anything but `open`) is named on the console when that happens; a row still at `open` is replaced silently, because `open` records no decision to lose.
