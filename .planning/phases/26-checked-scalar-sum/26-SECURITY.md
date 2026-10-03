---
phase: "26"
slug: "checked-scalar-sum"
status: verified
threats_open: 0
asvs_level: 1
created: "2026-10-03"
---

# Phase 26 — Security

This audit checks the twelve plan-authored threats at configured ASVS level 1, with high severity blocking advancement. It records source inspection plus hosted execution evidence; no tests or native programs were run locally. This is verification of the declared mitigations, not a claim that the compiler has no other threats.

## Trust Boundaries

| Boundary | Description | Data Crossing |
|---|---|---|
| Source → parser/checker | Untrusted source must remain bounded and obtain typed scalar admission. | Source bytes, tokens, CFG state |
| Checked core → independent validators | A producer's facts cannot substitute for independent scalar and authority checks. | Operations, types, provenance, back edges |
| Checked arithmetic → interpreter/C17 | Both engines must check overflow before addition. | U64 operands and terminal outcome |
| Child process → app caller/capture | Child results, operational errors, and evidence completeness remain distinct. | stdout, stderr, exit status, evidence bytes |
| Event producers → execution peer | Repeated events require independently checked occurrence and membership. | Invocation/site IDs, ordinals, terminal attribution |
| Expected answers → engine comparison | An implementation cannot author its own oracle. | Literal results and reached mutation outcomes |

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation and evidence | Status |
|---|---|---|---|---|---|---|
| T-26-01 | Denial of service | syntax/check scalar worklist | high | mitigate | `syntax/parser.go` bounds source bytes/tokens; `check/scalar.go` caps transfers at 65,536. `TestPhase26FrontendRecovery` and `TestPhase26AnalysisExhaustion` cover recovery and refusal. | closed |
| T-26-02 | Tampering | cyclic core admission | high | mitigate | Separate scalar derivations in `check/scalar.go`, `corevalidate/corevalidate.go`, and `originvalidate/originvalidate.go`; tracer, peer-independence, and cycle mutation tests. | closed |
| T-26-03 | Information disclosure | app defect output | medium | mitigate | `sum_to_n.schway` rejects 1001 before returning output; `TestPhase26ExactSumMatrix` and `TestPhase26PublicAppCLI` pin empty stdout and bounded defect stderr/status. | closed |
| T-26-04 | Tampering | interpreter/C checked addition | high | mitigate | Explicit check-before-add in both engines; `TestPhase26CheckedAddInterpreter`, `TestPhase26CheckedAddC17`, and `TestPhase26OverflowProcessOutcome` pin MAX+1 failure. CR-01 repair also checks terminal attribution and reserves its serialized size. | closed |
| T-26-05 | Repudiation | native outcome/evidence capture | high | mitigate | `TestPhase26FailureChannels` and `TestPhase26OverflowProcessOutcome` distinguish child status/streams, ToolError, and capture completeness. | closed |
| T-26-06 | Denial of service | checker/peer worklists | high | mitigate | All three analyses independently enforce the deterministic 65,536 bound with exhaustion seams and fail-closed diagnostics. Each package has `TestPhase26AnalysisExhaustion`. | closed |
| T-26-07 | Elevation of privilege | back-edge authority | high | mitigate | Checker and peers rederive owner/resource/loan/provenance restrictions. `TestPhase26BackEdgeAuthority`, category tests, and core/origin mutation controls retain refusals. | closed |
| T-26-08 | Tampering | cyclic operation classification | high | mitigate | Scalar U64/Bool copies only; unsupported event/authority carries remain refused. `TestPhase26ScalarCopyCycleBoundary`, peer mutations, and the acyclic path-oracle control cover the boundary. Checked-add terminal admission is separately constrained. | closed |
| T-26-09 | Tampering | expected-answer comparison | high | mitigate | `TestPhase26ExactSumMatrix` pins 0/55/500500 independently for each engine and rejects reached wrong-result/skipped-iteration mutations. | closed |
| T-26-10 | Repudiation | public app/evidence result | high | mitigate | Public CLI tests pin exact streams/status. `TestPhase26EvidenceCapacity` preserves child result while reporting incomplete, unverified capture; `TestPhase26SourceRepeatedCopyEvidence` checks actual source events with the independent peer. | closed |
| T-26-11 | Tampering | interpreter/C event producer | high | mitigate | Per-invocation/site ordinals with canonical zero omission; execution encoding, interpreter/native repeated-copy, and legacy byte-golden controls. | closed |
| T-26-12 | Repudiation | execution peer | high | mitigate | `TestPhase26PeerOccurrenceOrder` rejects duplicate, gapped, reordered, wrong-invocation/site and forged non-scalar events. Overflow controls reject wrong reason/source/type/outcome/position. | closed |

Paths above are relative to `internal/compiler/` unless an example is named. Test mappings and exact hosted commands remain in `26-VALIDATION.md`.

## Accepted Risks Log

No accepted risks. Resource/loan loop carries remain refused. The analysis cap is not a runtime termination proof. Event validation establishes structural attribution and occurrence integrity, not operand-value proof of overflow.

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|---|---|---|---|---|
| 2026-10-03 | 12 | 12 | 0 | GSD orchestrator, plan-authored ASVS L1 source audit |

Source inspection found all declared mitigations. The secure-phase workflow's authored-register/L1 short-circuit applies, so no deeper auditor dispatch was needed. Independent code review is clean after CR-01 repair. [Full hosted run 37118315516](https://github.com/szTheory/schway/actions/runs/37118315516) passed Ubuntu and macOS full/race suites and evidence aggregates at 46bee44ad87891d8a8547b2489f029d0fe237899. Earlier receipts are historical and are not substituted for this revision's checks.

## Sign-Off

- [x] All threats have disposition mitigate
- [x] Accepted risks documented
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-10-03 at the configured L1 depth.
