---
phase: 06-agent-feedback-and-performance-ratification
plan: 14
subsystem: compiler-diagnostics
tags: [defect-injection, anti-theater, repair, DX-04, D-06-26, D-06-27, D-06-28]

requires:
  - phase: 06-agent-feedback-and-performance-ratification
    provides: "06-12's five injectors and held-out corpus; 06-13's cmd/lang-repair driver, its byte-identity oracle, and its import-boundary lint"
provides:
  - "A fixture-substitution subprocess (stand-in) that lets any test drive the real cmd/lang-repair driver against arbitrary/mutated lang --json check output with zero driver code changes"
  - "Anti-theater guard 1: TestProseScrambleLeavesRepairBehaviourIdentical + TestProseScrambleFixtureKeepsStructuredFieldsIntact (D-06-27.1)"
  - "Anti-theater guard 2: TestVocabularyRemovalDrivesTheDriverRed + TestVocabularyRemovalGuardIsNotInert (D-06-27.2)"
  - "The differential-behaviour fallback oracle, split gate-side into internal/compiler/session (TestDifferentialFallbackOracleRunsInProcess), never inside the driver (D-06-26, D-06-28)"
  - "diagnostic.DriverEligible / cmd/lang-repair's driverEligible now also require a non-empty Kind (bug fix, see Deviations)"
affects: ["06-15"]

actuals:
  tokens: 12600
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Content-hash-keyed fixture-substitution subprocess: a tiny stand-in program, built once per test process, resolves which captured lang --json document to serve by SHA-256 of the CURRENT bytes of the source file it was told to check (its own last argv) -- never by counting invocations. This lets one stand-in binary correctly answer BOTH of the driver's calls (diagnose on pre-repair bytes, reverify on post-repair bytes) in a single-pass cycle with zero shared mutable state beyond a read-only captures directory."
    - "mutateCapture(raw, mode): a generic decode-to-map/marshal JSON mutation primitive, scoped mutations (strip_kind/span/replacement operate ONLY inside repairs[] elements, never causes[], which also has its own \"kind\")."
    - "Every guard is proven non-vacuous by construction: guard 1's scrambler is verified in both directions (TestProseScrambleFixtureKeepsStructuredFieldsIntact), guard 2's RED assertions are paired with a green control on the identical harness (TestVocabularyRemovalGuardIsNotInert), and both guards were demonstrated to actually fail once each during execution (recorded below, then reverted)."

key-files:
  created:
    - cmd/lang-repair/antitheater_test.go
    - cmd/lang-repair/testdata/match_diagnose_capture.json
    - cmd/lang-repair/testdata/match_reverify_capture.json
    - cmd/lang-repair/testdata/move_diagnose_capture.json
    - cmd/lang-repair/testdata/move_reverify_capture.json
    - cmd/lang-repair/testdata/borrow_diagnose_capture.json
    - cmd/lang-repair/testdata/borrow_reverify_capture.json
    - internal/compiler/session/session_phase6_oracle_test.go
  modified:
    - cmd/lang-repair/repair.go
    - internal/compiler/diagnostic/diagnostic.go
    - internal/compiler/diagnostic/diagnostic_test.go

key-decisions:
  - "Fixture-substitution subprocess over a driver-side test hook (06-RESEARCH.md Open Question 2): the stand-in is a separate, tiny program the test builds and passes as --lang; the real driver (repair.go, main.go) needed zero code changes for the mechanism itself, so there is no second, untested driver-internal code path. Proven faithful end to end before either guard depends on it (TestFixtureSubstitutionIsFaithful, the Task 1 tracer)."
  - "Guard 1's prose-scramble mechanism is scoped to the three classes (match/move/borrow) that actually go through the driver's own --json check/repair protocol. Cleanup and stale-evidence never decode a message/detail-bearing JSON field in their own repair mechanisms at all (cleanup: no JSON anywhere in its path -- pure cgen.EmitNative re-derivation; stale-evidence: its validateStatus closure decodes only {\"status\"}, asserted live by go/ast over repair_test.go) -- so their subtests prove that structural fact directly instead of faking a scramble the mechanism cannot even observe."
  - "Guard 2's four stripping modes and the differential-fallback oracle both use the move class's captures only (not all five) -- the plan's own Task 3 <behavior> list specifies four stripping-mode subtests, not a five-class matrix, unlike guard 1 which explicitly names all five classes."
  - "The differential-behaviour fallback oracle lives in internal/compiler/session/session_phase6_oracle_test.go, driving the REAL cmd/lang-repair BINARY as an external subprocess (built via `go build ./cmd/lang-repair`) rather than importing it -- Go forbids importing package main from anywhere, so this is the only way any test outside cmd/lang-repair itself can exercise the shipped driver at all."

patterns-established:
  - "A guard that can only ever assert one polarity (always-green or always-red) is untrustworthy until its own inertness is disproven: TestVocabularyRemovalGuardIsNotInert is the always-red guard's own green control, mirroring TestProseScrambleFixtureKeepsStructuredFieldsIntact's bidirectional check for the always-behavior-preserving guard."

requirements-completed: []  # DX-04 spans 06-11 through 06-15; NOT marked complete here per this plan's own instruction -- 06-15 also declares it

coverage:
  - id: D1
    description: "Fixture-substitution subprocess: a content-hash-keyed stand-in lets the real driver be driven against arbitrary/mutated lang --json check documents with zero driver code changes, proven faithful against the real binary end to end"
    requirement: "DX-04"
    verification:
      - kind: unit
        ref: "cmd/lang-repair/antitheater_test.go#TestFixtureSubstitutionIsFaithful"
        status: pass
      - kind: unit
        ref: "cmd/lang-repair/antitheater_test.go#TestBaselineCaptureContainsEligibleRepair"
        status: pass
      - kind: unit
        ref: "cmd/lang-repair/antitheater_test.go#TestMutateCaptureIdentityRoundTripIsByteStable"
        status: pass
    human_judgment: false
  - id: D2
    description: "Anti-theater guard 1 (prose-scramble): scrambling every message/detail string leaves repair behaviour (bytes, exit code, reported outcome) identical across match/move/borrow; cleanup/stale-evidence proven structurally prose-independent; the scrambler itself is verified bidirectionally"
    requirement: "DX-04"
    verification:
      - kind: unit
        ref: "cmd/lang-repair/antitheater_test.go#TestProseScrambleLeavesRepairBehaviourIdentical"
        status: pass
      - kind: unit
        ref: "cmd/lang-repair/antitheater_test.go#TestProseScrambleFixtureKeepsStructuredFieldsIntact"
        status: pass
      - kind: other
        ref: "Live falsification: temporarily made selectRepair require message text containing \"ownership transferred\"; match/move/borrow subtests failed as expected; reverted (repair.go confirmed byte-identical by diff)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Anti-theater guard 2 (vocabulary removal): stripping repairs[]/kind/span/replacement with prose intact drives the driver RED; a green control on the identical harness proves the guard is not inert"
    requirement: "DX-04"
    verification:
      - kind: unit
        ref: "cmd/lang-repair/antitheater_test.go#TestVocabularyRemovalDrivesTheDriverRed"
        status: pass
      - kind: unit
        ref: "cmd/lang-repair/antitheater_test.go#TestVocabularyRemovalGuardIsNotInert"
        status: pass
      - kind: other
        ref: "Live falsification: strip_kind subtest failed before the driverEligible Kind-check fix (deviation below), passed after"
        status: pass
    human_judgment: false
  - id: D4
    description: "The differential-behaviour fallback oracle lives CI-gate-side, in-process, in internal/compiler/session -- never inside the driver, which may not import internal/compiler/reduce"
    requirement: "DX-04"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_oracle_test.go#TestDifferentialFallbackOracleRunsInProcess"
        status: pass
      - kind: other
        ref: "go test ./... && go test -race ./... && go vet ./... (full suite)"
        status: pass
    human_judgment: false

duration: ~50 min
completed: 2026-09-06
status: complete
---

# Phase 06 Plan 14: DX-04 Anti-Theater Guards (Prose-Scramble and Vocabulary-Removal) Summary

**The two falsifiability-core guards for DX-04's repair exercise: a content-hash-keyed fixture-substitution subprocess proves the driver's behaviour is unchanged when every diagnostic message/detail string is scrambled to lorem ipsum, and unchanged-until-it-breaks when the structured repairs[]/kind/span/replacement channel is stripped instead -- both demonstrated to actually bite during execution, not merely asserted.**

## How each guard was proven to bite

Per this plan's own success criterion, a guard nobody demonstrated failing is a guard nobody can trust. Both guards were made to fail live, then reverted:

- **Guard 1 (prose-scramble).** Temporarily changed `selectRepair` in `repair.go` to require the diagnostic's own `message` text to contain `"ownership transferred"` before selecting any repair. Re-ran `TestProseScrambleLeavesRepairBehaviourIdentical`: the `move` subtest failed on a differing exit code (identity=0, scrambled=1) because the lorem-ipsum-scrambled message no longer matched; `match`/`borrow` failed earlier, at the harness's own fixture-quality precondition, because their real diagnose captures never contained that substring to begin with. Reverted via a saved pre-edit copy; `diff` confirmed `repair.go` is byte-identical to its pre-experiment state, and the full guard suite re-ran green.
- **Guard 2 (vocabulary removal).** First implementation of `TestVocabularyRemovalDrivesTheDriverRed`'s `strip_kind` subtest failed to go red: the driver exited 0 and reported `repaired`, because `driverEligible`/`diagnostic.DriverEligible` never actually checked `Kind` for eligibility (see Deviations below) -- an empty-`Kind` repair still spliced its `Span`/`Replacement` in just fine. This was the guard doing its job: catching a genuine gap in the driver's own claimed "structured channel only" contract. Fixed both eligibility checks to also require a non-empty `Kind`; re-ran and all four stripping-mode subtests went red as required, with the paired `TestVocabularyRemovalGuardIsNotInert` control staying green on the identical harness.

## Performance

- **Tasks:** 3 (Task 1 tracer, Task 2 auto+tdd, Task 3 auto+tdd)
- **Files:** 3 created (antitheater_test.go, oracle test, 6 testdata captures counted as one logical unit), 3 modified (repair.go, diagnostic.go, diagnostic_test.go)
- **Commits:** 3 (one per task)

## Accomplishments

- Built the fixture-substitution subprocess (Task 1): a tiny stand-in program, built once per test process, that resolves which captured `lang --json check` document to serve by SHA-256 of the CURRENT content of the source file it is told to check (its own last argv) -- correctly answering both the driver's diagnose and reverify calls in a single-pass cycle without any per-invocation counter. Proven faithful against the real `lang` binary end to end (`TestFixtureSubstitutionIsFaithful`), and proven to cost the driver zero code changes for the mechanism itself.
- Generated and committed six real per-class baseline captures (`match`/`move`/`borrow`, each with a `_diagnose_` and `_reverify_` capture) by running the real shipped `lang` binary against each injector's real mutated held-out fixture -- never fabricated by hand. `TestBaselineCaptureContainsEligibleRepair` guards their quality.
- Shipped anti-theater guard 1 (`TestProseScrambleLeavesRepairBehaviourIdentical`, `TestProseScrambleFixtureKeepsStructuredFieldsIntact`): scrambling every `message`/`detail` string to a fixed lorem-ipsum value leaves repaired bytes, exit code, and reported outcome identical across match/move/borrow; cleanup and stale-evidence proven structurally prose-independent rather than faked.
- Shipped anti-theater guard 2 (`TestVocabularyRemovalDrivesTheDriverRed`, `TestVocabularyRemovalGuardIsNotInert`): four stripping modes (`strip_repairs`, `strip_kind`, `strip_span`, `strip_replacement`) each drive the driver to a non-zero exit and an `unrepairable` outcome with prose left fully intact; a green control on the identical harness proves the guard is not inert.
- Shipped the differential-behaviour fallback oracle split gate-side (`internal/compiler/session/session_phase6_oracle_test.go#TestDifferentialFallbackOracleRunsInProcess`): drives the real `cmd/lang-repair` binary as an external subprocess, then, entirely after it has exited, re-runs the shipped `-O0`-vs-`-O3` reducer/mismatch comparison in-process over the repaired file. A `go/ast` assertion proves `cmd/lang-repair` never imports `internal/compiler/reduce`.
- Fixed a genuine gap (`diagnostic.DriverEligible` / `repair.go`'s `driverEligible`): both now require a non-empty `Kind`, not just `Applicability`/`Span`/`Replacement`.

## Task Commits

1. **Task 1: The fixture-substitution subprocess, proven on one real capture** - `69a13a9` (feat)
2. **Task 2: Anti-theater guard 1 -- the prose-scramble test** - `6fae4e8` (test)
3. **Task 3: Anti-theater guard 2 -- vocabulary removal must drive the driver RED** - `0d57744` (test)

## Files Created/Modified

- `cmd/lang-repair/antitheater_test.go` - Both anti-theater guards, the fixture-substitution stand-in, `mutateCapture` and its six modes
- `cmd/lang-repair/testdata/{match,move,borrow}_{diagnose,reverify}_capture.json` - Six real baseline captures from the shipped `lang` binary
- `internal/compiler/session/session_phase6_oracle_test.go` - The differential-behaviour fallback oracle, gate-side
- `cmd/lang-repair/repair.go` - `driverEligible` now also requires a non-empty `Kind` (Task 3 deviation)
- `internal/compiler/diagnostic/diagnostic.go` - `DriverEligible` now also requires a non-empty `Kind` (Task 3 deviation)
- `internal/compiler/diagnostic/diagnostic_test.go` - Added an empty-`Kind` test case to `TestMachineApplicableWithoutMaterialIsNotEligible`

## Decisions Made

See `key-decisions` in frontmatter. Summary: fixture-substitution subprocess over a driver-side hook (RESEARCH's own recommendation); guard 1 scoped precisely to the three classes that use the driver's own protocol, with cleanup/stale-evidence covered by structural (not faked) prose-independence proofs; guard 2 and the fallback oracle both use only the `move` class, matching the plan's own narrower Task 3 `<behavior>` spec; the fallback oracle drives the real driver binary as an external subprocess since `package main` can never be imported.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `driverEligible`/`diagnostic.DriverEligible` never actually required a non-empty `Kind`**
- **Found during:** Task 3, first run of `TestVocabularyRemovalDrivesTheDriverRed`'s `strip_kind` subtest
- **Issue:** D-06-27.2 requires "stripping only `kind` from each repair makes the driver go red." The shipped (06-11/06-13) `DriverEligible`/`driverEligible` functions only ever checked `Applicability == MachineApplicable && Span != nil && Replacement != ""` -- `Kind` was treated as pure reporting metadata, never load-bearing. Stripping `kind` alone therefore left the repair fully eligible and the driver applied it successfully, contradicting the plan's own explicit truth requirement and, independently, letting an unclassified/malformed repair record through the eligibility gate.
- **Fix:** Added `r.Kind != ""` to both `diagnostic.DriverEligible` (internal/compiler/diagnostic/diagnostic.go) and `cmd/lang-repair`'s local mirror `driverEligible` (repair.go), keeping the two definitions in lockstep as repair.go's own doc comment requires. Added a covering case to `TestMachineApplicableWithoutMaterialIsNotEligible`. Every existing production call site (match/move/borrow/insert_take repairs in check.go) already sets a non-empty `Kind` on every repair it emits, so this is purely a tightening of a previously-unenforced invariant with zero behavioural change to any real diagnostic.
- **Files modified:** `cmd/lang-repair/repair.go`, `internal/compiler/diagnostic/diagnostic.go`, `internal/compiler/diagnostic/diagnostic_test.go`
- **Verification:** `TestOnlyMachineApplicableIsDriverEligible`, `TestMachineApplicableWithoutMaterialIsNotEligible`, full `cmd/lang-repair` suite, full `go test ./...`, `go test -race ./...`, `go vet ./...` all clean.
- **Commit:** `0d57744` (Task 3)

**2. [Rule 4-adjacent, self-corrected] Task 1's `TestRepairDriverAndMainAreUnchangedByThisPlan` was removed in Task 3**
- **Found during:** Task 3, once the `driverEligible` fix above required touching `repair.go`
- **Issue:** Task 1 had added a test asserting `repair.go`/`main.go` exist unmodified "by this plan" (satisfying Task 1's own acceptance criterion at the time it was committed). Task 3's Rule 1 bug fix above legitimately modifies `repair.go` for a documented, necessary reason. Leaving the old test in place (it only checked file existence, not content-hash, so it would not have failed) would have been a misleading, no-longer-accurate claim in a test's own doc comment.
- **Fix:** Removed the test and its doc-comment reference; the file-level header comment now points to this SUMMARY for the Task 3 `repair.go` change instead.
- **Files modified:** `cmd/lang-repair/antitheater_test.go`
- **Verification:** `diff` against the pre-Task-3 committed `repair.go` shows exactly the two-line `driverEligible` change; no other drift.
- **Commit:** `0d57744` (Task 3)

---

**Total deviations:** 2 (1 Rule 1 bug fix, 1 self-corrective test removal directly caused by the first). **Impact:** The Rule 1 fix is necessary for D-06-27.2's own literal requirement to be satisfiable at all, and independently closes a real fail-open eligibility gap (an unclassified repair could otherwise be mechanically applied). No scope creep beyond `cmd/lang-repair`'s eligibility check and its mirror in `diagnostic.go`.

## Known Stubs

None. Every guard runs against real captures from the real shipped binary, the real injectors, and (for the fallback oracle) the real built `cmd/lang-repair` binary as an external subprocess.

## Issues Encountered

None blocking beyond the two deviations above, both resolved within their own task.

## User Setup Required

None.

## Next Phase Readiness

- 06-15 can build the remaining DX-04 surface (CLI dispatch, `isPhase6Corpus`, budget/debt register, the recorded non-gating agent-legibility exercise) on top of everything this plan and 06-12/06-13 shipped, without modification.
- DX-04 is NOT marked complete by this plan alone -- 06-15 also declares it, per this plan's explicit instruction.
- The fixture-substitution stand-in (`buildStandin`, `writeCaptureForContent`, `mutateCapture`) and the checked-in per-class captures are reusable by any future guard that needs to drive the real driver against synthetic `lang --json check` output.
- Full suite (`go test ./...`), `go test -race ./...`, and `go vet ./...` are all clean at HEAD.

---
*Phase: 06-agent-feedback-and-performance-ratification*
*Completed: 2026-09-06*

## Self-Check: PASSED

All key-files (created/modified) verified present on disk. All three task commit hashes (`69a13a9`, `6fae4e8`, `0d57744`) verified present in `git log --oneline --all`. All plan `<verify>` commands re-run via `scripts/assert-go-tests.sh` with `GOCACHE=/tmp/ai-lang-phase6-cache` and pass. Full suite, `-race`, and `go vet` all clean at HEAD.
