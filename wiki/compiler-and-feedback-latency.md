---
id: compiler-feedback-latency
title: Compiler and feedback latency
summary: A latency-first compiler architecture for fast, sound, structured AI and human development loops.
type: design
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-03
tags: [compiler, incremental, latency, ci, ai-feedback]
related: [vision, implementation-path, effects-and-capabilities, trust-validation-information-flow, residual-uncertainty-register, compute-efficiency-constitution, runtime-profiles-dogfooding, convergence-work-program, open-questions, research-ledger]
---

# Compiler and feedback latency

## Recommendation

Treat feedback latency as a product invariant and a language-design constraint,
not as a compiler optimization to attempt later. The metric that ultimately
matters is **time from an incorrect edit to sufficient evidence for the next
correct edit**. Parse/check latency, affected-test latency, diagnostic quality,
and runtime introspection all contribute to that loop.

The compiler should therefore have two deliberately different paths:

- a fast, sound development path that parses, resolves, checks, interprets or
  emits low-optimization code, and runs the smallest affected evidence set;
- an exhaustive release path that adds whole-program optimization, broad
  properties, fuzzing, model checks, security policy, and reproducibility.

The development path may defer expensive evidence, but it may not report a
program as checked when a required soundness obligation was skipped. Every
response names what was checked, deferred, stale, or unknown.

This is governed by **budgeted abundance**: use compute freely when it reduces
meaningful uncertainty or improves evidence, but eliminate redundant,
mis-scoped, unactionable, or uneconomical work. The full project discipline is
in the [compute-efficiency constitution](compute-efficiency-constitution.md).

## Candidate latency service levels

These are hypotheses for benchmarks, not promises:

| Interaction | Candidate warm p95 | Scope |
|---|---:|---|
| Parse/name/type query | 50 ms | one local query after a small edit |
| Sound affected check | 200 ms | invalidated semantic slice |
| Affected unit examples | 1 s | smallest proven affected set |
| Development reload | 2 s | changed component with compatible state |
| Structured incident query | 200 ms | bounded in-memory runtime evidence |

Each number must be reported with project size, change shape, machine class,
cache state, and backend. A headline latency without its workload is marketing.
Cold builds, clean CI builds, large public-signature changes, and optimized
release builds need separate budgets.

## Architecture

```text
source edits
    -> persistent query service
       -> parse/name queries
       -> exported signature + effect fingerprint
       -> dependency and change-impact graph
       -> affected contracts/specs
       -> fast execution backend
    -> immutable content cache
    -> optional release optimizer and exhaustive evidence
```

The CLI, editor, CI runner, formatter, and agent protocol are clients of the
same long-lived compiler service. They must not independently reconstruct the
program and disagree about its state.

### Stable, fine-grained queries

Model compilation as deterministic queries over declared inputs. Record the
dependency edges actually observed by each query and invalidate only consumers
whose semantic input changed. Rust's compiler uses a query dependency graph for
incremental compilation, while its documentation also shows the integration
work required at backend and code-generation boundaries. That is evidence for
the architecture, not evidence that fine-grained invalidation is automatic.
([rustc incremental compilation](https://rustc-dev-guide.rust-lang.org/queries/incremental-compilation-in-detail.html))

Use two identities:

- a source fingerprint for formatting, comments, and source-map changes;
- a semantic fingerprint for exported type, effect, contract, layout, and
  behavior-relevant changes.

A comment or private rename should not rebuild a service graph. Changing a
public effect row or wire schema should invalidate every affected consumer even
when the textual diff is tiny.

### Explicit module boundaries

Modules form an acyclic dependency graph and publish compact interfaces. A
consumer should read a dependency's exported semantic summary, not parse the
entire transitive source graph. Go similarly compiles packages separately and
places export data needed by importers in compiled output.
([Go compiler](https://go.dev/cmd/compile/))

Avoid whole-program type inference across module boundaries. Public types,
effects, capabilities, protocol versions, and relevant contracts are explicit
or materialized into a canonical checked interface. Private inference remains
available where it cannot create distant invalidation or invisible API drift.

### Fast backend and release backend

Do not make maximum optimization the price of running a test. The pure core can
start in a reference interpreter. Native development should use a deliberately
fast backend candidate such as Cranelift; a QBE control helps reveal the minimum
backend actually required. LLVM/PGO remains a later release candidate when a
measured workload needs its optimization or target breadth. Bytecode earns a
place only if repeated live execution justifies another IR and runtime.

Semantics and conformance tests are shared; optimization tiers may change speed
and representation, never results, overflow behavior, failure behavior, or
effect ordering.

### Specialization has a budget

Unbounded monomorphization, type-level evaluation, macro expansion, and generic
instantiation can make small source changes expensive. Generics should default
to a predictable implementation strategy; explicit or profile-guided
specialization is budgeted and explained. Go's own release notes record a build
speed regression after generics and later recovery, demonstrating that type
system choices affect the edit loop.
([Go 1.20 release notes](https://go.dev/doc/go1.20))

Dependent or refinement reasoning belongs on a progressive assurance ladder.
Lean demonstrates the leverage of dependent types and a small proof-checking
kernel, while its documentation also explains how type-level computation can
slow compilation. The lesson is to support optional proof artifacts and a
small trusted checker without putting arbitrary proof search in every edit.
([Lean dependent types](https://lean-lang.org/functional_programming_in_lean/Programming-with-Dependent-Types/),
[Lean reference](https://lean-lang.org/doc/reference/latest/))

## Affected evidence

The compiler owns one graph linking definitions to:

- callers and implementations;
- types, effects, capabilities, and provider bindings;
- examples, properties, fixtures, and generated boundaries;
- wire, storage, and actor-state schemas;
- architecture and security policies;
- runtime events and production evidence.

`verify --changed` traverses this graph and returns both the selected evidence
and the reason each item was selected or omitted. Selection is conservative:
when the graph cannot prove non-impact, it widens the slice.

```text
$ ai verify --changed --format json
{
  "checked": ["Checkout.place_order", "spec:payment_decline"],
  "deferred": [{"kind": "fuzz", "reason": "release_lane"}],
  "invalidated_by": ["Payments.charge effect changed"],
  "sound_for": ["types", "effects", "architecture", "selected_specs"]
}
```

## Worktrees, local caches, and CI

Git supports multiple working trees attached to one repository. Compiler state
must therefore use content identities and repository-relative names, never
assume one checkout path, and never let mutable analysis from one worktree
contaminate another. Immutable artifacts may be shared by digest; active query
state is namespaced per worktree.
([Git worktree](https://git-scm.com/docs/git-worktree))

Local and CI builds execute the same hermetic action graph. Go's build cache and
reproducible-build discussion show the value of content-derived action IDs, but
also expose a caution: undeclared foreign inputs can escape a cache key. FFI
headers, link libraries, environment, tools, and build-plugin inputs must be
declared.
([Go build cache](https://go.dev/cmd/go/),
[Go reproducible builds](https://go.dev/blog/rebuild))

## Compile-time execution policy

Arbitrary build scripts and compile-time metaprograms are latency,
reproducibility, and supply-chain hazards. Derivations and plugins must:

- run in a sandbox with explicit capabilities;
- declare content-addressed inputs and outputs;
- have time, memory, output-size, and recursion budgets;
- be deterministic unless their nondeterminism is declared and prevents reuse;
- expose generated semantic nodes and source maps;
- never block parsing or basic semantic inspection of the rest of the project.

The toolchain reports which plugin caused latency or invalidation. A package
cannot acquire network, process, filesystem, or secret authority during build
merely because it is a dependency.

## Self-observation

The compiler should explain its own work:

```text
ai compile explain --last
  43 ms parse/check
  11 queries invalidated
  public effect fingerprint changed in Payments
  18 downstream modules rechecked
  312 ms generated test inputs
  0 cache downloads; local rebuild was cheaper
```

CI can enforce budgets for query fan-out, cold/warm check time, generated code,
specialization count, cache size, and peak memory. Regressions become reviewable
data tied to a change rather than folklore about the compiler getting slower.

## Footguns

- **Overly coarse units:** package-level invalidation makes small changes large.
- **Overly fine queries:** bookkeeping and hashing can cost more than recompute.
- **Global inference:** a local implementation edit changes distant inferred APIs.
- **Optimization in the edit lane:** release-quality code generation blocks tests.
- **Specialization explosion:** convenient generics multiply codegen units.
- **Unsound cache hits:** undeclared files, libraries, environment, or plugins
  reuse stale output.
- **Semantic tier drift:** debug and release produce different behavior.
- **False affected-test precision:** an incomplete graph skips required evidence.
- **Fast but opaque errors:** low compiler latency still produces a slow repair
  loop if diagnostics omit cause and machine-applicable remedies.

## Decisions still requiring measurement

1. Which query granularity wins on the corpus rather than microbenchmarks?
2. Which dev backend gives the best check-to-run latency for service code?
3. Should generic code use dictionaries, specialization, or a hybrid by profile?
4. What assurance can safely remain asynchronous after a local edit?
5. Can stable semantic fingerprints survive ordinary refactors without hiding
   behavior changes?
6. What latency targets remain credible at 10 K, 100 K, and 1 M lines?

The exact experiments and backend comparison are specified in the
[convergence work program](convergence-work-program.md).
