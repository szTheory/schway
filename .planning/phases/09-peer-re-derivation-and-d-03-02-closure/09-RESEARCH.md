# Phase 09: Peer Re-Derivation and D-03-02 Closure - Research

**Researched:** 2026-09-10
**Domain:** Independent (N-version) static-analysis peer derivation in a Go compiler; loan-liveness dataflow restructuring; differential/mutation-kill test gating
**Confidence:** HIGH

## Summary

This phase has no external-library research surface — it is entirely an
in-repo structural exercise: extend `corevalidate`'s existing forward-set-
propagation peer substrate to re-derive interprocedural loan liveness and all
four `Callable` refusal classes, prove peer/producer agreement is load-bearing
via a seeded fault with a companion assertion, and collapse `check`'s
loan-liveness admission from two call sites into one. Every locked decision in
09-CONTEXT.md was re-verified against the shipped tree this session by reading
the actual source (not just grepping for strings), and every decision held up
**except two planning-document line/path references not yet flagged as stale**,
recorded below under Assumptions/Open Questions.

The single highest-risk structural fact this research surfaces, not fully
named in CONTEXT.md: **`computeLoanLastUses`'s last-use index is consumed
*inline*, during the same per-binding lowering loop that also decides
`ownership.move_while_borrowed` and `ownership.borrow_conflict`** (via the
`activeLoans`/`expiringLoans` state machine seeded from `loanUses[index]` at
each borrow binding). The three other `ownership.*` codes emitted from that
same loop (`use_after_move`, `borrow_requires_share`, `transfer_requires_take`)
do **not** depend on `computeLoanLastUses`'s timing index at all — they are
immediate per-binding facts (initialization state, ability presence). This
narrows OWN-09's restructure to exactly two liveness-timing-dependent codes,
not all five, and that narrowing is the load-bearing fact a plan must state
explicitly or it will either over-scope (rewrite five codes' worth of logic)
or under-scope (miss that `move_while_borrowed`/`borrow_conflict` genuinely
need a new post-assembly timing source).

The second highest-risk fact this research surfaces, not named in CONTEXT.md
at all: **`generateCallGraphCorpus` (D-09-23's prescribed corpus vehicle) is
an unexported function in a `_test.go` file inside package `check`**
(`internal/compiler/check/costcorpus_test.go:50`). Go's package-visibility
rules make this uncallable from `internal/compiler/session`'s `session_test`
package, external-test-package status notwithstanding — unexported symbols in
`_test.go` files are invisible outside their own package, full stop. D-09-23's
"same generator, disjoint consumers" therefore cannot be satisfied by direct
reuse; the plan must either export the generator from a non-test file in
`check`, or relocate/duplicate it into a shared test-support location. This is
a concrete decomposition question the planner must resolve, not an
implementation detail to discover at execution time.

**Primary recommendation:** decompose Phase 09 into (at minimum) three ordered
plans honoring D-09-10/D-09-27/D-09-28's hard ordering: **Plan A** builds the
peer's liveness derivation (OWN-07) plus the D-07-33 four-class closure
(OWN-08) plus the differential, written red-first against today's known
divergence, going green in the same plan (TRU-04) — because D-09-10 requires
the peer to land *fully first* and D-09-27 requires the differential in the
*same plan* as the peer. **Plan B** does the OWN-09 restructure and the
`computeLoanLastUses` deletion, strictly after Plan A's differential is green.
**Plan C** (can run in parallel with A/B once the corpus generator question is
settled) does the third chokepoint widening (D-09-28), OWN-05's structural
proofs, QLT-07's closure document, and the debt register. The chokepoint
widening must land **before** Plan A's peer manifest row claims to be a hard
gate, so in practice it either lands as Plan A's own first task or as a
zero-th plan.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Interprocedural loan-liveness derivation (producer) | Admission / `check` (post-assembly) | — | Already the sole producer of the accepted-program answer; OWN-09 moves its decision point, not its ownership |
| Interprocedural loan-liveness derivation (peer) | Admission / `corevalidate` | — | Reachability-closure peer, structurally independent, never imports `check`/`callgraph` (enforced by 4 shipped guard tests) |
| Call-graph cycle refusal | Admission / `check`'s `callgraph` package | `corevalidate` (own DFS) | Runs before liveness in both layers; not this phase's scope to change, only to feed TRU-04's cycle-peer differential |
| Callable/publication-safety refusal (4 classes) | Admission / `originvalidate` (producer) | `corevalidate` (peer, closing this phase) | `originvalidate` is itself a third derivation layer the peer must differ from, not just from `check` |
| Diagnostic identity/ordering | Admission / `check` + `diagnostic` package | — | `Code+Span+Causes` hash is computed at diagnostic construction time regardless of which pass emits it — moving *when* a diagnostic is built changes observable identity even with no code-string change |
| Cost/growth-exponent gating | Evidence / `session` package | `measure` package | Two independent chokepoints already exist by design (D-08-31/32); peer's own bound is a third, deliberately disjoint from `check`'s |
| Differential/divergence gating | Evidence / `session` package (`session_peer_gate_test.go`) | — | Single trusted harness; D-09-23 forbids a second, independently-truthed one |

## Standard Stack

Not applicable — no new external dependencies. This phase is a pure
in-repo extension of existing Go packages (`corevalidate`, `check`,
`originvalidate`, `session`). No `npm`/`pip`/`cargo` installs occur.

## Package Legitimacy Audit

Not applicable — this phase installs no external packages.

## Architecture Patterns

### System Architecture Diagram

```
                 ast.Program
                     │
                     ▼
        ┌────────────────────────┐
        │  check.Program (lower) │   <- OWN-09 target: NO ownership
        │  checkLinear/checkBranch│      decisions during this stage
        │  emits core.LinearOp*  │      after the restructure lands
        └───────────┬────────────┘
                     │ core.Program (assembled, per-function admitted
                     │ for everything EXCEPT loan-liveness timing)
                     ▼
        ┌────────────────────────────────────┐
        │ checkCallGraphAcyclic (callgraph)   │  cycle refusal, already
        └───────────┬──────────────────────────┘  runs before liveness
                     │ acyclic core.Program
                     ▼
        ┌────────────────────────────────────┐
        │ buildInterproceduralSummaries       │  callee-before-caller,
        │ (walks callgraph.Order BACKWARD —   │  produces UsesParam /
        │  D-08-42: Order is caller-before-   │  ReturnsBorrowOfParam
        │  callee, not callee-before-caller)  │  per function
        └───────────┬──────────────────────────┘
                     │ interproceduralSummaryTable (real, not zero-value)
                     ▼
        ┌────────────────────────────────────┐
        │ ONE loan-liveness decision point    │  <- OWN-09's destination:
        │ (extends today's                    │     checkInterproceduralLoan-
        │  checkInterproceduralLoanLiveness   │     Liveness to ALSO decide
        │  to cover intraprocedural-only      │     move_while_borrowed /
        │  move_while_borrowed too)           │     borrow_conflict, not just
        └───────────┬──────────────────────────┘   the interprocedurally-
                     │ diagnostics (deterministic order — D-09-13)          extended case
                     ▼
              check.Result{Program, Diagnostics}

   ── in parallel, independently ──

        core.Program (same artifact, handed to a SEPARATE validator)
                     │
                     ▼
        ┌────────────────────────────────────┐
        │ corevalidate.Validate               │  imports ONLY core (+ stdlib)
        │  - checkCallGraphAcyclic (own DFS)  │  4 shipped import-guard tests
        │  - v.peerPostorder / peerSignatures │  forbid check/ast/originvalidate/
        │  - buildLoanChainIndex (OWN-07:     │  callgraph imports
        │    must consult peer liveness bits  │
        │    before propagating over OpCall)  │
        │  - peerCallable (OWN-08: extend to  │
        │    all 4 PublishProblemsFor classes)│
        └───────────┬──────────────────────────┘
                     │ corevalidate.Result{Valid, Problems, PeerSignatures}
                     ▼
        session_peer_gate_test.go's exact-set differential
        (TestNoUndeclaredCheckPeerDivergenceAcrossCorpus +
         peerDivergenceExpected) — TRU-04's vehicle
```

### Recommended Project Structure

No new packages. New files, per D-09-04's discretion:

```
internal/compiler/corevalidate/
├── corevalidate.go                     # existing; buildLoanChainIndex gets
│                                        #   the OWN-07 consult-before-propagate
│                                        #   edit; peerCallable/derivePeerSignature
│                                        #   get OWN-08's 3 new class checks
├── corevalidate_peer_liveness.go       # NEW (D-09-04's suggested sibling file):
│                                        #   forward-propagation liveness bits,
│                                        #   doc comment distinguishing reused
│                                        #   SUBSTRATE from reused DERIVATION
├── corevalidate_seams.go or inline     # NEW unexported seam(s) for D-09-25's
│                                        #   corevalidate-side companion assertion
internal/compiler/check/
├── check.go                            # analyzeStraightLine/analyzeArmBody's
│                                        #   activeLoans/expiringLoans lose their
│                                        #   move_while_borrowed/borrow_conflict
│                                        #   emission; computeLoanLastUses deleted;
│                                        #   checkInterproceduralLoanLiveness
│                                        #   extended to cover the intraprocedural-
│                                        #   only case
internal/compiler/session/
├── session_peer_gate_test.go           # peerDivergenceExpected: 2 entries
│                                        #   retired (paths corrected below);
│                                        #   fed by the corpus generator
├── qlt02_budget_manifest.json          # + new peer-cost-metric row
├── session_phase6_budget.go            # QLT02GateEligibleMetrics() widened
│                                        #   to 3 (measure.GateEligibleMetrics()
│                                        #   too, disjoint chokepoint)
├── risk_lanes.json                     # + peer's own cost-scaling lane, OR
│                                        #   extend lane:interprocedural-cost-
│                                        #   scaling's rationale (planner
│                                        #   discretion)
.planning/phases/09-.../
├── 09-VALIDATION.md                    # NEW — required output, see below
├── PHASE-09-DEBT.md                    # NEW — written at planning time (D-09-44)
```

### Pattern 1: Forward set-propagation liveness bit, mirroring the existing peer idiom

**What:** A boolean (or small enum) fact computed by a single forward pass
over `function.Linear.Operations`, propagated through `OpMove`/`OpCopy`/
`OpBorrowShared`/`OpBorrowExclusive` hops via a `map[string]bool` keyed by
place ID, read at `OpReturn`.

**When to use:** Exactly `peerParameterEscapesOwned` and
`peerReturnDerivesFromBorrow`'s shape — this is OWN-07's own derivation
idiom, verified in tree.

**Example (verified, not hypothetical — this is the actual shipped function
OWN-07 extends):**
```go
// Source: internal/compiler/corevalidate/corevalidate.go:2099 (verified read this session)
func peerReturnDerivesFromBorrow(function *core.Function) bool {
	if function.Linear == nil {
		return false
	}
	paramTrace := map[string]bool{function.Parameter.ID: true}
	derived := make(map[string]bool)
	for _, operation := range function.Linear.Operations {
		switch operation.Kind {
		case core.OpBorrowShared, core.OpBorrowExclusive:
			if paramTrace[operation.SourceID] || derived[operation.SourceID] {
				derived[operation.TargetID] = true
			}
		case core.OpMove, core.OpCopy:
			if paramTrace[operation.SourceID] {
				paramTrace[operation.TargetID] = true
			}
			if derived[operation.SourceID] {
				derived[operation.TargetID] = true
			}
		}
	}
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpReturn && derived[operation.SourceID] {
			return true
		}
	}
	return false
}
```
OWN-07's new liveness bits are the **same shape** (a forward-propagated
membership set), but must additionally consult `v.peerPostorder`/
`v.peerSignatures` (callee-before-caller order, already built for the Phase 07
cycle peer) so an `OpCall` hop can ask "does the callee's declared contract
say this loan extends past the call" — `buildLoanChainIndex`
(`corevalidate.go:1133-1148`, verified read this session) today has **no**
such consultation: it sets `idx.parent[operation.TargetID] = SourceID` for
**every** operation with a non-empty `TargetID`, unconditionally, regardless
of `Kind`. This is the exact defect D-09-03 names.

### Pattern 2: Peer-side reachability closure, structurally opposite the producer's worklist

**What:** `check`'s `loanLivenessFixpoint` (check.go:1905) is a converging
backward monotone worklist. `corevalidate`'s `recomputeLoanEndpoints`
(corevalidate.go:1270-1403, verified read this session in full) is a
**one-shot** reachability closure (`blockReach`, BFS-once-per-block) plus a
static reduction over a fully-materialized use relation — never a worklist,
never iterated to convergence. This is the actual, already-shipped precedent
for "reachability closure, not bounded worklist" that D-09-28 says the peer's
cost curve is shaped like. The peer's new liveness-bit derivation for OWN-07
should follow the SAME non-iterative, single-pass shape (materialize once,
reduce once) to stay consistent with the package's already-established
independence argument — an iterated worklist inside `corevalidate` would
blur the very distinction D-09-28's "distinct cost curve" claim rests on.

**Example (verified excerpt, corevalidate.go:1215-1233):**
```go
func blockReach(blocks []core.Block, index map[string]int) []map[int]bool {
	reach := make([]map[int]bool, len(blocks))
	for i, block := range blocks {
		visited := make(map[int]bool, len(blocks))
		queue := append([]string(nil), block.Successors...)
		for len(queue) > 0 {
			id := queue[0]
			queue = queue[1:]
			successorIndex, known := index[id]
			if !known || visited[successorIndex] {
				continue
			}
			visited[successorIndex] = true
			queue = append(queue, blocks[successorIndex].Successors...)
		}
		reach[i] = visited
	}
	return reach
}
```

### Pattern 3: Companion-assertion seeded fault (the exact shape D-09-25 requires)

**What:** An unexported package-level `bool` seam, engaged by exactly one
same-package test, that disables ONE peer's refusal and asserts the OTHER
peer's independently-derived answer for the same fact is unchanged (still
refuses) — proving the two code paths are separate, not proving mere output
change.

**Example (verified, the actual shipped precedent, check_test.go:3322-3343):**
```go
// Source: internal/compiler/check/check_test.go:3314-3343 (verified read this session)
func TestVerifyCallableRefusalSeamCheckDisabledCorevalidateStillRefuses(t *testing.T) {
	defer func() { verifyCallableRefusalSeam = false }()
	verifyCallableRefusalSeam = true
	result := Program(mustParseProgram(t, readPhase07Fixture(t, "call_uncallable_callee.lang")))
	verifyCallableRefusalSeam = false
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected check's seam to admit the call, got %+v", result.Diagnostics)
	}
	coreResult := corevalidate.Validate(result.Program)
	if coreResult.Valid {
		t.Fatal("expected corevalidate to independently still refuse, got Valid == true")
	}
	found := false
	for _, problem := range coreResult.Problems {
		if problem.Code == core.CalleeNotCallable {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected %s, got %+v", core.CalleeNotCallable, coreResult.Problems)
	}
}
```
D-09-25 requires the OPPOSITE direction too: a `corevalidate`-side seam that
disables the peer's new liveness/endpoint refusal, with `check`'s
`checkInterproceduralLoanLiveness` (or the restructured single pass, post
OWN-09) asserted to still refuse. Both seams are unexported package-level
`bool`s — `corevalidate`'s scoping is enforced by the SAME four import-guard
tests that block `callgraph`/`check`/`ast`/`originvalidate` imports, which is
precisely D-09-25's "the seam's very scoping is the proof the code paths are
separate" argument, verified structurally sound: a `corevalidate`-package
seam is physically unreachable from `check`'s package.

### Anti-Patterns to Avoid

- **Extracting a shared "loan liveness" helper package for `check` and
  `corevalidate` to both call.** D-09-02 and ARCHITECTURE §4 explicitly reject
  this: "a shared implementation isn't a peer at all, it's a single point of
  failure wearing two names." Independence is proven by derivation method
  (forward set-propagation vs. backward memoized walk) plus import boundary,
  never by file layout.
- **Treating OWN-09 as "delete `computeLoanLastUses`'s algorithm."** Verified
  in tree: there is one algorithm (`loanLivenessFixpoint`) with two callers.
  The deletion target is the summary-blind CALL SITE and its
  `shadow:place:*` synthesis, not a second law.
- **Widening only one of `session.QLT02GateEligibleMetrics()` /
  `measure.GateEligibleMetrics()`.** Both must gain the peer's new metric name
  together, or `TestGateEligibleMetricSetsAgreeAcrossChokepoints` fails by
  design (verified: this test exists precisely to catch a one-sided widening).
- **Recomputing `originvalidate`'s backward combination law inside the
  peer for D-07-33's remaining three classes.** D-09-19 (corrected from the
  original researcher claim) requires a CONTAINMENT check — compare the
  peer's own forward-derived facts against the DECLARED origin — never a
  re-implementation of `RecomputeOrigin`/`RecomputeOriginPerReturn`'s
  backward combination rule (`originvalidate.go:234`, `:133`).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Peer signature/liveness substrate | A new `corevalidate` subpackage or a fresh postorder/memoization scheme | Extend `v.peerPostorder` (corevalidate.go:138, appended :472) and `v.peerSignatures` (:126, refined :505-547) | D-09-01: this substrate already exists for the structurally analogous Phase 07 signature/`Callable` peer; a new one doubles the ordering surface Phase 09's own risk statement warns about |
| Seeded-fault discrimination mechanism | An external mutation-testing tool (`go-mutesting`, `gremlins`) | Unexported package-level bool + deferred restore, matching `loanLivenessBoundSeam`/`callReturnTypeDerivationSeam`/`verifyCallableRefusalSeam` | D-09-26: mutation tools produce a broad score, not the single targeted witness-carrying fault criterion 2 names; also `go-mutesting` is inactive upstream and `gremlins` self-documents as not scaling (REQUIREMENTS.md Out of Scope) |
| Cross-peer field-set agreement check | A hand-rolled deep comparison for each new disclosure test | `reflect.DeepEqual` / the existing `structuralFieldsEqual` and `ClosureDigest` byte-equality precedent | Proves outputs agree without constraining implementation — the exact instrument D-09-30 needs |
| Peer's own cross-function relation lookup | A new adjacency/postorder build inside the new liveness-bit function | `v.peerAdjacency`/`v.peerPostorder`, already built by `checkCallGraphAcyclic` as its own DFS byproduct | Re-deriving adjacency a second time inside the same package is wasted work and a second place for the two structures to silently disagree |

**Key insight:** every piece of substrate OWN-07/OWN-08 need already exists in
`corevalidate` from Phase 07's signature/`Callable` peer work. The genuinely
new work is narrow: (1) a new forward-propagated liveness bit consulted by
`buildLoanChainIndex` before it crosses `OpCall`, and (2) an access-mode
payload added to the existing `peerReturnDerivesFromBorrow`-shaped walk. The
temptation this phase must resist is building infrastructure that already
ships.

## Runtime State Inventory

Not applicable — this is not a rename/refactor/migration phase. No stored
data, live service config, OS-registered state, secrets, or build artifacts
carry a renamed identifier. (checked explicitly: no rename verbs appear in
the phase description or CONTEXT.md; the phase adds/deletes code paths and
diagnostic emission timing, not identifiers.)

## Common Pitfalls

### Pitfall 1: Reading OWN-09 as "delete a law" and finding nothing to delete (or deleting `loanLivenessFixpoint` itself)

**What goes wrong:** A plan that searches for "the intraprocedural
loan-liveness law" as a separate function will either come up empty (there
is only one fixpoint function) or, worse, delete
`loanLivenessFixpoint`/`materializeLoanEndpoints` themselves — which are
also `checkInterproceduralLoanLiveness`'s, `checkBranch`'s arm-block
liveness, and `corevalidate`'s (no — corevalidate has its own,
`recomputeLoanEndpoints`, materially different) shared machinery.

**Why it happens:** D-08-09/D-08-27's own text (now corrected by D-09-07) used
"law" language that reads naturally as "algorithm," and the stale line
reference (`check.go:2979` vs. actual `:3743`) in `08-CONTEXT.md` made
re-verification easy to skip.

**How to avoid:** State D-09-07's finding at the top of any OWN-09-touching
plan task: one algorithm (`loanLivenessFixpoint`), two callers
(`analyzeStraightLine`/`analyzeArmBody`'s summary-blind shadow calls at
`check.go:3287`/`:2074`, and `checkInterproceduralLoanLiveness`'s
summary-aware call at `check.go:835`). Only the first caller and its
`shadow:place:*` synthesis (`computeLoanLastUses`, `check.go:3743-3855`) are
deleted.

**Warning signs:** A task description that says "delete the intraprocedural
loan-liveness algorithm" rather than "delete `computeLoanLastUses` and its
two call sites, replacing their `move_while_borrowed`/`borrow_conflict`
emission with an extension of `checkInterproceduralLoanLiveness`."

### Pitfall 2: Assuming all five `ownership.*` codes depend on `computeLoanLastUses`'s timing

**What goes wrong:** A plan tries to move `use_after_move`,
`borrow_requires_share`, and `transfer_requires_take` emission to
post-assembly too, because they live in the same per-binding loop as
`move_while_borrowed`/`borrow_conflict`.

**Why it happens:** All five codes are raised from the same
`analyzeStraightLine`/`analyzeArmBody` per-binding loop, so they look
uniformly coupled to `computeLoanLastUses` at a glance.

**How to avoid:** Verified this session — only two codes read
`loanUses[index]`/`activeLoans[...].lastUse`-derived state:
`ownership.move_while_borrowed` (check.go:2176, :3392, gated by
`activeLoans[source.place.ID]` still being non-empty, which is seeded from
`expiringLoans[loan.lastUse]` populated at each borrow binding using
`loanUses[index]` from `computeLoanLastUses`) and `ownership.borrow_conflict`
(via `conflictingLoan(activeLoans[...])`, same `activeLoans` set). The other
three fire from immediate per-binding facts independent of any last-use
timing: `use_after_move` from `!source.initialized` (a move having already
happened, tracked directly in `placeState`, not from loan timing),
`borrow_requires_share`/`transfer_requires_take` from ability presence
(`hasTypeAbility`). Restructuring only needs to touch the two timing-
dependent codes; the other three legitimately stay as immediate
per-binding lowering-time facts (they are not "ownership decisions" in the
loan-liveness sense OWN-09 targets — they are place/ability facts, not
timing facts).

**Warning signs:** A task diff that touches `use_after_move`'s
`!source.initialized` check or any `hasTypeAbility` call — that is scope
creep past what OWN-09 requires.

### Pitfall 3: The corpus generator D-09-23 names is unexported and package-private

**What goes wrong:** A plan assumes `session_peer_gate_test.go` (package
`session_test`) can call `generateCallGraphCorpus` directly, discovers at
execution time that it cannot (unexported symbol in a `_test.go` file in a
different package — Go visibility rules make this uncallable regardless of
build-tag or test-package tricks), and either stalls mid-plan or silently
duplicates the generator with drift risk.

**Why it happens:** D-09-23's phrasing ("feed it additionally from Phase 08's
`generateCallGraphCorpus`") reads like an ordinary cross-package test helper
reuse, which is common elsewhere in this codebase (`testsupport.ProjectPath`),
masking that this specific helper is unexported and test-file-scoped.

**How to avoid:** Verified this session: `generateCallGraphCorpus` is at
`internal/compiler/check/costcorpus_test.go:50`, unexported, package `check`.
Resolve this explicitly at plan-decomposition time — either (a) export it
(rename to `GenerateCallGraphCorpus`) and move it to a non-`_test.go` file in
`check` (acceptable production-code surface addition, since it returns a
`core.Program` and takes primitive args — no test-only types leak), or (b)
relocate it to a shared `testsupport`-style package both `check`'s tests and
`session`'s tests can import. Do not duplicate it — that reintroduces the
exact two-derivations-one-truth risk this phase exists to eliminate, this
time for the *test fixtures* rather than the production code.

**Warning signs:** A new function in `session_peer_gate_test.go` whose body
looks structurally identical to `corpusChain`/`corpusLayered` in
`costcorpus_test.go`.

### Pitfall 4: Diagnostic identity changes silently when emission moves from per-function to post-assembly

**What goes wrong:** A fixture that used to fail with (say)
`ownership.move_while_borrowed` from function A (because per-function early
exit stopped at the first error in source order) now fails with a different
first diagnostic from function B, because the post-assembly pass iterates
`program.Functions` in a different order (or the same order, but a
DIFFERENT first offending operation is found within a function due to the
whole-program view). Since `Code+Span+Causes` folds into the diagnostic's
`ID` (verified, `diagnostic.go:96-112`: `Error()` marshals
`{Schema,Code,Span,Causes}` and SHA-256-hashes it), this changes a
published, agent-facing diagnostic ID for an unrelated reason (ordering),
not a content reason.

**Why it happens:** Moving from N independent per-function early-exit passes
to 1 whole-program pass changes the *set* of programs for which "the first
error found" is well-defined the same way across both. A program with
errors in two different functions can flip which one is reported first.

**How to avoid:** D-09-13 requires "an explicit ordering-stability
assertion, not a corpus replay." Concretely: iterate `program.Functions` in
declaration order (matching source order, which the AST->core lowering
already preserves in `program.Funcs`/`program.Functions`) so cross-function
ordering is unchanged from today's "which function got processed by
`checkLinear`/`checkBranch` first" order. Add a same-package test that feeds
a synthetic two-function program with a known error in each function and
asserts the reported diagnostic's `Code` matches source declaration order,
not iteration-map order (Go map iteration over `program.Functions` if it is
ever a map, rather than a slice, would be a second silent ordering hazard —
verify `core.Program.Functions`'s type is a slice, not a map, before
assuming order is stable).

**Warning signs:** Any golden-fixture diff where only the diagnostic `Code`
or `ID` changed but the underlying source program did not.

### Pitfall 5: Treating `peerDivergenceExpected`'s two retiring entries' paths as `testdata/phase07/...`

**What goes wrong:** A plan task looks for
`testdata/phase07/twin_a_accept.lang` and
`testdata/phase07/relay_depth2_accept.lang` (as CONTEXT.md's D-09-03 names
them) to retire from `peerDivergenceExpected`, does not find them at that
path, and either creates a duplicate fixture at the wrong path or silently
assumes the entries do not exist yet.

**Why it happens:** CONTEXT.md's D-09-03 states the paths as
`testdata/phase07/twin_a_accept.lang` /
`testdata/phase07/relay_depth2_accept.lang`. **Verified this session by
reading `session_peer_gate_test.go:23-59` directly: the actual shipped map
keys are `testdata/phase08/twin_a_accept.lang` and
`testdata/phase08/relay_depth2_accept.lang`** (both mapped to
`"core.move_while_borrowed"`). This is a third stale planning-document
reference, in addition to the two D-09-45 already names — not yet flagged
anywhere in CONTEXT.md.

**How to avoid:** Use the verified paths (`testdata/phase08/...`) when
writing the plan task that retires these two entries. The code wins over
the planning document per this research task's own instruction.

**Warning signs:** A `grep` for `testdata/phase07/twin_a_accept.lang`
returning nothing.

## Code Examples

### `checkInterproceduralLoanLiveness`'s existing whole-program shape (the pass OWN-09 extends)

```go
// Source: internal/compiler/check/check.go:828-941 (verified read this session, excerpted)
func checkInterproceduralLoanLiveness(program core.Program, summaries interproceduralSummaryTable, spanByOperationID map[string]diagnostic.Span) []diagnostic.Diagnostic {
	var diagnostics []diagnostic.Diagnostic
	for _, function := range program.Functions {
		if function.Linear == nil || len(function.Linear.Operations) == 0 {
			continue
		}
		blocks := cfgBlocksForFunction(function)
		fixpoint, diag := loanLivenessFixpoint(function.ID, blocks, summaries, function.Span)
		// ... builds endpoints, chain := derivePlaceLoans(...), finds the
		// earliest offending OpMove, and emits check.interprocedural_loan_liveness
		// ONLY for the case where a CALL extends the loan (forward or backward
		// direction). It currently has NO branch for a purely intraprocedural
		// conflict — "Neither direction fired ... nothing new to report" at
		// check.go:936-938, because the per-function admission pass (today's
		// analyzeStraightLine/analyzeArmBody) already caught that case earlier.
		// OWN-09's restructure must add that branch here, emitting
		// ownership.move_while_borrowed/ownership.borrow_conflict for the
		// non-call-extended case, once the earlier per-function emission is
		// deleted.
	}
	return diagnostics
}
```

### `buildLoanChainIndex`'s unconditional OpCall propagation — the D-09-03 defect site

```go
// Source: internal/compiler/corevalidate/corevalidate.go:1133-1149 (verified read this session, full function)
func buildLoanChainIndex(operations []core.LinearOperation, checks *int) *loanChainIndex {
	idx := &loanChainIndex{
		bornAt: make(map[string]string, len(operations)),
		parent: make(map[string]string, len(operations)),
		memo:   make(map[string][]string, len(operations)),
		checks: checks,
	}
	for _, operation := range operations {
		if operation.Kind == core.OpBorrowShared || operation.Kind == core.OpBorrowExclusive {
			idx.bornAt[operation.TargetID] = operation.LoanID
		}
		if operation.TargetID != "" {
			idx.parent[operation.TargetID] = operation.SourceID
			// ^ no check on operation.Kind here at all: an OpCall's
			//   TargetID always gets parent[TargetID] = SourceID, exactly
			//   as if the call were a transparent OpCopy of its argument.
			//   OWN-07's fix: gate this specifically for OpCall by
			//   consulting the callee's peer-derived
			//   "returns-borrow-of-param" bit (via v.peerSignatures,
			//   analogous to how check.go's derivePlaceLoans already
			//   gates its own OpCall branch on summaries.lookup).
		}
	}
	return idx
}
```

### `diagnostic.Error`'s identity fold — why D-09-13's ordering concern is real

```go
// Source: internal/compiler/diagnostic/diagnostic.go:96-114 (verified read this session, exact)
func Error(code string, span Span, message string, causes ...Cause) Diagnostic {
	identity := struct {
		Schema string
		Code   string
		Span   Span
		Causes []Cause
	}{Schema: Schema, Code: code, Span: span, Causes: causes}
	encoded, _ := json.Marshal(identity)
	sum := sha256.Sum256(encoded)
	return Diagnostic{Schema: Schema, ID: "diagnostic:" + hex.EncodeToString(sum[:12]), Code: code, Severity: "error", Primary: span, Message: message, Causes: causes}
}
```
Note `Message` is NOT part of the identity hash — only `Schema+Code+Span+
Causes`. This means the ordering risk D-09-13 names is specifically about
WHICH diagnostic is returned first when multiple would apply (today's
per-function early return vs. tomorrow's whole-program scan), not about the
hash formula itself changing.

## State of the Art

| Old Approach (pre-Phase-09) | Current/Target Approach (Phase 09) | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Loan-liveness decided at two points: lowering-time (`computeLoanLastUses`, summary-blind, AST-shadow ops) and post-assembly (`checkInterproceduralLoanLiveness`, summary-aware, real ops) | One decision point, post-assembly, summary-aware for both intra- and inter-procedural cases | This phase (D-09-08/09/10) | Pattern B's masking bug (D-08-41) is fixed as a side effect; diagnostic emission order becomes a first-class risk (D-09-13) |
| `corevalidate`'s `Callable` peer independently re-derives only `core.origin_omitted` | Peer re-derives all 4 `PublishProblemsFor` classes | This phase (D-09-15 through D-09-21) | `Callable` agreement stops being vacuous on 3 of 4 classes; closes D-07-33 |
| `corevalidate`'s `buildLoanChainIndex` propagates every `OpCall` transparently | Gated by a peer-derived "returns-borrow-of-param" bit before propagating | This phase (D-09-01/02/03) | Closes D-03-02's interprocedural half for the peer |
| `check.interprocedural_loan_liveness` scheduled for promotion to `core.*` "at the moment corevalidate independently re-derives the same fact" (D-08-21) | Promotion commitment formally superseded; codes stay divergent by design | This phase (D-09-31/32/33) | Reverses a Phase 08 written commitment; documented as a reversal, not silently dropped |

**Deprecated/outdated:**
- `computeLoanLastUses` and its `shadow:place:*`/`shadow:op:*`/`shadow:loan:*`
  synthetic-operation scaffolding: deleted this phase (D-09-08 reverses
  D-08-41's "Not scheduled" disposition).
- The `08-CONTEXT.md` line reference `check.go:2979` for
  `computeLoanLastUses`: stale, actual location is `check.go:3743` (D-09-45a,
  re-verified this session).

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | The OWN-09 restructure needs to change only `ownership.move_while_borrowed` and `ownership.borrow_conflict` emission timing, not `use_after_move`/`borrow_requires_share`/`transfer_requires_take` | Common Pitfalls #2 | If wrong, the restructure is far larger than scoped — a plan sized against this assumption would blow the 2× scope-cut trigger (D-09-43) faster than expected. Verified by direct code read this session (which fields each check consults), so risk is LOW, but the interaction with `placeState.initialized` tracking after the restructure (does deferring `move_while_borrowed` change WHEN a place becomes uninitialized, which `use_after_move` depends on?) was not fully traced end-to-end and should be a plan-time verification step |
| A2 | `generateCallGraphCorpus` should be exported/relocated rather than duplicated | Common Pitfalls #3 | If a plan instead duplicates the generator into `session`'s test package, two generators can drift, silently reproducing the two-derivations-one-truth risk for TEST FIXTURES (lower severity than for production code, but still undermines TRU-04's "same generator, disjoint consumers" argument) |
| A3 | `core.Program.Functions` is a slice (not a map), so post-assembly iteration order is deterministic by declaration/lowering order | Common Pitfalls #4 | If it were a map, Pitfall 4's ordering-stability risk would be far more severe (nondeterministic across runs, not just "different from before"). Not independently re-verified against `core/core.go`'s struct definition in this session — recommend the planner confirm this one-line fact before writing the ordering-stability test |
| A4 | The peer's new OWN-07 liveness bit should be derived non-iteratively (single forward pass, no fixpoint), matching `recomputeLoanEndpoints`'s already-shipped non-worklist shape | Architecture Patterns #2 | If the peer's new derivation instead needs a converging fixpoint (e.g., because a loan can extend across more than one `OpCall` hop and needs multi-pass propagation to reach a fixed point), a single forward pass could under-propagate. This should be checked against the specific multi-hop chain fixtures (e.g., `relay_depth2_accept.lang`) at plan time — a single pass over `Linear.Operations` in program order, consulting ALREADY-FINALIZED callee summaries via the existing callee-before-caller `v.peerPostorder`, should suffice by the same argument `buildInterproceduralSummaries` already relies on (D-08-42), but this was not proven by tracing `relay_depth2_accept.lang` step-by-step this session |

**If this table is empty:** N/A — see above.

## Open Questions

1. **Does deferring `move_while_borrowed`'s check to post-assembly change
   `use_after_move`'s behavior for any fixture?**
   - What we know: `use_after_move` fires on `!source.initialized`, set
     directly when an `OpMove`/`take` succeeds during the SAME per-binding
     loop `move_while_borrowed` currently gates. If `move_while_borrowed`
     currently PREVENTS a move (by refusing it), removing that inline gate
     means the move proceeds during lowering (marking the place moved) even
     when the post-assembly pass would later refuse the whole program for
     `move_while_borrowed`.
   - What's unclear: whether any existing fixture has both an interprocedural
     conflict AND a subsequent `use_after_move`-shaped access to the SAME
     place within the same function, where today's inline refusal short-
     circuits before that access is ever reached, but a deferred refusal
     would let lowering proceed far enough to ALSO trip `use_after_move` —
     producing two errors where one existed, and changing which error a
     whole-program pass reports first (Pitfall 4's exact concern, but for a
     within-function interaction rather than cross-function).
   - Recommendation: add this as an explicit plan task — enumerate every
     existing fixture where `move_while_borrowed` currently fires, and for
     each, trace what the REST of that function's bindings would do if
     lowering proceeded past the move (rather than short-circuiting). This
     is bounded work given the ~58-program, ~28-line-average corpus
     (D-09-11's own characterization of corpus size).

2. **Does the peer's new liveness bit need multi-hop (>1 `OpCall`) propagation, and if so, does a single forward pass suffice?**
   - What we know: `relay_depth2_accept.lang` exists specifically to exercise
     a 2-hop relay chain (per its name and its presence in
     `peerDivergenceExpected` today). `check`'s own summary machinery
     (`buildInterproceduralSummaries`) already handles arbitrary-depth chains
     via callee-before-caller ordering over the WHOLE program (not per-hop).
   - What's unclear: whether the peer's `buildLoanChainIndex` consultation of
     a new liveness bit, computed function-by-function, correctly composes
     across 2+ hops without its own fixpoint — i.e., whether "already-
     finalized callee summary, consulted once, in postorder" is sufficient
     for the CHAIN-PROPAGATION problem `buildLoanChainIndex` solves (which is
     itself a chain-following structure, `carriedLoans`/`foldChain`), or
     whether the loan-liveness bit itself (as opposed to the "does this
     function return a borrow of its param" signature bit) needs a genuine
     transitive closure over the call graph.
   - Recommendation: trace `relay_depth2_accept.lang` through the proposed
     peer derivation by hand (or write the differential test first, per
     D-09-27's red-first discipline) before finalizing the peer's exact
     algorithm shape. This is precisely what D-09-27's red-first ordering is
     for — the test written against today's known divergence should exercise
     this exact case.

## Environment Availability

Skipped — this phase has no external tool/service/runtime dependencies beyond
the existing Go toolchain, already available and exercised by every prior
phase (`go test ./internal/compiler/check ./internal/compiler/corevalidate
./internal/compiler/pathoracle` was run and passed during D-09-40a's
inventory this session's predecessor context, per CONTEXT.md).

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go's standard `testing` package (`go test`) |
| Config file | none — no `go.mod` test-runner config beyond the module itself |
| Quick run command | `go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/pathoracle ./internal/compiler/session ./internal/compiler/originvalidate` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| OWN-07 | `corevalidate` re-derives interprocedural loan liveness without sharing an implementation with `check`; seeded fault diverges | unit + differential | `go test ./internal/compiler/corevalidate -run TestBuildLoanChainIndex` (new) | ❌ Wave 0 — new test |
| OWN-07 | Import independence extended to new liveness-bit code | unit | `go test ./internal/compiler/corevalidate -run TestValidatorImportsStayIndependent` | ✅ existing (`corevalidate_endpoint_internal_test.go:107` and 3 siblings) — extend coverage, not the test itself |
| OWN-08 | Peer re-derives all 4 `PublishProblemsFor` classes | unit | `go test ./internal/compiler/corevalidate -run TestPeerDoesNotRederiveNarrowedClasses` (existing, must be updated to assert re-derivation, not narrowing) | ✅ exists, needs semantic flip |
| OWN-09 | `computeLoanLastUses` deleted; single decision point | unit + regression | `go test ./internal/compiler/check -run TestOwnershipSequenceExhaustive` and `TestBranchSequenceExhaustive` (existing exhaustive enumeration, re-run post-restructure) | ✅ exists |
| OWN-09 | Diagnostic ordering stability | unit | new: `TestInterproceduralOrderingStability` (name at planner's discretion) | ❌ Wave 0 — new test (Pitfall 4) |
| TRU-04 | Zero divergence over recursion/diamond/deep-chain shapes, before either peer ships | integration/differential | `go test ./internal/compiler/session -run TestNoUndeclaredCheckPeerDivergenceAcrossCorpus` (existing, extended per D-09-23) | ✅ exists, needs corpus-source extension (see Open Question re: `generateCallGraphCorpus` visibility) |
| TRU-04 | Cycle-peer differential (D-09-22's disposition) | unit | `go test ./internal/compiler/corevalidate -run TestCyclePeer` (existing `corevalidate_cycle_peer_test.go`, extend) | ✅ exists |
| TRU-04 | Seeded endpoint-level fault + companion assertion (both directions) | unit | existing precedent `TestVerifyCallableRefusalSeamCheckDisabledCorevalidateStillRefuses` + new `corevalidate`-side companion (D-09-25) | ⚠️ one direction exists, one new — Wave 0 |
| TRU-04 | Counted-work linear-or-bounded cost lane | unit/benchmark | `go test ./internal/compiler/session -run TestGateEligibleMetricSetsAgreeAcrossChokepoints` (existing, extend to 3-way agreement) | ✅ exists, needs a 3rd chokepoint |
| QLT-07 | Loan-liveness Nyquist subset closed | documentation | none — `09-VALIDATION.md`'s new section, per D-09-42 | N/A — doc-only |

### Sampling Rate

- **Per task commit:** `go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/session`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd-verify-work`; additionally, the
  mid-phase gate structure Phase 08 used (a mandatory checkpoint before the
  `computeLoanLastUses` deletion lands, per D-09-10) should re-run the
  differential specifically.

### Wave 0 Gaps

- [ ] A new `corevalidate`-side companion seam + test proving `check`'s
      liveness refusal disabled still leaves `corevalidate` refusing (D-09-25,
      the direction NOT already covered by `TestVerifyCallableRefusalSeamCheckDisabledCorevalidateStillRefuses`)
- [ ] `TestInterproceduralOrderingStability` (or equivalent name) — a
      same-package `check` test asserting cross-function diagnostic-selection
      order is unchanged pre/post restructure (D-09-13)
- [ ] Resolution of `generateCallGraphCorpus`'s package-visibility gap before
      `session_peer_gate_test.go` can consume it (export it from a non-`_test.go`
      file in `check`, or relocate to a shared test-support location)
- [ ] Synthetic programs reaching every `ownership.*` code specifically through
      the AST-shadow path (D-09-11) — needed BEFORE the pre-deletion shadow-run
      widening, not after

## M001 Phase 3 Debt Closure (loan-liveness subset)

Per D-09-42, this closes in a dedicated `09-VALIDATION.md` section (not this
document), citing the `03-VALIDATION.md` rows verified already-satisfied per
D-09-40a's inventory:

- **Per-Task Verification Map:** 03-03-01, 03-03-02, 03-03-03, 03-04-01,
  03-04-02, 03-04-03, 03-05-01, 03-05-02, 03-05-03
- **Mutation-Kill Register:** path-oracle endpoint differential, uniform-join
  falsifier, validator endpoint recomputation, edge-specificity fixture pair,
  counted-work honesty
- **Generator Reachability Register:** `generatedLinearProgram` (extended),
  exhaustive ownership sequence enumeration, path-oracle metamorphic trials

All 33 named tests were verified present and passing (`go test
./internal/compiler/check ./internal/compiler/corevalidate
./internal/compiler/pathoracle` — all `ok`) at context-capture time
(D-09-40a). **This is a ratification/documentation task, not a
test-authorship task** — the planner should size it as a single `09-PLAN.md`
task writing the `09-VALIDATION.md` section plus one non-mutating pointer
line appended to `03-VALIDATION.md` (D-09-41/D-09-42), not as new test
development. **`03-VALIDATION.md`'s `nyquist_compliant: false` and its
OWN-04/loop-carried-liveness scope limitation are explicitly OUT of this
closure and must never be edited** — only the pointer line is added.

## Security Domain

`security_enforcement` is enabled (`.planning/config.json`:
`workflow.security_enforcement: true`, `security_asvs_level: 1`).

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | Not applicable — this phase is a compiler-internal admission-layer change with no auth surface |
| V3 Session Management | No | Not applicable |
| V4 Access Control | No | Not applicable |
| V5 Input Validation | Yes (narrow sense) | `corevalidate` is explicitly a **source-blind, adversarial-input-tolerant validator** — `carriedLoans`'s iterative (not recursive) walk with cycle detection (`corevalidate.go:1160-1181`, verified read) is exactly this: a defense against a corrupted/adversarial `core.Program` artifact causing unbounded recursion or a crash. Any new peer code must preserve this "never crash on a corrupted core artifact" property |
| V6 Cryptography | No | Not applicable — `diagnostic.Error`'s SHA-256 use is for identity/dedup, not a security boundary |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Adversarial/corrupted `core.Program` causing infinite loop or crash in a peer validator | Denial of Service | Iterative (not recursive) graph walks with visited-set cycle detection, exactly as `loanChainIndex.carriedLoans` and `blockReach` already do; any new OWN-07 traversal must follow the same shape |
| Two peers silently sharing an implementation, making "independent agreement" vacuous | Tampering (of the evidence, not the data) | Import-guard tests (4 shipped) + derivation-method independence (forward-propagation peer vs. backward-memoized producer), the entire subject of this phase |
| A seeded fault in shared code causing correlated (not independent) divergence, producing a false sense of security from criterion 2 | Repudiation of the "independence" claim itself | Companion assertion in BOTH directions (D-09-25), following Knight & Leveson's correlated-failure finding — this is the phase's own named mitigation for its own named risk |

## Sources

### Primary (HIGH confidence — all read directly from the shipped tree this session)

- `internal/compiler/corevalidate/corevalidate.go` (full read of lines 1-160,
  440-560, 1126-1420, 1942-2375) — peer substrate, `buildLoanChainIndex`,
  `recomputeLoanEndpoints`, `derivePeerSignature`,
  `peerParameterEscapesOwned`, `peerReturnDerivesFromBorrow`, `peerCallable`
- `internal/compiler/check/check.go` (full read of lines 1-120, 679-1010,
  1170-1260, 2060-2330, 3723-3860, plus targeted greps of the whole file) —
  `Program`, `deriveFunctionUsesParam`, `verifyCallableRefusal`,
  `analyzeArmBody`, `computeLoanLastUses`, `checkInterproceduralLoanLiveness`,
  `interproceduralLoanLivenessDiagnostic`
- `internal/compiler/check/check_test.go` (lines 3314-3357) —
  `TestVerifyCallableRefusalSeamCheckDisabledCorevalidateStillRefuses`,
  `TestInterproceduralDisclosedFieldSet` (lines 4920-4950)
- `internal/compiler/diagnostic/diagnostic.go` (lines 85-115) — `Error()`'s
  identity-hash construction
- `internal/compiler/originvalidate/originvalidate.go` (lines 280-346) —
  `checkForeignOriginOmitted`
- `internal/compiler/core/core.go` (targeted greps) — `ParameterContract`,
  `LinearOperation`, `DecodeInterface`'s closed-set Mode validation
  (lines 458-472)
- `internal/compiler/corevalidate/corevalidate_endpoint_internal_test.go`
  (lines 95-114) — the four-package import guard
- `internal/compiler/session/session_peer_gate_test.go` (lines 20-60,
  215-221) — `peerDivergenceExpected`'s actual entries (paths corrected vs.
  CONTEXT.md, see Pitfall 5)
- `internal/compiler/session/session_phase6_budget.go` (lines 40-77) —
  `QLT02MetricVocabulary`, `QLT02GateEligibleMetrics`
- `internal/compiler/session/qlt02_budget_manifest.json` (full read) —
  4 existing rows, confirming `recomputed_work_growth_exponent` already
  ratified at 1200 milliexponent
- `internal/compiler/session/risk_lanes.json` (lines 230-235) —
  `lane:interprocedural-cost-scaling`
- `internal/compiler/measure/statistics.go` (lines 130-138) —
  `GateEligibleMetrics`
- `internal/compiler/check/costcorpus_test.go` (lines 1, 50-64) —
  `generateCallGraphCorpus`'s package/export status
- `.planning/phases/08-interprocedural-loan-liveness-in-check/PHASE-08-DEBT.md`
  (lines 33-36, 236-253) — D-08-41's full text (Pattern B scope limit),
  D-08-42/D-08-43
- `.planning/phases/07-calls-signatures-and-call-graph-refusal/PHASE-07-DEBT.md`
  (lines 18-52, 72+) — D-03-02, D-07-33, D-07-46's full text
- `.planning/phases/08-interprocedural-loan-liveness-in-check/08-VALIDATION.md`
  (section headers) — the shape a `09-VALIDATION.md` should follow
- `.planning/config.json` — `workflow.nyquist_validation: true`,
  `workflow.security_enforcement: true`, `security_asvs_level: 1`

### Secondary (MEDIUM confidence)

- `.planning/phases/09-peer-re-derivation-and-d-03-02-closure/09-CONTEXT.md`
  — the primary input, itself independently re-verified against the tree by
  the orchestrator before this research session per its own text; treated as
  authoritative except where this session's direct reads found a discrepancy
  (Pitfall 5's stale `testdata/phase07/` vs. actual `testdata/phase08/` paths)

### Tertiary (LOW confidence)

- None — this phase required no web research; every claim traces to either
  the shipped tree (read directly) or CONTEXT.md (itself tree-verified).

## Metadata

**Confidence breakdown:**
- Standard stack: N/A — no external dependencies
- Architecture: HIGH — every cited function/line was read directly this
  session, not inferred from CONTEXT.md's summaries
- Pitfalls: HIGH for Pitfalls 1/2/3/5 (directly verified against code);
  MEDIUM for Pitfall 4 (the diagnostic-identity mechanism is verified, but
  the specific ordering behavior post-restructure is a design question for
  the plan, not yet resolved by this research)

**Research date:** 2026-09-10
**Valid until:** Until Phase 09 lands (this is a point-in-time snapshot of an
actively-changing tree; re-verify line numbers if planning is delayed and
other phases' work merges first)
