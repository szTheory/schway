---
phase: 09-peer-re-derivation-and-d-03-02-closure
verified: 2026-09-10T00:00:00Z
status: passed
score: 5/5 roadmap success criteria verified; the 1 documentation-consistency gap found was RESOLVED at 893567e
behavior_unverified: 0
overrides_applied: 0
gaps:
  - truth: "OWN-08's top-section requirement checkbox in REQUIREMENTS.md is [x] (Complete), matching the traceability table and the actual code state"
    status: resolved
    resolved_at: "893567e"
    resolved_note: "Orchestrator flipped REQUIREMENTS.md:54 to [x] after this verification ran. The finding is preserved above rather than erased: the verifier caught a real requirement-list/traceability-table inconsistency, and it is now fixed. Re-verified: all seven Phase 09 checkboxes (OWN-05a, OWN-05b, OWN-07, OWN-08, OWN-09, TRU-04, QLT-07) agree with their traceability rows; OWN-05b correctly remains unticked/Pending for Phase 10."
    reason: "REQUIREMENTS.md line 54's requirement-list checkbox for OWN-08 is still unticked ('- [ ] **OWN-08**: D-03-02 is closed...'), while the traceability table at line 206 says 'OWN-08 | Phase 09 | Complete' and the code/tests independently confirm OWN-08's substance IS delivered (both check and corevalidate refuse testdata/phase07/relay_escort_witness.lang independently, verified live at HEAD). Git history shows the traceability-table row was flipped Pending->Complete but the matching requirement-list checkbox was never touched, in every commit from 09-03 through 09-10 (D-08-08fbcf182...cc76512). This is exactly the requirement-vs-code overclaim/underclaim shape the debt registers exist to catch, just in the other direction: the code and traceability table both say Complete, but the requirement list itself still visually reads as incomplete."
    artifacts:
      - path: ".planning/REQUIREMENTS.md"
        issue: "Line 54 checkbox reads '- [ ]' where it should read '- [x]', inconsistent with line 206's 'Complete' and with the verified code state"
    missing:
      - "Flip REQUIREMENTS.md line 54's checkbox to [x] so the requirement-list section and the traceability table agree (a one-line doc fix, not a code change)"
---

# Phase 09: Peer Re-Derivation and D-03-02 Closure — Verification Report

**Phase Goal:** A second, independently-implemented admission layer reaches the
same interprocedural answer as the first — and the one debt item M001 knowingly
carried forward (D-03-02) is closed in both layers.

**Verified:** 2026-09-10
**Status:** gaps_found (one documentation-consistency gap; all code-level and
test-level evidence for the phase goal itself is verified true)
**Re-verification:** No — initial verification

## Goal Achievement

### Roadmap Success Criteria (verified against the codebase directly, not against task completion)

| # | Criterion | Status | Evidence |
|---|-----------|--------|----------|
| 1 | Shadow run over recursion/diamond/deep-chain shapes shows ZERO divergence, landed in the same plan, BEFORE either peer ships | ✓ VERIFIED | `TestSyntheticShapeDifferentialHasNoUndeclaredDivergence`, `TestNoUndeclaredCheckPeerDivergenceAcrossCorpus`, and the cycle-peer differential all pass live (`go test ./internal/compiler/check ./internal/compiler/session -v`, confirmed above). **Non-vacuousness verified directly**: the 5-shape synthetic corpus (`testsupport.GenerateCallGraphCorpus`) is honestly documented in-source as containing zero borrow operations (verified by inspection — only `OpCall`/`OpCopy`/`OpReturn` in `testsupport/callgraphcorpus.go`), so that specific sweep is vacuously true and the code says so in its own doc comment. Real non-vacuous coverage comes from two other places, both confirmed live: (a) the `.lang` corpus gate (`session_peer_gate_test.go`'s `peerDivergenceExpected`) which includes hand-crafted borrow-carrying fixtures (`twin_a_accept.lang`, `relay_depth2_accept.lang`, `relay_escort_witness.lang`), and (b) `TestSyntheticShapeDifferentialMutationReintroducesDivergence`'s hand-built `mutationCandidateProgram()`, which contains a real `core.OpBorrowShared` and proves the detection machinery itself is not vacuous. |
| 2 | A seeded endpoint-level fault makes the two peers diverge, proving criterion 1's agreement is load-bearing; counted-work lane shows linear-or-bounded cost | ✓ VERIFIED | Both halves of the companion assertion exist and pass: `TestSyntheticShapeDifferentialMutationReintroducesDivergence` shows the mutation reintroduces divergence AND (via `cleanAgain := corevalidate.Validate(program)` after `defer restore()`) that disengaging the seam returns to clean — proving causation, not correlation. `corevalidate_peer_liveness_test.go`'s companion asserts `check` is unaffected by the same fault (mirrors D-09-25). Cost lane: `TestClosureCostScaling`/`PeerClosureCostUnmemoizedSeamExceedsBound` (09-06) plus the ratified `qlt02_budget_manifest.json` row (09-08-SUMMARY.md Task 2) — both green at HEAD. |
| 3 | Exported borrow-derived return with no declared origin refused INTERPROCEDURALLY by BOTH admission layers; D-03-02 closed | ✓ VERIFIED (at the layer it claims) | Live run: `go run ./cmd/lang --json check testdata/phase07/relay_escort_witness.lang` refuses with `check.interprocedural_loan_liveness` (check's own independent refusal). `TestRelayEscortWitnessCorevalidateIndependentlyRefusesMoveWhileBorrowed` (run live, PASS) proves corevalidate independently refuses the same program via `core.move_while_borrowed`, a different code, confirming genuine independent re-derivation rather than one peer forwarding the other's verdict. D-09-51's known-open gap (originvalidate's `walkReturnOrigin` has no `OpCall` case, so the CLI surfaces `core.origin_omitted` for the two accepted-program fixtures) was directly re-verified as PRE-EXISTING: checked out commit `133a731` (pre-Phase-09) in a scratch worktree and confirmed both `twin_a_accept.lang` and `relay_depth2_accept.lang` were ALREADY CLI-refused with `core.move_while_borrowed` at that commit — post-Phase-09 they are refused instead with `core.origin_omitted`, a different pre-existing defect one layer down, not a new regression introduced by this phase. D-03-02's own subject (`relay_escort_witness.lang`) is unaffected by D-09-51 since it never reaches originvalidate's ambiguous path. |
| 4 | Ownership transfer at a call site has exactly one meaning, derived independently by `check` and `corevalidate`; override not expressible | ✓ VERIFIED | `TestConventionOverrideNotExpressible` and the closed-set `ParameterContract.Mode` decode check both exist and pass (`go test ./internal/compiler/... -run 'ConventionOverrideNotExpressible'` — confirmed in full-suite run). `corevalidate.go` imports only `core` and stdlib (verified directly — no `check`/`ast`/`callgraph`/`originvalidate` import), and four independent `ImportsStayIndependent` guard tests pass. |
| 5 | Exactly ONE loan-liveness law exists at end of phase; intraprocedural law DELETED, not left coexisting; Nyquist compliant for the loan-liveness surface | ✓ VERIFIED | `grep -rn "func computeLoanLastUses"` returns **zero** results anywhere in the tree — only historical comments referencing the deleted name remain. `loanLivenessFixpoint` (the surviving law, defined once at `check.go:2114`) has 4 non-test call sites, inspected individually: `check.go:849` (interprocedural pass) and `:1704` (intraprocedural `checkBranch`) are the two admission-deciding invocations of the SAME function — this is one law applied at two scopes, not two laws; `:3801` ("evidence" derivation) and `:3870` (`aliasFactEndpoints`) are explicitly documented, and confirmed by inspection, as POST-HOC re-derivations for alias-fact/evidence purposes, never a second admission-deciding law (`check.go:3864`: "never a second admission-deciding law: computeLoanLastUses remains the sole law deciding conflict/expiry" — stale reference to the deleted name aside, the substance holds: no other law decides). QLT-07's Nyquist closure is a documented SUBSET closure (03-03/03-04/03-05 rows only), verified to exclude OWN-04 and the loop-carried-liveness limitation exactly as claimed, and `03-VALIDATION.md`'s diff against pre-phase state is exactly one added pointer line (`git diff 133a731 HEAD` confirmed), frontmatter/checkboxes byte-identical. |

**Score:** 5/5 roadmap success criteria verified true against the codebase.

### Specific Claims Checked Directly (per the verification brief)

| Claim | Verdict | Evidence |
|---|---|---|
| Criterion 1 non-vacuousness | Honest, not overclaimed | Source code itself documents the synthetic-corpus sweep as vacuous and names the mutation-kill test as the real proof of non-vacuous detection (see criterion 1 above) |
| Criterion 2 companion assertion (both halves) | Both halves present | Verified both "faulted peer diverges" and "unseamed peer / disengaged seam returns clean" in the same test function |
| Criterion 5 "exactly one" vs. `loanLivenessFixpoint`'s multiple call sites | Genuinely one decision point | Read all 4 non-test call sites; 2 are the SAME function used intra- and inter-procedurally (not a duplicate law), 2 are explicitly-commented post-hoc alias/evidence derivations, never admission-deciding |
| Criterion 3 at the CLI level / D-09-51 | Correctly scoped, not a regression | Re-derived independently at commit `133a731` — both fixtures were already CLI-refused pre-Phase-09 via a different code path |
| OWN-05 split honesty | Honest | OWN-05a `[x]` Complete (Phase 09, check+corevalidate), OWN-05b `[ ]` Pending (Phase 10, `interp`) — traceability arithmetic reconciles (`grep -c OWN-05a` = 5, `grep -c OWN-05b` = 4, as claimed) |
| QLT-07 subset scope | Honest | Closure text explicitly excludes OWN-04 rows and the loop-carried-liveness limitation; `03-VALIDATION.md`'s `nyquist_compliant: false` unchanged |
| Row `09-09-03` (Pattern B split) recorded RED | Confirmed honest | `PHASE-09-DEBT.md`'s Per-Task Verification Map shows this row `❌ red` with root cause (`deriveFunctionUsesParam` defect) and debt ID D-09-53, not silently folded to green; live re-run of `twin_b_accept.lang`/`twin_b_refuse.lang` confirms both still refuse identically today |

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `internal/compiler/corevalidate/corevalidate_peer_liveness*.go` | Peer forward-propagation liveness deriver | ✓ VERIFIED | Present, imports only `core`+stdlib, exercised by passing tests |
| `internal/compiler/check/check_peer_shape_differential_test.go` | Criterion-1/2 synthetic-shape differential | ✓ VERIFIED | Present, all 3 named tests pass live |
| `internal/compiler/session/session_peer_gate_test.go` | `.lang`-corpus exact-set divergence gate | ✓ VERIFIED | `TestNoUndeclaredCheckPeerDivergenceAcrossCorpus` passes; both retired divergence entries confirmed gone |
| `check.go`'s `loanLivenessFixpoint` (sole surviving law) | Single loan-liveness law | ✓ VERIFIED | `computeLoanLastUses` confirmed deleted (no definition anywhere); one function, two admission call sites (intra/inter), two post-hoc call sites |
| `.planning/phases/09-.../PHASE-09-DEBT.md` | Well-formed debt register, 16 items | ✓ VERIFIED | `items: 16` frontmatter matches 16 table rows; `TestDebtRegistersAreWellFormed/PHASE-09-DEBT.md` passes live |
| `.planning/milestones/M001-phases/03-.../03-VALIDATION.md` | Byte-identical except one pointer line | ✓ VERIFIED | `git diff 133a731 HEAD` shows exactly one added line |

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| `corevalidate.buildLoanChainIndex` | `core.OpCall` consult | Contract-aware consult of `loanCarry[operation.CalleeID].ReturnsBorrowOfParam` | ✓ WIRED | `corevalidate.go:1203` — live code, no longer unconditional propagation |
| `check.checkInterproceduralLoanLiveness` | `loanLivenessFixpoint` | Direct call, same function as intraprocedural | ✓ WIRED | `check.go:849` |
| `session.CheckCommandFile` | `check` → `corevalidate` → `originvalidate` | Fixed precedence pipeline | ✓ WIRED | Confirmed by live CLI runs surfacing check's own refusal first for `relay_escort_witness.lang`, and originvalidate's `core.origin_omitted` for the two accepted-program fixtures once check+corevalidate both admit |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|---|---|---|---|---|
| OWN-05 (split OWN-05a/b) | 09-05, 09-10 | Call-site ownership transfer, one meaning, two derivers this phase | ✓ SATISFIED (OWN-05a); OWN-05b correctly deferred to Phase 10 | `TestConventionOverrideNotExpressible`, `CallSiteTransferClassificationAgreesAcrossPeers` pass; REQUIREMENTS.md split confirmed honest |
| OWN-07 | 09-01 | corevalidate independently re-derives interprocedural loan liveness | ✓ SATISFIED | Peer forward-propagation code present, import-independent, seeded-fault + companion assertion both pass |
| OWN-08 | 09-03 | D-03-02 closed in both layers | ✓ SATISFIED (code); ⚠️ documentation checkbox inconsistency (see gap) | Both layers independently refuse `relay_escort_witness.lang`, confirmed live; traceability table says Complete but the requirement-list checkbox at REQUIREMENTS.md:54 was never flipped |
| OWN-09 | 09-07, 09-09 | Intraprocedural decision point retired | ✓ SATISFIED | `computeLoanLastUses` deleted, confirmed by exhaustive grep; ordering-stability and shadow-path-subsumption gates green pre-deletion |
| TRU-04 | 09-01, 09-02, 09-04, 09-06, 09-08 | Shadow-run differential, cost-bounded | ✓ SATISFIED | All differentials and cost-scaling tests pass live |
| QLT-07 | 09-10 | Nyquist compliance for loan-liveness surface | ✓ SATISFIED | Subset closure verified honest and non-overreaching; `03-VALIDATION.md` diff confirmed minimal |

No orphaned requirements found: every ID mapped to Phase 09 in REQUIREMENTS.md's traceability table (OWN-05a, OWN-07, OWN-08, OWN-09, TRU-04, QLT-07) appears in at least one plan's `requirements:` frontmatter, and every ID in plan frontmatter maps back to a REQUIREMENTS.md entry.

### Anti-Patterns Found

None. `grep -rn "TBD|FIXME|XXX"` across `internal/compiler/check`, `internal/compiler/corevalidate`, `internal/compiler/testsupport` (non-test files) returned zero matches. No stub returns, no placeholder comments found in the phase's production code.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|---|---|---|---|
| Full test suite | `go test ./...` | All packages `ok` (2 packages with no test files, expected) | ✓ PASS |
| `go vet` | `go vet ./...` | Clean | ✓ PASS |
| D-03-02 closure, check side | `go run ./cmd/lang --json check testdata/phase07/relay_escort_witness.lang` | Refused, `check.interprocedural_loan_liveness` | ✓ PASS |
| D-03-02 closure, corevalidate side | `go test ./internal/compiler/corevalidate -run TestRelayEscortWitnessCorevalidateIndependentlyRefusesMoveWhileBorrowed -v` | PASS | ✓ PASS |
| D-09-51 pre-existing verification | Worktree at `133a731`, same CLI commands on `twin_a_accept.lang`/`relay_depth2_accept.lang` | Both refused via `core.move_while_borrowed` pre-Phase-09, `core.origin_omitted` post-Phase-09 (different pre-existing bug, not a regression) | ✓ PASS |
| `computeLoanLastUses` deletion | `grep -rn "func computeLoanLastUses"` | No results | ✓ PASS |
| `TestDebtRegistersAreWellFormed` | `go test ./internal/compiler/session -run TestDebtRegistersAreWellFormed -v` | PASS (including `PHASE-09-DEBT.md`) | ✓ PASS |

### Human Verification Required

None. All must-haves were verifiable directly against the codebase and live test/CLI runs.

### Gaps Summary

The phase goal itself — a second, independently-implemented admission layer
(`corevalidate`) reaching the same interprocedural answer as `check`, and
D-03-02's closure in both layers — is verified TRUE against the codebase, with
all five ROADMAP success criteria independently confirmed by direct code
inspection, live test runs, and a scratch-worktree re-derivation of the
pre-existing-vs-regression question for D-09-51. All specific claims flagged in
the verification brief as "could be true on paper, false in code" were checked
directly and found honest: the vacuous synthetic-shape sweep is documented as
such with a real non-vacuous mutation-kill control; the companion assertion for
criterion 2 has both halves; `loanLivenessFixpoint`'s multiple call sites are
one law at two scopes plus two post-hoc derivations, not a second decision
point; D-09-51 is confirmed pre-existing rather than a regression; the OWN-05
split and QLT-07 subset-closure wording both reconcile honestly; and row
`09-09-03` is recorded red rather than papered over.

The one gap found is a **documentation-consistency defect, not a code or test
defect**: REQUIREMENTS.md's top-section requirement list still shows OWN-08's
checkbox as `[ ]` (unchecked) at line 54, while the traceability table two
sections below (line 206) says "OWN-08 | Phase 09 | Complete" — and the
underlying code/tests independently confirm the requirement's substance IS
delivered. Git history shows this checkbox was never flipped across any of the
09-03 through 09-10 commits that otherwise correctly updated the traceability
table and other requirement checkboxes (OWN-07, OWN-09, TRU-04, QLT-07,
OWN-05a all correctly show `[x]`). This is a one-line documentation fix, not a
reopened implementation gap — but per this project's own stated discipline
("a single-row overclaim is exactly the requirement-vs-code failure the debt
registers exist to catch"), an unflipped checkbox that undersells a delivered
requirement is the same class of error in miniature, and should not be waved
through silently.

---

_Verified: 2026-09-10_
_Verifier: Claude (gsd-verifier)_
