# Phase 20 Plan Check

**Phase:** 20 — Nyquist, D-13-34, and the Frontier Fixture  
**Plans checked:** 7  
**Result:** REVISION READY — final independent sign-off pending

## Goal and requirement coverage

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

### Resolution submitted for independent sign-off

Plan 20-02 Task 1 now identifies `R1/R2/R3=0`, reconciled `R2b=23`, raw `R2b=24`, 639 documents, and 708 commands as the historical research snapshot. It requires a fresh execution-start scan of the now-planned tree, records exact findings and provisional owners, and explicitly permits the global frontier test to be red before reconciliation. The measured 2026-09-25 planning snapshot (`R2=10`, `R2b=25`, 651 documents, 737 commands) is a point-in-time observation, not a required execution count. Plan 02's task verifications now run the passing non-vacuous corpus check; Plan 03 enforces zero draft while preserving interim findings; Plan 07 is the first task that requires the exact global frontier and owner gate green, after new Phase 20 tests and source-row repairs land. The validation map and research distinguish the two snapshots and this sequence.

The new interim command passed in the current worktree. The global pin was intentionally observed red in the current worktree and remains a required Phase 20 completion gate. Independent sign-off must verify the revised dependency and verification sequence; this paragraph does not self-approve that check.

```yaml
issues_pending_recheck: []
```

**Recommendation:** Re-run independent plan sign-off on the revised seven-plan set. Preserve the `blocking-human` D-13-34 checkpoint and exact frontier/owner equality at Plan 07.
