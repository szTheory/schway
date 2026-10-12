---
phase: 30-local-modules
verified: 2026-10-12T00:00:00Z
status: passed
score: 4/4 must-haves verified
covered_files: [".planning/phases/30-local-modules/30-00-PLAN.md",".planning/phases/30-local-modules/30-00-SUMMARY.md",".planning/phases/30-local-modules/30-01-PLAN.md",".planning/phases/30-local-modules/30-01-SUMMARY.md",".planning/phases/30-local-modules/30-02-PLAN.md",".planning/phases/30-local-modules/30-02-SUMMARY.md",".planning/phases/30-local-modules/30-03-PLAN.md",".planning/phases/30-local-modules/30-03-SUMMARY.md",".planning/phases/30-local-modules/30-04-PLAN.md",".planning/phases/30-local-modules/30-04-SUMMARY.md",".planning/phases/30-local-modules/30-05-PLAN.md",".planning/phases/30-local-modules/30-05-SUMMARY.md",".planning/phases/30-local-modules/30-06-PLAN.md",".planning/phases/30-local-modules/30-06-SUMMARY.md",".planning/phases/30-local-modules/30-07-PLAN.md",".planning/phases/30-local-modules/30-07-SUMMARY.md",".planning/phases/30-local-modules/30-08-PLAN.md",".planning/phases/30-local-modules/30-08-SUMMARY.md",".planning/phases/30-local-modules/30-09-PLAN.md",".planning/phases/30-local-modules/30-09-SUMMARY.md","examples/phase30/adler32.schway","examples/phase30/adler32_cli.schway"]
covered_digest: "v3:sha256:2677409dd402bd76c6e14c5a1f51436ebe8ffa3d14b74bb672230ce81e612372"
behavior_unverified: 0
overrides_applied: 0
---

# Phase 30: Local Modules Verification Report

**Phase Goal:** A Schway program can import another local source file as a module through an explicit import and export list, with resolution closed under the entry directory and every module-graph error refused by name.
**Status:** passed
**Re-verification:** No (initial verification)

## Goal Achievement

### Observable Truths (ROADMAP success criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Entry imports `adler32` from a separate module file and prints the pinned Phase 29 answers on both engines; `..`, absolute and symlink escapes refused; no search path | VERIFIED | `examples/phase30/adler32.schway` (`module adler32`, export list) imported by `adler32_cli.schway`; `check` passes. TestPhase30AdlerModuleAppCLI (native, with mutated-library control) and TestPhase30AdlerModuleInterpreter PASS. Live probes: symlink to outside refused `module.symlink`; `..` and absolute paths refused at the grammar (`syntax.expected_*`, exit 2). TestPhase30SymlinkAndEscapeRefused, ImportPathSyntaxRefused, FIFO, CaseExactMatch, SingleIOSite PASS. |
| 2 | Non-exported name refused with a stable code; corevalidate re-derives visibility; cross-module admission is body-blind | VERIFIED | Live probe: `module.not_exported` ("module g does not export priv"). TestPhase30PrivateNameRefused, CorevalidateVisibilityPeer, CallAdmissionBodyBlind, ScopedContractReadsNoBody, CallableFlipIsASignatureChange PASS. |
| 3 | Self, 2-cycle and 3-cycle, missing, duplicate, header/path mismatch and collision each refused with own named diagnostic attributed to importing source; imported-file diagnostics name that file | VERIFIED | Live probes: `module.import_cycle` (self: c -> c; 2-cycle: a -> b -> a, `unit=b`), `module.not_found`, `module.path_mismatch`, `module.import_collision`, and `core.call_callee_unresolved unit=m` for an error inside an imported file. TestPhase30ImportCycles, GraphRefusals, ImportCollisions, ImportedFileDiagnosticNamesFile, PeerRefusalNamesUnit PASS. The 3-cycle and duplicate cases are covered by the tests above, not by my own probes. |
| 4 | Diamond loads and emits shared module once; no-import programs byte-identical C pinned by golden committed first; C names injective | VERIFIED | TestPhase30DiamondLoadsAndEmitsOnce, DiamondLoadsOnce, PreChangeGolden (+ NotInert), CNamesInjective, EmitOrderIndependent, ReversedImportOrderIdenticalC PASS. Golden commit 7ef4c459 precedes the first feature commit f3899d2b. |

**Score:** 4/4 truths verified (0 behavior-unverified)

### Requirements Coverage

| Requirement | Source Plans | Status | Evidence |
|-------------|--------------|--------|----------|
| MOD-03 | 30-00, 01, 02, 03, 05, 08, 09 | SATISFIED | Closed resolver, caps at N / N+1, imported-module header check (entry header exempt, per the dated D-06 amendment in REQUIREMENTS.md). |
| MOD-04 | 30-01, 04, 05, 07, 08, 09 | SATISFIED | Export-list visibility, corevalidate peer, signature-only admission. |
| MOD-05 | all ten plans | SATISFIED | Named graph refusals, diamond once, no-import byte identity, file-named diagnostics. |

All three IDs appear in PLAN frontmatter and in REQUIREMENTS.md (marked Complete, mapped to Phase 30). No orphaned requirements.

### Review and Gate Artifacts

- 30-REVIEW: critical CR-01 plus WR-01..WR-06 all `fixed` in 30-REVIEW-DISPOSITION; IN-01 and IN-02 deferred (known, tracked).
- 30-SECURITY: `threats_open: 0`. 30-VALIDATION: `validated`, `nyquist_compliant: true`.

### Behavioral Spot-Checks and Test Runs

| Check | Result |
|-------|--------|
| `go build ./...` and `go vet ./...` | PASS |
| `go test ./... -run Phase30 -count=1 -v` | 82 PASS lines (including subtests), 0 FAIL, 0 SKIP at the top level in my run |
| Live CLI probes for 8 module refusals plus an escape-by-symlink | all refused with the expected named code, exit 2 |

### Anti-Patterns

None found in the phase's example files. Known, tracked non-gaps: formatter var-body defect (Phase 29, pending todo), `debugmap.Build` panics on 5 committed sources (pinned in the golden), explain/query unknown-ID limitation on import-bearing sources, review IN-01 and IN-02.

### External-Blocker Note (not a gap)

Hosted full and race CI on ubuntu-latest and macos-latest has not run for Phase 30. No run ID is recorded. Per 30-TRANSITION-CHECK section 5 and plan 30-09, this is an external-blocker note only. No gap plan is opened; M006 receipts belong to Phase 32 (EVD-15). The case-sensitive-filesystem row of TestPhase30CaseExactMatch skips by design on APFS and is expected to run on Linux CI.

### Human Verification Required

None. Every truth has automated evidence (PROJECT.md zero-human-UAT preference).

### Gaps Summary

No gaps. The phase goal is achieved in the codebase.

---

_Verified: 2026-10-12_
_Verifier: Claude (gsd-verifier)_
