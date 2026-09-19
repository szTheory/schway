# Phase 15: Event Identity (`lang.execution/2`) - Context

**Gathered:** 2026-09-19
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 15 gives every dynamically reached function activation a deterministic,
human-auditable identity under `lang.execution/2`; records every executed
`OpCall` as an observable caller-to-callee edge; independently validates the
observed invocation structure; and bounds native call-DAG unfolding before C
emission. It makes the shared-leaf diamond comparable across interpreter,
`-O0`, `-O3`, and `-O3 -flto` without changing frozen `/0` or `/1` bytes.

This phase does not port branch/match emitters, add loops or recursion, alter
static event-ID derivation, introduce runtime tracing infrastructure, or add a
dependency.

</domain>

<decisions>
## Implementation Decisions

### Invocation Path Grammar

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

### Observable `OpCall` Event

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

### Independent Peer Validation

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

### Static Native Path Table

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

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Milestone Authority and Phase Contract

- `.planning/ROADMAP.md` §Phase 15 — goal, five success criteria, measured-risk
  gate, ordering, and closure targets.
- `.planning/REQUIREMENTS.md` §Observability / Native Emission — OBS-01 through
  OBS-04 and NAT-10.
- `.planning/PROJECT.md` — current milestone thesis, runtime posture, known
  event-identity debt, and explicit exclusions.
- `.planning/STANDING-VERDICTS.md` — hard constraints, dependency verdicts,
  anti-features, and evidence rules that apply to every M003 phase.
- `.planning/LANGUAGE-MATURITY.md` — actual expressible surface; especially no
  loops, recursion, or branch on a computed value.
- `.planning/research/M003/ADVERSARIAL-SYNTHESIS.md` — authoritative M003
  contradiction resolutions; its Phase 15 charter overrides conflicting
  source research.
- `.planning/research/M003/EMISSION-AND-EVENT-IDENTITY.md` §Cluster B — source
  analysis, context-path grammar, static-table alternatives, and peer
  re-derivation design considered in this discussion.

### Historical Debt and Fixtures

- `.planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/PHASE-11-DEBT.md`
  — D-11-51 shared-leaf collision record closed by this phase.
- `.planning/milestones/M002-phases/12-result-payloads/PHASE-12-DEBT.md` —
  D-12-21 event-identity continuation closed by this phase.
- `testdata/phase11/multi_function_diamond_call.lang` — executable shared-leaf
  collision fixture and four-tier acceptance gate.
- `testdata/phase07/deep_diamond_acyclic.lang` — 13-function/four-diamond
  ceiling measurement fixture; expected unfolded invocation count is 61.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- `internal/compiler/execution/execution.go`: closed schema constants and the
  canonical `Execution`/`Event` JSON model; add `/2`, `Invocation`, and the
  call-edge callee field without altering legacy canonical bytes.
- `internal/compiler/interp/interp.go`: explicit frame stack and dedicated
  `OpCall` arm already provide the parent frame, call operation, and child push
  point needed for invocation threading and preorder call events.
- `internal/compiler/cgen/cgen_program.go`: the multi-function emitter already
  resolves calls within one translation unit and centralizes per-function
  event emission; it is the integration point for preflight unfolding, the
  literal path table, threaded indices, and call events.
- `internal/compiler/session/session_phase5_compare.go`: fail-closed field
  routing and five-axis comparison pattern; `Invocation` and
  `CalleeFunctionID` require deliberate routing.
- `internal/compiler/session/session_phase11_differential_test.go`:
  `DiamondSharedLeaf` contains the inverted assertion that must be flipped,
  not deleted.
- `internal/compiler/pathoracle/pathoracle.go`: `MaxPaths` and named,
  fail-closed bounded enumeration are the nearest precedent, though its CFG
  path bound is semantically distinct from the whole-program invocation cap.

### Established Patterns

- Legacy schema bytes are frozen across bumps; schema-peek dispatch preserves
  old decode/validation behavior.
- Independent peers re-derive facts from `core.Program` and carry structural
  import guards rather than trusting producer artifacts.
- Every new guard needs a non-inertness proof; a green path alone is
  insufficient after Phase 14.
- Closed vocabularies, named diagnostics, deterministic fixtures, exact
  comparator field routing, and bidirectional set/differential checks are
  preferred over permissive fallbacks.
- Generated C stays readable and optimizer-independent; the threaded value is
  a compile-time table index inside one translation unit.

### Integration Points

- Interpreter frame creation, event helpers, and `OpCall` push logic.
- `/2` JSON decoding and native `validateExecution` uniqueness checks.
- `emitProgram` preflight and generated function signatures/call sites.
- Session comparator field routing, four-tier differential, and peer gate.
- Fixture-based boundary, removal, malformed-wire, and seeded-mismatch tests.

</code_context>

<specifics>
## Specific Ideas

- The user requested maximum-breadth adversarial consideration across security,
  product, architecture, compiler/runtime, C/C++, distributed systems,
  observability, API design, DevOps, testing, and human-audit lenses, followed
  by decisive one-shot recommendations. Four `gsd-advisor-researcher` fan-outs
  independently examined the four decision areas before synthesis.
- All four recommendations were accepted together as a coherent set. The
  unifying rule is that occurrence identity and causality must remain readable,
  independently derivable, bounded before allocation, and impossible to report
  green through omission.
- The ceiling measurement is a planning input, not a guessed constant:
  13 declarations unfold to 61 activation-path nodes.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 15-event-identity-lang-execution-2*
*Context gathered: 2026-09-19*
