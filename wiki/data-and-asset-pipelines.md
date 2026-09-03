---
id: data-asset-pipelines
title: Data, warehouse, and asset pipelines
summary: A typed pipeline design for large data, cleaning, lineage, event time, backfills, warehouses, columnar execution, and reproducible asset builds.
type: design
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-02
tags: [data, pipelines, assets, lineage]
related: [domain-data-distribution, type-system-failure-semantics, performance-observability-delivery, capability-gauntlet, research-ledger]
---

# Data, warehouse, and asset pipelines

## Reader and outcome

This note is for someone implementing a data/asset toolkit, connector, or
pipeline. After reading it, they should be able to define a typed, replayable,
observable pipeline without inventing another execution or metadata model.

## Recommendation

Treat a pipeline as a versioned graph of ordinary typed functions and explicit
effects. The compiler understands schemas, dependencies, resource bounds,
lineage, time semantics, partitioning, failure, and evidence. Runners decide
local versus distributed execution. Warehouses, object stores, queues, media
tools, and GPUs are capability providers.

Do not make one universal `pipeline` grammar part of the core yet. A first-party
kit can derive an inspectable graph from normal modules/functions and render a
concise pipeline view. The same substrate should support:

- bounded batch and unbounded stream data;
- ETL/ELT and warehouse models;
- data cleaning, validation, enrichment, and quarantine;
- database migration/backfill jobs;
- ML feature and evaluation datasets;
- source, media, shader, localization, and game-asset builds.

## Candidate data surface

```text
record RawOrder {
  order_id: Text,
  occurred_at: Text,
  amount: Text,
  currency: Text,
}

record CleanOrder {
  order_id: OrderId,
  occurred_at: Instant,
  amount: Money<Currency>,
}

fn clean_order(raw: RawOrder)
  -> Result<CleanOrder, DataIssue>
{
  Ok(CleanOrder(
    order_id: OrderId.parse(raw.order_id)?,
    occurred_at: Instant.parse(raw.occurred_at, policy: strict_iso8601)?,
    amount: Money.parse(amount: raw.amount, currency: raw.currency)?,
  ))
}
```

Missing, null, invalid, redacted, late, deleted, and unknown are not one value.
Boundary decoding makes each policy explicit. Cleaning does not silently coerce
bad values or overwrite the raw evidence.

```text
pipeline daily_orders(run: DataRun)
  with { RawLake, Warehouse, Quarantine, Lineage }
{
  RawLake.read<RawOrder>(partition: run.partition)
    |> map_result(clean_order)
    |> route_errors(to: Quarantine.dataset("bad_orders"))
    |> assert(order_quality)
    |> Warehouse.merge<CleanOrder>(
         table: "fact_orders",
         key: .order_id,
         run_id: run.id,
       )
}
```

This notation may be an editor projection. Semantically it is typed dataflow,
capability calls, validation evidence, and a declared sink contract.

## Dataset identity and lineage

A dataset reference includes more than a table name:

```text
DatasetRef<CleanOrder>(
  namespace: Warehouse.analytics,
  name: "fact_orders",
  version: SchemaVersion(7),
  partition: Date(2026-09-02),
  snapshot: SourceDigest(...),
  classification: Confidential,
)
```

Each run records:

- job/pipeline semantic ID and source/artifact version;
- exact input and output dataset identities, partitions, and snapshots;
- field-level lineage when derivable;
- providers, configuration, code, schemas, and transformation fingerprints;
- row/byte counts, rejected/quarantined records, distributions, and quality
  evidence;
- start, completion, failure, retry, owner, budget, and causal trace;
- declared versus observed lineage, with disagreement reported.

The internal lineage model should export through OpenLineage rather than invent
an isolated ecosystem. Runtime lineage is not authority to read a dataset;
capability and classification policy remain separate.

## Data contracts and quality

```text
data_contract order_quality for Dataset<CleanOrder> {
  require {
    row_count >= previous_success * 0.90
    unique(.order_id) == 100.percent
    not_null(.occurred_at) == 100.percent
    sum(.amount) approximately source_control_total tolerance 0.01.percent
    distribution(.currency).drift <= 0.05
  }

  on_violation {
    Critical => stop_before_publish
    Warning => publish_with_evidence
  }
}
```

Types catch row-shape and local validity failures. Dataset contracts catch
population, relationship, freshness, drift, reconciliation, and business
invariants. Both are needed. Generated checks can propose constraints from
profiles, but only an owner can decide which distribution changes are wrong.

Quality gates preserve representative failing samples under classification and
size limits. “Mostly valid” requires an explicit threshold and disposition; it
must not become silent data loss.

## Batch, stream, and time

The same pure transform may run in batch or stream mode. Execution semantics
remain explicit:

```text
stream sessions from ClickEvents {
  time: event_time(.occurred_at)
  watermark: bounded_out_of_order(max_lateness: 5.minutes)
  window: session(gap: 30.minutes)
  late: route(to: LateEvents)
  trigger: on_watermark with_updates until 24.hours
}
```

Processing time, ingestion time, and event time are distinct types. Watermarks
are estimates, not facts that no earlier event can arrive. Windowing, triggers,
allowed lateness, retractions/updates, ordering, and late-data disposition are
visible in the semantic graph and tests.

Backpressure and checkpoint cost use the same bounded-stream mechanisms as
services. Data skew, hot keys, slow partitions, state growth, and checkpoint
alignment appear in runtime events and profiles.

## Delivery guarantees without mythology

`at_most_once`, `at_least_once`, and `exactly_once_state` are boundary
contracts—not decorative labels. End-to-end exactly-once effects require a
replayable source plus a transactional or idempotent sink that participates in
the protocol. A runner cannot manufacture those properties for an arbitrary
email, HTTP API, or foreign database.

The compiler verifies that a chosen retry/checkpoint policy is compatible with
declared source and sink traits. Runtime evidence records offsets, checkpoints,
commit IDs, duplicate suppression, and reconciliation. When end-to-end
exactly-once is unavailable, make duplicates or loss explicit and supply
idempotency, deduplication, or compensation.

## Backfills and migrations

A backfill is a versioned deployment over historical data, not an ad hoc loop:

```text
backfill orders_v7 {
  code: artifact("orders-pipeline", version: "7.2.1")
  inputs: partitions(2024-01-01..2026-08-31, snapshot: immutable)
  output: shadow("fact_orders_v7")
  parallelism: 64
  budget: { warehouse: 20_000.slot_hours, deadline: 12.hours }
  verify: [order_quality, compare_to_v6, downstream_contracts]
  promote: atomic_alias_swap
}
```

Defaults should favor shadow output, resumable partition checkpoints,
idempotent writes, limited concurrency, dependency budgets, and validation
before promotion. The tool estimates affected partitions, bytes, cost, and
downstream lineage before execution. Production mutation requires a separate
capability and supports pause, resume, rollback, and audit.

Schema evolution is directional. Readers and writers may be compatible in one
direction but not another. Adding a field, changing nullability, reinterpreting
units, changing partition keys, or altering event-time policy selects affected
pipelines, contracts, backfills, caches, and consumers.

## Large-data execution and mechanical sympathy

The portable collection API should distinguish materialized collections,
lazy/streaming sequences, columnar batches, and external datasets. An operation
that collects an unbounded stream or multi-terabyte dataset into memory should
not look like an ordinary list conversion.

```text
Dataset<Order>
  |> filter(.occurred_at >= cutoff)
  |> group_by(.customer_id)
  |> aggregate(total: sum(.amount))
  |> execute(using: Warehouse, budget: query_budget)
```

The planner exposes pushdown, projection, join order, repartition, shuffle,
sort, materialization, cache, spill, and estimated row/byte cost. Execution has
bounded memory and disk spill rather than accidental OOM. Statistics are
versioned observations and bad estimates remain visible.

Apache Arrow is a strong interop and in-memory columnar target: contiguous
columns aid scanning/SIMD and its format supports zero-copy exchange. It is not
the universal source representation; mutation-heavy domain code and tiny
records may prefer row layouts. Untrusted columnar buffers require structural
and full validation plus allocation/decompression limits.

## Asset pipelines

Asset builds share the data-pipeline fundamentals but add content-addressed
identity and target variants:

```text
asset_pipeline ship_texture(source: Asset<Image>)
  with { ImageCodec, GpuCompressor, AssetStore }
{
  source
    |> decode(limit: 200.megapixels)
    |> color_convert(to: Srgb)
    |> mipmaps(filter: Lanczos3)
    |> compress(format: Bc7, quality: 0.8)
    |> publish(target: DesktopGpu)
}
```

Every stage declares tool/version, target, platform, inputs, outputs, resource
budgets, determinism, and licenses. Hermetic transforms and content hashes allow
incremental rebuild and cache sharing. Nondeterministic encoders record seeds
and remain non-reproducible unless demonstrated otherwise. External processes
receive sandboxed filesystem/process capabilities, not ambient machine access.

Shaders, CSS, localization, generated code, media, and models should all use
this substrate instead of bespoke build scripts. The final artifact manifest
links assets to their sources and transforms.

## Interactive data work

The semantic workspace/REPL can inspect samples, schemas, plans, and quality
results without loading the complete dataset. Every effectful evaluation has a
provider, budget, dataset snapshot, and recorded result reference. A useful
exploration can be promoted into a module, spec, data contract, or pipeline
without copy/pasting hidden notebook state.

## Operations and observability

Pipeline defaults expose:

- input/output rows and bytes, freshness, lag, watermark, and late data;
- throughput, backlog, skew, hot partitions, shuffle/spill, memory, and cost;
- quality failures, quarantine growth, duplicate/loss estimates, and drift;
- checkpoint duration/failure, recovery point, retry amplification, and sink
  commit state;
- dataset/job/run lineage and affected downstream consumers.

Alerts attach to user/business freshness and correctness objectives, not every
failed row. A quarantine that silently grows is still a correctness incident.

## Anti-patterns

- Treating `null`, malformed, absent, redacted, and late data as one state.
- Silent coercion or dropping bad records to make a dashboard green.
- Claiming end-to-end exactly-once because operator state is checkpointed.
- Running backfills directly into current production tables by default.
- Unbounded `collect`, shuffle, state, decompression, or external process use.
- Hidden notebook state and irreproducible manual cleaning.
- A pipeline DAG that is separate from the compiler dependency/effect graph.
- Treating field lineage as authorization.
- Making columnar layout universal rather than profile- and operation-aware.
- Caching derived data without code/schema/input fingerprints.

## Primary evidence

- [Apache Beam model](https://beam.apache.org/documentation/basics/) distinguishes
  bounded/unbounded data, event and processing time, watermarks, windows,
  triggers, state, and timers.
- [Apache Flink fault tolerance](https://nightlies.apache.org/flink/flink-docs-stable/docs/learn-flink/fault_tolerance/)
  explains that exactly-once operator state is not the same as each event being
  processed once and that end-to-end guarantees require replayable sources and
  transactional or idempotent sinks.
- [OpenLineage](https://openlineage.io/docs/) defines interoperable dataset,
  job, and run lineage events with extensible facets.
- [Apache Arrow columnar format](https://arrow.apache.org/docs/format/Columnar.html)
  specifies language-independent columnar buffers designed for adjacency,
  random access, SIMD, and zero-copy sharing.
- [Apache Arrow security considerations](https://arrow.apache.org/docs/dev/format/Security.html)
  require explicit validation and bounded allocation for untrusted data.
- [Great Expectations](https://docs.greatexpectations.io/docs/core/define_expectations/create_an_expectation/)
  demonstrates data expectations as explicit, executable assertions, including
  severity and partial-success policies.

All sources were last verified 2026-09-02. The unified language design is
project synthesis.
