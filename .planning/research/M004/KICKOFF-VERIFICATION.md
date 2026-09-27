# M004 Kickoff Verification

Date: 2026-09-27. Scope: milestone planning, living roadmap, historical evidence
disposition, and GSD routing. No production compiler capability was implemented.

## Artifact checks

- Four new phases, 22–25, each with five observable success criteria and a
  runnable source/input/output witness. Phase 21 is completed archived prework.
- 24 unique pending requirements; each occurs once in traceability and once in
  the matching roadmap phase's Requirements field; zero orphans/duplicate owners.
- Local Markdown link targets in PROJECT, PRODUCT-ROADMAP, LANGUAGE-MATURITY,
  REQUIREMENTS, ROADMAP and the five M004 research files resolve.
- All five research artifacts are substantive and have no continuation sentinel.
- SDK milestone switch: M003 → M004; outgoing cleanup cleared zero directories.
  The Phase 21 archive and completed UAT were not moved or replayed.
- Four empty phase directories retain routing placeholders only; no plans,
  contexts, execution summaries, or passing implementation claims were invented.
- `init.progress`: M004, four `not_started` phases, next phase 22.
- `roadmap.analyze`: phases 22–25, no missing details, zero completed phases/plans.
- `state.validate --strict`: `valid: true`, no warnings. STATE frontmatter has
  current phase 22 and four total phases. The installed canonical
  `state-contract.cjs` publisher refreshed state.json; its next action is
  `/gsd:progress --next`, labeled `plan phase 22`.
- `git diff --check`: passed.

The installed runtime's `state.sync` updates STATE.md but does not itself publish
state.json. The exported canonical publisher was invoked after sync; no second
state schema or handwritten routing answer was introduced.

## Focused regression checks

Command (12 named tests, plus their subtests):

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestLanguageMaturityCountsAreCurrent|TestSelfDescribingDocsGuardIsNotInert|TestPhase16EmitterCutsAreAmendedAndOwned|TestDebtRegistersAreWellFormed|TestDebtRegisterIdentifiersAreGloballyUnique|TestPhase20UnownedDebtPopulation|TestPhase20UnownedDebtCapRule|TestUnreachableClaimsViewIsCurrent|TestUnreachableClaimsViewCatchesHandEdit|TestUnreachableClaimsViewCatchesStaleEntry|TestD1243ControlIsUnconstructible|TestPhase18WrongSlotMutation)$' -count=1 -v
```

Result: all 12 passed after the planning/debt/maturity edits. The Phase 16
historical owner guard retains its four seeded mutations and now reads the M003
archive instead of requiring the live M004 charter to remain provisional.
D-12-43's two current wrong-slot witnesses passed; its authored disposition is
`CLOSED(d5ddb04)` / `MUTATION-KILLED`, with the original Phase 12 narrative and
ratification preserved. The generated claims view came from its canonical test
renderer, not an independently authored replacement. Only the current unowned
cohort changed from four to three; historical ten/13-item cohorts remain intact.

## Evidence limits

No full compiler suite, race lane, complete validation corpus, cross-host run,
new resource application, or Phase 21 UAT was run for this kickoff. The focused
checks support these planning repairs, not M004 runtime acceptance. Dated edits
touch inputs of archived Phase 16/21 report fingerprints; those reports retain
their recorded revision scope and were not relabeled as current verification.

All M004 requirements remain Pending. Next authorized planning action:
`$gsd-plan-phase 22` using the adopted milestone research and living roadmap.
