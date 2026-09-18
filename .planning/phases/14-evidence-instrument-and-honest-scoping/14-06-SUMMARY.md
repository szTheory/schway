---
phase: 14-evidence-instrument-and-honest-scoping
plan: 06
subsystem: evidence-instruments
tags: [go, testing, go-parser, groundedness-lint, static-analysis, evd-01, os-exec]

requires:
  - phase: 14-evidence-instrument-and-honest-scoping
    provides: "plan 14-01's tracer slice (static test index, Tier-A document scanner, R1/R2/unparseable classification, the original 26-entry pinned frontier) -- this plan extends that same file rather than creating a new one"
provides:
  - "internal/compiler/session/verification_groundedness_test.go extended with D-14-11 role-based document scoping (Tier A/Tier B with promotion), R3 (grep execution) and R2b (per-branch groundedness) classification, two corpus floors, TestStaticTestIndexMatchesGoTestList (D-14-10's accuracy control), and a re-measured 125-entry pinned frontier"
affects: ["14-09 (retires the freeform Status column -- must not leave Tier-A detection briefly undetectable; this plan already keys detection on the Per-Task Verification Map header and the Status/Grade column name)", "14-10 (Nyquist reconciliation, must drive the R1/R2/R2b/R3 frontier toward empty and adjudicate archived rows)", "QLT-10 (per-branch/R2b closure)", "any later plan touching .planning/**/*.md content"]

actuals:
  tokens: 17845
  tasks: 3
  commits: 6

tech-stack:
  added: []
  patterns:
    - "Illocutionary-role document scoping: basename-based enforced tier, frontmatter-declaration-gated exempt tier (opt-in only), and a content-based promotion rule keyed on a Status/Grade table column -- never a bare emoji scan of the row"
    - "Argv-form-only subprocess execution for document-supplied command text (grep/rg), bounded by a bounded writer + context timeout, exactly mirroring the existing git-status spawn shape in this same file"
    - "Shell-metacharacter detection outside quoted arguments, with command substitution ($() checked outside single quotes only (POSIX double quotes do not suppress it) -- distinct from the other four metacharacters, which are checked outside both quote kinds"
    - "Per-branch (union-then-branch) two-stage classification: a union check that preserves prior classification exactly, followed by a branch check that only fires when the union already passed"
    - "Runtime-assembled self-referential search patterns in a lint's own non-inertness/self-check subtests, so the pattern never appears as one contiguous literal in the lint's own source"

key-files:
  created: []
  modified:
    - internal/compiler/session/verification_groundedness_test.go

key-decisions:
  - "Kept Task 1 and Task 2's new classifiers (scope classification, grep execution, per-branch detection) as standalone functions with their own dedicated tests, NOT wired into the production measuredViolations/pinnedFrontier pipeline until Task 3. This was necessary because Task 1's and Task 2's own <verify> commands (`-run 'TestVerificationGroundedness'`) match TestVerificationGroundednessFrontierIsPinned by substring -- wiring broadened scope or new classifications into the pipeline before the frontier could be re-measured and re-pinned would have broken that test mid-task. Task 3 does the actual wiring and the one frontier re-pin, matching the plan's own task boundaries."
  - "R2 (union groundedness) is checked before R2b (per-branch groundedness) in classifyCommand: when every branch fails, the union also fails, so that case classifies R2 exactly as plan 14-01 pinned it (e.g. 09-VALIDATION.md:93's 'Mode.*Invalid|DecodeMode', both branches dead). R2b only fires when the union passes (>=1 branch resolves) but a specific OTHER branch does not -- this is what keeps the original 26 entries' classifications unchanged while adding the 11 new per-branch findings as a distinct class."
  - "classifyCommand reads projectRoot from index.root (already present on every testIndex from buildTestIndex) rather than adding a new parameter, so none of plan 14-01's existing call sites in TestVerificationGroundednessClassifier needed to change."
  - "Corpus floors (D-14-20) set to 420 documents / 520 commands -- a small buffer below the exact measured 427/533 -- so the floor still fails hard the moment the glob or extractor silently narrows, without making every single-document addition/removal a floor-breaking event. Verified the floor actually fails (naming both numbers) by temporarily raising it above the measured count and re-running."
  - "The frontier grew from 26 to 125 entries. This is NOT scope creep or an accident: it is the mechanical, by-design consequence of two decisions already ratified in 14-CONTEXT.md -- D-14-11's full .planning/**/*.md discovery (no document in the real corpus yet declares the `verification_role: proposal` exemption, so every RESEARCH/PLAN/CONTEXT/DISCUSSION-LOG/research-tree document defaults enforced) and D-14-16's R2b wiring. The literal's own doc comment records this attribution so the growth is auditable without re-deriving it."
  - "Read_first's citation of the archival-breakage grep row as being in '11-VALIDATION.md' does not match the verified live corpus: the actual currently-zero-match grep-over-the-roadmap row is `.planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md:113` (`grep -c \"S-008\" .planning/ROADMAP.md`), confirmed by direct execution (0 matches, exit 1) before writing any test against it. 11-VALIDATION.md:54's own roadmap-grep row (`flto|escalation, not a pass`) is NOT zero under whole-command R3 semantics -- the \"flto\" branch alone matches 10 times, so the row passes R3 even though its second alternation branch (`escalation, not a pass`) resolves to zero on its own; that is a per-branch (R2b) question, and R2b is explicitly scoped to go-test -run patterns only, not grep patterns, per the plan's own Task 2 action text. Used the real, verified 09-VALIDATION.md:113 finding as the test target instead of asserting a false claim to match the read_first's citation. Documented here rather than silently substituted."
  - "Two real bugs found and fixed while testing R3 (see Deviations): a double-quote/command-substitution gap in hasShellMetacharacterOutsideQuotes, and a self-referential search-pattern false positive in the argv-form-only self-check subtest."

patterns-established:
  - "Any future .planning/**-scanning Go test that needs to execute document-supplied command text directly (never through a shell) should reuse hasShellMetacharacterOutsideQuotes/classifyGrepGroundedness's argv-form-only shape rather than re-deriving shell-metacharacter detection from scratch -- especially the command-substitution-is-not-blocked-by-double-quotes nuance."

requirements-completed: [EVD-01]

coverage:
  - id: D1
    description: "Groundedness lint scopes documents by illocutionary role (D-14-11): an evidence basename is always enforced, a proposing document is exempt only via a positive frontmatter declaration, and a proposing document carrying a verdict token beside a command span is promoted to enforced -- renaming a file is never an escape."
    requirement: "EVD-01"
    verification:
      - kind: unit
        ref: "internal/compiler/session/verification_groundedness_test.go#TestVerificationGroundednessScopeByIllocutionaryRole"
        status: pass
    human_judgment: false
  - id: D2
    description: "Grep-shaped verification commands with a literal existing file operand and no shell metacharacter are executed directly (argv form, never a shell) and must yield >=1 match; a command carrying a pipe/redirect/conjunction/semicolon/command-substitution is never executed."
    requirement: "EVD-01"
    verification:
      - kind: unit
        ref: "internal/compiler/session/verification_groundedness_test.go#TestVerificationGroundednessGrepExecution"
        status: pass
      - kind: integration
        ref: "internal/compiler/session/verification_groundedness_test.go#TestVerificationGroundednessGrepOverRealCorpus"
        status: pass
    human_judgment: false
  - id: D3
    description: "Per-branch groundedness (R2b) is detected and pinned but not enforced to zero this phase, per D-14-16's explicit sizing decision."
    requirement: "EVD-01"
    verification:
      - kind: unit
        ref: "internal/compiler/session/verification_groundedness_test.go#TestVerificationGroundednessPerBranch"
        status: pass
    human_judgment: false
  - id: D4
    description: "Corpus floors (documents discovered, commands extracted) prevent the lint from silently going inert on an empty/narrowed corpus; the static test index is cross-checked for set equality against the toolchain's own whole-module test listing."
    requirement: "EVD-01"
    verification:
      - kind: unit
        ref: "internal/compiler/session/verification_groundedness_test.go#TestVerificationGroundednessCorpusIsNotEmpty"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/verification_groundedness_test.go#TestStaticTestIndexMatchesGoTestList"
        status: pass
    human_judgment: false
  - id: D5
    description: "The frontier literal equals the freshly measured violation set exactly (set equality, not a count/ceiling), re-measured over the whole enforced-tier corpus and re-pinned from 26 to 125 entries with the delta attributed."
    requirement: "EVD-01"
    verification:
      - kind: unit
        ref: "internal/compiler/session/verification_groundedness_test.go#TestVerificationGroundednessFrontierIsPinned"
        status: pass
    human_judgment: false

duration: 55min (first commit to last commit)
completed: 2026-09-17
status: complete
---

# Phase 14 Plan 06: Groundedness Lint Full Expansion Summary

**Extended the plan 14-01 groundedness lint with D-14-11 illocutionary-role document scoping, argv-form-only grep execution (R3), per-branch alternation groundedness (R2b), two corpus floors, a whole-module `go test -list` accuracy control, and a re-measured, re-pinned 125-entry exact violation frontier (up from 26).**

## Performance

- **Duration:** 55 min (commit-to-commit span)
- **Tasks:** 3
- **Files modified:** 1 (verification_groundedness_test.go, 976 -> ~1470 lines)

## Accomplishments

- **Task 1 (D-14-11 scope):** Tier A (asserting, enforced) vs Tier B (proposing, exempt) document classification. An evidence basename (VALIDATION/VERIFICATION/SUMMARY/UAT/MIDPHASE-GATE) is always enforced. A proposing basename (RESEARCH/PLAN/CONTEXT/DISCUSSION-LOG) or a document under `.planning/research/**` is exempt only via a positive `verification_role: proposal` frontmatter declaration -- never by filename accident, and exemption is opt-in only (no declaration defaults enforced). Any document -- regardless of basename or declaration -- whose table row carries both a command code-span and a verdict token in its own Status/Grade column is promoted to enforced, closing the rename-as-evasion path. Verdict-token detection is keyed on the Status/Grade COLUMN NAME (never a bare emoji scan of the row), confirmed against the real corpus that this correctly ignores a RESEARCH.md "File Exists?" ❌ cell, and confirmed to work identically whether the table still carries a Status column or has already migrated to a Grade column (the D-14-05 keying caveat this plan must not leave briefly undetectable).
- **Task 2 (R3 grep execution, R2b per-branch):** `classifyGrepGroundedness` executes a grep/rg-shaped command directly via `os/exec` (argv form, never a shell) when it has a literal existing file operand and no shell metacharacter, requiring >=1 match; zero matches or a missing file operand is a finding, never a skip. `hasShellMetacharacterOutsideQuotes` detects pipe/redirect/conjunction/semicolon/command-substitution outside quoted arguments. `classifyPerBranchGroundedness` splits a go-test `-run` pattern's first segment on top-level alternation and requires every branch to resolve, per D-14-16 -- detected and pinned, not enforced to zero this phase.
- **Task 3 (wiring, floors, accuracy control, re-pin):** Wired R2/R2b/R3 into `classifyCommand`, broadened production discovery from the narrow VALIDATION/VERIFICATION glob to the full D-14-11 enforced-tier scope, added two corpus floors (documents >= 420, commands >= 520, measured not copied), added `TestStaticTestIndexMatchesGoTestList` (D-14-10's accuracy control, 1066 names resolved on both sides, 5.4s), and re-measured/re-pinned the exact violation frontier: 125 entries, up from plan 14-01's 26, with the growth fully attributed in the literal's own doc comment.
- Verified against the real corpus: the R3 finding at `.planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md:113` (`grep -c "S-008" .planning/ROADMAP.md`, zero matches -- S-008 was archived out of the live ROADMAP.md) and 11 real R2b findings across `04-VERIFICATION.md`, `05-VERIFICATION.md`, and `09-VALIDATION.md`.
- Found and fixed two real bugs while testing (see Deviations).

## Task Commits

1. **Task 1: Scope by illocutionary role** -- `a116cb8` (test, RED) -> `fa0a7cc` (feat, GREEN)
2. **Task 2: Execute grep-shaped commands, detect per-branch groundedness** -- `9e445e8` (test, RED) -> `f8ddf2d` (feat, GREEN)
3. **Task 3: Corpus floors, accuracy control, re-pin the frontier** -- `4f4d773` (test, RED) -> `5b75b1e` (feat, GREEN)

_Note: all three tasks are `tdd="true"`. Each RED commit stubbed exactly the function(s) under test to a trivially-wrong return value (compiling cleanly, so the failure is a real assertion failure, never a compile error), confirmed the new subtests fail against the stub, then restored the real implementation for the GREEN commit._

## Files Created/Modified

- `internal/compiler/session/verification_groundedness_test.go` -- role-based scope classifier (Tier A/B + promotion), grep executor (R3), per-branch classifier (R2b), corpus floors, static-index accuracy control, re-pinned 125-entry frontier.

## Decisions Made

See `key-decisions` in frontmatter. Summary: (1) kept Task 1/2's new classifiers standalone until Task 3's wiring, to avoid breaking the frontier pin mid-task via the `-run` substring match on `TestVerificationGroundedness`; (2) R2 checked before R2b so the original 26 pinned classifications are unaffected; (3) `classifyCommand` reads `index.root` rather than adding a parameter, so no existing call sites changed; (4) corpus floors set with a small buffer below the exact measured values; (5) the 26->125 frontier growth is fully attributed to two already-ratified decisions (D-14-11 full-tree scope with no documents yet declaring the exemption, and D-14-16's R2b wiring), not scope creep; (6) the read_first's "phase-11" citation for the zero-match grep row does not match the verified corpus -- used the real, confirmed 09-VALIDATION.md:113 finding instead of asserting a false claim, and documented the discrepancy rather than silently substituting.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `hasShellMetacharacterOutsideQuotes` failed to detect command substitution inside double-quoted arguments**
- **Found during:** Task 2, first real test run after restoring the GREEN implementation
- **Issue:** The first implementation gated the `$(` check on the same `inSingle || inDouble` catch-all as the other four metacharacters (`|`, `>`, `;`, `&&`). POSIX double quotes suppress word-splitting and globbing but do NOT suppress command substitution -- only single quotes do. A command like `grep -c "$(cat match.txt)" match.txt` was wrongly treated as inert (no shell metacharacter), which would have let it execute a subprocess that itself embeds a subshell invocation.
- **Fix:** Restructured `hasShellMetacharacterOutsideQuotes` so the `$(` check runs independent of `inDouble`, gated only by `inSingle` -- command substitution is checked outside single quotes specifically, while the other four metacharacters remain checked outside both quote kinds.
- **Files modified:** `internal/compiler/session/verification_groundedness_test.go`
- **Verification:** `TestVerificationGroundednessGrepExecution/a_command_carrying_a_command_substitution_is_never_executed` failed before the fix (subprocess ran) and passes after.
- **Committed in:** `f8ddf2d` (Task 2 GREEN commit)

**2. [Rule 1 - Bug] The argv-form-only self-check subtest matched its own source line (false-positive count of 2, not 0)**
- **Found during:** Task 2, same test run
- **Issue:** The subtest proving this file never passes a command string to a shell wrote its search pattern (`sh", "-c"`) as one contiguous string literal directly in its own source and in its failure message -- so `grep -c` over the file's own source found 2 matches (the pattern literal itself and the message), not the expected 0, even though the file genuinely contains no shell invocation.
- **Fix:** Assembled the search pattern at runtime from parts (`strings.Join`) so the literal substring never appears contiguously in the file's own source text, and removed the literal from the failure message too (uses `%q` on the runtime value instead).
- **Files modified:** `internal/compiler/session/verification_groundedness_test.go`
- **Verification:** `grep -c 'sh", "-c"' internal/compiler/session/verification_groundedness_test.go` now prints `0`; the subtest passes.
- **Committed in:** `f8ddf2d` (Task 2 GREEN commit)

---

**Total deviations:** 2 auto-fixed (both Rule 1 bugs found via testing against real inputs, not against the corpus).
**Impact on plan:** Both fixes are necessary for R3's stated correctness property (never hands a document's command to a shell) and for the argv-form-only acceptance criterion to actually prove what it claims. No scope creep; no architectural change.

## Issues Encountered

None beyond the deviations above. The frontier's growth from 26 to 125 entries was anticipated in shape (D-14-11/D-14-16 both predict scope and classification growth) but its exact magnitude was not assumed -- measured, as D-14-19 requires, and recorded in the pinned literal's own doc comment for auditability.

## Threat Flags

None found beyond the plan's own `<threat_model>` (T-14-31 through T-14-36), which this plan's implementation addresses directly: T-14-31 (grep execution) via argv-form-only + the metacharacter guard + a subtest asserting no subprocess ran for a metacharacter-carrying command; T-14-32 (scope evasion by rename) via the promotion rule, independent of basename; T-14-33 (inert-by-empty-corpus) via the two corpus floors, proven to fail on a narrowed glob; T-14-34 (resolution mechanism drift) via `TestStaticTestIndexMatchesGoTestList`; T-14-35/T-14-36 not applicable by construction (no unbounded pipeline spawned, zero packages installed).

## User Setup Required

None -- no external service configuration required.

## Next Phase Readiness

- The groundedness lint is fully expanded, wired, green, and re-pinned at 125 entries. Every `<must_haves.truths>` item from the plan frontmatter is demonstrated by a passing subtest; every `<must_haves.prohibitions>` item is honored (no inline suppression channel exists anywhere in the file, no verification command is ever handed to a shell, and no finding can be silenced by renaming a file out of the enforced tier).
- Plan 14-09 (retires the freeform Status column) can proceed without a detection gap: enforced-tier detection already accepts either the Status column or the Grade column as a verdict signal.
- Plan 14-10 (Nyquist reconciliation) inherits the full 125-entry frontier -- its job is to drive the R1/R2/R2b/R3 classes toward empty (R2b stays pinned-not-enforced per D-14-16 until QLT-10) and to add `verification_role: proposal` frontmatter declarations to genuine RESEARCH/PLAN/CONTEXT/DISCUSSION-LOG proposals, which would shrink the frontier back down by removing the now-in-scope proposal-corpus duplication.
- No blockers for the next plan in this phase.

## Self-Check: PASSED

- `[ -f internal/compiler/session/verification_groundedness_test.go ]` -> FOUND
- Commits present in `git log --oneline --all`: `a116cb8`, `fa0a7cc`, `9e445e8`, `f8ddf2d`, `4f4d773`, `5b75b1e` -- all FOUND. The seventh, this plan's own metadata commit, is self-referential and cannot state its own SHA: verify it with `git log --oneline -1 -- .planning/phases/14-evidence-instrument-and-honest-scoping/14-06-SUMMARY.md`.
- `go test ./internal/compiler/session/... -run 'TestVerificationGroundedness' -count=1 -v` re-run: 9/9 subtests PASS (`TestVerificationGroundednessClassifier`, `TestVerificationGroundedness`, `TestVerificationGroundednessScopeByIllocutionaryRole`, `TestVerificationGroundednessFrontierIsPinned`, `TestVerificationGroundednessCorpusIsNotEmpty`, `TestVerificationGroundednessIsNotInert`, `TestVerificationGroundednessGrepExecution`, `TestVerificationGroundednessGrepOverRealCorpus`, `TestVerificationGroundednessPerBranch`)
- `go test ./internal/compiler/session/... -run 'TestStaticTestIndexMatchesGoTestList' -count=1 -v` re-run: PASS (1066 names resolved on both sides, 4.4s)
- `go build ./...` and `go vet ./...` re-run: both clean
- `go test ./...` re-run: exit 0, all 25 tested packages `ok` (session package 175.2s)
- `git status --short`: clean except the pre-existing untracked `.planning/milestone.lock`

---
*Phase: 14-evidence-instrument-and-honest-scoping*
*Completed: 2026-09-17*
