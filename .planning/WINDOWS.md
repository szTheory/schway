---
schema_version: 1
open_count: 4
waived_count: 0
fixed_count: 7
total_count: 11
last_updated: 2026-10-15T01:54:05.332Z
---

# Broken Windows Ledger

> Cross-phase defect register. With `workflow.windows_enforce` enabled, `/gsd-ship` blocks while `open_count > 0`.
> Waive with `gsd-tools windows waive <id> "<reason>"` (reason required).
> Mark fixed with `gsd-tools windows fixed <id>`.

| id | phase | kind | file | line | description | status | reason | recorded_at | resolved_at |
|----|-------|------|------|------|-------------|--------|--------|-------------|-------------|
| 1 | 14 | unrun-verify | .planning/milestones/M002-phases/08-interprocedural-loan-liveness-in-check/08-VALIDATION.md |  | Stale ~60s/~120s estimated-runtime prose (lines 27,39,94 in 08-VALIDATION.md; also 10/11/12/13-VALIDATION.md) not corrected by 14-05 -- out of scope (narrative, not a verification command span); no existing Phase 14 plan owns it, recommend QLT-10 | open |  | 2026-09-18T02:24:35.473Z |  |
| 2 | 16 | unrun-verify | internal/compiler/cgen/cgen_n1_convergence_test.go |  | Package-wide cgen verification remains red until Plan 16-06 flips its intentional pre-cutover expectations. | fixed |  | 2026-09-19T21:48:08.202Z | 2026-09-24T06:37:27.070Z |
| 3 | 16 | unrun-verify | .planning/phases/16-branch-match-emitter-port/16-19-PLAN.md |  | Full go test ./... ran but failed on pre-existing Phase 11 fixture, corpus, and validation-groundedness drift outside this plan scope. | fixed |  | 2026-09-23T22:41:36.175Z | 2026-09-24T06:37:27.152Z |
| 4 | 33 | deviation | internal/compiler/interp/interp.go |  | Plan 33-02 acceptance expected two live-resource error strings; source has three because RunWithForeignOutcomes and both routed callers retain the same required text. | open |  | 2026-10-13T17:04:15.066Z |  |
| 5 | 33 | deviation | scripts/check-receipts.sh |  | Preserved first-failure refusal ordering by resolving the expected digest before manifest-content checks. | fixed |  | 2026-10-13T17:23:51.848Z | 2026-10-13T17:26:32.776Z |
| 6 | 33 | deviation | .planning/milestones/M006-phases/32-clean-checkout-guide-and-final-receipts/*-SUMMARY.md; .planning/milestones/M006-phases/32.1-address-tech-debt-cli-exit-classes-and-module-aware-routes/*-SUMMARY.md |  | Historical Phase 32 and 32.1 coverage refs amended to the Phase 33-04 generic gate test names so groundedness resolves. | open |  | 2026-10-13T17:55:33.809Z |  |
| 7 | 34 | deviation | internal/compiler/session/session_phase34_prechange_test.go |  | Comparator now compares not_emitted outcomes and reaches a status-mutation control. | open |  | 2026-10-13T21:34:52.574Z |  |
| 8 | 34 | deviation | internal/compiler/surfacemanifest/manifest.go |  | Removed Box/Pair ability-only and linear-if contexts from the surface manifest after CheckCommandFile rejected their complete programs. | fixed |  | 2026-10-14T00:24:46.705Z | 2026-10-14T00:24:48.501Z |
| 9 | 36 | deviation | internal/compiler/session/session_phase36_prechange_test.go |  | Reject unsafe new-fixture ledger paths before file reads; traversal control passes. | fixed |  | 2026-10-15T01:31:06.038Z | 2026-10-15T01:31:07.387Z |
| 10 | 36 | deviation | internal/compiler/check/check_phase34_refusal_test.go |  | Updated Phase 34's stale one-argument scalar-call rejection assertion to the Phase 36 source contract while retaining its two-argument refusal. | fixed |  | 2026-10-15T01:53:35.425Z | 2026-10-15T01:54:05.136Z |
| 11 | 36 | deviation | internal/compiler/syntax/parser.go |  | Kept byte_len/byte_at builtin initializer routing ahead of named-call lookahead so nested byte_at indexes parse as source calls. | fixed |  | 2026-10-15T01:54:00.791Z | 2026-10-15T01:54:05.332Z |

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
  },
  {
    "id": 2,
    "kind": "unrun-verify",
    "phase": "16",
    "file": "internal/compiler/cgen/cgen_n1_convergence_test.go",
    "line": null,
    "description": "Package-wide cgen verification remains red until Plan 16-06 flips its intentional pre-cutover expectations.",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-09-19T21:48:08.202Z",
    "resolved_at": "2026-09-24T06:37:27.070Z",
    "milestone": null
  },
  {
    "id": 3,
    "kind": "unrun-verify",
    "phase": "16",
    "file": ".planning/phases/16-branch-match-emitter-port/16-19-PLAN.md",
    "line": null,
    "description": "Full go test ./... ran but failed on pre-existing Phase 11 fixture, corpus, and validation-groundedness drift outside this plan scope.",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-09-23T22:41:36.175Z",
    "resolved_at": "2026-09-24T06:37:27.152Z",
    "milestone": null
  },
  {
    "id": 4,
    "kind": "deviation",
    "phase": "33",
    "file": "internal/compiler/interp/interp.go",
    "line": null,
    "description": "Plan 33-02 acceptance expected two live-resource error strings; source has three because RunWithForeignOutcomes and both routed callers retain the same required text.",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-10-13T17:04:15.066Z",
    "resolved_at": null,
    "milestone": null
  },
  {
    "id": 5,
    "kind": "deviation",
    "phase": "33",
    "file": "scripts/check-receipts.sh",
    "line": null,
    "description": "Preserved first-failure refusal ordering by resolving the expected digest before manifest-content checks.",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-10-13T17:23:51.848Z",
    "resolved_at": "2026-10-13T17:26:32.776Z",
    "milestone": null
  },
  {
    "id": 6,
    "kind": "deviation",
    "phase": "33",
    "file": ".planning/milestones/M006-phases/32-clean-checkout-guide-and-final-receipts/*-SUMMARY.md; .planning/milestones/M006-phases/32.1-address-tech-debt-cli-exit-classes-and-module-aware-routes/*-SUMMARY.md",
    "line": null,
    "description": "Historical Phase 32 and 32.1 coverage refs amended to the Phase 33-04 generic gate test names so groundedness resolves.",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-10-13T17:55:33.809Z",
    "resolved_at": null,
    "milestone": null
  },
  {
    "id": 7,
    "kind": "deviation",
    "phase": "34",
    "file": "internal/compiler/session/session_phase34_prechange_test.go",
    "line": null,
    "description": "Comparator now compares not_emitted outcomes and reaches a status-mutation control.",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-10-13T21:34:52.574Z",
    "resolved_at": null,
    "milestone": null
  },
  {
    "id": 8,
    "kind": "deviation",
    "phase": "34",
    "file": "internal/compiler/surfacemanifest/manifest.go",
    "line": null,
    "description": "Removed Box/Pair ability-only and linear-if contexts from the surface manifest after CheckCommandFile rejected their complete programs.",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-10-14T00:24:46.705Z",
    "resolved_at": "2026-10-14T00:24:48.501Z",
    "milestone": null
  },
  {
    "id": 9,
    "kind": "deviation",
    "phase": "36",
    "file": "internal/compiler/session/session_phase36_prechange_test.go",
    "line": null,
    "description": "Reject unsafe new-fixture ledger paths before file reads; traversal control passes.",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-10-15T01:31:06.038Z",
    "resolved_at": "2026-10-15T01:31:07.387Z",
    "milestone": null
  },
  {
    "id": 10,
    "kind": "deviation",
    "phase": "36",
    "file": "internal/compiler/check/check_phase34_refusal_test.go",
    "line": null,
    "description": "Updated Phase 34's stale one-argument scalar-call rejection assertion to the Phase 36 source contract while retaining its two-argument refusal.",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-10-15T01:53:35.425Z",
    "resolved_at": "2026-10-15T01:54:05.136Z",
    "milestone": null
  },
  {
    "id": 11,
    "kind": "deviation",
    "phase": "36",
    "file": "internal/compiler/syntax/parser.go",
    "line": null,
    "description": "Kept byte_len/byte_at builtin initializer routing ahead of named-call lookahead so nested byte_at indexes parse as source calls.",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-10-15T01:54:00.791Z",
    "resolved_at": "2026-10-15T01:54:05.332Z",
    "milestone": null
  }
]
````
