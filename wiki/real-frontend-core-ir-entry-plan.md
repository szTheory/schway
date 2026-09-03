---
id: real-frontend-core-ir-entry-plan
title: Real frontend and core-IR entry plan
summary: The GSD-ready tracer-bullet plan that turns five ownership workbenches into one real source-to-core-to-native compiler spine without prematurely freezing syntax or runtime scope.
type: strategy
status: candidate
confidence: high
created: 2026-09-03
updated: 2026-09-03
tags: [bootstrap, frontend, core-ir, native, gsd, roadmap]
related: [ownership-evidence-roadmap, semantic-kernel-contract, semantic-kernel-probes, convergence-work-program, implementation-path, compiler-feedback-latency, compute-efficiency-constitution, research-ledger]
---

# Real frontend and core-IR entry plan

## Reader and action

This note is for the project author or an implementation agent starting the
first non-disposable Lang code. After reading it, they should be able to turn
the milestone into GSD phases and vertical slices without reopening broad
language ideation, freezing decorative syntax, or extending the Go workbenches
as a shadow compiler.

## Outcome

Build one honest tracer bullet:

```text
provisional canonical source
  -> lossless concrete syntax tree
  -> resolved typed ownership core
  -> deterministic interpreter events
  -> C17 development lowering
  -> native O0/O3 events
  -> compact evidence manifest + structured diagnostics
```

The tracer bullet succeeds when a small ownership/resource corpus parses,
formats idempotently, checks with stable diagnostics, runs in the interpreter,
lowers through Clang, and produces equivalent semantic events. It must retain
the negative controls from Spikes 001–005.

This is not a toy parser success criterion. The vertical must include a moved
value, shared and exclusive access, a branch-specific last use, a public
borrowed result, a noncopyable resource with partial initialization/cleanup,
and one C boundary.

## Decisions carried in

These are implementation inputs, not topics for the first planning discussion:

- affine owned values with separate `copy`, `drop`, `share`, `send`, and
  `escape` abilities;
- ordinary local loans end at proven CFG last use, including edge-specific
  endpoints;
- public borrowed results carry verified value/field origins, alternatives,
  and shared/exclusive access;
- fresh higher-ranked origins exist only at genuine callback lending scopes;
- compact content-bound typed-core validation runs at trust crossings, while
  full traces are mismatch-only evidence;
- target layout, initialized state, allocator identity, retention/capture,
  cleanup, and unwind policy are typed foreign-boundary obligations;
- optimizer attributes are derived from checked facts, never asserted as
  casual hints; and
- no mandatory tracing heap or service runtime is required by this slice.

The authoritative evidence is in the five spike READMEs and the ownership
evidence roadmap. Their Go/C notation is not the source grammar.

## Bootstrap recommendation

Use Go 1.24 and its standard library for the Stage 0 compiler host, with C17
emission plus Clang as the first native development backend.

This is a reversible bootstrap choice:

- Go is already installed and has sustained fast, dependency-free work across
  all semantic experiments.
- Explicit structs and state machines fit lossless syntax and typed-core work.
- A single static host tool keeps onboarding, CI, and agent execution simple.
- C17 emission exposes ABI and optimizer behavior immediately without choosing
  a large backend framework before the IR is trustworthy.
- The stable asset is the versioned core/evidence contract. Frontend passes can
  later be strangled into Lang, and the native backend can be replaced without
  changing source semantics.

This choice does not claim Go is the ideal permanent compiler host. Reconsider
it only if measured incremental memory/latency, native-backend integration, or
implementation safety becomes a bottleneck. Do not pay Rust dependency and
bootstrap compile cost preemptively; do not reject Rust or Cranelift after the
core contract exists and a fair backend measurement is possible.

## Milestone 0: source-to-native semantic spine

### Phase 0 — repository and evidence foundation

Deliver a production-code area outside `.planning`, one root command, and one
versioned wire-contract area. Establish:

- hermetic standard-library-only host builds;
- stable schema, node, symbol, CFG point/edge, diagnostic, and event identities;
- canonical JSON for typed-core fixtures and compact evidence manifests;
- explicit toolchain, target, source/core digest, flags, and policy metadata;
- one command surface with human and JSON projections; and
- measured cold/warm command time, peak memory, output bytes, and affected work.

Gate: a clean checkout can run the smallest fixture offline with one command,
and a deliberately stale manifest is rejected.

### Phase 1 — lossless syntax and canonical formatting

Implement a small conventional braced control surface only for the constructs
required by this milestone. Preserve whitespace/comments in a lossless CST,
then project one formatter-owned canonical source form.

Required constructs:

- module, import, and explicit export declarations;
- nominal records and closed variants;
- functions with named parameters and explicit public signatures;
- immutable bindings plus explicit mutable bindings;
- calls, blocks, `if`, exhaustive `match`, and `Result` propagation;
- provisional ownership modes sufficient to express own/borrow/exclusive/take;
  and
- resource acquisition with a lexically visible cleanup boundary.

Gate: parse→format→parse preserves the semantic tree; formatting is idempotent;
comments attach stably; malformed delimiter, match, signature, and ownership
edits produce bounded recovery and structured diagnostics. No user-defined
precedence, macros, implicit conversions, or effect grammar enters this phase.

### Phase 2 — resolved typed ownership core

Lower the CST into an immutable portable core with explicit:

- nominal types, ADTs, fields, functions, places, and operations;
- generic ability derivation;
- ownership, initialization, moves, shared/exclusive loans, and release;
- CFG blocks, edges, point/edge loan endpoints, and branch joins;
- public return origins and tagged alternatives;
- foreign layout/capture/allocator/unwind obligations; and
- source spans plus stable semantic identities.

Reuse semantic fixture *concepts* from Spikes 001–004, not their Go APIs. One
declarative rule inventory should generate human documentation, test cases, and
checker tables where practical; independently implemented validators consume
the wire contract without copying frontend inference.

Gate: the migrated bounded corpora retain their accept/reject outcomes,
diagnostic identities, reduced counterexamples, and mutation detection. A
coordinated false frontend fact remains an explicit expected escape.

### Phase 3 — deterministic interpreter

Execute the portable core directly with no native backend. Emit semantic events
for values, moves, loans, initialization, release, failure, and foreign-boundary
simulation. Separate semantic order from wall-clock measurements.

Gate: the relevant blocking semantic probes agree with the static checker;
every mismatch includes the smallest known source/core case, causal facts, and
an addressable trace. The interpreter remains a correctness oracle and live
development engine, not the production-performance claim.

### Phase 4 — minimal C17 lowering

Lower the same core into readable C17 plus generated C interface declarations.
Start with stack/inline values and an explicit allocator/resource shim. Carry
layout, initialized-field flags, reverse cleanup, allocator identity, callback
capture, and unwind policy from typed core into lowering decisions.

Generate `restrict`/no-alias-like facts only from exclusive ownership evidence.
Make ordinary foreign calls potentially aliasing and capturing unless an
audited adapter proves stronger facts. Catch Lang panic at the boundary; reject
foreign unwind/nonlocal exit unless a platform adapter explicitly contains it.

Gate: interpreter, O0, and O3 semantic events agree. The seven Spike 005 hostile
cases remain detected, including the optimizer-only false-no-alias divergence.
ASan/UBSan runs are separate evidence, and generated ABI declarations pass
target layout fixtures.

### Phase 5 — agent/human feedback loop

Expose one compiler service contract through a CLI first:

- `check` returns the smallest sound local diagnostics;
- `run --engine=interpreter|native` selects an execution path;
- `explain <diagnostic-id>` returns the cause graph and obligations;
- `query` returns symbols, types, effects, ownership, dependencies, and affected
  tests by stable identity;
- `verify` runs configured property, mutation, differential, optimized, and
  sanitizer lanes; and
- `evidence` emits or validates the compact manifest, with detailed events only
  on failure or request.

Gate: an agent can introduce, locate, and repair a move/borrow/cleanup defect
using bounded JSON output without scraping prose, while a human receives the
same causal result in a calm projection.

## Vertical slices

Plan each phase as a tracer bullet, not as horizontal infrastructure:

| Slice | User-visible capability | Semantic pressure | Required negative control |
|---|---|---|---|
| S1 canonical pure function | format/check/run one ADT transformation | CST, names, types, match, Result | missing match alternative |
| S2 owned buffer | transfer and reject use after move | places, abilities, stable diagnostic | move during active loan |
| S3 borrowed view | return a view tied to input | public origin/access summary | omitted second origin |
| S4 branch last use | release a loan on one CFG edge | liveness, edge identity, joins | omitted edge endpoint |
| S5 fallible resource | partial acquire and reverse cleanup | initialization, Result, release | missing second release |
| S6 C packet/borrow | pass record and callback view through C | layout, allocator, capture, unwind | field reorder and stale retention |
| S7 optimized native | compare interpreter/O0/O3 | lowering and optimizer contracts | false no-alias fact |
| S8 evidence loop | query/explain/verify the corpus | identities, manifests, bounded output | stale artifact and coordinated lie |

S1 establishes the spine end to end. Later slices deepen that same path. Do not
build the whole parser, checker, interpreter, and backend as four disconnected
horizontal projects.

## Test and evidence matrix

Every slice must combine only the evidence that answers a distinct question:

| Evidence | Always | Changed/affected | Verify/release |
|---|---:|---:|---:|
| deterministic fixtures | yes | yes | yes |
| formatter/CST properties | yes | yes | yes |
| bounded exhaustive ownership generation | affected kernel rules | yes | yes |
| seeded metamorphic properties | sampled | yes | extended |
| independent typed-core validation | package/cache boundary | yes | yes |
| interpreter/native differential | native slice | yes | all targets |
| O0/O3 negative controls | native slice | affected lowering | yes |
| ASan/UBSan | no | changed unsafe/FFI | yes |
| mutation score | focused changed rules | yes | full configured set |
| performance/invalidation series | cheap counters | changed budgets | declared machines |

Cache only hermetic successful results keyed by source/core/schema/toolchain/
target/policy inputs. A cache hit must report what evidence was reused and why
it remains valid. Never let cached “green” hide a changed negative control.

## Work that is intentionally deferred

Do not add these until the source-to-native spine passes its gates:

- general effect rows, provider graphs, resumable handlers, or runtime DI;
- async/await, actors, scheduler, backpressure, managed cyclic regions, or GC;
- packages, registry, remote cache, LSP, MCP server, web framework, SQL, HTTP,
  TLS, GUI, AI inference, or distributed workflow runtime;
- full generics, traits, variance, subtyping, partial moves, or borrow across
  suspension; and
- Cranelift/QBE/LLVM integration beyond the C17 control path.

Their preservation seams already exist in the design. Pull one forward only if
a current slice produces a reduced counterexample that cannot be represented
without it.

## Risk controls and stop conditions

- If the provisional grammar begins deciding semantics, move the fact into core
  IR and keep syntax replaceable.
- If independent validation imports frontend inference, delete that expansion
  and return to compact committed-fact checking.
- If C output becomes the semantic specification, restore interpreter/core
  primacy before adding features.
- If a new abstraction cannot preserve stable diagnostics and bounded affected
  work, simplify it before broadening the corpus.
- If a native attribute lacks a typed-core derivation, omit it even when that
  costs performance.
- If a test lane repeats another lane without catching a distinct seeded defect,
  remove it from the default loop.
- If S1 cannot run end to end quickly, reduce its language surface rather than
  postponing integration.

## GSD handoff

The first GSD milestone should be named **M001 — Source-to-Native Semantic
Spine**. Its planning input is this note plus the semantic kernel contract,
probe suite, ownership evidence roadmap, and Spikes 001–005.

Discussion should confirm only choices that materially change execution:

1. the provisional grammar token forms for the S1–S3 subset;
2. the production code directory/module name;
3. the first command name while the language remains codenamed Lang; and
4. the exact baseline machines used to ratify, rather than merely observe,
   feedback budgets.

Everything else above is either a carried decision, a reversible default, or a
phase gate. The next productive operation is phase planning and execution, not
another broad blind-spot pass.
