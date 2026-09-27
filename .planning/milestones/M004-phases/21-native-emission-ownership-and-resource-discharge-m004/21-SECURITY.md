---
phase: "21"
slug: "native-emission-ownership-and-resource-discharge-m004"
status: verified
# Count of open threats at or above workflow.security_block_on (high).
threats_open: 0
asvs_level: 1
created: "2026-09-27"
---

# Phase 21 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| Checked program to native emission | Public emitters accept only validated compiler products and retain fail-closed family boundaries. | Checked core programs, generated C, execution output |
| Generated C to compiler/native execution | The recorded comparison uses a named fixture, bounded compiler lanes, and explicit provenance. | Fixture bytes, emitted C, compiler/host facts, execution results |
| Archived planning records to recurring evidence checks | Phase 16 and M003 artifacts remain historical inputs; current guards verify ownership and exact identity. | Debt rows, requirement ownership, source paths, evidence receipts |
| Validation producer to checked-in evidence records | The producer is sequential and the consumer binds requested pairs, digests, and completion witnesses. | Test results, corpus records, manifests, child-process environment |

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-21-01 | Tampering | Resource-discharge contract | medium | mitigate | Contract parser and mutation tests reject malformed or over-admitting records. | closed |
| T-21-02 | Elevation | Family admission contract | high | mitigate | Contract records remain design-only; production refuses unproven families. | closed |
| T-21-03 | Elevation | Public C emitter dispatch | high | mitigate | Both public APIs share one validated lowering path; cut shapes refuse before serialization. | closed |
| T-21-04 | Tampering | Retired legacy emitters | medium | mitigate | Tests ensure the retired lowering bodies remain absent. | closed |
| T-21-05 | Tampering | Fixture/compiler evidence | medium | mitigate | Fixture and emitted-C digests retain compiler and host provenance. | closed |
| T-21-06 | Elevation | Optimized comparison claim | high | mitigate | Comparison claims stay bounded to the measured fixture, compiler, host, and lanes. | closed |
| T-21-07 | Repudiation | Archived emitter-debt decisions | medium | mitigate | Archived debt rows point to committed decisions and recurring probes. | closed |
| T-21-08 | Tampering | Derived unreachable-claims view | medium | mitigate | Tests derive and byte-compare the view against authored debt registers. | closed |
| T-21-09 | Elevation | LTO evidence scope | high | mitigate | Receipt checks reject optimizer and host claims beyond the recorded sample. | closed |
| T-21-05-01 | Tampering | Archived Phase 16 cut decision | medium | mitigate | Archive-backed guard checks the selected decision and refusal before serialization. | closed |
| T-21-05-02 | Spoofing | NAT-09 ownership register | high | mitigate | Guard checks archived ownership, current roadmap owner, and seeded mutations. | closed |
| T-21-05-03 | Elevation of privilege | Contract-to-emitter boundary | high | mitigate | Cut families remain refused; contract data is not treated as production admission or runtime proof. | closed |
| T-21-06-01 | Tampering | Corpus record and manifest | high | mitigate | Consumer-derived pairs, exact digests, and completion witnesses are checked. | closed |
| T-21-06-02 | Repudiation | Groundedness and reconciliation entries | high | mitigate | Full finding identity, obligations, measured frontier, and derived view are checked. | closed |
| T-21-06-03 | Denial of service | Corpus producer child process | medium | mitigate | Sequential execution and a bounded consumer elapsed check are present. The declared explicit child `GOCACHE` path is absent; retained as an open, below-threshold item. | open — below high threshold (non-blocking) |
| T-21-06-04 | Elevation of privilege | Evidence interpretation | high | mitigate | Historical refusal and measured-claim ceilings remain enforced. | closed |

*Only open threats at or above the configured high threshold count toward `threats_open`.*

## Accepted Risks Log

No accepted risks.

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-09-27 | 16 | 15 | 1 (medium, non-blocking) | gsd-security-auditor |

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed (no open high-or-higher threats)
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-09-27
