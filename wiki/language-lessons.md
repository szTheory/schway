---
id: language-lessons
title: Language lessons and nearest existing options
summary: Mechanisms worth borrowing, costs worth avoiding, and practical languages closest to the vision.
type: research
status: draft
confidence: mixed
created: 2026-09-02
updated: 2026-09-02
tags: [languages, comparison, lessons, footguns]
related: [vision, design-atlas, compiler-feedback-latency, domain-data-distribution, research-ledger, implementation-path]
---

# Language lessons and nearest existing options

## Conclusion

No existing language combines the whole thesis. The nearest production path is
likely a **Gleam or Elixir service layer on the BEAM, with Rust or Zig at narrow
native boundaries**, while the nearest research inspirations are Koka's typed
effects, Roc's platform-supplied I/O, Unison's content-addressed code and
abilities, SPARK's executable contracts, Lean's small proof kernel, and Pony's
capability-safe actors.

That combination is a laboratory, not the final language. It lets the project
test whether effect-aware architecture, supervision, structured diagnostics,
and agent introspection actually improve outcomes before building a runtime.

## Landscape

The “borrow” and “avoid” columns are synthesis: they are design judgments based
on the documented mechanisms, not claims made by each language project.

| Language/system | Borrow | Avoid carrying forward | Relevance |
|---|---|---|---|
| C | Stable ABI, transparent layouts, small runtime, hardware access | Undefined behavior, unsafe defaults, manual ownership everywhere | Essential interop floor and embedded reference |
| Rust | Ownership, explicit results, traits, unsafe boundary, editions, integrated tooling | Borrow-checker concepts leaking into ordinary service code; slow or noisy compile loops; macro/dependency sprawl | Strongest safety/toolchain reference for native code |
| Zig | First-class C interop, explicit allocators, compile-time evaluation, built-in testing/cross-compilation | Unrestricted compile-time cleverness and manual lifetime burden in application code | Excellent boundary and bootstrap reference |
| Go | Fast builds, one tool, formatting, simple packages, explicit errors, goroutines, strong compatibility | Nil, data races, orphan goroutines, repetitive error plumbing, implicit interface satisfaction surprises | Best simplicity and turnkey-toolchain reference |
| Erlang/OTP | Isolated processes, mailboxes, links, supervision, hot operation, per-process GC, runtime inspection | Dynamic protocol mistakes and assuming all workloads suit copying/message passing | Best runtime and failure-model reference |
| Elixir | Canonical formatter, pipelines, macros used to create approachable APIs, Mix, ExUnit, OTP ergonomics | Runtime metaprogramming opacity, dynamic typing across large generated changes | Best overall developer-experience reference |
| Gleam | Small static functional language, ADTs, results, BEAM and JS targets, Erlang/Elixir interop | Relying on backend differences or foreign code that erases guarantees | Closest production language to the service profile |
| Pony | Actors plus reference capabilities enabling safe transfer/share without locks | Requiring every application engineer to reason fluently about a large capability lattice | Key concurrency and ownership experiment |
| Swift | Ergonomic value types, ARC, actor isolation, compile-time data-race checks, C-family interop | Reference cycles, implicit ARC cost, and platform-specific ecosystem gravity | Strong concurrency and API-design reference |
| OCaml | Fast native functional language, inference, ADTs, modules/functors, separate compilation | Two-level module complexity and untyped experimental effects as the final model | Compiler implementation and module-system reference |
| Haskell | Purity, algebraic modeling, property testing culture, abstractions with laws | Laziness as invisible cost, extension accretion, type-level puzzles, monad-transformer stacks | Semantic and testing reference; warning about maximalism |
| Koka | Inferred effect rows, handlers, totality distinctions, optimized reference counting research | Depending on research-stage ecosystem/tooling for initial production adoption | Closest reference for the semantic center |
| Roc | Pure/effectful function distinction, platform-supplied I/O, and domain-specific hosts | Requiring one platform abstraction to carry every independently composable capability; current ecosystem/runtime maturity | Strong host/effect boundary reference |
| Unison | Content-addressed definitions, abilities, structural operations, transcripts, code mobility | Making a novel code database and novel language mandatory in the first prototype | Most AI-relevant source identity and distribution reference |
| SPARK/Ada | Contracts with execution and proof semantics, strong types, analyzable safe subset | Applying proof ceremony uniformly or presenting “proved” as “meets unstated intent” | Highest-assurance reference and optional proof tier |
| Lean | Dependent types, proof-producing automation, small trusted kernel, machine-oriented integration | Arbitrary proof search or type-level computation in the default fast edit lane | Progressive assurance and checker architecture reference |
| Ruby | Humane naming, blocks, readable DSLs, convention over configuration, joy | Open classes, method-missing magic, ambient mutation, runtime-only surprises | Human-readability and framework ergonomics reference |
| Raku | Grammars, multiple dispatch, rich domain expression, playful language design | Context-sensitive maximal syntax and user-defined language variation in canonical source | Parsing/DSL inspiration and complexity warning |
| Lisp/Clojure | Uniform trees, structural editing, hygienic-macro potential, immutable data | Assuming homoiconicity ensures semantic simplicity; macro dialects that defeat common tooling | Candidate alternate surface for benchmarking |
| F#/ML family | Inference, discriminated unions, units of measure, pragmatic functional-first style | Hidden allocation or abstraction costs without good feedback | Domain modeling and approachable FP reference |
| Scala | Powerful types, FP/OOP interop, contextual abstraction | Implicit resolution and feature interaction becoming a language inside the language | Strong warning about “best of every paradigm” |
| TypeScript | Gradual adoption, excellent structural tooling, enormous JS reach | Unsound escape routes, structural types conflating shape and meaning, configuration layering | Adoption and language-server reference |
| Python | AI/data ecosystem, readable baseline, enormous model familiarity | Dynamic failures, packaging fragmentation, runtime reflection, global mutation | Ecosystem bridge; reminder that training prevalence matters |
| Lua | Tiny embeddable runtime and simple C API | Minimal semantics insufficient for architecture and correctness goals | Embedding and extension-language reference |
| WebAssembly components | Typed cross-language interfaces, capability-based hosting, composable binaries | Treating current toolchain coverage or performance as universal | Safer polyglot boundary and portable plugin substrate |
| LLVM/MLIR | Mature optimization/backends and layered intermediate representations | Coupling source semantics directly to one backend's transient details | Native bootstrap path, not necessarily permanent core |

## Evidence-backed lessons

### Failure and concurrency

OTP defines workers, supervisors, restart strategies, and applications as common
structural patterns rather than application-specific frameworks. Erlang also
ships production-suitable tracing and inspection of processes, messages,
schedulers, and allocations. This supports making supervision and introspection
runtime concepts. ([OTP](https://www.erlang.org/docs/27/system/design_principles.html),
[runtime tools](https://www.erlang.org/doc/apps/runtime_tools/api-reference.html))

Swift detects many data races through actor isolation, while Pony's reference
capabilities constrain which mutable or immutable references may cross actors.
The ergonomic lesson is to provide strong isolation defaults and expose advanced
ownership only when data actually crosses a risky boundary. ([Swift concurrency](https://docs.swift.org/swift-book/documentation/the-swift-programming-language/concurrency/),
[Pony capabilities](https://tutorial.ponylang.io/reference-capabilities/reference-capabilities.html))

### Effects and dependency injection

Koka includes the effect row in every function type and can infer polymorphic
effects. Unison abilities pair effect interfaces with replaceable handlers.
OCaml demonstrates that handlers can also implement generators, coroutines, and
user-level threads, while explicitly noting that its current effects are not
statically effect-safe. ([Koka](https://koka-lang.github.io/koka/doc/book.html),
[Unison abilities](https://www.unison-lang.org/docs/fundamentals/abilities/),
[OCaml effects](https://ocaml.org/manual/5.2/effects.html))

Roc currently distinguishes pure and effectful functions and delegates all I/O
to an application platform. That suggests a capability-secure host boundary and
domain-specific runtime profiles. This design should preserve that explicitness
while allowing multiple small capabilities/providers rather than one indivisible
platform.
([Roc functions and effects](https://roc-lang.org/functional),
[Roc platforms](https://www.roc-lang.org/docs/main/langref/platforms/))

### Verification and testing

SPARK puts preconditions, postconditions, invariants, and data dependencies in
the language; the same contracts may execute or become proof obligations. Go's
toolchain integrates coverage-guided fuzzing and saves minimized failures as
regression inputs. Rust compiles unit, integration, example, and documentation
tests under one command. ([SPARK contracts](https://docs.adacore.com/spark2014-docs/html/ug/en/source/subprogram_contracts.html),
[Go fuzzing](https://go.dev/doc/security/fuzz/),
[Cargo tests](https://doc.rust-lang.org/cargo/commands/cargo-test.html))

Lean shows a progressive-assurance architecture in which automation produces
proof terms checked by a smaller kernel. Its dependent-type material also makes
the compile-time cost of type-level computation explicit. Borrow the small
checker and optional evidence ladder, not a requirement that every service
operation carry an interactive proof.
([Lean reference](https://lean-lang.org/doc/reference/latest/),
[Lean dependent types](https://lean-lang.org/functional_programming_in_lean/Programming-with-Dependent-Types/))

### Deterministic distributed testing

FoundationDB runs cluster software in deterministic simulation with simulated
time and injected network, machine, and datacenter failures, retaining seeds for
reproduction. Typed replaceable effects give this language a credible path to a
similar test world for services, provided foreign and truly concurrent sources
of nondeterminism are acknowledged.
([FoundationDB testing](https://apple.github.io/foundationdb/testing.html))

### Memory is an explicit trade space

Rust ownership provides memory safety without a collector. Erlang uses
per-process heaps and generational collectors, supporting independent short
collections but retaining a runtime and message representation. Go documents
the direct CPU-versus-memory tradeoff of tracing GC and soft container-aware
limits. Swift documents ARC's strong-reference-cycle failure mode. No one model
dominates every target profile. ([Rust ownership](https://doc.rust-lang.org/stable/book/ch04-00-understanding-ownership.html),
[Erlang GC](https://www.erlang.org/doc/apps/erts/garbagecollection.html),
[Go GC](https://go.dev/doc/gc-guide),
[Swift ARC](https://docs.swift.org/swift-book/documentation/the-swift-programming-language/automaticreferencecounting/))

### Interop should graduate beyond C

Zig shows the practical value of importing C headers and exposing ABI-compatible
types. The WebAssembly Component Model adds rich interface types and a canonical
ABI so components in languages with different memory models can compose without
sharing raw memory. Use C for reach; prefer typed components when isolation and
portability matter. ([Zig language reference](https://ziglang.org/documentation/0.15.1/),
[Component Model rationale](https://component-model.bytecodealliance.org/design/why-component-model.html))

### Evolution and builds are language design

Rust editions permit opt-in surface changes while keeping cross-edition crates
interoperable. Go's compatibility promise covers the language and core APIs.
Bazel's content-addressed remote caching depends on declared inputs and
reproducible actions; its documentation also records cache-poisoning and host-
tool leakage risks. ([Rust editions](https://doc.rust-lang.org/edition-guide/editions/),
[Go compatibility](https://go.dev/doc/go1compat),
[Bazel remote caching](https://bazel.build/remote/caching))

## Closest practical choices today

### 1. Gleam on BEAM

Best fit for statically typed functional service code, ADTs, explicit results,
BEAM concurrency, and access to Erlang/Elixir libraries. It does not provide the
proposed typed effect/capability architecture, native systems profile, or rich C
story by itself. ([Gleam](https://gleam.run/))

### 2. Elixir on BEAM

Best fit for the desired turnkey experience, formatter, tests, supervision,
runtime inspection, hot development, and production services. Types and effects
would remain conventions or external analysis rather than hard guarantees.

### 3. Rust with an actor/runtime framework

Best fit for native performance, C interoperability, memory safety, typed
errors, and mature tooling. It carries more ownership and compile-time
complexity than the calm application surface wants, and supervision is not a
universal runtime invariant.

### 4. Koka, Roc, or Unison as a research base

Best fit for exploring effects, handlers, platform-owned I/O, content identity,
and distribution. Koka's own documentation currently describes
its v3 ecosystem as lacking async libraries and package management; Roc's
platform model and evolving toolchain need direct experimentation; Unison's
codebase model changes adoption and tooling assumptions substantially. They are
inspiration and experiment hosts, not the conservative production default.
([Koka status](https://koka-lang.github.io/koka/doc/index.html),
[Roc platforms](https://www.roc-lang.org/docs/main/langref/platforms/),
[Unison big idea](https://www.unison-lang.org/docs/the-big-idea/))

### 5. Zig at the systems boundary

Best fit for compiling and integrating C, controlling allocation, and producing
portable native artifacts. It is not a managed distributed-service runtime.

## Lessons about maximalism

“Best of all worlds” should not mean importing every feature. Raku, Scala,
Haskell extensions, C++, and macro-heavy ecosystems show a recurring systems
effect: individually useful powers interact until there is no single obvious
way to read or reason about a program.

The alternative is **semantic maximalism over a minimal orthogonal core**:

- ADTs model state.
- Functions transform values.
- Effects describe external behavior.
- Capabilities grant authority.
- Handlers interpret effects.
- Components constrain dependency boundaries.
- Tasks and actors describe lifetime and isolation.
- Contracts describe obligations.

Every proposed feature should first try to lower into those concepts.
