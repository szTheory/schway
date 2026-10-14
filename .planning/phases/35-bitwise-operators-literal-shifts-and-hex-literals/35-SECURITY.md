---
phase: "35"
slug: "bitwise-operators-literal-shifts-and-hex-literals"
status: verified
threats_open: 0
asvs_level: 1
block_on: high
register_authored_at_plan_time: true
created: "2026-10-14"
---

# Phase 35 — Security

The read-only gsd-security-auditor verified all 24 distinct plan-authored mitigations at ASVS L1. No tests were executed during this security audit. All dispositions remain mitigate; no accepted risks or transfers were substituted for implementation.

## Trust Boundaries

Untrusted source crosses bounded parsing, numeric ownership and source admission into core. Raw core independently crosses core/origin validation before interpretation or C emission. C crosses installed Clang under unsigned/literal shift laws. Manifests and receipts cross into reviewer-facing claims through independent witnesses and negative controls.

## Threat Register

Paths abbreviated to peer/file below are relative to internal/compiler unless another root is explicit. Source line references record this audit's inspection; named execution receipts are separately scoped.

| Threat ID | Category | Component | Severity | Disposition | Mitigation evidence | Status |
|---|---|---|---|---|---|---|
| T-35-01 | Tampering | baseline comparator | high | mitigate | session/session_phase35_prechange_test.go:94,103,367,427 | closed |
| T-35-02 | Repudiation | numeric policy/resource history | medium | mitigate | check/check_phase35_baseline_test.go:66; syntax/parser_phase35_baseline_test.go:58; testdata/phase35/numeric-delta.json | closed |
| T-35-03 | Denial of Service | scalarExprAtom recursion | high | mitigate | syntax/parser.go:936,1046; syntax/parser_phase35_test.go:132 | closed |
| T-35-04 | Tampering | token adjacency/formatter | high | mitigate | syntax/parser.go:791; syntax/parser_phase35_test.go:18,393 | closed |
| T-35-05 | Tampering | numeric diagnostic ownership | medium | mitigate | syntax/lexer.go:58; syntax/parser.go:1499; syntax/numeric_phase35_test.go; check/check_phase35_numeric_test.go | closed |
| T-35-06 | Tampering | source shift admission | high | mitigate | check/scalar.go:785,896; check/check_phase35_test.go:16 | closed |
| T-35-07 | Elevation of Privilege | unused Byte bridge | high | mitigate | check/scalar.go:47,138; check/check_phase35_test.go:312; cmd/schway/main_phase321_run_input_test.go | closed |
| T-35-08 | Tampering | checker transfer | high | mitigate | check/scalar.go:307; check/check_phase35_test.go:167; session/session_phase35_peer_mutation_test.go | closed |
| T-35-09 | Tampering | corevalidate raw immediate | high | mitigate | corevalidate/corevalidate.go:1297,4786; corevalidate/corevalidate_phase35_test.go:72 | closed |
| T-35-10 | Tampering | origin admission/recompute/publication | high | mitigate | originvalidate/originvalidate.go:293,476,1365,1589; originvalidate/originvalidate_phase35_test.go:73,272,413; session/session_phase35_peer_mutation_test.go | closed |
| T-35-11 | Elevation of Privilege | raw unused-Byte bridge | high | mitigate | corevalidate/corevalidate.go:1213,4736,4752; originvalidate/originvalidate.go:1474,1508; TestPhase35UnusedByteIRAdmission and TestPhase35LegacyByteCompatibility in both peers | closed |
| T-35-12 | Tampering | interpreter numeric results | high | mitigate | interp/interp.go:1791; interp/interp_phase35_test.go:16,98 | closed |
| T-35-13 | Tampering | causal event inventory | high | mitigate | executionpeer/executionpeer.go:549; executionpeer/executionpeer_phase35_test.go | closed |
| T-35-14 | Elevation of Privilege | ancestry and Byte fixture | medium | mitigate | pathoracle/pathoracle.go:287; pathoracle/pathoracle_phase35_test.go; interp/interp_phase35_test.go:137 | closed |
| T-35-15 | Tampering | C shift emission | high | mitigate | cgen/cgen_program.go:2240,2766; cgen/cgen_phase35_test.go:245,472 | closed |
| T-35-16 | Denial of Service | event document capacity/projection | high | mitigate | cgen/cgen_program.go:251,901; cgen/cgen_phase35_test.go:352,472 | closed |
| T-35-17 | Tampering | emitter entry/classifier/signature | high | mitigate | cgen/cgen.go:112,122,133; cgen/cgen_phase35_test.go:162,352,579,620 | closed |
| T-35-18 | Tampering | optimized numeric outcomes | high | mitigate | session/session_phase35_values_test.go:57,121,175; 35-07-SUMMARY.md historical local receipts | closed |
| T-35-19 | Spoofing | source-free manifest claims | high | mitigate | surfacemanifest/manifest.go:146,209; session/session_phase35_surface_test.go:55,88,120,207 | closed |
| T-35-20 | Tampering | public entry/input projection | high | mitigate | cmd/schway/main_phase35_test.go; session/session_phase35_values_test.go:22; cmd/schway/main_phase321_run_input_test.go | closed |
| T-35-21 | Repudiation | lane/index/dispatch coverage | medium | mitigate | core/core_test.go:913,1069; session/session_phase35_values_test.go:57,121; 35-VALIDATION.md; 35-EVIDENCE.md | closed |
| T-35-22 | Tampering | Adler source rewrite | high | mitigate | examples/checksum/adler32.schway:17,20; session/session_phase35_adler_test.go:21,110; session/session_phase35_closeout_test.go:324 | closed |
| T-35-23 | Repudiation | summaries/docs/evidence | medium | mitigate | session/session_phase35_closeout_test.go:243,324; session/summary_coverage_test.go:24,123 | closed |
| T-35-SC | Tampering | dependency installation | high | mitigate | go.mod remains standard-library-only; no go.mod/go.sum delta from c4a2eec1; no installs performed | closed |

## Accepted Risks Log

No accepted risks. No unregistered SUMMARY threat flags.

## Evidence Scope

The existing mutation receipt records 3 no-op overlays and 15 actual arithmetic implementation mutants. Existing optimization receipts cover 33 vectors at O0/O3/O3-LTO and 16 UBSan shift rows on the local host; this audit did not rerun them or create hosted evidence. Phase 38 owns final two-host binding.

Source/raw-core admission controls submit malformed inputs and assert exact refusals; they are not executed implementation-source refusal-removal mutants. Actual implementation-source mutation evidence here applies to arithmetic transfers. L1 closure verifies the admission gates and their named controls by inspection.

Raw core validates immediate semantics, not original literal AST provenance; that authority remains in the source checker. Coordinated erasure of all identifying scalar families/reserved type facts remains the documented source-blind boundary of the unused-Byte guard. Executionpeer verifies causal event admission and fabricated-event refusal; it does not prove numeric answers. Readability and whole-loop cold/warm latency remain unmeasured.

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|---|---|---|---|---|
| 2026-10-14 | 24 | 24 | 0 | gsd-security-auditor; orchestrator recorded verdict |

## Sign-Off

- [x] All threats retain their planned mitigation disposition.
- [x] No accepted-risk substitution.
- [x] threats_open: 0 at ASVS L1/high threshold.
- [x] status: verified.

Approval: verified 2026-10-14, scoped to the evidence above.
