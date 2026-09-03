---
id: performance-observability-delivery
title: Performance, observability, profiling, and continuous delivery
summary: A unified but non-monolithic design for semantic instrumentation, profiling, benchmarks, load and fault tests, SLOs, and evidence-driven delivery.
type: design
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-02
tags: [performance, observability, profiling, delivery]
related: [context-telemetry-security, compiler-feedback-latency, capability-gauntlet, implementation-path, research-ledger]
---

# Performance, observability, profiling, and continuous delivery

## Reader and outcome

This note is for a runtime, compiler, toolkit, or application contributor. After
reading it, they should be able to place a performance or operational feature
at the correct layer and define evidence that catches a regression before or
during a safe rollout.

## Recommendation

Use one typed causal event substrate and stable semantic identities, but keep
the following instruments distinct:

| Instrument | Answers | Typical cost and cadence |
|---|---|---|
| Static explanation | What might allocate, block, suspend, copy, lock, or fan out? | every edit |
| Benchmark | Did a controlled operation become slower or larger? | changed/CI |
| Load test | What happens under a modeled demand distribution? | CI/preproduction |
| Fault test | What happens when dependencies, hosts, clocks, disks, or networks fail? | changed/nightly |
| Monitoring | Is user-visible service health within its objective? | continuous |
| Trace/event inspection | Why did this request, job, actor, or pipeline behave this way? | sampled/focused |
| Profile | Where are CPU, allocation, lock, I/O, scheduler, and memory resources consumed? | sampled/on demand |
| Canary | Does this exact artifact/config/data change behave safely on representative traffic? | delivery |

They correlate through artifact, component, symbol, build, deployment, trace,
and workload IDs. They do not collapse into one complicated observability API.

## Shift-left performance

The compiler can provide useful facts without claiming to solve performance
statically:

- visible allocation, copy, retain/release, boxing, dynamic dispatch, and
  representation conversions;
- blocking, suspension, scheduler hops, lock acquisition, actor sends, remote
  calls, retries, and fan-out;
- loop and collection operations with known complexity classes;
- unbounded queues, recursion, concurrency, payloads, or compile-time work;
- cache-key construction, materialization, spill, and serialization boundaries;
- native size, alignment, padding, vectorization, and target-feature reports;
- an explanation of why a changed public fact invalidated downstream work.

An editor or agent can request the terse view or the complete cost surface:

```text
lang inspect cost OrderApi.checkout
lang inspect allocation --changed --format json
lang inspect blocking --reachable-from Audio.render
lang inspect layout ParticleBatch --target aarch64-macos
```

The compiler should reject objectively incompatible combinations—for example,
blocking or allocating in a hard-realtime effect set. It should report likely
costs elsewhere, not turn heuristics into false correctness errors.

## Performance evidence

```text
benchmark checkout_hot_path {
  workload: CheckoutCases.production_shape(version: "2026-08")
  warmup: until_stable(max: 10.seconds)
  samples: 1000
  budget {
    p50 <= baseline * 1.03
    p99 <= baseline * 1.08
    allocations <= baseline
    peak_memory <= 128.mib
  }
}
```

This is candidate toolkit notation. A performance result must preserve:

- source, compiler, runtime, dependencies, flags, profile, and artifact digest;
- benchmark code/data/config versions;
- host CPU, memory, operating system, power/frequency controls, and isolation;
- warmup and cache state;
- distribution, sample count, uncertainty, outliers, and comparison method;
- CPU, wall time, allocation, I/O, energy, and artifact size where relevant.

One noisy measurement must not block a release. Verification should distinguish
a statistically credible regression, an inconclusive run, and an infrastructure
failure. Absolute budgets protect user requirements; relative budgets catch
drift. Critical benchmarks can run locally through the same hermetic command as
CI.

## Runtime instrumentation

Automatic low-cost events belong at semantically meaningful boundaries:

- capability/effect calls;
- component and foreign boundaries;
- task, actor, workflow, pipeline, and resource lifecycle;
- queues, pools, locks, caches, retries, and circuit state;
- request/job outcome and budget consumption;
- panic, cancellation, overload, migration, and deployment transitions.

Instrumenting every function or line by default creates cost and noise. Pure
function detail is recovered through sampling profiles or a scoped debug lease.
Generated semantic-convention attributes remain versioned adapters rather than
hard-coded source strings.

Cardinality and privacy are checked together. User IDs, raw URLs, SQL values,
cache keys, and model content are not safe metric labels. The compiler can flag
an unbounded label source; the runtime applies count, byte, rate, and retention
budgets and reports dropped evidence without recursively flooding telemetry.

## Monitoring and SLOs

```text
objective checkout_availability {
  good: Checkout.completed within 2.seconds
  valid: Checkout.started excluding InvalidRequest
  target: 99.9.percent over 28.days
  alert: burn_rate(window: 1.hour, factor: 14)
}
```

Objectives and alert policies belong in a first-party operations kit backed by
typed outcome events. The language can validate units, event fields, data
classification, and deployment ownership. It cannot choose the product’s
acceptable reliability target.

Default dashboards can derive latency, traffic, errors, and saturation, plus
domain outcomes. Pages require a user-visible or imminent symptom, an owner,
and an actionable response. Diagnostics and exploratory anomaly detection do
not automatically page a human.

Health, readiness, liveness, and correctness remain distinct. A process may be
alive but unable to serve correctly; a dependency may be degraded while the
user objective still succeeds through a cache or fallback.

## Continuous profiling

The runtime should expose a sampling profiler with stable semantic frames and
build IDs for:

- on-CPU and wall-clock stacks;
- allocation and retained-memory attribution;
- lock wait, contention, and lock-order graphs;
- I/O wait, scheduler delay, task/actor/mailbox time, and backpressure;
- GC or reference-count costs by profile;
- native/foreign frames with symbol and unsafe-boundary attribution.

Profile samples link to traces when possible, but profiling remains useful when
tracing is disabled. Sampling is the continuous default; heavy instrumentation,
heap snapshots, and full event recording require scoped leases and resource
budgets. Exporters should interoperate with pprof, Linux perf, and the evolving
OpenTelemetry profiles signal while the internal representation stays stable.

An AI-facing query should return a regression explanation, not a flame graph
image alone:

```text
lang inspect profile deployment:canary-42 \
  --compare deployment:control \
  --group-by symbol,allocation,lock \
  --format json
```

## Typed load and fault scenarios

```text
load_scenario checkout_burst {
  target: CheckoutApi.place
  traffic: ramp(from: 100.rps, to: 5_000.rps, over: 10.minutes)
  mix: CheckoutTraffic.production_snapshot(version: "2026-08")
  faults: [Payments.latency(p99: 3.seconds), Inventory.loss(rate: 0.01)]
  limits: { clients: 20_000, duration: 30.minutes, cost: 50.usd }
  require {
    successful.p99 <= 750.milliseconds
    error_rate <= 0.1.percent
    queues.inventory <= 5_000
    recovery_time <= 60.seconds
  }
}
```

The harness records the arrival model, concurrency, payload distribution,
connection reuse, warmup, generator saturation, environment, dependency
behavior, and results. It detects when the load generator—not the application—
is the bottleneck. Closed-loop tests must not hide latency by reducing offered
load when the system slows; arrival-rate and coordinated-omission policy are
explicit.

Representative load tests complement canaries. Synthetic traffic cannot reveal
every production input, dependency, cache, or scheduler interaction. Production
targets are denied by default and require an explicit capability and budget.

## Delivery as executable evidence

```text
lang verify changed
lang build --profile service --target linux-aarch64
lang attest artifact:build-73
lang deploy canary --artifact build-73 --policy checkout_rollout
lang promote --deployment canary-42
```

The toolchain derives an impact cone and runs the cheapest sound gates early:

1. format, parse, type/effect/resource, architecture, and policy checks;
2. affected specs, properties, contracts, static cost checks, and fuzz seeds;
3. hermetic multi-target build plus dependency, license, vulnerability, and
   provenance checks;
4. selected integration, compatibility, load, fault, and migration evidence;
5. signed immutable artifact and evidence manifest;
6. canary with black-box outcomes and white-box diagnostics;
7. automatic halt or rollback on declared policy; human escalation only when
   judgment is required;
8. promotion with the same artifact—never a rebuild.

Local and CI commands use the same engine. CI contributes isolated scale,
trusted credentials, target matrices, and signed provenance; it should not be
where developers first discover basic compiler, architecture, or test failures.

## Compiler and runtime self-observation

The compiler must consume its own discipline. Each query reports cache hits,
invalidations, CPU, memory, disk, dependency fan-out, and critical path. A slow
check can answer “why was this slow?” Compile-time extensions have budgets and
cannot perform undeclared network, clock, filesystem, or process effects.

The runtime reports instrumentation overhead and lets verification compare
telemetry-on with telemetry-off behavior. Observability that destabilizes the
application fails its purpose.

## Anti-patterns

- Treating benchmark means or a single run as truth.
- Alerting on every internal anomaly or retry.
- Unbounded metric labels and default capture of sensitive payloads.
- Instrumenting every function while leaving effects and user outcomes opaque.
- Load tests with toy payloads, no arrival model, or an overloaded generator.
- A CI-only command developers cannot reproduce locally.
- Rebuilding between verification and deployment.
- Automatically tuning production without bounded authority and rollback.
- Declaring a performance complexity that the implementation cannot verify.
- Merging monitoring, profiling, tracing, debugging, and load generation into
  one fragile subsystem.

## Primary evidence

- [Google SRE monitoring](https://sre.google/sre-book/monitoring-distributed-systems/)
  distinguishes black-box symptoms and white-box causes, defines the four
  golden signals, emphasizes tail latency, and warns against noisy complexity.
- [OpenTelemetry profiles](https://opentelemetry.io/docs/specs/otel/profiles/)
  defines a low-overhead profile signal correlated with traces, metrics, and
  logs; its status was Alpha when verified.
- [Google SRE canarying](https://sre.google/workbook/canarying-releases/)
  treats a canary as a partial, time-bounded deployment evaluated before
  promotion and explains why artificial load is complementary.
- [SLSA provenance](https://slsa.dev/spec/v1.2/provenance) defines verifiable
  information about where, when, and how an artifact was produced.

All sources were last verified 2026-09-02. Placement and syntax are project
synthesis.
