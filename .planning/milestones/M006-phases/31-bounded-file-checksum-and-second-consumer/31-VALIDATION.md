---
phase: "31"
slug: "bounded-file-checksum-and-second-consumer"
# status lifecycle: planned (plan-phase; TestPhase20ValidationLifecycle forbids committed drafts) → validated (closure plan / validate-phase)
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-10-12"
---

# Phase 31 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> The source of truth for each test is `31-RESEARCH.md` § Validation Architecture.
> Zero human UAT is required (PROJECT.md § Verification Operating Preference; 31-CONTEXT D-15).

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (Go 1.24.0 stdlib `testing`). Native lanes need clang. The ASan, mkfifo and non-root rows are gated by cited `env:` skips. |
| **Config file** | none (go.mod only) |
| **Quick run command** | Run each touched package with the flags `-run 'Phase31' -count=1`. Packages: check, corevalidate, originvalidate, pathoracle, cgen, interp, native, session, testsupport, cmd/schway. |
| **Session meta set** | `go test ./internal/compiler/session -run 'Inventory\|Groundedness\|Reconciliation\|ValidationLifecycle\|TestPhase28EntryExitClassManifest\|TestPhase272AcceptedCorpusByteIdentity\|TestPhase273ConformanceCorpusByteIdentity\|TestPhase30PreChangeGolden\|TestLanguageMaturityCountsAreCurrent\|TestExitCodeDocsTable\|TestNoSuppressionOutlivesItsWitness\|TestPhase16PublicEmitterConsumerInventory' -count=1` |
| **Convention set** | `go test ./internal/compiler/native -run 'TestSourceNeverSpawnsUnboundedProcesses' -count=1`, `go test ./internal/compiler/check -run 'TestSingleLoanLivenessLaw' -count=1`, `go test ./internal/compiler/exitcode -count=1` |
| **Full suite command** | Hosted CI only: `go test ./...` and `go test -race -timeout=20m ./...` on ubuntu-latest and macos-latest |
| **Estimated runtime** | About 5–60 s per focused package run |

---

## Sampling Rate

- **After every task commit:** run the focused `-run 'Phase31'` command for each package the task touched. Also run the session meta set whenever the task touches any of:
  - a `.schway` file;
  - `session.go`, `evidence.go` or `native_app.go`;
  - a register;
  - a planning doc that lists counts.
- **After every plan wave:** run the full quick command, the session meta set and the convention set.
- **Before `/gsd-verify-work`:** the hosted full and race CI runs must be green on both OSes, with their run IDs recorded.
- **Max feedback latency:** 60 s per focused run.

---

## Per-Task Verification Map

There is one row per success criterion or decision; the plan-checker maps tasks to rows. Test names are the ones that exist after execution (verified by `TestVerificationGroundedness`). Every row must have an armed control that has been seen to fail.

| Ref | Requirement | Secure Behavior | Test | Armed control | Status |
|-----|-------------|-----------------|------|---------------|--------|
| SC1 / D-05 / D-06 | APP-11 | Through `build` + `app run -- PATH`: empty→1, 0x00→65537, Wikipedia→300286872, 65,536×0xFF→2006388466 | `TestPhase31ChecksumPublicRoute` | mutated library flips the outputs | ✅ green |
| SC1 / D-07 | APP-11 | The same answers on the interpreter, from an independent host read | `TestPhase31ChecksumInterpreter` | the reader uses stat size, or skips IsRegular | ✅ green |
| SC2 / D-08 / D-09 / D-17 | APP-12 | 65,537 bytes → `FileTooLarge`. Missing, directory, FIFO, `/dev/null` and mode-000 → `FileUnavailable`. Exit 65, stdout empty, fixed stderr with no path, identical across engines | `TestPhase31CaseMatrix`, `TestPhase31ExpectedMatrixJSON` | the S_ISREG check deleted (`/dev/null` → 1) and the `>=`/`>` boundary mutant, both in `TestPhase31BoundedReadMutants` | ✅ green |
| SC2 / D-10 | APP-12 | Released exactly once on success, oversize, AllocFailed and EIO | `TestPhase31ReleaseExactlyOnce` | a double-release mutant on the oversize path is seen to fail under ASan | ✅ green |
| SC2 | APP-12 | The size check is a limit+1 read and never uses st_size, and growth after fstat is caught | `TestPhase31BoundedReadMutants` | an st_size-trusting mutant | ✅ green |
| SC3 / D-01 / D-03 / D-04 / D-16 | MOD-06 | Single copy (allow-list pinned by sha256); each app has exactly one `adler32`; `adler_low` Wikipedia→920; dual flip; reverse B-only control | `TestPhase31SingleCopyAndDualFlip`, `TestPhase31AdlerLowPublicRoute` | a copied library is detected; the B-only mutation leaves `adler_low` unflipped | ✅ green |
| SC3 / D-11 / D-19 | MOD-06 | `Modules[]` (every module, sorted) appears in identity, receipt and evidence; independent of order; no-import bytes unchanged | `TestPhase31ModuleDigests` | a comment-only library edit flips BuildID but not EmittedCDigest | ✅ green |
| SC4 / D-12 | EVD-14 | Every oracle row carries `cmd` and `python`; Go `hash/adler32` agrees | `TestPhase31OracleRows` | a row with empty cmd fails the meta-test | ✅ green |
| SC4 / D-13 | EVD-14 | Wrong modulus, wrong initial state, swapped halves, off-by-one bound (exit 65), signed-char retype, deleted bounds check (ASan, non-empty input), refused loop neighbours and library mutation are each seen to fail; the `-fsigned-char` flag lane passes | `TestPhase31Controls`, `TestPhase31ControlInventory` and `TestPhase31LoopNeighborRefusals` | the inventory maps each control to a subtest one-to-one | ✅ green |
| D-06 / D-16 | APP-11 / MOD-06 | A seeded fault in each peer for the third shape (check, corevalidate, originvalidate, pathoracle), with named refusals for neighbouring shapes | `TestPhase31FileBytesSeededFaults` in each peer package, `TestPhase31FileBytesNeighborRefusals` (check) and `TestPhase31FileErrorStderrTable` (cgen) | one seeded fault per peer | ✅ green |
| D-01 | MOD-06 | Moved files are hash-identical apart from the header, and the prechange golden bytes are unchanged | `TestPhase31MovedLibraryIdentity`, `TestPhase30PreChangeGolden`, `TestPhase30LibraryIsSingleCopy` (re-pinned) | the existing non-inert control | ✅ green |
| D-14 | APP-11 | The cap case runs on every race-enabled hosted run, with no `-short` or env gate | covered by `TestPhase31ChecksumPublicRoute` and `TestPhase31ChecksumInterpreter` | n/a (no gate exists) | ✅ green |
| D-08 | APP-12 | The case table in `checksum.expected.json` matches the matrix, `exitcode.AppInputRefused` and the `FileError` alternative order | `TestPhase31ExpectedMatrixJSON` | a table row that disagrees with the matrix fails | ✅ green |
| Security | APP-12 | A distinctive path segment never appears in stderr on any error row, including host refusals; a leading `-` path is safe | `TestPhase31CaseMatrix` | the emitter echoes argv | ✅ green |

*Status: ✅ green · ❌ red · ⚠️ flaky (no row is open)*

---

## Wave 0 Requirements

- [x] Every `TestPhase31*` file above.
- [x] `internal/compiler/testsupport` oracle rows with `cmd` and `python` fields (D-12).
- [x] An adapter fault-hook harness helper (OPEN, READ, MALLOC, FREE).
- [x] A `root` token added to the closed `env:` skip set (D-20).
- [x] The D-01 move and re-pin, landed atomically before any new example file.

---

## Manual-Only Verifications

All phase behaviors have automated verification.

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 60s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-10-12 (31-09)
