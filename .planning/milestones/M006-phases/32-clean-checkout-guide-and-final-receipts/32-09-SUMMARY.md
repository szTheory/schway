---
phase: 32-clean-checkout-guide-and-final-receipts
plan: 09
subsystem: ci/evidence
tags: [evd-15, freeze, ci, binding-gate]
requires: ["32-08"]
provides:
  - scripts/check-phase32-freeze.sh binding gate (freeze, headSha, origin tip)
  - TestPhase32FreezeGate with armed mismatch controls
  - ci.yml phase32 receipt run and upload in evidence-aggregate
  - TestPhase32CIHostMatrix with armed control
  - FREEZE_COMMIT and FREEZE_DIGEST
affects: [32-10, 32-11]
key-files:
  created:
    - scripts/check-phase32-freeze.sh
    - internal/compiler/session/session_phase32_freeze_test.go
  modified:
    - .github/workflows/ci.yml
    - internal/compiler/session/session_phase32_ci_test.go
key-decisions:
  - "Revision = codedigest: a commit after the freeze that touches no digest path is accepted; SAME_AS arguments are string-compared before any git lookup so a synthetic sha reports sha_mismatch"
requirements-completed: []
status: complete
commits: 2
plan_head_before: 4cdd91d4a75a1eeeafec499b7ede53a71f09071f
plan_head_after: 079e112f81bd9015c66ee47adb396cc46eb6c7fb
actuals:
  tokens: 9000
  tasks: 3
  commits: 2
coverage:
  - id: D-00-D-08c-freeze-binding
    description: "check-phase32-freeze.sh rejects a mismatched headSha, a mismatched origin tip, a post-freeze digest-path change, a non-descendant and a malformed sha, and accepts a planning-only HEAD move with an unchanged codedigest; its path list equals phase32DigestPaths"
    requirement: "EVD-15"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_freeze_gate_test.go#TestFreezeGate"
        status: pass
    human_judgment: false
  - id: D-08b-ci-wiring
    description: "evidence-aggregate runs verify-phase32.sh after the Phase 27 upload and uploads phase32-receipt-os-sha with if-no-files-found: error, with an armed control"
    requirement: "EVD-15"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase32_ci_test.go#TestPhase32CIHostMatrix"
        status: pass
    human_judgment: false
duration: 15min
completed: 2026-10-12
---

# Phase 32 Plan 09: CI wiring, freeze gate and the freeze Summary

Amended 2026-10-13: Phase 33-04 generalized this gate to `TestFreezeGate`; the historical coverage reference now points to its current source.

Added an armed freeze binding gate and the hosted Phase 32 receipt step, then froze the code on a fully green tree.

## Accomplishments

- `scripts/check-phase32-freeze.sh`: ordered checks (usage, bad_sha, sha_mismatch, unknown_commit, not_descendant, digest_path_changed_after_freeze). `TestPhase32FreezeGate` drives it in a temp repo with base/freeze/plan/late commits and observes each rejection plus the planning-only acceptance.
- `ci.yml` evidence-aggregate now runs `sh scripts/verify-phase32.sh` with `tee`, and uploads `phase32-receipt-<os>-<sha>` with `if-no-files-found: error`. `TestPhase32CIHostMatrix` pins it; its armed subtest removes the guard and is detected. Phase 27 steps untouched; no `continue-on-error`.

## Freeze

- FREEZE_COMMIT: `079e112f81bd9015c66ee47adb396cc46eb6c7fb` (the Task 2 CI commit; the Task 1 gate commit `7a203016` is its ancestor)
- FREEZE_DIGEST: `c1b4e85df4d7723659c62e8338c0c961c0fd930fb45f84b7c25de1fa718cd227`
- Full `go test ./... -count=1`: green. Session meta set plus `TestValidationRowGradesAreEarnedOverArchivedCorpus`: green.
- Local producer last row: `phase32 host result host=macOS codedigest=c1b4e85d...cd227 tree=clean cases=39 controls=4 programs=checksum,adler_low,adler32_cli,adler32_fixture,byte_index_oob,sum_to_n,fizzbuzz status=pass`
- `sh scripts/check-phase32-freeze.sh 079e112f... 079e112f...`: status=pass; `git status --porcelain` over the digest paths is empty.

## Deviations from Plan

None - plan executed exactly as written. Task 3 produced no code commit; it is this SUMMARY.

## Known Stubs

None.

## Self-Check: PASSED

Created files exist; commits 7a203016 and 079e112f are on the branch.
