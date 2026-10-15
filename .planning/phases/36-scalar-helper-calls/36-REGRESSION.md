---
phase: 36
status: passed
verified: 2026-10-15T06:55:23.602521+00:00
source_test_revision: b51043c6566f793893c8c4e143493e7ec3af5bcd
scope: focused-prior-phase-sample
---

# Phase 36 — Prior-phase regression evidence

This is the bounded affected sample from current-milestone Phases33–35, run before goal verification. It is not a full suite, whole-package run, broad cross-program native matrix, hosted receipt or human UAT. The user/project exact-selector policy overrides the generic Go full-suite fallback. Each command uses canonical normalize-test-command and run-with-timeout600, Go timeout600s, -count1, verbose actual output and GOCACHE=/tmp/ai-lang-verification-gocache.

Manifest: [/tmp/schway36-regression-b51043c6566f-_h2nianh/owner-results.json](/tmp/schway36-regression-b51043c6566f-_h2nianh/owner-results.json). Exact argv, normalized commands, exit status, expected compiled-root multiplicities, raw logs and SHA-256 are recorded there. All13 commands exited0, all80 expected root occurrences passed, no FAIL/SKIP.69 unique selected names expand to80 occurrences across owner packages; internal/external check tests legitimately share TestPhase35NumericRoutePolicy, so expected multiplicities come from Go TestGoFiles/XTestGoFiles. No zero-selector success is accepted.

| Owner package | Root occurrences | Result | Log |
|---|---:|---|---|
| ./cmd/schway | 4 | PASS | [1-schway.log](/tmp/schway36-regression-b51043c6566f-_h2nianh/1-schway.log) |
| ./internal/compiler/cgen | 4 | PASS | [2-cgen.log](/tmp/schway36-regression-b51043c6566f-_h2nianh/2-cgen.log) |
| ./internal/compiler/check | 9 | PASS | [3-check.log](/tmp/schway36-regression-b51043c6566f-_h2nianh/3-check.log) |
| ./internal/compiler/core | 2 | PASS | [4-core.log](/tmp/schway36-regression-b51043c6566f-_h2nianh/4-core.log) |
| ./internal/compiler/corevalidate | 6 | PASS | [5-corevalidate.log](/tmp/schway36-regression-b51043c6566f-_h2nianh/5-corevalidate.log) |
| ./internal/compiler/executionpeer | 1 | PASS | [6-executionpeer.log](/tmp/schway36-regression-b51043c6566f-_h2nianh/6-executionpeer.log) |
| ./internal/compiler/interp | 5 | PASS | [7-interp.log](/tmp/schway36-regression-b51043c6566f-_h2nianh/7-interp.log) |
| ./internal/compiler/native | 2 | PASS | [8-native.log](/tmp/schway36-regression-b51043c6566f-_h2nianh/8-native.log) |
| ./internal/compiler/originvalidate | 6 | PASS | [9-originvalidate.log](/tmp/schway36-regression-b51043c6566f-_h2nianh/9-originvalidate.log) |
| ./internal/compiler/pathoracle | 1 | PASS | [10-pathoracle.log](/tmp/schway36-regression-b51043c6566f-_h2nianh/10-pathoracle.log) |
| ./internal/compiler/session | 31 | PASS | [11-session.log](/tmp/schway36-regression-b51043c6566f-_h2nianh/11-session.log) |
| ./internal/compiler/surfacemanifest | 2 | PASS | [12-surfacemanifest.log](/tmp/schway36-regression-b51043c6566f-_h2nianh/12-surfacemanifest.log) |
| ./internal/compiler/syntax | 7 | PASS | [13-syntax.log](/tmp/schway36-regression-b51043c6566f-_h2nianh/13-syntax.log) |

## Corrections reached by this gate

- Three newly tracked helper examples now have exact module-aware public-route inventory rows and real CLI evidence, including U64 input-required exit64 and unchanged unused-Byte wrapper value. Commits ec9be36e/03229023.
- Phase34 baseline consumer composes the exact Phase36 source-hash ledger with13 rejection controls; old frozen bytes stay unchanged.
- Phase35 admission law mutation removes the equivalent shift-only preflight in its temporary overlay and reaches the original named semantic assertion; production checks stay intact.
- Plan08 summary has required truthful coverage metadata. The latter three corrections and reverse/reconfirm controls are recorded in [resolved debug](../../debug/resolved/phase36-regression-consumers.md), commits350340c0/b51043c6.

The initial receipt harness assumed unique test names per package and rejected an otherwise passing check package; its multiplicity expectation was corrected. The successful final capture above reran the whole bounded sample at one revision. Earlier failed attempts remain local diagnostic records and are not current success evidence.
