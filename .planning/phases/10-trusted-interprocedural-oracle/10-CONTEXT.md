# Phase 10: Trusted Interprocedural Oracle - Context

**Gathered:** 2026-09-11
**Status:** Ready for planning

<domain>
## Phase Boundary

The **non-admission** side of the stack becomes trustworthy enough to be the
authority everything native is differential-tested against in Phase 11.
Concretely, three things that today do not exist across a function boundary:

1. **`originvalidate` walks published origins across `OpCall`** (TRU-02),
   closing the inherited D-09-51 defect in the same change.
2. **`pathoracle` independently re-derives the cross-function loan-chain rule**
   (TRU-03), without importing `check` or `corevalidate`, and with that
   non-import enforced by a mechanism rather than a comment.
3. **`interp` executes a multi-function program** on a bounded call stack
   (SEM-08), runs drops in a defined order across frames (SEM-09), and derives
   the call-site ownership-transfer fact itself (OWN-05b).

Bound together by **criterion 4's cross-function loan-endpoint differential** at
a **declared, bounded composition depth** (QLT-04), which puts `interp`,
`pathoracle`, `check`, and `corevalidate` on the same enumerated inputs — the
gate that catches this phase's riskiest assumption (that the interpreter's frame
model agrees with the two admission layers about what is legal across a call
boundary) *before* `cgen` is built on top of the oracle in Phase 11.

**Not this phase** (each has its own phase and gate): multi-function C emission
and interprocedural `-O3`/LTO equivalence (Phase 11, NAT-04/05/06/07);
cross-run summary caching / QLT-06 (Phase 11, per D-08-37); the call-graph-shape
reachability register QLT-03 (Phase 11); `Result` payloads (Phase 12); the agent
loop (Phase 13).

**Explicitly re-scoped out of this phase by this discussion:** D-09-53's
`check.deriveFunctionUsesParam` code change. See D-10-27 — the debt register's
diagnosis was falsified by direct execution, and the real gap is a different,
correctly-scoped item.

</domain>

<decisions>
## Implementation Decisions

All twelve gray areas were researched in advisor mode (twelve parallel
`gsd-advisor-researcher` agents, `minimal_decisive` calibration) under the
developer's standing mandate to fan out across stakeholder-role lenses, run an
adversarial pass, draw on cross-ecosystem prior art, and synthesize one-shot
recommendations. The developer directed that the synthesized recommendation be
adopted in every case.

**Every load-bearing factual claim below was independently re-verified against
the shipped tree by the orchestrator before being locked.** Where a researcher
and the tree disagreed, the tree won and the correction is recorded inline. Two
claims in the orchestrator's own briefing to the researchers were wrong and were
corrected by the researchers against the tree (D-10-19, D-10-31). One **locked
Phase 09 debt item is reversed here** (D-10-27 supersedes D-09-53's
characterization); the reversal is stated as a reversal, with the superseded
text named.

### `originvalidate` across `OpCall` — TRU-02 and D-09-51

- **D-10-01 (the defect is real and is Criterion 1, not a side quest):**
  verified in tree — `walkReturnOrigin` (`originvalidate.go:171-218`) switches on
  `operation.Kind` with cases for `OpBorrowExclusive`, `OpBorrowShared`, and
  `OpForeignCall` (which reads `function.ForeignContract.Alias`), and has **no
  `case core.OpCall`**. It therefore treats a call boundary as fully transparent,
  walking through `operation.SourceID` into the argument's own provenance. D-09-51
  is not deferred scope that happens to land here — it **is** Phase 10 Success
  Criterion 1, verbatim.

- **D-10-02 (the widening is a narrow callee-contract map, NOT `core.Program`):**
  precompute a map keyed by callee ID, built **once** in `ValidatePublished`, and
  thread it as an explicit added parameter through
  `RecomputeOriginPerReturn` → `walkReturnOrigin` (and `RecomputeOrigin` /
  `PublishProblemsFor` as needed). **Do not pass `core.Program` down.** Rationale
  is mechanism-over-convention: a `core.Program` parameter makes
  `program.Functions[i].Linear` — the callee's *body* — reachable from inside a
  walk whose whole contract (SEM-05 body-blindness, and this package's own header
  comment) is that it never reads a callee body. With `core.Program` the
  body-blindness guarantee degrades to a comment a later editor can silently
  violate; with a narrow map the guarantee is a property of the **type
  signature**. — **Reversibility:** costly — undoing means re-widening every
  threaded signature and the 9 in-package test call sites again.

- **D-10-03 (this is not a green-field choice — copy the shipped precedent):**
  `corevalidate`'s own D-09-03 fix, landed one phase earlier for the structurally
  identical problem, already chose exactly this shape: `v.peerLoanCarry
  map[string]peerLoanCarryFact` built once per `Validate` call, threaded as a
  plain narrow parameter into `buildLoanChainIndex(operations, checks, loanCarry)`
  (`corevalidate.go:1192`), consulted at the `OpCall` branch via
  `loanCarry[operation.CalleeID]` (`corevalidate.go:1203`) — never a
  whole-`core.Program` handoff. Mirror it.

- **D-10-04 (keep the map's value type minimal — this is the anti-drift control):**
  the value type should carry only the origin fact the walk needs (an
  `Access`/derived-ness pair), **not** the full `core.PublicOrigin` (which
  additionally carries `Paths []string` this walk does not need). Nothing on the
  parameter type should be misusable to reach a body even by a future editor who
  wants to. Pair this with a same-package test that fails if `walkReturnOrigin`'s
  signature ever gains a `core.Function`- or `core.Program`-shaped parameter — a
  signature-text assertion is cheap and matches Criterion 1's own
  "build- or test-level control rather than convention" bar.

- **D-10-05 (blast radius, counted not estimated):** 2 external production call
  sites of `RecomputeOriginPerReturn` (`session/session.go:2641`,
  `session/session_phase7.go:455`), both already holding a `core.Program` in
  scope; 9 test call sites, all in
  `internal/compiler/originvalidate/originvalidate_test.go` (lines 278, 287, 346,
  487, 720, 765, 771, 800, 807); the in-package definitions and internal calls.
  `BuildInterface`/`ValidatePublished`'s own signatures **do not change**, so
  their ~19 call sites across `session`/`check` are untouched. This is a bounded,
  mechanical change — not the cross-cutting hazard the debt register's prose
  implies.

- **D-10-06 (one plan, two commits — widen then use):** land the widening and the
  `OpCall` case in the **same plan**, not a separate preceding plan (the widening
  has no test-observable behavior until the case consumes it), but sequence
  internally as Google-LSC-style widen-then-use: commit 1 adds the parameter and
  builds the map with the `case core.OpCall:` absent or inert, suite green;
  commit 2 adds the real consult and flips the fixture. This satisfies both the
  atomic-commit discipline and red-first testing.

- **D-10-07 (the mirror to `OpForeignCall` is real but NOT exact — say so):**
  both read a *declared* contract at the hop rather than re-walking a body, but
  `OpForeignCall`'s contract lives on the **current** function
  (`function.ForeignContract`, no lookup at all) while `OpCall`'s lives on a
  **different** function and requires the cross-function map. Calling it "the
  already-proven hop" without naming that difference is an overclaim; the plan's
  doc comment must state it.

- **D-10-08 (the gate):** `testdata/phase08/twin_a_accept.lang` moving from a
  `core.origin_omitted` refusal to a clean admit through the **full** `lang check`
  CLI (`session.CheckCommandFile`, fixed precedence check → corevalidate →
  originvalidate). Load-bearingness proven by a `TerminatorKindsOverride`-shaped
  nil-default seam (`originvalidate/export_test.go:14` already establishes this
  discipline in this package) that forces the new `OpCall` case to no-op,
  demonstrating the fixture regresses to the old wrong refusal — not merely that
  the case exists syntactically.

### `pathoracle`'s cross-function model — TRU-03

- **D-10-09 (COMPOSE callee paths; do NOT add a contract hop):** `OpCall` is
  handled by recursing into the callee via pathoracle's **own**
  `EnumeratePaths`/`linearizePath` and splicing the callee's enumerated concrete
  paths into the caller's path, then running the existing forward replay over the
  spliced sequence. — **Reversibility:** one-way — the composed enumeration
  becomes the substrate QLT-04's declared-depth corpus and criterion 4's
  differential are both built on; retreating to a contract hop afterwards
  invalidates the depth declaration and the differential's independence claim
  together.

- **D-10-10 (why, and it is not a style preference):** `check` and `corevalidate`
  have **already** claimed the contract-hop pattern for `OpCall` — verified at
  `corevalidate.go:1203` (`loanCarry[operation.CalleeID].ReturnsBorrowOfParam`)
  and `check.go:585/721/1806-1916`. If `pathoracle` also hops on a callee summary,
  all three procedures branch on `OpCall` and consult a callee-derived fact, and
  TRU-03's "third, structurally distinct decision procedure" claim collapses at
  exactly the one place — call boundaries — where TRU-03 is being asked to prove
  something. Composition is the one design where pathoracle's cross-function
  behavior **is the mechanism it already has, applied recursively**: the existing
  enumerator closing over itself is not a new abstraction.

- **D-10-11 (the property that actually matters, stated precisely):** the
  package-doc property TRU-03 rests on is **not** "touches zero other functions"
  (impossible for any correct interprocedural analysis — `check`'s fixpoint and
  `corevalidate`'s closure both also cross into callee-derived facts). It is
  "never converges, never closes a relation, only enumerates and replays concrete
  paths." Composition preserves that under call-crossing; a contract hop does not,
  because it collapses a callee's per-path distinctions into one summary fact
  **before** replay — precisely the join/reduction step the package doc says this
  package never performs. Write this into the doc comment; it is the first thing
  a reviewer will probe.

- **D-10-12 (two caps, two typed refusals):** `MaxPaths` (existing, per-function,
  `pathoracle.go:46`) stays as-is. Add a **separate** declared composition-depth
  cap with its own typed refusal carrying a `Code()` method, mirroring
  `pathCapError`'s shape (`pathoracle.go:87-110`). Do not fold composition depth
  into `MaxPaths` — one cap doing two jobs cannot report which limit fired.

- **D-10-13 (signature widening, consistent in shape with D-10-02):**
  `RecomputeEndpoints` widens with the **narrowest possible** callee-lookup
  capability, not a bare `core.Program`. Decided independently at this site from
  `originvalidate`'s, but deliberately the same *shape* — least-privilege
  parameter passing at both re-derivers.

- **D-10-14 (the discriminating test that proves this is not cosmetic):**
  construct a callee with **two distinct concrete paths where only one borrows a
  parameter into its return** — a per-path divergence, not a per-function
  property. Under composition the caller's composed paths correctly split into an
  edge endpoint (divergent) vs a point endpoint (non-divergent) depending on which
  callee path is spliced; a `ReturnsBorrowOfParam`-style contract hop necessarily
  reports the same conservative answer for every call site. A test asserting the
  split survives composition but collapses under a stubbed contract hop is the
  mutation-kill proof that the cross-function hop is load-bearing rather than
  inert. **This fixture is a required deliverable**, not an optional extra.

- **D-10-15 (cycle guard):** composition needs its own guard against mutual
  recursion across the call edge, distinct from `EnumeratePaths`'s existing
  intra-function back-edge handling. `callgraph` refuses cycles before this ever
  runs, so this is fail-closed defence for a corrupted core artifact — say so in
  the comment rather than implying it is reachable.

### Import control — Criterion 1's "mechanism rather than convention"

- **D-10-16 (THE CONCRETE GAP, newly found and previously unrecorded):**
  `originvalidate`'s two guard tests —
  `TestOriginValidatorImportsStayIndependent` (`originvalidate_test.go:33`, check
  at line 50) and `TestOriginValidateImportsNeitherCheckNorAst`
  (`originvalidate_test.go:57`, check at line 74) — forbid **only**
  `/compiler/check` and `/compiler/ast`. **Neither forbids `/compiler/corevalidate`.**
  Criterion 1 requires the `OpCall` walk to import "neither `check` **nor
  `corevalidate`**," and the existing mechanism structurally cannot catch a
  `corevalidate` import today. This is the single most concrete finding of the
  fan-out and is a required deliverable. D-09-05's "the mechanism already exists,
  extend its coverage" holds — but its coverage audit was scoped to
  `corevalidate`'s own guards and never checked `originvalidate`'s.

- **D-10-17 (harden from direct-import to transitive):** convert the guard from
  the current `go/parser` per-file `ImportsOnly` directory scan to a
  `go list -deps` transitive check against the same forbidden set. Direct-import
  scanning misses the realistic failure mode: nobody adds
  `import "compiler/check"` to `pathoracle` on purpose; they add a helper package
  that itself imports `check`. No new dependency — `go list` ships with the
  toolchain CI already requires. Apply to `originvalidate` (mandatory, it has the
  real gap) and to `pathoracle`'s `TestOracleImportsStayIndependent`
  (`pathoracle_test.go:51`) as the two re-derivers Criterion 1 names.

- **D-10-18 (test-level IS sufficient here — do not build new machinery):**
  `.github/workflows/ci.yml`'s `checks` job runs `go vet ./...`, `go build ./...`,
  `go test ./...` and `go test -race ./...` on every push and PR, on Linux and
  macOS. A `go test` guard therefore already fails the build. Criterion 1's
  "build- **or** test-level" is satisfied; adding `depguard`, a custom
  `go/analysis` analyzer, or an `internal/` restructure to a repo with zero prior
  linter infrastructure is unjustified machinery. **Do not consolidate** the
  guards into one shared table either — the existing redundancy (four
  near-identical `corevalidate` guards) means multiple deletions are needed to
  blind the mechanism, and that redundancy is worth more here than the
  maintainability win. A guard test is structural, not a production peer, so
  D-09-38's no-shared-helper rule does not forbid consolidation — it is simply not
  worth doing this phase.

- **D-10-19 (an asymmetry to record rather than silently fix):** `originvalidate`
  imports `compiler/callgraph` (`originvalidate.go:18`, calling `callgraph.Order`)
  — a package `check` also imports. `corevalidate`'s guard explicitly **forbids**
  that same import for itself
  (`corevalidate_endpoint_internal_test.go:107`, per D-07-19). The two
  re-derivers are held to different standards today. `callgraph` is a leaf
  utility with no compiler-internal deps besides `core`, so this is defensible —
  but it is defensible by *review*, not by mechanism, and the forbidden lists are
  hand-curated per package. Record it; do not change `originvalidate`'s
  `callgraph` import this phase.

- **D-10-20 (the residual weakness, stated rather than papered over):** every
  guard lives in the package it polices, so a single commit can add the forbidden
  import and edit the assertion together. Transitive checking does not fix this;
  neither would consolidation. Accepted as out of scope for a test-level
  mechanism, consistent with what Criterion 1 asks for. State it in the artifact.

### The interpreter call stack — SEM-08

- **D-10-21 (explicit `[]frame` stack + iterative dispatch; NOT Go recursion):**
  each `OpCall` pushes a heap-allocated frame and loops. This is not a preference:
  Go's stack exhaustion is a fatal runtime **throw**, and `recover()` **cannot**
  catch it (verified against the Go runtime's behavior and multiple `golang/go`
  issue reports). Any design that recurses in Go is therefore architecturally
  incapable of satisfying SEM-08's "a named refusal, **not** a host stack
  overflow." — **Reversibility:** one-way — the frame model is what Phase 11's
  five-axis comparator differentials against; changing it after Phase 11 starts
  moves every native baseline.

- **D-10-22 (the in-repo precedent, and it is load-bearing not decorative):**
  `callgraph.Order` (`callgraph.go:283-400`) is already an **iterative**
  three-color DFS with an explicit stack, and its doc comment (lines 25-26) states
  why: "never native Go recursion, so the compile-time traversal costs no Go
  stack." `interp`'s three execution paths are already iterative loops too
  (`runLinearBlocks` walks `currentID` through a `for {}`). Adopt the shape the
  repo already uses; do not invent a new one.

- **D-10-23 (`MaxCallDepth = 128`, and the reasoning is INVERTED from `MaxPaths`
  — this is the crux):** verified — `syntax/parser.go:16` enforces
  `maxFunctions = 1024`, and `callgraph` refuses every cycle, so an admitted
  program's call graph is a DAG whose longest chain is bounded above by 1024.
  `MaxPaths` (4096) sits **above** its real reachable maximum (64) so it never
  fires today; `MaxCallDepth` must sit **below** the real structural ceiling
  (1024) so that it **does** fire on a legitimately generated program — otherwise
  the refusal is decorative. 128 is comfortably below 1024 (a 129-function chain
  fixture leaves ample headroom for scaffolding) and orders of magnitude above any
  depth today's fixtures exercise. **Write the inversion into the constant's
  rationale comment**; a reader who pattern-matches it to `MaxPaths` will draw the
  wrong conclusion.

- **D-10-24 (exercise the refusal through the REAL pipeline):** the
  depth-exceeded test builds a genuine `.lang` source of 129 chained
  single-call functions and runs it through
  `syntax.Parse` → `check.Program` → `corevalidate.Validate` → `interp.Run`,
  following `checkedCallBasicProgram`'s existing pattern
  (`interp_test.go:30`). **Never** a hand-built `core.Program` that bypasses
  `corevalidate.Validate` — a test that proves the bound by constructing a program
  the real pipeline could never produce proves less than it appears to.

- **D-10-25 (the refusal is an `Outcome`, not a Go `error`):** model it like
  `OpDefect` — a typed outcome serializable into `CanonicalBytes` — **not** like
  `ErrCallUnsupported`'s bare error return. Phase 11's five-axis comparator needs
  it as comparable execution **data** when diffing against native's structurally
  different limit-hit behavior; a Go-side-channel error string is not comparable.

- **D-10-26 (per-frame `values` map):** each frame owns its own
  `values map[string]string`, so a moved-from place in the caller is genuinely
  inaccessible for the call's duration and the callee's places are a disjoint
  namespace seeded only by the passed argument. `Run`'s existing single-frame map
  becomes the stack's base, not a special case. Determinism is preserved because
  ordered output is already produced from ordered slices (`liveOrder`, the block
  walk), never by ranging a map — the plan must assert this explicitly for the new
  code rather than inherit it by inspection.

### D-09-53 — REVERSED

- **D-10-27 (D-09-53's diagnosis is FALSIFIED; this supersedes it):** PHASE-09-DEBT.md's
  D-09-53 records `deriveFunctionUsesParam`'s `default: usesParam = true` arm as
  a *defect* contradicting the function's own doc comment, with
  `TestInterproceduralLivenessTwinPatternB` stated as "UNAFFECTED." An empirical
  audit applied the register's own literal proposed narrowing (exclude
  `core.OpMove`/`core.OpCopy` from the default arm, `check.go:729-730`) in a
  throwaway patch and ran the full suite. Result: **8 named test failures**,
  including `TestInterproceduralLivenessTwinPatternB` itself — the very test the
  register cites as proof the law is sound and unaffected. Also failing:
  `TestDeriveFunctionUsesParamBehaviors`, `TestSummaryDerivationTwoHopChainPropagates`,
  `TestSummaryDerivationIsOnePassPerFunction`, `TestSummaryDerivationRequiresProgramOrder`,
  `TestCostCorpusLeafTemplatesDifferOnlyInTheCallee`, plus an ordering-stability
  baseline and a newly-surfaced `check`/`corevalidate` divergence. Six
  independently-authored, currently-passing Phase 08/09 tests **pin the current
  semantics as intended**. Orchestrator verification of the structural claim:
  `check.go:723-726` exempts **only `case core.OpReturn:`** — the operation
  itself, never the chain — so the doc comment's "the entire parameter-derived
  reference chain leads directly to its own terminating OpReturn" overstates what
  the code implements. **What is actually wrong is `twin_b_accept.lang`'s header
  comment and the doc comment's wording, not the derivation.**

- **D-10-28 (the REAL gap, correctly scoped and re-filed):** even with the
  narrowing applied, Pattern B's split **still cannot be demonstrated end to
  end**, because `corevalidate` has **no `usesParam` peer at all**. Verified
  directly: `corevalidate_peer_liveness.go:38-39` defines `peerLoanCarryFact` with
  exactly one field, `ReturnsBorrowOfParam` — Pattern A's fact, D-09-51's scope.
  `corevalidate`'s liveness walk therefore treats every `OpCall` argument
  unconditionally as a use and refuses the accept twin regardless. The honest debt
  item is **"`corevalidate` has no interprocedural `usesParam` peer,"** which is a
  missing *second detector*, not a broken existing one — and Key Lesson 2's exact
  failure condition (nothing but `check` watches this fact). File it as a new row
  in `PHASE-10-DEBT.md` with an **open** landing phase; do not silently re-point
  D-09-53.

- **D-10-29 (do NOT apply the code change in Phase 10):** a change that makes one
  fixture pass while breaking `TestCostCorpusLeafTemplatesDifferOnlyInTheCallee`
  — a mutation/differential-discipline test — is evidence the "fix" degrades a
  detector, not evidence of a test-update inconvenience. Landing it would be a
  unilateral semantic change to an ownership checker, against the majority of
  existing test evidence, with no second detector watching. **Deliverable
  instead:** correct `twin_b_accept.lang`'s header and
  `deriveFunctionUsesParam`'s doc comment to state the implemented contract (the
  exemption covers a zero-hop direct return of the parameter place, not an
  arbitrary identity chain), keep the observed-verdict assertions exactly as
  Phase 09 left them, and record D-10-28's re-filing.

- **D-10-30 (the no-third-deferral rule does NOT make this a violation):** this is
  not a second deferral of D-09-53. D-09-53's *premise* is withdrawn as
  incorrect, and its real content is re-filed as a differently-scoped item. The
  distinction must be written explicitly into `PHASE-10-DEBT.md` so no future
  reader records it as "deferred twice." The planner must **re-run the throwaway
  patch as a pre-flight** and confirm the 8 failures before writing the
  correction — the orchestrator verified the structural claims and the absence of
  a `corevalidate` peer directly, but did not itself re-execute the patched suite.

### Drop ordering across frames — SEM-09

- **D-10-31 (`OpCall` is not a terminator; there is almost no nonlocal exit — say
  so plainly):** enumerated against `core.go`: `OpReturn` (terminator, ordinary),
  `OpFail` (terminator, but an ordinary typed value propagated one boundary at a
  time — not unwinding), `OpDefect` (effectively total process-abort), the foreign
  process-root landing pad (`interp.go:369`, the one genuine nonlocal exit today,
  currently single-frame), and `OpCall` — which `core.go:559-564` documents as
  **never** a terminator. **Codename Lang has no exceptions and no unwinding.**
  SEM-09's second clause is therefore *narrow*, not vacuous: the only constructs
  that skip more than one frame are the foreign landing pad extended to N frames,
  and the SEM-08 depth refusal if implemented as a runtime abort. **This must be
  stated as a finding in the shipped artifact.** A plan that "satisfies" SEM-09's
  second clause with a test exercising only `OpReturn`/`OpFail` popping has proven
  the *first* clause only.

- **D-10-32 (NO event-schema change — reuse `FunctionID`):** verified —
  `execution.Event` already carries `FunctionID` (`execution/execution.go:45`) on
  every event. Use it for frame attribution rather than adding a frame/depth
  field. The reason is Phase 11: a `-O3 -flto` build can legitimately **inline** a
  callee, at which point there is no runtime frame for native to reproduce a depth
  value from — a depth field would manufacture guaranteed false divergences on the
  five-axis comparator. `FunctionID` is a static, inlining-invariant source fact.
  Place IDs are already function-namespaced, and `callgraph` refuses cycles so no
  function calls itself, making `FunctionID` a sufficient attribution key **for as
  long as recursion stays refused** — state that dependency explicitly.
  — **Reversibility:** one-way — a schema change after Phase 11's baselines exist
  requires regenerating every comparator baseline.

- **D-10-33 (the defined order):** on normal return (`OpReturn`/`OpFail`) the
  callee's frame must already be drained of live resources before it pops; promote
  this to a **`corevalidate`-checked invariant on the callee's signature**, checked
  once per function declaration, **not** re-derived by the caller per call site.
  On the abrupt path, drain is a strict LIFO flattening: innermost frame first,
  reverse-acquisition order within each frame (the existing single-frame rule,
  unchanged), then outward. This composes for free because call frames are already
  LIFO.

- **D-10-34 (`interp` must not be the sole authority on drop order):** the
  `corevalidate` signature invariant (D-10-33) and `cgen`'s multi-frame version of
  its existing single-frame `emitNonlocalPad` should both ship so that a second
  peer independently knows the order. If `cgen`'s multi-frame pad slips to Phase
  11, that is a **named, temporary single-authority gap** recorded in the debt
  register — not a silent acceptance.

- **D-10-35 (observed, not asserted — mechanically):** every ordering test asserts
  against `interp.CanonicalBytes` / `execution.Equal`, **never** against
  `frame.liveResources` or `live` map contents. Add a `frameDrainOrderForTest`
  nil-default unexported seam following `opCallGroupedArmForTest`'s shape
  (`interp.go:23-31`) that flips drain order, and assert the canonical bytes
  diverge.

### OWN-05b — `interp` as the third deriver

- **D-10-36 (confront the import; keep it, and make the boundary a mechanism):**
  verified — `interp` imports `corevalidate` (`interp.go:3-10`) and `Run`'s first
  act is `corevalidate.Validate(program)` (`interp.go:38`). `interp` never calls
  `validated.PeerSignatures()` or `validated.PeerSiteCoverage()` — grepped, zero
  hits — but that is true **by omission, not by enforcement**. Keep the import
  (hoisting `Validate` to the caller would weaken `interp`'s fail-closed posture
  and touch every call site, for an import-graph purity the current one-value
  domain does not need), and add a guard test that **fails the moment `interp`
  reads an ownership-bearing field of `corevalidate.Result`**. That converts
  "happens not to" into "cannot without a red test."

- **D-10-37 (state the narrower claim honestly):** this is **not** the mutual
  non-import independence `check` and `corevalidate` have from each other. The
  defensible claim is *independence of derivation mechanism for the ownership fact
  specifically, nested inside a shared, unrelated validation dependency*. A bug in
  `corevalidate.Validate` would feed `interp` bad input too (Knight & Leveson's
  correlated-fault result). Write the distinction into the debt register; do not
  let "interp is independent" stand unqualified.

- **D-10-38 (the derivation IS the execution behavior):** `interp`'s fact is the
  observable frame-partition behavior at the call boundary — the moved-from place
  deleted from the caller's `values` map, the callee's frame seeded only by the
  transferred value — not a separate classification step. An executor that never
  inspects a `Mode` string at all is structurally incapable of being a
  transcription of `check`/`corevalidate`'s static contract comparison. This is
  the WASM-reference-interpreter / Miri pattern, and `ARCHITECTURE.md:554-561`
  already prescribes it.

- **D-10-39 (one helper across the three arms is allowed; D-09-38 is not violated):**
  `runBranchArm`, `runLinearBlocks`, and `runLinear` each carry their own `OpCall`
  arm (`interp.go:156, 341, 454`). They should share **one** small in-package
  frame-partition helper. D-09-38 forbids a helper built for **cross-peer** reuse;
  three arms inside the same peer, same package, same derivation method are not
  three peers, and three hand-copies of delete-and-reinsert logic is a real drift
  risk. Draw that line explicitly in the doc comment.

- **D-10-40 (the one-value domain makes natural-input agreement near-vacuous —
  say it):** `ParameterContract.Mode` is hardcoded `"owned"` everywhere
  (`corevalidate.go:2163` — note the 09-CONTEXT citation of `:2011-2015` is
  **stale**; the mechanism and its "D-07-01: today's grammar has exactly one
  parameter form" comment are unchanged) and `core.LinearOperation` has no
  override field. A natural-input four-way differential on this fact **cannot
  disagree by construction**. What makes it non-vacuous now is a seeded-fault
  harness: a synthetic `core.Program` with `Mode` flipped to `"shared"`, proving
  (a) the closed-set decode check refuses it before any peer sees it, and (b) if
  that gate were bypassed, exactly which peer diverges. Document criterion 4 that
  way — not as "three engines agree."

- **D-10-41 (the mutant pairing is the actual content of "independent"):** name
  which peer alone catches which mutant. Mutate `runBranchArm`'s `OpMove` to skip
  `delete(values, operation.SourceID)` (move-as-copy) — `check`/`corevalidate`
  execute nothing and are blind; only `interp`'s behavior catches it. Mutate
  `derivePeerSignature` to hardcode `"shared"` — `interp` never consults `Mode`
  and is blind; only the differential against `corevalidate`'s declared contract
  catches it. Requiring this pairing is what makes "independent" falsifiable
  rather than rhetorical, and it is Phase 10's mutation obligation for this fact.

### The Pitfall-4 stack probe

- **D-10-42 (subprocess probe, not in-process measurement):** an in-process
  headroom-ratio measurement is an assertion wearing observation's clothes and is
  **rejected as the gate's evidence** (usable only as a supplementary fast
  sanity check). The probe re-execs the test binary env-guarded (Go's standard
  `TestHelperProcess` idiom — note: **no** subprocess, `SetMaxStack`, or `TestMain`
  pattern exists anywhere in this repo today, so this is genuinely new), the child
  pins a small host ceiling via `debug.SetMaxStack` for determinism across
  platforms, and the parent asserts on exit status and stderr.

- **D-10-43 (the honest finding is the STRONGER one — frame it that way):**
  because D-10-21 makes the interpreter's call stack an explicit heap `[]frame`,
  language call depth consumes **O(1) host stack**. The probe's result is
  therefore not a safety-margin ratio but "the two limits are structurally
  unrelated, here is the measurement." That is a better result than a near-miss
  ratio — and the artifact must **say so explicitly** rather than present it as a
  passed stress test that could have failed. Design the probe to demonstrate
  exactly that: with the language cap disabled via a nil-default unexported
  `maxCallDepthOverride` seam (matching `TerminatorKindsOverride`'s discipline —
  never a production-mutable exported global), the child sails past 100× `MaxCallDepth`
  without host collapse; with the cap enabled, the named refusal fires first.

- **D-10-44 (threshold fixed in source before the first run):** the ratio/depth
  multiplier is committed **before** the first green run, foreclosing the
  "threshold chosen after seeing the result" attack. Review should confirm the
  constant predates the first passing run.

### QLT-04 — the declared composition depth

- **D-10-45 (the definition, adopted from the tree not invented):** composition
  depth is **the number of `OpCall` hops a single loan-or-published-origin's carry
  relation crosses before reaching its terminal use** — a chain of `d+1`
  functions. This is already the load-bearing definition in
  `corevalidate_peer_liveness.go:64-77` and in the `relay_depthN_*` fixture-naming
  convention Phase 08/09 established. Do not invent a competing one.

- **D-10-46 (declare depth = 3):** depth 2 is the established necessity floor
  (D-09-49 Q2 proved depth-1-only is insufficient evidence of transitivity), so
  QLT-04 cannot declare less than 2 without contradicting the project's own prior
  finding. Declaring **3** = necessity-minimum plus one sufficiency margin, in the
  spirit of CBMC's `--unwind N` plus an unwinding assertion (a bound *and* a check
  that raising it finds nothing new). — **Reversibility:** reversible — the depth
  is a declared constant plus corpus; lowering it is a stated scope cut (see
  D-10-55).

- **D-10-47 (the finding that forces depth 3, verified in tree):** there is **no
  depth-3 fixture anywhere** — `grep -rn "depth3\|depth_3"` over `testdata/` and
  `internal/` returns nothing; `testdata/phase08/` contains only depth-1 twins and
  `relay_depth2_*`. Meanwhile `corevalidate_peer_liveness.go:64-77` asserts
  "depth-2, depth-3, and beyond all resolve correctly from the SAME single forward
  pass." That is **prose with no fixture behind it** — exactly the
  implied-by-the-corpus failure Criterion 4's wording exists to prevent. Stopping
  at 2 leaves that claim permanently unverified.

- **D-10-48 (the depth-3 fixture must carry a BORROW through the middle hop):**
  verified — `relay_depth2_accept.lang` composes an **owned pass-through**
  (`leaf` declares `-> Buffer` owned; `relay` forwards it), so the most
  interesting cell — a borrow surviving two hops — is untested even at the depth
  that exists. The new depth-3 fixture pair must vary **which hop carries the
  borrow**, not repeat the owned-passthrough shape; otherwise it is "one more of
  the same" that the uniform per-hop mechanism agrees on regardless of
  correctness.

- **D-10-49 (the product space, with its exclusions NAMED):** the space is small
  and genuinely exhaustible today — `ParameterContract.Mode` is **unreachable at
  anything but `"owned"`** (arity fixed at 1, D-07-01, `core.go:191-198`), so that
  whole dimension collapses to cardinality 1; `ReturnContract.Mode` has 3 legal
  values; `LoanEndpoint.Kind` has 2 (`point`/`edge`, `core.go:762-769`);
  accept/refuse is 2. Roughly a dozen structurally distinct boundary-crossing
  cases, not thousands. **The `Mode` collapse must be stated as a named exclusion
  in the shipped artifact**, never as a silently-pruned dimension.

- **D-10-50 (declared bidirectionally, or it is not declared):** a Go const with a
  `MaxPaths`-style rationale comment **plus** a
  `TestCompositionDepthCorpusReachesDeclaredBound`-shaped test that fails in
  **both** directions — if the corpus silently stops reaching the declared depth,
  **and** if someone raises the constant without extending the corpus. The
  bidirectional check is the actual deliverable; a bare constant is a description,
  not a declaration.

### Criterion 4 — the four-way differential

- **D-10-51 (EXTEND `session_peer_gate_test.go`; do not build a second harness):**
  D-09-23 already ratified this tradeoff — two gates that can drift on what
  divergence means is the same two-derivations-one-truth failure the phase exists
  to prevent. Verified: `TestNoUndeclaredCheckPeerDivergenceAcrossCorpus`
  (`session_peer_gate_test.go:226-279`) compares exactly two peers today
  (`session.CheckFile` vs `corevalidate.Validate`).

- **D-10-52 (the common projection already exists — `core.LoanEndpoint`):**
  verified — three of the four peers already speak the identical shape
  (`core.go:762-769`): `check` via `materializeLoanEndpoints` (`check.go:1726`),
  `corevalidate` via its already-implemented but **unexported**
  `(*validator).recomputeLoanEndpoints` (`corevalidate.go:1366`), and
  `pathoracle.RecomputeEndpoints` (already exported). The only missing seam is an
  exported `corevalidate.Result` accessor. Comparing only admit/refuse would be a
  **regression** from information already in the codebase.

- **D-10-53 (`interp` is accept-side only — bill it that way, always):** `interp`
  cannot produce a `LoanEndpoint` set; its answer is an ordered `Execution`. Its
  comparison is **metamorphic** (endpoint order → observed event order), not
  equality, and it exists only for programs the admission layers **accept**. The
  harness and every artifact must say **"three-way on refuse, four-way on
  accept,"** never "four-way" unqualified — matching D-09-37's precedent on
  overclaiming. `interp`'s one genuine contribution is frame-model/static-verdict
  agreement, which is exactly the phase's riskiest assumption; name it as its own
  separate claim rather than inflating its billing.

- **D-10-54 (upgrade `peerDivergenceExpected` from a bare map to an accountable
  one):** each entry becomes a struct carrying a **debt-register ID and landing
  phase**, with a companion test asserting every ID resolves to a still-open
  `PHASE-NN-DEBT.md` entry. An allowlist that grows silently is how a gate dies;
  this formalizes the discipline the existing comments already practice
  informally.

- **D-10-55 (the seeded fault, per D-09-25's pattern, per new peer pair):** an
  **unexported, package-scoped** seam inside `pathoracle` (and later `interp`'s
  frame teardown) that only that package can reach, plus the companion assertion
  proving the **other** peers' independently-derived endpoint sets are byte-for-byte
  unchanged while only the seeded peer disagrees. Go's package boundary makes
  cross-package reach structurally impossible — that is the actual proof of
  independence, not an assertion that happens to pass. Never seed at a point all
  four peers read from.

### `interp` stability — the Phase 11 precondition

- **D-10-56 (SEM-09 lands FIRST, then the freeze — ordering is non-negotiable):**
  the freeze must come **after** the drop-ordering work, or it is broken inside
  Phase 10 itself. D-10-32's no-new-field decision substantially de-risks this,
  but the ordering rule stands regardless: no freeze is declared until SEM-09's
  event emission is final. **This is the top review risk for the phase.**

- **D-10-57 (a mechanically-checked freeze, not a paragraph):** (1) freeze the
  `Schema` value and JSON shape via a golden corpus under
  `testdata/phase10/interp_oracle/` with a `TestInterpOracleGoldenCorpus` failing
  byte-for-byte on drift; (2) a named regeneration procedure requiring a
  **reviewed commit touching the golden files with a stated rationale** — never a
  bare `-update` flag a CI failure can silently absorb; (3)
  `TestInterpDeterministicAcrossRuns` at `-count=10` (Go's map-iteration
  randomization makes this genuinely informative); (4) a named escalation path for
  Phase 11 — a versioned schema bump plus regenerated goldens plus an explicit
  decision entry, never an in-place edit to the frozen goldens.

- **D-10-58 (the coverage floor is structural, not a percentage):**
  `interp_test.go` today is 97 lines containing **exactly one** test function
  (`TestOpCallGroupedArmMutationKilled`) — verified — for the component about to
  roughly double in size and become the authority for everything native. The
  floor: every `core.OperationKind` exercised through `Run` at **each** of the
  three execution paths, every refusal path reached by a named test, and
  Mutation-Kill Register rows for the new call-stack / cross-frame / ownership
  machinery. This must be an actual enumerated table in the phase's validation
  artifact, in the house form Phase 09 already ships — not a coverage number.

### Cross-cutting

- **D-10-59 (declared scope-cut trigger, in writing before planning, per Key
  Lesson 4):** verified — `ROADMAP.md:198-200` declares the milestone 2× trigger
  for **Phases 08 and 09 only**; Phase 10 inherits none and must declare its own.
  Phase 09's trigger was genuinely checked mid-phase (plan 09-08's own agenda item
  (f), `09-08-PLAN.md:177`) and measured **0.17×** (76,385 actual vs ~460,000
  estimated, `09-08-SUMMARY.md:151`) — a token-cost ratio, not a plan-count ratio
  (Phase 09 ran 10 plans; a naive plan-count reading would have misfired).

  > **Trigger:** if Phase 10's actual token cost (summed across completed plans)
  > exceeds **~2×** the initial per-plan token-estimate baseline for the same
  > plans — measured at a **mandatory mid-phase gate plan**, scheduled after the
  > `originvalidate` ∥ `pathoracle` ∥ `interp`-call-stack parallel wave lands and
  > **before** criterion 4's differential work begins — that is the trigger.
  >
  > **Cut order:** (1) QLT-04's declared composition depth reduced from 3 to 2,
  > with the reduced depth stated **verbatim** in the shipped artifact (Criterion
  > 4's own "stated rather than implied" requirement), remaining depth landing as
  > named debt in Phase 11; then (2) the `OpForeignCall`-adjacent slice of TRU-02,
  > deferred with a named Phase 11 fallback, mirroring D-09-21; then, only if
  > still over budget, (3) OWN-05b's assertion folded into criterion 4's
  > differential rather than standing alone — a consolidation permitted **only**
  > after coverage is provably preserved.
  >
  > **Never cut:** criterion 4's four-way differential **existence** (cutting its
  > depth is a scope cut; cutting its existence makes the phase's own gate
  > decorative); the `interp` stability freeze at full declared strength
  > (disqualified on Phase-11-dependency grounds alone); the Pitfall-4
  > stack probe (a named Gate in the roadmap's own criterion-2 text); **D-09-51's
  > `originvalidate` `OpCall` fix** (it is not carried scope — it *is* Success
  > Criterion 1 verbatim); and the seeded faults and their companion assertions
  > (D-10-14, D-10-41, D-10-55).
  >
  > Any cut is a deferral with a named landing phase recorded in
  > `PHASE-10-DEBT.md` at the gate's own commit, never a silent drop.

- **D-10-60 (new milestone-wide discipline, declared here):** an item that would
  be deferred a **second** time — any `D-NN-xx` row whose `Landing phase` cell
  would be rewritten to point past the phase it already named — must either be cut
  from the milestone explicitly via a REQUIREMENTS.md amendment, or becomes
  automatically never-cut. It cannot silently acquire a third landing phase.
  Verified: `TestDebtRegistersAreWellFormed` (`session/session_test.go:2613-2671`)
  enforces the register's **format only** — it does not track deferral hop count —
  so this is a declared rule, not something the suite will catch.

- **D-10-61 (`PHASE-10-DEBT.md` is written at PLANNING time, not phase end):**
  following D-09-44's precedent, in the mechanically-checked format. It must
  carry at minimum: D-10-27/28/29/30 (the D-09-53 reversal and re-filing),
  D-10-19 (the `callgraph` import asymmetry), D-10-20 (the guard's residual
  bypassability), D-10-31 (SEM-09's narrow second clause), D-10-34 (any
  single-authority gap if `cgen`'s multi-frame pad slips), D-10-37 (the narrowed
  independence claim for `interp`), D-10-40 (the one-value-domain vacuity),
  D-10-59 (the cut trigger), and D-10-60 (the no-third-deferral rule).

### Claude's Discretion

The developer directed that the synthesized recommendation be adopted for all
twelve gray areas, under the standing mandate to fan out across stakeholder-role
lenses, run an adversarial pass, draw on cross-ecosystem prior art, and
synthesize one-shot recommendations. Every decision above is therefore Claude's
synthesis under that instruction, grounded in twelve advisor research returns and
independently verified against the shipped tree.

Planner discretion remains over:

- Plan decomposition and wave ordering, subject to four hard constraints:
  **SEM-09 before the stability freeze** (D-10-56); the **mid-phase gate plan
  scheduled before criterion 4's differential work** (D-10-59); each differential
  landing in the **same plan** as the peer it tests, written red-first (D-09-27's
  inherited precedent); and **widen-then-use** as two commits inside one plan for
  `originvalidate` (D-10-06).
- Exact Go identifier names for the frame type, the depth constant's siblings,
  the composition-depth constant, the new seams, and the callee-contract map's
  value type.
- Whether `pathoracle`'s composition lives in `pathoracle.go` or a sibling file
  in the same package.
- The order in which the two re-derivers (`originvalidate`, `pathoracle`) are
  landed relative to `interp`'s call stack — they are independent at the
  implementation level per the roadmap's Parallel note.
- Whether the `corevalidate` `usesParam` peer re-filed by D-10-28 gets a proposed
  landing phase now or is left open — either is honest; recording it is required.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase charter and requirements
- `.planning/ROADMAP.md` — Phase 10 section (goal, 4 success criteria, riskiest
  assumption, the Parallel note requiring `interp` be *stable* before Phase 11);
  and lines 196-205 for the milestone 2× trigger's **scope** (Phases 08/09 only).
- `.planning/REQUIREMENTS.md` — SEM-08, SEM-09 (lines 28-32); OWN-05b (lines
  41-47); TRU-02, TRU-03 (lines 84-88); QLT-04 (line 107); and the traceability
  table rows at lines 200-216.
- `.planning/research/ARCHITECTURE.md` §5 "Interpreter Call Stack" (line 504
  onward — bounded depth, frame representation, drop ordering, ownership state at
  the call boundary), §"Site 3 — `originvalidate`" (line 89), §"Site 6 —
  `pathoracle`" (line 169).

### Inherited debt this phase must close or correct
- `.planning/phases/09-peer-re-derivation-and-d-03-02-closure/PHASE-09-DEBT.md`
  — **D-09-51** (the `originvalidate` `OpCall` gap; closed here by D-10-01
  through D-10-08) and **D-09-53** (whose diagnosis is **reversed** here by
  D-10-27; read the original text before writing the correction).
- `.planning/phases/09-peer-re-derivation-and-d-03-02-closure/09-CONTEXT.md`
  — D-09-05 (import-guard mechanism already exists), D-09-23/25/27 (the TRU-04
  differential, its seeded fault, the companion assertion, same-plan/red-first),
  D-09-34 through D-09-38 (OWN-05a, non-expressibility, **no shared helper for
  `interp`**), D-09-43 (the Phase 09 cut trigger's form), D-09-44 (debt register
  at planning time), D-09-49 Q2 (depth-1 proven insufficient).

### The base enumeration QLT-04 rebuilds
- `.planning/milestones/M001-phases/03-borrowed-views-and-cfg-lifetimes/03-VALIDATION.md`
  — M001 Phase 3's exhaustive endpoint enumeration: the Generator Reachability
  Register (lines 119-124) and the five-row conflict matrix (lines 64-77).

### Language and standing verdicts
- `.planning/LANGUAGE-MATURITY.md` — the language is far less expressive than the
  roadmap vocabulary implies. Read before assuming any construct exists.
- `.planning/STANDING-VERDICTS.md` — already-researched verdicts (the six
  dispatch sites, why `-flto` is load-bearing, anti-features).

### Source files this phase changes or must not break
- `internal/compiler/interp/interp.go` — the whole file; three execution paths,
  the `OpCall` arms at 156/341/454, `opCallGroupedArmForTest` at 23-31,
  `nonlocalExitDefectReason` at 369, `CanonicalBytes` at 411.
- `internal/compiler/execution/execution.go` — `Schema0` (line 9), `Event` with
  `FunctionID` (line 45), `Execution` (line 53). **The Phase 11 comparator
  contract.**
- `internal/compiler/originvalidate/originvalidate.go` — `RecomputeOriginPerReturn`
  (135), `walkReturnOrigin` (171-218), `RecomputeOrigin` (236), `PublishProblemsFor`
  (357), `ValidatePublished` (405); the `callgraph` import at line 18.
- `internal/compiler/originvalidate/originvalidate_test.go` — the two guard tests
  at 33 and 57 (checks at 50 and 74) that **omit `corevalidate`**.
- `internal/compiler/pathoracle/pathoracle.go` — package doc 1-28 (the identity
  claim), `MaxPaths` 33-46, `pathCapError` 87-110, `EnumeratePaths` 219,
  `RecomputeEndpoints` 354, `TerminatorKindsOverride` 59.
- `internal/compiler/callgraph/callgraph.go` — `Order`'s iterative DFS (283-400)
  and its "costs no Go stack" rationale (25-26); cycle refusal.
- `internal/compiler/corevalidate/corevalidate_peer_liveness.go` — `peerLoanCarryFact`
  (38-39, the **only** field is `ReturnsBorrowOfParam`), `derivePeerLoanCarry`,
  and the unverified "depth-3 and beyond" prose at 64-77.
- `internal/compiler/corevalidate/corevalidate.go` — `recomputeLoanEndpoints`
  (1366, unexported), `buildLoanChainIndex` (1192) and its `loanCarry` consult
  (1203), `derivePeerSignature`'s hardcoded `Mode: "owned"` (2163).
- `internal/compiler/check/check.go` — `deriveFunctionUsesParam` (700-741, the
  `OpReturn`-only exemption at 723-726), `materializeLoanEndpoints` (1726).
- `internal/compiler/session/session_peer_gate_test.go` — `peerDivergenceExpected`
  and `TestNoUndeclaredCheckPeerDivergenceAcrossCorpus` (226-279).
- `internal/compiler/syntax/parser.go` — `maxFunctions = 1024` (line 16),
  `maxArmsPerMatch` (521).
- `internal/compiler/core/core.go` — `LoanEndpoint` (762-769), `ParameterContract`/
  `ReturnContract` (191-198, 236-284), `OpCall` is-not-a-terminator doc (559-564),
  `TerminatorKinds()` (653-655).
- `testdata/phase08/` — `twin_a_accept.lang` (D-10-08's gate),
  `twin_b_accept.lang` (D-10-29's comment correction),
  `relay_depth2_accept.lang` (the owned-passthrough shape D-10-48 must not
  repeat).
- `.github/workflows/ci.yml` — the `checks` job that makes test-level guards
  build-failing (D-10-18).

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- **`corevalidate`'s `peerLoanCarry` threading pattern** (`corevalidate.go:177,
  1192, 1203`): a precomputed narrow map built once and passed as an explicit
  parameter. This is the shipped precedent `originvalidate`'s widening copies
  (D-10-03).
- **`callgraph.Order`'s iterative explicit-stack DFS** (`callgraph.go:283-400`):
  the in-repo shape for "traverse without consuming Go stack," and the direct
  model for `interp`'s frame stack (D-10-22).
- **`pathoracle`'s cap-plus-typed-error idiom** (`MaxPaths` + `pathCapError`
  with `Code()`, `pathoracle.go:33-110`): the house form for every bound this
  phase declares — but note D-10-23's inversion of its *direction*.
- **Nil-default unexported injection seams** (`TerminatorKindsOverride`
  `pathoracle.go:59`; `opCallGroupedArmForTest` `interp.go:23-31`;
  `deriveFunctionUsesParamObserved` `check.go:701`): the established way to make
  a limit or a path testable without a production-mutable exported global.
- **`core.LoanEndpoint`** (`core.go:762-769`): already the common answer shape for
  three of criterion 4's four peers (D-10-52).
- **`session_peer_gate_test.go`'s corpus walker**: the differential substrate to
  extend rather than duplicate (D-10-51).
- **`checkedCallBasicProgram`** (`interp_test.go:30`): the full-pipeline fixture
  pattern the depth-exceeded test must follow (D-10-24).

### Established Patterns
- **Independence = derivation method + import boundary**, never file layout.
  Guards are `go/parser` directory scans skipping `_test.go` files, per package.
- **Bounds are fail-closed and typed**, never truncating, with the real reachable
  maximum stated alongside the chosen cap in the rationale comment.
- **Differentials land in the same plan as the peer they test**, written red-first,
  with a seeded fault plus a companion assertion discriminating independence from
  mere change-detection (D-09-25/27).
- **Debt registers are written at planning time** in a mechanically-checked
  format, and every deferral names a landing phase.
- **Overclaiming is a recorded failure mode** — OWN-05's split into 05a/05b
  (D-09-37) is the precedent D-10-53's "three-way on refuse, four-way on accept"
  billing follows.

### Integration Points
- `session.CheckCommandFile`'s fixed precedence (check → corevalidate →
  originvalidate) is where D-10-08's gate is observed, and is the pipeline whose
  masking chain produced D-09-51 and D-09-53 in the first place.
- `execution.Event`/`Schema0` is the contract Phase 11's five-axis comparator
  consumes — every SEM-09 decision is constrained by what native `-O3 -flto` can
  faithfully reproduce (D-10-32).
- `interp.Run`'s `corevalidate.Validate` call is both `interp`'s fail-closed
  posture and OWN-05b's independence tension (D-10-36).

</code_context>

<specifics>
## Specific Ideas

- **State the `MaxCallDepth` inversion loudly** (D-10-23): a reader who
  pattern-matches it to `MaxPaths` will conclude the cap should sit *above* the
  reachable maximum and make the refusal decorative. The comment must say the
  direction is deliberately opposite and why.
- **`pathoracle`'s new doc comment must distinguish "enumerates and replays" from
  "touches only its own function"** (D-10-11) — the first is the property TRU-03
  rests on, the second is the property a reviewer will *think* it rests on, and
  composition preserves only the first.
- **Bill criterion 4 as "three-way on refuse, four-way on accept" everywhere**
  (D-10-53), including in plan titles and summaries, not only in a footnote.
- **The Pitfall-4 probe's honest result is the stronger one** (D-10-43) — write it
  up as a structural-decoupling measurement, never as a dramatic near-miss.
- **Re-run D-09-53's throwaway patch as a planning pre-flight** (D-10-30) before
  writing the correction, so the reversal rests on a re-executed result rather
  than on this document.

</specifics>

<deferred>
## Deferred Ideas

- **A `corevalidate` interprocedural `usesParam` peer** — newly identified here
  (D-10-28) as the real content behind D-09-53. A *missing second detector*, not
  a broken existing one; Key Lesson 2's exact exposure. Landing phase open;
  record it in `PHASE-10-DEBT.md` rather than assigning it under time pressure.
- **Multi-function `cgen` and interprocedural `-O3`/LTO equivalence** — Phase 11
  (NAT-04 through NAT-07). `cgen`'s multi-frame nonlocal pad (D-10-34) is the one
  piece that would ideally land here; if it slips, it is a named single-authority
  gap.
- **QLT-03's call-graph-shape reachability register** — Phase 11.
- **Persistent cross-run summary caching / QLT-06** — Phase 11, per D-08-37 and
  D-09-06.
- **`Result` payloads** — Phase 12. Note D-10-32's `FunctionID`-as-frame-identity
  decision is valid only while recursion stays refused; if Phase 12's `Result`
  match arms ever admit a recursive shape, that decision re-opens.
- **`originvalidate`'s `callgraph` import** (D-10-19) — the asymmetry with
  `corevalidate`'s forbidden-list is recorded, not resolved. Revisit only if
  `callgraph` ever gains a compiler-internal dependency beyond `core`.
- **Guard-test self-deletion resistance** (D-10-20) — a guard living in the
  package it polices can be weakened in the same commit as the violation.
  Accepted for a test-level mechanism; would need an out-of-package or CI-owned
  control to close.
- **`.planning/spikes` registry maintenance** (D-08-43, spike 006's missing row)
  — still un-owned by any phase, carried forward from Phase 09's deferred list.

</deferred>

---

*Phase: 10-Trusted Interprocedural Oracle*
*Context gathered: 2026-09-11*
