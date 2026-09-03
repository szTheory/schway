---
id: boundary-data-validation-persistence
title: Boundary data, validation, transactions, and pools
summary: A first-party data-boundary contract for exact JSON decoding, explicit changes, retry-safe transactions, and bounded database connection ownership.
type: design
status: candidate
confidence: medium
created: 2026-09-03
updated: 2026-09-03
tags: [json, validation, persistence, transactions, pooling]
related: [application-architecture-data, trust-validation-information-flow, type-system-failure-semantics, effects-and-capabilities, resources-locks-caching, domain-data-distribution, data-asset-pipelines, residual-uncertainty-register, capability-gauntlet, research-ledger]
---

# Boundary data, validation, transactions, and pools

## Reader and outcome

This is for contributors building the JSON/schema and relational-data kits.
After reading it, they should be able to implement the first boundary slice
without confusing parsing with validation, a proposed change with a committed
fact, or a connection pool with an invisible global.

## Verdict

JSON, Zod/Pydantic-like runtime validation, Ecto-like changes, transaction
plans, SQL, and connection pooling should be excellent first-party facilities.
They should not become unrelated language syntax. The current semantic kernel
already has the mechanisms they need:

- ADTs and nominal types distinguish presence, validity, identity, and state;
- typed derivations produce codecs and schemas from an inspectable declaration;
- effects and capabilities make parsing limits, repositories, clocks, and
  external authority explicit;
- affine resources model streams, cursors, transactions, and connection leases;
- `Result`, cancellation, and structured tasks preserve failures and cleanup;
- stable semantic identities version schemas, validators, queries, and evidence.

The early compiler requirement is therefore to preserve schema, field,
validator, change, transaction, and resource identities in typed IR. The policy
and implementation belong in independently versioned official kits.

## The complete boundary pipeline

```text
untrusted bytes
  -> bounded syntax value
  -> versioned wire shape
  -> permitted proposed change
  -> validated nominal command
  -> authorized transaction plan
  -> committed receipt + post-commit work
```

Each arrow establishes one named fact. No arrow silently establishes the facts
to its right.

| Stage | Establishes | Does not establish |
|---|---|---|
| JSON parse | valid bounded JSON syntax | expected fields, precision, business validity |
| typed decode | declared wire shape and conversion policy | domain invariant or authorization |
| change construction | which fields were proposed and permitted | current database truth |
| pure validation | local invariant under a named policy version | uniqueness, freshness, ownership |
| authorization | principal may request the operation now | database commit or external completion |
| transaction commit | database accepted one atomic unit under stated isolation | message delivery or remote side effect |
| post-commit workflow | attempted downstream consequences | global atomicity |

## JSON is a protocol, not a universal object model

The raw JSON kit should preserve enough information to avoid accidental loss:

- distinguish object, array, string, Boolean, null, and an exact bounded number
  token rather than eagerly converting every number to binary floating point;
- detect duplicate member names and never silently inherit a parser's
  first-wins or last-wins behavior;
- preserve source path and byte span when available;
- impose byte, nesting, collection, string, number-digit, and work limits;
- separate parsing from target-specific numeric and text conversion;
- keep canonicalization, pretty printing, and ordinary serialization as
  different named operations.

RFC 8259 permits implementation limits and records interoperability problems
around duplicate object names and numeric precision. RFC 8785 adds a particular
canonical JSON representation for hashing and signing, including stricter
number and member-name requirements. That makes canonicalization useful but not
an invisible default serializer or a definition of semantic equality.
([JSON](https://www.rfc-editor.org/rfc/rfc8259),
[JSON Canonicalization Scheme](https://www.rfc-editor.org/rfc/rfc8785.html))

```text
let syntax = Json.parse(
  input: body,
  limits: PublicApi.json_limits,
  duplicate_fields: reject,
)?

let request = Json.decode<CreateOrderV3>(
  value: syntax,
  unknown_fields: reject,
  numbers: exact,
)?
```

The convenient combined operation may fuse parsing and decoding so it does not
build an intermediate tree. Tooling still reports the two semantic stages and
their independent failures.

### Strictness is boundary-specific

There is no universal correct unknown-field policy:

| Boundary | Leading default | Reason |
|---|---|---|
| public command ingress | reject unknown fields | catches misspellings and ambiguous intent |
| tolerant response reader | ignore but report unknown fields | supports additive producer evolution |
| relay, proxy, or read-modify-write | preserve unknown fields | prevents an older intermediary destroying newer data |
| signed or hash-addressed data | exact declared canonical policy | byte identity and semantic identity must not drift |

Protocol Buffers demonstrates why presence and unknown-field preservation are
separate compatibility concerns: explicit presence distinguishes unset from a
default value, removed field numbers must not be reused, and unknown fields can
survive binary round trips but be lost through some transformations.
([Protocol Buffers evolution](https://protobuf.dev/programming-guides/proto3/))

### Missing, null, default, and unchanged are different

Patch and update inputs need an explicit presence algebra:

```text
data FieldPatch<a> = Unchanged | Set(value: a) | Clear
```

`Option<T>` describes a value that may be absent in a completed domain model;
it does not by itself describe whether an update omitted a field, explicitly
set it to null, requested deletion, or supplied a default-valued value. RFC
7396 gives null the special meaning “remove” and therefore cannot represent
every document model; JSON Patch is a distinct operation format. Lang should
derive each intentionally rather than treating a partial record as obvious.
([JSON Merge Patch](https://www.rfc-editor.org/rfc/rfc7396.html),
[JSON Patch](https://www.rfc-editor.org/rfc/rfc6902.html))

### Schema dialects and formats are versioned dependencies

Generated JSON Schema carries an explicit dialect, vocabulary set, schema ID,
and compatibility direction. Its `format` behavior cannot be assumed: JSON
Schema 2020-12 separates format annotation from the optional format-assertion
vocabulary. A Lang validator therefore records whether a format was asserted,
annotated, or checked later by application policy.
([JSON Schema 2020-12](https://json-schema.org/draft/2020-12),
[validation vocabulary](https://json-schema.org/draft/2020-12/json-schema-validation))

Schemas can themselves be untrusted or computationally hostile. Loading remote
references, regexes, recursive schemas, or custom vocabularies requires
capabilities, resolution policy, cycle handling, and work budgets. Generated
schemas are signed/versioned artifacts; arbitrary schemas never extend the
compiler's trusted type system.

## Validation should produce narrow evidence

Use three deliberately different abstractions:

```text
Decode<Wire>                  // representation and shape
Change<Base, Intent>          // permitted proposed fields and local errors
Verified<RuleVersion, Value>  // one named predicate was established
```

A Zod-like `parse` operation is a useful ergonomic baseline, but one successful
parse must not collapse representation, business validity, authorization, and
freshness into a generic “safe” value.

Validation errors are a bounded tree with stable codes, semantic field paths,
source spans where available, expected/observed shapes, redaction class, and
repair hints. Human messages may be localized without becoming the machine
contract. Independent cheap errors can accumulate; recursive, remote, or
expensive checks obey explicit work and disclosure budgets.

Identity follows the same rule:

```text
OrderId.parse(text:)          // correct representation for an OrderId
Orders.authorize(id:, actor:) // actor may address this order
Orders.load(id:)              // it currently exists in this consistency view
```

None implies the next.

### Validation can expire

Pure facts such as “this string has the Email grammar” may be stable. Facts
such as uniqueness, account state, entitlement, inventory, exchange rates, or
policy approval are observations of a versioned world. Their evidence names
the rule, relevant version/snapshot, principal, and expiry or is consumed
inside the transaction that rechecks it.

This prevents a compile-time-looking `Verified<T>` from concealing a
time-of-check/time-of-use race.

## Changes are explicit values, not active records

The first-party `Change<T>` should retain:

- the base value or identity and its observed version;
- the requested operation: insert, update, delete, replace, or domain action;
- field presence and the allowlist that admitted each external field;
- typed converted values plus rejected raw shapes under redaction;
- local validation errors and declared database constraints;
- the policy/schema versions used to derive the change;
- whether the change is still pure, authorized, and eligible for persistence.

Ecto changesets are strong precedent because they distinguish external casting
from internal changes, explicitly permit fields, separate validations from
database constraints, accumulate errors, and can operate without a database.
They also expose the key limit: a local uniqueness check is unsafe while the
database constraint is authoritative under concurrency.
([Ecto Changeset](https://hexdocs.pm/ecto/Ecto.Changeset.html))

```text
fn change_profile(base:, input:) -> Change<Player, UpdateProfile> {
  change(base:, intent: UpdateProfile)
    |> permit(input:, fields: [.display_name, .bio])
    |> parse(.display_name, as: PlayerName)
    |> validate(.bio, using: Text.max_graphemes(count: 280))
    |> expect_version(base.version)
}
```

Mass assignment, hidden callbacks, implicit association deletion, lazy loading,
and persistence methods on domain records are excluded from the ordinary path.

## Transactions are effect-restricted computations

Simple transactions should use ordinary control flow. A named, inspectable
transaction plan earns its cost when steps are composed dynamically, need
preflight validation, or must expose partial results and failure identity.
Ecto.Multi is useful precedent precisely because it is an introspectable data
structure of uniquely named operations; its own documentation recommends
ordinary transaction control flow for most simpler cases.
([Ecto Multi](https://hexdocs.pm/ecto/Ecto.Multi.html))

```text
transaction TransferFunds isolation Serializable retry safe {
  debit  = Accounts.debit(id: from, amount:)
  credit = Accounts.credit(id: to, amount:)
  audit  = Ledger.append(transfer: Transfer(debit:, credit:))

  after_commit Outbox.publish(event: FundsTransferred(audit:))
}
```

Candidate laws:

- the transaction owns one affine database session/lease;
- isolation, read-only status, deadline, and retry policy are explicit;
- retryable bodies admit only the database transaction capability plus pure,
  deterministic computation over captured inputs;
- time, randomness, model calls, network calls, file writes, email, and other
  external effects occur before the retryable body or through a transactional
  outbox/post-commit workflow;
- serialization retry reruns the whole transaction, never the failed statement
  alone;
- optimistic version failure is a typed stale result, not a silent overwrite;
- nested scopes do not pretend savepoints are independent transactions;
- commit uncertainty is distinct from a clean rollback and from known success.

PostgreSQL requires applications using serializable isolation to handle
serialization failures by retrying the whole transaction. It also recommends
keeping transactions no larger or longer than integrity requires and
controlling active connections.
([PostgreSQL transaction isolation](https://www.postgresql.org/docs/current/transaction-iso.html))

The compiler can enforce the effect restriction. It cannot infer that an
arbitrary SQL transaction expresses the correct business invariant.

## A connection pool is a bounded resource scheduler

The pool belongs in the database/runtime kit and is injected as a capability.
It is not a process-global singleton and is never accessed by an active record.

```text
use lease = OrdersDb.acquire(
  deadline: request.deadline,
  priority: interactive,
)?

lease.transaction(isolation: Serializable, using: transfer)
```

The contract includes:

- maximum open and in-flight connections plus a bounded wait queue;
- acquisition deadline, overload result, cancellation, and fairness policy;
- affine lease ownership and deterministic reset/check-in on every exit;
- connection health, authentication/TLS rotation, session-state reset, and
  prepared-query compatibility;
- transaction pinning and explicit rules for child tasks;
- separate queue, connect, server, decode, and total latency evidence;
- leak, long-hold, idle-in-transaction, saturation, and reconnect diagnostics;
- tenant/priority partitioning where one workload could starve another.

DBConnection exposes explicit checkout ownership and separately reports time
waiting for a pool, using a connection, and decoding a result. These are useful
precedents for making the resource and its bottleneck visible.
([DBConnection ownership](https://hexdocs.pm/db_connection/DBConnection.Ownership.html),
[DBConnection timing](https://hexdocs.pm/db_connection/DBConnection.LogEntry.html))

Query cancellation is best effort, not rollback evidence. PostgreSQL documents
that successfully dispatching a cancel request does not guarantee that it had
an effect. The client must still consume/close/reset the protocol state and
classify the final database outcome.
([PostgreSQL query cancellation](https://www.postgresql.org/docs/current/libpq-cancel.html))

## Placement decisions

| Concern | Core semantic hook | First-party kit | Application policy |
|---|---|---|---|
| dynamic boundary value | ADTs, `Result`, ownership | JSON parser/value | accepted media/schema |
| typed decoding | derivation metadata, nominal constructors | codecs and error tree | strictness/coercion |
| schema export | stable identity/version | JSON Schema/OpenAPI/protocol adapters | compatibility promises |
| validation | opaque types/contracts/evidence | validators and `Change<T>` | business rules |
| database access | effects, affine resources, cancellation | driver, typed queries, repository | credentials/topology |
| transactions | effect restriction and cleanup | isolation/retry/plan/outbox | invariant and retry budget |
| pooling | bounded resource primitive | pool implementation and telemetry | size/fairness/deadline |
| query performance | cost/evidence protocol | plan capture and cardinality tools | SLO/index decisions |

No JSON token, SQL keyword, changeset operator, pool, or database isolation
level needs core grammar. Their semantic identities must nevertheless be
queryable by agents and available to change-impact analysis.

## Adversarial kit probes

| ID | Adversary | Pass condition |
|---|---|---|
| JSON-001 | duplicate object name | strict decode rejects with both spans; no silent winner |
| JSON-002 | large integer becomes `F64` | target conversion is explicit or reports precision loss |
| JSON-003 | deep/huge/long-number input | bounded failure before unbounded stack, allocation, or CPU |
| JSON-004 | unknown command field is misspelled | command policy rejects it with a local repair |
| JSON-005 | older relay sees a newer field | preservation policy round-trips it or reports deliberate loss |
| JSON-006 | omitted, null, false, zero, and empty are decoded | presence/default policy distinguishes every intended case |
| JSON-007 | ordinary serialization is used for a signature | type/policy requires a named canonicalization version |
| JSON-008 | schema `format` annotation is treated as asserted | dialect/vocabulary mismatch is diagnosed |
| JSON-009 | hostile remote schema loads refs/regexes | capability and work budget contain resolution/evaluation |
| JSON-010 | validation error contains a secret | stable path/code remains while value is redacted |
| CHANGE-001 | request sets `is_admin` | field allowlist rejects mass assignment |
| CHANGE-002 | uniqueness was checked before a concurrent insert | database constraint decides; local check is not trusted |
| CHANGE-003 | stale base version updates a row | typed conflict; no lost update |
| CHANGE-004 | deletion is implied by omitted child association | explicit domain delete intent required |
| TX-001 | serializable conflict occurs after reads | entire eligible transaction retries within budget |
| TX-002 | retryable body sends email or invokes a model | effect-row rejection; stage outside or use outbox |
| TX-003 | commit response is lost | result is `CommitUnknown`, not rollback or success |
| TX-004 | nested transaction assumes independent commit | semantic/tooling warning exposes savepoint behavior |
| TX-005 | post-commit publication fails | durable outbox/workflow retains recoverable obligation |
| POOL-001 | requests exceed pool capacity | bounded queue sheds or times out with causal evidence |
| POOL-002 | cancelled waiter later receives a lease | lease is not leaked or delivered to a dead scope |
| POOL-003 | task exits while owning a connection | deterministic reset/check-in or quarantine |
| POOL-004 | one tenant saturates every connection | configured fairness/bulkhead policy is observable and enforced |
| POOL-005 | cancel request races query completion | final protocol/transaction outcome remains explicit |

## Admission and stopping rule

This pass does not justify a database-aware compiler, a universal schema
language, or ORM syntax. It justifies four early preservation points:

1. exact presence and dynamic-value distinctions in ordinary ADTs;
2. stable schema/validator/query identities and summaries in typed IR;
3. effect-restricted transaction bodies and affine resource leases;
4. versioned, redaction-aware evidence for every boundary transition.

Those hooks already fit the semantic-kernel direction. Build them through the
kernel and trust-flow experiments; implement JSON and persistence behavior as
the first production-style kit vertical after the native core is trustworthy.

Sources in this note were last verified 2026-09-03. Recommendations are project
synthesis rather than claims made by the cited projects.
