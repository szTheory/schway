---
schema_version: 1
open_count: 1
waived_count: 0
fixed_count: 0
total_count: 1
last_updated: 2026-09-18T02:24:35.473Z
---

# Broken Windows Ledger

> Cross-phase defect register. With `workflow.windows_enforce` enabled, `/gsd-ship` blocks while `open_count > 0`.
> Waive with `gsd-tools windows waive <id> "<reason>"` (reason required).
> Mark fixed with `gsd-tools windows fixed <id>`.

| id | phase | kind | file | line | description | status | reason | recorded_at | resolved_at |
|----|-------|------|------|------|-------------|--------|--------|-------------|-------------|
| 1 | 14 | unrun-verify | .planning/milestones/M002-phases/08-interprocedural-loan-liveness-in-check/08-VALIDATION.md |  | Stale ~60s/~120s estimated-runtime prose (lines 27,39,94 in 08-VALIDATION.md; also 10/11/12/13-VALIDATION.md) not corrected by 14-05 -- out of scope (narrative, not a verification command span); no existing Phase 14 plan owns it, recommend QLT-10 | open |  | 2026-09-18T02:24:35.473Z |  |

````json
[
  {
    "id": 1,
    "kind": "unrun-verify",
    "phase": "14",
    "file": ".planning/milestones/M002-phases/08-interprocedural-loan-liveness-in-check/08-VALIDATION.md",
    "line": null,
    "description": "Stale ~60s/~120s estimated-runtime prose (lines 27,39,94 in 08-VALIDATION.md; also 10/11/12/13-VALIDATION.md) not corrected by 14-05 -- out of scope (narrative, not a verification command span); no existing Phase 14 plan owns it, recommend QLT-10",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-09-18T02:24:35.473Z",
    "resolved_at": null
  }
]
````
