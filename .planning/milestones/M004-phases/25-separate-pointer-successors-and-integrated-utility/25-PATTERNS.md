# Phase 25: Separate Pointer Successors and Integrated Utility - Pattern Map

**Mapped:** 2026-10-01  
**Files analyzed:** 16 planned or likely touched path groups  
**Analogs found:** 16 / 16 role-level analogs; no exact production U64 by-pointer admission exists

The exact names for Phase 25 fixtures and the evidence script remain planning choices. The rows below distinguish those new paths from existing files whose bounded patterns are the closest analogs. All named existing analogs were checked with git ls-files and are tracked source. The phase research ran no builds, tests, or CI; none were run during this mapping.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| internal/compiler/syntax/parser.go, only if syntax must change | utility | transform | existing borrow forms in testdata/phase3/sequential_shared_then_exclusive_accept.schway | role-match; prefer no syntax change |
| internal/compiler/check/check.go | utility | transform / validation | its checkLocalFileByteTransferCaller and existing loan conflict diagnostics | exact role, same bounded source admission |
| internal/compiler/corevalidate/corevalidate.go | utility | transform / validation | buildLoanChainIndex and independently derived loan endpoints | exact role-match |
| internal/compiler/originvalidate/originvalidate.go | utility | transform / validation | backward return-origin walk and borrow access derivation | exact role-match |
| internal/compiler/pathoracle/pathoracle.go | utility | transform / validation | RecomputeEndpoints and linearized path loan state | exact role-match |
| internal/compiler/interp/interp.go | utility | request-response / model execution | OpForeignCall outcome and OpCall frame handling | exact role-match |
| internal/compiler/cgen/cgen.go | utility | transform | legacy shared/exclusive family classifiers and ABI manifest derivation | role-match only; restrict is not a Phase 25 contract |
| internal/compiler/cgen/cgen_program.go | utility | transform | sole production serializer and bounded Phase 24 application admission | exact role, but currently refuses pointer bodies |
| internal/compiler/check/*_test.go and independent validator tests | test | transform / negative controls | internal/compiler/session/session_borrow_conflict_test.go | exact role-match |
| internal/compiler/cgen/cgen_program_test.go | test | transform / native-output controls | Phase 24 emitter transfer and pre-serialization refusal tests | exact role-match |
| internal/compiler/interp and session integration tests | test | request-response / model execution | internal/compiler/interp/interp.go and internal/compiler/session/session_borrow_conflict_test.go | exact role-match |
| testdata/phase25/*.schway (new; exact fixture names for planning) | test | transform / validation | phase3 sequential/conflicting loan fixtures and phase5 restrict_borrow.schway | role-match; no family-specific U64 witness exists |
| examples/phase24/transfer.schway or a new examples/phase25 utility source | utility | file-I/O / request-response | examples/phase24/transfer.schway | exact application flow predecessor |
| examples/phase24/README.md or a new Phase 25 README/evidence index | utility | request-response / documentation | examples/phase24/README.md | exact documentation pattern |
| scripts/verify-phase25.sh (new, if split from existing verifier) | utility | batch / evidence | scripts/verify-phase24.sh | exact evidence-script pattern |
| .github/workflows/ci.yml | config | batch / evidence | current macOS/Linux evidence-aggregate job | exact host-lane pattern |

## Pattern Assignments

### internal/compiler/check/check.go (checker, transform / validation)

**Analog:** internal/compiler/check/check.go, checkLocalFileByteTransferCaller, lines 4039-4070.

This function is the closest admission pattern because it recognizes source AST shape, validates foreign contracts, then constructs explicit core places and operations. The key fixed-shape guard and refusal are:

~~~go
if body == nil || len(body.Bindings) != 2 || body.Result != body.Bindings[1].Name || ... {
    return refuse("check.local_owner_shape_unsupported", "PathToken caller may only receive one owner helper, borrow it once, and return U64", body.Span)
}
~~~

The body is then represented as ordered core operations, with the release paired to the acquisition operation and return last (lines 4064-4070):

~~~go
ops := []core.LinearOperation{
    {Kind: core.OpCall, SourceID: places[0].ID, TargetID: places[1].ID, CalleeID: callee.ID},
    {Kind: core.OpForeignCall, SourceID: places[1].ID, TargetID: places[2].ID, Foreign: localForeignContract(use)},
    {Kind: core.OpRelease, SourceID: places[1].ID, ReleasesOperationID: acquisitionID, Foreign: localForeignContract(release)},
    {Kind: core.OpReturn, SourceID: places[2].ID, TypeID: types[2].ID},
}
~~~

Extend this bounded recognizer only for the approved acquire → use → shared copy → exclusive copy → return composition. Preserve the typed-use failure and release pairing. Do not assume relaxing the binding count is sufficient: the checker currently hard-codes the operation order and indexes those two bindings directly.

### Borrow source witnesses and conflict tests

**Analogs:** testdata/phase3/sequential_shared_then_exclusive_accept.schway, lines 11-18; testdata/phase3/shared_exclusive_reject.schway, lines 10-14; internal/compiler/session/session_borrow_conflict_test.go, lines 72-130.

The accepted fixture makes the shared loan's last use precede the exclusive borrow:

~~~schway
fn relay(buffer: Buffer) -> Buffer {
  let first = borrow buffer
  let reviewed = borrow first
  let second = borrow mut buffer
  let inspected = borrow second
  let delivered = take buffer
  delivered
}
~~~

The reject fixture keeps the first shared loan live across the exclusive borrow (lines 10-14). The session test distinguishes sequential acceptance from the shared/exclusive conflict and checks the stable diagnostic code and cause order. Use this shape to keep family positives and overlap failures separate. Phase 25 also needs independent escape and family-specific witnesses; these fixtures do not establish pointer lowering or a U64 helper contract.

### Independent borrow validators and model replay

**Analogs:** internal/compiler/corevalidate/corevalidate.go, lines 2042-2064; internal/compiler/originvalidate/originvalidate.go, lines 292-304; internal/compiler/pathoracle/pathoracle.go, lines 885-929.

Core validation builds its own parent/loan index from operations and peer-derived callee facts:

~~~go
for _, operation := range operations {
    if operation.Kind == core.OpBorrowShared || operation.Kind == core.OpBorrowExclusive {
        idx.bornAt[operation.TargetID] = operation.LoanID
    }
    if operation.TargetID != "" {
        idx.parent[operation.TargetID] = operation.SourceID
    }
}
~~~

Origin validation records access mode while walking the source chain (lines 297-304): `OpBorrowExclusive` derives exclusive access and `OpBorrowShared` derives shared access. Pathoracle independently enumerates paths, linearizes each, and synthesizes loan endpoints (lines 902-929). Keep these as separate derivations in their own packages; do not make any peer import checker approval or a shared Phase 25 verdict. Extend each peer only with the facts it already owns and retain its current fail-closed handling of unknown/corrupt core.

The interpreter counterpart is internal/compiler/interp/interp.go, lines 1239-1245, 1307-1343, and 1396-1411. It models borrow operations as value aliases, foreign outcomes separately, and in-language calls with callee frames. Preserve deterministic scalar outputs and the error-before-helper sequence there. Model replay does not prove host file IO, C pointer arguments, or physical cleanup, so keep its evidence label distinct from native application receipts.

### C lowering, ABI manifest, and emitter controls

**Analogs:** internal/compiler/cgen/cgen_program.go, lines 641-643, 1937-1970; internal/compiler/cgen/cgen.go, lines 253-255, 353-355, 381-397, 1080-1099.

The production whole-program emitter currently stops before serialization for either legacy by-pointer family:

~~~go
if function.Match == nil && (selectsByPointerLowering(function, function.Linear) || selectsByPointerLoweringSharedOnly(function, function.Linear)) {
    return "", fmt.Errorf("function %q: by-pointer bodies are not supported by whole-program native emission this phase", function.ID)
}
~~~

The legacy classifiers recognize a first shared or exclusive borrow from the sole parameter, followed by a narrow reborrow chain. The legacy exclusive attribute helper can produce `restrict`; that is historical metadata only and must not be copied into Phase 25 C or manifests. The current generic operation emitter writes borrow values as scalar copies and calls a callee with the source local (cgen_program.go, lines 1937-1970). A successful Phase 25 witness therefore needs a checked family-specific C pointer parameter and compatible address-taking at the call site, not just removal of the refusal.

Use one checked ABI fact source for the C declaration/call and the attribute manifest. The Phase 25 manifest and generated C must agree on pointer representation while making no `restrict`, `noalias`, capture, alignment, or ownership claim. Preserve structural refusals for mutation, forwarding, retention, callbacks, nonlocal exits, and wider pointer forms before serialization.

Emitter controls live in internal/compiler/cgen/cgen_program_test.go, lines 685-717 and 1721-1753. The legacy pointer disposition test asserts refusal occurs before invocation serialization; the Phase 24 positive transfer test checks ordered acquire/use/release/return operations and one acquisition-paired release. Follow both patterns: family-specific positive tests must inspect actual C parameters and manifests, while unsupported shapes and true conflicts remain distinct refusals. Add reached wrong-result controls so each helper independently contributes to the 65/66 result.

### Integrated utility and native evidence

**Analogs:** examples/phase24/transfer.schway, lines 7-50; examples/phase24/README.md, lines 12-35; examples/phase23/file_byte.bindings.json, lines 1-12; examples/phase23/adapter.h, lines 42-48.

The source declares explicit acquire/use/release foreign contracts, wraps acquisition in an ordinary helper, then uses the owner and returns a U64. The README gives root-relative Go build, Schway app build with the explicit manifest, supplied one-byte inputs, expected 65/66 results, and typed 0x43 failure. The Phase 23 manifest enumerates trusted C sources, headers, include directory, symbols, and runtime dependency. Keep the adapter as owner of its file descriptor and the Schway resource contract as owner of the allocated buffer lifecycle; the new pointer helpers consume the copied U64.

The exclusive `touch` precedent in testdata/phase5/restrict_borrow.schway, lines 26-31, is only a legacy lowering shape. It returns an owned Buffer through an exclusive reborrow and is not evidence for a scalar read/copy helper or a Phase 25 `restrict` promise.

For evidence, scripts/verify-phase24.sh, lines 11-43 and 45-90, records host OS/architecture, Go and Clang versions, Clang target, git revision, tree state, elapsed time, and rejects skipped/missing focused test groups. .github/workflows/ci.yml, lines 48-71 and 79-101, already owns macOS and Ubuntu lanes and runs the Phase 24 and Phase 23 evidence scripts under its evidence aggregate. Extend that ownership for Phase 25; give each family × host × applicable lane an explicit receipt and leave absent evidence incomplete. Do not add a duplicate full-suite host job when the existing owner answers the same question.

## Shared Patterns

### Independent derivation

**Sources:** check/check.go, corevalidate/corevalidate.go, originvalidate/originvalidate.go, and pathoracle/pathoracle.go.

The checker builds core from source; each peer independently re-derives relevant loan, origin, or endpoint facts. Keep distinct implementations and tests. The planner should assign family-specific overlap and escape mutations to every peer that consumes those facts.

### Structured diagnostics

**Source:** internal/compiler/diagnostic/diagnostic.go, lines 97-117; borrow cause ordering in internal/compiler/session/session_borrow_conflict_test.go, lines 118-130.

`Diagnostic` already carries schema, stable ID, code, severity, primary span, message, causes, and repairs; `diagnostic.Error` creates a stable structured error. Keep the schema and attach the primary span to the offending use/argument, with origin/conflict causes where known. Existing borrow tests expect repair-bearing conflict diagnostics, including `narrow_to_shared_borrow`; do not carry those repair suggestions into a Phase 25 boundary unless the particular edit is proven behavior-preserving. A true overlap remains a borrow conflict; a safe unsupported shape remains an unsupported-shape refusal.

### Evidence boundaries and dependency posture

Reuse Go standard-library test/evidence helpers, the Phase 23 C adapter, installed Clang, and current host lanes. Keep interpreter replay, actual host IO, pointer-parameter inspection, and the physical owner observer as separate evidence claims. No external package or new runtime/build dependency has a demonstrated need in the research.

## No Analog Found

| File/behavior | Role | Data Flow | Reason |
|---------------|------|-----------|--------|
| Phase 25 shared U64 read/copy helper with actual pointer-parameter C and a checked source caller | utility | transform / request-response | Existing shared classifier and historical golden do not pass production whole-program emission and do not prove U64 address-taking or the manifest contract. |
| Phase 25 exclusive U64 read/copy helper with actual pointer-parameter C and no unsupported attributes | utility | transform / request-response | Existing exclusive classifier/fixture is legacy and can imply `restrict`; it is not a safe production ABI precedent. |
| Separate family × host native pointer receipts with reached wrong-result controls | test | batch / native execution | Phase 24 receipts prove owner transfer/cleanup and cannot establish scalar pointer behavior. |

## Metadata

**Analog search scope:** examples/phase23, examples/phase24, internal/compiler/check, corevalidate, originvalidate, pathoracle, interp, cgen, diagnostic, session, native, testdata/phase3, testdata/phase5, scripts, and .github/workflows.  
**Tracked-source gate:** all existing analog paths named above were verified with git ls-files; no runtime mirror paths are used.  
**Project skills:** .codex/skills and .agents/skills are absent in this checkout.  
**Pattern extraction date:** 2026-10-01
