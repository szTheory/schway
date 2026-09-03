---
id: context-telemetry-security
title: Context, telemetry, privacy, and operations
summary: A candidate design for typed propagation and useful-by-default observability without ambient authority, secret leakage, or unbounded cost.
type: design
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-03
tags: [context, telemetry, privacy, security, devops, sre]
related: [vision, effects-and-capabilities, trust-validation-information-flow, feature-coherence, networking-http-tls-security, example-tour, research-ledger]
---

# Context, telemetry, privacy, and operations

## Recommendation

Make telemetry a runtime semantic substrate and context a narrow typed channel.
Do not make logging calls, arbitrary implicit parameters, or untyped baggage the
foundation.

The runtime already observes meaningful transitions: capability calls, task and
actor lifecycle, message sends, remote boundaries, retries, failures, resource
pressure, and release identity. It can emit low-cost structural events for
these transitions. Programs enrich them through classified typed fields. Export,
retention, sampling, and payload collection remain policy decisions with explicit
budgets.

This delivers “observability for free” only in a disciplined sense: correlation,
operation identity, timing, outcome shape, and causal structure are automatic.
Business meaning and sensitive payloads are never guessed.

## Four things commonly called context

These must not collapse into one global map.

| Kind | Examples | Rule |
|---|---|---|
| Execution context | cancellation, deadline, trace identity, resource budget | May propagate structurally with tasks; inspectable and bounded. |
| Security context | authenticated principal, grants, trust boundary | Validated at ingress; never trusted merely because it propagated. |
| Business context | tenant, account, locale, feature decision | Explicit domain input when behavior or authorization depends on it. |
| Diagnostic context | experiment cohort, request correlation, debug tags | Classified, budgeted, and never an authority source. |

The key split is semantic importance. A trace ID can remain contextual. A
`TenantId` that changes which records a use case may read belongs in the
application/domain input, even if the edge adapter originally derived it from
request context.

## Candidate context surface

```text
context RequestExecution {
  trace: TraceContext @classification(public_identifier),
  deadline: Deadline,
  cancellation: Cancellation,
  budget: WorkBudget,
}

context AuthenticatedRequest {
  principal: Verified<Principal> @classification(restricted),
  tenant: Verified<TenantId> @classification(restricted),
}
```

Context parameters are type-directed but not invisible:

```text
fn handle(request: HttpRequest)
  -> Result<HttpResponse, RequestError>
  using { execution: RequestExecution, auth: AuthenticatedRequest }
{
  let command = PlaceOrder(
    tenant: auth.tenant.value,
    actor: auth.principal.value,
    lines: decode_lines(body: request.body)?,
  )

  place_order(command: command)
}
```

Rules:

- `using` arguments can always be written explicitly and queried by tools.
- Resolution is lexical and type-directed; there is no global mutable context.
- Imports of providers/context are syntactically distinct from value imports.
- Ambiguous or recursive resolution is a compile error with the candidate graph.
- Context does not perform implicit type conversion.
- Only declared context types participate; arbitrary values are never inferred.
- Child tasks inherit permitted execution context. Detached tasks require an
  explicit propagation statement.
- Remote propagation passes through a typed encoder and trust-boundary policy.
- Security/business values must be revalidated or converted to explicit inputs.

Scala 3's contextual-abstraction documentation is valuable negative and positive
evidence: it describes term inference as useful for context, DI, capabilities,
and type classes, while acknowledging that overly powerful implicits were
misused, challenged tooling, and produced poor failure messages. Its redesign
separates givens, using clauses, imports, and conversions. This proposal should
be narrower still. ([Scala 3 contextual abstractions](https://docs.scala-lang.org/scala3/reference/contextual/index.html))

## Propagation is a security boundary

Automatic propagation is convenient precisely because it is easy to forget.
OpenTelemetry warns that baggage may reach unintended third-party resources and
has no built-in integrity checks. W3C Trace Context forbids PII in trace fields
and calls for assessment of header abuse and trust boundaries.
([OTel baggage](https://opentelemetry.io/docs/concepts/signals/baggage/),
[W3C Trace Context](https://www.w3.org/TR/trace-context/))

Therefore every context field has metadata that survives through the compiler
and runtime:

```text
@propagation(local_task)
@classification(restricted)
@integrity(verified_at_ingress)
@retention(none)
tenant: Verified<TenantId>
```

Candidate propagation classes:

- `local_call`: ordinary lexical context only;
- `local_task`: inherited by structured child tasks;
- `actor_message`: requires typed message encoding;
- `trusted_service`: crosses an allowlisted boundary with integrity policy;
- `public_wire`: safe for external protocols;
- `never`: secrets, raw credentials, handles, and non-exportable values.

The defaults should be `local_call` and `retention(none)`, not “propagate and
record everything.”

Classification governs observation and retention; it is not input validation
or authority. A `Secret<VerifiedPrincipal>` may be valid and high integrity yet
still forbidden from telemetry, while attacker-controlled public text may be
safe to render through a context-specific text sink. See [Trust, validation,
and information flow](trust-validation-and-information-flow.md) for the
separate axes and source-to-sink analysis.

## Typed semantic events

Runtime events should be versioned data, not interpolated strings:

```text
event CapabilityCompleted {
  operation: OperationId,
  caller: DefinitionId,
  duration: Duration,
  outcome: OutcomeClass,
  attempts: Int,
  trace: TraceId,
  fields: EventFields,
}
```

Enrichment is explicit and classification-aware:

```text
observe current_operation {
  field order_kind = command.kind
    @classification(operational)
    @cardinality(low)

  field tenant = command.tenant
    @classification(restricted)
    @cardinality(high)
    @transform(pseudonymize)
}
```

The checker should reject or warn on:

- a `Secret`, access token, password, connection string, or payment payload;
- unbounded text or collection fields without size policy;
- high-cardinality fields sent to metric labels;
- propagation beyond a field's trust/classification policy;
- export to a sink whose clearance is too low;
- event schemas changed incompatibly;
- sampling that can remove mandatory audit/security events.

OWASP advises excluding or transforming session IDs, access tokens, sensitive
PII, passwords, connection strings, keys, and payment data; it also recommends
validation, sanitization, consistent schemas, and testing failures of the logging
system itself. Those should become default type/policy checks rather than a
checklist every application reimplements.
([OWASP Logging Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html))

## Wide events, traces, and projections

The proposal at [Logging Sucks](https://loggingsucks.com/) usefully emphasizes
high-dimensional structured “wide events,” correlation, and tail sampling. The
strongest idea is to accumulate request outcome context instead of scattering
sentence fragments through the code.

Adopt that as a **projection**, not a universal primitive:

```text
causal runtime events
  -> request/service-hop summary (wide event)
  -> trace/span export
  -> metrics with bounded labels
  -> audit stream with separate retention
  -> live agent/debug view
```

One event per request does not fit long-lived actors, streaming requests,
workflows with multiple durable outcomes, or concurrent fan-out. The internal
model should be a typed causal event graph; a wide event is one useful summary
of it. The examples in the article also demonstrate why unrestricted enrichment
is dangerous: exact queries, IPs, user identifiers, JWT subjects, and customer
or payment attributes can create privacy, security, and storage costs. Retention
by “VIP user” must be an explicit governed policy, never a language default.

## Cost model and degradation

Every deployment declares an observability budget:

```text
telemetry production {
  cpu_budget: 0.5.percent,
  memory_budget: 32.mebibytes,
  egress_budget: 2.mebibytes_per_second,
  tail_buffer: 10.seconds,
  payloads: deny,
  on_overflow: preserve_errors_then_sample,
}
```

The runtime reports actual overhead and dropped/sampled counts. Policies should
distinguish:

- structural events, which are compact and generally on;
- enrichment fields, checked against classification/cardinality budgets;
- payload capture, off by default;
- focused debug leases, time-bounded and auditable;
- mandatory audit events, isolated from ordinary diagnostic sampling.

Telemetry failure must not silently take down the application or silently lose
required audit evidence. Each application declares whether a sink failure is
`best_effort`, `degrade`, or `fail_closed`; the last is reserved for explicit
compliance/security requirements.

## Live agent introspection

CLI and MCP-like tools query the same typed runtime model:

```text
lang inspect effects --definition shop.orders/place_order --format json
lang inspect providers --service CheckoutApi --explain
lang inspect runtime --failures --since 10m --format json
lang inspect trace --id 7f... --causal --redaction agent_safe
lang debug focus --component CheckoutApi --duration 60s --budget 5MiB
```

The protocol returns stable IDs, schemas, causal links, source locations,
classification/redaction decisions, costs, and machine-applicable actions. It
does not require an agent to scrape colorful terminal prose, and it does not
grant access merely because the process can be inspected. Debug capabilities
are scoped, authenticated, rate-limited, and audited.

## Configuration, secrets, releases, and containers

The Twelve-Factor guidance separates deploy-varying config from code, treats
logs as event streams, and separates build, release, and run. The core ideas
remain useful, but environment variables are an input transport—not a complete
secret or configuration type system.
([config](https://12factor.net/config), [logs](https://12factor.net/logs),
[build/release/run](https://12factor.net/build-release-run))

Candidate operational contract:

```text
config CheckoutConfig {
  database_url: Secret<DatabaseUrl> from deploy,
  public_origin: Url from deploy,
  retry_limit: Int range 0..8 = 3,
}
```

- `from deploy` forbids a source literal and records configuration provenance.
- Secret values cannot be printed, serialized, compared for debugging, embedded
  in compiler artifacts, or exported as telemetry without an explicit unsafe
  declassification.
- `.env` is a development adapter, gitignored and never the semantic model.
- a generated config manifest tells Docker/Kubernetes/CI what must be supplied;
  missing or malformed inputs fail before serving traffic;
- a build is hermetic and content-addressed; a release combines immutable build
  identity with config identity; runtime mutation produces a new auditable
  configuration revision;
- the process writes structured events to stdout/OTLP and does not manage log
  rotation;
- health, readiness, graceful shutdown, signal handling, and resource limits are
  provided by the service profile with overrideable policy;
- build provenance should be emitted alongside artifacts; SLSA defines
  provenance as tracking build outputs back to their source and build process.
  ([SLSA provenance](https://slsa.dev/spec/v1.2/provenance))

Container support should be an artifact target, not syntax. `lang package
--target oci` can generate a minimal non-root image, SBOM, provenance,
configuration manifest, health contract, and deterministic entrypoint from the
same build graph.

## Open design questions

- Which context types are privileged by the language versus defined through a
  sealed core protocol?
- Can classification and propagation policies remain understandable without
  becoming a second type system?
- What is the minimum structural event set whose overhead is acceptable for
  service and native profiles?
- How are audit integrity, deletion/retention law, and right-to-erasure policies
  represented without pretending one global policy exists?
- Does focused debug instrumentation use hot code replacement, dormant probes,
  eBPF/platform probes, or handler replacement per backend?
