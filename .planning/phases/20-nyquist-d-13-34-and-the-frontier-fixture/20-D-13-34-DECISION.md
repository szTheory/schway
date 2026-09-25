# D-13-34 Decision Brief — M001 Held-Out Move/Borrow Pairs

**Status:** selected and implemented — **replace** (user choice, 2026-09-25).

## Live evidence

Ran:

```sh
GOCACHE=/private/tmp/phase20-gocache GOMAXPROCS=4 go test ./internal/compiler/session -run 'TestPhase6(HeldoutPairsAreAlphaRenamesOnly|DefectCorpusIsHeldOut)$' -count=1 -v
```

Both controls passed. `TestPhase6HeldoutPairsAreAlphaRenamesOnly` confirms that the weakness is still present; `TestPhase6DefectCorpusIsHeldOut` skips the two known-identical pairs and continues checking the structurally distinct match pair.

| Class | Held-out five-part summary | Derivation five-part summary | Result |
|---|---|---|---|
| Match | `{bindingCount:0, matchArmCount:3, borrowCount:0, takeCount:0, maxDepth:1}` | `{bindingCount:0, matchArmCount:2, borrowCount:0, takeCount:0, maxDepth:1}` | Distinct by one match arm |
| Move | `{bindingCount:2, matchArmCount:0, borrowCount:0, takeCount:2, maxDepth:1}` | `{bindingCount:2, matchArmCount:0, borrowCount:0, takeCount:2, maxDepth:1}` | Identical; identifier rename only |
| Borrow | `{bindingCount:3, matchArmCount:0, borrowCount:3, takeCount:0, maxDepth:1}` | `{bindingCount:3, matchArmCount:0, borrowCount:3, takeCount:0, maxDepth:1}` | Identical; identifier rename only |

The predicate is the existing identifier-independent tuple `(bindingCount, matchArmCount, borrowCount, takeCount, maxDepth)`. The current move fixtures are two-hop take chains; the borrow fixtures contain two shared borrows of one owner and a reborrow that keeps the first loan live. The injectors currently prove one `ownership.use_after_move` or one `ownership.borrow_conflict`, respectively, with a driver-eligible repair.

### D-06-29 claim today

D-06-29 says the held-out split blunts but cannot eliminate repair overfitting risk because three classes and six fixtures provide a small reachable-input space. The current evidence does **not** support the same structural-distinctness claim for all three classes: only match is distinct under the retro-strengthened predicate. The move and borrow examples are not independent structural holdouts. Their successful injector/repair checks still show behavior on those fixtures, but do not establish structural generalization beyond identifier spelling.

## Choice A — replace only the move and borrow held-out sources

This preserves the derivation fixtures and the five-component predicate. The proposed diffs add one operation to each held-out program, keeping the marker on the operation that the injector should corrupt:

```diff
--- a/testdata/phase6/heldout_move_defect.lang
+++ b/testdata/phase6/heldout_move_defect.lang
@@
-  let echoed = take delivered // lang:move-target
-  echoed
+  let echoed = take delivered
+  let returned = take echoed // lang:move-target
+  returned
```

Expected move summaries after implementation and measurement:

- Proposed held-out: `{bindingCount:3, matchArmCount:0, borrowCount:0, takeCount:3, maxDepth:1}`.
- Existing derivation: `{bindingCount:2, matchArmCount:0, borrowCount:0, takeCount:2, maxDepth:1}`.
- The injector changes the marked third take back to the already-moved `buffer`; acceptance still requires the clean source to check and the mutated source to produce exactly one `ownership.use_after_move` with a driver-eligible repair.

```diff
--- a/testdata/phase6/heldout_borrow_defect.lang
+++ b/testdata/phase6/heldout_borrow_defect.lang
@@
   let second = borrow buffer // lang:borrow-target
   let reviewed = borrow first
+  let audited = borrow reviewed
   second
```

Expected borrow summaries after implementation and measurement:

- Proposed held-out: `{bindingCount:4, matchArmCount:0, borrowCount:4, takeCount:0, maxDepth:1}`.
- Existing derivation: `{bindingCount:3, matchArmCount:0, borrowCount:3, takeCount:0, maxDepth:1}`.
- The added reborrow uses the first loan after the marked second borrow; acceptance still requires the clean source to check and escalation of the marked borrow to produce exactly one `ownership.borrow_conflict` with a driver-eligible repair.

These diffs are proposals, not verified fixtures. If chosen, the implementation must run the existing injection controls and assert the resulting summaries differ; if either proposed source fails those controls, adjust it before closing D-13-34.

As a feasibility check only, both proposed sources were written under `/private/tmp` and passed the current `lang check` CLI. They were not installed into `testdata/phase6`; the injector behavior and five-component summaries remain to be verified only if this branch is chosen.

**Expected disposition:** close D-13-34 only after those tests pass and commit the updated fixtures and structural probe. Restore D-06-29's structural inference for move and borrow while retaining its stated small-corpus residual risk. The open-unowned debt population would fall from five to four because this item is closed.

## Choice B — re-ratify the measured weakness and assign it to M004 Phase 21

This preserves every shipped fixture byte and the current predicate. It explicitly withdraws D-06-29's move/borrow structural-generalization inference; only the match class remains structurally distinct. The injector checks remain useful for those exact examples, but they do not establish diverse held-out structure.

Proposed D-13-34 debt row disposition and detail text:

```text
Landing phase: P21 (M004 — Native Emission Ownership and Resource Discharge)
Status: OPEN, owned by P21; this is not a closure.
Decision: Re-ratify the measured alpha-rename weakness. Preserve the M001 fixture bytes and D-13-33 five-component predicate. D-06-29's structural held-out inference is withdrawn for move and borrow; it applies structurally only to match. Keep the explicit limitation that six fixtures across three classes cannot establish generalization over the language's full program space. P21 owns a future review of whether the widened post-M003 source surface warrants structurally distinct replacements; ownership does not imply the debt is resolved.
Reopening/exit evidence: a P21 review either supplies structurally distinct valid held-out move and borrow fixtures that preserve the original injector defects and pass the five-component control, or records a further explicit re-ratification based on measured evidence.
```

P21 is the planned M004 phase named in the live roadmap and already used as a landing-phase owner for other M004 evidence debt. Under this choice, D-13-34 remains open, now assigned; it leaves the open-unowned population by ownership, not closure. The PRC-02 cap must be recalculated from the live ledger after the choice.

## Human choice required

Choose **replace** to restore structural separation for both classes, or **reratify** to preserve the shipped examples and explicitly narrow D-06-29 under P21 ownership. After the choice, Plan 06 Task 3 will implement only that branch and rerun the held-out, injector, debt-register, and live cap controls.

## Selected outcome and implementation evidence

The user selected **replace**. The held-out move and borrow sources now include
the proposed extra operation; the derivation fixtures and five-component
predicate remain unchanged. `TestPhase6DefectCorpusIsHeldOut` and
`TestPhase6HeldoutPairsAreStructurallyDistinct` passed: move summaries are
`{bindingCount:3, matchArmCount:0, borrowCount:0, takeCount:3, maxDepth:1}` vs
`{bindingCount:2, matchArmCount:0, borrowCount:0, takeCount:2, maxDepth:1}`;
borrow summaries are `{bindingCount:4, matchArmCount:0, borrowCount:4, takeCount:0,
maxDepth:1}` vs `{bindingCount:3, matchArmCount:0, borrowCount:3, takeCount:0,
maxDepth:1}`. Match remains 3 arms vs 2. The move and borrow injector controls
passed with exactly their intended defect. D-06-29's structural inference is
restored for all classes while the small-corpus residual is retained.

The live D-13-34 row is closed as `CLOSED(P20)` and the current open-unowned
population is four. Final commit and full-suite evidence are in
`20-06-SUMMARY.md`.
