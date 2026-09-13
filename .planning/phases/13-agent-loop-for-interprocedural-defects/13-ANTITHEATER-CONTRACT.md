# 13-01 Task 2: What `cmd/lang-repair/antitheater_test.go` Demands of a New Defect Class

Read in full (all 794 lines). RESEARCH.md Assumption A4 / Open Question 2 named
this file as the research pass's single LOW-confidence area; this document
closes that gap before any new defect class (`wrap_call_in_try`,
`use_matching_argument`, or the cycle-explainability showcase) is designed.

## 1. Capture-corpus files: which must exist per class, and their naming convention

Every class the anti-theater guards exercise needs exactly two checked-in
captures under `cmd/lang-repair/testdata/`, named `<class>_diagnose_capture.json`
and `<class>_reverify_capture.json` — read via `captureFile(t, name)`
(`antitheater_test.go:192-199`, `filepath.Join("testdata", name)`). The three
classes iterated today are `match`, `move`, `borrow`
(`antitheater_test.go:206`, `antitheater_test.go:434`, `antitheater_test.go:779`
all range over the literal slice `[]string{"match", "move", "borrow"}`). A
capture is produced by running the REAL shipped `lang` binary against the
injector's real mutated `heldout_*.lang` fixture — never fabricated by hand
(`antitheater_test.go:188-191`, citing "this plan's Task 1" of the ORIGINAL
06-13/06-14 phase, not this plan). `diagnose_capture.json` is `lang --json
check`'s own stdout against the mutated source; `reverify_capture.json` is the
same command's stdout against the POST-repair source. There is no companion
`heldout_`/`derivation_` distinction inside `cmd/lang-repair/testdata/` itself
— that split lives in `testdata/phase6/` (and now `testdata/phase13/`), one
layer up; the captures directory only ever holds the two JSON documents per
class name.

A NEW interprocedural class (`move_after_interprocedural_loan`,
`wrap_call_in_try`, `use_matching_argument`) that wants coverage from
`TestProseScrambleLeavesRepairBehaviourIdentical` or
`TestVocabularyRemovalDrivesTheDriverRed` needs both files staged under a new
`<class>_diagnose_capture.json` / `<class>_reverify_capture.json` name, AND a
new entry added to whichever hardcoded class list currently drives that guard
(see obligation 6 below — this is NOT automatic).

## 2. What `TestFixtureSubstitutionIsFaithful` requires of a new capture

`TestFixtureSubstitutionIsFaithful` (`antitheater_test.go:221-270`) does not
read the checked-in captures at all — it is the TRACER proving the
fixture-substitution mechanism itself is faithful, generated fresh every run:
it builds the real `lang` binary (`testsupport.BuildCLI`, line 230), runs the
real driver against a real injected `move` defect
(`antitheater_test.go:233-245`), independently captures that SAME real
binary's own diagnose/reverify documents for the exact mutated/repaired byte
pair it just observed (`antitheater_test.go:247-254`), stages them for the
stand-in keyed by content hash (`writeCaptureForContent`,
`antitheater_test.go:141-148`), and asserts the stand-in-driven run is
byte-identical to the real-binary-driven run (`antitheater_test.go:264-269`).
A new class inherits this test's coverage for free ONLY if it is added to
this function's own hardcoded fixture/injector pair — today hardcoded to
`heldout_move_defect.lang` / `session.MoveInjector{}` (line 233-234). It is
not parameterized over class the way the guard tests below are; extending it
to cover an interprocedural class means adding a second, source-varying
`t.Run` (or a loop), not merely dropping in a new capture pair.

## 3. What `TestMutateCaptureIdentityRoundTripIsByteStable` implies for a capture carrying a `repairs` array

`TestMutateCaptureIdentityRoundTripIsByteStable`
(`antitheater_test.go:433-448`) proves `mutateCapture(raw, "identity")` is
deterministic (two calls on the same input produce byte-identical output) and
idempotent (re-applying identity to its own output changes nothing) — a
precondition every scramble/strip mode's "compare against baseline" step
depends on, so an observed difference is attributable to the mutation, never
to `encoding/json`'s own re-marshal nondeterminism. It iterates the same
`{"match", "move", "borrow"}` list (line 434). A capture whose `repairs`
array is non-empty (any of the three D-13-09 classes, once captured) goes
through the exact same generic `map[string]interface{}` decode/marshal path
(`mutateCapture`, `antitheater_test.go:292-321`) as every other field —
nothing in this test or `mutateCapture` special-cases `repairs[]` structurally,
so a `repairs`-bearing capture requires no new handling here. The implication
is purely additive: adding a new class to the iterated list (obligation 6)
extends this test's own determinism proof to that class's capture for free,
PROVIDED the new class's capture is also added to whatever list drives this
specific test.

## 4. What the prose-scramble tests mean for the new repair-kind strings

`TestProseScrambleLeavesRepairBehaviourIdentical`
(`antitheater_test.go:625-649`) and
`TestProseScrambleFixtureKeepsStructuredFieldsIntact`
(`antitheater_test.go:772-794`) both scramble ONLY `message`/`detail` string
values (`scrambleProse`, `antitheater_test.go:360-377`) and assert every
OTHER field — explicitly including `repairs[].kind`
(`mutateCapture`'s `"scramble_prose"` case at line 306-307 calls
`scrambleProse`, which recurses into every map/slice but only ever rewrites a
`"message"`/`"detail"` key, per the switch at lines 362-369) — is byte-for-byte
untouched. `Repair.Kind` is neither `"message"` nor `"detail"`, so
`move_after_interprocedural_loan`, `wrap_call_in_try`, and
`use_matching_argument` are structurally IMMUNE to prose scrambling by
construction — nothing in these tests treats a repair kind string as prose,
and nothing needs to change in `scrambleProse` for a new kind to be safe.
The only actual dependency a new class has on these tests is procedural, not
mechanical: it must be added to the iterated class list (obligation 6) to be
COVERED by the guard at all; until then the guard proves nothing about the
new class, but it also does not need to — it proves nothing FALSE about it
either.

## 5. What `TestVocabularyRemovalDrivesTheDriverRed` / `TestVocabularyRemovalGuardIsNotInert` require when the repair-kind vocabulary grows

`TestVocabularyRemovalDrivesTheDriverRed`
(`antitheater_test.go:734-741`) strips one of `repairs[]` wholesale, or one
of `kind`/`span`/`replacement` from within each repair
(`mutateCapture`'s `"strip_kind"`/`"strip_span"`/`"strip_replacement"` cases,
`antitheater_test.go:312-316`, which delete the named field from every
repair object via `forEachRepair`, `antitheater_test.go:341-353`) and asserts
the driver goes RED (nonzero exit, `OutcomeUnrepairable`, source
byte-unmodified — `testVocabularyRemovalMode`,
`antitheater_test.go:705-727`). Its companion,
`TestVocabularyRemovalGuardIsNotInert` (`antitheater_test.go:750-770`), is
the load-bearing control proving the SAME harness goes GREEN with nothing
stripped — without it, a broken harness could make every RED assertion above
pass for the wrong reason. Both are hardcoded to the `move` class only
(`heldout_move_defect.lang` / `session.MoveInjector{}`, lines 707-708 and
752-753) — unlike the prose-scramble guard, THIS guard is not even
parameterized over the three existing classes, let alone a new one. Growing
the repair-kind vocabulary with three new strings requires NO change here
UNLESS a plan wants dedicated stripped-vocabulary coverage for an
interprocedural class specifically — the guard's structural claim ("the
driver depends only on the structured channel, never prose it could scrape")
is generic over Kind's exact string value, so it already covers any future
kind by construction; it just is not currently EXERCISED against one.

## 6. Where a new defect class must be registered

Four separate hardcoded points gate whether a new class's coverage is real
rather than merely possible:

- `TestBaselineCaptureContainsEligibleRepair` (`antitheater_test.go:205-219`)
  — iterates `{"match", "move", "borrow"}`.
- `TestMutateCaptureIdentityRoundTripIsByteStable`
  (`antitheater_test.go:433-448`) — same three-class list.
- `TestProseScrambleLeavesRepairBehaviourIdentical`
  (`antitheater_test.go:631-649`) — three `t.Run` calls plus two structurally
  distinct classes (`cleanup`, `stale_evidence`) that bypass the JSON
  protocol entirely; NOT a data-driven loop.
- `TestProseScrambleFixtureKeepsStructuredFieldsIntact`
  (`antitheater_test.go:779-794`) — same three-class list, nested with
  `{"diagnose", "reverify"}`.

None of these four lists is a single shared constant — each function declares
its own literal `[]string{"match", "move", "borrow"}` (or, for
`TestProseScrambleLeavesRepairBehaviourIdentical`, its own five `t.Run`
blocks). A new interprocedural class registers coverage only by being added
to EVERY list whose guard it wants to inherit; there is no single
registration point, and `repair_test.go`'s own
`TestRepairDriverFixesEveryDefectClassSinglePass` (`repair_test.go:430-447`)
is a FIFTH, entirely separate hardcoded list that must also be extended
(covered by this plan's own Task 3, not by anything in
`antitheater_test.go`).

## Obligations for plans 13-03, 13-05 and 13-06

1. **Capture generation (obligation 1).** Whichever plan first exercises
   `wrap_call_in_try` or `use_matching_argument` end-to-end must generate
   `<class>_diagnose_capture.json` / `<class>_reverify_capture.json` under
   `cmd/lang-repair/testdata/` from a REAL `lang --json check` run against a
   REAL fixture and its REAL post-repair state — never hand-written JSON
   (the same rule 06-13/06-14 already established, restated here for the
   avoidance of doubt).
2. **`TestFixtureSubstitutionIsFaithful` extension is optional but should be
   named explicitly if skipped.** This is a single hardcoded tracer, not a
   loop; extending it to a second class is extra work with a real cost
   (another `t.Run`, another real-binary round trip) that no phase task
   currently names. Whichever plan first ships `move_after_interprocedural_loan`,
   `wrap_call_in_try`, or `use_matching_argument` end-to-end should decide,
   and record the decision, rather than silently leaving it uncovered.
3. **No action needed for `repairs`-array determinism** (obligation 3) beyond
   registering the new class per obligation 6 — the generic JSON decode path
   already handles a `repairs` array of any shape.
4. **No action needed for prose-scramble safety** (obligation 4) — the new
   Kind strings are immune by construction. Registering per obligation 6 is
   what makes that immunity OBSERVED rather than merely true.
5. **No action needed for vocabulary-removal coverage** (obligation 5) unless
   a plan specifically wants a dedicated RED-on-strip proof for an
   interprocedural class; the existing guard's claim already generalizes.
6. **Registration (obligation 6) is the one MANDATORY step.** Whichever plan
   ships a new class's first real, driver-verified repair must add that
   class's name to every one of the four `antitheater_test.go` lists in
   §6 it wants coverage from, PLUS `repair_test.go:430-447`'s own
   `TestRepairDriverFixesEveryDefectClassSinglePass` list (Task 3 of THIS
   plan is the first instance of this obligation, for
   `move_after_interprocedural_loan` — see its own read_first pointing back
   at this document).
7. **`import_boundary_test.go`'s `forbiddenImportSubstr = "/internal/"` lint**
   (`import_boundary_test.go:17`) applies to every new non-test file under
   `cmd/lang-repair` a future plan might add — none is anticipated, but any
   helper needed from `internal/` must be re-declared locally in a `_test.go`
   file (which the lint exempts by construction, `scanForbiddenImports`
   skipping `_test.go` suffixes at `import_boundary_test.go:32`), never
   imported into `repair.go`/`main.go`.

## Contradictions found

None. No obligation extracted above contradicts anything in
13-VALIDATION.md's Per-Task Verification Map — the antitheater file's own
structural claims (repair-kind values are immune to prose scrambling;
structured-vocabulary removal is what actually gates the driver) are
consistent with, and in fact are the SOURCE of, D-13-11's and D-13-13's own
posture toward the new repair kind (a Kind string is identity-bearing
metadata, never prose the driver scrapes).
