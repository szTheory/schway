# Phase 36: Scalar Helper Calls - Discussion Log

> **Audit trail only.** Not input for planning/research/execution. Decisions are in 36-CONTEXT.md.

**Date:** 2026-10-14
**Areas discussed:** scalar composition; Bool signatures; imports/cycles; independent peer knowledge; dynamic invocation/capacity; callee defects; compatibility/consumer evidence.
**Selection authority:** User explicitly requested fan-out across relevant roles and automatic adoption of coherent recommendations. No interactive question/answer is invented. Three typed `gsd-advisor-researcher` agents completed read-only research (helper semantics; peer evidence; engines/compatibility), with adversarial source follow-ups. No compiler behavior test ran during discussion.

## Composition and Bool surface

| Option | Pros | Cons | Complexity | Selection |
|---|---|---|---|---|
| Name-only call argument | Closest to old ordinary calls, sufficient for CRC state | Awkward special restriction, extra temporary source bindings | Smaller grammar; same peer/engine work | Considered |
| One bounded scalar-expression argument | Composable computed/nested helpers; reuses expression parser/depth | More depth/recovery/order controls | Parser/AST/lowering plus all semantic consumers | Selected under delegated authority, D-03..05 |
| U64-only helpers | Smallest current signature expansion | Contradicts CALL-01's Bool contract | Smaller signature tests | Rejected |
| All four U64/Bool pairs with identifier tails | Fulfills requirement; no Bool-literal or general-tail expansion | Requires explicit Bool formal/result admission | Parser/check/IR peers/engines/manifest | Selected, D-02/D-04 |

Preserve left-to-right ordered core operations and exactly-once argument evaluation. C operand order cannot supply the language guarantee. Existing scalar effects and defects remain visible; Unit is not a value-returning helper signature.

## Imports and graph policy

Reuse closed whole-program module resolution, explicit export identities and existing all-function cycle refusal. An entry-reachable-only cycle policy and a new package/search-path layer were rejected because neither is required by the witness. Declaration-order/name-collision controls preserve deterministic identity. Existing bounds and authority restrictions remain.

## Independent peer knowledge

| Option | Pros | Cons | Complexity | Selection |
|---|---|---|---|---|
| Shared producer summaries / same semantic mutation required at every peer | Superficially uniform | Source-blind/causal peers lack necessary values; correlated trust | New proof protocol or false claims | Rejected |
| Independently enforced call contracts; result value unknown; engine answers; causal event checks | Fits actual interfaces and preserves independent boundaries | Requires precise per-knower evidence wording | Existing analyses plus focused controls | Selected, D-08/D-15/D-18/D-22 |

Missing/wrong-typed argument/result controls belong to raw peers; source lowering and engine expected answers detect same-typed substitutions. Callee numerical defects require dynamic engine rows. A plausible success trace alone cannot be refuted by executionpeer without scalar inputs.

## Repeated activations and capacity

| Option | Pros | Cons | Complexity | Selection |
|---|---|---|---|---|
| Reuse static invocation path with event Occurrence | Retains static tables | Does not uniquely name a dynamic activation | Narrower event changes | Considered |
| Populate existing InvocationSegment.Ordinal | Unique site/parent activation identity; no schema field | Requires bounded dynamic paths/counters and independent progression checks | Interpreter/cgen/peer/capacity | Selected, D-11..14 |
| Unlimited trace / raise all limits for CRC | Easy-looking complete capture | Unbounded work/allocation; conflates ordinary execution and evidence | New runtime burden | Rejected |

Keep static/depth/document/capture bounds. Finite repeated-call positive and named exhaustion rows are mandatory; no promise of complete per-call capture for every future CRC input. Do not impose the static 4096-node cap as a new semantic dynamic-call limit.

## Defect witness correction

| Option | Pros | Cons | Complexity | Selection |
|---|---|---|---|---|
| Widen scalar helper arguments to ByteView | Could put byte_at inside any such helper | New ownership boundary without named need | Broad signature/borrow/ABI work | Rejected |
| Keep U64/Bool calls; overflow/remainder new-callee rows; preserve old ByteView bounds-callee regression | Fits actual reachable semantics and ownership scope | Requires dated acceptance clarification | Existing wrapper plus new order control | Selected, D-16..18 |

A U64/Bool helper cannot access bytes. Existing fixture-owning wrapper→ByteView callee remains a bounds regression through its application/fixture routes. A `byte_at`-computed argument can fault before helper entry and must be labelled a caller defect. Both are distinct from new U64/Bool callee overflow/remainder rows.

## Compatibility and runnable witness

Selected pre-edit enumerated immutable C corpus and armed drift/omission/exemption controls; existing goldens remain frozen. Selected small local/imported reflected CRC-step helper in a >1-call loop, all four signature pairs, independent literal answers, O0/O3/LTO feature lanes, exact named focused tests, causal/capacity controls, and independently authored manifest witnesses. Public U64 input restrictions remain; ordinary app stdout is distinct from conformance outcome.value. Complete file CRC/guide/CLI stays Phase 37. No new dependency or backend.

## Sources and adversarial review

Advisors inspected actual parser, scalar builder, call graph, independent validators, execution schema/peer, interpreter frames, C invocation preflight/capacity, legacy fixture wrapper, manifest and previous compatibility patterns. Source-era estimates in the M007 research and standing verdicts were treated as historical where current code differed.

- [Rust operand order](https://doc.rust-lang.org/reference/expressions.html#evaluation-order-of-operands): precedent for explicit left-to-right operands.
- [Go evaluation order](https://go.dev/ref/spec#Order_of_evaluation): host-language contrast, not Schway authority.
- [WG14 N1570](https://www.open-std.org/jtc1/sc22/wg14/www/docs/n1570.pdf): C11 draft ordering clauses relevant to retained C17 code generation; one advisor called it C17 and the synthesis corrects that label.
- [Clang UBSan](https://clang.llvm.org/docs/UndefinedBehaviorSanitizer.html#available-checks): supplementary generated-C checks.

Adversarial pass covered Bool input/output asymmetry, unused Byte escape, same-type substitutions invisible to raw peers, unreachable recursion, legacy ownership compatibility, repeated/nested site identity, event reservation undercount, defect before invocation, native/application route mismatch, capture exhaustion, frozen corpus omission and manifest false claims. Further generic ecosystem/role fan-out would not change a scoped choice; implementation research owns representation feasibility.

## Agent's Discretion and Deferred Ideas

Planner may choose local data structures, AST layout, finite vector/corpus size and plan waves while preserving the context contracts. Broader signatures, arity-N, aggregate returns, recursion/dispatch, exact-return summaries, replay/telemetry/query frameworks and complete trace summarization are deferred. All three conditional seeds remain dormant on the named witnesses in the context. No todos matched. Persistent auto-advance is enabled; stop the implementation chain before starting Phase 37.
