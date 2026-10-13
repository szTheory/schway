---
phase: "28"
slug: "exit-classes-verify-report-and-entry-pruning"
status: verified
threats_open: 0
asvs_level: 1
created: "2026-10-13"
---

# Phase 28 — Security

> retroactive, no new tests. This register reconstructs the Phase 28 trust boundaries from its review, disposition, and verification records; it does not claim a new audit or hosted run.

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| Source text to exit-class mapping | A source refusal must remain distinguishable from a tool failure. | diagnostics and exit status |
| Report path to filesystem | Refused or oversized runs remove stale reports; symlink removal affects the link. | report path and report bytes |
| Relative path to project-root resolution | Native execution must preserve the caller's relative path through admission. | source path |
| Entry pruning to emitted C | Removing unreachable functions must preserve reachable-program bytes. | typed program and C output |
| Emitter registry to source | The public emitter inventory pins declared consumers to actual call sites. | source location metadata |

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-28-01 | Tampering | Source refusal and exit mapping | high | mitigate | `errors.Join` preserves source-refusal identity; `TestAppVerifyRefusalSurvivesReportWriteFailure` and uniform exit-class tests | closed |
| T-28-02 | Information disclosure | Verification report path | medium | mitigate | stale report removal and symlink-safe path handling; `TestAppVerifyRemovesStaleReportOnRefusal` | closed |
| T-28-03 | Spoofing | Relative native source path | medium | mitigate | caller path reaches admission; `TestPhase28NativeCommandRelativePath` and relative-path controls in `TestExitClassesUniformAcrossSurfaces` | closed |
| T-28-04 | Tampering | Entry pruning and emitted C | high | mitigate | reachable output identity is pinned by `TestPhase28PrunedEmitMatchesGolden`, `TestPhase272AcceptedCorpusByteIdentity`, and `TestPhase273ConformanceCorpusByteIdentity` | closed |
| T-28-05 | Repudiation | Public emitter inventory | low | mitigate | registry correction in commit `eface8a`; `TestPhase16PublicEmitterConsumerInventory` | closed |

Empty threat categories are not implied by omission: the source records no separate spoofing, elevation-of-privilege, or denial-of-service findings beyond the boundaries above.

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Disposition | Date |
|---------|------------|-----------|-------------|------|
| AR-28-01 | WR-01 | The line-number registry warning is a process finding; its structural symbol-based replacement remains future work. | deferred (per `28-REVIEW-DISPOSITION.md`) | 2026-10-11 |

## Notes

`28-VERIFICATION.md` reports hosted CI was not run for that revision. The ambiguous-entry binary observation hid its numeric exit code behind a pipe; automated tests pin exit 2 instead.

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-10-13 | 5 | 5 | 0 | Phase 33 retroactive reconstruction |

## Sign-Off

Retroactive, no new tests. Existing named tests and archived review-fix evidence support the mitigations above.

- [x] Each listed threat has a disposition.
- [x] Deferred process risk is recorded as deferred.
- [x] `threats_open: 0` reflects this register's thresholded open count.
- [x] No new tests, hosted CI, or fresh audit are claimed.

**Approval:** reconstructed 2026-10-13
