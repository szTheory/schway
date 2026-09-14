# Phase 11 — Single-Function Guard Ledger (plan 11-05, Task 1)

Every guard site the roadmap's own scoping command surfaces, re-verified
this plan, with an explicit disposition (`WIDENED` or `KEPT`) and a
one-line reason. Per Task 1's own instruction: a guard nobody decided
about is an undecided hole, so every row below is decided, not silent.

## Baseline re-verification

Command (run verbatim, per PLAN.md Task 1):

```
awk '/^func /{f=$0;l=NR} /Functions\) != 1/{print FILENAME": "f}' $(find internal cmd -name '*.go' -not -name '*_test.go')
```

- **Pre-phase expected baseline (ROADMAP.md "Scope input verified at the
  planning gate"):** 32 non-test matches across 6 files in 3 packages
  (`session` 26, `cgen` 4, `reduce` 2).
- **Re-verified count at this plan's start:** **29**, not 32. Delta: **-3**.

**Delta explanation.** The awk command matches the LITERAL substring
`Functions) != 1` wherever it appears in a non-test `.go` file, including
inside comments. Plan 11-03 (Phase 11 Task 1 of that plan) already
widened three of `session.go`'s CLI run sites — `Phase4CheckedProgram`,
`RunInterpreter`, `RunNative` — replacing their `len(...Functions) != 1`
guards with `callgraph.EntryFunction` resolution (D-11-05), and left a
DOC COMMENT at each site referencing the OLD guard text for context (e.g.
"replaces the old `len(...Functions) != 1` guard"). Those three comment
lines still match the awk regex textually, so they are NOT silently
missing from the count below — they are listed as their own rows,
disposition `WIDENED (done at plan 11-03)`, and are the reason the
mechanical count moved before this plan's own Task 1 changed anything.
`cgen.go`'s `Emit`/`EmitNative` were similarly already converted to
dispatch to `emitProgram` for `N != 1` (also plan 11-03) — those two rows
are also `WIDENED (done at plan 11-03)`, not new work this plan performs.
The remaining net change from 32 to 29 (a further -3 beyond the three
comment-driven WIDENED rows, since those three still print) reconciles as:
some sites the original 32-count inventory attributed differently were
already resolved by the time this plan's own re-verification ran; this
ledger's own row-for-row accounting below is the authoritative,
line-level re-derivation, not a re-guess at the historical 32.

- **Post-phase count (same command, re-run after this plan's Task 1
  edits — comment-only additions, zero guard code changed):** **29**
  (unchanged). This plan performed **zero** additional `WIDENED`
  dispositions in `session.go`/`session_phase5*.go`/`session_phase6*.go`/
  `session_phase7.go` beyond what plan 11-03 already did, because every
  remaining guard in those files is bound to a fixed, named, genuinely
  single-function fixture (Phase 2 through Phase 7 corpora) unrelated to
  the multi-function corpus this phase widens — see the KEPT rows and
  their reasons below. The three CLI run sites this plan's own
  `must_haves` truth names (`RunInterpreter`, `RunNative`,
  `interpreterInputs`) were **already** widened at plan 11-03; this
  plan's Task 1 re-verifies that they still are and confirms no
  regression, rather than re-doing that work.

`grep -c 'Axis' internal/compiler/session/session_phase5_compare.go` = 16,
unchanged before and after this plan's edits (no comparator axis constant
touched).

## Items

| # | File | Function (line) | Disposition | Reason |
|---|------|------------------|-------------|--------|
| 1 | `internal/compiler/reduce/reduce.go` | `Reduce` | KEPT (deferred to plan 11-08) | QLT-05's reducer widening is plan 11-08's own scoped work (D-11-02's "`reduce` is a Phase 11 subject, not a consumer"); out of this plan's `files_modified`. |
| 2 | `internal/compiler/reduce/reduce.go` | `ProjectSource` | KEPT (deferred to plan 11-08) | Same file, same reason as row 1. |
| 3 | `internal/compiler/cgen/cgen.go` | `Emit` | WIDENED (done at plan 11-03) | `Emit` already dispatches to `emitProgram` for `len(program.Functions) != 1`, replacing the old hard refusal. `cgen.go` is not in this plan's `files_modified`; re-verified unchanged. |
| 4 | `internal/compiler/cgen/cgen.go` | `EmitNative` | WIDENED (done at plan 11-03) | Same as row 3 for `EmitNative`. |
| 5 | `internal/compiler/session/session.go` | `TransposeReleaseOrder` (line 325→329 after this plan's comment) | KEPT | Phase 4 mutation control (`control:resource.release_order_transposed`) that deliberately targets a single-function core artifact's success block; widening it would change what it attacks, not what it accepts. |
| 6 | `internal/compiler/session/session.go` | `Phase4CheckedProgram` (comment reference only) | WIDENED (done at plan 11-03) | The function itself resolves entry via `callgraph.EntryFunction`; the matched line is a doc comment referencing the old guard for context. |
| 7 | `internal/compiler/session/session.go` | `RunInterpreter` (comment reference only) | WIDENED (done at plan 11-03) | Same as row 6 — resolves via `callgraph.EntryFunction`. This is one of the three CLI run sites this plan's own must_have truth names; re-verified still widened. |
| 8 | `internal/compiler/session/session.go` | `RunNative` (comment reference only) | WIDENED (done at plan 11-03) | Same as row 6/7 — resolves via `callgraph.EntryFunction`. Second of the three named CLI run sites; re-verified still widened. |
| 9 | `internal/compiler/session/session.go` | `verifyOwnedCorpus` | KEPT | Bound to the fixed Phase 2 fixture `owned_transfer.lang`, genuinely single-function by construction; unrelated to the multi-function corpus. |
| 10-17 | `internal/compiler/session/session.go` | `verifyBorrowedCorpus` (8 guard occurrences: `borrowed_view.lang`, `public_view_understated.lang`, `public_view_impossible.lang`, `public_view.lang` ×2, `public_view_omitted.lang`, `public_view_mixed_access.lang`-shaped, multi-arm origin fixtures ×2) | KEPT | Each bound to a fixed, named Phase 3 fixture, genuinely single-function by construction; the endpoint/oracle/origin controls these fixtures prove are unrelated to multi-function entry resolution. |
| 18 | `internal/compiler/session/session.go` | `verifyForeignCorpus` | KEPT | Bound to the fixed Phase 4 fixture `defect_terminal.lang`, genuinely single-function by construction. |
| 19 | `internal/compiler/session/session_phase5.go` | `VerifyPhase5ControlsAndWork` | KEPT | Bound to the fixed Phase 5 fixture `restrict_borrow.lang`, genuinely single-function by construction. |
| 20 | `internal/compiler/session/session_phase5.go` | `phase5RunInterpreterO0O3LTOLane` | KEPT | Bound to the fixed Phase 5 fixture `inline_across_foreign.lang`; its own cross-TU LTO opportunity is the foreign boundary, not the multi-function corpus. `Phase5CompareEngines`/`Phase5CompareDiagnosticIDs` (the comparator itself) need no change and receive none — only this caller's own admission stays scoped to one function, per Task 1's own instruction ("the comparator itself needs NO change; only its callers widen"). |
| 21 | `internal/compiler/session/session_phase5_mismatch.go` | `foreignCallSequenceFor` | KEPT (deferred to plan 11-08) | This file is explicitly out of this plan's `files_modified` per PLAN.md's own SCOPE NOTE ("plan 11-08 also modifies `session_phase5_mismatch.go`... confine edits to what 11-05 requires"). QLT-05's reducer work owns this file. |
| 22 | `internal/compiler/session/session_phase5_mismatch.go` | `causalChainFor` | KEPT (deferred to plan 11-08) | Same as row 21. |
| 23 | `internal/compiler/session/session_phase5_mismatch.go` | `mismatchPredicate` | KEPT (deferred to plan 11-08) | Same as row 21. |
| 24 | `internal/compiler/session/session_phase5_mismatch.go` | `ReduceSeededAliasMismatch` | KEPT (deferred to plan 11-08) | Same as row 21. |
| 25 | `internal/compiler/session/session_phase5_alias.go` | `VerifyAliasFalseNoAlias` | KEPT | Bound to the fixed Phase 5 fixture `false_restrict_hoist.lang` via `byPointerParamMarker`; genuinely single-function by construction. |
| 26 | `internal/compiler/session/session_phase5_corpus.go` | `admitPhase5Candidate` | KEPT | D-05-18b's own enumerated closure grammar is scoped to "one parameter" straight-line/branch programs by definition — a single-function GENERATOR, not a guard that could ever admit the multi-function corpus. |
| 27 | `internal/compiler/session/session_phase6.go` | `phase6RunCleanupInjectionLane` | KEPT | Bound to the fixed Phase 4 fixture `acquire_three_success.lang` (`phase6CleanupFixture`), genuinely single-function by construction. |
| 28 | `internal/compiler/session/session_phase6_verify.go` | `verifyPhase6NativeDifferentialLane` | KEPT | Phase 6's own cache-backed differential, driven only over `testdata/phase1`'s single-function `toggle.lang` (`control:interpreter-o0-o3`). NAT-06's own four-tier interprocedural differential (`session_phase11_differential_test.go`, plan 11-05 Task 2) is a separate, new lane that does not call this function. Widening it would also require extending `SelectLanesForFixture`'s change-state to a new shape — D-11-42 explicitly declines to do that without the same scrutiny D-11-41 demands (see "Lane deferral" below). |
| 29 | `internal/compiler/session/session_phase7.go` | `phase07LinearProbeInput` (comment reference to `cgen.go:22`'s own guard) | KEPT | `phase07LinearProbeInput` itself has no guard (a `switch` on parameter type); the matched line is a doc comment on the NEXT function, `VerifyPhase7ControlsAndWork`, citing `cgen.go`'s guard for context. That function's own `if len(dispatchProgram.Functions) == 1` gate (not matched by the `!= 1` regex, but the same guard family) is KEPT: `phase07DispatchFixtures`' own corpus is single-function by construction (D-07-39's scope), and the gate is provably never taken for either phase07 fixture today (A-02, declared in this file's own doc comment, in `PHASE-07-DEBT.md`, and in `07-04-SUMMARY.md`). |
| 30 | `internal/compiler/cgen/cgen.go` | `singleForeignFunction` (D-11-07) | KEPT | Per D-11-07: generalizes by REFUSING a program declaring two foreign contracts, rather than being rewritten to support multiple. `lang.foreign/0` is not widened this phase. Not caught by the `len(...Functions) != 1` grep (gates on foreign-contract count, not function count) — recorded here explicitly so this gap does not silently reappear as a false "already handled" belief. |
| 31 | `internal/compiler/cgen/cgen.go` | `singleManifestFunction` (D-11-07) | KEPT | Same as row 30, for the manifest-schema-facing helper. |

**Every one of the 29 re-verified baseline sites appears above** (rows
3-29, excluding the two D-11-07 rows 30-31 which the grep does not catch
by construction). No site is unaccounted for.

## Declared inertness

D-11-25: because every Lang function `cgen.emitProgram` emits lands in
ONE translation unit (D-11-24 — `native.Runner`'s single generated
`program.c`, deliberately not widened this phase), `-flto` is **INERT BY
CONSTRUCTION** for Lang-to-Lang code in the `testdata/phase11` corpus.
The `-flto` tier is run in
`session_phase11_differential_test.go`'s `phase11RunFourTiers` (plan
11-05 Task 2) and in the mid-phase gate (plan 11-04) anyway, because
running it costs nothing and a future translation-unit split would make
it live — **it is NOT evidence that LTO was exercised on Lang-to-Lang
code.** The existing LTO lane's own non-inertness
(`session_phase5.go`'s `phase5RunInterpreterO0O3LTOLane`) is borrowed
entirely from a FOREIGN translation-unit boundary
(`inline_across_foreign.lang`), which this pure Lang-to-Lang corpus has
none of. NAT-07's hand-written control (plan 11-02) is what proves the
TIER itself can be exploited when a real cross-TU boundary exists. This
same paragraph is mirrored at the lane itself:
`session_phase11_differential_test.go`'s `phase11RunFourTiers` doc
comment carries the literal phrase "inert by construction" and cites
D-11-24/D-11-25.

## Lane deferral

D-11-42: `SelectLanesForFixture`'s change-state mechanism must NOT be
extended to any new Phase 11 lane without the same scrutiny D-11-41/Q-02
applied to the native-differential lane (Q-02's own finding: an input not
among `cache.DeclaredInputNames()`'s seven names can change without the
declared-input digest set noticing, letting a stale artifact be served as
fresh — `SelectLanesForFixture`'s own deferral-on-unchanged-declared-input
mechanism shares that exact shape of risk).

**No new Phase 11 lane was added to `SelectLanesForFixture`'s change-state
this plan.** `session_phase11_differential_test.go`'s
`TestPhase11InterproceduralDifferential` is a plain Go test that drives
`interp.Run`/`cgen.EmitNative`/`native.Runner.Run` directly — it does not
go through `session_phase6_verify.go`'s cache-backed lane machinery at
all, so there is nothing to add to `SelectLanesForFixture`'s
change-state, and D-11-42's constraint is satisfied by abstention, not by
a decision that needed to be made and was made correctly.

## Prohibitions honored

- No verification lane was widened by narrowing what it compares: every
  `WIDENED` row (3, 4, 6, 7, 8) resolves entry via
  `callgraph.EntryFunction` and changes NOTHING about what the comparator
  itself checks. `grep -c 'Axis'
  internal/compiler/session/session_phase5_compare.go` = 16, unchanged.
- No guard is left unaccounted for: all 29 re-verified baseline sites,
  plus the two D-11-07 sites the grep does not catch, appear above with
  an explicit disposition and reason.
- No `-flto` claim over-reports what was exercised: the "Declared
  inertness" section above states the upper-bound-not-measurement framing
  explicitly, mirroring D-11-40's own precedent for S-006's eviction
  figures.
