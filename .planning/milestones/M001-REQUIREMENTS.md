# Requirements Archive: M001 Source-to-Native Semantic Spine

**Archived:** 2026-09-07
**Status:** SHIPPED

For current requirements, see `.planning/REQUIREMENTS.md`.

---

# Requirements: Codename Lang — M001 Source-to-Native Semantic Spine

**Defined:** 2026-09-03
**Core Value:** Give AI agents and human reviewers the shortest reliable path
from intent to sound, reproducible evidence without wasting iteration time or
hiding runtime costs.

## M001 Requirements

### Foundation and contracts

- [x] **FND-01**: A contributor can build and run the Stage 0 compiler from a
  clean checkout offline using Go 1.24 and its standard library.
- [x] **FND-02**: Compiler artifacts expose versioned, deterministic identities
  for schemas, nodes, symbols, CFG points/edges, diagnostics, and semantic events.
- [x] **FND-03**: A compact evidence manifest binds source/core digests,
  toolchain, target, flags, and policy, and rejects any deliberately stale input.
- [x] **FND-04**: Every compiler command can report wall time, peak memory,
  output bytes, cache status, and affected/recomputed work without changing
  semantic output.

### Canonical source and frontend

- [x] **SYN-01**: A contributor can parse and canonically format the Phase 1
  subset: a module, explicit export block, closed nominal variants, and one
  function with a named parameter, block, and exhaustive `match`; later phases
  add bindings, calls, and only the syntax demanded by their vertical slice.
- [x] **SYN-02**: Parse → format → parse preserves the semantic tree, formatting
  is idempotent, and comments remain stably attached.
- [x] **SYN-03**: Malformed delimiters, matches, signatures, and ownership forms
  recover within a bounded region and emit stable structured diagnostics.
- [x] **SYN-04**: The formatter owns one conventional source form and does not
  reorder named-argument evaluation or create whitespace-only choice churn.

### Typed semantic core

- [x] **SEM-01**: The frontend lowers source into an immutable, serializable core
  with explicit nominal types, ADTs, fields, functions, places, operations,
  source spans, and stable semantic identities.
- [x] **SEM-02**: Closed variant matches are exhaustive and deterministic;
  missing, unreachable, or subsumed alternatives are rejected before execution.
- [x] **SEM-03**: `Result` propagation and ignored-result rules produce explicit
  typed control flow; panic and cancellation cannot be erased as ordinary errors.
- [x] **OWN-01**: Noncopyable values transfer exactly once, and use-after-move or
  move-during-loan programs are rejected with the transfer and conflict causes.
- [x] **OWN-02**: Copy, drop, share, send, and escape abilities are derived
  independently, including through generic and aggregate types.
- [x] **OWN-03**: Shared and exclusive loans obey conflict rules and ordinary
  local loans end at proven CFG point/edge-specific last use.
- [x] **OWN-04**: Public borrowed results record verified field/alternative
  origins and access mode without inspecting provider bodies downstream.
- [x] **RES-01**: Partially initialized noncopyable resources release exactly
  once in reverse completed-acquisition order on return and typed failure.
- [x] **FFI-01**: Foreign contracts carry target layout, initialized state,
  allocator identity, capture/retention, aliasing, and unwind obligations.

### Execution and native equivalence

- [x] **INT-01**: A deterministic interpreter executes valid core programs and
  emits ordered semantic events for values, moves, loans, initialization,
  release, failure, and simulated foreign operations.
- [x] **INT-02**: Interpreter mismatches report the smallest known source/core
  case, causal facts, and addressable event trace.
- [x] **NAT-01**: The compiler emits readable C17 and interface declarations for
  the portable milestone subset, then builds and runs them with Clang.
- [x] **NAT-02**: Interpreter, native `-O0`, and native `-O3` agree on terminal
  outcome, semantic-event order, and live-resource state for valid probes.
- [x] **NAT-03**: Native validation detects layout mismatch, missing partial
  cleanup, allocator mismatch, stale callback retention, false no-alias facts,
  use-after-free, and foreign nonlocal-exit cleanup bypass.

### Agent and human feedback

- [x] **DX-01**: One root command provides `format`, `check`, and
  `run --engine=interpreter|native` with concise human output and versioned JSON.
- [x] **DX-02**: `explain` and `query` expose diagnostic cause graphs, symbols,
  types, ownership, dependencies, and affected tests by stable identity.
- [x] **DX-03**: `verify` selects deterministic, property, mutation,
  differential, optimized, and sanitizer lanes by changed risk, while `evidence`
  emits or validates the compact manifest and expands detail only on demand.
- [x] **DX-04**: An agent can introduce, locate, and repair representative match,
  move, borrow, cleanup, and stale-evidence defects without scraping prose.
- [x] **QLT-01**: The suite preserves every relevant negative control and reduced
  counterexample from Spikes 001–005 while making the coordinated false
  source-to-core claim an explicit expected escape.
- [x] **QLT-02**: Compiler feedback budgets are ratified as cold/warm
  distributions on declared machines, and regressions identify the responsible
  stage and affected-work expansion.

## Later Milestones

### Language semantics

- **EFF-01**: Static non-resumable capabilities and provider graphs with
  first-class test substitution and explicit authority.
- **CON-01**: Structured concurrency, tasks, actors, bounded queues,
  cancellation, backpressure, and supervision.
- **MEM-01**: Profile-specific reclamation beyond stack/inline ownership,
  including measured shared immutable and owner-confined cyclic strategies.
- **TRU-01**: Policy-scoped trust, confidentiality, validation, and safe-sink
  information-flow analysis.

### Ecosystem and workloads

- **PKG-01**: Hermetic direct-dependency packaging, lock graphs, build authority,
  provenance, migrations, and compatibility policy.
- **KIT-01**: First-party boundary kits for exact JSON/schema validation,
  SQL/changes/transactions/pools, HTTP/TLS, telemetry, caching, and workflows.
- **AI-01**: Model/tool capabilities, agent workflows, budgets, replay, evals,
  and live introspection over the stable compiler/runtime protocol.
- **APP-01**: Production dogfood applications spanning CLI, service, data,
  embedded/native, GUI, and compiler/runtime classes.

## Out of Scope for M001

| Feature | Reason |
|---------|--------|
| General algebraic effect handlers | Continuation linearity would enlarge the first ownership model; static calls preserve the seam |
| Async, actors, scheduler, GC, hot reload | Sequential ownership/resource semantics must become trustworthy first |
| Package manager and registry | No ecosystem contract should precede a usable language slice |
| SQL, HTTP, TLS, JSON derivation, GUI, AI runtime | First-party kits are downstream consumers of the core, not prerequisites |
| Full generics, traits, variance, subtyping, partial moves | High semantic and diagnostic surface without evidence they unlock M001 |
| Borrow across suspension | Intentionally rejected until structured concurrency storage experiments exist |
| LLVM/Cranelift/QBE integration | C17 is the control path until a measured bottleneck justifies backend complexity |
| Permanent syntax freeze | M001 validates a formatter-owned provisional grammar and stable core contract |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| FND-01 | Phase 1 | Complete |
| FND-02 | Phase 1 | Complete |
| FND-03 | Phase 1 | Complete |
| FND-04 | Phase 6 | Complete |
| SYN-01 | Phase 1 | Complete |
| SYN-02 | Phase 1 | Complete |
| SYN-03 | Phase 1 | Complete |
| SYN-04 | Phase 1 | Complete |
| SEM-01 | Phase 1 | Complete |
| SEM-02 | Phase 1 | Complete |
| SEM-03 | Phase 4 | Complete |
| OWN-01 | Phase 2 | Complete |
| OWN-02 | Phase 2 | Complete |
| OWN-03 | Phase 3 | Complete |
| OWN-04 | Phase 3 | Complete |
| RES-01 | Phase 4 | Complete |
| FFI-01 | Phase 4 | Complete |
| INT-01 | Phase 1 | Complete |
| INT-02 | Phase 5 | Complete |
| NAT-01 | Phase 1 | Complete |
| NAT-02 | Phase 5 | Complete |
| NAT-03 | Phase 5 | Complete |
| DX-01 | Phase 1 | Complete |
| DX-02 | Phase 6 | Complete |
| DX-03 | Phase 6 | Complete |
| DX-04 | Phase 6 | Complete |
| QLT-01 | Phase 5 | Complete |
| QLT-02 | Phase 6 | Complete |

**Coverage:**

- M001 requirements: 28 total
- Mapped to phases: 28
- Unmapped: 0 ✓

> **Traceability linter note.** `gsd-tools phase complete` reports EFF-01, CON-01,
> MEM-01, TRU-01, PKG-01, KIT-01, AI-01 and APP-01 as "found in body but missing
> from Traceability table". That is expected and must not be actioned: those eight
> live under **Later Milestones** and are deliberately outside M001. Adding them
> here would falsify the coverage counts above. The linter scans the whole document
> without respecting the milestone boundary.

## Definition of Done

M001 is complete only when every requirement is implemented, independently
verified, and represented by reproducible evidence from a clean checkout. A
green build without interpreter/native equivalence, preserved negative controls,
or bounded structured diagnostics is not sufficient.

---
*Requirements defined: 2026-09-03*
*Last updated: 2026-09-03 after initial roadmap definition*
