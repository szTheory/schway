---
id: language-toolchain-primer
title: Candidate language and toolchain primer
summary: A plain-language guide to the candidate syntax, effects, dependency injection, failure model, specifications, and check-to-release commands.
type: examples
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-02
tags: [syntax, effects, toolchain, onboarding]
related: [example-tour, effects-and-capabilities, type-system-failure-semantics, resources-locks-caching, modules-architecture-live-development, compiler-feedback-latency, design-baseline]
---

# Candidate language and toolchain primer

## Read this before the examples

The syntax is design fiction, not a final grammar. `lang` is a placeholder
command until the project has a name. The important proposal is the semantic
model underneath the punctuation.

## A complete small example

```text
fn join_member(command: JoinMember)
  -> Result<Member, JoinError>
  with { Members, Clock }
{
  let joined_at = Clock.now()
  let member = Member(
    id: command.member_id,
    email: command.email,
    joined_at:,
  )

  Members.insert(member:)?
  Ok(member)
}
```

Read the signature as: “`join_member` accepts a `JoinMember`, returns either a
`Member` or a declared `JoinError`, and is authorized to use the `Members` and
`Clock` capabilities.”

### Is `Clock.now()` a side effect?

Yes. Wall-clock time is nondeterministic. The effect is visible in
`with { Clock }`; the provider graph supplies the actual clock. The ordinary
call punctuation does not imply purity. A more visually explicit
`Clock.now!()` remains a syntax experiment, but it would not change semantics.

```text
provide production {
  Clock = SystemClock
  Members = PostgresMembers(pool: AppDatabase)
}

provide test overrides production {
  Clock = FixedClock(at: 2030-01-02T03:04:05Z)
  Members = MemoryMembers()
}
```

This is dependency injection without a reflective service locator. The
compiler resolves the provider graph, lifetimes, missing providers, ambiguity,
and cycles. Application functions depend on small capabilities rather than
concrete adapters.

Pure runtime-selected strategies may still be passed as ordinary values. They
do not need to become container entries.

## Named arguments and the elided value

```text
Inventory.release(reservation: reservation)
Inventory.release(reservation:)
```

The second form is exact-name punning: it means the first form when exactly one
unqualified local named `reservation` is in scope. It keeps the semantic role
visible without repeating a token. The current canonical-formatting candidate
accepts either spelling but rewrites the first to the second; ambiguity is an
error rather than a formatter guess. Shadowing and model familiarity still need
testing. Multi-role calls remain named; concise unary algebraic and pipeline
calls are an open exception.

## Values, variants, and pattern matching

```text
record Receipt { payment: PaymentId, at: Instant }

data CaptureError =
  | Declined { reason: DeclineReason }
  | Unavailable { retry_after: Duration }

match Payments.capture(invoice:) {
  Ok(payment:) => Ok(Receipt(payment:, at: Clock.now()))
  Err(Declined(reason:)) => Err(PaymentDeclined(reason:))
  Err(Unavailable(retry_after:)) => Err(TryLater(retry_after:))
}
```

Records are product types; `data` variants are sum types. Pattern matching must
be exhaustive, so adding a new error variant creates a precise obligation at
every affected match.

Values are immutable by default. `order with { status: Cancelled(...) }`
constructs an updated value rather than mutating shared state.

## Expected failure, defects, and `?`

```text
let reservation = Inventory.reserve(items: command.lines)?
```

`?` explicitly propagates the declared `Err` variant from a `Result`; it does
not catch exceptions. A caller must handle, translate, or propagate expected
failure. Cancellation is separate control flow. Broken invariants panic the
smallest supervised isolation boundary and produce causal evidence; ordinary
application code cannot catch and ignore them.

## Flat specifications

```text
spec "declined payment releases inventory" with test_providers {
  let command = fixtures.place_order()
  Inventory.expect_reserve(items: command.lines)
  Payments.stub_charge(result: Err(Declined(reason: CardExpired)))

  let result = place_order(command:)

  expect result == Err(PaymentDeclined(reason: CardExpired))
  expect Inventory.released_count() == 1
}
```

The style is deliberately flat: arrange, act, assert; ordinary bindings; no
nested setup precedence or shared-example metaprogramming. Types and contracts
can generate boundary cases and identify missing evidence, but cannot invent
the intended business oracle.

## What `check`, `verify`, and `release` mean

The words describe different feedback scopes, not three synonyms for compile.

| Command | Promise | Expected speed |
|---|---|---|
| `lang fmt` | Rewrite source to the one canonical layout. | immediate |
| `lang check` | Parse, resolve, type/effect/resource-check, enforce architecture, and report evidence obligations for the affected semantic slice. It executes no application code. | subsecond when warm |
| `lang test path-or-query` | Execute selected examples, properties, simulations, contracts, or integrations with explicit providers. | workload-dependent |
| `lang run` | Run or compatibly reload a development application with an inspectable provider/runtime graph. | interactive |
| `lang inspect` | Query types, effects, callers, providers, obligations, costs, live actors, queues, traces, and failure causes as text or structured data. | interactive |
| `lang verify changed` | Assemble the policy-required evidence for the semantic impact cone: tests, properties, contracts, architecture, security, compatibility, and budgets. | bounded but slower |
| `lang verify all` | Evaluate the complete configured evidence policy from hermetic inputs. | CI/release lane |
| `lang build` | Produce an artifact; it does not falsely imply that all release evidence passed. | profile-dependent |
| `lang release` | Produce a versioned immutable artifact plus provenance, compatibility, migration, and evidence manifests. It fails unless release policy is satisfied. | deliberate gate |

The earlier question “should missing specs block `check` or
`verify`/`release`?” means this: should unfinished evidence prevent a programmer
from quickly checking otherwise sound code, or should `check` identify the
exact obligation while the stronger gates refuse shipment?

The current recommendation is:

- `check` blocks unsound code and objective policy violations, but reports
  missing broader evidence as an obligation;
- `verify` blocks when its selected evidence policy is incomplete or failing;
- `release` requires the project’s complete release policy;
- safety-critical profiles may deliberately promote named obligations into
  `check` without making that the universal default.

This preserves a fast edit loop without letting “it compiled” masquerade as
“it is production-ready.”

## How an agent sees the same program

An agent should not scrape colorized compiler prose. The toolchain exposes a
versioned semantic protocol:

```text
lang inspect function join_member --format json
lang inspect effects --changed --format json
lang inspect impact --since HEAD~1 --format json
lang verify changed --format json
```

Responses include stable symbol IDs, definitions, callers, inferred and
declared effects, provider bindings, failure variants, architecture edges,
stale evidence, runtime correlation IDs, and suggested repairs. Text remains
the human review surface; this semantic graph is the machine reasoning surface.

## Where to go next

- The [example tour](example-tour.md) shows broader realistic code.
- [Effects and capabilities](effects-and-capabilities.md) explains why this is
  authority tracking as well as DI.
- [Compiler feedback](compiler-and-feedback-latency.md) describes how the loop
  stays fast.
- The [capability gauntlet](capability-gauntlet.md) records which software
  classes remain design claims rather than demonstrated capability.
