# Cross-AI Plan Review — Phase 07

## Overall assessment

The plans show unusually strong attention to negative controls, independent derivation, bounded diagnostics, and source-grounded staging. However, they are not executable as written. Several required data flows do not exist in the current architecture: `check` has no interface-summary input, `corevalidate` has no summary input or exportability facts, and `core.LinearOperation` has no callee identity or source span. These are foundational gaps that affect Plans 01–05. Replanning is recommended before execution.

Overall risk: **HIGH**.

---

## Plan 01 — `lang.interface/1` and summary peer

### Summary

The schema-first staging is sound, but the plan does not define an implementable producer/peer comparison boundary or an unambiguous digest construction. It also claims validation behavior that merely removing `omitempty` does not provide.

### Strengths

- The separate interface schema is real and appropriately isolated from `core.Program`: `InterfaceSchema` is independent of `core.Schema`, supporting the claim that an interface bump need not move core bytes. See [core.go:142](<~/projects/ai-lang/internal/compiler/core/core.go:142>) and [core.go:155](<~/projects/ai-lang/internal/compiler/core/core.go:155>).
- Reusing the existing full SHA-256 representation is correct: `digest` emits `sha256:` plus all 32 bytes in lowercase hex. See [originvalidate.go:404](<~/projects/ai-lang/internal/compiler/originvalidate/originvalidate.go:404>).
- The plan correctly identifies two genuinely separate replay paths. They dispatch independently at [corevalidate.go:841](<~/projects/ai-lang/internal/compiler/corevalidate/corevalidate.go:841>) and [corevalidate.go:1039](<~/projects/ai-lang/internal/compiler/corevalidate/corevalidate.go:1039>).

### Concerns

- **HIGH — The peer has nothing to compare against.** `corevalidate.Validate` accepts only a `core.Program`; neither it nor its validator contains an interface summary. See [corevalidate.go:53](<~/projects/ai-lang/internal/compiler/corevalidate/corevalidate.go:53>). Wiring summary comparison into replay sites therefore requires a new API such as `ValidateInterface(program, summary)`, or putting the summary into `core.Program`. Neither is planned.
- **HIGH — `/0` decoding is not actually wired.** `CheckSummary` currently unmarshals directly into `core.Interface`. Merely defining `InterfaceV0` does not make `/0` documents use it. See [originvalidate.go:459](<~/projects/ai-lang/internal/compiler/originvalidate/originvalidate.go:459>). The plan needs schema peeking and explicit `/0` versus `/1` decoding.
- **HIGH — Required-field validation is absent.** Go JSON decoding accepts missing non-`omitempty` fields as zero values. Current `CheckSummary` checks only JSON validity and `CoreDigest`, not schema or field completeness. See [originvalidate.go:459](<~/projects/ai-lang/internal/compiler/originvalidate/originvalidate.go:459>). The threat-model claim that missing `Mode`, `Return`, or `Foreign` fields are refused is therefore unsupported.
- **HIGH — `ClosureDigest` is self-referential as specified.** “Digest of the function’s own `/1` signature bytes” includes `ClosureDigest` unless a separate digest-preimage type explicitly omits it. No canonical preimage, ordering rule, or domain separator is specified.
- **HIGH — Several signature facts have no authoritative source.** The AST and core parameter types contain only name and type; neither declares `owned/shared/exclusive` or `Drops`. See [ast.go:85](<~/projects/ai-lang/internal/compiler/ast/ast.go:85>) and [core.go:173](<~/projects/ai-lang/internal/compiler/core/core.go:173>). Inferring a declared ownership convention from the body would contradict the stated signature-only contract.
- **MEDIUM — The protocol projection is a distinct schema surface.** It currently exposes only origin paths/access. See [protocol.go:132](<~/projects/ai-lang/internal/compiler/protocol/protocol.go:132>) and [session.go:1209](<~/projects/ai-lang/internal/compiler/session/session.go:1209>). The plan should decide whether the projection is the complete `/1` artifact or a deliberately lossy command answer.

### Suggestions

- Define an explicit `ValidateInterface(program, interface)` boundary and keep ordinary `Validate(program)` separate.
- Add `DecodeInterface` with schema-first dispatch, strict `/1` validation, duplicate-function checks, mode-domain checks, digest-format checks, and unknown-schema refusal.
- Define a canonical `functionDigestPreimage` excluding `ClosureDigest`, with sorted callee `(ID,digest)` pairs and a domain separator.
- Resolve where parameter mode, `Drops`, `Fresh`, and publishability originate before minting `/1`.

### Risk assessment

**HIGH.** The schema can be implemented, but its main trust and validation mechanisms are currently undefined.

---

## Plan 02 — Parser, `OpCall`, and dispatch registration

### Summary

The parser correction and phase-scoped dispatch controls are well researched. The core representation and execution staging, however, are incomplete: there is no field that records the callee, no route by which checking obtains a callee interface, and the proposed interpreter/cgen behavior risks pretending calls executed when they did not.

### Strengths

- The plan correctly discovered that bare calls are actively rejected by the parser, rather than merely unsupported by the checker. See [parser.go:411](<~/projects/ai-lang/internal/compiler/syntax/parser.go:411>).
- Reusing `callArguments()` is appropriate; the existing `try_call` constructor already establishes the `Callee`/`Arguments` AST shape. See [parser.go:461](<~/projects/ai-lang/internal/compiler/syntax/parser.go:461>).
- The exhaustive-dispatch analysis is accurate: the in-process control requires actual corpus occurrence, while cgen is guarded by `len(program.Functions) == 1`. See [core_test.go:280](<~/projects/ai-lang/internal/compiler/core/core_test.go:280>) and [core_test.go:347](<~/projects/ai-lang/internal/compiler/core/core_test.go:347>).
- The same cgen limitation exists in production at [cgen.go:16](<~/projects/ai-lang/internal/compiler/cgen/cgen.go:16>) and [cgen.go:46](<~/projects/ai-lang/internal/compiler/cgen/cgen.go:46>).

### Concerns

- **HIGH — `OpCall` cannot identify its callee.** `core.LinearOperation` has source, target, loan, type, foreign-edge, release, allocator, and reason fields, but no `CalleeID`. See [core.go:261](<~/projects/ai-lang/internal/compiler/core/core.go:261>). The plan says to emit a callee identity without specifying the field, JSON representation, validation rules, or frozen-core-byte impact.
- **HIGH — `check` has no summary input.** `session.Check` parses source and calls `check.Program(ast.Program)`; `check.Program` receives no `core.Interface`. See [session.go:598](<~/projects/ai-lang/internal/compiler/session/session.go:598>) and [check.go:44](<~/projects/ai-lang/internal/compiler/check/check.go:44>). The statement that call admission consults only the callee’s `/1` summary is not implementable through the planned APIs.
- **HIGH — Closure-digest chaining precedes cycle refusal.** Plan 02 introduces callee digest chaining, while cycle detection does not arrive until Plan 04. A cyclic source can therefore reach an undefined or recursively nonterminating digest computation before the cycle gate exists.
- **HIGH — The cgen instructions contradict the live switches.** Adding `OpCall` to the four grouped copy/move/borrow cases would generate copy-like C and events, not return “unsupported.” See [cgen.go:259](<~/projects/ai-lang/internal/compiler/cgen/cgen.go:259>) and [cgen.go:1651](<~/projects/ai-lang/internal/compiler/cgen/cgen.go:1651>). `OpCall` needs separate explicit unsupported cases.
- **HIGH — The proposed interpreter handling fabricates semantics.** The current interpreter mutates a value map for every operation and returns the resulting value. See [interp.go:391](<~/projects/ai-lang/internal/compiler/interp/interp.go:391>). Treating `OpCall` like an `ownedEvent` pass-through makes a call appear to execute without invoking the callee. Yet the exhaustive control invokes the interpreter on corpus functions at [core_test.go:329](<~/projects/ai-lang/internal/compiler/core/core_test.go:329>).
- **HIGH — Existing parser expectations cannot remain unchanged.** `TestFallibleCallUnconsumedRejected` explicitly requires `syntax.Parse` to produce the diagnostic. See [syntax_test.go:1002](<~/projects/ai-lang/internal/compiler/syntax/syntax_test.go:1002>). Once enforcement moves to `check`, that assertion must change. The shipped CLI matrix also currently expects `format --check` to reject the fixture syntactically. See [native_test.go:1173](<~/projects/ai-lang/internal/compiler/native/native_test.go:1173>).
- **MEDIUM — Call analysis lacks required context.** `analyzeStraightLine` currently receives only one function’s ID, parameter, type fact, and body. See [check.go:1804](<~/projects/ai-lang/internal/compiler/check/check.go:1804>). It cannot resolve callee summaries without a deliberate signature/API change.

### Suggestions

- Add a required `CalleeID string json:"callee_id,omitempty"` fact, valid only for `OpCall`, and validate exclusivity against foreign-only fields.
- Introduce a pre-body signature pass producing a summary table, then pass that immutable table into body admission.
- Move cycle construction/refusal before closure-digest chaining, or defer non-leaf closure digests until Plan 04.
- Make interpreter and cgen return explicit named “known but unsupported in this phase” errors, and adjust exhaustive controls so “recognized” is not confused with “executed.”
- Update parser-only and format expectations explicitly.

### Risk assessment

**HIGH.** The syntax portion is ready, but the essential IR and summary-consumption architecture is missing.

---

## Plan 03 — Callable ⊆ publishable

### Summary

The fail-closed intent is correct, but the plan misidentifies what “publishable” means in the current code. Export membership is not the D-04-03 predicate, and `corevalidate` cannot rederive either export membership or publication eligibility using the proposed inputs.

### Strengths

- The proposed exact diagnostic and negative-control approach is appropriate for an admission boundary.
- The zero value of `Callable == false` is a sound fail-closed default once strict `/1` decoding exists.
- Extracting a real fixture is useful, though the shape is already machine-tested inline.

### Concerns

- **HIGH — Publishable is not equivalent to exported.** D-04-03 explicitly defines callable as “`originvalidate.ValidatePublished` would publish it.” The implementation recomputes return-origin safety across every function; it does not consult the AST export list. See [04-CONTEXT.md:48](<~/projects/ai-lang/.planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-CONTEXT.md:48>) and [originvalidate.go:346](<~/projects/ai-lang/internal/compiler/originvalidate/originvalidate.go:346>). A merely unexported callee is therefore the wrong negative control.
- **HIGH — `export_callee` is likely an invalid repair.** Adding the function to `export {}` does not repair an unsafe borrow-derived return. The existing decisive fixture checks clean but fails publication with `core.origin_omitted`. See [originvalidate_test.go:230](<~/projects/ai-lang/internal/compiler/originvalidate/originvalidate_test.go:230>).
- **HIGH — The peer cannot derive `Callable` from export data.** `core.Program` contains no export set. See [core.go:10](<~/projects/ai-lang/internal/compiler/core/core.go:10>). More importantly, deriving the real predicate requires running an independent equivalent of `ValidatePublished`, not checking names.
- **HIGH — Caller admission still lacks a summary channel.** The same missing `check.Program` interface input from Plan 02 blocks the central requirement.
- **MEDIUM — The “missing fixture” claim is overstated.** There is no standalone file, but the exact clean-but-unpublishable source is already embedded and asserted in [check_exclusive_test.go:46](<~/projects/ai-lang/internal/compiler/check/check_exclusive_test.go:46>) and again in [originvalidate_test.go:230](<~/projects/ai-lang/internal/compiler/originvalidate/originvalidate_test.go:230>). The plan should extract this known witness rather than invent an export-based substitute.
- **HIGH — The relay/escort source cannot be copied verbatim.** The recorded witness uses nested `relay(borrow mut buffer)`, which D-07-01 now makes ungrammatical. See [04-CONTEXT.md:64](<~/projects/ai-lang/.planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-CONTEXT.md:64>). It must be deliberately converted to A-normal form and shown semantically equivalent.

### Suggestions

- Define `Callable` as the absence of the independently recomputed publication problems for that function—not as export membership.
- Extract the existing `exclusive_borrow_clean` inline source into a fixture.
- Do not offer `export_callee`; use a repair matching the actual defect, likely adding/correcting the return-origin contract, and mark it confirmation-required unless a precise source span and replacement exist.
- Specify an A-normal relay/escort witness: bind the exclusive borrow first, pass that binding to `relay`, then move the owner.
- Add per-function publication APIs so one invalid function does not reduce only to the current first-problem whole-program result.

### Risk assessment

**HIGH.** The planned predicate would enforce a different rule from D-04-03.

---

## Plan 04 — Call graph and cycle diagnostic

### Summary

Iterative three-color DFS, all-function roots, edge deduplication, and bounded diagnostics are strong choices. The plan nevertheless depends on missing core facts and cannot currently construct its promised span-bearing diagnostic.

### Strengths

- Iterative traversal is the right defense against host-stack failure.
- Bounding only diagnostic projection, not legal traversal depth, is a strong separation of correctness from resource reporting.
- The diagnostic-ID concern is real: code, primary span, and ordered causes are hashed verbatim. See [diagnostic.go:108](<~/projects/ai-lang/internal/compiler/diagnostic/diagnostic.go:108>).
- Modeling the new package after `pathoracle`’s typed errors and import-independence tests is consistent with existing practice. See [pathoracle.go:87](<~/projects/ai-lang/internal/compiler/pathoracle/pathoracle.go:87>) and [pathoracle_test.go:46](<~/projects/ai-lang/internal/compiler/pathoracle/pathoracle_test.go:46>).

### Concerns

- **HIGH — Graph edges still lack `CalleeID`.** This plan cannot enumerate `OpCall` edges until Plan 02 defines an actual callee-bearing core field.
- **HIGH — Required call-site spans do not exist in core.** `LinearOperation` has no `Span`. See [core.go:261](<~/projects/ai-lang/internal/compiler/core/core.go:261>). A package accepting only `core.Program` cannot return the closing call-site span promised for `Primary` and every member cause.
- **HIGH — “Reuse `isDeclaredFunctionName`” is impossible literally.** It is an unexported helper inside `corevalidate`, while `callgraph` is prohibited from importing `corevalidate`. See [corevalidate.go:1654](<~/projects/ai-lang/internal/compiler/corevalidate/corevalidate.go:1654>). The rule must be specified independently in terms of `CalleeID`, not reused as code.
- **MEDIUM — Cycle canonicalization is underspecified.** Rotation normalizes the same directed witness, but graphs can contain multiple cycles. Root or adjacency ordering may select different valid witnesses and therefore different IDs. `Order` should define sorted root/edge order and a deterministic rule for selecting one cycle.
- **MEDIUM — “Direct” and “mutual” fixtures are confused.** Plan 04 calls a two-function cycle “direct,” while Plan 05 separately defines self, mutual, and indirect cases. Use self/direct = length 1, mutual = length 2, indirect = length ≥3 consistently.
- **MEDIUM — Unresolved callee refusal lacks a typed identity.** `cycleError` covers cycles only. The plan needs a separate error code/type for nonexistent callees so `check` and `corevalidate` can assert the same stable fact.

### Suggestions

- Have graph edges carry `{callerID, calleeID, operationID}`. Let `check` retain an operation-ID-to-source-span map for diagnostic projection, rather than adding source spans to serialized core unless that is an intentional schema decision.
- Define deterministic sorted traversal and canonical cycle selection.
- Specify a separate unresolved-callee error type and mapping.
- Rename/redivide fixtures consistently by cycle length.

### Risk assessment

**HIGH.** The algorithm is sound, but its input and diagnostic data are absent.

---

## Plan 05 — Independent cycle peer and mutation matrix

### Summary

The corpus design is excellent, especially the insistence on diamonds and shared leaves. The main risks are that QLT-08 mutations arrive after the controls they supposedly gate, several production-global fault seams are unsafe, and the peer depends on unresolved representation problems from earlier plans.

### Strengths

- The diamond/shared-leaf corpus is materially better than a simple deep chain for distinguishing gray re-entry from black revisitation.
- Synthetic `core.Program` tests are the right way to give `corevalidate` a disjoint reachable input space.
- Testing unreachable cycles by rooting traversal at all declared functions is appropriately fail-closed.
- The existing `pathoracle` precedent supports synthetic artifact validation independent of parsing. See [pathoracle.go:112](<~/projects/ai-lang/internal/compiler/pathoracle/pathoracle.go:112>).

### Concerns

- **HIGH — QLT-08 is not enforced “in the plan that introduces” each control.** Plan 02 introduces both dispatch controls without a seeded mutation. Plan 04 introduces cycle refusal but defers the parser-shaped mutation-kill corpus to Plan 05. The phase-wide matrix is useful, but it does not satisfy the stated temporal discipline.
- **HIGH — Exported mutable production seams conflict with `go test -race`.** Package-level override variables are global mutable state. The existing precedent at [pathoracle.go:51](<~/projects/ai-lang/internal/compiler/pathoracle/pathoracle.go:51>) is already delicate; multiplying these across production `check` and `corevalidate` paths risks races and order dependence, especially under shuffled or parallel tests.
- **HIGH — Independent disable tests require architecture not specified earlier.** There is no planned API allowing `check`’s refusal to be disabled while still passing the resulting cyclic artifact into `corevalidate`, or vice versa, without exported production-global switches.
- **HIGH — Phase-wide completeness can become a self-authored registry.** A table listing “every control” proves completeness only if independently compared with the authoritative phase control list. Otherwise a newly introduced control can be omitted from both the table and its mutations.
- **MEDIUM — The synthetic builder must satisfy the entire validator first.** `corevalidate.run` validates schema, module identity, body union, types, places, and operation consistency before any proposed call-graph peer. See [corevalidate.go:80](<~/projects/ai-lang/internal/compiler/corevalidate/corevalidate.go:80>). The plan should specify whether cycle validation runs before body replay or provide a fully valid synthetic builder.
- **MEDIUM — Determinism needs production ordering, not only repeated tests.** Map-backed fixture builders can happen to pass twice. Root IDs, adjacency lists, emitted members, and matrix rows should all be explicitly sorted.

### Suggestions

- Move each required mutation kill into the plan that introduces its control, or weaken and formally amend QLT-08’s timing requirement before execution.
- Prefer unexported injected predicates on a private walker configuration used by same-package tests. Avoid mutable exported globals on production paths.
- Derive the phase-wide control list from `Phase7RequiredControls()` and compare it with mutation coverage by exact set equality.
- Place whole-program cycle validation at a clearly defined point in `corevalidate.run`, before per-function replay can mask it with an unrelated error.
- Add deterministic sorting assertions for graph nodes, edges, witnesses, and mutation results.

### Risk assessment

**HIGH.** The test strategy is strong, but its control plumbing and timing do not presently satisfy QLT-08.

---

## Recommended replanning order

1. Define the interprocedural data model: `CalleeID`, signature ownership authority, publication/callability fact, and digest preimage.
2. Define the APIs: pre-body signature pass, `check` summary input, strict interface decoding, and `corevalidate` summary-comparison entry point.
3. Move cycle refusal before closure-digest chaining.
4. Decide honest Phase 07 behavior for interpreter and cgen without fabricating call execution.
5. Reassign mutation kills to the plans that introduce each control.
6. Only then retain the existing five-plan decomposition.

With those changes, the staged shape can work. Without them, execution is likely to discover architectural changes mid-plan, and several green tests could certify stubs or a different semantic rule than the milestone requires.
