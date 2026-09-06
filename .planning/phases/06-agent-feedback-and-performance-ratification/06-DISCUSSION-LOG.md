# Phase 6: Agent Feedback and Performance Ratification - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-09-06
**Phase:** 06-agent-feedback-and-performance-ratification
**Areas discussed:** query/explain addressing, Change-risk + cache substrate, Budget declaration + machines, Repair-exercise mechanism
**Mode:** advisor (USER-PROFILE.md present; `vendor_philosophy: opinionated` → calibration tier `minimal_decisive`; `NON_TECHNICAL_OWNER = false`, overridden by explicit `technical_background: true` and `explanation_depth: practical-detailed:technical`)

---

## Gray area selection

All four presented areas were selected. The developer added a standing
instruction for the research fan-out, quoted here as data:

> "for each fan out for each decision point consider breadth+depth through all
> relevant stakeholder-role lenses (security, product, technical, design, etc
> roles...) and consider pros/cons/tradeoffs/examples,
> antipatterns/patterns/bestpractices/footguns/lessons learned even research
> online and get insight from other products/ecosystems as well if helpful....
> goal is to fan out consider all the angles adversarial pass as well then
> synthesize perfect one-shot recommendations for each"

Four `gsd-advisor-researcher` agents were spawned in parallel (model: sonnet),
one per area, each carrying that instruction plus scouted codebase facts and
the project's established decisions.

---

## query/explain addressing surface

| Option | Description | Selected |
|--------|-------------|----------|
| Joined query + synthesized DAG | One `query` resolving any existing ID form; `explain` synthesizes a bounded, depth-capped cause DAG from the flat `causes[]`. Mint `lang.explain/0` and `lang.query/0` as net-new documents. Kythe-style join key, no persisted graph store. | ✓ |
| Per-kind subcommands | `query symbol` / `query type` / `query ownership` / `query deps` / `query tests`, each a flat projection; `explain` mechanically expands `causes[]` with no edges. Cheapest to ship, defers the cause-graph criterion, adds two more ad-hoc vocabularies. | |

**User's choice:** Joined query + synthesized DAG
**Notes:** Prior art weighed — Kythe VName (one canonical identity as universal
join key), SARIF 2.1 `graphs`/`graphTraversals` (studied as the cautionary
case: expressive, near-universally unimplemented because the traversal model is
heavy to author and consume), LSP `relatedInformation` (deliberately flat, no
graph), rustc `--explain` (flat code → static prose, no query language), Bazel
`query`/`cquery` (one query language over one label-addressed graph, confirming
the single-addressing-scheme pattern scales). Decisive argument: the repo
already carries five ID vocabularies and already has an honest-join precedent
in `debug-map`; minting a sixth would be the wrong move, and M001's call-free,
single-parameter language makes cause chains shallow enough that a synthesized
DAG is both honest and cheap.

---

## Change-risk + cache substrate

| Option | Description | Selected |
|--------|-------------|----------|
| Artifact cache + static risk table | Content-addressed local cache of compiled/instrumented binaries ONLY — never verdicts; checker/comparator/sanitizer always re-run. Static `risk_lanes.json` selection table, gitignored local state file for "changed", conservative widening, deferred lanes never render as `pass`. | ✓ |
| Selection only, cache deferred | Ship lane selection now; `cache_status` honestly reports `not_implemented` following the `peak_rss_status` precedent. Zero new soundness surface, but the expensive `-O3`/`-flto`, sanitizer, and per-mutant recompiles stay uncached and SC2's "reports reused cache inputs" is satisfied only in absence. | |

**User's choice:** Artifact cache + static risk table
**Notes:** Prior art weighed — Go build cache `ActionID`, Nix/Guix
input-addressed vs content-addressed derivations, ccache/sccache footgun class
(`__TIME__`, headers, compiler upgrades), rustc incremental red-green graph and
its soundness-bug history, Gradle/Bazel declared-input discipline, Nx `affected`
selection, and the Ekstazi/STARTS regression-test-selection soundness
literature. Decisive argument: caching *artifacts* rather than *verdicts* is
the structural property that removes the false-green vector entirely, while
still capturing where the real seconds live. A rustc query graph or Bazel
action graph was rejected as over-engineering given M001 has no call graph and
no modules. The known unsound holes were required to be documented rather than
papered over, consistent with the project's existing expected-escape practice.

---

## Budget declaration + machines

| Option | Description | Selected |
|--------|-------------|----------|
| Gate work, observe wall time | `recomputed_work` is the only hard gate (exact, machine-independent, already pinned). Wall-clock cold/warm p50/p95 ratified as loose bounded observations per `machine_id`; CoV auto-demotes noisy metrics. `peak_rss` stays honestly "unavailable" for M001. JSON budget manifest beside `qlt01_registry.json`. | ✓ |
| Statistical wall-clock gate | Hard p95 cold/warm thresholds per declared host with a real statistical pipeline (CoV/bootstrap, quarantine state, exception review). Answers SC4's wording literally, but the project has one laptop-class Apple-silicon host and no prior distribution to set a credible threshold from. | |

**User's choice:** Gate work, observe wall time
**Notes:** Prior art weighed — rustc-perf's `instructions:u`-over-wall-time
rationale (the single most on-point precedent, since `recomputed_work` is its
direct analogue), SPEC CPU2017 reportable-vs-informational run rules, LLVM
benchmarking guidance and LNT, Go `benchstat`, Firefox Talos noisy-test
quarantine, Chromium perf bots, Clang `-ftime-trace` and rustc `-Ztime-passes`
for stage attribution. Decisive argument: the project already owns an exact,
deterministic, machine-independent cost metric pinned to constants across the
tree, and a single laptop-class Apple-silicon host with battery/thermal/P-E-core
confounds and no prior distribution cannot support a credible first p95 gate —
gating it would produce exactly the flaky thresholds
`wiki/compute-efficiency-constitution.md` forbids. Peak-RSS was examined
directly and deferred: `ru_maxrss` unit divergence (bytes on macOS, kilobytes on
Linux) plus the Go-runtime RSS confound make it a portability trap on the only
available host, so honest "unavailable" wins for M001. This closes the
long-standing STATE.md blocker "Baseline machines for ratified feedback budgets
remain to be chosen before Phase 6."

---

## Repair-exercise mechanism

| Option | Description | Selected |
|--------|-------------|----------|
| Split: gate + recorded exercise | Deterministic driver gates CI offline (five injectors, byte-identity oracle, prose-scramble + vocabulary-removal + marker mutation-kill, `cmd/`-only import boundary). Agent-legibility claim — repair rounds vs the `p95:2` budget, tokens, whether protocol-only sufficed — recorded as non-gating evidence on a per-milestone cadence. | ✓ |
| Deterministic gate only | Ship the CI-gated driver and its anti-theater guards; skip the recorded agent exercise entirely. Accepts DX-04 as a mechanism-only claim and defers agent-legibility to a later milestone, with SC3's "an agent can locate and repair" left unfalsified. | |

**User's choice:** Split: gate + recorded exercise
**Notes:** Prior art weighed — rustc `Applicability` enum and `rustfix`
suggestion-span design (the most on-point precedent for turning a diagnostic
into a machine-applicable edit), Clang `FixItHint` and `clang-apply-replacements`
overlapping-replacement footguns, ESLint `--fix` multi-pass fixer conflicts, Go
analysis `SuggestedFix`/`TextEdit`, SARIF `fixes[]`/`artifactChanges`, LSP
`CodeAction`/`WorkspaceEdit`, and the automated-program-repair overfitting
literature (Qi et al. 2015; Smith et al. 2015 *Is the cure worse than the
disease?*; the 2025 follow-up), plus Defects4J/QuixBugs/SWE-bench validity
critiques and mutation-testing equivalent-mutant theory. Decisive argument: a
deterministic repairer alone reproduces this project's own three-gate-failure
shape — a green test whose reachable input space omitted the hard case — while
a live LLM in the suite would break the offline/deterministic/stdlib-only
constraints. Splitting gate from evidence resolves both, and mirrors the
project's existing template of fail-closed mutation-killed gates paired with
honestly-labelled non-gating evidence. The prose-scramble and
vocabulary-removal tests were identified as the falsifiability core of DX-04 —
without them the exercise is theater by construction.

---

## Cross-area synthesis raised during discussion

A schema-versioning conflict was surfaced across areas and resolved as a single
coordinated decision rather than three independent ones: `explain`/`query` are
net-new documents needing no bump, but the cache fields and the budget/machine
fields both extend the existing `protocol.Lane` and `protocol.Metrics` structs,
which under the additive discipline is one coordinated bump —
`lang.verify-lane/0` → `/1` and `lang.command/0` → `/1`, `/0` bytes frozen —
with the hard constraint that none of the new fields may enter
`Result.Finalize()`'s identity struct (FND-04's "without changing semantic
output"). Recorded as D-06-31 and D-06-32.

## Claude's Discretion

- Exact node-count budget and default `--depth` for `explain`.
- Exact CoV threshold for the noisy-metric auto-demotion rule.
- Exact truncation code strings (should follow closest in-tree precedent).
- Whether the `risk_lanes.json` audit is a new executable audit or an extension
  of the existing QLT-01 audit harness.
- File/package placement within `internal/compiler/` for the new cache and
  budget code, following the Phase 5 sibling-file convention.

## Deferred Ideas

- Peak-RSS via `getrusage` — revisit when a Linux machine exists to validate
  the unit conversion.
- Persistent compiler service / query daemon — post-M001.
- Remote cache, LSP, MCP server, package registry — out of scope per PROJECT.md.
- A real `budget { … }` DSL and its parser — JSON manifest ships instead.
- Statistical wall-clock gating — deferred until a dedicated runner fleet exists.
- `OpCall` and interprocedural loan liveness / `-O3` equivalence — already
  M002's lead charter item (D-05-32/D-05-33).
- Widening the LTO lane beyond one fixture — existing recorded debt D-05-41.
