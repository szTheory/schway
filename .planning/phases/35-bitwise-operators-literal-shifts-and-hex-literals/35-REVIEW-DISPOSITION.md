---
phase: 35
review: 35-REVIEW.md
titles: json
findings:
  - id: WR-01
    severity: warning
    disposition: fixed
    title: "Formatter joins separated right-angle tokens into a shift operator"
  - id: WR-02
    severity: warning
    disposition: fixed
    title: "Parenthesized negative shift amounts miss the shift refusal"
open: 0
total: 2
recorded: 2026-10-14T22:31:10.825Z
---

# Phase 35: Code Review Disposition

| Finding | Severity | Disposition | Source |
|---------|----------|-------------|--------|
| WR-01 | warning | fixed | 35-REVIEW-FIX.md (not in the current review) |
| WR-02 | warning | fixed | 35-REVIEW-FIX.md (not in the current review) |

Dispositions: `open` (recorded, not yet triaged), `fixed`, `skipped`, `deferred`.
Set `deferred` by hand and put the reason in the Source cell; both are preserved. A `|` in the reason is kept as prose and escaped on the next run.
Re-running the gate keeps every row it can. A row the current review no longer reports is kept and its Source cell flagged, so a finding does not leave this record silently. ONE exception: when a finding id is REUSED by a different finding, the earlier decision cannot keep a row — the id is taken — and it is dropped. A RECORDED decision (anything but `open`) is named on the console when that happens; a row still at `open` is replaced silently, because `open` records no decision to lose.
