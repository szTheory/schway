---
phase: 12-result-payloads
reviewed: 2026-09-13T00:00:00Z
depth: standard
files_reviewed: 32
files_reviewed_list:
  - internal/compiler/ast/ast.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_n1_convergence_test.go
  - internal/compiler/cgen/cgen_payload_tracer_test.go
  - internal/compiler/check/check.go
  - internal/compiler/check/check_payload_test.go
  - internal/compiler/core/core.go
  - internal/compiler/core/core_convention_absence_test.go
  - internal/compiler/core/core_single_derivation_test.go
  - internal/compiler/core/core_test.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/corevalidate/corevalidate_alternative_details_test.go
  - internal/compiler/corevalidate/corevalidate_result_payload_probe_test.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interp_oracle_golden_test.go
  - internal/compiler/interp/interp_test.go
  - internal/compiler/native/native.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/originvalidate/originvalidate_payload_test.go
  - internal/compiler/pathoracle/pathoracle_payload_test.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_payload_control_test.go
  - internal/compiler/session/session_payload_replay_test.go
  - internal/compiler/syntax/parser.go
  - testdata/phase08/relay_depth2_accept.lang
  - testdata/phase12/payload_arity_mismatch.lang
  - testdata/phase12/payload_binder_on_nullary.lang
  - testdata/phase12/payload_borrow_interaction.lang
  - testdata/phase12/payload_drop_obligation.lang
  - testdata/phase12/payload_duplicate_payload_type.lang
  - testdata/phase12/payload_layout_mismatch.golden.c
  - testdata/phase12/payload_missing_binder.lang
  - testdata/phase12/payload_resource_refused.lang
  - testdata/phase12/payload_tracer.lang
findings:
  critical: 0
  warning: 0
  info: 0
  total: 0
status: clean
---

# Phase 12: Code Review Report (re-review after gap closure)

**Reviewed:** 2026-09-13
**Depth:** standard
**Files Reviewed:** 32
**Status:** clean

## Summary

This is a re-review of Phase 12 (result payloads) after gap-closure plans
12-06/12-07/12-08 closed the four findings raised in the prior 12-REVIEW.md
(CR-01, WR-01, WR-02, IN-01). I re-read every file in the required set
against the code as it exists now, cross-checked each of the four prior
findings' claimed dispositions against the actual committed code (not just
against PHASE-12-DEBT.md's narrative), and looked for anything new the
gap-closure plans might have introduced.

**Verification of the four prior dispositions (all confirmed factually
correct against the code):**

- **CR-01 (duplicate payload-type mis-resolution) — confirmed FIXED.**
  `check.go`'s `program.Data` walk (lines 158-176) now refuses a data
  declaration whose alternatives collide on `PayloadType` at declaration
  time (`check.duplicate_payload_type`), positioned before any function
  body is checked. Both engines' independent first-match derivations are
  gone: `cgen.go` and `interp.go` no longer define
  `alternativeNameForPayloadType` at all (grepped, zero hits in production
  code) — both now call the single shared `core.AlternativeNameForPayloadType`
  (`core.go:107-125`), which reports ambiguity via an error return rather
  than silently returning the first match. `cgen.go`'s `emitBranch` header
  logic also now calls the shared `core.LookupAlternativeDetail` rather
  than a local per-package copy.
- **WR-01 (misleading binder-source diagnostic) — confirmed FIXED.**
  `check.go:3447-3465`'s `analyzePayloadArm` now branches the
  `check.binder_on_nullary_alternative` message on which side supplied the
  binder, naming the offending binder identifier and both the pattern and
  value alternatives when it was construction-side. Traced the arm shapes
  by hand (`Nope(x) => Nope` and `Nope => Ok(w)`) against the branch logic
  and both produce the documented message text.
- **WR-02 (fault-injection seam could no-op vacuously) — confirmed FIXED.**
  `cgen.go:78-93` adds `payloadSlotSwapInjectedWriteCount` /
  `PayloadSlotSwapInjectedWriteCount()`, incremented only on
  `wrongPayloadSlot`'s successful branch (`cgen.go:2145-2146`).
  `session_payload_control_test.go:246-249` asserts `injected >= 1` before
  trusting the mutated-beat comparison, with a diagnostic failure message
  naming exactly what fixture-shape regression would cause a silent no-op.
- **IN-01 (`emitBranch`'s local alternative-detail reimplementation) —
  confirmed FIXED.** `cgen.go:1877` calls `core.LookupAlternativeDetail`
  directly; no local reimplementation remains in `cgen.go`.

**New code introduced by the gap-closure plans (06-08) was reviewed for
new defects.** I specifically traced:
- `check.go`'s new `check.duplicate_payload_type` walk for scoping
  correctness (per-declaration only, resets `firstAlternativeByPayloadType`
  inside the `program.Data` loop — confirmed it does not leak state across
  data types) and ordering (runs after the resource-payload refusal walk
  and before `functionNames` is built, so no function body is checked
  before either refusal fires).
- `core.AlternativeNameForPayloadType`'s three-way switch
  (`0`/`1`/`>1` matches) for off-by-one or short-circuit errors — none
  found; the ambiguous case correctly collects every colliding name into
  `matches` before reporting, so the error message names all colliding
  alternatives, not just the first two.
- `corevalidate.go`'s `OpConstructPayload`/`OpDestructurePayload` cases
  (lines 2011-2043) for an asymmetry: `OpConstructPayload`'s inline
  `v.check` does not re-verify `target.TypeID != ""` the way
  `OpDestructurePayload`'s does. Traced this back through the earlier,
  generic per-operation structural loop (`corevalidate.go:990-1023`),
  which already requires `places[operation.TargetID]` to exist for every
  operation kind except the four terminator kinds and
  `OpDestructurePayload` (which is exempted because it names its produced
  place via `PayloadTargetID` instead, checked separately on the next
  line) — `OpConstructPayload` is not in that exemption list, so its
  target's existence is already guaranteed by the earlier pass before the
  per-kind switch ever runs. Not a defect; the two cases' apparent
  asymmetry is accounted for by an earlier shared check.
- `cgen.go`'s `wrongPayloadSlot` fallback (`ok == false`, no other
  payload-carrying alternative to misdirect into) — confirmed the fallback
  path (`cgen.go:2156-2158`) writes the *correct* value rather than
  silently corrupting something else, matching its doc comment.

**Fixture and test review.** All nine `testdata/phase12/*.lang` fixtures'
header comments were checked against what the corresponding code path
actually does (not just what the comment claims) — each one's described
refusal/acceptance matches the diagnostic code and message the described
code path would actually produce. `payload_layout_mismatch.golden.c` is
correctly a committed *mutated* fixture (transposed struct fields) with no
untransposed counterpart ever intended to be committed, matching its own
doc comment. `cgen_n1_convergence_test.go`'s five-row table matches the
measured D-12-36 finding recorded in PHASE-12-DEBT.md verbatim (fixture,
legacy/program success bits, and byte-identity bit all cross-checked).

No new Critical, Warning, or Info findings are raised by this re-review.
All reviewed files meet quality standards; the four previously-raised
findings are genuinely closed, not merely marked closed.

---

_Reviewed: 2026-09-13_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
