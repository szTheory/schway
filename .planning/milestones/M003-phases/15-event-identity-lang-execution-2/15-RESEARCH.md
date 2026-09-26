# Phase 15: Event Identity (`lang.execution/2`) - Research

**Researched:** 2026-09-19
**Domain:** Versioned execution documents, deterministic call-occurrence identity, and bounded native C emission
**Confidence:** HIGH

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-15-01:** `Event.Invocation` is a required, non-empty field only under
  `lang.execution/2`. `/0` and `/1` document bytes and validation semantics
  remain frozen. — **Reversibility: one-way** — changing the public `/2` grammar
  after consumers and goldens exist requires a schema migration.
- **D-15-02:** The canonical grammar is
  `inv:entry:{E(entryID)}(/{E(opCallID)}#{ordinal})*`. The entry frame is
  `inv:entry:{E(entryID)}`; a child appends its executed `OpCall` operation ID
  and the count of prior executions of that same call site within the same
  parent invocation.
- **D-15-03:** Always emit `#0` in the current acyclic, non-iterating language.
  The ordinal is reserved now so later repeated execution does not require a
  second identity-grammar change.
- **D-15-04:** `E` is a single canonical UTF-8 byte escaping rule: RFC 3986
  unreserved ASCII and `:` remain literal; every other byte is uppercase
  `%HH`. `/`, `#`, and `%` are therefore never ambiguous delimiters. `/2`
  validation rejects malformed and non-canonical spellings.
- **D-15-05:** Invocation identity is occurrence ancestry, not content
  identity. Do not hash event payloads or paths. An edit that changes an ID on
  the selected entry-to-call chain intentionally changes the invocation;
  unrelated edits do not.
- **D-15-06:** Static `Event.ID` derivation remains unchanged. `/2` uniqueness
  is the pair `(invocation, id)`; `/0` and `/1` retain ID-only uniqueness.
- **D-15-07:** Every successfully admitted executed `OpCall` emits exactly one
  caller-owned `function.called` event immediately before child activation.
  Resolution or depth failure emits no successful-call edge.
- **D-15-08:** The call event carries `id: <OpCall ID>:event:called`,
  `function_id: <caller function ID>`, `invocation: <parent invocation>`, and a
  required `callee_function_id: <resolved callee function ID>`, plus the
  existing `source_place`, `target_place`, and `type_id` facts from the
  operation. — **Reversibility: one-way** — field ownership and event ordering
  become part of the public `/2` protocol.
- **D-15-09:** Event order is strict depth-first preorder: the call event comes
  before the callee event subsequence; caller execution resumes afterward.
  Do not add paired call-completion/span events; `function.returned` already
  records completion.
- **D-15-10:** The removal negative control compares the `function.called`
  projection: removing one reachable call removes exactly its one call-edge
  record. Callee-local events may also disappear as the necessary dynamic
  consequence, so “exactly one event” must not be asserted over the unfiltered
  full document.
- **D-15-11:** A non-importing peer performs its own entry resolution and
  call-DAG unfolding. It validates canonical invocation membership,
  `(invocation,id)` uniqueness, invocation-to-function consistency,
  single-parent call-edge ownership, and stack/preorder discipline.
- **D-15-12:** General `/2` validation checks the exact observed causal
  structure but does not require equality with every statically admissible
  invocation. Static-set equality is valid only as a separately named control
  for straight-line/full-coverage fixtures; requiring it generally would
  reject future untaken branches.
- **D-15-13:** The differential is guarded in both directions. One seeded fault
  corrupts production invocation/edge derivation while the peer remains
  correct; another corrupts the peer acceptance/derivation seam while
  production remains correct, and the test fails if a formerly detected
  divergence silently resolves.
- **D-15-14:** Peer failures are named and actionable: report the violated
  class, event index, invocation, and expected parent/function where
  applicable. Unknown schema, missing required `/2` fields, malformed grammar,
  traversal-bound exhaustion, and unclassified event kinds fail closed.
- **D-15-15:** Native `emitProgram` uses a static invocation-path table and
  threads an unsigned table index through internal generated functions. It
  does not build paths at runtime, hash paths, truncate, or silently fall back.
- **D-15-16:** The fixed ceiling is 4096 invocation nodes. A node is one
  dynamically reachable activation occurrence represented by its complete
  static call-site chain, with the entry activation counted as node one.
- **D-15-17:** The 13-function `deep_diamond_acyclic.lang` fixture unfolds to
  61 nodes, not 13: `T0=1`, then `T1=5`, `T2=13`, `T3=29`, `T4=61` for its four
  stacked diamonds. The chosen ceiling is therefore 67.1 times the measured
  adversarial fixture.
- **D-15-18:** Count once after call-graph and entry validation but before path
  table allocation or C serialization. Use checked/saturating expansion and
  refuse on the attempted 4097th node.
- **D-15-19:** Refusal uses stable code
  `cgen.invocation_path_table_exceeded` and records the entry, `limit: 4096`,
  `observed_at_least: 4097`, and deterministic first-overflow invocation/call
  operation context. There is no flag, environment override, or adaptive
  source-size heuristic.
- **D-15-20:** Pin the 61-node measurement, test 4096 accepted and 4097
  refused, and mutation-kill removal/bypass of the preflight. A future emitted
  byte-size limit, if measurements justify one, is a separate bound with a
  separate diagnostic; it must not silently redefine the node contract.

### the agent's Discretion

Implementation factoring, helper names, and exact diagnostic prose remain at
the planner's discretion so long as the codes, wire fields, ordering, bounds,
and independent-derivation properties above are preserved.

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|---|---|---|
| OBS-01 | Shared-leaf activations have distinct identities. | `/2` pair uniqueness, invocation threading, flipped diamond control. |
| OBS-02 | Calls expose causal edges. | Caller-owned preorder `function.called` projection and removal control. |
| OBS-03 | Identity has an independent peer derivation. | Non-importing DAG unroller plus two-direction seeded faults. |
| OBS-04 | `/0` and `/1` bytes stay frozen; `/2` owns the field. | Schema-peek validation, legacy byte goldens, versioned C writers. |
| NAT-10 | Re-invoking fixture agrees on interpreter, `-O0`, `-O3`, `-O3 -flto`. | Existing four-tier harness becomes positive for `DiamondSharedLeaf`. |
</phase_requirements>

## Project Constraints (from AGENTS.md)

- Keep the Stage 0 host to Go 1.24 standard library; generated native code is readable C17 for installed Clang. [VERIFIED: AGENTS.md]
- Prefer standard-library-only, shallow audited boundaries; Phase 15 must add no dependency. [VERIFIED: AGENTS.md]
- Preserve deterministic fixtures and separate negative controls, properties, mutation, differential execution, and sanitizers into explicit cost lanes. [VERIFIED: AGENTS.md]
- Preserve macOS/Linux portability and do not encode current Apple arm64 host facts into wire contracts or target facts. [VERIFIED: AGENTS.md]
- Keep untrusted input, unsafe operations, FFI, secrets, and build authority explicit; no ambient-authority escape. [VERIFIED: AGENTS.md]
- This research is within the active GSD planning workflow; execution must likewise use the appropriate GSD entry point. [VERIFIED: AGENTS.md]

## Summary

Implement `/2` as an additive protocol branch: retain static `Event.ID`, add an `Invocation` occurrence path and `CalleeFunctionID` only for `/2`, and require `(Invocation, ID)` uniqueness. The existing model is a single shared `execution.Event` struct and has only `Schema0` and `Schema1`; `/1` validation currently rejects duplicate `ID` values globally. [VERIFIED: internal/compiler/execution/execution.go:8-11,41-57] [VERIFIED: internal/compiler/native/native.go:553-590]

The work has three coupled producers/consumers: interpreter frames calculate path/ordinal state at the `OpCall` push point; `emitProgram` preflights and serializes a static path table while threading a table index; a new non-importing peer reconstructs the same call-occurrence tree from `core.Program` and validates observed `/2` event causality. Do not place the peer in `interp` or `cgen`, and do not use a static admissible-set equality check as general `/2` validation. [VERIFIED: 15-CONTEXT.md]

**Primary recommendation:** Plan six ordered slices: frozen-wire contract and parser; independent invocation peer; interpreter producer; native static-table producer/preflight; comparator and Phase 11 four-tier cutover; then boundary/mutation/non-inertness evidence (with the peer and producer faults in opposite directions).

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|---|---|---|---|
| `/2` wire grammar and structural validation | API / Backend | — | The execution JSON schema/model and native decoder own protocol admission. |
| Invocation derivation during interpreted execution | API / Backend | — | `runFrameStack` owns the active call-frame stack and emits ordered events. [VERIFIED: internal/compiler/interp/interp.go:643-851] |
| Invocation derivation during C emission | API / Backend | Native runtime | Go emission builds tables; generated C only carries indices and records literal paths. |
| Causal-document peer validation | API / Backend | — | A separate Go package traverses `core.Program`, never engine output derivation helpers. |
| Four-tier agreement | API / Backend | Native runtime | Session runs interpreter and native O0/O3/LTO, then compares every engine pair. [VERIFIED: internal/compiler/session/session_phase11_differential_test.go:73-147] |

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---|---|---|---|
| Go standard library | Go 1.24.0 available | Model, JSON, deterministic tables, tests | Project constraint; no dependency is warranted. [VERIFIED: environment probe] |
| Installed Clang | Apple Clang 21.0.0 available | Compile generated C at O0/O3/LTO | Existing native differential mechanism. [VERIFIED: environment probe] |

**Installation:** none. No external package or Package Legitimacy Audit applies.

## Architecture Patterns

### System Architecture Diagram

```text
checked core.Program
  ├─ entry + ordered OpCall traversal ──> invocation-tree peer ──> /2 structural verdict
  ├─ interp frame stack ──> invocation + function.called ──> execution /2
  └─ cgen preflight/unfold ──> static literal paths + child-index tables ──> C17 ──> execution /2
                                                                        │
interpreter / O0 / O3 / O3-LTO documents ──> Phase5CompareEngines ────┘
```

### Required Producer Pattern

1. Add `Schema2`, `Event.Invocation`, and `Event.CalleeFunctionID`; use omission for legacy marshalling but `/2` validation must require non-empty `Invocation` and require `CalleeFunctionID` for `function.called`. This is necessary because a non-omitempty new field would alter `/0`/`/1` canonical bytes. The current struct has no version-specific encoding layer and `CanonicalBytes` directly calls `json.Marshal`. [VERIFIED: internal/compiler/execution/execution.go:41-57,83-88]
2. Centralize canonical byte escaping and grammar parse/format in a new small protocol package or `execution` helper that neither imports `interp` nor `cgen`. The peer must use the parser, while each producer uses the formatter. Do not let the peer call production unfolding/emission helpers. [VERIFIED: 15-CONTEXT.md]
3. Make the call event before `partitionFrameForCall` mutates/returns a child only after the callee resolves and the depth admission succeeds. The current arm increments the caller index, calls `partitionFrameForCall`, and only then checks `maxCallDepth`; emitting beforehand would wrongly report an edge for resolution/depth failure. [VERIFIED: internal/compiler/interp/interp.go:787-825]
4. Pass invocation through every event constructor (including immediate match-arm return, transition events, terminal events, leak/nonlocal/depth events) that can participate in a `/2` execution. The current helpers derive static IDs from operation IDs, so merely changing returns leaves transition collisions. [VERIFIED: internal/compiler/interp/interp.go:623-641,691-851]
5. In C, produce a literal path table plus static per-call-site child-index tables. A single constant child index at a source call site is unsound when that caller function itself has multiple invocation contexts; select child index from the parent `lang_inv` with a generated static table. This still meets the locked no-runtime-path-construction rule and is auditable C. [VERIFIED: 15-CONTEXT.md] [ASSUMED]

### Native Preflight/Emission Pattern

Run `callgraph.Order` and `callgraph.EntryFunction`, then expand nodes in deterministic function-operation order. `EntryFunction` derives an in-degree-zero root and uses reachable-closure size to resolve multiple candidates; `Order` uses deterministic sorted graph traversal and distinguishes shared black revisits from gray cycles. [VERIFIED: internal/compiler/callgraph/callgraph.go:243-330,449-500]

The node builder must maintain `{index, functionID, invocation, parentIndex, incomingCallID}` and stop before appending the 4097th node. Record the first attempted child context in a typed error with code `cgen.invocation_path_table_exceeded`; use no test-only production override. `deep_diamond_acyclic.lang` explicitly stacks four diamonds; its declared functions/calls yield the locked recurrence `T0=1, T1=5, T2=13, T3=29, T4=61`. [VERIFIED: testdata/phase07/deep_diamond_acyclic.lang:1-87] [VERIFIED: 15-CONTEXT.md]

Generated function prototypes currently take exactly one value argument, and `emitProgramFunction` records all transitions/returns into one global event buffer. Extend every internal generated prototype/definition/call to accept an unsigned invocation index, extend `LANG_EVENT` and `lang_record_event`/writer to carry invocation and optional callee ID, and write literal `lang.execution/2` only on the multi-function `/2` path. [VERIFIED: internal/compiler/cgen/cgen_program.go:236-245,329-435] [VERIFIED: internal/compiler/cgen/cgen.go:1667-1713]

### Independent Peer Pattern

Create a dedicated peer package with a structural import test (`go list -deps` or equivalent) proving it imports neither `internal/compiler/interp` nor `internal/compiler/cgen`. It independently resolves entry, indexes functions/operations, expands the bounded call tree, parses canonical invocations, and consumes the document sequentially with an explicit activation stack. [VERIFIED: 15-CONTEXT.md]

For each event, validate schema/required fields, canonical membership, `(invocation,id)` uniqueness, invocation-implied function identity, and kind classification. For `function.called`, validate caller invocation/function, resolved `callee_function_id`, call-site event ID, exactly one child ownership, and that the next child subsequence is preorder-nested. Permit an admissible invocation not to occur in general validation; reserve static-set equality for a named straight-line/full-coverage control. [VERIFIED: 15-CONTEXT.md]

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---|---|---|---|
| Event occurrence identity | Payload/path hash or global counter | Locked readable ancestry grammar | Hash collides for content-identical leaf activations; counter is not independently derivable. [VERIFIED: EMISSION-AND-EVENT-IDENTITY.md:343-350] |
| Entry resolution | `Functions[0]` or an engine-local convention | Independent reproduction of the documented in-degree/closure algorithm | Entry is not source-order-defined. [VERIFIED: internal/compiler/callgraph/callgraph.go:243-330] |
| Native path strings | `snprintf`, runtime concatenation, truncation | Emitter-built static literal path and child-index tables | Preserves deterministic, readable `/2` values and bound-before-allocation. [VERIFIED: 15-CONTEXT.md] |
| Peer validation | Reuse interpreter/cgen helpers | Separate traversal/state machine | Reusing production derivation invalidates OBS-03. [VERIFIED: 15-CONTEXT.md] |

## Common Pitfalls

1. **Breaking frozen schemas.** Adding a non-omitempty field to the shared model changes JSON bytes even for `/0` and `/1`; preserve legacy omission and keep ID-only uniqueness for those schemas. [VERIFIED: internal/compiler/execution/execution.go:41-57,83-88] [VERIFIED: 15-CONTEXT.md]
2. **Only fixing returns.** `ownedEvent` also derives static `operation.ID + ":event"`; a re-invoked callee with transitions still collides unless every `/2` event gets its frame invocation. [VERIFIED: internal/compiler/interp/interp.go:846-851]
3. **Call-edge timing wrong.** Emit only after callee resolution and depth admission; no edge for a rejected push. [VERIFIED: internal/compiler/interp/interp.go:787-825] [VERIFIED: 15-CONTEXT.md]
4. **Using the full-event removal assertion.** Removing a call also removes its descendants; compare only the `function.called` projection. [VERIFIED: 15-CONTEXT.md]
5. **Treating call-site index as globally constant in generated C.** Shared functions can execute in several parent contexts. Child selection must be indexed by parent invocation, or code must be cloned per occurrence; plan the former static-table shape. [ASSUMED]
6. **Unbounded/preallocated expansion.** Count/append in checked order and refuse the attempted 4097th node before allocating/serializing the complete table. [VERIFIED: 15-CONTEXT.md]
7. **One-way mutation proof.** A producer fault alone cannot establish peer independence; seed both producer corruption and peer acceptance corruption, with restored control required to detect the divergence. [VERIFIED: 15-CONTEXT.md]

## Code Examples

The public `/2` protocol skeleton must preserve the following locked values verbatim: `"lang.execution/2"`, `"inv:entry:{E(entryID)}(/{E(opCallID)}#{ordinal})*"`, `"function.called"`, `"cgen.invocation_path_table_exceeded"`, `"limit: 4096"`, and `"observed_at_least: 4097"`. [VERIFIED: 15-CONTEXT.md]

```go
// Producer-only shape; parser/peer own a separate derivation.
childInvocation := parentInvocation + "/" + escape(operation.ID) + "#0"
events = append(events, Event{
    Schema: execution.Schema2, ID: operation.ID + ":event:called",
    Kind: "function.called", FunctionID: caller.function.ID,
    Invocation: parentInvocation, CalleeFunctionID: callee.ID,
    SourcePlace: operation.SourceID, TargetPlace: operation.TargetID, TypeID: operation.TypeID,
})
```

## Validation Architecture

### Test Framework

| Property | Value |
|---|---|
| Framework | Go standard `testing` package, Go 1.24.0 available [VERIFIED: environment probe] |
| Config file | none |
| Quick run command | `go test ./internal/compiler/session -run 'TestPhase15DiamondFrontierMoved' -count=1` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|---|---|---|---|---|
| OBS-01 | Diamond has distinct `(invocation,id)` and restored collision fails. | four-tier + mutation | `go test ./internal/compiler/session -run '^TestPhase11InterproceduralDifferential$' -count=1` | ✅ target test exists; expectation must flip [VERIFIED: internal/compiler/session/session_phase11_differential_test.go:300-340] |
| OBS-02 | One admitted call emits preorder `function.called`; call removal changes only call projection by one. | unit + fixture differential | focused new peer/interp/cgen tests | ❌ Wave 0 |
| OBS-03 | Independent peer accepts valid documents and catches producer/peer seeded faults in both directions. | peer + mutation | focused new peer package tests | ❌ Wave 0 |
| OBS-04 | Legacy bytes/semantics frozen; `/2` required fields, grammar and pair uniqueness enforced. | model/native golden | `go test ./internal/compiler/{execution,native} -count=1` | ✅ native validation suite exists; `/2` cases are Wave 0 [VERIFIED: internal/compiler/native/native_test.go:886-924] |
| NAT-10 | Re-invoking fixture compares all four tiers. | integration | `go test ./internal/compiler/session -run '^TestPhase11InterproceduralDifferential$' -count=1 -v` | ✅; currently preserves the negative control [VERIFIED: test run 2026-09-19] |

### Required Non-Inertness Controls

- Pin the 61-node measurement from the real fixture and synthesize 4096/4097 node programs or core fixtures; prove a bypassed/disabled preflight reaches the wrong path, while restored preflight returns `cgen.invocation_path_table_exceeded`.
- Producer fault: alter only production invocation/call-edge derivation; peer must reject while its derivation stays correct.
- Peer fault: alter peer grammar/ownership acceptance or derivation; a deliberately corrupted document must become wrongly accepted, and restored peer must reject it. This is the bidirectional control required by D-15-13, not two copies of one producer test. [VERIFIED: 15-CONTEXT.md]
- Retain and flip `DiamondSharedLeaf`; do not delete its current documented-collision branch. The current test intentionally accepts the old failure at all native optimizations. [VERIFIED: internal/compiler/session/session_phase11_differential_test.go:300-340]

### Wave 0 Gaps

- [ ] Invocation grammar canonical/non-canonical byte cases, including UTF-8 escaping, `%`/`/`/`#`, empty fields, and malformed ordinal.
- [ ] Legacy `/0` and `/1` canonical-byte corpus pin before producer changes.
- [ ] Peer structural import guard and named actionable error assertions.
- [ ] Static-table 61/4096/4097 boundary and preflight-bypass mutation proof.
- [ ] `function.called` projection removal control and strict preorder assertion.

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---|---|---|
| V2 Authentication | no | No identity is an authentication credential. |
| V3 Session Management | no | Invocation is deterministic execution evidence, not session state. |
| V4 Access Control | no | No authorization boundary is introduced. |
| V5 Input Validation | yes | Fail closed on unknown schema/kind, missing `/2` fields, malformed/non-canonical grammar, and table-bound exhaustion. [VERIFIED: 15-CONTEXT.md] |
| V6 Cryptography | no | Do not hash paths or introduce cryptography. [VERIFIED: 15-CONTEXT.md] |

| Threat Pattern | STRIDE | Standard Mitigation |
|---|---|---|
| Path-table expansion DoS | Denial of service | Checked/saturating preflight, refusal on node 4097, no override. [VERIFIED: 15-CONTEXT.md] |
| Ambiguous delimiter/noncanonical path | Tampering | One byte escape/strict parser; reject alternate spellings. [VERIFIED: 15-CONTEXT.md] |
| Forged causal edge | Tampering | Peer checks callee, parent ownership, function consistency and preorder. [VERIFIED: 15-CONTEXT.md] |
| Silent evidence weakening | Repudiation | Legacy byte pins, comparator routing, flipped Phase 11 control, and bidirectional fault tests. [VERIFIED: 15-CONTEXT.md] |

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|---|---|---|---|---|
| Go | compiler/tests | ✓ | go1.24.0 | — |
| Clang | native O0/O3/LTO differential | ✓ | Apple Clang 21.0.0 | — |

No external services, package managers, or new dependencies are required.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|---|---|---|
| A1 | A per-parent static child-index table is required unless native function bodies are cloned per invocation occurrence. | Architecture Patterns / Pitfall 5 | Generated C could select the wrong child identity for shared/re-invoked callers. |

## Open Questions

1. **Exact legacy producer scope for `/2`.**
   - What we know: `/0` and `/1` bytes/semantics are frozen; Phase 15's native work is `emitProgram`, while existing single-function writers hardcode legacy schema strings. [VERIFIED: internal/compiler/cgen/cgen_program.go:291-299] [VERIFIED: internal/compiler/cgen/cgen.go:1702-1712]
   - What's unclear: whether interpreter `/2` is selected only for multi-function executions or via an explicit versioned run mode.
   - Recommendation: make selection explicit in the execution producer API and pin both legacy and multi-function behavior before editing; do not implicitly upgrade all legacy emitters.

2. **Peer package home.**
   - What we know: the peer must not import either engine, and `session` already owns cross-engine comparison. [VERIFIED: 15-CONTEXT.md]
   - Recommendation: use a new small `internal/compiler/executionpeer` package with an import guard, then let `session` invoke it; avoid a peer embedded in `session` if that would couple it to engine imports. [ASSUMED]

## Sources

### Primary (HIGH confidence)

- [15-CONTEXT.md](15-CONTEXT.md) — locked grammar, event shape/order, peer, bounds, and required controls.
- [ROADMAP.md](../../ROADMAP.md) — Phase 15 success criteria and requirement ownership.
- [execution model](~/projects/ai-lang/internal/compiler/execution/execution.go) — current schemas and canonical JSON model.
- [interpreter](~/projects/ai-lang/internal/compiler/interp/interp.go) — frame/call/event sequencing.
- [native validator](~/projects/ai-lang/internal/compiler/native/native.go) — legacy validation and uniqueness seam.
- [native emitter](~/projects/ai-lang/internal/compiler/cgen/cgen_program.go) — multi-function generation seam.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — local Go/Clang probes and project constraints.
- Architecture: HIGH — locked decisions plus source-of-truth code paths; one C child-index implementation detail is explicitly assumed.
- Pitfalls: HIGH — directly traceable to current frame, schema, C event, and Phase 11 test behavior.

**Research date:** 2026-09-19
**Valid until:** phase implementation begins; decisions are locked project-local authority.
