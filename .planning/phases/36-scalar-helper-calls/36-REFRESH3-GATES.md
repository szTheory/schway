# Phase 36 freshness gates — 2026-10-16

**Source HEAD:** `52cdc9954297e0d002d4fa8ca20fe444ddb32f85`  
**Captured:** `2026-10-16T03:58:54Z` UTC  
**Scope:** Fresh Phase36 owner acceptance roots, bounded current regression union, mechanical prohibitions, and required security/UI/code-review gates. Prior reports and UAT remain intact. This Phase36 gate pass changed only its evidence and review artifacts; it did not edit implementation, fixtures, configuration, shared living documents, historical receipts, or legacy dispositions.

## Exact local test policy

No full project or package suites were run. The fresh Go checks use only exact selectors recorded in the JSON receipts, with `GOCACHE=/tmp/ai-lang-verification-gocache`, `-count=1`, `-timeout=600s`, and JSON event output. Package-specific anchored regexes and raw logs are recorded in the receipts; this avoids embedding regex pipes in a Markdown table. No hosted command or push was run.

## Gate outcomes

| Gate | Result | Evidence |
|---|---|---|
| Phase36 mapped task roots | 86/86 passed; 13 packages; 0 missing/fail/skip; 74.718s | [36-REFRESH3-RUN.json](36-REFRESH3-RUN.json) |
| Bounded regression union | 252/252 root executions passed; 16 packages; 0 missing/fail/skip; 137.091s | [36-REFRESH3-REGRESSION.json](36-REFRESH3-REGRESSION.json) |
| Prohibition mechanics | both violation RED controls exit 1 as expected; clean/default GREEN each 2/2; both canonical producers green | [36-REFRESH3-PROHIBITIONS.json](36-REFRESH3-PROHIBITIONS.json) |
| Security | 24/24 unique authored Level 1 threats closed; 0 open; `block_on: high`; no new boundary in the changed emitter path | [36-SECURITY.md](36-SECURITY.md) |
| UI applicability | N/A, non-blocking; no frontend files/spec | [36-UI-REVIEW.md](36-UI-REVIEW.md) |
| Code review | no new Phase36 finding; WR-02 remains deferred warning | [36-REVIEW.md](36-REVIEW.md), [36-REVIEW-DISPOSITION.md](36-REVIEW-DISPOSITION.md) |

## Gate command evidence

The fresh owner commands are the 14 argv entries in `36-REFRESH3-RUN.json`.
Each records the actual package and exact anchored test names, with the cache,
uncached execution, timeout, and JSON options described above.

The regression commands are the 16 argv entries in `36-REFRESH3-REGRESSION.json`; their names come from `35-REFRESH3-RUN.json` and were executed fresh. The Phase35 inventory has 251 unique package/name selectors and 252 actual roots because `TestPhase35NumericRoutePolicy` runs once in each of the check package's internal and external test packages. The cgen group includes the current Phase33 CR-05 dynamic defect repair root. Phase33 CR-03/04/06 remain separate mandatory M007 obligations; they are not silently closed by these passing roots.

The direct prohibition controls use `GSD_PROHIB_SUBJECT` to select each closed fixture; the absent-subject case unsets that variable. For both bad subjects, exit 1 was expected and the assertion observed `actual: violation`, `expected: green`. Clean and absent-subject controls each exited 0 with two passes. These exact commands ran:

```sh
GSD_PROHIB_SUBJECT=.planning/phases/36-scalar-helper-calls/prohibition-outcome-violation.json node --test .planning/phases/36-scalar-helper-calls/36-prohibition-check.test.cjs
GSD_PROHIB_SUBJECT=.planning/phases/36-scalar-helper-calls/prohibition-causal-violation.json node --test .planning/phases/36-scalar-helper-calls/36-prohibition-check.test.cjs
GSD_PROHIB_SUBJECT=.planning/phases/36-scalar-helper-calls/prohibition-clean.json node --test .planning/phases/36-scalar-helper-calls/36-prohibition-check.test.cjs
env -u GSD_PROHIB_SUBJECT node --test .planning/phases/36-scalar-helper-calls/36-prohibition-check.test.cjs
```

Then these exact canonical commands ran independently for the two request files and both returned exit 0 with structured `status: green`, `flagged: false`, `kind: node-test`, and `failFirstProof: violation-fixture`:

```sh
node /Users/jon/.codex/gsd-core/bin/gsd-tools.cjs check prohibition-enforcement /tmp/schway36-refresh3-prohibitions/outcome-request.json --raw
node /Users/jon/.codex/gsd-core/bin/gsd-tools.cjs check prohibition-enforcement /tmp/schway36-refresh3-prohibitions/causal-request.json --raw
```

Security performed the authored zero-open ASVS-1/high-block shortcut. The phase still has 24 unique planned threat IDs, 31 occurrences across eight plans, and no new trust boundary from the inspected dynamic native defect branch. The idempotent `verification.append-audit` call returned `{"appended":false}`; `36-SECURITY.md` therefore remains byte-preserved. The UI skill gate was verified with:

```sh
node /Users/jon/.codex/gsd-core/bin/gsd-tools.cjs check ui.safety-gate 36 --raw
```

It returned `{"frontend":false,"hasUiFiles":false,"hasUiSpec":false,"block":false}`. No interactive UI menu was needed for this N/A safety-gate path, and no screenshots or numeric UI score were fabricated.

## Review disposition and limits

Current code review inspected the changed scalar `OpDefect` emission path and its new behavioral test, then checked the path against the Phase36 scalar-call contracts. The fresh controls show terminal defect attribution for native dynamic helper calls; the application shell retains its existing user-facing failure. No Phase36 finding was added. WR-02 remains a deferred warning because the comment at `corevalidate.go` still describes a same-type rule that cross-typed U64/Bool calls no longer require. Phase33 CR-03/04/06 remain separately owned audit blockers. `36-REVIEW-DISPOSITION.md`, `36-SECURITY.md`, and `36-UI-REVIEW.md` were not rewritten because their decisions/applicability remain unchanged.

The receipts include current input hashes, post-run source bindings, raw-log hashes, exact per-package selectors, and result counts. These are local Darwin/arm64 observations on the recorded source HEAD; no Linux/hosted result, full-suite result, or subjective readability claim follows.
