---
phase: 14-evidence-instrument-and-honest-scoping
plan: 08
subsystem: process
tags: [go, nat03-mutations, axis-movement-law, escape-registry, suppression-enumerator, evd-05]

requires:
  - phase: 14-evidence-instrument-and-honest-scoping
    provides: "plan 14-07's closed four-kind witness grammar (debtRegisterWitnessTokenProblem), the closed escape registry (debtRegisterEscapeRegistry, its callback-invocation-unsubjected entry backed by probe:TestRetainedPointerEscapeIsStillUnsubjected), and the module-wide suppression enumerator (TestNoSuppressionOutlivesItsWitness) that supersedes this plan's deleted single-file marker guard"
provides:
  - "internal/compiler/session/session_phase5_alias.go's AssertMutationMovesAnAxis is now the SOLE axis-movement law: it handles all seven NAT-03 controls including the sanitizer allocator-mismatch control (previously handled only by a thin dispatcher, now deleted)"
  - "internal/compiler/session/session_phase5_alias_test.go's TestEveryMutationMovesItsClaimedAxis enumerates all seven mutation-table rows with zero skips: a Subjected row is asserted through the single law, an unsubjected row is admitted with a declared escape that must resolve in the closed registry"
  - "The superseded single-file marker guard (TestNoNAT03RowRemainsPending) is deleted; every surviving prose mention of the PENDING-05-08 marker string, across production code, test code, and error messages, is gone"
affects: []

actuals:
  tokens: 3517
  tasks: 3
  commits: 4

tech-stack:
  added: []
  patterns:
    - "Declared-state branching in a table-driven test (row.Subjected true/false) replaces a fixture-path exclusion -- every row is admitted into the law, never excluded from it"
    - "Manual RED demonstration for a test-only task: temporarily break real production data (clear a field), run the test, observe the exact expected failure, revert via git checkout, then commit the finished passing test -- used when the task has no separate GREEN implementation step because a prior task already made the data correct"

key-files:
  created: []
  modified:
    - internal/compiler/session/session_phase5.go
    - internal/compiler/session/session_phase5_alias.go
    - internal/compiler/session/session_phase5_test.go
    - internal/compiler/session/session_phase5_alias_test.go
    - internal/compiler/session/witness_registry_test.go

key-decisions:
  - "Task 2 (removing the per-row exclusion) produced a single test commit rather than a RED-then-GREEN commit pair. There was no separate production-code GREEN step: Task 1's collapse already made the real NAT03Mutations() table data satisfy the new declared-state check (the sanitizer row already had Subjected: true and a working law case after Task 1; the callback-invocation row's escape already resolved per plan 14-07's registry). Genuine RED was still demonstrated -- not asserted from memory -- by temporarily clearing the real EscapeID field, running the test, observing the exact expected failure message, then reverting via `git checkout` before committing the finished test. This satisfies the task's own acceptance criterion ('temporarily clearing that row's escape identifier makes the test fail... recorded... before reverting') literally."
  - "Extended the plan's explicit five-prose-site list by two: a negative-assertion literal in this plan's own Task 1 RED test (session_phase5_test.go, checking the error string did NOT contain 'PENDING-05-08') and a descriptive hand-off comment in witness_registry_test.go (plan 14-07's SUMMARY hand-off note, which itself named the marker string it predicted 14-08 would remove). Both are legitimate prose, not production defects, but both are literal occurrences of the retired marker string, and the plan's own must_haves truth ('not as prose in a test file') applies module-wide, not just to the five sites the plan enumerated in its own research. Rule 2 (missing critical): the plan's acceptance criterion is a repo-wide grep with no site list attached, so satisfying it required finding and fixing these two additional occurrences."

patterns-established:
  - "A table-driven guard test with a declared-state branch, calling the shared session-package witness-resolution helper (debtRegisterWitnessTokenProblem) directly rather than duplicating resolution logic, is the pattern for any future per-row admit-with-escape check."

requirements-completed: [EVD-05]

coverage:
  - id: D1
    description: "The mutation axis-movement law has exactly one implementation: Phase5AssertMutationMovesAnAxis (the thin dispatcher) is deleted from session_phase5.go; the sanitizer allocator-mismatch control's case lives directly in AssertMutationMovesAnAxis's own switch in session_phase5_alias.go, delegating to the unmodified assertAllocatorMismatchMovesAxis helper; the default-arm error message names the unsupported control with no stale marker text"
    requirement: "EVD-05"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase5_test.go#TestAssertMutationMovesAnAxisHandlesSanitizerControl"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase5_test.go#TestAssertMutationMovesAnAxisUnknownControlNamesControlNoMarker"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase5_test.go#TestNAT03SanitizerRowMovesItsClaimedAxis"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase5_test.go#TestAssertMutationMovesAnAxisFailsOnAMislabeledAxis"
        status: pass
    human_judgment: false
  - id: D2
    description: "Zero per-row exclusions: TestEveryMutationMovesItsClaimedAxis's fixture-path skip is deleted, replaced by a branch on each row's own declared state. All 7 mutation-table rows produce a subtest, none skipped. The unsubjected callback-invocation row is admitted with its declared escape (escape:callback-invocation-unsubjected), which resolves in the closed escape registry and carries its own probe (probe:TestRetainedPointerEscapeIsStillUnsubjected) -- proven load-bearing by temporarily clearing the row's EscapeID and observing the test fail with the expected message, then reverting"
    requirement: "EVD-05"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase5_alias_test.go#TestEveryMutationMovesItsClaimedAxis"
        status: pass
    human_judgment: false
  - id: D3
    description: "The superseded single-file marker guard (TestNoNAT03RowRemainsPending) is deleted in the same phase that lands its replacement (plan 14-07's TestNoSuppressionOutlivesItsWitness); the replacement demonstrably covers strictly more surface -- it enumerates Skip calls, //go:build constraints, comments, and BasicLit strings across every .go file (production and test), catching the exact two surfaces (the error message, the test-file prose) the deleted guard's single-file line-match missed. No occurrence of the stale PENDING-05-08 marker string survives anywhere under internal/ or cmd/, in any form"
    requirement: "EVD-05"
    verification:
      - kind: unit
        ref: "internal/compiler/session/witness_registry_test.go#TestNoSuppressionOutlivesItsWitness"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/witness_registry_test.go#TestSuppressionWitnessGuardIsNotInert"
        status: pass
      - kind: other
        ref: "grep -rn 'PENDING-05-08' internal cmd (exit 1, zero matches) and grep -rn 'TestNoNAT03RowRemainsPending' internal cmd (exit 1, zero matches)"
        status: pass
    human_judgment: false

duration: 9min (commit-to-commit span, 23:57:13 -> 00:05:55 local, across the plan's 4 task commits)
completed: 2026-09-18
status: complete
---

# Phase 14 Plan 08: Collapse the Axis-Movement Law and Retire Its Superseded Guard Summary

**The mutation axis-movement law has exactly one implementation and zero per-row exclusions: the thin cross-plan dispatcher is deleted, the sanitizer allocator-mismatch control is folded directly into the surviving switch, the callback-invocation row is admitted with a witnessed declared escape instead of a silent skip, and the superseded single-file marker guard is deleted in favor of plan 14-07's module-wide suppression enumerator, which is proven to cover strictly more surface.**

## Performance

- **Duration:** 9 min (commit-to-commit span, 2026-09-17T23:57:13-04:00 -> 2026-09-18T00:05:55-04:00)
- **Started:** 2026-09-17T23:57:13-04:00
- **Completed:** 2026-09-18T00:05:55-04:00
- **Tasks:** 3
- **Files modified:** 5

## Accomplishments

- `Phase5AssertMutationMovesAnAxis`, the thin cross-plan dispatcher in `session_phase5.go` that added one case in front of `AssertMutationMovesAnAxis` and otherwise delegated verbatim, is deleted. The sanitizer allocator-mismatch control's case (`return assertAllocatorMismatchMovesAxis(ctx, mutation)`) now lives directly in `AssertMutationMovesAnAxis`'s own switch in `session_phase5_alias.go` -- there is exactly one axis-movement law.
- The default-arm error message no longer carries the stale `PENDING-05-08` marker text: `AssertMutationMovesAnAxis: unsupported control %q` names the control and nothing else.
- `TestEveryMutationMovesItsClaimedAxis` enumerates all 7 NAT-03 mutation-table rows with zero skips. Each row branches on its own declared state: a `Subjected: true` row is asserted through the single surviving law; the one `Subjected: false` row (`control:native.sanitize.retained_pointer`) is admitted with its declared escape (`escape:callback-invocation-unsubjected`), verified to resolve against the closed escape registry plan 14-07 added.
- `TestNoNAT03RowRemainsPending`, the retired single-file marker guard that read exactly one file for exactly one trimmed comment line (and missed the identical marker surviving in an error message and in two test files -- D-14-22 instance 4), is deleted. Every surviving prose occurrence of the `PENDING-05-08` string -- five sites the plan named plus two additional sites found during execution (a negative-assertion literal in this plan's own new test, and a descriptive comment in `witness_registry_test.go`) -- is rewritten. `grep -rn 'PENDING-05-08' internal cmd` now returns zero matches.
- Re-ran the module-wide suppression guard (plan 14-07's `TestNoSuppressionOutlivesItsWitness`) after the deletion: it passes and its verbose output enumerates 12 suppression sites across the module (`native_lto_test.go` x2, `symbols_test.go`, `session_payload_replay_test.go` x2, `session_phase11_differential_test.go`, `session_phase5_corpus_test.go` x3, `session_phase6_evidence_test.go`, `session_phase6_injectors_test.go` x2) -- none of them the retired marker, confirming the guard covers the surface the deleted single-file guard watched (the alias.go declaration site) plus the two surfaces it missed (the error-message string literal and the test-file prose), by scanning every `.go` file's comments, string literals, and `Skip` calls module-wide rather than one file's comment lines.

## Task Commits

1. **Task 1 RED: two failing falsifiers for the collapse** — `36d835e` (test)
2. **Task 1 GREEN: fold the sanitizer case in, delete the dispatcher** — `c52530f` (feat)
3. **Task 2: remove the per-row exclusion, admit the escape row** — `7f9302e` (test)
4. **Task 3: delete the superseded guard and remaining stale prose** — `f8a5a15` (feat)

_Note on Task 2's single commit: see `key-decisions` -- Task 1's collapse already made the real mutation-table data satisfy the new declared-state check, so there was no separate GREEN implementation step to commit. Genuine RED was demonstrated by temporarily clearing the real `EscapeID` field (not committed), observing the exact expected failure, then reverting via `git checkout` before writing the single commit that both introduces and satisfies the check._

**Metadata commit:** (this SUMMARY + REQUIREMENTS.md + STATE.md + ROADMAP.md, committed separately)

## Files Created/Modified

- `internal/compiler/session/session_phase5.go` — `Phase5AssertMutationMovesAnAxis` deleted.
- `internal/compiler/session/session_phase5_alias.go` — sanitizer allocator-mismatch case folded into `AssertMutationMovesAnAxis`'s switch; default-arm error message rewritten; doc comment updated to state this is the single axis-movement law.
- `internal/compiler/session/session_phase5_test.go` — call site retargeted from the deleted dispatcher to the surviving law; two new RED-then-GREEN falsifiers added; `TestNoNAT03RowRemainsPending` deleted; stale prose rewritten at two sites.
- `internal/compiler/session/session_phase5_alias_test.go` — `TestEveryMutationMovesItsClaimedAxis`'s fixture-path skip replaced by a declared-state branch; stale prose rewritten at one site.
- `internal/compiler/session/witness_registry_test.go` — one descriptive comment naming the retired marker string rewritten.

## Decisions Made

See `key-decisions` in frontmatter for full detail. In summary: (1) Task 2 is a single test commit, not a RED/GREEN pair, because Task 1 already made the real data correct -- RED was demonstrated manually (clear-and-revert) rather than via a separate failing commit; (2) two additional stale-marker prose sites beyond the plan's named five were found and fixed, because the plan's own acceptance criterion is a repo-wide grep with no site list, and both were genuine surviving occurrences of the retired string.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Two additional stale-marker prose sites beyond the plan's named five**
- **Found during:** Task 3 (final `grep -rn 'PENDING-05-08' internal cmd` check)
- **Issue:** The plan's `<read_first>` named five prose sites to rewrite. After fixing those, the marker string still survived in two places the plan's research did not enumerate: (a) a negative-assertion string literal this plan's own Task 1 RED test added (`session_phase5_test.go`, asserting the error message did NOT contain `"PENDING-05-08"` -- ironically itself an occurrence of the string), and (b) a descriptive hand-off comment in `witness_registry_test.go` (plan 14-07's own SUMMARY hand-off note, predicting "plan 14-08 removes it entirely").
- **Fix:** Reworded the Task 1 test to assert on the actual retired substring (`"row not yet subjected"`) rather than the marker's ID form, and reworded the `witness_registry_test.go` comment to describe the removal in the past tense without repeating the marker string.
- **Files modified:** `internal/compiler/session/session_phase5_test.go`, `internal/compiler/session/witness_registry_test.go`
- **Verification:** `grep -rn 'PENDING-05-08' internal cmd` returns zero matches (exit 1); `go test ./internal/compiler/session/...` still passes.
- **Committed in:** `f8a5a15` (Task 3 commit)

---

**Total deviations:** 1 auto-fixed (1 missing critical). **Impact:** Necessary to satisfy the plan's own literal acceptance criterion (a repo-wide grep, not a five-site list) and its `must_haves.truths` claim that the marker "no longer survives anywhere ... not as prose in a test file." No scope creep -- both fixes are prose-only comment/string rewrites with no behavior change.

## Issues Encountered

None beyond the deviation documented above.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- EVD-05 closes: the mutation axis-movement law has exactly one implementation, zero per-row exclusions, and the superseded guard is gone alongside every surviving copy of its stale marker.
- `go test ./...` is green across all 25 packages; `go vet ./...` is clean.
- `git status --short` clean except the pre-existing untracked `.planning/milestone.lock`.
- No blockers for the next plan in this phase.

---
*Phase: 14-evidence-instrument-and-honest-scoping*
*Completed: 2026-09-18*

## Self-Check: PASSED

- FOUND: commit `36d835e` (test, RED, Task 1)
- FOUND: commit `c52530f` (feat, GREEN, Task 1)
- FOUND: commit `7f9302e` (test, Task 2)
- FOUND: commit `f8a5a15` (feat, Task 3)
- `grep -rn 'Phase5AssertMutationMovesAnAxis' internal cmd` prints nothing (exit 1)
- `grep -c 'PENDING-05-08' internal/compiler/session/session_phase5_alias.go` prints `0`
- `grep -rn 'PENDING-05-08' internal cmd` prints nothing (exit 1)
- `grep -rn 'TestNoNAT03RowRemainsPending' internal cmd` prints nothing (exit 1)
- `grep -c 't.Skip' internal/compiler/session/session_phase5_alias_test.go` prints `0`
- `go test ./internal/compiler/session/... -run 'TestEveryMutationMovesItsClaimedAxis|TestAssertMutationMovesAnAxisFailsOnAMislabeledAxis|TestNAT03MutationTableHasSevenRows|TestNoSuppressionOutlivesItsWitness|TestSuppressionWitnessGuardIsNotInert' -count=1 -v` prints only `--- PASS:` lines, 7 subtests under `TestEveryMutationMovesItsClaimedAxis`, none skipped
- `go test ./...` reports `ok` for all 25 packages
- `go vet ./...` is clean
- `git status --short` clean except the pre-existing untracked `.planning/milestone.lock`
