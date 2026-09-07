---
phase: 03-borrowed-views-and-cfg-lifetimes
verified: 2026-09-04T18:00:00Z
status: passed
score: 4/4 must-haves verified
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 3/4
  gaps_closed:
    - "SC3/SC4 multi-arm origin leak (03-VERIFICATION.md's third-round finding): RecomputeOriginPerReturn now walks every core.OpReturn in a function's Linear.Operations, not just the first; RecomputeOrigin is a conservative combiner over the per-return results. Independently reproduced against the shipped binary with a hand-written, out-of-corpus program (module owned.verifier_probe_alpha, first arm owned/second arm shared-borrow): lang interface export now exits 2 with core.origin_omitted instead of exiting 0 with a public_origin-less summary."
    - "Access-mode disagreement across arms (new must-have added by 03-10): a hand-written, out-of-corpus program (module owned.verifier_probe_beta, one arm shared-borrow/one arm exclusive-borrow) recomputes to the AccessConflicting sentinel and is refused with core.origin_omitted (undeclared) — verified live. The declared-and-conflicting path is proven refused by TestDeclaredConflictingAccessIsRefused (re-run, passes) since match-bodied functions cannot declare an origin through the current frontend grammar; the sentinel's containment (it can never reach a written summary) is independently traced in originvalidate.go and confirmed by the same test."
  gaps_remaining: []
  regressions: []
gaps: []
deferred: []
---

# Phase 3: Borrowed Views and CFG Lifetimes Verification Report

**Phase Goal:** Local borrows remain ergonomic through precise last-use inference while public borrowed results remain explicit and separately checkable.
**Verified:** 2026-09-04
**Status:** passed
**Re-verification:** Yes — after gap closure (03-10, following two prior gap-closure rounds: 03-08, 03-09)

## Re-Verification Summary

This is the third re-verification of Phase 3. The prior round (03-VERIFICATION.md, `status: gaps_found`, score 3/4) found that both originally-reported gaps (GAP 1/CR-01, mixed-access reborrow-chain misreport; GAP 2/D-03-02, undeclared origin exporting as fully owned) were genuinely closed for single-arm, straight-line functions, but a post-closure code review (03-REVIEW.md, `9ab2e94`) found — and that verification independently reproduced — a new instance of the same defect class: `originvalidate.RecomputeOrigin` only inspected the FIRST `core.OpReturn` in a function's flat `Linear.Operations`, so a match-arm-bodied function whose non-first arm returned a live borrow was exported with no `public_origin` field at all, defeating SC3 and SC4 for that shape.

Gap-closure plan 03-10 was executed to close this. Per this verification's standing instruction not to trust either the SUMMARY's or the review's closure claims at face value, I re-read the production code directly and independently reproduced every required probe against the freshly-built shipped binary, using hand-written, out-of-corpus programs distinct from every fixture in `testdata/phase3/` and from the prior verification's own reproduction program.

### Independent code reading

`internal/compiler/originvalidate/originvalidate.go` now defines `RecomputeOriginPerReturn` as the package's sole backward-walk site (confirmed: `grep -c 'for current != function.Parameter.ID'` → `1`). It collects every `core.OpReturn` in `function.Linear.Operations` into `returnOps` (no `if returnOp == nil` first-only guard remains — that exact line from the prior verification's cited defect is gone) and walks backward from each independently, using one `sourceOf` map built once over the whole flat operations list. `RecomputeOrigin` is now a pure combiner: no derived return → not-ok; all derived returns agree on access → that access; disagreement → the `AccessConflicting` sentinel (never a silently-picked winner). `ValidatePublished` gained a declared-access domain check that refuses any declared `Access` outside `{"shared","exclusive"}` before ever comparing against a recomputed answer, so a declared summary can never claim the sentinel and have it match a genuinely conflicting recomputation.

### Independent live reproduction (hand-written, out-of-corpus, this verification's own programs)

Built the shipped binary fresh (`go build -o /tmp/lang ./cmd/lang`).

**Probe 1 — multi-arm omitted-origin leak** (`module owned.verifier_probe_alpha`, first arm `take`s and returns owned, second arm `borrow`s and returns the loan, no declared origin — the exact shape the prior verification falsified SC3/SC4 with, written independently with different naming):
- `lang --json check` → `"status":"pass"`, exit 0.
- `lang --json interface export` → `"status":"invalid"`, **exit 2**, diagnostic `core.origin_omitted`: `"no declared origin, but body derives origin [pulse] with access \"shared\""`. No summary file written.
- This is the exact reversal of the prior verification's falsifying result (which was exit 0 with a `public_origin`-less summary on an analogous program). **The gap is closed.**

**Probe 2 — access-mode disagreement across arms** (`module owned.verifier_probe_beta`, one arm returns a shared reborrow, the other an exclusive reborrow, no declared origin):
- `lang --json check` → `"status":"pass"`, exit 0.
- `lang --json interface export` → `"status":"invalid"`, **exit 2**, diagnostic `core.origin_omitted`: `"...body derives origin [pulse] with access \"conflicting\""` — confirming the combiner produced the `AccessConflicting` sentinel rather than silently picking `shared` or `exclusive`, and that the sentinel correctly routes through the same omitted-origin refusal path since match-bodied functions cannot declare an origin in the current grammar. The declared-and-conflicting path (a hand-mutated summary declaring the sentinel itself) is proven refused by `TestDeclaredConflictingAccessIsRefused`, re-run live and passing — it specifically targets the one fixture whose own recomputed answer IS the sentinel, which would slip through if the domain check ran after rather than before the comparison.

**Non-regression of the two previously-closed single-arm gaps** — re-run live against the shipped binary:
- `testdata/phase3/public_view_mixed_access.lang` → exit 2, `core.origin_access_mismatch`. Unchanged.
- `testdata/phase3/public_view_omitted.lang` → exit 2, `core.origin_omitted`. Unchanged.

**SC1 and SC2** re-confirmed, not merely assumed carried-forward:
- SC1: `control:cfg.path_oracle_disagreement` re-run live via the gate (below), passes; `pathoracle` package still imports neither `check` nor `ast` (unchanged mechanism from prior verifications).
- SC2: `branch_one_arm_shared_accept.lang` checks clean (exit 0); `branch_one_arm_shared_reject.lang` rejected with `ownership.move_while_borrowed`. Both re-run live against the shipped binary.

**The gate** — `sh scripts/verify-phase3.sh` re-run live from the repo root: exit 0. Confirmed eleven required `control:` IDs in the script's own required-control loop (`ownership.exclusive_conflict`, `ownership.exclusive_move`, `core.loan_endpoint_mismatch`, `cfg.path_oracle_disagreement`, `origin.understated_summary`, `origin.impossible_summary`, `origin.stale_summary`, `origin.omitted_summary`, `origin.mixed_access_chain`, `origin.multi_arm_omitted`, `origin.multi_arm_access_conflict` — up from nine), all present in the `lane:borrowed-origin-controls` and sibling lanes' JSON output. Both expected escapes (`escape:coordinated-source-core-lie`, `escape:coordinated-frontend-summary-lie`) present in `expected_escapes`, neither widened nor appearing as a detected `control:escape:...` entry anywhere in the gate's JSON output.

**Sentinel containment** — traced independently in the code (not merely per the review's claim): `checkBranch` (the sole producer of match-bodied `core.Function` values, per `03-REVIEW.md`'s own trace, which this verification does not re-derive but whose conclusion is consistent with the domain-check code read directly) never sets `PublicOrigin`, so a match function can never carry a declared sentinel from honest source. `ValidatePublished`'s domain check independently refuses any declared `Access` outside `{"shared","exclusive"}` before the declared-vs-recomputed comparison runs — read directly at `originvalidate.go:227-237`, confirmed to execute unconditionally inside the `function.PublicOrigin != nil` branch, before `containsAll`/access-mismatch comparisons. `TestDeclaredConflictingAccessIsRefused` re-run live, passes, and specifically exercises the one fixture whose recomputed answer IS the sentinel (the case that would slip through a mis-ordered check).

**Full test suite**: `go build ./...`, `go vet ./...` clean. `go test ./...` — 14 packages, all pass, 0 failures (re-run live, not merely cited from context). Named tests re-run individually: `TestMultiArmOmittedOriginRejected`, `TestOwnedArmsStillPublish`, `TestMultiArmAccessConflictDerivesNeitherArm`, `TestDeclaredConflictingAccessIsRefused`, `TestOriginAccessMismatchRejected`, `TestOmittedOriginRejected` — all PASS.

**No debt markers** (`TODO`/`FIXME`/`XXX`/`TBD`/`HACK`/`PLACEHOLDER`) found in any file this round's plan modified.

## Goal Achievement

### Observable Truths (ROADMAP Success Criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Shared and exclusive borrow conflicts are accepted or rejected identically by the core checker and bounded path oracle | ✓ VERIFIED | `pathoracle` package (imports neither `check` nor `ast`); `control:cfg.path_oracle_disagreement` re-run live via the gate, exit 0; unchanged and re-confirmed across all three verification rounds |
| 2 | A branch-specific last use ends a loan on the correct CFG edge without a manual scope block; omitting that edge is detected | ✓ VERIFIED | Re-confirmed live against the shipped binary: `branch_one_arm_shared_accept.lang` checks clean; `branch_one_arm_shared_reject.lang` rejected with `ownership.move_while_borrowed`. D-03-01 (architectural: mechanism is `discoverLoanLastUses` per-arm, not the CFG dataflow) carried, non-blocking, unchanged |
| 3 | A public borrowed view names all verified field/alternative origins and access modes without downstream body inspection | ✓ VERIFIED | All three known instances of the defect class closed and independently re-verified live: single-arm omission (GAP 2/D-03-02), single-arm mixed-access chain (GAP 1/CR-01), and multi-arm omission/access-conflict (this round's 03-10 fix). Probes 1 and 2 above, run against hand-written out-of-corpus programs distinct from every fixture, both correctly reject |
| 4 | Separate compilation rejects stale, omitted, or impossible public-origin summaries while retaining the coordinated-frontend-lie limitation explicitly | ✓ VERIFIED | "Stale" rejected (`TestStaleSummaryRejectedBeforeOtherChecks`, unchanged). "Omitted" and "impossible" rejected for both single-arm and multi-arm shapes — verified live in this round via Probes 1 and 2 plus non-regression re-runs of `public_view_omitted.lang`/`public_view_mixed_access.lang`. Coordinated-frontend-lie limitation retained and named (`escape:coordinated-frontend-summary-lie`), present in the gate's `expected_escapes`, never promoted to a detected control |

**Score:** 4/4 truths verified (0 present, behavior-unverified)

### Deferred Items

None.

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/compiler/check/check.go` (loanLivenessFixpoint, materializeLoanEndpoints, discoverLoanLastUses) | Edge-specific loan liveness | ✓ VERIFIED (behavior), ⚠️ architecturally split (D-03-01, carried, non-blocking) | Unchanged from prior verifications |
| `internal/compiler/pathoracle/pathoracle.go` | Independent bounded path-oracle re-derivation | ✓ VERIFIED | Unchanged from prior verifications |
| `internal/compiler/corevalidate/corevalidate.go` (recomputeLoanEndpoints) | Independent re-derivation via reachability closure | ✓ VERIFIED | Unchanged from prior verifications |
| `internal/compiler/originvalidate/originvalidate.go` | Independent origin/access re-derivation, sound across single-arm and multi-arm shapes | ✓ VERIFIED | `RecomputeOriginPerReturn` walks every `OpReturn`; `RecomputeOrigin` combines conservatively; `ValidatePublished`'s domain check contains the `AccessConflicting` sentinel. Read directly and probed live; matches SUMMARY's claims |
| `scripts/verify-phase3.sh` | Phase 3 gate requiring every required control | ✓ VERIFIED | Re-run live: exit 0, eleven required control IDs present (up from nine), both expected escapes present, none promoted to detected controls |

### Data-Flow Trace (Level 4)

`originvalidate.RecomputeOriginPerReturn`'s per-return output flows into `RecomputeOrigin`'s combiner, whose output flows into `ValidatePublished`'s comparison, and from there into `interface export`'s CLI exit code and (on success only) the written summary file. Confirmed genuinely wired and input-dependent: the four hand-written/fixture programs exercised this round (Probe 1 multi-arm-omitted, Probe 2 access-conflict, `public_view_mixed_access.lang`, `public_view_omitted.lang`) produced four distinct, input-dependent outcomes, all correctly rejecting. No stub, no disconnected path, no static fallback.

### Key Link Verification

| From | To | Via | Status | Details |
|------|-----|-----|--------|---------|
| `check.go` PublicOrigin construction | `originvalidate.ValidatePublished` | `session.InterfaceExportCommandFile` only | ⚠️ PARTIAL | Unchanged (WR-01, carried, non-blocking): not wired into `lang check`/`lang run`. Does not affect SC3/SC4's own truth — `interface export` is the publication path these criteria govern |
| `RecomputeOriginPerReturn`'s per-return backward walk | Every `OpReturn` in a function | function iteration | ✓ WIRED | Read directly: `returnOps` collects every `core.OpReturn` encountered while scanning `operations` in index order, no first-only guard remains; each is walked independently via `walkReturnOrigin` sharing one `sourceOf` map. Probed live via Probes 1 and 2 |
| `loanLivenessFixpoint` (`checkBranch`) | Admission decision (`conflictingLoan`/`expiringLoans`) | none | ✗ NOT_WIRED | Unchanged from prior verifications (D-03-01, carried, non-blocking) |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Multi-arm omitted origin (this round's target defect, hand-written, out-of-corpus) | `lang interface export` on `owned.verifier_probe_alpha` (first arm owned, second arm shared borrow) | exit 2, `core.origin_omitted`, no summary file written | ✓ PASS |
| Multi-arm access-mode disagreement (hand-written, out-of-corpus) | `lang interface export` on `owned.verifier_probe_beta` (one arm shared, one arm exclusive) | exit 2, `core.origin_omitted` naming access `"conflicting"` — sentinel correctly derived, never silently resolved to one arm | ✓ PASS |
| Mixed-access reborrow chain, single-arm (GAP 1/CR-01, closed, non-regression) | `lang interface export public_view_mixed_access.lang` | exit 2, `core.origin_access_mismatch` | ✓ PASS |
| Omitted origin, single-arm (GAP 2/D-03-02, closed, non-regression) | `lang interface export public_view_omitted.lang` | exit 2, `core.origin_omitted` | ✓ PASS |
| Branch-specific last use accepted/rejected (SC2, non-regression) | `lang check` on `branch_one_arm_shared_accept.lang` / `branch_one_arm_shared_reject.lang` | exit 0 clean / rejected `ownership.move_while_borrowed` | ✓ PASS |
| Sentinel containment via declared-access domain check | `go test ./internal/compiler/originvalidate/... -run TestDeclaredConflictingAccessIsRefused -v` | PASS | ✓ PASS |
| Phase 3 gate | `sh scripts/verify-phase3.sh` | exit 0, eleven required controls + two expected escapes present, none promoted | ✓ PASS |

### Probe Execution

Not applicable — no `scripts/*/tests/probe-*.sh` files exist for this phase; `scripts/verify-phase3.sh` is the phase's own gate script and was run directly above, live, exit 0.

### Requirements Coverage

| Requirement | Source Plans | Description | Status | Evidence |
|-------------|--------------|--------------|--------|----------|
| OWN-03 | 03-01, 03-02, 03-03, 03-04, 03-05, 03-07 | Shared and exclusive loans obey conflict rules and ordinary local loans end at proven CFG point/edge-specific last use | ✓ SATISFIED | SC1, SC2 re-confirmed live this round. `.planning/REQUIREMENTS.md` currently shows `[ ]`/"Pending" — this row should be updated to Complete |
| OWN-04 | 03-06, 03-07, 03-08, 03-09, 03-10 | Public borrowed results record verified field/alternative origins and access mode without inspecting provider bodies downstream | ✓ SATISFIED | SC3, SC4 now hold across single-arm and multi-arm shapes — independently re-verified live this round (Probes 1 and 2), not merely re-read from the SUMMARY. `.planning/REQUIREMENTS.md` currently shows `[ ]`/"Gaps Found" — this row should be updated to Complete now that this verification confirms the gap is closed |

All requirement IDs declared across the ten plans (`OWN-03`: 03-01/02/03/04/05/07; `OWN-04`: 03-06/07/08/09/10) are accounted for; no orphaned requirement IDs found in `REQUIREMENTS.md`'s Phase 3 row beyond these two.

**Recommended REQUIREMENTS.md update (this verification's decision, per 03-10-SUMMARY.md's explicit deferral):**
- `OWN-03`: `[ ]`/"Pending" → `[x]`/"Complete"
- `OWN-04`: `[ ]`/"Gaps Found" → `[x]`/"Complete"
- Phase 3 row in ROADMAP.md's own checklist: mark complete.

### Anti-Patterns Found

No `TODO`/`FIXME`/`XXX`/`TBD`/`HACK`/`PLACEHOLDER` markers found in any file this round's plan (03-10) modified. No stub/placeholder patterns found — every code path traced this round performs real, input-dependent computation.

## Non-Regression

- `go build ./...`, `go vet ./...` pass, re-run live.
- `go test ./...` — 14 packages, 0 failures, re-run live.
- Named tests re-run individually and pass: `TestMultiArmOmittedOriginRejected`, `TestOwnedArmsStillPublish`, `TestMultiArmAccessConflictDerivesNeitherArm`, `TestDeclaredConflictingAccessIsRefused`, `TestOriginAccessMismatchRejected`, `TestOmittedOriginRejected`.
- `sh scripts/verify-phase3.sh` re-run live: exit 0, eleven required controls present, two expected escapes present.
- `lang check`/`lang interface export` re-run directly against a freshly-built shipped binary for two new hand-written, out-of-corpus probe programs (this verification's own, distinct from every prior verification's and every fixture's) plus the two previously-closed-gap fixtures and both SC2 branch fixtures.
- `03-DEBT.md` and `03-REVIEW.md` read directly: D-03-01 and WR-01 remain open, accepted, non-blocking, unedited by this round; IN-01 (new, informational only, not a defect) is carried forward, non-blocking.

## Gaps Summary

None. All four ROADMAP Success Criteria for Phase 3 hold, independently re-verified live against the shipped binary rather than trusted from SUMMARY.md or 03-REVIEW.md claims. The multi-arm origin-leak defect that caused the second re-verification to fail SC3/SC4 is closed: `RecomputeOriginPerReturn` walks every `core.OpReturn` in a function's body (not just the first), `RecomputeOrigin` combines per-arm results conservatively with a named sentinel for disagreement rather than a silently-picked winner, and the sentinel is proven unable to reach a written interface summary or round-trip as a declared value.

**Carried, accepted, non-blocking debt (unchanged, does not affect this phase's observable truths):**
- **D-03-01** — the CFG dataflow's linear cost never reaches the admission-deciding code path (OWN-03's architectural debt). Open and accepted.
- **WR-01** — `originvalidate.ValidatePublished` is reached only via `session.InterfaceExportCommandFile`, not wired into `lang check`/`lang run`.
- **IN-01** (new, informational, from `03-REVIEW.md`) — the shared `sourceOf`-map safety proof depends on an invariant enforced in `check.go` (monotonic place/operation IDs never reset per arm), outside `originvalidate`'s control and untested locally. Not a defect in current code.

---

_Verified: 2026-09-04_
_Verifier: Claude (gsd-verifier)_
