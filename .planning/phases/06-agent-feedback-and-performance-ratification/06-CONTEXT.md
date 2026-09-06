# Phase 6: Agent Feedback and Performance Ratification - Context

**Gathered:** 2026-09-06
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 6 delivers the **agent-facing protocol layer over the already-shipped
compiler**. It adds two new inspection commands (`explain`, `query`), makes
`verify` select evidence lanes by changed risk over a local artifact cache,
extends `evidence` with on-demand trace expansion, ships a protocol-only
repair exercise, and ratifies cost budgets with declared-machine reporting.

Requirements delivered: **FND-04, DX-02, DX-03, DX-04, QLT-02**.

**This phase is not language surface.** No new `OperationKind`, no new syntax,
no `OpCall` (deferred to M002 as its lead charter item, D-05-32/D-05-33). The
`.lang` language M001 ships is frozen as of Phase 5: one parameter per
function, no Lang-to-Lang calls, affine ownership, shared/exclusive borrows
with CFG last-use liveness, typed failure via ok/err edges, `OpRelease`
cleanup, exhaustive match, and declared C foreign contracts.

</domain>

<decisions>
## Implementation Decisions

All four gray areas were researched in advisor mode (four parallel
`gsd-advisor-researcher` agents, `minimal_decisive` calibration) with a
deliberate breadth pass across stakeholder-role lenses, an adversarial pass,
and cited ecosystem prior art. The developer selected the recommended option
in every case.

### `query` / `explain` addressing surface (DX-02)

- **D-06-01:** `query` is **one joined addressing surface**, not per-kind
  subcommands. It resolves any of the five stable ID vocabularies that already
  exist in the tree — `diagnostic:<hex>`, debugmap's `core_id` /
  `operation_id` / `point_id`, evidence IDs and content digests, `control:*`,
  and `lane:*` — so a caller never has to know which subcommand matches which
  ID prefix. **Do not mint a sixth ID vocabulary.** The join model is Kythe's
  (one canonical identity encoding used as the join key on every edge), and
  the in-tree precedent is `debug-map`'s existing span → `core_id` →
  `operation_id` join. — **Reversibility:** costly — the addressing grammar is
  a published agent-facing contract; changing it later invalidates every agent
  workflow and every recorded evidence reference built on it.

- **D-06-02:** `explain` **synthesizes a bounded cause DAG on demand** from the
  flat `diagnostic.Causes` list plus span/binding correlation. The graph is
  **computed per cold invocation, never persisted** — no graph store, no
  daemon, no new persisted core type. Nodes are
  `{id, kind, detail, span, availability}`, reusing the existing honest
  `not_captured` / `optimized_out` availability vocabulary rather than
  fabricating a value. Edges are typed: `caused_by`, `narrows`, `same_binding`.
  Because the graph is synthesized, **determinism across cold invocations is a
  testable obligation** and must be tested.

- **D-06-03:** Bounding discipline. `explain` takes `--depth=N` (default 3)
  plus a hard node-count budget; `query` returns bounded lists paginated by
  `--cursor`, not by depth. Both truncate with a **stable truncation code**,
  mirroring the shipped 64 KiB-plus-one output-bound discipline. Expansion is
  explicit and opt-in — smallest sufficient context by default.

- **D-06-04:** Schema minting. `explain` and `query` are **net-new documents**
  and get net-new `/0` schemas: **`lang.explain/0`** and **`lang.query/0`**,
  each with a content-derived `id` following the existing `Finalize()` pattern.
  They are NOT crammed into `lang.command/0`. Rule of thumb established here:
  *a new capability gets a new `/0` schema; only extending an existing
  document's fields earns a `/1` bump.* — **Reversibility:** one-way — these
  are published versioned product-API schemas; per the standing project rule,
  `/0` bytes are frozen once shipped and can only be superseded additively by
  a `/1`.

- **D-06-05:** Command surface (concrete):
  ```
  lang explain <id> [--depth=N] [--json]
  lang query <id-or-pattern> [--kind=symbol|type|ownership|dependency|test] [--depth=N] [--cursor=C] [--json]
  ```
  Rejected alternative, recorded so it is not re-litigated: per-kind
  subcommands (`query symbol`, `query type`, …). Rejected because it
  reproduces the subcommand-sprawl anti-pattern — the agent must already know
  what it is looking for in order to route the call — and because it produces
  no cause graph at all, which SC1 names directly. SARIF's `graphs` /
  `graphTraversals` were studied as the cautionary case: expressive, and in
  practice almost never emitted, because the traversal model is heavy to
  author and consume.

### Change-risk selection and cache substrate (DX-03)

- **D-06-06:** **Cache artifacts, never verdicts.** The cache stores expensive
  intermediates only — compiled binaries, instrumented (ASan/UBSan) binaries,
  per-mutant binaries. The checker, the five-axis comparator, and the sanitizer
  classifier **always re-run fresh** against whatever binary (cached or newly
  built) is on disk this invocation. A cache hit therefore changes *what is
  skipped* (recompilation) and never *what is asserted*. This is the single
  structural property that closes the stale-green hole, and it is the load-
  bearing constraint of this whole area. — **Reversibility:** one-way — caching
  a verdict even once creates a false-proof vector that no later refactor can
  retroactively remove from already-published evidence.

- **D-06-07:** Cache key = an explicitly **declared** input list, hashed into
  one content ID per artifact (Go's `ActionID` is the model):
  - fixture source bytes
  - declared build flags / opt-level / target triple
  - Clang binary identity — a **probed digest**, not merely the `--version`
    string (the shipped tool-identity probe already does this kind of work)
  - linked ASan / UBSan / libc++ runtime identity where applicable
  - the frozen C translation-unit bytes for FFI cases
  - **the mutation runner's own source hash** — a mutant binary built by a
    changed mutator MUST miss the cache
  - the Go toolchain version that built the harness

  **Anything not on this list is an escape by construction.** Document the
  list; do not claim completeness.

- **D-06-08:** "Changed" is measured against a **gitignored local state file**
  recording last-seen input-hash per (fixture × lane) — not `git diff`, not a
  committed baseline. Rationale: a no-diff commit can still change effective
  risk through toolchain drift, and a committed baseline would import
  cross-machine staleness. First run on any machine is cold and is reported as
  cold.

- **D-06-09:** On-disk layout `$XDG_CACHE_HOME/lang-verify/<sha256>/{artifact,
  meta.json}`, sharded by the first two hex characters. `meta.json` records the
  **full declared-input list** used to compute the key — the transparency
  analogue of `go build -x` / `GODEBUG=gocachehash`. Local only. No remote
  cache (explicitly out of scope for M001), no daemon, no persistent service.

- **D-06-10:** Risk→lane selection is a **static declared table**, not a
  derived dependency graph: `internal/compiler/session/risk_lanes.json`,
  following the `session/qlt01_registry.json` checked-in-registry precedent,
  with an executable audit. Justification: M001 has no call graph, no modules,
  and no interprocedural edges to invalidate — a rustc query graph or a Bazel
  action graph would be enormous over-engineering.

- **D-06-11:** **Conservative widening / fail-closed.** Any fixture whose kind
  is not exhaustively classified, or whose declared-input set cannot be fully
  computed (e.g. the Clang identity probe fails), **selects all its lanes** and
  is treated as `not_cacheable` for that run. Ambiguity and silence always
  resolve to "run it," never to "skip it."

- **D-06-12:** Reporting vocabulary. `protocol.Lane` gains
  `cache_status` ∈ `artifact_reused | artifact_recomputed | not_cacheable |
  unavailable` — deliberately **not** `hit`/`miss`, which would wrongly imply
  a verdict was cached — plus `selection_reason`
  (`selected: <kind> matched` / `deferred: no declared dependency` /
  `widened: undeclared-input risk`). **A lane that was not run renders
  `status: "deferred"`, structurally never `pass`.** `protocol.Metrics` gains
  `cache_inputs_reused_count`, denominated in the same counted-unit currency as
  `recomputed_work`.

- **D-06-13:** Documented soundness argument and its known holes (recorded, not
  hidden, per house style). The argument is an *artifact*-cache argument, not a
  test-selection argument: identical declared inputs ⇒ identical intermediate
  artifact. Known escapes to record in the phase's debt/escape register:
  (1) undeclared environment — locale, `ulimit`, filesystem case-sensitivity;
  (2) a Clang change that does not alter its reported version string (the
  ccache `__TIME__`-class footgun); (3) any future nondeterministic codegen
  silently breaking the "same inputs ⇒ same artifact" premise; (4) a
  hand-edited or partially deleted cache directory being indistinguishable
  from a cold one — there is no integrity check beyond content-hash lookup,
  consistent with the existing stance that SHA-256 is content identity, never
  proof.

### Budgets, declared machines, and metric honesty (FND-04, QLT-02)

- **D-06-14:** **`recomputed_work` is the only hard gate.** It is deterministic,
  machine-independent, and already pinned to exact constants across the tree
  (`16n+13`, `LinearWorkLimit = 16*facts+14`, `RecomputedWork` exactly 8, the
  403-check constant). Gating on it requires **zero statistics**. This is the
  project's own equivalent of rustc-perf's `instructions:u`, chosen for the
  same reason: small regressions are unreliable to detect in wall time, and a
  deterministic counter has no noise floor to fight.

- **D-06-15:** Wall-clock cold/warm distributions are **ratified as bounded
  observations, never as hard p95 gates**. Rejected alternative, recorded:
  a full statistical wall-clock gate (LNT/Talos-style with CoV/bootstrap and
  quarantine). Rejected because the project has exactly one laptop-class Apple-
  silicon host with no CI fleet — battery state, P/E core scheduling, and
  thermal throttling all confound into a single sample stream, there is no
  cross-machine averaging, and **there is no prior distribution from which to
  set a credible first threshold.** Gating on it would produce precisely the
  "noisy microbenchmark thresholds that fail every edit" that
  `wiki/compute-efficiency-constitution.md` forbids.

- **D-06-16:** Budgets live in a checked-in **`session/qlt02_budget_manifest.json`**,
  mirroring the `qlt01_registry.json` precedent — a flat JSON registry, not Go
  constants (the JSON diff is the review surface) and not the wiki's
  `budget { … }` DSL (whose own text calls those numbers hypotheses; too early
  to build a parser in M001). Row shape:
  `{machine_id, metric, gate_type: "hard"|"observed", value_or_bound, unit,
  ratified_at, ratified_by_commit}`. An **executable audit** parallel to the
  QLT-01 audit cross-checks declared `machine_id`s against live probe output,
  so a silent "edit the number until it is green" cannot pass unreviewed.

- **D-06-17:** A **declared machine** is a self-describing probe, not a
  free-text host-class name. `machine_id` = short hash of
  `{os, arch, cpu_model, logical_cores, go_version, clang_version}`, reusing
  the existing tool-identity-probe pattern. **No hostnames, serials, or MAC
  addresses** — host fingerprints are a needless leak. This also directly
  serves PROJECT.md's constraint that target facts must not accidentally encode
  the current Apple arm64 host.

- **D-06-18:** On an **undeclared** machine, `verify` runs in
  **observation-only mode**: it reports the full distribution and the computed
  `machine_id`, marks `ratified: false`, and **neither consults nor writes the
  budget manifest**. Fail-closed on *ratification*, not on *running* — any
  machine can still produce evidence.

- **D-06-19:** Statistical discipline reuses Phase 2's shipped mechanism: 20
  warm samples, a small fixed n for cold (cold state cannot be re-established
  in-process), reporting **p50 and p95**. Noise margin is the coefficient of
  variation, and **any metric whose CoV exceeds a fixed threshold across its 20
  warm samples is automatically demoted to `observed` for that run** — this
  mechanizes "quarantine unstable metrics as observations" instead of leaving
  it to per-run human judgment.

- **D-06-20:** **Peak RSS stays `"unavailable"` for M001.** Do not implement
  `getrusage`. `ru_maxrss` is bytes on macOS and kilobytes on Linux, and a Go
  process's RSS is dominated by runtime/GC allocation unrelated to compiler
  work; implementing it on the only available host would trade an honest gap
  for exactly the host-encoding trap PROJECT.md warns against. **Record the
  reason in a code comment at the `peak_rss_status` sites** so the gap reads as
  deliberate rather than neglected. Revisit when a second (Linux) machine
  exists to validate the unit conversion.

- **D-06-21:** **Stage attribution without a tracing runtime.** Introduce
  explicit start/stop timestamp pairs at the small number of fixed, sequential
  stage boundaries the pipeline already crosses — parse, check/validate,
  lower/emit, native-compile, link — recorded into the existing session/lane
  bookkeeping structs. No global tracer, no span model; PROJECT.md forbids a
  mandatory global tracing runtime.

- **D-06-22:** **Blocking rule.** A regression is blocking **if and only if**
  `recomputed_work` for the affected lane or stage exceeds its ratified
  ceiling — exact, deterministic, no noise judgment. A wall-clock or
  output-bytes observation exceeding its loose bound is **never blocking on its
  own**; it surfaces as a flagged observation that must cite the responsible
  stage's work-count delta before it can become a gate candidate in a future
  ratification. This mirrors rustc-perf's practice of confirming a wall-time
  signal against instruction counts before acting on it.

### DX-04 repair exercise

- **D-06-23:** **Two-part answer: a deterministic CI gate plus a separately
  recorded, non-gating agent exercise.** The gate proves the *protocol* is
  mechanically sufficient; the recorded exercise proves the *diagnostics are
  legible to a consumer who did not author them*. Rejected alternative,
  recorded: a deterministic repairer alone. Rejected because a repairer written
  against the same fixtures it repairs, by the same author, with no held-out
  set, is by definition not evidence that the protocol is agent-legible — it is
  the overfitting failure named in the automated-program-repair literature
  (Qi et al. 2015; Smith et al. 2015) and the exact shape of this project's own
  three-gate-failure lesson: *a green test whose reachable input space omitted
  the hard case.* A live LLM in the test suite was rejected in the same breath:
  it would violate the offline / stdlib-only / deterministic constraints and
  flake the gate.

- **D-06-24:** `diagnostic.Repair` must **grow from a classification into an
  applicable edit**. Today it is `{kind, detail}` — no span, no replacement
  text. Minimal additive extension on `lang.diagnostic/1`: optional `span`,
  optional `replacement`, and a rustfix-style
  `applicability: MachineApplicable | RequiresConfirmation | Unspecified`
  (an enum, not a boolean — only `MachineApplicable` is driver-eligible;
  everything else routes to the recorded non-gating exercise).
  **Critically: `span`, `replacement`, and `applicability` join `detail` in the
  NON-identity-bearing bucket. Only `kind` participates in semantic identity**
  — otherwise a coordinate shift would spuriously change a diagnostic's ID.
  `lang.diagnostic/0` bytes remain frozen. — **Reversibility:** one-way — the
  identity-bearing/non-identity-bearing split is a published contract; moving a
  field across that line later changes every previously-published diagnostic ID.

- **D-06-25:** Five defect injectors, one per named class:
  - **match** — new small held-out `.lang` defect corpus at source granularity
    (remove an arm / duplicate a pattern). Exhaustiveness lives above the C
    layer the existing runners mutate.
  - **move** — held-out `.lang` fixtures injecting affine use-after-move.
  - **borrow** — held-out `.lang` fixtures injecting a loan conflict
    (overlapping exclusive + shared borrow).
  - **cleanup** — **reuse the existing release-omission mutation runner
    directly**; it already targets `lang:release-site` markers, which is
    exactly a missing-`OpRelease` defect.
  - **stale-evidence** — not a source defect at all: re-touch source (or the
    core artifact) *after* a manifest was captured so its SHA-256 no longer
    binds. Located via `lang evidence --validate`'s existing mismatch report,
    not via a source diff; repaired by recapturing a manifest that binds.

- **D-06-26:** Success oracle must **reject the degenerate "delete the
  offending code" repair**. For match/move/borrow/cleanup: `lang check` clean
  **AND** the result byte-identical to the pre-defect original wherever the
  defect was a single mechanical mutation (proving the repair *reversed* the
  injected defect rather than producing a different plausible program) —
  falling back to a differential-behaviour check through the existing
  reducer/mismatch machinery only for classes where byte-identity is too
  strict. For stale-evidence: `--validate` passes **AND** the manifest's bound
  digest matches current `lang format`-canonical source, not merely "some
  manifest exists."

- **D-06-27:** **Three anti-theater guards — these are the real deliverable of
  DX-04, not the repairer itself:**
  1. **Prose-scramble test.** CI runs the driver twice: once against real
     diagnostics, once with every `message` and `detail` string replaced by
     lorem ipsum (`kind` / `span` / `replacement` untouched), asserting
     **identical repair behaviour and outcome**. This is "without scraping
     prose," operationalized as an executable claim.
  2. **Structured-vocabulary-removal test.** A third run strips `repairs[]` /
     `kind` / `span` / `replacement` from the JSON with prose left intact, and
     asserts the driver **goes RED**. This proves the mechanism genuinely
     depends on the structured channel rather than silently falling back to
     prose it could scrape.
  3. **Marker mutation-kill.** Each injector's target marker/span is itself
     mutation-tested: if the target disappears, the exercise **refuses** rather
     than silently passing — the same fail-closed marker-count guard Phase 5
     shipped for exactly this failure mode.

- **D-06-28:** **Protocol-only enforced structurally, not by convention.** The
  repair driver lives in `cmd/`, invokes the **shipped `lang` binary as a
  subprocess** and never imports `internal/compiler/*`, and reads only `--json`
  output. A CI import-boundary lint **fails the build** if the driver ever
  imports anything under `internal/` or opens a `.lang` source file outside the
  JSON `span` / `replacement` it was handed by the binary.

- **D-06-29:** Held-out corpus discipline: the defect-injection fixtures used
  by the CI gate MUST be distinct from any fixtures used to hand-derive the
  driver's kind→edit mapping. This blunts — it cannot eliminate, given how
  small the language is — the overfitting concern, and the residual is honest
  debt to record.

- **D-06-30:** Budget and cadence. The CI gate is **single-pass per defect**
  (inject → diagnose → apply → reverify), bounded and O(1) per class, with no
  repair loop. The **recorded exercise** is the one that measures
  `repair_rounds` against the `agent.repair_rounds.p95: 2` intent from
  `wiki/compute-efficiency-constitution.md`, plus token and tool-call cost, and
  whether protocol-only access sufficed. It is **evidence, not a gate**, runs
  on a slower (per-milestone / on-demand) cadence, and reports honestly
  unmeasured values rather than fabricating them.

### Cross-cutting: schema versioning

- **D-06-31:** One coordinated additive bump, decided across areas rather than
  per-area. `explain` and `query` are net-new documents at `lang.explain/0` and
  `lang.query/0` and need **no** bump to `lang.command/0` on their own account.
  However, D-06-12 (`cache_status`, `selection_reason`,
  `cache_inputs_reused_count`) and D-06-15/17/21 (`machine_id`, `gate_verdict`,
  `cold_or_warm`, `stage_breakdown`) both add fields to the **existing**
  `protocol.Lane` and `protocol.Metrics` structs. Under the standing additive
  discipline that is **one** coordinated bump: **`lang.verify-lane/0` → `/1`
  and `lang.command/0` → `/1`, with all `/0` bytes frozen byte-for-byte.**
  Phase 1 evidence bytes remain byte-identical. — **Reversibility:** one-way —
  published schema versions cannot be unpublished.

- **D-06-32:** **No new metrics field may enter `Result.Finalize()`'s identity
  struct.** FND-04's "without changing semantic output" is a hard constraint;
  timing values correctly stay out of identity today and must continue to.
  This is a required verification item, not an assumption.

### Claude's Discretion

- Exact node-count budget and default `--depth` for `explain`; exact CoV
  threshold for the auto-demotion rule in D-06-19; exact truncation code
  strings — all should follow the closest existing in-tree precedent rather
  than being invented.
- Whether the `risk_lanes.json` audit is a new executable audit or an extension
  of the existing QLT-01 audit harness.
- File/package placement within `internal/compiler/` for the new cache and
  budget code, following the established sibling-file convention (Phase 5 kept
  every new surface in `session_phase5*.go` siblings rather than editing
  `session.go`).

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase 6 roadmap refs (from ROADMAP.md)
- `wiki/compiler-and-feedback-latency.md` — candidate latency service levels
  table; the `verify --changed` affected-evidence model with its
  `checked[] / deferred[{kind,reason}] / invalidated_by[] / sound_for[]` JSON
  sketch; the warning that undeclared foreign inputs (FFI headers, link
  libraries, environment, tools) escape a cache key; §"Stable, fine-grained
  queries" on source vs semantic fingerprints; §"Self-observation".
- `wiki/compute-efficiency-constitution.md` — §"Budget manifest" (the
  `budget development { … }` / `budget ci_pull_request { … }` sketch, explicitly
  labelled hypotheses); the prevent/explain/gate triad; §"Measurement and
  regression protocol" (10 rules, incl. "ratchet only stable metrics with a
  noise margin", "quarantine unstable metrics as observations"); §"AI
  development-loop efficiency" (stable IDs, compact typed facts, deltas,
  smallest sufficient context, `agent.repair_rounds.p95: 2`); §"CI and
  verification lanes" (the keystroke/save/commit/PR/nightly/release ladder).
- `wiki/performance-observability-and-delivery.md` — shift-left performance,
  performance evidence, monitoring/SLOs, compiler and runtime self-observation.
- `wiki/ai-native-runtime-and-evals.md` — typed tools and authority, budget
  algebra, replay honesty, change impact for AI artifacts.

### Project-level authority
- `.planning/PROJECT.md` — constraints (feedback latency, correctness, runtime
  posture, bootstrap, dependencies, portability, evidence, security) and the
  Key Decisions table, incl. "Treat structured diagnostics and evidence as a
  versioned product API".
- `.planning/REQUIREMENTS.md` — FND-04, DX-02, DX-03, DX-04, QLT-02 verbatim.
- `.planning/ROADMAP.md` §Phase 6 and §"M002 Charter" — the `OpCall` deferral
  that fixes this phase's boundary.
- `.planning/STATE.md` §"Accumulated Context" — the full decision ledger for
  Phases 1–5 and the open blocker "Baseline machines for ratified feedback
  budgets remain to be chosen before Phase 6" (closed by D-06-17/D-06-18).

### Prior-phase context and debt
- `.planning/phases/05-native-equivalence-and-adversarial-evidence/05-CONTEXT.md`
- `.planning/phases/05-native-equivalence-and-adversarial-evidence/05-DEBT.md`
  — carry-forward entries, incl. D-03-02 open past the milestone and D-05-41.
- `.planning/phases/04-fallible-resources-and-c-boundary/04-CONTEXT.md`
- `.planning/phases/03-borrowed-views-and-cfg-lifetimes/03-CONTEXT.md`
- `.planning/phases/03-borrowed-views-and-cfg-lifetimes/03-DEBT.md`
- `.planning/spikes/MANIFEST.md`

### Ecosystem prior art cited during discussion (external, for the researcher)
- Kythe schema / storage model — the "one canonical identity used as the join
  key" pattern behind D-06-01.
- SARIF 2.1.0 `graphs` / `graphTraversals` — studied as the cautionary case for
  D-06-02 (expressive, near-universally unimplemented).
- Go build cache `ActionID` (`cmd/go/internal/cache/cache.go`) — the declared-
  input content-key model behind D-06-07.
- Ekstazi / STARTS regression-test-selection soundness literature — behind
  D-06-13's explicit hole list.
- rustc-perf `instructions:u` rationale and collector README; SPEC CPU2017 run
  and reporting rules — behind D-06-14/D-06-15/D-06-22.
- rustc `Applicability` enum and `rustfix` — behind D-06-24.
- Qi et al. 2015 and Smith et al. 2015 (APR patch-overfitting) — behind D-06-23
  and D-06-29.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- **`internal/compiler/protocol`** — `Result`, `Metrics`, `Lane`,
  `EvidenceSummary`, `InterfaceSummary`, `DebugMapSummary`, plus `Finalize()`'s
  content-derived identity and the `JSON`/`Human` dual projection. Every new
  command plugs in here. `Metrics` already carries `elapsed_ns`,
  `output_bytes`, `recomputed_work`; `peak_rss_status` is hardcoded
  `"unavailable"` at ~15 producer sites; **there is no `cache_status` field and
  no cache code anywhere in the tree.**
- **`internal/compiler/debugmap` + `lang debug-map`** — the direct precedent
  and probable substrate for D-06-01's join: it already joins source span →
  `core_id` → `operation_id`/`point_id`, already accepts a single-ID query
  argument, and already reports honest `availability: not_captured |
  optimized_out`.
- **`internal/compiler/diagnostic`** — `Diagnostic{schema, id, code, severity,
  primary_span, message, causes[], repairs[]}`; `Cause{kind, detail, span}`;
  `Repair{kind, detail}`; `Error()` / `ErrorWithRepairs()`; the established
  rule that repair *kinds* participate in identity while prose does not. This
  is what D-06-02 reads and D-06-24 extends.
- **`session/qlt01_registry.json` + `session/qlt01.go` + its executable audit**
  — the checked-in-JSON-registry-with-live-cross-check precedent that both
  `risk_lanes.json` (D-06-10) and `qlt02_budget_manifest.json` (D-06-16) copy.
- **Phase 4/5 mutation runners** — the release-omission runner targeting the
  LAST `lang:release-site` marker is reused verbatim as D-06-25's cleanup
  injector. The fail-closed marker-count guard is the model for D-06-27's third
  guard.
- **`internal/compiler/reduce`** — five moves, strict predicate, bounded budget,
  `lang.mismatch/0`, and `session.SignatureFromDisagreement`. Available as the
  fallback differential oracle in D-06-26.
- **Tool-identity probes** — already bounded (64 KiB-plus-one) with five-second
  deadlines, already used for sanitizer availability. Directly reusable for
  D-06-07's Clang digest and D-06-17's `machine_id`.
- **Phase 2's bounded gate** — already reports five 20-sample warm
  distributions *without* ratifying an SLO. D-06-19 reuses this machinery; the
  ratification is the only new part.
- **`lang evidence` / `evidence --validate`** — the stale-evidence injector,
  locator, and repair path in D-06-25 are all already-shipped surface.

### Established Patterns

- **Additive `/1` bumps with `/0` bytes frozen byte-for-byte.** Precedents:
  `lang.diagnostic/0`→`/1`, `lang.evidence/0`→`/1`. Phase 1 evidence bytes are
  frozen and must remain byte-identical.
- **Independent re-derivation at trust crossings (D-12).** Checker and
  independent validator never share a mechanism; they re-derive the same fact
  by materially different algorithms. Anything Phase 6 asserts twice must
  follow this.
- **Sibling-file convention.** Phase 5 put every new surface in
  `session_phase5*.go` siblings and left `session.go` untouched; the Phase 5
  corpus dispatch even lives at the CLI layer (`isPhase5Corpus` in
  `cmd/lang/main.go`) specifically to avoid editing `session.VerifyCorpus`.
- **Honest unavailability over fabrication.** `availability: not_captured` /
  `optimized_out`, `peak_rss_status: "unavailable"`, and the coordinated
  source-to-core false claim documented as an **expected escape** rather than a
  claimed control.
- **Fail-closed.** Phase 5 shipped fail-closed field routing and a marker-count
  guard that refuses when its mutation target disappears.
- **Bounded everything with stable truncation codes**, and explicit counted
  `recomputed_work` pinned to exact constants in tests.
- **Standing process rules adopted after three gate failures that shared one
  shape — a green test whose reachable input space omitted the hard case:**
  (1) mutation-kill every differential, (2) interrogate what inputs a property
  test actually reaches, (3) **drive the shipped binary on hand-written
  programs**, not only the gate's own corpus. D-06-27 and D-06-28 exist because
  of this rule and must be planned as first-class work, not as test polish.

### Integration Points

- `cmd/lang/main.go` — flat arg-matching dispatcher (no `flag` package, no
  cobra). Two new command arms for `explain` and `query`; `--json` extraction
  already exists and is shared.
- `protocol.Lane` / `protocol.Metrics` — the single coordinated `/1` bump
  (D-06-31), plus the hard constraint that none of it reaches
  `Result.Finalize()`'s identity struct (D-06-32).
- `session.VerifyCorpusFile` / `session.VerifyPhase5ControlsAndWork` — where
  lane selection and cache-status reporting land. Both currently recompute
  everything on every invocation.
- `internal/compiler/cache` — new package (D-06-06 … D-06-09).
- `cmd/<repair-driver>` — new binary, subprocess-only, import-boundary-linted
  (D-06-28).
- `scripts/verify-phase6.sh` — the per-phase release cost lane, following
  `verify-phase1.sh` … `verify-phase5.sh`.

</code_context>

<specifics>
## Specific Ideas

- The developer's explicit instruction for this discussion was a deliberate
  breadth-then-depth fan-out: consider every relevant stakeholder-role lens
  (security, product, architecture, DevOps, domain specialists), weigh
  pros/cons/trade-offs with real examples, name anti-patterns, best practices,
  footguns and lessons learned, research other products and ecosystems online,
  run an adversarial pass, then synthesize one decisive recommendation per
  decision point. All four areas were researched under that instruction and
  every recommendation was accepted. **Downstream agents should treat these
  decisions as researched and locked, not as defaults to re-open.**
- The load-bearing framing the developer selected across three of four areas is
  the same one: **prefer the deterministic, exactly-assertable mechanism as the
  gate, and report everything noisy or unprovable honestly alongside it rather
  than gating on it.** Cache artifacts not verdicts; gate work-counts not wall
  time; gate the repair mechanism but record agent legibility. Planning should
  preserve that shape.
- `explain`'s cause graph is synthesized per cold invocation. Its determinism
  is a testable obligation and should get an explicit test, not an assumption.
- The prose-scramble test (D-06-27.1) and the vocabulary-removal test
  (D-06-27.2) are the falsifiability core of DX-04. If planning has to cut
  scope, these are the last things to cut — without them the repair exercise is
  theater by construction.

</specifics>

<deferred>
## Deferred Ideas

- **Peak-RSS measurement via `getrusage`** — deferred out of M001 (D-06-20).
  Revisit when a second, Linux, machine exists to validate the bytes-vs-
  kilobytes unit conversion and to separate Go-runtime RSS from compiler work.
- **A persistent compiler service / query daemon** — the long-term
  architecture sketched in `wiki/compiler-and-feedback-latency.md`. Explicitly
  a post-M001 decision; PROJECT.md lists "a permanent compiler host" as
  out of scope for this milestone.
- **Remote cache, LSP, MCP server, package registry** — named out of scope in
  PROJECT.md; the Phase 6 cache is local-only by decision (D-06-09).
- **A real `budget { … }` DSL and its parser** — the wiki sketch's own text
  calls its numbers hypotheses; M001 ships the JSON manifest instead (D-06-16).
  A DSL becomes worth building when budgets are stable enough to be worth a
  grammar.
- **Statistical wall-clock gating** with CoV/bootstrap thresholds and a
  quarantine workflow — deferred until a dedicated, thermally-stable,
  multi-machine runner fleet exists (D-06-15).
- **`OpCall` and interprocedural loan liveness / `-O3` equivalence** — already
  M002's lead charter item (D-05-32/D-05-33). Untouched by this phase.
- **Widening the LTO lane beyond one fixture** — existing recorded debt D-05-41,
  landing "Phase 5+"; not pulled into Phase 6 scope.

</deferred>

---

*Phase: 6-agent-feedback-and-performance-ratification*
*Context gathered: 2026-09-06*
