# Phase 19: Numeric Literals and `OpConst` - Context

**Gathered:** 2026-09-24
**Status:** Ready for planning

<domain>
## Phase Boundary

Let Lang name a value it was not given by adding numeric literals and the
`OpConst` core operation. Deliver VAL-01 through VAL-03: parse and format a
literal, check and execute it in both engines, register the operation at all
six dispatch sites, and compare interpreter, `-O0`, `-O3`, and `-O3 -flto`
results. Keep the scalar execution projection byte-identical for existing
values, or justify each changed golden.

This phase adds one fixed-width unsigned type. Signed integers, additional
widths, floats, a numeric tower, arithmetic, comparison operators, `Bool`,
loops, and other new control-flow or value capabilities remain outside scope.
</domain>

<decisions>
## Implementation Decisions

The user asked for recommendations after a broad, adversarial review across
language design, compiler, backend, portability, security, product, and
operations perspectives, then explicitly approved following the
recommendations automatically.

### Unsigned type and bounds

- **D-19-01:** Use one architecture-independent `U64` type with values from 0
  through 18,446,744,073,709,551,615. Reject a literal outside this range with
  a source diagnostic before it reaches executable core; never truncate or
  wrap it. Native targets must provide an exact 64-bit unsigned representation
  or fail closed. Do not make the source type width target-dependent.

### Literal spelling and formatting

- **D-19-02:** Accept decimal, `0x` hexadecimal, and `0b` binary integer
  literals. Permit `_` separators only between digits. Do not accept octal
  forms or type suffixes. Preserve the literal's original token spelling in
  formatted source while canonicalizing whitespace and layout.

### Literal typing and source use

- **D-19-03:** Support direct initialization such as `let count = 42`. Every
  numeric literal has the single `U64` type. Do not add literal suffixes,
  contextual numeric inference, or a general untyped/arbitrary-precision
  constant system. Existing signature and type-checking rules remain the
  authority at function boundaries.

### Runtime representation and evidence

- **D-19-04:** Serialize a numeric scalar's execution value as a canonical
  decimal string, independent of its source radix or separators. Preserve
  D-12-18's byte-identical scalar projection for existing values. If any
  serialized execution golden must move, justify that movement explicitly.
- **D-19-05:** Keep `OpConst` honest at all six dispatch sites and retain the
  four-tier literal-bearing differential fixture required by the roadmap.
  Treat successful registration alone as insufficient evidence; the literal
  must execute end to end.

### the agent's Discretion

Parser/core representation, helper boundaries, fixture names, diagnostic text
and exact test factoring are open to the planner, provided the decisions above,
the existing independent admission peers, the six-site exhaustive controls,
the five-axis semantic comparison, and the zero-external-production-dependency
constraint remain intact.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase and project authority

- `.planning/ROADMAP.md` §Phase 19 — goal, scope exclusions, requirements,
  success criteria, golden-preservation gate, and sizing.
- `.planning/REQUIREMENTS.md` §Values — VAL-01, VAL-02, and VAL-03.
- `.planning/PROJECT.md` — portability, correctness, native path, and source
  authority constraints.
- `.planning/STATE.md` — current milestone and session position.
- `.planning/STANDING-VERDICTS.md` — zero-production-dependency constraint,
  six dispatch sites, and evidence anti-patterns.
- `.planning/LANGUAGE-MATURITY.md` — actual current surface and the absence of
  arithmetic and numeric values before this phase.
- `.planning/research/M003/ADVERSARIAL-SYNTHESIS.md` §P19 — authoritative
  M003 phase framing and evidence gate.

### Prior phase decisions

- `.planning/phases/18-branch-on-a-computed-value/18-CONTEXT.md` — immediately
  preceding source-to-core and branch boundary; arithmetic remains deferred.
- `.planning/milestones/M002-phases/12-result-payloads/12-CONTEXT.md` —
  D-12-17/D-12-18 value-domain widening and evidence-invisible scalar
  serialization contract.

### Implementation seams

- `internal/compiler/syntax/token.go`, `internal/compiler/syntax/lexer.go`,
  `internal/compiler/syntax/parser.go`, and
  `internal/compiler/syntax/format.go` — tokenization, lossless parsing, and
  formatter-owned source projection.
- `internal/compiler/ast/ast.go` — source AST shapes.
- `internal/compiler/core/core.go` — `OperationKind`, `AllOperationKinds`, and
  typed core operation representation.
- `internal/compiler/interp/interp.go` — current `{tag, payload string}`
  scalar value model and D-12-18 projection.
- `internal/compiler/cgen/cgen_program.go` — surviving production C emitter.

### Language precedent consulted

- [Go specification: constants](https://go.dev/ref/spec#Constants) — exact,
  untyped constants require arbitrary-precision representation and contextual
  conversion rules; useful precedent, deliberately not adopted here.
- [Rust Reference: integer literal expressions](https://doc.rust-lang.org/reference/expressions/literal-expr.html)
  — suffix and inference rules are coupled to a multi-type numeric system.
- [C++ draft: integer literals](https://eel.is/c%2B%2Bdraft/lex.icon) — radix,
  separators, and suffix examples; useful lexical comparison, not a source
  compatibility target.
- [Zig documentation: integer literals](https://ziglang.org/documentation/master/#Integer-Literals)
  — examples of decimal, binary, hexadecimal, and between-digit separators.
</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- The lossless syntax `Tree` retains source bytes and trivia; formatter logic
  already projects a canonical layout while keeping semantic token text.
- `core.AllOperationKinds()` and `TestAllOperationKindsHandledAtEverySite`
  provide the forcing function for making `OpConst` real at all six peers.
- The interpreter's value struct already has a scalar payload string, which
  can preserve the established scalar serialization while the internal value
  domain widens.
- `emitProgram` is the surviving production C lowering path and should own
  native constant emission.

### Established Patterns

- Syntax is lossless; semantic AST/core projections are separate from source
  token storage.
- Every new operation is independently handled by `check`, `corevalidate`,
  `interp`, `cgen`, `pathoracle`, and `originvalidate`.
- Interpreter/native agreement is established through the existing five-axis
  comparator and distinct optimization tiers, not by emitter registration.
- Deterministic fixtures and mutation controls test distinct claims; a green
  wired path alone is not evidence that a literal executes.

### Integration Points

- Lexer tokens feed parser and formatter, then AST checking produces a typed
  core `OpConst`; independent validators admit that core before interpreter
  and C emitter execution.
- Numeric runtime serialization feeds existing execution goldens and the
  four-tier comparator.
- Exact-width target representation crosses from Lang's architecture-neutral
  `U64` semantics into C17 code generation and must fail closed if unavailable.
</code_context>

<specifics>
## Specific Ideas

Use a direct source form such as `let count = 42`; accept `0x` and `0b` for
systems-oriented notation without adding octal ambiguity; preserve source
spelling but serialize runtime values as decimal. Existing values must retain
D-12-18's byte-identical scalar projection.
</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within Phase 19's boundary. Signed types, more widths,
floating point, arithmetic, and comparisons remain deferred per the roadmap.
</deferred>

---

*Phase: 19-numeric-literals-and-opconst*
*Context gathered: 2026-09-24*
