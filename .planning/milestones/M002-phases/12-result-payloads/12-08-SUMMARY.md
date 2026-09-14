---
phase: 12-result-payloads
plan: 08
subsystem: process
tags: [debt-register, checkpoint, cr-01, d-12-43, phase-close]

requires:
  - phase: 12-result-payloads
    provides: "12-06 (CR-01 source-layer closure, WR-01), 12-07 (CR-01 engine-layer closure, WR-02, IN-01), 12-VERIFICATION.md's two routed human-verification items"
provides:
  - "D-12-43 ratified as criterion 2's terminal finding at an explicit blocking-human checkpoint (closing 12-VERIFICATION.md's human-verification item 2)"
  - "D-12-44: CR-01's disposition recorded as a decision (FIX), with evidence, restriction, and GEN-01 lifting condition"
  - "D-12-45: WR-01/WR-02/IN-01 dispositions recorded, none deferred"
  - "A like-for-like phase-close evidence table for re-verification"
affects: ["Phase 12 re-verification", "M003 (GEN-01's real generic Result<T, E>, the named D-12-44 lifting condition)"]

actuals:
  tokens: 9200
  tasks: 3
  commits: 1

tech-stack:
  added: []
  patterns:
    - "Checkpoint outcome recorded as a dated, verbatim-quoted paragraph appended to an existing debt-register detail section, per D-12-36's own precedent for recording a checkpoint outcome without rewriting prior text"

key-files:
  modified:
    - .planning/phases/12-result-payloads/PHASE-12-DEBT.md

key-decisions:
  - "D-12-43 RATIFIED (not slipped): the developer selected option id `ratify` at the blocking-human checkpoint, accepting D-12-43 as criterion 2's terminal state as worded. No enabling work is scheduled; no future phase is named as owner; ROADMAP.md gains no new phase or scope from this decision."
  - "D-12-44 recorded: CR-01's disposition is FIX (not ratified debt), closed by plan 06's check.duplicate_payload_type declaration-time refusal plus plan 07's shared core.AlternativeNameForPayloadType/core.LookupAlternativeDetail resolver. Restriction: two alternatives of one data type may not declare the same payload type. Lifting condition: GEN-01's real generic Result<T, E> (M003) at T == E."
  - "D-12-45 recorded: WR-01 (plan 06 Task 3), WR-02 (plan 07 Task 3), and IN-01 (plan 07 Task 1) are each closed; none deferred."

requirements-completed: [RES-02, RES-03]

coverage:
  - id: D1
    description: "D-12-43's status as criterion 2's terminal finding was decided at an explicit blocking-human checkpoint (Task 1), not ratified silently inside a task."
    requirement: "RES-03"
    verification:
      - kind: manual_procedural
        ref: "Task 1 checkpoint: developer replied `ratify` verbatim; outcome transcribed into PHASE-12-DEBT.md's D-12-43 detail section with date and verbatim words"
        status: pass
    human_judgment: true
    rationale: "This deliverable IS the human decision itself -- a checkpoint outcome cannot be verified by an automated test; it is the checkpoint's transcript, cross-checked here against the register text it produced."
  - id: D2
    description: "CR-01's disposition is recorded as decision D-12-44 with an ID, chosen branch, stated reason, evidence, restriction, and lifting condition -- not inferred from a passing test suite."
    requirement: "RES-02"
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session/... -run TestDebtRegistersAreWellFormed -v -count=1"
        status: pass
      - kind: other
        ref: "grep -c -E '^| D-12-(44|45) |' PHASE-12-DEBT.md == 2; grep -c -E '^### D-12-(43|44|45)' PHASE-12-DEBT.md == 3"
        status: pass
    human_judgment: false
  - id: D3
    description: "WR-01, WR-02, and IN-01 each have a recorded disposition (D-12-45); none silently dropped."
    requirement: "RES-02"
    verification:
      - kind: other
        ref: "PHASE-12-DEBT.md ### D-12-45 detail section names all three findings with plan/task references and states none was deferred"
        status: pass
    human_judgment: false
  - id: D4
    description: "The debt register stays mechanically well-formed after every edit: items count matches table rows, every row has a detail section, every severity is from the closed vocabulary, every landing phase is non-empty."
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session/... -run TestDebtRegistersAreWellFormed -v -count=1 (11 registers, all PASS, including PHASE-12-DEBT.md)"
        status: pass
    human_judgment: false
  - id: D5
    description: "The whole workspace suite is green at phase close, and every control named in the planning brief's must-not-regress list still passes."
    verification:
      - kind: integration
        ref: "go test ./... (25 package result lines, zero FAIL)"
        status: pass
      - kind: integration
        ref: "ten named must-not-regress controls, each re-run individually -- see table below"
        status: pass
    human_judgment: false

duration: 25min
completed: 2026-09-13
status: complete
---

# Phase 12 Plan 08: D-12-43 Checkpoint Resolution + CR-01/WR/IN Dispositions Summary

**D-12-43 ratified at an explicit blocking-human checkpoint (developer chose `ratify`); CR-01's FIX disposition and the WR-01/WR-02/IN-01 dispositions are now recorded as decisions D-12-44/D-12-45 in `PHASE-12-DEBT.md`, which stays mechanically well-formed; the whole workspace suite and all ten must-not-regress controls are green at phase close.**

## Performance

- **Duration:** ~25 min
- **Tasks:** 3
- **Files modified:** 1 (`PHASE-12-DEBT.md`)

## Accomplishments

- **Task 1 (checkpoint, `gate="blocking-human"`):** Presented D-12-43's status question to the human via the plan's own three-option decision. The human replied with option id `ratify`, selecting "RATIFY — D-12-43 is criterion 2's terminal state as worded." This closes `12-VERIFICATION.md`'s human-verification item 2 — the finding was previously reached by measurement inside plan 05 without a blocking human checkpoint; it now has one.
- **Task 2:** Edited `PHASE-12-DEBT.md` additively:
  - Appended a dated ratification paragraph to the existing `### D-12-43` detail section, recording the checkpoint date (2026-09-13), the chosen option id (`ratify`), and the developer's verbatim words. The original measurement narrative is textually unchanged.
  - Added row and detail section `D-12-44` (severity `blocker`, CLOSED — plans 06-07): CR-01's disposition is FIX, restating `12-06-PLAN.md`'s `<cr01_disposition>` reasoning for why (b) as literally described (an `AlternativeName` IR field) was rejected on corpus-byte-stability grounds and which narrowed form of its intent (plan 07's shared `core` resolver) was adopted instead. Names the shipped diagnostic code, fixture path, ambiguity error text, the restriction ("two alternatives of one data type may not declare the same payload type"), and the GEN-01 lifting condition.
  - Added row and detail section `D-12-45` (severity `info`, CLOSED — plans 06-07): WR-01, WR-02, IN-01 dispositions, one paragraph each, stating none was deferred.
  - Updated frontmatter `items:` from 8 to 10 to match the new table row count.
- **Task 3:** Re-ran the full evidence set with no code changes. Confirmed `grep -rn "duplicate_payload_type" internal/compiler/` now returns matches (previously zero, per `12-VERIFICATION.md`) — the clearest single-line proof CR-01 closed. Confirmed `go test ./...` green across all 25 packages. Confirmed all ten must-not-regress controls PASS. Confirmed five of the six D-12-36 legacy `cgen` emitters are byte-unchanged since before plans 06-07; `emitBranch` is the sole, planned exception.

## Task Commits

1. **Task 1 + Task 2 (combined, no separable code artifact for Task 1 — it is a checkpoint answer transcribed by Task 2's edit):** `1292c2d` (docs) — "docs(12-08): record D-12-43 ratification and CR-01/WR/IN dispositions"
2. **Task 3:** No commit — verification-only, no files modified (per the plan's explicit instruction not to edit `12-VERIFICATION.md`, `REQUIREMENTS.md`, `ROADMAP.md`, or `STATE.md` in this task).

**Plan metadata:** committed separately below.

## Files Created/Modified

- `.planning/phases/12-result-payloads/PHASE-12-DEBT.md` — added rows D-12-44, D-12-45; amended D-12-43's detail section with a dated ratification paragraph; updated frontmatter `items: 8` → `items: 10`; added an "Amended at phase close" note under the intro paragraph. No existing row deleted or reworded (`git diff` shows one removed line: the frontmatter `items: 8` line itself, which is the required count update).

## Checkpoint Outcome (Task 1)

**Decision:** Is D-12-43 accepted as criterion 2's terminal state, or does enabling work get named and slipped to a future phase?

**Developer's answer (verbatim):** `ratify`

**Option selected:** "RATIFY — D-12-43 is criterion 2's terminal state as worded."

**Consequence:** No enabling work is scheduled. No future phase is named as owner of reopening D-12-43. `.planning/ROADMAP.md` gains no new phase and no new scope from this decision. D-12-43's landing phase remains "OPEN and UNOWNED", governed by its existing reopening condition (a future plan changing how a payload-carrying return's terminal value is derived).

## Phase-Close Evidence Table

Re-run verbatim from `12-VERIFICATION.md`'s "Behavioral Spot-Checks" table:

| Behavior | Command | Observed Result | Status |
|----------|---------|------------------|--------|
| N=1 convergence differential | `go test ./internal/compiler/cgen/... -run TestN1ConvergenceDifferential -v -count=1` | 5/5 subtests PASS | ✓ PASS |
| Payload tracer three-engine agreement | `go test ./internal/compiler/cgen/... -run TestPayloadTracer -v -count=1` | PASS | ✓ PASS |
| Payload pattern refusals | `go test ./internal/compiler/check/... -run TestPayloadPatternRefusals -v -count=1` | 5/5 subtests PASS (grew from 4/4 — `duplicate_payload_type` subtest added by plan 06) | ✓ PASS |
| Resource payload refusal | `go test ./internal/compiler/check/... -run TestResourcePayloadRefused -v -count=1` | 2/2 subtests PASS | ✓ PASS |
| Exhaustive dispatch control | `go test ./internal/compiler/core/... -run TestAllOperationKindsHandledAtEverySite -v -count=1` | PASS | ✓ PASS |
| Corpus characterization replay | `go test ./internal/compiler/session/... -run TestPayloadCorpusCharacterizationReplay -v -count=1` | All non-skipped fixtures PASS (97 subtests: 62 PASS, 35 SKIP), byte-identical | ✓ PASS |
| Debt registers well-formed | `go test ./internal/compiler/session/... -run TestDebtRegistersAreWellFormed -v -count=1` | PASS (11 registers, including amended PHASE-12-DEBT.md) | ✓ PASS |
| Layout mutation control (D-12-37) | `go test ./internal/compiler/session/... -run TestPayloadLayoutMutationRefused -v -count=1` | PASS | ✓ PASS |
| Slot-swap mutation control (D-12-38) | `go test ./internal/compiler/session/... -run TestPayloadSlotSwapMutationKilled -v -count=1` | PASS (measures/pins absence, per ratified D-12-43; WR-02's injected-write assertion also passes) | ✓ PASS (honesty-preserving regression test, not a "bug caught" claim) |
| Full workspace suite (run once) | `go test ./...` | All 25 package result lines `ok`/no-test-files, zero FAIL | ✓ PASS |
| Duplicate-payload-type ambiguity check exists | `grep -rn "duplicate_payload_type" internal/compiler/` | **BEFORE (per 12-VERIFICATION.md): no match.** **AFTER (this run): 8 matches** — `core.go:101` (comment), `check.go:166` (the refusal), `check_payload_test.go` (6 matches: fixture name, expected code x3, comment, zero-count assertion) | ✓ PASS — confirms CR-01 resolved |

## Must-Not-Regress Controls (all ten, individually re-run)

| Control | Command | Result |
|---|---|---|
| `TestPayloadTracerThreeEngineAgreement` | `go test ./internal/compiler/cgen/... -run TestPayloadTracerThreeEngineAgreement -v -count=1` | PASS |
| `TestPayloadPatternRefusals` | `go test ./internal/compiler/check/... -run TestPayloadPatternRefusals -v -count=1` | PASS (5/5 subtests) |
| `TestResourcePayloadRefused` | `go test ./internal/compiler/check/... -run TestResourcePayloadRefused -v -count=1` | PASS (2/2 subtests) |
| `TestAllOperationKindsHandledAtEverySite` | `go test ./internal/compiler/core/... -run TestAllOperationKindsHandledAtEverySite -v -count=1` | PASS |
| `TestPayloadCorpusCharacterizationReplay` | `go test ./internal/compiler/session/... -run TestPayloadCorpusCharacterizationReplay -v -count=1` | PASS (97 subtests: 62 PASS, 35 SKIP) |
| `TestPayloadLayoutMutationRefused` | `go test ./internal/compiler/session/... -run TestPayloadLayoutMutationRefused -v -count=1` | PASS |
| `TestPayloadSlotSwapMutationKilled` | `go test ./internal/compiler/session/... -run TestPayloadSlotSwapMutationKilled -v -count=1` | PASS |
| `TestN1ConvergenceDifferential` | `go test ./internal/compiler/cgen/... -run TestN1ConvergenceDifferential -v -count=1` | PASS (5/5 subtests) |
| `TestDebtRegistersAreWellFormed` | `go test ./internal/compiler/session/... -run TestDebtRegistersAreWellFormed -v -count=1` | PASS (11 registers) |
| `TestC03ResultPayloadOriginAcrossOpCall` | `go test ./internal/compiler/corevalidate/... -run TestC03ResultPayloadOriginAcrossOpCall -v -count=1` | PASS (2/2 subtests) |

All ten PASS. `go test ./...` output (25 package result lines):

```
ok      github.com/codename-lang/lang/cmd/lang
ok      github.com/codename-lang/lang/cmd/lang-repair
ok      github.com/codename-lang/lang/internal/compiler/ability
?       github.com/codename-lang/lang/internal/compiler/ast   [no test files]
ok      github.com/codename-lang/lang/internal/compiler/cache
ok      github.com/codename-lang/lang/internal/compiler/callgraph
ok      github.com/codename-lang/lang/internal/compiler/cgen
ok      github.com/codename-lang/lang/internal/compiler/check
ok      github.com/codename-lang/lang/internal/compiler/core
ok      github.com/codename-lang/lang/internal/compiler/corevalidate
ok      github.com/codename-lang/lang/internal/compiler/debugmap
ok      github.com/codename-lang/lang/internal/compiler/diagnostic
ok      github.com/codename-lang/lang/internal/compiler/evidence
ok      github.com/codename-lang/lang/internal/compiler/execution
ok      github.com/codename-lang/lang/internal/compiler/interp
?       github.com/codename-lang/lang/internal/compiler/interp/interptestdirect   [no test files]
ok      github.com/codename-lang/lang/internal/compiler/measure
ok      github.com/codename-lang/lang/internal/compiler/native
ok      github.com/codename-lang/lang/internal/compiler/originvalidate
ok      github.com/codename-lang/lang/internal/compiler/pathoracle
ok      github.com/codename-lang/lang/internal/compiler/protocol
ok      github.com/codename-lang/lang/internal/compiler/reduce
ok      github.com/codename-lang/lang/internal/compiler/session
ok      github.com/codename-lang/lang/internal/compiler/syntax
ok      github.com/codename-lang/lang/internal/compiler/testsupport
```

25 lines total (23 `ok`, 2 `[no test files]`), zero `FAIL`.

## D-12-36 Legacy Emitter Byte-Stability Check

Diffed `internal/compiler/cgen/cgen.go` across the full CR-01 closure range (`e8b82b5~1` — before plan 07's first commit — through `e07613b`, plan 07's last code commit; plan 06's commits are included transitively since `e8b82b5~1` postdates them):

```
git diff --stat e8b82b5~1 e07613b -- internal/compiler/cgen/cgen.go
 internal/compiler/cgen/cgen.go | 75 +++++++++++++++++++++++++++---------------
 1 file changed, 48 insertions(+), 27 deletions(-)
```

Function-signature-level diff (`git diff ... | grep -E '^\+.*func |^-.*func '`):

```
+func PayloadSlotSwapInjectedWriteCount() int {
-func alternativeNameForPayloadType(dataType core.DataType, payloadType string) string {
```

Hunk headers touched: the top-of-file `payloadSlotSwapForTest` region (new counter, WR-02), `emitBranch` (line ~1874 — IN-01's fix), the deletion of `alternativeNameForPayloadType` and `payloadFieldNames` region, and `emitBranchOperations` (CR-01's engine-layer resolver rewiring, twice).

**Five of the six D-12-36 legacy emitters are byte-unchanged**: `emitLinear`, `emitLinearBorrowedByPointer`, `emitLinearBorrowedByPointerPlain`, `emitLinearForeign`, `emitMatch` — none of their function names appear anywhere in the diff. **`emitBranch` is the single deliberate, planned exception**: plan 07 Task 1 replaced its per-alternative `detail` lookup loop with a call to the shared `core.LookupAlternativeDetail` helper (closing IN-01), and `emitBranchOperations` (called from `emitBranch`) was rewired to the shared `core.AlternativeNameForPayloadType` resolver (closing CR-01's engine-layer half). This is exactly the single named exception the plan's Task 3 action text calls for.

`internal/compiler/interp/interp.go`, `internal/compiler/check/check.go`, and `internal/compiler/core/core.go` also changed in this range — expected, since plan 07's shared resolver and plan 06's refusal live there:

```
+func LookupAlternativeDetail(dataType DataType, name string) AlternativeDetail {
+func AlternativeNameForPayloadType(dataType DataType, payloadType string) (string, error) {
-func alternativeNameForPayloadType(program core.Program, parameterTypeName, payloadType string) string {
```

`git diff --stat HEAD~3 -- internal/compiler/cgen/cgen.go internal/compiler/interp/interp.go internal/compiler/check/check.go internal/compiler/core/core.go` (the plan's literal verify range) returned empty output — the three commits immediately preceding this plan's own commit are all `docs(12-07)` state-sync commits that touch none of these four files, so this narrower range shows no diff, which is expected and not a failure signal. The wider, more informative range used above (`e8b82b5~1`..`e07613b`) spans the actual CR-01/WR-02/IN-01 code changes and is the one that demonstrates the emitter byte-stability claim.

## Decisions Made

- **D-12-43 RATIFIED** at plan 08 Task 1's blocking-human checkpoint. See "Checkpoint Outcome" above.
- **D-12-44 recorded**: CR-01's disposition is FIX. See `PHASE-12-DEBT.md`'s `### D-12-44` section for full evidence, restriction, and lifting condition.
- **D-12-45 recorded**: WR-01/WR-02/IN-01 dispositions, none deferred. See `PHASE-12-DEBT.md`'s `### D-12-45` section.

## Deviations from Plan

None — plan executed exactly as written. Task 1's checkpoint was answered by the human before this executor was spawned (continuation agent); Task 2 transcribed that answer without re-asking or second-guessing it. Task 3 ran verification-only with no code changes, as specified. One clarification, not a deviation: Task 3's literal `git diff --stat HEAD~3` verify command returned empty output because the three commits immediately before this plan's own are state-sync `docs` commits — this is recorded above with the wider, informative range also captured for the SUMMARY's own evidentiary purpose.

## Issues Encountered

None.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- Phase 12's two human-verification items (`12-VERIFICATION.md`) are both closed: CR-01's disposition is a recorded decision (D-12-44), and D-12-43 was ratified at an explicit blocking-human checkpoint (D-12-43's amended detail section).
- `PHASE-12-DEBT.md` is mechanically well-formed (10 items, `TestDebtRegistersAreWellFormed` green) and no prior row was rewritten.
- The workspace suite is green across all 25 packages; the spot-check that previously proved CR-01 open (`grep -rn "duplicate_payload_type" internal/compiler/` — no match) now proves it closed (8 matches).
- RES-02 and RES-03 are declared by plans 06, 07, and this plan (shared IDs). This is the LAST plan declaring them — per `requirements.ready-ids`, they become ready for `requirements.mark-complete` now that all three plans' summaries exist.
- Ready for phase re-verification. `12-VERIFICATION.md`, `.planning/REQUIREMENTS.md`, `.planning/ROADMAP.md`, and `.planning/STATE.md` were left byte-unchanged by this plan's own tasks (STATE.md/ROADMAP.md updates below are this executor's standard close-out step, run after Task 3, not part of Task 3 itself).

---
*Phase: 12-result-payloads*
*Completed: 2026-09-13*

## Self-Check: PASSED

- FOUND: `.planning/phases/12-result-payloads/12-08-SUMMARY.md`
- FOUND: `.planning/phases/12-result-payloads/PHASE-12-DEBT.md` (modified, verified via `git diff --stat`)
- FOUND: commit `1292c2d` (Task 1+2 combined docs commit)
- Re-ran plan-level `<verification>`: `TestDebtRegistersAreWellFormed` green (11 registers); `go test ./...` green across all 25 packages; all ten must-not-regress controls individually re-run and PASS; `git diff --stat` for this plan's own edit names only `PHASE-12-DEBT.md`.
- Re-ran Task 2's acceptance criteria: `items:` frontmatter (10) matches table row count; `D-12-44`/`D-12-45` rows present exactly once each with matching `### <ID>` sections; `D-12-43`'s original measurement narrative textually unchanged (only additions, verified via `git diff`); no pre-existing row deleted or reworded; no file outside `PHASE-12-DEBT.md` modified by Task 2.
- Re-ran Task 3's acceptance criteria: `go test ./...` exits 0, 25 package result lines recorded; `grep -rn "duplicate_payload_type" internal/compiler/` returns matches in both `check.go` and a test file, before/after states quoted above; evidence table present with one row per `12-VERIFICATION.md` Behavioral Spot-Check command; all ten must-not-regress controls listed with observed PASS; five of six D-12-36 emitters confirmed byte-unchanged with `emitBranch` named as the sole planned exception; `12-VERIFICATION.md`, `REQUIREMENTS.md`, `ROADMAP.md`, `STATE.md` confirmed byte-unchanged by this plan's task work (`git status --short` before the state-update step below showed only `PHASE-12-DEBT.md` and this new SUMMARY).
