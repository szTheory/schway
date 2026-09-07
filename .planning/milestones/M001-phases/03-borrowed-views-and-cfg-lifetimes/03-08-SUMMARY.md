---
phase: 03-borrowed-views-and-cfg-lifetimes
plan: "08"
subsystem: compiler-origin-validation
tags: [go, ownership, borrowing, origin-validation, tdd, mutation-testing]

# Dependency graph
requires:
  - phase: 03-borrowed-views-and-cfg-lifetimes
    provides: "03-06's originvalidate package (independent origin/access re-derivation) and 03-VERIFICATION.md's CR-01 finding"
provides:
  - "RecomputeOrigin's OpBorrowExclusive branch guarded first-seen, symmetric with OpBorrowShared"
  - "public_view_mixed_access.lang regression fixture for a two-hop shared/exclusive reborrow chain"
  - "Recorded mutation-kill (revert-and-fail) and shipped-binary out-of-corpus proof for CR-01's fix"
affects: [03-09, OWN-04, ROADMAP-SC3, ROADMAP-SC4]

# Actuals (#2632)
actuals:
  tokens: 6500
  tasks: 2
  commits: 2

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Mutation-kill via a throwaway detached git worktree: revert the production guard there, prove the paired regression tests fail, remove the worktree, never touch the main tree's history"

key-files:
  created:
    - testdata/phase3/public_view_mixed_access.lang
  modified:
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/originvalidate/originvalidate_test.go
    - internal/compiler/testsupport/cli_test.go

key-decisions:
  - "Guarded OpBorrowExclusive with the same first-seen check already applied to OpBorrowShared — a one-line symmetric fix, not a redesign of RecomputeOrigin's backward-walk algorithm."
  - "Used a hand-written fixture with distinct variable names (`exclusive`/`shared`) rather than reusing CR-01's `y`/`z` naming, to keep the regression readable independent of the review doc."
  - "Task 2 required no source changes — it is a verification/demonstration task (mutation-kill + non-regression floor); no per-task commit was made for it since the working tree had no diff to stage."

requirements-completed: [OWN-04]

coverage:
  - id: D1
    description: "RecomputeOrigin derives access `shared` for a mixed shared/exclusive reborrow chain (closest-to-return hop decides)"
    requirement: "OWN-04"
    verification:
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_test.go#TestMixedAccessChainDerivesShared"
        status: pass
    human_judgment: false
  - id: D2
    description: "ValidatePublished rejects the fixture's declared borrow mut(buffer) against a shared-only body as core.origin_access_mismatch, and the reverse hop ordering (exclusive-of-shared) is proven to still derive correctly and pass"
    requirement: "OWN-04"
    verification:
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_test.go#TestMixedAccessChainRejectedAsAccessMismatch"
        status: pass
    human_judgment: false
  - id: D3
    description: "The shipped ./cmd/lang binary rejects the mixed-access fixture at `interface export`, exiting non-zero with core.origin_access_mismatch in its JSON diagnostics"
    requirement: "OWN-04"
    verification:
      - kind: integration
        ref: "internal/compiler/testsupport/cli_test.go#TestMixedAccessChainRejectedThroughCLI"
        status: pass
      - kind: e2e
        ref: "manual: built binary run against an out-of-corpus hand-written program (scratch.oob_mixed_access/peek/payload) — exit 2, core.origin_access_mismatch"
        status: pass
    human_judgment: false
  - id: D4
    description: "The guard is proven load-bearing by mutation-kill: reverting it in a throwaway detached worktree makes both regression tests fail"
    verification:
      - kind: other
        ref: "recorded revert-and-fail output below, ## Mutation-Kill section"
        status: pass
    human_judgment: false
  - id: D5
    description: "The Phase 3 non-regression floor (all seven required controls, both expected escapes, Phase 1/2 goldens unmoved) still holds unchanged"
    verification:
      - kind: other
        ref: "sh scripts/verify-phase3.sh (exit 0)"
        status: pass
    human_judgment: false

# Metrics
duration: 11min
completed: 2026-09-04
status: complete
---

# Phase 3 Plan 08: Mixed-Access Reborrow Chain Origin Guard Summary

**Guarded `originvalidate.RecomputeOrigin`'s exclusive-access branch first-seen so a mixed shared/exclusive reborrow chain derives `shared` (not `exclusive`), closing GAP 1 / CR-01 with a regression fixture, a mutation-kill, and a shipped-binary out-of-corpus proof.**

## Performance

- **Duration:** 11 min
- **Started:** 2026-09-04T15:27:04Z
- **Completed:** 2026-09-04T15:38:23Z
- **Tasks:** 2 completed
- **Files modified:** 3 (1 created, 3 modified — `originvalidate.go` counted once)

## Accomplishments

- `RecomputeOrigin`'s `OpBorrowExclusive` case is now guarded exactly like the paired `OpBorrowShared` case (`if derivedAccess == "" { derivedAccess = "exclusive" }`), so the backward walk's hop nearest the returned place decides the derived access mode.
- Added `testdata/phase3/public_view_mixed_access.lang` — the two-hop `borrow mut` → `borrow` chain 03-REVIEW.md CR-01 identified as untested, declared `borrow mut(buffer)` over a body that only ever hands back a shared reborrow.
- Added `TestMixedAccessChainDerivesShared` and `TestMixedAccessChainRejectedAsAccessMismatch` (the latter also exercises the reverse hop ordering — exclusive reborrow of a shared loan — proving the guard is symmetric, not one-sided).
- Added `TestMixedAccessChainRejectedThroughCLI`, the shipped-binary falsifier: `interface export` on the fixture through the real built `./cmd/lang` binary exits non-zero with `core.origin_access_mismatch` in its JSON diagnostics.
- Confirmed via mutation-kill (revert-and-fail in a throwaway detached worktree) that the guard is load-bearing, and via an out-of-corpus hand-written program that the fix generalizes beyond the fixture that ships with the gate.
- Re-ran `sh scripts/verify-phase3.sh` unchanged: all seven required controls and both expected escapes still present, exit 0 — the guard is a no-op for every existing single-hop-off-the-parameter fixture.

## Task Commits

Each task was committed atomically per the TDD RED/GREEN cycle (Task 03-08-01 carried `tdd="true"`):

1. **Task 03-08-01 (RED): failing test for mixed-access chain guard** - `e049c11` (test)
2. **Task 03-08-01 (GREEN): guard RecomputeOrigin's exclusive branch** - `410c5a7` (feat)

**Task 03-08-02** (mutation-kill + non-regression floor) produced no source diff — it is a verification/demonstration task. Its evidence is recorded below under Mutation-Kill and Shipped-Binary Out-of-Corpus Proof; no commit was made because the working tree had nothing to stage after it ran.

**Plan metadata:** committed alongside this SUMMARY (see final metadata commit).

## Files Created/Modified

- `testdata/phase3/public_view_mixed_access.lang` - the mixed shared/exclusive reborrow-chain regression fixture (new)
- `internal/compiler/originvalidate/originvalidate.go` - `RecomputeOrigin`'s `OpBorrowExclusive` case now first-seen guarded
- `internal/compiler/originvalidate/originvalidate_test.go` - `TestMixedAccessChainDerivesShared`, `TestMixedAccessChainRejectedAsAccessMismatch`
- `internal/compiler/testsupport/cli_test.go` - `TestMixedAccessChainRejectedThroughCLI`

## Decisions Made

- Guarded only the `OpBorrowExclusive` branch, matching the existing `OpBorrowShared` branch's shape exactly — no broader redesign of the backward-walk algorithm, per the plan's prohibition against relaxing `ValidatePublished`'s first-problem-only accumulation.
- Used distinct fixture variable names (`exclusive`, `shared`) instead of CR-01's `y`/`z`, and a distinct out-of-corpus program (`scratch.oob_mixed_access`/`peek`/`payload`) for the D-11 shipped-binary demonstration, so neither proof depends on names appearing anywhere else in the repo.
- Discovered during the out-of-corpus demonstration that `data` is a reserved keyword (`TokenData`) and cannot be used as a parameter name — used `payload` instead. This is a pre-existing grammar fact, not a defect, and required no code change.

## Deviations from Plan

None - plan executed exactly as written. The TDD RED/GREEN split (two commits for Task 1) follows directly from the task's own `tdd="true"` attribute and the plan's explicit instruction ("Write the failing tests first... the test must fail against the current tree before the guard lands").

## Mutation-Kill

Per D-09 and the plan's Task 03-08-02 action, the guard's load-bearing status was proven by reverting it in a throwaway detached git worktree (`git worktree add --detach <tmp> HEAD`, never touching the main tree's index or history) and re-running the two regression tests. Verbatim failing output:

```
=== RUN   TestMixedAccessChainDerivesShared
    originvalidate_test.go:110: expected access shared (closest-to-return hop), got "exclusive"
--- FAIL: TestMixedAccessChainDerivesShared (0.00s)
=== RUN   TestMixedAccessChainRejectedAsAccessMismatch
    originvalidate_test.go:133: expected exactly core.origin_access_mismatch, got []
--- FAIL: TestMixedAccessChainRejectedAsAccessMismatch (0.00s)
FAIL
FAIL	github.com/codename-lang/lang/internal/compiler/originvalidate	0.222s
FAIL
```

The worktree was removed immediately after (`git worktree remove --force <tmp>`); the main tree was never modified by this demonstration.

## Shipped-Binary Out-of-Corpus Proof (D-11)

A hand-written program NOT present in any corpus or fixture directory, using a different module name, function name, and parameter name than the shipped fixture:

```
module scratch.oob_mixed_access

export {
  fn peek
}

fn peek(payload: Buffer) -> borrow mut(payload) Buffer {
  let handle = borrow mut payload
  let glimpse = borrow handle
  glimpse
}
```

Run through the freshly built `./cmd/lang` binary:

```
$ lang --json interface export oob_mixed_access.lang out_summary.json
exit code: 2
diagnostic code: core.origin_access_mismatch
message: "s1:scratch.oob_mixed_access:fn:peek: declared access \"exclusive\", body derives \"shared\""
```

This confirms the fix generalizes beyond the shipped `public_view_mixed_access.lang` fixture, not merely to the one input the gate itself carries.

## Non-Regression Floor

`sh scripts/verify-phase3.sh` re-run after both tasks: exit 0. All seven required control IDs present (`control:ownership.exclusive_conflict`, `control:ownership.exclusive_move`, `control:core.loan_endpoint_mismatch`, `control:cfg.path_oracle_disagreement`, `control:origin.understated_summary`, `control:origin.impossible_summary`, `control:origin.stale_summary`). `escape:coordinated-frontend-summary-lie` present exactly once and never promoted to `control:escape:coordinated-frontend-summary-lie` (count 0, as required). Phase 1 and Phase 2 verify runs embedded in the same script both reported `status: pass` with their own goldens/controls unmoved (D-13).

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- ROADMAP SC3's access-mode clause and SC4's `impossible` clause are now true for the mixed-access reborrow-chain shape that 03-VERIFICATION.md's CR-01 found untested and reachable through the shipped binary.
- SC4's `omitted` clause (D-03-02: a borrow-derived return with no declared origin exports indistinguishable from a fully-owned return) remains open — that is 03-09's scope, not this plan's. `OWN-04` is intentionally NOT yet fully closable: the shared-requirement gate correctly withholds marking it complete until 03-09's sibling plan also finishes.
- No blockers for 03-09; it operates on the same `originvalidate.go`/`ValidatePublished` file this plan touched but a structurally different code path (admission-time detection of an undeclared origin, not access-mode comparison).

---
*Phase: 03-borrowed-views-and-cfg-lifetimes*
*Completed: 2026-09-04*

## Self-Check: PASSED

- FOUND: `.planning/phases/03-borrowed-views-and-cfg-lifetimes/03-08-SUMMARY.md`
- FOUND: `testdata/phase3/public_view_mixed_access.lang`
- FOUND: commit `e049c11` (test(03-08): add failing test for mixed-access reborrow chain guard)
- FOUND: commit `410c5a7` (feat(03-08): guard RecomputeOrigin's exclusive branch on first-seen)
