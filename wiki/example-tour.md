---
id: example-tour
title: Example language tour
summary: A realistic but provisional syntax tour across domain logic, services, tests, supervision, AI inference, interop, and operations.
type: examples
status: seed
confidence: low
created: 2026-09-02
updated: 2026-09-02
tags: [syntax, examples, developer-experience, design-probe]
related: [vision, language-toolchain-primer, effects-and-capabilities, type-system-failure-semantics, syntax-cognitive-ergonomics, feature-coherence, compiler-feedback-latency, application-architecture-data, concurrency-memory, domain-data-distribution, context-telemetry-security, performance-observability-delivery, resources-locks-caching, data-asset-pipelines, modules-architecture-live-development, ai-native-runtime-evals, native-low-level-profile, capability-gauntlet, design-atlas, implementation-path]
---

# Example language tour

## Status

This is **executable-design fiction**: consistent enough to evaluate, not a
committed grammar. It deliberately uses familiar syntax so discussion focuses
on semantics. A Lisp-like or AST-native surface should be benchmarked against
the same examples later.

For a compact explanation of `with`, `Clock.now()`, provider-based DI,
named-argument punning, and the toolchain commands, read the
[language and toolchain primer](language-and-toolchain-primer.md) first.

Working conventions:

- canonical formatter owns layout;
- calls and constructors use named arguments;
- an exact matching local may use `label:` as shorthand for `label: label`;
- public functions have explicit input, output, error, and capability shapes;
- immutable bindings are default;
- `Result` propagation uses `?` but never hides thrown exceptions;
- pipes make data flow visible without positional arguments;
- specs live next to the component contract in the semantic model, even if an
  editor folds them into another view.

## 1. Domain types

```text
module shop.orders.domain

opaque type OrderId = Id<"order">
opaque type Sku = Text matching /^[A-Z0-9-]{3,32}$/

record OrderLine {
  sku: Sku,
  quantity: Int range 1..10_000,
  unit_price: Money<Usd>,
}

data OrderStatus =
  | Pending
  | Confirmed { at: Instant, payment: PaymentId }
  | Cancelled { at: Instant, reason: CancelReason }

record Order {
  id: OrderId,
  lines: NonEmpty<OrderLine>,
  status: OrderStatus,
}
```

The types carry domain distinctions and useful generator boundaries. Regex
syntax is shown as a convenience; whether grammars, regexes, or refinements are
core features remains open.

## 2. A pure domain transition

```text
fn cancel(order: Order, reason: CancelReason) -> Result<Order, CancelError>
requires { reason.is_present }
ensures { result => result.ok?.id == order.id }
{
  match order.status {
    Pending | Confirmed { ... } => Ok(order with {
      status: Cancelled(at: logical_time, reason: reason),
    })

    Cancelled { ... } => Err(AlreadyCancelled(order: order.id))
  }
}
```

This example exposes a flaw: a pure function cannot obtain `logical_time`.
Either the caller supplies `at: Instant`, or the function declares `Clock`.
The compiler should report this as a missing value/capability and offer both
repairs rather than silently reading a global clock.

The more domain-pure version is:

```text
fn cancel(order: Order, reason: CancelReason, at: Instant)
  -> Result<Order, CancelError>
{
  match order.status {
    Pending | Confirmed { ... } => Ok(order with {
      status: Cancelled(at:, reason:),
    })

    Cancelled { ... } => Err(AlreadyCancelled(order: order.id))
  }
}
```

## 3. Ports as capabilities

```text
capability Inventory {
  reserve(items: NonEmpty<OrderLine>)
    -> Result<Reservation, InventoryError>
  release(reservation: Reservation) -> Unit
}

capability Payments {
  charge(amount: Money<Usd>, idempotency_key: OrderId)
    -> Result<PaymentId, PaymentError>
}

capability Orders {
  save(order: Order) -> Result<Unit, StoreError>
}
```

These are application ports. Production adapters may use Postgres and Stripe;
domain/application code cannot name those concrete packages.

## 4. Application use case

```text
fn place_order(command: PlaceOrder)
  -> Result<Order, PlaceOrderError>
  with { Inventory, Payments, Orders, Clock }
{
  let reservation = Inventory.reserve(items: command.lines)?
  defer Inventory.release(reservation:) unless committed

  let payment = Payments.charge(
    amount: command.total,
    idempotency_key: command.id,
  )?

  let order = Order(
    id: command.id,
    lines: command.lines,
    status: Confirmed(at: Clock.now(), payment:),
  )

  Orders.save(order:)?
  committed = true
  Ok(order)
}
```

The example is intentionally uncomfortable: compensation after a successful
charge but failed save is a distributed transaction problem. The language
should reveal the incomplete failure design, not make `defer` look sufficient.
A verifier could emit an obligation such as:

```text
E4107: effect sequence can leave an externally visible partial commit

Payments.charge succeeds -> Orders.save fails -> payment remains captured

repair candidates:
  1. require Payments.authorize + capture after persistence
  2. add compensating Payments.refund and specify its failure policy
  3. execute through a durable workflow with idempotent steps
```

## 5. Flat local specifications

```text
spec "empty orders are unrepresentable" {
  compile_reject PlaceOrder(lines: [])
}

spec "declined payment releases the reservation" {
  let inventory = Inventory.memory(stock: fixtures.in_stock)
  let world = CheckoutWorld(
    inventory:,
    payments: Payments.decline(reason: CardDeclined),
    orders: Orders.memory(),
    clock: Clock.fixed(at: @2026-09-02T12:00:00Z),
  )

  let result = run place_order(command: fixtures.order) with world

  expect result == Err(PaymentFailed(reason: CardDeclined))
  expect inventory.events == [Reserved, Released]
}

property "saved totals equal line totals" {
  for_all command: PlaceOrder.generated() {
    let result = run place_order(command:) with fixtures.successful_order_world

    expect result.ok?.total == command.lines.sum(by: .line_total)
  }
}

cover place_order each PlaceOrderError
```

Each spec stands alone, uses ordinary variables, and has visible arrange, act,
and assert paragraphs. Helpers are ordinary functions rather than nested hooks
or shared-example precedence. `compile_reject` tests the static model.
`property` supplies an authored oracle over generated data. `cover` is an
obligation, not proof that each error behavior is correct.

## 6. Adapter and composition root

```text
adapter PostgresOrders implements Orders
with { Database, Trace }
{
  fn save(order: Order) -> Result<Unit, StoreError> {
    Database.execute(
      statement: sql<"insert_order">,
      parameters: order.to_row(),
    )?
  }
}

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

SQL is a typed resource compiled against a schema snapshot, not an unchecked
string. Secrets are non-printable capability values. The composition graph is
available to the compiler and runtime. Providers are statically resolved; this
is not a runtime map keyed by strings or types. The compiler rejects cycles,
missing bindings, ambiguity, and capture of request providers by service-lived
providers.

## 7. HTTP without framework grammar

```text
route POST "/orders" {
  input: Json<CreateOrderRequest>,
  output: Json<OrderResponse>,
  errors: {
    InvalidInput -> 422,
    OutOfStock -> 409,
    PaymentFailed -> 402,
    DependencyUnavailable -> 503,
  },
  handler: fn(request:) {
    request
      |> CreateOrderRequest.validate(request: _)?
      |> PlaceOrder.from_request(request: _)
      |> place_order(command: _)?
      |> OrderResponse.from_order(order: _)
  },
}
```

`route` should probably be a typed official-library builder over ordinary
language constructs, not permanent HTTP syntax. The formatter may render the
builder declaratively because its schema is known.

## 8. Supervised stateful service

```text
actor InventoryCache receives InventoryMessage {
  state: CacheState,
  mailbox: bounded(capacity: 10_000, overflow: Reject),

  on Get(sku:) -> Reply<Option<Stock>> {
    reply state.stock.get(key: sku)
  }

  on Refresh(snapshot:) {
    state = CacheState.from_snapshot(snapshot: snapshot)
  }
}

supervisor CheckoutWorkers {
  strategy: one_for_one,
  restart: max 3 within 10.seconds,

  child inventory_cache: InventoryCache(
    initial: CacheState.empty(),
    shutdown: graceful(within: 5.seconds),
  )
}
```

The message type is closed and exhaustively handled. Mailbox capacity,
overflow, restart intensity, and shutdown are not production afterthoughts.

## 9. Structured parallel work

```text
fn build_dashboard(account: AccountId)
  -> Result<Dashboard, DashboardError>
  with { Accounts, Billing, Analytics, Task }
{
  within Task.group(deadline: 800.ms) {
    let profile = Task.start {
      Accounts.profile(account: account)
    }

    let balance = Task.start {
      Billing.balance(account: account)
    }

    let activity = Task.start {
      Analytics.recent(account: account, limit: 20)
    }

    Dashboard(
      profile: profile.await()?,
      balance: balance.await()?,
      activity: activity.await()?,
    )
  }
}
```

Leaving the group cancels unfinished children. Trace context and budgets follow
the tasks automatically. There is no detached spawn in the default profile.

## 10. Durable distributed workflow

```text
workflow ResearchAnswer(input: ResearchRequest)
  -> Result<Answer, ResearchError>
  with { Model, Search, Store, Remote }
{
  step evidence = Remote.run(
    placement: requires(capabilities: { Search }),
    retry: exponential(max_attempts: 3, budget: 20.seconds),
    idempotency_key: input.id,
  ) {
    Search.collect(query: input.question)
  }

  step draft = Model.generate(
    model: policy.reasoning_model,
    input: AnswerPrompt(request: input, evidence: evidence),
    output: schema<AnswerDraft>,
    budget: ModelBudget(max_tokens: 4_000, max_cost: 0.20.usd),
    cache: private(scope: input.tenant),
  )

  step verdict = verify_claims(draft: draft, evidence: evidence)
  Store.save(value: verdict)
  Ok(verdict.answer)
}
```

Model inference is an official capability rather than core grammar. The durable
workflow supplies checkpoints, idempotency, retries, placement, cost, and trace
semantics that ordinary async code should not pretend to provide.

## 11. C interoperability

```text
foreign c library zstd {
  header: "zstd.h",
  link: system("zstd", minimum: "1.5"),

  fn compress(
    destination: BorrowedMut<Bytes>,
    source: Borrowed<Bytes>,
    level: CInt,
  ) -> Result<ByteCount, ZstdError>
  contract {
    blocking: true,
    thread_safe: true,
    retains_arguments: false,
    panic_crossing: forbidden,
  }
}
```

Generated bindings should carry ownership and threading contracts. The compiler
cannot prove a C library honors them; the boundary remains marked `Foreign` and
can be isolated in a process or Wasm component.

## 12. Restricted embedded profile

```text
profile embedded {
  runtime: none,
  allocation: region_only,
  panic: abort,
  capabilities: { Gpio, Timer, Serial },
  forbid: { Network, Filesystem, Model, Remote, dynamic_loading },
}

fn sample(sensor: Sensor, using region: Region<1024.bytes>)
  -> Reading
  with { Gpio, Timer }
{
  // Same value and effect semantics; a stricter resource policy.
}
```

The profile restricts programs rather than changing their meaning. Whether one
source semantics can lower efficiently to both BEAM-like and no-runtime targets
is a major open experiment.

## 13. Agent-native tool loop

Human command:

```text
ai check --changed
```

Structured result:

```json
{
  "status": "rejected",
  "change": "chg_7m2",
  "diagnostic_delta": {"added": 1, "removed": 3},
  "diagnostics": [{
    "id": "E4107",
    "symbol": "shop.orders/place_order",
    "kind": "partial_external_commit",
    "effect_path": ["Payments.charge", "Orders.save"],
    "counterexample": {"charge": "ok", "save": "unavailable"},
    "repairs": [
      {"id": "authorize_capture", "confidence": "candidate"},
      {"id": "durable_workflow", "confidence": "candidate"}
    ]
  }],
  "affected_tests": ["place_order/declined_payment_releases_reservation"],
  "cache": {"hit_actions": 37, "executed_actions": 2}
}
```

Candidate protocol operations:

| Operation | Returns |
|---|---|
| `program.symbol` | Canonical signature, docs, contracts, owners, stable ID |
| `program.graph` | Direct dependency/effect/call graph with reasons |
| `change.preview` | Semantic diff, blast radius, required verification |
| `change.apply` | Transactional AST edit with preconditions |
| `verify.changed` | Diagnostic delta and evidence bundle |
| `runtime.tree` | Nodes, supervisors, actors, tasks, mailboxes, budgets |
| `runtime.trace` | Causal typed events scoped to a request or symbol |
| `runtime.explain_failure` | Minimal causal packet and relevant source contracts |
| `runtime.replay` | Deterministic handler inputs where capture policy permits |
| `cost.explain` | Build/runtime/inference cost attribution by component |

LSP, DAP, and BSP demonstrate the value of versioned structured boundaries for
editors, debuggers, and builds. Rust's compiler also emits hierarchical JSON
diagnostics with applicability metadata. This language should offer those human
integrations while designing a lower-noise, semantically richer agent protocol.
([LSP](https://microsoft.github.io/language-server-protocol/specifications/specification-current),
[DAP](https://microsoft.github.io/debug-adapter-protocol/),
[BSP](https://build-server-protocol.github.io/docs/specification),
[rustc JSON](https://doc.rust-lang.org/nightly/rustc/json.html))

## 14. Typed context and automatic telemetry

```text
fn create_order(request: HttpRequest)
  -> Result<HttpResponse, RequestError>
  using { execution: RequestExecution, auth: AuthenticatedRequest }
{
  let command = PlaceOrder(
    tenant: auth.tenant.value,
    actor: auth.principal.value,
    lines: CreateOrderRequest.decode(body: request.body)?.lines,
  )

  observe current_operation {
    field order_kind = command.kind
      @classification(operational)
      @cardinality(low)
  }

  place_order(command: command)
    |> OrderResponse.from_result(result: _)
}
```

The runtime supplies trace identity, deadline, cancellation, operation name,
duration, effect calls, and typed outcome. The program supplies business meaning.
Tenant and principal arrive through verified edge context but become explicit
command fields because authorization and behavior depend on them. A secret or
unclassified high-cardinality field is rejected from telemetry rather than
redacted after an incident.

## 15. Domain events, consistency, and deterministic failure

```text
event OrderEvent version 1 {
  OrderPlaced { order_id: OrderId, total: Money<Usd> }
  InventoryReleased { order_id: OrderId, reservation: ReservationId }
}

capability Orders.Read {
  get(id: OrderId, consistency: ReadConsistency)
    -> Result<Option<Observed<Order>>, ReadError>
}

capability Orders.Write {
  append(
    expected: Version,
    events: NonEmpty<OrderEvent>,
    idempotency_key: IdempotencyKey,
  ) -> Result<CommitReceipt, WriteError>
}

simulation inventory_partition {
  seed: 0x82f4
  world: ServiceWorld.deterministic()

  fault Network.partition(between: [checkout, inventory], at: 3.seconds)

  always Inventory.reserved_not_oversold
  eventually Orders.terminal_or_compensated(within: 30.seconds)
}
```

`event` and `simulation` are candidate projections from an official kit, not
accepted core keywords. Their types, effects, schemas, faults, and obligations
must remain ordinary queryable semantic nodes. `Orders.Read` and `Orders.Write`
make authority separable without forcing every application to use CQRS. The
simulation works because clock, network, scheduling, storage, and randomness
are effects; exact replay is promised only when every relevant source of
nondeterminism is captured.

## 16. Opaque identity with a UUIDv7 service provider

```text
opaque type PlayerId = Id<Player>

fn register_player(input: NewPlayer)
  -> Result<Player, RegisterError>
  with { Ids<Player>, Clock, Players.Write }
{
  let player = Player(
    id: PlayerId.new(),
    joined_at: Clock.now(),
    name: input.name,
  )

  Players.Write.insert(player:)?
  Ok(player)
}

service GameServer {
  provide Ids<Player> = Uuid.v7(clock: Clock, random: Random)
}
```

Domain code cannot mix `PlayerId` with `MatchId` or inspect UUID bits. A test
can provide deterministic IDs. A database adapter stores UUIDv7 as 128-bit
binary where supported. A native array or emulator may use a different opaque
index type; UUID is a profile default, not a core representation.

## 17. Verified modular monolith

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

The compiler rejects cycles and access to component internals. The deployment
has one primary transactional boundary, but tables still have owners. Running
`architecture extract Orders` returns the serialization, compatibility,
security, idempotency, and rollout obligations that would appear if it became a
service; it never silently turns a local call into a remote one.

## 18. Strict JSON to validated domain data

```text
record CreatePlayerRequest derives { Json.Schema } {
  name: Text,
  email: Text,
}

fn decode_create_player(body: Bytes) -> Validation<CreatePlayer> {
  let request = Json.decode<CreatePlayerRequest>(
    bytes: body,
    mode: strict,
    unknown_fields: reject,
    duplicate_fields: reject,
    limits: JsonLimits.default,
  )?

  validate CreatePlayer(
    name: PlayerName.parse(text: request.name),
    email: EmailAddress.parse(text: request.email),
  )
}
```

Decoding yields path-structured errors and never makes raw text a validated
email/name by assertion. JSON Schema and OpenAPI are versioned derived
artifacts, not alternate sources of truth.

## 19. Explicit change data and typed SQL

```text
fn change_profile(player:, input:) -> Change<Player> {
  change(value: player)
    |> cast(input:, allow: [.display_name, .bio])
    |> validate(.display_name, using: PlayerName.validate)
    |> validate(.bio, using: Text.max_graphemes(count: 280))
}

query find_player(id: PlayerId) -> Option<PlayerRow> using schema GameDb {
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
    let change = change_profile(player:, input:).require_valid()?
    repo.update(change:)?
  })
}
```

Records do not save themselves. There is no ambient repository or lazy query.
The SQL resource is compiled against a schema snapshot and reports parameters,
nullability, result shape, cardinality assumptions, and migration drift.

## 20. Consumer contract verification

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

The consumer publishes its minimal dependency; provider verification replays
the interaction. The same mechanism protects an in-process port today and a
remote service tomorrow. It complements schema, domain, integration, and E2E
evidence.

## 21. Expected failure, defect, and supervision

```text
fn settle(invoice: Invoice)
  -> Result<Receipt, SettleError>
  with { Payments, Clock }
{
  match Payments.capture(invoice:) {
    Ok(payment:) => Ok(Receipt(payment:, at: Clock.now()))
    Err(Declined(reason:)) => Err(PaymentDeclined(reason:))
    Err(Unavailable(retry_after:)) => Err(DependencyUnavailable(retry_after:))
  }
}

actor SettlementWorker receives SettlementMessage {
  on Settle(invoice:) {
    match settle(invoice:) {
      Ok(receipt:) => reply Settled(receipt:)
      Err(error: Retryable) => retry(error:, using: settlement_retry)
      Err(error:) => reply Rejected(error:)
    }
  }
}
```

Expected failures are handled, translated, or propagated. Adding a payment
error variant breaks the exhaustive match. An impossible invariant panics the
worker, produces a causal failure packet, and invokes supervisor policy; it
cannot be caught and ignored by ordinary application code.

## 22. State machine as an inspectable projection

```text
machine OrderLifecycle {
  state Pending
  state Paid(payment: PaymentId)
  state Shipped(tracking: TrackingId)
  state Cancelled(reason: CancelReason)

  on Pay(payment:)   Pending -> Paid(payment:)
  on Ship(tracking:) Paid(_) -> Shipped(tracking:)
  on Cancel(reason:) Pending -> Cancelled(reason:)
}
```

The kit lowers this to state/event ADTs and a pure transition function. It can
generate a diagram, transition coverage, boundary cases, and bounded traces.
Whether `machine` is dedicated grammar or an editor rendering of ordinary
typed declarations remains an experiment.

## 23. Cross-platform UI with explicit commands

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

`update` is pure; commands declare effects separately. Native and web hosts
share the semantic center but expose platform capabilities for accessibility,
text input, windows, menus, and rendering rather than promising false
uniformity.

## 24. Bounded AI workflow and evaluation

```text
workflow ClassifyIncident(incident: Incident)
  -> Result<Route, TriageError>
  with { TriageModel, ReadRunbook, HumanReview }
{
  let evidence = ReadRunbook.call(
    input: RunbookQuery(service: incident.service),
  )?

  let reply = TriageModel.infer(
    input: TriagePrompt(incident:, evidence:),
    policy: Policy(
      deadline: 3.seconds,
      max_tokens: 800,
      max_cost: 0.02.usd,
      data_class: Internal,
    ),
  )?

  match validate_route(reply.value) {
    Ok(route:) if route.confidence >= 0.90 => Ok(route)
    _ => HumanReview.route(subject: incident, evidence: reply.trace_ref)
  }
}

eval incident_triage for ClassifyIncident {
  dataset: IncidentCases(version: "2026-09-01")
  samples_per_case: 5
  require {
    unsafe_auto_routes == 0
    accuracy.lower_confidence_bound >= 0.95
    p95_cost <= 0.02.usd
  }
}
```

The model is an effect, not an oracle. The workflow has explicit authority and
budgets; low-confidence output routes to a human. The eval is a versioned
evidence artifact with repeated samples, not a deterministic unit assertion.
An editor may draw the agent graph from this ordinary typed control flow.

## 25. Native parsing with caller-owned memory

```text
fn decode_frame(
  bytes: BytesView,
  using arena: borrow mut Arena,
) -> Result<borrow Frame, FrameError | OutOfMemory>
{
  let header = FrameHeader.parse(bytes: bytes.take(count: 8))?
  let payload = bytes.slice(
    from: 8,
    count: checked_usize(header.payload_length)?,
  )?

  let frame = arena.allocate<Frame>()?
  frame.initialize(header:, payload:)
  Ok(frame.borrow())
}
```

The result cannot outlive `arena`. Bounds, integer conversion, and allocation
stay checked in release builds. Exact borrow notation and inference boundaries
remain experimental, but the lifetime, OOM, and no-hidden-copy requirements do
not.

## 26. Lexical resources and honest cleanup failure

```text
fn copy_file(from: Path, to: Path)
  -> Result<ByteCount, CopyError>
  with { Files }
{
  use input = Files.open_read(path: from)?
  use output = Files.create_write(path: to)?

  let copied = Streams.copy(
    from: input,
    to: output,
    buffer_bytes: 64.kib,
  )?

  output.commit()?
  Ok(copied)
}
```

Ownership prevents either handle from escaping or being used after close. The
scope guarantees cleanup on return, error, panic, and cancellation. `commit`
is explicit because a write becoming durable can fail; automatic cleanup must
not silently convert that failure into success.

## 27. Typed cache policy rather than invisible memoization

```text
cache ProductDescriptions {
  key: ProductId
  value: ProductDescription
  capacity: 128.mib
  freshness: 5.minutes
  stale_while_revalidate: 30.seconds
  negative: 10.seconds for ProductMissing
  coalesce: in_flight
}

fn description(product_id:)
  -> Result<ProductDescription, CatalogError>
  with { ProductDescriptions, Catalog }
{
  ProductDescriptions.get_or_load(
    key: product_id,
    load: fn() { Catalog.fetch_description(product_id:) },
  )?
}
```

The cache declares key, clock-dependent freshness, capacity, negative behavior,
and duplicate-request suppression. The provider owns storage and eviction. The
runtime reports hit state, age, origin load, coalescing, evictions, and tenant
classification without making cache use ambient.

## 28. A batch/stream pipeline with lineage and event time

```text
fn clean_order(raw:) -> Validation<Order> {
  validate Order(
    id: OrderId.parse(text: raw.order_id),
    amount: Money.parse(text: raw.amount, currency: raw.currency),
    occurred_at: Instant.parse(text: raw.occurred_at),
  )
}

pipeline DailyOrders with { OrdersSource, Warehouse, Quarantine } {
  OrdersSource.read(schema: RawOrder.schema, mode: unbounded)
    |> event_time(using: .occurred_at)
    |> watermark(max_lateness: 10.minutes)
    |> map_validated(using: clean_order, invalid: Quarantine.write)
    |> window(tumbling: 1.day)
    |> aggregate(using: summarize_orders)
    |> Warehouse.upsert(
         table: "daily_orders",
         idempotency: [.window, .merchant_id],
       )
}
```

The graph carries source/dataset identity, schemas, lineage, watermarks,
quality results, and sink guarantees. A local runner and distributed runner may
execute it differently. Neither may erase the distinction between event time
and processing time or call an outcome exactly-once without the necessary
source and sink contracts.

## 29. Performance and delivery evidence

```text
benchmark "decode one million orders" for Json.decode_stream<RawOrder> {
  input: fixtures.orders_1m
  require {
    p95 <= 850.milliseconds
    peak_memory <= 96.mib
    allocations_per_row <= 2
  }
}

load "checkout saturation" against CheckoutService {
  traffic: ramp(from: 100.rps, to: 5_000.rps, over: 10.minutes)
  faults: [Inventory.delay(p99: 800.milliseconds)]
  require {
    errors.rate <= 0.1.percent
    latency.p99 <= 500.milliseconds
    queues.checkout <= 2_000
  }
}

release production {
  require: [verify.all, benchmark.changed, load.checkout_saturation]
  canary: Percent(value: 2, for: 20.minutes)
  promote_when: CheckoutSlo.healthy
  rollback_when: CheckoutSlo.burn_rate > 2
}
```

These are versioned evidence artifacts, not core control-flow syntax. The
compiler binds them to stable symbols, inputs, artifact/config IDs, budgets,
and provenance. Static cost explanations, benchmarks, load/fault tests,
monitoring, traces, profiles, and canaries remain distinct views.

## 30. Explicit modules and a typed live workspace

```text
module shop.orders

export {
  type PlaceOrder
  type PlaceOrderError
  fn place_order
}

import shop.inventory.{Inventory}
import shop.payments.{Payments}
```

```text
lang live
> use providers test
> inspect effects shop.orders.place_order
> run place_order(command: fixtures.valid_order)
> promote last as spec "valid order reserves and captures"
```

Everything is private unless exported; imports name their dependency and the
compiler rejects module/component cycles. The live workspace executes the same
typed language, records providers and effects, and can promote a discovery into
ordinary durable source. It does not pretend a real payment or file write can
be undone.

## What the examples reveal

1. Effects and named data make important behavior legible without much syntax.
2. Distributed compensation cannot be solved by prettier error propagation.
3. Architecture policy and effect inspection may prevent more slop than a novel
   grammar.
4. Tests can be local and generated, but useful oracles remain authored intent.
5. Actor and task models serve different lifetimes and should coexist.
6. AI inference benefits from the same capabilities, budgets, tests, and traces
   as any other nondeterministic external system.
7. The agent protocol may be as differentiating as the source language.
8. DI becomes smaller when effect/capability contracts and provider lifetimes
   are part of the semantic model.
9. Automatic telemetry is credible only when propagation, privacy, cardinality,
   overhead, and trust are checked together.
10. Exact-name argument punning removes repetition without losing parameter-role
    labels or returning to positional arguments.
11. Domain events and resilience patterns can be turnkey declarative kits while
    their semantic ingredients remain a small general-purpose core.
12. A modular monolith can have compiler-enforced boundaries without paying the
    operational and semantic cost of a network.
13. UUIDv7, JSON, SQL, contracts, state machines, and GUI all fit without
    becoming universal grammar when official kits emit ordinary semantic nodes.
14. Flat specs preserve locality and readability better than a nested test DSL.
15. AI-specific workflows fit ordinary effects, tools, budgets, state, and
    evidence; graph syntax need not enter the semantic core.
16. Low-level credibility requires explicit lifetime, layout, allocation, and
    unsafe contracts even if ordinary service code infers most of them.
17. Resource safety requires both automatic scope cleanup and explicit handling
    of fallible commit/flush outcomes.
18. Caching is a visible consistency and capacity policy, not an annotation that
    silently memoizes effects.
19. Data and asset pipelines reuse types, effects, resource bounds, provenance,
    and evidence while making event-time and delivery limits explicit.
20. Performance and delivery tools compound when correlated by stable semantic
    identity, but remain more honest as separate instruments.
21. A typed live workspace can provide Lisp-like immediacy without a second
    language or unrecorded session truth.
