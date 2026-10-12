---
phase: "30"
slug: "local-modules"
# status lifecycle: planned (plan-phase; TestPhase20ValidationLifecycle forbids committed drafts) → validated (30-09 / validate-phase)
status: planned
nyquist_compliant: false
wave_0_complete: false
created: "2026-10-12"
---

# Phase 30 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> The source of truth for each test is `30-RESEARCH.md` § Validation Architecture.
> Zero human UAT is required (PROJECT.md § Verification Operating Preference).

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (Go 1.24.0 stdlib `testing`) |
| **Config file** | none (go.mod only) |
| **Quick run command** | each task's own `<automated>` command from its PLAN.md (the Phase30 tests are created by the plans; 30-09 rewrites this row once they exist) |
| **Session meta set** | `go test ./internal/compiler/session -run 'Inventory\|Groundedness\|ValidationLifecycle\|TestPhase28EntryExitClassManifest\|TestPhase272AcceptedCorpusByteIdentity\|TestPhase273ConformanceCorpusByteIdentity\|TestLanguageMaturityCountsAreCurrent\|TestExitCodeDocsTable\|TestEntryFunctionCalledOnlyViaResolveEntry\|TestPhase16PublicEmitterConsumerInventory' -count=1` |
| **Full suite command** | hosted CI only: `go test ./...` and `go test -race -timeout=20m ./...` on ubuntu-latest and macos-latest |
| **Estimated runtime** | focused runs take ~5-60 s per package |

---

## Sampling Rate

- **After every task commit:** run the focused `-run 'Phase30'` command for each touched package. Also run the session meta set whenever a `.schway` file, a `session.go`/`evidence.go` line, or a planning doc that lists a `go test` command changes.
- **After every plan wave:** run every Phase30 row for the wave's packages, plus `TestPhase30PreChangeGolden`, `TestPhase272AcceptedCorpusByteIdentity` and `TestPhase273ConformanceCorpusByteIdentity`.
- **Before `/gsd-verify-work`:** hosted full and race CI must be green on both OSes, with the run IDs recorded.
- **Max feedback latency:** 60 s per focused run.

---

## Per-Task Verification Map

This map gives one row per success criterion or decision; the plan-checker maps tasks to rows. The commands come from `30-RESEARCH.md` § Validation Architecture.

| Ref | Requirement | Secure Behavior | Test | Armed control | Status |
|-----|-------------|-----------------|------|---------------|--------|
| SC1 / D-17 | MOD-03 | Imported Adler-32 gives the pinned answers 0..9 on native and on the interpreter | `TestPhase30AdlerModuleAppCLI`, `TestPhase30AdlerModuleInterpreter` | wrong expected row fails | ⬜ pending |
| SC1 | MOD-03 | The library exists as exactly one copy | `TestPhase30LibraryIsSingleCopy` | copy `fn adler32(` into the entry | ⬜ pending |
| SC1 / D-25 | MOD-03 | `..`, absolute and separator imports are refused as syntax errors | `TestPhase30ImportPathSyntaxRefused` | one row per form: "slash" / "absolute" / "dotdot" (`admitModulePathFormForTest`, 30-01) | ⬜ pending |
| SC1 / D-08 | MOD-03 | Symlink escapes and in-root symlinks are refused | `TestPhase30SymlinkAndEscapeRefused` | skip per-component Lstat; open without Root | ⬜ pending |
| D-10 | MOD-03 | A FIFO is refused without blocking | `TestPhase30FIFORefusedWithoutBlocking` | open before Lstat | ⬜ pending |
| D-09 | MOD-03 | Case must match exactly | `TestPhase30CaseExactMatch` | EqualFold accept | ⬜ pending |
| D-10 | MOD-03 | TOCTOU swaps are refused, size is never taken from Stat, each file is read once, and there is a single IO site | `TestPhase30FakeFSToctouAndStatSize`, `TestPhase30LoaderReadsEachFileOnce`, `TestPhase30SingleIOSite` | size from Stat and drop SameFile (`sizeFromStatForTest`, `skipSameFileForTest`, 30-03); re-read (`rereadForTest`, 30-04); unclassified source read and stray `os.OpenRoot` (synthetic sources, 30-08) | ⬜ pending |
| SC2 | MOD-04 | A private import is refused (`module.not_exported`) | `TestPhase30PrivateNameRefused` | seam admits private | ⬜ pending |
| SC2 | MOD-04 | corevalidate re-derives visibility independently | `TestPhase30CorevalidateVisibilityPeer`, `TestPhase30ValidatedProgramKeepsModuleFacts` | peer ignores Exported; `json:"-"` clone | ⬜ pending |
| SC2 | MOD-04 | Admission is body-blind | `TestPhase30CallAdmissionBodyBlind`, `TestPhase30CallableFlipIsASignatureChange` | body-read seam | ⬜ pending |
| SC3 / D-24 | MOD-05 | Self, 2-, 3- and via-entry cycles are refused | `TestPhase30ImportCycles` | visited-set instead of gray/black | ⬜ pending |
| SC3 / D-19 | MOD-05 | missing, path_mismatch, duplicate, duplicate_import and invalid_segment are refused | `TestPhase30GraphRefusals` | sibling accepted fixture per case, plus `skipGraphCheckForTest` rows for not_found / duplicate / path_mismatch (30-04) | ⬜ pending |
| SC3 / D-04 / D-20 | MOD-05 | Collisions and imported `main`/data/foreign are refused, naming both origins | `TestPhase30ImportCollisions` | seam disables the collision checks (`skipImportCollisionForTest`, 30-05) | ⬜ pending |
| SC3 / D-11 | MOD-05 | A diagnostic inside an imported file names that file, and IDs are distinct per unit | `TestPhase30ImportedFileDiagnosticNamesFile`, `TestPhase30UnitIdentity`, `TestPhase30UnitOmitempty` | skip stamping (`check.SetSkipUnitStampingForTest`, engaged in 30-08); unit left out of the ID (`unitOutOfIdentityForTest`, 30-02); drop omitempty (`TestPhase30GateUnitOmitemptyInvariance`, 30-00) | ⬜ pending |
| D-23 | MOD-05 | `unit=` projection, and the repair driver declines cross-unit repairs | `TestPhase30UnitProjection`, `TestPhase30RepairDriverDeclinesCrossUnit`, `TestPhase30DebugMapUnit` | always print; driver ignores unit; key by bare name | ⬜ pending |
| SC4 | MOD-05 | A diamond is loaded and emitted once | `TestPhase30DiamondLoadsAndEmitsOnce` | diamond counted twice (`modload.SetDiamondTwiceForTest`, engaged in 30-08) | ⬜ pending |
| SC4 / D-15 | MOD-05 | No-import programs stay byte-identical | `TestPhase30PreChangeGolden`, `TestPhase30PreChangeGoldenIsNotInert`, plus the existing 272/273 tests | non-inert control | ⬜ pending |
| SC4 / D-12 / D-18 | MOD-05 | C names are injective and a reversed import order gives identical C | `TestPhase30CNamesInjective`, `TestPhase30EmitOrderIndependent`, `TestPhase30ReversedImportOrderIdenticalC` | entry allocated second (`importedFirstForTest`, 30-07); skip canonical sort (`skipCanonicalImportSortForTest`, 30-07); merge in import order (`modload.SetMergeInImportOrderForTest`, engaged in 30-08). The plain `_` join control belongs to `TestPhase30GateCNamesInjective` (30-00). | ⬜ pending |
| D-13 / D-21 | MOD-03 | Caps hold at exact N and refuse N+1 | `TestPhase30CapsExactNAndNPlusOne`, `TestPhase30ImportsLimitParse` | `>`→`>=` per cap; count after open | ⬜ pending |
| D-02 / D-03 | MOD-03 | Grammar and formatter | `TestPhase30ImportGrammar`, `TestPhase30FormatImport`, `TestPhase30GeneratedImportRoundTrips` | formatter sorts imports | ⬜ pending |
| D-22 | MOD-03 | Single-file commands refuse imports | `TestPhase30SingleFileCommandsRefuseImports` | `check.Program` ignores Imports (`check.SetIgnoreUnloadedImportsForTest`, engaged in 30-08) | ⬜ pending |
| D-11 / MOD-04 | MOD-04 | Every consumer of the merged program keeps two same-named private functions distinct | `TestPhase30SameBareNameTwoModules` | seeded fault in lib's helper only, per consumer (corevalidate, originvalidate, pathoracle, callgraph) | ⬜ pending |
| D-15 / D-22 | MOD-05 | The no-import debug-map stays guard-free and byte-identical; only the module path applies MaxCheckWork | `TestPhase30DebugMapNoImportUnguarded`, `TestPhase30ModuleWorkLimit` | guard the no-import path (`debugMapGuardNoImportForTest`, 30-08) | ⬜ pending |
| Security | MOD-03 | No absolute path or PathError text appears in diagnostics | `TestPhase30ErrorsNeverLeakAbsolutePaths` | format `err.Error()` | ⬜ pending |
| D-14 | all | Gate flags (a)-(f) | `TestPhase30Gate*` and `30-00-GATE-VERDICT.md` | each flag has its own control | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `testdata/phase30/prechange-corpus.json` and `TestPhase30PreChangeGolden` are committed first, on their own (D-15).
- [ ] Gate tests (a)-(f) and `30-00-GATE-VERDICT.md` follow the Phase 29 shape (D-14).
- [ ] A modload filesystem-interface seam (a fake FS) exists before any graph code.

---

## Manual-Only Verifications

All phase behaviors have automated verification.

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
