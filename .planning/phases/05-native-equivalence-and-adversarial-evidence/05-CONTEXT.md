# Phase 5: Native Equivalence and Adversarial Evidence - Context

**Gathered:** 2026-09-05
**Status:** Ready for planning
**Source:** advisor-mode discussion at Phase 04 close-out; eight parallel research
fan-outs (alias-fact emission, callback retention, allocator mismatch + UAF,
sanitizer lane design, equivalence corpus + axes, minimization + causal trace,
QLT-01 spike audit, carried debt + `OpCall`), each run through a multi-role
stakeholder pass plus an adversarial false-green pass, then cross-checked
against the shipped tree, the wiki corpus, and the Phase 2/3/4 failure history.
The user directed that the synthesized recommendations be adopted directly.

<domain>
## Phase Boundary

The source-to-native subset built in Phases 1–4 must survive `-O3` and
deliberately seeded boundary defects, with each defect caught by an
*independently meaningful* lane. Requirements INT-02, NAT-02, NAT-03, QLT-01.

This phase is **evidence-shaped, not surface-shaped**. It adds no new language
capability the roadmap did not ask for. What it adds is machinery the repository
does not have at all today: the first optimizer-attribute emission path derived
from a checked fact, the first ASan/UBSan lanes, the first test-case reducer,
the first LTO build tier, and the first executable traceability registry.

**The constraint that shapes the whole phase:** three of NAT-03's seven hostile
mutations currently have **no subject to attack**. Phase 4 emits zero
optimizer-visible attributes (D-04-13), so "false no-alias facts" has nothing to
falsify — recorded as D-04-32. The language has no closures, no calls-into-Lang
and no registration surface, so "stale callback retention" has nothing to
falsify. And the checker's affine ownership plus CFG-precise loan liveness makes
source-level use-after-free unconstructible by design. Phase 5's first job is
therefore to *create honest subjects*, and its second job is to prove the lanes
that attack them are not vacuous.

Phase 4 already ships three of the seven: layout mismatch, missing partial
cleanup (release-omitted and release-order-transposed), and foreign
nonlocal-exit cleanup bypass.

**A second-order finding that outranks all of the above (D-05-24):** a green
`-O0` vs `-O3` differential over today's corpus proves close to nothing.
Without `-flto`, a call to a symbol defined in a separate translation unit is
opaque at the call site — Clang cannot inline it, cannot prove it pure, cannot
eliminate effects across it. Every current corpus program's only observable
effect crosses exactly that boundary. So the existing differential is green *by
construction*, not by evidence, which is precisely the failure shape D-09/D-10
exist to prevent.

</domain>

<decisions>
## Implementation Decisions

### Alias-fact emission — giving "false no-alias facts" a subject

- **D-05-01:** Phase 5 emits **exactly one optimizer-visible attribute:
  `restrict`, and only on parameters of functions cgen itself generates**, only
  where `check` proves and `corevalidate` independently re-confirms that the
  parameter is exclusively borrowed or uniquely owned for the entire duration of
  the call. This is the same boundary D-04-14 already drew for `_Noreturn` on
  `lang_defect`: an attribute is admissible when it is a property of a function
  cgen owns end-to-end, and inadmissible when it is a claim about opaque code.
  **`restrict` is never emitted on a foreign extern declaration.** No `noalias`,
  no `nonnull`, no `returns_nonnull`, no `__attribute__((malloc))`, no
  LLVM-IR-level attributes.
  — **Reversibility:** costly — removing the attribute later is one line, but
  `control:alias.false_no_alias` and the manifest justification binding are built
  on top of it, and NAT-03 loses its subject again if it goes.

  **Why stop at C-source `restrict`.** Rust spent five years fighting exactly the
  next step up this lattice: `noalias` on `&mut` was disabled in 2016 after real
  LLVM miscompiles (rust-lang/rust#31681, PR #31545), lived behind
  `-Zmutable-noalias` for years, was only re-enabled by default on LLVM 12 in
  2021 — and regressed once more (#84958). C17 §6.7.3.1 `restrict` is the older,
  simpler, better-understood half of the same lattice, it is a promise about the
  *parameter object for the duration of that function's execution*, and that is
  exactly what an exclusive loan already proves. Swift's exclusivity enforcement
  gives the complementary rule: never claim a stronger guarantee than the static
  prover actually established.

- **D-05-02 — THE BLOCKING PRECONDITION, verified in the shipped tree during
  this discussion:** `cgen` emits Lang-derived functions with **by-value**
  parameters — `fmt.Fprintf(&out, "static %s %s(%s %s) {\n", typeName,
  functionName, typeName, parameterName)` at `internal/compiler/cgen/cgen.go:157`
  — and `testdata/phase2/owned_transfer.golden.c` confirms the only pointer
  parameters in emitted C belong to cgen's own runtime helpers
  (`lang_write_buffer_hex(const LANG_BUFFER *value)`, `lang_write_bytes(const
  char *data, size_t length)`). **A by-value struct parameter cannot be
  meaningfully `restrict`-qualified. D-05-01 therefore has no subject either
  until a by-pointer lowering exists.**

  Phase 5 must build that lowering first: an **additive** emitter path that
  lowers an exclusively-borrowed `Buffer` parameter by pointer, selected only for
  the new Phase 5 fixtures, leaving the Phase 1/2/3/4 emitters and their goldens
  byte-untouched. This is exactly the additive-emitter-path pattern D-04-20
  already established for streaming events, and it is the reason the alias-fact
  work is **two** deliverables, not one.
  — **Reversibility:** costly — a second lowering path for the same source shape
  is a lasting fork in cgen that later phases must keep coherent.

  **If the planner concludes the by-pointer lowering is too invasive to land
  safely in this phase, that is a `checkpoint:decision`, not a silent
  substitution.** Emitting `restrict` on a by-value parameter, or on a runtime
  helper whose non-aliasing is a cgen-internal convention rather than a checked
  ownership fact, would be a false claim of precisely the class D-04-13 refused.
  Do not do it. Escalate instead.

- **D-05-03:** `control:foreign.no_unproven_attributes` (D-04-13) is **narrowed,
  never deleted**, and stays in the required-control set. Its new form asserts
  two things: (a) foreign extern declarations carry zero attributes — unchanged
  from Phase 4 and still mutation-killed by injecting `restrict` there; and (b)
  **every** entry in the manifest's `emitted_attributes` has a matching
  justification that `corevalidate` independently re-derived. An attribute with
  no proven backing fact is a hard build failure, not an absent mutation signal.

- **D-05-04:** The manifest records the attribute→proof binding as
  `emitted_attributes: [{attr, core_node, parameter, justified_by: <borrow-fact
  identity>}]` — additive and `omitempty` per D-13, on the existing
  `lang.evidence/1` manifest, no schema bump. `emitted_attributes: []` was
  deliberately made a *field* rather than an omission in Phase 4 precisely so
  Phase 5 could populate it. `corevalidate` re-derives the justification from the
  core artifact's own loan facts; it never reads what `cgen` wrote (D-12).

- **D-05-05:** `control:alias.false_no_alias` injects `restrict` on a parameter
  the checker did **not** prove exclusive, and asserts the observable signature
  `interpreter == -O0 != -O3`. The fixture is **engineered, not discovered**: it
  must contain a repeated load/store or loop-shaped access over the falsely
  `restrict`-ed parameter that Clang reliably hoists or vectorizes at `-O3` and
  not at `-O0`, matching the shape spike 005 iteration 4 measured (returns `2` at
  `-O0`, `1` at `-O3` on Apple clang 21 / arm64). **A mutation that produces no
  divergence is a broken fixture and must fail the lane, never pass it** — the
  compiler version is recorded in evidence so a toolchain upgrade that kills
  observability surfaces as a red lane rather than a quiet false green.

### Stale callback retention — reframed, with a named residual

- **D-05-06:** Phase 5 does **not** build a callback-registration seam. NAT-03's
  "stale callback retention" is subjected as a **retained-pointer lifetime
  defect**: a hostile frozen foreign TU stashes a pointer to a borrowed buffer in
  `static` storage and dereferences it after the Lang side has released it. This
  is the FFI-003 shape, it needs no calls-into-Lang, and it is caught by the
  sanitizer lane (D-05-10).
  — **Reversibility:** reversible — a real registration surface remains additive
  if a later milestone needs one.

  **Why not build the seam, when Phase 4 built the `defect` terminator for the
  symmetric problem.** D-04-15 built its subject because it cost one
  `OperationKind` and depended on nothing else. This case is the mirror image:
  a function pointer into cgen-emitted Lang code *is* calls-into-Lang, which
  reopens D-04-01 and creates the consumer for D-03-02's known-unreachable
  unsoundness, and forces the cross-function differential rebuild Phase 4 spent a
  whole decision avoiding. Worse, the cheap version is self-defeating: a
  synchronous single-shot callback that fires before anything can go stale has an
  empty reachable input space for *staleness* — it would re-create D-10's failure
  mode one level down while looking closed. And every mature ecosystem surveyed
  (GObject `GWeakRef`, JNI global/weak-global refs, N-API `napi_ref`, Rust
  `Box::into_raw` + `catch_unwind`) locates callback safety at the retained
  context pointer, not the invocation mechanism. ASan's documented blind spot
  (google/sanitizers#652) is specifically *deferred* callback firing, so a real
  seam would not even buy better coverage.

- **D-05-07:** The residual is **named and gate-visible**, not absorbed. Add
  `escape:callback-invocation-unsubjected` to the Phase 5 escape set, worded:
  *"NAT-03 stale-callback-retention is subjected via the retained-pointer lane
  (`static`-stashed borrowed pointer dereferenced post-release, detected by
  ASan). The callback-invocation mechanism itself — registration, later firing,
  generation tokens — has no subject, because M001 has no calls-into-Lang by
  design (D-04-01). The pointer-lifetime root cause is tested; the
  invocation-time variant is not."* NAT-03 is honestly discharged as **six
  subjected plus one subsumed-with-named-escape**, never as a silent 7/7.

### Allocator mismatch and use-after-free

- **D-05-08:** Allocator mismatch keeps Phase 4's **static** contract-layer
  refusal unchanged and gains a **dynamic** half that attacks a *different
  artifact*: a **second frozen foreign TU exposing a real second allocator
  identity** (`posix_memalign`/`aligned_alloc` paired with its matching free),
  resolved by declared symbol exactly as the existing frozen TUs are. The fixture
  allocates through the arena allocator and routes the release through the
  plain-malloc wrapper's `free`. Detection is ASan's own
  **`alloc-dealloc-mismatch`** diagnostic — an existing, load-bearing, pointer-tag
  check, so this project writes zero new detection logic.
  **`alloc_dealloc_mismatch=1` must be pinned explicitly in the harness: it
  defaults to OFF on macOS**, and a contributor on this host running defaults
  would get a false green with no warning.

- **D-05-09:** Use-after-free is injected into a **hostile variant of a frozen
  foreign TU** that returns a pointer into storage the wrapper has already freed,
  exercised through a real allocate/use/free/use-again cycle. The dangling read
  must feed an observable the optimizer cannot eliminate, so it is not dead-store
  removed. **It must not be injected into cgen's own release emission** — that
  artifact is already attacked by Phase 4's release-omitted and
  release-order-transposed controls, and reusing it would collapse two of the
  seven distinct questions into one. Nor may it be a standalone hostile C file
  unrelated to the boundary: that tests ASan, not this compiler.

  This preserves Phase 4's independence rule verbatim — the static allocator
  check attacks contract metadata, the dynamic checks attack foreign fixtures,
  and the release controls attack the emitter's own output, so no author keeping
  two sides aligned can satisfy all of them.

- **D-05-10:** Detection for both is **ASan only, never a bare native run**.
  A plain `-O0`/`-O3` run of either fixture is UB, not a guaranteed crash: the
  freed block is often re-served intact and the program completes cleanly.
  ASan's redzones plus its 256 MB default quarantine make the trap deterministic
  regardless of allocator reuse timing. **Do not build determinism on
  `MallocScribble` / `MALLOC_PERTURB_`** — those perturb freed bytes without
  trapping the access, and they differ between macOS libmalloc and glibc, so they
  would fail this project's cross-host requirement.

### Sanitizer lane design

- **D-05-11:** Isolation is **structural, not procedural**. A dedicated
  `lane:native-sanitize` compiles and links a **separate binary** from the same
  generated C17 with
  `-std=c17 -O1 -g -fno-omit-frame-pointer -fsanitize=address,undefined
  -fno-sanitize-recover=all`. That binary is never handed to the equivalence
  comparator. The guarantee is enforced at the **type level** — the comparator's
  signature simply does not accept a sanitizer-lane evidence value — because a
  boolean "sanitizer-tainted" flag on a shared table is one refactor away from
  laundering sanitizer noise as a semantic mismatch.
  — **Reversibility:** reversible.

- **D-05-12:** `-O1`, not `-O0` and not `-O3`. `-O0` is too far from shipped
  codegen to answer the question; `-O3` raises false-positive and
  inlining-obscures-blame risk, and no major sanitizer CI (Chromium, OSS-Fuzz,
  KASAN) runs its primary tier at full optimization. If `-O3`-specific UB later
  matters it is added as a **named deferred obligation**, never silently skipped.

- **D-05-13:** Options are pinned **by the harness, not contributor-configurable**:
  `ASAN_OPTIONS=halt_on_error=1:abort_on_error=1:symbolize=0:detect_leaks=0:detect_odr_violation=0:alloc_dealloc_mismatch=1`
  and `UBSAN_OPTIONS=halt_on_error=1:print_stacktrace=0`. `detect_leaks=0` is
  explicit because LeakSanitizer is unreliable-to-absent on Apple clang arm64
  (llvm/llvm-project#115992). UBSan runs the `-fsanitize=undefined` umbrella
  minus the noisy non-UB additions `implicit-conversion` and
  `unsigned-integer-overflow`, which are **named as an escape rather than
  silently dropped**; `alignment`, `bounds`, `pointer-overflow`, and `nullability`
  stay on because they map onto FFI-boundary defect classes.

- **D-05-14:** A clean sanitizer run is **not** evidence. The lane carries an
  **always-on positive control** — the deliberately-defective retained-pointer
  fixture from D-05-06 — required to produce a report on *every* `verify`
  invocation, not merely once at authoring time. Controls key on the specific
  diagnostic substring **plus** exit code (`ERROR: AddressSanitizer:
  heap-use-after-free`, `alloc-dealloc-mismatch`, UBSan `runtime error:`), never
  on "process exited nonzero" alone, because abort-without-report is exactly the
  false-green mode. Two required controls:
  `control:native.sanitize.retained_pointer` and
  `control:native.sanitize.ubsan_no_recover` — the latter proving
  `-fno-sanitize-recover=all` actually took effect, which is what catches a
  typo'd options string.

- **D-05-15:** Absence of the sanitizer runtime reports **`tool_missing` /
  operational**, never pass — the exact posture `control:foreign.unwind_forbidden`
  took for `nm` in D-04-19. Availability is probed by compiling and running a
  two-line ASan+UBSan smoke fixture.

- **D-05-16:** Symbolized stack traces carry absolute paths and ASLR-influenced
  addresses and are **never admitted verbatim into the content-bound manifest**.
  Only a stable classification tuple `{sanitizer, check_kind,
  report_signature_substring, exit_code}` is hashed; raw stderr is retained as a
  truncated artifact under the existing 64 KiB-plus-one bound with its stable
  stage/stream truncation code.

- **D-05-17:** Cost placement: the sanitizer lane does **not** run in
  `edit`/`check`. It is mandatory in `verify` and `release`, listed in the Phase 5
  required-control set duplicated verbatim between
  `Phase5RequiredControls()` and `scripts/verify-phase5.sh`, with a
  `TestPhase5RequiredControlsMatchScript` mirroring the Phase 4 test — so a
  missing or `tool_missing` lane is a visible named obligation, never an implicit
  pass.

### The equivalence corpus and its axes

- **D-05-18:** The milestone corpus is the union of the existing hand-written
  `testdata/phase1..phase4` sets (kept **byte-identical**, D-13) plus a new
  `testdata/phase5` set in two parts:
  **(a) a hand-written adversarial subset of 6–10 programs**, each engineered to
  trigger one *named* `-O3`/LTO transformation — inlining across the foreign call
  under `-flto`, dead-store elimination of an unused borrowed acquire, reordering
  of two independent event emissions, tail-collapse of a release ladder, and a
  typed-failure / `defect` path that dies by signal or truncates stdout; and
  **(b) a bounded enumerated closure** generated from the typed core over the
  current grammar — one parameter, `Byte | Buffer`, ≤2 ADT alternatives,
  straight-line or single-level-branch body, 0 or 1 foreign call, `borrow`/`take`/
  `try`/`discard` as the only fallible consumers.
  The enumeration bound (e.g. depth ≤3, ≤4 statements) is a **versioned constant
  asserted equal between Go and `scripts/verify-phase5.sh`**, mirroring the
  required-control-set duplication pattern, so corpus scope cannot drift silently
  any more than a control set can.
  — **Reversibility:** reversible — the bound is a constant.

  **Deliberately NOT adopted:** the wiki's 64-workload expansion, the 24 blocking
  probes, and the extended semantic matrix. Those are a superset design for a
  later, larger language. Claiming them over a grammar with one parameter, two
  types, and no loops would be overclaiming.

- **D-05-19:** `-O3` alone is insufficient and a new sampled tier
  `lane:native-differential-lto` / `control:interpreter-o0-o3-lto` builds the
  **hand-written adversarial subset only** at `-O3 -flto`. Without LTO the
  separately-compiled foreign TU is opaque at the call site, so the optimizer
  cannot legally perturb the very code paths SC1 is about, and the existing
  differential is green by construction. Restricting LTO to the adversarial
  subset keeps the real link-time cost bounded **while guaranteeing by
  construction — not by sampling luck — that the perturbable programs are the
  ones actually built with it.**
  — **Reversibility:** reversible.

- **D-05-20:** Compared axes are SC1's three, each made explicit rather than
  assumed: terminal outcome as the closed `value | typed_failure | defect` set
  **including the ok payload and the err-edge ADT alternative**; semantic-event
  order with causal identity; live-resource ledger including released-count.
  Two axes are **added** because SC1's wording silently omitted them: **exit
  status and signal** for abort/defect programs, and **diagnostic-ID equivalence
  for reject-programs** across engines — the latter under its own control,
  distinct from `interpreter-o0-o3`, since reject-programs never execute.

- **D-05-21:** The exclusion list — addresses, wall-clock time, allocator
  identity, hash/PRNG seed, thread identity — becomes an **asserted list in the
  comparator**, with a **fail-closed field-routing test**: adding a field to
  `Outcome`, the event record, or the resource ledger without routing it to
  either "compared" or "excluded" fails the build. Today the exclusion is prose
  in the wiki and default behaviour in code, which is the shape that lets a
  pointer leak into a compared field and either flake or pass by accident of
  allocator determinism.

- **D-05-22:** Every one of NAT-03's seven mutations must **cite, in
  `scripts/verify-phase5.sh`, the specific corpus program it is injected into**,
  and a test must assert the mutation actually **moves at least one compared
  axis** on that program before the mutation is trusted as a detector. A mutation
  injected into a program whose behaviour it does not change is a false green
  wearing a control's name.

### Mismatch minimization and causal trace

- **D-05-23:** Reduction operates on the **typed core artifact only** — the
  operation list and block graph — never on source text. `ddmin` and C-Reduce are
  offset- and text-based, which collides head-on with this project's "function-local
  ordinals, never source offsets" identity discipline, and text reduction can
  produce candidates that are structurally invalid. The reduced **source** case is
  a pretty-printed *projection* of the reduced core, produced by the same
  reduction run — never independently re-parsed and re-reduced, which would yield
  two non-corresponding "smallest" cases.
  — **Reversibility:** reversible.

  Exactly **five reduction moves**, applied greedily in a **fixed deterministic
  order** to fixpoint, with no randomization and therefore no seed to record:
  (1) drop an unused binding; (2) drop an unreachable/non-matching `match` arm;
  (3) drop a foreign acquisition/release stage not on the diverging causal path;
  (4) collapse a two-arm branch to the arm containing the divergence; (5) truncate
  a straight-line body to its terminator plus the minimal prefix reaching the
  diverging operation. This is HDD with a five-move alphabet — the right size for
  a grammar with one parameter and no loops. Do **not** build a general
  delta-debugging engine; `bugpoint`'s documented failure was exactly a
  pass-oblivious generic reducer that was slow and confusing.

- **D-05-24:** The interestingness predicate is
  `same diverging axis AND same first-diverging OperationID (or causal role, if
  position shifts under reduction) AND same disagreeing engine pair`. A candidate
  that still fails, but on a different axis, at a different operation, or between
  a different engine pair, is **rejected as uninteresting**. Additionally, a
  candidate that changes the foreign-boundary call sequence or argument shapes on
  the causal path is rejected — that is the local analogue of C-Reduce's
  UB-drift failure, and the foreign boundary is the one place in this language
  where UB genuinely lives.

- **D-05-25:** Work bound `MaxReductionAttempts = 64` check+compile+execute
  cycles, counted in the same `RecomputedWork` currency as every other lane, with
  each attempt's work summed into `total_recomputed_work`. Exhausting the budget
  is **not** a lane failure; it sets `minimality: "budget_exhausted"` instead of
  `"fixpoint"`. That is INT-02's own "smallest **known** case" hedge made
  mechanical and auditable rather than silently degrading.

- **D-05-26:** The trace is a **new top-level `lang.mismatch/0` document** — a
  new schema at `/0` per this project's convention, so `lang.diagnostic/1` and
  `lang.execution/0` identities do not move (D-13). Fields: `schema`,
  `diverging_axis`, `engine_pair`, `diverging_operation_id`, `reduced_core`,
  `reduced_source`, `minimality`, `total_recomputed_work`, `reduction_attempts`,
  `event_window` (the last **8** events per engine side, bounded — not the full
  trace), `causal_chain` (the ordered acquisition/borrow/control-edge steps
  leading to the diverging operation), and `causes` — which **reuses the existing
  diagnostic cause-graph shape verbatim**, so an agent parses one cause format
  project-wide. The primary consumer is an AI repair agent: it must be able to
  propose a fix from this document alone, without opening the full execution log.

- **D-05-27:** Three mutation-kills are required, because a reducer has three
  distinct ways to be vacuous. (a) **No-op reducer** — make `Reduce()` the
  identity; the lane must go red via a strict size-decrease assertion on a corpus
  of known-reducible seeded mismatches. (b) **Predicate too loose** — accept any
  failure; seed two distinct mismatches and assert the reducer for A never returns
  a case matching B's signature. (c) **Non-determinism** — randomize move order;
  assert byte-identical reduced output across repeated runs. A predicate that is
  too *tight* is caught by the same size-decrease assertion in (a).

### QLT-01 — spike control preservation

- **D-05-28:** A **machine-readable control registry** plus an **executable
  audit**, not a coverage document. One row per spike negative control or reduced
  counterexample, carrying `spike_id`, `control_mechanism` (the hazard itself —
  e.g. "false `restrict` diverges between `-O0` and `-O3`", "`longjmp` bypasses
  cleanup with no sanitizer report"), and **exactly one** of `live_descendant`
  (a fixture plus its `control:` ID) or `waived` (`{reason, citation to the
  shipped-language limitation, owner, phase}`).
  **"Not relevant" is valid only as a `waived` row with a specific falsifiable
  citation** — a superseded design decision such as "no generics in M001" or
  "Box/Pair rejected at the execution-admission gate" — never as free text.
  — **Reversibility:** reversible.

  A documentary mapping alone is the DO-178C spreadsheet-RTM failure mode: it
  decays silently and is caught at audit time rather than commit time. This
  project already answered the general form of this question in Phase 4 with
  fail-closed sets duplicated between Go and the shell gate and equality-tested.

- **D-05-29:** `TestQLT01RegistryComplete` walks the registry and fails if any row
  has neither a live descendant nor a waiver, **and** fails if a `control:` ID a
  row cites no longer exists in the shipped required-control set. The second half
  is what prevents the "ported fixture that rotted" species of theater.

- **D-05-30:** The coordinated source-to-core lie gets a **distinct gate-visible
  `escape:coordinated-source-to-core-false-claim`** and — decisively — **an
  adversarial program that actually constructs matching-but-false source/core
  artifacts and is asserted to pass the gate as a named escape.** Demonstrating
  reachability is the only thing that distinguishes this from the prose-only
  version D-04-31 item 1 currently records. The existing
  `originvalidate.KnownEscape` / `corevalidate.KnownEscape` machinery is reused,
  not replaced.

- **D-05-31:** Minimum honest scope: extract **control mechanism and identity**
  from all five spikes into the registry. A full one-to-one code port of every
  spike fixture is explicitly **not** required — controls whose mechanism is
  superseded are legitimately waived with a citation, and force-porting them into
  fixtures that no longer test anything real would be the worse outcome. The
  extraction is a real, non-trivial task — five prose READMEs plus Go and C, none
  machine-readable today — and must be sized as its own planning line, likely the
  single largest task in the phase.

### Carried debt and scope

- **D-05-32:** **`OpCall` is OUT of Phase 5.** The roadmap's Phase 5 goal and all
  four success criteria are about evidence, not language surface. Pulling
  `OpCall` in would land a new `OperationKind` at six dispatch sites (D-12a)
  simultaneously with alias-fact emission, the first sanitizer lanes, the first
  reducer, and the QLT-01 registry — the exact fingerprint of the failure that
  cost Phase 2 a remediation round, Phase 3 a mid-phase gate plus three
  gap-closure plans (one fixing a defect its predecessor's fix introduced), and
  Phase 4 thirteen plans and five review rounds. It would also violate D-04-26's
  own stated precondition.
  — **Reversibility:** one-way for this milestone — M001 ships without
  Lang-to-Lang calls.

- **D-05-33:** **The roadmap must say so explicitly; silence is the real
  footgun.** `OpCall`, interprocedural equivalence, and D-03-02's full closure
  become the lead item of **M002's charter**. No phase in the current roadmap
  owns them — Phase 6 is agent feedback and performance ratification, not
  language surface. Record in the debt register: *"D-03-02 remains open,
  unreachable-but-unfixed, lift condition unchanged (callable ⊆ publishable,
  D-04-03). M001 ships without Lang-to-Lang calls. Interprocedural equivalence is
  explicitly outside M001's proof scope."*

- **D-05-34:** **`discoverLoanLastUses` retirement lands IN Phase 5** — its
  precondition ("`check.go` is not itself landing a new `OperationKind`") holds
  for the first time since D-04-26 was written, and holds *only* because D-05-32
  keeps `OpCall` out. **Phase 5 is the last eligible slot in M001**: Phase 6 has
  no language-surface work, so deferring again would mean either a dedicated
  debt-cleanup plan (violating D-08) or an unowned carry past the milestone.
  Sequence it **early**, before or alongside the alias-fact work, so the later
  sanitizer and reducer work runs against a single liveness law.

- **D-05-35:** The retirement follows the **Rust NLL migration procedure**, not a
  green-test-suite deletion. (a) Broaden `loanLivenessFixpoint`'s scope from
  `checkBranch` arm blocks only to the full domain `discoverLoanLastUses` covers
  — straight-line bodies plus branch arms. (b) Run **both laws in shadow mode**
  over `TestOwnershipSequenceExhaustive`'s full ~113,164-case enumeration and
  `TestBranchSequenceExhaustive`'s per-arm equivalent, **logging** every
  accept/reject divergence; the candidate never gates behaviour during this pass.
  (c) Require **zero divergences** as a hard gate. (d) Only then delete
  `discoverLoanLastUses`, re-point accept/reject onto the fixpoint, and promote
  `core.LoanEndpoint` from decorative to load-bearing. (e) **Any divergence is a
  Phase-5-blocking finding, not a footnote** — fix the fixpoint, never widen the
  old scanner, and re-run the full enumeration before deleting anything.

  Rust ran the old AST borrowck and NLL in parallel migration mode for years and
  deleted the old checker only when the divergence population converged to zero —
  on data, never on a schedule. GitHub's Scientist formalizes the same pattern:
  log mismatches, don't merely assert their absence.

- **D-05-36:** **D-04-33 (`Alias`) is folded into Phase 5's foreign/alias-fact
  emission plan** per D-08 — that plan already has cgen's attribute emission and
  the foreign contract surface open. Do **both** cheap discharges while the file
  is open: add `Alias` to `corevalidate.foreignContractFieldsCSafe` /
  `validCIdentifier` and to cgen's `unsafeForeignContractField` alongside its
  siblings, **and** add the regression test asserting no `EmitForeign*` output
  ever contains an `Alias` value. The guard rejects nothing accepted today; the
  test is what makes a future splice site fail loudly instead of silently
  reopening the closed injection class.

- **D-05-37:** One gap the phase must own that SC1's wording hides: **no prior
  phase proved that `-O3` code generation preserves the CFG-precise last-use loan
  expiry semantics `check.go` gates on.** SC1 re-litigates Phase 3's exhaustive
  differentials and Phase 4's publication gate against the *native* backend for
  the first time. Treat it as an explicit criterion item, not an assumed corollary
  — this is exactly what D-10 exists to force into the open before the gate rather
  than after.

### Method — carried forward, still binding

- **D-05-38:** D-09/D-10/D-11 stand unchanged and are newly load-bearing, because
  three of this phase's deliverables (sanitizer lanes, the reducer, the LTO tier)
  are brand-new machinery whose first green run proves nothing until it has been
  shown it can go red. Mutation-kill every oracle. Interrogate what inputs a green
  test actually reaches. Drive the shipped `./cmd/lang` binary on hand-written
  out-of-corpus programs.

- **D-05-39:** D-12/D-12a and D-13 stand. `check` and `corevalidate` remain
  independent derivations with no shared helpers — this now extends to the
  attribute justification (D-05-04) and the release/liveness law (D-05-35). No
  `lang.core/2`, no `lang.evidence/2`, no `lang.diagnostic/2`; the only new schema
  is `lang.mismatch/0`. Assert byte-identity of Phase 1–4 core programs, goldens,
  and evidence-manifest IDs **by test, before anything else lands**.

- **D-05-40:** Given the phase's size — five net-new subsystems plus a debt
  retirement — a **mandatory mid-phase gate** is adopted in advance, following
  Phase 3's precedent rather than discovering the need after a wave failure. It
  falls after the alias-fact and liveness-retirement work and before the reducer
  and QLT-01 registry, since the latter two consume a settled differential.

### Claude's Discretion

Plan decomposition, wave structure, and plan count (subject to D-05-40's
mid-phase gate and D-05-34's early sequencing of the liveness retirement); the
exact enumeration bound constant in D-05-18; the concrete spelling of lane and
control identifiers beyond those named above; the registry file format and
location for D-05-28 (YAML, JSON, or a Go table — provided it is machine-readable
and audited by D-05-29); the `lang.mismatch/0` field encoding details; the
concrete second-allocator symbol names in D-05-08; and the shape of the
by-pointer lowering in D-05-02 — provided every decision above holds, the
additive-path constraint keeps prior goldens byte-identical, and the D-05-02
escalation rule is honoured rather than worked around.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase scope
- `.planning/ROADMAP.md` §"Phase 5" — goal, four success criteria, canonical refs
- `.planning/REQUIREMENTS.md` — INT-02, NAT-02, NAT-03, QLT-01
- `.planning/PROJECT.md` §"Constraints" (evidence lanes answer distinct questions
  and have explicit cost lanes; optimizer attributes and FFI guarantees must
  derive from checked facts), §"Out of Scope"

### Contract and intent
- `wiki/semantic-kernel-probes.md` §"Differential equivalence" — the authoritative
  match definition and the named-not-dropped exclusion rule
- `wiki/semantic-kernel-probes.md` §"Comparison axes", §"Generator and reducer
  strategy" (reduce by typed AST and semantic identity, preserving the violated
  claim), §"CI lanes" ("Not run is never rendered as pass")
- `wiki/semantic-kernel-probes.md` §"Unsafe and foreign boundaries" — FFI-003,
  FFI-004, FFI-010; §"Optimizer and profile monotonicity" — OPT-001..OPT-008
- `wiki/ownership-evidence-roadmap.md` §"What the testing strategy actually is",
  §"Per-spike execution checklist", §"Reopening and stopping rules"
- `wiki/native-and-low-level-profile.md` §"Checked semantics and optimization" —
  a generated binding is a candidate contract, never proof
- `wiki/compute-efficiency-constitution.md` — cost-lane placement for D-05-17
- `wiki/semantic-kernel-contract.md` §"Backend-independent optimizer contract",
  §"Observable behavior and telemetry"

### Prior art
- `.planning/spikes/MANIFEST.md` — the registry extraction's starting point (D-05-28)
- `.planning/spikes/005-native-ffi-provenance-cleanup/README.md` — iteration 4 is
  the direct template for D-05-05's fixture; iteration 5 (`longjmp` bypasses
  cleanup with **no** sanitizer report) is the load-bearing caution for D-05-11
- `.planning/spikes/005-native-ffi-provenance-cleanup/native/` and `lab/` — read
  the actual C and Go, not only the README
- `.planning/spikes/004-independent-certificate-checker/README.md` — the origin of
  D-12's independence rule

### Phase 4 carry-forward (read before planning)
- `.planning/phases/04-fallible-resources-and-c-boundary/04-CONTEXT.md` —
  D-04-01..D-04-29; **D-04-13 in full** (the attribute refusal D-05-01/D-05-03
  narrows) and **D-04-14** (the `_Noreturn` exemption whose boundary D-05-01 reuses)
- `.planning/phases/04-fallible-resources-and-c-boundary/04-DEBT.md` — **D-04-26
  in full** (the liveness-retirement precondition D-05-34 turns on), **D-04-32**
  (the alias-fact subject gap), **D-04-33** (the `Alias` pin), D-04-30, D-04-31
- `.planning/phases/04-fallible-resources-and-c-boundary/04-VERIFICATION.md` and
  `04-VALIDATION.md` — the truths this phase must not regress
- `.planning/phases/03-borrowed-views-and-cfg-lifetimes/03-DEBT.md` — the D-03-01
  history behind D-05-34, and the **D-03-02** entry D-05-33 must re-record

### Implementation surfaces named by decisions above
- `internal/compiler/cgen/cgen.go:157` — the by-value parameter emission that
  D-05-02 identifies as the blocking precondition
- `internal/compiler/session/session.go` — the lane table, required-control set,
  and the existing mutation runners D-05-08/D-05-09 must peer, not fork
- `internal/compiler/native/native.go` — Clang invocation, multi-TU compilation,
  and `validateExecution`
- `scripts/verify-phase4.sh` — the shape `verify-phase5.sh` copies (peer, not fork)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/compiler/session/session.go` — the lane table, `RecomputedWork`
  accounting, and the fail-closed required-control set with its verbatim
  duplication into the shell gate plus a set-equality test. Phase 5 adds
  `Phase5RequiredControls()` and `TestPhase5RequiredControlsMatchScript` on the
  same pattern. The existing marker-count guards (`>1` and `0` refuse to run
  rather than mutate an ambiguous target) are the pattern every new mutation
  runner must copy.
- `internal/compiler/session/session.go`'s `LayoutMutationRunner`,
  `OwnedBackendMutationRunner`, the release-omission and release-transposition
  runners, and the two nonlocal-pad runners — five worked examples of the
  "attack a specific artifact, fail closed, count nonzero work" shape.
- `internal/compiler/evidence/evidence.go` — content-bound manifests; gains the
  populated `emitted_attributes` (D-05-04) and the sanitizer classification tuple
  (D-05-16), both additive and `omitempty`.
- `TestOwnershipSequenceExhaustive` (~113,164 cases) and
  `TestBranchSequenceExhaustive` — the shadow-run oracle D-05-35 requires. Note
  they are **checker** oracles over core, not execution differentials; they never
  launch a process, which is exactly why D-05-18's execution corpus is separate
  work and not a corollary of them.
- `scripts/verify-phase4.sh` + `scripts/assert-go-tests.sh` — the bounded gate,
  the fail-closed exact-target selector, and the warm-observation helper. Phase 5
  **copies the shape into `verify-phase5.sh`**; it does not fork the frozen Phase
  4 script and does not extract a shared helper, per the duplication rationale
  already recorded there.
- `internal/compiler/native/native.go` — already resolves foreign sources by
  declared symbol (`ForeignSourcePathForSymbol`) and already compiles foreign TUs
  as separate invocations, so D-05-08's second allocator TU and D-05-09's hostile
  variant slot into existing machinery rather than needing new plumbing.

### Established Patterns
- Negative controls are exact IDs required fail-closed by the gate, each with
  nonzero counted work; expected escapes are `escape:` IDs that must appear and
  must never be claimed as covered.
- An additive emitter path leaves prior emitters and their goldens untouched
  (D-04-20's streaming events) — the precedent D-05-02's by-pointer lowering follows.
- Semantic identity uses function-local ordinals, never source offsets — which is
  why D-05-23 reduces core structure rather than source text.
- A host-tool-dependent control returns `tool_missing`/operational, never pass
  (D-04-19's `nm -u`) — the posture D-05-15 reuses for the sanitizer runtime.
- Evidence claims content identity only; it never overclaims translation proof.

### Integration Points
- `internal/compiler/native/native.go` `Run` — gains a third and fourth build
  configuration (`-O3 -flto` for D-05-19, and the separate sanitizer binary for
  D-05-11). The sanitizer build must be a distinct artifact with a distinct
  output path, never a flag toggled on the differential's own build.
- `native.validateExecution` — Phase 4 widened it for typed failure, defect-abort,
  and nonlocal exit. D-05-20 widens the *comparison* rather than the validator:
  exit status and signal become compared data, and D-05-21 adds the fail-closed
  field-routing test that makes future widening impossible to do silently.
- `internal/compiler/cgen/cgen.go` — the by-value parameter emission at line 157
  is the blocking precondition (D-05-02); `unsafeForeignContractField` and its
  peers gain `Alias` (D-05-36).
- `internal/compiler/corevalidate/corevalidate.go` — gains the independent
  attribute-justification re-derivation (D-05-04) and `Alias` in
  `foreignContractFieldsCSafe` (D-05-36).
- `internal/compiler/check/check.go` — `loanLivenessFixpoint` broadens to the full
  straight-line-plus-branch domain and `discoverLoanLastUses` is deleted only
  after a zero-divergence shadow pass (D-05-35).
- Any new lane joins the required-control set and must carry nonzero work; any new
  diagnostic joins the `lang.diagnostic/1` taxonomy without moving an existing ID.

</code_context>

<specifics>
## Specific Ideas

**The LTO finding is the sharpest result in the whole research set.** Without
`-flto`, code in a separate translation unit is opaque to link-time optimization,
and every current corpus program's only observable effect crosses exactly that
boundary. A green `-O0`/`-O3` differential over today's corpus is therefore
proving codegen determinism, not equivalence under optimization. If Phase 5 ships
without D-05-19's LTO tier and D-05-18(a)'s engineered adversarial subset, SC1 is
satisfiable while meaning nothing — which is the same species of empty claim
D-04-15 refused to accept for panic.

**Three of NAT-03's seven mutations had no subject; two remain partly
unsubjected after this phase, and both are named.** "False no-alias facts" gets a
real subject only after the by-pointer lowering (D-05-02) lands — that lowering,
not the attribute, is the actual deliverable and the actual risk. "Stale callback
retention" is subsumed into the retained-pointer lane with
`escape:callback-invocation-unsubjected` recording exactly what is not tested.
Honest arithmetic for NAT-03 is six subjected plus one subsumed-with-escape, and
the phase should say so rather than presenting 7/7.

**Every new lane in this phase needs an always-on positive control, not a
one-time authoring check.** The sanitizer lane, the reducer, and the LTO tier are
all new machinery whose failure mode is silent inertness — the sanitizer not
actually linked, the reducer no-opping, the LTO tier building without `-flto`.
D-05-14's always-on defective fixture is the general answer and should be applied
to all three, not only to ASan.

**`alloc_dealloc_mismatch` defaults to OFF on macOS.** This project's primary
host is Apple clang on arm64. A contributor running the suite with default
`ASAN_OPTIONS` gets a false green on D-05-08's control with no warning at all.
That single default is the most likely concrete false-green in the phase and is
why D-05-13 pins options in the harness rather than the environment.

**Do not delete `discoverLoanLastUses` on a green suite.** Rust ran both borrow
checkers in parallel for years and deleted the old one on divergence data, not on
a schedule. The ~113,164-case enumeration already exists; use it as a shadow-run
divergence log, not merely as a regression assertion that happens to still pass.

**Layered gates keep paying for themselves.** Phase 2's close-out and Phase 3's
three gap-closure rounds each showed a layer finding defects the previous layer
missed, including one defect *introduced* by the previous layer's fix. Keep
review, verification, validation, and security as distinct passes, and adopt
D-05-40's mid-phase gate in advance rather than after a wave failure.

</specifics>

<deferred>
## Deferred Ideas

- **Lang-to-Lang calls (`OpCall`), interprocedural loan liveness in both
  admission layers, call-graph construction, cycle refusal, a bounded interpreter
  call stack, and the cross-function rebuild of Phase 3's exhaustive
  differentials** — **M002's lead charter item** (D-05-32/D-05-33), gated on
  callable ⊆ publishable (D-04-03). Not Phase 6: Phase 6 is agent feedback and
  performance ratification, not language surface.
- **Interprocedural `-O3` equivalence** — out of M001's proof scope by
  construction, since M001 ships without calls. Must be stated explicitly in the
  roadmap and debt register rather than implied by SC1's wording.
- **A real foreign callback-registration seam** — generation tokens, registration
  handles, unregister races, `FFI-004`/`FFI-005`'s thread-and-timing variants
  (D-05-06). Additive whenever a milestone needs calls-into-Lang.
- **LeakSanitizer, MemorySanitizer, ThreadSanitizer** — LSan is unreliable on the
  primary host (D-05-13 pins `detect_leaks=0`); MSan and TSan have no subject in a
  single-threaded language with no uninitialized-read surface.
- **UBSan `implicit-conversion` and `unsigned-integer-overflow`** — not UB,
  known-noisy; named as an escape rather than silently dropped (D-05-13).
- **`-O3`-level sanitizer runs** — the lane builds at `-O1` (D-05-12); an
  `-O3` sanitizer tier is a named deferred obligation if `-O3`-specific UB later
  matters.
- **The wiki's 64-workload expansion, the 24 blocking probes, and the extended
  semantic matrix** — a superset design for a larger language (D-05-18).
- **A storable, matchable `Result` value and payload-carrying alternatives** —
  unchanged from D-04-30; M002 or a Phase 6 error-taxonomy requirement.
- **Full one-to-one code ports of superseded spike fixtures** — waived rows with
  citations instead (D-05-31).
- **Loop-carried loan liveness, generics, closures, threads, async** — unchanged;
  this phase adds no loop, recursion, or call construct.

### Accepted residual limitations (record, do not engineer around)

- **The callback-invocation mechanism is unsubjected.** The pointer-lifetime root
  cause is tested; registration-then-later-firing is not, because M001 has no
  calls-into-Lang. Recorded as `escape:callback-invocation-unsubjected` (D-05-07).
- **The coordinated multi-artifact lie survives, one layer wider.** Phase 5 makes
  it *demonstrable* (D-05-30) rather than merely asserted, which is strictly
  better evidence — but demonstrating an escape is not closing it, and no control
  in this repository closes a coordinated lie.
- **Optimizer observability is host- and version-bound.** D-05-05's divergence was
  measured on Apple clang 21 / arm64. The compiler version is recorded in
  evidence and a lost divergence fails the lane, but the *shape* of the fixture
  that provokes `-O3` remains an engineered artifact of this toolchain, not a
  portable law.
- **A differential is existential, not universal.** It proves these engines agree
  on these programs on these axes under these opt levels — nothing more. The
  enumeration bound (D-05-18) and the exclusion list (D-05-21) are the honest
  statement of that scope, which is why both are asserted in code rather than
  described in prose.
- **Quarantine of the foreign boundary is permanent and non-discharging**
  (D-04-31 item 4, inherited unchanged). Sanitizer evidence catches defects the
  sanitizer can see; spike 005 iteration 5 already recorded a real cleanup bypass
  that produced **no** sanitizer report at all.
- **`discoverLoanLastUses`'s retirement is the last M001 opportunity.** If it
  slips again it survives the milestone; D-05-34 records that explicitly so the
  next deferral is a visible decision rather than a quiet one.

</deferred>

---

*Phase: 5-Native Equivalence and Adversarial Evidence*
*Context gathered: 2026-09-05*
