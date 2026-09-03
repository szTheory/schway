---
id: resources-locks-caching
title: Resources, locks, synchronization, and caching
summary: Safe lexical resource lifetimes, explicit sync/async behavior, deadlock defenses, and typed cache policy with deterministic tests and runtime evidence.
type: design
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-02
tags: [resources, concurrency, caching, safety]
related: [native-low-level-profile, concurrency-memory, runtime-profiles-dogfooding, effects-and-capabilities, performance-observability-delivery, research-ledger]
---

# Resources, locks, synchronization, and caching

## Reader and outcome

This note is for someone designing or using native resources, concurrency, or
caches. After reading it, they should be able to make acquisition, lifetime,
cleanup, blocking, synchronization, staleness, and capacity explicit without
turning every call into manual plumbing.

## Recommendation

Use one resource model across files, sockets, memory, locks, transactions,
tasks, subscriptions, foreign handles, and provider lifetimes:

- acquisition returns an owned linear resource or an expected error;
- lexical scope guarantees exactly-once cleanup on success, error, panic, and
  cancellation;
- ownership prevents use after close and accidental concurrent use;
- async cleanup is awaited before the enclosing structured scope completes;
- cleanup failure is represented rather than overwritten by another failure;
- critical resources never depend on nondeterministic GC finalization;
- providers bind application/request/task scopes and close in dependency order.

Caches reuse the same ideas—explicit ownership, lifetime, clocks, budgets,
providers, events, and tests—but add freshness and invalidation semantics.

The strongest cache posture is: **make caching removable within its declared
consistency contract**. If removing or losing the cache destroys authoritative
truth, it is a storage system and must declare durability, recovery, and
consistency instead of using cache terminology.

## Lexical resources

```text
fn import_orders(path: Path)
  -> Result<ImportSummary, ImportError>
  with { Files, Database }
{
  use file = Files.open(path:, mode: Read)?
  use transaction = Database.begin(isolation: Serializable)?

  let summary = decode_orders(from: file)
    |> insert_orders(using: transaction)?

  transaction.commit()?
  Ok(summary)
}
```

`use` binds an owned resource and closes it at lexical scope exit in reverse
acquisition order. Moving it transfers cleanup responsibility. Borrowing it
cannot outlive the owner. Returning it transfers ownership explicitly.

`defer` remains useful for a small ordinary action, but `use` carries stronger
resource semantics and tool visibility. A generic `defer close()` pattern can
forget ownership, double-close, perform fallible cleanup invisibly, or close a
resource while child work still uses it.

## Fallible cleanup and transactions

Destructors should release process-local memory/handles and must not become an
arbitrary failure channel. Operations whose completion matters—flush, commit,
sync, publish, graceful shutdown—are explicit and return `Result`.

If both the body and cleanup fail, preserve both:

```text
ResourceFailure(
  primary: Some(ParseFailed(...)),
  cleanup: [FileCloseFailed(...)],
)
```

Do not replace the original cause with the last close error. During panic,
cleanup failures join the causal failure packet; they do not trigger recursive
unwinding. Profiles may use abort semantics, but resource ownership remains
defined.

A committed transaction records that rollback is no longer required. Rollback
is best-effort cleanup; commit failure remains an explicit application result.

## Structured lifetime and cancellation

```text
use server = Tcp.listen(address: config.address)?

scope connections {
  for await socket in server.accept() {
    spawn supervised handle_connection(own socket)
  }
}
```

The scope does not close the listener or return while children still own
accepted sockets. Cancellation propagates from the owner to children, waits for
bounded cleanup, then escalates according to policy. A child cannot cancel its
parent unless given that authority.

Cleanup receives a narrow cancellation shield and deadline so it can release
resources without becoming immortal. Runtime inspection shows resources still
open, their owner, acquisition site, age, child users, and cleanup state.

## Sync, async, and blocking

Suspension and blocking are distinct effects:

- `await` marks a suspension point where other work may run;
- `Blocking` marks an operation that occupies an OS/runtime worker;
- nonblocking async I/O carries cancellation and deadline semantics;
- foreign calls declare blocking and callback/reentrancy behavior;
- realtime/interrupt contexts reject both ordinary blocking and unbounded
  suspension;
- holding a non-async lock across `await` is rejected by default.

Do not transparently turn synchronous functions into asynchronous ones or vice
versa. Adapters can run blocking work in a bounded blocking pool, but queueing,
deadline, cancellation, and saturation remain visible.

## Locks are the escape hatch

Actors, immutable values, ownership transfer, and structured channels are the
default. Shared mutable memory uses guards:

```text
use guard = cache_state.lock(rank: CacheState)?
guard.value.record_hit(key:)
```

Candidate defenses:

1. `Mutex<T>` exposes `T` only through a lexical guard.
2. Guards are affine and cannot escape their lock/resource scope.
3. Ordinary guards cannot cross `await`, actor send, or unknown foreign calls.
4. Static lock ranks reject known order inversions.
5. Development/runtime lock dependency tracking records observed order and
   reports cycles for dynamic cases.
6. Try-lock and timeout return typed outcomes; they do not pretend to solve
   liveness.
7. Poison/invariant recovery, fairness, priority inheritance, and interrupt
   safety are explicit lock-type policies.
8. Every blocking wait participates in deadlock and causal diagnostics.

Static ranks cannot prove arbitrary data-dependent locking safe. Runtime
lock-graph validation and schedule testing complement them. “Deadlock-free” is
credible only for a restricted concurrency subset or a verified protocol.

## Cache semantics

A cache is a performance copy of an authoritative value unless explicitly
declared otherwise. If losing it loses truth, it is storage and must use storage
durability semantics.

```text
capability CustomerCache {
  get(
    key: CustomerCacheKey,
    freshness: Freshness,
  ) -> Result<CacheLookup<CustomerView>, CacheError>

  put(
    key: CustomerCacheKey,
    value: CustomerView,
    expires: Expiry,
  ) -> Result<Unit, CacheError>
}
```

Useful result states are not just hit/miss:

```text
data CacheLookup<T> =
  | Fresh { value: T, age: Duration }
  | Stale { value: T, age: Duration, revalidate_by: Instant }
  | Miss
  | Negative { reason: NegativeReason, age: Duration }
```

The caller or official policy chooses whether stale or negative values are
acceptable. Security/authentication/authorization results default to no
cross-request reuse unless their complete authority context is in the key and a
reviewed policy permits it.

## What can be shifted left

Cache correctness divides into three enforcement layers:

| Obligation | Compiler/type/effect system | Official cache toolkit/runtime | Application evidence |
|---|---|---|---|
| stable canonical key encoding | enforce | provide/version | compatibility fixtures |
| explicit semantic inputs | derive candidates and flag captures/effects | record resolved key fields | seeded omitted-input defects |
| tenant/auth/privacy partition | track classified values and capability scope | partition storage and redact evidence | adversarial cross-tenant tests |
| freshness and allowed staleness | require policy at use site | calculate age, validate, reject illegal stale use | domain oracle and outage tests |
| update/invalidation ordering | require revision/fencing type where selected | conditional writes/deletes and leases | reordered/dropped-event schedules |
| duplicate fills and overload | expose concurrency/budget contract | single-flight, admission, bounded queues | burst/load/fault tests |
| capacity and eviction | require bounded policy outside ephemeral local scope | enforce bytes/items/tenant quotas | memory-pressure tests |
| origin/business truth | cannot infer | cannot invent | author-supplied contract and end-to-end tests |

The compiler can prove that a declared pure computation depends only on named
semantic inputs. It cannot discover that an external record changed, decide how
stale a customer balance may be, or infer whether an authorization decision can
be shared. Those are typed policy inputs plus tests, not optimizer guesses.

Do not add a general `cache` keyword to the v1 grammar. Core types, effects,
stable serialization, clocks, revisions, and capabilities provide the semantic
hooks; a first-party declarative kit should prove that those hooks are enough.
Promote syntax only if repeated programs otherwise contain error-prone ceremony
that derivation cannot remove.

## Candidate declarative cache contract

```text
cached fn customer_view(
  customer: CustomerId,
  viewer: ViewerPolicy,
) -> Result<CustomerView, LoadError>
  with { Customers, CustomerViews }
using CustomerViews.policy(
  authority: Customers,
  key: derive(customer, viewer, schema CustomerView.v3),
  freshness: bounded(max_age: 30.seconds),
  validate: CustomerRevision,
  invalidate: on CustomerChanged(customer:, revision:),
  fill: single_flight(deadline: inherit),
  capacity: 128.mib per tenant,
  negative: [NotFound for 5.seconds],
)
```

This is candidate toolkit notation, not accepted core syntax. Its checked
manifest should make these facts explicit:

- authoritative source and whether cache failure falls back, fails, or serves
  explicitly permitted stale data;
- key fields, canonical encoding, namespace, code/schema/model/config versions,
  and privacy partition;
- freshness clock, age calculation, validation token, and stale policy;
- event/version ordering and conditional update/delete rule;
- fill ownership, coalescing, deadline, cancellation, and retry budget;
- item/byte/tenant capacity, admission, eviction, and overload response;
- cacheable error variants and their shorter negative lifetime;
- encryption, residency, redaction, and destruction policy where classified
  data is present;
- metrics/events required to establish correctness and value.

Expiry uses monotonic elapsed time inside one process. Wall-clock timestamps
are converted at protocol/storage boundaries with explicit skew and precision
rules; they are not trusted as a total order across distributed writers.

An external cache is an effect because it can fail, be stale, and consume
network authority. A bounded in-process memoization of a total pure function may
remain observationally transparent if eviction changes only performance. Its
resource cost and capacity are still visible.

## Cache key correctness

A typed key should include every semantic input that changes the output:

- stable function/query/schema version;
- tenant, principal/authorization scope where relevant;
- locale, timezone, feature/config/model version;
- normalized arguments and upstream data version;
- privacy partition and target/profile representation.

```text
record CustomerCacheKey {
  customer: CustomerId,
  tenant: TenantId,
  viewer_policy: PolicyVersion,
  schema: SchemaVersion,
}
```

The compiler can derive candidate keys for pure functions and flag a captured
effect/config value missing from an explicit key. It cannot know every external
source of truth. Raw secrets and PII should not be used as observable cache keys;
use scoped opaque digests when needed.

Key encoding is a versioned protocol, not the host language's incidental hash
or object layout. Map/set ordering, Unicode normalization, floats, time zones,
defaults, and schema evolution must canonicalize or be rejected. A key explains
its fields and provenance without revealing their sensitive raw values.

## Invalidation, staleness, and stampedes

Supported policies should compose explicitly:

- immutable content-addressed values with long lifetime;
- TTL/freshness with jitter to avoid synchronized expiry;
- validation/version tokens;
- event-driven invalidation with version checks;
- stale-while-revalidate and bounded stale-if-error;
- write-through/write-behind only with declared consistency and failure policy;
- negative caching with shorter, typed lifetime;
- request coalescing/single-flight per key;
- admission and LRU/LFU/size-aware eviction under a hard memory budget.

Retries and cache fills share the same deadline. Single-flight followers do not
extend the original work indefinitely; cancellation policy decides whether the
fill continues when the initiating caller leaves. A failed fill is not cached
forever. Refresh fan-out is bounded.

Domain-event invalidation is attractive, but dual writes to database and cache
are not atomic by wish. Use transactionally recorded outbox/version events or
accept a measured staleness window. A late invalidation must not delete a newer
version.

Use monotonically comparable revisions or fencing/lease tokens for cache fills
and invalidations. A fill begun before a write must not publish an older value
after the newer invalidation. Facebook's memcache design introduced leases to
arbitrate stale concurrent fills and reduce thundering herds, demonstrating that
single-flight and ordering are related correctness concerns rather than two
independent tuning knobs. ([Scaling Memcache at
Facebook](https://www.usenix.org/conference/nsdi13/technical-sessions/presentation/nishtala))

## Memoization

Pure, deterministic functions with stable value identity are eligible for
memoization, but it should not be silently automatic. Unbounded memoization is
a memory leak with a flattering name. The declaration includes key, scope,
capacity, eviction, and cost:

```text
memoize compile_schema
  scope: compiler_session
  key: semantic_inputs
  capacity: 512.mib
  eviction: cost_aware_lru
```

Compiler queries and asset/data stages benefit especially from content-addressed
memoization because inputs and tool versions are already known.

Build and compiler caches admit a stronger contract than application caches:
when actions are hermetic and deterministic, the cache key can identify all
semantic inputs and reuse becomes exact rather than freshness-based. Bazel's
remote-cache documentation also records the failure cases: host tools and
untracked environment can create incorrect shared hits, concurrent input
changes can upload invalid results, and an untrusted writer can poison the
cache. ([Bazel remote caching](https://bazel.build/remote/caching), [Bazel
hermeticity](https://bazel.build/basics/hermeticity))

Lang's build cache should therefore:

- hash compiler/runtime/backend versions, target/profile, flags, environment
  inputs, source/interface identities, dependencies, tools, and declared data;
- sandbox actions and reject undeclared reads, writes, time, randomness, and
  network access in cacheable mode;
- separate content-addressed blobs from signed action-result attestations;
- permit developer read access while limiting trusted shared writes to CI or an
  equivalent builder identity;
- verify blob digests on read and quarantine/diagnose inconsistent results;
- explain every hit, miss, non-cacheable action, and invalidation;
- compare hash/serialization/upload/download cost with recomputation before
  enabling remote reuse.

## Deterministic testing and evidence

Test providers control clock, cache capacity, eviction, invalidation delivery,
load failure, and concurrency schedules. Required scenarios include:

- hit, miss, stale, negative, eviction, and corrupt entry;
- concurrent cold miss and stampede suppression;
- write/invalidation reordering and dropped invalidation;
- clock jumps and expiry boundaries;
- provider outage, partial timeout, and memory pressure;
- tenant/authorization key separation;
- schema/code version change and rolling mixed versions.

Runtime events expose hit/miss/stale/negative rates, origin load saved and
amplified, fill latency/failure, coalesced waiters, evictions, capacity, key
cardinality, invalidation lag, and served-stale age. Hit rate alone can be high
while correctness, memory, or backend load is poor.

## Anti-patterns

- GC finalizers as the primary file/socket/lock cleanup mechanism.
- Hiding `commit`, `flush`, or graceful shutdown in an infallible destructor.
- A lock guard surviving an arbitrary suspension point.
- Assuming lock timeouts make a protocol deadlock-free.
- Using a cache as truth while configuring it for eviction.
- Cache keys that omit tenant, authorization, locale, schema, or code version.
- Automatic retries plus cache refresh plus client retry multiplying load.
- Identical TTLs causing synchronized expiry.
- Invalidating an unversioned key after a newer write.
- Unbounded memoization or a cache without a capacity policy.
- Treating a high cache-hit percentage as sufficient performance evidence.
- Using wall-clock timestamps as the only ordering defense across writers.
- Treating the language's ordinary object hash or serialization as a stable key
  protocol.
- Sharing cache entries across authority or privacy partitions because the data
  payload happens to be identical.
- Letting every library bring its own invisible cache, retry, worker pool, and
  freshness defaults.
- Trusting all developers or remote builders to publish executable cache
  artifacts without identity, provenance, or digest verification.

## Primary evidence

- [Rust `Drop`](https://doc.rust-lang.org/book/ch15-03-drop.html) demonstrates
  compiler-inserted lexical cleanup for resources such as files, sockets, and
  locks.
- [Go context](https://go.dev/blog/context) models cancellation/deadline trees
  so child work stops and releases resources when a request ends.
- [Linux lockdep design](https://docs.kernel.org/5.17/locking/lockdep-design.html)
  demonstrates runtime lock-class and acquisition-dependency cycle validation.
- [RFC 9111](https://www.rfc-editor.org/rfc/rfc9111.html) defines HTTP cache
  keys, freshness, validation, `Vary`, and privacy considerations.
- [Go singleflight](https://pkg.go.dev/golang.org/x/sync/singleflight) provides
  duplicate in-flight call suppression per key.
- [Redis eviction](https://redis.io/docs/latest/develop/reference/eviction/)
  demonstrates that hard memory limits, eviction policy, expiration, and
  observability materially affect cache behavior.

All sources were last verified 2026-09-02. Syntax and generalized policies are
project synthesis.
