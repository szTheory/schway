# Phase 28 Exit-Pin Inventory (plan 28-01)

Recorded at PLAN_BASE c2a6f719ec4b823b330781949b7d37682ef0f9c2 (D-12, PITFALLS 13), before any pin is edited. Line numbers are as observed at PLAN_BASE and are approximate for later plans, which must re-check before editing. Plan 28-01 touches no production source.

## Production literals

Literals that plan 28-02 moves behind the exitcode table.

| file:line | literal or assertion | disposition | owning plan |
| --- | --- | --- | --- |
| internal/compiler/protocol/protocol.go:681-697 | ExitCode returns 0 (pass), 2 (invalid), 3 (operational), 4 (mismatch), default 64 | derive from exitcode table | 28-02 |
| cmd/schway/main.go:25-28 | const block exitSuccess 0, invalid 2, operational 3, usage 64 | derive from exitcode table | 28-02 |
| cmd/schway/main.go:242 | app-run `return 65` (input_too_long / input_contains_nul refusal) | named constant, value preserved | 28-02 |
| cmd/schway/main.go:245 | app-run `return 125` (runner error) | named constant, value preserved | 28-02 |
| cmd/schway/main.go:251 | `128 + outcome.SignalNumber` (signaled child) | named helper, value preserved | 28-02 |
| cmd/schway/main.go:254 | app-run `return 124` (timed out) | named constant, value preserved | 28-02 |
| cmd/schway/main.go:257 | app-run `return 125` (launch error) | named constant, value preserved | 28-02 |
| cmd/schway/main.go:260 | app-run `return 125` (invalid process outcome) | named constant, value preserved | 28-02 |
| cmd/schway-repair/main.go:12-15 | const block exitSuccess 0, exitUnrepaired 1, exitOperational 3, exitUsage 64 (plus exitUnrepaired 1) | derive from exitcode table | 28-02 |
| internal/compiler/native/native.go:37 | EventCapacityExitStatus = 69 | preserve; reference from table | 28-02 |
| internal/compiler/native/native.go:630 | comparison `adjudication.ExitCode == EventCapacityExitStatus` | preserve | 28-02 |
| internal/compiler/native/native_app.go:424 | `ExitCode: 0` in RunExited outcome | named constant | 28-02 |
| internal/compiler/native/sanitize.go:227 | `exitCode == 0` (and the -1 sentinel, out of scope per D-21) | named constant for 0 only | 28-02 |
| internal/compiler/session/session_phase5.go:456 | `report.ExitCode == 0` | named constant | 28-02 |
| internal/compiler/session/session_phase5_sanitize.go:180, 205, 218 | `ExitCode == 0` comparisons on three sanitizer reports | named constant | 28-02 |

## Generated-C exit text

Literals inside cgen string literals. Plan 28-02 pins them with a scan test and does not edit them.

| file | text | count | meaning read from surrounding C |
| --- | --- | --- | --- |
| internal/compiler/cgen/cgen.go | exit(65) | 1 | application arithmetic or limit refusal |
| internal/compiler/cgen/cgen.go | exit(69) | 1 | event capacity exhausted (matches EventCapacityExitStatus) |
| internal/compiler/cgen/cgen_program.go | exit(0) | 2 | normal completion paths |
| internal/compiler/cgen/cgen_program.go | exit(65) | 10 | application refusal paths (overflow, remainder by zero, output limit, input acquire) |
| internal/compiler/cgen/cgen.go | return 0; | 46 | success |
| internal/compiler/cgen/cgen.go | return 1; | 14 | generic failure |
| internal/compiler/cgen/cgen.go | return 71; | 1 | unreachable after schway_defect |
| internal/compiler/cgen/cgen.go | return 74; | 24 | write failure |
| internal/compiler/cgen/cgen_program.go | return 0; | 21 | success |
| internal/compiler/cgen/cgen_program.go | return 1; | 4 | generic failure |
| internal/compiler/cgen/cgen_program.go | return 2; | 1 | UseError.UnsupportedByte path |
| internal/compiler/cgen/cgen_program.go | return 64; | 2 | wrong argument count or over-long argument |
| internal/compiler/cgen/cgen_program.go | return 65; | 9 | application refusal |
| internal/compiler/cgen/cgen_program.go | return 70; | 1 | entry name missing |
| internal/compiler/cgen/cgen_program.go | return 74; | 16 | write failure |

Flag: 70, 71, 74, 2 and 64 are generated-program child statuses outside the D-07 host exit table; the host surfaces them only as an opaque child exit code.

## Pins that will flip

| file or test | pin | owning plan |
| --- | --- | --- |
| internal/compiler/native/native_test.go:1521, 1564 | cutAfterCheckRefusal for foreign_origin_omitted (cut-M004 native exit 3 becomes check-class refusal on native) | 28-06 |
| internal/compiler/session/session_test.go | TestPublishedOriginValidatedOnCheckAndRun native assertion | 28-06 |
| internal/compiler/session/session_phase16_production_paths_test.go:22 | TestPhase16ProductionPathsPreserveM004Refusal restrict_borrow entry | 28-06 |
| internal/compiler/session/session_phase271_admission_test.go | corevalidate/build, corevalidate/interp, origin/interp error-text asserts | 28-06 |
| internal/compiler/session/session_phase271_admission_test.go | the */verify rows | 28-07 |
| internal/compiler/session/session_phase271_admission_test.go:234 | TestPhase271VerifyFileRouteWritesOnlyFailReport | 28-07 |
| internal/compiler/callgraph/callgraph_entry_test.go:140 | TestEntryFunctionEmitsUnreachableFunctions | 28-04 |
| internal/compiler/cgen/cgen_program_test.go:1254 | TestEmitProgramZeroCallEdges | 28-05 |
| internal/compiler/session/session_phase11_differential_test.go:377 | ZeroCallEdges subtest of TestPhase11InterproceduralDifferential | 28-05 |
| internal/compiler/session/session_payload_replay_test.go:165 | payloadCorpusExpectedSkips zero_call entry | 28-05 |
| testdata/phase272/accepted-corpus-hashes.json and testdata/phase273/conformance-corpus-hashes.json | zero_call entries, via named allow-list entries only (goldens stay frozen, D-23) | 28-05 |
| internal/compiler/session/session_admission_divergence_test.go:66 | knownAdmissionDivergences payload_borrow_interaction | 28-05 |
| internal/compiler/session/session_phase273_appverify_test.go and cmd/schway/main_test.go:562, 582 | `len(Engines)` asserts (value 3) and ApplicationVerifySchema consumers (main_test.go:551) | 28-03 |

## Pins to preserve

Re-run before and after the owning plan.

| location | pin |
| --- | --- |
| cmd/schway/main_test.go (lines 105-376) | every exit 65 pin and the exit 64 usage pin at line 115 |
| scripts/verify-phase27.sh, scripts/check-phase27-receipts.sh:99, examples/phase27/README.md:41, 53 | exit 65 receipts and documentation |
| internal/compiler/native/native.go:37, 630 | exit 69 comparison |
| internal/compiler/cgen/cgen_phase273_test.go:86 | TestPhase273CapacityContractLiteralsAgree |
| internal/compiler/native/native_test.go:1710 | TestPhase273RunnerEventCapacityAdjudication |
| internal/compiler/session/session_phase273_appverify_test.go | exit 69 and 65 verify pins |
| internal/compiler/testsupport cli_test.go and cli_phase6_test.go; session_test.go | compiler-domain 2, 3, 4 pins |
| internal/compiler/native/native_test.go:1483 | cut-M004 native exit 3 fixtures in phase4CorpusMatrix (other than the foreign_origin_omitted entry that flips) |
| native run with an empty PATH | exit 3 operational |

Note: the 65 hits in scripts/verify-phase25.sh (lines 98-181) are U64 values (the byte 0x41 and receipts), not exit codes.

## New pins

| pin | owning plan |
| --- | --- |
| app-run exit 124 (timeout), 125 (launch or runner failure), 128+n (signaled) | 28-02 |
| cross-surface exit-2 fixtures: duplicate_function_name, foreign_origin_omitted, payload_borrow_interaction | 28-06 |
| documentation table of exit classes | 28-07 |

## Corpus audit

Source: testdata/phase28/entry-exit-class-manifest.json (155 files). Entry shape counts: single-root 87, unchecked 62, multi-root-main 5, multi-root-no-main 1, no-function 0, no-root 0.

| file | entryShape | entry | entryError | deadFunctions |
| --- | --- | --- | --- | --- |
| testdata/phase11/multi_function_gate_corpus.schway | multi-root-main | main | none | s1:phase11.multi_function_gate_corpus:fn:touch |
| testdata/phase11/multi_function_unreachable.schway | multi-root-main | main | none | s1:phase11.multi_function_unreachable:fn:orphan |
| testdata/phase11/multi_function_zero_call.schway | multi-root-main | none | core.entry_ambiguous | none |
| testdata/phase12/payload_borrow_interaction.schway | multi-root-no-main | none | core.entry_ambiguous | none |
| testdata/phase13/heldout_shared_callee_twin_alpha.schway | multi-root-main | main | none | s1:phase13.heldout_shared_callee_twin_alpha:fn:beta |
| testdata/phase13/heldout_shared_callee_twin_mirror.schway | multi-root-main | main | none | s1:phase13.heldout_shared_callee_twin_mirror:fn:alpha |

Tiebreak-without-main entries that resolve: 0 (D-03: no resolving program depends on the closure-size tiebreak without a function named main). The four dead-function programs match research exactly.

Files whose check class will change when check starts resolving the entry: testdata/phase12/payload_borrow_interaction.schway (check exit 0 today, then an entry_ambiguous refusal) and testdata/phase11/multi_function_zero_call.schway (check exit 0 today, tie with main, same refusal). No checked file has shape no-function or no-root, so none of those apply.

Observed baseline classes cross-checked against research: payload_borrow_interaction check exit 0, build and interpreter run exit 3, native run exit 3 (native.tool_failure under the missing-clang runner); duplicate_function_name check exit 2 core.duplicate_function_id with build and interpreter run exit 3; foreign_origin_omitted exit 2 on check, build and interpreter run, and exit 3 on native run. All match research.

## Baseline runs

Observed at PLAN_BASE in package internal/compiler/session, all passing in one run (package line ok):

| test | status |
| --- | --- |
| TestPhase273ConformanceCorpusByteIdentity | PASS |
| TestPhase272AcceptedCorpusByteIdentity | PASS |
| TestPhase28EntryExitClassManifest | PASS |
| TestPhase16PublicEmitterConsumerInventory | PASS |
