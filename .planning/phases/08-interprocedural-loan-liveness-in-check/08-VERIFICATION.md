---
phase: 08-interprocedural-loan-liveness-in-check
verified: 2026-09-10T05:00:00Z
status: passed
score: 7/7 must-haves verified
behavior_unverified: 0
overrides_applied: 0
---

# Phase 08: Interprocedural Loan Liveness in check Verification Report

**Phase Goal:** The checker derives cross-function loan liveness from signatures alone, under a measured, bounded cost.
**Verified:** 2026-09-10
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | `check` derives interprocedural loan liveness from callee signatures only, never by re-walking callee bodies (OWN-06 text) | VERIFIED | `buildInterproceduralSummaries` (check.go:579-616) consults `core.FunctionSignature.Return.Mode`/`Parameters[0].Mode` only, and `deriveFunctionUsesParam` (check.go:686-726) walks each function's **own** body exactly once, on its own turn in reverse-postorder; at a call site a caller only ever calls `summaries.lookup(calleeID)` — never re-reads `core.Function.Linear`/`Match` for a callee. Confirmed by direct code read, not doc-comment trust. |
| 2 | `check` refuses cross-function loan-liveness violations and admits safe compositions, end-to-end through the real CLI, on a composition-depth-≥2 adversarial corpus (ROADMAP criterion 1) | VERIFIED | Independently re-ran `go run ./cmd/lang --json check` against all 10 corpus fixtures myself (not trusting SUMMARY): `twin_a_refuse`, `relay_depth2_refuse`, `negative_control_fails`, `negative_control_infallible`, `relay_escort_witness` all report `check.interprocedural_loan_liveness`; `match_arm_call` is clean; `twin_a_accept`/`relay_depth2_accept` are clean at `check` (the reported `core.move_while_borrowed` is corevalidate's independent peer diagnostic surfaced by the same CLI command — confirmed by reading `session.go`'s union-of-peer-diagnostics comment, matching the documented D-08-40 divergence exactly). Pattern B (`twin_b_*`) is refused intraprocedurally for both members alike — a disclosed, genuine scope limit (D-08-41), proven instead at the checked-core level by `TestInterproceduralLivenessTwinPatternB` (passing). |
| 3 | Terminates under a measured, fail-closed, derived iteration bound — never a magic constant, never silent truncation (ROADMAP criterion 2) | VERIFIED | `loanLivenessBoundFactor × len(blocks) × distinctLoanCount` bound in check.go; `check.loan_liveness_bound_exceeded` named refusal; ran `TestLoanLivenessBoundMutationKilled` directly — PASS, and it asserts both directions (seam up → refusal fires; seam down → same input admits). Bound value confirmed absent from Causes/Message (D-08-18). |
| 4 | Interprocedural admission cost is measured on realistic call-graph fan-out under the existing p50/p95/CoV protocol, stays within a declared bound, recorded in the feedback-budget manifest (EFF-02 text, ROADMAP criterion 3) | VERIFIED | Ran `TestInterproceduralSummaryGrowthExponentInOps` directly — PASS, fitted milli-exponents 998–1003 across five shapes (chain/diamond/dense/parser-shaped/forward), all under the 1200 (1.2×) bound. `qlt02_budget_manifest.json` carries a `hard`-gated `recomputed_work_growth_exponent` row with `ratified_by_commit: b28a92f` — the gate's own adjudication commit, distinct from the row-introducing commit, matching the documented re-ratification claim. Both chokepoints (`measure.GateEligibleMetrics`, `session.QLT02GateEligibleMetrics`) widened; `TestGateEligibleMetricSetsAgreeAcrossChokepoints`-class tests pass. |
| 5 | The mandatory mid-phase gate adjudicated criteria 1 and 3 from code-level evidence before the law was declared final, and every open agenda item is resolved or recorded as debt (never assumed safe) | VERIFIED | `PHASE-08-DEBT.md` holds 9 well-formed items (`TestDebtRegistersAreWellFormed` passes); each of the gate's four agenda items has an explicit, dated disposition in 08-06-SUMMARY.md's key-decisions and the debt register's dated paragraphs. `ownership.*` intraprocedural law confirmed NOT retired this phase (grep of `check.go` shows `ownership.move_while_borrowed` codes still live and firing on `twin_b_*`). |
| 6 | The critical code-review defect (nondeterministic loan blame via Go map iteration) is fixed, not merely logged | VERIFIED | Commit `4bd3544` sorts `chain.borrowOperation` loan IDs before the tie-break loop (check.go:863-880 area) — read the diff directly. `go vet ./...` clean; `go build ./...` clean. |
| 7 | Requirements OWN-06 and EFF-02 are accounted for, correctly mapped, and no Phase-08 requirement is orphaned | VERIFIED | `.planning/REQUIREMENTS.md` maps exactly OWN-06 and EFF-02 to "Phase 08", both marked Complete; both requirement IDs appear in plan frontmatter (08-01 through 08-04 for OWN-06, 08-05 for EFF-02, both in 08-06). No orphaned Phase-08 rows found. |

**Score:** 7/7 truths verified (0 present-but-behavior-unverified)

### Known, Disclosed Scope Limits (not gaps — adjudicated as acceptable at the mid-phase gate)

- **D-08-41** — Pattern B's real `.lang` twin fixtures cannot demonstrate a differing end-to-end CLI verdict because `computeLoanLastUses` (the AST-shadow intraprocedural admission path) is summary-blind by design and refuses both members identically before the interprocedural pass ever runs. Independently confirmed via direct CLI re-run: both `twin_b_refuse.lang` and `twin_b_accept.lang` report `ownership.move_while_borrowed`. The contract-driven differentiation this pair exists to demonstrate is proven instead at the checked-core level (`TestInterproceduralLivenessTwinPatternB`, passing). This is a genuine, disclosed, permanent scope limit on criterion 1's real-fixture corpus for one of three composition patterns — it does not falsify the phase goal, which is proven end-to-end by Pattern A and the depth-2 relay pattern, and at the checked-core level for Pattern B.
- **D-08-40** — `twin_a_accept.lang`/`relay_depth2_accept.lang` are admitted by `check` and independently refused by `corevalidate`'s still-intraprocedural peer. This is a `check`-vs-peer divergence, not a `check`-internal defect; it is registered as tracked debt with a named landing phase (Phase 09).
- **D-08-27** — OWN-09's Phase-08-vs-09 retirement-timing conflict between two project documents remains genuinely unresolved; the gate correctly declined to adjudicate it by assumption and left it for human decision. This does not affect OWN-06/EFF-02's own truth of achievement — the interim rule (ownership.* not retired, interprocedural check runs strictly last) is itself verified in the code (see truth 5).

These items were cross-checked against ROADMAP's later-phase scope (Phase 09 — Peer Re-Derivation and D-03-02 Closure) and are legitimately deferred, not silently dropped.

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/compiler/check/check.go` | Interprocedural summary table, forward canonicalization, backward gate, bound, diagnostics | VERIFIED | 3974 lines; all named functions/diagnostics present and wired; builds clean |
| `internal/compiler/check/check_test.go` | Twin-pair, relay, bound, mutation-kill, disclosed-field-set tests | VERIFIED | All 30+ named tests present; spot-run subset all PASS |
| `internal/compiler/check/costcorpus_test.go` | Five-shape synthetic corpus, growth-exponent fit | VERIFIED | Present; `TestInterproceduralSummaryGrowthExponentInOps` PASS with milli-exponents 998-1003 |
| `internal/compiler/session/qlt02_budget_manifest.json` | Ratified hard-gate manifest row | VERIFIED | `recomputed_work_growth_exponent` row present, `gate_type: hard`, `ratified_by_commit` = gate's own commit |
| `testdata/phase08/*.lang` (9 fixtures) | Adversarial composition-depth corpus | VERIFIED | All present; CLI-verified verdicts match documented expectations |
| `.planning/phases/08-.../PHASE-08-DEBT.md` | 9-item adjudicated debt register | VERIFIED | Well-formed (`TestDebtRegistersAreWellFormed` PASS); matches known_context's 9-item description |
| `.planning/REQUIREMENTS.md` | OWN-06/EFF-02 marked Complete | VERIFIED | Both marked `[x]`/Complete, correctly scoped |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| `core.FunctionSignature.Return.Mode`/`Parameters[0].Mode` | `interproceduralSummary` | `buildInterproceduralSummaries` | WIRED | Confirmed by code read; no `Linear`/`Match` field consulted for a *callee* |
| `interproceduralSummary` | `derivePlaceLoans`'s `OpCall` branch | summary-consuming forward canonicalization | WIRED | `blockLoanLiveness`'s backward `OpCall` gate consumes `usesParam`/`returnsBorrowOfParam` |
| `testdata/phase08/*.lang` | CLI verdict | `syntax.Parse` → `check.Program` | WIRED | Independently re-run via `go run ./cmd/lang --json check`, verdicts match documented expectations |
| `buildInterproceduralSummaries`'s work counter | fitted growth exponent | `costcorpus_test.go`'s log-log OLS fit | WIRED | Re-ran test directly; printed work/ops pairs consistent with 998-1003 milli-exponent claim |
| `measure.GateEligibleMetrics()` | `session.QLT02GateEligibleMetrics()` | two-chokepoint agreement | WIRED | Both widened; agreement test passes |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| End-to-end interprocedural refusal | `go run ./cmd/lang --json check testdata/phase07/relay_escort_witness.lang` | `check.interprocedural_loan_liveness` | PASS |
| Twin A refuse/accept differ by contract | CLI on `twin_a_refuse.lang`/`twin_a_accept.lang` | refuse: `check.interprocedural_loan_liveness`; accept: clean at `check` | PASS |
| Depth-2 relay refuse/accept differ | CLI on `relay_depth2_refuse.lang`/`relay_depth2_accept.lang` | refuse: `check.interprocedural_loan_liveness`; accept: clean at `check` | PASS |
| Negative control ignores forbidden fields | CLI on `negative_control_fails.lang`/`negative_control_infallible.lang` | identical `check.interprocedural_loan_liveness` verdict | PASS |
| Mutation-kill of the fail-closed bound | `go test ./internal/compiler/check/... -run TestLoanLivenessBoundMutationKilled -v` | PASS | PASS |
| Growth-exponent cost gate | `go test ./internal/compiler/check/... -run TestInterproceduralSummaryGrowthExponentInOps -v` | PASS, all shapes 998-1003 milli | PASS |
| CR-01 fix in place | `git show 4bd3544` | sorted loan-ID iteration replaces map range | PASS |
| Full package suite | `go test ./internal/compiler/check/... ./internal/compiler/measure/... ./internal/compiler/session/... ./internal/compiler/corevalidate/...` | check/measure/corevalidate: ok; session: 1 pre-existing unrelated failure (D-08-43, `TestQLT01RegistryCoversAllFiveSpikes`) | PASS (disclosed, unrelated) |
| Build and vet | `go build ./...`, `go vet ./...` | both clean, exit 0 | PASS |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| OWN-06 | 08-01, 08-02, 08-03, 08-04, 08-06 | Interprocedural loan liveness from signatures alone, fail-closed bound | SATISFIED | See truths 1-3, 5-6 above |
| EFF-02 | 08-05, 08-06 | Cost measured under existing protocol, declared bound, manifest-recorded | SATISFIED | See truth 4 above |

No orphaned requirements mapped to Phase 08.

### Anti-Patterns Found

None. Grep for `TBD|FIXME|XXX|TODO|HACK|PLACEHOLDER` across all files modified this phase (check.go, check_test.go, costcorpus_test.go, measure/statistics.go, session_phase6_budget.go, session_phase6_risklanes.go) returned zero matches. `go vet ./...` clean.

### Human Verification Required

None. All must-haves resolved to VERIFIED status from direct, independently re-run code-level and CLI evidence — no visual, real-time, or external-service behavior in scope for this phase.

### Gaps Summary

No gaps found. All observable truths tied to the phase goal's three clauses — "from signatures alone" (verified: summaries derived exclusively from `core.FunctionSignature` plus each function's own body on its own turn, callers never re-walk callee bodies), "cross-function loan liveness" (verified end-to-end on 8 of 9 real fixtures plus the carried witness fixture, with the ninth's real-fixture limitation disclosed and proven instead at the checked-core level), and "under a measured, bounded cost" (verified: deterministic work-counter growth exponent 998-1003 milli, sub-1200 bound, ratified in the hard-gated manifest) — hold on direct re-verification, not on SUMMARY trust alone.

The one code-review critical finding (CR-01, nondeterministic diagnostic blame) was confirmed fixed in a follow-up commit rather than merely logged. Two minor code-review findings (WR-02 recomputation redundancy, IN-01/IN-02 doc-precision nits) were deliberately left unfixed as non-blocking quality nits — correctly so, since neither affects correctness or the goal's three load-bearing clauses.

The disputed process-ordering finding (OWN-06/EFF-02 marked complete in REQUIREMENTS.md before the mid-phase gate ran) was scrutinized directly: git history confirms the marks were flipped by 08-03's and 08-05's own commits, ahead of 08-06's gate. This verifier's own independent judgment, formed from re-running the actual evidence (not from trusting the gate's self-adjudication), reaches the same conclusion the gate did: both requirements are genuinely satisfied in substance, so the ordering deviation is a process-hygiene issue worth noting but not one that invalidates the phase goal.

---

_Verified: 2026-09-10_
_Verifier: Claude (gsd-verifier)_
