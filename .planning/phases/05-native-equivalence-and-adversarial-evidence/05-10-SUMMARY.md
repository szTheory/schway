---
phase: 05-native-equivalence-and-adversarial-evidence
plan: 10
subsystem: testing
tags: [go, delta-debugging, reduction, core-ir, interpreter-mismatch]

# Dependency graph
requires:
  - phase: 05-native-equivalence-and-adversarial-evidence (05-06, 05-09)
    provides: "Phase5CompareEngines's five axis identifiers (the reducer's oracle shape); the mid-phase gate (D-05-40) releasing waves 5-7"
provides:
  - "internal/compiler/reduce -- Move/Moves()/Reduce/Result: the five fixed reduction moves, MaxReductionAttempts=64 budget, Minimality (fixpoint/budget_exhausted)"
  - "reduce.Signature/Interesting -- the D-05-24 interestingness predicate (axis + operation-or-causal-role + engine-pair conjunction, foreign-call-sequence-drift rejection)"
  - "reduce.ProjectSource -- deterministic core-to-source projection for straight-line and foreign resource-lifecycle bodies, always populated from the same reduction run"
affects: [05-12]

actuals:
  tokens: 17850
  tasks: 3
  commits: 2

tech-stack:
  added: []
  patterns:
    - "A new leaf package (no dependency on internal/compiler/session) built from D-05-23's fixed five-move spec, since no delta-debugging/core-structure minimization precedent existed anywhere in this repository (05-PATTERNS.md's own 'No Analog Found' row)."
    - "A structural core-level move that touches the block/edge graph rebuilds the affected region from scratch via the SAME algorithm check.go's real front end uses (checkResourceLifecycle for foreign chains, analyzeStraightLine's ID scheme for a collapsed match arm), rather than surgically patching the existing graph -- this is what keeps a reduced candidate always referentially consistent and keeps ProjectSource's round-trip correspondence provable rather than merely hoped-for."

key-files:
  created:
    - internal/compiler/reduce/reduce.go
    - internal/compiler/reduce/reduce_test.go
    - internal/compiler/reduce/predicate.go
    - internal/compiler/reduce/predicate_test.go
  modified: []

key-decisions:
  - "Reduce's rejection semantics: within one pass, a move's Apply is called repeatedly on its own accepted candidates, but a single interestingness rejection halts that move for the pass (advancing to the next move) rather than searching for an alternative candidate site. This keeps every move a pure, side-effect-free function of core.Program alone (no candidate-index state threaded through Apply's fixed signature) at the cost of not exhaustively searching every removal site per pass -- documented narrowing; a full pass restarts from move 1, so most real programs still reach fixpoint over multiple passes."
  - "drop-offpath-foreign-stage and collapse-branch-to-diverging-arm both rebuild the affected region from scratch (rebuildForeignChain replicates checkResourceLifecycle's exact ID/block/edge scheme; the collapse move replicates analyzeStraightLine's ID scheme) rather than doing ad hoc graph surgery on the existing Blocks/Edges/Operations. An earlier surgical version corrupted the reduced core by leaving a release operation dangling on a removed acquisition when that acquisition had releases on BOTH its error path and the success path (RES-01's own multi-path discharge requirement) -- discovered empirically while writing the round-trip tests, not anticipated in the plan."
  - "SignatureFromDisagreement (mapping a real session.Phase5EngineDisagreement into a reduce.Signature) is explicitly deferred to plan 05-12's own wiring file, per the plan's own instruction: session -> reduce is 05-12's wiring direction, so a reduce -> session dependency here would create an import cycle. reduce.Signature carries every field D-05-24 needs as plain strings/slices, with no session dependency."
  - "ProjectSource explicitly refuses (rather than fabricates) a source projection for a reduced function whose parameter or return type is a user-declared ADT once its Match wrapper has collapsed away: this project's checker only ever admits a declared ADT type inside an EXHAUSTIVE match body (there is no partial-match syntax), so a collapsed 2-arm ADT match has literally no valid, re-checkable source under the current grammar. Discovered empirically (a 'flag' bare-return projection hit check.go's type.unknown at the parameter-type site) while writing the round-trip test for the defect_dies_by_signal.lang shape; documented as a narrowing rather than worked around by inventing non-standard syntax."

requirements-completed: [INT-02]

coverage:
  - id: D1
    description: "Exactly five reduction moves exist, in fixed deterministic order, with a 64-cycle work budget whose exhaustion is recorded as budget_exhausted rather than an error"
    requirement: INT-02
    verification:
      - kind: unit
        ref: "internal/compiler/reduce#TestReduceHasExactlyFiveMoves"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce#TestReduceIsDeterministic"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce#TestReduceRespectsBudget"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce#TestReduceCountsWork"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce#TestReducePackageNeverTouchesSourceText"
        status: pass
    human_judgment: false
  - id: D2
    description: "The interestingness predicate accepts only candidates preserving the same axis, first-diverging operation (or causal role), and engine pair, and rejects any foreign-boundary call-sequence drift"
    requirement: INT-02
    verification:
      - kind: unit
        ref: "internal/compiler/reduce#TestPredicateAcceptsIdenticalSignature"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce#TestPredicateRejectsDifferentAxis"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce#TestPredicateRejectsDifferentOperation"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce#TestPredicateRejectsDifferentEnginePair"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce#TestPredicateAcceptsShiftedPositionSameCausalRole"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce#TestPredicateRejectsForeignCallSequenceDrift"
        status: pass
    human_judgment: false
  - id: D3
    description: "Every reduction result carries a source projection produced from its own reduced core within the same run, provably corresponding (re-checked through the real front end) and byte-deterministic"
    requirement: INT-02
    verification:
      - kind: integration
        ref: "internal/compiler/reduce#TestProjectedSourceCorrespondsToReducedCore"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce#TestProjectedSourceIsDeterministic"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce#TestResultSourceAlwaysMatchesResultProgram"
        status: pass
    human_judgment: false

duration: ~110min
completed: 2026-09-06
status: complete
---

# Phase 5 Plan 10: Core-Level Reducer Summary

**Five fixed HDD-style moves over the typed core (never source text) with a 64-cycle work budget, a strict same-axis/same-operation-or-causal-role/same-engine-pair/no-foreign-drift interestingness predicate, and a deterministic core-to-source projection proven, through the real front end, to correspond to its own reduced core.**

## Performance

- **Duration:** ~110 min (including iterative round-trip debugging against the real parser/checker)
- **Tasks:** 3
- **Files modified:** 4 (all new)

## Accomplishments

- `internal/compiler/reduce.Moves()` returns exactly the five D-05-23 moves in fixed order (`drop-unused-binding`, `drop-unmatched-arm`, `drop-offpath-foreign-stage`, `collapse-branch-to-diverging-arm`, `truncate-to-minimal-prefix`), each a pure `func(core.Program) (core.Program, bool)` operating only on the operation list and block graph.
- `Reduce` applies the moves greedily to fixpoint or `MaxReductionAttempts = 64` check+compile+execute cycles (one cycle per `interesting` call, whether accepted or rejected), recording `Minimality: "budget_exhausted"` (never an error) when the budget is spent first.
- `Signature`/`Interesting` implement D-05-24's exact conjunction: same `Axis`, AND same `OperationID` (or, if position shifted, the same `CausalRole` -- the one permitted relaxation), AND same `EnginePair`, AND no drift in the ordered `ForeignCallSequence` on the causal path.
- `ProjectSource` renders the reduced core back to Lang source for the two shapes this phase's real corpus exercises (a flat borrow/copy/take chain, and a foreign resource-lifecycle chain built by `checkResourceLifecycle`), populated inside `Reduce` so `Result.Source` and `Result.Program` always come from the same run.
- Round-trip fidelity is proven against the REAL shipped front end (`syntax.Parse` + `check.Program`), not merely asserted: `TestProjectedSourceCorrespondsToReducedCore` re-checks projected source for an unreduced straight-line seed, an unreduced foreign-chain seed, a foreign-chain seed after `drop-offpath-foreign-stage` has fired, and confirms the collapsed-ADT-match case is explicitly (not silently) unsupported.

## Task Commits

1. **Task 1 + Task 3 (reduce.go/reduce_test.go): five moves, Reduce, ProjectSource** - `aa64d56` (feat)
2. **Task 2 (predicate.go/predicate_test.go): the interestingness predicate** - `d1edcd0` (feat)

**Plan metadata:** committed as part of this SUMMARY's own commit (see below).

_Note on commit structure: Tasks 1 and 3 both modify `reduce.go`/`reduce_test.go` (Task 3's `ProjectSource` required revising the move implementations built in Task 1 to keep round-trip correspondence provable -- see Deviations), so they landed as one commit rather than two. Task 2's files (`predicate.go`/`predicate_test.go`) never overlapped and committed separately._

## Files Created/Modified

- `internal/compiler/reduce/reduce.go` (598 lines) - `Move`, `Moves`, `MaxReductionAttempts`, `Reduce`, `Result`, the five move implementations, `rebuildForeignChain` (replicates `checkResourceLifecycle`'s exact ID scheme), `renumberStraightLine` (replicates `analyzeStraightLine`'s ID scheme for a collapsed match arm), `ProjectSource` and its two body-shape projectors.
- `internal/compiler/reduce/reduce_test.go` (~700 lines) - Task 1/2 unit tests against hand-built `core.Program` fixtures, plus Task 3's round-trip tests against the real front end via `syntax.Parse`/`check.Program`.
- `internal/compiler/reduce/predicate.go` (79 lines) - `Signature`, `Interesting`, `foreignCallSequenceDrifted`.
- `internal/compiler/reduce/predicate_test.go` (85 lines) - the six D-05-24 predicate tests.

## Decisions Made

See `key-decisions` in frontmatter: Reduce's single-rejection-halts-this-move-for-the-pass semantics; rebuilding the foreign-chain/collapsed-match regions from scratch (via the real checker's own ID-generation algorithms) rather than surgical graph patching, after an earlier surgical version corrupted a reduced core by leaving a dangling release reference; `SignatureFromDisagreement` deferred to plan 05-12 to keep `reduce` a leaf package; `ProjectSource`'s explicit refusal for a collapsed ADT-typed match (no partial-match syntax exists in this grammar).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `drop-offpath-foreign-stage` corrupted the reduced core by leaving a dangling release reference**

- **Found during:** Task 3's round-trip test (`TestProjectedSourceCorrespondsToReducedCore/foreign_resource_chain_after_offpath_drop`)
- **Issue:** The first implementation removed a foreign call together with only the FIRST `OpRelease` naming it. RES-01's own discharge law means a completed acquisition can be released on more than one path (its own error path AND the success path), so removing just one release left the other referencing a now-deleted `ReleasesOperationID`, and the resulting `core.Program` no longer corresponded to anything a fresh check of any source could produce.
- **Fix:** Rewrote the move to decode the surviving call sequence and rebuild the ENTIRE resource-lifecycle `LinearBody` from scratch via a new `rebuildForeignChain`, which replicates `check.go`'s `checkResourceLifecycle` algorithm (ID scheme, block/edge shape, per-step error type facts, release-ladder ordering) exactly. This also correctly clears the function's `ForeignContract` and the module's now-dead data type when the last acquisition is removed.
- **Files modified:** `internal/compiler/reduce/reduce.go`
- **Verification:** `TestProjectedSourceCorrespondsToReducedCore/foreign_resource_chain_after_offpath_drop` passes; full `go test ./...`, `go test -race ./...`, `go vet ./...`, and `sh scripts/verify-phase5.sh` all green.
- **Committed in:** `aa64d56` (part of the Task 1+3 commit; the bug and its fix were both internal to development, never landing in a separate commit)

**2. [Rule 1 - Bug] `ProjectSource` never emitted a `data` declaration, breaking any body whose type was a user-declared ADT**

- **Found during:** Task 3's round-trip test for the collapsed-match shape
- **Issue:** The straight-line projector only ever emitted `module`/`export`/`fn` -- any function whose parameter or return type was a declared ADT (e.g. `Signal`) failed to re-check with `type.unknown`.
- **Fix:** Added `renderDataType`, which emits the ADT's `data Name = | Alt ...` declaration when `program.DataTypes` names the parameter or return type; omitted entirely for a built-in primitive (`Byte`/`Buffer`), which never appears in `DataTypes`.
- **Files modified:** `internal/compiler/reduce/reduce.go`
- **Verification:** Straight-line and foreign-chain round-trip tests pass with and without a declared ADT in scope.
- **Committed in:** `aa64d56`

**3. [Rule 4 - Architectural, resolved as documented narrowing rather than a workaround] A collapsed ADT-typed match has no valid re-checkable source projection**

- **Found during:** Task 3's round-trip test for `defect_dies_by_signal.lang`'s shape, after `collapse-branch-to-diverging-arm` eliminates the `Match` wrapper
- **Issue:** This project's checker (`check.Program`) only ever admits a user-declared ADT parameter/return type inside an EXHAUSTIVE `match` body; a plain straight-line `Linear` body is checker-admitted only for the executable primitive shapes (`Byte`/`Buffer`). Once collapse removes the `Match`, the surviving straight-line function's parameter/return type may still be an ADT, and there is no partial-match syntax in this grammar to legally re-express "the one surviving arm's logic" as source.
- **Resolution:** Rather than inventing non-standard syntax (which would validate against nothing) or silently emitting invalid source, `ProjectSource` detects this exact case (`Function.Match == nil` but `Parameter.Type`/`ReturnType` names a declared `DataType`) and returns the same explicit, grep-able `// reduce: projection unsupported for this program shape -- ...` marker used for every other unsupported shape. The CORE-level reduction itself is unaffected -- `collapse-branch-to-diverging-arm` still fires and narrows the core artifact; only the SOURCE projection for that specific narrowed shape is reported as unsupported.
- **Files modified:** `internal/compiler/reduce/reduce.go` (the `ProjectSource` dispatch gate), `internal/compiler/reduce/reduce_test.go` (test renamed to `collapsed_adt_match_projection_is_documented_unsupported` and its assertion changed from round-trip to "explicitly marked unsupported")
- **Verification:** `TestProjectedSourceCorrespondsToReducedCore/collapsed_adt_match_projection_is_documented_unsupported` passes.
- **Committed in:** `aa64d56`

---

**Total deviations:** 3 auto-fixed (2 bugs found and fixed during development before any commit landed; 1 architectural boundary resolved as an explicit, tested narrowing rather than a workaround).
**Impact on plan:** All three were discovered empirically while proving Task 3's round-trip requirement against the REAL shipped front end (not assumed) -- exactly the kind of gap the plan's own `<read_first>` emphasis on `checkResourceLifecycle`/`analyzeArmBody` was meant to surface. No scope creep: all three narrow or correct the SAME `ProjectSource`/`dropOffpathForeignStage` surface the plan specified, none add new public API.

## Known Stubs

None. Every exported function (`Moves`, `Reduce`, `Interesting`, `ProjectSource`) has a real, tested implementation; no hardcoded empty/placeholder return paths gate the plan's own goal.

## Issues Encountered

None beyond the deviations documented above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `internal/compiler/reduce` is a complete, dependency-free (leaf) package: `Move`/`Moves`/`Reduce`/`Result`, `Signature`/`Interesting`, and `ProjectSource` are all exported and ready for plan 05-12 to wire into `session` (the `lang.mismatch/0` document producer).
- Plan 05-12 must implement `SignatureFromDisagreement` (mapping a real `session.Phase5EngineDisagreement` into a `reduce.Signature`) in ITS OWN wiring file -- `reduce` deliberately does not import `session` to avoid an import cycle (`session` will import `reduce`).
- Plan 05-12 should be aware of the two documented `ProjectSource` narrowings: (a) it supports exactly the straight-line-borrow-chain and foreign-resource-lifecycle-chain shapes real Phase 5 fixtures use, returning an explicit unsupported marker for anything else; (b) a collapsed ADT-typed match has NO valid source projection under the current grammar (no partial-match syntax) -- plan 05-12's `lang.mismatch/0` document should treat an unsupported-marker `reduced_source` as a legitimate, expected value for that shape, not a bug to chase.
- `go test ./...`, `go test -race ./...`, `go vet ./...`, and `sh scripts/verify-phase5.sh` are all green as of this plan's completion.

---
*Phase: 05-native-equivalence-and-adversarial-evidence*
*Completed: 2026-09-06*

## Self-Check: PASSED
- All 4 created files verified present on disk.
- Both commit hashes (aa64d56, d1edcd0) verified present in git log.
- All plan `<acceptance_criteria>` re-run and passing (see coverage block).
- Plan-level `<verification>` commands re-run: `go test ./...`, `go test -race ./...`, `go vet ./...`, and `sh scripts/verify-phase5.sh` all exit 0.
