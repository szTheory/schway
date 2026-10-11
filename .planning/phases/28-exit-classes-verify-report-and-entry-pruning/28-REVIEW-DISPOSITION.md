---
phase: 28
review: 28-REVIEW.md
titles: json
findings:
  - id: CR-01
    severity: critical
    disposition: fixed
    title: "Emitter consumer registry line number is stale, so the inventory test fails at HEAD"
  - id: WR-01
    severity: warning
    disposition: deferred
    title: "Registry pins are fragile and were re-grounded blind"
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
total: 7
recorded: 2026-10-11T00:00:00.000Z
---

# Phase 28: Code Review Disposition

| Finding | Severity | Disposition | Source |
|---------|----------|-------------|--------|
| CR-01 | critical | fixed | eface8a (re-pinned `session.go:EmitNative` 3047→3042; the shift came from 28-08 removing 5 lines, not 307e14a) |
| WR-01 | warning | deferred | Process finding, not a code defect. Line-number pins are by design in the Phase 16 emitter inventory; the mitigation is to run the session meta/inventory tests after any `session.go` edit, which the orchestrator now does (verified passing at eface8a). Structural fix (symbol-based pins) left for a future phase. |
| WR-02 | warning | fixed | 28-REVIEW-FIX.md (first review) |
| WR-03 | warning | fixed | 28-REVIEW-FIX.md (first review) |
| WR-04 | warning | fixed | 28-REVIEW-FIX.md (first review) |
| IN-01 | info | fixed | 28-REVIEW-FIX.md (first review) — not re-reported by the 2026-10-11 re-review |
| IN-02 | info | fixed | 28-REVIEW-FIX.md (first review) — not re-reported by the 2026-10-11 re-review |

The first review's WR-01 ("Duplicate-`main` ambiguity reports the wrong roots and a misleading hint", fixed in 28-REVIEW-FIX.md) lost its row because the re-review reused the id WR-01 for a different finding.

Dispositions: `open` (recorded, not yet triaged), `fixed`, `skipped`, `deferred`.
Set `deferred` by hand and put the reason in the Source cell; both are preserved. A `|` in the reason is kept as prose and escaped on the next run.
