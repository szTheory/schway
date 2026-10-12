---
phase: 30
review: 30-REVIEW.md
titles: json
findings:
  - id: CR-01
    severity: critical
    disposition: fixed
    title: "Cap diagnostics (count, depth, total bytes) are swallowed; users get `module.imports_not_loaded`"
  - id: WR-01
    severity: warning
    disposition: fixed
    title: "Entry self-import (and any alias-to-entry) inflates the module count before it is known to be the entry"
  - id: WR-02
    severity: warning
    disposition: fixed
    title: "Repeated cap diagnostics for the same refused target"
  - id: WR-03
    severity: warning
    disposition: fixed
    title: "An imported name that is exported but not declared passes `check`, leaving a dangling import fact"
  - id: WR-04
    severity: warning
    disposition: fixed
    title: "Import name that collides with a foreign symbol of the entry is not refused at the import"
  - id: WR-05
    severity: warning
    disposition: fixed
    title: "Peer-refusal unit is taken from the first `s1:` ID in the detail, which is the callee for the visibility codes"
  - id: WR-06
    severity: warning
    disposition: fixed
    title: "Engine-level error classification: all non-NotExist `Lstat`/`Open` failures are reported as \"resolves outside the entry directory\""
  - id: IN-01
    severity: info
    disposition: deferred
    title: "Test-only seams are read on production hot paths as unsynchronized globals"
  - id: IN-02
    severity: info
    disposition: deferred
    title: "FIFO control test can leak a blocked goroutine and its seam on failure"
open: 0
total: 8
recorded: 2026-10-12T00:00:00Z
---

# Phase 30: Code Review Disposition

| Finding | Severity | Disposition | Source |
|---------|----------|-------------|--------|
| CR-01 | critical | fixed | 11d64746 (session and CLI route tests; reverting the fix makes them fail) |
| WR-01 | warning | fixed | 105f71b5 |
| WR-02 | warning | fixed | 9f7962fb |
| WR-03 | warning | fixed | 6df11250 |
| WR-04 | warning | fixed | 05fade38 (foreign symbols; data constructors have no collision path) |
| WR-05 | warning | fixed | 26e905e3 (structured `Problem.FunctionID`; phase16 inventory line re-pinned 3089→3098) |
| WR-06 | warning | fixed | 093c2fa5 (matches the os.Root escape text, because Go 1.24 has no sentinel; unrecognized errors fail closed as `module.read_failed`) |
| IN-01 | info | deferred | Same class as Phase 29 WR-04 (seams in production packages); hardening them behind a test-only build mechanism is a follow-up |
| IN-02 | info | deferred | Test hygiene only; the control still observes its seam on success. Fix when that test file is next touched |

Dispositions: `open` (recorded, not yet triaged), `fixed`, `skipped`, `deferred`.
Set `deferred` by hand and put the reason in the Source cell; both are preserved. A `|` in the reason is kept as prose and escaped on the next run.
Fixes were applied by `/gsd-code-review 30 --fix` (gsd-code-fixer, scope critical+warning). Details are in 30-REVIEW-FIX.md.
