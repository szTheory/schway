# Phase 13: Agent Loop for Interprocedural Defects - Pattern Map

**Mapped:** 2026-09-13
**Files analyzed:** 10 (4 modified, 6 created/extended-only)
**Analogs found:** 10 / 10 (all analogs are same-file precedent or sibling-file precedent already in the codebase)

All analog paths below were verified with `git ls-files` — every path is
tracked source, not a gitignored mirror.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/compiler/check/check.go` (blame resolver, B1/B2/B3) | service (compiler admission pass) | transform (program → diagnostics) | Same file: `buildInterproceduralSummaries` (:670-722) for the callee-before-caller resolver-data-structure pattern | exact — same-file precedent, same program shape |
| `internal/compiler/check/check.go` (3 repair emission sites) | service (diagnostic emitter) | transform | Same file: `borrowConflictDiagnosticPostAssembly` (:1223-1263) and `interproceduralLoanLivenessDiagnostic` (:1265-1289) | exact — identical emission-site shape, different code/cause payload |
| `internal/compiler/check/check_ordering_stability_test.go` | test (golden/pinned-ID table) | CRUD-adjacent (assert-only) | Same file: existing pinned-table rows (see excerpt below) | exact — re-pin in place, no structural change |
| `internal/compiler/session/session_phase6_explain.go` | service (session-layer synthesis) | transform (diagnostic → cause DAG) | Same file: `buildExplainGraph` (:139-235), `explainSpanStrictlyContains` (:112-123) | exact — same-file precedent for the new guard + node fields |
| `internal/compiler/protocol/protocol.go` | model/config (wire schema struct) | request-response (JSON wire) | `internal/compiler/core/core.go:127-150` `core.Function.ForeignContract` additive-omitempty field (D-04-23) | exact — documented in-repo precedent for the exact discipline needed |
| New interprocedural injectors (`session_phase6_injectors.go` or phase-13 sibling) | service (fault-injection harness) | event-driven (marker-triggered mutation) | Same file: `MatchInjector`/`MoveInjector` (:94-140+), `markerGuard` (:55-65), `matchInjectSkippingGuard` (:117-131) | exact — same `Injector` interface, same guard contract |
| `testdata/phase13/` + `HELDOUT.sha256` | config/fixture (test data + manifest) | file-I/O | `testdata/phase6/` directory layout + `cmd/lang/main.go`'s `isPhase6Corpus` (:251-254) | role-match — layout precedent; manifest itself (`HELDOUT.sha256`) is net-new, no direct analog, modeled on the sealed-corpus intent described in D-13-27 |
| Topology-distinctness / sealed-digest control tests | test (structural/mutation-kill control) | CRUD-adjacent (assert-only) | `internal/compiler/session/session_phase6_injectors_test.go` `TestPhase6DefectCorpusIsHeldOut` (byte-inequality control to be *strengthened*, not copied verbatim) | role-match — same control *shape*, stronger predicate per D-13-26 |
| `cmd/lang-repair/repair_test.go`, `antitheater_test.go` | test (integration, driver-level) | request-response (subprocess JSON protocol) | Same files: `TestUnrepairableDefectFailsTheGate`, `TestRepairDriverFixesEveryDefectClassSinglePass`, `TestRepairDriverSourceNeverReferencesHeldoutFixtures` (all pre-existing, extend only) | exact — extend existing tests, no new file needed |
| `cmd/lang-repair/import_boundary_test.go` (respect, not modify) | test (structural lint) | — | Same file: `scanForbiddenImports` (:25-48), `forbiddenImportSubstr = "/internal/"` (:18) | exact — new tests must satisfy this unchanged lint |

## Pattern Assignments

### `internal/compiler/check/check.go` — blame resolver (D-13-01..08)

**Analog:** same file, `buildInterproceduralSummaries` (:670-722)

**Callee-before-caller ordering pattern** (check.go:698-704) — the exact
technique D-13-03's B3 tie-break must call a second time (not thread out of
`buildInterproceduralSummaries`, which does not export its local `order`
variable — see RESEARCH.md Code Example 4):
```go
order, err := callgraph.Order(program)
if err != nil {
    return result, 0
}
// ...
for i := len(order) - 1; i >= 0; i-- {   // callee-before-caller walk
    functionID := order[i]
    // ...
}
```
Budget this as a small shared helper (e.g. `calleeBeforeCallerOrder(program) ([]string, error)`)
called by both `buildInterproceduralSummaries` and the new blame resolver —
one call site for the technique, two callers, satisfying "no new ordering
authority is introduced."

**Resolver data structure pattern (D-13-02)** — invert
`program.Functions[i].Linear.Operations` into a lookup map, exactly as
`spanByOperationID` is already threaded through `checkInterproceduralLoanLiveness`.
No new plumbing; compose with the existing map the same way.

**In-scope place enumeration (D-13-10's uniqueness gate)** — already
available on `resolveCallBinding`'s own signature, no plumbing needed:
```go
// check.go:2958 [VERIFIED]
func resolveCallBinding(functionID string, opOrdinal int, binding ast.Binding,
    places map[string]*placeState, calleeContracts map[string]calleeContract,
    typeFact core.TypeFact, foreignSymbols map[string]foreignSymbolInfo,
) (core.LinearOperation, core.Place, *diagnostic.Diagnostic)
```
```go
// check.go:3847-3862 [VERIFIED]
type placeState struct {
    place          core.Place
    declared       diagnostic.Span
    initialized    bool
    movedAt        *diagnostic.Span
    moveTargetID   string
    moveTargetName string
}
```
Iterate `places`, keep `initialized == true`, match each `place.place.TypeID`'s
constructor against `contract.ParameterType`.

**Typed refusal pattern for `blame_undetermined` (D-13-07)** — model on the
project-wide `{Code}`-plus-wrapped-`Err` shape:
```go
// internal/compiler/session/session_phase6_injectors.go:30-41 [VERIFIED]
type InjectorError struct {
    Code string
    Err  error
}
func (e *InjectorError) Error() string { return e.Code + ": " + e.Err.Error() }
func (e *InjectorError) Unwrap() error { return e.Err }
```
`ExplainError` in `session_phase6_explain.go:15-20` is the `{Code}`-only
sibling shape (no wrapped `Err`) if `blame_undetermined` needs only a code,
not a wrapped cause.

**Whole-function span, already free (D-13-15's peer input)** — do not
byte-search; both peers already carry a `Span` field:
```go
// core.go:127-150 [VERIFIED]
type Function struct {
    // ...
    Span diagnostic.Span `json:"span"`
}
```
```go
// ast.go:69-76 [VERIFIED]
type FuncDecl struct {
    Name         string
    Parameter    Parameter
    ReturnOrigin *BorrowOrigin
    ReturnType   TypeRef
    Body         Body
    Span         diagnostic.Span
}
```
`core.Function.Span` is already consumed today by `verifyCallableRefusal`'s
`core.callee_not_callable` Primary (whole-function granularity, not a
call-site token):
```go
// check.go:1546-1580 [VERIFIED]
if !callable {
    diag := diagnostic.Error(
        core.CalleeNotCallable, function.Span, "call target is not callable",
        diagnostic.Cause{Kind: "callee", Detail: operation.CalleeID},
    )
    return &diag
}
```

**`core.call_graph_cycle` exemption (D-13-05)** — Primary stays at the
closing operation, blame set published as secondary causes, exempt from B3:
```go
// check.go:541-560 [VERIFIED]
func checkCallGraphAcyclic(program core.Program, spanByOperationID map[string]diagnostic.Span) *diagnostic.Diagnostic {
    if _, err := callgraph.Order(program); err != nil {
        if cycle, ok := callgraph.CycleError(err); ok {
            members := cycle.Members()
            edgeOperationIDs := cycle.MemberEdgeOperationIDs()
            causes := []diagnostic.Cause{{Kind: "cycle_length", Detail: fmt.Sprintf("%d", len(members))}}
            // ... one cycle_member cause per edge, span from spanByOperationID ...
            primary := spanByOperationID[cycle.ClosingOperationID()]
            diag := diagnostic.Error(core.CallGraphCycle, primary, "call graph contains a cycle; recursion is refused", causes...)
            return &diag
        }
    }
}
```

---

### `internal/compiler/check/check.go` — three repair emission sites (D-13-09)

**Analog:** same file, `borrowConflictDiagnosticPostAssembly` (:1223-1263) —
the exact model for class 1 (`move_after_interprocedural_loan`), including
the `:stmt`-suffixed span-channel convention and `core.Place.Name`-based
text reconstruction:
```go
// check.go:1223-1263 [VERIFIED]
func borrowConflictDiagnosticPostAssembly(newBorrow, blockingBorrow core.LinearOperation, blockingLastUseSpan diagnostic.Span, places map[string]core.Place, spanByOperationID map[string]diagnostic.Span) diagnostic.Diagnostic {
    causes := []diagnostic.Cause{
        {Kind: "borrow_created_here", Span: spanPointer(spanByOperationID[blockingBorrow.ID])},
        {Kind: "borrow_used_later", Span: spanPointer(blockingLastUseSpan)},
        {Kind: "loan", Detail: blockingBorrow.LoanID},
        {Kind: "owner", Detail: newBorrow.SourceID},
        {Kind: "type", Detail: newBorrow.TypeID},
    }
    repairs := []diagnostic.Repair{{Kind: "create_loan_after_conflicting_loan_ends"}}
    if newBorrow.Kind == core.OpBorrowExclusive {
        // Repair.Span must cover the WHOLE statement (Replacement rewrites
        // all of it) -- the ":stmt"-suffixed entry, NOT the bare RHS-token
        // span used for the diagnostic's own Primary/causes.
        newBorrowSpan, ok := spanByOperationID[newBorrow.ID+":stmt"]
        if !ok {
            newBorrowSpan = spanByOperationID[newBorrow.ID]
        }
        targetName, sourceName := "", ""
        if places != nil {
            targetName = places[newBorrow.TargetID].Name
            sourceName = places[newBorrow.SourceID].Name
        }
        repairs = append(repairs, diagnostic.Repair{
            Kind: "narrow_to_shared_borrow", Span: &newBorrowSpan,
            Replacement:   "let " + targetName + " = borrow " + sourceName,
            Applicability: diagnostic.ApplicabilityMachineApplicable,
        })
    }
    return diagnostic.ErrorWithRepairs(
        "ownership.borrow_conflict", spanByOperationID[newBorrow.ID], "...", causes,
        repairs...,
    )
}
```
D-13-09.1's reorder replacement is the same technique with `CalleeID` added
to reconstruct both swapped statements.

**Current (pre-repair) shape of the three target codes** — these are the
exact call sites to convert from `diagnostic.Error` to
`diagnostic.ErrorWithRepairs`:

1. `check.interprocedural_loan_liveness` (check.go:1276-1289):
```go
func interproceduralLoanLivenessDiagnostic(move, borrow, call core.LinearOperation, causeDetail string, spanByOperationID map[string]diagnostic.Span) diagnostic.Diagnostic {
    borrowSpan := spanByOperationID[borrow.ID]
    callSpan := spanByOperationID[call.ID]
    causes := []diagnostic.Cause{
        {Kind: "borrow_created_here", Span: &borrowSpan},
        {Kind: "loan_extended_by_call", Span: &callSpan, Detail: call.CalleeID},
        {Kind: "callee_return_contract", Detail: causeDetail},
    }
    primary := spanByOperationID[move.ID]
    return diagnostic.Error(
        "check.interprocedural_loan_liveness", primary,
        "cannot transfer ownership while an interprocedurally-extended loan is still live", causes...,
    )
}
```

2. `syntax.fallible_call_not_consumed` (check.go:3055-3057, inside `resolveCallBinding`):
```go
if _, isForeign := foreignSymbols[binding.RHS.Callee]; isForeign {
    diag := diagnostic.Error("syntax.fallible_call_not_consumed", binding.RHS.Span, "a fallible call must be the operand of `try`")
    return core.LinearOperation{}, core.Place{}, &diag
}
```
Replacement: `"try " + Callee + "(" + Arguments[0] + ")"`, Span:
`binding.RHS.Span`.

3. `check.call_argument_type_mismatch` (check.go:2993-3000, inside `resolveCallBinding`):
```go
if !callArgumentTypeCheckSeam && (typeFact.Shape.Constructor == "" || contract.ParameterType == "" || typeFact.Shape.Constructor != contract.ParameterType) {
    causes := []diagnostic.Cause{
        {Kind: "callee", Detail: contract.ID},
        {Kind: "argument_type", Detail: typeFact.Shape.Constructor},
        {Kind: "declared_parameter_type", Detail: contract.ParameterType},
    }
    diag := diagnostic.Error(checkCallArgumentTypeMismatch, binding.RHS.Span, "call argument type does not match the callee's declared parameter type", causes...)
    return core.LinearOperation{}, core.Place{}, &diag
}
```
D-13-10's uniqueness gate applies here: emit `use_matching_argument` repair
only when exactly one initialized in-scope place's `TypeID` constructor
matches `contract.ParameterType`; otherwise keep `diagnostic.Error` (no
repair) so the driver returns `unrepairable`.

**`Error` vs `ErrorWithRepairs` identity construction (the D-13-09a churn
mechanism, must be understood before touching any of the three sites):**
```go
// internal/compiler/diagnostic/diagnostic.go:108-152 [VERIFIED]
func Error(code string, span Span, message string, causes ...Cause) Diagnostic {
    identity := struct {
        Schema string
        Code   string
        Span   Span
        Causes []Cause
    }{Schema: Schema /* "lang.diagnostic/0" */, Code: code, Span: span, Causes: causes}
    encoded, _ := json.Marshal(identity)
    sum := sha256.Sum256(encoded)
    return Diagnostic{Schema: Schema, ID: "diagnostic:" + hex.EncodeToString(sum[:12]), ...}
}

func ErrorWithRepairs(code string, span Span, message string, causes []Cause, repairs ...Repair) Diagnostic {
    // repairs sorted by Kind then Detail
    repairKinds := make([]string, len(repairs)) // ONLY Kind, never Span/Replacement/Applicability
    identity := struct {
        Schema      string
        Code        string
        Span        Span
        Causes      []Cause
        RepairKinds []string
    }{Schema: Schema1 /* "lang.diagnostic/1" */, Code: code, Span: span, Causes: causes, RepairKinds: repairKinds}
    // ...
}
```
The `Schema` string alone differs between the two hashed structs — switching
any of the three sites to `ErrorWithRepairs` churns that code's ID
unconditionally, even on an instance where the repair never fires.

**Applicability / driver-eligibility gate (all three sites and D-13-08's
`blame_undetermined` dual-repair case must go through this unchanged):**
```go
// diagnostic.go:71-95 [VERIFIED]
func NormalizeApplicability(applicability string) string {
    if applicability == "" {
        return ApplicabilityUnspecified
    }
    return applicability   // never defaults toward MachineApplicable
}

func DriverEligible(r Repair) bool {
    return r.Applicability == ApplicabilityMachineApplicable && r.Kind != "" && r.Span != nil && r.Replacement != ""
}
```

---

### `internal/compiler/check/check_ordering_stability_test.go` — six re-pinned rows (D-13-09a)

**Analog:** same file's existing pinned table (verified exact current values):
```go
"phase07/call_type_mismatch.lang":          {"check.call_argument_type_mismatch", "diagnostic:eec74c3869e957e7eccaafb8"},
"phase07/relay_escort_witness.lang":        {"check.interprocedural_loan_liveness", "diagnostic:58c1b5b2072cda60f77d721e"},
"phase08/relay_depth2_refuse.lang":         {"check.interprocedural_loan_liveness", "diagnostic:300248748c05f20eb1fe3948"},
"phase08/twin_a_refuse.lang":               {"check.interprocedural_loan_liveness", "diagnostic:e01ad6316e27899deeb610f7"},
"phase08/twin_b_refuse.lang":               {"check.interprocedural_loan_liveness", "diagnostic:d15b65a04fd8515f92e999e2"},
"phase4/fallible_call_unconsumed.lang":     {"syntax.fallible_call_not_consumed", "diagnostic:91c8b8c7a5d359d9a14fe6a8"},
```
These six rows (spanning all three D-13-09 target codes) WILL churn once
`ErrorWithRepairs` lands on those codes. Rows for `core.callee_not_callable`
and `core.call_graph_cycle` are untouched. Structure: same table shape, same
test — only regenerate and re-pin the hash values, do not restructure the
test.

---

### `internal/compiler/session/session_phase6_explain.go` — function attribution + narrows guard (D-13-14, D-13-19, D-13-22)

**Analog:** same file, `buildExplainGraph` (:139-235) and
`explainSpanStrictlyContains` (:112-123).

**Current containment check to guard (D-13-19's sleeper bug fix site):**
```go
// session_phase6_explain.go:112-123 [VERIFIED]
func explainSpanStrictlyContains(parent, child *diagnostic.Span) bool {
    if parent == nil || child == nil {
        return false
    }
    if parent.Start > child.Start || parent.End < child.End {
        return false
    }
    return parent.Start != child.Start || parent.End != child.End
}
```
Guard to add: a node whose span equals a `core.Function.Span` is a
function-scope node and may be a `caused_by` parent but never a `narrows`
parent; `narrows` additionally requires parent and child to resolve to the
same `function_id`. Wire this into the `narrows`-selection block:
```go
// session_phase6_explain.go:183-198 [VERIFIED — the exact block to extend]
if edgeKind == protocol.EdgeCausedBy && cause.Span != nil {
    best := -1
    for candidate := range items {
        parentSpan := items[candidate].node.Span
        if !explainSpanStrictlyContains(parentSpan, cause.Span) {
            continue
        }
        if best == -1 || explainSpanWidth(parentSpan) < explainSpanWidth(items[best].node.Span) {
            best = candidate
        }
    }
    if best != -1 {
        parent = best
        edgeKind = protocol.EdgeNarrows
    }
}
```

**Correlation-key precedent to avoid re-breaking (D-13-16's basis — do NOT
put function identity in `Detail`):**
```go
// session_phase6_explain.go:94-110 [VERIFIED]
var explainCorrelationKinds = map[string]bool{
    "place": true, "owner": true, "loan": true, "transfer_target": true,
}
func explainCorrelationKey(kind, detail string) string {
    if detail == "" || !explainCorrelationKinds[kind] {
        return ""
    }
    return kind + ":" + detail
}
```

**Node/edge synthesis entry points to extend with `FunctionID`/`FunctionName`:**
```go
// session_phase6_explain.go:153-221 [VERIFIED — root + per-cause node construction]
items := []explainBuilt{{
    node: protocol.ExplainNode{
        ID: diag.ID, Kind: diag.Code, Detail: diag.Message, Span: &rootSpan,
        Availability: string(debugmap.Available),
    },
    depth: 0,
}}
// ... per cause:
items = append(items, explainBuilt{
    node:  protocol.ExplainNode{ID: causeID, Kind: cause.Kind, Detail: cause.Detail, Span: cause.Span, Availability: availability},
    depth: depth,
})
```
Both construction sites need the peer-re-derived `function_id`/`function_name`
lookup (once from `core.Program.Functions[].Span`, once from
`ast.Program.Funcs[].Span`, per D-13-15) applied by innermost-containing-span
resolution — reuse `explainSpanStrictlyContains`-style bounds checks, not a
new search primitive.

**`ExplainCommandFile`'s existing local-variable capture (where the
peer-re-derived function tables get threaded in):**
```go
// session_phase6_explain.go:45-74 [VERIFIED]
func ExplainCommandFile(path, diagnosticID string, depth int) (protocol.Result, error) {
    // ...
    diagnostics := parsed.Diagnostics
    work := 1
    if len(diagnostics) == 0 {
        checked := check.Program(parsed.Program)
        diagnostics = checked.Diagnostics
        work = checked.Work
    }
    target, found := findExplainDiagnostic(diagnostics, diagnosticID)
    // ...
    nodes, edges, truncated, graphWork := buildExplainGraph(target, resolveExplainDepth(depth))
    result.Explain = &protocol.ExplainSummary{
        Schema: protocol.ExplainSchema, RootID: target.ID, Nodes: nodes, Edges: edges, Truncated: truncated,
    }
    return completeCommand(result, started, work+graphWork), nil
}
```
`checked.Program` (the `core.Program` with `Functions[].Span`) and
`parsed.Program` (the `ast.Program` with `Funcs[].Span`) are both already
local values here — pass both into `buildExplainGraph` (or a new sibling
helper) for the peer-re-derived resolution, no new state needs to be
plumbed from elsewhere.

---

### `internal/compiler/protocol/protocol.go` — additive struct changes (D-13-14, D-13-18)

**Analog:** `internal/compiler/core/core.go:127-150`
`core.Function.ForeignContract`, the documented in-repo precedent for
additive-omitempty leaving serialized bytes unchanged (D-04-23):
```go
// core.go:143-149 [VERIFIED]
// ForeignContract is the Phase 4 D-04-12 fact (additive, omitempty):
// present only when this function calls a declared `foreign C {}` symbol.
// Every pre-Phase-4 function, and every Phase 4 function that calls no
// foreign symbol, leaves this nil, so its serialized bytes are unchanged
// (D-04-23).
ForeignContract *ForeignContract `json:"foreign_contract,omitempty"`
```

**Current `ExplainNode`/`ExplainSummary` shape to extend:**
```go
// protocol.go:218-251 [VERIFIED]
type ExplainNode struct {
    ID           string           `json:"id"`
    Kind         string           `json:"kind"`
    Detail       string           `json:"detail,omitempty"`
    Span         *diagnostic.Span `json:"span,omitempty"`
    Availability string           `json:"availability"`
}
// add: FunctionID string `json:"function_id,omitempty"`
//      FunctionName string `json:"function_name,omitempty"`

type ExplainSummary struct {
    Schema    string        `json:"schema"`
    RootID    string        `json:"root_id"`
    Nodes     []ExplainNode `json:"nodes"`
    Edges     []ExplainEdge `json:"edges"`
    Truncated string        `json:"truncated,omitempty"`
}
// add: Functions []ExplainFunction `json:"functions,omitempty"`
// new type ExplainFunction struct { ID, Name string; Span diagnostic.Span }
```
`ExplainSchema = "lang.explain/0"` (protocol.go:198) does not bump per
D-13-18 — same reasoning as `ForeignContract`, `ExplainSummary` carries no
identity hash.

---

### New interprocedural defect injectors (D-13-24)

**Analog:** same file, `MatchInjector` + `markerGuard` +
`matchInjectSkippingGuard` (session_phase6_injectors.go:43-131):
```go
// :43-53 [VERIFIED] — the interface every new injector implements
type Injector interface {
    Name() string
    Inject(source []byte) ([]byte, error)
}

// :55-65 [VERIFIED] — the fail-closed guard contract
func markerGuard(count int, want string) error {
    if count == 0 {
        return &InjectorError{Code: InjectorTargetMissingCode, Err: fmt.Errorf("%s marker count is %d, want at least 1", want, count)}
    }
    return nil
}

// :94-115 [VERIFIED] — full injector shape to clone per new interprocedural class
const matchTargetMarker = "// lang:match-target"
type MatchInjector struct{}
func (MatchInjector) Name() string { return "match" }
func (MatchInjector) Inject(source []byte) ([]byte, error) {
    lines, index := lastMarkerLine(source, matchTargetMarker)
    if err := markerGuard(markerCount(lines, matchTargetMarker), "match arm target"); err != nil {
        return nil, err
    }
    mutated := append(append([]string(nil), lines[:index]...), lines[index+1:]...)
    return []byte(strings.Join(mutated, "\n")), nil
}

// :117-131 [VERIFIED] — the not-inert twin every new injector needs
// (guard-disabled variant demonstrating the silent-pass failure the guard
// prevents; never called from the real Inject path)
func matchInjectSkippingGuard(source []byte) []byte {
    lines, index := lastMarkerLine(source, matchTargetMarker)
    if index == -1 {
        return append([]byte(nil), source...)
    }
    mutated := append(append([]string(nil), lines[:index]...), lines[index+1:]...)
    return []byte(strings.Join(mutated, "\n"))
}
```
Typed refusal shape to reuse verbatim:
```go
// :24-41 [VERIFIED]
const InjectorTargetMissingCode = "phase6.injector_target_missing"
type InjectorError struct {
    Code string
    Err  error
}
func (e *InjectorError) Error() string { return e.Code + ": " + e.Err.Error() }
func (e *InjectorError) Unwrap() error { return e.Err }
```
Each new interprocedural injector needs its own marker constant (e.g.
`// lang:interprocedural-blame-target`), its own `Inject`, and its own
`*InjectSkippingGuard` twin extending `TestEveryInjectorRefusesWhenMarkerDisappears`
and `TestInjectorTargetChoiceIsSpecified` (D-13-30a).

---

### `testdata/phase13/` fixture corpus + `HELDOUT.sha256` manifest (D-13-24, D-13-27)

**Analog:** `testdata/phase6/` directory layout (verified via `git ls-files`):
```
testdata/phase6/README
testdata/phase6/derivation_borrow_defect.lang
testdata/phase6/derivation_match_defect.lang
testdata/phase6/derivation_move_defect.lang
testdata/phase6/heldout_borrow_defect.lang
testdata/phase6/heldout_match_defect.lang
testdata/phase6/heldout_move_defect.lang
testdata/phase6/stale_evidence_subject.lang
```
`heldout_`/`derivation_` prefix convention and README contract carry
forward unchanged into `testdata/phase13/`.

**Marker-file corpus dispatch pattern (cmd/lang/main.go:229-254), the
phase-13 sibling this needs:**
```go
// cmd/lang/main.go:244-254 [VERIFIED]
func isPhase6Corpus(corpus string) bool {
    _, err := os.Stat(filepath.Join(corpus, "heldout_match_defect.lang"))
    return err == nil
}
```
`isPhase7Corpus` is checked BEFORE `isPhase6Corpus` in the dispatch chain
(main.go:165-172) — a `isPhase13Corpus` sibling should follow the same
ordering discipline (checked before the corpus it could otherwise be
swallowed by).

`HELDOUT.sha256` itself has no direct in-repo analog (net-new artifact);
model its accompanying "assert current bytes match" test on the shape of
`TestPhase6DefectCorpusIsHeldOut`'s byte-comparison test in
`session_phase6_injectors_test.go` (assert-only, no goldens/bless button per
D-13-27).

---

### `cmd/lang-repair/repair_test.go`, `antitheater_test.go` — extend only (D-13-32, tests-only)

**Analog:** same files' existing named tests (`TestUnrepairableDefectFailsTheGate`,
`TestRepairDriverFixesEveryDefectClassSinglePass`,
`TestRepairDriverSourceNeverReferencesHeldoutFixtures`) — extend each with
subtests/scope for the three new interprocedural classes rather than adding
parallel test functions, per D-13-29's "extend `TestUnrepairableDefectFailsTheGate`"
and D-13-27's "extend `TestRepairDriverSourceNeverReferencesHeldoutFixtures`
beyond `cmd/lang-repair` to any non-test file in `internal/compiler/check`."

**Structural import-boundary lint every new test must respect (no source
change, re-declare rather than import):**
```go
// cmd/lang-repair/import_boundary_test.go:15-18 [VERIFIED]
const forbiddenImportSubstr = "/internal/"
```
Any helper a new test needs that lives under `internal/` (e.g. a blame-rule
constant) must be re-declared locally in the test file, not imported —
`scanForbiddenImports` (import_boundary_test.go:25-48) parses every
non-test `.go` file under `cmd/lang-repair` and fails the build on any
`internal/`-prefixed import.

## Shared Patterns

### Typed `{Code}`-plus-wrapped-`Err` failure shape
**Sources:** `internal/compiler/session/session_phase6_injectors.go:30-41` (`InjectorError`, wraps `Err`), `internal/compiler/session/session_phase6_explain.go:15-20` (`ExplainError`, `{Code}`-only).
**Apply to:** any new `blame_undetermined` refusal type in `check.go`. Use
the `{Code}`-only shape (`ExplainError`) if no underlying Go error needs
wrapping; use the wrapped shape (`InjectorError`) if one does.
```go
type ExplainError struct{ Code string }
func (e *ExplainError) Error() string { return e.Code }
```

### Fail-closed defaults
**Sources:** `diagnostic.go:71-95` (`NormalizeApplicability`/`DriverEligible`), `session_phase6_injectors.go:55-65` (`markerGuard`).
**Apply to:** every new repair-emission gate (D-13-07, D-13-10) and every
new injector. Never default toward `MachineApplicable`; refuse (return an
error / emit no repair) rather than emit a plausible-but-wrong edit.

### Diagnostic identity construction (`Error` vs `ErrorWithRepairs`)
**Source:** `internal/compiler/diagnostic/diagnostic.go:108-152`.
**Apply to:** all three D-13-09 repair emission sites — switching from
`Error` to `ErrorWithRepairs` unconditionally changes the code's ID (Schema
string alone differs in the hashed struct); plan a re-pin task for
`check_ordering_stability_test.go`'s six affected rows in the same wave.

### Anti-theater / not-inert twin tests
**Sources:** `session_phase6_injectors.go:117-131` (`matchInjectSkippingGuard`), `cmd/lang-repair/antitheater_test.go` (not read in full this pass — budget a dedicated read before writing new controls, per RESEARCH.md Open Question 2).
**Apply to:** every new injector (D-13-30a), the topology-distinctness
control (D-13-30b — kill with an alpha-renamed copy of a derivation
fixture, asserting red), the sealed-digest control (D-13-30c — kill with a
one-byte flip in a temp copy), and the criterion-3 control (D-13-30d — kill
by applying the repair at the detection-site function instead of the
compiler-named one, asserting the program does NOT re-check clean).

### Peer re-derivation (`check` vs `corevalidate`; core-table vs AST)
**Source:** the general pattern already established across the codebase (D-08-17/D-09-31; distinct `checkCallArgumentTypeMismatch`/`core.CallArgumentTypeMismatch` constants with explicit "never unified" doc comments at check.go:1291-1303).
**Apply to:** D-13-06 (blame computed in `check` only, `corevalidate` never
re-derives it) and D-13-15 (function resolution independently derived once
from `core.Program.Functions[].Span`, once from `ast.Program.Funcs[].Span`).

## No Analog Found

None — every file in scope has at least a role-match analog already in the
tree. The only genuinely net-new artifact is `testdata/phase13/HELDOUT.sha256`
itself (a sealed-manifest file, not code), which has no direct precedent but
is fully specified by D-13-27's text (sha256 manifest + a test asserting
current bytes match, no bless button).

## Metadata

**Analog search scope:** `internal/compiler/check/`, `internal/compiler/session/`, `internal/compiler/protocol/`, `internal/compiler/diagnostic/`, `internal/compiler/core/`, `internal/compiler/ast/`, `cmd/lang-repair/`, `cmd/lang/`, `testdata/phase6/`
**Files scanned:** 12 read directly this session (line ranges non-overlapping with RESEARCH.md's prior reads where possible; all analog paths confirmed tracked via `git ls-files`)
**Pattern extraction date:** 2026-09-13
