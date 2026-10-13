# Schway

<!-- schway-current:start -->
Current publication identity (2026-09-28): Schway uses the Go module `github.com/szTheory/schway`, the `schway` and `schway-repair` commands, `.schway` source files, `schway.*` and `schway:*` protocol identifiers, and `schway_` and `SCHWAY_` native ABI symbols. Preserve older spellings only where they document historical implementation evidence.
<!-- schway-current:end -->


## What This Is

Schway is a general-purpose programming language and toolchain for
AI-authored, human-audited production software. It aims to combine fast,
structured feedback with a small coherent semantic core, usable low-level
ownership and resource control, deterministic evidence, and a calm canonical
source form that remains readable when humans mostly review rather than type it.

## Name and Migration Decision

**Schway** is the chosen public name for both the language and project, decided
2026-09-28. Treat this decision as settled; do not reopen naming or generate new
shortlists unless later clearance finds a material direct conflict. An
exact-name web search on the decision date did not surface a programming
language named Schway; it did surface an unrelated content-marketing service at
[schway.com](https://schway.com/), which is an accepted tradeoff for this
personal open-source project.

On 2026-09-28, `github.com/szTheory/schway` was recorded as the canonical public repository path. The repository was confirmed public and empty before the initial source publication. The isolated publication copy now uses the Schway module path, commands, `.schway` source suffix, protocol identifiers, and native ABI names.

The history rewrite and privacy audit passed before publication. The sanitized `main` preserves all 1,538 original commits, their ordered parent lists and 18 merge commits, plus the three milestone tags. The current tree and every selected ref were scanned for personal paths, contact tokens, owner-name tokens, phone-shaped strings, credentials, and secrets. Confirmed and unclassified candidates were both zero. The initial push is limited to sanitized `main` and the three milestone tags; the hosted Ubuntu evidence receipt is pending that push.

The first three milestones were deliberately narrower than the full vision:
prove one real source-to-native semantic spine (M001), prove it survives a
function boundary (M002), then prove it survives computation (M003), before
expanding into effects, services, agents, databases, GUIs, or an ecosystem.
Calls between Schway functions, computed-value matches, and U64 constants run
end to end. Phase 26 now adds checked U64 addition, comparison/Bool branching,
and scalar loops through an ordinary `sum_to_n` application; general strings,
collections, modules, and resource-carrying loops remain future work.

The development destination is a usable general-purpose language: small native
programs, FizzBuzz and file utilities, reusable byte-oriented libraries, JSON,
HTTP, and low-level systems libraries. These are capability goals, with new
surface admitted through real programs and explicit safety evidence. See
[PRODUCT-ROADMAP.md](PRODUCT-ROADMAP.md) for the living delivery order and
[LANGUAGE-MATURITY.md](LANGUAGE-MATURITY.md) for the observed frontier.

## Core Value

Give an AI agent and a human reviewer the shortest reliable path from intent to
sound, reproducible evidence without wasting iteration time or hiding runtime
costs.

## Verification Operating Preference

Turn acceptance criteria into early deterministic checks at the lowest layer that proves each claim: unit, seam, smoke, integration, or end-to-end checks, with relevant failure controls. Put a check in recurring CI when its regression value justifies its runtime and maintenance cost. Aim for zero human UAT when automated evidence covers the criteria; hand off only genuinely subjective, external, or user-authority properties that lack repeatable proxies. Prefer standard-library and local code, with no new dependency by default.

This preference does not waive mandatory workflow gates, required user authorization, or acceptance decisions explicitly designated as human-only. Future phase plans must name the concrete verification commands that prove their acceptance criteria and arrange for high-value recurring checks to run in existing CI.

During `/gsd-verify-work`, the agent runs every derived checkpoint itself before
presenting any (CLI command, exit code, stdout/stderr, docs claim), records each
as `result: pass` with `source: automated` and its covering test, and adds a
focused test (with a mutation/armed control) for any checkpoint not already
pinned in `go test ./...` CI. Only checkpoints with no repeatable proxy reach
the user; when none remain, the session completes with no human prompt.
Settled 2026-10-11 (Phase 29 UAT); do not re-ask.

From Phase 32 on, every plan SUMMARY carries a frontmatter `coverage:` block,
as in the GSD summary template. Each deliverable is either
`human_judgment: false` with passing `ref: <file>#<TestName>` rows, or
`human_judgment: true` with a `rationale` naming why no repeatable proxy
exists. `/gsd-verify-work` then auto-passes deliverables deterministically
instead of re-deriving checkpoints from prose (Phase 31's SUMMARYs lacked the
block, so its UAT was derived by hand). `TestSummaryCoverageGroundedness`
(session meta set) enforces the block and resolves every named test. Settled
2026-10-12 (Phase 31 UAT); do not re-ask.

When an automated verification report becomes stale, preserve completed UAT and
refresh the report from current source and automated evidence instead of
replaying user checkpoints. If the phase is already marked complete and
`$gsd-execute-phase` therefore no-ops, dispatch the GSD verifier directly; then
re-query `init.progress` and update the handoff. Keep hosted CI receipts distinct
from local or container runs, and reserve human UAT for criteria that remain
subjective, external, or user-authority decisions after automation.

When a verifier reports only a required hosted CI receipt that cannot be run
because this checkout has no remote or runner access, treat it as an external
blocker: do not create an implementation gap plan or human UAT. Record the
existing CI job as the next action, keep the phase incomplete, and resume its
verifier after the receipt arrives without replaying completed plans.

## Current Milestone: M006 Bounded Byte Utility and Local Modules

**Goal:** A developer can run an ordinary Schway `checksum` program that reads a
size-limited file and prints one specified checksum with exact success and
error results, built from a reusable local module that a second consumer
imports without copying.

**Target features:**
- Admit only the byte operations the checksum consumer needs: byte views,
  lengths, bounds-checked indexing, and borrowing a buffer across a scalar loop.
  Every operation gets independent interpreter and native evidence, plus
  out-of-bounds and wrong-result controls.
- Specify the checksum program's results exactly for empty, one-byte, maximum
  and oversized input, a missing or unreadable file, and the error-path cleanup
  of the owned buffer.
- Add a local module boundary with explicit visibility, where a second consumer
  imports the checksum library without copying it.
- Carry over M005 debt:
  - N2: re-acquire two-host hosted receipts at the current revision.
  - N1: keep the named capacity reason when the app verify report would exceed
    its size limit.
  - W2/W3: split the conflated exit classes in `build` and `run --engine=native`.

**Decision method:** design decisions are researched with multiple lenses,
including an adversarial pass. Each gets one synthesized recommendation, which
is auto-followed. Copy-local code is preferred over new dependencies.

## Current State

**2026-10-10 M005 Practical Computation shipped** (5 phases, 25 plans; archive
`.planning/milestones/M005-*`, tag `vM005`). Ordinary Schway source now sums
bounded U64 input (`examples/sum_to_n.schway`: 0/55/500500, 1,001 refused, checked
overflow exits 65) and prints exact FizzBuzz through 1,000
(`examples/fizzbuzz.schway`, byte-identical to pinned fixtures) with U64
equality/remainder, static zero-divisor refusal, dynamic checked remainder failure,
and a shared 65,536-byte output ceiling. Both programs have independently pinned
interpreter/native answers and dual-host receipts (gate at 6224f8b). Inserted
phases 27.1–27.3 closed audit debt: origin validation on the build route,
straight-line scalar `let` admission (R1), a named
`native.event_capacity_exhausted` outcome so `app verify` reports `unsupported`
instead of a signal (W1), and guide outputs under ignored `.schway/` (W4).
Known limits: `app verify` on `sum_to_n` verifies only n ≤ 14 and loses the named
reason above n ≈ 172 (report size limit, N1); hosted evidence (run 38091738463 at
d7ffbcd) predates the final review fixes (N2). M006 started 2026-10-10 (see
Current Milestone above).

### Historical state records (superseded by dated closeouts)

**M004 opened 2026-09-27.** Phase 21 is completed, archived contract/retirement
prework; its six plans and seven UAT cases must not be replayed. Phase 22 is
complete: `lang app build` retains an artifact from the admitted source and
declared local build inputs, while `lang app run` accepts bounded U64 input,
starts the selected app once, and preserves ordinary output and process
outcomes. Evidence capture and explicit replay use separate routes. The older
`lang run` remains a synthetic-input conformance harness with optimization-tier
replay. Phase 22 linked local C but did not admit Schway foreign calls or live
foreign ownership; all three foreign/by-pointer families remain refused. Its
macOS receipt reports incomplete dependency closure and `cacheable: false`.
The refreshed 2026-09-28 verifier confirms 5/5 roadmap truths and the completed
objective README contract UAT remains unchanged. Subjective readability is not
claimed.

Phase 23 now admits a bounded file-byte allocation that stays live through
Schway-directed use and generated local cleanup. Its focused script passes on
macOS and in a local Linux ARM64 container; the full Go suite passes on macOS.
Security review covers 18/18 declared threats, and code review is clean. The
configured hosted Ubuntu CI receipt remains pending, and Phase 23 stays in
progress with its required CI evidence outstanding. No Git remote is configured
in this checkout, so the hosted job could not be run here. Phase 24 owns
transfer and typed-error cleanup; Phase 25 owns the bounded shared/exclusive
pointer families.
The following shipped records describe evidence at their recorded revisions.

**Shipped M002 — Interprocedural Semantic Spine (2026-09-14).** Phases 07-13,
61 plans, 183 tasks over 6 days; 386 commits, 441 files changed,
+130,172 / -1,527 lines against `f2490a8`.

`OpCall` is a real `core.OperationKind` wired at all six dispatch sites
(`check`, `corevalidate`, `interp`, `cgen`, `pathoracle`, `originvalidate`).
A multi-function Schway program parses, checks, is independently re-validated by
three non-importing peers, interprets, lowers to multi-function C17, and agrees
across interpreter / `-O0` / `-O3` / `-O3 -flto` on the five-axis comparator.
`Result` values carry payloads. `lang-repair` fixes interprocedural loan-liveness
defects through the JSON protocol alone.

**Phase 14 complete (2026-09-18) — Evidence Instrument and Honest Scoping.**
13 plans. The systemic finding below — *the instrument reports green because
something is wired, not because it runs* — now has a mechanism rather than a
set of patches: claims carry evidence grades with declared ceilings, every guard
carries a seeded-fault proof, the corpus-wide run record cannot report a result
it did not finish producing, and the `>=EXERCISED` bar is applied reflexively to
Phase 14's own validation artifact. Debt rows require a named owner; zero
`UNOWNED(none-yet-scheduled)` rows remain in `PHASE-14-DEBT.md`.

**Phase 16 complete (2026-09-26) — Branch/Match Emitter Port.** All 26 plans
are complete. Branch and match lowering now use `emitProgram`, and the retired
emitters are removed. The remaining foreign and by-pointer families stay an
explicit M004 boundary under NAT-08/NAT-09. The refreshed verifier passed all
4 goal truths, the full Go suite passed, and the already-complete 21/21
automated UAT was preserved without rerunning it.

**Shipped M003 — Computation and Honest Instruments (2026-09-26).** Phases
14–20: 7 phases, 85 plans, and 114 tasks. All 33 requirements are complete and
all seven phase verifications pass. The milestone audit is `tech_debt`: Nyquist
validation remains partial for Phases 14, 17, and 18, and four open unowned debt
items remain within the PRC-02 cap. Phase 21 belongs to M004 and is excluded
from these totals. See `.planning/milestones/M003-MILESTONE-AUDIT.md` for the
audit and `.planning/milestones/M003-ROADMAP.md` for the full phase record.

The evidence layer now grades claims against exercised behavior and pins moved
refusal frontiers. Schema `/2` event identity distinguishes repeated callee
activations. Branch and match lowering use the single `emitProgram` authority;
the retired emitters are gone, while three foreign/by-pointer families remain
explicitly refused under M004 ownership. Return facts are independent from
parameter facts; computed `Result` matches and U64 `OpConst` values run through
the checked core, interpreter, and native path.

**M003 entry-state gaps** (the snapshot as M003 began, corrected 2026-09-17
against the tree by the research fan-out — `.planning/research/M003/`; three
claims previously recorded here were wrong and are restated below):

- **DX-06 (partial, and worse than recorded)** — this section previously said
  the contract-boundary blame resolver was "built, tested, and compile-time
  exhaustive, but structurally unreachable," and that it would close
  automatically once `sameType(ReturnType, Parameter.Type)` lifted. **Both
  halves are wrong.** The *entire* blame subsystem — `resolveBlame`,
  `resolveCycleBlame`, `classifyDeclaredCause` — is test-only dead code: zero
  production call sites, and nothing ever sets `blameFact.Violated = true`
  (`classifyDeclaredCause` hardcodes `false`; `true` occurs only in
  `check_blame_test.go`). Lifting the invariant does **not** close it: a
  B1-shaped diagnostic needs a *user-declared* contract field that the
  declaring function's own admission cannot verify, and every
  `FunctionSignature` field today is producer-derived. B1's real home is
  separate compilation (previously forecast as M006). M003 corrects D-13-02b's reopening condition
  rather than restating the automatic-closure claim a third time.
- **DX-07 (partial)** — two of three new repairable interprocedural classes
  ship (`move_after_interprocedural_loan`, `wrap_call_in_try`);
  `use_matching_argument` was withdrawn empirically, its repair byte-identical
  to the original on every real trigger. See `PHASE-13-DEBT.md` D-13-10a. This
  one *does* close when the invariant lifts.
- **LTO evidence is scoped.** `emitProgram` produces one translation unit
  without `restrict` and refuses foreign/by-pointer program shapes. The
  opt-in Phase 21 comparison found semantic equality across interpreter,
  `-O0`, `-O3`, and `-O3 -flto` for one emitted multi-function fixture on
  Darwin arm64 with Apple Clang 21.0.0; see D-14-45 and
  `21-LTO-EVIDENCE.md`. This is not evidence about optimizer activity,
  performance, cleanup, or other hosts/toolchains. NAT-07's composition
  control remains hand-written C, not emitted by `cgen` (D-11-24).
- **Nyquist validation is partial, and "cheap to close" was only half true** —
  Phases 12 and 13 are literal frontmatter flips; Phase 07 is roughly an hour.
  But Phase 11 has six or more command cells elided with `…` and therefore
  un-runnable, and **three `go test -run` patterns across the VALIDATION corpus
  resolve to zero tests and exit 0 with `[no tests to run]`** — two in Phase 08
  (one naming a mechanism deleted in Phase 09) and one in **Phase 09**, a phase
  certified `validated` / `nyquist_compliant: true` on a green row whose
  evidence is a summary's prose. These are provenance defects in the measuring
  instrument, not coverage holes; the substance is intact.
- **10 open, unowned debt items** in three clusters: the single-function
  emitter work (D-11-02 → D-12-36 — a **port-then-delete**, not a deletion:
  `emitProgram` is a strict *subset* that refuses `Match`, `Blocks` and
  `ForeignContract`, emits zero `restrict`, and hardcodes
  `"live_resources":[]`), event identity (D-11-51 → D-12-21), and the
  single-type-per-function invariant (D-13-02b, D-13-10a).
- **D-13-34 points backwards** — M001's `testdata/phase6` move and borrow
  held-out pairs are alpha-renames of each other (match genuinely differs).
  Blast radius determined: DX-04's written claim survives on corpus-independent
  evidence; what is void is D-06-29's anti-overfitting inference for two of
  three source classes. M002 is unaffected.

**The systemic finding.** Those are not six unrelated defects. DX-06, the
`-flto` inertness, the zero-test VALIDATION patterns, and
`TestEveryMutationMovesItsClaimedAxis` silently skipping two rows are four
instances of one failure mode: *the instrument reports green because something
is wired, not because it runs.* It is the same mode the M002 audit caught in
the integration checker. M003 Phase 14 mechanizes the fix rather than patching
the instances.

## Shipped Milestone: M003 Computation and Honest Instruments

Completed 2026-09-26: 7 phases, 85 plans, 114 tasks, 33/33 requirements, and
7/7 passing phase verifications. The audit records non-blocking technical debt
and no requirement, integration, or flow gaps.

**Goal:** Prove a meaning survives *computation* — the first value Schway creates
rather than moves — on instruments that cannot report green for work that is
merely wired.

**Thesis:** M001 proved one meaning survives lowering. M002 proved it survives a
function boundary. M003 proves it survives computation, and makes the measuring
instruments honest enough that the claim means what it says.

**Delivered scope:**

- Evidence instruments that grade at EXERCISED+ or name themselves unreachable —
  a groundedness lint, a closed evidence vocabulary
  (`DEFINED | WIRED | REACHABLE | EXERCISED | MUTATION-KILLED`), and
  `.planning/UNREACHABLE-CLAIMS.md` for criterion-shaped defects. The three
  wrong claims above are corrected here.
- Event identity that survives a shared-leaf diamond — context-path invocation
  identity under a `lang.execution/2` bump, uniqueness moved to
  `(invocation, id)`, independently re-derivable by a non-importing peer
  because the call graph is a guaranteed DAG.
- One emission law instead of two — **completed in Phase 16**: branch/match
  lowering uses `emitProgram`, three emitters were deleted, and the remaining
  foreign/by-pointer families were formally cut with M004 ownership.
- A function's return type may differ from its parameter type.
- A branch that can discriminate a **computed** value — today `match` is a
  whole-function-body form; Phase 18 admitted computed scrutinees and exposed
  payload results after spike S-010. This is not general statement branching.
- Literals and `OpConst` — Schway can name a value it was not given.
- A moved refusal frontier, pinned by a test.

**Explicitly out of this milestone:** `if` as surface syntax, comparison
operators, `Bool`, `OpBinary`, loops and back edges, arity-N, DX-06 closure,
the three cut emitter families, and DX-08/09/11/13.

**The governing gate — constructibility precondition.** No requirement is
admitted unless its `.lang` fixture is checked in *first* as a refused frontier
fixture with its diagnostic pinned by a test; the phase gate is that the pinned
diagnostic moved. This is the only gate that fires *before* the work, and it
would have refused comparison operators on day one — a `Bool` that cannot be
branched on is the DX-06 failure mode for the fourth time.

**Historical risk assessment.** The M003 research anticipated ownership and
evidence changes at computed branches. Its percentage estimates were judgments
without calibration and are not current forecasts. Future estimates name the
specific changing assumption, executable witness, and decision gate instead.

**WIP limits.** M003 may not close with more than 5 unowned debt items. Every
debt item names an owning phase when it is recorded.

## Shipped Milestone: M005 Practical Computation (2026-10-10)

**Goal:** A developer can run ordinary Schway source that sums bounded U64
input and prints FizzBuzz with defined arithmetic, loop, and output behavior.

**Target features:**
- Define the admitted U64 arithmetic, comparison, remainder, Bool, and overflow
  behavior identically for the interpreter and native C17 path.
- Admit scalar loops with CFG fixed-point analysis; keep resource and loan state
  across loop back edges refused until independently proved.
- Run `sum_to_n` through the retained application route with bounded input,
  exact output, and arithmetic boundary behavior.
- Print FizzBuzz from ordinary Schway source with bounded text output, backed by
  independent interpreter/native expected answers and changed-assumption
  controls.

M004 is complete through Phase 25, including application execution, live
allocation ownership and discharge, transfer across calls and typed errors, and
the separate bounded shared/exclusive pointer successors. Its hosted receipts
remain scoped to their recorded revisions; M005 adds no new resource claim.

## Milestone Arc

[PRODUCT-ROADMAP.md](PRODUCT-ROADMAP.md) is the living authority for future
capability order. The M003-era M004–M006 forecast is superseded prospectively:

| Horizon | Capability | Promotion gate |
|---|---|---|
| Shipped M005 | Defined scalar computation and loops; `sum_to_n` followed by FizzBuzz | Admit defined arithmetic and scalar loop state with independent interpreter/native evidence; keep resource and loan carries refused |
| Mid term | Reusable byte-oriented libraries, arity-N/aggregates as needed, local modules, bounded JSON; HTTP as a separate branch | A consumer needs each abstraction; untrusted byte/error/resource behavior is specified |
| Long term | Separate compilation/contracts, ecosystem, richer generics/effects/concurrency | Concrete clients and measured limits justify the added semantic surface |

**Amended 2026-10-12 (Phase 32, DX-19): the provisional checksum frontier is retired.**
`examples/checksum.schway` was retired and deleted; its tombstone sha256 is
`ae95e550df67114c3464dda9cc24779f8a2efd69bd2b99dd0ccb2acc235cc154`. Its last
pinned refusal was `syntax.expected_rbrace` at 521-527. The supported route is
`examples/checksum/checksum.schway`, and the guide is
`examples/checksum/README.md`. The frozen goldens keep the path through the
dated retired-path allow-lists `phase28RetiredEntries` and `phase30RetiredPaths`
(absence allowed, re-appearance a failure), and `TestPhase32ChecksumFrontierRetired`
guards the retirement.

D-12-43 already became constructible in Phase 18; aggregates are not its
reopening gate. DX-06 remains tied to a separately checked user-declared
contract, without a speculative milestone number.

## Requirements

### Validated

- ✓ Checked U64 addition and defined overflow; `<` and Bool branching; scalar
  U64/Bool loops with independent deterministic fixed points and explicit
  ownership/resource/loan/provenance back-edge refusals — M005 Phase 26
  (U64-01, FLOW-01, FLOW-02).
- ✓ Ordinary `sum_to_n` source runs through the public app route with literal
  expected answers, rejected-input/overflow tuples, reached mutations, dynamic
  event identities, and capacity separation — M005 Phase 26 (APP-07).

- ✓ A clean checkout can format, check, interpret, and natively run a small
  canonical Schway program through one production command — M001 (Phase 1;
  `lang format` / `lang check` / `lang run --engine={interp,native}`).
- ✓ One portable typed core gives ownership, borrowing, cleanup, and foreign
  boundaries the same meaning in the checker, interpreter, and native path —
  M001 (Phases 2-4; `check`, `corevalidate`, `originvalidate`, `interp`, `cgen`
  independently derive the same facts).
- ✓ Interpreter, native `-O0`, and native `-O3` executions produce equivalent
  semantic outcomes and events for the milestone corpus — M001 (Phase 5;
  five-axis comparator, `-flto`-proven-non-inert `-O3` tier, ASan/UBSan lane,
  adversarial + bounded-enumeration corpora). Scope note: M001 ships without
  calls between Schway functions, so this holds intraprocedurally by construction.
- ✓ Invalid programs fail early with stable, bounded, machine-readable cause
  graphs and calm human projections — M001 (Phases 1-3, 6; `lang.diagnostic`,
  `lang explain` cause DAG, rustfix-style `diagnostic.Repair` applicability).
- ✓ Compact content-bound evidence manifests make successful work reusable
  without treating frontend claims as proofs of source correctness — M001
  (Phases 1-2, 6; `evidence --validate`, `internal/compiler/cache` which is
  structurally incapable of holding a verdict).
- ✓ The default edit loop remains fast, deterministic, offline-capable, and
  measurable; expensive evidence runs only when its distinct question is
  relevant — M001 (Phase 6; `risk_lanes.json` changed-risk lane selection,
  `internal/compiler/measure` p50/p95/CoV protocol, `qlt02_budget_manifest.json`).

- ✓ Storable/matchable `Result` values with payload-carrying alternatives —
  M002 (Phase 12; `OpConstructPayload`/`OpDestructurePayload` with a single
  shared `core.AlternativeNameForPayloadType` derivation read by `check`,
  `interp`, and `cgen`, and affine drop obligation held on the unmoved
  alternative). Closes D-04-30, deferred from M001 Phase 4. Scope note: a
  `data` declaration whose alternatives collide on payload type is refused
  (`check.duplicate_payload_type`) rather than resolved — see D-12-44.

- ✓ Calls between Schway functions: `OpCall` is a real `OperationKind` at all six dispatch
  sites, gated on callable ⊆ publishable (D-04-03) — M002 (Phase 07;
  SEM-04/05/06/07, both exhaustive-dispatch controls green, digest-bound
  signature summaries, no caller admission ever reads a callee body).
- ✓ Call-graph construction refuses cycles — direct, mutual, and indirect —
  with a named code, never a hang; the interpreter executes calls on a bounded
  stack with a documented ceiling — M002 (Phases 07, 10; SEM-07, SEM-08).
- ✓ Interprocedural loan liveness is derived in both admission layers from
  signatures alone and independently re-derived by a non-importing peer —
  M002 (Phases 08-09; OWN-06/07/08/09, EFF-02). **Closes D-03-02**, the one
  debt item knowingly carried past M001.
- ✓ Cross-function loan-endpoint differentials rebuild Phase 3's exhaustive
  enumeration at a declared bounded composition depth, with `pathoracle` and
  `originvalidate` re-deriving without importing `check` or `corevalidate` —
  M002 (Phase 10; QLT-04, TRU-02, TRU-03, TRU-04; all three transitive-
  dependency guards clean).
- ✓ Interpreter, `-O0`, `-O3`, and `-O3 -flto` produce equivalent semantic
  outcomes and events for the interprocedural corpus, with the tier proven
  non-inert by an engineered composition-only divergence — M002 (Phase 11;
  NAT-04/05/06/07, QLT-03/05/06). This is the claim M001 could not make,
  because M001 ships without calls.
- ✓ `lang explain`'s cause DAG stays bounded across function boundaries and
  names the function each cause step belongs to — M002 (Phase 13; DX-05).
- ✓ Every shipped claim is graded at EXERCISED or above, or names itself
  unreachable with an unblocking trigger, so no instrument can report green for
  work that is merely wired — M003 (Phase 14; evidence grades with WIRED /
  EXERCISED ceilings, seeded-fault proofs on every guard, a completion-witnessed
  corpus-wide run record with a measured timing margin, a live `//go:build`
  suppression allowlist, and owner-required debt rows). Validated in Phase 14:
  Evidence Instrument and Honest Scoping.
- ✓ Event identity distinguishes repeated activations of the same callee —
  M003 Phase 15 (`lang.execution/2`); diamond, sibling-call, and caller-owned
  callee identity controls are covered by the phase verification.
- ✓ One emission law lowers all admitted programs — M003 Phase 16 (NAT-08,
  NAT-09); branch/match lowering uses `emitProgram`, the retired emitters are
  absent, and the remaining foreign/by-pointer families are an explicit M004
  refusal boundary with assigned ownership.

- ✓ A function's return type can differ from its parameter type — M003 Phase 17;
  checker, core validator, origin validator, interpreter, and native emission
  preserve the separate facts, and the held-out JSON repair path is exercised.
- ✓ A branch can discriminate a computed value — M003 Phase 18; computed
  terminal matches and payload values execute through checked core, interpreter,
  and native emission.
- ✓ Schway can name a U64 value through a source literal — M003 Phase 19; canonical
  `OpConst` facts are independently validated and agree across interpreter and
  native optimization tiers.
- ✓ The post-M003 evidence corpus and refused checksum frontier are reconciled —
  M003 Phase 20; remaining partial Nyquist coverage is documented in the audit.
- ✓ A retained native application accepts bounded caller U64 input, runs once
  with ordinary streams, and keeps execution evidence separate from explicit
  differential replay; local build inputs use a closed manifest and content-bound
  identity — M004 Phase 22 (APP-02–06, FFI-02, EVD-11). The current evidence is
  macOS-only; host dependency closure remains incomplete/non-cacheable, and local
  C behavior is still trusted input rather than a Schway foreign call.

- ✓ U64 equality/remainder with static zero-divisor refusal and defined dynamic
  checked failure — M005 Phase 27 (U64-02).
- ✓ Exact bounded FizzBuzz through ordinary source, with the shared 65,536-byte
  output limit for both M005 programs — M005 Phase 27 (APP-08, APP-09).
- ✓ Independently pinned cross-program and dual-host evidence, reached negative
  controls, and clean-checkout documentation — M005 Phase 27 (EVD-12, DX-16);
  guide outputs made clean-tree safe in Phase 27.3.

### Active

M006 requirements live in [REQUIREMENTS.md](REQUIREMENTS.md) (IDs are assigned
at milestone kickoff). In scope: the bounded checksum byte utility, the local
module boundary with a second consumer, and M005 carry-over N2.
Tracked but not scheduled: N4 (guide n ≤ 14 note), R1b, and the D-27.1-10
dispositions.

Phase 28 (2026-10-10) closed N1 (EVD-13: `app verify` keeps the named capacity
reason at any n via a fixed-size outcome digest, report schema
`lang.app-verification/2`), W2/W3 and N3 (DX-17: one `exitcode` table, admission
before any tool, so check/build/both run engines give one exit class per
source), and DX-18 (`main` wins entry selection; functions `main` never calls
are pruned after full-program admission). Local verifier passed 4/4; hosted CI
not yet run at this revision.

Phase 29 (2026-10-11) closed BYTE-01, BYTE-02, BYTE-03 and APP-10: ordinary
source reads a read-only `ByteView` in a scalar loop (`byte_len`, `byte_at`)
with a bounds check that survives -O3 and LTO (exit 65,
`schway: byte index out of bounds`), and Adler-32 is pinned on both engines
over a test-only foreign fixture owner (`schway build --manifest` plus
`schway app run -- N`). Verifier 5/5, threats 35/35 closed, UAT 6/6 automated;
hosted run 38186068052 kept the spike 008 cap at 65,536 bytes.

Phase 30 (2026-10-12) closed MOD-03, MOD-04 and MOD-05. An entry writes
`import adler32 { fn adler32 }` and calls the imported function by its bare
name. Modules resolve only to exact-case regular files under the entry's
directory: symlinks, `..`, FIFOs and case aliases are refused. Only exported
names are visible, and corevalidate re-derives that independently. Cycles,
missing modules, header/path mismatches, collisions and the caps (16 modules,
depth 8, 4 MiB) are each refused with a named `module.*` code. A diagnostic
inside an imported file carries `unit=<module>`. No-import output is
byte-identical to a pre-change golden. The Phase 29 Adler-32 program now
imports its library from one file, with the 10 pinned answers on both engines.
Verifier 4/4, threats 36/36 closed, UAT 6/6 automated. Hosted run 38283564836 (b6f45bb8) passed
on ubuntu and macOS. A formatter defect carried over from Phase 29 is tracked as a
todo.

Phase 32 (2026-10-12) closed DX-19 and EVD-15, and with them M006.
`examples/checksum/README.md` is now an executable guide. `TestPhase32GuideCleanCheckout` copies the
tracked files into a temp dir and runs the guide's commands, matching exit code, stdout and stderr. The
provisional `examples/checksum.schway` is retired (see the dated amendment above).
Hosted run 38315799626 (headSha 6c304271) passed on ubuntu and macOS. Its receipts
(`32-RECEIPT.json`: 7 programs, 39 cases per host) are bound to codedigest
`c1b4e85d…`, which covers the code paths and excludes `.planning` and `*.md`. Edits to living docs
therefore keep the receipt current, and a later code change reopens it (D-10a). The formatter
defect is fixed, and the hosted macOS aggregate now has a 20-minute test budget. Verifier
3/3, UAT 0 human items. The 7 code-review findings (2 Warning, 5 Info) all fail closed and
are deferred past the freeze in one todo.

Phase 32.1 (2026-10-12, inserted from the M006 audit) tightened DX-17, DX-18 and MOD-03 and
re-took EVD-15. `run` on an unbindable entry refuses usage/64 `tool.run_input_required` on both
engines. `check` refuses a `ByteView` `main`. The interpreter matches native on abort-like
defects. The execution peer adopts `main` precedence independently. One accessor
(`check.ModuleAwareRoutes`) names each route's module boundary, the route audit pins it, and the
usage text lists every route. Hosted run 38335175306 (headSha 9f8f878f) passed on ubuntu and macOS.
The receipts are re-bound at freeze codedigest `b9253457…`. Verifier 6/6, threats 21/21 closed,
UAT 31/31 automated with 0 human items. Review warnings WR-01/WR-02 and a coverage-`kind` pin are
deferred to the next code-moving phase, because any code commit reopens the receipts.

Atomic IDs and acceptance for M005 are archived in
[milestones/M005-REQUIREMENTS.md](milestones/M005-REQUIREMENTS.md).

### Deferred — Named Landing, Not Dropped

- Byte views, bounds/indexing, arity-N, aggregates, and local modules → reusable
  byte-oriented library consumers.
- DX-06 / B1 blame → the separate-compilation contract boundary (D-13-02b).
- General strings, recursive data, collections, and generics → promote only
  the minimum needed by a selected library/program. FizzBuzz needs none of them.
- CI lane duplication and partial historical Nyquist coverage → named
  maintenance candidates; fix when they block or measurably slow a selected
  capability, preserving the evidence each lane supplies.

### Out of Scope

- General effect rows, resumable handlers, and runtime dependency injection —
  preserve typed-core seams, then design from a working ownership spine.
- Async/await, actors, schedulers, backpressure, supervision, or managed cyclic
  heaps — defer until sequential resource semantics survive native lowering.
- Package registry, remote cache, LSP, MCP server, web/SQL/TLS/GUI kits, AI
  inference, and distributed workflows — later milestones built on the stable
  compiler service contract.
- Full generics, subtyping, variance, partial moves, borrow across suspension,
  reflection, macros, and user-defined precedence — they would broaden the
  search space before the kernel oracle is trustworthy.
- A permanent compiler host or final optimizing backend decision — Go 1.24 and
  readable C17/Clang remain reversible Stage 0 choices; neither M001 nor M002
  gave a reason to revisit either, and M002's multi-function C17 emission and
  `-flto` tier exercised the backend harder than M001 did.
- Recursive programs — `callgraph`'s three-color DFS refuses every cycle before
  any liveness derivation runs, so a liveness differential over an admitted
  recursive program structurally cannot exist (TRU-04, D-09-22). Revisit only
  alongside a bounded-recursion admission design.

## Context

**Shipped M001 — Source-to-Native Semantic Spine (2026-09-07).** 6 phases,
62 plans, 174 tasks over 5 days. Details in
`.planning/milestones/M001-ROADMAP.md`.

**Shipped M002 — Interprocedural Semantic Spine (2026-09-14).** 7 phases
(07-13), 61 plans, 183 tasks over 6 days; 386 commits, 441 files changed,
+130,172 / -1,527 lines, 61 new `.lang` fixtures. ~102,280 lines of Go across
25 packages; `go build ./...`, `go vet ./...`, and `go test ./...` are clean.
Details in `.planning/milestones/M002-ROADMAP.md`.

**Shipped M003 — Computation and Honest Instruments (2026-09-26).** Phases
14–20: 7 phases, 85 plans, 114 tasks, and 33/33 requirements verified. The
milestone audit is `tech_debt`: partial Nyquist coverage in Phases 14, 17, and
18 plus four open unowned debt items within the five-item limit. Phase 21 is
assigned to M004 and excluded. Full records are in
`.planning/milestones/M003-ROADMAP.md`,
`.planning/milestones/M003-REQUIREMENTS.md`, and
`.planning/milestones/M003-MILESTONE-AUDIT.md`.

What exists now is a real vertical compiler that crosses function boundaries:
lossless canonical frontend → typed core with a per-function CFG and `OpCall` →
an affine ownership and borrow checker with interprocedural loan liveness → three
independent, non-importing validators (`corevalidate`, `originvalidate`,
`pathoracle`) → deterministic interpreter oracle on a bounded call stack →
multi-function readable C17 through Clang at `-O0`/`-O3`/`-flto`, with an
ASan+UBSan lane, a core-level HDD reducer, and a content-bound evidence/cache
layer keyed on the call-graph closure. `Result` values carry payloads.
`cmd/lang-repair` fixes interprocedural defects through the JSON protocol alone.

**M001's structural gap is closed.** `OpCall` landed at all six dispatch sites,
D-03-02 closed in Phase 09, and interprocedural `-O3`/LTO equivalence is now
inside the proof scope with an engineered composition-only negative control
proving the tier non-inert.

**M002's principal outstanding liability is its debt register, not its
requirements.** 77 items across 7 registers, 26 closed, 0 open blockers, but 10
open and unowned — which is why the milestone audit's status is `tech_debt`
rather than `passed`. The three clusters, their dependency chains, and the two
honest partials (DX-06, DX-07) are enumerated in `## Current State` above and in
full in `.planning/milestones/M002-MILESTONE-AUDIT.md`.

**One process observation worth carrying forward.** The M002 integration checker
returned an unqualified "all 31 requirements satisfied, ready for release,"
marking DX-06 and DX-07 `✓ WIRED` on the strength of wiring alone. Its
structural findings held under spot-check and were adopted; its
requirement-satisfaction column was rejected against the tree. An integration
checker that grades requirements from wiring will systematically convert an
honest partial into a false green, because wiring is exactly what a structurally
unreachable defect class still has.

The design corpus in `wiki/` records intent, primary-source research, synthesis,
decisions, examples, and open questions with explicit provenance. Five bounded
workbenches in `.planning/spikes/` tested local affine ownership, CFG
edge-specific loan liveness, separate-compilation origins and generic abilities,
independent typed-core certificate validation, and native C ABI/provenance/
cleanup hazards; their surviving contracts are now production code, and their
hazards are reconciled row-by-row in `qlt01_registry.json`.

Canonical planning inputs:

- `.planning/LANGUAGE-MATURITY.md` — where the language actually stands
  (expressiveness vs assurance) versus how the roadmap sounds. **Read this
  before answering any "how far along are we" question.**
- `.planning/STANDING-VERDICTS.md` — dependency verdicts, anti-features, and
  load-bearing facts already researched. Reopen only on new evidence.
- `wiki/real-frontend-core-ir-entry-plan.md`
- `wiki/semantic-kernel-contract.md`
- `wiki/semantic-kernel-probes.md`
- `wiki/ownership-evidence-roadmap.md`
- `wiki/compiler-and-feedback-latency.md`
- `wiki/compute-efficiency-constitution.md`
- `wiki/residual-uncertainty-register.md`
- `.planning/spikes/MANIFEST.md`

## Constraints

- **Feedback latency**: Optimize for the whole generate → verify → run → observe
  → repair loop; report cold and warm distributions rather than one flattering
  number.
- **Correctness**: Safe code has defined behavior; optimizer attributes and FFI
  guarantees must derive from checked facts.
- **Low-level capability**: Usable ownership, borrowing, deterministic cleanup,
  target layout, and C interop are milestone requirements, not post-v1 add-ons.
- **Runtime posture**: No mandatory global tracing heap or service runtime; the
  first native slice uses stack/inline values and explicit resources.
- **Bootstrap**: Go 1.24 standard library hosts Stage 0; readable C17 emitted to
  installed Clang is the development native path.
- **Dependencies**: Prefer standard-library-only and shallow audited boundaries;
  every dependency must earn more than copy-local implementation.
- **Source authority**: Text remains Git-friendly and formatter-owned; parsing is
  lossless enough to preserve comments and produce stable recovery.
- **Evidence**: Deterministic fixtures, negative controls, properties, mutation,
  differential execution, and sanitizers answer distinct questions and have
  explicit cost lanes.
- **Portability**: macOS and Linux are the initial host priorities; wire contracts
  and target facts must not accidentally encode the current Apple arm64 host.
- **Security**: Untrusted input, unsafe operations, FFI, secrets, and build
  authority stay explicit; no generic `untaint` or ambient authority escape.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Build M001 as vertical source-to-native slices | Cross-layer feedback reveals semantic mistakes sooner than disconnected compiler layers | ✓ Good — every phase surfaced consumer gaps (e.g. 03-02's native execution-document validator) that a layered build would have hidden until integration |
| Use Go 1.24 stdlib for Stage 0 | Fast builds, current spike evidence, simple offline onboarding, and low dependency depth | ✓ Good — zero external deps through 6 phases; `go/ast`/`go/parser` doubled as a structural-test substrate |
| Emit readable C17 through Clang first | Exposes real ABI and optimizer behavior while keeping backend replacement reversible | ✓ Good — real `-O3`/LTO/`restrict` divergences and ASan findings only exist because the backend is real |
| Make the deterministic interpreter the semantic oracle | Prevents C output or optimizer accidents from defining the language | ✓ Good — the five-axis comparator's authority rests on it; it caught two real `--engine=native` bugs in Phase 4 |
| Use affine values with separate copy/drop/share/send/escape abilities | Avoids conflating distinct semantic and cost properties | ✓ Good — abilities derived independently through aggregates and generics; one accepted override (OV-02-01) |
| Infer local borrow ends at CFG last use; declare public origins | Preserves local ergonomics without hiding public ABI relationships | ✓ Good — `loanLivenessFixpoint` became the sole liveness law after a 230,692-comparison shadow run, and M002 Phase 09 closed the interprocedural half (D-03-02) with an independently re-deriving peer |
| Validate compact typed-core facts independently at trust crossings | Gives cheap cache/release evidence without claiming source-to-core proof | ✓ Good — and its known escape is executable, not hypothetical (`escape:coordinated-source-to-core-false-claim`) |
| Keep syntax provisional but canonical | Enables real parser/formatter evidence without freezing decorative choices early | ✓ Good — losslessness and idempotence held through generated programs and malformed AI edits |
| Treat structured diagnostics and evidence as a versioned product API | AI effectiveness depends more on precise verifier feedback than exotic syntax | ✓ Good — the `/0`→`/1` schema bump landed across 12 literal sites with `/0` bytes provably frozen; `cmd/lang-repair` repairs 5 defect classes through the JSON protocol alone |
| Defer `OpCall` and interprocedural equivalence out of M001 | Landing a new `OperationKind` at six dispatch sites alongside alias facts, sanitizers, the reducer, and the QLT-01 registry matched the fingerprint of the failures that cost Phases 2-4 extra remediation rounds | ✓ Good — M002 spent 7 phases and 61 plans on exactly that one addition, which vindicates the deferral; `OpCall` is now real at all six sites |
| Historical D-12-43 unconstructibility ratification (2026-09-13) | Phase 12 could not expose the wrong payload slot through its grammar; the historical ratification and measurements remain archived | Superseded by Phase 18's source-constructible payload return and wrong-slot mutation; the dated debt disposition records current evidence |
| Cut M002 phases where a *gate* becomes meaningful, not where implementation could parallelize | A new `OperationKind` at six dispatch sites has no safe partial state; the useful boundary is "can this claim now be independently checked?" | ✓ Good — all 7 phases verified passed, and each gate caught something (Phase 09's requirement-vs-traceability defect, Phase 10's fixture escape hatch, Phase 12's blocker) |
| Split OWN-05 into OWN-05a/OWN-05b (D-09-37) rather than flip one row Complete on two of three derivers | A single-row overclaim is exactly the requirement-vs-code failure the debt registers exist to catch | ✓ Good — the split is why the milestone audit found zero orphaned or overclaimed requirements across three independent sources |
| Eliminate Phase 11's two human-judgment items with tests rather than adjudicate them | A debt note is read once; a test runs on every CI invocation | ✓ Good — `TestSeedEntryHazardIsReal` and `session_admission_divergence_test.go` both fail in *both* directions, catching a silently resolved divergence as well as a new one |
| Keep DX-06 and DX-07 evidence-based instead of weakening their criteria | DX-06 needs a user-declared contract field its declaring function cannot verify; DX-07's `use_matching_argument` repair was byte-identical on real triggers | ✓ Good — M003 closes DX-07 through a sealed held-out repair proof; DX-06 remains owned by the future separate-compilation contract capability (D-13-02b) |
| Reject the integration checker's requirement-satisfaction column while adopting its structural findings | It graded requirements from wiring, and wiring is exactly what a structurally unreachable defect class still has | ✓ Good — caught two would-be false greens (DX-06, DX-07) that contradicted both Phase 13's own verification and a `grep` of the tree |
| Keep the first public application ABI to checked `U64 -> U64` with one bounded opaque argv token | A small caller-selected input/output witness proves ordinary application execution without prematurely defining general strings or IO | ✓ Phase 22 — inputs `7` and `42` produce independent results through one retained-app launch; wider inputs remain deferred |
| Separate ordinary app capture from explicit differential verification | Application effects must occur once on the caller's route; evidence capture alone cannot prove semantic correctness | ✓ Phase 22 — one run emits ordinary streams; `app verify` separately checks isolated/replayable inputs against independent expected values |
| Keep local C build authority closed and mark incomplete host closure honestly | Declared inputs and stable identity do not prove the installed SDK/linker/runtime inventory is complete | ✓ Phase 22 — manifest and known inputs are content-bound; receipts remain `dependency_closure: incomplete` and `cacheable: false` |
| Keep cache refusal codes stable while retaining typed Clang probe causes and finite subprocess deadlines | Callers keep a fail-closed contract while diagnostics distinguish timeout, launch, and command failures; bounded work remains cancellable | ✓ Good — three default-parallel full suites, capped-parallel and race suites, vet, and build all pass under verified captured receipts |
| Reuse existing macOS/Linux full and race CI lanes when focused coverage adds no distinct signal | Recurring CI value must justify its runtime and maintenance cost | ✓ Good — Plan 08's unchanged CI blob and `not_added` disposition remain verified after G-18-16 closure |
| Route branch and match lowering through the public `emitProgram` authority | A single emitter law prevents admitted source shapes from depending on a legacy backend | ✓ Good — M003 Phase 16 passed all 4 verification truths; foreign and by-pointer families retain explicit M004 ownership |
| Require a fixture-first constructibility gate for every admitted language requirement | It catches built-but-unreachable behavior before production implementation begins | ✓ Good — M003's phases used refused frontier fixtures and pinned diagnostics to make each new surface constructible |
| Separate application execution from differential replay | Real IO cannot safely run once per optimization tier on the user's input | Adopted for M004; explicit verification uses isolated/replayable inputs |
| Lower bounded pointer borrows without `restrict` | Schway exclusivity and an optimizer alias promise are distinct obligations | Prospective D-16-07 amendment; retain per-family proofs and broad-shape refusals |
| Maintain a program-driven living roadmap | The user wants automatic planning suggestions and steady progress toward usable software | Agent-executed planning reviews in AGENTS.md; no background service or new planning framework |
| Start loop admission with independently derived U64/Bool scalar fixed points | A runnable sum proves useful iteration without widening resource/loan authority | ✓ Phase 26 — exact source-to-app answers and independent refusals pass on both hosts |
| Distinguish repeated events by invocation/site occurrence while preserving legacy bytes | Static event IDs alone alias loop iterations; zero omission preserves existing documents | ✓ Phase 26 — source-authored copies emit 0/1/2 and independent mutations reject forged sequences |
| Validate checked-add terminal evidence separately from arithmetic semantics | Structural event membership cannot prove operand arithmetic; native size admission must include possible failure records | ✓ Phase 26 CR-01 — peer attribution/outcome and N−1/N capture-bound controls pass; actual overflow is exercised in each engine |
| Name generated-C event-capacity exhaustion (exit 69 + exact stderr) instead of aborting | A signal hid a bounded-evidence limit as a crash; a named outcome lets app verify report `unsupported` without masking a mismatch | ✓ Phase 27.3 — n=14 passes, n=15 unsupported, wrong expected stays fail; one-level exhaustion fails as divergence (WR-01) |
| Insert decimal tech-debt phases from the milestone audit before closing | Audit warnings W1/W4 and R1 were concrete user-facing defects, not cosmetic | ✓ Good — 27.1–27.3 closed R1, W1, W4 with hosted evidence; residual N1–N4 tracked |
| Compose local modules as one merged whole program with closed in-root resolution | Separate compilation and per-module interfaces are not needed for a second consumer, and a closed resolver (no search path, no symlinks, exact case, one IO site) keeps the import surface a non-escalating boundary | ✓ Phase 30 — imported Adler-32 gives identical answers on both engines; no-import bytes unchanged against a pre-change golden; module digests in identity are Phase 31 (MOD-06) |
| Keep post-freeze work `.planning`-only while hosted receipts bind the code digest | `scripts/codedigest.sh` hashes the HEAD listing of the code paths, so even a test-only commit makes the living claims `stale_bound_claim` and forces a re-dispatch | ✓ Phase 32.1 UAT — a test pin (31d71e94) was reverted byte-for-byte (7ab2f48e) and deferred; the digest returned to `b9253457…` |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition:**
1. Move disproved active requirements to Out of Scope with evidence.
2. Move verified requirements to Validated with a phase reference.
3. Add requirements exposed by reduced counterexamples.
4. Record decisions and their measured outcomes.
5. Check that the description and core value have not drifted.
6. Refresh PRODUCT-ROADMAP and LANGUAGE-MATURITY when observed facts change;
   propose the three most useful next capabilities with examples and checker gates.

**After each milestone:**
1. Review all sections against shipped behavior.
2. Reconfirm the core value.
3. Audit deferred scope and preservation seams.
4. Update context with adopter, performance, and correctness evidence.

---
*Last updated: 2026-10-12 after Phase 32.1*
