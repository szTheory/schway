---
phase: "16"
slug: "branch-match-emitter-port"
status: verified
threats_open: 0
asvs_level: 1
created: "2026-09-24"
---

# Phase 16 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| Checked compiler facts → emitted native C | Admission, branch/match lowering, and resource serialization must use validated semantic facts. | Source-derived program structure, checked places, layouts, event data |
| Host probe → admission decision | Host availability and restrict evidence must not be mistaken for a proven-safe lowering. | Probe result, target/compiler facts, refusal decision |
| Planning/evidence records → compiler gates | Corpus, manifests, and ownership records must be tied to current executable witnesses and digests. | Test names, fixture/program/artifact digests, ownership metadata |

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| 16-01-PLAN:T-16-01 | Tampering | `emitProgram` admission order | high | mitigate | Preserve graph/entry → shape → preflight → serialization and exercise the direct tracer. (verified by gsd-security-auditor) | closed |
| 16-01-PLAN:T-16-02 | Repudiation | resource document tail | medium | mitigate | Derived-value mutation control proves the serialized field is wired. (verified by gsd-security-auditor) | closed |
| 16-02-PLAN:T-16-01 | Tampering | branch admission/order | high | mitigate | Ordered stage tests plus mutation control before serialization. (verified by gsd-security-auditor) | closed |
| 16-02-PLAN:T-16-03 | Information disclosure | event attribution | low | accept | Accepted risk recorded in Accepted Risks Log: Events contain only existing checked IDs/places under frozen schema-2 rules. | closed |
| 16-03-PLAN:T-16-04 | Tampering | payload field lowering | high | mitigate | Consume checker layout facts and mutation-test the connection. (verified by gsd-security-auditor) | closed |
| 16-03-PLAN:T-16-05 | Denial of service | defect terminal | medium | mitigate | Preserve event-before-abort and narrow `_Noreturn` structural control. (verified by gsd-security-auditor) | closed |
| 16-04-PLAN:T-16-06 | Tampering | restrict admission evidence | high | mitigate | Immutable exact-shape source, dual-host lanes, exhaustive extension refusals. (verified by gsd-security-auditor) | closed |
| 16-04-PLAN:T-16-07 | Repudiation | unavailable host result | high | mitigate | Three-state result; unavailable cannot satisfy admission. (verified by gsd-security-auditor) | closed |
| 16-05-PLAN:T-16-06 | Elevation of privilege | by-pointer admission | high | mitigate | Blocking-human checkpoint with mandatory dual-host/fence evidence. (verified by gsd-security-auditor) | closed |
| 16-06-PLAN:T-16-06 | Elevation of privilege | pointer-shape admission | high | mitigate | Decision-linked predicate and exhaustive refusal test. (verified by gsd-security-auditor) | closed |
| 16-06-PLAN:T-16-08 | Repudiation | NAT-09 amendment/debt | high | mitigate | Family bijection, substantive fields, seeded mismatch control. (verified by gsd-security-auditor) | closed |
| 16-07-PLAN:T-16-09 | Tampering | convergence bytes | high | mitigate | Exact both-mode differential on three tracked fixtures. (verified by gsd-security-auditor) | closed |
| 16-07-PLAN:T-16-10 | Repudiation | golden ledger | high | mitigate | Four-entry bijection, fault table, executable witnesses. (verified by gsd-security-auditor) | closed |
| 16-07-PLAN:T-16-11 | Tampering | semantic comparator | high | mitigate | Independent four-tier comparison and seeded semantic fault. (verified by gsd-security-auditor) | closed |
| 16-08-PLAN:T-16-12 | Elevation of privilege | dispatch authority | high | mitigate | Blocking-human decision after three machine-verifiable gates. (verified by gsd-security-auditor) | closed |
| 16-09-PLAN:T-16-09 | Tampering | public emitter dispatch | high | mitigate | Atomic history and focused cgen/core gates prove the single schema-2 authority. (verified by gsd-security-auditor) | closed |
| 16-10-PLAN:T-16-10a | Tampering | schema-2 projection | high | mitigate | Mutation tests reject actual-to-expected copying. (verified by gsd-security-auditor) | closed |
| 16-10-PLAN:T-16-10b | Repudiation | consumer registry | high | mitigate | AST/path-aware exhaustive inventory is fail-closed. (verified by gsd-security-auditor) | closed |
| 16-11-PLAN:T-16-11 | Tampering | historical artifact registry | high | mitigate | Digest, provenance, refusal, and manifest-bijection tests fail closed. (verified by gsd-security-auditor) | closed |
| 16-12-PLAN:T-16-12 | Elevation | Phase 4/5 session helpers | high | mitigate | Registry, refusal-first routing, and witness citation enforcement prevent a hidden production path. (verified by gsd-security-auditor) | closed |
| 16-13-PLAN:T-16-13a | Repudiation | validation record | high | mitigate | Command, result, revision, witness, and digest fields are required and validated. (verified by gsd-security-auditor) | closed |
| 16-13-PLAN:T-16-13b | Tampering | cut evidence linkage | high | mitigate | Registry/witness/manifest cross-checks and mutation tests fail closed. (verified by gsd-security-auditor) | closed |
| 16-14-PLAN:T-16-G14-01 | Elevation of privilege | `Phase16ControlNativeC` | high | mitigate | Strict public-emitter delegation plus behavioral refusal equality test. (verified by gsd-security-auditor) | closed |
| 16-14-PLAN:T-16-G14-02 | Tampering | file-backed frozen evidence | high | mitigate | Canonical checked-program digest, source digest, refusal identity, and artifact digest must all match. (verified by gsd-security-auditor) | closed |
| 16-14-PLAN:T-16-G14-03 | Repudiation | production control results | high | mitigate | AST non-bypass inventory and injected fallback mutation prevent historical C from being reported as live emission. (verified by gsd-security-auditor) | closed |
| 16-15-PLAN:T-16-G15-01 | Tampering | consumer registry | high | mitigate | Bidirectional AST/registry comparison with unique exact identities. (verified by gsd-security-auditor) | closed |
| 16-15-PLAN:T-16-G15-02 | Repudiation | mutation controls | medium | mitigate | Error-class-specific stale, duplicate, missing, and witness mutations exercise the shared validator. (verified by gsd-security-auditor) | closed |
| 16-16-PLAN:T-16-16-01 | Tampering | Frozen evidence and manifest | medium | mitigate | Bind fixture, canonical program, artifact, and exact refusal; seed mutation failures. (verified by gsd-security-auditor) | closed |
| 16-16-PLAN:T-16-16-02 | Repudiation | Phase 11 refusal witness | low | mitigate | Verify the current named by-pointer refusal before returning frozen bytes. (verified by gsd-security-auditor) | closed |
| 16-17-PLAN:T-16-17-01 | Tampering | Corpus run record and manifest | medium | mitigate | Derive artifacts from executed producer output and independently recompute both digests. (verified by gsd-security-auditor) | closed |
| 16-17-PLAN:T-16-17-02 | Repudiation | Completion witnesses | low | mitigate | Require one completion witness per pair and one final batch witness. (verified by gsd-security-auditor) | closed |
| 16-18-PLAN:T-16-18-01 | Repudiation | R2b finding ownership | medium | mitigate | Require explicit P20 landing ownership for every retained R2b row. (verified by gsd-security-auditor) | closed |
| 16-18-PLAN:T-16-18-02 | Tampering | Pinned frontier | medium | mitigate | Preserve exact set equality and seeded mutation controls against current scanner output. (verified by gsd-security-auditor) | closed |
| 16-19-PLAN:T-16-19-01 | Repudiation | Debt witnesses | low | mitigate | Run current exact test names through debt and generated-view validators. (verified by gsd-security-auditor) | closed |
| 16-19-PLAN:T-16-19-02 | Tampering | Machine facts and budget verdict | medium | mitigate | Separate deterministic injected assertions from bounded probe fixtures; preserve typed failures and declared machine identity checks. (verified by gsd-security-auditor) | closed |
| 16-20-PLAN:T-16-20-01 | Tampering | Maturity snapshot | low | mitigate | Retain independent AST/source count verification and assert the updated snapshot in the focused test. (verified by gsd-security-auditor) | closed |
| 16-21-PLAN:T-16-21-01 | Tampering | Frozen evidence and consumer registry | medium | mitigate | Verify all provenance links and seed faults at each registry/manifest edge. (verified by gsd-security-auditor) | closed |
| 16-21-PLAN:T-16-21-02 | Repudiation | Phase 11 gate result | low | mitigate | Keep live structural controls and named executable refusal witnesses. (verified by gsd-security-auditor) | closed |
| 16-22-PLAN:T-16-22-01 | Spoofing | Fixture discovery | low | mitigate | Store the evidence input outside `.lang` discovery and test the canonical manifest identity. (verified by gsd-security-auditor) | closed |
| 16-22-PLAN:T-16-22-02 | Tampering | Fixture/provenance references | medium | mitigate | Update every registry and test path while checking source/program/artifact digests remain fixed. (verified by gsd-security-auditor) | closed |
| 16-23-PLAN:T-16-23-01 | Tampering | Validation evidence | medium | mitigate | Re-execute package-pattern pairs and independently verify both digests and witnesses. (verified by gsd-security-auditor) | closed |
| 16-23-PLAN:T-16-23-02 | Tampering | Protocol site inventory | low | mitigate | Retain exact file and total checks after accounting for the declared helper removal. (verified by gsd-security-auditor) | closed |
| 16-24-PLAN:T-16-24-01 | Tampering | Validation corpus selector | low | mitigate | Filter by explicit frontmatter status and exercise draft/validated controls. (verified by gsd-security-auditor) | closed |
| 16-24-PLAN:T-16-24-02 | Spoofing | Phase 4 evidence row | medium | mitigate | Cite only existing tests and state the active post-M004 behavior accurately. (verified by gsd-security-auditor) | closed |
| 16-24-PLAN:T-16-24-03 | Tampering | Run record | medium | mitigate | Re-export live pairs and execute the existing sequential producer with completion witnesses. (verified by gsd-security-auditor) | closed |
| 16-25-PLAN:T-16-25-01 | Repudiation | NAT-09 ownership | medium | mitigate | Name P21 in roadmap and requirements authority and point all three rows to it. (verified by gsd-security-auditor) | closed |
| 16-25-PLAN:T-16-25-02 | Tampering | Debt ownership guard | medium | mitigate | Assert exact P21 ownership and kill a seeded UNOWNED mutation. (verified by gsd-security-auditor) | closed |
| 16-26-PLAN:T-16-26-01 | Repudiation | NAT-09 verification evidence | medium | mitigate | Keep both named tests executable and run the focused assertions after the selector edit. (verified by gsd-security-auditor) | closed |
| 16-26-PLAN:T-16-26-02 | Tampering | Exact R2b frontier ownership | medium | mitigate | Run both groundedness controls and retain the current frontier/landing map unless a real finding is diagnosed. (verified by gsd-security-auditor) | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on count toward threats_open*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-16-01 | 16-02-PLAN.md:T-16-03 | Event attribution is limited to existing checked IDs and places under frozen schema-2 rules; the auditor found no implementation mitigation gap. | Recorded Phase 16 plan disposition | 2026-09-24 |

*Accepted risks do not resurface in future audit runs.*

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-09-24 | 49 | 49 | 0 | gsd-security-auditor (ASVS L1) |

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-09-24
