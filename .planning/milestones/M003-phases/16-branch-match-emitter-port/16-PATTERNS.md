# Phase 16: Branch/Match Emitter Port - Pattern Map

**Mapped:** 2026-09-19  
**Files analyzed:** 10 anticipated modified/created files  
**Analogs found:** 9 / 10

## File Classification

| New/Modified File | Role | Data Flow | Closest tracked analog | Match quality |
|---|---|---|---|---|
| `internal/compiler/cgen/cgen.go` | emitter/dispatch | transform | `internal/compiler/cgen/cgen.go` | exact |
| `internal/compiler/cgen/cgen_program.go` | emitter | transform | `internal/compiler/cgen/cgen_program.go` | exact |
| `internal/compiler/cgen/cgen_n1_convergence_test.go` | test | transform | `internal/compiler/cgen/cgen_n1_convergence_test.go` | exact |
| `internal/compiler/cgen/cgen_program_test.go` | test | request-response/refusal | `internal/compiler/cgen/cgen_program_test.go` | exact |
| `internal/compiler/cgen/cgen_test.go` | test/golden | file-I/O/transform | `internal/compiler/cgen/cgen_test.go` | exact |
| `internal/compiler/core/core_test.go` | test/provenance guard | file-I/O/transform | `internal/compiler/core/core_test.go` | exact |
| `internal/compiler/session/session_phase11_differential_test.go` | integration test | request-response | `internal/compiler/session/session_phase11_differential_test.go` | exact |
| `internal/compiler/session/session_test.go` | document-validation test | file-I/O | `internal/compiler/session/session_test.go` | exact |
| `.planning/REQUIREMENTS.md` | requirements/config | batch | `.planning/REQUIREMENTS.md` | exact |
| `.planning/phases/16-branch-match-emitter-port/PHASE-16-DEBT.md` | debt register/document | batch | `.planning/milestones/M002-phases/10-trusted-interprocedural-oracle/PHASE-10-DEBT.md` | role-match |
| `testdata/phase16/restrict_readonly_probe.c` (if the decision checkpoint needs a checked-in probe) | test fixture | file-I/O | `testdata/phase5/restrict_borrow.golden.c` via `cgen_test.go` | partial; no existing hand-written C probe harness |

All listed analogues were verified with `git ls-files`; none is a generated or ignored mirror.

## Pattern Assignments

### `internal/compiler/cgen/cgen.go` (emitter/dispatch, transform)

**Analog:** same file, `Emit`/`EmitNative` dispatch at lines 112-173.

**Dispatch pattern** (lines 112-138 and 142-173): validate once at each public boundary, normalize with `validated.Program()`, then select a lowering. Phase 16 replaces the function-count/shape selection with the one surviving call, preserving the validation prefix in both APIs.

```go
validated := corevalidate.Validate(program)
if !validated.Valid {
    return "", fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
}
program = validated.Program()
if len(program.Functions) != 1 {
    return emitProgram(program, false)
}
```

**Cutover rule:** Lines 118-137 and 148-172 are the exact legacy routing fork. The atomic deletion is limited to `emitMatch`, `emitBranch`, and `emitLinear`; retain named refusal/cut paths for foreign and unadmitted pointer families rather than silently admitting them.

---

### `internal/compiler/cgen/cgen_program.go` (emitter, transform)

**Analog:** same file, `emitProgram` at lines 412-490 and `emitProgramFunction` at lines 652-760.

**Admission/ordering pattern** (lines 415-473): graph order and entry are established first; each ordered function is admitted/refused by shape; only then do invocation/output preflight and allocation start.

```go
order, err := callgraph.Order(program)
if err != nil { return "", err }
entry, err := callgraph.EntryFunction(program)
if err != nil { return "", err }
// Establish supported shape before preflight.
preflightNodes, err := preflightInvocationPathTable(program, entry)
if err != nil { return "", err }
```

Port the match admission and arm lowering inside this ordered pipeline; do not move preflight above shape checks. The current branch refusal at lines 434-436 is the seam to replace, while foreign and excluded pointer refusals remain named boundaries.

**Ordinary linear-body pattern** (lines 667-717): use the program emitter's local allocator and operation loop, preserve move markers and event attribution, and record into the shared `/2` event buffer.

```go
names := newCNames(reserved...)
locals[parameter.ID] = names.allocate(cLocal(parameter.Name), "place", 0)
for _, operation := range function.Linear.Operations {
    switch operation.Kind {
    case core.OpCopy, core.OpMove, core.OpBorrowShared, core.OpBorrowExclusive:
        // declare local, emit C assignment, then lang_record_event(...)
    }
}
```

**Return/output boundary** (lines 736-750): body helpers return values and record events; `main` alone serializes the document. Do not copy `emitLinear`'s `/1` tail. Current `main` serialization at lines 600-615 is the integration seam; replace its literal empty `live_resources` tail with a derived surviving-emitter value.

---

### `internal/compiler/cgen/cgen_program.go` branch helpers (emitter, transform)

**Analog:** `internal/compiler/cgen/cgen.go`, `emitBranch` lines 1780-1931 and `emitBranchOperations` lines 2002-2134.

**Checker-derived payload layout** (lines 1838-1859): resolve the data type from the program and project `check.PayloadRecordLayout`; do not calculate C field layout locally.

```go
layout := check.PayloadRecordLayout(dataType)
for index, alternative := range dataType.Alternatives {
    field := layout.Fields[index+1]
    payloadFields[alternative] = payloadFieldInfoT{name: field.Name, cType: field.CType}
}
```

**Arm lowering pattern** (lines 1919-1928): validate an arm block, emit the selected case, then delegate its straight-line operations to a helper. Adapt that helper to return through the program function and shared schema `/2` writer, rather than reproducing an inline `main` tail.

```go
for _, arm := range function.Match.Arms {
    block, known := blocksByID[arm.BlockID]
    if !known { return "", fmt.Errorf("arm %q references unknown block %q", arm.ID, arm.BlockID) }
    fmt.Fprintf(&out, "    case %s: {\\n", alternativeBySource[arm.Pattern])
    if err := emitBranchOperations(/* arm data */); err != nil { return "", err }
}
```

**Defect boundary** (lines 2049-2067): retain the generated `lang_defect` path and its `_Noreturn` exemption only. The Phase 16 port must not introduce generic optimizer attributes; preserve the event-before-abort sequence.

**Payload operation pattern** (lines 2108-2134): use `core.AlternativeNameForPayloadType` and the projected checker field name, then record `value.payload_destructured`.

---

### `internal/compiler/cgen/cgen_n1_convergence_test.go` (test, transform)

**Analog:** same file, `TestN1ConvergenceDifferential` lines 43-120.

**Table-driven characterization pattern:** retain exactly five table entries (guard at lines 78-80), parse/check each fixture once, drive legacy public dispatch and direct `emitProgram`, then assert success/refusal and exact bytes where both are admitted.

```go
legacyOutput, legacyErr := Emit(program)
programOutput, programErr := emitProgram(program, false)
if row.legacyOK && row.programOK {
    identical := legacyOutput == programOutput
    if identical != row.identical { t.Fatalf(/* includes both outputs */) }
}
```

Flip the scoped three fixture expectations to identity before deletion. Keep explicit foreign and unadmitted-by-pointer refusal rows; do not delete the characterization merely because public dispatch is later unified. Extend the pattern to exercise both `Emit` and `EmitNative` for the three required fixtures.

---

### `internal/compiler/cgen/cgen_program_test.go` (test, request-response/refusal)

**Analog:** same file, `TestUnsupportedProgramShapePrecedesSchema2Preflight` lines 195-215 and `TestInvocationPreflightGuardIsNotInert` lines 241-261.

**Refusal-precedence and non-inertness pattern:** construct/check a fixture, invoke public `EmitNative`, assert the intended error text, then assert serialization was not reached. Pair a normal control with a bypass/seeded-fault control where the guard itself needs proof.

```go
_, err = cgen.EmitNative(checked.Program)
if err == nil { t.Fatal("expected ... to be refused") }
if !strings.Contains(err.Error(), "...") { t.Fatalf("...: %v", err) }
if cgen.InvocationSerializationReachedForTest() { t.Fatal("... reached C serialization") }
```

Use this structure for D-16-03 ordering and for the exact-source-shape refusal matrix if the bounded `restrict` path is proposed.

---

### `internal/compiler/cgen/cgen_test.go` (test/golden, file-I/O/transform)

**Analog:** `TestExistingEmittersAreByteIdentical`, lines 58-99; by-pointer golden/mutation marker, lines 661-688.

**Golden test pattern:** table fixtures with mode, check source through `session.Check`, call the selected public emitter, read the tracked golden with `testsupport.ProjectPath`, and compare whole strings.

```go
if test.native { generated, err = cgen.EmitNative(checked.Program) } else { generated, err = cgen.Emit(checked.Program) }
golden, err := os.ReadFile(testsupport.ProjectPath(test.golden...))
if generated != string(golden) { t.Fatalf("...\\n--- got ---\\n%s\\n--- want ---\\n%s", generated, golden) }
```

For an admitted `restrict` port, copy the phase-5 structural marker assertion (`strings.Count(..., "/* lang:by-pointer-param */") == 1`, lines 675-687) but add explicit rejection cases for every prohibited extension. A golden alone is not the shape fence.

---

### `internal/compiler/core/core_test.go` (test/provenance guard, file-I/O)

**Analog:** `previousPhaseGoldenCDigests` and `TestPreviousPhaseGoldenCUnchanged`, lines 148-208.

**Digest-map/bijection pattern:** maintain one fixed path-to-SHA-256 map, enumerate tracked golden paths deterministically, fail on an unpinned disk item, hash contents with `sha256.Sum256`, then fail on a pinned-but-missing entry.

```go
want, known := previousPhaseGoldenCDigests[relative]
if !known { t.Fatalf("unpinned golden.c file %s found on disk...", relative) }
sum := sha256.Sum256(data)
got := hex.EncodeToString(sum[:])
if got != want { t.Fatalf("golden.c bytes moved for %s: got sha256 %s, want %s", relative, got, want) }
```

Put the typed four-entry Phase-16 change ledger adjacent to this map and add bidirectional ledger↔map validation: path uniqueness, 64-hex fields, old digest equality, current/new digest equality, semantic witness, N=1 fixture, moved responsibility, structural reason, and review disposition. Add negative controls for missing, duplicate, stale, and map-mismatched entries.

---

### `internal/compiler/session/session_phase11_differential_test.go` (integration test, request-response)

**Analog:** `phase11RunFourTiers`, lines 90-125, and `phase11CompareFourTiers`, lines 127-145.

**Independent semantic comparator pattern:** emit once using `cgen.EmitNative`, run interpreter plus native `-O0`, `-O3`, and `-O3 -flto`, then pass all four documents to `session.Phase5CompareProgramEngines`.

```go
o0, err := runner.Run(ctx, cSource, "-O0", []string{input})
o3, err := runner.Run(ctx, cSource, "-O3", []string{input})
ltoRunner := runner; ltoRunner.LTO = true
o3lto, err := ltoRunner.Run(ctx, cSource, "-O3", []string{input})
if err := session.Phase5CompareProgramEngines(fixture, program, engines); err != nil { t.Fatalf("%s: four-tier disagreement: %v", fixture, err) }
```

Keep this test as defense in depth beside byte identity. Its comments at lines 78-89 establish that a one-translation-unit pure-Lang corpus does not prove LTO non-inertness.

---

### `internal/compiler/session/session_test.go` and Phase-16 debt artifact (document validation, file-I/O/batch)

**Analog:** `TestDebtRegistersAreWellFormed`, lines 2981-3014, and `checkDebtRegister`, lines 3131-3148; register structure in `PHASE-10-DEBT.md`, lines 1-20 and 42-57.

**Register pattern:** use YAML frontmatter with an accurate `items:` count, an `## Items` table, and exactly one `### <ID>` detail section per row. Let the existing glob-driven well-formedness test validate the new artifact; add phase-specific assertions for the required M004 owner, real prerequisite, reopening condition, and `-flto` consequence, because generic format validation does not prove honesty.

```go
for _, path := range registers {
    t.Run(filepath.Base(path), func(t *testing.T) { checkDebtRegister(t, path) })
}
```

**No-third-deferral authority:** copy the explicit amendment language shape from `PHASE-10-DEBT.md` lines 384-403: an item deferred a second time must be cut by a `REQUIREMENTS.md` amendment or becomes never-cut. Do not merely change a landing-phase cell.

---

### `.planning/REQUIREMENTS.md` (requirements/config, batch)

**Analog:** Native Emission requirements at lines 93-104.

**Amendment pattern:** preserve the normative NAT-08/NAT-09 bullets and append an explicit, reviewable amendment under their authority. The existing wording already requires a recorded amendment that names the landing milestone and `-flto` consequence (lines 95-101); Phase 16 must name every excluded legacy family and M004 rather than altering those requirements to conceal a deferral.

## Shared Patterns

### Validation and serialization order

**Source:** `internal/compiler/cgen/cgen_program.go:415-473`  
**Apply to:** `emitProgram` ports and refusal tests.

Graph/entry validation → supported shape → invocation/output preflight → name allocation/serialization is load-bearing. Test that invalid shapes cannot reach serialization.

### Checker-owned payload facts

**Source:** `internal/compiler/cgen/cgen.go:1838-1859`, `2108-2134`  
**Apply to:** match/branch payload lowering.

Derive struct field names and C types from `check.PayloadRecordLayout`; resolve payload alternative identity through `core.AlternativeNameForPayloadType`.

### Deterministic evidence

**Sources:** `internal/compiler/cgen/cgen_test.go:58-99`; `internal/compiler/core/core_test.go:170-208`; `internal/compiler/session/session_phase11_differential_test.go:90-145`  
**Apply to:** convergence, golden ledger, and semantic verification.

Use exact C for characterization/integrity, a checked provenance ledger for approved changes, and the four-tier comparator for semantics. These are separate controls.

### Explicit debt authority

**Sources:** `internal/compiler/session/session_test.go:2981-3014`; `.planning/milestones/M002-phases/10-trusted-interprocedural-oracle/PHASE-10-DEBT.md:384-403`  
**Apply to:** M004 cut and requirements amendment.

The generic debt parser verifies artifact shape; the phase must separately assert the substantive cut information and never silently introduce a third deferral.

## No Analog Found

| File | Role | Data Flow | Reason |
|---|---|---|---|
| `testdata/phase16/restrict_readonly_probe.c` / Linux CI harness, if created | test fixture/harness | file-I/O | Existing repository evidence covers emitted by-pointer golden C, not a checked-in hand-written, dual-host exact-shape C17 probe. Use the Phase-5 marker/golden test structure plus the research-defined C shape; do not pretend a generic native test proves it. |

## Metadata

**Analog search scope:** `internal/compiler/cgen`, `internal/compiler/core`, `internal/compiler/session`, `.planning` debt/requirements artifacts, `testdata`  
**Tracked files scanned:** 10 primary analogues  
**Pattern extraction date:** 2026-09-19
