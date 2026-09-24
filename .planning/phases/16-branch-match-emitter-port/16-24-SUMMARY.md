---
phase: 16-branch-match-emitter-port
plan: 24
status: complete
completed: 2026-09-24
---

# Plan 16-24 Summary: Close Validation Evidence Frontier Gaps

Updated `allPrimaryValidationRows` to read validation status and primary table schema. Validated, partial, and complete documents always participate in evidence grading. Draft or planned documents participate once their primary table has adopted the Grade column, preserving existing graded drafts such as Phase 14 while deferring pre-schema Phase 17's pending rows. Added regression checks for these eligibility rules.

Corrected Phase 4's stale test references after the M004 cut. The affected rows now describe the live public-evidence refusal/unreachability behavior and cite existing tests; they no longer claim the retired sidecar-digest path remains an active capability.

Regenerated the corpus pair export and run record after these changes. The final set contains 31 consumer-derived package-pattern pairs and a completed sequential run. The manifest binds the current pair bytes and JSONL body, with 31 pair completion witnesses and one batch completion witness.

## Verification

- Focused draft/schema eligibility, Phase 14 row-floor, Phase 4 current controls, validation grade, run-record integrity, and 17-site lane-pin checks passed.
- `GOCACHE=/tmp/ai-lang-go-cache go test ./...` passed across all packages.
- `GOCACHE=/tmp/ai-lang-go-cache go test -race ./...` passed across all packages.

## Deviations

The initial status-only draft filter also hid Phase 14, an older draft with a fully graded table and an explicit permanent row-floor guard. The selector was narrowed to defer only drafts without a Grade schema. This preserves older completed evidence while excluding ungraded future work.
