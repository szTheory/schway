---
phase: 13-agent-loop-for-interprocedural-defects
verified: 2026-09-13T00:00:00Z
status: passed
score: 3/3 success criteria verified (2 of 3 honestly reported as partially met, per ratified terminal findings)
behavior_unverified: 0
overrides_applied: 0
re_verification: No — initial verification
---

# Phase 13: Agent Loop for Interprocedural Defects — Verification Report

**Phase Goal:** An AI agent hitting a cross-function defect gets a bounded,
correctly attributed cause chain and can repair it through the JSON protocol
alone.
**Verified:** 2026-09-13
**Status:** passed
**Re-verification:** No — initial verification

## Summary Verdict

The three ratified terminal findings (D-13-02b, D-13-10a, D-13-34) are
**accurately recorded** in `PHASE-13-DEBT.md` and `13-VALIDATION.md`, and
nothing was weakened to make a test go green. Every test I re-ran
independently against the tree passed for real, with genuine assertions
(exact outcome strings, sha256-sealed fixtures, non-inert mutation-kill
guards) — not theater. Criterion 1 (DX-05) is fully met. Criteria 2 and 3
(DX-06/DX-07) are honestly partial, exactly as claimed, for the reasons
claimed.

One **additional finding not elevated to `PHASE-13-DEBT.md`** is worth the
developer's attention (see "Cross-Cutting Integrity Findings" below):
13-06 discovered that D-13-28's twin-pair fixture design assumption (the true
fix lives in the shared callee `sink`) was empirically false — the shipped
repair operates entirely inside the caller (`alpha`/`beta`), never `sink`.
This is substantively the same root cause as D-13-02b (already ratified) and
the SUMMARY documents it in full, but it was resolved as a same-plan "Rule 1
deviation" rather than escalated through the one blocking-human checkpoint
the phase used for D-13-33/34. I judge this a documentation/process
observation, not a defect: the substance is already covered by D-13-02b's
ratified language ("criterion 3's twin pair... does not exercise the
contract-boundary rule's distinctive claim"), the tests are honest about what
they now prove, and no control was weakened. Recommend folding this into
`PHASE-13-DEBT.md`'s D-13-02b entry for future readers who read the debt
register but not 13-06-SUMMARY.md.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | DX-05: `lang explain`'s cause DAG names the function for each step, peer-re-derived, with the `narrows` function-scope guard in place and truncation codes unchanged | ✓ VERIFIED | `protocol.ExplainNode.FunctionID/FunctionName`, `ExplainSummary.Functions` exist (`protocol.go:242-243,282`). `resolveExplainFunction` independently derives from `table.coreFuncs` AND `table.astFuncs`, refusing on disagreement (`session_phase6_explain.go:262-296`). `TestExplainFunctionAttribution`, `TestExplainFunctionIdentitySurvivesReorder`, `TestExplainNilSpanOmitsFunction`, `TestNarrowsFunctionScopeGuard(+IsNotInert)` all re-run green. `grep -rn function_budget` returns empty; only `truncated:explain.node_budget`/`truncated:explain.depth` exist. Edge vocabulary still exactly 3 values. `ExplainSchema = "lang.explain/0"` unchanged. |
| 2 | DX-07: `lang-repair` reaches and fixes new interprocedural defect classes through the JSON protocol alone, proven on a held-out split | ⚠️ PARTIALLY MET (honestly, as claimed) | Exactly **two** `MachineApplicable` kinds ship: `move_after_interprocedural_loan` (`check.go:1703`) and `wrap_call_in_try` (`check.go:3645`). `use_matching_argument` is never emitted (`grep -n 'Kind: "use_matching_argument"'` → zero hits); `check.go:3535-3538` documents why and returns `var repairs []diagnostic.Repair` (empty). Both shipped classes reach outcome `repaired` on **sealed held-out** fixtures via `TestRepairDriverFixesEveryDefectClassSinglePass` (re-run green, subtests `interprocedural_loan`, `fallible_consume`). `testdata/phase13/HELDOUT.sha256` verifies against all 5 fixtures (`shasum -a 256 -c` → 5×OK). `cmd/lang-repair/repair.go`/`main.go` byte-unchanged all phase (`git log` shows zero phase-13 commits touching either file). |
| 3 | DX-06: Cross-function blame attribution points at the correct fix location; repair-then-re-check regression covers the non-obvious-function case | ⚠️ PARTIALLY MET (honestly, as claimed) | `TestTwinPairBlame` (both `alpha`/`mirror` subtests) and `TestTwinPairBlameGuard` (the D-13-30d mutation kill) re-run green. `resolveBlame`/`resolveCycleBlame` exist, are exhaustively tested, and are genuinely unwired in production (`grep -c resolveBlame\(` → 1, the definition only) — matches D-13-02b's claim that B1 is structurally unreachable because `sameType(function.ReturnType, function.Parameter.Type)` is enforced at `check.go:255`, `:3148`, `:3399` as an admission precondition independent of any call. `blameFieldWitness` is a genuine compile-time (not runtime) exhaustiveness guard (`check.go:629-646`); `blame_undetermined` exists and routes both sites to `RequiresConfirmation` (`check.go:838-853`). |

**Score:** 3/3 criteria have direct, re-run evidence. Two are honestly
partial, exactly matching the ratified findings' framing — no laundering
observed.

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `protocol.ExplainNode.FunctionID/FunctionName/Blame` | additive fields | ✓ VERIFIED | Present, `omitempty`, documented (`protocol.go:224-246`) |
| `protocol.ExplainSummary.Functions` | whole-function span table | ✓ VERIFIED | `ExplainFunction{ID,Name,Span}` present (`protocol.go:257-266,282`) |
| `resolveExplainFunction` (peer re-derivation) | core+AST independent derivation | ✓ VERIFIED | `session_phase6_explain.go:262-296`, refuses on disagreement via `ErrExplainFunctionResolutionDisagreement` |
| `narrows` function-scope guard | whole-function-span exclusion + same-function_id requirement | ✓ VERIFIED | `session_phase6_explain.go:393-450`; non-inert twin `buildExplainGraphSkippingNarrowsGuard` exists and its test flips red without the guard |
| `resolveBlame`/`resolveCycleBlame` (B1/B2/B3) | contract-boundary blame rule | ✓ VERIFIED (present, correctly unwired) | `check.go:769,832`; `blameFieldWitness` compile-time guard at `:642` |
| `move_after_interprocedural_loan` repair | new MachineApplicable kind, backward-direction-gated | ✓ VERIFIED | `check.go:1636-1707`, `callIsLastUse` gate confirmed |
| `wrap_call_in_try` repair | new MachineApplicable kind | ✓ VERIFIED | `check.go:3645`; span bug found+fixed in 13-06 (verified: full call-expression span, not callee-identifier-only) |
| `use_matching_argument` (withdrawn) | no repair emitted | ✓ VERIFIED | Zero `Kind: "use_matching_argument"` occurrences; honest `unrepairable` path confirmed |
| `testdata/phase13/HELDOUT.sha256` | sealed manifest | ✓ VERIFIED | `shasum -a 256 -c` → 5/5 OK |
| `cmd/lang-repair/repair.go`, `main.go` | byte-unchanged (D-13-32) | ✓ VERIFIED | No phase-13 commit touches either file (`git log` confirms) |

### Key Link Verification

| From | To | Via | Status | Details |
|------|-----|-----|--------|---------|
| `check.go` diagnostics | `cmd/lang-repair` driver | `MachineApplicable` Repair + JSON protocol | ✓ WIRED | `TestRepairDriverFixesEveryDefectClassSinglePass` drives real built `lang` binary end-to-end, exact outcome strings asserted |
| `session_phase6_explain.go` | `protocol.ExplainNode` | function attribution fields | ✓ WIRED | `applyExplainFunction` sets fields from peer-verified resolution |
| `resolveBlame` | production emission sites | — | ⚠️ NOT WIRED, BY DESIGN | Confirmed unwired (grep count 1); this is the documented, ratified consequence of D-13-02b, not an oversight |

### Behavioral Spot-Checks / Re-run Tests

All named in the verification context's "verify these specifically" section,
plus the full VALIDATION.md map's `-run` patterns, were spot-checked or
re-run directly against the tree (not read from a SUMMARY claim):

| Test | Result |
|------|--------|
| `TestBlameClassifiesEveryInterproceduralCode` | ✓ PASS |
| `TestRepairDriverFixesEveryDefectClassSinglePass` (7 subtests) | ✓ PASS |
| `TestTwinPairBlame` (`alpha`, `mirror`) | ✓ PASS |
| `TestTwinPairBlameGuard` | ✓ PASS |
| `TestUseMatchingArgumentUniquenessGate` (3 subtests) | ✓ PASS |
| `TestUnrepairableDefectFailsTheGate` (3 subtests) | ✓ PASS |
| `TestWrapCallInTryReverifyFailed` | ✓ PASS |
| `TestHeldoutOutcomeSetContainsNoLaundering` | ✓ PASS |
| `TestRepairDriverSourceNeverReferencesHeldoutFixtures` + not-inert | ✓ PASS |
| `TestCorpusTopologyDisjoint` (2 subtests) + not-inert | ✓ PASS |
| `TestHeldoutCorpusSealed` + not-inert | ✓ PASS |
| `TestNarrowsFunctionScopeGuard` + not-inert | ✓ PASS |
| `TestPhase6DefectCorpusIsHeldOut` (`match` PASS, `move`/`borrow` SKIP with D-13-34 reason) + not-inert | ✓ PASS (as claimed — narrow skip, not a blanket regression) |
| `TestRepairDriverImportsStayOutsideInternal` + not-inert | ✓ PASS |
| `go build ./...` | ✓ PASS |
| `shasum -a 256 -c testdata/phase13/HELDOUT.sha256` | ✓ 5/5 OK |
| `git diff --quiet -- cmd/lang-repair/repair.go cmd/lang-repair/main.go` (whole phase) | ✓ clean |
| `git diff --quiet -- testdata/phase6/` | not independently re-checked this run, but `TestPhase6DefectCorpusIsHeldOut`'s skip messages and D-13-34's explicit record are consistent with no edits |

No `go test -run` pattern found anywhere in `13-VALIDATION.md`'s Automated
Command column or in any `<verify><automated>` block across the seven PLAN
files matches zero tests — I spot-ran a broad sample and every pattern
resolved to real, existing test functions. The two historically-broken
patterns 13-07 found and fixed (`TestOrderingStability`,
`TestImportBoundary`) do not recur elsewhere.

### Requirements Coverage

| Requirement | Status | Evidence |
|---|---|---|
| DX-05 | ✓ SATISFIED (Complete, as REQUIREMENTS.md states) | Function attribution, peer re-derivation, narrows guard, stable truncation codes — all directly verified in the tree |
| DX-06 | ⚠️ PARTIAL, as REQUIREMENTS.md states | Contract-boundary rule implemented, tested, and correct but structurally unreachable (B1) at this language maturity — ratified terminal finding D-13-02b, confirmed by direct code inspection (`sameType` admission-time precondition) |
| DX-07 | ⚠️ PARTIAL, as REQUIREMENTS.md states | Two of three intended classes ship and are proven `repaired` on held-out fixtures; `use_matching_argument` withdrawn as a structurally-unrepairable no-op — ratified terminal finding D-13-10a, confirmed by direct code inspection (zero emission sites) |

No orphaned requirements found — DX-05/06/07 are the only three mapped to
Phase 13 and all three are addressed.

### Anti-Patterns Found

Scanned every file touched this phase (`check.go`, `protocol.go`,
`session_phase13_injectors.go`, `session_phase6_explain.go`,
`session_phase6_injectors.go`, `repair_test.go`, `antitheater_test.go`) for
`TBD|FIXME|XXX|TODO|HACK|PLACEHOLDER`: **zero matches.** No debt markers, no
unresolved theater flags in shipped code.

### Cross-Cutting Integrity Findings

- **Controls were not weakened to pass.** The `testdata/phase6` structural
  predicate stays strengthened and unweakened; `move`/`borrow` subtests are
  narrowly `t.Skip`'d (not the whole test), `match` still runs and passes.
  `use_matching_argument` was not downgraded to a non-`MachineApplicable`
  applicability to dodge a red test — it emits nothing, and the driver
  honestly reports `unrepairable`.
- **QLT-08 D-13-30(a)-(d) seams all present and independently re-run green**,
  including the withdrawn `use_matching_argument` correctness seam, which is
  recorded as withdrawn with a reasoned justification rather than silently
  dropped.
- **One finding not folded into `PHASE-13-DEBT.md`:** 13-06 empirically
  discovered that D-13-28's twin-pair fixture design assumption (true fix
  lives in shared callee `sink`) was wrong — the shipped repair for
  `move_after_interprocedural_loan` operates entirely within the mutated
  caller's own body for both fixture halves, never inside `sink`, because
  this is a second instance of the same `sameType`-admission-precondition
  structural property behind D-13-02b. This is documented thoroughly and
  honestly in `13-06-SUMMARY.md`'s "Criterion 3 verdict" section, and the
  resulting tests (`TestTwinPairBlame`/`TestTwinPairBlameGuard`) were built
  around the verified behavior rather than the mistaken assumption — so
  nothing shipped is dishonest. But `PHASE-13-DEBT.md`'s three-item register
  does not name this specific sub-finding, and it was resolved as a
  same-plan "Rule 1 deviation" rather than routed through the phase's one
  `gate="blocking-human"` checkpoint (which covered only D-13-33/34).
  Recommend a short addendum to `PHASE-13-DEBT.md`'s D-13-02b entry pointing
  at 13-06-SUMMARY.md, so a reader of the debt register alone (not every
  SUMMARY) gets the full picture. This does not change DX-06's already-
  "Partial" disposition and is not a blocker.

### Human Verification Required

None. Every must-have in this verification context was checkable directly
against source and by re-running named tests.

## Gaps Summary

No blocking gaps found. The phase delivers exactly what it honestly claims:
criterion 1 fully met; criteria 2 and 3 partially met for well-evidenced,
ratified, structural reasons (language maturity, not engineering shortfall).
The one process note above (twin-pair assumption correction not mirrored in
`PHASE-13-DEBT.md`) is a documentation-completeness recommendation, not a
gap blocking phase closure.

---

_Verified: 2026-09-13_
_Verifier: Claude (gsd-verifier)_
