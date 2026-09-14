---
phase: 10-trusted-interprocedural-oracle
plan: 06
subsystem: process
tags: [debt-register, mid-phase-gate, cost-measurement, doc-correction, scope-cut-trigger]

# Dependency graph
requires:
  - phase: 10-trusted-interprocedural-oracle
    provides: "plans 10-01 through 10-05 (the originvalidate/pathoracle/interp-call-stack parallel wave), whose actual token cost this gate measures against their own declared estimates"
provides:
  - "D-09-53's diagnosis reversed and re-filed: deriveFunctionUsesParam's doc comment and testdata/phase08/twin_b_accept.lang's header now state the implemented zero-hop-only contract, with the real gap re-filed as D-10-28 (corevalidate has no interprocedural usesParam peer), zero logic change"
  - "D-10-59's scope-cut trigger measured as a token-cost ratio (0.19x) and recorded in PHASE-10-DEBT.md with an explicit DID NOT FIRE verdict"
  - "All four hard ordering constraints confirmed against git history: SEM-09 before the freeze, this gate before criterion 4, originvalidate's widen-then-use as two commits, each differential landing with its own peer"
  - "A green full suite (go test ./...) and a green race suite over interp/pathoracle/originvalidate/corevalidate, gating entry to plans 10-07 and 10-08"
affects: [10-07, 10-08, 10-09]

# Actuals (#2632)
actuals:
  tokens: 9200
  tasks: 3
  commits: 2

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Debt-register amendment in place (adding a dated paragraph and evidence table to an existing D-NN-xx row's Detail section, plus updating its Landing-phase cell) rather than a new row, mirroring D-09-43's precedent for a trigger measurement"

key-files:
  created: []
  modified:
    - internal/compiler/check/check.go
    - internal/compiler/check/check_ordering_stability_test.go
    - testdata/phase08/twin_b_accept.lang
    - .planning/phases/10-trusted-interprocedural-oracle/PHASE-10-DEBT.md

key-decisions:
  - "The diagnostic-ordering-stability baseline id for twin_b_accept.lang shifted after the header rewrite (the id is content-hash derived, not code-derived) -- updated with an inline justification per that test's own stated contract, rather than treated as a code regression, since the CODE (check.interprocedural_loan_liveness) stayed byte-identical"
  - "No cut taken at the checkpoint gate: the measured 0.19x ratio is far under the ~2x trigger and all four ordering constraints held, so 'Proceed unchanged' (the plan's documented default) is correct on the merits, not merely auto-selected"

patterns-established: []

requirements-completed: [SEM-08, SEM-09, TRU-02, TRU-03]

coverage:
  - id: D1
    description: "deriveFunctionUsesParam's doc comment and twin_b_accept.lang's header state the implemented contract (zero-hop-only OpReturn exemption) rather than the falsified intended-split narrative, with zero logic change"
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/check/... ./internal/compiler/corevalidate/... ./internal/compiler/session/..."
        status: pass
      - kind: other
        ref: "git diff -U0 internal/compiler/check/check.go (comment-only)"
        status: pass
    human_judgment: false
  - id: D2
    description: "D-10-59's token-cost trigger is measured (0.19x) and recorded with an explicit DID NOT FIRE verdict"
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session/... -run TestDebtRegistersAreWellFormed -v"
        status: pass
      - kind: other
        ref: "PHASE-10-DEBT.md D-10-59 amendment (sums, ratio, date, verdict)"
        status: pass
    human_judgment: false
  - id: D3
    description: "All four hard ordering constraints confirmed against git history"
    verification:
      - kind: other
        ref: "git log --oneline -- internal/compiler/originvalidate/originvalidate.go; ls .planning/phases/10-trusted-interprocedural-oracle/*-SUMMARY.md"
        status: pass
    human_judgment: false
  - id: D4
    description: "The full suite and the race suite over the four modified packages are green, admitting criterion 4's differential work unchanged"
    verification:
      - kind: integration
        ref: "go test ./..."
        status: pass
      - kind: integration
        ref: "go test -race ./internal/compiler/interp/... ./internal/compiler/pathoracle/... ./internal/compiler/originvalidate/... ./internal/compiler/corevalidate/..."
        status: pass
    human_judgment: false

duration: 40min
completed: 2026-09-11
status: complete
---

# Phase 10 Plan 6: Mid-Phase Gate — Cost Measurement, Ordering Confirmation, and D-09-53 Reversal Summary

Phase 10's actual token cost measured at 0.19x its declared per-plan baseline (71,773 actual vs 375,000 estimated across plans 10-01–10-05) — the D-10-59 scope-cut trigger DID NOT FIRE — and the D-09-53 doc-comment/fixture-header correction landed with zero logic change, clearing plans 10-07 and 10-08 to proceed unchanged.

## Performance

- **Duration:** 40 min
- **Started:** 2026-09-11
- **Completed:** 2026-09-11
- **Tasks:** 3 completed (Task 3 is a checkpoint:decision)
- **Files modified:** 4

## Accomplishments

- Landed the D-10-29 deliverable: `deriveFunctionUsesParam`'s doc comment (check.go) now states the contract the switch actually implements — the `case core.OpReturn:` exemption covers a **zero-hop direct return of the parameter place only**, never an arbitrary identity chain (e.g. `taken = take buffer; taken`) that merely terminates in a return. `testdata/phase08/twin_b_accept.lang`'s header states the OBSERVED verdict (refused identically to `twin_b_refuse.lang`, via `check.interprocedural_loan_liveness`) and why, replacing the falsified intended-split narrative, and re-files the real gap as D-10-28 (`corevalidate` has no interprocedural `usesParam` peer at all). Zero logic change: `git diff -U0 internal/compiler/check/check.go` shows only comment lines added/removed.
- Measured the D-10-59 scope-cut trigger as a token-cost ratio: summed `estimate.tokens` across plans 10-01 through 10-05 (375,000) against each plan's own `actuals.tokens` (71,773) = **0.19x**, recorded in `PHASE-10-DEBT.md`'s `### D-10-59` section with the two sums, the ratio, the date, and an explicit **DID NOT FIRE** verdict, following D-09-43's precedent form. `items:` count is unchanged (12) since no cut row was added.
- Confirmed all four hard ordering constraints against git history and the phase directory's own state (not against plan documents): (1) plan 10-05 (SEM-09) has a SUMMARY, plan 10-09 (the stability freeze) does not yet — SEM-09 landed first; (2) plan 10-08 (criterion 4's differential) has no SUMMARY yet — this gate precedes it; (3) `git log --oneline -- internal/compiler/originvalidate/originvalidate.go` shows two distinct commits from plan 10-02 (`bb4aa02` widen, `6648b72` use); (4) plan 10-02's origin-walk differential and plan 10-03's path-composition differential each land their own peer's test in the same plan, per their own SUMMARYs.
- The checkpoint gate ran both required verifications green: `go test ./...` (0 failures across every package) and `go test -race` over `interp`/`pathoracle`/`originvalidate`/`corevalidate` (no DATA RACE). With the ratio at 0.19x and all four constraints holding, **"Proceed unchanged" was selected** (the plan's documented default, auto-ratified under `workflow.auto_advance: true` since the checkpoint's gate is the default `"blocking"`, not `gate="blocking-human"`) — **no cut is taken**. This is stated explicitly, per the plan's own requirement that a no-cut gate not be indistinguishable from a skipped one.

## Task Commits

Each task was committed atomically:

1. **Task 1: Land the D-09-53 reversal as an artifact correction** - `3adacf1` (docs)
2. **Task 2: Measure the D-10-59 trigger and confirm the four hard ordering constraints** - `0bf8801` (docs)
3. **Task 3: Mid-phase gate checkpoint:decision** - no separate commit; auto-ratified "Proceed unchanged" under `workflow.auto_advance: true`, realized by Tasks 1-2's already-landed work and this SUMMARY's own record of the verification run

**Plan metadata:** (this commit) `docs(10-06): complete mid-phase gate plan`

## Files Created/Modified

- `internal/compiler/check/check.go` — `deriveFunctionUsesParam`'s doc comment corrected to state the zero-hop-only OpReturn exemption, citing D-10-27/D-10-29. Comment-only.
- `internal/compiler/check/check_ordering_stability_test.go` — `twin_b_accept.lang`'s baseline diagnostic id updated (content-hash shift from the fixture header rewrite, code unchanged), with an inline justification comment.
- `testdata/phase08/twin_b_accept.lang` — header comment corrected to state the observed refusal verdict and re-file D-10-28.
- `.planning/phases/10-trusted-interprocedural-oracle/PHASE-10-DEBT.md` — D-10-59's Landing-phase cell updated and its Detail section amended with the trigger measurement, evidence table, verdict, and the four confirmed ordering constraints.

## Decisions Made

See `key-decisions` in frontmatter.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Correcting `twin_b_accept.lang`'s header shifted its diagnostic baseline id**
- **Found during:** Task 1, `go test ./internal/compiler/check/...`
- **Issue:** `TestInterproceduralDiagnosticOrderingStability` failed: `phase08/twin_b_accept.lang: baseline mismatch: want id="diagnostic:5fe0501da5249e177b0ada39", got id="diagnostic:9077559bb65dc5fe33cb57bb"`. The diagnostic `code` (`check.interprocedural_loan_liveness`) was unchanged; only the content-hash-derived `id` shifted because the fixture's header comment bytes changed. This is a direct, expected consequence of a comment-only edit to a fixture the test hashes, not a logic regression.
- **Fix:** Updated the baseline table entry for `phase08/twin_b_accept.lang` to the new id, with an inline comment explaining the shift is content-hash-derived and the observed code is byte-identical, per the test's own stated contract ("a difference here must be justified in the owning plan's SUMMARY, not silently regenerated").
- **Files modified:** `internal/compiler/check/check_ordering_stability_test.go`
- **Verification:** `go test ./internal/compiler/check/... -run 'TestInterproceduralLivenessTwinPatternB|TestDeriveFunctionUsesParamBehaviors|TestCostCorpusLeafTemplatesDifferOnlyInTheCallee|TestInterproceduralDiagnosticOrderingStability' -v` and the full `go test ./internal/compiler/check/... ./internal/compiler/corevalidate/... ./internal/compiler/session/...` both exit 0, no test expectation other than this one id changed.
- **Committed in:** `3adacf1`

---

**Total deviations:** 1 auto-fixed (Rule 1).
**Impact on plan:** Necessary consequence of a comment-only fixture edit hitting a content-hash-derived test id; no code behavior moved, no scope creep.

## Issues Encountered

None beyond the deviation above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- D-10-59's trigger measured well under budget (0.19x); Phase 10 has ample headroom for plans 10-07 and 10-08's criterion-4 differential work without needing any of the three named cuts.
- All four hard ordering constraints held — no escalation needed.
- The suite and race suite are both green, satisfying the plan's own gating requirement for criterion 4's differential work to begin.
- D-10-28 (corevalidate has no interprocedural usesParam peer) remains OPEN with no assigned landing phase, per 10-CONTEXT.md's explicit planner-discretion grant — candidate homes are Phase 11 or M003, unchanged by this plan.
- Ready for plan 10-07.

## Self-Check: PASSED

- `internal/compiler/check/check.go` — FOUND
- `internal/compiler/check/check_ordering_stability_test.go` — FOUND
- `testdata/phase08/twin_b_accept.lang` — FOUND
- `.planning/phases/10-trusted-interprocedural-oracle/PHASE-10-DEBT.md` — FOUND
- Commit `3adacf1` — FOUND (`git log --oneline --all`)
- Commit `0bf8801` — FOUND
- `git diff -U0 internal/compiler/check/check.go` — comment-only (verified above)
- `go test ./internal/compiler/check/... ./internal/compiler/corevalidate/... ./internal/compiler/session/...` — PASS
- `go test ./internal/compiler/session/... -run TestDebtRegistersAreWellFormed -v` — PASS (9/9 registers, including PHASE-10-DEBT.md)
- `go test ./...` — PASS (0 failures)
- `go test -race ./internal/compiler/interp/... ./internal/compiler/pathoracle/... ./internal/compiler/originvalidate/... ./internal/compiler/corevalidate/...` — PASS, no DATA RACE
- `git log --oneline -- internal/compiler/originvalidate/originvalidate.go` — shows `bb4aa02` and `6648b72` (two commits, plan 10-02)

---
*Phase: 10-trusted-interprocedural-oracle*
*Completed: 2026-09-11*
