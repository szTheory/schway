---
phase: 13-agent-loop-for-interprocedural-defects
plan: 01
subsystem: check
tags: [diagnostic-repairs, interprocedural, ownership, lang-repair, json-protocol]

# Dependency graph
requires:
  - phase: 08-interprocedural-loan-liveness-in-check
    provides: check.interprocedural_loan_liveness's original diagnostic.Error emission (D-08-20..25) and derivePlaceLoans' forward/backward direction classification
  - phase: 06-agent-feedback-and-performance-ratification
    provides: diagnostic.ErrorWithRepairs, the driver-eligibility contract (DriverEligible/NormalizeApplicability), and cmd/lang-repair's kind-agnostic single-pass driver
provides:
  - A real, driver-verified repair (move_after_interprocedural_loan) for one of DX-07's three required interprocedural defect classes
  - A direction-aware safety gate (callIsLastUse) proving the repair is only emitted when the swap is actually semantics-preserving
  - testdata/phase13/ corpus root with its README contract and first derivation fixture
  - .planning/phases/13-agent-loop-for-interprocedural-defects/13-ANTITHEATER-CONTRACT.md, closing RESEARCH.md's single LOW-confidence area
affects: [13-02, 13-03, 13-04, 13-05, 13-06, 13-07]

# Actuals (#2632)
actuals:
  tokens: 9709
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "diagnostic.Error -> diagnostic.ErrorWithRepairs conversion, unconditional even on the zero-repair fallback path, to make a schema-churn consequence (D-13-09a) deliberate and reviewable rather than accidental"
    - "Direction-aware repair gating: a diagnostic emission site that can be reached via two structurally different derivations may only be repair-eligible on ONE of them; empirically verify (splice-and-recheck), never assume symmetry"

key-files:
  created:
    - testdata/phase13/README
    - testdata/phase13/derivation_interprocedural_loan_defect.lang
    - .planning/phases/13-agent-loop-for-interprocedural-defects/13-ANTITHEATER-CONTRACT.md
  modified:
    - internal/compiler/check/check.go
    - internal/compiler/check/check_ordering_stability_test.go
    - internal/compiler/syntax/parser.go
    - cmd/lang-repair/repair_test.go

key-decisions:
  - "The move_after_interprocedural_loan repair is emitted ONLY for the BACKWARD direction (call is the loan's own recorded last use), never the FORWARD direction (loan propagated through the call onto a place read still later) -- empirically verified by splicing the swap onto both twin_a_refuse.lang-shaped and twin_b_refuse.lang-shaped fixtures: the backward swap re-checks clean, the forward swap does not (the diagnostic still fires, identically). This was NOT explicit in 13-CONTEXT.md/13-PATTERNS.md and required discovering the distinction empirically before shipping the repair."
  - "check_ordering_stability_test.go re-pins FIVE rows, not the plan's stated four -- phase08/twin_b_accept.lang also carries check.interprocedural_loan_liveness and churns identically from the same unconditional schema switch. The plan's own text omitted it; corrected here rather than silently narrowing the must_haves truth to match the plan."
  - "internal/compiler/syntax/parser.go's call-binding Binding.Span (Rule 1 bugfix, not in the plan's files_modified) was truncated at the callee token, never including the call's own closing paren -- latent until check.go's new \":stmt\" span-widening needed the whole statement span. Fixed at the source since a full `go test ./...` run before and after confirmed no existing pinned diagnostic ID depends on the old, truncated value."

requirements-completed: [DX-07]

coverage:
  - id: D1
    description: "check.interprocedural_loan_liveness carries a machine-applicable move_after_interprocedural_loan repair for the backward-direction case, spliceable through the JSON protocol alone"
    requirement: "DX-07"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestRelayEscortWitnessRefusesInterproceduralLiveness"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_ordering_stability_test.go#TestInterproceduralDiagnosticOrderingStability"
        status: pass
      - kind: integration
        ref: "cmd/lang-repair/repair_test.go#TestRepairDriverFixesInterproceduralLoanLivenessSinglePass"
        status: pass
    human_judgment: false
  - id: D2
    description: "cmd/lang-repair driver source is unmodified across this plan (D-13-32)"
    requirement: "DX-07"
    verification:
      - kind: other
        ref: "git diff --quiet -- cmd/lang-repair/repair.go cmd/lang-repair/main.go"
        status: pass
    human_judgment: false
  - id: D3
    description: "cmd/lang-repair/antitheater_test.go's obligations for a new defect class are written down before any new class is designed"
    requirement: "DX-07"
    verification: []
    human_judgment: true
    rationale: "The contract document's completeness and accuracy against a 794-line file is a judgment call best confirmed by a human reviewer familiar with the anti-theater guard design, not something a test can assert."

# Metrics
duration: 27min
completed: 2026-09-13
status: complete
---

# Phase 13 Plan 01: Interprocedural-Loan-Liveness Repair Tracer Summary

**`check.interprocedural_loan_liveness` now carries a real, driver-verified `move_after_interprocedural_loan` repair for the backward-direction case, proven end-to-end through `cmd/lang-repair`'s untouched JSON-protocol driver, with a direction-aware safety gate that refuses the repair on the forward-direction case where the swap would not actually fix anything.**

## Performance

- **Duration:** 27 min
- **Started:** 2026-09-13T18:23:45Z (approx, prior session boundary)
- **Completed:** 2026-09-13T18:50:44Z
- **Tasks:** 3
- **Files modified:** 7 (4 modified, 3 created)

## Accomplishments
- Converted `check.interprocedural_loan_liveness` from `diagnostic.Error` to `diagnostic.ErrorWithRepairs` (D-13-09a's deliberate, unconditional schema `/0 -> /1` switch)
- Discovered and correctly handled a real correctness hazard: the "swap the move and call statements" repair is only semantics-preserving for the BACKWARD direction (call is the loan's own last use); the FORWARD direction (loan propagated through the call onto a later read) is NOT fixed by the swap, verified empirically against real fixtures both ways
- Extended `check.go`'s `:stmt` span-widening (previously take/borrow-only) to call bindings, and fixed a latent parser bug (`Binding.Span` truncated at the callee token for call bindings) that the new span channel exposed
- Proved the repair end-to-end through the real, unmodified `cmd/lang-repair` driver on a real corpus fixture
- Read `cmd/lang-repair/antitheater_test.go` in full and recorded its obligations for future interprocedural defect classes in `13-ANTITHEATER-CONTRACT.md`
- Re-pinned five `check_ordering_stability_test.go` rows (one more than the plan named) with the actual regenerated IDs, verified via a full corpus scan that no other row changed

## Task Commits

Each task was committed atomically:

1. **Task 1: End-to-end `move_after_interprocedural_loan`** - `da19a7f` (feat)
2. **Task 2: Write down `antitheater_test.go`'s obligations** - `88d3ee3` (docs)
3. **Task 3: Regression for the repair driver path** - `c4ee457` (test)

**Plan metadata:** (this commit)

## Files Created/Modified
- `internal/compiler/check/check.go` - `interproceduralLoanLivenessDiagnostic` now builds via `ErrorWithRepairs`, gated to the backward direction (`callIsLastUse`); adds `unionSpan`/`calleeNameFromID` helpers; both call-binding sites now record a `":stmt"` span entry
- `internal/compiler/check/check_ordering_stability_test.go` - five `check.interprocedural_loan_liveness` rows re-pinned with regenerated IDs and explanatory comments
- `internal/compiler/syntax/parser.go` - call-binding `Binding.Span` now covers the whole statement (through the closing paren), not just the callee token
- `cmd/lang-repair/repair_test.go` - `TestRepairDriverFixesInterproceduralLoanLivenessSinglePass` added
- `testdata/phase13/README` - corpus prefix contract, carried forward from `testdata/phase6/README`
- `testdata/phase13/derivation_interprocedural_loan_defect.lang` - the tracer's real, backward-direction fixture
- `.planning/phases/13-agent-loop-for-interprocedural-defects/13-ANTITHEATER-CONTRACT.md` - obligations extracted from `antitheater_test.go`

## Decisions Made
- Gated the repair to the backward direction only (`callIsLastUse` parameter), after empirically proving the forward-direction swap does not resolve the conflict on real fixtures (`testdata/phase07/relay_escort_witness.lang`, `testdata/phase08/twin_a_refuse.lang`, `testdata/phase08/relay_depth2_refuse.lang` correctly emit zero repairs; `testdata/phase08/twin_b_accept.lang`, `testdata/phase08/twin_b_refuse.lang` correctly emit a repair that splices clean)
- Re-pinned `phase08/twin_b_accept.lang` in addition to the plan's named four rows, since it independently carries `check.interprocedural_loan_liveness` and churns from the same unconditional schema switch
- Fixed the parser's call-binding `Binding.Span` truncation as a Rule 1 bugfix rather than working around it in `check.go`, after confirming via a full `go test ./...` run (before and after) that no other pinned diagnostic ID depends on the old value

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Gated the repair to the backward direction only**
- **Found during:** Task 1, verifying the repair against real fixtures
- **Issue:** The plan's action text and PATTERNS.md described "both statements re-emitted in swapped order" without distinguishing the two directions `checkInterproceduralLoanLiveness` can reach this diagnostic through (forward: loan propagated through the call onto a later read; backward: the call is itself the loan's recorded last use). An unconditional implementation emits a `MachineApplicable` repair on `testdata/phase07/relay_escort_witness.lang` (a plan-cited read_first fixture) that, when applied, still refuses with the identical diagnostic -- a repair that does not repair, exactly the failure D-13-11's safety argument exists to prevent.
- **Fix:** Added a `callIsLastUse bool` parameter to `interproceduralLoanLivenessDiagnostic`, set `true` only at the backward-direction call site (check.go, where `candidate := function.Linear.Operations[lastUse]`), `false` at the forward-direction call site (`chain.extendedByCall`). The repair is now gated on this flag in addition to the existing span/name-availability checks.
- **Files modified:** internal/compiler/check/check.go
- **Verification:** Manually spliced the repair onto `relay_escort_witness.lang`, `twin_a_refuse.lang` (forward, unfixed -- now correctly emit zero repairs) and `twin_b_accept.lang`, `twin_b_refuse.lang` (backward, fixed -- verified clean re-check both before and after the `callIsLastUse` gate); `TestRelayEscortWitnessRefusesInterproceduralLiveness`'s pre-existing "expected zero repairs" assertion now passes for the correct reason.
- **Committed in:** da19a7f (Task 1 commit)

**2. [Rule 1 - Bug] Widened `syntax` parser's call-binding statement span**
- **Found during:** Task 1, implementing the repair's `:stmt` span channel
- **Issue:** `internal/compiler/syntax/parser.go`'s `linearBody` set a call binding's `Binding.Span` to `spanFrom(bindingStart, source)`, where `source` is the CALLEE token (parsed before the `(` is even seen) -- so the span stopped short of the arguments and closing paren, unlike every other binding kind (take/borrow/borrow_mut/copy), whose `source` token IS the RHS's own last token. This produced a truncated repair `Span` whose `Replacement` splice corrupted the source (observed directly: `let delivered = take buffer(borrowed)` after a bad splice).
- **Fix:** Changed the call-binding's `Span` to `diagnostic.Span{Start: bindingStart.Span.Start, End: end}`, where `end` is the closing paren's own End (already computed by `p.callArguments()`, already used two lines below for `body.Span.End`).
- **Files modified:** internal/compiler/syntax/parser.go
- **Verification:** A full `go test ./...` run before and after this change produces byte-identical results outside `check.interprocedural_loan_liveness`'s own five expected-to-churn rows, confirming no other pinned diagnostic ID depends on the old, truncated span value. The repaired fixture splices and re-checks clean.
- **Committed in:** da19a7f (Task 1 commit)

**3. [Rule 1 - Bug] Re-pinned a fifth `check_ordering_stability_test.go` row (`phase08/twin_b_accept.lang`)**
- **Found during:** Task 1, running the full ordering-stability corpus scan
- **Issue:** The plan's must_haves truth and acceptance criteria stated "exactly four" rows of code `check.interprocedural_loan_liveness` would churn. A full corpus scan found FIVE such rows in the pinned table; `phase08/twin_b_accept.lang` (already pinned with this exact code, per plan 09-09's own historical note) churns identically from the same unconditional `/0 -> /1` schema switch, independent of the `callIsLastUse` gate (confirmed by re-running the scan with the parser fix reverted).
- **Fix:** Re-pinned all five rows with their regenerated IDs and an explanatory comment on each; verified via a full-corpus mismatch scan that no row of any OTHER code changed.
- **Files modified:** internal/compiler/check/check_ordering_stability_test.go
- **Verification:** `TestInterproceduralDiagnosticOrderingStability` passes; `git diff` on the file touches only the five named fixture-key lines (plus explanatory comments).
- **Committed in:** da19a7f (Task 1 commit)

---

**Total deviations:** 3 auto-fixed (3 Rule 1 bugs, all discovered while implementing Task 1 and required for the tracer's own correctness claim)
**Impact on plan:** All three fixes were necessary for the shipped repair to actually be correct (the direction gate) and applicable (the span fix), and for the plan's own must_haves truth to match reality (the fifth row). No scope creep -- nothing beyond what Task 1's own acceptance criteria required.

## Issues Encountered
None beyond the three deviations above, which are the actual substance of what this tracer plan exists to surface (13-01's objective explicitly frames it as "far cheaper to learn it here than after ten layers are committed").

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- The tracer path (`check` -> `lang --json check` -> `cmd/lang-repair`) is proven real for one interprocedural defect class, with the driver genuinely unmodified.
- The direction-asymmetry finding (repairs are only valid for the backward direction) is a load-bearing fact for plans 13-03/13-05/13-06: any future interprocedural repair emission site must be checked for the same kind of directional/structural asymmetry before shipping a `MachineApplicable` repair -- do not assume a "swap" or "reorder" repair is safe merely because it type-checks; verify by splicing and re-checking on a real fixture in both directions the diagnostic can fire from.
- `13-ANTITHEATER-CONTRACT.md` names the one mandatory registration obligation (adding a new class's name to `antitheater_test.go`'s four hardcoded class lists plus `repair_test.go`'s own `TestRepairDriverFixesEveryDefectClassSinglePass` list) for any future class that wants that coverage; plans 13-03/13-05/13-06 should consult it before designing new repair-kind classes.
- No blockers for 13-02.

## Self-Check: PASSED

- `internal/compiler/check/check.go`, `internal/compiler/check/check_ordering_stability_test.go`, `internal/compiler/syntax/parser.go`, `cmd/lang-repair/repair_test.go` all exist and carry the described changes (`git show da19a7f`, `c4ee457`).
- `testdata/phase13/README`, `testdata/phase13/derivation_interprocedural_loan_defect.lang` exist on disk.
- `.planning/phases/13-agent-loop-for-interprocedural-defects/13-ANTITHEATER-CONTRACT.md` exists on disk.
- `git log --oneline --all --grep="13-01"` returns 3 commits (da19a7f, 88d3ee3, c4ee457).
- All plan-level `<verification>` commands re-run clean: `go build ./... && go vet ./internal/compiler/check/...`; `go test ./...`; `TestInterproceduralDiagnosticOrderingStability` (5 re-pinned rows, documented); end-to-end CLI `lang-repair` reports `repaired` and the repaired file re-checks clean; `git diff --quiet -- cmd/lang-repair/repair.go cmd/lang-repair/main.go` exits 0.

---
*Phase: 13-agent-loop-for-interprocedural-defects*
*Completed: 2026-09-13*
