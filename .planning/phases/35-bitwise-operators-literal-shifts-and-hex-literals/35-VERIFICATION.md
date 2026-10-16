---
phase: 35-bitwise-operators-literal-shifts-and-hex-literals
verified: 2026-10-16T03:41:27Z
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
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-COMPATIBILITY-DELTA.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-EVIDENCE.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REFRESH3-GATES.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REFRESH3-PROHIBITION.json"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REFRESH3-RUN.json"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REFRESH3-VALIDATION.json"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REVIEW-DISPOSITION.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-SECURITY.md"
  - ".planning/phases/36-scalar-helper-calls/36-REFRESH2-GATES.md"
  - "cmd/schway/main_phase35_test.go"
  - "docs/agent-surface.md"
  - "examples/bitwise.schway"
  - "examples/checksum/adler32.schway"
  - "internal/compiler/cgen/cgen.go"
  - "internal/compiler/cgen/cgen_phase33_scalar_defect_test.go"
  - "internal/compiler/cgen/cgen_phase35_test.go"
  - "internal/compiler/cgen/cgen_phase36_caller_shapes_test.go"
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
covered_digest: "v3:sha256:45003bdd8adb489f20e3b92db061704de08396230cc788f48d474570f48b6dfb"
behavior_unverified: 0
overrides_applied: 0
decision_coverage: "23/23 CONTEXT decisions honored"
---

# Phase 35: Bitwise Operators, Literal Shifts, and Hex Literals Verification Report

**Phase Goal:** Ordinary Schway scalar code can use U64 `^`, `&`, `|`, literal-amount `<<` and `>>`, and hex literals, with one defined meaning on every engine and optimization level, no C17 shift undefined behavior, and no new defect path.
**Verified:** 2026-10-16T03:41:27Z
**Status:** passed
**Prior report:** passed, 38/38; no `gaps:` section. This was a freshness verification, not a gap-closure run.

## Goal Achievement

### Observable Truths

The contract is the five roadmap success criteria plus the 34 unique plan truths. One plan truth restates the Adler roadmap criterion, giving 38 distinct truths.

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | U64 xor, and, and or have independently pinned values in interpreter and native O0, O3, and O3-LTO; check/core/origin derive transfers independently and executionpeer checks event-free causality. | ✓ VERIFIED | Current value matrix and interpreter vector tests passed; REFRESH3 raw logs independently parsed with 15/15 transfer mutants killed and all 33 vectors passing at O0/O3/LTO. Executionpeer’s causal-only boundary is preserved. |
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
| 27 | Old generated C/core/schema identifiers remain byte-identical outside declared feature/resource amendments. | ✓ VERIFIED | Frozen pre-change comparisons pass; source revision `9bc84d95` has no source change relative to HEAD `04f0ff1b` (only later gate docs). |
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

### Freshness evidence and boundaries

I independently verified receipt inputs and raw log hashes against current files and parsed the raw Go test logs. `35-REFRESH3-VALIDATION.json` binds 122/122 current inputs and 13/13 logs; its 72 expected root executions have zero missing, failed or skipped roots. `35-REFRESH3-RUN.json` binds 181/181 inputs and 16/16 logs; its 252 executions represent 251 unique package/name roots and 240 distinct test names, all passing with zero missing/fail/skip. The one-count differences are documented duplicate selectors, not omitted results. These are parent-produced runs; my independent work was hash checking and raw-log parsing, not rerunning their full matrix.

The current source revision is `9bc84d95da311beb24cf3e5621f1483346eaf01b`; current HEAD `04f0ff1b` adds only gate documents. I independently ran the focused checks below. Parent evidence also records 15 transfer mutants and 26 admission mutants killed, 33-vector O0/O3/LTO and 16-case UBSan lanes, 24/24 authored security mitigations closed, and a 29-file independent review with zero new Phase 35 findings. The last items are parent evidence and are labeled as such; the mutation suite was not duplicated.

The implementation gives each new operation a pure U64 result. C shifts use unsigned widened values and checker-validated literal counts from 0 to 63, so C17’s signed/oversized shift cases do not arise on these sites. The operations do not introduce defect events or reasons. This establishes the Phase 35 “no new defect path” contract for the added operations; it does not assert that every pre-existing defect path has parity.

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

### Focused behavioral checks run for this verification

Every Go command used the isolated cache `GOCACHE=/tmp/ai-lang-verification-gocache`, `-count=1`, and `-timeout=600s`.

Three grouped Go commands are shown exactly as run:

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax -run '^(TestPhase35ScalarGrammar|TestPhase35FormatterRoundTrip|TestPhase35ExpressionDepth|TestPhase35NumericOwnership)$' -count=1 -v -timeout=600s
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -run '^(TestPhase35NativeShiftTracer|TestPhase35ShiftSiteControls|TestPhase33ScalarExplicitDefectShells|TestPhase36UnusedByteStructuredHelper)$' -count=1 -v -timeout=600s
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase35PreChangeEmitterCorpus|TestPhase35PreChangeCorpusControls|TestPhase35ExistingAdlerAnswers)$' -count=1 -v -timeout=600s
```

| Behavior | Result | Status |
|---|---|---|
| Grammar, formatting, depth and number ownership | Exit 0; all 4 roots and boundary subcases pass. | ✓ PASS |
| Interpreter bitwise/shift vectors and event freedom — `go test ./internal/compiler/interp -run '^TestPhase35BitwiseShiftVectorsEventFree$' -count=1 -v -timeout=600s` | Exit 0; 33 vector cases pass. | ✓ PASS |
| Native shift C tracer/site controls and adjacent defect/Byte compatibility tests | Exit 0; four named roots pass. The defect test exercises dynamic caller shells and does not prove no-call parity. | ✓ PASS |
| Frozen emitters and Adler outputs | Exit 0; frozen corpus, controls and both-engine Adler outputs pass. | ✓ PASS |
| Optimization matrix — `go test ./internal/compiler/session -run '^TestPhase35BitwiseOptimizationMatrix$' -count=1 -v -timeout=600s` | Exit 0; 33 vectors across O0/O3/O3-LTO pass. | ✓ PASS |
| Shift undefined-behavior sanitizer — `go test ./internal/compiler/session -run '^TestPhase35ShiftUBSan$' -count=1 -v -timeout=600s` | Exit 0; all 16 UBSan shift cases pass. | ✓ PASS |
| Test-tier prohibition adapter — `node --test testdata/phase35/prohibition-evidence.test.cjs` | Exit 0; violation fixture proves fail-first, clean fixture passes. Both BIT-01 and BIT-02 prohibitions are green/unflagged in the receipt. | ✓ PASS |

### Requirements coverage

All eight plans declare BIT-01 through BIT-04 only. `REQUIREMENTS.md` assigns exactly those four to Phase 35; no orphaned Phase 35 requirement was found.

| Requirement | Source plans | Description | Status | Evidence |
|---|---|---|---|---|
| BIT-01 | 03, 05, 06, 07 | U64 `^`, `&`, `|`; independent scalar transfers and literal answers across engines/optimization levels; causal event-free checks remain distinct from numeric proof. | ✓ SATISFIED | Independent vector and optimization tests pass; 15/15 parent-reported transfer mutants killed; causal adapter controls pass. |
| BIT-02 | 02, 03, 04, 06, 07 | Literal U64 shifts only for counts 0..63; named refusals for invalid/nonliteral shifts; defined low64 behavior and no unsafe generated C. | ✓ SATISFIED | Literal admission, unsigned C-site inspection, 33-vector matrix, positive unsafe-site controls and 16-case UBSan pass. |
| BIT-03 | 02, 07 | Case-insensitive hex digits, exact spelling retention, malformed/overflow refusal with named codes. | ✓ SATISFIED | Lexer/parser/formatter and named refusal tests pass. |
| BIT-04 | 08 | Adler uses `b << 16`; pinned outputs remain identical on interpreter and native. | ✓ SATISFIED | Direct `TestPhase35ExistingAdlerAnswers` passed for both engines. |

### Anti-patterns and probes

No unreferenced `TBD`, `FIXME` or `XXX` marker was found in the inspected changed implementation. The `placeholder` word in two existing `check.go` comments describes an intentionally unused field; it is not a stub or debt marker. No empty implementation or disconnected data path was found in the phase-owned code.

| Probe | Command | Result | Status |
|---|---|---|---|
| No phase-declared or conventional phase probe was found | Searched all eight PLAN/SUMMARY files and `scripts/**/tests/probe-*.sh` | No probe path present. | N/A |

### Scope and outstanding audit debt

The Phase 33 ledger still records CR-03 and CR-04 as deferred BLOCKER audit obligations, and CR-06 as a deferred BLOCKER for the legacy no-call native scalar defect path. CR-05 is fixed for dynamic scalar callers. Phase 36 WR-02 remains deferred comment warning. These are not Phase 35 requirements and do not falsify the added operators’ defined numeric semantics or the “no new defect path” result. In particular, preservation of legacy no-call C bytes does not establish native no-call defect-document parity. The Phase 35 checks establish that the new pure operators themselves produce no new defect/event path; the external obligations remain open for the required audit before Phase 38.

No human-only verification item remains for this software contract. UI applicability is false; no visual claim is made. There is no Phase 35 gap to defer to a later roadmap phase.

### Gaps summary

No Phase 35 goal gap was found. The independent source inspection and focused checks support the 38/38 truths. External Phase 33 audit obligations remain visible above and are not closed by this verification.

---

_Verified: 2026-10-16T03:41:27Z_
_Verifier: gsd-verifier agent_
