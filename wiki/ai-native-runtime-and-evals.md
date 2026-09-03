---
id: ai-native-runtime-evals
title: AI-native runtime, agents, and evaluation
summary: A future-resilient design for model calls, tools, agent workflows, traces, budgets, replay, and stochastic evaluation without vendor-specific core syntax.
type: design
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-03
tags: [ai, agents, evaluation, runtime]
related: [effects-and-capabilities, trust-validation-information-flow, context-telemetry-security, domain-data-distribution, compiler-feedback-latency, capability-gauntlet, research-ledger]
---

# AI-native runtime, agents, and evaluation

## Conclusion

The language should be AI-native in its **semantics and feedback loop**, not by
enshrining today’s model vendors or agent frameworks in grammar.

- Model inference is a typed, budgeted, nondeterministic effect.
- A tool is a capability with input/output/error schemas and explicit authority.
- An agent is a supervised workflow that combines models, tools, state, guards,
  and handoffs.
- An agent graph is an inspectable projection, not necessarily the source of
  truth.
- An eval is versioned stochastic evidence, not a boolean unit test.
- A trace is a causal graph with classified payloads, not an excuse to record
  every prompt and secret.

This design survives changes in model architecture, hosting, APIs, and
orchestration fashion because it is built from the same effects, contracts,
state machines, structured concurrency, durable workflows, and telemetry used
elsewhere.

## Candidate model capability

```text
capability Model<In, Out> {
  infer(
    input: In,
    policy: InferencePolicy,
  ) -> Result<ModelReply<Out>, ModelError>
}

record InferencePolicy {
  model: ModelSelector,
  deadline: Deadline,
  max_tokens: TokenCount,
  max_cost: Money<Usd>,
  data_class: DataClass,
  residency: ResidencyPolicy,
  cache: CachePolicy,
  retries: RetryPolicy,
}
```

`Out` is decoded and validated against a derived schema. Schema conformance
does not establish factual or business correctness. The reply carries usage,
latency, provider/model identity, request provenance, safety outcomes, and an
opaque reference to classified raw evidence.

The compiler can ensure that a budget and provider exist. It cannot promise a
model will obey an unstated intention.

## Typed tools and authority

```text
tool ReadCustomer : Tool<CustomerQuery, CustomerView, LookupError> {
  authority: read CustomerData
  idempotency: safe
  trust: internal
}

tool RefundPayment : Tool<RefundCommand, RefundReceipt, RefundError> {
  authority: write Payments
  idempotency: keyed(by: RefundCommand.idempotency_key)
  approval: required(when: amount > 500.usd)
  trust: internal
}
```

A remote MCP or function-calling schema may describe a tool’s shape, but remote
annotations are untrusted input. The application’s trusted adapter grants the
actual authority, resource limits, timeout, redaction, and audit policy. Text
from a user, model, retrieval result, or tool response can never mint a
capability.

Model output and retrieved content retain low-integrity influence even after
schema validation. Decoding establishes shape; an application policy and
capability boundary separately authorize an action. The candidate type and
analysis model is specified in [Trust, validation, and information
flow](trust-validation-and-information-flow.md).

Useful tool properties include:

- read, write, destructive, privileged, or open-world authority;
- idempotency key and retry safety;
- input/output/error and data-classification schemas;
- deadline, cost, rate, concurrency, and payload budgets;
- transaction or compensation boundary;
- approval policy and reason;
- trusted adapter identity and protocol version.

## Agent workflows without magical graphs

```text
workflow ResolveRefund(case: SupportCase)
  -> Result<Resolution, ResolutionError>
  with { TriageModel, ReadCustomer, RefundPayment, HumanApproval }
{
  let triage = TriageModel.infer(
    input: TriageInput.from(case:),
    policy: refund_triage_policy,
  )?

  let customer = ReadCustomer.call(input: CustomerQuery.from(case:))?

  match decide_refund(triage: triage.value, customer:) {
    Reject(reason:) => Ok(Resolution.rejected(reason:))
    Refund(command:) if command.amount <= 500.usd =>
      RefundPayment.call(input: command)?.into_resolution()
    Refund(command:) => {
      HumanApproval.require(subject: command, reason: "large refund")?
      RefundPayment.call(input: command)?.into_resolution()
    }
  }
}
```

An editor can render the control/data/effect graph, simulate paths, display
authority at each node, and suggest parallelism. The checked-in semantics remain
ordinary typed control flow. A specialized visual or declarative projection may
be admitted later if it proves easier to review and evolve.

Short work uses structured tasks. Long-running, crash-resilient work uses a
durable workflow kit that records every nondeterministic result before making
the next decision. Actors own long-lived concurrent state. These mechanisms
should compose but remain distinct.

## Handoffs

A handoff is not “send the entire chat to another agent.” It transfers an
explicit envelope:

```text
record Handoff<Goal, Context, Authority> {
  goal: Goal,
  context: Context,
  authority: Authority,
  budget: RunBudget,
  evidence: EvidenceRefs,
  completion: CompletionContract,
}
```

The receiver cannot acquire greater authority or budget than the sender
delegated. Ownership, cancellation, deadline, expected output, and escalation
path remain visible. Input filtering and authorization occur before any
effectful action.

## Budget algebra

AI workloads make several resource dimensions user-visible at once:

```text
record RunBudget {
  deadline: Deadline,
  tokens: TokenCount,
  cost: Money<Usd>,
  tool_calls: Count,
  parallelism: Count,
  memory: Bytes,
}
```

Child tasks receive a partition or explicitly shared view of the parent budget;
they cannot mint more. Retries consume the same deadline and attempt budget.
Queues and fan-out are bounded. The runtime can explain which dimension stopped
a run and where the resources went.

Static analysis can catch missing or obviously incompatible budgets. Runtime
enforcement handles values known only during execution.

## Replay honesty

There are two different operations:

- **Recorded replay** reuses stored model/tool/nondeterministic outputs and
  deterministically reruns surrounding code.
- **Live rerun** calls models and tools again and may produce a different
  result.

The toolchain must never label the second operation deterministic replay. A
durable workflow event history records model selection, prompt/artifact hashes,
tool schemas, sampling parameters, outputs, and decisions needed to resume. It
also requires explicit versioning when workflow code changes.

## Evals are a native evidence family

```text
eval refund_triage for ResolveRefund {
  dataset: RefundCases(version: "2026-09-01")
  samples_per_case: 5
  graders: [
    ExactField("decision"),
    Rubric("policy adherence", judge: calibrated_policy_judge),
    ToolAuthorityInvariant(),
  ]
  require {
    policy_adherence.lower_confidence_bound >= 0.98
    unauthorized_tool_calls == 0
    p95_cost <= 0.08.usd
  }
}
```

This is candidate toolkit notation, not a core keyword decision. The semantic
artifact records:

- objective and versioned dataset;
- typical, edge, adversarial, and regression partitions;
- reference labels and who produced them;
- model/provider/sampling/prompt/tool/retrieval versions;
- grader definitions, calibration, and known bias;
- sample count, seeds where applicable, distribution, confidence, and
  thresholds;
- cost, latency, safety, authority, and trajectory metrics;
- failing traces and minimized reproducible cases.

Pairwise, classification, and criteria-based grading are preferred over vague
open-ended scores where possible. Model judges must be calibrated against human
expert labels. Continuous evals run on every relevant change or sampled
production traces, not only before launch.

## Change impact for AI artifacts

Prompts, model policies, tool schemas, retrieval corpora, datasets, graders,
and safety policy are versioned semantic artifacts. The dependency graph can
answer:

```text
lang inspect impact --changed prompt:refund_triage
lang verify changed --include evals --format json
```

Changing a prompt selects the evals and contracts that depend on it. Changing a
tool’s authority or schema selects callers, simulations, and compatibility
checks. A model alias resolving to a new snapshot is a deployable configuration
change with provenance—not an invisible environment mutation.

## Telemetry and privacy

Built-in tracing should create spans/events for runs, model generations, tool
calls, guards, handoffs, approvals, retries, budgets, and outcomes. The event
schema should align with open telemetry conventions where useful.

Content is classified separately from structural metadata:

- structural timing, model/tool identity, outcome, and usage may be safe to
  retain broadly;
- prompts, outputs, retrieval passages, and tool payloads may contain secrets,
  PII, copyrighted material, or attacker-controlled instructions;
- payload capture is redacted or off by default according to declared data
  class;
- narrowly scoped debug leases can temporarily increase capture with owner,
  expiry, audit, and cost limits.

## Local inference and accelerator support

Do not embed a transformer engine in the language runtime initially. Define a
stable `Model` capability and efficient tensor/buffer/accelerator interop.
Promote local scheduling into the runtime only after workloads demonstrate a
shared need for batching, cache reuse, accelerator placement, admission
control, or isolation that an ordinary provider cannot meet.

This keeps the door open to on-device inference, GPU execution, model
compilation, and new architectures without coupling the core language to one
generation of inference machinery.

## Anti-patterns

- A core `agent` DSL coupled to one vendor’s abstractions.
- Treating valid JSON as a correct decision.
- Granting tool authority from model-produced names or remote annotations.
- Implicit retries of write tools without idempotency evidence.
- Unbounded autonomous loops, recursive delegation, or parallel tool fan-out.
- Full-history handoffs with ambient credentials.
- Logging every prompt/output by default.
- Calling one stochastic run a regression test.
- Using the same uncalibrated model as author and sole judge.
- Claiming deterministic replay while making live external calls.

## Primary evidence

- [OpenAI evaluation best practices](https://developers.openai.com/api/docs/guides/evaluation-best-practices)
  describes objective/dataset/metric design, edge and adversarial cases,
  continuous evaluation, and judge limitations.
- [OpenAI agent evals](https://developers.openai.com/api/docs/guides/agent-evals)
  distinguishes trace grading for workflows from repeatable dataset evals.
- [OpenAI function calling](https://developers.openai.com/api/docs/guides/function-calling)
  documents strict schema conformance and its constraints.
- [OpenAI Agents tracing](https://openai.github.io/openai-agents-python/tracing/)
  records agents, generations, tools, guardrails, and handoffs and documents
  sensitive-data controls.
- [MCP tools](https://modelcontextprotocol.io/specification/2025-06-18/server/tools)
  defines input/output schemas and warns that tool annotations are untrusted.
- [MCP authorization](https://modelcontextprotocol.io/specification/2025-06-18/basic/authorization)
  shows that transport authorization is a separate concern.
- [Temporal workflow definition](https://docs.temporal.io/workflow-definition)
  explains deterministic workflow constraints and moving external actions into
  activities.
- [OpenTelemetry generative AI conventions](https://opentelemetry.io/docs/specs/semconv/gen-ai/)
  provide an evolving interoperability target for model and agent telemetry.

All sources were last verified 2026-09-02. The design recommendations are
synthesis, not claims made by those projects.
