---
gsd_state_version: 1.0
current_phase: 05
current_phase_name: Native Equivalence and Adversarial Evidence
status: executing
stopped_at: Completed 05-10-PLAN.md
last_updated: "2026-09-06T14:46:23.114Z"
last_activity: 2026-09-05
last_activity_desc: Phase 05 execution started
state_head: d1edcd06afd650448e037901743113ab98679025
progress:
  total_phases: 6
  completed_phases: 4
  total_plans: 47
  completed_plans: 43
  percent: 67
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-09-03)

**Core value:** Give AI agents and human reviewers the shortest reliable path
from intent to sound, reproducible evidence without wasting iteration time or
hiding runtime costs.
**Current focus:** Phase 05 — Native Equivalence and Adversarial Evidence

## Current Position

Phase: 05 (Native Equivalence and Adversarial Evidence) — EXECUTING
Plan: 11 of 14
Status: Ready to execute
Last activity: 2026-09-05 — Phase 05 execution started

Progress: ██████████ [███████░░░] 67%

## Performance Metrics

**Velocity:**

- Total plans completed: 30
- Average duration: 11 min
- Total execution time: 105 min

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 02 | 7 | - | - |
| 03 | 10 | - | - |
| 04 | 13 | - | - |
**Per-Plan Metrics:**

| Plan | Duration | Tasks | Files |
|------|----------|-------|-------|
| Phase 01 P01 | 10 min | 3 tasks | 18 files |
| Phase 01 P02 | 4 min | 3 tasks | 9 files |
| Phase 01 P03 | 16 min | 3 tasks | 14 files |
| Phase 02 P01 | 10 min | 2 tasks | 14 files |
| Phase 02 P02 | 9 min | 3 tasks | 6 files |
| Phase 02 P03 | 12 min | 2 tasks | 8 files |
| Phase 02 P04 | 13 min | 2 tasks | 4 files |
| Phase 02 P05 | 12min | 2 tasks | 8 files |
| Phase 02 P06 | 10 min | 2 tasks | 9 files |
| Phase 02 P07 | 9 min | 2 tasks | 8 files |
| Phase 03 P01 | 95 min | 3 tasks | 19 files |
| Phase 03 P02 | 70 min | 3 tasks | 26 files |
| Phase 03 P06 | 70 min | 3 tasks | 17 files |
| Phase 03 P03 | 100 min | 3 tasks | 6 files |
| Phase 03 P04 | 95 min | 3 tasks | 7 files |
| Phase 03 P05 | 95 min | 3 tasks | 7 files |
| Phase 03-borrowed-views-and-cfg-lifetimes P07 | 130min | 3 tasks | 13 files |
| Phase 03 P08 | 11 min | 2 tasks | 4 files |
| Phase 03 P09 | 22 min | 3 tasks | 8 files |
| Phase 03 P10 | 45 min | 3 tasks | 9 files |
| Phase 04 P01 | ~5h | 4 tasks | 25 files |
| Phase 04 P02 | 3h | 3 tasks | 16 files |
| Phase 04 P03 | ~2h | 3 tasks | 15 files |
| Phase 04 P04 | ~2h | 3 tasks | 19 files |
| Phase 04-fallible-resources-and-c-boundary P05 | ~2h | 3 tasks | 11 files |
| Phase 04 P06 | ~2h | 3 tasks | 9 files |
| Phase 04 P07 | 55 min | 3 tasks | 7 files |
| Phase 04 P08 | 30 min | 3 tasks | 3 files |
| Phase 04 P09 | 25 min | 2 tasks | 2 files |
| Phase 04 P10 | 20min | 2 tasks | 3 files |
| Phase 04 P11 | 35 min | 2 tasks | 3 files |
| Phase 04 P12 | 45min | 3 tasks | 5 files |
| Phase 04 P13 | 30 min | 4 tasks | 10 files |
| Phase 05 P01 | 110 min | 3 tasks | 8 files |
| Phase 05 P02 | ~70min | 3 tasks | 3 files |
| Phase 05 P03 | 35 min | 3 tasks | 2 files |
| Phase 05 P04 | 90 min | 3 tasks | 8 files |
| Phase 05 P05 | ~140min | 3 tasks | 8 files |
| Phase 05 P06 | 75 min | 3 tasks | 5 files |
| Phase 05 P07 | 180 min | 3 tasks | 4 files |
| Phase 05 P08 | 95min | 3 tasks | 13 files |
| Phase 05 P09 | 36min | 3 tasks | 6 files |
| Phase 05 P10 | ~110min | 3 tasks | 4 files |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md and the provenance-rich wiki research ledger.

- M001 uses vertical source-to-native slices.
- Stage 0 uses Go 1.24 stdlib and readable C17/Clang reversibly.
- The deterministic interpreter is the semantic oracle.
- Syntax remains provisional; stable typed-core/evidence identities are the asset.
- Phase 1 command/evidence contracts are executable; Phase 2 should extend the same vertical path with affine ownership and independent abilities.
- [Phase 02]: Linear type, place, operation, and point IDs use function-local semantic ordinals, never source offsets.
- [Phase 02]: Phase 1 match artifacts retain lang.core/0 and lang.execution/0 while owned linear artifacts use /1.
- [Phase 02]: Ownership token normalization lives at the parser boundary so the Phase 1 lexer remains unchanged.
- [Phase 02]: Box and Pair use a request-local package-private structural conjunction with no arbitrary production masks.
- [Phase 02]: Generic parser limits are depth 64 and 4,096 type nodes with declaration-bounded recovery.
- [Phase 02]: Legacy Error construction remains diagnostic/0; only repair-bearing ownership diagnostics select diagnostic/1.
- [Phase 02]: Straight-line shared loans expire immediately after their precomputed final use through a linear indexed schedule.
- [Phase 02]: Ownership checker work counts type nodes, the complete last-use scan, and the inspected forward prefix.
- [Phase 02]: Validator ability and transition authorization is independently implemented from checker and interpreter behavior, sharing only inert core records.
- [Phase 02]: The owned final return operation is the explicit final claim, and canonical validation work is exactly 16n+13.
- [Phase 02]: Each compile/run stdout/stderr stream has an independent 64 KiB max-plus-one bound and stable stage/stream truncation code.
- [Phase 02]: Native execution decodes stdout only as exactly one strict execution document; successful-run stderr is operational failure.
- [Phase 02]: Interpreter/O0/O3 equality covers complete ordered semantic execution facts while excluding physical observations.
- [Phase 02]: Legacy match Emit bytes remain frozen for Phase 1 evidence while native execution uses dedicated strict-JSON emission.
- [Phase 02]: Owned evidence selects lang.evidence/1 after independent core admission while Phase 1 evidence remains byte-identical on /0.
- [Phase 02]: SHA-256 is content identity only; the coordinated source/core lie is an expected escape, never a detected control.
- [Phase 02]: The bounded phase gate runs shared test/race/vet work once and reports five 20-sample warm distributions without ratifying an SLO.
- [Phase 02]: Generated linear C records executed operation events and serializes the returned runtime place through one counted 64 KiB output layer.
- [Phase 02]: The exact-one owned backend mutation must run at O0 and O3 and produce semantic mismatch exit 4 before control:backend.runtime_causality is admitted.
- [Phase 02]: Generated C uses one global ordinary-identifier allocator that preserves legacy source-derived names when unique and adds deterministic category/ordinal suffixes only for actual collisions.
- [Phase 02]: Compiler-spawned tool identity probes have independent 64 KiB-plus-one stdout/stderr bounds and five-second deadlines.
- [Phase 02]: `Buffer` grants `share` (OV-02-01), resolving a self-contradiction between the ability table and the shipped move-while-borrowed control fixture; `Buffer` remains noncopyable.
- [Phase 02]: `borrow` is gated on `AbilityShare` in the checker as defence in depth; the gate is unreachable from source today and guarded by a self-invalidating enumeration test.
- [Phase 02]: Loan liveness is transitive across reborrows and copies-of-loans in both admission layers, with the test oracle re-derived by fixed-point closure so it cannot mirror the production law.
- [Phase 02]: The formatter classifies an opening brace by the declaration keyword that opened the line, and that classification survives a trailing comment.
- [Phase 03]: [Phase 03-01]: A match with any arm body requires every arm to carry one; bare-arm-only matches remain the fully separate, untouched Phase 1 code path.
- [Phase 03]: [Phase 03-01]: A match function whose arms carry linear bodies now legitimately carries both core.Match and core.Linear at once — the third case Function.HasClosedBody/Match.HasBlocks had to learn.
- [Phase 03]: [Phase 03-01]: analyzeArmBody is a deliberate duplicate of analyzeStraightLine (not a shared helper), so the Phase 1/2 straight-line path carries zero risk from the CFG addition.
- [Phase 03]: [Phase 03-02]: An exclusive loan is gated on the same AbilityShare requirement a shared loan is gated on, not a new ability.
- [Phase 03]: [Phase 03-02]: corevalidate independently re-derives the five-row loan conflict matrix via running per-owner high-water-mark maps, a materially different mechanism from check.go's active-loan-set (D-12).
- [Phase 03]: [Phase 03-02]: D-02-09/D-07 closed — checkLinear refuses a straight-line linear function whose parameter shape has no native execution lowering (only Byte/Buffer) at check time, with ability derivation still running first.
- [Phase 03]: [Phase 03-06]: Origin paths are scoped to a function's own single parameter name — this reduced language has one parameter per function and no field-path-bearing executable shape. — Box/Pair are rejected at check.go's execution-admission gate before a body is ever analyzed, so a richer field-path grammar would have no honest source input to exercise this phase.
- [Phase 03]: [Phase 03-06]: A borrowed-view function's ability set is not separately re-derived with an explicit escape:false witness; PublicOrigin carries only Paths and Access. — corevalidate's existing type-fact loop recomputes abilities for every fact in linear.Types and requires an exact match for the same shape; a synthetic view TypeFact would desync from that independent recomputation.
- [Phase 03]: [Phase 03-03]: Backward worklist loan liveness is scoped to checkBranch's arm blocks only; checkLinear/analyzeStraightLine keep discoverLoanLastUses unchanged to protect two already-shipped regression suites.
- [Phase 03]: [Phase 03-03]: The uniform-join fixture pair is built on 03-01's per-arm alias isolation, not a shared pre-branch loan; only ONE fixture flips under the seeded uniform-join fault, matching the plan frontmatter's must_haves claim over the task prose's stronger 'both flip' claim.
- [Phase 03]: [Phase 03-04]: corevalidate.recomputeLoanEndpoints independently re-derives loan endpoints via a reachability closure + reduction, never check.go's iterative worklist fixpoint.
- [Phase 03]: [Phase 03-04]: loanChainIndex replaces both replayStraightLine and replayBlocks' O(n)-per-operation loansForPlace copy-and-rescan with a memoized, cycle-safe parent-pointer chain; LinearWorkLimit moves to 16*facts+14.
- [Phase 03]: [Phase 03-05]: discoverLoanLastUses (Phase 2, unchanged) is the sole law deciding admission in both checkLinear and checkBranch; loanLivenessFixpoint only produces the decorative LoanEndpoint facts checkBranch reports, never gating accept/reject.
- [Phase 03]: [Phase 03-05]: Mid-phase gate closed clean with two non-blocking debt items recorded in 03-DEBT.md (D-03-01: discoverLoanLastUses' own quadratic work stays uncounted; D-03-02: an exported borrow-derived return with no declared origin exports as if fully owned).
- [Phase 03]: verifyBorrowedCorpus reuses 03-04/03-05's existing lane constructors unmodified rather than duplicating mutation logic — keeps the Phase 3 gate an assembly of already-proven parts and avoids breaking already-shipped regression tests
- [Phase 03]: Guarded RecomputeOrigin's OpBorrowExclusive branch first-seen (symmetric with OpBorrowShared), closing CR-01/OWN-04's impossible-half gap — The hop nearest the returned place must decide the derived access mode; an earlier hop further from the return overwriting it was the exact defect 03-VERIFICATION.md found reachable through the shipped binary
- [Phase 03]: ValidatePublished recomputes an origin unconditionally for every function, refusing publication of an undeclared borrow-derived return with core.origin_omitted; the gate lives entirely on the publication path (ValidatePublished/interface export), never restored in check.go's admission path.
- [Phase 03]: Promoted RecomputeOriginPerReturn to the package's sole backward-walk site; RecomputeOrigin is now a pure conservative combiner over it, with AccessConflicting as an undeclarable sentinel for disagreeing arms.
- [Phase 04]: D-04-04 checkpoint resolved edge-based-now: typed failure is a two-successor control-flow edge in core, not a storable Result value; core.DataType.Alternatives is untouched.
- [Phase 04]: The tracer's ok payload and function parameter are both typed Byte (not an opaque Handle) to avoid generalizing cgen/interp's scalar writer this plan.
- [Phase 04]: The declared failure ADT's error value is always its first declared alternative in both engines, a documented narrowing pending real case-analysis syntax.
- [Phase 04]: checkFallibleLinear supports exactly one shape this plan (sole binding is the try-call, immediately returned); richer shapes are refused with check.foreign_call_shape_unsupported.
- [Phase 04]: checkFallibleLinear dispatches between the 04-01 tracer shape (single try binding, immediately returned) and the new resource-lifecycle shape (a sequence of try/discard bindings whose result is the function's own parameter), kept as two separate functions so the shipped tracer fixture stays byte-unaffected.
- [Phase 04]: A discard-acquired resource is never tracked for release this plan (documented narrowing): only a try_call whose ok/err edges diverge and is named by some OpRelease is tracked.
- [Phase 04]: corevalidate's release-order rederivation walks BACKWARD from every failure/return edge over the block/edge graph, independent of check's forward accumulation; tracked-acquisition membership uses the same 'named by some OpRelease' rule as check.go/cgen, not an edge-divergence heuristic.
- [Phase 04]: cgen's live-resource accounting is a genuine runtime ledger in generated C (a static array), not a compile-time-only literal, so the release-omission mutation is observable at the native layer.
- [Phase 04]: The release-omission mutation runner targets the LAST lang:release-site marker (the success block's own final release), since this project's fixtures always succeed at runtime and an earlier err-block release is unreachable.
- [Phase 04]: [Phase 04]: core.ForeignContract's Layout obligation describes the frozen private header's real record (payload, one field), not the {ok,value} ABI result struct cgen already emits inline; the two are distinct declared C shapes.
- [Phase 04]: [Phase 04]: InitializedState/Capture/Retention/Aliasing are compiler-derived fixed structural facts this phase (no new foreign C {} policy syntax), since the language has no closures/threads/partial-init; recorded in the sidecar manifest's unchecked_obligations list, not claimed proven.
- [Phase 04]: [Phase 04]: OpDefect (D-04-15) is scoped to a match arm's terminal position this plan (checkBranch/analyzeArmBody), never checkResourceLifecycle's straight-line resource shape -- the shipped witness needs no call surface at all, the stronger structural claim SC3 asks for.
- [Phase 04]: RunNative's foreign-source auto-wiring now resolves by declared symbol (native.ForeignSourcePathForSymbol) instead of hardcoding the first frozen TU — A second frozen foreign TU exists as of plan 05; hardcoding the first one silently broke linking any program declaring the newer symbol
- [Phase 04]: originvalidate/pathoracle now walk every terminator (return, typed failure, defect) via core.TerminatorKinds(), mutation-killed independently in both packages (D-04-29 closed)
- [Phase 04]: core.ForeignContract.Alias reuses the shared/exclusive access vocabulary; a foreign call declared to borrow/retain its argument is recognised as borrow-derived, refused with core.foreign_origin_omitted when undeclared (D-04-28 closed)
- [Phase 04]: originvalidate.ValidatePublished wired into lang check/run at the CLI command-file layer (not into RunInterpreter/RunNative themselves) to avoid regressing non-publication tests that drive the engines directly (WR-01 closed)
- [Phase 04]: discoverLoanLastUses now counts its own transitive-scan work (D-04-25 closed); retiring the quadratic derivation itself is re-recorded as dated debt D-04-26, carried to Phase 5
- [Phase 04]: Second-/third-stage typed-failure differential proven via genuine execution (test-double foreign object + interp.RunLinearBlockDirect), not a hand-constructed document; a structural core-artifact mutation approach was tried and reverted after corevalidate's own core.fail_reached_without_err_edge correctly refused it
- [Phase 04]: session.runNativeInputs derives lang run --engine=native's per-input Expect from the interpreter's own verdict, fixing two real bugs (multi-arm defect fixture, single-input nonlocal-exit fixture) discovered by driving the shipped binary on out-of-corpus programs
- [Phase 04]: corevalidate's terminal-block structural check ships as at-least-one/non-entry (core.terminal_block_unreachable), not the literal exactly-one form 04-VERIFICATION.md asked for, to keep discard_because.lang's legitimate ok/err merge valid
- [Phase 04]: Both new tests consume cgen.BannedOptimizerAttributes rather than a hard-coded literal, matching the existing header-only test's stated reason
- [Phase 04]: TestAttributeScanLaneCoversEveryInspectableLayer pins lane:foreign-no-unproven-attributes RecomputedWork at exactly 8, closing 04-VERIFICATION.md gap 2's artifact-count blind spot
- [Phase 04]: Used in-tree revert/run/restore for the mutation-kill demonstration instead of a detached git worktree, per this session's no-worktree constraint
- [Phase 04]: Pinned TestAcyclicChainsStillValidateUnderCycleGuard's accepting-path Checks constant to 403, measured from the actual pre-guard build
- [Phase 04]: Promoted okEdgeInto's plural map to the primary representation (assumption-delta promote): the single-edge case is the len==1 degenerate branch, not a separate side table — Keeps exactly one source of truth for which ok edge(s) target a block; an add-alongside side table would reintroduce the last-writer-wins collapse this plan closes
- [Phase 04]: corevalidate audits ForeignContract.Symbol's identifier shape (foreign.symbol_not_identifier) and cgen independently refuses the same hostile input on its three Validate-bypassing EmitForeign* entry points, closing 04-VERIFICATION.md gap 2's C-injection hole
- [Phase 04]: [Phase 04] 04-13: full C-identifier shape for allocator/unwind/nonlocal_exit (DD-04-13-01), a third independent predicate at check.go's source-admission layer, plus comment-safety/C-type-expression audits in corevalidate and cgen for every remaining spliced ForeignContract field -- closing 04-VERIFICATION.md gap 2b's code-injection hole at the C boundary.
- [Phase 05]: selectsByPointerLowering gates on function.PublicOrigin == nil to keep declared borrow-return functions (identical op shape) on their existing lowering path — public_view_mixed_access.lang structurally matches the new by-pointer predicate; PublicOrigin is the only structural fact that distinguishes them
- [Phase 05]: computeLoanLastUses (loanLivenessFixpoint) is the sole liveness law deciding admission in both analyzeStraightLine and analyzeArmBody; discoverLoanLastUses retired after a recorded zero-divergence shadow run over 230,692 admission-site comparisons (D-04-26/D-03-01 closed).
- [Phase 05]: [Phase 05-03]: All three tasks were authored and committed as a single cohesive file/commit rather than three incremental commits since the tasks share helpers in one small file.
- [Phase 05]: [Phase 05-03]: classifySanitizerRun is a pure (exitCode, stderrText) function with no process I/O, so the unclassified-abort false-green mode and path/address-exclusion property are unit-testable on synthetic input.
- [Phase 05]: [Phase 05-04]: deriveAliasFacts additionally excludes function.PublicOrigin != nil, mirroring selectsByPointerLowering's own identical guard, required for the checker/cgen agreement test to hold on public_view_mixed_access.lang — Same operation shape as the Phase 5 fixture but a distinct, already-shipped semantic category
- [Phase 05]: [Phase 05-04]: cgen never imports check's AliasFact type -- restrict is emitted unconditionally inside emitLinearBorrowedByPointer (only ever reached when the fact holds by construction), reading the justification loan ID directly off function.Linear.Operations[0].LoanID — TestAliasFactAgreesWithByPointerSelection proves selectsByPointerLowering and deriveAliasFacts decide the identical condition, so no data dependency between check and cgen is needed (D-12)
- [Phase 05]: [Phase 05-04]: corevalidate.recomputeAliasJustifications feeds a synthetic single-block core.LinearBody into the EXISTING recomputeLoanEndpoints rather than adding a parallel mechanism for straight-line bodies — recomputeLoanEndpoints already refuses when Linear.Blocks is empty, which every straight-line function's serialized core is by design
- [Phase 05]: [Phase 05-05]: The bounded closure's 'single-level branch' axis is checkFallibleLinear's own ok/err block fork (a try/discard call), never a match{} construct -- check.Program dispatches per-function on hasTryCall, so a straight-line function calling a foreign symbol is already block-graph-shaped in core.
- [Phase 05]: [Phase 05-05]: phase5ForeignChainSource declares the real frozen lang_res_open symbol, not a synthetic name -- native.ForeignSourcePathForSymbol only resolves the closed set of real frozen translation units, so a made-up symbol would check-admit but never link natively.
- [Phase 05]: [Phase 05-06]: axis:exit-status-signal is backed by two new independently-set execution.Execution fields (ExitSignaled/ExitSignal) rather than a value derived from Outcome.Kind, so all five Phase 5 comparator axes are independently seedable.
- [Phase 05]: [Phase 05-06]: Phase5CompareDiagnosticIDs is a separate function from Phase5CompareEngines under its own control:diagnostic.reject_program_id_equivalence, since reject-programs never execute and never produce an execution.Execution document.
- [Phase 05]: [Phase 05-06]: TestLTOTierIsNotInert captures compiled binary bytes via a new Runner.LastBinary() recorder rather than reusing Run's temp directory, which is deleted before Run returns.
- [Phase 05]: [Phase 05-07]: deriveAliasFacts and cgen.selectsByPointerLowering are mathematically equivalent for every checker-admitted straight-line program in this language's current grammar — Proven both analytically and empirically (8+ constructed .lang shapes); the false-restrict fixture uses a SHARED borrow chain (deriveAliasFacts's own exclusive-first gate makes it trivially zero-fact) rather than a discovered divergent case, since none exists
- [Phase 05]: [Phase 05-07]: added cgen.selectsByPointerLoweringSharedOnly/emitLinearBorrowedByPointerPlain, an additive by-pointer-without-restrict lowering path mutually exclusive with the existing exclusive-borrow gate — Required so the mutation runner's fail-closed marker-count guard has a real marker to attack in normally-compiled, unmutated generated C; verified against the full corpus (including 05-05's enumerated closure) for zero regression
- [Phase 05]: Empirically falsified posix_memalign+free as an ASan allocator-mismatch mechanism; used C++ operator-new/delete (_Znwm/_ZdlPv) linked via -lc++ instead, verified on-host to produce a genuine alloc-dealloc-mismatch report.
- [Phase 05]: Collapsed a three-distinct-foreign-symbol retained-pointer fixture design to one call-count-dispatched symbol after confirming cgen only calls the first declared symbol in a resource-lifecycle sequence.
- [Phase 05]: D-05-41 (LTO lane one-fixture sample) accepted as recorded debt, landing phase Phase 5+ — TestPhase5CorpusThreeEngineAgreement and TestLTOTierIsNotInert already cover the correctness gap; widening now is optional
- [Phase 05]: [Phase 05-10]: drop-offpath-foreign-stage and collapse-branch-to-diverging-arm rebuild the affected core region from scratch via the checker's own ID-generation algorithms, rather than surgical graph patching -- an earlier surgical version left a dangling release reference.
- [Phase 05]: [Phase 05-10]: ProjectSource explicitly reports (never fabricates) an unsupported-shape marker for a collapsed ADT-typed match, since this project's checker admits a declared ADT type only inside an exhaustive match body and there is no partial-match syntax.

### Pending Todos

None yet.

### Blockers/Concerns

- Baseline machines for ratified feedback budgets remain to be chosen before Phase 6.
- Nine Phase 02 debt items are carried into Phase 3; see `.planning/phases/02-owned-values-and-abilities/02-DEBT.md`. Two are deadline-bearing: D-02-05 (`__LANG_` → `_LANG_` before any further C artifact is frozen) and D-02-03 (the Θ(N²) checker cost, which OWN-03's CFG liveness should remove anyway).
- Phase 3 stays one phase. The plan-checker's split recommendation is answered by a mandatory
  mid-phase gate after wave 5 (03-05, the OWN-03 terminal) plus 03-06's re-pin to a parallel
  wave-2 track. See ROADMAP §Phase 3, "Decision (2026-09-04)".
- Process debt adopted as standing rules after three gate failures shared one shape — a green test whose reachable input space omitted the hard case: mutation-kill every differential, interrogate what inputs a property test actually reaches, and drive the shipped binary on hand-written programs rather than only the gate's own corpus.
- 03-03 flagged three documented deviations (straight-line scope, fixture-pair interpretation, single-fixture-flip) for human review before 03-04/03-05 build further on core.LoanEndpoint's current arm-block-only population.
- [Phase 03-04] carried forward from 03-03, unresolved: checkLinear/analyzeStraightLine still use the old discoverLoanLastUses liveness law, not checkBranch's new backward worklist -- two liveness derivations coexist in the checker; 03-05 is the phase's designated gate for adjudicating this.
- [Phase 03-05] Two mid-phase-gate debt items recorded in 03-DEBT.md, non-blocking: D-03-01 (D-05/D-02-03's quadratic-cost fix never reached the admission-deciding code path) and D-03-02 (an exported borrow-derived return with no declared origin exports indistinguishable from a fully-owned return). Both should be closed before Phase 4 introduces cross-function calls.

### Roadmap Evolution

- Phase 1 edited: removed generic web-app MVP mode; retained tracer-first vertical planning

## Deferred Items

| Category | Item | Status | Deferred At | Milestone |
|----------|------|--------|-------------|-----------|
| Runtime | Effects, async, actors, scheduling, managed heaps | Deferred | Initialization | Post-M001 |
| Ecosystem | Packages and first-party application kits | Deferred | Initialization | Post-M001 |

## Session Continuity

Last session: 2026-09-06T14:46:14.601Z
Stopped at: Completed 05-10-PLAN.md
Resume file: None
