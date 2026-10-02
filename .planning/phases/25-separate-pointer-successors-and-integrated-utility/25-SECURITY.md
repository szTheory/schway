---
phase: "25"
slug: "separate-pointer-successors-and-integrated-utility"
status: verified
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: "2026-10-02"
---

# Phase 25 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| Source to checked facts | A malformed or mutated source shape could be misclassified as a shared or exclusive loan. | Source forms and checked borrow facts |
| Checked core to peer validators and C serializer | Forged family, origin, or loan-lifetime facts could admit invalid C pointer use or unsupported optimizer assumptions. | Core operations, loan identities, ABI declarations, and manifests |
| Model behavior to native claim | Model replay could be misrepresented as native pointer or host-I/O evidence. | Program inputs, generated C, compiler and runtime results |
| Refusal to diagnostic | Vague or unsafe advice could conceal a conflict or induce a safety bypass. | Source spans, refusal classes, and cause spans |
| CI host to evidence receipt | A skipped family, lane, or host could be mistaken for completion. | Host, family, lane, compiler, flags, and expected results |
| Source claims to roadmap/maturity | Historical or model-only evidence could be presented as new native proof. | Inspection, local checks, hosted results, and historical receipts |

---

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-25-01 | Tampering / Elevation | Shared source-family admission | high | mitigate | Checker preserves borrow-family facts and fixed caller composition; incompatible-source controls cover refusal. | closed |
| T-25-02 | Tampering | Serializer and ABI derivation | high | mitigate | Exact pointer declarations, call operands, and manifests derive from checked facts; unsupported candidates refuse before serialization. | closed |
| T-25-03 | Information Disclosure / Tampering | Manifest promises | medium | mitigate | Published ABI facts derive from checked source without unsupported alias, capture, alignment, or ownership claims. | closed |
| T-25-04 | Tampering / Elevation | Core/origin peer derivation | high | mitigate | Independent core and origin peers reject wrong-family, conflict, and escape mutations. | closed |
| T-25-05 | Tampering | Path-sensitive loan lifetime | high | mitigate | Path oracle derives straight-line loan uses, overlap, and escape from operations; endpoint, call-boundary, overlap, and escape controls pass. | closed |
| T-25-06 | Repudiation / Tampering | Independence of guarantees | medium | mitigate | Peer validators derive obligations independently and do not use checker flags or serialized endpoint tables as proof. | closed |
| T-25-07 | Tampering / Elevation | Exclusive source admission and caller composition | high | mitigate | Positive source-family checks and wrong-family, mutation, ordering, and legacy-caller controls cover bounded admission. | closed |
| T-25-08 | Tampering | Borrow conflict and escape checking | high | mitigate | Core, origin, and path peers independently reject mutated overlap and escape programs. | closed |
| T-25-09 | Repudiation / Information Disclosure | Structured diagnostics | medium | mitigate | Stable diagnostics include primary and applicable cause spans, function identity, and no unsafe suggested repair. | closed |
| T-25-10 | Tampering / Elevation | Exclusive lowering and attributes | high | mitigate | Exclusive C and manifest facts match the checked family; wrong-result and structural refusal controls are present. | closed |
| T-25-11 | Tampering | Error order and model/native boundary | high | mitigate | Interpreter and native tests verify expected answers, inherited error order, and reached wrong-result controls. | closed |
| T-25-12 | Repudiation / Information Disclosure | Structured diagnostics | medium | mitigate | Stable source-attributed escape diagnostics preserve primary and cause spans and distinguish refusal classes. | closed |
| T-25-13 | Repudiation / Tampering | Clean-checkout utility contract | medium | mitigate | Documented inputs and evidence checks cover 65/66 results and the typed 0x43 error path. | closed |
| T-25-14 | Tampering / Repudiation | Host/family/lane receipts | high | mitigate | Verification rejects skipped tests and marks absent family/host/lane rows incomplete; existing CI runs the matrix script. | closed |
| T-25-15 | Repudiation / Information Disclosure | Living capability claims | medium | mitigate | Roadmap and maturity records distinguish inspection, executed checks, hosted results, and historical receipts. | closed |

The T-25-14 fail-incomplete control is implemented. Hosted run `36971855722`
at branch head `421b5b94eb867e940a207dfd64d7971dc8198172` and PR merge
revision `a90c27c5b432ef6fc59fbafaa68b50a1374ae138` passed the required native
Linux/x86_64 and macOS/arm64 evidence aggregates. All three families (foreign,
shared, exclusive) passed baseline `-O0`, optimized `-O2`, and ASan+UBSan on
both hosts: 18/18 distinct rows, each returning expected/actual 65 and 66 and
preserving the typed `0x43` failure before helper calls. Linux used Go
`linux/amd64`, target `x86_64-pc-linux-gnu`, and Ubuntu Clang 18.1.3; macOS used
Go `darwin/arm64`, target `arm64-apple-darwin25.6.0`, and Apple Clang 21. The
`checks` and `current evidence aggregate` jobs passed on both hosts; checks
included vet, build, full Go tests, and race tests. This hosted receipt closes
EVD-10. Individual evidence-script logs mark the other host incomplete because
each invocation runs on one native host; the paired successful aggregate job
supplies that host's independently executed rows.

The audit trail and verdict remain unchanged: SECURED, ASVS L1, all 15
mitigations closed, `threats_open: 0`. The 2026-10-02 security audit records
the peer and diagnostic assessment; the later hosted run adds evidence for
T-25-14 without changing that assessment's scope or claiming that source
inspection alone proves native behavior.

Later plan-local threats are traceability aliases for the existing controls:
T-25-16 → canonical T-25-13, T-25-17 → T-25-14, T-25-18 → T-25-15,
T-25-19 → T-25-14, and T-25-20 → T-25-15. These aliases connect later plan
work to existing audit evidence; they do not add audited controls or findings.

---

## Accepted Risks Log

No accepted risks.

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-10-02 | 15 | 15 | 0 | GSD security auditor, ASVS L1 |

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** SECURED / ASVS L1; EVD-10 host evidence bound by run 36971855722.
