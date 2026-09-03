---
id: application-architecture-data
title: Application architecture and data boundaries
summary: A verified modular-monolith default with explicit identity, validation, SQL, contracts, workflows, and extraction boundaries.
type: design
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-03
tags: [architecture, monolith, data, validation, contracts]
related: [vision, domain-data-distribution, boundary-data-validation-persistence, trust-validation-information-flow, effects-and-capabilities, feature-coherence, type-system-failure-semantics, example-tour, research-ledger]
---

# Application architecture and data boundaries

## Recommendation

Make a **verified modular monolith** the service-profile default: one deployable,
one runtime, and one primary transactional consistency boundary until evidence
justifies separation. Within it, components expose typed ports, own their data,
form a directed acyclic graph, and cannot access another component's internals.

This does not make remote systems second-class. A remote boundary is a stricter
contract that adds serialization, versioning, deadlines, cancellation,
idempotency, authentication, partial failure, telemetry, and rollout evidence.
Extraction is explicit because a local call and a network call are not
semantically interchangeable.

The detailed JSON/presence, validation-freshness, change, transaction-retry,
and connection-pool contract is in [Boundary data, validation, transactions,
and pools](boundary-data-validation-and-persistence.md). This note retains the
application-architecture view; the boundary note owns those lower-level laws.

Spring Modulith demonstrates useful enforceable rules—no module cycles and
access through module APIs only—without requiring microservice deployment.
([Spring Modulith verification](https://docs.spring.io/spring-modulith/reference/verification.html))

## Architecture contract

```text
application Shop deploy monolith {
  component Catalog {
    expose { Products.Read }
    own storage { products }
  }

  component Orders {
    use { Catalog.Products.Read, Payments.Charge }
    expose { Checkout.PlaceOrder, Orders.Read }
    own storage { orders, order_events }
  }

  component Admin {
    use { Orders.Read }
  }
}
```

This is candidate manifest/projection syntax, not necessarily core grammar. Its
semantic graph supports:

- cycle and internal-access rejection;
- provider lifetime and authority checks;
- data ownership and migration ownership;
- affected checks/specs after a change;
- architecture diagrams generated from truth rather than annotations;
- a report showing what prevents component extraction;
- one-process calls marked `local`, remote adapters marked `remote`.

A component may use the monolith's primary transaction only across a declared
consistency boundary. The default is not “every table is globally reachable.”
Shared writes require an explicit owner or coordinated application operation.

## When to extract a service

Do not extract for fashion, namespace organization, or imagined future scale.
Consider extraction when at least one persistent force exists:

- independent scaling or hardware placement;
- independent failure containment or security trust zone;
- independent ownership/release cadence;
- data sovereignty or regulatory isolation;
- a protocol boundary already exists for external consumers.

Before extraction, the tool should enumerate missing remote obligations:
serialization, compatibility, deadlines, idempotency, delivery/ordering,
authentication, telemetry, deployment coexistence, and contract verification.

An anti-corruption adapter translates concepts when models disagree. It applies
between components in one process just as it does around a legacy or remote
system. It should be used for semantic mismatch, not as ceremony around every
call; it adds code, latency at remote boundaries, and maintenance.
([Anti-corruption layer](https://learn.microsoft.com/en-us/azure/architecture/patterns/anti-corruption-layer))

## Identity: opaque first, UUIDv7 by profile

Domain code sees `Id<Entity>`, never “a UUID string” or “an integer primary
key.” The service-profile default candidate is UUIDv7 because it supports
coordination-free generation and time-ordered database locality. The default
provider is still a capability so tests, imports, databases, and privacy policy
can replace it.

```text
opaque type PlayerId = Id<Player>

fn register_player(input: NewPlayer)
  -> Result<Player, RegisterError>
  with { Ids<Player>, Clock }
{
  Ok(Player(
    id: PlayerId.new(),
    joined_at: Clock.now(),
    name: input.name,
  ))
}
```

Candidate profile defaults:

| Context | Default candidate | Reason |
|---|---|---|
| Distributed/service domain identity | UUIDv7 behind `Id<Entity>` | Local generation plus sortable locality |
| Purely local dense array/index | bounded integer/index type | Compactness, locality, arithmetic bounds |
| Database-internal surrogate | provider-selected | Let storage evidence choose |
| Deterministic name identity | namespaced deterministic ID | Reproducible mapping, explicit collision domain |
| Security bearer token | cryptographic secret/token type, not UUID | Identity is not authority or unguessability |
| Foreign/legacy identity | opaque wrapper over foreign representation | Preserve compatibility without contaminating the domain |

RFC 9562 defines UUIDv7 with a Unix-epoch millisecond time field and recommends
binary 128-bit database storage where feasible. It also warns that UUIDs are
not security capabilities and should not be assumed hard to guess. Time order
can leak creation chronology, and monotonic generation under clock rollback or
high same-millisecond volume needs a specified provider policy. UUIDv7 is a
strong service default, not a universal primary-key law.
([RFC 9562](https://www.rfc-editor.org/rfc/rfc9562))

## Boundary decoding and validation

JSON is a first-party codec/schema capability, not core language syntax.
Boundary decoding has two phases:

```text
untrusted bytes
  -> syntactic Json.Value
  -> Decode<CreatePlayerRequest>
  -> validated nominal CreatePlayer
  -> domain operation
```

```text
record CreatePlayerRequest derives { Json.Schema } {
  name: Text,
  email: Text,
}

record CreatePlayer {
  name: PlayerName,
  email: EmailAddress,
}

fn CreatePlayer.from_request(request: CreatePlayerRequest)
  -> Validation<CreatePlayer>
{
  validate CreatePlayer(
    name: PlayerName.parse(text: request.name),
    email: EmailAddress.parse(text: request.email),
  )
}

fn create_player(request: HttpRequest) -> Result<HttpResponse, RequestError> {
  let wire = Json.decode<CreatePlayerRequest>(
    bytes: request.body,
    mode: strict,
    unknown_fields: reject,
    duplicate_fields: reject,
    limits: JsonLimits.default,
  )?

  let input = CreatePlayer.from_request(request: wire)?
  ...
}
```

The codec reports path-structured errors, source spans when available, all safe
independent validation failures, byte/depth/item limits, and exact coercions.
Strictness is explicit. Duplicate field behavior must not depend on whichever
parser happened to be installed: RFC 8259 notes that duplicate-name behavior
varies between implementations and that interoperable objects use unique
names. It also documents number precision interoperability limits.
([RFC 8259](https://www.rfc-editor.org/rfc/rfc8259))

Derived JSON Schema/OpenAPI is useful for interoperability, but generated
schema is an artifact with a version and semantic diff. OpenAPI 3.1 bases its
data types on JSON Schema 2020-12, which makes that a practical first export
target. ([OpenAPI 3.1](https://spec.openapis.org/oas/v3.1.0))

## Changes and persistence: Ecto-inspired, not ActiveRecord-shaped

Persistence should use explicit data and repository capabilities:

- records do not save themselves;
- no global connection or ambient transaction;
- no hidden lazy loading or implicit query on field access;
- queries are typed values/resources with visible cardinality and effects;
- validation is pure where possible; database constraints are checked by the
  repository and translated to typed outcomes;
- updates are explicit `Change<T>` values that preserve field-level errors and
  authorized input fields;
- migrations, schema snapshots, and compatibility are versioned artifacts.

```text
fn change_profile(player: Player, input: UpdateProfileRequest)
  -> Change<Player>
{
  change(value: player)
    |> cast(input:, allow: [.display_name, .bio])
    |> validate(.display_name, using: PlayerName.validate)
    |> validate(.bio, using: Text.max_graphemes(count: 280))
}

query find_player(id: PlayerId) -> Option<PlayerRow> using schema ShopDb {
  sql"""
  select id, display_name, bio
  from players
  where id = :id
  """
}

fn update_profile(id:, input:)
  -> Result<Player, UpdateError>
  with { Players.Write }
{
  Players.Write.transaction(fn(repo:) {
    let player = repo.get_for_update(id:)?
    let changed = change_profile(player:, input:).require_valid()?
    repo.update(change: changed)?
  })
}
```

The typed SQL literal is a candidate toolkit form compiled against a schema
snapshot, not an ORM grammar. It keeps SQL's real semantics visible and can
report nullability, cardinality, indexes, parameters, and migration drift.
Ecto changesets provide useful precedent for filtering external fields,
casting, validation, constraints, and schemaless changes.
([Ecto Changeset](https://hexdocs.pm/ecto/Ecto.Changeset.html))

## Contract testing

Contract tests belong at independently evolving component, service, event, and
foreign boundaries. Use two complementary layers:

1. **Schema compatibility:** directional checks on fields, variants, effects,
   errors, and protocol versions.
2. **Consumer examples:** the consumer records the minimal request/response or
   message behavior it actually depends on; provider verification replays those
   examples against the provider.

```text
contract Checkout consumes InventoryApi {
  interaction "reserve available stock" {
    given InventoryState.in_stock(sku: fixtures.sku, count: 3)

    let response = InventoryApi.reserve(
      items: [Line(sku: fixtures.sku, quantity: 2)],
      idempotency_key: fixtures.order_id,
    )

    expect response matches Ok(Reservation(id: _, expires_at: _))
  }
}
```

Pact documents this consumer-driven shape: consumer tests define minimal
expectations and provider verification replays them without needing both
services running together. Contract tests do not replace domain properties,
integration tests, or a small number of end-to-end journeys; they answer a
narrow compatibility question. ([Pact](https://docs.pact.io/getting_started/how_pact_works))

For a modular monolith, the same contract machinery protects an internal port
without paying network cost. If extracted later, the behavioral evidence
already exists.

## State machines, grammars, workflows, and sagas

All four are high-value semantic structures, but not equally core:

| Concept | Candidate placement | Compiler/tool leverage |
|---|---|---|
| State transition | Core ADTs + pure functions | Exhaustiveness and illegal-transition checks |
| Declarative state machine | Official kit/projection first | Diagrams, path generation, reachability, model checking |
| Grammar/parser | Official typed parsing kit | Typed AST, ambiguity/resource diagnostics, fuzz generation |
| Finite pipeline | Ordinary functions + structured tasks/streams | Effect, cancellation, and backpressure graph |
| Durable workflow | Official runtime kit | Checkpoints, replay, idempotency, versioning, placement |
| Saga | Durable workflow policy | Compensation graph, manual intervention, recovery evidence |

```text
machine OrderLifecycle {
  state Pending
  state Paid(payment: PaymentId)
  state Shipped(tracking: TrackingId)
  state Cancelled(reason: CancelReason)

  on Pay(payment:)      Pending -> Paid(payment:)
  on Ship(tracking:)    Paid(_) -> Shipped(tracking:)
  on Cancel(reason:)    Pending -> Cancelled(reason:)
}
```

This should lower to a nominal state ADT, event ADT, and pure transition
function. The tooling can flag unreachable states, unhandled events, missing
authorization, and untested transitions. W3C SCXML shows the breadth of generic
state-machine semantics; that breadth is also a warning against freezing a
large statechart model into the core grammar immediately.
([W3C SCXML](https://www.w3.org/TR/scxml/))

A saga does not create atomicity. Compensation can fail and may require manual
reconciliation, so every compensation has its own idempotency, retry,
observability, and terminal policy. ([Saga pattern](https://learn.microsoft.com/en-us/azure/architecture/patterns/saga))

## GUI direction

GUI should begin as an official `Ui` host/capability with a pure
model-event-update-view center, inspired by the Elm Architecture. Renderers may
target native controls, web, terminal, or a GPU scene, while platform-specific
capabilities remain explicit.

```text
record Model { player: Player, saving: Bool }

data Event = NameEdited(text: Text) | SavePressed | Saved(Result<Player, SaveError>)

fn update(model:, event:) -> Transition<Model, Command> {
  match event {
    NameEdited(text:) => stay(model with { player.name: text })
    SavePressed => command(model with { saving: true }, SavePlayer(model.player))
    Saved(Ok(player:)) => stay(Model(player:, saving: false))
    Saved(Err(error:)) => report(model with { saving: false }, error:)
  }
}

fn view(model:) -> Ui<Node> {
  Column([
    TextField(value: model.player.name, on_change: NameEdited),
    Button(label: "Save", event: SavePressed, disabled: model.saving),
  ])
}
```

The compiler can check event exhaustiveness, state ownership, effect commands,
accessibility labels, and stale async responses. It cannot promise identical
native behavior across platforms. Text input, IME, accessibility, focus,
windowing, GPU, and platform conventions are first-class toolkit work—not a
reason to add GUI keywords to the language.
([Elm Architecture](https://guide.elm-lang.org/architecture/))

## Footguns deliberately excluded

- Transparent local-to-remote substitution.
- A framework-generated microservice per component.
- Global ORM models, hidden lazy queries, and ActiveRecord callbacks.
- Validation that silently coerces at an external trust boundary.
- Treating generated OpenAPI as the complete behavioral contract.
- UUIDs used as authorization or secrecy.
- Event sourcing as the default persistence strategy.
- Sagas presented as rollback-equivalent distributed transactions.
- A state-machine or GUI DSL whose expansion is not inspectable ordinary
  semantic nodes.
