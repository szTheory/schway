---
phase: quick
plan: 261002-4wy
subsystem: evidence
tags: [validation-corpus, run-record, provenance]
requires: []
provides:
  - Complete graded Phase 25 validation map with receipt-backed test evidence
  - Fresh provenance-bound validation corpus execution record
affects: [phase16-validation-corpus, phase25-validation]
tech-stack:
  added: []
  patterns: [consumer-derived pair export, argv-preserving evidence producer]
key-files:
  modified:
    - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md
    - testdata/phase16/validation-corpus-run-record.jsonl
    - testdata/phase16/validation-corpus-run-record.manifest.json
decisions:
  - "Set each Phase 25 row to EXERCISED only after its primary evidence had a passing non-skipped test in the fresh batch."
  - "Kept all Non-inertness cells as em dashes because no distinct passing twin is claimed."
  - "Discarded failed and stale producer outputs; bound only the final batch for the corrected map and current HEAD."
metrics:
  duration: 2h52m
  completed: 2026-10-02
  status: complete
actuals:
  tasks: 3
  commits: 0
---

# Quick 261002-4wy Summary

**The ten-row Phase 25 validation map now uses receipt-backed EXERCISED grades, and the current 34-pair execution record is bound to its exact export and source revision.**

## Performance

- **Workflow span:** approximately 2 hours 52 minutes across the initial attempt and resumed executions.
- **Accepted producer duration:** 58.215968 seconds, measured with a monotonic clock.
- **Tasks:** all three tasks completed; no commits were made in this worker run as directed by the quick workflow orchestrator.
- **Files modified:** the Phase 25 validation map and the two declared evidence artifacts.

## Accomplishments

- Added Grade and Non-inertness columns to the complete ten-row Phase 25 map, preserving the archived table's column order and existing requirements, security decisions, scope boundaries, and statuses.
- Replaced the bare-name evidence cells for T-25-04 and T-25-06 with their complete executable commands and package operands, preserving all eight named controls.
- Set all ten rows to EXERCISED from the accepted record. Every row's primary evidence resolves to at least one passing non-skipped test. All Non-inertness cells remain em dashes; no mutation-killed grade is claimed.
- The eight named T-25-04 and T-25-06 controls all have passing, non-skipped entries in the accepted record.

## Accepted receipt

- The consumer export contains 34 ordered package and pattern pairs in 37,820 exact bytes. Pair SHA-256: `14f14924a5cb4f30c3f84db5a3dbf9b47c3265bd6bd5dbafecd50ccc926a3b93`.
- The producer ran with the writable Go cache inherited by child processes. Its raw output is 3,124,525 bytes with SHA-256 `afc2f50c524d4389167b1a5f9ec9f5a5c17816151fc876e1b0d078cc935c36fd`.
- The batch completed in 58.215968 seconds, emitted exactly 34 pair witnesses and one final batch witness for 34 pairs, recorded zero test failures, and had empty producer stderr and child stderr sidecar.
- The manifest records schema `phase16-validation-corpus-run-record/v1`, revision `4cd95136647dcb2678563c29f66cdcf2b2321ee2`, completion, production time `2026-10-02T11:02:18.621Z`, both exact digests, and elapsed `58.215968s`.
- Re-export after the Grade-only edits was byte-for-byte identical to the original exported pair array.

## Verification

- The consumer pair-export check passed.
- The isolated pair-export, archived-corpus grade, and anchored-name tests passed.
- The focused groundedness frontier and three-class tests passed.
- The exact uncached full Go suite passed; the session package completed in 51.368 seconds, and every other reported package passed.
- The no-test-weakening source-scope check passed. Whitespace checks passed, and the implementation diff contains only the Phase 25 validation map and the two evidence artifacts.
- These checks passed before summary creation. The post-summary focused and full-suite reruns are the final gate; their results will be returned without further document edits.

## Discarded attempts and stale receipts

- The initial producer invocation omitted the writable cache from its child environment. All 34 child cases failed while accessing the restricted default cache; its output and 0.749656-second timing were discarded.
- A later successful 34-pair batch took 117.055149 seconds but used the earlier pair export, before the Phase 25 map was corrected. Its pair digest was `da7a697f964ffc814b8c32a25bfbb4db1460533a61d9ea04ba5b98a0dc885225`; it is stale and was replaced by the accepted batch.
- The first corrected-map batch used the writable cache and produced all completion witnesses, but its record contained six groundedness failures because the new map commands had not yet been pinned. Its 91.032961-second output was rejected and never bound to a manifest.
- The P25 pin/owner quick task was subsequently committed and verified. The fresh accepted batch was run at that current HEAD after both groundedness checks passed.

## Commits

None in this worker run. The quick workflow orchestrator will commit planning metadata separately.

## Self-Check: PASSED

The corrected validation map, JSONL record, manifest, and summary exist. On-disk pair and record digests match the manifest; the JSONL has the exact 34 pair witnesses, one final batch witness, and zero test failure actions.
