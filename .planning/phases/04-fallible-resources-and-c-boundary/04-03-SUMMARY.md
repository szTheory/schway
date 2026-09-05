---
phase: 04-fallible-resources-and-c-boundary
plan: "03"
subsystem: compiler
tags: [ffi, c17, foreign-contract, conformance-unit, layout, evidence, mutation-testing]

requires:
  - phase: 04-fallible-resources-and-c-boundary
    provides: "plan 01/02's core.ForeignContract, checkFallibleLinear/checkResourceLifecycle, the byte-frozen native/lang_foreign_resource.c + private header, the release-order materialization/rederivation pair, verifyForeignCorpus's existing required controls"
provides:
  - "The complete core.ForeignContract field set: InitializedState, Capture, Retention, Aliasing, and Layout (RecordLayout/LayoutField), alongside the existing Symbol/Allocator/Unwind/NonlocalExit/Fails"
  - "core.LinearOperation.Allocator, the additive allocator-identity fact shared by OpForeignCall acquisitions and the OpRelease that discharges them"
  - "evidence.Manifest.ForeignDigest, digest-bound from a new lang.foreign/0 sidecar manifest, additive on the /1 identity struct only"
  - "cgen.EmitForeignManifest/EmitForeignHeader/EmitForeignConformance: the three inspectable layers generated from one contract"
  - "cgen.BannedOptimizerAttributes/NoreturnExemption/ScanForBannedAttributes: the zero-attribute control machinery"
  - "native.Runner.CompileConformanceUnit: a separate, bounded, timed, compile-only invocation reporting native.conformance_failed"
  - "session.LayoutMutationRunner/LayoutProbeContract and two new verifyForeignCorpus required controls: control:foreign.layout_mismatch, control:foreign.no_unproven_attributes"
affects: [04-04, 04-05, 04-06, 04-07]

actuals:
  tokens: 19100
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Obligations this phase's language cannot yet exercise (capture/retention/aliasing/callback_retention) are structurally fixed, compiler-derived facts (standardForeignObligations) rather than per-symbol declarations, and are explicitly named in the sidecar manifest's unchecked_obligations list rather than silently proven."
    - "The generated header, the sidecar manifest, and the conformance unit's obligation comment block are all derived from the SAME core.ForeignContract value, so a contract mutation moves all three in lockstep -- no independently hand-maintained comment exists to drift."
    - "The conformance unit's target-layout obligation (Layout.ForeignTypeName) describes the ACQUIRED RESOURCE's internal record (matching the frozen private header's own struct), not the ABI result struct cgen already emits inline -- the two are deliberately different C types serving different purposes."
    - "A dedicated two-field probe contract/fixture (LayoutProbeContract + foreign_layout_mismatch.golden.c) demonstrates the general N-field conformance mechanism without touching the byte-frozen, single-field production fixture, which cannot itself express a field transposition."

key-files:
  created:
    - internal/compiler/native/native_conformance_test.go
    - testdata/phase4/foreign_layout_mismatch.golden.c
  modified:
    - internal/compiler/core/core.go
    - internal/compiler/check/check.go
    - internal/compiler/check/check_test.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/corevalidate_test.go
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_test.go
    - internal/compiler/evidence/evidence.go
    - internal/compiler/evidence/evidence_test.go
    - internal/compiler/native/native.go
    - internal/compiler/native/foreign_resource.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_test.go

key-decisions:
  - "core.ForeignContract's Layout obligation describes the frozen private header's ACTUAL record (native/lang_foreign_resource_private.h's lang_foreign_resource_block, one field: payload) -- not the {ok, value} by-value ABI result struct cgen's emitLinearForeign already generates inline. These are two distinct, independently-declared C shapes serving different purposes (the acquired resource's internal layout vs. the call's wire result), and conflating them would have made the real frozen fixture's one-field record unable to host a meaningful field-order proof at all."
  - "InitializedState/Capture/Retention/Aliasing are compiler-derived (check.go's standardForeignObligations), not sourced from new `foreign C {}` policy syntax, because this plan's files_modified excludes the syntax/ast/lexer/parser layer and existing plan01/02 corpus fixtures cannot be edited to add new declaration keys without risking their already-shipped tests. This is documented as a fixed structural fact of this phase's language (no closures/threads/partial-init), not an omission-tolerant default -- a future phase with richer foreign shapes must replace it with real per-symbol declarations."
  - "A dedicated 2-field probe contract (session.LayoutProbeContract) and its own frozen fixture (foreign_layout_mismatch.golden.c) drive the layout-transposition mutation-kill demonstration, independent of the real single-field production symbol, which structurally cannot express a 2-field transposition."
  - "The generated `_LANG_` header and the conformance unit's own typedef/extern block are inlined directly into the conformance unit's returned string (EmitForeignConformance calls EmitForeignHeader and prepends its output) rather than written to a second physical file on disk -- native.Runner.CompileConformanceUnit compiles one self-contained temporary source per invocation, consistent with its own bounded/timed single-source-file contract."

patterns-established:
  - "Three-layer generation from one contract: EmitForeignManifest (JSON, authoritative), EmitForeignHeader (obligation comments + self-layout asserts, generated FROM the same contract value), EmitForeignConformance (the one place Lang's declaration and the foreign private header meet, compiled separately and never linked)."
  - "A fail-closed internal-consistency validator (corevalidate.foreignLayoutConsistent) and a fail-closed cross-artifact identity validator (releaseAllocatorMatches) both read only the core artifact, independently of check's own bookkeeping."

requirements-completed: [FFI-01]

coverage:
  - id: D1
    description: "core.ForeignContract carries every FFI-01 obligation category (layout, initialized state, allocator, capture, retention, aliasing, unwind) and is content-bound into evidence without moving any pre-existing manifest identifier"
    requirement: FFI-01
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestForeignContractCarriesEveryObligation"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestForeignFieldsAreOmittedWhenAbsent"
        status: pass
      - kind: unit
        ref: "internal/compiler/evidence/evidence_test.go#TestForeignSidecarManifestDigestBinds"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestPreviousPhaseManifestIDsUnchanged"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestAllocatorIdentityMismatchRejected"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestForeignContractInternallyValidated"
        status: pass
    human_judgment: false
  - id: D2
    description: "Three inspectable layers (sidecar manifest, generated header, conformance unit) are all generated from one contract, and the conformance unit compiles on its own, never linked, over the real frozen private header"
    requirement: FFI-01
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_test.go#TestGeneratedForeignHeaderNamesAreAllocated"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_test.go#TestObligationCommentsAreGeneratedFromJSON"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_test.go#TestConformanceUnitAssertsEveryField"
        status: pass
      - kind: unit
        ref: "internal/compiler/native/native_conformance_test.go#TestConformanceUnitCompilesSeparately"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_test.go#TestPrivateHeaderIsIncludedOnlyByConformanceUnit"
        status: pass
    human_judgment: false
  - id: D3
    description: "A field transposition in a frozen boundary fixture is a compile-time refusal, and zero optimizer-visible attributes are emitted, mutation-killed by injection"
    requirement: FFI-01
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestLayoutMutationIsCompileTimeRefusal"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestLayoutMutationAttacksFrozenFixtureOnly"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestNoUnprovenAttributesEmitted"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestAttributeInjectionMakesControlFail"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestNoreturnExemptionIsNamed"
        status: pass
      - kind: integration
        ref: "internal/compiler/session/session_test.go#TestVerifyPhase4ForeignLayoutControls"
        status: pass
    human_judgment: false

duration: ~2h (single continuous session)
completed: 2026-09-05
status: complete
---

# Phase 4 Plan 3: Fallible Resources and C Boundary — Three-Layer Foreign Contract Summary

**The complete FFI-01 obligation set lives in one authoritative `core.ForeignContract`, generates three lockstep-derived inspectable layers (sidecar manifest, `_LANG_` header, and a conformance translation unit compiled separately and never linked), and proves both its target-layout claim and its zero-optimizer-attribute claim with mutation-killed controls.**

## Performance

- **Duration:** ~2h (single continuous session)
- **Tasks:** 3/3 completed
- **Files modified:** 15 (2 created, 13 modified)

## Accomplishments

- `core.ForeignContract` gained `InitializedState`, `Capture`, `Retention`, `Aliasing`, and `Layout` (`*core.RecordLayout`, an ordered `LayoutField` list plus record-level size/alignment and the foreign-side type name), completing every obligation category FFI-01 names alongside the existing `Symbol`/`Allocator`/`Unwind`/`NonlocalExit`/`Fails`.
- Every new obligation is required and refused when absent (`corevalidate`'s `foreign.obligation_undeclared`/`foreign.layout_invalid`) — never a silently-defaulted omission. The three the language cannot yet exercise (capture, retention, aliasing — plus callback retention, which has no field at all) are named honestly in the `lang.foreign/0` sidecar manifest's `unchecked_obligations` list.
- `core.LinearOperation.Allocator` is a new additive fact carried on `OpForeignCall` and copied onto the `OpRelease` that discharges it; `corevalidate.releaseAllocatorMatches` independently refuses a release whose allocator disagrees, and `check.go`'s `releaseAllocatorMismatch` defensively re-asserts the same invariant on its own emission.
- `evidence.Manifest.ForeignDigest` is bound from a new `lang.foreign/0` sidecar manifest's content digest, set only when a program declares a foreign block; the `/1` identity struct's `omitempty` tag on the new trailing field keeps every pre-existing manifest identifier byte-for-byte unchanged (`TestPreviousPhaseManifestIDsUnchanged` still green).
- `cgen.EmitForeignManifest`/`EmitForeignHeader`/`EmitForeignConformance` are three new sibling emitters (never invoked from `Emit`/`EmitNative`) generating, respectively: the authoritative JSON sidecar; a `_LANG_`-namespaced header whose obligation comment block is written from the SAME contract value the manifest serializes (so a contract mutation moves both in lockstep); and a conformance translation unit carrying `_Static_assert(sizeof/_Alignof/offsetof)` triples against the real frozen private header, defining no symbol.
- `native.Runner.CompileConformanceUnit` compiles a conformance unit as its own separate, bounded, timed `-c`-only invocation, reporting the distinct `native.conformance_failed` code so a conformance refusal is never mistaken for a program compile failure.
- `cgen.BannedOptimizerAttributes`/`NoreturnExemption`/`ScanForBannedAttributes` back a real zero-attribute control: every corpus fixture's emitted C and sidecar manifest is scanned clean, and injecting a banned token into a copy of real emitted output is caught.
- `session.LayoutMutationRunner` (with a purpose-built `LayoutProbeContract`) and a new golden fixture (`testdata/phase4/foreign_layout_mismatch.golden.c`, a two-field record with its fields deliberately transposed) drive `control:foreign.layout_mismatch`; `control:foreign.no_unproven_attributes` joins it as a second new required control in `verifyForeignCorpus`.

## Task Commits

Each task was committed atomically:

1. **Task 04-03-01: Complete the authoritative foreign contract and bind its sidecar manifest into evidence** — `032be33` (feat)
2. **Task 04-03-02: Generate the three inspectable layers and compile the conformance unit separately** — `a045875` (feat)
3. **Task 04-03-03: The layout mutation against the frozen fixture and the zero-attribute control against the emitter** — `17e8f3b` (feat)

## Files Created/Modified

- `internal/compiler/core/core.go` — `ForeignContract` obligation fields, `RecordLayout`/`LayoutField`, `LinearOperation.Allocator`
- `internal/compiler/check/check.go` — `standardForeignLayout`/`standardForeignObligations`/`buildForeignContract`, `releaseAllocatorMismatch`
- `internal/compiler/corevalidate/corevalidate.go` — `foreignLayoutConsistent`, `releaseAllocatorMatches`, obligation-completeness check
- `internal/compiler/evidence/evidence.go` — `Manifest.ForeignDigest`, `hasForeignContract`, `/1`-identity-only trailing field
- `internal/compiler/cgen/cgen.go` — `EmitForeignManifest`/`EmitForeignHeader`/`EmitForeignConformance`, `BannedOptimizerAttributes`/`NoreturnExemption`/`ScanForBannedAttributes`
- `internal/compiler/native/native.go` — `Runner.CompileConformanceUnit`
- `internal/compiler/native/foreign_resource.go` — `ForeignResourcePrivateHeaderPath`
- `internal/compiler/session/session.go` — `LayoutMutationRunner`, `LayoutProbeContract`, two new `verifyForeignCorpus` lanes
- `testdata/phase4/foreign_layout_mismatch.golden.c` — the seeded two-field transposed private-header fixture
- Test files: `check_test.go`, `corevalidate_test.go`, `evidence_test.go`, `cgen_test.go`, `native_conformance_test.go` (new), `session_test.go`

## Decisions Made

See `key-decisions` in the frontmatter. The most consequential: the target-layout obligation describes the frozen private header's real internal record (`lang_foreign_resource_block`, one field), not the ABI result struct cgen already emits, and a dedicated two-field probe contract/fixture drives the transposition mutation-kill demonstration since the real production record structurally cannot express one.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Adding `ForeignDigest` to the `/1` identity struct moved every manifest ID, including pre-Phase-4 and non-foreign programs**
- **Found during:** Task 1, first run of `TestPreviousPhaseManifestIDsUnchanged`/`TestOwnedEvidenceBindings`
- **Issue:** `encoding/json` includes every struct field's key in the marshaled identity bytes regardless of value, so appending an untagged `ForeignDigest string` field changed the identity JSON — and therefore the SHA-256-derived manifest ID — for every program, not just ones declaring a foreign block.
- **Fix:** Tagged the new field `json:",omitempty"` so it is entirely absent from the identity JSON when empty, restoring byte-identical manifest IDs for every program with no foreign block.
- **Files modified:** internal/compiler/evidence/evidence.go
- **Verification:** `TestPreviousPhaseManifestIDsUnchanged` and `TestOwnedEvidenceBindings` both green
- **Committed in:** 032be33 (Task 1 commit)

**2. [Rule 1 - Bug] The new release-allocator lookup broke the pre-existing "invented" release-order mutation test**
- **Found during:** Task 1, first run of `TestReleaseOrderMutationMatrix/invented`
- **Issue:** My first `releaseAllocatorMatches` implementation failed with `core.release_target_unknown` whenever a release's `ReleasesOperationID` did not resolve to a real operation — but plan 02's own "invented" mutation deliberately sets an unresolvable ID and expects `core.release_order_mismatch` from `checkReleaseOrder`'s later sequence comparison, not an earlier failure from this new check.
- **Fix:** `releaseAllocatorMatches` now silently returns `true` (deferring to `checkReleaseOrder`'s own comparison) when the acquisition cannot be resolved by ID, only comparing allocators when a real acquisition is found.
- **Files modified:** internal/compiler/corevalidate/corevalidate.go
- **Verification:** `TestReleaseOrderMutationMatrix` (all four rows) and `TestAllocatorIdentityMismatchRejected` both green
- **Committed in:** 032be33 (Task 1 commit)

---

**Total deviations:** 2 auto-fixed (both Rule 1 — bugs found and fixed during implementation, before any commit landed). **Impact on plan:** Both auto-fixes were necessary for correctness against already-shipped Phase 4 plan 01/02 invariants; no scope creep.

## Issues Encountered

None.

## Revert-and-Fail Demonstrations (D-09)

### 1. Layout mutation (seeded, not a revert — the frozen fixture is committed already-wrong)

Ran `session.LayoutMutationRunner` against the committed `testdata/phase4/foreign_layout_mismatch.golden.c` directly:

```
DEMO OUTPUT:
native.conformance_failed: exit status 1: ~
error: static assertion failed due to requirement '__builtin_offsetof(struct lang_foreign_layout_probe_block, first) == 0':
lang_foreign_layout_probe_block.first offset must match the declared layout
   49 | _Static_assert(offsetof(lang_foreign_layout_probe_block, first) == 0, "lang_foreign_layout_probe_block.first offset must match the declared layout");
      |                ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
/Applications/Xcode.app/.../__stddef_offsetof.h:16:24: note: expanded from macro 'offsetof'
   16 | #define offsetof(t, d) __builtin_offsetof(t, d)
      |                        ^
~ note: expression evaluates to '1 == 0'
   49 | _Static_assert(offsetof(lang_foreign_layout_probe_block, first) == 0, ...);
      |                ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~^~~~
~ error: static assertion failed due to requirement
'__builtin_offsetof(struct lang_foreign_layout_probe_block, second) == 1':
lang_foreign_layout_probe_block.second offset must match the declared layout
   52 | _Static_assert(offsetof(lang_foreign_layout_probe_block, second) == 1, ...);
      |                ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
```

`err.(*native.ToolError).Code == "native.conformance_failed"` — the expected refusal, over the real `clang` toolchain.

### 2. Zero-attribute control (production-code injection, in the live working tree — file backed up, edited, tested, restored)

Backed up `internal/compiler/cgen/cgen.go`, injected a banned token (`restrict`) into `EmitForeignHeader`'s generated extern declaration line, ran the control:

```
--- FAIL: TestNoUnprovenAttributesEmitted (0.00s)
    session_test.go:1338: foreign_acquire_one.lang: found banned attribute tokens [restrict]
FAIL
FAIL	github.com/codename-lang/lang/internal/compiler/session	0.137s
```

Restored `cgen.go` from the backup (`diff` confirmed byte-identical to the pre-mutation state), re-ran the same command:

```
ok  	github.com/codename-lang/lang/internal/compiler/session	0.193s
```

Full `go test -count=1 ./...` and `go test -race ./...` were re-run clean after the restore, confirming no residual state from the injection.

## Verification Performed

- `env GOCACHE=/tmp/ai-lang-phase4-cache go test -count=1 ./...` — pass
- `go test -race ./...` — pass
- `go vet ./...` — clean
- `git diff <phase-start>..HEAD -- testdata/phase1 testdata/phase2 testdata/phase3` — empty
- `go.mod`/`go.sum` — no new requirement
- `sh scripts/verify-phase3.sh` — exits 0
- `native/lang_foreign_resource.c` and `native/lang_foreign_resource_private.h` — byte-identical to their state at the end of plan 01 (`git diff --stat` empty)
- All six named tests per task's `<verify>` block, run via `scripts/assert-go-tests.sh` — pass

## Known Stubs

None that block this plan's own success criteria. Documented narrowings (not stubs, stated in `key-decisions`): `InitializedState`/`Capture`/`Retention`/`Aliasing` are compiler-derived rather than per-symbol-declared this phase (no new `foreign C {}` syntax was added); `capture`/`retention`/`aliasing`/`callback_retention` are recorded in the sidecar manifest's `unchecked_obligations` list, never claimed as proven.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- The complete `core.ForeignContract`, the three inspectable layers, the conformance-unit compile machinery, and the two new required controls are all in place for plan 04 (the terminal `defect` operation) and later plans to extend directly.
- Carried-forward debt, unchanged from plans 01/02: `core.TerminatorKinds()` still returns `{OpReturn, OpFail}`; the defect terminator (`OpDefect`) is plan 04's job. `originvalidate`/`pathoracle` still only walk `OpReturn` — D-04-29's widening to `{OpReturn, OpFail, OpDefect}` remains explicitly out of scope for plans 01/02/03, carried to plan 06 per the phase's own source-coverage audit.
- New, plan-03-local narrowing (documented in `key-decisions`, not carried debt): the foreign obligation surface has no per-symbol declaration syntax for `initialized_state`/`capture`/`retention`/`aliasing` — a future phase adding richer foreign shapes needs real syntax here.
- No blockers.

## Self-Check: PASSED

All key files verified present on disk; all three task commits (`032be33`, `a045875`, `17e8f3b`) verified present in `git log`; full test suite, race detector, vet, and `verify-phase3.sh` all green as of this summary.

---
*Phase: 04-fallible-resources-and-c-boundary*
*Completed: 2026-09-05*
