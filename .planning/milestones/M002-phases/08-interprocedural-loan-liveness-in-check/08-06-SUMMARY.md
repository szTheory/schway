---
phase: 08-interprocedural-loan-liveness-in-check
plan: 06
subsystem: check
tags: [mid-phase-gate, debt-register, cost-gate, ratification, requirements-closure, ownership, interprocedural]

# Dependency graph
requires:
  - phase: 08-interprocedural-loan-liveness-in-check
    plan: 01
    provides: "checkInterproceduralLoanLiveness, buildInterproceduralSummaries, the forward-canonicalization interprocedural fact entry"
  - phase: 08-interprocedural-loan-liveness-in-check
    plan: 02
    provides: "deriveFunctionUsesParam, the corrected caller-before-callee callgraph.Order traversal direction"
  - phase: 08-interprocedural-loan-liveness-in-check
    plan: 03
    provides: "The nine-fixture testdata/phase08 adversarial corpus criterion 1 is adjudicated from, plus relay_escort_witness.lang"
  - phase: 08-interprocedural-loan-liveness-in-check
    plan: 04
    provides: "The derived fail-closed loanLivenessBound and its mutation-killed refusal"
  - phase: 08-interprocedural-loan-liveness-in-check
    plan: 05
    provides: "The five-shape synthetic call-graph corpus, the growth-exponent fit, both widened gate chokepoints, and the qlt02_budget_manifest.json row this plan re-ratifies"
provides:
  - "A mandatory mid-phase gate that adjudicated criterion 1 (nine real .lang fixtures plus relay_escort_witness.lang) and criterion 3 (fitted milli-exponent 998-1003 across five call-graph shapes) from re-run, code-level evidence -- not from prose"
  - "PHASE-08-DEBT.md updated to 9 tracked items: three existing items (D-08-26, D-08-27, D-08-38) carry dated gate-review resolutions; four new items (D-08-40 through D-08-43) formalize findings the gate surfaced or the phase's prior plans left implicit"
  - "The recomputed_work_growth_exponent manifest row re-ratified at this gate's own adjudication commit (b28a92f), not at the commit that introduced the row (D-08-33)"
  - "checkInterproceduralLoanLiveness's doc comment records the liveness law's final scope: the two consulted callee-signature fields, the two-tier production/consumption split, the check.*-to-core.* Phase 09 promotion schedule, and ownership.*'s non-retirement"
  - "OWN-06 and EFF-02 confirmed complete in REQUIREMENTS.md (already marked by 08-03/08-05 ahead of this gate, per Deviations below); OWN-07/08/09 confirmed untouched"
affects: [09-peer-re-derivation-and-d-03-02-closure]

# Actuals (#2632)
actuals:
  tokens: 5915
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "A mid-phase gate adjudicates from RE-RUN code-level evidence, not from prior plans' prose: every fixture verdict and every fitted exponent in this SUMMARY was independently re-executed against the working tree at gate time, not copied from 08-03/08-05's own SUMMARYs."
    - "A cost-bound manifest row's ratified_by_commit is re-pointed at the gate's own adjudication commit via a two-commit sequence within one task: commit the debt-register edit first, then edit the manifest row to reference that already-known hash, so ratified_by_commit never has to reference its own not-yet-existing commit."

key-files:
  created: []
  modified:
    - .planning/phases/08-interprocedural-loan-liveness-in-check/PHASE-08-DEBT.md
    - internal/compiler/session/qlt02_budget_manifest.json
    - internal/compiler/check/check.go
    - .planning/REQUIREMENTS.md

key-decisions:
  - "Task 1's checkpoint:decision was auto-adjudicated under auto_advance/yolo mode (config.json: mode=yolo, workflow.auto_advance=true), matching the task's own <default> instruction: adjudicate every agenda item, record any unresolved one as debt with a named landing phase, and declare the law final only because criterion 3's measurement genuinely shipped (08-05, re-verified live at this gate)."
  - "Agenda (a) OWN-09's document conflict is NOT resolved by this gate -- the plan's own text says only a human decision closes it, and auto-mode adjudication does not manufacture that decision. Recorded as carried debt (D-08-27), with the interim rule (ownership.* not retired, interprocedural check runs last) reconfirmed through this phase's close-out."
  - "Agenda (b) criterion 4's accepted-program disclosure vehicle is NOT commissioned now -- doing so would be a Rule 4 architectural change (new schema/protocol surface), not something a mid-phase gate decides unilaterally. Carried as debt (D-08-26), landing Phase 09 alongside corevalidate's own peer work."
  - "Agenda (c) resolved: both D-08-09 admission paths (the core-operation path deciding criterion 1, the AST-shadow path deciding intraprocedural loan expiry) remain load-bearing and neither is vestigial. Retiring either before Phase 09's independent peer exists would repeat RETROSPECTIVE Key Lesson 2 -- explicitly the same reasoning D-08-27's own interim rule already applies to ownership.* wholesale."
  - "Agenda (d) resolved: both planner assumptions (the bound's unspecified integer-overflow contract, the growth-exponent milli-unit rounding tie-break) are ratified as-is. The overflow assumption is already recorded in loanLivenessBoundFactor's own doc comment with an explicit reopen trigger (iteration/loops/collections/arity>1); the rounding contract is pinned by TestGrowthExponentRoundingBoundary (08-05). Neither needed a new debt-register entry -- both are already load-bearing, tested or documented, design decisions rather than open questions."
  - "The scope-cut trigger (D-08-38) did NOT fire: 08-04 landed at its own estimated scope and 08-05's cost-gate instrument work shipped in full. Resolved at this gate, not carried further."
  - "Four new debt items were added, none previously tracked as their own register row: D-08-40 (the two accepted-program peer-divergence fixtures, formalizing what 08-03's SUMMARY only noted in prose), D-08-41 (Pattern B's real-fixture scope limit on criterion 1, likewise previously only a SUMMARY note), D-08-42 (stale callgraph.Order direction in 08-CONTEXT.md/08-01 -- code is correct, two planning docs are not), D-08-43 (the pre-existing .planning/spikes registry gap behind the one non-Phase-08 test failure, so it reads as tracked rather than unexplained)."
  - "OWN-06 and EFF-02 were ALREADY marked complete in REQUIREMENTS.md before this plan ran -- by 08-03's commit (9ad04ce) and 08-05's commit (9cfff62) respectively, both ahead of this gate. This appears to be a limitation of the shared-ID gate (gsd-tools requirements.ready-ids): it scans sibling *-PLAN.md files present in the phase directory at scan time, and 08-06-PLAN.md likely did not exist yet when 08-03 ran. This plan does not revert those marks (they are now correct in substance, since the gate did adjudicate favorably), but documents the ordering violation of the plan's own must_haves truth (\"OWN-06 and EFF-02 are marked complete in REQUIREMENTS.md only after the gate has adjudicated\") as a deviation rather than silently treating Task 3(b) as fully executed-as-designed."

requirements-completed: [OWN-06, EFF-02]

coverage:
  - id: D1
    description: "The mandatory mid-phase gate adjudicated criterion 1 (per-fixture verdicts for all nine testdata/phase08 fixtures plus relay_escort_witness.lang) and criterion 3 (fitted milli-exponent per shape, ratio-stability deltas) from re-run code-level evidence"
    requirement: "OWN-06"
    verification:
      - kind: e2e
        ref: "go run ./cmd/lang --json check testdata/phase08/*.lang testdata/phase07/relay_escort_witness.lang (all 10 fixtures re-run live at gate time; see Accomplishments for the full verdict table)"
        status: pass
      - kind: unit
        ref: "go test ./internal/compiler/check/... -run GrowthExponentInOps -v (re-run live; milli-exponents 998-1003 across all five shapes, recomputed by direct log-log OLS fit over the printed ops/work pairs)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Every gate agenda item (a/b/c/d) carries an explicit, recorded disposition -- resolved with the resolution, or recorded as debt with a named landing phase -- none left assumed-safe"
    requirement: "OWN-06"
    verification:
      - kind: other
        ref: "PHASE-08-DEBT.md dated resolution paragraphs in D-08-26, D-08-27, D-08-38, plus this SUMMARY's key-decisions"
        status: pass
    human_judgment: false
  - id: D3
    description: "The recomputed_work_growth_exponent manifest row's ratified_at/ratified_by_commit name the gate's own adjudication commit (b28a92f), not the commit that introduced the row (f642ae1)"
    requirement: "EFF-02"
    verification:
      - kind: other
        ref: "internal/compiler/session/qlt02_budget_manifest.json (ratified_by_commit: b28a92f927e40f5c0c6279dfef1368c660fe75c5)"
        status: pass
      - kind: unit
        ref: "go test ./internal/compiler/session/... -run 'TestDebtRegistersAreWellFormed|QLT02|BudgetAudit' -v"
        status: pass
    human_judgment: false
  - id: D4
    description: "The liveness law is declared final in code (checkInterproceduralLoanLiveness's doc comment) because the cost-gate instrument work was not cut and criterion 3's measurement genuinely shipped; OWN-06 and EFF-02 read complete in REQUIREMENTS.md; OWN-07/08/09 remain untouched"
    requirement: "OWN-06"
    verification:
      - kind: other
        ref: "internal/compiler/check/check.go (checkInterproceduralLoanLiveness doc comment)"
        status: pass
      - kind: other
        ref: ".planning/REQUIREMENTS.md (OWN-06 [x]/Complete, EFF-02 [x]/Complete, OWN-07/08/09 [ ]/Pending, unmodified by this plan)"
        status: pass
    human_judgment: false
  - id: D5
    description: "Full-suite confirmation: go test ./... && go vet ./... stay green except the one pre-existing, unrelated TestQLT01RegistryCoversAllFiveSpikes failure, now recorded as D-08-43"
    verification:
      - kind: unit
        ref: "go test ./... && go vet ./..."
        status: pass
    human_judgment: false
  - id: D6
    description: "Criterion 4's legibility half: a reader can tell, without consulting the callee's source, which callee-signature field a diagnostic's liveness answer depended on"
    verification: []
    human_judgment: true
    rationale: "This is a judgement about a reader's experience of the rendered diagnostic, not an assertion about bytes -- the field SET is machine-checked separately by TestInterproceduralDisclosedFieldSet (08-03). Self-assessed during this plan (auto mode, no blocking-human gate on this <human-check>): the relay_depth2_refuse.lang diagnostic's third cause reads callee_return_contract detail=\"...fn:relay:return.mode=shared\", naming the exact field and its value without requiring the callee's source -- PASS by inspection, but a genuine human reader's confirmation was not collected and would strengthen this."

# Metrics
duration: ~45min
completed: 2026-09-10
status: complete
---

# Phase 08 Plan 06: The Mandatory Mid-Phase Gate Summary

**The mid-phase gate re-ran criterion 1's ten-fixture corpus and criterion 3's five-shape growth-exponent fit live against the working tree, adjudicated all four agenda items explicitly (two resolved, two carried as named debt), re-ratified the cost-bound manifest row at its own commit, declared the interprocedural liveness law final in code, and expanded the debt register from 5 to 9 tracked items to formalize what prior plans had only noted in prose.**

## Performance

- **Duration:** ~45 min
- **Started:** 2026-09-10T03:19 (following 08-05's completion commit)
- **Completed:** 2026-09-10T04:05 (approx)
- **Tasks:** 3 completed (Task 1 `checkpoint:decision`, auto-adjudicated; Tasks 2-3 `type="auto"`)
- **Files modified:** 4 (0 created)

## Accomplishments

- **Task 1 -- the gate itself.** Re-ran all ten criterion-1 fixtures live via `go run ./cmd/lang --json check` and confirmed every verdict matches its documented expectation:

  | Fixture | Verdict |
  |---|---|
  | `twin_a_refuse.lang` | `check.interprocedural_loan_liveness` |
  | `twin_a_accept.lang` | clean at `check` (corevalidate diverges: `core.move_while_borrowed`, registered residual D-08-40) |
  | `twin_b_refuse.lang` | `ownership.move_while_borrowed` (intraprocedural, documented scope limit D-08-41) |
  | `twin_b_accept.lang` | `ownership.move_while_borrowed` (same, identical to refuse member) |
  | `relay_depth2_refuse.lang` | `check.interprocedural_loan_liveness` |
  | `relay_depth2_accept.lang` | clean at `check` (corevalidate diverges: `core.move_while_borrowed`, registered residual D-08-40) |
  | `negative_control_fails.lang` | `check.interprocedural_loan_liveness` |
  | `negative_control_infallible.lang` | `check.interprocedural_loan_liveness` (identical to fails member, per 08-03's field-equality assertion) |
  | `match_arm_call.lang` | clean (no diagnostics) |
  | `relay_escort_witness.lang` (07) | `check.interprocedural_loan_liveness` |

  Re-fit criterion 3's growth exponent directly from `go test ./internal/compiler/check/... -run GrowthExponentInOps -v`'s printed ops/work pairs via an independent log-log OLS fit: chain 1003, diamond 1002, dense 1003, parser-shaped 1000, forward 998 (all milli-exponent, bound 1200); ratio-stability deltas between S=128 and 4S=512 all under 0.2% against the 15% tripwire. Adjudicated all four agenda items (see key-decisions): (a) OWN-09 carried, unresolved by design; (b) criterion 4's disclosure vehicle carried; (c) both D-08-09 admission paths resolved as still load-bearing; (d) both planner assumptions ratified. Confirmed the scope-cut trigger did not fire, so the liveness law may be declared final.
- **Task 2 -- transcription and re-ratification.** `PHASE-08-DEBT.md` grew from 5 to 9 tracked items: D-08-26/27/38 gained dated gate-review paragraphs (38 resolved outright; 26/27 carried); four new items (D-08-40 through D-08-43) formalize the peer-divergence residual, Pattern B's scope limit, the stale `callgraph.Order` direction in two planning docs, and the pre-existing spike-registry test failure. `TestDebtRegistersAreWellFormed` passes against the expanded register. The `recomputed_work_growth_exponent` manifest row's `ratified_at`/`ratified_by_commit` now point at this plan's own debt-register commit (`b28a92f`), committed via a two-step sequence (debt register first, then the manifest edit referencing that already-known hash) so the row never has to reference its own not-yet-existing commit.
- **Task 3 -- declaring the law final.** `checkInterproceduralLoanLiveness`'s doc comment now states, in one place, the four facts that make up the law's final scope: the two consulted callee-signature fields, the body-free two-tier production/consumption split, the `check.*`-to-`core.*` Phase 09 promotion schedule, and `ownership.*`'s deliberate non-retirement. `OWN-06` and `EFF-02` read complete in `REQUIREMENTS.md`; `OWN-07/08/09` are untouched. Full suite plus `go vet` re-run clean except the one pre-existing, now-tracked `TestQLT01RegistryCoversAllFiveSpikes` failure (D-08-43).

## Task Commits

Each task was committed atomically:

1. **Task 1: The mandatory mid-phase gate** - no code commit (checkpoint:decision, auto-adjudicated; its dispositions are transcribed by Task 2's commit)
2. **Task 2a: Record the adjudications in the debt register** - `b28a92f` (docs)
3. **Task 2b: Re-ratify the manifest row at the gate's commit** - `9d13d6c` (chore)
4. **Task 3: Declare the liveness law final** - `cb4118f` (docs)

**Plan metadata:** commit pending (this SUMMARY + STATE.md/ROADMAP.md/REQUIREMENTS.md)

## Files Created/Modified

- `.planning/phases/08-interprocedural-loan-liveness-in-check/PHASE-08-DEBT.md` - 5 -> 9 tracked items; dated gate-review resolutions on D-08-26/27/38
- `internal/compiler/session/qlt02_budget_manifest.json` - `recomputed_work_growth_exponent` row's `ratified_at`/`ratified_by_commit` re-pointed at this gate's own commit
- `internal/compiler/check/check.go` - `checkInterproceduralLoanLiveness`'s doc comment gains the law's final-scope closing paragraph
- `.planning/REQUIREMENTS.md` - confirmed unchanged in substance (OWN-06/EFF-02 already complete; see Deviations)

## Decisions Made

See `key-decisions` in frontmatter for the full set: the auto-adjudication basis, each agenda item's disposition (a carried unresolved, b carried, c resolved, d ratified), the scope-cut trigger's resolution, the four new debt items, and the pre-existing REQUIREMENTS.md ordering deviation.

## Deviations from Plan

### Auto-fixed Issues

None this plan required a Rule 1-3 auto-fix to production code — Tasks 2 and 3 executed as written.

### Documented Ordering Deviation (not a Rule 1-3 auto-fix — no code path changed to "fix" this)

**1. OWN-06 and EFF-02 were already marked complete in `REQUIREMENTS.md` before this gate ran.**
- **Found during:** Required-reading pass, before Task 1
- **Issue:** This plan's own `must_haves.truths` states "OWN-06 and EFF-02 are marked complete in REQUIREMENTS.md only after the gate has adjudicated." Git history shows `9ad04ce` (08-03's completion commit) already flipped OWN-06 to `[x]`/Complete, and `9cfff62` (08-05's completion commit) already flipped EFF-02 to `[x]`/Complete — both well before this plan (08-06) ran the gate. The most likely mechanism: `gsd-tools requirements.ready-ids`'s shared-ID gate scans sibling `*-PLAN.md` files present in the phase directory at scan time; if `08-06-PLAN.md` did not yet exist when 08-03/08-05 ran their own `update_requirements` step, the gate could not see that OWN-06/EFF-02 were also declared by a not-yet-created sibling plan, so it (incorrectly, from this plan's own must_haves) treated them as ready.
- **Resolution:** Not reverted. The gate's own adjudication (this plan) confirms both marks are now correct IN SUBSTANCE — criterion 1 and criterion 3 both hold on re-run evidence, and the liveness law is genuinely declared final by this same plan. Reverting the marks to `[ ]` and re-flipping them within this same plan would create a confusing double-edit in the same file with no substantive benefit, since the final state is correct either way. Documented here instead, per Rule 2 (missing critical process step -- the ordering guarantee itself, not the correctness of the final marks) so a future gate-authoring pass can consider hardening `requirements.ready-ids` to also block on requirement IDs declared only by a not-yet-existing sibling, if that gate's own tooling is ever revisited.
- **Files affected:** `.planning/REQUIREMENTS.md` (no edit made by this plan)
- **Verification:** `git log -p -- .planning/REQUIREMENTS.md | grep -B1 'OWN-06\|EFF-02'` confirms the exact prior commits and lines.
- **Impact:** No functional impact — both requirements are genuinely satisfied as of this gate. Process-ordering-only deviation, fully disclosed rather than silently treated as Task 3(b) having executed the mark itself.

---

**Total deviations:** 1 documented (process-ordering only, no code changed, no scope creep).
**Impact on plan:** None on the phase's technical outcome. The gate's own adjudication independently confirms OWN-06/EFF-02's correctness on the evidence, regardless of when the checkbox was flipped.

## Issues Encountered

None beyond the documented ordering deviation above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- The interprocedural loan-liveness law is declared final: `check.*` consults exactly `return.mode` and `parameters[0].mode`, never a callee's body; production is callee-before-caller, one pass; the code is scheduled for `core.*` promotion in Phase 09 once `corevalidate` independently re-derives the same fact.
- `PHASE-08-DEBT.md` carries 9 dated, machine-shape-checked items for Phase 09 (and later phases) to consult: D-08-27 (OWN-09's document conflict, needs a human decision), D-08-26 (criterion 4's accepted-program disclosure vehicle), D-08-40 (the two accepted-program peer-divergence fixtures, concrete input for `corevalidate`'s own peer extension), D-08-41 (Pattern B's real-fixture scope limit, informational), D-08-42 (stale planning-doc callgraph-order direction, informational), D-08-43 (the pre-existing spike-registry test failure, informational, not Phase 08's to fix).
- Phase 09's peer re-derivation plan has a stable, code-documented target to write its independence argument and mutation-kill evidence against.
- OWN-06 and EFF-02 are complete; OWN-07, OWN-08, OWN-09 remain Pending, correctly mapped to Phase 09.
- No blockers for Phase 09 planning.

## Self-Check: PASSED

- FOUND: .planning/phases/08-interprocedural-loan-liveness-in-check/PHASE-08-DEBT.md
- FOUND: internal/compiler/session/qlt02_budget_manifest.json
- FOUND: internal/compiler/check/check.go
- FOUND: .planning/REQUIREMENTS.md
- FOUND commit: b28a92f
- FOUND commit: 9d13d6c
- FOUND commit: cb4118f
- Re-ran Task 1 acceptance criteria: all ten criterion-1 fixture verdicts re-confirmed live via `go run ./cmd/lang --json check`; growth exponent re-fit live from raw test output (998-1003 milli, all shapes) -- PASS
- Re-ran Task 2 acceptance criteria: `go test ./internal/compiler/session/... -run 'TestDebtRegistersAreWellFormed|QLT02|BudgetAudit' -v` -- all PASS; manifest row's `ratified_by_commit` (`b28a92f927e40f5c0c6279dfef1368c660fe75c5`) confirmed different from the introducing commit (`f642ae1`) -- PASS
- Re-ran Task 3 acceptance criteria: `go build ./...` clean; `go run ./cmd/lang --json check testdata/phase07/relay_escort_witness.lang` emits `check.interprocedural_loan_liveness` -- PASS; `.planning/REQUIREMENTS.md` shows OWN-06/EFF-02 complete, OWN-07/08/09 untouched -- PASS
- Re-ran plan-level `<verification>`: `go test ./internal/compiler/session/... -run TestDebtRegistersAreWellFormed` exits 0; `go test ./... && go vet ./...` -- only the pre-existing, now-tracked (D-08-43) `TestQLT01RegistryCoversAllFiveSpikes` failure -- PASS

---
*Phase: 08-interprocedural-loan-liveness-in-check*
*Completed: 2026-09-10*
