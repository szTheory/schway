---
phase: 06-agent-feedback-and-performance-ratification
plan: 12
subsystem: compiler-diagnostics
tags: [defect-injection, mutation-testing, repair, DX-04, D-06-25, D-06-27, D-06-29]

requires:
  - phase: 06-11
    provides: "diagnostic.Repair{Span,Replacement,Applicability} and DriverEligible(), the exact shape every injector's mutated diagnostic must carry"
provides:
  - "session.Injector interface, five concrete injectors (match/move/borrow/cleanup/stale_evidence), and session.AllInjectors()"
  - "The shared fail-closed marker guard (phase6.injector_target_missing) every injector answers to, plus its non-inertness proof"
  - "testdata/phase6/'s held-out corpus with the heldout_/derivation_ prefix split (D-06-29)"
  - "New DriverEligible repairs on match.non_exhaustive, ownership.use_after_move, and ownership.borrow_conflict"
affects: ["06-13", "06-14", "06-15"]

actuals:
  tokens: 15800
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Marker-driven source-granularity injector: a trailing `// lang:<class>-target` comment names the single line an injector mutates; markerGuard refuses (phase6.injector_target_missing) when the marker's count is zero, and the LAST marked line is always the stable target when more than one exists."
    - "Thin-adapter injector over an already-shipped mutation runner: CleanupInjector calls ReleaseOmissionMutationRunner.Mutate (a method extracted from Run so there is exactly one release-marker scan in the tree) rather than reimplementing the scan."
    - "Self-mapping-arm gate on a diagnostic repair: match.non_exhaustive only offers a MachineApplicable add_missing_arm repair when every present arm already maps its own pattern to itself, which is both a safe mechanical guess AND the condition that keeps a pre-existing /0-schema golden fixture (Off => On, which does not self-map) byte-identical."
    - "Held-out/derivation corpus split by filename prefix, asserted disjoint-and-non-empty by a test that scales its per-class check from whatever heldout_*.lang fixtures exist on disk."

key-files:
  created:
    - internal/compiler/session/session_phase6_injectors.go
    - internal/compiler/session/session_phase6_injectors_test.go
    - testdata/phase6/README
    - testdata/phase6/heldout_match_defect.lang
    - testdata/phase6/derivation_match_defect.lang
    - testdata/phase6/heldout_move_defect.lang
    - testdata/phase6/derivation_move_defect.lang
    - testdata/phase6/heldout_borrow_defect.lang
    - testdata/phase6/derivation_borrow_defect.lang
    - testdata/phase6/stale_evidence_subject.lang
  modified:
    - internal/compiler/check/check.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_borrow_conflict_test.go

key-decisions:
  - "match.non_exhaustive's new add_missing_arm repair is gated on allArmsSelfMap(arms): a missing arm's target value is only a safe mechanical guess when every present arm already maps pattern to itself. This is not a fixture-path heuristic -- it is a semantic condition on the match's own existing arms -- and it is what keeps TestPhase1DiagnosticGoldenUnchanged's frozen lang.diagnostic/0 byte-for-byte identical (Off => On does not self-map, so the Phase 1 fixture never reaches the new repair path)."
  - "ownership.use_after_move gains moveTargetName (the surface binding name alongside the existing moveTargetID place ID) so its use_transfer_target repair can write real source text (Replacement) rather than only an internal place-ID Detail -- wired only in analyzeStraightLine, matching this file's existing asymmetry for insert_take (analyzeArmBody's own use_after_move repair stays classification-only)."
  - "ownership.borrow_conflict gains a narrow_to_shared_borrow repair only when the NEW loan is the conflicting borrow_mut (downgrade exclusive to shared, a genuine single-statement rewrite); the symmetric new-shared-vs-existing-exclusive case has no equally mechanical fix and stays classification-only. This required threading the failing ast.Binding into borrowConflictDiagnostic and updating TestBorrowConflictCauseChain's expected repair-kind list."
  - "CleanupInjector reuses ReleaseOmissionMutationRunner by calling a new Mutate method extracted from its existing Run (Run now calls Mutate too), rather than reimplementing the marker scan -- verified by a go/ast test that the injectors file contains no second `lang:release-site` literal."
  - "StaleEvidenceInjector's re-touch is a fixed appended comment line, not randomized content -- deterministic, and the locator is evidence.Validate's own mismatch report; the injector computes no source diff (asserted by go/ast: the file imports no reduce/differential package)."
  - "Cleanup and stale-evidence need no held-out .lang fixtures of their own: cleanup targets generated C (an existing NAT-03 fixture's cgen output already carries the marker), and stale-evidence is not a source defect at all, so there is no kind->edit mapping to overfit for either class."

patterns-established:
  - "phase6.injector_target_missing (session.InjectorError) is the one uniform refusal code all five injectors answer to, even where the underlying mechanism (e.g. ReleaseOmissionMutationRunner.Mutate's native.backend_control_invalid) has its own pre-existing code -- CleanupInjector wraps rather than replaces it, so Unwrap still reaches the original error."

requirements-completed: []  # DX-04 spans 06-11 through 06-15; NOT marked complete here per this plan's own instruction

coverage:
  - id: D1
    description: "Five defect injectors exist, one per named class, each answering to session.Injector"
    requirement: "DX-04"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_injectors_test.go#TestMatchDefectInjectorProducesExactlyOneDefect"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_injectors_test.go#TestMoveDefectInjectorProducesExactlyOneDefect"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_injectors_test.go#TestBorrowDefectInjectorProducesExactlyOneDefect"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_injectors_test.go#TestCleanupInjectorReusesReleaseOmissionRunner"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_injectors_test.go#TestStaleEvidenceInjectorBreaksManifestBinding"
        status: pass
    human_judgment: false
  - id: D2
    description: "Each of the four source-granularity injectors produces exactly one mechanical change, and every injector refuses (phase6.injector_target_missing) rather than silently passing when its target marker disappears"
    requirement: "DX-04"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_injectors_test.go#TestEveryInjectorProducesExactlyOneMechanicalChange"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_injectors_test.go#TestInjectorTargetChoiceIsSpecified"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_injectors_test.go#TestEveryInjectorRefusesWhenMarkerDisappears"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_injectors_test.go#TestInjectorMarkerCountGuardIsNotInert"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_injectors_test.go#TestInjectorRefusalPropagatesToExerciseFailure"
        status: pass
      - kind: other
        ref: "go test -race ./internal/compiler/session/... and go test -race ./... (full suite)"
        status: pass
    human_judgment: false
  - id: D3
    description: "The held-out corpus is structurally distinct from the derivation corpus (D-06-29)"
    requirement: "DX-04"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_injectors_test.go#TestPhase6DefectCorpusIsHeldOut"
        status: pass
    human_judgment: false

duration: ~35 min (not precisely timestamped at start; the three task commits span 2026-09-06T21:42:54-04:00 to 2026-09-06T21:51:12-04:00, excluding substantial upfront reading/design time before the first commit)
completed: 2026-09-06
status: complete
---

# Phase 06 Plan 12: Defect Injectors and Held-Out Corpus Summary

**Five DX-04 defect injectors (match, move, borrow, cleanup, stale-evidence), a held-out/derivation-split fixture corpus, and a fail-closed marker guard that every injector provably cannot silently bypass.**

## Held-Out Fixtures (D-06-29 — read before touching 06-13)

**Scored by the CI gate; NEVER consult when hand-deriving the repair driver's kind->edit mapping:**
- `testdata/phase6/heldout_match_defect.lang`
- `testdata/phase6/heldout_move_defect.lang`
- `testdata/phase6/heldout_borrow_defect.lang`

**Derivation-only; the ONLY fixtures the driver's kind->edit mapping may consult, never wired to any injector or scored by the gate:**
- `testdata/phase6/derivation_match_defect.lang`
- `testdata/phase6/derivation_move_defect.lang`
- `testdata/phase6/derivation_borrow_defect.lang`

**Neither prefix (no kind->edit mapping to overfit for this class at all):**
- `testdata/phase6/stale_evidence_subject.lang` — a small legitimate program to capture an evidence manifest over, then re-touch. Cleanup needs no dedicated fixture: it reuses `testdata/phase4/acquire_three_success.lang`'s already-generated C.

`testdata/phase6/README` states this rule in full. `testdata/phase6/heldout_match_defect.lang` also doubles as the corpus's characteristic marker file for a future `isPhase6Corpus`-shaped dispatch (06-15).

## Anti-Theater Guard 3 Confirmation

Every one of the five injectors, given an input shaped like an eligible one but with its target marker stripped, refuses with the typed `phase6.injector_target_missing` error rather than silently returning an unmutated source:

- `TestEveryInjectorRefusesWhenMarkerDisappears` iterates `session.AllInjectors()` (not a hand-written list), so a sixth injector added later without its own guard fails this test loudly.
- `TestInjectorMarkerCountGuardIsNotInert` proves the guard is load-bearing, not decorative: `matchInjectSkippingGuard` (a guard-DISABLED twin, never called by `MatchInjector.Inject` itself) demonstrably returns the marker-absent source completely unmutated, and that unmutated source checks clean — exactly the silent-pass theatre the guard exists to prevent. The guarded path, given the identical input, refuses.
- `TestInjectorRefusalPropagatesToExerciseFailure` drives the injector through `RunDefectInjectionExercise` and asserts the reported `Status` is `"fail"`, not merely that `Inject` returned an error — fail-closed all the way to the exercise's own result, not just at the injector boundary.

## Performance

- **Tasks:** 3 (Task 1 tracer, Task 2 auto+tdd, Task 3 auto+tdd)
- **Files:** 3 modified, 10 created
- **Commits:** 3 (one per task, each independently builds and passes its own `<verify>` block plus the full `internal/compiler/...` suite)

## Accomplishments

- Added `session.Injector` (interface), `session.MatchInjector`, `MoveInjector`, `BorrowInjector`, `CleanupInjector`, `StaleEvidenceInjector`, `session.AllInjectors()`, the shared `markerGuard`/`markerCount`/`lastMarkerLine` primitives, and `session.InjectorError{Code: "phase6.injector_target_missing"}`.
- Extended three existing diagnostics with real `Span`/`Replacement`/`ApplicabilityMachineApplicable` repairs so each injector's mutated fixture produces a genuinely `DriverEligible` diagnostic (06-11's `diagnostic.DriverEligible` predicate):
  - `match.non_exhaustive` gains `add_missing_arm`, gated on `allArmsSelfMap` (see Deviations).
  - `ownership.use_after_move` gains a fully-populated `use_transfer_target` (new `placeState.moveTargetName` field carries the surface binding name).
  - `ownership.borrow_conflict` gains `narrow_to_shared_borrow` when the new (conflicting) loan is exclusive.
- Extracted `ReleaseOmissionMutationRunner.Mutate` (marker-scan-and-delete only, no execution) out of its existing `Run`, so `CleanupInjector` calls the SAME method `Run` does — one release-marker scan in the whole tree, asserted by a `go/ast` test.
- Built the `testdata/phase6/` corpus with the `heldout_`/`derivation_` prefix split (D-06-29) and its README.
- Proved end-to-end, for each of match/move/borrow: a clean held-out fixture -> `Injector.Inject` -> `session.Check` -> exactly one rejecting diagnostic carrying a `DriverEligible` repair, with the unmutated fixture always checking clean first.
- Proved cleanup and stale-evidence end-to-end against real artifacts: `CleanupInjector` against `testdata/phase4/acquire_three_success.lang`'s generated C, `StaleEvidenceInjector` against a real `evidence.Build`/`evidence.Validate` manifest round-trip (skips gracefully if `clang` is unavailable on the host).
- Delivered anti-theater guard 3 as a first-class deliverable (Task 3), not test polish — see confirmation section above.

## Task Commits

1. **Task 1: Held-out corpus plus one injector, end-to-end to a repair-bearing diagnostic** — `6ada503` (feat)
2. **Task 2: The remaining four injectors — move, borrow, cleanup, stale-evidence** — `b5cd15d` (feat)
3. **Task 3: Anti-theater guard 3 — every injector's marker is itself mutation-tested** — `d43429e` (test)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing critical functionality] `match.non_exhaustive`, `ownership.use_after_move`, and `ownership.borrow_conflict` carried no `DriverEligible` repair material before this plan**
- **Found during:** Task 1 (match), Task 2 (move, borrow)
- **Issue:** The plan's Task 1/2 acceptance criteria require each injector's mutated fixture to produce a diagnostic whose `Repairs` contains a `DriverEligible` entry (non-nil `Span`, non-empty `Replacement`, `Applicability: MachineApplicable`). 06-11 extended `diagnostic.Repair`'s shape and wired exactly ONE call site (`insert_take`) end-to-end; the three call sites this plan's injectors target still only carried classification-only repairs (`Kind` alone, no `Span`/`Replacement`).
- **Fix:** Extended all three call sites following 06-11's own established pattern (`insert_take`'s `Span: &takeSpan, Replacement: "take " + binding.RHS.Source, Applicability: MachineApplicable` shape):
  - `match.non_exhaustive`: `add_missing_arm`, gated on a new `allArmsSelfMap` helper (see below).
  - `ownership.use_after_move`: `use_transfer_target` now carries `Span`/`Replacement` via a new `placeState.moveTargetName` field, wired only in `analyzeStraightLine` (matching the file's existing `insert_take` asymmetry — `analyzeArmBody`'s own copy is untouched).
  - `ownership.borrow_conflict`: `narrow_to_shared_borrow`, added only when the new loan is `borrow_mut` (downgrade to shared is the one uncontroversial single-statement fix); required threading the failing `ast.Binding` into `borrowConflictDiagnostic` (4 call sites, one shared function).
- **Files modified:** `internal/compiler/check/check.go`
- **Verification:** `go test ./internal/compiler/check/...`, `go test ./internal/compiler/session/...`, both `-race`; manual end-to-end runs against each held-out fixture.
- **Commit:** `6ada503` (match), `b5cd15d` (move, borrow)

**2. [Rule 1 - Bug avoidance / architectural guard] Naively wiring `match.non_exhaustive` onto `ErrorWithRepairs` broke a frozen Phase 1 golden**
- **Found during:** Task 1, first implementation attempt
- **Issue:** `match.non_exhaustive` was, since Phase 1, emitted via `diagnostic.Error` (schema `lang.diagnostic/0`). `TestPhase1DiagnosticGoldenUnchanged` pins the EXACT byte output of this diagnostic on `testdata/phase1/non_exhaustive.lang` (an `Off => On` match missing the `On` arm). Switching the call site to `ErrorWithRepairs` unconditionally changes the diagnostic's schema to `/1` (and therefore its ID) for EVERY caller, including that frozen Phase 1 fixture — this would have violated D-06-31's "Phase 1 evidence bytes remain byte-identical" and `lang.diagnostic/0`'s frozen-bytes invariant (D-06-24), since `ErrorWithRepairs` always uses `Schema1` even with zero repairs attached.
- **Fix:** Added `allArmsSelfMap(arms []ast.MatchArm) bool` as the gate on whether a repair is offered at all: only when every present arm already maps its own pattern to itself (a genuinely semantic condition on the match's own existing arms, not a fixture-path heuristic) does the call site use `ErrorWithRepairs`; otherwise it still calls the original `diagnostic.Error`. `testdata/phase1/non_exhaustive.lang`'s `Off => On` does not self-map, so it never reaches the new repair path, and `TestPhase1DiagnosticGoldenUnchanged` stays green unmodified. The held-out `heldout_match_defect.lang` fixture was deliberately designed with self-mapping arms (`Red => Red`, etc.) specifically so it DOES reach the new repair path.
- **Files modified:** `internal/compiler/check/check.go`
- **Verification:** `TestPhase1DiagnosticGoldenUnchanged` (unmodified, still green); `TestMatchDefectInjectorProducesExactlyOneDefect`.
- **Commit:** `6ada503`

**3. [Rule 1 - Bug] `TestBorrowConflictCauseChain`'s expected repair-kind list needed updating**
- **Found during:** Task 2
- **Issue:** Adding `narrow_to_shared_borrow` to `ownership.borrow_conflict` changed `shared_exclusive_reject.lang`'s repair set from `[create_loan_after_conflicting_loan_ends]` to `[create_loan_after_conflicting_loan_ends, narrow_to_shared_borrow]` (sorted), failing the pre-existing test's exact-match assertion.
- **Fix:** Updated the test's expected list with a comment explaining why the new repair appears for this specific fixture (its conflicting loan is the new exclusive one).
- **Files modified:** `internal/compiler/session/session_borrow_conflict_test.go`
- **Verification:** `go test ./internal/compiler/session/...`
- **Commit:** `b5cd15d`

**Total deviations:** 3 auto-fixed (2× Rule 2 missing critical functionality, 1× Rule 1 bug/regression-guard). **Impact:** All three were necessary for the plan's own acceptance criteria to be satisfiable at all (a `DriverEligible` repair cannot exist without material to carry it), and the second deviation is itself a load-bearing correctness discovery — a naive implementation would have silently broken a frozen Phase 1 compatibility guarantee. None expanded scope beyond what Task 1/2's acceptance criteria already required.

## Known Stubs

None. Every injector is fully functional against real fixtures and the real checker/evidence pipeline; no placeholder logic or hardcoded empty values.

## Issues Encountered

None blocking. `TestStaleEvidenceInjectorBreaksManifestBinding` skips gracefully (`t.Skip`) if the host has no `clang` on `PATH` — consistent with every other NAT-03/evidence test in this tree that depends on the native toolchain; it ran and passed on this host.

## User Setup Required

None.

## Next Phase Readiness

- 06-13 (repair driver): `session.AllInjectors()`, each `Injector.Inject`, and the `phase6.injector_target_missing` refusal code are the stable seam to drive defect injection from. `diagnostic.DriverEligible` (06-11) is already proven reachable through all three extended repairs.
- 06-13 MUST read `testdata/phase6/README` before touching any fixture: `derivation_*` fixtures are the only ones its kind->edit mapping may consult; `heldout_*` fixtures are CI-gate-scored only. 06-13 is expected to add a test asserting no `derivation_*` reference ever names a `heldout_*` path (per this plan's design, not yet implemented here).
- 06-13's byte-identity oracle (D-06-26) can rely on `TestEveryInjectorProducesExactlyOneMechanicalChange`'s guarantee: every source-granularity injector's mutation is exactly one line removed or rewritten.
- 06-14 (the remaining two anti-theater guards — prose-scramble and structured-vocabulary-removal, D-06-27.1/.2) and 06-15 (CLI dispatch, `isPhase6Corpus`, budget/debt register) are unaffected by anything in this plan beyond consuming its outputs.
- DX-04 is NOT marked complete — 06-13, 06-14, and 06-15 remain.
- Full suite (`go test ./...`), `go test -race ./...`, and `go vet ./...` are all clean at HEAD.

---
*Phase: 06-agent-feedback-and-performance-ratification*
*Completed: 2026-09-06*

## Self-Check: PASSED

All key-files (created/modified) verified present on disk with `[ -f ]`; all three task commit hashes (`6ada503`, `b5cd15d`, `d43429e`) verified present in `git log --oneline --all`. Full suite, `-race`, and `go vet` all clean at HEAD (re-verified after Task 3's commit).
