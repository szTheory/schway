---
phase: 03-borrowed-views-and-cfg-lifetimes
plan: "10"
subsystem: ownership
tags: [originvalidate, origin-recomputation, match-arms, gap-closure, phase3-gate]

requires:
  - phase: 03-borrowed-views-and-cfg-lifetimes
    provides: "03-06 through 03-09's single-arm origin recomputation and publication-path gate (RecomputeOrigin, ValidatePublished, control:origin.omitted_summary, control:origin.mixed_access_chain)"
provides:
  - "RecomputeOriginPerReturn as the package's sole backward-walk site — one per-return derivation per core.OpReturn, not just the first"
  - "RecomputeOrigin reimplemented as a pure conservative combiner over RecomputeOriginPerReturn, signature unchanged"
  - "AccessConflicting sentinel for arms that disagree on access mode, with a declared-access domain check making it undeclarable"
  - "control:origin.multi_arm_omitted and control:origin.multi_arm_access_conflict as fail-closed required Phase 3 gate controls (eleven total)"
  - "Two new corpus fixtures pinning the multi-arm omitted and multi-arm access-conflict shapes"
affects: [phase-04-cross-function-calls, own-04-requirement-closure]

actuals:
  tokens: 9028
  tasks: 3
  commits: 5

tech-stack:
  added: []
  patterns:
    - "Plural-primary/singular-combiner promotion: RecomputeOriginPerReturn is the one derivation; RecomputeOrigin degenerates to it, never the reverse"
    - "Conservative combination law: disagreement produces a named sentinel, never a silently-picked winner"

key-files:
  created:
    - testdata/phase3/public_view_multi_arm_omitted.lang
    - testdata/phase3/public_view_multi_arm_access_conflict.lang
  modified:
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/originvalidate/originvalidate_test.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_test.go
    - internal/compiler/testsupport/cli_test.go
    - scripts/verify-phase3.sh
    - .planning/phases/03-borrowed-views-and-cfg-lifetimes/03-DEBT.md

key-decisions:
  - "Promoted RecomputeOriginPerReturn to primary per the plan's assumption-delta decision; RecomputeOrigin is now a pure combiner, never a second backward-walk site (grep -c 'for current != function.Parameter.ID' == 1, verified)."
  - "Chose the AccessConflicting sentinel over a fourth problem code, reusing core.origin_access_mismatch's existing assertion target; made it undeclarable via a domain check that runs before any declared-vs-recomputed comparison."
  - "control:origin.multi_arm_access_conflict injects a declared origin (shared) onto the checked function rather than adding a new fixture per declared mode, since the undeclared and cross-declared-mode paths are already covered by originvalidate_test.go unit tests."

requirements-completed: [OWN-03, OWN-04]

coverage:
  - id: D1
    description: "A match-bodied function whose non-first arm returns a live, unreleased borrow is refused publication (core.origin_omitted) instead of exporting with no public_origin field"
    requirement: "OWN-04"
    verification:
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_test.go#TestMultiArmOmittedOriginRejected"
        status: pass
      - kind: integration
        ref: "internal/compiler/testsupport/cli_test.go#TestMultiArmOmittedOriginRejectedThroughCLI"
        status: pass
    human_judgment: false
  - id: D2
    description: "Two arms deriving different access modes resolve to neither arm's answer (AccessConflicting sentinel), refused with core.origin_access_mismatch whichever mode is declared, undeclared refused with core.origin_omitted, and the sentinel itself is never a declarable value"
    requirement: "OWN-04"
    verification:
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_test.go#TestMultiArmAccessConflictDerivesNeitherArm"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_test.go#TestDeclaredConflictingAccessIsRefused"
        status: pass
    human_judgment: false
  - id: D3
    description: "The Phase 3 gate requires both new multi-arm control IDs fail-closed (eleven total) with both expected escapes still surfaced and never promoted to detected controls"
    verification:
      - kind: integration
        ref: "sh scripts/verify-phase3.sh"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestVerifyPhase3ControlsAndWork"
        status: pass
    human_judgment: false
  - id: D4
    description: "The fix is load-bearing: reverting the per-return promotion in a throwaway detached worktree makes the new tests fail, and the shipped binary refuses an out-of-corpus hand-written multi-arm program"
    verification:
      - kind: manual_procedural
        ref: "mutation-kill in /tmp/ai-lang-mutation-kill (detached worktree, removed after use) plus shipped-binary run on owned.lens_peek"
        status: pass
    human_judgment: false

duration: 45min
completed: 2026-09-04
status: complete
---

# Phase 3 Plan 10: Multi-arm origin leak — every OpReturn is walked, disagreeing arms resolve to neither Summary

**`RecomputeOriginPerReturn` walks every `core.OpReturn` in a function (not just the first); `RecomputeOrigin` is now a conservative combiner that refuses to let any single match arm's answer win.**

## Performance

- **Duration:** 45 min
- **Started:** 2026-09-04T16:45:36Z
- **Completed:** 2026-09-04T17:30:00Z
- **Tasks:** 3
- **Files modified:** 9 (2 new fixtures, 7 modified)

## Accomplishments
- Closed the third-round gap 03-VERIFICATION.md found: a match-bodied function whose first arm returns owned and second arm returns a live borrow used to export with no `public_origin` field at all. `RecomputeOriginPerReturn` is now the package's sole backward-walk site, walking every `OpReturn` in operation order.
- `RecomputeOrigin` is reimplemented as a pure conservative combiner over that per-return slice per the plan's combination law: no borrow-derived return → not-ok (byte-identical to before); all agree → that access; disagree → the new `AccessConflicting` sentinel, never a silently-picked winner.
- Added a declared-access domain check to `ValidatePublished`: a declared `Access` outside `{shared, exclusive}` is refused before any comparison, so a mutated summary can never declare the sentinel and match a conflicting recomputation.
- Wired both new defects (`control:origin.multi_arm_omitted`, `control:origin.multi_arm_access_conflict`) into the Phase 3 gate — eleven required controls total, `sh scripts/verify-phase3.sh` passes clean with both expected escapes still surfaced and never promoted to detected controls.
- Proved the fix load-bearing with a mutation-kill (revert-and-fail in a throwaway detached worktree) and a shipped-binary run on an out-of-corpus hand-written program.
- Every single-arm outcome and every owned-arm fixture is byte-identical to before this plan; OWN-04 stays unchecked in REQUIREMENTS.md (phase verifier's decision).

## Task Commits

Each task followed RED-GREEN commits (tasks 1 and 2 are `tdd="true"`; task 3 is a single `auto` commit):

1. **Task 03-10-01: multi-arm omitted-origin leak**
   - `test(03-10)` `640b5f5` — failing tests + `public_view_multi_arm_omitted.lang` fixture (RED: `RecomputeOriginPerReturn` undefined)
   - `feat(03-10)` `96e54be` — `RecomputeOriginPerReturn` as the sole backward-walk site; `RecomputeOrigin` reimplemented as combiner (GREEN)
2. **Task 03-10-02: disagreeing arms / declared-access domain check**
   - `test(03-10)` `0a98d9f` — failing test + `public_view_multi_arm_access_conflict.lang` fixture (RED: declaring the sentinel on the fixture whose own recomputed answer IS the sentinel incorrectly reported no problems)
   - `feat(03-10)` `d576a7e` — declared-access domain check in `ValidatePublished` (GREEN)
3. **Task 03-10-03: wire both controls into the Phase 3 gate**
   - `feat(03-10)` `a057063` — two new lanes in `verifyBorrowedCorpus`, `scripts/verify-phase3.sh`'s required-control loop, `TestVerifyPhase3ControlsAndWork`/`TestPhase3VerifierScriptContract` extended, `03-DEBT.md` closure entry

**Plan metadata:** committed alongside this SUMMARY.

_Note: RED commits in this plan intentionally fail to build/pass — that is the expected RED-phase state for `tdd="true"` tasks._

## Files Created/Modified
- `testdata/phase3/public_view_multi_arm_omitted.lang` - first arm owned, second arm returns a live shared borrow, no declared origin (match functions cannot declare one); checks clean
- `testdata/phase3/public_view_multi_arm_access_conflict.lang` - first arm returns a live shared borrow, second arm returns a live exclusive borrow; checks clean
- `internal/compiler/originvalidate/originvalidate.go` - `ReturnOrigin`, `RecomputeOriginPerReturn`, `AccessConflicting`, reimplemented `RecomputeOrigin` combiner, `ValidatePublished`'s declared-access domain check
- `internal/compiler/originvalidate/originvalidate_test.go` - 10 new tests covering the per-return derivation, the multi-arm omitted case, the access-conflict case, the domain check, and the generalized every-return invariant
- `internal/compiler/session/session.go` - two new lanes inside `verifyBorrowedCorpus`'s existing `lane:borrowed-origin-controls` lane
- `internal/compiler/session/session_test.go` - `TestVerifyPhase3ControlsAndWork` extended to require both new control IDs
- `internal/compiler/testsupport/cli_test.go` - `TestMultiArmOmittedOriginRejectedThroughCLI`; `TestPhase3VerifierScriptContract` extended
- `scripts/verify-phase3.sh` - required-control loop extended to eleven IDs
- `.planning/phases/03-borrowed-views-and-cfg-lifetimes/03-DEBT.md` - dated closure entry; D-03-01 and WR-01 left open, unedited

## Decisions Made
- **Promote, not add-alongside** (per the plan's own `<assumption_delta_decision>`): `RecomputeOriginPerReturn` is the one and only backward-walk site; `RecomputeOrigin` is a pure combiner. Verified structurally: `grep -c 'for current != function.Parameter.ID' internal/compiler/originvalidate/originvalidate.go` prints exactly `1`.
- **Sentinel over a fourth problem code**: `AccessConflicting` (string `"conflicting"`) reuses `core.origin_access_mismatch` as its assertion target rather than introducing a new diagnostic code, keeping every existing control's assertion target stable (no new `lang.core` schema field, no new `OperationKind`).
- **Domain check placement**: the declared-access domain check runs immediately inside the `function.PublicOrigin != nil` branch, before the understated/mismatch comparisons — verified necessary, not redundant, by strengthening `TestDeclaredConflictingAccessIsRefused` to declare the sentinel on the ONE fixture whose own recomputed answer IS the sentinel (declared == recomputed); without the domain check this test failed RED (`[]` problems, i.e. incorrectly accepted).
- **`control:origin.multi_arm_access_conflict` injects rather than adds a third fixture**: the gate exercises the declared-and-conflicting path via mutation injection (mirroring the understated/impossible controls' existing pattern) since the frontend has no grammar for a per-arm origin declaration; the undeclared form of the same fixture is covered by `originvalidate_test.go` unit tests, and both paths are load-bearing (see Mutation-Kill below).

## Deviations from Plan

None — plan executed exactly as written. Both fixtures checked clean on the first draft (no adjustment needed per the plan's own contingency).

## Mutation-Kill (D-09)

Performed in a throwaway detached worktree (`/tmp/ai-lang-mutation-kill`, created via `git worktree add ... HEAD --detach` from commit `d576a7e`, removed via `git worktree remove --force` immediately after). Reverted `RecomputeOriginPerReturn`'s loop to capture only the first `core.OpReturn` (restoring the exact pre-fix defect, inline in the promoted function rather than the old standalone loop, so the revert exercises the same code path the fix replaced):

```diff
 	var returnOps []*core.LinearOperation
 	for index := range operations {
 		operation := operations[index]
 		if operation.Kind == core.OpReturn {
-			returnOps = append(returnOps, &operations[index])
+			if len(returnOps) == 0 {
+				returnOps = append(returnOps, &operations[index])
+			}
 			continue
 		}
 		sourceOf[operation.TargetID] = operation
 	}
```

Verbatim failing output:

```
=== RUN   TestMultiArmOmittedOriginRejected
    originvalidate_test.go:290: expected RecomputeOrigin to succeed on the multi-arm leak
--- FAIL: TestMultiArmOmittedOriginRejected (0.00s)
=== RUN   TestMultiArmAccessConflictDerivesNeitherArm
    originvalidate_test.go:324: expected exactly 2 per-return origins, got [{OperationID:s1:owned.public_view_multi_arm_access_conflict:fn:choose:op:2 Paths:[flag] Access:shared Derived:true}]
--- FAIL: TestMultiArmAccessConflictDerivesNeitherArm (0.00s)
FAIL
FAIL	github.com/codename-lang/lang/internal/compiler/originvalidate	0.196s
```

```
=== RUN   TestMultiArmOmittedOriginRejectedThroughCLI
    cli_test.go:518: expected non-zero exit for a multi-arm omitted origin: {'{"schema":"lang.command/0","command":"interface","status":"pass", ... "functions":[{"id":"s1:owned.public_view_multi_arm_omitted:fn:choose","name":"choose"}]},"lanes":[],"expected_escapes":["escape:coordinated-frontend-summary-lie"], ... }' Stderr:[] Exit:0}
--- FAIL: TestMultiArmOmittedOriginRejectedThroughCLI (0.28s)
FAIL	github.com/codename-lang/lang/internal/compiler/testsupport	0.449s
```

All three named tests fail against the reverted production code — the promotion is proven load-bearing, not incidentally passing.

## Shipped-Binary Out-of-Corpus Evidence (D-11)

Built `./cmd/lang` fresh from the final tree. Ran a hand-written program not present in any corpus, with module/type/alternative/parameter naming distinct from every phase3 fixture:

```
module owned.lens_peek

export {
  type Lens
  fn peek
}

data Lens =
  | Open
  | Shut

fn peek(state: Lens) -> Lens {
  match state {
    Open => {
      let claimed = take state
      claimed
    }
    Shut => {
      let glance = borrow state
      glance
    }
  }
}
```

- `lang --json check lens_peek.lang`: `"status":"pass"`, zero diagnostics, **exit 0**.
- `lang --json interface export lens_peek.lang summary.json`: `"status":"invalid"`, one diagnostic `core.origin_omitted` ("no declared origin, but body derives origin [state] with access \"shared\""), **exit 2**.
- `summary.json` was never created (`ls` reports "No such file or directory").

## Non-Regression Evidence

- All five pre-existing origin/branch fixtures re-run through the freshly-built shipped binary, unchanged outcomes: `public_view_omitted.lang` → exit 2, `core.origin_omitted`; `public_view_mixed_access.lang` → exit 2, `core.origin_access_mismatch`; `public_view.lang` → exit 0, publishes with `paths:["buffer"],access:"shared"`; `branch_view.lang` → exit 0, no `public_origin`; `borrowed_view.lang` → exit 0, no `public_origin`.
- All six pre-existing `originvalidate` unit tests pass unchanged (`TestOriginUnderstatedRejected`, `TestOriginAccessMismatchRejected`, `TestMixedAccessChainDerivesShared`, `TestMixedAccessChainRejectedAsAccessMismatch`, `TestOmittedOriginRejected`, `TestStaleSummaryRejectedBeforeOtherChecks`).
- `go build ./...`, `go vet ./...`, `go test ./...` all pass clean.
- `sh scripts/verify-phase3.sh` exits 0: eleven required control IDs present, `escape:coordinated-frontend-summary-lie` surfaced once as an expected escape and zero times as `control:escape:coordinated-frontend-summary-lie`.
- Phase 1 and Phase 2 goldens untouched (`git status --short testdata/` shows no golden diffs); Phase 1/Phase 2 verify invocations inside the gate script pass unchanged.
- `grep -c 'D-03-01' .planning/phases/03-borrowed-views-and-cfg-lifetimes/03-DEBT.md` → 4 (still present, not dropped).
- `grep -n 'OWN-04' .planning/REQUIREMENTS.md` → still `[ ]`/"Gaps Found" — not flipped by this plan.

## What Inputs the Green Tests Actually Reach

Per the standing process debt (interrogate what inputs a property test actually reaches): `TestPublishedOriginConsistentWithEveryReturn` walks every `.lang` file in `testdata/phase3` and checks every function's combined answer against its per-return elements — but it SKIPS any file that produces diagnostics, so it only ever reaches check-clean programs (by design: `RecomputeOrigin`/`RecomputeOriginPerReturn` are never called on a rejected program in the real pipeline either). It also only reaches functions whose `Linear` body is non-nil; a `Match`-only function with zero `Linear.Operations` (none exist in the current corpus — every match-bodied fixture in `testdata/phase3` has linear arm bodies) would return `nil` from `RecomputeOriginPerReturn` and be silently skipped by the invariant loop rather than asserted on, since `derivedCount == 0` in that case is indistinguishable from "no arm derives a borrow." This is the same "every return owned" degenerate case §2 of the combination law already covers correctly (not-ok), so the shape is not unreachable in the sense of untested outcome, but the invariant test's own coverage of a zero-`Linear.Operations` shape specifically is not currently exercised by any phase3 fixture — recorded here per the standing process rule rather than left implicit.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
ROADMAP SC3 and SC4 now hold for multi-arm functions as well as single-arm ones. Both 03-VERIFICATION.md gaps (SC3, SC4) are closed with fail-closed regression controls. D-03-01 and WR-01 remain open, accepted, non-blocking debt items carried unchanged into Phase 4. OWN-04's requirement checkbox in REQUIREMENTS.md is deliberately left unchecked — flipping it, and re-running the phase verifier to confirm no further gap exists, is the next step outside this plan's scope.

---
*Phase: 03-borrowed-views-and-cfg-lifetimes*
*Completed: 2026-09-04*

## Self-Check: PASSED

All key files verified present on disk; all five task commit hashes (640b5f5, 96e54be, 0a98d9f, d576a7e, a057063) verified present in git log.
