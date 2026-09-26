---
phase: 14-evidence-instrument-and-honest-scoping
plan: 07
subsystem: process
tags: [go, debt-register, evd-03, evd-04, witness-grammar, executed-probes, suppression-enumerator, generated-view]

requires:
  - phase: 14-evidence-instrument-and-honest-scoping
    provides: "plan 14-04's closed three-form owning-phase vocabulary and debtRegisterProblems (the pure, non-fataling well-formedness law this plan extends with Grade/Witness)"
provides:
  - "internal/compiler/session/session_test.go's checkDebtRegister extended with Grade and Witness columns under a closed four-kind executed witness grammar (probe:/callsite:/escape:/env:), plus the two D-14-29 closure assertions"
  - "internal/compiler/session/witness_registry_test.go -- four executed probes (TestB1BlameIsStructurallyUnreachable, TestD1243ControlIsUnconstructible, TestLTOInertnessOnMultiFunctionEmission, TestRetainedPointerEscapeIsStillUnsubjected), the module-wide suppression enumerator (TestNoSuppressionOutlivesItsWitness) with its three-seeded-fault non-inertness proof, and the generated .planning/UNREACHABLE-CLAIMS.md view with fail-closed anti-decay guards"
  - "PHASE-13-DEBT.md, PHASE-14-DEBT.md, PHASE-11-DEBT.md, PHASE-12-DEBT.md fully migrated to Grade/Witness (32 rows across four registers)"
affects: ["plan 14-08 (removes the PENDING-05-08 marker this plan left carrying an interim citation, as part of collapsing the axis-movement law)", "any future plan authoring a new *-DEBT.md register or a new t.Skip call"]

actuals:
  tokens: 36581
  tasks: 4
  commits: 5

tech-stack:
  added: []
  patterns:
    - "Closed four-kind witness grammar (probe:/callsite:/escape:/env:) enforced by regex + resolution function per kind, mirroring the existing closed-vocabulary-map pattern (debtRegisterSeverities) but with EXECUTED resolution instead of set membership"
    - "Two-tier suppression-citation resolution mode (suppressionModeFull vs suppressionModePendingOnly): the live module's blanket comment/string-literal scan uses a narrower grammar than a Skip call's own citation or an isolated synthetic test fixture, to avoid colliding with pre-existing unrelated codebase vocabulary"
    - "Generated, byte-compared view derived from a closed syntactic predicate over register data (Witness cell contains a probe: token) rather than a hand-curated list"

key-files:
  created:
    - internal/compiler/session/witness_registry_test.go
    - testdata/phase14/blame_unreachable_admission_refusal.lang
    - testdata/phase14/multi_function_match_refusal.lang
    - .planning/UNREACHABLE-CLAIMS.md
  modified:
    - internal/compiler/session/session_test.go
    - internal/compiler/session/session_phase6_injectors_test.go
    - internal/compiler/session/session_phase11_differential_test.go
    - internal/compiler/native/native_lto_test.go
    - internal/compiler/native/symbols_test.go
    - internal/compiler/cgen/cgen_program_test.go
    - internal/compiler/session/session_payload_replay_test.go
    - internal/compiler/session/session_phase5_corpus_test.go
    - internal/compiler/session/session_phase6_evidence_test.go
    - internal/compiler/session/verification_groundedness_test.go
    - .planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/PHASE-13-DEBT.md
    - .planning/milestones/M002-phases/12-result-payloads/PHASE-12-DEBT.md
    - .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/PHASE-11-DEBT.md
    - .planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md
    - .planning/LANGUAGE-MATURITY.md
    - .planning/REQUIREMENTS.md

key-decisions:
  - "Grade/Witness columns migrated in FULL for four registers (PHASE-11, PHASE-12, PHASE-13, PHASE-14-DEBT.md -- 32 rows) rather than only the specific rows D-14-30's day-one table names. Once a register's Items table carries the Grade/Witness header, every row in that table must satisfy the column-presence-and-non-emptiness law, so partial per-row mechanization inside an already-columned table is not constructible; DEFINED/n/a is the honest floor for a row with no executed claim to substantiate (see the DEFINED exemption decision below), making full-file migration cheap once that exemption exists."
  - "debtRegisterGradeWitnessRowProblems exempts DEFINED-graded rows from the D-14-29 'below-bar row needs a witness or explicit withdrawal' requirement -- extending the plan's literal text. D-14-03's own ladder treats DEFINED as 'the floor, always permitted': a bare declaration asserts nothing a witness could substantiate. Without this exemption, PHASE-14-DEBT.md's D-14-46/D-14-47 (genuinely open, prospective debt with no probe to cite yet) would have no honest recording: marking them '(withdrawn)' would misstate that the claim was retracted, when it was not."
  - "The 87 rows across 02-DEBT.md..PHASE-10-DEBT.md remain on debtRegisterGradeWitnessExemptions (frozen prior art, unmigrated) -- deriving and witnessing that much historical data with real executed probes is out of this plan's budget. Recorded as debt (see Deferred/Debt below), not silently dropped."
  - "TestNoSuppressionOutlivesItsWitness's LIVE-MODULE comment/string-literal surface is checked ONLY for the PENDING-NN-NN citation shape (suppressionModePendingOnly), not the full probe:/callsite:/escape:/env:/D-XX-NN grammar. A first implementation attempt at the full grammar produced 115 false positives: this codebase already uses 'escape:' as its own, unrelated, pre-existing anti-theater vocabulary (session_phase5_escapes.go's EscapeCoordinatedSourceToCoreFalseClaim, originvalidate.go's KnownEscape -- a 'coordinated lie' escape hatch, nothing to do with this plan's NAT03Mutation escape registry) and 'D-XX-NN' as its own pervasive decision/finding cross-reference convention used in nearly every doc comment, not exclusively as a suppression citation. Skip-call sites still use the full grammar (this codebase's Skip citations use only this plan's shapes), as do the isolated synthetic-package non-inertness tests."
  - "The call-site witness scan (D-13-02b's callsite:internal/compiler/check.resolveBlame=0) counts PRODUCTION (non-_test.go) call sites only, excluding unit tests that call resolveBlame directly to exercise it in isolation. This matches the exact claim D-13-02b makes -- resolveBlame is unit-tested but never reached from the production checking pipeline -- and is a deliberate scoping decision the plan text did not spell out at this granularity."
  - "cgen_program_test.go's t.Skipf on a missing checked-in fixture (testdata/phase07/cycle_self.lang, which exists) was a Rule 1 bug, not a legitimate skip: fixture absence would be repo corruption. Changed to t.Fatalf while adding suppression citations to the genuinely uncited skips."

patterns-established:
  - "A witness cell's resolution logic (debtRegisterWitnessTokenProblem) is reused, not reimplemented, everywhere a probe:/escape:/env: citation can appear -- the register's own Witness column, a t.Skip call's citation, and the suppression enumerator's synthetic-fixture tests all call the same function, so 'resolves' means one thing project-wide."

requirements-completed: [EVD-03, EVD-04]

coverage:
  - id: D1
    description: "checkDebtRegister extended with Grade and Witness columns under a closed four-kind executed witness grammar (probe:/callsite:/escape:/env:); free text and a bare phase identifier both fail; a missing call-site symbol fails rather than counting zero; the two D-14-29 closure assertions (below-bar-needs-witness-or-withdrawal, no identifier claimed by two graded registers) are implemented with passing negative subtests"
    requirement: "EVD-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestDebtRegisterWitnessGrammarIsClosed"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestDebtRegisterIdentifierUniquenessGuardIsNotInert"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestDebtRegistersAreWellFormed"
        status: pass
    human_judgment: false
  - id: D2
    description: "Four executed probes back the day-one register contents: TestB1BlameIsStructurallyUnreachable (D-13-02b/D-13-10a), TestD1243ControlIsUnconstructible (D-12-43), TestLTOInertnessOnMultiFunctionEmission (D-14-45), TestRetainedPointerEscapeIsStillUnsubjected (the retained_pointer escape). Each asserts a structural refusal/fact, not a message string, and each is proven to actually depend on its fixture (the neutralized-probe seeded fault in TestSuppressionWitnessGuardIsNotInert)"
    requirement: "EVD-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/witness_registry_test.go#TestB1BlameIsStructurallyUnreachable"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/witness_registry_test.go#TestD1243ControlIsUnconstructible"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/witness_registry_test.go#TestLTOInertnessOnMultiFunctionEmission"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/witness_registry_test.go#TestRetainedPointerEscapeIsStillUnsubjected"
        status: pass
    human_judgment: false
  - id: D3
    description: "Every suppression surface in the module (Skip/Skipf/SkipNow calls, //go:build constraints, comments, string literals, across production and test files) is enumerated and its citation checked for resolution; the guard is proven non-inert on three independent seeded fault kinds (uncited skip, flipped callsite count, neutralized/XPASS probe); the module's blanket comment/string-literal surface is deliberately scoped to the PENDING-NN-NN shape only (see key-decisions) to avoid colliding with unrelated pre-existing codebase vocabulary"
    requirement: "EVD-04"
    verification:
      - kind: unit
        ref: "internal/compiler/session/witness_registry_test.go#TestNoSuppressionOutlivesItsWitness"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/witness_registry_test.go#TestSuppressionWitnessGuardIsNotInert"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/witness_registry_test.go#TestSuppressionWitnessOnlyInStringLiteralIsFound"
        status: pass
    human_judgment: true
    rationale: "The scoping decision (PENDING-only for the live blanket comment/string-literal scan, full grammar for Skip sites and isolated synthetic tests) is a judgment call that trades literal coverage of D-14-25's text for avoiding a 115-false-positive collision with this codebase's own pre-existing 'escape:' and 'D-XX-NN' conventions. A human should confirm this tradeoff is acceptable rather than treating the mechanical PASS as the full story."
  - id: D4
    description: ".planning/UNREACHABLE-CLAIMS.md exists, is git-tracked, and holds exactly the six known qualifying rows (D-13-02b, D-13-10a, D-13-34, D-14-45, D-11-02, D-12-43), each with grade, witness and unblocking trigger; the view is regenerated in memory and byte-compared (no blessing path); the non-zero-row-count, hand-edit-detection, and stale-entry anti-decay guards are each proven load-bearing"
    requirement: "EVD-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/witness_registry_test.go#TestUnreachableClaimsViewIsCurrent"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/witness_registry_test.go#TestUnreachableClaimsViewNonZeroRowCountIsLoadBearing"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/witness_registry_test.go#TestUnreachableClaimsViewCatchesHandEdit"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/witness_registry_test.go#TestUnreachableClaimsViewCatchesStaleEntry"
        status: pass
      - kind: other
        ref: "manually appended one character to the checked-in .planning/UNREACHABLE-CLAIMS.md and confirmed TestUnreachableClaimsViewIsCurrent failed, then reverted"
        status: pass
    human_judgment: false

duration: 26min (commit-to-commit span across 5 task commits; full-suite `go test ./...` ran after, 160s for internal/compiler/session alone)
completed: 2026-09-18
status: complete
---

# Phase 14 Plan 07: Witness Grammar, Executed Probes, Suppression Enumerator, and the Unreachable-Claims View Summary

**A closed four-kind executed witness grammar (`probe:`/`callsite:`/`escape:`/`env:`) now backs every unreachable claim in four fully-migrated debt registers, four new probes prove the day-one claims currently hold, a module-wide suppression enumerator requires every `Skip` call to cite a resolvable witness, and `.planning/UNREACHABLE-CLAIMS.md` is a generated, byte-compared view derived from the registers.**

## Performance

- **Duration:** 26 min (commit-to-commit span, 23:19:56 → 23:45:13 local, across the plan's 5 task commits)
- **Started:** 2026-09-17T23:19:56-04:00
- **Completed:** 2026-09-17T23:45:13-04:00
- **Tasks:** 4
- **Files created/modified:** 19 (4 created, 15 modified in the task commits; +2 REQUIREMENTS.md/SUMMARY.md in the metadata commit)

## Accomplishments

- `internal/compiler/session/session_test.go`: `checkDebtRegister` extended in place with `Grade` and `Witness` columns under a closed four-kind witness grammar. A witness token is exactly one of: `probe:<TestName>` (resolved via a static `go/parser` scan of every `*_test.go` file, never `go test -list`), `callsite:<pkgPath>.<Symbol>=<N>` (an abstract-syntax-tree call-site count over the package's PRODUCTION files only; a missing symbol fails, never reads as a count of zero), `escape:<id>` (must resolve in the closed `debtRegisterEscapeRegistry`, whose own entry must itself carry a resolving `probe:`), or `env:<id>` (a closed, never-closing, explicitly-not-debt environmental set — `clang`, `otool`). Free text and a bare phase identifier (`P17`) both fail by construction. Both D-14-29 closure assertions are implemented: a below-bar row needs a witness or an explicit `(withdrawn)` marker (with `DEFINED` exempted as the no-claim floor), and no identifier is claimed by two graded registers.
- `internal/compiler/session/witness_registry_test.go` (new, 810 lines): four executed probes, the suppression enumerator (`TestNoSuppressionOutlivesItsWitness`, walking every `.go` file — production and test — for Skip calls, `//go:build` constraints, comments, and string literals), its three-seeded-fault non-inertness proof, and the generated-view machinery (`deriveUnreachableClaims`/`renderUnreachableClaimsView`) with four dedicated anti-decay tests.
- Two new fixtures under `testdata/phase14/`: `blame_unreachable_admission_refusal.lang` (a function whose declared return type contradicts its own parameter type — the exact admission-time refusal D-13-02b's claim depends on) and `multi_function_match_refusal.lang` (a two-function, all-Match-bodied program with one call edge — the exact shape `cgen.emitProgram` refuses, backing D-14-45).
- `PHASE-13-DEBT.md`, `PHASE-14-DEBT.md`, `PHASE-11-DEBT.md`, `PHASE-12-DEBT.md` fully migrated to Grade/Witness (3+3+16+10 = 32 rows). The four claims D-14-30's day-one table names (D-13-02b, D-13-10a, D-13-34, D-14-45, D-11-02, D-12-43 — six rows, two registers each contributing one real probe) carry real `WIRED`/`DEFINED (withdrawn)` grades and resolving `probe:` witnesses; every other row in those four files is `DEFINED`/`n/a` (a bare declaration, no executed claim to substantiate).
- `.planning/UNREACHABLE-CLAIMS.md` generated and checked in: 6 entries, exactly the six known qualifying rows, byte-compared against a fresh in-memory regeneration on every test run.
- Collateral: two existing tests (`TestLanguageMaturityCountsAreCurrent`, the groundedness lint's `TestVerificationGroundednessFrontierIsPinned`) needed re-derivation once the two new testdata fixtures and the two new test names existed — both fixed (see Deviations).

## Task Commits

1. **Task 1 RED: witness grammar** — `689fde3` (test)
2. **Task 1 GREEN + Task 2: probes and data migration** — `72ecb6f` (feat)
3. **Task 3 RED: suppression enumerator** — `17ff8d1` (test)
4. **Task 3 GREEN: cite every remaining uncited Skip call** — `f1651e0` (feat)
5. **Task 4 GREEN: generated UNREACHABLE-CLAIMS.md view** — `b8242d1` (feat)

_Note: Task 1 and Task 2 are tightly coupled by the plan's own design (Task 1's `<verify>` literally states "exits 0 after Task 2 supplies the probes"), so their RED/GREEN pair spans two commits: `689fde3` is the RED commit (the grammar law compiles and correctly fails — both because two registers lack the new columns and because one probe token doesn't resolve yet), `72ecb6f` is the GREEN commit that supplies the missing data and the probe. Task 4 (generating the view) had no meaningful RED state of its own to commit separately — deriving the view requires the fully-migrated Task 1-3 state to exist first — so it is GREEN-only, consistent with its `type="auto"` (non-`tdd`) frontmatter._

**Metadata commit:** (this SUMMARY + REQUIREMENTS.md + STATE.md + ROADMAP.md)

## Files Created/Modified

- `internal/compiler/session/witness_registry_test.go` — four executed probes, the suppression enumerator, the generated-view machinery.
- `testdata/phase14/blame_unreachable_admission_refusal.lang`, `testdata/phase14/multi_function_match_refusal.lang` — new fixtures.
- `.planning/UNREACHABLE-CLAIMS.md` — new generated view.
- `internal/compiler/session/session_test.go` — Grade/Witness grammar, escape/environmental registries, the two closure assertions, `debtRegisterGradeWitnessExemptions`.
- `PHASE-13-DEBT.md`, `PHASE-14-DEBT.md`, `PHASE-11-DEBT.md`, `PHASE-12-DEBT.md` — Grade/Witness columns added to every row.
- `session_phase6_injectors_test.go`, `session_phase11_differential_test.go`, `native_lto_test.go`, `native/symbols_test.go`, `cgen_program_test.go`, `session_payload_replay_test.go`, `session_phase5_corpus_test.go`, `session_phase6_evidence_test.go` — Skip-call citations added or corrected.
- `verification_groundedness_test.go` — two stale pinned-frontier entries removed.
- `.planning/LANGUAGE-MATURITY.md` — corpus counts re-derived (126→128 programs, 4244→4311 lines).

## Decisions Made

See `key-decisions` in frontmatter for full detail. In summary: (1) Grade/Witness migrated in full for four registers rather than row-by-row, because a columned table cannot host partial mechanization; (2) `DEFINED` is exempted from the witness-or-withdrawal requirement, extending the plan's literal D-14-29 text to make genuinely-open, not-yet-probed debt recordable honestly; (3) the remaining 87 rows across `02-DEBT.md`..`PHASE-10-DEBT.md` stay exempt, frozen prior art, recorded as debt below; (4) the live-module suppression scan is deliberately narrowed to the `PENDING-NN-NN` shape after a full-grammar first attempt produced 115 false positives against this codebase's own pre-existing, unrelated `escape:`/`D-XX-NN` conventions; (5) the call-site witness for `resolveBlame` counts production call sites only, matching the exact claim being witnessed.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `cgen_program_test.go`'s defensive Skip on a missing checked-in fixture**
- **Found during:** Task 3 (suppression enumerator's first full-module scan)
- **Issue:** `t.Skipf("no self-recursion fixture available: %v", err)` guarded a read of `testdata/phase07/cycle_self.lang`, which is checked into the repo and always present — a read failure there is repo corruption, not a legitimate skip condition, and the skip was also uncited.
- **Fix:** Changed to `t.Fatalf`.
- **Files modified:** `internal/compiler/cgen/cgen_program_test.go`
- **Verification:** `go test ./internal/compiler/cgen/...` still passes (fixture exists, so the Fatalf path is never hit in practice).
- **Committed in:** `f1651e0`

**2. [Rule 3 - Blocking] Collateral test breakage from two new testdata fixtures and two new test names**
- **Found during:** end-of-Task-4 full-suite run
- **Issue:** `testdata/phase14/*.lang` (2 new files) shifted `LANGUAGE-MATURITY.md`'s machine-checked corpus counts (126→128 programs, 4244→4311 lines); creating `TestUnreachableClaimsViewIsCurrent` and `TestNoSuppressionOutlivesItsWitness` made two `pinnedFrontier` entries in the groundedness lint stale (the patterns those entries recorded as dead now resolve to real, passing tests).
- **Fix:** Updated `LANGUAGE-MATURITY.md`'s stated counts; removed the two now-resolved entries from `pinnedFrontier` in `verification_groundedness_test.go` — exactly D-14-18's "each reconciliation shrinks the pinned literal" design.
- **Files modified:** `.planning/LANGUAGE-MATURITY.md`, `internal/compiler/session/verification_groundedness_test.go`
- **Verification:** `go test ./internal/compiler/session/... -run 'TestLanguageMaturityCountsAreCurrent|TestSelfDescribingDocsGuardIsNotInert|TestVerificationGroundednessFrontierIsPinned'` passes.
- **Committed in:** `b8242d1`

---

**Total deviations:** 2 auto-fixed (1 bug, 1 blocking). **Impact:** Both are necessary consequences of adding real, checked-in evidence (a fixture, a passing test) to a tree that mechanically re-derives its own counts and pins its own known-dead patterns — exactly the "instrument catches drift" behavior this phase exists to install. No scope creep.

## Issues Encountered

**Scope-narrowing on Task 3's suppression enumerator (documented as a decision, not silently absorbed).** D-14-25's literal text asks the enumerator to check every comment and string literal in the module for pending-marker, decision-identifier, and escape-identifier citations. A first implementation attempt at the full grammar produced 115 false positives: this codebase already uses `escape:` as its own, unrelated, pre-existing anti-theater vocabulary (`session_phase5_escapes.go`'s `EscapeCoordinatedSourceToCoreFalseClaim`, `originvalidate.go`'s `KnownEscape`) and `D-XX-NN` as a pervasive decision/finding cross-reference used in the vast majority of doc comments, not exclusively as a suppression citation. The live module's blanket comment/string-literal scan is scoped to the `PENDING-NN-NN` shape only; Skip-call sites and the isolated synthetic-package non-inertness tests still use the full grammar. See `coverage` D3's `human_judgment: true` rationale.

**87 debt-register rows remain unmigrated to Grade/Witness**, recorded as debt below rather than silently deferred.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- EVD-03 and EVD-04 both close: every unreachable claim this plan's scope covers is justified by an executed probe or an honest `DEFINED`/no-claim floor, and every suppression surface this plan's scope covers cites a resolvable witness.
- Plan 14-08 has a clean, documented hand-off: the `PENDING-05-08` marker in `session_phase5_alias_test.go` now carries an interim citation (accepted by shape under the suppression enumerator) and a code comment stating plan 14-08 removes it entirely when collapsing the axis-movement law (EVD-05).
- `go test ./...` is green across all 25 packages; `go vet ./...` is clean.
- No blockers for the next plan in this phase.

## Debt Registered (not a new *-DEBT.md row -- recorded here per plan instruction; a future plan may promote these to PHASE-14-DEBT.md rows)

- **87 rows across `02-DEBT.md`..`PHASE-10-DEBT.md` remain on `debtRegisterGradeWitnessExemptions`**, frozen prior art without Grade/Witness cells. Migrating them with real, honest, executed probes is a substantial standalone effort (this plan's own migration of 32 rows across four files, with only 6 needing real probes, was already this plan's largest single cost). Trigger: a future plan or QLT item decides the remaining historical corpus is worth the same treatment.
- **The suppression enumerator's live-module scan is narrower than D-14-25's literal text** (PENDING-only, not the full grammar, for comments/string-literals). Trigger: a future plan builds a proper disambiguation layer — a full `*-CONTEXT.md` decision-ID index across all archived milestones, plus a second closed registry distinguishing this plan's `escape:` witness-grammar namespace from the pre-existing `escape:`-prefixed anti-theater vocabulary — at which point the full grammar can be applied module-wide without the 115-false-positive collision.

---
*Phase: 14-evidence-instrument-and-honest-scoping*
*Completed: 2026-09-18*

## Self-Check: PASSED

- FOUND: `internal/compiler/session/witness_registry_test.go`
- FOUND: `.planning/UNREACHABLE-CLAIMS.md`
- FOUND: `testdata/phase14/blame_unreachable_admission_refusal.lang`
- FOUND: `testdata/phase14/multi_function_match_refusal.lang`
- FOUND: commit `689fde3` (test, RED, Task 1)
- FOUND: commit `72ecb6f` (feat, GREEN, Task 1+2)
- FOUND: commit `17ff8d1` (test, RED, Task 3)
- FOUND: commit `f1651e0` (feat, GREEN, Task 3)
- FOUND: commit `b8242d1` (feat, GREEN, Task 4)
- `go test ./internal/compiler/session/... -run 'TestDebtRegisterWitnessGrammarIsClosed|TestB1BlameIsStructurallyUnreachable|TestD1243ControlIsUnconstructible|TestPhase6HeldoutPairsAreAlphaRenamesOnly|TestLTOInertnessOnMultiFunctionEmission|TestRetainedPointerEscapeIsStillUnsubjected|TestNoSuppressionOutlivesItsWitness|TestSuppressionWitnessGuardIsNotInert|TestUnreachableClaimsViewIsCurrent' -count=1 -v` prints only `--- PASS:` lines
- `go test ./...` reports `ok` for all 25 packages
- `go vet ./...` is clean
- `git status --short` clean except the pre-existing untracked `.planning/milestone.lock`
