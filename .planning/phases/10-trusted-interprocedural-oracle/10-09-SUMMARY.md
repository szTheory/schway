---
phase: 10-trusted-interprocedural-oracle
plan: 09
subsystem: interp
tags: [golden-corpus, stability-freeze, mutation-testing, coverage-floor, determinism]

# Dependency graph
requires:
  - phase: 10-trusted-interprocedural-oracle
    provides: "plan 10-01's frame stack (D-10-21), plan 10-04's MaxCallDepth=128 depth-exceeded Outcome, plan 10-05's cross-frame drain order (D-10-33/D-10-35) -- SEM-09's event emission final BEFORE this freeze, per D-10-56's hard ordering requirement"
provides:
  - "TestInterpOracleGoldenCorpus: a byte-for-byte golden corpus (testdata/phase10/interp_oracle/, 6 programs) freezing interp's emitted Schema and Execution shape across every reachable terminal path, with regeneration gated on a reviewed commit rather than a flag (D-10-57 clause 1/2)"
  - "TestInterpOracleCorpusDeterministic + TestInterpSchemaMatchesGoldenCorpus: determinism proven over the FULL corpus at -count=10, and a schema-bump-requires-golden-regeneration guard (D-10-57 clause 3/4)"
  - "The structural coverage floor (D-10-58) as three enumerated tables in 10-VALIDATION.md: an operation-kind x execution-path matrix (no blank cells), a refusal-path table, and a six-row Mutation-Kill Register"
affects: [11]

# Actuals (#2632)
actuals:
  tokens: 13967
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "A golden corpus mixing real-pipeline-driven programs (checkedProgramFromFixture, generateAndCheckCallDepthChain) with deliberately synthetic core.Program probes (oracleTypedFailureProgram) for the one terminal path -- core.OpFail -- no real check.go-emitted body can ever place in a shape interp actually visits, since interp unconditionally follows the ok edge of every OpForeignCall this phase"
    - "Regeneration-by-reviewed-commit, never by flag: the golden test's own doc comment documents the exact scratch-test procedure a reviewer runs once, inspects by hand, and deletes -- no flag.Bool-style mechanism exists anywhere in the package"
    - "A structural coverage floor expressed as PROVABLY UNREACHABLE cells grounded in existing codebase doc comments (D-04-15's match-arm-only OpDefect rule, checkFallibleLinear's Match-less-only Blocks rule) rather than bare percentage claims"
  patterns_established: []

key-files:
  created:
    - internal/compiler/interp/interp_oracle_golden_test.go
    - testdata/phase10/interp_oracle/single_frame_return.golden.json
    - testdata/phase10/interp_oracle/cross_frame_normal_return.golden.json
    - testdata/phase10/interp_oracle/typed_failure.golden.json
    - testdata/phase10/interp_oracle/defect_terminal.golden.json
    - testdata/phase10/interp_oracle/cross_frame_nonlocal_landing_pad.golden.json
    - testdata/phase10/interp_oracle/call_depth_exceeded_refusal.golden.json
  modified:
    - internal/compiler/interp/interp_test.go
    - .planning/phases/10-trusted-interprocedural-oracle/10-VALIDATION.md

key-decisions:
  - "core.OpFail is frozen via a deliberately synthetic core.Program (oracleTypedFailureProgram), never a real .lang fixture: interp's own runLinearBlocks doc comment already states every OpForeignCall unconditionally follows the ok edge this phase, so the err block a fallible call's OpFail lives in is structurally unreachable through any real, checked Lang program -- freezing it required isolating interp's own mechanism directly, mirroring moveAsCopyProbeProgram's established precedent"
  - "The operation-kind coverage matrix's real gaps (core.OpBorrowShared/core.OpBorrowExclusive untested at every path; core.OpCopy/core.OpMove each tested at only one of three; core.OpCall untested at runBranchArm) were closed with new tests (TestOperationKindCoverageAcrossAllThreePaths, TestCallFromBothMatchArmsAcrossFrames) rather than annotated as gaps, per Task 3's own instruction"
  - "Four refusal-path tests were added to close previously-untested guard clauses; TestRunRefusesInvalidBodyUnion documents a genuine finding rather than a clean pass: Run's own HasClosedBody guard is provably unreachable in practice, since corevalidate.Validate (Run's unconditional first act) always refuses that exact shape first as core.invalid_body"
  - "10-VALIDATION.md's Per-Task Verification Map TBDs are resolved by binding each row to its actual originating plan/task/wave (e.g. the OWN-05b guard traces to plan 10-01 Task 3, wave 1); nyquist_compliant set to true since all ten rows now bind to a real automated command, verified green"

patterns-established:
  - "Golden-corpus regeneration procedure documented entirely in the test's own doc comment (the exact scratch-test snippet, run-once command, and manual git-diff-review step), never automated -- a CI failure can never silently absorb a drift"

requirements-completed: [SEM-08, SEM-09, OWN-05b]

coverage:
  - id: D1
    description: "interp's emitted Schema value and Execution document shape are frozen byte-for-byte against a 6-program golden corpus covering every reachable terminal path (single/cross-frame return, typed failure, defect, cross-frame nonlocal landing pad, depth-exceeded refusal)"
    requirement: SEM-08
    verification:
      - kind: unit
        ref: "internal/compiler/interp/interp_oracle_golden_test.go#TestInterpOracleGoldenCorpus"
        status: pass
      - kind: other
        ref: "ls testdata/phase10/interp_oracle/ | wc -l (returns 6)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Regeneration is gated on a reviewed commit with a stated rationale, never a flag; the Phase 11 escalation path (versioned schema bump + regenerated goldens + decision entry) is named and an in-place edit is explicitly forbidden"
    requirement: SEM-08
    verification:
      - kind: other
        ref: "grep -rn 'flag.Bool' internal/compiler/interp/ (prints nothing)"
        status: pass
      - kind: unit
        ref: "internal/compiler/interp/interp_oracle_golden_test.go#TestInterpSchemaMatchesGoldenCorpus"
        status: pass
    human_judgment: false
  - id: D3
    description: "interp is deterministic across the full golden corpus at -count=10 (Go's map-iteration randomization gets a real chance to surface an accidental map range)"
    requirement: SEM-09
    verification:
      - kind: unit
        ref: "internal/compiler/interp/interp_oracle_golden_test.go#TestInterpOracleCorpusDeterministic"
        status: pass
      - kind: other
        ref: "go test ./internal/compiler/interp/... -run 'TestInterpDeterministicAcrossRuns|TestInterpOracleGoldenCorpus|TestInterpOracleCorpusDeterministic' -count=10"
        status: pass
    human_judgment: false
  - id: D4
    description: "The structural coverage floor is an enumerated table (operation-kind x execution-path matrix, refusal-path table, six-row Mutation-Kill Register) with no blank cells and no percentage, satisfying Phase 11's hard precondition that interp be STABLE"
    requirement: OWN-05b
    verification:
      - kind: unit
        ref: "internal/compiler/interp/interp_test.go#TestOperationKindCoverageAcrossAllThreePaths"
        status: pass
      - kind: unit
        ref: "internal/compiler/interp/interp_test.go#TestCallFromBothMatchArmsAcrossFrames"
        status: pass
      - kind: other
        ref: "grep -c 'TBD' .planning/phases/10-trusted-interprocedural-oracle/10-VALIDATION.md (returns 0)"
        status: pass
    human_judgment: true
    rationale: "Whether the PROVABLY UNREACHABLE cells' structural reasoning is sound, and whether the coverage floor genuinely satisfies Phase 11's hard precondition, is a reviewer judgment the plan's own must_haves flagged as needing confirmation (the EDGE flagged_assumption carried from plans 10-01/10-04)."

duration: 70min
completed: 2026-09-11
status: complete
---

# Phase 10 Plan 9: Trusted Interprocedural Oracle — Stability Freeze Summary

`interp` is now STABLE, not merely working: its emitted schema and shape are frozen byte-for-byte against a six-program golden corpus covering every reachable terminal path, determinism is proven over that full corpus at `-count=10`, and the previously-absent coverage story is replaced with an enumerated structural floor (operation-kind matrix, refusal-path table, six-row Mutation-Kill Register) with zero blank cells.

## Performance

- **Duration:** ~70 min
- **Started:** 2026-09-11
- **Completed:** 2026-09-11
- **Tasks:** 3 completed
- **Files modified:** 9 (2 new test files counted as 1 golden test file + 6 golden fixtures, 1 modified test file, 1 modified validation doc)

## Accomplishments

- `TestInterpOracleGoldenCorpus` freezes 6 programs' `interp.CanonicalBytes` output byte-for-byte under `testdata/phase10/interp_oracle/`: single-frame return (`defect_terminal.lang`'s "Go" arm), cross-frame normal return (`call_basic.lang`), a typed failure (a deliberately synthetic `core.Program`, since interp's own `runLinearBlocks` always follows the ok edge and can never reach a real fallible call's err-block `OpFail`), a real defect (`defect_terminal.lang`'s "Halt" arm), a cross-frame abrupt exit through the foreign nonlocal landing pad, and the SEM-08 depth-exceeded refusal over a genuine 129-function chain. No `-update` flag or any rewrite mechanism exists anywhere in the package; the test's own doc comment documents regeneration as a reviewed commit with a stated rationale and the exact scratch-test procedure a reviewer runs.
- `TestInterpOracleCorpusDeterministic` extends the determinism guarantee from a single fixture to the full corpus, and `TestInterpSchemaMatchesGoldenCorpus` scans every committed golden's own schema strings against the package's declared closed set (`Schema`, `execution.Schema1`) so a schema bump can never land without regenerating the goldens in the same commit — the mechanical half of the named three-part Phase 11 escalation path (versioned bump + regenerated goldens + decision entry) documented in the golden test's own doc comment.
- The structural coverage floor (D-10-58) replaced the absent coverage story with three enumerated tables in `10-VALIDATION.md`: an operation-kind x execution-path matrix with every cell naming a test or stating `PROVABLY UNREACHABLE` with a structural reason (grounded in D-04-15's match-arm-only `OpDefect` rule and `checkFallibleLinear`'s Match-less-only `Blocks` rule, never a bare assertion); a refusal-path table for the six named refusals; and a six-row Mutation-Kill Register for the call-stack, cross-frame, and ownership machinery this phase added.
- Building the matrix surfaced real (not merely structural) gaps: `core.OpBorrowShared`/`core.OpBorrowExclusive` had never been exercised by any interp test at any path; `core.OpCopy`/`core.OpMove` were each exercised at only one of the three reachable paths; `core.OpCall` had never been driven through `runBranchArm`. Per Task 3's own instruction, these were closed with new tests (`TestOperationKindCoverageAcrossAllThreePaths`, `TestCallFromBothMatchArmsAcrossFrames`) rather than annotated as gaps.
- Four refusal-path tests were added to close previously-untested guard clauses. `TestRunRefusesInvalidBodyUnion` documents a genuine finding, not a clean pass: `Run`'s own `!function.HasClosedBody()` guard is provably unreachable via `Run()` in practice, because `corevalidate.Validate` (Run's unconditional first act) always refuses that exact shape first as `core.invalid_body` — confirmed directly by driving the mutated program through `Run` and observing the actual error, not assumed.
- `10-VALIDATION.md`'s Per-Task Verification Map TBDs are resolved: every row now binds to its actual originating plan/task/wave, and `nyquist_compliant` is set to `true` since all ten rows bind to a real automated command, verified green.

## Task Commits

Each task was committed atomically:

1. **Task 1: Freeze the oracle — a byte-for-byte golden corpus** - `5811cc4` (test)
2. **Task 2: Determinism at -count=10 and the Phase 11 escalation path** - `141559b` (test)
3. **Task 3: The structural coverage floor, as an enumerated table** - `2f236ef` (test)

## Files Created/Modified

- `internal/compiler/interp/interp_oracle_golden_test.go` — new file: `TestInterpOracleGoldenCorpus`, `TestInterpSchemaMatchesGoldenCorpus`, `TestInterpOracleCorpusDeterministic`, `interpOracleCorpus`, `checkedProgramFromFixture`, `oracleTypedFailureProgram`.
- `testdata/phase10/interp_oracle/*.golden.json` — 6 committed golden files, one per corpus program.
- `internal/compiler/interp/interp_test.go` — `TestOperationKindCoverageAcrossAllThreePaths`, `oracleBorrowSequenceProgram`, `TestCallFromBothMatchArmsAcrossFrames`, and four refusal-path tests (`TestRunRefusesCorevalidateInvalidProgram`, `TestRunRefusesAbsentFunction`, `TestRunRefusesInvalidBodyUnion`, `TestRunFrameStackRefusesUnknownOperationKind`).
- `.planning/phases/10-trusted-interprocedural-oracle/10-VALIDATION.md` — Per-Task Verification Map fully bound, `nyquist_compliant: true`, three new structural coverage-floor tables appended.

## Decisions Made

See `key-decisions` in frontmatter.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Initial coverage-matrix assertions had off-by-one event-count and wrong Input-field expectations**
- **Found during:** Task 3, first run of `TestOperationKindCoverageAcrossAllThreePaths` and `TestCallFromBothMatchArmsAcrossFrames`
- **Issue:** `TestOperationKindCoverageAcrossAllThreePaths` expected 4 events but the terminal `OpReturn` also emits a `function.returned` event (5 total); `TestCallFromBothMatchArmsAcrossFrames` checked `event.Input == input` on a `function.returned` event, but `terminalOutcome`'s `OpReturn` case never populates `Input`.
- **Fix:** Updated the expected event-kind slice to include `"function.returned"`; changed the helper-call assertion to check `event.FunctionID` has the `:fn:helper` suffix instead of `event.Input`.
- **Files modified:** internal/compiler/interp/interp_test.go
- **Verification:** Both tests pass; re-ran the full `internal/compiler/interp` suite and `go test ./...` to confirm no other test's expectations were affected.
- **Committed in:** `2f236ef` (fixed before commit)

---

**Total deviations:** 1 auto-fixed (Rule 1 — a test-authoring bug caught and fixed before committing, not a production-code defect).
**Impact on plan:** No scope creep; the fix corrected the plan author's own test expectations against interp's actual (correct, unchanged) behavior.

## Issues Encountered

None beyond the deviation above.

## User Setup Required

None - no external service configuration required.

## Known Stubs

None — this plan adds test infrastructure and documentation only; no production code paths were stubbed.

## Next Phase Readiness

- `interp` is STABLE per the roadmap's own hard precondition for Phase 11: its emitted schema and shape are frozen byte-for-byte across every reachable terminal path, determinism holds over the full corpus at `-count=10`, and the coverage floor is an enumerated table with no blank cells.
- The three-part Phase 11 escalation path (versioned schema bump + regenerated goldens + a decision entry) is named and mechanically enforced by `TestInterpSchemaMatchesGoldenCorpus`; an in-place golden edit without a version bump is explicitly forbidden.
- `PHASE-10-DEBT.md`'s existing open items remain open and unchanged by this plan (this plan touches only `interp`/`interp_test`/`10-VALIDATION.md`): D-10-28 (corevalidate has no interprocedural `usesParam` peer, landing phase deliberately open), D-10-34 (cgen's multi-frame nonlocal pad, named Phase 11 debt), and `deferred-items.md`'s two open findings (`corevalidate.peerDeriveOriginFacts` has no `core.OpCall` case, from plan 10-07; `corevalidate.Result.LoanEndpoints()` naturally incomplete for a fail-fast-refused branched function, from plan 10-08) — both reflected here as open, not re-logged as duplicates.
- This is Phase 10's last plan. Ready for `/gsd-verify-work 10` and Phase 11 planning.

## Self-Check: PASSED

- `internal/compiler/interp/interp_oracle_golden_test.go` — FOUND
- `internal/compiler/interp/interp_test.go` — FOUND
- `.planning/phases/10-trusted-interprocedural-oracle/10-VALIDATION.md` — FOUND
- `testdata/phase10/interp_oracle/` (6 golden files) — FOUND
- Commit `5811cc4` — FOUND (`git log --oneline --all`)
- Commit `141559b` — FOUND
- Commit `2f236ef` — FOUND
- `go test ./internal/compiler/interp/... -run 'TestInterpDeterministicAcrossRuns|TestInterpOracleGoldenCorpus|TestInterpOracleCorpusDeterministic' -count=10` — PASS
- `go test -race ./internal/compiler/interp/...` — PASS, no DATA RACE
- `go test ./internal/compiler/interp/... -shuffle=on -count=2` — PASS
- `grep -rn 'flag.Bool' internal/compiler/interp/` — prints nothing
- `grep -c 'TBD' .planning/phases/10-trusted-interprocedural-oracle/10-VALIDATION.md` — prints 0
- `go test ./...` — PASS (no non-`ok` lines except pre-existing no-test-file packages)
- `go test -race ./...` — PASS
- `go build ./... && go vet ./...` — PASS

---
*Phase: 10-trusted-interprocedural-oracle*
*Completed: 2026-09-11*
