---
id: memory-reclamation-policy
title: Memory reclamation and latency policy
summary: The accepted v1 memory posture: ownership-first semantics, no mandatory global tracing heap, explicit managed islands, and measured cleanup latency.
type: design
status: accepted
confidence: high
created: 2026-09-02
updated: 2026-09-03
tags: [memory, ownership, latency, runtime, decision]
related: [vision, ownership-lifetime-decisive-study, native-low-level-profile, concurrency-memory, runtime-profiles-dogfooding, resources-locks-caching, performance-observability-delivery, convergence-work-program, research-ledger, open-questions]
---

# Memory reclamation and latency policy

## Reader and outcome

This decision is for compiler, runtime, and library implementers. After reading
it, they should be able to add storage or concurrency machinery without
silently imposing a global collector, unpredictable cleanup, or manual memory
management on every Lang program.

## Accepted v1 posture

Lang will not require a global tracing garbage-collected heap. Its universal
source model is ownership-first:

1. Values are immutable and owned by default. The compiler chooses stack,
   inline, unique-heap, or reuse representations without changing value
   meaning.
2. Local borrows remove copies without creating owning aliases. Public,
   stored, concurrent, and foreign boundaries expose ownership explicitly.
3. Resources such as files, sockets, locks, mappings, and foreign handles have
   lexical ownership and a deterministic close attempt. They never depend on
   tracing finalizers.
4. Arenas and regions provide explicit bulk lifetimes for parsers, frames,
   requests, compiler passes, and other phase-shaped work.
5. Immutable cross-owner sharing is explicit. A `Shared<T>`-like form may use
   precise reference counting; cycles are not admitted accidentally.
6. Application and service programs may opt an isolated actor or region into
   managed tracing when a cyclic graph genuinely needs it. Managed references
   cannot escape that owner except through a checked snapshot, immutable
   share, or region-bound handle; a deep move is legal only when ownership
   proves the moved graph is independent.

This resolves the architectural question. It does **not** prematurely choose
the implementation of immutable sharing or the collector inside a managed
region. Those remain measured experiments.

```text
ordinary value       compiler-owned representation; no global GC assumption
unique buffer        move/borrow; deterministic release
arena value          cannot outlive the arena
Shared<T>             explicit immutable sharing; cycle policy checked
Managed<R, T>         graph value confined to managed region R
Resource<T>           exactly one close attempt; failure handled explicitly
```

## The actual contract is latency, not an algorithm name

“No GC” does not imply smooth execution. Immediate reference counting can
release a long chain on the unlucky operation that removes the last reference.
An arena can defer the same work until reset. Deterministic destructors can run
arbitrary or blocking cleanup. An allocator can fault pages or contend. A
concurrent tracing collector trades pauses for CPU, barriers, memory headroom,
and mutator assistance.

Go's collector guide makes the fundamental CPU-versus-memory tradeoff explicit
and documents latency from brief pauses, GC CPU competition, and allocation
assistance. OpenJDK's ZGC shows that a sophisticated concurrent collector can
keep pauses very small, but only with substantial runtime machinery, barriers,
and a throughput/memory tradeoff. ([Go GC guide](https://go.dev/doc/gc-guide),
[OpenJDK ZGC](https://wiki.openjdk.org/display/zgc/Main))

Therefore every memory implementation must satisfy an observable budget:

- report allocation rate, live/retained bytes, peak RSS, reclamation CPU,
  memory headroom, and p50/p99/p99.9/max stall distributions;
- attribute allocation, sharing, retention, and reclamation to a component,
  task, actor, region, or foreign boundary;
- expose why a value escaped, copied, became shared, or entered a managed
  region;
- distinguish hard bounds from measured objectives;
- fail explicitly on exhausted memory rather than enter unbounded reclamation
  thrash or silently switch memory strategies.

The service profile may offer goals such as `latency`, `balanced`, and
`footprint`, but these are budget policies rather than a public menu of
collector algorithms. V1 should ship one implementation per admitted storage
form. Supporting several interchangeable collectors before one is proven would
multiply stack-map, barrier, FFI, debugging, tuning, and support costs.

## Cleanup must be boring

Implicit memory reclamation may only perform compiler-known, non-failing
structural work. Lang should not permit arbitrary user code, I/O, locks, or
blocking in an implicit value destructor.

External cleanup is a resource operation with two levels:

- scope exit guarantees that cleanup is attempted on every ordinary exit;
- operations whose failure matters, such as commit, flush, durable sync, or
  protocol shutdown, are explicit and return `Result` before fallback cleanup.

This preserves deterministic acquisition/release without hiding slow network
operations or secondary failures inside `drop`. Rust specifies recursive
lexical destruction and also illustrates why destruction order and cleanup
interactions become language semantics rather than an allocator detail.
([Rust destructors](https://doc.rust-lang.org/reference/destructors.html))

Pure memory release may be scheduled differently by profile:

| Profile | Reclamation rule |
|---|---|
| `freestanding` | no hidden allocation or collector; all storage and release bounds are explicit |
| `native-core` | compiler-known unique/arena release; sharing is explicit and linked on demand |
| `application` | same semantics; bounded per-owner deferred release may smooth non-observable memory work |
| `service` | same semantics plus isolated managed regions and scheduler/accounting integration when enabled |

Deferring release cannot delay an observable resource close, cross a memory
limit without policy, or make destructor side effects observable—because
implicit destructors have no such effects.

## Realtime and latency-critical work

Hard realtime is not inferred from using no GC. A restricted `critical` or
realtime region must be mechanically checkable:

- no unbounded allocation, reclamation, lock acquisition, blocking, tracing,
  page fault assumption, or foreign call;
- storage is preallocated, stack-based, or drawn from a bounded arena/pool;
- loops and cleanup have declared or derived work bounds;
- calls require an effect/cost set legal for the region;
- failure policy is explicit before entry.

Application/service latency goals remain measured SLOs, not hard realtime
claims. They use scheduler work budgets, bounded deferred reclamation, and
load/fault tests to control tails.

## Managed-region admission rule

A tracing region is admitted only when all of the following are true:

1. The workload contains useful cyclic or identity-rich graphs that are
   materially awkward under unique ownership, arenas, weak edges, or explicit
   graph indexes.
2. The region has one owner and no concurrently mutable references escape it.
3. Its memory, pause, CPU, promotion/retention, and cross-boundary-copy costs
   are inspectable.
4. Cancellation, panic, FFI pinning, snapshots, and actor restart have defined
   behavior.
5. A graph corpus demonstrates better whole-system ergonomics and performance
   than the ownership/RC alternatives.

BEAM's per-process heaps show why isolation can localize collection work, but
message copies, large shared binaries, mailbox behavior, and total memory still
need measurement. ORCA and Verona show credible co-designed actor/region
families, not a ready-made Lang collector. ([Erlang garbage
collector](https://www.erlang.org/doc/apps/erts/garbagecollection.html),
[ORCA](https://www.ponylang.io/media/papers/OGC.pdf),
[Project Verona](https://www.microsoft.com/en-us/research/project/project-verona/))

## War game

| Failure mode | Shift-left response | Runtime evidence |
|---|---|---|
| reference-count cascade | compiler estimates release depth; bounded deferred release outside critical regions | release-queue depth and stall histogram |
| reference cycle | ordinary shared graph rejects or requires weak/managed edge | cycle witness and owning region |
| arena teardown spike | warn on high-cost reset; permit incremental reset outside realtime | reset work and retained pages |
| atomic RC contention | prefer move/local share; mark cross-thread `Shared`; profile cache-line contention | atomic retain/release and contention counts |
| managed-region pause | isolated owner, pause/work budget, no global roots | per-region pause distribution |
| collector thrash near limit | admission/headroom policy and fail-fast boundary | GC CPU, allocation debt, limit breaches |
| resource hidden in finalizer | reject finalizer-only resource type | leaked-resource witness |
| cleanup blocks cancellation | explicit async close with deadline; fallback policy | cleanup duration and cancellation delay |
| FFI hides pointers or blocks | typed pin/root/ownership plus blocking class | foreign residency and scheduler stall |
| OOM changes semantics | one explicit profile policy; no silent fallback collector | allocation failure path and evidence |

## Decisive experiments

The ownership laboratory should compare unique values, arenas, precise
RC/reuse, explicit shared values, and one isolated managed region. The complete
64-workload protocol is in the [ownership and lifetime decisive
study](ownership-and-lifetime-decisive-study.md); its memory subset includes:

- persistent trees and DAGs;
- a cyclic compiler or GUI graph;
- bursty actors exchanging small records and large buffers;
- parser/request arenas;
- last-reference release of adversarial chains;
- cross-thread sharing and FFI pinning;
- memory pressure, cancellation, panic, and OOM.

Perceus establishes precise RC/reuse as a serious candidate for a functional,
cycle-free core; it does not settle cycles, cross-thread accounting, or this
project's latency tails. ([Perceus](https://doi.org/10.1145/3453483.3454032))

The first implementation does not need to ship a managed region to satisfy the
native-capable v1 requirement. The result chooses representations and later
runtime code; it does not reopen the accepted decisions that the global heap is
optional, resources are lexical, and managed graphs are confined.
