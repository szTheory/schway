# M002 Milestone Shape: Options, Sequencing, and Recommendation

**Project:** Codename Lang
**Milestone under decision:** M002 — Interprocedural Semantic Spine
**Researched:** 2026-09-08
**Overall confidence:** MEDIUM-HIGH (project-internal claims HIGH — read from primary planning
artifacts; external precedent claims HIGH for dates/facts, MEDIUM for the
analogy drawn from them to this project; all "should" statements are labeled
judgment)

---

## 0. Reading base and provenance

Read in full before drafting: `.planning/PROJECT.md`, `.planning/RETROSPECTIVE.md`,
`.planning/MILESTONES.md`, `wiki/ownership-evidence-roadmap.md`,
`wiki/semantic-kernel-contract.md`, `wiki/compiler-and-feedback-latency.md`,
`wiki/compute-efficiency-constitution.md`, `wiki/residual-uncertainty-register.md`.
Skimmed for scope-boundary confirmation: `wiki/modules-architecture-and-live-development.md`,
`wiki/language-lessons.md`. All wiki citations below (Rust RFC 2094, Swift
SE-0304/0377/0390/0504, LLVM docs, Go memory model, etc.) are the project's own
primary-source research, re-cited here where load-bearing for this decision;
I verified their applicability, not their existence.

External sources for the pace/pitfalls and post-mortem section were fetched via
web search on 2026-09-08 (see Sources). Confidence on each is marked inline.

---

## 1. Option space

### (a) Pure interprocedural depth — the charter as written, nothing else

**Scope:** `OpCall` at all six dispatch sites, gated callable ⊆ publishable
(D-04-03); call-graph construction with cycle refusal and a bounded interpreter
call stack; interprocedural loan liveness in `check` and `corevalidate`
(closes D-03-02); cross-function rebuild of Phase 3's exhaustive loan-endpoint
differentials; interprocedural `-O3`/LTO equivalence; storable/matchable
`Result` with payload-carrying alternatives (D-04-30). Nyquist debt (Phases 3,
5, 6) stays carried, restated as still-open in the M002 charter rather than
silently dropped.

**What it proves:** the semantic spine generalizes past one function body — the
single structural gap M001 shipped with. It answers the highest-blast-radius
open question: does independent re-derivation, loan liveness, and native
equivalence survive a call boundary, or does the M001 architecture only work
because it never had to cross one.

**What it defers:** validation debt closure, modules/separate compilation as a
first-class unit, and any agent-loop investment beyond what calls force.

**Risk:** MEDIUM. This is exactly the charter the retrospective and PROJECT.md
already committed to (`PROJECT.md:185`, "M002's lead charter"). The risk is not
scope-fit, it's execution risk: RETROSPECTIVE.md's own numbers show
interprocedural-shaped work (Phase 3, "moved from single-consumer changes to
five-consumer coordinated changes") already cost 70-95 min/plan versus ~10
min/plan for single-consumer work, and cost 3 corrective plans for one
derivation (`originvalidate.RecomputeOrigin`). `OpCall` at six dispatch sites is
structurally *more* consumers than that, not fewer.

**Case FOR:** it is the smallest coherent unit that keeps the semantic-kernel
promise honest — a language whose core claim ("every semantic guarantee holds
across function boundaries") cannot be asserted with one function per program.
Minimizes scope creep; every plan traces to one of six enumerated
deliverables. Matches `wiki/residual-uncertainty-register.md`'s explicit stop
rule: "the next unknown-reduction work is implementation, measurement, and
counterexample reduction — not more unconstrained enumeration."

**Case AGAINST:** ships M002 with *more* validation debt than M001 if Nyquist
gating is not itself re-instated as a phase gate — repeating the exact failure
mode RETROSPECTIVE.md names ("Validation that is not gated slips... tracked
exactly the phases where it was a gate condition"). Also: cross-function loan
liveness is D-03-02's exact debt item; closing it *is* the milestone, so this
option has almost no slack if the riskiest assumption (that the existing
liveness/origin architecture generalizes without a redesign) turns out false.

---

### (b) Charter + validation-debt closure (Nyquist for Phases 3, 5, 6)

**Scope:** (a) plus retroactively closing Nyquist validation for the three
M001 phases that shipped non-compliant.

**What it proves:** the whole shipped spine — not just the newly-built
interprocedural half — meets its own evidentiary bar. Removes the asymmetry
where M002's brand-new interprocedural code is validated to a standard the
Phase 3/5/6 code it *depends on* was not.

**What it defers:** nothing new semantically; this is pure debt-service.

**Risk:** LOW-MEDIUM technically (it's retrospective work against a stable
target), but HIGH schedule-displacement risk: it competes for the same
attention budget as the six-dispatch-site `OpCall` work, and MILESTONES.md
shows M001's own validation slipped specifically *because* it was queued
behind feature work every time.

**Case FOR:** "a control you have never seen fail is a claim" (RETROSPECTIVE.md
Key Lesson 1) applies as much to unvalidated phases as to unmutation-killed
controls — Phases 3, 5, 6 currently carry claims, not evidence, and Phase 3 is
precisely the phase M002 is about to build directly on top of (loan liveness,
origins). Closing Phase 3's Nyquist gap *before* rebuilding its interprocedural
half is arguably a prerequisite, not a nice-to-have: you should not extend a
derivation whose own local correctness hasn't been checked to the standard the
project claims to hold itself to.

**Case AGAINST:** it is backward-facing work on a codebase that is about to
change underneath it. Nyquist-validating Phase 3's *intraprocedural* loan
liveness the same milestone it gets extended to interprocedural liveness risks
validating a soon-obsolete shape, wasting the validation the moment D-03-02
closes. Better sequencing (see §3) is to fold the *relevant* slice of Phase 3
Nyquist closure into the phase that rebuilds loan liveness (the derivation is
being touched anyway — closing its gate there is nearly free) and explicitly
defer Phase 5/6 Nyquist to a later milestone when their surface is stable
again.

---

### (c) Charter + separate compilation / modules

**Scope:** (a) plus landing `wiki/modules-architecture-and-live-development.md`'s
module/component/package split — explicit exports, acyclic dependency graph,
compact exported semantic summaries consumed instead of transitive source.

**What it proves:** that calls are meaningful *across* compilation units, not
just across functions in one file — the compiler-and-feedback-latency
document's central architectural claim ("a consumer should read a dependency's
exported semantic summary, not parse the entire transitive source graph").

**What it defers:** nothing from the charter, but adds an entire new subsystem
(export surfaces, acyclic graph verification, compact interface materialization)
in the same milestone as the interprocedural rebuild.

**Risk:** HIGH. This combines the milestone's already-highest-risk item
(cross-function loan liveness rebuild, closing the one deliberately-carried
debt item) with a second large, independently risky subsystem. RETROSPECTIVE.md's
lesson about Phase 3 cost growth was specifically about *coordinating five
consumers of one concept* (loan liveness) — modules would add a sixth-plus
axis (compilation-unit boundary) to every one of those consumers simultaneously.

**Case FOR:** "call boundary" and "module boundary" are the same
architectural question asked twice if you build calls first and modules
second — you'll rebuild the interprocedural call-graph and origin work again
once real separate compilation exists, because M001's Phase 3 origin work was
explicitly "public origins... verified against the core" only within one
compilation unit (`corevalidate` proof). Building them together means one
rebuild, not two.

**Case AGAINST:** this is precisely the failure pattern RETROSPECTIVE.md
diagnoses M001 avoiding on purpose: "Deferring `OpCall`... was correct for the
risk budget" — landing a new dispatch mechanism *alongside* another
large structural change (there: alias facts, sanitizers, reducer, QLT-01
registry; here: modules) "matched the fingerprint of the failures that cost
Phases 2-4 extra remediation rounds." Modules is exactly the kind of
non-charter addition that inflates plan count under review pressure (the
3→7→10→13→14→15 plan-count drift RETROSPECTIVE.md records). The
`residual-uncertainty-register.md` stop rule also does not rank modules as
NOW-tier: Frontier 1 ("compiler architecture and feedback economics", where
module boundaries live) is explicitly *after* Frontier 0 (semantic kernel,
where interprocedural ownership lives).

---

### (d) Charter + agent-loop leverage (better repair, evidence reuse, faster warm loop)

**Scope:** (a) plus extending Phase 6's agent surface (`lang explain`, cause
DAG, `lang-repair`, changed-risk lanes, budget manifest) to cover the new
interprocedural defect classes: cross-function borrow violations, call-cycle
refusals, interprocedural `Result` payload mismatches.

**What it proves:** that the agent-facing contract — the actual stated Core
Value of the project ("the shortest reliable path from intent to sound,
reproducible evidence... without wasting iteration time") — holds up once
errors can originate in a different function than the one being edited, which
is a materially different diagnostic-locality problem than M001 ever faced.

**What it defers:** validation debt, modules.

**Risk:** MEDIUM. Bounded because it extends existing infrastructure (Phase 6
patterns: 5 defect injectors, held-out corpus, cause DAG) rather than building
new architecture, but every new `OpCall`-caused defect class needs its own
injector, corpus, and mutation-kill (RETROSPECTIVE.md Lesson 1) — this is
real, multiplicative work, not decoration.

**Case FOR:** this is the option most directly aligned with the *stated*
Core Value rather than the architectural charter. A milestone that lands
interprocedural calls but produces opaque or distant diagnostics when a loan
escapes across three functions technically satisfies the charter while
violating the product thesis. `wiki/compute-efficiency-constitution.md`'s
efficiency review checklist explicitly asks "Can an agent/human query why the
work happened?" for every material feature — interprocedural calls is about as
material as a feature gets.

**Case AGAINST:** it risks scope-inflating the exact deliverables that are
already the highest-risk item in the milestone (see (a)'s risk analysis).
Diagnostics *for* interprocedural defects cannot be well-designed until the
interprocedural defect *taxonomy* is known from having built the checker/
`corevalidate` work — building them in parallel from day one means redesigning
the diagnostic surface mid-flight when the taxonomy turns out different than
guessed. Better as a dependent later phase than a co-equal charter addition.

---

### (e) Charter narrowed to a hard proof core, with a deliberately experimental high-leverage spike alongside

**Scope:** Narrow the charter's guaranteed deliverable to the smallest set that
*must* ship for D-03-02 to close and for the milestone's name ("Interprocedural
Semantic Spine") to be true: `OpCall` dispatch + call-graph + cross-function
loan liveness + interprocedural `-O3` equivalence. Explicitly time-box
`Result` payload work (D-04-30) and Nyquist closure as stretch/fallback items
that can slip to M003 without re-scoping the milestone name. Run one bounded,
declared, falsifiable spike alongside — e.g., a Cranelift/QBE dev-backend
timing probe (Frontier 3's still-open "development backend" question) or a
recursive-call / cycle-refusal stress corpus at scale — following the spike
discipline already established in `.planning/spikes/`.

**What it proves:** the hard proof core (interprocedural soundness), while
using the project's own established spike methodology (bounded, falsifiable,
disposable, gated) to retire one *additional* piece of uncertainty from the
residual-uncertainty register without inflating the milestone's committed
surface.

**What it defers:** everything not on the critical path to "calls are sound
across function boundaries, natively and in the interpreter." Result payloads
and Nyquist debt become explicitly negotiable, not silently dropped —
consistent with RETROSPECTIVE.md Lesson 4 ("deferring a large structural
change is legitimate and should be written down at the moment it is decided").

**Risk:** LOW-MEDIUM. This is the only option that actively reduces M002's own
risk profile relative to the charter-as-written, by explicitly building slack
into the plan-count budget the retrospective shows tends to be consumed by
gap-closure work anyway (3→15 plan drift under review pressure). The spike is
risk-additive but bounded by construction (spikes 001-005 all closed with a
gate and a result, never expanded into the compiler — see
`ownership-evidence-roadmap.md`'s "per-spike execution checklist" and "stop
when the gate is answered").

**Case FOR:** it is honest about M001's own lesson that phase counts drift
upward under review pressure specifically when gap-closure work is absorbed as
silent scope rather than declared and renegotiated. Declaring Result payloads
and Nyquist as explicitly-negotiable *before* the milestone starts means that
if (when, per Phase 3's history) cross-function loan liveness needs three
corrective plans, that absorbs budget from declared-optional work instead of
either blowing the milestone timeline or getting silently cut under pressure
near the end.

**Case AGAINST:** it under-delivers against the milestone charter as already
written into `PROJECT.md`, requiring a renegotiation with the user before
starting — a cost the other options don't pay. It also risks reading as
hedging rather than commitment on the one item (D-03-02) the project has
already flagged as the single carried debt item; shipping M002 without closing
it a second time would be a genuine credibility cost.

---

### (f) Better option: charter as the spine, Nyquist closure folded into the touching phase, agent-loop extension as a dependent final phase, modules explicitly excluded and named as M003's lead candidate

This is not a fully distinct scope so much as a specific ordering/inclusion
policy across (a)-(e), and it's the one I recommend (see §6). Summarized here
for completeness of the option space; full sequencing in §3.

**Scope:** the charter (a) as the spine; fold Phase-3-relevant Nyquist closure
into whichever phase rebuilds loan liveness rather than as a separate pass
(cheap when already touching the file, expensive as a standalone pass later —
this is (b)'s own "before AGAINST" case, resolved); land the agent-loop
extension for interprocedural defect classes as the *last* phase, once the
interprocedural defect taxonomy is actually known (resolves (d)'s AGAINST);
explicitly exclude modules (c) and Phase 5/6 Nyquist closure from M002, naming
both in PROJECT.md's Out of Scope with the reasoning stated at decision time
(per Lesson 4).

**What it proves:** the interprocedural spine, cleanly, with the debt this
milestone is *responsible for* (interprocedural loan liveness) closed to the
same evidentiary standard as the rest of the milestone rather than carried
again.

**What it defers, explicitly, in writing:** modules/separate compilation
(named M003 candidate), Phase 5/6 Nyquist closure (named debt, explicit
carry), any agent-loop work beyond what the new defect classes strictly
require.

**Risk:** MEDIUM, same technical risk profile as (a) — the core rebuild is the
same hard problem regardless of packaging — but LOWER schedule risk than (a)
because it pre-negotiates what happens when (not if, per Phase 3's own
history) the interprocedural rebuild needs more plans than budgeted.

---

## 2. Multi-lens evaluation

Evaluated against (a) charter-only, (c) charter+modules, (e) narrowed-core+spike,
and (f) the recommended shape, since these four are the genuinely distinct
strategic postures (b) and (d) are absorbed into (f)'s ordering policy.

| Lens | (a) charter-only | (c) charter+modules | (e) narrowed+spike | (f) recommended |
|---|---|---|---|---|
| **Language designer** — is the semantic core still small and coherent? | Yes. One new `OperationKind`, no new kernel law. Matches `semantic-kernel-contract.md`'s own scoping discipline (kernel owns only behavior whose disagreement changes program meaning; modules are explicitly listed as *outside* the kernel table). | Strained. Modules pull in export/privacy/acyclicity semantics that are real new kernel-adjacent laws (structural-port conformance, witness requirements) — not wrong to eventually build, but they are a second semantic surface landing in the same milestone as the first cross-function proof. | Yes, more so than (a) — narrowing to only what's required for D-03-02 keeps the "coherent core" claim tightest. | Yes — same as (a), with explicit scope discipline that keeps future additions (modules) named rather than crept in. |
| **Compiler engineer** — is build order sane, blast radius bounded? | Sane by RETROSPECTIVE.md's own evidence: this is the exact next item in the vertical-slice order (call is the one thing every layer already assumes will eventually exist — six dispatch sites are already named). Blast radius is the known cost: five-plus consumer coordination, Phase 3-shaped. | Blast radius compounds: coordinating five-plus consumers of *loan liveness* while simultaneously coordinating consumers of a *new module boundary* is a cross-product, not a sum, of the two risks RETROSPECTIVE documents separately (Phase 3's five-consumer cost, and D-02-03's "fixing one of N independent peers is not fixing the item" — modules would add peers to fix, not just work to do). | Bounded, deliberately tighter than (a). The spike is walled off by the project's own spike discipline (disposable, gated, time-boxed). | Sane, same order as (a); Nyquist-folding reduces total distinct build passes over the loan-liveness code (touch it once, not twice). |
| **Verification engineer** — does the evidence actually prove what it claims? | Mostly yes for the new work (independent re-derivation, mutation-kill discipline is contractual per RETROSPECTIVE Lesson 1-3), but the milestone *as a whole* would claim interprocedural soundness while resting on three still-unvalidated phases underneath — a claim, not fully evidence, exactly the shape Lesson 1 warns about. | Same gap as (a), plus a second: module-boundary export-surface claims ("consumer never touches a body field," M001's Phase 3 already demonstrated this within one compilation unit) would need re-proving under real separate compilation, and there's no budget slack to do that rigorously alongside the loan-liveness rebuild. | Tightest: narrowing the guaranteed core to exactly what's independently re-derivable and mutation-killed within budget is the most honest evidentiary posture of the four. | Best: closes the specific Phase-3 gap this milestone's work depends on, which is the one instance where "evidence proves what it claims" is genuinely load-bearing (you cannot honestly claim interprocedural loan-liveness proof if the intraprocedural derivation it extends was never Nyquist-validated). |
| **Security engineer** — new trust boundaries and escapes? | Calls are a new trust boundary by construction: callable ⊆ publishable is a new admission gate (D-04-03), and call-graph cycle refusal is a new fail-closed device that needs its own negative control (per the project's "escapes are declared and executable, not hypothetical" pattern — `escape:coordinated-source-to-core-false-claim` is the model to replicate for calls). | All of (a)'s new surface, plus a second: module boundaries are precisely where `wiki/semantic-kernel-contract.md`'s nominal-authority rules ("structural conformance never grants authority") get tested for real, and where accidental widening of a public surface across a compilation-unit boundary is easiest to miss under time pressure. | Same new surface as (a), scoped tighter — fewer new interacting boundaries to reason about at once. | Same as (a); explicitly deferring modules means the module-boundary trust question gets its own dedicated milestone attention later instead of sharing review bandwidth with the call-boundary question now. |
| **SRE / release engineer** — CI cost, determinism, reproducibility | Manageable: extends existing verify-phaseN.sh gate pattern, existing five-axis comparator now needs an interprocedural corpus, budget-manifest lane already exists to catch cost regressions. | Higher CI cost and higher risk of an under-gated seam: two new large subsystems (calls, modules) both need their own gate scripts and both compete for the same "mandatory mid-phase gate" review bandwidth that Phase 3 already showed straining under five-consumer coordination alone. | Lowest CI/cost risk of the four — narrower committed surface, bounded spike cost by construction. | Comparable to (a); folding Nyquist into the touching phase is *cheaper* in CI terms than a separate validation pass over already-shipped code (no need to reconstruct old phase context). |
| **AI agent consumer** — faster, more precise feedback loop? | Neutral-to-negative in the short run: new interprocedural defect classes (cross-function borrow escape, cycle refusal, call-graph diagnostics) will initially lack the tuned cause-DAG/repair support Phase 6 built for intraprocedural defects, until a follow-up closes that gap. | Same gap as (a), plus module-boundary defects (illegal cross-module borrow, cyclic dependency) also lack agent-facing repair support — a wider gap, opened in the same milestone. | Same short-run gap as (a); spike doesn't address it either way. | Best of the four for the *stated* Core Value: explicitly scheduling the agent-loop extension as the milestone's own final phase (not deferred indefinitely, not built blind before the defect taxonomy is known) closes the gap the same milestone it's opened, once it's cheap to do correctly. |
| **Human auditor** — bounded audit time? | Bounded: same architecture (one new `OperationKind`, existing dual-validator pattern) an auditor already knows how to read from M001. | Not bounded the same way: an auditor now needs to hold two new mental models simultaneously (call semantics + module semantics) to audit any one PR, which is exactly the kind of coordinated-review load that RETROSPECTIVE.md's plan-count drift (13-15 plans, "review pressure") correlates with. | Most bounded: narrower committed surface is easier to audit exhaustively within a fixed time budget. | Bounded, same as (a); explicit named deferrals (modules, Phase 5/6 Nyquist) reduce what the auditor has to hold in their head as "still open" versus "silently missing." |
| **Product/strategy** — what makes this language worth existing over Rust/Zig/Swift? | Directly on-thesis: Rust's NLL took years to reach the ergonomics Lang is trying to get right the first time (see §4); proving the *interprocedural* half of that promise, honestly and with independent re-derivation, is close to the entire differentiated bet of the project (AI-authored, human-audited, evidence not vibes). | Modules matter eventually (Rust/Swift/Zig all have real module stories the project will need to beat or match) but are not the differentiator — ownership+evidence is. Spending this milestone's scarce coordination budget on modules dilutes the one claim competitors can't easily match. | On-thesis and disciplined, but under-claims relative to what the project has already told itself (and, per PROJECT.md, the user) it will ship this milestone — a real cost if credibility with stated commitments matters. | Best balance: ships the full differentiated claim (interprocedural soundness) while explicitly protecting the evidentiary rigor that *is* the differentiator, and defers modules to when it can get the same rigor rather than diluted rigor now. |

**Where the lenses disagree, named explicitly:**

- **Verification engineer vs. Product/strategy on (a):** verification wants
  Nyquist closed before claiming milestone-wide soundness; product/strategy is
  satisfied shipping the charter as originally scoped, on schedule, since the
  Nyquist gap is already disclosed debt, not a secret. This is resolved, not
  ignored, by (f): fold only the *directly load-bearing* slice of Nyquist
  closure (Phase 3, since interprocedural liveness extends it) into the
  build, and leave Phase 5/6 as still-disclosed, not silently expanded scope.
- **Compiler engineer vs. Language designer on (c) modules:** the compiler
  engineer's "build order sane" case for landing modules alongside calls (avoid
  rebuilding the call-graph/origin work twice) is real and not wrong — but the
  language designer's coherence concern and the SRE's CI/review-bandwidth
  concern outweigh it *for this specific milestone*, because M001's own
  history (Phase 3 cost inflation from *one* new cross-cutting concept) is
  direct evidence against stacking two at once. This is a real tradeoff, not a
  clean win for either side — recorded as a deliberate risk acceptance, not
  a dismissal.
- **AI agent consumer vs. Verification engineer on when to build agent-loop
  support:** the agent-consumer lens wants repair/diagnostic support built
  *alongside* the semantic work so nothing ships without a feedback loop;
  verification (and compiler engineering, via the "you don't know the defect
  taxonomy yet" argument) wants it built *after*. (f) resolves this in favor of
  sequencing it last within the same milestone rather than choosing one lens
  over the other permanently.

---

## 3. Sequencing analysis — recommended option (f)

Riskiest assumption named per phase; the gate is the mechanism that would
catch that assumption failing, following the project's own established
pattern (mandatory mid-phase gates, shadow-runs before deletion, mutation
kills on every differential).

### Phase 1 — Call-graph construction and cycle refusal

**Goal:** Build the call graph over the existing typed core (six dispatch
sites get `OpCall` as a real, structurally admitted `OperationKind`); refuse
cycles with a stable diagnostic; bound the interpreter call stack.

**Unblocks:** every other phase — nothing interprocedural can be checked,
validated, or interpreted without a call graph existing first.

**Why here:** it is the one piece of infrastructure every later phase
depends on and the one piece that has no dependency on the harder loan-liveness
question — cheapest correct place to start, matching the project's own
"vertical slice, cheapest falsifiable question first" discipline.

**Riskiest assumption:** that the existing per-function CFG/terminator
registry (`checkReleaseOrder`'s backward walk, the terminal-block registry from
Phase 4) generalizes to a call-graph walk without a structural rework — M001's
graph walks (release-order rederivation, cause DAG) were all *intraprocedural*.

**Catching gate:** a cycle-safe, visited-set-guarded call-graph walk proven
against a hand-authored corpus including at least one deliberately cyclic and
one deliberately deep (near-bound) call chain, with the refusal path
mutation-killed (an un-guarded walk must be shown to hang or blow the bound on
the same corpus before the guard is trusted) — same discipline as
`checkReleaseOrder`'s D-04 cyclic-refusal work in M001.

### Phase 2 — Interprocedural loan liveness in `check` (closes half of D-03-02)

**Goal:** Extend `loanLivenessFixpoint` (the sole liveness law since M001's
05-02 shadow-run retirement of the old derivation) across call boundaries in
the fast admission-deciding path.

**Unblocks:** Phase 3 (the `corevalidate` peer), Phase 4 (differentials can't
run without both peers agreeing on what "loan ends here" means across a call).

**Why here:** this is the actual charter centerpiece and the single carried
debt item (D-03-02); it must land before its independent-checker peer so the
peer has a stable target to validate against, matching M001's own order
(`check` before `corevalidate` in every phase).

**Riskiest assumption:** that a monotone intraprocedural-shaped dataflow
extends to interprocedural liveness without becoming the "whole-program
inference" pattern `compute-efficiency-constitution.md` explicitly warns
against ("Avoid whole-program inference... whose cost or invalidation radius is
difficult to bound") — i.e., that this can stay a bounded, summary-based
analysis (function-level origin/ability summaries, not full inlining or
re-walking callee bodies from every call site).

**Catching gate:** a scale/cost measurement (same shape as M001's 230,692-
comparison shadow run) on a corpus with realistic call-graph fan-out, checked
against the `compute-efficiency-constitution.md` budget discipline — if cost
grows worse than linear in call-graph size, that is the falsification signal
requiring redesign before Phase 3 begins, not after.

### Phase 3 — Independent re-derivation in `corevalidate` + fold in Phase-3-relevant Nyquist closure

**Goal:** `corevalidate`'s peer derivation (reachability-closure-plus-reduction,
per M001's established pattern of *never* sharing `check`'s worklist
implementation) extended across calls; simultaneously close the Nyquist
validation gap for the *loan-liveness-adjacent* slice of the original Phase 3
work this phase is directly modifying.

**Unblocks:** Phase 4's differential (needs two independently-derived answers
to compare) and the Nyquist debt-service commitment from (f)'s scope decision.

**Why here:** RETROSPECTIVE.md's D-02-03 lesson ("fixing one of two peers is
not fixing the item") means Phase 2 alone is not "closing D-03-02" — only
having both independent peers agree closes it, exactly as M001 required for
every other trust crossing. Folding Nyquist closure in here is free-ish: the
file is already open, the mental model is already loaded, and it avoids a
second, more expensive standalone pass later (resolves option (b)'s AGAINST
case).

**Riskiest assumption:** that `corevalidate`'s independent derivation strategy
(reachability closure) generalizes to cross-function reachability without
becoming quadratic-or-worse in call-graph size — this is precisely the shape of
bug class D-02-03/D-03-01 already caught once in M001 ("a quadratic cost fix
landed in the wrong half... the fix never reached the admission-deciding
path").

**Catching gate:** a counted-work lane (per-operation, not per-wall-clock,
following M001's established practice) proving linear-or-declared-bounded work
in call-graph size, PLUS a deliberate seeded endpoint-level fault (same pattern
as Phase 3's exhaustive-path-enumeration oracle in M001) proving the two peers
actually diverge when one is wrong — a control that has never been made to
fail is a claim, not evidence (Lesson 1).

### Phase 4 — Interprocedural differential + native `-O3`/LTO equivalence

**Goal:** Rebuild Phase 3's exhaustive loan-endpoint differential across
function boundaries; extend the five-axis interpreter/`-O0`/`-O3` comparator to
call-containing programs; prove `-flto`-non-inert `-O3` equivalence for calls
specifically (the exact claim M001 explicitly could not make).

**Unblocks:** this *is* the milestone's headline claim ("Interprocedural
`-O3` equivalence... M001 could not claim because it ships without calls" —
PROJECT.md:75). Nothing downstream needs to unblock; this is where the
milestone's central risk gets adjudicated.

**Why here:** it depends on Phases 1-3 (call graph, both liveness peers) being
done and stable; it is deliberately placed before the lower-risk `Result`
payload work so the milestone's hardest, most irreversible claim gets tested
with maximum remaining schedule slack, not minimum (mirrors M001's own
Phase 5 placement of the native-equivalence proof after the ownership/FFI
machinery, not before).

**Riskiest assumption:** that cross-function inlining, devirtualization, or
alias-analysis decisions at `-O3`/LTO do not introduce a *new* class of
divergence beyond the intraprocedural `restrict`/false-no-alias class M001
already found and fixed (Phase 5's `control:alias.false_no_alias`) — LTO
crossing function boundaries is exactly where a frontend's no-alias/`restrict`
claims are most likely to overreach, per the project's own citation of LLVM's
UB manual on optimizer-visible attribute obligations.

**Catching gate:** at least one deliberately engineered interprocedural
adversarial program per new alias/inlining-sensitive attribute the frontend
now emits across a call boundary, proven to produce the same
`interpreter == -O0 != -O3` divergence shape M001's Phase 5 used as its
positive control for the intraprocedural case — i.e., replicate the
`false_no_alias` mutation-kill methodology one call-boundary later, not assume
it still applies.

### Phase 5 — Storable/matchable `Result` values with payload alternatives (D-04-30)

**Goal:** Land the deferred `Result` payload work from M001's Phase 4.

**Unblocks:** nothing structurally downstream in M002 — this is deliberately
placed after the hard interprocedural proof, not interleaved with it, because
it is independent surface-area work that does not touch the call graph, loan
liveness, or native equivalence machinery.

**Why here:** per (e)'s reasoning, this is the item with the most legitimate
claim to being negotiable/movable if Phases 2-4 run long (which, per Phase 3's
own history, is the base rate to expect, not the exception) — placing it last
means a schedule slip here does not cascade into the milestone's harder,
already-scheduled-tight claims.

**Riskiest assumption:** that payload-carrying alternatives interact cleanly
with the now-interprocedural ownership model (a `Result` payload returned
across a call boundary is exactly a borrowed/owned-origin question) rather
than being a genuinely separate, previously-deferred concern — if it turns out
D-04-30 secretly depends on interprocedural origins in a way not yet
discovered, that's a real risk to surface here, not silently absorb.

**Catching gate:** an explicit compatibility check — does a `Result<T,E>`
payload's origin/ownership story reduce to already-proven Phase 2-4 machinery,
or does it require new interprocedural rules? If the latter, this phase
re-scopes (possibly slips to M003) rather than silently expanding Phase 2-4's
already-tight budget.

### Phase 6 — Agent-loop extension for interprocedural defect classes

**Goal:** Extend Phase 6 (M001)'s cause-DAG, repair applicability, and
changed-risk lane machinery to the new defect classes this milestone
introduces: cross-function borrow escape, call-cycle refusal, interprocedural
`-O3` divergence, `Result` payload defects.

**Unblocks:** closes the AI-agent-consumer lens gap named in §2 — the actual
Core Value claim ("shortest reliable path from intent to sound, reproducible
evidence... without wasting iteration time") is not fully true for
interprocedural code until this phase lands.

**Why here, last:** the defect taxonomy is only known once Phases 1-5 exist;
building repair/diagnostic support blind (option (d), done in parallel)
risks building the wrong cause-DAG edges or the wrong repair applicability
classes, exactly the kind of premature-specificity waste
`compute-efficiency-constitution.md`'s efficiency checklist warns against.

**Riskiest assumption:** that the existing cause-DAG/repair architecture
(bounded depth, `truncated:explain.*` codes, rustfix-style `Applicability`)
extends across a call boundary without a redesign — a borrow that escapes
three functions away may need a materially longer or differently-shaped cause
chain than anything Phase 6 (M001) was built and bounded against.

**Catching gate:** the existing DX-04-style held-out/derivation-split fixture
corpus methodology (5 M001 defect injectors, held-out corpus, fail-closed
marker guard) replicated for at least 2-3 new interprocedural defect classes,
proving the repair driver actually reaches and fixes the new defect shape
through the JSON protocol alone — not merely that the diagnostic fires.

### Phase 7 (optional, only if scheduled a bounded spike per §5) — one high-leverage experiment

Run alongside Phase 1 or Phase 4 (see §5 for which), not as a blocking
sequential phase — genuinely parallelizable because it does not gate any
charter deliverable.

**Parallelization note:** Phase 1 (call graph) and a spike (§5) can run
concurrently since the spike, by construction, doesn't touch production code
paths the charter depends on. Phases 2 and 3 cannot run in parallel with each
other (Phase 3 depends on Phase 2's liveness law existing as a stable target
to validate against, mirroring M001's `check`-before-`corevalidate` order).
Phase 6 cannot start meaningfully in parallel with Phases 1-5 (its taxonomy
input doesn't exist yet) but its *scaffolding* (extending the injector
framework, held-out corpus infrastructure) could begin once Phase 4 lands,
overlapping with Phase 5.

---

## 4. Pace vs. quality

### Practices that let this project go fast without losing rigor

1. **Fold validation into the phase that touches the code, not a later pass.**
   Directly answers "validation that is not gated slips" — the gate cost is
   amortized against work already being done, so it doesn't compete for
   separate schedule; this is (f)'s Phase 3 design.
2. **Budget the mutation-kill into the plan that introduces the control, not a
   later gate** (RETROSPECTIVE.md Lesson 1, stated as a rule already). For
   M002 specifically: every new differential (call-graph cycle refusal,
   interprocedural liveness fixpoint, cross-function `-O3` equivalence) needs
   its mutation-kill written in the *same plan* that introduces it, not
   deferred to a mid-phase gate.
3. **Retire the old law the same phase you introduce the better one**
   (Lesson 2). For M002: if interprocedural liveness requires a new fixpoint
   shape rather than a pure extension of `loanLivenessFixpoint`, the old
   intraprocedural-only law must be shadow-run-and-deleted in the same phase
   it's superseded, not carried for "roughly two phases" the way M001's
   two-derivation period ran (03-03 to 05-02).
4. **Fix at both peers, in the same commit family, or the item isn't fixed**
   (Lesson 3 / D-02-03's exact failure). Every M002 fix to loan-liveness logic
   needs an explicit checklist line: "landed in `check`? landed in
   `corevalidate`? landed in `interp`? landed in `cgen`? landed in native's
   execution-document validator?" — M001's Phase 2 gap was a *fifth* consumer
   nobody had listed; M002 should enumerate consumers before starting each
   plan, not discover them mid-audit.
5. **Declare deferred scope in writing at decision time, not discovery time**
   (Lesson 4). Applies directly to (f)'s explicit naming of modules and
   Phase 5/6 Nyquist as deferred-with-reasoning in PROJECT.md, done now rather
   than at milestone close.
6. **Treat phase-count drift as a signal, not noise.** RETROSPECTIVE.md's
   3→7→10→13→14→15 sequence is a symptom of absorbing gap-closure work as plan
   count instead of renegotiating scope. For M002: if Phase 2 or 3 (the
   highest-risk phases) start exceeding ~2x their initial plan estimate,
   that's the trigger to renegotiate Phase 5 (Result payloads) out to M003
   explicitly, per §3's design — not to silently keep adding plans.
7. **Archival and path changes are code changes, budget for them** (Lesson 6)
   — irrelevant to mid-milestone pace directly, but relevant to not
   underestimating the milestone-close cost.

### Anti-patterns of "going fast" specific to a compiler project

- **Landing a new dispatch mechanism under load with other large structural
  changes** — the exact anti-pattern M001 avoided on purpose by deferring
  `OpCall`, and the exact anti-pattern option (c) (charter+modules) would
  reintroduce.
- **Trusting the producer's claim at a trust crossing instead of independent
  re-derivation** — fast in the moment, but it's precisely what
  `escape:coordinated-source-to-core-false-claim` demonstrates is
  *executable*, not hypothetical, in this architecture.
- **Skipping the shadow-run before deleting an old derivation** — fast to
  delete, but the 230,692-comparison shadow run existing at all is what makes
  M001's "fixpoint became the sole liveness law" claim credible instead of
  assumed.
- **Running validation "when convenient"** — RETROSPECTIVE.md's own diagnosis
  of exactly this pattern (Nyquist compliance tracked gate conditions, not
  intentions).
- **Optimizing the fast/dev path by skipping a soundness obligation rather
  than deferring it explicitly** — `compiler-and-feedback-latency.md`'s
  explicit rule: "the development path may defer expensive evidence, but it
  may not report a program as checked when a required soundness obligation
  was skipped."
- **Whole-program/global inference creeping in under time pressure** because
  it's easier to implement than a properly bounded summary-based analysis —
  `compute-efficiency-constitution.md` names this directly as an anti-pattern
  ("Global inference/coherence whose cost depends on the whole ecosystem"),
  and it is the single most likely failure mode for interprocedural loan
  liveness specifically, since the naive implementation of "just look at the
  callee body" *is* whole-program inference in disguise.

### Real project post-mortems, and what they tell M002 specifically

- **Rust NLL → Polonius** (2018-2026, [Rust Blog Aug 2026](https://blog.rust-lang.org/2026/08/04/enabling-polonius-alpha-on-nightly/),
  [EuroRust 2024 retrospective](https://eurorust.eu/2024/talks/the-first-six-years-in-the-development-of-polonius/)).
  VERIFIED WEB. Polonius (the interprocedural/flow-sensitive generalization of
  NLL's borrow model) spun out in 2018, hit a wall where "certain programs were
  considerably slower than NLL to the extent that using that
  implementation/formulation... was a non-starter," required a full
  reformulation in 2023, targeted stabilization in 2024, and only reached
  nightly alpha in August 2026 — eight years end to end, with a stated 10-20%
  compile-time cost accepted as the price of the expressiveness gain even at
  alpha. **Lesson for M002 (judgment):** generalizing a borrow/liveness model
  from local/lexical to flow-sensitive-across-boundaries is a known
  multi-year-scale hard problem even for a team with vastly more resources
  than this project. M002's ambition (interprocedural liveness in one
  milestone) is not comparable in scope to Polonius's full ambition — Lang's
  liveness model is intentionally simpler (function-boundary summaries, not
  arbitrary flow-sensitive region inference) — but the Rust experience is
  direct evidence that "the cost model looked fine intraprocedurally and then
  wasn't across a boundary" is a real, not hypothetical, failure mode, which
  is exactly why Phase 2's catching gate (§3) is a cost-scaling measurement,
  not just a correctness proof.
- **Swift ownership (SE-0377 parameter ownership, SE-0390 noncopyable types,
  SE-0304 structured concurrency, SE-0504 cancellation shields)**. Already
  cited as primary sources in `semantic-kernel-contract.md`. VERIFIED
  (project's own primary-source research, confirmed applicable). Swift
  sequenced ownership *modifiers* (borrowing/consuming, SE-0377, accepted
  2023) separately from and before full noncopyable *types* (SE-0390),
  and structured-concurrency cancellation semantics went through a second
  proposal (SE-0504, cancellation shields) years after the first
  (SE-0304) once real cleanup-after-cancellation needs surfaced. **Lesson for
  M002 (judgment):** Swift's own sequencing discipline validates (f)'s choice
  to *not* try to land the full ownership-across-boundary story and the
  Result-payload story in one undifferentiated pass — Swift's own team split
  semantically-adjacent-but-separable concerns (parameter modifiers vs. type
  system vs. cancellation) across proposals precisely so each could be
  independently evaluated and shipped when ready, matching (f)'s Phase 5
  placement of `Result` payloads as separable and movable.
- **Zig stage1→stage2 self-hosted compiler transition** ([Zig
  "Goodbye to the C++ Implementation"](https://ziglang.org/news/goodbye-cpp/),
  [HN discussion Aug 2022](https://news.ycombinator.com/item?id=32529113)).
  VERIFIED WEB. The self-hosted (stage2) compiler shipped as default with
  known regressions (async/await incomplete, "some fresh bugs") and the
  removal of stage1 left 650 open issues labeled against the now-deleted
  codebase. **Lesson for M002 (judgment):** shipping a structural rewrite
  (here: liveness derivation extended across calls) "as the new default" while
  quietly carrying known regressions/gaps is a real, executed pattern in a
  peer systems-language project — it is exactly the failure mode (f) tries to
  avoid by explicitly naming carried debt (Phase 5/6 Nyquist, modules) rather
  than letting it accumulate as undocumented "fresh bugs" the way Zig's
  transition did. The 650-issue number is a caution about the cost of *not*
  naming debt explicitly at transition time.
- **Go generics rollout** ([Go 1.18 issue #49569](https://github.com/golang/go/issues/49569),
  [Go 1.18/1.20 release notes](https://go.dev/blog/go1.18)). VERIFIED WEB. Go
  1.18 shipped generics with a measured 15-18% compile-time regression,
  root-caused specifically to *front-end* cost (`types2`/`noder2`), not the
  SSA backend — i.e., the cost showed up exactly where it wasn't initially
  budgeted for, and took until Go 1.20 (two releases later) to bring back in
  line. **Lesson for M002 (judgment):** this is the single most directly
  relevant precedent to Phase 2's catching gate. A new cross-cutting semantic
  feature's cost frequently shows up in *checking/type-front-end* work, not
  the place engineers instinctively profile first (codegen/backend) —
  reinforces that Phase 2's gate must measure `check`/`corevalidate` cost
  specifically, not just end-to-end `-O3` build time, and must do so *before*
  Phase 3-4 build on top of an unmeasured cost shape.
- **CompCert and seL4** ([CACM
  2010](https://cacm.acm.org/research/sel4-formal-verification-of-an-operating-system-kernel/),
  [seL4 Verification](https://sel4.systems/Verification/proofs.html)).
  VERIFIED WEB (widely-reported facts, well-established in the formal-methods
  literature). CompCert's correctness proof is specifically valuable because
  Csmith fuzzing found wrong-code bugs in every other tested compiler but not
  CompCert as of 2011 — i.e., the proof paid for itself by catching a defect
  class (miscompilation) that black-box testing alone systematically misses.
  seL4's methodology is "verification-aware structuring... several
  intermediate languages, each equipped with a formalized operational
  semantics" — proof-friendly IR staging, not proof bolted on after the fact.
  **Lesson for M002 (judgment):** this validates the project's own existing
  architecture (independent re-derivation at trust crossings, `check` vs.
  `corevalidate` as genuinely separate implementations, not proof in the Coq
  sense but the same underlying insight — a second *independent* encoding
  catches what a single implementation's tests cannot) rather than suggesting
  M002 needs actual machine-checked proofs. The relevant transferable lesson
  is proof-friendly *staging* (Phase 1→2→3→4 in §3, each adding one
  independently-checkable layer) over attempting end-to-end verification in
  one undifferentiated pass — which is what options (a)-(e) all already do
  structurally, and what (c) (modules bundled in) would compromise by
  collapsing two independently-checkable concerns into one pass.

---

## 5. High-leverage experiments (bounded, time-boxed, M002-scoped)

Each follows the project's own spike discipline (`ownership-evidence-roadmap.md`'s
per-spike checklist: falsifiable question, named competing mechanisms,
oracle that doesn't reuse the mechanism under test, at least one injected
defect, stop when the gate is answered).

### 1. Interprocedural liveness cost-scaling probe (run alongside Phase 1)

**Hypothesis:** function-summary-based interprocedural liveness stays linear
(or otherwise boundedly sub-quadratic) in call-graph size, not blowing up the
way naive whole-callee-body re-walking would.

**Cheapest disproving experiment:** generate a synthetic call-graph corpus
with controlled fan-out/depth (mirroring M001's own "explicit block/edge/
transfer counts and 1,000-origin interface growth" scale series from
`ownership-evidence-roadmap.md`'s testing-strategy table) and measure counted
work (not wall clock) against call-graph size before writing the real Phase 2
implementation — a cheap Go-workbench-style throwaway model, not production
code, exactly like Spikes 001-005.

**Decision it informs:** whether Phase 2 can extend the existing
`loanLivenessFixpoint` architecture directly, or whether a summary-caching
layer (memoized per-function origin/ability facts, consumed rather than
recomputed at each call site) must be designed in from the start — this is a
Phase-2-blocking decision, so the spike must complete before Phase 2's plan is
written, not during it.

### 2. Dev-backend timing probe: does Cranelift/QBE change the -O3 equivalence risk profile? (run alongside Phase 4, non-blocking)

**Hypothesis:** `residual-uncertainty-register.md`'s Frontier 3 "development
backend" question (Cranelift leading, QBE as complexity control, versus the
current Clang-at-`-O0`/`-O3` approach) remains correctly deferred — i.e., that
switching dev backends would not have changed how Phase 4's interprocedural
`-O3`/LTO divergence-finding worked.

**Cheapest disproving experiment:** after Phase 4 lands its adversarial
interprocedural divergence corpus (the `-O3`/LTO equivalence work), attempt to
reproduce just one or two of the found divergences under a fast dev-backend
prototype (Cranelift via its Go/Rust bindings, or QBE directly) purely as a
measurement, not a production path — if the same divergence classes surface at
a different optimization tier's boundary, that's information; if they don't
reproduce at all (because a simpler backend doesn't do the same interprocedural
optimization), that's equally informative about which backend actually
generates the risk.

**Decision it informs:** whether Frontier 3's "development backend" choice is
genuinely low-risk to defer further (current posture) or whether the
interprocedural work has just made backend choice materially more consequential
than it was for the M001 intraprocedural case — informs M003 prioritization,
not M002 scope.

### 3. Recursive/mutually-recursive call stress corpus (run alongside Phase 1)

**Hypothesis:** the bounded interpreter call stack and cycle-refusal design
correctly distinguish "legal deep recursion" from "illegal cycle" without a
false-positive/false-negative rate that would make ordinary recursive Lang
programs (a parser, a tree walker — exactly the dogfood verticals named in
`ownership-evidence-roadmap.md`'s Experiment 007) unusable.

**Cheapest disproving experiment:** hand-author or generate a small corpus of
legal deeply-recursive programs (parser-shaped, per the project's own planned
007 dogfood target) alongside genuinely cyclic programs, run them through the
Phase 1 cycle-refusal/stack-bound logic, and check both directions: does the
refusal ever fire on legal recursion, and does the bound ever silently accept
a cycle it should refuse?

**Decision it informs:** whether the call-stack bound is a fixed constant
(simple, but risks false-positive refusal on legitimate deep recursion) or
needs to be a declared/measured budget per `compute-efficiency-constitution.md`'s
budget-manifest pattern (`agent.repair_rounds.p95`-style, but for call depth)
— a Phase-1-blocking design decision best resolved before, not after, Phase 1
ships its refusal semantics as a fixed contract other phases build on.

### 4. Nyquist-closure cost measurement for Phase 3 (M001) before committing to fold it into M002 Phase 3

**Hypothesis:** closing Nyquist validation for M001's original Phase 3 (loan
liveness) is genuinely cheap when done as part of the code the M002 rebuild is
already touching — i.e., that (f)'s Phase 3 design (fold, don't defer) is
actually correct sizing, not wishful scheduling.

**Cheapest disproving experiment:** before finalizing the M002 phase plan,
spend a small time-boxed pass (measured in hours, not days) auditing exactly
what M001 Phase 3's Nyquist gap consists of (per
`.planning/milestones/M001-MILESTONE-AUDIT.md`) and estimate incremental cost
of closing it *within* the M002 Phase 3 rebuild versus as a fully separate
pass.

**Decision it informs:** whether (f)'s Phase 3 design holds as scoped, or
whether Nyquist closure for Phase 3 should itself be named an explicit
stretch/fallback item (like Phase 5's Result payloads) rather than a
committed part of the phase — this is the one place in the recommended
sequencing where the estimate hasn't been empirically checked yet, and it's
cheap to check before the milestone's plans are finalized.

### 5. (Optional, lower priority) `Result` payload / interprocedural origin interaction probe (run before Phase 5, informed by Phase 4's output)

**Hypothesis:** `Result` payload origins (D-04-30) reduce cleanly to the
already-proven Phase 2-4 interprocedural origin machinery, with no new kernel
rule needed.

**Cheapest disproving experiment:** once Phase 4 lands, take 2-3 hand-authored
programs that return a `Result` with a borrowed payload across a function
boundary and check whether the existing origin/ability derivation from Phases
2-4 already produces the right answer, or whether it silently falls through to
an unsound default.

**Decision it informs:** whether Phase 5 is genuinely independent surface work
(as assumed in §3) or secretly depends on new interprocedural rules — directly
gates whether Phase 5 can safely be the "negotiable, movable to M003" item (f)
treats it as.

---

## 6. Explicit recommendation

**Recommendation: Option (f)** — the charter (a) as the milestone's spine,
with Nyquist closure folded specifically into the phase that rebuilds
loan liveness (not deferred as a separate pass, not expanded to cover
Phases 5/6), the agent-loop extension (d) landed as the milestone's own final
phase once the interprocedural defect taxonomy actually exists, and modules
(c) explicitly excluded and named as M003's lead candidate rather than
silently dropped.

**Reasoning, stated plainly:**

1. The charter as written (a) is already the right-sized unit of semantic
   risk — it is the single carried debt item from M001 (D-03-02), it matches
   the project's own residual-uncertainty-register stop rule ("the next
   unknown-reduction work is implementation... not more unconstrained
   enumeration"), and re-litigating whether to build it is not warranted by
   this research.
2. What the charter-as-written under-specifies is *sequencing and honesty
   about validation debt*, which is exactly where M001's own retrospective
   says the project's real failures lived (Nyquist slipping, phase-count
   drift, fixing one of N peers). (f) is not a scope change from (a); it is a
   scheduling and disclosure discipline applied to (a), directly answering
   the retrospective's own named lessons.
3. Modules (c) is real, valuable, eventually necessary work — but stacking it
   onto the milestone that is already this project's hardest technical bet
   (extending five-plus-consumer loan liveness and native equivalence across
   a call boundary, the exact shape that cost M001's Phase 3 its worst
   plan-count and time-per-plan inflation) repeats the precise anti-pattern
   M001 avoided on purpose when it deferred `OpCall` out of Phase 4/5. Rust's
   own NLL→Polonius history (§4) is independent evidence that generalizing a
   local liveness model across boundaries is hard enough on its own to deserve
   an undivided milestone.
4. The agent-loop extension (d) is the right final phase, not a parallel
   workstream, because you cannot correctly design cause-DAG edges or repair
   applicability for a defect taxonomy you haven't built yet — building it in
   parallel risks the Phase-6-shaped rework M001 avoided by *not* speculatively
   building diagnostics ahead of the semantics they describe.

**What I would NOT do:**

- **I would not bundle modules/separate compilation into M002** (rejecting
  option (c) as scoped), even though the compiler-engineer lens has a real
  point about avoiding a second call-graph/origin rebuild later. The
  coordination-cost evidence from M001's own Phase 3 is too direct a warning,
  and the residual-uncertainty-register itself ranks modules (Frontier 1) below
  the semantic kernel (Frontier 0, where interprocedural ownership lives) in
  priority. I would instead name modules explicitly in PROJECT.md's Out of
  Scope for M002, with the reasoning captured now — per Lesson 4 — so that
  when M003 revisits it, the "we could have built calls and modules together"
  tradeoff is visible, not silently forgotten.
- **I would not run Nyquist closure for Phases 5 and 6 (M001) inside M002.**
  These phases' surfaces (native optimization tier, agent-facing schema) are
  not what M002 is directly rebuilding, so validating them now risks
  validating a shape that changes again once Phase 6 (M002) extends the
  agent surface anyway. I would carry them forward as still-disclosed debt,
  explicitly, rather than either quietly dropping them or expensively
  validating work that's about to be touched again.
- **I would not treat the high-leverage spikes in §5 as scope-additive.**
  They are explicitly bounded, disposable, and gated the same way M001's
  Spikes 001-005 were — if experiment 1 (liveness cost-scaling) or experiment
  3 (recursion stress corpus) surfaces a real problem, that's a design input
  to Phase 1/2, not a commitment to build a production-grade solution to
  every question a spike raises. The project's own stop rule applies: "stop
  when the gate is answered; do not expand the spike into the compiler."
- **I would not let Phase 5 (`Result` payloads) or the Phase-3-folded Nyquist
  closure become fixed commitments that block the milestone if Phase 2/3 (the
  genuinely hard, historically-expensive work) run over budget.** Per Lesson 6
  (phase-count drift), I would treat both as explicitly negotiable-out-to-M003
  the moment Phase 2 or 3 shows the same 2x-plus plan-count inflation Phase 3
  showed in M001, rather than silently absorbing the overrun as more plans on
  an already-committed scope.
- **I would not skip the cost-scaling measurement (spike 1, §5) and go
  straight to implementing Phase 2's production liveness extension.** The Go
  1.18 generics precedent (§4) and Rust's Polonius performance wall (§4) are
  both direct evidence that this exact class of feature — a local analysis
  generalized across a new boundary — has a real, not hypothetical, risk of
  hitting an unexpected cost cliff in the front-end/checking layer
  specifically. Skipping the cheap probe to save a few days risks losing far
  more time discovering the cliff mid-Phase-2, after committing to an
  architecture the cost problem then forces a redesign of.

---

## Sources

**Project-internal (primary, read in full 2026-09-08):**
- `~/projects/ai-lang/.planning/PROJECT.md`
- `~/projects/ai-lang/.planning/RETROSPECTIVE.md`
- `~/projects/ai-lang/.planning/MILESTONES.md`
- `~/projects/ai-lang/wiki/ownership-evidence-roadmap.md`
- `~/projects/ai-lang/wiki/semantic-kernel-contract.md`
- `~/projects/ai-lang/wiki/compiler-and-feedback-latency.md`
- `~/projects/ai-lang/wiki/compute-efficiency-constitution.md`
- `~/projects/ai-lang/wiki/residual-uncertainty-register.md`
- `~/projects/ai-lang/wiki/modules-architecture-and-live-development.md` (skimmed)
- `~/projects/ai-lang/wiki/language-lessons.md` (skimmed)

**External (fetched via web search 2026-09-08; confidence marked inline per claim):**
- [Enabling the next iteration of the borrow checker on nightly — Rust Blog, Aug 4 2026](https://blog.rust-lang.org/2026/08/04/enabling-polonius-alpha-on-nightly/)
- [The first six years in the development of Polonius — EuroRust 2024](https://eurorust.eu/2024/talks/the-first-six-years-in-the-development-of-polonius/)
- [Polonius current status and roadmap](https://rust-lang.github.io/polonius/current_status.html)
- [Goodbye to the C++ Implementation of Zig — Zig News](https://ziglang.org/news/goodbye-cpp/)
- [Zig is now self-hosted by default — Hacker News discussion, Aug 2022](https://news.ycombinator.com/item?id=32529113)
- [cmd/compile: Go 1.18 compile time regression from generics — golang/go#49569](https://github.com/golang/go/issues/49569)
- [Go 1.18 Release Notes](https://go.dev/blog/go1.18)
- [seL4: Formal Verification of an Operating-System Kernel — CACM](https://cacm.acm.org/research/sel4-formal-verification-of-an-operating-system-kernel/)
- [seL4 Proofs](https://sel4.systems/Verification/proofs.html)

Already embedded (project's own primary-source research, re-verified applicable
here, not re-fetched): Rust RFC 2094 (NLL), Swift SE-0304/0377/0390/0504,
Go memory model, LLVM UndefinedBehavior/LangRef docs, rustc-dev-guide incremental
compilation — all cited in full in `semantic-kernel-contract.md` and
`compiler-and-feedback-latency.md`.
