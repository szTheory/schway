# Phase 27: Exact FizzBuzz and Cross-Host Evidence - Context

**Gathered:** 2026-10-03
**Status:** Ready for planning

<domain>
## Phase Boundary

Deliver a bounded FizzBuzz program in ordinary Schway source for `0 ≤ n ≤ 1,000`, with exact output and defined U64 remainder behavior. Close the shared 65,536-byte application output contract for both M005 programs (`sum_to_n` and FizzBuzz), independently compare interpreter and native results with pinned answers and reached controls, establish native evidence on macOS and Linux, and document clean-checkout commands. Keep the existing Go standard-library host, interpreter, C17/Clang route, and application runner. Do not broaden into general strings, broad IO, general division, a new backend, or resource/loan loop carries.

The FizzBuzz result covers `1…n`: use `FizzBuzz` for multiples of 15, `Fizz` for other multiples of 3, `Buzz` for other multiples of 5, and decimal digits otherwise. Each result line ends in LF; `n = 0` emits no bytes. Inputs above 1,000 fail before any application output. A write that exceeds the shared byte ceiling is a non-success outcome; earlier streamed bytes are not rolled back.
</domain>

<decisions>
## Implementation Decisions

### Bounded source output
- **D-27-01:** Add the smallest explicit source output surface needed by the two examples: `write_text` for compile-time text literals, `write_u64` for decimal U64, and `write_newline` for LF. Do not introduce general string values, interpolation, general formatting, or broad IO.
- **D-27-02:** Let explicitly-writing applications use a narrow `Unit`/no-result entry mode. Preserve the existing `U64 -> U64` app entry behavior that prints the returned value followed by LF. Do not infer print suppression from whether a function happens to write.
- **D-27-03:** Enforce one shared 65,536-byte output limit incrementally at each write chunk. A write that crosses the limit fails; prior bytes may already be visible, and no rollback guarantee is made. Callers and evidence use the process outcome to distinguish success from failure. Keep the public input bound at 1,000; exercise the output boundary with a lowered test-only limit.

### Arithmetic and failure behavior
- **D-27-04:** Add only U64 equality and remainder needed for FizzBuzz. Reject a statically known zero remainder divisor during checking. Check a dynamic divisor before native C `%` so it never reaches C's undefined zero-divisor case. Both engines report the stable reason `U64 remainder by zero`; the public app boundary returns status 65 and exact stderr `schway: U64 remainder by zero\n`. A zero-divisor witness that fails before writing has empty stdout; if a later failure follows successful writes, the streamed prefix can remain.
- Preserve the Phase 26 independent scalar-loop fixed-point analyses and explicit refusal of resource, loan, loan-derived, and unsupported cyclic state across loop back edges.

### Evidence and user guidance
- **D-27-05:** Compare each engine independently against literal expected answers, then compare engines. Include boundary cases, reached wrong-arithmetic/skipped-iteration controls, overflow and zero-remainder controls, repeated loop-event identity, output-limit boundary controls, and actual macOS/Linux native receipts tied to declared inputs. An unavailable host is recorded as unavailable, not a pass. Historical receipts remain identified by their revision and host.
- **D-27-06:** Add one Phase 27 clean-checkout guide/evidence index linked from the root README. It documents how to build and run both programs, exact results, input and output bounds, failure behavior, and host evidence provenance; keep command truth in one place rather than scattering copies.
- **D-27-07:** Keep the dependency tree shallow. Prefer a small copy-local implementation over another dependency unless a concrete need justifies the dependency's maintenance and security cost. No new dependency or backend is needed for this phase.

### the agent's Discretion
- Choose exact syntax and internal operation encoding for the bounded writes, provided the surface stays literal/decimal/LF-only and checker, validators, interpreter, native code, and app behavior agree.
- Choose the focused test-only mechanism for lowering the output limit without changing the public 65,536-byte contract.
- Organize evidence fixtures and the phase guide to match established repository patterns while keeping expected answers independent of both engines.
</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase scope and acceptance
- `.planning/ROADMAP.md` §Phase 27 — goal, success criteria, runnable witness, evidence boundary, and phase ownership.
- `.planning/REQUIREMENTS.md` — U64-02, APP-08, APP-09, EVD-12, and DX-16 acceptance requirements and traceability.
- `.planning/PROJECT.md` — project value and constraints for feedback, correctness, runtime posture, dependencies, source authority, evidence, portability, and security.
- `.planning/research/SUMMARY.md` — M005 research synthesis and source-inspected implementation boundaries.

### Living direction and preserved decisions
- `.planning/PRODUCT-ROADMAP.md` — ranked next capabilities, consumer-first scope, and current evidence/debt references.
- `.planning/LANGUAGE-MATURITY.md` — demonstrated language surface, refusal frontier, and evidence witnesses.
- `.planning/STANDING-VERDICTS.md` — durable project-wide judgments that remain in force.
- `.planning/milestones/M005-phases/26-checked-scalar-sum/26-CONTEXT.md` and `.planning/milestones/M005-phases/26-checked-scalar-sum/26-VERIFICATION.md` — Phase 26 contract, verified behavior, and loop/refusal boundaries.
- `.planning/milestones/M004-phases/25-separate-pointer-successors-and-integrated-utility/25-CONTEXT.md` — prior independent-validation and pointer-family boundaries.

### Existing source, output, and documentation patterns
- `examples/sum_to_n.schway` — current runnable source pattern and U64-returning application witness.
- `examples/phase22/README.md` and `examples/phase24/README.md` — examples of public command documentation and evidence scope.
- `internal/compiler/ast/ast.go`, `internal/compiler/syntax/lexer.go`, `internal/compiler/syntax/parser.go`, and `internal/compiler/syntax/format.go` — source representation and syntax/formatting integration points.
- `internal/compiler/core/core.go`, `internal/compiler/check/scalar.go`, and `internal/compiler/check/check.go` — core operations, scalar CFG admission, and producer checker.
- `internal/compiler/corevalidate/corevalidate.go`, `internal/compiler/originvalidate/originvalidate.go`, and `internal/compiler/pathoracle/pathoracle.go` — independent validation and path-analysis peers.
- `internal/compiler/interp/interp.go`, `internal/compiler/cgen/cgen_program.go`, and `internal/compiler/cgen/cgen.go` — interpreter and C17 emission consumers.
- `internal/compiler/session/session.go`, `internal/compiler/native/native_app.go`, and `cmd/schway/main.go` — application entry contract, native app behavior, and public commands.

No external specification is normative for this phase; the language and app contracts above are the project acceptance source. Official Rust, Go, and C references were consulted only as comparative implementation guidance for complete writes and zero-divisor behavior.
</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `examples/sum_to_n.schway` — demonstrates ordinary source with mutable scalar state, checked addition, comparison, branches, and a pre-tested loop.
- Existing U64 decimal app output and `schway_write_bytes` — reuse the bounded output path and its 65,536-byte ceiling rather than create a second IO mechanism.
- Phase 26 independent scalar analyses and evidence controls — extend their supported scalar domain without weakening authority/resource refusals.
- Separate operation consumers in checker, core validator, origin validator, path oracle, interpreter, and C17 emitter — each affected semantic boundary needs a coordinated implementation and independent peer derivation.

### Established Patterns
- Source is formatter-owned and Git-friendly; syntax/operator admission must remain synchronized with lossless parsing and formatting.
- A checked fact is not established by the producer checker alone. Independent validators derive their own admission, and each engine is compared with independent expected answers.
- Application stdout, stderr, exit status, and evidence capacity are separate contracts. Streaming C writes can leave an already-written prefix after a later failure.
- Cross-host claims are tied to actual source/build inputs and named hosts; historical receipts do not prove new Phase 27 behavior.
- Standard-library-only and shallow dependency boundaries are project constraints; Phase 27 needs no third-party package or new backend.

### Integration Points
- Add equality, remainder, and bounded-write semantics through AST/syntax, core, checking/validation, interpreter, and C17 generation.
- Extend app entry validation and both execution routes so explicit Unit-returning writes do not gain an automatic trailing `0\n`, while `sum_to_n` retains its U64 auto-print contract.
- Add the FizzBuzz source witness and connect its app/input/output cases to independent evidence and the clean-checkout guide.
</code_context>

<specifics>
## Specific Ideas

The discussion considered security, product behavior, compiler architecture, C semantics, interpreter/native parity, app UX, cross-host operations, evidence quality, documentation, and dependency maintenance. The synthesized recommendation is deliberately narrow: explicit literal/decimal/LF writes, defined checked remainder, incremental output bounds, a distinct no-result entry contract, and independent evidence.

User preference: “another copy and paste is better than another dep.” Treat this as a preference for small copy-local code and a shallow audited dependency tree; add a dependency only when a concrete need earns its maintenance and security cost.
</specifics>

<deferred>
## Deferred Ideas

- General strings, Unicode APIs, dynamic formatting, collections, and broad IO.
- General division and arithmetic beyond the exact FizzBuzz needs.
- Whole-output buffering, transactional stdout, or rollback guarantees.
- Resource, loan, or loan-derived values carried across loop back edges.
- New dependencies or backends without a named consumer and evidence-backed need.
</deferred>

---

*Phase: 27-exact-fizzbuzz-and-cross-host-evidence*
*Context gathered: 2026-10-03*
