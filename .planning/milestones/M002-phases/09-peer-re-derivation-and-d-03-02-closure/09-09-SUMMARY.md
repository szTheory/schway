---
phase: 09-peer-re-derivation-and-d-03-02-closure
plan: 09
subsystem: compiler-check
tags: [ownership, loan-liveness, interprocedural, diagnostic-identity, deletion, gate]

requires:
  - phase: 09-peer-re-derivation-and-d-03-02-closure
    provides: "plan 09-07's three pre-deletion gates (ordering-stability baseline, D-09-49 Q1 enumeration, shadow-path reachability corpus) and plan 09-08's mandatory mid-phase gate authorization of the computeLoanLastUses deletion"
provides:
  - "Exactly ONE loan-liveness decision point in check: checkInterproceduralLoanLiveness's post-assembly pass, extended to the purely-intraprocedural case; lowering (analyzeArmBody/analyzeStraightLine) makes no ownership timing decision"
  - "computeLoanLastUses and its shadow:place/shadow:loan scaffolding DELETED in the same commit that removed the two lowering-time raises (D-09-10: never an intermediate state where neither path decides)"
  - "TestOwnershipSequenceExhaustive's 225,890-case differential moved to the post-assembly decision point on BOTH sides, byte-identical, assertSupportEqual unrelaxed"
  - "Plan 09-07's three pre-deletion gates re-run and reproduced, with every difference named, classified, and justified in writing (never a wholesale regeneration)"
  - "D-09-53: Pattern B's split (twin_b_refuse refused, twin_b_accept admitted) does NOT materialize, root-caused to a pre-existing deriveFunctionUsesParam defect, recorded as new debt rather than silently accommodated"
affects: ["09-10 (marks OWN-09 complete only if this plan's evidence is accepted; carries OWN-08/TRU-04 forward per this plan's own requirement-marking rule)", "Phase 10 (candidate landing site for D-09-51's originvalidate fix and D-09-53's deriveFunctionUsesParam fix)"]

actuals:
  tokens: 27500
  tasks: 4
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Evidence-only re-derivation: a deleted admission-deciding function's mechanics survive under a new name and role (loanFinalUseEvidence), feeding test-visible fields with zero production readers, never re-consulted for a decision"
    - "A same-channel ':stmt'-suffixed span key lets a post-assembly repair reconstruct a whole-statement span from a per-operation span map without adding a second threading path through checkLinear/checkBranch's return signatures"
    - "Precedence rule: a timing-independent inline lowering failure anywhere in the WHOLE PROGRAM gates the post-assembly pass off entirely (pre-existing D-08-27 rule), so pairing a lowering-time code against a newly-deferred code no longer demonstrates ordering claims that assume both decide at the same layer"

key-files:
  created: []
  modified:
    - internal/compiler/check/check.go
    - internal/compiler/check/check_test.go
    - internal/compiler/check/check_exclusive_test.go
    - internal/compiler/check/check_ordering_stability_test.go
    - internal/compiler/check/check_shadow_subsumption_test.go
    - testdata/phase2/move_while_borrowed.lang
    - testdata/phase3/exclusive_exclusive_reject.lang
    - testdata/phase3/exclusive_move_reject.lang
    - testdata/phase3/shared_exclusive_reject.lang
    - .planning/phases/09-peer-re-derivation-and-d-03-02-closure/PHASE-09-DEBT.md

key-decisions:
  - "Task 1's straightLineSupportAtDecisionPoint composes analyzeStraightLine with the extended checkInterproceduralLoanLiveness for the exhaustive test's production side; oracleStraightLine's oracleDeferredLoanLiveness defers the SAME two decisions, independently derived (no shared identifier with production loan-liveness machinery, verified by source scan)."
  - "LoanFinalUses/States (test-visible evidence, zero production readers, grep-verified) are re-derived by loanFinalUseEvidence -- the deleted computeLoanLastUses' mechanics kept verbatim under a new name and a strictly evidence-only role, never consulted for an admission decision. This was chosen over deriving these fields post-hoc from the (possibly truncated) real Operations list, because the independent oracle populates them from the FULL declared body even when a later per-binding fact fails partway through the walk -- production must answer for the same full body, not a truncated prefix."
  - "The new-loan-vs-existing-loan borrow_conflict scan and the extended move_while_borrowed scan share the SAME lastUseIndexByLoan derivation (materializeLoanEndpoints over the real assembled function) and pick whichever offending event has the lower operation index, mirroring the single forward walk the deleted lowering-time code used to perform."
  - "testOnlyForceUniformLoanJoin is re-attached inside checkInterproceduralLoanLiveness itself (forcing a loan's last use to its own arm block's last operation), gated to function.Match != nil -- its pre-restructure scope (analyzeArmBody only, never analyzeStraightLine) -- rather than left in analyzeArmBody where it would now be inert (the real admission decision moved to the post-assembly pass)."
  - "Four testdata fixtures (implicit copy of a borrowed/exclusive view -> explicit take) are edited: Buffer withholds Copy, and once lowering stops deciding the timing-dependent code inline and proceeds through every binding unconditionally, the OLD implicit copy would newly fail with the timing-independent ownership.transfer_requires_take BEFORE the deferred code the fixture exists to demonstrate ever decides -- the precedence rule masking an unintended defect instead of the intended one. These fixtures are also consumed directly by internal/compiler/session's borrow-conflict tests (out of this plan's file scope); fixing the fixture, not weakening any assertion, restored their passing status with zero session-package edits."
  - "A same-channel ':stmt'-suffixed span entry (analyzeStraightLine/analyzeArmBody's existing CallSpans map, merged into spanByOperationID exactly like every other entry) carries the WHOLE binding statement's span, needed only by borrowConflictDiagnosticPostAssembly's narrow_to_shared_borrow repair (which rewrites the whole statement, not just the RHS token the diagnostic's own Primary/causes use) -- discovered as a genuine repair-correctness bug via cmd/lang-repair's own single-pass repair test going from pass to 'reverify_failed' failing to a corrected pass."
  - "Pattern B's promised split (D-08-41's harm, D-09-08's justification for authorizing the deletion) does NOT materialize: both twin_b fixtures now report check.interprocedural_loan_liveness identically, root-caused to deriveFunctionUsesParam's default branch treating a plain OpMove/OpCopy forward-to-return as a 'use' (contradicting its own doc comment). Recorded as new debt (D-09-53), NOT fixed here (out of Task 2's authorized scope, Phase-08-vintage code, un-audited blast radius on other committed fixtures) and NOT accommodated by adjusting the twin fixtures -- exactly Task 4(b)'s own named contingency."
  - "TestDiagnosticSelectionFollowsFunctionDeclarationOrder's fixture pair is redesigned (both functions now carry a deferred code) rather than left pairing a lowering-time code against a deferred one, which no longer demonstrates the rule under test post-restructure (a lowering-time failure ANYWHERE in the program gates the post-assembly pass off for the WHOLE program, masking the deferred code regardless of declaration order)."

patterns-established:
  - "When an admission-deciding function is deleted but its algorithm must survive as pure evidence for an independent-oracle comparison, rename and re-scope it explicitly in its own doc comment rather than leaving its old identity (and the risk of a caller re-consulting it for a decision) ambiguous."

requirements-completed: []

coverage:
  - id: D1
    description: "Exactly one loan-liveness decision point exists in check (checkInterproceduralLoanLiveness's post-assembly pass); computeLoanLastUses and its two summary-blind call sites are deleted in the same commit that removes the lowering-time raises"
    requirement: "OWN-09"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestSingleLoanLivenessLaw"
        status: pass
      - kind: static
        ref: "grep -v '^[[:space:]]*//' internal/compiler/check/check.go | grep -c computeLoanLastUses = 0"
        status: pass
    human_judgment: false
  - id: D2
    description: "TestOwnershipSequenceExhaustive proves production/oracle agreement over the full 225,890-case enumeration at the moved (post-assembly) decision point, assertSupportEqual unrelaxed, case count explicitly asserted"
    requirement: "OWN-09"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestOwnershipSequenceExhaustive"
        status: pass
    human_judgment: false
  - id: D3
    description: "Plan 09-07's three pre-deletion gates (ordering-stability baseline, shadow-path reachability/refusal corpus, timing-independent fence) re-run and reproduce or are justified in writing"
    requirement: "OWN-09"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_ordering_stability_test.go#TestInterproceduralDiagnosticOrderingStability"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_shadow_subsumption_test.go#TestShadowPathSubsumptionCorpus"
        status: pass
    human_judgment: false
  - id: D4
    description: "Pattern B's twin pair (twin_b_refuse.lang/twin_b_accept.lang) exercised end to end through the CLI; the outcome (split does not materialize) is reported and root-caused, not silently accommodated"
    requirement: "OWN-09"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestInterproceduralLivenessTwinPatternBRealFixtures"
        status: pass
      - kind: other
        ref: "go run ./cmd/lang --json check testdata/phase08/twin_b_refuse.lang && go run ./cmd/lang --json check testdata/phase08/twin_b_accept.lang (both refused with check.interprocedural_loan_liveness)"
        status: pass
    human_judgment: true
    rationale: "D-08-41's recorded harm is NOT removed by this deletion (a finding of the first importance per Task 4(b)); a human must decide whether D-09-53's diagnosed-but-unfixed deriveFunctionUsesParam defect should be scoped into a follow-up plan before Phase 09 is considered fully closed on OWN-09's split-verdict claim."
  - id: D5
    description: "go test ./... && go vet ./... exits 0, including cmd/lang-repair's own single-pass repair test (discovered regression, fixed via a new ':stmt' span channel)"
    verification:
      - kind: integration
        ref: "go test ./... (full repo) && go vet ./..."
        status: pass
    human_judgment: false

duration: ~5h
completed: 2026-09-10
status: complete
---

# Phase 09 Plan 09: Collapse to One Loan-Liveness Decision Point Summary

**Deleted the summary-blind `computeLoanLastUses` early call site under D-09-08's authorization, extended `checkInterproceduralLoanLiveness` to become the SOLE post-assembly decision point for `ownership.move_while_borrowed`/`ownership.borrow_conflict`, reproduced all three of plan 09-07's pre-deletion gates with every difference justified in writing, and discovered — rather than silently absorbed — that Pattern B's promised twin-fixture split does not materialize, root-causing it to a pre-existing `deriveFunctionUsesParam` defect recorded as new debt (D-09-53).**

## Performance

- **Duration:** ~5 hours (deep investigation of CFG-aware post-assembly detection, span-channel plumbing, and cross-package masking effects)
- **Tasks:** 4 completed
- **Files modified:** 10 (0 created)

## Accomplishments

- **Task 1+2 (one commit, per D-09-10):** Deleted `computeLoanLastUses` and its `shadow:place:*`/`shadow:loan:*` scaffolding. `analyzeArmBody`/`analyzeStraightLine` now emit `core.OpMove`/`core.OpBorrowShared`/`core.OpBorrowExclusive` unconditionally, making no ownership-timing decision. `checkInterproceduralLoanLiveness` is extended with a new-loan-vs-existing-loan conflict scan (`ownership.borrow_conflict`) and a "neither direction" intraprocedural classification (`ownership.move_while_borrowed`), both sharing the pass's existing `lastUseIndexByLoan` derivation — never a second liveness law. `loanFinalUseEvidence` (a renamed, evidence-only replacement for the deleted function, mechanically identical) keeps `LoanFinalUses`/`States` populated for the FULL declared body regardless of where a later per-binding fact truncates the real walk, matching the independent oracle's own behavior. `testOnlyForceUniformLoanJoin` is re-attached inside the post-assembly pass, gated to branch-bodied functions.
- **Task 1's exhaustive differential:** `straightLineSupportAtDecisionPoint` composes lowering with the extended post-assembly pass; `oracleDeferredLoanLiveness` defers the oracle's matching two decisions, verified independent by source scan (references no production loan-liveness identifier). All 225,890 cases (explicit counter asserted) still compare byte-identically via unrelaxed `assertSupportEqual`.
- **Task 3:** Re-ran and reproduced (or justified in writing) plan 09-07's ordering-stability baseline, shadow-path reachability register (re-derived for `lastUseIndexByLoan`, pre-deletion sets retained as historical record), and refusal baseline. `TestDiagnosticSelectionFollowsFunctionDeclarationOrder`'s fixture pair was redesigned since pairing a lowering-time code against a now-deferred one no longer demonstrates the rule under test.
- **Task 4:** `TestSingleLoanLivenessLaw` extended with the second retired identifier and an exact, named count of `loanLivenessFixpoint` call sites (4: `checkInterproceduralLoanLiveness`, `checkBranch`, `aliasFactEndpoints`, `loanFinalUseEvidence`). `twin_b_refuse.lang`/`twin_b_accept.lang` exercised end to end through the CLI: **the split does NOT materialize** — both refuse identically with `check.interprocedural_loan_liveness`, diagnosed to `deriveFunctionUsesParam`'s pre-existing, unrelated defect and recorded as D-09-53 rather than fixed or accommodated. All five `ownership.*` codes and `check.interprocedural_loan_liveness` verified to still exist and fire on a triggering fixture; corpus-wide peer-divergence gate stays green; `go test ./... && go vet ./...` exits 0.
- **Discovered and fixed a genuine repair-correctness bug** (Rule 1): `borrowConflictDiagnosticPostAssembly`'s `narrow_to_shared_borrow` repair initially used the borrow operation's own RHS-token span for its `Span` field, but its `Replacement` text rewrites the WHOLE binding statement — applying the repair literally corrupted source (`cmd/lang-repair`'s `TestRepairDriverFixesEveryDefectClassSinglePass/borrow` regressed to `reverify_failed`). Fixed by adding a `":stmt"`-suffixed span entry to the same span channel, carrying the whole-statement span lowering already has available.
- **Discovered and fixed four testdata fixtures** (Rule 1) whose implicit-copy "keep the loan referenced" idiom (`let observed = view`) newly failed with the timing-independent `ownership.transfer_requires_take` (Buffer withholds Copy) once lowering stopped short-circuiting at the timing-dependent binding — masking the very code each fixture exists to demonstrate. Rewritten to `let observed = take view`, preserving the load-bearing later reference without requiring Copy. Three of these fixtures are also consumed directly by `internal/compiler/session`'s own borrow-conflict tests (out of this plan's file scope); the fix restored their passing status with zero session-package edits.

## Task Commits

1. **Tasks 1+2 (one commit per D-09-10's "never an intermediate state" rule)** — `b8fe3df` (feat)
2. **Task 3 (pre-deletion baseline reproduction)** — `ef7f698` (test)
3. **Task 4b (Pattern B debt record)** — `cc327d5` (docs)

Task 4's code changes (`TestSingleLoanLivenessLaw` extension, `TestInterproceduralLivenessTwinPatternBRealFixtures` update) landed inside commit `b8fe3df` since they modify `check_test.go`, the same file Tasks 1-2 already touched — documented here rather than forced into an artificial partial-file split.

**Plan metadata:** (this commit)

## Files Created/Modified

- `internal/compiler/check/check.go` — deleted `computeLoanLastUses`; extended `checkInterproceduralLoanLiveness` with borrow-conflict detection and intraprocedural move classification; added `loanFinalUseEvidence`, `moveWhileBorrowedDiagnosticPostAssembly`, `borrowConflictDiagnosticPostAssembly`, `lastUseSpanAt`, `placesByID`; re-attached `testOnlyForceUniformLoanJoin`
- `internal/compiler/check/check_test.go` — `straightLineSupportAtDecisionPoint`, `oracleDeferredLoanLiveness`, rewritten `TestOwnershipSequenceExhaustive`/`oracleStraightLine`, extended `TestSingleLoanLivenessLaw`, updated `TestInterproceduralLivenessTwinPatternBRealFixtures`; deleted three `computeLoanLastUses`-dependent tests
- `internal/compiler/check/check_exclusive_test.go`, `testdata/phase2/move_while_borrowed.lang`, `testdata/phase3/{exclusive_exclusive_reject,exclusive_move_reject,shared_exclusive_reject}.lang` — masking-avoidance fixture edits
- `internal/compiler/check/check_ordering_stability_test.go`, `internal/compiler/check/check_shadow_subsumption_test.go` — baseline/register updates with written justifications
- `.planning/phases/09-peer-re-derivation-and-d-03-02-closure/PHASE-09-DEBT.md` — D-09-53 recorded

## Decisions Made

See `key-decisions` in frontmatter for the full account.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `narrow_to_shared_borrow` repair span corrupted source when applied**
- **Found during:** Task 2, verified via `cmd/lang-repair`'s own regression test
- **Issue:** The repair's `Span` used the RHS-token span (e.g. 6 characters covering just `buffer`), but its `Replacement` rewrote the whole `let NAME = borrow mut SOURCE` statement — applying it literally inserted the whole-statement text into a 6-character hole, corrupting the program
- **Fix:** Added a `":stmt"`-suffixed span entry to the existing `CallSpans` channel, carrying the whole binding statement's span (`binding.Span`, distinct from `binding.RHS.Span`); `borrowConflictDiagnosticPostAssembly` consults it with a fallback to the RHS span
- **Files modified:** `internal/compiler/check/check.go`
- **Verification:** `cmd/lang-repair`'s `TestRepairDriverFixesEveryDefectClassSinglePass/borrow` (regressed, then fixed)
- **Committed in:** `b8fe3df`

**2. [Rule 1 - Bug] Four fixtures' implicit-copy loan-liveness idiom masked their own intended diagnostic**
- **Found during:** Task 1-2, via `TestCheckerVerdictsUnchanged`, `TestExclusiveBorrowLowersToCore`, and `internal/compiler/session`'s `TestBorrowConflictMatrix`/`TestBorrowConflictCauseChain`/`TestExclusiveMoveRejected`
- **Issue:** `let observed = view` (an implicit copy of a borrowed/exclusive view, used purely to keep the loan's last use extending past a move or a second borrow) failed with the timing-independent `ownership.transfer_requires_take` once lowering stopped short-circuiting before it — Buffer withholds Copy
- **Fix:** Rewrote to `let observed = take view`/`take first`, preserving the load-bearing later reference (still extends the loan's last use) without requiring Copy
- **Files modified:** `internal/compiler/check/check_exclusive_test.go`, `testdata/phase2/move_while_borrowed.lang`, `testdata/phase3/exclusive_exclusive_reject.lang`, `testdata/phase3/exclusive_move_reject.lang`, `testdata/phase3/shared_exclusive_reject.lang`
- **Verification:** `TestCheckerVerdictsUnchanged`, `TestExclusiveBorrowLowersToCore`, `internal/compiler/session`'s three borrow-conflict tests, all pass with zero session-package edits
- **Committed in:** `b8fe3df`

---

**Total deviations:** 2 auto-fixed (both Rule 1 - bug), plus the plan's own explicitly-anticipated Task 3/4 baseline-reproduction and Pattern-B-non-materialization findings (documented above and in `key-decisions`, not deviations from the deviation-rule framework since the plan itself names these as expected possible outcomes with a prescribed handling).
**Impact on plan:** Both auto-fixes were necessary for correctness (a corrupting repair, and fixtures that no longer demonstrated their own stated purpose). No scope creep — neither touched any file outside this plan's declared `files_modified` plus the testdata fixtures the masking bug required.

## Known Stubs

None.

## Threat Flags

None — the threat register's five `mitigate`-disposition items (T-09-30 through T-09-34, T-09-33, T-09-23) are all directly addressed: the oracle's independence is source-scan-verified (T-09-30), the 225,890-case enumeration is counter-asserted and unreduced (T-09-31), `assertSupportEqual` is byte-identical to its pre-plan state (T-09-32), the shadow-path refusal baseline reproduces with zero new admissions (T-09-03), every ordering-stability difference is named/classified/justified (T-09-21), `loanLivenessFixpoint`/`materializeLoanEndpoints`/`cfgBlockSpec` all survive (T-09-27), Tasks 1-2 landed in one commit (T-09-28), all diagnostic trigger conditions/spans are preserved (T-09-29), `testOnlyForceUniformLoanJoin` is re-attached and its dependent tests pass (T-09-34), `qlt02_budget_manifest.json` has zero diff and its audit is green (T-09-33), and the three timing-independent guards stay byte-identical (T-09-23).

## Issues Encountered

**D-09-53 (Pattern B's split does not materialize):** the single most significant finding of this plan. D-08-41's recorded harm ("Pattern B's twin pair could not demonstrate a differing end-to-end CLI verdict") is NOT removed by this authorized deletion — both `twin_b_refuse.lang` and `twin_b_accept.lang` now reach `checkInterproceduralLoanLiveness` (the summary-blind mask is correctly gone) but both are refused identically with `check.interprocedural_loan_liveness`. Root-caused directly (not merely suspected) to `deriveFunctionUsesParam`'s `default: usesParam = true` branch treating a plain `OpMove`/`OpCopy` forward-to-return as a "use," contradicting its own doc comment's stated exemption — a pre-existing, Phase-08-vintage defect this deletion merely stopped masking, discovered one layer further down the pipeline than any prior planning document anticipated (the same "one validator's own defect masks another's" shape D-09-08 and D-09-51 both already name). Per Task 4(b)'s own explicit contingency, this is reported and recorded as new debt (D-09-53, landing phase not yet decided — candidate Phase 10, alongside D-09-51) rather than fixed inline (out of this plan's authorized scope, Rule 4 territory: an un-audited blast radius on other committed Phase 08 fixtures) or accommodated by adjusting the twin fixtures.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- Exactly one loan-liveness decision point exists in `check`; the summary-blind path is fully deleted, not left coexisting (OWN-09's core claim, now evidenced).
- Plan 09-07's strongest evidence (the 225,890-case exhaustive differential) still holds, at the later decision point, with neither size nor strictness reduced.
- Two open debt items now point at Phase 10 as a candidate landing site: D-09-51 (`originvalidate`'s `OpCall` transparency defect) and D-09-53 (`deriveFunctionUsesParam`'s identity-forward defect) — both "a validator's own pre-existing defect, unmasked one layer at a time by this phase's authorized deletions."
- OWN-09 stays whatever plan 09-10's own gate decides given this plan's evidence (this plan does not itself flip requirement status per its own requirement-marking rule — OWN-09 is also carried by plan 09-10).
- No blockers for plan 09-10, beyond the two named open debt items above requiring an explicit landing-phase decision before Phase 09 is considered fully closed.

---
*Phase: 09-peer-re-derivation-and-d-03-02-closure*
*Completed: 2026-09-10*

## Self-Check: PASSED

- FOUND: internal/compiler/check/check.go
- FOUND: internal/compiler/check/check_test.go
- FOUND: internal/compiler/check/check_exclusive_test.go
- FOUND: internal/compiler/check/check_ordering_stability_test.go
- FOUND: internal/compiler/check/check_shadow_subsumption_test.go
- FOUND: testdata/phase2/move_while_borrowed.lang
- FOUND: testdata/phase3/exclusive_exclusive_reject.lang
- FOUND: testdata/phase3/exclusive_move_reject.lang
- FOUND: testdata/phase3/shared_exclusive_reject.lang
- FOUND: .planning/phases/09-peer-re-derivation-and-d-03-02-closure/PHASE-09-DEBT.md
- FOUND commit b8fe3df (feat(09-09): collapse check to one loan-liveness decision point)
- FOUND commit ef7f698 (test(09-09): reproduce plan 09-07's pre-deletion baselines)
- FOUND commit cc327d5 (docs(09-09): record Pattern B split non-materialization as debt)
- `go test ./internal/compiler/check/... -count=2 -shuffle=on` and `go test ./... && go vet ./...` both verified green before this SUMMARY was finalized
