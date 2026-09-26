---
phase: 14-evidence-instrument-and-honest-scoping
plan: 05
subsystem: testing
tags: [go, go-parser, ast, self-describing-docs, evd-06, evd-08, qlt-02, budget-manifest, language-maturity]

requires:
  - phase: 14-evidence-instrument-and-honest-scoping
    provides: "14-01's groundedness-lint conventions (phaseArtifactGlob, static go/parser test-index resolution) reused as the AST-walk pattern for guard counting"
provides:
  - "TestLanguageMaturityCountsAreCurrent: independent re-derivation (via go/parser AST walk, never a subprocess) of every count .planning/LANGUAGE-MATURITY.md states about the tree -- guard total, per-file count, per-package count, including-tests total, per-package breakdown, corpus program count, corpus line count"
  - "A corrected .planning/LANGUAGE-MATURITY.md: guard total 32->22 (8 files in 2 packages, not 6 in 3; 48 including tests, not 55), the reduce-package row removed (Phase 11 already widened reduce.Reduce to accept multi-function seeds, so its remaining checks are shortcuts, not refusals), and a fresh corpus figure (126 programs, 4,244 lines, up from 89/3,096 at 2026-09-11)"
  - "An observed suite_wall_clock_ns row in qlt02_budget_manifest.json: 191890000000 ns, from one clean cold `go test ./...` run taken during this task"
  - "TestSelfDescribingDocsGuardIsNotInert: proves both self-checks are not inert via three seeded faults (changed stated count, rewritten re-verify command, manifest row removed) plus unmodified-copy controls"
affects: [evidence-instrument-and-honest-scoping, self-describing-docs, budget-manifest]

actuals:
  tokens: 7178
  tasks: 3
  commits: 5

tech-stack:
  added: []
  patterns:
    - "AST-based (go/parser + go/ast), comment-immune structural matching as the independent re-derivation mechanism for a document's own stated counts, replacing a document's self-supplied shell re-verify command as authority"
    - "Checker functions factored to take a document/manifest PATH argument (not a fixed embedded path) so the same function serves both the real-corpus test and a non-inertness test running against t.TempDir() copies"

key-files:
  created:
    - internal/compiler/session/self_describing_docs_test.go
  modified:
    - .planning/LANGUAGE-MATURITY.md
    - internal/compiler/session/qlt02_budget_manifest.json
    - internal/compiler/session/session_phase6_budget.go

key-decisions:
  - "Guard re-derivation predicate is a go/ast BinaryExpr match on `len(X.Functions) != 1` (mirroring the document's own concept exactly), walked via go/parser rather than shelled out to the document's awk line. This is comment-immune by construction: ast.Inspect never visits *ast.Comment nodes, so prose that merely mentions the pattern (found 3 times in session.go, once in session_phase7.go) is never miscounted as a guard, unlike the document's own awk command which counted all 4 as guards."
  - "Corrected the reduce-package guard-table row to zero/removed rather than leaving a stale non-zero figure: Phase 11 (D-11-29/D-11-30) already widened reduce.Reduce to accept multi-function seeds, so its remaining len(Functions) comparisons are <=1/==1 single-function-path shortcuts, not != 1 refusals, and structurally fall outside the guard table's own predicate. Corrected the document's adjacent prose claim ('reduce.Reduce refuses a multi-function seed outright') for the same reason -- leaving it uncorrected while fixing only the count would have been dishonest."
  - "Did not trust CONTEXT.md's own parenthetical note ('It says 32 guards; there are 26') as a target to hit. Per the plan's explicit instruction, ran the independent derivation and wrote whatever it returned (22), which differs from both the document's stale 32 and the planning note's 26 -- the planning note is itself an informal estimate, not a second authority to reconcile against."
  - "Widened QLT02MetricVocabulary() to admit 'suite_wall_clock_ns' (deviation, see below) -- required because TestBudgetAuditRefusesUndeclaredMachine already runs AuditQLT02BudgetManifest against the real checked-in manifest and asserts zero failures; a new metric name outside the closed vocabulary would have failed that pre-existing test."
  - "Split each TDD task's single logical change into a genuine RED (test file referencing not-yet-existing symbols, or run against the not-yet-corrected document) then GREEN commit, rather than writing the finished code and reverse-engineering a red state after the fact -- both RED states were captured by actually running the test/build before committing, not asserted from memory."

patterns-established:
  - "Checker-function path-argument factoring (see key-decisions) is the reusable non-inertness pattern for any future self-describing-document guard in this codebase."

requirements-completed: [EVD-06, EVD-08]

coverage:
  - id: D1
    description: "TestLanguageMaturityCountsAreCurrent independently re-derives every count LANGUAGE-MATURITY.md states about the tree and fails on any drift"
    requirement: "EVD-06"
    verification:
      - kind: unit
        ref: "internal/compiler/session/self_describing_docs_test.go#TestLanguageMaturityCountsAreCurrent"
        status: pass
    human_judgment: false
  - id: D2
    description: "LANGUAGE-MATURITY.md's stated counts corrected to the fresh 2026-09-17 measurement (guard total, per-package breakdown, corpus program/line counts)"
    requirement: "EVD-06"
    verification:
      - kind: unit
        ref: "internal/compiler/session/self_describing_docs_test.go#TestLanguageMaturityCountsAreCurrent"
        status: pass
    human_judgment: false
  - id: D3
    description: "Both self-checks proven non-inert: seeded count change fails naming line/values; rewritten embedded re-verify command does not change the verdict (independence proof); manifest row removal fails the manifest check"
    requirement: "EVD-06"
    verification:
      - kind: unit
        ref: "internal/compiler/session/self_describing_docs_test.go#TestSelfDescribingDocsGuardIsNotInert"
        status: pass
    human_judgment: false
  - id: D4
    description: "qlt02_budget_manifest.json carries an observed suite-wall-clock baseline from a cold `go test ./...` measurement taken during this task, not copied from any document"
    requirement: "EVD-08"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestBudgetManifestLoads"
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestQLT02BudgetManifestFileUnchangedDuringAudit"
      - kind: other
        ref: "python3 -c \"import json;rows=json.load(open('internal/compiler/session/qlt02_budget_manifest.json'));print(len(rows))\" -> 6 (5 pre-existing + 1 new)"
        status: pass
    human_judgment: false

duration: 55min
completed: 2026-09-18
status: complete
---

# Phase 14 Plan 05: Machine-Check the Self-Describing Docs Summary

**Independent go/parser AST re-derivation makes `.planning/LANGUAGE-MATURITY.md`'s counts machine-checked (32→22 guards corrected) and records the first real cold `go test ./...` wall-clock baseline (191.89s) in `qlt02_budget_manifest.json`.**

## Performance

- **Duration:** 55 min
- **Started:** 2026-09-18T01:20:00Z (approx.)
- **Completed:** 2026-09-18T02:15:00Z (approx.)
- **Tasks:** 3
- **Files modified:** 4 (1 created, 3 modified)

## Accomplishments

- `internal/compiler/session/self_describing_docs_test.go` (new): `TestLanguageMaturityCountsAreCurrent` walks `internal/` and `cmd/` with `go/parser`/`go/ast`, matching the exact structural shape `len(X.Functions) != 1` (never shelling out, never regex-matching the document's own embedded `awk` command), and asserts every count `.planning/LANGUAGE-MATURITY.md` states about the tree against that independent derivation. It also re-derives the `.lang` corpus program count and total line count.
- `.planning/LANGUAGE-MATURITY.md` corrected to the fresh measurement: guard total **32→22**, across **8 files in 2 packages** (not 6 in 3), **48 including tests** (not 55); per-package breakdown `session` 26→20, `cgen` 4→2; the `reduce` row removed entirely (see Deviations/Decisions — Phase 11 already widened `reduce.Reduce`); corpus figure refreshed to **126 programs, 4,244 lines** (up from 89/3,096 at the 2026-09-11 re-assessment).
- `internal/compiler/session/qlt02_budget_manifest.json` gained an `observed` `suite_wall_clock_ns` row: **191,890,000,000 ns (191.89s)**, from a real clean cold `go test ./...` run taken during this task on `machine:4797d76b7863` — not copied from RESEARCH.md's ~192.7s partial re-measurement, CONTEXT.md, or any archived `~60s`/`~120s` VALIDATION.md prose.
- `TestSelfDescribingDocsGuardIsNotInert` proves both self-checks go red on a seeded staleness and green otherwise, across 5 subtests: unmodified doc copy passes; a stated-count edit fails naming the document path and both values; a rewritten embedded re-verify command does **not** change the verdict (the independence proof — the checker never executes that command); unmodified manifest copy passes; a manifest copy with the wall-clock row removed fails.

## Task Commits

Each task was committed as a genuine RED-then-GREEN pair (test file referencing not-yet-existing symbols or run against the not-yet-corrected document, confirmed failing/not-building by actually running it, then the corresponding fix):

1. **Task 1: Machine-check LANGUAGE-MATURITY.md's counts** —
   - `305e05a` `test(14-05): add failing test for LANGUAGE-MATURITY.md count self-check` (RED — confirmed failing against the stale document, 8 findings logged, see below)
   - `c186ef8` `feat(14-05): correct LANGUAGE-MATURITY.md's counts to the re-derived values` (GREEN)
2. **Task 2: Measure the suite wall-clock cold and record it** —
   - `f33e9af` `feat(14-05): record a cold full-suite wall-clock baseline (EVD-08)` (single commit, `type="auto"`, no TDD — includes the required `QLT02MetricVocabulary()` widening, see Deviations)
3. **Task 3: Prove both self-checks are not inert** —
   - `f43a89e` `test(14-05): add failing test for the self-describing docs guards' non-inertness` (RED — package did not build, referencing `checkSuiteWallClockObservedRow`/`suiteWallClockMetric` before they existed)
   - `7c710ee` `feat(14-05): implement the suite wall-clock manifest-row check` (GREEN)

**Plan metadata:** committed as part of this same close-out (see Final Commit below).

## Files Created/Modified

- `internal/compiler/session/self_describing_docs_test.go` (new, 502 lines) — `TestLanguageMaturityCountsAreCurrent`, `TestSelfDescribingDocsGuardIsNotInert`, and their shared helpers (`scanFunctionsGuards`, `corpusStats`, `checkLanguageMaturityDoc`, `checkSuiteWallClockObservedRow`).
- `.planning/LANGUAGE-MATURITY.md` — guard counts, per-package table, corpus figures corrected; new "re-assessed 2026-09-17" entries; explicit notes marking the embedded `awk`/shell commands as approximate-only, no longer authoritative.
- `internal/compiler/session/qlt02_budget_manifest.json` — new `suite_wall_clock_ns` observed row (6th row).
- `internal/compiler/session/session_phase6_budget.go` — `QLT02MetricVocabulary()` widened by one entry (deviation, see below).

## Before/After Numbers (Task 1, per acceptance criteria)

| Figure | Before (stated) | After (re-derived, now stated) |
|---|---|---|
| Non-test guard total | 32 | 22 |
| Files carrying guards | 6 | 8 |
| Packages carrying guards | 3 (session, cgen, reduce) | 2 (session, cgen) |
| Guards including tests | 55 | 48 |
| `session` package guards | 26 | 20 |
| `cgen` package guards | 4 | 2 |
| `reduce` package guards | 2 | 0 (row removed) |
| Corpus programs | 89 (2026-09-11) | 126 (2026-09-17) |
| Corpus lines | 3,096 (2026-09-11) | 4,244 (2026-09-17) |

RED-phase failure output (Task 1, against the original stale document, `go test ./internal/compiler/session/... -run TestLanguageMaturityCountsAreCurrent -count=1 -v`):
```
self_describing_docs_test.go:317: .../LANGUAGE-MATURITY.md:96: stated non-test guard total 32 does not match re-derived 22
self_describing_docs_test.go:317: .../LANGUAGE-MATURITY.md:96: stated guard file count 6 does not match re-derived 8
self_describing_docs_test.go:317: .../LANGUAGE-MATURITY.md:96: stated guard package count 3 does not match re-derived 2
self_describing_docs_test.go:317: .../LANGUAGE-MATURITY.md:96: stated including-tests guard total 55 does not match re-derived 48
self_describing_docs_test.go:317: .../LANGUAGE-MATURITY.md:101: stated guard count for package "session" is 26, does not match re-derived 20
self_describing_docs_test.go:317: .../LANGUAGE-MATURITY.md:102: stated guard count for package "cgen" is 4, does not match re-derived 2
self_describing_docs_test.go:317: .../LANGUAGE-MATURITY.md:103: stated guard count for package "reduce" is 2, does not match re-derived 0
self_describing_docs_test.go:317: .../LANGUAGE-MATURITY.md: could not locate the corpus-count sentence in the expected shape
--- FAIL: TestLanguageMaturityCountsAreCurrent (0.09s)
```

## Task 2 Measurement Record (per acceptance criteria)

- **Command:** `go clean -testcache && time go test ./... -count=1` (run in full on this session's host; `time` wrapper via Bash `{ time ...; }`)
- **Raw wall clock:** `3:11.89 total` (real) = **191.89s** = **191,890,000,000 ns**; user 102.63s, system 67.95s, 88% CPU
- **Host identifier:** `machine:4797d76b7863` (via `measure.ProbeMachine`/`measure.MachineID`, confirmed live on this host by a throwaway test, matching all 5 pre-existing manifest rows)
- **Go version:** `go1.24.0 darwin/arm64` (Darwin 25.6.0, arm64)
- **No figure was copied from any document.** RESEARCH.md's ~192.7s and CONTEXT.md's citation of it are both explicitly a partial/order-of-magnitude re-measurement (RESEARCH.md A2 says so directly: "I did not do a single clean full-suite cold timing run in this research pass"); this task's 191.89s is the first actual clean cold measurement recorded as a manifest artifact, and it happens to be close to that estimate — coincidence of a correct estimate, not evidence the estimate was trusted.
- The full run's only failure was the pre-existing, out-of-scope `internal/compiler/native` package failure (`TestSourceNeverSpawnsUnboundedProcesses` against `.claude/skills/.../lab.go` and `internal/compiler/session/verification_groundedness_test.go`), exactly as flagged in this plan's prior-work context.

### Stale `~60s`/`~120s` wall-clock figures still present under `.planning/` (enumerated, not fixed — per Task 2's explicit scope boundary)

Per D-14-12 ("archived `*-VALIDATION.md` are in scope for detection and append-only for repair — a dead row is never rewritten in place"), and because these are narrative *estimated-runtime* prose, not `go test -run` verification command spans (the EVD-01 groundedness lint landed in 14-01/14-06 asserts only over command code-spans, never over prose runtime estimates), none of these were edited by this plan. Every file below still states a stale short figure:

| File | Lines | Stale figure | Owning phase/plan |
|---|---|---|---|
| `.planning/milestones/M002-phases/08-interprocedural-loan-liveness-in-check/08-VALIDATION.md` | 27, 39, 94 | "~60 seconds (full suite)" / "60 seconds" / "< 60s" | None of Phase 14's existing plans scope narrative runtime prose (the groundedness lint only asserts over verification-command code-spans). Recommend a new debt row landing at **QLT-10** (the reconciliation phase named by D-14-16), or a future Phase-14 plan if one is added for prose-only staleness. |
| `.planning/milestones/M002-phases/10-trusted-interprocedural-oracle/10-VALIDATION.md` | 28, 37 | "~120 seconds" / "120 seconds" | Same — QLT-10. |
| `.planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md` | 28, 41, 132 | "~60 seconds" / "60 seconds" / "< 60s" | Same — QLT-10. |
| `.planning/milestones/M002-phases/12-result-payloads/12-VALIDATION.md` | 26, 35, 92 | "~60 seconds" / "60 seconds" / "< 60s" | Same — QLT-10. |
| `.planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/13-VALIDATION.md` | 26, 46, 185 | "~60-120 seconds" / "~120 seconds" / "~60-120s" | Same — QLT-10. |

`.planning/REQUIREMENTS.md:63` and the Phase 14 planning documents (`CONTEXT.md`, `RESEARCH.md`, `ROADMAP.md`, `DISCUSSION-LOG.md`, `14-VALIDATION.md`) also mention the ~60s figure, but only as meta-commentary describing the staleness (e.g. "documents claim ~60s") — those are not themselves stale claims and were left as-is.

## Decisions Made

See frontmatter `key-decisions` for the full list. The most consequential: the guard-count predicate is a structural (`go/ast`) match, not a text match, which is what makes it comment-immune — this is *why* the independently re-derived total (22) differs from both the document's stale 32 and the planning-note's informal 26 (`CONTEXT.md`'s "Claude's Discretion" aside), and the plan's own instruction not to trust either planning-stage figure was followed literally.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Widened `QLT02MetricVocabulary()` to admit the new metric name**
- **Found during:** Task 2 (adding the `suite_wall_clock_ns` manifest row)
- **Issue:** `AuditQLT02BudgetManifest` enforces a closed metric vocabulary (`QLT02MetricVocabulary()`), and `TestBudgetAuditRefusesUndeclaredMachine` (pre-existing, in `session_phase6_budget_test.go`) already runs that audit against the real checked-in manifest and asserts **zero** failures against the live machine ID. Adding a manifest row naming a metric outside the closed vocabulary would have failed that pre-existing test — a genuine blocker to completing Task 2's acceptance criteria ("the existing manifest-loading test passes unchanged").
- **Fix:** Added `"suite_wall_clock_ns"` to `QLT02MetricVocabulary()` in `internal/compiler/session/session_phase6_budget.go`, with a doc comment stating it is gate_type `"observed"` only (never `"hard"` — `QLT02GateEligibleMetrics()` is unchanged, so this metric can never gate a build).
- **Files modified:** `internal/compiler/session/session_phase6_budget.go` (not in the plan's declared `files_modified`, but directly required by Task 2's own change — see Rule 3's scope boundary: "directly caused by the current task's changes").
- **Verification:** `go test ./internal/compiler/session/... -run 'TestBudgetAuditRefusesUndeclaredMachine|TestEvaluateBudgetAgreesWithDemote|TestGrowthExponentIsAlsoAHardGate|TestGateEligibleMetricSetsAgreeAcrossChokepoints' -count=1 -v` — all pass; `TestEvaluateBudgetAgreesWithDemote` in particular iterates every non-gate-eligible metric in the vocabulary (including the new one) and confirms it never reports a blocking verdict, so the widening did not silently create a new hard gate.
- **Committed in:** `f33e9af` (Task 2's commit)

**2. [Rule 2 - Missing Critical] Corrected an adjacent stale factual claim about `reduce.Reduce` while correcting the guard count**
- **Found during:** Task 1 (re-deriving the per-package breakdown)
- **Issue:** The document's guard-table "consequences" prose claimed `reduce.Reduce` "refuses a multi-function seed outright." The independent re-derivation found zero `!= 1` guards in the `reduce` package (the pattern the table's own predicate counts), and inspecting `internal/compiler/reduce/reduce.go`'s own doc comment confirmed why: Phase 11 (D-11-29/D-11-30) already widened `Reduce` to accept multi-function seeds. Leaving that prose claim uncorrected while fixing only the numeric count directly beside it would have been the exact kind of honesty gap this plan exists to close.
- **Fix:** Removed the `reduce` row from the guard breakdown table and rewrote the "two consequences" bullet to state the correction, citing D-11-29/D-11-30, and noting the claim was accurate as of the 2026-09-11 re-assessment.
- **Files modified:** `.planning/LANGUAGE-MATURITY.md`
- **Verification:** Manual inspection of `internal/compiler/reduce/reduce.go`'s package doc comment and `seed_validate.go`, confirming the widened `Seed`-accepting behavior; `TestLanguageMaturityCountsAreCurrent` passes with the `reduce` row absent (its independent scan finds zero `reduce`-package matches for the guard predicate, so a present-but-zero row would have needed special-casing — removing it is the cleaner, honest correction).
- **Committed in:** `c186ef8` (Task 1's GREEN commit)

---

**Total deviations:** 2 auto-fixed (1 blocking, 1 missing-critical-honesty). **Impact on plan:** Both were necessary corrections directly adjacent to the plan's own declared scope (a manifest vocabulary that had to admit the new row's metric name; a document claim that would have remained false right next to the very count being corrected). No scope creep beyond what each task's own change required.

## Issues Encountered

None beyond the deviations above.

## Known Stubs

None.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- Both self-describing artefacts (`LANGUAGE-MATURITY.md`, `qlt02_budget_manifest.json`) are now machine-checked and currently true; `go test ./...` will fail the suite the next time either drifts from the tree.
- The enumerated `~60s`/`~120s` stale prose in five archived `*-VALIDATION.md` files (table above) remains unresolved and unowned by any existing Phase 14 plan — recommend either a new debt row at the phase's mid-phase/close-out gate, or explicit scoping into a future QLT-10 plan, so it is not silently forgotten.
- `internal/compiler/native`'s `TestSourceNeverSpawnsUnboundedProcesses` failure remains, exactly as flagged in this plan's prior-work context, unresolved by design (owned by the phase's post-merge gate, not this plan).

## Self-Check: PASSED

- `[ -f internal/compiler/session/self_describing_docs_test.go ]` → FOUND
- `[ -f .planning/LANGUAGE-MATURITY.md ]` → FOUND (modified)
- `[ -f internal/compiler/session/qlt02_budget_manifest.json ]` → FOUND (modified)
- Commits present in `git log --oneline --all`: `305e05a`, `c186ef8`, `f33e9af`, `f43a89e`, `7c710ee` — all FOUND
- All four plan-level `<verify>` command sets re-ran clean in this final pass (see "Task Commits" / body above)
- `go test ./...` re-run: only the pre-existing, out-of-scope `internal/compiler/native` failure; every other package including `internal/compiler/session` passed

---
*Phase: 14-evidence-instrument-and-honest-scoping*
*Completed: 2026-09-18*
