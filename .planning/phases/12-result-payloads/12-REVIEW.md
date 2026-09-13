---
phase: 12-result-payloads
reviewed: 2026-09-13T00:00:00Z
depth: standard
files_reviewed: 30
files_reviewed_list:
  - internal/compiler/ast/ast.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_n1_convergence_test.go
  - internal/compiler/cgen/cgen_payload_tracer_test.go
  - internal/compiler/check/check.go
  - internal/compiler/check/check_payload_test.go
  - internal/compiler/core/core.go
  - internal/compiler/core/core_convention_absence_test.go
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
  - testdata/phase12/payload_layout_mismatch.golden.c
  - testdata/phase12/payload_missing_binder.lang
  - testdata/phase12/payload_resource_refused.lang
  - testdata/phase12/payload_tracer.lang
findings:
  critical: 1
  warning: 2
  info: 1
  total: 4
status: issues_found
---

# Phase 12: Code Review Report

**Reviewed:** 2026-09-13T00:00:00Z
**Depth:** standard
**Files Reviewed:** 30
**Status:** issues_found

## Summary

Phase 12 adds payload-carrying `data` alternatives (`Ok(Buffer)`, `Err(Fault)`, ...) across the whole pipeline: parser, AST, check, core, corevalidate, interp, cgen, native, and session's mutation-testing harnesses. The additive design is careful and deliberately documented (D-12-05/D-12-10/D-12-14/D-12-25/D-12-27), and the fail-closed refusals that were implemented (`check.payload_arity_mismatch`, `check.missing_payload_binder`, `check.binder_on_nullary_alternative`, `check.resource_payload_refused`) are correctly wired and exercised by dedicated fixtures.

However, one real correctness gap survives the review: nothing anywhere in the pipeline refuses (or even flags) two distinct alternatives of the same data type declaring the *same* payload type (e.g. `Ok(Buffer) | Also(Buffer)`). Both `cgen.alternativeNameForPayloadType` and `interp.alternativeNameForPayloadType` resolve an operation's `PayloadType` back to an alternative name by linear scan and return the *first* match — silently misattributing `OpConstructPayload`/`OpDestructurePayload` to the wrong alternative whenever two alternatives share a payload type. The code's own comments acknowledge this ("unambiguous only because this tracer's alternatives declare distinct payload types... a data type with two alternatives sharing one payload type is future work") but the compiler never refuses the input that breaks the assumption — it silently produces wrong tag/field mappings instead, which is exactly the kind of silent-incorrectness this project's own design philosophy (explicit, fail-closed refusal for every other named hazard in this phase) argues against. `session.PayloadProbeDataType()` even constructs exactly such a same-payload-type data type (`First(Byte)`, `Second(Byte)`) as a legal `core.DataType`, confirming the shape is reachable and accepted by every validator.

Two secondary quality issues are also noted below (WARNING), plus one documentation nit (INFO).

## Critical Issues

### CR-01: Two alternatives sharing a payload type silently mis-resolve, with no refusal anywhere in the pipeline

**File:** `internal/compiler/check/check.go:3231-3238` (payloadTypeNamesForeignReturnType is a sibling case, not the bug site), and the actual resolution sites:
`internal/compiler/cgen/cgen.go:1939-1946` (`alternativeNameForPayloadType`)
`internal/compiler/interp/interp.go:839-848` (`alternativeNameForPayloadType`)

**Issue:** Both cgen and interp resolve an `OpConstructPayload`/`OpDestructurePayload` operation's `PayloadType` back to a declaring alternative name via a linear scan over `dataType.AlternativeDetails`, returning the *first* alternative whose `PayloadType` matches:

```go
func alternativeNameForPayloadType(dataType core.DataType, payloadType string) string {
	for _, detail := range dataType.AlternativeDetails {
		if detail.PayloadType == payloadType {
			return detail.Name
		}
	}
	return ""
}
```

Nothing in `check.go`, `core.NewDataType`, or `corevalidate.go` refuses a `data` declaration where two alternatives declare the *same* payload type (e.g. `data Outcome = | Ok(Buffer) | Also(Buffer)`). `core.NewDataType` (core/core.go:54-70) only checks that `AlternativeDetails` names are unique and reference a real alternative — it does not check `PayloadType` values for collisions. `corevalidate.go`'s `AlternativeDetails` re-validation (lines 298-306) likewise only checks name uniqueness, not payload-type uniqueness.

Consequently, a program declaring two alternatives with the same payload type compiles and runs without any diagnostic, but `OpConstructPayload` (in a match arm that constructs the *second* alternative, e.g. `=> Also(v)`) will emit a tag write for `Also` (from `alternativeBySource`, which correctly keys by operation.PayloadType→altName resolution) alongside a struct-field write and destructure logic resolved by the *same* ambiguous helper — in `cgen.go`'s `emitBranchOperations` (`OpConstructPayload`/`OpDestructurePayload` cases, ~2083-2144) and `interp.go`'s matching cases (~709-731), both consumers independently call the same ambiguous `alternativeNameForPayloadType` and will consistently pick the *first* declared alternative with that payload type rather than the one the arm actually names via `operation.TargetID`/pattern. This can silently construct/destructure the wrong alternative's field, or (worse) let the interpreter and the -O0/-O3 compiled C path *agree* with each other while both are wrong relative to the source program's actual intent, since both consumers share the identical bug — meaning the project's own "interp/cgen must independently agree" convergence tests (the exact test class this phase is built to strengthen, e.g. `cgen_n1_convergence_test.go`) would NOT catch this defect, because both sides derive from the same flawed helper rather than being independent implementations for this one specific fact.

This directly contradicts the phase's own stated design principle — every other new hazard this phase introduces (arity mismatch, missing binder, binder on nullary, resource-carrying payload) is refused explicitly and fail-closed at check time. This one is not: it's presently a silent correctness hazard rather than a documented restriction enforced by a diagnostic.

`session.PayloadProbeDataType()` (`internal/compiler/session/session.go:612-620`) already constructs exactly this shape (`First(Byte)`, `Second(Byte)`) as a valid `core.DataType`, confirming both that the shape is legal today and that no downstream validator rejects it.

**Fix:** Add a declaration-time refusal in `check.go` (alongside the existing `check.resource_payload_refused` walk, since both run over `program.Data` before any function body is checked) that rejects two alternatives of the same data type declaring the same non-empty `PayloadType`, e.g.:

```go
seenPayloadTypes := make(map[string]string) // payloadType -> first alternative name
for _, declaration := range program.Data {
    seenPayloadTypes = map[string]string{}
    for _, alternative := range declaration.Alternatives {
        if alternative.PayloadType == "" {
            continue
        }
        if first, exists := seenPayloadTypes[alternative.PayloadType]; exists {
            result.Diagnostics = append(result.Diagnostics, diagnostic.Error(
                "check.duplicate_payload_type", alternative.Span,
                fmt.Sprintf("alternative %q and %q both declare payload type %q; this phase requires distinct payload types per alternative so a construct/destructure operation can be resolved unambiguously", first, alternative.Name, alternative.PayloadType),
            ))
            continue
        }
        seenPayloadTypes[alternative.PayloadType] = alternative.Name
    }
}
```

Alternatively (or additionally), remove the ambiguity at the root by giving `core.LinearOperation` an explicit `AlternativeName` fact for `OpConstructPayload`/`OpDestructurePayload` rather than relying on `PayloadType` round-tripping through a name lookup at all — this would also simplify `alternativeNameForPayloadType` out of existence in both cgen and interp.

## Warnings

### WR-01: `analyzePayloadArm`'s "binder on nullary alternative" diagnostic conflates pattern-side and construction-side binders

**File:** `internal/compiler/check/check.go:3372-3414`

**Issue:** `binder` is computed once as `arm.Binder` (pattern-side) falling back to `arm.ConstructBinder` (value/construction-side):

```go
binder := arm.Binder
if binder == "" {
    binder = arm.ConstructBinder
}
...
if patternDetail.PayloadType != "" {
    ...
} else if binder != "" {
    return fail(diagnostic.Error("check.binder_on_nullary_alternative", arm.Span, "binder present on an alternative declared with no payload"))
}
```

When the pattern alternative is nullary but the arm's *value* side supplies a construction binder for a payload-carrying alternative (e.g. `Empty => Ok(w)` where `Empty` is nullary and `Ok` carries a payload), this code path reports `check.binder_on_nullary_alternative` — a diagnostic whose message and code both describe a binder attached to the *pattern's* alternative, when the actual binder in the source was written in the *value* position for a different (payload-carrying) alternative entirely. The refusal is arguably still correct (there is genuinely no place to source `w`'s value from, since D-12-14 provides only one shared payload place per arm and the pattern side produced none), but the diagnostic code/message actively mislead a developer debugging this case — they will look at the pattern's nullary alternative for a binder that isn't textually there.

**Fix:** Either distinguish this case with its own diagnostic code (e.g. `check.payload_binder_has_no_source`), or make the message conditionally reference which side of the arm the binder came from:

```go
} else if binder != "" {
    reason := "binder present on an alternative declared with no payload"
    if arm.Binder == "" && arm.ConstructBinder != "" {
        reason = "construction binder names a payload source, but this arm's pattern alternative declares no payload to bind"
    }
    return fail(diagnostic.Error("check.binder_on_nullary_alternative", arm.Span, reason))
}
```

### WR-02: `wrongPayloadSlot` silently falls back to the correct write when no alternative-mismatch target exists, without signaling the fault-injection seam is a no-op

**File:** `internal/compiler/cgen/cgen.go:60-75`, `2115-2137`

**Issue:** `wrongPayloadSlot` is the fault-injection seam backing `TestPayloadSlotSwapMutationKilled` (D-12-38). When `payloadSlotSwapForTest` is `true` but `dataType` has no *other* payload-carrying alternative to misdirect into, `wrongPayloadSlot` returns `ok == false`, and the caller falls back to emitting the *correct* write with no comment or log distinguishing "mutation requested but inapplicable" from the ordinary non-mutated path (`2135-2137` uses the identical code as the plain `else` branch at `2138-2140`). If a future data type change (e.g. reducing a payload-carrying data type to a single payload alternative) accidentally removes the only fixture this control depends on, `SetPayloadSlotSwapForTest(true)` would silently stop injecting any fault at all and the mutation-kill test would pass vacuously (proving nothing) rather than failing loudly to say "the seam has no alternative to attack."

**Fix:** Have the test harness (or `wrongPayloadSlot` itself, via a package-level counter/flag) assert that `ok == true` was reached at least once whenever `payloadSlotSwapForTest` is active, so a fixture regression that silently disables the seam is caught by the test suite rather than passing green.

## Info

### IN-01: `emitBranch`'s per-arm `detail` lookup ignores its own zero-value initialization pattern

**File:** `internal/compiler/cgen/cgen.go:1841-1848`

**Issue:** Inside `emitBranch`'s per-alternative loop, `detail` is initialized to the zero value and then overwritten only `if candidate.Name == alternative` is found — functionally equivalent to (and duplicating) the already-defined `lookupAlternativeDetail` helper in `check.go`, but re-implemented locally rather than reused (cgen would need to import the pattern, not necessarily the helper, since `lookupAlternativeDetail` is unexported in `check`). This is a minor duplication rather than a bug — flagged as info only because the surrounding code elsewhere in this phase is unusually careful about "read the SAME shared derived fact, never a second independent derivation" (see `payloadFieldBySource`/`PayloadRecordLayout` commentary throughout), and this one spot quietly reintroduces a small independent derivation of alternative-detail lookup.

**Fix:** Export `lookupAlternativeDetail` from `check` (or inline a one-line equivalent with a short comment noting the duplication is intentional/local), for consistency with the "one shared derivation" convention this phase otherwise holds itself to.

---

_Reviewed: 2026-09-13T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
