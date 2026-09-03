---
id: domain-data-distribution
title: Domain, data, and distributed-system semantics
summary: Placement decisions for domain modeling, CQRS, event sourcing, consistency, resilience, storage, and deterministic simulation.
type: design
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-02
tags: [domain-modeling, distributed-systems, cqrs, event-sourcing, resilience]
related: [vision, effects-and-capabilities, compiler-feedback-latency, context-telemetry-security, example-tour, research-ledger]
---

# Domain, data, and distributed-system semantics

## Recommendation

Make domain facts legible to the compiler without making one enterprise
architecture mandatory. The language core should know nominal data, variants,
contracts, effects, versioned schemas, capabilities, and state transitions.
An official service toolkit can define commands, queries, events, aggregates,
projections, retries, streams, and persistence protocols on top.

This yields a placement rule:

| Concern | Placement |
|---|---|
| Domain identity and invariants | Core types and contracts |
| Read versus write authority | Capability operations/effects |
| Commands, events, aggregates, projections | Official declarative kit, visible in semantic graph |
| Consistency and transaction promises | Typed adapter contracts |
| Retry, timeout, backpressure, pooling | Runtime primitives plus official policies |
| Load balancing and service discovery | Host/runtime adapter, not source-language grammar |
| Durable database implementation | External adapter or specialized official kit |
| Recovery evidence | Toolchain/deployment policy |

## Nouns, verbs, and events

- **Nouns** are opaque or nominal identifiers, records, and algebraic data
  types. `OrderId` and `PaymentId` do not unify merely because both wrap text.
- **Verbs** are functions or capability operations. Their results, errors,
  effects, authorization, and idempotency are part of the contract.
- **Events** are immutable, versioned facts represented as closed variants at a
  component boundary and open/version-aware schemas at a distributed boundary.

```text
opaque type OrderId = Uuid

type PlaceOrder = {
  order_id: OrderId,
  customer: CustomerId,
  lines: NonEmpty<OrderLine>,
}

event OrderEvent version 1 {
  OrderPlaced { order_id: OrderId, total: Money }
  PaymentDeclined { order_id: OrderId, reason: DeclineReason }
}
```

`event` is shown as candidate declarative-kit notation, not a committed keyword.
It must lower to ordinary typed data and version metadata that every tool can
inspect. If an annotation or derivation provides the same leverage with less
grammar, prefer it.

## Structural behavior, nominal meaning

Use structural typing for small behavioral ports and nominal typing for domain
meaning.

```text
capability OrderReader {
  find(id: OrderId) -> Result<Option<Order>, ReadError>
}

// No `implements OrderReader` is required.
type PostgresOrders { ... }
fn PostgresOrders.find(id: OrderId) -> Result<Option<Order>, ReadError> with Database
```

Go interfaces are satisfied implicitly by matching method sets. This supports
consumer-defined, narrow ports and easy test doubles without coupling a
provider to every consumer interface.
([Go specification](https://go.dev/ref/spec))

Do not extend that rule blindly to public domain data. TypeScript documents a
structural compatibility model and also documents deliberate unsoundness in
parts of it. Two equal record shapes can still mean different security,
currency, tenancy, lifecycle, or validation states.
([TypeScript compatibility](https://www.typescriptlang.org/docs/handbook/type-compatibility))

Candidate hybrid:

- capabilities/interfaces match structurally by default;
- public domain types, IDs, validated values, and resources are nominal;
- private anonymous records may be structural and row-polymorphic;
- public/foreign/security boundaries can require an explicit conformance
  witness even when method matching is structural;
- interface satisfaction is queryable, and accidental future satisfaction can
  be linted when authority is involved.

## Command-query separation and CQRS

Command-query separation is a good local default: a query returns information
without changing externally visible state; a command may change state and
returns a receipt, identifier, or explicit outcome rather than a convenient
copy of a read model.

```text
capability Orders.Read {
  get(id: OrderId, consistency: ReadConsistency)
    -> Result<Option<Observed<Order>>, ReadError>
}

capability Orders.Write {
  place(command: PlaceOrder, idempotency_key: IdempotencyKey)
    -> Result<CommitReceipt<OrderEvent>, PlaceOrderError>
}
```

Separate capabilities let architecture policy, deployment, tests, and telemetry
distinguish read from write authority. They do not require CQRS. CQRS—the
architectural use of distinct write and read models—adds projection lag,
operational components, schema evolution, and repair work. It should be an
official template chosen when scaling or domain needs justify it, not the
default for a small program.

## Consistency is a promise, not a flag

Consistency belongs in the operation contract and must be supported by the
provider. Candidate policies include `linearizable`, `snapshot`, `causal`, and
`eventual(max_staleness:)`. A generic boolean such as `strong: true` is too
lossy.

```text
type Observed<a> = {
  value: a,
  version: Version,
  observed_at: Instant,
}

let order = Orders.Read.get(
  id: order_id,
  consistency: eventual(max_staleness: 2.seconds),
)?
```

The type checker can ensure that an adapter claims the requested contract; only
the implementation and production evidence can establish that the claim is
true. Cross-service transactions are never inferred from local syntax.

## Event sourcing and immutable audit

Event sourcing deserves an official kit because typed effects and pure state
transitions make it unusually testable:

```text
aggregate Order {
  state: OrderState
  command: OrderCommand
  event: OrderEvent

  fn decide(state:, command:)
    -> Result<NonEmpty<OrderEvent>, OrderError>

  fn evolve(state:, event:) -> OrderState
}
```

Exact-name argument punning—`state:` meaning `state: state`—removes repetition
without dropping the role label. The formatter owns this canonical shorthand.

Akka's event-sourcing documentation describes append-only events, rebuilding
state through replay or snapshots, pure event handlers, and single-writer
requirements per persistence identity. These are obligations the kit must make
visible, not details to hide.
([Akka Event Sourcing](https://doc.akka.io/libraries/akka-core/current/typed/persistence.html))

The kit must address:

- deterministic `evolve` functions and replay;
- optimistic concurrency and a single logical writer;
- event, snapshot, actor-state, and projection migrations;
- idempotent consumers and deduplication scope;
- projection lag, repair, and rebuild;
- retention, storage growth, encryption, and erasure/privacy conflicts;
- snapshots as optimization, never the only source of durable truth;
- distinction between domain events, integration events, telemetry, and audit
  records.

An immutable audit log is not automatically an event store, and neither is a
backup. Each has different access, retention, privacy, and recovery contracts.

## Resilience policies

The happy path for a remote operation should require a deadline and bounded
resource policy. Runtime/official-kit primitives should include:

- timeouts and propagated deadlines;
- cancellation;
- retry classification, capped exponential backoff, and jitter;
- idempotency keys and deduplication contracts;
- one declared retry owner across a call chain;
- circuit breaking, bulkheads, rate limits, and load shedding;
- bounded queues and non-blocking backpressure;
- connection-pool capacity, wait deadlines, and health evidence;
- readiness, health, service discovery, and load-balancer adapters.

AWS guidance emphasizes idempotent APIs, bounded retry attempts, exponential
backoff with jitter, and retrying at one layer to avoid multiplying load.
([Making retries safe](https://aws.amazon.com/builders-library/making-retries-safe-with-idempotent-APIs/),
[Limit retries](https://docs.aws.amazon.com/wellarchitected/2023-04-10/framework/rel_mitigate_interaction_failure_limit_retries.html))

Retries must never be an invisible default. The compiler can reject a retry
policy without an idempotency or explicit at-most-once risk contract. Runtime
events expose attempts, delay, owner, budget, and final cause.

Reactive Streams defines non-blocking backpressure so a fast producer cannot
overwhelm a slower consumer. The core runtime should own bounded demand,
cancellation, and terminal-state semantics; transformations and connectors can
live in the toolkit.
([Reactive Streams](https://www.reactive-streams.org/))

## Built-in storage: narrow and honest

Provide a stable runtime-local table/cache interface and an official embedded
durable key-value/log interface. On BEAM, a local table can map to ETS. This is
valuable for actor state, memoization, indexes, rate limits, and test worlds.

Do not promise that a bundled database eliminates Postgres, Redis, or a durable
log in general. Erlang's own documentation presents ETS as in-memory term
storage and Mnesia as a distributed DBMS with transactions, replication,
backup, and operational tradeoffs. Those mechanisms are powerful precisely
because their scope and failure model are specific.
([Erlang tables and databases](https://www.erlang.org/doc/system/tablesdatabases.html),
[Mnesia overview](https://www.erlang.org/docs/25/apps/mnesia/mnesia_overview.html))

Storage adapters declare durability, consistency, replication, eviction,
capacity, transaction, backup, restore, and partition behavior. `Cache` is not
an alias for truth.

## Deterministic worlds, replay, and time travel

The effect model enables a high-leverage testing feature: run the entire
program against a deterministic world that supplies clock, randomness,
network, disk, scheduler, process failure, and external responses.

```text
simulation checkout_partition {
  seed: 0x82f4
  world: ServiceWorld.deterministic()

  fault Network.partition(between: [checkout, inventory], at: 3.seconds)
  fault Process.crash(actor: OrderCoordinator, after_events: 2)

  always Inventory.reserved_not_oversold
  eventually Orders.terminal_or_compensated(within: 30.seconds)
}
```

FoundationDB tests an entire simulated cluster with simulated time and injected
machine, network, and datacenter failures; deterministic seeds make failures
reproducible when uncontrolled nondeterminism is absent. This is unusually
aligned with typed, replaceable effects.
([FoundationDB simulation](https://apple.github.io/foundationdb/testing.html),
[client testing](https://apple.github.io/foundationdb/client-testing.html))

Claims must remain bounded:

- replay is exact only for captured effects and deterministic code;
- real kernels, networks, CPUs, foreign libraries, and hardware still need
  integration and fault-injection tests;
- a recorded production session may contain secrets or PII and needs the same
  classification, retention, and access controls as telemetry;
- multithreading and racy foreign code can break reproducibility.

Optional state-machine specifications may compile to a model-checking format
or run in a bounded checker. This should supplement executable code, not imply
that arbitrary distributed programs have been proven correct.

## Schema and state evolution

Evolution is part of semantics from the first durable byte:

- wire fields and variants have stable identities;
- deleted identities are reserved and cannot be reused;
- unknown fields/variants can be retained where the protocol supports it;
- compatibility is checked in both producer and consumer directions;
- actor state, snapshots, event streams, database schemas, and projections each
  have explicit migration paths;
- rolling deployment checks old/new coexistence, downgrade, and rollback.

Protocol Buffers documents why removed numeric identifiers should be reserved,
how unknown fields are preserved in binary messages, and why some nominally
wire-compatible changes still require a coordinated rollout. These are strong
defaults for a language-owned schema tool.
([Protocol Buffers evolution](https://protobuf.dev/programming-guides/proto3/))

Hot reload is straightforward for stateless compatible functions. Stateful hot
upgrade requires a checked state/protocol migration. Erlang release handling
supports code replacement but documents explicit upgrade instructions,
old/new-code coexistence, state transformation, and the difficulty introduced
by complex dependencies. “Hot reload everything” is therefore not a safe
default.
([Erlang release handling](https://www.erlang.org/doc/system/release_handling.html))

## Backup and disaster recovery

For every durable provider, the deployment contract should declare:

- recovery point objective and recovery time objective;
- backup scope, cadence, retention, encryption, and ownership;
- restore procedure and last successful restore test;
- region/failure-domain assumptions;
- schema and application version compatibility;
- behavior when a dependency is unavailable or restored stale.

The toolchain can generate manifests, schedule tests, reject missing policy, and
attach restore evidence to releases. The language cannot make an untested
infrastructure recovery claim true.
