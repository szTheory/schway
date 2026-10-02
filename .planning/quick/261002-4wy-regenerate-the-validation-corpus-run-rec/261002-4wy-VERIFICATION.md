---
phase: quick
plan: 261002-4wy
verified: 2026-10-02T11:17:00Z
status: passed
score: 4/4 must-haves verified
covered_files:
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md
  - .planning/quick/261002-4wy-regenerate-the-validation-corpus-run-rec/261002-4wy-PLAN.md
  - .planning/quick/261002-4wy-regenerate-the-validation-corpus-run-rec/261002-4wy-SUMMARY.md
  - testdata/phase16/validation-corpus-run-record.jsonl
  - testdata/phase16/validation-corpus-run-record.manifest.json
covered_digest: "v1:sha256:60f8632fbbcc84c131888756e6742d49e8c32a0c10162ea5faff58d24c248e30"
behavior_unverified: 0
overrides_applied: 0
---

# Quick 261002-4wy Verification Report

**Goal:** Complete the Phase 25 graded validation map and refresh its derived corpus execution receipt.
**Verified:** 2026-10-02T11:17:00Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

| # | Must-have truth | Status | Evidence |
|---|---|---|---|
| 1 | All ten Phase 25 rows use the graded schema, have honest Grade and Non-inertness cells, and cite executable evidence; T-25-04/06 preserve their named controls. | VERIFIED | [25-VALIDATION.md](../../phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md) has ten twelve-column rows, `EXERCISED` and em dash in each row. The T-25-04 and T-25-06 commands match the plan exactly. The record contains passing, non-skipped pass events for all eight named controls. The archived grader test passed. |
| 2 | The JSONL is a completed producer batch over exactly the current consumer-derived pairs. | VERIFIED | A fresh `TestExportValidationCorpusPairs` export produced 34 ordered pairs in 37,820 bytes with SHA-256 `14f14924a5cb4f30c3f84db5a3dbf9b47c3265bd6bd5dbafecd50ccc926a3b93`, matching the manifest. The consumer pair-export and archived-grade tests passed. The current JSONL contains 34 pair completion witnesses in matching order, one final 34-pair batch witness, and zero `fail` actions. Its digest matches the manifest. The summary records a successful producer invocation with inherited writable cache, empty stderr, and 58.215968s elapsed. |
| 3 | The manifest binds the receipt to exact exported bytes, current revision, completion witnesses, and measured elapsed time. | VERIFIED | Manifest schema is v1, `completed` is true, revision matches `git rev-parse HEAD`, both SHA-256 values match the current files, `produced_at` parses as UTC RFC3339, and `elapsed` is positive (`58.215968s`). Record witness counts and final ordering match the consumer contract. |
| 4 | Isolated archived grading, groundedness, and the full Go suite pass, with no test or grade-policy weakening. | VERIFIED | Focused pair-export, archived-grade, groundedness-frontier, and three-class tests passed. `GOCACHE=/private/tmp/schway-gocache go test -count=1 ./...` passed after the summary was written. `git diff --exit-code HEAD -- cmd internal scripts go.mod go.sum` and `git diff --check` passed. The diff is limited to the Phase 25 map, the two evidence artifacts, and quick-workflow metadata. |

**Score:** 4/4 must-haves verified (0 present, behavior-unverified)

## Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `.planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md` | Complete graded Phase 25 map | VERIFIED | Ten rows; schema and grades accepted by archived-corpus grader. |
| `testdata/phase16/validation-corpus-run-record.jsonl` | Raw test JSONL and producer witnesses | VERIFIED | 3,124,525 bytes; SHA-256 matches manifest; 34 pair witnesses, one final batch witness, zero fail actions. |
| `testdata/phase16/validation-corpus-run-record.manifest.json` | Schema-v1 binding and provenance | VERIFIED | Exact pair/record digests; current HEAD revision; completed true; positive elapsed and RFC3339 timestamp. |

## Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| Phase 25 validation map | `requestedCorpusPairs` | Consumer-derived row evidence | WIRED | Fresh exported bytes match the manifest pair digest; consumer alignment test passed. |
| `TestExportValidationCorpusPairs` | `scripts/evidence-run-record.sh` | Pair export and discrete argv pairs | WIRED | Producer receives exported package/pattern pairs; current record witnesses align with the ordered export. |
| Manifest | `TestValidationRowGradesAreEarnedOverArchivedCorpus` | Digest, revision, duration, witnesses | WIRED | Focused grader passed against the current checked-in artifacts. |

## Data-Flow Trace

| Artifact | Data | Source | Produces real data | Status |
|---|---|---|---|---|
| JSONL run record | Go test events | Recorded Go test invocations for exported package/pattern pairs | Yes; contains test run/output/pass events and producer completion witnesses | FLOWING |
| Manifest | Digests and batch metadata | Exact exported pair bytes, exact JSONL bytes, Git revision, production timestamp, measured duration | Yes; both hashes and revision were recomputed from current files/state | FLOWING |

## Behavioral and Scope Checks

| Check | Result |
|---|---|
| Focused pair-export, archived-grade, groundedness frontier, and three-class tests | PASS |
| T-25-04 and T-25-06 named controls | PASS; all eight have pass events and no skips/failures |
| Full uncached Go suite after summary | PASS; `go test -count=1 ./...` |
| No-test-weakening source/dependency check | PASS; no diff under `cmd`, `internal`, `scripts`, `go.mod`, or `go.sum` |
| Whitespace check | PASS |

A duplicate producer replay was started during verification and interrupted at the orchestrator's request before completion. It did not change tracked files and is not counted as evidence. Producer execution evidence is the accepted run documented in the summary, corroborated by the current receipt's exact pair digest, test events, witness structure, consumer tests, and full suite.

## Requirements Coverage

No formal requirement IDs are declared in the quick plan. The source audit's stated requirements are covered by the must-haves above.

## Anti-Patterns Found

None affecting the declared implementation artifacts. The implementation diff contains no source, test, script, dependency, grade-policy, or exemption changes.

## Human Verification Required

None.

## Gaps Summary

No gaps found. The current graded map, exported pair list, bound receipt, focused gates, and full Go suite satisfy the quick task's acceptance criteria.

---

_Verified: 2026-10-02T11:17:00Z_  
_Verifier: gsd-verifier_
