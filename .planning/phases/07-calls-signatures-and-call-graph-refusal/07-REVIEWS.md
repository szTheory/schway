---
phase: 07
reviewers: [codex, opencode, antigravity, gemini, cursor]
reviewed_at: "2026-09-08T15:15:30Z"
plans_reviewed: [07-01-PLAN.md, 07-02-PLAN.md, 07-03-PLAN.md, 07-04-PLAN.md, 07-05-PLAN.md]
models:
  codex: "unknown"
  opencode: "unknown"
  antigravity: "unknown"
  gemini: "unknown"
  cursor: "unknown"
model_sources:
  codex: "unknown"
  opencode: "unknown"
  antigravity: "unknown"
  gemini: "unknown"
  cursor: "unknown"
---

# Cross-AI Plan Review — Phase 07

**Lanes that produced a review:** codex, opencode, antigravity (3 of 5).
**Lanes that failed:** gemini (account ineligible — Gemini Code Assist free tier
is no longer supported for this client), cursor (not authenticated).
`claude` was skipped for independence (this session is Claude Code).

All three reviewers that ran had repo access and cited `file:line` evidence.
All three returned **Risk: HIGH**.

## Consensus Summary

Three independent reviewers converged, without prompting, on the same class of
defect: **the plans specify data flows that the current architecture cannot
carry.** The design intent (independent re-derivation, fail-closed defaults,
bounded diagnostics, mutation-killed controls) is rated strong by all three; the
plumbing to realize it is missing from the plans, not merely from the code.

The orchestrator independently verified every finding below against the shipped
tree. All are CONFIRMED, and two are worse than the reviewer stated.

### Agreed Concerns

**C-1 — `Callable` cannot be derived by either producer or peer (all three reviewers, HIGH).**
07-03 has both `originvalidate` and `corevalidate` derive `Callable` from "the
program's own declared export set". No export set survives lowering:
`ast.Program.Exports` is produced at `syntax/parser.go:160` and dropped at the
AST-to-core boundary; `core.Program` (`core/core.go:10-16`) and `core.Function`
(`core/core.go:25-47`) carry no export or publication marker; the word "export"
appears nowhere in the core schema. **Verified stronger than reported:** both
`originvalidate.BuildInterface(core.Program)` (`originvalidate.go:415`) and
`corevalidate.Validate(core.Program)` (`corevalidate.go:56`) take *only* a
`core.Program`, so this breaks the producer too, not just the peer. SEM-06 is
unimplementable as planned without a core-schema change.

**C-2 — `core.OpCall` has nowhere to record its callee (codex + antigravity, HIGH).**
`core.LinearOperation` (`core/core.go:261-300`) has `SourceID`, `TargetID`,
`LoanID`, `TypeID`, the three foreign-edge fields, `ReleasesOperationID`,
`Allocator`, and `Reason` — and no callee identity. `grep` across all five plans
and CONTEXT.md returns **zero** occurrences of `CalleeID` or any equivalent. Yet
07-04 repeatedly says "an `OpCall` whose callee **resolves** to no declared
function", presupposing a resolution mechanism that is never specified. The
existing `OpForeignCall` precedent does not generalize: it carries symbol
identity on `Function.ForeignContract`, which works only because a function has
at most one foreign contract — a function can call many callees. Without this
field `callgraph` has no edges, and Plans 04 and 05 have no input.

**C-3 — `check` has no summary input channel (codex, HIGH; implied by opencode).**
`check.Program(ast.Program) Result` (`check/check.go:44`) receives no
`core.Interface`, and `session.Check` (`session.go:598`) passes none. The phase's
central claim — that call admission consults only the callee's `/1` summary — is
not reachable through any planned API.

### Highest-value single finding

**C-4 — 07-03 enforces a different rule than D-04-03 (codex, HIGH — CONFIRMED).**
D-04-03 defines callable as "`originvalidate.ValidatePublished` would publish
it." `ValidatePublished` (`originvalidate.go:346-358`) recomputes **return-origin
safety** per function and refuses with `core.origin_omitted`; it never consults
an export list. Export membership and publishability are different predicates.
The plan's negative control (an unexported callee) therefore tests the wrong
rule, and its proposed `export_callee` repair cannot fix an unsafe
borrow-derived return. This would have shipped a green gate enforcing a rule the
milestone did not ask for.

### Corrections to Phase 07 planning artifacts

**A-03 in `07-CONTEXT.md` is partly wrong — my error, surfaced by codex.**
I recorded that `exclusive_borrow_clean` "does not exist under any name today."
It does exist, as inline Go source at `check/check_exclusive_test.go:46-58`
(module `owned.exclusive_borrow_clean`, `fn relay`), and the clean-but-
unpublishable witness is asserted again at `originvalidate_test.go:230`. There is
no `testdata/` **file**, which is what I actually checked. The Wave 0 task should
be "extract the existing inline witness", not "author a new one" — and the
existing witness is the *right* one, because it fails publication with
`core.origin_omitted`, i.e. it exercises C-4's real predicate.

Relatedly, the `relay`/`escort` witness recorded in D-04-03's narrative uses
nested `relay(borrow mut buffer)`, which **D-07-01 has just made ungrammatical**.
It cannot be copied verbatim; it must be deliberately converted to A-normal form
and shown semantically equivalent.

### Other confirmed findings

- **`LinearOperation` has no `Span`** (`core/core.go:261-300`) — verified. 07-04
  promises `Primary` = the closing call site's span and a span on every
  `cycle_member` cause, from a package that accepts only `core.Program`. Not
  constructible. Codex's suggestion is sound: keep an operation-ID-to-span map on
  the `check` side for diagnostic projection rather than putting spans in
  serialized core.
- **"Reuse `isDeclaredFunctionName`" is impossible literally** (codex) — it is
  unexported in `corevalidate` (`corevalidate.go:1654`), and 07-04 line 199 puts
  `corevalidate` in `callgraph`'s own forbidden-import list. The precedence rule
  must be restated independently. Note it also matches on `Name`, not ID.
- **Strict `/1` decoding is unspecified** (codex) — removing `omitempty` does not
  make Go reject missing fields; `CheckSummary` (`originvalidate.go:459`)
  unmarshals straight into `core.Interface` and checks only JSON validity and
  `CoreDigest`. The threat-model claim that a missing `Mode`/`Return`/`Foreign`
  is refused is currently unsupported, and nothing routes `/0` documents to the
  pinned `InterfaceV0` decoder.
- **`ClosureDigest` is self-referential as specified** (codex) — "digest of the
  function's own `/1` signature bytes" includes `ClosureDigest` unless a separate
  preimage type excludes it. No canonical preimage, callee ordering, or domain
  separator is defined. This is A-06's open base case, and it is larger than A-06
  recorded.
- **Digest chaining is ordered before cycle refusal** (codex) — 07-02 chains
  callee digests; cycle detection lands in 07-04. A cyclic program can reach a
  non-terminating digest computation before the gate that would refuse it exists.
- **Interpreter and cgen would fabricate call semantics** (codex) — folding
  `OpCall` into the grouped copy/move/borrow arms (`cgen.go:259`,
  `interp.go:391`) emits copy-like C and pass-through values, making a call look
  executed when no callee ran. Both need explicit "recognized but unsupported
  this phase" arms so a green exhaustive-dispatch control does not certify a stub.
- **Existing tests must change and the plans say they need not** (codex) —
  `TestFallibleCallUnconsumedRejected` (`syntax_test.go:1004`) pins the refusal
  at `syntax.Parse`, and `native_test.go:1173` expects `format --check` to reject
  the fixture syntactically. D-07-01 moves that enforcement.
- **QLT-08's timing discipline is not met** (codex) — 07-02 introduces both
  dispatch controls with no seeded mutation, and 07-04 introduces cycle refusal
  while deferring its mutation corpus to 07-05. Either move each kill into the
  plan that introduces its control, or amend QLT-08 explicitly.
- **Exported mutable fault seams on production paths** (codex, MEDIUM) — package
  globals risk `-race` and test-order dependence; prefer unexported injected
  predicates on a private walker config.
- **Canonical rotation does not make cycle choice deterministic** (codex, MEDIUM)
  — rotation normalizes *one* witness; a graph with several cycles can still
  yield different witnesses under different root orderings. Needs sorted roots
  and adjacency plus a deterministic witness-selection rule.

### Agreed Strengths

- The diamond/shared-leaf corpus requirement is singled out by two reviewers as
  materially better than a deep chain for separating gray re-entry from black
  revisit.
- Iterative three-color DFS over an explicit stack (not Go recursion) is the
  right DoS posture; bounding the *diagnostic* at 32 while leaving *traversal*
  unbounded is called a strong separation of correctness from reporting.
- Cause-order diagnostic-ID sensitivity is real and correctly identified
  (`diagnostic.go:108-117`).
- Modeling `callgraph` on `pathoracle` — typed errors, synthetic artifacts,
  static import-independence test — matches shipped practice.
- Recognizing `cgen`'s `len(Functions) == 1` gate rather than pretending `OpCall`
  emission is exercised (`core_test.go:347-352`).
- `Callable`'s Go zero value being `false` is a sound fail-closed default —
  *once* strict `/1` decoding exists.

### Divergent Views

- **Antigravity** proposes accepting `ErrorWithRepairs` with an `export_callee`
  repair for `core.callee_not_callable`. **Codex** argues that repair is invalid
  because export membership is not the predicate (C-4). Codex is correct on the
  evidence: `ValidatePublished` never reads exports. Resolve in codex's favor.
- **Antigravity** rates the three-color DFS complexity concern MEDIUM (fearing
  exponential traversal); the plans already specify white/gray/black, which
  bounds it at O(V+E). This one is already handled.
- **OpenCode** rates the phase-scoped list proliferation MEDIUM drift risk;
  neither other reviewer raises it. It is a real but manageable maintenance cost.

---

## Codex Review

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

---

## OpenCode Review

I’m checking the cited source files against the plan claims before I review the plans.
I’ve confirmed the key code paths the plans talk about. I’m checking the test scaffolding and isolation patterns next.
I’ve found the main test and independence hooks. I’m checking one structural assumption now: whether the core artifact actually carries the export information the later plans want to re-derive from.
The core artifact itself doesn’t carry an export set, so I’m checking whether the plan’s wording matches some other in-tree representation or if that’s a real gap.
I found a concrete mismatch: the core IR doesn’t keep an export set. I’m confirming whether the checker still retains that fact anywhere, or whether the plan is assuming a field that doesn’t exist.
I have the likely blocker: the planned peer needs export metadata, but the core IR doesn’t obviously carry it. I’m grabbing the AST/source-side representation to separate “needs a new core field” from “already present somewhere else.”
I’m verifying whether exports are already threaded through the checker. If they aren’t, that confirms the plan needs a structural addition, not just a re-derivation.
**Summary**
The phase 07 plan is well-structured: it uses explicit gates, phase-scoped control lists, mutation-kill seams, and independent derivations in the right places. The biggest problem is in 07-03: it assumes `Callable` can be re-derived from a core-only artifact, but the current core IR does not preserve publishability/export data, so that independence claim is not implementable as written.

**Strengths**
- Good reuse of proven patterns: the mutation seam shape matches `pathoracle`’s test-only override pattern and four-beat mutation test style (`pathoracle.go:51-66`, `pathoracle_test.go:268-295`).
- Good avoidance of false coverage: the plan explicitly recognizes `cgen`’s `len(Functions) == 1` gate, so it does not pretend `OpCall` emission is exercised when it cannot be (`core_test.go:347-352`, `session.go:2560-2565`).
- Good control drift protection: the phase-07 lane/script approach mirrors the Phase 5 set-equality pattern (`session_phase5.go:20-27`, `session_phase5_test.go:291-295`).
- Good diagnostic discipline: the plan accounts for cause-order identity and bounded truncation with the same kind of load-bearing semantics already used elsewhere (`diagnostic.go:108-117`, `protocol.go:123-129`, `session_phase6_evidence.go:21-25`).

**Concerns**
- **HIGH**: 07-03’s `Callable` peer cannot be implemented from the current core shape alone. Source exports live in `ast.Program.Exports` (`ast.go:5-9, 47-51`), but `check.Program` lowers into `core.Program` with only `Schema`, `Module`, `ModuleID`, `DataTypes`, and `Functions` (`check.go:44-125`, `core.go:10-16`). `core.Function` also has no publication marker (`core.go:25-47`). The plan needs a durable core-side publishability fact or the peer independence claim breaks.
- **MEDIUM**: The plan introduces a lot of phase-scoped lists, scripts, and tests across 07-01 to 07-05. That is correct shape, but it is high drift risk because the repo already relies on strict duplication checks to keep these lists aligned (`session_phase5.go:20-27`, `session_phase5_test.go:291-295`).
- **MEDIUM**: The cycle-refusal work depends on two independent graph walkers, shadowing precedence, and several mutation seams. That is the right design, but it is easy to accidentally share logic or miss a root-order edge case unless the synthetic fixtures are extremely strict.
- **LOW**: Some acceptance around `cgen` is documentation-only by design, not behavioral. That is fine, but the plan should keep saying that explicitly so a green control is not mistaken for real multi-function codegen coverage.

**Suggestions**
- Add a concrete core-level publishability marker before 07-03, or revise `Callable` so it is derived from a fact already preserved in `core.Program`.
- Keep the phase-07 control lists and verify scripts under strict set-equality tests, like Phase 5, to prevent silent drift.
- Make the synthetic fixtures the primary truth for 07-04/07-05, especially the diamond/shared-leaf corpus and the shadowing case.
- Add one explicit note in the plan summary that `cgen`’s `OpCall` arm is hygiene-only this phase, not runtime coverage.

**Risk Assessment**
**HIGH**. The overall structure is strong, but the `Callable` independence assumption is currently unsatisfied by the core shape, and that fact sits on the critical path for the later call-admission and cycle-refusal phases.

---

## Antigravity Review

# Summary
This phase provides a rigorous, highly decoupled plan for implementing interprocedural calls, signature summaries, and call-graph cycle detection in Codename Lang. The plan exhibits an excellent defensive posture, utilizing fail-closed invariant checking, redundant derivations (peer validations in `corevalidate` and `originvalidate`), and sophisticated fault-injection techniques (`TestStage0SummaryMutationMatrix`, `TestCallGraphMutationMatrix`) to ensure security controls are active. However, there are a few critical schema omissions in `core.go` that will cause the implementation steps to fail because the necessary facts (such as export lists and callee IDs) are not persisted into the `core.Program` artifact that the peers and call-graph builder consume.

# Strengths
- **Defensive Peer Architecture**: The `corevalidate` peer re-derives `Callable` and cycle refusal entirely independently using a disjoint reachable input space (hand-built synthetic ASTs). This strictly isolates the checker logic from the validation logic, ensuring a compromised or buggy `check` artifact is caught. (e.g. `corevalidate.go:352-361` precedent).
- **Fail-Closed Access Control**: Implementing `Callable` with a Go zero-value of `false` perfectly satisfies the requirement that absence of the fact results in refusal.
- **Stability and API Integrity**: Explicit handling of canonical rotation for diagnostic causes in cycles guarantees stable `Diagnostic.ID`s across discovery orderings (`diagnostic.go:96-151`, D-07-16), which is essential for downstream tooling (`lang-repair`).
- **Mutation Testing and Seams**: The use of explicit fault-injection seams (following `pathoracle.go:51-59` style) and the requirement that every control must fail a seeded mutation or fail the matrix (QLT-08) is a gold-standard approach to verifying control effectiveness.
- **Explicit Cycle Traversal**: Using an explicit stack with three-color DFS instead of native Go recursion prevents DoS attacks on deeply nested call graphs (T-07-16).

# Concerns
- **HIGH: Missing Export Set in `core.Program`** (Plan 03, Tasks 1 & 3): The plan specifies that `originvalidate` and `corevalidate` must independently derive `Callable` by reading "the program's own declared export set and function table". However, `core.Program` (defined in `internal/compiler/core/core.go:10-16`) does not currently persist the module's `export` block. Since both validators operate solely on the compiled `core.Program` and do not have access to the AST (`ast.Program`), they structurally cannot know which functions are exported unless a schema change adds this information (e.g., an `Exported []string` field on `core.Program` or an `Exported bool` on `core.Function`). This omission breaks the core mechanism of SEM-06.
- **HIGH: Missing `CalleeID` for `core.OpCall`** (Plan 01, Task 1 & Plan 04): The plan adds `OpCall` to `core.OperationKind` but does not specify how the call target is persisted in `core.LinearOperation`. Since functions are not currently first-class values passed via `SourceID`, `core.LinearOperation` (`core.go:261-285`) needs an additive field (like `CalleeID string`) to store the resolved function ID. Without this, the `callgraph` traversal has no edges to follow.
- **MEDIUM: Call Graph Re-entry vs Subgraph Visits**: For `deep_diamond_acyclic.lang` (Plan 05, Task 2), the plan expects a bounded execution for a 200-node graph. With diamonds and shared leaves, a naive DFS without a "black" (fully visited) set will traverse paths exponentially (O(2^N)), resulting in a timeout. The plan states "three-color DFS" (white/gray/black), which correctly solves this by memoizing visited nodes, but the `gray` (on-stack) vs `black` (finished) distinction must be strictly enforced so that `Order` remains O(V+E) and doesn't timeout the 60-second limit.

# Suggestions
- **Widen `core.Program` for Export Set**: Explicitly add an `Exported bool` field to `core.Function` or an `ExportedFunctions []string` array to `core.Program` in `core.go` during Plan 01 so that `originvalidate` and `corevalidate` can read the access-control facts directly from the checked core. 
- **Add `CalleeID` to `core.LinearOperation`**: In Plan 01, explicitly add an `omitempty` string field (e.g., `CalleeID`) to `LinearOperation` so `OpCall` has a place to record its target function.
- **Diagnostic Repair Kind**: For the `core.callee_not_callable` code decision (Plan 03), accept the `ErrorWithRepairs` option (`export_callee`) as proposed, since it provides actionable, direct remediation for developers.

# Risk Assessment
**HIGH**
While the conceptual design, test plans, and security postures are extraordinarily thorough and well thought out, the plan fails to account for necessary data in the `core.Program` schema. Specifically, the absence of an export set and a callee identifier in the intermediate representation means that the independent validations and call graph traversals structurally cannot be implemented without unplanned schema modifications. Resolving these omissions in the plan before execution will lower the risk to LOW.

---

## Gemini Review

> Lane failed. `IneligibleTierError: This client is no longer supported for Gemini Code Assist for individuals.` The account must migrate to Antigravity (which the `antigravity` lane above already uses). This lane is permanently unavailable on this host until that changes.

---

## Cursor Review

> Lane failed. `Error: Authentication required. Please run 'agent login' first, or set CURSOR_API_KEY environment variable.` Recoverable by authenticating the Cursor CLI.
