---
phase: "30"
slug: "local-modules"
status: verified
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: "2026-10-12"
---

# Phase 30 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| import text -> parser | untrusted program text reaches the `import` grammar; paths cannot be written in the grammar | module path, import items |
| module name -> filesystem | the closed resolver maps a validated name under the entry root; it is the only module IO site | root-relative path, file bytes |
| imported module -> check scope | only exported, signature-only facts from another unit reach the importer | callee contracts |
| typed core -> corevalidate | a forged cross-module call must be refused by an independent visibility law | core Module/Exported/Imports facts |
| merged program -> cgen | imported functions must get injective, non-reserved C names | C identifiers |
| diagnostic -> repair driver | a repair whose unit is another file must never be applied to the entry | unit, span, replacement |
| diagnostics -> user/agent | refusals carry module names and root-relative paths only, never host paths or OS error text | diagnostic JSON and human lines |
| test seams -> production build | `...ForTest` seams default off; exported setters have no non-test callers | package-level test switches |

---

## Threat Register

Evidence for each row comes from the security audit of 2026-10-12. Test status was taken from that day's run: all 76 Phase30 tests pass, and the full `go test ./...` passes after `b1c70f03`.

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-30-01 | Tampering | no-import pre-change golden | high | mitigate | single commit 7ef4c459 (test + golden only, revision = PLAN_BASE 4ea05fd7), never regenerated; TestPhase30PreChangeGolden, TestPhase30PreChangeGoldenIsNotInert | closed |
| T-30-02 | Denial of service | FIFO import | medium | mitigate | `IsRegular()` before Open (`modload/resolve.go:368` vs `:379`); TestPhase30FIFORefusedWithoutBlocking | closed |
| T-30-03 | Tampering | in-root symlink alias | high | mitigate | per-component Lstat refusal (`resolve.go:327-335`, `:362`); TestPhase30SymlinkAndEscapeRefused | closed |
| T-30-04 | Tampering | import path text | high | mitigate | `modulePathSpan` admits only dotted identifiers (`syntax/parser.go:393-411`); `validateModulePath`; TestPhase30ImportPathSyntaxRefused | closed |
| T-30-05 | Denial of service | import fan-out per file | medium | mitigate | `MaxImportsPerModule = 16`, `module.imports_limit` at parse; TestPhase30ImportsLimitParse | closed |
| T-30-06 | Tampering | import repairs | medium | mitigate | duplicate-import delete offered only when the later items are a subset of the earlier ones (`parser.go:260-295`); TestPhase30ImportRepairs | closed |
| T-30-07 | Tampering | cross-unit repair applied to the entry | high | mitigate | unit decoded structurally and cross-unit repairs declined (`cmd/schway-repair/repair.go:354`); TestPhase30RepairDriverDeclinesCrossUnit | closed |
| T-30-08 | Spoofing | diagnostic ID collision across units | medium | mitigate | Unit is in both identity structs, `InUnit` recomputes the ID; TestPhase30UnitIdentity | closed |
| T-30-09 | Repudiation | existing diagnostic IDs changing | medium | mitigate | Unit omitempty in identity, `InUnit("")` is identity, pinned ID d34396fcc8ea6530a1f04dc3; TestPhase30GateUnitOmitemptyInvariance | closed |
| T-30-10 | Tampering | symlink escape | high | mitigate | per-component Lstat plus `os.OpenRoot` second layer (`modload/fs.go:15-17`); TestPhase30SymlinkAndEscapeRefused | closed |
| T-30-11 | Denial of service | FIFO/device hang | high | mitigate | regular-file check before Open; TestPhase30FIFORefusedWithoutBlocking | closed |
| T-30-12 | Tampering | TOCTOU swap | high | mitigate | `sameFile(Lstat, Stat)` after Open, then one bounded read (`resolve.go:383-387`); TestPhase30FakeFSToctouAndStatSize | closed |
| T-30-13 | Spoofing | case-fold / homoglyph aliasing | medium | mitigate | lowercase ASCII segments and an exact byte-equal directory-listing match (`resolve.go:191-215`, `:405-431`); TestPhase30CaseExactMatch | closed |
| T-30-14 | Denial of service | huge listing / oversized module | medium | mitigate | `MaxDirectoryEntries = 65536`, limit+1 read, size from the bytes read; TestPhase30DirectoryListingBounded, TestPhase30PerModuleByteCap | closed |
| T-30-15 | Information disclosure | host paths / OS error text | medium | mitigate | `diagnose` renders module names and root-relative paths only (`resolve.go:263-301`); TestPhase30ErrorsNeverLeakAbsolutePaths | closed |
| T-30-16 | Denial of service | fan-out and deep chains | high | mitigate | 16 modules and depth 8 checked at enqueue before open; 4 MiB total after each read (`load.go:291-339`); TestPhase30CapsExactNAndNPlusOne, TestPhase30CapPrecedence | closed |
| T-30-17 | Denial of service | quadratic corevalidate cost on merged size | high | mitigate | `module.program_limit` at 1,024 functions / 4,096 declarations (`merge.go:58-90`); TestPhase30GateMergedProgramLimit | closed |
| T-30-18 | Denial of service | recursion / cycles | medium | mitigate | iterative worklist and three-color DFS on an explicit stack; TestPhase30ImportCycles, TestPhase30DiamondLoadsOnce | closed |
| T-30-19 | Tampering | re-read between load and identity | medium | mitigate | each file read once and the entry matched by `sameFile` without reopening; TestPhase30LoaderReadsEachFileOnce | closed |
| T-30-20 | Elevation of privilege | calling another module's private function | high | mitigate | `module.not_exported` (`check/check.go:2340-2344`), scope admits exported fns only; TestPhase30PrivateNameRefused | closed |
| T-30-21 | Tampering | admission depending on callee body | medium | mitigate | `functionContract` reads signature fields only (`check.go:2086-2093`); TestPhase30CallAdmissionBodyBlind | closed |
| T-30-22 | Spoofing | name shadowing via imports | medium | mitigate | collision refusals naming both origins, no repair (`check.go:2316-2358`); TestPhase30ImportCollisions | closed |
| T-30-23 | Repudiation | import-bearing source checked as single-file | medium | mitigate | `module.imports_not_loaded` fails closed (`check.go:99-104`); TestPhase30CheckRefusesUnloadedImports | closed |
| T-30-24 | Spoofing | forged core cross-module call | high | mitigate | `visibilityAllows` at both replay sites (`corevalidate.go:2914`, `:3311`, `:5278-5290`); TestPhase30CorevalidateVisibilityPeer | closed |
| T-30-25 | Tampering | clone dropping module facts | high | mitigate | named JSON tags carry the facts through `cloneProgram`; `dropModuleFactsInCloneForTest` sits in the production path; TestPhase30ValidatedProgramKeepsModuleFacts | closed |
| T-30-26 | Elevation of privilege | library main/export hijacking the entry | medium | mitigate | entry-module filter in `EntryFunction` and entry preference in `findFunction`; TestPhase30EntryIgnoresLibraryMain, TestPhase30FindFunctionPrefersEntryModule | closed |
| T-30-27 | Tampering | C symbol collision across modules | high | mitigate | length-prefixed `moduleCName` plus `allocate` (`cgen/cgen.go:1607-1620`); TestPhase30CNamesInjective | closed |
| T-30-28 | Tampering | reserved C identifiers | medium | mitigate | the D-18 grammar refuses leading `_`, `__` and trailing `_`; names begin `SCHWAY_M`; TestPhase30SegmentGrammar | closed |
| T-30-29 | Tampering | second module IO site | high | mitigate | `loadSourcePath` is the only `modload.Load` route; `os.OpenRoot` appears only at `modload/fs.go:16`; TestPhase30SingleIOSite | closed |
| T-30-30 | Repudiation | single-file commands ignoring imports | medium | mitigate | fail-closed per command; TestPhase30SingleFileCommandsRefuseImports | closed |
| T-30-31 | Information disclosure | host paths in route results | medium | mitigate | route-level TestPhase30ErrorsNeverLeakAbsolutePaths | closed |
| T-30-32 | Denial of service | debug-map bypassing the work cap | medium | mitigate | `checkLoaded` sums Work and applies `MaxCheckWork` on the module path; TestPhase30ModuleWorkLimit, TestPhase30DebugMapNoImportUnguarded | closed |
| T-30-33 | Tampering | library copied rather than imported | high | mitigate | TestPhase30LibraryIsSingleCopy with a copy control; adler32 functions carry `Module == "adler32"` | closed |
| T-30-34 | Repudiation | overclaiming separate compilation or hashing | medium | mitigate | `## Non-claims` in `examples/phase30/README.md`; non-claim sentences in PRODUCT-ROADMAP and LANGUAGE-MATURITY | closed |
| T-30-35 | Spoofing | wrong pinned answers | medium | mitigate | `testsupport.Phase29FixtureExpected` with answers from zlib and Go literals | closed |
| T-30-SC | Tampering | package installs | high | mitigate | `go.mod`/`go.sum` unchanged since 4ea05fd7; no package manifests added | closed |

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-30-01 | T-30-24 | A forged core whose Exported/Imports facts agree with a forged call escapes the visibility law by design. This is the frontend trust boundary, documented at `corevalidate.go:5251-5257`, the same class as the Phase 29 `escape:coordinated-source-core-lie`. Phase 31 module digests (MOD-06) are the stated cross-check. | 30-06 threat model; security audit | 2026-10-12 |
| AR-30-02 | T-30-02/03/10/11/12 (scope) | The entry path is user authority, as before this phase. `modload.Load` opens the entry with `os.Open` and a bounded read, with no Lstat, regular-file or Root check. The FIFO, symlink and TOCTOU mitigations apply to imported modules, which are resolved from untrusted import text. | security audit | 2026-10-12 |

*Accepted risks do not resurface in future audit runs.*

---

## Notes (non-register observations)

- `modload/load.go:57-90` exports `Set*ForTest` setters from a non-test file so that session tests (30-08) can arm modload seams. No non-test caller exists under `internal/` or `cmd/`. This matches the existing `cgen.SetPayloadSlotSwapForTest` convention, and the seams are reachable only through the package API. Hardening these behind a test-only build mechanism would be the same follow-up as Phase 29 WR-04.

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-10-12 | 36 | 36 | 0 | /gsd-secure-phase (gsd-security-auditor, ASVS 1, mitigations verified against code) |

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-10-12
