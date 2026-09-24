---
phase: 16-branch-match-emitter-port
plan: 17
plan_head_before: b470e7c0ed937d4d1164108857a3a617bd54ee03
commits: 2
subsystem: testing
tags: [validation-corpus, run-record, evidence-grade, fail-closed]

requires:
  - phase: 16-18
    provides: Current groundedness frontier and P20 ownership policy
provides:
  - Validation corpus manifest and run record refreshed from the live 31-pair consumer-derived set
  - Persisted-record mutation controls for changed pair bytes, changed JSONL bytes, and missing completion witnesses
affects: [NAT-08, NAT-09, validation-evidence]

actuals:
  tokens: 1031423
  tasks: 2
  commits: 2

tech-stack:
  added: []
  patterns:
    - Run the evidence producer over the exact consumer-derived pair export and bind both pair bytes and raw JSONL bytes by SHA-256
    - Re-digest mutated fixture bodies to prove completion witnesses are independently required

key-files:
  created:
    - .planning/phases/16-branch-match-emitter-port/16-17-SUMMARY.md
  modified:
    - internal/compiler/session/evidence_grade_test.go
    - testdata/phase16/validation-corpus-run-record.jsonl
    - testdata/phase16/validation-corpus-run-record.manifest.json

key-decisions:
  - "Regenerate the evidence record from the live consumer-derived 31-pair export and preserve actual pass/fail/skip output."
  - "Keep pair and batch completion witnesses fail-closed, including when a modified record body is re-bound by a new digest."

requirements-completed: [NAT-08, NAT-09]

coverage:
  - id: D1
    description: The checked-in run record and manifest bind to the current live pair export and contain a complete producer batch.
    requirement: NAT-08
    verification:
      - kind: unit
        ref: "SHA-256 comparison and JSONL witness audit: 31 derived pairs, 31 unique pair witnesses, 1 batch witness"
        status: pass
      - kind: unit
        ref: "go test ./internal/compiler/session -run '^(TestValidationRowGradesAreEarnedOverArchivedCorpus|TestRunRecordCarriesACompletionWitness|TestRunRecordCompletenessGuardIsNotInert)$' -count=1 -v"
        status: fail
    human_judgment: true
    rationale: "The digest and witnesses validate, but the grade test also reaches unrelated existing evidence-row defects in 04-VALIDATION.md and 17-VALIDATION.md."
  - id: D2
    description: Tampered pair/record bytes and missing persisted pair/batch witnesses are independently rejected.
    requirement: NAT-09
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session -run '^TestCheckedInCorpusRecordRejectsTamperingAndVacuity$' -count=1 -v"
        status: pass
    human_judgment: false

duration: 12min
completed: 2026-09-23
status: complete
---

# Phase 16 Plan 17: Live Validation Corpus Evidence Summary

**The corpus evidence now binds to the live 31-pair set and a completed sequential run, with persisted tampering and missing-witness controls.**

## Performance

- **Duration:** About 12 minutes.
- **Started:** Approximately 2026-09-23T19:55:00-04:00.
- **Completed:** 2026-09-23T20:07:39-04:00.
- **Tasks:** 2/2 implemented.
- **Files modified:** 3 planned files.

## Accomplishments

- Exported the exact package-pattern pairs from the validation consumer. The live set contains 31 pairs and hashes to `07e30b53013291e76df6cd4d01d06ddbf4184be0fa33c230000aa9c17d4814cb`.
- Ran `scripts/evidence-run-record.sh` sequentially with `GOCACHE=/tmp/ai-lang-go-cache`. The resulting JSONL contains every actual test result, 31 distinct pair-completion witnesses, and one 31-pair batch-completion witness.
- Updated the manifest to bind the pair export and record bytes. Independent recomputation matched both SHA-256 values; elapsed time was `122.422s`, under the 15-minute cap.
- Added fixture-based refusal cases for changed source pair bytes, changed JSONL bytes, and absent pair or batch completion even when the changed record body is re-digested.

## Task Commits

1. **Task 1: Regenerate corpus run evidence from the live producer** - `69df5de` (`fix`).
2. **Task 2: Exercise refusal paths for stale and incomplete corpus evidence** - `aaa85da` (`test`).

## Files Created/Modified

- `internal/compiler/session/evidence_grade_test.go` - Added persisted-record negative controls while keeping exact digest and completion validation unchanged.
- `testdata/phase16/validation-corpus-run-record.jsonl` - Replaced the stale record with the complete output from the live 31-pair run.
- `testdata/phase16/validation-corpus-run-record.manifest.json` - Updated source revision, pair and record digests, production timestamp, and elapsed duration.

## Decisions Made

- Used the consumer's export seam as the only source of requested pairs, so the producer and validator operate on the same current bytes.
- Kept observed failures and skips in the run record. No success result or completion witness was hand-authored.
- Re-digested altered JSONL in the missing-witness controls so those tests reach and exercise the completion-witness guard rather than failing earlier at the record-digest check.

## Deviations from Plan

No guard was weakened and no out-of-scope validation document was changed. The initial producer attempt used the inaccessible default Go cache and was discarded; the final record was generated with the writable `/tmp/ai-lang-go-cache`.

## Issues Encountered

- The plan's focused grade command reaches the refreshed corpus digest and record, then fails on existing independent issues: `04-VALIDATION.md` row `04-03-01` declares `EXERCISED` but derives `WIRED` from its evidence command, and `17-VALIDATION.md` row `17-01-01` lacks a `Grade` column. These documents are outside this plan's declared files.
- `GOCACHE=/tmp/ai-lang-go-cache go test ./...` was run and failed in `internal/compiler/session`: the same two validation-row issues above, plus five existing Phase 11 gate tests reject by-pointer bodies as unsupported by whole-program native emission. Other reported packages passed.
- The focused checked-in mutation test passed. The plan's focused grade command and full-suite gate remain red for the existing issues listed above; the validation guards were not relaxed to force a green result.
- The sandbox denied `.git/index.lock` on the first GSD commit attempt. The same GSD commits succeeded after retrying with repository metadata access; no unrelated files were staged.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

The corpus digest and complete execution record are refreshed and independently bound to current source bytes. The remaining grade and Phase 11 gate failures need their existing owners; they were not edited under this plan's three-file scope.

## Self-Check

PASSED. SUMMARY exists, and both task commits (`69df5de`, `aaa85da`) are present in the repository log.

---
*Phase: 16-branch-match-emitter-port*
*Completed: 2026-09-23*
