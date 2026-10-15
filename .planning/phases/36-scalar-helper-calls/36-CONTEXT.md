# Phase 36: Scalar Helper Calls - Context

**Gathered:** 2026-10-14
**Status:** Ready for planning
**Decision authority:** The user requested all relevant decisions, specialist fan-out, reputable primary sources, adversarial review to diminishing returns, and automatic adoption of a coherent recommendation. Three typed advisor researchers completed read-only research. Decisions below are delegated recommendations, not fabricated interactive answers. Persistent `workflow.auto_advance: true` governs the next planning/execution handoff.

<domain>
## Phase Boundary

Deliver CALL-01/CALL-02: direct, arity-one, acyclic scalar value calls with U64/Bool arguments and results, including imported local-module helpers, used in scalar expressions and assignments. Prove a `crc_step`-style helper called repeatedly from a loop and through an importer. Preserve existing defect reasons, exit classes, source authority, module resolution, and no-new-call generated C bytes.

This explicitly includes adding Bool scalar signatures and initialized Bool identifier tails, which are absent at the baseline. It does not admit Bool literal tails, arity-N, ByteView scalar-to-scalar arguments, indirect dispatch, recursion, closures, aggregate returns, a new backend, a new dependency, or a new proof/replay framework. Existing ByteView callers may invoke U64/Bool helpers; existing owning-wrapper-to-ByteView calls remain supported. Phase 37 owns the complete file CRC program, second CRC importer, public interpreter file CLI, and executable guide. Phase 38 owns final two-host binding after the audit.

</domain>

<decisions>
## Implementation Decisions

### Bounded composition and source contract

- **D-01:** Use Go 1.24 standard library and small local code over dependencies. Reuse direct `OpCall`, `core.Function.Parameter`, existing module composition, interpreter frames, and readable C17/Clang. Preserve current core/execution/diagnostic schema names and wire fields. Sharing mechanical identity formatting is acceptable; sharing a producer's semantic verdict between independent peers is not.
- **D-02:** Admit all four signature pairs U64→U64, U64→Bool, Bool→U64, Bool→Bool for scalar helpers. Require exactly one parameter, exact argument type, and target type from the declared return type. No implicit Bool/U64 conversion, numeric truthiness, Unit value result, Byte operand conversion, or new resource/pointer/ByteView value-call signature. Preserve existing U64/ByteView/unused-Byte and Unit-body routes separately. The unused Byte bridge stays provably unused even in a call argument.
- **D-03:** Add a direct named-call scalar expression whose single argument is an already admitted scalar expression, including computed and nested calls. Reuse the existing 64-level expression-depth accounting and bounded delimiter-aware recovery; count call-argument recursion in that bound. Calls bind as atoms tighter than infix operators. Keep legacy non-scalar call grammar and ordinary arity-one refusals intact. Name-only arguments were considered but rejected here because an arbitrary special restriction would impede composition while the bounded scalar parser and ordered lowering already exist.
- **D-04:** Calls may appear wherever scalar expressions are already admitted, including binding/assignment RHS, operands, conditions, byte-index expressions, and another helper's argument, subject to the existing context's type rules. Preserve the current identifier-tail source form: a Bool helper returns an initialized Bool local or formal by identifier; bind a call result before returning it. No general expression-tail expansion or Bool literals follow. Keep `true`/`false` contextual identifiers and existing unsupported forms/hints.
- **D-05:** Fully evaluate the argument once before helper entry. Lower nested calls and sibling operands left-to-right into explicit fresh typed places and ordered `OpCall` operations; complete the call before evaluating the next operand or storing its assignment result. Do not emit nested C expressions that rely on C operand order. An argument defect prevents helper entry, later operands, and the store. Preserve existing admitted scalar writes and defect behavior; do not invent a purity annotation or silently duplicate effects.

### Calls, imports, and static guarantees

- **D-06:** Resolve local and imported helper targets through the existing closed whole-program module resolver, explicit exports, stable semantic IDs, and signature lookup. Reuse in-root/case/symlink/import-cycle limits unchanged. Test colliding names, declaration order, and a private/unimported target. No package manager, module search path, separate compilation, or newly module-aware query/interface route.
- **D-07:** Preserve the existing whole-program call-graph policy and refusal code: direct and mutual recursion are refused even in unused functions or unreachable blocks. Keep existing graph/depth/resource bounds, including the 4096 static invocation-path bound and interpreter depth limit 128; do not equate acyclicity with bounded dynamic work or allocate an unbounded unfolded DAG before refusal.
- **D-08:** Check, corevalidate, and originvalidate independently enforce callee existence/signature, initialized compatible argument, scalar-safe call context, correct result type, and a valid fresh call-result target. A call initializes its target to a value-unknown scalar of the declared result type. Do not import producer summaries or implement mandatory exact-return constant propagation. Preserve local constant/zero-divisor analysis; a helper result that happens to be zero follows the dynamic remainder-defect route unless independently proved by an existing law.
- **D-09:** Keep the distinction between a fresh static call-result temporary and assignment to a mutable local. Repeated execution of that static producer in a loop is legitimate; reusing a prior static producer's target in malformed raw IR is not. Include prior constant/copy/store/call targets, wrong fields/types, missing arguments/results, and positive branch/loop assignment shapes in direct independent validator controls. Check success must not mask a raw-peer refusal.
- **D-10:** Preserve all existing resource, loan, provenance, back-edge and path-oracle restrictions. A scalar value call does not authorize ownership/loan/event state on loop back edges or widen by-reference contracts. Update actual scalar-kind/CFG trigger inventories and explicit event allowances needed for this feature, rather than blanket allowlisting `OpCall` everywhere. Preserve ByteView argument ownership through the old owning-wrapper seam only.

### Dynamic invocation and bounded evidence

- **D-11:** Give repeated calls at the same static site distinct activations using the existing `InvocationSegment.Ordinal`: zero-based, contiguous per `(parent dynamic invocation, static OpCall ID)`. A fresh parent activation resets its child-site counters; separate sites and nested parents remain distinct. First/nonrepeated activations retain ordinal zero and old canonical spelling. Reusing the same invocation path with only event occurrences was considered but rejected because the existing public invocation grammar explicitly names caller-owned call-site occurrences. No new wire field/schema is needed.
- **D-12:** Interpreter, C emission, and executionpeer independently construct/check ordinal progression and ancestry. Keep event `Occurrence` for repetitions inside one activation; it does not replace activation identity. Reject missing/skipped/duplicate/reordered ordinals, wrong parent/site/callee, repeated terminal events, missing/unbalanced calls/returns, forged defects, and events after the terminal defect. Test one site called more than once, two sites targeting one callee, nested helper calls, a helper with its own loop, and a later-iteration defect.
- **D-13:** Retain bounded static preflight and existing execution-document (16 MiB), application-capture (64 KiB), runtime event/output, and call-depth limits. Dynamic counter/path construction must be bounded and checked before wrap or allocation. The planner may choose a small local bounded representation; it must distinguish static source refusal from runtime evidence exhaustion. Do not silently turn the 4096 static-node ceiling into a new semantic input/call-count limit. No global tracing heap or resident runtime.
- **D-14:** Native capacity reservation must account for calls, callee return/failure/defect records, invocation text, and repeated loop activations. The current sum of static operations is not a complete dynamic bound. Require a finite repeated-call witness whose captured trace is complete and independent exact boundary/exhaustion controls. Preserve named conformance capacity exit 69 and application capture's incomplete/capacity-exhausted status, separate from the application's own streams/outcome. Optional evidence exhaustion must not invent success or silently suppress a program defect. Do not promise complete per-call evidence for arbitrary Phase 37 CRC file sizes; Phase 37 must apply the existing ordinary-run versus capture/verification distinction.
- **D-15:** Executionpeer proves observed causal call structure, argument/result place/type attribution where represented, invocation ancestry, event order and terminal defect attribution. It has no scalar argument/result payload and cannot prove arithmetic or distinguish a plausible success trace from a numerically required defect. Keep Phase 35 D-18's boundary; no numeric replay, new operand events, interprocedural evaluator, or identical-mutant-at-every-peer requirement.

### Defects and acceptance clarification

- **D-16:** New U64/Bool helpers propagate their actual callee overflow/remainder reason through nested calls and loops on interpreter and native, exit 65, with no normal caller continuation or fabricated result. Exercise dynamic inputs so static zero/overflow refusal cannot substitute for execution. Pin exact reason/stdout/stderr/exit through the relevant application seam and capture attribution separately. No new failure kind or optimizer attribute.
- **D-17:** A U64/Bool-only helper cannot read bytes: `byte_at` requires the function's ByteView formal. Preserve the existing owning-wrapper→ByteView scalar callee bounds-defect row on interpreter fixture/file runners and native application C. This is a compatibility regression, not a new ByteView scalar-to-scalar admission. Separately test `helper(byte_at(bytes, index))` inside a ByteView caller: OOB occurs in the caller before helper entry, emits no helper invocation, and preserves existing live-owner abort-like behavior. Never label that as a bounds fault inside a U64/Bool helper.
- **D-18:** **Dated acceptance clarification (2026-10-14):** Preserve the original CALL-01/02 and Phase 36 roadmap demands as provenance; interpret them through D-08/D-15/D-17. New-callee defect rows are overflow/remainder; bounds remains the existing ByteView-callee regression plus the new argument-order control. Check lowers the actual source argument/result; raw peers independently enforce valid IR contracts; engines catch same-typed argument/result substitutions and swallowed numerical defects; executionpeer catches causal and attribution forgeries. Record this in the live requirements and final verification. Do not claim that a source-blind peer can reject well-typed IR merely because it differs from source intent.

### Compatibility, evidence, and runnable gain

- **D-19:** Before compiler edits, commit an enumerated literal generated-C baseline through all relevant current emitter routes for tracked source programs without the newly admitted scalar calls/signatures. Record non-emitting rows explicitly. Keep earlier frozen goldens immutable; no broad path exemptions, regenerated expected C, or accidental removal of old programs. Arm live C-byte drift, omitted-row, and false-exemption controls. Intentional new source fixtures get explicit rows, not a waiver for unrelated output changes.
- **D-20:** Make a small ordinary-source `crc_step`-style U64 helper that performs a reflected bit step using Phase 35 operators and is called repeatedly in a loop. Add a local call and an importing module using that helper, with literal independently calculated answers. This is a bounded helper witness, not the Phase 37 complete CRC file consumer. Cover all four U64/Bool signature pairs, nested/computed arguments, result used in an arithmetic expression and mutable assignment, and a Bool result used in a condition. Use discriminating nonidentity values; identity/zero alone cannot detect dropped arguments or results.
- **D-21:** Keep public route claims exact: parameterized U64 `main` still needs input and public conformance `run` does not gain a new U64 input flag/default. A suitable unused-Byte public wrapper may demonstrate a constant-fed helper while preserving unused-Byte laws. Native retained applications and the existing interpreter application seam prove exact decimal stdout separately. The ByteView owning wrapper uses its existing application/fixture path; do not assert that the conformance shell admits it. Public interpreter file routing remains Phase 37.
- **D-22:** Plan finite controls per actual knower: checker source-lowering argument/result selection, independent raw-core/origin refusal-removal tests, each engine's literal answers and dropped-argument/result/swallowed-defect controls, and executionpeer's event/ordinal/attribution controls. Use fresh positive/no-op/restored baselines; a compile/vet failure, missing test selector, unrelated assertion, or source rejected before the target is not a killed semantic mutant. Reuse the small overlay/local-seam patterns from Phase 35; no mutation dependency or generic framework.
- **D-23:** Extend deterministic `schway.surface/1` with exact scalar call contexts, U64/Bool input/result pairs, arity one, no recursion, bounded argument expressions, Bool identifier-tail admission, and unchanged absent Bool literals/Unit value calls/ByteView scalar-call arguments. Preserve source-free discovery and its independently authored complete-program witnesses, finite admission/limit/source census, bidirectional missing/false-admission/refusal/hint/restriction controls, and no producer-fed expected manifest. Only change the specific obsolete scalar-expression-call absence claim.
- **D-24:** Use focused exact named selectors with `GOCACHE=/tmp/ai-lang-verification-gocache`, `-count=1 -v -timeout=600s`. No local full suite, whole-package run, broad cross-program native matrix, push, hosted dispatch, or hosted receipt binding. Run bounded helper answer/defect rows on interpreter and native O0/O3/O3-LTO; sanitizer checks supplement independent expected values. State local host/toolchain, no skips as passes, and separate causal capture from numeric/physical proof. Human readability and complete cold/warm feedback-loop distributions remain unmeasured.
- **D-25:** Target about eight coherent plans, each tied to the runnable helper or a named safety obligation. Freeze first, then bounded source/core contracts, independent analyses, engine/evidence integration, and finite compatibility/manifest/consumer acceptance. Extra plans need an explicit witness or safety reason, not one plan per role. All SUMMARY coverage uses real named tests and the closed kind vocabulary; prefer automated acceptance over human UAT for deterministic properties.
- **D-26:** At this transition retain dormant z1g/trz/z1s with the table below. Record source inspection separately from fresh execution and historical reports. Append dated living-document amendments; preserve earlier decisions/receipts. The next three capabilities remain Phase 36 helper composition, Phase 37 file CRC/second consumer/interpreter CLI/guide, and Phase 38 final hosted binding. No query database, telemetry framework, or source-form study is justified by generic forecasts.

### Planner's Discretion

Local AST/helper/test layout, bounded dynamic invocation implementation, precise test vector count and corpus enumeration, independent-peer refusal wording consistent with existing codes, and dependency-aware wave split. Research should validate the ordinal/capacity representation before plan commitments; it may choose mechanics but cannot silently replace unique activation identity, inflate evidence guarantees, or widen signature/ownership boundaries.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Ownership and settled boundaries

- `AGENTS.md` — active workflow and living-transition requirements.
- `.planning/ROADMAP.md` Phase 36 and successors; `.planning/REQUIREMENTS.md` CALL-01/02 and dated clarification; `.planning/PROJECT.md` verification/dependency preferences; `.planning/STATE.md` current position and local-test restrictions.
- `.planning/research/M007/SUMMARY.md` — named CRC consumer and milestone scope; its source estimates predate Phase 35.
- `.planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-CONTEXT.md` D-03/D-05/D-07/D-16..18 — expression order/depth, unused Byte, independent knowers.
- `.planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-VERIFICATION.md`, `.planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-ADMISSION-MUTATION-EVIDENCE.md` — historical accepted behavior and valid implementation-mutation discipline.
- `.planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-CONTEXT.md` — byte identity, diagnostics, finite independent manifest witnesses and refusals.
- `.planning/phases/33-carried-debt-bounded-capture-read-and-generic-freeze-gate/33-CONTEXT.md` D-12..16 — owner/reason-scoped defects, bounded capture and capacity distinctions.
- `.planning/milestones/M006-phases/32.1-address-tech-debt-cli-exit-classes-and-module-aware-routes/32.1-CONTEXT.md` — public input, exit and module-aware route contracts.

### Source and evidence witnesses

- `internal/compiler/syntax/parser.go`, `internal/compiler/syntax/format.go`, `internal/compiler/ast/ast.go` — scalar routing, bounded expression grammar, identifier tails, lossless formatting.
- `internal/compiler/check/check.go`, `internal/compiler/check/scalar.go`, `internal/compiler/core/core.go`, `internal/compiler/callgraph/callgraph.go` — old call lowering, scalar fixed point/signatures, arity-one IR, all-function cycle check.
- `internal/compiler/corevalidate/corevalidate.go`, `internal/compiler/originvalidate/originvalidate.go`, `internal/compiler/pathoracle/pathoracle.go` — independently enforced scalar/origin/authority contracts.
- `internal/compiler/execution/execution.go`, `internal/compiler/execution/invocation.go`, `internal/compiler/executionpeer/executionpeer.go` — actual event payload, existing ordinal grammar, bounded causal membership and current loop restrictions.
- `internal/compiler/interp/interp.go`, `internal/compiler/cgen/cgen.go`, `internal/compiler/cgen/cgen_program.go`, `internal/compiler/native/native_app.go` — interpreter frames, static path preflight, event reservation, C generation and capture boundaries.
- `internal/compiler/session/session.go`, `cmd/schway/main.go` — conformance/app entry-input distinctions.
- `internal/compiler/cgen/cgen_phase29_fixture_test.go`, `internal/compiler/session/session_phase29_interp_fixture_test.go`, `internal/compiler/session/session_phase29_native_fixture_test.go` — existing ByteView owning-wrapper/callee evidence; resolve exact selectors from source before planning.
- `internal/compiler/session/session_phase35_prechange_test.go`, `internal/compiler/session/session_phase35_peer_mutation_test.go`, `internal/compiler/session/session_phase35_admission_mutation_test.go` — finite compatibility and implementation-control precedents.
- `internal/compiler/surfacemanifest/manifest.go`, `internal/compiler/surface/refusals.go`, `internal/compiler/session/session_phase34_surface_test.go`, `testdata/phase35/surface-census.json` — manifest and independent witness/census consumers.
- `examples/checksum/adler32.schway`, `examples/checksum/checksum.schway`, `examples/phase29/byte_index_oob.schway` — current loop, importing wrapper and bounds witnesses.

### Living records and primary references

- `.planning/PRODUCT-ROADMAP.md`, `.planning/LANGUAGE-MATURITY.md`, `.planning/sources/DASHBIT-AI-ERA.md`.
- `.planning/seeds/SEED-261008-z1g-bounded-symbol-and-caller-query.md`, `.planning/seeds/SEED-261008-trz-consumer-triggered-failure-observability.md`, `.planning/seeds/SEED-261008-z1s-compare-contract-and-source-forms.md`.
- [Rust operand evaluation order](https://doc.rust-lang.org/reference/expressions.html#evaluation-order-of-operands) — left-to-right explicit language precedent, not wholesale Rust grammar adoption.
- [Go order of evaluation](https://go.dev/ref/spec#Order_of_evaluation) — contrasting host-language rule; Schway's ordered core is its authority.
- [WG14 N1570](https://www.open-std.org/jtc1/sc22/wg14/www/docs/n1570.pdf) §6.5/§6.5.2.2 — C11 committee-draft ordering clauses relevant to the retained C17 backend; do not mislabel N1570 as a C17 draft.
- [Clang UBSan checks](https://clang.llvm.org/docs/UndefinedBehaviorSanitizer.html#available-checks) — supplemental C backend fault detection, not an independent Schway numeric specification.

</canonical_refs>

<code_context>
## Existing Code Insights

No Phase 36 directory, SPEC, CONTEXT, plan, interruption checkpoint, codebase maps, matching todo, or global advisor profile existed at initialization. Three source-focused advisors supplied compiler, evidence/security and engine/compatibility findings. These are inspections/design inferences, not executed feature evidence.

### Reusable Assets

Existing direct `OpCall`, one `Parameter`, declared signature lookup, merged local modules, cycle checker, static invocation paths, interpreter frames, named add/remainder/byte defects, canonical ordinal formatting, application runners, golden comparators and finite mutation overlays all remain useful. The general call machinery predates this phase; scalar builder expression calls, Bool formals/results and repeated call activation handling do not.

### Established Patterns and Integration Points

Bounded parser/AST → scalar source lowering/core → independent core/origin analyses and graph gates → interpreter/C emission and causal peer → session/public routes. Manifest is a leaf consumer. Unknown scalar call results are sufficient for safe intraprocedural analysis. Source provenance, numeric answers, causal structure, physical cleanup, capture completeness and host receipts are separate claims.

`invocationEventCapacity` currently sums static operations; `trackEventOccurrences` keys on cyclic scalar copies; engines and causal-peer append ordinal zero; executionpeer's repeated-event allowance is narrower than calls. Extending only `scalar.go` would leave the loop runnable gain incomplete. Existing U64/Bool-only helpers cannot access bytes; keep D-17's witness separation.

</code_context>

<specifics>
## Runnable Gain and Transition Review

Use a small reflected CRC step such as shifting right and conditionally XORing the reflected polynomial, with an arity-one U64 state. Call it more than once from a bounded loop and from an importer; assert literal hand/independently calculated answers. Full ISO-HDLC file folding and its final XOR remain Phase 37 work. Include nonidentity U64 and mixed Bool signature rows so argument/result substitutions are observable.

| Capability / user-visible program | Current blocker | Smallest complete slice | Checker/guarantee and evidence/debt | Owner / next action | Observation changing priority |
|---|---|---|---|---|---|
| Imported `crc_step` helper in a scalar loop | Scalar-expression calls, Bool signatures and repeated call evidence absent | Phase 36 arity-one local/imported helper with a >1-activation loop and exact answer | CALL-01/02; independent IR contracts, argument/result/defect engine controls, ordinal/causal/capacity controls; D-18 clarification | Plan/execute Phase 36; validate source/evidence scaling before commitments | A concrete consumer needs another signature/authority shape or the bounded identity representation fails |
| `crc32 FILE`, second importer and interpreter CLI use | Helper support plus public interpreter file route absent | Bounded ordinary ISO-HDLC module, second consumer and test-run guide | APP-13/14, MOD-07, DX-24; independent vectors, file/cleanup outcomes, mutation, CLI parity; capture limits remain explicit | Phase 37 | A named file consumer needs a different input/error contract, or capture overhead changes its ordinary-run design |
| Reproducible checksum/CRC claims on two hosts | Final code and audit not frozen | One M007 manifest and final revision-bound Ubuntu/macOS receipts | EVD-17; generic freeze/receipt negative controls, no local substitute for hosted binding | Phase 38 after milestone audit/gap closure; explicit push/dispatch authority at its checkpoint | Source changes after freeze or host/toolchain evidence disagrees |

| Seed | Disposition | Current source witness | Named consumer / trigger assessment | Next action |
|---|---|---|---|---|
| z1g | retain dormant | `internal/compiler/session/session_phase6_query.go#QueryCommandFile`; ordinary source search | `crc_step` declaration/call navigation; no measured query/search gap established | Consumer owner must name/measure a task gap before promotion |
| trz | retain dormant | `internal/compiler/native/native_app.go#readApplicationCapture`; `internal/compiler/session/session_app_verify.go` | Helper/CRC failure; current source and capture explain known boundaries, no unexplained failure; old allocation issue resolved in Phase 33 | Review one actually unexplained failure before extending observation |
| z1s | retain dormant | `internal/compiler/syntax/format.go#Format` | Helper source review; no actual reviewer contract/source-form dispute | Require named dispute and independent evidence before experiment |

Evidence status: this discussion ran no Go/native behavior tests. Source inspection and primary documentation support decisions only. Phase 35 verification and M006 receipts remain historical at their recorded revisions. Final hosted binding is pending in Phase 38. Human readability and whole-loop cold/warm distributions are unmeasured.

</specifics>

<deferred>
## Deferred Ideas

ByteView scalar-to-scalar arguments, arity-N, aggregate returns, general expression/Bool literal tails, indirect/recursive calls, constant-return interprocedural propagation, value replay in executionpeer, general trace summarization, a new query/observability framework, and separate compilation require a named consumer and their own ownership. No new dependency is justified here. Numeric spelling debt from Phase 35 remains preserved. Complete CRC/second consumer/interpreter file CLI/guide stay Phase 37; hosted final evidence Phase 38; word counter/longest_line and scalar aggregates stay future M008.

</deferred>

---
*Phase: 36-scalar-helper-calls*
*Context gathered: 2026-10-14*
