---
phase: 38-milestone-receipts
reviewed: 2026-10-16T23:22:12Z
depth: standard
files_reviewed: 5
files_reviewed_list:
  - .github/workflows/ci.yml
  - scripts/verify-phase32.sh
  - scripts/check-milestone-binding.sh
  - testdata/receipts/M007.manifest
  - internal/compiler/session/session_phase38_receipts_test.go
findings:
  critical: 4
  warning: 2
  info: 0
  total: 6
status: issues_found
---

# Phase 38: Code Review Report

**Reviewed:** 2026-10-16T23:22:12Z  
**Depth:** standard  
**Files Reviewed:** 5  
**Status:** issues_found

## Summary

The five scoped files were reviewed at the committed preparation source. Four blockers prevent the M007 producer/binding or freeze-preflight path from meeting its acceptance contract. The preparation should not be frozen until they are fixed and the affected focused controls are rerun. Two warnings cover a metadata type-validation gap and an inventory expansion beyond the specified 16-control/341-event scope.

**Source binding:** Base `d8cccd4a3aedae454ae53f5623314d5909bd9058`; reviewed commit `c2da3cf17ea2f7ef663d540138fbbe2d7cdbb14f`; codedigest `3a05bab4400428771cf9f5ee6eb7f3f4d2cbafea58b75fdc5f66764b6cea7d82`.

| File | SHA-256 |
|---|---|
| `.github/workflows/ci.yml` | `db35170e99b83b06a56c3c9a0695965514d43bad872c3165373f93f8d9ab8625` |
| `scripts/verify-phase32.sh` | `2c3d92c3a2df7f4bd9a6de97c85322f5858f5ec67d9d2d2371befd4f664dca86` |
| `scripts/check-milestone-binding.sh` | `336ad3103b604d803ff3e91465cef99d8a09b70905b4fe8fc173638c6827e1a6` |
| `testdata/receipts/M007.manifest` | `dd766978e6ef82c3a265e992306a013b3e58108e9570e399e5d9d9ee94a2bac7` |
| `internal/compiler/session/session_phase38_receipts_test.go` | `c5882f7cc48c5528da8cb0af0b8a07de266d3e2663ca6749653598d3b3eda353` |

**Evidence:** Source inspection covered all five files and their acceptance boundaries. Newly run checks were the focused `TestPhase38FreezeWorktreeControls/clean-source-and-planning-descendant` test and a bounded `go test -json` invocation for `TestPhase32CodeDigest`; the latter exited successfully and emitted package `start`, child-test events, and package `pass`. The freeze test passed while its fixture provided preparation paths that did not exist. The executor's earlier Task 02 focused acceptance command is a prior receipt; it uses synthetic JSON events and does not run the full M007 producer. No full producer, full/race suite, hosted workflow, or final phase receipt was run in this review.

## Narrative Findings (AI reviewer)

### Critical Issues

#### CR-01: [BLOCKER] Real `go test -json` package events always fail control capture

**File:** `scripts/verify-phase32.sh:304-318`; `scripts/check-milestone-binding.sh:239-248`  
**Severity:** BLOCKER

**Issue:** The producer accepts only `run`, `pass`, `skip`, `fail`, and `output` actions, and requires every `run`/`pass` test name to be in the manifest. Go emits a package-level `Action: "start"` without a `Test` before test events and a package-level `Action: "pass"` without a `Test` after a successful package run. The producer rejects the first event as an unexpected action; the binding parser would also reject the final package pass as an unexpected empty test name. Thus actual controls cannot produce or bind a receipt. The synthetic event fixtures omit both package events and do not expose this failure.

**Reproduction:** `GOCACHE=/tmp/schway-phase38-review-gocache go test -json -count=1 -run '^TestPhase32CodeDigest$' ./internal/compiler/session` exited 0 and emitted package-level `start` and `pass` records with no `Test` field.

**Fix:** In both parsers, handle package lifecycle events separately from test events: accept the expected package-level `start` and successful package-level `pass` records, continue to reject package-level failure, and keep exact run/pass/skip validation for every required test. Add a parser control using real captured Go JSON output.

#### CR-02: [BLOCKER] The codedigest control omits emitted subtest events

**File:** `testdata/receipts/M007.manifest:232,249`; `scripts/verify-phase32.sh:304-318`; `scripts/check-milestone-binding.sh:239-248`  
**Severity:** BLOCKER

**Issue:** The `codedigest` selector runs `^TestPhase32CodeDigest$`, which runs that test's subtests as well. The manifest inventories only `TestPhase32CodeDigest`, while the successful Go JSON output also contains child names such as `TestPhase32CodeDigest/sh-equals-go` and `TestPhase32CodeDigest/subdirectory-equals-root`. Once package lifecycle records are handled, the producer and binding checker still reject the first unlisted child as unexpected. Other controls with a root-only event declaration need the same selector-versus-emitted-name check.

**Fix:** Inventory every required child emitted by each root selector, or narrow selectors to an explicitly intended set and inventory all resulting events. Keep the exact child-name set tied to a real JSON capture so omitted descendants cannot silently disappear.

#### CR-03: [BLOCKER] M007 completion compares 106 executed rows with the 39-row M006 table

**File:** `scripts/verify-phase32.sh:112-118,351-364`  
**Severity:** BLOCKER

**Issue:** In M007 mode, `case_table` is deliberately replaced with the inherited M006 rows at line 115, so `table_count` remains 39. The producer executes those 39 rows, then adds 67 M007 rows printed by the Python inventory at lines 351-353, making `case_count` 106. The final equality at line 364 therefore always fails with `incomplete_cases`, even after control-event parsing succeeds.

**Fix:** Compute the completion total from the full M007 manifest inventory, or compare inherited and supplemental counts independently and then require their sum to equal the manifest's 106 distinct rows. Add a focused accounting control that exercises the actual M007 count path.

#### CR-04: [BLOCKER] Freeze preflight accepts nonexistent or altered preparation records

**File:** `scripts/check-milestone-binding.sh:94-103`; `internal/compiler/session/session_phase38_receipts_test.go:593-598`  
**Severity:** BLOCKER

**Issue:** `freeze_data` checks that each preparation entry has the expected keys, a hash-shaped `sha256`, and `source_sha`/`codedigest` strings matching values in the freeze JSON. It never resolves the path, reads the record, or compares its bytes with the supplied hash. The positive freeze fixture supplies nonexistent review/security/validation/audit paths and arbitrary hashes, yet `TestPhase38FreezeWorktreeControls/clean-source-and-planning-descendant` expects the freeze gate to pass. A freeze can therefore pass without the review, security result, validation record, or audit it claims to bind.

**Fix:** Resolve every preparation path with `safe_path`, require the file to exist within the repository, hash its bounded bytes and compare against `sha256`, then validate the record's declared source SHA/codedigest and required disposition. Change the positive fixture to create real source-bound records and add missing-file, changed-byte, and stale-source refusals.

### Warnings

#### WR-01: [WARNING] Artifact and log identity fields accept non-integer JSON values

**File:** `scripts/check-milestone-binding.sh:289-291,304-308`  
**Severity:** WARNING

**Issue:** Top-level run IDs, artifact IDs, and log byte counts use `positive_int`, but artifact `run_id`/`attempt` and log `artifact_id` are compared to integers without type checks. Python equality accepts values such as `9001.0 == 9001`, `1.0 == 1`, and `True == 1`, so malformed JSON identity fields pass the same-attempt and artifact-membership checks despite the strict integer contract.

**Fix:** Require `positive_int(a["run_id"])`, `positive_int(a["attempt"])`, and `positive_int(item["artifact_id"])` before comparing. Also require the three dispatch input values to have exact boolean types so numeric zero cannot stand in for `false`.

#### WR-02: [WARNING] Manifest exceeds the scoped control inventory by one root and 133 event names

**File:** `testdata/receipts/M007.manifest:3,239,257-389`; `internal/compiler/session/session_phase38_receipts_test.go:177-198`  
**Severity:** WARNING

**Issue:** The requested inventory is 16 control roots and 341 event names. The manifest declares 17 roots and 474 unique event names. The entire difference is the additional `bitwise-optimization` root and its 133 event names; the manifest-binding test explicitly requires the full `TestPhase35BitwiseOptimizationMatrix` vector matrix even though the scoped bitwise witnesses name the locked-264 and alternate-8 results. This expands the hosted producer inventory beyond the stated finite scope on both hosts.

**Fix:** Remove or narrow the extra matrix inventory and its matching assertions to the approved bitwise witnesses, then assert the exact scoped root/event totals. If the full matrix is a newly accepted requirement, update the phase's scoped inventory contract before freeze so the producer and audit agree.

---

_Reviewed: 2026-10-16T23:22:12Z_  
_Reviewer: the agent (gsd-code-reviewer)_  
_Depth: standard_
