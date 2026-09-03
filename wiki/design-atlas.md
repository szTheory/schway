---
id: design-atlas
title: Design-decision atlas
summary: A stakeholder-indexed map of the major language, compiler, runtime, ecosystem, and governance decisions.
type: atlas
status: draft
confidence: mixed
created: 2026-09-02
updated: 2026-09-02
tags: [decision-map, tradeoffs, architecture, evaluation]
related: [vision, language-lessons, feature-coherence, effects-and-capabilities, type-system-failure-semantics, syntax-cognitive-ergonomics, compiler-feedback-latency, application-architecture-data, concurrency-memory, domain-data-distribution, blind-spots-boundaries, context-telemetry-security, open-questions]
---

# Design-decision atlas

## How to use this atlas

Each choice must be evaluated as a system. A locally elegant feature can create
global costs in compiler complexity, inference time, build caching, FFI safety,
runtime predictability, or human audit.

For every candidate, record:

1. Which failure it prevents.
2. Which guarantees are static, dynamic, tested, proved, or merely conventional.
3. What new complexity and escape hatches it creates.
4. What structured evidence an agent receives.
5. What experiment could falsify the choice.

## Accountability lenses

| Lens | Optimizes for | Asks first | Typical blind spot |
|---|---|---|---|
| AI author | Few reliable edit/repair iterations | Can I query exact constraints and apply a local repair? | Generated code can pass shallow checks while missing intent |
| AI evaluator | Grounded, discriminating evidence | What oracle separates plausible from correct? | Benchmarks can reward familiar training data |
| Human reviewer | Fast comprehension and trust calibration | What changed semantically and what can it affect? | Compact diffs can hide broad generated behavior |
| Application engineer | Delivery speed without latent debt | Is the paved road sufficient for ordinary work? | Escape hatches accumulate into a second language |
| Language designer | Orthogonal, teachable semantics | Does this feature compose or duplicate another mechanism? | Elegant calculus can produce poor everyday ergonomics |
| Compiler engineer | Correctness, incrementalism, diagnostics | Can this be checked locally, cached, and explained? | Compiler architecture may constrain language design prematurely |
| Runtime engineer | Scheduling, memory, isolation, introspection | What invariants must hold under failure and load? | Fast microbenchmarks may ignore tail latency and operability |
| Software architect | Explicit boundaries and change containment | Can forbidden dependencies be represented and rejected? | One mandated architecture can misfit small or unusual systems |
| SRE/operator | Predictable failure, cost, and recovery | Can I explain this incident from production evidence? | Instrumentation itself can add cost, cardinality, or risk |
| Security engineer | Least authority and auditable trust boundaries | What ambient power or unsafe code exists? | Safety controls can drive users toward opaque bypasses |
| Performance engineer | Mechanical sympathy and measurable budgets | Where are allocation, copying, contention, and cache misses? | Manual optimization can damage clarity and portability |
| Ecosystem steward | Stability, supply-chain health, coherent libraries | Who maintains this dependency and how does it evolve? | Central curation can bottleneck experimentation |
| FinOps/cost owner | Total compute, storage, network, and inference cost | What does this change cost per build and per request? | Cheap runtime choices may consume expensive human time |
| Release engineer | Reproducibility and safe migration | Can the exact artifact and inputs be reconstructed? | Hermeticity can create heavy configuration |

## Cross-lens scorecard

This is a hypothesis map, not a verdict. `++` means strong expected leverage;
`--` means a material risk requiring evidence.

| Candidate mechanism | AI | Audit | Architecture | SRE | Performance | Compiler cost |
|---|---:|---:|---:|---:|---:|---:|
| Canonical AST-backed source | ++ | ++ | + | + | + | - |
| Static algebraic types | ++ | ++ | + | + | + | - |
| Typed effects/capabilities | ++ | ++ | ++ | ++ | + | -- |
| Typed provider/context graph | ++ | ++ | ++ | ++ | neutral | - |
| Exhaustive pattern matching | ++ | ++ | + | + | neutral | - |
| Executable contracts | ++ | ++ | + | + | neutral | -- |
| Structured concurrency | + | ++ | ++ | ++ | + | - |
| Supervised actors | + | + | ++ | ++ | workload-dependent | -- |
| Content-addressed builds | + | + | + | + | ++ build-time | -- |
| Restricted typed derivation | ++ | + | + | neutral | + | - |
| Unrestricted macros | + | -- | -- | -- | uncertain | -- |
| General aspect weaving | - | -- | -- | - | uncertain | -- |
| C ABI boundary | + | neutral | + | - | ++ | - |
| Wasm component boundary | + | ++ | ++ | + | -/neutral | -- |
| Built-in agent protocol | ++ | ++ | + | ++ | + | -- |
| Bidirectional local inference | ++ | ++ | + | neutral | + build-time | - |
| Verified modular monolith | + | ++ | ++ | ++ | + | - |
| Strict typed boundary codecs | ++ | ++ | ++ | + | neutral | - |
| UUIDv7 identity provider | + | + | + | + | + storage | neutral |
| Per-object memory strategy | - | -- | - | -- | uncertain | -- |

## Compounding bundles

### Effect spine

```text
effect types
  + capabilities
  + handlers
  + architecture policy
  + test substitution
  + automatic spans
  + remote handlers
```

One semantic mechanism may replace DI containers, mocking frameworks, ambient
permissions, tracing wrappers, and some RPC scaffolding.

The provider graph supplies construction and lifetime; effects describe the
authority a function can exercise. These are related but not interchangeable.

### Evidence spine

```text
ADTs + exhaustive matching + contracts + properties + shrinking + mutation checks
```

Types enumerate the valid shape; contracts state local truth; generators explore
the shape; shrinking produces a minimal witness; mutation analysis asks whether
the oracle notices meaningful faults.

### Change spine

```text
canonical AST + stable symbol IDs + query compiler + semantic diff + content hashes
```

This supports precise context retrieval, transaction-like edits, incremental
verification, reproducible caches, and low-noise review.

### Operations spine

```text
structured tasks + isolated actors + supervision + propagated context + typed events
```

This aligns source-level ownership, runtime failure domains, distributed traces,
and incident explanation.

## Decision inventory

### Source representation and syntax

Questions to answer:

- Is text the authority, or is an AST/code database authoritative with text as
  a deterministic projection?
- Must parse/format/parse preserve identity exactly?
- Are function arguments always named? Are operators, pipelines, constructors,
  indexing, and closures exceptions?
- Are blocks delimited by braces, indentation, keywords, or a uniform tree?
- How much type information is explicit at public boundaries versus inferred?
- Are imports always qualified and symbol-level?
- Can the language represent semantic diffs independent of formatting?
- Are comments attached to syntax nodes so automated rewrites preserve them?
- Is there exactly one string interpolation, collection, and call syntax?
- May an exact-name argument use `label:` for `label: label`, and how does the
  formatter handle shadowing and renames?
- Do effectful calls carry a visible marker in addition to their checked effect
  row?
- Does Unicode improve domain expression enough to justify confusables and
  input/tooling cost?

Candidate default: conventional expression syntax, named calls, canonical
formatting, stable AST identities, ASCII operators, and semantic-diff support.
Do not choose Lisp syntax solely for parser convenience; benchmark human and AI
performance on equivalent programs.

Footguns: too many equivalent forms, whitespace-sensitive meaning that formats
poorly, user-configurable formatting, implicit imports, operator overloading
without semantic constraints, and macros that rewrite uninspectable syntax.

Use the admission gate and interaction war game in
[Feature coherence](feature-coherence.md) before promoting a syntax proposal.

### Types and data modeling

Questions to answer:

- Hindley–Milner-style inference, bidirectional typing, local inference, or
  explicit annotations?
- Nominal, structural, or hybrid records and interfaces?
- Algebraic data types and exhaustive pattern matching as the universal state
  model?
- Newtypes, opaque types, refinement types, units of measure, dependent types,
  or contracts?
- Is `null` absent in favor of `Option`?
- Are numeric conversions checked and explicit? What are overflow semantics?
- Are collections persistent, mutable, or chosen by profile?
- Does dynamic typing exist at typed boundaries such as JSON and foreign data?
- How are schema evolution and backward compatibility represented?
- Can public types expose representation accidentally?

Candidate default: nominal public/domain types, structural consumer-owned
behavioral interfaces and private records where safe, ADTs, no null, checked
conversions, local inference, explicit public signatures, opaque domain types,
and refinements/contracts added progressively.

Footguns: structural compatibility that accepts semantically different values,
boolean blindness, implicit numeric narrowing, bottom/null inhabiting every type,
and type-level programming whose diagnostics require a specialist.

### Errors, partiality, and totality

Questions to answer:

- Which failures are `Result`, cancellation, panic, process exit, or effect?
- Are error sets inferred as rows or named explicitly?
- Can a function be declared total and terminating?
- How are resource cleanup and rollback guaranteed?
- Do retries belong to callers, handlers, supervisors, or policy?
- Can ignored errors compile?
- How are foreign exceptions translated?

Candidate default: typed results for expected failure, exhaustive handling at
component boundaries, scoped cleanup, panic for violated invariants, and
supervision for process failure.

Footguns: exceptions invisible in signatures, catch-all recovery, string errors,
automatic retries without idempotency, and treating cancellation as failure.

### Effects, capabilities, and dependency injection

Questions to answer:

- Algebraic effect rows, capability parameters, or both?
- Are effects inferred internally and explicit on public APIs?
- Are handlers one-shot or multi-shot; lexical, dynamic, actor-local, or task-local?
- Can handlers resume, fork, migrate, or retry computations?
- How do effect polymorphism and higher-order functions stay readable?
- Which effects are built in: state, allocation, I/O, time, random, panic,
  async, trace, unsafe, foreign?
- Can architecture policy allow an abstract capability while forbidding a
  concrete adapter?
- How does an effect cross an FFI or network boundary?

Candidate default: typed capability/effect rows, lexical handlers, explicit
public effects, inference inside implementations, and limited continuation power
until benchmarks justify more.

Footguns: effect-row noise, effect masking, handler order changing semantics,
multi-shot continuations duplicating resources, and “pure” wrappers around
unsafe foreign behavior.

### Contracts, tests, and evaluation

Questions to answer:

- Which contracts run in development, production, or proof mode?
- Can types derive useful boundary generators and shrinkers?
- How are examples colocated without bloating implementation views?
- What must every public API specify: examples, error coverage, properties, or
  explicit waiver?
- How are time, randomness, concurrency, and external effects made deterministic?
- Are state-machine and model-based tests first class?
- Can failing generated cases become stable regression fixtures automatically?
- Is mutation testing incremental and risk-targeted?
- Can the toolchain admit “no oracle exists” rather than generate tautologies?
- How are flaky and performance-sensitive tests quarantined without being ignored?

Candidate default: examples and properties colocated with public contracts;
unit, integration, fuzz, mutation, and proof modes in one runner; deterministic
effect handlers; changed-code test selection; saved minimal counterexamples.

Footguns: generated assertions that restate the implementation, coverage as a
correctness proxy, global mocks, snapshot tests for unstable noise, and making
all contracts production checks regardless of cost.

### Modules and architecture

Questions to answer:

- What is the unit of encapsulation, compilation, deployment, supervision, and
  ownership—and should these align?
- Are dependency cycles always forbidden or only across components/layers?
- Are ports ordinary capability interfaces?
- Is hexagonal layering fixed by the language or declared as project policy?
- Can tiny programs avoid architecture ceremony?
- Are imports limited to direct declared dependencies?
- How are friend/internal APIs expressed without ecosystem leakage?
- Can the compiler compute change impact and ownership from the graph?

Candidate default: packages contain components; components expose typed ports;
the component graph is acyclic; internal module cycles may be collapsed or
rejected; policy presets offer hexagonal architecture without hard-coding its
vocabulary into semantics.

Footguns: mandatory enterprise layering for scripts, dependency injection by
runtime reflection, barrel modules that hide actual dependencies, and cyclic
initialization.

### Concurrency and distribution

Questions to answer:

- Structured tasks, actors, channels, data parallelism, or a composable subset?
- Is shared mutable memory forbidden, isolated, ownership-checked, or lock-based?
- Are actor message protocols typed and versioned?
- How are backpressure, mailbox bounds, deadlines, cancellation, and fairness
  represented?
- What failure propagation and restart strategies are defaults?
- Is location transparent, or must remote latency/failure remain syntactically
  visible?
- Can tasks migrate with code and data? What authority follows them?
- How are delivery semantics, idempotency, ordering, and deduplication specified?
- Can deterministic replay cover scheduling and external effects?

Candidate default: structured concurrency for finite work; supervised actors for
long-lived isolated state; immutable or uniquely owned messages; explicit
`remote` boundary; bounded mailboxes and deadlines by default.

Footguns: orphan goroutines, unbounded queues, silent actor death, distributed
calls that look local, global executors, blocking foreign calls on schedulers,
and shared state disguised by a thread-safe container.

See [Domain, data, and distributed-system semantics](domain-data-and-distribution.md)
for CQS/CQRS, consistency, event sourcing, resilience policies, storage,
deterministic simulation, upgrades, and recovery.

### Memory and resource model

Questions to answer:

- Tracing GC, per-actor GC, ARC, ownership/borrowing, regions, arenas, or profile-
  dependent lowering?
- Which guarantees must be source semantics versus backend optimization?
- Can most application code ignore memory while hot paths opt into ownership?
- How are cycles handled under reference counting?
- Are allocation and copying visible effects or measured costs?
- Can values cross actors or FFI without copying safely?
- What are latency, footprint, and real-time guarantees per profile?
- How does the runtime respect container/cgroup limits?

Candidate direction: immutable values plus ownership inference; managed per-
isolate heaps for the service profile; explicit regions/arenas for native and
embedded profiles; no raw pointer outside audited foreign/unsafe components.
This requires prototypes before commitment.

Footguns: one global heap for independently failing services, invisible large
copies, ARC cycles, borrow annotations infecting application code, finalizers,
and a “soft” memory limit presented as a hard guarantee.

### Runtime and observability

Questions to answer:

- Is telemetry always on, sampled, compiled out, or profile-controlled?
- Which semantic operations emit events automatically?
- Can events be correlated across tasks, actors, nodes, foreign calls, builds,
  tests, and model inference?
- What is safe to inspect or mutate in production?
- How are cardinality, PII, secrets, retention, and observer overhead bounded?
- Can an agent request a focused trace without globally increasing noise?
- Does every failure retain a bounded causal packet for replay?
- How are scheduler, GC, allocation, mailbox, and dependency costs exposed?

Candidate default: typed runtime events, automatic causal context propagation,
low-cost ring buffers, tail/error-triggered capture, OpenTelemetry export, and a
read-only agent interface by default.

Footguns: log strings as APIs, high-cardinality labels, accidental baggage/PII
propagation, observer effects, always-on full payload capture, and remote debug
ports with ambient authority.

### Metaprogramming, parsing, and DSLs

Questions to answer:

- Can ordinary functions plus typed schemas replace most macros?
- Should parsing expressions or grammars be first-class library values?
- Are derivations declarative and hygienic?
- Can generated code be inspected, formatted, attributed, and cached?
- May plugins perform I/O or execute arbitrary host code during compilation?
- How are plugin/compiler version compatibility and sandboxing handled?

Candidate default: parser combinators or typed grammar DSL in the official
toolkit; restricted hygienic derivation over typed AST; plugins as capability-
sandboxed Wasm components. Avoid user-defined syntax initially.

Footguns: Raku-scale context sensitivity in a language seeking canonicality,
Ruby-style method-missing magic across public APIs, unhygienic macros, and
networked build scripts.

### Interoperability

Questions to answer:

- Is C ABI support source-level, generated bindings, header translation, or all
  three?
- Who owns memory across the boundary?
- How are callbacks, panics, exceptions, threads, blocking, and async translated?
- Which types have stable layouts and ABIs?
- Are foreign components capability-isolated?
- Is Wasm/WIT a primary package boundary or deployment target only?
- Can BEAM terms/processes be used without erasing type guarantees?

Candidate default: narrow generated C bindings with explicit ownership contracts;
WIT-like rich interfaces for safe components; BEAM interop in the first service
backend; subprocess/port isolation preferred for crash-prone native libraries.

Footguns: claiming FFI safety from signatures alone, callbacks after owner
destruction, foreign exceptions crossing frames, blocking the scheduler, and
ABI stability freezing poor representations.

### Build, packages, and standard toolkit

Questions to answer:

- Is one official command responsible for format/check/test/build/run/doc/package?
- Are builds hermetic and network-disabled after dependency resolution?
- Are toolchains pinned as dependencies and automatically acquired?
- Is the graph content-addressed; can local and CI caches interoperate?
- Are direct and transitive dependencies separately budgeted?
- Are build-time plugins more privileged than runtime packages?
- Can dependencies be vendored and audited offline?
- What qualifies a library for `core`, `official`, `incubator`, or community status?
- Can official libraries version independently without compatibility chaos?
- What warm/cold latency budgets apply to parse, check, affected tests, reload,
  and release at representative repository sizes?
- How are semantic query dependencies and invalidations explained?
- Which proof, fuzz, optimization, and model-check obligations may be deferred
  from the sound edit lane, and how is that incompleteness represented?
- Can parallel Git worktrees share immutable cache artifacts without sharing
  mutable compiler state?

Candidate default: one tool; lockfile with complete graph/provenance; direct
imports only; hermetic sandboxed builds; local content-addressed cache first;
optional remote cache; small stable core plus independently versioned official
kits on a coordinated release train.

Footguns: implicit network access, arbitrary install scripts, feature unification
at a distance, dependency diamonds with differing semantics, a standard library
that cannot evolve, a fragmented “official” ecosystem, coarse invalidation,
specialization explosion, and fast diagnostics that omit enough cause to repair.

See [Compiler and feedback latency](compiler-and-feedback-latency.md) for the
latency-first compiler proposal.

### Versioning and governance

Questions to answer:

- What compatibility promise begins at 1.0?
- Can editions alter surface syntax while sharing one IR and ecosystem?
- How long are compiler, runtime, ABI, toolkit, and protocol versions supported?
- Are migrations machine-generated and semantically checked?
- What evidence must a feature proposal include?
- Who can expand the language, official toolkit, or unsafe surface?
- How are deprecations measured against ecosystem telemetry without violating privacy?

Candidate default: opt-in editions, cross-edition linking, explicit language and
toolchain requirements, long deprecation windows, automatic migrations, and a
feature budget requiring use cases, alternatives, interaction analysis,
implementation cost, and removal strategy.

Footguns: semantic versioning that ignores behavior, compiler and package
versions conflated, permanent experimental syntax, and governance driven by
novelty rather than compounding leverage.

### AI inference as a platform capability

Questions to answer:

- Is model inference merely a library effect or a scheduler/runtime primitive?
- How are structured outputs, streaming, tool calls, retries, caching, budgets,
  privacy, and provenance represented?
- Are model/provider identities values, capabilities, deployment configuration,
  or types?
- Can deterministic tests substitute recorded or model-based handlers?
- How does telemetry record tokens and cost without capturing sensitive prompts?
- Can local accelerators be used without putting tensor semantics in the core?

Candidate default: an official `Model` capability with typed schemas, streaming,
cost budgets, cache policy, and test handlers. Keep provider APIs and tensor
kernels outside the language. Promote runtime scheduling only after measured
workloads show cross-cutting leverage.

Footguns: provider-specific grammar, nondeterminism hidden behind a pure-looking
function, retries that multiply cost, caches that leak private inputs, and tests
whose only oracle is another model.

## Global anti-patterns

- **Kitchen-sink syntax:** every borrowed feature adds another way to express the
  same concept.
- **Framework in the grammar:** ORM, HTTP, containers, and current AI APIs evolve
  faster than language editions.
- **Static-or-nothing ideology:** expensive proof obligations applied to low-risk
  glue encourage bypasses.
- **Runtime magic:** automatic behavior with no type, graph, or trace representation.
- **Escape-hatch normalization:** `unsafe`, dynamic calls, suppression, and raw
  foreign access become routine.
- **Metric theater:** line coverage, type coverage, or lint counts treated as
  evidence of business correctness.
- **Universal backend fiction:** service, HPC, browser, and microcontroller
  profiles quietly acquire different semantics.
- **Agent-only representation:** token savings make code hostile to incident
  response and review.
- **Human-only diagnostics:** tools force agents to scrape unstable prose and
  terminal formatting.
