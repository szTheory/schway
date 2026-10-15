# Phase 37: CRC-32 Consumer, Interpreter File Route, and Guide - Pattern Map

**Mapped:** 2026-10-15  
**Files analyzed:** 30 planned implementation, test, fixture, and evidence owners  
**Analogs found:** 30 / 30 (one partial; no complete production Go reader analog)

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| examples/crc32/lib/crc32.schway | utility | transform | examples/checksum/adler32.schway; examples/helper_calls/crc_step.schway | role-match |
| examples/crc32/crc32.schway | controller | file-I/O | examples/checksum/checksum.schway | exact |
| examples/crc32/crc_low.schway | controller | file-I/O | examples/checksum/adler_low.schway | exact |
| examples/crc32/adapter.c | utility / host adapter | file-I/O | examples/checksum/adapter.c | exact copy |
| examples/crc32/adapter.h | config / ABI declaration | transform | examples/checksum/adapter.h | exact copy |
| examples/crc32/file_bytes.bindings.json | config | transform | examples/checksum/file_bytes.bindings.json | exact copy |
| examples/crc32/crc32.expected.json (proposed) | config / test fixture | batch | examples/checksum/checksum.expected.json | role-match |
| examples/crc32/README.md | utility / executable guide | file-I/O | examples/checksum/README.md | role-match |
| internal/compiler/testsupport/crc32_oracle.go (proposed) | utility / oracle | transform | internal/compiler/testsupport/checksum_oracle.go | role-match |
| internal/compiler/testsupport/crc32_oracle_test.go (proposed) | test | transform | internal/compiler/testsupport/checksum_oracle_test.go | role-match |
| internal/compiler/testsupport/crc32_matrix_test.go (proposed) | test | request-response | internal/compiler/native/native_phase31_matrix_test.go | role-match |
| internal/compiler/interp/filebytes_reader.go (proposed) | utility / host adapter | file-I/O | internal/compiler/native/bindings.go; examples/checksum/adapter.c | partial |
| internal/compiler/interp/filebytes_reader_test.go (proposed) | test | file-I/O | internal/compiler/native/native_phase31_fault_test.go | role-match |
| internal/compiler/interp/interp.go | utility / interpreter | transform | RunWithFileBytes and RunApplication in same file | exact substrate |
| internal/compiler/interp/interp_phase37_test.go (proposed) | test | transform | internal/compiler/interp/interp_phase36_test.go; interp_phase321_defect_test.go | role-match |
| internal/compiler/session/session.go | service | request-response | RunInterpreterApplication, BuildApplicationFile, RunNativeCommandFile | role-match |
| internal/compiler/session/session_phase37_crc32_test.go (proposed) | test | request-response | session_phase31_single_copy_test.go; pathoracle_phase31_filebytes_test.go | role-match |
| cmd/schway/main.go | controller | request-response | runApplicationRun and existing run dispatch | exact |
| cmd/schway/main_phase37_file_run_test.go (proposed) | test | request-response | cmd/schway/main_test.go; internal/compiler/session/session_phase321_run_input_test.go | role-match |
| cmd/schway/main_phase32_routes_test.go | test / route inventory | request-response | same file | exact |
| cmd/schway/main_phase34_routes_test.go | test / diagnostic inventory | request-response | same file | exact |
| cmd/schway/main_phase37_guide_test.go (proposed) | test / guide runner | file-I/O | cmd/schway/main_phase32_guide_test.go | partial |
| README.md | utility / exit contract | request-response | existing Exit codes section | exact |
| internal/compiler/session/session_phase28_exitdocs_test.go | test / exit inventory | request-response | same file | exact |
| internal/compiler/surfacemanifest/manifest.go (conditional) | utility / public facts | transform | Current and canonical in surfacemanifest | role-match |
| testdata/phase34/surface-witnesses.json (conditional) | config / test fixture | batch | same file | exact |
| internal/compiler/session/session_phase34_limit_consumers_test.go (conditional) | test / consumer inventory | request-response | same file | exact |
| .planning/phases/37-.../37-prohibition-check.test.cjs (proposed) | test / evidence adapter | request-response | .planning/phases/36-scalar-helper-calls/36-prohibition-check.test.cjs | role-match |
| .planning/phases/37-.../prohibition-*.json (proposed) | config / closed fixtures | batch | tracked Phase 36 prohibition fixtures | role-match |
| .planning/phases/37-.../37-PROHIBITION-EVIDENCE.md (proposed) | utility / evidence record | transform | .planning/phases/36-scalar-helper-calls/36-PROHIBITION-EVIDENCE.md | role-match |

Test and phase-local evidence filenames are proposals, not locked API names. Surface-manifest files are conditional: D-12 requires route-specific truth only if current scoped facts cannot express the input route without changing the ten-field wire shape or claiming a universal FileBytes cap. No new exit-code value is implied.

## Pattern Assignments

### Shared CRC source and consumers

**Analogs:** examples/checksum/adler32.schway, adler_low.schway, checksum.schway; examples/helper_calls/crc_step.schway.

Use the Adler module for a pure ByteView loop and the helper witness for right-shift/XOR/if forms. Keep all CRC work in the new library. Phase 36 crc_step is a four-step helper witness, not a second full CRC implementation.

~~~schway
// examples/checksum/adler32.schway:8-22
fn adler32(bytes: ByteView) -> U64 {
  var a = 1
  var b = 0
  var i = 0
  var n = byte_len(bytes)
  while i < n {
    a = a + byte_at(bytes, i)
    a = a % 65521
    b = b + a
    b = b % 65521
    i = i + 1
  }
  var r = b << 16
  r = r + a
  r
}
~~~

~~~schway
// examples/helper_calls/crc_step.schway:8-15
fn crc_step(state: U64) -> U64 {
  var next = state >> 1
  if state & 1 == 1 {
    next = next ^ 0xEDB88320
  } else {
    next = state >> 1
  }
  next
}
~~~

Both entries follow checksum.schway's exact foreign declarations and FileError alternatives, acquire PathToken, borrow ByteView, call the shared module and return U64. adler_low.schway lines 45-56 is the low-16 reduction pattern. Imported foreign/data declarations remain unsupported, so repeat them in both entries. These excerpts establish admitted source forms only; do not copy Adler's modulus.

~~~schway
// examples/checksum/checksum.schway:13-49
foreign C {
  fn schway_file_bytes_acquire(path: PathToken) -> FileBytes {
    unwind: forbidden
    nonlocal_exit: forbidden
    mode: "acquire"
    abi: schway_file_bytes_acquire_fn
    allocator: "libc_malloc"
    release: "schway_file_bytes_release"
    fails: FileError
  }
}
data FileError =
  | FileTooLarge
  | FileUnavailable
  | AllocFailed
fn main(path: PathToken) -> U64 {
  let owner = try schway_file_bytes_acquire(path)
  let view = schway_file_bytes_view(owner)
  let sum = adler32(view)
  sum
}
~~~

### CRC adapter triple

**Analog:** examples/checksum/adapter.c, adapter.h, file_bytes.bindings.json.

Copy the Phase 31 three-file contract without semantic edits and check byte drift. C opens once with O_NONBLOCK and supported O_NOCTTY/O_CLOEXEC flags, fstats the descriptor, reads limit+1, rejects oversize before shrink, checks successful close, and emits no streams. adapter.h lines 7-45 pins ABI layout/types; the manifest is the closed local source/header/include/symbol list.

~~~c
// examples/checksum/adapter.c:76-108
descriptor = SCHWAY_FILE_BYTES_OPEN(
    path, O_RDONLY | O_NONBLOCK | SCHWAY_FILE_BYTES_NOCTTY | SCHWAY_FILE_BYTES_CLOEXEC);
if (descriptor < 0) return acquire_failure(SCHWAY_FILE_BYTES_UNAVAILABLE);
if (SCHWAY_FILE_BYTES_FSTAT(descriptor, &details) != 0 || !S_ISREG(details.st_mode)) {
  return acquire_cleanup_failure(descriptor, NULL, SCHWAY_FILE_BYTES_UNAVAILABLE);
}
while (received < (size_t)SCHWAY_FILE_BYTES_LIMIT + 1u) {
  ssize_t amount = SCHWAY_FILE_BYTES_READ(
      descriptor, data + received, (size_t)SCHWAY_FILE_BYTES_LIMIT + 1u - received);
  if (amount > 0) {
    received += (size_t)amount;
    continue;
  }
  if (amount == 0) break;
  if (errno == EINTR) continue;
  return acquire_cleanup_failure(descriptor, data, SCHWAY_FILE_BYTES_UNAVAILABLE);
}
if (received > (size_t)SCHWAY_FILE_BYTES_LIMIT) {
  return acquire_cleanup_failure(descriptor, data, SCHWAY_FILE_BYTES_TOO_LARGE);
}
~~~

The adapter's successful close is checked before ownership is returned (lines 101-113). The fixed ordered alternatives are also enforced at cgen_program.go lines 3316-3320.

### Source admission and interpreter application

**Analogs:** internal/compiler/session/session.go, session_admit.go, session_modules.go, session_filebytes.go; internal/compiler/interp/interp.go; internal/compiler/pathoracle/pathoracle.go.

Load imports through the common source-path helper. Run the peer sequence on the whole checked merged program before resolveEntry prunes it. Then enforce the exact owner path and foreign contract. Existing helpers: session_modules.go lines 24-49 and session_filebytes.go lines 8-17.

~~~go
// session_admit.go:24-25, 33-55
// Admission runs on the FULL checked program, before resolveEntry prunes.
func admitProgram(checked CheckResult) (core.Program, []diagnostic.Diagnostic) {
    program := checked.Program
    validated := corevalidate.Validate(program)
    if !validated.Valid {
        return program, []diagnostic.Diagnostic{peerRefusalDiagnostic(validated, program.ModuleID)}
    }
    program = validated.Program()
    if problems := originvalidate.ValidatePublished(program); len(problems) > 0 {
        return program, []diagnostic.Diagnostic{originPeerRefusalDiagnosticIn(problems[0], program, checked.Tree, checked.UnitTrees)}
    }
    if err := pathoracle.ValidateLocalOwnerPaths(program); err != nil {
        return program, []diagnostic.Diagnostic{diagnostic.Error("core.pathoracle_refused", diagnostic.Span{}, err.Error())}
    }
    return program, nil
}
~~~

RunInterpreterApplication at session.go lines 1184-1210 has the checked/admitted/entry sequence but accepts U64 only. RunWithFileBytes at interp.go lines 601-657 is an evidence-oriented injected-reader API; retain that seam and add a public application boundary with the fixed production reader. Interpreter admission does not resolve a manifest; native alone requires the explicit manifest.

The frame-stack engine and output sink are the reuse point. Current RunApplication installs the sink and creates evidence for scalar calls:

~~~go
// interp.go:1128-1138
base.applicationOutput = sink
if hasScalarValueCalls(program) {
    base.evidence = newInterpreterEvidence()
}
execution, err := runProgramFrameStack(program, base)
capacityErr := errors.Is(err, ErrEvidenceCapacityExhausted)
~~~

runProgramFrameStack auto-creates evidence (interp.go lines 1695-1713); frames carry invocation identity, call ordinals and the evidence pointer (1362-1378). Add a capture-disabled ordinary application mode that avoids retaining events, identities and growing counters. Preserve the enabled path. TestPhase36InterpreterEvidenceBoundaries (interp_phase36_test.go:302-388) pins output and real terminal outcomes through exhaustion, including a later defect. TestPhase321AbortLikeDefectFileRoute pins file-owner defect and normal-return behavior.

### Independent production Go reader

**Partial analogs:** internal/compiler/native/bindings.go and examples/checksum/adapter.c.

native.readBindingFile (bindings.go:66-86) opens one descriptor, checks regular-file type, and reads through io.LimitReader. It ignores deferred close errors and lacks nonblocking open. The C adapter supplies the stronger descriptor/FIFO and close pattern, but no production Go reader meets D-11. testsupport.ChecksumReadFile in checksum_oracle.go:29-58 is the independent test reader/oracle and must not be imported into production.

### Opaque CLI route and native one-shot run

**Analogs:** cmd/schway/main.go, main_test.go, internal/compiler/native/native_app.go.

Dispatch the separator-bearing form before generic extractors. Parse only pre-separator options, pass one opaque suffix token, and keep existing no-tail run --engine=... SOURCE as conformance. runApplicationRun at main.go lines 211-264 is the parser precedent.

~~~go
// main.go:36-66, 120-130
if len(args) > 0 && args[0] == "app" {
    return runApplication(args)
}
args, jsonMode, ok := extractJSON(args)
// ...
if len(args) == 3 && args[0] == "run" && strings.HasPrefix(args[1], "--engine=") {
    // Existing no-tail route
}
~~~

TestPhase22AppRunKeepsOpaqueTokenStreamsAndChildStatus (main_test.go:526-532) pins --json as opaque child input. main_test.go:328-345 pins empty input as (65, "", ""), plus the 4,096/4,097-byte and NUL boundaries. Preserve these tuples for both file routes.

No combined source-build/run route exists. Compose session.BuildApplicationFile with a private temporary artifact, then native.RunApplication once; preserve direct streams, direct argv, timeout, and cleanup. native_app.go lines 143-218 shows binding resolution/private Clang staging; lines 400-455 shows bounded child execution/outcome mapping.

### Independent oracle, tuple matrix, and reached mutations

**Analogs:** testsupport/checksum_oracle.go, checksum_oracle_test.go, checksum_matrix_test.go; native/native_phase31_matrix_test.go; session/session_phase31_single_copy_test.go; native/native_phase31_controls_test.go and native_phase31_fault_test.go; pathoracle/pathoracle_phase31_filebytes_test.go.

ChecksumOracleRows (checksum_oracle.go:60-103) records Python command/provenance and literal values; checksum_oracle_test.go:13-46 cross-checks with Go's independent library. checksum.expected.json:1-24 shows case tuples with exact streams and engine ownership. Cross-check recurring values with Go hash/crc32.ChecksumIEEE, never as the sole oracle.

TestPhase31CaseMatrix (native_phase31_matrix_test.go:20-40, 52-110) compares each engine to literal tuples with exact streams and a timeout. Phase 37 should invoke both public CLI forms, compare each to expected tuples, then compare engines. Keep root-sensitive unreadable testing a named case; do not skip the matrix.

TestPhase31SingleCopyAndDualFlip (session_phase31_single_copy_test.go:113-160, 232-292) combines a scoped scan, one imported function per merged program, both consumers' changed outputs, and changed module digests/build identities. TestPhase31Controls (native_phase31_controls_test.go:153-235) is the finite reached-mutation pattern. TestPhase31BoundedReadMutants and TestPhase31ReleaseExactlyOnce (native_phase31_fault_test.go:164-190, 272-369) cover exact mutation anchors, bounds, FIFO refusal, and physical cleanup. TestPhase31FileBytesSeededFaults (pathoracle_phase31_filebytes_test.go:17-73, 138-160) deep-copies checked core and refuses forged owner shapes.

Proposed focused test owners: testsupport/crc32_oracle.go and _test.go, testsupport/crc32_matrix_test.go, interp/filebytes_reader_test.go, interp/interp_phase37_test.go, and session/session_phase37_crc32_test.go. Keep the five CRC mutations finite and reached: each must check/admit/build and fail at its discriminating assertion. Restore source/overlays and compare raw bytes.

### Route, exit, surface, guide, and prohibition inventories

**Analogs:** cmd/schway/main_phase32_routes_test.go, main_phase34_routes_test.go, main_phase32_guide_test.go; session_phase28_exitdocs_test.go and session_phase34_limit_consumers_test.go; surfacemanifest/manifest.go; tracked Phase 36 prohibition artifacts.

Add the new route and both entries to current inventories with a registered owning test. Exact current names: TestPhase32PublicRouteAudit, TestPhase34DiagnosticRouteInventory, and TestPhase34SourceRoutePresentation. The Phase 32 audit parses dispatch and requires rows; Phase 34 also binds handler/call/module/source class/owner.

Preserve the existing exit table: Invalid 2, Operational 3, Mismatch 4, Usage 64, AppInputRefused 65, EventCapacity 69, AppTimeout 124, AppLaunchError 125. Root README lines 71-101 documents meanings; TestExitCodeDocsTable (session_phase28_exitdocs_test.go:33-92) catches missing, duplicate and wrong rows. There is no new exit code.

surfacemanifest.Current (manifest.go:105-137) states FileBytes has no universal language cap. Update the producer, witness and TestPhase34SurfaceLimitConsumers only if D-12 needs an interpreter-specific limit/input-form fact; preserve the independent finite census and ten-field wire shape.

Preserve TestPhase32GuideCleanCheckout, TestPhase32GuideMatchesExpectedJSON and TestPhase32GuideParserNegativeControls. The Phase 32 copy helper at main_phase32_guide_test.go:202-289 walks the checkout, including untracked nonignored files; do not reuse unchanged. Add a tracked-files-only copy path, exact command tuple checks, fixture/table agreement, and armed guide drift.

For D-23, use tracked Phase 36 artifacts 36-prohibition-check.test.cjs and 36-PROHIBITION-EVIDENCE.md. Keep the Phase 37 adapter and fixtures local and dependency-free; bind exact source/root selectors, accept only the named RED, and run bad/clean/default through the canonical producer. Preserve unresolved history and separate structural from numeric claims.

## Shared Patterns

**Admission and authority.** Apply module-aware check, corevalidate, originvalidate and pathoracle to the full program before pruning. Then enforce exact foreign contracts and owner shape. Native alone resolves the manifest. Sources: session_admit.go:24-55; pathoracle.go:1633-1665.

**Outcomes and streams.** Keep exit, stdout, stderr, typed FileError, checked defect, cleanup and evidence status distinct. File errors use empty stdout, one fixed line, and no private path. Empty path stays (65, "", ""). Sources: main.go:211-280; native_app.go:400-455; checksum_oracle.go:126-163.

**FileBytes ownership.** Keep ordered FileError alternatives and exact acquire/view/release ABI. PathToken entry: one acquire, one view, one consuming call, one release, returned U64. Scope 65,536 bytes to CRC adapters; native FileBytes remains binding-dependent. Sources: checksum.schway:13-49; cgen_program.go:3215-3250, 3316-3320; adapter.c:76-113.

**Independent evidence.** Compare both engines separately to literal expectations before parity. Keep numerical oracle, source admission, causal evidence and native physical cleanup as separate claims. Sources: checksum_oracle.go:60-103; native_phase31_matrix_test.go:52-110; session_phase31_single_copy_test.go:242-292.

## No Analog Found

| File/behavior | Role | Data Flow | Reason |
|---|---|---|---|
| Exact production Go reader combining nonblocking open, descriptor fstat, cap+1 read, checked close, and primary-error-preserving cleanup | utility / host adapter | file-I/O | native.readBindingFile is partial; testsupport.ChecksumReadFile is a test oracle and must not be imported. Implement the fixed Go adapter independently, using the C adapter only as a host-I/O reference. |

## Metadata

**Analog search scope:** examples/checksum, examples/helper_calls, cmd/schway, internal/compiler/session, interp, native, testsupport, pathoracle, surfacemanifest, exitcode, testdata/phase34, and tracked Phase 36 prohibition artifacts.  
**Tracked-source gate:** every named analog was verified with git ls-files --; no ignored runtime mirror or untracked path is referenced.  
**Pattern extraction date:** 2026-10-15

