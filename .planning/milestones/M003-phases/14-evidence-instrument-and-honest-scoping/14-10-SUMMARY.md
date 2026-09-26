---
phase: 14-evidence-instrument-and-honest-scoping
plan: 10
subsystem: process
tags: [go, groundedness-lint, reconciliation, debt-register, evd-01, evd-03, generated-view, evidence-run-record]

requires:
  - phase: 14-evidence-instrument-and-honest-scoping
    provides: "plan 14-06's role-based Tier-A scope and R1/R2/R2b/R3 classifier (verification_groundedness_test.go), plan 14-07's closed witness grammar and generated-view discipline (witness_registry_test.go, .planning/UNREACHABLE-CLAIMS.md), plan 14-09's re-pinned 124-entry frontier and evidence-run-record.sh"
provides:
  - "internal/compiler/session/session_test.go's closed four-member reconciliation verdict vocabulary (renamed/superseded/obsolete-by-design/under-scoped, D-14-12) with mechanically checked obligations, TestReconciliationVerdictsCarryTheirObligations, and parseReconciliationEntries"
  - "66 reconciliation entries in PHASE-14-DEBT.md (D-14-55..D-14-120), one per runnability/groundedness/grep finding in the pre-plan pinned frontier, each a debt row with a witness carrying a \"```reconciliation\" fenced block, never rewriting the archived document it was found in"
  - ".planning/EVIDENCE-RECONCILIATION.md -- a generated, byte-compared view (witness_registry_test.go's deriveEvidenceReconciliation/renderEvidenceReconciliationView) with all four D-14-14 anti-decay guards proven"
  - "scripts/assert-reconciliation-touched.sh -- the archive-edit coupling rule (a commit touching an archived VALIDATION/VERIFICATION doc without touching the reconciliation view exits 1)"
  - "the pinned groundedness frontier re-measured and reduced from 124 to 58 entries: runnability (R1), groundedness (R2) and grep (R3) now assert empty (TestVerificationGroundednessThreeClassesAreEmpty); per-branch (R2b, 10 entries) stays pinned with every member owned via the new r2bLandingPhases register (all P20, ROADMAP's QLT-10 assignment)"
  - "14-VALIDATION.md's Per-Task Verification Map filled with 31 real rows (one per task across all ten plans in this phase), every Automated Command executed and confirmed to resolve"
affects: ["Phase 20 (QLT-10) -- inherits the per-branch (R2b) frontier class this plan deliberately left pinned, plus D-14-121's deferred corpus-wide grade-cap re-verification for 14-VALIDATION.md"]

actuals:
  tokens: 50857
  tasks: 3
  commits: 5

tech-stack:
  added: []
  patterns:
    - "Closed reconciliation verdict vocabulary (renamed/superseded/obsolete-by-design/under-scoped) with per-verdict obligations reusing existing primitives -- classifyCommand for \"resolves\", debtRegisterCallSiteWitnessMatches for \"symbol absent from tree\" (the anti-source-search rule), a new reconciliationCommitExists for \"commit is real\""
    - "A markdown-safe machine-readable grammar (\"```reconciliation\" fenced code blocks inside a debt register's Detail prose) for structured data that cannot live in a pipe-delimited table cell because it legitimately contains unescaped \"|\" characters"
    - "Lint exclusion via a live reconciliation-set lookup (reconciledFindings), not archive editing -- a finding stays textually present in the document it was found in; only the FRONTIER's own comparison excludes it, so TestReconciliationVerdictsCarryTheirObligations keeps independently re-checking every entry's obligation on every run"
    - "Generated-view command-span self-reference guard: a view quoting a bad command verbatim (for traceability) must never render it as a bare backtick-delimited span matching the lint's own command-detection regex, or the view promotes itself to Tier-A and finds violations in its own quotations (\"cmd: \" prefix breaks the anchored match while staying legible)"

key-files:
  created:
    - .planning/EVIDENCE-RECONCILIATION.md
    - scripts/assert-reconciliation-touched.sh
  modified:
    - internal/compiler/session/session_test.go
    - internal/compiler/session/witness_registry_test.go
    - internal/compiler/session/verification_groundedness_test.go
    - internal/compiler/session/evidence_grade_test.go
    - .planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md
    - .planning/phases/14-evidence-instrument-and-honest-scoping/14-VALIDATION.md
    - .planning/REQUIREMENTS.md

key-decisions:
  - "Reconciliation entries carry Grade=DEFINED/Witness=n/a in PHASE-14-DEBT.md's Items table (never a probe: token) -- deliberately kept OUT of the Grade/Witness closed-vocabulary law plan 14-07 built, so the 66 new rows do not inflate .planning/UNREACHABLE-CLAIMS.md (confirmed unchanged, still 12 entries). The obligation check is a separate, dedicated law (TestReconciliationVerdictsCarryTheirObligations) operating on the \"```reconciliation\" block, not routed through the witness grammar."
  - "Verdict assignment for all 66 findings was researched individually against the real tree (git log -S for renames/deletions, grep for surviving coverage), never guessed: every 'renamed' replacement and 'superseded' covering command was verified via classifyCommand to actually resolve before being recorded; every 'obsolete-by-design' deleted symbol was verified absent via the AST scan before being recorded. See the Verification section for the exact commands run."
  - "Anchored two 14-VALIDATION.md patterns with a trailing $ on each alternation branch (never wrapping the whole group in ^(...)$, which breaks D-14-16's per-branch/R2b splitter) after discovering an unanchored TestValidationRowGradesAreEarned citation substring-matched TestValidationRowGradesAreEarnedOverArchivedCorpus itself, causing a recursive corpus-scan invocation that pushed the session package past Go's 600s default per-package timeout. Reproduced and fixed with direct before/after timing (session package: ~600-608s timing out, three times, isolated from machine contention by killing stray go-test processes between runs -> 268-347s clean pass after the anchor fix)."
  - "Declined to remove 14-VALIDATION.md's stale validationGradeBarExemptions entry (whose stated reason, 'never filled in', is now literally false since this plan filled it in) after the removal, when tested, surfaced the SAME corpus-wide timeout risk across nine other unrelated already-frozen archived files on this specific machine at this specific load. Reverted rather than landed under that uncertainty; recorded as debt D-14-121 with the exact reproduction and reasoning, rather than either silently leaving the stale rationale or blindly forcing the change through."

patterns-established:
  - "A reconciliation obligation's 'resolves' check is defined once (classifyCommand) and reused by both the renamed and superseded verdicts, so 'resolves' means the same thing everywhere the lint uses it."

requirements-completed: [EVD-03, EVD-01]

coverage:
  - id: D1
    description: "Closed four-member reconciliation verdict vocabulary (renamed/superseded/obsolete-by-design/under-scoped) with a mechanically checked obligation per verdict; a fifth value fails naming the offender; each obligation has a passing positive and passing negative subtest, including the rename-to-a-dead-pattern case and the obsolete-by-design-symbol-still-exists case"
    requirement: "EVD-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestReconciliationVerdictsCarryTheirObligations"
        status: pass
    human_judgment: false
  - id: D2
    description: "66 reconciliation entries authored in PHASE-14-DEBT.md (D-14-55..D-14-120), one per runnability/groundedness/grep finding in the pre-plan 124-entry pinned frontier; every entry's obligation independently re-verified to hold against the live tree; no archived document modified (git diff --stat -- .planning/milestones is empty); no inline suppression syntax or count-keyed ceiling present"
    requirement: "EVD-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestReconciliationVerdictsCarryTheirObligations/every_real_reconciliation_entry's_obligation_holds"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestReconciliationVerdictsCarryTheirObligations/the_reconciliation_ledger_exactly_covers_the_live_corpus's_R1/R2/R3_findings"
        status: pass
      - kind: other
        ref: "git diff --stat -- .planning/milestones (empty output, confirmed)"
        status: pass
      - kind: other
        ref: "grep -nE 'nolint|ignore-file|maxViolations|allowedCount' internal/compiler/session/session_test.go (no matches, confirmed)"
        status: pass
    human_judgment: false
  - id: D3
    description: ".planning/EVIDENCE-RECONCILIATION.md is a generated, byte-compared view with all four D-14-14 anti-decay guards proven: frontmatter entries:N matches the derivation, hand-edit detection, stale-entry detection (never auto-pruned), and non-zero-row-count is load-bearing. scripts/assert-reconciliation-touched.sh implements the archive-coupling rule, exercised against synthetic git repositories for both verdicts (0=coupled, 1=uncoupled) before being run against this repository's own real HEAD~1..HEAD range"
    requirement: "EVD-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/witness_registry_test.go#TestEvidenceReconciliationViewIsCurrent"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/witness_registry_test.go#TestEvidenceReconciliationViewNonZeroRowCountIsLoadBearing"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/witness_registry_test.go#TestEvidenceReconciliationViewCatchesHandEdit"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/witness_registry_test.go#TestEvidenceReconciliationViewCatchesStaleEntry"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/witness_registry_test.go#TestEvidenceReconciliationViewCountMatchesFrontmatter"
        status: pass
      - kind: other
        ref: "manual synthetic-repo exercise of scripts/assert-reconciliation-touched.sh (coupled=exit 0, uncoupled=exit 1), then sh scripts/assert-reconciliation-touched.sh HEAD~1..HEAD against this repository (exit 0)"
        status: pass
    human_judgment: false
  - id: D4
    description: "The pinned groundedness frontier's runnability (R1), groundedness (R2) and grep (R3) classes assert empty over a live corpus scan with reconciled findings excluded; per-branch (R2b, 10 entries) remains non-empty with every member carrying a closed-vocabulary landing phase (all P20); corpus floors (documents and commands) asserted non-degraded in the same run"
    requirement: "EVD-01"
    verification:
      - kind: unit
        ref: "internal/compiler/session/verification_groundedness_test.go#TestVerificationGroundednessThreeClassesAreEmpty"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/verification_groundedness_test.go#TestVerificationGroundednessFrontierIsPinned"
        status: pass
    human_judgment: false
  - id: D5
    description: "14-VALIDATION.md's Per-Task Verification Map filled with 31 rows spanning all ten plans in this phase; every Automated Command executed during this task and confirmed to resolve; grades declared at the derived ceiling per TestValidationRowGradesAreEarnedOverArchivedCorpus (corrected iteratively against real derivation output, not guessed)"
    requirement: "EVD-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/evidence_grade_test.go#TestValidationRowGradesAreEarnedOverArchivedCorpus"
        status: pass
      - kind: other
        ref: "each of the 31 rows' Automated Command executed individually during this task (see Verification section); all passed or produced expected, correctly-classified output"
        status: pass
    human_judgment: true
    rationale: "Grade declarations for the 31 rows were fit iteratively against one mechanized signal (the derivation ladder); whether a specific grade is the MOST accurate available characterization (vs. merely non-overclaiming) is a judgment call worth a human's read, per this file's own D-14-02 discipline that grading below the derived ceiling is always safe but not necessarily maximally informative."
  - id: D6
    description: "All eleven phase requirement rows (EVD-01..08, DX-08, DX-09, PRC-01) confirmed marked complete in REQUIREMENTS.md with supporting evidence; EVD-01's own text updated from its stale pre-lint baseline description to describe the shipped lint and this plan's reconciliation of it. One new debt item (D-14-121) recorded honestly rather than silently absorbed into a requirement's marked-complete status."
    requirement: "EVD-01"
    verification:
      - kind: other
        ref: "grep -n 'EVD-0\\|DX-08\\|DX-09\\|PRC-01' .planning/REQUIREMENTS.md (all eleven present and marked [x] Complete)"
        status: pass
    human_judgment: true
    rationale: "Confirming eleven requirements' prose is honestly current (not just checked off) requires reading each against the real shipped mechanism; this plan verified EVD-01's own text directly (since this plan is its primary closer) but did not perform a full textual audit of the other ten, which were closed by their own owning plans."

duration: 165min (first RED commit to final metadata commit; dominated by per-finding forensic research for 66 reconciliation entries and by diagnosing/fixing a real corpus-wide test timeout regression)
completed: 2026-09-18
status: complete
---

# Phase 14 Plan 10: Nyquist Reconciliation — Frontier Emptied, Phase Closed Summary

**66 runnability/groundedness/grep findings reconciled outside the archives under a closed four-verdict vocabulary with mechanically checked obligations, the pinned groundedness frontier driven from 124 to 58 entries (three of four classes now empty), `.planning/EVIDENCE-RECONCILIATION.md` shipped as a byte-compared generated view, and 14-VALIDATION.md's own verification map filled with 31 real, executed rows — closing Phase 14.**

## Performance

- **Duration:** ~165 min (RED commit `a352717` to the final metadata commit)
- **Tasks:** 3
- **Files modified:** 8 (2 created, 6 modified in task commits; plus SUMMARY.md/STATE.md/ROADMAP.md in the metadata commit)

## Accomplishments

- **Task 1:** Implemented the closed four-member reconciliation verdict vocabulary (`renamed`/`superseded`/`obsolete-by-design`/`under-scoped`, D-14-12) in `session_test.go`, each with a mechanically checked obligation reusing existing primitives (`classifyCommand` for "resolves," `debtRegisterCallSiteWitnessMatches` for the anti-source-search "symbol absent from tree" check, a new `reconciliationCommitExists` for "the cited commit is real"). Authored 66 reconciliation entries in `PHASE-14-DEBT.md` (D-14-55..D-14-120) as debt rows with a `` ```reconciliation ``` `` fenced block in each Detail section (never the Items table, which cannot host raw command text containing an unescaped `|`) — one entry per runnability, groundedness, and grep finding in the pre-plan 124-entry pinned frontier. Every finding was individually researched against the real tree (git history for renames/deletions, live grep/test-name checks for surviving coverage) before being recorded; no finding's verdict was guessed.
- **Task 2:** Generated `.planning/EVIDENCE-RECONCILIATION.md` from the 66 reconciliation entries, reusing plan 14-07's `UNREACHABLE-CLAIMS.md` generated-view discipline (regenerate in memory, byte-compare, no blessing path). All four D-14-14 anti-decay guards proven live: frontmatter count cross-check, hand-edit detection, stale-entry detection (never auto-pruned), and non-zero-row-count as a load-bearing assertion. `scripts/assert-reconciliation-touched.sh` implements the archive-coupling rule (a commit touching an archived `*-VALIDATION.md`/`*-VERIFICATION.md` document without also touching the reconciliation view exits 1), exercised against synthetic git repositories for both verdicts before being run against this repository's own real history.
- **Task 3:** Excluded every reconciled finding from `TestVerificationGroundednessFrontierIsPinned`'s comparison and re-measured the pinned literal: 124 → 58 entries. `TestVerificationGroundednessThreeClassesAreEmpty` asserts runnability (R1), groundedness (R2) and grep (R3) are empty, and that per-branch (R2b, 10 entries) stays non-empty with every member carrying a closed-vocabulary landing phase — all `P20`, per `ROADMAP.md`'s QLT-10 (Phase 20) assignment, consistent with D-14-16's sizing decision that per-branch closure is judgment work deferred past this phase. Filled `14-VALIDATION.md`'s Per-Task Verification Map with 31 real rows spanning all ten plans in this phase, every command executed and confirmed to resolve. Confirmed all eleven phase requirements (`EVD-01`..`08`, `DX-08`, `DX-09`, `PRC-01`) are marked complete in `REQUIREMENTS.md`; updated `EVD-01`'s own text, which had gone stale describing the pre-lint baseline rather than the shipped mechanism.
- **Found and fixed a real performance regression this same task exposed** (see Deviations): an unanchored citation in the newly-filled `14-VALIDATION.md` recursively re-invoked the corpus-wide grade-cap test inside its own run-record generation, pushing the session package past Go's 600s default per-package timeout. Diagnosed via direct, repeated, isolated measurement (not assumed), fixed by anchoring, and verified `go test ./...` now exits 0 with no `-timeout` override, session package at 291-347s across repeated clean runs.

## Task Commits

1. **Task 1 RED: add failing test for the reconciliation verdict vocabulary** — `a352717` (test)
2. **Task 1 GREEN: implement the vocabulary and reconcile 66 findings** — `8cc9ba8` (feat)
3. **Task 2 RED: add failing test for the generated reconciliation view** — `9db46b1` (test)
4. **Task 2 GREEN: generate the view and the archive-coupling script** — `6685877` (feat)
5. **Task 3: empty three frontier classes, fill 14-VALIDATION.md, close phase evidence** — `0dcb460` (feat; includes the timeout-regression fix, since it was discovered and fixed within this same task before the frontier re-pin could be honestly asserted)

_Note: Task 1 and Task 2 are `tdd="true"`. Each RED commit stubbed the function under test (`reconciliationEntryProblem` returning `""` unconditionally for Task 1; the checked-in `.planning/EVIDENCE-RECONCILIATION.md` simply did not exist yet for Task 2) so failure was genuine and observed, not asserted from memory, before the GREEN commit restored/added the real implementation. Task 3 is `type="auto"` (no `tdd` attribute) and is GREEN-only, consistent with its frontmatter._

**Plan metadata:** committed separately (this SUMMARY + STATE.md + ROADMAP.md)

## Files Created/Modified

- `internal/compiler/session/session_test.go` — the reconciliation verdict vocabulary, obligation checker, `parseReconciliationEntries`, `TestReconciliationVerdictsCarryTheirObligations`.
- `internal/compiler/session/witness_registry_test.go` — `deriveEvidenceReconciliation`/`renderEvidenceReconciliationView`, the five `TestEvidenceReconciliationView*` tests.
- `internal/compiler/session/verification_groundedness_test.go` — `pinnedFrontier` reduced 124→58 (R1/R2/R3 entries removed), `reconciledFindings`, `r2bLandingPhases`, `TestVerificationGroundednessThreeClassesAreEmpty`, `TestVerificationGroundednessFrontierIsPinned` updated to exclude reconciled findings.
- `internal/compiler/session/evidence_grade_test.go` — untouched in the final committed state (an exemption-removal experiment was tried and reverted; see Deviations/D-14-121).
- `.planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md` — 66 new reconciliation rows (D-14-55..D-14-120) plus D-14-121 (the timeout-regression debt row).
- `.planning/phases/14-evidence-instrument-and-honest-scoping/14-VALIDATION.md` — Per-Task Verification Map filled with 31 rows.
- `.planning/EVIDENCE-RECONCILIATION.md` — new generated view (66 entries).
- `scripts/assert-reconciliation-touched.sh` — new coupling script.
- `.planning/REQUIREMENTS.md` — EVD-01's text updated to describe the shipped mechanism.

## Decisions Made

See `key-decisions` in frontmatter for full detail. In summary: (1) reconciliation entries stay `Grade=DEFINED`/`Witness=n/a` in the Items table, deliberately outside the witness-grammar law, so they don't inflate `UNREACHABLE-CLAIMS.md`; (2) every one of the 66 verdicts was individually researched against real git history and live test resolution, never guessed; (3) two `14-VALIDATION.md` patterns were anchored with a trailing `$` per branch (not a wrapping `^(...)$`, which breaks per-branch/R2b splitting) after a real, reproduced, three-times-confirmed timeout regression traced to a recursive self-match; (4) removing `14-VALIDATION.md`'s stale grade-bar exemption was attempted, found to expose the same timeout risk across nine unrelated files, and reverted rather than landed under uncertainty — recorded as debt D-14-121 instead.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Recursive corpus-scan self-invocation from an unanchored pattern citation**
- **Found during:** Task 3, after filling `14-VALIDATION.md`'s Per-Task Verification Map and running `go test ./...` to verify — the `session` package failed with a goroutine-dump timeout at 600-608s (Go's default per-package ceiling), reproduced identically three separate times after killing every stray `go-test`/`go-build` process between attempts to rule out machine contention.
- **Issue:** Two of the 31 new `14-VALIDATION.md` rows cited the plan's own literal `-run 'TestValidationRowGradesAreEarned'` command (copied verbatim from the plan's `<automated>` text, as every other row's citation was). As an unanchored regex, this string is also a substring of `TestValidationRowGradesAreEarnedOverArchivedCorpus` — the very (expensive, run-record-generating) test that scans `14-VALIDATION.md`. Once merged into the `session` package's consolidated run-record alternation, this caused the corpus-wide test to recursively re-invoke itself inside its own generated subprocess batch. No other archived `*-VALIDATION.md` document cited this pattern (confirmed by grep), so this is directly attributable to this task's own new content, not a pre-existing defect.
- **Fix:** Anchored each alternation branch with a trailing `$` (`TestValidationGradeVocabularyIsClosed$\|TestValidationGradeOrderIsTotal$\|TestValidationRowGradesAreEarned$` for 14-09-T1; `TestValidationRowGradesAreEarned$` for 14-09-T2). A wrapping `^(...)$` was tried first and rejected: it broke `TestVerificationGroundednessThreeClassesAreEmpty`'s per-branch (R2b) splitter, which expects each alternative to carry its own anchors, producing a spurious new R2b finding.
- **Files modified:** `.planning/phases/14-evidence-instrument-and-honest-scoping/14-VALIDATION.md`
- **Verification:** `go test ./internal/compiler/session/... -run 'TestValidationRowGradesAreEarnedOverArchivedCorpus' -count=1 -v` dropped from timing out at exactly its configured budget (three repeated measurements: 600.230s, 300.07s/300.518s at a 300s budget, 400.24s/401.078s at a 400s budget I tried and reverted) to a clean PASS at 105-108s across three repeated clean runs. Full `go test ./internal/compiler/session/... -count=1` dropped from 600-608s (FAIL, timeout) to 261-347s (PASS) across four repeated runs. Full `go test ./... -count=1` (no `-timeout` override, matching the plan's own literal acceptance criterion) now exits 0 with every package `ok`.
- **Committed in:** `0dcb460` (Task 3 commit)

**2. [Rule 1 - Bug, three sub-fixes] Grade declarations on the newly-filled 14-VALIDATION.md rows exceeded their derived ceiling**
- **Found during:** Task 3, iterating `TestValidationRowGradesAreEarnedOverArchivedCorpus` against the real corpus after the recursion fix above.
- **Issue:** Three of the 31 rows were initially declared `EXERCISED` but the mechanized derivation ladder (D-14-03) only supports a lower ceiling for their evidence shape: `14-01-T2` (`TestVerificationGroundednessFrontierIsPinned`) and `14-10-T3` (`TestVerificationGroundednessThreeClassesAreEmpty`) derived only `WIRED`; `14-02-T1` (a `go run ./cmd/lang --json check ...` CLI invocation, not a `go test -run` pattern) derived only `REACHABLE` per the ladder's own CLI-invocation rung.
- **Fix:** Declared grades corrected to their derived ceilings (`WIRED`, `WIRED`, `REACHABLE` respectively) — never claiming above what the mechanized cap supports, per this plan's own EVD-02-inherited discipline.
- **Files modified:** `.planning/phases/14-evidence-instrument-and-honest-scoping/14-VALIDATION.md`
- **Verification:** `TestValidationRowGradesAreEarnedOverArchivedCorpus` passes cleanly for `14-VALIDATION.md` (and all fourteen documents) after the corrections.
- **Committed in:** `0dcb460` (Task 3 commit)

**3. [Rule 4 - Architectural, declined] Removing 14-VALIDATION.md's stale grade-bar exemption**
- **Found during:** Task 3, after filling `14-VALIDATION.md` for real, noticing `validationGradeBarExemptions`'s stated reason for exempting it ("template row never filled beyond plan-time placeholder") was now literally false.
- **What was tried:** Removed the exemption entry and re-ran `TestValidationRowGradesAreEarnedOverArchivedCorpus`.
- **What it exposed:** The SAME corpus-wide timeout symptom (many files failing "declared grade exceeds ceiling" at exactly the configured budget), now spanning nine OTHER, unrelated, already-frozen archived files whose own exemptions were never touched — indicating the run-record generation was not completing within its 300s budget on this specific machine at this specific load, not a logic defect in the removed exemption itself.
- **Decision:** Reverted (`git checkout --`) rather than landed under that uncertainty. This is a genuine architectural/scope question (whether `evidence-run-record.sh`'s sequential-batch design, or its timeout budget, needs a second pass — the exact kind of change plan 14-09 itself made once already) that this plan did not have the budget to safely resolve. Recorded as debt `D-14-121` with the full reproduction, rather than either silently leaving the stale exemption rationale in place or forcing an unverified change through.
- **Files modified:** none (reverted); `.planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md` gained the D-14-121 row documenting the attempt and its outcome.
- **Committed in:** `0dcb460` (Task 3 commit, as the debt row; no code change)

---

**Total deviations:** 2 auto-fixed (both Rule 1 — a genuine performance regression this task's own new content caused, and three grade-declaration corrections against the mechanized cap), 1 declined-and-recorded (Rule 4-adjacent — an architectural risk surfaced but not resolved, honestly deferred as debt rather than forced through).
**Impact on plan:** The recursion fix and grade corrections were necessary for the plan's own literal acceptance criterion (`go test ./...` exits 0, no `-timeout` override) to actually hold — exactly the failure mode (an instrument or its supporting fixtures silently going stale or silently becoming too expensive under a legitimate edit) this phase exists to retire. The declined exemption removal is scope discipline, not a shortcut: forcing it through without being able to verify it on this machine would have risked landing an unverified corpus-wide grade change.

## Issues Encountered

- **Machine contention during diagnosis.** This session's own repeated `go test` invocations (used to isolate the recursion bug from machine load) left orphaned child subprocesses (the run-record generator's own sequential `go test` batch) running past their parent's completion, causing several early diagnostic runs to show inflated timings and, once, spurious failures in the unrelated `testsupport` package (native-differential-lane timeouts) that were confirmed to be contention artifacts, not real regressions, once the process tree was fully cleaned between runs. The final, clean, repeated measurements (in Verification, above) are the ones this SUMMARY's claims rest on.
- **Load average on the shared build machine stayed elevated (4.7-9.3, 18 cores, 12 logged-in users) throughout this plan's execution**, which is disclosed here because it is the most plausible explanation for why the corpus-wide grade-cap test's cost varies run-to-run (105s to 268s observed for the same test) — a real, external cost-lane variable this plan did not attempt to control, only to stay safely under budget against.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- Phase 14 is complete: all ten plans executed, `go test ./...` exits 0 (25 packages, no `-timeout` override), `go vet ./...` clean, `go build ./...` clean.
- The groundedness lint's pinned frontier holds zero runnability, groundedness and grep findings, with the per-branch class pinned and owned (`P20`).
- `.planning/EVIDENCE-RECONCILIATION.md` and `.planning/UNREACHABLE-CLAIMS.md` are both live, generated, byte-compared views with proven anti-decay guards.
- All eleven phase requirements are marked complete with supporting evidence named.
- **Open debt inherited by later phases:** D-14-121 (14-VALIDATION.md's grade-bar exemption re-verification, blocked on a run-record-generation cost/timeout question `evidence-run-record.sh` itself may need a second optimization pass to resolve — mirroring plan 14-09's own prior fix); the per-branch (R2b) groundedness frontier class (10 entries, owned `P20`/QLT-10/Phase 20); the 87 unmigrated `02-DEBT.md`..`PHASE-10-DEBT.md` rows (recorded in plan 14-07's own debt list); the live-module suppression scanner's narrower-than-literal scope (also plan 14-07's debt).
- No blockers for Phase 15.

---
*Phase: 14-evidence-instrument-and-honest-scoping*
*Completed: 2026-09-18*

## Self-Check: PASSED

- FOUND: `.planning/EVIDENCE-RECONCILIATION.md`
- FOUND: `scripts/assert-reconciliation-touched.sh` (executable)
- FOUND: `.planning/phases/14-evidence-instrument-and-honest-scoping/14-10-SUMMARY.md` (this file)
- Commits present in `git log --oneline --all`: `a352717`, `8cc9ba8`, `9db46b1`, `6685877`, `0dcb460` — all FOUND. The plan metadata commit is self-referential and cannot state its own SHA: verify with `git log --oneline -1 -- .planning/phases/14-evidence-instrument-and-honest-scoping/14-10-SUMMARY.md`.
- `go test ./internal/compiler/session/... -run 'TestReconciliationVerdictsCarryTheirObligations' -count=1 -v` re-run: all `--- PASS:` lines, no failures.
- `go test ./internal/compiler/session/... -run 'TestEvidenceReconciliationViewIsCurrent|TestEvidenceReconciliationViewNonZeroRowCountIsLoadBearing|TestEvidenceReconciliationViewCatchesHandEdit|TestEvidenceReconciliationViewCatchesStaleEntry|TestEvidenceReconciliationViewCountMatchesFrontmatter' -count=1 -v` re-run: all PASS.
- `go test ./internal/compiler/session/... -run 'TestVerificationGroundednessFrontierIsPinned|TestVerificationGroundednessThreeClassesAreEmpty|TestVerificationGroundednessCorpusIsNotEmpty|TestUnreachableClaimsViewIsCurrent|TestDebtRegistersAreWellFormed' -count=1 -v` re-run: all PASS.
- `sh scripts/assert-reconciliation-touched.sh HEAD~1..HEAD` re-run: `exit=0`.
- `sh scripts/assert-go-tests.sh --self-test ./internal/compiler/session TestVerificationGroundednessFrontierIsPinned` re-run: `exit=0`.
- `git diff --stat -- .planning/milestones` re-run: empty (no archived document modified by this plan).
- `grep -nE 'nolint|ignore-file|maxViolations|allowedCount' internal/compiler/session/session_test.go` re-run: no matches.
- `go build ./...` and `go vet ./...` re-run: both clean.
- Full, unfiltered `go test ./...` re-run after the final commit, exactly as the plan's own verify command specifies (no `-timeout` override): `EXIT:0`, every package `ok`, session package `291.758s`-`346.515s` across repeated runs (comfortably under Go's 600s default per-package ceiling).
- `git status --short`: clean except the pre-existing untracked `.planning/milestone.lock`.
