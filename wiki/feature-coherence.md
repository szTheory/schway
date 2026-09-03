---
id: feature-coherence
title: Feature coherence and admission framework
summary: A gate for choosing the smallest set of orthogonal features that compounds into a powerful, explainable language.
type: design
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-03
tags: [language-design, feature-selection, syntax, coherence]
related: [vision, design-atlas, effects-and-capabilities, trust-validation-information-flow, type-system-failure-semantics, syntax-cognitive-ergonomics, compiler-feedback-latency, application-architecture-data, concurrency-memory, domain-data-distribution, example-tour, research-ledger]
---

# Feature coherence and admission framework

## Recommendation

Do not optimize for the fewest features or the longest feature checklist.
Optimize for **maximum leverage from a small set of orthogonal concepts**.

A feature earns syntax or semantic privilege only if it solves an important,
repeated problem; composes with the semantic spine; can be explained precisely
by tools; and beats a library or tooling solution after accounting for its
whole-life cost.

This makes language design a portfolio problem. Adding one feature changes the
value and cost of others. Algebraic data types make exhaustive pattern matching
far more valuable. Typed effects reduce the need for dependency-injection
frameworks and general aspect weaving. A canonical AST and formatter make
semantic diffs and source transformations cheaper. The unit of evaluation is
therefore a coherent feature cluster, not a feature in isolation.

## The admission gate

Every candidate must answer these questions before entering the core:

| Gate | Passing evidence |
|---|---|
| Job | It solves a frequent, consequential job in the target workloads. |
| Non-duplication | Existing core concepts cannot solve the job as clearly or safely. |
| Compounding leverage | It improves at least three important workflows, or one safety-critical workflow. |
| Canonicality | Common code has one obvious representation and mechanically equivalent forms normalize. |
| Static leverage | The compiler, verifier, or editor gains useful facts rather than merely new spelling. |
| Explainability | Resolution and failure can be rendered as a small causal explanation and queried as data. |
| Cost | Syntax, compiler, runtime, tooling, learning, migration, and backend costs are acceptable together. |
| Profiles | Its meaning is stable across service, native, Wasm, and restricted profiles, or restrictions are explicit. |
| Containment | Unsafe or complex escape hatches are lexical, visible, and auditable. |
| Falsifiability | A corpus, benchmark, or usability study can show that the feature failed to earn its keep. |
| Evolution | The feature can be revised or removed without silently changing program meaning. |

The default burden of proof increases as a feature moves left in this ladder:

```text
tool/editor -> community package -> incubator -> official toolkit
            -> core interface -> language semantics -> dedicated syntax
```

A broad official toolkit can make something turnkey without freezing it into
the compiler. This is the preferred strangler-fig path: incubate externally,
measure, standardize the interface, and move inward only when stability and
ubiquity justify the compatibility burden. Movement outward should remain
possible when independent versioning is healthier.

## Current feature disposition

These are candidates, not accepted decisions.

| Candidate | Current disposition | Why |
|---|---|---|
| Algebraic data types | Core semantics | Give closed data shapes, typed errors, protocol states, and derivable generators. |
| Pattern matching | Core syntax | Compounds with ADTs, errors, messages, parsers, and exhaustive state handling. |
| Typed effects and handlers | Semantic spine | Unify authority, substitution, testing, architecture, telemetry, and sandboxing. |
| Capabilities/provider graph | Core interface plus composition syntax | Replace reflective DI and make the executable dependency graph statically inspectable. |
| Structural behavioral interfaces | Core type-system candidate | Consumer-defined ports and test doubles need no provider coupling; keep domain identities nominal. |
| Named arguments | Default syntax, with exact-name punning | Improve local meaning and reorder-stable diffs; `reservation:` may canonically mean `reservation: reservation`. Symbolic operators and obvious unary calls remain under study. |
| Pipeline notation | Small syntax | Exposes data flow, but must not create placeholder or precedence dialects. |
| Contracts/specifications | Core semantics plus toolchain | Supply runtime checks, proof obligations, generation boundaries, and local evidence. |
| Structured concurrency | Core semantics/runtime | Makes task lifetime, cancellation, failure, and supervision inspectable. |
| General aspect weaving | Reject from core | Hidden join points and precedence obscure control flow and weaken local reasoning. |
| Typed effect interception | Keep | Captures useful AOP jobs at explicit semantic boundaries with typed ordering and scope. |
| Unrestricted macros | Reject initially | Create language dialects and make generated code, tooling, and audits harder. |
| Typed derivation/staging | Incubate | May remove boilerplate if expansion is deterministic, inspectable, hygienic, bounded, and formatter-owned. |
| Grammars and parser combinators | Official toolkit first | High leverage for protocols, but do not require privileged syntax until corpus evidence supports it. |
| Regex literals | Toolkit or tiny literal form | Familiar and useful, but backtracking, portability, and readability policies matter more than punctuation. |
| Data layout contracts | Native/restricted profile experiment | Mechanical sympathy matters, but ordinary service code should not carry layout ceremony. |
| Broad data-structure catalog | Official toolkit | Users need it out of the box; the compiler does not need to own most structures. |
| ORM | Opinionated official package candidate | Common job with major semantic tradeoffs; one universal ORM should not define the language. |
| AI inference | `Model` capability and official package first | Budgets, schemas, provenance, caching, and nondeterminism matter; vendor churn argues against core syntax. |
| Telemetry | Runtime semantic event substrate | Automatic boundaries are valuable, while payload and export policy must remain budgeted and configurable. |
| Perl-style taint bit | Reject as universal runtime semantics | Source-to-sink protection matters, but one bit conflates provenance, validity, encoding, confidentiality, and authority. |
| Typed trust-flow boundary | Core IR hook plus official security policy and verifier | Reuses nominal types, capabilities, and evidence; local checks stay fast while stronger global analysis is targeted. |
| Domain commands/events/aggregates | Official declarative kit first | The compiler should see their ordinary types, effects, contracts, and schemas without freezing one architecture into grammar. |
| CQRS and event sourcing | Official service kit, opt in | Powerful for selected domains; projection lag, migrations, storage, and recovery are too costly as universal defaults. |
| Retries/backoff/jitter | Typed runtime policy | Common and dangerous; require bounded attempts, one retry owner, and idempotency evidence rather than invisible automation. |
| Deterministic simulation | Runtime/toolchain priority | Clock, randomness, network, disk, and scheduling handlers can create reproducible failure worlds. |
| Built-in distributed database | Reject as a universal promise | A local table/cache and embedded store are useful; general durability and distributed recovery remain explicit provider concerns. |
| Dependent types/proofs | Progressive assurance experiment | A small checker and optional proof artifacts may add leverage; arbitrary proof search/type computation cannot tax every edit. |
| Bidirectional local inference | Core type-checker candidate | Rich local inference and checking can preserve concise source without whole-program inference. |
| Verified modular monolith | Service-profile default policy | Strong boundaries and one deployable avoid premature distributed cost while preserving an extraction path. |
| UUIDv7 | Service-toolkit identity-provider default | Good distributed generation and index locality; representation remains hidden and is neither universal nor secret. |
| JSON/schema validation | Official boundary kit | Ubiquitous and high leverage, but coercion, limits, duplicate fields, and evolution must be explicit. |
| Ecto-like changes and Repo | Official data kit | Keeps validation, queries, constraints, and persistence explicit without active-record magic. |
| Consumer contract verification | Toolchain/official test kit | Protects independently evolving ports and services; complements rather than replaces integration tests. |
| Declarative state machines | Official kit/projection first | Compile to ordinary ADTs/transitions and earn syntax only through corpus evidence. |
| GUI | Official host/capability kit later | Pure update/view semantics compose; platform behavior and accessibility remain explicit adapter concerns. |
| Per-object memory-strategy choice | Reject | Multiplies ownership, cycle, FFI, optimization, and diagnostic semantics. Use explicit profiles and lexical regions. |

## Structural behavior without structural domain identity

Go-style implicit interface satisfaction earns serious consideration for ports.
A component should be able to accept the behavior it needs without forcing each
provider to declare allegiance to a centrally owned interface. This improves
hexagonal boundaries, small test doubles, and incremental adaptation.

The rule should not make all data structural. `CustomerId`, `TenantId`,
`VerifiedPrincipal`, and `Usd` may have identical representations while carrying
incompatible meaning and authority. The coherent candidate is therefore:

- nominal opaque types and public domain data;
- structural matching for small behavioral interfaces/capabilities;
- structural anonymous records inside private implementation code;
- explicit witnesses at foreign, security, or public compatibility boundaries
  when silent conformance would be risky.

See [Domain, data, and distributed-system semantics](domain-data-and-distribution.md)
for the detailed proposal and evidence.

## Why pattern matching earns its keep

Pattern matching is not ornamental convenience here. One construct can express:

- exhaustive domain-state transitions;
- typed error recovery;
- actor and protocol message dispatch;
- destructuring of records and persistent collections;
- parsing over algebraic syntax trees;
- compiler analyses over the language's own intermediate representation.

It gives the checker facts that nested conditionals do not express as directly.
The useful subset should be intentionally constrained:

- patterns bind names once;
- exhaustiveness and unreachable arms are checked;
- closed types require exhaustive handling unless `_` is explicitly justified;
- guards are pure expressions, so matching does not hide I/O or mutation;
- matching does not invoke arbitrary user code;
- extractors are ordinary typed functions or parser constructs, not invisible
  control flow;
- the formatter gives equivalent pattern trees one canonical layout.

```text
match result {
  Ok(order:) => confirm(order: order)
  Err(PaymentDeclined { reason: }) => ask_for_payment(reason: reason)
  Err(error: Retryable) if error.attempts < 3 => retry(error: error)
}
```

Whether subtype/class patterns such as `error: Retryable` belong is unresolved;
tagged data patterns are easier to make exhaustive and portable.

## Why general AOP does not

Aspect-oriented programming was created for behavior that cross-cuts a
system's primary decomposition. The original design uses aspect programs and
weaving at join points. That identifies a real problem, but unrestricted
pointcuts let behavior attach far from the code being audited.
([Kiczales et al.](https://homepages.cwi.nl/~storm/teaching/reader/KiczalesEtAl97.pdf))

This design should recover the valuable jobs without accepting invisible
weaving:

| Cross-cutting job | Coherent mechanism |
|---|---|
| Authorization | Explicit capability plus policy handler |
| Transactions | Typed scoped handler with commit/rollback protocol |
| Retries/timeouts | Handler combinators whose order is visible and queryable |
| Telemetry | Runtime events at capability, task, actor, and boundary transitions |
| Caching | Explicit handler with key, consistency, and invalidation contract |
| Audit | Typed audit effect with mandatory retention/export policy |

Interception is allowed only at declared semantic operations. The compiler must
show the resolved handler stack in source order and machine-readable form.

## Data-oriented design and mechanical sympathy

Data-oriented design is partly a language concern, but mostly a representation,
compiler, profiler, and library concern. The native profile should explore:

- value types with predictable inline representation;
- explicit `packed`, alignment, and ABI layout only at boundaries;
- array-of-structures and structure-of-arrays views generated from one schema;
- arenas/regions and bounded allocation policies;
- iterators that reveal vectorization and allocation barriers;
- a `layout explain` command showing size, padding, cache-line crossings, and
  representation conversions;
- profiles that report locality and allocation against source-level data shapes.

Intel's optimization guidance explicitly treats improved locality and sequential,
small-stride access as key optimizations and documents AoS-to-SoA layout
tradeoffs. That supports exposing layout evidence, not forcing one layout on all
workloads. ([Intel optimization manual](https://cdrdv2-public.intel.com/821612/248966-Optimization-Reference-Manual-V1-050.pdf))

An ordinary record must retain the same value semantics across backends even if
its physical representation differs. Layout promises enter the type/ABI only
when the author explicitly asks for them.

## Data structures: batteries without compiler bloat

The distribution should feel complete on day one. That does not imply keywords
for every structure.

- **Semantic core:** arrays/slices, text/bytes, tuples/records/variants,
  persistent sequence/map/set interfaces, iterator/stream protocols.
- **Official toolkit:** mutable and persistent maps/sets/vectors, deque, heap,
  ordered collections, bit sets, tries, graphs, union-find, ring buffers, LRU,
  channels, spatial structures, and well-tested algorithms.
- **Specialized toolkit:** roaring bitmaps, Bloom/Cuckoo filters, probabilistic
  sketches, lock-free structures, and storage-aware indexes.

Probabilistic structures should make their uncertainty and budget visible:

```text
let seen = Bloom<String>.create(
  expected_items: 1_000_000,
  false_positive_rate: 0.001,
  memory_budget: 2.mebibytes,
)?
```

The constructor should reject incompatible promises rather than silently
selecting surprising parameters. The toolkit should expose complexity,
allocation, determinism, thread-safety, serialization, and profile support as
machine-readable metadata.

## Interaction war game

Before accepting a cluster, run it through representative programs and inspect
pairwise interactions:

| Interaction | Failure to prevent |
|---|---|
| Effects × async | Cancellation or suspension silently changes handler lifetime. |
| Effects × DI | Runtime service lookup erases the effect contract. |
| Context × concurrency | Tenant/principal leaks into detached work. |
| Context × distribution | Untrusted baggage becomes authorization input. |
| Pattern matching × evolution | Catch-all arms hide newly added protocol variants. |
| Contracts × effects | A precondition performs I/O or becomes nondeterministic. |
| Telemetry × privacy | Automatic fields export secrets or PII. |
| Telemetry × performance | Cardinality or event volume defeats latency/cost budgets. |
| Layout × FFI | Backend-specific padding silently changes ABI. |
| Macros × canonical source | Generated dialects cannot be explained or semantically diffed. |
| Hot reload × state | New code runs against incompatible actor/component state. |
| Defaults × security | A convenient development provider reaches production. |
| Nominal data × structural ports | Accidental shape compatibility grants domain meaning or authority. |
| Local inference × public API | A body edit silently changes callers or invalidates the whole graph. |
| Refinements × fast feedback | Solver/proof work blocks ordinary semantic checking. |
| Monolith × component data | One transaction becomes justification for global table access. |
| Local × remote extraction | A network boundary is presented as a transparent deployment choice. |
| UUIDv7 × privacy | Sortable identity leaks chronology or is mistaken for authorization. |
| JSON × validation | Parser coercion or duplicate fields bypass domain invariants. |
| Contract tests × confidence | Selected examples are mistaken for end-to-end correctness. |
| Actor × external effects | Restart repeats an effect that was not idempotent. |
| Structured tasks × blocking FFI | A foreign call stalls scheduler progress or ignores cancellation. |
| Retry × pools/backpressure | Recovery traffic amplifies overload and exhausts queues. |
| State machine × schema evolution | A new state cannot coexist with old persisted actors/messages. |
| GUI × host portability | Shared view syntax hides native accessibility/input differences. |

The minimum corpus should include a CLI, HTTP service, durable workflow,
actor-based service, parser, data pipeline, native/FFI loop, and restricted
embedded program. A feature that looks elegant in only one of these has not yet
earned general-purpose status.

## Syntax quality protocol

Syntax should be selected empirically, not by first-impression taste alone.

1. Express the same corpus with two or three coherent surfaces.
2. Canonically format each one.
3. Measure parse ambiguity, token count, edit locality, diff churn, and diagnostic
   precision.
4. Ask humans to audit seeded defects and AI agents to generate and repair them.
5. Inspect pathological cases, not only a landing-page example.
6. Keep the semantic AST stable while changing the surface.

The desired surface is calm because it has few context-sensitive rules and one
obvious shape—not because it compresses every idea into punctuation.

## Promotion rule

No feature becomes accepted because another language has it or because an AI
can generate it. Promotion requires an explicit decision record containing the
job, competing mechanisms, corpus evidence, costs by stakeholder lens,
interaction results, escape hatch, and migration story.
