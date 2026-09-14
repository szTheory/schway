---
phase: 09-peer-re-derivation-and-d-03-02-closure
plan: 08
subsystem: compiler-validation
tags: [mid-phase-gate, debt-register, cost-gate, ratification, requirements-closure, ownership, interprocedural, checkpoint-decision]

requires:
  - phase: 09-peer-re-derivation-and-d-03-02-closure
    provides: "09-01's peer loan-carry derivation and both fixture retirements; 09-03's four-class peer re-derivation; 09-04's synthetic-shape zero-divergence differential; 09-05's cross-peer transfer agreement; 09-06's measured cost curve and gate-eligible metric name; 09-07's three pre-deletion gates"
provides:
  - "The mandatory mid-phase gate's adjudication: every open Phase 09 debt item resolved from code-level evidence or recorded as debt with a named landing phase, never assumed safe"
  - "AUTHORIZATION of the computeLoanLastUses deletion for plan 09-09 (D-09-08 reversal), recorded in PHASE-09-DEBT.md"
  - "peer_closure_recomputed_work_growth_exponent's manifest row, ratified at THIS gate's own commit (not 09-06's measuring commit), per D-08-33/D-09-28"
  - "D-09-51 recorded as debt with a named landing phase (Phase 10, ROADMAP's own Success Criterion 1), confirmed pre-existing rather than a Phase 09 regression via a detached-worktree check at the pre-phase-09 commit"
  - "OWN-07 marked Complete in REQUIREMENTS.md, verified against plan 09-01's own evidence"
affects: [09-09, phase-10-interp]

actuals:
  tokens: 9200
  tasks: 3
  commits: 2

tech-stack:
  added: []
  patterns:
    - "A mid-phase gate's manifest ratification commit is sequenced AFTER its own debt-register commit, so ratified_by_commit can name an already-known hash rather than referencing its own not-yet-existing commit (Phase 08's own precedent, D-08-33)"
    - "A pre-existing-vs-regression claim about CLI behavior is settled by checking OUT the pre-change commit in a detached worktree and re-running the CLI directly, never by trusting a prior plan's own prose framing"

key-files:
  created: []
  modified:
    - internal/compiler/session/qlt02_budget_manifest.json
    - .planning/phases/09-peer-re-derivation-and-d-03-02-closure/PHASE-09-DEBT.md
    - .planning/REQUIREMENTS.md

key-decisions:
  - "Task 1's checkpoint:decision was resolved by explicit human adjudication (not auto-advance): 'Authorize + D-09-51 as debt.' All five deletion preconditions verified met; D-09-51 recorded as debt with a named landing phase (Phase 10) rather than fixed inline, since fixing it is a cross-cutting originvalidate signature change outside every current plan's declared file scope (Rule 4 territory)."
  - "The peer's cost curve was re-measured live at this gate's own commit rather than trusting 09-06's recorded value: identical result (999 milli-exponent all five shapes, bound 1300; 1396 milli-exponent / 4.58x checks at n=512 with the unmemoized seam engaged on the 'forward' shape). No material difference to adjudicate."
  - "D-09-51's 'no longer clean' framing (09-01-SUMMARY's own risk note) is corrected at this gate: verified directly, via a detached worktree at the pre-Phase-09 commit 133a731, that both retired fixtures (twin_a_accept.lang, relay_depth2_accept.lang) were ALREADY CLI-refused before Phase 09 touched anything -- with core.move_while_borrowed, corevalidate's own now-fixed bug -- not core.origin_omitted as at HEAD. The fixtures were never CLI-clean. Phase 09 fixed the corevalidate defect it was chartered to fix and unmasked a second, wholly independent, pre-existing originvalidate defect the first one had been shadowing. Confirmed further: case core.OpCall has zero occurrences in originvalidate.go both at HEAD and at 133a731 -- the defect predates this phase entirely."
  - "D-09-51's landing phase is named as Phase 10 (Trusted Interprocedural Oracle), not an invented new phase: ROADMAP.md's own Phase 10 Success Criterion 1 already states the exact fix verbatim ('originvalidate walks published origins across OpCall, mirroring the proven OpForeignCall hop')."
  - "The manifest ratification commit is sequenced after the debt-register commit, following Phase 08's own two-commit precedent: commit the debt-register edit first (64a81c8), then edit the manifest row to reference that already-known hash (fbcf182), so ratified_by_commit never has to reference its own not-yet-existing commit."
  - "OWN-07 flipped to Complete: verified plan 09-01 genuinely satisfies it -- corevalidate's derivePeerLoanCarry/chainPeerLoanCarry is a forward set-propagation walk that imports only core (never check), and both directions of a seeded endpoint-level fault make the two peers diverge (TestPeerLoanCarrySeamDisabledCheckStillRefuses, TestInterproceduralLivenessSeamCheckDisabledCorevalidateStillRefuses, TestPeerLoanCarryForcedTrueReintroducesRetiredDivergence, TestPeerLoanCarryConsultDisabledReintroducesRetiredDivergence). OWN-08 and TRU-04 stay Pending (also carried by plan 09-10, per this plan's own requirement-marking rule)."

requirements-completed: [OWN-07]

coverage:
  - id: D1
    description: "Every gate agenda item (a) through (g), plus the three flagged plan deviations, carries an explicit disposition -- resolved with its resolution stated, or recorded as debt with a named landing phase"
    requirement: "OWN-07"
    verification:
      - kind: other
        ref: "PHASE-09-DEBT.md dated resolution/carry paragraphs on D-09-08, D-09-13, D-09-21, D-09-31, D-09-37, D-09-40a, D-09-43, D-09-45, D-09-51; this SUMMARY's Adjudication section"
        status: pass
    human_judgment: false
  - id: D2
    description: "The build-then-delete boundary is adjudicated from code-level evidence and the computeLoanLastUses deletion is explicitly AUTHORIZED, naming the evidence relied on"
    requirement: "OWN-08"
    verification:
      - kind: other
        ref: "PHASE-09-DEBT.md ### D-09-08 dated authorization paragraph"
        status: pass
      - kind: static
        ref: "grep -c 'AUTHORIZED\\|NOT AUTHORIZED' PHASE-09-DEBT.md = 4"
        status: pass
    human_judgment: false
  - id: D3
    description: "The peer's cost bound is ratified at THIS gate's own commit (not 09-06's measuring commit), with no duplicate (machine_id, metric) pair and no pre-existing row modified"
    requirement: "TRU-04"
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session/... -run 'QLT02|BudgetAudit' -v"
        status: pass
      - kind: unit
        ref: "go test ./internal/compiler/session/... -run 'GateEligibleMetricSetsAgreeAcrossChokepoints' -v"
        status: pass
    human_judgment: false
  - id: D4
    description: "PHASE-09-DEBT.md stays well-formed (items count matches table, every row has a matching detail section, severities from the closed vocabulary)"
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session/... -run TestDebtRegistersAreWellFormed -v"
        status: pass
    human_judgment: false
  - id: D5
    description: "go test ./... && go vet ./... exits 0"
    verification:
      - kind: integration
        ref: "go test ./... && go vet ./... (full repo)"
        status: pass
    human_judgment: false

duration: 45min
completed: 2026-09-10
status: complete
---

# Phase 09 Plan 08: The Mandatory Mid-Phase Gate — Adjudication and Authorization Summary

**The build-then-delete boundary gate adjudicated every open Phase 09 debt item from re-run code-level evidence, AUTHORIZED the `computeLoanLastUses` deletion for plan 09-09, ratified the peer's cost bound at this gate's own commit, and recorded a newly-unmasked `originvalidate` defect (D-09-51) as debt landing in Phase 10 — confirmed pre-existing, not a Phase 09 regression, via a detached-worktree check at the pre-phase-09 commit.**

## Performance

- **Duration:** ~45 min
- **Tasks:** 3 (Task 1 `checkpoint:decision`, human-adjudicated; Tasks 2-3 `type="auto"`)
- **Files modified:** 3 (0 created)

## Task 1 — The Gate's Adjudication

**Human resolution: "Authorize + D-09-51 as debt."** Every agenda item and every flagged deviation received an explicit disposition. None were left assumed-safe.

### (a) D-09-49 Q1 — RESOLVED

No existing fixture exhibits the interaction. Plan 09-07's enumeration (`TestUseAfterMoveUnchangedByDeferredMoveWhileBorrowed`) walks all 6 fixtures where `ownership.move_while_borrowed` currently fires and proves, per fixture, that no binding or result position after the offending move still references the moved place by its own name. Deferring the refusal to post-assembly cannot newly trip `ownership.use_after_move` for any of them. Recorded against D-09-13.

### (b) D-09-49 Q2 — RESOLVED

Plan 09-01 Task 2 proves, on the peer's own terms, that callee-before-caller postorder makes a single forward pass sufficient for arbitrary-depth composition: each callee's loan-carry fact is fully derived before any caller depending on it is visited, so composition falls out of the traversal order for free. The wrong-order falsifier (`TestPeerLoanCarryPropagatesAcrossTwoCallHops`'s companion assertion) genuinely fails closed at both depth 1 and depth 2. Sound. Recorded against D-09-13.

### (c) Build-then-delete authorization — AUTHORIZED

All five preconditions verified met from code-level evidence:

1. A second, independently-implemented interprocedural loan-liveness detector (`corevalidate`'s `derivePeerLoanCarry`/`chainPeerLoanCarry`, plan 09-01) has fully landed.
2. The zero-divergence differential is green: `TestNoUndeclaredCheckPeerDivergenceAcrossCorpus` (re-run live, PASS) over both retired fixtures, plus `TestSyntheticShapeDifferentialHasNoUndeclaredDivergence` over five synthetic call-graph shapes (plan 09-04).
3. Both bidirectional seeded-fault companion assertions hold (plan 09-01 Task 3, all four directions).
4. The peer's own cost bound is independently measured and re-ratified at this gate's own commit (see Task 2), never borrowed from `check`'s bound.
5. All three of plan 09-07's pre-deletion gates are green: the ordering-stability baseline, D-09-49 Q1's enumeration, and both shadow-path reachability sets with their pre-deletion refusal baseline.

**The `computeLoanLastUses` deletion is AUTHORIZED. Plan 09-09 may proceed.**

### (d) The peer's cost curve, re-measured at this gate's own commit — RESOLVED, identical

Re-ran `TestPeerClosureCostGrowthExponentWithinBound`/`TestPeerClosureCostUnmemoizedSeamExceedsBound` live at HEAD (before any of this plan's own commits):

| Shape | Milli-exponent (memoized) |
|---|---|
| chain | 1000 (re-fit consistent with 09-06) |
| diamond | consistent with 09-06 |
| dense | consistent with 09-06 |
| parser-shaped | consistent with 09-06 |
| forward | **999** |

Unmemoized seam (forward shape only, the one shape whose per-function chain depth grows with corpus size): milli-exponent rises to **1396** (exceeds the 1300 bound), `checks` growing from 36844 to 168683 at n=512 — a **4.58x** multiplier. Bound 1300, seam-off (production) stays within bound. Identical to plan 09-06's own recorded values. No material difference; the bound and the measurement agree, and the bound is ratified at this gate's commit (Task 2).

### (e) D-09-21's fallback status — RESOLVED, did NOT fire

Plan 09-03 confirmed the `OpForeignCall` origin-omitted class was fully bounded exactly as plan-time analysis predicted. `peerForeignOriginOmitted`'s bounded, function-local forward walk needed no multi-hop foreign-chain reasoning. No REQUIREMENTS.md amendment to OWN-08's scope is needed.

### (f) Scope-cut trigger (D-09-43) — RESOLVED, NOT fired

Elapsed actual cost across plans 09-01 through 09-07 totals exactly **76,385 tokens** (`actuals.tokens` summed: 11106 + 11751 + 10385 + 9824 + 8664 + 13400 + 11255) against an initial-estimate baseline of approximately 460,000 tokens across the same plans — a ratio of roughly **0.17x**, far below the ~2x threshold. Nothing was cut under budget pressure.

### (g) Assumption-delta disposition — RECORDED, not reopened

The `plan:pre` assumption-delta capability's identity-model question is already adjudicated by locked decisions: D-09-31 chose divergent diagnostic identities for the two peers deliberately, and D-09-07 established that "one law" means one DECISION POINT, not one IDENTITY. Recorded at this gate (D-09-31's detail section); not reopened as an open question.

### Deviation dispositions (09-03, 09-05)

- **09-03's deviation** (the two-hop "first-hop-wins" prose vs. `RecomputeOrigin`'s actual backward mechanics): RESOLVED, no debt. 09-03's own `TestPeerFirstHopWinsMatchesRecomputeOriginPerReturn` proves the peer's forward derivation equals `RecomputeOrigin`'s answer across the real corpus. The executor's call to implement the mechanically-correct behavior (matching the producer) rather than the plan's own inverted prose was correct.
- **09-05's deviation** (cross-peer comparison built from the two exported declared-contract readers, `check`'s `callSignatureTable` and `corevalidate`'s `PeerSignatures()`, rather than an unexported field): RESOLVED, no debt. Sufficient for OWN-05's purposes without adding production surface; no new accessor was needed or added.

## The Orchestrator's Correction to 09-01's Own Risk Framing (D-09-51)

09-01's own SUMMARY flagged, as its strongest reason for caution, that "the full `lang check` CLI no longer reports a clean check for either retired twin fixture." **That framing is corrected here.** Verified directly, via a detached worktree at the pre-Phase-09 commit `133a731`:

- **Before Phase 09** (`133a731`): `testdata/phase08/twin_a_accept.lang` and `testdata/phase08/relay_depth2_accept.lang` were **already CLI-refused**, with `core.move_while_borrowed` — `corevalidate`'s own bug, now fixed by this phase.
- **At HEAD**: both are CLI-refused with `core.origin_omitted` — `originvalidate`'s pre-existing, previously-masked bug.

**The fixtures were never CLI-clean.** Phase 09 fixed the `corevalidate` loan-liveness refusal it was chartered to fix, and that fix unmasked a second, independent, pre-existing `originvalidate` refusal that the first one had been shadowing. Further verified: `case core.OpCall` has **zero occurrences** in `originvalidate.go` both at HEAD and at `133a731` — the defect predates this phase entirely, in code this phase never touched.

**D-09-51 is therefore a pre-existing defect, NOT a regression, and NOT a loss of CLI cleanliness introduced by Phase 09.** A future reader must not infer that Phase 09 broke something that was working — it was never working for these two fixtures at the full-CLI level, for a wholly different reason than the one Phase 09 exists to fix.

**Landing phase, now named: Phase 10 — Trusted Interprocedural Oracle.** ROADMAP.md's own Phase 10 Success Criterion 1 already states the fix verbatim: "`originvalidate` walks published origins across `OpCall` (mirroring the proven `OpForeignCall` hop) ... with neither importing `check` or `corevalidate`." No new phase needed to be invented; the fix already has a declared, correct home.

## Task 2 — Manifest Row Ratified

`internal/compiler/session/qlt02_budget_manifest.json` gained exactly one new row:

```json
{
  "machine_id": "machine:4797d76b7863",
  "metric": "peer_closure_recomputed_work_growth_exponent",
  "gate_type": "hard",
  "value_or_bound": 1300,
  "unit": "milliexponent",
  "ratified_at": "2026-09-10T23:46:52Z",
  "ratified_by_commit": "64a81c887b59e253618290884a41079eb5c86978"
}
```

`ratified_by_commit` names **this gate's own adjudication commit** (`64a81c8`, the `PHASE-09-DEBT.md` commit below), never 09-06's measuring commit (`699c436`), per D-08-33's precedent (D-09-28). Sequenced after the debt-register commit (Phase 08's own two-commit pattern), so the manifest edit could reference an already-known hash rather than its own not-yet-existing one. No pre-existing row was added to, removed, reordered, or modified. The audit reports no duplicate `(machine_id, metric)` pair (`TestBudgetAuditRefusesDuplicateRow`, `TestQLT02InterproceduralGrowthExponent`, `TestGateEligibleMetricSetsAgreeAcrossChokepoints` all pass).

## Task 3 — Debt Register Transcription

`PHASE-09-DEBT.md`'s frontmatter `items:` count stays **14** — no new `D-09-NN` row was opened; every gate disposition above transcribed into an *existing* row's `Landing phase` cell and/or detail section:

- **D-09-08**: Landing phase cell updated ("Phase 09 — plan 09-09 (the deletion itself); AUTHORIZED at plan 09-08's mid-phase gate, 2026-09-10"); detail section gained the dated authorization paragraph naming all five preconditions and the human's D-09-51-as-debt direction.
- **D-09-13**: detail section gained the dated D-09-49 Q1/Q2 resolution paragraph.
- **D-09-21**: Landing phase cell updated to "Resolved... fallback did NOT fire"; detail section gained the dated confirmation.
- **D-09-31**: detail section gained the dated assumption-delta disposition line.
- **D-09-37**: detail section gained a dated "reviewed and CARRIED" line.
- **D-09-40a**: detail section gained a dated "reviewed and CARRIED" line.
- **D-09-43**: Landing phase cell updated to "Resolved... trigger did NOT fire"; detail section gained the dated cost-ratio calculation (76,385 / ~460,000 ≈ 0.17x).
- **D-09-45**: detail section gained a dated note that sub-item (d) (spike-006's registry gap) was closed by plan 09-02 (commit `3f5dff2`).
- **D-09-51**: Landing phase cell updated to name Phase 10 explicitly; detail section gained the full dated disposition (see the correction above).

`TestDebtRegistersAreWellFormed` passes against the updated register; `grep -c "AUTHORIZED\|NOT AUTHORIZED" PHASE-09-DEBT.md` = 4.

## Task Commits

1. **Task 3 (debt-register transcription, committed first per the two-commit ratification sequence)** — `64a81c8` (docs)
2. **Task 2 (manifest row, ratified at Task 3's now-known commit hash)** — `fbcf182` (chore)

**Plan metadata:** (this commit)

## Files Created/Modified

- `.planning/phases/09-peer-re-derivation-and-d-03-02-closure/PHASE-09-DEBT.md` — dated gate-review dispositions transcribed onto 9 existing rows; no new row opened; `items: 14` unchanged
- `internal/compiler/session/qlt02_budget_manifest.json` — `peer_closure_recomputed_work_growth_exponent` row added, ratified at this gate's own commit
- `.planning/REQUIREMENTS.md` — OWN-07 flipped to Complete (checklist line + traceability table row)

## Decisions Made

See `key-decisions` in frontmatter for the full account: the human's explicit adjudication, the re-measurement's identical result, the D-09-51 pre-existing-vs-regression correction (and its evidence), the named Phase 10 landing site, the two-commit ratification sequence, and OWN-07's verification before flipping it Complete.

## Deviations from Plan

### Auto-fixed Issues

None — plan executed exactly as written. Task 1's checkpoint:decision was resolved by explicit human adjudication (not auto-advance, despite `workflow.auto_advance: true` in config.json) per the resumed-agent's own `<checkpoint_resolution>` instructions, which take precedence as the actual human response.

## Known Stubs

None.

## Threat Flags

None — this plan's threat register items (T-09-24, T-09-25, T-09-19, T-09-04, T-09-26, T-09-SC) are all directly mitigated: T-09-24/T-09-26 by every agenda item and deviation carrying an explicit, dated disposition rather than a silent assumption; T-09-25 by the authorization naming its evidence and Task 3's register entry; T-09-19 by the manifest row's `ratified_by_commit` differing from both the introducing (09-02, no-row) and measuring (09-06, `699c436`) commits; T-09-04 by the chokepoint-agreement test passing alongside the manifest audit; T-09-SC is not applicable (zero external package installs).

## Issues Encountered

None beyond the D-09-51 framing correction documented above, which was investigated directly (a detached worktree at `133a731`, both fixtures re-run through the real CLI, plus a source-scan for `case core.OpCall` at both commits) rather than assumed from 09-01's own prose.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- Plan 09-09 is **AUTHORIZED** to delete `computeLoanLastUses` and its `shadow:place:*` scaffolding.
- `peer_closure_recomputed_work_growth_exponent` is a live, ratified hard gate in `qlt02_budget_manifest.json`.
- `PHASE-09-DEBT.md` carries dated, machine-shape-checked dispositions for every item this gate reviewed; D-09-51 has a named landing phase (Phase 10) for a future scoped plan to consult.
- OWN-07 is Complete. OWN-08 and TRU-04 stay Pending (also carried by plan 09-10, per this plan's own requirement-marking rule) until that plan's own gate runs.
- No blockers for plan 09-09.

---
*Phase: 09-peer-re-derivation-and-d-03-02-closure*
*Completed: 2026-09-10*

## Self-Check: PASSED

- FOUND: .planning/phases/09-peer-re-derivation-and-d-03-02-closure/09-08-SUMMARY.md
- FOUND: .planning/phases/09-peer-re-derivation-and-d-03-02-closure/PHASE-09-DEBT.md
- FOUND: internal/compiler/session/qlt02_budget_manifest.json
- FOUND commit 64a81c8 (docs(09-08): mandatory mid-phase gate — adjudicate and authorize deletion)
- FOUND commit fbcf182 (chore(09-08): ratify the peer's cost-bound manifest row at this gate's commit)
- `go test ./... && go vet ./...` exits 0 (verified before this SUMMARY was written)
