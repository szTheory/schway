---
id: native-low-level-profile
title: Native and low-level programming profile
summary: The semantics, safety boundaries, compiler obligations, and workload proofs needed for credible systems, embedded, realtime, and FFI programming.
type: design
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-03
tags: [native, systems, memory, ffi]
related: [semantic-kernel-contract, semantic-kernel-probes, ownership-lifetime-decisive-study, concurrency-memory, memory-reclamation-policy, runtime-profiles-dogfooding, resources-locks-caching, type-system-failure-semantics, compiler-feedback-latency, compute-efficiency-constitution, convergence-work-program, capability-gauntlet, implementation-path, research-ledger]
---

# Native and low-level programming profile

## Conclusion

The first implementation is now explicitly **native-capable**: usable ownership,
lexical resources, C interop, and native execution are v1 requirements rather
than future compatibility promises. The language can preserve one
value/effect/module model across profiles, but credible native work requires
defined memory, layout, aliasing, allocation, concurrency, FFI, and panic
semantics from the first typed IR onward.

Ordinary application code should not pay Rust-like ownership syntax everywhere.
Native resource code cannot receive safety by hiding lifetime and aliasing
facts. The candidate compromise is:

- immutable, owned values and lexical resource scopes by default;
- compiler-inferred borrowing where local and unambiguous;
- explicit ownership/borrow/region notation at public, stored, concurrent, and
  unsafe boundaries;
- arenas and caller-provided allocators as ordinary capabilities;
- a small unsafe module boundary with machine-readable proof obligations;
- no release mode that silently changes checked operations into undefined
  behavior.

The [semantic kernel contract](semantic-kernel-contract.md) now fixes the v0
cross-feature laws, and the [probe suite](semantic-kernel-probes.md) makes the
native claim falsifiable across interpreter and optimized lowering.

## Semantics that must exist before calling it systems-capable

| Concern | Required design answer | Failure if deferred |
|---|---|---|
| Object lifetime | Ownership, moves, borrows, destruction order, escape rules | use-after-free, leaks, implicit copies |
| Aliasing and provenance | What pointers may derive from, alias, compare, and access | backend optimizations miscompile valid-looking code |
| Initialization and validity | Valid bit patterns, uninitialized storage, padding, partial initialization | undefined reads and FFI unsoundness |
| Layout and ABI | Size, alignment, field order, calling convention, enum/tag representation | corrupt foreign calls and persisted data |
| Integer operations | Overflow, narrowing, shifts, divide-by-zero, conversion | debug/release divergence and security defects |
| Concurrency memory model | Data races, atomics, ordering, fences, thread/interrupt sharing | hardware-dependent misbehavior |
| Volatile and MMIO | Reads/writes that must occur and address-space meaning | broken devices and optimizer reordering |
| Allocation and OOM | Allocator ownership, fallibility, fragmentation, bounded regions | hidden aborts and realtime failure |
| Stack and recursion | Stack growth, overflow behavior, recursion bounds | process corruption or missed deadlines |
| Panic and unwinding | Cleanup and ABI behavior across native/foreign boundaries | double failure, leaked resources, UB |
| Pinning/address sensitivity | Self-references, async state, intrusive structures, DMA buffers | moved address-sensitive values |
| Endianness and alignment | Explicit wire/device conversion and unaligned access | portability and protocol defects |
| Dynamic linking | Symbol/version/loading/unloading lifetime | stale pointers and incompatible plugins |
| Startup/runtime | Entry, libc/no-libc, TLS, signals, interrupts, runtime requirements | inability to build kernels or firmware |
| Intrinsics | SIMD, assembly, GPU and target-feature boundaries | opaque unsafe escape hatches |

## Candidate ownership surface

Most code stays simple:

```text
fn checksum(bytes: BytesView) -> U32 {
  bytes.fold(initial: 0, using: crc32_step)
}
```

Resource ownership becomes explicit when it crosses a meaningful boundary:

```text
fn read_frame(
  socket: borrow Socket,
  arena: borrow mut Arena,
) -> Result<borrow Frame, ReadError>
with { Network }
{
  let storage = arena.allocate_bytes(count: max_frame_size)?
  let count = socket.read(into: storage)?
  Frame.parse(bytes: storage.slice(to: count))
}
```

The returned frame is tied to the arena borrow. It cannot outlive the arena or
cross to another actor unless copied, immutably shared under the profile’s
rules, or uniquely transferred. Diagnostics should explain the resource story
in domain language, not expose inference internals first.

The exact syntax is unresolved. The invariant is not.

## Allocation and OOM

Like Zig libraries, native libraries should accept allocator/region authority
from callers instead of assuming a process-global heap. Allocation is an
effect when it can fail or consume a governed resource.

```text
fn decode_packet(bytes:, using arena: borrow mut Arena)
  -> Result<Packet, DecodeError | OutOfMemory>
```

Profiles may bind convenient defaults:

- service: managed allocation and process isolation;
- native application: tracing/region hybrid with a default application heap;
- embedded: caller-provided static arenas, no heap unless declared;
- realtime: bounded allocation phases and no allocation in marked critical
  sections.

OOM must not be an uninspectable universal abort. A profile may choose abort at
its outer policy boundary, but reusable code preserves fallibility.

## Checked semantics and optimization

Safety is a semantic promise, not a debug flag. Normal arithmetic, indexing,
conversions, and pointer operations have defined checked behavior in every
build profile. When raw hardware performance needs unchecked operations, the
operation is visibly distinct and locally justified:

```text
unsafe module simd_crc
  proves { bounds: caller, alignment: 32, target: avx2 }
{
  fn crc_blocks(bytes: AlignedView<32>) -> U32 {
    intrinsic.avx2_crc_unchecked(bytes:)
  }
}
```

Release modes select optimization, debug information, instrumentation, and
panic strategy. They do not quietly change source arithmetic or bounds
semantics. The optimizer may eliminate checks it proves redundant.

## Unsafe is a reviewed component, not a magic word

Lexical `unsafe` merely identifies where unchecked power is used; it does not
make that code sound. An unsafe component should declare:

- safety invariants and which caller/provider establishes each;
- memory, thread, interrupt, and reentrancy assumptions;
- valid layouts, alignments, provenance, and lifetimes;
- foreign ABI and unwind/panic contract;
- target features and portability limits;
- tests, fuzzers, sanitizers, model checks, and supported targets;
- transitive callers and artifacts that rely on its claims;
- reviewer/owner and last evidence version.

The semantic graph must make unsafe dependence visible transitively. Project
policy may forbid unsafe outside named components or dependencies.

## FFI and C

```text
foreign C library zlib {
  fn compress2(
    dest: out Buffer,
    source: borrow BytesView,
    level: CInt,
  ) -> CInt
  abi: platform_c
  unwind: forbidden
  blocking: false
}
```

Bindings need more than matching parameter shapes:

- ownership and lifetime of every pointer;
- nullability, length, alignment, mutability, and aliasing;
- callbacks and which thread/reentrancy rules apply;
- error-code translation and `errno` policy;
- whether calls block or may retain memory;
- ABI, target, symbol version, and unwind behavior;
- who initializes and shuts down foreign global state.

Foreign claims are quarantined: a generated binding can create a candidate
contract, never proof that the C implementation obeys it. Sanitizers, fuzzing,
header/ABI checks, and runtime guards become release evidence.

## Async, pinning, and blocking

Rust demonstrates that compiled async state may become address-sensitive and
require pinning. This language should avoid exposing pinning through ordinary
service code, but must specify it for native futures, intrusive collections,
DMA, and self-referential structures. The compiler may infer immobility inside
a lexical task; storing or exporting an address-sensitive value requires an
explicit stable-storage type.

Blocking foreign calls declare `blocking: true`; the runtime moves them to an
appropriate pool or rejects them in an incompatible context. Realtime and
interrupt contexts have statically restricted effect sets.

## Atomics, threads, actors, and interrupts

Actors and structured tasks remain the default concurrency interface. Native
shared memory is an explicit capability:

```text
fn increment(counter: borrow Atomic<U64>) -> U64
with { SharedMemory }
{
  counter.fetch_add(value: 1, ordering: Relaxed)
}
```

Atomic order is never an inferred decorative default. Higher-level atomic
types can encode common safe protocols. Data races are invalid programs, and
the toolchain should integrate race detection, schedule exploration, and weak
memory litmus tests. Interrupt handlers use a restricted profile: no blocking,
no ordinary allocation, explicit shared state, bounded work, and target-aware
calling convention.

## Data-oriented and performance tooling

The type system should permit representation choices without exposing them in
domain APIs:

```text
layout ParticleBatch as soa {
  position: Vector3<F32>,
  velocity: Vector3<F32>,
  lifetime: F32,
}
```

Dedicated syntax has not earned admission; an official layout derivation may
be enough. What must be first-class in tooling is explanation:

- size, alignment, padding, and cache-line footprint;
- allocation and copy sites;
- boxing, dynamic dispatch, and reference-count operations;
- vectorization success/failure and target features;
- actor/task scheduler hops and contention;
- effect/provider boundaries preventing optimization;
- profile-specific generated code and backend assumptions.

## Compiler/backend contract

LLVM optimizations distinguish immediate undefined behavior, poison values,
and undef-like values and rely on subtle pointer and attribute contracts. The
front end therefore needs a written portable core-IR semantics stricter than
“whatever LLVM currently accepts.” Lowering must preserve those semantics.

Evidence should include:

- differential execution across optimization levels and backends;
- randomized IR translation validation where practical;
- sanitizer and interpreter comparison;
- conformance tests for overflow, aliasing, layout, atomics, and panic;
- bounded compile-time evaluation with visible cost and cache keys;
- deterministic/reproducible artifact generation.

Compile-time code is still code. Like Zig’s evaluation quota, it needs resource
budgets and termination diagnostics. Avoid unrestricted procedural macros in
the compiler process; derivation should be typed, sandboxed, deterministic, and
cacheable.

## Lessons from Rust and Zig

| Lesson | Adopt | Avoid |
|---|---|---|
| Rust ownership | Safe abstraction boundaries, explicit unsafe obligations, strong concurrency guarantees | Requiring lifetime notation in ordinary service code before inference proves it necessary |
| Rust `Pin` | Make address sensitivity a real semantic property | Let async lowering surprise users with obscure pin/lifetime errors |
| Rust FFI | Explicit ABI and unwind contracts | Treating `extern` signatures as sufficient safety evidence |
| Rust compiler experience | Rich diagnostics and incremental queries | Unbounded trait/type/macro work and monomorphization cost hidden from authors |
| Zig allocators | Caller-controlled allocation and honest OOM | Allocator plumbing in every ordinary application API |
| Zig comptime | Powerful compile-time evaluation with quotas | Unbounded arbitrary build-time execution |
| Zig cross-compilation | First-class target descriptions and C integration | Host-environment leakage into supposedly hermetic builds |
| Zig build modes | Explicit performance/safety choices | A fast release setting that silently disables basic safety guarantees |

## Primary evidence

- [Rust undefined behavior reference](https://doc.rust-lang.org/stable/reference/behavior-considered-undefined.html)
  enumerates data-race, pointer, aliasing, validity, ABI, assembly, and runtime
  assumptions and notes that unsafe code must still avoid undefined behavior.
- [Rust type layout](https://doc.rust-lang.org/reference/type-layout.html)
  documents representations, size, alignment, and `repr(C)` constraints.
- [Rust `Pin`](https://doc.rust-lang.org/std/pin/)
  explains address-sensitive values and async/self-referential structures.
- [Rust Nomicon FFI](https://doc.rust-lang.org/nomicon/ffi.html) and
  [unwinding](https://doc.rust-lang.org/nomicon/unwinding.html) document ABI and
  cross-language unwind hazards.
- [Zig language reference](https://ziglang.org/documentation/master/) documents
  explicit allocators, fallible OOM, compile-time branch quotas, build modes,
  C translation, atomics, pointers, assembly, and unchecked behavior.
- [LLVM undefined behavior](https://llvm.org/docs/UndefinedBehavior.html) and
  [language reference](https://llvm.org/docs/LangRef.html) document backend
  semantics a frontend must lower into correctly.

All sources were last verified 2026-09-02. The candidate design is project
synthesis.

The ownership-model alternatives, 64-workload laboratory, edge-case laws, and
stakeholder gauntlet are specified in the [ownership and lifetime decisive
study](ownership-and-lifetime-decisive-study.md). Their place in the full
execution sequence remains in the [convergence work
program](convergence-work-program.md).
