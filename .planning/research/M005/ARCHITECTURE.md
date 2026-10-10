# Architecture Patterns

**Domain:** Schway compiler/interpreter architecture for bounded scalar computation  
**Project:** Schway, M005 Practical Computation  
**Researched:** 2026-10-02  
**Confidence:** HIGH for current integration points (source inspection); MEDIUM for the recommended loop/event protocol until a Phase 26 source witness pins it.

## Recommended Architecture

Keep one semantic route: lossless syntax and AST → `check` produces typed `core.Program` → independent core/publication/path peers validate that representation → the retained interpreter and C17 emitter execute it → the session harness compares semantic outcomes, events, and exact application output against an independent expected answer. Do not add a VM or backend for this scope. The project already has a portable typed core, an interpreter, and a single production C emitter, and M005's witness needs scalar integer operations and CFG control flow rather than a new execution model.

Add arithmetic and Bool as explicit typed core operations. Comparisons produce Bool; condition terminators/edges consume Bool. Define U64 operator behavior before lowering: use fixed-width unsigned values and explicitly specify wrapping add/subtract/multiply and remainder-by-zero behavior. C17 unsigned operations already have modulo semantics, but language semantics must be explicit and the interpreter must implement them directly rather than inheriting C behavior by accident. If M005 chooses checked overflow instead, represent overflow as a first-class semantic outcome in core and both engines; do not rely on compiler flags or host traps. Division/remainder by zero must have a defined result or a defined defect/refusal path, never C undefined behavior. The C reference for unsigned arithmetic is WG14 draft N1570 §6.2.5: [WG14 N1570](https://oifans.cn/docs/c11/n1570.html).

Represent loops using the existing `core.Block`, `core.Edge`, and block operation lists. Lower a source `while condition { body }` into a header, body, and exit with an explicit back edge; lower `if/else` into ordinary conditional CFG successors so FizzBuzz can branch on computed comparisons. Prefer this narrow structured surface over admitting arbitrary user-authored gotos or inventing a second loop IR. The AST/formatter remains the source authority; the core CFG remains the shared semantic authority. Add a forward fixed-point analysis over scalar initialization and value/type facts, and keep the existing backward loan analysis conservative: any back-edge state carrying a resource, owner, or loan is refused until a dedicated fixed-point proof establishes merge and discharge behavior.

At the execution boundary, use the existing application route for `sum_to_n` and FizzBuzz. Define a small bounded output operation (or an equally explicit typed output intrinsic) with an explicit byte ceiling and defined overflow outcome. Keep stdout/stderr as application streams; execution evidence remains a separate schema/document. Do not encode a program's printed output as the compiler's JSON execution document. Both interpreter and emitted C must expose the same stdout bytes, exit result, and semantic execution evidence for the supported witness.

### Component Boundaries

| Component | Responsibility | M005 integration |
|-----------|---------------|------------------|
| `internal/compiler/syntax`, `ast` | Parse and preserve source structure | Add only the needed operator, `while`, conditional, and bounded-output syntax. Keep formatter-owned stable spelling and source spans. |
| `internal/compiler/check` | Type-check source and construct canonical typed core | Derive operand/result types, branch predicates, U64 operation facts, CFG, and scalar fixed-point state. Emit stable refusal diagnostics before malformed core can proceed. |
| `internal/compiler/core` | Portable semantic representation | Add explicit arithmetic/comparison/output operations or terminators; use blocks and edges for loop and conditional flow. Avoid embedding source-only syntax choices. |
| `internal/compiler/corevalidate` | Independent structural and semantic validation of serialized core | Independently derive operator signatures, CFG reachability/dominance/merge rules, scalar initialization, and loop fixed-point facts. Do not import check's conclusions. |
| `internal/compiler/originvalidate` | Independent publication/origin safety validation | Verify source-produced operands and targets, Bool branch origins, output argument provenance, and allowed operation forms without trusting check's origin map. |
| `internal/compiler/pathoracle` | Independent path/obligation oracle | Replace unconditional cycle refusal only for admitted scalar-only loops. Bound oracle work separately from program semantics; refuse unsupported ownership/loan cycles. |
| `internal/compiler/interp` | Reference execution engine | Execute CFG with explicit scalar state and operator semantics; produce output and dynamic events using identities independent of native code. |
| `internal/compiler/cgen` | Sole C17 lowering authority | Emit the same typed operations and CFG, including explicit loop branches; ensure every emitted C expression is defined for every checked input. Preserve conservative attributes and no new backend. |
| `internal/compiler/executionpeer`, `execution` | Validate evidence schema and event membership/identity | Extend identity rules for repeated dynamic execution; distinguish evidence/document exhaustion from application behavior. |
| `internal/compiler/session`, `native` | Integration, differential evidence, process capture | Run interpreter/native against independent expected answers; capture stdout, stderr, exit status, and bounded evidence separately. Retain the app build/run boundary and one process launch. |

### Data Flow

```text
.schway source
   ↓ parse / format-preserving AST
typed operators + structured while / if + bounded output
   ↓ check: canonical typed core and CFG
independent corevalidate + originvalidate + pathoracle
   ↓
┌─────────────────────┬────────────────────────┐
│ interpreter         │ core → C17 → Clang app │
└─────────────────────┴────────────────────────┘
   ↓                             ↓
semantic result/events     stdout/stderr/exit
   └──────────────┬──────────────┘
                  ↓ session compares each to independent oracle
```

For the first witness, `sum_to_n` should exercise one bounded U64 input and the loop-carried accumulator/index; use a predeclared bound that guarantees termination and pin behavior at zero, the maximum accepted input, and arithmetic boundaries. FizzBuzz then composes the same arithmetic, comparisons, branches, loop, and output without adding a new runtime abstraction. Expected output should be a separately authored literal or independently computed reference, never derived from interpreter output and reused as the native expectation.

## Required Independent Derivations Across the Six Consumers

The current Phase 18/19 pattern carries evidence farther than a frontend-only feature: `OpConst` is explicitly handled in `check`, `corevalidate`, `originvalidate`, `pathoracle`, `interp`, and `cgen`. M005 must preserve that discipline for each new operation.

| Consumer | Independent fact to derive for M005 | Failure if omitted |
|----------|-------------------------------------|--------------------|
| `check` | Source typing, U64 operator definition, Bool predicate typing, block/edge construction, scalar fixed-point facts, and loop-state refusal diagnostics | Source may be accepted with ambiguous overflow, uninitialized loop-carried values, or invalid branch types. |
| `corevalidate` | Operator arity/types, result fact consistency, valid conditional terminators, CFG fixed point / reachability, and no owner/loan carry across back edges | Mutated or imported core could bypass frontend checks; a checked program's claims become self-authenticating. |
| `originvalidate` | Bool and scalar results come from admitted source operations; condition and output operands are valid, initialized, and not fabricated | Publication safety can accept impossible provenance despite structural consistency. |
| `pathoracle` | Cycle/loop classification, scalar-only eligibility, finite traversal/evidence bound, and refusal when loop obligations include resource/loan state | Either the oracle rejects all loops forever or an unbounded/unsound path walk silently omits iterations and obligations. |
| `interp` | Runtime operation semantics and condition/loop behavior, including termination/output bounds and dynamic event occurrence counting | The source reference engine can disagree with the specified language or emit duplicate event identities. |
| `cgen` | Same operation semantics translated to defined C17, correct loop CFG, explicit Bool branch lowering, and output cap behavior | Interpreter-only success gives a false production claim; C UB or compiler optimization may diverge from the language. |

The operation-kind registry/dispatch controls must include the new core kinds at every consumer. Reuse validation gates so each kind is demonstrably reached in positive and negative cases. A checker peer omitted from the work leaves a forged-core gap; the interpreter omitted leaves no independent executable semantics; the C17 path omitted fails the actual production route. Any one of these omissions makes cross-engine agreement partial rather than evidence for the milestone.

## Dynamic Event Identity and Loop-Oracle Limits

`lang.execution/2` currently identifies schema-2 events by `(invocation, event ID)`. `execution.InvocationSegment` includes an ordinal, but the current acyclic call graph gives every segment ordinal zero; `executionpeer` derives a bounded, statically unfolded call-membership table. That makes two dynamic executions of the same operation in one function activation indistinguishable under a loop. Do not silently reuse static IDs for loop events.

Before admitting repeated operations, extend the evidence identity contract (likely a schema `/3` change) with a per-activation dynamic occurrence ordinal for each event-producing operation, or an equivalent canonical dynamic path counter. Define whether the counter increments on every executed operation or only event-producing operations, how it resets on a new function activation, and how overflow/exhaustion is reported. The peer must independently check canonical, monotonic occurrence identities and that an event is a member of an admitted function/operation; the interpreter and native emitter must each derive their own counters. The comparator must compare identity and event order, not just count repeated IDs. This is distinct from call invocation identity: a loop iteration is not a new function activation.

The path oracle's search bound is an evidence/tooling ceiling, not a language step limit and not permission to truncate application semantics. If the oracle cannot finish validating an evidence document or coverage claim, return an explicit incomplete/exhausted result; never turn that state into a valid execution or a source-level nontermination result. Bound runtime capture/output separately. For a first loop witness, choose an explicitly bounded input domain, include a zero-iteration case, and pin the iteration/output ceiling. Do not use a test-only timeout as the language's termination guarantee.

## Resource and Loan Carry Refusal Boundary

Scalar-only cycles may be admitted after checker and peers converge on a finite dataflow fixed point. A back edge whose incoming state includes a live owner, a transfer obligation, a shared/exclusive loan, or a loan-derived value remains refused. Do not merge or union those facts optimistically: ownership is path-sensitive and loan expiry depends on all future loop iterations. A loop that acquires and releases a resource within each iteration is also out of this slice unless the checker, core peer, origin peer, path oracle, interpreter, and C emitter all prove exactly-once discharge for zero, one, and repeated iterations, including early exits. Prefer a whole-loop refusal code with a stable source span and named back-edge cause over a partial execution.

This keeps M005 independent from completed M004 cleanup claims and avoids smuggling loop-specific destructor, break/continue, or error-precedence semantics into the scalar witness. A future consumer that requires a resource to cross an iteration should trigger a separate ownership-loop design and decisive native evidence, not simply relax the current refusal.

## Structured Loop and Branch Options

| Option | Fit for the witness | Recommendation |
|--------|---------------------|----------------|
| `while condition { body }` lowered to the existing CFG | Directly expresses bounded `sum_to_n` and FizzBuzz iteration; ordinary break-on-condition maps to one back edge and one exit | Recommend. Keep condition evaluation at the header and use explicit edges. |
| `loop { ... }` with `break` / `continue` | More general; creates extra exit and continue targets and makes early-exit resource discharge a relevant concern | Defer until a concrete witness needs it. A single `while` covers M005 and narrows CFG/evidence states. |
| `if/else` as structured statement lowered to CFG | Needed to choose FizzBuzz labels and to branch on computed Bool values | Admit the minimal form required by FizzBuzz, with no general expression-level control-flow abstraction unless the source witness needs it. |
| Arbitrary jumps / user-authored CFG | No user benefit for the selected programs; would expose unstructured topology and expand diagnostics/analysis obligations | Do not admit. |

## Patterns to Follow

### Pattern 1: Typed core operation with independent peers
**What:** Give each semantic operator its own core kind and closed operand/result contract, then implement a separate derivation in all six consumers. Avoid hiding arithmetic inside a generic stringly-typed foreign call or source-only evaluator.  
**When:** Every operation can affect type safety, execution, or evidence.  
**Example:** A typed `U64 remainder` has two U64 inputs and a U64 result; a comparison has two U64 inputs and a Bool result; branch edges require Bool.

### Pattern 2: Structured source, canonical CFG
**What:** Keep `while` and `if/else` as readable source constructs while representing them as explicit core blocks and edges.  
**When:** User-facing control flow must be simple while validators and both engines need one control-flow model.  
**Example:** A `while i <= n` header evaluates the condition, chooses body/exit, and the body back edge returns to the same header.

### Pattern 3: Separate program output from compiler evidence
**What:** Capture bounded stdout/stderr as application data and execution events as a separate evidence document.  
**When:** A program emits user-visible text while compiler tooling also reports execution records.  
**Example:** FizzBuzz's exact newline-delimited stdout is compared as bytes; evidence still records terminal outcome and semantic events independently.

## Anti-Patterns to Avoid

### Reusing acyclic path enumeration for loops
**What:** Remove `cfg_back_edge` and let existing path enumeration traverse cycles.  
**Why bad:** It can fail to terminate or exhaust by enumerating paths forever; a capped walk can omit real loop executions while appearing complete.  
**Instead:** Use finite monotone fixed-point analysis for static facts, a separately bounded dynamic evidence model, and explicit incomplete status on oracle exhaustion.

### Treating interpreter/native agreement as the expected answer
**What:** Compute expected output from one engine and compare the other engine to it.  
**Why bad:** Shared wrong assumptions can agree, especially when both implement wrapping, loop bounds, or output truncation incorrectly.  
**Instead:** Independently specify edge cases and exact output; keep a wrong-result or changed-assumption control that reaches the disputed behavior.

### Admitting a broad generic loop construct first
**What:** Add `loop`, `break`, `continue`, nested labels, and ownership-bearing exits to unlock one summation and FizzBuzz.  
**Why bad:** Adds topology and lifetime cases that the witnesses do not need.  
**Instead:** Start with one structured `while` and conditional branches; promote other exits only with a concrete runnable consumer.

## Scalability Considerations

| Concern | At 100 users/programs | At 10K | At 1M |
|---------|----------------------|-------|-------|
| CFG fixed-point work | Use worklists and bounds derived from finite lattice height and CFG size; measure cold/warm distributions | Keep transfer functions monotone and sparse; avoid rescanning all blocks per fact | Need measured profiling and possibly bitsets/SCC scheduling; preserve same derivation and refusal contract |
| Dynamic event volume | Bound per-run document bytes and report exhaustion explicitly | Add streaming or bounded summaries only if the execution evidence consumer needs it | Separate full traces from summaries; never claim full equivalence from a truncated trace |
| Application output | Explicit small byte ceiling with a stable failure outcome | Keep cap configurable only if its value is part of reproducible run identity | Stream to caller with backpressure only after a real application contract requires it |

These are design thresholds, not claims about measured performance. Current project constraints require cold and warm feedback distributions before making latency claims.

## Evidence and Inspection Notes

- **Source inspection (this research run):** `core.OperationKind` has `OpConst` and payload operations; `core.Block`/`Edge` are already explicit; `check.loanLivenessFixpoint` calls cycle detection and returns `check.cfg_back_edge`; `pathoracle` has `pathoracle.cfg_back_edge`; `cgen.emitProgram` is the current whole-program C17 route; `interp` executes the core; schema `/2` event identity includes call invocation paths. No tests were run for this research task.
- **Historical receipts (not newly re-run):** M003 established the six-operation-consumer discipline, `/2` invocation identity for acyclic calls, and interpreter/native comparison across its shipped witnesses. These receipts do not prove loop behavior.
- **Newly run checks:** None. Recommendations above are architecture analysis, not implementation or test evidence.

## Sources

- Current project contracts: `.planning/PROJECT.md`, `.planning/PRODUCT-ROADMAP.md`, `.planning/LANGUAGE-MATURITY.md`, `.planning/REQUIREMENTS.md` (read for this research task).
- Current source witnesses: `internal/compiler/core/core.go`; `internal/compiler/check/check.go`; `internal/compiler/corevalidate/corevalidate.go`; `internal/compiler/originvalidate/originvalidate.go`; `internal/compiler/pathoracle/pathoracle.go`; `internal/compiler/interp/interp.go`; `internal/compiler/cgen/cgen_program.go`; `internal/compiler/execution/{execution.go,invocation.go}`; `internal/compiler/executionpeer/executionpeer.go`; `internal/compiler/session/session_phase5_compare.go`.
- WG14, *N1570 Committee Draft — C11*, §6.2.5, integer types and unsigned modulo arithmetic: https://oifans.cn/docs/c11/n1570.html. Used only for the C17 lowering fact; this source is the C11 committee draft, not a C17 standard text.
