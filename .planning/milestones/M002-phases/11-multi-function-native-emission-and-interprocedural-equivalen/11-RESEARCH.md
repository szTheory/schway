# Phase 11: Multi-Function Native Emission and Interprocedural Equivalence - Research

**Researched:** 2026-09-11
**Domain:** Multi-function C code generation (cgen), interprocedural differential
testing, HDD program reduction across function boundaries, `restrict`/LTO
semantics, cache soundness for interprocedural facts.
**Confidence:** HIGH for code-site claims (all re-read this session); MEDIUM for
Clang/LLVM `restrict`/LTO external semantics (cited, not independently compiled
in this research session — Q-04/Q-07 in 11-CONTEXT.md are the falsification
steps the plan must still run); LOW/ASSUMED nowhere — this phase's 11-CONTEXT.md
already ran a seven-way adversarial research fan-out and this document's job is
to re-verify its load-bearing code claims against the actual repository, not to
re-litigate the decisions.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

All 50 decisions in `11-CONTEXT.md` (D-11-01 through D-11-50) are locked
implementation decisions from a seven-way parallel research fan-out with an
adversarial pass each. They are reproduced by reference, not copied verbatim
here (the file is 843 lines) — **the planner MUST read `11-CONTEXT.md` in full
before planning**, per that file's own `<canonical_refs>` instruction. The
headline structural decisions, restated for a planner who reads only this
document first:

- **D-11-01/D-11-02/D-11-03/D-11-04:** `cgen` gains an additive `emitProgram`
  path in a new `cgen_program.go`, reached only when `len(Functions) != 1`. The
  six existing single-function whole-TU emitters (`emitLinear`, `emitLinearBorrowedByPointer`,
  `emitLinearBorrowedByPointerPlain`, `emitLinearForeign`, `emitBranch`,
  `emitMatch`) are left byte-for-byte untouched and deleted only in Phase 12
  (declared debt, `PHASE-11-DEBT.md`, gated on Q-05's N=1 convergence
  differential). Body emitters emit statements/events only; the new TU
  assembler owns includes/typedefs/support/ledger/prototypes/`main`. Exactly
  ONE `emitCall` helper writes every Lang-to-Lang call, called from exactly two
  places; it never mints an attribute.
- **D-11-05/D-11-06:** Entry point resolves via new `callgraph.EntryFunction(program)`
  (unique in-degree-zero root over `buildAdjacency`); zero-or-many roots is a
  named fail-closed refusal. `session.go`'s three run sites switch from
  `Functions[0].Name` to this resolver. Do NOT add an `Exported`/`EntryFunctionID`
  field to `core.Program` this phase (would move frozen `/0` bytes) — assert
  agreement with the single `ast.Export` instead.
- **D-11-07:** `singleForeignFunction`/`singleManifestFunction` stay
  single-function-scoped; with N functions they refuse two foreign contracts
  rather than picking one.
- **D-11-08:** Two-tier name allocation — one `cNames` for all type/function
  names (deterministic via `callgraph.Order`), fresh per-function `cNames` for
  locals seeded with reserved list + all allocated globals.
- **D-11-09/D-11-10 (SCOPE REDUCTION, evidence-driven, flagged for human
  attention):** Phase 11 emits **zero** call-boundary alias attributes,
  permanently for this phase (not a waypoint). Rationale: Lang functions take
  exactly one parameter, no globals, no callbacks, no address-escaping foreign
  contracts, and `check` already refuses passing one place to two calls — two
  pointers to one object cannot exist across a Lang call boundary in M002, so a
  call-boundary `restrict` promises about a hazard that cannot exist. NAT-05 is
  satisfied by an explicit, visible empty set with a generated comment citing
  this decision — a requirement weakened by evidence, must be stated as such in
  `11-VERIFICATION.md`.
- **D-11-11/D-11-12/D-11-13:** The `EmittedAttribute` discharge-pair design
  (`justified_by`/`discharged_by`) is designed and recorded but NOT built this
  phase. `lang.attributes/0` is the schema name reserved for it, minted only
  when D-11-09 is revisited. A pre-existing gap (`evidence.go:233` binds
  `ForeignDigest` only under `hasForeignContract`, so a pure-Lang `restrict`
  claim is not digest-bound into evidence today) is disclosed, not fixed.
- **D-11-14 through D-11-19:** The mid-phase gate is a conjunction with a
  second independent knower: `cgen.ScanForBannedAttributes` returns empty over
  emitted artifacts (reads output bytes) AND `check` independently reports
  N≥1 functions that *would have* carried `restrict` (reads the program) AND
  structural floors (≥2 functions, ≥1 call edge) AND interpreter ≡ `-O0`.
  Suppression is token-local (one `Fprintf` gains one interpolated qualifier),
  diff-locality-tested, permanent (not a temporary flag), and anti-vacuity
  mutation-killed (re-running under the justified profile must go red).
- **D-11-20/D-11-21/D-11-22 (ROADMAP AMENDMENTS REQUIRED, flagged for human
  attention):** Criterion 3's divergence signature `interpreter == -O0 != -O3`
  is factually wrong for the interprocedural sequel — measured on this host
  (Apple clang 21.0.0, arm64-apple-darwin25.6.0) it is
  `interpreter == -O0 == -O3 != (-O3 -flto)`. "Fails red before its fix"
  presupposes a constructible shipped defect that D-11-09 makes inexpressible
  in M002. A toolchain-pinned re-measurement clause is required (exploitation
  is non-monotonic in inlining aggressiveness: `-O1` correct, `-O2`/`-O3`
  single-TU diverges, full-LTO 2-TU inlines everything and folds back to
  correct).
- **D-11-23/D-11-24/D-11-25/D-11-26/D-11-27:** NAT-07's control is a
  coordinated two-site mutation across two/three translation units, extending
  `AliasFactMutationRunner`, shipped as a **hand-written-C matrix test** in
  `native_lto_test.go`, pinned to the recorded Clang version — built BEFORE
  `cgen` learns to write more than one source file. The production corpus
  stays single-TU (`native.Runner`'s single `program.c` is NOT widened this
  phase), so `-flto` is inert by construction for Lang-to-Lang code in
  criterion 2's corpus; the control's non-inertness is borrowed from the
  existing foreign-TU boundary. This must be disclosed in writing in
  `11-VERIFICATION.md`, not implied.
- **D-11-28 through D-11-37 (`reduce`, QLT-05):** Never inline (destroys the
  defect class Phase 11 is about). `reduce` gains exactly two new whole-program
  moves — `drop-call-site` (OpCall→OpCopy) and `drop-orphan-function` — run
  FIRST, before the five existing per-function moves become per-function loops.
  `Reduce` takes a `Seed{Program, EntryFunctionID}`. `MaxReductionAttempts`
  becomes a derived bound
  (`AttemptsPerFunction*len(Functions) + callSiteCount`, `AttemptsPerFunction=64`).
  QLT-05 re-verification is strict field equality (`Axis`, `EnginePair`,
  `OperationID`, no `CausalRole` fallback), never `Interesting`-equality.
  `RefusedShapes()` ships as an enumerated, test-asserted refusal register.
  `foreignCallSequenceFor` returning `nil` for multi-function programs is a
  silent-slippage hole that MUST be fixed (widen across `callgraph`
  reverse-postorder + add a dynamic second knower from the `-O0` event stream).
  Flaky-predicate tolerance is explicitly not built.
- **D-11-38 through D-11-42 (QLT-06, caching):** Cache **nothing new** this
  phase — `ArtifactSpec.FixtureSource` already hashes the whole `.lang` source
  file, which strictly dominates any call-graph-closure key in a
  single-compilation-unit language; a closure key would be soundness-loosening,
  not soundness-improving. QLT-06 splits a/b (flagged for human attention):
  06a Complete at Phase 07 (`ClosureDigest` mutation-killed by two knowers);
  06b Complete-by-abstention at Phase 11, discharged by a structural test that
  `cache.DeclaredInputNames()` stays at seven names and `internal/compiler/cache`
  does not import `core`/`originvalidate`. S-006's eviction figures
  (100%/92%/43%) are a model upper bound, not a measurement — do not re-quote
  as measured. A LIVE CACHE SOUNDNESS HOLE is disclosed: `internal/compiler/cgen/*.go`
  is not a declared cache input, so editing `cgen` with unchanged `.lang`
  fixtures can serve a stale binary — must be fixed before any cache work
  (Q-02 confirms it).
- **D-11-43 through D-11-50 (QLT-03, shape register):** New
  `qlt03_shape_register.json`, reusing `qlt01.go`'s skeleton shape but never
  extending `qlt01_registry.json` itself. Cells are enumerated mechanically as
  an axis product (unclassified cell defaults to `gap` and fails); only
  `unreachable` rows are hand-written, each naming a falsifier test. Five axes:
  edge topology; argument mode across the edge; callee body shape;
  returned-value origin class; composition-depth bucket. Full enumeration of
  argument-mode × return-origin (20 cells) + pairwise over the rest (~40 rows
  target), not the full 2,560-cell product. Negative claims carry one
  admissible proof mechanism per class (`grammar`/`refused`/`generator`/
  `bounded_exhaustive`/`gap`) — a corpus-absence assertion alone is rejected as
  theater. `witness` is not boolean — a thinly-reached cell (`reached_thin`)
  counts as a gap for gating. The register is the honest home for disclosed
  Phase 10 trust gaps, headlined by the `peerDeriveOriginFacts` `OpCall` gap.

### Claude's Discretion (flagged for human attention at the planning checkpoint)

1. D-11-09/D-11-10 — emitting zero call-boundary alias attributes materially
   weakens what NAT-05 claims, even though the evidence for it is strong.
2. D-11-20/D-11-21/D-11-22 — three roadmap amendments to criterion 3.
3. D-11-39 — splitting QLT-06 into a/b, one of which is satisfied by
   abstention.

### Deferred Ideas (OUT OF SCOPE)

- Deleting the six single-function emitters — Phase 12, gated on Q-05.
- Call-boundary alias attributes and the `EmittedAttribute` discharge-pair
  schema — designed, built only when D-11-09 is revisited; no construct on the
  current roadmap creates two pointers to one object.
- `lang.attributes/0` — minted only alongside the discharge schema.
- Per-function translation units / widening `native.Runner` beyond one
  `program.c` — only if criterion 2's LTO tier must be non-inert over the
  corpus itself.
- A closure-keyed native cache — revisit only if Lang gains separate
  compilation units (M003).
- Reviewing the D-09-51 negative-control verdict flip — dissolved as a Phase 11
  precondition by D-11-09, but still open, unreviewed.
- `peerCalleeFrameDrained`'s forward-pass ordering assumption and the
  `callgraph` import asymmetry — both remain review-enforced, not
  mechanism-enforced. NAT-06 leans on the latter.
- Flaky-predicate tolerance in the reducer — explicitly not built.
- Arithmetic, iteration, strings, arrays — on no roadmap at all.

</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| NAT-04 | `cgen` emits multi-function C17; `Emit`/`EmitNative` no longer refuse >1 function | Guard inventory (below) pinpoints exactly which `cgen` guards move; `emitProgram`/`emitCall` design in D-11-01–D-11-04 is the mechanism. `session` and `reduce` guards (32 total, not 2) are the wider surface that must also move for the program to be *runnable*, not merely emittable. |
| NAT-05 | Every call-boundary alias/capture promise names its derived checked fact | D-11-09 shows the honest answer is an emitted, explained empty set (zero promises), not a populated promise set — see "Alias attributes" section below and the `restrict`/LTO literature review. |
| NAT-06 | Interpreter/`-O0`/`-O3`/`-O3 -flto` agree on the interprocedural corpus, five-axis comparator | Five axes verified at `session_phase5_compare.go:19-25`. Entry-resolver unification (D-11-05) is what keeps oracle and binary from disagreeing about which function *is* the program. NAT-06 leans on the `callgraph` import-asymmetry independence gap carried from Phase 10 (see Phase 10 carry-forward section). |
| NAT-07 | Engineered composition-only negative control proves the `-O3`/LTO tier non-inert | `false_no_alias` precedent studied; `restrict`/LTO/inlining literature cited; D-11-20–D-11-27 give the corrected divergence signature and the hand-written-C matrix design (Q-04). |
| QLT-03 | Call-graph-shape reachability register, naming provably-unreached shapes | Register design (D-11-43–D-11-50); `testsupport.GenerateCallGraphCorpus`/`CallGraphCorpusShapes` verified to have five shapes with **zero borrow ops** (Phase 09 finding) — the predicted headline negative row. |
| QLT-05 | HDD reducer output on multi-function programs re-verified to reproduce the same property | `reduce.Reduce`/`ProjectSource` guards verified line-for-line; the two new moves and the strict-equality re-verification design in D-11-28–D-11-37; blocked/narrowed by the `peerDeriveOriginFacts` `OpCall` gap per the Critical Path. |
| QLT-06 | No interprocedural fact cacheable without a callee-invalidates-caller test; keys derive from call-graph closure | `cache.DeclaredInputNames()` verified as exactly seven names, none naming `cgen`; `ClosureDigest` verified absent from `internal/compiler/cache/`. D-11-38–D-11-42 show why "cache nothing new" is the sound answer and name the live `cgen`-as-undeclared-input hole. |

</phase_requirements>

## Summary

Phase 11's job is to make a *checked* multi-function Lang program *executable*
and *provably equivalent* across four engines (interpreter, `-O0`, `-O3`,
`-O3 -flto`) — closing the gap Phases 07-10 opened (calls are admitted and
checked but `lang run` refuses them on both engines today). This research
re-verified, line-by-line, every code claim `11-CONTEXT.md`'s seven-way research
fan-out made, and every one checked out: the guard inventory is exactly 32
non-test `len(Functions) != 1` sites across 6 files in 3 packages (`session` 26,
`cgen` 4, `reduce` 2) — not the two `cgen` entry points the roadmap's criterion 1
names — plus two additional single-function assumptions the grep pattern does
not catch (`singleForeignFunction`/`singleManifestFunction`). `corevalidate.peerDeriveOriginFacts`
(corevalidate.go:2407) genuinely has no `core.OpCall` case in its `switch` —
confirmed by reading the function body — which is the single Phase 10 carry-forward
fact that four independent researchers hit from four directions and that this
plan must branch on (Q-01 decides which branch). `cache.DeclaredInputNames()`
is genuinely exactly seven names with no `cgen` entry, confirming the live cache
hole. `reduce.Reduce` and `reduce.ProjectSource` genuinely hard-refuse
`len(Functions) != 1`, confirming the reducer is a first-class subject of this
phase, not a downstream consumer.

**Primary recommendation:** Follow `11-CONTEXT.md`'s 50 locked decisions as
written. The single highest-leverage scope decision is D-11-09: Phase 11 emits
**zero** call-boundary alias attributes, because Lang's one-parameter,
no-globals, no-callback, no-address-escaping-foreign-contract shape makes a
call-boundary `restrict` promise about a hazard that provably cannot exist in
M002 — this converts NAT-05 from "prove a nontrivial promise" into "prove an
honest absence," which is achievable, testable, and does not stake correctness
on LLVM's scoped-noalias machinery (whose documented fragility across LLVM
versions is real and cited below). The mid-phase gate must ship BEFORE any call
lowering with attributes is even considered, and the tracer-first sequencing
below (thinnest end-to-end slice → gate → reducer/cache/register) is the
recommended execution order.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Multi-function C emission (`emitProgram`, `emitCall`) | Compiler backend (`cgen`) | — | `cgen` is the sole C-emission boundary; D-11-01 through D-11-04 keep it additive |
| Entry-point resolution | Compiler frontend/graph (`callgraph`) | Compiler driver (`session`) | `callgraph.EntryFunction` derives the fact once; `session`'s three run sites and `cgen`'s assembler both consume it so they never disagree |
| Interprocedural equivalence proof | Verification harness (`session`) | Native toolchain (`native`), Interpreter (`interp`) | `session_phase5_compare.go`'s five-axis comparator already owns cross-engine comparison; it must be widened to drive multi-function programs, not reinvented |
| Alias/capture attribute justification | Compiler backend (`cgen`) emission + Independent validator (`corevalidate`) | — | `corevalidate.ValidateEmittedAttributes` is the proven second knower for any attribute claim (D-05-01 precedent); this phase emits none, so the "knower" role narrows to confirming the empty set is honest |
| HDD reduction over multi-function programs | Compiler tooling (`reduce`) | Verification harness (`session`, mismatch re-verification) | `reduce` operates on `core.Program` only; it is upstream of `session`'s re-verification lane, which is the second knower for QLT-05 |
| Cache soundness for interprocedural artifacts | Build cache (`cache`) | Verification harness (`session_phase6_verify.go`) | `cache` package is a dependency-free leaf; declaring `cgen` as an input is a `cache`-side fix even though the hole is exposed by `session`'s consult site |
| Call-graph shape reachability register | Verification harness (`session`) | Test corpus generator (`testsupport`) | The register is a `session`-owned audit artifact over `testsupport.GenerateCallGraphCorpus`'s declared production space |

## Standard Stack

This phase adds zero new external dependencies — confirmed against
`.planning/STANDING-VERDICTS.md`'s zero-external-dependency record, which
`11-CONTEXT.md`'s `<canonical_refs>` names as a hard constraint. Every
mechanism (call-graph traversal, HDD reduction, differential testing, caching)
extends existing in-repo machinery.

### Core

| Component | Status | Purpose | Why Standard (in this repo) |
|-----------|--------|---------|------------------------------|
| `cgen.emitProgram` (new) | To build | Whole-program C17 assembler for N>1 functions | Additive sibling pattern (D-04-20) already used four times in this file (`emitLinearBorrowedByPointer` etc.) |
| `callgraph.EntryFunction` (new) | To build | Deterministic, fail-closed program-entry resolution | Extends `buildAdjacency`/`Order`, already the proven acyclicity substrate (SEM-07) |
| `reduce`'s two new moves (`drop-call-site`, `drop-orphan-function`) | To build | Whole-program-level HDD moves, run before intra-function moves | J-Reduce's dependency-graph-closure pattern, already cited in this repo's own design vocabulary |
| Hand-written-C 4×3 LTO matrix test | To build | NAT-07's composition-only negative control | Extends `AliasFactMutationRunner` (`session_phase5_alias.go:87`), proven idempotent/fail-closed shape |

### Supporting

| Component | Status | Purpose | When Used |
|-----------|--------|---------|-----------|
| `cache.DeclaredInputNames()` gains an eighth name (`cgen_source` or similar) | To build | Closes the live cache-soundness hole (D-11-41) | Before any interprocedural cache work is even considered |
| `qlt03_shape_register.json` (new) | To build | QLT-03's mechanically-enumerated axis-product register | Session-owned audit artifact, sibling of `qlt01_registry.json`'s skeleton (not an extension of it) |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Additive `emitProgram` alongside six frozen single-function emitters | Refactor the six emitters into one parameterized multi-function emitter now | Rejected in D-11-01: risks moving `testdata/phase5/restrict_borrow.golden.c` and four pinned digests in the same diff that introduces calls, making digest movement unattributable. Phase 12 does the deletion once N=1 convergence is proven (Q-05). |
| Zero call-boundary alias attributes (D-11-09) | Emit `restrict`/noalias-shaped attributes now, closing NAT-05's prose literally | Rejected: no two pointers to one object can exist across a Lang call boundary in M002 (one parameter, no globals, no callbacks, no address-escaping foreign contracts, `check` already refuses aliased dual-call arguments) — the promise would be about a hazard that cannot exist, and would stake correctness on LLVM's scoped-noalias machinery, whose fragility is documented (Rust `noalias` history below). |
| Closure-keyed interprocedural cache (QLT-06's literal prose) | Wire `ClosureDigest` into `cache.Input` this phase | Rejected in D-11-38: `ArtifactSpec.FixtureSource` already hashes the entire `.lang` source file; Lang has one module, one file, no separate compilation units, so whole-program hashing strictly dominates any closure key — a closure key could only ever admit MORE hits on LESS evidence, a soundness-loosening change. |
| Full HDD with callee inlining across function boundaries | Never inline (D-11-28) | Inlining destroys the defect class Phase 11 is about — the call boundary is where the alias promise lives; inlining also forces ID renumbering, invalidating `Signature.OperationID` |

**Installation:** none — no new packages.

**Version verification:** N/A — no new packages. Toolchain identity for
NAT-07's control must be pinned and recorded (`clang --version` — this
research session's host reports Apple clang 21.0.0, arm64-apple-darwin25.6.0,
per D-11-20/D-11-22; the plan must re-capture this on its own execution host,
not assume this value is still current).

## Package Legitimacy Audit

Not applicable — this phase adds zero external packages (confirmed: no new
`go.mod` dependency is proposed anywhere in `11-CONTEXT.md`, and the
zero-external-dependency record in `.planning/STANDING-VERDICTS.md` is a named
hard constraint).

## Guard Inventory — Re-Verified This Session

Re-ran the roadmap's own command against the current tree:

```
awk '/^func /{f=$0;l=NR} /Functions\) != 1/{print FILENAME": "f}' $(find internal cmd -name '*.go' -not -name '*_test.go')
```

`[VERIFIED: awk output, run this session]` — **32 non-test matches** (55
including `_test.go` files), matching the roadmap's scope input exactly.
Per-file breakdown (non-test):

| File | Count | Functions |
|------|-------|-----------|
| `internal/compiler/session/session.go` | 15 | `TransposeReleaseOrder`, `Phase4CheckedProgram`, `RunInterpreter`, `interpreterInputs`, `RunNative`, `verifyOwnedCorpus`, `verifyBorrowedCorpus` (×9 distinct call sites inside one function body — the awk pattern counts each `func` header once but the guard recurs inside), `verifyForeignCorpus` |
| `internal/compiler/session/session_phase5_mismatch.go` | 4 | `foreignCallSequenceFor`, `causalChainFor`, `mismatchPredicate`, `ReduceSeededAliasMismatch` |
| `internal/compiler/cgen/cgen.go` | 4 | `Emit`, `EmitNative`, `emitLinear`, `emitBranchOperations` |
| `internal/compiler/reduce/reduce.go` | 2 | `Reduce`, `ProjectSource` |
| `internal/compiler/session/session_phase5.go` | 2 | `VerifyPhase5ControlsAndWork`, `phase5RunInterpreterO0O3LTOLane` |
| `internal/compiler/session/session_phase5_alias.go` | 1 | `VerifyAliasFalseNoAlias` |
| `internal/compiler/session/session_phase5_corpus.go` | 1 | `admitPhase5Candidate` |
| `internal/compiler/session/session_phase6.go` | 1 | `phase6RunCleanupInjectionLane` |
| `internal/compiler/session/session_phase6_verify.go` | 1 | `verifyPhase6NativeDifferentialLane` |
| `internal/compiler/session/session_phase7.go` | 1 | `phase07LinearProbeInput` |

**26 in `session` across 9 files** (the CONTEXT.md figure of "session 26"
counts the `Functions) != 1` occurrences project-wide inside the `session`
package across all its files — `session.go` itself contributes 15 by function
count, but re-grepping the raw pattern occurrence count, not function-header
count, inside `session.go` alone also surfaces the actual `verifyBorrowedCorpus`-style
repeated guard shape; the roadmap's "×9" annotation refers to the same guard
string recurring across nine near-identical `verify*Corpus` bodies). Distinguish
by category:

- **Must widen to make N>1 runnable at all (NAT-04's CLI-visible half):**
  `session.go`'s `RunInterpreter`, `RunNative`, `interpreterInputs` — these are
  the CLI `lang run` path itself. Confirmed: `[VERIFIED: internal/compiler/session/session.go:783]`
  `if len(program.Functions) != 1 {` inside `RunNative`, and the parallel guard
  in the interpreter path at line 759 (`if len(checked.Program.Functions) != 1`).
- **Must widen to make N>1 provable (NAT-06's equivalence proof, the harder
  half):** the Phase 5/6/7 verification lanes (`verifyBorrowedCorpus` ×9,
  `verifyOwnedCorpus`, `verifyForeignCorpus`, `VerifyPhase5ControlsAndWork`,
  `phase5RunInterpreterO0O3LTOLane`, `verifyPhase6NativeDifferentialLane`,
  `VerifyAliasFalseNoAlias`) — these ARE the five-axis equivalence machinery;
  widening them is materially harder than widening the CLI run path because
  each embeds fixture-shape assumptions (single input, single interpreted
  execution) that the five-axis comparator's per-fixture drivers assume.
- **Must widen or the phase cannot claim criterion 4 at all:** `cgen.go`'s
  `Emit`/`EmitNative` (line 35, 65) — the two the roadmap's criterion 1
  literally names — plus `emitLinear`/`emitBranchOperations`, which are
  reached only through the single-function dispatch and therefore never see
  `core.OpCall` today (both already carry a `core.OpCall` case that
  unconditionally errors "Lang-to-Lang calls are not supported by native
  emission this phase" — `[VERIFIED: internal/compiler/cgen/cgen.go:339-345]`
  `"D-07-39/A-02: OpCall is registered but not lowered by native emission this
  phase. Emit/EmitNative hard-fail on len(program.Functions) != 1 ... so this is
  forward hygiene for Phase 11's multi-function C emission, not a live gap."`).
- **Must widen or QLT-05 is unreachable:** `reduce.go`'s `Reduce` (line 117)
  and `ProjectSource` (line 637) — confirmed by direct read, both hard-error
  with a message naming the exact function count
  (`[VERIFIED: internal/compiler/reduce/reduce.go:117-119]`
  `"reduce: seed must carry exactly one function, got %d"`), and
  `[VERIFIED: internal/compiler/reduce/reduce.go:636-639]`
  `"expected exactly one function, got %d"` via `unsupportedProjection`.
- **Defensive, stay as-is:** the fixture-forging helpers deep in `session.go`
  (lines 1569-1996, e.g. `TestForeignManifestBytesUnchangedForPriorPhases`-adjacent
  fixture mutators) are test-fixture-construction helpers that intentionally
  assert their OWN single-function fixture shape — these are not part of the
  32-guard non-test count (they live in `_test.go` files, per the awk output's
  second run) and are correctly excluded already.

**Two additional single-function assumptions the grep does NOT catch**
(confirmed by direct read, matching D-11-07 exactly):
`[VERIFIED: internal/compiler/cgen/cgen.go:1868,1908]` `singleForeignFunction`
("returns the one function in program that declares a foreign contract, or an
error if none does") and `singleManifestFunction` ("returns the one function in
program that EmitForeignManifest has a lang.foreign/0 sidecar to say something
about"). Neither is gated by `len(Functions) != 1`, but both iterate assuming
at most one match matters; D-11-07's answer (refuse two foreign contracts
rather than pick one) is correct and requires no new schema.

## Standard Approach: `cgen` Today

`[VERIFIED: internal/compiler/cgen/cgen.go]`, full file read (2,217 lines).

**Dispatch shape:** `Emit`/`EmitNative` (lines 29-90) both: validate via
`corevalidate.Validate`, refuse `len(program.Functions) != 1`, then dispatch on
the single function's shape (`Match+Linear` → `emitBranch`; `Linear` with
`len(Blocks) > 0` → `emitLinearForeign`; `Linear` selecting by-pointer lowering →
`emitLinearBorrowedByPointer`/`emitLinearBorrowedByPointerPlain`; else
`emitLinear`; else `emitMatch`). Each of the six emitters writes a **complete
translation unit including its own `int main(int argc, char **argv)`** — this
is exactly why D-11-01 says there is no per-function seam to loop over; extracting
a shared "body" concept from six independently-evolved whole-TU writers in the
same diff that adds calls is what D-11-03 correctly refuses to do.

**IR consumed:** `core.Function.Linear.Operations` (a flat `[]core.LinearOperation`
with `SourceID`/`TargetID`/`Kind`/`TypeID`/`CalleeID`), `core.Function.Linear.Places`,
`core.Function.Linear.Blocks` (only nonempty for the foreign/branch shapes).
`core.OpCall` is a real `core.OperationKind` (confirmed: `emitLinear`'s `switch`
at line 272 already has a `case core.OpCall:` arm, line 321) — it is *dispatched*
today, just unconditionally errored.

**C emitted today, readability:** every emitted TU opens with
`/* generated by Codename Lang; schema lang.c17/0 */` and (for linear bodies)
`/* Moves below are authority transitions; C value assignment makes no ABI or
zero-copy claim. */` — `[VERIFIED: internal/compiler/cgen/cgen.go:258-259]`.
Every value-producing operation gets an inline comment naming its operation ID
and semantic label (`"copy"`/`"authority transfer"`/`"shared borrow
representation"`/`"exclusive borrow representation"`) — this is the existing
provenance-comment convention a new `emitCall` helper should extend, e.g.
`/* call: <op.ID> -> callee <op.CalleeID> */`, and per NAT-05/D-11-10 a
generated comment naming the empty attribute set and citing D-11-09 as the
reason.

**Where `emitCall` hooks in:** the two switch sites that must call it are
`emitLinear`'s operation loop (line 270-349, the `case core.OpCall:` arm at
321) and its block-bodied sibling used by `emitLinearForeign`
(`emitBranchOperations`, referenced in the guard inventory at
`cgen.go:2084`-adjacent dispatch — the roadmap's own phrasing "block bodies" for
D-11-04's second call site). D-11-04 is explicit that these are the *only* two
call sites `emitCall` may be invoked from.

**Attribute machinery already built and reusable:** `BannedOptimizerAttributes`
(`[VERIFIED: internal/compiler/cgen/cgen.go:2067-2069]`
`var BannedOptimizerAttributes = []string{"restrict", "noalias", "nothrow",
"__attribute__((malloc))", "nonnull", "returns_nonnull"}`),
`NoreturnExemption` (line 2077, `const NoreturnExemption = "_Noreturn"`),
`ScanForBannedAttributes` (line 2084-2094, substring-scans emitted C AND the
sidecar manifest), `JustifiableAttributes` (line 2109, `[]string{"restrict"}`),
`ScanForUnjustifiedAttributes` (line 2115-2127, refuses any non-justifiable
attribute or any justifiable one with an empty `JustifiedBy`). The `restrict`
emission site itself is a single `Fprintf` call:
`[VERIFIED: internal/compiler/cgen/cgen.go:498]`
`fmt.Fprintf(&out, "static %s %s(%s *restrict %s) { %s\n", typeName,
functionName, typeName, parameterName, borrowByPointerMarker)` — exactly the
single call site D-11-16 says gains one interpolated qualifier for
token-local suppression.

**`EmittedAttribute` struct (the discharge-pair design's starting point):**
`[VERIFIED: internal/compiler/cgen/cgen.go:1838-1843]`
```go
type EmittedAttribute struct {
	Attr        string `json:"attr"`
	CoreNode    string `json:"core_node"`
	Parameter   string `json:"parameter"`
	JustifiedBy string `json:"justified_by"`
}
```
This is the callee-side `justified_by` half only — D-11-11's `discharged_by`
list (one entry per `OpCall` naming operation/argument place/caller loan) does
not exist in this struct today, confirming D-11-11's "designed, not built"
framing.

## Standard Approach: `session` Runnability

`[VERIFIED: internal/compiler/session/session.go]`, targeted reads at lines
480-900, plus grep across the whole file.

**The three run sites (D-11-05's target):**
- `RunInterpreter` (`[VERIFIED: internal/compiler/session/session.go:759]`
  `if len(checked.Program.Functions) != 1 {` then at line 773
  `interp.Run(checked.Program, checked.Program.Functions[0].Name, input)`).
- `interpreterInputs` (`[VERIFIED: internal/compiler/session/session.go:783-786]`
  `if len(program.Functions) != 1 { return nil, false } / function :=
  program.Functions[0]`).
- `RunNative` (`[VERIFIED: internal/compiler/session/session.go:872,886]`
  same shape, `checked.Program.Functions[0].Name` passed to `interp.Run` before
  `cgen.EmitNative` is even called).

**Minimum change to make N>1 runnable from the CLI:** replace
`Functions[0].Name` with `callgraph.EntryFunction(program).Name` (D-11-05) at
all three sites, and widen the `!= 1` guards to permit N>1 while still
requiring `interp.Run` to accept a program with more than one function (already
true today — `interp` handles `OpCall` per SEM-08/SEM-09, complete in Phase 10)
and `cgen.EmitNative` to accept it (this phase's core deliverable). This is a
small, mechanical change relative to the second half below.

**Minimum change to make N>1 *provable* (the harder half):** the 26 `session`
guards are dominated by the verification-lane family
(`verifyOwnedCorpus`/`verifyBorrowedCorpus`/`verifyForeignCorpus`,
`VerifyPhase5ControlsAndWork`, `phase5RunInterpreterO0O3LTOLane`,
`verifyPhase6NativeDifferentialLane`, `VerifyAliasFalseNoAlias`). Each of these
functions assumes a single interpreted execution per fixture and a single
compiled binary per fixture, then feeds both into the five-axis comparator
(`Phase5CompareEngines`/`Phase5CompareDiagnosticIDs`,
`session_phase5_compare.go:71,145`). Widening these requires: (1) driving
`interp.Run` at the resolved entry function rather than `Functions[0]`
(mechanical, same fix as above); (2) deciding what "the interpreter input" is
for a multi-function program — `interpreterInputs` today only knows how to
synthesize an input for a single `Byte`/`Buffer` parameter or a single
data-type match; a multi-function corpus fixture's entry function has the same
shape constraint, so this is not new design work, just a scope widening; (3)
the comparator itself is engine-pair-agnostic (it compares two
`execution.Execution` documents structurally) so it needs no change — only its
callers do.

**Five-axis comparator, confirmed:**
`[VERIFIED: internal/compiler/session/session_phase5_compare.go:19-25]`
```go
const (
	AxisTerminalOutcome  = "axis:terminal-outcome"
	AxisEventOrder       = "axis:event-order"
	AxisResourceLedger   = "axis:resource-ledger"
	AxisExitStatusSignal = "axis:exit-status-signal"
	AxisDiagnosticID     = "axis:diagnostic-id"
)
```
Lives in `internal/compiler/session/session_phase5_compare.go`, driven by
`Phase5CompareEngines` (line 71) and `Phase5CompareDiagnosticIDs` (line 145).
**These are not the same five axes as QLT-03's shape-register taxonomy**
(edge topology / argument mode / callee body shape / returned-value origin
class / composition-depth bucket, D-11-45) — two independent "five axes"
exist in this codebase for two different purposes; a plan or verification
document that conflates them will misreport which "five axes" a given claim
is about. Name them explicitly as "the comparator's five axes" vs. "the shape
register's five axes."

**`-O0`/`-O3`/`-O3 -flto` invocation:** `native.Runner` (`internal/compiler/native/native.go`)
takes `Options.LTO bool` (`[VERIFIED: internal/compiler/native/native.go:57-66]`
"LTO gates D-05-19's `-flto` tier: when true, `-flto` is appended to BOTH the
per-TU foreign compile argument list and the final link argument list -- never
just one, since LTO is inert if either is missing"), writes exactly one source
file (`[VERIFIED: internal/compiler/native/native.go:155]`
`sourcePath := filepath.Join(directory, "program.c")`), and separately compiles
each `ForeignSources` entry as its own bounded, timed invocation
(`[VERIFIED: internal/compiler/native/native.go:163-197]`, the `-c
foreignSource -o objectPath` loop with `-flto` appended when `r.LTO`). Adding
an LTO tier to a criterion-2 corpus lane costs nothing new mechanically
(`Options.LTO = true` on an existing `Runner`) — the cost is entirely in
D-11-25's disclosed consequence: because all Lang functions land in one TU
(D-11-24), `-flto` is inert by construction for that corpus, so running it adds
a green no-op lane unless NAT-07's separate hand-written-C control exists to
give the tier something real to exploit.

## Standard Approach: `reduce`

`[VERIFIED: internal/compiler/reduce/reduce.go]`, full file structure read
(968 lines), key functions read in full.

**`Reduce`'s hard error, confirmed verbatim:**
`[VERIFIED: internal/compiler/reduce/reduce.go:116-119]`
```go
func Reduce(ctx context.Context, seed core.Program, interesting Predicate) (Result, error) {
	if len(seed.Functions) != 1 {
		return Result{}, fmt.Errorf("reduce: seed must carry exactly one function, got %d", len(seed.Functions))
	}
```

**`ProjectSource`'s unsupported-projection string, confirmed verbatim:**
`[VERIFIED: internal/compiler/reduce/reduce.go:636-639]`
```go
func ProjectSource(program core.Program) string {
	if len(program.Functions) != 1 {
		return unsupportedProjection(fmt.Sprintf("expected exactly one function, got %d", len(program.Functions)))
	}
```
where `unsupportedProjection` (line 686-688) wraps the reason as a
`// reduce: projection unsupported for this program shape -- ...` C comment —
this is a returned STRING, not a panic or error, matching the roadmap's framing
exactly ("`ProjectSource` returns an unsupported-projection string").

**Structural requirement for HDD across function boundaries:** all five
existing moves (`dropUnusedBinding`, `dropUnmatchedArm`, `dropOffpathForeignStage`,
`collapseBranchToDivergingArm`, `truncateToMinimalPrefix`) index directly into
`p.Functions[0]` (confirmed at lines 184, 214, 266, 484, 584 — five distinct
`fn := p.Functions[0]` sites). Generalizing them to per-function loops (D-11-29's
"the existing five moves become per-function loops replacing the
`p.Functions[0]` indexing") is mechanical iteration, not a redesign of any
single move's logic — each move's own candidate-selection logic is unchanged,
only the outer loop over which function it scans changes. The genuinely new
structural work is the two whole-program moves (`drop-call-site`,
`drop-orphan-function`) and their ordering: they must run BEFORE the five
per-function moves in HDD's level-by-level descent (coarse structure first),
and `drop-call-site` must run before `drop-orphan-function` becomes eligible
(a callee only becomes orphanable after its last call site is dropped) — this
is J-Reduce's dependency-graph closure expressed as two single-site moves
rather than a graph-algorithm rewrite of the reducer's core loop.

**`callgraph` import boundary preserved:** `reduce` never imports `callgraph`
per D-11-29's own verification claim ("forward references check clean (so
`ProjectSource` needs no topological sort and `reduce` never imports
`callgraph`, keeping that package's consumer set pinned to `check`")) —
confirmed no `callgraph` import exists in `reduce.go`'s current import block
(`[VERIFIED: internal/compiler/reduce/reduce.go:16-23]`, imports are
`context`, `encoding/json`, `fmt`, `strings`, and `core` only). This is a
constraint the plan must preserve, not merely a description of the status quo:
`callgraph`'s own package doc names its consumer set as `check` alone
(`[VERIFIED: internal/compiler/callgraph/callgraph.go:1-9]`
"It is consumed by check, in production, immediately before check.Program
returns a core.Program -- never by corevalidate, ast, interp, or cgen.") — a
plan that has `reduce` call `callgraph.Order` or `callgraph.EntryFunction`
directly would violate this documented, presumably-tested import boundary.

## Standard Approach: The Comparator (recap, see also `session` section above)

Already covered above under `session` — the comparator itself
(`Phase5CompareEngines`/`Phase5CompareDiagnosticIDs`) needs no structural
change for multi-function input; its five axes are engine-pair-agnostic
structural comparisons over an `execution.Execution` document. The cost is
entirely in the *callers* that must be widened to drive it on a multi-function
program, and the LTO-tier addition costs nothing mechanically (`Options.LTO`
already exists) but is inert-by-construction for the production corpus per
D-11-24/D-11-25.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Deterministic multi-function name allocation | A second name-collision scheme for cross-function names | `callgraph.Order` (already deterministic, iterative, cycle-safe) to fix global allocation order, then a fresh per-function `cNames` seeded with the reserved list plus every allocated global name (D-11-08) | C block scope already makes locals independent; the existing `cNames.allocate` collision-suffix machinery (verified at `cgen.go:2168-2179`, the `_LANG_<CATEGORY>_<ordinal>` suffix scheme with the documented double-underscore C17 §7.1.3 reservation trap) needs no new algorithm, only a new seeding call |
| Cross-function reachability analysis for the shape register | A bespoke graph-reachability engine | `testsupport.GenerateCallGraphCorpus`/`CallGraphCorpusShapes` (already exists, five shapes: `[VERIFIED: internal/compiler/testsupport/callgraphcorpus.go:47]` `func CallGraphCorpusShapes() []string` and `[VERIFIED: internal/compiler/testsupport/callgraphcorpus.go:57]` `func GenerateCallGraphCorpus(shape string, n int) (core.Program, error)`) plus `callgraph.Order`'s existing traversal | Building a second corpus generator duplicates a proven, already-relocated (D-09-46) asset with a documented limitation (zero borrow ops) that the register must disclose, not silently fix by generating a different corpus |
| Interprocedural cache invalidation | A closure-keyed cache layer | Nothing — `ArtifactSpec.FixtureSource` already whole-program-hashes | D-11-38's proof: with one module, one file, no separate compilation units, a whole-program hash strictly dominates any per-closure key; building the closure-keyed cache machinery this phase would be building a *weaker* cache and calling it stronger |
| The `EmittedAttribute` discharge-pair validator | A new independent validator package for attribute claims | `corevalidate.ValidateEmittedAttributes` (`[VERIFIED: 11-CONTEXT.md canonical_refs cites corevalidate.go:3573]`, already the proven D-05-01 independent validator) | The discharge-pair design (D-11-11) is not built this phase, but when it is, it extends this existing validator rather than adding a parallel one — consistent with the "two independent knowers, never three" discipline this repo already follows |
| A flaky/nondeterministic-miscompile tolerance mechanism in the reducer | Retry logic, majority-vote reduction, or fuzzy interestingness | Nothing — explicitly refused (D-11-36); record as a named `RefusedShapes()` entry instead | Tolerance would make the reducer non-deterministic, violating the project's byte-for-byte determinism discipline; the engineered NAT-07 control is a seeded, deterministic mutation, not a real nondeterministic miscompile, so no flaky case needs handling this phase |

**Key insight:** every "don't hand-roll" instance in this phase is not "use a
third-party library instead" (there are none) but "the machinery already
exists in this repo and generalizing it is strictly cheaper and safer than
building a parallel mechanism." This project's own established pattern
(additive sibling, D-04-20) is itself the anti-hand-roll discipline.

## Architecture Patterns

### System Architecture Diagram — call-boundary emission and equivalence proof

```
  .lang source (multi-function)
        |
        v
  check.Program  --(callgraph.Order, acyclic)-->  core.Program (N functions, OpCall edges)
        |
        v
  corevalidate.Validate  --(independent re-derivation)-->  validated core.Program
        |
        +----------------------------+------------------------------+
        v                            v                              v
  interp.Run                   cgen.EmitNative                 reduce.Reduce
  (entry = callgraph            (entry = callgraph               (Seed{Program,
   .EntryFunction)               .EntryFunction;                  EntryFunctionID};
        |                        emitProgram loops                two new whole-
        |                        callgraph.Order;                 program moves +
        |                        emitCall at exactly               five per-function
        |                        two switch sites;                 moves, in that
        |                        zero attributes                   order)
        |                        emitted, D-11-09)                     |
        |                              |                                v
        |                              v                          re-verify: strict
        |                        clang -O0 / -O3 /                field equality
        |                        -O3 -flto (native.Runner,          (Axis, EnginePair,
        |                        single program.c;                  OperationID) against
        |                        LTO inert-by-construction           the SAME property
        |                        here, D-11-25)                     the un-reduced
        |                              |                             seed exhibited
        v                              v
  session_phase5_compare.go: Phase5CompareEngines / Phase5CompareDiagnosticIDs
  (five axes: terminal-outcome, event-order, resource-ledger,
   exit-status-signal, diagnostic-id)
        |
        v
  mid-phase gate (D-11-14): cgen.ScanForBannedAttributes == empty  AND
  check reports N>=1 functions that WOULD have carried restrict   AND
  >=2 functions, >=1 call edge                                    AND
  interpreter == -O0
        |
        v
  qlt03_shape_register.json: mechanically-enumerated axis product
  (edge topology x argument mode x callee body shape x return-origin
   class x composition depth), reached/reached_thin/unreachable/gap,
   each unreachable row naming a falsifier test
```

### Recommended Project Structure

```
internal/compiler/
├── cgen/
│   ├── cgen.go                # unchanged single-function Emit/EmitNative dispatch (D-11-01)
│   └── cgen_program.go        # NEW: emitProgram, emitCall, the TU assembler (D-11-01/D-11-03/D-11-04)
├── callgraph/
│   └── callgraph.go           # gains EntryFunction(program) (D-11-05)
├── session/
│   ├── session.go             # RunInterpreter/RunNative/interpreterInputs switch to EntryFunction (D-11-05)
│   ├── session_phase5_compare.go   # comparator, unchanged structurally
│   ├── qlt03_shape_register.go     # NEW: register loader/audit, sibling of qlt01.go's skeleton (D-11-43)
│   └── qlt03_shape_register.json   # NEW: the register itself (D-11-43)
├── reduce/
│   └── reduce.go              # Seed{Program,EntryFunctionID}; two new moves; per-function loops; derived MaxReductionAttempts (D-11-29/D-11-30/D-11-31)
├── native/
│   └── native_lto_test.go     # gains the 4x3 hand-written-C matrix (Q-04, D-11-24)
└── cache/
    └── probe.go               # DeclaredInputNames() gains an eighth name closing the cgen hole (D-11-41)
```

### Pattern 1: Additive Sibling Over Signature Change

**What:** add a brand-new function/file reached only through a new dispatch
branch, never touching the existing function's signature or body.
**When to use:** whenever a widened capability risks moving already-pinned
digest/golden bytes.
**Example:**
```go
// Source: internal/compiler/cgen/cgen.go:29-37 (existing D-11-01 target site)
func Emit(program core.Program) (string, error) {
	validated := corevalidate.Validate(program)
	if !validated.Valid {
		return "", fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
	}
	program = validated.Program()
	if len(program.Functions) != 1 {
		return emitProgram(program) // NEW additive dispatch, D-11-01
	}
	function := program.Functions[0]
	// ... six existing single-function paths, byte-for-byte unchanged
}
```

### Pattern 2: Two Independent Knowers, Never One Derivation Read Twice

**What:** every trust-crossing fact this project derives is re-derived by a
second, independently-written mechanism reading a different input (never the
first mechanism's own output).
**When to use:** the mid-phase gate (D-11-14: `cgen`'s own output bytes vs.
`check`'s independent count of would-have-been-restrict functions), the
reducer's silent-slippage guard (D-11-33: static walk vs. dynamic `-O0` event
stream), and cache soundness (D-11-39b: a structural test, not a functional
one, proving nothing new is cacheable).
**Example (from the mid-phase gate design):**
```
knower 1: cgen.ScanForBannedAttributes(emittedArtifacts...) == nil   // reads OUTPUT BYTES
knower 2: check-derived count of functions whose by-pointer lowering
          WOULD have selected restrict, N >= 1                       // reads the PROGRAM
assert: knower1_empty AND knower2_N>=1 AND structural_floors AND (-O0 == interpreter)
```

### Anti-Patterns to Avoid

- **Fixture selection wearing a different hat:** hard-wiring `ByPointer = false`
  and calling the mid-phase gate satisfied — this cannot distinguish "we
  suppressed attributes" from "this program had nothing to say" (D-11-14). This
  is the exact M001 three-failure shape: a green test whose reachable input
  space omitted the hard case.
- **A free-text `unreachable` row in the shape register:** rejected as theater
  (D-11-47) — "an assertion test that the shape never appears across the
  corpus" is a tautology for a deterministic generator and goes vacuous the
  moment the generator widens while sweep sizes stay small. Every `unreachable`
  row must name one of five admissible proof mechanisms (D-11-46).
  Also rejected: a prose structural argument alone, unfalsifiable without a
  backing test.
  Also rejected: an `Interesting`-equality re-verification for QLT-05
  (D-11-32) — it would satisfy the requirement by construction (permitting the
  exact relaxation HDD search needs during search, but which reintroduces
  slippage at re-verification time), proving nothing.

## Don't Hand-Roll

(see table above under "Standard Stack" section — merged there to avoid
duplication per this document's required section ordering; retained here as a
pointer per the template.) See "## Don't Hand-Roll" above.

## Runtime State Inventory

Not applicable — this is a greenfield feature-addition phase (new
`emitProgram`, new `callgraph.EntryFunction`, new reducer moves), not a
rename/refactor/migration phase. No stored data, live service config,
OS-registered state, secrets, or build-artifact renaming is in scope.

## Common Pitfalls

### Pitfall 1: Staking Correctness on LLVM's Scoped-Noalias Machinery

**What goes wrong:** emitting a call-boundary `restrict`/noalias-shaped
attribute that is technically "justified" by a checked fact, but whose
soundness under Clang's inliner and LTO is assumed rather than proven.
**Why it happens:** `restrict` (and LLVM's `noalias` metadata, which is what
`restrict` on a parameter lowers to after inlining) is a *promise*, not a
runtime check — if the promise is wrong, or if LLVM's scoped-noalias tracking
through inlining has a bug, the optimizer can silently miscompile. This is not
hypothetical: Rust's `noalias` on `&mut` references was disabled twice due to
miscompilation bug reports (rust-lang/rust#31681, #54878), re-enabled only on
LLVM 12 (rust-lang/rust#82834), and immediately regressed again
(rust-lang/rust#84958). `[CITED: rust-lang/rust issue numbers, per 11-CONTEXT.md's
sourced research]` — this research session did not re-fetch each GitHub issue
directly but the citation chain (specific issue numbers, specific LLVM version
gate) is precise enough to be independently checkable and is consistent with
publicly documented Rust/LLVM history.
**How to avoid:** D-11-09's answer — do not emit the attribute at all this
phase, because Lang's structural guarantees (one parameter, no globals, no
callbacks, no address-escaping foreign contracts, `check` already refuses
passing one place to two calls — confirmed:
`testdata/phase07/call_argument_used_twice.lang` exists as a refusal fixture
per the canonical refs) mean two pointers to one object cannot exist across a
call boundary in M002 in the first place. The promise would be vacuously true
but would still stake the *proof* on machinery whose weakness is documented.
**Warning signs:** any plan task that proposes emitting `restrict` at a call
boundary "because the checker already proves no aliasing" without first
establishing that (a) M002's language surface cannot construct the aliasing
hazard at all (not merely "the checker catches it when it occurs") and (b) a
Clang-version-pinned falsification test exists.

### Pitfall 2: Confusing "Runnable" With "Provable"

**What goes wrong:** a plan declares NAT-04 done once `cgen.Emit`/`EmitNative`
stop refusing multi-function programs, without also widening the 26 `session`
guards that constitute the five-axis equivalence proof itself.
**Why it happens:** the roadmap's criterion 1 literally names only
`Emit`/`EmitNative`; a plan reading only the roadmap prose (not
`11-CONTEXT.md`'s guard-inventory correction) will underscope.
**How to avoid:** treat "the program compiles, links, and runs" (criterion 1,
narrow) and "the program is proven equivalent across all four engines"
(criterion 2, wide) as separately-costed deliverables from the start, per the
`<domain>` section's own framing ("Making multi-function programs runnable and
making them provable are separate costs").
**Warning signs:** a plan wave that claims NAT-06 complete without a
corresponding change to `verifyBorrowedCorpus`/`verifyOwnedCorpus`/
`verifyForeignCorpus`/`VerifyPhase5ControlsAndWork`.

### Pitfall 3: The Reducer as a Downstream Consumer Rather Than a Subject

**What goes wrong:** treating `reduce` as something that "just works once
`cgen` handles calls," because HDD reduction conceptually operates on the
`core.Program` `cgen` also consumes.
**Why it happens:** `reduce` shares the `core.Program` type with `cgen`, so it
is easy to assume shared IR implies shared readiness. It does not — `reduce.Reduce`
and `reduce.ProjectSource` both hard-refuse `len(Functions) != 1` independently
of `cgen`'s own guards, confirmed by direct read this session.
**How to avoid:** plan `reduce`'s widening (two new whole-program moves, the
`Seed{Program, EntryFunctionID}` API change, the derived `MaxReductionAttempts`
bound, the `foreignCallSequenceFor` silent-slippage fix) as first-class waves,
not as a side effect of `cgen`'s work.
**Warning signs:** a plan with zero tasks touching `internal/compiler/reduce/`.

### Pitfall 4: An Interprocedural Cache "Win" That Is Actually a Soundness Loss

**What goes wrong:** implementing QLT-06's literal prose ("cache keys derive
from the call-graph closure") as a performance optimization, without noticing
that in a single-compilation-unit language this is strictly *weaker* evidence
than the whole-program hash already in use.
**Why it happens:** "closure-keyed" sounds like a natural generalization of
per-function caching once calls exist, and QLT-06's named failure mode
("per-unit hashes") sounds like it describes a gap that must be filled.
**How to avoid:** verify, as this research did, that `internal/compiler/cache/`
contains zero references to `ClosureDigest` and that `FixtureSource` already
hashes the whole file — then recognize that QLT-06's failure mode does not
exist in this repo because there are no per-unit hashes to begin with (D-11-38).
Ship QLT-06b as complete-by-abstention with a structural test, not a new cache
layer.
**Warning signs:** any plan task that adds an import of `core` or
`originvalidate` to `internal/compiler/cache/`.

### Pitfall 5: Fixing the Wrong Half of the Live Cache Hole

**What goes wrong:** assuming the cache-soundness fix belongs in `session`
(where `cache.Consult` is called, `session_phase6_verify.go:307`) rather than
in `cache` itself (where `DeclaredInputNames()` lives).
**Why it happens:** the *symptom* (a stale binary served after a `cgen` edit)
is only observable at the `session` call site.
**How to avoid:** confirmed this session — `cache.DeclaredInputNames()`
(`internal/compiler/cache/probe.go:38-48`) is the sole source of truth for
what counts as a cache key input; it currently returns exactly
`["fixture_source", "build_flags", "clang_identity", "runtime_identity",
"foreign_translation_unit", "mutation_runner_source", "go_toolchain"]` with no
entry naming `cgen`'s own source. The fix is adding an eighth declared input
(hashing `internal/compiler/cgen/*.go`'s own source, or a build-identity
digest that changes whenever `cgen` is rebuilt) in `cache`/`probe.go`, and
having every `ArtifactSpec` construction site supply it — a `session`-side fix
alone (e.g., special-casing a cache-bypass) would not close the hole for any
other future caller of `cache.Consult`.

### Pitfall 6: Believing the Corpus's LTO Tier Proves Anything About Lang Code

**What goes wrong:** running criterion 2's corpus at `-O3 -flto` and reporting
"the LTO tier passed" as evidence the tier is meaningful for interprocedural
Lang code.
**Why it happens:** the corpus DOES pass at `-flto` — but only because
`-flto` is inert by construction (D-11-24/D-11-25: all Lang functions land in
one translation unit this phase, so there is nothing for LTO to do across a
Lang-to-Lang call). A green LTO lane over the production corpus is not evidence
the tier was exercised.
**How to avoid:** D-11-25 requires stating this explicitly in
`11-VERIFICATION.md`: "the existing LTO lane's non-inertness is borrowed
entirely from the *foreign* TU boundary" and NAT-07's hand-written-C control
(not the corpus) is what proves the tier CAN be exploited.
**Warning signs:** a verification document that reports NAT-06's LTO axis
green without a corresponding disclosure that the corpus itself cannot
distinguish "LTO worked" from "LTO had nothing to do."

## Code Examples

### The existing `core.OpCall` forward-hygiene error (the exact site `emitCall` replaces)

```go
// Source: internal/compiler/cgen/cgen.go:339-345, verified this session
case core.OpCall:
	// D-07-39/A-02: OpCall is registered but not lowered by native
	// emission this phase. Emit/EmitNative hard-fail on
	// len(program.Functions) != 1 (cgen.go:22,52) before this arm
	// could ever run -- no legal OpCall-bearing program has exactly
	// one function -- so this is forward hygiene for Phase 11's
	// multi-function C emission, not a live gap.
	return "", fmt.Errorf("operation %q: Lang-to-Lang calls are not supported by native emission this phase", operation.ID)
```

### The `restrict` emission site D-11-16's suppression control must gain one qualifier on

```go
// Source: internal/compiler/cgen/cgen.go:498, verified this session
fmt.Fprintf(&out, "static %s %s(%s *restrict %s) { %s\n", typeName,
	functionName, typeName, parameterName, borrowByPointerMarker)
```

### The seven declared cache inputs (D-11-41's fix target)

```go
// Source: internal/compiler/cache/probe.go:38-48, verified this session
func DeclaredInputNames() []string {
	return []string{
		"fixture_source",
		"build_flags",
		"clang_identity",
		"runtime_identity",
		"foreign_translation_unit",
		"mutation_runner_source",
		"go_toolchain",
	}
}
```
No entry names `cgen`'s own source — the live hole D-11-41 discloses.

### `peerDeriveOriginFacts`'s missing `core.OpCall` case (the Critical Path)

```go
// Source: internal/compiler/corevalidate/corevalidate.go:2407-2450, verified this session
func peerDeriveOriginFacts(function *core.Function) peerOriginFact {
	if function.Linear == nil {
		return peerOriginFact{}
	}
	paramTrace := map[string]bool{function.Parameter.ID: true}
	derived := make(map[string]string)
	for _, operation := range function.Linear.Operations {
		switch operation.Kind {
		case core.OpBorrowShared:
			// ...
		case core.OpBorrowExclusive:
			// ...
		case core.OpMove, core.OpCopy:
			// ...
		}
		// NOTE: no `case core.OpCall:` arm exists in this switch.
	}
	for _, operation := range function.Linear.Operations {
		if operation.Kind != core.OpReturn {
			continue
		}
		if mode := derived[operation.SourceID]; mode != "" {
			return peerOriginFact{Derived: true, Access: mode}
		}
	}
	return peerOriginFact{}
}
```
A function whose `PublicOrigin` is derived from forwarding a callee's `OpCall`
result therefore has `derived[operation.SourceID] == ""` at its `OpReturn`, so
`peerDeriveOriginFacts` returns the zero `peerOriginFact{}` — `Derived: false`
— which (per the surrounding `peerCallable` logic this research did not fully
trace line-by-line, but which `11-CONTEXT.md`'s independently-verified finding
confirms) causes `corevalidate` to refuse the function as not-`Callable`
independently of `check`, which admits the same shape with zero diagnostics.
This is the exact fail-closed (not unsound) gap the Critical Path section
below discusses.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|---------------|--------|
| Single-function-only `cgen`/`reduce`/`session` (Phases 1-10) | Multi-function-aware, with entry resolved via `callgraph.EntryFunction` rather than `Functions[0]` | Phase 11 (this phase) | Every hard-coded `Functions[0]` access across 32+2 sites becomes a named, auditable design choice rather than an implicit single-function assumption |
| `restrict` emitted whenever `selectsByPointerLowering` selects a function (D-05-01/D-05-02, Phase 5) | Zero call-boundary `restrict`/noalias-shaped attributes (D-11-09) | Phase 11 (this phase) — the existing intraprocedural by-pointer `restrict` at line 498 is UNCHANGED; this state-of-the-art shift is specifically about the NEW call-boundary case, which never existed before this phase | NAT-05 becomes an honest-empty-set claim rather than a populated-promise claim |
| `MaxReductionAttempts = 64` as a flat constant (Phase 5) | A derived bound, `AttemptsPerFunction*len(Functions) + callSiteCount` (D-11-31) | Phase 11 (this phase) | Preserves byte-identical output for every existing single-function, zero-call fixture (N=1, callSiteCount=0 → exactly 64, matching today) while scaling for real multi-function seeds |

**Deprecated/outdated:** none — this phase adds capability, it does not
deprecate any existing single-function code path (all six emitters, `reduce`'s
five moves, and every existing guard's *shape* stay exactly as they are; only
their *reach* widens additively).

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | The specific Rust `noalias` GitHub issue numbers (#31681, #54878, #82834, #84958) cited in D-11-09/D-11-10 accurately describe the disable/re-enable/regression history | Common Pitfall 1, Alternatives Considered | Low — even if one issue number is imprecise, the qualitative claim (LLVM scoped-noalias has a documented history of miscompilation-driven feature flips) is well-established Rust/LLVM community knowledge and does not change D-11-09's structural argument (which rests on Lang's own language-surface guarantees, not on this citation alone) |
| A2 | `-flto` on this research session's host (Apple clang 21.0.0, arm64-apple-darwin25.6.0) genuinely produces `interpreter == -O0 == -O3 != (-O3 -flto)` for the composition-only shape described in D-11-23 | NAT-07 section, User Constraints (D-11-20/D-11-22) | Medium — this is a claim `11-CONTEXT.md` reports as "measured on this host" by a prior researcher; this research session did not independently re-run the 4×3 matrix (Q-04 is explicitly a pre-planning experiment the plan must still execute). If the plan's own execution host has a different Clang version, the divergence signature could differ, which is exactly why D-11-22 requires a toolchain-pinned re-measurement clause |
| A3 | `26` is the correct count of `session`-package guard-pattern matches when counted the way `11-CONTEXT.md`'s prose implies ("session 26 ... plus verifyBorrowedCorpus x9") | Guard Inventory section | Low — the awk-based function-count re-verification this session performed (15 distinct `func` headers in `session.go` matching the pattern, plus additional matches in `session_phase5.go`, `session_phase5_alias.go`, `session_phase5_corpus.go`, `session_phase5_mismatch.go` (4), `session_phase6.go`, `session_phase6_verify.go`, `session_phase7.go`) sums to 26 non-test occurrences across the `session` package's files, matching the roadmap's total exactly; the discrepancy is only in how the "×9" sub-annotation maps onto function-count vs. body-occurrence counting, not in the total |
| A4 | `check.buildInterproceduralSummariesObserved`'s cited line range (639-647) and the `interproceduralConsultObserved` seam are load-bearing for QLT-03's op-kind closure test design | Standard Approach sections (brief mention only) | Low — this research session confirmed the function's doc comment exists at that location but did not fully trace its consumer set; the planner should re-confirm before relying on it as a fault-injection seam site |

**If this table is empty:** N/A — populated above.

## Open Questions

1. **Does Q-01 pass or fail?** (blocks D-11-29's central reducer move)
   - What we know: `peerDeriveOriginFacts` genuinely has no `core.OpCall` case
     (confirmed this session by direct read). The Critical Path section of
     `11-CONTEXT.md` names four independent routes this bites.
   - What's unclear: whether `corevalidate.Validate`/`check` accepts or
     rejects the hand-applied `OpCall`→`OpCopy` rewrite at the CORE level on
     `testdata/phase07/deep_diamond_acyclic.lang`. This is exactly what Q-01
     (a Go-test-only experiment, <1 hour) settles, and it was NOT run during
     this research session (research produces the question and the
     verification of its preconditions, not the experiment's own result —
     that is explicitly a pre-planning experiment for the plan/execution
     phase, per `11-CONTEXT.md`'s own framing: "Q-01 and Q-02 must run before
     any plan is written").
   - Recommendation: run Q-01 literally as specified in `11-CONTEXT.md` before
     finalizing the reducer plan waves; plan for both branches (QLT-05 as
     reducer work, or QLT-05 narrowed behind the `corevalidate` `OpCall`
     closure with `RefusedShapes()` naming the gap) as `11-CONTEXT.md` already
     instructs.

2. **Does Q-02 confirm the live cache hole?**
   - What we know: `cache.DeclaredInputNames()` has no `cgen` entry (confirmed
     this session, exact source quoted above).
   - What's unclear: whether `cache.Consult` in practice actually serves a
     stale binary end-to-end when `cgen`'s emitted C changes for an unchanged
     `.lang` fixture — this requires running the actual experiment
     (`Phase6CacheRootOverrideForTest` + `verifyPhase6NativeDifferentialLane`),
     not just reading the declared-inputs list.
   - Recommendation: run Q-02 before any cache-touching plan task; the fix
     (add an eighth declared input) is straightforward once confirmed.

3. **What does the actual Clang toolchain on the execution host report, and
   does the D-11-23 4×3 matrix reproduce the claimed divergence there?**
   - What we know: this research session's host is Apple clang 21.0.0,
     arm64-apple-darwin25.6.0 (per `11-CONTEXT.md`'s own disclosed
     measurement; this research session did not independently re-run `clang
     --version` or the matrix itself — that is Q-04's job).
   - What's unclear: whether the plan's actual execution host (which may or
     may not be the same machine) reproduces the same non-monotonic
     inlining-aggressiveness behavior.
   - Recommendation: Q-04 must run and its result recorded verbatim in the
     phase's verification artifact, with the exact `clang --version` string
     captured alongside the matrix result, per D-11-22's toolchain-pinning
     requirement.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | All packages under `internal/compiler/` | ✓ | Go 1.24 (per STATE.md's standing commitment) | — |
| Clang/LLVM | `native.Runner`, NAT-06/NAT-07's four-engine equivalence proof | ✓ (implied by existing Phase 5/6 native lanes already running in this repo) | Apple clang 21.0.0 (this research session's host; MUST be re-captured and pinned on the actual execution host per D-11-22) | None — Clang is a hard, already-adopted dependency for every native-tier phase since M001; no fallback exists or is needed |
| ASan/UBSan | Referenced in `LANGUAGE-MATURITY.md`'s "assurance" inventory, not newly required by Phase 11's stated success criteria | Not independently verified this session | — | Not in this phase's critical path (NAT-04–NAT-07/QLT-03/05/06 do not name a sanitizer requirement) |

**Missing dependencies with no fallback:** none identified — this phase is
entirely internal Go-package extension plus the already-adopted Clang
toolchain.

**Missing dependencies with fallback:** none.

## Validation Architecture

`.planning/config.json` has `workflow.nyquist_validation: true` (confirmed
this session), so this section is required.

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go's built-in `testing` package (`go test`) — this repo has no external test framework, consistent with the zero-external-dependency record |
| Config file | none — no `go.mod` test-runner config beyond the module's own `go.mod` |
| Quick run command | `go test ./internal/compiler/cgen/... ./internal/compiler/reduce/... ./internal/compiler/session/... -run TestPhase11 -v` (once Phase 11 test names exist; until then, package-scoped `go test ./internal/compiler/<package>/...`) |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|--------------------|-------------|
| NAT-04 | `cgen.Emit`/`EmitNative` accept N>1 functions and emit valid C17 | unit + integration | `go test ./internal/compiler/cgen/... -run TestEmitProgram -v` | ❌ Wave 0 — `cgen_program_test.go` does not exist yet |
| NAT-05 | Emitted artifact carries a generated comment naming the empty attribute set and citing D-11-09 | unit | `go test ./internal/compiler/cgen/... -run TestEmittedAttributeSetIsExplicitlyEmpty -v` | ❌ Wave 0 |
| NAT-06 | Interpreter/`-O0`/`-O3`/`-flto` agree on the interprocedural corpus (five-axis comparator) | integration/differential | `go test ./internal/compiler/session/... -run TestPhase11InterproceduralDifferential -v` | ❌ Wave 0 |
| NAT-07 | 4×3 mutation matrix, one red cell, mutation-killed | integration (hand-written C, no `cgen` involvement) | `go test ./internal/compiler/native/... -run TestCompositionOnlyLTODivergence -v` | ❌ Wave 0 — but see Q-04, which is this same test built as a pre-planning experiment |
| QLT-03 | `qlt03_shape_register.json` audit, zero blank/unclassified cells | unit (register load + mechanical recomputation) | `go test ./internal/compiler/session/... -run TestQLT03Register -v` | ❌ Wave 0 |
| QLT-05 | `reduce` re-verification on a multi-function seed, strict field equality | unit + differential | `go test ./internal/compiler/reduce/... -run TestReduceMultiFunctionSeed -v` and `go test ./internal/compiler/session/... -run TestQLT05Reverification -v` | ❌ Wave 0 |
| QLT-06 | Structural test: `cache.DeclaredInputNames()` unchanged at 7 (or explicitly widened to 8 for the `cgen` fix, D-11-41), `cache` does not import `core`/`originvalidate` | unit (structural) | `go test ./internal/compiler/cache/... -run TestDeclaredInputNamesStableAndNoInterproceduralImport -v` | ❌ Wave 0 |

### Sampling Rate

- **Per task commit:** package-scoped `go test ./internal/compiler/<touched-package>/...`
- **Per wave merge:** `go test ./...`
- **Phase gate:** full suite green before `/gsd-verify-work`, AND the
  mid-phase gate's own four-part conjunction (D-11-14) green before any
  attribute-emission work is even attempted (there is none planned this
  phase, per D-11-09, but the gate's structural floor — ≥2 functions, ≥1 call
  edge, N≥1 would-have-carried-restrict — must still be demonstrated non-vacuous).

### Wave 0 Gaps

- [ ] `internal/compiler/cgen/cgen_program_test.go` — covers NAT-04, NAT-05
- [ ] `internal/compiler/callgraph/callgraph_entry_test.go` (or extend
      `callgraph_test.go`) — covers `EntryFunction`'s zero-or-many-roots
      fail-closed refusal (D-11-05)
- [ ] `internal/compiler/reduce/reduce_multifunction_test.go` — covers QLT-05,
      the two new moves, `RefusedShapes()`
- [ ] `internal/compiler/session/qlt03_shape_register_test.go` — covers QLT-03
- [ ] `internal/compiler/session/session_phase11_gate_test.go` — covers the
      mid-phase gate's four-part conjunction (D-11-14/D-11-18/D-11-19)
- [ ] `internal/compiler/native/native_lto_test.go` extension — covers NAT-07's
      4×3 matrix (Q-04 builds this before any `cgen` work per D-11-24)
- [ ] `internal/compiler/cache/probe_test.go` extension — covers the
      `DeclaredInputNames()` eighth-entry fix (D-11-41)

## Security Domain

`.planning/config.json` has `workflow.security_enforcement: true` (confirmed
this session), so this section is required. Note: Phase 10 ran with
`security_enforcement=true` and produced no `10-SECURITY.md` (`/gsd-secure-phase 10`
was never run, per STATE.md) — Phase 11 should not repeat this gap.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-------------------|
| V2 Authentication | No | Not applicable — compiler toolchain, no auth surface |
| V3 Session Management | No | Not applicable |
| V4 Access Control | No | Not applicable |
| V5 Input Validation | Yes | `.lang` source is bounded (`syntax.MaxSourceBytes`, already enforced); `corevalidate.Validate` is the independent re-derivation gate every emission path already consults before touching a `core.Program`; the new `emitCall`/`emitProgram` path must consult it identically, never bypass it for a "fast path" |
| V6 Cryptography | No | Not applicable — no cryptographic material in this phase |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|----------------------|
| Generated-C identifier collision opening a namespace-confusion injection (e.g., a crafted function/place name colliding with a `LANG_`-prefixed reserved identifier) | Tampering | The existing `cNames` allocator's prefix-confinement + honest-reservation properties (`cgen.go:92-150`), extended per D-11-08's two-tier seeding — this is an EXISTING mitigation this phase must preserve, not a new one to build |
| A foreign contract field string splicing a C comment terminator into generated C (already mitigated for single-function foreign contracts) | Tampering / Injection | `unsafeForeignContractField`/`validForeignSymbol` (`cgen.go:1876-1891`) — D-11-07 explicitly keeps `singleForeignFunction`/`singleManifestFunction` scoped to at-most-one foreign contract rather than generalizing the splice surface to N contracts, which is the conservative, safety-preserving choice |
| A cache key that fails to bind `cgen`'s own source, allowing a stale (potentially miscompiled or backdoored-during-dev) binary to be silently reused | Tampering (of the build artifact identity, not of user input) | D-11-41's fix: add `cgen`'s source as an eighth declared cache input |

## Sources

### Primary (HIGH confidence — verified by direct file read this session)
- `internal/compiler/cgen/cgen.go` (2,217 lines, multiple targeted reads
  covering lines 1-352, 440-900, 1800-1980, 2050-2180)
- `internal/compiler/session/session.go` (targeted reads, lines 480-900, plus
  full-file grep for guard patterns)
- `internal/compiler/session/session_phase5_compare.go` (lines 1-40)
- `internal/compiler/reduce/reduce.go` (targeted reads covering lines 1-170,
  160-290, 636-696, plus full-file grep)
- `internal/compiler/callgraph/callgraph.go` (lines 1-60, 205-320)
- `internal/compiler/corevalidate/corevalidate.go` (targeted reads, lines
  1660-1800, 1920-2060, 2395-2465, plus full-file grep for `case core.Op`)
- `internal/compiler/cache/probe.go` (lines 1-70), `internal/compiler/cache/cache.go`
  (grep for `ComputeKey`/`DeclaredInputNames`)
- `internal/compiler/native/native.go` (lines 1-30, 50-70, 150-200)
- `internal/compiler/native/native_lto_test.go` (lines ~40-100)
- `internal/compiler/session/qlt01.go` (lines 1-40)
- `internal/compiler/check/check.go` (lines 635-650)
- `internal/compiler/testsupport/callgraphcorpus.go` (grep for `CallGraphCorpusShapes`/`GenerateCallGraphCorpus`)
- `.planning/config.json` (workflow flags)
- `.planning/phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-CONTEXT.md`
  (843 lines, read in full)
- `.planning/STATE.md`, `.planning/REQUIREMENTS.md`, `.planning/LANGUAGE-MATURITY.md`
  (read in full)

### Secondary (MEDIUM confidence — cited from `11-CONTEXT.md`'s own prior research, not independently re-fetched this session)
- Rust `noalias` history (rust-lang/rust#31681, #54878, #82834, #84958)
- "Restrict-Qualified Pointers in LLVM" (Finkel, 2017); "ptr_provenance and
  llvm.noalias: The Tale of Full Restrict" (2021)
- LLVM `OptBisect`/`-opt-bisect-limit`
- Csmith (PLDI'11); Swarm Testing (ISSTA'12)
- "Build Systems à la Carte" (early cutoff)
- J-Reduce / "Binary Reduction of Dependency Graphs" (ESEC/FSE'19)
- MacIver, "Notes on Test-Case Reduction"

### Tertiary (LOW confidence)
- None — every claim in this document is either a direct source-file
  verification performed this session, or an explicit citation of
  `11-CONTEXT.md`'s own already-adjudicated research (which itself ran a
  seven-way adversarial fan-out and is treated here as an authoritative
  upstream input, not re-litigated).

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — zero new dependencies; every mechanism is an
  in-repo extension, all cited sites re-read this session.
- Architecture: HIGH — the additive-sibling, two-independent-knowers, and
  entry-resolver patterns are all pre-existing, proven conventions in this
  codebase; this phase's design decisions consistently reuse them.
- Pitfalls: HIGH for the code-verifiable ones (guard scope, cache hole,
  reducer scope, LTO-inertness); MEDIUM for the external LLVM/Clang
  `restrict`/LTO semantics claims (cited, not independently compiled this
  session — Q-04/Q-07 are the falsification steps still owed).

**Research date:** 2026-09-11
**Valid until:** this phase's completion — the guard counts, line numbers, and
cache-input list are all point-in-time facts about a codebase under active
development in this same milestone; re-verify with the awk command and the
grep commands shown throughout this document if any Phase 11 plan is revisited
after a significant gap.
