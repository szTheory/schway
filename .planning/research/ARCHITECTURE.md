# Architecture Research: M002 Interprocedural Semantic Spine

**Domain:** Compiler internals — adding Lang-to-Lang calls to an existing
6-phase, peer-validated, source-to-native Go compiler for Codename Lang.
**Researched:** 2026-09-08
**Confidence:** HIGH for structural claims (grounded in files read directly
from `~/projects/ai-lang`); MEDIUM for external-comparison claims
(Rust/LLVM, cited but not independently re-verified against primary sources
this session); LOW/inference explicitly flagged where the answer is a design
recommendation rather than a fact already in the codebase.

**Reading key:** `[READ]` = read from this codebase this session. `[EXT]` =
external primary source, cited. `[INFER]` = design inference/recommendation,
not yet decided in the codebase.

---

## 1. The Six Dispatch Sites — the single most important deliverable

`[READ]` The "six dispatch sites" are not a metaphor — they are a literal,
named, tested table. `internal/compiler/core/core.go:244-249` states it
directly:

> "This is the single table every dispatch site (check, corevalidate, interp,
> cgen, pathoracle, originvalidate) is tested against (D-04-22)."

`internal/compiler/core/core_test.go:266-280`, `TestAllOperationKindsHandledAtEverySite`
— the `control:kind.exhaustive_dispatch` control — drives every declared
`core.OperationKind` through all six real programs and asserts none crash or
wrongly reject. `internal/compiler/session/session.go:2501-2574` is the
session-layer, CLI-observable sibling lane (`lane:kind-exhaustive-dispatch`),
proving the same claim through the shipped `lang verify` binary, not just
in-process tests.

The six sites, in the order the registry and both control tests name them:

### Site 1 — `check` (producer/admission site)
**File:** `internal/compiler/check/check.go`
**Entry points:** `Program()` (line 44), the dispatch fork at
`checkFallibleLinear` (line 1269) which routes to `checkForeignTracer`
(1357) or `checkResourceLifecycle` (1499), plus the two straight-line/branch
walkers `analyzeStraightLine` (1804) / `analyzeArmBody` (784).
**Shape:** `check` is a *producer*, not a `switch operation.Kind` consumer —
it does not walk an already-built operation list and dispatch on `.Kind`;
it constructs `core.LinearOperation`s from AST as it derives ownership,
borrow, and admission facts. The nearest thing to a "dispatch" here is
`resolveForeignStep` (line 1316), which already resolves a fallible call's
callee name against two disjoint buckets — `foreignSymbols` and
`functionNames` (`Program()`, line 87-90) — and **already emits a
dedicated diagnostic, `core.call_target_not_foreign` (line 1326), the
instant a callee resolves into `functionNames` instead of `foreignSymbols`.**
This is the exact refusal M002 must flip into acceptance. `Program()`'s own
comment (lines 77-81) says it directly: "the foreign symbol table and the
Lang function-name set both exist purely so a fallible call's callee can be
resolved against one or the other (D-04-01/D-04-02)."
**Cost of `OpCall`:** New — a real `OpCall`-emitting path parallel to
`checkForeignTracer`/`checkResourceLifecycle` (call it `checkLangCall`),
reusing `resolveForeignStep`'s existing bucket-resolution shape but
producing `core.OpCall` instead of refusing. Also new: cross-function loan
liveness admission (see §4) and the `callable ⊆ publishable` gate read
against the callee's declared `core.PublicOrigin`/`core.Interface` summary
(see §2) rather than the callee's body — `check` on the caller side must
never open the callee's `Linear` body, it is source-blind to sibling
functions.

### Site 2 — `corevalidate` (independent structural + admission re-derivation)
**File:** `internal/compiler/corevalidate/corevalidate.go`
**Entry points:** two independent `switch operation.Kind` blocks — one in
`replayStraightLine` (line 903-1008) and one in `replayBlocks` (line
1098-1210+) — because straight-line and branch-shaped bodies are replayed by
genuinely different code, not a shared helper (mirroring the same
duplication discipline `check.go` already has between straight-line and arm
analysis).
**Shape:** classic exhaustive `switch operation.Kind { case core.OpCopy: ...
default: return v.check(false, "core.unknown_operation", ...) }`. The
`default:` arm is fail-closed — an unregistered kind is rejected, not
silently accepted (`core.unknown_operation` at line 1004).
**Cost of `OpCall`:** New `case core.OpCall:` in *both* switches (straight-line
and branch), each independently re-deriving: (a) that the callee's declared
signature is structurally consistent with the call site's argument/return
places, (b) that the call transfers/borrows ownership exactly as `check`
claims, (c) the interprocedural half of loan liveness (§4) — this is the
literal site of the Lesson-3 hazard the retrospective names: "Fixing one of
N independent peers is not fixing the item... every fix has N sites." A fix
landed only in `replayStraightLine` and not `replayBlocks` (or vice versa)
reproduces exactly the D-02-03/D-03-01 failure class the retrospective
records ("A quadratic cost fix landed in the wrong half").

### Site 3 — `originvalidate` (published-origin re-derivation)
**File:** `internal/compiler/originvalidate/originvalidate.go`
**Entry point:** `RecomputeOriginPerReturn` (line 133), whose backward walk
(`walkReturnOrigin`, line 169) already has a `switch operation.Kind` (line
182) handling `OpBorrowExclusive`, `OpBorrowShared`, and — critically —
`OpForeignCall` (line 191), where a foreign call's declared
`ForeignContract.Alias` ("borrow"/"retain") is treated as a real access-mode
hop, exactly like an in-language borrow, per D-04-28.
**Shape:** per the D-04-22 registry comment (core.go:274-279), `originvalidate`
does **not** switch exhaustively on every kind today — it only discriminates
terminators (`isTerminatorKind`, a membership test against
`core.TerminatorKinds()`) to find where to start each backward walk, then
switches on the *specific* kinds relevant to origin propagation. "Handled" at
this site currently means "the walk completes without error," a weaker
contract than the other five sites (explicitly named as D-04-29 debt,
out of scope for that plan).
**Cost of `OpCall`:** New `case core.OpCall:` in `walkReturnOrigin`'s switch,
exactly mirroring the existing `OpForeignCall` handling — a call's declared
callee-side origin (its `PublicOrigin`/`Interface` summary, §2) becomes a
real access-mode hop, so a caller returning a value borrow-derived through a
callee's borrow-returning function is still recognized as borrow-derived, not
silently treated as owned. This is the *direct mechanism* that closes
D-03-02 ("an exported borrow-derived return with no declared origin... in
its interprocedural half").

### Site 4 — `interp` (deterministic oracle)
**File:** `internal/compiler/interp/interp.go`
**Entry points:** two switches — one in the straight-line walk inside `Run`
(implicit, around line 83 in the pre-M002 single-function-body shape) and one
in `runBranchArm` (line 83-138, the shown excerpt) for arm-bodied match
functions. Both are exhaustive `switch operation.Kind` with a `default:`
returning an error on unknown kind.
**Shape:** notably, `interp.go`'s own comments at lines 100-101 and 121-123
explicitly acknowledge cases (`OpForeignCall`, `OpRelease`) that "no match
arm can produce... this phase" but are still handled, "solely so
control:kind.exhaustive_dispatch's six-site table finds every kind handled at
every site." This is the codebase *already documenting* the cost of the
six-site discipline for M001's own operations — the same tax M002 pays again
for `OpCall`.
**Cost of `OpCall`:** New — this is the deepest architectural change of the
six, because `interp.Run` today has no notion of a call stack at all (see
§5). `case core.OpCall:` must push a new frame, evaluate the callee
function's body (found via `findFunction`, already multi-function-aware —
see below), and pop/return a value, all while the deterministic-oracle
authority property (`interp` defines correctness, not C or optimizer
behavior) is preserved.

### Site 5 — `cgen` (C17 emission)
**File:** `internal/compiler/cgen/cgen.go`
**Entry points:** at least five separate `switch operation.Kind` sites across
`emitLinear`/`emitLinearBorrowedByPointer`/`emitLinearBorrowedByPointerPlain`
(lines 259, 466, 651 — near-identical triplicated straight-line emitters),
`emitLinearForeign`'s per-block emitter (lines 1118, 1158), and
`emitBranchOperations` (line 1651-1729, the excerpt read this session).
**Shape:** exhaustive `switch` with explicit "known but unsupported" cases
(e.g. `case core.OpForeignCall, core.OpFail:` inside `emitBranchOperations`,
line 1723-1729) rather than falling through to `default:`, specifically so
the six-site control finds a *named* unsupported case rather than an unknown
kind — the same discipline as `interp`.
**Structural fact with the largest blast radius:** `cgen.Emit` (line 16) and
`cgen.EmitNative` (line 46) both hard-fail today with `"C emitter expects one
function"` the instant `len(program.Functions) != 1`
(`internal/compiler/cgen/cgen.go:22,52`). **cgen currently has no
multi-function C emission at all** — every generated translation unit is one
`int main(...)` (or a `static` helper for the borrowed-by-pointer variants,
itself only ever called *from* `main`, never emitting a second Lang-level
function). This is the single largest new-construction item in M002, not an
incremental `case core.OpCall:` addition: cgen must grow (a) per-function C
function definitions for every non-entry-point Lang function, (b) a real
call-site expression compiling `core.OpCall` to a C function call, (c) a
calling convention decision (§6), and (d) `main`'s own shape changes from
"the one function" to "call the entry function." `check`, `corevalidate`, and
`interp` already iterate `program.Functions` (plural) and are not gated this
way — cgen is the outlier.
**Cost of `OpCall`:** New `case core.OpCall:` at every existing switch site
(five-plus locations, all of which need the same lowering logic — a strong
argument for factoring one shared `emitCall` helper before or during OpCall
landing, rather than five independent copies drifting), *plus* the
prerequisite multi-function-program emission infrastructure above.

### Site 6 — `pathoracle` (independent path-enumeration oracle)
**File:** `internal/compiler/pathoracle/pathoracle.go`
**Entry point:** `RecomputeEndpoints` (line 354) → `linearizePath` (line
279), which reads `operation.Kind` in exactly two ways: `isTerminatorKind`
(membership test, same shape as `originvalidate`'s) to close a path, and an
`if operation.Kind == core.OpBorrowShared || core.OpBorrowExclusive` at line
306 to record a loan's birth.
**Shape:** per its own package doc (lines 1-28) and the D-04-22 registry
note, `pathoracle` is deliberately the *third, structurally distinct*
loan-liveness decision procedure — a genuinely different mechanism class
(exhaustive concrete-path enumeration, never a fixpoint or a relation
closure) from both `check.go`'s `loanLivenessFixpoint` and
`corevalidate.go`'s `recomputeLoanEndpoints`. Like `originvalidate`, it does
not switch exhaustively on every `OperationKind` today — "handled" here means
"the walk completes without error," per the same D-04-29 weaker-contract
note.
**Cost of `OpCall`:** New — `linearizePath` must decide what an `OpCall`
means for the place-inheritance chain it tracks (does a loan cross a call
boundary through an argument, and does the callee's declared origin
re-enter the chain on return — mirroring `originvalidate`'s new
`OpForeignCall`-shaped hop). Because `pathoracle` explicitly must not import
`check` or `corevalidate` (package doc, lines 4-7), it cannot reuse whatever
interprocedural liveness derivation `check`/`corevalidate` build for §4 — it
must independently re-derive the cross-function loan-chain rule from the
core artifact alone, a *third* independent derivation of the same
interprocedural fact, exactly matching the intraprocedural precedent it
already set for M001.

### Summary table

| # | Site | File | Kind of dispatch | `OpCall` cost class |
|---|------|------|-------------------|----------------------|
| 1 | check | `internal/compiler/check/check.go` | AST→core producer; existing callee-bucket resolution (`resolveForeignStep`) already refuses Lang callees | Flip an existing refusal into a new producer path + cross-function admission |
| 2 | corevalidate | `internal/compiler/corevalidate/corevalidate.go` | Two independent exhaustive `switch operation.Kind` (straight-line, branch) | New `case` in both, each re-deriving cross-function loan facts independently |
| 3 | originvalidate | `internal/compiler/originvalidate/originvalidate.go` | Partial switch on origin-relevant kinds during backward walk | New `case core.OpCall` mirroring existing `OpForeignCall` hop; closes D-03-02 |
| 4 | interp | `internal/compiler/interp/interp.go` | Two exhaustive `switch operation.Kind` (straight-line, arm) | Deepest: needs a real call stack (new, see §5) |
| 5 | cgen | `internal/compiler/cgen/cgen.go` | 5+ exhaustive `switch operation.Kind` sites | Largest: needs multi-function C emission (new infrastructure, see §6), not just a case arm |
| 6 | pathoracle | `internal/compiler/pathoracle/pathoracle.go` | Partial switch during path linearization | New, third independent cross-function loan-chain re-derivation |

**Maintainer's lens:** six real sites plus the `AllOperationKinds()`
registry test plus two exhaustive-dispatch controls (`core_test.go`'s
in-process one and `session.go`'s CLI-observable lane) means landing
`OpCall` costs a minimum of **eight to ten independent edits** before the
kind is "real" per this project's own definition of real (D-04-22's own
standard). The retrospective's Lesson 3 is not hypothetical for M002 — it is
the literal shape of every plan that touches `OpCall`.

---

## 2. Function Signature as an Interprocedural Contract

`[READ]` The codebase already has the artifact this question asks about,
built for a narrower purpose (separate-compilation origin publication,
Phase 3 OWN-04) but structurally reusable:

- `core.Interface` (`internal/compiler/core/core.go:148-153`) — "a subset of
  `Program` containing only module identity, a digest binding it to the
  exact `core.Program` it was derived from, and per-function signatures —
  explicitly no `Linear` or `Match` body."
- `core.FunctionSignature` (`core.go:164-171`) — "deliberately has no
  Linear/Match field at all — not merely an omitted one — so a consumer
  decoding this type structurally cannot reach a body even by accident."
  Carries `Parameter`, `ReturnType`, `PublicOrigin`, and `Abilities`.
- Built by `originvalidate.BuildInterface` (`originvalidate.go:415-434`) —
  note the *producer* of this artifact is `originvalidate`, not `check`: the
  summary is derived independently from the checked core, source-blind,
  consistent with `originvalidate`'s whole package posture ("never trusted,
  only recomputed from `core.Program`/`core.Function` facts").
- Consumed today by `session.go:1027` (`RunInterfaceCommandFile`, writes the
  digest-bound summary to disk) and by `protocol.go:133`
  ("directly from a core.Interface summary — never from a body field").

### Three designs, evaluated against this architecture specifically

**Design A — Full whole-program analysis.** Every call site's admission
decision is computed by walking the entire program's call graph and every
callee body in one global pass (closest analogue: an old-style whole-program
alias analysis).
- *Pro:* maximally precise; no summary staleness.
- *Con, decisively against this architecture:* it destroys the
  peer-verification model. Every one of `check`/`corevalidate`/`interp`/
  `cgen`/`pathoracle`/`originvalidate` would need the *entire* program in
  scope to answer any per-function question, and `corevalidate`'s
  source-blind posture ("intentionally does not know source or reuse
  checker/interpreter authorization code," `corevalidate.go:1-3`) and
  `originvalidate`'s and `pathoracle`'s explicit "does not import check"
  constraints become incoherent — a whole-program analysis has to trust
  *something* upstream, and this project's entire architecture is built to
  never do that. `[INFER]` Reject.

**Design B — Per-function summaries with declared origins (Rust's
approach).** `[EXT]` Rust's borrow checker computes per-function facts and,
at a call boundary, consults the callee's *signature* (lifetime parameters
and their variance/outlives relationships declared in the function
signature) rather than re-borrow-checking the callee's body at every call
site — this is exactly what makes separate compilation and generic
functions tractable in rustc's NLL/Polonius-family borrow checker. See the
rustc dev guide's description of how function signatures carry the complete
interprocedural borrow contract (region/lifetime bounds are part of the
signature, not inferred at each call site):
https://rustc-dev-guide.rust-lang.org/borrow_check.html and
https://rustc-dev-guide.rust-lang.org/mir/borrowck.html . `[EXT, MEDIUM
confidence — cited but not re-verified against the exact current rustc
source this session]`
- *Pro:* well-proven design; matches this project's own `PublicOrigin`
  precedent almost exactly (a declared-not-inferred cross-function contract,
  the same posture Codename Lang already chose for OWN-04's public borrow
  origins over inferring them).
- *Con:* if the "summary" lives as fields directly on `core.Function` (or is
  read straight off the callee's own `core.Function` inside the same
  `core.Program`), a caller-side check that reads the callee's *checked*
  facts is trusting the checker's own claim about the callee, not an
  independently re-derived fact — this is exactly the shape the
  retrospective's `escape:coordinated-source-to-core-false-claim` names as
  the project's one accepted, deliberately-not-solved hazard. Using the
  signature *as checked by `check` alone* would import a second layer of
  the same coordinated-lie risk one level higher (a lying producer *and* a
  lying summary, both self-consistent).

**Design C — A separate summary artifact, validated independently
(fits the re-derivation principle).** `[INFER, HIGH confidence — this is the
architecture the codebase already half-built]` Extend `core.Interface`/
`core.FunctionSignature` from an origin-only publication artifact into the
full interprocedural call contract: add fields for the parameter's ownership
requirement (owned/borrowed-shared/borrowed-exclusive), and — critically —
have this summary re-derived independently by more than one of the six
sites (at minimum: `corevalidate`, since it is the trust boundary before
`interp`/`cgen` ever run, and ideally re-verified again by `originvalidate`
for the origin-specific subset it already owns). A call site's admission
decision reads the callee's `FunctionSignature`, never the callee's
`Linear`/`Match` body — enforced structurally by `FunctionSignature`
literally not having a body field, exactly as `FunctionSignature`'s own doc
comment already argues for the origin case.
- *Pro:* directly continues the pattern the codebase already committed to
  (declared-not-inferred contracts, body-stripped artifacts, independent
  re-derivation at every trust crossing). Reuses `core.Interface`'s existing
  digest-binding (`CoreDigest`) so a stale summary is detectable, not
  silently trusted — this answers the "genuinely independent, not sharing an
  implementation" half of §4 for free at the signature layer.
- *Con:* it is the design with the most new work — the summary must be
  computed by *at least two* independent producers (the checker's own
  admission-time summary, and a `corevalidate`-side or dedicated
  `callsummary`-package recomputation from the checked core), not merely
  extended once. This is explicitly the cost the retrospective's Lesson 3
  predicts, paid deliberately rather than by accident.

**Recommendation:** Design C. It is not a new idea for this codebase — it is
`core.Interface`/`core.FunctionSignature` widened to carry the full call
contract, produced and independently re-derived the same way `PublicOrigin`
already is. Design B (trust the checker's own signature) is acceptable only
as the *producer's* half of Design C, never as the sole source of truth a
caller's admission decision rests on.

---

## 3. Call Graph Placement

`[READ]` No call graph exists in the codebase today — `grep`ing for
"call graph," "callgraph," or a graph/adjacency structure over functions
in `internal/compiler` returns nothing. `session.go` does contain generic
graph-walk discipline ("graph walks carry visited-set guards and named
refusal codes rather than hanging," per the retrospective's own Patterns
Established section) but no *function* call graph.

### Which component should own it

`[INFER]` Three candidates, evaluated:

- **`core` package.** Attractive because `core.Program` already holds every
  `core.Function` and is the one artifact every one of the six sites reads.
  But `core` is a passive data-shape package (types + two small pure
  functions, `AllOperationKinds`/`TerminatorKinds`) with zero decision logic
  today — adding cycle-detection *logic* here would break that
  separation and make `core` a second place decisions live, undermining the
  "independent re-derivation" principle (a fact computed once in `core` and
  reused everywhere is the opposite of six independent derivations).
- **A new package, `internal/compiler/callgraph`.** `[INFER, recommended]`
  Mirrors the existing precedent exactly: `pathoracle` is *already* "a
  package that owns one specific whole-function derivation, consumed only by
  tests and the verify harness (`session.go`) — never by `check`,
  `corevalidate`, `interp`, or `cgen`" (pathoracle.go's own doc comment,
  paraphrased). A `callgraph` package built the same way — reads
  `core.Program` only, builds the adjacency relation from `OpCall`
  operations' callee references, refuses a cycle with a named, typed error
  (matching `pathoracle`'s `backEdgeError`/`pathCapError` shape exactly) —
  is a natural sibling, not a novel pattern. It should be **consumed by
  `check`** at admission time (a caller must know its own call graph
  position to refuse recursion before ever emitting `OpCall`), and
  **independently re-derived by `corevalidate`** (same cycle-refusal
  question, asked again from the checked core artifact alone, per the
  established two-peer discipline) — this is the natural extension of the
  six-site pattern to a whole-program-shaped fact, not a violation of it.
- **`session` (the verify/CLI orchestration layer).** Wrong layer — `session`
  already depends on `check`/`corevalidate`/`interp`/`cgen`/`pathoracle`/
  `originvalidate`; putting graph logic here would make it a seventh implicit
  dispatch site with none of the discipline (no dedicated `_test.go`, no
  registry entry) the real six have.

### Interaction with `cache`'s content-binding — is it a conflict?

`[READ]` `internal/compiler/cache/cache.go`'s package doc (lines 1-25) is
explicit: the cache is "a local, content-addressed store for expensive
intermediate ARTIFACTS... keyed on an explicitly DECLARED input list
(D-06-07)" and structurally "stores ARTIFACTS ONLY, never a verdict,
judgement, or pass/fail outcome" — enforced by
`TestCacheExportedSurfaceStoresNoVerdict`.

**This is not a conflict, once correctly framed.** A call graph is a
whole-program *fact*, but the cache doesn't need to store the call graph
itself as a first-class object — it needs the call graph's *content
identity* folded into the `cache.Input` list for any per-function artifact
that is affected by interprocedural analysis. Concretely: today a cached
artifact (a compiled binary, an instrumented binary) is keyed only on
per-unit declared inputs. Once a function can call another function,
compiling function A's artifact is no longer independent of function B's
source — the correct fix is **widen the declared input list**, not change
`cache`'s per-unit shape: add a `cache.Input{Name: "callgraph-digest",
Digest: <hash of the reachable call subgraph's content>}` (or, more
conservatively and more in keeping with `cache.ComputeKey`'s existing
whole-program-blind design, key every cached native artifact on the whole
program's core digest rather than a single function's, the same way
`core.Interface.CoreDigest` already binds a summary to its exact source
program). This keeps `cache` structurally unaware of call graphs — it never
learns what a call graph *means*, it only receives one more named, digested
input — which is exactly the "content-bound, structurally cannot hold a
verdict" property doing its job at a new scale, not being violated by it.
`[INFER]`

### Cycle refusal — where does the refusal code live so all peers agree

`[READ]` The precedent is `pathoracle`'s `backEdgeError`
(`pathoracle.go:112-125`): "OWN-03 is scoped to acyclic CFGs this phase...
this package independently re-derives that rejection rather than trusting
check.go's own `loanLivenessFixpoint` to have caught it first, so a
corrupted or synthetic CFG that check.go never saw is still refused here."
`[INFER]` The same discipline applies one level up: cycle refusal for the
*call* graph must be re-derived independently by at least two peers — the
proposed `callgraph` package (consumed by `check` at admission time, refusing
before an `OpCall` is even emitted) and `corevalidate` (refusing again from
the checked core artifact alone, exactly as it re-derives every other
admission fact `check` already claimed). A single stable diagnostic code
(e.g. `core.call_graph_cycle`, following the existing `core.*` code
namespace convention seen in `core.call_target_not_foreign`,
`core.borrow_conflict`, `core.move_while_borrowed`) should be shared as a
*string constant*, never as shared cycle-detection logic — the code names
the same fact, the derivation stays independent, matching this project's
own "shared inert records only" pattern (retrospective, Patterns
Established).

---

## 4. Interprocedural Loan Liveness in Two Admission Layers

`[READ]` What M001 did intraprocedurally: `check.go`'s
`loanLivenessFixpoint` (line 603-716) is, per the retrospective, "the sole
liveness law" after "a 230,692-comparison shadow run" against the
now-retired `discoverLoanLastUses" (Phase 5, `05-02`). It is a backward
worklist fixpoint over per-block live-loan sets on the function's own CFG
(`cfgBlockSpec`), described in the package doc as never enumerating a
concrete path, only converging abstract per-block summaries.

Independently, `corevalidate.go`'s `recomputeLoanEndpoints` (line 694-828)
computes "ONE bounded reachability closure" over an explicit
`(loan, block, ordinal)` use relation — "a single global relation, never
per-block summaries and never a path" (corevalidate's own doc comment at
`replayStraightLine`, lines 878-889, explicitly contrasts its
`ownerLiveSharedUntil`/`ownerLiveExclusiveUntil` running high-water marks
against `ownerBlockedUntil`'s single whole-function aggregate — two
genuinely different mechanisms even *within* corevalidate).

And `pathoracle` is the acknowledged *third* mechanism class (exhaustive
concrete-path enumeration), consumed only by tests/verify, never production.

**So M001 already has three structurally distinct liveness derivations for
the intraprocedural case, by design** — this is the existing discipline
M002 must extend, not a new invention.

### What changes for the interprocedural case

The cross-function question is genuinely new: does a loan taken in function
A and passed by reference into function B remain live across the call, and
when does it end — at the call site (if B only borrows transiently) or does
B's own return extend it (if B returns a borrow derived from its parameter,
per `PublicOrigin`)? This is precisely D-03-02's interprocedural half: "an
exported borrow-derived return with no declared origin exports
indistinguishable from a fully-owned return... in the INTERPROCEDURAL half
of the hazard" (`M001-MILESTONE-AUDIT.md:36`).

### Design options for keeping `check` and `corevalidate` genuinely
independent while agreeing

`[INFER]`

1. **Signature-mediated liveness (recommended, extends §2 Design C
   directly).** Neither `check` nor `corevalidate` ever look inside the
   callee's body to decide interprocedural liveness. Both consult the
   callee's `core.FunctionSignature`/`core.Interface` summary (declared
   ownership-requirement per parameter, declared origin/access on return)
   and apply that declared contract using their own, already-different
   intraprocedural mechanism (fixpoint on the caller side in `check`,
   reachability-closure on the caller side in `corevalidate`) to decide
   whether the call site itself extends or ends the loan. This is exactly
   how the existing `OpForeignCall`/`ForeignContract.Alias` precedent
   already works for foreign calls (§1, Site 3) — `OpCall` should be a
   drop-in structural sibling of `OpForeignCall` for liveness purposes, not
   a new mechanism. Independence is preserved because the *summary* is a
   shared inert fact (like `core.PublicOrigin` already is), never shared
   derivation code — `check` and `corevalidate` each fold the summary into
   their own pre-existing, differently-shaped intraprocedural liveness
   engine.
2. **Full body inlining/substitution at admission time.** Treat a call as
   "paste the callee's operations into the caller's CFG and re-run the
   existing intraprocedural fixpoint/closure." Rejected: breaks separate
   compilation, breaks recursion entirely (§5/§3), and makes the summary
   artifact (§2) pointless — if the body is always available and always
   trusted, there is no reason for a body-stripped `Interface` to exist.
3. **A dedicated interprocedural liveness pass shared as a library.**
   Compute cross-function liveness once, call it from both `check` and
   `corevalidate`. Rejected outright — this is precisely the anti-pattern
   the retrospective calls out as "Fixing one of N independent peers is not
   fixing the item" in reverse: a shared implementation isn't a peer at all,
   it's a single point of failure wearing two names, and it collapses the
   verification value the six-site architecture exists to buy.

**Agreement mechanism:** the same one M001 already uses for
`check`↔`corevalidate` intraprocedurally — no shared code, same declared
input (here: the callee's summary digest, not its body), same admission
question, and a test suite that runs both derivations against a shared
corpus and asserts identical accept/reject verdicts (the direct
interprocedural analogue of Phase 3's "exhaustive loan-endpoint
differentials," which M002's charter explicitly names for cross-function
rebuild).

---

## 5. Interpreter Call Stack

`[READ]` `interp.Run` (`interp.go:17-50`) has no call stack today — it
looks up exactly one function by name (`findFunction`, already
multi-function-aware: it searches `program.Functions` by name, so the
*data model* already supports multiple functions per program, only the
*execution model* does not yet call between them), and either runs its
linear body (`runLinear`) or its selected match arm (`runBranchArm`,
lines 56-140+) to a single `Execution`/`Outcome`. There is no `Frame` type,
no stack slice, and no depth counter anywhere in `internal/compiler/interp`.

### Design `[INFER]`

- **Bounded depth.** Follow the existing fail-closed-cap precedent exactly:
  `pathoracle.MaxPaths` (const, 4096, with an explicit rationale comment
  about DoS surface) and the diagnostic/error shape `pathCapError`/
  `backEdgeError` implement (`Code()` method, typed error, refuses rather
  than truncates). A `MaxCallDepth` constant with the same shape —
  `interp.callDepthExceededError{functionID string, depth, limit int}`,
  refused fail-closed, never silently deepened — is the natural sibling.
  The cap must be a genuine, documented, conservative ceiling (like
  `MaxPaths`'s "well above [64], a decimal order of magnitude" reasoning),
  not tuned to just barely pass the milestone's own corpus.
- **Frame representation.** `[INFER]` A `frame` struct parallel to the
  existing per-invocation state `runLinear`/`runBranchArm` already build
  locally (`values map[string]string` in `runBranchArm`, line 72) — each
  call pushes `{functionID, values map[string]string, liveResources
  []string}` (mirroring `Execution.LiveResources`'s existing shape, which
  already exists for the single-function fallible-resource case in Phase 4)
  onto a `[]frame` stack. `Run`'s existing single-frame `values` map becomes
  the base of this stack rather than a special case.
- **Drop/cleanup ordering on return and on early exit.** `[READ]` The
  existing single-function precedent for reverse-order cleanup is Phase 4's
  three-stage acquisition ("exact reverse-order partial cleanup," per
  PROJECT.md's Context section) enforced by `check.go`'s
  `releaseAllocatorMismatch` (line 1679) and `corevalidate`'s
  `checkReleaseOrder` (`corevalidate.go:1258`). `[INFER]` For calls, the
  same reverse-order discipline must extend across the frame boundary: on
  an ordinary `OpReturn`, the callee's own live resources must already be
  fully released (a callee cannot return with any live foreign acquisition
  — this should be a `corevalidate`-checked invariant on the callee's
  *signature*, not something the caller re-derives per call site) before its
  frame pops; on early exit (`OpFail`/`OpDefect` inside the callee), the same
  cleanup-then-terminate sequencing already proven for the single-frame case
  applies unchanged *within* the popping frame, and the caller's own frame
  is untouched (the callee's failure is a normal value returned across the
  call boundary via `OpFail`'s existing typed-failure shape — Codename Lang
  has no exceptions or unwinding to design here, only ordinary return-value
  propagation, which simplifies this considerably compared to a
  language with unwinding).
- **Ownership state at the call boundary.** `[INFER]` The interpreter's
  existing `values map[string]string` per-place ownership tracking (move =
  `delete` + reinsert, per `runBranchArm`'s `OpMove` case at line 87-90)
  must be partitioned per frame — a moved-from place in the caller becomes
  genuinely inaccessible to the caller for the call's duration, and the
  callee's own places are a disjoint namespace seeded only by the
  passed argument(s). This is a natural extension of the existing per-frame
  `values` map design, not a new ownership model.

---

## 6. C17 Lowering of Calls

`[READ]` As established in §1 Site 5, `cgen.Emit`/`cgen.EmitNative`
currently refuse any program with more than one function
(`cgen.go:22,52`). The nearest existing analogue for "calling something
that isn't inline code" is `emitLinearForeign` (line 749+), which declares
an `extern` C symbol (`fmt.Fprintf(&out, "extern %s %s(unsigned char
argument);\n\n", ...)`, line 811) and calls it — but that call target is a
foreign C function outside this compiler's control, never a second Lang
function this same `cgen` invocation must also emit.

### Design `[INFER]`, extending existing precedent

- **Calling convention.** `[INFER]` Given the executable shapes in scope
  today (`Byte`, `Buffer` per `linearProbeInput`, `core_test.go:255-264`)
  are small and currently passed by value (or by pointer for the
  `restrict`-qualified borrowed-view lowering, `emitLinearBorrowedByPointer`,
  line 460), the lowest-risk first cut is: **ordinary C value-parameter
  calling convention for owned/copy parameters, pointer parameters for
  borrowed views** — i.e., reuse the exact same by-value-vs-by-pointer
  fork `cgen` already implements for a single function's own parameter
  (`selectsByPointerLowering`/`selectsByPointerLoweringSharedOnly`,
  `cgen.go:33-38`), applied uniformly to every call argument, not just the
  entry function's own parameter.
- **sret/byval for aggregates.** `[READ]` The one real aggregate in the
  codebase today is `ForeignContract.Layout`/`RecordLayout`
  (`core.go:79-116`) — the `{ok, value}` two-field by-value ABI result
  struct `emitLinearForeign`'s `emitLinearForeignOutputSupport` always
  generates for a fallible foreign call's return. `[INFER]` `OpCall`
  returning a non-scalar value should follow the same precedent: an
  explicit named C struct return by value at `-O0`, which Clang is free to
  convert to the platform's `sret` (struct-return-via-hidden-pointer) ABI
  internally at `-O3` — this compiler should never *hand-emit* an `sret`
  pointer parameter itself; that is exactly the kind of ABI-shape claim
  `native/foreign_retained.go` and the sanitizer/zero-attribute discipline
  (`BannedOptimizerAttributes`, `cgen.go:2017`) exist to keep out of
  hand-authored C, letting the real Clang ABI define ABI rather than the
  generator asserting one.
- **How ownership transfer at a call maps to C.** `[READ]` Every existing
  ownership operation already lowers to "C value assignment... [that] makes
  no ABI or zero-copy claim" (the file-level comment at `cgen.go:1598`,
  emitted verbatim into every generated file). `[INFER]` A move-by-argument
  `OpCall` should lower identically — plain C value copy into the callee's
  parameter, with the *authority* transfer being a purely typed-core-level
  fact that C's type system knows nothing about, exactly as `OpMove`
  already does intraprocedurally (`case core.OpCopy, core.OpMove, ...:`
  share one emission branch today, `cgen.go:1652`, distinguished only by
  the emitted comment label, never by different C).
- **Which LLVM/Clang attributes may be emitted, and from which checked
  facts.** `[READ]` The existing precedent is exact and load-bearing:
  `restrict` is the *only* justifiable optimizer attribute today
  (`JustifiableAttributes`, `cgen.go:2058`), emitted *exclusively* when
  `check`/`corevalidate` have proven no-alias for that specific parameter
  (`EmittedAttribute{Attr: "restrict", ..., JustifiedBy: loanID}`,
  `cgen.go:732`), and every other candidate optimizer attribute is
  explicitly *banned* (`BannedOptimizerAttributes`, listing `"restrict",
  "noalias", "nothrow", "__attribute__((malloc))", "nonnull",
  "returns_nonnull"`, `cgen.go:2017`) unless justified the same way. `[INFER]`
  For calls, this means: a call argument may only carry `restrict` if the
  *callee's declared signature* (its `FunctionSignature`'s ownership
  requirement, §2) plus the caller's own proven no-alias fact jointly
  justify it — never inferred from "the callee looks like it doesn't alias
  its argument" by reading the callee's body (source-blind, cross-function,
  matching the whole architecture's posture). `nothrow`/`_Noreturn` remains
  reserved for the one existing exemption (`lang_defect`, `cgen.go:1696`) —
  a call to an ordinary Lang function must never be marked `nothrow` unless
  the language actually has no interprocedural nonlocal-exit path for it
  (Codename Lang's `setjmp`/`longjmp` nonlocal-exit pad,
  `internal/compiler/native/foreign_nonlocal.go`, exists precisely because C
  boundary calls *can* exit nonlocally — the same caution applies once
  Lang-to-Lang calls exist inside a function that itself sits between two
  foreign-call landing pads).
- **How `-flto` changes what must be proven.** `[READ]` M001's `-flto`
  discipline already proves attributes are "not-inert" — the PROJECT.md
  Context section credits it with catching "real `-O3`/LTO/`restrict`
  divergences" and the `-O3` tier is explicitly "`-flto`-proven-non-inert."
  `[INFER]` Once calls exist, LTO can *inline across the new call
  boundary* and thereby expose any `restrict`/no-alias claim that was only
  locally true (true within the caller's own body) but globally false
  (false once the callee's actual aliasing behavior is visible to the
  optimizer post-inlining) — this is a strictly *harder* version of the
  existing single-function `false_no_alias` engineered negative control
  (Phase 5, "an actual `interpreter == -O0 (2) != -O3 (7)` divergence on
  Apple clang 21"). M002's "interprocedural `-O3`/LTO equivalence" charter
  item is precisely the claim that this new, larger attack surface still
  holds, and it needs its own interprocedural engineered negative control
  (a function that would falsely justify `restrict` only if the checker
  failed to consult the callee's real declared contract), not merely a
  rerun of the existing intraprocedural one.

---

## 7. `Result` with Payloads

`[READ]` `core.DataType.Alternatives` is `[]string` today
(`core.go:18-23`) — bare nullary tags, no payload. D-04-30
(`04-DEBT.md:85-111`) records the deferred design explicitly and already
names the one-way-door shape: "add a sibling `alternative_details
[]Alternative omitempty` field keyed by name — additive to `core.DataType`,
never a shape change to the existing `Alternatives []string` field." D-04-04
"fixed Phase 4's failure representation as a two-successor control-flow edge
in core, not a storable value: no `Result` type, no generics, no
payload-carrying alternatives" — the error payload today "rides the `err`
edge as an ordinary nullary ADT," and the ok payload is "an ordinary place
on the ok successor block." Critically, D-04-30's own text warns: "a
storable `Result` value bolted onto an edge-based core is a genuine
re-lowering of the failure representation, not an extension of it" — the
cost of adding this is explicitly *not* incremental.

### Layout in the three engines `[INFER]`, following D-04-30's own additive
plan

- **Core IR:** `core.DataType` gains `Alternatives []string` (unchanged,
  frozen per D-13/D-04-23) plus a new sibling `AlternativeDetails
  []Alternative` (omitempty), where `Alternative{Name string, PayloadType
  string}` names each tag's carried type (or empty for a still-nullary
  tag, so existing ADTs like the Phase 4 failure ADTs stay byte-identical).
  `core.LinearOperation` needs a way to represent "construct a payload-
  carrying alternative" and "match/destructure a payload out of one" — most
  naturally as two new fields on the existing operation shape (a
  `PayloadSourceID`/`PayloadTargetID` pair, following the exact precedent
  `OkEdgeID`/`ErrEdgeID`/`ErrTargetID` already set for `OpForeignCall`:
  additive, omitempty, populated only on the relevant kind) rather than a
  wholly new `LinearOperation` shape.
- **Interpreter:** `interp`'s `values map[string]string` model is
  string-keyed and currently only carries scalar-ish values (bytes/buffers
  per the probe inputs). A payload-carrying alternative needs a real tagged
  representation in this map's value type — `[INFER]` the least invasive
  extension keeps `values` string-keyed by place ID but widens what a
  "value" *is* (a small tagged struct: `{tag string, payload string}` or
  similar), which every existing `case core.OpCopy/OpMove/...:` arm must
  still round-trip unchanged for non-`Result` types (an additive change to
  the value representation, not a new dispatch case, so it does not add a
  *seventh* case to the six sites' switches — only `Result`-specific
  construct/match need new cases).
- **C17:** the natural lowering is the same tagged-union-by-value shape
  already established for the two-field `{ok, value}` ABI struct
  `emitLinearForeignOutputSupport` generates for fallible foreign calls
  (`core.go:79-116`'s `RecordLayout`/`LayoutField`, and
  `cgen.go:1383-1395+`) — a C `struct { <enum> tag; union { ... } payload;
  }` (or, given the zero-attribute/no-hand-authored-ABI-tricks discipline,
  a struct with an explicit per-alternative field set rather than a real C
  `union`, to keep every byte's provenance checker-derivable and avoid
  reintroducing the kind of unproven-ABI-claim risk the `restrict`
  discipline exists to prevent).

### Interaction with the affine drop obligation

`[READ]` Codename Lang has no explicit `OpDrop` operation
(`core.AllOperationKinds()` has none) — "drop" is a *type-level ability*
(`core.AbilityDrop`, `internal/compiler/ability/ability.go`), not a runtime
op; the affine discipline is enforced by requiring every place be moved or
copied by function end, not by an explicit destructor call. `[INFER]` A
payload-carrying `Result` alternative composes this exactly the way any
other aggregate composes today: the alternative's payload place is subject
to the same move/copy-by-end-of-function admission rule as any other place,
and *matching* a `Result` (destructuring which alternative it is and
extracting the payload) must produce a fresh, independently-tracked place
for the extracted payload — mirroring how `checkBranch`'s existing bare-ADT
match already binds one place per arm today, just now with a real payload
value flowing into that place rather than nothing. The one genuinely new
hazard: a payload that is itself a resource (a Phase 4 fallible-foreign-
acquisition-derived value) flowing through a `Result` alternative and *not*
matched on some path — this is a new admission question for `check`/
`corevalidate` (does every path either consume the payload or is the
alternative never matched?), not a new mechanism; it reuses the exact
same reverse-order-release admission machinery Phase 4 already built
(`releaseAllocatorMismatch`, `checkReleaseOrder`), extended to a payload
living inside a matched alternative instead of a bare acquired place.

---

## 8. Build Order

`[INFER]`, reasoned from dependencies actually present in the code read
this session, and from the retrospective's own recorded failure modes
(three gate failures sharing "a green test whose reachable input space
omitted the hard case"; two liveness derivations coexisting for two phases;
a quadratic fix landing in only one of two peers).

### Phase-shaped sequence

**Stage 0 — Signature/summary artifact (prerequisite for everything else).**
Extend `core.FunctionSignature`/`core.Interface` (§2, Design C) with the
ownership-requirement-per-parameter and full return-origin contract needed
to admit a call without opening a callee body. This is a pure data-shape
and single-producer (`originvalidate.BuildInterface`-style) change with no
dependency on call semantics existing yet — it can be built and tested in
isolation against the existing Phase 4 corpus (calls don't need to exist to
validate the summary shape is correct for every existing function). **Land
first; nothing else can be built without it**, because every one of the six
sites' `OpCall` handling reads this summary rather than a callee body.

**Stage 1 — `core.OpCall` registration + `check` producer path (Site 1).**
Add `core.OpCall` to `AllOperationKinds()` (this alone forces every one of
the six sites' exhaustive switches to fail to compile/fail the registry test
until each adds a `case`, which is *useful* — it converts "six sites, easy
to forget one" into "six compile/test failures, impossible to forget one").
Flip `resolveForeignStep`'s `core.call_target_not_foreign` refusal into a
real Lang-callee producer path in `check.go`, gated on `callable ⊆
publishable` read against Stage 0's summary for the callee. **Depends on:**
Stage 0.

**Stage 2 — Call graph package + cycle refusal (§3).** New
`internal/compiler/callgraph` package, `pathoracle`-shaped: reads
`core.Program`'s `OpCall` operations (now real after Stage 1), builds the
adjacency relation, refuses cycles with a typed, named error. Consumed by
`check` (pre-admission refusal) — the earliest point a cyclic program can be
rejected cheaply, before any interprocedural liveness work runs on a
malformed graph. **Depends on:** Stage 1 (needs real `OpCall` operations to
walk).

**Stage 3 — `corevalidate` independent re-derivation (Site 2) + independent
cycle re-check.** New `case core.OpCall` in both `replayStraightLine` and
`replayBlocks`; independent call-graph cycle re-derivation from the checked
core artifact (never trusting `check`'s Stage 2 verdict). **Depends on:**
Stage 1 (needs the operation shape finalized) and Stage 0 (needs the summary
shape finalized so `corevalidate`'s independent admission check reads the
same contract `check` reads, without sharing derivation code). **Can
proceed in parallel with:** Stage 2, once Stage 1 is done — `corevalidate`
doesn't need `callgraph` to exist as a package; it independently re-derives
cycle refusal itself, per §3.

**Stage 4 — Interprocedural loan liveness in both admission layers (§4).**
The M002 charter's own single largest, most novel item — signature-mediated
liveness extension to `check`'s `loanLivenessFixpoint`-adjacent admission
and `corevalidate`'s `recomputeLoanEndpoints`-adjacent admission,
independently, per §4 Design 1. **Depends on:** Stage 0 (summary must carry
enough origin/ownership detail), Stage 1 (needs real call sites to attach
liveness facts to), Stage 3 (extends the same switch cases Stage 3 lands).
**This is the single highest-risk integration point in the whole
milestone** — it is exactly the shape of failure the retrospective names
three times (Phase 2, 3, 4 each cost extra remediation rounds from "a green
test whose reachable input space omitted the hard case"), and it is the one
explicitly-carried debt item (D-03-02) the milestone exists to close. Budget
a mid-phase gate here specifically, mirroring Phase 3's own mid-phase-gate
precedent for the intraprocedural version of this exact problem.

**Stage 5 — `originvalidate` (Site 3) + `pathoracle` (Site 6) independent
re-derivations.** Both extend an existing partial-switch site with an
`OpCall`-shaped hop mirroring the already-proven `OpForeignCall` pattern
(§1, Sites 3 and 6). Lower risk than Stage 4 because the mechanism
(mirror an existing, working `OpForeignCall` case) is already proven in
this codebase, not novel. **Depends on:** Stage 0, Stage 1, and ideally
Stage 4 (so the origin/liveness facts they re-derive are stable) — but the
two packages' own independence from each other (`pathoracle` must not
import `originvalidate` or vice versa) means **these two can be built in
parallel** with each other, and largely in parallel with Stage 4 once
Stage 1 is done, since both read the summary/operation shape, not the
liveness engine's internals.

**Stage 6 — Interpreter call stack (§5, Site 4).** New `Frame`
type, bounded depth, per-frame ownership partitioning, cross-frame
drop/cleanup ordering. **Depends on:** Stage 1 (real `OpCall` operations),
Stage 4 conceptually (the interpreter must agree with the admission layers
about what's legal, though it can be built against Stage 1's operations
alone and validated against Stage 4's corpus once both exist) — **can
proceed in parallel with Stage 5**, since interp doesn't depend on
originvalidate/pathoracle. This is the deterministic-oracle authority path;
land and stabilize it before Stage 7, since Stage 7's cgen work needs a
trusted interpreter oracle to differential-test against (the whole
five-axis-comparator discipline depends on interp being right first).

**Stage 7 — C17 multi-function emission + call lowering (§6, Site 5).** The
largest single-site cost (§1): first the prerequisite — `cgen.Emit`/
`EmitNative` must stop refusing `len(program.Functions) != 1` and grow real
multi-function C emission (every non-entry function becomes a real C
function definition) — *then* `case core.OpCall` at every existing switch
site, ideally refactored into one shared `emitCall` helper rather than five
independent copies. **Depends on:** Stage 6 (must differential-test against
a working, trusted interpreter — this project's whole native-equivalence
discipline requires the oracle to exist first), Stage 4 (restrict/no-alias
justification for call arguments depends on the interprocedural liveness
facts being real). **This is the second-highest-risk integration point** —
it is genuinely new infrastructure (multi-function C emission has never
existed in this codebase), not an incremental case-arm addition, and it is
where the `-flto`/interprocedural-`-O3`-equivalence charter item lives.

**Stage 8 — Cross-function differential/exhaustive-dispatch rebuild.**
Rebuild Phase 3's exhaustive loan-endpoint differentials cross-function, and
extend `control:kind.exhaustive_dispatch`'s fixture set (both
`core_test.go`'s in-process control and `session.go`'s CLI-observable lane)
to include `core.OpCall`-carrying fixtures at every one of the six sites,
plus a genuinely new interprocedural engineered negative control (§6, the
harder `-flto` version of `false_no_alias`). **Depends on:** every prior
stage — this is the milestone-closing verification stage, matching Phase
5's own role in M001 (adversarial native evidence landing after the
semantic spine, not alongside it).

**Stage 9 — `Result` payloads (§7).** Structurally independent of every
prior stage's *call* machinery — D-04-30 was deferred for Phase 4 reasons
unrelated to calls, and nothing about payload-carrying alternatives strictly
requires `OpCall` to exist first. **Can be built in parallel with Stages
1-8**, by a different work-stream, provided it lands its own `case`
additions at the same six sites without colliding with the `OpCall` case
additions (a real scheduling risk if both land in the same plan/file at the
same time — the retrospective's D-08 "fold into whichever plan already
touches the file" note argues for landing `Result` payload cases in the
*same* plans that touch each site for `OpCall`, not as a separate pass over
the same six files, to avoid the "two coexisting laws" / repeated-file-touch
inefficiency the retrospective flags for M001).

### Rationale summary

- **Prerequisite chain:** Stage 0 (summary) → Stage 1 (`OpCall` producer) →
  {Stage 2 (call graph) ∥ Stage 3 (corevalidate)} → Stage 4 (interprocedural
  liveness, the hardest item) → {Stage 5 (origin/pathoracle) ∥ Stage 6
  (interp)} → Stage 7 (cgen, needs interp as oracle) → Stage 8 (closing
  verification).
- **Parallelizable:** Stage 2 ∥ Stage 3 (once Stage 1 lands); Stage 5 ∥
  Stage 6 (once Stage 4 lands); Stage 9 (`Result` payloads) against
  everything, coordinated only at the file-touch level per site.
- **Riskiest integration points, named explicitly:** Stage 4
  (interprocedural loan liveness — the milestone's own lead charter item and
  the one explicitly-carried debt item) and Stage 7 (cgen multi-function
  emission — new infrastructure, not an incremental change, and the
  dependency for the interprocedural `-O3`/LTO equivalence claim). Both
  deserve their own mid-phase gates, following the Phase 3 precedent the
  retrospective already validates as effective for exactly this failure
  shape.

---

## Sources

- `[READ]` `~/projects/ai-lang/.planning/PROJECT.md`
- `[READ]` `~/projects/ai-lang/.planning/RETROSPECTIVE.md`
- `[READ]` `~/projects/ai-lang/.planning/ROADMAP.md`
- `[READ]` `~/projects/ai-lang/.planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-DEBT.md`
- `[READ]` `~/projects/ai-lang/.planning/milestones/M001-MILESTONE-AUDIT.md`
- `[READ]` `~/projects/ai-lang/internal/compiler/core/core.go`
- `[READ]` `~/projects/ai-lang/internal/compiler/core/core_test.go`
- `[READ]` `~/projects/ai-lang/internal/compiler/check/check.go`
- `[READ]` `~/projects/ai-lang/internal/compiler/corevalidate/corevalidate.go`
- `[READ]` `~/projects/ai-lang/internal/compiler/originvalidate/originvalidate.go`
- `[READ]` `~/projects/ai-lang/internal/compiler/interp/interp.go`
- `[READ]` `~/projects/ai-lang/internal/compiler/cgen/cgen.go`
- `[READ]` `~/projects/ai-lang/internal/compiler/pathoracle/pathoracle.go`
- `[READ]` `~/projects/ai-lang/internal/compiler/cache/cache.go`
- `[READ]` `~/projects/ai-lang/internal/compiler/ability/ability.go`
- `[READ]` `~/projects/ai-lang/internal/compiler/session/session.go` (lines 2480-2580, exhaustive-dispatch lane)
- `[EXT]` rustc dev guide, borrow checking / MIR borrowck (signature-mediated interprocedural borrow facts): https://rustc-dev-guide.rust-lang.org/borrow_check.html , https://rustc-dev-guide.rust-lang.org/mir/borrowck.html — MEDIUM confidence, cited from prior knowledge of rustc's architecture, not re-fetched and re-verified against current source this session.

---
*Architecture research for: Codename Lang M002 (Interprocedural Semantic Spine)*
*Researched: 2026-09-08*
