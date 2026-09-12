# Requirements: Codename Lang — M002 Interprocedural Semantic Spine

**Defined:** 2026-09-08
**Core Value:** Give an AI agent and a human reviewer the shortest reliable path
from intent to sound, reproducible evidence without wasting iteration time or
hiding runtime costs.

REQ-IDs continue M001's category vocabulary (`SEM`, `OWN`, `NAT`, `TRU`, `QLT`,
`EFF`, `RES`, `DX`) and its numbering.

## M002 Requirements

### Semantic Core — Calls

- [x] **SEM-04**: A Lang function can call another Lang function; `OpCall` is a
      real `core.OperationKind` handled at all six dispatch sites (`check`,
      `corevalidate`, `interp`, `cgen`, `pathoracle`, `originvalidate`) with
      both exhaustive-dispatch controls green.

- [x] **SEM-05**: A callee signature summary — extending `core.Interface` /
      `core.FunctionSignature` — is a digest-bound artifact carrying everything
      a caller needs for admission; no caller admission reads a callee body.

- [x] **SEM-06**: A call is admitted only when the callee is callable ⊆
      publishable (D-04-03); a call to a non-publishable target is refused with
      a stable diagnostic code.

- [x] **SEM-07**: The compiler constructs a call graph and refuses cycles —
      direct, mutual, and indirect (including cycles through `Result` matching)
      — with a named refusal code, never a hang.

- [x] **SEM-08**: The interpreter executes calls on a bounded call stack with a
      documented fixed ceiling; exceeding it is a named refusal, not a host
      stack overflow.

- [x] **SEM-09**: Drop and cleanup obligations run in the defined order on
      normal return and on every nonlocal exit across a call boundary.

### Ownership Across Boundaries

- [x] **OWN-05a**: Ownership transfer at a call site (move vs borrow, per the
      callee's declared parameter convention) has one meaning, derived
      independently by `check` and `corevalidate`; call-site override of a
      callee's declared convention is not expressible in source and is
      fail-closed at the core layer.

- [x] **OWN-05b**: The same call-site ownership-transfer fact is derived
      independently by `interp` — the third of OWN-05's original three
      derivers — verified there by TRU-03 and in Phase 11 by NAT-06. Split
      from OWN-05a (D-09-37) rather than left as a single row flipped Complete
      on two of three derivers, because Phase 09 ships only `check` and
      `corevalidate`; a single-row overclaim is exactly the requirement-vs-code
      failure the debt registers exist to catch.

- [x] **OWN-06**: `check` derives interprocedural loan liveness from callee
      signatures only — never by re-walking callee bodies — and terminates
      under a fail-closed iteration bound.

- [x] **OWN-07**: `corevalidate` independently re-derives the same
      interprocedural loan-liveness facts without sharing an implementation
      with `check`; a seeded endpoint-level fault makes the two peers diverge.

- [x] **OWN-08**: D-03-02 is closed — an exported borrow-derived return with no
      declared origin is refused in the interprocedural case, in both admission
      layers.

- [x] **OWN-09**: The intraprocedural loan-liveness DECISION POINT is retired in
      Phase 09, one phase after the interprocedural law landed in Phase 08 —
      not the same phase, by deliberate decision: retiring the sole decision
      point before a second, independently-implemented detector existed would
      have violated Key Lesson 2 (never fix the one detector that catches a
      class of bug without first having a second one watching). "Decision
      point," not "law," because there was never a second law: one algorithm
      (`loanLivenessFixpoint`) had two callers, one of them deliberately
      summary-blind (`computeLoanLastUses`, deleted in Phase 09) — a reader who
      searches for a second law will not find one.

### Native Lowering and Equivalence

- [x] **NAT-04**: `cgen` emits multi-function C17; `Emit` / `EmitNative` no
      longer refuse programs with more than one function.

- [x] **NAT-05**: Every aliasing or capture promise emitted at a call boundary
      (`restrict`, noalias-shaped attributes) derives from a checked fact, and
      the deriving fact is named in the emitted artifact.

- [x] **NAT-06**: Interpreter, native `-O0`, native `-O3`, and `-O3 -flto`
      produce equivalent semantic outcomes and events for the interprocedural
      corpus on the five-axis comparator.

- [x] **NAT-07**: The interprocedural `-O3`/LTO tier is proven non-inert by at
      least one engineered composition-only negative control that reproduces an
      `interpreter == -O0 != -O3` divergence and fails red before its fix.

### Independent Re-Derivation

- [x] **TRU-02**: `originvalidate` extends its published-origin walk across
      `OpCall`, mirroring the already-proven `OpForeignCall` hop.

- [x] **TRU-03**: `pathoracle` independently re-derives the cross-function
      loan-chain rule without importing `check` or `corevalidate`.

- [x] **TRU-04**: A shadow-run differential over recursion, diamond, and
      deep-chain call graphs shows zero divergence between every peer that
      derives a given interprocedural fact, before either peer ships.
      **"Recursion" is satisfied as a cycle-peer differential, not a liveness
      differential (D-09-22):** `callgraph`'s three-color DFS refuses
      `core.call_graph_cycle` before any liveness derivation runs, so a
      *liveness* differential over an admitted recursive program cannot exist
      in this language. Both `check`'s and `corevalidate`'s cycle-refusal
      layers must independently agree on refusal-or-not AND on the witness,
      for self, mutual, and indirect (3-or-more-hop) cycles. A future reader
      must not infer from "recursion" that this compiler admits recursive
      programs and go hunting for a liveness fact that structurally cannot
      exist.

### Evidence and Quality

- [x] **QLT-03**: A call-graph-shape reachability register records which
      cross-function shapes the generator actually reaches, and names shapes it
      provably does not.

- [x] **QLT-04**: Cross-function loan-endpoint differentials rebuild Phase 3's
      exhaustive endpoint enumeration at a declared, bounded composition depth.

- [x] **QLT-05**: HDD reducer output on multi-function programs is re-verified
      to reproduce the same property as its input.

- [x] **QLT-06**: No interprocedural fact is marked cacheable until a
      callee-changes-invalidates-caller regression test gates it; interprocedural
      cache keys derive from the call-graph closure, not per-unit hashes.

- [x] **QLT-07**: Nyquist validation is compliant for the loan-liveness surface,
      closing M001 Phase 3's validation debt. Closed for exactly the
      loan-liveness-scoped rows (03-03/03-04/03-05) in
      `09-VALIDATION.md`'s "M001 Phase 3 Debt Closure (loan-liveness subset)"
      section; OWN-04's rows (03-06/03-07) and the loop-carried-liveness
      exclusion stay explicitly outside this closure and
      `03-VALIDATION.md`'s own `nyquist_compliant: false` is unchanged
      (D-09-40a, D-09-41, D-09-42).

- [x] **QLT-08**: Every new interprocedural control is mutation-killed in the
      plan that introduces it — no control ships having never been seen to fail.

### Cost and Feedback Latency

- [x] **EFF-02**: Interprocedural admission cost is measured on realistic
      call-graph fan-out under the existing p50/p95/CoV protocol, stays within a
      declared bound, and is recorded in the feedback-budget manifest.

### Result Values

- [ ] **RES-02**: `Result` values with payload-carrying alternatives are
      storable and matchable; moving out of a matched payload obeys the affine
      drop obligation (D-04-30).

- [ ] **RES-03**: `Result` layout — tagged union, with niche optimization where
      a checked ability fact permits it — has one meaning in the core IR, the
      interpreter, and emitted C17.

### Agent Loop

- [ ] **DX-05**: `lang explain`'s cause DAG stays bounded when causes span
      functions, and names the function each cause step belongs to.

- [ ] **DX-06**: Cross-function blame attribution points at the correct fix
      location; a repair-then-re-check regression covers cases where the
      non-obvious function is the right one.

- [ ] **DX-07**: `lang-repair` reaches and fixes at least three new
      interprocedural defect classes through the JSON protocol alone, proven on
      a held-out fixture split.

## Future Requirements

Tracked, not in this roadmap.

### Modules — M003 lead candidate

- **MOD-01**: Separate compilation with per-module signature artifacts.
- **MOD-02**: Cross-module origin and ability contracts.

### Language Surface

- **GEN-01**: Ability-bounded generics, enabling a real generic `Result<T, E>`.
- **STD-01**: A deterministic standard library.
- **SVC-01**: A stable compiler-service contract (the prerequisite for LSP/MCP).

### Carried Validation Debt

- **QLT-09**: Nyquist validation for M001 Phases 5 and 6 — deliberately deferred
  because both surfaces are touched again by M002 (native tier, agent surface);
  validating now risks validating a shape that is about to change.

## Out of Scope

| Feature | Reason |
|---------|--------|
| Modules / separate compilation | Stacking it on the milestone that already extends five-plus-consumer loan liveness across a call boundary repeats the coordination-cost failure M001 avoided by deferring `OpCall`. Named as M003's lead candidate, not silently dropped. |
| Indirect / dynamic dispatch (function pointers, closures, trait objects) | Keeps the M002 call graph closed and typed — the property that makes cycle refusal and summary-based liveness tractable. |
| Whole-program borrow inference | The design Rust abandoned pre-1.0; breaks separate compilation and the evidence-manifest cache. Declared contracts instead. |
| Formal SMT / Alive2-style proof of interprocedural equivalence | Alive2, the state of the art, explicitly does not attempt interprocedural proof. Extend the empirical five-axis comparator instead. |
| Z3 / CVC5 / any cgo-bound SMT | Breaks the zero-external-dependency record; M002's properties are graph-reachability problems solvable by extending existing fixpoints. |
| Mutation-testing frameworks as a CI gate | `go-mutesting` upstream inactive; `gremlins` v0.6.0 is pre-1.0 and self-documents as not scaling. Hand-authored mutation-kill controls continue (QLT-08). |
| ThinLTO | Trades cross-module precision for build speed the project does not need, and would double the equivalence-proof surface. |
| MSan, `-fsanitize=cfi` | Different defect class (uninitialized memory) and a security-hardening mechanism for an indirect-call surface M002 does not have. |
| Implicit `From`-style error conversion at a propagation operator | Ambient conversion authority, against the project's own no-ambient-authority constitution. |
| Call-site override of a callee's declared ownership convention | The C++ overload-resolution cautionary case. |
| Implicit ARC-style refcount fallback for arguments | Hides runtime cost — directly against the core value. |
| Full generics, subtyping, variance, partial moves, effect rows, async | Unchanged from M001: they broaden the search space before the kernel oracle is trustworthy. |

## Traceability

Every M002 requirement maps to exactly one phase. Phase numbering continues from
M001 (which ended at Phase 06), so M002 runs Phases 07-13.

| Requirement | Phase | Status |
|-------------|-------|--------|
| SEM-04 | Phase 07 | Complete |
| SEM-05 | Phase 07 | Complete |
| SEM-06 | Phase 07 | Complete |
| SEM-07 | Phase 07 | Complete |
| SEM-08 | Phase 10 | Complete |
| SEM-09 | Phase 10 | Complete |
| OWN-05a | Phase 09 | Complete |
| OWN-05b | Phase 10 | Complete |
| OWN-06 | Phase 08 | Complete |
| OWN-07 | Phase 09 | Complete |
| OWN-08 | Phase 09 | Complete |
| OWN-09 | Phase 09 | Complete |
| NAT-04 | Phase 11 | Complete |
| NAT-05 | Phase 11 | Complete |
| NAT-06 | Phase 11 | Complete |
| NAT-07 | Phase 11 | Complete |
| TRU-02 | Phase 10 | Complete |
| TRU-03 | Phase 10 | Complete |
| TRU-04 | Phase 09 | Complete |
| QLT-03 | Phase 11 | Complete |
| QLT-04 | Phase 10 | Complete |
| QLT-05 | Phase 11 | Complete |
| QLT-06 | Phase 11 | Complete |
| QLT-07 | Phase 09 | Complete |
| QLT-08 | Phase 07 | Complete |
| EFF-02 | Phase 08 | Complete |
| RES-02 | Phase 12 | Pending |
| RES-03 | Phase 12 | Pending |
| DX-05 | Phase 13 | Pending |
| DX-06 | Phase 13 | Pending |
| DX-07 | Phase 13 | Pending |

**Coverage:**

- M002 requirements: 31 total (OWN-05 split into OWN-05a/OWN-05b, D-09-37)
- Mapped to phases: 31 ✓
- Unmapped: 0

**Per-phase counts:** 07 → 5, 08 → 2, 09 → 6, 10 → 6, 11 → 7, 12 → 2, 13 → 3.
(Phase 10 gains OWN-05b, the `interp` peer split from OWN-05, D-09-37; Phase
09 stays at 6 since OWN-05a fills the row OWN-05 previously occupied.)

**Note on QLT-08:** mapped to Phase 07, the phase that establishes it, but it is
a standing discipline enforced in every phase 07-13. Listed once so coverage
stays unambiguous.

## Scope-Cut Order

If the milestone runs over budget, cut in this order — explicitly, in writing,
at the moment it is decided (Key Lesson 4):

1. **RES-02, RES-03** (`Result` payloads) — structurally independent of the call
   machinery. Slip to M003; do not silently absorb as extra plans.

2. **QLT-07** (Nyquist fold-in) — if the pre-flight cost probe shows the fold-in
   is not cheap, re-scope to a stretch item and carry it as disclosed debt
   alongside QLT-09. **Not triggered:** the pre-flight inventory (D-09-40a,
   run 2026-09-10) found the fold-in cost within both pre-registered bounds
   (≤ 1 additional plan, zero new production files); QLT-07 shipped
   COMMITTED, closed in Phase 09's own `09-VALIDATION.md`.

3. **DX-05, DX-06, DX-07** (agent loop) — narrow to the two highest-value defect
   classes rather than cutting the mutation-kill / repair-then-re-check
   discipline for what does ship.

**Trigger:** if the loan-liveness phases exceed ~2x their initial plan estimate,
renegotiate `Result` payloads out to M003 immediately rather than adding plans.

**Never cut:** SEM-04, SEM-07, OWN-06, OWN-07, OWN-08, NAT-06, NAT-07. D-03-02 is
the one deliberately-carried debt item this milestone exists to close; deferring
it again would repeat M001's defer-under-load pattern on the same item for a
second consecutive milestone.

---
*Requirements defined: 2026-09-08*
*Last updated: 2026-09-08 after M002 roadmap creation*
