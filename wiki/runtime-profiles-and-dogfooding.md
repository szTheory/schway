---
id: runtime-profiles-dogfooding
title: Runtime profiles, scheduling, memory isolation, and dogfooding
summary: A compiled-first execution posture that keeps low-level deployment small while admitting an optional BEAM-inspired application runtime and a compounding self-host path.
type: strategy
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-02
tags: [runtime, scheduling, memory, dogfood, bootstrap]
related: [vision, author-intent-model, concurrency-memory, memory-reclamation-policy, native-low-level-profile, resources-locks-caching, convergence-work-program, implementation-path, research-ledger, open-questions]
---

# Runtime profiles, scheduling, memory isolation, and dogfooding

## Reader and outcome

This note is for the bootstrap and runtime teams. After reading it, they should
be able to decide which execution services a program needs, preserve one source
meaning across profiles, and move useful tooling into Lang without making an
immature self-hosted compiler the only way to repair itself.

## Recommended posture

Lang is **compiled-first**, but “compiled” does not mean “runtime-free.” A
runtime is any support the generated program depends on: allocation, panic,
unwinding, task scheduling, async I/O, garbage collection, actors, supervision,
telemetry, or dynamic loading. Ahead-of-time code can use all, some, or none of
those services.

Adopt a monotonic profile ladder:

| Profile | Required execution support | Intended work |
|---|---|---|
| `freestanding` | compiler intrinsics plus explicitly supplied target hooks | kernels, firmware, boot code, alloc-free libraries |
| `native-core` | small versioned support library; explicit allocator/panic policy | CLIs, libraries, parsers, engines, embedded applications |
| `application` | structured task scheduler, async I/O, timers, cancellation, resource registry | desktop, network, data, and concurrent applications |
| `service` | application runtime plus actors, supervision, bounded queues, overload control, live inspection | servers, distributed components, durable agents |

Each higher profile adds services and legal effects. It cannot weaken bounds,
ownership, cleanup, overflow, or failure meaning. Libraries declare their
minimum profile and remain usable in larger profiles. A program may avoid the
scheduler or collector without forking the language.

`managed-region` is an application/service extension, not another rung:
isolated tracing regions or actor heaps may be enabled for cycle-heavy graphs
only if they win the measured tradeoff. No profile requires a global tracing
heap; the accepted boundary is detailed in [Memory reclamation and latency
policy](memory-reclamation-policy.md).

The reference interpreter is a development and conformance engine, not the
production definition of “runtime.” Native AOT compilation remains the first
shipping path. JIT, bytecode, BEAM, and Wasm remain later execution choices
under the same semantic contract.

Rust's platform-agnostic `no_std` core demonstrates that one language can make
runtime facilities profile-dependent rather than universal. ([Embedded Rust
`no_std`](https://doc.rust-lang.org/stable/embedded-book/intro/no-std.html))

## What belongs where

| Concern | Core semantics | Small native support | Optional application/service runtime | Official toolkit |
|---|---|---|---|---|
| ownership, borrows, sendability | yes | enforcement metadata/helpers | inspection only | collections/adapters |
| lexical resource cleanup | yes | panic/unwind/drop glue | cancellation coordination | files, sockets, transactions |
| allocation | capability and failure contract | selected allocator hook | scoped heaps/arenas/managed regions | policies and diagnostics |
| tasks and cancellation | lifetime/effect rules | none for sequential programs | scheduler, timers, async I/O | combinators |
| actors and supervision | protocols, isolation, exit meaning | none | mailboxes, scheduling, restart tree | patterns and adapters |
| tracing GC | no universal requirement | none | only for a selected managed region/profile | tuning and evidence |
| caching | purity/effects, stable keys, clocks, versions | hashing/serialization primitives | coalescing, timers, capacity accounting | typed cache policy |
| telemetry | typed semantic event contract | cheap hook/compile-out path | buffers, sampling, live queries | exporters and policy |

This keeps the language semantic core small without pretending that scheduling,
memory reclamation, or operations are ordinary third-party details.

## What to borrow from BEAM

Erlang supplies four unusually coherent mechanisms:

1. **Isolation as the unit of failure and reclamation.** Each process has its
   own stack and heap. Its collector is a per-process generational semi-space
   collector, so collection work is localized rather than a universal heap
   stop. ([Erlang garbage collector](https://www.erlang.org/doc/apps/erts/garbagecollection.html))
2. **Preemption as a responsiveness contract.** Processes are preempted after
   consuming a reduction budget rather than trusting user code to yield
   voluntarily. ([Erlang process scheduling](https://www.erlang.org/doc/apps/erts/erlang.html))
3. **Message-passing ownership boundaries.** Ordinary message data is copied
   between processes, with reference-counted binaries and literals as notable
   local exceptions. This helps isolation but makes communication cost real.
   ([Erlang process efficiency](https://www.erlang.org/doc/system/eff_guide_processes.html))
4. **Runtime visibility and supervision.** Processes, mailboxes, links, exits,
   reductions, memory, and schedulers are inspectable operational entities.

Borrow the contracts, not the implementation wholesale:

- tasks and actors have bounded work slices under application/service profiles;
- an actor owns a mutable region and processes one behavior at a time;
- unique values can move, immutable values can be shared under a known policy,
  and copies are measurable;
- allocation/reclamation is attributable to an owning task, actor, or region;
- blocking and long foreign work leave ordinary scheduler lanes;
- supervisors own restart and escalation; they do not make side effects atomic;
- mailboxes are bounded and typed, unlike an unbounded-default interpretation
  of classic actors.

### Costs and footguns the design must retain

- Per-process GC reduces pause scope; it does not guarantee flat total memory,
  bounded latency, or no large-object retention.
- Copying messages improves isolation but can waste bandwidth and allocation;
  sharing large binaries introduces reference-counting and retention costs.
- Selective receive can scan long queues, so closed protocols and bounded
  mailboxes remain important.
- A long native call can occupy an ordinary scheduler. Erlang's NIF guidance
  recommends short calls or separate dirty CPU/I/O schedulers; Lang should make
  that classification part of the FFI effect contract. ([Erlang NIF
  scheduling](https://www.erlang.org/docs/26/man/erl_nif.html))
- Process priorities can starve other work and do not automatically solve
  priority inversion. The happy path should prefer budgets and admission over
  user-selected priority tiers.

## Candidate native scheduling design

Begin with compiler-inserted safe points and bounded cooperative slices that
are **preemptive from application code's point of view**:

- function/backedge/allocation/effect safe points spend a task budget;
- a depleted budget yields to the scheduler;
- async I/O, waits, and system calls are explicit suspension points;
- `BlockingCpu` and `BlockingIo` foreign work use bounded dedicated executors;
- the compiler reports loops whose profile cannot guarantee a safe-point bound;
- realtime/freestanding profiles reject implicit safe points and require an
  explicit worst-case execution contract;
- later asynchronous preemption is an implementation experiment, not a source
  semantic promise.

Go's runtime distinguishes blocked, synchronous, and asynchronous safe points,
showing both the leverage and implementation complexity of arbitrary
preemption. ([Go preemption implementation](https://go.dev/src/runtime/preempt.go))

This candidate offers bounded responsiveness early without immediately taking
on signal-safe arbitrary-instruction preemption, precise stack maps everywhere,
or a universal tracing collector. Test CPU loops, allocation loops, FFI calls,
priority inversion, cancellation, and tail latency before committing it.

## Accepted memory boundary; experimental implementations

Use one ownership language with several storage implementations selected at
explicit regions, not per object:

```text
actor Catalog owns region catalog_memory {
  state: CatalogGraph,
  memory: Region.managed(limit: 256.mib, pause: 2.ms),
}

fn decode_frame(bytes: borrow Bytes, using arena: Arena)
  -> Result<borrow Frame, DecodeError>
{
  ...
}
```

- unique/borrowed values and lexical resources form the accepted universal foundation;
- arenas provide bulk local allocation and deterministic release;
- immutable shared values use a measured RC/reuse or interned representation;
- an actor/region may opt into local tracing only if the graph/cycle workload
  justifies it; tracing is never a silent fallback;
- mutable references cannot cross an actor/region boundary;
- cross-boundary data is moved, immutably shared, or visibly copied;
- profiles report message bytes, copying, retains, release-queue work, heap
  growth, and collection pause distributions.

Implicit value destruction cannot execute arbitrary user code, I/O, locks, or
blocking work. Pure memory release may be budgeted/deferred by the application
runtime, but observable resource close remains lexical and explicit when its
failure matters. This prevents “no GC” from merely moving an unbounded pause
into a reference-count cascade, recursive destructor, or arena reset.

Pony's ORCA research is important because it co-designs type-enforced isolation,
causal message delivery, actor ownership, and concurrent reclamation rather
than bolting an actor library onto a global shared heap. It demonstrates a
serious hybrid design family, not a ready-made choice for Lang. ([ORCA
paper](https://www.ponylang.io/media/papers/OGC.pdf)) Project Verona similarly
investigates isolated regions and the absence of concurrent mutation as a basis
for scalable memory management. ([Project Verona](https://www.microsoft.com/en-us/research/project/project-verona/))

## Runtime alternatives

| Alternative | What it optimizes | Why it is not the leading universal choice |
|---|---|---|
| no runtime beyond libc/target hooks | footprint, control, embedded reach | cannot supply supervised actors, async I/O, preemption, or live process evidence |
| one BEAM-like VM for every program | uniform concurrency and operations | conflicts with low-level layout, direct C ABI, small binaries, and freestanding control |
| one Go-like managed native runtime | simple deployment and strong service ergonomics | global runtime/collector becomes mandatory for libraries and restricted systems |
| each library brings its own executor/GC | local freedom | scheduler conflicts, duplicated pools, incompatible cancellation, poor inspection |
| explicit monotonic profiles | honest cost and broad reach | requires conformance discipline and careful runtime ABI evolution |

The profile ladder is the leading candidate because it matches the stated
preference for powerful defaults with explicit escape boundaries. “Runtime
minimalism” is not a goal by itself; the goal is the minimum runtime that earns
its cost for a selected program.

## Compounding dogfood ladder

Dogfood early, but self-host late enough that the language can carry the load.
Rust's compiler uses several bootstrap stages and documents that a fully current
stage is expensive; Go likewise needs an earlier Go compiler to build its
self-hosted toolchain. ([Rust bootstrapping](https://rustc-dev-guide.rust-lang.org/building/bootstrapping/what-bootstrapping-does.html),
[Go source installation](https://go.dev/doc/install/source))

| Level | Lang-written artifact | Dividend | Safety rail |
|---|---|---|---|
| D0 | executable semantic examples and corpus fixtures | tests every compiler change | host harness remains authoritative |
| D1 | pure standard modules and ownership/resource probes | exercises types, values, layout, and native lowering | interpreter/native differential checks |
| D2 | corpus/evidence runner | improves conformance, benchmarking, provenance, and AI feedback immediately | host runner can replay the same manifest |
| D3 | diagnostic renderer, cache/invalidation inspector, semantic-query CLI | improves daily compiler and agent work | versioned protocol; tools are replaceable clients |
| D4 | formatter and build/action-graph driver | dogfoods canonical source and hermetic caching | bootstrap binary and golden output retained |
| D5 | package resolver, language server, selected compiler passes | moves substantial toolchain value into Lang | mixed host/Lang build supported |
| D6 | compiler frontend and later backend/runtime portions | self-hosting and ecosystem credibility | reproducible two-stage build plus diverse double compilation |

### Recommended first Lang-written application

Promote a **Lang corpus and evidence runner** ahead of the generic JSON cleaning
CLI. The runner should:

- read versioned fixture/evidence manifests;
- stream and strictly validate structured data;
- invoke the reference interpreter and native artifacts;
- compare results, failures, cleanup, and event traces;
- hash complete semantic/toolchain inputs and explain cache hit/miss/invalidation;
- record wall/CPU/memory/disk/output-size samples with provenance;
- minimize and persist failing cases for humans and agents.

It contains the valuable JSON/data stress slice but directly accelerates Lang's
own feedback loop. It is deliberately a client of a stable compiler protocol,
not a compiler pass, so an early bug cannot make the compiler impossible to
repair. The generic data cleaner can remain a corpus workload inside it and
graduate later if outside users value it.

## Host implementation posture

Do not choose the host language by taste alone. Compare at least Rust, OCaml,
and Go for frontend iteration, persistent-query memory, parser/IR ergonomics,
Cranelift/QBE integration, packaging, build time, profiling, and dependency
surface. Rust is the leading integration candidate; its compile-time cost is a
material risk, not a footnote. Keep semantic fixtures, schemas, IR serialization,
and conformance tests host-neutral so changing the host does not redefine Lang.

Do not require the in-tree compiler to build itself during ordinary frontend
development. Use a trusted bootstrap snapshot, stage only the affected tools,
and reserve full double compilation for release/nightly evidence.

## Experiments and exit criteria

1. Compile one sequential native program with no scheduler or tracing collector.
2. Run CPU-bound and I/O-bound structured tasks under the candidate safe-point
   scheduler; measure p50/p99/p99.9 latency, fairness, overhead, and cancellation.
3. Compare unique move, immutable share, visible copy, and managed actor-region
   messaging across small records, trees, and large byte buffers.
4. Build graph/cycle-heavy actor, compiler, and GUI probes with no tracing GC
   and with a local managed region.
5. Block, spin, allocate, and re-enter through C to verify FFI classification
   and scheduler containment.
6. Build the D2 corpus/evidence runner in Lang and retain a host replay oracle.
7. Bootstrap two successive compiler stages and measure clean/incremental time,
   peak memory, artifact size, reproducibility, and semantic equivalence.

The runtime posture converges when the sequential profile does not pay for
unused scheduler/GC services, the service profile meets responsiveness and
isolation budgets, foreign work cannot silently stall it, and one Lang-written
tool measurably shortens the language's own verified development loop.
