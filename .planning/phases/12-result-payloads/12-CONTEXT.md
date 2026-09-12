# Phase 12: `Result` Payloads - Context

**Gathered:** 2026-09-12
**Status:** Ready for planning — **criterion 3's pre-flight probe is ANSWERED: BRANCH A (accepted), 2026-09-12. See D-12-04a.**

<domain>
## Phase Boundary

An **alternative can carry a payload** — constructed, stored, matched with a
binder, and moved out of under the affine discipline — and that payload has
**one meaning** across the core IR, `interp`, and emitted C17, proven on the
five-axis comparator rather than asserted per engine.

This is the first phase of M002 whose subject is the **language surface** rather
than the interprocedural machinery. Phases 07-11 made multi-function programs
admissible and then executable; this phase makes a *value* richer. D-04-30's own
recorded warning governs the cost estimate: "a storable `Result` value bolted
onto an edge-based core is a genuine re-lowering of the failure representation,
not an extension of it" — the cost is explicitly **not** incremental.

**Where the tree actually starts** (verified against the shipped source
2026-09-12, not inferred from the roadmap):

- `core.DataType` is `{ID, Name, Alternatives []string, Span}`
  (`internal/compiler/core/core.go:22-27`) — bare nullary tags, no payload field.
- `core.MatchArm` is `{ID, EdgeID, Pattern string, Value string, BlockID string
  omitempty}`. **There is no binder anywhere in the pattern representation.**
  `Pattern` is one alternative name.
- The parser reads **exactly one identifier then `=>`**
  (`internal/compiler/syntax/parser.go:530`), so a payload binder is genuine
  new parser work. `maxArmsPerMatch` and `maxAlternatives` (4096) already cap
  the surface, fail-closed.
- `ast.MatchArm` is `{Pattern, Value, Body, Span}` where `Value` is the arm's
  **bare-name result expression**, mutually exclusive with `Body`
  (`internal/compiler/ast/ast.go:111-120`). `ast.Alternative` is `{Name, Span}`.
- `interp`'s entire value model is `values map[string]string`
  (`internal/compiler/interp/interp.go:308`). The foreign err edge stores the
  literal string `"err"` (`:667`).
- Phase 4's failure is a **two-successor control-flow edge**, not a value
  (D-04-04). No generics exist (GEN-01 is M003), so `Result<T, E>` is not
  expressible.
- Nothing named `AlternativeDetails`, `PayloadType`, or `PayloadTargetID`
  exists anywhere in the tree. This area is greenfield.

**In scope:**

1. Payload-carrying alternatives in the source grammar, the core IR, `check`,
   `corevalidate`, `interp`, `cgen`, `pathoracle`, and `originvalidate`
   (RES-02, RES-03).
2. Criterion 3's **pre-flight probe**, run and answered **before planning**
   (D-12-01..D-12-04).
3. The inherited **D-11-02 six-emitter deletion** and the **D-11-52** Match-body
   lift, as explicitly-labeled inherited-NAT-debt-closure work (D-12-28..D-12-33).
4. Criterion 2's engineered anti-vacuity control (D-12-34..D-12-38).

**Not in scope** (each with a named home):

- **Resource-carrying payloads** — refused this phase by a named fail-closed
  diagnostic; the full rule is declared debt with a three-part landing condition
  (D-12-22..D-12-27).
- **Niche optimization** — designed and recorded, not built, because the
  precondition cannot presently exist (D-12-19..D-12-21).
- Generics / a real `Result<T, E>` (GEN-01, M003).
- `lang.foreign/0` widening (D-11-07, still OPEN and unowned).
- D-11-51's per-invocation event identity — **flagged as likely to be tripped
  first by this phase's fixtures**, but its fix needs its own reviewed plan
  (D-12-17).
- QLT-09's Phase-5 Nyquist portion — deliberately **not** closed mid-phase
  (D-12-39).
- Arithmetic, iteration, `if`, strings, arrays (on no roadmap).

</domain>

<decisions>
## Implementation Decisions

All eight gray areas were researched in advisor mode (eight parallel
`gsd-advisor-researcher` agents, `minimal_decisive` calibration, advisor model
`sonnet`) under the developer's standing mandate to fan out across
stakeholder-role lenses, run an adversarial pass, draw on cross-ecosystem prior
art, and synthesize one-shot recommendations. The developer adopted the
synthesized recommendation in **every** case, after being shown the three areas
(2, 6, 7) where adoption carries a cost worth stating.

**Every load-bearing factual claim below was independently re-verified against
the shipped tree by the orchestrator before being locked.** Where a researcher
and the tree disagreed, **the tree won** and the correction is recorded inline
at the decision it changes (D-12-13, D-12-36, and the line-number ledger in
D-12-29). This phase also **deviates from a recorded design plan** (D-04-30) and
that deviation is stated as a deviation, with the superseded text named
(D-12-05).

### Criterion 3 — the pre-flight probe (hard entry gate)

- **D-12-01 (form: a committed single-branch Go test, Q-01 shape — not a
  throwaway spike):** criterion 3's question is **binary** ("does payload origin
  and ownership reduce to already-proven Phase 08-11 machinery, yes or no"),
  which is structurally identical to Phase 11's Q-01 ("does `corevalidate`
  accept this core-level rewrite") and structurally *unlike* Phase 09's S-008
  replacement (a **gradational** threshold question over a count). The
  gradational form does not apply, so the threshold-inventory shape is wrong
  here. Land the probe as a committed Go test plus a companion clean-edit
  assertion, run **before planning**, left permanently in the suite.
  Rationale is this project's own discipline rather than precedent-worship: a
  `.planning/spikes/` prose verdict leaves **no standing regression**, so if
  `peerDeriveOriginFacts` or the callee-contract map changes later, nothing
  re-fires the check that made the phase-entry gate decision — the gate's
  evidentiary basis would live in an editor's memory of a document, which is
  exactly what mechanism-over-convention forbids.
  — **Reversibility:** reversible — a committed test can be deleted; nothing
  else depends on its existence.

- **D-12-02 (what the probe executes, and its oracle):** one fixture built at
  the **darkest corner**, not a representative average: a callee returns a
  payload-carrying value whose payload is sourced from a `PublicOrigin`
  forwarded across `OpCall`; the caller matches it and moves the payload out of
  the matched arm. The oracle is the **four admission peers** (`check`,
  `corevalidate`, `originvalidate`, `pathoracle`) — independent by per-package
  import allowlist, so none of them is its own oracle and none reuses the
  mechanism under test.

- **D-12-03 (the injected defect, which is what makes the oracle non-vacuous):**
  inject at D-10-C01's boundary and verify the peers **diverge in the
  already-known, already-attributed way** — `corevalidate`'s
  `peerDeriveOriginFacts` refusing the shape as not-`Callable` while `check`
  admits it with zero diagnostics. A probe whose peers merely agree proves
  nothing; the pre-registered divergence proves the oracle can see.
  **Verified in tree 2026-09-12:** `peerDeriveOriginFacts`
  (`internal/compiler/corevalidate/corevalidate.go:2407`) has cases for
  `core.OpBorrowShared`, `core.OpBorrowExclusive`, and `core.OpMove`/`core.OpCopy`
  — and **no `case core.OpCall`**. D-10-C01 is still open; the probe's premise
  holds.

- **D-12-04 (BRANCH A / BRANCH B pre-registered, and the M003 slip trigger
  stated so it cannot be argued away after the fact):** write both branches down
  **before** running.
  - **BRANCH A (accepted, phase proceeds):** every peer's verdict on the fixture
    is **fully attributable** to an already-catalogued carry-forward gap —
    D-10-C01, D-11-51, or D-11-52 — with **no new divergence root cause**.
  - **BRANCH B (slip to M003):** fires when **either** (a) any peer's verdict
    traces to a divergence root cause **not** already named in
    D-10-C01/D-11-51/D-11-52, **or** (b) closing the divergence requires adding
    a **new dispatch-site case arm of a kind not already inventoried** across
    the six Phase 08-11 dispatch sites.
  The (b) clause carries its own explicit exclusion, because the distinction is
  the whole gate: **filling an existing empty switch arm is in-scope
  maintenance, not a new interprocedural rule.** Phase 12 is already named as
  the roadmap's **first scope cut**, so BRANCH B's response is an explicit slip,
  never extra plans.
  — **Reversibility:** one-way in spirit — once BRANCH B fires, the phase's
  scope decision is made by the pre-registration, not renegotiated; the point of
  pre-registering is that the verdict cannot be rationalized afterward.

- **D-12-04a (PROBE RUN — VERDICT: BRANCH A (accepted). Phase 12 proceeds.):**
  run 2026-09-12, **after** D-12-01..D-12-04 were committed (`97a7c03`) and
  **before** any planning — so the pre-registration is timestamped in git ahead
  of the result and the verdict cannot be back-fitted.
  **Artifact:**
  `internal/compiler/corevalidate/corevalidate_result_payload_probe_test.go` —
  four tests, committed and permanent:
  `TestC03ResultPayloadOriginAcrossOpCall` (the probe, two subtests),
  `TestC03PeerDeriveOriginFactsOpCallGapStillOpen` (pins the gap the attribution
  depends on), `TestC03ProbeFixturesAreUnmodified` (companion clean-edit
  assertion). Full suite, `go vet`, and `gofmt` all green.

  **How the two reduction shapes were chosen.** Payload-carrying alternatives do
  not exist yet, so no real payload fixture can be run. The probe drives the two
  shapes the payload question *reduces* to, both already expressible:
  `testdata/phase07/call_from_both_match_arms.lang` (the match-arm-with-a-call
  half — whose own header already names it **"Phase 12's forward guard, since a
  future `Result` match arm's own call must be picked up the same way"**, i.e. it
  was authored for this phase), and
  `testdata/phase08/relay_depth2_accept.lang` (the forwarded-callee-result half).
  All four peers were driven **directly**, never through `session`, so no peer is
  consulted through another peer: `check` via `loadCheckedProgram` (which fails on
  any diagnostic, so a clean load *is* check's verdict), `corevalidate.Validate`,
  `originvalidate.ValidatePublished`, and `pathoracle.BuildCalleeLookup` +
  `RecomputeEndpoints`.

  **Result, shape by shape:**
  - **Match-arm-call shape:** all four peers clean — `check` zero diagnostics,
    `corevalidate.Valid == true`, zero `originvalidate` problems, no `pathoracle`
    error. No divergence at all.
  - **Forwarded-callee-result shape:** all four peers clean. Note this is itself a
    **change from the fixture's recorded state** — see D-12-04c.
  - **Synthesized D-10-C01 shape** (the darkest corner, built because it is
    *inexpressible* in committed testdata — see D-12-04b): refused by two peers
    with **two already-catalogued codes**:
    - `corevalidate` → **`core.callee_not_callable`** (`core.CalleeNotCallable`,
      `core/core.go:587`) — D-10-C01's signature refusal, and D-10-C04's own flip
      target.
    - `originvalidate` → **`core.origin_understated`**
      (`originvalidate/originvalidate.go:473`) with the detail *"declared origin
      [buffer] does not cover the body-derived origin **[]**"*. The empty
      body-derived origin is D-10-C01's mechanism **made directly visible**:
      `peerDeriveOriginFacts` cannot walk through the `OpCall`, so it derives
      *nothing* and the declared origin is reported as uncovered.
    Both codes were verified **pre-existing** in the tree, not introduced by this
    probe.

  **Why this is BRANCH A and not BRANCH B.** Every verdict is fully attributable
  to **D-10-C01**, an already-catalogued carry-forward item. No refusal traces to
  an uncatalogued root cause (clause (a) not triggered), and closing either
  refusal means **filling `peerDeriveOriginFacts`'s existing empty switch with a
  `case core.OpCall:` arm** — which D-12-04's clause (b) explicitly excludes as
  *"in-scope maintenance, not a new interprocedural rule"*. No new
  `OperationKind` and no new dispatch site is implied. The refusals are
  **fail-closed and conservative, not unsound**, which is exactly why D-12-27 and
  D-12-28 defer resource-carrying payloads rather than this phase slipping.

  **Two findings the probe produced that plans must carry forward:**
  1. The payload-forwarding fixture shape is **blocked**, confirmed empirically
     rather than inferred — any Phase 12 fixture needing a forwarded payload
     origin hits `core.callee_not_callable` / `core.origin_understated`. This is
     D-12-28's premise, now measured.
  2. `peerDivergenceExpected`
     (`internal/compiler/session/session_peer_gate_test.go:43`) has exactly **one
     live entry** left (`testdata/phase07/duplicate_function_name.lang` →
     `core.duplicate_function_id`), and its own comment records that D-09-51's
     retirement **"unmasked a SEPARATE, pre-existing `originvalidate` finding —
     one this test is structurally blind to (it never calls
     `originvalidate.ValidatePublished`)"**. The probe **does** call it, so it
     sees what that gate cannot. Plans should not treat the peer-gate register as
     a complete divergence inventory.

- **D-12-04b (the darkest-corner fixture is INEXPRESSIBLE in testdata, and that
  is part of the finding):** no committed fixture declares a borrow-returning
  `PublicOrigin` sourced from forwarding a callee's result — **because the gap is
  open**. D-10-C01's own text says it "constrained which fixtures Phase 10 plans
  10-07/10-08 could express". So the probe **synthesizes** the shape at core
  level, Q-01 style: clone `relay_depth2_accept.lang`'s checked program (where
  `relay` forwards `leaf`'s result and declares no origin at all) and attach
  `&core.PublicOrigin{Paths: []string{"buffer"}, Access: "shared"}` to `relay`.
  A **clean baseline is asserted first** (both peers admit the unmodified
  fixture), and `assertOnlyPublicOriginChanged` proves the edit is a single
  additive field — function count, every function ID/Name/ReturnType, every
  operation count, and every operation ID **and Kind** unchanged. That is what
  makes the verdict attributable to the added declared origin alone.

- **D-12-04c (TREE CORRECTION — `relay_depth2_accept.lang`'s own header is
  stale):** the fixture's header states that "corevalidate ... also refuses this
  fixture (`core.move_while_borrowed`) even though `check`'s own interprocedural
  law ... correctly admits it". **Verified no longer true**: the probe measured
  `corevalidate.Valid == true` on the unmodified fixture. Phase 09's **D-09-03**
  closed it deliberately — retiring `relay_depth2_accept.lang` and
  `twin_a_accept.lang` from `peerDivergenceExpected` was a *required assertion* of
  Phase 09 — and the register's retirement comment says so, but the fixture's own
  header was never updated. The probe's first draft pinned D-10-C01 against this
  fixture and went red for exactly this reason; the fixture was not wrong, the
  header was. Recorded per D-09-45 rather than silently fixed. **A plan touching
  this fixture should correct its header.**

### Core IR payload representation

- **D-12-05 (TWO NEW `OperationKind`s — an explicit DEVIATION from D-04-30's own
  named design, stated as a deviation):** payload construction and payload
  destructuring each get a **new `core.OperationKind`**. This **supersedes**
  D-04-30's recorded plan, which the ROADMAP lists as a canonical ref and which
  reads: "add a sibling `alternative_details []Alternative omitempty` field keyed
  by name — additive to `core.DataType`, never a shape change to the existing
  `Alternatives []string` field", extended by ARCHITECTURE §7's [INFER] to "a
  `PayloadSourceID`/`PayloadTargetID` pair, following the exact precedent
  `OkEdgeID`/`ErrEdgeID`/`ErrTargetID` already set for `OpForeignCall`".
  **Why the recorded plan is rejected:** riding existing kinds makes the
  *meaning* of `OpCopy`/`OpMove` conditional on whether an omitempty field is
  populated, so `TestAllOperationKindsHandledAtEverySite`
  (`control:kind.exhaustive_dispatch`, `internal/compiler/core/core_test.go:267`)
  **passes vacuously** for the new payload semantics: the switch still covers
  every kind while the new affine behavior hides inside an arm the control never
  forces anyone to re-examine. That is precisely the "guarantee degrades to a
  comment a later editor can silently violate" failure mode this project's
  mechanism-over-convention discipline exists to prevent.
  **Why the cited precedent does not actually support the recorded plan:**
  `OkEdgeID`/`ErrEdgeID`/`ErrTargetID` live on `OpForeignCall`, a kind whose
  **single** meaning already *is* "a call with two outcome edges". Those fields
  refine one meaning; they do not bolt a second, silently-different meaning onto
  a general-purpose kind.
  **Cross-ecosystem prior art, both pointing the same way:** Rust MIR keeps
  `AggregateKind::Adt` out of `Rvalue::Use` rather than folding variant
  construction into an ordinary use; Swift SIL keeps `enum` and
  `unchecked_enum_data` as **distinct, verifier-checked** instructions. Both
  drew the line for the same reason — a sum-type variant with a payload is a new
  *kind* of value-producing/consuming event, not a shape variant of a scalar
  move.
  Accepted cost, counted not estimated: **~16-20 independent edits** across the
  six literal dispatch sites (D-04-22 prices one kind at a minimum of 8-10),
  including the body-blind `originvalidate` and the independence-critical
  `pathoracle`.
  — **Reversibility:** costly — undoing means removing two kinds from
  `AllOperationKinds()` and unwinding the arms at all six dispatch sites plus
  the exhaustive-dispatch control's own fixtures.

- **D-12-06 (argue from affine obligations, not edit count — the reason the
  hybrid was also rejected):** destructuring **mints a fresh place carrying a
  fresh ownership obligation** that no existing kind models, so it must get a new
  kind under every candidate considered. Once that is conceded, giving
  construction the same treatment costs one more kind's 8-10 edits and removes
  the conditional-meaning hazard on `OpCopy`/`OpMove` **entirely**. D-04-22's
  cost is stated as a **floor precisely to buy the anti-vacuity guarantee** — the
  price is the feature, not the obstacle.

- **D-12-07 (`PayloadType` is a bare string):** match the existing all-string
  type-naming convention already used by `core.Function.ReturnType`,
  `core.Parameter.Type`, and `core.MatchArm.Pattern`. A type-**ID** reference
  would be the only ID-based type reference anywhere in the IR — an inconsistent
  parallel mechanism in a language with no generics to disambiguate.

- **D-12-08 (no `lang.core/2` bump):** `AllOperationKinds()` gaining two entries
  and `core.DataType` gaining an additive struct do **not** move the serialized
  bytes of any fixture that does not use them. `Schema = "lang.core/0"` and
  `Schema1 = "lang.core/1"` (`core.go:10-11`) both hold. This follows D-11-06's
  own reasoning, which refused an `Exported`/`EntryFunctionID` field on
  `core.Program` because "any populated field (even `omitempty`) moves frozen
  core bytes and forces a `lang.core/2` bump against D-05-39" — the difference
  is that a **new kind string** is only ever present on an operation that uses
  it, whereas a new struct field on an existing type is serialized corpus-wide.
  — **Reversibility:** one-way if violated — a schema bump is a published
  contract change; the plan must assert byte-identity over the corpus, not
  assume it.

- **D-12-09 (the parallel-array hazard gets a mechanism, not a comment):**
  `Alternatives []string` alongside a payload-detail list keyed by name is a
  genuine parallel-array desynchronization hazard — a name present in one and
  absent from the other is **representable but invalid**. Close it two ways:
  (i) a **single constructor function** that is the only permitted way to build a
  `core.DataType` with payload-carrying alternatives, enforcing name-set equality
  at construction; **and** (ii) one `corevalidate` arm asserting the same
  invariant **over the whole corpus**. This mirrors how Protobuf's `oneof` and
  FlatBuffers' union-type-vector pairing are kept in sync by generated accessor
  code rather than by convention.

- **D-12-10 (one minting authority for the destructured place):** the
  destructured payload's place ID is minted by the **same lowering pass that
  already mints every other place, edge, and block ID** in this compiler — never
  a second minting authority introduced by a validator or by `interp`. Name the
  field `PayloadTargetID`, following the `OkEdgeID`/`ErrTargetID` field-naming
  precedent exactly.

### Source syntax and the pattern binder

- **D-12-11 (`| Ok(Buffer)` declaring, `Ok(v) => { ... }` binding):**
  constructor-application syntax, the spelling used with only cosmetic variance
  by Rust, Swift, OCaml, Scala, Kotlin, and (via `|payload|` capture) Zig.
  Decided on **least surprise for both audiences**: this language's primary
  *author* is an AI agent and its primary *reviewer* is a human, so the syntax
  both have already internalized beats a novel or terser one. No researched
  ecosystem — Rust, Swift, OCaml, Zig, Erlang, Austral, Koka, Val/Hylo — spells
  payload extraction any other way, so the alternatives have **no
  least-surprise precedent to lean on**.

- **D-12-12 (general `data` feature; `Result` is a USER FIXTURE, not a
  compiler-blessed built-in):** any user-declared `data` type's alternatives may
  carry payloads. `data Outcome = | Ok(Buffer) | Err(Fault)` is the feature;
  RES-02 is satisfied by **a real type**, with one canonical fixture named
  `Result`, rather than by special-casing the requirement's wording.
  Three reasons, in order of weight: (i) a blessed monomorphic built-in is
  exactly the ambient, special-cased authority this project's no-ambient-authority
  constitution argues against; (ii) the general form is what today's
  `data`/`match` machinery **already gives away for free** — alternatives are
  already a list, `match` already dispatches on tag; (iii) it is the form GEN-01
  (M003, "ability-bounded generics, enabling a real generic `Result<T, E>`") can
  absorb **without a breaking migration**, whereas a built-in is something the
  generics work would have to unbuild.
  — **Reversibility:** one-way — the spelling is a published source-language
  contract; every fixture, golden file, and `lang format` output written under it
  would have to be rewritten to change it.

- **D-12-13 (TREE CORRECTION — the binder is genuine new parser work):** the
  researcher's synthesis claimed this "extends today's 'one bound place per arm'
  (`MatchArm.Value`)". **Verified wrong.** `ast.MatchArm.Value`
  (`internal/compiler/ast/ast.go:111-120`) is the arm's **bare-name result
  expression**, mutually exclusive with `Body` per its own doc comment
  ("Exactly one of Value and Body is populated") — it is **not** a binder slot.
  The parser reads `pattern := p.identifier("syntax.expected_pattern")` followed
  immediately by `=>` (`internal/compiler/syntax/parser.go:530-533`), so `Ok(v)`
  requires inserting an optional `( identifier )` after the pattern identifier,
  with new named refusals for the malformed forms. ARCHITECTURE §7's phrase
  "`checkBranch`'s existing bare-ADT match already binds one place per arm today"
  refers to `checkBranch`'s **internal** place handling
  (`internal/compiler/check/check.go:1597`), not to any AST-level binder.
  Recorded rather than silently fixed, per this project's stale-reference
  discipline (D-09-45). Plans must budget real parser work here.

- **D-12-14 (bind MOVES — via a real `OpMove`, and the diagnostics must be
  loud):** binding a payload emits a real move at arm entry, from the
  scrutinee's payload slot into a fresh, independently-tracked place; moving that
  place further still requires the existing explicit `take`. This matches
  ARCHITECTURE §7's requirement that matching "must produce a fresh,
  independently-tracked place for the extracted payload".
  **The researcher's own flagged risk is adopted as a constraint, not noted and
  dropped:** move-on-bind *looks* like naming but *is* consuming — that is
  Rust RFC 2005's exact match-ergonomics surprise, where a move/non-move
  distinction became invisible in the pattern. The diagnostics must be loud
  enough that an AI author and a human reviewer both see the consumption.

- **D-12-15 (three new named fail-closed refusals):** with payloads, a pattern
  can be wrong in ways set-membership exhaustiveness cannot express. Add
  arity-mismatch, binder-on-nullary-alternative, and
  missing-binder-for-payload-alternative as named refusals in the `check.*`
  namespace, layered **on top of** the existing set-membership exhaustiveness
  check rather than replacing it.

- **D-12-16 (rejected: bare tag plus body-level `take`):** recorded so it is not
  relitigated. `Ok => { let v = take flag }` keeps `MatchArm.Pattern` a bare
  string and is the smallest grammar diff, but it needs **flow-sensitive
  narrowing the checker does not have** (inside the `Ok` arm, `flag`'s type must
  be understood as "the Ok alternative, payload accessible", with no name in the
  pattern to hang the fact on) — more type-system machinery to build than a
  binder field, not less. It also overloads `take`'s established meaning for a
  subtly different job, and a new narrowing pass is exactly the soundness-gap
  class the never-guess discipline exists to prevent.

### Interpreter value model

- **D-12-17 (widen to a struct: `value{tag, payload string}`,
  `values map[string]value`):** the Go compiler then forces **every** access site
  to be touched, with no silent misses; a zero-value struct (`tag == ""`) is
  trivially "this is a scalar" with **no nil-interface state** to mishandle; and
  it stays flat and allocation-cheap, matching the package's existing
  plain-string idiom. An interface/sum-type encoding was rejected: it buys
  extensibility for a third value shape that does not exist (there are exactly
  two — scalar, and tagged-with-payload), while adding a type switch per arm and
  a nil-interface footgun the struct form **cannot even express**.
  Blast radius, counted: ~17 sites in `internal/compiler/interp/interp.go` — 10
  `.values[...]` accesses, the three literal `map[string]string{...}` frame seeds
  at `:205`, `:219`, `:286`, the three frame constructors `newFlatFrame` (`:357`),
  `newBlockFrame` (`:374`), `newArmFrame` (`:427`), and the struct field at
  `:308` — plus the caller-argument deletion logic at `:452-470`.
  — **Reversibility:** costly — undoing means narrowing the type back across all
  ~17 sites.

- **D-12-18 (evidence-invisible BY CONSTRUCTION, not by review):** every
  non-payload site writes `tag: ""`, and serialization **special-cases
  `tag == ""` to emit the bare string**, so a scalar value's serialized bytes are
  byte-identical to today's. This is the same reasoning as D-11-01: Phase 11
  chose an additive emitter path specifically so pinned digests could not move in
  the diff that introduced calls, because "when a digest moves you cannot tell
  which change did it. The additive path freezes them **by construction** rather
  than by test." The four pinned golden-C digests
  (`internal/compiler/core/core_test.go:156-159`:
  `testdata/phase1/generated.golden.c`,
  `testdata/phase2/owned_transfer.golden.c`,
  `testdata/phase4/foreign_layout_mismatch.golden.c`,
  `testdata/phase5/restrict_borrow.golden.c`) must not move for this reason.
  — **Reversibility:** one-way if violated — once frozen evidence bytes move,
  Phase 11's equivalence evidence was produced under a different representation
  and cannot be re-attributed.

- **D-12-19 (the proof is a corpus characterization replay):** replay the
  **entire** pre-Phase-12 fixture corpus and assert **byte-identical execution
  documents**. This is the mechanism that actually discriminates a real
  regression from a no-op, because it exercises the real oracle/native comparator
  path rather than a hand-picked property. Add a narrow property test that a
  scalar value survives an arbitrary `copy`/`move`/`borrow` sequence unchanged.

- **D-12-20 (two representations REJECTED as this project's own named
  anti-patterns):** (i) **string-encoding** tag and payload into one string is
  **in-band signalling** — the literal `"err"` already shares a namespace with
  user values, a payload containing the delimiter is a correctness bug, and the
  interpreter's value space stops being checker-derivable; (ii) **parallel side
  maps** (`values` plus `tags` plus `payloads`) is **parallel-map
  desynchronization** — a place present in one map and absent from another is
  representable-but-invalid, and frame construction/teardown plus the
  caller-argument deletion logic would have to keep N maps in lockstep across
  call boundaries. Neither is a narrow diff; both defer cost into a shape the
  project rejects on principle.

- **D-12-21 (D-11-51 is FLAGGED, not silently absorbed):** widening the value
  type is orthogonal to D-11-51 — event-ID identity derives from the callee's
  static `OpReturn`, untouched by the value shape. **But** a multi-call-site
  fixture exercising a payload-carrying return is exactly the kind of new fixture
  likely to trip that open defect **for the first time**. Planning must flag the
  interaction explicitly so it surfaces as an anticipated constraint rather than
  a mystery `validateExecution` refusal ("duplicate execution event id"). D-11-51
  remains **OPEN and UNOWNED**; its fix needs `interp.go` **and**
  `cgen_program.go` and therefore its own reviewed plan — Phase 12 does not claim
  it.

### C17 layout and the niche-optimization clause

- **D-12-22 (FLAT per-alternative-field struct, not a real C `union`):** emit a
  struct with a tag plus **one field per alternative**, the tag saying which is
  live. Reasons in order of weight: (i) a real `union` would be this project's
  **first non-checker-derived ABI claim**, since nothing in the current pipeline
  tracks which member is active — directly against the `restrict` discipline's own
  prohibition on asserting a property the checker did not derive; (ii) reading an
  inactive member is **UB in C17** outside the common-initial-sequence exception
  (§6.5.2.3), and `-O3 -flto` with TBAA plus sanitizers under `-Werror` is
  **exactly** the proof regime this project runs — a `union` is adversarial to
  the project's own equivalence claim; (iii) every byte's provenance stays
  checker-derivable with **zero new machinery**, reusing `core.RecordLayout` /
  `core.LayoutField` exactly as `standardForeignLayout`
  (`internal/compiler/check/check.go:3118`) already generates for the `{ok, value}`
  ABI struct.
  **The counter-argument is acknowledged and priced, not waved away:** a flat
  struct is not literally a "tagged union" as RES-03's words say, and wasted
  space cuts against the core value's refusal to hide runtime costs. But the only
  payload types available today are `Byte` and `Buffer`, so the waste is **one
  tag byte over two possible payload types** — measured and small, not
  hypothetical.
  — **Reversibility:** costly — the emitted C struct shape is what golden files
  and `_Static_assert` conformance pairs are written against.

- **D-12-23 (NICHE OPTIMIZATION IS VACUOUSLY SATISFIABLE — and that is the
  finding, stated loudly rather than glossed):** niche optimization requires a
  provably **uninhabited** bit pattern in the payload type to hide the tag in.
  The only payload types at this maturity are `Byte` and `Buffer`, and `Byte` is
  plausibly **fully inhabited** — every bit pattern of an `unsigned char` is a
  valid `Byte`. **There is therefore no niche to exploit**, and RES-03's niche
  clause is not merely undischarged, it is **uninstantiable**.
  This is the same *shape* of finding as D-11-09, where Phase 11 emitted zero
  call-boundary alias attributes because "two pointers to one object cannot exist
  across a Lang call boundary in M002" — a mechanism whose precondition cannot
  presently exist, so the zero state became the phase's **terminal** state rather
  than a waypoint. Criterion 2 must say so **explicitly**, as an
  absence-of-applicable-input finding; a silent "no niche fired" reading must not
  be allowed to stand in for "niche-restriction logic was verified", which are
  very different verification states.

- **D-12-24 (niche is DESIGNED-AND-RECORDED, not built):** record in the phase
  debt register, in the mechanically-checked `*-DEBT.md` format
  (`TestDebtRegistersAreWellFormed`, `internal/compiler/session/session_test.go:2613`):
  the candidate mechanism is an **extension of the existing checker-derived
  layout path** (`RecordLayout`/`LayoutField`, `standardForeignLayout`'s pattern)
  with a new **inhabitance/niche fact derived by the checker** from a type's
  enumerated bit-pattern space — explicitly **not** a hand-declared ability.
  **Reopening condition:** a payload type is added whose declared representation
  has a provably invalid bit pattern (a non-null-guaranteed pointer, or a
  range-restricted integer).

- **D-12-25 (layout is ONE shared derived fact, not three independent
  derivations):** compute layout **once** and have the core IR, `interp`, and
  `cgen` all read that single `RecordLayout`-shaped fact. Agreement is then a
  **fact of construction**, and the comparator's job is proving each consumer
  **reads the shared fact correctly** — not proving three derivations coincide.
  This is a deliberate, narrow exception to the independence discipline that
  governs the four **admission** peers: independence is the evidence for a
  *judgement* (is this program admissible), whereas layout is a *definition*, and
  three competing definitions of one layout is not independence but ambiguity.
  — **Reversibility:** costly — three consumers would have to grow their own
  derivations to reverse it.

- **D-12-26 ("one meaning" means observable-behavior agreement, stated
  precisely):** `interp` has **no byte layout at all** — `values` is a
  `map[string]string` with no size, alignment, or offset anywhere. So criterion
  2's "one meaning in the core IR, the interpreter, and emitted C17" **cannot**
  mean byte-identical layout across all three. It means: the same alternative is
  live, the same payload value is extracted, and the same events are emitted in
  the same order — with layout-as-bytes a **C-only** obligation policed by
  `_Static_assert`/`offsetof` pairs for internal self-consistency. Write this
  down in the phase artifacts; a weaker claim stated precisely is worth more than
  a stronger claim that cannot be true.

### The affine hazard — resource payloads

- **D-12-27 (REFUSE resource-carrying payloads this phase, named and
  fail-closed):** a named refusal rejects any alternative whose payload type
  **structurally contains** a Phase-4 tracked-resource-derived value. RES-02 and
  criterion 1 are satisfied with `Byte`, `Buffer`, and nullary-ADT payloads —
  ordinary move-by-function-end, no resource involved.
  **The honest cost, stated rather than buried:** criterion 1's *resource* half
  is **deferred, not met**. Criterion 1 reads "the unmoved alternative's payload
  is still dropped exactly once", and the resource-specific reading of that is
  what this decision defers.
  — **Reversibility:** reversible — a refusal is removed when the rule lands;
  nothing built on top of it has to be unwound.

- **D-12-28 (why refusal rather than admission — the reason is the two open
  peer gaps, not timidity):** both sub-hazards land **precisely** on
  `corevalidate`'s two open, unreviewed gaps.
  - **Construction-side** (a resource moved *into* a payload, then returned) is
    "exactly the kind of multi-step dependency chain" **D-10-C02** warns
    `peerCalleeFrameDrained` (`corevalidate.go:2823`) may not survive — it
    detects resource-escape-via-return with a **single FORWARD pass** over
    `linear.Operations`, correct only under an **unstated and unenforced**
    assumption that operations are declaration-ordered by dependency, while its
    own cited precedent `peerParameterEscapesOwned` walks **BACKWARD** and is
    order-independent.
  - **Match-side / forwarding** (a payload constructed in a callee and forwarded
    out) hits **D-10-C01**'s missing `case core.OpCall` in
    `peerDeriveOriginFacts` — **re-verified open in tree 2026-09-12** (D-12-03).
    `corevalidate` independently refuses what `check` admits with zero
    diagnostics; fail-closed and conservative, not unsound, but it already
    **constrained which fixtures Phase 10 could express**, and any Phase 12
    fixture needing that shape hits the same wall.
  - **D-10-C04** sits on the same fault line and is **OPEN and UNOWNED**: Phase
    10's D-09-51 fix flipped `negative_control_fails.lang` and
    `negative_control_infallible.lang` from `check.interprocedural_loan_liveness`
    to `core.callee_not_callable`, and the executor's explicit request for human
    review **has not happened**.
  Admitting the rule now therefore risks discovering a genuinely new
  interprocedural fact **after fixtures are written** — the exact "hidden need
  for a new interprocedural rule" criterion 3 says must trigger a slip, not a
  silent absorption.

- **D-12-29 (what "dropped exactly once" can even mean here):** with **no
  `OpDrop`** in `AllOperationKinds()` and "drop" being a **type-level ability**
  (`core.AbilityDrop`, `internal/compiler/core/core.go:505`,
  `internal/compiler/ability/ability.go`), "dropped exactly once" is **not a
  countable runtime event**. It can only mean a static **by-function-end
  accounting property** — the same property `check`'s worklist and
  `corevalidate`'s set-propagation already derive independently. Say this in the
  phase artifacts rather than letting the criterion's wording imply a destructor
  count that does not exist.

- **D-12-30 (debt landing condition — three parts, all required):** the full
  resource-in-payload rule lands only when **D-10-C01 is closed** (the `OpCall`
  arm added to `peerDeriveOriginFacts`) **AND D-10-C02 is proven order-independent
  or fixed** **AND D-10-C04 is reviewed**. Record in the phase debt register with
  a non-empty landing phase per the register's mechanically-checked format.

### Inherited D-11-02 emitter deletion

- **D-12-31 (TAKE IT — gate-then-delete, inside Phase 12):** honor the
  contractual landing phase Phase 11 wrote into `PHASE-11-DEBT.md` ("Phase 12 —
  a green N=1 convergence differential (Q-05), landing inside Phase 12's
  per-dispatch-site `Result` plans, never as a second sweep"), with **no stated
  reversal needed**. Sequence: (i) the **N=1 convergence differential** lands and
  goes green as its own plan — a single-function program must produce identical
  output through `emitProgram` as through its legacy single-function emitter;
  (ii) **D-11-52's Match-body refusal is lifted** in `emitProgram`; (iii) the six
  emitters are **deleted**, with digests re-pinned **in a diff containing nothing
  else**; (iv) payload `case`s are then written **exactly once**, against a single
  emitter family.
  This is structurally **rustc's `-Z borrowck=migrate` pattern** — run old and
  new in parallel, diff, cut over only once parity is proven, then delete the old
  checker — and it is this project's own build-then-delete pattern (D-09-10,
  D-11-01/02) applied as written.
  **Why not re-defer:** re-deferral pays the "two coexisting laws" tax **twice** —
  duplicating every payload `case` across both families now **and** still owing
  the deletion later — and the D-11-52 collision **does not shrink by waiting**,
  it just moves.
  — **Reversibility:** one-way — once the six emitters are deleted and digests
  re-pinned, restoring them means re-establishing which emitter produced each
  committed golden.

- **D-12-32 (the collision, named so planning does not rediscover it):**
  `emitMatch` is simultaneously (i) one of the six emitters D-11-02 schedules for
  deletion, (ii) the emitter `Result` payload matching most needs, and (iii) the
  exact capability `emitProgram` currently refuses for multi-function programs —
  `"function %q: multi-function branch bodies are not supported by native
  emission this phase"` (`internal/compiler/cgen/cgen_program.go:152`). Phase 12
  cannot touch `Result` matching in native emission without landing in the middle
  of this. D-11-52 also records that **every** existing `defect` terminator is
  reached through a `core.Match` arm, because no arithmetic, `if`, or loops exist
  to reach it another way.

- **D-12-33 (TREE CORRECTION — `11-CONTEXT.md`'s emitter line numbers are
  stale):** Phase 11's own work moved them. Verified against the shipped tree
  2026-09-12, as the ledger plans must use:

  | Emitter / writer | Recorded in `11-CONTEXT.md` | Verified 2026-09-12 |
  |---|---|---|
  | `emitMatch` | 152 | **`cgen.go:152`** ✓ unchanged |
  | `emitLinear` | 234 | **`cgen.go:234`** ✓ unchanged |
  | `emitLinearBorrowedByPointer` | 456 | **`cgen.go:532`** |
  | `emitLinearBorrowedByPointerPlain` | 649 | **`cgen.go:726`** |
  | `emitLinearForeign` | 793 | **`cgen.go:870`** |
  | `emitBranch` | 1597 | **`cgen.go:1674`** |
  | `emitLinearForeignOutputSupport` | 1383-1395 | **`cgen.go:1516`** |
  | `emitProgram` | (new file) | **`cgen_program.go:128`** |
  | D-11-52 refusal | — | **`cgen_program.go:152`** |

  Also verified: `emitLinearOutputSupport` at `cgen.go:1481`,
  `emitBranchOperations` at `cgen.go:1762`, `emitProgramFunction` at
  `cgen_program.go:339`. Recorded rather than silently corrected, per D-09-45.

- **D-12-34 (scope hygiene — the deletion is NOT a RES-* success criterion):**
  deleting the six emitters serves **NAT-04..NAT-07** (Phase 11's requirements),
  not RES-02/RES-03. Represent it as **explicitly-labeled inherited-debt-closure
  plans** with their own **non-requirement-shaped** verification (differential
  green; digest identity), so Phase 12's actual requirement-shaped criteria stay
  requirement-shaped. This is how the project's one-phase-per-requirement /
  100%-coverage discipline survives an inherited obligation — the roadmap
  explicitly rejects phases "whose success criteria could only read as task
  completions", and smuggling the deletion into a RES-* criterion would do
  exactly that.

- **D-12-35 (what replaces the roadmap's now-stale scheduling note):** the
  ROADMAP's instruction — "land `Result` payload `case` additions in the *same
  plans* that already touch each dispatch site for `OpCall`, not as a second
  sweep over the same six files" — is **moot**: `OpCall` landed across Phases
  07-11, all complete, so there are **no remaining `OpCall` plans to fold into**.
  Its stated **purpose** (avoid a second sweep; avoid two coexisting laws) still
  binds. The replacement mechanism is D-12-31's ordering itself: by deleting the
  legacy family **before** payload `case`s are written, there is **exactly one
  law** to write them into, so a second sweep is impossible by construction
  rather than by instruction.

- **D-12-36 (the fallback, and its single trigger):** re-deferral with a stated
  reversal (naming D-11-02's superseded landing-phase text, per the D-09-08 /
  D-09-30 / D-10-27 precedent) becomes correct **only** if the N=1 convergence
  differential cannot be made green cheaply. That is a finding for the pre-flight
  probe window, not a mid-phase improvisation.

### Criterion 2 — the anti-vacuity control

- **D-12-37 (two controls, because the obvious one is structurally
  insufficient):** retain a **frozen-fixture** control in the
  `control:foreign.layout_mismatch` direction — a committed
  `testdata/phase12/*.golden.c` declaring the payload struct's alternative slots
  transposed or resized relative to the checker-derived layout fact, killed by
  the generated `_Static_assert`/`offsetof` pair under `-Werror`. It is
  **necessary** for the C-side obligation and **structurally incapable** of
  catching the real bug: a `_Static_assert` polices the **struct declaration's**
  `sizeof`/`_Alignof`/`offsetof`, never **which field a given match arm actually
  reads**. A `cgen` bug that reads the right-shaped struct at the **wrong slot**
  for a correct tag leaves every `_Static_assert` green.

- **D-12-38 (the DECISIVE control is a reverted production-hunk mutation):**
  seed a bug in `cgen`'s match-arm/constructor codegen that reads or writes the
  **wrong alternative's payload slot** for a correct tag, and require a genuine
  interp-vs-native **value** divergence — the Phase-5 `false_no_alias` shape
  (which produced "an actual `interpreter == -O0 (2) != -O3 (7)` divergence on
  Apple clang 21"), not a mere compile failure. Pair it with a **companion
  unmutated cross-fixture run** proving the comparator reports **agreement** when
  the bug is absent — that companion is what discriminates real value-identity
  checking from a harness that screams on any diff. Per D-10's rule, this is the
  **other** mutation direction (a production hunk, reverted cleanly, never left
  committed) and deliberately attacks a different artifact than D-12-37's frozen
  fixture.

- **D-12-39 (TREE CORRECTION — no harness extension is needed; the payload-value
  channel already exists):** the researcher hedged that this "requires the
  differential harness to compare extracted payload values at match sites, which
  it may not do today... if that channel doesn't exist in the harness yet, adding
  it is itself the fix". **The tree answers it.**
  `internal/compiler/session/session_phase5_compare.go:89-96` compares
  `left.Outcome.Value != right.Outcome.Value` on `axis:terminal-outcome`, and its
  own comment reads: "the closed value|typed_failure|defect set, **INCLUDING the
  ok payload and the err-edge ADT alternative**". So the slot-swap mutation is
  catchable on an **existing** axis with **no harness extension**, provided the
  wrong-slot payload reaches the terminal outcome. `execution.Event` additionally
  carries `FunctionID`/`SourcePlace`/`TargetPlace`/`TypeID`
  (`internal/compiler/execution/execution.go:41+`), giving `axis:event-order`
  partial mid-function observability. Recorded per D-09-45 — this materially
  reduces the control's cost and plans must not budget a harness extension that
  is already there.

- **D-12-40 (extend an existing axis; do NOT add a lane):** the five axes are
  `axis:terminal-outcome`, `axis:event-order`, `axis:resource-ledger`,
  `axis:exit-status-signal`, `axis:diagnostic-id`
  (`internal/compiler/session/session_phase5_compare.go:20-24`). Work within
  them. **D-11-42** is a standing constraint: `SelectLanesForFixture`'s
  change-state mechanism "MUST NOT BE EXTENDED to any new lane without the same
  scrutiny D-11-41/Q-02 applied to the native-differential lane", because "an
  undeclared input can move silently while the declared set reports no change".

- **D-12-41 (if a control turns out unconstructible, that is a DEFECT IN THE
  CRITERION):** per **D-11-36**, a control that cannot be built or turns out
  flaky must be **escalated as a defect in the criterion**, never silently
  treated as a tolerance requirement. D-12-23's niche finding is the live
  instance: the niche-specific control is **unconstructible** because there is no
  niche, and criterion 2 must state that as a finding rather than skip the check.

- **D-12-42 (QLT-09's Phase-5 Nyquist portion is NOT closed this phase):**
  Phase 12 *is* the native-tier change QLT-09 was waiting on ("deliberately
  deferred because both surfaces are touched again by M002 (native tier, agent
  surface); validating now risks validating a shape that is about to change"),
  but the change is still **in flight within** Phase 12. Closing it mid-phase
  would conflate a fresh semantic-verification commit with a Nyquist-debt
  closure — D-11-01's digest-conflation concern. `12-VALIDATION.md` scopes itself
  to criterion 2's new control and **explicitly declines** to close QLT-09's
  Phase-5 portion. It becomes validatable once Phase 12's layout work is frozen
  (after the D-12-31 deletion lands), making end-of-Phase-12 or Phase 13 the
  defensible home, decided against a **pre-registered threshold** in the
  QLT-07/D-09-40 style rather than folded silently into this criterion.
  Note `workflow.nyquist_validation` is enabled, so `12-VALIDATION.md` is
  produced regardless — the question was only what it should claim.

### Claude's Discretion

- Plan decomposition and wave structure, subject to D-12-31's ordering and
  D-12-34's labeling requirement.
- Which file each new derivation lives in, subject to the existing
  package-independence import allowlists.
- Exact diagnostic code strings for D-12-15's three refusals and D-12-27's
  resource refusal, within the established `check.*` / `core.*` namespacing.
- Whether `Fault` (the example payload ADT in `| Err(Fault)`) or another name is
  used in fixtures.
- Whether the corpus characterization replay (D-12-19) lives as one test or a
  table-driven family.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase 12's own charter

- `.planning/ROADMAP.md` § "Phase 12: `Result` Payloads" — goal, the three
  success criteria, the riskiest assumption, the gate, **and the now-stale
  scheduling note superseded by D-12-35**.
- `.planning/REQUIREMENTS.md` — RES-02 (line 153) and RES-03 (line 157), the
  only two requirements this phase carries.
- `.planning/research/ARCHITECTURE.md` **§7 "`Result` with Payloads"**
  (lines 656-730) — D-04-30's recorded shape, the three-engine layout [INFER],
  and the affine-interaction paragraph that names this phase's one genuinely new
  hazard. **§6** (lines 640-654) for the `-flto` / interprocedural-negative-control
  reasoning. **Stage 9** (lines 855-868) for the structural-independence claim and
  the original scheduling argument.
- `.planning/research/SUMMARY.md` lines 384, 641, 699, 729 — the `Result` payload
  stage summary and its assumed-not-confirmed independence note.
- `.planning/research/OPTIONS.md` lines 39, 209, 462 — `Result` payload work as a
  stretch/fallback item and the scope-cut order.

### Project-level durable context — read before re-deriving anything

- `.planning/LANGUAGE-MATURITY.md` — **mandatory.** The language is far less
  expressive than the roadmap vocabulary implies: no arithmetic, no iteration, no
  `if`, no strings, no arrays; `Byte`/`Buffer` only. Includes the 32-site
  single-function guard inventory. **Do not read `wiki/example-tour.md` as a
  description of the language** — its effect rows, `?`, generics, and `spec`
  blocks are unimplemented design target.
- `.planning/STANDING-VERDICTS.md` — the six dispatch sites are literal, named,
  and test-enforced; D-04-22's 8-10-edit floor; why `-flto` is load-bearing; why
  `core.Interface`/`core.FunctionSignature` is the home for callee summaries.
- `.planning/PROJECT.md` § "Current Milestone: M002" — and the Out of Scope
  table in `REQUIREMENTS.md`, which rules out implicit `From`-style error
  conversion, call-site ownership override, implicit ARC fallback, and full
  generics, each with a recorded reason.

### Inherited obligations and open gaps — all load-bearing for this phase

- `.planning/phases/11-multi-function-native-emission-and-interprocedural-equivalen/PHASE-11-DEBT.md`
  — **mandatory.** D-11-02 (the six-emitter deletion, landing phase = Phase 12),
  D-11-52 (the `emitMatch` collision), D-11-51 (duplicate `function.returned`
  event IDs), D-11-42 (the lane-addition constraint), D-11-07, D-11-36, and the
  five Phase 10 carry-forwards D-10-C01..D-10-C05.
- `.planning/phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-CONTEXT.md`
  — D-11-01 (the additive emitter path and the digest-attribution reasoning),
  D-11-03 (body-emitter vs TU-assembler split), D-11-04 (one `emitCall`),
  D-11-05 (`callgraph.EntryFunction`), D-11-06 (why no new `core.Program` field),
  D-11-08 (two-tier name allocation), D-11-09 (the zero-attribute terminal state
  — the precedent D-12-23/D-12-24 follow). **Its emitter line numbers are stale;
  use D-12-33's verified ledger instead.**
- `.planning/phases/10-trusted-interprocedural-oracle/10-CONTEXT.md` — D-10-02
  (narrow callee-contract map, not `core.Program`), D-10-04 (minimal value type
  plus the signature-text assertion), D-10-27 (a recorded reversal's form).
- `.planning/phases/09-peer-re-derivation-and-d-03-02-closure/09-CONTEXT.md` —
  D-09-02 (independence is derivation method plus import boundary, never file
  location), D-09-10 (build-then-delete), D-09-45 (the stale-reference
  discipline this CONTEXT.md follows in D-12-13/D-12-33/D-12-39).

### Live source the plans must read, with verified locations

- `internal/compiler/core/core.go` — `DataType` :22-27, `Function` :29-52,
  `Match` :842, `MatchArm` :861, `AllOperationKinds()` :704,
  `TerminatorKinds()` :711, `AbilityDrop` :505, `Schema`/`Schema1` :10-11,
  `RecordLayout`/`LayoutField` :79-130.
- `internal/compiler/core/core_test.go` —
  `TestAllOperationKindsHandledAtEverySite` :267
  (`control:kind.exhaustive_dispatch`); the four pinned golden-C digests :156-159.
- `internal/compiler/ast/ast.go` :111-120 — `MatchArm` (and `Alternative`); the
  doc comment that falsifies the binder claim (D-12-13).
- `internal/compiler/syntax/parser.go` :525-560 — the match-arm parse, including
  `p.identifier("syntax.expected_pattern")` at :530; `maxAlternatives` :17,
  `maxArmsPerMatch`.
- `internal/compiler/interp/interp.go` — `values map[string]string` :308; frame
  seeds :205/:219/:286; `newFlatFrame` :357, `newBlockFrame` :374,
  `newArmFrame` :427; `placeTypeIndex` :406; caller-argument handling :452-470;
  operation arms :622-642, :666-667, :693, :727.
- `internal/compiler/cgen/cgen.go` and `cgen_program.go` — **use D-12-33's
  verified line ledger**, not `11-CONTEXT.md`'s.
- `internal/compiler/corevalidate/corevalidate.go` — `peerDeriveOriginFacts`
  :2407 (**no `case core.OpCall`** — D-10-C01, re-verified open),
  `peerCalleeFrameDrained` :2823 (D-10-C02's forward-pass assumption).
- `internal/compiler/check/check.go` — `checkBranch` :1597,
  `standardForeignLayout` :3118.
- `internal/compiler/session/session_phase5_compare.go` — the five axis
  constants :20-24; `comparePhase5Pair` :89 with the `Outcome.Value` comparison
  and the "INCLUDING the ok payload" comment (D-12-39).
- `internal/compiler/session/session.go` :565-587 — `LayoutMutationRunner`
  (`control:foreign.layout_mismatch`), the frozen-fixture precedent D-12-37
  follows; `testdata/phase4/foreign_layout_mismatch.golden.c` is the fixture.
- `internal/compiler/native/native.go` :553+ — `validateExecution`, including the
  duplicate-event-id refusal D-12-21 flags.
- `internal/compiler/execution/execution.go` :36-50 — `Outcome{Kind, Value}` and
  `Event`.
- `internal/compiler/ability/ability.go` — the ability vocabulary D-12-24's
  recorded niche design would extend.
- `internal/compiler/session/session_test.go` :2613 —
  `TestDebtRegistersAreWellFormed`, the format D-12-24/D-12-30's debt rows must
  satisfy.
- `testdata/phase4/defect_terminal.lang` — the canonical `data` + `match` +
  `defect` source shape the new syntax extends.

### External prior art cited in the decisions

- Rust MIR (`AggregateKind::Adt` kept out of `Rvalue::Use`), `rustc_abi` enum
  layout and the niche/tagged-union RFCs, RFC 2005 match ergonomics (D-12-14's
  named risk), and `-Z borrowck=migrate` (D-12-31's pattern:
  <https://github.com/rust-lang/rust/issues/46908>,
  <https://blog.rust-lang.org/2022/08/05/nll-by-default/>).
- Swift SIL (`enum` vs `unchecked_enum_data` as distinct verifier-checked
  instructions); Swift resilient enum layout and spare-bit discriminants.
- C17 §6.5.2.3 active-member and common-initial-sequence rules (D-12-22).
- Protobuf `oneof` / FlatBuffers union-type-vector pairing (D-12-09's
  generated-accessor mechanism).

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- **`core.RecordLayout` / `core.LayoutField` plus `standardForeignLayout`
  (`check.go:3118`)** — the checker-derived layout mechanism already shipped for
  the `{ok, value}` foreign ABI struct. D-12-22/D-12-25 extend it rather than
  inventing a payload-layout path; its `_Static_assert` conformance generation is
  the C-side policing D-12-37 reuses.
- **`ast.Alternative{Name, Span}`** already exists as a struct, so a payload
  type field there is additive rather than a shape change.
- **`control:foreign.layout_mismatch` / `LayoutMutationRunner`
  (`session.go:565-587`)** — a complete frozen-fixture mutation control,
  including the rule that no correct version of the mutated fixture is ever
  committed. D-12-37 clones the pattern for payload layout.
- **`axis:terminal-outcome`'s existing `Outcome.Value` comparison
  (`session_phase5_compare.go:89-96`)** — already compares "the ok payload and
  the err-edge ADT alternative", so D-12-38's control needs **no** new harness
  channel (D-12-39).
- **`corevalidate`'s peer substrate** — `peerPostorder`, `peerSignatures`, and
  the forward-set-propagation idiom (`peerDeriveOriginFacts` :2407,
  `peerParameterEscapesOwned`, `peerCalleeFrameDrained` :2823). D-09-01's lesson
  applies: extend the substrate, do not build a second one.
- **`reduce`'s `Axis`-keyed predicate machinery** (`reduce/predicate.go:14-18`,
  `reduce/mismatch.go:75`) — already parameterized over the five axis names, so a
  payload-layout divergence reduces without new reducer plumbing.

### Established Patterns

- **Additive omitempty sibling fields, never a shape change** — every
  post-Phase-1 fact on `core.Function` (`Match`, `Linear`, `PublicOrigin`,
  `ForeignContract`) follows it, and `MatchArm.BlockID`'s own doc comment records
  the byte-freeze reasoning. D-12-08 depends on this holding.
- **The six literal dispatch sites plus `control:kind.exhaustive_dispatch`** —
  the control that makes D-12-05's new kinds visible and that D-12-05 argues
  would have passed vacuously under D-04-30's shape.
- **Independence by derivation method plus import boundary** (D-09-02) — governs
  the four admission peers. D-12-25 carves a deliberate, narrow exception for
  *layout*, on the grounds that layout is a definition rather than a judgement.
- **Fail-closed named refusals with layer-namespaced codes** —
  `core.callee_not_callable`, `check.interprocedural_loan_liveness`,
  `core.origin_omitted`, `syntax.expected_pattern`. D-12-15 and D-12-27 add to
  this namespace.
- **Declared caps, fail-closed** — `maxAlternatives` 4096, `maxArmsPerMatch`,
  `pathoracle`'s `MaxPaths`, `interp`'s `MaxCallDepth`. A payload feature must
  not introduce an unbounded dimension without a declared cap.
- **Build-then-delete** (D-09-10, D-11-01/02) — D-12-31 applies it as written.
- **Record corrections, never silently fix** (D-09-45) — D-12-13, D-12-33, and
  D-12-39 follow it.

### Integration Points

- `core.DataType` and `core.LinearOperation` — the two IR types that grow
  (D-12-05, D-12-07, D-12-09, D-12-10).
- `syntax/parser.go:530` — where the pattern binder is parsed (D-12-11, D-12-13).
- `check.go:checkBranch` :1597 — where arm-level place binding already happens
  and where the `OpMove`-on-bind of D-12-14 is emitted.
- `interp.go` ~17 sites — the value-type widening (D-12-17..D-12-20).
- `cgen_program.go:152` — D-11-52's refusal, lifted by D-12-31 step (ii).
- `cgen.go`'s six emitters — deleted by D-12-31 step (iii); see D-12-33's ledger.
- `corevalidate.go` :2407 and :2823 — not modified this phase, but both gate
  D-12-30's debt landing condition; the probe (D-12-03) reads :2407's behavior.
- `session_phase5_compare.go` — the axis the D-12-38 control runs on.

</code_context>

<specifics>
## Specific Ideas

- **The developer's standing research mandate**, restated verbatim in this
  discussion and applied to all eight areas: for each decision point, fan out
  across breadth and depth through every relevant stakeholder-role lens
  (security, product, technical, design, and the technical specialties —
  architect, devops, the relevant language/framework expertise, and so on),
  consider pros/cons/tradeoffs/examples, anti-patterns/patterns/best
  practices/footguns/lessons-learned, research online and draw insight from other
  products and ecosystems where helpful, run an **adversarial pass**, then
  synthesize one-shot recommendations. For UI/UX-shaped questions, add existing
  patterns, design patterns, the **principle of least surprise**, conventions,
  and elegant design style. The developer noted the mandate should be adapted to
  the project's actual domain rather than applied literally — here it was adapted
  to compiler/IR, type-system, C-backend/ABI, runtime, and verification lenses,
  with the UI/UX half applied to **area 3** on the grounds that source syntax
  **is** a UX surface whose primary author is an AI agent and whose primary
  reviewer is a human.
- `data Outcome = | Ok(Buffer) | Err(Fault)` is the concrete shape agreed for the
  general feature; `Result` appears as one canonical fixture name, not a built-in.
- The probe fixture is deliberately the **darkest corner**, not a representative
  average (D-12-02).
- Criterion 2's honest claim is **observable-behavior agreement**, with
  layout-as-bytes a C-only obligation (D-12-26) — a weaker claim stated precisely
  beats a stronger one that cannot be true.

</specifics>

<deferred>
## Deferred Ideas

- **Resource-carrying payloads** — refused this phase (D-12-27). Landing
  condition is three-part and explicit (D-12-30): D-10-C01 closed **and** D-10-C02
  proven order-independent or fixed **and** D-10-C04 reviewed.
- **Niche optimization** — designed and recorded, not built (D-12-24). Reopening
  condition: a payload type arrives whose declared representation has a provably
  invalid bit pattern. Currently **uninstantiable** because `Byte` is plausibly
  fully inhabited (D-12-23).
- **D-10-C01's missing `case core.OpCall`** in `peerDeriveOriginFacts` — not
  repaired here; the option to close it first and then admit resource payloads
  was considered and declined as importing Phase 10 debt repair on top of the
  Phase 11 NAT-* debt this phase already takes.
- **D-10-C02's order-independence** for `peerCalleeFrameDrained` — not proven
  here.
- **D-10-C04's human review** of Phase 10's negative-control verdict flip —
  still **OPEN and UNOWNED**. Phase 12 does not claim it.
- **D-11-51's per-invocation event identity** — flagged as likely to be tripped
  first by this phase's fixtures (D-12-21), but its fix needs `interp.go` **and**
  `cgen_program.go` and therefore its own reviewed plan. Not claimed here.
- **D-11-07's `lang.foreign/0` widening** — still OPEN and unowned.
- **QLT-09's Phase-5 Nyquist portion** — deliberately not closed mid-phase
  (D-12-42); end-of-Phase-12 or Phase 13, against a pre-registered threshold.
- **A real generic `Result<T, E>`** — GEN-01, M003. D-12-12 is chosen
  specifically so generics can absorb it without a breaking migration.
- **Revisiting D-12-16's bare-tag-plus-`take` syntax** — would need a
  flow-sensitive narrowing pass; recorded so it is not relitigated without that
  prerequisite.

</deferred>

---

*Phase: 12-result-payloads*
*Context gathered: 2026-09-12*
