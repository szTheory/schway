---
phase: 29-byte-views-and-the-adler-32-library
plan: 09
subsystem: checker-and-validators
tags: [release-while-view-live, byte-views, loan-liveness, seeded-fault, d-22]
requires: [29-06, 29-07]
provides: [release-while-view-live-laws, owner-loop-block-pins, move-neighbor-pins]
affects: [29-10, 29-11, 29-12]
tech-stack:
  added: []
  patterns: [independent law per peer, per-package unexported law toggle, PLAN_BASE characterization in a scratch worktree]
key-files:
  created:
    - internal/compiler/check/check_phase29_release_test.go
    - internal/compiler/corevalidate/corevalidate_phase29_release_test.go
    - internal/compiler/pathoracle/pathoracle_phase29_release_test.go
  modified:
    - internal/compiler/check/check.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/export_test.go
    - internal/compiler/pathoracle/pathoracle.go
    - internal/compiler/pathoracle/export_test.go
    - internal/compiler/corevalidate/corevalidate_phase29_fixture_test.go
    - internal/compiler/pathoracle/pathoracle_phase29_fixture_test.go
key-decisions:
  - "A view place is a ByteView-typed target of a foreign borrow; scoping to ByteView keeps the Phase 23 file_byte owner (scalar borrow result read after release) free of false positives"
  - "check runs the law inside a new post-assembly function checkFixtureOwnerPostAssembly (law, then existing check.cfg_back_edge for cyclic blocks, then the closed five-operation shape), called from checkFixtureBytesEntry before the core is returned"
  - "Moves of the owner are not refused by move_while_borrowed on the fixture shape because the fixture view is a foreign borrow, not a loan; pinned with the codes actually observed, plus loan-form rows for the existing move codes"
requirements-completed: [BYTE-03]
status: complete
actuals:
  tokens: 45000
  tasks: 3
  commits: 3
plan_head_before: 677e04bcd847698372744c4536c5cc3a82dce6ed
plan_head_after: 1dae2b85c2856dfd2885ffcee0d4694104306468
completed: 2026-10-11
---

# Phase 29 Plan 09: Release-while-view-live laws Summary

check, corevalidate and pathoracle each independently refuse an owner release placed before the last read of a ByteView borrowed from it, with `check.release_while_view_live`, `core.release_while_view_live` and `pathoracle.release_while_view_live` (D-22 family), each proven load-bearing by a seeded fault and an unexported per-package toggle.

D-24 outcome: NOT triggered. R2 and M1 were refused at PLAN_BASE (677e04b) by every listed peer with an existing code, so no D-22 family extension was made and no re-plan stop was needed.

## Tasks

1. Tracer (358393b): three laws, toggles (`disableReleaseWhileViewLiveLawForTest` in each package; export_test setters for corevalidate and pathoracle), `TestPhase29ReleaseBeforeViewUseRefused` in all three. Each law is evaluated before that peer's closed-shape matcher (check: first step of `checkFixtureOwnerPostAssembly`; corevalidate: top of `replay`; pathoracle: before `validateFixtureOwnerFunction`). With the law toggled off the observed code becomes check.fixture_owner_shape_unsupported, core.invalid_target and pathoracle.fixture_owner_order respectively.
2. Neighbors (47e3cb2): `TestPhase29MoveOwnerWhileViewLive` (check, corevalidate), `TestPhase29LoopNeighborParseRefusals` (check), unspellable release row, accepted-fixture positive control.
3. Loop blocks (1dae2b8): `TestPhase29OwnerLoopBlockRefused` in all three packages.

## PLAN_BASE characterization (scratch worktree at 677e04b, since removed)

Entries available at PLAN_BASE: check = `loanLivenessFixpoint` over `cfgBlocksForFunction` (the only post-assembly entry for a forged core; no fixture post-assembly function existed), corevalidate = `Validate`, pathoracle = `ValidateFixtureOwnerPaths`.

| Row | Peer | planBaseCode | headCode | amendment |
|-----|------|--------------|----------|-----------|
| R1 | check | check.cfg_back_edge | check.release_while_view_live | D-22 |
| R1 | corevalidate | core.scalar_backedge_authority | core.scalar_backedge_authority | none |
| R1 | pathoracle | pathoracle.fixture_owner_shape | pathoracle.release_while_view_live | D-22 |
| R2 | check | check.cfg_back_edge | check.cfg_back_edge | none |
| R2 | corevalidate | core.scalar_backedge_authority | core.scalar_backedge_authority | none |
| R2 | pathoracle | pathoracle.fixture_owner_shape | pathoracle.fixture_owner_shape | none |
| M1 | check | check.cfg_back_edge | check.cfg_back_edge | none |
| M1 | corevalidate | core.scalar_backedge_authority | core.scalar_backedge_authority | none |

For R1 the existing PLAN_BASE codes in check and pathoracle name only the cycle or the closed shape, not the release, so the D-22 law now fires first and R1 falls under the D-22 amendment in check and pathoracle. In corevalidate the Phase 26 `core.scalar_backedge_authority` law runs before the replay, so R1 keeps its existing code and the D-22 law is not reached in a cyclic core. With each law toggle set every row equals its planBaseCode. All three PLAN_BASE codes occur in production source at 677e04b (git grep verified). Plan 29-11 cites this table in the D-22 ROADMAP note.

## Source-level rows (parse refusals)

`TestPhase29LoopNeighborParseRefusals` pins syntax.expected_linear_result at the `while` token for `let moved = owner` and for an explicit release inside a `while` body (labelled `parse refusal (D-22)`, not ownership proofs). Explicit `schway_fixture_bytes_release(owner)` in straight-line main is refused with syntax.expected_rbrace (release is not user-spellable). `let moved = owner` in straight-line main is refused by check.fixture_owner_shape_unsupported.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Law false-positived on the Phase 23 file_byte owner**
- **Found during:** Task 1
- **Issue:** the first version treated any foreign borrow target as a view; Phase 23's scalar borrow result is legitimately read after release (TestPhase23AcquisitionSeedsCoreOwnerObligation failed).
- **Fix:** a view is the ByteView-typed target of a foreign borrow, in all three laws.
- **Commit:** 358393b

**2. [Plan wording] Earlier 29-06/29-07 pins now observe the D-22 code**
- Because each law runs before the closed matchers, the corevalidate rows "release ordered before the call" (was core.invalid_target) and "view returned from main" (was core.final_claim_mismatch), and the pathoracle rows "release ordered before the call" (was fixture_owner_order) and "view returned instead of the call result" (was fixture_owner_terminal), now report the D-22 code first. Both returned-view rows are genuine views read after the release. The pins were updated in place with comments; the old codes are observed with the toggle set.
- **Files modified:** corevalidate_phase29_fixture_test.go, pathoracle_phase29_fixture_test.go
- **Commit:** 358393b

**3. [Plan assumption] Move of the owner is not refused by move_while_borrowed on the fixture shape**
- **Found during:** Task 2
- **Issue:** the plan expected ownership.move_while_borrowed / core.move_while_borrowed. The fixture view is an OpForeignCall borrow, not an OpBorrowShared loan, so the loan-based move laws do not see it. A forged owner move before the call is refused by existing codes instead: corevalidate core.place_uninitialized (release of the moved-out owner) or core.fixture_owner_lifecycle (release of the moved place, caught only by the closed lifecycle matcher); check check.fixture_owner_shape_unsupported (closed shape).
- **Fix:** no new code (D-22 prohibition honored). Tests pin the observed fixture-shaped codes and add loan-form rows (view borrowed as OpBorrowShared) that pin the existing codes: core.move_while_borrowed, ownership.move_while_borrowed (plain copy read) and check.interprocedural_loan_liveness (call read).
- **Commit:** 47e3cb2

**4. [Plan wording] check characterization entry**
- check had no forged-core post-assembly entry at PLAN_BASE; the PLAN_BASE column uses `loanLivenessFixpoint`, and the HEAD entry `checkFixtureOwnerPostAssembly` delegates cyclic block graphs to the same function after the law, so a toggled-off row equals the PLAN_BASE code.

**Total deviations:** 4 (1 bug fix, 3 plan-assumption or wording adjustments). **Impact:** none on scope; flagged safety observation below.

## Safety observation (for 29-11 and review)

Straight-line owner-move-while-view-live on the fixture shape is refused only by generic initialization and closed-shape matchers, not by a loan law, because foreign borrows carry no loan. This is refused today but is not an independent move-while-borrowed law; no new code was added because D-22 authorizes only the release family.

## Verification

- All Task verify commands pass: `TestPhase29(ReleaseBeforeViewUseRefused|MoveOwnerWhileViewLive|LoopNeighborParseRefusals|OwnerLoopBlockRefused)` across check, corevalidate, pathoracle (three PASS lines for ReleaseBeforeViewUse and OwnerLoopBlock); Phase 23/24/25 focused tests, TestValidatorImportsStayIndependent, session corpus identity (272, 273), TestPhase29FixtureChecks, TestPhase29NativeFixtureVectors, TestPhase29NeighborBaseline, TestPhase16PublicEmitterConsumerInventory, TestPhase28EntryExitClassManifest pass.
- Session meta-tests (Inventory|Groundedness|Reconciliation|ValidationLifecycle) pass across ./internal/... and ./cmd/...
- Acceptance greps pass; pathoracle imports neither check nor corevalidate; `git worktree list` shows no scratch worktree.

## Known Stubs

None.

## Threat Flags

None. T-29-26 and T-29-27 are mitigated and pinned.

## Self-Check: PASSED
