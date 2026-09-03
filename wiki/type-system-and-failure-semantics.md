---
id: type-system-failure-semantics
title: Type system and failure semantics
summary: A fast, progressive type system that makes domain meaning, effects, and failure obligations precise without turning ordinary application code into proofs.
type: design
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-03
tags: [types, errors, assurance, compiler, ai]
related: [vision, semantic-kernel-contract, semantic-kernel-probes, trust-validation-information-flow, boundary-data-validation-persistence, effects-and-capabilities, feature-coherence, compiler-feedback-latency, domain-data-distribution, research-ledger, open-questions]
---

# Type system and failure semantics

## Recommendation

Use a sound static type system with bidirectional local inference, algebraic
data types, exhaustive pattern matching, nominal domain types, structural
behavioral ports, row-polymorphic effects, and progressive refinements.

Keep the ordinary edit lane decidable, predictable, and fast. Contracts,
refinements, model checking, and proofs form higher assurance tiers whose
obligations are explicit and cacheable. Do not make arbitrary theorem proving,
global inference, or unconstrained type-level computation prerequisites for a
sub-second local check.

This is the type-system equivalent of the language's larger thesis: put enough
meaning in the program that the compiler can be a strong critic, but never hide
an expensive or heuristic critic behind an apparently simple compile command.

The executable v0 rules—including evaluation order, public inference
boundaries, ownership-aware structural conformance, cancellation, panic, and
numeric validity—are consolidated in the [semantic kernel
contract](semantic-kernel-contract.md). This note retains the broader type and
assurance rationale.

## Candidate semantic core

| Mechanism | Job | Why it earns its place | Principal cost or footgun |
|---|---|---|---|
| Algebraic data types | Enumerate valid shapes and states | Powers errors, protocols, state machines, generators, and exhaustive matching | Large variants can become catch-all dumping grounds |
| Exhaustive matching | Prove case coverage over closed data | Produces local, actionable feedback after evolution | Wildcards can defeat evolution checks |
| Opaque nominal types | Separate equal representations with different meaning | Prevents ID, currency, tenant, and trust-boundary confusion | Wrapper/conversion ceremony if inference and derivation are weak |
| Structural ports | Express the behavior a consumer needs | Supports consumer-owned interfaces and tiny test doubles | Accidental conformance if used for domain identity or security authority |
| Parametric generics | Reuse algorithms without erasing types | High leverage for collections, results, effects, and adapters | Specialization can explode compile time and binary size |
| Effect rows | Expose authority and nondeterminism | Unifies DI, testing, sandboxing, architecture, and telemetry | Effect soup and confusing handler resolution |
| Affine resources | Prevent double close/use-after-move at resource boundaries | Useful for files, sockets, buffers, FFI, and native code | Explicit lifetime syntax can tax ordinary service code |
| Refinements/contracts | State value predicates and postconditions | Improve validation, generation, runtime checking, and optional proof | Solver latency, undecidability, annotations, false confidence |
| Dynamic boundary value | Accept unknown foreign data before validation | Necessary for JSON, reflection, migration, and interop | Dynamic values spreading into domain code |

Source influence, domain validation, output encoding, confidentiality, and
authority remain distinct. The candidate mechanics and tiered analysis are in
[Trust, validation, and information flow](trust-validation-and-information-flow.md).
They reuse nominal types and capability contracts rather than add a universal
`trusted` coercion.

Bidirectional typing divides work into synthesizing a type from an expression
and checking an expression against an expected type. It is a promising basis
for local inference and better error locality without requiring complete
annotations everywhere. The approach is well established across increasingly
expressive type systems, but this project's exact rules still need a prototype.
([Bidirectional Typing](https://arxiv.org/abs/1908.05839))

## Nominal meaning, structural behavior

The candidate rule is intentionally asymmetric:

- public domain records, variants, identities, validated values, units, and
  security witnesses are nominal;
- small consumer-owned capabilities and ports may be satisfied structurally;
- anonymous structural records are allowed inside private implementation code;
- public structural compatibility never silently grants authority;
- a provider's extra methods do not become visible through a narrower port;
- effect and failure signatures participate in behavioral conformance;
- ambiguous or accidental conformance at public boundaries requires an explicit
  witness.

This takes the useful part of Go's implicit interface satisfaction without
inheriting TypeScript's deliberate unsoundness. It also keeps hexagonal design
local: a consumer declares the narrow behavior it uses; a provider does not
import the consumer merely to announce conformance.

```text
port LoadOrder {
  get(id: OrderId) -> Result<Option<Order>, LoadError>
    with { Database.Read }
}

record PostgresOrders { pool: Pool }

fn PostgresOrders.get(self:, id:) -> Result<Option<Order>, LoadError>
  with { Database.Read }
{ ... }

// Conformance is structural and compiler-visible. OrderId and TenantId are
// still incompatible even if both use the same 128-bit representation.
```

## Inference policy

Inference should remove repetition, not erase public meaning.

- Infer local binding, lambda, generic, and private effect types.
- Require parameter, result, error, and effect contracts on exported component
  operations.
- Check an implementation against its declared public shape rather than
  inferring an unstable public API from the body.
- Never infer a widening coercion, authority grant, network boundary, or
  loss-of-precision conversion.
- Explain every inserted dictionary, handler, conversion, allocation, copy, or
  specialization through the semantic query API.
- Bound inference to a module or explicit strongly connected private group;
  public components themselves form a DAG.

Type classes/traits should use explicit coherence rules. Avoid global orphan
instances, import-dependent method meaning, overlapping implicit resolution,
and user-defined implicit conversions. A derived implementation is a named,
inspectable artifact with provenance.

## Progressive assurance ladder

```text
shape -> exhaustive cases -> executable contract -> generated property
      -> bounded model check -> proof certificate
```

Each rung adds evidence without redefining the function:

1. **Shapes:** ADTs, opaque types, `Option`, and `Result` exclude broad classes
   of invalid state.
2. **Contracts:** preconditions, postconditions, and invariants can run at a
   boundary, seed generators, or become release obligations.
3. **Properties:** authored oracles run over generated and shrunk examples.
4. **Model checks:** bounded concurrency/state-machine exploration finds a
   counterexample or reports its explored bound.
5. **Proofs:** critical code may carry a certificate checked by a small trusted
   verifier.

Refinement systems such as Liquid Haskell demonstrate that predicates can be
reduced to SMT obligations, while their work also exposes the importance of
termination and annotation boundaries. Lean demonstrates a much stronger
dependent-type/proof environment and a small proof-checking kernel. Those are
inspirations for an optional assurance lane, not evidence that full dependent
typing belongs in the default application language.
([Liquid Haskell papers](https://ucsd-progsys.github.io/liquidhaskell/papers/),
[Lean reference](https://lean-lang.org/doc/reference/latest/))

## Failure taxonomy

The language should not have one undifferentiated notion of exception.

| Kind | Representation | Required handling |
|---|---|---|
| Ordinary absence | `Option<T>` | Match, default, or propagate through an optional computation |
| Expected domain/external failure | `Result<T, E>` | Match, translate with context, or explicitly propagate with `?` |
| Cancellation/deadline | Task control outcome | Preserve as control flow; do not accidentally translate to a business error |
| Defect/invariant violation | `panic` | Terminate the smallest isolation boundary and retain causal evidence |
| Actor/process death | Typed exit observed by supervisor | Restart, stop, or escalate according to declared policy |
| Fatal runtime failure | Root termination | Crash the runtime only when isolation/recovery policy cannot contain it |

An ignored `Result` is a compile error. Catch-all exception handling is not an
ordinary control-flow feature. Panic interception exists only at test,
supervisor, runtime, and foreign boundaries; application code cannot catch a
defect and quietly continue with unknown invariants.

Network timeout, overload, disconnection, and dependency unavailability are
expected failures at their boundary. They may be handled by a visible retry,
fallback, circuit, or durable-workflow policy. They do not crash the whole app.
An impossible internal state is a defect: crash the owning actor/task boundary,
record enough evidence to fix it, and let supervision apply its declared policy.
This follows the useful distinction made by Rust between recoverable `Result`
and unrecoverable panic while adopting Erlang/OTP's smaller fault-isolation
unit. ([Rust error handling](https://doc.rust-lang.org/book/ch09-00-error-handling.html),
[Erlang robustness](https://www.erlang.org/doc/system/robustness))

```text
fn settle(invoice: Invoice)
  -> Result<Receipt, SettleError>
  with { Payments, Clock }
{
  match Payments.capture(invoice:) {
    Ok(payment:) => Ok(Receipt(payment:, at: Clock.now()))
    Err(Declined(reason:)) => Err(PaymentDeclined(reason:))
    Err(Unavailable(retry_after:)) => Err(DependencyUnavailable(retry_after:))
  }
}
```

If `Payments.capture` later adds `FraudReview`, the match fails to compile. The
compiler reports the missing case, affected callers, declared contracts, and a
minimal counterexample shape in one structured diagnostic.

## What “if it compiles, it is good” can honestly mean

Compilation cannot prove that requirements are correct, product behavior is
desirable, an external service honored its contract, or a latency target will
hold in production. It can make objective slop illegal:

- unresolved placeholders, ignored results, dead branches, and stale waivers;
- unhandled variants and impossible public effect widening;
- undeclared dependency edges, component cycles, and internal API access;
- use of unvalidated boundary data as a domain type;
- resource leaks expressible in the ownership model;
- detached tasks, unbounded default queues, and lock-across-await in safe code;
- telemetry export of secrets or disallowed classifications;
- public behavior with no evidence or explicit, expiring waiver at release.

The stronger slogan is: **if a release verifies, every known obligation has
evidence, an owner, or an explicit expiring exception.** That is achievable and
more useful than pretending compilation proves taste or intent.

## Anti-patterns rejected early

- General null plus nullability inference.
- Checked exceptions as a separate hierarchy from ordinary typed results.
- Exceptions whose possible types are discovered from implementations.
- Class inheritance as the primary reuse or domain-polymorphism mechanism.
- Import-order-dependent implicit resolution.
- Arbitrary user-defined coercions and truthiness.
- Whole-program inference that turns a local edit into a global check.
- Type-level computation with no fuel, cache key, or diagnostic budget.
- A universal `Dynamic`/`Any` escape that silently contaminates trusted code.
- Wildcard matches at public protocol boundaries without a compatibility policy.

## Experiments required before acceptance

1. Implement the same service slice with nominal ports, structural ports, and
   explicit adapters; inject accidental matches and evolve each interface.
2. Compare row-polymorphic effects with explicit capability parameters on a
   large effect-heavy module.
3. Measure bidirectional-check latency and diagnostic locality after private and
   public edits.
4. Run a failure-evolution benchmark: add variants and measure AI/human repair
   accuracy with and without exhaustive matching.
5. Compare runtime contracts, SMT refinements, and proof certificates on three
   high-value invariants; record annotation and cache cost.
6. Define exactly which assurance obligations block `check`, `test`, `verify`,
   and `release`.
