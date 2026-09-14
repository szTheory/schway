# Phase 13: Agent Loop for Interprocedural Defects - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-09-13
**Phase:** 13-agent-loop-for-interprocedural-defects
**Areas discussed:** Blame-vs-detection rule, Which 3 repairable classes, Function naming + schema, Held-out split construction

**Mode:** advisor (`$HOME/.claude/gsd-core/USER-PROFILE.md` present).
`vendor_philosophy: opinionated` → calibration tier **`minimal_decisive`**
(1–2 options per area). `technical_background: true` and
`explanation_depth: practical-detailed:technical` → `NON_TECHNICAL_OWNER = false`,
so no plain-language reframing was applied.

**Research:** the user supplied a standing fan-out prompt asking for
breadth + depth across all relevant stakeholder-role lenses, pros/cons/tradeoffs,
antipatterns/best-practices/footguns/lessons-learned, online research into other
products and ecosystems, an adversarial pass, then a synthesized one-shot
recommendation per decision point. Four `gsd-advisor-researcher` agents ran in
parallel (model: opus) — one per selected gray area — each instructed accordingly.
All four returned; the tables below are the synthesized results.

**Notable convergence:** two agents working independently (blame rule, explain
schema) arrived at the same conclusion that function attribution must live on
`protocol.ExplainNode` and never on `diagnostic.Cause`, for the same reason —
`Causes` is marshalled wholesale into the diagnostic identity sha256.

---

## Blame-vs-detection rule (DX-06)

| Option | Description | Selected |
|--------|-------------|----------|
| A: Contract-boundary blame | B1/B2/B3 algorithm over `functionByOperationID`. Blame the party that broke its own declared signature; caller-blame as default; cycles exempt. Computed in `check` only, `corevalidate` does not re-derive. Fail-open hazard closed by exhaustiveness guard + explicit `blame_undetermined` routing. No-op for all seven shipped codes, so zero ID churn. | ✓ |
| B: Detection-site + annotate | Keep the current Primary span, express blame only as explain-graph metadata. Smallest diff and zero risk, but fails criterion 3 outright — an agent acts on `primary_span` first. | |

**User's choice:** A — Contract-boundary blame.

**Notes:** Descends from Findler–Felleisen contract blame as formalized by
Wadler & Findler's blame calculus (*Well-Typed Programs Can't Be Blamed*, ESOP
2009), which transfers directly because `FunctionSignature` / `callable ⊆
publishable` (D-04-03) is a declared boundary contract. Primary/secondary
presentation follows rustc's `MultiSpan`; the "carry why back to the right site"
machinery follows rustc's `ObligationCauseCode`, which exists precisely because
inference's default of blaming the detection site produced the misattributed-
lifetime-error bug class.

Two things were recorded as explicit limitations rather than smoothed over:

1. **Fail-open hazard.** B2-as-default is fail-open if B1's predicate is
   incomplete. Closed by a compile-time exhaustiveness guard per declared-contract
   field plus a `blame_undetermined` route (D-13-07).
2. **A case the rule does NOT survive.** Mutual consistency with incompatible
   intent — both sides individually honest, the callee's honest contract simply
   wrong for the program's intent. Repair-then-re-check cannot falsify the
   choice; criterion 3's gate is blind. Mitigated by emitting both sites as
   `RequiresConfirmation` so `DriverEligible` refuses to auto-apply (D-13-08).

Alternatives surveyed but not tabled: GHC/Haskell type-error blame and type-error
slicing, Flow/TypeScript's caller-side "argument is not assignable" framing,
Racket/Eiffel contract blame, Infer `bug_trace`, CodeQL path-problem results.
Infer supplied the machine-actionability lesson: an agent needs one authoritative
edit site plus an ordered function-attributed chain, not an undifferentiated span
set.

---

## Which 3 repairable classes (DX-07)

| Option | Description | Selected |
|--------|-------------|----------|
| Set A (3 distinct families) | `interprocedural_loan_liveness` (`move_after_interprocedural_loan`) + `syntax.fallible_call_not_consumed` (`wrap_call_in_try`) + `check.call_argument_type_mismatch` (`use_matching_argument`, gated on a unique in-scope match). Cycle refusal becomes explain+blame only, returning `unrepairable`. | ✓ |
| Set B (lower-risk third) | Swap the third for call-site `use_after_move` reusing the existing `use_transfer_target` kind. Inputs already recorded at `check.go:3051`. Lowest implementation risk, but vulnerable to a criterion-2 novelty challenge since both the code and repair kind already exist. | |

**User's choice:** Set A.

**Notes:** The structural insight driving the whole area: `cmd/lang-repair` is
kind-agnostic by design and has no kind→edit table, so "repairable" is entirely a
`check.go` **emission** question. Set A was preferred for evidence breadth — three
genuinely distinct defect families (ownership / effect / type) rather than three
variations on ownership.

**Cycle refusal settled as explain-only.** The source already settles it twice
(`checkCallGraphAcyclic`'s doc; `cfgBackEdgeDiagnostic` at `check.go:2143-2146`
restating D-07-31c). The only candidate edits are "delete one call edge" — every
choice changes program meaning and yields a clean-checking program that does
something different, the exact failure DX-07 must not ship — or "inline the
callee", which is unbounded and not span-local. This **reinterprets ROADMAP.md's
scope-cut note**, which names cycle refusal among the two highest-value classes:
that value is explainability, not repair. Recorded as D-13-12 so the note is not
later read as a repair commitment.

**The key defensive argument** (D-13-11): the loan-liveness reorder can honestly
ship `MachineApplicable` in Lang where rustc ships `MaybeIncorrect`, because
`take` is a pure compile-time ownership transfer with no observable runtime
effect — relocating it past a call cannot change meaning, whereas reordering in
Rust can move a `Drop`.

Precedent surveyed: rustc `Applicability` + `cargo fix`'s MachineApplicable-only
rule (the direct ancestor of `DriverEligible`) and the 2018-edition rustfix
incidents; ESLint `--fix`'s independently-reached non-overlapping + bounded-rerun
design; Go `analysis.SuggestedFix`'s package confinement; TypeScript's codefix
catalogue shipping cross-file edits essentially only for imports; Error
Prone/Refaster and clang-tidy `FixItHint` confidence notions.

---

## Function naming + schema (DX-05)

| Option | Description | Selected |
|--------|-------------|----------|
| A: Explain-time projection | `function_id` + `function_name` (omitempty) on `ExplainNode`, optional `functions` table on `ExplainSummary`, resolved from innermost containing function span and peer-re-derived from the AST. Plus the `narrows` function-scope guard fixing cross-function mis-parenting. No `lang.explain/0` bump for the fields. | ✓ |
| A-minus: fields only, no guard | Ship the fields, defer the `narrows` guard as recorded debt. Satisfies criterion 1's literal wording but leaves the DAG shape wrong across function boundaries — display fixed, derivation still incorrect. | |

**User's choice:** A — fields *and* guard.

**Notes:** Modeled on SARIF's `logicalLocation{name, fullyQualifiedName,
kind:"function"}` split from physical location, with the implementers' lesson
applied: SARIF's `run.logicalLocations` + `index`/`parentIndex` indirection is
routinely ignored by consumers (GitHub code scanning included), so identity must
be readable **inline** on the node rather than behind a join. LSP models "this
cause is over there" via `DiagnosticRelatedInformation` with no logical field,
which is why LSP clients cannot group related information by function without
re-parsing. Infer buries the procedure name in prose `description` and every
downstream tool regexes it.

Rejected shapes and why, on verified in-repo grounds rather than taste:
- **Encoding into `Cause.Detail`** — `explainCorrelationKey` is literally
  `kind + ":" + detail` over place/owner/loan/transfer_target, so mutating
  `Detail` silently repartitions `same_binding` correlation and changes the **DAG
  shape**, *and* `Detail` sits inside the identity hash. Double kill.
- **Function identity on `diagnostic.Span`** — `Span` is embedded in
  `Diagnostic.Primary` and in every `Cause` inside the marshalled identity struct,
  so it would churn effectively every diagnostic ID in the project.
- **Producer-side field on `diagnostic.Cause`** — deferred to M003 modules; if
  landed then, it must be non-identity-bearing, projecting `Causes` to
  `{Kind, Detail, Span}` exactly as `RepairKinds` does for repairs.

**The sleeper bug, confirmed real:** `explainSpanStrictlyContains` is pure byte
containment, so a whole-function-declaration span becomes the narrowest
containing ancestor of every later cause in that body — manufacturing `narrows`
edges asserting a relation the rule never meant. This is a *derivation* bug, not
a display bug, which is why A-minus was rejected. The guard keeps the closed
three-value edge vocabulary intact by routing cross-function parenting to
`caused_by`.

**Schema bump verdict:** not for the fields (additive + `omitempty`;
`ExplainSummary` has no identity hash; `corevalidate` never touches explain;
in-repo precedent `core.Function.ForeignContract` / D-04-23). The **real** bump
trigger is the guard — if it flips any edge kind on unchanged single-function
input, that is a published-output semantic change and earns `lang.explain/1`
(D-13-20).

**Truncation verdict:** reuse `node_budget` / `depth` unchanged; no
`truncated:explain.function_budget`. Cross-function expansion adds entries to the
same flat `diag.Causes` list `ExplainMaxNodes` already bounds — no new unbounded
dimension, and a second axis would need its own ordering contract.

---

## Held-out split construction

| Option | Description | Selected |
|--------|-------------|----------|
| A: Mechanical + topology-disjoint | New marker-driven interprocedural injectors in the existing `Injector`/`markerGuard` shape over hand-authored bases differing in call-graph topology. Split rule: topology-triple disjointness + held-out hop distance ≥ 2 + fail-closed baseline. Sealed by `HELDOUT.sha256`, property assertions only, criterion-3 twin pair mandatory. | ✓ |
| B: Two hand-authored corpora | Write the held-out broken programs directly, as M001 effectively did. Cheapest, no new machinery. Carries M001's documented weakness forward. | |

**User's choice:** A.

**Notes:** Option B was shown to repeat a weakness already present in shipped
M001 evidence — `heldout_move_defect.lang` and `derivation_move_defect.lang` are
near alpha-renamings (`item`→`buffer`, `moved_once`→`delivered`) and
`TestPhase6DefectCorpusIsHeldOut` asserts only byte-inequality, which any rename
passes.

**The sharpest finding of the whole discussion:** the overfittable artifact in
Phase 13 is **not** the driver — `repair.go` has no kind→edit table to overfit.
It is **`check`'s choice of which function's span the repair points at**. So the
split must vary the dimension a blame rule could overfit: call-graph topology and
detection-to-fix hop distance. Varying identifiers is ceremonial; varying
topology is not. This became D-13-25 and reshapes the whole fixture design.

**Criterion 3's fixture must be a twin pair.** The natural construction — shared
callee `sink` with two callers `alpha`/`beta`, diagnostic firing in `alpha`, only
a fix in `sink` making the whole program re-check clean — is passed **by luck** by
a rule that always blames the callee. So it ships with a mirror fixture where the
unique whole-program-clean fix is the *caller*. Always-blame-callee fails the
twin; always-blame-detection-site fails the first. Either half alone is a coin
flip.

**Anti-contamination:** sealed `HELDOUT.sha256` manifest so post-hoc tuning is a
loud reviewable diff; no goldens for held-out (property assertions only) because
the rustc `tests/ui --bless` lesson is that a regenerable expectation is not
evidence.

Rejected: bounded-exhaustive / grammar-based generation (Korat, Alloy small-scope,
csmith/YARPGen). Genuinely tractable given Lang's tiny surface, but there is no
oracle for *which function is the correct fix location* in a generated program —
only "does it check clean", which is exactly the plausible-vs-correct trap APR
fell into (Smith et al., FSE 2015).

---

## Follow-up decisions

### Cascade discipline for byte-span splicing

| Option | Description | Selected |
|--------|-------------|----------|
| One repair per pass | At most one driver-eligible repair per check/apply cycle, then re-run. Designs the cascade footgun out rather than testing around it. | ✓ |
| All non-overlapping, descending offset | rustfix/ESLint model. Fewer round-trips, but needs an overlap-detection control that itself needs mutation-killing. | |
| Leave driver untouched | Phase 13 changes only `check.go` emission; cascade risk handled purely by fixture discipline. | |

**User's choice:** One repair per pass.

**Notes:** Verified after the fact that this **confirms shipped behavior and
requires no code change**. `Repair()` (`cmd/lang-repair/repair.go:262-290`)
already performs exactly one diagnose → at most one apply → at most one reverify,
with no loop construct anywhere in the function (D-06-30), and `subprocess_count`
is structurally always 1 or 2. The 2018-edition `cargo fix` class of bug is
already designed out. Net effect: **`cmd/lang-repair` source needs no changes in
Phase 13** — tests extended only.

### `declare_foreign_symbol` shell repair (`check.go:3123`)

| Option | Description | Selected |
|--------|-------------|----------|
| Leave as-is, record as debt | Out of Phase 13's scoped domain; note in deferred ideas. | ✓ |
| Complete it | Give it a real Span/Replacement, making it driver-eligible — a fourth repair class, exceeding what criterion 2 asks. | |
| Delete the shell | Restores consistency but changes an existing diagnostic's published `RepairKinds`, which IS identity-bearing under schema `/1`, churning that diagnostic's ID. | |

**User's choice:** Leave as-is, record as debt.

**Notes:** It is currently the only repair in the tree advertising a `Kind` it
cannot apply. Both resolutions widen scope, so the phase boundary was kept clean
and the inconsistency recorded in CONTEXT.md's deferred ideas.

### M001 `testdata/phase6` corpus distinctness

| Option | Description | Selected |
|--------|-------------|----------|
| Retro-strengthen phase6 too | Apply the new distinctness control to M001's corpus as well. Closes a real hole in shipped evidence. | ✓ |
| Phase 13 fixtures only, record debt | Keep the phase boundary tight; leave the hole open but documented. | |

**User's choice:** Retro-strengthen phase6 too.

**Notes:** Caveat carried into CONTEXT.md as D-13-33 — M001's fixtures are
largely intraprocedural, so the topology triple degenerates there and a weaker
structural predicate is needed. If applying it turns currently-green tests red,
that is a real hole in shipped M001 evidence being surfaced, not a regression to
suppress; escalate rather than weaken the predicate.

---

## Claude's Discretion

Left to the planner (enumerated in full in CONTEXT.md `<decisions>`):

- Field naming granularity for the explain node (`function_id` + `function_name`
  vs a nested object vs SARIF `fully_qualified_name`).
- Whether the `functions` table on `ExplainSummary` ships now or waits for a
  consumer.
- Whether `functionByOperationID` is materialized once on `check.Result` or
  rebuilt per diagnostic.
- Whether `ExplainDefaultDepth` stays 3 for cross-function chains.
- Scope of class 3's uniqueness gate (all initialized in-scope places vs
  parameter + prior `let`s before the call site).
- New corpus directory vs extending `testdata/phase6/` (and the
  `isPhase6Corpus` dispatch sibling).
- Whether the topology triple is computed from the existing call-graph package or
  re-derived independently.
- Held-out fixture count per class (minimum: the twin pair, ≥ 2 for the blame
  class).
- Whether a non-`Callable` callee is B1 or B2 — leaning B2.

## Deferred Ideas

- `check.go:3123`'s `declare_foreign_symbol` shell repair — resolve in a future
  phase.
- The `Cause.Span`-is-identity-bearing / `Repair.Span`-is-not inconsistency —
  recorded as debt; forcing only if M003 requires producer-side function identity.
- Producer-side function identity on `diagnostic.Cause` — M003 modules candidate.
- Function attribution on raw `lang check` diagnostics (not just `explain`).
- Bounded-exhaustive / grammar-based fixture generation — revisit when a
  differential blame oracle exists.
- Foreign-C boundary causes with no enclosing Lang function — omit field or
  sentinel kind.

## Process Note

No `todo.match-phase` matches for phase 13 (`todo_count: 0`), so no todos were
folded or reviewed. No SPEC.md exists for this phase, so requirements were not
pre-locked. No `.continue-here.md` blocking anti-patterns were present. No
`.planning/codebase/` maps exist, so codebase scouting used the grep fallback.
