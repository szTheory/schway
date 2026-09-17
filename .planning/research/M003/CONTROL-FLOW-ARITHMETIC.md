# M003 Research: Control Flow and Arithmetic

**Researched:** 2026-09-17
**Mode:** Feasibility + scope (single decision point)
**Overall confidence:** HIGH on codebase claims (every one is a path+line I read), MEDIUM-HIGH on the scope verdict.

---

## Verdict (one-shot recommendation)

**Yes, M003 should do arithmetic and `if` — and must explicitly *not* do loops.** Arithmetic and `if` are cheap relative to their payoff because they require **zero change to the CFG shape**: `if` desugars into the arm-terminating two-block topology `check`/`cgen`/`interp` already build for `match`, and arithmetic is a pure new-`OperationKind` addition of exactly the shape the project has now executed twice (`OpCall` in M002, `OpConstructPayload`/`OpDestructurePayload` in Phase 12). Loops are categorically different: the acyclic-CFG invariant is independently, deliberately, fail-closed enforced in at least three derivers (`check.cfg_back_edge` at `check.go:2734`, `pathoracle.cfg_back_edge` at `pathoracle.go:243`, `corevalidate.checkCallGraphAcyclic` at `corevalidate.go:481`), the trusted oracle `pathoracle` is a **path enumerator** (`EnumeratePaths`, `pathoracle.go:412`, `MaxPaths = 4096`) rather than a fixpoint and has no meaning under a back edge, `loanLivenessBound`'s own doc comment (`check.go:2634-2650`) names iteration as the thing that invalidates it, and affine soundness across a back edge is the single hardest problem in the space (Austral refuses it outright; Rust needs drop flags). Three further payoffs make arithmetic the *highest-leverage* M003 content rather than merely a nice addition: (1) a match arm's returned value in native is currently the **compile-time literal `arm.Pattern`** (`cgen.go:1902` → `cgen.go:2021`), so arithmetic inside an arm is the first thing that makes a branch produce a computed runtime value — which is precisely what makes the ratified-unconstructible **D-12-43** value-divergence control constructible; (2) any comparison operator produces `Bool` from `Byte`, which forces `sameType(ReturnType, Parameter.Type)` to lift and thereby closes the **DX-06 / DX-07 / D-13-02b / D-13-10a** cluster as a *prerequisite* rather than a parallel task; (3) arithmetic adds O(1) work per operation and is therefore neutral against the `recomputed_work_growth_exponent` **hard** gate of 1.2 in `internal/compiler/session/qlt02_budget_manifest.json`, which loops would not be. Scope: **one milestone, six phases, two new `OperationKind`s and no more.**

---

## Current state, grounded

Everything below was read, not assumed.

### The CFG is real but structurally acyclic by refusal

| Fact | Evidence |
|---|---|
| `core.Block` = `{ID, PointID, OperationIDs, Successors []string}` | `internal/compiler/core/core.go:918-923` |
| `core.Edge` = `{ID, FromBlockID, ToBlockID, Pattern}` — keyed by **match arm pattern** | `core.go:927-931` |
| `Blocks`/`Edges`/`LoanEndpoints` are `omitempty`; a straight-line body never populates them | `core.go:902-912` |
| Only a `match` with arm bodies produces blocks (`Match.HasBlocks()`) | `core.go:975-982` |
| Every arm block has **exactly one successor (the join)** today | `materializeLoanEndpoints` doc comment, `check.go:2800-2810` |
| Cycles in the block graph are refused fail-closed by an iterative three-colour DFS | `detectCFGCycle`, `check.go:2570-2609`; called at `check.go:2733` |
| The refusal is a coded diagnostic, non-repairable, whole-topology | `cfgBackEdgeDiagnostic`, `check.go:2623-2629` |
| A **second, independent** back-edge refusal exists in the trusted oracle | `pathoracle.backEdgeError`, `pathoracle.go:130-143`, raised at `:243` |
| A **third** acyclicity derivation exists for the call graph in the non-importing peer | `corevalidate.checkCallGraphAcyclic`, `corevalidate.go:450-481` |

So: `loanLivenessFixpoint` **is** a backward monotone worklist over a finite lattice (`check.go:2722-2790`) and the *algorithm* would converge on a cyclic CFG. But the CFG it runs on is acyclic **by construction and by triple refusal**, and the analysis is only correct-by-accident in a way three separate load-bearing details depend on.

### The three details a back edge breaks

1. **`loanLivenessBound = 4 * blockCount * (distinctLoanCount + 1)`** (`check.go:2656-2658`, factor at `:2653`). Its own comment says this is *"a PLANNER ASSUMPTION (08-04-PLAN.md OWN-06/precision) that must be re-opened if a later milestone makes block or loan counts program-controlled (iteration, loops, collections)"* (`check.go:2645-2650`). With a back edge a block can be re-queued more than a constant number of times; the true worklist bound is O(blocks × lattice height) = O(blocks × loans). Factor 4 becomes a **false-positive refusal**, not a safety net.
2. **`materializeLoanEndpoints`** (`check.go:2812-2872`) classifies each loan endpoint as `point` (last use inside a block) or `edge` (at a genuine successor divergence). Both notions presuppose a DAG. A loan born before a loop and used inside it is live on the back edge and has **no single last use** — the Rust "reborrow inside a loop" shape.
3. **`analyzeArmBody`'s soundness argument is explicitly mutual exclusion**: *"one arm's move can never be observed as a false use-after-move by a sibling arm that never runs at the same time (mutually exclusive control flow — the two arms' places never collide because their IDs are distinct global ordinals)"* (`check.go:2887-2891`). A back edge is exactly the construct that makes a block run *again alongside itself*. This is not a bound to retune; it is the argument itself.

### `pathoracle` is an enumerator, not a fixpoint — the deepest obstacle

`EnumeratePaths(functionID, entryBlockID, idx, MaxPaths)` (`pathoracle.go:412`, `pathoracle_compose.go:211`) enumerates **acyclic entry-to-return paths**, capped at `MaxPaths = 4096` (`pathoracle.go:67`), with interprocedural splicing capped at `MaxCompositionDepth = 3` (`pathoracle_compose.go:85`). Its own comment says MaxPaths *"never fires on any program the parser can produce today"* (real reachable max: 64 = `maxArmsPerMatch`). Under a loop the set of *acyclic path shapes* stays finite but the set of *executions* does not — so the oracle's claim would silently weaken from "all executions" to "all acyclic path shapes" **without any code failing**. That is a semantics downgrade to the project's trusted oracle, smuggled in. Migrating `pathoracle` from enumeration to a fixpoint is the single largest cost item in loops, and it is invisible in a naive estimate.

### Arithmetic does not exist at any layer — four distinct absences

| Layer | Current state | Evidence |
|---|---|---|
| **Lexer** | No numeric-literal path at all. Digits are only accepted as identifier *continuation* characters (`unicode.IsDigit` at `lexer.go:106`). A leading `1` falls through `punctuation()` (`lexer.go:130-163`), gets width 0, and becomes `TokenUnknown` + `syntax.unexpected_byte`. | `lexer.go:101-124`, `:130-163` |
| **Tokens** | No `+ - * / % < >=` etc. Punctuation set is exactly `-> => { } ( ) : . \| =` and `< >` (used for type args). | `token.go:30-43`, `lexer.go:130-163` |
| **AST** | Fully **name-based** — no expression tree. `LinearBody.Result` is a `string`; `RHS{Kind, Source string, Callee string, Arguments []string}`. | `ast` pkg, `LinearBody` at `:150-159`, `RHS` at `:167-181` |
| **Core IR** | Twelve `OperationKind`s, none of them constant or arithmetic. Every operation *reads an initialized place* (`SourceID`); there is no literal operand anywhere. | `core.go:629-683`, `AllOperationKinds()` at `core.go:820` |
| **Interpreter** | The value domain is **`struct { tag, payload string }`** (`interp.go:368-370`). There is no numeric domain in the oracle. | `interp.go:355-370` |
| **Native** | `Byte` → `unsigned char`; `Buffer` → `struct { unsigned char bytes[4]; size_t length; }`. Runtime values are printed via `lang_write_byte(unsigned char)`. | `cgen.go:345`, `:1574`, `cgen_program.go:206-208` |
| **Types** | Exactly two value types, hardcoded by string in three packages. | `cgen_program.go:206-208`, `cgen.go:344`, `evidence.go:287-291` |

### The finding that reframes the whole question

In the **branch** emitter, a match arm's returned value is written as a **compile-time literal**:

```go
// cgen.go:1902  — returnLiteral is bound to arm.Pattern
emitBranchOperations(..., typeName, arm.Pattern)
// cgen.go:2021  — and written verbatim
fmt.Fprintf(out, "      if (!lang_write_json_string(%s)) return 74;\n", strconv.Quote(returnLiteral))
```

Contrast the **straight-line** emitters, which write the real runtime value: `lang_write_byte(locals[source.ID])` at `cgen.go:399`, `:724`, `:911`, `:1337`.

This *is* D-12-43 ("the current grammar never lets a match arm's result expose raw payload bytes — so a wrong-slot write is invisible to every comparator axis by construction"). **Arithmetic inside a branch arm is the minimal construct that makes a branch return a computed runtime value, and therefore the minimal construct that turns the five-axis comparator's value axis from vacuous to live on branching programs.** That converts a ratified "unconstructible" finding into a closable one.

### Scale of the corpus today

115 `.lang` files, 4,189 lines total (verified: `find . -name '*.lang' -not -path './.git/*' | wc -l`). Bounds: `maxFunctions = 1024`, `MaxTokens = 1<<17` (`parser.go:14-16`), `maxArmsPerMatch = 64`, `interp.MaxCallDepth = 128` (`interp.go:32`). The ownership differential is `TestOwnershipSequenceExhaustive`, 2 sweeps × (48⁰+48¹+48²+48³) = **225,890 cases**, with a derived case counter that fails if the generator silently stops early (`check_test.go:95-160`).

---

## Sequencing options

| Option | Surface added | Assurance cost | Unblocks | Verdict |
|---|---|---|---|---|
| **A. Arithmetic only** (literals + one integer type + binary ops, A-normal form) | numeric token, `OpConst`, `OpBinary`, one numeric type, interp numeric domain | Two new `OperationKind`s × 6 dispatch sites × D-04-22's "8-10 independent edits" each. **No CFG change.** Bound/endpoint/arm logic untouched. New: overflow semantics + the C-emission UB question. | Real computation; D-12-43 becomes constructible *if* the arithmetic can sit inside an arm | **Take** |
| **B. `if` only** (Bool + `if/else`, desugared to a two-arm match) | `Bool` builtin data type, `if`/`else` tokens, desugar in the parser | Near-zero **if arms terminate**, because the lowered topology is byte-identically today's arm topology (`Match.HasBlocks()`, one successor to a join). No phi, no drop flags, no merge rule. Needs a scrutinee producer, i.e. a comparison operator, i.e. option A. | Nothing on its own — with no comparison operator there is no way to *produce* a `Bool` | **Take, but strictly after A** |
| **C. Arithmetic + `if`** | A ∪ B, plus comparison operators | A's cost plus the `sameType(ReturnType, Parameter.Type)` lift (a comparison is `Byte -> Bool`) | D-12-43, DX-06, DX-07, D-13-02b, D-13-10a. First Lang program that computes something and chooses on the result. | **This is M003** |
| **D. Loops** | back edge, loop-carried values | `loanLivenessBound` re-derivation; `materializeLoanEndpoints` redesign; `analyzeArmBody`'s mutual-exclusion argument invalidated; `pathoracle` enumeration → fixpoint migration; three independent cycle refusals must be *selectively* relaxed (relaxing three independently-written derivations in lockstep is exactly the "two coexisting laws" anti-pattern in STANDING-VERDICTS); interp needs a fuel budget; native needs a matching budget or the comparator diverges on non-termination; growth-exponent hard gate at risk | Real programs | **M004** |
| **E. A numeric tower** (signed + unsigned, multiple widths, float) | many types, conversions, promotion rules | Signed overflow UB; IEEE754 cross-tier determinism (FMA contraction, x87 excess precision, `-ffast-math` adjacency); combinatorial conversion matrix | Nothing M003 needs | **Reject for M003** |

**Answers to the sub-questions.**

1. **Order.** Arithmetic → comparison → `if`. Not the other way: `if` with no way to produce a `Bool` is inert, and an inert feature is exactly the DX-06 failure mode this project just paid for.
2. **Can any be deferred?** Loops, yes and must be. Arithmetic, no — it is the forcing function for three open debt items.
3. **Is `if` cheap given `match`?** *Yes, conditionally on one design decision: `if` arms must terminate* (each arm ends in `OpReturn`/`OpFail`/`OpDefect`, per `TerminatorKinds()` at `core.go:825`), exactly as match arms do today. Under that constraint `if` is a **parser desugaring plus a `Bool` data declaration** and costs essentially nothing in the IR, the checker, the validators, the oracle, or cgen. The moment `if` arms are allowed to *fall through to shared code*, you acquire a value-merge (phi or block parameters) and an ownership-state merge (Austral Rule 3 or Rust drop flags) simultaneously — which is a different, much larger feature. **Do not let "if" quietly become "if with a join" during planning.** That is the single highest-risk scope creep in this milestone.
4. **Does arithmetic need a numeric tower?** No. **One fixed-width unsigned integer type is enough for M003**, and is strictly better than two: it makes signed-overflow UB *unconstructible* rather than merely handled (see below).

---

## Recommended scope for M003

Six phases. Ordered by what unblocks what, and by the adversarial finding that the emitter cluster must come first.

| # | Phase | Why here | Closes |
|---|---|---|---|
| **P14** | **Return type may differ from parameter type** | Prerequisite: a comparison operator is `Byte -> Bool`, which is unrepresentable under `sameType(ReturnType, Parameter.Type)` (`check.go:255`, `:3148`, `:3399`). Doing this first also makes the *already-built* `resolveBlame` (`check.go:769`, zero call sites) reachable without writing new code. | DX-06, DX-07, D-13-02b, D-13-10a |
| **P15** | **Retire the six single-function emitters** | The three-deferral rule (D-10-60) forbids a third deferral, and D-12-36 already stated the reversal. Do it **before** adding a new operation kind to emission, or the new arithmetic case must be written six times. `cgen.Emit`/`EmitNative`/`emitLinear`/`emitBranchOperations` all hard-fail on `len(Functions) != 1`. | D-11-02, D-12-36, D-10-60 |
| **P16** | **Numeric literals + one integer type + `OpConst`** | First numeric token in the lexer, first numeric domain in the interpreter oracle (`interp.value` is `{tag, payload string}` today — widen it deliberately, following D-12-17/D-12-18's precedent of a single widened struct with a byte-identical projection for the existing case). One new `OperationKind` at six dispatch sites. | — |
| **P17** | **`OpBinary`: arithmetic + comparison, with a decided overflow law** | Second and final new `OperationKind`. Includes the C-emission width/cast discipline and the defect-on-overflow terminal. | — |
| **P18** | **`Bool` + `if`, arm-terminating; computed branch return values at five axes** | The desugar is cheap; the *evidence* is the phase. This is where the branch emitter stops writing `arm.Pattern` as a literal and starts writing a computed value. | **D-12-43** |
| **P19** | **Arithmetic evidence + budget re-ratification + Nyquist close** | Edge-value bounded enumeration, mutation controls on the overflow check, re-ratify the growth-exponent gate, and `/gsd-validate-phase 07 08 11 12 13`. | Nyquist debt |

### Explicitly deferred to M004 (design committed now, work not scheduled)

- Loops of any kind, and with them: `loanLivenessBound` re-derivation, `materializeLoanEndpoints` redesign, the three coordinated back-edge-refusal relaxations, `pathoracle`'s enumeration→fixpoint migration, interpreter fuel, native step budgets.
- Signed integers, multiple widths, floats, saturating operators.
- Any `if` that falls through to a shared join, and therefore phi/block-parameters and drop flags.
- Recursion (already out of scope per PROJECT.md; a bounded-recursion admission design and a loop design should be researched together, since they are the same termination question).

### Explicitly *not* in M003 either

- Event identity (D-11-51 → D-12-21). It is a *function*-identity collision in shared-leaf diamonds; arithmetic multiplies operations per function, not functions, so it does not worsen. It deserves an owner but it is not coupled to this milestone and adding it would make M003 a four-cluster milestone. Give it to M004 or a dedicated slot.
- `testdata/phase6` held-out pair repair (D-13-34) — cheap, but unrelated; fold into P19 only if it is genuinely a few hours.

---

## Integer semantics design

This is where the project can most easily give away its own guarantee, so the design is stated as rules, not options.

### Rule 1 — Unsigned only. One width. No signed integers in M003.

Signed overflow is UB in C and LLVM will exploit it (this is the documented purpose of `-fwrapv`, and of `-fsanitize=signed-integer-overflow`, whose interaction with `-fwrapv` is itself a known landmine: since GCC 8, `-fwrapv` effectively disables the UBSan overflow check — and this project **runs an ASan/UBSan lane**, so a `-fwrapv`-based mitigation would silently blind an existing evidence lane). The correct move for a project whose style is "make the hazard structurally impossible, then police it with a control" is: **there is no signed integer type in M003.** Unsigned arithmetic in C is defined as modulo 2^N by the standard, unconditionally.

Recommended type: `U64`, lowered to `uint64_t` (`<stdint.h>`). Keep `Byte` as a *byte*, not a number — do not overload it.

### Rule 2 — Defeat integer promotion explicitly, and prove it with a control.

`unsigned char` and `unsigned short` **promote to `int`** in C. At `unsigned char` width this happens to be harmless (`a+b ≤ 510`, `a*b ≤ 65025`, both far inside `int`), but at `uint16_t` a multiply can reach ~4.29e9 and overflow `int` — **signed UB reached from purely unsigned source types.** This is the canonical trap and it is width-dependent, which makes it exactly the kind of thing that survives review.

The rule: **cgen emits an explicit cast on every operand and on the result**, never relying on promotion:

```c
/* U64 add, wrapping family */
uint64_t t3 = (uint64_t)((uint64_t)t1 + (uint64_t)t2);
```

And a **non-importing structural control** (same posture as `commentSafeForeignField`/`validForeignCType`, `cgen.go:1443`, `:1470`): scan the emitted C for any arithmetic expression whose operands are not explicitly cast to the declared unsigned width, and for any occurrence of a signed integer type in an arithmetic context. Make the control fail under a fault-injection seam that removes one cast (the `loanLivenessBoundSeam` / `SetDisableEmptyTagSerializationForTest` pattern, `check.go:2684`, `interp.go:391`).

### Rule 3 — Two operator families, no silent default.

Follow Zig's explicit-overflow stance, reduced from three families to two:

| Family | Spelling | Semantics | C lowering | Interp |
|---|---|---|---|---|
| **Checked** (default) | `+ - *` | On overflow: terminal **defect** with a named reason | explicit pre-check (`if (b != 0u && a > UINT64_MAX / b) lang_defect("...")`), then the cast-wrapped op | compute in `uint64`, check, emit the same defect outcome |
| **Wrapping** | `+% -% *%` | Total; two's-complement / modular | the cast-wrapped op, no check | `uint64` wrap |

Zig's saturating family (`+| -| *|`) is real prior art but adds a third semantics with no M003 demand. Defer.

**Why defect rather than wrap-by-default:** PROJECT.md's constraint is "Safe code has defined behavior" *and* the whole project is built on named refusals over silent coercion. Silent wrap on `+` is defined but wrong-answer-shaped. Defect is already a first-class, five-axis-comparable outcome: `OpDefect` (`core.go:650-656`), the `"defect"` outcome kind with an empty value and a `function.defected` event carrying the reason (`cgen.go:2050-2072`), and a `_Noreturn lang_defect` that is the single documented exemption from the zero-attribute control. **Reusing it costs almost nothing and immediately gives arithmetic a comparator-visible failure channel.**

**Why not `__builtin_*_overflow`:** it works, but the project emits *readable* C and derives optimizer attributes from checked facts; an explicit pre-check is auditable by a human reviewer and by the structural control in Rule 2.

### Rule 4 — Division and modulo: divisor zero is a defect.

`/` and `%` with a zero divisor → the same terminal defect, checked before the division at every tier. In C, integer division by zero is UB; there is no unsigned escape. WebAssembly's choice here is a **trap** on `idiv_u` by zero, which is the same shape as a defect. (Signed `INT_MIN / -1` overflow-trap does not arise: Rule 1 removes signed types.)

### Rule 5 — Shifts: out-of-range shift amount is a defect, not a mask.

C makes a shift by ≥ width UB. WebAssembly **masks the shift amount modulo N** and is total. Masking is cheaper and equally defined — but it is a silent wrong answer, and this project consistently prefers a named refusal. Take the defect. Record the Wasm alternative in the decision so a future milestone can revisit for a performance reason rather than re-deriving the question.

Right-shift is logical (no signed types ⇒ no arithmetic-shift ambiguity, which is exactly Rule 1 paying off a second time).

### Rule 6 — The interpreter oracle and the native tiers agree by *shared derivation*, not by coincidence.

Every semantic decision above must be derived from **one shared table in `core`**, in the style of `core.AlternativeNameForPayloadType` (the "single shared derivation read by `check`, `interp`, and `cgen`" precedent that Phase 12 established, cited in PROJECT.md's validated-requirements list). Concretely: one `core` function mapping an operator to `{arity, operand type, result type, overflow policy, identity/edge values}`, read by all six dispatch sites. Never a per-site `switch` on operator spelling — that is exactly how the interpreter stops being an oracle.

### Rule 7 — Float is deferred, and say why in the decision record.

IEEE-754 bit-exact agreement between a Go-hosted interpreter and Clang at `-O3 -flto` is achievable but requires actively suppressing FMA contraction (`-ffp-contract=off`), guaranteeing no excess precision, and matching Go's own float semantics — and the payoff for an AI agent writing its first real Lang program is near zero. Defer, named.

---

## Termination and the interpreter oracle

M003's answer is the strong one: **do not create the problem.** With no loops and no recursion (call-graph cycles refused by name, PROJECT.md validated requirement SEM-07), every admitted program terminates *structurally*, the interpreter remains a total function, and `pathoracle` keeps enumerating a complete, finite set of executions. Arithmetic does not change this: a checked-overflow defect is a terminal outcome, not a non-termination.

**Commit the M004 posture now, so it constrains M003's IR choices:**

1. **Structured control flow only, WebAssembly-style.** No `goto`, no arbitrary branch targets. In Wasm, *"labels can only be referenced from within the associated structured control instruction … branches can only be directed outwards"*; a branch to a `block` is a `break`, a branch to a `loop` is a `continue`. This makes reducibility a **syntactic** property, so "does this CFG have an irreducible back edge" is never a question an analysis has to answer. For a C17-emitting backend this is doubly right: structured Lang control flow maps to structured C control flow, so the emitted C stays readable and no `goto`-based state machine is ever needed.
2. **Counted loops first, general `while` later or never.** A counted loop (`for i in 0..n`, trip count a runtime `U64`) is a Dafny `decreases` clause that is *inferred trivially because the counter is the variant*. Termination then needs no proof obligation, no fuel, no admission rule — it is a consequence of the construct. This is the single design choice that keeps the interpreter total and the comparator honest under loops.
3. **If a general `while` is ever admitted, both tiers need a budget and the budget must be comparator data.** The precedent already exists and is well-designed: `interp.MaxCallDepth = 128` produces `callDepthExceededDefectReason` as *"comparable execution DATA — modeled like the existing core.OpDefect arm … never a bare Go error (D-10-25) — because Phase 11's five-axis comparator needs the refusal as data when diffing against native's structurally different limit-hit behavior"* (`interp.go:24-40`). A step budget must follow that exact shape. **But note the cost the comparator imposes:** a step counter in native is an observable side effect the optimizer must preserve, so `-O3 -flto` would carry it. That is a real per-operation cost against the compute-efficiency constitution, and it is the strongest argument for counted loops over `while`.
4. **`pathoracle` must move from enumeration to fixpoint, explicitly and with a stated weakening.** Under loops, `EnumeratePaths` no longer enumerates executions. Migrating it silently would downgrade the project's trusted oracle without any test going red. Name this as M004's largest single line item.

---

## Ownership × control flow: the hard cases

These are M004's problems. They are listed here because **the M003 IR decisions must not foreclose the good answers.**

### Case 1 — Move inside one branch but not the other (conditional move)

Two established answers:

- **Austral, Rule 3:** *"linear variables defined outside an `if` statement must be used consistently in all branches: they must either be consumed in every branch or appear zero times in every branch."* The checker runs per-branch and compares the resulting state tables; a mismatch is a compile error. **No runtime state.**
- **Rust:** allows the inconsistency and inserts **drop flags** — a runtime bit per conditionally-moved local, elaborated in MIR (`DropIfInit` → drop elaboration).

**Recommendation: Austral Rule 3, with a named diagnostic code.** Drop flags introduce per-value runtime state whose value must then be proven identical across interpreter / `-O0` / `-O3` / `-flto` — the five-axis comparator would have to carry a whole new class of runtime fact, and cgen would acquire dynamic drop state it has never had. Refusing the program is cheaper, more auditable, and matches the project's "named refusal" posture. Austral's own case for this is simplicity: its entire linearity checker is under a thousand lines.

**M003 sidesteps this entirely** because `if` arms terminate — there is no state to merge.

### Case 2 — Move out of a loop body

**Austral, Rule 5: *"linear variables defined outside a loop cannot be consumed inside the loop."*** The reasoning is exact: it would be consumed on the first iteration and consumed-again on the second. A value may be created *and* consumed within one iteration; each iteration gets its own.

**Recommendation: adopt Austral Rule 5 verbatim as M004's default**, with one escape hatch that is a *typing* rule rather than a dataflow rule — see Case 4.

### Case 3 — A loan live across a back edge

Rust's answer is a fixpoint: *"the set of in-scope loans at each point is found via a fixed-point dataflow computation"* (RFC 2094), with lifetimes as sets of program points grown to a fixpoint, plus false unwind edges so every infinite loop still has an exit node so drop-liveness stays well-defined. Reborrows inside a loop are the characteristic hard pattern.

Concretely for this codebase: the *fixpoint* is fine (it is already a monotone backward worklist over a finite lattice, `check.go:2722`). What must change is (a) the bound, re-derived as O(blocks × loans) rather than 4×blocks×(loans+1), and (b) `materializeLoanEndpoints`, whose point/edge dichotomy has no meaning for a loan live on a back edge. A third endpoint kind — or a decision that a loan live on a back edge is simply refused in M004 — is the design choice.

Heed the named precedent in LANGUAGE-MATURITY.md: **Polonius**. Rust's Datalog borrow-checker prototype *"doesn't scale and has no path to stabilization"*; the current native rewrite reports a worst-case 2-3× slowdown. Cost-scaling before generalizing is a standing project rule (STANDING-VERDICTS, process anti-patterns) and this is the exact shape it names.

### Case 4 — Loop-carried ownership state: the one design choice that matters

Instead of asking "was this owned value consumed on some iteration?" (a dataflow question, and the hard one), make the loop-carried value an **explicit block/loop parameter**. Then a loop body that consumes an owned value must *produce a replacement of the same type* to pass to the next iteration — and the check becomes a **typing rule**, which is local, cheap, and needs no widening or narrowing.

This is not novel; it is the convergent answer in three systems:

- **Cranelift** uses **block parameters instead of phi nodes** — *"Cranelift represents φ-nodes with block parameters rather than explicit φ-instructions"*; the stated benefits are simpler reasoning, fewer conversion bugs, and trivially simple inlining (call arguments simply become jump arguments).
- **MLIR's `scf` dialect** threads loop-carried values as **`iter_args`** on `scf.for`/`scf.while`, so loop-carried state is explicit in the op's signature.
- **WebAssembly** gives `block`/`loop`/`if` a type signature, so branches *"consume operands matching their target's type signature."*

**Recommendation: when M004 adds loops, use block parameters, never phi nodes.** And record this now, because `core.Block` (`core.go:918`) currently has no parameter list — adding one later is an additive `omitempty` field of exactly the kind the project has done repeatedly (`Blocks`/`Edges`/`LoanEndpoints` at `core.go:902-912`, `CalleeID` at `core.go:876`, `PayloadType` at `core.go:889`), so nothing about M003 needs to pre-build it. It just must not be designed *against*.

### Case 5 — Drop obligations that differ per path

Today `corevalidate.CalleeFrameNotDrained` (`corevalidate.go:784-810`) enforces per-function that a tracked acquisition is released or ownership-transferred out. Under Austral Rule 3 + Rule 5 this generalizes cleanly (every path has the same obligation set, by refusal). Under drop flags it does not. Another reason for Rule 3.

---

## Evidence/corpus strategy at affordable cost

The genuine risk is combinatorial: control flow multiplies path space, arithmetic multiplies value space, and the project already carries a **hard** gate at `recomputed_work_growth_exponent = 1200` milliexponent (1.2) and `peer_closure_recomputed_work_growth_exponent = 1300` (`internal/compiler/session/qlt02_budget_manifest.json`).

**Five rules to keep it affordable.**

1. **Keep the value dimension orthogonal to the ownership dimension.** Do **not** extend `TestOwnershipSequenceExhaustive`'s 48-symbol alphabet (`check_test.go:114`) with arithmetic symbols. Adding 8 operators takes 48³ → 56³, a ~1.55× cost on 225,890 cases, and tests nothing new about ownership — arithmetic operations are ownership-inert (they read copyable operands and produce a fresh place). Build a **separate** arithmetic enumeration. Orthogonality is the entire cost control.
2. **Enumerate over a declared edge-value alphabet, not the type's domain.** `{0, 1, 2, MAX/2, MAX-1, MAX}` × `{0, 1, 2, MAX/2, MAX-1, MAX}` × ~8 operators = **288 cases**, exhaustive over the declared alphabet, with a derived case counter that fails if the generator stops early (the `check_test.go:146-160` pattern). Add `{0}` as a divisor and `{63, 64, 65}` as shift amounts to hit each defect boundary from both sides. This is a few hundred cases, not a few hundred thousand.
3. **One new risk lane, shaped like the existing one.** Add `lane:arithmetic-differential` to `internal/compiler/session/risk_lanes.json` with `declared_inputs: ["fixture_source", "build_flags", "clang_identity", "go_toolchain"]` — byte-for-byte the dependency shape of the existing `lane:native-differential` — plus a pure-source `lane:arithmetic-negative-controls` with `["fixture_source"]`. Reusing the two existing input shapes means changed-risk selection needs no new logic.
4. **Re-ratify the growth-exponent gate deliberately, not by discovering a red build.** Arithmetic adds O(1) work per operation, so the exponent should not move; make that a *prediction recorded before the measurement*, and treat a move as a finding. (This is also the cheapest possible early warning that something in the design is secretly superlinear.)
5. **The highest-value mutation controls are two, and both are cheap.** (a) Delete the overflow pre-check from cgen → the five-axis comparator must go red on a MAX+1 fixture. (b) Change one operand cast width in cgen → the structural C-scan control from Rule 2 must go red. Both are the `loanLivenessBoundSeam` shape. A control nobody has seen fail is not evidence — the project's own words.

**One free win:** the new arithmetic fixtures are *reducible*. `reduce.Reduce` currently hard-errors on a multi-function seed (LANGUAGE-MATURITY.md, verified inventory), so P15's emitter/guard retirement should widen `reduce` in the same phase or the HDD reducer will be useless on exactly the new fixtures.

---

## Role-lens disagreements

| Lens | Position | Where it conflicts |
|---|---|---|
| **Language designer** | One unsigned integer type, `if` desugared to `match`, checked `+` with explicit `+%` for wrapping. Minimal coherent surface. | Conflicts with **Numerics**, which wants a real signed type because "a language with no negative numbers is not a language." Resolved in favour of the designer *for one milestone*, with signedness named as deferred. |
| **Compiler/IR engineer** | Block parameters, not phi. Arithmetic in A-normal form as binding RHS — no nested expression grammar. | Conflicts with the **Product/AI-agent** lens, which wants `let d = (a + b) * c` to just work. Resolved: A-normal form is *already* the language's shape (`let x = ...` with name-only operands, `ast.RHS`), and an agent emitting three `let`s instead of one nested expression is a trivial cost. Nested expressions are pure syntax and can land any time later without touching the IR. |
| **Dataflow specialist** | The existing worklist would converge on a cyclic CFG; no widening needed (finite lattice of loan-ID sets). "Loops are not that hard." | Conflicts with the **Borrow-checker** lens, which observes that the *bound* (`check.go:2656`), the *endpoint materialization* (`check.go:2812`), and the *mutual-exclusion soundness argument* (`check.go:2887-2891`) all break even though the fixpoint itself does not. **The dataflow lens is right about the algorithm and wrong about the cost.** This disagreement is the crux of the split verdict. |
| **Borrow-checker specialist** | Austral Rule 3 + Rule 5; refuse rather than elaborate drop flags. | Conflicts with the **Product** lens (refusing move-out-of-loop is a real ergonomic loss). Resolved by Case 4: block parameters give back most of the expressiveness without dataflow. |
| **C/UB specialist** | Unsigned-only, one width, explicit casts on every operand, structural scan of emitted C, never `-fwrapv` (it would blind the existing UBSan lane). | No conflict — this is the one lens nobody disagrees with. Note it *strengthens* the language-designer position rather than opposing it. |
| **Formal methods** | The oracle's authority rests on totality. Keep it total by construction (no loops), and when that ends, prefer a structural termination argument (counted loops = trivially inferred `decreases`) over a fuel budget. | Conflicts with the **Dataflow/Product** lenses, which would accept general `while` + fuel. Resolved: fuel makes the interpreter total but *breaks tier agreement*, because native has no step counter and adding one is an optimizer-visible cost at `-O3 -flto`. |
| **Numerics** | Defer float, hard. IEEE-754 agreement across a Go interpreter and Clang `-O3 -flto` needs FMA-contraction suppression and excess-precision control, for near-zero M003 payoff. | No conflict. |
| **Product / AI agent** | The single most blocking absence is **arithmetic**, not loops. An agent can express repetition by unrolling or by calling a helper; it cannot express `1 + 1` at all, and cannot even *lex* `1`. | Agrees with the verdict, and is the strongest independent support for it. |

---

## Prior art and lessons (with sources)

- **WebAssembly — structured control flow.** `block`/`loop`/`if`/`br`/`br_if`/`br_table`, with labels referenceable *only from within* the enclosing structured instruction and branches directed outwards only. Branch to a `block` = `break`; branch to a `loop` = `continue`. Branch operands must match the target's type signature. **Lesson:** reducibility becomes syntactic, and a C17 backend gets structured C for free. Adopt wholesale for M004. ([spec](https://webassembly.github.io/spec/core/syntax/instructions.html))
- **WebAssembly — integer semantics.** `iadd`/`isub`/`imul` are modulo 2^N; `idiv_*`/`irem_*` **trap** on zero (and signed division traps on `INT_MIN / -1`); shift and rotate amounts are taken **modulo N**. **Lesson:** trapping division is the mainstream choice; masking shifts is the mainstream choice and the one place this project should knowingly diverge (defect instead of mask) for auditability. ([spec](https://webassembly.github.io/spec/core/syntax/instructions.html))
- **Zig — explicit overflow.** Three families: default `+ - *` (overflow detected in Debug/Safe, wraps in Fast), wrapping `+% -% *%`, saturating `+| -| *|`; plus `@addWithOverflow`/`@subWithOverflow`/`@mulWithOverflow`. `/` on signed runtime operands is disallowed — you must pick `@divTrunc`/`@divFloor`/`@divCeil`/`@divExact`. Division by zero is illegal behavior. `>>` is logical for unsigned, arithmetic for signed. **Lesson:** the *spelling* of overflow policy belongs in the source, not in a compiler flag. Adopt two of the three families. **Anti-lesson:** Zig's default operator changing meaning by optimization mode is exactly the divergence this project's five-axis comparator exists to forbid — do **not** copy that. ([docs](https://ziglang.org/documentation/master/#Operators))
- **Rust NLL (RFC 2094).** Liveness computed for *lifetimes*, not variables; region inference is a fixpoint where *"each lifetime variable begins as an empty set and we iterate over the constraints, repeatedly growing the lifetimes until they are big enough"*; in-scope loans at each point come from a fixed-point dataflow computation; false unwind edges are added for infinite loops so the CFG has a final exit node and drop-liveness stays well-defined; `StorageDead` is a shallow write, moves are deep writes. **Lesson:** loops need a genuine loop-aware liveness *and* a drop-liveness story, and even Rust needed a synthetic-edge hack to make it well-founded. ([RFC 2094](https://rust-lang.github.io/rfcs/2094-nll.html))
- **Rust drop elaboration / drop flags.** `DropIfInit` in MIR build, lowered by drop elaboration into flag-guarded drop sites. **Lesson:** the price of permitting conditional moves is per-value runtime state — which this project would then have to prove equal across four execution tiers. ([rustc DROP_IF_INIT discussion](https://github.com/rust-lang/compiler-team/issues/558))
- **Polonius.** The Datalog borrow-checker *"doesn't scale and has no path to stabilization"*; the native rewrite reports worst-case 2-3× slowdown. **Lesson:** the cost cliff in this exact analysis shape is historically real; measure before generalizing (which is already a standing project rule). ([status](https://rust-lang.github.io/polonius/current_status.html), [2025h1 goal](https://rust-lang.github.io/rust-project-goals/2025h1/Polonius.html))
- **Austral.** *"Rule 3: linear variables defined outside an `if` statement must be used consistently in all branches — consumed in every branch or appear zero times in every branch."* *"Rule 5: linear variables defined outside a loop cannot be consumed inside the loop."* Linearity checker under a thousand lines of OCaml; borrowing (lexical lifetimes, explicit borrow construct, references cannot escape) is the escape hatch. **Lesson:** this is the *best-matched* prior art in the whole list — a linear-types systems language that solved affine-×-control-flow by refusal rather than elaboration, and stayed small. Adopt both rules. ([how the checker works](https://borretti.me/article/how-australs-linear-type-checker-works), [linear types tutorial](https://austral-lang.org/tutorial/linear-types))
- **Cranelift.** Block parameters rather than φ-instructions; stated benefits: simpler reasoning, fewer conversion bugs, trivially simple inlining (call args become jump args), and more convenient materialization in the e-graph mid-end. **Lesson:** for a hand-written checker with independent re-derivers, block parameters are strictly easier to re-derive than phi nodes. ([Cranelift isel](https://cfallin.org/blog/2020/09/18/cranelift-isel-1/), [aegraph](https://cfallin.org/blog/2026/04/09/aegraph/), [inliner](https://fitzgen.com/2025/11/19/inliner.html))
- **MLIR `scf` dialect.** `scf.for`/`scf.while` with loop-carried values as `iter_args`; *"structured"* means no gotos. `scf.while` is a two-region before/after design supporting arbitrary conditions. **Lesson:** loop-carried state as an explicit signature, not as implicit mutation — the same idea as block parameters, arrived at independently. ([scf dialect](https://mlir.llvm.org/docs/Dialects/SCFDialect/), [scf.while RFC](https://discourse.llvm.org/t/rfc-add-scf-while-to-scf-dialect/1977))
- **Swift SIL / OSSA.** *"Ownership SSA is an augmented version of SSA that enforces ownership invariants for SSA values … allows verification that SIL in OSSA form can be validated statically as not containing use after free errors or leaked memory."* **Lesson:** ownership and SSA compose; a static verifier over the ownership-annotated IR is a viable fourth independent deriver if this project ever wants one. ([SIL.md](https://github.com/swiftlang/swift/blob/main/docs/SIL/SIL.md))
- **Dafny — termination.** Every loop and recursive function has an explicit or inferred `decreases` clause: a measure that strictly decreases and is bounded below. Dafny infers it in common cases. **Lesson:** a counted loop *is* an inferred `decreases`. Choosing counted loops buys a termination proof for free and removes the fuel question entirely. ([termination tutorial](https://dafny.org/latest/OnlineTutorial/Termination), [reference](https://dafny.org/dafny/DafnyRef/DafnyRef))
- **C/UB and UBSan.** Unsigned arithmetic is defined modulo 2^N; but `unsigned char`/`unsigned short`/`uint8_t`/`uint16_t` are *"a minefield"* because they promote to signed `int` and thereby inherit signed UB. `-fsanitize=signed-integer-overflow` emits a check branch calling `__ubsan_handle_*_overflow`; since GCC 8, `-fwrapv` **disables** that check. Overflow checks can also inhibit vectorization. **Lesson:** the mitigation is not a flag, it is emitting explicit casts and having no signed types; and `-fwrapv` would actively degrade an existing evidence lane. ([Clang UBSan docs](https://releases.llvm.org/10.0.0/tools/clang/docs/UndefinedBehaviorSanitizer.html), [MaskRay, all about UBSan](https://maskray.me/blog/2023-01-29-all-about-undefined-behavior-sanitizer), [Regehr on overflow checking in LLVM](https://blog.regehr.org/archives/1384), [Hurchalla on unsigned promotion](https://jeffhurchalla.com/2019/01/16/c-c-surprises-and-undefined-behavior-due-to-unsigned-integer-promotion/), [Nayuki, C/C++ integer rules](https://www.nayuki.io/page/summary-of-c-cpp-integer-rules))

---

## Adversarial pass

**A1. "Widen the type system first — that's the debt that's actually ratified."**
*Concede the priority, reject the framing.* A comparison operator has type `Byte -> Bool`; it is literally unrepresentable while `sameType(ReturnType, Parameter.Type)` is enforced (`check.go:255`, `:3148`, `:3399`). So this is not an alternative to arithmetic — it is **arithmetic's first phase**, and it makes the already-built, zero-call-site `resolveBlame` (`check.go:769`) reachable with no new resolver code. That is why P14 is first.

**A2. "The single-function emitter cluster is more urgent, and it's at its third deferral."**
*Concede entirely, and reorder because of it.* D-11-02 → D-12-36 is now at the point where a third deferral would violate D-10-60 — the project's own no-third-deferral rule — and `cgen.Emit`, `EmitNative`, `emitLinear`, `emitBranchOperations` all hard-fail on `len(Functions) != 1`, alongside 26 more guards in `session` and 2 in `reduce`. Adding a new arithmetic case to **six** single-function emitters would mean writing the new operation six times and would make the deletion strictly harder afterwards. **This is the strongest adversarial argument in the set, and its correct consequence is P15 before P16, not "cancel M003's arithmetic."**

**A3. "Control flow will invalidate the whole evidence corpus."**
*Concede for loops; reject for arithmetic and `if`.* For loops it is true in a way that is worse than it sounds: `pathoracle`'s path enumeration would silently stop meaning what it says, with no test going red. For arithmetic and arm-terminating `if`, the CFG topology is **unchanged** — `if` lowers into the same one-successor-to-a-join arm blocks `Match.HasBlocks()` already produces — so the ownership corpus, the endpoint materialization, the path oracle, and the composition depth are all untouched. The genuinely new evidence is one orthogonal value-domain enumeration of a few hundred cases. This objection is the *reason for the split*, not a reason against the milestone.

**A4. "You're about to make the reducer, comparator, and 32 `len(Functions) != 1` guards carry a second new axis simultaneously — exactly the fingerprint that forced `OpCall` out of M001."**
*Concede, and let it set the hard cap.* The M001→M002 lesson, recorded in PROJECT.md's own decision table, is that landing a new `OperationKind` at six dispatch sites *alongside* other machinery is what cost Phases 2-4 extra remediation rounds. **Consequence: M003 adds at most two new `OperationKind`s (`OpConst`, `OpBinary`) and nothing else semantic.** No event identity, no modules, no generics, no collections. If a seventh thing appears in planning, cut it.

**A5. "Event identity (D-11-51 → D-12-21) will be crossing a third milestone boundary."**
*Partially concede.* It will, and that is uncomfortable. But it is a *function*-identity collision in shared-leaf diamond call graphs; arithmetic multiplies operations per function, not functions, and operation IDs are already globally ordinal (`analyzeArmBody`, `check.go:2874-2891`). It does not worsen under M003 and it is not on M003's critical path. It needs an owner; it does not need this milestone.

**A6. "Checked-overflow-as-defect is a runtime cost you've hidden."**
*Reject.* It is explicit, it is the point of the `+` vs `+%` split, and PROJECT.md's core value is specifically "without … hiding runtime costs." A user who wants the free version writes `+%` and gets exactly one instruction.

**A7. "Unsigned-only is a toy language."**
*Concede, and bound it.* A language with no negative numbers is not shippable. But M003 is not shipping a language; it is extending an assurance spine, and Rule 1 converts "signed overflow UB" from a thing to be defended against into a thing that **cannot be constructed** — which is the same move the project already made with call-graph cycles and with `sameType`. Signedness is named, deferred, and should be M004's or M005's *first* numeric item.

---

## Proposed M003 requirements this implies (draft REQ text)

Written user-centric and testable, in the style of PROJECT.md's validated rows.

- **TYP-01** — A function's return type may differ from its parameter type, and a contract-boundary mismatch names the correct fix location. *Testable:* `resolveBlame` (`check.go:769`) acquires at least one production call site; a B1-shaped interprocedural diagnostic is constructible from real source; DX-06 and DX-07 both flip from partial to satisfied; D-13-02b and D-13-10a close.
- **NAT-08** — Multi-function programs are emitted, reduced, and verified through one emitter path; no production code branches on `len(program.Functions) != 1`. *Testable:* the `awk` inventory in LANGUAGE-MATURITY.md returns zero non-test hits; `reduce.Reduce` accepts a multi-function seed; D-11-02 / D-12-36 close and D-10-60 is honoured rather than retired.
- **NUM-01** — A program can name an integer value in source and the value survives unchanged to every execution tier. *Testable:* a numeric-literal fixture returns the same value from `interp`, `-O0`, `-O3`, and `-O3 -flto`; `OpConst` is handled at all six dispatch sites and `TestAllOperationKindsHandledAtEverySite` stays green.
- **NUM-02** — A program can add, subtract, multiply, divide, and compare fixed-width unsigned integers, and every tier produces the identical result or the identical named failure. *Testable:* the edge-value enumeration (declared alphabet × operators, with a derived case counter) agrees across all five axes; `OpBinary` at all six dispatch sites.
- **NUM-03** — Overflow, division by zero, and out-of-range shift are named terminal outcomes, never undefined behaviour, and never differ by optimization level. *Testable:* a MAX+1 fixture produces the same `defect` outcome and the same `function.defected` reason at all four execution tiers; the ASan/UBSan lane is clean; deleting the emitted overflow pre-check turns the comparator red (mutation control).
- **NUM-04** — No emitted C performs arithmetic on a signed integer type or relies on integer promotion. *Testable:* a non-importing structural scan of emitted C refuses any uncast arithmetic operand or signed arithmetic type, and the scan is proven to fail under a one-cast-removed fault-injection seam.
- **CTL-01** — A program can branch on a computed condition, and each branch may return a *computed* value rather than a compile-time-known alternative name. *Testable:* a branching fixture whose arms return different computed values agrees across all five axes; `cgen`'s branch emitter no longer writes `arm.Pattern` as the returned literal on such a function; **D-12-43's value-divergence control becomes constructible and is constructed**.
- **CTL-02** — Every admitted program still terminates by construction; no new construct introduces unbounded execution. *Testable:* `check.cfg_back_edge`, `pathoracle.cfg_back_edge`, and the call-graph cycle refusal all remain reachable and green; `pathoracle` still enumerates a complete path set for every corpus program; no fuel or step budget exists anywhere.
- **QLT-10** — Arithmetic evidence runs in its own declared cost lane and does not move the ratified work-growth exponent. *Testable:* new lanes in `risk_lanes.json` with existing input shapes; `recomputed_work_growth_exponent` re-measured and re-ratified at or below 1200 milliexponent, with the prediction recorded before the measurement.
- **QLT-11** — Nyquist validation debt for Phases 07, 08, 11, 12, and 13 is closed. *Testable:* `/gsd-validate-phase 07 08 11 12 13` reconciles; the milestone audit's `not_validated_phases` list is empty.

---

## Confidence + what would change my mind

**Confidence: HIGH** on every codebase claim (each cites a path and a line I read in this session) and on the integer-semantics design. **MEDIUM-HIGH** on the scope verdict — the main residual uncertainty is whether P14 (return-type widening) and P15 (emitter retirement) together are large enough to crowd out P16-P19, in which case M003 becomes debt-retirement-plus-literals and arithmetic slips a phase.

**What would change my mind:**

1. **If `if` cannot in fact be made arm-terminating in the surface syntax without being unusable.** My whole "`if` is cheap" claim rests on every arm ending in a terminator. If discussion concludes that agents will constantly want fall-through `if`, then `if` costs a value merge *and* an ownership merge, and it should move to M004 alongside loops — leaving M003 as arithmetic-only. Probe this in `/gsd-discuss-phase` before planning P18.
2. **If P15 (retiring six single-function emitters + 26 `session` guards + 2 `reduce` guards) is measured at more than ~2 phases.** Then M003 is the debt-retirement milestone and arithmetic starts in M004. Spike this first; it is cheap to measure and it is the largest unknown in the plan.
3. **If a real agent-authored program demand shows loops are blocking in a way unrolling is not.** My Product-lens claim is that arithmetic blocks harder than iteration. If an actual attempted program refutes that, the split still holds (loops are still expensive) but the urgency shifts and M004 should follow immediately.
4. **If the interpreter's `value` widening turns out to be invasive.** `interp.value` is `{tag, payload string}` (`interp.go:368`) and D-12-18 deliberately made the scalar projection byte-identical to a plain string, with a mutation seam guarding it. If adding a numeric domain cannot preserve that property, the serialized-execution goldens across the whole corpus move, and P16 gets materially larger.
5. **If measurement shows arithmetic *is* superlinear in `recomputed_work`.** I predict it is O(1) per operation and exponent-neutral. A measured move above 1.2 would mean something in the derivation chain is quadratic in operation count, and would be a finding worth stopping for.
6. *Not* changed by: an argument that wrapping should be the default for `+`. I considered and rejected it on the project's own stated constraints; a new argument would need to show the checked-overflow pre-check costs something the compute-efficiency constitution actually forbids.

---

## Sources

**Codebase (read this session; path:line).**
`internal/compiler/core/core.go` — `OperationKind` set :629-683, `AllOperationKinds` :820, `TerminatorKinds` :825, `LinearOperation` :831-897, `LinearBody` :899-912, `Block` :918-923, `Edge` :927-931, `LoanEndpoint` :933-960, `HasClosedBody` :966-972, `Match`/`MatchArm` :974-1005.
`internal/compiler/check/check.go` — `detectCFGCycle` :2570-2609, `cfgBackEdgeDiagnostic` :2612-2629, `loanLivenessBoundFactor` :2631-2653, `loanLivenessBound` :2655-2658, `loanLivenessBoundSeam` :2660-2684, `loanLivenessFixpoint` :2711-2790, `materializeLoanEndpoints` :2792-2872, `analyzeArmBody` :2874-2891, `resolveBlame` :769, `sameType` enforcement :255/:3148/:3399.
`internal/compiler/pathoracle/pathoracle.go` — `MaxPaths` :56-67, `TerminatorKindsOverride` :69-77, `backEdgeError` :130-143/:243, `EnumeratePaths` call :412. `pathoracle_compose.go` — `MaxCompositionDepth` :21-85, composed enumeration :211.
`internal/compiler/corevalidate/corevalidate.go` — `peerGrayVsVisitedMutationForTest` :440-448, `checkCallGraphAcyclic` :450-500, `CalleeFrameNotDrained` (in core.go) :784-810.
`internal/compiler/interp/interp.go` — `MaxCallDepth` :11-32, `callDepthExceededDefectReason` :34-40, `maxCallDepthOverride` :42-58, `frame` :301-353, `value` :355-370, serialization seam :372-415.
`internal/compiler/cgen/cgen.go` — entry points :112/:142/:317/:615/:809/:953, `LANG_BUFFER` typedef :345, runtime byte write :399/:724/:911/:1337, `commentSafeForeignField` :1443, `validForeignCType` :1470, `lang_write_byte` :1574/:1609, branch emit call with `arm.Pattern` :1902, `emitBranchOperations` :1979-2078. `cgen_program.go` — type lowering :195-215.
`internal/compiler/syntax/token.go` — full token set :7-44. `lexer.go` — keywords :10-22, identifier/digit handling :101-124, `punctuation` :130-163. `parser.go` — bounds :12-16, `linearBody` :384-450.
`internal/compiler/ast` — `LinearBody` :150-159, `Binding` :161-165, `RHS` :167-181.
`internal/compiler/check/check_test.go` — `TestOwnershipSequenceExhaustive` :95-160, derived case counter :146-160.
`internal/compiler/evidence/evidence.go` — `evidenceInput` type switch :284-294.
`internal/compiler/session/qlt02_budget_manifest.json`, `internal/compiler/session/risk_lanes.json`.
`.planning/PROJECT.md`, `.planning/LANGUAGE-MATURITY.md`, `.planning/STANDING-VERDICTS.md`, `.planning/milestones/M002-MILESTONE-AUDIT.md`.
Corpus measurement: 115 `.lang` files / 4,189 lines (re-run the LANGUAGE-MATURITY.md commands to reverify).

**External.**
- [WebAssembly Core Specification — Instructions](https://webassembly.github.io/spec/core/syntax/instructions.html)
- [Zig Language Reference — Operators](https://ziglang.org/documentation/master/#Operators)
- [Rust RFC 2094 — Non-Lexical Lifetimes](https://rust-lang.github.io/rfcs/2094-nll.html)
- [rust-lang/compiler-team #558 — DROP to DROP_IF (drop elaboration / drop flags)](https://github.com/rust-lang/compiler-team/issues/558)
- [rustc_borrowck/src/nll.rs](https://github.com/rust-lang/rust/blob/main/compiler/rustc_borrowck/src/nll.rs)
- [Polonius — Current status and roadmap](https://rust-lang.github.io/polonius/current_status.html)
- [Rust Project Goals 2025h1 — Scalable Polonius support on nightly](https://rust-lang.github.io/rust-project-goals/2025h1/Polonius.html)
- [Austral — How Austral's Linear Type Checker Works (Rules 3 and 5)](https://borretti.me/article/how-australs-linear-type-checker-works)
- [Austral Tutorial — Linear Types](https://austral-lang.org/tutorial/linear-types)
- [Cranelift — A New Backend, Part 1: Instruction Selection (block parameters vs phi)](https://cfallin.org/blog/2020/09/18/cranelift-isel-1/)
- [Cranelift — The acyclic e-graph mid-end optimizer](https://cfallin.org/blog/2026/04/09/aegraph/)
- [A Function Inliner for Wasmtime and Cranelift](https://fitzgen.com/2025/11/19/inliner.html)
- [MLIR — 'scf' Dialect](https://mlir.llvm.org/docs/Dialects/SCFDialect/)
- [LLVM Discourse — RFC: Add scf.while to scf dialect](https://discourse.llvm.org/t/rfc-add-scf-while-to-scf-dialect/1977)
- [Swift SIL documentation (OSSA)](https://github.com/swiftlang/swift/blob/main/docs/SIL/SIL.md)
- [Dafny — Termination tutorial (decreases clauses)](https://dafny.org/latest/OnlineTutorial/Termination)
- [Dafny Reference Manual](https://dafny.org/dafny/DafnyRef/DafnyRef)
- [Clang — UndefinedBehaviorSanitizer documentation](https://releases.llvm.org/10.0.0/tools/clang/docs/UndefinedBehaviorSanitizer.html)
- [MaskRay — All about UndefinedBehaviorSanitizer](https://maskray.me/blog/2023-01-29-all-about-undefined-behavior-sanitizer)
- [John Regehr — Efficient Integer Overflow Checking in LLVM](https://blog.regehr.org/archives/1384)
- [Jeffrey Hurchalla — Surprises and Undefined Behavior From Unsigned Integer Promotion](https://jeffhurchalla.com/2019/01/16/c-c-surprises-and-undefined-behavior-due-to-unsigned-integer-promotion/)
- [Nayuki — Summary of C/C++ integer rules](https://www.nayuki.io/page/summary-of-c-cpp-integer-rules)
