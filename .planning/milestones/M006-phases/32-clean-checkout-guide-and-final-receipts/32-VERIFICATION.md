---
phase: 32-clean-checkout-guide-and-final-receipts
verified: 2026-10-12T23:30:00Z
status: passed
score: 3/3 must-haves verified
covered_files:
  - ".planning/phases/32-clean-checkout-guide-and-final-receipts/32-01-PLAN.md"
  - ".planning/phases/32-clean-checkout-guide-and-final-receipts/32-01-SUMMARY.md"
  - ".planning/phases/32-clean-checkout-guide-and-final-receipts/32-02-PLAN.md"
  - ".planning/phases/32-clean-checkout-guide-and-final-receipts/32-02-SUMMARY.md"
  - ".planning/phases/32-clean-checkout-guide-and-final-receipts/32-03-PLAN.md"
  - ".planning/phases/32-clean-checkout-guide-and-final-receipts/32-03-SUMMARY.md"
  - ".planning/phases/32-clean-checkout-guide-and-final-receipts/32-04-PLAN.md"
  - ".planning/phases/32-clean-checkout-guide-and-final-receipts/32-04-SUMMARY.md"
  - ".planning/phases/32-clean-checkout-guide-and-final-receipts/32-05-PLAN.md"
  - ".planning/phases/32-clean-checkout-guide-and-final-receipts/32-05-SUMMARY.md"
  - ".planning/phases/32-clean-checkout-guide-and-final-receipts/32-06-PLAN.md"
  - ".planning/phases/32-clean-checkout-guide-and-final-receipts/32-06-SUMMARY.md"
  - ".planning/phases/32-clean-checkout-guide-and-final-receipts/32-07-PLAN.md"
  - ".planning/phases/32-clean-checkout-guide-and-final-receipts/32-07-SUMMARY.md"
  - ".planning/phases/32-clean-checkout-guide-and-final-receipts/32-08-PLAN.md"
  - ".planning/phases/32-clean-checkout-guide-and-final-receipts/32-08-SUMMARY.md"
  - ".planning/phases/32-clean-checkout-guide-and-final-receipts/32-09-PLAN.md"
  - ".planning/phases/32-clean-checkout-guide-and-final-receipts/32-09-SUMMARY.md"
  - ".planning/phases/32-clean-checkout-guide-and-final-receipts/32-10-PLAN.md"
  - ".planning/phases/32-clean-checkout-guide-and-final-receipts/32-10-SUMMARY.md"
  - ".planning/phases/32-clean-checkout-guide-and-final-receipts/32-11-PLAN.md"
  - ".planning/phases/32-clean-checkout-guide-and-final-receipts/32-11-SUMMARY.md"
  - "cmd/schway/main_phase32_guide_test.go"
  - "examples/checksum/README.md"
  - "scripts/check-phase32-freeze.sh"
  - "scripts/check-phase32-receipts.sh"
  - "scripts/verify-phase32.sh"
covered_digest: "v3:sha256:3c934c099400a9a3e2b0fbb6755c4a4752630ba028c9320d3ac0057e0ee2ee8d"
behavior_unverified: 0
---

# Phase 32: Clean-checkout guide and final receipts Verification Report

**Phase Goal:** A developer can reproduce every M006 program from a clean checkout using documented commands, and the milestone's native claims are bound to hosted macOS and Linux receipts at the final revision.
**Status:** passed
**Re-verification:** No, initial verification

## Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| SC1 | Guide doc-test from a clean checkout | VERIFIED | `examples/checksum/README.md` exists with the 65,536-byte limit. `go test ./cmd/schway -run Phase32 -v`: GuideCleanCheckout (2.23s, real clang run, 0 skips), GuideMatchesExpectedJSON, GuideParserNegativeControls, NativePolicy and PublicRouteAudit all PASS. |
| SC2 | `examples/checksum.schway` retired, with a dated amendment | VERIFIED | The file is absent. PROJECT.md carries "Amended 2026-10-12 (Phase 32, DX-19)" with the tombstone sha256 and the supported route. Remaining mentions of the path are in frozen goldens and tests under dated retired-path allow-lists (absence allowed, re-appearance fails). |
| SC3 | Hosted macOS and Linux receipts bound to the final revision (codedigest) | VERIFIED | `sh scripts/codedigest.sh` prints c1b4e85d...cd227, matching the recorded one. The freeze gate passes (079e112f -> HEAD d308dd9f). `gh run view 38315799626` reports conclusion success, with the ubuntu and macos aggregate jobs and the checks jobs all success. `check-phase32-receipts.sh` on the two committed logs exits 0: status=pass, hosts macOS and Linux, 7 programs, 39 cases. `32-RECEIPT.json` has runId 38315799626 and the same codedigest. The run headSha 6c304271 is an ancestor of HEAD, and `git diff 6c304271..HEAD` outside `.planning` is empty, so no code changed after the run. |

## Requirements Coverage

| Requirement | Description | Status | Evidence |
|-------------|-------------|--------|----------|
| DX-19 | Guide commands, expected results, limit and error outcomes; frontier retired | SATISFIED | SC1 and SC2 above. |
| EVD-15 | Hosted macOS and Linux receipts at the final M006 revision | SATISFIED | SC3. Both IDs are marked complete in REQUIREMENTS.md. No orphaned Phase 32 requirements. |

## Session meta-tests

`go test ./internal/compiler/session -run 'Phase32|TestPhase28EntryExitClassManifest|TestPhase30PreChange|TestValidationRowGradesAreEarnedOverArchivedCorpus|TestSummaryCoverageGroundedness|TestLanguageMaturityCountsAreCurrent'` passes.

## Orchestrator deviations assessed

- **8c565e8b (corpusRetiredTests allow-list):** it is an ancestor of the freeze commit, so it is inside the digested revision. It touches only `evidence_grade_test.go`. It is the same retired-path pattern as `phase28RetiredEntries` and `phase30RetiredPaths`, and the meta-tests pass with it. Accepted.
- **32-REVIEW.md:** 2 Warning and 5 Info findings, all deferred in 32-REVIEW-DISPOSITION.md (open: 0). The disposition states each fails closed or is a hardening gap, and none affects the recorded receipt's soundness. The todo `.planning/todos/pending/2026-10-12-phase32-review-hardening-after-m006.md` exists. Fixing them would touch digest paths and reopen the receipts (D-10a). WARNING-level note only: WR-01 and WR-02 are real latent gate weaknesses. WR-02 is `codedigest.sh` being cwd-dependent with no empty-listing check. I recomputed the digest from the repo root and it matches, so it is not blocking.
- **Sequential execution instead of worktrees:** process only, no goal impact.

## Caveat

`check-phase32-receipts.sh` states that it checks consistency and digest binding, not authenticity. Authenticity rests on the run record, which I confirmed read-only through `gh`.

## Human Verification Required

None. Every item has an automated proxy.

## Gaps Summary

No gaps. Anti-pattern scanning was not re-run beyond the targeted tests.

_Verifier: Claude (gsd-verifier)_
