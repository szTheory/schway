# Project Research Summary

**Project:** Codename Lang
**Milestone:** M002 — Interprocedural Semantic Spine
**Domain:** Compiler engineering — extending an existing intraprocedural,
peer-validated, source-to-native Go compiler with Lang-to-Lang calls
**Researched:** 2026-09-08
**Confidence:** MEDIUM-HIGH overall (project-internal architectural claims HIGH,
grounded in code read directly; external precedent HIGH for facts/dates,
MEDIUM for the analogy drawn to this project; a few explicit LOW/inference
judgments are flagged where noted)

---

## Executive Summary

M002 adds one new `OperationKind` — `OpCall` — to a compiler that already has
a working intraprocedural semantic spine (M001: checker, two independent
validators, deterministic interpreter, readable-C17-through-Clang native
path, five-axis differential comparator). That single addition is deceptively
large: it must land at all six existing dispatch sites (`check`,
`corevalidate`, `interp`, `cgen`, `pathoracle`, `originvalidate`), each of
which pays its own cost, and it requires building call-graph construction,
cycle refusal, a bounded interpreter call stack, interprocedural loan
liveness in two independently-derived admission layers, and interprocedural
`-O3`/LTO equivalence — the one claim M001 could not make because it shipped
without calls. All four research lenses (stack, features, architecture,
pitfalls) and the fifth roadmap-option lens (OPTIONS.md) converge on the same
verdict: the milestone charter as already written in PROJECT.md is correctly
scoped, the recommended technique for every hard sub-problem is a bounded,
declared, independently-re-derived summary (never whole-program inference or
re-walking a callee's body), and the single biggest execution risk is not
"can this be built" but "will `check` and `corevalidate` drift while building
it," the same D-02-03/D-03-01 failure class M001 already paid for once.

The recommended approach: build the interprocedural spine exactly as
charter-scoped (no modules, no full generics), using signature-level
"declared contract, verified not trusted" summaries as the interprocedural
analog of M001's public-origin pattern (validated architecturally by
Rust NLL/Polonius and Val/Hylo's parameter-convention precedent), sequence
work so the hardest, most novel item (interprocedural loan liveness) lands
before the largest new-infrastructure item (multi-function C emission), and
treat Nyquist-for-Phase-3 as a fold-in of the phase that is already touching
that code rather than a separate pass. Every dependency decision holds the
zero-external-production-dependency record: `pgregory.net/rapid` and Go's
native fuzzer are test-only additions; SMT solvers, mutation-testing
frameworks, and `x/tools/go/callgraph` are explicitly rejected as imports
(read-only, where useful at all).

Key risks, and how research says to mitigate them: (1) cross-function
borrow/lifetime unsoundness at composition boundaries — mitigated by
requiring declared, checker-verified contracts at every call site, never a
re-walk of the callee body, with an adversarial corpus of call chains depth
≥2 with divergent loan states; (2) `check`/`corevalidate` silently drifting
on a new cross-cutting fact — mitigated by landing both peers in the same
plan/phase with a shadow-run diff on adversarial (not just in-corpus)
call-graph shapes before either ships; (3) an interprocedural `-O3`/LTO
"proof" that is actually inert because it lacks a composition-specific
negative control — mitigated by replicating M001's `false_no_alias`
methodology one call boundary later, proven to fail red before being trusted;
(4) whole-program/global inference creeping in under time pressure because
it is easier to implement than a properly bounded summary — the single most
likely failure mode per the project's own compute-efficiency constitution,
mitigated by a pre-Phase-2 cost-scaling spike. External precedent (Rust's
eight-year NLL→Polonius generalization, Go 1.18's generics front-end cost
cliff) corroborates that "local analysis generalized across a boundary" is a
class of problem with real, not hypothetical, cost/soundness risk — which is
exactly why the roadmap below front-loads a cheap falsification spike before
committing to the production liveness architecture.

---

## Key Findings

### Recommended Stack

Zero new external **production** dependencies — the zero-dependency record
established across M001's 6 phases holds. Two test-only additions and one
dev-tool addition earn their place; everything else researched (SMT solvers,
mutation-testing frameworks, `x/tools/go/callgraph`, libFuzzer, CFI, MSan,
ThinLTO) is explicitly rejected for M002's scope, either because it breaks
the dependency posture (cgo-bound Z3) or because it solves a harder/different
problem than M002 has (VTA/CHA/RTA solve Go-style dynamic dispatch; `OpCall`
is closed, typed, direct dispatch). See the verdict list below.

**Core additions:**
- `pgregory.net/rapid` v1.3.0 (test-only) — property-based generators +
  shrinking for call-graph shapes and cross-function loan scenarios; pure Go,
  zero transitive dependencies, integrated shrinker beats hand-rolled
  generators for producing minimal interprocedural counterexamples.
- `go test -fuzz` (Go 1.24 stdlib, already in use) — differential fuzzing of
  serialized call-graph/program encodings across the four engines, extending
  M001's five-axis comparator pattern rather than replacing it.
- `golang.org/x/perf/cmd/benchstat` (dev-tool, `go install`, not a `go.mod`
  dependency) — paired A/B statistical comparison for call-graph construction
  time, interprocedural admission cost, and native call-overhead by
  optimization tier, feeding the existing `internal/compiler/measure`
  protocol.

### Expected Features

**Must have (table stakes), consolidated from the language/feature survey:**
- Per-parameter ownership convention derived solely from the callee's
  declared signature (move/borrow), never overload-resolved at the call site
  — the C++ move/copy-overload anti-pattern is explicitly rejected.
- Declared, not inferred, ownership and origin contract at every function
  boundary — the direct interprocedural extension of M001's existing
  "declare public origins" pattern, matching Rust NLL/Polonius's own
  signature-as-summary design and Val/Hylo's fully-local convention system.
- `callable ⊆ publishable` gate as a single, uniform static predicate
  (D-04-03) — not piecemeal visibility-modifier logic the way Rust's `pub` +
  auto-trait leakage or Austral's export lists approximate it.
- Static, whole-corpus call graph, direct calls only (no first-class function
  values/dynamic dispatch in M002 scope) with SCC-based cycle detection and
  hard refusal — puts Lang in the MISRA-C/embedded-Rust camp (refuse
  recursion) rather than the Ada/SPARK camp (bounded-and-proven recursion),
  a deliberate, conservative scope cut.
- Bounded interpreter call stack producing a diagnosable verdict, never a
  host segfault.
- Independent re-derivation of every new interprocedural fact in both `check`
  and `corevalidate` (peer architecture, non-negotiable per project rule).
- Tagged-union `Result`/payload representation, exhaustive match-arm
  checking as a hard error, move-out-of-matched-payload as the default
  ownership semantics, and an error-propagation operator that reuses the
  existing three-stage cleanup/nonlocal-exit machinery rather than adding a
  fourth coexisting law.

**Should have (differentiators):**
- Signature-as-cache-key discipline — a callee body changing without its
  signature changing must not force caller re-verification, protecting the
  evidence-manifest/cache investment.
- Statically provable per-function stack bound (GNATstack-style), essentially
  free once recursion is refused (the call graph is a DAG).
- Niche-filling layout optimization for `Result`, derived from the ability
  system's already-checked non-null/non-zero facts — sequenced after the
  plain tagged-union representation is proven sound at all three engines.

**Defer (explicit anti-features for M002):**
- Call-site override of a callee's declared ownership convention.
- Implicit reference-counting/ARC-style fallback for arguments.
- Whole-program/cross-module borrow inference (the design Rust itself
  abandoned pre-1.0).
- A second, coexisting interprocedural liveness law alongside the
  intraprocedural one (retrospective Key Lesson 2 — retiring the old law is
  part of the same phase, not a follow-up).
- Implicit `From`-style error-type conversion at the propagation operator —
  in tension with the project's "no ambient authority escape" constitution;
  require explicit conversion instead.
- Indirect/dynamic dispatch calls (function pointers, closures) — no
  first-class function-value feature is in M002's charter.
- Formal SMT-based (Alive2-style) interprocedural equivalence proof — even
  Alive2, the state-of-the-art LLVM formal tool, has not solved this
  interprocedurally; differential testing is the correct, achievable tier.
- A single -O3 run with no engineered negative control as "sufficient"
  interprocedural equivalence evidence.

### Architecture Approach

Extend the existing six-dispatch-site architecture, never collapse it into
shared derivation code. The interprocedural contract is carried by widening
`core.Interface`/`core.FunctionSignature` (already body-stripped, already
digest-bound) into a full call contract — ownership requirement per
parameter, return-origin contract — produced once and **independently
re-derived by at least two of the six sites** (at minimum `corevalidate`,
ideally `originvalidate` for its origin-specific subset), never trusted from
the checker's own claim alone. A new `internal/compiler/callgraph` package,
built in the same shape as the existing `pathoracle` package (reads
`core.Program` only, refuses cycles with a named typed error), is consumed
by `check` at admission time and independently re-derived by `corevalidate`.
The interpreter gains a real call stack (`Frame` type, bounded depth, typed
refusal, per-frame ownership partitioning) for the first time. `cgen` carries
the largest single-site cost: it currently hard-refuses any program with more
than one function and must grow genuine multi-function C emission before
`OpCall` lowering is even possible.

**Major components (new or substantially extended):**
1. **Signature/summary artifact** (`core.FunctionSignature`/`core.Interface`,
   extended) — the interprocedural contract every dispatch site reads;
   never a callee body.
2. **`internal/compiler/callgraph`** (new package) — call-graph construction,
   SCC computation, cycle refusal; pathoracle-shaped, consumed by `check`,
   independently re-derived by `corevalidate`.
3. **Interprocedural loan liveness** (extends `check`'s `loanLivenessFixpoint`
   and `corevalidate`'s `recomputeLoanEndpoints`) — signature-mediated, never
   body-inlining, independently derived in each engine's own pre-existing
   mechanism shape.
4. **Interpreter call stack** (`interp`, new `Frame` infrastructure) — the
   deterministic-oracle authority path; must exist and stabilize before
   native equivalence work can differential-test against it.
5. **Multi-function C17 emission + call lowering** (`cgen`, new
   infrastructure) — largest blast-radius item; ordinary value/pointer
   calling convention reusing the existing by-value-vs-by-pointer fork,
   `restrict` only ever justified by a jointly-checked caller+callee fact.
6. **`Result`/payload representation** (`core.DataType.AlternativeDetails`,
   additive per D-04-30) — structurally independent of the call machinery,
   can be built in parallel, must land its per-site cases in the same plans
   that touch each site for `OpCall` to avoid a second pass over the same
   six files.

### Critical Pitfalls

1. **Cross-function borrow/lifetime unsoundness at composition boundaries**
   — mirrors real, accepted Rust unsoundness issues (#84366, #106431, #71550)
   that all occur specifically at signature/composition boundaries, not
   within a single function body. Avoid by requiring declared callee
   contracts checked at the call site, never a re-walk of the callee's CFG;
   gate on an adversarial corpus of composition-depth-≥2 divergent-loan-state
   cases.
2. **`check`/`corevalidate` silently drift on a new cross-function fact** —
   the direct generalization of M001's own D-02-03 ("a quadratic cost fix
   landed in the wrong half"). Avoid by landing both peers in the same
   plan/phase and shadow-run diffing on adversarial call-graph shapes
   (recursion, diamonds, deep chains) before either ships.
3. **An `-O3`/LTO tier that looks proven interprocedurally but is inert** —
   Rust's multi-year `noalias`+inlining saga (#31681/#54878) and LLVM
   #122537 both show a locally-true no-alias fact silently becoming false
   once composition (inlining) exposes it. Avoid by building a
   composition-only engineered negative control, proven to fail red, before
   trusting the equivalence claim — exactly the harder sequel to M001's
   `false_no_alias`.
4. **Fixpoint non-termination / SCC mishandling / native stack blowup** —
   the call graph is not guaranteed acyclic or bounded the way a CFG is.
   Avoid with explicit SCC computation, visited-set-guarded (never native
   recursive) traversal, and an iteration-count fail-closed guard on every
   new fixpoint.
5. **Generator/reducer under-coverage on multi-function programs** — an
   intraprocedural-tuned generator and single-function-tuned HDD reducer both
   systematically under-reach call-graph shape (depth, fan-out, SCC
   structure) unless it is explicitly modeled as its own generation
   dimension with its own reachability register (mirroring D-04-21).

---

## Recommended Milestone Shape — Reconciled

**OPTIONS.md's recommendation:** ship the charter as already written in
PROJECT.md (no modules, no full generics), fold Phase-3-relevant Nyquist
closure into whichever phase rebuilds loan liveness (not a separate pass, not
expanded to Phases 5/6), land the agent-loop/diagnostics extension as the
milestone's own final phase once the interprocedural defect taxonomy is
actually known, and explicitly name modules as M003's lead candidate rather
than silently dropping them.

**How the other four documents relate to that recommendation:**

- **STACK.md supports it directly** — nothing in the stack research implies
  a scope larger than the charter; every "explicitly not added" item
  (SMT, mutation frameworks, x/tools callgraph as import, CFI, MSan,
  ThinLTO) reinforces staying inside the charter's actual proof obligations
  rather than reaching for heavier machinery.
- **FEATURES.md supports it and adds texture, not contradiction** — its
  §6 "breadth pass" ranks generics/full `Result<T,E>` genericity (rank 1),
  a minimal effect surface (rank 2, flagged premature), a stdlib (rank 3),
  bounded recursion (rank 4, explicitly premature), and the
  compiler-service contract (rank 5, "correctly sequenced after M002") as
  *post-M002* work — i.e., FEATURES.md independently arrives at "don't widen
  M002's scope," just from the feature-precedent angle rather than the
  process-risk angle OPTIONS.md uses. No contradiction.
- **ARCHITECTURE.md supports it and gives OPTIONS.md's recommendation its
  load-bearing technical justification** — the six-dispatch-site cost table
  and the "8-10 independent edits minimum" finding is exactly *why*
  OPTIONS.md's coordination-cost argument against bundling modules holds:
  ARCHITECTURE.md independently shows the charter alone is already a
  five-plus-consumer-coordination problem (Site 4 needs a new call stack
  from scratch, Site 5 needs new multi-function emission infrastructure)
  before modules are even considered.
- **PITFALLS.md supports it and sharpens the fold-Nyquist-in decision** —
  Pitfall 2 (`check`/`corevalidate` drift) and its "land both peers in the
  same plan" prescription is the direct mechanism argument for why folding
  Phase-3 Nyquist closure into the phase that already re-touches loan
  liveness is efficient rather than merely convenient: the file is already
  open, the review context is already loaded, and deferring it risks
  validating a shape that changes again once D-03-02 closes anyway.

**No document complicates or contradicts the OPTIONS.md recommendation.**
The strongest tension surfaced is internal to OPTIONS.md itself (the
verification-engineer lens wants full Nyquist closure now; product/strategy
is fine shipping charter-as-scoped since the gap is disclosed debt) and
OPTIONS.md resolves it the same way this synthesis adopts: fold only the
*directly load-bearing* slice (Phase 3, since interprocedural liveness
extends it), carry Phases 5/6 Nyquist forward as still-disclosed debt.

**Reconciled answer: adopt OPTIONS.md's option (f) as written.** Charter as
the spine; Nyquist-for-Phase-3 folded into the loan-liveness rebuild phase;
agent-loop/diagnostics extension as the final phase; modules and Phase 5/6
Nyquist explicitly named Out of Scope in PROJECT.md, not silently dropped.

---

## Reconciled Build Order

ARCHITECTURE.md's 10 dependency-ordered stages and OPTIONS.md's 6-7 phase
sequence describe the same underlying dependency graph at different
granularity — ARCHITECTURE.md is closer to plan-level detail (what must
literally compile before what), OPTIONS.md is phase-level (what a phase
gate should assert). They do not disagree on ordering; they differ in
where the phase boundaries fall and in whether Nyquist/agent-loop/Result
are separate phases or folded/deferred. This synthesis reconciles them into
one recommended phase sequence, mapping ARCHITECTURE.md's stages onto it and
noting where the two sources cut the boundaries differently.

### Recommended phase sequence

**Phase 0 (pre-Phase-1 spike, non-blocking to charter scope): Interprocedural
liveness cost-scaling probe.** Not a phase in either source's numbering, but
OPTIONS.md §5's Experiment 1 is explicitly Phase-2-blocking — it must
complete *before* the production liveness design is committed. Cheapest
disproving experiment: synthetic call-graph corpus, counted-work measurement,
throwaway Go workbench (spike discipline, not production code). Informs
whether the fixpoint can extend directly or needs a memoized summary-caching
layer designed in from the start. Grounded in real precedent (Go 1.18's
15-18% front-end-specific generics regression; Rust's Polonius performance
wall) that this exact failure shape is not hypothetical.

**Phase 1 — Call-graph construction and cycle refusal.**
Maps to ARCHITECTURE.md Stage 0 (signature/summary artifact — prerequisite
for everything) + Stage 1 (`core.OpCall` registration + `check` producer
path) + Stage 2 (callgraph package + cycle refusal). OPTIONS.md's Phase 1
matches this directly. **Where the sources cut differently:** ARCHITECTURE.md
treats Stage 0 (signature extension) as strictly separable and landable in
isolation against the existing Phase 4 corpus before any call exists;
OPTIONS.md folds it wordlessly into "Phase 1." Take ARCHITECTURE.md's finer
cut for planning: land Stage 0 first as its own plan/gate (pure data-shape
change, testable without calls existing), then Stage 1+2 in the same phase.
Registering `core.OpCall` in `AllOperationKinds()` immediately forces all six
dispatch-site switches to fail compilation/registry tests until each adds a
case — deliberately used as a forcing function, not an accident.
Run the recursive/mutually-recursive stress-corpus spike (OPTIONS.md §5
Experiment 3) alongside this phase, since it is genuinely non-blocking to
charter deliverables.

**Phase 2 — Interprocedural loan liveness in `check` (Site 1, `check`'s
half of D-03-02).**
Maps to ARCHITECTURE.md Stage 4's `check`-side half, pulled forward of
`corevalidate`'s peer re-derivation per M001's own established order
(`check` before `corevalidate` in every phase, per ARCHITECTURE.md's own
observation). Riskiest assumption: the fixpoint stays a bounded,
summary-based analysis rather than becoming whole-program inference in
disguise (directly the failure mode `compute-efficiency-constitution.md`
names). Gate: a scale/cost measurement on realistic call-graph fan-out,
informed by Phase 0's spike.

**Phase 3 — Independent re-derivation in `corevalidate` (Site 2) +
call-graph cycle re-derivation (peer half of Stage 2) + Phase-3-relevant
Nyquist closure.**
Maps to ARCHITECTURE.md Stage 3 (corevalidate independent re-derivation,
explicitly notable as "can proceed in parallel with Stage 2" once Stage 1
lands — but OPTIONS.md sequences it strictly after Phase 2 for the
independent-target reason above; **this is a real granularity difference,
not a contradiction**: ARCHITECTURE.md is describing what *can* build in
parallel at the implementation-dependency level, OPTIONS.md is describing
what a *phase gate* should require in sequence for the two-peer-agreement
question to be meaningful. Take OPTIONS.md's sequencing for phase planning
— `check`'s liveness law needs to exist as a stable target before its
independent peer can be validated against it — while noting `corevalidate`'s
own engineering work could start earlier if staffing allows) + the rest of
Stage 4 (the `corevalidate` half of interprocedural loan liveness). This is
also where Nyquist-for-Phase-3 folds in, per OPTIONS.md's option (f) and
Pitfall 2's "land both peers in the same plan" mechanism. Gate: counted-work
lane proving linear-or-declared-bounded cost, plus a deliberate seeded
endpoint-level fault proving the two peers actually diverge when one is
wrong.

**Phase 4 — Interprocedural differential + native `-O3`/LTO equivalence.**
Maps to ARCHITECTURE.md Stage 5 (`originvalidate`/`pathoracle`, lower risk,
mirrors the already-proven `OpForeignCall` pattern) + Stage 6 (interpreter
call stack — must land and stabilize as the trusted oracle before Stage 7)
+ Stage 7 (multi-function C emission + call lowering, the largest new
infrastructure item) + Stage 8 (cross-function differential/exhaustive-
dispatch rebuild). OPTIONS.md collapses these four ARCHITECTURE stages into
one phase because they share a single headline claim (interprocedural `-O3`
equivalence) and the same "needs a trusted interpreter oracle first"
dependency chain. **Reconciliation: keep ARCHITECTURE.md's internal stage
order as the plan sequence within this phase** (Sites 3+6 in parallel with
each other → Site 4 (interp) stabilizes → Site 5 (cgen) → Stage 8 closing
verification), but treat it as one phase for gating purposes, matching
OPTIONS.md's "this is where the milestone's central risk gets adjudicated"
framing. Gate: at least one engineered interprocedural adversarial program
per new alias/inlining-sensitive attribute, proven to reproduce the
`interpreter == -O0 != -O3` divergence shape M001's Phase 5 established.
Run the dev-backend timing probe (OPTIONS.md §5 Experiment 2) alongside this
phase, non-blocking.

**Phase 5 — `Result` payloads (D-04-30, ARCHITECTURE.md §7 / Stage 9).**
Both sources agree this is structurally independent of the call machinery
and can be built in parallel with or after Phases 1-4, landing its per-site
cases in the same plans that touch each site for `OpCall` (per
ARCHITECTURE.md's own scheduling-risk warning) rather than a second sweep.
Sequenced last because it is the milestone's most legitimately negotiable
item if Phases 2-4 run over budget. Gate: an explicit compatibility check —
does `Result` payload origin/ownership reduce to already-proven Phase 2-4
machinery, or does it secretly need new interprocedural rules (OPTIONS.md §5
Experiment 5)?

**Phase 6 — Agent-loop extension for interprocedural defect classes.**
Both sources agree this must be last: the interprocedural defect taxonomy
(cross-function borrow escape, cycle refusal, `-O3` divergence, `Result`
payload defects) is only knowable once Phases 1-5 exist. Gate: the existing
DX-04-style held-out/derivation-split fixture methodology replicated for
2-3 new defect classes, proving `lang-repair` actually reaches and fixes the
new shape through the JSON protocol, not merely that the diagnostic fires.

**Summary of where the two sources disagreed and what was taken:**

| Disagreement | ARCHITECTURE.md | OPTIONS.md | Reconciled |
|---|---|---|---|
| Stage 0 (signature) as its own gate vs. folded into Phase 1 | Separable, land-first | Implicit in "Phase 1" | Took ARCHITECTURE.md's finer cut — land Stage 0 first as its own plan |
| corevalidate (Stage 3) parallel with callgraph (Stage 2)? | Yes, can proceed in parallel once Stage 1 lands | No — corevalidate's liveness peer needs check's liveness law to exist first | Took OPTIONS.md's sequencing for the *liveness* half specifically; corevalidate's non-liveness engineering (structural re-derivation, cycle re-check) can still start earlier per ARCHITECTURE.md |
| Number of phases | 10 stages | 6-7 phases | Reconciled to 6 phases + 1 pre-phase spike, using OPTIONS.md's phase granularity for gating and ARCHITECTURE.md's stage granularity for intra-phase plan sequencing |
| Nyquist closure | Not addressed (out of scope for that document) | Folded into Phase 3 | Adopted OPTIONS.md's fold, justified independently by PITFALLS.md Pitfall 2 |

---

## The Six Dispatch Sites — Per-Site Cost of `OpCall`

Every phase touching `OperationKind` pays at all six. Registering `OpCall` in
`core.AllOperationKinds()` forces every exhaustive switch to fail to
compile/fail the registry test until each site adds a case — used
deliberately as a forcing function.

| # | Site | Dispatch shape | `OpCall` cost | Risk class |
|---|------|------|------|------|
| 1 | `check` | AST→core producer; `resolveForeignStep` already refuses Lang callees via `core.call_target_not_foreign` | Flip the existing refusal into a real producer path (`checkLangCall`) + cross-function admission read against the callee's signature summary | Moderate — mechanical flip of an existing refusal, but first landing of the summary-consumption pattern |
| 2 | `corevalidate` | Two independent exhaustive switches (straight-line, branch) | New `case core.OpCall` in *both*, each independently re-deriving structural consistency, ownership transfer, and interprocedural loan liveness | High — literal site of the D-02-03/D-03-01 "fix one of two peers" hazard |
| 3 | `originvalidate` | Partial switch on origin-relevant kinds during backward walk | New `case core.OpCall` mirroring the already-proven `OpForeignCall` hop | Low — proven mechanism, just a new case |
| 4 | `interp` | Two exhaustive switches (straight-line, arm), no call stack today | Deepest change: needs a real `Frame`/call-stack from scratch (new subsystem, not a case arm) | High — new infrastructure, and the deterministic-oracle authority path everything else differential-tests against |
| 5 | `cgen` | 5+ exhaustive switch sites; `Emit`/`EmitNative` hard-refuse `len(Functions)!=1` today | Largest: needs multi-function C emission infrastructure *before* any `case core.OpCall` addition is meaningful; a shared `emitCall` helper recommended over 5 independent copies | Highest — genuinely new infrastructure, largest blast radius of the six |
| 6 | `pathoracle` | Partial switch during path linearization; must not import `check`/`corevalidate` | New, third independent re-derivation of the cross-function loan-chain rule (cannot reuse Sites 1/2's derivation by design) | Moderate — proven mechanism shape (mirrors `originvalidate`), but must stay structurally independent |

Plus the two exhaustive-dispatch controls (`core_test.go`'s in-process
`control:kind.exhaustive_dispatch` and `session.go`'s CLI-observable lane) —
landing `OpCall` costs a minimum of **8-10 independent edits** before it is
"real" per the project's own D-04-22 standard.

---

## High-Risk Integration Points and Their Gates

Drawing PITFALLS.md's pitfall→phase→gate mapping together with
ARCHITECTURE.md's risk ranking (Stage 4 = highest risk, Stage 7 = second
highest):

| Integration point | Named risk (Pitfall #) | Architecture stage | Gate that catches failure |
|---|---|---|---|
| Interprocedural loan liveness in `check` (signature-mediated, summary-based) | Pitfall 1 (cross-function unsoundness), Pitfall 4 (fixpoint non-termination) | Stage 4 (ARCHITECTURE's own "single highest-risk integration point") | Composition-depth-≥2 divergent-loan-state adversarial corpus; iteration-count fail-closed guard on the fixpoint; pre-Phase-2 cost-scaling spike |
| `check` ↔ `corevalidate` agreement on every new interprocedural fact | Pitfall 2 | Stages 3+4 | Shadow-run diff over recursion/diamond/deep-chain call graphs showing zero divergence before either layer ships, landed in the same plan |
| `restrict`/no-alias emission surviving Clang's inliner across a call boundary | Pitfall 3 | Stage 7 (cgen) | Engineered composition-only negative control (harder sequel to `false_no_alias`), proven to fail red pre-fix |
| Call-graph/interprocedural traversal code (native Go recursion vs. explicit worklist) | Pitfall 4 | Stages 2, 4 | Indirect-cycle and pathological-depth-but-acyclic corpus; native-stack-overflow probe distinct from the language-level bound |
| Generator/HDD reducer coverage of multi-function programs | Pitfall 5 | Stage 8 (closing verification) | Dedicated call-graph-shape reachability register (mirroring D-04-21); reducer output re-verified to reproduce the same property as its input |
| Content-bound cache keyed per-unit vs. call-graph-shaped facts | Pitfall 6 | Not explicitly staged in ARCHITECTURE.md — first phase making any interprocedural fact cacheable (likely Stage 7/Phase 4) | Callee-changes-invalidates-caller regression test, gating before any interprocedural fact is marked cacheable |
| `lang explain`/`lang-repair` blame attribution across a call boundary | Pitfall 7 | Phase 6 (agent-loop extension) | Repair-then-re-check regression test covering cases where the non-obvious function is the correct fix location |

**Highest-risk points, named explicitly (both sources agree):** interprocedural
loan liveness (Phase 2/3, Stage 4) and multi-function C emission + native
equivalence (Phase 4, Stage 7). Both deserve their own mid-phase gates,
mirroring M001's own Phase 3 mid-phase-gate precedent.

---

## Stack Decisions — Verdict List

The zero-external-production-dependency record is a hard constraint; every
verdict below respects it.

| Item | Verdict | One-line reason |
|---|---|---|
| `pgregory.net/rapid` v1.3.0 | **Adopt** (test-only) | Pure Go, zero transitive deps, integrated shrinker beats hand-rolled generators for interprocedural counterexamples |
| `go test -fuzz` (Go 1.24 stdlib) | **Adopt** | Already the toolchain in use; extend M001's comparator pattern with a fuzz-driven front end |
| `golang.org/x/perf/cmd/benchstat` | **Adopt** (dev-tool, `go install`, never `go.mod`) | Official golang.org/x sub-repo; paired A/B comparison front end for the existing p50/p95/CoV measurement protocol |
| `golang.org/x/tools/go/callgraph/{cha,rta,vta}` | **Read, don't import** | Solves Go's dynamic-dispatch problem (interfaces, first-class function values); `OpCall` is closed/typed direct dispatch, a simpler problem; VTA is self-documented experimental |
| LLVM `CallGraphSCCPass` / Rust NLL-polonius design | **Read, don't import** (no embeddable dependency exists) | Cited as architectural precedent only — SCC bottom-up traversal order, signature-as-interprocedural-contract |
| Z3/CVC5 (any cgo-bound SMT binding) | **Reject** | Requires building/linking external C/C++ library, breaks the zero-dependency record; M002's properties are graph-reachability problems solvable by extending existing fixpoints, not SMT-shaped |
| Alloy | **Reject** as CI/build dependency; offline design-time use only | Standalone Java tool, not embeddable; use exactly like `.planning/spikes/` — never wired into the pipeline |
| `gopter` / `testing/quick` | **Reject** | Weaker maintenance/shrinking (gopter) or frozen/unmaintained with no real shrinking (`testing/quick`) versus `rapid` |
| libFuzzer via cgo | **Reject** | Requires a cgo/C++ coverage-instrumented build; Go's native fuzzer already provides coverage-guided mutation for the actual target (Lang programs as data) |
| `go-mutesting` / `gremlins` (mutation testing) | **Reject** as CI gate; narrow periodic audit only if adopted later | Upstream inactive (go-mutesting) or pre-1.0 and self-documented as not scaling to large modules (gremlins v0.6.0); continue hand-authored mutation-kill controls instead |
| `-fsanitize=cfi` | **Reject** for M002 | Security-hardening mechanism, not correctness-proving; `OpCall` is direct/typed dispatch, no indirect-call attack surface yet |
| MSan | **Reject** for M002 | Requires instrumented libc++/custom runtime; detects a different defect class (uninitialized memory) than M002's charter |
| ThinLTO | **Reject** | Trades cross-module precision for build speed the project doesn't need yet; would double the equivalence-proof surface |
| A second/separate comparator for interprocedural equivalence | **Reject** | Would duplicate M001's five-axis comparator infrastructure instead of extending its corpus and controls |

---

## Anti-Features and What NOT to Do — Consolidated

From FEATURES.md's anti-feature tables and OPTIONS.md's "what I would not
do":

**Design/feature anti-patterns:**
- Call-site override of a callee's declared ownership convention (the C++
  overload-resolution cautionary tale).
- Implicit reference-counting/ARC-style fallback for arguments that don't
  cleanly move or borrow.
- Whole-program/cross-module inference of borrow relationships instead of
  declared contracts — precisely the design Rust abandoned pre-1.0, and it
  breaks separate compilation and the evidence-manifest cache.
- A second, coexisting interprocedural loan-liveness law that doesn't retire
  the intraprocedural one in the same phase (retrospective Key Lesson 2).
- Indirect/dynamic dispatch calls (function pointers, closures, trait-
  object-style dispatch) admitted into the M002 call graph.
- Silent depth-based stack growth with no fixed, documented ceiling.
- Implicit `From`-style error-type conversion at the propagation operator.
- Attempting a formal SMT-based (Alive2-style) proof of interprocedural
  equivalence this milestone — genuinely unsolved even by state-of-the-art
  tools, disproportionate to the charter.
- Treating a single `-O3` run with no engineered negative control as
  sufficient interprocedural equivalence evidence.

**Process/scoping anti-patterns (OPTIONS.md's "what I would NOT do"):**
- Do not bundle modules/separate compilation into M002, even though it
  would avoid rebuilding the call-graph/origin work later — the
  coordination-cost evidence from M001's own Phase 3 is too direct a
  warning; name modules explicitly as M003's lead candidate instead.
- Do not run Nyquist closure for Phases 5 and 6 (M001) inside M002 — their
  surfaces are about to be touched again by M002's own work (native tier,
  agent surface), so validating now risks validating a soon-obsolete shape.
- Do not treat the bounded spikes/experiments as scope-additive — stop when
  the gate is answered, do not expand a spike into production code.
- Do not let `Result` payloads or the Phase-3-folded Nyquist closure become
  fixed commitments that block the milestone if the harder phases (loan
  liveness, corevalidate re-derivation) run over budget.
- Do not skip the cost-scaling measurement spike and go straight to
  production liveness implementation — Go 1.18 generics and Rust's Polonius
  wall are direct evidence this exact class of feature has real cost-cliff
  risk in the front-end/checking layer specifically.
- Do not build the agent-loop/diagnostics extension in parallel with the
  semantic work — the defect taxonomy isn't known until Phases 1-5 exist;
  building blind risks the wrong cause-DAG edges.

---

## Bounded Experiments/Spikes Worth Running

All follow the project's own spike discipline (falsifiable question, named
competing mechanisms, an oracle that doesn't reuse the mechanism under test,
at least one injected defect, stop when the gate is answered).

| Spike | Hypothesis | Decision it informs | When |
|---|---|---|---|
| Interprocedural liveness cost-scaling probe | Function-summary-based liveness stays linear/sub-quadratic in call-graph size | Whether Phase 2 extends `loanLivenessFixpoint` directly or needs a memoized summary-caching layer designed in from the start (Phase-2-**blocking**) | Before Phase 1 finishes / before Phase 2's plan is written |
| Recursive/mutually-recursive call stress corpus | Bounded call stack + cycle refusal correctly distinguish legal deep recursion from illegal cycles, without false positives on realistic (parser-shaped) programs | Whether the call-stack bound is a fixed constant or needs a declared/measured budget (Phase-1-blocking design decision) | Alongside Phase 1 |
| Dev-backend timing probe (Cranelift/QBE vs. Clang) | Switching dev backends would not have changed how Phase 4's `-O3`/LTO divergence-finding worked | Whether Frontier 3's "development backend" deferral remains correctly low-risk, or whether interprocedural work makes backend choice materially more consequential | Alongside Phase 4, non-blocking, informs M003 not M002 scope |
| Nyquist-closure cost measurement for M001 Phase 3 | Closing Nyquist validation for the loan-liveness phase is genuinely cheap when folded into the code M002 is already touching | Whether the fold-in Phase 3 design holds as scoped, or whether it should itself become an explicit stretch/fallback item | Before finalizing the M002 phase plan (cheap, hours not days) |
| `Result` payload / interprocedural origin interaction probe | `Result` payload origins reduce cleanly to already-proven Phase 2-4 machinery, no new kernel rule needed | Whether Phase 5 is genuinely independent surface work (safe to treat as negotiable/movable to M003) or secretly depends on new interprocedural rules | After Phase 4 lands, before Phase 5 starts |

---

## Scope-Cut Order

If the milestone runs over budget (Phase 3's own M001 history — 70-95
min/plan for five-consumer-coordinated work versus ~10 min/plan for
single-consumer work, and 3 corrective plans for one derivation — is the
base rate to expect, not the exception), OPTIONS.md names `Result` payloads
and folded Nyquist closure as the negotiable items. This synthesis confirms
that ordering and makes the trigger explicit:

1. **First cut: `Result` payloads (Phase 5).** Structurally independent of
   the call machinery (confirmed by both ARCHITECTURE.md §7 and OPTIONS.md
   Phase 5's design), most legitimate claim to being movable — slip to M003
   explicitly, do not silently absorb as extra plans on the committed scope.
2. **Second cut: Phase-3-folded Nyquist closure.** If the pre-flight cost
   measurement spike (see experiments table) shows the fold-in is not
   actually cheap, re-scope it to an explicit stretch/fallback item rather
   than a committed part of Phase 3 — carry it forward as still-disclosed
   debt alongside Phases 5/6 Nyquist, per Lesson 4 (declare deferred scope
   in writing at decision time).
3. **Third cut: agent-loop/diagnostics extension (Phase 6) scope narrowing.**
   If Phases 1-5 consume more budget than planned, Phase 6 is the next most
   negotiable — it is additive to the charter's core semantic claim (calls
   are sound across function boundaries) rather than load-bearing for it;
   narrow to the highest-value 2-3 defect classes (cross-function borrow
   escape, cycle refusal) rather than the full taxonomy, and defer the
   remainder rather than cutting corners on the mutation-kill/repair-then-
   re-check discipline for what does ship.

**Explicit trigger, per Lesson 6 (phase-count drift as signal):** if Phase 2
or Phase 3 (the highest-risk phases) start exceeding ~2x their initial plan
estimate, that is the trigger to renegotiate Phase 5 out to M003
immediately — not to silently keep adding plans. **What must never be cut:**
the charter's own headline claims — `OpCall` at all six sites, cycle
refusal, interprocedural loan liveness closing D-03-02, and interprocedural
`-O3`/LTO equivalence — since D-03-02 is the one deliberately-carried debt
item the milestone exists to close, and cutting it would reproduce M001's
own "defer under load" pattern for a second consecutive milestone on the
same debt item.

---

## Implications for Roadmap

### Phase 1: Call-graph construction and cycle refusal
**Rationale:** Every later phase depends on `OpCall` existing as a real,
structurally admitted operation and on a call graph existing to reason about;
has no dependency on the harder loan-liveness question, so it is the
cheapest correct place to start.
**Delivers:** `core.OpCall` registered at all six sites (forcing compile/test
failures until each adds a case); signature/summary artifact extended;
`internal/compiler/callgraph` package with SCC-based cycle refusal; bounded
interpreter call-stack ceiling declared (not yet wired into full call
execution).
**Addresses:** Table-stakes items — static call graph, cycle refusal,
declared/not-inferred contract at the boundary, `callable ⊆ publishable`
gate.
**Avoids:** Pitfall 4 (fixpoint/SCC mishandling, native stack blowup) via
explicit visited-set-guarded traversal from day one.

### Phase 2: Interprocedural loan liveness in `check`
**Rationale:** The milestone's actual charter centerpiece and single carried
debt item (D-03-02); must land before its independent-checker peer so the
peer has a stable target to validate against, matching M001's own
`check`-before-`corevalidate` order.
**Delivers:** Signature-mediated interprocedural liveness extension to
`loanLivenessFixpoint`.
**Uses:** Stack — no new dependency; this is compiler engineering against
the existing fixpoint architecture, informed by the pre-Phase-1/2
cost-scaling spike.
**Implements:** Architecture component — the signature-as-interprocedural-
contract design (Design C from ARCHITECTURE.md §2).

### Phase 3: `corevalidate` independent re-derivation + Nyquist-Phase-3 fold-in
**Rationale:** Per D-02-03's lesson, Phase 2 alone does not close D-03-02 —
only two independently-derived, agreeing peers do. Folding Nyquist closure
here is close to free (file already open, mental model already loaded).
**Delivers:** `corevalidate`'s independent cross-function loan-liveness and
cycle re-derivation; Phase-3-relevant Nyquist validation closed.
**Addresses:** Pitfall 2's exact prescription (land both peers in the same
plan, shadow-run diff on adversarial call-graph shapes).

### Phase 4: Interprocedural differential + native `-O3`/LTO equivalence
**Rationale:** Depends on Phases 1-3 being stable; deliberately placed
before the lower-risk `Result` work so the milestone's hardest, most
irreversible claim gets tested with maximum remaining schedule slack.
**Delivers:** `originvalidate`/`pathoracle` `OpCall` cases (mirroring the
proven `OpForeignCall` pattern); a real interpreter call stack (`Frame`
type, stabilized as the trusted oracle); multi-function C17 emission +
call lowering in `cgen`; cross-function exhaustive-dispatch rebuild; the
engineered composition-only `-O3`/LTO negative control.
**Avoids:** Pitfall 3 (inert proof tier) via the mandatory composition-only
negative control gate.

### Phase 5: `Result` payloads (D-04-30)
**Rationale:** Structurally independent of the call machinery; the
milestone's most legitimately negotiable item if Phases 2-4 run over budget.
**Delivers:** `core.DataType.AlternativeDetails`, payload-carrying
alternative construct/match in all three engines, tagged-union C17 layout.
**Addresses:** `Result`/sum-type table stakes from FEATURES.md §4.

### Phase 6: Agent-loop extension for interprocedural defect classes
**Rationale:** The defect taxonomy is only known once Phases 1-5 exist;
building diagnostics blind risks the wrong cause-DAG shape.
**Delivers:** Cause-DAG/repair-applicability coverage for cross-function
borrow escape, call-cycle refusal, interprocedural `-O3` divergence, and
`Result` payload defects.
**Avoids:** Pitfall 7 (misattributed cross-function diagnostics) via a
repair-then-re-check regression test for each new defect class.

### Phase Ordering Rationale

- **Dependency chain:** Phase 1 (call graph/summary) → Phase 2 (`check`
  liveness) → Phase 3 (`corevalidate` peer + Nyquist fold) → Phase 4
  (interp/cgen/native equivalence, needs a trusted interpreter oracle first)
  → Phase 5 (`Result`, parallelizable but sequenced last as the negotiable
  item) → Phase 6 (agent-loop, needs the defect taxonomy to exist).
- **Grouping rationale:** ARCHITECTURE.md's finer 10-stage breakdown collapses
  cleanly into these 6 phases because Stages 5-8 all share the same
  headline dependency (a stabilized interpreter oracle) and the same
  headline claim (interprocedural `-O3`/LTO equivalence).
- **Pitfall avoidance built into the sequencing itself:** peer-pair phases
  (2+3) are adjacent and gated on shadow-run agreement; the native-equivalence
  phase (4) is gated on an engineered negative control before being trusted;
  the diagnostics phase (6) is deliberately last so it is built against a
  known, not guessed, defect taxonomy.

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 2/3 (interprocedural loan liveness):** Highest-risk, most novel
  item in the milestone — external precedent (Rust NLL/Polonius) shows this
  class of problem has real multi-year-scale difficulty even for larger
  teams; the pre-Phase-2 cost-scaling spike is a research-phase-equivalent
  activity that should run before planning commits to an architecture.
- **Phase 4 (native `-O3`/LTO equivalence, multi-function cgen):** Largest
  new-infrastructure item (`cgen` currently refuses multi-function programs
  entirely); needs research into calling-convention and `sret`/aggregate
  ABI choices specific to Lang's `Result` shapes before this phase is
  planned in detail.
- **Phase 6 (agent-loop extension):** Diagnostic blame-attribution across a
  call boundary is a genuinely open UX/design question (rustc's own
  multi-year investment here is cited precedent); worth a dedicated design
  pass once the defect taxonomy exists.

Phases with standard, well-documented patterns (skip deep research-phase):
- **Phase 1 (call graph/cycle refusal):** SCC/Tarjan-style cycle detection
  is well-trodden; the pattern to follow (`pathoracle`-shaped package,
  visited-set-guarded walks) already exists in the codebase.
- **Phase 3 (`corevalidate` peer re-derivation):** Mechanism is a direct
  extension of M001's own established reachability-closure pattern, not a
  new design.
- **Phase 5 (`Result` payloads):** D-04-30 already names the additive design
  plan in detail; low novelty risk.

---

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | MEDIUM-HIGH | Primary sources (pkg.go.dev, LLVM/Clang docs, go.dev) verified for every library/flag claim; project-fit judgments (e.g., "gopter is weaker than rapid") are explicitly labeled comparative inference, not sourced facts |
| Features | MEDIUM-HIGH | Rust/Ada/SPARK/LLVM precedent is primary-sourced and strong; Val/Hylo and Austral (pre-1.0 projects) are MEDIUM since their own public docs are the primary source; anti-feature judgments are calibrated inference against M001's own decisions, explicitly tagged |
| Architecture | HIGH for structural claims (grounded in code read directly from the repo — line numbers cited throughout); MEDIUM for external comparisons (Rust rustc-dev-guide cited from prior knowledge, not re-fetched this session) |
| Pitfalls | MEDIUM-HIGH | General compiler-literature classes verified from primary sources (issue trackers, CompCert/seL4 papers, WG14 provenance TS); the project-specific pitfall→phase mapping is HIGH confidence as direct inference from this project's own recorded architecture and retrospective |
| Options/sequencing | MEDIUM-HIGH | Project-internal claims HIGH (read from primary planning artifacts in full); external precedent (Rust Polonius timeline, Go 1.18 generics regression, Zig stage1→stage2, CompCert/seL4) verified via web search with dates, MEDIUM confidence on the analogy drawn to this specific project |

**Overall confidence:** MEDIUM-HIGH — the milestone shape and build order
recommendations rest on convergent, independently-derived findings across
five research documents that were not coordinated with each other during
research, which strengthens confidence in the reconciled recommendation
specifically because the disagreements found (see Reconciled Build Order
table) were narrow and resolvable, not fundamental.

### Gaps to Address

- **Interprocedural liveness cost model is unmeasured.** Every source agrees
  this is the highest-risk unknown; the pre-Phase-2 cost-scaling spike (see
  Bounded Experiments) is the recommended way to close this gap before
  planning commits to a specific fixpoint architecture — treat it as a
  planning prerequisite, not an optional nice-to-have.
- **`Result` payload / interprocedural-origin interaction is unverified.**
  D-04-30's design plan predates M002's interprocedural work; it is assumed
  independent but not yet confirmed — the Phase 5 pre-flight probe (Bounded
  Experiments) should run before Phase 5 planning, and if it surfaces a
  hidden dependency on Phase 2-4 machinery, Phase 5 must be replanned (or
  moved to M003) rather than silently absorbing new interprocedural rules.
- **Nyquist-fold-in cost for Phase 3 is an estimate, not measured.**
  OPTIONS.md itself flags this as "the one place in the recommended
  sequencing where the estimate hasn't been empirically checked yet" — the
  cheap pre-flight audit (Bounded Experiments) should run before the M002
  phase plan is finalized, not discovered mid-Phase-3.
- **Diagnostic blame-attribution design for cross-function causes has no
  concrete algorithm yet**, only a design principle ("blame whichever side
  violates its own declared contract"). This needs a dedicated design pass
  in Phase 6's own planning, informed by the actual defect taxonomy Phases
  1-5 produce — do not attempt to pre-design this before that taxonomy
  exists.
- **cgen's calling-convention and aggregate-return ABI choices for `OpCall`
  are design recommendations (`[INFER]` in ARCHITECTURE.md), not yet
  validated against real generated C.** Should be an early decision point
  within Phase 4's planning, informed by the `ForeignContract.Layout`
  precedent already in the codebase.

## Sources

### Primary (HIGH confidence)
- `pkg.go.dev/pgregory.net/rapid`, `pkg.go.dev/golang.org/x/tools/go/callgraph/{cha,rta,vta}`, `pkg.go.dev/golang.org/x/perf/cmd/benchstat` — verified 2026-09-08
- `go.dev/doc/security/fuzz/`, `go.dev/src/testing/fuzz.go` — Go native fuzzing constraints
- `llvm.org/docs/LangRef.html`, `llvm.org/docs/LinkTimeOptimization.html`, `clang.llvm.org/docs/{AddressSanitizer,MemorySanitizer,ControlFlowIntegrity}.html` — verified 2026-09-08
- Rust two-phase borrows, NLL architecture, Polonius design, `Try`/`?` RFC, tagged-union/niche RFC — rustc-dev-guide and rust-lang/rfcs, accessed 2026-09-08
- Val/Hylo parameter conventions, Austral linear types — project language docs, accessed 2026-09-08
- AdaCore Platinum SPARK / GNATstack case study; MISRA C Rule 17.2 / undecidability analysis (arXiv:2212.13933)
- Alive2 paper and repo (interprocedural scope limitation), EMI/Orion original paper
- rust-lang/rust issues #84366, #106431, #71550, #123241, #31681, #54878; llvm/llvm-project#122537, #98978; WG14 N3005 provenance TS; CompCert structure docs; seL4 verification papers — all accessed 2026-09-08
- Rust Polonius alpha announcement (Aug 2026), EuroRust 2024 Polonius retrospective, Zig stage1→stage2 transition, golang/go#49569 (Go 1.18 generics regression), seL4/CompCert CACM coverage — fetched via web search 2026-09-08
- This project's own codebase, read directly: `core.go`, `check.go`, `corevalidate.go`, `originvalidate.go`, `interp.go`, `cgen.go`, `pathoracle.go`, `cache.go`, `ability.go`, `session.go`, plus `PROJECT.md`, `RETROSPECTIVE.md`, `ROADMAP.md`, `M001-MILESTONE-AUDIT.md`, `04-DEBT.md`, and the `wiki/` design corpus

### Secondary (MEDIUM confidence)
- `go-gremlins/gremlins`, `zimmski/go-mutesting`/`avito-tech/go-mutesting` maintenance-status assessment (GitHub project pages, dated within research window)
- YARPGen/Csmith bug-count figures (secondary reporting via arXiv survey, not primary Intel paper)
- Swift SR-8546/exclusivity enforcement, seL4 static-stack-sizing convention, `cargo-call-stack` — general ecosystem knowledge, cross-referenced not independently re-verified this pass

### Tertiary (LOW confidence, flagged for validation)
- Rust NLL/Polonius interprocedural design stance cited in STACK.md as "well-established consensus," not re-fetched from rust-lang primary sources this pass — flag for verification if it becomes load-bearing for a specific design decision beyond architectural precedent
- seL4 static stack sizing convention — general formal-verification-kernel knowledge, not independently re-verified

---
*Research completed: 2026-09-08*
*Ready for roadmap: yes*
