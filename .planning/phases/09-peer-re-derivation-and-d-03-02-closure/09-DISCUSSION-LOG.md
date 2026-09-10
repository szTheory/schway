# Phase 09: Peer Re-Derivation and D-03-02 Closure - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-09-10
**Phase:** 09-peer-re-derivation-and-d-03-02-closure
**Mode:** advisor (`USER-PROFILE.md` present; `vendor_philosophy: opinionated` →
calibration tier `minimal_decisive`; `NON_TECHNICAL_OWNER = false`, overridden by
`technical_background: true`)
**Areas discussed:** Peer independence mechanism; OWN-09 law retirement; TRU-04
shadow-run gate design; D-07-33 narrowed-peer closure scope; S-008 / QLT-07
status; D-08-26 accepted-program disclosure; OWN-05 call-site convention
override; diagnostic-code promotion mechanics

---

## How this discussion ran

The developer selected **all eight** gray areas and supplied their standing
boilerplate research mandate: for each decision point, fan out across all
relevant stakeholder-role lenses, consider pros/cons/tradeoffs/examples,
antipatterns/patterns/best-practices/footguns/lessons-learned, research online
and draw insight from other products and ecosystems, run an adversarial pass,
then synthesize one-shot best recommendations.

Eight `gsd-advisor-researcher` agents ran in parallel (model `sonnet`). The
orchestrator then independently re-verified every factual claim against the
shipped tree before locking anything. **Three researcher claims required
correction; the tree won in each case.** The developer selected "Adopt all 8".

---

## Area 1 — Peer independence mechanism

| Option | Description | Selected |
|--------|-------------|----------|
| Extend the existing `peerSignatures`/`peerPostorder` substrate | Add peer-derived liveness bits via forward set-propagation, folded into the existing postorder loop; make `buildLoanChainIndex` contract-aware | ✓ |
| Build a wholly new liveness-specific peer mechanism | Own DFS/ordering pass, own memo table, possibly a new subpackage | |

**User's choice:** Adopt synthesized recommendation (extend the existing substrate).
**Notes:** Grounding finding — `corevalidate` already ships a working
peer-derivation substrate for the structurally analogous Phase 07 problem
(`v.peerPostorder`, `v.peerSignatures` refined callee-before-caller at
`corevalidate.go:505-547`). A second from-scratch mechanism buys no additional
independence — that comes from derivation *method* and the import boundary, never
from file or package location — while doubling the ordering surface the phase's
own risk statement warns about.

**Orchestrator correction (→ D-09-05):** the researcher left open whether the peer
may reuse `callgraph.Order` as legitimate shared substrate. It may not, and this
is not a judgement call — `corevalidate_endpoint_internal_test.go:107` already
forbids importing `compiler/callgraph` (alongside `check`, `ast`, and
`originvalidate`), and production `corevalidate.go` imports only `core`. Four
import-guard tests exist. The "what structurally prevents drift" sub-question is
therefore already answered by shipped machinery.

---

## Area 2 — OWN-09: which law actually dies

| Option | Description | Selected |
|--------|-------------|----------|
| Delete `computeLoanLastUses`, restructure to lower-then-decide | No ownership decisions during lowering; one decision pass over the assembled `core.Program` | ✓ |
| Leave it as permanent disclosed debt | Status quo — D-08-41 stands, two admission-time entry points forever | |

**User's choice:** Adopt synthesized recommendation (delete).
**Notes:** This **reverses** Phase 08's D-08-41 disposition ("accepted, permanent,
disclosed scope limitation… Landing phase: Not scheduled"), deliberately and in
writing. Rejected alternatives: making `computeLoanLastUses` summary-aware (still
leaves two call sites into one algorithm — the coexistence OWN-09 forbids), and
re-scoping "the intraprocedural law" to mean something narrower (goalpost-moving
with no basis in the code). Cross-ecosystem anchor: rustc's `-Zborrowck=migrate`
two-law period was explicitly temporary and ended at an edition boundary; the
status-quo option is the documented "strangler fig that never finishes."

**Orchestrator verification (→ D-09-07) — the most consequential finding in the
set:** the researcher's central claim was verified correct and reframes the
requirement. `computeLoanLastUses` (at `check.go:3743`, **not** `:2979` as
`08-CONTEXT.md`'s D-08-09 records) calls the *same* `loanLivenessFixpoint`, fed
synthetic `shadow:place:*` operations and a permanent zero-value
`interproceduralSummaryTable{}`. There was never a second law — one algorithm,
two callers, one deliberately summary-blind. OWN-09's "the intraprocedural law is
DELETED" means deleting the early summary-blind call site and its scaffolding, not
deleting an algorithm.

**Document conflict closed:** `REQUIREMENTS.md:173`'s Phase 09 mapping stands as
operative (Phase 08 is shipped at `4ab5c65`; reopening it is not viable); OWN-09's
own requirement text is what gets corrected. This closes D-08-27, which Phase 08's
mid-phase gate explicitly left for a human.

---

## Area 3 — TRU-04 shadow-run gate design

Four sub-decisions were adjudicated.

| Sub-decision | Options considered | Selected |
|---|---|---|
| (a) "recursion" shape | Cycle-peer differential ✓ / deep self-similar chains / defer to Phase 10 | Cycle-peer differential |
| (b) harness vehicle | Extend `session_peer_gate_test.go` ✓ / reuse `generateCallGraphCorpus` as program source ✓ / build a separate harness ✗ | Extend + feed |
| (c) seeded fault | Unexported seam + companion assertion ✓ / mutation matrix (supplement) / external mutation tool ✗ | Seam + companion |
| (d) "same plan" | Test-red-first inside one plan ✓ | Test-red-first |
| (e) counted-work lane | Peer gets its own bound ✓ / share `check`'s ✗ | Own bound |

**User's choice:** Adopt synthesized recommendation.
**Notes:** (a) is a category error as written — Phase 07 refuses cycles before any
liveness derivation runs, so there is no admitted recursive program and a
*liveness* differential over recursion cannot exist; the disposition must be
written into TRU-04's own definition-of-done, not left in a test comment. (c)'s
adversarial core: Knight & Leveson's 1986 correlated-failure result means "a
seeded fault makes them diverge" proves nothing on its own — the discriminating
assertion is that the **unseamed** peer is *unaffected*. (b) rejects a second
harness because two gates that can drift on the meaning of "divergence" is the
same two-derivations-one-truth failure the phase exists to prevent.

**Orchestrator corrections:**
- **(→ D-09-24)** The seam precedent is `verifyCallableRefusalSeam` at
  `check.go:1183` with `TestVerifyCallableRefusalSeamCheckDisabledCorevalidateStillRefuses`
  at `check_test.go:3322` — in **`check`**, not `corevalidate_mutation_matrix_test.go`
  as researched. It disables `check`'s refusal and asserts `corevalidate` still
  refuses: the exact companion-assertion shape, running the *opposite* direction
  from the recommendation. Resolution: land seams in **both** directions.
- **(→ D-09-28)** `qlt02_budget_manifest.json` **already** carries
  `recomputed_work_growth_exponent`, and **both** chokepoints were already widened
  by Phase 08 with `TestGateEligibleMetricSetsAgreeAcrossChokepoints` proving they
  agree — so no re-widening is needed to reuse that name. But an unpriced cost was
  surfaced: since the peer's closure curve is a different curve and duplicate
  `(machine_id, metric)` rows are a validation failure, the peer's own bound needs
  a **distinct metric name**, which means a **third chokepoint widening** with its
  own test obligation — the same trap D-08-32 caught in Phase 08.

---

## Area 4 — D-07-33 narrowed-peer closure scope

| Option | Description | Selected |
|--------|-------------|----------|
| Full closure in Phase 09 | Peer re-derives all four `PublishProblemsFor` classes | ✓ |
| Partial closure | Two body-only classes now; cut foreign-origin-omitted with a named landing phase | (fallback only) |

**User's choice:** Adopt synthesized recommendation (full closure).
**Notes:** The literal OWN-08 reading — "no declared origin" maps to
`core.origin_omitted`, the one class the peer already covers, therefore OWN-08
closes without touching the rest — was considered and **rejected**. `Callable` is
defined project-wide as the full D-04-03 predicate; declaring OWN-08 closed while
three of four classes stay provably vacuous on the peer side would be Key Lesson
3's failure mode and D-03-02 repeating its own M001 history on the same item for
a second consecutive milestone.

Two findings made full closure cheaper than the debt implied: `peerReturnDerivesFromBorrow`
(`corevalidate.go:2099`) is already a forward walk tracking presence only, so the
extension is an access-mode payload, not a duplicate of `originvalidate.go`; and
arity 1 collapses path-understatement to a boolean the peer already computes,
leaving access mode as the only genuinely new derivation. The peer performs a
**containment check** against its own derived facts, not a recomputation — which
is what keeps it independent of `originvalidate` as well as `check`.
`checkForeignOriginOmitted` (a ~55-line function-local backward walk) is the
precedent that the foreign hop does not reopen spike 005's PARTIAL surface.

The partial option is retained as a **fallback with a declared trigger**: if
execution — not plan-time analysis — finds the foreign hop unbounded, split the
debt row, cut that class with a named landing phase, and amend REQUIREMENTS.md to
carve it explicitly out of OWN-08's scope, since OWN-08 is on the never-cut list.

---

## Area 5 — S-008 / QLT-07 status (the phase's planning blocker)

| Option | Description | Selected |
|--------|-------------|----------|
| Run S-008 as a formal spike | Own directory, `SPIKE.md`, competing mechanisms, oracle, injected defect | |
| Replace with a bounded inventory/estimation pre-flight pass | Same bounded intent, pre-registered numeric threshold, no spike apparatus | ✓ |
| Commit QLT-07 blind | Skip the probe, accept estimation risk | |
| Declare QLT-07 stretch now | Apply scope-cut order immediately | |

**User's choice:** Adopt synthesized recommendation (bounded inventory pass).
**Notes:** Verified — `.planning/spikes/` contains only 001–006; S-007 and S-008
have never been run. Two independent reasons drove the substitution: S-008 fails
the project's own spike test (no genuine competing mechanisms, no oracle
independent of a mechanism under test — it is a scoping question, not a validity
question, and forcing spike shape means inventing artificial arms that
`CONVENTIONS.md` itself argues against); and a new spike directory would reproduce
D-08-43's already-open, un-owned registry gap a second time, since
`TestQLT01RegistryCoversAllFiveSpikes` already fails over spike 006's missing row.

**Threshold pre-registered before the count is taken** (this is what keeps the
gate non-decorative): QLT-07 is committed iff the loan-liveness-scoped rows
(03-03/03-04/03-05 only, excluding OWN-04's 03-06/03-07 rows) need **≤1 additional
plan** and open **zero files** OWN-07/OWN-08's work does not already touch.
Otherwise it is re-declared stretch in writing at that moment.

**Closure mechanics:** the archived `03-VALIDATION.md` is never amended in place —
its own loop-carried-liveness clause is structurally un-closable while the
language has no loops, so its boolean can never legitimately flip to `true`. The
subset closes in a new `09-VALIDATION.md` section, with one non-mutating pointer
line appended to the M001 document. Supplemental-filing pattern: reference the
original approval record, never rewrite it.

---

## Area 6 — D-08-26 accepted-program disclosure

| Option | Description | Selected |
|--------|-------------|----------|
| New opt-in runtime vehicle | e.g. `lang check --explain-admission`, mirroring `evidence --validate`'s opt-in `Trace` pattern | |
| Permanent test-of-record, both peers, sets must match | No runtime vehicle ever; peer gets its own closed-set test; cross-peer equality asserted | ✓ |

**User's choice:** Adopt synthesized recommendation (permanent test-of-record).
**Notes:** Verified — the accepted-side consulted field set is a **compile-time
constant**: `check.go:594-609` fires exactly `return.mode` and
`parameters[0].mode` for every declared signature, recorded through the shipped
`interproceduralConsultObserved` seam (`check.go:641`) and already proven closed
by `TestInterproceduralDisclosedFieldSet` (`check_test.go:4898`). A runtime record
would reprint a constant every run at gated cost with zero incremental
information — the opposite of the refusal-side disclosure, which genuinely varies.
The research surfaced a fourth candidate vehicle Phase 08 had not considered (an
opt-in flag mirroring `protocol.go:223-230`'s `Trace *TraceSummary` pattern); it
is technically buildable and still rejected on the constant-payload argument.

**The independence tension was resolved, not dodged:** requiring the two peers'
consulted field sets to *match* is a contract-conformance check ("both read only
what SEM-05 permits"), not a derivation-shape check — it constrains nothing about
worklist-vs-closure, exactly as the shipped `structuralFieldsEqual` and
`ClosureDigest` comparisons already prove outputs without constraining
implementation. It is strictly stronger than verdict-agreement alone, because two
implementations can agree on a verdict while one secretly reads a third field.
This is the independence instrument Phase 08's gate anticipated when it sent both
disclosure questions here together.

---

## Area 7 — OWN-05 call-site convention override

| Option | Description | Selected |
|--------|-------------|----------|
| Negative-space proof only | Grammar/parser test + `core`-level "no field exists" test | |
| Seam-backed refusal control | New unexported seam + mutation-kill, D-07-47/D-08-15 shape | |
| Both, scoped correctly per layer | Negative-space at source + confirm the **existing** closed-set `Mode` decode check at core | ✓ |

**User's choice:** Adopt synthesized recommendation (both, per layer).
**Notes:** Verified — `core.LinearOperation` (read in full) has no per-call
convention-override field; `CalleeID` is its only Phase-07 addition. Both
`corevalidate.derivePeerSignature` and `originvalidate.go:811` hardcode
`Mode: "owned"` with an explicit "today's grammar has exactly one parameter form"
comment. The one core-level field a hostile producer could abuse
(`ParameterContract.Mode`) is already validated against its closed three-value set
at decode time.

A new seam was **rejected**: there is no undecoded slot for an override to hide
in, so it would be dead weight invented to fill a shape the wire format does not
have — the opposite of D-07-47/D-08-15, where a real reachable-looking path
existed and was proven closed. The adversarial point preserved: "not expressible
in source" is **not** the same claim as "not expressible in core," and
`corevalidate`'s whole role is validating a core it did not produce; both claims
get stated separately.

**Honesty finding:** OWN-05 names three derivers, Phase 09 ships two. It must not
be flipped to `Complete` at phase end — split into OWN-05a/OWN-05b or mark the row
partial. Same class of conflict as D-08-27, handled now rather than deferred.
Phase 09 must also **not** extract a shared helper "ready for `interp`" — that is
ARCHITECTURE §4's single-point-of-failure-wearing-two-names anti-pattern.

---

## Area 8 — Diagnostic-code promotion mechanics

| Option | Description | Selected |
|--------|-------------|----------|
| Promote to shared `core.interprocedural_loan_liveness` | Per D-08-21's literal commitment; both peers emit one code | |
| Keep divergent codes; formally supersede D-08-21 | Each peer's code tracks its own derivation; sameness documented as a relationship | ✓ |

**User's choice:** Adopt synthesized recommendation (keep divergent, supersede).
**Notes:** This **supersedes** a commitment Phase 08 made in writing. The trigger
D-08-21 set ("both peers must agree on-code") was never achievable, and D-08-17's
own reasoning already said why: the peer computes through a reachability closure,
not a worklist, so it "cannot fail the same way," and a shared code would assert
an agreement the two mechanisms are structurally incapable of having. The corpus
already demonstrates it — `corevalidate` refuses the relay-escort witness via
`core.move_while_borrowed` while `check` refuses the same program via
`check.interprocedural_loan_liveness`, and `session_peer_gate_test.go` celebrates
that as two honest independent refusals.

Promotion would force one of two bad outcomes: synthesizing `check`'s three-role
cause template onto a closure that has no natural "the call that extended the
loan" (reverse-engineering the worklist into the peer — the literal anti-pattern
this phase exists to prevent), or the same code with different causes, which per
the `Code + Span + Causes` identity fold yields different diagnostic IDs depending
on which layer refused — a worse dispatch contract than two distinct stable codes.

In-tree precedent: `check.call_argument_type_mismatch` / `core.CallArgumentTypeMismatch`
(D-07-46) is already a divergent-code pair for one defect. `core.call_graph_cycle`
is the shared-code precedent and is justified by something absent here — a shared
peer-agnostic witness type. Ecosystem agreement: SARIF separates `ruleId` from
result fingerprinting; Postgres SQLSTATE classes track subsystem.

---

## Claude's Discretion

The developer directed that the synthesized recommendation be adopted for all
eight areas under their standing research mandate. Planner discretion remains
over: plan decomposition and wave ordering (subject to three hard constraints —
build-then-delete, differential-in-the-same-plan-written-red-first, and the third
chokepoint widening landing before any peer manifest row claims to be a hard
gate); whether the peer's derivation lives in `corevalidate.go` or a sibling file;
exact Go identifier names; whether OWN-05 splits into two rows or stays one
partial row; and the order the four `PublishProblemsFor` classes are closed in.

## Deferred Ideas

- `interp`'s third derivation of ownership transfer — Phase 10 (TRU-03), Phase 11
  (NAT-06). Explicitly not prepared for by extracting a shared helper.
- Persistent cross-run summary caching — Phase 11 / QLT-06, per D-08-37, which
  binds the peer as well as `check`.
- A runtime accepted-program disclosure artifact — **closed** as "no vehicle, by
  design," not deferred.
- Promotion of `check.interprocedural_loan_liveness` to `core.*` — **closed** as
  superseded, not deferred.
- `.planning/spikes` registry maintenance (spike 006's missing row, D-08-43) —
  still un-owned; cheap to close if Phase 09 touches the spike table.
- Re-opening the two-bit summary model — when arity widens past 1 (D-07-07) or
  `Result` match arms can independently borrow-vs-move a parameter (Phase 12).

## Scope creep

None. All eight areas clarified how to implement what Phase 09 already scopes; no
new capabilities were proposed or absorbed.
