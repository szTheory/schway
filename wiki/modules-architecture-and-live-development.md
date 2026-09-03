---
id: modules-architecture-live-development
title: Modules, architecture conventions, and live development
summary: Explicit private-by-default modules, verified component boundaries, conventional project layout, canonical punning, and a typed transactional successor to the REPL.
type: design
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-02
tags: [modules, architecture, repl, compiler]
related: [application-architecture-data, compiler-feedback-latency, language-toolchain-primer, design-baseline, implementation-path, research-ledger]
---

# Modules, architecture conventions, and live development

## Reader and outcome

This note is for a compiler, toolchain, or application-architecture
contributor. After reading it, they should be able to create a module/component,
predict every legal dependency, and use the live workspace without introducing
hidden state or a second dynamic language.

## Recommendation

Separate four ideas that many languages conflate:

1. A **file** is a source and diff unit.
2. A **module** is a namespace, privacy, recursion, and compilation unit.
3. A **component** is an architectural ownership and capability boundary made
   from modules.
4. A **package** is a separately versioned and resolved distribution unit made
   from components.

The default filesystem layout projects these semantics predictably, but paths
are not the only architecture truth. The compiler verifies the declared graph,
not folder-name folklore.

## Modules and exports

Modules may export more than one item. The safe default is private; public API
is an explicit, reviewable block:

```text
module shop.orders

export {
  type OrderId
  type PlaceOrder
  fn place_order
  capability OrderQueries
}

record InternalRow { ... }
fn translate_store_error(error:) -> PlaceOrderError { ... }
```

Advantages over capitalization- or annotation-only export:

- the complete public surface is locally scannable;
- removing, renaming, or adding API produces a concentrated semantic diff;
- docs, compatibility, contract, and evidence obligations derive from one
  declaration;
- accidental exposure and wildcard re-export are difficult;
- generated code cannot make an item public merely by choosing a capitalized
  name.

Multiple exports are normal; API-size budgets and cohesion checks discourage a
god module. Re-exports are explicit and retain origin. Wildcard imports,
ambient preludes, import-for-side-effect, and runtime monkey-patching are not
part of ordinary modules.

```text
import shop.money { Money, Currency }
import shop.orders.types as order_types
```

Qualified use is the default when names or authority would be ambiguous. The
formatter groups and sorts imports by stable package/module identity, not line
length preference. Unused and undeclared direct imports are errors.

## Dependency and recursion rules

The package/component/module dependency graph is acyclic. The compiler prints
the shortest cycle and architectural repair options. This improves incremental
compilation as well as reasoning: an implementation or public-signature change
has a directed impact cone.

Mutually recursive types/functions may live in one module inside an explicit
`recursive` group. This avoids manufacturing tiny cyclic modules to express one
coherent algorithm. Top-level effectful initialization is forbidden; providers
and application startup own effects and order. Pure constant initialization
must be deterministic and cycle-free.

Structural ports remain consumer-owned:

```text
module shop.checkout.application

capability InventoryNeeded {
  reserve(items: NonEmpty<OrderLine>)
    -> Result<Reservation, InventoryError>
}
```

An adapter can satisfy the shape without depending on a central interface
package. Public/security/foreign conformance receives an explicit witness so
accidental structural matches do not silently grant authority.

## Components and verified architecture

```text
component orders {
  owns { schema "orders", topic "order-events" }
  exports { api, events }
  may_use { payments.api, inventory.api, platform.clock }

  layer domain     may_use { shared_kernel }
  layer application may_use { domain, ports }
  layer ports       may_use { domain }
  layer adapters    may_use { application, ports, infrastructure }
  layer interface   may_use { application, api }
}
```

This is candidate manifest notation. It encodes enforceable facts:

- component graph and permitted layer directions;
- exported modules/types/capabilities/events;
- owned storage, queues, assets, and configuration namespaces;
- allowed cross-component APIs and forbidden implementation imports;
- effect/authority budgets by layer;
- service-extraction and contract boundaries.

The compiler can enforce a declared architecture perfectly. It cannot invent
perfect domain boundaries. Templates provide an opinionated modular-monolith
default; teams can declare another acyclic policy deliberately.

## Conventional folder projection

The service template renders a component locally:

```text
src/
  orders/
    domain/
    application/
    ports/
    adapters/
    interface/
    specs/
```

The project formatter/linter enforces this mapping for template-created
projects. `lang move` performs a semantic module move, updates imports, previews
the architectural impact, and preserves stable symbol identity where possible.

Escape hatch: a package may declare a custom deterministic module-to-path map.
That is useful for foreign layouts and migrations, but it cannot change the
component graph or bypass policy. There is no generic `utils`, `common`, or
`shared` dumping ground by default. A shared kernel is named, small, dependency
free, and carries an API/evolution budget.

Folder convention is valuable for discovery and diff locality; making every
DDD vocabulary word core grammar would freeze one architecture fashion.

## Canonical exact-name punning

The project should reduce equivalent surface forms. When a named argument or
record field resolves to an unqualified local of the exact same name:

```text
Inventory.release(reservation: reservation)
```

the formatter canonicalizes it to:

```text
Inventory.release(reservation:)
```

The long form remains accepted by the parser for migration/paste friendliness
but is not canonical checked-in source. Punning is allowed only in named calls,
constructors, updates, and patterns where name resolution is exact. It does not
apply across qualification, conversion, aliasing, or a different spelling.

This policy reduces source entropy and tokens without removing the semantic
role label. The corpus experiment still tests whether frequent punning creates
shadowing mistakes; if it does, the formatter policy can be reversed without
changing the semantic model.

## A typed live workspace, not merely a REPL

The useful successor to a classic read-eval-print loop is a persistent semantic
workspace shared by the editor, terminal, debugger, tests, and agents:

```text
lang live

> :type checkout
> :effects checkout
> :providers checkout
> :why-error diagnostic:184
> :sample Dataset.orders partition:today rows:20
> :profile checkout for:5.seconds
> :trace request:01K...
> :promote cell:17 to spec "declined payment releases inventory"
```

Every cell is a transaction with:

- source and semantic dependencies;
- inferred/declared types, effects, authority, and resource budget;
- provider world and runtime/deployment target;
- result or structured failure;
- dataset/config/artifact snapshot references;
- whether external effects occurred and whether they can be replayed,
  compensated, or neither.

Pure cells are deterministic and content-cacheable. Effectful cells require an
explicit provider and execute capability; preview/simulated providers are the
default for dangerous operations. Undo can revert workspace state, not pretend
to undo an email or arbitrary network write.

Sessions can be saved as canonical transcripts. Useful expressions promote
directly into modules, specs, benchmarks, evals, pipeline stages, or runbooks
with provenance. This addresses the classic REPL failure where knowledge lives
only in ephemeral process state.

## Live code and state evolution

Redefining a pure function creates a new semantic version in the workspace and
updates eligible call sites transactionally. Running resources, actors,
workflows, native frames, and foreign callbacks may still reference older code.
The runtime reports them explicitly.

Automatic reload is limited to compatible stateless changes. Stateful changes
require a typed migration; ABI/layout or unsafe native changes generally
restart the owning isolation boundary. The experience should be fast, but never
pretend that all live state can be safely reinterpreted.

## Interpreter, JIT, and compiler

The answer is **both interpretation and compilation**, with one semantic core:

```text
canonical source
  -> resolved typed/effect/ownership IR
       -> reference interpreter
       -> fast development native code generator / JIT
       -> optimized AOT backend
       -> later BEAM and Wasm backends
```

- The reference interpreter supplies executable semantics, constant evaluation,
  deterministic tests, and a differential oracle. It need not run production
  services quickly.
- A fast native backend is required in v1 because the author explicitly chose
  usable low-level ownership/resource programming in the first implementation.
- The development generator prioritizes compile latency and usable code. A
  Cranelift spike is the leading candidate because it supports library use,
  JIT/AOT, multiple major architectures, and explicitly prioritizes compilation
  speed and a defined IR without undefined behavior.
- An optimized LLVM backend may follow for peak native performance and targets,
  but its compile time and more subtle IR contract must earn the cost.
- BEAM remains a valuable service backend for OTP interop and mature process
  semantics, but it is no longer the foundation that defines v1 semantics.

JIT is optional. A bytecode/register interpreter may deliver the better first
live-workspace loop. The decision is benchmarked using startup, private/public
edit latency, code quality, debugger/profiler integration, implementation size,
memory, and semantic fidelity.

Every backend consumes the same portable core contract and runs differential
and conformance suites. Backend-specific capabilities and representation limits
remain explicit profiles rather than silent behavior changes.

## Module information as compiler leverage

Explicit exports, no import cycles, pure initialization, component ownership,
and stable symbol identity improve both correctness and speed:

- private implementation edits avoid downstream type checking/code generation;
- public fingerprints select the exact invalidation cone;
- components and packages compile in parallel by DAG layer;
- direct dependency declarations make cache keys and hermetic builds complete;
- no side-effectful module initialization means analysis/build order does not
  change behavior;
- semantic moves can preserve identity and caches where meaning did not change.

This is a genuine compounding bundle: architecture constraints improve agent
context selection, human comprehension, incremental compilation, test
selection, release evidence, and future service extraction simultaneously.

## Anti-patterns

- Public-by-default definitions or wildcard re-exports.
- Import cycles disguised through runtime lookup or a service locator.
- Effectful package/module initialization.
- Equating a directory tree with verified architecture.
- One-interface-per-file ceremony or a single-export restriction.
- A central interfaces package that every component depends on.
- Custom path mappings that bypass component policy.
- A dynamic REPL with semantics different from compiled programs.
- Pretending workspace undo reverses external effects.
- Reloading state/ABI changes without compatibility and migration evidence.
- Choosing one heavy optimizer for both subsecond editing and peak release code.

## Primary evidence

- [Go specification](https://go.dev/ref/spec) makes direct or indirect package
  import cycles illegal and defines import dependencies and initialization
  order, demonstrating the build/reasoning leverage of an acyclic package graph.
- [Clojure REPL guidance](https://clojure.org/guides/repl/guidelines_for_repl_aided_development)
  describes the REPL as a powerful program interface while warning that its
  ephemeral state is poor breakage detection and should be preserved elsewhere.
- [Erlang code loading](https://www.erlang.org/doc/system/code_loading.html)
  demonstrates that live systems can concurrently contain current and old code
  and that further replacement can terminate processes still using old code.
- [Cranelift](https://cranelift.dev/) describes a library-oriented JIT/AOT
  backend designed for compiler speed, security, relatively simple
  implementation, and an IR without undefined behavior.
- [LLVM ORC](https://llvm.org/docs/ORCv2.html) supports eager/lazy and concurrent
  JIT compilation and illustrates that compilation timing and optimization are
  composable tradeoffs rather than one universal choice.

All sources were last verified 2026-09-02. The module surface and staged
execution strategy are project synthesis.
