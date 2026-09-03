---
id: semantic-kernel-probes
title: Semantic kernel probe suite
summary: An executable adversarial corpus and evidence contract for falsifying the candidate ownership, effect, failure, resource, arithmetic, concurrency, and FFI semantics.
type: strategy
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-03
tags: [semantics, probes, interpreter, differential-testing, evidence]
related: [semantic-kernel-contract, ownership-lifetime-decisive-study, trust-validation-information-flow, convergence-work-program, compiler-feedback-latency, native-low-level-profile, type-system-failure-semantics, effects-and-capabilities, capability-gauntlet, implementation-path, research-ledger, open-questions]
---

# Semantic kernel probe suite

## Reader and outcome

This is for the first compiler contributor and for an AI agent proposing a
semantic change. After reading it, they should be able to add a kernel feature,
state its expected outcome and trace, run it against the interpreter and native
backend, and know whether the feature survives.

The suite is not a collection of happy-path examples. It is an executable
argument about language meaning.

## Gate-zero deliverable

Experiment 0 is complete only when one command can:

1. parse and type-check every positive probe;
2. reject every negative probe with a stable diagnostic code and primary span;
3. execute valid probes in the reference interpreter;
4. lower the portable subset to one native backend;
5. compare terminal outcome, semantic events, initialized/released resources,
   and declared nondeterminism;
6. emit machine-readable evidence plus a concise human rendering;
7. reduce a mismatch to the smallest still-failing probe.

The initial implementation need not parse the final pretty syntax. A small
explicit kernel notation is preferable to letting parser design block semantic
evidence.

## Corpus record

Every probe is data with a stable identity:

```text
Probe {
  id: "OWN-007",
  claim: "an exclusive borrow prevents shared access until its last use",
  source: "probes/ownership/exclusive_last_use.lang",
  mode: CompileReject,
  expected: {
    diagnostic_code: "ownership.conflicting_borrow",
    primary_symbol: "buffer",
    cause_chain: ["exclusive borrow here", "shared use here"],
  },
  profiles: [NativeCore, Application, Service],
  features: [Borrow, Mutation, ControlFlow],
}
```

Paths are illustrative; stable probe and symbol IDs, not paths or line numbers,
are the durable protocol.

## Machine evidence shape

```json
{
  "schema": "lang.probe-evidence/0",
  "probe_id": "RES-004",
  "compiler": "langc-dev+semantic-hash",
  "profile": "application",
  "frontend": {
    "accepted": true,
    "types": [],
    "effects": [],
    "ownership": []
  },
  "executions": {
    "interpreter": { "outcome": {}, "events": [] },
    "native_dev": { "outcome": {}, "events": [] }
  },
  "equivalence": {
    "status": "match",
    "ignored_operational_fields": ["wall_time", "address"]
  },
  "cost": {
    "wall_ms": 0,
    "cpu_ms": 0,
    "peak_bytes": 0,
    "queries_recomputed": 0
  }
}
```

Machine output is lossless and schema-versioned. Human output leads with the
cause, violated law, and smallest repair. Neither contains raw compiler data
structures as an accidental API.

## The first 24 blocking probes

These are the minimum gate. A native backend is not credible until all 24 agree
with the interpreter.

| ID | Claim under test | Expected evidence |
|---|---|---|
| VAL-001 | Closed variant matching is exhaustive | added variant rejects affected match with missing constructor |
| VAL-002 | Nominal IDs with equal representation are incompatible | compile rejection; no implicit conversion candidate |
| VAL-003 | Structural port checks ownership/error/effect shape | near-match explains every incompatible member dimension |
| EVAL-001 | Calls and named arguments evaluate left to right | ordered capability trace is identical across optimization levels |
| EVAL-002 | `?` releases exited resources before returning `Err` | primary error plus reverse release trace |
| OWN-001 | A noncopyable binding transfers exactly once | use-after-move rejected at use with transfer origin |
| OWN-002 | Shared borrow blocks mutation | conflict spans and inferred borrow endpoint |
| OWN-003 | Exclusive borrow blocks every other borrow | conflict and minimal scope repair |
| OWN-004 | Local last-use inference shortens a borrow | valid program executes without manual scope block |
| OWN-005 | Returned borrow cannot outlive declared source | escape rejected at return/storage boundary |
| OWN-006 | Partial move is unavailable in v1 | diagnostic proposes whole destructure or replacement operation |
| OWN-007 | Escaping closure cannot borrow a local | capture rejection with `copy`/`take`/shared-handle choices |
| OWN-008 | Ordinary borrow cannot cross suspension | rejection names borrow, owner, and suspension point |
| EFF-001 | Undeclared capability is impossible | call rejected with originating dependency chain |
| EFF-002 | Provider resolution is unique and coherent | ambiguity lists candidates and composition paths |
| EFF-003 | Lexical test override cannot escape | lifetime rejection at task/store/return boundary |
| EFF-004 | Capability call cannot duplicate continuation | no continuation value exists in kernel IR/API |
| FAIL-001 | Ignored `Result` is illegal | diagnostic proposes match, `?`, or reasoned discard |
| FAIL-002 | Panic is not catchable as expected failure | catch-like construct rejected; supervisor boundary shown |
| CANCEL-001 | Cancellation is observed only at declared points | deterministic trace distinguishes request from observation |
| RES-001 | Resources release in reverse completed-acquisition order | exact release trace on return, `Err`, cancellation, and panic |
| RES-002 | Automatic release cannot overwrite primary failure | primary `Err` plus attached `ReleaseIssue` |
| NUM-001 | ordinary overflow is identical in every build | same defined defect panic in interpreter/dev/release native |
| FFI-001 | panic cannot cross non-unwinding C boundary | shim translates or process-boundary termination matches contract |

## Extended semantic matrix

### Values, inference, and matching

| ID | Adversary | Pass condition |
|---|---|---|
| VAL-004 | wildcard hides a new exported variant | rejected unless explicitly future-accepting |
| VAL-005 | pattern guard calls `Clock.now()` | rejected as effectful guard |
| VAL-006 | inferred public error/effect widening | implementation rejected against declared signature |
| VAL-007 | overlapping structural implementations | explicit witness required or ambiguity rejected |
| VAL-008 | recursive public component dependency | minimal cycle rendered and rejected |
| VAL-009 | untyped literal exceeds target type | compile-time representability error |

### Ownership and storage

| ID | Adversary | Pass condition |
|---|---|---|
| OWN-009 | branch moves value on one path then joins | valid only when every continuing path owns a value |
| OWN-010 | loop-carried mutable borrow | fixed point converges locally; diagnostic stays at loop |
| OWN-011 | arena view escapes arena reset | compile rejection or explicit owning copy |
| OWN-012 | iterator invalidated by mutation | exclusivity rejection before execution |
| OWN-013 | task spawn captures non-transfer-safe handle | rejection identifies unsafe field preventing transfer |
| OWN-014 | actor message contains sender borrow | rejection proposes own/share serialization boundary |
| OWN-015 | last shared reference releases a million-node chain | correct release count plus p99.9/max queue/stall evidence |
| OWN-016 | cycle formed in non-managed shared values | construction rejected or explicit managed region required |
| OWN-017 | managed-region handle crosses owner | static rejection or typed remote/weak handle only |
| OWN-018 | pinned callback state moved after registration | rejection tied to registration resource |

### Effects and providers

| ID | Adversary | Pass condition |
|---|---|---|
| EFF-005 | two transitive defaults satisfy one capability | no arbitrary order; composition root must choose |
| EFF-006 | provider depends on itself through aliases | normalized cycle rejected with shortest path |
| EFF-007 | narrower port hides provider's stronger authority | caller can invoke only declared operations/effects |
| EFF-008 | higher-order `map` invokes effectful callback | callback row propagates without unrelated effects |
| EFF-009 | foreign function falsely appears pure | foreign boundary requires declared unsafe/effect contract |
| EFF-010 | retry wrapper encloses nested retrying provider | static/policy warning shows multiplicative attempt bound |
| EFF-011 | test override changes unrelated provider | evidence rejects graph delta outside override scope |
| EFF-012 | capability stored beyond provider lifetime | rejected with provider and storage lifetimes |

### Failure, cancellation, panic, and cleanup

| ID | Adversary | Pass condition |
|---|---|---|
| FAIL-003 | catch-all maps defect to success | impossible in ordinary code |
| FAIL-004 | new expected error variant reaches exhaustive caller | caller fails to compile with affected chain |
| FAIL-005 | error translated without original cause | policy rejects missing causal attachment at boundary |
| CANCEL-002 | cancellation races with remote completion | provider contract records sent/completed/observed facts; no rollback claim |
| CANCEL-003 | child ignores cancellation | scope waits only to budget, then escalates with child identity |
| CANCEL-004 | two children fail concurrently | stable aggregate retains both causes independent of schedule |
| PANIC-001 | child task panics while sibling owns resource | sibling cancellation and release precede parent-boundary panic |
| PANIC-002 | actor panics repeatedly | supervisor intensity limit produces typed escalation evidence |
| RES-003 | construction fails after three of five resources | only initialized resources release, in reverse order |
| RES-004 | body returns `Err` and release reports issue | `Err` remains primary; issue is causally attached |
| RES-005 | correctness requires file flush | omission of explicit `finish` violates must-finish state/protocol |
| RES-006 | release tries to allocate, await, panic, or call network | restricted-release checker rejects each operation |
| OOM-001 | fallible arena exhausts | typed `OutOfMemory`; no panic or partial publication |
| OOM-002 | ordinary heap exhausts inside isolated actor | actor exits, owner resources accounted, supervisor informed |

### Numerics, data validity, and deterministic artifacts

| ID | Adversary | Pass condition |
|---|---|---|
| NUM-002 | signed add overflow under release optimization | same defined defect panic as interpreter |
| NUM-003 | wrapping and saturating arithmetic | exact declared result, never implicit fallback |
| NUM-004 | ordinary divide by zero and minimum/-1 | same defined defect panic on all targets; checked form returns its declared value |
| NUM-005 | ordinary shift by negative/width-or-greater count | same defined defect panic; named checked/masking form behaves as declared |
| NUM-006 | narrowing/sign conversion loses information | ordinary conversion rejected; explicit policy succeeds |
| NUM-007 | NaN and signed zero used as map key | no default `Eq`/`Hash`; selected wrapper defines behavior |
| NUM-008 | strict float corpus across backends | allowed bit/result set declared; mismatch reduced |
| DATA-001 | unordered map enters canonical serializer | rejected or explicitly sorted/canonicalized |
| DATA-002 | uninitialized padding crosses FFI/hash boundary | initialization/check or rejection before crossing |
| DATA-003 | invalid UTF-8 becomes `Text` | typed decode failure; unchecked construction remains unsafe |
| DATA-004 | byte/scalar/grapheme index is omitted | ambiguous indexing rejected; explicit unit API succeeds |
| DATA-005 | canonically equivalent unnormalized text compares | exact equality differs until explicit normalization witness |

### Concurrency primitives, volatile access, and stack behavior

| ID | Adversary | Pass condition |
|---|---|---|
| ATOMIC-001 | load/store uses an illegal memory order | compile rejection names permitted orders |
| ATOMIC-002 | message-passing litmus uses release/acquire | interpreter/model checker and native outcomes fit happens-before contract |
| ATOMIC-003 | `SeqCst` litmus across threads | only outcomes allowed by one total atomic order appear |
| VOL-001 | volatile flag used for thread synchronization | static/policy rejection; atomic or lock proposed |
| VOL-002 | MMIO access optimized away or duplicated | semantic/native trace preserves declared volatile operations |
| STACK-001 | recursion exhausts an ordinary profile stack | defined fatal runtime outcome, not memory corruption |
| STACK-002 | unbounded recursion enters realtime/interrupt code | profile/budget rejection before deployment |
| INIT-001 | top-level initializer requests clock/network/file | compile rejection routes work to explicit startup |

### Trust, validation, and information flow

These probes use existing nominal types, capabilities, and typed-IR flow
summaries. They do not imply that a Boolean taint bit or general security
lattice belongs in kernel semantics. The full rationale is in [Trust,
validation, and information flow](trust-validation-and-information-flow.md).

| ID | Adversary | Pass condition |
|---|---|---|
| FLOW-001 | HTTP bytes go directly to a shell program | rejection points from source to raw sink and proposes structured process API |
| FLOW-002 | regex capture is treated as universally trusted | specific nominal validation rule required |
| FLOW-003 | HTML-text encoding enters JavaScript context | context-type mismatch |
| FLOW-004 | parameterized SQL receives hostile string data | accepted as data; generated protocol does not concatenate syntax |
| FLOW-005 | untrusted string selects SQL identifier | allowlisted `SqlIdentifier` construction required |
| FLOW-006 | valid JSON becomes an authorized command | schema decoding alone is insufficient |
| FLOW-007 | validated secret is logged | confidentiality sink still rejects it |
| FLOW-008 | model output proposes a typed destructive call | schema succeeds; authorization rejects missing capability/policy evidence |
| FLOW-009 | untrusted branch selects privileged executable | targeted control-influence path is reported |
| FLOW-010 | validator changes one aggregate field | other fields retain their flow facts |
| FLOW-011 | actor/queue/cache round trip drops provenance | schema preserves or explicitly re-establishes facts |
| FLOW-012 | foreign operation has no flow summary | conservative propagation and explicit release obligation |
| FLOW-013 | dependency claims sanitizer behavior | trusted provenance and conformance evidence required |
| FLOW-014 | declassification omits purpose or authority | rejection; no transition or audit event |
| FLOW-015 | static model misses native propagation | instrumented adversarial build exposes source-to-sink trace |
| FLOW-016 | flow diagnostic contains secret bytes | only redacted shapes and semantic IDs appear |
| FLOW-017 | global analysis explores indiscriminately | policy scope and budget bound work; deferred work is explicit |
| FLOW-018 | cached validation survives policy change | semantic invalidation selects dependent facts, tests, and artifacts |

### Unsafe and foreign boundaries

| ID | Adversary | Pass condition |
|---|---|---|
| FFI-002 | nullable C pointer imported as reference | binding requires `Option<NonNull<_>>` or checked adapter |
| FFI-003 | C retains a borrowed buffer after return | contract/lifetime rejection; owned transfer or registration required |
| FFI-004 | callback fires after captured state release | registration shutdown prevents call or evidence traps violation |
| FFI-005 | callback arrives on undeclared thread | adapter rejects/queues per declared threading contract |
| FFI-006 | foreign exception crosses C boundary | conformance harness reports fatal contract violation, never safe continuation |
| FFI-007 | `errno` read after an intervening call | generated adapter captures it immediately |
| FFI-008 | out pointer only partially initialized | validity check prevents typed value construction |
| FFI-009 | integer-reconstituted pointer lacks provenance | safe dereference impossible; evidence build traps unsafe claim |
| FFI-010 | incorrect `noalias` reaches optimizer | IR validator rejects unproven attribute |
| FFI-011 | CPU-long call marked nonblocking | scheduler watchdog identifies contract violation and provider owner |
| FFI-012 | target-feature intrinsic runs on unsupported CPU | guarded dispatch or defined fatal defect before instruction |

### Optimizer and profile monotonicity

| ID | Adversary | Pass condition |
|---|---|---|
| OPT-001 | optimizer removes pure allocation | allowed only if outcome/accounting contract is unchanged |
| OPT-002 | optimizer reorders two capability calls | semantic trace mismatch catches it |
| OPT-003 | optimizer sinks release past visible effect | semantic trace mismatch catches it |
| OPT-004 | backend poison introduced by checked overflow lowering | differential mismatch or IR validator catches it |
| OPT-005 | profile disables bounds checks | conformance rejects semantic drift |
| OPT-006 | freestanding program requests contained panic | profile checker rejects missing service |
| OPT-007 | realtime section allocates, recurses without a bound, or performs unbounded release | static budget/effect rejection |
| OPT-008 | telemetry sampling changes application audit result | audit must be a capability effect; test rejects telemetry dependence |

## The original twelve ownership workloads

These remain compact cross-feature corpus programs. The later [ownership and
lifetime decisive study](ownership-and-lifetime-decisive-study.md) expands them
into 64 systematically partitioned workloads without changing these stable
fixtures or IDs.

The compact probes above must also appear inside realistic code so a model that
works only on toy terms cannot pass:

1. streaming JSON parser with borrowed slices;
2. arena-backed syntax tree returned from a parser;
3. compiler graph containing cycles;
4. iterator/view invalidation during mutation;
5. escaping closure and callback registration;
6. structured async socket/file lifetime;
7. large actor-message transfer and immutable share;
8. C callback retaining caller data;
9. partial resource initialization;
10. primary failure plus release issue;
11. fallible bounded arena/OOM;
12. intrusive or address-sensitive pinned structure.

Each workload has positive, compile-reject, runtime-defect, cancellation, and
optimized-native variants where meaningful.

## Comparison axes

The suite compares more than yes/no correctness.

| Lens | Measurements |
|---|---|
| semantic | outcome, event order, releases, leaks, allowed nondeterminism |
| diagnostic | primary-span distance, cause completeness, repair applicability, stability |
| AI loop | tokens supplied, repair rounds, wrong-fix rate, semantic queries used |
| human audit | time to answer owner/effect/failure/cleanup questions, reviewer errors |
| compiler | cold/warm wall and CPU, peak memory, invalidation cone, cache bytes |
| runtime | copies, retains, allocations, peak memory, release queue, p50/p99/p99.9/max stalls |
| unsafe | obligation count, transitive reach, sanitizer/fuzzer coverage, escapes |
| portability | interpreter/backend/target/profile conformance and conditional code |

Thresholds begin as recorded observations. They become gates only after the
corpus has realistic scale and repeated measurements with noise retained.

## Generator and reducer strategy

Handwritten examples define intent; generated programs search the seams.

- Generate well-typed programs from the typed core, then mutate ownership,
  effects, order, initialization, and profile facts one at a time.
- Generate paired programs expected to be equivalent under pure rewrites.
- Enumerate small task schedules and weak-memory litmus cases under explicit
  bounds.
- Randomize allocation/layout/target properties in the interpreter without
  exposing them as stable language behavior.
- Reduce failures by typed AST and semantic identity, preserving the violated
  claim rather than only textual parseability.
- Seed every fixed compiler bug into the permanent corpus with its prior bad
  outcome and the smallest reproducer.

Miri provides useful precedent for interpreting a compiler IR to catch
use-after-free, uninitialized data, invalid values, races, and alias violations,
while explicitly acknowledging incomplete FFI and schedule coverage. Lang should
copy the layered-evidence posture, not claim that one interpreter proves native
foreign behavior. ([Miri](https://github.com/rust-lang/miri))

## Differential equivalence

Two executions match when:

- terminal outcome constructors and typed payloads match;
- capability events have the same partial order and causal identities;
- every acquired resource is released exactly once when the contract requires;
- cancellation request/observation and panic boundary agree;
- atomics produce an outcome allowed by the declared memory model;
- strict numeric operations match exactly and relaxed operations remain inside
  their declared result set;
- ignored operational details are named by schema, never silently dropped.

Address, wall time, thread identity, allocator identity, hash seed, and sampled
telemetry may differ unless a probe explicitly promotes them into policy.

WebAssembly's executable validation and small-step execution specification is a
useful standard for stating rules independently of one implementation; it also
models traps explicitly rather than treating every invalid execution as an
ordinary value. ([WebAssembly execution](https://webassembly.github.io/spec/core/exec/))

## CI lanes

```text
edit      parse + affected type/effect/ownership negatives
check     all kernel static probes + interpreter affected execution
verify    full interpreter/native differential + generated/reduced corpus
release   all profiles/targets + sanitizers + FFI + repeated performance tails
nightly   schedule exploration + fuzzing + randomized layouts + backend matrix
```

Every deferred lane produces a named obligation. “Not run” is never rendered as
“pass.” Cache keys include the semantic contract version, target, profile,
backend, optimization mode, and probe generator seed.

## Acceptance and change control

A semantic proposal must supply:

1. the kernel law it adds or changes;
2. positive and negative probes;
3. interactions with ownership, effects, failure, cleanup, cancellation, FFI,
   profiles, and optimization;
4. a structured diagnostic and repair story;
5. interpreter and native evidence;
6. compile/runtime/AI/human cost deltas;
7. a removal or migration path if experimental.

No syntax-only example can accept a semantic feature. No benchmark can accept a
feature whose interpreter and native meaning disagree.
