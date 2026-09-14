# Phase 11 Mid-Phase Gate Adjudication

**Plan:** 11-04
**Gate function:** `session.VerifyPhase11ZeroAttributeGate` (`internal/compiler/session/session_phase11_gate.go`)
**Corpus:** `testdata/phase11/multi_function_gate_corpus.lang`
**Date:** 2026-09-12

This is the mandatory mid-phase gate (D-11-14). Nothing in waves 4, 5, or 6
is admitted until this document is ratified.

## 1. Corpus size and file paths

- 1 fixture file: `testdata/phase11/multi_function_gate_corpus.lang`.
- Two additional inline, non-committed corpora exist only inside
  `internal/compiler/session/session_phase11_gate_test.go` as Go string
  literals, used exclusively to exercise the N=0 and N=2 boundaries
  (`phase11ZeroWouldCarryCorpus`, `phase11TwoWouldCarryCorpus`); they are
  not part of the gate's own committed corpus and are not counted below.

## 2. Function count and call-edge count of the gate corpus

- Function count: **3** (`identity`, `main`, `touch`).
- Call-edge count: **1** (`main` &rarr; `identity`, one `core.OpCall`
  operation).
- Both floors met: function count &ge; 2, call-edge count &ge; 1
  (`TestPhase11GateIsNonVacuous`, `TestPhase11ZeroAttributeGate`).

## 3. N — the would-have-carried-restrict count

- **N = 1** — `touch` alone. `touch`'s body is the exact
  exclusive-borrow-then-reborrow-to-terminator chain
  `testdata/phase5/restrict_borrow.lang` uses on its own Buffer parameter;
  neither `main` nor `identity` (Byte parameters, no borrow chain) match.
- Derivation used: `phase11WouldCarryRestrict` (`session_phase11_gate.go`)
  — session's OWN re-derivation of the identical structural condition
  `cgen.SelectsByPointerLowering`/`selectsByPointerLowering` decides,
  reading only `core.Function`/`core.LinearBody`. It shares zero helpers
  with `cgen` (D-12); `session_phase11_gate.go` calls into `cgen` only for
  emission (`cgen.EmitNative`) and scanning
  (`cgen.ScanForBannedAttributes`), never for this count.
- Both sides of the N-threshold are asserted, not only the passing one:
  `TestPhase11GateFailsAtNZero` (N=0, the gate FAILS) and
  `TestPhase11GateCountsAdjacentWouldCarryFunctions` (N=2, on a distinct
  corpus with two would-carry functions — the gate still passes and N is
  counted as 2, not merged or double-counted).

## 4. Artifacts scanned by `cgen.ScanForBannedAttributes` and its result

- **2 artifacts scanned**, under the corpus's default `AttributesSuppressed`
  invocation:
  1. The whole-program C17 translation unit `cgen.EmitNative(program)`
     produces (`emitProgram`'s own output — every function's own body,
     `main`, and the generated empty call-boundary attribute-set comment).
  2. One counterfactual single-function artifact,
     `cgen.EmitNative(singleFunctionProgram)` for `touch` alone, extracted
     from the checked+validated corpus program and emitted under the
     gate's own active `AttributeSuppressionProfile`.
- **Result: empty.** `cgen.ScanForBannedAttributes` returns no banned
  token across either artifact (`TestPhase11ZeroAttributeGate`,
  `report.BannedAttributesFound` has length 0).
- Consistency check only (D-11-15, not the second knower): the
  whole-program call boundary's own `[]cgen.EmittedAttribute` value is
  definitionally empty — `emitCall` (`cgen_program.go`) contains no
  attribute-recording statement at all — reported as
  `ManifestEmptyAttributes: true`, explicitly not treated as an
  independent knower.

## 5. `-O0` vs interpreter agreement result (five-axis comparator)

- **Pass.** `session.RunNative` (called unconditionally, never gated on
  the scan above) reports no diagnostics and no `EngineMismatch` for the
  corpus at both `-O0` and `-O3`; `report.InterpreterAgreesAtO0 == true`.
  `TestPhase11ZeroAttributeGate`, `TestPhase11GateCountsAdjacentWouldCarryFunctions`
  both assert this directly.

## 6. Mutation-kill result under the justified profile

- **Went RED, as required** — the mutation-kill's own re-run.
  `TestPhase11GateMutationKill` re-runs the
  identical lane on the identical corpus with
  `profile = cgen.AttributesJustified` instead of
  `cgen.AttributesSuppressed`. The counterfactual `touch` artifact then
  carries the literal `restrict` token
  (`static LANG_BUFFER LANG_TOUCH(LANG_BUFFER *restrict LANG_BUFFER_LANG_PARAMETER_0)`),
  `cgen.ScanForBannedAttributes` catches it, `report.BannedAttributesFound`
  is non-empty, and `report.Passed()` is `false`. A gate green under both
  profiles would have been measuring nothing (D-11-18); this one is not.

## 7. Diff-locality result

- **Pass.** `TestPhase11SuppressionIsDiffLocal` emits `touch` alone under
  both profiles and line-diffs the two outputs after stripping every
  `cgen.BannedOptimizerAttributes` token: the ONLY divergent line is the
  function signature's own qualifier
  (`*LANG_BUFFER_LANG_PARAMETER_0` vs. `*restrict LANG_BUFFER_LANG_PARAMETER_0`),
  and after stripping the banned token the two lines are structurally
  identical. `cgen.NoreturnExemption` (`_Noreturn`) is absent from both
  outputs symmetrically (this fixture has no defect terminator), so its
  presence never diverges between profiles.

## Adjudication

Per D-11-09, multi-function C emission's own call-boundary attribute set
is proven independently non-vacuous and empty: a multi-function program
containing a genuine would-have-carried-restrict function compiles,
links, runs, and agrees with the interpreter at `-O0` (and `-O3`), with
**zero** alias or capture attributes emitted at any Lang-to-Lang call
boundary. The gate's own scan is proven not vacuous by a real mutation
kill (item 6) and a diff-locality check (item 7) — not by inspection
alone.

**This zero-attribute state is TERMINAL for Phase 11**, not a waypoint.
No call lowering with alias attributes is admitted this phase. NAT-05
ships satisfied by this explicit, generated, empty call-boundary
attribute set (Task 1, `cgen_program.go`'s
`emitCallBoundaryAttributeComment`) — this **is a requirement weakened by
evidence** (D-11-10), not a clean pass as originally scoped, and must be
written up as such in `11-VERIFICATION.md`.

### Known divergence, flagged for human review (not silently resolved)

`go run ./cmd/lang --json check testdata/phase11/multi_function_gate_corpus.lang`
reports `status: invalid` with `core.origin_omitted` for `touch`. This is
**not** a defect this plan introduced: it is a pre-existing incompatibility
between two independently-shipped invariants that this plan's own corpus
is the first to combine —

- D-05-02's `selectsByPointerLowering` (and this plan's own
  `phase11WouldCarryRestrict` re-derivation) REQUIRES `PublicOrigin == nil`
  (an undeclared origin) as part of its own structural gate.
- `originvalidate.ValidatePublished` (consulted only by the CLI `check`
  command's `publishedOriginProblemFile`, never by `session.Check` or any
  existing Phase 5/6+ production or test path) unconditionally flags ANY
  function — exported or not — whose body derives a borrow-based return
  origin without a declared `PublicOrigin`.

Any function matching `selectsByPointerLowering`'s structural shape
necessarily derives such an origin (that is what the shape means), so this
divergence is not specific to `touch`: running
`testdata/phase5/restrict_borrow.lang` — the already-shipped, existing
fixture `touch`'s body is modeled on verbatim — through the identical CLI
`check` command reproduces the same `core.origin_omitted` diagnostic. No
prior test or production path exercises `restrict_borrow.lang`-shaped
functions through the CLI `check` command; this plan's own `<verify>`
block is the first place that combination is asked for. All of this
plan's own verification instead runs through `session.Check` +
`corevalidate.Validate` — the same admission path `VerifyPhase11ZeroAttributeGate`,
every `cgen_program_test.go`/`session_phase11_gate_test.go` test, and
every other production run site in this codebase already use — where the
corpus checks and validates cleanly (`TestPhase11ZeroAttributeGate` and
siblings all pass). Flagged here for human review rather than worked
around silently; it does not affect the gate's own verdict above, which
never depends on the CLI `check` command.

## Summary of concrete values (D-11-19)

- function count: 3
- call-edge count: 1
- N = 1 (would-have-carried-restrict functions)
- artifacts scanned: 2 (whole-program TU + 1 counterfactual)
- mutation-kill: red under the justified profile, as required

## Ratification

**Selected option: A — Gate PASSED.**

All four conjuncts met on a provably non-vacuous corpus (both sides of the
N-threshold asserted), non-vacuity floors met, mutation kill red under the
justified profile, diff-locality confirmed. Waves 4-6 (NAT-06's
differential lanes, QLT-03's register, QLT-06's cache fix, QLT-05's
reducer) are **admitted**. The zero-attribute state is recorded as
TERMINAL for this phase per D-11-09; NAT-05 is to be written up in
`11-VERIFICATION.md` as a requirement weakened by evidence (D-11-10), not
a clean pass.

Recorded automatically under auto-mode execution (no `gate="blocking-human"`
attribute on this task) based on the concrete, reproducible conjunct
values in items 1-7 above; every value was measured by the test suite
in this same commit, not asserted from narrative. A human reviewer should
still confirm this ratification and the flagged CLI-check divergence
before treating `touch`'s exact fixture shape (and the CLI check
command's own origin-publication strictness) as fully settled — consistent
with 11-03-SUMMARY.md's own precedent of flagging structural deviations
for human review rather than treating executor-recorded ratification as
final for a `one-way` reversibility decision.

**Date:** 2026-09-12
