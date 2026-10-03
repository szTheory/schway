# Phase 26 wave gates

## Wave 1 — 2026-10-02

Plan 26-01's focused hosted run 37073989672 passes on Ubuntu and macOS at
`b2034083c2274a2d1d301394def0835047dd5eda`. Summary and task commits exist;
the source witness and scalar checker helper exist; the summary self-check
passes. Schema drift and UI gates report no block; the codebase drift gate
has no structure map to inspect.

The broader integration run 37074516513 at
`1de217f878a6253415c1524d2190d40cf310086f` passes vet and build on both hosts,
but fails the full suites and evidence aggregates. No full-suite pass or race
pass is claimed. Failures identify:

- Four new scalar-operation fields absent from the closed core inventory.
- The emitter consumer moving from line 979 to 1006 in core_test.go.
- The new source witness changing the corpus to 152 programs / 4,948 lines.
- M004 archive moves absent from the Phase 23 contract lookup, Phase 24
  evidence links, and Phase 23/25 groundedness locator records.
- Phase 26's still-draft validation document and test selectors owned by the
  four remaining plans.

The first four categories are repaired in this wave's integration correction.
D-14-144 receives a dated locator amendment; historical validation documents
and receipts are preserved. Its generated reconciliation row is re-derived
from the authored record. The repairs have source/diff checks only until the
next hosted run.

The active phase's unimplemented test selectors and draft status remain
visible failures. Execution continues to their named owning plans under the
project's autonomous execution setting; no test is suppressed or marked
passed. The final phase gate must rerun full hosted validation after those
plans and the validation document are complete. Wave 1 full-gate failure
count: 1.

## Wave 2 — 2026-10-02

Plans 26-02 and 26-03 have focused hosted passes on both Ubuntu and macOS:
run 37081675751 at `1677ca3799f18fe2b699943773149560ce89ec57` and run
37085604802 at `d653312ee81917ada28c932ca84ba90f7381af96`. Their summaries,
artifacts, commits, and self-checks are present. Schema and UI gates report
no block; codebase drift has no structure map to inspect.

The broader run 37085873013 at
`0368ec65dfa7814ce5b4fa1fa5737d0f2f99eabc` passes vet and build on both hosts.
Both full suites and Phase 6 evidence aggregates fail; race checks are skipped.
The Phase 23, 24, and 25 aggregates pass. Remaining failures are:

- Shifted C emitter test call locations and the new checked-add conformance
  consumer. This integration correction updates their explicit registry.
- Five stale unparseable frontier entries in current milestone research files.
  The hosted groundedness test confirms they no longer occur; this correction
  removes those five entries without adding exceptions for new findings.
- Active Phase 26 validation selectors for Plans 04/05, mixed-package regex
  alternatives, and draft lifecycle status. Plan 05 and phase closeout own
  resolving these against implemented tests and actual hosted receipts.
- The checked-in validation-corpus receipt no longer matches the current
  corpus digest. Refresh it through the hosted receipt job after finalizing
  the active validation rows; do not synthesize a receipt locally.

Metadata corrections have source/diff checks pending the next hosted run.
No broad-suite or race pass is claimed. Full-gate failures remain visible
while the remaining owning plans execute; the final phase gate must pass.
Wave 2 full-gate failure count: 1.

## Wave 3 Plan 04 integration correction — 2026-10-03

Plan 26-04's initial full gate, run 37089290645 at `8894da69d816aeca566197a109a9a66d6cdf8330`, failed on Ubuntu and macOS. It exposed an unconditional schema-2 writer extension that changed legacy generated C and required `uint64_t` declarations in programs that do not use U64, along with missing identifier reservations, comparison routing for `Execution.Events.Occurrence`, a changed stable `duplicate_pair` diagnostic, and stale public-emitter consumer line references. The original and full failure evidence is retained in `26-04-integration-red-evidence.json`.

Commits `fa85fac` and `9784f76` gate occurrence support on an eligible cyclic scalar-copy site, preserve previous writer output otherwise, restore the peer's stable duplicate-pair diagnostic, add occurrence to semantic comparison routing, classify generated identifiers, align every affected emitter callsite, and expand focused CI to include the affected cgen, executionpeer, and session packages. Compile-only build and source checks passed locally. Focused hosted run 37090502900 passed cgen and executionpeer full packages plus the Phase 26 event, interpreter/native repetition, authority, ordering, comparison, and public-emitter inventory controls on both hosts. Its full session package still fails on the active receipt/frontier/reconciliation/lifecycle rows; Ubuntu's focused session job also reports existing Phase 5 verification failures because its Clang invocation cannot find `-lc++`.

The final full hosted gate 37090678253 passes vet and build on both hosts. Full `go test ./...` fails on the same active Phase 26 validation blockers on Ubuntu and macOS:

- `TestValidationRowGradesAreEarnedOverArchivedCorpus`: checked-in digest `14f14924a5cb4f30c3f84db5a3dbf9b47c3265bd6bd5dbafecd50ccc926a3b93` differs from current digest `af2dafc4f3f15ecb2087e41b8fddc4ca08fff37232f55bc1ed7904c58499da4d` across 34 pairs.
- `TestReconciliationVerdictsCarryTheirObligations`: an unreconciled R2 finding remains at `26-VALIDATION.md:49`.
- `TestVerificationGroundednessFrontierIsPinned`: active R2b findings remain at lines 42–47 and 50 of `26-VALIDATION.md`.
- `TestPhase20ValidationLifecycle`: `26-VALIDATION.md` remains draft.
- `TestVerificationGroundednessThreeClassesAreEmpty`: reports R1=0, R2=1, R3=0, and R2b=52.

The Phase 23, 24, and 25 evidence aggregates pass. Phase 6 aggregates fail on the active Phase 26 findings. Race checks and dependent Phase 15 aggregate checks were skipped after test failures; the validation corpus receipt and Phase 26 workflow jobs were skipped in the full workflow. No full-suite or race pass is claimed, and none of these active findings were suppressed. Plan 05 and phase closeout own the validation rows, lifecycle transition, and hosted receipt refresh.

Plan 04's loop tests inject `OpCopy` into checked core. They verify core/backend occurrence behavior and do not establish the requested public source witness `let snapshot = i`; Plan 05 owns that source-level consumer and must add it before claiming the source witness.
