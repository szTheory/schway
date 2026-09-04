---
phase: 03-borrowed-views-and-cfg-lifetimes
verified: 2026-09-04T00:00:00Z
status: gaps_found
score: 3/4 must-haves verified
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 2/4
  gaps_closed:
    - "GAP 1 / CR-01 (originvalidate mixed-access reborrow-chain misreport) — RecomputeOrigin's OpBorrowExclusive branch is now first-seen guarded; independently reproduced fixed against the shipped binary (public_view_mixed_access.lang now exits 2 with core.origin_access_mismatch)"
    - "GAP 2 / D-03-02 (undeclared origin exports as if fully owned, single-arm shape) — ValidatePublished now runs RecomputeOrigin unconditionally and refuses publication with core.origin_omitted; independently reproduced fixed against the shipped binary (public_view_omitted.lang now exits 2 with core.origin_omitted)"
  gaps_remaining: []
  regressions:
    - "New gap discovered by the post-gap-closure code review (03-REVIEW.md, committed 9ab2e94) and independently reproduced here: RecomputeOrigin only inspects the FIRST OpReturn in a function's flat Operations list. A match-arm-bodied function whose non-first arm returns a live borrow evades both the access-mismatch and the newly-added omitted-origin gate entirely. This is not a regression in the literal sense (the single-arm gaps are genuinely closed) but a previously-undiscovered instance of the same root defect class, surfaced only after the gap-closure plans landed, and it defeats the very SC3/SC4 clauses this round of gap closure was meant to satisfy."
gaps:
  - truth: "A public borrowed view names all verified field/alternative origins and access modes without downstream body inspection (ROADMAP SC3)"
    status: failed
    reason: "Confirmed live and reachable, independently of 03-REVIEW.md's own reproduction: originvalidate.RecomputeOrigin builds its backward-walk chain from only the first core.OpReturn encountered while scanning function.Linear.Operations in index order; every subsequent OpReturn (one per match arm, per check.go's own arm-lowering, which appends all arms' returns into the same flat Operations slice) is never examined. A match-bodied function whose first arm returns an owned value and whose second arm returns a live, unreleased borrow of the parameter is exported with NO public_origin field at all — indistinguishable from a function that owns its return outright."
    artifacts:
      - path: "internal/compiler/originvalidate/originvalidate.go"
        issue: "RecomputeOrigin (lines ~59-76): `if returnOp == nil { returnOp = &operations[index] }` only ever captures the first OpReturn; later arms' OpReturn operations are skipped via `continue` and never walked."
    missing:
      - "RecomputeOrigin must collect every OpReturn in the function, walk backward from each, and combine per-arm answers conservatively (any borrow-derived arm with no covering declaration must trigger core.origin_omitted; disagreeing access modes across arms must not silently resolve to one arm's answer)"
      - "A regression fixture pairing an owned-returning arm with a borrow-returning arm (or two arms disagreeing on access mode), wired as a new required control in scripts/verify-phase3.sh — the existing control:origin.omitted_summary and control:origin.mixed_access_chain controls only exercise the single-arm (non-match) shape and do not catch this"
  - truth: "Separate compilation rejects stale, omitted, or impossible public-origin summaries while retaining the coordinated-frontend-lie limitation explicitly (ROADMAP SC4)"
    status: failed
    reason: "'Omitted' is explicitly named in this criterion's text. For the single-arm shape it is now genuinely rejected (verified live: public_view_omitted.lang exits 2 with core.origin_omitted). But for a match-arm-bodied function, the same category of omission is NOT rejected — verified live against the shipped binary with a hand-written, out-of-corpus program (owned.branch_origin_leak: first arm returns owned, second arm returns a live borrow). `lang check` passes; `lang interface export` exits 0 and writes a summary for `choose` with no public_origin field at all. This is the same root cause as the SC3 gap above (RecomputeOrigin's first-OpReturn-only scan) and it defeats the omitted-origin gate 03-09 built specifically to close this criterion's 'omitted' clause."
    artifacts:
      - path: "internal/compiler/originvalidate/originvalidate.go"
        issue: "Same root cause as SC3's gap — ValidatePublished calls RecomputeOrigin unconditionally now (03-09's fix), but RecomputeOrigin itself is blind to every arm but the first, so a branch-shaped function's real omission is never detected."
    missing:
      - "Same fix as SC3: RecomputeOrigin must reason over every OpReturn, not just the first"
      - "A control/test asserting interface export rejects a match-arm-bodied function whose non-first arm returns a borrow-derived value with no covering declaration (currently: exit 0, silent success, confirmed against the shipped ./cmd/lang binary)"
deferred: []
---

# Phase 3: Borrowed Views and CFG Lifetimes Verification Report

**Phase Goal:** Local borrows remain ergonomic through precise last-use inference while public borrowed results remain explicit and separately checkable.
**Verified:** 2026-09-04
**Status:** gaps_found
**Re-verification:** Yes — after gap closure (03-08, 03-09)

## Re-Verification Summary

The prior 03-VERIFICATION.md (initial verification) failed SC3/SC4 on two grounds: GAP 1 (CR-01, a mixed shared/exclusive reborrow chain misreporting a stronger access mode than the body grants) and GAP 2 (D-03-02, an undeclared origin exporting identically to a fully-owned return). Both gap-closure plans (03-08, 03-09) have landed. I re-tested both original findings directly against the codebase and the shipped binary rather than trusting the SUMMARY.md or 03-DEBT.md closure claims.

**Both original gaps are genuinely closed** for the shape they were built against (single-arm, straight-line, non-match functions):

- `public_view_mixed_access.lang` (the exact CR-01 shape: `borrow mut buffer` then `borrow` the reborrow, declared `borrow mut(buffer)`) now produces `core.origin_access_mismatch` and exits 2 through the shipped `./cmd/lang` binary. Reading `originvalidate.go:88-91` confirms `OpBorrowExclusive` is now guarded first-seen, symmetric with `OpBorrowShared`.
- `public_view_omitted.lang` (the exact D-03-02 shape: `relay`'s `borrow mut`/reborrow/return-the-exclusive-loan with no declared origin) now produces `core.origin_omitted` and exits 2 through the shipped binary. Reading `originvalidate.go:119-127` confirms `ValidatePublished` now calls `RecomputeOrigin` unconditionally, including when `PublicOrigin == nil`.
- Both fixtures are wired into `scripts/verify-phase3.sh`'s required-control set (nine controls total, up from seven); the gate was re-run live and passes with all nine controls present.

**However, a fresh code review performed after the gap-closure plans landed (03-REVIEW.md, committed `9ab2e94`) found a new Critical, and per this verification's explicit instruction I independently reproduced it rather than taking the review at face value in either direction.**

### Independent reproduction of the review's Critical finding

I read `internal/compiler/originvalidate/originvalidate.go` directly (not the review's excerpt) and confirmed the code is exactly as described:

```go
var returnOp *core.LinearOperation
for index := range operations {
    operation := operations[index]
    if operation.Kind == core.OpReturn {
        if returnOp == nil {
            returnOp = &operations[index]
        }
        continue
    }
    sourceOf[operation.TargetID] = operation
}
```

`returnOp` is captured only once, on the first `OpReturn` encountered. This is correct for a straight-line function (exactly one `OpReturn`), but `check.go`'s match-arm lowering appends one `OpReturn` per arm into the same flat `function.Linear.Operations` slice — so for a branch-shaped function, every arm past the first is invisible to `RecomputeOrigin`.

I then built the shipped binary (`go build -o /tmp/lang ./cmd/lang`) and ran a hand-written, out-of-corpus program through it — not the review's own reproduction, an independently written one with different naming and structure:

```
module owned.branch_origin_leak
export { type Switch, fn choose }
data Switch = | On | Off
fn choose(flag: Switch) -> Switch {
  match flag {
    On => { let moved = take flag; moved }
    Off => { let view = borrow flag; view }
  }
}
```

- `lang --json check` on this source: `"status":"pass"`, zero diagnostics.
- `lang --json interface export` on this source: `"status":"pass"`, exit 0. The written summary:
  ```json
  {"functions":[{"id":"...:fn:choose","name":"choose","parameter":{...},"return_type":"Switch","abilities":["copy","drop","share","send","escape"]}]}
  ```
  There is **no `public_origin` field** at all — structurally identical to a function that owns and returns a brand-new `Switch`, even though the `Off` arm hands the caller a live, unreleased borrow of `flag`.

I confirmed no fixture in `testdata/phase3/` exercises this shape: `branch_view.lang` and `borrowed_view.lang` (the phase's only match-bodied fixtures) return an owned (`take`d) value from every arm; `public_view_mixed_access.lang` and `public_view_omitted.lang` (the phase's only origin fixtures) are both single-arm straight-line functions. `scripts/verify-phase3.sh`'s nine required controls do not cover this shape, confirmed by re-running the gate live (passes, but exercises none of this).

**Verdict: this is a real, independently-reproduced, live gap against SC3 and SC4 as literally worded**, in the exact mechanism (`originvalidate`) both gap-closure plans touched without tracing what "first `OpReturn`" means once a function has more than one. It is not a regression of the closed gaps (the single-arm fixtures still pass); it is a previously-undiscovered instance of the same defect class, surfaced by the post-closure review's deeper scrutiny of "the least-reviewed part of the phase."

I also note `.planning/REQUIREMENTS.md` currently marks `OWN-04` as `[x]`/"Complete" (set by the 03-09 gap-closure commit, `fe5c638`). That mark was premature: it was written before the post-closure review ran and found this new Critical. `OWN-04` should remain unchecked until this gap is closed.

## Goal Achievement

### Observable Truths (ROADMAP Success Criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Shared and exclusive borrow conflicts are accepted or rejected identically by the core checker and bounded path oracle | ✓ VERIFIED | `internal/compiler/pathoracle` package (imports neither `check` nor `ast`), `control:cfg.path_oracle_disagreement` lane re-run live, exit 0; unchanged from initial verification, re-confirmed |
| 2 | A branch-specific last use ends a loan on the correct CFG edge without a manual scope block; omitting that edge is detected | ✓ VERIFIED | Empirically re-confirmed against the shipped binary: `branch_one_arm_shared_accept.lang` checks clean; `branch_one_arm_shared_reject.lang` rejected with `ownership.move_while_borrowed`. D-03-01's architectural note (mechanism is `discoverLoanLastUses` per-arm, not the CFG dataflow) is carried, non-blocking, unchanged from initial verification |
| 3 | A public borrowed view names all verified field/alternative origins and access modes without downstream body inspection | ✗ FAILED | GAP 1/CR-01 (original) and GAP 2/D-03-02 are genuinely closed for single-arm functions (verified live). A NEW, independently-reproduced defect remains: a match-arm-bodied function whose non-first arm returns a live borrow exports with no `public_origin` field at all — `RecomputeOrigin` only inspects the first `OpReturn` in `Linear.Operations` |
| 4 | Separate compilation rejects stale, omitted, or impossible public-origin summaries while retaining the coordinated-frontend-lie limitation explicitly | ✗ FAILED | "Stale" rejected (`TestStaleSummaryRejectedBeforeOtherChecks`). "Omitted" and "impossible" are now correctly rejected for single-arm functions (verified live: `public_view_omitted.lang` → `core.origin_omitted`, `public_view_mixed_access.lang` → `core.origin_access_mismatch`), but "omitted" is NOT rejected for a multi-arm match-bodied function whose non-first arm returns a live borrow — verified live against the shipped binary (`owned.branch_origin_leak`, exit 0, no `public_origin` field). Coordinated-frontend-lie limitation is retained and named (`escape:coordinated-frontend-summary-lie`) |

**Score:** 3/4 truths verified (0 present, behavior-unverified) — SC1 and SC2 unchanged and verified; SC3 and SC4 remain failed on a newly-discovered instance of the same root-cause class as the originally-reported gaps

### Deferred Items

None.

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/compiler/check/check.go` (loanLivenessFixpoint, materializeLoanEndpoints, discoverLoanLastUses) | Edge-specific loan liveness | ✓ VERIFIED (behavior), ⚠️ architecturally split (D-03-01, carried, non-blocking) | Unchanged from initial verification |
| `internal/compiler/pathoracle/pathoracle.go` | Independent bounded path-oracle re-derivation | ✓ VERIFIED | Unchanged from initial verification |
| `internal/compiler/corevalidate/corevalidate.go` (recomputeLoanEndpoints) | Independent re-derivation via reachability closure | ✓ VERIFIED | Unchanged from initial verification |
| `internal/compiler/originvalidate/originvalidate.go` | Independent origin/access re-derivation | ✗ Real bug (new instance) | `RecomputeOrigin`'s single-arm gaps (CR-01 original, D-03-02) are fixed and regression-tested; the function still only inspects the first `OpReturn`, so a multi-arm (match-bodied) function's non-first-arm borrow-derived return evades both `core.origin_access_mismatch` and `core.origin_omitted` — confirmed live against the shipped binary |
| `scripts/verify-phase3.sh` | Phase 3 gate requiring every required control | ✓ VERIFIED (for what it covers) | Re-run live: exit 0, nine required control IDs present, both expected escapes present. Does not cover the multi-arm origin-leak shape — no control exists for it |

### Data-Flow Trace (Level 4)

`originvalidate.RecomputeOrigin`'s output flows to `ValidatePublished`'s comparison and from there to `interface export`'s CLI output and the written summary file — confirmed genuinely wired and input-dependent (the three hand-written programs used in this verification — mixed-access, omitted single-arm, and multi-arm branch-origin-leak — produced three different, input-dependent outcomes: rejected/rejected/silently-accepted). The remaining defect is not a stub or disconnected path; it is a real computation that is blind to part of its own input (every `OpReturn` past the first).

### Key Link Verification

| From | To | Via | Status | Details |
|------|-----|-----|--------|---------|
| `check.go` PublicOrigin construction | `originvalidate.ValidatePublished` | `session.InterfaceExportCommandFile` only | ⚠️ PARTIAL | Unchanged from initial verification (WR-01, carried, non-blocking): not wired into `lang check`/`lang run` |
| `RecomputeOrigin`'s per-return backward walk | Every `OpReturn` in a function | function iteration | ✗ NOT_WIRED (new finding) | Confirmed by reading `originvalidate.go:59-70`: only the first `OpReturn` encountered is captured into `returnOp`; every subsequent `OpReturn` in the same `Operations` slice (one per match arm) is skipped via `continue` and never walked backward |
| `loanLivenessFixpoint` (`checkBranch`) | Admission decision (`conflictingLoan`/`expiringLoans`) | none | ✗ NOT_WIRED | Unchanged from initial verification (D-03-01, carried, non-blocking) |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Mixed-access reborrow chain (GAP 1/CR-01, closed) | `lang interface export public_view_mixed_access.lang` | exit 2, `core.origin_access_mismatch` | ✓ PASS |
| Omitted origin, single-arm (GAP 2/D-03-02, closed) | `lang interface export public_view_omitted.lang` | exit 2, `core.origin_omitted` | ✓ PASS |
| Omitted origin, non-first match arm returns a live borrow (new, hand-written, out-of-corpus) | `lang interface export` on `owned.branch_origin_leak` (first arm owned, second arm shared borrow) | exit 0, no `public_origin` field in the written summary — should reject or flag | ✗ FAIL |
| Phase 3 gate | `sh scripts/verify-phase3.sh` | exit 0, all 9 required controls + 2 expected escapes present | ✓ PASS (does not cover the new gap's shape) |

### Probe Execution

Not applicable — no `scripts/*/tests/probe-*.sh` files exist for this phase; `scripts/verify-phase3.sh` is the phase's own gate script and was run directly above.

### Requirements Coverage

| Requirement | Source Plans | Description | Status | Evidence |
|-------------|--------------|--------------|--------|----------|
| OWN-03 | 03-01, 03-02, 03-03, 03-04, 03-05, 03-07 | Shared and exclusive loans obey conflict rules and ordinary local loans end at proven CFG point/edge-specific last use | ✓ SATISFIED | SC1, SC2 verified; unchanged from initial verification. `.planning/REQUIREMENTS.md` correctly shows `[ ]`/"Pending" |
| OWN-04 | 03-06, 03-07, 03-08, 03-09 | Public borrowed results record verified field/alternative origins and access mode without inspecting provider bodies downstream | ✗ BLOCKED | SC3, SC4 both fail on a newly-discovered defect class (multi-arm origin leak). `.planning/REQUIREMENTS.md` currently shows `[x]`/"Complete" (set by commit `fe5c638`, 03-09) — this is premature and should be reverted to Pending until the multi-arm gap is closed |

All requirement IDs declared across the nine plans (`OWN-03`: 03-01/02/03/04/05/07; `OWN-04`: 03-06/07/08/09) are accounted for; no orphaned requirement IDs found in `REQUIREMENTS.md`'s Phase 3 row beyond these two.

### Anti-Patterns Found

No `TODO`/`FIXME`/`XXX`/`TBD`/`HACK`/`PLACEHOLDER` markers found in the phase's modified files. The defect found is a real logic gap (incomplete iteration over a multi-valued structure), not stub/placeholder code, and is already honestly recorded in `03-REVIEW.md`'s renumbered `CR-01`.

## Non-Regression

- `go build ./...`, `go vet ./...` pass.
- `go test ./internal/compiler/originvalidate/... -run 'TestMixedAccess|TestOrigin' -v` — all 6 tests pass, confirming the two closed gaps are genuinely regression-tested, not just fixed incidentally.
- `sh scripts/verify-phase3.sh` re-run live: exit 0, nine required controls present.
- `lang check`/`lang interface export` re-run directly against the shipped binary for all three hand-written probe programs used in this report (mixed-access, omitted single-arm, multi-arm branch-origin-leak).

## Gaps Summary

The two gaps from the initial verification (GAP 1/CR-01, GAP 2/D-03-02) are genuinely, verifiably closed — confirmed independently against the shipped binary, not merely by reading the SUMMARY.md claims. Both are now regression-tested and wired into the Phase 3 gate as required controls.

However, a new instance of the same defect class was found by the post-closure code review and independently reproduced here: **`originvalidate.RecomputeOrigin` only examines the first `OpReturn` in a function's `Linear.Operations` list.** `check.go`'s match-arm lowering appends one `OpReturn` per arm into that same flat list, so a match-bodied function whose first arm returns owned and whose second (or later) arm returns a live, unreleased borrow is exported with no `public_origin` field at all — the exact "omitted" failure mode ROADMAP SC4 names, for a shape the 03-09 gap-closure plan's own fix did not anticipate. This is confirmed live and reachable through the shipped `./cmd/lang` binary on an honest, hand-written, out-of-corpus program requiring no adversarial mutation to reach — the same severity class as the original CR-01 and D-03-02 findings.

**Recommended remediation before OWN-04 is marked complete:**
1. Fix `RecomputeOrigin` to walk backward from every `OpReturn` in `function.Linear.Operations`, not just the first, and combine per-arm results conservatively (any arm deriving an uncovered borrow triggers `core.origin_omitted`; disagreeing access modes across arms must not silently resolve to one arm's answer).
2. Add a regression fixture pairing an owned-returning arm with a borrow-returning arm (or two arms with disagreeing access modes), and wire a new required control into `scripts/verify-phase3.sh`.
3. Revert `.planning/REQUIREMENTS.md`'s `OWN-04` row from `[x]`/"Complete" back to `[ ]`/"Pending" until the above is closed.

D-03-01 (OWN-03's architectural debt — the CFG dataflow's linear cost never reaches the admission-deciding code path) remains open, accepted, and non-blocking, unchanged from the initial verification; it does not affect OWN-03's observable truth.

---

_Verified: 2026-09-04_
_Verifier: Claude (gsd-verifier)_
