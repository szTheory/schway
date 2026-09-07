---
phase: 05-native-equivalence-and-adversarial-evidence
plan: 04
subsystem: compiler-codegen
tags: [cgen, check, corevalidate, restrict, alias-analysis, ownership, ffi-security]

requires:
  - phase: 05-native-equivalence-and-adversarial-evidence
    provides: "05-01's selectsByPointerLowering/emitLinearBorrowedByPointer by-pointer emitter and the Phase 1-4 byte-identity tripwire; 05-02's loanLivenessFixpoint as the sole liveness law (core.LoanEndpoint load-bearing, discoverLoanLastUses retired)"
provides:
  - "check.AliasFact/deriveAliasFacts: an admission-gating fact proving a straight-line function's sole parameter is exclusively borrowed for the WHOLE call, grounded in loanLivenessFixpoint's own materialized endpoints, exposed on check.Result with no core schema change"
  - "cgen emits C `restrict` on exactly the by-pointer parameter selectsByPointerLowering already selects, with a justification binding (attr/core_node/parameter/justified_by) populated in the lang.foreign/0 sidecar's emitted_attributes for a non-foreign, by-pointer-lowered function (singleManifestFunction)"
  - "corevalidate.ValidateEmittedAttributes: an independent re-derivation (via the existing recomputeLoanEndpoints machinery, never check's fixpoint) that refuses any emitted attribute with no proven, independently-matching justification (core.attribute_unjustified)"
  - "control:foreign.no_unproven_attributes narrowed, not deleted: BannedOptimizerAttributes/ScanForBannedAttributes unchanged (restrict still banned at every foreign extern declaration); JustifiableAttributes/ScanForUnjustifiedAttributes govern manifest entries only"
affects: [05-05, 05-06, 05-07]

actuals:
  tokens: 12631
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Three independent derivations of the same structural condition (check's deriveAliasFacts, cgen's selectsByPointerLowering, corevalidate's recomputeAliasJustifications), proven to agree by a cross-package corpus-enumeration test rather than by sharing a helper (D-12)"
    - "A post-hoc re-derivation (aliasFactEndpoints in check.go, the synthetic single-block CFG fed to corevalidate's recomputeLoanEndpoints) that reuses an EXISTING liveness/endpoint mechanism on real operation IDs after admission, rather than adding a second admission-deciding law"
    - "An optimizer-attribute exemption is legal only with an independently re-derived justification carried alongside it (JustifiableAttributes + ScanForUnjustifiedAttributes on the producer side, ValidateEmittedAttributes on the independent-validator side) -- narrower than NoreturnExemption's unconditional exemption, and the second precedent for 'named exemption, not silent omission'"

key-files:
  created: []
  modified:
    - internal/compiler/check/check.go
    - internal/compiler/check/check_exclusive_test.go
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_test.go
    - internal/compiler/cgen/export_test.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/corevalidate_exclusive_test.go
    - testdata/phase5/restrict_borrow.golden.c

key-decisions:
  - "deriveAliasFacts additionally excludes function.PublicOrigin != nil, mirroring selectsByPointerLowering's own identical guard: testdata/phase3/public_view_mixed_access.lang has the identical exclusive-borrow-then-reborrow-to-terminator operation shape as the Phase 5 fixture, so the guard is required for TestAliasFactAgreesWithByPointerSelection to hold, not merely a nice-to-have."
  - "cgen never imports check's AliasFact type. Because TestAliasFactAgreesWithByPointerSelection proves selectsByPointerLowering and deriveAliasFacts decide the identical condition, emitLinearBorrowedByPointer emits restrict unconditionally (it is only ever reached when the fact holds by construction) and reads the justification's loan ID directly off function.Linear.Operations[0].LoanID -- the same identity check's AliasFact.LoanID carries, without a data dependency between the two packages."
  - "EmitForeignManifest's selector was generalized (new singleManifestFunction) rather than singleForeignFunction itself, because EmitForeignHeader/EmitForeignConformance have no equivalent for a by-pointer (non-foreign) function -- widening singleForeignFunction would have given those two emitters a case they cannot honestly serve."
  - "corevalidate's recomputeAliasJustifications feeds a synthetic single-block core.LinearBody (real operation IDs, one synthesized Block with all OperationIDs, no successors) into the EXISTING recomputeLoanEndpoints, rather than adding a parallel loan-endpoint mechanism for straight-line bodies -- recomputeLoanEndpoints already refuses (returns nil) when Linear.Blocks is empty, and straight-line functions never populate that field in their own serialized core."

patterns-established:
  - "A promoted test-only export (cgen.SelectsByPointerLowering, moved from export_test.go's package-cgen-only visibility to a real cgen.go export) is the mechanism for a cross-package agreement test when the two packages must share zero helpers but a test still needs to compare their independent decisions."

requirements-completed: [NAT-03]

coverage:
  - id: D1
    description: "check independently derives a whole-call exclusivity alias fact (AliasFact), exposed on check.Result, that exists only for the exact structural condition cgen's selectsByPointerLowering also selects"
    requirement: "NAT-03"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_exclusive_test.go#TestAliasFactRequiresWholeCallExclusivity"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_exclusive_test.go#TestAliasFactAgreesWithByPointerSelection"
        status: pass
    human_judgment: false
  - id: D2
    description: "cgen emits restrict on exactly one cgen-owned by-pointer parameter, populates the lang.foreign/0 sidecar's emitted_attributes with a justification binding, and every prior-phase manifest byte and generated-C golden stays unchanged"
    requirement: "NAT-03"
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_test.go#TestRestrictEmittedOnlyWithAliasFact"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_test.go#TestRestrictNeverOnForeignExtern"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_test.go#TestEmittedAttributesCarryJustification"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_test.go#TestForeignManifestBytesUnchangedForPriorPhases"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_test.go#TestPhase5ByPointerLoweringGolden"
        status: pass
    human_judgment: false
  - id: D3
    description: "corevalidate independently re-derives the attribute justification from core alone (never reading cgen's own claimed justification as evidence) and refuses an unbacked or wrong-name attribute as a hard build failure (core.attribute_unjustified)"
    requirement: "NAT-03"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_exclusive_test.go#TestAttributeJustificationIsIndependentlyRederived"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_exclusive_test.go#TestUnjustifiedAttributeIsRefused"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_exclusive_test.go#TestUnprovenAttributeNameIsRefused"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_exclusive_test.go#TestAttributeValidatorImportsStayIndependent"
        status: pass
      - kind: manual_procedural
        ref: "Temporarily disabled the JustifiedBy mismatch check in ValidateEmittedAttributes, confirmed TestUnjustifiedAttributeIsRefused failed with the expected error, then restored the check and confirmed green (per plan acceptance criteria's 'demonstrate, then restore')."
        status: pass
    human_judgment: false
  - id: D4
    description: "No prior-phase golden, sidecar byte, or manifest ID moved; full go build/vet/test/test-race ./... stays green"
    requirement: "NAT-03"
    verification:
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestPreviousPhaseGoldenCUnchanged"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestPreviousPhaseManifestIDsUnchanged"
        status: pass
      - kind: other
        ref: "go build ./... && go vet ./... && go test ./... && go test -race ./..."
        status: pass
    human_judgment: false

duration: ~90min
completed: 2026-09-06
status: complete
---

# Phase 5 Plan 4: Emit and Independently Validate `restrict` Summary

**C `restrict` is now emitted on exactly cgen's own by-pointer parameter, bound to a proof `check` independently derives and `corevalidate` independently re-derives from core alone -- NAT-03's "false no-alias facts" mutation finally has an honest subject.**

## Performance

- **Duration:** ~90 min
- **Tasks:** 3 completed
- **Files modified:** 8

## Accomplishments

- **`check.AliasFact`/`deriveAliasFacts`** (Task 1): a new admission-gating fact proving a straight-line function's sole parameter is covered by an exclusive loan across *every* operation from its first use to the function's own terminator -- the C17 §6.7.3.1 "for the duration of that function's execution" condition. It is grounded in `loanLivenessFixpoint`'s own materialized endpoints (via a new `aliasFactEndpoints` helper that runs the fixpoint over the function's REAL post-admission operation IDs), never a re-implemented last-use scan. Exposed on `check.Result.AliasFacts` -- no `core.Program` field, no schema bump (D-05-39). `TestAliasFactRequiresWholeCallExclusivity` proves all four required cases (the fixture yields one fact; a loan expiring before the terminator yields zero; a shared loan yields zero; every Phase 1-4 fixture yields zero). `TestAliasFactAgreesWithByPointerSelection` proves this fact and cgen's `selectsByPointerLowering` decide the *identical* condition on every corpus fixture, including the one Phase 3 fixture (`public_view_mixed_access.lang`) that shares the exact same operation shape but must NOT get a fact (its `PublicOrigin` marks it a distinct, already-shipped semantic category).
- **`restrict` emission + justification binding** (Task 2): `emitLinearBorrowedByPointer` now emits `<TYPE> *restrict <param>`, legal without importing check's `AliasFact` type because the agreement test above proves cgen's own gate is the same condition. `foreignManifestDocument.EmittedAttributes` changes from `[]string` to `[]EmittedAttribute` (same JSON key, same empty-array encoding for every existing sidecar -- asserted, not assumed, by `TestForeignManifestBytesUnchangedForPriorPhases`). `EmitForeignManifest` now resolves its function via a new `singleManifestFunction` selector (foreign-contract case delegates verbatim to the unchanged `singleForeignFunction`; otherwise falls back to the sole by-pointer-lowered function) and populates one `EmittedAttribute{Attr: "restrict", CoreNode, Parameter, JustifiedBy}` entry, reading `JustifiedBy` directly off `function.Linear.Operations[0].LoanID`. `control:foreign.no_unproven_attributes` is narrowed via new `JustifiableAttributes`/`ScanForUnjustifiedAttributes` (manifest-entry scope only); `BannedOptimizerAttributes`/`ScanForBannedAttributes` (declaration-site scope) are completely unchanged -- `restrict` is still banned in every foreign extern declaration, proven by `TestRestrictNeverOnForeignExtern`.
- **Independent re-derivation in corevalidate** (Task 3): `corevalidate.AttributeClaim` (a local decoding of the sidecar's JSON, never importing `cgen.EmittedAttribute`), `recomputeAliasJustifications` (feeds a synthetic single-block `core.LinearBody` view of a straight-line function's own real operations into the *existing* `recomputeLoanEndpoints` reachability-closure-plus-reduction mechanism -- `recomputeLoanEndpoints` itself refuses when `Linear.Blocks` is empty, which every straight-line function's serialized core is by design, so this synthesis is the honest way to reuse it rather than duplicate it), and the exported `ValidateEmittedAttributes`, which refuses an unproven attribute name or a mismatched justification with the new `core.attribute_unjustified` diagnostic ID (joining `lang.diagnostic/1` without moving any existing ID). Manually verified load-bearing: temporarily disabled the mismatch check, confirmed `TestUnjustifiedAttributeIsRefused` failed with the expected error, then restored it.

## Task Commits

Each task was committed atomically:

1. **Task 1: Prove exclusive-for-the-whole-call in check, as an admission-gating fact** - `1fc291a` (feat)
2. **Task 2: Emit restrict on the by-pointer parameter and populate emitted_attributes with its justification binding** - `8a3d34b` (feat)
3. **Task 3: Independently re-derive the justification in corevalidate and narrow control:foreign.no_unproven_attributes** - `8798c5e` (feat)

## Files Created/Modified

- `internal/compiler/check/check.go` - `AliasFact`, `AliasFactExclusiveBorrow`/`AliasFactUniqueOwner`, `aliasFactEndpoints`, `deriveAliasFacts`; `Result.AliasFacts`; wired into `Program()`'s straight-line append site
- `internal/compiler/check/check_exclusive_test.go` - `TestAliasFactRequiresWholeCallExclusivity`, `TestAliasFactAgreesWithByPointerSelection`, supporting fixtures/helpers
- `internal/compiler/cgen/cgen.go` - promoted `SelectsByPointerLowering` export; `restrict` emission in `emitLinearBorrowedByPointer`; `EmittedAttribute` type; `emittedAttributeForByPointerParameter`; `singleManifestFunction`; `EmitForeignManifest` generalized; `JustifiableAttributes`/`ScanForUnjustifiedAttributes`
- `internal/compiler/cgen/cgen_test.go` - `TestRestrictEmittedOnlyWithAliasFact`, `TestRestrictNeverOnForeignExtern`, `TestEmittedAttributesCarryJustification`, `TestForeignManifestBytesUnchangedForPriorPhases`
- `internal/compiler/cgen/export_test.go` - removed the now-redundant test-only `SelectsByPointerLowering` alias (promoted to a real export in cgen.go)
- `internal/compiler/corevalidate/corevalidate.go` - `AttributeClaim`, `AttributeUnjustifiedError`, `recomputeAliasJustifications`, `ValidateEmittedAttributes`
- `internal/compiler/corevalidate/corevalidate_exclusive_test.go` - `TestAttributeJustificationIsIndependentlyRederived`, `TestUnjustifiedAttributeIsRefused`, `TestUnprovenAttributeNameIsRefused`, `TestAttributeValidatorImportsStayIndependent`
- `testdata/phase5/restrict_borrow.golden.c` - updated to `LANG_BUFFER *restrict LANG_BUFFER_LANG_PARAMETER_0` on the by-pointer signature line (the one deliberate generated-C change this plan makes to a pinned golden)

## Decisions Made

See `key-decisions` in frontmatter: the `PublicOrigin` exclusion required for cross-package agreement, cgen never importing check's `AliasFact` type (relying instead on the proven structural equivalence), generalizing `EmitForeignManifest`'s own selector rather than `singleForeignFunction` itself, and feeding a synthetic single-block CFG into corevalidate's existing `recomputeLoanEndpoints` rather than adding a parallel mechanism.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Promoted cgen.SelectsByPointerLowering from a test-only accessor to a real export**
- **Found during:** Task 1, writing `TestAliasFactAgreesWithByPointerSelection`
- **Issue:** The plan's own acceptance criteria require a cross-package test (in `check`'s test suite) proving `deriveAliasFacts` and cgen's `selectsByPointerLowering` agree on every corpus fixture. `selectsByPointerLowering` was exposed only via `export_test.go` (`package cgen`, a Go test-only file invisible to importers -- including `check`'s own test package).
- **Fix:** Moved the exported wrapper from `export_test.go` into `cgen.go` itself as a genuine, permanent export (`cgen.SelectsByPointerLowering`), removing the now-redundant test-only alias. Production dispatch (`Emit`/`EmitNative`) is unaffected -- it still calls the unexported `selectsByPointerLowering` directly.
- **Files modified:** `internal/compiler/cgen/cgen.go`, `internal/compiler/cgen/export_test.go`
- **Verification:** `TestAliasFactAgreesWithByPointerSelection` passes for every corpus fixture; full `go build ./...`/`go vet ./...` clean.
- **Committed in:** `1fc291a` (Task 1 commit)

**2. [Rule 1 - Bug] Excluded declared-borrow-return functions (PublicOrigin != nil) from deriveAliasFacts**
- **Found during:** Task 1, running `TestAliasFactRequiresWholeCallExclusivity`'s Phase 1-4 enumeration
- **Issue:** `testdata/phase3/public_view_mixed_access.lang` has the exact same exclusive-borrow-then-reborrow-to-terminator operation shape as the Phase 5 tracer fixture, so the initial derivation wrongly produced a fact for it -- the same additivity gap 05-01's own summary documented for `selectsByPointerLowering` and fixed the same way.
- **Fix:** Added the identical `function.PublicOrigin != nil` exclusion `selectsByPointerLowering` already uses.
- **Files modified:** `internal/compiler/check/check.go`
- **Verification:** `TestAliasFactRequiresWholeCallExclusivity` and `TestAliasFactAgreesWithByPointerSelection` both pass for the full Phase 1-4 corpus.
- **Committed in:** `1fc291a` (Task 1 commit -- caught and fixed before commit, not a follow-up)

---

**Total deviations:** 2 auto-fixed (1 blocking/test-visibility, 1 bug/additivity gap). **Impact:** Both were necessary for the plan's own acceptance criteria (the cross-package agreement test) to be satisfiable at all and to hold across the full corpus. No scope creep.

### Acceptance-criteria wording note (not a deviation, documented per plan's own instruction)

Task 2's literal acceptance grep (`grep -c 'restrict' testdata/phase5/restrict_borrow.golden.c` returns exactly `1`) does not hold as written: `grep -c` counts matching LINES, and the fixture's own module name (`phase5.restrict_borrow`, chosen in plan 05-01) makes "restrict" a substring of nearly every generated identifier/comment line, so the count is 8, not 1. The real intent -- exactly one `*restrict` qualifier token, on the function signature line -- is asserted precisely by `TestRestrictEmittedOnlyWithAliasFact` (`strings.Count(generated, "*restrict ")`) and the existing `TestPhase5ByPointerLoweringGolden`'s by-pointer-marker count. This is a wording artifact of the literal grep against a fixture name chosen in a prior plan, not a functional gap.

## Issues Encountered

None beyond the deviations above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- NAT-03's "false no-alias facts" mutation now has an honest subject: `restrict` is emitted, justified, and independently re-validated. 05-05+ can build the mutation runner that injects a false `restrict` (or corrupts the justification binding) and confirms `corevalidate.ValidateEmittedAttributes` refuses it.
- `cgen.SelectsByPointerLowering` and `cgen.EmittedAttribute`/`ScanForUnjustifiedAttributes`/`JustifiableAttributes` are now stable, genuinely exported surface for later plans (e.g. the mutation runner, the QLT-01 control registry) to consume directly, rather than needing test-only access.
- `EmitForeignManifest`'s generalized `singleManifestFunction` selector is scoped exactly to itself; `EmitForeignHeader`/`EmitForeignConformance` remain foreign-contract-only, so later plans extending the by-pointer path do not need to reason about header/conformance-unit generalization that does not exist.
- No blockers.

---
*Phase: 05-native-equivalence-and-adversarial-evidence*
*Completed: 2026-09-06*

## Self-Check: PASSED

All 8 modified files verified present on disk with expected content. All 3 task commit hashes (`1fc291a`, `8a3d34b`, `8798c5e`) verified in `git log`. Plan-level `<verification>` block re-run clean: all eight named tests pass (`TestAliasFactRequiresWholeCallExclusivity`, `TestRestrictEmittedOnlyWithAliasFact`, `TestRestrictNeverOnForeignExtern`, `TestEmittedAttributesCarryJustification`, `TestAttributeJustificationIsIndependentlyRederived`, `TestUnjustifiedAttributeIsRefused`, `TestPreviousPhaseGoldenCUnchanged`, `TestPreviousPhaseManifestIDsUnchanged`); `go build ./...`, `go vet ./...`, `go test ./...`, and `go test -race ./...` all clean, including the previously-fixed `TestDebtRegistersAreWellFormed`.
