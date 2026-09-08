# Phase 07: Calls, Signatures, and Call-Graph Refusal - Research

**Researched:** 2026-09-08
**Domain:** Interprocedural compiler admission (typed-core call surface, digest-bound signature summaries, cycle refusal over an explicit call graph)
**Confidence:** HIGH for structural/file:line claims (all read this session); MEDIUM for sizing; LOW/ASSUMED flagged inline

## Summary

Phase 07 is not a green-field call feature — it is a **forcing function** wired through nine files that already exist and already dispatch on `core.OperationKind`. The single highest-value finding of this research is that **CONTEXT.md's framing of D-07-01 as "widening one predicate at `check.go:1316`" undersells the work by a full architectural layer**: the parser today does not merely lack call syntax, it **actively refuses** a bare call in a binding RHS at parse time (`syntax.fallible_call_not_consumed`, `parser.go:404-421`), unconditionally, before the callee's identity (foreign vs. Lang) is even known. Making `let r = f(x)` legal for a Lang callee therefore requires **parser work** (a new `ast.RHS.Kind`, e.g. `"call"`) plus **moving the foreign/Lang admission decision from parse-time to check-time** — not a checker-only change. Separately, `resolveForeignStep` (`check.go:1316`) is on the *fallible-foreign* path only (`checkForeignTracer`/`checkResourceLifecycle`, entered via `hasTryCall`), which a non-fallible Lang call never reaches; the natural home for the new admission logic is a **new case inside `analyzeStraightLine` (`check.go:1804`) and `analyzeArmBody` (`check.go:784`)**, which already carry the per-binding, name-keyed scope map (`places map[string]*placeState`) that D-07-01's "any in-scope binding" rule needs.

Two more corrections carry real planning weight. First, `debugmap.go:54` is a **comment** listing example schema constants, not a functional `/1`-bump site — the real bump-touch count is closer to 3 (`core.go` const, `originvalidate.BuildInterface`, `protocol.go`'s `InterfaceSummary`/`InterfaceFunctionAnswer` projection and `session.go`'s `InterfaceExportCommandFile`), not the 4 CONTEXT.md estimated, and CONTEXT.md's cited function name `RunInterfaceCommandFile` does not exist — the actual producer entry point is `session.InterfaceExportCommandFile` (`session.go:1028`). Second, and most consequential for Success Criterion 2: **`testdata/phase3/exclusive_borrow_clean` does not exist anywhere in the tree**, nor does any file with the `escort`/dangling-alias witness shape from D-04-03's narrative — it was a discussion example, never materialized as a fixture. Phase 07 must create both the single-function "clean but uncallable" fixture and (for the callgraph corpus) the two-function `relay`/`escort` dangling-alias witness from scratch; there is no existing negative control to point at.

On the dispatch-site cost side, the news is better than sized: `cgen.Emit`/`EmitNative` hard-fail on `len(Functions) != 1` (`cgen.go:22,52`), and **no accepted program containing a real `core.OpCall` can ever have exactly one function** (a call needs a distinct callee; self-recursion is refused by the very cycle check this phase adds). Both `core_test.go`'s `TestAllOperationKindsHandledAtEverySite` and `session.go`'s `lane:kind-exhaustive-dispatch` gate `cgen.Emit` behind `if len(program.Functions) == 1`, so **cgen's runtime dispatch is never exercised by any legal OpCall fixture under either control as currently written**. The minimum change to keep both controls green is adding `core.OpCall` to `AllOperationKinds()` and to `encountered`/`encounteredKinds` bookkeeping (which iterate all operations regardless of the cgen guard) — cgen's four switch statements can still gain a documented, structurally-unreachable `case core.OpCall:` (mirroring `interp.go`'s existing precedent of listing `OpForeignCall`/`OpRelease` in `runBranchArm` "solely so the six-site table finds it handled"), but this is defensive/forward-compatible hygiene for Phase 11, not something the two exhaustive-dispatch controls require this phase.

Both phase-scoped required-kind fixture lists (`core_test.go`'s in-process `fixtures` slice and `session.go`'s `dispatchFixtures`/`lane:kind-exhaustive-dispatch`) are **literal, hand-maintained lists per phase** — Phase 07 needs its own new fixture list and its own required-kinds check (`[]core.OperationKind{core.OpCall}` at minimum), following the exact Phase-4 pattern rather than inventing a new mechanism.

**Primary recommendation:** Sequence Stage 0 exactly as CONTEXT.md specifies (widen `Interface`/`FunctionSignature`, land the `corevalidate` peer, gate against the ~39-function M001 Phase 1-4 corpus) — that work is a clean, well-understood extension of an existing, already-proven artifact. Budget Stage 1's `check` producer plan as **parser-plus-checker** work (grammar change, new RHS kind, relocated admission diagnostic, new `analyzeStraightLine`/`analyzeArmBody` call case), not checker-only, and budget one plan-sized task purely for manufacturing the two missing negative-control fixtures (`exclusive_borrow_clean`-equivalent, and the `relay`/`escort` two-function witness) before SEM-06's gate can be written at all.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SEM-04 | `OpCall` is a real `core.OperationKind` handled at all six dispatch sites with both exhaustive-dispatch controls green | Pattern 1 (exact edit sites per dispatch site), Pattern 2 (what the two controls actually require), Pitfall 4 (cgen is structurally unreachable this phase) |
| SEM-05 | Digest-bound callee signature summary extending `core.Interface`/`FunctionSignature`; no caller admission reads a callee body | Architecture Diagram (Stage 0 producer flow), Code Examples (`analyzeStraightLine`'s scope map), Open Question 3 (`ClosureDigest` base case) |
| SEM-06 | Call admitted only when callee is callable ⊆ publishable; refusal has a stable diagnostic code | Pitfall 3 (the negative-control fixtures this predicate needs do not exist and must be authored), Security Domain (V4 mapping) |
| SEM-07 | Call graph constructed; cycles (direct/mutual/indirect, including through `Result` matching) refused with a named code, never a hang | Pattern 3 (`callgraph` package template), Pitfall 2 (diamond corpus requirement), Security Domain (DoS bound) |
| QLT-08 | Every new interprocedural control mutation-killed in the plan that introduces it | Validation Architecture (Phase Requirements → Test Map row; confirms no reusable QLT-08 registry exists, `qlt01_registry.json` is a distinct M001 mechanism) |
</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Call syntax admission (`let r = f(x)`) | Frontend — parser (`syntax`) | Frontend — checker (`check`) | Parser must stop universally refusing a bare call RHS; checker resolves foreign-vs-Lang and admits/refuses by identity |
| Callee signature summary (Stage 0) | Frontend — `originvalidate` (producer) | Frontend — `corevalidate` (independent peer) | `core.Interface`/`FunctionSignature` is already the body-stripped, digest-bound artifact; D-07-20 requires two independent derivations, never one |
| `OpCall` semantic admission | Frontend — `check` (producer) | `corevalidate`, `originvalidate` (independent re-derivation) | check emits `OpCall` from the callee's summary alone; both validation peers re-derive independently, never sharing the producer's state |
| Cycle refusal | Frontend — new `internal/compiler/callgraph` package, invoked from `check` | `corevalidate` (independent peer, synthetic-artifact only) | D-07-14: cycle refusal runs over `check`'s own completed in-memory `core.Program`, never AST-side; `corevalidate` re-derives with disjoint reachable input space |
| Interpreter execution of `OpCall` | Execution — `interp` | — | Out of scope for real call-stack semantics this phase (SEM-08 is Phase 10); dispatch-site presence only |
| Native lowering of `OpCall` | Codegen — `cgen` | — | Structurally unreachable this phase (see Summary); registered for exhaustive-dispatch hygiene only, real multi-function emission is Phase 11 (NAT-04) |
| Loan-chain re-derivation across a call | Frontend — `pathoracle` | — | Deferred: pathoracle's cross-function loan-chain rule is TRU-03/Phase 10, not this phase; Phase 07 only needs pathoracle to not crash walking an `OpCall`-bearing function |

## Package Legitimacy Audit

No external packages are introduced by this phase's production code. `pgregory.net/rapid` (test-only, per STANDING-VERDICTS.md "Adopt") is **not yet present in `go.mod`/`go.sum`** — verified by `grep pgregory go.mod go.sum` returning nothing. If any Stage 0/1/2 plan wants property-based generation for the synthetic-`core.Program` peer tests (D-07-19), the import must be added at that point and is a **test-only dependency**, which does not affect the zero-external-production-dependency record. `[VERIFIED: go.mod:1-3, go.sum grep — this session]`

| Package | Registry | Age | Downloads | Source Repo | Verdict | Disposition |
|---------|----------|-----|-----------|-------------|---------|-------------|
| `pgregory.net/rapid` | Go module proxy | pre-existing standing verdict, not re-audited this session | — | github.com/flyingmutant/rapid | Not re-checked (already `Adopt` in STANDING-VERDICTS.md) | Add to `go.mod` as test-only when first plan needs it; no production impact |

**Packages removed due to [SLOP] verdict:** none.
**Packages flagged as suspicious [SUS]:** none.

## Standard Stack

No new production libraries. The "stack" for this phase is entirely in-repo Go code extending existing packages (`core`, `check`, `corevalidate`, `originvalidate`, `interp`, `cgen`, `pathoracle`, `session`, `diagnostic`, `protocol`) plus one new package, `internal/compiler/callgraph`, modeled directly on `pathoracle`'s existing posture (see Architecture Patterns below). `[VERIFIED: package inventory via find/grep — this session]`

## Architecture Patterns

### System Architecture Diagram

```
   source (.lang)
        │
        ▼
  ┌───────────┐   syntax.linearBody() must stop unconditionally
  │  syntax   │   refusing a bare call RHS (parser.go:404-421) and
  │ (parser)  │   instead emit a new RHS.Kind ("call") — foreign-vs-
  └─────┬─────┘   Lang identity is NOT known at parse time
        │ ast.Program
        ▼
  ┌───────────┐   check.Program() (check.go:44) already builds
  │   check   │   foreignSymbols + functionNames maps; a bare-call
  │(producer) │   binding now resolves against BOTH:
  │           │     - functionNames[callee] → emit core.OpCall via a
  │           │       new case in analyzeStraightLine (:1804) /
  │           │       analyzeArmBody (:784), consulting the callee's
  │           │       digest-bound core.Interface (Stage 0 artifact)
  │           │       — never the callee's own core.Function body
  │           │     - foreignSymbols[callee] → still requires try/
  │           │       discard; unwrapped bare call now refused HERE
  │           │       (relocated from parser), same diagnostic code
  │           │   After emitting OpCall for the whole core.Program,
  │           │   check runs the new callgraph package over its own
  │           │   completed in-memory Program and refuses cycles
  │           │   BEFORE returning it (D-07-14) — no AST-side graph.
  └─────┬─────┘
        │ core.Program (OpCall-bearing, acyclic)
        ▼
  ┌────────────────────────────────────────────────────────────┐
  │           corevalidate (independent peer)                   │
  │  - replayStraightLine (:841, case core.OpCall added ~:904)  │
  │  - replayBlocks       (:1039, case core.OpCall added ~:1099)│
  │  - independent signature-summary re-derivation (Stage 0)    │
  │  - independent callgraph traversal over SYNTHETIC            │
  │    core.Program values only (D-07-19), disjoint input space  │
  └─────┬────────────────────────────────────────────────────────┘
        │ validated core.Program
        ├──────────────► pathoracle.RecomputeEndpoints (must not error;
        │                 cross-function loan-chain rule deferred to Ph.10)
        ├──────────────► originvalidate.RecomputeOriginPerReturn (must not
        │                 crash) + BuildInterface (widened: Stage 0 producer)
        ├──────────────► interp.Run (per-function dispatch case for OpCall;
        │                 no real call-stack semantics this phase)
        └──────────────► cgen.Emit  — GATED: only invoked when
                          len(program.Functions) == 1, which no legal
                          OpCall-bearing program ever satisfies; dispatch
                          case is hygiene, not exercised this phase
```

### Recommended Project Structure

```
internal/compiler/
├── core/core.go            # Interface/FunctionSignature widened to /1; OpCall registered
├── syntax/parser.go        # new RHS.Kind "call"; syntax.fallible_call_not_consumed
│                           # refusal narrows to "foreign callee, not try/discard-wrapped"
├── check/check.go          # new call-admission case in analyzeStraightLine/analyzeArmBody;
│                           # callgraph invocation before Program() returns
├── corevalidate/corevalidate.go  # case core.OpCall in both replay* switches;
│                                 # independent signature-summary + callgraph peers
├── originvalidate/originvalidate.go  # BuildInterface widened (Stage 0 producer)
├── interp/interp.go        # dispatch case for OpCall (no call-stack semantics)
├── cgen/cgen.go            # dispatch case for OpCall (structurally unreachable this phase)
├── pathoracle/pathoracle.go # must not error walking an OpCall-bearing function
├── callgraph/              # NEW package — sibling of pathoracle, not a novel pattern
│   ├── callgraph.go        # three-color DFS, backEdge-shaped typed errors
│   └── callgraph_test.go   # synthetic core.Program builder + mutation matrix
├── diagnostic/diagnostic.go # no structural change; Causes already participate in ID (:96-112)
├── protocol/protocol.go    # truncated:core.call_cycle_bound convention copy
└── session/session.go      # new phase-07 dispatchFixtures list + required-kinds check;
                             # InterfaceExportCommandFile writes /1, decodes frozen /0
testdata/phase07/            # new corpus: two-function calls, diamonds+shared-leaves,
                             # self-call, unreachable-cycle, foreign-shadowing, unresolvable
                             # callee, exclusive_borrow_clean-equivalent, relay/escort witness
```

### Pattern 1: Six-site registration is a forcing function, not a metaphor

**What:** Adding `core.OpCall` to `AllOperationKinds()` (`core.go:250-252`) breaks the `default:` arm of every exhaustive switch simultaneously.
**When to use:** Always, for any new `OperationKind` — this is D-04-22's established discipline.
**Verified sites and exact current line numbers this session:**

| Site | File | Switch(es) | Current line(s) |
|---|---|---|---|
| `check` | `check.go` | new admission case (not a switch on `OperationKind` directly — see Pattern 2) | `analyzeStraightLine` binding-kind switch starts `:1804`; the `switch binding.RHS.Kind` is inline per-binding, `"take"` case at `:1877` |
| `corevalidate` | `corevalidate.go` | `replayStraightLine` switch `operation.Kind` | func starts `:841`; `case core.OpForeignCall:` at `:962`; `default:` at `:1003` |
| `corevalidate` | `corevalidate.go` | `replayBlocks` switch `operation.Kind` | func starts `:1039`; `case core.OpForeignCall:` at `:1161`; `default:` at `:1216` |
| `interp` | `interp.go` | `runBranchArm` switch | `case core.OpForeignCall:` at `:97` (documented-unreachable-this-phase precedent to copy) |
| `interp` | `interp.go` | `runLinearBlocks`, `runLinear` switches | multiple `case core.Op*` blocks, `:221-318` and `:409-428` |
| `cgen` | `cgen.go` | four case lists `core.OpCopy, core.OpMove, core.OpBorrowShared, core.OpBorrowExclusive:` | `:260`, `:467`, `:652`, `:1652` (one off from CONTEXT.md's `:259,:466,:651` — verify against live tree before editing, do not trust either citation blindly) |
| `pathoracle` | `pathoracle.go` | no exhaustive `OperationKind` switch — walks terminators/blocks only | confirmed via `TestAllOperationKindsHandledAtEverySite`'s own doc comment: "pathoracle and originvalidate do not switch exhaustively... 'handled' means the walk completes without error" |
| `originvalidate` | `originvalidate.go` | same as pathoracle — terminator-membership walk, `isTerminatorKind` (`:44-51`) | no `OperationKind` switch to edit for dispatch; `BuildInterface` (`:415-434`) is the Stage 0 producer edit, separate concern |

`[VERIFIED: internal/compiler/{core,corevalidate,interp,cgen,pathoracle,originvalidate}/*.go — read this session]`

### Pattern 2: The two exhaustive-dispatch controls are literal fixture lists, not generic sweeps

**What:** `core_test.go`'s `TestAllOperationKindsHandledAtEverySite` (`:266-357`, the fixtures slice at `:272-284`) and `session.go`'s `lane:kind-exhaustive-dispatch` (`:2501-2568`, `dispatchFixtures` slice at `:2513`) are each a **hand-maintained list of `.lang` file paths**, run through the full `session.Check → corevalidate.Validate → pathoracle/originvalidate/interp → cgen(if len==1)` pipeline, plus a **separate hand-maintained list of required `core.OperationKind` values** that must appear in `encountered`/`encounteredKinds`.
**What this means for Stage 1:** neither control requires a *synthetic* per-kind fixture — both require a **real, admitted, two-function `.lang` program producing `core.OpCall`** added to the phase's own fixture list (`core_test.go`'s combined-corpus list is cumulative across all phases; `session.go`'s list is phase-scoped, e.g. `dispatchFixtures` for Phase 4 only exercises `OpForeignCall, OpRelease, OpFail, OpDefect`). Phase 07 needs:
1. One new `.lang` fixture (or more) added to `core_test.go`'s `fixtures` slice (`:272-284`), containing a real, legal `OpCall`.
2. A **new, phase-scoped block in `session.go`** mirroring `:2501-2568` exactly — its own `dispatchFixtures` list and its own required-kinds check `[]core.OperationKind{core.OpCall}` (at minimum) — because the existing Phase 4 lane is scoped to Phase 4's four kinds only and must not be silently repurposed.
`cgen.Emit` will not run for this fixture in either control (needs ≥2 `Functions`), which is a genuine, favorable finding: **cgen's real dispatch behavior for `OpCall` is untested by either control this phase**, and the plan should say so explicitly rather than claim coverage it does not have (this is the D-07-04-style "declared, not implied" discipline the milestone already practices).
`[VERIFIED: internal/compiler/core/core_test.go:266-357, internal/compiler/session/session.go:2501-2568 — read this session]`

### Pattern 3: `pathoracle` as the `callgraph` package template

**What:** `pathoracle`'s package doc (`:1-29`) states its posture precisely: "consumed only by tests and the verify harness — never by check, corevalidate, interp, or cgen," imports neither `check` nor `corevalidate` nor `ast`, reads only `core.Function`'s own declared Blocks/Edges/Operations. Its typed-error shape:
```go
// Source: internal/compiler/pathoracle/pathoracle.go:112-125
type backEdgeError struct {
    functionID, blockID string
}
func (e *backEdgeError) Error() string {
    return fmt.Sprintf("pathoracle.cfg_back_edge: function %q block %q participates in a cycle", e.functionID, e.blockID)
}
func (e *backEdgeError) Code() string { return "pathoracle.cfg_back_edge" }
```
and the companion `pathCapError` (`:96-107`) with its own `Code()` method and an `errors.As`-style accessor function (`PathCapError(err error) (*pathCapError, bool)`).
**When to use for `callgraph`:** `internal/compiler/callgraph` should be a **sibling package**, NOT called by `check`/`corevalidate`/`interp`/`cgen` as an import — rather, `check` calls it directly (D-07-14: "check emits OpCall normally, then runs callgraph over its own completed in-memory core.Program and refuses before returning it"), and `corevalidate` implements **its own independent traversal**, not an import of `callgraph` (this is the same "no shared derivation" discipline as D-07-22's "no shared `callsummary` package"). The typed-error shape (`cycleError` or similar, with a `Code()` method returning `"core.call_graph_cycle"`) should mirror `backEdgeError`/`pathCapError` exactly: a private struct, an `Error() string`, a `Code() string`, and an `errors.As`-style accessor.
**Difference from pathoracle:** pathoracle is consumed only by tests/verify harness; `callgraph` is consumed by `check` in production (a real admission dependency), so it needs a stable public entry point (e.g. `callgraph.FindCycle(program core.Program) (reversePostorder []string, err error)`) rather than pathoracle's test/verify-only surface.
`[VERIFIED: internal/compiler/pathoracle/pathoracle.go:1-125 — read this session]`

### Anti-Patterns to Avoid

- **Widening `resolveForeignStep` directly for Lang calls.** `resolveForeignStep` (`check.go:1316-1355`) is entered only from `checkForeignTracer`/`checkResourceLifecycle`, which are reached only when `hasTryCall(body)` is true (`check.go:115-118`). A non-fallible Lang call (no `try`/`discard`) never reaches this function. Do not try to make one predicate serve both call families; the argument-shape *rule* generalizes (arity-1, in-scope-binding), but the *code path* does not.
- **Assuming the parser already accepts `f(x)` as a legal binding RHS.** It does not — it is a hard parse-time refusal today (`syntax.fallible_call_not_consumed`, `parser.go:404-421`), independent of callee identity.
- **Assuming `debugmap.go:54` needs a functional edit for the `/1` bump.** It is a comment (`"...(core.Schema1, evidence.Schema1, core.InterfaceSchema, ...) — no existing..."`) listing example schema constants for illustration, not a decode/encode site.
- **Citing `session.RunInterfaceCommandFile`.** No such function exists; the actual name is `session.InterfaceExportCommandFile` (`session.go:1028`).
- **Assuming `cgen` needs real multi-function-call emission this phase.** It cannot be exercised by any legal `OpCall` fixture under either exhaustive-dispatch control (see Pattern 2); building it would be scope creep into Phase 11 (NAT-04).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Cycle detection over a call graph | A bespoke recursive-descent walker | Iterative three-color (white/gray/black) DFS over an explicit stack (D-07-18, locked) | Native Go recursion risks the exact stack-blowup Pitfall 4 names; explicit worklist is already the project's established pattern (pathoracle, loanLivenessFixpoint) |
| Independent re-derivation "proof" | A single shared `callsummary` package imported by both `check` and `corevalidate`/`originvalidate` | Two independently-implemented derivations (D-07-20/D-07-22), enforced by a static import-independence test | A shared implementation cannot diverge by construction — exactly what QLT-08's seeded-fault discipline needs to exercise |
| Content-vs-authenticity signature trust | A keyed/signed digest scheme | Unkeyed content digest (`ClosureDigest`, same shape as `CoreDigest`) plus independent re-derivation as the forgery answer (D-07-13) | Explicitly out of scope; digests prove staleness only, never forgery — do not build crypto machinery this phase |
| Diagnostic identity stability | Ad hoc string formatting of cycle membership | Canonical rotation (lexicographically smallest function ID at index 0) before constructing `Causes`, exploiting the existing `Diagnostic.ID` derivation (`diagnostic.go:96-112`, sha256 over `{Schema, Code, Span, Causes}`) | Without rotation, the same cycle discovered from two root orderings yields two different published IDs — an unstable versioned API surface |

**Key insight:** every "don't hand-roll" item in this phase is really the same insight restated — the project already has the exact structural precedent (pathoracle for callgraph, diagnostic.go's rotation-sensitive ID for D-07-16, the `/0`→`/1` bump discipline for Interface) and the discipline is to **reuse the shape**, not invent an adjacent one.

## Runtime State Inventory

Not applicable — Phase 07 is additive (new grammar, new operation kind, new package), not a rename/refactor/migration. No stored data, live service config, OS-registered state, secrets, or build artifacts carry a name this phase changes.

## Common Pitfalls

### Pitfall 1: Treating the parser refusal as cosmetic

**What goes wrong:** A plan that scopes Stage 1's `check` producer path as "checker-only" will discover mid-plan that `let r = f(x)` never reaches `check` at all — it is refused by the parser before an `ast.Program` even exists.
**Why it happens:** CONTEXT.md's own wording ("this widens the existing predicate... one predicate, not a new grammar") describes the *admission rule* shape, which is correctly grammar-free at the checker level, but elides that the **parser's blanket refusal of any bare call** is itself a piece of grammar that must be relaxed.
**How to avoid:** Size Stage 1's first plan to include: (1) a new `ast.RHS.Kind` (e.g. `"call"`) parsed at `parser.go`'s existing `if p.peek().Kind == TokenLParen` branch (`:404-421`) instead of unconditionally emitting `"call_unconsumed"` + `syntax.fallible_call_not_consumed`; (2) relocating the "must be try/discard-wrapped" refusal to `check.go`, where `foreignSymbols`/`functionNames` (`check.go:82` area, built in `Program()`) are available to distinguish a foreign callee (still refused bare) from a Lang callee (now admitted).
**Warning signs:** A plan whose only diff is inside `check.go`'s existing foreign-call functions, with no `syntax/parser.go` or `syntax/parser_test.go` changes.

### Pitfall 2: Building the cycle refusal test corpus as chains, not diamonds

**What goes wrong:** A straight `A→B→C→…→Z` acyclic chain corpus will not kill the highest-value seeded mutation (swapping the on-stack "gray" check for a plain "visited" check), because reverse postorder never revisits a node on a pure chain.
**Why it happens:** A chain is the easiest fixture to write and "looks" pathological (deep), but depth alone does not exercise re-entry detection.
**How to avoid:** Build the "pathological-depth-but-acyclic" corpus with **diamonds and shared leaves** — `A→B, A→C, B→D, C→D`, repeated/nested — per D-07-specifics. This is already flagged as locked in CONTEXT.md; the research confirms no existing fixture of this shape exists anywhere in `testdata/` today (`grep -rl "fn escort\|fn relay" testdata/` finds only single-function `relay`-named fixtures, none multi-function).
**Warning signs:** A callgraph test file whose only acyclic fixtures are single-chains.

### Pitfall 3: Assuming `exclusive_borrow_clean` and the `relay`/`escort` witness exist

**What goes wrong:** A plan step that reads "wire the existing `exclusive_borrow_clean` fixture into the SEM-06 negative control" will fail at execution — no such *file* exists in `testdata/`. **CORRECTED (D-07-44):** the witness itself **does** exist, as inline Go source at `check/check_exclusive_test.go:46-58` (`module owned.exclusive_borrow_clean`, `fn relay`), asserted again at `originvalidate_test.go:230`. The task is to **extract** it verbatim into `testdata/phase07/`, not to author a new one — and it is the *correct* witness because it fails publication with `core.origin_omitted`, which is D-07-31's real predicate.
**Why it happens:** The fixture name and the `relay`/`escort` witness program are quoted verbatim from Phase 4's `04-CONTEXT.md` discussion narrative (D-04-03), which describes them as the **decisive research witness**, not as artifacts that were ever committed to `testdata/`.
**How to avoid:** Treat fixture creation as its own task: (1) a single-function fixture that checks clean today but must become `Callable == false` once D-04-03's predicate is implemented (candidates: any function whose `PublicOrigin.Access == "exclusive"` with no further constraint — `testdata/phase3/exclusive_exclusive_reject.lang` and `exclusive_move_reject.lang` exist but are *rejection* fixtures, not clean-but-uncallable ones; none of the existing phase3 `exclusive_*` fixtures is a clean-accept case per a quick content check, so a new one must be authored); (2) the two-function `relay`/`escort` dangling-alias witness, verbatim from `04-CONTEXT.md`'s D-04-03 block, as a **new** `testdata/phase07/` fixture proving the call-admission gate actually closes the D-03-02-adjacent hole this phase exists to prevent from reopening.
**Warning signs:** A plan or PLAN.md task listing `testdata/phase3/exclusive_borrow_clean` as an input file.

### Pitfall 4: Believing cgen "needs" `OpCall` handling this phase

**What goes wrong:** Sizing a cgen plan/task assuming real code generation for Lang-to-Lang calls, duplicating Phase 11's scope.
**Why it happens:** The six-dispatch-site doctrine (D-04-22) is easy to over-read as "every site must fully execute every kind."
**How to avoid:** Confirm from `cgen.go:22,52`'s `len(Functions) != 1` guard and both exhaustive-dispatch controls' `if len(program.Functions) == 1` gating (`core_test.go:346`, `session.go:2559`) that cgen's `Emit`/`EmitNative` will never actually be invoked on a legal `OpCall`-bearing program this phase. The task is: add `core.OpCall` to the existing `case core.OpCopy, core.OpMove, core.OpBorrowShared, core.OpBorrowExclusive:` lists (four sites) with a doc comment citing this exact reachability argument (mirroring `interp.go:97`'s "no fixture this phase can produce this" precedent) — a small, safe, forward-compatible edit, not new emission logic.
**Warning signs:** A cgen task with acceptance criteria referencing `.c` output correctness for a two-function program.

## Code Examples

### The existing binding-scope map D-07-01's admission rule should consult

```go
// Source: internal/compiler/check/check.go:1822-1825 (analyzeStraightLine)
places := map[string]*placeState{
    parameterName: {place: result.Places[0], declared: parameterSpan, initialized: true},
}
// ... each subsequent binding.Name is added to `places` as it is processed,
// which is exactly the "any in-scope binding" set D-07-01's admission rule
// needs to validate a call argument against (not just the function parameter).
```

### The parser's current unconditional refusal (must become conditional)

```go
// Source: internal/compiler/syntax/parser.go:404-421 (linearBody)
source := p.identifier("syntax.expected_binding_source")
if p.peek().Kind == TokenLParen {
    // D-04-06: a fallible foreign call is admissible only as the
    // operand of `try` (or `discard ... because`, not yet a parsed
    // construct this phase). A bare call in a binding right-hand
    // side is refused here, at parse time...
    p.problem("syntax.fallible_call_not_consumed", source, "a fallible call must be the operand of `try`")
    _, end := p.callArguments()
    body.Bindings = append(body.Bindings, ast.Binding{
        Name: name.Text, RHS: ast.RHS{Kind: "call_unconsumed", Source: source.Text, Span: source.Span}, Span: spanFrom(bindingStart, source),
    })
    body.Span.End = end
    continue
}
```

### The dispatch pattern to copy for a new `core.OpCall` case (corevalidate)

```go
// Source: internal/compiler/corevalidate/corevalidate.go:962-972 (replayStraightLine, OpForeignCall case)
case core.OpForeignCall:
    if !v.targetMatches(function, index, operation, places, produced) {
        return false
    }
    initialized[operation.TargetID] = true
    produced[operation.TargetID] = true
    errTarget, errKnown := places[operation.ErrTargetID]
    if !v.check(errKnown && operation.ErrTargetID != operation.TargetID && operation.ErrTargetID != operation.SourceID && !produced[operation.ErrTargetID] && errTarget.TypeID != "", "core.invalid_target", operation.ErrTargetID) {
        return false
    }
    initialized[operation.ErrTargetID] = true
    produced[operation.ErrTargetID] = true
// OpCall (non-fallible, single successor) is structurally closer to
// OpCopy/OpMove (single TargetID, no OkEdgeID/ErrEdgeID/ErrTargetID) than
// to OpForeignCall — model its replay case on OpCopy/OpMove's shape
// (:904-912, :946-956) plus a callee-ID resolution check reusing
// isDeclaredFunctionName (:1654) or its inverse.
```

### The `isDeclaredFunctionName` precedent D-07's `callgraph` node identity must reuse

```go
// Source: internal/compiler/corevalidate/corevalidate.go:359 (usage) / :1654 (definition)
if !v.check(!isDeclaredFunctionName(v.program.Functions, function.ID, function.ForeignContract.Symbol), "core.call_target_not_foreign", operation.ID) {
    return false
}
// isDeclaredFunctionName(functions []core.Function, excludeFunctionID, name string) bool
// — the exact name-collision precedence rule callgraph's node-identity
// derivation must reuse for the foreign-symbol-shadowing fixture (D-07
// specifics item 3).
```

### Diagnostic ID identity (why D-07-16's rotation is load-bearing)

```go
// Source: internal/compiler/diagnostic/diagnostic.go:104-112 (Error)
identity := struct {
    Schema string
    Code   string
    Span   Span
    Causes []Cause
}{Schema: Schema, Code: code, Span: span, Causes: causes}
encoded, _ := json.Marshal(identity)
sum := sha256.Sum256(encoded)
return Diagnostic{Schema: Schema, ID: "diagnostic:" + hex.EncodeToString(sum[:12]), ...}
// Causes is part of the hashed identity struct verbatim — an unrotated
// Causes slice (different order per discovery root) produces a different
// ID for the semantically identical cycle. Confirmed this session.
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|---------------|--------|
| `core.Interface`/`FunctionSignature` (`lang.interface/0`) carries origin + abilities only | `lang.interface/1` carries a total per-parameter ownership mode, a total return contract, `Callable`, `Fails`, closure-derived `ForeignReach`/`ClosureDigest` | This phase (Stage 0) | Every consumer of the summary after Phase 07 must decode `/1` to admit a call; `/0` stays decodable-but-never-admissible (frozen bytes) |
| No Lang-to-Lang call surface; `try`/`discard`-wrapped foreign calls are the only call-shaped RHS | Bare `let r = f(x)` to a Lang function is legal; the identical bare-call syntax to a foreign symbol remains refused, now at check-time instead of parse-time | This phase (Stage 1) | `syntax.fallible_call_not_consumed`'s enforcement layer moves; existing negative fixture `testdata/phase4/fallible_call_unconsumed.lang` must still be refused, but by a different code path |
| No call graph exists | `internal/compiler/callgraph`, iterative 3-color DFS, roots = all declared functions | This phase (Stage 2) | First whole-program-shaped structure in the compiler; explicitly excludes edges from the digest-bound summary (D-07-11) to avoid re-importing the rejected whole-program Design A |

**Deprecated/outdated:** none — this is additive; nothing existing is removed except the parser's blanket bare-call refusal, which narrows rather than disappears.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | The new call-admission logic belongs in `analyzeStraightLine`/`analyzeArmBody` rather than a wholly new top-level `checkLangCall` function called from `Program()` | Architecture Patterns, Pattern 1 | Low — this is an internal code-organization choice within `check`, reversible, and either shape satisfies SEM-04/06; flagged as the most natural fit given the existing `places` scope map, not verified against a spike |
| A2 | `ForeignReach`'s three fields should be plain strings (not an enumerated Go type), following the `ForeignContract.Unwind`/`NonlocalExit` precedent | Standard Stack / Claude's Discretion item in CONTEXT.md | Low — CONTEXT.md already leaves this to planner discretion; this is a recommendation grounded in an in-tree precedent (`core.go:63-64`), not a locked decision |
| A3 | `pgregory.net/rapid`'s STANDING-VERDICTS.md "Adopt" verdict is still current (not re-audited against the live registry this session) | Package Legitimacy Audit | Low — test-only, and the verdict is a standing project record; re-verify with `go list -m` if a plan actually adds it |

## Open Questions (RESOLVED — see 07-CONTEXT.md amendments and 07-REVIEWS.md)

> **All three questions below were resolved after this file was written.** Q1 and
> Q2 are answered by name inside `07-03-PLAN.md`; **Q3's recommendation below is
> now WRONG** and is superseded by **D-07-37**. Read the resolution markers, not
> the original recommendations.


1. **Where exactly does the relocated `syntax.fallible_call_not_consumed` refusal for a foreign callee live once the parser stops emitting it universally?**
   - What we know: the parser currently emits it unconditionally at `parser.go:404-421`; `check.go`'s `Program()` already builds `foreignSymbols`/`functionNames` maps that can distinguish the two cases.
   - What's unclear: whether the diagnostic *code* (`syntax.fallible_call_not_consumed`) should be preserved verbatim when re-emitted from `check` (keeping `testdata/phase4/fallible_call_unconsumed.lang`'s expected code unchanged, per `check_test.go:791`), or whether it becomes a new `core.*`/`check.*`-namespaced code now that it's a semantic rather than syntactic refusal.
   - Recommendation: preserve the existing code string for backward test compatibility (it is already asserted as a diagnostic *code*, not a diagnostic *source layer*, in `check_test.go`), and confirm this in Stage 1's first plan before touching the parser.
   - **RESOLVED** in `07-03-PLAN.md:201` — the recommendation was adopted. See also **D-07-40**: `TestFallibleCallUnconsumedRejected` (`syntax_test.go:1004`) and `native_test.go:1173` are deliberate planned edits, not incidental breakage.

2. **Does `resolveForeignStep`'s argument-shape check need to change at all, or does it stay exactly as-is (foreign-only) while a parallel, new predicate handles the Lang-call case?**
   - What we know: `resolveForeignStep` (`check.go:1316-1355`) is reachable only via `hasTryCall`-gated functions; a bare Lang call never reaches it.
   - What's unclear: whether CONTEXT.md's D-07-01 wording ("widens the existing predicate") was intended as a literal code-reuse instruction or a conceptual description of the *rule* (arity-1, in-scope-binding) being reused.
   - Recommendation: treat it as conceptual reuse only; `resolveForeignStep` stays untouched (it is correctly foreign-only), and the new Lang-call predicate is new code sharing the same *shape*, not the same function.
   - **RESOLVED** in `07-03-PLAN.md:429-431` — conceptual reuse only, as recommended.

3. **What is the minimal `ForeignReach`/`ClosureDigest` computation for Stage 0, given the language has no calls yet at the point Stage 0 lands?**
   - What we know: Stage 0 must widen every existing function's summary (all ~39 single-function Phase 1-4 corpus programs) with a total `ForeignReach` and `ClosureDigest` before any `OpCall` exists.
   - What's unclear: for a function with zero callees (every function in the Phase 1-4 corpus), is `ClosureDigest` simply `hash(own signature)` with an empty callee-digest list, or is there a reserved sentinel?
   - ~~Recommendation: `ClosureDigest = hash(own FunctionSignature bytes)`~~ — **SUPERSEDED BY D-07-37. This recommendation is self-referential and must not be implemented:** the "own `FunctionSignature` bytes" include the `ClosureDigest` field itself. The cross-AI review (codex, HIGH) caught this. The locked definition is a canonical preimage — the `/1` signature **with `ClosureDigest` zeroed**, prefixed with a domain separator, followed by callee `(ID, ClosureDigest)` pairs **sorted by ID**; the zero-callee base case is that preimage with an empty callee list. Implemented in `07-01-PLAN.md` Task 3. Note also **D-07-38**: the chaining arm is deliberately deferred to `07-08-PLAN.md`, in a wave after cycle refusal, because the chain terminates only on a DAG.

## Environment Availability

Skipped — this phase has no external tool/service dependencies beyond the existing Go toolchain (`go 1.24`, confirmed via `go.mod`) and Clang/native toolchain already required by prior phases (unchanged by Phase 07; cgen's real emission work is out of scope this phase per Pattern 2/Pitfall 4).

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` (`go test ./...`) — no external test framework in this repo |
| Config file | none — table-driven Go tests, fixtures under `testdata/phaseN/` |
| Quick run command | `go test ./internal/compiler/check/... ./internal/compiler/corevalidate/... ./internal/compiler/core/...` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| SEM-04 | `OpCall` handled at all six dispatch sites, both exhaustive-dispatch controls green | unit + CLI | `go test ./internal/compiler/core/... -run TestAllOperationKindsHandledAtEverySite` + `go test ./internal/compiler/session/... -run TestVerify` (lane assertions) | ❌ Wave 0 — needs new phase-07 fixture(s) added to both lists |
| SEM-05 | Digest-bound `lang.interface/1` carries everything a caller needs; no body read | unit | new `originvalidate`/`corevalidate` tests over the Stage 0 gate corpus | ❌ Wave 0 |
| SEM-06 | callable ⊆ publishable; refusal has a stable code | unit | new `check`/`originvalidate` test using the (currently missing) clean-but-uncallable fixture | ❌ Wave 0 — fixture must be authored first (Pitfall 3) |
| SEM-07 | Direct/mutual/indirect cycles refused, bounded, never hang | unit | new `internal/compiler/callgraph` tests + `check`-level integration tests over diamond/shared-leaf and chain corpora | ❌ Wave 0 — package does not exist yet |
| QLT-08 | Every new interprocedural control mutation-killed in the introducing plan | unit, `TestXMutationKilled`/`TestXMutationMatrix` naming convention | project-wide convention, e.g. `go test ./internal/compiler/callgraph/... -run MutationKilled` | ❌ Wave 0 — no `qlt08`-scoped registry exists (the existing `qlt01_registry.json`/`qlt01.go`/`qlt01_test.go` in `internal/compiler/session/` is a **different**, M001-era spike-descendant registry, not reusable for QLT-08's per-control mutation-kill discipline) |

### Sampling Rate

- **Per task commit:** targeted package tests for the file(s) touched (`go test ./internal/compiler/<package>/...`)
- **Per wave merge:** `go test ./...` (full suite; the project has no separate slow/fast test tiers documented)
- **Phase gate:** full suite green, plus both exhaustive-dispatch controls (in-process `core_test.go` and CLI-observable `lane:kind-exhaustive-dispatch`-equivalent for Phase 07) before `/gsd-verify-work`

### Wave 0 Gaps

- [ ] `testdata/phase07/*.lang` — two-function call corpus: basic call, diamond+shared-leaves (parser-shaped, per Pitfall 2), self-call, unreachable cycle, foreign-symbol-shadowing, unresolvable-callee-forged-artifact, both-match-arms-call (D-07-28 `Result`-blindness guard)
- [ ] a clean-but-uncallable fixture for SEM-06 — **CORRECTED (D-07-44): EXTRACT** the existing inline witness at `check/check_exclusive_test.go:46-58`; do not author a new one (see Pitfall 3)
- [ ] the `relay`/`escort` two-function dangling-alias witness referenced by D-04-03 — no file exists, **and D-07-44 adds:** the recorded source uses nested `relay(borrow mut buffer)`, which **D-07-01 has made ungrammatical**. It must be deliberately converted to A-normal form and shown semantically equivalent (`07-05-PLAN.md` Task 2), never copied verbatim.
- [ ] `internal/compiler/callgraph/callgraph.go` + `callgraph_test.go` — the package itself does not exist
- [ ] a new phase-07-scoped block in `session.go` mirroring `:2501-2568`, with its own `dispatchFixtures` list and required-kinds check
- [ ] Framework install: none — stdlib `testing` already covers everything; `pgregory.net/rapid` add is optional/test-only if a plan wants property-based synthetic-`core.Program` generation for the `corevalidate` peer's mutation matrix (D-07-19)

## Security Domain

`security_enforcement` is enabled in `.planning/config.json` (`security_asvs_level: 1`, absent-is-enabled rule also applies). This phase's ASVS-relevant surface is narrow: it is a compiler admission boundary, not a network-facing or user-auth surface.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | not applicable — compiler has no auth surface |
| V3 Session Management | no | not applicable |
| V4 Access Control | yes (analogously) | `callable ⊆ publishable` (SEM-06) is the project's own access-control-shaped invariant: a caller must not admit a call to a non-publishable target. Standard control: fail-closed refusal with a stable diagnostic code (`core.*` namespace), never a silent drop |
| V5 Input Validation | yes | every new `core.Program`/`core.Interface` field the parser/checker/validators consume must be validated fail-closed (no `omitempty`-style silent absence tolerated per D-07-08/D-07-09's "absence must be an error" rule) |
| V6 Cryptography | yes (narrow) | `ClosureDigest`/`CoreDigest` are unkeyed SHA-256 content digests (`sha256.Sum256`, already in use at `originvalidate.go:399-402`) — explicitly integrity-only, never authenticity (D-07-13); do not introduce keyed/HMAC digests this phase, that is out of scope |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Coordinated producer-and-peer lie (a single implementation self-certifying) | Repudiation / Tampering | Independent re-derivation, never a shared `callsummary` package (D-07-20/D-07-22), enforced by a static import-independence test asserting `corevalidate` does not import `originvalidate` |
| Forged artifact naming a nonexistent or unresolvable callee | Tampering | Must refuse through the existing `core.unknown_*`/`core.call_target_not_foreign` path, never silently dropped from the call graph (a dropped edge is how a cycle escapes detection) |
| Diagnostic-ID instability enabling downstream dedup/repair confusion | Tampering (of derived trust artifacts) | Canonical rotation of cycle-membership Causes before hashing (D-07-16), with a required test asserting one identical ID across two discovery-root orderings |
| Unbounded diagnostic output as a DoS vector on a dense/adversarial call graph | Denial of Service | Dedupe edges at graph-build time (bounds E by V²), bound the diagnostic (32 `cycle_member` causes max, `truncated:core.call_cycle_bound`), never bound the traversal itself (traversal stays O(V+E)) |

## Sources

### Primary (HIGH confidence — read this session)
- `internal/compiler/core/core.go` (full file) — `Interface`, `InterfaceSchema`, `FunctionSignature`, `Parameter`, `OperationKind`, `AllOperationKinds`, `TerminatorKinds`
- `internal/compiler/check/check.go` (targeted reads: `:1,150`, `:297-330`, `:784-900`, `:1018-1125`, `:1254-1432`, `:1499-1520`, `:1804-1900`) — dispatch, `hasTryCall` gating, `resolveForeignStep`, `analyzeStraightLine`
- `internal/compiler/corevalidate/corevalidate.go` (function-signature scan + `:340-370`, `:890-1015`) — both replay switches, `isDeclaredFunctionName`
- `internal/compiler/originvalidate/originvalidate.go` (full file) — `BuildInterface`, `ValidatePublished`, `CheckSummary`, `FunctionAnswer`
- `internal/compiler/pathoracle/pathoracle.go` (`:1-130`) — package doc, `backEdgeError`, `pathCapError`
- `internal/compiler/interp/interp.go` (`:60-150`, function-signature scan) — `runBranchArm`'s documented-unreachable-kind precedent
- `internal/compiler/cgen/cgen.go` (function-signature scan) — four switch sites, `len(Functions) != 1` guard lines
- `internal/compiler/diagnostic/diagnostic.go` (`:1-140`) — `Cause`, `Diagnostic.ID` identity derivation
- `internal/compiler/protocol/protocol.go` (`:100-200`) — `InterfaceSummary`/`InterfaceFunctionAnswer`, `TraceSummary`'s truncation convention
- `internal/compiler/debugmap/debugmap.go` (`:30-70`) — confirmed `InterfaceSchema` mention is a comment, not a functional site
- `internal/compiler/syntax/parser.go` (`:370-495`) — `linearBody`, `tryCallBinding`, the bare-call refusal
- `internal/compiler/ast/*.go` (`RHS`, `Binding`, `LinearBody` struct definitions)
- `internal/compiler/core/core_test.go` (`:200-360`) — `TestAllOperationKindsRegistered`, `TestAllOperationKindsHandledAtEverySite`
- `internal/compiler/session/session.go` (`:1020-1050`, `:2480-2580`) — `InterfaceExportCommandFile`, `lane:kind-exhaustive-dispatch`
- `internal/compiler/session/qlt01.go`, `qlt01_registry.json` — confirmed this is a distinct, M001-era spike-descendant registry, not QLT-08's mechanism
- `go.mod` — confirmed `pgregory.net/rapid` absent
- Filesystem searches (`find`, `grep -rl`) confirming `testdata/phase3/exclusive_borrow_clean` and any `relay`/`escort` two-function fixture do not exist anywhere in the tree
- `.planning/phases/07-.../07-CONTEXT.md`, `.planning/REQUIREMENTS.md`, `.planning/STATE.md`, `.planning/STANDING-VERDICTS.md`, `.planning/LANGUAGE-MATURITY.md`, `.planning/milestones/M001-phases/04-.../04-CONTEXT.md`, `.planning/config.json` — read in full this session

### Secondary (MEDIUM confidence)
- None used — this phase's research surface was entirely in-repo; no web/docs lookups were needed or performed (no external library research applies to this phase's scope).

### Tertiary (LOW confidence)
- None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — no new production dependencies; all structural claims read from source this session
- Architecture: HIGH for existing-code claims (file:line verified); MEDIUM for the *placement* recommendation (new call-admission code in `analyzeStraightLine`/`analyzeArmBody` vs. a wholly new function) — reversible, low-risk either way
- Pitfalls: HIGH — each pitfall is grounded in a specific, quoted, in-tree fact (parser refusal, missing fixtures, cgen gating), not speculation

**Research date:** 2026-09-08
**Valid until:** effectively the life of Phase 07's planning — this is milestone-scoped, in-repo structural research tied to the exact commit this session read; re-verify file:line citations if the tree changes materially before planning begins (e.g., after Stage 0 lands, before Stage 1 planning starts).
