---
milestone: M002
milestone_name: Interprocedural Semantic Spine
audited: 2026-09-13T00:00:00Z
status: tech_debt
scores:
  requirements: 29/31 satisfied, 2/31 honestly partial, 0 unsatisfied, 0 orphaned
  phases: 7/7 verified passed
  integration: 6/6 OpCall dispatch sites wired, 0 orphaned exports
  flows: 3/3 end-to-end flows complete
gaps: {}
partial:
  requirements:
    - id: "DX-06"
      status: "partial"
      phase: "Phase 13"
      claimed_by_plans: ["13-02-PLAN.md", "13-06-PLAN.md"]
      completed_by_plans: ["13-02-SUMMARY.md", "13-06-SUMMARY.md"]
      verification_status: "passed (criterion honestly reported partial)"
      debt_row: "PHASE-13-DEBT.md D-13-02b"
      evidence: >-
        Contract-boundary blame resolver is built, tested, and compile-time
        exhaustive (blameFieldWitness), but deliberately unwired: `grep -rn
        'resolveBlame(' --include='*.go' internal/` returns exactly 1 hit, the
        definition at check.go:769, and zero call sites. The rule cannot be
        exercised because no B1-shaped (contract-violation) interprocedural
        diagnostic is constructible while `sameType(ReturnType,
        Parameter.Type)` is enforced at function admission
        (check.go:255, :3148, :3399) independent of any call. Ratified
        terminal finding on the D-12-43 precedent.
    - id: "DX-07"
      status: "partial"
      phase: "Phase 13"
      claimed_by_plans: ["13-01-PLAN.md", "13-04-PLAN.md", "13-05-PLAN.md", "13-06-PLAN.md"]
      completed_by_plans: ["13-01-SUMMARY.md", "13-04-SUMMARY.md", "13-05-SUMMARY.md", "13-06-SUMMARY.md"]
      verification_status: "passed (criterion honestly reported partial)"
      debt_row: "PHASE-13-DEBT.md D-13-10a"
      evidence: >-
        Requirement asks for at least three new repairable interprocedural
        classes; exactly two ship. `move_after_interprocedural_loan`
        (check.go:1703) and `wrap_call_in_try` (check.go:3645) both emit
        MachineApplicable repairs that reach `repaired` on sealed held-out
        fixtures. The third, `use_matching_argument`, was withdrawn
        empirically: it appears in check.go only at :3503 and :3535, both
        comments recording why no repair is emitted. Its replacement was
        byte-identical to the original on every real trigger, because the
        single-type-per-function invariant makes the "correct" argument always
        the one already passed.
tech_debt:
  total_items: 77
  closed_or_resolved: 26
  open_and_unowned: 10
  blockers_open: 0
  registers:
    - phase: 07-calls-signatures-and-call-graph-refusal
      items: 11
      warnings: 3
    - phase: 08-interprocedural-loan-liveness-in-check
      items: 9
      warnings: 3
    - phase: 09-peer-re-derivation-and-d-03-02-closure
      items: 16
      warnings: 9
    - phase: 10-trusted-interprocedural-oracle
      items: 12
      warnings: 7
    - phase: 11-multi-function-native-emission-and-interprocedural-equivalen
      items: 16
      warnings: 9
    - phase: 12-result-payloads
      items: 10
      warnings: 4
    - phase: 13-agent-loop-for-interprocedural-defects
      items: 3
      warnings: 3
  unowned:
    - id: D-10-C04
      severity: warning
      reqs: [OWN-09, TRU-02]
      item: "D-09-51 negative-control verdict flip is unreviewed"
    - id: D-11-02
      severity: info
      reqs: [NAT-04, NAT-05, NAT-06, NAT-07]
      item: "Six single-function emitters not deleted; re-deferred at D-12-36"
    - id: D-11-27
      severity: info
      reqs: [NAT-07]
      item: "No phase claims the D-09-51 flip review NAT-07 was declared not to depend on"
    - id: D-11-51
      severity: warning
      reqs: [NAT-06]
      item: "Shared-leaf diamond call graphs collide on event identity"
    - id: D-12-21
      severity: warning
      reqs: [NAT-06]
      item: "D-11-51 interaction anticipated, not absorbed; blocked on D-11-51"
    - id: D-12-36
      severity: warning
      reqs: [NAT-04, NAT-05, NAT-06, NAT-07]
      item: "D-11-02 deletion re-deferred with a stated reversal"
    - id: D-12-43
      severity: warning
      reqs: [RES-03]
      item: "Decisive value-divergence control unconstructible against current representation"
    - id: D-13-02b
      severity: warning
      reqs: [DX-06]
      item: "B1 contract-violation blame structurally unreachable at this maturity"
    - id: D-13-10a
      severity: warning
      reqs: [DX-07]
      item: "use_matching_argument withdrawn as unrepairable"
    - id: D-13-34
      severity: warning
      reqs: []
      item: "M001 testdata/phase6 move and borrow held-out pairs are structurally identical — a hole in shipped M001 evidence"
nyquist:
  compliant_phases: ["09", "10"]
  partial_phases: []
  not_validated_phases: ["07", "08", "11", "12", "13"]
  missing_phases: []
  overall: partial
doc_defects:
  - file: .planning/ROADMAP.md
    line: 851
    issue: >-
      "30 of 30 M002 requirements mapped" contradicts the per-phase table
      directly below it (5+2+6+6+7+2+3 = 31) and REQUIREMENTS.md's own
      "M002 requirements: 31 total". The OWN-05 -> OWN-05a/OWN-05b split
      (D-09-37) was applied to the table but not the headline.
    severity: info
    fixed: true
---

# Milestone M002 — Interprocedural Semantic Spine — Audit Report

**Audited:** 2026-09-13
**Scope:** Phases 07-13
**Status:** `tech_debt` — every requirement is satisfied or honestly partial,
no blockers remain, and the accumulated debt needs a disposition before M003.

**Milestone thesis (ROADMAP.md:10-22):** add exactly one new
`core.OperationKind` — `OpCall` — and prove every semantic guarantee M001
established intraprocedurally still holds across a function boundary, under two
standing rules: *no fact crosses a trust boundary without being independently
re-derived*, and *no callee body is ever read to admit a caller*.

**Scale:** 385 commits, 141 Go files touched, +44,575 / −1,466 lines, 61 new
`.lang` fixtures, since the milestone opened at `f2490a8`.

---

## 1. Requirements Coverage (3-source cross-reference)

All 31 M002 requirements were cross-referenced against three independent
sources: REQUIREMENTS.md's traceability table, each phase's VERIFICATION.md
requirements table, and each plan SUMMARY.md's `requirements` frontmatter.

| Phase | Requirements | Count | All three sources agree |
|-------|--------------|-------|-------------------------|
| 07 | SEM-04, SEM-05, SEM-06, SEM-07, QLT-08 | 5 | Yes |
| 08 | OWN-06, EFF-02 | 2 | Yes |
| 09 | OWN-05a, OWN-07, OWN-08, OWN-09, TRU-04, QLT-07 | 6 | Yes |
| 10 | SEM-08, SEM-09, TRU-02, TRU-03, QLT-04, OWN-05b | 6 | Yes |
| 11 | NAT-04, NAT-05, NAT-06, NAT-07, QLT-03, QLT-05, QLT-06 | 7 | Yes |
| 12 | RES-02, RES-03 | 2 | Yes |
| 13 | DX-05, DX-06, DX-07 | 3 | Yes |

- **Satisfied:** 29
- **Partial (honest, developer-ratified):** 2 — DX-06, DX-07
- **Unsatisfied:** 0
- **Orphaned:** 0 — every traceability ID appears in at least one plan's
  frontmatter and in its owning phase's VERIFICATION.md requirements table;
  every frontmatter ID maps back to a traceability row.

**Checkbox/traceability consistency:** 29 requirements read `[x]` with a
`Complete` row. DX-06 and DX-07 read `[ ]` with a `Partial — <reason>` row that
names its owning debt entry. The two agree in both directions, so the
requirement-list-vs-traceability defect Phase 09's verifier caught (OWN-08,
fixed at `893567e`) is not repeated.

**The FAIL gate does not fire.** `partial` is not `unsatisfied`: both partials
have a passing phase verification, a named terminal finding, a debt row, and an
explicit developer ratification. Neither is an unexplained shortfall.

---

## 2. Phase Verifications

| Phase | Status | Score | Notes |
|-------|--------|-------|-------|
| 07 Calls, Signatures, Call-Graph Refusal | passed | 4/4 | Re-verified after gap closure; PVG-01/02/03 closed by plans 07-10/11/12 |
| 08 Interprocedural Loan Liveness in `check` | passed | 7/7 | Initial verification |
| 09 Peer Re-Derivation and D-03-02 Closure | passed | 5/5 | One documentation gap found and resolved at `893567e` |
| 10 Trusted Interprocedural Oracle | passed | 10/10 | Initial verification |
| 11 Multi-Function Native Emission | passed | 5/5 | Both `human_verification` items resolved `automated` (see below) |
| 12 `Result` Payloads | passed | 3/3 | Re-verified after gap closure; CR-01 closed by plans 12-06/07/08 |
| 13 Agent Loop for Interprocedural Defects | passed | 3/3 | 2 of 3 criteria honestly partial |

**Phase 11's two human-judgment items were eliminated rather than adjudicated**
— the stronger outcome. The `reduce.Seed` entry-ID hazard (11-REVIEW.md WR-01)
got the fail-closed `Seed.Validate` guard the review suggested, with
`TestSeedEntryHazardIsReal` as an anti-vacuity control that fails if the hazard
ever stops existing. The `CheckCommandFile`-vs-`session.Check` admission
divergence got `session_admission_divergence_test.go`, a sweep that pins the
exact divergence set across every committed fixture and fails in **both**
directions — a new divergence or a silently resolved one. No debt row was
filed, on the reasoning that a debt note is read once while a test runs on every
CI invocation.

---

## 3. Cross-Phase Integration

Checked by `gsd-integration-checker`, with its load-bearing claims independently
re-verified against the tree (see §3.1 — the checker overclaimed on two
requirements and its findings are adopted only where they survived that check).

### `OpCall` at six dispatch sites (SEM-04, the milestone's central claim)

| # | Site | Wired |
|---|------|-------|
| 1 | `check` | Yes |
| 2 | `corevalidate` | Yes |
| 3 | `interp` | Yes — via `partitionFrameForCall` + `runFrameStack` |
| 4 | `cgen` | Yes — C17 call with aliasing promises |
| 5 | `pathoracle` | Yes — `pathoracle_compose.go` splices concrete callee paths |
| 6 | `originvalidate` | Yes — origin walk consults callee return contract |

### Peer independence (TRU-03, OWN-07) — re-verified directly

```
go list -deps ./internal/compiler/pathoracle/...      -> no check, corevalidate, originvalidate
go list -deps ./internal/compiler/corevalidate/...    -> no check, originvalidate
go list -deps ./internal/compiler/originvalidate/...  -> no check, corevalidate
```

All three transitive-dependency guards are clean. `interp` imports
`corevalidate` deliberately, as a validation gate before execution, and
`TestInterpDoesNotReadCorevalidateOwnershipFields` proves it never reads the
ownership fields it would need to be borrowing an answer.

### Orphaned exports

Zero. Each phase's `provides:` frontmatter was traced to a consuming site in a
later phase.

### End-to-end flows

| Flow | Phases | Result |
|------|--------|--------|
| Multi-function program: parse → check → corevalidate → interp, then → cgen → native | 07-11 | Complete; interpreter and native agree |
| Interprocedural loan-liveness defect → diagnostic → repair → re-check clean | 08, 13 | Complete through the real `cmd/lang-repair` driver |
| Four-engine equivalence (interp / -O0 / -O3 / -O3 -flto) | 10-12 | Complete |

### Build and test

`go build ./...`, `go vet ./...`, and `go test ./...` are clean across all 25
packages.

### 3.1 Where the integration checker was overruled

The checker returned an unqualified "all 31 requirements satisfied,
integration-complete and ready for release," marking **DX-06 and DX-07 as
`✓ WIRED`**. That column is rejected. It contradicts both phase 13's own
verification and the tree:

- For **DX-06** it cited `TestRepairDriverFixesInterproceduralLoanLivenessSinglePass`
  — a *repair-driver* test — as evidence for *blame attribution*. These are
  different claims. `grep -rn 'resolveBlame(' --include='*.go' internal/`
  returns exactly one hit: the definition at `check.go:769`. Zero call sites.
  The blame resolver is deliberately unwired, exactly as D-13-02b records.
- For **DX-07** it reported the requirement satisfied on the strength of one
  repair class. The requirement asks for three. Two ship
  (`check.go:1703`, `check.go:3645`); `use_matching_argument` appears in
  `check.go` only at `:3503` and `:3535`, both comments explaining why it emits
  nothing.

It also reported 23 packages where `go list ./...` returns 25, and flagged
`cache/probe_test.go` as "occasionally flaky" — `go test ./internal/compiler/cache/...
-count=1` passes cleanly, and no flake was reproduced.

Its **structural** findings — the six dispatch sites, the independence guards,
the absence of orphaned exports, the three E2E flows — were spot-checked and
held, and are adopted above. Its **requirement-satisfaction** column was not.
This is worth recording as a process observation: an integration checker that
grades requirements from wiring alone will systematically convert an honest
partial into a false green, because wiring is exactly what a structurally
unreachable defect class still has.

---

## 4. Tech Debt

**77 items across 7 registers. 26 closed or resolved. 0 open blockers.**
The single `blocker`-severity item ever filed (D-12-44, CR-01's duplicate
payload-type ambiguity) is closed by plans 12-06 and 12-07.

| Phase | Items | Warnings | Closed/resolved |
|-------|-------|----------|-----------------|
| 07 | 11 | 3 | 1 |
| 08 | 9 | 3 | 3 |
| 09 | 16 | 9 | 5 |
| 10 | 12 | 7 | 4 |
| 11 | 16 | 9 | 7 |
| 12 | 10 | 4 | 6 |
| 13 | 3 | 3 | 0 |

### The 10 items that are OPEN and UNOWNED

These carry into M003 with no phase claiming them. This is the milestone's
principal outstanding liability and the reason the audit status is `tech_debt`
rather than `passed`.

| ID | Sev | Requirements | Item |
|----|-----|--------------|------|
| D-10-C04 | warning | OWN-09, TRU-02 | The D-09-51 negative-control verdict flip is unreviewed |
| D-11-02 | info | NAT-04..07 | Six single-function emitters not deleted |
| D-11-27 | info | NAT-07 | No phase claims the D-09-51 flip review |
| D-11-51 | warning | NAT-06 | Shared-leaf diamond call graphs collide on event identity |
| D-12-21 | warning | NAT-06 | D-11-51 interaction anticipated, not absorbed — blocked on D-11-51 |
| D-12-36 | warning | NAT-04..07 | D-11-02's deletion re-deferred, with a stated reversal |
| D-12-43 | warning | RES-03 | Decisive value-divergence control unconstructible against the current representation |
| D-13-02b | warning | DX-06 | B1 contract-violation blame structurally unreachable |
| D-13-10a | warning | DX-07 | `use_matching_argument` withdrawn as unrepairable |
| D-13-34 | warning | — | M001's `testdata/phase6` move and borrow held-out pairs are structurally identical |

Three clusters are worth naming, because they are not ten independent problems:

**Cluster A — the single-function emitter deletion (D-11-02 → D-12-36, plus
D-11-27).** Deferred once in Phase 11, re-deferred in Phase 12 with the reversal
stated openly. Each deferral is individually defensible; the pattern is what
matters, and Phase 10's own D-10-60 ("the no-third-deferral rule") was written
precisely to catch it. **A third deferral would violate a rule this milestone
set for itself.** M003 should either land the deletion or retire D-10-60
explicitly.

**Cluster B — event identity (D-11-51 → D-12-21).** D-12-21 cannot be closed
until D-11-51 is, and neither has an owner. This is a real dependency chain
sitting unowned across a milestone boundary.

**Cluster C — the single-type-per-function invariant (D-13-02b, D-13-10a, and
Phase 13's twin-pair sub-finding).** Three findings, one root cause:
`sameType(ReturnType, Parameter.Type)` is enforced at every function's
admission, so a callee cannot contradict its own contract, all in-scope places
share one type, and the real fix always lands in the caller. These are not
independent debts — they close together, automatically, the moment return type
may differ from parameter type. The blame resolver and its exhaustiveness guard
are already built and waiting.

**D-13-34 is the one item pointing backwards rather than forwards.** M001's
`heldout_move_defect.lang`/`derivation_move_defect.lang` and the borrow pair are
alpha-renames of each other, not structurally distinct programs — a genuine hole
in *shipped* M001 evidence, found only because Phase 13 chose the stricter
retro-strengthening branch and then reported what it found instead of quietly
fixing the fixtures. It affects no live M002 requirement, but it does weaken a
claim M001 already made.

---

## 5. Nyquist Validation Coverage

| Phase | VALIDATION.md | `status` | `nyquist_compliant` | Verdict |
|-------|---------------|----------|---------------------|---------|
| 07 | exists | draft | false | NOT-VALIDATED |
| 08 | exists | draft | false | NOT-VALIDATED |
| 09 | exists | validated | true | COMPLIANT |
| 10 | exists | validated | true | COMPLIANT |
| 11 | exists | draft | false | NOT-VALIDATED |
| 12 | exists | draft | false | NOT-VALIDATED |
| 13 | exists | draft | true | NOT-VALIDATED |

**Overall: partial.** No phase is missing a VALIDATION.md, and no phase is a
genuine PARTIAL (`validated` + `nyquist_compliant: false`). Five phases are
NOT-VALIDATED, which is a coverage TODO — validate-phase never reconciled the
file — not a compliance failure.

**Phase 13 is a special case worth not misreading.** Its `nyquist_compliant:
true` was earned: plan 13-07 re-ran every row of the Per-Task Verification Map
for real, and in doing so found and corrected two ungrounded `-run` patterns
that matched no test — a class of dead verification command that exits 0 forever
because `go test -run` succeeds when its pattern matches nothing. The `status`
field simply stayed `draft` because validate-phase never ran as a skill. The
substance is there; the lifecycle marker is not.

Closing this is cheap and does not require a new phase:

```
/gsd-validate-phase 07
/gsd-validate-phase 08
/gsd-validate-phase 11
/gsd-validate-phase 12
/gsd-validate-phase 13
```

---

## 6. Documentation Defect Found and Fixed

`ROADMAP.md:851` read "30 of 30 M002 requirements mapped to exactly one phase"
while the per-phase table immediately below it sums to 31, and REQUIREMENTS.md
states "M002 requirements: 31 total (OWN-05 split into OWN-05a/OWN-05b,
D-09-37)". The split was applied to the table and to REQUIREMENTS.md but never
to the ROADMAP headline. Corrected to 31 as part of this audit.

This is the same shape as the OWN-08 defect Phase 09's verifier caught: a
derived summary line left stale when the thing it summarizes changed underneath
it. Both were found by cross-referencing, not by reading either document alone.

---

## 7. Verdict

M002 delivered what it set out to deliver. `OpCall` is real at all six dispatch
sites; cross-function loan liveness is derived from signatures alone and
independently re-derived by a peer that cannot see the first derivation; the
oracle, the native emitter, and the interpreter agree across four optimization
tiers; `Result` payloads have one meaning in three engines; and the agent loop
reaches cross-function defects through the JSON protocol.

Two requirements fell short of their stated bar, both for the same structural
reason, both reported honestly rather than papered over, and both ratified.
Ten debt items — three clusters, really — cross into M003 unowned, and one of
them (the emitter deletion) is one deferral away from breaking a rule this
milestone wrote for itself.

Nothing here blocks completing the milestone. The debt disposition is a
judgment call, not a gate.
