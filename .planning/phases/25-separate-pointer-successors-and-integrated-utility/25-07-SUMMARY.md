---
phase: 25-separate-pointer-successors-and-integrated-utility
plan: "07"
subsystem: evidence
tags: [hosted-ci, native-evidence, security, roadmap]

requires:
  - phase: 25-06
    provides: "The integrated utility, fail-incomplete evidence matrix, and local evidence records."
provides:
  - "Phase 25 validation and security records bound to hosted run 36971855722 and its 18 native rows."
  - "Current product roadmap and language maturity claims reconciled with dual-host evidence and bounded refusals."
affects: [25-verification, M004-closeout, product-roadmap, language-maturity]

actuals:
  tokens: 7777
  tasks: 2
  commits: 2
commits: 2
plan_head_before: a7500a5965e6e6563762701e1bf2e3589b7b0bc5

tech-stack:
  added: []
  patterns:
    - "Bind evidence claims to hosted run, branch head, merge revision, native host, family, and lane."
    - "Keep historical and local evidence distinct from current hosted receipts."

key-files:
  created:
    - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-07-SUMMARY.md
  modified:
    - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md
    - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-SECURITY.md
    - .planning/PRODUCT-ROADMAP.md
    - .planning/LANGUAGE-MATURITY.md

key-decisions:
  - "Run 36971855722 closes EVD-10 using the paired native Linux/x86_64 and macOS/arm64 aggregate jobs at merge revision a90c27c5b432ef6fc59fbafaa68b50a1374ae138."
  - "Preserve SECURED / ASVS L1, 15/15 mitigations closed, and zero open threats; hosted evidence supplements rather than expands the security audit."
  - "Keep the exact utility-chain pointer admission and existing mutation, forwarding, retention, callback, nonlocal-exit, wider-shape, and unsupported-attribute refusals."

requirements-completed: [NAT-11, NAT-12, NAT-13, EVD-10, DX-14, DX-15]

duration: "7min"
completed: 2026-10-02
status: complete
---

# Phase 25 Plan 07: Hosted Evidence and Living Claims Summary

**Phase 25's validation, security, roadmap, and maturity records now agree on the successful 18-row dual-host receipt and the bounded pointer refusal frontier.**

## Performance

- **Duration:** 7 min
- **Started:** 2026-10-02T07:15:46Z
- **Completed:** 2026-10-02T07:21:50Z
- **Tasks:** 2
- **Files modified:** 4 planned artifacts, plus this summary

## Accomplishments

- Bound the validation and security records to GitHub Actions run `36971855722`, branch head `421b5b94eb867e940a207dfd64d7971dc8198172`, and merge revision `a90c27c5b432ef6fc59fbafaa68b50a1374ae138`.
- Recorded all 18 native host/family/lane rows: foreign, shared, and exclusive families each passed baseline `-O0`, optimized `-O2`, and ASan+UBSan on Linux/x86_64 and macOS/arm64, with expected/actual 65/66 and typed `0x43` failure before helper calls.
- Reconciled both living documents with the current result while preserving dated historical claims, the existing security verdict, exact pointer refusal boundaries, and all three fully specified ranked capabilities.

## Evidence

Both task-specific verification commands passed:

- `GOCACHE=/private/tmp/schway-gocache go test -count=1 -run '^TestPhase25(EvidenceScript|EvidenceIndex)$' ./internal/compiler/native`
- `GOCACHE=/private/tmp/schway-gocache go test -count=1 -run '^TestPhase25LivingRoadmap$' ./internal/compiler/native`
- `rg -q '36971855722'` against each task's target documents.
- `git diff --check`

Hosted `checks` and `current evidence aggregate` jobs passed on both hosts. Linux used Go `linux/amd64`, target `x86_64-pc-linux-gnu`, Ubuntu Clang 18.1.3; macOS used Go `darwin/arm64`, target `arm64-apple-darwin25.6.0`, Apple Clang 21. Each host passed vet, build, full Go tests, race tests, and the current evidence aggregate. Cold min/median/max was 9.350/9.540/9.980s (Linux) and 12.470/12.670/12.930s (macOS); warm was 0.390/0.490/0.550s and 0.820/1.400/2.130s respectively. The hosted receipt is distinct from source inspection, local checks, and historical runs.

## Task Commits

1. **Task T-25-11: Bind the hosted matrix to Phase 25 validation and security** — `2daa40c` (docs)
2. **Task T-25-12: Reconcile current product and maturity claims** — `ba1477e` (docs)

## Files Created/Modified

- `.planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md` — current hosted matrix and EVD-10 sign-off.
- `.planning/phases/25-separate-pointer-successors-and-integrated-utility/25-SECURITY.md` — hosted T-25-14 receipt with unchanged SECURED audit.
- `.planning/PRODUCT-ROADMAP.md` — current Phase 25 status and dated provenance.
- `.planning/LANGUAGE-MATURITY.md` — current capability evidence and dated hosted amendment.

## Decisions Made

- The two successful native aggregate jobs jointly supply all host rows; an individual script's absent-other-host rows are scoped to that one-host invocation.
- Existing pointer scope and the three ranked next capabilities remain unchanged.

## Deviations from Plan

None — plan executed as written.

## Issues Encountered

The first git commit attempt was denied by the sandbox's `.git/index.lock` restriction. The same verified staged changes were committed through the authorized repository write path; no content was lost or altered.

## Next Phase Readiness

All four planned documents are ready for the fresh Phase 25 goal-backward verification. Security remains SECURED / ASVS L1 with all 15 mitigations closed and zero open threats.

## Self-Check: PASSED

- All four modified artifacts exist and include current run `36971855722` evidence where applicable.
- Task commits `2daa40c` and `ba1477e` exist.
- Hosted provenance, matrix rows, security verdict, refusal frontier, and all three ranked recommendations are recorded.

---
*Phase: 25-separate-pointer-successors-and-integrated-utility*
*Completed: 2026-10-02*
