# Phase 35 closeout transition check — 2026-10-14

Phase 35 implementation now includes the public 264/8 bitwise witness, hexadecimal spelling, bounded literal shifts and the one-line Adler rewrite. Final phase acceptance is owned by `35-VERIFICATION.md`; this transition review does not substitute for that verdict. The earlier dated implementation-only pending notes remain historical.

## Evidence and repaired obligations

New local evidence: `35-REGRESSION.md` final refresh records 94 distinct names/98 package-qualified executions, no missing/fail/skip. `35-ADMISSION-MUTATION-EVIDENCE.md` records 26 actual refusal-removal controls; `35-MUTATION-EVIDENCE.md` records the earlier 15 arithmetic controls. Final verification must refresh the arithmetic controls against the repaired validator snapshots. Independent shift-target freshness gates, formatter token boundaries and grouped-negative refusal fixes are committed; review is clean and the L1 security register has 24 closed threats. Historical optimization/UBSan rows remain explicitly bound to their original local receipt until final verification reruns them. Phase 38 hosted receipts, human readability and whole-loop cold/warm distributions remain pending or unmeasured.

## Three next capabilities

| User-visible program | Current blocker | Smallest complete slice | Checker/guarantee changes | Evidence/debt | Owner/next action | Observation changing priority |
|---|---|---|---|---|---|---|
| Imported `crc_step` called inside a checksum loop | Scalar-expression calls are absent | One arity-one acyclic imported U64/Bool helper and one loop call | Independent argument/result, import identity, arity and call-defect checks | CALL-01/02; independent helper results and defect controls | Phase 36 discussion/planning; no implementation started here | Named consumer requires another call shape or arity |
| CRC-32 file tool, second importer and public interpreter CLI parity | Helper calls and file-input interpreter CLI parity are absent | Bounded ISO-HDLC CRC-32 library, file program, second importer and guide | Preserve file bounds/cleanup; independent CRC answers, failure and reuse controls | APP-13/14, MOD-07, DX-24 | Phase 37 after helper slice | Actual consumer needs a different bounded file/failure contract |
| Reproducible checksum/CRC verification on macOS and Linux | Current final hosted receipts are unbound | Run and bind both hosts at one final source/fixture revision | Defined target-independent operator behavior and receipt provenance; no backend change | EVD-17 and Phase 38 receipt owner; M006 historical | Phase 38 evidence closure after consumer integration | One host exposes a semantic, sanitizer or receipt-binding difference |

## Conditional seeds

Re-read trigger metadata and the relevant z1g/trz/z1s seeds plus the Dashbit appraisal. Retain all dormant. z1g: `session_phase6_query.go#QueryCommandFile` remains sufficient for the named crc_step review; no measured query gap. trz: `native_app.go#readApplicationCapture`, `TestPhase33CaptureReadBounded` and application verification remain the witnesses for checksum failure diagnosis; current defects were diagnosed from source and named controls, so no unexplained capture gap. z1s: `syntax/format.go#Format` remains the canonical-source witness; the token-spacing defect was a settled-contract violation, not a contested source-design choice. Revisit only on those named triggers; no new database, telemetry runtime or proof backend follows from the appraisal.

The active requirements and phase owners remain in ROADMAP.md/REQUIREMENTS.md. Phase 35 has a runnable gain; Phase 36 must retain its named helper/loop witness rather than become another unbounded enabling phase.
