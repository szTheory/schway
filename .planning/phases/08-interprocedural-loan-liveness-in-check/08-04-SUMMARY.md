---
phase: 08-interprocedural-loan-liveness-in-check
plan: 04
subsystem: check
tags: [ownership, loan-liveness, fixpoint, iteration-bound, cfg, mutation-kill, go]

# Dependency graph
requires:
  - phase: 08-interprocedural-loan-liveness-in-check
    plan: 03
    provides: "The nine-fixture adversarial corpus and the mechanically-enforced twin discipline that establishes success criterion 1"
provides:
  - "loanLivenessFixpoint's cycle pre-walk converted from a self-calling closure to an explicit-stack iterative three-colour DFS (D-08-19a), ported from callgraph.Order's own structure"
  - "The coded check.cfg_back_edge diagnostic (D-08-19b), replacing a bare fmt.Errorf-wrapped string, with a single cycle_block cause naming the offending block"
  - "A derived, auditable, fail-closed iteration bound on the worklist loop -- loanLivenessBoundFactor * blocks * (distinctLoans + 1) -- and its named refusal, check.loan_liveness_bound_exceeded (D-08-14)"
  - "loanLivenessBoundSeam, mutation-killed in both directions by TestLoanLivenessBoundMutationKilled (D-08-16/QLT-08)"
  - "TestLoanLivenessBoundValueIsNotInDiagnosticIdentity, proving the bound's own computed integer never appears in the refusal's Causes or Message and never moves its published ID (D-08-18)"
affects: [09-peer-re-derivation-and-d-03-02-closure]

# Actuals (#2632)
actuals:
  tokens: 7954
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Every intraprocedural or whole-program fixpoint/traversal in this codebase now uses an explicit-stack iterative DFS, never native recursion, for its own cycle detection: callgraph.Order set the shape (D-07-18), loanLivenessFixpoint's cycle pre-walk now ports it verbatim (D-08-19a) rather than inventing a second one."
    - "A whole-function-property refusal (no single offending operation) carries Primary = the function's own declaration span and Causes that name only categorical facts, never the computed numeric value that gates the refusal -- so retuning an internal threshold can never move a diagnostic's published ID (D-08-18, extending D-07-16/D-07-43's cycle-witness-stability precedent to a bound value)."
    - "A derived bound over a lattice-height-bounded fixpoint must account for the fixpoint's own per-visit accounting floor (here: one transfer-function evaluation per block regardless of loan count) or a zero-loan legitimate input computes a zero bound and trips immediately -- the '+1' floor in loanLivenessBound(blocks, loans) = factor*blocks*(loans+1) is load-bearing, not decorative, and was caught by running the full existing suite against the literal formula before committing."

key-files:
  created: []
  modified:
    - internal/compiler/check/check.go
    - internal/compiler/check/check_test.go
    - .planning/phases/08-interprocedural-loan-liveness-in-check/PHASE-08-DEBT.md

key-decisions:
  - "The bound formula is loanLivenessBoundFactor * blockCount * (distinctLoanCount + 1), not the literal blockCount * distinctLoanCount the plan's prose named. Rule 1 auto-fix: the literal zero-loan-count formula computes a bound of exactly zero for any function with no borrows, and since loanLivenessFixpoint runs unconditionally for every checked function (including every non-borrowing one), this would refuse nearly every legal program on its very first non-trivial worklist iteration -- directly violating the plan's own must_haves.truths criterion that 'no legal program trips it'. The '+1' floor preserves the required doubling-block-count-doubles-the-bound property (the floor is a constant additive term inside a factor multiplied by blockCount, so scaling blockCount still scales the whole product linearly) and, for the single-block reborrow-chain shape, reduces to exactly the pre-existing, independently-derived 4*n+4 ceiling TestLivenessWorkScale/TestReborrowChainWorkIsLinear already assert -- a strong independent confirmation the fix is the intended formula, not an arbitrary patch."
  - "The bound refusal's Causes name the function ID and a categorical 'block_count' Kind with NO Detail value (not the actual block count number), rather than literally disclosing the count as the plan's prose could be read to suggest. This is required by the plan's OWN Task 3(b) acceptance criterion -- two runs with differing computed bounds (varied by block count) must produce IDENTICAL diagnostic IDs -- which is only possible if Causes never encode the value that varies between those two runs. Verified directly by TestLoanLivenessBoundValueIsNotInDiagnosticIdentity."
  - "loanLivenessFixpoint gained a diagnostic.Span parameter (the caller's own function-level span) rather than deriving one internally, since the function itself has no natural span of its own -- checkInterproceduralLoanLiveness passes function.Span, checkBranch passes function.Body.Span, and the two post-hoc, single-block, cycle-incapable call sites (aliasFactEndpoints, computeLoanLastUses' shadow path) pass a zero-valued diagnostic.Span{} since their error path is structurally unreachable (single block, no successors, can never cycle)."

patterns-established:
  - "Whole-function-property refusal diagnostics (no single offending operation) use the function's own declaration span as Primary and categorical (value-free) Causes, following D-08-19b/D-08-14's own shape -- a precedent future whole-program or whole-function bound/topology refusals in this codebase should follow."

requirements-completed: [OWN-06]

coverage:
  - id: D1
    description: "loanLivenessFixpoint's cycle pre-walk uses an explicit stack, not native recursion, demonstrated on a 50,000-deep synthetic chain and structurally proven absent of any self-calling closure"
    requirement: "OWN-06"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCFGBackEdgeWalkIsIterative"
        status: pass
    human_judgment: false
  - id: D2
    description: "check.cfg_back_edge is a coded diagnostic.Diagnostic carrying the offending block ID as a cycle_block cause, not a bare formatted error string"
    requirement: "OWN-06"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestBackEdgeRejected"
        status: pass
    human_judgment: false
  - id: D3
    description: "loanLivenessFixpoint terminates under a derived fail-closed bound and produces the named check.loan_liveness_bound_exceeded refusal (never a hang, never a truncated live-in map) when reached; the bound scales with block count"
    requirement: "OWN-06"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestLoanLivenessBoundExceededRefusesRatherThanTruncates"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestLoanLivenessBoundScalesWithBlockCount"
        status: pass
    human_judgment: false
  - id: D4
    description: "The bound's numeric value appears in neither the diagnostic's Causes nor its Message, so retuning the bound cannot move any existing diagnostic ID"
    requirement: "OWN-06"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestLoanLivenessBoundValueIsNotInDiagnosticIdentity"
        status: pass
    human_judgment: false
  - id: D5
    description: "The bound control is mutation-killed through the unexported loanLivenessBoundSeam, flipped and deferred-restored by a same-package test, proven in both directions"
    requirement: "OWN-06"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestLoanLivenessBoundMutationKilled"
        status: pass
    human_judgment: false
  - id: D6
    description: "PHASE-08-DEBT.md's D-08-15 detail section names the shipped seam and mutation-kill test, and TestDebtRegistersAreWellFormed stays green with the items count and table shape unchanged"
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session/... -run TestDebtRegistersAreWellFormed"
        status: pass
    human_judgment: false
  - id: D7
    description: "No existing verdict moved and no regression was introduced: the whole pre-existing suite plus go vet stay green (excluding the pre-existing, unrelated session-package spike-registry gap), including under -count=2 -shuffle=on"
    verification:
      - kind: unit
        ref: "go test ./... && go vet ./..."
        status: pass
      - kind: unit
        ref: "go test ./internal/compiler/check/... -count=2 -shuffle=on"
        status: pass
    human_judgment: false

# Metrics
duration: ~55min
completed: 2026-09-10
status: complete
---

# Phase 08 Plan 04: The Derived Loan-Liveness Bound and Iterative CFG Pre-Walk Summary

**The only genuinely iterative fixpoint in Phase 08 -- `loanLivenessFixpoint`'s worklist -- now terminates under a derived, block-count-scaled, fail-closed bound with a named, identity-stable refusal, its cycle pre-walk converted from native recursion to an explicit-stack DFS ported from `callgraph.Order`, and the bound control mutation-killed through a seeded seam that has actually been seen to fail.**

## Performance

- **Duration:** ~55 min
- **Started:** 2026-09-09T22:00 (approx, following 08-03's completion commit)
- **Completed:** 2026-09-10
- **Tasks:** 3 completed (all `type="auto"`, one `tdd="true"`)
- **Files modified:** 3 (`internal/compiler/check/check.go`, `internal/compiler/check/check_test.go`, `PHASE-08-DEBT.md`)

## Accomplishments

- **Task 1 -- explicit-stack cycle pre-walk, coded refusal.** Replaced `loanLivenessFixpoint`'s self-calling `var walk func(id string) error` closure with `detectCFGCycle`, an iterative three-colour (white/gray/black) DFS over an explicit `cfgWalkFrame` stack, structurally ported from `callgraph.Order` (D-07-18/D-08-19a). `loanLivenessFixpoint`'s own signature changed from returning a bare `error` to `*diagnostic.Diagnostic`; the cycle refusal is now `check.cfg_back_edge`, built by the new `cfgBackEdgeDiagnostic` with a single `cycle_block` cause naming the offending block, replacing the old `fmt.Errorf`-wrapped string (D-08-19b). `TestCFGBackEdgeWalkIsIterative` go/ast-scans the function body to assert no self-calling closure remains and exercises the walk on a synthetic 50,000-deep chain to prove the conversion is load-bearing under depth, not merely structurally absent.
- **Task 2 -- the derived bound and its named refusal.** Added `loanLivenessBoundFactor` (4) and `loanLivenessBound(blocks, loans) = factor * blocks * (loans + 1)`, checked against the worklist's own running `work` total inside the loop; reaching it returns `check.loan_liveness_bound_exceeded` and a zero-valued `loanLivenessResult`, never a truncated live-in map. The unexported `loanLivenessBoundSeam` (false in production) forces the trip for tests. The doc comment states the bound's justification in the required terms: an internal-consistency assertion over a provably-convergent monotone finite-lattice fixpoint, not a DoS defense, with the overflow non-contract this plan's own planner assumption requires.
- **Task 3 -- mutation-kill and identity stability.** `TestLoanLivenessBoundMutationKilled` follows `TestCallReturnTypeDerivationMutationKilled`'s exact defer-before-flip shape, proving both directions: refusal with the seam up, clean admission with it down. `TestLoanLivenessBoundValueIsNotInDiagnosticIdentity` constructs the refusal twice with differing computed bounds (via block count) and proves identical diagnostic IDs, with neither run's own bound value appearing as a substring in the Message or any Cause. `PHASE-08-DEBT.md`'s D-08-15 detail section now names the actually-shipped witness (`loanLivenessBoundSeam`, `TestLoanLivenessBoundMutationKilled`) rather than planning-time placeholder wording, with the register's `items:` count and table shape unchanged.

## Task Commits

Each task was committed atomically:

1. **Task 1: Explicit-stack cycle pre-walk and a coded `check.cfg_back_edge`** - `63246d9` (feat)
2. **Task 2: The derived fail-closed bound and its named refusal** - `9a668d1` (feat)
3. **Task 3: Mutation-kill the bound and prove its value cannot move a diagnostic ID** - `e3b3be7` (test)

**Plan metadata:** commit pending (this SUMMARY + STATE.md/ROADMAP.md/REQUIREMENTS.md)

_Note: `workflow.tdd_mode` is `false` for this project (config.json); Task 2 was marked `tdd="true"` in the plan, but per this project's established convention (see 08-03-SUMMARY.md's identical note), the `<behavior>`/`<implementation>` split was developed and verified together within Task 2's single commit rather than as separate RED/GREEN commits -- consistent with how every other Phase 08 plan has already handled `tdd="true"` under this project's `tdd_mode: false` setting._

## Files Created/Modified

- `internal/compiler/check/check.go` - `cfgWalkFrame`, `detectCFGCycle`, `cfgBackEdgeDiagnostic`, `loanLivenessBoundFactor`, `loanLivenessBound`, `loanLivenessBoundSeam`, `loanLivenessBoundExceededDiagnostic`; `loanLivenessFixpoint`'s signature and body (explicit-stack pre-walk, bound check, `*diagnostic.Diagnostic` return); all four production call sites updated
- `internal/compiler/check/check_test.go` - `TestCFGBackEdgeWalkIsIterative`, `TestLoanLivenessBoundScalesWithBlockCount`, `TestLoanLivenessBoundExceededRefusesRatherThanTruncates`, `TestLoanLivenessBoundMutationKilled`, `TestLoanLivenessBoundValueIsNotInDiagnosticIdentity`; `TestBackEdgeRejected` widened to assert on the diagnostic's Code/cause; all seven pre-existing test call sites updated to the new signature
- `.planning/phases/08-interprocedural-loan-liveness-in-check/PHASE-08-DEBT.md` - D-08-15 detail section updated to name the landed seam and test

## Decisions Made

See `key-decisions` in frontmatter for the three load-bearing decisions this plan made while landing the bound: (1) the `+1` floor correction to the literal bound formula (Rule 1 auto-fix, verified against the full existing suite); (2) the categorical (value-free) `block_count` cause shape, required by the plan's own identity-stability acceptance criterion; (3) threading a caller-supplied `diagnostic.Span` through `loanLivenessFixpoint` rather than deriving one internally.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] The literal `blockCount * distinctLoanCount` bound formula computes zero for any loan-free function, refusing nearly every legal program**
- **Found during:** Task 2, while deriving the bound formula and reasoning through its behavior on the existing test corpus before committing
- **Issue:** The plan's action text specifies `loanLivenessBoundFactor * len(blocks) * distinctLoanCount`. `loanLivenessFixpoint` runs unconditionally for every checked function via `checkBranch` and `checkInterproceduralLoanLiveness`, including every function that borrows nothing at all. For such a function, `distinctLoanCount == 0`, so the literal formula computes a bound of exactly zero. Since the worklist loop still costs one transfer-function evaluation per block regardless of loan involvement, `work` becomes nonzero after the very first block is processed, and any function requiring more than one worklist pass (i.e. essentially every multi-block, loan-free function processed via `checkBranch`) would trip the bound refusal on its own next iteration -- directly violating the plan's own `must_haves.truths` claim that "every existing fixture and every testdata/phase08 fixture stays well under the bound; no legal program trips it."
- **Fix:** Changed the formula to `loanLivenessBoundFactor * blockCount * (distinctLoanCount + 1)`. The `+1` preserves the doubling-block-count-doubles-the-bound property (a constant additive term inside a product linear in `blockCount` still scales linearly with `blockCount`) and, for the single-block reborrow-chain shape `TestLivenessWorkScale`/`TestReborrowChainWorkIsLinear` already independently pin at `4*n+4`, reduces to exactly that same ceiling -- confirming the fix against a pre-existing, independently-derived bound rather than merely asserting a patched one.
- **Files modified:** `internal/compiler/check/check.go`
- **Verification:** `go test ./internal/compiler/check/...` (full package, including all pre-existing fixtures and `TestLivenessWorkScale`/`TestReborrowChainWorkIsLinear` at n=10,000) passes; `go test ./... && go vet ./...` shows only the pre-existing, unrelated `TestQLT01RegistryCoversAllFiveSpikes` failure.
- **Committed in:** `9a668d1`

---

**Total deviations:** 1 auto-fixed (1 Rule 1 -- a genuine correctness gap in the plan's literal formula, caught before commit by reasoning through the existing test corpus rather than discovered as a later regression; no scope creep beyond the fix itself, and the fix preserves every acceptance criterion the plan's literal text also required, including the doubling-block-count property).
**Impact on plan:** Necessary for the bound to satisfy its own stated acceptance criterion ("no legal program trips it"); does not change the bound's derivation rationale, its identity-stability guarantee, or the seam's shape.

## Issues Encountered

None beyond the deviation above -- caught by design-time reasoning and confirmed by running the full existing suite (not merely the plan's own new tests) before each task's commit, per this project's standing process rule to interrogate what input space a change actually reaches rather than trusting a green suite in isolation.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Success criterion 2 is fully met: the only genuinely iterative fixpoint in Phase 08 terminates under a derived, auditable, fail-closed bound, producing a named refusal rather than a hang or a silent under-approximation.
- The bound control has been mutation-killed in both directions, and its own numeric value is proven incapable of moving a published diagnostic ID.
- The cycle pre-walk cannot exhaust the native Go stack, demonstrated (not merely argued) at 50,000-deep synthetic input.
- `PHASE-08-DEBT.md`'s D-08-15 now names the concrete shipped witness for any future auditor or Phase 09 reader.
- No blockers for 08-05 or 08-06.

## Self-Check: PASSED

- FOUND: internal/compiler/check/check.go
- FOUND: internal/compiler/check/check_test.go
- FOUND: .planning/phases/08-interprocedural-loan-liveness-in-check/PHASE-08-DEBT.md
- FOUND commit: 63246d9
- FOUND commit: 9a668d1
- FOUND commit: e3b3be7
- Re-ran Task 1 acceptance criteria: `go test ./internal/compiler/check/... -run 'BackEdge|CFGBackEdge|LoanLivenessFixpoint' -v` -- all PASS
- Re-ran Task 2 acceptance criteria: `go test ./internal/compiler/check/... -run 'LoanLivenessBound' -v` -- all PASS; `go test ./internal/compiler/check/... -count=2 -shuffle=on` -- PASS
- Re-ran Task 3 acceptance criteria: `go test ./internal/compiler/check/... -run 'LoanLivenessBoundMutationKilled|LoanLivenessBoundValueIsNotInDiagnosticIdentity' -v` -- all PASS; `go test ./internal/compiler/session/... -run TestDebtRegistersAreWellFormed -v` -- all PASS
- Re-ran plan-level `<verification>`: `go test ./internal/compiler/check/... -count=2 -shuffle=on` exits 0; `go test ./internal/compiler/session/... -run TestDebtRegistersAreWellFormed` exits 0; `go test ./... && go vet ./...` -- only the pre-existing, unrelated `TestQLT01RegistryCoversAllFiveSpikes` failure (confirmed present before this phase began) -- PASS

---
*Phase: 08-interprocedural-loan-liveness-in-check*
*Completed: 2026-09-10*
