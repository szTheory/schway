---
phase: "31"
slug: "bounded-file-checksum-and-second-consumer"
status: verified
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: "2026-10-12"
---

# Phase 31 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| repo example tree -> module loader | which file is the importable library is decided by directory placement (Phase 30 D-05) | module source bytes |
| argv[1] PATH -> generated C -> adapter | an untrusted path chosen by the caller crosses into `open` | path bytes (any length, NUL, leading dash) |
| file content -> adapter buffer | untrusted bytes of untrusted length | file bytes, capped at 65,536 |
| adapter status -> generated stderr | the only error text a caller sees | fixed `FileError.<Alt>` line, never the path |
| path -> Go host reader | the interpreter's independent read of the same path | file bytes, capped at 65,536 |
| source text -> core program -> peers | a malformed owner shape must be refused before any C is emitted | core owner-shape facts |
| file system state / allocator -> adapter | growth after fstat, device/FIFO, mid-read error, allocation failure | descriptor state, allocation results |
| receipt file on disk -> app run | a receipt may be tampered with between build and run | `modules` rows (name, relative path, digest) |
| library module -> two consumers | one source of truth whose change must reach both builds | library bytes and identity |
| oracle literals / control list -> CI | wrong literals or unrun controls give false assurance | expected checksums, control inventory |

---

## Threat Register

Evidence for each row comes from the security audit of 2026-10-12 (L1, grep depth plus the named tests). Test status was taken from that day's run: every `TestPhase31*` test passes, and the full local `go test ./...` passes in all 27 packages at `2271c0e2`. Hosted CI has not yet run at this revision.

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-31-01 | Tampering | `examples/checksum/adler32.schway` move | medium | mitigate | `git mv` keeps the bytes; TestPhase31MovedLibraryIdentity pins the sha256 with a non-inert control | closed |
| T-31-02 | Repudiation | dated Phase 30 receipts | low | mitigate | `.planning/phases/30-local-modules` unchanged since 70710df6 (empty `git diff`) | closed |
| T-31-03 | Information disclosure | stderr on acquire failure | high | mitigate | `examples/checksum/adapter.c` has no stdio output calls; cgen prints only the fixed `FileError.<Alt>` table; TestPhase31FileErrorStderrTable, TestPhase31CaseMatrix (path segment absent) | closed |
| T-31-04 | Denial of service | adapter read on a FIFO or device | high | mitigate | `O_NONBLOCK` open (`adapter.c:78`), then `fstat` + `S_ISREG` before any read (`adapter.c:80`); TestPhase31CaseMatrix FIFO and `/dev/null` rows | closed |
| T-31-05 | Tampering | size check | high | mitigate | read-based limit+1 probe (`adapter.c:83-98`), no `st_size` reference; TestPhase31BoundedReadMutants/stat-size-mutant | closed |
| T-31-06 | Elevation of privilege | shape capture by the Phase 23 matcher | medium | mitigate | distinct plural prefix, new branch dispatched first; TestPhase31FileBytesNeighborRefusals, Phase 23/29 tests unchanged | closed |
| T-31-07 | Denial of service | `ChecksumReadFile` on a FIFO | medium | mitigate | `O_NONBLOCK` open and `IsRegular` before read (`testsupport/checksum_oracle.go:34-39`); TestPhase31CaseMatrix | closed |
| T-31-08 | Tampering | route-supplied alternative name | medium | mitigate | `RunWithFileBytes` refuses an alternative outside `Fails` (`interp/interp.go:403`) | closed |
| T-31-09 | Elevation of privilege | owner-shape admission | high | mitigate | each peer refuses each forged neighbour alone (defects found and fixed in 3cc5bb5b); TestPhase31FileBytesSeededFaults in check, corevalidate, originvalidate, pathoracle | closed |
| T-31-10 | Tampering | cgen stderr table | medium | mitigate | table derived from `program.DataTypes`, declaration order pinned (WR-02, 949f2386); TestPhase31FileErrorStderrTable, TestPhase31FileErrorDeclarationIsPinned | closed |
| T-31-11 | Tampering | TOCTOU growth after fstat | high | mitigate | limit+1 read probe; TestPhase31BoundedReadMutants/growth-after-fstat and stat-size-mutant | closed |
| T-31-12 | Denial of service | FIFO open blocking | high | mitigate | `O_NONBLOCK`; TestPhase31BoundedReadMutants/nonblocking-mutant hangs to the 5 s deadline while the faithful build returns at once | closed |
| T-31-13 | Tampering | double free or leak on error paths | high | mitigate | TestPhase31ReleaseExactlyOnce (six-path counting observer); TestPhase31BoundedReadMutants/double-release-asan | closed |
| T-31-14 | Denial of service | unbounded read of a huge file | medium | mitigate | read loop bounded at 65,537 bytes (`adapter.c:87`); TestPhase31BoundedReadMutants (1 MiB file) | closed |
| T-31-15 | Tampering | receipt `modules` rows | high | mitigate | `validateReceiptModules` checks sort, uniqueness, canonical relative paths, identity and the entry link (`native/native_app.go`, hardened by WR-03 7256e8dc); TestPhase31ValidateReceiptModulesTable, TestPhase31ModuleDigests | closed |
| T-31-16 | Information disclosure | host paths in receipts | medium | mitigate | only entry-relative slash paths are recorded; absolute, backslash, `..`, empty and `.` segments refused; TestPhase31ValidateReceiptModulesTable | closed |
| T-31-17 | Repudiation | library edit not reflected in identity | high | mitigate | a comment-only library edit flips the BuildID; TestPhase31ModuleDigests | closed |
| T-31-18 | Information disclosure | error stderr on both engines and host refusals | high | mitigate | distinctive non-ASCII path segment asserted absent on every error row; TestPhase31CaseMatrix, TestPhase31CaseMatrixNativeBoundaries | closed |
| T-31-19 | Tampering | path that looks like an option | low | mitigate | the path is `argv[1]` after `--`; the leading-dash row prints the checksum; TestPhase31CaseMatrix | closed |
| T-31-20 | Denial of service | FIFO and devices on either engine | high | mitigate | `O_NONBLOCK` + regular-file check on both readers; deadline-bound FIFO rows, `/dev/null` → FileUnavailable on both; TestPhase31CaseMatrix | closed |
| T-31-21 | Tampering | silent library fork (copy) | high | mitigate | sha256-frozen allow-list and a single `adler32` per app; TestPhase31SingleCopyAndDualFlip/copied-library-is-detected | closed |
| T-31-22 | Repudiation | stale identity after a library change | high | mitigate | dual flip moves both BuildIDs and the shared module row; TestPhase31SingleCopyAndDualFlip | closed |
| T-31-23 | Tampering | oracle rows | high | mitigate | recorded Python zlib command plus Go `hash/adler32` cross-check; TestPhase31OracleRows (armed empty-cmd control) | closed |
| T-31-24 | Repudiation | controls wired but not run | high | mitigate | go/parser inventory maps each EVD-14 control to one subtest; TestPhase31ControlInventory | closed |
| T-31-25 | Tampering | signed-char dependence | medium | mitigate | TestPhase31Controls/signed-char-retype fails as required; signed-char-flag-lane passes | closed |
| T-31-26 | Repudiation | README and register claims | medium | mitigate | every stated value names its test; groundedness and validation-lifecycle meta-tests pass; no hosted receipt claimed | closed |
| T-31-SC | Tampering | package installs | high | mitigate | `go.mod`/`go.sum` unchanged since 70710df6; no package manifests added | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on count toward threats_open*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|

No accepted risks. Code-review info items IN-02 (the 64K cap lives in the host readers, not in `RunWithFileBytes`) and IN-03 (unstable sort and duplicated prefix trimming) are deferred in 31-REVIEW-DISPOSITION.md. Neither opens a threat: both shipped readers enforce the cap, and module names are unique.

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-10-12 | 27 | 27 | 0 | /gsd-secure-phase (L1 orchestrator; register authored at plan time, auditor skipped by the short-circuit rule) |

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-10-12
