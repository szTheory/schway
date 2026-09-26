---
phase: 14-evidence-instrument-and-honest-scoping
plan: 02
subsystem: parser-diagnostics
tags: [go, parser, diagnostic-identity, distinctness-gate, dx-08]

requires:
  - phase: 14-evidence-instrument-and-honest-scoping
    provides: "plan 14-01's groundedness lint (not a functional dependency, same phase/wave)"
provides:
  - "internal/compiler/syntax/parser.go's recoverRegion + skipped_region cause: three structurally distinct broken programs now mint three distinct diagnostic IDs and three distinct result: IDs instead of colliding on one"
  - "internal/compiler/check/diagnostic_distinctness_test.go — the diagnostic_distinctness gate (DX-08): a lexer-only corpus-distinctness predicate, a live 1.0-distinctness assertion, and a frozen pre-fix non-inertness proof"
  - "testdata/distinctness/ — an 11-member parse-failing corpus plus a frozen collision_control.json capture"
affects: ["any future plan that adds identity-bearing Cause content to a diagnostic", "the phase that eventually lands `if`/conditional surface syntax (the spiral trio becomes a valid-program fixture and must be replaced)"]

actuals:
  tokens: 7172
  tasks: 3
  commits: 4

tech-stack:
  added: []
  patterns:
    - "recoverRegion beside recoverUntil: a token-skipping recovery primitive that additionally returns the discarded diagnostic.Span, keeping the same early-return-at-boundary convention"
    - "Widen a fixed-arity reporting function (p.problem) to a trailing variadic causes parameter so every existing call site compiles unchanged"
    - "Lexer-only, kind-only corpus distinctness predicate (rename-immune by construction) as an independent-of-the-instrument-under-test corpus gate"
    - "Frozen, never-regenerated JSON capture (collision_control.json) as a non-inertness control, read-only, asserted via mtime/size stability"

key-files:
  created:
    - testdata/distinctness/spiral_full.lang
    - testdata/distinctness/spiral_narrow.lang
    - testdata/distinctness/spiral_bare.lang
    - testdata/distinctness/for_range.lang
    - testdata/distinctness/arithmetic.lang
    - testdata/distinctness/unclosed_brace.lang
    - testdata/distinctness/stray_semicolon.lang
    - testdata/distinctness/lone_else.lang
    - testdata/distinctness/empty_file.lang
    - testdata/distinctness/one_token_hard_case.lang
    - testdata/distinctness/dangling_pipe.lang
    - testdata/distinctness/collision_control.json
    - testdata/distinctness/README.md
    - internal/compiler/syntax/parser_skipped_region_test.go
    - internal/compiler/check/diagnostic_distinctness_test.go
    - .planning/phases/14-evidence-instrument-and-honest-scoping/deferred-items.md
  modified:
    - internal/compiler/syntax/parser.go

key-decisions:
  - "recoverRegion advances past the unexpected token BEFORE measuring the discarded extent, so causes[0].span never re-covers the token already reported at primary_span -- this is what makes the third spiral member's skipped region exactly one token wide (matches 14-CONTEXT.md D-14-36's measured shape) rather than including the unexpected token itself."
  - "The skipped_region cause is attached ONLY when recoverRegion actually discarded >= 1 token (discarded=true); when the parser is already at a boundary immediately after the unexpected token, no cause is added and the diagnostic is byte-identical to before the fix -- this is what confines published-ID churn to fixtures whose declaration recovery discarded at least one token, per the plan's own budget statement."
  - "diagnostic_distinctness_test.go lives in package check_test, not package check as check_blame_test.go's literal analog does -- session.CheckCommandFile is needed for end-to-end result: ID computation, and session imports check, so a package-check file importing session would create a check->session->check cycle. check_exclusive_test.go and check_payload_test.go already establish check_test as the sanctioned package for check-package tests that need session."
  - "The published-ID churn enumeration (Task 2) was run as a one-time scripted before/after comparison over the whole testdata/**/*.lang corpus, not committed as a regression test -- documented in this SUMMARY's churn table and flagged human_judgment:true in coverage (no committed test re-asserts the predicted diff set on every future run)."

patterns-established:
  - "Any future task that widens a diagnostic-reporting function's causes should follow p.problem's variadic-trailing-parameter shape rather than adding an overload or changing existing call sites."

requirements-completed: [DX-08]

coverage:
  - id: D1
    description: "The spiral trio (if v { v } else { v } / if v { v } / if v) mints three distinct diagnostic-ID sets and three distinct result: IDs via an identity-bearing skipped_region cause; primary_span stays one token wide and identical across all three; a successfully-parsed program is unaffected"
    requirement: "DX-08"
    verification:
      - kind: unit
        ref: "internal/compiler/syntax/parser_skipped_region_test.go#TestSkippedRegionCauseSeparatesSpiralTrio"
        status: pass
      - kind: unit
        ref: "internal/compiler/syntax/parser_skipped_region_test.go#TestValidProgramUnaffectedBySkippedRegionFix"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/diagnostic_distinctness_test.go#TestDiagnosticDistinctnessIsOne"
        status: pass
    human_judgment: false
  - id: D2
    description: "The diagnostic_distinctness gate ships with a lexer-only, rename-immune corpus-distinctness predicate and proves itself non-inert against a frozen pre-fix collision capture it can only read"
    requirement: "DX-08"
    verification:
      - kind: unit
        ref: "internal/compiler/check/diagnostic_distinctness_test.go#TestDistinctnessCorpusMembersAreStructurallyDistinct"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/diagnostic_distinctness_test.go#TestDiagnosticDistinctnessGuardIsNotInert"
        status: pass
    human_judgment: false
  - id: D3
    description: "Published-ID churn from the parser fix is confined to the new testdata/distinctness/ members plus fixtures whose declaration recovery discarded >= 1 token, enumerated by a before/after diff over every testdata/**/*.lang fixture"
    requirement: "DX-08"
    verification: []
    human_judgment: true
    rationale: "The enumeration was a one-time scripted comparison (Python driving the built lang binary) documented in this SUMMARY's churn table, not a committed Go regression test -- a human (or a future plan) must re-verify the predicted-diff-set property holds if the parser's recovery logic changes again."

duration: 15min (first commit to last commit)
completed: 2026-09-18
status: complete
---

# Phase 14 Plan 02: Diagnostic Distinctness Summary

**A `skipped_region` diagnostic.Cause attached to `syntax.expected_declaration` separates three structurally distinct broken programs that previously collided on one diagnostic ID and one `result:` ID, shipped alongside a lexer-only distinctness gate proven non-inert against a frozen pre-fix capture.**

## Performance

- **Duration:** 15 min (commit-to-commit span)
- **Started:** 2026-09-17T20:59:03-04:00
- **Completed:** 2026-09-17T21:13:51-04:00
- **Tasks:** 3
- **Files created/modified:** 17 (16 created, 1 modified)

## Accomplishments

- `internal/compiler/syntax/parser.go`: `recoverRegion` (beside `recoverUntil`) now records the byte extent discarded by declaration-level error recovery; `p.problem` widened to a trailing variadic `causes ...diagnostic.Cause` parameter (all 15 existing call sites compile unchanged); `parseProgram`'s `default:` arm attaches `diagnostic.Cause{Kind: "skipped_region"}` whenever recovery actually discarded >= 1 token.
- `internal/compiler/diagnostic/diagnostic.go` is byte-unchanged — `Causes` was already in the identity basis (`{Schema, Code, Span, Causes}`), so this alters identity *content*, not the schema. `lang.diagnostic/0` stands.
- Measured on the spiral trio (identical `module spiral` / `export {}` preamble, so `primary_span` lands at the same offset in all three):

  | | skipped region | `causes[0].span` | diagnostic ID | result ID |
  |---|---|---|---|---|
  | `if v { v } else { v }` | 18 bytes | `{29,47}` | `diagnostic:8dfde85de26dc0c32d44f471` | `result:22f6dee9ad4fe3deab8c6d96` |
  | `if v { v }` | 7 bytes | `{29,36}` | `diagnostic:9a3c7270623eb28104900185` | `result:ec552b8e69feea9c94e657e5` |
  | `if v` | 1 byte | `{29,30}` | `diagnostic:9f30034ada3f1c67eada489e` | `result:7be1d32d7e70f3aa863d2444` |

  `primary_span` is `{26,28}` — the "if" token, one token wide — identical across all three.
- `internal/compiler/check/diagnostic_distinctness_test.go` (package `check_test`): three assertions — `TestDistinctnessCorpusMembersAreStructurallyDistinct` (lexer-only, pairwise non-trivia token-kind-sequence inequality over `testdata/distinctness/*.lang`), `TestDiagnosticDistinctnessIsOne` (unique diagnostic-ID sets and unique `result:` IDs both equal corpus size, live), `TestDiagnosticDistinctnessGuardIsNotInert` (reads the frozen `collision_control.json` and asserts it reports 1/3, never the healthy 3/3).
- `testdata/distinctness/`: an 11-member parse-failing corpus (the spiral trio, a for-style iteration form, an arithmetic expression, an unclosed brace, a stray semicolon, a lone `else`, an empty file, a one-token-hard-case that shares `spiral_bare.lang`'s skipped-region byte-width but is structurally distinct, and a dangling-pipe data declaration) plus `collision_control.json` (frozen, from before the fix landed) and `README.md` (predicate, frozen-capture discipline, additive-only membership rule).
- Verified `TestDistinctnessCorpusMembersAreStructurallyDistinct` fails the corpus (not the compiler) on a temporary identifier-renamed near-duplicate of `spiral_bare.lang`; the temp fixture was then removed and `git status --porcelain testdata/distinctness` confirmed clean.

## Task Commits

1. **Task 1: Build the distinctness corpus and freeze the pre-fix collision control** — `de405e6` (test)
2. **Task 2: Attach the discarded recovery extent as a `skipped_region` cause, and enumerate published-ID churn** — `0f3c02e` (test, RED) → `672686b` (feat, GREEN)
3. **Task 3: Ship the distinctness gate with its corpus predicate and frozen non-inertness control** — `365e673` (test)

_Note: Task 2 is `tdd="true"`: `0f3c02e` is the intentional RED (`syntax.expected_declaration` diagnostics carry no `skipped_region` cause yet, confirmed failing on an assertion, not a compile error), `672686b` is GREEN (the parser fix)._

## Files Created/Modified

- `internal/compiler/syntax/parser.go` — `recoverRegion`, variadic `p.problem`, `parseProgram`'s `default:` arm attaches `skipped_region`.
- `internal/compiler/syntax/parser_skipped_region_test.go` — RED/GREEN behavioral proof for the spiral trio.
- `internal/compiler/check/diagnostic_distinctness_test.go` — the shipped distinctness gate (3 tests).
- `testdata/distinctness/*.lang` (11 fixtures), `collision_control.json`, `README.md`.
- `.planning/phases/14-evidence-instrument-and-honest-scoping/deferred-items.md` — one pre-existing, unrelated finding logged (see Issues Encountered).

## Decisions Made

See `key-decisions` in frontmatter.

## Published-ID Churn Table (Task 2's enumeration)

Ran `go test ./cmd/lang --json check` (via the built binary) over every `testdata/**/*.lang` fixture (126 total: 115 pre-existing + 11 new distinctness members) before and after the parser change, diffing `id` fields of all diagnostics and the top-level result. **8 fixtures changed; 0 unexplained rows:**

| File | In predicted set? | Why |
|---|---|---|
| `testdata/distinctness/spiral_full.lang` | yes | new `testdata/distinctness/` member |
| `testdata/distinctness/spiral_narrow.lang` | yes | new `testdata/distinctness/` member |
| `testdata/distinctness/spiral_bare.lang` | yes | new `testdata/distinctness/` member |
| `testdata/distinctness/for_range.lang` | yes | new `testdata/distinctness/` member |
| `testdata/distinctness/arithmetic.lang` | yes | new `testdata/distinctness/` member |
| `testdata/distinctness/lone_else.lang` | yes | new `testdata/distinctness/` member |
| `testdata/distinctness/one_token_hard_case.lang` | yes | new `testdata/distinctness/` member |
| `testdata/phase1/malformed.lang` | yes | pre-existing fixture whose declaration recovery discards a 26-byte region (verified: `causes[0].span={150,176}`) — its FIRST diagnostic is `syntax.unexpected_byte` (pinned in `check_ordering_stability_test.go:294` as `diagnostic:ccb9bcd29f3e8d96fa0368b6`, **unchanged**, confirmed still passing); the moved ID belongs to the LATER `syntax.expected_declaration` diagnostic in its diagnostics list, which no pinned artifact references (D-14-38: its only in-repo mention is `check_blame_test.go:750`'s code-string set-membership check, not an ID) |

Fixtures that hit the declaration-recovery arm but discarded **zero** tokens (`testdata/distinctness/stray_semicolon.lang`) or never reached the arm at all (`unclosed_brace.lang` → `syntax.expected_rbrace`; `dangling_pipe.lang` → `syntax.expected_alternative`; `empty_file.lang` → `syntax.expected_module`) correctly show **no** ID movement, confirming the budget: broken-program diagnostic IDs may move; no valid program's evidence moves, and movement is confined to programs that reach `parseProgram`'s `default:` arm and had >= 1 token discarded.

`go test ./...` (full suite) confirms `check_ordering_stability_test.go`'s pinned baselines are unaffected: all packages other than the one pre-existing, unrelated failure below report `ok`.

## Deviations from Plan

None — plan executed as written. `recoverRegion`'s exact placement of the advance-past-unexpected step was an implementation detail not spelled out byte-for-byte in the plan text, chosen to reproduce the measured wide/narrower/one-token shape in `14-CONTEXT.md` D-14-36.

## Debt rows to register

Per Task 3's instruction, this plan writes no file under `.planning/**` other than this SUMMARY. Plan 14-04 creates `PHASE-14-DEBT.md`; plan 14-07 registers these two rows:

1. **Nothing forces the distinctness corpus to grow as the language grows.** When the conditional form (`if`) enters the language, the spiral trio (`spiral_full.lang`, `spiral_narrow.lang`, `spiral_bare.lang`) becomes a *valid* program and must be replaced in `testdata/distinctness/` by the then-current not-in-language surface, or the corpus predicate would need to be re-derived against a program that no longer refuses to parse. **Trigger:** any milestone that lands `if`/conditional surface syntax (not yet on any milestone's roadmap per `.planning/PROJECT.md`). **Proposed owning phase:** the phase that lands conditional surface syntax, whichever milestone that turns out to be.
2. **Putting more content on `Cause.Span` deepens the already-recorded, half-enforced rule that a coordinate shift must never move a diagnostic's ID.** This is now bounded-widened for broken programs only — a whitespace edit inside a broken program's tail (past the point where declaration recovery starts) can now move that program's diagnostic ID, where before this fix it could not (the swallowed region was never in the identity basis). This is an acceptable, deliberately bounded widening of an already-recorded hole (Phase 13's coordinate-shift discipline), not a new defect, but it should be reconciled explicitly rather than rediscovered. **Trigger:** any future plan that attaches identity-bearing `Cause.Span` content derived from parser recovery on a program that DOES parse successfully (today only refused programs reach this code path). **Proposed owning phase:** whichever future phase next widens `Cause` content on a parse-successful diagnostic path.

## Issues Encountered

**One pre-existing, unrelated `go test ./...` failure, confirmed present before this plan's first commit.** `TestSourceNeverSpawnsUnboundedProcesses` (`internal/compiler/native/native_test.go`) fails on two files, neither touched by this plan:
- `.claude/skills/spike-findings-ai-lang/sources/005-native-ffi-provenance-cleanup/lab/lab.go` (committed at `58103f5`, before this phase started)
- `internal/compiler/session/verification_groundedness_test.go` (added by plan 14-01, `f71aba8`)

Confirmed pre-existing via `git show 7bd3e7d:...` (the commit at this plan's start) plus an isolated `-run TestSourceNeverSpawnsUnboundedProcesses` run against that exact tree state, which reproduces the identical failure. Per the executor's scope-boundary rule, this is logged to `.planning/phases/14-evidence-instrument-and-honest-scoping/deferred-items.md` and NOT fixed by this plan — neither file is in this plan's `files_modified`. This plan's own scoped verify command (`go test ./internal/compiler/syntax/... ./internal/compiler/diagnostic/... ./internal/compiler/check/... -count=1`) is green, and the repo-wide `go test ./...` run introduces **zero new failures** beyond this one pre-existing, unrelated defect.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- DX-08 closed: the spiral trio mints three distinct result IDs and diagnostic-ID sets; the diagnostic schema is unchanged; every moved published ID is enumerated and accounted for.
- Two debt items recorded above for plan 14-04/14-07 to formalize into `PHASE-14-DEBT.md`.
- One pre-existing, unrelated `go test ./...` failure logged to `deferred-items.md` for a future plan to address (fix the two exec.Command call sites, or add a scoped exemption for frozen `.claude/skills/**/sources/**` snapshots).
- No blockers for the next plan in this phase.

---
*Phase: 14-evidence-instrument-and-honest-scoping*
*Completed: 2026-09-18*

## Self-Check: PASSED

- FOUND: testdata/distinctness/ (11 .lang fixtures + collision_control.json + README.md), all git-tracked
- FOUND: internal/compiler/syntax/parser_skipped_region_test.go
- FOUND: internal/compiler/check/diagnostic_distinctness_test.go
- FOUND: commit de405e6 (test, Task 1)
- FOUND: commit 0f3c02e (test, RED, Task 2)
- FOUND: commit 672686b (feat, GREEN, Task 2)
- FOUND: commit 365e673 (test, Task 3)
- `go test ./internal/compiler/syntax/... ./internal/compiler/diagnostic/... ./internal/compiler/check/... -count=1` exits 0
- `go test ./internal/compiler/check/... -run 'TestDistinctnessCorpusMembersAreStructurallyDistinct|TestDiagnosticDistinctnessIsOne|TestDiagnosticDistinctnessGuardIsNotInert' -count=1 -v` prints 3 `--- PASS:` lines
- `git diff --stat -- internal/compiler/diagnostic/diagnostic.go` prints nothing
- `go test ./...` reports exactly one FAIL package (`internal/compiler/native`), confirmed pre-existing and unrelated to this plan's diff
