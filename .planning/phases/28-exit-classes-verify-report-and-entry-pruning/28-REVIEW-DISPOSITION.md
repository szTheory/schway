---
phase: 28
review: 28-REVIEW.md
titles: json
findings:
  - id: WR-01
    severity: warning
    disposition: fixed
    title: "Duplicate-`main` ambiguity reports the wrong roots and a misleading hint"
  - id: WR-02
    severity: warning
    disposition: fixed
    title: "A source refusal can be masked as exit 3 when the report write fails"
  - id: WR-03
    severity: warning
    disposition: fixed
    title: "A stale report file survives a refused or oversize run"
  - id: WR-04
    severity: warning
    disposition: fixed
    title: "A non-diagnostic `resolveEntry` error is silently tolerated in app verify"
  - id: IN-01
    severity: info
    disposition: fixed
    title: "`appRunExitCode` collides with child exit statuses by design"
  - id: IN-02
    severity: info
    disposition: fixed
    title: "`Phase4CheckedProgram` discards the pruned program"
open: 0
total: 6
recorded: 2026-10-11T03:04:10.516Z
---

# Phase 28: Code Review Disposition

| Finding | Severity | Disposition | Source |
|---------|----------|-------------|--------|
| WR-01 | warning | fixed | 28-REVIEW-FIX.md |
| WR-02 | warning | fixed | 28-REVIEW-FIX.md |
| WR-03 | warning | fixed | 28-REVIEW-FIX.md |
| WR-04 | warning | fixed | 28-REVIEW-FIX.md |
| IN-01 | info | fixed | 28-REVIEW-FIX.md |
| IN-02 | info | fixed | 28-REVIEW-FIX.md |

Dispositions: `open` (recorded, not yet triaged), `fixed`, `skipped`, `deferred`.
Set `deferred` by hand and put the reason in the Source cell; both are preserved. A `|` in the reason is kept as prose and escaped on the next run.
Re-running the gate keeps every row it can. A row the current review no longer reports is kept and its Source cell flagged, so a finding does not leave this record silently. ONE exception: when a finding id is REUSED by a different finding, the earlier decision cannot keep a row — the id is taken — and it is dropped. A RECORDED decision (anything but `open`) is named on the console when that happens; a row still at `open` is replaced silently, because `open` records no decision to lose.
