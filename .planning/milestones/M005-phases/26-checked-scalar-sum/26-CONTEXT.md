# Phase 26: Checked Scalar Sum - Context

**Gathered:** 2026-10-02  
**Status:** Ready for planning

<domain>
## Phase Boundary

Deliver a runnable `sum_to_n` program from ordinary Schway source through the
public application route. For input `0`, `10`, and `1,000`, stdout is exactly
`0\n`, `55\n`, and `500500\n`. Accept `0 ≤ n ≤ 1,000`; a larger input fails
without application output. U64 addition is checked and has matching
interpreter and native behavior. Admit Bool conditions and predicate-controlled
scalar loops with bounded CFG fixed-point analysis. Continue to refuse
ownership, resource, loan, and loan-derived provenance across a loop back edge,
and fail closed when analysis exceeds its deterministic bound. Phase 27 owns
remainder, equality for FizzBuzz, bounded text output, and the full cross-host
evidence contract.

</domain>

<decisions>
## Implementation Decisions

### Loop source form and scalar state

- **D-26-01:** Use function-local mutable scalar bindings (`var`) for the
  `U64` loop counter and accumulator. Keep this mutability limited to copyable
  `U64`/`Bool` state; it does not admit mutation of owners, resources, loans,
  pointers, or aggregate values. The user explicitly selected this option.
- **D-26-02:** Use a conventional pre-tested `while condition { ... }` form.
  Re-evaluate the Bool condition before each iteration; `n = 0` performs zero
  iterations. Use `if/else` block control flow with Bool conditions, without
  implicit truthiness. The user approved the specialist recommendation.
- **D-26-03:** Use infix `+` for U64 addition and `<` for the Phase 26 sum
  predicate. Keep equality and remainder with Phase 27, where FizzBuzz needs
  them. Do not add other arithmetic or the full relational/equality operator
  set without a phase requirement.

An illustrative loop shape is:

```schway
var i = 0
var total = 0
while i < n {
  i = i + 1
  total = total + i
}
```

The exact AST, type facts, lowering, and formatter mechanics remain planning
work. Keep value-yielding `if`, pattern conditions, short-circuit condition
chains, iterator loops, `break`, and `continue` out of this slice unless a
Phase 26 acceptance criterion cannot be met otherwise.

### Checked U64 overflow and caller-visible failure

- **D-26-04:** Overflow is a defined checked failure, never a wrapped success.
  When it reaches the application entry, report a program-level non-success
  with a nonzero outcome, bounded deterministic stderr, and no stdout. Keep
  this distinct from compiler/tool failures and from evidence-capacity
  exhaustion. The user approved the specialist recommendation.
- **D-26-05:** Do not make arithmetic overflow a catchable source-level typed
  error in Phase 26. A future named consumer that must recover and continue is
  the observation that would justify revisiting that contract.
- **D-26-06:** The bounded `sum_to_n` inputs cannot overflow, so independently
  exercise overflow with a direct arithmetic witness near `U64::MAX`; pin the
  expected program outcome rather than inferring it from interpreter/native
  agreement alone.

### Scalar loop guarantees and evidence

- **D-26-07:** Accept cyclic control flow only after finite, monotone,
  deterministic fixed-point analysis establishes the admitted scalar state.
  Do not remove `check.cfg_back_edge` as a shortcut or treat the existing
  acyclic path enumerator as a loop proof. Independent validators must derive
  the admitted facts themselves.
- **D-26-08:** Keep any loop-carried owner, resource, loan, or loan-derived
  provenance refused with stable source-attributed diagnostics. Analysis
  bounds fail closed. Where loop execution can repeat an event, preserve
  distinguishable dynamic occurrences; evidence exhaustion is never a
  language-level result.
- **D-26-09:** Compare interpreter and native execution separately with
  independently pinned answers. Include the three accepted sum values, the
  rejected `1,001` boundary, direct overflow, reached wrong-result and
  skipped-iteration controls, and resource/loan back-edge refusal controls.
  Check in and exercise the `sum_to_n` source witness before broadening loop
  support. Source inspection and historical M004 receipts are not Phase 26
  execution evidence.

### Dependencies and implementation scope

- **D-26-10:** Retain Go 1.24 standard library, the existing interpreter,
  readable C17/Clang emission, and the current application runner. Add no
  dependency or backend for this slice; copy-local code is preferred unless a
  concrete need outweighs its dependency-tree and audit cost.

### The agent's Discretion

- Choose the precise grammar productions and canonical formatting for `var`,
  `while`, `if/else`, `<`, and checked infix `+`, with source spans and stable
  recovery.
- Choose the scalar lattice, CFG worklist and fail-closed analysis budget,
  independently derived validation facts, and the source-attributed refusal
  diagnostics, while preserving D-26-07 and D-26-08.
- Choose how loop-event occurrence identity is represented and what focused
  evidence proves it; do not conflate repeated dynamic events or application
  outcomes with exhausted tooling capacity.
- Use `if/else` as block control flow in the initial slice. Keep branch values,
  condition chains, `break`/`continue`, iterator forms, and nested-loop
  guarantees outside unless required by the approved witnesses.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase boundary, requirements, and product direction

- `.planning/ROADMAP.md` § Phase 26 and § Phase 27 — runnable goals, exact
  accepted inputs/outputs, refusal frontier, and the equality/remainder handoff.
- `.planning/REQUIREMENTS.md` § U64 semantics, § Scalar control flow, § Runnable
  applications, and § Traceability — U64-01, U64-02, FLOW-01/02, and APP-07
  ownership.
- `.planning/PROJECT.md` § Core Value, § Constraints, and § Verification
  Operating Preference — defined behavior, portability, explicit costs, and
  evidence policy.
- `.planning/PRODUCT-ROADMAP.md` § Checker and guarantee activation map, §
  Current three recommendations, and § Phase 26 discussion — current
  capability order and dated design decisions.
- `.planning/LANGUAGE-MATURITY.md` § M005 kickoff, § Phase 26 discussion
  amendment, § What exists, § What prevents ordinary programs, and § Next
  useful thresholds — source witnesses, refusals, evidence provenance, and
  next capabilities.
- `.planning/STANDING-VERDICTS.md` — dependency posture, independent derivation,
  evidence controls, and known process pitfalls.
- `.planning/research/SUMMARY.md` — M005 source-inspection findings, proposed
  architecture, cross-ecosystem arithmetic alternatives, and failure modes.
- `.planning/milestones/M004-phases/25-separate-pointer-successors-and-integrated-utility/25-CONTEXT.md`
  — accepted copy-local dependency posture and completed source/compiler
  boundary immediately preceding this phase.

### Current language and evidence seams

- `examples/checksum.schway` and
  `internal/compiler/session/session_phase20_test.go` — provisional loop-shaped
  source and the refused checksum frontier; these are not admitted loop
  semantics.
- `internal/compiler/ast/ast.go`, `internal/compiler/syntax/lexer.go`,
  `internal/compiler/syntax/parser.go`, and
  `internal/compiler/syntax/format.go` — current lossless source representation,
  parser, and formatter to extend.
- `internal/compiler/core/core.go` — operation kinds and CFG facts; each new
  operation/control fact must stay registered and validated.
- `internal/compiler/check/check.go` and
  `internal/compiler/check/check_test.go` — current cycle refusal and
  `TestBackEdgeRejected` negative control.
- `internal/compiler/corevalidate/corevalidate.go`,
  `internal/compiler/originvalidate/originvalidate.go`, and
  `internal/compiler/pathoracle/pathoracle.go` — independent facts and bounded
  path rules that must not accept an unproved resource/loan loop carry.
- `internal/compiler/interp/interp.go` and
  `internal/compiler/cgen/cgen_program.go` — independently implemented
  interpreter and sole production C serializer.
- `internal/compiler/native/native_app.go` and `cmd/schway/main.go` — bounded
  application input, stream/status behavior, and public command integration.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- `schway app run` already accepts bounded U64 input and preserves ordinary
  stdout, stderr, and process outcomes; reuse it for the runnable sum rather
  than the synthetic-input conformance route.
- The checked core already represents control flow as blocks and edges, and
  the interpreter/native app path can compare actual program outcomes with
  separate expected answers.
- Go 1.24, the standard library, C17/Clang, and existing app output limits
  provide the full implementation stack; no external dependency is indicated.
- `examples/checksum.schway` offers a provisional source witness for the
  historical refusal boundary only; its syntax and behavior are not a
  supported language contract.

### Established Patterns

- `check`, `corevalidate`, `originvalidate`, `pathoracle`, `interp`, and `cgen`
  are separate semantic consumers. Each affected peer must independently
  derive or validate new scalar/control-flow facts.
- `check.cfg_back_edge` currently rejects cyclic CFGs; `TestBackEdgeRejected`
  and `TestPhase20ChecksumFrontier` pin relevant refused behavior. Simply
  admitting cycles would bypass the current proof boundary.
- The application runner keeps program streams/status separate from evidence
  capture and replay. Preserve that separation; evidence exhaustion must not
  become overflow behavior or a successful program result.
- U64 source literals and `OpConst` already exist, but the operation inventory
  has no addition or comparison operation. Native lowering must implement the
  ratified checked semantics explicitly instead of inheriting a host-language
  overflow policy.

### Integration Points

- Extend the syntax/AST/formatter, checker, and core together for scalar
  addition, comparison, Bool conditions, and structured control flow.
- Update the independent core/origin/path validators with bounded scalar
  fixed-point derivation while retaining resource, loan, and provenance
  refusals across back edges.
- Implement the same checked addition and control-flow behavior independently
  in the interpreter and C emitter; preserve exact app output and failure
  boundaries.
- Ensure repeated loop execution cannot alias dynamic evidence events, and
  compare each engine with independently authored expected answers plus reached
  controls.

</code_context>

<specifics>
## Specific Ideas

- Recommended source shape uses local mutable `U64` state and a pre-tested
  `while i < n` loop; use familiar infix `+` and no implicit Bool truthiness.
- `sum_to_n(0)`, `sum_to_n(10)`, and `sum_to_n(1000)` must print `0\n`, `55\n`,
  and `500500\n`; input above `1,000` fails before application output.
- Use a direct near-maximum U64 witness for checked overflow. At the app
  boundary it is a nonzero program failure with bounded stderr and empty
  stdout, not a catchable source error.
- External precedent supports explicit language policy rather than accidental
  host behavior: Go defines unsigned wraparound
  ([specification](https://go.dev/ref/spec#Integer_operators)); Rust exposes
  `u64.checked_add` as an `Option` result
  ([API](https://doc.rust-lang.org/std/primitive.u64.html#method.checked_add));
  and distinguishes predicate `while` from infinite `loop` and iterator `for`
  ([reference](https://doc.rust-lang.org/reference/expressions/loop-expr.html)).
  These are comparisons, not dependencies or requirements for Schway.
- Source and planning documents were inspected, but no Phase 26 tests, builds,
  or native executions were run. Historical M004/Phase 25 receipts establish
  only their recorded behavior and revisions.

</specifics>

<deferred>
## Deferred Ideas

- Phase 27 owns U64 equality and remainder, exact FizzBuzz text, shared
  application output bounds, and full macOS/Linux evidence.
- Catchable overflow, wrapping/saturating arithmetic, subtraction,
  multiplication, division, iterator/`for` loops, `break`/`continue`, arbitrary
  jumps, nested-loop guarantees, and owner/resource/loan/provenance state
  across loop back edges remain outside this phase.
- A bounded checksum module or JSON configuration reader remains after a named
  consumer demonstrates the need; no such capability is pulled into Phase 26.

</deferred>

---

*Phase: 26-checked-scalar-sum*  
*Context gathered: 2026-10-02*
