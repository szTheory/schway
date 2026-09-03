---
id: effects-and-capabilities
title: Effects and capabilities as the semantic spine
summary: A candidate model unifying dependency injection, architecture, testing, security, telemetry, and distribution.
type: design
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-03
tags: [effects, capabilities, dependency-injection, architecture]
related: [vision, semantic-kernel-contract, semantic-kernel-probes, trust-validation-information-flow, boundary-data-validation-persistence, design-atlas, feature-coherence, compiler-feedback-latency, domain-data-distribution, context-telemetry-security, example-tour, research-ledger]
---

# Effects and capabilities as the semantic spine

## Proposal

Use four related but distinct concepts:

- A **value type** says what data is.
- An **error type** says which expected failures are returned.
- An **effect set** says which external or nonlocal operations may occur.
- A **capability handler** supplies the authority and implementation for those
  operations in a lexical scope.

Conceptually:

```text
fn place_order(command: PlaceOrder)
  -> Result<Order, PlaceOrderError>
  with { Inventory, Payments, Clock }
```

The function cannot access a database, network, clock, random source, secret,
filesystem, process launcher, or model unless its declared capabilities permit
it. Tests install deterministic handlers; production installs adapters.

For the first ownership-capable kernel, “handler” means a statically resolved,
one-shot provider call—not a first-class resumable continuation. General
one-shot handlers remain experimental; multi-shot handlers are outside v1
because duplicating or discarding a continuation can duplicate or abandon
captured linear resources. The precise executable boundary is in the
[semantic kernel contract](semantic-kernel-contract.md).

Koka provides evidence that effect rows can be inferred and polymorphic, while
Unison demonstrates effect interfaces paired with handlers. These establish
feasibility of the semantic mechanism, not the ergonomics of this proposal.
([Koka](https://koka-lang.github.io/koka/doc/book.html),
[Unison](https://www.unison-lang.org/docs/fundamentals/abilities/writing-abilities/))

Roc supplies another important model. Current Roc syntax distinguishes pure
functions (`->`) from effectful functions (`=>`), uses `!` in effectful function
names, and lets the application's platform supply all I/O. A Roc application is
built against exactly one platform, whose host controls I/O and other runtime
services. The high-leverage ideas to borrow are no ambient standard I/O and an
explicit host contract. This proposal keeps capabilities independently
composable rather than requiring one monolithic platform object.
([Roc functions and effects](https://roc-lang.org/functional),
[Roc platforms](https://www.roc-lang.org/docs/main/langref/platforms/))

## Why this is higher leverage than ordinary DI

An interface parameter can swap implementations, but a typed capability can
also answer:

- Which authority does this function require?
- Which layer introduced that authority?
- Can the call run in a sandbox, test, browser, embedded target, or remote node?
- Which operations should create trace spans?
- Which failures, retries, and costs cross the boundary?
- Can the compiler prove that domain code is deterministic?
- What is the smallest handler graph needed to execute this test?

This makes the dependency graph semantic rather than framework metadata.

## First-class provider graph, not a service locator

The useful part of a DI container should be built in: construction, lifetimes,
cleanup, default bindings, scoped overrides, cycle detection, and a queryable
graph. Runtime name lookup, reflection, hidden globals, and decorator scanning
should not be.

```text
service CheckoutApi {
  provide Database = Postgres.connect(config: CheckoutConfig.database_url)
    lifetime service
    close Database.close

  provide Orders = PostgresOrders(database: Database)
  provide Inventory = InventoryClient(endpoint: CheckoutConfig.inventory_url)
  provide Payments = StripePayments(secret: Secrets.require(name: "stripe"))
  provide Clock = Clock.system()

  scope request from Http.request {
    provide RequestExecution = RequestExecution.from_http(request: Http.request)
    provide AuthenticatedRequest = authenticate(request: Http.request)?
  }

  expose http = CheckoutHttp()
}
```

The compiler resolves a closed construction graph, reports missing or ambiguous
providers, rejects cycles, validates lifetimes, and generates deterministic
startup/shutdown code. The graph is available through compiler and runtime
queries. A request-scoped value cannot leak into a service-scoped provider.

Candidate lifetimes are `service`, `request`, `task`, `actor`, and `call`. New
lifetimes require evidence; a menu of framework-specific scopes would recreate
container complexity.

Default precedence is lexical and visible:

```text
run target
  with lexical test/debug overrides
  over application composition bindings
  over explicit platform defaults
```

There are no process-global overrides. Libraries may ship constructors and
explicit development/test providers, but may not silently choose production
infrastructure. The application composition root owns deploy defaults.

Most injection should correspond to authority or nondeterminism—database,
network, time, randomness, filesystem, model, secrets, foreign code—because
those are effects worth substituting and auditing. “Only DI effects” is too
absolute, however. A pure runtime-selected policy or algorithm may also be
supplied as an ordinary typed value. Pure helpers with no substitution job use
normal functions/modules, not the provider graph.

This avoids two opposite failures:

- passing every capability manually until signatures become plumbing; and
- hiding every dependency behind an ambient service locator.

Provider synthesis is intentionally narrower than general implicit term
inference. It resolves declared provider types only at composition boundaries;
every synthesized argument can be rendered explicitly, and the explanation
includes the full candidate path.

## Candidate surface

```text
capability Inventory {
  reserve(items: NonEmpty<OrderLine>)
    -> Result<Reservation, InventoryError>

  release(reservation: Reservation) -> Unit
}

fn place_order(command: PlaceOrder)
  -> Result<Order, PlaceOrderError>
  with { Inventory, Payments, Clock }
{
  let reservation = Inventory.reserve(items: command.lines)
    |> map_error(error: _, using: PlaceOrderError.from_inventory)?

  let charge = Payments.charge(
    amount: command.total,
    idempotency_key: command.id,
  )
    |> on_error(action: fn(error:) {
      Inventory.release(reservation:)
      return error
    })?

  Ok(Order.confirmed(
    command:,
    reservation:,
    charge:,
    confirmed_at: Clock.now(),
  ))
}
```

An exact-name argument may use **label punning**: `reservation:` is canonical
shorthand for `reservation: reservation`. It preserves the semantic role and
reorder-stable call shape while removing pure repetition. It applies only when
the local identifier exactly matches the parameter label; renaming either side
expands or updates the call through a semantic edit. The formatter must not
invent implicit positional binding.

All multi-argument calls use names. The `with` row is explicit on public
functions. The `?` operator only propagates a typed `Result`; it never throws.
The exact cleanup
syntax is unresolved because compensation, transactions, cancellation, and
panic need one coherent resource model.

`Clock.now()` is an effectful call. The qualifier says which capability
operation is invoked; the function's `with { Clock }` contract grants that
authority; and the provider graph supplies `Clock.system()` in production or
`Clock.fixed(...)` in a test. Whether the call itself also needs an effect marker
such as Roc's `Clock.now!()` is a syntax experiment. The semantic model must not
depend on punctuation to detect the effect.

## Inference policy

Candidate balance:

- Infer effects within private functions.
- Print the inferred row in editor hovers and agent queries.
- Require an explicit effect contract on public component boundaries.
- Reject an implementation whose inferred effects exceed its public contract.
- Offer a machine-applicable edit to widen the contract, but identify which
  dependency introduced the new authority.
- Allow effect-polymorphic higher-order functions without forcing callers to
  spell the propagated row.

Example:

```text
fn map(items: List<a>, transform: fn(item: a) -> b with effects)
  -> List<b>
  with effects
```

The row variable `effects` is ordinary to the compiler but can be rendered as
“whatever `transform` requires” for humans.

## Architecture enforcement

Architecture should use general dependency and capability rules rather than
hard-code the words “hexagonal” or “clean architecture” into the grammar.

```text
policy service_architecture {
  component domain {
    allow capabilities: {}
    allow dependencies: { domain }
  }

  component application {
    allow capabilities: { Clock, Inventory, Payments }
    allow dependencies: { domain, ports }
  }

  component adapters {
    allow capabilities: { Database, Network, Secrets, Trace }
    allow dependencies: { ports }
  }
}
```

The compiler can reject an SQL call in `domain`, a concrete Stripe dependency
in `application`, or a cycle between components. Small programs use a default
single-component policy and pay no layering ceremony.

## Testing and specifications

Handlers make effects deterministic without runtime patching:

```text
spec place_order {
  example "payment decline releases inventory" {
    run place_order(command: fixtures.valid_order)
      with Inventory = Inventory.memory(stock: fixtures.stock)
      with Payments = Payments.decline(reason: card_declined)
      with Clock = Clock.fixed(at: @2026-09-02T12:00:00Z)

    expect result == Err(PaymentDeclined(reason: card_declined))
    expect Inventory.events == [Reserved, Released]
  }
}
```

The compiler knows that every `PlaceOrderError` variant should be reachable or
justified. It can derive valid structural generators from `PlaceOrder`, propose
boundary values, and save minimized failures. The developer still supplies
properties and expected outcomes.

Tests can override one leaf while reusing the resolved world:

```text
spec CheckoutApi {
  world test = CheckoutApi.providers
    with Database = Database.memory()
    with Payments = Payments.fake()
    with Clock = Clock.fixed(at: @2026-09-02T12:00:00Z)

  example "decline releases inventory" using test {
    run place_order(command: fixtures.valid_order)
      with Payments = Payments.decline(reason: CardDeclined)

    expect Inventory.events == [Reserved, Released]
  }
}
```

An override is type-checked against the same contract, and its scope appears in
the evidence manifest. Overrides cannot alter unrelated bindings or escape the
test task.

## Telemetry and operations

Each capability operation is a natural semantic span. A production handler can
emit a typed event without application wrappers:

```text
CapabilityCall {
  capability: Payments,
  operation: charge,
  caller: shop.orders/place_order,
  trace: 7f...,
  duration: 83.ms,
  outcome: Err(CardDeclined),
  attempts: 1,
  cost: Money(usd: 0.0002),
}
```

OpenTelemetry already standardizes correlated traces, metrics, logs, resources,
and context propagation. The runtime should export compatible signals while
retaining a richer typed internal event model. Arbitrary baggage must be
treated as untrusted and potentially sensitive. ([OpenTelemetry overview](https://opentelemetry.io/docs/specs/otel/overview/),
[context propagation](https://opentelemetry.io/docs/concepts/context-propagation/))

See [Context, telemetry, privacy, and operations](context-telemetry-and-security.md)
for propagation, classification, budgets, wide-event projections, and live
agent inspection.

## Concurrency and distribution

Effects do not replace structured concurrency or actors. They describe the
authority to create and communicate with them.

- `Task` permits finite child work whose lifetime is scoped.
- `Actor<Message>` permits messages to long-lived isolated state.
- `Remote<World>` permits crossing a node boundary with an explicit interface.
- A handler decides local thread pool, BEAM process, test simulation, or remote
  transport.

Location transparency should stop at the source boundary. Remote operations
must retain visible latency, partial failure, versioning, cancellation, and
delivery semantics even if the handler syntax looks similar.

Timeout, retry, exponential-backoff/jitter, pooling, backpressure, circuit
breaking, and load shedding should compose as typed policies around remote
capabilities. They belong in runtime primitives and official kits rather than
becoming hidden behavior of every network call. See
[Domain, data, and distributed-system semantics](domain-data-and-distribution.md).

## Effects that may deserve first-class status

| Effect/capability | Why it matters | Default constraint |
|---|---|---|
| State | Distinguishes local mutation from pure transformation | Lexically isolated; not ambient |
| Allocate | Makes embedded and cost profiles analyzable | Usually inferred/measured, explicit in restricted profiles |
| Clock | Enables deterministic tests | No ambient wall clock |
| Random | Enables deterministic replay | Seeded handler in tests |
| Filesystem | High authority and platform variance | Path/root capability scoped |
| Network | Distributed failure and data exfiltration | Destination/protocol scoped where practical |
| Secrets | Sensitive data provenance | Non-serializable, redacted, propagation-restricted |
| Database | Transactions, latency, schema contracts | Abstract port in application layer |
| Trace | Observation policy | Automatic at effect boundaries; payload opt-in |
| Task/Actor | Lifetime and concurrency | Structured/bounded defaults |
| Remote | Partial failure and versioning | Never inferred from an ordinary local call |
| Model | Nondeterminism, privacy, tokens, monetary cost | Schema, budget, cache, and provenance required |
| Foreign | Trust boundary | Ownership/threading/blocking contract required |
| Unsafe | Guarantee escape | Smallest lexical scope, rationale, audit event |

## Footguns and guardrails

### Effect soup

If routine signatures contain dozens of implementation effects, readability
collapses. Use abstract domain capabilities at application boundaries and lower
them to concrete database/network/trace effects inside adapters.

### Handler-order semantics

Stacked handlers can make behavior depend on order. Prefer named handler graphs
with explicit dependencies. Reject ambiguous compositions.

### Lifetime capture

A service provider retaining request/task context leaks identity, memory, and
authority across requests. Provider lifetime checking should be analogous to a
borrow check over scopes: a longer-lived value cannot capture a shorter-lived
provider unless it copies an explicitly transferable value.

### Authority lifecycle

A capability proves that a value is the only route to an operation; it does not
prove the provider, credential, lease, principal, or policy remains valid
forever. Long-lived capabilities may be attenuated to a narrower operation or
resource set, delegated with an explicit lifetime and budget, expire, be
revoked, or fail because their provider is shutting down.

Core typing preserves scope and prevents an attenuated capability from
regaining authority. Runtime revocation remains an expected typed operation
failure and a causal event, not a retroactive type error. Code that requires
irrevocable local authority uses a shorter lexical capability backed by owned
state; remote credentials, agent tools, debug leases, and rotating secrets
must assume revocation.

### Default-provider drift

If dependencies acquire defaults everywhere, behavior becomes import-sensitive
and surprising. Only composition roots and explicit platform profiles select
defaults. Library providers require opt-in imports and appear in resolution
explanations.

### Ambient context as authorization

Trace/deadline propagation is not proof of identity or tenancy. Adapters verify
security context and pass authorization-relevant values into application/domain
commands explicitly.

### Masking authority

A function must not relabel `Network + Secrets` as harmless `CustomerLookup`
without the adapter boundary remaining queryable. Abstraction hides details from
callers, not from audit tools.

### Resumable continuation hazards

Multi-shot handlers may duplicate a continuation that owns a socket, lock,
transaction, or unique value. Start with one-shot handlers. Add resumability
only alongside an ownership rule that proves duplication safe.

### Retry multiplication

Nested handlers can each retry, producing exponential calls and cost. Retry is
a typed policy with a single owner, an idempotency requirement, attempt budget,
and trace representation.

### False purity through foreign code

Foreign declarations are effectful by default. Removing effects requires an
audited contract and cannot guarantee more than the foreign implementation and
ABI actually provide.

## Questions that require prototypes

1. Can effect rows stay understandable on realistic higher-order service code?
2. Can the compiler infer useful effects incrementally without slowing the edit
   loop?
3. Should error variants be part of the effect row or remain ordinary results?
4. Can ownership and one-shot effects share a simple model for resources?
5. Which handler operations lower efficiently to BEAM, native, and Wasm?
6. Can typed events be captured with low enough overhead for default production
   use?
7. Does effect-aware context improve agent edit correctness versus the same
   programs in Gleam, Rust, and Elixir?
8. Can provider lifetime errors be explained more simply than ownership errors
   while preventing request-context leaks?
9. Does `Clock.now()` versus `Clock.now!()` produce better agent accuracy and
   human auditing once effects are already known to the compiler?
10. Is structural satisfaction safe for ordinary capability ports, and where
    should public/security boundaries require an explicit witness?
