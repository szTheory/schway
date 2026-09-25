# Phase 20 Plan Check

**Phase:** 20 — Nyquist, D-13-34, and the Frontier Fixture  
**Plans checked:** 7  
**Result:** PASS — 0 blockers, 0 warnings, 0 advisories

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

The 18 `20-VALIDATION.md` command cells now use exact test names, named prefixes, or separate commands joined by `&&`; none needs a Markdown escape converted into shell syntax. A read-only parser check confirmed that every table row has the expected cell count, each command cell has one intact code span, contains no pipe or backslash, and splits into shell-parseable `go test` or `go run` invocations. The two relevant existing grade-cap tests pass together. All seven plans still pass `frontmatter.validate --schema plan` and `verify.plan-structure` with no errors or warnings.

The existing Phase 14 groundedness gate was also run. It is red before Phase 20 execution because the newly added research and validation documents cite planned tests that do not exist yet, and because those documents move the measured corpus frontier. This is execution work assigned to Plans 01–03, 05, and 07; it is not evidence that the command-map cells are malformed. The gate reported R2=10 and R2b=25 post-reconciliation at this planning snapshot. Those counts must be remeasured and fully reconciled during the phase; the `20-VALIDATION.md` rows remain `pending` and `status: planned` until then.

```yaml
issues: []
```

**Recommendation:** Proceed with the seven-plan execution graph. Keep the D-13-34 human checkpoint, and require the planned live groundedness and debt-cap gates before phase close.
