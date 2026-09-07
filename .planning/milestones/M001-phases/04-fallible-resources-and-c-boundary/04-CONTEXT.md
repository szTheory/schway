# Phase 4: Fallible Resources and C Boundary - Context

**Gathered:** 2026-09-04
**Status:** Ready for planning
**Source:** advisor-mode discussion at Phase 03 close-out; four parallel research
fan-outs (call surface, failure representation, C boundary, panic/nonlocal exit),
each cross-checked against the wiki corpus, `.planning/spikes/005-*`, and the
Phase 2/3 failure history.

<domain>
## Phase Boundary

A noncopyable resource crosses one audited C boundary while partial
initialization, failure propagation, and cleanup remain defined. Requirements
SEM-03, RES-01, FFI-01.

This phase extends the same vertical source-to-native spine Phases 1–3 built. It
introduces exactly one new call form (**foreign only**), one new failure control
shape (**an edge, not a value**), one new resource lifecycle, and one terminal
defect outcome. It does **not** open new subsystems: no sanitizer lanes, no
callback registration, no dynamic-library handles, no `errno` translation, no
symbol versioning, no header ingestion, no optimizer-attribute emission.

**The constraint that shapes the whole phase:** the language today has no call
construct of any kind, and `core.DataType.Alternatives` is `[]string` — every
ADT alternative is nullary. Every Phase 4 success criterion needs at least one of
those two things built first. The decisions below choose the call form and
deliberately avoid needing the payload-carrying alternatives.

</domain>

<decisions>
## Implementation Decisions

### Call surface

- **D-04-01:** Phase 4 adds **`OpForeignCall` only**. No Lang-to-Lang calls.
  A `foreign C { }` declaration block is the only callable surface; acquisition
  steps are direct calls to declared foreign C symbols inside a linear body.
  — **Reversibility:** costly — adding `OpCall` later is additive to the IR, but
  every admission-layer differential written this phase is single-function by
  construction and must be rebuilt cross-function when it lands.

- **D-04-02:** `OpForeignCall.Callee` resolving to a Lang `Function.ID` is a hard
  refusal (`core.call_target_not_foreign`) in **both** `check` and `corevalidate`,
  independently derived. This is a fail-closed control, not a convention.

- **D-04-03:** The lift condition for Lang-to-Lang calls is recorded as
  **callable ⊆ publishable** — a function is callable only if
  `originvalidate.ValidatePublished` would publish it. This reuses the
  already-proven 03-09/03-10 machinery instead of restoring the `check.go` gate
  03-06 deliberately removed, so `exclusive_borrow_clean` keeps checking clean
  while becoming uncallable.

  **Why D-04-01 rather than adding `OpCall` now.** `.planning/phases/03-*/03-DEBT.md`
  accepted D-03-02 as non-blocking for exactly one reason: *"nothing in the
  executable semantics ever invokes one Lang function from another … the gap is
  … currently unreachable as an executable unsoundness, because there is no
  consumer."* Adding `OpCall` creates the consumer. The research produced the
  decisive witness — this program `lang check`s clean and returns a dangling
  alias, where `relay` is byte-for-byte the `exclusive_borrow_clean` fixture
  03-09 kept accepting on purpose:

  ```
  fn relay(buffer: Buffer) -> Buffer {
    let view = borrow mut buffer
    view
  }
  fn escort(buffer: Buffer) -> Buffer {
    let aliased = relay(borrow mut buffer)
    let delivered = take buffer
    aliased
  }
  ```

  No analogous program exists under D-04-01: a foreign function has no body, so
  no derivation can disagree with its declaration, and the entire D-03-02 defect
  class is structurally absent. The residual risk migrates to "does the C
  implementation honour its declared contract", which the project has already
  adjudicated as a quarantined, non-provable claim
  (`wiki/native-and-low-level-profile.md`: *"a generated binding can create a
  candidate contract, never proof that the C implementation obeys it"*). That is
  a known, named, fenced residual — categorically different from a newly created
  unfenced one.

### Failure representation

- **D-04-04:** Typed failure is a **two-successor control-flow edge in core**, not
  a data value. No `Result` type, no generics, no payload-carrying alternatives.
  `core.DataType.Alternatives` is **not touched this phase**.
  — **Reversibility:** one-way — a storable `Result` value bolted onto an
  edge-based core is a genuine re-lowering, not an extension. When error
  *translation between layers* is needed (M002, or Phase 6 error taxonomies),
  the payload-alternative debt comes due at full price.

- **D-04-05:** The error payload is an ordinary **existing nullary ADT** riding
  the `err` edge (e.g. `data AcquireError = | FirstRefused | SecondRefused`,
  which the current parser, checker, and `cgen` already handle). The ok payload
  is an ordinary place on the ok successor block. Both payloads ride the edges,
  so no alternative ever carries a payload.

- **D-04-06:** `try` and `discard … because "<rationale>"` are the **only two
  admissible consumers** of a fallible operation. `acquire` is not an expression;
  it is reachable only as the operand of one of those two. `let x = acquire …`
  does not parse. There is no severity dial: the core IR has no encoding for a
  fallible operation without a failure successor, so the ignored-result rule is
  an IR well-formedness invariant, not a lint.

  **Why not a must-use lint.** Rust's `#[must_use]` (RFC 1940) was deliberately a
  lint; it is suppressible and has documented coverage holes (unused `==` escaped
  it because HIR sees binops as distinct from calls). Go pushed the equivalent
  fully out of the compiler into `errcheck`. Zig went the other way — `!T` is
  control flow and `errdefer` exists precisely because partial-acquisition
  rollback is the hard case, which is RES-01 verbatim.

- **D-04-07:** The reverse-order release sequence is **materialized as explicit
  `OpRelease` operations into each failure block by `check`**, so `interp` and
  `cgen` consume one order rather than each deriving one. `corevalidate`
  independently *rederives* the expected order from the block/edge graph and
  compares — never reads what `check` wrote.

- **D-04-08:** Terminal outcome is a **closed named set carried as its own axis**,
  separate from the returned value: `value | typed_failure | defect`, with
  `cancelled` **reserved but unconstructible** and a fail-closed control asserting
  no engine emits it. This is the cancellation seam: async is milestone
  Out-of-Scope and cancellation is vacuously satisfied this phase, but adding it
  later must be an enum member plus a reviewed change, never a quiet retrofit as
  an `AcquireError` variant.

- **D-04-09:** There is **no error-value constructor anywhere in the IR** — no
  `OpMakeErr`, no `Err(...)` expression. The only producer of `typed_failure` is
  an `OpFail` terminator, and the only producer of `OpFail` is an `err` edge
  carrying a place of the function's declared failure type. This is what
  structurally enforces SEM-03's *"panic and cancellation cannot be erased as
  ordinary errors"*: a panic path has no syntactic or IR-level route into
  `typed_failure`.

### The C boundary

- **D-04-10:** The boundary is a **hand-written, byte-frozen, separately-compiled
  foreign translation unit that wraps real libc `malloc`/`free`** and returns a
  record by value. The compiler forms its contract from Lang source and **never
  parses the foreign TU's header**.
  — **Reversibility:** costly — the fixture is a frozen golden in the same class
  as `owned_transfer.golden.c`; moving it is a visible reviewable golden move.

  **The reframe that settles shim-vs-libc.** Quarantine is a property of the
  **information flow, not the authorship**. A shim whose private header the
  generated C `#include`s destroys quarantine even though the shim is foreign;
  a frozen TU the compiler has never parsed is quarantined even though it lives
  in-repo. Pure libc was rejected on a decisive defect: `malloc` has no record
  layout, so SC2's "target layout inspectable" has no subject at all, and six of
  Phase 5's seven hostile mutations become non-injectable.

- **D-04-11:** A **generated conformance TU** (`lang_foreign_conformance.c`) is the
  single, explicit, auditable place where Lang's declaration and the foreign TU's
  own private header are permitted to meet. It defines no symbols, is compiled
  but not linked, and carries `_Static_assert(sizeof/_Alignof/offsetof)` pairs
  across the two declarations. Under the existing `-Werror`, a field-order
  mutation becomes a **compile-time refusal**, converting spike 005 iteration 2's
  silent runtime corruption into a hard failure.

- **D-04-12:** Obligations are inspectable in **three layers derived from one
  authoritative `core.ForeignContract`**: (a) a generated `_LANG_`-namespaced
  header with the extern declaration plus a generated obligation comment block
  plus self-layout `_Static_assert`s; (b) the conformance TU; (c) a
  `lang.foreign/0` sidecar manifest digest-bound into `evidence.Manifest` via a
  new `foreign_digest` **omitempty** field. The JSON is authoritative and the C
  comments are generated *from* it, so the two can never drift.

- **D-04-13:** **cgen emits zero optimizer-visible attributes this phase** — no
  `restrict`, `noalias`, `nothrow`, `__attribute__((malloc))`, `nonnull`, or
  `returns_nonnull` — and the refusal is itself a fail-closed control
  (`control:foreign.no_unproven_attributes`), scanning all emitted C and the
  manifest's `emitted_attributes` and asserting both empty. Mutation-kill:
  injecting `restrict` into the emitter must make the control fail.
  `emitted_attributes: []` is deliberately a *field*, not an omission, so Phase 5
  can assert on it.
  — **Reversibility:** reversible — Phase 5 adds attributes only alongside the
  proven fact each derives from.

  **Adjudication of the one research conflict.** Two of the four agents proposed
  emitting `nothrow` derived from an admitted `unwind: forbidden` declaration.
  Overruled. Phase 4 has no alias analysis, no capture analysis, and no proof
  obligations to derive attributes from; `nothrow` is an LLVM-visible promise
  about a callee that cannot be inspected. Spike 005 iteration 4 showed a false
  `restrict` returning `2` at `-O0` and `1` at `-O3` on this exact host — these
  are promises the optimizer acts on, so an unproven one produces wrong code, not
  slow code. The strongest honest Phase 4 statement about the optimizer is a
  proven absence. The declared `unwind` fact stays load-bearing at the
  *admission* layer (D-04-15), which is where it can actually be checked.

- **D-04-14:** `_Noreturn` on the generated `lang_defect` function is **exempt**
  from D-04-13. It is not a claim about a foreign callee; it is a property of a
  function cgen itself emits, every path of which provably ends in `abort()`.

### Panic and foreign nonlocal exit

- **D-04-15:** Phase 4 introduces a **real but deliberately terminal `defect`
  operation**: abort-only at the process root, no catch, no containment, no
  unwinding, and **no cleanup**. This matches `wiki/semantic-kernel-contract.md`
  verbatim — *"an abort-only profile can host panic only at its process/root
  boundary"* and *"a process-root fatal termination does not promise external
  cleanup."*
  — **Reversibility:** reversible — containment and typed defect propagation are
  additive to this shape; nothing in Phase 6's actor/supervision space is
  foreclosed.

  **Why not satisfy SC3's first half negatively.** "There is no unwind path" is
  good evidence about the *boundary* and worthless evidence about *panic* if
  panic does not exist — a criterion whose evidence has an empty reachable input
  space by construction is the maximal instance of the failure mode that already
  cost this project three gap-closure plans (03-08, 03-09, 03-10; standing rule
  D-10). Adding the `defect` terminator costs one `OperationKind`, depends on
  none of the other decisions (a defect terminator in a match arm needs no call
  surface), and converts the claim into one with a reachable witness.

- **D-04-16:** SC3's second half is **detection, not prevention** — the criterion
  says "cannot **silently** bypass". Prevention against opaque C is unachievable
  and claiming it would be an unproven attribute, the same class as the false
  `restrict`. Prevention *is* delivered statically at the declaration layer:
  **a foreign declaration that does not state its `unwind`/`nonlocal_exit` policy
  is refused admission**, independently, by both `check` and `corevalidate`.
  There is no default value.

- **D-04-17:** Detection is one **process-root `setjmp` landing pad** (one per
  process — not per call, not per borrow; the wiki's cost constraint is respected)
  plus a **cleanup ledger in `static` storage**. On a foreign nonlocal exit the
  pad emits `foreign.nonlocal_exit` plus one `resource.leaked` per still-live
  acquisition, then terminates as a defect.

- **D-04-18:** The pad **must not run releases**. C17 §7.13.2.1 leaves non-`volatile`
  automatic objects changed since `setjmp` indeterminate after `longjmp`;
  releasing from indeterminate handles converts a leak into a use-after-free.
  The ledger lives in `static` storage for the same reason — a stack-resident
  ledger would itself be UB to read in the pad. This is exactly why spike 005's
  `native/nonlocal_exit.c` used `static int acquired/released`. Enforced as
  `control:defect.no_release_on_defect` (zero `resource.released` events after a
  defect terminal record), so it is an invariant rather than a comment.

- **D-04-19:** The non-unwind control is **`nm -u` on the linked binary** —
  underscore-normalized for the Mach-O/ELF difference, **allowlist** (not
  denylist), so any new undefined symbol fails until a human adds it with
  rationale. It must return operational/`tool_missing` on a host without `nm`,
  never pass. **Do not** build the control on unwind-section absence: measured on
  this host (Apple clang 21, arm64), `-fno-exceptions
  -fno-asynchronous-unwind-tables -fno-unwind-tables` compiled clean under the
  existing `-Werror -pedantic` and `otool -l | grep -c __unwind_info` returned
  `0`, but Darwin/arm64 mandates compact unwind for many function shapes, so a
  section assertion would pass or fail for the wrong reason.

- **D-04-20:** Events must be **streamed at the point they occur**, not
  accumulated and written once at the end. Today's buffered `lang_write_events()`
  means an aborting process emits **zero** events, which would make SC4's
  interpreter/native agreement unfalsifiable on exactly the paths SC3 is about.
  The terminal record must be the last thing written, and its **absence must be a
  failure, never a tolerated truncation** — the bounded-writer caps
  (`MaxStreamBytes`, `LANG_OUTPUT_LIMIT`) can otherwise make a missing release
  event look like a passing run.
  This is an **additive emitter path**; `emitEventSupport` and the Phase 1/2/3
  emitters are not touched, so those goldens stay byte-identical (D-13).

### Method — carried forward from Phase 3, still binding

- **D-04-21:** D-09/D-10/D-11 stand unchanged. Mutation-kill every oracle
  (revert-and-fail in a throwaway detached worktree). Interrogate what inputs a
  green property test actually *reaches*. Drive the shipped `./cmd/lang` binary
  on hand-written **out-of-corpus** programs, not only the gate's own corpus.

- **D-04-22:** D-12/D-12a stand and **widen**. `check` and `corevalidate` remain
  independent derivations with no shared helpers. A new `OperationKind` must now
  be handled at **six** sites — `check`, `corevalidate`, `interp`, `cgen`,
  `pathoracle`, `originvalidate` — and several of those have **two** dispatch
  switches each (`corevalidate` straight-line + branch walkers; `interp`
  `runLinear` + `runBranchArm`; `cgen` both walkers). Required control:
  `control:kind.exhaustive_dispatch`, a table-driven test over a new
  `core.AllOperationKinds()` registry asserting every kind is handled at every
  site, mutation-killed by deleting one `case`.

- **D-04-23:** D-13 stands. No `lang.core/2`. Every new core field is additive and
  `omitempty`. Assert byte-identity of Phase 1/2/3 core programs **and**
  evidence-manifest IDs **by test**, before anything else lands.

- **D-04-24:** D-15 stands and is newly load-bearing: a test that runs a program
  which `abort()`s or `longjmp`s must handle nonzero exit, signals, and truncated
  output within the existing bounds. The SIGABRT assertion goes through
  `ProcessState.Sys().(syscall.WaitStatus)` — **never** a hardcoded `134`, which
  is a shell's encoding and this codebase correctly uses no shell.

### Carried Phase 3 debt — disposition

- **D-04-25:** **D-03-01 gets the cheap half in Phase 4:** instrument
  `discoverLoanLastUses` to count its own transitive-scan work honestly, matching
  the `result.Work++`-per-operation convention `blockLoanLiveness` already
  established. This closes the metric-honesty half of D-05.

- **D-04-26:** **Retiring `discoverLoanLastUses` carries to Phase 5.** The
  complete fix D-05 intended (drive both admission and endpoint materialization
  from `loanLivenessFixpoint`) is deliberately deferred: performing that surgery
  inside the same `check.go` that is simultaneously gaining `OpForeignCall`,
  `OpFail`, `OpRelease`, and `OpDefect` is precisely the compounding-wave-defect
  shape Phases 2 and 3 both hit. Re-record it as dated open debt in `04-DEBT.md`.

- **D-04-27:** **WR-01 is closed this phase:** wire
  `originvalidate.ValidatePublished` into `lang check`/`lang run`, not only the
  `interface export` path. This is newly load-bearing because foreign
  declarations become a **second surface where an alias fact can be silently
  absent**, and `originvalidate` must learn about foreign returns regardless
  (D-04-28). Per D-08, fold it into whichever plan already touches
  `originvalidate`; do not create a debt-cleanup plan.

- **D-04-28:** `originvalidate` must recognise a value returned from an
  `OpForeignCall` whose declaration says it borrows or retains an argument as
  **borrow-derived**, with a new `core.foreign_origin_omitted` refusal. Without
  this, D-04-01's second-order risk is live: the `foreign` declaration surface is
  itself a signature carrying origin and access facts, so foreign-only does not
  escape origin reasoning — it moves it somewhere the facts are *asserted* rather
  than derived.

- **D-04-29:** `originvalidate` and `pathoracle` must walk **every** terminator —
  `{OpReturn, OpFail, OpDefect}` — not just `OpReturn`. Both walk only `OpReturn`
  today. **This is the single highest-risk item in the phase:** it reproduces
  exactly the defect class 03-08/03-09/03-10 closed three times (an analysis that
  walked *a* return rather than *every* return), and if either walker misses a
  terminator the failure path becomes invisible to origin verification and path
  enumeration **while the O0/O3 differential still passes**, because it only
  compares paths the oracle enumerated. Required control asserting the walked
  terminator set equals the full set, mutation-killed by deleting `OpFail` from
  either walker.

### Claude's Discretion

Plan decomposition, wave structure, plan count, the concrete spelling of the
`foreign` block and of `try` / `discard … because`, `OperationKind` naming (the
four research fan-outs proposed overlapping names — `OpForeignCall`,
`OpAcquire`/`OpResourceAcquire`, `OpRelease`/`OpResourceRelease`, `OpFail`,
`OpDiscard`, `OpDefect` — reconcile into one minimal set at planning time), the
exact `lang.foreign/0` field layout, the fixture type shape, and the diagnostic
code names — provided every decision above holds and the D-04-22 six-site
dispatch obligation is discharged.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase scope
- `.planning/ROADMAP.md` §"Phase 4" — goal, four success criteria, canonical refs
- `.planning/ROADMAP.md` §"Phase 5" — the Phase 4/5 line; NAT-03's seven hostile
  mutations are Phase 5's to run, and Phase 4 must leave them something to attack
- `.planning/REQUIREMENTS.md` — SEM-03, RES-01, FFI-01
- `.planning/PROJECT.md` §"Out of Scope" — bounds how much panic/isolation and
  generics machinery is legitimate here

### Contract and intent
- `wiki/semantic-kernel-contract.md` §"Failure, cancellation, and panic" — the
  authoritative panic intent; the abort-only profile; "panic never unwinds across
  a foreign ABI"
- `wiki/semantic-kernel-contract.md` §"Resources and cleanup" — split `release`
  from fallible `finish`; reverse completed-acquisition order; "only fully
  initialized fields release"
- `wiki/semantic-kernel-contract.md` §"Layout, validity, unsafe, and FFI" —
  safe memory laws; unsafe as an obligation boundary; **foreign declarations are
  quarantined claims** (the source of D-04-10's reframe)
- `wiki/semantic-kernel-contract.md` §"Backend-independent optimizer contract",
  §"Observable behavior and telemetry"
- `wiki/native-and-low-level-profile.md` §"FFI and C", §"Checked semantics and
  optimization", §"Unsafe is a reviewed component, not a magic word",
  §"Compiler/backend contract"
- `wiki/memory-reclamation-policy.md` §"Cleanup must be boring" — implicit
  reclamation may only do compiler-known non-failing structural work

### Prior art
- `.planning/spikes/005-native-ffi-provenance-cleanup/README.md` — the direct
  validated prior art. Iteration 2 (linker success is not ABI evidence),
  iteration 3 (partial init + reverse-order cleanup + allocator identity),
  iteration 4 (false `restrict` changes the answer at `-O3`), iteration 5
  (`longjmp` bypasses cleanup with no sanitizer report). Gates 1–9.
- `.planning/spikes/005-native-ffi-provenance-cleanup/native/` and `lab/` — read
  the actual C and Go, not only the README. `native/nonlocal_exit.c`'s
  `static int acquired/released` is the ledger D-04-17 generalizes.
- `.planning/spikes/004-independent-certificate-checker/README.md` — keeping the
  two admission layers independent

### Phase 3 carry-forward (read before planning)
- `.planning/phases/03-borrowed-views-and-cfg-lifetimes/03-CONTEXT.md` — D-01..D-15
- `.planning/phases/03-borrowed-views-and-cfg-lifetimes/03-DEBT.md` — D-03-01
  (open), the D-03-02 closure, the multi-arm closure, and WR-01. **Read the
  D-03-02 entry in full** — its "currently unreachable because there is no
  consumer" is the load-bearing premise of D-04-01.
- `.planning/phases/03-borrowed-views-and-cfg-lifetimes/03-VERIFICATION.md` — the
  truths this phase must not regress
- `.planning/phases/03-borrowed-views-and-cfg-lifetimes/03-VALIDATION.md` — the
  mutation-kill and generator-reachability registers
- `.planning/phases/02-owned-values-and-abilities/02-OVERRIDES.md` — OV-02-01

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/compiler/core/core.go` — the additive `omitempty` precedent to follow
  verbatim: `Block`/`Edge`/`LoanEndpoint`/`MatchArm.BlockID` all landed without
  moving a Phase 1/2 golden byte. Add `core.AllOperationKinds()` as the single
  registry every dispatch site is tested against (D-04-22).
- `internal/compiler/session/session.go` — the control table, lane work
  accounting, and the fail-closed required-control set (eleven today). The
  existing marker-count guards (`>1` and `0` refuse to run rather than mutate an
  ambiguous target) are the pattern the new backend mutation runners must copy.
- `internal/compiler/session/session.go`'s `OwnedBackendMutationRunner` — the
  pattern for mutating the *emitter's own output*, which is what makes the
  omitted-release control attack a different artifact than the layout control.
- `scripts/verify-phase3.sh` + `scripts/assert-go-tests.sh` — the bounded gate and
  fail-closed exact-target selector. Phase 4 **copies the shape into
  `verify-phase4.sh`**; it does not fork the frozen Phase 3 script and does not
  extract a shared helper (the same duplication rationale recorded there applies).
- `internal/compiler/evidence/evidence.go` — content-bound manifests; gains
  `foreign_digest` omitempty.
- `internal/compiler/interp/interp.go` — its `default:` case already errors on an
  unknown operation kind. That is the fail-closed model the other packages should
  copy, not the one to relax.

### Established Patterns
- Negative controls are exact IDs required fail-closed by the gate, each with
  nonzero counted work.
- Semantic identity uses function-local ordinals, never source offsets.
- Evidence claims content identity only; it never overclaims translation proof.
- Generated C keeps one global ordinary-identifier namespace with the `_LANG_`
  prefix; source-derived names when unique, deterministic suffix only on real
  collision.

### Integration Points
- Any new lane joins the gate's required-control set and must carry nonzero work.
- Any new diagnostic joins the repair-bearing `lang.diagnostic/1` taxonomy;
  identity is `{schema, code, span, causes, repair_kinds}`, so prose changes must
  not move a diagnostic ID.
- `internal/compiler/native/native.go` — **`validateExecution` currently
  hard-rejects any `Outcome.Kind != "returned"`, any nonzero exit, any stderr
  byte, and any non-empty `LiveResources`. Typed failure, defect-abort, and the
  nonlocal-exit case all violate that, so an aborting program cannot be run at
  all today.** It needs an explicit expected-terminal-outcome axis, widened
  precisely, plus a regression test asserting Phase 1–3 documents are still
  rejected on the old grounds.
- `internal/compiler/cgen/cgen.go` — `emitEventSupport` buffers events and writes
  once at the end; the new resource/foreign boundary needs an **additive**
  streaming emitter path (D-04-20), leaving the existing emitters untouched.
- `internal/compiler/native/native.go` — `Run` writes and compiles one source;
  it must grow multi-TU compilation, with the conformance TU compiled as a
  **separate invocation** so its failure is distinguishable from the program's.

</code_context>

<specifics>
## Specific Ideas

**The three-stage fixture is mandatory, not optional.** With only two
acquisitions the release *set* and the release *order* coincide, so a
differential over the two-stage fixture SC1 literally asks for stays green even
if `cgen` derives its cleanup order from its own `goto`-ladder structure rather
than from the materialized `OpRelease` list. Phase 4 must ship a **three**-
acquisition fixture and a mutation that **transposes** two `OpRelease` emissions
and turns the differential red. This is the sharpest finding in the whole
research set and it directly instantiates D-10.

**The two mutation directions must attack different artifacts.** Layout mutation
attacks the frozen foreign fixture; omitted-release mutation attacks the
production emitter's own output. An author keeping both sides aligned cannot
satisfy both controls — passing one requires the emitter to be right, passing the
other requires the fixture to be independent. This is the structural answer to
"an in-repo shim is self-confirming."

**Layered gates keep paying for themselves.** Phase 2's close-out and Phase 3's
three gap-closure rounds both showed each layer finding defects the previous
layer missed, including one defect *introduced* by the previous layer's fix.
Keep review, verification, validation, and security as distinct passes.

**Roadmap gap surfaced by this discussion, for the record:** NAT-03's "false
no-alias facts" hostile mutation has **no subject** until Phase 5 itself first
adds a *proven* alias-fact emission path — because D-04-13 means Phase 4 emits no
attributes to falsify. That dependency is not currently stated in ROADMAP.md
§Phase 5 and should be added when Phase 5 is planned.

</specifics>

<deferred>
## Deferred Ideas

- **Lang-to-Lang calls (`OpCall`)** — Phase 5, gated on **callable ⊆ publishable**
  (D-04-03). Brings with it interprocedural loan liveness in both admission
  layers, call-graph construction, cycle refusal, a bounded interpreter call
  stack, and rebuilding the Phase 3 exhaustive differentials cross-function.
- **A storable, matchable `Result` value and payload-carrying alternatives** —
  M002 or Phase 6, when error *translation between layers* is actually needed.
  The additive move is a sibling `alternative_details []Alternative omitempty`
  keyed by name, never a shape change to `Alternatives []string`. Record as debt
  now rather than discovering it later (D-04-04).
- **Retiring `discoverLoanLastUses`** — Phase 5 (D-04-26).
- **Allocator-mismatch and stale-callback-retention detection** — Phase 5's
  NAT-03. Phase 4 only carries the allocator identity in the contract and
  independently refuses a release whose declared allocator differs from its
  acquisition's.
- **ASan/UBSan lanes** — Phase 5, isolated from semantic-equivalence evidence.
- **Callback registration, generation tokens, dynamic-library handles, `errno`
  translation, symbol versioning, `blocking`/`reentrant` enforcement, header
  ingestion** — declared and refused-if-non-default this phase; never exercised,
  because the language has no calls-into-Lang, no closures, and no threads.
- **Panic containment, isolation boundaries, `catch_unwind` analogues,
  cancellation shields/budgets/checks** — Phase 6 and beyond; PROJECT.md defers
  actors, supervision, and structured task scopes.
- **Loop-carried loan liveness** — still deferred; this phase adds no loop or
  recursion construct (Phase 3 scope note).
- **Full DWARF/CodeView emission, crash capture, proof/SMT tiers** — unchanged
  from Phase 3's D-03.

### Accepted residual limitations (record, do not engineer around)

- **The coordinated three-way lie.** An author who simultaneously edits the
  frozen fixture, the Lang `foreign` declaration, and the conformance TU's
  expectations passes Phase 4 on a wrong boundary. This is QLT-01's documented
  source-to-core escape class, one layer down.
- **The pad cannot see a `longjmp` to a foreign-owned `jmp_buf` established below
  it** (reachable only through foreign-invoked callbacks), and foreign
  `exit()`/`_Exit()`/`raise()` is not observable at all.
- **One small by-value record on one host.** Unions, bit fields, vectors,
  variadics, packed/aligned records, and aggregate-passing thresholds remain
  open, as do ELF, x86-64, GCC, and non-Apple Clang. Spike 005's Known
  Limitations fence this and Phase 4 inherits the fence.
- **Quarantine is permanent and non-discharging.** A green gate never proves the
  foreign implementation obeys its declared contract; the manifest's
  `unchecked_obligations` list is the honest record.

</deferred>

---

*Phase: 4-Fallible Resources and C Boundary*
*Context gathered: 2026-09-04*
