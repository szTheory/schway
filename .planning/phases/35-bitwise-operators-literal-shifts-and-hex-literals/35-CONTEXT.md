# Phase 35: Bitwise Operators, Literal Shifts, and Hex Literals - Context

**Gathered:** 2026-10-13
**Status:** Ready for planning
**Decision authority:** The user requested all relevant gray areas, specialist fan-out, primary-source research and adversarial passes to diminishing returns, then explicitly delegated following the coherent recommendations. Three typed advisor researchers completed initial and adversarial passes. Selections below are agent recommendations under that delegation, not invented interactive user answers. Persistent `workflow.auto_advance` governs the planning/execution handoff.

<domain>
## Phase Boundary

Deliver BIT-01..BIT-04: U64 `^`, `&`, `|`, literal-only `<<` and `>>`, complete the already partially admitted hex-literal contract, and replace Adler's 16 doublings with `b << 16`. Preserve the written runnable expression `x ^ (y << 3) & 0xFF` by adding bounded scalar grouping. No scalar helper calls, general constant expressions, complement, rotations, new numeric types, backend, dependency, runtime event or defect path.

Phase 36 owns scalar calls; Phase 37 owns the CRC file consumer, second consumer, interpreter file CLI and guide; Phase 38 owns final hosted receipts. DX-22's dynamic-shift refusal is completed here, without moving its requirement owner from Phase 34.

</domain>

<decisions>
## Implementation Decisions

### Scope and compatibility

- **D-01:** Use Go 1.24 standard-library/local implementation and the existing C17/Clang backend. No expression parser library, bigint package, new proof/replay framework or dependency. Keep legacy core/execution/diagnostic schema identifiers and current exit classes; new op kinds reuse existing fields rather than adding a mandatory wire field.
- **D-02:** Before compiler changes, commit an enumerated literal generated-C baseline for the current tracked source corpus through the relevant existing emitter routes. Record non-emitting cases explicitly. Keep all old goldens immutable and arm a reached byte-drift control. Programs without the new operators, grouping or hex extension retain their emitted C bytes; Adler's deliberately changed source is an explicit separate exception with unchanged behavioral answers. Pin previously admitted scalar/linear numeric policies and Bool identifier uses before conversion/parser changes.

### Expression grammar and bounded grouping

- **D-03:** Use descending precedence `% > + > << >> > & > ^ > | > (<, ==)`, left association within each tier. Preserve the existing shared comparison tier and relative order of old operators. U64-only bitwise operators return U64; Bool operands remain refused. Evaluate ordinary operands in the existing left-to-right lowering order. Keep `&&`, `||`, `!` and subtraction refused; admitting `|` must not split `||` into admitted operations.
- **D-04:** Add transparent scalar parentheses for the exact roadmap witness. Parentheses preserve the inner expression kind rather than creating a runtime/core operation: `(3)` and nested groups around a literal are literal amounts; `(1 + 2)` and names remain non-literals even if constant propagation knows their value. Preserve lossless CST punctuation and source spans. Do not infer that grouping admits calls, tuples, Bool literal tails or other absent forms.
- **D-05:** Bound recursive scalar expression nesting at 64, with a distinct named expression-depth resource refusal and `parser.expression_depth` manifest entry. Cover grouping and recursively nested builtin indices; use iterative delimiter-aware recovery and preserve sibling diagnostics. Reuse a numeric limit value if appropriate, but do not label expression depth as type depth. Remove or bound the existing unsupported-`!` recursive path when touching the atom parser. This is an explicit parser resource-bound amendment, not a claim that source/token caps alone prevent stack exhaustion. Characterize old nested-index admission before implementation and record any resource-refusal delta separately; do not silently rebaseline old successful C witnesses.
- **D-06:** Preserve nested type closing tokens, comparisons, ownership keywords, declaration/match `|` separators, comments and strings. Planner may choose context-local adjacent-angle recognition or dedicated shift tokens with contextual type-close splitting; require contiguous source bytes for `<<`/`>>` and a nested-generic parse/format golden. Formatter output must reparse to the same admitted tree, be idempotent, keep numeric spelling, and place infix operators coherently without changing old formatting goldens.
- **D-07:** Under D-03 the written witness means `x ^ ((y << 3) & 0xFF)`. Pin `x=0x100`, `y=0x21` to literal decimal answer `264`; `(x ^ (y << 3)) & 0xFF` gives `8` and is a discriminating control. Document both rather than silently treating the original expression as a final mask. **2026-10-14 research clarification:** declarations require one parameter; public run binds Byte/Buffer fixtures but not U64, while scalar checking currently accepts only U64/ByteView inputs. To preserve the published public runnable gain, admit the narrow bridge `fn main(unused: Byte) -> U64` with a provably unused Byte parameter and U64/Bool local computation. Keep one formal parameter, existing fixture input, arity and all U64 input-required behavior; no new CLI flag or default U64 input. Byte is not a U64 operand or convertible scalar; source check and independent core/origin laws must refuse any use of that placeholder. This is a named witness-enabling admission obligation, not general Byte scalar support. Both public conformance routes return outcome value `264`; they retain their existing execution-document/human projection rather than becoming bare-stdout application routes. Separately pin `264\n` from the existing U64 application/native runner and interpreter application seam. Record the bridge's actual tests before claiming it runnable.

### U64 semantics, shift representation and refusals

- **D-08:** XOR/AND/OR are ordinary per-bit operations on exactly 64-bit unsigned values. `x << n` keeps the low 64 bits of `x * 2^n`; `x >> n` is logical right shift. There is no overflow defect, signed shift, masked count, target-dependent word width or dynamic amount. Counts 0, 31, 32 and 63 have independently authored expected vectors; include maximum/high-bit/all-zero values and differing operators/associations, not only engine agreement.
- **D-09:** Lower shifts to distinct immediate-only core kinds, recommended `OpShiftLeftU64` and `OpShiftRightU64`, using the existing `ConstU64` string for the canonical base-10 amount. `RightID` must be empty. Permit `ConstU64` only on `OpConst` and these shift kinds; require presence, canonical decimal spelling, range 0..63, U64 source/result, valid initialized source and fresh/valid target facts. No named amount place, constant-def lookup or producer-issued literal witness may substitute for this contract. Old core serialization/IDs stay unchanged when no new feature appears.
- **D-10:** Source checking independently requires a number leaf after transparent grouping; refuse names, computed constants, unsupported negative counts and representable 64+ counts with `surface.unsupported_shift_amount` and the one-line hint: `Shift amounts must be integer literals from 0 to 63; write a literal in that range.` Decimal and admitted hex amount spellings canonicalize only in core. Malformed numeric tokens and unrepresentable U64 values keep their existing named lexical/type failures, with an explicit code/span delta ledger if a targeted hex diagnostic is refined. Do not admit unary minus or subtraction to diagnose shifts.
- **D-11:** Corevalidate and originvalidate each independently enforce the immediate shape, U64 types and range, including missing/noncanonical/oversized amounts, extraneous `RightID`, wrong kind fields and malformed raw-core producer input. Their shift contract refusal uses the registered code/hint where diagnostics are available. Removing each refusal must be reached and caught in a table-driven control. Neither peer has source AST provenance: they prove valid IR immediate semantics, not that arbitrary raw core was generated from a particular literal token. Source syntax admission belongs to check; a new source-provenance protocol is out of scope.
- **D-12:** Generated shifts must contain a checked literal amount below 64. Use an unsigned non-promoted left type, e.g. `(uint64_t)((unsigned long long)lhs << 16)` and analogous right shift, so C integer promotion cannot introduce signed-shift UB. `uint64_t` is exact-width when available; unsigned long long is at least 64 bits, and the final cast yields the required low 64 bits even on a wider implementation. No unchecked fallback, compiler assumption or new optimizer attribute. A finite generated-C corpus scan must identify/count the actual new shift sites and fail on a dynamic or 64+ shift mutant; do not pass merely because no shift was found.

### Hex spelling and existing numeric policies

- **D-13:** Hex is partially implemented already: lexer accepts `0x`/`0X`, mixed-case ASCII digits, and existing between-digit underscores; generic number AST/raw token and formatter preserve spelling. Complete uppercase `0X` conversion on the legacy linear path with the smallest local change. Keep leading zeros bounded by existing source/token limits and overflow by value, not digit count. No new hex token, suffix, alternate radix, separator placement, Unicode digit normalization or automatic uppercase/underscore rewrite.
- **D-14:** Preserve existing decimal/binary policies rather than normalizing them as collateral work. Legacy `parseU64Literal` treats `010` as decimal 10 and recognizes lowercase binary prefix; scalar base-0/fallback conversion treats `010` as 8 and accepts `0B10` as 2, while legacy rejects `0B10`. Record and test these route differences as deferred consistency debt. A shared converter is allowed only if it preserves explicit route policies and earns its complexity; the preferred patch is narrow `0X` support. The no-feature C contract forbids changing `010` now.
- **D-15:** Malformed numeric candidates retain `syntax.malformed_numeric_literal` and the complete raw candidate span (`0x`, invalid radix digits, trailing ASCII/Unicode suffix, bad separators). Use exact token/span ownership to avoid a second parser expectation for that same cause; retain other causes and limit diagnostics. Valid-but-overflowing hex keeps a named existing type/range code: legacy `check.literal_out_of_range`, scalar `type.invalid_u64` unless a targeted documented hex-only refinement is needed. Do not silently promise a uniform historical code. Pin code/count/span and spelling-preserving format/Tree.Bytes() for accepted and rejected inputs, including EOF, maximum U64, maximum+1 and long leading zeros.

### Independent evidence and manifest

- **D-16:** Add independently implemented transfer cases in check, corevalidate and originvalidate, including their actual scalar fixed-point trigger/inventory paths. Keep each peer's type/shape/range admission independent of the producer and of a shared semantics evaluator. For exact-result transfer mutants, feed discriminating literal results into a downstream observable analysis obligation (e.g. statically zero remainder) or independently test that peer's transfer result; do not claim an isolated returned value was checked if the peer only analyzes safety. Each table row names the peer, op, seeded fault, reached branch and named failure; restore every mutable control.
- **D-17:** Keep pure operations event-free in interpreter, C evidence projection, causal peer and capacity inventories, with no new runtime reason or event. `executionpeer.Validate` explicitly checks observed causal structure and receives no entry argument; it cannot prove the numeric result of a pure operation. Independently test per-op causal admission/no-event behavior and reject fabricated step/defect events at those operations. Retain literal expected-answer and wrong-result controls in interpreter/native, at `-O0`, `-O3` and `-O3 -flto`, with a focused UBSan shift lane. New value replay or operand events would exceed this slice and defeat its event-free contract.
- **D-18:** **Evidence-backed acceptance clarification:** BIT-01/roadmap criterion 1 currently require a wrong-result mutant for every op at every peer, including the causal-only execution peer. Before writing plans, amend the live Phase 35 criterion and BIT-01 wording with a dated, source-grounded note: value-transfer mutants at check/core/origin and engine answer mutants; per-op event-free/forged-event controls at executionpeer. Preserve the original demand in the dated amendment/provenance, name this limitation in verification, and do not count causal admission as numeric validation. This recommendation is auto-followed under the user's explicit decision delegation; it preserves the goal of defined, independently tested answers without introducing runtime events. No other requirement or phase owner changes.
- **D-19:** Extend the existing deterministic `schway.surface/1` manifest with all five operators, grouping, precedence/association, literal-only count/range, low-bit left-shift rule, hex spelling and new expression-depth bound. Add the implemented shift refusal to its catalog. Keep ten top-level fields/source-free discovery and independent literal complete-source witness matrix plus finite source/admission/limit census; no producer-fed expected manifest or weakened discovery tests. Arm removal, false-admission, missing refusal/hint and weakened restriction controls, with restoration. Keep scalar expression calls absent until Phase 36.

### Runnable consumer, lanes and phase cost

- **D-20:** Replace Adler's 16 `r = r + r` operations with the equivalent `b << 16`, retaining the existing final addition and module/cleanup contract. The maximum bounded `b` makes this equivalent to the old checked additions. Existing Phase 29/31 literal answer, checksum and `adler_low` output tests must pass unchanged, on both engines. Do not update expected answers to the new producer or require file input through the public interpreter CLI before Phase 37. Include a reached wrong-shift mutant in a focused additional test.
- **D-21:** Only focused named local selectors/narrow CLI checks with `GOCACHE=/tmp/ai-lang-verification-gocache`; no local project-wide suite, whole package runs, broad native matrix, push or hosted dispatch/binding. Plan the small optimization/literal-shift matrix as named feature tests, reusing existing runners. Existing recurring CI owns broad checks; Phase 38 owns hosted two-host final codedigest receipts. State local host/toolchain and incomplete host lanes honestly; human readability and cold/warm whole-loop latency remain unmeasured.
- **D-22:** Aim for about eight coherent plans or fewer. Baseline before compiler edits, then grammar/hex and core contracts, independently derived validators/execution, focused evidence/manifest, consumer/guide/closeout. Additional plans require a named runnable witness or safety obligation; do not split every role or op into a plan. SUMMARY coverage names real tests and the closed kind vocabulary; prioritize deterministic automation over human UAT.
- **D-23:** Retain dormant seeds z1g/trz/z1s with the scoped transition disposition below. Do not promote a database, telemetry framework or syntax study from generic forecasts. Update living documents only with newly inspected facts and preserve archived/differently dated claims; the three useful next capabilities remain Phase 35 computation, Phase 36 imported helpers and Phase 37 file CRC/second consumer/CLI parity.

### Planner's Discretion

- Exact local helper, op-name and test-file layout, consistent with the above contracts.
- Shift-token/type-close strategy and deterministic formatter spacing once pinned.
- Named expression-depth code/precise boundary counting, with 64-level bound and reachable malformed controls.
- Concrete finite baseline corpus, per-peer observability sinks and focused matrix arrangement; preserve historical receipts and goldens.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Ownership and prior contracts

- `AGENTS.md` — GSD workflow and living-transition rules.
- `.planning/ROADMAP.md` Phase 35 and successors; `.planning/REQUIREMENTS.md` BIT-01..04/DX-22 ownership; `.planning/PROJECT.md` verification/dependency preferences; `.planning/STATE.md` local lane restrictions.
- `.planning/research/M007/SUMMARY.md` — scoped consumer/evidence choice; its hex/terminator-only appraisal is corrected by this source inspection.
- `.planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-CONTEXT.md` — frozen identity, exact-span recovery ownership, manifest independence and dynamic-shift ownership.
- `.planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-REVIEW-FIX.md` — actual limit route/scope consumers and finite census limitations.
- `.planning/phases/33-carried-debt-bounded-capture-read-and-generic-freeze-gate/33-CONTEXT.md` — bounded capture, defect reason ownership and receipt lanes.
- `.planning/milestones/M006-phases/32.1-address-tech-debt-cli-exit-classes-and-module-aware-routes/32.1-CONTEXT.md` — entry/route/exit restrictions, with later debt amendments preserved.

### Source and evidence witnesses

- `internal/compiler/syntax/parser.go`, `internal/compiler/syntax/lexer.go`, `internal/compiler/syntax/token.go`, `internal/compiler/syntax/tree.go`, `internal/compiler/syntax/format.go` — lossless tokens, Pratt tiers, recursive atoms, type closes, recovery and formatting.
- `internal/compiler/check/check.go`, `internal/compiler/check/scalar.go`, `internal/compiler/core/core.go` — numeric route policies, source literal admission, canonical core and abstract transfer.
- `internal/compiler/corevalidate/corevalidate.go`, `internal/compiler/originvalidate/originvalidate.go`, `internal/compiler/executionpeer/executionpeer.go`, `internal/compiler/execution/execution.go` — independent scalar laws and causal-only evidence boundary.
- `internal/compiler/interp/interp.go`, `internal/compiler/cgen/cgen.go`, `internal/compiler/cgen/cgen_program.go`, `internal/compiler/native/native.go` — engines, emitter validation gate and event/capacity inventories.
- `internal/compiler/surface/refusals.go`, `internal/compiler/surfacemanifest/manifest.go`, `testdata/phase34/surface-witnesses.json` — current installed facts and independently authored witnesses.
- `internal/compiler/session/session_phase34_surface_test.go`, `internal/compiler/session/session_phase34_limit_consumers_test.go`, `internal/compiler/session/session_phase34_prechange_test.go` — bidirectional finite census, consumer limits and immutable C precedent.
- `internal/compiler/syntax/parser_phase19_test.go`, `internal/compiler/check/check_phase19_test.go`, `internal/compiler/interp/interp_phase19_integration_test.go` — existing numeric spelling/conversion tests; researcher resolves exact named cases from source.
- `examples/checksum/adler32.schway`, `examples/checksum/checksum.schway`, `examples/checksum/adler_low.schway`, `examples/checksum/README.md` — actual ordinary-source consumer.
- `internal/compiler/session/session_phase29_interp_fixture_test.go`, `internal/compiler/session/session_phase29_native_fixture_test.go`, `internal/compiler/session/session_phase31_interp_test.go`, `internal/compiler/native/native_phase31_modules_test.go`, `internal/compiler/native/native_phase31_matrix_test.go` — existing engine/consumer answer and output tests; researcher resolves exact focused selectors before plans cite them.

### Living records and primary references

- `.planning/PRODUCT-ROADMAP.md`, `.planning/LANGUAGE-MATURITY.md`, `.planning/sources/DASHBIT-AI-ERA.md`.
- `.planning/seeds/SEED-261008-z1g-bounded-symbol-and-caller-query.md`, `.planning/seeds/SEED-261008-trz-consumer-triggered-failure-observability.md`, `.planning/seeds/SEED-261008-z1s-compare-contract-and-source-forms.md`.
- [Rust expression precedence](https://doc.rust-lang.org/reference/expressions.html#expression-precedence) — contrasting grammar, not wholesale adoption; existing Schway comparisons remain left-associative.
- [Go integer literals](https://go.dev/ref/spec#Integer_literals) and [Go scanner](https://go.dev/src/go/scanner/scanner.go) — contrasting separator/prefix policies, not Schway's numeric spec.
- [Rust integer tokens](https://doc.rust-lang.org/reference/tokens.html#integer-literals) — suffix/prefix alternatives deliberately not added.
- [WG14 N2176 C17 draft](https://www.open-std.org/jtc1/sc22/wg14/www/docs/n2176.pdf) §6.5.7 — count, promoted-width and unsigned-shift law; advisors consulted it, parent web parser could not decode this PDF.
- [Clang UBSan](https://clang.llvm.org/docs/UndefinedBehaviorSanitizer.html#available-checks) — targeted shift sanitizer lane, not proof of all semantics.

</canonical_refs>

<code_context>
## Existing Code Insights

No codebase maps, phase SPEC/context/plans/checkpoint or advisor profile were present. Scout used relevant parser, lexer, checker/core, manifest, formatting and causal-peer source. No Go/native behavior check ran during this discussion.

### Reusable Assets and Patterns

- Existing Pratt parser and token-based formatter can preserve old precedence/bytes. Type angle tokens and `||` refusal require contextual care.
- Generic `TokenNumber.Text`, lossless `Tree.Source`/`Bytes()`, existing numeric validation and named diagnostics already supply much of BIT-03.
- `ConstU64` is already an omitempty decimal core value field; kind exclusivity must be explicitly extended for immediate shifts.
- Three independent scalar analyses can observe exact constant transfers, but `originHasScalarCFG` and explicit admission laws must cover the new kinds. Causal event checking does not supply numeric replay.
- Existing baseline, literal answer, mutation, bounded source-census and event-free tests provide narrow evidence patterns. Preserve their authority distinctions.

### Integration Points

Parser/token/format → scalar checker/core → independent core/origin laws → interp/cgen/native → session/public run; manifest is a leaf consumer. Imported Adler exercises the existing module route, without widening module-aware commands or scalar calls.

</code_context>

<specifics>
## Runnable Witness and Transition Review

Pin `x=0x100`, `y=0x21`, `x ^ (y << 3) & 0xFF` to outcome value `264` from both public conformance engine routes using the unused-Byte bridge in D-07. Pin bare `264\n` separately on the existing application/native runner and interpreter application seam; public conformance streams are not application stdout. Pin grouping/final-mask alternative to value `8`. Numeric output stays decimal; hex output is not added.

| Capability/program | Current blocker | Smallest complete slice | Guarantee/evidence | Owner/next action | Observation changing priority |
|---|---|---|---|---|---|
| Bitwise main and Adler `b << 16` | New operators, grouping and literal-shift core contracts absent; hex conversion incomplete | This bounded Phase 35 computation slice | BIT-01..04, shift shape/range, literal vectors, per-peer transfers/causality, immutable no-feature C | Plan/execute Phase 35 | A concrete consumer requires dynamic shifts or another bound |
| Imported `crc_step` helper in a loop | Scalar expression calls remain absent | Arity-one acyclic imported helper returning U64/Bool | CALL-01/02, independent call/argument/result and callee-defect controls | Phase 36 | A named second consumer needs arity-N or a changed call boundary |
| File `crc32` plus second importer and CLI parity | Operator/helper prerequisites and public interpreter file route | Existing bounded file/cleanup path with ordinary CRC module and executable guide | APP-13/14, MOD-07, DX-24; independent ISO-HDLC vectors and outcome parity | Phase 37; final hosted EVD-17 in Phase 38 | A real file tool requires a smaller API or different outcome |

| Seed | Disposition | Current source witness | Named consumer and trigger | Next action |
|---|---|---|---|---|
| z1g | retain dormant | `internal/compiler/session/session_phase6_query.go` and ordinary source search | `crc_step` navigation; no measured query accuracy/cost gap established | Consumer owner names/measures a gap before promotion |
| trz | retain dormant | `internal/compiler/native/native_app.go#readApplicationCapture`, `internal/compiler/session/session_app_verify.go` | Adler/CRC failure; no unexplained failure established; old allocation issue resolved in Phase 33 | Review one actual unexplained failure before expanding observation |
| z1s | retain dormant | `internal/compiler/syntax/format.go#Format` | Bitwise expression review; a bounded grammar decision is resolved here, no actual reviewer/source-form dispute established | Require named dispute and independent evidence before experiment |

This is source inspection and primary-source research, not newly executed compiler evidence. Prior Phase 33/34 checks and M006 receipts remain historical at their named revisions. Human readability and whole-loop cold/warm latency are unmeasured. No full suite/native matrix/hosted receipt was run here.

</specifics>

<deferred>
## Deferred Ideas

- Numeric route unification (`010`, `0B10`) requires a separately owned semantic compatibility decision; preserve the discrepancy as named debt now.
- Scalar result replay within executionpeer needs runtime inputs/CFG/value evidence and is not part of this event-free slice.
- Dynamic shifts, complement, rotations, numeric suffixes/new bases/types, hex output, arity-N, aggregates, writable buffers and additional backend remain outside Phase 35.
- Scalar helpers Phase 36; CRC consumer/second importer/interpreter file CLI/guide Phase 37; final two-host hosted binding Phase 38.
- Word counter/longest_line remain future M008, followed by a named bounded config parser/JSON consumer.

</deferred>

---
*Phase: 35-bitwise-operators-literal-shifts-and-hex-literals*
*Context gathered: 2026-10-13*
