---
phase: "15"
slug: "event-identity-lang-execution-2"
status: verified
threats_open: 0
asvs_level: 1
created: "2026-09-22"
---

# Phase 15 — Security

## Trust Boundaries

| Boundary | Description | Data Crossing |
|---|---|---|
| Execution JSON → native validator | Versioned engine output is admitted before it becomes evidence. | Untrusted serialized execution document |
| Checked call DAG → native emitter | Source graph fanout can amplify during native serialization. | Checked program structure |
| Engine evidence → session verdict | Agreement is accepted only after independent structural validation. | Execution documents |
| Repository workflow → hosted CI | Checked-in workflow commands select recurring cross-platform evidence. | CI configuration |

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|---|---|---|---|---|---|---|
| T-15-01 | Denial of Service | peer/emitter invocation expansion | high | mitigate | 4096-node bounds in `executionpeer` and native preflight, with 4096/4097 controls | closed |
| T-15-02 | Tampering | invocation parser | high | mitigate | Canonical formatter/parser round trip and alternate-spelling rejection | closed |
| T-15-03 | Tampering | causal evidence and peer gate | high | mitigate | Caller-owned call events plus independent `executionpeer.Validate` before Schema 2 comparison | closed |
| T-15-04 | Repudiation | schema and evidence routing | high | mitigate | Explicit Schema 2 route, frozen legacy bytes, peer/comparator and four-tier controls | closed |
| T-15-G1-01 | Tampering | Schema 2 native admission | high | mitigate | Canonical wire-seam test and class-specific `native.invalid_execution` refusals | closed |
| T-15-G1-02 | Repudiation | Plan 01 coverage declaration | medium | mitigate | Machine-readable all-automated coverage references the native admission seam | closed |
| T-15-G2-01 | Tampering | program-aware comparator | high | mitigate | Direct legacy peer-bypass/equivalence regression and retained Schema 2 peer refusal | closed |
| T-15-G2-02 | Repudiation | CI evidence provenance | high | mitigate | Cross-platform aggregate names all five current Phase 15 seams and is source-pinned by test | closed |
| T-15-SC | Tampering | dependency boundary | low | accept | No package-manager install or dependency was added | closed |
| T-15-G1-SC | Tampering | dependency supply chain | low | accept | Gap closure added no dependency | closed |
| T-15-G2-03 | Denial of Service | recurring aggregate CI | low | accept | Five focused seams supplement, rather than duplicate, the retained baseline | closed |
| T-15-G2-SC | Tampering | GitHub Action/package supply chain | low | accept | No action, package, or dependency version changed | closed |

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---|---|---|---|---|
| AR-15-01 | T-15-SC | No dependency change occurred in Phase 15. | Project policy | 2026-09-22 |
| AR-15-02 | T-15-G1-SC | The Schema 2 admission closure introduced no dependency. | Project policy | 2026-09-22 |
| AR-15-03 | T-15-G2-03 | Focused evidence stays bounded while baseline CI remains intact. | Project policy | 2026-09-22 |
| AR-15-04 | T-15-G2-SC | The CI closure changed no action or dependency version. | Project policy | 2026-09-22 |

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|---|---|---|---|---|
| 2026-09-22 | 12 | 12 | 0 | gsd-secure-phase (ASVS L1) |

## Sign-Off

- [x] All threats have a disposition.
- [x] Accepted risks are documented.
- [x] `threats_open: 0` confirmed.
- [x] `status: verified` set in frontmatter.

**Approval:** verified 2026-09-22
