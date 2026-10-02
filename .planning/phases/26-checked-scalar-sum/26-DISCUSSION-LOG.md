# Phase 26: Checked Scalar Sum - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in `26-CONTEXT.md` — this log preserves alternatives.

**Date:** 2026-10-02  
**Phase:** 26-checked-scalar-sum  
**Areas discussed:** Control syntax and loop state, U64 comparison surface, overflow failure surface

---

The user asked for specialist fan-out across relevant product, language API,
compiler, safety, portability, diagnostics/formatting, and evidence lenses,
including adversarial alternatives and external ecosystem precedent where
useful. Three `gsd-advisor-researcher` agents reviewed the candidate decisions.
The synthesis used primary references for Rust predicate/infinite/iterator
loop distinctions, Rust `u64.checked_add`, and Go unsigned overflow policy.
The user prefers standard-library and copy-local code over adding a dependency.

No Phase 26 project tests, builds, or native execution were run. Current source
and planning evidence were inspected; M004 and Phase 25 receipts remain
historical evidence for their recorded scope and revisions only.

## Control syntax and loop state

| Option | Description | Selected |
|--------|-------------|----------|
| Scoped mutable scalar bindings | Function-local `var` for copyable `U64`/`Bool` loop state; keep resource and loan values refused across back edges | ✓ |
| Explicit immutable loop state | A new form that carries values from one iteration to the next without scalar assignment | |
| Other | User-specified update form | |

**User's choice:** `1` (scoped mutable scalar bindings).

The follow-up choice was the predicate-loop keyword:

| Option | Description | Selected |
|--------|-------------|----------|
| `while condition { ... }` | Pre-tested predicate loop with a clearly named condition | ✓ |
| `loop condition { ... }` | Reuse the provisional checksum example's `loop` keyword | |
| Other | User-specified spelling | |

**User's choice:** The user subsequently approved the fan-out recommendations
for all areas, including conventional pre-tested `while`. The initial `loop
buffer` checksum example is provisional and refused; it does not establish
iterator or loop syntax.

**Additional recommendation adopted:** Use ordinary Bool `if/else` blocks for
control flow, not implicit truthiness or value-yielding conditionals in this
slice. Keep iterator forms, `break`/`continue`, and nested-loop guarantees out
of scope.

## U64 comparison surface

| Option | Description | Selected |
|--------|-------------|----------|
| Minimal Phase 26 subset | Admit `<` for `sum_to_n`; add equality together with remainder in Phase 27 | ✓ |
| Full operator set | Admit `<`, `<=`, `>`, `>=`, `==`, and `!=` before Phase 27 | |
| Other | User-specified comparison set | |

**User's choice:** Adopt the specialist recommendation: only `<` in Phase 26.
It is enough for the sum loop. Equality and remainder remain Phase 27 work.

## Overflow failure surface

| Option | Description | Selected |
|--------|-------------|----------|
| Application failure | Nonzero program outcome, bounded stderr, and no stdout; not catchable in source | ✓ |
| Typed source error/result | Source can catch or propagate arithmetic overflow as a typed value | |
| Other | User-specified failure contract | |

**User's choice:** Adopt the specialist recommendation: checked overflow is a
program-level non-success at the app boundary, with bounded deterministic
stderr and no stdout. It is not a catchable source-level typed error in Phase
26. A direct near-`U64::MAX` witness is needed because accepted sum inputs
cannot overflow.

## the agent's Discretion

- Select the concrete scalar lattice, worklist, deterministic analysis bound,
  CFG lowering, and independent validator derivations.
- Select stable diagnostics and how repeated event occurrences are identified.
- Choose exact grammar/formatter productions for the ratified source forms.
  The recommendation is familiar infix `+`, `<`, and block `if/else`.
- The user approved the specialist recommendations as a set; the earlier
  explicit answer selected mutable scalar bindings.

## Deferred Ideas

- Source-catchable overflow if a named consumer needs recovery and continuation.
- Phase 27 equality, remainder, exact FizzBuzz output, and full dual-host
  evidence; broader arithmetic and iterator/control constructs remain outside
  this phase.
- A checksum or JSON consumer until a named user program justifies its byte
  and module abstractions.

