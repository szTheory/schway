---
phase: 03-borrowed-views-and-cfg-lifetimes
plan: "05"
subsystem: compiler-frontend
tags: [ownership, cfg, loan-liveness, path-oracle, mutation-kill, property-testing, mid-phase-gate]

requires:
  - phase: 03-borrowed-views-and-cfg-lifetimes
    provides: "03-01's real CFG (core.Block/Edge/LoanEndpoint, checkBranch), 03-02's exclusive loans and conflict matrix, 03-03's backward-worklist loanLivenessFixpoint/materializeLoanEndpoints (checkBranch's sole LoanEndpoint producer), and 03-04's independent reachability-closure recomputeLoanEndpoints in corevalidate"
provides:
  - "internal/compiler/pathoracle: a bounded, exhaustive acyclic-path-enumeration third decision procedure for loan endpoints, imported by no production package, agreeing with production LoanEndpoints by whole-value equality on real branch fixtures"
  - "A recorded revert-and-fail demonstration (D-09): neutralizing checkBranch's materializeLoanEndpoints call makes the oracle differential, the uniform-join-fault test, and the verify lane all fail"
  - "A seeded uniform-join-fault endpoint-level disagreement, naming the specific loan"
  - "A path-count cap (MaxPaths=4096) that rejects fail-closed rather than truncating"
  - "session.PathOracleDisagreementLane wired into the verify path as control:cfg.path_oracle_disagreement with nonzero work"
  - "Generators extended to reach match-arm bodies, the exclusive-borrow spelling, a zero-binding arm, an in-arm reborrow, and a comment trailing an arm's own header line, with the reachable AND unreachable space stated explicitly"
  - "The phase's mandatory mid-phase gate: both open items (the checker's two liveness derivations; 03-06's dropped origin gate) adjudicated with code-level evidence and recorded as dated debt in 03-DEBT.md"
affects: [03-07-debug-lineage-and-phase-close]

actuals:
  tokens: 17320
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Exhaustive acyclic path enumeration (pathoracle.EnumeratePaths) as a third mechanism class, distinct from both check.go's backward-worklist fixpoint and corevalidate.go's reachability-closure-plus-reduction: concrete per-path replay combined only at the end, never an abstract per-block summary and never a single global relation closure"
    - "Per-path forward transitive place->loan inheritance (pathoracle.linearizePath), independently re-deriving discoverLoanLastUses' law over core.LinearOperation place IDs rather than ast.Binding names, since the oracle must not import check"
    - "Real declared core.Edge IDs read directly from the artifact for edge-endpoint identity, rather than reimplementing production's ID-formatting formula, so whole-value equality holds without coupling to an implementation-internal string convention"
    - "Endpoint-level fault construction (TestOracleDisagreesWithUniformJoinFault) when an existing test-only admission seam (testOnlyForceUniformLoanJoin) provably cannot reach the fact under test"

key-files:
  created:
    - internal/compiler/pathoracle/pathoracle.go
    - internal/compiler/pathoracle/pathoracle_test.go
    - .planning/phases/03-borrowed-views-and-cfg-lifetimes/03-DEBT.md
  modified:
    - internal/compiler/check/check_test.go
    - internal/compiler/syntax/syntax_test.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_test.go

key-decisions:
  - "The path oracle reads only function.Linear.Blocks/Edges/Operations and returns nil for a straight-line function (no Blocks) — mirroring production's own scope, since checkBranch is the sole producer of observable LoanEndpoint records. This is a deliberate scope match, not an evasion: it means the oracle differential compares exactly the artifact production claims to produce, never manufacturing a disagreement against a fact production never asserts."
  - "TestOracleDisagreesWithUniformJoinFault does NOT toggle testOnlyForceUniformLoanJoin to produce its differential, because that seam only perturbs analyzeArmBody's admission-time discoverLoanLastUses consumer — never loanLivenessFixpoint/materializeLoanEndpoints, which run afterward on the already-lowered operations and are checkBranch's sole LoanEndpoint producer. The test instead reconfirms the seam's admission-level effect (matching TestUniformJoinPlacementFlipsBothVerdicts) and separately constructs, at the endpoint level, what uniform-join placement would have produced had it also corrupted materializeLoanEndpoints, then proves the oracle disagrees with that construction."
  - "generatedOwnershipBodyBytes' alphabet grew from 3 kinds (36 symbols) to 4 kinds (48 symbols) to add borrow_mut, and the check_test.go oracle (oracleStraightLine) gained an independently-derived five-row conflict-matrix check (oracleConflictingLoan) so TestOwnershipSequenceExhaustive's ~113,164-case differential now covers shared/exclusive overlap, not just single-loan sequences — this was necessary, not cosmetic: before this change no generated case could ever create two overlapping loans of different access modes on the same owner, so a conflict-matrix defect could have hidden behind this specific green differential indefinitely."
  - "TestBranchSequenceExhaustive builds its 2,401 two-block programs directly against ast.Program structs (bypassing the parser), and derives its independent verdict by applying the EXISTING oracleStraightLine differential once per arm — legitimate only because 03-01's per-arm aliasing makes every arm's admission decision structurally independent of its sibling, a fact this test's own 100% pass rate (on the first run, no fix needed) independently reconfirms rather than assumes."
  - "The mid-phase gate's two open items are adjudicated as real, non-blocking, dated debt (03-DEBT.md D-03-01, D-03-02), not silently accepted and not fixed within this plan's scope, matching the phase's own carried-debt convention (02-DEBT.md) and the gate's explicit instruction not to widen scope into rewiring checkLinear without evidence of a wrong verdict."

requirements-completed: []

coverage:
  - id: D1
    description: "A bounded acyclic path oracle independently enumerates every acyclic entry-to-return path, linearizes each, and decides loan endpoints without calling any production liveness code; agrees with production by whole-value equality on real branch fixtures"
    requirement: OWN-03
    verification:
      - {kind: unit, ref: "internal/compiler/pathoracle#TestOracleEnumeratesAllAcyclicPaths", status: pass}
      - {kind: unit, ref: "internal/compiler/pathoracle#TestOracleImportsStayIndependent", status: pass}
      - {kind: unit, ref: "internal/compiler/pathoracle#TestOracleAgreesWithProduction", status: pass}
      - {kind: unit, ref: "internal/compiler/pathoracle#TestOracleFailsOnUnterminatedLoan", status: pass}
    human_judgment: false
  - id: D2
    description: "The differential is demonstrated load-bearing by a recorded revert-and-fail (D-09), a seeded uniform-join fault at the endpoint level, a fixed-seed metamorphic trial batch, and a fail-closed path-count cap; the verify harness runs it as a lane with nonzero work"
    requirement: OWN-03
    verification:
      - {kind: unit, ref: "internal/compiler/check#TestOracleDisagreesWithUniformJoinFault", status: pass}
      - {kind: unit, ref: "internal/compiler/check#TestOracleMetamorphicTrials", status: pass}
      - {kind: unit, ref: "internal/compiler/pathoracle#TestOraclePathCountCapRejects", status: pass}
      - {kind: unit, ref: "internal/compiler/session#TestPathOracleLaneRecordsControl", status: pass}
    human_judgment: false
  - id: D3
    description: "The property lanes reach match-arm bodies, the exclusive-borrow spelling, a zero-binding arm, an in-arm reborrow, and a comment trailing an arm's own header line; the reachable and unreachable space is stated explicitly"
    requirement: OWN-03
    verification:
      - {kind: unit, ref: "internal/compiler/syntax#TestGeneratedBranchBodiesRoundTrip", status: pass}
      - {kind: unit, ref: "internal/compiler/syntax#TestGeneratedKindsIncludeExclusiveBorrow", status: pass}
      - {kind: unit, ref: "internal/compiler/check#TestBranchSequenceExhaustive", status: pass}
      - {kind: unit, ref: "internal/compiler/check#FuzzOwnershipLinear", status: pass}
    human_judgment: false
  - id: D4
    description: "Mid-phase gate: both open items (the checker's two liveness derivations; 03-06's dropped origin gate) adjudicated with code-level evidence, using the path oracle as the independent instrument, and recorded as dated debt"
    requirement: OWN-03
    verification: []
    human_judgment: true
    rationale: "This is a review-and-judgment deliverable, not a test assertion. The evidence (code reads, a live revert-and-fail run, a live neutralize-and-observe run against 03-DEBT.md's two concrete instances) is recorded verbatim in this summary and in 03-DEBT.md for a human to confirm the adjudication before wave 6 (03-07) proceeds."

duration: 95min
completed: 2026-09-04
status: complete
---

# Phase 03 Plan 05: Bounded Path Oracle, Mutation Kill, and the Mid-Phase Gate Summary

**A bounded, exhaustive-path-enumeration oracle independently agrees with production's loan-endpoint dataflow, is proven load-bearing by a recorded revert-and-fail and an endpoint-level seeded fault, and the phase's mandatory mid-phase gate closes with both open items adjudicated from code-level evidence and two new debt items recorded rather than silently assumed safe.**

## Performance

- **Duration:** ~95 min
- **Started:** 2026-09-04 (session start)
- **Completed:** 2026-09-04
- **Tasks:** 3 completed
- **Files modified:** 7 (3 created, 4 modified — plus this SUMMARY and 03-DEBT.md)

## Accomplishments

- Built `internal/compiler/pathoracle`: `EnumeratePaths` exhaustively walks
  every acyclic block-to-block path from a function's declared entry block
  to a terminal (no-successor) block, in the exact successor order the core
  artifact declares. `linearizePath` replays one concrete path forward,
  independently re-deriving the retained straight-line last-use law
  (forward transitive place->loan inheritance) over `core.LinearOperation`
  place IDs — a genuinely different mechanism from both check.go's backward
  worklist fixpoint (abstract per-block live-sets, converged) and
  corevalidate.go's reachability-closure-plus-reduction (one global
  materialized relation, one BFS closure): no shared helper, no queue, no
  relation closure anywhere in this package.
- `RecomputeEndpoints` combines per-path answers into `core.LoanEndpoint`
  records, reading real declared `core.Edge` IDs from the artifact for edge
  endpoints (never reimplementing production's ID-formatting formula), and
  agrees with production's materialized `LoanEndpoints` by whole-value
  equality on `branch_one_arm_shared_accept.lang`
  (`TestOracleAgreesWithProduction`).
- A late terminal guard (`unterminatedLoanError`) fires when a path carries
  a loan but never reaches `OpReturn` — a hard failure rather than a
  silent default to the loan's own creation site
  (`TestOracleFailsOnUnterminatedLoan`), and a fail-closed path-count cap
  (`MaxPaths=4096`, well above the language's real ceiling of 64 arms)
  rejects an over-cap synthetic function rather than truncating
  (`TestOraclePathCountCapRejects`).
- **Recorded revert-and-fail demonstration (D-09), performed live this
  session:** `checkBranch`'s `linear.LoanEndpoints = materializeLoanEndpoints(...)`
  line was temporarily replaced with `linear.LoanEndpoints = nil`, and all
  four differentials that depend on it failed exactly as required —
  captured verbatim below — then the change was reverted with a
  byte-identical `git diff` confirming zero residual drift.
- `TestOracleDisagreesWithUniformJoinFault` first reconfirms (independently
  of `TestUniformJoinPlacementFlipsBothVerdicts`) that 03-03's
  `testOnlyForceUniformLoanJoin` seam only perturbs admission-time
  bookkeeping, never `materializeLoanEndpoints` — then constructs, at the
  endpoint level, what uniform-join placement would have produced had it
  also corrupted endpoint materialization, and proves the oracle
  independently disagrees, naming the specific loan.
- `TestOracleMetamorphicTrials` runs a fixed 64-trial batch reordering two
  independent shared-loan bindings and alpha-renaming their places per
  trial, asserting oracle/production agreement on every trial and its own
  trial count.
- `session.PathOracleDisagreementLane` wires the oracle into the verify
  path as `lane:path-oracle-disagreement`, recording
  `control:cfg.path_oracle_disagreement` with nonzero work, using the
  Phase 1 `addLane` shape (explicit status on every path).
- Extended `linearGeneratedKinds` with `"borrow mut "`
  (`TestGeneratedKindsIncludeExclusiveBorrow` proves it is actually
  reached, not merely listed) and added a new `generatedBranchProgram`
  generator (`TestGeneratedBranchBodiesRoundTrip`) reaching arm bodies with
  a zero-binding arm, an in-arm reborrow, the exclusive spelling, and a
  comment trailing an arm's own header line — with a doc comment stating
  the reachable AND unreachable space explicitly (no loop/back edge, no
  mixed bare/body match, no generic-typed match scrutinee — the grammar
  forbids the last one entirely).
- `TestBranchSequenceExhaustive` extends the exhaustive ownership-sequence
  differential from one block to a two-block branch at bounded length
  (2,401 deterministic two-block programs, count-asserted), checked
  through the real `Program(...)` entry point against an independently
  computed per-arm verdict.

## Task Commits

1. **Task 03-05-01: Bounded acyclic path enumeration as a third decision procedure** — `f6c663b` (feat)
2. **Task 03-05-02: Mutation-kill the oracle and cap the enumeration (D-09)** — `4179320` (test)
3. **Task 03-05-03: Make the generators reach branching, exclusive, and cross-arm shapes (D-10)** — `68ed4f7` (test)

**Plan metadata:** (this commit)

## Files Created/Modified

- `internal/compiler/pathoracle/pathoracle.go` — the oracle package
- `internal/compiler/pathoracle/pathoracle_test.go` — enumeration, import-independence, agreement, cap, unterminated-loan tests
- `internal/compiler/check/check_test.go` — exclusive-borrow-aware oracle, uniform-join endpoint disagreement, metamorphic trials, two-block exhaustive differential
- `internal/compiler/syntax/syntax_test.go` — exclusive-borrow-reaching kind, branch-body generator and round-trip test
- `internal/compiler/session/session.go` — `PathOracleDisagreementLane`
- `internal/compiler/session/session_test.go` — lane control test
- `.planning/phases/03-borrowed-views-and-cfg-lifetimes/03-DEBT.md` — mid-phase gate's two carried debt items

## Decisions Made

See `key-decisions` in frontmatter.

## Deviations from Plan

None — plan executed as written. (The mid-phase gate's own two open items
were adjudicated as required rather than "fixed," per the plan's explicit
instruction not to widen scope into rewiring `checkLinear` without evidence
of a wrong verdict; see the Mid-Phase Gate section below and 03-DEBT.md.)

## Issues Encountered

- The first `RecomputeEndpoints` implementation had a real bug, caught by
  `TestOracleMetamorphicTrials` before any commit: `neededBeyond` marking
  started at index 0 of a path's block sequence rather than at the loan's
  own birth-block index, so a loan born strictly inside a later block was
  incorrectly marked "needed beyond" an EARLIER, unrelated block (the
  shared `entry` block), producing spurious edge endpoints alongside the
  correct point endpoint. Fixed by scoping the `neededBeyond` marking loop
  to `[birthIndex, deathIndex)` instead of `[0, deathIndex)`. Verified: the
  fix was written, then all pathoracle and check package tests were
  re-run and passed; no broken version was ever committed.

## User Setup Required

None — no external service configuration required.

## Mutation-Kill / Non-Regression Evidence (recorded verbatim per D-09)

**Revert-and-fail, performed live this session.** `checkBranch`'s
`linear.LoanEndpoints = materializeLoanEndpoints(functionID, cfgBlocks, edgeIDLookup, fixpoint)`
line (check.go:352) was temporarily replaced with
`linear.LoanEndpoints = nil` (edgeIDLookup discarded via `_ =` to keep the
build green), and the following four failures were captured verbatim:

```
=== RUN   TestOracleAgreesWithProduction
    pathoracle_test.go:120: branch_one_arm_shared_accept.lang: oracle disagrees with production
         got:  [{ID:...loan:1 ... Kind:point BlockID:...block:arm:0 ...}]
         want: []
--- FAIL: TestOracleAgreesWithProduction (0.00s)
```

```
=== RUN   TestPathOracleLaneRecordsControl
    session_test.go:937: expected the honest, loan-bearing program to drive the control to pass:
    {... ID:lane:path-oracle-disagreement Status:fail Controls:[] RecomputedWork:0 ...}
--- FAIL: TestPathOracleLaneRecordsControl (0.00s)
```

```
=== RUN   TestOracleDisagreesWithUniformJoinFault
    check_test.go:339: fixture carries no LoanEndpoints to corrupt: &{... LoanEndpoints:[]}
--- FAIL: TestOracleDisagreesWithUniformJoinFault (0.00s)
```

```
=== RUN   TestOracleMetamorphicTrials
    check_test.go:485: seed=0: want exactly 2 loan endpoints (both independent shared loans), got []
--- FAIL: TestOracleMetamorphicTrials (0.00s)
```

The change was then reverted (`mv check.go.bak check.go`) and confirmed
byte-identical: `git diff internal/compiler/check/check.go` produced no
output, and the full suite (`go build ./...`, `go test ./...`) was green
again before any further work.

**Seeded uniform-join fault (endpoint level), recorded verbatim.**
`TestOracleDisagreesWithUniformJoinFault` constructs the endpoint set
uniform-join placement would have produced (the fixture's real point
endpoint for `s1:owned.branch_one_arm_shared_accept:fn:choose:loan:1`,
replaced by an edge endpoint on `...:edge:entry:arm:0`) and confirms
`pathoracle.RecomputeEndpoints` disagrees, naming loan `...:loan:1` by both
its real point endpoint and the corrupted edge claim. It also
independently reconfirms `testOnlyForceUniformLoanJoin`'s ADMISSION-level
effect (the fixture's diagnostics go from 0 to non-empty when the seam is
engaged), matching `TestUniformJoinPlacementFlipsBothVerdicts`.

**Exhaustive differentials, extended this plan.**
`TestOwnershipSequenceExhaustive` (48-symbol alphabet, ~113,164 cases,
extended to include `borrow_mut` and an independently-derived conflict
matrix) and `TestBranchSequenceExhaustive` (2,401 two-block programs) both
pass, proving `discoverLoanLastUses` — the law that actually decides
admission in both `checkLinear` and `checkBranch` (see Mid-Phase Gate
below) — agrees with two independent oracle implementations across every
alphabet-reachable shape.

**Non-regression.**
`git diff --stat HEAD~3..HEAD -- testdata/phase1/ testdata/phase2/` is
empty — no Phase 1/2 golden moved.
`sh scripts/verify-phase2.sh` exits 0 and reports all Phase 1/2 controls
through a freshly built binary.
`env GOCACHE=/tmp/ai-lang-phase3-cache go test ./...`,
`go test -race ./...`, and `go vet ./...` all pass with zero findings.

## Known Stubs

None.

## Mid-Phase Gate (mandatory — ROADMAP §Phase 3, "Decision, 2026-09-04")

This plan is the OWN-03 terminal and carries the phase's mandatory
mid-phase gate. Both open items named in the phase handoff are adjudicated
below with code-level evidence, using the path oracle built in this plan
as the independent instrument the gate asked for.

### Open item 1 — the checker's two liveness derivations

**The premise needed correcting first.** Reading `check.go` directly (not
assuming from prior summaries) shows the two derivations do not compete to
answer the SAME question. `discoverLoanLastUses` (Phase 2, unchanged) is
read by BOTH `analyzeStraightLine` (`checkLinear`) and `analyzeArmBody`
(`checkBranch`'s per-arm helper) — its `loanUses[index].index` feeds
`loan.lastUse` directly, which drives `conflictingLoan`/`expiringLoans`:
the actual accept/reject decision, in every function the checker admits,
straight-line or branch. `loanLivenessFixpoint`/`materializeLoanEndpoints`
(03-03's new dataflow) runs ONLY inside `checkBranch`, AFTER
`analyzeArmBody` has already returned a verdict — its sole output,
`linear.LoanEndpoints`, is never read by any admission-deciding code
anywhere in the checker. So "do the two derivations agree" is not the
right question — only one of them ever produces a verdict.

**(a) Do the checker's two derivations agree on every reachable program
shape?** Reframed correctly: is `discoverLoanLastUses` — the single law
that actually decides every verdict — CORRECT on every reachable shape?
Using the path oracle as the independent instrument this session (not
assuming 03-04's "no new disagreement found" note is sufficient):

- `TestOwnershipSequenceExhaustive`, extended this plan from 36 to 48
  symbols specifically to add the exclusive-borrow spelling and an
  independently-derived five-row conflict matrix, re-verifies
  `discoverLoanLastUses`'s straight-line law against a materialized-edges-
  plus-fixed-point-closure oracle across ~113,164 cases — a genuinely
  different mechanism from the transitive-propagation law under test.
- `TestBranchSequenceExhaustive` (new this plan) drives 2,401 synthetic
  two-block branch programs through the real `Program(...)` entry point
  and independently re-derives each arm's own verdict via the SAME oracle
  applied per arm — legitimate because 03-01's per-arm aliasing makes each
  arm's admission fully independent of its sibling, which this test's
  100%-pass-on-first-run result reconfirms rather than assumes.
- No disagreement was found in either exhaustive sweep.

**(b) Does OWN-03 hold for straight-line programs under the old
propagation, or only for branch-bodied ones?** It holds for both, for a
structural reason, not merely an empirical one: a straight-line function's
CFG is a single block by construction — there is no branch, hence no
divergent edge to place an endpoint on, so "edge-specific last use" is
vacuously satisfied (every ending is necessarily a point ending, since
there is only one path). 03-03's own `TestStraightLineEndpointsUnchanged`
already proves this concretely: it runs `loanLivenessFixpoint` (the CFG
dataflow) on a single synthetic block built from a real straight-line
body's actual operations and confirms it reproduces
`discoverLoanLastUses`'s answer exactly — the two derivations are
mathematically forced to agree on a degenerate one-block CFG, and this
session's exhaustive differential (a) independently reconfirms
`discoverLoanLastUses` itself is correct on that shape.

**(c) Is the split defensible, or is it debt?** Both — narrowly. The split
between "the law that decides" (`discoverLoanLastUses`, unchanged,
correct, independently re-verified this session) and "the law that
reports facts" (`loanLivenessFixpoint`, correct, but decorative for
admission purposes) is a DEFENSIBLE end state for correctness: no wrong
verdict was found, and none is structurally possible for straight-line
programs. It is NOT a defensible end state for cost: D-05 explicitly
required OWN-03's CFG liveness to "remove the quadratic factor [D-02-03]
by construction" and make `recomputed_work` count the propagation
honestly. Neither happened on the path that actually matters —
`discoverLoanLastUses` remains the potentially-quadratic law deciding
every verdict, and its own propagation work is still never counted
(confirmed by inspection: every `result.Work++` in `analyzeStraightLine`
and `analyzeArmBody` sits in the surrounding loop, none inside
`discoverLoanLastUses` itself). This is recorded as **D-03-01** in
`03-DEBT.md`, not fixed here, per the plan's explicit instruction not to
widen scope into rewiring `checkLinear` without evidence of a wrong
verdict — there is none.

### Open item 2 — 03-06's dropped straight-line origin gate

**Confirmed by reading `originvalidate.go` directly.**
`ValidatePublished` begins `if function.PublicOrigin == nil { continue }`
— it never calls `RecomputeOrigin` (the independent, body-derived
recomputation) for a function that declares no origin at all.
`RecomputeOrigin` only ever CHECKS a declaration that already exists; it
is never used to DETECT that one is missing. `BuildInterface` then copies
`PublicOrigin` (nil) straight into the exported `FunctionSignature`.

**Is there a reachable program that reaches a public borrowed return with
no verified origin?** Yes, concretely:
`internal/compiler/check/check_exclusive_test.go`'s `exclusive_borrow_clean`
fixture exports `relay(buffer: Buffer) -> Buffer` returning an exclusively-
borrow-derived value with NO `borrow(path)` annotation, and checks with
zero diagnostics. `interface export` on this exact module would produce a
`FunctionSignature{PublicOrigin: nil, Abilities: [drop, share, send, escape]}`
for `relay` — structurally identical to a function returning a brand-new,
unrelated, fully-owned `Buffer`.

**Does the parse-time grammar rule have the same reach as the dropped
gate?** No — they enforce different things. The grammar rule constrains
the SPELLING of an annotation once one is written (`borrow` in
return-type position requires `(path)`); it does nothing to require that
anyone write one at all. A borrow-derived return with a plain
(unannotated) return type was never within the grammar rule's reach —
only the dropped `check.go` gate could have caught it, and 03-06 removed
that gate entirely (rather than narrowing it) specifically to protect
`exclusive_borrow_clean`'s own already-shipped acceptance. The
substitution 03-06's summary claimed ("mandatory annotation is now
enforced grammatically") is incomplete: it is mandatory only for whoever
chooses to write one.

**Why not blocking today.** This language has no cross-function call
construct yet, so nothing ever invokes one Lang function from another and
could act on the misleading exported signature — the gap is real in the
`interface export` ARTIFACT but currently inert as an executable
unsoundness. Recorded as **D-03-02** in `03-DEBT.md`, with a scoped fix
recommendation (an `interface export`/`ValidatePublished`-time check,
unconditional on `PublicOrigin != nil`, distinct from `check.go`'s own
admission path so `lang check` keeps accepting `exclusive_borrow_clean`
exactly as today) — before Phase 4 introduces calls.

### Gate verdict

**Clean, with two items recorded as dated, non-blocking debt** (03-DEBT.md
D-03-01, D-03-02). No wrong verdict was found in either sweep; the checker's
admission law is single-sourced and independently re-verified at both the
straight-line and branch level this session. Wave 6 (03-07) may proceed;
03-07 should read 03-DEBT.md alongside 02-DEBT.md when assembling the
phase's carried-debt closure.

## Next Phase Readiness

- OWN-03's four required truths (ROADMAP criteria 1-2) are independently
  re-verified by a third, mechanically different mechanism, load-bearing
  by a recorded revert-and-fail.
- The mid-phase gate is closed; 03-07 has both open items adjudicated with
  evidence and two concrete debt items to fold into the phase's final
  debt-closure accounting, alongside 02-DEBT.md's carried items.
- No blockers for 03-07.

---
*Phase: 03-borrowed-views-and-cfg-lifetimes*
*Plan: 05*
*Completed: 2026-09-04*

## Self-Check: PASSED
