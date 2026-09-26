---
phase: "19"
slug: "numeric-literals-and-opconst"
status: verified
threats_open: 0
asvs_level: 1
register_authored_at_plan_time: true
created: "2026-09-26"
---

# Phase 19 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| Source text to compiler | Numeric source bytes are untrusted until fully tokenized and checked. | Decimal, hexadecimal, binary, separators, malformed tails, and magnitudes |
| Checked core to peers and engines | Core facts may be forged and must be independently validated before execution. | `OpConst`, type identity, canonical U64 payload |
| Interpreter/native output to evidence | U64 values cross runtime and target-width boundaries into observable execution documents. | Exact-width values and serialized decimal results |
| Operation registry to six consumers | A registered operation must be handled by every analysis and execution consumer. | Operation kind, execution result, differential evidence |

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-19-01 | Tampering | `interp.value.String` | high | mitigate | U64 projection is canonical decimal; legacy scalar replay and seeded projection mutation control pass (`interp_phase19_test.go`, `session_payload_replay_test.go`). | closed |
| T-19-02 | Tampering | numeric source frontier | medium | mitigate | Overflow and malformed witnesses retain full-span refusals (`session_phase19_test.go::TestPhase19NumericRefusalFrontiers`). | closed |
| T-19-03 | Tampering | numeric lexer | high | mitigate | Whole-token grammar and full-span malformed refusals are covered by `syntax_test.go::TestPhase19NumericToken` and `TestPhase19NumericMalformed`. | closed |
| T-19-04 | Denial of service | lexer scan | medium | mitigate | Numeric candidate scanning is bounded by existing source/token limits and advances once; malformed numeric tests pass (`syntax/lexer.go`, `syntax_test.go`). | closed |
| T-19-05 | Tampering | checker literal conversion | high | mitigate | Checked 64-bit parsing rejects max+1 before `OpConst`; admission/range tests pass (`check_phase19_test.go::TestPhase19LiteralAdmission`, `TestPhase19LiteralRange`). | closed |
| T-19-06 | Elevation of privilege | checker type boundary | medium | mitigate | Literals have fixed U64 identity and mismatched return types are refused (`check_phase19_test.go::TestPhase19LiteralType`). | closed |
| T-19-07 | Tampering | `corevalidate` OpConst | high | mitigate | Independent validation accepts canonical facts and rejects forged facts (`corevalidate_phase19_test.go::TestPhase19OpConstIndependentAdmission`, `TestPhase19ForgedOpConstFactsRefused`). | closed |
| T-19-08 | Tampering | path and origin peers | medium | mitigate | Constant targets are source-free roots; path and origin tests verify no inherited source or parameter origin (`pathoracle_test.go::TestPhase19ConstantPathRoots`, `originvalidate_test.go::TestPhase19ConstantOriginStopsAtRoot`). | closed |
| T-19-09 | Tampering | native constant lowering | high | mitigate | Generated C uses `uint64_t`, checked `UINT64_C` constants, and exact-width compile guards; native max-value and missing-macro controls pass (`cgen_program.go:673-674`, `cgen_program_test.go::TestPhase19U64NativeExactWidthAndOutput`, `TestPhase19ExactWidthTargetGuardRejectsMissingMacro`). | closed |
| T-19-10 | Tampering | execution serialization | high | mitigate | Interpreter and native paths emit canonical decimal strings and retain scalar replay evidence (`interp.go:498`, `cgen_program.go:879`, `session_payload_replay_test.go::TestPayloadCorpusCharacterizationReplay`). | closed |
| T-19-11 | Denial of service | numeric output writer | medium | mitigate | U64 output reuses the bounded output writer (`cgen_program.go:879`); max-value serialization and the full Go suite pass. | closed |
| T-19-12 | Tampering | exhaustive dispatch controls | high | mitigate | The real literal fixture reaches all six consumers; kind/site mutation controls are exercised (`core_test.go::TestAllOperationKindsHandledAtEverySite`, `session_phase19_test.go::TestPhase19Dispatch`, `session_phase7_mutation_test.go`). | closed |
| T-19-13 | Repudiation | four-tier evidence | medium | mitigate | Four tiers assert the expected value and compare all pairs; the seeded wrong-result control is rejected (`session_phase19_test.go::TestPhase19FourTierLiteral`, `TestPhase19WrongResultControl`). | closed |
| T-19-14 | Tampering | scalar goldens | high | mitigate | Scalar corpus replay and frozen generated-C baselines pass; no unreviewed golden movement is recorded (`session_payload_replay_test.go`, `19-SCALAR-REVIEW.md`). | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on count toward `threats_open`*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

## Accepted Risks Log

No accepted risks.

## Threat Flags

No unregistered threat flags were recorded in the seven Phase 19 summaries.

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-09-26 | 14 | 14 | 0 | Codex |

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-09-26
