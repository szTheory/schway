# Phase 13: Agent Loop for Interprocedural Defects - Research

**Researched:** 2026-09-13
**Domain:** Diagnostic blame attribution, `lang explain` cause-DAG function attribution, and JSON-protocol repair emission for cross-function defects — pure codebase verification, no external research (per scope).
**Confidence:** HIGH (every claim below is either directly read from source this session, or explicitly marked ASSUMED)

## Summary

This phase has no framework or library research surface — it is entirely a
codebase-grounded verification of 33 already-locked decisions in
`13-CONTEXT.md`. This document exists to (a) empirically confirm or refute
CONTEXT.md's central factual claims against the shipped source, and (b) hand
the planner exact file:line coordinates, existing test names, and one
significant correction CONTEXT.md did not make explicit.

**The central finding:** D-13-04's "all seven codes land in B2, zero Primary
spans move" claim is **VERIFIED** — all eight interprocedural codes'
`diagnostic.Error`/`diagnostic.Error` calls were read directly and every one
places `Primary` on the calling side (see Section 1 table). The pinned IDs in
`check_ordering_stability_test.go` are exactly as CONTEXT.md quotes them.

**The correction CONTEXT.md does not make explicit:** three of D-13-04's
"zero churn" codes are the *exact same* three codes D-13-09 is about to give
their first-ever `Repair`. Every one of them currently builds its diagnostic
with `diagnostic.Error` (schema `lang.diagnostic/0`, identity =
`{Schema, Code, Span, Causes}`). Attaching a `Repair` requires switching to
`diagnostic.ErrorWithRepairs` (schema `lang.diagnostic/1`, identity adds
`RepairKinds`) — and that switch **always** changes the sha256 identity, even
on fixtures where the new repair never fires (Schema alone differs in the
struct that gets hashed). D-13-04's zero-churn claim is true for *blame
adoption*; it is **not** true for *repair addition*, and D-13-09's own
"Reversibility: costly" note already half-says this for removal — the
forward direction needs to be equally explicit in the plan. Four rows in
`check_ordering_stability_test.go`'s pinned table **will** change their
`diagnostic:<hash>` value once D-13-09's three repairs land, and the plan
must include a task to re-pin them (not silently let the test start failing
mid-implementation).

**Primary recommendation:** build the blame resolver and the three repair
emission sites as one integrated `check.go` workstream (they share
`spanByOperationID`/`places`/`calleeContract` plumbing already local to
`Program()`), build the explain-side function attribution as a second,
largely-independent workstream (it touches only
`session_phase6_explain.go` + `protocol.go` + `ExplainCommandFile`'s local
variable capture), and treat the held-out corpus + `testdata/phase6`
retro-strengthening as a third, sequenced-last workstream since its D-13-26
topology triple depends on the call-graph shapes the first workstream's
fixtures will need to exist anyway.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Blame resolution (which function is Primary) | `check` (compiler / admission layer) | — | D-13-06: single-peer by design; `corevalidate` never re-derives blame |
| Repair emission (Span/Replacement/Applicability) | `check` (compiler / admission layer) | — | `cmd/lang-repair` is kind-agnostic (D-13-32); all repair *content* is a `check.go` emission question |
| Function attribution on explain nodes | `session` (agent-surface layer, `session_phase6_explain.go`) | `protocol` (wire schema) | Explain is a session-layer synthesis over `check`'s existing flat `Causes`; it is not itself an admission-layer concern |
| Repair application (splicing bytes) | `cmd/lang-repair` (driver tier) | — | Confirmed unchanged this phase (D-13-32); walled off from `internal/` by structural lint |
| Held-out fixture corpus + injectors | `session` (`session_phase6_injectors.go` sibling) | `internal/compiler/check` (topology-distinctness control) | Injectors are session-layer per D-13-24's precedent; the anti-overfitting topology control must live in `check`'s own path per D-13-27's "any non-test file in the blame-rule path" extension |

## User Constraints

<user_constraints>
### Locked Decisions

All 33 decisions D-13-01 through D-13-33 in `13-CONTEXT.md` are LOCKED. This
research does not reopen any of them. See `13-CONTEXT.md` in full for exact
text; the load-bearing ones are cited by ID throughout this document.

### Claude's Discretion (from CONTEXT.md, copied verbatim)

- Field naming granularity for D-13-14 (`function_id` + `function_name` vs a
  nested `function` object vs SARIF's `fully_qualified_name`).
- Whether the `functions` table on `ExplainSummary` ships now or waits for a
  consumer needing whole-function spans.
- Whether `functionByOperationID` is materialized once on `check.Result` or
  rebuilt per diagnostic.
- Whether `ExplainDefaultDepth` stays 3 for cross-function chains.
- Whether class 3's uniqueness gate counts all initialized in-scope places or
  only the function parameter plus prior `let`s bound before the call site.
- New corpus directory (`testdata/phase13/`) vs extending `testdata/phase6/`.
- Whether the topology triple is computed from the existing call-graph
  package or re-derived independently.
- Number of held-out fixtures per class (minimum: the D-13-28 twin pair).
- Whether a non-`Callable` callee is a callee contract violation (B1) or
  caller misuse (B2) — lean B2.

### Deferred Ideas (OUT OF SCOPE — copied verbatim)

- `check.go:3123`'s `declare_foreign_symbol` shell repair — left as-is, debt.
- `Cause.Span` identity-bearing / `Repair.Span` not — not reopened.
- Producer-side function identity on `diagnostic.Cause` — M003 candidate.
- Function attribution on raw `lang check` diagnostics (not `explain`).
- Bounded-exhaustive / grammar-based fixture generation.
- Foreign-C boundary causes with no Lang function — small, settle in-plan.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| DX-05 | `lang explain`'s cause DAG stays bounded across functions, names the owning function | Section 4 maps exact `ExplainNode`/`ExplainSummary` additions and every pinned test at risk |
| DX-06 | Cross-function blame attribution points at the correct fix location; repair-then-re-check regression for the non-obvious case | Sections 1-2 verify all 8 codes' current Primary sites and the resolver's available inputs; Section 8 covers D-13-28's twin-pair fixture |
| DX-07 | `lang-repair` reaches/fixes ≥3 new interprocedural classes via JSON protocol alone, on a held-out split | Section 3 maps all three emission sites concretely; Section 6 maps the injector/harness contract; Section 7 assesses D-13-33's blast radius |
</phase_requirements>

## Standard Stack

Not applicable — this phase adds zero new dependencies (Go stdlib only, per
the project's zero-external-production-dependency record, unaffected here).
No package legitimacy audit is required.

## Architecture Patterns

### System Architecture Diagram

```
lang --json check FILE
        |
        v
check.Program(ast.Program) --------------------------------------+
  |  builds core.Program.Functions[]                              |
  |  spanByOperationID (local, built during per-function walk)    |
  |  buildInterproceduralSummaries(program, table)                |
  |    -> callgraph.Order(program)   [caller-before-callee, RAW]  |
  |       walked BACKWARD internally for callee-before-caller     |
  |       (this reversed order is NOT returned/exported today)    |
  |  checkInterproceduralLoanLiveness(...)                        |
  |  [NEW] blame resolver: functionByOperationID (D-13-02)        |
  |    consumed by [NEW] B1/B2/B3 rule (D-13-03) at every one     |
  |    of the 8 interprocedural emission sites                    |
  |  [NEW] 3 repair-emission sites (D-13-09) attach Repair{}       |
  |    -> flips diagnostic.Error to diagnostic.ErrorWithRepairs   |
  |    -> CHURNS 4 pinned rows in check_ordering_stability_test.go|
  v
diagnostic.Diagnostic{Primary, Causes, Repairs}  (JSON wire)
  |                                                    |
  v (lang explain)                                     v (lang-repair)
session.ExplainCommandFile                     cmd/lang-repair.Repair()
  parses+checks SRC fresh (no cache)             runLangCheck (subprocess 1)
  buildExplainGraph(target, depth)               selectRepair (first eligible)
  [NEW] function attribution:                    applyRepair (splice bytes)
    - once from core.Program.Functions[].Span    runLangCheck (subprocess 2)
    - once from ast.Program.Funcs[].Span         -> Outcome{Repaired|...}
    (peer re-derivation, D-13-15)                 (D-13-32: already single-pass,
  [NEW] narrows function-scope guard (D-13-19)     no cmd/lang-repair change)
  v
protocol.ExplainSummary{Nodes: [...function_id/function_name], Functions: [...]}
```

### Recommended Project Structure (files this phase touches)

```
internal/compiler/check/check.go          # blame resolver + 3 repair sites (bulk of phase)
internal/compiler/check/check_ordering_stability_test.go  # 4 rows need re-pinning
internal/compiler/session/session_phase6_explain.go       # function attribution + narrows guard
internal/compiler/session/session_phase6_explain_test.go  # existing pinned-shape tests to re-run
internal/compiler/protocol/protocol.go                    # additive struct fields only
internal/compiler/session/session_phase6_injectors.go (or a phase-13 sibling)  # new interprocedural injectors
testdata/phase13/ (or testdata/phase6/ extension)          # new corpus + HELDOUT.sha256
cmd/lang/main.go                                            # isPhaseNCorpus sibling if new dir
cmd/lang-repair/repair_test.go, antitheater_test.go        # tests only, per D-13-32
```

### Pattern: peer-independent resolver inputs already exist, do not build a byte-search

`core.Function` already carries `Span diagnostic.Span` unconditionally
(`internal/compiler/core/core.go:149`, no `omitempty`) — **already consumed
today** by `verifyCallableRefusal` for `core.callee_not_callable`'s Primary
(`internal/compiler/check/check.go:1576-1579`, `diagnostic.Error(core.CalleeNotCallable, function.Span, ...)`)
and by two other whole-function-Primary diagnostics
(`core.invalid_body` at check.go:200, `type.return_mismatch` at check.go:256).
`ast.FuncDecl` independently carries its own `Span diagnostic.Span`
(`internal/compiler/ast/ast.go:69-76`). **D-13-15's peer re-derivation is
therefore two straightforward loops over two already-spanned structures —
not a byte-offset containment search that has to be invented.** This
narrows (favorably) CONTEXT.md's own framing in the `<domain>` section
("function identity is resolvable by mapping a byte offset to its enclosing
function declaration") — the mapping is a direct field read on both peers,
not a search.

```go
// Source: internal/compiler/core/core.go:127-150 [VERIFIED]
type Function struct {
    ID   string `json:"id"`
    Name string `json:"name"`
    ...
    Span diagnostic.Span `json:"span"`
}
```

```go
// Source: internal/compiler/ast/ast.go:69-76 [VERIFIED]
type FuncDecl struct {
    Name         string
    Parameter    Parameter
    ReturnOrigin *BorrowOrigin
    ReturnType   TypeRef
    Body         Body
    Span         diagnostic.Span
}
```

### Anti-Patterns to Avoid

- **Do not treat `buildInterproceduralSummaries`' internal reversed loop as
  an already-exported callee-before-caller order.** See Section 2 — it is a
  local variable. D-13-03's B3 reuses the same *authority*
  (`callgraph.Order`, reversed) but must call it again, not thread a value
  out of `buildInterproceduralSummaries`.
- **Do not assume `check.go:1550-1580`'s `core.callee_not_callable` Primary
  span is call-site granularity.** It is `function.Span` — the whole calling
  function's declaration. Any repair or explain-side reasoning that assumes
  a call-site token here will be wrong.
- **Do not flip `diagnostic.Error` to `diagnostic.ErrorWithRepairs` on any
  of the three D-13-09 codes without a plan task to re-pin
  `check_ordering_stability_test.go`.** See Section "Diagnostic ID churn"
  below — this is guaranteed, not conditional.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| callee-before-caller function ordering | A second topological sort | `callgraph.Order(program)`, reversed (same technique `buildInterproceduralSummaries` already uses internally at check.go:698) | D-13-03 explicitly forbids introducing new ordering authority |
| Whole-function span lookup | AST byte-range search | `core.Function.Span` / `ast.FuncDecl.Span` (both already exist, see above) | Both fields are already populated and already consumed by existing diagnostics |
| Statement-level span for the loan-liveness reorder repair | A new span-computation pass | The `:stmt`-suffixed `spanByOperationID` entries `borrowConflictDiagnosticPostAssembly` already reads (check.go:1244) | Same channel, same convention, already proven for the analogous intraprocedural repair |
| Repair applicability gating | A new gate mechanism | `diagnostic.NormalizeApplicability` / `DriverEligible` (diagnostic.go:75-95) — never defaults toward driver-eligible | Existing fail-closed discipline; D-13-07/D-13-10 explicitly inherit it |

**Key insight:** almost every mechanism D-13-09/D-13-14/D-13-19 need already
exists in the codebase in a directly analogous form (post-assembly repair
reconstruction, span-by-operation-ID threading, whole-function Span fields).
The phase's actual novelty is the **blame rule** (B1/B2/B3) and the
**cross-function narrows guard** — everything else is composition of
existing plumbing.

## Runtime State Inventory

Not applicable — this is not a rename/refactor/migration phase.

## Common Pitfalls

### Pitfall 1: Assuming Error→ErrorWithRepairs is free because D-13-04 says "zero churn"

**What goes wrong:** a plan reads D-13-04 ("zero Primary spans move today")
and infers the pinned diagnostic IDs are entirely stable through the phase,
then is surprised when `check_ordering_stability_test.go` goes red mid-Set-A
implementation.
**Why it happens:** D-13-04's claim is scoped to *blame adoption* (moving
Primary spans). D-13-09 is a separate, later decision that *adds Repairs* to
three of those same codes. `diagnostic.Error`'s identity struct is
`{Schema: "lang.diagnostic/0", Code, Span, Causes}`;
`diagnostic.ErrorWithRepairs`'s is `{Schema: "lang.diagnostic/1", Code, Span,
Causes, RepairKinds}` (`internal/compiler/diagnostic/diagnostic.go:108-152`,
read in full this session). The `Schema` string alone differs between the
two structs being hashed, so switching a code from `Error` to
`ErrorWithRepairs` changes its sha256 **even for a diagnostic instance that
ends up with zero repairs attached** (e.g., `check.call_argument_type_mismatch`
on a fixture where D-13-10's uniqueness gate finds ≥2 matches and emits no
repair — if the code path unconditionally calls `ErrorWithRepairs`, that
instance's ID still churns relative to today's `Error`-built ID).
**How to avoid:** the plan must include an explicit task, in the same wave
as the emission-site changes, to regenerate and re-pin these four rows in
`check_ordering_stability_test.go` (verified exact current values, read this
session):
  - `phase07/relay_escort_witness.lang` → `check.interprocedural_loan_liveness`, `diagnostic:58c1b5b2072cda60f77d721e` (line 193)
  - `phase08/relay_depth2_refuse.lang` → `check.interprocedural_loan_liveness`, `diagnostic:300248748c05f20eb1fe3948` (line 219)
  - `phase08/twin_a_refuse.lang` → `check.interprocedural_loan_liveness`, `diagnostic:e01ad6316e27899deeb610f7` (line 220)
  - `phase08/twin_b_refuse.lang` → `check.interprocedural_loan_liveness`, `diagnostic:d15b65a04fd8515f92e999e2` (line 253)
  - `phase4/fallible_call_unconsumed.lang` → `syntax.fallible_call_not_consumed`, `diagnostic:91c8b8c7a5d359d9a14fe6a8` (line 297)
  - `phase07/call_type_mismatch.lang` → `check.call_argument_type_mismatch`, `diagnostic:eec74c3869e957e7eccaafb8` (line 185)

  (That is 6 rows across the three affected codes, not 4 — corrected count
  after enumerating every table entry for each code; see the grep output
  this session for the full pinned table.)
**Warning signs:** `go test ./internal/compiler/check/...` failing on
`TestInterproceduralDiagnosticOrderingStability` or the sibling pinned-ID
test immediately after Set A's repair-emission tasks land, with a diff that
shows only the hash changing and the Code/fixture staying the same — that
is the expected, intentional churn, not a regression.

### Pitfall 2: `explainSpanStrictlyContains` narrows-flip risk is lower than D-13-20 implies

**What goes wrong:** the plan over-budgets time defending against every
existing explain fixture's edge kind flipping once the D-13-19 guard lands.
**Why it happens:** D-13-20 names `core.call_graph_cycle`'s `cycle_member`
causes as "the candidate" for a flip. Read directly
(`checkCallGraphAcyclic`, check.go:541-560): `cycle_member` cause spans come
from `spanByOperationID[edgeOperationIDs[i]]` — narrow per-`OpCall` token
spans, never a whole-function span. The **only** diagnostics whose Primary
equals a whole `core.Function.Span` today are `core.callee_not_callable`,
`core.invalid_body`, and `type.return_mismatch`
(`diagnostic.Error(code, function.Span, msg)` with **zero** `Cause` spans in
every one of those three call sites — read directly, no `Cause{..., Span:
...}` argument passed at any of check.go:200, :256, :1576-1579). A `narrows`
edge can only form when a *cause* carries a non-nil `Span`
(`buildExplainGraph`, `session_phase6_explain.go:183`,
`if edgeKind == protocol.EdgeCausedBy && cause.Span != nil`). Since none of
the three whole-function-Primary diagnostics has any spanned cause today,
**no existing shipped fixture can exercise the containment search that the
D-13-19 guard changes** — there is nothing for it to flip.
**How to avoid:** run the existing explain fixture suite as D-13-20
prescribes (do not skip it — this is source-level reasoning, not a
substitute for the actual test run), but calibrate expectations: the likely
outcome is that `lang.explain/1` is **not** triggered by D-13-19 landing,
because the flip condition cannot currently be constructed from any shipped
diagnostic. This is favorable evidence, not proof — a fixture combining
`core.callee_not_callable` with a later multi-cause diagnostic in the same
explain call (if one existed) could still be at risk; none does today.
**Warning signs:** if `TestExplainCauseGraphIsDeterministic` or
`TestExplainEdgeKindVocabularyIsClosed`
(`internal/compiler/session/session_phase6_explain_test.go:265`, `:204`)
changes output on an unmodified fixture, that is the real signal, not a
theoretical risk.

### Pitfall 3: D-13-10's uniqueness gate is easy to make vacuous given single-type-per-function maturity

**What goes wrong:** a held-out fixture for
`check.call_argument_type_mismatch`'s `use_matching_argument` repair
accidentally has either zero or every initialized place match, because the
"exactly one match" case is harder to construct than it sounds.
**Why it happens:** per `LANGUAGE-MATURITY.md`/D-07-09, every Lang function
has exactly one parameter and (per `sameType` enforcement at the head of
`checkLinear`/`checkBranch`/`checkFallibleLinear`) exactly one `TypeFact` —
every initialized place inside one function shares the *same* constructor
(`Byte` or `Buffer`). `resolveCallBinding`'s signature already carries
`places map[string]*placeState` (check.go:2958) — this **is** the in-scope
place set, no plumbing needed (see Section 3) — but because every place in
the function shares one constructor, the "matches" set is naturally either
{every initialized place} (≥2 easily) or {the single parameter, if nothing
else is in scope yet} (exactly 1). Constructing the exactly-one case
requires the call site to occur **before** any intervening `let` produces a
second initialized place of the same type.
**How to avoid:** build the derivation fixture with the call as the body's
*first* binding (only the parameter is in scope → uniqueness holds, repair
fires) and the held-out mismatch/ambiguous-refusal fixture with a prior
`let` in scope (two matches → `unrepairable`, per D-13-10's own required
positive test).
**Warning signs:** a fixture author writes a call-argument-type-mismatch
program with 3+ `let` bindings before the mismatched call site expecting a
clean unique-match repair — it will not fire; D-13-10 requires it not to.

## Code Examples

### 1. All eight interprocedural codes' current Primary span (D-13-04 empirical verification)

Read directly from `internal/compiler/check/check.go` and
`internal/compiler/core/core.go` this session. **All eight land on the
caller side (B2-shaped) — D-13-04's "zero churn from blame adoption" claim
is VERIFIED.**

| Code | Emission site | Primary span expression | Caller or callee? |
|------|---------------|--------------------------|--------------------|
| `check.interprocedural_loan_liveness` | check.go:1276-1288 `interproceduralLoanLivenessDiagnostic` | `spanByOperationID[move.ID]` — the offending `OpMove` | Caller (the function doing the move) [VERIFIED] |
| `core.call_graph_cycle` | check.go:541-560 `checkCallGraphAcyclic` | `spanByOperationID[cycle.ClosingOperationID()]` — the closing edge's own OpCall | Whole-cycle; exempt from B3 per D-13-05 [VERIFIED] |
| `check.call_argument_type_mismatch` | check.go:2993-3000 `resolveCallBinding` | `binding.RHS.Span` — the call site token | Caller [VERIFIED] |
| `check.call_arity_unsupported` | check.go:2959-2965 `resolveCallBinding` | `binding.RHS.Span` — the call site token | Caller (self-contained syntax defect) [VERIFIED] |
| `core.callee_not_callable` | check.go:1546-1580 `verifyCallableRefusal` | `function.Span` — **the whole calling function's declaration**, not a call-site token | Caller, coarse granularity [VERIFIED] |
| `check.call_return_type_unrepresentable` | check.go:3016-3026 `resolveCallBinding` | `binding.RHS.Span` — the call site token | Caller [VERIFIED] |
| `check.foreign_call_shape_unsupported` | check.go:2939-2942 `checkFallibleLinear` | `body.Span` — the function's own body | N/A — not Lang-to-Lang; only one party exists [VERIFIED] |
| `syntax.fallible_call_not_consumed` | check.go:3055-3057 `resolveCallBinding` | `binding.RHS.Span` — the call site token | Caller [VERIFIED] |

```go
// Source: internal/compiler/check/check.go:1546-1580 [VERIFIED]
// core.callee_not_callable's Primary is coarse — the WHOLE calling
// function's declaration span, not a call-site token.
for _, function := range functions {
    if function.Linear == nil { continue }
    for _, operation := range function.Linear.Operations {
        if operation.Kind != core.OpCall { continue }
        entry, ok := table.lookup(operation.CalleeID)
        callable := ok && entry.Callable
        if !callable {
            diag := diagnostic.Error(
                core.CalleeNotCallable, function.Span, "call target is not callable",
                diagnostic.Cause{Kind: "callee", Detail: operation.CalleeID},
            )
            return &diag
        }
    }
}
```

**Note on `core.callee_not_callable`'s B1/B2 defensibility:** CONTEXT.md's
own Claude's Discretion list flags this as "the one field where both
readings are defensible" — the code today implements B2 (blame the caller's
whole function), matching D-13-04's claim, but a callee that is genuinely
not `Callable` is arguably the *callee's own* declared-fact violation
(non-`Callable` is a property the callee's own declaration determines). The
plan should settle this explicitly rather than let the shipped B2 behavior
silently stand in as the "decision" — it is currently an artifact of
implementation order (the cycle-gate ordering comment at check.go:451-456
explains *why* this fires here), not a considered blame choice.

### 2. Pinned diagnostic IDs D-13-04 cites (verified exact table, check_ordering_stability_test.go:184-220)

```go
// Source: internal/compiler/check/check_ordering_stability_test.go [VERIFIED]
"phase07/call_type_mismatch.lang":          {"check.call_argument_type_mismatch", "diagnostic:eec74c3869e957e7eccaafb8"},
"phase07/call_uncallable_callee.lang":      {"core.callee_not_callable", "diagnostic:5735171812e7ad4920c3bb72"},
"phase07/cycle_indirect.lang":              {"core.call_graph_cycle", "diagnostic:1e6c260432c9010ac6a196a9"},
"phase07/cycle_mutual.lang":                {"core.call_graph_cycle", "diagnostic:39bb0a1a48d08fc9674307a4"},
"phase07/cycle_self.lang":                  {"core.call_graph_cycle", "diagnostic:6e836fd6f98202bf58f1dfdb"},
"phase07/cycle_through_match_arm.lang":     {"core.call_graph_cycle", "diagnostic:0fe333002fca961ebee8f00a"},
"phase07/cycle_unreachable.lang":           {"core.call_graph_cycle", "diagnostic:03edcf9d691da106106c2fc0"},
"phase07/foreign_symbol_shadowing.lang":    {"core.call_graph_cycle", "diagnostic:3fa66ee8773176867ffb9b4e"},
"phase07/relay_escort_witness.lang":        {"check.interprocedural_loan_liveness", "diagnostic:58c1b5b2072cda60f77d721e"},
"phase08/negative_control_fails.lang":      {"core.callee_not_callable", "diagnostic:79999a6354e9192f2578c976"},
"phase08/negative_control_infallible.lang": {"core.callee_not_callable", "diagnostic:d5d34a1946a3587c54c6177f"},
"phase08/relay_depth2_refuse.lang":         {"check.interprocedural_loan_liveness", "diagnostic:300248748c05f20eb1fe3948"},
"phase08/twin_a_refuse.lang":               {"check.interprocedural_loan_liveness", "diagnostic:e01ad6316e27899deeb610f7"},
"phase08/twin_b_refuse.lang":               {"check.interprocedural_loan_liveness", "diagnostic:d15b65a04fd8515f92e999e2"},
"phase4/fallible_call_unconsumed.lang":     {"syntax.fallible_call_not_consumed", "diagnostic:91c8b8c7a5d359d9a14fe6a8"},
```

**Rows that WILL churn when D-13-09's repairs land** (three
`check.interprocedural_loan_liveness` rows, one `syntax.fallible_call_not_consumed`
row, one `check.call_argument_type_mismatch` row — six rows total, all six
quoted above): every row whose Code is one of the three D-13-09 target
codes. Rows for `core.callee_not_callable` and `core.call_graph_cycle` are
NOT in D-13-09's Set A and stay untouched — those two codes' IDs are
genuinely stable through this phase.

### 3. Diagnostic identity construction (D-13-16/D-13-17's basis, verified in full)

```go
// Source: internal/compiler/diagnostic/diagnostic.go:108-152 [VERIFIED]
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
    // ... repairs sorted by Kind then Detail ...
    repairKinds := make([]string, len(repairs)) // only Kind, never Span/Replacement/Applicability
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

Confirms D-13-16 (Span/RepairKinds separation exact), D-13-17 (D-13-14's
additive `ExplainNode` fields touch neither struct — genuinely zero-churn),
and the Pitfall 1 finding (Schema string alone forces a churn on any
`Error`→`ErrorWithRepairs` conversion, independent of whether any repair
condition is actually met on a given instance).

### 4. `buildInterproceduralSummaries`' ordering is NOT exported (D-13-03's B3 needs its own call)

```go
// Source: internal/compiler/check/check.go:670-722 [VERIFIED]
// callgraph.Order's own doc comment names its result "reverse postorder",
// but empirically that order lists a CALLER before every function it
// calls -- the opposite of callee-before-caller. This loop walks
// callgraph.Order's result BACKWARD (last-visited-first) to get the
// genuine callee-before-caller traversal.
func buildInterproceduralSummaries(program core.Program, table callSignatureTable) (interproceduralSummaryTable, int) {
    order, err := callgraph.Order(program)   // order is a LOCAL var
    if err != nil { return result, 0 }
    for i := len(order) - 1; i >= 0; i-- {   // callee-before-caller walk
        functionID := order[i]
        // ... consumed here, never returned ...
    }
    return result, work   // <- only the summary map and a cost int leave this function
}
```

**Finding:** D-13-03's text ("the callee-before-caller topological order
`buildInterproceduralSummaries` already computes") is correct about the
*algorithm* (same `callgraph.Order` authority, same reversal technique) but
imprecise about *reuse* — nothing is exported for the blame resolver to
consume directly. The B3 tie-break code must call `callgraph.Order(program)`
a second time and reverse it itself (cheap, and explicitly sanctioned — "no
new ordering authority is introduced" — but it is new *code*, not a value
pulled from an existing return). The plan should budget this as a small,
separate helper (e.g. `calleeBeforeCallerOrder(program) ([]string, error)`)
that both `buildInterproceduralSummaries` and the new blame resolver call,
rather than duplicating the reversal loop inline in two places — this also
naturally satisfies "no new ordering authority" by construction (one call
site for the technique, two callers).

### 5. `resolveCallBinding`'s signature already exposes the in-scope place set (D-13-10's uniqueness gate input)

```go
// Source: internal/compiler/check/check.go:2958 [VERIFIED]
func resolveCallBinding(functionID string, opOrdinal int, binding ast.Binding,
    places map[string]*placeState, calleeContracts map[string]calleeContract,
    typeFact core.TypeFact, foreignSymbols map[string]foreignSymbolInfo,
) (core.LinearOperation, core.Place, *diagnostic.Diagnostic)
```

```go
// Source: internal/compiler/check/check.go:3847-3862 [VERIFIED]
type placeState struct {
    place        core.Place   // place.TypeID identifies the place's type
    declared     diagnostic.Span
    initialized  bool
    movedAt      *diagnostic.Span
    moveTargetID string
    moveTargetName string
}
```

**Finding:** the `places` parameter IS the in-scope place enumeration D-13-10
needs — no plumbing required, directly contradicting the possibility CONTEXT.md
flags ("Report whether in-scope place enumeration is available at that point
or must be plumbed" — it is available). The uniqueness gate is: iterate
`places`, keep entries where `initialized == true`, and match each
`place.TypeID`'s constructor against `contract.ParameterType`. See Pitfall 3
for why this maturity level makes the exactly-one-match case narrower to
construct than it first appears (single parameter, single type-fact per
function).

## State of the Art

Not applicable in the external sense (no library/tooling evolution to
track). Internally: this is the first phase to attach `Repair` to any of
these three codes (D-08-25's "Repairs is nil" is being reversed) and the
first to add per-node function identity to `lang.explain`.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `TestExplainNodeOrderIsStableOnTies` and the other seven tests in `session_phase6_explain_test.go` (listed by name in Code Example section headers above) are the complete set of shape-pinning tests at risk from D-13-19/D-13-20 — I read the test file's function names via grep, not each test body in full. | Explain-side surface | Low — grep-derived function name list is reliable; body-level assertions were spot-checked only for the graph-construction logic itself (`buildExplainGraph`, read in full), not each test's individual fixture data |
| A2 | No shipped fixture combines a whole-function-Primary diagnostic (`core.callee_not_callable`/`core.invalid_body`/`type.return_mismatch`) with an explain call that also has spanned sibling causes elsewhere in the same graph — Pitfall 2's "zero flip risk" conclusion depends on this holding for every fixture, not just the three I checked call sites for. | Common Pitfalls, Pitfall 2 | Medium — if some fixture routes a whole-function-Primary diagnostic through `lang explain` where causes ARE later added by a different, unexamined code path, the flip could still occur. D-13-20's own prescribed action (run the actual test suite before landing) remains the authoritative check; this finding narrows expectation, does not replace the test run. |
| A3 | `TestPhase6DefectCorpusIsHeldOut`'s full body (read only the first 25 lines) asserts nothing beyond byte-inequality across the `heldout_`/`derivation_` prefix split, as CONTEXT.md's D-13-24 states. | Section 7 | Low — CONTEXT.md's own text already asserts this and the read portion confirms the byte-collection logic; the remaining ~15 unread lines almost certainly just do the inequality comparison D-13-24 describes. |
| A4 | `antitheater_test.go`'s 794 lines were not read this session (only referenced) — I did not verify exactly what a new repair class must satisfy beyond what `repair_test.go`'s two named tests show. | Section 6 | Medium — the planner should budget a dedicated read of `antitheater_test.go` in the plan's first wave rather than trust this document's silence on its exact contract. |

## Open Questions

1. **Does `TestUnrepairableDefectFailsTheGate`'s existing fixture
   (`testdata/phase1/non_exhaustive.lang`) generalize, or does each new
   interprocedural class need its own dedicated unrepairable-case test?**
   - What we know: the existing test uses a single-function fixture unrelated
     to any interprocedural code; `repair_test.go:538-558` (read in full)
     shows it asserts `SubprocessCount == 1` and `Status == unrepairable`.
   - What's unclear: whether D-13-29's extension of this test to "the new
     classes" means adding new `t.Run` subtests with interprocedural
     unrepairable fixtures (most likely, given D-13-10's explicit "needs a
     positive test" requirement) or is satisfied by the existing single
     fixture alone.
   - Recommendation: plan for new subtests — D-13-10's own text ("The
     resulting `unrepairable` outcomes ... need a positive test, not a
     workaround") reads as requiring class-specific coverage, not reuse of
     the existing generic one.

2. **`antitheater_test.go`'s exact acceptance contract for a new defect
   class** — not read this session (794 lines, out of budget). The plan's
   first wave should include a dedicated read/summary task before any new
   injector or fixture is written, since D-13-24/D-13-30 explicitly extend
   its precedent.

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` (no external test framework anywhere in the tree) |
| Config file | none — standard `go test` |
| Quick run command | `go test ./internal/compiler/check/... ./internal/compiler/session/... -run TestInterprocedural` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| DX-06 (blame resolver, B1/B2/B3) | New rule classifies every interprocedural code correctly, `blame_undetermined` fires for unclassified declared facts | unit | `go test ./internal/compiler/check/... -run TestBlame` | ❌ Wave 0 — new test file/functions |
| DX-06 criterion 3 (twin pair) | Repair-then-re-check regression; naive detection-site fix leaves twin's other caller broken | integration | `go test ./cmd/lang-repair/... -run TestTwinPairBlame` (or equivalent name) | ❌ Wave 0 |
| DX-07 (3 new repairable classes) | Each of the 3 Set A classes reaches `Repaired` via the JSON protocol, on held-out fixtures | integration | `go test ./cmd/lang-repair/... -run TestRepairDriverFixesEveryDefectClassSinglePass` (extended) | ✅ exists, extend |
| DX-07 (held-out non-contamination) | `TestRepairDriverSourceNeverReferencesHeldoutFixtures` extended beyond `cmd/lang-repair` to `internal/compiler/check` | static/AST-inspection | `go test ./... -run TestRepairDriverSourceNeverReferencesHeldoutFixtures` (extended) | ✅ exists, extend scope |
| DX-05 (function attribution correctness) | D-13-23's three-function-plus-decoy fixture; test-side independent resolver; reorder-source mutation | unit | `go test ./internal/compiler/session/... -run TestExplainFunctionAttribution` | ❌ Wave 0 |
| DX-05 (narrows guard) | Function-scope span never parents via `narrows`; guard-disabled mutation fails | mutation-kill | `go test ./internal/compiler/session/... -run TestNarrowsFunctionScopeGuard` | ❌ Wave 0 |
| QLT-08 (mutation-kill every new control) | Every new injector, the topology-distinctness control, the sealed-digest control, the criterion-3 control — each individually mutation-killed | mutation-kill | per-control, see D-13-30(a-d) | ❌ Wave 0 |

### Sampling Rate

- **Per task commit:** targeted `go test ./internal/compiler/check/...` / `./internal/compiler/session/...` / `./cmd/lang-repair/...` scoped to the package touched
- **Per wave merge:** `go test ./...` (full suite — this project has no slow-integration-test carve-out; `go test ./...` is already the standing full-suite command per every prior phase's SUMMARY)
- **Phase gate:** full suite green, plus explicit re-run of `check_ordering_stability_test.go` to confirm the six expected-churn rows changed and nothing else did

### Wave 0 Gaps

- [ ] Blame resolver unit tests (`internal/compiler/check/*_test.go`) — new file needed, covers DX-06
- [ ] D-13-23's three-function decoy fixture + independent test-side resolver — new `.lang` fixture + test, covers DX-05
- [ ] D-13-28's twin-pair fixture (mirror included) — new `.lang` fixtures, covers DX-06 criterion 3 (the phase's riskiest-assumption gate)
- [ ] New interprocedural injectors in `session_phase6_injectors.go` or a phase-13 sibling — covers DX-07 criterion 2
- [ ] `testdata/phase13/HELDOUT.sha256` manifest + its own test — covers D-13-27
- [ ] Framework install: none — stdlib only

## Security Domain

`.planning/config.json`'s `workflow.security_enforcement` was not directly
read this session; treat as enabled per this document's own instructions
(absent = enabled) and STATE.md's note that Phase 11 shipped a full
`11-SECURITY.md` under this same default.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V5 Input Validation | yes | `resolveCallBinding`'s existing fail-closed admission gates; the new blame resolver must be equally fail-closed (D-13-07's `blame_undetermined` routing is exactly this) |
| V2/V3/V4 Auth/Session/Access Control | no | Not applicable — this is a compiler diagnostic surface, no auth boundary |
| V6 Cryptography | no | The sha256 diagnostic-identity hash is a content-addressing scheme, not a security cryptography boundary |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Fail-open blame default (B2 silently absorbing an unclassified B1 case) | Tampering (a repair could mask a real callee defect) | D-13-07's mandatory `blame_undetermined` routing with a compile-time exhaustiveness guard per declared-contract field |
| Overfit repair (fixes the derivation fixture, not the defect class) | Repudiation-adjacent (evidence looks valid but isn't) | D-13-26's topology-disjointness split, D-13-27's sealed held-out manifest, D-13-29's outcome-string assertion |
| Repair driver escaping its declared span | Tampering | `cmd/lang-repair`'s existing `CodeSpanOutOfRange` refusal (repair.go:237-242) — already proven, unaffected this phase (D-13-32) |

## Sources

### Primary (HIGH confidence — direct source read this session)

- `internal/compiler/check/check.go` (read in full 1-330, 440-565, 660-750, 1190-1530, 1546-1600, 2100-2180, 2900-3100) — every interprocedural emission site, `buildInterproceduralSummaries`, `resolveCallBinding`, `verifyCallableRefusal`, `checkCallGraphAcyclic`
- `internal/compiler/check/check_ordering_stability_test.go` (grepped pinned table, lines 184-314) — exact pinned diagnostic IDs
- `internal/compiler/diagnostic/diagnostic.go` (read in full) — `Error`/`ErrorWithRepairs` identity construction, `DriverEligible`, `NormalizeApplicability`
- `internal/compiler/core/core.go` (read lines 120-160; grepped `CallGraphCycle`/`Function struct`) — `Function.Span` field existence
- `internal/compiler/ast/ast.go` (grepped `FuncDecl`) — `FuncDecl.Span` field existence
- `internal/compiler/session/session_phase6_explain.go` (read in full) — `buildExplainGraph`, `explainSpanStrictlyContains`, `explainCorrelationKey`, `ExplainCommandFile`
- `internal/compiler/protocol/protocol.go` (read lines 190-270) — `ExplainNode`/`ExplainSummary`/`ExplainSchema`/edge vocabulary
- `internal/compiler/session/session_phase6_injectors.go` (read in full) — `Injector` interface, `markerGuard`, all five injectors, `AllInjectors`
- `cmd/lang-repair/repair.go` (read in full) — `Repair()`'s single-pass structure, `driverEligible`, `applyRepair`, `selectRepair`
- `cmd/lang-repair/repair_test.go` (read `TestRepairDriverSourceNeverReferencesHeldoutFixtures`, `TestRepairDriverFixesEveryDefectClassSinglePass`, `TestUnrepairableDefectFailsTheGate`, `phase6Fixture` helper)
- `cmd/lang/main.go` (grepped `isPhase6Corpus`/`isPhase7Corpus`) — marker-file corpus dispatch pattern
- `internal/compiler/session/session_phase6_injectors_test.go` (read `TestPhase6DefectCorpusIsHeldOut`'s first 25 lines) — byte-inequality-only assertion, confirms D-13-24's own critique of M001's methodology
- `testdata/phase6/` (directory listing) — exactly 6 `.lang` fixtures (match/move/borrow × heldout/derivation) + README + stale_evidence_subject.lang; zero interprocedural fixtures, confirming D-13-33's "topology triple degenerates" concern is not merely likely but structurally certain (no call graph exists in any of these 6 files)

### Secondary (MEDIUM confidence)

- None — no external documentation was consulted per this task's explicit scope restriction (all prior external research already done by the four advisor agents referenced in CONTEXT.md's `<specifics>` section).

### Tertiary (LOW confidence)

- None.

## Metadata

**Confidence breakdown:**
- Diagnostic Primary-span verification (Section "Code Examples" #1): HIGH — every code's emission site read directly, exact line numbers cited
- Diagnostic ID churn finding: HIGH — identity struct read in full, the Error→ErrorWithRepairs schema difference is unambiguous Go source
- Explain-side narrows-flip risk assessment: MEDIUM — reasoned from source but not from an actual `go test` run; the prescribed D-13-20 test run remains authoritative
- `antitheater_test.go`'s exact new-class contract: LOW — not read this session, flagged as Open Question 2 with an explicit Wave 0 recommendation

**Research date:** 2026-09-13
**Valid until:** stable — this is direct-source verification of a codebase under active development within this same milestone; re-verify any claim if the referenced file's line numbers have shifted by the time the plan executes (expected to be hours, not weeks, given this is the final phase of an in-progress milestone).
