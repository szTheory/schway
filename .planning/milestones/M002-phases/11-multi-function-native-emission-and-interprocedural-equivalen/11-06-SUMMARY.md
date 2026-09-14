---
phase: 11-multi-function-native-emission-and-interprocedural-equivalen
plan: 06
subsystem: session
tags: [qlt03, reachability-register, mechanical-enumeration, corevalidate, testsupport]

# Dependency graph
requires:
  - phase: 11-multi-function-native-emission-and-interprocedural-equivalen
    provides: "11-04's mid-phase gate ratification admitting waves 4-6; 11-05's guard ledger and four-tier differential; testsupport.GenerateCallGraphCorpus/CallGraphCorpusShapes (Phase 08/09) as the declared production space this register audits"
provides:
  - "qlt03_shape_register.go/.json: the five-axis call-graph-shape reachability register -- 35 mechanically enumerated cells (20 full argument-mode x return-origin, 15 pairwise topology x shape x depth), 7 reached, 28 unreachable, each unreachable row naming exactly one of five closed proof mechanisms and a real, existing falsifier test"
  - "AuditQLT03Register: control-ID-per-failure audit over absent cells, unclassified rows, gap/reached_thin rows, out-of-set mechanisms, and duplicate cell keys, proven able to fail via an export_test.go seam without editing the committed register"
  - "The headline borrow_of_callee_result finding (D-11-49): 4 rows (one per argument mode) classified `refused`, citing corevalidate.peerDeriveOriginFacts' missing core.OpCall case (D-10-C01) and testdata/phase10/relay_depth3_refuse.lang as the falsifying fixture"
affects: [11-07, 11-08, 11-09]

# Actuals (#2632)
actuals:
  tokens: 58000
  tasks: 3
  commits: 1

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Two-slice cell identity instead of one forced 5-tuple: rows are keyed by only the axes their own designed slice varies over (argument_mode x return_origin, or edge_topology x callee_body_shape x composition_depth_bucket), joined with a declared separator and an axis-set discriminator -- avoids fabricating meaningless placeholder values for axes a given slice does not probe, and makes duplicate-key collision structurally impossible across slices."
    - "Mechanical structural classifiers over core.LinearOperation (qlt03ClassifyArgumentMode/qlt03ClassifyReturnOrigin/qlt03ClassifyBodyShape/qlt03EdgeCompositionDepth) derive every axis value from operation graph structure alone, never from which testsupport builder function constructed a given function -- so reclassification survives a future rewrite of the corpus builders' own implementation."
    - "generator-mechanism unreachability is grounded in a COMMITTED literal (QLT03CommittedGeneratorOpKinds) or a NAMED structural-invariant test (TestQLT03StructuralGeneratorLimits) citing the corpus builders' own construction, never a bare 'we ran the corpus and it wasn't there' scan -- the D-11-47 distinction between an admissible mechanism and a vacuous drift detector."

key-files:
  created:
    - internal/compiler/session/qlt03_shape_register.go
    - internal/compiler/session/qlt03_shape_register.json
    - internal/compiler/session/qlt03_shape_register_test.go
    - internal/compiler/session/export_test.go
  modified: []

key-decisions:
  - "Row identity is a two-slice design (argument_mode x return_origin OR edge_topology x callee_body_shape x composition_depth_bucket), not a single forced 5-tuple with placeholder defaults for the axes a given slice doesn't vary -- matches 11-CONTEXT.md's own wording ('FULL enumeration of argument-mode x return-origin (20 cells) plus pairwise coverage over the rest') literally rather than reading it as one 5-axis grid, and eliminates an entire class of accidental cross-slice key collisions a fixed-default design would have required careful arithmetic to avoid."
  - "Argument mode across every edge in this generator is uniformly 'move' (never 'copy' or a borrow mode) -- an additional finding beyond D-11-45's own prediction (which only named the borrow column empty), grounded in the corpus builders' own construction (every OpCall argument is sourced from either the function's own parameter directly or a prior OpCall's own target, never an OpCopy target) and pinned by TestQLT03StructuralGeneratorLimits rather than silently asserted."
  - "The headline borrow_of_callee_result gap is proven via testdata/phase10/relay_depth3_refuse.lang -- an already-committed Phase 10 fixture whose own file header independently documents the identical corevalidate.peerDeriveOriginFacts OpCall gap -- rather than constructing a new raw core.Program by hand, since a real, already-checked .lang fixture is stronger evidence than a synthetic one built solely for this register."
  - "All four rows of the borrow_of_callee_result column (one per argument mode: move, copy, borrow_shared, borrow_exclusive) are classified `refused`, not just the `move` cell 11-CONTEXT.md's own prose singled out -- the corevalidate refusal is a property of the RETURN's own origin, independent of how an unrelated, later argument might be passed, so narrowing the refused classification to one argument mode would have been an unjustified asymmetry in the register's own mechanical enumeration."

requirements-completed: [QLT-03]

coverage:
  - id: D1
    description: "The five-axis taxonomy is declared as closed Go value sets (4 argument modes, 5 return origins, 3 callee body shapes, 3 composition-depth buckets, 5 edge topologies matching testsupport.CallGraphCorpusShapes()), and the generator's emittable op-kind set is pinned against a committed literal ({call, copy, return}) rather than a second recomputation"
    requirement: "QLT-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/qlt03_shape_register_test.go#TestQLT03GeneratorOpKindClosure"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/qlt03_shape_register_test.go#TestQLT03AxisCompletenessAgainstCorpusShapes"
        status: pass
    human_judgment: false
  - id: D2
    description: "Cells are enumerated mechanically (35 rows: 20 full argument-mode x return-origin, 15 pairwise topology x shape x depth) with exact case-sensitive cell identity; a case- or whitespace-differing key is reported as a duplicate-row error, not accepted as a distinct cell"
    requirement: "QLT-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/qlt03_shape_register_test.go#TestQLT03CellIdentityIsExactStringEquality"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/qlt03_shape_register_test.go#TestQLT03ReachedRecomputationIsIdempotent"
        status: pass
    human_judgment: false
  - id: D3
    description: "The committed register (35 rows, 7 reached / 28 unreachable) loads with a discriminated-union invariant enforced (exactly one of reached/unreachable per row), a closed five-value proof-mechanism set enforced with free text and nonexistent falsifier tests rejected, and the borrow_of_callee_result family named as the headline refused negative with its diagnostic and falsifying fixture"
    requirement: "QLT-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/qlt03_shape_register_test.go#TestQLT03RegisterRowsAreDiscriminated"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/qlt03_shape_register_test.go#TestQLT03RegisterRejectsOutOfSetMechanism"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/qlt03_shape_register_test.go#TestQLT03RegisterRejectsNonexistentFalsifierTest"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/qlt03_shape_register_test.go#TestQLT03RegisterRejectsBothOrNeitherPopulated"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/qlt03_shape_register_test.go#TestQLT03BorrowOfCalleeResultIsRefused"
        status: pass
    human_judgment: false
  - id: D4
    description: "AuditQLT03Register returns zero failures over the committed register, is proven able to FAIL via an export_test.go fault-injection seam without editing the committed JSON, recomputes reach at three corpus sizes without drift, runs swarm (shape-omitting and op-kind-omitting) configurations, and is deterministic under repetition and t.Parallel() -- never mutating qlt03_shape_register.json"
    requirement: "QLT-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/qlt03_shape_register_test.go#TestQLT03Register"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/qlt03_shape_register_test.go#TestQLT03StatusRecomputation"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/qlt03_shape_register_test.go#TestQLT03AuditCanFail"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/qlt03_shape_register_test.go#TestQLT03SwarmConfigurations"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/qlt03_shape_register_test.go#TestQLT03RegisterIsDeterministicUnderRepeatAndParallel"
        status: pass
    human_judgment: false

duration: 105min
completed: 2026-09-12
status: complete
---

# Phase 11 Plan 6: QLT-03 Call-Graph-Shape Reachability Register Summary

**A 35-row, mechanically enumerated axis-product register records exactly which cross-function call-graph shapes `testsupport.GenerateCallGraphCorpus` reaches (7 cells) and names, for every one of the remaining 28 cells, one admissible closed-set proof mechanism plus a real falsifier test -- with the headline `borrow_of_callee_result` gap (`corevalidate.peerDeriveOriginFacts` has no `core.OpCall` case) classified `refused` and proven against an already-committed Phase 10 fixture.**

## Performance

- **Duration:** ~105 min
- **Started:** 2026-09-12
- **Completed:** 2026-09-12
- **Tasks:** 3 completed (implemented and committed as one coherent commit; see Deviations)
- **Files modified:** 4 created, 0 modified

## Accomplishments

- **The five-axis taxonomy** (`internal/compiler/session/qlt03_shape_register.go`) is declared as closed Go value sets: edge topology (5 values, identical to `testsupport.CallGraphCorpusShapes()`), argument mode across the edge (`move`, `copy`, `borrow_shared`, `borrow_exclusive`), callee body shape (`relay`, `leaf_use`, `leaf_pass`), returned-value origin class (`fresh_owned`, `param_forwarded`, `borrow_of_own_param`, `borrow_of_callee_result`, `foreign_derived`), and a composition-depth bucket (`depth_1`, `depth_2_3`, `depth_4_plus`).
- **`QLT03CommittedGeneratorOpKinds()` pins the generator's own emittable op-kind set** at `{call, copy, return}`; `TestQLT03GeneratorOpKindClosure` sweeps every shape at n in {8, 24, 48} and compares the OBSERVED set against this committed literal, firing the instant a borrow or foreign-call op-kind is added -- before any row is even consulted.
- **Mechanical structural classifiers** (`qlt03ClassifyArgumentMode`, `qlt03ClassifyReturnOrigin`, `qlt03ClassifyBodyShape`, `qlt03EdgeCompositionDepth`) derive every axis value from `core.LinearOperation` structure alone (backward traces through Move/Copy/Borrow/Call chains), never from which `testsupport` builder function constructed a given function.
- **35 mechanically enumerated cells**: a FULL 20-cell argument-mode x return-origin product plus a 15-cell pairwise-coverage slice over (edge topology x callee body shape x composition-depth bucket), built via a Latin-square-style formula (`depth = (topology_index + shape_index) mod 3`) -- the full 2,560-cell product is not taken (D-11-45).
- **7 cells reached, 28 unreachable.** The generator reaches exactly `(move, fresh_owned)` and `(move, param_forwarded)` in the argument-mode slice, and exactly one `depth_1` cell per topology in the pairwise slice (`chain/relay`, `diamond/leaf_pass`, `dense/leaf_use`, `parser-shaped/relay`, `forward/leaf_pass`). Argument mode is uniformly `move` across the ENTIRE sweep -- an additional finding beyond D-11-45's own prediction that only the borrow column would be empty.
- **The headline negative (D-11-49, D-10-C01):** all four `borrow_of_callee_result` cells (one per argument mode) are classified `refused`, citing `core.callee_not_callable` and `corevalidate.peerDeriveOriginFacts`' missing `core.OpCall` case. `TestQLT03BorrowOfCalleeResultIsRefused` proves this directly against `testdata/phase10/relay_depth3_refuse.lang` -- an already-committed Phase 10 fixture whose own file header independently documents this exact gap.
- **The closed five-value proof-mechanism table** (`grammar`, `refused`, `generator`, `bounded_exhaustive`, `gap`) is enforced by `LoadQLT03Register`: an out-of-set mechanism (including free text) is rejected, and a non-`gap` row naming a falsifier test that does not exist as a real `func Test...` symbol in the package's own `_test.go` files (verified via `go/parser`) is rejected.
- **`witness` is not boolean** (D-11-48): `ReachedEvidence.Thin()` implements the `< 3 occurrences AND 1 shape AND max_depth == 1` gate; no committed row currently trips it (all reached cells clear the threshold, several with a genuinely close margin, e.g. `forward/leaf_pass/depth_1` at exactly 3 occurrences).
- **`AuditQLT03Register`** fails on any absent enumerated cell, unclassified row, both/neither-populated row, `gap` row, `reached_thin` row, out-of-set mechanism, or duplicate cell key -- each with its own `control:qlt03.*` identifier -- and is proven able to FAIL (`TestQLT03AuditCanFail`) via an `export_test.go` fault-injection seam (`SetQLT03InjectExtraRequiredCellForTest`) that perturbs the audit's own cell enumeration WITHOUT editing the committed `qlt03_shape_register.json`.
- **Drift protections**: `TestQLT03StatusRecomputation` (statuses hold at n=8/24/48 individually), `TestQLT03AxisCompletenessAgainstCorpusShapes` (cross-checks the edge-topology axis against the live corpus shape list), `TestQLT03SwarmConfigurations` (shape-omitting and op-kind-omitting sweeps; this generator's own shape independence means no omission-only cell was found, an honest negative result rather than a fabricated positive), and `TestQLT03RegisterIsDeterministicUnderRepeatAndParallel` (identical verdicts under `t.Parallel()`, register file mtime unchanged).

## Task Commits

Implemented and committed as one coherent commit rather than three (see Deviations for the reasoning):

1. **Tasks 1+2+3: QLT-03 register, loader, audit, mutation kill, and drift protection** -- `679bf47` (feat)

## Files Created/Modified

- `internal/compiler/session/qlt03_shape_register.go` -- axes, committed op-kind literal, structural classifiers, sweep/aggregation, cell enumeration, `QLT03Row`/`ReachedEvidence`/`UnreachableProof`, `LoadQLT03Register`, `AuditQLT03Register`
- `internal/compiler/session/qlt03_shape_register.json` -- the committed 35-row register data artifact
- `internal/compiler/session/qlt03_shape_register_test.go` -- all 15 named tests (Q-03 closure, structural limits, headline falsifier, cell identity, idempotency, discriminated rows, loader rejection tests, top-level audit, status recomputation, axis completeness, mutation kill, swarm configurations, determinism under parallel)
- `internal/compiler/session/export_test.go` -- new file: `SetQLT03InjectExtraRequiredCellForTest` (the D-11-50 mutation-kill seam) and `setQLT03RegisterBytesForTest` (a loader-rejection test seam)

## Decisions Made

See `key-decisions` above. The two load-bearing ones: (1) row identity uses two DISJOINT axis-set slices (argument-mode x return-origin, and topology x shape x depth) rather than forcing every row into one 5-tuple with placeholder defaults, matching 11-CONTEXT.md's own literal wording and eliminating an entire class of cross-slice key-collision arithmetic; and (2) the headline `borrow_of_callee_result` refusal is proven against an already-committed, independently-documented Phase 10 fixture (`relay_depth3_refuse.lang`) rather than a hand-built synthetic program, since real committed evidence is stronger than a purpose-built one.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking issue] `qlt03_shape_register.json` created ahead of its plan-assigned task**
- **Found during:** Task 1, first `go build` after adding `//go:embed qlt03_shape_register.json` to `qlt03_shape_register.go`
- **Issue:** The plan's `files_modified` lists `qlt03_shape_register.json` under Task 2 only, but Go's `//go:embed` directive requires the embedded file to exist at COMPILE time, so the package cannot build at all once Task 1's `qlt03_shape_register.go` references it.
- **Fix:** Created a minimal placeholder (`[]`) alongside Task 1's own file, then replaced it with the full 35-row register once Task 2's classification logic existed.
- **Files modified:** `internal/compiler/session/qlt03_shape_register.json`
- **Verification:** `go build ./internal/compiler/session/...` succeeds at every intermediate step.
- **Committed in:** `679bf47` (final committed state only; the placeholder never reached its own commit)

### Architectural Note (process deviation, not a code defect)

**2. Three tasks implemented and committed as one commit, not three**

The plan's own per-task commit cadence assumes each task's file-level changes can stand alone as a separately-buildable, separately-green commit. In practice, this register's three tasks are tightly interdependent in a way that makes that split counterproductive rather than merely inconvenient:

- Task 3's own drift protections (`TestQLT03StatusRecomputation`, `TestQLT03AxisCompletenessAgainstCorpusShapes`) and its mutation-kill (`TestQLT03AuditCanFail`) require `AuditQLT03Register` and the loaded register to already exist and already be green -- which is Task 2's own deliverable.
- The loader's own falsifier-test-existence check (`LoadQLT03Register` rejecting a nonexistent `falsifier_test` symbol, Task 2's own acceptance criterion) requires the test file to already declare every falsifier test the JSON cites -- which are Task 3's own drift-protection test names for the "generator"-mechanism rows.
- Building the actual 35-row JSON content (Task 2) required first computing the mechanical sweep's real reach data using Task 1's classifiers, which in turn could only be validated as correct once Task 3's audit existed to check it end-to-end.

Splitting into three separately-committed states would have required either (a) forward-declaring Task 3's falsifier-test symbols with no real assertions yet, purely so Task 2's loader validation would pass -- a form of test-theater the plan's own must_haves explicitly reject in spirit -- or (b) temporarily weakening the loader's falsifier-existence check, then re-strengthening it in a later commit. Both are worse than one commit that is green, complete, and internally consistent throughout. This is a process deviation from the plan's stated per-task commit cadence, not a scope or correctness deviation: every task's own acceptance criteria are met, and all three tasks' verify commands pass against the final tree.
- **Impact:** None on correctness or the register's own content. Documented here per the executor's Rule-4-adjacent obligation to flag, not silently absorb, a deviation from the plan's own process instructions.

---

**Total deviations:** 1 auto-fixed (Rule 3, blocking issue), 1 architectural/process note (commit cadence).
**Impact on plan:** None on scope or correctness. The blocking-issue fix was required for the package to build at all; the commit-cadence note reflects a genuine interdependency discovered during implementation, not a shortcut taken to save time.

## Issues Encountered

None beyond the deviations above -- both discovered and resolved during normal implementation, not left as open questions.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `qlt03_shape_register.go`/`.json` is now the authoritative QLT-03 register for the rest of Phase 11 -- a future plan touching call-graph-shape reachability should extend this register, not create a second one, and per D-11-43 it stays SEPARATE from `qlt01_registry.json` (verified: `git diff --name-only internal/compiler/session/qlt01_registry.json` from this plan's base is empty).
- The `borrow_of_callee_result` `refused` finding is consistent with, and cites, the same `D-10-C01`/`corevalidate.peerDeriveOriginFacts` gap already carried as open, unowned debt in `PHASE-11-DEBT.md` -- this plan does not close that gap, only documents its reachability-register implications honestly.
- `TestQLT03StructuralGeneratorLimits`'s own pinned invariants (argument mode always `move`; only `forward` topology ever produces composition depth > 1; `forward`'s own callees are always `leaf_pass`) are now load-bearing facts about `testsupport.GenerateCallGraphCorpus` -- any future widening of that generator (e.g. adding a borrow-carrying shape) will need this register's rows re-audited, and `TestQLT03GeneratorOpKindClosure`/`TestQLT03StructuralGeneratorLimits` will surface the surprise immediately.
- No blockers to proceeding with plan 11-07.

---
*Phase: 11-multi-function-native-emission-and-interprocedural-equivalen*
*Completed: 2026-09-12*

## Self-Check: PASSED

All created files verified present on disk (`internal/compiler/session/qlt03_shape_register.go`,
`qlt03_shape_register.json`, `qlt03_shape_register_test.go`, `export_test.go`). Commit hash `679bf47`
verified present in `git log`. `go build ./...` clean. `go vet ./internal/compiler/session/...` clean.
`go test ./...` green across all 23 packages (independently re-verified by the orchestrator: exit 0,
0 FAIL). `go test ./internal/compiler/session/... -run TestQLT03 -v -count=2` prints 15 `--- PASS:`
lines per run (30 total), identical verdicts across both runs. `git status --porcelain
internal/compiler/session/qlt03_shape_register.json` is empty after every test run in this session.
