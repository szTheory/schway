---
phase: 12-result-payloads
plan: 02
subsystem: compiler-payload-representation
tags: [core-ir, ast, parser, check, corevalidate, interp, cgen, payload, D-12-05, D-12-25, tracer]

# Dependency graph
requires:
  - phase: 12-result-payloads (plan 01)
    provides: RE-DEFER ratification of D-12-31 (six legacy cgen emitters kept, byte-untouched), PHASE-12-DEBT.md
provides:
  - Two new core.OperationKind values (OpConstructPayload, OpDestructurePayload) registered in AllOperationKinds() (now 12 kinds)
  - core.AlternativeDetail / DataType.AlternativeDetails, guarded by the D-12-09 NewDataType constructor
  - ast.Alternative.PayloadType and ast.MatchArm.Binder, plus parser support for `Ok(Buffer)` / `Ok(v) =>` / `=> Ok(v)`
  - checkBranch's payload-arm lowering (analyzePayloadArm) and the shared PayloadRecordLayout derivation
  - corevalidate/interp/cgen support for both new operation kinds, proven by a real interpreter/-O0/-O3 three-engine agreement test
  - Ratified published surface (source spelling, six diagnostic code strings, D-12-19 replay shape) for plans 03-05 to cite verbatim
affects: [12-03, 12-04, 12-05]

# Actuals (#2632)
actuals:
  tokens: 17250
  tasks: 3
  commits: 2

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Bare-value match arms whose scrutinee type declares a payload alternative are lowered through checkBranch's block/edge scaffolding (D-12-05), not the plain Match-only path — analyzePayloadArm builds operations directly from arm.Pattern/Value/Binder instead of parsing an ast.LinearBody"
    - "A genuinely type-widening/narrowing operation (source type != produced-place type) sets operation.TypeID to its SOURCE's own type (the universal source.TypeID==operation.TypeID pre-switch law every corevalidate replay applies) and lets the newly-produced place carry the different type independently — never reuses targetMatches when target.TypeID would need to differ from operation.TypeID"
    - "Shared checker-derived layout fact (check.PayloadRecordLayout, D-12-25) consumed by cgen only; interp needs no C-specific layout since its own value{tag,payload} widening carries the fact directly"
    - "PayloadType-to-alternative-name resolution (alternativeNameForPayloadType, duplicated independently in interp and cgen per D-09-02) is unambiguous only when a data type's alternatives declare DISTINCT payload types — a documented, deliberate fixture constraint (D-12-41), not a general n-alternatives-one-type solution"

key-files:
  created:
    - testdata/phase12/payload_tracer.lang
    - internal/compiler/cgen/cgen_payload_tracer_test.go
  modified:
    - internal/compiler/core/core.go
    - internal/compiler/ast/ast.go
    - internal/compiler/syntax/parser.go
    - internal/compiler/check/check.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/interp/interp.go
    - internal/compiler/cgen/cgen.go
    - internal/compiler/native/native.go
    - internal/compiler/session/session.go
    - internal/compiler/core/core_test.go
    - internal/compiler/core/core_convention_absence_test.go
    - internal/compiler/interp/interp_test.go
    - internal/compiler/interp/interp_oracle_golden_test.go

key-decisions:
  - "Task 1 checkpoint ratified as-is (developer: \"Ratify as proposed\") — see the full ratification table below, which plans 03-05 cite verbatim."
  - "Fixture uses TWO data types (Fault + Outcome), matching the plan's original hard-constraints fixture, with Ok(Buffer)/Err(Fault) declaring DISTINCT payload types — this is what lets PayloadType alone disambiguate which alternative an operation belongs to in interp/cgen, at the cost of not yet supporting two alternatives sharing one payload type (D-12-41, deferred)."
  - "PayloadRecordLayout is exported (deviation from the plan's proposed unexported payloadRecordLayout name) so cgen can read the ONE shared layout fact without check re-deriving it or cgen re-deriving it independently (D-12-25's whole point)."
  - "Two structural bugs were discovered and fixed in-scope (Rule 3, blocking): cgen.emitBranch and session.interpreterInputs both assumed program.DataTypes has exactly one entry. A payload alternative referencing another declared data type as its own payload (Err(Fault)) needs two. Both now resolve the scrutinee's own data type BY NAME against function.Parameter.Type; byte-identical for every pre-Phase-12 single-data-type program."
  - "native.go's execution-event validator's allow-list needed the two new event kinds added (Rule 2, missing critical functionality) — without this every native run of a payload-carrying program was refused as \"unknown execution event kind\" even though both engines individually produced correct output."
  - "The payload's actual runtime bytes are a fixed canned literal per payload type (Buffer -> {1,2,3,4}, matching linearInput's existing convention; a nullary payload type -> its own zero tag), not real per-invocation data — this project's single-CLI-argument protocol (`argv[1]` names only the selected alternative) has no channel to carry arbitrary payload bytes per run. Documented as a functionality gap (D-12-41), not an architectural one: the round trip is still genuinely exercised (destructure -> construct -> return), only the payload's own content is synthetic."

requirements-completed: [RES-02, RES-03]

coverage:
  - id: D1
    description: "Two new core.OperationKind values (OpConstructPayload, OpDestructurePayload) registered in AllOperationKinds(), forcing real fixture coverage via the exhaustive-dispatch control"
    requirement: "RES-02"
    verification:
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestAllOperationKindsRegistered"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestAllOperationKindsHandledAtEverySite"
        status: pass
    human_judgment: false
  - id: D2
    description: "checkBranch lowers a bare-value payload match arm into destructure/construct/return operations, minted from the shared nextIndex authority, corevalidated by independently-written case arms"
    requirement: "RES-02"
    verification:
      - kind: unit
        ref: "go run ./cmd/lang --json check testdata/phase12/payload_tracer.lang (zero diagnostics)"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/... (full package suite)"
        status: pass
    human_judgment: false
  - id: D3
    description: "One payload-carrying alternative constructed, bound, and re-constructed on a single path, proven identical across interpreter, native -O0, and native -O3"
    requirement: "RES-03"
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_payload_tracer_test.go#TestPayloadTracerThreeEngineAgreement"
        status: pass
    human_judgment: false
  - id: D4
    description: "Payload struct layout derived exactly once in check (PayloadRecordLayout) and read only by cgen; no second, independent layout derivation exists"
    requirement: "RES-03"
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_payload_tracer_test.go#TestPayloadTracerThreeEngineAgreement (asserts the emitted C declares the tag field)"
        status: pass
    human_judgment: false
  - id: D5
    description: "All four pinned golden-C digests remain byte-identical after this plan's diff"
    verification:
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestPreviousPhaseGoldenCUnchanged"
        status: pass
    human_judgment: false
  - id: D6
    description: "The published surface (source spelling, six diagnostic code strings, D-12-19 replay shape) is developer-ratified in writing before being committed"
    verification: []
    human_judgment: true
    rationale: "A developer decision recorded verbatim in this SUMMARY (Task 1's checkpoint), not a mechanically-checkable fact."

duration: 95min
completed: 2026-09-12
status: complete
---

# Phase 12 Plan 02: Payload Round Trip Tracer Summary

**One payload-carrying alternative (Buffer/Fault) is constructed, destructured, and re-constructed on a single path — two new core.OperationKinds, a shared checker-derived C layout, and full interpreter/-O0/-O3 agreement — committed as the skeleton plans 03-05 expand.**

## Performance

- **Duration:** 95 min (Task 1 checkpoint resolution + Tasks 2-3 implementation)
- **Started:** 2026-09-12T22:10:00Z (approx.)
- **Completed:** 2026-09-12T23:45:00Z (approx.)
- **Tasks:** 3
- **Files modified:** 13 modified, 2 created

## Accomplishments

- `core.AllOperationKinds()` now returns 12 kinds; `OpConstructPayload`/`OpDestructurePayload` are real, distinct kinds (never a conditionally-populated field on `OpCopy`/`OpMove`), forcing the exhaustive-dispatch control to demand real fixture coverage.
- `core.NewDataType` is the sole permitted constructor for a `DataType` carrying payload alternatives — a name present in `Alternatives` but absent (or duplicated) in `AlternativeDetails` is structurally unconstructible.
- The parser accepts `data Outcome = | Ok(Buffer) | Err(Fault)`, a pattern binder `Ok(v) =>`, and a construction argument `=> Ok(v)`, sharing one `optionalPayloadBinder` helper and two new grammar-level refusals (`syntax.expected_binder`, `syntax.expected_rparen`).
- `checkBranch` lowers a bare-value payload match arm (no braces) into real Linear IR — `OpDestructurePayload` then `OpConstructPayload` then `OpReturn` — minted from the same `nextIndex` authority every other operation uses, with a freshly-derived `TypeFact` for the payload place.
- `corevalidate`, `interp`, and `cgen` each independently support both new kinds; `testdata/phase12/payload_tracer.lang`'s `identity` function runs identically through the interpreter, native `-O0`, and native `-O3` (`TestPayloadTracerThreeEngineAgreement`).
- All four pinned golden-C digests are byte-identical; `go test ./...` exits 0.
- The published surface (source spelling, six diagnostic codes, D-12-19 replay shape) was developer-ratified before any of it was committed.

## Task Commits

Each code task was committed atomically:

1. **Task 1: Ratify the published surface** - checkpoint:decision, no code commit (decision recorded verbatim below, per plan and per the 12-01 precedent)
2. **Task 2: Core IR, AST, and parser — the two new kinds and the binder** - `0c90466` (feat)
3. **Task 3: End-to-end "a payload survives a round trip"** - `924efc0` (feat)

**Plan metadata:** (this commit, following SUMMARY creation)

## Task 1 Checkpoint: Developer Ratification

**Ratified: "Ratify as proposed" (verbatim).** The amend option was explicitly presented and explicitly not chosen — no value below was amended from the planner's proposal.

**Source spelling (D-12-11/D-12-12):**

| Element | Ratified spelling |
|---|---|
| Declaration | `| Ok(Buffer)` |
| Pattern binder | `Ok(v) =>` |
| Construction | `Ok(v)` in arm-value position |

`Result`/`Outcome` is a fixture name, never a compiler-blessed built-in — the feature is general over any user-declared `data` type.

**Diagnostic code strings, ratified verbatim:**

| Code | Refuses |
|---|---|
| `check.payload_arity_mismatch` | a pattern whose binder count does not match the alternative's declared payload arity |
| `check.binder_on_nullary_alternative` | a binder on an alternative declared with no payload |
| `check.missing_payload_binder` | a payload-carrying alternative matched with no binder |
| `check.resource_payload_refused` | D-12-27's fail-closed refusal of any alternative whose payload type structurally contains a Phase-4 tracked-resource-derived value |
| `syntax.expected_binder` | opening paren in pattern position, no identifier inside |
| `syntax.expected_rparen` | unclosed binder |

**D-12-19 corpus replay:** ratified as **one globbing test** (matching the `phaseArtifactGlob` / `TestDebtRegistersAreWellFormed` precedent), NOT a hand-enumerated table.

These exact strings and this exact spelling are what plans 03, 04, and 05 must cite.

## Files Created/Modified

- `internal/compiler/core/core.go` - `OpConstructPayload`/`OpDestructurePayload` consts, `AllOperationKinds()` append, `PayloadType`/`PayloadTargetID` additive fields on `LinearOperation`, `AlternativeDetail` struct, `DataType.AlternativeDetails`, `NewDataType` constructor
- `internal/compiler/ast/ast.go` - `Alternative.PayloadType`, `MatchArm.Binder`
- `internal/compiler/syntax/parser.go` - `optionalPayloadBinder` helper; `dataDecl` and `matchExpr` both extended
- `internal/compiler/check/check.go` - `NewDataType` migration; bare-payload-arm routing into `checkBranch`; `analyzePayloadArm`; `PayloadRecordLayout` (exported); `lookupAlternativeDetail`/`dataTypeHasPayload` helpers; `ownershipSupport.Types` additive field
- `internal/compiler/corevalidate/corevalidate.go` - `matchBranchStructural`'s generic place/target checks extended for `OpDestructurePayload`'s `PayloadTargetID`; `replayBlocks` gains independently-written `OpConstructPayload`/`OpDestructurePayload` case arms
- `internal/compiler/interp/interp.go` - `type value struct{tag, payload string}`; `frame.values` and all three frame constructors widened; new dispatch case arms; `alternativeNameForPayloadType` helper
- `internal/compiler/cgen/cgen.go` - `emitBranch`/`emitBranchOperations` gain payload struct emission reading `check.PayloadRecordLayout`; data-type-by-name lookup fix (see Deviations)
- `internal/compiler/native/native.go` - execution-event validator allow-list gains the two new event kinds
- `internal/compiler/session/session.go` - `interpreterInputs` resolves the scrutinee's data type by name instead of assuming array length 1
- `testdata/phase12/payload_tracer.lang` - the tracer fixture (`Outcome{Ok(Buffer), Err(Fault)}`, identity function)
- `internal/compiler/cgen/cgen_payload_tracer_test.go` - `TestPayloadTracerThreeEngineAgreement`
- `internal/compiler/core/core_test.go`, `core_convention_absence_test.go`, `interp_test.go`, `interp_oracle_golden_test.go` - pinned-count/field-list tripwires updated deliberately; test helper call sites widened to the new `map[string]value` shape

## Decisions Made

See `key-decisions` in the frontmatter above for the full list. Summarized:

1. Task 1's ratification is final and verbatim (see table above).
2. The fixture keeps the plan's original two-data-type shape (`Outcome`/`Fault`) with distinct payload types per alternative, rather than simplifying to one data type — this is what makes `PayloadType` alone sufficient to resolve "which alternative" in interp/cgen (documented as D-12-41's deliberate scope limit, not a general solution).
3. `PayloadRecordLayout` is exported (not `payloadRecordLayout` as literally named in the plan's action text) so `cgen` can consume it directly — the D-12-25 "one shared fact" requirement is otherwise unreachable across packages.
4. `operation.TypeID` for both new kinds names the SOURCE place's own type (the universal `source.TypeID == operation.TypeID` law every corevalidate replay site applies), never the produced place's type — the produced place carries its own, genuinely different, `TypeID` independently. This mirrors `OpForeignCall`'s existing `TargetID`/`ErrTargetID` precedent (different places, different types, one shared `operation.TypeID` naming only the source).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] `cgen.emitBranch` assumed exactly one `core.DataType` per program**
- **Found during:** Task 3 (native run of the tracer fixture failed with `native.compile_failed`/data-type lookup errors)
- **Issue:** `emitBranch` refused with `len(program.DataTypes) != 1`. The tracer fixture legitimately declares two data types (`Fault` and `Outcome`, since `Err`'s payload type is itself a declared ADT) — a shape the plan's own hard-constraints fixture required but no prior emitter had exercised.
- **Fix:** `emitBranch` now resolves the scrutinee's own data type BY NAME against `function.Parameter.Type`, never by array position. For every pre-Phase-12 program (which always has exactly one data type matching the scrutinee), the resolved value — and every emitted byte — is unchanged.
- **Files modified:** `internal/compiler/cgen/cgen.go`
- **Verification:** `TestPreviousPhaseGoldenCUnchanged` still passes (four pinned digests byte-identical); `TestPayloadTracerThreeEngineAgreement` passes.
- **Committed in:** `924efc0` (Task 3 commit)

**2. [Rule 3 - Blocking] `session.interpreterInputs` had the identical one-data-type assumption**
- **Found during:** Task 3 (native run via `session.RunNativeFile`, before the emitBranch fix was even reached)
- **Issue:** `interpreterInputs` gated its Match-arm branch on `len(program.DataTypes) == 1`, returning `nil, false` for the tracer fixture and short-circuiting the whole three-engine run with `os.ErrInvalid`.
- **Fix:** Same by-name resolution as `emitBranch` above.
- **Files modified:** `internal/compiler/session/session.go`
- **Verification:** `go test ./internal/compiler/session/...` passes; `TestPayloadTracerThreeEngineAgreement` passes.
- **Committed in:** `924efc0` (Task 3 commit)

**3. [Rule 2 - Missing Critical] `native.go`'s execution-event validator rejected the two new event kinds**
- **Found during:** Task 3 (native run failed with `native.invalid_execution: unknown execution event kind` even after both compile and runtime succeeded)
- **Issue:** The validator's per-event-kind `switch` has a fail-closed `default: return errors.New("unknown execution event kind")`. Without an explicit allow-list entry, every native execution of a payload-carrying program was refused as malformed, even though the emitted JSON was correct.
- **Fix:** Added `"value.payload_constructed"`/`"value.payload_destructured"` to the existing linear-transition case alongside `value.copied`/`value.transferred`/etc. (identical field-shape requirements — non-terminal, both source and target places, a type ID).
- **Files modified:** `internal/compiler/native/native.go`
- **Verification:** `TestPayloadTracerThreeEngineAgreement` passes; `go test ./internal/compiler/native/...` passes.
- **Committed in:** `924efc0` (Task 3 commit)

**4. [Rule 1 - Bug] C compound-literal syntax error in the payload construction site**
- **Found during:** Task 3 (native compile failed: brace-init list used on the right-hand side of a plain assignment)
- **Issue:** `payloadCannedInitializer("Buffer")` originally returned a bare `{{1u,2u,3u,4u},4u}` brace list, which is legal only in a declaration initializer, not a plain assignment (`field = {...};` is a C syntax error outside a declaration).
- **Fix:** Wrapped in a C99 compound literal cast: `(LANG_BUFFER){{1u, 2u, 3u, 4u}, 4u}`.
- **Files modified:** `internal/compiler/cgen/cgen.go`
- **Verification:** clang `-std=c17 -Wall -Wextra -Werror -pedantic` at both `-O0` and `-O3` compiles clean; `TestPayloadTracerThreeEngineAgreement` passes.
- **Committed in:** `924efc0` (Task 3 commit)

---

**Total deviations:** 4 auto-fixed (2 blocking, 1 missing critical, 1 bug). **Impact:** All four were required for the tracer's own stated success criterion (real interpreter/-O0/-O3 agreement) to be reachable at all — none is scope creep; each was discovered by actually running the fixture through every named engine, not inferred in advance.

## Known Stubs / Simplifications

- **Payload byte content is a fixed canned literal, not real per-invocation data.** This project's single-CLI-argument protocol (`argv[1]` names only the selected alternative, e.g. `"Ok"`) has no channel to carry arbitrary payload bytes through a `lang run` invocation. `payloadCannedInitializer` hardcodes `{1,2,3,4}` for `Buffer` and a zero tag for a nullary payload type (mirroring `linearInput`'s existing Buffer/Byte convention). The destructure -> construct -> return round trip is still genuinely exercised end to end; only the payload's own content is synthetic rather than caller-supplied. Documented as D-12-41's functionality gap, not an architectural one. Resolving this (a real multi-argument or structured-input CLI protocol) is out of this plan's scope and has no named owning plan.
- **`alternativeNameForPayloadType` (interp and cgen, independently written per D-09-02) is unambiguous only when a data type's alternatives declare DISTINCT payload types.** Two alternatives sharing one payload type (e.g. `Ok(Buffer) | Retry(Buffer)`) cannot be disambiguated by `operation.PayloadType` alone with the current `core.LinearOperation` shape. This tracer's fixture deliberately avoids that case (`Ok(Buffer)`, `Err(Fault)` — distinct types). A general n-alternatives-sharing-one-type solution (e.g. naming the alternative directly on the operation) is future work; no plan currently owns it.
- **`check.PayloadRecordLayout`'s declared `Size`/`Alignment`/`Offset` fields are informational, not mechanically verified against a real C compiler's actual struct layout this plan.** No `_Static_assert`/`offsetof` conformance check exists yet for the payload struct (that is plan 05's `PayloadLayoutMutationRunner`, per `PATTERNS.md`). This plan's own emitted C compiles and runs correctly under `-Wall -Wextra -Werror -pedantic` at `-O0`/`-O3`, which is the load-bearing proof for a tracer; the layout FACT itself is not yet independently cross-checked against the compiler's own ABI.

These are flagged for plan 05's anti-vacuity controls, not silently dropped.

## Issues Encountered

None beyond the four auto-fixed deviations above, all resolved within Task 3's own scope.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plans 03, 04, and 05 can cite the ratified source spelling and six diagnostic code strings verbatim from the table above.
- `check.PayloadRecordLayout` is exported and ready for plan 05's conformance-control work.
- The `alternativeNameForPayloadType` limitation (distinct-payload-types-only) should be surfaced to plan 03 if its fixtures need two alternatives sharing one payload type — none of plan 03's named fixtures (`payload_arity_mismatch.lang`, `payload_binder_on_nullary.lang`, etc., per `artifacts_this_phase_produces`) appear to need this, but it is worth an explicit check before authoring them.
- The `emitBranch`/`interpreterInputs` by-name data-type resolution fix is a general capability improvement (not scoped narrowly to this fixture) and should make any future multi-data-type branch program work correctly without further changes.

## Self-Check: PASSED

- FOUND: internal/compiler/core/core.go
- FOUND: internal/compiler/ast/ast.go
- FOUND: internal/compiler/syntax/parser.go
- FOUND: internal/compiler/check/check.go
- FOUND: internal/compiler/corevalidate/corevalidate.go
- FOUND: internal/compiler/interp/interp.go
- FOUND: internal/compiler/cgen/cgen.go
- FOUND: internal/compiler/native/native.go
- FOUND: internal/compiler/session/session.go
- FOUND: testdata/phase12/payload_tracer.lang
- FOUND: internal/compiler/cgen/cgen_payload_tracer_test.go
- FOUND commit: 0c90466 (Task 2)
- FOUND commit: 924efc0 (Task 3)
- `go test ./...` exits 0
- `go test ./internal/compiler/core/... -run TestPreviousPhaseGoldenCUnchanged -count=1` exits 0 (four pinned digests byte-identical)
- `go test ./internal/compiler/cgen/... -run TestPayloadTracer -count=1` exits 0

---
*Phase: 12-result-payloads*
*Completed: 2026-09-12*
