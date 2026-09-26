---
phase: "18"
slug: "branch-on-a-computed-value"
status: verified
threats_open: 0
asvs_level: 1
created: "2026-09-25"
---

# Phase 18 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

## Trust Boundaries

| Boundary | Description | Data Crossing |
|-----------|-------------|---------------|
| `.lang` source to parser/checker | Untrusted source may be malformed or adversarial. | Source text, syntax tree, control flow |
| Checked core to independent validation peers | Checker facts are not trusted as peer evidence. | Computed place, type, ownership, origin |
| Validated core to interpreter/native emitter | Wrong dispatch or payload routing could alter program behavior. | Branch selection, return values, native output |
| Tests and mutation hooks to comparator evidence | Missing execution tiers or inert mutation instrumentation could give false confidence. | Engine outcomes and injected corruption |
| CI configuration to compiler evidence | Missing or skipped checks could remove recurring protection. | Build, vet, test, race, Clang execution |

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-------------|----------|-------------|------------|--------|
| T-18-01 | Denial of service | Parser, CFG, peer traversal, liveness | high | mitigate | Bounded parser/core inputs, cycle refusal, and fail-closed liveness bound `4 × blocks × (distinct loans + 1)`; covered by Phase 18 checker, liveness, and full-suite tests. | closed |
| T-18-02 | Tampering / Elevation of privilege | Checker and independent core/origin peers | high | mitigate | Peers independently derive computed place, type, ownership, and payload origin; forged, wrong-type, arm-local, sibling-arm, and understated-origin controls are exercised by corevalidate/originvalidate tests. | closed |
| T-18-03 | Tampering | C emitter, mutation harness, comparator, CI wiring | high | mitigate | Four execution tiers agree for valid inputs; seeded wrong-slot mutation requires positive injection and exact terminal-outcome divergence; comparator anti-controls seed each claimed axis. Full/race CI lanes retain installed Clang coverage. | closed |
| T-18-04 | Tampering | Interpreter callee frame and computed-match dispatch | high | mitigate | Computed callees execute entry-prefix operations before dispatch and bind only the selected arm's Result value place; interpreter and session regressions pass. | closed |
| T-18-05 | Tampering | Branch arm return representation and native emitter | high | mitigate | Independent validation rejects forged arm value/type/source/place combinations; native emitter regression and source-to-native session differential tests pass. | closed |

## Accepted Risks Log

No accepted risks.

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-09-25 | 5 | 5 | 0 | GSD verify-work orchestrator (ASVS L1 evidence review) |

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-09-25
**Status:** passed
