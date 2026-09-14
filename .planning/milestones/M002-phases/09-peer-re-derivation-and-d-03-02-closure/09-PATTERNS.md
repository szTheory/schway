# Phase 09: Peer Re-Derivation and D-03-02 Closure - Pattern Map

**Mapped:** 2026-09-10
**Files analyzed:** 10 work items (all extensions to existing tracked files; no
greenfield production file)
**Analogs found:** 10 / 10 (all in-tree, all git-tracked, verified via
`git ls-files`)

Like Phase 08, this phase is dominated by additive clauses inside files that
already contain a structurally near-identical instance of the pattern one
function away. The one genuine novelty is that two of the ten items
(1 and 3) require *contrasting*-shape analogs to actively avoid — `check`'s
backward memoized walk and `originvalidate`'s backward recomputation — since
copying either would collapse peer independence. Those contrasts are called
out explicitly below, not just the "copy this" analogs.

## File Classification

| New/Modified File (work item) | Role | Data Flow | Closest Analog (COPY) | Contrasting Analog (DO NOT COPY) | Match Quality |
|---|---|---|---|---|---|
| 1. `corevalidate.go` (or sibling) — peer interprocedural liveness derivation | service (peer dataflow pass) | transform (forward set-propagation, memoized per function) | `peerReturnDerivesFromBorrow` (`corevalidate.go:2099`), `peerParameterEscapesOwned` (`:2061`), postorder refinement loop `chainPeerClosureDigests` (`:495-549`, the loop CONTEXT.md cites as `:505-547`) | `check.go`'s `buildInterproceduralSummaries` (`:579`) walked backward over `callgraph.Order`, `deriveFunctionUsesParam` (`:686`) | exact (copy), exact (contrast) |
| 2. `buildLoanChainIndex` `OpCall` consult (`corevalidate.go:1133`) | transform (canonicalization gate) | transform | `loanChainIndex`/`carriedLoans`/`foldChain` (`:1126-1202`) — same file, same struct, extend in place | n/a | exact |
| 3. Access-mode payload + containment check on peer's forward origin walk | transform + controller (derivation + comparison) | transform | `peerReturnDerivesFromBorrow` (`:2099-2131`), `peerCallable` (`:2349`) | `originvalidate.go`'s `RecomputeOrigin` (`:236`), `RecomputeOriginPerReturn` (`:135`), `checkForeignOriginOmitted` (`:290-345`) — narrow-technique precedent only, never code | exact (copy), exact (contrast) |
| 4. Delete `computeLoanLastUses` (`check.go:3743`); restructure `check` to lower-then-decide | transform → deletion + control-flow restructure | transform (two-pass: lower, then decide) | `checkInterproceduralLoanLiveness` (`check.go:828`) — the shape to move toward | the two call sites being retired: `analyzeArmBody` (`:2069`, uses index at `:2074`), `analyzeStraightLine` (`:3280`, uses index at `:3287`), with their `activeLoans`/`expiringLoans` state machines at `:2176`/`:3392` | exact |
| 5. Seeded-fault seam + companion assertion (both directions) | config/utility (fault-injection seam) + test | event-driven (test-only flip) | `verifyCallableRefusalSeam` (`check.go:1183`, read at `:1250`) + `TestVerifyCallableRefusalSeamCheckDisabledCorevalidateStillRefuses` (`check_test.go:3322`) | n/a — this is the shape to replicate in the reverse direction (`corevalidate`-side seam, `check` asserts still-refuses) | exact |
| 6. Cycle-peer differential (self/mutual/indirect) | test | event-driven (synthetic `core.Program` construction) | `corevalidate_cycle_peer_test.go` — `syntheticProgram`/`syntheticFunction` (`:35`, `:79`), `TestCyclePeerRefusesSyntheticMutualCycle` (`:156`), `TestCyclePeerRefusesSelfEdge` (`:174`) | n/a | exact |
| 7. Corpus-wide divergence gate widened with Phase 08 corpus | test | batch (corpus replay) | `peerDivergenceExpected` (`session_peer_gate_test.go:23`) + `TestNoUndeclaredCheckPeerDivergenceAcrossCorpus` (`:218`) — extend consumer, not a new harness | n/a | exact |
| 8. Third gate-eligible metric chokepoint widening (peer's own bound) | config (gate-eligibility declaration) + data | transform/batch | `measure.Demote`/`gateEligibleMetricSet()` (`statistics.go:160-164`, `:134-137`... see below), `session.QLT02GateEligibleMetrics()`/`QLT02MetricVocabulary()` (`session_phase6_budget.go:46-66`), `TestGateEligibleMetricSetsAgreeAcrossChokepoints` (`session_phase6_budget_test.go:237`) — Phase 08 already did exactly this widening once (commit `b28a92f`) | n/a | exact |
| 9. Import-independence guard extension | test | request-response (static source-text scan) | Four existing guards: `corevalidate_cycle_peer_test.go:146` (`TestCyclePeerTestFileImportsStayIndependent`), `corevalidate_endpoint_internal_test.go:107` (package-internal, scans imports of files under this test dir), `corevalidate_exclusive_test.go:216` (`TestAttributeValidatorImportsStayIndependent`, string-scans `corevalidate.go` source text), `corevalidate_test.go:148` (broadest string-scan, six forbidden substrings) | n/a | exact |
| 10. `09-VALIDATION.md` + `PHASE-09-DEBT.md` | test-fixture/doc (data) | batch | `08-VALIDATION.md` (frontmatter + Per-Task Verification Map shape), `PHASE-08-DEBT.md` (frontmatter + `## Items` table + `### <ID>` detail sections), `TestDebtRegistersAreWellFormed` (`session_test.go:2613`) | n/a | exact |

## Pattern Assignments

### 1. `corevalidate`'s peer interprocedural liveness derivation

**Analog to copy — the postorder refinement loop the new bits fold into**
(`corevalidate.go:495-549`, verbatim excerpt, this is the loop CONTEXT.md's
D-09-01 cites as `:505-547`):
```go
func (v *validator) chainPeerClosureDigests() {
	order := v.peerPostorder
	if closureDigestDiscoveryOrderForTest {
		// ...seam...
	}
	for _, id := range order {
		signature, ok := v.peerSignatures[id]
		if !ok {
			continue
		}
		var pairs []peerCalleeDigestPair
		if !closureDigestEmptyCalleesForTest {
			for _, calleeID := range v.peerAdjacency[id] {
				calleeSignature := v.peerSignatures[calleeID]
				pairs = append(pairs, peerCalleeDigestPair{ID: calleeID, ClosureDigest: calleeSignature.ClosureDigest})
				// ... join Foreign/Fails ...
			}
		}
		digest, err := peerComputeClosureDigest(signature, pairs)
		if err != nil {
			continue // fail closed: unreachable in practice
		}
		signature.ClosureDigest = digest
		v.peerSignatures[id] = signature
	}
}
```
This is a **callee-before-caller** memoized refinement loop, reading
`v.peerSignatures[calleeID]` (already-final by construction, per
`v.peerPostorder`'s own doc comment at `:130-138`) before writing the caller's
entry. **Fold the new liveness bits into this exact loop shape** — read each
callee's already-derived liveness fact from `v.peerSignatures[calleeID]`
before deriving the caller's own, exactly as `ClosureDigest`/`Foreign`/`Fails`
are joined today.

**Analog to copy — the forward set-propagation idiom itself**
(`corevalidate.go:2099-2131`, verbatim, `peerReturnDerivesFromBorrow`):
```go
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
And its sibling `peerParameterEscapesOwned` (`:2061-2080`) — same shape, one
forward pass over `function.Linear.Operations` building a `map[string]bool`
of derived places, a second short pass checking `OpReturn`. **The new
liveness derivation should be a third function of this exact shape**: single
forward pass(es) over `function.Linear.Operations`, a plain `map[string]bool`
(or two, if tracking both "uses param" and "returns borrow of param" as
distinct bits) as the propagated set, no recursion, no worklist, no
`callgraph` import.

**Contrasting shape — copy the ROLE, never the MECHANISM**
(`check.go:579-619`, `buildInterproceduralSummaries`, verbatim excerpt):
```go
func buildInterproceduralSummaries(program core.Program, table callSignatureTable) (interproceduralSummaryTable, int) {
	summaries := make(map[string]interproceduralSummary, len(program.Functions))
	result := interproceduralSummaryTable{summaries: summaries}
	order, err := callgraph.Order(program) // FORBIDDEN import for corevalidate
	if err != nil {
		return result, 0
	}
	// ...
	for i := len(order) - 1; i >= 0; i-- { // walked BACKWARD to get callee-before-caller
		functionID := order[i]
		// ...
		if function, ok := functionByID[functionID]; ok {
			usesParam, usesWork := deriveFunctionUsesParam(function, result) // memoized lookup of already-derived callees
			summary.usesParam = usesParam
			work += usesWork
		}
		summaries[functionID] = summary
	}
	return result, work
}
```
and `deriveFunctionUsesParam` (`check.go:686-728`) itself iterates
`function.Linear.Operations` to a **fixpoint within one function** (a small
inner worklist: `for { changed := false; ...; if !changed { break } }`),
consulting `summaries.lookup(operation.CalleeID)` for a callee already
finalized by the outer backward-order walk. The peer's derivation must
**never** import `callgraph`, **never** consume `callgraph.Order`'s
caller-before-callee list, and must be a **single deterministic forward pass**
(mirroring `peerReturnDerivesFromBorrow`'s shape) rather than an inner
fixpoint loop over one function's own operations — the two are allowed to
look similar in shape (both are set-propagation over operations) but must
differ in **direction and callee-ordering source**: `check` derives
callee-before-caller from `callgraph.Order` walked backward; the peer derives
callee-before-caller from its own `v.peerPostorder` (`:472`, its own DFS
byproduct).

**Placement (D-09-04):** either inline in `corevalidate.go` beside
`peerReturnDerivesFromBorrow`, or a sibling file `corevalidate_peer_liveness.go`
in the same package — planner's discretion. Either way the new function's doc
comment must state explicitly (mirroring `peerReturnDerivesFromBorrow`'s own
doc comment style at `:2082-2098`, which names both "independent
re-derivation" and "materially different mechanism") that what is reused from
the postorder loop is **substrate**, and the derivation itself is new.

---

### 2. `buildLoanChainIndex` `OpCall` consult

**Analog — the exact struct and functions to extend, same file**
(`corevalidate.go:1126-1202`, verbatim, already fully quoted above under
required reading). Today:
```go
for _, operation := range operations {
	if operation.Kind == core.OpBorrowShared || operation.Kind == core.OpBorrowExclusive {
		idx.bornAt[operation.TargetID] = operation.LoanID
	}
	if operation.TargetID != "" {
		idx.parent[operation.TargetID] = operation.SourceID // unconditional, even across OpCall
	}
}
```
**Extension:** before the unconditional `idx.parent[operation.TargetID] =
operation.SourceID` line, add an `operation.Kind == core.OpCall` branch that
consults the peer's own liveness bit for `operation.CalleeID` (the map/table
built by work item 1) — if the callee is not derived to return a borrow of
its parameter, do **not** set `idx.parent[operation.TargetID]`, breaking the
chain at the call boundary (a fresh, owned identity), matching `check`'s
`derivePlaceLoans` `OpCall` branch's own D-08-07 fix in spirit (do not import
or copy that function — this is `corevalidate`'s own struct and its own
`buildLoanChainIndex` caller, `recomputeLoanEndpoints` at `:1295`). `carriedLoans`
(`:1160-1182`) and `foldChain` (`:1188-1202`) need no change — they walk
`idx.parent`/`idx.bornAt` however they were built.

**Fixture retirement this change must produce:** `session_peer_gate_test.go`'s
`peerDivergenceExpected` (`:56-58`) currently pins:
```go
"testdata/phase08/twin_a_accept.lang":       "core.move_while_borrowed",
"testdata/phase08/relay_depth2_accept.lang": "core.move_while_borrowed",
```
Both entries must be **deleted** once the `OpCall` consult lands — D-09-03
names this a required assertion, not a side effect.

---

### 3. Access-mode payload + containment check on the peer's forward origin walk

**Analog to extend — same function, same file** (`peerReturnDerivesFromBorrow`,
`corevalidate.go:2099-2131`, already quoted above). Today `derived` is
`map[string]bool` (presence only). **Extension (D-09-17/D-09-18):** widen the
value to carry access mode (`shared`/`exclusive`), following the `switch
operation.Kind` structure already present — `OpBorrowShared` sets mode
`"shared"`, `OpBorrowExclusive` sets `"exclusive"`, and an `OpMove`/`OpCopy`
hop forwards the source's already-derived mode to the target (mirroring how
`paramTrace`/`derived` already forward membership). First-hop-wins direction
(D-09-19) means: once a place's derived mode is set, a later hop must never
overwrite it — same "if `derived[TargetID]` not already set" guard idiom
`peerReturnDerivesFromBorrow` and `peerParameterEscapesOwned` already use for
their boolean maps.

**Consumer — `peerCallable`** (`corevalidate.go:2349-2357`, verbatim):
```go
func peerCallable(function *core.Function) bool {
	if forcePeerCallableAlwaysTrue {
		return true
	}
	if function.PublicOrigin == nil {
		return !peerReturnDerivesFromBorrow(function)
	}
	return true // vacuous agreement for the other 3 classes today — this is D-09-16's target
}
```
**Extension:** when `function.PublicOrigin != nil`, instead of `return true`
unconditionally, compare the declared `PublicOrigin.Access` (and, once arity
widens past 1, `.Paths`) against the peer's own newly-access-mode-aware
forward walk result — a **containment check** (declared access is at least as
restrictive as / matches derived access), not a call into
`RecomputeOrigin`/`RecomputeOriginPerReturn`.

**Contrasting shape — technique-only precedent, never code, never import**
(`originvalidate.go:135-266`, `RecomputeOriginPerReturn`/`RecomputeOrigin`,
verbatim excerpt of the combination law):
```go
func RecomputeOrigin(function core.Function) (paths []string, access string, ok bool) {
	perReturn := RecomputeOriginPerReturn(function)
	// ...combine every per-return derived access; conflict -> AccessConflicting...
}
```
This is a **backward** per-return walk (`walkReturnOrigin`, `:171-218`,
walking `current = operation.SourceID` back toward `function.Parameter.ID`
via a `sourceOf` map built once) with a **combination law across multiple
returns**. The peer must not call this, must not replicate its exact
combination law unverified, and must prove first-hop-wins equivalence by
seeded-fault test rather than assumption (D-09-19).

**Named precedent for the bounded foreign-crossing decision**
(`originvalidate.go:282-345`, `checkForeignOriginOmitted`, verbatim excerpt):
```go
func checkForeignOriginOmitted(function core.Function) *Problem {
	if function.ForeignContract == nil {
		return nil
	}
	alias := function.ForeignContract.Alias
	if alias != "borrow" && alias != "retain" {
		return nil
	}
	// ...backward walk from every OpReturn, scoped to this ONE function's own
	// ForeignContract.Alias, checking whether current == foreignCallTarget...
}
```
This is the ~55-line proof-by-existence D-09-20 cites: a narrow, function-local
foreign-crossing check is possible without spike 005. The peer's forward walk
should mirror this **narrowness** (bounded to one function's own declared
contract), never cross `OpForeignCall` in its general propagation, and never
share code with this function.

---

### 4. Delete `computeLoanLastUses`; restructure `check` to lower-then-decide

**Deletion target** (`check.go:3743`, confirmed correct line per D-09-45's
correction of the stale `08-CONTEXT.md` reference):
```go
func computeLoanLastUses(parameterName string, body *ast.LinearBody) (map[int]loanUse, int) {
	const parameterPlaceID = "shadow:place:parameter"
	// ... shadow:place:*/shadow:loan:* synthesis over ast.LinearBody, before
	// any core.Function exists ...
}
```

**Its two call sites, to be retired** (both confirmed by `grep`):
- `analyzeArmBody` (`check.go:2069`), consuming the index at `:2074`:
  `loanUses, discoveryWork := computeLoanLastUses(parameterName, body)`
- `analyzeStraightLine` (`check.go:3280`), consuming the index at `:3287`:
  `loanUses, fixpointWork := computeLoanLastUses(parameterName, body)`

**Their state machines that consume the deleted index, and raise
`ownership.move_while_borrowed` during lowering** (verbatim, both nearly
identical, `:2160-2179` inside `analyzeArmBody` and `:3376-3399` inside
`analyzeStraightLine`):
```go
case "take":
	if loans := activeLoans[source.place.ID]; len(loans) > 0 {
		// ... pick lowest-sorted loan ID as blocking ...
		return fail(diagnostic.ErrorWithRepairs(
			"ownership.move_while_borrowed", binding.RHS.Span, "cannot transfer ownership while a future-used shared loan is live", causes,
			diagnostic.Repair{Kind: "move_after_last_borrow_use"},
		))
	}
	kind = core.OpMove
	// ...
```
Per D-09-09, this decision (and the `expiringLoans`/`activeLoans` bookkeeping
around it) moves out of lowering entirely; lowering emits `core.OpMove`
unconditionally and the single post-assembly pass decides.

**Shape to move toward — already exists and is already interprocedural**
(`check.go:828-902+`, `checkInterproceduralLoanLiveness`, verbatim excerpt):
```go
func checkInterproceduralLoanLiveness(program core.Program, summaries interproceduralSummaryTable, spanByOperationID map[string]diagnostic.Span) []diagnostic.Diagnostic {
	var diagnostics []diagnostic.Diagnostic
	for _, function := range program.Functions {
		if function.Linear == nil || len(function.Linear.Operations) == 0 {
			continue
		}
		blocks := cfgBlocksForFunction(function)
		fixpoint, diag := loanLivenessFixpoint(function.ID, blocks, summaries, function.Span)
		// ...
		endpoints := materializeLoanEndpoints(function.ID, blocks, edgeID, fixpoint, summaries)
		// ... walk sorted borrowedLoanIDs, find earliest offending OpMove ...
	}
	return diagnostics
}
```
This already runs **post-assembly, over the whole `core.Program`**, fed real
`interproceduralSummaryTable` — the exact "one pass over the assembled
`core.Program` decides" shape D-09-09 requires. The restructure's job is to
make this the **only** decision point by deleting the two early callers
above, not to build a new pass.

---

### 5. Seeded-fault seam + companion assertion (both directions)

**Exact precedent to replicate — declaration** (`check.go:1183`):
```go
var verifyCallableRefusalSeam = false
```
**Read site** (`check.go:1250`, inside `verifyCallableRefusal`):
```go
if verifyCallableRefusalSeam {
	callable = true
}
```
**Exact precedent to replicate — the companion-assertion test**
(`check_test.go:3322-3343`, verbatim, this is THE shape D-09-24/D-09-25
require, already running in the direction this phase's precedent needs):
```go
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
**Land in both directions (D-09-24):** (a) this exact `check`-disabled /
`corevalidate`-still-refuses test already generalizes to the new liveness
fact — write its liveness-scoped sibling the same way; (b) the sharper,
*new* direction is a `corevalidate`-side seam (unexported package-level bool
in `corevalidate.go`, e.g. inside the new liveness derivation or
`buildLoanChainIndex`'s `OpCall` consult) with a test disabling it and
asserting `check`'s `checkInterproceduralLoanLiveness` still refuses via
`check.interprocedural_loan_liveness` — mirroring this test's shape exactly
but swapped. State D-09-25's Knight & Leveson discrimination argument in that
new test's own doc comment, following this test's own doc-comment style
(lines above the func, naming the discrimination this test provides).

---

### 6. Cycle-peer differential (self/mutual/indirect)

**Analog — the synthetic `core.Program` builders and existing cycle tests**
(`corevalidate_cycle_peer_test.go`, function names confirmed by grep):
```go
func syntheticProgram(edges map[string][]string) core.Program { /* :35 */ }
func syntheticFunctionID(name string) string { return "s1:cyclepeer:fn:" + name } // :67
func syntheticFunction(name string, callees []string) core.Function { /* :79 */ }
func hasProblem(problems []Problem, code string) bool { /* :122 */ }

func TestCyclePeerTestFileImportsStayIndependent(t *testing.T) { /* :137 */ }
func TestCyclePeerRefusesSyntheticMutualCycle(t *testing.T) { /* :156 */ }
func TestCyclePeerRefusesSelfEdge(t *testing.T) { /* :174 */ }
func TestCyclePeerRefusesCycleWithNoEntryPoint(t *testing.T) { /* :190 */ }
func TestCyclePeerAcceptsAcyclicDiamond(t *testing.T) { /* :210 */ }
```
These already build exactly the self-cycle, mutual-cycle, and (via
`syntheticProgram`'s general `edges map[string][]string`) indirect/3+-hop-cycle
shapes D-09-22's disposition requires. **Extend this file** with a
`check`-side companion (both `check`'s callgraph refusal and `corevalidate`'s
independent refusal must agree on refusal-or-not AND on the witness) rather
than inventing a new synthetic-program builder — `syntheticProgram`/
`syntheticFunction` are the reusable construction, `hasProblem` the reusable
assertion helper.

---

### 7. Corpus-wide divergence gate widened with Phase 08's corpus

**Analog — the exact map and test to extend** (`session_peer_gate_test.go:15-23`
doc comment + declaration, `:218` test signature, `:256-268` the two-directional
exactness check):
```go
// peerDivergenceExpected is 07-10's named, commented, hand-maintained ...
var peerDivergenceExpected = map[string]string{
	// ...
	"testdata/phase08/twin_a_accept.lang":       "core.move_while_borrowed",
	"testdata/phase08/relay_depth2_accept.lang": "core.move_while_borrowed",
}

func TestNoUndeclaredCheckPeerDivergenceAcrossCorpus(t *testing.T) {
	// ...
	for path, wantCode := range peerDivergenceExpected {
		// fails if a declared entry no longer diverges
	}
	// ...
	if _, declared := peerDivergenceExpected[path]; !declared {
		t.Errorf("UNDECLARED divergence: %s ... is not in peerDivergenceExpected", path, gotCode)
	}
}
```
**Extension (D-09-23):** feed this **same** test additionally from Phase 08's
`generateCallGraphCorpus` (chain/diamond/dense/parser-shaped/forward shapes,
ported as real `core.Program` values, not `.lang` fixtures) as an
**additional synthetic source**, not a second harness. Do not stand up a
parallel `peerDivergenceExpected`-equivalent map for the synthetic corpus —
extend this test's own source-enumeration loop to also iterate the generated
programs, applying the same two-directional exactness discipline.

---

### 8. Third gate-eligible metric chokepoint widening

**Both chokepoints, current state (already widened once by Phase 08, per
D-09-28's finding)** — `measure/statistics.go:130-132`:
```go
func GateEligibleMetrics() []string {
	return []string{"recomputed_work", "recomputed_work_growth_exponent"}
}
```
`measure/statistics.go:160-174`, `Demote` (already consulting a set, not a
literal — this is Phase 08's own D-08-32 fix already landed):
```go
func Demote(requested string, metric string, summary Summary, err error) string {
	if err != nil {
		return VerdictNotRatified
	}
	if !gateEligibleMetricSet()[metric] {
		return VerdictObserved
	}
	if summary.CoV > CoVDemotionThreshold {
		return VerdictObserved
	}
	if requested == VerdictBlocking {
		return VerdictBlocking
	}
	return requested
}
```
`session/session_phase6_budget.go:46-66`:
```go
func QLT02MetricVocabulary() []string {
	return []string{"recomputed_work", "elapsed_ns", "output_bytes", "recomputed_work_growth_exponent"}
}
// ...
func QLT02GateEligibleMetrics() []string {
	return []string{"recomputed_work", "recomputed_work_growth_exponent"}
}
```
**The agreement test to re-derive, not just re-run**
(`session_phase6_budget_test.go:237-252`, verbatim):
```go
func TestGateEligibleMetricSetsAgreeAcrossChokepoints(t *testing.T) {
	sessionSet := QLT02GateEligibleMetrics()
	measureSet := measure.GateEligibleMetrics()
	if len(sessionSet) != len(measureSet) {
		t.Fatalf("... not set-equal")
	}
	// ... every session-side member must appear in measure-side ...
}
```
**Extension (D-09-28):** add a **third, distinct** metric name for the peer's
own closure-curve bound (e.g. `peer_closure_recomputed_work` or similar — an
exact identifier is planner discretion) to **both** `GateEligibleMetrics()`
and `QLT02GateEligibleMetrics()` (and `QLT02MetricVocabulary()` if the name is
genuinely new rather than reused) in the same wave, then re-run/re-derive
`TestGateEligibleMetricSetsAgreeAcrossChokepoints` unchanged in shape — it
already forbids a one-sided widening by construction, no new test logic
needed, just a new passing name on both sides.

**Manifest row to add** (`session/qlt02_budget_manifest.json`, existing row
shape, verbatim, to copy for the new metric):
```json
{
  "machine_id": "machine:4797d76b7863",
  "metric": "recomputed_work_growth_exponent",
  "gate_type": "hard",
  "value_or_bound": 1200,
  "unit": "milliexponent",
  "ratified_at": "2026-09-10T03:27:15Z",
  "ratified_by_commit": "b28a92f927e40f5c0c6279dfef1368c660fe75c5"
}
```
Ratify the peer's own new row at the mid-phase gate, per D-09-28, with its own
distinct `unit` (never `"count"` — the peer's cost curve is a reachability
closure, not a worklist, so it must not borrow `check`'s bound value).

**Risk lane to add or extend** (`session/risk_lanes.json:230-235`, the
existing `lane:interprocedural-cost-scaling` entry, verbatim):
```json
{
  "fixture_kind": "phase8_interprocedural",
  "lane_id": "lane:interprocedural-cost-scaling",
  "declared_inputs": ["go_toolchain"],
  "rationale": "A changed-risk lane (D-08-36) firing when internal/compiler/check's interprocedural summary/liveness code or internal/compiler/callgraph changes: ..."
}
```
Either widen this lane's rationale/scope to also cover `corevalidate`'s new
closure-curve code, or add a sibling lane — planner discretion; either way
follow this exact JSON row shape.

---

### 9. Import-independence guard extension

**Four existing guards, verified at the cited locations:**

1. `corevalidate_cycle_peer_test.go:137-155`
   (`TestCyclePeerTestFileImportsStayIndependent`) — scans this one test
   file's own imports.
2. `corevalidate_endpoint_internal_test.go` (package-internal test, the
   loop confirmed around line 107) — walks `.go` files in the test directory
   and rejects any import path suffix matching `/compiler/check`,
   `/compiler/ast`, `/compiler/originvalidate`, or `/compiler/callgraph`
   (verbatim, this exact function):
   ```go
   if strings.HasSuffix(path, "/compiler/check") || strings.HasSuffix(path, "/compiler/ast") || strings.HasSuffix(path, "/compiler/originvalidate") || strings.HasSuffix(path, "/compiler/callgraph") {
       t.Fatalf("%s imports %s, which corevalidate must never depend on", entry.Name(), path)
   }
   ```
3. `corevalidate_exclusive_test.go:200-216` (`TestAttributeValidatorImportsStayIndependent`)
   — string-scans `corevalidate.go`'s own source text for `"compiler/check"` and
   `"compiler/cgen"`.
4. `corevalidate_test.go:144-152` — the broadest scan, six forbidden
   substrings against `corevalidate.go`'s source text:
   ```go
   for _, forbidden := range []string{"compiler/ast", "compiler/check", "compiler/interp", "compiler/cgen", "compiler/session", "compiler/ability"} {
       if strings.Contains(string(source), forbidden) {
           t.Fatalf("validator imports forbidden producer/engine package %q", forbidden)
       }
   }
   ```

**Which the new code falls under:** the new liveness/closure code lives
inside `corevalidate.go` itself (or a sibling production file in the same
package) — guard 4 (`corevalidate_test.go:148`) and guard 3
(`corevalidate_exclusive_test.go:216`) already cover any production file
matching `corevalidate.go`'s source text scan **only if the scan target stays
`corevalidate.go` literally**; if the new derivation lands in a sibling file
per D-09-04's discretion, **that sibling file's path must be added to guards
3 and 4's scanned-file list**, or a fifth guard authored for it — do not
assume a new file is automatically covered. New test files driving the new
derivation (e.g. a `corevalidate_peer_liveness_test.go`) fall under guard 1's
pattern if they import `check` as an ordinary same-package-external-test
dependency (external `_test` package, per `corevalidate_test.go:1-2`'s own
documented distinction) to drive a companion assertion.

---

### 10. `09-VALIDATION.md` and `PHASE-09-DEBT.md`

**Analog — frontmatter and Per-Task Verification Map shape**
(`08-VALIDATION.md:1-9`, verbatim frontmatter):
```yaml
---
phase: "08"
slug: "interprocedural-loan-liveness-in-check"
status: draft
nyquist_compliant: false
wave_0_complete: true
created: "2026-09-09"
---
```
followed by `## Test Infrastructure`, `## Sampling Rate`, and `## Per-Task
Verification Map` (a `| Task ID | Plan | Wave | Requirement | Threat Ref |
Secure Behavior | Test Type | Automated Command | File Exists | Status |`
table). Phase 09's `09-VALIDATION.md` must additionally carry the
**"M001 Phase 3 Debt Closure (loan-liveness subset)"** section D-09-42
mandates, citing the specific 03-03/03-04/03-05 rows satisfied, with its own
scoped `nyquist_compliant` flag — a section shape with no existing sibling
example in-tree, since no prior phase has closed an inherited M001 debt item;
follow `03-VALIDATION.md`'s own Per-Task Verification Map row shape when
quoting it (read-only, never amend in place, per D-09-41).

**Analog — debt register frontmatter and item-table shape**
(`PHASE-08-DEBT.md:1-9`, verbatim frontmatter):
```yaml
---
phase: 08-interprocedural-loan-liveness-in-check
recorded: 2026-09-09
status: accepted
disposition: gate-adjudicated
items: 9
blocking: 0
---
```
followed by `## Items` (`| ID | Source | Threat/Req | Severity | Landing
phase | Item |`) and one `### D-08-NN` prose section per row (verified —
`PHASE-08-DEBT.md` has exactly 9 rows and 9 matching `###` sections, e.g.
`### D-08-15`, `### D-08-26` ... `### D-08-43`). **Mechanically enforced by**
`TestDebtRegistersAreWellFormed` (`session_test.go:2613-2626`, calling
`checkDebtRegister` per discovered `*-DEBT.md` file), which validates:
frontmatter starts with `---\n`, has a terminated block, an `items:` count
line parseable via `fmt.Sscanf("items: %d", ...)` that must equal the `##
Items` table's row count, and (per D-09-44's own required-format list) one
`### <ID>` detail section per row with severity from `{blocker, warning,
info}`.

`PHASE-09-DEBT.md` must be written **at planning time** (D-09-44), carrying at
minimum: D-09-08, D-09-13, D-09-21, D-09-31, D-09-37, D-09-40, D-09-43, and
the resolutions of D-08-26/D-08-27/D-08-40 (which `PHASE-08-DEBT.md`'s own
`D-08-26`/`D-08-27`/`D-08-40` rows already point at "Phase 09" as landing
phase — those rows' own `### D-08-NN` prose is the direct source text to
carry forward/resolve).

## Shared Patterns

### Forward set-propagation over `function.Linear.Operations`, one map, no recursion
**Source:** `corevalidate.go:2061-2131` (`peerParameterEscapesOwned`,
`peerReturnDerivesFromBorrow`).
**Apply to:** work items 1 and 3 — every new peer-derived fact this phase
adds. Never a backward walk (that is `originvalidate`'s and `check`'s shape),
never a `callgraph` import.

### Callee-before-caller consumption via the peer's OWN postorder byproduct
**Source:** `corevalidate.go:130-138` (`v.peerPostorder`'s doc comment),
`:495-549` (`chainPeerClosureDigests`, the consuming loop).
**Apply to:** work item 1's fold-in point. Never `callgraph.Order` (forbidden
import, D-09-05).

### Unexported package-level bool seam + defer-restored companion-assertion test, run in BOTH directions
**Source:** `check.go:1183`/`:1250` (`verifyCallableRefusalSeam`),
`check_test.go:3322` (`TestVerifyCallableRefusalSeamCheckDisabledCorevalidateStillRefuses`).
**Apply to:** work item 5 — both the existing-direction generalization and the
new `corevalidate`-side seam. State the Knight & Leveson discrimination
argument in the new test's own doc comment (D-09-25).

### Exact-set (both-directions) divergence/agreement gating — extend the consumer, never fork the harness
**Source:** `session_peer_gate_test.go:23` (`peerDivergenceExpected`), `:218`
(`TestNoUndeclaredCheckPeerDivergenceAcrossCorpus`); `session_phase6_budget_test.go:237`
(`TestGateEligibleMetricSetsAgreeAcrossChokepoints`).
**Apply to:** work items 7 and 8. A new source of programs/metrics feeds an
existing exactness test; it never gets its own parallel exactness harness.

### Two independent chokepoints must widen together, or a "hard" gate is decorative
**Source:** `measure/statistics.go:130-132` + `session/session_phase6_budget.go:64-66`,
proven equal by `TestGateEligibleMetricSetsAgreeAcrossChokepoints`.
**Apply to:** work item 8's third widening — the exact trap D-08-32 already
caught once; do not widen only one side.

### Import independence is a static-scan test, not a design guideline
**Source:** the four guards under work item 9.
**Apply to:** any new file in `corevalidate` — verify it is covered by an
existing scan's file list, or extend a scan's file list / author a fifth
guard. Never rely on code review alone.

### Debt/validation documents follow a mechanically-checked frontmatter + table + per-row-detail-section shape
**Source:** `PHASE-08-DEBT.md` (whole file), `TestDebtRegistersAreWellFormed`
(`session_test.go:2613`).
**Apply to:** work item 10. Write `PHASE-09-DEBT.md` at planning time, not
phase end (D-09-44), and let the existing test validate its shape rather than
hand-checking row/section correspondence.

## No Analog Found

None among the ten work items above — every one has at least one same-file or
same-package existing instance to copy the shape of, following the same
pattern Phase 08's own PATTERNS.md observed ("every mechanism this phase
needs already exists in the tree in miniature"). The two items requiring an
explicit contrasting-shape callout (1 and 3) are documented above precisely
because the risk here is not "no analog exists" but "the wrong analog is
one function away and looks tempting to copy."

One partial gap worth naming for the planner: **no existing in-tree section
shape closes an inherited M001-milestone debt item from within a later
milestone's phase document** (work item 10's "M001 Phase 3 Debt Closure"
section, D-09-42). `08-VALIDATION.md` and `07-VALIDATION.md` are both
same-milestone (M002) documents with no cross-milestone closure precedent to
copy structurally — the planner should treat this section's exact heading and
field shape as new prose, anchored only by `TestDebtRegistersAreWellFormed`'s
unrelated debt-register format and by `03-VALIDATION.md`'s own frontmatter
fields being quoted (never edited) from the read-only source.

## Metadata

**Analog search scope:** `internal/compiler/corevalidate/` (all `.go` files),
`internal/compiler/check/check.go` and `check_test.go`,
`internal/compiler/originvalidate/originvalidate.go`,
`internal/compiler/session/` (`session_peer_gate_test.go`,
`session_phase6_budget.go`, `session_phase6_budget_test.go`,
`qlt02_budget_manifest.json`, `risk_lanes.json`, `session_test.go`),
`internal/compiler/measure/statistics.go`,
`.planning/phases/08-interprocedural-loan-liveness-in-check/` (`08-VALIDATION.md`,
`PHASE-08-DEBT.md`).
**Files scanned:** 16 (all confirmed git-tracked via `git ls-files`).
**Pattern extraction date:** 2026-09-10
