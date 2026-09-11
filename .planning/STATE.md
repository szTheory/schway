---
gsd_state_version: 1.0
milestone: M002
milestone_name: Interprocedural Semantic Spine
current_phase: 09
current_phase_name: Peer Re-Derivation and D-03-02 Closure
status: executing
stopped_at: Completed 09-10-PLAN.md — Phase 09 complete
last_updated: "2026-09-11T02:53:53.333Z"
last_activity: 2026-09-10
last_activity_desc: Phase 09 mandatory mid-phase gate adjudicated; computeLoanLastUses deletion AUTHORIZED for plan 09-09
state_head: 2598ef5757a7c65deb8ffdb8b78f070947abe643
progress:
  total_phases: 7
  completed_phases: 2
  total_plans: 28
  completed_plans: 28
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-09-07)

**Core value:** Give AI agents and human reviewers the shortest reliable path
from intent to sound, reproducible evidence without wasting iteration time or
hiding runtime costs.
**Current focus:** Phase 09 — Peer Re-Derivation and D-03-02 Closure
six dispatch sites and prove every M001 guarantee survives a function boundary.

**Durable context (survives context clears — read before re-deriving):**

- `.planning/LANGUAGE-MATURITY.md` — the language is far less expressive than
  the roadmap vocabulary implies: no arithmetic, no iteration, no Lang-to-Lang
  calls yet. Assurance stack ~60-70% built; language surface ~5-10%.
- `.planning/STANDING-VERDICTS.md` — already-researched verdicts (deps,
  anti-features, the six dispatch sites, why `-flto` is load-bearing).

## Current Position

Phase: 09 (Peer Re-Derivation and D-03-02 Closure) — EXECUTING
Plan: 10 of 10
Status: Ready to execute — computeLoanLastUses deletion AUTHORIZED by plan 09-08's gate
Last activity: 2026-09-10 — Completed 09-08-PLAN.md (mandatory mid-phase gate)

Progress: [--------------------] 0% (0/7 phases)

## M002 Phase Map

| Phase | Name | Requirements | Status |
|-------|------|--------------|--------|
| 07 | Calls, Signatures, and Call-Graph Refusal | 5 | Not started |
| 08 | Interprocedural Loan Liveness in `check` | 2 | Not started |
| 09 | Peer Re-Derivation and D-03-02 Closure | 6 | Not started |
| 10 | Trusted Interprocedural Oracle | 5 | Not started |
| 11 | Multi-Function Native Emission and Equivalence | 7 | Not started |
| 12 | `Result` Payloads | 2 | Not started |
| 13 | Agent Loop for Interprocedural Defects | 3 | Not started |

Phase numbering continues from M001 (which ended at Phase 06). Full detail:
`.planning/ROADMAP.md`.

## Performance Metrics

**Velocity:**

- Total plans completed: 77
- Average duration: 11 min
- Total execution time: 105 min

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 02 | 7 | - | - |
| 03 | 10 | - | - |
| 04 | 13 | - | - |
| 05 | 14 | - | - |
| 06 | 15 | - | - |
| 07 | 12 | - | - |
| 08 | 6 | - | - |
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
| Phase 05 P11 | 45min | 3 tasks | 3 files |
| Phase 05 P12 | 95min | 3 tasks | 8 files |
| Phase 05 P13 | ~50min | 2 tasks | 5 files |
| Phase 05 P14 | ~35 min (continuation) | 3 tasks | 5 files |
| Phase 06-agent-feedback-and-performance-ratification P01 | 40min | 3 tasks | 3 files |
| Phase 06 P02 | 55 min | 3 tasks | 6 files |
| Phase 06-agent-feedback-and-performance-ratification P04 | 55 min | 3 tasks | 4 files |
| Phase 06 P08 | 35min | 3 tasks | 4 files |
| Phase 06-agent-feedback-and-performance-ratification P11 | 18 min | 3 tasks | 3 files |
| Phase 06-agent-feedback-and-performance-ratification P03 | 70min | 3 tasks | 4 files |
| Phase 06 P05 | 55min | 3 tasks | 4 files |
| Phase 06 P12 | ~35 min | 3 tasks | 13 files |
| Phase 06 P06 | 65min | 3 tasks | 8 files |
| Phase 06 P13 | 55min | 3 tasks | 6 files |
| Phase 06 P07 | 95min | 3 tasks | 8 files |
| Phase 06-agent-feedback-and-performance-ratification P14 | 50 min | 3 tasks | 11 files |
| Phase 06 P09 | 55min | 3 tasks | 4 files |
| Phase 06 P10 | 55min | 3 tasks | 9 files |
| Phase 06 P15 | ~40 min | 3 tasks | 12 files |
| Phase 07 P01 | 55min | 3 tasks | 8 files |
| Phase 07 P02 | 90 min | 3 tasks | 11 files |
| Phase 07 P03 | 55 min | 3 tasks | 16 files |
| Phase 07 P04 | ~70 min | 3 tasks | 14 files |
| Phase 07 P05 | ~140min | 3 tasks | 10 files |
| Phase 07 P06 | 95min | 3 tasks | 12 files |
| Phase 07 P07 | 150 min | 3 tasks | 13 files |
| Phase 07 P08 | 64min | 2 tasks | 11 files |
| Phase 07 P09 | 90 min | 4 tasks | 11 files |
| Phase 07 P10 | 105min | 3 tasks | 9 files |
| Phase 07 P11 | 90 min | 3 tasks | 12 files |
| Phase 07 P12 | 55min | 3 tasks | 12 files |
| Phase 08 P01 | ~33min | 3 tasks | 5 files |
| Phase 08 P02 | ~50 min | 3 tasks | 2 files |
| Phase 08 P03 | 30 min | 3 tasks | 12 files |
| Phase 08 P04 | ~55min | 3 tasks | 3 files |
| Phase 08 P05 | 65min | 3 tasks | 8 files |
| Phase 08 P06 | ~45min | 3 tasks | 4 files |
| Phase 09 P01 | 33min | 3 tasks | 9 files |
| Phase 09 P02 | 45min | 3 tasks | 7 files |
| Phase 09 P03 | 55min | 3 tasks | 5 files |
| Phase 09 P04 | 55min | 3 tasks | 2 files |
| Phase 09 P05 | 55min | 3 tasks | 3 files |
| Phase 09 P07 | 55min | 3 tasks | 2 files |
| Phase 09 P06 | 50min | 3 tasks | 7 files |
| Phase 09 P09 | ~5h | 4 tasks | 10 files |
| Phase 09 P10 | 50min | 3 tasks | 5 files |

## Accumulated Context

### Decisions

Full decision log lives in PROJECT.md (Key Decisions) and the provenance-rich
`wiki/` research ledger. Per-phase decisions from M001 are archived with their
phase artifacts under `.planning/milestones/M001-phases/`.

Standing architectural commitments carried into M002:

- Vertical source-to-native slices, not layered compiler construction.
- Go 1.24 stdlib for Stage 0; readable C17 through Clang as the reversible
  native path.
- The deterministic interpreter is the semantic oracle.
- Syntax stays provisional; stable typed-core, diagnostic, and evidence
  identities are the asset.
- Trust-crossing facts are re-derived independently (`corevalidate`,
  `originvalidate`), never trusted from the producer.
- [Phase 07]: 07-01: DecodeInterface routes CheckSummary; /0 stays decodable but never admissible for a call — D-07-36 closes codex's HIGH finding that strict /1 decoding was unspecified and /0 dispatch was unwired
- [Phase 07]: Callable is publication safety (D-04-03), not export membership: PublishProblemsFor extracted, corevalidate summary peer wired at both replay sites, narrowed to core.origin_omitted (D-07-33) — 07-02 review-driven decisions D-07-31/D-07-32/D-07-33/D-07-20/D-07-21/D-07-22 landed exactly as ratified at the checkpoint
- [Phase 07]: core.LinearOperation.CalleeID added additive+omitempty per D-07-29; the parser change and the foreign-bare-call refusal relocation to check (D-07-40) landed together in Task 1 since they are inseparable within one green commit.
- [Phase 07]: 07-04: phase07 lane fires all three Phase 07 controls (phase07_in_process, phase07_lane, dispatch.recognized_not_executed) since the lane's own fixtures genuinely re-prove all three claims through the CLI path; cmd/lang main.go isPhase7Corpus dispatch added as Rule 3 blocking fix (not in plan's files_modified) since the lane is not CLI-observable without it
- [Phase 07]: 07-05: check.Program builds an immutable post-body signature table (reusing core.FunctionSignature, structurally body-free) and refuses a call to a non-Callable callee with core.callee_not_callable; corevalidate independently re-derives the same refusal via peerCallable. Discovered relay_escort_witness.lang checks clean while corevalidate refuses (core.move_while_borrowed) -- the interprocedural half of D-03-02, left open for Phase 08/09. — Two genuinely independent SEM-06 derivations (check's table consult, corevalidate's own peerCallable) satisfy QLT-08 without collapsing to one derivation read twice; the D-03-02 divergence is documented as a named, tested finding rather than hidden.
- [Phase 07]: 07-06: callgraph.Order (iterative white/gray/black DFS, no native recursion, roots = all declared functions, edges from CalleeID only) wired into check before it returns a core.Program; canonical rotation plus cross-cycle lexicographically-smallest witness selection makes the core.call_graph_cycle diagnostic ID deterministic regardless of discovery order (D-07-16/D-07-43); diagnostic bounded at 32 cycle_member causes while the traversal itself stays unbounded; spans projected from OpCall operation IDs via new check-side emission bookkeeping, no Span added to core.LinearOperation (D-07-35); three fault-injection seams (gray-vs-visited, self-edge, unresolved-edge-drop) each independently mutation-killed against a diamond/shared-leaf corpus a chain fixture could never kill.
- [Phase 07]: 07-07: corevalidate grows its own independently-written whole-program cycle peer (checkCallGraphAcyclic) over disjoint synthetic-only input space; run() split into structural-then-replay passes to place it correctly; remaining SEM-07 corpus (indirect/unreachable/match-arm/shadowing) landed; bilateral independent-disable proof and phase-wide completeness matrix close QLT-08 for Phase 07.
- [Phase 07]: [Phase 07] 07-08: ClosureDigest chained over callee summary digests only after callgraph.Order proves acyclicity; corevalidate independently re-derives the chain over its own postorder, matching the producer byte-for-byte wherever D-07-33's narrowed Callable scope already agrees; callee-changes-invalidates-caller mutation-killed on both sides; SEM-05 closed.
- [Phase 07]: Checkpoint auto-ratified: accepted all four proposed diagnostic code strings (check.call_argument_type_mismatch, check.call_return_type_unrepresentable, core.CallArgumentTypeMismatch, core.CallReturnTypeMismatch) and both ordered Causes shapes verbatim.
- [Phase 07]: 07-09: check.resolveCallBinding gates argument-type match and derives OpCall's TargetID.TypeID from the callee's declared return contract (fail-closed); corevalidate independently re-derives both refusals with no shared helper. Closes 07-VERIFICATION.md's single FAILED truth and 07-REVIEW.md CR-01.
- [Phase 07]: 07-10: CheckCommandFile consults corevalidate.Validate (refusing union, fixed precedence check-then-peer-then-originvalidate); interface export/core report a peer refusal as StatusInvalid with the peer's own code instead of tool.operation_failed. Closes 07-REVIEW.md CR-04/PVG-03.
- [Phase 07]: 07-11: consume-on-call closes CR-01/PVG-01 -- resolveCallBinding consumes call arguments (ability-decided), corevalidate independently re-derives from the core artifact alone; D-07-07 corrected, D-07-52 records the accepted implicit-transfer residual — Checkpoint auto-ratified under auto_advance/yolo mode
- [Phase 07]: 07-12: Foreign/Fails made closure-derived over the proven-acyclic call graph in both originvalidate (joinForeignReach/joinFails) and corevalidate (peerJoinForeignReach/peerJoinFails, independently written, no shared helper beyond core.ForeignReachConflict); closes 07-REVIEW.md CR-03/PVG-02 and incidentally IN-01 (index-coupling fix). Phase 07 has no open post-verification gaps.
- [Phase 08]: 08-01: interprocedural fact enters through derivePlaceLoans' forward canonicalization pass (D-08-07), never the backward transfer function; new post-acyclicity check.interprocedural_loan_liveness law closes D-03-02's interprocedural half on relay_escort_witness.lang, with both admission paths proven to agree via a mutation-tested differential (D-08-09).
- [Phase 08]: 08-02: fixed a real callgraph.Order ordering-direction bug in buildInterproceduralSummaries (its result is caller-before-callee, not callee-before-caller as previously assumed) -- required for UsesParam transitivity to hold; deriveFunctionUsesParam's fixpoint-loop shape makes the program-order-vs-reversed cost claim genuinely falsifiable (171.3x observed multiplier at k=512, matching spike S-006's 192x finding).
- [Phase 08]: 08-03: fixture-header + peerDivergenceExpected register the corevalidate residual for the two ACCEPT twins (check admits, corevalidate's still-intraprocedural loanChainIndex refuses via core.move_while_borrowed) as concrete input for Phase 09's peer re-derivation.
- [Phase 08]: 08-03: Pattern B's real .lang twin pair cannot demonstrate a differing end-to-end CLI verdict (computeLoanLastUses' summary-blind AST-shadow path refuses both members identically, ownership.move_while_borrowed) -- tests assert the TRUE observed outcome; the contract-driven backward gate itself stays proven at the checked-core level (08-02's own synthetic test).
- [Phase 08]: 08-04: loanLivenessFixpoint's cycle pre-walk converted to an explicit-stack DFS ported from callgraph.Order, its cycle refusal promoted to a coded check.cfg_back_edge diagnostic, and the worklist given a derived fail-closed bound (factor*blocks*(loans+1), the +1 floor a Rule 1 auto-fix over the plan's literal formula) with a named check.loan_liveness_bound_exceeded refusal, mutation-killed via loanLivenessBoundSeam and proven identity-stable against its own retuning — Success criterion 2 required the only genuinely iterative fixpoint in Phase 08 to terminate under a fail-closed bound with a named refusal rather than a hang or silent truncation; the literal blockCount*distinctLoanCount formula computes zero for any loan-free function and would have refused nearly every legal program, caught by reasoning through the existing suite before committing
- [Phase 08]: 08-05: interprocedural cost gate shipped -- five-shape synthetic call-graph corpus fits buildInterproceduralSummaries' own deterministic work counter to <=1.2 growth exponent against operation count; both hardcoded gate-eligibility chokepoints (measure.Demote, session.QLT02GateEligibleMetrics) widened together and proven to agree; a latent EvaluateBudget bug (literal-string gating instead of set membership) surfaced by that widening was Rule-1-fixed so the new metric genuinely blocks on its bound; ratified hard manifest row + lane:interprocedural-cost-scaling changed-risk lane land the gate in the feedback-budget ledger
- [Phase 08]: 08-06: Mid-phase gate adjudicated all four agenda items from re-run evidence (a/b carried as named debt, c/d resolved), scope-cut trigger did not fire, liveness law declared final; debt register expanded 5->9 items; manifest row re-ratified at gate's own commit; OWN-06/EFF-02 confirmed complete (found already marked by 08-03/08-05 ahead of the gate, a process-ordering deviation documented, not reverted since correct in substance).
- [Phase 09]: 09-01: corevalidate's own forward-derived interprocedural loan-carry peer (derivePeerLoanCarry/chainPeerLoanCarry) folded into the existing v.peerPostorder substrate; buildLoanChainIndex's OpCall consult retires both peerDivergenceExpected accept-twin divergences (D-08-40 resolved); bidirectional seeded-fault companion assertions prove the agreement load-bearing. Discovered and documented (not fixed) a pre-existing originvalidate OpCall-transparency defect this fix unmasks (PHASE-09-DEBT.md D-09-51) -- the full CLI's clean-check claim for both fixtures is not satisfied end-to-end, a separate validator's pre-existing bug.
- [Phase 09]: [Phase 09]: 09-02: generateCallGraphCorpus relocated verbatim from check's costcorpus_test.go to testsupport.GenerateCallGraphCorpus/CallGraphCorpusShapes (D-09-46), unblocking session/corevalidate consumption with zero production import of check; a third gate-eligible metric (peer_closure_recomputed_work_growth_exponent) declared at both chokepoints plus the QLT-02 vocabulary with no manifest row yet (D-09-28); spike-006's QLT-01 registry gap closed (waived, since AllShippedControlIDs predates Phase 08), making go test ./... unconditionally green for the rest of Phase 09 (D-08-43 resolved).
- [Phase 09]: [Phase 09] 09-03: peerDeriveOriginFacts adds an access-mode payload to peerReturnDerivesFromBorrow's existing forward walk; peerOriginContained performs a containment check (never a recomputation of originvalidate's backward combination law); peerForeignOriginOmitted independently re-derives the fourth class with a bounded, function-local walk. peerCallable now consults all four PublishProblemsFor classes, closing D-07-33; TestPeerDoesNotRederiveNarrowedClasses flipped to TestPeerRederivesFormerlyNarrowedClasses. D-09-21's fallback did not fire. OWN-08 stays Pending in REQUIREMENTS.md since 09-08's mid-phase gate also carries it.
- [Phase 09]: [Phase 09]: 09-04: synthetic-shape zero-divergence differential (package check, D-09-50's vehicle split) proves TRU-04 criterion 1 over diamond/deep-chain/dense/parser-shaped/forward shapes; discovered the five shapes never contain a borrow op, making the natural sweep provably vacuous, so Task 3's mutation-kill hand-built a twin_a_accept.lang-shaped fixture instead. TRU-04's 'recursion' shape settled as a cycle-peer witness-agreement differential (self/mutual/indirect), never a liveness one -- category-error disposition recorded in code (D-09-22).
- [Phase 09]: OWN-05 stays Pending after 09-05 (D-09-37): only check+corevalidate proven this plan; interp is Phase 10
- [Phase 09]: No new seam minted for core-layer convention override (D-09-35): existing ParameterContract.Mode closed-set decode check confirmed as the fail-closed control
- [Phase 09]: [Phase 09] 09-07: D-09-49 Q1 settled by enumeration (no fixture exhibits the interaction across all 6 move_while_borrowed fixtures); shadowPathReachableCodes={move_while_borrowed,borrow_conflict} proven by source scan + testOnlyForceUniformLoanJoin perturbation, not merely predicted
- [Phase 09]: [Phase 09]: 09-06: peer's cost sweep injects a genuine borrow into every generated function (raw corpus has none, D-09-04) via injectBorrowForLoanCarry, fits flat-linear against a self-derived 1300 milli-exponent bound (all five shapes ~999-1000), and mutation-kills it via peerClosureUnmemoizedSeamForTest (foldChain memo-write skip) on the 'forward' shape (0.999->1.396 exponent, 4.58x checks at n=512); Result.PeerConsultedFields() closes D-08-26's positive half, proven identical to check's set (D-09-30). No qlt02_budget_manifest.json row added (D-09-28 -- ratified at 09-08's gate).
- [Phase 09]: Deleted computeLoanLastUses; checkInterproceduralLoanLiveness is now the sole loan-liveness decision point (D-09-07/D-09-08/D-09-09)
- [Phase 09]: Pattern B's twin split does not materialize; recorded as new debt D-09-53 (pre-existing deriveFunctionUsesParam defect), not fixed in this plan
- [Phase 09]: OWN-05 split into OWN-05a (Phase 09, Complete) and OWN-05b (Phase 10, Pending); QLT-07 closed for the loan-liveness subset, ratified in 09-VALIDATION.md; OWN-09/TRU-04 requirement texts corrected in place

### Pending Todos

Three bounded pre-phase spikes, recorded in ROADMAP.md ("Pre-Phase Spikes"):

- **S-006 interprocedural liveness cost-scaling probe** — **ANSWERED
  2026-09-09, gate released** (`.planning/spikes/006-interprocedural-liveness-cost-scaling/`,
  VALIDATED). A memoized summary cache must be designed into Phase 08 from the
  start: memoized derivation is linear in program size on every call-graph
  shape (~5 work units/op, one derivation per function), while the charitable
  unmemoized arm is quadratic (76x the memoized cost at 512 functions, and the
  multiplier doubles with size) and the naive one exhausts a 4M-unit budget at
  32 functions. `loanLivenessFixpoint` itself extends directly — the new
  subsystem is the summary table plus its invalidation, not a new dataflow
  engine — reverse postorder over Phase 07's proven-acyclic call graph
  suffices, program-order bodies are a load-bearing invariant (reversed: 192x
  penalty, quadratic in body length), and a persistent CROSS-run summary cache
  must not be assumed (one leaf edit invalidates 92% worst / 43% mean on the
  parser-shaped graph).
- **S-007 recursive / mutually-recursive stress corpus** — runs alongside
  Phase 07, non-blocking; informs whether the call-stack ceiling is a fixed
  constant or a declared budget.
- **S-008 Nyquist fold-in cost measurement** — runs before Phase 09 planning;
  determines whether QLT-07 stays committed or becomes a declared stretch item.

Also: a `Result` payload / interprocedural-origin interaction probe must be
answered before Phase 12 is planned.

### Blockers/Concerns

- **D-03-02 (open past M001):** an exported borrow-derived return with no
  declared origin exports indistinguishable from a fully-owned return, in the
  INTERPROCEDURAL half of the hazard. The single-function half closed in
  Phase 3. Owned by M002's `OpCall` charter (D-05-32/D-05-33).
- **M001 ships without Lang-to-Lang calls.** Interprocedural `-O3` equivalence
  is outside M001's proof scope by construction, not by omission.
- Remaining Phase 2/4 debt is registered in the archived `*-DEBT.md` files and
  summarized in `.planning/milestones/M001-MILESTONE-AUDIT.md`: five open
  Phase 2 items (info/warning level), D-04-30 (storable/matchable `Result`
  values, deferred to M002), and D-04-31's four accepted residual limitations.
- Nyquist validation is compliant for Phases 1, 2, and 4; Phases 3, 5, and 6
  are not-validated. Overall status: partial.
- **Standing process rules** adopted after three M001 gate failures that shared
  one shape — a green test whose reachable input space omitted the hard case:
  mutation-kill every differential, interrogate what inputs a property test
  actually reaches, and drive the shipped binary on hand-written programs
  rather than only the gate's own corpus.
- Plan 09-09 HALTED before any code change: TestOwnershipSequenceExhaustive (check_test.go) calls analyzeStraightLine directly and asserts DiagnosticCode==ownership.move_while_borrowed/borrow_conflict synchronously against an oracle that computes those codes inline. This contradicts Task 1's must_have truth 'Lowering makes no loan-liveness decisions' and Task 2's acceptance criterion that this same test 'pass unchanged'. Empirically confirmed via a reverted experimental edit (length=2 case=198): removing the move_while_borrowed raise from analyzeStraightLine made production return DiagnosticCode:"" while the oracle still returned ownership.move_while_borrowed, failing assertSupportEqual. Needs an architectural decision (Rule 4) on how TestOwnershipSequenceExhaustive's contract is meant to change before the computeLoanLastUses deletion can proceed. Tree left green, no commits made.

### Roadmap Evolution

- Phase 1 edited: removed generic web-app MVP mode; retained tracer-first
  vertical planning.
- `OpCall` and interprocedural equivalence deferred out of M001 (Phase 5,
  D-05-32) and promoted to M002's lead charter.
- **M002 roadmap (2026-09-08):** adopted `research/SUMMARY.md`'s reconciled
  build order with **one documented departure** — SUMMARY's Phase 4 is split
  into Phases 10 and 11 at the ARCHITECTURE.md Stage 6 / Stage 7 boundary.
  Reason: as one phase it would carry 12 of 30 requirements and two brand-new
  subsystems (interpreter `Frame` stack; multi-function C emission) at M001's
  recorded 70-95 min/plan coordinated-change rate, and it would hold the
  milestone's central risk unadjudicated for its whole length. The "cgen must
  differential-test against a trusted interpreter" dependency SUMMARY itself
  names is a gate, so it is used as one.
- **Scope-cut order carried into the roadmap up front** (Phase 12 → QLT-07 →
  Phase 13 narrowing), with the explicit trigger: if Phase 08 or 09 exceeds
  ~2x its initial plan estimate, renegotiate Phase 12 out to M003 immediately
  rather than adding plans.
- **Two mandatory mid-phase gates** recorded, following M001 Phase 3's
  precedent: Phase 08 (interprocedural loan liveness) and Phase 11
  (multi-function C emission before alias-attribute call lowering).

## Deferred Items

| Category | Item | Status | Deferred At | Milestone |
|----------|------|--------|-------------|-----------|
| Runtime | Effects, async, actors, scheduling, managed heaps | Deferred | Initialization | Post-M001 |
| Ecosystem | Packages and first-party application kits | Deferred | Initialization | Post-M001 |

## Session Continuity

Last session: 2026-09-11T02:53:53.199Z
Stopped at: Completed 09-10-PLAN.md — Phase 09 complete
durable context recorded in LANGUAGE-MATURITY.md and STANDING-VERDICTS.md
Resume file: None
Next command: `/gsd-discuss-phase 07`

## Operator Next Steps

- Review `.planning/ROADMAP.md` (note the documented departure from the
  research-reconciled 6-phase order, and the pre-phase spike decision).
- S-006 is answered; Phase 08 planning is unblocked. Kick off S-007 alongside Phase 07.
- Discuss and plan Phase 07 with /gsd-discuss-phase.

### Gate override — Phase 08 decision coverage (2026-09-09)

`check.decision-coverage-plan` returned `passed: false` during `/gsd-plan-phase 08` with
`total: 32, covered: 0, uncovered: []` and the message "decisions could not be fully parsed".
This is a parser false-negative, not a dropped decision: 08-CONTEXT.md carries 40 `D-08-NN`
decisions and all 40 are cited across `08-01..08-06-PLAN.md` + `PHASE-08-DEBT.md` (verified by
direct set difference; zero uncovered). The parser rejects bullets whose `:` sits inside the bold
span, e.g. `- **D-08-01 (the falsification that decides the area):**`. Override accepted by the
user; verify-phase should re-surface it.
