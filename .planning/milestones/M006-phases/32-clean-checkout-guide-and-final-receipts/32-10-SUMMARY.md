---
phase: 32-clean-checkout-guide-and-final-receipts
plan: 10
subsystem: ci/evidence
tags: [evd-15, freeze, hosted-dispatch, checkpoint]
requires: ["32-09"]
provides:
  - RUN_ID of the single hosted dispatch bound to the frozen codedigest
affects: [32-11]
key-files:
  created: []
  modified: []
key-decisions:
  - "The user pushed and dispatched once; the orchestrator bound the run with check-phase32-freeze.sh and armed it at use with a synthetic headSha"
requirements-completed: []
status: complete
commits: 1
actuals:
  tasks: 3
  commits: 1
coverage:
  - id: D-00-D-08c-run-bound-to-freeze
    description: "The dispatched run's headSha, the pushed HEAD and the origin tip are one commit with no digest-path change since FREEZE_COMMIT; a synthetic headSha is rejected with sha_mismatch"
    requirement: "EVD-15"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_freeze_gate_test.go#TestFreezeGate"
        status: pass
    human_judgment: false
duration: 5min
completed: 2026-10-12
---

# Phase 32 Plan 10: Hosted dispatch for the frozen M006 revision

Amended 2026-10-13: Phase 33-04 generalized the freeze gate to `TestFreezeGate`; the historical coverage reference now points to its current source.

The single human step ran. The user pushed `phase26-validation` and dispatched `ci.yml` once. The resulting run is bound to the freeze from 32-09.

## Freeze (from 32-09)

- FREEZE_COMMIT: `079e112f81bd9015c66ee47adb396cc46eb6c7fb`
- FREEZE_DIGEST: `c1b4e85df4d7723659c62e8338c0c961c0fd930fb45f84b7c25de1fa718cd227`
- PUSHED_HEAD: `6c3042713fadab198d827cb259eb7705431055e1`

## Task 1: freeze re-proved before the push

- Branch `phase26-validation`. No digest path is dirty. `sh scripts/codedigest.sh` equals FREEZE_DIGEST.
- `phase32 freeze gate: status=pass freeze=079e112f81bd9015c66ee47adb396cc46eb6c7fb candidate=6c3042713fadab198d827cb259eb7705431055e1`
- Armed control 1, a synthetic headSha: `phase32 freeze gate: status=fail reason=sha_mismatch`, exit 1.
- Armed control 2, the commit before the freeze used as the base: `phase32 freeze gate: status=fail reason=digest_path_changed_after_freeze`, exit 1.
- Session meta set: ok.
- The push sent 41 commits, from `origin/phase26-validation` to PUSHED_HEAD.

## Task 2: human action

The user ran `! git push -q origin phase26-validation && gh workflow run ci.yml --ref phase26-validation`. It printed https://github.com/szTheory/schway/actions/runs/38315799626.

## Task 3: run bound to the freeze

- RUN_ID: `38315799626`. URL: https://github.com/szTheory/schway/actions/runs/38315799626
- Event `workflow_dispatch`, headSha `6c3042713fadab198d827cb259eb7705431055e1`, which equals PUSHED_HEAD.
- `git rev-parse origin/phase26-validation` also equals PUSHED_HEAD.
- Only one dispatched run exists for PUSHED_HEAD, so it cannot be cancelled by a second dispatch.
- Binding: `sh scripts/check-phase32-freeze.sh FREEZE_COMMIT 6c304271… PUSHED_HEAD origin` printed `phase32 freeze gate: status=pass freeze=079e112f81bd9015c66ee47adb396cc46eb6c7fb candidate=6c3042713fadab198d827cb259eb7705431055e1`.
- Armed at use with a synthetic headSha: `phase32 freeze gate: status=fail reason=sha_mismatch`, exit 1.

## Deviations

The orchestrator ran Tasks 1 and 3 inline. This checkpoint plan has no code work.

## Next

32-11 waits for run 38315799626 to finish. It then downloads both `phase32-receipt-*` artifacts and ingests them into `32-RECEIPT.json` as a `.planning`-only change.
