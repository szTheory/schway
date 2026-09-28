---
phase: "23-live-local-allocation-and-discharge"
slug: "live-local-allocation-and-discharge"
status: verified
threats_open: 0
asvs_level: 1
created: "2026-09-28"
---

# Phase 23 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| Caller argv → generated entry | Untrusted path bytes enter one retained native process. | Bounded path token |
| Checked foreign declaration → C adapter | A declared ABI grants trusted local C build authority but does not prove arbitrary C behavior. | Symbols, ABI facts, local C inputs |
| Adapter → Lang owner | Pointer, length, status, and allocator pairing cross the foreign memory boundary. | Resource identity, byte, failure status |
| Candidate core → independent peers | Candidate core may omit or forge cleanup while retaining plausible metadata. | Acquisition, operation, release facts |
| Generated C → libc allocator | Pointer identity and lifetime become physical process facts. | Allocation, use, release events |
| Compiler event stream → evidence report | Semantic records may be plausible despite incorrect native calls. | Events and separate observer receipt |
| Checked-in example → public CLI / CI host → evidence claim | Documentation or a skipped host lane must not overstate the exercised route. | Commands, expected answers, host/compiler/target receipt |

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-23-01 | Tampering | argv and generated entry | high | mitigate | Direct argv execution; bounded byte length; NUL refusal; no shell or C-text interpolation. | closed |
| T-23-02 | Denial of service | adapter file read | high | mitigate | Regular-file check; one-byte read plus EOF probe; EINTR retry; bounded diagnostics. | closed |
| T-23-03 | Elevation of privilege | owner pointer and release ABI | high | mitigate | Pointer-bearing ABI conformance; noncopyable owner; exact allocator/destructor pairing; pre-serialization shape refusal. | closed |
| T-23-04 | Tampering | core release list | high | mitigate | Seed obligations from successful acquisition; replay terminal paths; reached all-release-deleted control. | closed |
| T-23-05 | Elevation of privilege | forged contract and allocator identity | high | mitigate | Recompute operation ABI and pairing in source-blind peers; reject wrong owner and fabricated release. | closed |
| T-23-06 | Repudiation | compiler event stream | medium | mitigate | Keep events plausible in mutation tests so independent peer verdict carries the proof. | closed |
| T-23-07 | Spoofing | model evidence | high | mitigate | Model-only receipt disclaims host IO and physical cleanup; report tests enforce those fields. | closed |
| T-23-08 | Tampering | supplied foreign outcomes | medium | mitigate | Check outcomes against per-operation result, failure, mode, and owner with independent expected answers. | closed |
| T-23-09 | Tampering | native destructor | high | mitigate | Separate pointer-identity observer and four reached physical-destruction mutations. | closed |
| T-23-10 | Denial of service | duplicate/wrong free controls | high | mitigate | Observer rejects invalid use/release before libc access; subprocess deadlines bound controls. | closed |
| T-23-11 | Repudiation | compiler event sidecar | medium | mitigate | Mutation controls require a separate native receipt while retaining plausible compiler events. | closed |
| T-23-12 | Spoofing | example/evidence output | high | mitigate | Authored expected outputs; focused script runs native observer and public contract groups. | closed |
| T-23-13 | Repudiation | host receipt | high | mitigate | Script records host, compiler, target, revision, tree, and status; rejects skipped or unmatched tests; Linux CI wiring is present. | closed |
| T-23-14 | Denial of service | CI cost | medium | mitigate | One focused script step in the existing host matrix; no duplicated full/race/sanitizer suites. | closed |
| T-23-15 | Denial of service | adapter path/open/read | high | mitigate | Nonblocking open, regular-file admission, bounded reads, and typed failures. | closed |
| T-23-16 | Tampering | failed acquisition record | high | mitigate | Failures initialize `{NULL, 0}`; partial allocations are freed before return; injected failure controls. | closed |
| T-23-17 | Tampering | source owner binding | high | mitigate | Bind successful owning acquisition and exercise reached discard/copy refusals. | closed |
| T-23-18 | Elevation of privilege | C-emission contract | high | mitigate | Validate exact per-operation symbol, mode, status, and allocator pairing before serialization. | closed |

*Status: open · closed · open — below high threshold (non-blocking)*

## Accepted Risks Log

No accepted risks.

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-09-28 | 18 | 18 | 0 | gsd-security-auditor (ASVS L1) |

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-09-28
