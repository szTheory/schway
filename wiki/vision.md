---
id: vision
title: Vision and design principles
summary: A general-purpose language optimized for AI authorship, human audit, fast evidence, and production reliability.
type: vision
status: candidate
confidence: high
created: 2026-09-02
updated: 2026-09-02
tags: [vision, principles, ai-first, human-auditable]
related: [design-atlas, feature-coherence, effects-and-capabilities, compiler-feedback-latency, compute-efficiency-constitution, runtime-profiles-dogfooding, domain-data-distribution, blind-spots-boundaries, context-telemetry-security, convergence-work-program, example-tour]
---

# Vision and design principles

## One-sentence vision

Create a general-purpose language in which AI can produce and repair software
reliably, while expert humans can rapidly audit its architecture, behavior,
cost, and failure modes.

## Design center

**Intent:** AI will increasingly author most source code. Humans will review,
audit, diagnose, and occasionally edit it. The source is therefore a durable
shared representation between machine reasoning and human reasoning, not just
text optimized for manual typing.

**Synthesis:** “AI-first” should primarily mean *evidence-first*. A compact
syntax saves inference tokens once. A precise compiler, fast tests, structured
diagnostics, and live causal traces prevent or shorten many later inference
rounds.

The north-star loop is:

```text
intent
  -> constrained source
  -> static obligations
  -> executable specifications
  -> hermetic build
  -> supervised runtime
  -> causal evidence
  -> structured repair
```

## Priority order

When values conflict, use this provisional order:

1. Semantic correctness and explicit failure behavior.
2. Fast, structured feedback for AI and humans.
3. Human auditability and architectural legibility.
4. Production reliability, security, and operability.
5. Predictable performance and resource cost.
6. Interoperability and a practical adoption path.
7. Expressive power and concision.
8. Ease of implementing the language itself.

This order is a candidate, not a locked decision. It intentionally values total
human-and-machine time over keystrokes.

## Principles

### 1. Semantic density, not character density

Spend tokens when they remove ambiguity: names, domain types, effect sets,
units, invariants, and error variants. Eliminate punctuation, ceremony, and
repeated configuration that contribute no new information.

### 2. One canonical surface

Use a formatter-owned representation, deterministic import ordering, stable
name resolution, and minimal synonymous syntax. Canonical source enables stable
generation, smaller diffs, semantic caching, and reliable automated rewrites.

### 3. Functional core, explicit effects

Immutable values, algebraic data types, exhaustive pattern matching, and total
functions are the default. Mutation, I/O, nondeterminism, time, secrets,
networking, and foreign code remain available but visible through effects and
capabilities.

### 4. Make illegal architecture difficult

The compiler knows the component and dependency graph. Cycles, forbidden layer
edges, ambient capabilities, undeclared dependencies, and transitive imports
are rejected. Architecture policy should be inspectable data, not naming
convention folklore.

### 5. Locality over indirection

Keep a component's public contract, examples, properties, dependencies, and
operational expectations close enough to inspect together. Avoid annotation
scavenger hunts, magical discovery, and configuration spread across unrelated
files.

### 6. Convention over configuration, with typed escape hatches

The default project builds, tests, formats, documents, packages, observes, and
ships without assembling a toolchain. Escape hatches are explicit, local,
capability-scoped, searchable, and carry a rationale. Unsafe power should have
an audit trail and a narrow blast radius.

### 7. Specifications form a ladder

Examples, tables, boundary cases, properties, state machines, contracts, and
proofs should share data and tooling. Public APIs surface missing obligations.
Counterexamples become permanent regression cases.

The system may derive inputs and obligations from types; it cannot infer
business truth. Generated tests without a meaningful oracle are scaffolding,
not evidence.

### 8. Expected failure is data

Expected domain and infrastructure failures use typed results. Exhaustive
matching exposes every supported failure. Panics are reserved for violated
internal invariants and trigger supervised recovery, not routine branching.

### 9. Concurrency is structured; services are supervised

Tasks have parents, lifetimes, cancellation, deadlines, resource budgets, and
trace context. Long-lived state is isolated behind actors. Supervision policies
make restart and escalation explicit. Unbounded orphan work is not the happy
path.

### 10. Observe semantics, not log strings

Effect boundaries, task creation, messages, retries, state transitions, errors,
allocations, and external calls produce typed events with causal context. Logs,
metrics, traces, profiles, and replay data are projections of the same runtime
event model.

### 11. Agents receive structured truth

The toolchain exposes the syntax tree, symbol graph, types, effect graph,
dependency graph, diagnostics, change impact, test obligations, runtime tree,
and traces through a versioned protocol. Human output is a rendering of this
model, not the only interface.

### 12. Optimize the change, not just the build

The compiler is incremental and query-based. It should verify the smallest
affected slice, report the blast radius before edits, and return diagnostic
deltas after edits. Hermetic inputs enable content-addressed local and CI cache
reuse.

### 13. Dependencies have a carrying cost

Imports are direct and explicit. The lock graph, licenses, provenance,
capabilities, build scripts, binary size, compile time, and known vulnerabilities
are queryable. Projects may enforce budgets for dependency depth and cost.
Copying a tiny stable implementation can be cheaper than adopting a framework,
but copied code must retain license and provenance.

### 14. Batteries are curated, not fused to the compiler

Keep the semantic core small. Ship a coherent first-party toolkit for common
work—serialization, HTTP, testing, telemetry, time, crypto interfaces, data
access, actors, and deployment—but version most kits independently. Promote
libraries into the supported set only after real use stabilizes their shape.

### 15. Interop is a boundary with a trust level

Support the C ABI for reach and WebAssembly components for typed, capability-
restricted composition. Foreign calls declare safety, ownership, threading,
blocking, and error contracts. Unsafe adapters cannot silently contaminate the
rest of the program's guarantees.

### 16. One language, multiple explicit profiles

The first implementation is native-capable and includes usable ownership,
lexical resources, C interop, and native execution. Distributed services and AI
agents remain the first production application thesis; `beam`, `wasm`, and
`embedded` profiles follow a shared conformance model. Profiles may restrict
effects and allocation; they must not quietly change program meaning.

### 17. Operational cost is part of correctness

CPU, memory, allocation, network, storage, inference tokens, retries, and
latency are observable resources. Budgets can be tested and enforced at
component boundaries. Defaults should respect container limits without hiding
the underlying tradeoff.

### 18. Evolution is designed before 1.0

Separate language edition, compiler version, runtime ABI, standard interfaces,
and toolkit package versions. Provide machine-applicable migrations. Code from
different supported editions must interoperate through one canonical IR.

### 19. Metaprogramming must preserve tooling

Prefer ordinary functions, typed derivation, declarative schemas, and compiler
plugins over arbitrary compile-time execution. Expansion is deterministic,
inspectable, cacheable, capability-restricted, and produces source maps.

### 20. Features must compound

A feature earns core-language status when it strengthens multiple goals at
once. Typed effects are attractive because they may improve architecture,
testing, security, telemetry, distribution, and agent reasoning simultaneously.
Niche convenience alone belongs in a library.

### 21. Context is narrow, typed, and distrustful

Deadlines, cancellation, trace identity, and bounded execution budgets may flow
with structured work. Tenant, principal, and other authorization-relevant facts
become explicit domain/application inputs after validation. Arbitrary values do
not become ambient merely to save parameters, and propagated metadata is never
trusted solely because it arrived in context.

### 22. Operational defaults carry provenance

Deploy configuration, secrets, provider bindings, build identity, release
identity, telemetry policy, and generated artifacts are linked in an inspectable
evidence graph. Development conveniences such as `.env`, in-memory adapters, and
hot reload cannot silently become production behavior.

### 23. Feedback latency is an architectural budget

The fast development lane is designed before expensive language features are
accepted. Public semantic boundaries, query invalidation, type-level
computation, specialization, macros, backends, and generated evidence all spend
from a measured latency budget. Optimize time-to-correct, and never disguise a
deferred obligation as a completed check.

### 24. Domain meaning is first-class; architecture brands are not

Nominal domain types, commands, events, state transitions, invariants, read and
write authority, consistency, and schema evolution should be inspectable. The
compiler does not mandate DDD, CQRS, event sourcing, Rails, or clean architecture
as universal doctrine. Opinionated official kits project those patterns through
general semantics when their operational cost is justified.

### 25. Design for time, overload, and change

Partial failure, deadlines, retries, backpressure, bounded queues, schema/state
migrations, mixed-version operation, backup, restore, and rollback are normal
production states. They belong in the language system's evidence model even
when their implementations live in runtime or toolkit layers.

### 26. Begin as a verified modular monolith

The service profile defaults to one deployable and one primary transactional
boundary with compiler-verified internal components. Components form a DAG,
expose ports, own data, and cannot reach through one another's internals.
Distribution is introduced only with explicit partial-failure, compatibility,
security, and operational contracts.

### 27. Progressive assurance protects the fast lane

Sound shapes, effects, and architecture checks run during ordinary editing.
Contracts, generated properties, bounded model checks, and proofs add explicit
evidence tiers. Expensive assurance is cached and selectable, while release
status never implies that a deferred obligation passed.

### 28. Failure is contained at the smallest honest boundary

Expected absence and failure are typed values. Cancellation remains structured
control flow. Violated internal invariants terminate the owning task or actor
and are handled by supervision; external partial failure is handled by visible
resilience or workflow policy rather than catch-all exceptions.

### 29. Human source and machine semantics share one program

Humans receive a calm canonical textual projection. Agents can query and edit a
stable typed/effect graph instead of reconstructing it from files. Alternate
projections may exist, but they cannot change name resolution, behavior, or
program identity.

### 30. Spend compute deliberately; waste none

Do not ration analysis, tests, optimization, or AI inference when they deliver
meaningful correctness, evidence, or user value. Measure the whole cost in
human wait, model iterations, wall/CPU time, memory, disk, network, and runtime
resources. Eliminate duplicate work, narrow affected scope soundly, choose
appropriate algorithms and representations, schedule within resource budgets,
and cache only when reuse is cheaper than recomputation. See the
[compute-efficiency constitution](compute-efficiency-constitution.md).

### 31. Dogfood for compounding value, not symbolism

Move executable examples, official modules, evidence runners, diagnostics,
cache inspection, formatting, and build tooling into Lang as soon as each slice
is trustworthy and shortens the next development loop. Keep a host-language
oracle and reproducible bootstrap until staged self-hosting is cheaper and safer
than maintaining the bridge. A compiler rewrite earns priority through measured
development leverage, not language-project tradition. See [Runtime profiles and
dogfooding](runtime-profiles-and-dogfooding.md).

## What “slop-resistant” can mean objectively

The compiler cannot enforce taste, but it can make these conditions visible or
illegal:

- undeclared or cyclic dependencies;
- unused or over-broad capabilities;
- hidden I/O, global mutation, or nondeterminism;
- non-exhaustive state and error handling;
- boolean blindness and primitive obsession at public boundaries;
- oversized public surfaces and accidental visibility;
- duplicated structures or semantically equivalent branches;
- unbounded concurrency, retries, queues, or resource use;
- ignored results, placeholder panics, stale suppressions, and dead code;
- public behavior with no example, property, contract, or explicit waiver;
- dependency or build-cost budget regressions;
- generated edits lacking verification evidence.
- unvalidated boundary data entering nominal domain code;
- detached tasks, unbounded default mailboxes, or implicit remote calls;
- duplicate JSON fields or ambiguous boundary coercions under strict policy;
- public identifiers that are visually confusable or violate project naming policy.

Warnings should be high precision. A noisy quality gate teaches both humans and
agents to suppress the instrument.

## Non-goals for the first implementation

- Replace C, CUDA, Python, JavaScript, and the BEAM ecosystem immediately.
- Put model-provider APIs or ORM syntax into the language grammar.
- Prove arbitrary business requirements automatically.
- Guarantee “perfect architecture” independent of context.
- Make every workload equally fast through one memory/runtime strategy.
- Allow unrestricted macros merely because the compiler can execute them.
