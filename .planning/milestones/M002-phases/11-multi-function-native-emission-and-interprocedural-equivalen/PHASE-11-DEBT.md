---
phase: 11-multi-function-native-emission-and-interprocedural-equivalen
recorded: 2026-09-11
status: accepted
disposition: planning-time
items: 16
blocking: 0
---

# Phase 11 — Declared Deferred Scope and Phase 10 Carry-Forward Inputs (D-11-02)

**Written:** 2026-09-11, at phase **open** — not at phase close, per D-11-02's
explicit instruction ("record it in PHASE-11-DEBT.md at phase open, not at
phase close"). Shape follows the mechanically-checked debt-register format
`TestDebtRegistersAreWellFormed` (`internal/compiler/session/session_test.go:2613`)
enforces on every `*-DEBT.md` register: an `items:` count matching the
`## Items` table, one `### <ID>` detail section per row, a severity from the
closed vocabulary (blocker, warning, info), and a non-empty landing phase.

This register carries two kinds of rows:

1. **Phase 11's own recorded-not-built decisions** (D-11-02, D-11-07, D-11-11,
   D-11-12, D-11-13, D-11-27, D-11-36, D-11-40, D-11-42) — decisions this
   phase makes and designs but deliberately does not implement, each with an
   explicit closing condition.
2. **Phase 10 carry-forward items** (D-10-C01 through D-10-C05) — the five
   residual trust gaps STATE.md's "Phase 10 carry-forward" section names,
   restated here as **INPUT TO PHASE 11**, not closed history, since Phase 11
   lowers multi-function Lang to C and differential-tests it against the
   Phase 10 oracle whose residual gaps these are.

Also recorded (not a debt row): **Phase 10 ran with
`security_enforcement=true` and produced no `10-SECURITY.md`.** `/gsd-secure-phase 10`
was never run. See the Notes section at the end of this file.

---

## Items

| ID | Source | Threat/Req | Severity | Landing phase | Item |
|---|---|---|---|---|---|
| D-11-02 | 11-CONTEXT.md (D-11-02), RE-DEFERRED 2026-09-12 (D-12-36, `12-01-SUMMARY.md`) | NAT-04..NAT-07 | info | UNOWNED(single-function-emitter-deletion) | THE SIX SINGLE-FUNCTION EMITTERS ARE NOT DELETED THIS PHASE. `emitLinear`, `emitLinearBorrowedByPointer`, `emitLinearBorrowedByPointerPlain`, `emitLinearForeign`, `emitBranch`, and `emitMatch` remain in `cgen` unchanged by Phase 11's multi-function emission work. Their deletion was gated on Q-05's green N=1 convergence differential landing as part of Phase 12's own per-dispatch-site `Result` plans; that gate fired RE-DEFER, not DELETE — see the dated sub-paragraph below |
| D-11-07 | 11-CONTEXT.md (D-11-07) | NAT-04, FFI-01 | warning | P11 | `singleForeignFunction` AND `singleManifestFunction` STAY SINGLE-FUNCTION AND GENERALIZE BY REFUSING a program with two foreign contracts, rather than being rewritten to support multiple. `lang.foreign/0` is NOT widened. These are single-function assumptions the phase's own `len(Functions) != 1` grep inventory does not catch, since they gate on foreign-contract count, not function count |
| D-11-11 | 11-CONTEXT.md (D-11-11) | NAT-05 | info | UNOWNED(emittedattribute-discharge-pair-design) | THE `EmittedAttribute` DISCHARGE-PAIR DESIGN IS DESIGNED AND RECORDED, NOT BUILT. Callee-side `justified_by` plus caller-side `discharged_by`, refused on EQUALITY (never containment), is a complete design this phase writes down but does not implement, because Phase 11 emits zero call-boundary alias attributes by construction (D-11-09) |
| D-11-12 | 11-CONTEXT.md (D-11-12) | NAT-05 | info | UNOWNED(sidecar-schema-verdict-design) | THE SIDECAR-NORMATIVE-PLUS-INLINE-COMMENT CHANNEL, WITH THE `lang.attributes/0` SCHEMA VERDICT, IS DESIGNED AND RECORDED, NOT BUILT. Companion to D-11-11: the channel a discharge pair would be published through is specified, not shipped, for the same reason (zero attributes to discharge this phase) |
| D-11-13 | 11-CONTEXT.md (D-11-13) | TRU-02, OWN-04 | warning | P11 | PRE-EXISTING GAP, NOT INTRODUCED BY PHASE 11. `evidence.go:233` binds `ForeignDigest` only under `hasForeignContract`, so a pure-Lang `restrict` claim is never digest-bound into evidence. Live before Phase 11 opened; closing it is explicitly `lang.attributes/0`'s job, not this phase's emission work |
| D-11-27 | 11-CONTEXT.md (D-11-27) | NAT-07 | info | UNOWNED(d-09-51-negative-control-review) | NAT-07's DESIGN DOES NOT DEPEND ON REVIEWING D-09-51's UNREVIEWED NEGATIVE-CONTROL VERDICT FLIP, because D-11-09 deletes the dependent emission rule the flip would have interacted with. The flip review itself (item 4 of the Phase 10 carry-forward list, D-10-C04 below) remains open and unowned by Phase 11 — recorded here so NAT-07's design is not mistaken for having closed it |
| D-11-36 | 11-CONTEXT.md (D-11-36) | QLT-05 | warning | UNOWNED(flaky-predicate-escalation) | FLAKY-PREDICATE TOLERANCE IN THE REDUCER IS EXPLICITLY NOT BUILT THIS PHASE. If Phase 11's engineered control (the anti-vacuity criterion QLT-05's reducer must satisfy) turns out flaky, that is a criterion-3 problem surfacing in criterion 4, and it MUST be escalated as a defect in the criterion, not silently treated as a reducer tolerance requirement |
| D-11-40 | 11-CONTEXT.md (D-11-40), corrects S-006 (`.planning/spikes/006-interprocedural-liveness-cost-scaling/`) | QLT-06 | warning | P11 | CORRECTION: S-006's 100% / 92% / 43% EVICTION FIGURES ARE AN UPPER BOUND ON A MODEL, NOT A MEASUREMENT. `ClosureDigest` chains over signature summaries only, and never over function bodies (`originvalidate.go:561-590`), so a body-only edit does not move a caller's digest at all. Any artifact that re-quotes these three figures as measured eviction rates MUST NOT do so; state them as the analytical upper bound they are |
| D-11-42 | 11-CONTEXT.md (D-11-42) | QLT-02 | warning | P11 | `SelectLanesForFixture`'s CHANGE-STATE MECHANISM MUST NOT BE EXTENDED to any new Phase 11 lane without the same scrutiny D-11-41/Q-02 applied to the native-differential lane. Deferring a lane because its declared inputs did not move shares D-11-41's exact hole — an undeclared input can move silently while the declared set reports no change |
| D-10-C01 | STATE.md "Phase 10 carry-forward" item 1; PHASE-10-DEBT.md D-10-19-adjacent (`corevalidate.peerDeriveOriginFacts` has no `core.OpCall` case) | QLT-05, TRU-02 | warning | P11 | INPUT TO PHASE 11, NOT CLOSED HISTORY. `corevalidate.peerDeriveOriginFacts` has no `case core.OpCall:` arm. Any function declaring a borrow-returning `PublicOrigin` sourced from forwarding a callee's result is refused as not-`Callable` by `corevalidate` independently of `check`, which admits the same shape with zero diagnostics. Pre-existing; silently affects Phase 08's `relay_depth2_refuse.lang` too. Fail-closed (conservative, not unsound), but it constrained which fixtures Phase 10 plans 10-07/10-08 could express, so any Phase 11 fixture needing that shape hits the same wall |
| D-10-C02 | STATE.md "Phase 10 carry-forward" item 2; 10-REVIEW.md WR-01 | TRU-02, TRU-03 | warning | P11 | INPUT TO PHASE 11, NOT CLOSED HISTORY. `corevalidate.peerCalleeFrameDrained` detects resource escape via return with a single FORWARD pass over `linear.Operations`, correct only under an unstated and unenforced assumption that operations are declaration-ordered by dependency. Its own cited precedent, `peerParameterEscapesOwned`, walks BACKWARD and is order-independent. Disclosed residual risk, not a blocker |
| D-10-C03 | STATE.md "Phase 10 carry-forward" item 3; 10-REVIEW.md WR-02, D-10-19 | TRU-02, TRU-03 | info | P11 | INPUT TO PHASE 11, NOT CLOSED HISTORY. The independence guards disagree: `originvalidate` permits importing `internal/compiler/callgraph`; `corevalidate`'s equivalent forbidden list forbids it. The four peers' independence is enforced by hand-curated per-package import lists, so this asymmetry is defensible by REVIEW, not by MECHANISM. Phase 11's NAT-06 leans on that independence holding |
| D-10-C04 | STATE.md "Phase 10 carry-forward" item 4; 10-02-SUMMARY.md Deviations (D-09-51 fix) | OWN-09, TRU-02 | warning | UNOWNED(negative-control-flip-human-review) | INPUT TO PHASE 11, NOT CLOSED HISTORY. Plan 10-02's D-09-51 fix flipped two negative controls: `negative_control_fails.lang` and `negative_control_infallible.lang` moved from `check.interprocedural_loan_liveness` to `core.callee_not_callable`, because closing the `OpCall` transparent-walk defect also flows through `check`'s SEM-06 Callable gate. The executor documented the reasoning chain and explicitly flagged it for human review. **That review has not happened.** D-11-27 records that NAT-07's design does not depend on it, but the review itself remains open |
| D-10-C05 | STATE.md "Phase 10 carry-forward" item 5 | none — dead code, not a live risk | info | UNOWNED(dead-code-d-10-c05) | INPUT TO PHASE 11, NOT CLOSED HISTORY (informational). `interp.Run`'s `!function.HasClosedBody()` guard is provably unreachable: `corevalidate.Validate` always catches that shape first (found by Phase 10 plan 10-09). Dead defensive code, harmless, recorded here so a future Phase 11 reader does not mistake it for live protection when reasoning about `interp`'s refusal surface |
| D-11-51 | 11-05 Task 2 (`session_phase11_differential_test.go`'s `DiamondSharedLeaf` subtest, discovered against `testdata/phase11/multi_function_diamond_call.lang`) | NAT-06 | warning | UNOWNED(shared-leaf-diamond-event-identity) | NEWLY DISCOVERED, NOT FIXED. A callee invoked from TWO OR MORE distinct static call sites in one run (a genuine "diamond, shared leaf" call graph) emits its own `function.returned` event with the IDENTICAL event ID on every invocation, because both `interp.Run` (`terminalOutcome`) and `cgen.emitProgram` (`emitProgramFunction`) derive that ID from the callee's OWN STATIC OpReturn operation ID — a per-DECLARATION identity, not a per-INVOCATION one. `interp.Run` performs no duplicate-ID validation and returns such a document successfully; `native.go`'s own decode-time validator (`validateExecution`, "duplicate execution event id") correctly refuses to trust it, identically and consistently across `-O0`, `-O3`, and `-O3 -flto` (the refusal is structural, not optimizer-dependent). No prior fixture in this repository ever executed a shared-leaf diamond — `testdata/phase07/deep_diamond_acyclic.lang`'s own shared-leaf diamond is check-only, never driven through `interp.Run` or native emission — so this gap was never previously observable. |
| D-11-52 | 11-05 Task 2 (`session_phase11_differential_test.go`'s `DivergingCallee` subtest) | NAT-06 | info | P12 | NOT EXPRESSIBLE THIS PHASE, extends D-11-02. Every existing `defect` terminator in this codebase is reached through a `core.Match` arm (no arithmetic, no `if`, no loops exist at this maturity to reach `defect` any other way), and `cgen.emitProgram` explicitly refuses any Match-bodied function inside a multi-function program ("multi-function branch bodies are not supported by native emission this phase"). A diverging callee therefore cannot be lowered to native in a multi-function program this phase — a direct structural consequence of D-11-02's own scope decision (the six single-function emitters, including `emitMatch`, are deliberately not deleted or generalized this phase), not a fixture-authoring gap. |

## Spike verdicts

Both of this plan's pre-planning experiments (11-CONTEXT.md § Pre-Planning
Experiments) are committed as single-branch Go tests, not prose. Both
resolved to their BRANCH A outcome.

- **Q-01** — `TestQ01CoreLevelOpCallToOpCopyRewrite`
  (`internal/compiler/corevalidate/corevalidate_opcall_rewrite_spike_test.go`):
  **BRANCH A (accepted)**. Rewriting the FIRST `core.OpCall` operation in
  `testdata/phase07/deep_diamond_acyclic.lang`'s checked `core.Program` to
  `core.OpCopy` in place (ID/SourceID/TargetID preserved byte-for-byte,
  `CalleeID` cleared) is ACCEPTED by `corevalidate.Validate` —
  `Valid == true`. `TestQ01RewrittenProgramStillChecks` proves the rewrite is
  a single-field edit (function count, operation count, and every operation
  ID unchanged before/after), attributing the accept verdict to the `Kind`
  change alone.

  **Consequence for plan 11-08:** the core-level `OpCall`-to-`OpCopy` rewrite
  is viable as designed (D-11-29's central move). QLT-05's reducer can
  express "drop a call site" as this single-field core edit; plan 11-08
  proceeds as the full drop-call-site reducer move, NOT narrowed behind a
  `RefusedShapes()`-gated scope. D-10-C01 above records that this rewrite's
  viability is orthogonal to `peerDeriveOriginFacts`'s missing `OpCall` case
  — the rewrite REMOVES the call before that gap would ever be reached on
  this fixture, rather than closing the gap itself.

- **Q-02** — `TestQ02StaleCgenServesReusedArtifact`
  (`internal/compiler/session/session_phase6_cache_hole_test.go`):
  **BRANCH A (hole reproduces)**. Running `verifyPhase6NativeDifferentialLane`
  once over an unchanged single-function fixture populates the cache under a
  key computed from `phase6ArtifactSpec`'s seven declared inputs. Compiling a
  deliberately perturbed C source (standing in for a `cgen` rewrite, with no
  `cgen` source file touched) into a byte-distinguishable binary, then
  re-deriving `phase6ArtifactSpec` for the SAME unchanged `.lang` source and
  the SAME `clangPath`, yields the IDENTICAL key, and `cache.Consult` reports
  `cache.StatusArtifactReused` and serves the ORIGINAL (now stale) artifact —
  never the bytes the perturbed source would have produced.
  `TestQ02DeclaredInputNamesStillSevenNoCgen` pins `cache.DeclaredInputNames()`
  at exactly its current seven names, none containing `cgen`.

  **Consequence for plan 11-07:** the D-11-41 cache-soundness hole is
  CONFIRMED, not withdrawn. Plan 11-07 ships a real eighth declared cache
  input (a digest over `cgen`'s emitted C, or an equivalent content identity
  for the lowering step) rather than a withdrawal note. `DeclaredInputNames()`'s
  widening from seven to eight names is the visible diff this pinned
  expectation exists to surface.

## Notes

**Phase 10 security posture.** Phase 10 ran with
`workflow.security_enforcement=true` in `.planning/config.json` but produced
no `10-SECURITY.md`; `/gsd-secure-phase 10` was never run. This is not a
Phase 11 debt item with its own `D-11-NN` identifier — it is a Phase 10
process gap, restated here (per STATE.md's own note) because Phase 11's
threat model (this plan's `<threat_model>`, and every subsequent Phase 11
plan's) inherits trust assumptions from the Phase 10 oracle that were never
run through that retroactive security review.

## Detail

### D-11-02 — the six single-function emitters are not deleted this phase

first-recorded: M002

`emitLinear`, `emitLinearBorrowedByPointer`, `emitLinearBorrowedByPointerPlain`,
`emitLinearForeign`, `emitBranch`, and `emitMatch` remain exactly as Phase 10
left them. Phase 11 adds multi-function emission ALONGSIDE these, not instead
of them. Deleting them is explicitly Phase 12's job, gated on Q-05's green N=1
convergence differential (proving the new multi-function path produces
identical output to the old single-function path at N=1) landing as part of
Phase 12's per-dispatch-site `Result` plans — never as a second sweep back
through `cgen` after Phase 11 closes.

**2026-09-12 — RATIFIED: RE-DEFERRED (D-12-36), superseding the text above.**
Phase 12 plan 01 landed the gate: `TestN1ConvergenceDifferential`
(`internal/compiler/cgen/cgen_n1_convergence_test.go`, committed `1bac938`)
drove both the legacy single-function `Emit` path and `emitProgram` over the
same five representative single-function shapes. Measured result:
`emitProgram` refuses four of the five outright — `testdata/phase1/toggle.lang`
(match-only), `testdata/phase3/borrowed_view.lang` (branch/match+linear),
`testdata/phase4/foreign_acquire_one.lang` (foreign-call blocks), and
`testdata/phase4/defect_terminal.lang` (branch with a defect terminator) all
fail with "multi-function branch bodies are not supported" or "multi-function
foreign-call bodies are not supported" — and on the fifth,
`testdata/phase2/owned_transfer.lang`, both paths succeed but their outputs
DIFFER, both textually (a call-boundary attribute comment and an extra
`#include` shift the preamble) and structurally (`emitProgram` writes a
`static` function plus a separate `main`; `emitLinear` writes one inline
`main`). The differential is therefore red-by-measurement, not merely
unproven, and making it green requires porting branch bodies, foreign-call
block bodies, and both by-pointer lowering variants into `emitProgram`,
converging the preamble, and re-pinning the four frozen golden-C digests —
roughly 1,500 lines of `cgen.go` emitter logic. That is not "cheap" under any
reading, which is D-12-36's single named trigger.

The developer ratified **RE-DEFER** at Phase 12 plan 01's Task 2 checkpoint
(recorded verbatim in `.planning/phases/12-result-payloads/12-01-SUMMARY.md`),
reasoning that the "two coexisting laws" tax D-12-31 feared does not
materialize for payload `case` arms specifically: `emitProgram` cannot express
a match body at all, so the payload match surface stays exclusively
legacy-family territory this phase, and the payload arms are still written
exactly once. This **supersedes** the landing-phase text this section
originally recorded — verbatim, from the Items table's original `Landing
phase` cell: *"Phase 12 — a green N=1 convergence differential (Q-05), landing
inside Phase 12's per-dispatch-site `Result` plans, never as a second sweep."*
That commitment is withdrawn. The six emitters remain undeleted with **no
currently-owned landing phase** — see the Items table's updated `Landing
phase` cell above and `PHASE-12-DEBT.md`'s `D-12-36` row for the re-filed
debt item.

### D-11-07 — `singleForeignFunction`/`singleManifestFunction` generalize by refusal

first-recorded: M002

Both helpers keep their single-function assumption. Rather than rewriting them
to iterate over multiple foreign contracts, Phase 11 makes them REFUSE a
program declaring two foreign contracts, closed-set style. `lang.foreign/0`,
the manifest schema, is not widened to describe multiple foreign symbols.
This is deliberately narrower than a general fix: the phase's own 32-site
single-function-guard inventory (ROADMAP.md "Scope input verified at the
planning gate") was built from a `len(Functions) != 1` grep, which does not
surface a foreign-contract-count assumption at all — this row exists so that
gap does not silently reappear as a false "already handled" belief.

### D-11-11 — the `EmittedAttribute` discharge-pair design, designed not built

first-recorded: M002

D-11-09 (11-CONTEXT.md) establishes that Phase 11 emits ZERO call-boundary
alias attributes: with one parameter per function, no globals, no callbacks,
and no address-escaping foreign contracts admitted by `check`, two pointers to
one object cannot exist across a Lang call boundary in M002. A call-boundary
`restrict` attribute would promise about a hazard that cannot exist.

Given that, the callee-side `justified_by` / caller-side `discharged_by`
discharge-pair design (D-11-11), refused on EQUALITY rather than containment,
and the sidecar-normative-plus-inline-comment publication channel with its
`lang.attributes/0` schema verdict (D-11-12), are both fully designed and
recorded in 11-CONTEXT.md but have nothing to discharge this phase. Building
the mechanism against an empty attribute set would be speculative generality
with zero live test coverage. Both land only once a future phase's call
boundary can actually carry an attribute worth discharging.

### D-11-12 — the sidecar-normative-plus-inline-comment channel, designed not built

first-recorded: M002

Companion row to D-11-11: the publication channel a discharge pair would be
written through — a sidecar manifest entry treated as normative, mirrored by
a human-readable inline comment at the emission site — plus the
`lang.attributes/0` schema verdict that channel would validate against, are
both fully specified in 11-CONTEXT.md but not implemented this phase, for the
identical reason D-11-11 states: Phase 11 emits zero call-boundary alias
attributes (D-11-09), so there is nothing for the channel to carry yet.
Building the channel against an empty payload would be speculative
generality with zero live test coverage; it lands once D-11-11's mechanism
has a real discharge pair to publish.

### D-11-13 — the pre-existing `ForeignDigest` gap

first-recorded: M002

`evidence.go:233` binds `ForeignDigest` into evidence only under
`hasForeignContract`. A function whose return carries a `restrict` claim over
a purely-Lang code path (no foreign contract at all) is therefore never
digest-bound. This predates Phase 11 and is not introduced or worsened by
Phase 11's emission work; it is recorded here because `lang.attributes/0`
(D-11-12) is the schema whose eventual job includes closing it.

### D-11-27 — NAT-07 does not depend on the unreviewed D-09-51 flip review

first-recorded: M002

NAT-07's design (the `-O3 -flto` equivalence matrix) does not require
resolving whether Plan 10-02's D-09-51 fix correctly flipped
`negative_control_fails.lang`/`negative_control_infallible.lang`, because
D-11-09 deletes the emission rule that flip would have interacted with. This
row exists specifically to prevent NAT-07's design from being read as having
closed that review — it has not. The review itself is D-10-C04 below, and
stays open and unowned.

### D-11-36 — flaky-predicate tolerance is not built

first-recorded: M002

QLT-05's reducer relies on an engineered control being deterministic. Phase
11 does not build any tolerance mechanism for that control turning out flaky.
If it does, the correct response is to treat it as a defect IN THE CONTROL
(a criterion-3 problem: the control itself is unsound) that has surfaced
during criterion-4 work, escalate it explicitly, and fix the control — never
to quietly add retry/tolerance logic to the reducer that would mask a
genuinely nondeterministic predicate.

### D-11-40 — S-006's eviction figures are a model bound, not a measurement

first-recorded: M002

`ClosureDigest` (Phase 07) chains strictly over callee SIGNATURE summaries,
never over function bodies — confirmed at `originvalidate.go:561-590`. S-006's
100% / 92% / 43% eviction figures were derived from a MODEL of edit
frequency and cache-key sensitivity, not from an instrumented measurement of
real cache behavior under real edits. A body-only edit to a leaf function
therefore does NOT move any caller's `ClosureDigest` at all, which is exactly
the gap that makes the model's figures an upper bound rather than an
observed rate. Any future artifact (a `*-VERIFICATION.md`, a `*-SUMMARY.md`,
this register itself) that re-quotes 100%/92%/43% as measured MUST instead
cite them as the analytical upper bound S-006 establishes.

### D-11-42 — `SelectLanesForFixture`'s change-state extension constraint

first-recorded: M002

Q-02 (this plan, Task 2) demonstrates that an input NOT among
`cache.DeclaredInputNames()`'s seven names (here, `cgen`'s own source) can
change without the declared-input digest set noticing, letting a stale
artifact be served as fresh. `SelectLanesForFixture`'s own change-state
mechanism has the SAME shape of risk: a lane deferred because its declared
inputs did not move is trusting that its declared-input list is complete.
This row is a standing constraint on every future lane addition in Phase 11
and beyond, not a one-time fix: each new lane's declared-input list must be
scrutinized with the same skepticism D-11-41/Q-02 applied here, not assumed
complete by default.

### D-10-C01 — `corevalidate.peerDeriveOriginFacts` has no `core.OpCall` case

first-recorded: M002

Restated from STATE.md's Phase 10 carry-forward item 1, itself carried from
`PHASE-10-DEBT.md`. Any function declaring a borrow-returning `PublicOrigin`
sourced from forwarding a callee's result is refused as not-`Callable` by
`corevalidate` independently of `check`, which admits the same shape with
zero diagnostics. Pre-existing since Phase 09/10; silently affects Phase 08's
`relay_depth2_refuse.lang` too. Fail-closed (conservative, not unsound), but
it constrained which fixtures Phase 10 plans 10-07/10-08 could express.

This plan's Q-01 experiment (`TestQ01CoreLevelOpCallToOpCopyRewrite`) settles
the ADJACENT question of whether the reducer can sidestep this gap entirely
by rewriting the `OpCall` away before `peerDeriveOriginFacts` would ever see
it — verdict: BRANCH A (accepted), so plan 11-08 can. This row remains open
regardless: any Phase 11 fixture that actually NEEDS the borrow-forwarding
shape (rather than rewriting past it) still hits this wall unchanged.

### D-10-C02 — `corevalidate.peerCalleeFrameDrained`'s unstated forward-pass ordering assumption

first-recorded: M002

Restated from STATE.md's Phase 10 carry-forward item 2
(`10-REVIEW.md` WR-01). `peerCalleeFrameDrained` detects resource escape via
return with a single FORWARD pass over `linear.Operations`, correct only
under an unstated, unenforced assumption that operations are
declaration-ordered by dependency. Its own cited precedent,
`peerParameterEscapesOwned`, walks BACKWARD and is order-independent.
Disclosed residual risk, not a blocker; Phase 11's multi-function emission
does not touch this function, but any Phase 11 fixture with an unusual
operation-declaration order could expose it for the first time.

### D-10-C03 — the `callgraph` import asymmetry between `originvalidate` and `corevalidate`

first-recorded: M002

Restated from STATE.md's Phase 10 carry-forward item 3
(`10-REVIEW.md` WR-02, D-10-19). `originvalidate` permits importing
`internal/compiler/callgraph`; `corevalidate`'s own forbidden-import guard
forbids it. The four peers' independence is therefore enforced by
hand-curated, disagreeing per-package import lists — defensible by REVIEW,
not by MECHANISM. Phase 11's NAT-06 (the four-tier interprocedural agreement
matrix) leans on this independence holding; this row flags that the
independence guarantee is weaker than a mechanism-enforced one.

### D-10-C04 — the unreviewed D-09-51 negative-control verdict flip

first-recorded: M002

Restated from STATE.md's Phase 10 carry-forward item 4 (`10-02-SUMMARY.md`
Deviations). Plan 10-02's D-09-51 fix flipped
`negative_control_fails.lang` and `negative_control_infallible.lang` from
`check.interprocedural_loan_liveness` to `core.callee_not_callable`, because
closing the `OpCall` transparent-walk defect also flows through `check`'s
SEM-06 Callable gate. The executor documented the reasoning chain and
explicitly flagged it for human review at the time. That review has not
happened. D-11-27 above records that NAT-07's design does not depend on this
review completing, but the review itself remains OPEN and UNOWNED by any
Phase 11 plan.

### D-10-C05 — `interp.Run`'s unreachable `!function.HasClosedBody()` guard

first-recorded: M002

Restated from STATE.md's Phase 10 carry-forward item 5 (found by Phase 10
plan 10-09). `corevalidate.Validate` always catches an unclosed body first,
so this guard inside `interp.Run` is provably unreachable dead defensive
code. Harmless, but recorded here so a future Phase 11 reader tracing
`interp`'s refusal surface does not mistake this line for live protection
against a shape `corevalidate` has already excluded upstream.

### D-11-51 — shared-leaf diamond call graphs collide on event identity

first-recorded: M002

Plan 11-05 Task 2's own `DiamondSharedLeaf` subtest is the first fixture in
this repository to actually EXECUTE (not merely check) a call graph where
the SAME callee is invoked from two or more distinct static call sites
(`multi_function_diamond_call.lang`: `main` calls both `left` and `right`,
each of which calls the shared leaf `leaf`). Both engines derive an
executed function's `function.returned` event ID from that function's own
STATIC `OpReturn` operation ID (`interp.go`'s `terminalOutcome`,
`cgen_program.go`'s `emitProgramFunction`) — a per-declaration identity,
never a per-invocation one. `leaf` therefore emits the identical event ID
twice in one run. `interp.Run` performs no duplicate-ID validation and
returns the document successfully anyway; `native.go`'s own decode-time
validator (`validateExecution`) correctly refuses to trust any decoded
document carrying two events with the same ID ("duplicate execution event
id"), consistently across all three native tiers (the refusal is
structural, not optimizer-dependent, so there is no CROSS-TIER
disagreement — but no four-tier agreement on a comparable document either,
since three of the four tiers never produce one).

Fixing this for real requires giving each function's events a
per-invocation-unique identity — e.g. threading a call-site-qualified
suffix through both `interp.Run`'s frame stack and `cgen.emitProgram`'s
per-function event-ID derivation, independently, so the two engines' own
identity schemes keep agreeing byte-for-byte. That is a change to
`interp.go` and `cgen_program.go`, neither of which is in plan 11-05's own
declared `files_modified`, and large enough to the two engines' shared
event-identity convention that it needs its own reviewed plan. Plan
11-05's own `DiamondSharedLeaf` subtest asserts the HONEST, consistent
refusal (interpreter succeeds with a duplicate-ID document; all three
native tiers refuse identically) rather than silently weakening the test
or picking an easier "diamond" that never actually re-invokes the shared
leaf, which would misreport a narrower proof as a wider one.

### D-11-52 — a diverging callee is not expressible in a multi-function program this phase

first-recorded: M002

Extends D-11-02. Every existing `defect` terminator in this codebase
(`defect_terminal.lang`, `defect_dies_by_signal.lang`) is reached through a
`core.Match` arm — no arithmetic, no `if`, no loops exist at this maturity
to reach `defect` any other way. `cgen.emitProgram` explicitly refuses any
Match-bodied function inside a multi-function program ("multi-function
branch bodies are not supported by native emission this phase"), so a
diverging callee cannot be lowered to native in a multi-function program
this phase at all. This is a direct structural consequence of D-11-02's
own scope decision (the six single-function emitters, including
`emitMatch`, are deliberately not deleted or generalized this phase), not
a fixture-authoring gap plan 11-05 could have engineered around. Plan
11-05's own `DivergingCallee` subtest is an honest, named `t.Skip`, not a
silently absent case.
