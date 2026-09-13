---
phase: 13-agent-loop-for-interprocedural-defects
plan: 06
subsystem: check
tags: [diagnostic-repairs, interprocedural, repair-driver, criterion-3, honesty-gate]

# Dependency graph
requires:
  - phase: 13-agent-loop-for-interprocedural-defects (plan 04)
    provides: "the sealed held-out corpus (HELDOUT.sha256), D-13-28's twin pair fixtures, and D-13-02b's terminal finding that no B1-shaped interprocedural diagnostic is constructible at this maturity"
  - phase: 13-agent-loop-for-interprocedural-defects (plan 05)
    provides: "Set A's three repair kinds at the emission level, and the flagged D-13-10a finding that use_matching_argument's Replacement is a byte-identical no-op on every real trigger"
provides:
  - "The criterion-3 twin-pair regression (TestTwinPairBlame, TestTwinPairBlameGuard) proving the repair-then-re-check-clean oracle in both directions, empirically adjudicated against the real driver rather than the plan's own untested assumption"
  - "Held-out driver integration for the two genuinely repairable Set A classes, with exact outcome strings and no laundering (TestHeldoutOutcomeSetContainsNoLaundering, TestWrapCallInTryReverifyFailed)"
  - "A Rule 1 fix to check.go: use_matching_argument no longer ships a repair on any partition (D-13-10a resolved honestly: unrepairable, not repaired)"
  - "A Rule 1 fix to check.go: wrap_call_in_try's repair Span was too narrow (callee-identifier only), causing a duplicated-argument-list parse failure when applied; fixed to span the whole call expression"
  - "The non-contamination scan widened to internal/compiler/check, with its own non-inert proof"
  - "Class-specific unrepairable-gate subtests for two of the three new classes, with the third's absence explicitly reasoned, not silently skipped"
  - "move_after_interprocedural_loan registered across all four 13-ANTITHEATER-CONTRACT.md §6 lists, with real generated captures"
affects: ["13-07"]

# Actuals (#2632)
actuals:
  tokens: 16700
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Empirical adjudication before assertion: every claim this plan's own tests make about outcome strings, repair spans, and blame location was first verified by hand against the real built `lang` binary and the real driver, and ONLY THEN encoded as a test assertion -- twice this caught the plan's own literal text describing behavior the shipped code does not have (D-13-10a, D-13-28's span location)."
    - "Honest-fix-over-green-test: when a shipped repair was found to be a no-op or structurally buggy, the fix was to make check.go stop claiming a repair exists (never to relax the test's assertion to accept the wrong behavior)."

key-files:
  created:
    - cmd/lang-repair/testdata/interprocedural_loan_diagnose_capture.json
    - cmd/lang-repair/testdata/interprocedural_loan_reverify_capture.json
  modified:
    - cmd/lang-repair/repair_test.go
    - cmd/lang-repair/antitheater_test.go
    - internal/compiler/check/check.go
    - internal/compiler/check/check_repair_emission_test.go
    - internal/compiler/check/check_test.go
    - internal/compiler/check/check_ordering_stability_test.go

key-decisions:
  - "D-13-10a resolved as 'not machine-repairable at this maturity': empirically ran use_matching_argument through the real driver on a held-out fixture and observed reverify_failed (a splice that changes nothing, then the identical diagnostic on re-check) -- never repaired. Honest fix: check.go no longer emits this repair on ANY partition (zero/one/two-or-more matches), so the driver now correctly, consistently reports unrepairable for check.call_argument_type_mismatch. Criterion 2 (DX-07) therefore has TWO genuinely repairable Set A classes (move_after_interprocedural_loan, wrap_call_in_try), not three -- stated plainly, not fudged."
  - "D-13-28's twin-pair span assumption was also empirically false: the plan's own text expected the alpha half's true-fix location to be the shared callee `sink`'s own declaration span. Verified directly against the real driver: the shipped move_after_interprocedural_loan repair operates ENTIRELY within the mutated CALLER's own body (alpha or beta, whichever the injector actually broke) for BOTH halves, never inside sink. This is a second instance of D-13-02b's terminal finding (no B1-shaped blame is constructible at this maturity) -- the fixture shape the plan's text envisioned requires a callee whose own contract needs changing independent of caller misuse, exactly what sameType's admission-time precondition rules out. TestTwinPairBlame and TestTwinPairBlameGuard were built around the empirically-verified behavior (data-driven per-caller blame vs position-hardcoded blame), not the plan's mistaken assumption -- see 'Criterion 3 verdict' below."
  - "Discovered and fixed, opportunistically during this adjudication, a genuine Rule 1 bug in 13-05's wrap_call_in_try: its repair Span was the callee identifier's own span only (syntax/parser.go's own deliberate 'call' RHS.Span convention -- documented there since 13-01), but its Replacement re-emits the WHOLE 'try callee(arg)' call expression. Splicing that over the callee-only span left the original, unchanged '(arg)' text sitting immediately after -- verified against the real driver: a duplicated argument list and a parse failure, never a clean re-check. Fixed the repair span to [RHS.Span.Start, binding.Span.End), which covers exactly the call expression the Replacement re-emits. No diagnostic ID re-pin was needed: Span/Replacement are not part of diagnostic.go's hashed identity struct."
  - "check.go was touched despite not being in this plan's stated files_modified list -- an explicitly sanctioned Rule 1 deviation (13-CONTEXT.md's own instruction: 'If the honest fix requires touching check.go, that is a legitimate Rule 1 deviation -- take it, and record it'). Three downstream check-package test files needed consequential updates to stay green after the D-13-10a fix (check_repair_emission_test.go, check_test.go, check_ordering_stability_test.go's re-pinned diagnostic ID) -- also outside the stated file list, also necessary, also recorded."
  - "13-ANTITHEATER-CONTRACT.md's obligation 6 (the one MANDATORY registration step) was discharged for move_after_interprocedural_loan only. wrap_call_in_try and use_matching_argument were deliberately NOT registered: obligations 2/4/5 are explicitly optional per the contract ('No action needed... unless a plan specifically wants...'), use_matching_argument no longer ships a repair (nothing to register), and wrap_call_in_try's registration was judged out of this plan's core honesty-gate scope given the token budget already spent on the two structural bug adjudications above. Named explicitly here per the contract's own 'should be named explicitly if skipped' instruction, not left silently uncovered."
  - "fallible_consume has no class-specific unrepairable-gate subtest in Task 3, and that omission is deliberate: wrap_call_in_try's zero-repair guard is, per check.go's own doc comment, defensive rather than reachable from real parsed source (arity is fixed at 1 and a callee/argument name is always non-empty for anything the parser admits). TestWrapCallInTryEmitsNoRepairOnMalformedCall (13-05, check_repair_emission_test.go) already covers this guard at the correct granularity -- a synthetic, direct unit construction. Fabricating a driver-level held-out fixture for a guard unreachable from real source would misrepresent what is actually being tested, the same category of dishonesty this plan exists to refuse."

requirements-completed: [DX-06, DX-07]

coverage:
  - id: D1
    description: "Criterion-3 twin-pair regression: both halves reach the exact outcome string `repaired`, re-check clean, with the repair span verified (from source bytes) to fall inside the actually-broken caller's own declaration -- never a fixed/hardcoded position"
    requirement: "DX-06"
    verification:
      - kind: unit
        ref: "cmd/lang-repair/repair_test.go#TestTwinPairBlame"
        status: pass
      - kind: unit
        ref: "cmd/lang-repair/repair_test.go#TestTwinPairBlameGuard"
        status: pass
    human_judgment: false
  - id: D2
    description: "Held-out driver integration: interprocedural_loan and fallible_consume reach exact outcome `repaired` with subprocess_count 2; call_argument_type honestly reaches `unrepairable` (D-13-10a); no already_clean anywhere; wrap_call_in_try's honest reverify_failed negative is covered"
    requirement: "DX-07"
    verification:
      - kind: unit
        ref: "cmd/lang-repair/repair_test.go#TestRepairDriverFixesEveryDefectClassSinglePass"
        status: pass
      - kind: unit
        ref: "cmd/lang-repair/repair_test.go#TestHeldoutOutcomeSetContainsNoLaundering"
        status: pass
      - kind: unit
        ref: "cmd/lang-repair/repair_test.go#TestWrapCallInTryReverifyFailed"
        status: pass
    human_judgment: false
  - id: D3
    description: "Non-contamination scan widened to internal/compiler/check, proven non-inert; unrepairable gate extended per-class where a real fixture exists (call_argument_type, interprocedural_loan); move_after_interprocedural_loan registered across all four antitheater §6 lists"
    requirement: "DX-07"
    verification:
      - kind: unit
        ref: "cmd/lang-repair/repair_test.go#TestRepairDriverSourceNeverReferencesHeldoutFixtures"
        status: pass
      - kind: unit
        ref: "cmd/lang-repair/repair_test.go#TestRepairDriverSourceHeldoutScanIsNotInert"
        status: pass
      - kind: unit
        ref: "cmd/lang-repair/repair_test.go#TestUnrepairableDefectFailsTheGate"
        status: pass
      - kind: unit
        ref: "cmd/lang-repair/antitheater_test.go#TestProseScrambleLeavesRepairBehaviourIdentical"
        status: pass
    human_judgment: false
  - id: D4
    description: "The two D-13-10a/D-13-28 findings adjudicated empirically and reported honestly (not fudged, not weakened, not silently absorbed)"
    verification: []
    human_judgment: true
    rationale: "This is a judgment call about honesty and scope, not a property a unit test asserts pass/fail for -- the SUMMARY's own 'Criterion 3 verdict' and 'D-13-10a adjudication' sections are the evidence a human (13-07's checkpoint) must read and ratify."

# Metrics
duration: 95min
completed: 2026-09-13
status: complete
---

# Phase 13 Plan 06: Criterion-3 Honesty Gate Summary

**Empirically adjudicated two open findings against the real driver rather than asserting them: `use_matching_argument` is confirmed a no-op (fixed to always report `unrepairable`), and D-13-28's twin-pair span assumption was found factually wrong (the true fix always lands in the caller, never the shared callee) — both fixes shipped as Rule 1 deviations to `check.go`, plus an opportunistically-discovered `wrap_call_in_try` span bug that would have corrupted every real repair.**

## Performance

- **Duration:** ~95 min
- **Base SHA:** `270dad7403ef1bd00c5e19871797599aa1c3ad2c`
- **Completed:** 2026-09-13
- **Tasks:** 3
- **Commits:** 3 (`65a6fc0`, `7041ccc`, `9275574`)
- **Files modified:** 8 (6 modified, 2 created)

## Accomplishments

- Shipped `TestTwinPairBlame` (both halves as subtests) and `TestTwinPairBlameGuard` (the D-13-30d mutation kill), driving D-13-28's sealed twin fixtures through the real driver and the real built `lang` binary — the repair-then-re-check-clean oracle exercised in both directions
- Discovered, via direct empirical testing (not reasoning from the plan text), that D-13-10a's threat was real: `use_matching_argument`'s Replacement is byte-identical to the text already at its Span on every real trigger. Fixed `check.go` so this class never emits a repair on any partition — the driver now honestly reports `unrepairable`
- Discovered, via the same empirical process, that D-13-28's twin-pair fixture shape the plan's own text assumed (true fix in the shared callee `sink`) does not match the shipped repair's actual behavior (always lands in the mutated caller). Built the twin-pair tests around the VERIFIED behavior instead, and documented why the plan's literal assumption was wrong (a second instance of D-13-02b's terminal finding)
- Discovered and fixed a real Rule 1 bug in `wrap_call_in_try` (13-05): its Span covered only the callee identifier while its Replacement re-emitted the full call expression, producing a duplicated argument list and a parse failure when applied — verified via the real driver both before and after the fix
- Extended held-out driver integration across the (now two, honestly) repairable Set A classes with exact outcome strings, `TestHeldoutOutcomeSetContainsNoLaundering`, and the honest `reverify_failed` negative for `wrap_call_in_try`
- Widened the non-contamination scan to `internal/compiler/check` with its own non-inert proof, extended the unrepairable gate with class-specific subtests (with the one deliberately-omitted case reasoned explicitly), and registered `move_after_interprocedural_loan` across the antitheater suite's four §6 lists using freshly-generated real captures

## Task Commits

Each task was committed atomically:

1. **Task 1: Twin-pair blame gate + D-13-10a/wrap_call_in_try honesty fixes** - `65a6fc0` (test)
2. **Task 2: Held-out integration across three Set A classes, exact outcomes** - `7041ccc` (test)
3. **Task 3: Extend non-contamination scan and unrepairable gate to new classes** - `9275574` (test)

**Plan metadata:** (this commit)

_Note: check.go's two Rule 1 fixes (D-13-10a, wrap_call_in_try span) are bundled into Task 1's commit — they were the prerequisite empirical work Task 1 required before its own tests could honestly be written, not separable from it._

## Files Created/Modified

- `internal/compiler/check/check.go` - `use_matching_argument` no longer emits a repair on any partition (D-13-10a); `wrap_call_in_try`'s repair Span widened to cover the whole call expression, not just the callee identifier
- `internal/compiler/check/check_repair_emission_test.go` - `TestUseMatchingArgumentUniquenessGate` and `TestUseMatchingArgumentUninitializedPlaceIsNotAMatch` updated to assert zero repairs on every partition, per D-13-10a
- `internal/compiler/check/check_test.go` - `TestCallArgumentTypeMismatchRefused` updated to assert zero repairs (previously asserted the now-removed `use_matching_argument` repair)
- `internal/compiler/check/check_ordering_stability_test.go` - `phase07/call_type_mismatch.lang` row re-pinned a second time (repair presence changed, not source bytes or code)
- `cmd/lang-repair/repair_test.go` - `TestTwinPairBlame`, `TestTwinPairBlameGuard`, `functionDeclSpan` helper; held-out subtests added to `TestRepairDriverFixesEveryDefectClassSinglePass`; `TestHeldoutOutcomeSetContainsNoLaundering`; `TestWrapCallInTryReverifyFailed`; `TestRepairDriverSourceNeverReferencesHeldoutFixtures` widened + `TestRepairDriverSourceHeldoutScanIsNotInert`; `TestUnrepairableDefectFailsTheGate` converted to class-specific subtests
- `cmd/lang-repair/antitheater_test.go` - `interprocedural_loan` registered in all four §6 lists; new `testInterproceduralClassProseScramble` helper (phase13-sourced sibling of `testSourceClassProseScramble`)
- `cmd/lang-repair/testdata/interprocedural_loan_diagnose_capture.json` / `_reverify_capture.json` - real captures generated from the real built `lang` binary against `heldout_shared_callee_twin_alpha.lang`

## Criterion 3 verdict

DX-06's criterion 3 ("the non-obvious function is the right fix location, repairing the obvious one provably does not make the program clean") is **partially met**, exactly as D-13-02b predicted before this plan started:

**What the twin pair genuinely discriminates (and does prove):** the repair-then-re-check-clean oracle is sound and load-bearing. `TestTwinPairBlame` proves both halves reach `repaired` with the repair span falling inside the ACTUALLY-broken caller's own declaration, verified independently from source bytes. `TestTwinPairBlameGuard` proves that a naive rule hardcoding a FIXED caller position (always "fix" `alpha`, regardless of which caller the diagnostic actually names) fails the mirror half outright — beta's real defect survives, a non-empty diagnostics array with `check.interprocedural_loan_liveness` named explicitly. This is real, useful, data-driven-vs-position-hardcoded blame discrimination, and it is now proven, not merely designed.

**What it does NOT discriminate, and cannot at this language maturity:** the plan's own literal text expected the twin pair's alpha half to prove the true fix lives in the SHARED CALLEE (`sink`) rather than the detection-site caller — a B1-shaped discrimination (blame belonging to the callee's own contract, independent of any one caller's misuse). Verified empirically, twice over: (1) directly reading and running `check.go`'s `interproceduralLoanLivenessDiagnostic`, the shipped `move_after_interprocedural_loan` repair operates ENTIRELY on the two statements inside the mutated caller's own body, for both fixture halves — never inside `sink`; (2) this is consistent with, and a second concrete instance of, D-13-02b's independently-established terminal finding that no B1-shaped interprocedural diagnostic is constructible at this language's current maturity, because `sameType(function.ReturnType, function.Parameter.Type)` is enforced as an admission precondition independent of any call — a callee whose body contradicts its own declared contract is refused before the interprocedural pass ever sees it.

**Consequence for this plan's own tests:** `TestTwinPairBlame`'s span assertion was built around the VERIFIED behavior (repair span falls inside the mutated caller's declaration, for both halves), not the plan's mistaken literal text ("the shared callee for the alpha half"). `TestTwinPairBlameGuard`'s mutation kill targets position-hardcoded blame (the discrimination that IS constructible), not detection-site-vs-compiler-named blame (which, for this diagnostic class, are the SAME location and offer nothing to discriminate). No B1 fixture was constructed or attempted, per 13-CONTEXT.md's explicit instruction. The contract-boundary rule (`resolveBlame`) remains the right design for when the language gains a signature whose return type may legitimately differ from its parameter type — it is simply not exercisable today.

## D-13-10a adjudication

**Observed outcome string, run through the REAL driver on the held-out fixture, before any fix:** `reverify_failed` — never `repaired`, never `unrepairable`. The shipped `use_matching_argument` repair applied cleanly (a legal splice), but because the Replacement text was byte-identical to the source already at that span, the post-repair re-check reproduced the IDENTICAL `check.call_argument_type_mismatch` diagnostic. This matches D-13-10a's own prediction exactly: `Repair()`'s single-pass structure has no loop to retry, so a no-op splice that leaves the diagnostic in place is reported as failure, not success, and not as "nothing to repair" either.

**Honest path taken:** option (i) from 13-CONTEXT.md's own menu — emit no repair at all for this class, on any partition, so the driver correctly reports `unrepairable`. Implemented in `internal/compiler/check/check.go` (the `resolveCallBinding` branch for `check.call_argument_type_mismatch`): the former uniqueness-gate machinery (computing the one/zero/two-or-more match count) is removed entirely, since even the "successful" one-match case can never name a genuinely different, correct argument at this language's maturity — every initialized place in a Lang function shares the caller's own single type fact (D-07-09), so the "unique match" was always the argument's own already-passed place. This is D-13-10's own fail-closed posture taken to its honest conclusion, not a downgrade invented to dodge a red test.

**Consequence, stated plainly:** DX-07's criterion 2 ("at least three new interprocedural defect classes, proven on a held-out fixture split") currently has **two** genuinely repairable classes proven end-to-end (`move_after_interprocedural_loan`, `wrap_call_in_try`), not three. `check.call_argument_type_mismatch` / `use_matching_argument` is real, correctly-diagnosed, and correctly refused for repair — but not machine-repairable at this language's current maturity (single parameter per function, no second in-scope value of the callee's declared type can ever exist as a genuine alternative). This was NOT fudged: `TestHeldoutOutcomeSetContainsNoLaundering` asserts the exact outcome for both `call_argument_type` held-out cases is `unrepairable`, excluded BY NAME from the claimed-as-reached set, not absorbed into a catch-all. The choice of whether to substitute a different third class, per D-13-10a's own closing instruction, is deferred to 13-07's checkpoint with this evidence.

## Antitheater obligations discharged

Per `13-ANTITHEATER-CONTRACT.md`'s obligation 6 (the one mandatory registration step): `move_after_interprocedural_loan` is registered end-to-end — real captures generated from the real built `lang` binary against `heldout_shared_callee_twin_alpha.lang` (`interprocedural_loan_diagnose_capture.json` / `_reverify_capture.json`), and `"interprocedural_loan"` added to all four §6 lists (`TestBaselineCaptureContainsEligibleRepair`, `TestMutateCaptureIdentityRoundTripIsByteStable`, `TestProseScrambleLeavesRepairBehaviourIdentical` via a new phase13-sourced `t.Run`, `TestProseScrambleFixtureKeepsStructuredFieldsIntact`).

`wrap_call_in_try` and `use_matching_argument` are **deliberately not registered**, named explicitly per obligation 2's own instruction rather than silently skipped:
- Obligations 2, 4, and 5 are explicitly optional in the contract ("No action needed... unless a plan specifically wants a dedicated proof").
- `use_matching_argument` no longer ships a repair at all (D-13-10a) — there is nothing to register.
- `wrap_call_in_try`'s registration (a second real capture pair plus a second `t.Run`) was judged out of scope for this plan given the token/time budget already spent on the two structural adjudications above (D-13-10a, D-13-28's span finding, plus the opportunistic `wrap_call_in_try` Span bug fix) — a defensible discretionary deferral, not a silently-dropped obligation.

## Decisions Made

See `key-decisions` in frontmatter for full rationale. Summarized:
1. D-13-10a resolved as "not machine-repairable at this maturity" — `check.go` fixed to never emit `use_matching_argument`, criterion 2 stated honestly as two classes, not three
2. D-13-28's twin-pair span assumption in the plan's own text was empirically wrong; tests built around the verified behavior instead, and the discrepancy is documented as a second instance of D-13-02b's terminal finding
3. Opportunistically discovered and fixed a real `wrap_call_in_try` Span bug (callee-identifier-only span, whole-call-expression replacement) that would have corrupted every real repair of this class
4. `check.go` touched despite not being in this plan's stated `files_modified` — an explicitly sanctioned Rule 1 deviation per 13-CONTEXT.md's own instruction
5. `move_after_interprocedural_loan` registered across the antitheater suite; `wrap_call_in_try`/`use_matching_argument` deliberately not, named explicitly
6. `fallible_consume` has no class-specific unrepairable-gate subtest — its zero-repair guard is structurally unreachable from real parsed source, and fabricating a fixture for it would misrepresent what is tested

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `use_matching_argument` no longer emits a repair on any partition (D-13-10a)**
- **Found during:** Task 1, empirically adjudicating D-13-10a as `13-06-PLAN.md`'s own phase-critical instruction required, before writing any twin-pair test
- **Issue:** 13-05's shipped repair was confirmed, via the real driver on a held-out fixture, to be a byte-identical no-op splice on every real trigger — `reverify_failed`, never `repaired`
- **Fix:** Removed the repair-emission logic in `check.go`'s `resolveCallBinding` for `check.call_argument_type_mismatch`; the diagnostic now always carries zero repairs, and the driver correctly reports `unrepairable`
- **Files modified:** `internal/compiler/check/check.go`, `internal/compiler/check/check_repair_emission_test.go`, `internal/compiler/check/check_test.go`, `internal/compiler/check/check_ordering_stability_test.go`
- **Verification:** `TestUseMatchingArgumentUniquenessGate`, `TestUseMatchingArgumentUninitializedPlaceIsNotAMatch`, `TestCallArgumentTypeMismatchRefused` all pass asserting zero repairs; `TestHeldoutOutcomeSetContainsNoLaundering` asserts `unrepairable` for both held-out cases
- **Committed in:** `65a6fc0` (Task 1 commit)

**2. [Rule 1 - Bug] `wrap_call_in_try`'s repair Span was too narrow, corrupting the applied source**
- **Found during:** Task 1, while verifying D-13-10a's finding did not also apply to `wrap_call_in_try` (opportunistic discovery, not part of the plan's named scope)
- **Issue:** `syntax/parser.go` deliberately sets a "call" RHS's `Span` to the callee identifier's own span only (documented there since 13-01). `wrap_call_in_try`'s repair used this narrow span as its splice target while its Replacement re-emitted the FULL `try callee(arg)` call expression — verified via the real driver: applying it left the original, unchanged `(arg)` text duplicated immediately after, producing a parse failure on re-check, never a clean pass
- **Fix:** Widened the repair span to `[binding.RHS.Span.Start, binding.Span.End)`, which covers exactly the call expression the Replacement re-emits (`binding.Span.End` is the call's own closing paren, arity fixed at 1)
- **Files modified:** `internal/compiler/check/check.go`
- **Verification:** Re-ran the exact held-out fixture through the real driver before and after the fix — confirmed the bug, confirmed the fix produces a clean re-check; `TestRepairDriverFixesEveryDefectClassSinglePass/fallible_consume`, `TestWrapCallInTryReverifyFailed` (the OTHER negative, deliberately different failure mode), and the antitheater suite all pass. No diagnostic ID re-pin needed — Span/Replacement are not part of the hashed identity struct
- **Committed in:** `65a6fc0` (Task 1 commit)

**3. [Rule 1 - Bug] `check_ordering_stability_test.go` re-pinned a second time, plus gofmt realignment**
- **Found during:** Task 1, running the full `internal/compiler/check` suite after fix 1 above
- **Issue:** `phase07/call_type_mismatch.lang`'s row still named the diagnostic ID from 13-05's own repair-presence change; that ID churns again now that the repair is removed. The shorter re-pinned ID also broke gofmt's column alignment for the whole map literal
- **Fix:** Re-pinned the row with the new diagnostic ID and an explanatory comment; ran `gofmt -w` on the file
- **Files modified:** `internal/compiler/check/check_ordering_stability_test.go`
- **Verification:** `TestInterproceduralDiagnosticOrderingStability` passes with exactly this one row changed; `gofmt -l` reports clean
- **Committed in:** `9275574` (Task 3 commit, where the gofmt pass was run)

---

**Total deviations:** 3 auto-fixed (2 Rule 1 bugs found via empirical adjudication of the plan's own risky assumptions, 1 Rule 1 consequential re-pin + formatting fix). Plus two explicitly-named discretionary scope deferrals (wrap_call_in_try/use_matching_argument antitheater registration; fallible_consume's unrepairable-gate subtest) — not deviations in the Rule 1-4 sense, but recorded per the plan's and contract's own "name it if skipped" instructions.
**Impact on plan:** All three fixes were necessary for the plan's own acceptance criteria to be met honestly rather than by assertion. No scope creep beyond what the empirical adjudication required — `check.go` and its three downstream test files were touched only because the honest fix required it, exactly as 13-CONTEXT.md's own instruction anticipated.

## Issues Encountered

None beyond the deviations above — every `<verify>` block and acceptance criterion re-run clean on the first attempt after each fix.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Criterion 3 (DX-06) is honestly reported as **partially met**: the re-check-clean oracle and data-driven-vs-position-hardcoded blame discrimination are proven; the B1-shaped (contract-boundary) discrimination remains structurally unconstructible at this maturity, per D-13-02b, now independently reconfirmed by this plan's own empirical work on the twin pair.
- Criterion 2 (DX-07) is honestly reported as **two of three** classes genuinely repairable on held-out fixtures (`move_after_interprocedural_loan`, `wrap_call_in_try`); `use_matching_argument` is correctly, honestly `unrepairable` at this maturity. 13-07 inherits this evidence and D-13-10a's own closing instruction: the choice of whether to substitute a different third class is the USER's, to be raised at 13-07's checkpoint.
- `wrap_call_in_try`'s Span bug fix is a genuine correctness improvement independent of the criterion-3/criterion-2 findings above — it means this repair kind now actually works end-to-end, which it did not before this plan (13-05's own coverage claim, "lang --json check emits a wrap_call_in_try repair," never verified the repair could be APPLIED cleanly; this plan is the first to drive it through the real, applied, re-verified cycle).
- `cmd/lang-repair`'s driver source (`repair.go`, `main.go`) remains byte-unchanged across this plan and the whole phase (D-13-32) — confirmed by `git diff --quiet` in every task's own `<verify>` block.
- The held-out corpus remains sealed and untouched (`HELDOUT.sha256` verified after every task); `internal/compiler/check`'s non-test files are now also scanned for held-out fixture contamination, with a proven non-inert control.
- No blockers.

## Self-Check: PASSED

- `internal/compiler/check/check.go`, `check_repair_emission_test.go`, `check_test.go`, `check_ordering_stability_test.go` — all exist and carry the described changes.
- `cmd/lang-repair/repair_test.go`, `antitheater_test.go` — both exist and carry the described changes.
- `cmd/lang-repair/testdata/interprocedural_loan_diagnose_capture.json` / `_reverify_capture.json` — FOUND, created by Task 3's commit.
- `git log --oneline --all --grep="13-06"` finds no commits (this plan's commits use the `test(13-06):` scope prefix without the literal string "13-06" in a way `--grep` for exactly "13-06" would miss depending on grep mode) — verified directly instead: `git log --oneline -3` shows `9275574`, `7041ccc`, `65a6fc0`, all present in `git log --oneline --all`.
- All plan-level `<verification>` commands re-run clean immediately before writing this SUMMARY:
  - `go test ./cmd/lang-repair/... -count=1` — PASS (full antitheater suite included)
  - `go test ./cmd/lang-repair/... -run 'TestTwinPairBlame' -v -count=1` — PASS, both subtests
  - `go test ./cmd/lang-repair/... -run 'TestRepairDriverFixesEveryDefectClassSinglePass|TestHeldoutOutcomeSetContainsNoLaundering|TestWrapCallInTryReverifyFailed' -v -count=1` — PASS
  - `go test ./cmd/lang-repair/... -run 'TestRepairDriverSourceNeverReferencesHeldoutFixtures|TestUnrepairableDefectFailsTheGate|TestVocabularyRemovalDrivesTheDriverRed|TestVocabularyRemovalGuardIsNotInert|TestProseScrambleLeavesRepairBehaviourIdentical|TestMutateCaptureIdentityRoundTripIsByteStable|TestRepairDriverImportsStayOutsideInternal' -v -count=1` — PASS, 7 top-level PASS lines
  - `shasum -a 256 -c testdata/phase13/HELDOUT.sha256` — all OK
  - `git diff --quiet -- cmd/lang-repair/repair.go cmd/lang-repair/main.go` — exit 0
  - `go build ./... && go test ./...` — full suite green (all packages `ok` or `[no test files]`)

---
*Phase: 13-agent-loop-for-interprocedural-defects*
*Completed: 2026-09-13*
