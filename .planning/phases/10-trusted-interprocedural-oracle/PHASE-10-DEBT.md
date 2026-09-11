---
phase: 10-trusted-interprocedural-oracle
recorded: 2026-09-11
status: accepted
disposition: planning-time
items: 12
blocking: 0
---

# Phase 10 — Declared Deferred Scope, Reversals, and Recorded Asymmetries (D-10-61)

**Written:** 2026-09-11, at *planning* time — not at phase end.

Per RETROSPECTIVE Key Lesson 4 and D-09-44's precedent, deferred scope is declared
in writing at the moment it is decided. Shape follows the mechanically-checked
debt-register format `TestDebtRegistersAreWellFormed`
(`internal/compiler/session/session_test.go:2613`) enforces on every `*-DEBT.md`
register: an `items:` count matching the `## Items` table, one `### <ID>` detail
section per row, a severity from the closed vocabulary (blocker, warning, info),
and a non-empty landing phase.

**One inherited Phase 09 row is REVERSED here.** D-09-53's *diagnosis* is
falsified (D-10-27), its real content re-filed under a different scope (D-10-28),
and the distinction between "reversed premise" and "deferred twice" is stated
explicitly (D-10-30) so no future reader records it as a third deferral.

**The reversal rests on a re-executed result, not on a document.** The
planner re-ran D-09-53's own literal proposed narrowing (a throwaway patch adding
`case core.OpMove, core.OpCopy:` as an inert arm before `default:` in
`deriveFunctionUsesParam`, `check.go:729-730`) against the shipped tree on
2026-09-11 and ran `go test ./...`. Result, verbatim: `TestDeriveFunctionUsesParamBehaviors`,
`TestInterproceduralLivenessTwinPatternB`, `TestInterproceduralLivenessTwinPatternBRealFixtures`,
`TestSummaryDerivationTwoHopChainPropagates`, `TestSummaryDerivationIsOnePassPerFunction`,
`TestSummaryDerivationRequiresProgramOrder`, `TestCostCorpusLeafTemplatesDifferOnlyInTheCallee`,
and `TestInterproceduralDiagnosticOrderingStability` all fail in package `check`,
plus `TestNoUndeclaredCheckPeerDivergenceAcrossCorpus` newly fails in package
`session` — the check/corevalidate divergence D-10-27 predicted. The patch was
reverted and `git diff internal/compiler/check/check.go` is empty.

---

## Items

| ID | Source | Threat/Req | Severity | Landing phase | Item |
|---|---|---|---|---|---|
| D-10-19 | 10-CONTEXT.md (D-10-19) | TRU-02, TRU-03 | info | Not scheduled — recorded, not resolved; revisit only if `callgraph` gains a compiler-internal dependency beyond `core` | RECORDED ASYMMETRY, deliberately not fixed. `originvalidate` imports `internal/compiler/callgraph` (`originvalidate.go:18`, calling `callgraph.Order`), a package `check` also imports. `corevalidate`'s own guard explicitly FORBIDS that same import for itself (`corevalidate_endpoint_internal_test.go:107`, per D-07-19). The two re-derivers are therefore held to different standards today. `callgraph` is a leaf utility whose only compiler-internal dependency is `core`, so the asymmetry is defensible — but it is defensible by REVIEW, not by mechanism, and each package's forbidden list is hand-curated. Phase 10 does NOT change `originvalidate`'s `callgraph` import; it records the asymmetry so a future reader does not discover it as an undocumented inconsistency |
| D-10-20 | 10-CONTEXT.md (D-10-20) | TRU-02, TRU-03, T-10-01 | warning | Not scheduled — accepted as out of scope for a test-level mechanism, consistent with Success Criterion 1's own "build- or test-level" wording | RESIDUAL WEAKNESS OF THE IMPORT GUARD, stated rather than papered over. Every import guard this phase hardens lives inside the package it polices, so a single commit can add the forbidden import AND edit the assertion together. Upgrading the guards from a direct `go/parser` scan to a transitive `go list -deps` check (D-10-17) does not fix this; neither would consolidating them into one shared table (which D-10-18 rejects for an unrelated reason). Closing it would require an out-of-package or CI-owned control. Accepted, because Criterion 1 asks for a build- or test-level mechanism and `go test` already fails CI's `checks` job (`.github/workflows/ci.yml`) |
| D-10-27 | 10-CONTEXT.md (D-10-27), REVERSES PHASE-09-DEBT.md D-09-53; re-executed as a planning pre-flight 2026-09-11 | OWN-09 | warning | Phase 10 — plan 10-06 (the artifact correction lands at the mid-phase gate) | REVERSAL. D-09-53 records `deriveFunctionUsesParam`'s `default: usesParam = true` arm as a DEFECT contradicting the function's own doc comment, and names `TestInterproceduralLivenessTwinPatternB` as "UNAFFECTED." That DIAGNOSIS IS FALSIFIED. Applying the register's own literal proposed narrowing in a throwaway patch and running `go test ./...` produces 8 named failures in package `check`, including `TestInterproceduralLivenessTwinPatternB` itself, plus a newly-surfaced `check`/`corevalidate` divergence in `TestNoUndeclaredCheckPeerDivergenceAcrossCorpus`. Six independently-authored, currently-passing Phase 08/09 tests PIN THE CURRENT SEMANTICS AS INTENDED. Structural verification: `check.go:723-726` exempts ONLY `case core.OpReturn:` — the operation itself, never the chain — so the doc comment's "the entire parameter-derived reference chain leads directly to its own terminating OpReturn" overstates what the code implements. What is actually wrong is `twin_b_accept.lang`'s header comment and the doc comment's wording, NOT the derivation |
| D-10-28 | 10-CONTEXT.md (D-10-28); re-files the real content of PHASE-09-DEBT.md D-09-53 | OWN-05a, OWN-09, TRU-04 | warning | OPEN — deliberately unassigned. Not assigned under time pressure, per 10-CONTEXT.md's explicit planner-discretion grant; candidate homes are Phase 11 (adjacent to native lowering already touching `check`) or M003 | THE REAL GAP, correctly scoped and re-filed. Even WITH D-09-53's narrowing applied, Pattern B's split still cannot be demonstrated end to end, because `corevalidate` HAS NO `usesParam` PEER AT ALL. Verified directly: `corevalidate_peer_liveness.go:38-39` defines `peerLoanCarryFact` with exactly one field, `ReturnsBorrowOfParam` — Pattern A's fact, D-09-51's scope. `corevalidate`'s liveness walk therefore treats every `OpCall` argument unconditionally as a use and refuses the accept twin regardless of what `check` derives. The honest debt item is "corevalidate has no interprocedural usesParam peer" — a MISSING SECOND DETECTOR, not a broken existing one, and Key Lesson 2's exact failure condition (nothing but `check` watches this fact). This row is NOT a re-pointing of D-09-53; it is a differently-scoped new item |
| D-10-29 | 10-CONTEXT.md (D-10-29) | OWN-09 | info | Phase 10 — plan 10-06 Task 1 (doc-comment and fixture-header corrections only, zero logic change) | DO NOT APPLY THE CODE CHANGE IN PHASE 10. A change that makes one fixture pass while breaking `TestCostCorpusLeafTemplatesDifferOnlyInTheCallee` — a mutation/differential-discipline test — is evidence the "fix" DEGRADES a detector, not evidence of a test-update inconvenience. Landing it would be a unilateral semantic change to an ownership checker, against the majority of existing test evidence, with no second detector watching. DELIVERABLE INSTEAD: correct `testdata/phase08/twin_b_accept.lang`'s header and `deriveFunctionUsesParam`'s doc comment to state the IMPLEMENTED contract (the exemption covers a zero-hop direct return of the parameter place, not an arbitrary identity chain), keep the observed-verdict assertions exactly as Phase 09 left them, and record D-10-28's re-filing |
| D-10-30 | 10-CONTEXT.md (D-10-30), interacts with D-10-60 | OWN-09 | info | Phase 10 — recorded here at planning time; nothing further to land | THE NO-THIRD-DEFERRAL RULE IS NOT VIOLATED BY D-10-27, and the distinction is written down so no future reader records it as "deferred twice." D-09-53 is not being deferred a second time: its PREMISE is WITHDRAWN AS INCORRECT, and its real content is re-filed as a differently-scoped item (D-10-28) with its own landing-phase cell. A withdrawn premise plus a new row is structurally different from a `Landing phase` cell rewritten to point past the phase it already named, which is what D-10-60 forbids. The planner re-ran D-09-53's own proposed patch as a pre-flight on 2026-09-11 and confirmed the 8 failures BEFORE writing the correction, exactly as D-10-30 requires — the reversal rests on a re-executed result, not on a planning document |
| D-10-31 | 10-CONTEXT.md (D-10-31), 10-RESEARCH.md SEM-09 row | SEM-09 | warning | Phase 10 — plan 10-05 states the finding in the shipped artifact; no deferred work | SEM-09's SECOND CLAUSE IS NARROW, NOT VACUOUS — stated plainly rather than left for a reviewer to discover. Enumerated against `core.go`: `OpReturn` (terminator, ordinary), `OpFail` (terminator, but an ordinary typed value propagated one boundary at a time — not unwinding), `OpDefect` (effectively total process-abort), the foreign process-root landing pad (`interp.go:369`, the one genuine nonlocal exit today, currently single-frame), and `OpCall` — which `core.go:559-564` documents as NEVER a terminator. Codename Lang has no exceptions and no unwinding. The only constructs that skip more than one frame are therefore the foreign landing pad extended to N frames, and the SEM-08 depth refusal. A plan that "satisfies" SEM-09's second clause with a test exercising only `OpReturn`/`OpFail` popping has proven the FIRST clause only |
| D-10-34 | 10-CONTEXT.md (D-10-34) | SEM-09, NAT-06 | warning | Phase 11 — `cgen`'s multi-frame nonlocal pad, IF it does not land in Phase 10 plan 10-05; the gap is named at that plan's own commit, never silently accepted | NAMED TEMPORARY SINGLE-AUTHORITY GAP. `interp` must not be the sole authority on cross-frame drop order. Two peers are supposed to know the order independently: the `corevalidate`-checked callee-signature invariant (D-10-33, landing in plan 10-05) and `cgen`'s multi-frame version of its existing single-frame `emitNonlocalPad`. If `cgen`'s multi-frame pad slips to Phase 11 — the expected outcome, since multi-function C emission is Phase 11's own charter (NAT-04 through NAT-07) — then between Phase 10's end and that landing, `interp` plus the `corevalidate` signature invariant are the only two authorities, and the third (the emitted-native one) is absent. That is a NAMED gap recorded here, not a silent acceptance |
| D-10-37 | 10-CONTEXT.md (D-10-37) | OWN-05b, TRU-03 | warning | Phase 10 — plan 10-01 writes the narrowed claim into the guard test's doc comment; recorded here so no artifact restates it unqualified | THE OWN-05b INDEPENDENCE CLAIM IS NARROWER THAN `check`/`corevalidate`'s, and must never stand unqualified. `interp` imports `corevalidate` (`interp.go:3-10`) and `Run`'s first act is `corevalidate.Validate(program)` (`interp.go:38`). This is NOT the mutual non-import independence `check` and `corevalidate` have from each other. The defensible claim is INDEPENDENCE OF DERIVATION MECHANISM FOR THE OWNERSHIP FACT SPECIFICALLY, NESTED INSIDE A SHARED, UNRELATED VALIDATION DEPENDENCY. A bug in `corevalidate.Validate` would feed `interp` bad input too — Knight and Leveson's correlated-fault result. The import is deliberately KEPT (hoisting `Validate` to the caller would weaken `interp`'s fail-closed posture and touch every call site for an import-graph purity the current one-value domain does not need); the mitigation is a guard test that fails the moment `interp` reads an ownership-bearing field of `corevalidate.Result` |
| D-10-40 | 10-CONTEXT.md (D-10-40) | OWN-05a, OWN-05b, QLT-04 | warning | Phase 10 — plan 10-08 documents criterion 4 this way in the shipped artifact; reopens only when the parameter-mode domain gains a second reachable value | THE ONE-VALUE DOMAIN MAKES NATURAL-INPUT AGREEMENT ON THE OWNERSHIP FACT NEAR-VACUOUS — say it rather than let the differential imply more than it proves. `ParameterContract.Mode` is hardcoded `"owned"` everywhere (`corevalidate.go:2163`, whose own comment cites D-07-01: today's grammar has exactly one parameter form), and `core.LinearOperation` has no override field. A natural-input four-way differential on this fact CANNOT DISAGREE BY CONSTRUCTION. What makes criterion 4 non-vacuous is the seeded-fault harness: a synthetic `core.Program` with `Mode` flipped to `"shared"`, proving (a) the closed-set decode check refuses it before any peer sees it, and (b) if that gate were bypassed, exactly which peer diverges. NOTE: 09-CONTEXT.md's citation of this mechanism at `corevalidate.go:2011-2015` is STALE; the mechanism is at `:2163` and is otherwise unchanged |
| D-10-59 | 10-CONTEXT.md (D-10-59), ROADMAP.md:198-200 (whose 2x trigger is scoped to Phases 08 and 09 ONLY) | all six phase requirements | info | Phase 10 — plan 10-06, the mandatory mid-phase gate, scheduled after the parallel wave and BEFORE criterion 4's differential work | DECLARED SCOPE-CUT TRIGGER, fixed in writing before planning finished. TRIGGER: if Phase 10's actual token cost summed across completed plans exceeds ~2x the initial per-plan token-estimate baseline for the same plans, measured at plan 10-06's gate. CUT ORDER: (1) QLT-04's declared composition depth reduced from 3 to 2, with the reduced depth stated VERBATIM in the shipped artifact and the remaining depth landing as named Phase 11 debt; then (2) the `OpForeignCall`-adjacent slice of TRU-02, deferred with a named Phase 11 fallback, mirroring D-09-21; then, only if still over budget, (3) OWN-05b's assertion folded into criterion 4's differential rather than standing alone, permitted ONLY after coverage is provably preserved. NEVER CUT: criterion 4's differential EXISTENCE; the `interp` stability freeze at full declared strength; the Pitfall-4 stack probe; D-09-51's `originvalidate` `OpCall` fix (it is Success Criterion 1 verbatim, not carried scope); and the seeded faults with their companion assertions (D-10-14, D-10-41, D-10-55). Any cut is a deferral with a named landing phase recorded here at the gate's own commit, never a silent drop. Phase 09's own trigger measured 0.17x (76,385 actual vs ~460,000 estimated) — a TOKEN-COST ratio, never a plan-count ratio |
| D-10-60 | 10-CONTEXT.md (D-10-60) | milestone-wide process rule | info | Milestone-wide — declared here, enforced by review; NOT enforced by the suite | NEW MILESTONE-WIDE DISCIPLINE, declared at Phase 10 planning. An item that would be deferred a SECOND time — any `D-NN-xx` row whose `Landing phase` cell would be rewritten to point past the phase it already named — must either be cut from the milestone explicitly via a REQUIREMENTS.md amendment, or becomes automatically never-cut. It cannot silently acquire a third landing phase. Verified at planning time that the suite will NOT catch this: `TestDebtRegistersAreWellFormed` (`session/session_test.go:2613-2671`) enforces the register's FORMAT ONLY — an `items:` count matching the table, one detail section per row, a closed severity vocabulary, and a non-empty landing phase — and does not track deferral hop count. This is therefore a declared rule enforced by review, and the gap in mechanical enforcement is stated rather than assumed closed |

## Detail

### D-10-19 — the `callgraph` import asymmetry between the two re-derivers

`originvalidate.go:18` imports `internal/compiler/callgraph` and calls
`callgraph.Order`. `corevalidate`'s own internal guard
(`corevalidate_endpoint_internal_test.go:107`) names `/compiler/callgraph` in its
forbidden list, per D-07-19. So the two packages Success Criterion 1 names as "the
two re-derivers" are held to different import standards today, for the same
utility package.

`callgraph`'s only compiler-internal dependency is `core`, so it cannot become a
route by which `originvalidate` acquires a transitive dependency on `check` or
`corevalidate` — which is what Criterion 1 actually forbids. The asymmetry is
therefore defensible. It is defensible **by review**, though, not by mechanism:
each package's forbidden list is hand-curated, and nothing in the build notices
that the two lists disagree.

Phase 10 hardens both guards from direct-import scanning to transitive
`go list -deps` checking (D-10-17) and adds `/compiler/corevalidate` to
`originvalidate`'s list (D-10-16), but **deliberately does not** add
`/compiler/callgraph`. Doing so would break the shipped build for a purity
argument the current dependency shape does not support.

### D-10-20 — the guard's residual bypassability

Each import guard lives inside the package it polices:
`originvalidate_test.go`'s two guards police `originvalidate`;
`pathoracle_test.go:51`'s guard polices `pathoracle`. A single commit can
therefore add a forbidden import and weaken the assertion that would have caught
it, and no other package's tests notice.

Transitive checking (D-10-17) raises the bar against the *realistic* failure mode
— a helper package that itself imports `check` — but does nothing about a
deliberate same-commit edit. Consolidation into one shared table would not help
either, and is rejected for a separate reason (D-10-18: the existing redundancy
across four near-identical guards means multiple deletions are needed to blind
the mechanism, and that redundancy is worth more than the maintainability win).

Closing this would require a control outside the policed package — a CI-owned
check, or an `internal/` restructure. Success Criterion 1 asks for a "build- or
**test**-level import control rather than convention," and a `go test` guard
already fails CI's `checks` job. The residual weakness is accepted and stated.

### D-10-27 — D-09-53's diagnosis is falsified

PHASE-09-DEBT.md's D-09-53 records `deriveFunctionUsesParam`'s
`default: usesParam = true` arm as a defect contradicting the function's own doc
comment ("true UNLESS the entire parameter-derived reference chain leads directly
to its own terminating OpReturn"), and states that
`TestInterproceduralLivenessTwinPatternB` is "UNAFFECTED."

The planner applied the register's own literal proposed narrowing — excluding
`core.OpMove`/`core.OpCopy` from the default arm at `check.go:729-730` — in a
throwaway patch and ran the full suite on 2026-09-11. The result contradicts the
register on both points. Eight named `check` tests fail:

- `TestDeriveFunctionUsesParamBehaviors` (subtest: a leaf function whose body reads its parameter)
- `TestInterproceduralLivenessTwinPatternB` (subtest: callee uses its parameter) — **the very test the register cites as unaffected**
- `TestInterproceduralLivenessTwinPatternBRealFixtures` (subtest: `twin_b_accept.lang`)
- `TestSummaryDerivationTwoHopChainPropagates`
- `TestSummaryDerivationIsOnePassPerFunction`
- `TestSummaryDerivationRequiresProgramOrder`
- `TestCostCorpusLeafTemplatesDifferOnlyInTheCallee` (subtest: callee uses its parameter)
- `TestInterproceduralDiagnosticOrderingStability`

plus `TestNoUndeclaredCheckPeerDivergenceAcrossCorpus` in package `session` — a
newly-surfaced `check`/`corevalidate` divergence, which is the more serious
signal: the narrowing does not merely move a verdict, it makes two peers disagree
where they currently agree.

Six independently-authored, currently-passing Phase 08/09 tests therefore **pin
the current semantics as intended**. The structural claim in the register is also
overstated: `check.go:723-726` exempts only `case core.OpReturn:` — the operation
itself, never a chain leading to it — so the doc comment describes a contract the
code has never implemented.

**Conclusion:** the derivation is not the defect. The doc comment's wording and
`twin_b_accept.lang`'s header comment are.

### D-10-28 — the real gap: `corevalidate` has no interprocedural `usesParam` peer

Even with D-09-53's narrowing applied, Pattern B's split still cannot be
demonstrated end to end. `corevalidate_peer_liveness.go:38-39` defines
`peerLoanCarryFact` with exactly one field, `ReturnsBorrowOfParam` — Pattern A's
fact, the one D-09-51 scoped. There is no `usesParam` field, no `usesParam`
derivation, and no `usesParam` consult anywhere in `corevalidate`. Its liveness
walk therefore treats every `OpCall` argument unconditionally as a use, and
refuses the accept twin regardless of what `check` derives.

This is a **missing second detector**, not a broken existing one — Key Lesson 2's
exact exposure, where exactly one peer watches a fact and nothing catches it
drifting. It is filed as its own row, with its own landing phase, precisely so
that it is not mistaken for a re-pointing of D-09-53.

**Landing phase deliberately OPEN.** 10-CONTEXT.md grants the planner discretion
here and 10-RESEARCH.md recommends leaving it open rather than assigning it under
time pressure. Candidate homes: Phase 11 (which already opens `check` for native
lowering) or M003.

### D-10-29 — the deliverable is an artifact correction, not a code change

The narrowing breaks `TestCostCorpusLeafTemplatesDifferOnlyInTheCallee`, a
mutation/differential-discipline test whose whole purpose is to detect a detector
degrading. A change that makes one fixture pass while breaking that test is
evidence of degradation, not of a test-update inconvenience.

What Phase 10 ships instead, in plan 10-06 Task 1, with **zero logic change**:

1. `deriveFunctionUsesParam`'s doc comment is corrected to state the implemented
   contract — the exemption covers a zero-hop direct return of the parameter
   place, never an arbitrary identity chain that happens to terminate in a
   return.
2. `testdata/phase08/twin_b_accept.lang`'s header comment is corrected to state
   the verdict the fixture actually receives and why, replacing the intended-split
   narrative.
3. The observed-verdict assertions Phase 09 left in place stay byte-identical.
4. D-10-28's re-filing is cross-referenced from both corrected comments.

### D-10-30 — why this is a reversal, not a second deferral

D-10-60 forbids an item silently acquiring a third landing phase. D-09-53 is not
acquiring one. Its premise — "`deriveFunctionUsesParam`'s default arm is a defect"
— is **withdrawn as incorrect** on re-executed evidence, and separately, a
different gap that the register's prose gestured at is filed fresh as D-10-28 with
its own landing-phase cell.

The distinction matters and is recorded here rather than left implicit: a
withdrawn premise closes a row; a rewritten `Landing phase` cell defers one. Only
the second is what D-10-60 governs.

D-10-30 also imposed a procedural obligation on the planner: re-run the throwaway
patch **before** writing the correction, so the reversal rests on a re-executed
result rather than on this document. That was done on 2026-09-11; the result is
transcribed in D-10-27 and at the head of this register, and the patch was
reverted (`git diff internal/compiler/check/check.go` empty).

### D-10-31 — SEM-09's second clause is narrow, and that must be said

SEM-09 requires drops to run in the defined order "on normal return **and on every
nonlocal exit** across a call boundary." Enumerated against `core.go`, Codename
Lang's nonlocal-exit surface is:

| Construct | Terminator? | Skips more than one frame? |
|---|---|---|
| `OpReturn` | yes | no |
| `OpFail` | yes | no — an ordinary typed value propagated one boundary at a time |
| `OpDefect` | yes | effectively a total process abort |
| foreign process-root landing pad (`interp.go:369`) | n/a | **yes** — the one genuine nonlocal exit today, currently single-frame |
| `OpCall` | **no** (`core.go:559-564`) | n/a |

The language has no exceptions and no unwinding. So SEM-09's second clause is
*narrow* — but not vacuous: the two constructs that genuinely skip more than one
frame are the foreign landing pad extended to N frames, and the SEM-08 depth
refusal. A plan that satisfies the clause with a test exercising only
`OpReturn`/`OpFail` popping has proven the **first** clause only, and plan 10-05
must not do that.

### D-10-34 — the temporary single-authority gap on cross-frame drop order

D-10-34 requires that `interp` not be the sole authority on drop order across
frames. Two independent peers are supposed to know it:

1. The `corevalidate`-checked callee-signature invariant (D-10-33) — a function's
   frame must be drained of live resources before it pops, checked once per
   function declaration, never re-derived per call site. **This lands in plan
   10-05.**
2. `cgen`'s multi-frame version of its existing single-frame `emitNonlocalPad`.

Item 2 is expected to slip to Phase 11, because multi-function C emission is
Phase 11's own charter (NAT-04 through NAT-07) and `cgen` today refuses
`len(Functions) != 1` outright. Between Phase 10's end and that landing, the
emitted-native authority on drop order is **absent**, and only `interp` plus the
`corevalidate` invariant know the order.

That is recorded here as a named gap at planning time. If plan 10-05 finds
`cgen`'s pad is cheaply extendable, landing it there closes this row early; if
not, this row carries forward into Phase 11 with its landing phase already named.

### D-10-37 — the narrowed OWN-05b independence claim

`interp.go:3-10` imports `corevalidate`, and `Run`'s first act (`interp.go:38`) is
`corevalidate.Validate(program)`. So OWN-05b's "third independent deriver" claim
is structurally different from the OWN-05a claim `check` and `corevalidate` earned
from each other, which rests on mutual non-import.

The honest claim: **independence of derivation mechanism for the ownership fact
specifically, nested inside a shared, unrelated validation dependency.** A bug in
`corevalidate.Validate` feeds `interp` bad input too — Knight and Leveson's
correlated-fault result applies and is not being waved away.

The import is deliberately **kept**. Hoisting `Validate` to `Run`'s callers would
weaken `interp`'s fail-closed posture (it would become possible to run an
unvalidated program) and would touch every call site, buying an import-graph
purity the current one-value domain does not need.

The mitigation is mechanical rather than rhetorical: a guard test that fails the
moment `interp` reads an ownership-bearing field of `corevalidate.Result`
(`PeerSignatures()`, `PeerSiteCoverage()`). Today `interp` reads none of them —
grepped, zero hits — but that is true by omission. The guard converts "happens not
to" into "cannot without a red test." Plan 10-01 lands it and writes this narrowed
claim into its doc comment.

### D-10-40 — the one-value domain, and why criterion 4 is billed with a seeded fault

`ParameterContract.Mode` is hardcoded `"owned"` at `corevalidate.go:2163`, whose
own inline comment cites D-07-01 ("today's grammar has exactly one parameter
form"). `core.LinearOperation` carries no call-site override field, and arity is
fixed at 1.

Therefore a natural-input differential on the call-site ownership-transfer fact
**cannot disagree by construction**. Reporting "three engines agree" over such a
corpus would be an overclaim of exactly the kind D-09-37 already caught once.

What makes criterion 4 non-vacuous on this fact is the seeded-fault harness: a
synthetic `core.Program` with `Mode` flipped to `"shared"`, proving (a) the
closed-set decode check refuses it before any peer sees it, and (b) if that gate
were bypassed, exactly which peer diverges. Plan 10-08 documents criterion 4 this
way in the shipped artifact.

**Stale citation corrected:** 09-CONTEXT.md cites this mechanism at
`corevalidate.go:2011-2015`. It is at `:2163`. The mechanism and its comment are
otherwise unchanged.

### D-10-59 — the declared scope-cut trigger

ROADMAP.md:198-200 declares the milestone 2x trigger for **Phases 08 and 09
only**. Phase 10 inherits none and therefore declares its own, in writing, before
planning finished, per Key Lesson 4.

**Trigger.** If Phase 10's actual token cost, summed across completed plans,
exceeds ~2x the initial per-plan token-estimate baseline for the same plans. The
baseline is each PLAN.md's own `estimate.tokens` frontmatter value. Measured at
the **mandatory mid-phase gate**, plan 10-06 — scheduled after the
`originvalidate` ∥ `pathoracle` ∥ `interp`-call-stack parallel wave lands and
**before** criterion 4's differential work begins.

**Cut order.**

1. QLT-04's declared composition depth reduced from 3 to 2, with the reduced
   depth stated **verbatim** in the shipped artifact (Criterion 4's own "stated
   rather than implied" requirement), remaining depth landing as named Phase 11
   debt.
2. The `OpForeignCall`-adjacent slice of TRU-02, deferred with a named Phase 11
   fallback, mirroring D-09-21.
3. Only if still over budget: OWN-05b's assertion folded into criterion 4's
   differential rather than standing alone — permitted **only** after coverage is
   provably preserved.

**Never cut.** Criterion 4's four-way differential **existence** (cutting its
depth is a scope cut; cutting its existence makes the phase's own gate
decorative); the `interp` stability freeze at full declared strength (disqualified
on Phase-11-dependency grounds alone); the Pitfall-4 stack probe (a named Gate in
the roadmap's own criterion-2 text); D-09-51's `originvalidate` `OpCall` fix (it
is Success Criterion 1 verbatim, not carried scope); and the seeded faults with
their companion assertions (D-10-14, D-10-41, D-10-55).

Any cut is a deferral with a named landing phase recorded in this register at the
gate's own commit, never a silent drop.

**Calibration note.** Phase 09's trigger was genuinely checked mid-phase (plan
09-08's agenda item (f), `09-08-PLAN.md:177`) and measured **0.17x** — 76,385
actual against ~460,000 estimated (`09-08-SUMMARY.md:151`). That is a
**token-cost** ratio. Phase 09 ran 10 plans against an initial estimate of fewer,
so a naive plan-count reading would have misfired. Plan 10-06 must measure tokens.

### D-10-60 — the no-third-deferral rule

Declared here as a new milestone-wide discipline.

An item that would be deferred a **second** time — any `D-NN-xx` row whose
`Landing phase` cell would be rewritten to point past the phase it already named —
must either be cut from the milestone explicitly via a REQUIREMENTS.md amendment,
or becomes automatically never-cut. It cannot silently acquire a third landing
phase.

Verified at planning time that the suite will not enforce this:
`TestDebtRegistersAreWellFormed` (`session/session_test.go:2613-2671`) checks the
register's **format only** — the `items:` count matching the table, one `### <ID>`
detail section per row, a severity from the closed vocabulary, and a non-empty
landing phase. It does not read landing-phase history and cannot count deferral
hops.

This is therefore a **declared rule enforced by review**, and that limitation is
stated rather than assumed closed. D-10-30 records why D-10-27's reversal is not
itself a violation of it.
