---
phase: 03-borrowed-views-and-cfg-lifetimes
plan: "09"
subsystem: compiler-origin-validation
tags: [go, ownership, borrowing, origin-validation, tdd, mutation-testing, gap-closure]

# Dependency graph
requires:
  - phase: 03-borrowed-views-and-cfg-lifetimes
    provides: "03-06's originvalidate package (independent origin/access re-derivation), 03-08's mixed-access guard and public_view_mixed_access.lang, and 03-VERIFICATION.md's D-03-02/GAP 2 finding"
provides:
  - "core.origin_omitted: ValidatePublished recomputes an origin for every function, including those with no declared origin, and refuses publication when the body derives a real borrow-derived return"
  - "public_view_omitted.lang: the borrow-derived-return-with-no-annotation control fixture 03-VERIFICATION.md found missing"
  - "control:origin.omitted_summary and control:origin.mixed_access_chain wired into the Phase 3 gate's fail-closed nine-control required set"
  - "D-03-02 closed in 03-DEBT.md; D-03-01 remains open, accepted, non-blocking"
affects: [OWN-04, ROADMAP-SC3, ROADMAP-SC4, phase-04-planning]

# Actuals (#2632)
actuals:
  tokens: 4500
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Mutation-kill via a throwaway detached git worktree: revert the production branch there, prove the paired regression tests fail, remove the worktree, never touch the main tree's history (same pattern 03-08 established)"

key-files:
  created:
    - testdata/phase3/public_view_omitted.lang
  modified:
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/originvalidate/originvalidate_test.go
    - internal/compiler/testsupport/cli_test.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_test.go
    - scripts/verify-phase3.sh
    - .planning/phases/03-borrowed-views-and-cfg-lifetimes/03-DEBT.md

key-decisions:
  - "ValidatePublished now calls RecomputeOrigin unconditionally for every function, branching on whether PublicOrigin is nil afterward, rather than gating the call itself — preserves the existing understated/access_mismatch branches byte-for-byte while adding the omitted branch as a sibling."
  - "The gate lives entirely in originvalidate.ValidatePublished (the publication path); check.go's admission path is untouched, so lang check still accepts exclusive_borrow_clean / relay exactly as before while lang interface export on the same source now refuses to publish it."
  - "public_view_omitted.lang reuses check_exclusive_test.go's exact relay shape (borrow mut, reborrow shared, return the exclusive loan) so the fixture pins the same defect the fixture_disposition names, rather than inventing an unrelated shape."
  - "The new mixed_access_chain gate control needs no mutation injection (unlike understated/impossible) — check.go's honest producer already constructs the mismatching declaration itself, which is exactly why 03-VERIFICATION.md found it reachable through the shipped binary."

requirements-completed: [OWN-04]

coverage:
  - id: D1
    description: "originvalidate.ValidatePublished refuses to publish a borrow-derived return with no declared origin, coded core.origin_omitted"
    requirement: "OWN-04"
    verification:
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_test.go#TestOmittedOriginRejected"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_test.go#TestExclusiveBorrowCleanShapeChecksButCannotPublish"
        status: pass
    human_judgment: false
  - id: D2
    description: "The new check does not over-fire on an owned return with no declared origin, straight-line or match-bodied"
    requirement: "OWN-04"
    verification:
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_test.go#TestOwnedReturnWithNoOriginStillPublishes"
        status: pass
    human_judgment: false
  - id: D3
    description: "The shipped ./cmd/lang binary refuses to publish the omitted-origin fixture (exit 2, core.origin_omitted, no summary written) while lang check on the same source exits 0"
    requirement: "OWN-04"
    verification:
      - kind: integration
        ref: "internal/compiler/testsupport/cli_test.go#TestOmittedOriginRejectedThroughCLI"
        status: pass
      - kind: e2e
        ref: "manual: built binary run against an out-of-corpus hand-written program (scratch.oob_omitted/glimpse/payload) — check exit 0, export exit 2, core.origin_omitted"
        status: pass
    human_judgment: false
  - id: D4
    description: "Both new origin defects are required, fail-closed, nonzero-work controls in the Phase 3 gate (control:origin.omitted_summary, control:origin.mixed_access_chain)"
    requirement: "OWN-04"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestVerifyPhase3ControlsAndWork"
        status: pass
      - kind: other
        ref: "sh scripts/verify-phase3.sh (exit 0, nine required controls present)"
        status: pass
    human_judgment: false
  - id: D5
    description: "The fix is proven load-bearing by mutation-kill: reverting the unconditional-recomputation branch in a throwaway detached worktree makes both regression tests fail"
    verification:
      - kind: other
        ref: "recorded revert-and-fail output below, ## Mutation-Kill section"
        status: pass
    human_judgment: false
  - id: D6
    description: "D-03-02 is closed in 03-DEBT.md with a dated entry naming the resolving problem code and both control IDs; D-03-01 remains open, accepted, non-blocking"
    verification:
      - kind: other
        ref: ".planning/phases/03-borrowed-views-and-cfg-lifetimes/03-DEBT.md, ## Closure section"
        status: pass
    human_judgment: false

# Metrics
duration: 22min
completed: 2026-09-04
status: complete
---

# Phase 3 Plan 09: Omitted-Origin Publication Gate Summary

**`originvalidate.ValidatePublished` now recomputes an origin for every function including those with none declared, refusing to publish a borrow-derived return with no annotation as `core.origin_omitted` — closing GAP 2/D-03-02 with a nine-control fail-closed gate, while `lang check` keeps accepting `exclusive_borrow_clean` unchanged.**

## Performance

- **Duration:** 22 min
- **Started:** 2026-09-04T15:41:00Z
- **Completed:** 2026-09-04T16:03:00Z
- **Tasks:** 3 completed
- **Files modified:** 8 (1 created, 7 modified)

## Accomplishments

- `ValidatePublished`'s unconditional `PublicOrigin == nil { continue }` short-circuit is replaced: `RecomputeOrigin` runs for every function; when it succeeds for a function that declared no origin, publication is refused with a new `core.origin_omitted` problem naming the derived paths and access mode.
- Added `testdata/phase3/public_view_omitted.lang` — the same borrow-derived-return shape as the already-shipped `exclusive_borrow_clean` / `relay` fixture, exported here as the control 03-VERIFICATION.md found missing.
- Added `TestOmittedOriginRejected`, `TestOwnedReturnWithNoOriginStillPublishes` (straight-line AND match-bodied owned returns both still publish clean), and `TestExclusiveBorrowCleanShapeChecksButCannotPublish` (embeds the exact `relay` source, machine-checking the fixture_disposition).
- Added `TestOmittedOriginRejectedThroughCLI`: the shipped `./cmd/lang` binary exits 2 with `core.origin_omitted` on the fixture and writes no summary file.
- Wired `control:origin.omitted_summary` and `control:origin.mixed_access_chain` into `verifyBorrowedCorpus`'s origin-controls lane, `TestVerifyPhase3ControlsAndWork`, and `scripts/verify-phase3.sh`'s required-control loop — the gate now requires nine control IDs fail-closed.
- Confirmed via mutation-kill (revert-and-fail in a throwaway detached worktree) that the new branch is load-bearing, and via a hand-written out-of-corpus program that the fix generalizes beyond the shipped fixture.
- Closed D-03-02 in `03-DEBT.md` with a dated entry naming the resolving code and both control IDs; D-03-01 is left unedited, still open, accepted, non-blocking.

## Task Commits

Each task was committed atomically per the TDD RED/GREEN cycle (Task 03-09-01 carried `type="tracer" tdd="true"`):

1. **Task 03-09-01 (RED): failing test for omitted-origin rejection** - `d1977a3` (test)
2. **Task 03-09-01 (GREEN): recompute origin unconditionally, refuse omitted publication** - `5f556da` (feat)
3. **Task 03-09-02: wire both new origin defects as fail-closed required controls** - `2eb7b8f` (feat)
4. **Task 03-09-03: dated debt closure** - `275ad38` (docs) — the mutation-kill and shipped-binary demonstration were verification/demonstration steps producing no source diff of their own; their evidence is recorded below.

**Plan metadata:** committed alongside this SUMMARY (see final metadata commit).

## Files Created/Modified

- `testdata/phase3/public_view_omitted.lang` - the omitted-origin control fixture (new)
- `internal/compiler/originvalidate/originvalidate.go` - `ValidatePublished` recomputes unconditionally, adds `core.origin_omitted`
- `internal/compiler/originvalidate/originvalidate_test.go` - `TestOmittedOriginRejected`, `TestOwnedReturnWithNoOriginStillPublishes`, `TestExclusiveBorrowCleanShapeChecksButCannotPublish`
- `internal/compiler/testsupport/cli_test.go` - `TestOmittedOriginRejectedThroughCLI`
- `internal/compiler/session/session.go` - two new lane checks in `verifyBorrowedCorpus`, two new required control IDs
- `internal/compiler/session/session_test.go` - `TestVerifyPhase3ControlsAndWork` asserts nine required IDs
- `scripts/verify-phase3.sh` - nine-control required list
- `.planning/phases/03-borrowed-views-and-cfg-lifetimes/03-DEBT.md` - dated D-03-02 closure entry

## Decisions Made

- Kept the declared-origin comparison branch (understated/access_mismatch), the first-problem-only accumulation, and problem ordering exactly as they were — the only change is that `RecomputeOrigin` is now called unconditionally and its result is inspected against a nil `PublicOrigin` before falling through to the existing comparison.
- Did not restore any `check.go`-side rejection; the gate lives entirely on the publication path (`ValidatePublished` / `lang interface export`), per the plan's explicit prohibition and the fixture_disposition's decision.
- No new grammar, keyword, or opt-out spelling was introduced — the only way to publish a borrow-derived return remains declaring `borrow(path)` / `borrow mut(path)`, which the language already spells.
- Used a distinct out-of-corpus program (`scratch.oob_omitted`/`glimpse`/`payload`) for the D-11 shipped-binary demonstration, matching 03-08's precedent of using names that appear nowhere else in the repo.

## Deviations from Plan

None - plan executed exactly as written. The TDD RED/GREEN split (two commits for Task 1) follows directly from the task's own `tdd="true"` attribute and the plan's explicit instruction to write the failing tests first.

## Mutation-Kill

Per D-09 and the plan's Task 03-09-03 action, the `core.origin_omitted` branch's load-bearing status was proven by reverting it in a throwaway detached git worktree (`git worktree add --detach <tmp> HEAD`, never touching the main tree's index or history) and re-running the two regression tests. Verbatim failing output:

```
=== RUN   TestOmittedOriginRejected
    originvalidate_test.go:177: expected exactly core.origin_omitted, got []
--- FAIL: TestOmittedOriginRejected (0.00s)
FAIL
FAIL	github.com/codename-lang/lang/internal/compiler/originvalidate	0.259s
FAIL
```

```
=== RUN   TestOmittedOriginRejectedThroughCLI
    cli_test.go:488: expected non-zero exit for an omitted origin declaration: {Stdout:... Exit:0}
--- FAIL: TestOmittedOriginRejectedThroughCLI (0.34s)
FAIL
FAIL	github.com/codename-lang/lang/internal/compiler/testsupport	0.566s
FAIL
```

The worktree was removed immediately after (`git worktree remove --force <tmp>`); the main tree was never modified by this demonstration.

## Shipped-Binary Out-of-Corpus Proof (D-11)

A hand-written program NOT present in any corpus or fixture directory, using a different module name, function name, and parameter name than the shipped fixture:

```
module scratch.oob_omitted

export {
  fn glimpse
}

fn glimpse(payload: Buffer) -> Buffer {
  let handle = borrow mut payload
  let peek = borrow handle
  handle
}
```

Run through the freshly built `./cmd/lang` binary:

```
$ lang --json check oob_omitted.lang
exit code: 0

$ lang --json interface export oob_omitted.lang out_summary.json
exit code: 2
diagnostic code: core.origin_omitted
message: "s1:scratch.oob_omitted:fn:glimpse: no declared origin, but body derives origin [payload] with access \"exclusive\""
(out_summary.json was not created)
```

This confirms `check` and `interface export` genuinely disagree by design on the identical source, and that the fix generalizes beyond the shipped `public_view_omitted.lang` fixture.

## Non-Regression Floor

`sh scripts/verify-phase3.sh` re-run after all three tasks: exit 0. All nine required control IDs present (`control:ownership.exclusive_conflict`, `control:ownership.exclusive_move`, `control:core.loan_endpoint_mismatch`, `control:cfg.path_oracle_disagreement`, `control:origin.understated_summary`, `control:origin.impossible_summary`, `control:origin.stale_summary`, `control:origin.omitted_summary`, `control:origin.mixed_access_chain`). `escape:coordinated-frontend-summary-lie` present exactly once and never promoted to `control:escape:coordinated-frontend-summary-lie` (count 0). Phase 1 and Phase 2 verify runs embedded in the same script both reported `status: pass` with their own goldens/controls unmoved (D-13). `go build ./...`, `go vet ./...`, and the full `go test ./...` suite all pass.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- ROADMAP SC3 and SC4's `omitted` clause are now true: a public borrowed view's declaration cannot silently omit a real, body-derivable origin, and the exported summary can no longer misreport a borrow-derived return as fully owned.
- `OWN-04` is now closable: both sibling gap-closure plans (03-08's CR-01 fix and this plan's D-03-02 fix) are complete, and `requirements.ready-ids`'s shared-requirement gate can now mark OWN-04 complete since both declaring plans (03-06/03-07 originally, 03-08/03-09 as the gap closures) have finished.
- D-03-01 remains open, accepted, non-blocking debt for Phase 4+: `loanLivenessFixpoint`'s linear cost still never reaches the admission-deciding code path, though OWN-03's observable truth is independently proven correct by the per-arm `discoverLoanLastUses` mechanism.
- No blockers for the phase's own end-of-phase re-verification against 03-VERIFICATION.md's two `failed` truths.

---
*Phase: 03-borrowed-views-and-cfg-lifetimes*
*Completed: 2026-09-04*

## Self-Check: PASSED

- FOUND: `.planning/phases/03-borrowed-views-and-cfg-lifetimes/03-09-SUMMARY.md`
- FOUND: `testdata/phase3/public_view_omitted.lang`
- FOUND: commit `d1977a3` (test(03-09): add failing test for omitted-origin rejection)
- FOUND: commit `5f556da` (feat(03-09): recompute origin unconditionally, refuse omitted publication)
- FOUND: commit `2eb7b8f` (feat(03-09): wire omitted-origin and mixed-access controls into Phase 3 gate)
- FOUND: commit `275ad38` (docs(03-09): close D-03-02, D-03-01 remains open and non-blocking)
