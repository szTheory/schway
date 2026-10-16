---
phase: 35-bitwise-operators-literal-shifts-and-hex-literals
verified: 2026-10-16T21:13:17Z
status: passed
score: "38/38 truths verified"
covered_files:
  - ".planning/phases/33-carried-debt-bounded-capture-read-and-generic-freeze-gate/33-REVIEW-DISPOSITION.md"
  - ".planning/phases/33-carried-debt-bounded-capture-read-and-generic-freeze-gate/33-REVIEW.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-01-PLAN.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-01-SUMMARY.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-02-PLAN.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-02-SUMMARY.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-03-PLAN.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-03-SUMMARY.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-04-PLAN.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-04-SUMMARY.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-05-PLAN.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-05-SUMMARY.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-06-PLAN.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-06-SUMMARY.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-07-PLAN.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-07-SUMMARY.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-08-PLAN.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-08-SUMMARY.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-371-REFRESH-2026-10-16-VERIFICATION-HISTORY.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-371-REFRESH-GATES.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-371-REFRESH-RUN.json"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-COMPATIBILITY-DELTA.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-EVIDENCE.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-PROHIBITION-ENFORCEMENT.json"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-PROHIBITION-WIRING.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REFRESH3-GATES.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REFRESH3-PROHIBITION.json"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REFRESH3-RUN.json"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REFRESH3-VALIDATION.json"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REVIEW-DISPOSITION.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-SECURITY.md"
  - ".planning/phases/36-scalar-helper-calls/36-REFRESH2-GATES.md"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION-03.jsonl"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION-04.jsonl"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION-05.jsonl"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION-06.jsonl"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION-07.jsonl"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION-09.jsonl"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION-10.jsonl"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION-12.jsonl"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION-13.jsonl"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION-15.jsonl"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/371-REFRESH-COMPLETION-2026-10-16.json"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/371-REFRESH-DISPOSITION-AMENDMENT-2026-10-16.md"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/371-REFRESH-SYNTAX-20261016.jsonl"
  - "cmd/schway/main_phase35_test.go"
  - "docs/agent-surface.md"
  - "examples/bitwise.schway"
  - "examples/checksum/adler32.schway"
  - "internal/compiler/cgen/cgen.go"
  - "internal/compiler/cgen/cgen_exit_literals_test.go"
  - "internal/compiler/cgen/cgen_phase33_scalar_defect_test.go"
  - "internal/compiler/cgen/cgen_phase35_test.go"
  - "internal/compiler/cgen/cgen_phase36_caller_shapes_test.go"
  - "internal/compiler/cgen/cgen_phase371_legacy_defect_test.go"
  - "internal/compiler/cgen/cgen_program.go"
  - "internal/compiler/check/check.go"
  - "internal/compiler/check/check_phase35_baseline_test.go"
  - "internal/compiler/check/check_phase35_numeric_test.go"
  - "internal/compiler/check/check_phase35_test.go"
  - "internal/compiler/check/scalar.go"
  - "internal/compiler/core/core.go"
  - "internal/compiler/core/core_phase35_test.go"
  - "internal/compiler/core/core_test.go"
  - "internal/compiler/corevalidate/corevalidate.go"
  - "internal/compiler/corevalidate/corevalidate_phase35_test.go"
  - "internal/compiler/executionpeer/executionpeer.go"
  - "internal/compiler/executionpeer/executionpeer_phase33_terminal_test.go"
  - "internal/compiler/executionpeer/executionpeer_phase35_test.go"
  - "internal/compiler/executionpeer/executionpeer_phase371_test.go"
  - "internal/compiler/interp/interp.go"
  - "internal/compiler/interp/interp_phase35_test.go"
  - "internal/compiler/native/native.go"
  - "internal/compiler/originvalidate/originvalidate.go"
  - "internal/compiler/originvalidate/originvalidate_phase35_test.go"
  - "internal/compiler/pathoracle/pathoracle.go"
  - "internal/compiler/pathoracle/pathoracle_phase35_test.go"
  - "internal/compiler/session/session_phase34_compat_test.go"
  - "internal/compiler/session/session_phase34_limit_consumers_test.go"
  - "internal/compiler/session/session_phase34_surface_test.go"
  - "internal/compiler/session/session_phase35_adler_test.go"
  - "internal/compiler/session/session_phase35_admission_mutation_test.go"
  - "internal/compiler/session/session_phase35_closeout_test.go"
  - "internal/compiler/session/session_phase35_peer_mutation_test.go"
  - "internal/compiler/session/session_phase35_prechange_test.go"
  - "internal/compiler/session/session_phase35_surface_test.go"
  - "internal/compiler/session/session_phase35_values_test.go"
  - "internal/compiler/session/session_phase371_compatibility_test.go"
  - "internal/compiler/session/summary_coverage_test.go"
  - "internal/compiler/surface/refusals.go"
  - "internal/compiler/surfacemanifest/manifest.go"
  - "internal/compiler/surfacemanifest/manifest_test.go"
  - "internal/compiler/syntax/format.go"
  - "internal/compiler/syntax/lexer.go"
  - "internal/compiler/syntax/numeric_phase35_test.go"
  - "internal/compiler/syntax/parser.go"
  - "internal/compiler/syntax/parser_phase35_baseline_test.go"
  - "internal/compiler/syntax/parser_phase35_test.go"
  - "testdata/phase16/public-emitter-consumers.json"
  - "testdata/phase35/numeric-delta.json"
  - "testdata/phase35/numeric-policy-baseline.json"
  - "testdata/phase35/prechange-corpus.json"
  - "testdata/phase35/prohibition-causal-bad.txt"
  - "testdata/phase35/prohibition-evidence.test.cjs"
  - "testdata/phase35/prohibition-lanes-bad.txt"
  - "testdata/phase35/recovery-after.json"
  - "testdata/phase35/shift-sites.json"
  - "testdata/phase35/surface-census.json"
  - "testdata/phase35/surface-witnesses.json"
  - "testdata/phase35/value-vectors.json"
  - "testdata/phase371/compatibility-amendments.json"
  - "testdata/phase371/pre-emitter-census.json"
covered_digest: "v3:sha256:533112f30cd6b2885619e81d5efdcd405bddfdd9fe7cc3d8b3a69b30d2ed3c64"
behavior_unverified: 0
overrides_applied: 0
decision_coverage:
  honored: 23
  total: 23
  not_honored: []
human_verification: []
---

# Phase 35: Bitwise Operators, Literal Shifts, and Hex Literals Verification Report

**Phase Goal:** Ordinary Schway scalar code can use U64 `^`, `&`, `|`, literal-amount `<<` and `>>`, and hex literals, with one defined meaning on every engine and optimization level, no C17 shift undefined behavior, and no new defect path.
**Verified:** 2026-10-16T21:13:17Z
**Status:** passed
**Re-verification:** No — the prior passed report had no gaps; must-haves were re-established from the roadmap and eight plans.
**Prior report:** Passed 38/38 with no `gaps:` section, but became stale after the Phase 37.1 source-delta refresh.

## Goal Achievement

### Observable Truths

The contract is the five roadmap success criteria plus the 34 unique plan truths. One plan truth restates the Adler roadmap criterion, giving 38 distinct truths.

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | U64 xor, and, and or have independently pinned values in interpreter and native O0, O3, and O3-LTO; check/core/origin derive transfers independently and executionpeer checks event-free causality. | ✓ VERIFIED | Current Phase 37.1 raw logs and directly rerun value matrix support the three optimization lanes; current source-bound mutation receipts record 15/15 independent transfer mutants killed. Executionpeer’s causal-only boundary is preserved. |
| 2 | Literal shifts at 0, 31, 32, and 63 agree across engines/optimization levels; low-64 left shift, nonliteral and out-of-range refusal, and safe C sites are defined. | ✓ VERIFIED | Checker/core/origin source enforces literal 0..63; native shift scanner, 33-vector matrix, 16-case UBSan run and admission mutants pass. Emission widens unsigned operands and casts to `uint64_t`. |
| 3 | Hex digits are case-insensitive, spelling is preserved, and malformed/overflow tokens receive named diagnostics. | ✓ VERIFIED | Lexer/parser ownership, formatter round-trip and refusal registry inspected; syntax grammar, numeric ownership and formatter tests pass. |
| 4 | Adler uses `b << 16`; pinned checksum and `adler_low` answers remain identical on both engines. | ✓ VERIFIED | `TestPhase35ExistingAdlerAnswers` passed in both engines; source imports and cleanup path inspected. |
| 5 | No-feature C output remains frozen; new operators/refusals are registered and manifest controls are reached. | ✓ VERIFIED | `TestPhase35PreChangeEmitterCorpus` and controls passed across the three emitter APIs; current manifest/census enforcement is present and fresh receipt roots pass. |
| 6 | The pre-edit corpus pins tracked source, old C/core identity, explicit refusals and all three emitter entry points. | ✓ VERIFIED | Frozen corpus and comparator were inspected and the exact named corpus/control tests passed. |
| 7 | Legacy scalar `010`/`0B10` route meanings and Bool identifiers remain unchanged. | ✓ VERIFIED | Numeric baseline fixtures and both checker conversions inspected; current regression receipt binds the route controls. |
| 8 | Prior nested-index admission remains distinct from the new depth-64 amendment. | ✓ VERIFIED | Before-state and resource-delta fixtures are separate; baseline tests pass in the bound regression run. |
| 9 | `x ^ (y << 3) & 0xFF` parses as `x ^ ((y << 3) & 0xFF)` while preserving number-leaf identity. | ✓ VERIFIED | Pratt parser and transparent group AST inspected; `TestPhase35ScalarGrammar` and formatter round-trip pass. |
| 10 | Depth 64 succeeds and 65 refuses across groups/indices; adjacency, formatter boundaries and sibling recovery remain defined. | ✓ VERIFIED | Bounded parser counter and named refusal inspected; depth/grammar roots pass, including 63/64/65/128 boundary cases. |
| 11 | Hex spelling admission is narrow; malformed-token ownership, route-specific overflow codes and old number policies remain intact. | ✓ VERIFIED | Lexer exact-span and route-local conversion inspected; numeric ownership and formatter spelling tests pass. |
| 12 | Five exact-width pure operations lower to core; shifts carry only canonical decimal `ConstU64` 0..63 and empty `RightID`. | ✓ VERIFIED | Checker lowering and core operation registration inspected; producer/shape controls are green in current receipts. |
| 13 | Non-number shift amounts refuse even when compile-time constant, preserving lexical/overflow ownership. | ✓ VERIFIED | Checker requires a number AST leaf before lowering; named shift refusal and admission mutants are bound and green. |
| 14 | The checker allows only one unused Byte formal for U64/Bool work and refuses Byte use/read/copy/conversion/operands without changing arity or U64 input requirements. | ✓ VERIFIED | Source bridge/refusal code inspected; `TestPhase35UnusedByteSourceAdmission` is in the fresh run roots. |
| 15 | Checker transfer mutants for every new operation reach independent scalar facts and safety sinks. | ✓ VERIFIED | Five transfer arms and their reached mutants are present; parent’s hash-bound mutation logs show all 5 checker cases and 15 total transfer mutants killed. |
| 16 | Core and origin validators independently reject malformed shift shape, type and range without producer provenance. | ✓ VERIFIED | Both validators inspect raw fields and enforce immediate/type/fresh-target laws independently; their fresh raw-log roots and negative controls pass. |
| 17 | Checker, corevalidator and originvalidator independently derive all five transfers, with faults reached through real analysis paths. | ✓ VERIFIED | Distinct transfer implementations and path-triggered mutants inspected; current logs show 15/15 transfer mutants killed. |
| 18 | Both IR peers allow only the unused Byte bridge and reject Byte read/copy/conversion/operand use. | ✓ VERIFIED | Independent peer shape laws and their admission controls are present and pass in the current receipt roots. |
| 19 | Old core/identity/schema bytes remain stable; wrong-kind immediates and nonempty shift `RightID` are rejected independently. | ✓ VERIFIED | Pre-change corpus and raw core/origin field checks inspected; freshness roots pass the wrong-shape controls. |
| 20 | Interpreter computes exact-width values without adding events, defects or reasons. | ✓ VERIFIED | `TestPhase35BitwiseShiftVectorsEventFree` passed all 33 cases; interpreter operation arms only produce U64 values. |
| 21 | Pathoracle recognizes all new value operations and preserves loan-free ancestry. | ✓ VERIFIED | Dispatch and phase tests inspected; current regression raw logs contain passing pathoracle roots. |
| 22 | Executionpeer permits event-free operations and rejects forged step/defect events without numeric replay. | ✓ VERIFIED | `eventFreeOperationKinds` contains all five; causal controls pass. Numeric proofs remain with value-knowing interpreter/native peers. |
| 23 | Byte stays Byte at the unused bridge and produces U64 locals without conversion. | ✓ VERIFIED | Source/core fixture and public conformance path preserve Byte input; the bridge contract is exercised in current receipts. |
| 24 | Generated shifts use checked 0..63 literals, unsigned non-promoted operands and exact-width casts; the site scan is non-vacuous. | ✓ VERIFIED | Both simple and branched emitter paths emit unsigned operands and `uint64_t` results; tracer and positive-site/unsafe controls pass. |
| 25 | Event projection, schema/capacity, scalar signatures and path selectors recognize all five operations without new events or optimizer facts. | ✓ VERIFIED | Emitter/native inventories and exhaustive operation selectors inspected; current cgen roots and regression log roots pass. |
| 26 | The unused-Byte conformance signature returns U64 without changing application entry, defaults or conversions. | ✓ VERIFIED | Fixture harness retains Byte input and U64 result; public conformance witness is in the fresh receipts. |
| 27 | Old generated C/core/schema identifiers remain byte-identical outside declared feature/resource amendments. | ✓ VERIFIED | Frozen pre-change comparisons and current regression receipts show the no-feature baseline remains unchanged; current HEAD adds no post-20a source changes. |
| 28 | Public interpreter/native routes return 264, alternate association returns 8, and application output is `264\n`. | ✓ VERIFIED | Independent literal matrix pins both associations; public witness and application-output roots pass in the receipt logs. |
| 29 | Literal values hold at O0/O3/O3-LTO and shift UBSan; engine agreement alone is not numeric proof. | ✓ VERIFIED | Direct matrix (33 vectors) and UBSan (16 shift cases) passed; independent expected values and wrong-answer controls are hash-bound. |
| 30 | `schway.surface/1` retains ten top-level fields and source-free discovery while documenting operators, grouping, precedence, counts, low64, hex and depth. | ✓ VERIFIED | Manifest and guide inspected; source-free route, surface tests and removal/false-admission controls pass in current receipts. |
| 31 | Complete-source census and admission/limit/removal/false-admission/hint mutants stay reached without weakening legacy tests. | ✓ VERIFIED | Independent witnesses/census and bidirectional controls are wired to manifest tests; current regression logs pass. |
| 32 | Operator adjacency, missing/equal operands and left association retain pinned behavior. | ✓ VERIFIED | Grammar suite covers touching/missing operators and association; `TestPhase35ScalarGrammar` passed. |
| 33 | Shift boundaries, grouping, nonliteral refusal, low64 behavior and logical right shift are exact. | ✓ VERIFIED | Syntax/parser admission plus interpreter/native vector tests passed; UBSan cases are clean. |
| 34 | Maximum U64 with leading zeros succeeds; max+1 fails without floating rounding, signed conversion or digit cutoff. | ✓ VERIFIED | Numeric ownership tests passed; arbitrary-precision lexical path and named overflow were inspected. |
| 35 | Adler edge fixtures preserve outputs/import/cleanup/order and reach the wrong-shift control. | ✓ VERIFIED | Direct Adler answer test passed; wrong-shift and imported source controls pass in hash-bound logs. |
| 36 | Evidence summaries use actual runs and a closed coverage vocabulary, separating fresh checks, inspection, history and incomplete hosts. | ✓ VERIFIED | I rehashed all receipt inputs/logs and parsed every raw test root; prohibition adapter test passed. Current host scope is explicitly local Darwin/arm64. |
| 37 | Dormant seeds keep explicit consumer/trigger dispositions; living-document updates derive from inspected current facts. | ✓ VERIFIED | Phase transition artifacts record the dispositions; fresh decision-coverage query reports 23/23 honored. Mutable global roadmaps were inspected but are not in the fingerprint. |
| 38 | Adler/resource changes are explicit deltas while old C/core/ID/schema/numeric baselines remain frozen. | ✓ VERIFIED | Compatibility delta and frozen baseline fixtures align; pre-change emitter and numeric ownership tests passed. |

**Score:** 38/38 truths verified (0 present, behavior-unverified).

### Current source delta and evidence

The original report bytes are preserved in `35-371-REFRESH-2026-10-16-VERIFICATION-HISTORY.md` (SHA-256 `ecaeaa79f187f1f90ecbea3c63a1d2d6abf06b42fe13d369ecf77439f9e05787`). Its five Phase 35-covered production paths were rechecked against the Phase 37.1 delta: `cgen_program.go`, `corevalidate.go`, `executionpeer.go`, `syntax/parser.go`, and the adjacent Phase 33 terminal test. The code changes validate defect-reason UTF-8/provenance and terminal serialization; the Phase 35 scalar transfer, shift-width, and lexer/Pratt behavior remains unchanged. The five pure operations remain in `eventFreeOperationKinds`; the current causal test exercises normal event-free admission and fabricated-event rejection.

Current source HEAD is `6572cd0ccf8d9f778d2ea22e4d6e498891c64f32`. The Phase 37.1 source delta ends at `20a519468839a0698e2c4e93fc51300509c6abbf`; `20a519..6572cd` adds no Go, C, or Schway source changes. The new compatibility tests and fixtures are included in the fingerprint below.

The parent-produced bounded regression receipt at `.planning/phases/37.1-close-gap-legacy-terminal-outcomes/371-REFRESH-COMPLETION-2026-10-16.json` records 250 package/test selectors, 251 root events, 240 distinct names, and zero missing, failed, or skipped roots. It reused 220 unchanged green roots with their original source basis, reran five formerly failing roots, and ran 26 previously unrun roots; the pressure selector was excluded. I inspected the current Phase 35 receipt logs and independently ran the focused checks listed below. This is bounded evidence, not a full workspace or two-host run.

| Current raw log slice | SHA-256 | Phase 35 evidence |
|---|---|---|
| `371-REFRESH-SYNTAX-20261016.jsonl` | `ff45e01c6fb34463137cde721db26c3f6e19b95ac5ecb03aaee872b5b91a9b77` | Lexer/parser and numeric ownership roots. |
| `37.1-371-REFRESH-REGRESSION-03.jsonl` / `-04.jsonl` | `b3e09bace8c936ed185d3eb0f051f9f98a5df6ea41c885b3abfb0cd5697026ed` / `fc9196be67943a5c437a9143a376c0c3e860f8f19d6354725e5774b690e7a957` | Native tracer, unsigned shift sites, and emitter controls. |
| `37.1-371-REFRESH-REGRESSION-05.jsonl` / `-06.jsonl` / `-07.jsonl` | `447e2cc026b29fe8d4035f02ed03e9e3e97c274e718cc3b23f298e24ad5e2743` / `92a2d0ada2c03af592bb2d171389048bdf3dd34fac68049a540920b8acc4190c` / `8d57863fb24598246e02d7d2d888303677f7013149ab3b173c2f6cb3f55a729a` | Checker admission/transfers, core shape, and independent core transfer roots. |
| `37.1-371-REFRESH-REGRESSION-09.jsonl` / `-10.jsonl` | `01b5bf38c5319f6378d153e7499de3872ddd955785842ddf09027145e2cffc3c` / `accf0ed71574c43316b4cbc24f0a7fe62fbc203946651a6a7c1b08ef17c2ea89` | Causal event controls, interpreter values, and wrong-answer controls. |
| `37.1-371-REFRESH-REGRESSION-12.jsonl` / `-13.jsonl` | `30241c989eda846408b4c4b48330ff2af7b35c4099e147010a6fbaee8114760e` / `9926fe611d6364d952369e2581cc2b9015045f8bdc2c0be0fe34cc96275e7a9c` | Independent origin transfers/admission and pathoracle ancestry. |
| `37.1-371-REFRESH-REGRESSION-15.jsonl` | `3a0329496a644dce9acd534d6177437f64f51825d4cb11bed323ba51045eb1c9` | Adler, admission/transfer mutants, value matrix, shift UBSan, surface and evidence roots. |

Current regression/security/review gates are recorded in `35-371-REFRESH-GATES.md`: a 19-root/254-subtest D18 proof union has zero skips; the current delta review is clean; supplemental security controls close 13/13 with zero open; UI is not applicable. These are parent-produced results, distinguished from my source inspection and focused tests. Phase 35 makes no Linux or hosted-receipt claim; the current directly executed native lanes are local Darwin/arm64.

The C emitter shifts an `unsigned long long` operand by a checked decimal literal in 0..63 and casts the result to `uint64_t`. The count stays below the operand width and unsigned left shift supplies low-64-bit behavior after the cast. The interpreter computes on `uint64`; these pure operations add no defect events, terminal reasons, or optimizer facts. Existing Phase 33 audit obligations for unrelated legacy defect paths remain separate and are not closed by this Phase 35 result.

**Verification boundary:** The fingerprint covers the Phase 35 plan/summaries automatically, the original Phase 35 must-have sources/tests/fixtures, current changed test/data files, and the current receipts/raw logs. Planning-root roadmaps are inert inputs and are not fingerprinted.

### Required artifacts

The GSD artifact query passed all 48 plan declarations (45 unique paths). Source implementations, tests, fixtures and public witnesses were checked for substantive content and wiring through the links below. The main implementations are `syntax/parser.go`, `check/scalar.go`, `core/core.go`, both independent validators, `interp/interp.go`, `cgen/cgen_program.go`, and their phase tests. The frozen corpus, literal vectors, shift-site manifest and surface census are consumed by their named comparators rather than merely existing.

| Artifact group | Status | Evidence |
|---|---|---|
| Parser, lexer, formatter and refusal registry | ✓ VERIFIED | Source/diagnostic ownership inspected; grammar, depth, numeric and round-trip tests pass. |
| Checker lowering and core operation contracts | ✓ VERIFIED | Independently registered op kinds and checker transfer tests; 48/48 plan artifact declarations passed. |
| Core/origin validators, interpreter, pathoracle and executionpeer | ✓ VERIFIED | Separate laws and consumers inspected; current raw test roots and focused interpreter test pass. |
| C emitter/native runner, manifests and public conformance | ✓ VERIFIED | Safe shift emission inspected; native tracer, frozen emitter corpus, value matrix and manifest roots pass. |
| Source examples, docs, vectors and compatibility/evidence records | ✓ VERIFIED | Examples/tests consume the fixtures; Adler and closeout evidence roots pass. |

### Key link verification

The generic key-link query found 1/24 by literal path-reference matching. That result was insufficient for these function/operation/fixture links, so I traced every declared connection through implementation or its named consumer. The traces below are supported by current source and the exact tests or freshly hash-checked roots noted in the details.

| From | To | Via | Status | Details |
|---|---|---|---|---|
| pre-change session comparator | C emitter | frozen inventory to all emitter outputs | ✓ WIRED | `TestPhase35PreChangeEmitterCorpus` compares the three entry points. |
| numeric baseline test | checker scalar conversion | baseline to both routes | ✓ WIRED | Baseline fixture is consumed by both conversion routes and their controls. |
| parser | formatter | grouped operator AST round trip | ✓ WIRED | Parser and formatter test passes, including spelling preservation. |
| lexer | checker | token/span to route-local U64 conversion | ✓ WIRED | Numeric fixtures and route refusal consumers are present in current roots. |
| parser | refusal registry | bounded depth refusal and recovery | ✓ WIRED | Named refusal is registered; depth and sibling recovery tests pass. |
| parser | checker scalar | literal number leaf to shift immediate | ✓ WIRED | Source-only AST check enforces the literal rule; shift admission controls pass. |
| checker scalar | core | five operation kinds and transfer facts | ✓ WIRED | Producer registers operation and expected fields; transfer roots pass. |
| checker Byte bridge | checker refusal path | unused formal only | ✓ WIRED | The single unused formal is preserved; use/conversion paths refuse. |
| core | corevalidate | raw operation fields to independent shape/range law | ✓ WIRED | Core validator checks direct and CFG paths; raw roots pass. |
| core | originvalidate | raw fields to independent origin law | ✓ WIRED | Origin recomputation/publish paths and their negative controls pass. |
| originvalidate transfer | observed fact/safety obligation | peer-specific transfer to reached fault | ✓ WIRED | Analysis-trigger mutants are in the bound Phase 35 roots. |
| validated core | interpreter | runtime values and event projection | ✓ WIRED | 33 interpreter vectors passed with event-free assertions. |
| core operation kinds | pathoracle | ancestry dispatch | ✓ WIRED | Five kinds dispatch through the pathoracle consumer; phase roots pass. |
| interpreter output | executionpeer | causal event classification | ✓ WIRED | No-event admission and fabricated step/defect controls pass; no numeric replay is claimed. |
| value vectors | interpreter tests | independent literals to comparator | ✓ WIRED | Fixture is read by the direct interpreter vector test. |
| core shift immediate | C emitter | checked literal to actual unsigned C site | ✓ WIRED | C source arms, native tracer, site scan and compiled matrix pass. |
| C emitter | native inventories | pure kind to event/schema projection | ✓ WIRED | Inventories recognize the ops; no new event/optimizer fact is emitted. |
| core Byte signature | C emitter fixture harness | preserved Byte input to U64 result | ✓ WIRED | Signature and Byte fixture consumers are covered in current roots. |
| bitwise example | public CLI test | public source to interpreter/native outcomes | ✓ WIRED | Public witness roots pass; literal outputs are pinned independently. |
| value vectors | session matrix | fixture to O0/O3/LTO/UBSan | ✓ WIRED | Direct 33-vector matrix and 16-case UBSan tests pass. |
| surface manifest | surface session tests | source census and drift controls | ✓ WIRED | Session surface roots pass with non-vacuous census fixtures. |
| Adler source | Adler session test | `b << 16` to checksum/adler_low | ✓ WIRED | Direct both-engine output test passes. |
| evidence index | closeout test | run references to closed evidence kinds | ✓ WIRED | Prohibition adapter and closeout evidence roots pass. |
| evidence index | language maturity record | inspected facts to dated capability disposition | ✓ WIRED | Current transition artifacts and 23/23 decision coverage support the scoped link. |

### Data-flow trace

| Artifact | Value | Source | Produces real data | Status |
|---|---|---|---|---|
| `examples/bitwise.schway` | witness result | Parsed and checked source locals executed by interpreter/native compiler | Yes; both engines produce independently pinned 264, and alternate association produces 8. | ✓ FLOWING |
| `examples/checksum/adler32.schway` | checksum and `adler_low` | Imported Schway source with `b << 16` | Yes; existing checksum answers are observed through both engines. | ✓ FLOWING |

### Behavioral Spot-Checks

All Go commands used `GOCACHE=/tmp/ai-lang-verification-gocache-count1-v120`, `-count=1`, and `-timeout=600s`. Each command selected named Phase 35 tests; no full package or workspace suite was run.

| Behavior | Command | Result | Status |
|---|---|---|---|
| Values agree at O0/O3/O3-LTO | `go test ./internal/compiler/session -run '^TestPhase35BitwiseOptimizationMatrix$' -count=1 -v -timeout=600s` | Exit 0; 33 independently expected vectors pass across all three native lanes. | ✓ PASS |
| C17-safe shift results | `go test ./internal/compiler/session -run '^TestPhase35ShiftUBSan$' -count=1 -v -timeout=600s` | Exit 0; all 16 shift UBSan cases pass. | ✓ PASS |
| Checker literal-shift admission and transfer | `go test ./internal/compiler/check -run '^(TestPhase35ShiftAdmission|TestPhase35ScalarTransferMutants)$' -count=1 -v -timeout=600s` | Exit 0; valid counts and invalid/nonliteral controls plus five operator transfers pass. | ✓ PASS |
| Independent core validation | `go test ./internal/compiler/corevalidate -run '^TestPhase35ScalarTransferMutants$' -count=1 -v -timeout=600s` | Exit 0; core validator rejects five transfer faults and CFG variants. | ✓ PASS |
| Independent origin validation | `go test ./internal/compiler/originvalidate -run '^TestPhase35ScalarTransferMutants$' -count=1 -v -timeout=600s` | Exit 0; origin validator rejects five transfer faults and CFG variants. | ✓ PASS |
| Parser, numeric ownership, and spelling | `go test ./internal/compiler/syntax -run '^(TestPhase35ScalarGrammar|TestPhase35NumericOwnership|TestPhase35FormatterRoundTrip|TestPhase35ExpressionDepth)$' -count=1 -v -timeout=600s` | Exit 0; precedence/grouping, hex case/spelling, overflow/malformed refusals, and depth boundaries pass. | ✓ PASS |
| Emitted unsigned shift sites and event projection | `go test ./internal/compiler/cgen -run '^TestPhase35UnsignedShiftEmissionAndEventProjection$' -count=1 -v -timeout=600s` | Exit 0; checked C sites and pure event projection pass. | ✓ PASS |
| Event-free causal controls | `go test ./internal/compiler/executionpeer -run '^TestPhase35PureOperationCausalControls$' -count=1 -v -timeout=600s` | Exit 0; all five operations admit no-event execution and reject fabricated events. This is causal evidence only. | ✓ PASS |
| Frozen emitter compatibility and Adler answers | `go test ./internal/compiler/session -run '^(TestPhase35PreChangeEmitterCorpus|TestPhase35PreChangeCorpusControls|TestPhase35ExistingAdlerAnswers)$' -count=1 -v -timeout=600s` | Exit 0; old emitter outputs stay frozen and checksum/adler_low answers match on both engines. | ✓ PASS |
| Engine values and wrong-answer controls | `go test ./internal/compiler/session -run '^(TestPhase35FocusedValueAnswers|TestPhase35EngineWrongAnswerControls)$' -count=1 -v -timeout=600s` | Exit 0; exact values pass and all five wrong-answer controls are rejected. | ✓ PASS |
| Public source witness | `go test ./cmd/schway -run '^TestPhase35PublicConformanceWitness$' -count=1 -v -timeout=600s` | Exit 0; public interpreter/native witness returns 264, alternate grouping returns 8. | ✓ PASS |
| Evidence-prohibition adapter | `node --test testdata/phase35/prohibition-evidence.test.cjs` | Exit 0; the active adapter invokes the named Go evidence-grounding test and rejects false causal/lane claims. | ✓ PASS |

These checks complement the current hash-bound receipt roots above. The native matrix and UBSan runs were directly re-executed on the current checkout; the rest are exact named tests, not broad suites.

### Requirements coverage

All eight plans declare BIT-01 through BIT-04 only. `REQUIREMENTS.md` assigns exactly those four to Phase 35; no orphaned Phase 35 requirement was found.

| Requirement | Source plans | Description | Status | Evidence |
|---|---|---|---|---|
| BIT-01 | 03, 05, 06, 07 | U64 `^`, `&`, `|`; independent scalar transfers and literal answers across engines/optimization levels; causal event-free checks remain distinct from numeric proof. | ✓ SATISFIED | Independent vector and optimization tests pass; 15/15 parent-reported transfer mutants killed; causal adapter controls pass. |
| BIT-02 | 02, 03, 04, 06, 07 | Literal U64 shifts only for counts 0..63; named refusals for invalid/nonliteral shifts; defined low64 behavior and no unsafe generated C. | ✓ SATISFIED | Literal admission, unsigned C-site inspection, 33-vector matrix, positive unsafe-site controls and 16-case UBSan pass. |
| BIT-03 | 02, 07 | Case-insensitive hex digits, exact spelling retention, malformed/overflow refusal with named codes. | ✓ SATISFIED | Lexer/parser/formatter and named refusal tests pass. |
| BIT-04 | 08 | Adler uses `b << 16`; pinned outputs remain identical on interpreter and native. | ✓ SATISFIED | Direct `TestPhase35ExistingAdlerAnswers` passed for both engines. |

### Test Quality Audit

Requirement-linked Phase 35 tests were scanned for disabled/skipped patterns and circular fixture generation. Writes in requirement-linked tests create temporary alternate source, corpus inputs, mutation overlays, or compiler wrappers; tests do not write expected value fixtures from the system under test. `value-vectors.json` contains the fixed decimal expectations consumed by the interpreter/native comparators.

| Test files | Linked requirements | Active / skipped | Circular | Assertion strength | Verdict |
|---|---|---|---|---|---|
| `syntax/*phase35*`, `syntax/numeric_phase35_test.go` | BIT-02, BIT-03 | Active; 0 disabled | No | Value, refusal, spelling, recovery | PASS |
| `check/check_phase35*`, `corevalidate/corevalidate_phase35_test.go`, `originvalidate/originvalidate_phase35_test.go` | BIT-01, BIT-02 | Active; 0 disabled | No | Independent transfer and negative controls | PASS |
| `interp/interp_phase35_test.go`, `session/session_phase35_values_test.go`, `cmd/schway/main_phase35_test.go` | BIT-01, BIT-02, BIT-03 | Active; 0 disabled | No | Fixed expected values and public behavior | PASS |
| `cgen/cgen_phase35_test.go`, `executionpeer/executionpeer_phase35_test.go`, `pathoracle/pathoracle_phase35_test.go` | BIT-01, BIT-02 | Active; 0 disabled | No | Generated-site, causal, and ancestry assertions | PASS |
| `session/session_phase35_adler_test.go`, `session/session_phase35_prechange_test.go` | BIT-04, compatibility | Active; 0 disabled | No | Both-engine expected answers and frozen output comparisons | PASS |
| `testdata/phase35/prohibition-evidence.test.cjs` | BIT-01, BIT-02 prohibitions | Active; 0 skipped | No | Violation fixtures plus clean evidence subject | PASS |

**Disabled requirement tests:** 0. **Circular expected-value patterns:** 0. **Insufficient assertions:** 0.

### Test-tier prohibitions

Both plan prohibitions are wired to `testdata/phase35/prohibition-evidence.test.cjs`; current `35-PROHIBITION-ENFORCEMENT.json` reports `green`, `flagged: false`, and `failFirst: true` for each.

| Requirement | Must not claim | Enforcement evidence | Result |
|---|---|---|---|
| BIT-01 | Engine agreement or causal event admission is independent numeric proof. | Violation fixture plus clean Phase 35 evidence subject; named adapter test passed. | ✓ VERIFIED |
| BIT-02 | Historical hosted receipts or skipped optimization/sanitizer lanes are fresh Phase 35 evidence. | Lane-claim violation fixture plus clean Phase 35 evidence subject; named adapter test passed. | ✓ VERIFIED |

Neither test-tier prohibition is unverified or flagged.

### Decision Coverage

The fresh non-blocking context decision query returned `total: 23`, `honored: 23`, and `not_honored: []`. The relevant `z1g`, `trz`, and `z1s` seeds remain dormant on their named current source witnesses; no seed was promoted by this Phase 35 refresh.

### Anti-patterns and probes

No unreferenced `TBD`, `FIXME` or `XXX` marker was found in the inspected changed implementation. The `placeholder` word in two existing `check.go` comments describes an intentionally unused field; it is not a stub or debt marker. No empty implementation or disconnected data path was found in the phase-owned code.

| Probe | Command | Result | Status |
|---|---|---|---|
| No phase-declared or conventional phase probe was found | Searched all eight PLAN/SUMMARY files and `scripts/**/tests/probe-*.sh` | No probe path present. | N/A |

### Scope and outstanding audit debt

The Phase 33 ledger still records CR-03 and CR-04 as deferred BLOCKER audit obligations, and CR-06 as a deferred BLOCKER for the legacy no-call native scalar defect path. CR-05 is fixed for dynamic scalar callers. Phase 36 WR-02 remains deferred comment warning. These are not Phase 35 requirements and do not falsify the added operators’ defined numeric semantics or the “no new defect path” result. In particular, preservation of legacy no-call C bytes does not establish native no-call defect-document parity. The Phase 35 checks establish that the new pure operators themselves produce no new defect/event path; the external obligations remain open for the required audit before Phase 38.

### Human Verification Required

N/A — infrastructure/foundation phase with no visual or external-service UX. All acceptance criteria are verifiable programmatically; there are no behavior-unverified truths. There is no Phase 35 gap to defer to a later roadmap phase.

### Gaps summary

No Phase 35 goal gap was found. The independent source inspection and focused checks support the 38/38 truths. External Phase 33 audit obligations remain visible above and are not closed by this verification.

---

_Verified: 2026-10-16T21:13:17Z_
_Verifier: gsd-verifier agent_
