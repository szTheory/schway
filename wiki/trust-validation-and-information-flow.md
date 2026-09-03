---
id: trust-validation-information-flow
title: Trust, validation, and information-flow boundary
summary: A candidate hybrid that preserves Perl taint mode's source-to-sink insight while separating provenance, validity, encoding, confidentiality, and authority.
type: design
status: candidate
confidence: medium
created: 2026-09-03
updated: 2026-09-03
tags: [security, taint, information-flow, validation, ai]
related: [design-baseline, semantic-kernel-contract, semantic-kernel-probes, boundary-data-validation-persistence, type-system-failure-semantics, effects-and-capabilities, context-telemetry-security, ai-native-runtime-evals, native-low-level-profile, compiler-feedback-latency, convergence-audit, residual-uncertainty-register, research-ledger, open-questions]
---

# Trust, validation, and information-flow boundary

## Reader and outcome

This is for the type-checker, security-tooling, boundary-kit, and AI-runtime
contributors. After reading it, they should be able to distinguish five
different security questions, implement the minimum v0 hooks without a second
general-purpose type system, and build probes that reveal whether stronger
information-flow checking earns its cost.

## Verdict

Perl taint mode identified a real and still-important invariant:

```text
externally influenced data must not accidentally control a dangerous sink
```

Lang should adopt that invariant, but not Perl's single hidden runtime taint
bit. The candidate is a layered hybrid:

1. Dangerous APIs accept structured or nominal sink-specific values rather
   than arbitrary strings.
2. Boundary data begins in an explicit input/provenance wrapper and reaches
   domain code through typed parsing and validation.
3. Confidentiality and secrets remain separate from input integrity.
4. Capabilities—not data labels—grant authority.
5. The typed IR records source, transformation, validation, encoding, sink,
   endorsement, and declassification summaries.
6. Fast local flow checks run during `check`; targeted interprocedural and
   control-influence checks run during `verify` and `release`.
7. Optional dynamic shadow tracking belongs in adversarial test builds, not in
   every production scalar.

This uses the existing semantic kernel—nominal types, opaque constructors,
generics, effects, capabilities, `Result`, modules, and evidence. V0 needs a
stable flow-summary hook in typed IR, but no new taint operator, implicit cast,
runtime header bit, or general security lattice in core syntax.

## What Perl proved—and where it stopped

Perl marks many external values, conservatively propagates taint through
expressions, and prevents tainted values from reaching selected process and
filesystem operations. Its documentation also records the limitations:

- some sinks such as `print`, `syswrite`, symbolic methods, and hash keys are
  not checked;
- a whole expression may become tainted after merely accessing tainted data;
- taint support has runtime cost even when the feature is unused;
- a regex capture can clear taint, but Perl does not establish that the regex
  is a sufficient validator;
- taint checking is explicitly a guard against mistakes, not a replacement for
  security reasoning.

That is a valuable historical result. It demonstrates that language-supported
source-to-sink checks can prevent accidents, while a Boolean `dirty/clean`
state cannot explain *safe for what*, *validated by which rule*, *secret to
whom*, or *authorized to do what*. ([Perl security](https://perldoc.perl.org/perlsec))

## Five axes that must not collapse into `Trusted`

| Axis | Question | Candidate representation | Example |
|---|---|---|---|
| Influence/integrity | Who or what may have affected this value? | `Input<Origin, T>` plus IR provenance | HTTP body, file, model output, foreign callback |
| Semantic validity | Which predicate or domain invariant was established? | opaque nominal type or `Verified<Rule, T>` | `Email`, `OrderId`, `Verified<UploadPolicy, File>` |
| Interpretation context | In which grammar or sink position is this value safe? | structured API or context-specific type | `HtmlText`, `UrlSegment`, `SqlIdentifier`, `OsArg` |
| Confidentiality | Who may observe, retain, or export it? | `Secret<T>` and classification metadata | access token, payment data, tenant identifier |
| Authority | What effects may code perform? | scoped capability/provider | read customer data, execute process, capture payment |

Two further facts attach to transitions rather than values:

- **Evidence:** which validator, policy version, principal, test, or approval
  justified a transition;
- **Cost:** bounds on input size, parsing work, allocation, recursion, model
  tokens, or external calls.

Examples of why one bit fails:

- A syntactically valid email remains untrusted as proof of account ownership.
- HTML-escaped text is not safe as a JavaScript program or URL scheme.
- Parameterized SQL safely transports a hostile string as data without making
  that string generally trusted.
- A secret may be fully validated but still forbidden from telemetry.
- A schema-valid model-produced tool call still has no permission to execute.

OWASP requires output handling specific to HTML body, attribute, JavaScript,
CSS, and URL contexts and warns that generic response interceptors can apply
the wrong or duplicate encoding. W3C Trusted Types similarly protects specific
DOM injection sinks with unforgeable typed values rather than blessing all
strings. ([OWASP XSS prevention](https://cheatsheetseries.owasp.org/cheatsheets/Cross_Site_Scripting_Prevention_Cheat_Sheet.html),
[W3C Trusted Types](https://www.w3.org/TR/trusted-types/))

## Candidate boundary model

### External data enters explicitly

```text
let body: Input<HttpRequest, Bytes> = Http.body(request:)

let command = PlaceOrder.decode(input: body)?
place_order(command: command)
```

`Input<Origin, T>` is written here as an opaque source-level qualifier whose
provenance fact survives in typed IR while its representation is erased in
ordinary compiled profiles. That zero-metadata representation is a design
requirement to test, not permission for the optimizer to erase security
checks. The qualifier permits inspection, bounded parsing, and pure mapping,
but it does not implicitly convert to domain identities, code, paths,
capabilities, or sink-specific representations.

Standard combinators preserve influence:

```text
Input.map(input: body, with: Bytes.trim)
  : Input<HttpRequest, Bytes>

Json.decode(input: body, limits: JsonLimits.default)
  : Result<Input<HttpRequest, JsonValue>, JsonError>
```

Decoding establishes a structural fact. It does not assert that the resulting
business command is permitted, belongs to the current tenant, or came from a
high-integrity principal.

### Validation produces a specific fact

```text
fn PlaceOrder.decode(input: Input<HttpRequest, Bytes>)
  -> Result<PlaceOrder, RequestProblem>
{
  let fields = Json.decode(input:, limits: OrderJson.limits)?

  PlaceOrder.validate(
    customer: CustomerId.parse(input: fields.customer_id)?,
    lines: OrderLines.parse(input: fields.lines)?,
  )
}
```

The private constructor of `PlaceOrder` is the proof boundary. A reusable
policy may instead return `Verified<PlaceOrderRuleV3, CandidateOrder>`. The
compiler records the validator identity and policy version, but ordinary code
does not carry a tower of visible wrappers once a nominal domain type is
constructed.

There is no generic `trust`, `clean`, or `sanitize` operation. A validator
establishes only its declared postcondition. Changing that postcondition or
policy version participates in dependency impact and selects affected tests.

### Prefer safe sinks over sanitizers

```text
Html.element("p") {
  Html.text(value: customer_name)       // encodes as text for this node
}

Database.query(
  statement: sql"SELECT * FROM orders WHERE id = {order_id}",
)                                       // value becomes a protocol parameter

Process.run(
  executable: Tool.git,
  arguments: [OsArg("show"), OsArg(revision)],
)                                       // no shell parsing
```

The ordinary APIs make the data/code distinction structural. Raw HTML,
dynamic SQL identifiers, shell programs, format strings, regular expressions,
network destinations, and foreign symbols use narrower constructors or
capabilities because each can change interpretation or authority.

OWASP recommends prepared/parameterized SQL and recommends avoiding direct OS
commands where a language API exists; when process execution is necessary,
argument separation and validation are distinct defenses. ([OWASP SQL
injection prevention](https://cheatsheetseries.owasp.org/cheatsheets/SQL_Injection_Prevention_Cheat_Sheet.html),
[OWASP command injection defense](https://cheatsheetseries.owasp.org/cheatsheets/OS_Command_Injection_Defense_Cheat_Sheet.html))

### Endorsement and declassification are effects

An integrity upgrade and a confidentiality downgrade are security decisions,
not casts:

```text
let approved = Integrity.endorse(
  proposal: tool_proposal,
  under: RefundToolPolicy.v4,
  evidence: approval,
)?

let public_receipt = Privacy.declassify(
  value: restricted_receipt,
  purpose: CustomerReceipt,
  fields: ReceiptDisclosure.allowed,
)?
```

These operations require capabilities, emit mandatory audit evidence, name the
policy and purpose, and return a nominal result. They cannot be supplied by an
untrusted package or inferred from a remote annotation.

Jif's information-flow model separates confidentiality from integrity and
names declassification and endorsement as distinct authority-sensitive
operations. Lang borrows the separation, not Jif's complete principal lattice
or pervasive source syntax in v0. ([Jif decentralized label
model](https://www.cs.cornell.edu/jif/doc/jif-3.0.0/dlm.html),
[Programming in Jif](https://www.cs.cornell.edu/jif/doc/jif-2.0.0/jif_programming.html))

## AI-specific application

AI makes Perl's original concern more important because natural-language data
and instructions occupy the same representation.

```text
let retrieved: Input<OpenWorld, Text> = Search.fetch(query: query)?
let output: Input<ModelProvider, Text> = Model.generate(
  policy: RefundAssistant,
  evidence: retrieved,
)?

let proposed = ToolProposal.decode(input: output)?
let authorized = RefundPolicy.authorize(
  proposal: proposed,
  principal: command.actor,
  budget: command.budget,
)?

Tools.invoke(call: authorized)
```

The transitions are intentionally different:

1. Retrieval and model generation produce low-integrity input.
2. Schema decoding establishes shape only.
3. Application policy checks action, target, principal, budget, freshness, and
   current state.
4. The capability boundary performs the effect.

Model text, retrieved documents, tool results, MCP descriptions, and generated
code never mint capabilities, select an unapproved provider, lower data
classification, or declare themselves safe. MCP requires clients to treat tool
annotations as untrusted unless the server is trusted; NIST defines prompt
injection around untrusted input entering a higher-trust prompt and documents
indirect agent hijacking through external resources. ([MCP tools](https://modelcontextprotocol.io/specification/2025-06-18/server/tools),
[NIST prompt injection](https://csrc.nist.gov/glossary/term/prompt_injection),
[NIST agent-security findings](https://www.nist.gov/blogs/caisi-research-blog/insights-ai-agent-security-large-scale-red-teaming-competition))

This architecture reduces the blast radius of prompt injection; it does not
prove that a model understands which text is an instruction. The decisive
security boundary remains typed authorization, constrained capabilities,
least privilege, approvals for material actions, result validation, and
adversarial evaluation.

## Information-flow analysis tiers

### `check`: local and compositional

The fast lane uses declared summaries and local flow:

- preserve source influence through values, fields, collections, matches, and
  standard pure transformations;
- require nominal validators and sink-specific values;
- reject direct source-to-dangerous-sink paths;
- enforce secret/classification restrictions at telemetry, serialization, and
  network boundaries;
- infer local facts but require public source/sink/validator summaries;
- render the shortest source → transform → sink path and a typed repair.

Local analysis is compatible with the project's latency goal. CodeQL's own
documentation describes local flow as sufficiently fast and precise for many
queries while warning that whole-program data-flow graphs can be large and
slow. ([CodeQL data-flow analysis](https://codeql.github.com/docs/writing-codeql-queries/about-data-flow-analysis/))

### `verify`: targeted global and control influence

The stronger lane follows summaries across components and dependencies for
named policies. It can model:

- field-sensitive and partially validated aggregates;
- storage, serialization, caches, queues, actors, tasks, and callbacks;
- selected control dependencies where untrusted data chooses a dangerous
  action even if its bytes do not reach the sink;
- policy-version drift and missing foreign/dependency summaries;
- confidentiality flows to logs, model providers, analytics, and network
  destinations;
- integrity flows into executable selection, authorization, provider choice,
  code generation, deployment, and irreversible writes.

```text
let executable = match request.mode {
  "admin" => Tool.admin,
  _ => Tool.viewer,
}

Process.run(executable:, arguments: [])
```

The executable contains no request bytes, but the request controls which
authority is exercised. Flow Caml demonstrates how a program-counter label can
track such implicit flows and also why full information-flow checking is a
language-wide commitment: its design extends typing across expressions,
references, exceptions, modules, and function effects. That machinery must
earn its ordinary-language cost rather than enter v0 by aspiration. ([Flow
Caml overview](https://pauillac.inria.fr/~simonet/publis/simonet-flowcaml-nutshell.pdf))

### `release`: complete policy and evidence

Release verification additionally requires:

- complete public and foreign flow summaries or an explicit expiring waiver;
- dependency and generated-code provenance;
- mutation tests that remove or weaken validators;
- adversarial source-to-sink fixtures;
- runtime evidence where static modeling is incomplete;
- review of every endorsement, declassification, raw sink, and assumption;
- target/profile agreement and a schema-versioned machine report.

### Instrumented and dynamic builds

Optional shadow labels can reveal native, foreign, reflection, plugin, or
schedule behavior missed by static models. They are a test/fuzz mode with
measured overhead and complete instrumentation requirements, not a production
semantic dependency. Clang's DataFlowSanitizer is instructive: it provides
general dynamic labels, but uninstrumented functions require explicit ABI
modeling and fully instrumented C++ libraries must be rebuilt. ([Clang
DataFlowSanitizer](https://clang.llvm.org/docs/DataFlowSanitizer.html))

## Flow summaries and foreign code

Every public operation may expose a compiler-checked summary:

```text
flow summary decode_header {
  source input -> result.value
  validates Utf8 & HeaderGrammarV2
  preserves classification
  bounded by limits
}

foreign fn legacy_escape(input: borrow CChar) -> OwnedCString
  unsafe
  flow { input -> result }
```

The compiler derives ordinary Lang summaries where possible. Foreign,
reflection, dynamic plugin, procedural generation, and opaque dependency
summaries are claims with provenance and evidence. Missing summaries default
to conservative propagation, never silent cleanliness.

Clang and CodeQL both model taint with sources, propagators, sinks, and explicit
filters or barriers. Their documentation makes the domain-modeling burden
visible: library behavior unavailable to the analyzer must be summarized, and
partial validation may require several independent flow facts rather than one
flag. ([Clang taint configuration](https://clang.llvm.org/docs/analyzer/user-docs/TaintAnalysisConfiguration.html),
[CodeQL flow state](https://codeql.github.com/docs/codeql-language-guides/using-flow-labels-for-precise-data-flow-analysis/))

## Composition with the rest of Lang

| Existing mechanism | Trust-flow role |
|---|---|
| Nominal types | Represent domain validity, sink contexts, and authorization tokens |
| Structural ports | Carry flow summaries, but never grant trust or authority by shape alone |
| Effects/capabilities | Control who may validate against external state, endorse, declassify, or invoke a sink |
| Ownership/borrowing | Preserve facts through moves and borrows and reduce alias-analysis ambiguity |
| `Result` | Makes failed parsing, validation, authorization, and policy transitions explicit |
| Pattern matching | Refines variant-specific facts without arbitrary sanitizer callbacks |
| Modules/components | Separate ingress, validation, policy, domain, and egress boundaries |
| Context | Carries verified execution facts only within declared propagation rules |
| Telemetry | Records structural flow decisions while enforcing confidentiality and retention |
| Caching | Includes authority, policy, classification, and validation version in key/eligibility decisions |
| Serialization | Re-establishes facts from a declared schema; hidden in-memory labels never magically cross a wire |
| Concurrency | Moves/copies/share handles preserve facts; actor/task boundaries use declared schemas |
| Contracts/evidence | Generate adversarial cases and prove which transition or sink rule was exercised |

## Anti-patterns and footguns

- One `trusted` Boolean that conflates validity, confidentiality, and authority.
- A universal sanitizer or `untaint()` cast.
- Treating regex success, JSON validity, escaping, authentication, or model
  confidence as general trust promotion.
- Encoding at ingress rather than at the destination context.
- Treating an internal database, queue, cache, tool server, or model as trusted
  merely because it is inside the network.
- Making every generic function state every possible source label explicitly.
- Whole-program implicit-flow checking on every keystroke.
- Ignoring control influence at authorization or executable-action sinks.
- Silent loss of labels through serialization, FFI, reflection, or dependency
  calls.
- Allowing a package to declare itself a sanitizer or endorsement authority.
- Using dynamic labels as a substitute for static sink types and capability
  isolation.
- Logging raw source-to-sink paths when they contain secrets or PII.
- Claiming noninterference while ignoring timing, resource, termination, or
  foreign side channels.
- Treating taint analysis as proof that the application's security policy is
  correct.

## Why this remains a small kernel

“Small” should mean a small number of orthogonal semantic laws and a small
trusted checker, not an artificially tiny standard library or weak safety
posture. Lean is useful precedent: rich elaboration and automation produce
artifacts checked by a smaller kernel that omits unification and other
elaboration machinery. The architectural benefit is reduced trusted surface
and independent checking—not an automatic theorem that every compile or
program runs faster. ([Lean elaboration and kernel](https://lean-lang.org/doc/reference/latest/Elaboration-and-Compilation/))

For Lang:

- sink-specific nominal types reuse the ordinary type checker;
- source/sink summaries reuse stable typed IR and incremental queries;
- capabilities reuse the effect/provider system;
- stronger flow analysis is targeted and cacheable;
- dynamic tracking is absent from ordinary runtime profiles;
- policy taxonomies live in independently versioned official kits.

Compile speed still depends on algorithms, query granularity, invalidation,
solver bounds, monomorphization, backend work, and diagnostics. Every trust
feature therefore carries edit-time, memory, cache, and false-positive budgets.

## Probe matrix

| ID | Adversary | Pass condition |
|---|---|---|
| FLOW-001 | HTTP bytes go directly to a shell program | compile rejection points from source to raw sink and proposes structured process API |
| FLOW-002 | regex capture is treated as universally trusted | rejected unless it constructs a specific nominal type with a declared rule |
| FLOW-003 | HTML-text encoding enters JavaScript context | context-type mismatch |
| FLOW-004 | parameterized SQL receives hostile string data | accepted as data; generated protocol never concatenates it as syntax |
| FLOW-005 | untrusted string selects SQL identifier | rejected or requires allowlisted `SqlIdentifier` construction |
| FLOW-006 | valid JSON becomes an authorized domain command | schema decoding alone is insufficient; policy/nominal transition required |
| FLOW-007 | secret is fully validated then logged | confidentiality sink rejects it despite validity |
| FLOW-008 | model output proposes a typed destructive call | schema succeeds; authorization still rejects absent capability/policy evidence |
| FLOW-009 | untrusted branch selects privileged executable | targeted control-influence analysis reports the decision path |
| FLOW-010 | sanitizer changes only one field of an aggregate | unaffected fields retain their flow facts |
| FLOW-011 | actor/queue/cache round trip drops provenance | schema and policy summary preserve or explicitly re-establish facts |
| FLOW-012 | foreign function has no flow summary | conservative propagation plus release obligation; never implicit clean result |
| FLOW-013 | dependency claims sanitizer behavior | claim requires trusted-kit provenance and conformance evidence |
| FLOW-014 | declassification omits purpose or authority | compile rejection and no audit event emitted |
| FLOW-015 | static model misses a native/foreign propagation | instrumented adversarial build reports the source-to-sink trace |
| FLOW-016 | flow diagnostic contains secret input bytes | diagnostic contains semantic IDs and redacted shapes only |
| FLOW-017 | global flow analysis follows every source to every sink | analysis is policy-scoped, incremental, budgeted, and reports deferred work honestly |
| FLOW-018 | policy changes while cached validation survives | semantic invalidation selects cached facts, callers, tests, and artifacts |

## Remaining experimental seams

1. Whether `Input<Origin, T>` is the best surface or merely a projection of IR
   influence facts.
2. Which facts are field-sensitive in the fast lane.
3. Which high-risk sinks require control-influence checking during `check`
   rather than `verify`.
4. Whether confidentiality is expressed through annotations, wrappers, or a
   narrow qualifier algebra.
5. How policy identity and version participate in public API compatibility.
6. How much provenance is retained at runtime without leaking data or adding
   unbounded cardinality.
7. Which dependency summaries are trusted, generated, audited, or treated as
   unproven obligations.
8. Whether a richer Jif/Flow-Caml-style lattice earns admission for selected
   high-assurance components.

These are executable questions. None requires a Boolean runtime taint mode in
ordinary Lang programs.

## Recommendation status

Adopt the **hybrid direction as a strong candidate** and preserve its typed-IR
hooks in Experiment 0. Do not add dedicated taint syntax or pervasive runtime
metadata yet. Promote stronger flow semantics only when the probe corpus shows
that nominal safe sinks plus targeted analysis leave important preventable
failures—or when they reduce AI repair and audit cost enough to pay for their
compiler complexity.

Sources in this note were last verified 2026-09-03. Recommendations are project
synthesis, not claims made by the cited systems.
