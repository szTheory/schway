# Phase 17: Return Type ≠ Parameter Type - Pattern Map

**Mapped:** 2026-09-22  
**Files analyzed:** 19 planned production, test, fixture, and evidence artifacts  
**Analogs found:** 19 / 19

## File Classification

| New/Modified File | Role | Data Flow | Closest tracked analog | Match quality |
|---|---|---|---|---|
| `internal/compiler/check/check.go` | service | transform | `internal/compiler/check/check.go` | exact seam extension |
| `internal/compiler/check/check_test.go` | test | request-response | `internal/compiler/check/check_test.go` | exact |
| `internal/compiler/check/check_repair_emission_test.go` | test | request-response | `internal/compiler/check/check_repair_emission_test.go` | exact |
| `internal/compiler/corevalidate/corevalidate.go` | service | transform | `internal/compiler/corevalidate/corevalidate.go` | exact seam extension |
| `internal/compiler/corevalidate/corevalidate_test.go` | test | transform | `internal/compiler/corevalidate/corevalidate_test.go` | exact |
| `internal/compiler/originvalidate/originvalidate.go` | service | transform | `internal/compiler/originvalidate/originvalidate.go` | exact seam extension |
| `internal/compiler/originvalidate/originvalidate_test.go` | test | transform | `internal/compiler/originvalidate/originvalidate_test.go` | exact |
| `internal/compiler/cgen/cgen_program.go` | service | transform | `internal/compiler/cgen/cgen_program.go` | exact sole-emitter extension |
| `internal/compiler/cgen/cgen_program_test.go` | test | transform | `internal/compiler/cgen/cgen_payload_tracer_test.go` | role-match |
| `internal/compiler/reduce/reduce.go` | service | transform | `internal/compiler/reduce/reduce.go` | exact |
| `internal/compiler/reduce/reduce_test.go` | test | transform | `internal/compiler/reduce/reduce_test.go` | exact |
| `internal/compiler/session/session_phase17_test.go` (new) | test | batch | `internal/compiler/session/session_phase11_differential_test.go` | role/data-flow match |
| `internal/compiler/session/session_phase17_repair_test.go` (new) | test | file-I/O | `internal/compiler/session/session_phase13_injectors_test.go` | role/data-flow match |
| `internal/compiler/session/session_phase17_injectors.go` (new, only if a dedicated injector is needed) | utility | transform | `internal/compiler/session/session_phase13_injectors.go` | exact |
| `cmd/lang-repair/repair.go` | controller | request-response | `cmd/lang-repair/repair.go` | exact protocol consumer |
| `cmd/lang-repair/repair_test.go` | test | file-I/O | `cmd/lang-repair/repair_test.go` | exact |
| `cmd/lang-repair/import_boundary_test.go` | test | transform | `cmd/lang-repair/import_boundary_test.go` | exact |
| `testdata/phase17/*.lang` (new tracer/frontier/derivation/held-out corpus) | config | file-I/O | `testdata/phase13/{derivation,heldout}_call_argument_mismatch.lang` | role/data-flow match |
| `testdata/phase17/HELDOUT.sha256` (new) | config | file-I/O | `testdata/phase13/HELDOUT.sha256` | exact |

`PHASE-13-DEBT.md` is a tracked historical authority, not an implementation target: preserve the Phase 13 corpus as a no-op/unrepairable control and put the D-13-02b successor witness in Phase 17 evidence/tests rather than editing the historical record.

## Pattern Assignments

### `internal/compiler/check/check.go` (service, transform)

**Analog:** same tracked file.

**Signature admission and fact minting** (lines 3146-3178; also the match admission head at 247-258):

```go
parameterType := coreType(function.Parameter.Type)
if !sameType(function.ReturnType, function.Parameter.Type) {
    return core.Function{}, []diagnostic.Diagnostic{
        diagnostic.Error("type.return_mismatch", function.Span,
            "linear result must have the parameter type")}, typeNodeCount(parameterType), nil
}
derived, err := ability.Derive(parameterType)
typeID := functionID + ":type:0"
```

Replace the single-type assumption with locally derived parameter and return facts. Do not pass a checked return summary into either validator peer. Update all three checker admission heads (ordinary linear, fallible linear, and match) together.

**Named source-call diagnostic** (lines 1748-1775 and 3460-3559):

```go
causes := []diagnostic.Cause{
    {Kind: "callee", Detail: contract.ID},
    {Kind: "argument_type", Detail: typeFact.Shape.Constructor},
    {Kind: "declared_parameter_type", Detail: contract.ParameterType},
}
diag := diagnostic.ErrorWithRepairs(
    checkCallArgumentTypeMismatch, binding.RHS.Span,
    "call argument type does not match the callee's declared parameter type", causes)
```

Preserve diagnostic primary span and ordered TYP-02 causes. Make TYP-03 use the existing separate code with ordered `callee`, `declared_return_type`, and caller-visible type causes; it must become source-reachable rather than remain seam-only.

**Repair emission** (current deliberately-disabled precedent at lines 3545-3580; unit partition harness in `check_repair_emission_test.go:104-218`): enumerate initialized caller places from `places`, select a real unique place whose fact is the callee parameter fact, and emit only a sealed, machine-applicable `use_matching_argument` repair at `binding.RHS.Span`. Retain zero/multiple/uninitialized fail-closed partitions.

### `internal/compiler/{corevalidate,originvalidate}/*.go` (services, transform)

**Analogs:** `internal/compiler/corevalidate/corevalidate.go:2310-2360,2820-2855`; `internal/compiler/originvalidate/originvalidate.go:900-950`.

**Independent ability/contract derivation:**

```go
parameterContract := core.ParameterContract{
    ID: function.Parameter.ID, Name: function.Parameter.Name, Type: function.Parameter.Type,
    Mode: parameterContractMode(),
    Drops: hasDropAbility && !peerParameterEscapesOwned(function),
}
returnContract := core.ReturnContract{Type: function.ReturnType}
returnContract.Fresh = returnContract.Mode == "owned" && hasDropAbility
```

Replace `hasDropAbility` with each peer's own parameter-type and return-type lookups, respectively. The two packages must implement their derivations separately; shared primitive ability rules are fine, but no shared return-contract helper or checker-produced artifact.

**Directional call peer** (`corevalidate.go:2838-2855`):

```go
argumentTypeMatches := argumentConstructor != "" &&
    callee.Parameter.Type != "" && argumentConstructor == callee.Parameter.Type
if !v.check(disableCallArgumentTypePeerForTest || argumentTypeMatches,
    core.CallArgumentTypeMismatch, operation.ID) { return false }

targetTypeMatches := targetConstructor != "" &&
    callee.ReturnType != "" && targetConstructor == callee.ReturnType
return v.check(disableCallReturnTypePeerForTest || targetTypeMatches,
    core.CallReturnTypeMismatch, operation.ID)
```

Keep fail-closed peer checks and distinct diagnostic codes. Add return-only seeded faults in each peer test; confirm parameter facts remain intact, a single peer fails/diverges, and a coordinated fault makes the agreement gate fail.

### `internal/compiler/cgen/cgen_program.go` (service, transform)

**Analog:** same tracked file, the only production emission family.

**C type allocation/prototype** (lines 554-571 and 639-652):

```go
functionNames := make([]string, len(functions))
typeNames := make([]string, len(functions))
// ... typeNames[index] is derived from the function's one current type
fmt.Fprintf(&out, "static %s %s(%s, unsigned int);\\n",
    typeNames[index], functionNames[index], typeNames[index])
```

Port this to a parameter-C-type/return-C-type pair through allocation, prototypes, `emitProgramFunction` definitions, and OpCall lowering. Do not revive any retired emitter.

**Entry input/output distinction** (lines 1128-1165):

```go
fmt.Fprintf(&out, "  %s lang_entry_input = %s;\\n", entryTypeName, initializer)
fmt.Fprintf(&out, "  %s lang_entry_output = %s(lang_entry_input, 0u);\\n",
    entryTypeName, functionNames[entryIndex])
```

Input must use entry parameter C type; call result and JSON rendering must use entry return C type. Add generated-C structural assertions alongside execution.

### `internal/compiler/reduce/reduce.go` (service, transform)

**Analog:** same tracked file, specifically source projection at lines 861-962 and 1103-1111.

```go
if isDeclaredDataType(program, fn.Parameter.Type) || isDeclaredDataType(program, fn.ReturnType) {
    return unsupportedProjection(...)
}
fmt.Fprintf(&out, "fn %s(%s: %s) -> %s {\\n%s%s}\\n\\n",
    fn.Name, fn.Parameter.Name, fn.Parameter.Type, fn.ReturnType, body.String(), terminal)
```

Audit fallback TypeID/type declaration selection so return facts never silently inherit parameter facts; preserve the established explicit handling for a distinct declared return type.

### `internal/compiler/session/session_phase17_test.go` (new test, batch)

**Analog:** `internal/compiler/session/session_phase11_differential_test.go:25-210`.

```go
func phase11RunFourTiers(t *testing.T, ctx context.Context, program core.Program,
    entryName, input string) map[string]execution.Execution {
    interpreted, err := interp.Run(program, entryName, input)
    cSource, err := supplier(program)
    o0, err := runner.Run(ctx, cSource, "-O0", []string{input})
    o3, err := runner.Run(ctx, cSource, "-O3", []string{input})
    ltoRunner.LTO = true
    o3lto, err := ltoRunner.Run(ctx, cSource, "-O3", []string{input})
}
```

Use a checked-in nominal `Resource -> Result` program with `main -> classify`; check it, run interpreter/O0/O3/O3-LTO, compare all tiers, then assert C prototypes/definitions/calls and entry rendering contain the two distinct types. Source-frontier tests should pin the pre-widening refusal then assert movement to the named calls and causes, avoiding earlier ownership/declaration errors.

### `testdata/phase17/*` and held-out integrity test (config, file-I/O)

**Analog:** `testdata/phase13/derivation_call_argument_mismatch.lang:1-23`, `testdata/phase13/heldout_call_argument_mismatch.lang:1-48`, and `internal/compiler/session/session_phase13_injectors_test.go:451-500`.

```go
sum := sha256.Sum256(data)
got := hex.EncodeToString(sum[:])
if got != entry.digest { t.Fatalf("%s: digest mismatch ...", entry.relPath) }
// Every heldout_*.lang on disk must also occur in HELDOUT.sha256.
```

Create parser-valid tracer, two frontier negatives, plus a derivation/held-out repair pair before repair emission. The held-out graph must retain the `main -> relay -> dispatch` depth distinction, be byte-sealed in a new Phase 17 manifest, and historical Phase 13 held-outs remain unchanged controls.

### `cmd/lang-repair/{repair.go,repair_test.go,import_boundary_test.go}` (controller/tests, request-response/file-I/O)

**Analogs:** `repair.go:409-455`, `repair_test.go:570-617`, and `import_boundary_test.go:20-96`.

```go
outcome, err := Repair(context.Background(), langBinary, sourcePath)
if outcome.Status != OutcomeRepaired { t.Fatalf(...) }
if outcome.SubprocessCount != 2 { t.Fatalf(...) }
verify := testsupport.RunCLI(t, langBinary, nil, "--json", "check", sourcePath)
```

The driver remains a protocol-only consumer: do not inspect diagnostic prose. Extend its existing tests for exact `repaired`, exactly two real JSON checks, independent clean re-check, original held-out bytes unchanged, plus red controls stripping/mutating kind, span, replacement, and prose. Maintain the AST-based source/import scans; add a production-source scan that rejects held-out fixture path references.

## Shared Patterns

### Independent admission peers

**Sources:** `check.go:3146-3178`; `corevalidate.go:2310-2360,2820-2855`; `originvalidate.go:900-950`.

Parameter ownership (`Drops`) and return ownership (`Fresh`) are independently derived at all three layers, followed by peer agreement. Keep test seams local, default false, restore them with `defer`/`t.Cleanup`, and use one-return-side-only mutation before the coordinated disagreement control.

### Stable source diagnostics

**Source:** `check.go:1748-1775,3460-3559`; test contract at `check_test.go:3817-3875`.

Use `diagnostic.Error`/`ErrorWithRepairs`, primary call span, exact distinct checker code, and ordered structured causes. Tests inspect code, severity, cause kind/order/details, and repair shape—not prose.

### Evidence before repair

**Sources:** `session_phase13_injectors_test.go:451-500`; `repair_test.go:570-617`.

Seal all `heldout_*.lang` bytes before repair implementation. Operate on `t.TempDir()` copies, assert source bytes remain exactly unchanged, and use one real repair pass (two driver subprocesses) plus an independent third clean check.

### Test-only boundary guards

**Sources:** `cmd/lang-repair/import_boundary_test.go:20-96`; tracked paths verified with `git ls-files`.

Use Go AST parsing to scan non-test production source. Extend boundary guards to new derivation files and ensure no production source encodes a held-out path.

## No Analog Found

None. Phase-specific test and fixture filenames are new, but Phase 11 differential evidence and Phase 13 integrity/repair controls provide direct patterns.

## Metadata

**Analog search scope:** `internal/compiler/{check,corevalidate,originvalidate,cgen,reduce,session}`, `cmd/lang-repair`, `testdata/phase07`, `testdata/phase13`  
**Tracked-source gate:** verified all 20 named code/evidence analogs with `git ls-files`  
**Files scanned:** 20 primary analogs plus focused test/fixture searches  
**Pattern extraction date:** 2026-09-22
