---
phase: 31
review: 31-REVIEW.md
titles: json
findings:
  - id: WR-01
    severity: warning
    disposition: fixed
    title: "Checksum adapter open lacks O_NOCTTY and O_CLOEXEC"
  - id: WR-02
    severity: warning
    disposition: fixed
    title: "Generated C maps acquire status N to FileError.Alternatives[N-1] without pinning the declaration order"
  - id: WR-03
    severity: warning
    disposition: fixed
    title: "validateReceiptModules does not reject empty or '.' segments, drive prefixes, or duplicate paths"
  - id: IN-01
    severity: info
    disposition: skipped
    title: "close() failure after a complete read is reported as FileUnavailable"
  - id: IN-02
    severity: info
    disposition: deferred
    title: "The 64K limit lives only in the host reader; interp.RunWithFileBytes does not enforce it"
  - id: IN-03
    severity: info
    disposition: deferred
    title: "Module-row sort is not stable and the entryPrefix trimming is hand-rolled in two packages"
open: 0
total: 6
recorded: 2026-10-12T00:00:00Z
---

# Phase 31: Code Review Disposition

| Finding | Severity | Disposition | Source |
|---------|----------|-------------|--------|
| WR-01 | warning | fixed | 85a0fe39. The `require-open-flags` observer mode backs `TestPhase31BoundedReadMutants/open-flags-mutant`. |
| WR-02 | warning | fixed | 949f2386. A FileError declaration that is not exactly FileTooLarge, FileUnavailable, AllocFailed is refused, in check (`check.file_bytes_error_alternatives`) and independently in cgen. The orchestrator auto-followed refusal over a name-keyed table: the adapter status codes are a fixed ABI order and the shape is closed. |
| WR-03 | warning | fixed | 7256e8dc. Covered by `TestPhase31ValidateReceiptModulesTable`. |
| IN-01 | info | skipped | Design choice, and the Go reader mirrors it. A failed close after a full read means the file could not be read reliably, so it is FileUnavailable under D-09. |
| IN-02 | info | deferred | The cap is a property of the host byte route. Both shipped routes, the adapter and `testsupport.ChecksumReadFile`, enforce it, and `TestPhase31CaseMatrix` pins both. Enforce it inside `RunWithFileBytes` if a third caller appears. |
| IN-03 | info | deferred | Module names are unique, which receipt validation refuses otherwise, so sort stability cannot change the output. Consolidate the duplicated trimming the next time either package is touched. |

Dispositions: `open` (recorded, not yet triaged), `fixed`, `skipped`, `deferred`.
Set `deferred` by hand and put the reason in the Source cell; both are preserved. A `|` in the reason is kept as prose and escaped on the next run.
The fixes were applied by gsd-code-fixer, scope critical+warning. Details are in 31-REVIEW-FIX.md. The focused Phase 23/28/29/30/31 tests and the session meta set were re-run green on the merged branch.
