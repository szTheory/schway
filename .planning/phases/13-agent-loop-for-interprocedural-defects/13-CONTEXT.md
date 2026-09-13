# Phase 13: Agent Loop for Interprocedural Defects - Context

**Gathered:** 2026-09-13
**Status:** Ready for planning

<domain>
## Phase Boundary

An AI agent hitting a **cross-function** defect gets (a) a bounded cause chain in
which every step **names the function it belongs to**, (b) blame that points at
the **function that must actually be edited** rather than the function where the
violation happened to be detected, and (c) at least **three interprocedural
defect classes it can repair through the `lang --json check` protocol alone**,
proven on a held-out fixture split rather than on the fixtures the repairs were
derived from.

This is the final phase of M002 and the only one whose subject is the **agent
loop** rather than the semantics. Phases 07-12 built the interprocedural defect
taxonomy; this phase *serves* it. No new semantic guarantee is introduced here —
if a defect class does not already have a diagnostic code in
`internal/compiler/check/check.go`, it is out of scope.

**What the shipped source actually looks like** (verified 2026-09-13 against the
tree, not inferred from the roadmap):

- `protocol.ExplainNode` is `{ID, Kind, Detail, Span, Availability}` — there is
  **no function field** (`internal/compiler/protocol/protocol.go:223-229`).
- `diagnostic.Span` is `{Start int, End int}` — **byte offsets only**, no file,
  no function, no line/col (`internal/compiler/diagnostic/diagnostic.go:16-19`).
- `buildExplainGraph` parents causes by pure byte-offset containment
  (`narrows`), binding correlation (`same_binding`), or fallback (`caused_by`).
  Truncation codes today are exactly `truncated:explain.node_budget` and
  `truncated:explain.depth`
  (`internal/compiler/session/session_phase6_explain.go:115-206`).
- `cmd/lang-repair` is **kind-agnostic by design**: it splices whatever
  `Span`/`Replacement` a `MachineApplicable` repair names and never maps a kind
  to an edit (`cmd/lang-repair/repair.go:1-21`). Therefore "repairable" is
  **entirely a `check.go` emission question**, not a driver question.
- `Repair()` already performs **exactly one** diagnose → at most one apply → at
  most one reverify, with no loop construct anywhere in the function (D-06-30);
  `subprocess_count` is always exactly 1 or 2
  (`cmd/lang-repair/repair.go:262-290`).
- Existing interprocedural codes: `check.interprocedural_loan_liveness`,
  `core.call_graph_cycle`, `check.call_argument_type_mismatch`,
  `check.call_arity_unsupported`, `core.callee_not_callable`,
  `check.call_return_type_unrepresentable`,
  `check.foreign_call_shape_unsupported`, `syntax.fallible_call_not_consumed`.
- Programs are **single-file** today (`ExplainCommandFile(path, ...)`), so
  function identity is resolvable by mapping a byte offset to its enclosing
  function declaration.

**Language-surface constraint that bounds every fixture in this phase:** per
`.planning/LANGUAGE-MATURITY.md` there is no arithmetic, no iteration, no `if`,
no strings, no arrays. Calls are arity-1
(`check.call_arity_unsupported`). Every fixture below is constructed from
`Byte`/`Buffer`, alternatives + `match`, `take`/`borrow`, and calls — nothing
else is available.

</domain>

<decisions>
## Implementation Decisions

### Blame attribution (DX-06)

- **D-13-01:** Adopt **contract-boundary blame**: blame the party that violated
  its **own declared signature**, with the detection site demoted to a typed
  secondary. This is Findler–Felleisen contract blame as formalized by Wadler &
  Findler's blame calculus ("well-typed programs can't be blamed"), which
  transfers directly because `FunctionSignature` / `callable ⊆ publishable`
  (D-04-03) *is* a declared boundary contract. Rejected: keeping the
  detection-site Primary and expressing blame only as explain metadata — an
  agent acts on `primary_span` first, repairs the caller, and re-check stays
  red, which is exactly the failure criterion 3 exists to catch.
  — **Reversibility:** costly — reversing means moving Primary spans on
  already-published diagnostics, which churns their sha256 IDs and the pinned
  fixtures in `check_ordering_stability_test.go`.

- **D-13-02:** The resolver data structure is an inversion of
  `program.Functions[i].Linear.Operations` into
  `functionByOperationID map[string]string`. This is exact and free: it needs
  **no** `core.Function.Span` field and **no** byte-offset range search, and it
  composes with the `spanByOperationID` map `checkInterproceduralLoanLiveness`
  already threads. Every cause is then typed as either a **use fact**
  (operation → owning function) or a **declared fact** (a `FunctionSignature`
  field — `Parameters[].Mode`, `Parameters[].Drops`, `Return`, `Callable`,
  `Abilities` → the declaring function). The existing `callee_return_contract`
  cause, whose Detail is literally `"<calleeID>:return.mode=<Mode>"`, is already
  a declared fact naming its owner.

- **D-13-03:** The rule is three-branch, lexicographic, **no fixed-point
  iteration**:
  - **B1 (contract violation, unary):** a function's own body contradicts its
    own declared signature → blame that function; Primary is the span inside
    that function where the contradiction is realized.
  - **B2 (contract misuse, default):** every function is internally consistent
    with its own declaration → blame the **caller** at the call operation.
  - **B3 (ties and cycles):** if B1 yields more than one function, take the
    minimum in the callee-before-caller topological order
    `buildInterproceduralSummaries` already computes; break residual ties by
    `program.Functions` index. **No new ordering authority is introduced.**

- **D-13-04:** **All seven currently-shipped interprocedural codes land in B2**,
  so adopting the rule moves **zero Primary spans today** —
  `interproceduralLoanLivenessDiagnostic`'s Primary is already the caller's
  offending `OpMove`, and `check.call_argument_type_mismatch`'s is already the
  caller's call site. The pinned IDs in
  `internal/compiler/check/check_ordering_stability_test.go` (e.g.
  `phase07/relay_escort_witness.lang → diagnostic:58c1b5b2072cda60f77d721e`) are
  **untouched**. All B1-shaped codes are new in Phase 13, so their IDs are
  unpublished and free. **Planner must verify this claim empirically before
  relying on it.**

- **D-13-05:** `core.call_graph_cycle` is **exempt from B3**. The blame set is
  the whole cycle, published as secondary causes; Primary stays at the closing
  operation.

- **D-13-06:** Blame is computed in **`check` only**. `corevalidate` must **not**
  re-derive it. D-08-17 / D-09-31 already record that the two peers cannot fail
  the same way (backward worklist vs. forward reachability closure) and that
  reverse-engineering one peer's cause shape into the other is the exact
  antipattern the peer discipline exists to prevent. **The peer gate agrees on
  refusal, never on blame site.**

- **D-13-07 (fail-open closure, MANDATORY):** B2-as-default is fail-open if B1's
  predicate is incomplete — a missed contract-violation case silently becomes
  caller-blame, and a caller-side repair can mask a broken callee contract while
  turning the check green. Required: enumerate B1 **per declared-contract field**
  with a **compile-time exhaustiveness guard**, and route any declared fact B1
  did not classify to an explicit **`blame_undetermined`** outcome
  (`RequiresConfirmation`, both sites published) rather than falling through to
  B2. This is non-negotiable — it is the difference between this rule and a
  heuristic.

- **D-13-08 (accepted limitation, record it, do not paper over it):** the rule
  does **not** survive *mutual consistency with incompatible intent* — caller and
  callee each individually honest, the real defect being that the callee's honest
  contract is the wrong contract for the program's intent. Both weakening the
  caller and strengthening the callee make the program check clean, so
  repair-then-re-check **cannot** falsify the choice and criterion 3's gate is
  blind. This is the classic gradual-typing "blame is correct but useless"
  failure. Mitigation matches the project's fail-closed posture: emit **both**
  sites as alternative repairs with `Applicability = RequiresConfirmation`,
  never `MachineApplicable`, so `DriverEligible` refuses to auto-apply either.

### Repairable defect classes (DX-07)

- **D-13-09:** Ship **Set A — three genuinely distinct defect families**, biased
  toward evidence breadth for the milestone:
  1. **`check.interprocedural_loan_liveness`** → repair kind
     `move_after_interprocedural_loan` (reverses D-08-25's `Repairs is nil`).
     Span: the `:stmt`-suffixed entry for the offending `OpMove` extended
     through the `:stmt` entry of the call that extends the loan — the channel
     `borrowConflictDiagnosticPostAssembly` (`check.go:1244`) already uses.
     Replacement: both statements re-emitted in swapped order, reconstructed
     from `core.Place.Name` + `CalleeID`, exactly as the existing post-assembly
     borrow-conflict repair reconstructs
     `"let " + targetName + " = borrow " + sourceName`.
  2. **`syntax.fallible_call_not_consumed`** (`check.go:3056`) → repair kind
     `wrap_call_in_try`. Span: `binding.RHS.Span`. Replacement:
     `"try " + Callee + "(" + Arguments[0] + ")"`. Interprocedural by
     construction — only the callee's *declaration* (`foreignSymbols`) makes the
     call fallible. Zero semantic-drift risk.
  3. **`check.call_argument_type_mismatch`** (`check.go:2999`) → repair kind
     `use_matching_argument`. Span: the argument token within
     `binding.RHS.Span`. Replacement: the name of the in-scope place whose type
     fact constructor equals `contract.ParameterType`.
  — **Reversibility:** costly — each lands a published repair kind in
  `lang.diagnostic/1`'s identity-bearing `RepairKinds`, so removing one later
  churns that diagnostic's ID.

- **D-13-09b (NARROWING, found empirically by the 13-01 tracer 2026-09-13 —
  amends D-13-09.1):** the "swap the move and call statements" repair is
  semantics-preserving only for the **backward** direction of
  `check.interprocedural_loan_liveness`, where the call is the loan's own
  recorded last use. In the **forward** direction — the loan propagated through
  the call onto a place that is read still *later* — the true last use lies
  beyond the call, and the swap does **not** fix the program. Verified by
  splicing the repair onto real `testdata/phase07` and `testdata/phase08`
  fixtures in both directions, not by reasoning. `interproceduralLoanLivenessDiagnostic`
  therefore takes a `callIsLastUse` gate and emits the `MachineApplicable` repair
  only when the edit is actually correct; the forward direction emits no repair
  and the driver honestly returns `unrepairable`. This is the fail-closed posture
  D-13-10 applies to `use_matching_argument`, arrived at independently for a
  second class. **Downstream plans must not widen this gate.**

- **D-13-09a (CORRECTION, found by research 2026-09-13 — supersedes the
  optimistic reading of D-13-17):** all three target codes currently build via
  `diagnostic.Error` (schema `lang.diagnostic/0`). Attaching a `Repair` forces a
  switch to `diagnostic.ErrorWithRepairs` (schema `/1`), and the `Schema` string
  itself is inside the hashed identity struct — so **their IDs change
  unconditionally**, whether or not a repair actually fires on a given program.
  **Six specific rows in `internal/compiler/check/check_ordering_stability_test.go`
  will churn and must be re-pinned as an explicit, called-out plan task**, not
  discovered mid-execution. This does **not** weaken D-13-17: D-13-17's claim is
  scoped to D-13-14 (the explain-side fields), which touches no identity payload
  and remains zero-churn. The two churn sources are independent; only this one is
  real, and it is a deliberate, reviewable consequence of shipping repairs.

- **D-13-10 (the safety gate on class 3):** emit `use_matching_argument`
  **only when exactly one initialized in-scope place matches**. On zero or ≥2
  matches emit **no repair at all** (`diagnostic.Error`; driver returns
  `unrepairable`). This is the one class where a clean-checking-but-semantically-
  different program is reachable, and the uniqueness gate **is** the entire
  safety argument. It is rustc's "there is a value of this type in scope"
  `MaybeIncorrect` downgrade converted into a **precondition** rather than an
  applicability downgrade — consistent with `NormalizeApplicability`'s refusal to
  ever default toward driver-eligible. The resulting `unrepairable` outcomes on
  genuinely-invalid programs are **correct behavior and need a positive test**,
  not a workaround.

- **D-13-11 (the argument the plan MUST make explicitly):** the loan-liveness
  reorder is `MachineApplicable` in Lang in a way its Rust analogue is not,
  because `take` is a **pure compile-time ownership transfer with no observable
  runtime effect** — relocating it past a call cannot change program meaning,
  whereas reordering in Rust can move a `Drop`. This is why rustc ships borrowck
  suggestions as `MaybeIncorrect`/help-only and Lang can honestly ship
  `MachineApplicable`. State this in the plan: *"rustc refuses to auto-fix borrow
  errors"* is the first objection a reviewer will raise.

- **D-13-12 (settled, and it reinterprets the roadmap):** **`core.call_graph_cycle`
  is NOT machine-repairable.** The source already settles this twice —
  `checkCallGraphAcyclic`'s doc ("No repairs (a cycle has no local, mechanical
  edit)") and `cfgBackEdgeDiagnostic` (`check.go:2143-2146`) restating D-07-31c
  intraprocedurally. The only candidate edits are "delete one call edge" (every
  choice changes program meaning and yields a clean-checking program that does
  something different — the exact failure DX-07 must not ship) or "inline the
  callee" (unbounded, not span-local). **ROADMAP.md's scope-cut note names cycle
  refusal as one of the two highest-value classes; that value is
  *explainability*, not *repair*.** Cycle refusal becomes the phase's showcase
  for DX-05/DX-06 — its `cycle_member` causes already carry per-edge spans — while
  honestly returning `unrepairable` from the driver. **The plan must say this
  explicitly so the scope-cut note is not later read as a repair commitment.**

- **D-13-13:** The `move_after_interprocedural_loan` kind is a **new string**,
  not a reuse of the existing `move_after_last_borrow_use`. The existing kind is
  Kind-only and **not** driver-eligible today; reusing it would silently change
  an existing diagnostic's driver behavior.

### Explain schema and function attribution (DX-05)

- **D-13-14:** Add `FunctionID string \`json:"function_id,omitempty"\`` and
  `FunctionName string \`json:"function_name,omitempty"\`` to
  `protocol.ExplainNode`, plus
  `Functions []ExplainFunction \`json:"functions,omitempty"\`` on
  `ExplainSummary` where `ExplainFunction{ID, Name, Span}`. The name is **inline
  on every node** so an agent never *has* to join; the table exists only to hand
  back the whole-function span for repair targeting. This is SARIF's
  `logicalLocation{name, fullyQualifiedName, kind:"function"}` split from
  physical location — with the lesson SARIF implementers learned applied: the
  `run.logicalLocations` + `index`/`parentIndex` indirection is the part
  consumers (GitHub code scanning included) routinely ignore, so identity must be
  readable inline.
  — **Reversibility:** reversible — additive `omitempty` fields on a
  non-identity-bearing struct.

- **D-13-15:** Resolution is by **innermost containing function span**, and is
  **peer-re-derived** — once from the core function table, once from the AST —
  satisfying the project's peer-re-derivation discipline. Nothing derived once
  and trusted.

- **D-13-16 (refused, on verified in-repo grounds):** do **not** encode the
  function into `Cause.Kind` or `Cause.Detail`. `explainCorrelationKey` is
  literally `kind + ":" + detail` over `place/owner/loan/transfer_target`
  (`session_phase6_explain.go:105-110`), so mutating `Detail` would silently
  repartition `same_binding` correlation and **change the DAG shape** — *and*
  `Detail` sits inside the identity sha256. Double kill. Likewise do not put
  function identity on `diagnostic.Span`: `Span` is embedded in
  `Diagnostic.Primary` **and** in every `Cause` inside the marshalled identity
  struct (`diagnostic.go:109-117`), so populating it would churn effectively
  every diagnostic ID in the project.

- **D-13-17:** **The explain-side change churns no diagnostic IDs.** Nothing in
  the identity payload (`{Schema, Code, Span, Causes}` / `+RepairKinds`) is
  touched by D-13-14. **Scope this claim to D-13-14 only** — D-13-09a records a
  separate, real churn caused by the `/0 → /1` schema switch when repairs are
  attached. The `cmd/lang-repair/testdata/*_capture.json` corpus is unaffected by
  D-13-14; whether it is affected by D-13-09a is a plan-task question.

- **D-13-18:** `lang.explain/0` does **not** bump for the fields — additive,
  `omitempty`, `ExplainSummary` carries no identity hash, and `corevalidate`
  never touches explain. In-repo precedent is explicit:
  `core.Function.ForeignContract` is documented as "additive, omitempty ... so
  its serialized bytes are unchanged" (D-04-23). `lang.diagnostic/0 → /1` bumped
  only because the *identity basis* changed, which this does not.

- **D-13-19 (the sleeper bug — confirmed real, and it is a derivation bug, not a
  display bug):** `explainSpanStrictlyContains` is pure byte containment, so any
  node whose span is a whole function declaration becomes the narrowest
  containing ancestor of **every** subsequent cause anywhere in that body,
  manufacturing `narrows` edges that assert a containment relation the rule was
  never written to mean once causes cross functions. **Guard:** a node whose span
  *equals* a `core.Function.Span` is a function-scope node and may be a
  `caused_by` parent but **never** a `narrows` parent; additionally `narrows`
  requires parent and child to resolve to the **same `function_id`**.
  Cross-function parenting is `caused_by`, which keeps the **closed three-value
  edge vocabulary intact**. (SARIF reinforces this: `threadFlowLocations` carry
  an explicit integer `nestingLevel` for call depth precisely because call
  structure must never be inferred from coordinates.)
  — **Reversibility:** costly — it changes published DAG shape.

- **D-13-20 (the actual schema-bump trigger):** before landing D-13-19, run the
  **existing** explain fixtures. If any edge kind flips on an **unchanged
  single-function** input — `core.call_graph_cycle`'s `cycle_member` causes are
  the candidate, since their spans come from `spanByOperationID` — that is a
  semantic change to published output and **that** earns `lang.explain/1`. A
  silent shape change is by far the worse outcome.

- **D-13-21:** **Reuse `truncated:explain.node_budget` and
  `truncated:explain.depth` unchanged. Do NOT add
  `truncated:explain.function_budget`.** A new code would not violate criterion 1
  (existing codes persist), but cross-function expansion adds entries to the same
  flat `diag.Causes` list that `ExplainMaxNodes` already bounds — there is no new
  unbounded dimension, and a second bounding axis would need its own ordering
  contract to stay deterministic. The genuine cross-function risk is that
  `ExplainDefaultDepth = 3` truncates legitimate caller→callee→callee chains;
  that is a default-value question with an existing `--depth` escape hatch, not a
  new code.

- **D-13-22:** Phase 13 also carries the `blame` marker from D-13-01 on
  `ExplainNode`, riding on D-13-14's additive-field decision.

- **D-13-23 (the falsifiable test for DX-05, defeats hardcoding / "use the root's
  function for every node" / single-derivation):** a three-function fixture where
  the cause chain runs caller → callee → callee-of-callee, with a **decoy**
  function whose byte range lies between two visited ones. Assert each node's
  `function_id` equals the innermost AST function containing its span, computed
  by a **test-side resolver written independently** over `parsed.Program` (never
  calling the production core-side resolver). Then the discriminating mutation:
  **reorder the function declarations in the source** so every byte offset shifts
  but no function identity changes — assert all spans move and all `function_id`
  values are unchanged. Additionally: a cause with `Span == nil` must emit **no**
  function field and `availability: not_captured` rather than guessing. And a
  guard-mutation test in the repo's existing fault-injection style: a cause whose
  span exactly equals a function declaration span must parent a later in-body
  cause with `caused_by`, and disabling the function-scope guard must make that
  test **fail**.

### Held-out fixture split (DX-07 criterion 2, DX-06 criterion 3)

- **D-13-24:** Construct the corpus **mechanically** — new marker-driven
  interprocedural injectors in `session_phase6_injectors.go`'s existing
  `Injector` / `markerGuard` shape — run against **hand-authored base programs
  whose call-graph shapes differ by construction**. Rejected: two hand-authored
  disjoint corpora, which repeats M001's documented weakness (its
  `heldout_move_defect.lang` and `derivation_move_defect.lang` are near
  alpha-renamings — `item`→`buffer`, `moved_once`→`delivered` — and
  `TestPhase6DefectCorpusIsHeldOut` asserts only byte-inequality, which any
  rename passes).

- **D-13-25 (the sharpest finding in this discussion — the plan must be built
  around it):** the overfittable artifact in Phase 13 is **not the driver**.
  `repair.go` has no kind→edit table to overfit. It is **`check`'s choice of
  which function's span the repair points at**. Therefore the split must vary the
  dimension a blame rule could overfit: **call-graph topology and
  detection-to-fix hop distance**. A held-out corpus that varies identifiers is
  ceremonial; one that varies topology is not.

- **D-13-26 (split rule, precise):** prefixes `heldout_` / `derivation_` and the
  M001 README contract carry forward, plus three structural obligations, each a
  test:
  1. **Topology disjointness.** For each interprocedural class, compute
     `(function count, call-edge count, detection-to-fix hop distance)` from the
     parsed program; held-out and derivation must **not** be equal. This replaces
     byte-inequality, which alpha-renaming defeats.
  2. **Strictly-harder held-out.** Every held-out interprocedural fixture must
     have **hop distance ≥ 2** (the fix site is *not* the function the
     diagnostic's Primary span lands in); derivation fixtures may be ≤ 1. This
     kills "held-out contains only the easy member of the class".
  3. **Fail-closed baseline.** Every held-out base must `check` clean unmutated
     and produce exactly one diagnostic mutated, with a `DriverEligible` repair.

- **D-13-27 (anti-contamination / one-shot discipline):** seal the held-out
  corpus with a committed `testdata/phase13/HELDOUT.sha256` manifest plus a test
  asserting current bytes match — editing a held-out fixture after a repair rule
  is tuned then becomes a **loud, reviewable diff** rather than silent test-set
  contamination. Extend `TestRepairDriverSourceNeverReferencesHeldoutFixtures`
  beyond `cmd/lang-repair` to any non-test file in the blame-rule path
  (`internal/compiler/check`). **No goldens for held-out**: assertions are
  properties (`outcome == "repaired"`, re-check clean, fix span falls inside the
  expected function), so there is **no `--bless`-shaped regenerate button**. The
  rustc `tests/ui --bless` lesson is that a regenerable expectation is not
  evidence.

- **D-13-28 (criterion 3's fixture MUST be a TWIN PAIR — either half alone is a
  coin flip):** calls are arity-1, so build from a **shared callee with two
  callers**:
  - `fn sink(buffer: Buffer) -> Buffer { let held = take buffer … }` — takes
    ownership.
  - `fn alpha(buffer: Buffer) -> Buffer { let view = borrow buffer; let r = sink(view) … }`
  - `fn beta(buffer: Buffer) -> Buffer { let view = borrow buffer; let r = sink(view) … }`
  - `fn main(buffer: Buffer) -> Buffer { … alpha/beta … }`

  The diagnostic fires at a call site in `alpha`. The naive detection-site repair
  — fix `alpha`'s argument — leaves `beta` still broken, so the program does
  **not** re-check clean. The only repair that makes the whole program clean is
  in `sink` (narrow its parameter to a shared borrow). **The re-check-clean
  oracle, not a golden blame assertion, is what discriminates right blame from
  wrong blame.**

  **But a rule that always blames the callee passes this by luck.** So it ships
  with the **mirror fixture** in the same held-out set, where one caller is legal
  and one illegal, making the unique whole-program-clean fix the **caller**.
  Always-blame-callee fails the twin; always-blame-detection-site fails the
  first. **The pair is the control.**

- **D-13-29 (outcome laundering closed):** assert the **exact outcome string**
  per held-out case. Over the whole corpus assert the observed outcome set
  contains **no `already_clean`** (that means an injector went inert) and **no
  `unrepairable`** for classes claimed as reached. Keep
  `TestUnrepairableDefectFailsTheGate` extended to the new classes so
  `unrepairable` can never read as a pass.

- **D-13-30 (QLT-08 mutation-kill, per new control):**
  (a) each new injector gets a guard-disabled twin mirroring
  `matchInjectSkippingGuard`, plus extension of
  `TestEveryInjectorRefusesWhenMarkerDisappears` and
  `TestInjectorTargetChoiceIsSpecified` to the new markers;
  (b) the **topology-distinctness control** is killed by feeding it an
  alpha-renamed copy of a derivation fixture and asserting red — this is the
  control that proves the split is not ceremonial, so it is the one most in need
  of a not-inert proof;
  (c) the sealed-digest control is killed by a one-byte flip in a temp copy;
  (d) the criterion-3 control is killed by applying the repair at the
  **detection-site** function instead of the compiler-named one and asserting the
  program does **not** re-check clean.

- **D-13-31 (rejected alternative, recorded so it is not re-proposed):**
  bounded-exhaustive / grammar-based generation (Korat, Alloy small-scope,
  csmith/YARPGen) is genuinely tractable given Lang's tiny surface, **but there
  is no oracle for *which function is the correct fix location* in a generated
  program** — only "does it check clean", which is precisely the
  plausible-vs-correct trap APR fell into. Revisit when a differential blame
  oracle exists.

### Driver and scope boundaries

- **D-13-32:** **One repair per pass.** This **confirms shipped behavior and
  requires no change**: `Repair()` already performs exactly one diagnose → at
  most one apply → at most one reverify with **no loop construct anywhere in the
  function** (D-06-30), and `subprocess_count` is structurally always 1 or 2.
  The byte-offset cascade footgun (two repairs in one response invalidating each
  other — the class of bug behind the 2018-edition `cargo fix` incidents, which
  is why rustfix applies only non-overlapping suggestions per pass) is therefore
  **already designed out**. **`cmd/lang-repair` needs no changes in Phase 13.**

- **D-13-33:** Retro-strengthen `testdata/phase6`'s distinctness control as well,
  not only phase 13's. Caveat the planner must handle: M001's fixtures are
  largely **intraprocedural**, so the topology triple from D-13-26 degenerates
  there — a weaker structural predicate is needed for them, and applying it may
  turn currently-green tests red. If it does, that is a **real hole in shipped
  M001 evidence being surfaced**, not a regression to suppress; escalate rather
  than weaken the predicate.

### Claude's Discretion

- Field naming granularity for D-13-14 (`function_id` + `function_name` vs a
  nested `function` object vs SARIF's `fully_qualified_name`) — the latter
  matters only once M003 modules add a namespace segment.
- Whether the `functions` table on `ExplainSummary` ships now or waits for a
  consumer needing whole-function spans (the inline name alone satisfies
  criterion 1).
- Whether `functionByOperationID` is materialized once on `check.Result` or
  rebuilt per diagnostic.
- Whether `ExplainDefaultDepth` stays 3 for cross-function chains.
- Whether class 3's uniqueness gate counts all initialized in-scope places or
  only the function parameter plus prior `let`s bound before the call site.
- New corpus directory (`testdata/phase13/`) vs extending `testdata/phase6/` —
  note `cmd/lang/main.go`'s `isPhase6Corpus` marker-file dispatch would need a
  phase-13 sibling.
- Whether the topology triple is computed from the existing call-graph package or
  re-derived independently (the peer-re-derivation discipline likely demands the
  latter).
- Number of held-out fixtures per class (minimum: the D-13-28 twin pair, so ≥ 2
  for the blame class).
- Whether a non-`Callable` callee is a callee contract violation (B1) or caller
  misuse (B2) — lean B2, but it is the one field where both readings are
  defensible.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project-level durable context (read FIRST — these override roadmap vocabulary)

- `.planning/LANGUAGE-MATURITY.md` — the language is far less expressive than
  the roadmap implies: no arithmetic, no iteration, no `if`, no strings/arrays,
  `Byte`/`Buffer` only. **Every fixture in this phase must be constructible in
  that surface.** Do not read `wiki/example-tour.md` as a description of the
  language.
- `.planning/STANDING-VERDICTS.md` — already-researched verdicts (deps,
  anti-features, the six dispatch sites, why `-flto` is load-bearing).
- `.planning/ROADMAP.md` §"Phase 13: Agent Loop for Interprocedural Defects" —
  goal, three success criteria, riskiest assumption, and the scope-cut note
  (**reinterpreted by D-13-12** — cycle refusal is high-value for
  explainability, not repair).
- `.planning/REQUIREMENTS.md` §"Agent Loop" — DX-05, DX-06, DX-07 verbatim.

### Explain / cause DAG (DX-05)

- `internal/compiler/session/session_phase6_explain.go` — `buildExplainGraph`,
  `explainCorrelationKey` (:105-110), `explainSpanStrictlyContains` (:115-123,
  **the D-13-19 sleeper**), truncation codes (:170, :202), and the node-ordering
  contract pinned by `TestExplainNodeOrderIsStableOnTies`.
- `internal/compiler/protocol/protocol.go:200-265` — `ExplainMaxNodes`,
  `ExplainNode`, `ExplainEdge` (closed `caused_by`/`narrows`/`same_binding`
  vocabulary), `ExplainSummary`, `ExplainDefaultDepth`.
- M001 Phase 6's `lang.explain/0` cause DAG design:
  `.planning/milestones/M001-phases/06-agent-feedback-and-performance-ratification/06-CONTEXT.md`
  and its `06-RESEARCH.md`.

### Diagnostics, identity, and repairs (DX-06, DX-07)

- `internal/compiler/diagnostic/diagnostic.go` — `Span`, `Cause`, `Repair`, the
  closed `Applicability` vocabulary, `NormalizeApplicability`'s refusal to
  default toward driver-eligible, `DriverEligible`, and the sha256 identity
  construction in `Error` / `ErrorWithRepairs` (**including the comment
  explaining why `Repair.Span` is deliberately non-identity-bearing — the exact
  technique D-13-16 relies on**).
- `internal/compiler/check/check.go` — the emission sites: `:1200-1320`
  (borrow-conflict post-assembly repair reconstruction, the model for D-13-09.1),
  `:2136-2153` (`cfgBackEdgeDiagnostic`, D-07-31c), `:2939-3095` (foreign call
  shape, `syntax.fallible_call_not_consumed` at `:3056`,
  `check.call_argument_type_mismatch` at `:2999`), `:3123` (the
  `declare_foreign_symbol` shell repair — **deferred, see below**).
- `internal/compiler/check/check_ordering_stability_test.go` — the pinned
  diagnostic IDs D-13-04 claims are untouched. **Verify empirically.**
- `cmd/lang-repair/repair.go` — the walled-off driver. Package doc explains the
  kind-agnostic design; `Repair()` (`:262-290`) is the D-13-32 single-pass
  structure.
- `cmd/lang-repair/import_boundary_test.go` — the structural lint forbidding
  `internal/` imports. Any new test must not breach it.
- `cmd/lang-repair/antitheater_test.go` — 794 lines proving the repair suite is
  not theater. **Every new class must survive it.**
- `cmd/lang-repair/repair_test.go` — including
  `TestUnrepairableDefectFailsTheGate` and
  `TestRepairDriverSourceNeverReferencesHeldoutFixtures`, both extended by
  D-13-27 / D-13-29.

### Fixture corpus and injectors (DX-07 criterion 2)

- `internal/compiler/session/session_phase6_injectors.go` — DX-04's
  held-out/derivation-split methodology: the `Injector` interface, `markerGuard`,
  the typed `phase6.injector_target_missing` refusal, `matchInjectSkippingGuard`,
  and the not-inert / marker-disappears / target-choice tests that D-13-30
  extends.
- `cmd/lang-repair/testdata/` — the current three-class capture corpus
  (borrow / match / move), `*_diagnose_capture.json` + `*_reverify_capture.json`.
- `testdata/phase6/` — M001's corpus and its **weak** byte-inequality
  `TestPhase6DefectCorpusIsHeldOut`, retro-strengthened per D-13-33.
- `cmd/lang/main.go` — `isPhase6Corpus` marker-file dispatch, which needs a
  phase-13 sibling if a new corpus directory is created.

### Prior-phase decisions this phase depends on

- D-08-25 (`interprocedural_loan_liveness` currently has `Repairs is nil` —
  reversed by D-13-09.1).
- D-08-17 / D-09-31 (the two admission peers cannot fail the same way;
  reverse-engineering one peer's cause shape into the other is the antipattern —
  the basis for D-13-06).
- D-06-24 (`Repair.Span`/`Replacement`/`Applicability` deliberately
  non-identity-bearing).
- D-06-28 (the driver is a protocol consumer, not a library consumer).
- D-06-30 (single-pass repair cycle — the basis for D-13-32).
- D-07-31c (a back edge has no local mechanical edit — the intraprocedural
  statement of D-13-12).
- D-04-03 (callable ⊆ publishable — the declared contract D-13-01 assigns blame
  against).
- D-04-23 (additive `omitempty` fields leave serialized bytes unchanged — the
  precedent for D-13-18).
- D-07-42 (same-package-only seam convention for QLT-08).

### External precedent cited in the decisions above

- Wadler & Findler, *Well-Typed Programs Can't Be Blamed* (ESOP 2009) — the
  formal basis for D-13-01.
- rustc: `Applicability` enum, `MultiSpan` primary/secondary labels,
  `ObligationCauseCode`, and the misattributed-lifetime-error bug class that
  motivated it. rustfix / `cargo fix` 2018-edition incidents → non-overlapping
  suggestions per pass (D-13-32).
- SARIF `logicalLocation{name, kind:"function"}` and `threadFlowLocations`
  `nestingLevel` — the model for D-13-14 and the reinforcement for D-13-19.
- APR overfitting literature (Smith et al., FSE 2015, *Is the cure worse than the
  disease?*) — why the split must be drawn **before** the repairs are written
  (D-13-27).

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- **`spanByOperationID` threading in `checkInterproceduralLoanLiveness`** — the
  blame resolver's `functionByOperationID` map composes directly with it; no new
  plumbing.
- **`buildInterproceduralSummaries`' callee-before-caller topological order** —
  D-13-03's B3 tie-break reuses it, introducing **no new ordering authority**.
- **`borrowConflictDiagnosticPostAssembly` (`check.go:1244`)** — already
  reconstructs statement text from `core.Place.Name`
  (`"let " + target + " = borrow " + source"`). D-13-09.1's reorder replacement
  is the same technique with `CalleeID` added.
- **`Injector` / `markerGuard` / `matchInjectSkippingGuard`** — a proven
  fail-closed mutation harness with an existing not-inert proof. D-13-24 extends
  it rather than inventing a parallel mechanism.
- **`core.Function.ForeignContract`'s additive-`omitempty` precedent (D-04-23)**
  — the documented in-repo basis for D-13-18's no-schema-bump claim.
- **`Repair.Span` excluded from the identity hash while `RepairKinds` is
  included** — the exact identity-bearing / non-identity-bearing separation
  technique D-13-16 relies on, and the fallback shape if `Cause` ever must grow a
  field.

### Established Patterns

- **Peer re-derivation** (`check` vs `corevalidate`) — constrains D-13-06 (blame
  is single-peer, by design) and shapes D-13-15 (function resolution IS
  peer-re-derived, from core table and AST independently).
- **Fail-closed defaults** — `NormalizeApplicability` never defaults toward
  driver-eligible; `markerGuard` refuses rather than returning unmutated source;
  `unrepairable` is never conflated with a pass. D-13-07, D-13-10, and D-13-29
  all inherit this posture.
- **Anti-theater testing** — `antitheater_test.go`,
  `TestInjectorMarkerCountGuardIsNotInert`. D-13-30 is the phase's obligation
  under this pattern.
- **Typed `{Code}`-plus-wrapped-`Err` failures** — `native.ToolError`,
  `evidence.ValidationError`, `session.InjectorError`, `DriverError`. Any new
  typed refusal (e.g. `blame_undetermined`) should match.
- **Structural import boundary** — `cmd/lang-repair` may not import `internal/`;
  it re-declares `driverEligible` rather than importing it. New tests must respect
  this.

### Integration Points

- `internal/compiler/check/check.go` — blame resolver + three new repair emission
  sites. **The bulk of the phase.**
- `internal/compiler/session/session_phase6_explain.go` — function attribution on
  nodes + the `narrows` function-scope guard.
- `internal/compiler/protocol/protocol.go` — two additive struct changes.
- `internal/compiler/session/session_phase6_injectors.go` (or a phase-13 sibling)
  — new interprocedural injectors.
- `testdata/phase13/` — new corpus + `HELDOUT.sha256`.
- **`cmd/lang-repair/` — source unchanged (D-13-32); tests extended only.**

</code_context>

<specifics>
## Specific Ideas

- The user asked for maximum-breadth adversarial research across stakeholder-role
  lenses before each decision, then a single decisive recommendation per decision
  point. Four `gsd-advisor-researcher` agents ran in parallel (blame rule,
  repairable classes, explain schema, held-out split) at `minimal_decisive`
  calibration. **All four independently converged**, most notably on putting
  function attribution on `ExplainNode` rather than `diagnostic.Cause` to avoid
  ID churn. Full tables and rationale: `13-DISCUSSION-LOG.md`.

- The user selected the researcher's recommended option in **every** area, and
  then chose the *stricter* branch on two of three follow-ups (one-repair-per-pass
  over the more permissive batch mode; retro-strengthening M001's corpus rather
  than only phase 13's) — consistent with this project's established
  fail-closed / adversarial-evidence posture.

- **Framing to carry into the plan:** "rustc refuses to auto-fix borrow errors"
  is the first objection a reviewer raises against D-13-09.1. The answer is
  D-13-11 — `take` has no observable runtime effect in Lang, so the reorder is
  semantics-preserving in a way Rust's is not. Lead with it.

</specifics>

<deferred>
## Deferred Ideas

- **`check.go:3123`'s `declare_foreign_symbol` shell repair** — the only repair
  in the tree that advertises a `Kind` it cannot apply (no `Span`, no
  `Replacement`, therefore never `DriverEligible`). Deliberately **left as-is and
  recorded as debt**: it is not one of the three chosen classes, and resolving it
  either way widens scope (completing it makes a fourth repair class; deleting it
  changes an existing diagnostic's published `RepairKinds`, which **is**
  identity-bearing under schema `/1` and would churn that diagnostic's ID). A
  future phase should pick one.

- **`Cause.Span` is identity-bearing while `Repair.Span` deliberately is not** —
  the "a coordinate shift must never move a diagnostic's ID" principle is only
  half-enforced. Not reopened here (D-13-14 routes around it entirely). Recorded
  as debt; it becomes forcing only if M003 modules require producer-side function
  identity on `Cause`.

- **Producer-side function identity on `diagnostic.Cause`** — the right answer
  once M003 modules mean `explain` can no longer see callee source. When landed,
  it must be **non-identity-bearing**, projecting `Causes` to
  `{Kind, Detail, Span}` in the identity struct exactly as `RepairKinds` does for
  repairs. M003 candidate, not Phase 13.

- **Function attribution on raw `lang check` diagnostics** (not just `explain`) —
  possibly wanted by DX-06 consumers later. Phase 13 treats `explain` as the only
  surface that needs it.

- **Bounded-exhaustive / grammar-based fixture generation** — rejected for now
  per D-13-31 (no oracle for correct fix location). Revisit when a differential
  blame oracle exists.

- **Foreign-C boundary causes** — a cause inside `ForeignContract` territory has
  no Lang function. Omit the field, or introduce a distinct sentinel kind. Small
  enough that the planner may settle it; noted so it is not discovered
  mid-execution.

</deferred>

---

*Phase: 13-agent-loop-for-interprocedural-defects*
*Context gathered: 2026-09-13*
