---
phase: 07-calls-signatures-and-call-graph-refusal
plan: 02
subsystem: compiler-validation
tags: [go, originvalidate, corevalidate, ownership, interprocedural, mutation-testing]

requires:
  - phase: 07-01
    provides: "lang.interface/1 schema (core.FunctionSignature with Callable/ClosureDigest fields), strict DecodeInterface, zero-callee ClosureDigest preimage"
provides:
  - "originvalidate.PublishProblemsFor(core.Function) []Problem -- per-function publication-safety predicate, extracted from ValidatePublished"
  - "Callable derived from PublishProblemsFor (D-07-31): publication safety, never export membership"
  - "testdata/phase07/clean_but_unpublishable.lang -- extracted negative control (D-07-44)"
  - "corevalidate's independent signature-summary peer, wired into both replayStraightLine and replayBlocks (D-07-20/D-07-21/D-07-22)"
  - "Five seeded D-07-24 faults as automated, deterministic, -race-clean tests, including the bilateral fault that fails the gate"
  - "PHASE-07-DEBT.md in the mechanically-checked debt-register format, with D-07-33's narrowing confirmed"
affects: [07-05, 07-07, 09]

actuals:
  tokens: 210000
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Independent structural re-derivation via materially different mechanism (map lookup + forward set-propagation vs linear scan + backward walk), never a second caller of the producer's own implementation"
    - "export_test.go pattern for exposing unexported fault-injection seams to a package's own external test package, split across two packages by which seam each fault needs"

key-files:
  created:
    - testdata/phase07/clean_but_unpublishable.lang
    - internal/compiler/corevalidate/corevalidate_summary_peer_test.go
    - internal/compiler/corevalidate/corevalidate_mutation_matrix_test.go
    - internal/compiler/corevalidate/export_test.go
    - internal/compiler/originvalidate/export_test.go
  modified:
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/originvalidate/originvalidate_test.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/corevalidate_endpoint_internal_test.go
    - internal/compiler/check/check_exclusive_test.go
    - .planning/phases/07-calls-signatures-and-call-graph-refusal/PHASE-07-DEBT.md

key-decisions:
  - "D-07-31/D-07-32 implemented exactly as the checkpoint ratified: Callable = len(PublishProblemsFor(fn))==0, no export field added anywhere."
  - "D-07-33's narrowing implemented literally: the peer independently re-derives ONLY core.origin_omitted; the other three classes are declared, tested false agreements (TestPeerDoesNotRederiveNarrowedClasses), not silently claimed coverage."
  - "Fault seams split across originvalidate and corevalidate packages' own TestStage0SummaryMutationMatrix, because Go's export_test.go pattern only reaches a package's OWN external test package -- a cross-package bilateral test cannot flip two different packages' unexported vars in one function without either exporting them for real (forbidden by D-07-42) or splitting by which seam it needs."
  - "Fault 3 (bilateral) demonstrated by combining corevalidate's own forced-true peer value (flipped live in that subtest) with the producer's forced-true value as an independently-proven literal (established by originvalidate_test.go's fault4 subtest on the same fixture), then feeding both into the shared sweep-report function -- an engineering compromise for the cross-package constraint, documented in both test files' comments."

requirements-completed: [SEM-05, SEM-06, QLT-08]

coverage:
  - id: D1
    description: "Callable is publication safety (D-04-03), not export membership: PublishProblemsFor extracted, ValidatePublished delegates with byte-identical whole-program first-problem behavior"
    requirement: "SEM-05"
    verification:
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_test.go#TestPublishProblemsForMatchesValidatePublishedAcrossCorpus"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_test.go#TestCallableIsPublicationSafetyNotExportMembership"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_test.go#TestBuildInterfaceNeverConsultsExportList"
        status: pass
    human_judgment: false
  - id: D2
    description: "testdata/phase07/clean_but_unpublishable.lang extracted verbatim (D-07-44) from check_exclusive_test.go:46-58, checks clean, fails publication with core.origin_omitted"
    requirement: "SEM-06"
    verification:
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_test.go#TestCallableIsPublicationSafetyNotExportMembership"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_exclusive_test.go#TestExclusiveBorrowLowersToCore"
        status: pass
    human_judgment: false
  - id: D3
    description: "corevalidate's independent signature-summary peer wired into BOTH replayStraightLine and replayBlocks; zero divergence on structural fields over the whole testdata/phase1..phase4+phase07 corpus"
    requirement: "SEM-06"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_summary_peer_test.go#TestSummaryPeerStructuralFieldsMatchProducerAcrossCorpus"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_summary_peer_test.go#TestSummaryPeerRunsAtBothReplaySites"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_endpoint_internal_test.go#TestValidatorImportsStayIndependent"
        status: pass
    human_judgment: false
  - id: D4
    description: "The peer's Callable re-derivation is narrowed to core.origin_omitted only; the three excluded classes are declared false agreements, not silently claimed"
    requirement: "SEM-06"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_summary_peer_test.go#TestSummaryPeerCallableAgreesOnOriginOmittedClass"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_summary_peer_test.go#TestPeerDoesNotRederiveNarrowedClasses"
        status: pass
    human_judgment: false
  - id: D5
    description: "Five seeded D-07-24 faults as deterministic, -race-clean, automated tests, including the bilateral fault that FAILS the gate rather than passing it"
    requirement: "QLT-08"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_mutation_matrix_test.go#TestStage0SummaryMutationMatrix"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_test.go#TestStage0SummaryMutationMatrix"
        status: pass
    human_judgment: false
  - id: D6
    description: "PHASE-07-DEBT.md conforms to the mechanically-checked debt-register format and confirms D-07-33's narrowing; fixes the pre-existing frontmatter failure"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestDebtRegistersAreWellFormed"
        status: pass
    human_judgment: false

duration: 90min
completed: 2026-09-08
status: complete
---

# Phase 07 Plan 02: Callable as Publication Safety + Independent corevalidate Summary Peer Summary

**`Callable` is now literally `len(originvalidate.PublishProblemsFor(fn))==0` (publication safety, never export membership), backed by an independent `corevalidate` structural peer wired into both replay shapes and five seeded, deterministic, -race-clean faults including the bilateral one that fails the gate.**

## Performance

- **Duration:** ~90 min
- **Tasks:** 3 (plus one checkpoint:decision, auto-approved under active auto-mode)
- **Files modified:** 11 (5 created, 6 modified)

## Accomplishments

- Extracted `originvalidate.PublishProblemsFor(core.Function) []Problem` from `ValidatePublished`'s loop body (D-07-32); `ValidatePublished` now delegates and preserves its whole-program first-problem contract exactly, proven byte-identical across the whole `testdata/phase1..phase4+phase07` corpus.
- `BuildInterface`'s `Callable` field is now derived from `PublishProblemsFor` (D-07-31): publication safety, not export membership. No export field was added to `core.Program`/`core.Function`; the string `"Exports"` never appears in `originvalidate.go` (asserted by a dedicated structural test).
- Extracted `testdata/phase07/clean_but_unpublishable.lang` verbatim from `check_exclusive_test.go:46-58` (D-07-44, correcting amendment A-03): checks clean, fails publication with `core.origin_omitted`. `check_exclusive_test.go` and `originvalidate_test.go` now both read this single extracted source instead of embedding it a second/third time.
- Landed `corevalidate`'s independent (D-07-20/D-07-22) structural signature-summary peer: `derivePeerSignature` re-derives Parameters/Return/Abilities/Fails/Foreign via a direct type-fact map lookup and forward set-propagation, materially different from `originvalidate.BuildInterface`'s linear scan and backward per-return walk. Never imports `originvalidate`, never calls `BuildInterface`/`PublishProblemsFor` (enforced by `TestValidatorImportsStayIndependent`'s extended forbidden-import list). Wired into **both** `replayStraightLine` and `replayBlocks` (D-07-21), plus the pure `lang.core/0` match-only dispatch path.
- `Callable`'s peer re-derivation is narrowed to the `core.origin_omitted` class only (D-07-33). `TestPeerDoesNotRederiveNarrowedClasses` names the exclusion directly: for `core.origin_understated`, `core.origin_access_mismatch`, and foreign-origin-omitted, the peer's `Callable` stays `true` -- a declared **false agreement**, not silently claimed coverage.
- Five seeded D-07-24 faults, all deterministic (`-count=2 -shuffle=on` identical), `-race`-clean, all unexported (D-07-42): fault 1 (producer-only type-fact lookup), fault 2 (two-site, replayBlocks-only wiring), fault 3 (bilateral -- asserts the exact report string `"no divergence detected under bilateral fault"` as a **gate failure**, never a pass), fault 4 (producer-side Callable), fault 5 (peer-side Callable).
- `corevalidate.SummaryPeerControls` exports this plan's three control names for `07-07`'s completeness matrix; each backed by a non-empty seeded-mutation list, asserted directly.
- Restructured `PHASE-07-DEBT.md` into the mechanically-checked debt-register format (`items:` frontmatter, `## Items` table, `### D-ID` detail sections) `TestDebtRegistersAreWellFormed` enforces -- fixing the pre-existing failure noted in this plan's prior-wave context -- and confirmed D-07-33's narrowing entry now that the peer it describes actually ships.

## Task Commits

1. **Checkpoint: ratify Callable as publication safety, not export membership** -- auto-approved (`⚡ Auto-selected: Proceed as decided`); `AUTO_CFG=true`, gate defaulted to `blocking` (not `blocking-human`), plan's own `<default>Proceed as decided.</default>` is the first/recommended option.
2. **Task 1: PublishProblemsFor, Callable, and the extracted negative control** -- `1941a83` (feat)
3. **Task 2: The corevalidate signature-summary peer, at BOTH replay sites, with its narrowing declared** -- `9b62c1a` (feat)
4. **Task 3: The three seeded faults -- Stage 0's QLT-08 gate, including the bilateral one that must FAIL** -- `8def161` (test)

_Task 3's commit also carries the `PHASE-07-DEBT.md` register restructure, since Task 3 is what confirms D-07-33's entry now that the peer ships._

## Files Created/Modified

- `internal/compiler/originvalidate/originvalidate.go` - `PublishProblemsFor` extraction, `Callable` derivation, fault seams (`typeFactExactIDMatchOverride`, `forceCallableAlwaysTrue`)
- `internal/compiler/originvalidate/originvalidate_test.go` - new corpus-wide equivalence/Callable tests, structural export-list test, fault 1/4 subtests
- `internal/compiler/originvalidate/export_test.go` - test-only setters for originvalidate's fault seams (never compiled into production)
- `testdata/phase07/clean_but_unpublishable.lang` - extracted negative control (D-07-44)
- `internal/compiler/check/check_exclusive_test.go` - reads the extracted fixture instead of an inline copy (deviation, see below)
- `internal/compiler/corevalidate/corevalidate.go` - `derivePeerSignature`, `recordSummaryPeer`, `peerCallable`/`peerReturnDerivesFromBorrow`/`peerParameterEscapesOwned`, `Result.PeerSignatures`/`PeerSiteCoverage`, fault seams, `SummaryPeerControls`
- `internal/compiler/corevalidate/corevalidate_summary_peer_test.go` - Tests 1-5 from Task 2's `<behavior>`
- `internal/compiler/corevalidate/corevalidate_mutation_matrix_test.go` - faults 2/3/5 + control-completeness assertion
- `internal/compiler/corevalidate/corevalidate_endpoint_internal_test.go` - extended import-independence forbidden list (`originvalidate`)
- `internal/compiler/corevalidate/export_test.go` - test-only setters for corevalidate's fault seams
- `.planning/phases/07-calls-signatures-and-call-graph-refusal/PHASE-07-DEBT.md` - restructured to the mechanically-checked register format, D-07-33 confirmed

## Decisions Made

See `key-decisions` in frontmatter. The most consequential: **fault seams had to be split across both packages' own `TestStage0SummaryMutationMatrix`** rather than living in one function, because Go's `export_test.go` pattern only reaches a package's own external test package -- a single cross-package test cannot flip two different packages' unexported vars simultaneously without either exporting them for real (which D-07-42 forbids) or splitting by which seam it needs. Fault 3 (bilateral) is demonstrated by combining corevalidate's own live-flipped peer value with the producer's forced-true value as an independently-proven literal constant (established by originvalidate's own fault 4 subtest on the identical fixture) fed into a shared `sweepReportBilateralAgreement` function. This is an honest engineering compromise for a real Go visibility constraint, not a shortcut around the fault's intent -- both halves of the "bilateral" claim are independently, mechanically proven; they are just proven in two cooperating test functions rather than one.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2/3 - Missing critical / blocking] `check_exclusive_test.go` modified but not listed in `files_modified`**
- **Found during:** Task 1
- **Issue:** Task 1's own acceptance criteria required `check_exclusive_test.go` and `originvalidate_test.go` to both read the extracted fixture "instead of embedding it," but the plan frontmatter's `files_modified` list omitted `check_exclusive_test.go`.
- **Fix:** Modified it as the acceptance criteria required (its inline `clean` source variable now reads `testdata/phase07/clean_but_unpublishable.lang`).
- **Files modified:** `internal/compiler/check/check_exclusive_test.go`
- **Verification:** `go test ./internal/compiler/check/...` passes.
- **Committed in:** `1941a83`

**2. [Rule 3 - Blocking] `export_test.go` files needed in both packages, not listed in `files_modified`**
- **Found during:** Task 3
- **Issue:** The bilateral fault (fault 3) requires flipping unexported seams in both `originvalidate` and `corevalidate` from cooperating test functions. Go's package visibility rules make an unexported package-level var reachable ONLY from that package's own test binary; D-07-42 explicitly forbids making the seams exported in production code.
- **Fix:** Added `export_test.go` (package-matching filename-suffixed `_test.go`, so excluded from every non-test build) in each package, exposing restore-returning setter functions to that package's own external test package only. Split `TestStage0SummaryMutationMatrix` across both packages by which seam each of the five faults needs.
- **Files modified:** `internal/compiler/originvalidate/export_test.go`, `internal/compiler/corevalidate/export_test.go` (both new)
- **Verification:** `go test ./internal/compiler/corevalidate/... ./internal/compiler/originvalidate/... -run MutationMatrix -v` shows a PASS line for `TestStage0SummaryMutationMatrix` in both packages; `-count=2 -shuffle=on` deterministic; `-race` clean.
- **Committed in:** `8def161`

**3. [Rule 3 - Blocking] `PHASE-07-DEBT.md` needed full register-format restructure, not just an appended entry**
- **Found during:** pre-flight (prior-wave context flagged this as a known pre-existing failure)
- **Issue:** `TestDebtRegistersAreWellFormed` (session package) requires `items:` frontmatter, a `## Items` table with specific columns, and matching `### D-ID` detail sections. The existing `PHASE-07-DEBT.md` (written before this mechanical check existed) had none of this structure -- appending D-07-33's deferral text alone (as Task 2's action literally requested) would not have satisfied the test.
- **Fix:** Restructured the whole file into the register format, converting "Deferred item 1" and "Deferred item 2" into `D-03-02` and `D-07-33` register rows with matching detail sections, while leaving the "scope-cut trigger," "accepted limitations," and "explicitly rejected" sections as prose (bulleted, non-`### D-`-headed) so they are not swept into the mechanical row/detail-section parity check.
- **Files modified:** `.planning/phases/07-calls-signatures-and-call-graph-refusal/PHASE-07-DEBT.md`
- **Verification:** `go test ./internal/compiler/session/... -run TestDebtRegistersAreWellFormed -v` passes for all six registers including `PHASE-07-DEBT.md`.
- **Committed in:** `8def161`

**4. [Rule 1 - Bug, in my own implementation] `peerReturnDerivesFromBorrow`'s first draft missed the copy-then-borrow chain**
- **Found during:** Task 2, while running `TestSummaryPeerCallableAgreesOnOriginOmittedClass`
- **Issue:** My initial forward-propagation walk only treated `OpBorrowShared`/`OpBorrowExclusive` as "derived" when the SourceID was the parameter place ID directly. `check.go`'s match-arm lowering always copies the scrutinee into a fresh place before each arm borrows it, so the borrow's SourceID is almost always a copy of the parameter, not the parameter's own place ID -- causing a false negative (Callable wrongly `true`) on `testdata/phase3/public_view_multi_arm_access_conflict.lang`.
- **Fix:** Added a `paramTrace` set (mirroring `peerParameterEscapesOwned`'s existing pattern) that also propagates through `OpMove`/`OpCopy`, so a borrow of a copy-of-the-parameter is recognized as tracing to the parameter.
- **Files modified:** `internal/compiler/corevalidate/corevalidate.go`
- **Verification:** `TestSummaryPeerCallableAgreesOnOriginOmittedClass` and `TestSummaryPeerStructuralFieldsMatchProducerAcrossCorpus` both pass over the whole corpus.
- **Committed in:** `9b62c1a`

---

**Total deviations:** 4 (2 missing-file additions required by acceptance criteria, 1 pre-existing structural-format fix, 1 self-caught implementation bug). **Impact:** All were necessary for correctness or for satisfying this plan's own acceptance criteria; none represent scope creep beyond what Task 1-3 already asked for.

## Known Stubs

None. `PublishProblemsFor`, the peer, and all five fault seams are fully wired and exercised.

## Threat Flags

None beyond what `07-02-PLAN.md`'s own `<threat_model>` already registered (T-07-08 through T-07-13, T-07-SC) -- all mitigated as designed; no new security-relevant surface was introduced outside that register.

## Issues Encountered

**`go run ./cmd/lang --json check testdata/phase07/clean_but_unpublishable.lang` does not satisfy Task 1's second `<verify>` automated command as literally written.** The plan's verify step expects zero error-severity diagnostics from this CLI invocation. However, `session.CheckCommandFile` (the `lang check` CLI command's handler) has, since Phase 4's D-04-27/WR-01, unconditionally run `originvalidate.ValidatePublished` after a successful `session.Check` and turned any problem into an error diagnostic -- this is pre-existing, established behavior, not something this plan touches or should touch (changing it would be an architectural change to a shipped CLI contract, well outside Task 1's scope). Every fixture of this exact "checks clean but fails publication" shape -- including the pre-existing `testdata/phase3/public_view_omitted.lang`, whose own header comment is now stale on this exact point -- produces the identical `core.origin_omitted` error diagnostic from `lang check`. I verified this is not specific to the new fixture. The acceptance criteria that actually govern this task ("Checking that fixture produces zero error-severity diagnostics" via `session.Check`, separately, "`ValidatePublished` on its checked program returns exactly one problem whose Code is `core.origin_omitted`") are both fully satisfied and covered by passing Go tests (`TestExclusiveBorrowCleanShapeChecksButCannotPublish`, `TestCallableIsPublicationSafetyNotExportMembership`). I did not modify `CheckCommandFile`'s fusion behavior. Recording this as an unrun/inapplicable `<verify>` command rather than silently marking the task done without addressing the discrepancy.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Stage 0's D-07-25 gate is closed for this plan's scope: `Callable` is the correct D-04-03 predicate, independently re-derived by two peers over the whole existing M001 corpus with both replay shapes exercised, and all five seeded faults behave as specified (including the bilateral one failing the gate).
- `07-05` (call-admission predicate) can now consume `Callable` with confidence it means publication safety, not export membership.
- `07-07`'s phase-wide completeness matrix has `corevalidate.SummaryPeerControls` to compare against by exact set equality.
- Phase 09 has a fully-declared, tested-false-agreement narrowing (D-07-33) to close, plus the loan-liveness peer (D-03-02) -- both entries are now in the mechanically-checked `PHASE-07-DEBT.md` register.
- Ready for `07-03`.

---
*Phase: 07-calls-signatures-and-call-graph-refusal*
*Completed: 2026-09-08*
