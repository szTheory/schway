---
id: concurrency-memory
title: Concurrency and memory model
summary: Structured tasks, supervised actors, bounded flow, and profile-specific memory implementations under one explicit source model.
type: design
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-02
tags: [concurrency, actors, memory, runtime, performance]
related: [vision, effects-and-capabilities, compiler-feedback-latency, runtime-profiles-dogfooding, memory-reclamation-policy, domain-data-distribution, type-system-failure-semantics, research-ledger]
---

# Concurrency and memory model

## Recommendation

Provide three safe concurrency lanes with visibly different lifetimes:

1. **Structured tasks** for finite concurrent work.
2. **Supervised actors** for long-lived isolated mutable state and services.
3. **Pure data parallelism and bounded streams** for bulk work and flow.

Do not expose ordinary shared mutable memory in the service profile. A future
native profile may provide a lexically marked `Shared<T>` escape with atomics,
locks, race detection, and model-checking hooks. “Actor,” “task,” “stream,” and
“thread” are not synonyms and must not share an ambiguous spawn API.

## Structured tasks

Every task belongs to a lexical group or runtime supervisor. Leaving a group
waits for, cancels, or explicitly detaches children according to its declared
policy. Deadline, cancellation, trace context, tenant-safe execution metadata,
and resource budget propagate structurally.

```text
within Task.group(deadline: 800.ms, failure: cancel_siblings) {
  let profile = Task.start { Accounts.profile(account:) }
  let balance = Task.start { Billing.balance(account:) }
  let activity = Task.start { Analytics.recent(account:, limit: 20) }

  Dashboard(
    profile: profile.await()?,
    balance: balance.await()?,
    activity: activity.await()?,
  )
}
```

The type/effect checker rejects a task handle escaping its owning group unless
ownership moves to an explicit supervisor. Blocking calls must declare
`Blocking`; the runtime moves them to an appropriate pool or rejects them for
that profile. Cancellation cleanup is scoped and bounded rather than an
arbitrary exception caught anywhere.

Swift task groups are useful precedent for scoped child tasks. The key project
requirement is stronger: all default child creation follows a structured
lifetime, and the semantic graph exposes it to agents.
([Swift TaskGroup](https://developer.apple.com/documentation/swift/taskgroup))

## Supervised actors

Actors own mutable state and process closed message protocols sequentially.
They have bounded mailboxes, explicit overflow behavior, startup/shutdown
contracts, stable state schemas, and supervisor policy.

```text
actor InventoryCache receives CacheMessage {
  state: CacheState,
  mailbox: bounded(capacity: 10_000, overflow: Reject),

  on Get(sku:) -> Reply<Option<Stock>> {
    reply state.stock.get(key: sku)
  }

  on Refresh(snapshot:) {
    state = CacheState.from_snapshot(snapshot:)
  }
}

supervisor CheckoutWorkers {
  strategy: one_for_one,
  restart: max 3 within 10.seconds,
  child inventory_cache: InventoryCache(initial: CacheState.empty()),
}
```

Values crossing an actor boundary must be immutable/shareable, copied, or
uniquely transferred. Actor failure discards the current message mutation and
becomes a typed exit observed by the supervisor. External side effects are not
rolled back automatically; idempotency and workflow policy remain explicit.

BEAM/OTP's worker/supervisor model is the leading operational reference and a
later conformance target because it treats supervision and process isolation as
runtime primitives. Native execution remains the first implementation target.
The source model should not inherit every BEAM semantic accidentally; mailbox
bounds, protocol types, deadlines, and overload behavior are stronger defaults
this project must add.

## Streams, backpressure, and pools

Every asynchronous stream declares demand/buffer semantics. Every resource pool
declares capacity, queue capacity, acquisition deadline, fairness policy, and
failure behavior. Unbounded is an explicit reviewed exception, never a default.

```text
Orders.events()
  |> Stream.map_parallel(
    concurrency: 16,
    ordered: false,
    buffer: 64,
    fn(event:) { Projections.apply(event:) },
  )
  |> Stream.run(deadline: 30.seconds)
```

Backpressure propagates demand rather than merely moving an unbounded queue
upstream. Reactive Streams provides a useful minimal precedent for asynchronous
non-blocking demand and terminal signals.
([Reactive Streams](https://www.reactive-streams.org/))

Connection pools are typed resources, not globals:

```text
provide Database = Postgres.pool(
  size: Runtime.cpu_count * 2,
  queue: bounded(capacity: 128),
  acquire_deadline: 100.ms,
  idle_timeout: 30.seconds,
) lifetime service
```

The default may be computed from a container budget, but `runtime explain
Database` shows the resolved values, provenance, saturation, and rejected
waiters. Hidden adaptive policy is unacceptable if it cannot be inspected.

## Retry and overload are one system

Retries consume the same deadline, concurrency, and rate budgets as original
work. A retry policy must name one owning layer and prove or explicitly waive
idempotency. Exponential backoff with jitter belongs in a typed runtime toolkit,
not as invisible client behavior.

```text
Remote.call(
  request:,
  deadline: 2.seconds,
  retry: Retry.exponential(
    on: [Unavailable, Reset],
    max_attempts: 3,
    max_elapsed: 1.5.seconds,
    jitter: full,
    idempotency_key: request.id,
  ),
)
```

The compiler/runtime rejects nested retry policies without an explicit shared
budget. Supervisors do not restart a permanently failing child forever; restart
intensity and escalation are mandatory.

## Shared-memory native escape hatch

Safe service code cannot take locks because it has no shared mutable references.
A native-only boundary can opt in:

```text
native unsafe shared Counter {
  value: Atomic<U64>,
}

fn increment(counter: Shared<Counter>) -> U64 {
  counter.value.fetch_add(value: 1, ordering: Relaxed)
}
```

The `unsafe shared` region is searchable and carries proof/test obligations.
Locks cannot be held across `await`; lock order may be declared and checked;
atomics require an ordering; model-check and sanitizer configurations belong in
the standard verifier. Go's memory model gives data-race-free programs a strong
sequentially consistent interpretation but still permits programs with races,
and its race detector is dynamic. This project should make the safe subset the
ordinary language and isolate the lower-level escape.
([Go memory model](https://go.dev/ref/mem),
[Go race detector](https://go.dev/doc/articles/race_detector))

## Scheduling and isolation lessons from BEAM

The attractive property is not “GC” in isolation. BEAM combines small isolated
process heaps, message passing, preemptive reduction budgets, supervision, and
runtime inspection. Erlang's collector is per-process generational copying GC,
and each process owns its stack and heap. ([Erlang garbage
collector](https://www.erlang.org/doc/apps/erts/garbagecollection.html)) The
runtime preempts a process after it consumes a reduction budget. ([Erlang
scheduling](https://www.erlang.org/doc/apps/erts/erlang.html))

The costs are equally instructive: most messages are copied, large binaries are
shared with reference counting, selective receive can scan a mailbox, and long
native calls can stall normal scheduler work unless split or isolated on a
dirty executor. ([Erlang process efficiency](https://www.erlang.org/doc/system/eff_guide_processes.html),
[Erlang NIF scheduling](https://www.erlang.org/docs/26/man/erl_nif.html))

The native candidate therefore uses compiler safe points and bounded work
budgets for application/service tasks, isolates blocking CPU and I/O work,
attributes memory to an owner, and leaves tracing collection to explicit
managed actor/region experiments. The full profile and dogfood rationale is in
[Runtime profiles and dogfooding](runtime-profiles-and-dogfooding.md).

## Memory strategy by profile

The architectural posture is now accepted: no profile requires one global
tracing heap. Ordinary code uses owned immutable values, local borrows,
compiler-known release, lexical resources, and explicit arenas. Immutable
cross-owner sharing and cyclic managed graphs are visibly different storage
forms.

| Profile | Default | Optional machinery |
|---|---|---|
| `freestanding` | fixed/stack/caller-provided storage; no hidden collector | explicitly bounded arenas or target hooks |
| `native-core` | unique owned storage and arenas; deterministic structural release | explicit immutable sharing linked on demand |
| `application` | same semantics with scheduler/accounting support | bounded deferred pure-memory release; isolated managed region |
| `service` | same semantics plus per-owner memory budgets | actor-local managed graphs when enabled |
| Wasm component target | Lang ownership projected into host linear memory/resources | copy/borrow boundary evidence |
| BEAM target | Lang ownership projected onto process heaps/messages | backend collector and binary-retention evidence |

Do not let each object arbitrarily choose GC versus ARC versus manual lifetime.
The remaining experiment compares precise RC/reuse and one isolated managed
region; it no longer decides whether every program must carry a collector.
`memory explain` reports allocation, retention, copies, sharing, release work,
and per-region pause distributions.

The complete decision, including destructor restrictions, realtime rules, and
the release-latency war game, is in [Memory reclamation and latency
policy](memory-reclamation-policy.md).

## Concurrency footgun register

The verifier, runtime, or profiler must have an answer for each of these:

| Failure | Candidate prevention/evidence |
|---|---|
| Data race | Isolation/sendability; `Shared` escape only; race/sanitizer runs |
| Deadlock | No locks in service subset; lock-order/await checks in native escape |
| Livelock/starvation | Scheduler/fairness telemetry and bounded retry/yield policy |
| Priority inversion | Explicit scheduler class and lock diagnostics where applicable |
| Lost wakeup/ABA | No raw condition variables in safe subset; typed atomic primitives |
| False sharing/cache contention | Layout/profiling evidence in native profile |
| Unbounded queue/mailbox | Bounded default with explicit overflow result |
| Head-of-line blocking | Per-operation latency/queue evidence and partitioned workers |
| Orphan work | Structured task ownership and supervisor transfer only |
| Cancellation leak | Scoped cleanup, cancellation tests, deadline propagation |
| Blocking scheduler thread | `Blocking` effect and dedicated executor |
| Actor reentrancy | Non-reentrant default; explicit protocol when enabled |
| Retry/thundering herd | One owner, shared budget, jitter, admission control |
| Version-skewed message | Versioned protocol and mixed-version verification |
| External duplicate effect | Idempotency key and delivery contract |
| Panic after external commit | Durable workflow/compensation obligation diagnostic |

## Experiments required

1. Map structured tasks and typed cancellation to BEAM processes without making
   cheap finite tasks unnaturally heavy.
2. Benchmark bounded mailboxes and overflow defaults under burst and overload.
3. Inject every footgun in the register and measure prevention, diagnosis, and
   false positives.
4. Compare precise RC/reuse and one isolated managed region against the accepted
   unique/arena foundation on trees, cycles, actors, and adversarial release
   chains; use BEAM as an operational comparison rather than a source mandate.
5. Build a small emulator or renderer to test unique buffers, data layout,
   deterministic scheduling, and shared-memory escape ergonomics.
6. Measure compiler-inserted safe-point overhead and fairness against
   asynchronous preemption on CPU loops, allocation loops, and foreign calls.
