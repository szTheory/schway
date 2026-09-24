---
status: diagnosed
trigger: "Phase 16 verify-work: focused groundedness tests report an unowned R2b finding at 16-VERIFICATION.md:143"
created: 2026-09-24T04:16:22Z
updated: 2026-09-24T04:17:32Z
---

## Current Focus

hypothesis: "The report added a grouped alternation after the frontier snapshot; the R2b splitter treats the group's parentheses as part of individual branches, so a valid two-test command becomes an unowned scanner finding."
test: "Compare the original and ungrouped equivalent Go test selectors; inspect exact tuple ownership in pinnedFrontier and r2bLandingPhases."
expecting: "Both selectors select the same two names, while only the grouped selector produces an invalid split branch under the current classifier."
next_action: "Return diagnosis to the Phase 16 verify-work orchestrator for a report-command correction and focused rerun."
bug_class: Bohrbug
candidate_causes:
  - "code: classifyPerBranchGroundedness uses a literal pipe split without group awareness"
  - "data: a new scanner-visible report command uses grouped alternation"
  - "config: exact frontier and landing-phase registers were last updated before the report row"
and_gate: "yes: the false R2b requires both grouped regex syntax and the scanner's literal pipe split; the test failures additionally require the resulting tuple to be absent from exact ownership registers"

## Symptoms

expected: "Phase 16 full Go suite passes with every retained verification command assigned to its owning phase."
actual: "Only the two groundedness tests fail; they report an unowned R2b command at 16-VERIFICATION.md:143."
errors: "unexpected extra violation not in the pinned frontier; R2b finding has no landing phase"
reproduction: "GOCACHE=/tmp/ai-lang-gocache go test ./..."
started: "After the Phase 16 verification report added its NAT-09 owner-law command."

## Eliminated

## Evidence

- timestamp: 2026-09-24T04:16:22Z
  checked: "knowledge-base.md and current test/report source"
  found: "No exact prior KB resolution; report line 143 cites a two-branch session test command; current pinnedFrontier and r2bLandingPhases have no matching Phase 16 report entry."
  implication: "A report edit may have advanced the scanner input without advancing its exact-key ownership registers."
- timestamp: 2026-09-24T04:17:32Z
  checked: "Focused groundedness test reproduction"
  found: "Both tests fail on exactly 16-VERIFICATION.md:143; R1=0 R2=0 R3=0, R2b=16, and corpus floors hold (574 documents, 625 commands)."
  implication: "One new R2b record is the remaining groundedness failure."
- timestamp: 2026-09-24T04:17:32Z
  checked: "Per-branch classifier and test index"
  found: "classifyPerBranchGroundedness splits the regex on every literal pipe. The report's '^(TestPhase16EmitterCutsAreAmendedAndOwned|TestDebtRegistersAreWellFormed)$' therefore yields '^(' and 'TestDebtRegistersAreWellFormed)$' fragments; the latter is an invalid standalone regexp. Both test names exist in the session package."
  implication: "R2b is a scanner artifact from a valid grouped alternation, not a missing test."
- timestamp: 2026-09-24T04:17:32Z
  checked: "History and semantically equivalent selector"
  found: "ba5ed41 introduced the owner-law row after 919e8d7 pinned the frontier. The ungrouped '^TestPhase16EmitterCutsAreAmendedAndOwned$|^TestDebtRegistersAreWellFormed$' selector lists the same two tests and runs successfully."
  implication: "Changing only the report's selector to the equivalent top-level alternation should remove the artificial R2b without adding P20 debt."

## Resolution

root_cause: "A Phase 16 verification report row was added after the exact frontier snapshot. Its valid grouped -run alternation is mis-split by classifyPerBranchGroundedness, creating an artificial R2b tuple absent from pinnedFrontier and r2bLandingPhases."
fix: "Recommended smallest correction: rewrite the report's -run pattern as '^TestPhase16EmitterCutsAreAmendedAndOwned$|^TestDebtRegistersAreWellFormed$' (same two tests, independent anchored branches), then rerun focused groundedness and the Phase 16 verification freshness gate. If preserving the exact grouped command, add its exact R2b tuple to both pinnedFrontier and r2bLandingPhases with P20 ownership, but that would defer a classifier false positive."
verification: "Diagnosis only. Reproduced focused failures and proved equivalent ungrouped selector lists and runs the same two tests; no fix applied."
files_changed: []
