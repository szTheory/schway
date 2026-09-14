# Phase 10: Trusted Interprocedural Oracle - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-09-11
**Phase:** 10-trusted-interprocedural-oracle
**Areas discussed:** pathoracle's cross-call model, originvalidate's program reach (D-09-51), Import-control enforcement, D-09-53's blast radius, Call stack shape and ceiling, Drop ordering and nonlocal exit, OWN-05b independence, The Pitfall-4 stack probe, QLT-04's composition depth, The four-way differential harness, "interp must be stable", Scope-cut trigger

**Mode:** advisor (`minimal_decisive` calibration, `technical_background: true`).
All twelve areas were selected by the developer, who directed the standing
fan-out mandate: research each decision point across all relevant
stakeholder-role lenses, consider pros/cons/tradeoffs/anti-patterns/best
practices/footguns, draw on cross-ecosystem prior art, run an adversarial pass,
and synthesize one-shot recommendations — with the synthesized recommendation
adopted in every case.

Twelve `gsd-advisor-researcher` agents ran in parallel. Every load-bearing
factual claim was then re-verified against the shipped tree by the orchestrator;
corrections are recorded inline in CONTEXT.md at the decision they change.

---

## pathoracle's cross-call model (TRU-03)

| Option | Description | Selected |
|--------|-------------|----------|
| Path composition | Recurse into the callee via pathoracle's own `EnumeratePaths`/`linearizePath`, splicing callee paths into the caller's. Preserves "enumerate and replay, never summarize." Needs a second declared depth cap. | ✓ |
| Contract hop | Treat `OpCall` as one hop reading the callee's declared signature, mirroring `originvalidate`'s `OpForeignCall`. Cheap and bounded, but makes pathoracle a third consumer of the same summary abstraction `check` and `corevalidate` already consult. | |

**User's choice:** adopt synthesized recommendation (composition).
**Notes:** Decisive factor verified in tree — `corevalidate.go:1203` and
`check.go:585/721/1806-1916` have *already* claimed the contract-hop pattern for
`OpCall`. A third hop would collapse all three procedures into the same shape at
exactly the call boundary TRU-03 is asked to prove something about. The
discriminating fixture (a callee with two paths, only one borrowing a parameter
into its return) became a required deliverable, D-10-14.

---

## originvalidate's program reach (TRU-02 / D-09-51)

| Option | Description | Selected |
|--------|-------------|----------|
| Narrow callee-contract map | Precompute a minimal ID-keyed origin-fact map once in `ValidatePublished`; thread it as an explicit parameter. Body-blindness becomes a property of the type signature. | ✓ |
| Pass `core.Program` down | Simpler signature, but makes `Functions[i].Linear` reachable from a walk whose contract is that it never reads a callee body — mechanism degrades to a comment. | |

**User's choice:** adopt synthesized recommendation (narrow map).
**Notes:** Not a green-field choice — `corevalidate`'s own D-09-03 fix shipped
this exact shape one phase earlier. Blast radius counted rather than estimated:
2 external production call sites, 9 in-package test call sites,
`BuildInterface`/`ValidatePublished` signatures unchanged so their ~19 call sites
are untouched. Sequencing settled as widen-then-use in two commits inside one
plan (D-10-06).

---

## Import-control enforcement (Criterion 1)

| Option | Description | Selected |
|--------|-------------|----------|
| Patch existing guards only | Add `corevalidate` to `originvalidate`'s forbidden list; keep direct-import `go/parser` scanning. Smallest diff, leaves the transitive blind spot open. | |
| Patch + harden to transitive | Same fix, plus convert the scan to a `go list -deps` transitive check against the same forbidden set. No new dependency. | ✓ |

**User's choice:** adopt synthesized recommendation (patch + transitive).
**Notes:** This area produced the fan-out's single most concrete finding, which
the orchestrator verified directly: `originvalidate`'s two guard tests
(`originvalidate_test.go:33` and `:57`, checks at `:50` and `:74`) forbid only
`/compiler/check` and `/compiler/ast` — **neither forbids `/compiler/corevalidate`**,
which Criterion 1 explicitly names. A previously-unrecorded gap. Rejected:
building `depguard`/a custom analyzer/an `internal/` restructure (CI's existing
`go test ./...` already makes a test-level guard build-failing); and
consolidating the guards (existing redundancy is worth more than the
maintainability win). Two residual weaknesses were recorded rather than papered
over: `originvalidate` imports `callgraph` while `corevalidate`'s guard forbids
exactly that import for itself (D-10-19), and every guard lives in the package it
polices (D-10-20).

---

## D-09-53's blast radius

| Option | Description | Selected |
|--------|-------------|----------|
| Apply the narrowing | Exclude `OpMove`/`OpCopy` from `deriveFunctionUsesParam`'s `default` arm, as the debt register proposes. | |
| Do not touch the code; correct the record | Retire D-09-53's characterization as a code defect; fix the fixture header and doc comment; re-file the real gap as a new, correctly-scoped item. | ✓ |

**User's choice:** adopt synthesized recommendation (correct the record).
**Notes:** The strongest result of the whole fan-out. The researcher applied the
register's own literal proposed narrowing in a throwaway patch and ran the full
suite: **8 named failures**, including `TestInterproceduralLivenessTwinPatternB`
— the exact test D-09-53's writeup claims is "UNAFFECTED." Six
independently-authored Phase 08/09 tests pin the current semantics as intended.
Orchestrator verification of the structural claims: `check.go:723-726` exempts
only `case core.OpReturn:` (the operation, not the chain), so the doc comment
overstates the implemented contract; and `corevalidate_peer_liveness.go:38-39`
defines `peerLoanCarryFact` with exactly one field, `ReturnsBorrowOfParam` —
**`corevalidate` has no `usesParam` peer at all**, so Pattern B's split cannot be
shown end to end regardless of what `check` does. A change that breaks
`TestCostCorpusLeafTemplatesDifferOnlyInTheCallee` (a mutation-discipline test)
degrades a detector rather than fixing one. Recorded as a reversal of a locked
Phase 09 item (D-10-27), with the real gap re-filed (D-10-28) and an explicit
note that this is **not** a second deferral (D-10-30).

---

## Call stack shape and ceiling (SEM-08)

| Option | Description | Selected |
|--------|-------------|----------|
| Explicit `[]frame` stack, iterative dispatch | Each `OpCall` pushes a heap frame and loops; depth checked before push; refusal modeled as a typed `Outcome`. | ✓ |
| Go recursion + depth counter | Minimal change, closest to today's three switches — but leaves the language and host limits coupled. | |

**User's choice:** adopt synthesized recommendation (explicit frame stack).
**Notes:** Settled by a verified Go-runtime fact rather than preference:
`recover()` cannot catch goroutine stack exhaustion (a fatal throw, not a panic),
so any Go-recursive design is architecturally incapable of satisfying "a named
refusal, not a host stack overflow." In-repo precedent found:
`callgraph.Order`'s iterative DFS exists for exactly this reason
(`callgraph.go:25-26`). The ceiling discussion produced the phase's sharpest
insight: `syntax/parser.go:16` enforces `maxFunctions = 1024` (verified), so
`MaxCallDepth` must sit **below** the structural ceiling to be reachable —
the opposite of `MaxPaths`, which sits above its reachable maximum and never
fires. `MaxCallDepth = 128`, exercised through the real pipeline with a
129-function generated source, never a synthetic `core.Program`.

---

## Drop ordering and nonlocal exit (SEM-09)

| Option | Description | Selected |
|--------|-------------|----------|
| Reuse `Event.FunctionID`, no schema change | Frame attribution via a field that already exists and is inlining-invariant. | ✓ |
| Add an explicit frame/depth field | Unambiguous frame identity, but a `-O3 -flto` build that inlines a callee has no frame to reproduce a depth from — manufacturing false divergences in Phase 11. | |

**User's choice:** adopt synthesized recommendation (no schema change).
**Notes:** Full enumeration of every way control leaves a callee frame produced a
finding that must be stated rather than quietly satisfied: `OpCall` is documented
as **never** a terminator (`core.go:559-564`), `OpFail` is ordinary typed-value
propagation, and the language has no exceptions or unwinding — so SEM-09's
"every nonlocal exit" clause is *narrow*, not vacuous, covering only the foreign
process-root landing pad extended to N frames and the SEM-08 depth refusal.
Verified: `execution.Event` already carries `FunctionID` at `execution.go:45`.
Also captured: `interp` must not be the sole authority on drop order (D-10-34).

---

## OWN-05b independence

| Option | Description | Selected |
|--------|-------------|----------|
| Keep the `corevalidate` import; enforce the boundary with a guard test | Derivation is the frame-partition execution behavior; a test fails the moment `interp` reads an ownership-bearing field of `corevalidate.Result`. | ✓ |
| Hoist `Validate` out of `Run` to sever the import | Cleanest import graph, but weakens `interp`'s fail-closed posture and touches every call site. | |

**User's choice:** adopt synthesized recommendation (keep + guard).
**Notes:** Verified — `interp` does import `corevalidate` (`interp.go:3-10`) and
`Run` calls `Validate` first (`:38`), but never reads `PeerSignatures()` or
`PeerSiteCoverage()`; that is true by omission, not enforcement, which the guard
test converts. Two honesty requirements were locked: the independence claim is
narrower than `check`↔`corevalidate`'s mutual non-import and must be written that
way (D-10-37), and with `Mode` hardcoded `"owned"` everywhere the natural-input
agreement is near-vacuous until a second parameter form exists (D-10-40). The
mutant-pairing requirement (which peer alone catches which mutant) became the
phase's concrete mutation obligation for this fact (D-10-41). A single in-package
helper across the three `OpCall` arms was ruled compatible with D-09-38, which
forbids *cross-peer* helpers.

---

## The Pitfall-4 stack probe

| Option | Description | Selected |
|--------|-------------|----------|
| Subprocess-driven probe | Re-exec the test binary env-guarded; child pins a small host ceiling via `debug.SetMaxStack`; parent asserts exit status and output. | ✓ |
| In-process headroom measurement | Fast and safe, but extrapolates from a documented default — an assertion wearing observation's clothes. | |

**User's choice:** adopt synthesized recommendation (subprocess).
**Notes:** Noted that no subprocess/`SetMaxStack`/`TestMain` pattern exists
anywhere in this repo today, so this is a genuinely new pattern rather than an
extension. Because the frame stack is explicit (D-10-21), language call depth
consumes O(1) host stack, so the probe's honest result is "the two limits are
structurally unrelated," which must be written up as such rather than dressed as
a near-miss stress test (D-10-43). Threshold fixed in source before the first run
(D-10-44). Seam follows `TerminatorKindsOverride`'s nil-default unexported
discipline.

---

## QLT-04's composition depth

| Option | Description | Selected |
|--------|-------------|----------|
| Declare depth 2 | Matches existing fixtures exactly, zero new corpus work — but leaves the "depth-3 and beyond" claim unverified forever. | |
| Declare depth 3 + bidirectional reachability check | Necessity floor (2) plus one sufficiency margin, CBMC-unwinding-assertion style, with a new fixture pair. | ✓ |

**User's choice:** adopt synthesized recommendation (depth 3).
**Notes:** Orchestrator verified all three load-bearing claims: no depth-3 fixture
exists anywhere (`grep -rn "depth3\|depth_3"` returns nothing);
`relay_depth2_accept.lang` composes an **owned pass-through**, so a borrow
surviving two hops is untested even at the depth that does exist; and
`corevalidate_peer_liveness.go:64-77` asserts depth-3-and-beyond correctness as
**prose with no fixture behind it** — precisely the implied-by-the-corpus failure
Criterion 4's wording exists to prevent. Product space computed and found small
(~a dozen structurally distinct cases) because `ParameterContract.Mode` is
unreachable at anything but `"owned"` — a collapse that must be named as an
explicit exclusion, not silently pruned.

---

## The four-way differential harness

| Option | Description | Selected |
|--------|-------------|----------|
| Extend `session_peer_gate_test.go`'s corpus walker | Reuses the substrate D-09-23 ratified; three peers already share the `core.LoanEndpoint` shape. | ✓ |
| Stand up a new dedicated harness package | Isolates the wiring, but creates two corpus walkers that can drift on what "the corpus" and "divergence" mean. | |

**User's choice:** adopt synthesized recommendation (extend).
**Notes:** The researcher corrected the orchestrator's own briefing: `pathoracle`
is **not** test-only — it is called from production at `session.go:2050,2067,2635`
and `session_phase7.go:449` (verified). The "four different answer shapes"
problem collapsed on inspection: `check`, `corevalidate` and `pathoracle` already
all speak `core.LoanEndpoint` (`core.go:762-769`), with only `corevalidate`'s
accessor unexported. `interp` genuinely cannot, so its comparison is metamorphic
and accept-side only — billed as "three-way on refuse, four-way on accept"
everywhere (D-10-53), following D-09-37's anti-overclaiming precedent. The
expected-divergence allowlist gains a debt-ID-and-landing-phase struct plus a
test asserting each ID is still open (D-10-54).

---

## "interp must be stable, not merely working"

| Option | Description | Selected |
|--------|-------------|----------|
| Documented convention | A paragraph asserting stability. Nothing enforces it. | |
| Sequenced mechanically-checked freeze | SEM-09 lands first, then golden corpus + determinism run + structural coverage floor + a named escalation path. | ✓ |

**User's choice:** adopt synthesized recommendation (mechanical freeze).
**Notes:** Verified — `interp_test.go` is 97 lines containing **exactly one**
test function, for the component about to double in size and become the authority
for everything native. The top review risk identified: freezing before SEM-09's
event work lands would break the freeze inside Phase 10 itself, so the ordering is
non-negotiable (D-10-56). D-10-32's no-new-field decision substantially de-risks
this but does not remove the rule. Coverage floor defined structurally
(operation-kind × execution-path × refusal-path, plus Mutation-Kill rows) rather
than as a percentage.

---

## Scope-cut trigger

| Option | Description | Selected |
|--------|-------------|----------|
| Token-cost ratio (~2×) at a scheduled mid-phase gate plan | Proven in-tree mechanism; observable mid-phase, not only in hindsight; independent of plan count. | ✓ |
| Named-checkpoint trigger | Ties the trigger to a deliverable, but Phase 10's plan sequence is `TBD`, so there is no non-arbitrary N to name yet. | |

**User's choice:** adopt synthesized recommendation (token ratio at a gate plan).
**Notes:** Verified — `ROADMAP.md:198-200` scopes the milestone 2× trigger to
**Phases 08 and 09 only**, so Phase 10 inherits none. Phase 09's trigger was
genuinely checked mid-phase and measured 0.17× (76,385 actual vs ~460,000
estimated) — a token ratio, not a plan count, which matters because Phase 09 ran
10 plans. Also verified: `TestDebtRegistersAreWellFormed` enforces register
**format only** and does not track deferral hop count, so the new no-third-deferral
rule (D-10-60) had to be declared rather than assumed enforced. D-09-51 was
ruled never-cut on the grounds that it is not carried scope at all — it is
Success Criterion 1 verbatim.

---

## Claude's Discretion

The developer directed that the synthesized recommendation be adopted for all
twelve areas under the standing fan-out mandate. Planner discretion remains over
plan decomposition and wave ordering (subject to four hard sequencing
constraints), Go identifier names, whether `pathoracle`'s composition lives in
`pathoracle.go` or a sibling file, the relative ordering of the two re-derivers
against `interp`'s call stack, and whether the re-filed `corevalidate`
`usesParam` peer gets a proposed landing phase now or is left open. See
CONTEXT.md's "Claude's Discretion" section.

## Deferred Ideas

Recorded in CONTEXT.md's `<deferred>` section: the `corevalidate` interprocedural
`usesParam` peer (newly identified, landing phase open); multi-function `cgen`
and interprocedural equivalence (Phase 11); `cgen`'s multi-frame nonlocal pad;
QLT-03's reachability register (Phase 11); cross-run summary caching (Phase 11 /
QLT-06); `Result` payloads (Phase 12, which re-opens D-10-32 if recursion is ever
admitted); `originvalidate`'s `callgraph` import asymmetry; guard-test
self-deletion resistance; and the still-un-owned `.planning/spikes` registry gap
carried from Phase 09.
