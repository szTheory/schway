# Phase 08: Interprocedural Loan Liveness in `check` - Pattern Map

**Mapped:** 2026-09-09
**Files analyzed:** 10 (9 modified, 1 new corpus/generator area; test files counted separately)
**Analogs found:** 10 / 10 (all in-tree, all tracked)

This phase is unusual: CONTEXT.md and RESEARCH.md already pin every mechanism
to a specific existing function this phase edits in place. There is
effectively **no greenfield file** — every "new" thing (summary bits, seam,
diagnostic code, cost metric) is an additive clause inside an existing
tracked file, following a pattern instance that already exists one function
away. All paths below were verified tracked via `git ls-files`.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog (same file, different function) | Match Quality |
|---|---|---|---|---|
| `internal/compiler/check/check.go` — `buildCallSignatureTable` extension (two new bits) | service (compiler pass, table builder) | batch (once per `check.Program` run, memoized) | same file: `Callable` bit already on `callSignatureTable`/`core.FunctionSignature`, built post-body-check | exact (same table, same shape of fact) |
| `internal/compiler/check/check.go` — `derivePlaceLoans` `OpCall` branch | transform (forward dataflow canonicalization) | transform (single forward pass over `core.LinearOperation`s) | same file: existing reborrow-chain branch (`OpBorrowShared`/`OpBorrowExclusive`) in the same function | exact |
| `internal/compiler/check/check.go` — `blockLoanLiveness` `OpCall` clause | transform (backward dataflow transfer function) | transform | same file: existing chain-ancestor "use" walk over `operation.SourceID` | exact |
| `internal/compiler/check/check.go` — `loanLivenessFixpoint` iteration bound + explicit-stack conversion | service (fixpoint driver / control-flow bound) | event-driven (worklist) | same file: existing `work` counter accounting in the same function; `callgraph.Order`'s iterative 3-color DFS for the stack conversion | exact (counter), role-match (stack conversion, cross-file) |
| `internal/compiler/check/check.go` — `computeLoanLastUses` `"call"` case | transform (AST-shadow admission path) | transform | same file: existing `"take"`/`"borrow"`/`"borrow_mut"` cases in the same `switch` | exact |
| `internal/compiler/check/check.go` — new diagnostic `check.interprocedural_loan_liveness` + cause chain | controller (diagnostic construction) | request-response (per-check-run refusal) | same file: `borrowConflictDiagnostic` (`check.go:~2793-2830`) | exact |
| `internal/compiler/check/check.go` — new unexported seam for the iteration bound | config/utility (fault-injection seam) | event-driven (test-only flip) | same file: `callReturnTypeDerivationSeam` (`check.go:493`) | exact |
| `internal/compiler/check/check_test.go` — new twin-pair, relay-depth, negative-control, memo, program-order, mutation-kill tests | test | request-response (table-driven checks) | same file: `TestReborrowChainWorkIsLinear` (`:737`), `TestCallReturnTypeDerivationMutationKilled` (`:3412`), `TestRelayEscortWitnessChecksCleanPendingInterproceduralLiveness` (`:2338`, to be flipped) | exact |
| `internal/compiler/measure/statistics.go` — widen `Demote`'s `"recomputed_work"` chokepoint | utility (verdict computation) | transform | same file: `Demote` itself (widen in place), guarded by `TestDemoteHasExactlyOnePromotionPassthrough` | exact |
| `internal/compiler/session/session_phase6_budget.go` — widen `QLT02GateEligibleMetrics()` | config (gate-eligibility declaration) | transform | same file: the function itself (widen in place) | exact |
| `internal/compiler/session/qlt02_budget_manifest.json` | config (data, not code) | batch | existing rows (`recomputed_work` hard, `elapsed_ns`/`output_bytes` observed) | exact |
| `internal/compiler/session/risk_lanes.json` | config (data) | batch | existing 38 lane entries, e.g. `lane:deterministic` | exact |
| new cost-corpus generator (Go file under `internal/compiler/check` or `internal/compiler/session`, test-only) | utility (synthetic test-fixture generator) | batch/transform | `.planning/spikes/006-interprocedural-liveness-cost-scaling/iplive/corpus.go` (`Generate(shape, n)`) + `scaling.go` (`Growth.ExponentInOps`) — **port, do not import** (separate throwaway module) | exact (source to port) |
| `testdata/phase08/*.lang` (new fixtures: twin pairs A/B, relay-depth-2, negative control) | test fixture (data) | request-response | `testdata/phase3/shared_shared_accept.lang`, `testdata/phase3/public_view.lang` family (borrow/take binding grammar reused unchanged) | exact |
| `testdata/phase07/relay_escort_witness.lang` | test fixture (data, unchanged content — only its consuming test's assertion flips) | request-response | itself — no analog needed, this is the fixture | n/a (pre-existing, assertion flips in `check_test.go`) |

## Pattern Assignments

### `internal/compiler/check/check.go` — summary-bit production (extends `buildCallSignatureTable`)

**Analog:** the existing `Callable` bit on `callSignatureTable` / `core.FunctionSignature`, same file, `check.go:560-631`.

**Existing table shape to extend** (`check.go:594-598`):
```go
type callSignatureTable struct {
	entries map[string]core.FunctionSignature
}

func (t callSignatureTable) lookup(calleeID string) (core.FunctionSignature, bool) {
	entry, ok := t.entries[calleeID]
	return entry, ok
}
```
Per RESEARCH.md's Open Question 1 (planner discretion, D-08-01..06): do **not** widen `core.FunctionSignature` itself (that is the schema bump D-08-04 forbids). Add a sibling in-memory map keyed identically to `entries`, e.g. `type interproceduralSummary struct { usesParam, returnsBorrowOfParam bool }` plus `summaries map[string]interproceduralSummary`, populated in `callgraph.Order`'s RPO (`internal/compiler/callgraph/callgraph.go:294`) callee-before-caller, memoized for the run only.

**Builder precedent to mirror** (`check.go:619-631`):
```go
func buildCallSignatureTable(program core.Program) (callSignatureTable, error) {
	if callSignatureTableBuildObserved != nil {
		callSignatureTableBuildObserved()
	}
	iface, err := originvalidate.BuildInterface(program)
	if err != nil {
		return callSignatureTable{}, err
	}
	entries := make(map[string]core.FunctionSignature, len(iface.Functions))
	for _, signature := range iface.Functions {
		entries[signature.ID] = signature
	}
	return callSignatureTable{entries: entries}, nil
}
```
Same shape: build once, no exported/unexported setter after construction (doc comment at `check.go:587-593` states the immutability-by-construction invariant this new map must also honor).

**Doc-comment precedent to copy the style of** (`check.go:599-618`, the paragraph explaining `Callable`'s "body-derived bit on a body-blind table" shape) — write the equivalent paragraph for `UsesParam`/`ReturnsBorrowOfParam`, citing D-08-05's two-tier consumption/production reading explicitly (per D-08-05's own instruction: "Write this into the phase's decision record explicitly").

---

### `internal/compiler/check/check.go` — forward canonicalization (`derivePlaceLoans` `OpCall` branch)

**Analog:** the existing reborrow-chain branch in the same function, `check.go:1049-1068` (verbatim, verified):
```go
func derivePlaceLoans(operations []core.LinearOperation) placeLoanChain {
	chain := placeLoanChain{
		latestLoan: make(map[string]string, len(operations)),
		parentLoan: make(map[string]string, len(operations)),
	}
	for _, operation := range operations {
		inherited := chain.latestLoan[operation.SourceID]
		if operation.Kind == core.OpBorrowShared || operation.Kind == core.OpBorrowExclusive {
			chain.parentLoan[operation.LoanID] = inherited
			if operation.TargetID != "" {
				chain.latestLoan[operation.TargetID] = operation.LoanID
			}
			continue
		}
		if inherited != "" && operation.TargetID != "" {
			chain.latestLoan[operation.TargetID] = inherited
		}
	}
	return chain
}
```
**Extension (D-08-07):** add an `operation.Kind == core.OpCall` branch before the generic fallthrough: look up the callee's `ReturnsBorrowOfParam` bit from the new summary map (via `operation.CalleeID`, per D-07 call-graph plumbing), and if true, `chain.latestLoan[operation.TargetID] = chain.latestLoan[operation.SourceID]` — otherwise let the existing generic fallthrough run unmodified (it already produces the correct fresh-identity result since `inherited` is `""` for a fresh call result). **Do not touch the backward transfer function for this fact** — see Anti-Pattern below.

---

### `internal/compiler/check/check.go` — backward "use" clause (`blockLoanLiveness` `OpCall` clause)

**Analog:** the existing chain-ancestor use-walk in the same function, `check.go:1094-1099` (verbatim):
```go
for loan := chain.latestLoan[operation.SourceID]; loan != "" && !recorded[loan]; loan = chain.parentLoan[loan] {
    work++ // one unit per chain-ancestor step walked
    live[loan] = true
    recorded[loan] = true
    uses = append(uses, loanBlockUse{loanID: loan, operationIndex: index, operationID: operation.ID})
}
```
**Extension (D-08-08):** an `OpCall` whose callee's `UsesParam` bit is true enters this same walk over `operation.SourceID` — no new loop, no new lattice value, just widen the condition that gates entry into this existing walk to include `OpCall` alongside whatever already triggers it (a plain reference read).

---

### `internal/compiler/check/check.go` — fixpoint bound + explicit-stack conversion (`loanLivenessFixpoint`)

**Analog:** the function's own existing `work` counter (`check.go:1142-1225`, verbatim above in full) — extend with the derived bound `k × len(blocks) × distinctLoanCount` (D-08-14) checked against the running `work` total inside the `for len(queue) > 0` loop, refusing fail-closed via the new `check.interprocedural_loan_liveness`-adjacent seam (see below) rather than an ad hoc panic.

**Native-recursion → explicit-stack analog:** `internal/compiler/callgraph/callgraph.go:294`'s `Order` — its iterative three-color DFS is the shape to port for converting the pre-walk at `check.go:1157-1180`:
```go
const (
    unvisited = 0
    inWork    = 1
    done      = 2
)
state := make(map[string]int, len(blocks))
var walk func(id string) error
walk = func(id string) error {
    switch state[id] {
    case inWork:
        return fmt.Errorf("check.cfg_back_edge: block %q participates in a cycle", id)
    case done:
        return nil
    }
    state[id] = inWork
    for _, successor := range byID[id].successors {
        if err := walk(successor); err != nil {
            return err
        }
    }
    state[id] = done
    return nil
}
```
This `walk` closure (currently native-recursive) is the exact target of D-08-19a's explicit-stack conversion, and its bare `fmt.Errorf("check.cfg_back_edge: ...")` at the `inWork` case is the exact target of D-08-19b's promotion to a coded `diagnostic.Diagnostic`. Read `callgraph.Order`'s own explicit-stack implementation (`internal/compiler/callgraph/callgraph.go`, three-color state machine over an explicit slice-as-stack) as the structural template — same three-state coloring, same "push/pop a stack slice" replacement for the call-stack frame.

---

### `internal/compiler/check/check.go` — `computeLoanLastUses` `"call"` case (D-08-09, AST-shadow path)

**Analog:** the function's own existing `switch` cases, `check.go:2996-3006` (verbatim):
```go
kind := core.OpCopy
loanID := ""
switch binding.RHS.Kind {
case "take":
    kind = core.OpMove
case "borrow":
    kind = core.OpBorrowShared
    loanID = fmt.Sprintf("shadow:loan:%d", index)
case "borrow_mut":
    kind = core.OpBorrowExclusive
    loanID = fmt.Sprintf("shadow:loan:%d", index)
}
```
**Extension:** add `case "call":` mirroring the semantics the real `derivePlaceLoans`/`blockLoanLiveness` path now implements (alias-if-`ReturnsBorrowOfParam`, use-if-`UsesParam`) — via the parallel `buildCalleeContracts`/`calleeContract` table already used by this AST-shadow path (`check.go:552-572`), not the `core.FunctionSignature`-based table (this path predates `core.LinearOperation` and works over `ast.LinearBody` directly). **This must land alongside the `derivePlaceLoans` fix, with a differential test that fails if only one path is fixed** (D-08-09) — see the test pattern below.

---

### `internal/compiler/check/check.go` — new diagnostic `check.interprocedural_loan_liveness`

**Analog:** `borrowConflictDiagnostic`, same file, verbatim (`check.go:~2793-2822`):
```go
func borrowConflictDiagnostic(span diagnostic.Span, blocking *loanState, ownerID, ownerTypeID string, binding ast.Binding) diagnostic.Diagnostic {
	causes := []diagnostic.Cause{
		{Kind: "borrow_created_here", Span: spanPointer(blocking.borrowedAt)},
		{Kind: "borrow_used_later", Span: spanPointer(blocking.lastUseSpan)},
		{Kind: "loan", Detail: blocking.id},
		{Kind: "owner", Detail: ownerID},
		{Kind: "type", Detail: ownerTypeID},
	}
	repairs := []diagnostic.Repair{{Kind: "create_loan_after_conflicting_loan_ends"}}
	...
	return diagnostic.ErrorWithRepairs(
		"ownership.borrow_conflict", span, "cannot create a loan while a conflicting loan is live", causes,
		repairs...,
	)
}
```
**New function shape (D-08-20..25):**
- Mint code `"check.interprocedural_loan_liveness"` (never reuse `ownership.*`).
- `Primary` = the failing operation's span (the `take`), via `spanByOperationID` (already used for cycle spans — D-07-51/WR-02 confirms it is directly reusable).
- Fixed three-role `Causes` (D-08-23), **ID-only and spanless for cause 3** (D-08-24 — `diagnostic.Span{Start, End int}` has no file field, so a cross-function callee cause cannot carry a span; follow the existing spanless-ID shape already in this file, `{Kind: "loan", Detail: blocking.id}` at `check.go:2809`):
  1. `{Kind: "borrow_created_here", Span: <borrow span, caller>}`
  2. `{Kind: "loan_extended_by_call", Span: <call-site span, caller>, Detail: "<calleeID>"}`
  3. `{Kind: "callee_return_contract", Detail: "<calleeID>:return.mode=<Mode>"}` or `":parameters[0].mode=<Mode>"`
- Use `diagnostic.Error` (not `ErrorWithRepairs`) — `Repairs: nil` per D-08-25, since `borrowConflictDiagnostic`'s repair-bearing shape does not apply (the fix lives in a different function than `Primary`). Follow `ownership.borrow_requires_share` / `ownership.transfer_requires_take`'s classification-only precedent instead.
- **Never embed the bound's numeric value in a `Cause` or `Message`** (D-08-18) — `Causes` fold into diagnostic ID identity (`internal/compiler/diagnostic/diagnostic.go:96-112`).

---

### `internal/compiler/check/check.go` — new fault-injection seam for the iteration bound

**Analog:** `callReturnTypeDerivationSeam`, same file, verbatim (`check.go:493`, declaration) and (`check.go:1928`, read site):
```go
// Unexported, false in production, set only from a same-package test that
// defers the restore immediately (see TestCallReturnTypeDerivationMutationKilled).
var callReturnTypeDerivationSeam = false
```
read site:
```go
if callReturnTypeDerivationSeam {
    derivedTypeID = argument.place.TypeID
} else {
    ...
}
```
**New seam (D-08-16):** e.g. `var loanLivenessBoundSeam = false` (or a small override struct), unexported package-level var, false in production, read inside `loanLivenessFixpoint`'s worklist loop to force the bound to trip on an otherwise-unreachable-from-source input. `loanLivenessFixpoint` already takes `[]cfgBlockSpec` (an unexported type), so any test exercising it is same-package by construction — the seam route "fits naturally" per D-08-16's own text. **Not** the exported `pathoracle.TerminatorKindsOverride` shape (`internal/compiler/pathoracle/pathoracle.go:51,59`).

---

### `internal/compiler/check/check_test.go` — new tests

**Analog 1 — work-ratio / growth-exponent pinning test**, verbatim (`check_test.go:737-753`):
```go
func TestReborrowChainWorkIsLinear(t *testing.T) {
	series := []int{10, 100, 1_000, 10_000}
	work := make([]int, len(series))
	for index, n := range series {
		block := cfgBlockSpec{id: "chain:block:ratio", operations: reborrowChainOperations(n), successors: nil}
		result, err := loanLivenessFixpoint("chain", []cfgBlockSpec{block})
		...
		work[index] = result.work
	}
	for index := 1; index < len(series); index++ {
		operationRatio := float64(series[index]) / float64(series[index-1])
		workRatio := float64(work[index]) / float64(work[index-1])
		if workRatio > operationRatio*2 {
			t.Fatalf(...)
		}
	}
}
```
Use this exact pattern for `TestSummaryDerivationRequiresProgramOrder` (D-08-11, program-order-reversed-vs-forward comparison) and for the summary-derivation cost test.

**Analog 2 — seam mutation-kill test**, verbatim (`check_test.go:3412-3425+`):
```go
func TestCallReturnTypeDerivationMutationKilled(t *testing.T) {
	defer func() { callReturnTypeDerivationSeam = false }()
	...
	callReturnTypeDerivationSeam = true
	op, target, diag := resolveCallBinding(...)
	...
}
```
Use this exact defer-and-restore shape for the new `loanLivenessBoundSeam` mutation-kill test (D-08-16, QLT-08 mutation-kill discipline).

**Analog 3 — the fixture that must flip**, verbatim (`check_test.go:2338-2364`, name and assertion both change):
```go
func TestRelayEscortWitnessChecksCleanPendingInterproceduralLiveness(t *testing.T) {
	source := readPhase07Fixture(t, "relay_escort_witness.lang")
	result := Program(mustParseProgram(t, source))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected zero error diagnostics (the D-03-02 interprocedural finding), got %+v", result.Diagnostics)
	}
	...
}
```
D-08-28.3 requires this renamed (e.g. `TestRelayEscortWitnessRefusesInterproceduralLiveness`) and re-asserted to expect exactly one `check.interprocedural_loan_liveness` diagnostic once D-08-09 lands.

**New differential test for D-08-09 (no direct in-tree analog — construct from the two admission call sites):** a fixture engineered so `computeLoanLastUses` (called at `check.go:1328` and `:2529`) and the real `derivePlaceLoans` path disagree if only one is fixed — following the structural shape of `TestCallReturnTypeDerivationMutationKilled` (construct two derivations designed to disagree, assert they agree post-fix).

---

### `internal/compiler/measure/statistics.go` — widen `Demote`'s chokepoint

**Current code to widen** (`statistics.go:120-135`, verbatim):
```go
func Demote(requested string, metric string, summary Summary, err error) string {
	if err != nil {
		return VerdictNotRatified
	}
	if metric != "recomputed_work" {
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
**Extension (D-08-32):** widen `if metric != "recomputed_work"` to a set-membership check against the same eligible-metrics list `QLT02GateEligibleMetrics()` returns (or a locally mirrored constant set), so a newly named growth-exponent metric is not silently demoted. **The `TestDemoteHasExactlyOnePromotionPassthrough` structural guard must be re-derived**, not just re-run — its own doc comment states it scans for "exactly one `return VerdictBlocking`" via `go/ast`:
```go
func TestDemoteHasExactlyOnePromotionPassthrough(t *testing.T) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "statistics.go", nil, 0)
	...
	ast.Inspect(demoteFunc.Body, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok || len(ret.Results) != 1 { return true }
		ident, ok := ret.Results[0].(*ast.Ident)
		if ok && ident.Name == "VerdictBlocking" { count++ }
		return true
	})
	...
}
```
Widening `Demote` must not add a second `return VerdictBlocking` — the re-derived test must keep asserting exactly one.

---

### `internal/compiler/session/session_phase6_budget.go` — widen `QLT02GateEligibleMetrics()`

**Current code** (`session_phase6_budget.go:55-57`, verbatim):
```go
func QLT02GateEligibleMetrics() []string {
	return []string{"recomputed_work"}
}
```
**Extension (D-08-32):** widen the returned slice to include the new growth-exponent metric name(s), consistent with whatever name Wave 4 mints (e.g. `recomputed_work_per_op_exponent`) — this is the second of the two chokepoints that must widen together or the manifest row is decorative. Cross-reference `QLT02MetricVocabulary()` (`session_phase6_budget.go:~44-47`, the closed three-value set `{"recomputed_work", "elapsed_ns", "output_bytes"}`) — if the new metric is a genuinely new name rather than reusing `recomputed_work`, that vocabulary constant likely needs widening too (verify at implementation time; not spelled out as a separate decision in CONTEXT.md, but structurally implied by "a row naming a metric outside this set is malformed").

---

### `internal/compiler/session/qlt02_budget_manifest.json` — new row(s)

**Existing row shape to copy** (verbatim, current file):
```json
{
  "machine_id": "machine:4797d76b7863",
  "metric": "recomputed_work",
  "gate_type": "hard",
  "value_or_bound": 4,
  "unit": "count",
  "ratified_at": "2026-09-07T03:50:03Z",
  "ratified_by_commit": "e9de50ec8dcd106497e3658ca30c5914efb5ade2"
}
```
New row(s) for the growth-exponent metric: same shape, `"gate_type": "hard"`, `"unit"` appropriate to the fitted value (e.g. a bare ratio/exponent — no existing `unit` string for this; planner discretion, but keep it distinct from `"count"`/`"nanoseconds"`/`"bytes"`). Ratify at the mid-phase gate per D-08-33 (before the liveness law is declared final), with a live `machine_id` probe present for the invocation.

---

### `internal/compiler/session/risk_lanes.json` — new `lane:interprocedural-cost-scaling`

**Existing entry shape to copy** (verbatim, one of 38 current entries):
```json
{
  "fixture_kind": "pure_match",
  "lane_id": "lane:deterministic",
  "declared_inputs": ["fixture_source"],
  "rationale": "Determinism replay only re-parses/re-runs the same fixture source; nothing else it emits can move independently."
}
```
New entry: `"lane_id": "lane:interprocedural-cost-scaling"`, firing on changes to `internal/compiler/check`'s interprocedural summary/liveness code or `internal/compiler/callgraph` (D-08-36), with a `rationale` explaining it is a changed-risk lane (not the default edit loop) because a multi-shape size sweep is too costly for the default loop per `wiki/compiler-and-feedback-latency.md`.

---

### New cost-corpus generator (test-only, port not import)

**Source to port** (throwaway spike module, NOT importable — separate `go.mod`): `.planning/spikes/006-interprocedural-liveness-cost-scaling/iplive/corpus.go`, verbatim excerpt:
```go
func leafUse(id string) Function {
	return Function{
		ID: id, ParamLoan: "p",
		Blocks: []Block{{ID: "entry", Ops: []Operation{{ID: id + ":0", Kind: OpUse, Loan: "p"}}}},
	}
}

func leafPass(id string) Function {
	return Function{
		ID: id, ParamLoan: "p",
		Blocks:  []Block{{ID: "entry", Ops: []Operation{}}},
		Returns: "p",
	}
}

func relay(id string, probeKind string, callees ...string) Function {
	fn := Function{ID: id, ParamLoan: "p"}
	ops := []Operation{}
	for i, callee := range callees {
		result := fmt.Sprintf("r%d", i)
		ops = append(ops, Operation{
			ID: fmt.Sprintf("%s:call%d", id, i), Kind: OpCall,
			Loan: "p", Callee: callee, Result: result,
		})
	}
	...
	switch probeKind {
	case "after":
		ops = append(ops,
			Operation{ID: id + ":borrow", Kind: OpBorrow, Loan: "l", Place: "own"},
			Operation{ID: id + ":probe", Kind: OpCall, Loan: "l", Callee: callees[0], Result: "pr"},
			Operation{ID: id + ":move", Kind: OpMove, Place: "own"},
			Operation{ID: id + ":useres", Kind: OpUse, Loan: "pr"},
		)
	case "before":
		ops = append(ops,
			Operation{ID: id + ":borrow", Kind: OpBorrow, Loan: "l", Place: "own"},
			Operation{ID: id + ":move", Kind: OpMove, Place: "own"},
			Operation{ID: id + ":probe", Kind: OpCall, Loan: "l", Callee: callees[0], Result: "pr"},
		)
	}
	...
}
```
Port this `Generate(shape, n)` template style — real `core.Program`/`core.LinearOperation` values built directly in Go, never through the `.lang` parser (D-08-34) — into a new **test-only** file in `internal/compiler/check` (or `internal/compiler/session`, planner discretion), producing `chain`/`diamond`/`dense`/`forward` shapes per D-08-35 (drop `tree`). Also port `scaling.go`'s `Growth.ExponentInOps` fitting logic for the growth-exponent computation consumed by the new `TestQLT02InterproceduralGrowthExponent`-style test (RESEARCH.md's test map, Wave 4).

---

### `testdata/phase08/*.lang` — new fixtures

**Analog:** `testdata/phase3/shared_shared_accept.lang`, `testdata/phase3/public_view.lang` and the `public_view_*` family — reuse the existing borrow/take binding grammar **unchanged**. Read one of these to confirm exact `let`/`borrow`/`take`/`call` syntax before authoring the twin pairs; do not invent new grammar.

**Required fixtures per D-08-28** (mechanical checklist, not free-form):
1. Pattern A twin (refuse iff callee `ReturnContract.Mode ∈ {shared, exclusive}`, accept iff `"owned"`) — differ in **exactly** the callee's declared return-mode field.
2. Pattern B twin (`borrow; move; call`, refuse iff callee uses its parameter) — differ in **exactly** the callee's declared parameter-mode/body-uses-param fact.
3. Depth-≥2 relay chain (`caller → relay → leaf`, relay's `Return.Mode` forwards leaf's).
4. Negative-control pair varying only `Fails`/`Foreign.*` — both members must resolve identically.
5. (Should-add, not gate-blocking) both-match-arms call fixture.

## Shared Patterns

### Fault-injection seams — unexported, same-package, defer-restored
**Source:** `internal/compiler/check/check.go:493` (`callReturnTypeDerivationSeam`), test usage `check_test.go:3412`.
**Apply to:** the new iteration-bound seam (D-08-16).
```go
var callReturnTypeDerivationSeam = false
// ... read at a call site ...
// ... in a same-package test:
defer func() { callReturnTypeDerivationSeam = false }()
callReturnTypeDerivationSeam = true
```

### Diagnostic construction — span-bearing causes first, then ID-bearing detail causes, `Primary` on the failing operation
**Source:** `internal/compiler/check/check.go` `borrowConflictDiagnostic` (~2793-2822).
**Apply to:** the new `check.interprocedural_loan_liveness` diagnostic (D-08-20..25). Never a repair when the fix lives in a different function than `Primary` (`ownership.borrow_requires_share` / `ownership.transfer_requires_take`'s classification-only precedent, not `borrow_conflict`'s repair-bearing shape).

### Deterministic work counters — never wall-clock for the hard gate
**Source:** `internal/compiler/check/check.go:1142-1225` (`loanLivenessFixpoint`'s existing `work` counter, `work++` per transfer-function evaluation and per worklist reinsertion).
**Apply to:** every new counted quantity this phase adds (summary-derivation work, cost-corpus work) — thread a plain `int` counter alongside the result, record wall-clock separately and mark it observed-only in the manifest.

### Two independent chokepoints must both widen together, or a "hard" gate silently cannot block
**Source:** `internal/compiler/measure/statistics.go:131-135` (`Demote`) + `internal/compiler/session/session_phase6_budget.go:55-57` (`QLT02GateEligibleMetrics`).
**Apply to:** any plan task introducing a new gate-eligible metric name — both files must change in the same wave, with `TestDemoteHasExactlyOnePromotionPassthrough` re-derived (not just re-run).

### Table-driven test with a work-ratio or growth-exponent assertion across an input-size series
**Source:** `internal/compiler/check/check_test.go:737-753` (`TestReborrowChainWorkIsLinear`).
**Apply to:** `TestSummaryDerivationRequiresProgramOrder`, the cost-gate's `TestQLT02InterproceduralGrowthExponent`.

## No Analog Found

None. Every file/function this phase touches has a same-file, same-role, near-identical-shape existing instance to copy from — this is the expected shape of a phase whose own research summary states "every mechanism this phase needs already exists in the tree in miniature."

## Metadata

**Analog search scope:** `internal/compiler/check/`, `internal/compiler/measure/`, `internal/compiler/session/`, `internal/compiler/callgraph/`, `internal/compiler/core/`, `internal/compiler/diagnostic/`, `testdata/phase3/`, `testdata/phase07/`, `.planning/spikes/006-interprocedural-liveness-cost-scaling/`.
**Files scanned:** 14 (all confirmed git-tracked via `git ls-files`).
**Pattern extraction date:** 2026-09-09
```
