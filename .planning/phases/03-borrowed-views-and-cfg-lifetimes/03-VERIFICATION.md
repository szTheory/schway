---
phase: 03-borrowed-views-and-cfg-lifetimes
verified: 2026-09-04T00:00:00Z
status: gaps_found
score: 2/4 must-haves verified
behavior_unverified: 0
overrides_applied: 0
gaps:
  - truth: "A public borrowed view names all verified field/alternative origins and access modes without downstream body inspection (ROADMAP SC3)"
    status: failed
    reason: "Two independently confirmed, empirically reproduced defects mean the exported summary is not trustworthy as 'all verified origins/access modes': (1) originvalidate.RecomputeOrigin reports a stronger access mode than the body actually grants for a mixed-access reborrow chain (03-REVIEW.md CR-01), so a declared borrow mut(x) can pass verification against a body that only ever hands back a shared reborrow; (2) a borrow-derived return with NO declared origin exports a FunctionSignature indistinguishable from a fully-owned return (03-DEBT.md D-03-02) — the summary silently omits a real, verifiable origin rather than naming it."
    artifacts:
      - path: "internal/compiler/originvalidate/originvalidate.go"
        issue: "RecomputeOrigin (lines ~75-100): the OpBorrowExclusive case unconditionally overwrites derivedAccess on every hop, unlike the guarded OpBorrowShared case, so an outer exclusive borrow further from the return silently overrides a closer, correct shared answer. ValidatePublished (lines ~110-135) never invokes RecomputeOrigin when function.PublicOrigin == nil, so an undeclared-but-real origin is never detected or reported."
    missing:
      - "Fix RecomputeOrigin's OpBorrowExclusive branch to guard on first-seen (`if derivedAccess == \"\" { derivedAccess = \"exclusive\" }`), matching the shared case, plus a regression fixture exercising a mixed shared/exclusive reborrow chain"
      - "An admission-time or ValidatePublished-time check that runs RecomputeOrigin unconditionally for exported functions and rejects/flags a real body-derived origin the declaration omitted entirely"
  - truth: "Separate compilation rejects stale, omitted, or impossible public-origin summaries while retaining the coordinated-frontend-lie limitation explicitly (ROADMAP SC4)"
    status: failed
    reason: "'Omitted' summaries are explicitly named in this success criterion's text and are the one category confirmed NOT rejected: lang interface export on a function with a borrow-derived return but no declared origin annotation (e.g. the pre-existing, already-shipped exclusive_borrow_clean/relay shape) exits 0 and produces a FunctionSignature with no public_origin field at all — verified directly against the shipped ./cmd/lang binary, not inferred from a test. 'Impossible' summaries are also incompletely caught: CR-01's access-mode bug (see the paired gap above) means a genuinely impossible borrow mut(x) declaration against a shared-reborrow body is NOT rejected by ValidatePublished, because RecomputeOrigin itself computes the wrong (matching) answer. 'Stale' summaries are correctly rejected (TestStaleSummaryRejectedBeforeOtherChecks) and are not implicated."
    artifacts:
      - path: "internal/compiler/originvalidate/originvalidate.go"
        issue: "Same as above — ValidatePublished's PublicOrigin==nil short-circuit means 'omitted' is definitionally never reached by this control, and CR-01 means a subset of 'impossible' declarations pass as if valid."
    missing:
      - "A control/test that constructs an exported, borrow-derived-return function with no declared origin and asserts interface export rejects or otherwise flags it (currently: exit 0, silent success)"
      - "A regression fixture for the mixed-access reborrow chain asserting ValidatePublished correctly returns core.origin_access_mismatch"
deferred: []
---

# Phase 3: Borrowed Views and CFG Lifetimes Verification Report

**Phase Goal:** Local borrows remain ergonomic through precise last-use inference while public borrowed results remain explicit and separately checkable.
**Verified:** 2026-09-04
**Status:** gaps_found
**Re-verification:** No — initial verification

## Adjudication of the Three Open Findings Named in the Verification Request

### 1. CR-01 (originvalidate mixed-access bug) — bears on OWN-04. Confirmed live and reachable.

I read `internal/compiler/originvalidate/originvalidate.go` directly (not the review's excerpt) and confirmed the code is exactly as CR-01 describes: `RecomputeOrigin`'s `OpBorrowExclusive` case unconditionally overwrites `derivedAccess` on every backward hop, while the `OpBorrowShared` case is correctly guarded to only set the value once (first-seen, i.e. closest to the return).

I then built the shipped binary (`go build -o /tmp/lang ./cmd/lang`) and ran a hand-written program through it, per the request's instruction to determine reachability empirically rather than by inspection alone:

```
fn view(buffer: Buffer) -> borrow mut(buffer) Buffer {
  let y = borrow mut buffer
  let z = borrow y
  z
}
```

`lang check` passes (exit 0). `lang interface export` also passes (exit 0) and its exported summary reads `paths=[buffer] access=exclusive` — the tool asserts the caller receives write access to `buffer`, when the function body's actual final return (`z`) is only ever a shared reborrow of the exclusively-borrowed `y`. This is a real, reachable false claim of a stronger capability than the body grants, produced by the only code in the compiler that checks a declared origin's access mode against the body (`check.go` copies the declaration verbatim with no verification — confirmed by reading `check.go` around the `PublicOrigin` construction).

No fixture or test in the shipped test suite exercises this shape (`grep` for mixed-access/reborrow patterns across `originvalidate/` and `testdata/phase3/` returns nothing).

**Verdict: OWN-04 is NOT fully earned.** This is exactly the kind of defect OWN-04 exists to prevent, it is empirically reachable through the shipped binary today, and it is untested. This alone is sufficient to fail ROADMAP Success Criterion 3 ("names all verified... access modes").

### 2. D-03-01 (loanLivenessFixpoint's output unused for admission) — bears on OWN-03. Assessed as non-blocking; OWN-03 is earned.

I independently confirmed the mechanical claim by reading `check.go`: `loanLivenessFixpoint`/`materializeLoanEndpoints` populate `linear.LoanEndpoints` only inside `checkBranch`, after the admission verdict from `analyzeArmBody` has already been returned (`check.go:341-352`); `discoverLoanLastUses` (called at `check.go:700` inside `analyzeArmBody` and `check.go:1019` inside `analyzeStraightLine`) is what feeds `loan.lastUse`, which drives `conflictingLoan`/`expiringLoans` — the actual accept/reject decision, in every function the checker admits.

However, `discoverLoanLastUses` is invoked *per arm* (via `analyzeArmBody`, called once per match arm with the arm's own body), and per 03-01's per-arm-aliasing design each arm is a straight-line body with no internal branch — so a per-arm transitive last-use scan is, by construction, already edge-specific: there is exactly one edge per arm, and the loan's last use within that arm's body is that edge's last use. I verified this behaviorally, not just by argument, running the shipped binary on the `branch_one_arm_shared_accept.lang`/`branch_one_arm_shared_reject.lang` fixture pair (a loan that ends before a same-arm move is accepted; a loan still live at a same-arm move is rejected with a causal diagnostic naming the loan) — both behaved exactly as required, with the omission-detection half also directly true (the reject fixture is in fact the "omitting that edge is detected" case).

The independent-oracle argument in 03-DEBT.md is not merely asserted — I confirmed `TestBranchSequenceExhaustive` and `TestOwnershipSequenceExhaustive` exist and are part of the standing (passing) suite, giving genuine independent-mechanism corroboration that `discoverLoanLastUses`'s per-arm answer is correct.

**Verdict: OWN-03's observable truth (SC1, SC2) holds.** The mechanism that decides is not the newly built CFG dataflow, which is a real, documented gap against D-05's original ambition (retiring the quadratic law) and against the literal words "proven CFG point/edge-specific" — the proof exists, but via a different, per-arm mechanism than the CFG dataflow that reports the same fact for corroboration only. This is legitimate architectural debt (rightly recorded, not blocking), not an observable behavior failure. I am not blocking OWN-03 on it, consistent with 03-05's mid-phase gate, whose reasoning I independently re-derived from the code rather than accepting on faith.

### 3. D-03-02 (undeclared origin exports as if fully owned) — bears on OWN-04. Confirmed live and reachable; blocks SC4 directly.

Built and ran the shipped binary against a hand-written program shaped exactly like the pre-existing, already-shipped `exclusive_borrow_clean` fixture but exported:

```
fn relay(buffer: Buffer) -> Buffer {
  let view = borrow mut buffer
  let reviewed = borrow view
  view
}
```

`lang check` passes. `lang interface export` also passes (exit 0) and produces `{"return_type":"Buffer","abilities":["drop","share","send","escape"]}` with **no `public_origin` field at all** — structurally identical to a function that owns and returns a brand-new `Buffer`. `ValidatePublished`'s `if function.PublicOrigin == nil { continue }` guard means `RecomputeOrigin` is never invoked to check whether an origin *should* have been declared.

ROADMAP Success Criterion 4's own wording is: "Separate compilation rejects stale, omitted, or impossible public-origin summaries." "Omitted" is named explicitly, and this is precisely the case that is not rejected — confirmed against the shipped binary, not inferred from a summary claim. 03-06-PLAN.md's own `<done>` block claimed "Stale, omitted, and impossible public-origin summaries are all rejected by recomputation" — that claim is false for the omitted case, as 03-06's own SUMMARY.md deviations section documents (it deliberately dropped the check.go-side rejection to avoid breaking `exclusive_borrow_clean`, substituting only a parse-time grammar rule that governs the *spelling* of an annotation once one is written, not whether one must be written at all).

**Verdict: OWN-04 is NOT earned on SC4** for the same underlying reason as SC3 above — the two findings share one root cause (`ValidatePublished`'s unconditional skip on `PublicOrigin == nil`, and `check.go`'s zero-verification copy of the declaration).

## Goal Achievement

### Observable Truths (ROADMAP Success Criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Shared and exclusive borrow conflicts are accepted or rejected identically by the core checker and bounded path oracle | ✓ VERIFIED | `internal/compiler/pathoracle` package (imports neither `check` nor `ast`, confirmed by its own independence test), `control:cfg.path_oracle_disagreement` lane passes in `scripts/verify-phase3.sh` (re-run live, exit 0), mutation-kill precedent (`TestOracleDisagreesWithUniformJoinFault`) exists in the standing suite |
| 2 | A branch-specific last use ends a loan on the correct CFG edge without a manual scope block; omitting that edge is detected | ✓ VERIFIED | Empirically confirmed against the shipped binary: `branch_one_arm_shared_accept.lang` (loan ends before same-arm move) checks clean; `branch_one_arm_shared_reject.lang` (loan live at same-arm move) rejected with `ownership.move_while_borrowed`, cause chain names the loan. See adjudication #2 above for the D-03-01 caveat (mechanism is `discoverLoanLastUses` per-arm, not the new CFG dataflow, which is independently-corroborating only) |
| 3 | A public borrowed view names all verified field/alternative origins and access modes without downstream body inspection | ✗ FAILED | Empirically confirmed against the shipped binary: (a) a mixed shared/exclusive reborrow chain exports `access=exclusive` when the body only grants shared access (CR-01, see adjudication #1); (b) a borrow-derived return with no declared origin exports with no `public_origin` field at all, indistinguishable from an owned value (D-03-02, see adjudication #3) |
| 4 | Separate compilation rejects stale, omitted, or impossible public-origin summaries while retaining the coordinated-frontend-lie limitation explicitly | ✗ FAILED | "Stale" is genuinely rejected (`TestStaleSummaryRejectedBeforeOtherChecks`, in the standing suite). "Omitted" is not rejected — confirmed live against the shipped binary (adjudication #3). "Impossible" is only partially caught — the CR-01 bug means a genuinely impossible declaration (mismatched access on a mixed-access chain) is not detected (adjudication #1). The coordinated-frontend-lie limitation IS retained and named (`escape:coordinated-frontend-summary-lie`, surfaced in every `verify` run, never reported as a detected control) |

**Score:** 2/4 truths verified (0 present, behavior-unverified)

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/compiler/check/check.go` (loanLivenessFixpoint, materializeLoanEndpoints, discoverLoanLastUses) | Edge-specific loan liveness | ✓ VERIFIED (behavior), ⚠️ architecturally split (D-03-01) | Present, wired, behaviorally correct per adjudication #2; the CFG dataflow's output is not consumed by admission logic |
| `internal/compiler/pathoracle/pathoracle.go` | Independent bounded path-oracle re-derivation | ✓ VERIFIED | Package exists, imports independently, `MaxPaths` cap enforced, lane passes in gate |
| `internal/compiler/corevalidate/corevalidate.go` (recomputeLoanEndpoints) | Independent re-derivation via reachability closure | ✓ VERIFIED | Confirmed present; `control:core.loan_endpoint_mismatch` lane passes |
| `internal/compiler/originvalidate/originvalidate.go` | Independent origin/access re-derivation | ✗ STUB-LIKE (real bug) | Exists, wired into `ValidatePublished` and the CLI, but `RecomputeOrigin`'s exclusive-access branch is unguarded (CR-01) and `ValidatePublished` never invokes recomputation for an undeclared origin (D-03-02) — both confirmed live against the shipped binary |
| `scripts/verify-phase3.sh` | Phase 3 gate requiring every required control | ✓ VERIFIED | Re-run live: exit 0, all required control IDs present, both expected escapes present and never reported as detected controls |

### Data-Flow Trace (Level 4)

`originvalidate.RecomputeOrigin`'s output flows to `ValidatePublished`'s comparison and from there to `interface export`'s CLI output — the chain is genuinely wired and produces real, non-static output that varies with input (confirmed: the two hand-written programs above produced different, input-dependent access/paths values). The defect is not a stub or a disconnected data path; it is a real computation that computes the wrong answer on an untested input class, and a real check that is unconditionally skipped on a specific input shape (no declared origin).

### Key Link Verification

| From | To | Via | Status | Details |
|------|-----|-----|--------|---------|
| `check.go` PublicOrigin construction | `originvalidate.ValidatePublished` | `session.InterfaceExportCommandFile` only | ⚠️ PARTIAL | Wired for `interface export`; confirmed NOT wired for `lang check`/`lang run` (`session.Check`, `RunInterpreter`, `RunNative` never call `ValidatePublished` — grepped and confirmed; matches 03-REVIEW.md WR-01, not independently reproduced with a runtime consumer since none exists yet, consistent with the review's own assessment) |
| `loanLivenessFixpoint` (`checkBranch`) | Admission decision (`conflictingLoan`/`expiringLoans`) | none | ✗ NOT_WIRED | Confirmed: `linear.LoanEndpoints` is written once (`check.go:352`) and never read by any admission-deciding code (D-03-01, independently re-confirmed by reading `check.go`) |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Shared/exclusive conflict matrix accept/reject | `lang check testdata/phase3/branch_one_arm_shared_{accept,reject}.lang` | accept exits 0; reject exits 2 with `ownership.move_while_borrowed` | ✓ PASS |
| Mixed-access reborrow chain access mode | `lang interface export` on a hand-written `borrow mut(x)` declaration over a body whose only return is a shared reborrow | reports `access=exclusive` (should be `shared`, or rejected) | ✗ FAIL |
| Undeclared origin on a borrow-derived exported return | `lang interface export` on a hand-written exported function returning a borrow-derived value with no origin annotation | exit 0, `public_origin` field absent (should reject or flag) | ✗ FAIL |
| Phase 3 gate | `sh scripts/verify-phase3.sh` | exit 0, all 7 required controls + 2 expected escapes present | ✓ PASS |

### Probe Execution

Not applicable — this phase has no `scripts/*/tests/probe-*.sh` files; `scripts/verify-phase3.sh` is the phase's own gate script and was run directly above (Behavioral Spot-Checks table).

### Requirements Coverage

| Requirement | Source Plans | Description | Status | Evidence |
|-------------|--------------|--------------|--------|----------|
| OWN-03 | 03-01, 03-02, 03-03, 03-04, 03-05, 03-07 | Shared and exclusive loans obey conflict rules and ordinary local loans end at proven CFG point/edge-specific last use | ✓ SATISFIED | SC1, SC2 both verified (see truths table). D-03-01's architectural gap is real but does not falsify the observable truth — independently re-derived and confirmed above, not accepted on the summary's word alone |
| OWN-04 | 03-06, 03-07 | Public borrowed results record verified field/alternative origins and access mode without inspecting provider bodies downstream | ✗ BLOCKED | SC3, SC4 both failed (see truths table). CR-01 and D-03-02 are both live, reachable, empirically reproduced defects in the one mechanism (`originvalidate`) that exists to satisfy this requirement |

All requirement IDs declared across the seven plans (`OWN-03`: 03-01/02/03/04/05/07; `OWN-04`: 03-06/07) are accounted for; no orphaned requirement IDs found in `REQUIREMENTS.md`'s Phase 3 row beyond these two. `REQUIREMENTS.md` correctly still shows both as unchecked/"Pending" — that write belongs to the orchestrator's `phase.complete` step, not this report. Per this verification, **OWN-03 is earned; OWN-04 is not**.

### Anti-Patterns Found

No `TODO`/`FIXME`/`XXX`/`TBD`/`HACK`/`PLACEHOLDER` markers found in `internal/compiler/originvalidate/originvalidate.go`, `internal/compiler/check/check.go`, or `internal/compiler/session/session.go`. No debt-marker gate violation. The defects found are real logic bugs and a scoping gap, not stub/placeholder code — both are already honestly recorded in `03-REVIEW.md` (CR-01) and `03-DEBT.md` (D-03-01, D-03-02), which this report independently re-verified against the codebase rather than trusting.

## Deferred Items

None. Neither CR-01 nor D-03-02 is named or clearly covered by any later phase's goal or success criteria in `.planning/ROADMAP.md` (Phase 4 covers fallible resources and the C boundary; nothing there addresses origin/access verification). They are Phase 3's own unresolved debt against Phase 3's own success criteria, not intentionally deferred work.

## Non-Regression

- `go build ./...`, `go vet ./...`, `go test ./...`, `go test -race ./...` all pass (orchestrator-provided, not independently re-run in full given cost; `sh scripts/verify-phase3.sh` — which itself runs `go test ./...`, `go test -race ./...`, and `go vet ./...` — was re-run live above and passed).
- `sh scripts/verify-phase3.sh` re-run live: exit 0.
- Phase 1 and Phase 2 golden/control non-regression is asserted by the same script and was observed passing in this run.

## Gaps Summary

Two of four ROADMAP success criteria for this phase fail on live, empirically reproduced evidence, both rooted in `internal/compiler/originvalidate/originvalidate.go`:

1. **CR-01** — `RecomputeOrigin`'s exclusive-access branch is unguarded, letting a mixed shared/exclusive reborrow chain report a stronger access mode (`exclusive`) than the body actually grants (`shared`). Confirmed live: `lang interface export` accepts and misreports such a function today.
2. **D-03-02** — `ValidatePublished` never checks whether an origin *should* have been declared; a borrow-derived exported return with no origin annotation exports identically to a fully-owned value. Confirmed live: `lang interface export` on such a function exits 0 with no `public_origin` field.

Both were already honestly surfaced by this phase's own process (03-REVIEW.md and 03-DEBT.md, produced by 03-05's mid-phase gate and 03-07's code review) — this verification's contribution is confirming both are real and reachable through the shipped binary (not merely theoretical), and that they are sufficient to fail OWN-04's two success criteria as literally worded in ROADMAP.md, notwithstanding that neither is an executable-unsoundness today (this language has no cross-function call construct yet, so nothing currently consumes a misleading exported summary at runtime).

OWN-03 (SC1, SC2) is earned. D-03-01 is real architectural debt (the CFG dataflow's linear cost never reaches the admission-deciding code path) but does not falsify OWN-03's observable truth, which is independently proven correct by a different, per-arm mechanism.

**Recommended remediation before OWN-04 is marked complete:**
1. Guard `RecomputeOrigin`'s `OpBorrowExclusive` case exactly like `OpBorrowShared` (first-seen wins), plus a regression fixture for a mixed-access reborrow chain.
2. Add an admission-time or `ValidatePublished`-time check that runs `RecomputeOrigin` unconditionally for exported functions and rejects (or requires an explicit opt-out for) a real, undeclared origin.

---

_Verified: 2026-09-04_
_Verifier: Claude (gsd-verifier)_
