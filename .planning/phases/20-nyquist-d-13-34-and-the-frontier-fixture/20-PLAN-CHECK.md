# Phase 20 Plan Check

**Phase:** 20 — Nyquist, D-13-34, and the Frontier Fixture  
**Plans checked:** 8 in revised set
**Result:** PASS WITH WARNING — 0 blockers, 1 scope warning in the eight-plan goal-backward recheck

The prior PASS applied to the seven-plan pre-integration set. Wave 1 revealed source-derived maturity, emitter-registry, and skip-witness changes plus a red full suite; the eight-plan revision below records the source-grounded goal-backward recheck.

## Prior seven-plan sign-off (superseded by Wave 1 integration)

### Goal and requirement coverage

All five roadmap success criteria and all four requirements have executable plan coverage:

| Requirement / criterion | Coverage | Assessment |
|---|---|---|
| QLT-10: reconcile named records and clear every draft status | 20-02 reconciles 07/08/11; 20-03 reconciles the other eight records and adds a full-corpus zero-draft guard; 20-07 repairs four remaining cited rows | Covers all 11 draft files in the research inventory; statuses and grades remain evidence-dependent. |
| QLT-11: refused checksum fixture with a moved diagnostic | 20-01 | Exact source digest, executable intent witnesses, current refusal, and same-byte historical comparison are specified. The record correctly calls the M003-open result a reconstruction from commit `d21db90e67750bb19976c4206f4c23c68cd06207`, not a contemporaneous pin. |
| QLT-12: content-address closure work and lower repeated cost | 20-04 | Key covers source, transitive/system headers, SDK/sysroot, link inputs, compiler/toolchain, target and flags; incomplete manifests bypass reuse; binaries alone are cached and verdicts/comparisons run fresh. Timing helper measures cold/warm full-suite runs against both Phase 14 references. |
| PRC-02: no more than five open unowned items | 20-05 and 20-07, then 20-06 | The denominator is the ten-ID M002 cohort plus seven distinct live M003 rows (13 at start). Three evidence-checked closures and a conditional, source-supported P21 owner reduce it to nine; four QLT-10 repairs must then bring it to five before the D-13-34 choice. If the P21 ownership is unsupported, the plan requires another real disposition rather than waiving the cap. |
| D-13-34 and D-06-29 | 20-06 | Preparatory evidence precedes a `blocking-human` checkpoint; either selected outcome explicitly handles the inference and preserves the existing structural predicate. The human decision is not used to meet the debt cap. |

The phase directory has a `20-VALIDATION.md`. It has 18 task rows for 19 plan tasks; the omitted row is the manual checkpoint 20-06-02, explicitly documented as having no automated result. All plan requirements fields account for QLT-10, QLT-11, QLT-12, and PRC-02. Research A1/A2 are resolved with the population and historical-evidence limits recorded. No Phase 20 CONTEXT.md, REVIEWS.md, or project skill directories were found. The responsibility map and patterns align with the planned Go/session, fixture, cache, and debt-register seams.

## Structure and ordering

All seven plans specify 2–3 tasks and have complete task fields. Dependencies are acyclic and waves are consistent: 20-03 follows 20-02; 20-07 follows 20-03 and 20-05; 20-06 follows 20-07. Other plans are Wave 1. Same-wave file ownership is separated. The 20-06 checkpoint is correctly marked `autonomous: false`; the subsequent task is explicitly conditional on the human's selection.

Plan estimates range from 12,000 to 26,000 tokens, below the 100,000-token budget. Confidence is low for all estimates because there are no project calibration samples, so the figures are rough. Plan 20-03 touches 10 files, at the warning threshold.

## Final command-map check

The 18 `20-VALIDATION.md` command cells use exact test names, named prefixes, or separate commands joined by `&&`; none depends on a Markdown escape becoming shell syntax. An independent parser confirmed all 18 table rows have nine cells, one intact command code span, no pipe or backslash, and shell-parseable commands. The D-13-34 task remains a `blocking-human` checkpoint, and no plan selects an outcome in advance.

### Final independent sign-off

Plan 20-02 Task 1 now identifies `R1/R2/R3=0`, reconciled `R2b=23`, raw `R2b=24`, 639 documents, and 708 commands as the historical research snapshot. It requires a fresh execution-start scan of the now-planned tree, records exact findings and provisional owners, and explicitly permits the global frontier test to be red before reconciliation. The measured 2026-09-25 planning snapshot (`R2=10`, `R2b=25`, 651 documents, 737 commands) is a point-in-time observation, not a required execution count. Plan 02's task verifications now run the passing non-vacuous corpus check; Plan 03 enforces zero draft while preserving interim findings; Plan 07 is the first task that requires the exact global frontier and owner gate green, after new Phase 20 tests and source-row repairs land. The validation map and research distinguish the two snapshots and this sequence.

Independent sign-off confirms that the sequence is sound: Plan 02 records the historical research baseline separately from a fresh execution-start scan and does not demand a green old pin; Plans 02/03 permit the expected interim red state while recording exact findings and owners; Plan 07 runs after those reconciliations and is the first plan requiring the exact global frontier and ownership gates to pass. The measured current planning scan (R2=10/R2b=25, 651 documents, 737 commands) is identified as a point-in-time observation, not an execution requirement.

The historical checksum comparison is accurately bounded: no contemporaneous M003-open fixture pin existed, so Plan 01 reproduces the old checker result for identical bytes and explicitly labels it reconstructed. The cache plan requires a complete input manifest and bypass on undiscoverable inputs, keeps execution/comparison fresh, and records corruption/invalidation controls. The debt plan keeps the audit cohort, live M003 additions, generated-view subset, and raw archive count distinct; it specifies a source-derived 13-item population, evidence-backed disposition path, seeded overflow control, and retains D-13-34 as a blocking human choice. The 11 draft files and the copy-safe VALIDATION command map are covered. All seven plan structures validate; dependencies are acyclic and consistent, each automated task has Files/Action/Verify/Done, and 20-06 preserves the human checkpoint.

### Timing-distribution sign-off

Plan 20-04 Task 3 specifies three paired cold/warm full-suite runs after prewarming a private Go build cache. Every run uses `-count=1`; each pair has a fresh empty closure-artifact cache followed by a warm reuse of that same cache. The helper is bounded to six full-suite invocations and 30 minutes total, with incomplete or failed pairs reported as non-passing evidence. `20-CLOSURE-TIMING.md` must show all six raw samples, min/median/max for cold and warm, paired deltas, exact environment facts, and warm-median comparisons against both Phase 14 references. Phase 14's recorded measurement is a cold `go test ./... -count=1` run on `machine:4797d76b7863` using Go 1.24.0 on darwin/arm64; the plan requires a non-comparable-baseline blocker when host/toolchain conditions differ. Three paired samples provide the requested distributions and expose spread while correctly avoiding a high-confidence percentile claim.

### Verification-direction sign-off

Each of the 18 runnable `<automated>` commands across Plans 20-01 through 20-07 has an immediately adjacent, non-empty `<fails_when>` sibling. The statements identify non-zero exit or Go's no-tests-selected output and the task-specific rejected condition, including a wrong checksum diagnostic, empty scanner corpus, draft status, cache reuse on changed input, incomplete timing pairs, a seeded sixth debt item, and stale frontier ownership. Independent inspection confirms the pair count and specificity; all seven plans pass `verify.plan-structure` without errors or warnings.

```yaml
issues_pending_recheck: []
```

### Final structural and command checks

All seven `verify.plan-structure` results are valid with no errors or warnings. The 18 automated rows in `20-VALIDATION.md` each have nine table cells, one intact command code span, and pass `bash -n`; the manual 20-06-02 checkpoint is intentionally absent from automated rows. Every automated check in the seven plans has a stated failing direction.

**Final independent sign-off: PASS.** All prior sign-off properties remain present in the revised plans: cold/warm timing distributions use three paired full-suite samples and compare against both Phase 14 references; Plan 02 separates historical and fresh execution-start groundedness snapshots and permits the expected interim red pin; validation commands remain executable and are sequenced so Plan 07 performs the final exact frontier/owner gate; the four requirements and five roadmap success criteria have coverage; the debt denominator is source-derived from the ten-ID M002 cohort plus seven distinct live M003 rows, with an evidence-backed disposition path and seeded cap control; the checksum comparison labels the historical result as reconstructed from the named commit; the closure cache key covers source, transitive/system headers, SDK/sysroot, link inputs, compiler/toolchain, target, and flags and bypasses reuse for incomplete inputs; and D-13-34 remains a blocking-human decision with both outcomes prepared from evidence.

All seven plan files validate structurally, their declared dependencies are acyclic and wave-consistent, and every automated command has a specific adjacent `<fails_when>`. No blocker, warning, or advisory remains for plan execution. Preserve the blocking-human D-13-34 checkpoint and exact frontier/owner equality at Plan 07.

## Wave 1 integration revision awaiting review

Plan 04 remains Wave 1 and now implements plus smoke-tests the bounded timing harness only. Its full-suite timing report moved to new autonomous Plan 08, Wave 4, after Plan 07's maturity, emitter-consumer, skip-witness, archive, debt, and exact-frontier reconciliation. Plan 08 first requires an unfiltered green `go test ./... -count=1` preflight, then runs three paired cold/warm full-suite samples. Plan 06's D-13-34 human checkpoint moves to Wave 5 after that timing plan. Plan 07 explicitly depends on Plans 01, 03, 04, and 05, so its source-derived registry updates use the final checksum fixture and cache code.

The revised `20-VALIDATION.md` has 21 automated rows; the decision checkpoint remains manual. A local read-only pass found all eight plan frontmatters and structures valid without warnings, all 21 commands paired with immediate failing directions and shell/regex parseable, and all 21 table commands copy-safe without Markdown escape conversion. These checks do not claim the full suite is green while Wave 1 is still in progress. Plan 07 starts only after Plan 04's final source is committed and its integration inputs are remeasured.

### Goal-backward recheck of the eight-plan revision

| Phase outcome | Executable path | Decision |
|---|---|---|
| Zero draft VALIDATION records and truthful 07/08/11 evidence | Plans 02/03 reconcile the eleven drafts; Plan 07 repairs four remaining source rows, corpus digest, exact frontier, and generated views | Covered |
| Refused checksum source with a moved diagnostic | Plan 01 checks in the exact source and compares current refusal with the explicitly reconstructed M003-open result | Covered |
| Content-addressed 112-program closure with fresh verdicts and measured latency | Plan 04 builds and smoke-tests the cache/harness; Plan 07 restores source-derived guards; Plan 08 requires a green full suite before three paired cold/warm measurements | Covered |
| No more than five open unowned debt items | Plan 05 derives the thirteen-item source set; Plan 07 closes/owns evidenced overflow, verifies the seeded cap, and keeps D-13-34 counted until decision | Covered |
| D-13-34 and D-06-29 | Plan 06 prepares both concrete outcomes, blocks for a human choice, then verifies the selected fixture/debt and inference result | Covered |

The dependency graph is acyclic and exact-wave consistent: Plans 01/02/04/05 are Wave 1, 03 Wave 2, 07 Wave 3, 08 Wave 4, and 06 Wave 5. There is no same-wave modified-file overlap. The live guard outputs directly identify Plan 07's assigned work: maturity 140/4602 stated versus 141/4639 derived; emitter registry stale line 468 versus current call line 501; four uncited native-test skips plus two Phase 20 cache skips; and 54 live reconciliation findings versus 66 ledger rows requiring reviewed stale-entry removal. Plan 07 depends on Plan 04 so those positions are re-derived after its final source settles. Every automated task has an adjacent failure direction, and all eight structures validate.

**Warning — Plan 07 scope:** it modifies 14 unique files across three tasks. This exceeds the plan-checker's 10-file warning threshold but remains below its 15-file blocker threshold. The work is sequenced within the plan: archived-row repair, source-derived inventory/witness reconciliation, then debt/frontier close. If execution context degrades, split the inventory task into a dependent plan without weakening its guards. Plan 04's source remains uncommitted at this review point, so the exact call positions and skip count must be checked once more after its final commit.

Final source-grounded pass after Plan 04 committed used a fresh 112-closure paired run (cold: 0 reused/112 recomputed; warm: 112 reused/0 recomputed) and passing deterministic timing-harness smoke test. Direct reruns of the current maturity and emitter guards fail exactly at the Plan 07 targets (140/4602 stated versus 141/4639 derived; stale EmitNative line 468 versus live line 501). The suppression guard reports four uncited `native_test.go` skips and two uncited Phase 20 cache skips. Plan 07 explicitly re-derives exact call and skip locations after source edits. This source check changes no issue classification: zero blockers, one Plan 07 scope warning.

```yaml
issues:
  - plan: "20-07"
    dimension: scope_sanity
    severity: warning
    required_property: "A plan's source and evidence edits fit comfortably within the execution context budget"
    description: "Plan 07 touches 14 unique files, above the 10-file warning threshold and one below the 15-file blocker threshold."
    fix_hint: "If context pressure appears, move the five-file maturity/emitter/skip inventory task into a dependent plan while preserving its existing guards and timing order."
```
