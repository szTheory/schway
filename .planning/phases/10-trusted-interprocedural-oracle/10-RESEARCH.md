# Phase 10: Trusted Interprocedural Oracle - Research

**Researched:** 2026-09-11
**Domain:** Compiler internals — Go implementation of an interprocedural call-stack
interpreter, two independent re-derivation packages (`originvalidate`,
`pathoracle`), and a four-peer differential harness. No external libraries;
100% in-repo Go stdlib work against `github.com/codename-lang/lang` (go.mod:
`go 1.24`, zero external deps).
**Confidence:** HIGH

## Summary

Phase 10 has no unknown technology — it is a `[VERIFIED]` shipped-precedent
extension exercise inside a single Go module, and the CONTEXT.md produced by
`/gsd-discuss-phase` already re-verified every load-bearing claim against the
tree with file:line citations (D-10-01 through D-10-61). This research pass
independently re-opened every cited file and re-confirmed the citations
(all matched byte-for-byte); nothing in CONTEXT.md's factual claims was
found stale. Rather than re-deriving decisions CONTEXT.md already locked,
this document translates those decisions into the RESEARCH.md shape the
planner consumes: it names the exact shipped precedents to mirror, the
exact signatures to widen, the exact test files each new guard/fixture
belongs in, and the pitfalls specific to Go's stack-overflow semantics and
this repo's "independent re-derivation" discipline.

There are three structurally distinct workstreams, confirmed independent at
the Go package-import level (`go list -deps` closure), that the roadmap's
Parallel note licenses to build concurrently: (1) `originvalidate` widening
its callee-contract map and adding an `OpCall` case to `walkReturnOrigin`
(mirrors `corevalidate`'s already-shipped `peerLoanCarry` threading
pattern, D-09-03); (2) `pathoracle` composing callee path enumerations
recursively through its own `EnumeratePaths`, never adding a contract hop
(the one irreversible design commitment in the phase — see D-10-09); (3)
`interp` growing an explicit heap-allocated `[]frame` stack with a bounded
depth constant, never Go native recursion (D-10-21 — this is the one
constraint no design alternative satisfies, because Go's stack-overflow
`throw` cannot be `recover()`'d). All three converge at a mandatory
mid-phase gate (D-10-59) before the phase's fourth workstream — the
three-way-on-refuse/four-way-on-accept differential (D-10-53) — begins.

**Primary recommendation:** Do not invent new patterns. Every subsystem this
phase adds has a byte-identical shipped sibling already in this codebase
(`callgraph.Order`'s iterative DFS for the frame stack; `corevalidate`'s
`peerLoanCarry` map-threading for `originvalidate`'s widening;
`pathCapError`'s typed-cap-error shape for the new composition-depth cap;
`opCallGroupedArmForTest`/`TerminatorKindsOverride`'s nil-default seam
convention for every new fault-injection point). The planning task is
almost entirely "locate the shipped precedent, mirror its shape, wire the
one new case" — not general compiler-construction research.

## User Constraints

<user_constraints>

### Locked Decisions

All sixty-one `D-10-NN` decisions in `10-CONTEXT.md` are locked — synthesized
by twelve parallel advisor-researcher agents and independently re-verified
against the shipped tree by the orchestrator (and re-verified a third time
in this research pass; see Verification Log below). They cover: the
`originvalidate` widening shape (D-10-01–08), `pathoracle`'s composition
design (D-10-09–15), the import-control mechanism gap (D-10-16–20), the
interpreter call stack (D-10-21–26), the D-09-53 reversal (D-10-27–30),
drop ordering across frames (D-10-31–35), OWN-05b's narrowed independence
claim (D-10-36–41), the Pitfall-4 stack probe (D-10-42–44), QLT-04's
declared composition depth (D-10-45–50), the four-way differential
(D-10-51–55), the interp stability freeze (D-10-56–58), and cross-cutting
discipline (D-10-59–61). Full text is authoritative and is not reproduced
here in full — see `10-CONTEXT.md` `<decisions>`. Key load-bearing points
restated for planner convenience:

- `originvalidate`'s `OpCall` widening threads a **narrow callee-contract
  map** (never `core.Program`), built once in `ValidatePublished`, mirroring
  `corevalidate`'s shipped `peerLoanCarry` pattern exactly (D-10-02/03).
- `pathoracle` **composes** callee path enumerations recursively through its
  own `EnumeratePaths`/`linearizePath` — it does NOT add a contract hop, even
  though `check` and `corevalidate` both already use contract hops at
  `OpCall`, because a third contract-hop procedure would collapse TRU-03's
  "structurally distinct" claim exactly at call boundaries (D-10-09/10/11).
  **This is a one-way, irreversible design commitment.**
- The interpreter call stack is an **explicit heap-allocated `[]frame` stack
  with iterative dispatch**, never Go recursion — Go's stack-exhaustion
  `throw` cannot be `recover()`'d, so recursion is architecturally incapable
  of satisfying SEM-08's "named refusal, not host stack overflow" (D-10-21).
  `MaxCallDepth = 128`, deliberately **below** the real structural ceiling
  (1024, from `maxFunctions`) so it actually fires — the inverse direction
  from `pathoracle.MaxPaths` (D-10-23).
- Import guards must be **hardened from direct-import (`go/parser`
  `ImportsOnly` scan) to transitive (`go list -deps`)**, and
  `originvalidate`'s guard must be extended to also forbid `corevalidate`
  (currently forbids only `check`/`ast` — a concrete, previously-unrecorded
  gap, D-10-16/17).
- D-09-53's diagnosis is **reversed**: the register's proposed narrowing
  breaks 8 currently-passing tests when empirically applied. The real gap is
  `corevalidate` having no interprocedural `usesParam` peer at all — re-filed
  as new debt (D-10-27–30), not silently re-pointed. **The planner must
  re-run the throwaway patch as a pre-flight** and confirm the 8 failures
  before writing the correction (D-10-30).
- SEM-09's nonlocal-exit clause is **narrow, not vacuous** — Codename Lang
  has no exceptions/unwinding; `OpCall` is never a terminator
  (`core.go:559-564`); the only >1-frame-skip constructs are the foreign
  landing pad extended to N frames and the SEM-08 depth refusal if
  implemented as an abort (D-10-31). Frame attribution reuses the existing
  `execution.Event.FunctionID` field — **no new event-schema field**, because
  a depth field would manufacture guaranteed false divergences once Phase
  11's `-flto` build inlines callees (D-10-32).
- Criterion 4's differential is **"three-way on refuse, four-way on
  accept"** — never unqualified "four-way" — because `interp` cannot produce
  a `core.LoanEndpoint` and only exists for programs the admission layers
  accept (D-10-53). It **extends** `session_peer_gate_test.go`, never a
  second harness (D-10-51).
- QLT-04 declares **composition depth = 3** (necessity-minimum 2, per
  D-09-49 Q2, plus one sufficiency margin) — and no depth-3 fixture exists
  anywhere in the tree today (verified: `grep -rn "depth3\|depth_3"` over
  `testdata/`/`internal/` returns nothing) (D-10-45–50).
- Four hard planning constraints bind wave ordering regardless of planner
  discretion: **SEM-09 before the stability freeze** (D-10-56); **the
  mid-phase gate scheduled before criterion 4's differential work**
  (D-10-59); **each differential lands in the same plan as the peer it
  tests, red-first** (D-09-27 precedent); **widen-then-use as two commits
  in one plan** for `originvalidate` (D-10-06).

### Claude's Discretion

- Plan decomposition and wave ordering, subject to the four hard constraints
  above.
- Exact Go identifier names for the frame type, the depth constant's
  siblings, the composition-depth constant, the new seams, and the
  callee-contract map's value type.
- Whether `pathoracle`'s composition lives in `pathoracle.go` or a sibling
  file in the same package.
- The relative landing order of `originvalidate`, `pathoracle`, and
  `interp`'s call stack — independent at the implementation level per the
  roadmap's Parallel note.
- Whether the `corevalidate` `usesParam` peer (D-10-28) gets a proposed
  landing phase now or is left open — either is honest; recording it is
  required.

### Deferred Ideas (OUT OF SCOPE)

- A `corevalidate` interprocedural `usesParam` peer (D-10-28) — landing
  phase open, record in `PHASE-10-DEBT.md`.
- Multi-function `cgen` and interprocedural `-O3`/LTO equivalence — Phase 11
  (NAT-04–07). `cgen`'s multi-frame nonlocal pad (D-10-34) would ideally
  land here; if it slips, it is a named single-authority gap.
- QLT-03's call-graph-shape reachability register — Phase 11.
- Persistent cross-run summary caching / QLT-06 — Phase 11.
- `Result` payloads — Phase 12. D-10-32's `FunctionID`-as-frame-identity
  decision holds only while recursion stays refused.
- `originvalidate`'s `callgraph` import asymmetry (D-10-19) — recorded, not
  resolved this phase.
- Guard-test self-deletion resistance (D-10-20) — accepted residual for a
  test-level mechanism.
- `.planning/spikes` registry maintenance (D-08-43 spike 006 row) — still
  un-owned, carried forward.

</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SEM-08 | Interpreter executes calls on a bounded call stack with a documented fixed ceiling; exceeding it is a named refusal, not a host stack overflow | `callgraph.Order`'s iterative explicit-stack DFS (`callgraph.go:283-400`) is the exact shape to mirror for `interp`'s `[]frame` stack. `MaxCallDepth = 128` sits below the real structural ceiling (`maxFunctions = 1024`, `syntax/parser.go:16`), unlike `pathoracle.MaxPaths` which sits above its ceiling — direction is inverted, must be stated in the rationale comment (D-10-23). Refusal modeled as an `Outcome`/`Execution` value (like `OpDefect`), not a bare Go `error` (D-10-25), because Phase 11's five-axis comparator needs it as comparable data. |
| SEM-09 | Drop/cleanup obligations run in defined order on normal return and every nonlocal exit across a call boundary | `core.go:559-564` documents `OpCall` as never a terminator; the language has no exceptions/unwinding, so the nonlocal-exit surface is narrow: the existing single-frame foreign landing pad (`interp.go:369`) extended to N frames, plus the SEM-08 depth refusal if modeled as abort (D-10-31). Order: LIFO across frames, reverse-acquisition within a frame (existing single-frame rule unchanged) (D-10-33). Reuse `execution.Event.FunctionID` (`execution.go:45`) for frame attribution — no schema change (D-10-32). Must be observed via `interp.CanonicalBytes`/`execution.Equal`, never internal map state, via a `frameDrainOrderForTest` seam mirroring `opCallGroupedArmForTest` (D-10-35). |
| TRU-02 | `originvalidate` extends its published-origin walk across `OpCall`, mirroring the proven `OpForeignCall` hop | `walkReturnOrigin` (`originvalidate.go:171-218`) switches on `operation.Kind` with cases for `OpBorrowExclusive`/`OpBorrowShared`/`OpForeignCall`, no `case core.OpCall` — confirmed absent by direct read. Mirror is real but not exact: `OpForeignCall`'s contract lives on the current function (no lookup); `OpCall`'s lives on a different function (requires the new callee-contract map) — state this difference in the plan's doc comment (D-10-07). |
| TRU-03 | `pathoracle` independently re-derives the cross-function loan-chain rule without importing `check` or `corevalidate` | Compose, don't contract-hop (D-10-09). Package doc's real claim is "never converges, never closes a relation, only enumerates and replays concrete paths" (D-10-11) — write this precisely; composition preserves it, a contract hop would not. New composition-depth cap needs its own typed refusal (mirroring `pathCapError`, `pathoracle.go:87-110`), separate from `MaxPaths` (D-10-12). |
| QLT-04 | Cross-function loan-endpoint differentials rebuild M001 Phase 3's exhaustive endpoint enumeration at a declared, bounded composition depth | Depth = 3 (D-10-46). Definition: number of `OpCall` hops a loan/origin carry relation crosses (already load-bearing at `corevalidate_peer_liveness.go:64-77`) (D-10-45). No depth-3 fixture exists in `testdata/` today (confirmed via grep — zero hits) (D-10-47). New fixture must vary which hop carries the borrow, not repeat `relay_depth2_accept.lang`'s owned-passthrough shape (D-10-48, confirmed by direct read: `leaf` returns owned `Buffer`, no borrow). Declared bidirectionally: a Go const plus a test that fails if the corpus stops reaching the depth AND if someone raises the constant without extending the corpus (D-10-50). |
| OWN-05b | Same call-site ownership-transfer fact derived independently by `interp` | Confirmed: `interp` imports `corevalidate` and calls `corevalidate.Validate` at `Run`'s first line (`interp.go:38`), but never reads `PeerSignatures()`/`PeerSiteCoverage()` (grepped, zero hits) — true by omission, needs a guard test making it a mechanism (D-10-36). Claim must be stated narrowly: independence of derivation mechanism nested inside a shared validation dependency, not mutual non-import independence (D-10-37). The derivation IS the observable frame-partition behavior — moved-from place deleted, callee frame seeded only from the transferred value — not a `Mode`-string inspection (D-10-38, confirmed: `runBranchArm`'s `OpMove` at `interp.go` does `delete(values, operation.SourceID)`). |

</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Interprocedural call execution (frame stack, drop ordering) | Interpreter (`interp`) | Core IR (`core.OpCall` shape) | `interp` is the deterministic semantic oracle (STATE.md standing commitment); it owns runtime execution semantics, `core` owns the static operation shape it dispatches on. |
| Published-origin re-derivation across calls | Admission / static analysis (`originvalidate`) | Core IR (`core.FunctionSignature`/`ReturnContract`) | `originvalidate` is a trust-boundary re-deriver consuming only the checked-core artifact, never a callee body — mirrors `corevalidate`'s existing tier placement. |
| Loan-chain path re-derivation across calls | Admission / static analysis (`pathoracle`) | Core IR (same) | Same tier as `originvalidate` — an independent oracle over `core.Function`'s own Blocks/Edges/Operations, never touching `check`/`corevalidate`. |
| Import-boundary enforcement (mechanism, not convention) | Build/test tooling (`go test` + `go list -deps`) | CI (`.github/workflows/ci.yml`) | Criterion 1 requires "build- or test-level," not runtime; `go test` already fails CI's `checks` job (`ci.yml:64`), so no new machinery is needed — extend the existing guard tests' scope and depth. |
| Cross-peer differential (four-way on accept, three-way on refuse) | Test/verification tier (`session` package test) | All four production tiers (`check`, `corevalidate`, `pathoracle`, `interp`) | The differential is a verification-layer consumer of four independently-implemented production deriver, never itself a fifth derivation — it must only compare, never compute the fact a second way. |

## Standard Stack

This phase adds **no external dependencies**. `go.mod` declares zero
non-stdlib imports `[VERIFIED: go.mod]` and every new subsystem (frame
stack, composition cap, transitive import guard) is built from Go's
standard library (`go/parser`, `go/token`, `os/exec` for the stack-overflow
subprocess probe, `runtime/debug.SetMaxStack`) plus this repository's own
established internal packages (`core`, `execution`, `callgraph`,
`corevalidate`, `testsupport`).

### Core (existing internal packages this phase extends)
| Package | Purpose | Why this one |
|---------|---------|---------------|
| `internal/compiler/interp` | Deterministic interpreter — grows a `[]frame` stack | Already the project's declared semantic oracle (STATE.md) |
| `internal/compiler/originvalidate` | Published-origin re-derivation — grows an `OpCall` case | Existing TRU-02 owner, `OpForeignCall` precedent already proven |
| `internal/compiler/pathoracle` | Path-enumeration oracle — grows call composition | Existing TRU-03 owner, `EnumeratePaths`/`MaxPaths`/`pathCapError` precedent already proven |
| `internal/compiler/corevalidate` | Peer re-deriver — supplies the `peerLoanCarry` threading pattern to mirror | Already shipped the D-09-03 fix for the structurally identical widening problem |
| `internal/compiler/callgraph` | Supplies the iterative explicit-stack DFS shape (`Order`, `callgraph.go:283-400`) | The only in-repo precedent for "traverse without consuming Go stack"; its own doc comment states the rationale (lines 25-26) |
| `internal/compiler/execution` | `Event`/`Execution`/`Outcome` schema `interp` must not silently change | Phase 11's five-axis-comparator contract; `FunctionID` field (`execution.go:45`) already exists and is reused, not extended |
| `internal/compiler/session` | Hosts `session_peer_gate_test.go`, the differential to extend | D-09-23 precedent: extending an existing gate beats building a second one |

### Supporting (Go stdlib, no new go.mod entries)
| Package | Purpose | When to Use |
|---------|---------|-------------|
| `go/parser`, `go/token` (`ImportsOnly` mode) | Existing direct-import guard mechanism | Already used by `TestOriginValidatorImportsStayIndependent` et al.; keep for the direct-scan half |
| `os/exec` | Subprocess re-exec for the Pitfall-4 native-stack-overflow probe | New this phase — no subprocess pattern exists anywhere in the repo today (`[VERIFIED: grep -rn "os/exec" internal/` returns zero hits]`); Go's standard `TestHelperProcess` idiom is the correct new pattern (D-10-42) |
| `runtime/debug.SetMaxStack` | Pins a small deterministic host stack ceiling in the probe's child process | Needed so the probe's pass/fail is platform-independent (D-10-42) |
| `go list -deps` (via `os/exec` from the test, or a `go:generate`-free direct shell invocation in the test) | Transitive import-guard hardening | Replaces/augments the direct-scan guard so a helper package that itself imports `check`/`corevalidate` is caught (D-10-17) |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Go native recursion for call frames | `[]frame` heap stack (chosen) | Recursion is **disqualified**, not merely worse: Go's stack-exhaustion `throw` is fatal and `recover()`-proof (verified against documented Go runtime behavior; multiple `golang/go` issue reports agree — `[CITED: golang/go issue tracker, general knowledge of runtime.throw semantics]` `[ASSUMED: no specific issue number verified this session]`) |
| A `depguard`/custom `go/analysis` linter for import control | `go test` + `go/parser`/`go list -deps` (chosen) | Criterion 1 only requires "build- or test-level"; CI already fails on `go test` (`ci.yml:64`); adding new lint infrastructure to a repo with zero prior linter tooling is unjustified machinery (D-10-18) |
| A contract-hop for `pathoracle`'s `OpCall` case | Recursive composition (chosen) | A third contract-hop collapses TRU-03's "structurally distinct" claim exactly at call boundaries — the one place independence is being tested (D-10-10/11) |
| A depth field on `execution.Event` | Reuse existing `FunctionID` (chosen) | A depth field manufactures guaranteed false divergences once Phase 11's `-O3 -flto` build inlines a callee and no runtime frame exists for native to reproduce a depth from (D-10-32) |
| A second differential harness for criterion 4 | Extend `session_peer_gate_test.go` (chosen) | D-09-23 already ratified this tradeoff for the same reason: two gates that can drift on what divergence means recreate the exact "two derivations, one truth" failure the phase exists to prevent |

**Installation:** none — no `go get`/`npm install` step this phase.
Verify the module still resolves cleanly before starting: `go build ./...`
and `go vet ./...` (both already in CI, `ci.yml:62-63`).

## Package Legitimacy Audit

**Not applicable.** This phase installs zero external packages —
`go.mod` (`[VERIFIED: go.mod, read this session]`) declares only `module
github.com/codename-lang/lang` and `go 1.24`, with no `require` block.
Every new symbol this phase adds lives in existing internal packages or
Go's standard library. The Package Legitimacy Gate protocol is skipped
per its own trigger condition ("whenever this phase installs external
packages").

## Architecture Patterns

### System Architecture Diagram

```
                         ┌─────────────────────────────┐
                         │   syntax.Parse (untrusted)   │
                         └──────────────┬───────────────┘
                                        ▼
                         ┌─────────────────────────────┐
                         │   check.Program (producer)   │  <- Site 1
                         │  builds core.Program w/OpCall │
                         └──────────────┬───────────────┘
                                        │  core.Program (checked-core artifact,
                                        │  callee bodies NEVER re-crossed)
              ┌─────────────────────────┼──────────────────────────┐
              ▼                         ▼                          ▼
   ┌────────────────────┐   ┌────────────────────┐     ┌────────────────────┐
   │ corevalidate.Validate│   │ originvalidate.     │     │  pathoracle.        │
   │  (Site 2, peer re-  │   │ ValidatePublished    │     │  EnumeratePaths /   │
   │  derivation)        │   │  (Site 3 — TRU-02)   │     │  RecomputeEndpoints │
   │                      │   │  NEW: OpCall case    │     │  (Site 6 — TRU-03)  │
   │  peerLoanCarry map   │   │  widens callee-      │     │  NEW: OpCall        │
   │  (shipped precedent  │   │  contract map,       │     │  composition,       │
   │  D-09-03)            │   │  walkReturnOrigin    │     │  recurses into own  │
   │                      │   │  gains case core.OpCall│    │  enumerator,        │
   │                      │   │  (mirrors            │     │  splices callee     │
   │                      │   │  OpForeignCall hop)  │     │  paths, own         │
   │                      │   │                      │     │  composition-depth  │
   │                      │   │                      │     │  cap                │
   └──────────┬───────────┘   └──────────┬───────────┘     └──────────┬───────────┘
              │                          │                            │
              │           (neither `originvalidate` nor `pathoracle`  │
              │            imports `check` or `corevalidate` —        │
              │            enforced by transitive go-list-deps guard) │
              └──────────────────────────┴────────────────────────────┘
                                        │
                                        ▼
                         ┌─────────────────────────────┐
                         │       interp.Run (Site 4)    │  <- deterministic oracle
                         │  NEW: []frame call stack,    │
                         │  MaxCallDepth=128 refusal,    │
                         │  per-frame values map,        │
                         │  LIFO drop drain across frames│
                         │  (imports corevalidate.Validate│
                         │   as fail-closed precondition, │
                         │   never reads its ownership    │
                         │   fields — OWN-05b guard)       │
                         └──────────────┬───────────────┘
                                        │  execution.Execution
                                        │  (Schema0, FunctionID-tagged events)
                                        ▼
              ┌─────────────────────────────────────────────────┐
              │  session_peer_gate_test.go's extended differential│
              │  (Criterion 4): three-way on refuse               │
              │  (check/corevalidate/pathoracle endpoints),        │
              │  four-way on accept (+ interp's metamorphic         │
              │  event-order comparison) — at declared depth-3      │
              │  composition corpus                                 │
              └─────────────────────────────────────────────────┘
```

### Recommended Project Structure

No new packages/directories — every change lands inside existing package
boundaries:

```
internal/compiler/
├── interp/
│   ├── interp.go              # NEW: frame type, []frame stack, MaxCallDepth,
│   │                           #      OpCall arms replaced at all 3 exec paths
│   └── interp_test.go         # NEW: depth-exceeded, drop-order, mutant-kill tests
├── originvalidate/
│   ├── originvalidate.go      # NEW: callee-contract map param, case core.OpCall
│   └── originvalidate_test.go # NEW: transitive-guard hardening, corevalidate forbidden
├── pathoracle/
│   ├── pathoracle.go          # NEW: composition-depth cap + typed error (or sibling file)
│   └── pathoracle_test.go     # NEW: transitive-guard hardening, discriminating fixture
├── corevalidate/
│   └── corevalidate.go        # NEW: RecomputeEndpoints exported accessor (D-10-52)
├── session/
│   └── session_peer_gate_test.go  # EXTENDED: three/four-way corpus walker, accountable
│                                   #           peerDivergenceExpected struct
testdata/phase10/
├── relay_depth3_*.lang        # NEW: depth-3 fixtures, borrow varied per hop (D-10-48)
├── call_depth_exceeded.lang or generated 129-function chain (D-10-24)
└── interp_oracle/             # NEW: golden corpus for the stability freeze (D-10-57)
```

### Pattern 1: Narrow-map threading for cross-function widening (`originvalidate`)
**What:** Precompute a `map[calleeID]<minimal fact>` once at the top-level
entry point, thread it as an explicit parameter through the call chain that
needs it — never widen to `core.Program`.
**When to use:** Any re-deriver that needs one fact about a *different*
function without gaining body access to that function.
**Example (the shipped precedent to mirror):**
```go
// Source: internal/compiler/corevalidate/corevalidate.go:1192,1203 (read this session)
func buildLoanChainIndex(operations []core.LinearOperation, checks *int, loanCarry map[string]peerLoanCarryFact) *loanChainIndex {
    // ...
    if operation.Kind == core.OpCall && !disablePeerLoanCarryConsultForTest &&
        !forcePeerLoanCarryTrueForTest && !loanCarry[operation.CalleeID].ReturnsBorrowOfParam {
        continue // callee's own declared contract does not carry
    }
}
```
`originvalidate`'s widening must follow this exact shape: build the map once
in `ValidatePublished`, thread it into `RecomputeOriginPerReturn` →
`walkReturnOrigin`, consult via `map[operation.CalleeID]`.

### Pattern 2: Iterative explicit-stack traversal (never native Go recursion)
**What:** An explicit slice-backed stack (`[]stackFrame` / `[]frame`) with a
`for { ... }` driver loop, never a recursive function call.
**When to use:** Any traversal or execution model whose depth is
user/program-controlled and must be bounded by a *named refusal*, not a Go
runtime crash.
**Example (the shipped precedent):**
```go
// Source: internal/compiler/callgraph/callgraph.go:283-292 (read this session)
type stackFrame struct {
    id             string
    nextChildIndex int
}
func Order(program core.Program) ([]string, error) {
    // ...
    stack := []stackFrame{{id: root}}
    for len(stack) > 0 {
        top := &stack[len(stack)-1]
        // push/pop mutate `stack` directly — no Go call-stack growth
    }
}
```
`interp`'s frame stack follows this shape but stores execution state
(`values map[string]string`, live-resource tracking) per element, not just
traversal position.

### Pattern 3: Typed, capped, fail-closed refusal (never truncation)
**What:** A private error type carrying a stable `Code()` method, returned
instead of silently truncating or looping forever.
**When to use:** Every new bound this phase introduces (`MaxCallDepth`,
`pathoracle`'s composition-depth cap).
**Example:**
```go
// Source: internal/compiler/pathoracle/pathoracle.go:87-110 (read this session)
type pathCapError struct {
    functionID string
    limit      int
}
func (e *pathCapError) Error() string {
    return fmt.Sprintf("pathoracle.path_count_exceeded: function %q exceeds the declared cap of %d acyclic entry-to-return paths", e.functionID, e.limit)
}
func (e *pathCapError) Code() string { return "pathoracle.path_count_exceeded" }
```

### Pattern 4: Nil-default unexported injection seam (fault-injection, never production-mutable)
**What:** An unexported package-level `var xOverride func() T` (or bool),
defaulting to nil/false in production, flipped only by a same-package test.
**When to use:** Every mutation-kill/discriminating-fixture obligation this
phase adds (D-10-14, D-10-35, D-10-41, D-10-42/43, D-10-55).
**Example:**
```go
// Source: internal/compiler/pathoracle/pathoracle.go:59-64 (read this session)
var TerminatorKindsOverride func() []core.OperationKind
func recognizedTerminatorKinds() []core.OperationKind {
    if TerminatorKindsOverride != nil {
        return TerminatorKindsOverride()
    }
    return core.TerminatorKinds()
}
```
```go
// Source: internal/compiler/interp/interp.go:23-31 (read this session)
var opCallGroupedArmForTest = false
```

### Anti-Patterns to Avoid
- **Passing `core.Program` down into a re-deriver's walk:** makes a callee's
  *body* reachable from inside a walk whose whole contract is body-blindness
  — the guarantee degrades from a type-signature property to a comment a
  future editor can silently violate (D-10-02).
- **A contract-hop in `pathoracle` at `OpCall`:** collapses TRU-03's
  independence claim exactly where it is being tested (D-10-10).
- **Recursion for the interpreter's call stack:** architecturally
  incapable of a named refusal — Go's `throw` on stack exhaustion cannot be
  caught (D-10-21).
- **Consolidating the four (soon five+) near-identical import guards into
  one shared table:** the existing redundancy means multiple deletions are
  needed to blind the mechanism; that redundancy is a real defense, not
  duplication debt, for a guard test (D-10-18).
- **Ranging a Go map to produce ordered output:** `interp`'s determinism
  depends on producing events from ordered slices (`liveOrder`, block walk)
  never a bare `range` over a map — must be asserted explicitly for new
  per-frame code, not inherited by inspection (D-10-26).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Bounded-depth traversal without consuming host stack | A custom recursion-depth counter with `recover()` | `[]frame`/`[]stackFrame` explicit stack (callgraph.Order's shape) | `recover()` cannot catch Go's stack-overflow `throw` — this is not a style preference, it is a hard platform constraint (D-10-21) |
| Transitive import boundary enforcement | A hand-rolled AST import-graph walker | `go list -deps` (stdlib toolchain command, already available in CI) | No new dependency, and avoids re-implementing what the Go toolchain already computes correctly (D-10-17) |
| Cross-function fact propagation ordering | A worklist/fixpoint loop for `pathoracle`'s composition | Recursion through `pathoracle`'s own `EnumeratePaths`, guarded by `callgraph`'s pre-existing cycle refusal | The call graph is already a proven-acyclic DAG by the time `pathoracle` runs; composition needs only its own defense-in-depth cycle guard (D-10-15), not a new convergence algorithm — that would reintroduce exactly the "join/reduction step" TRU-03's independence claim depends on avoiding |
| Stack-overflow-vs-language-cap distinction proof | An in-process headroom-ratio measurement | A subprocess probe using Go's `TestHelperProcess` idiom + `debug.SetMaxStack` | An in-process measurement is "an assertion wearing observation's clothes" — rejected as gate evidence, usable only as a supplementary sanity check (D-10-42) |

**Key insight:** every "don't hand-roll" item in this phase is not about
avoiding a third-party library (there are none) — it is about not
re-inventing a pattern this specific codebase has already proven correct
once. The research task was locating the precedent, not designing a new
mechanism.

## Common Pitfalls

### Pitfall 1: Treating the `OpForeignCall`/`OpCall` mirror as exact
**What goes wrong:** A plan or doc comment claims `originvalidate`'s new
`OpCall` case is "the same mechanism" as `OpForeignCall` without noting that
`OpForeignCall`'s contract lives on the *current* function
(`function.ForeignContract`, zero lookup) while `OpCall`'s lives on a
*different* function and requires a new cross-function map.
**Why it happens:** Both hops read a declared contract rather than
re-walking a body, so the surface similarity is real and easy to overstate.
**How to avoid:** State the difference explicitly in the widening's doc
comment (D-10-07).
**Warning signs:** A review comment or plan section that says "mirrors
OpForeignCall" with no qualifier.

### Pitfall 2: Recursing in Go for the interpreter's call stack
**What goes wrong:** A natural first implementation calls
`interp.execFunction` recursively at each `OpCall`. This compiles, passes
small tests, and then either (a) never triggers the depth refusal because
the real ceiling is Go's host stack, or (b) crashes the whole test binary
with an uncatchable `runtime: goroutine stack exceeds ... - fatal error:
stack overflow` the moment a legitimately deep (but still-language-legal)
chain is exercised.
**Why it happens:** Go recursion is the path of least resistance and
"looks like" every other language's call-stack model.
**How to avoid:** `[]frame` heap stack + iterative dispatch loop, mirroring
`callgraph.Order`, from the first commit (D-10-21/22).
**Warning signs:** Any `func (i *interp) execFunction(...)` that calls
itself; any depth test that only exercises small N.

### Pitfall 3: `MaxCallDepth`'s direction inverted from `MaxPaths`'s
**What goes wrong:** A reader pattern-matches `MaxCallDepth` to
`pathoracle.MaxPaths` (which sits comfortably *above* its real reachable
maximum so it never fires) and sets `MaxCallDepth` similarly high, making
the SEM-08 refusal decorative — never actually reachable by a legal
program.
**Why it happens:** Both are "declared bounds with a rationale comment," so
they look like the same pattern.
**How to avoid:** `MaxCallDepth` must sit **below** the real structural
ceiling (1024, from `maxFunctions`) so a legitimately-admitted 129-function
chain actually triggers it (D-10-23). Write the inversion into the
constant's own rationale comment.
**Warning signs:** A `MaxCallDepth` value ≥ 1024, or a comment that reads
like `MaxPaths`'s "well above the real maximum" language without noting the
opposite direction.

### Pitfall 4: Proving the stack-overflow-vs-language-cap distinction with an in-process measurement
**What goes wrong:** A plan writes a test that measures Go's remaining
stack headroom in-process (e.g. via a recursive probe function) and asserts
a ratio, presenting it as proof the two limits are unrelated.
**Why it happens:** It's simpler to write than a subprocess harness and
"looks like" a stress test.
**How to avoid:** Use a re-exec'd subprocess (`os/exec` + Go's
`TestHelperProcess` idiom, new to this repo — confirmed no existing
subprocess/`SetMaxStack`/`TestMain` pattern via
`grep -rn "os/exec" internal/` returning zero hits), pin the child's host
stack ceiling deterministically via `runtime/debug.SetMaxStack`, and assert
on exit status/stderr from the parent (D-10-42). Frame the honest result as
"O(1) host stack per language call depth — the two limits are structurally
unrelated," not a near-miss safety ratio (D-10-43). Fix the threshold
constant in source **before** the first green run (D-10-44).
**Warning signs:** Any stack-depth assertion that runs in the same process
as the test binary; a threshold that reads as chosen after seeing a passing
number.

### Pitfall 5: Overclaiming criterion 4 as unqualified "four-way"
**What goes wrong:** A plan, commit message, or the shipped artifact
describes the differential as "four engines agree" without the
refuse/accept qualifier, because `interp` genuinely cannot produce a
`core.LoanEndpoint` and only participates on the accept side.
**Why it happens:** "Four-way differential" is a shorter, cleaner phrase.
**How to avoid:** Every artifact states "three-way on refuse, four-way on
accept" verbatim (D-10-53), following the OWN-05a/05b split precedent
(D-09-37) for not overclaiming coverage that doesn't exist.
**Warning signs:** Plan titles or PLAN.md success criteria using "four-way"
with no qualifier.

### Pitfall 6: Applying the D-09-53 code change instead of the artifact correction
**What goes wrong:** A plan re-reads PHASE-09-DEBT.md's D-09-53 at face
value and applies the literal proposed narrowing to
`deriveFunctionUsesParam`'s default arm, believing it closes real debt.
**Why it happens:** The debt register reads as a straightforward, already-
scoped fix.
**How to avoid:** Re-run the throwaway patch as a pre-flight (exclude
`core.OpMove`/`core.OpCopy` from the default arm at `check.go:729-730`,
run the full suite) and confirm the 8 named failures — including
`TestInterproceduralLivenessTwinPatternB`, the very test the register cited
as proof the law was unaffected — before writing the correction instead
(D-10-27/30). Deliverable is a doc-comment/fixture-header correction plus a
newly-filed, correctly-scoped debt item (D-10-28/29), not a semantic code
change.
**Warning signs:** Any diff touching `check.go:723-730`'s default arm this
phase.

## Code Examples

### The `OpCall` case `originvalidate` needs to add (shape, not literal diff)
```go
// Source: internal/compiler/originvalidate/originvalidate.go:171-218 (read this session)
// walkReturnOrigin's existing switch — add a case core.OpCall here, consulting
// the new narrow callee-contract map (threaded in, never core.Program):
switch operation.Kind {
case core.OpBorrowExclusive:
    if derivedAccess == "" { derivedAccess = "exclusive" }
case core.OpBorrowShared:
    if derivedAccess == "" { derivedAccess = "shared" }
case core.OpForeignCall:
    if derivedAccess == "" && function.ForeignContract != nil {
        switch function.ForeignContract.Alias {
        case "borrow": derivedAccess = "shared"
        case "retain": derivedAccess = "exclusive"
        }
    }
// NEW: case core.OpCall:
//   if derivedAccess == "" {
//       if fact, ok := calleeContracts[operation.CalleeID]; ok {
//           derivedAccess = fact.Access // minimal value type, D-10-04
//       }
//   }
}
```

### The transitive import guard (new shape, extending the existing direct-scan guard)
```go
// Existing shape (direct-import scan), Source: internal/compiler/originvalidate/originvalidate_test.go:33-55 (read this session)
func TestOriginValidatorImportsStayIndependent(t *testing.T) {
    // ... go/parser ImportsOnly scan over the package's own .go files,
    // checking each import's path suffix against a forbidden list.
    // MUST be extended to also forbid "/compiler/corevalidate" (D-10-16),
    // AND supplemented with a `go list -deps` transitive check (D-10-17)
    // so a helper package that itself imports check/corevalidate is caught.
}
```

### The `Run` entry point's existing single-frame map, becoming the stack base
```go
// Source: internal/compiler/interp/interp.go:37-50 (read this session)
func Run(program core.Program, functionName, input string) (Execution, error) {
    validated := corevalidate.Validate(program) // fail-closed precondition; OWN-05b
                                                  // guard: never read validated's
                                                  // ownership-bearing fields (D-10-36)
    if !validated.Valid {
        return Execution{}, fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
    }
    program = validated.Program()
    function, ok := findFunction(program, functionName)
    // ... existing single-function dispatch to runLinear/runBranchArm.
    // NEW: this becomes frame[0] of an explicit []frame stack; OpCall
    // pushes frame[n+1] instead of recursing.
}
```

### The three `OpCall` arms today (all three currently return `ErrCallUnsupported`)
```go
// Source: internal/compiler/interp/interp.go:156,341,454 (read this session)
// runBranchArm (line 156), runLinearBlocks (line 341), runLinear (line 454)
// each have:
case core.OpCall:
    // D-07-39: recognized, never faked. See ErrCallUnsupported.
    return Execution{}, fmt.Errorf("operation %q: %w", operation.ID, ErrCallUnsupported)
// All three must gain real push-frame-and-continue behavior this phase,
// sharing ONE small in-package frame-partition helper (D-10-39) — this is
// allowed under D-09-38 because all three arms are the same peer/package/
// derivation method, not three cross-peer reuses.
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|---------------|--------|
| `interp.Run` looks up exactly one function by name, no cross-function execution | `interp.Run` executes a multi-function call graph on a bounded `[]frame` stack | This phase (SEM-08) | `interp` becomes the trusted oracle Phase 11's five-axis comparator differentials against — the single highest-leverage change in the phase |
| `originvalidate` treats `OpCall` as fully transparent (walks straight through) | `originvalidate` re-derives origin facts across `OpCall` via a callee-contract map | This phase (TRU-02, closes D-09-51) | Closes the last silent-transparency gap in the published-origin re-deriver; `twin_a_accept.lang` moves from refusal to clean admit through the full CLI |
| `pathoracle` only enumerates within one function | `pathoracle` composes callee enumerations recursively across `OpCall` | This phase (TRU-03) | Third, genuinely structurally-distinct interprocedural decision procedure — makes the criterion-4 differential meaningful rather than three peers agreeing by construction |

**Deprecated/outdated:**
- D-09-53's original diagnosis (`deriveFunctionUsesParam`'s default arm as a
  defect) is superseded — the diagnosis is falsified by direct execution;
  what's actually wrong is the doc comment's overstated wording (D-10-27).

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Go's `runtime.throw` on stack exhaustion is fatal and cannot be recovered via `recover()` (no specific `golang/go` issue number verified this session — this is documented, stable Go runtime behavior but was not re-confirmed against a live crash reproduction in this research pass) | Standard Stack / Pitfall 2 | Low — this is long-standing, widely-documented Go runtime behavior (stack overflow is one of the few `recover()`-proof fatal errors alongside out-of-memory); if somehow wrong, the consequence is merely that recursion becomes viable, which only *simplifies* the phase, it does not block it. CONTEXT.md's D-10-21 states this was "verified against the Go runtime's behavior and multiple golang/go issue reports" during the discuss-phase advisor pass — this research pass did not independently re-run a crash reproduction to confirm it, so it is downgraded to `[ASSUMED]` here pending that direct falsification. |

**If a stronger confirmation is wanted before planning:** a two-line Go
program that recurses unboundedly inside a `recover()`-wrapped goroutine,
run once, would convert this to `[VERIFIED]` — cheap, and worth doing as
part of Wave 0 tooling validation given how load-bearing D-10-21 is (it is
marked "one-way — the frame model is what Phase 11's five-axis comparator
differentials against").

## Open Questions

1. **Exact value type for `originvalidate`'s callee-contract map (D-10-04)**
   - What we know: must carry only the `Access`/derived-ness pair the walk
     needs, never the full `core.PublicOrigin` (which additionally carries
     `Paths []string` this walk doesn't need).
   - What's unclear: exact field names/shape — left to planner discretion
     per CONTEXT.md.
   - Recommendation: mirror `peerLoanCarryFact`'s one-field-struct shape
     (`corevalidate_peer_liveness.go:38-39`) for consistency with the
     pattern it's mirroring.

2. **Landing phase for the re-filed `corevalidate` `usesParam` peer
   (D-10-28)**
   - What we know: it is real, newly-identified debt, Key Lesson 2's exact
     exposure (only one detector watches the fact).
   - What's unclear: whether to propose Phase 11 (adjacent to the native
     lowering work already touching `check`) or leave landing phase open.
   - Recommendation: leave open in `PHASE-10-DEBT.md`, per CONTEXT.md's
     explicit discretion grant — do not assign under time pressure.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | All of this phase | Yes `[VERIFIED: go.mod declares go 1.24; confirmed by reading go.mod this session]` | 1.24 | — |
| `go list` (ships with toolchain) | Transitive import-guard hardening (D-10-17) | Yes — part of the Go toolchain, no separate install | matches Go toolchain version | — |
| CI runners (Linux + macOS) | Verifying the Pitfall-4 subprocess probe is platform-independent | Yes `[VERIFIED: .github/workflows/ci.yml, matrix runs-on line 46, read this session]` | — | — |

No missing dependencies. No blocking items.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go's built-in `testing` package (no third-party test framework anywhere in the repo) |
| Config file | none — plain `go test` |
| Quick run command | `go test ./internal/compiler/interp/... ./internal/compiler/originvalidate/... ./internal/compiler/pathoracle/... ./internal/compiler/corevalidate/...` |
| Full suite command | `go test ./...` (and `go test -race ./...`, both already CI-gated, `ci.yml:64-65`) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| SEM-08 | Depth-exceeded refusal fires through the real pipeline on a genuine 129-function chain | integration | `go test ./internal/compiler/interp/... -run TestCallDepthExceeded -v` | ❌ Wave 0 — build the 129-function `.lang` fixture and test, following `checkedCallBasicProgram`'s pattern (`interp_test.go:30`) |
| SEM-08 (Pitfall 4 gate) | Native-stack-overflow probe distinct from language-level bound | subprocess integration | `go test ./internal/compiler/interp/... -run TestNativeStackHeadroomIndependentOfCallDepth -v` | ❌ Wave 0 — new `TestHelperProcess`-shaped subprocess test |
| SEM-09 | Drop order observed via canonical bytes, both normal-return and every nonlocal exit | unit + differential | `go test ./internal/compiler/interp/... -run TestFrameDrainOrder -v` | ❌ Wave 0 — needs `frameDrainOrderForTest` seam (mirrors `opCallGroupedArmForTest`) |
| TRU-02 | `walkReturnOrigin` gains `case core.OpCall`; `twin_a_accept.lang` admits clean through full CLI | unit + CLI gate | `go test ./internal/compiler/originvalidate/... -run TestOpCallOriginWalk -v` then `go test ./internal/compiler/session/... -run TestCheckCommandFile` (existing CLI harness) | ❌ Wave 0 test; existing CLI harness ✅ |
| TRU-02 (import guard hardening) | `originvalidate` transitively imports neither `check` nor `corevalidate` | build/test guard | `go test ./internal/compiler/originvalidate/... -run TestOriginValidat.*Imports.*Independent -v` | ✅ exists, needs extension (add `corevalidate` to forbidden list + transitive scan) |
| TRU-03 | Composition splits caller's endpoints per-path depending on which callee path is spliced (discriminating fixture) | unit | `go test ./internal/compiler/pathoracle/... -run TestCompositionDiscriminatesPerPathBorrow -v` | ❌ Wave 0 — required deliverable fixture per D-10-14 |
| QLT-04 | Depth-3 composition corpus reaches the declared bound, bidirectionally checked | corpus + bidirectional gate | `go test ./internal/compiler/session/... -run TestCompositionDepthCorpusReachesDeclaredBound -v` | ❌ Wave 0 — new depth-3 fixtures + test |
| OWN-05b | `interp` never reads `corevalidate.Result`'s ownership-bearing fields | guard | `go test ./internal/compiler/interp/... -run TestInterpDoesNotReadCorevalidateOwnershipFields -v` | ❌ Wave 0 |
| Criterion 4 | Three-way-on-refuse, four-way-on-accept differential, seeded faults per peer pair | differential + mutation-kill | `go test ./internal/compiler/session/... -run TestNoUndeclaredCheckPeerDivergenceAcrossCorpus -v` (extended) | ✅ exists (`session_peer_gate_test.go:226`), needs extension to 3rd/4th peer |
| Stability freeze | Golden corpus byte-for-byte, deterministic across `-count=10` runs | golden + flake check | `go test ./internal/compiler/interp/... -run TestInterpOracleGoldenCorpus -v` and `-run TestInterpDeterministicAcrossRuns -count=10` | ❌ Wave 0 — new `testdata/phase10/interp_oracle/` golden corpus |

### Sampling Rate
- **Per task commit:** targeted package test (`go test ./internal/compiler/<pkg>/...`)
- **Per wave merge:** `go test ./...` (full suite, matching CI's `checks` job)
- **Phase gate:** `go test ./...` and `go test -race ./...` green before
  `/gsd-verify-work`, matching `ci.yml:64-65` exactly

### Wave 0 Gaps
- [ ] `internal/compiler/interp/interp_test.go` — depth-exceeded (through
      real pipeline, D-10-24), drop-order (D-10-35 seam), native-stack-probe
      (D-10-42), mutant-kill pairs for OWN-05b (D-10-41)
- [ ] `internal/compiler/originvalidate/originvalidate_test.go` — `OpCall`
      widening test, `corevalidate`-forbidden extension, transitive-guard
      upgrade, D-10-08's `TerminatorKindsOverride`-shaped seam for the gate
- [ ] `internal/compiler/pathoracle/pathoracle_test.go` — composition test,
      discriminating per-path-borrow fixture (D-10-14, required deliverable),
      transitive-guard upgrade
- [ ] `testdata/phase10/` — depth-3 fixture pair varying which hop carries
      the borrow (D-10-48), the 129-function depth-exceeded fixture (or
      generated), `interp_oracle/` golden corpus
- [ ] `internal/compiler/session/session_peer_gate_test.go` — extended to
      three/four peers, `peerDivergenceExpected` upgraded to an accountable
      struct with debt-register ID + landing phase (D-10-54)
- [ ] Framework install: none — `testing` stdlib only

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-------------------|
| V2 Authentication | No | N/A — compiler internals, no auth surface |
| V3 Session Management | No | N/A |
| V4 Access Control | No | N/A |
| V5 Input Validation | Yes | Fail-closed bounds on every new limit (`MaxCallDepth`, composition-depth cap), matching `pathCapError`'s house form — refuse rather than truncate; untrusted `.lang` source is already gated upstream by `syntax.Parse`'s `MaxSourceBytes`/`MaxTokens`/`maxFunctions` (`syntax/parser.go:12-19`) |
| V6 Cryptography | No | N/A — no crypto surface this phase |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|----------------------|
| Denial of service via unbounded call-graph depth or path composition | Denial of Service | `MaxCallDepth` (interp) and the new composition-depth cap (pathoracle) — both fail-closed, typed, refuse rather than hang or truncate (D-10-12, D-10-23) |
| Host process crash (uncatchable stack overflow) from a crafted deeply-chained program | Denial of Service | `[]frame` explicit heap stack makes language call depth cost O(1) host stack, structurally decoupling the two limits (D-10-21, D-10-43) |
| Silent trust-boundary bypass — a re-deriver accidentally consulting a callee body it should never read | Tampering / Elevation of Privilege (in the compiler-trust sense) | Narrow callee-contract map (never `core.Program`) as a type-signature-level guarantee, not a comment (D-10-02); transitive import guard forbidding `check`/`corevalidate`/`ast` (D-10-16/17) |

## Sources

### Primary (HIGH confidence — read directly this session)
- `go.mod` — confirms zero external dependencies
- `internal/compiler/originvalidate/originvalidate.go` (lines 160-225) —
  `walkReturnOrigin`'s switch, confirmed no `case core.OpCall`
- `internal/compiler/originvalidate/originvalidate_test.go` (lines 20-80) —
  both guard tests, confirmed forbidden list omits `corevalidate`
- `internal/compiler/callgraph/callgraph.go` (lines 1-30, 270-400) —
  `Order`'s iterative three-color DFS and its "costs no Go stack" doc
  comment
- `internal/compiler/pathoracle/pathoracle.go` (lines 1-115) — package doc,
  `MaxPaths`, `TerminatorKindsOverride`, `pathCapError`
- `internal/compiler/interp/interp.go` (lines 1-60, 140-470) — `Run`,
  `ErrCallUnsupported`, `opCallGroupedArmForTest`, all three `OpCall` arms
  (lines 156, 341, 454)
- `internal/compiler/execution/execution.go` (lines 1-60) — `Schema0`,
  `Event.FunctionID` (line 45), `Outcome`, `Execution`
- `internal/compiler/syntax/parser.go` (lines 1-25, `maxArmsPerMatch` at
  521) — `maxFunctions = 1024`
- `internal/compiler/core/core.go` (`FunctionSignature`, `ParameterContract`,
  `ReturnContract`, `OpCall`'s never-a-terminator doc, `LoanEndpoint`) —
  confirmed `Mode: "owned"` hardcoded reasoning and `TerminatorKinds()`
- `internal/compiler/corevalidate/corevalidate_peer_liveness.go` (lines
  1-80) — `peerLoanCarryFact`, `derivePeerLoanCarry`, confirmed single
  `ReturnsBorrowOfParam` field and the unverified depth-3 prose
- `internal/compiler/corevalidate/corevalidate.go` (lines 1185-1210,
  1360-1372, 2155-2170) — `buildLoanChainIndex`'s `loanCarry` consult,
  `recomputeLoanEndpoints`, hardcoded `Mode: "owned"`
- `internal/compiler/check/check.go` (lines 695-735) —
  `deriveFunctionUsesParam`'s `OpReturn`-only exemption confirmed
- `internal/compiler/session/session_peer_gate_test.go` (lines 220-282) —
  `TestNoUndeclaredCheckPeerDivergenceAcrossCorpus`'s two-peer walker,
  `peerDivergenceExpected` confirmed as a bare `map[string]string`
- `internal/compiler/corevalidate/corevalidate_endpoint_internal_test.go`
  (line 107) — confirmed `callgraph` in `corevalidate`'s own forbidden list
- `testdata/phase08/twin_a_accept.lang`, `relay_depth2_accept.lang` — read
  in full; confirmed the owned-passthrough shape D-10-48 must not repeat
- `.github/workflows/ci.yml` (lines 46-65) — confirmed `go vet`/`go
  build`/`go test`/`go test -race` on Linux+macOS matrix
- `.planning/research/ARCHITECTURE.md` §5 (lines 504-563) and the Stage
  5/6/7/8 build-order narrative (lines 795-884)
- `.planning/phases/10-trusted-interprocedural-oracle/10-CONTEXT.md` —
  full decision log, independently re-verified in this session
- `.planning/REQUIREMENTS.md`, `.planning/STATE.md` — requirement text and
  project history

### Secondary (MEDIUM confidence)
- none this phase — no web/docs lookups were needed; the entire domain is
  in-repo precedent

### Tertiary (LOW confidence)
- Go runtime stack-overflow/`recover()` semantics (A1 in Assumptions Log) —
  training knowledge, not re-confirmed via a live crash reproduction this
  session

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — zero external deps, fully verified via `go.mod`
- Architecture: HIGH — every pattern cited has a byte-verified shipped
  precedent in this exact repository
- Pitfalls: HIGH for all except Pitfall 2's Go-runtime claim (see A1),
  which is MEDIUM pending a direct crash-reproduction falsification

**Research date:** 2026-09-11
**Valid until:** effectively until this phase lands (the tree itself is the
source of truth and will change under active development within the phase);
30 days is a reasonable outer bound for any claim not already tied to a
specific line number.
