# Phase 26 planning amendments

## 2026-10-03 — Separate capture from independent verification

Independent goal verification identified an inaccurate sentence in
`26-05-PLAN.md`, Task 2 behavior: complete evidence was described as setting
`Verified=true`. That sentence is superseded by this dated correction:

> With complete event evidence, repeated scalar-copy occurrences are 0/1/2
> and the captured document passes independent `executionpeer.Validate`.
> `ApplicationEvidenceReport.Verified` remains false because capture is not a
> semantic verification verdict. Forced tiny capacity reports
> `capacity_exhausted` and `Verified=false`, retaining the actual child result.

This reconciles the plan with the existing Phase 22 capture contract and
D-26-08/D-26-09's separation of application outcome, evidence completeness,
and independent validation. It does not waive a requirement or alter the
locked decisions. The original plan remains preserved for auditability.

`TestPhase26SourceRepeatedCopyEvidence` validates interpreter-projected and
native-captured events independently. `TestPhase26EvidenceCapacity` pins the
incomplete-capture boundary. Both are exercised by the passing hosted receipts
in `26-VERIFICATION.md`; that independent report passes all five roadmap truths
with no unverified behavior. No implementation change is required.
