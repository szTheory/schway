# Phase 12: `Result` Payloads - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-09-12
**Phase:** 12-result-payloads
**Areas discussed:** Criterion 3's pre-flight probe, Core IR payload representation, Source syntax and the pattern binder, Interpreter value model widening, C17 layout and the niche-optimization clause, The affine hazard (unmatched resource payloads), Inherited D-11-02 emitter deletion, Anti-vacuity control for criterion 2

**Mode:** advisor (`$HOME/.claude/gsd-core/USER-PROFILE.md` present).
`vendor_philosophy: opinionated` → calibration tier **`minimal_decisive`**.
`technical_background: true` + `explanation_depth: practical-detailed:technical`
→ `NON_TECHNICAL_OWNER = false` (the explicit technical flag overrides every
inferred non-technical signal, per the advisor tie-breaker), so **no
product-language reframing** was applied.
Advisor model: `sonnet`. Eight `gsd-advisor-researcher` agents spawned in
parallel, one per selected area.

All eight gray areas offered were selected. The developer restated their standing
fan-out research mandate (see CONTEXT.md `<specifics>`) and asked that it be
adapted to this project's domain rather than applied literally.

---

## Area selection

| Option | Description | Selected |
|--------|-------------|----------|
| Criterion 3's pre-flight probe | Hard entry gate; probe must be answered before planning; failure branch is an M003 slip | ✓ |
| Core IR payload representation | D-04-30's omitempty-field shape vs new OperationKinds; `lang.core/2` risk | ✓ |
| Source syntax and the pattern binder | No binder exists in the grammar; general feature vs built-in `Result` | ✓ |
| Interpreter value model widening | `values map[string]string` cannot hold tag+payload | ✓ |
| C17 layout and the niche-optimization clause | Real `union` vs flat struct; is niche built or declared | ✓ |
| The affine hazard: unmatched resource payloads | ARCHITECTURE's one genuinely new hazard; admit or refuse | ✓ |
| Inherited D-11-02 emitter deletion | Phase 11 scheduled it here; collides with D-11-52 | ✓ |
| Anti-vacuity control for criterion 2 | Without a seeded fault, criterion 2 passes vacuously | ✓ |

**Notes:** no area was declined. No scope creep was raised by the developer —
every deferred item in CONTEXT.md originates from a prior phase's debt register or
from a researcher's own recommendation, not from an in-discussion request.

---

## 1. Criterion 3's pre-flight probe

| Option | Description | Selected |
|--------|-------------|----------|
| Committed single-branch Go test, Q-01 shape | Phase 11's precedent; permanent regression; companion clean-edit assertion | ✓ |
| Throwaway `.planning/spikes/` workbench | Matches the spike discipline; prose verdict; lower ceremony | |
| Bounded static inventory pass, S-008-replacement shape | Count affected sites against a pre-registered threshold | |

**User's choice:** adopt the synthesized recommendation (committed Go test).
**Notes:** the researcher's decisive argument was that criterion 3's question is
**binary**, matching Q-01, not **gradational**, which is what the S-008
replacement's threshold shape is for — so the inventory form is structurally the
wrong instrument. The spike-prose option was rejected on the project's own
mechanism-over-convention grounds: it leaves the gate's evidentiary basis with no
standing regression. The orchestrator re-verified D-10-C01's premise in tree
before locking (`peerDeriveOriginFacts` at `corevalidate.go:2407` has cases for
`OpBorrowShared`, `OpBorrowExclusive`, `OpMove`/`OpCopy` and **no** `case
core.OpCall`).

---

## 2. Core IR payload representation

| Option | Description | Selected |
|--------|-------------|----------|
| Two new `OperationKind`s | ~16-20 edits across six sites; new semantics visible to the exhaustive-dispatch control | ✓ |
| Follow D-04-30 as written | Ride existing kinds via additive omitempty fields; cheapest; accepts the vacuous-pass hazard | |
| Hybrid: construct rides, destructure gets a kind | Half the edit cost; the new affine obligation still visible at all six sites | |

**User's choice:** Two new OperationKinds.
**Notes:** this was presented as a genuine fork because the researcher **rejected
D-04-30's own named design**, which the ROADMAP lists as a canonical ref — so
adopting it means deviating from a recorded design plan, and CONTEXT.md D-12-05
states that deviation explicitly with the superseded text named. The researcher's
two strongest points: (i) the `OkEdgeID`/`ErrEdgeID` precedent D-04-30 leans on
does **not** support it, because those fields refine `OpForeignCall`'s single
existing meaning rather than bolting a second meaning onto a general-purpose
kind; (ii) Rust MIR and Swift SIL both independently drew the same line
(`AggregateKind::Adt` out of `Rvalue::Use`; `enum` vs `unchecked_enum_data` as
distinct verifier-checked instructions). The hybrid was rejected because
destructuring must get a new kind under every candidate anyway, so the marginal
cost of also giving construction one buys the removal of the conditional-meaning
hazard entirely. The hybrid's rejection rationale partly relied on RES-03 making
construction layout-relevant — which area 5 then found is vacuous — and that
weakening was surfaced in the question rather than hidden.

---

## 3. Source syntax and the pattern binder

| Option | Description | Selected |
|--------|-------------|----------|
| `\| Ok(Buffer)` / `Ok(v) => {...}`, general `data`, move-on-bind | Constructor-application; the spelling Rust/Swift/OCaml/Scala/Kotlin/Zig all use | ✓ |
| Bare tag + body-level `take` | `Ok => { let v = take flag }`; smallest grammar diff, `MatchArm.Pattern` untouched | |

**User's choice:** adopt the synthesized recommendation.
**Notes:** decided on least surprise for both audiences — the AI agent that
authors Lang code and the human who reviews it. No researched ecosystem (Rust,
Swift, OCaml, Zig, Erlang, Austral, Koka, Val/Hylo) spells payload extraction the
bare-tag way, so it has no least-surprise precedent. The bare-tag option was
further rejected because it needs flow-sensitive narrowing the checker does not
have — **more** type-system machinery than the binder field it saves — and it
would overload `take`'s established meaning.

On the `Result`-as-built-in question the researcher came down firmly on a
**general `data` feature with `Result` as a user fixture**: a blessed monomorphic
built-in is the ambient special-cased authority the project's constitution argues
against, and the general form is what GEN-01 (M003) can absorb without a breaking
migration.

The researcher's own flagged risk was adopted as a binding constraint rather than
noted and dropped: move-on-bind *looks* like naming but *is* consuming, which is
Rust RFC 2005's exact match-ergonomics surprise, so the diagnostics must be loud.

**Orchestrator correction (tree wins).** The researcher claimed this "extends
today's 'one bound place per arm' (`MatchArm.Value`)". Verified wrong —
`ast.MatchArm.Value` (`ast/ast.go:111-120`) is the arm's bare-name **result
expression**, mutually exclusive with `Body` per its own doc comment, not a
binder. The parser reads one identifier then `=>` (`syntax/parser.go:530-533`),
so `Ok(v)` is genuine new parser work. Recorded as CONTEXT.md D-12-13 per the
D-09-45 stale-reference discipline.

---

## 4. Interpreter value model widening

| Option | Description | Selected |
|--------|-------------|----------|
| Widen to a struct `value{tag, payload string}` | Compiler forces every site; zero-value struct = scalar; no nil-interface state | ✓ |
| Widen to an interface / sum type | Extensible; but dynamic dispatch, nil-interface footguns, type switch per arm | |
| Keep `map[string]string`, encode into the string | Narrowest diff — rejected as in-band signalling | |
| Parallel side maps | Zero churn on existing arms — rejected as parallel-map desynchronization | |

**User's choice:** adopt the synthesized recommendation.
**Notes:** options 3 and 4 were rejected as the project's own named
anti-patterns, not merely as inferior. The decisive design move is making the
change **evidence-invisible by construction**: every non-payload site writes
`tag: ""` and serialization special-cases `tag == ""` to emit the bare string, so
no frozen evidence byte moves — the same reasoning as D-11-01's additive emitter
path ("when a digest moves you cannot tell which change did it"). Proof is a
corpus characterization replay asserting byte-identical execution documents, plus
a narrow scalar round-trip property test.

The researcher explicitly flagged rather than absorbed the D-11-51 interaction: a
multi-call-site payload-return fixture is exactly what would trip that open
defect for the first time. Orchestrator verified the blast radius at ~17 sites in
`interp/interp.go`, matching the researcher's 15-20 estimate.

---

## 5. C17 layout and the niche-optimization clause

| Option | Description | Selected |
|--------|-------------|----------|
| Flat per-alternative-field struct; niche designed-not-built | Checker-derivable provenance, no UB, reuses `RecordLayout`/`standardForeignLayout` | ✓ |
| Real C `union` payload field | Smallest footprint; matches "tagged union" literally | |

**User's choice:** adopt the synthesized recommendation.
**Notes:** the `union` was rejected on three grounds — it would be the project's
first non-checker-derived ABI claim (nothing tracks the active member), reading an
inactive member is UB in C17 outside the common-initial-sequence rule, and
`-O3 -flto` with TBAA plus sanitizers under `-Werror` is precisely the proof
regime that would bite. The space counter-argument was priced rather than waved
away: with only `Byte` and `Buffer` available, the waste is one tag byte over two
payload types.

The significant finding is that **niche optimization is uninstantiable at this
maturity** — `Byte` is plausibly fully inhabited, so there is no niche to hide a
tag in, and RES-03's niche clause is vacuously satisfiable. The researcher
identified this as the same *shape* of finding as D-11-09 (a mechanism whose
precondition cannot presently exist), so niche becomes designed-and-recorded debt
with an explicit reopening condition rather than inert implemented code.

Also settled here: layout is **one shared derived fact** all three engines read,
not three independent derivations — a deliberate narrow exception to the
independence discipline, justified on the grounds that layout is a *definition*
whereas independence is evidence for a *judgement*.

---

## 6. The affine hazard: unmatched resource payloads

| Option | Description | Selected |
|--------|-------------|----------|
| Refuse resource payloads, declare debt | Named fail-closed refusal; RES-02 met with Byte/Buffer/nullary-ADT payloads | ✓ |
| Admit it — attempt the full rule | Extend Phase 4's tracked-resource analysis through payload construction and match arms | |
| Close D-10-C01 first, then admit | Add the missing `OpCall` arm as a prerequisite, then admit on the repaired foundation | |

**User's choice:** Refuse resource payloads, declare debt.
**Notes:** presented as a genuine fork because it **narrows criterion 1's literal
reading** — "the unmoved alternative's payload is still dropped exactly once"
would be exercised with no resource involved — and CONTEXT.md D-12-27 states that
cost rather than burying it.

The researcher's argument is that both sub-hazards land precisely on
`corevalidate`'s two open unreviewed gaps: construction-then-return is the
multi-step dependency chain D-10-C02 warns `peerCalleeFrameDrained`'s unenforced
ordering assumption may not survive, and payload-forwarded-out-of-a-callee hits
D-10-C01's missing `OpCall` arm. D-10-C04 sits on the same fault line and remains
unreviewed. So admitting now risks discovering a genuinely new interprocedural
fact **after** fixtures are written — the exact trigger criterion 3 says must
cause a slip rather than a silent absorption.

The third option (repair D-10-C01 first) was declined because it imports Phase 10
debt repair on top of the Phase 11 NAT-* debt area 7 already brings in, while
D-10-C02 and D-10-C04 would still be open.

Also settled: with no `OpDrop` and drop being a type-level ability, "dropped
exactly once" can only mean a static by-function-end accounting property, not a
countable runtime event — recorded so the criterion's wording is not read as
implying a destructor count that does not exist.

---

## 7. Inherited D-11-02 emitter deletion

| Option | Description | Selected |
|--------|-------------|----------|
| Take it: gate-then-delete | N=1 differential green, lift D-11-52, delete, then write payload cases once | ✓ |
| Re-defer with a stated reversal | Amend D-11-02's landing phase per the D-09-08/D-09-30/D-10-27 precedent | |
| Minimal: lift D-11-52 only, defer deletion | Take only the strictly-blocking piece; smallest honest scope | |

**User's choice:** Take it: gate-then-delete.
**Notes:** presented as a genuine fork because the work serves NAT-04..NAT-07,
not RES-02/RES-03, and materially grows the phase. The researcher identified the
ordering as structurally **rustc's `-Z borrowck=migrate`** pattern — run both,
diff, cut over once parity is proven, then delete — and argued re-deferral pays
the "two coexisting laws" tax **twice** (duplicate every payload `case` now, and
still owe the deletion later), since the D-11-52 collision does not shrink by
waiting.

Scope hygiene was settled explicitly: the deletion is represented as
**explicitly-labeled inherited-debt-closure plans** with non-requirement-shaped
verification (differential green, digest identity), so Phase 12's RES-* criteria
stay requirement-shaped and the roadmap's rejection of task-completion-shaped
criteria is respected.

Re-deferral is retained as the **fallback with a single trigger**: the N=1
differential cannot be made green cheaply — a finding for the pre-flight probe
window, not a mid-phase improvisation.

**Orchestrator correction (tree wins).** `11-CONTEXT.md`'s emitter line numbers
are stale, moved by Phase 11's own work. Verified current values recorded as a
ledger in CONTEXT.md D-12-33 (`emitMatch` 152 and `emitLinear` 234 unchanged;
`emitLinearBorrowedByPointer` 456→532, `...Plain` 649→726, `emitLinearForeign`
793→870, `emitBranch` 1597→1674, `emitLinearForeignOutputSupport` 1383-1395→1516;
`emitProgram` at `cgen_program.go:128` with D-11-52's refusal at `:152`).

---

## 8. Anti-vacuity control for criterion 2

| Option | Description | Selected |
|--------|-------------|----------|
| Frozen-fixture struct-shape control, `foreign_layout_mismatch` style | Strong precedent; necessary but structurally insufficient alone | ✓ (retained as baseline) |
| Reverted production-hunk emitter mutation — wrong alternative's slot for a correct tag | Produces a genuine interp-vs-native **value** divergence; the decisive control | ✓ (decisive) |

**User's choice:** adopt the synthesized recommendation (both, with the
production-hunk mutation as the decisive one).
**Notes:** the researcher's key insight is that a `_Static_assert` polices the
**struct declaration's** `sizeof`/`_Alignof`/`offsetof`, never **which field a
match arm actually reads** — so a `cgen` bug reading the right-shaped struct at
the wrong slot leaves every assert green while native produces a value `interp`
does not. Per D-10's rule the two controls deliberately attack **different
artifacts** (a frozen committed fixture with no correct version in tree, versus a
cleanly-reverted production hunk). A companion unmutated cross-fixture run
discriminates real value-identity checking from a harness that screams on any
diff. Work within the existing five axes; D-11-42 forbids casually adding a lane.

Also settled: criterion 2's "one meaning in the core IR, the interpreter, and
emitted C17" **cannot** mean byte-identical layout, because `interp` has no
sizes, alignments, or offsets at all. It means observable-behavior agreement,
with layout-as-bytes a C-only obligation.

**Orchestrator correction (tree wins).** The researcher hedged that the harness
may not compare extracted payload values and that "adding it is itself the fix".
The tree answers it: `session_phase5_compare.go:89-96` already compares
`left.Outcome.Value != right.Outcome.Value` on `axis:terminal-outcome`, with its
own comment reading "INCLUDING the ok payload and the err-edge ADT alternative".
So **no harness extension is needed** provided the wrong-slot payload reaches the
terminal outcome, and plans must not budget one. Recorded as CONTEXT.md D-12-39.

On Nyquist debt: `12-VALIDATION.md` scopes itself to criterion 2's new control and
**explicitly declines** to close QLT-09's Phase-5 portion mid-phase, to avoid
conflating a semantic-verification commit with a debt closure (D-11-01's
digest-conflation concern).

---

## Claude's Discretion

- Plan decomposition and wave structure, subject to D-12-31's ordering and
  D-12-34's labeling requirement.
- Which file each new derivation lives in, subject to the package-independence
  import allowlists.
- Exact diagnostic code strings for the three new pattern refusals and the
  resource-payload refusal, within the established namespacing.
- Fixture naming (e.g. whether `Fault` is the example error payload ADT).
- Whether the corpus characterization replay is one test or a table-driven
  family.

## Deferred Ideas

Every deferred item originates from a prior phase's debt register or from a
researcher's recommendation — none from an in-discussion scope request. Full list
with landing conditions is in CONTEXT.md `<deferred>`:

- Resource-carrying payloads (three-part landing condition: D-10-C01 closed,
  D-10-C02 proven order-independent, D-10-C04 reviewed).
- Niche optimization (reopening condition: a payload type with a provably invalid
  bit pattern).
- D-10-C01's missing `case core.OpCall`; D-10-C02's order-independence;
  D-10-C04's human review (OPEN and UNOWNED).
- D-11-51's per-invocation event identity — flagged as likely to be tripped first
  by this phase's fixtures; needs its own reviewed plan.
- D-11-07's `lang.foreign/0` widening (OPEN and unowned).
- QLT-09's Phase-5 Nyquist portion — end-of-Phase-12 or Phase 13, against a
  pre-registered threshold.
- A real generic `Result<T, E>` — GEN-01, M003.
- Revisiting the bare-tag-plus-`take` syntax — would need a flow-sensitive
  narrowing pass.
