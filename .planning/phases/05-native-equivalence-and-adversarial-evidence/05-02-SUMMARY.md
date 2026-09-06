---
phase: 05-native-equivalence-and-adversarial-evidence
plan: 02
subsystem: compiler-checker
tags: [ownership, borrow-checking, liveness, dataflow, technical-debt-retirement]

requires:
  - phase: 05-native-equivalence-and-adversarial-evidence
    provides: "05-01's promoted core.LoanEndpoint facts (by-pointer lowering, D-04-33 Alias audit) and the Phase 1-4 byte-identity tripwire this plan's own regression tests build on"
provides:
  - "loanLivenessFixpoint widened to cover straight-line bodies AND branch arms via computeLoanLastUses, replacing discoverLoanLastUses as the sole liveness law"
  - "A recorded zero-divergence shadow-mode migration (225,890 + 4,802 admission-site comparisons) authorizing the retirement, per the Rust NLL migration procedure"
  - "core.LoanEndpoint promoted from decorative to load-bearing; corevalidate.recomputeLoanEndpoints confirmed independent (D-12) by a new source-scan test"
  - "D-04-26 and D-03-01 closed"
affects: [05-03, 05-04, 05-05, 05-06, 05-07]

actuals:
  tokens: 6595
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Shadow-mode (GitHub-Scientist style) law migration: run both laws side by side, log divergences, gate deletion on a recorded zero-divergence run over the FULL enumeration (never a merely-green suite) plus an observed-case-count floor"
    - "Historical Work-accounting formulas an independent oracle pins exactly are preserved via a separate, unpinned field (FixpointWork) rather than disturbed by an internal mechanism swap"
    - "Comment-exempt source-scan tests (TestSingleLoanLivenessLaw, TestLivenessLawsStayIndependent) that build the scanned identifier via string concatenation so the test's own source line cannot trip its own scan"

key-files:
  created: []
  modified:
    - internal/compiler/check/check.go
    - internal/compiler/check/check_test.go
    - internal/compiler/core/core.go

key-decisions:
  - "Task 1's widening call and Task 2's shadow-comparison apparatus were both DELETED in Task 3 once discoverLoanLastUses was retired -- the plan's own instruction ('no dead seam left behind') took precedence over leaving the machinery in place for possible future reuse. The zero-divergence proof is preserved in git history (Task 2's commit) and in this SUMMARY, not as a permanently-running regression test."
  - "Straight-line's `result.Work` keeps the historical flat `len(bindings)+1` accounting term for the last-use derivation's cost, decoupled from which mechanism (discoverLoanLastUses, then computeLoanLastUses) actually computes the map -- TestOwnershipSequenceExhaustive/TestOwnershipWorkSeries pin this field exactly against an independent oracle/hand-derived formula that has no way to reproduce loanLivenessFixpoint's own loan-chain-shape-dependent cost. The fixpoint's REAL cost is counted honestly via the separate FixpointWork field, folded into checkLinear's function-level RecomputedWork total -- not hidden, just kept out of the one field an external oracle already pins. Arm bodies have no such pin, so analyzeArmBody's own discoveryWork term is the fixpoint's real cost directly."
  - "TestLoanLivenessFixpointCoversStraightLine (Task 1) uses testdata/phase3/shared_shared_accept.lang, not a phase2 fixture as the plan's read_first suggested -- no ACCEPTED phase2 fixture carries a borrow (both phase2 borrow fixtures are REJECT controls: move_while_borrowed.lang, reborrow_while_moved.lang)."

patterns-established:
  - "A liveness/admission law retirement is authorized by an executable zero-divergence test over the FULL case enumeration with a case-count floor, never by a green suite alone -- the divergence population is DATA, not prose."

requirements-completed: [NAT-02]

coverage:
  - id: D1
    description: "loanLivenessFixpoint's domain covers straight-line bodies as well as branch arms (widened, not a second fixpoint entry point)"
    requirement: "NAT-02"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestLoanLivenessFixpointCoversStraightLine"
        status: pass
    human_judgment: false
  - id: D2
    description: "Both liveness laws ran side by side in shadow mode over the full ownership (225,890 cases) and branch (4,802 cases) enumeration with zero divergences, gated by an observed-case-count floor of 113,000"
    requirement: "NAT-02"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go (TestLoanLivenessShadowZeroDivergence, run and recorded clean in commit b580771; deleted in commit cc7d81f per the plan's own dead-seam-removal instruction once the retirement it authorized landed)"
        status: pass
    human_judgment: true
    rationale: "The authorizing test itself was intentionally deleted after use (Task 3's explicit instruction); its pass/fail state is now historical (git log), not re-verifiable by re-running the current test suite. A human reviewer should confirm the retained TestOwnershipSequenceExhaustive/TestBranchSequenceExhaustive (which still exercise the single remaining law against an independent oracle) are an acceptable ongoing substitute for the deleted shadow harness."
  - id: D3
    description: "discoverLoanLastUses deleted; computeLoanLastUses (loanLivenessFixpoint-based) is the sole law deciding admission in both checkLinear/analyzeStraightLine and checkBranch/analyzeArmBody; core.LoanEndpoint is load-bearing"
    requirement: "NAT-02"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestSingleLoanLivenessLaw"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestLivenessLawsStayIndependent"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestOwnershipSequenceExhaustive"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestBranchSequenceExhaustive"
        status: pass
    human_judgment: false
  - id: D4
    description: "No checker verdict and no Phase 1-4 core byte changed by the retirement"
    requirement: "NAT-02"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCheckerVerdictsUnchanged"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestPreviousPhaseCoreBytesUnchanged"
        status: pass
    human_judgment: false

duration: ~70min
completed: 2026-09-05
status: complete
---

# Phase 5 Plan 2: Retire discoverLoanLastUses Summary

**Widened loanLivenessFixpoint to cover straight-line bodies, shadow-ran it against discoverLoanLastUses over 230,692 admission-site comparisons with zero divergences, then deleted discoverLoanLastUses and made the fixpoint the sole liveness law everywhere in the checker.**

## Performance

- **Duration:** ~70 min
- **Tasks:** 3 completed
- **Files modified:** 3 (check.go, check_test.go, core.go)

## Accomplishments

- Widened `loanLivenessFixpoint` to also cover straight-line function bodies (the degenerate single-block CFG), constructed inside `analyzeStraightLine` itself rather than adding a second entry point. Added `TestLoanLivenessFixpointCoversStraightLine` proving the widened fixpoint produces endpoints for a straight-line accepted fixture with borrows where production's serialized core carries none (byte-identity for straight-line functions is preserved -- `LinearBody.LoanEndpoints` still stays absent there).
- Ran `discoverLoanLastUses` (the old forward chain-inheritance law) and `loanLivenessFixpoint` (the candidate law, via a new `candidateLoanUses` helper that builds a synthetic single-block operation stream directly from the AST body) side by side in shadow mode over the COMPLETE `TestOwnershipSequenceExhaustive` (225,890 comparisons across both the shareable and non-shareable type sweeps) and `TestBranchSequenceExhaustive` (4,802 comparisons across both arms of 2,401 two-block programs) enumerations. Zero divergences recorded, well above the 113,000-case floor. Demonstrated the harness itself catches a seeded divergence (and clears again once restored) via a pure, non-`t.Run`-based assertion function (`t.Run` failures propagate to the whole package's exit code, which would have made the falsifying demonstration itself fail CI).
- Deleted `discoverLoanLastUses` entirely, repointed both admission sites (`analyzeStraightLine`, `analyzeArmBody`) onto `computeLoanLastUses` (the renamed candidate law, now the sole law), deleted the now-dead shadow-mode apparatus (recorder type, package var, comparison function, both shadow tests -- "no dead seam left behind"), and promoted `core.LoanEndpoint`'s doc comment from decorative to load-bearing. Added `TestSingleLoanLivenessLaw` (source-scan proof the retired identifier no longer appears as a live reference) and `TestLivenessLawsStayIndependent` (proves `corevalidate.go` still neither imports `compiler/check` nor calls any of its liveness helpers outside a comment, preserving D-12's two-independent-derivations invariant). D-04-26 and D-03-01 are hereby closed.

## Task Commits

Each task was committed atomically:

1. **Task 1: Broaden loanLivenessFixpoint to the full straight-line-plus-branch domain** - `c5d0cb9` (feat)
2. **Task 2: Shadow-run both liveness laws over the full enumeration and record the divergence population** - `b580771` (feat)
3. **Task 3: Delete discoverLoanLastUses, repoint admission onto the fixpoint, promote core.LoanEndpoint** - `cc7d81f` (feat)

## Files Created/Modified

- `internal/compiler/check/check.go` - `computeLoanLastUses` (sole liveness law), `FixpointWork` field, admission sites repointed, `discoverLoanLastUses` and the shadow-mode apparatus deleted
- `internal/compiler/check/check_test.go` - `TestLoanLivenessFixpointCoversStraightLine`, `TestSingleLoanLivenessLaw`, `TestLivenessLawsStayIndependent`; `TestLastUseDiscoveryWorkIsCounted`/`TestLastUseDiscoveryWorkSeries` repointed to `computeLoanLastUses`; the Task 2 shadow-mode tests added then removed per Task 3
- `internal/compiler/core/core.go` - `core.LoanEndpoint`'s doc comment promoted from decorative/report-only to load-bearing

## Decisions Made

See `key-decisions` in frontmatter: the FixpointWork/Work-accounting split (preserves an independent oracle's exact pin while still counting the real fixpoint cost honestly), deleting the shadow apparatus in Task 3 rather than keeping it as a permanent regression harness (per the plan's own "no dead seam left behind" instruction), and using `testdata/phase3/shared_shared_accept.lang` instead of a phase2 fixture for the Task 1 coverage test (no accepted phase2 fixture carries a borrow).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug/Design] Isolated FixpointWork from ownershipSupport.Work rather than folding the widened fixpoint's cost directly in**
- **Found during:** Task 1, while implementing the straight-line widening
- **Issue:** The plan's action text ("count the fixpoint's straight-line work into the existing checker RecomputedWork accounting") is ambiguous between "fold it into `ownershipSupport.Work`" and "fold it into the function-level RecomputedWork total (as checkBranch already does for its own fixpoint call)". `ownershipSupport.Work` returned by `analyzeStraightLine` is compared field-by-field, EXACTLY, against an independent oracle (`oracleStraightLine`) by `TestOwnershipSequenceExhaustive` and against a hand-derived formula by `TestOwnershipWorkSeries` -- both required to stay green by this plan's own verify commands. Folding the fixpoint's real, loan-chain-shape-dependent cost directly into `.Work` would have required deriving a matching closed-form oracle formula for arbitrary loan-chain shapes across ~113,000+ generated cases, which is not achievable without literally reimplementing `blockLoanLiveness`'s own internal walk inside the oracle (defeating the point of an independent oracle).
- **Fix:** Added a separate `FixpointWork` field to `ownershipSupport` (not compared by `assertSupportEqual`), folded into `checkLinear`'s own function-level `RecomputedWork` return value -- mirroring exactly where checkBranch already folds its own `loanLivenessFixpoint` call's work (into its own function-level `work`, never into `analyzeArmBody`'s returned `ownershipSupport.Work`). `result.Work` itself keeps its pre-existing formula unchanged throughout all three tasks.
- **Files modified:** `internal/compiler/check/check.go`
- **Verification:** `TestOwnershipSequenceExhaustive`, `TestOwnershipWorkSeries`, `TestBranchSequenceExhaustive` all green after every task; the real fixpoint cost is nonzero and counted (verified via `TestLoanLivenessFixpointCoversStraightLine`'s own nonzero-work assertion in Task 1, and via `checkLinear`'s return value).
- **Committed in:** `c5d0cb9` (Task 1), carried through `cc7d81f` (Task 3)

---

**Total deviations:** 1 auto-fixed (1 design clarification, made necessary by an independent-oracle pin the plan's own required verify commands protect). No scope creep -- the fixpoint's cost is honestly counted, just not in the one field an external oracle already pins exactly.

## Issues Encountered

- The plan's own final `<verification>` block names `TestLoanLivenessShadowZeroDivergence`, which Task 3's own instructions require deleting ("delete the shadow recorder's dual-computation call sites... keeping the recorder type itself only if some test still needs it; otherwise delete it too. No dead seam is left behind"). This is a genuine internal inconsistency in the plan (the final verification block was evidently not updated after Task 3's delete instruction was finalized). Resolved by running the plan's verification with that one identifier omitted (all other named tests pass); the zero-divergence proof itself is preserved in git history at commit `b580771` and recorded in this SUMMARY's `coverage` block (D2), rather than as a permanently-running test.
- `TestDebtRegistersAreWellFormed` (`internal/compiler/session/session_test.go`) fails on the unmodified tree with `04-DEBT.md: frontmatter declares items: 4 but the Items table holds 5 rows` -- confirmed pre-existing in 05-01-SUMMARY.md (reproduced via `git stash` there, before any of this phase's changes) and unrelated to any file this plan touches. Out of scope per the deviation rules' scope boundary; not fixed.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Exactly one liveness law (`loanLivenessFixpoint`/`computeLoanLastUses`) decides admission everywhere in the checker; `core.LoanEndpoint` is load-bearing. D-04-26 and D-03-01 are closed -- `check.go` is no longer landing a new `OperationKind` (D-05-32 keeps `OpCall` out) and is not itself carrying a second liveness law, so the rest of Phase 5's alias-fact, by-pointer, and corevalidate work (05-04+) builds against a single, honest law.
- `corevalidate.recomputeLoanEndpoints` remains a confirmed-independent re-derivation (D-12), now covered by an explicit regression test (`TestLivenessLawsStayIndependent`) rather than only a documented invariant.
- No blockers.

---
*Phase: 05-native-equivalence-and-adversarial-evidence*
*Completed: 2026-09-05*

## Self-Check: PASSED

All 3 modified files verified present on disk with expected content; all 3 task commit hashes (`c5d0cb9`, `b580771`, `cc7d81f`) verified in `git log`. Plan-level `<verification>` block re-run with the one now-deleted-by-design test identifier omitted (`TestLoanLivenessShadowZeroDivergence` -- see Issues Encountered): all remaining named tests pass (`TestSingleLoanLivenessLaw`, `TestLivenessLawsStayIndependent`, `TestOwnershipSequenceExhaustive`, `TestBranchSequenceExhaustive`, `TestCheckerVerdictsUnchanged`, `TestPreviousPhaseCoreBytesUnchanged`). `go build ./...`, `go vet ./...`, `go test ./...`, and `go test -race ./...` all clean except the confirmed pre-existing `TestDebtRegistersAreWellFormed` failure (unrelated, documented in 05-01-SUMMARY.md).
