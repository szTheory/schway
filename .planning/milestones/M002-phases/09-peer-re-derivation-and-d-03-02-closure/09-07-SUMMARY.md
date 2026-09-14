---
phase: 09-peer-re-derivation-and-d-03-02-closure
plan: 07
subsystem: check
tags: [ownership, loan-liveness, diagnostic-identity, gate, testdata]

requires:
  - phase: 09-peer-re-derivation-and-d-03-02-closure
    provides: "09-01's phase-wide artifact registry and decision-coverage baseline"
provides:
  - "An explicit diagnostic-ordering-stability rule assertion plus a literal pre-restructure baseline table over every currently-refused testdata fixture (check_ordering_stability_test.go)"
  - "D-09-49 Q1 settled by enumeration: deferring ownership.move_while_borrowed to post-assembly does not change ownership.use_after_move for any of the 6 fixtures where it currently fires"
  - "A fence proving the three timing-independent ownership.* codes fire from per-binding facts only, never activeLoans/loanUses"
  - "A literal pre-deletion AST-shadow-path refusal baseline (shadowPathRefusalBaseline) plan 09-09 must reproduce, plus a source-scan-and-perturbation-backed reachable/unreachable classification of all five ownership.* codes"
affects: ["09-08 (mid-phase gate consumes these baselines and the D-09-49 Q1 answer as agenda evidence)", "09-09 (the deletion plan re-runs every gate in this plan and must reproduce or justify every baseline entry)"]

actuals:
  tokens: 11255
  tasks: 3
  commits: 2

tech-stack:
  added: []
  patterns:
    - "Literal in-test baseline tables (never golden-file replays) for facts a restructure must reproduce or justify in writing"
    - "Dynamic corpus re-derivation cross-checked against a declared table, so an enumeration claim is falsifiable rather than merely asserted"
    - "Source-scan guard-window extraction (backward search from a raise site to its nearest matching guard pattern) reused symmetrically for both a fence (guard must NOT mention X) and a reachability proof (guard MUST mention X)"
    - "Reuse of the existing testOnlyForceUniformLoanJoin fault-injection seam as a genuine falsifying perturbation, not merely a passing test"

key-files:
  created:
    - internal/compiler/check/check_ordering_stability_test.go
    - internal/compiler/check/check_shadow_subsumption_test.go
  modified: []

key-decisions:
  - "D-09-49 Q1 ANSWER: no existing fixture exhibits the interaction, enumerated over 6 fixtures where ownership.move_while_borrowed currently fires (phase08/twin_b_accept.lang, phase08/twin_b_refuse.lang, phase2/move_while_borrowed.lang, phase2/reborrow_while_moved.lang, phase3/branch_one_arm_shared_reject.lang, phase3/exclusive_move_reject.lang) — in every case the moved place's only later reference is through the borrowed view's own binding name, never the moved place's own name again."
  - "shadowPathReachableCodes = {ownership.move_while_borrowed, ownership.borrow_conflict} (guards consult activeLoans/loanUses); shadowPathUnreachableCodes = {ownership.use_after_move, ownership.borrow_requires_share, ownership.transfer_requires_take} (guards consult only !source.initialized / hasTypeAbility) — classification proven by source scan AND by a testOnlyForceUniformLoanJoin perturbation that flips a swept program's move_while_borrowed verdict."
  - "Combined Task 2 and Task 3 into a single commit since both share the same output file (check_shadow_subsumption_test.go) and were authored together; documented here rather than forced into two artificial partial commits of the same file."

patterns-established:
  - "Pre-restructure baseline tables committed as literal data, with a header comment naming the plan that must reproduce them or justify each difference in writing"

requirements-completed: []

coverage:
  - id: D1
    description: "Explicit diagnostic-ordering-stability rule assertion (declaration order decides, verified both directions) plus structural slice-not-map precondition"
    requirement: "OWN-09"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_ordering_stability_test.go#TestDiagnosticSelectionFollowsFunctionDeclarationOrder"
        status: pass
    human_judgment: false
  - id: D2
    description: "Literal pre-restructure ordering baseline over every currently-refused testdata fixture (32 fixtures)"
    requirement: "OWN-09"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_ordering_stability_test.go#TestInterproceduralDiagnosticOrderingStability"
        status: pass
    human_judgment: false
  - id: D3
    description: "D-09-49 Q1 settled by enumeration over all 6 move_while_borrowed fixtures; zero violate the positive no-subsequent-access property"
    requirement: "OWN-09"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_shadow_subsumption_test.go#TestUseAfterMoveUnchangedByDeferredMoveWhileBorrowed"
        status: pass
    human_judgment: false
  - id: D4
    description: "Fence proving the three timing-independent ownership.* codes fire from per-binding facts only (dynamic + structural)"
    requirement: "OWN-09"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_shadow_subsumption_test.go#TestTimingIndependentOwnershipCodesFireFromPerBindingFacts"
        status: pass
    human_judgment: false
  - id: D5
    description: "AST-shadow-path reachability register (both sets declared, union is exactly the five ownership.* codes), a perturbation proof of dependence, and a literal pre-deletion refusal baseline over the extended exhaustive generator"
    requirement: "OWN-09"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_shadow_subsumption_test.go#TestShadowPathSubsumptionCorpus"
        status: pass
    human_judgment: false

duration: 55min
completed: 2026-09-10
status: complete
---

# Phase 09 Plan 07: Pre-Deletion Gates for the AST-Shadow Loan-Liveness Path Summary

**Three gates authored against CURRENT checker behavior — an explicit diagnostic-ordering-stability assertion, D-09-49 Q1 settled by enumeration, and a proven (not asserted) AST-shadow-path reachability corpus — all green before plan 09-09 deletes `computeLoanLastUses`.**

## Performance

- **Duration:** 55 min
- **Started:** 2026-09-10T22:07:40Z (approx, per session state)
- **Completed:** 2026-09-10
- **Tasks:** 3 completed
- **Files modified:** 2 (both newly created test files)

## Accomplishments

- Authored `TestDiagnosticSelectionFollowsFunctionDeclarationOrder` and `TestInterproceduralDiagnosticOrderingStability`, stating D-09-13's ordering rule explicitly (not via corpus replay) and committing a literal `(fixture, code, id)` baseline over all 32 currently-refused testdata fixtures that plan 09-09 must reproduce or justify.
- Settled D-09-49 Q1 by enumeration: `TestUseAfterMoveUnchangedByDeferredMoveWhileBorrowed` walks every one of the 6 fixtures where `ownership.move_while_borrowed` fires today and proves, per fixture, that no binding or result position after the offending move still references the moved place — deferring the refusal to post-assembly cannot newly trip `ownership.use_after_move` for any existing fixture.
- Fenced the three timing-independent codes (`ownership.use_after_move`, `ownership.borrow_requires_share`, `ownership.transfer_requires_take`) both dynamically (each still fires on a triggering fixture) and structurally (source-scan proving each raise site's guard mentions `initialized`/`hasTypeAbility` and never `activeLoans`/`loanUses`).
- Declared and proved the AST-shadow-path reachability register: `shadowPathReachableCodes` = `{move_while_borrowed, borrow_conflict}`, `shadowPathUnreachableCodes` = the other three — proven by source scan (reachable codes' guards DO consult `activeLoans`) and by a genuine falsifying perturbation (the existing `testOnlyForceUniformLoanJoin` seam flips a swept program's verdict, demonstrating real dependence rather than mere co-location in the same function). Committed a literal pre-deletion refusal baseline (40 entries) over the extended `generatedOwnershipBody` exhaustive alphabet.
- Verified throughout: `internal/compiler/check/check.go` is byte-identical to its pre-plan state; `computeLoanLastUses` and its two summary-blind call sites remain present and reachable (grep-count verify), authorizing nothing about their deletion — that is plan 09-09's charter, gated by plan 09-08.

## Task Commits

Each task was committed atomically:

1. **Task 1: Explicit diagnostic-ordering-stability assertion** - `99e75cc` (test)
2. **Task 2 + Task 3: D-09-49 Q1 enumeration, timing-independent fence, and shadow-path reachability corpus** - `d11ad74` (test) — combined into one commit since both tasks share the same output file (`check_shadow_subsumption_test.go`) and were authored together; documented here rather than forced into two artificial partial commits of the same file.

**Plan metadata:** (this commit)

## D-09-49 Q1: The Answer

**No existing fixture exhibits the interaction.** Enumerated over all 6 fixtures where `ownership.move_while_borrowed` currently fires (`phase08/twin_b_accept.lang`, `phase08/twin_b_refuse.lang`, `phase2/move_while_borrowed.lang`, `phase2/reborrow_while_moved.lang`, `phase3/branch_one_arm_shared_reject.lang`, `phase3/exclusive_move_reject.lang`), in every case the moved place's only later reference is through the borrowed view's own binding name (a different name already holding the alias), never the moved place's own name again — so deferring the refusal to post-assembly cannot newly trip `ownership.use_after_move` for any of them. No exception set is needed; nothing is raised for plan 09-08's gate on this specific question.

## Deviations from Plan

### Auto-fixed Issues

None — plan executed exactly as written, with one process note (not a code deviation): Tasks 2 and 3 landed in a single commit since they modify the same file and were authored as one coherent unit (see Task Commits above).

## Known Stubs

None. This plan authors test-only gates against current production behavior; no stub, placeholder, or unwired data path was introduced.

## Threat Flags

None. All new surface is test-only (package `check`, `_test.go` files); no new network endpoint, auth path, file-access pattern, or schema change at a trust boundary was introduced.

## Self-Check: PASSED

- FOUND: internal/compiler/check/check_ordering_stability_test.go
- FOUND: internal/compiler/check/check_shadow_subsumption_test.go
- FOUND commit 99e75cc (test(09-07): explicit diagnostic-ordering-stability gate against current behavior)
- FOUND commit d11ad74 (test(09-07): settle D-09-49 Q1 and prove AST-shadow-path subsumption)
- `go test ./... && go vet ./...` exits 0 (verified before this SUMMARY was written)
- `grep -v '^[[:space:]]*//' internal/compiler/check/check.go | grep -c 'computeLoanLastUses'` = 3 (unchanged; check.go untouched by this plan)
