---
id: capability-gauntlet
title: General-purpose capability gauntlet
summary: A workload and quality grid that turns general-purpose ambition into staged, falsifiable evidence.
type: strategy
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-03
tags: [evaluation, workloads, roadmap, convergence]
related: [design-baseline, author-intent-model, implementation-path, trust-validation-information-flow, language-ecosystem-lifecycle-backcast, residual-uncertainty-register, native-low-level-profile, runtime-profiles-dogfooding, memory-reclamation-policy, networking-http-tls-security, compiler-feedback-latency, compute-efficiency-constitution, performance-observability-delivery, resources-locks-caching, data-asset-pipelines, modules-architecture-live-development, ai-native-runtime-evals, convergence-audit, convergence-work-program, open-questions]
---

# General-purpose capability gauntlet

## Why this exists

“Can do everything” is not a useful acceptance criterion. This gauntlet names
representative software classes, the semantic stress each applies, current
readiness, and the proof that would move the claim forward. It prevents a good
service language from being mistaken for a demonstrated systems language and
prevents long-term ambition from distorting the first shippable slice.

## Evidence levels

| Level | Meaning |
|---:|---|
| G0 | Incompatible or intentionally out of scope. |
| G1 | Requirements and semantic risks are recorded. |
| G2 | A coherent design maps the workload to core/profile/toolkit mechanisms. |
| G3 | A representative vertical prototype works with tests, failures, inspection, and benchmarks. |
| G4 | Comparative non-toy implementation meets explicit acceptance thresholds. |
| G5 | Production pilot survives operation, upgrade, incident, and user feedback. |

Current levels describe **design evidence through 2026-09-03**, not an implemented
language. No class is above G2 yet.

## Workload matrix

| Software class | Primary stress | Likely profile/kits | Now | Decisive next proof |
|---|---|---|---:|---|
| Pure algorithm/data-structure library | generics, ADTs, iteration, allocation, predictability | portable core | G2 | property-tested collection/parser corpus with competitive compile/runtime cost |
| CLI and automation | startup, files, processes, signals, packaging | native/service CLI kit | G2 | single-binary tool with shell completion, errors, cancellation, upgrade |
| JSON parser/validator | bytes, Unicode, recursion, limits, schemas, fuzzing | portable/native JSON kit | G2 | standards corpus, differential/fuzz tests, streaming and adversarial limits |
| HTTP web service | async I/O, routing, validation, cancellation, telemetry | service + HTTP/JSON kit | G2 | production-shaped modular monolith with SQL and failure injection |
| HTTP server implementation | sockets, buffers, parsing, TLS, identity, intermediaries, backpressure | native + network kit | G2 | interoperable HTTP/1.1 server through a real proxy, fuzzed/differential parser, TLS failure, replay, SSRF, and slow-client tests |
| Custom network protocol | framing, endianness, timeouts, compatibility | native/service protocol kit | G2 | versioned binary protocol with partial reads and fault simulation |
| Stream processor | bounded flow, checkpoints, ordering, time | service + streams | G2 | burst/late-data workload with recovery and backpressure proof |
| Batch ETL/data cleaning | schemas, invalid/missing data, quarantine, lineage, backfill | native/service + data kit | G2 | stream a large dirty dataset through typed validation with bounded memory and reproducible lineage |
| Warehouse/analytics job | columnar layout, partitioning, pushdown, evolution, remote compute | native/service + Arrow/warehouse adapters | G2 | compile one typed model to local and warehouse execution and compare results/cost |
| Asset/build pipeline | content addressing, external tools, incremental graphs, reproducibility | native + asset/build kit | G2 | import/process/package media or game assets with hermetic cache and provenance |
| Load/fault test harness | workload models, virtual time, saturation, distributed coordination | native/service performance kit | G2 | reproduce overload and dependency failure with bounded generators and causal evidence |
| Profiler/observability collector | low overhead, sampling, unwinding, correlation, privacy | native runtime + operations kit | G2 | correlate CPU/allocation/lock profiles with traces and stable symbols under measured budget |
| Distributed actor service | supervision, mailboxes, partitions, placement | service actor runtime | G2 | multiplayer/session workload under node loss and rolling upgrade |
| Durable workflow/saga | replay, versioning, idempotency, compensation | service workflow kit | G2 | multi-day simulated workflow across code versions and failed compensation |
| AI agent/tool workflow | stochastic calls, authority, budgets, traces, evals | service + AI/eval kits | G2 | bounded agent with adversarial evals, approvals, replay, cost policy |
| Database client/ORM-like layer | protocols, pooling, schemas, transactions | service/native SQL/change kit | G2 | typed Postgres slice with migrations, cancellation, contract tests |
| Embedded KV/storage engine | pages, checksums, crash recovery, mmap/direct I/O | native storage kit | G1 | crash-consistent B-tree/LSM slice with model/fault tests |
| Relational database engine | query planning, storage, concurrency control, recovery | native + compiler/storage kits | G1 | small SQL engine with transactions and crash/recovery corpus |
| Compiler/interpreter | trees, graphs, incremental queries, diagnostics, bootstrap | native compiler kit | G2 | self-hosting subset or second implementation with conformance suite |
| Plugin host | dynamic loading, ABI/versioning, sandbox, authority | native/Wasm component kit | G1 | unload/upgrade-safe Wasm and native plugin comparison |
| Browser/Wasm application | compact artifacts, host effects, JS/DOM boundary | Wasm + web UI kit | G1 | non-toy offline developer tool with accessibility and profiling |
| Native desktop GUI | event loop, rendering, accessibility, platform conventions | native + UI host kit | G1 | same developer tool on macOS/Linux with explicit platform escapes |
| Mobile application | lifecycle, permissions, energy, native SDKs | mobile host profile | G1 | background/foreground app with platform API bindings and upgrade |
| Game logic/server | deterministic ticks, latency, networking, hot state | native/service game kits | G2 | rollback-networked simulation plus authoritative server |
| Game engine | ECS/data layout, assets, scripting, jobs, tools | native + GPU/UI/audio | G1 | small 2D engine with editor, asset reload, profiler, packaged game |
| Renderer/GPU compute | SIMD, layout, GPU memory, shader/driver interop | native GPU kit | G1 | triangle-to-scene renderer with frame-time and validation evidence |
| Audio/DSP | realtime bounds, SIMD, no allocation/blocking | realtime native profile | G1 | glitch-free graph under load with statically checked realtime effects |
| Emulator/VM | bit precision, dispatch, memory maps, timing, JIT option | native | G1 | cycle-aware small-console emulator with conformance ROMs |
| OS userland utility | syscalls, signals, permissions, zero-runtime option | native | G1 | core utility without libc plus sanitizer/conformance runs |
| Kernel | no runtime, address spaces, interrupts, atomics, linker/startup | freestanding native | G1 | bootable minimal kernel with allocator, interrupt, driver, panic path |
| Device driver | MMIO, DMA, interrupts, pinning, unsafe containment | embedded/freestanding | G1 | real or emulated driver with bounded unsafe contract and fault tests |
| Embedded firmware | static memory, cross-build, power, peripherals, no OS | embedded | G1 | microcontroller sensor/transport firmware with no heap and size budget |
| Hard realtime control | WCET, bounded allocation, interrupts, priority inversion | realtime embedded | G1 | deadline proof/measurement with forbidden-effect enforcement |
| HPC/numerics | arrays, SIMD, parallelism, accelerators, numerical semantics | native/HPC kit | G1 | representative kernel against C/Fortran/Rust baselines |
| Cryptographic library | constant time, zeroization, side channels, auditability | restricted native crypto capability, not the official TLS implementation path | G1 | known-answer tests, side-channel tooling, independent review |
| Local AI inference | tensors, accelerators, batching, memory, model formats | native AI provider | G1 | small supported model through external provider, then measured runtime case |
| ML training/data science | dynamic exploration, tensor autodiff, GPU/distribution | likely Python interop first | G1 | notebook/REPL and accelerator interop study before native ambition |
| Serverless/edge function | cold start, package size, quotas, observability | native/Wasm/service | G1 | deploy same typed handler to Wasm edge and container target |

## Every prototype runs the same quality gauntlet

A workload is not proven merely because it produces the expected happy-path
output.

| Dimension | Questions the evidence must answer |
|---|---|
| Expressiveness | Is important domain meaning direct, local, and free of framework ceremony? |
| Static correctness | Which invalid states, effects, architecture edges, ownership errors, and compatibility breaks are rejected? |
| Feedback latency | Warm/private edit, public edit, clean build, selected test, reload, and CI times? |
| Total efficiency | What wall, CPU, memory, disk, network, cache, model-token, tool-call, and human-wait cost produced the evidence? Which work was avoided? |
| Runtime efficiency | Throughput, tail latency, memory, allocation, startup, artifact size, energy where relevant? |
| Predictability | Can tools explain copies, allocation, blocking, scheduling, retries, and performance cliffs? |
| Resource safety | Are acquisition, ownership, cleanup, OOM, cancellation, locks, and foreign handles correct on every exit path? |
| Failure behavior | What happens under malformed input, OOM, cancellation, dependency loss, overload, panic, and partial failure? |
| Concurrency | Are lifetimes, bounds, race freedom, backpressure, fairness, and shutdown explicit? |
| Observability | Can human and agent find cause, authority, state, queue, budget, and evidence without log archaeology? |
| Data integrity | Are schema, lineage, event time, quality, replay, quarantine, and destructive migration behavior explicit? |
| Security/privacy | Least authority, untrusted input, secret/PII handling, supply chain, sandbox, denial-of-service limits? |
| Deployment | Hermetic build, cross-target, container/single binary, configuration, health, rollback, recovery? |
| Evolution | Source, data, wire, durable state, runtime, provider, and mixed-version upgrade behavior? |
| Interop | C/host/Wasm boundary correctness, overhead, debuggability, and unsafe claim containment? |
| AI editability | First-pass correctness, repair rounds, token/context cost, diagnostic usefulness, affected evidence? |
| Human audit | Time to answer what changed, can fail, has authority, costs resources, and needs migration? |
| Ecosystem cost | Direct/transitive dependencies, provenance, maintenance owner, compile impact, upgrade burden? |
| Lifecycle survivability | Can it build offline, reject compromised/stale updates, migrate across editions, recover from a broken compiler/provider, survive maintainer loss, and retire cleanly? |

## Tiered roadmap interpretation

The rows should not all receive equal priority.

1. **Foundation proofs:** pure library plus a native Lang corpus/evidence runner
   containing a streaming JSON/data workload,
   interpreter/native differential execution, C interop, lexical resources,
   ownership, and a compiler query service.
2. **Thesis proof:** a production-shaped modular-monolith web service plus a
   bounded AI workflow. This exercises types, effects, DI, SQL, HTTP, tests,
   evals, telemetry, supervision, upgrades, and the agent protocol together.
3. **Runtime credibility:** HTTP server or emulator, then embedded firmware.
   These deepen already-usable native memory, FFI, layout, concurrency, and
   cross-build rules.
4. **Specialized proof:** GUI/game tool, storage engine, renderer, realtime, or
   kernel according to adopter pull.

The semantic core must not make later rows impossible, but v1 need not claim
them. A profile can mature independently if cross-profile boundaries remain
typed and the portable subset is explicit.

## Convergence rule

Broad exploration has reached diminishing returns when:

- every gauntlet row maps to named core/profile/toolkit mechanisms or a clearly
  recorded incompatibility;
- no new workload reveals a contradiction in the core model across two review
  passes;
- the remaining high-impact unknowns have executable experiments;
- a 20–30 program syntax corpus covers the foundation, thesis, native, failure,
  concurrency, evolution, and AI cases;
- warm and clean compiler budgets, runtime budgets, and evidence policies are
  numeric;
- one first vertical proof is selected.

After that point, adding more conceptual rows is less valuable than moving the
highest-priority row from G2 to G3.

## Update protocol

For each prototype:

1. Record the commit/artifact, toolchain version, target, and benchmark host.
2. Score every quality dimension with links to evidence and failures.
3. Promote a level only when its definition is fully met.
4. Record workarounds and profile escapes; frequent escapes are design data.
5. Add a new row only when it introduces a genuinely distinct semantic stress.
6. Revisit the design baseline when two independent workloads expose the same
   core mismatch.
