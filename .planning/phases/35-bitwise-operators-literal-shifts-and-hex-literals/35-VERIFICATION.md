---
phase: 35-bitwise-operators-literal-shifts-and-hex-literals
verified: 2026-10-16T01:20:31Z
status: passed
score: "38/38 truths verified"
covered_files:
  - ".planning/phases/33-carried-debt-bounded-capture-read-and-generic-freeze-gate/33-REFRESH2-FIX.md"
  - ".planning/phases/33-carried-debt-bounded-capture-read-and-generic-freeze-gate/33-REVIEW-DISPOSITION.md"
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
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-CONTEXT.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-EVIDENCE.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REFRESH2-GATES.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REFRESH2-PROHIBITION.json"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REFRESH2-RUN.json"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REFRESH2-VALIDATION.json"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-RESEARCH.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REVIEW-DISPOSITION.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REVIEW.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-SECURITY.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-VALIDATION.md"
  - ".planning/phases/36-scalar-helper-calls/36-REVIEW-DISPOSITION.md"
  - "cmd/schway/main_phase35_test.go"
  - "docs/agent-surface.md"
  - "examples/bitwise.schway"
  - "examples/checksum/adler32.schway"
  - "internal/compiler/cgen/cgen.go"
  - "internal/compiler/cgen/cgen_phase35_test.go"
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
  - "internal/compiler/surface/refusals.go"
  - "internal/compiler/surfacemanifest/manifest.go"
  - "internal/compiler/surfacemanifest/manifest_test.go"
  - "internal/compiler/syntax/format.go"
  - "internal/compiler/syntax/lexer.go"
  - "internal/compiler/syntax/numeric_phase35_test.go"
  - "internal/compiler/syntax/parser.go"
  - "internal/compiler/syntax/parser_phase35_baseline_test.go"
  - "internal/compiler/syntax/parser_phase35_test.go"
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
covered_digest: "v3:sha256:11dd587871b0e437e1d7954e488da261ceee59622140df929c93248c33937297"
behavior_unverified: 0
overrides_applied: 0
decision_coverage: "23/23 CONTEXT decisions honored"
---

# Phase 35: Bitwise Operators, Literal Shifts, and Hex Literals Verification Report

**Phase Goal:** Ordinary Schway scalar code can use U64 `^`, `&`, `|`, literal-amount `<<` and `>>`, and hex literals, with one defined meaning on every engine and optimization level, no C17 shift undefined behavior, and no new defect path.
**Verified:** 2026-10-16T01:20:31Z
**Status:** passed
**Re-verification:** No gap-closure re-verification. This is a freshness refresh of the prior passed report. The prior report had no `gaps:` frontmatter, so Step 0 selected initial verification mode; current source and refreshed receipts were checked again.

## Goal Achievement

### Observable Truths

The 38 truths below are the five non-negotiable roadmap criteria plus 34 plan truths, with the Phase 35-08 Adler statement merged into roadmap criterion 4 because it restates that contract.

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | U64 xor, and, and or produce independently pinned values in interpreter and native O0, O3, and O3-LTO; checker, core, origin, and execution peers have distinct roles and reached negative controls. | ✓ VERIFIED | Current focused run receipt: 184 distinct test names, 196 root executions, 16 packages, no failures/skips/missing; 15 transfer mutants and 26 admission mutants killed. Value vectors are independently authored; executionpeer tests causal events only. |
| 2 | Literal shifts at 0, 31, 32, and 63 agree across engines and optimization levels; left shift is low-64; invalid counts and nonliteral amounts are refused; emitted C shift sites are safe. | ✓ VERIFIED | Focused core/origin admission tests, native shift scanner and answer tests, O0/O3/LTO and UBSan rows in the refreshed receipts; source emits unsigned wide operands and checked literal amounts. |
| 3 | Hex accepts case-insensitive digits, preserves formatter spelling, and rejects malformed or overflowing tokens with named diagnostics. | ✓ VERIFIED | Syntax grammar, numeric ownership, formatter round-trip tests; lexer/parser and refusal code/manifest inspected. |
| 4 | Adler uses `b << 16`; pinned checksum and `adler_low` outputs stay byte-identical on both engines. | ✓ VERIFIED | `TestPhase35ExistingAdlerAnswers` and `TestPhase35AdlerWrongShiftControl` passed; source and imported consumer path inspected. |
| 5 | Programs without new numeric syntax preserve generated C bytes; all new operators and refusal codes are registered and manifest controls are reached. | ✓ VERIFIED | Frozen pre-change corpus compares all declared emitters; surface manifest and census tests pass, including removal and false-admission controls. |
| 6 | Pre-edit corpus pins tracked source, old C/core identity, explicit refusals, and all three emitter entry points. | ✓ VERIFIED | `session_phase35_prechange_test.go`, `prechange-corpus.json`, and refreshed regression evidence are present and hash-bound; baseline comparator executed. |
| 7 | Legacy scalar `010` stays 8, `0B10` stays 2, legacy `010` stays 10, `0B10` remains refused, and Bool identifiers stay legal. | ✓ VERIFIED | Numeric policy baseline and both checker conversion routes inspected; route policy tests pass. |
| 8 | Prior nested-index admission remains separate from the declared depth-64 resource amendment. | ✓ VERIFIED | Frozen baseline and phase-specific nested-index ledger distinguish the old path from the new depth boundary; baseline tests pass. |
| 9 | Grouping parses `x ^ (y << 3) & 0xFF` as `x ^ ((y << 3) & 0xFF)` while preserving number-leaf identity. | ✓ VERIFIED | Pratt parser, transparent group AST, formatter reparse path and grammar/formatter tests inspected and exercised. |
| 10 | Expression depth 64 succeeds and 65 refuses across groups and indices; adjacency, old generic and formatter boundaries, and sibling recovery remain defined. | ✓ VERIFIED | Bounded parser counter, named depth refusal and iterative recovery are present; `TestPhase35ExpressionDepth` and grammar tests pass. |
| 11 | Hex spelling is narrowly admitted; malformed token ownership, route-specific overflow codes, and old numeric policy remain intact. | ✓ VERIFIED | Lexer exact-span candidate ownership, parser diagnostics and numeric tests; case-preserving formatter golden passes. |
| 12 | Five exact-width pure operations lower to core; shifts carry only canonical decimal `ConstU64` 0..63 and empty `RightID`. | ✓ VERIFIED | `scalar.go` lowering and core op registrations inspected; checker/core contract tests pass. |
| 13 | Non-number shift amounts refuse even when compile-time constant, with lexical and overflow ownership retained. | ✓ VERIFIED | Source-only AST number-leaf check and named refusal are wired; shift-admission and route-policy controls pass. |
| 14 | The source checker admits only one unused Byte formal for U64/Bool local work and refuses Byte use, read, copy, conversion, and operands without changing arity or U64 input-required behavior. | ✓ VERIFIED | Checker bridge/refusal source and `TestPhase35UnusedByteSourceAdmission` pass. |
| 15 | Checker transfer mutants for each new operation are reached through independent scalar facts and safety sinks. | ✓ VERIFIED | Five checker transfer arms and per-operation mutants are exercised by fresh phase test selectors. |
| 16 | Both IR peers independently reject invalid shift shape, type, and range without relying on producer provenance. | ✓ VERIFIED | `corevalidate` and `originvalidate` each inspect raw operation fields and enforce type/range/fresh target; independent negative controls pass. |
| 17 | Checker, core validator, and origin validator independently derive all five value transfers, with faults reached through actual analysis paths. | ✓ VERIFIED | Separate transfer implementations and path-reached mutant tests in all three peers; refreshed transfer mutants show 15/15 killed. |
| 18 | Each IR peer permits only the unused Byte formal bridge and rejects Byte read, copy, conversion, or operation use. | ✓ VERIFIED | IR-level bridge checks independently enforce the narrow shape; source and IR controls pass. |
| 19 | Old core and identity/schema bytes stay stable; wrong-kind `ConstU64` and nonempty shift `RightID` are rejected independently. | ✓ VERIFIED | Pre-change golden plus core/origin field validation; shift target and raw-shape controls pass. |
| 20 | Interpreter evaluates exact-width values without adding events, defects, or reasons. | ✓ VERIFIED | Interpreter uses U64 bit operations and validated unsigned literal shifts; vector, no-event and wrong-answer controls pass. |
| 21 | The new operations are recognized by pathoracle and preserve loan-free ancestry. | ✓ VERIFIED | Pathoracle dispatch and phase tests inspected; current regression selectors include the pathoracle phase controls. |
| 22 | Executionpeer admits event-free operations and rejects forged per-operation step/defect events without pretending to replay numeric results. | ✓ VERIFIED | `eventFreeOperationKinds` includes all five; event fabrication controls pass. Independent value tests live with actual value-knowing peers. |
| 23 | The Byte fixture input stays Byte on the unused bridge and produces U64 locals without conversion. | ✓ VERIFIED | Interpreter/core fixture and public conformance path preserve the Byte argument; exact witness tests pass. |
| 24 | Generated shifts use a checked 0..63 literal, unsigned non-promoted operand, and final exact-width cast; the site scan is non-vacuous. | ✓ VERIFIED | Both simple and branched C emitter paths use unsigned wide operands and `uint64_t` casts; shift-site fixture has positive sites and unsafe-shift controls pass. |
| 25 | Event projection, schema size, capacity, scalar signature, and path selectors recognize all five pure kinds without new events or optimizer facts. | ✓ VERIFIED | Emitter/native inventories and exhaustive selector tests include the operation kinds; no event schema/optimizer additions found. |
| 26 | The unused-Byte conformance signature returns U64 and does not alter application entry, input defaults, or conversions. | ✓ VERIFIED | Existing fixture harness and public routes retain Byte input; source witness computes U64 locals; `TestPhase35PublicConformanceWitness` passes. |
| 27 | Old generated C and core/schema identifiers remain byte-identical outside declared feature/resource amendments. | ✓ VERIFIED | Baseline corpus was frozen before edits and compared after; source diff since receipt revision contains only planning gate documents. |
| 28 | Public interpreter and native conformance routes expose value 264, the alternate association yields 8, and application output is `264\n`. | ✓ VERIFIED | Public conformance test and focused value tests compare literal outcomes and application output separately through existing routes. |
| 29 | Literal vectors survive O0/O3/O3-LTO and shift UBSan; engine agreement alone is not numeric proof. | ✓ VERIFIED | Fresh local raw logs cover each optimization and UBSan lane; separate literal vectors and wrong-answer mutants provide the numeric oracle. |
| 30 | `schway.surface/1` retains ten top-level fields and source-free discovery while documenting operators, grouping, precedence, counts, range, low64, hex, and depth. | ✓ VERIFIED | Manifest fields and guide entries inspected; public manifest route and surface tests pass. |
| 31 | Complete-source census and admission/limit/removal/false-admission/hint mutants stay reached without weakening legacy tests. | ✓ VERIFIED | Independent census and source witnesses feed manifest tests; copied controls are present and pass. |
| 32 | Operator adjacency, missing operands, equal operands, and left association retain pinned behavior. | ✓ VERIFIED | Grammar and expression tests cover token adjacency, missing operands and association; lexer/parser cases are present. |
| 33 | Shift boundaries, adjacency, grouping, nonliteral refusal, low64 behavior, and exact logical right shift are defined. | ✓ VERIFIED | Boundary vectors include 0/31/32/63 and 64; typed validators and native/interpreter vector tests pass. |
| 34 | Maximum U64 with leading zeroes succeeds; max+1 fails without float rounding, signed conversion, or digit-count cutoff. | ✓ VERIFIED | Numeric lexer/parser tests compare exact integer values and named overflow; numeric ownership tests pass. |
| 35 | Adler edge fixtures preserve outputs, module import/cleanup/order behavior, and reach the wrong-shift control. | ✓ VERIFIED | Existing checksum and `adler_low` tests run both engines; imported source and cleanup path are inspected; wrong-shift control passes. |
| 36 | Evidence summaries use actual runs and a closed coverage-kind vocabulary, separating fresh checks, inspection, history, and incomplete host claims. | ✓ VERIFIED | Current RUN/VALIDATION/PROHIBITION JSON are source/log hash-bound; evidence test and TAP producer reject bad claims. |
| 37 | Dormant seeds retain explicit consumer/trigger dispositions; living docs update only from inspected current facts. | ✓ VERIFIED | Phase discussion/research and transition artifacts record the seed dispositions and source-backed updates; decision coverage is 23/23. |
| 38 | Adler and resource changes are explicit deltas while old C/core/ID/schema/numeric baselines remain immutable. | ✓ VERIFIED | Compatibility delta, pre-change corpus, numeric policy baseline, and old nested-index ledger align; baseline tests pass. |

**Score:** 38/38 truths verified (0 present, behavior-unverified).

### Required Artifacts

GSD `verify.artifacts` was run against all eight plans: 48 declarations, 45 unique paths, all passed their declared content predicates. Implementation artifacts were also source-read and traced into consumers/tests. No declared implementation is a stub.

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `internal/compiler/session/session_phase35_prechange_test.go` | Frozen baseline test | ✓ VERIFIED | Baseline comparator exercised. |
| `testdata/phase35/prechange-corpus.json` | Old source and C/core golden | ✓ VERIFIED | Revision-pinned input, comparator passed. |
| `testdata/phase35/numeric-policy-baseline.json` | Old numeric route policy | ✓ VERIFIED | Both checker routes compared. |
| `internal/compiler/syntax/parser.go` | Parser ownership, grouping, depth | ✓ VERIFIED | Source and grammar/depth tests. |
| `internal/compiler/syntax/format.go` | Stable hex spelling and grouping | ✓ VERIFIED | Formatter round-trip test. |
| `internal/compiler/surface/refusals.go` | Named shift/depth refusals | ✓ VERIFIED | Codes/hints registered and tested. |
| `internal/compiler/syntax/lexer.go` | Numeric token/span ownership | ✓ VERIFIED | Exact malformed/overflow diagnostics tested. |
| `internal/compiler/check/check.go` | Route integration and recovery | ✓ VERIFIED | Source path and check controls. |
| `internal/compiler/syntax/parser_phase35_test.go` | Grammar/recovery tests | ✓ VERIFIED | Named grammar selectors pass. |
| `internal/compiler/syntax/numeric_phase35_test.go` | Numeric lexer/parser tests | ✓ VERIFIED | Boundary and malformed controls pass. |
| `internal/compiler/check/check_phase35_numeric_test.go` | Route numeric tests | ✓ VERIFIED | Named numeric route controls pass. |
| `testdata/phase35/numeric-delta.json` | Declared numeric amendments | ✓ VERIFIED | Consumed by the baseline/delta checks. |
| `internal/compiler/core/core.go` | Five new core operation kinds | ✓ VERIFIED | Kinds are represented and consumed downstream. |
| `internal/compiler/check/scalar.go` | Checker lowering/transfers | ✓ VERIFIED | Literal admission and transfer tests pass. |
| `internal/compiler/core/core_phase35_test.go` | Core producer contract | ✓ VERIFIED | Producer and field-shape tests pass. |
| `internal/compiler/check/check_phase35_test.go` | Checker controls | ✓ VERIFIED | Five operation paths and mutants reached. |
| `internal/compiler/corevalidate/corevalidate.go` | Independent core validation | ✓ VERIFIED | Raw shape, type, range, target checks. |
| `internal/compiler/originvalidate/originvalidate.go` | Independent origin validation | ✓ VERIFIED | Separate transfers and safety checks. |
| `internal/compiler/corevalidate/corevalidate_phase35_test.go` | Core validator controls | ✓ VERIFIED | Admission and transfer mutant tests pass. |
| `internal/compiler/originvalidate/originvalidate_phase35_test.go` | Origin validator controls | ✓ VERIFIED | Admission and analysis-trigger tests pass. |
| `internal/compiler/interp/interp.go` | Interpreter operation semantics | ✓ VERIFIED | Literal value and event controls pass. |
| `internal/compiler/pathoracle/pathoracle.go` | Path ancestry dispatch | ✓ VERIFIED | Five kinds recognized by consumer. |
| `internal/compiler/executionpeer/executionpeer.go` | Event-free causal classification | ✓ VERIFIED | Pure op set includes all five kinds. |
| `internal/compiler/interp/interp_phase35_test.go` | Value and event tests | ✓ VERIFIED | Vectors and wrong-answer controls pass. |
| `internal/compiler/pathoracle/pathoracle_phase35_test.go` | Pathoracle tests | ✓ VERIFIED | Included in current named regression. |
| `internal/compiler/executionpeer/executionpeer_phase35_test.go` | Causal event tests | ✓ VERIFIED | Forged step/defect events rejected. |
| `examples/bitwise.schway` | Runnable bitwise witness | ✓ VERIFIED | Public witness returns the pinned outcome. |
| `internal/compiler/core/core_test.go` | Exhaustive operation registration | ✓ VERIFIED | New kinds registered across exhaustive sites. |
| `testdata/phase35/value-vectors.json` | Independent expected values | ✓ VERIFIED | Consumed by interpreter and native tests. |
| `internal/compiler/cgen/cgen.go` | Native emission integration | ✓ VERIFIED | Emitter paths wired to scalar selection. |
| `internal/compiler/cgen/cgen_program.go` | Safe C operation emission | ✓ VERIFIED | Unsigned operand and exact-width cast observed. |
| `internal/compiler/cgen/cgen_phase35_test.go` | C shift-site checks | ✓ VERIFIED | Positive site count and unsafe controls pass. |
| `testdata/phase35/shift-sites.json` | Expected emitted shift sites | ✓ VERIFIED | Scanner consumes fixture. |
| `internal/compiler/surfacemanifest/manifest.go` | Public operator/refusal contract | ✓ VERIFIED | Source-free manifest route reflects operators. |
| `cmd/schway/main_phase35_test.go` | Public run witness | ✓ VERIFIED | Interpreter/native outcome path passes. |
| `internal/compiler/session/session_phase35_values_test.go` | Value/optimization/UBSan lanes | ✓ VERIFIED | Fresh receipts include these current named tests. |
| `internal/compiler/session/session_phase35_surface_test.go` | Manifest/census tests | ✓ VERIFIED | Positive census and removal controls pass. |
| `testdata/phase35/surface-witnesses.json` | Independent surface fixtures | ✓ VERIFIED | Consumed by route and manifest tests. |
| `testdata/phase35/surface-census.json` | Independent admission census | ✓ VERIFIED | Census is non-vacuous and tested. |
| `examples/checksum/adler32.schway` | Adler shift consumer | ✓ VERIFIED | Source uses `b << 16`. |
| `docs/agent-surface.md` | Operator and shift guide | ✓ VERIFIED | Required semantics documented. |
| `internal/compiler/session/session_phase35_adler_test.go` | Adler regression tests | ✓ VERIFIED | Both engines and wrong-shift control pass. |
| `internal/compiler/session/session_phase35_closeout_test.go` | Evidence/closeout tests | ✓ VERIFIED | Evidence claims and docs checked. |
| `35-COMPATIBILITY-DELTA.md` | Scoped compatibility changes | ✓ VERIFIED | Baseline versus amendments explicit. |
| `35-EVIDENCE.md` | Evidence index | ✓ VERIFIED | Current receipt references traceable. |

### Key Link Verification

All 24 declared links were manually traced through source and named consumers. The generic lexical link query verified one directly and reported 23 “target not referenced” because it searches literal target path strings; those links use function/operation semantics instead. The manual traces below confirm the actual connections.

| From | To | Via | Status | Details |
|---|---|---|---|---|
| prechange session test | C emitter | frozen inventory to emitter outcomes | ✓ WIRED | Three emitter entry points are compared to the pinned corpus. |
| numeric baseline check test | checker scalar lowering | baseline to both checker routes | ✓ WIRED | Both source-conversion routes are explicitly exercised. |
| parser | formatter | lossless operator/group AST round trip | ✓ WIRED | Parse-format-parse idempotence and spelling tests pass. |
| lexer | checker | exact numeric token/span to route-local conversion | ✓ WIRED | Source positions and named overflow/refusal paths are asserted. |
| parser | refusal registry | depth counter to named refusal and recovery | ✓ WIRED | Depth failure and sibling recovery are exercised. |
| parser | checker scalar | number leaf to literal amount | ✓ WIRED | Group-preserving AST admits only a number leaf. |
| checker scalar | core | five operation kinds to transfers | ✓ WIRED | Producer and safety sink tests exercise each operation. |
| checker scalar | checker scalar | unused Byte formal bridge | ✓ WIRED | Byte is kept unused; uses/conversions refuse. |
| core | corevalidate | raw fields to independent shape/range laws | ✓ WIRED | Linear and CFG analyses reject malformed shapes. |
| core | originvalidate | raw fields to independent origin laws | ✓ WIRED | Origin recomputation and publish paths reach tests. |
| originvalidate | originvalidate | peer transfer to observed fact/safety obligation | ✓ WIRED | Actual analysis-trigger tests reach each fault. |
| corevalidate | interpreter | validated core to runtime values/events | ✓ WIRED | Literal value vectors pass with unchanged events. |
| core | pathoracle | op kinds to ancestry dispatch | ✓ WIRED | New operations flow to loan-free ancestry. |
| interpreter document | executionpeer | actual execution to causal classification | ✓ WIRED | No-event admission and fabricated events are checked; no numeric replay. |
| value vectors | interpreter tests | independent literals to comparator | ✓ WIRED | Vector file is consumed by the interpreter tests. |
| core | C emitter | validated immediate to actual C shift site | ✓ WIRED | Native output runs and static site scan sees nonzero sites. |
| C emitter | native runtime | pure kind to event projection/inventories | ✓ WIRED | No new event or optimizer fact is introduced. |
| core | C emitter | unused-Byte fixture signature | ✓ WIRED | Existing harness preserves Byte input and U64 result. |
| bitwise example | public CLI test | existing public run to engine outcomes | ✓ WIRED | CLI conformance test passes on interpreter and native. |
| value vectors | session tests | literal fixtures to Clang O0/O3/LTO/UBSan | ✓ WIRED | Fresh focused lanes and wrong-answer controls are in receipts. |
| manifest | session surface tests | census to manifest and drift controls | ✓ WIRED | Bidirectional removal/false-admission controls pass. |
| Adler source | Adler session tests | shift source to checksum and adler_low | ✓ WIRED | Existing imported outcomes remain byte-identical. |
| evidence index | closeout session tests | actual run references to closed coverage kinds | ✓ WIRED | Test-tier false-claim controls pass. |
| evidence index | language maturity doc | source facts/receipts to scoped next capabilities | ✓ WIRED | Current source inspection and named owners are reflected. |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|---|---|---|---|---|
| `examples/bitwise.schway` | `witness` | Parsed/checker-validated core from source locals; interpreted or emitted native code | Yes. Public test observes pinned `264` and native/interpreter outcomes. | ✓ FLOWING |
| `examples/checksum/adler32.schway` | `r`, `checksum`, `adler_low` | Imported source computation including `b << 16`; existing consumer routes | Yes. Existing literal outputs are compared across both engines. | ✓ FLOWING |

### Behavioral Spot-Checks

Focused named tests were run for this verification. The refreshed task/regression receipts were also checked against every input hash and raw log hash; they bind to source revision `f19721428ec02008b17c3dbbb13a4169f035ef89`. The coverage fingerprint includes the current mutation-test sources, Phase 33 debt test, Phase 34 recovery consumer/fixture, TAP adapter, and both negative-claim subjects named by these receipts. The current HEAD adds only phase gate/review documentation after that source revision.

```text
go test ./internal/compiler/syntax -run '^(TestPhase35ScalarGrammar|TestPhase35ExpressionDepth|TestPhase35NumericOwnership|TestPhase35FormatterRoundTrip)$' -count=1 -v -timeout=600s
go test ./internal/compiler/check -run '^(TestPhase35ShiftAdmission|TestPhase35NumericRoutePolicy|TestPhase35UnusedByteSourceAdmission)$' -count=1 -v -timeout=600s
go test ./internal/compiler/corevalidate -run '^(TestPhase35ShiftImmediateAdmission|TestPhase35ScalarTransferMutants|TestPhase35ShiftTargetFreshness)$' -count=1 -v -timeout=600s
go test ./internal/compiler/originvalidate -run '^(TestPhase35ShiftImmediateAdmission|TestPhase35ScalarTransferMutants|TestPhase35AnalysisTriggers)$' -count=1 -v -timeout=600s
go test ./internal/compiler/interp -run '^(TestPhase35BitwiseShiftVectorsEventFree|TestPhase35InterpreterWrongAnswerControls)$' -count=1 -v -timeout=600s
go test ./internal/compiler/cgen -run '^(TestPhase35NativeShiftTracer|TestPhase35ShiftSiteControls|TestPhase35UnsignedShiftEmissionAndEventProjection)$' -count=1 -v -timeout=600s
go test ./internal/compiler/session -run '^(TestPhase35FocusedValueAnswers|TestPhase35ExistingAdlerAnswers|TestPhase35AdlerWrongShiftControl)$' -count=1 -v -timeout=600s
go test ./internal/compiler/executionpeer -run '^TestPhase35PureOperationCausalControls$' -count=1 -v -timeout=600s
```

| Behavior | Command | Result | Status |
|---|---|---|---|
| Parser grouping, depth, numeric ownership, formatter | Focused syntax command above | Exit 0 | ✓ PASS |
| Literal-only checker admission and Byte bridge | Focused checker command above | Exit 0 | ✓ PASS |
| Core/origin shape, transfers, target freshness | Focused validator commands above | Exit 0 | ✓ PASS |
| Interpreter vectors and wrong-answer controls | Focused interpreter command above | Exit 0 | ✓ PASS |
| C emission and positive shift-site scanner | Focused C generator command above | Exit 0 | ✓ PASS |
| Public values and Adler regression | Focused session command above | Exit 0 | ✓ PASS |
| Causal event-free admission and fabricated-event refusal | Focused executionpeer command above | Exit 0 | ✓ PASS |
| Fresh phase task selectors | `35-REFRESH2-VALIDATION.json` | 66 distinct names, 72 actual roots, 13 packages; 0 missing/fail/skip; all 122 input hashes and 13 raw-log hashes matched | ✓ PASS |
| Fresh current plus carried regression selectors | `35-REFRESH2-RUN.json` | 184 distinct names, 196 actual roots, 16 packages; 0 missing/fail/skip; all 177 input hashes and 16 raw-log hashes matched | ✓ PASS |
| Actual-source mutation controls | `35-REFRESH2-RUN.json` | 15 transfer and 26 admission mutants killed | ✓ PASS |
| Optimization and shift sanitizer lanes | Fresh raw logs referenced by `35-REFRESH2-RUN.json` | Literal answer matrix at O0, O3, O3-LTO and shift UBSan rows passed | ✓ PASS |

The receipts establish fresh local evidence, not a full workspace suite or hosted multi-host run. Neither is a Phase 35 success criterion.

### Prohibition Enforcement

The two test-tier descriptors in `35-08-PLAN.md` project exactly into `35-REFRESH2-PROHIBITION.json`; their six recorded input hashes matched. Both canonical `check prohibition-enforcement` queries returned green. The actual causal-false-claim and skipped-lane-claim subjects returned red, while clean and default subjects returned green. No human waiver is inferred.

| Prohibition | Enforcement | Result |
|---|---|---|
| Do not treat engine agreement or causal admission as numeric proof | Node TAP adapter invokes the Go evidence-grounding test with a bad causal subject and clean control | ✓ VERIFIED; false claim rejected, clean/default accepted |
| Do not describe historical or skipped lanes as fresh evidence | Same adapter checks the false lane claim and clean control | ✓ VERIFIED; false claim rejected, clean/default accepted |

### Probe Execution

No phase plan or success criterion declares a `probe-*.sh`; this is a language/compiler feature phase, not a migration or tooling phase. No probe was applicable.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|---|---|---|---|---|
| BIT-01 | 35-03 through 35-07 | U64 bitwise operators with independent value and causal evidence | ✓ SATISFIED | Truths 1, 12, 15, 17, 20, 22, 28-31. |
| BIT-02 | 35-02 through 35-07 | Literal shifts limited to 0..63 with safe C semantics | ✓ SATISFIED | Truths 2, 10, 13, 16, 24, 29, 33. |
| BIT-03 | 35-02 and 35-07 | Hex literal support, diagnostics, and spelling preservation | ✓ SATISFIED | Truths 3, 9, 11, 34. |
| BIT-04 | 35-08 | Adler shift with unchanged outputs on both engines | ✓ SATISFIED | Truths 4 and 35. |

No additional requirement is mapped to Phase 35 without a plan owner. DX-22 remains owned by Phase 34; its dynamic-shift refusal slice is explicitly assigned to BIT-02 and verified here.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---|---|---|---|
| Modified compiler files | Various | Broad grep candidates such as empty diagnostic slices, optional empty inventories, and a local “placeholder” fixture comment | ℹ️ INFO | Context inspection showed existing sentinel/helper behavior, not a Phase 35 value path or placeholder implementation. No phase output depends on a static empty fallback. |
| All files named by phase plans | — | Unreferenced `TBD`, `FIXME`, or `XXX` debt markers | None found | Debt-marker gate passes. |

### Decision Coverage

`check.decision-coverage-verify` reports 23/23 trackable CONTEXT decisions honored by shipped artifacts. No decision is unresolved or silently overridden.

### Human Verification Required

None. Runtime behavior that determines the goal has named passing tests and fresh receipt evidence; no deferred human checks were found in task plans.

### Gaps Summary

No Phase 35 goal gaps remain. The source implements and connects all required scalar syntax, core operations, independent validators, interpreter semantics, native emission, manifest/refusal rules, and Adler consumer behavior. The 26-file Phase 35 review found zero new findings, and the recorded security review passed its 24 applicable controls with no open Phase 35 finding. The fresh receipts bind to the production-source revision and raw logs rather than relying on SUMMARY claims.

Separate carried debt remains visible: Phase 33 CR-03 and CR-04 are confirmed deferred BLOCKER false-rejection cases, owned by the execution-peer maintainer for the mandatory M007 audit before the Phase 38 freeze. The current review found they do not falsify Phase 35 bitwise, shift, hex, or Adler witnesses; they are not fixed, accepted risk, or a global-clean claim. Phase 36 WR-02 is a separate deferred comment warning. These items do not create a Phase 35 gap.

---

_Verified: 2026-10-16T01:20:31Z_
_Verifier: the agent (gsd-verifier)_
