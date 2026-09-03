---
id: semantic-kernel-contract
title: Semantic kernel contract
summary: The smallest executable set of language laws that must agree across the checker, reference interpreter, native lowering, runtime profiles, and agent protocol.
type: design
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-03
tags: [semantics, ownership, effects, failure, ffi, convergence]
related: [design-baseline, type-system-failure-semantics, effects-and-capabilities, trust-validation-information-flow, ownership-lifetime-decisive-study, native-low-level-profile, memory-reclamation-policy, semantic-kernel-probes, compiler-feedback-latency, convergence-audit, convergence-work-program, research-ledger, open-questions]
---

# Semantic kernel contract

## Reader and outcome

This is for the compiler/runtime contributor implementing the first executable
Lang subset. After reading it, they should be able to reject an incompatible
frontend, interpreter, IR, backend optimization, or runtime profile before it
becomes architecture.

This note is a **prototype contract**, not a frozen language specification.
Rules here are deliberately firmer than the surface syntax. A rule changes only
when an executable probe demonstrates that it is incoherent, unusable, or too
expensive—not because a backend happens to prefer different behavior.

## Direct recommendation

Build one strict, expression-oriented, memory-safe core with:

- immutable owned values and explicit local mutation;
- type-declared implicit copying, otherwise ownership transfer;
- shared and exclusive borrows with local control-flow inference;
- lexically scoped resources with restricted automatic release;
- algebraic data, exhaustive matching, `Option`, and typed `Result`;
- non-resumable capability operations and statically resolved providers;
- structured tasks, explicit transfer/share boundaries, and cooperative
  cancellation;
- defined evaluation, arithmetic, initialization, release, and panic behavior;
- raw memory and foreign behavior quarantined behind checked unsafe contracts;
- a deterministic reference interpreter whose observable event trace is the
  oracle for every lowering.

The key narrowing is that v1 does **not** include general multi-shot algebraic
effect handlers. Research shows that a continuation which captures a linear
resource cannot safely be discarded or resumed multiple times without tracking
control-flow linearity. One-shot capability calls retain the architectural
leverage—authority, DI, testing, policy, and telemetry—without making captured
continuations part of the first ownership model. ([Soundly Handling
Linearity](https://doi.org/10.1145/3632896), [Koka effect
handlers](https://koka-lang.github.io/koka/doc/book.html))

## What belongs in the kernel

The kernel owns only behavior whose disagreement would change program meaning
or safety across implementations.

| Kernel law | Why it must be central | What stays outside |
|---|---|---|
| Values and types | Defines valid states and representation-independent behavior | JSON, SQL, HTTP, UUID policy |
| Evaluation order | Orders effects, moves, failure, and cleanup | Formatter layout |
| Ownership and borrowing | Prevents use-after-free, double release, and data races | Exact allocator/collector algorithm |
| Effects and providers | Defines authority and substitution | Concrete database/network/model clients |
| Failure and cancellation | Prevents accidental recovery or lost causes | Retry, circuit, saga policy |
| Resource exit | Determines what happens on return, error, cancel, and panic | Domain compensation |
| Arithmetic and memory validity | Prevents build-mode and backend semantic drift | Specialized numeric packages |
| Text and byte validity | Prevents encoding, indexing, normalization, and FFI ambiguity | Locale, collation, segmentation policy |
| Unsafe and FFI obligations | Bounds the safe-language claim | Header parsers and provider selection |
| Observable trace | Makes interpreter/backend comparison possible | Sampling, export, and retention policy |

Architecture policy, contracts, telemetry projections, actors, durable
workflows, caches, and AI tools build on this kernel. They may be first-party,
but they do not redefine it.

## Semantic result of an execution

At the kernel boundary, executing a closed computation produces exactly one
terminal control outcome plus an ordered trace:

```text
Outcome<T, E> =
  | Returned(T)
  | Failed(E)
  | Cancelled(CancelCause)
  | Panicked(Defect)
  | Terminated(FatalRuntimeFailure)

Execution<T, E> = {
  outcome: Outcome<T, E>,
  events: List<SemanticEvent>,
  live_resources: Set<ResourceId>,
}
```

`Result<T, E>` is a value inside normal execution. `Failed(E)` above means a
top-level `Err(E)` after propagation. Cancellation and panic are intentionally
not error variants that ordinary catch-all code can erase.

Safe code has no undefined outcome. Unsafe code that violates a declared
obligation leaves the language guarantee, and evidence builds should turn every
checkable violation into a precise trap.

## Values and type checking

### Data has nominal meaning

- Public records, variants, identities, validated values, units, trust labels,
  and resource states are nominal.
- Private anonymous records may be structural.
- Small consumer-owned behavioral ports may be satisfied structurally.
- Structural conformance includes parameter ownership, results, errors, and
  effects; extra provider methods are hidden by the narrower port.
- Authority, trust, identity, or representation compatibility never arises from
  structural coincidence alone.
- There is no general null. Absence is `Option<T>`.
- There is no truthiness, arbitrary implicit conversion, or import-dependent
  method meaning.
- A validity, provenance, confidentiality, or integrity fact never grants an
  effect capability. Structural coincidence and remote metadata cannot mint
  authority.

Boundary trust policy is built from nominal values, opaque constructors,
capabilities, and typed-IR flow summaries rather than a hidden Boolean runtime
taint bit. The kernel preserves the facts needed by the [trust and
information-flow boundary](trust-validation-and-information-flow.md); specific
source, sink, classification, validation, encoding, endorsement, and
declassification taxonomies remain versioned policy.

### Inference is local and directional

- Bidirectional checking infers local bindings, literals, lambdas, generics,
  private effects, and local borrow extents.
- Exported operations declare parameters, ownership modes, results, expected
  failures, and authority effects.
- Inference never invents an authority grant, lossy conversion, allocation
  policy, foreign guarantee, public error widening, or cross-task sharing.
- Recursive private groups must be declared together; public component
  dependencies remain acyclic.
- V1 omits general subtyping, impredicative polymorphism, arbitrary type-level
  execution, and overlapping implicit instances.

Bidirectional typing is established as a way to combine synthesis and checking
without requiring complete annotation or global inference. Lang still has to
measure its own diagnostic and incremental behavior. ([Bidirectional
Typing](https://arxiv.org/abs/1908.05839))

### Algebraic data is closed by default

- Matches over closed variants are exhaustive.
- A wildcard over an exported variant must be explicitly marked as accepting
  future cases; the ordinary wildcard does not suppress evolution diagnostics.
- Guards are pure and cannot move, mutate, allocate through a governed
  allocator, suspend, or invoke a capability.
- A scrutinee is evaluated once. Arms and guards are considered top to bottom.
- Unreachable or subsumed arms are errors, not warnings.

### Text is not bytes

- `Bytes` is an arbitrary byte sequence. `Text` is valid Unicode scalar text;
  invalid external encodings produce a typed boundary error.
- The default implementation is UTF-8, but programs do not infer byte offsets
  from character positions.
- APIs name their unit: bytes, Unicode scalar values, or grapheme clusters.
  `Text` has no ambiguous integer indexing and no constant-time character-index
  promise.
- Equality is exact scalar-sequence equality. Normalization, locale-aware
  comparison, case folding, collation, and grapheme segmentation are explicit
  operations with versioned Unicode data.
- A normalized value carries a type/witness such as `Normalized<NFC, Text>`;
  normalization is not silently repeated on assignment, hashing, or equality.
- Foreign/OS strings declare encoding, NUL/sentinel policy, ownership, and
  lifetime at the boundary.

Rust's UTF-8 `str` invariant and byte-counting length show the value of a valid
text type while also showing why a bare “string index” is ambiguous. Unicode
defines several normalization forms and conformance tests rather than one
universal implicit equality rule. ([Rust `str`](https://doc.rust-lang.org/std/primitive.str.html),
[Unicode normalization](https://unicode.org/reports/tr15/))

## Evaluation and sequencing

Lang is strict and eager by default.

1. Subexpressions evaluate left to right as written.
2. Named arguments evaluate in source order, not declaration or alphabetical
   order. The formatter cannot reorder them.
3. A call evaluates its receiver/callee, then arguments, then enters the body.
4. `and` and `or` short-circuit left to right.
5. A pipeline evaluates its input once, then invokes each stage in order.
6. Assignment evaluates to `Unit`; assignment chaining is invalid.
7. `?` inspects one `Result`; on `Err`, it begins an ordinary typed return after
   releasing exited lexical resources.
8. A panic or observed cancellation stops ordinary evaluation and begins the
   corresponding boundary-exit protocol.

Top-level declarations are effect-free. Constants use bounded, deterministic
compile-time evaluation; provider composition, resource acquisition, task
creation, and registration happen in an explicit fallible program/component
startup function. Module import order is never hidden control flow.

Rust now specifies left-to-right evaluation for ordinary multi-operand
expressions; its documented compound-assignment exception is a useful example
of the surprise Lang should avoid. ([Rust expression
order](https://doc.rust-lang.org/reference/expressions.html#evaluation-order-of-operands))

The optimizer may reorder or eliminate work only when the program has the same
allowed outcome, capability-operation order, atomic behavior, explicit bounded
allocation behavior, and resource-exit trace.

## Ownership and mutation

### Default ownership laws

Every runtime value is in one of these states:

```text
owned -> borrowed_shared* | borrowed_exclusive | moved | released
```

- Binding a noncopyable value transfers ownership unless the receiving
  operation borrows it.
- A type is implicitly `Copy` only when its declaration proves that duplication
  is semantically invisible and release-free. Size affects lint/cost evidence,
  not meaning.
- Copying any other value is explicit (`copy` for a declared clone operation,
  with allocation/cost visible in the semantic graph).
- A shared borrow permits reads and other shared borrows.
- An exclusive borrow permits mutation and excludes every other live borrow.
- The owner cannot be moved, released, or mutably accessed while borrowed.
- A borrow ends at its last control-flow use when local inference proves it;
  diagnostics can render the inferred extent.
- Public returned borrows declare their source relationship. Lifetimes are not
  inferred across a public ABI.
- Duplication, abandonment, cross-isolation transfer, concurrent sharing,
  escapability, and address stability are distinct type properties in typed
  IR. Surface names and which are user-declared remain experimental.

Swift's noncopyable and parameter-ownership work demonstrates that declaration-
site `borrowing`/`consuming` modes can control ownership without forcing special
punctuation at ordinary call sites. Rust's non-lexical lifetimes demonstrate
the ergonomic value—and implementation cost—of control-flow-based local borrow
extents. ([Swift parameter ownership](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0377-parameter-ownership-modifiers.md),
[Swift noncopyable types](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0390-noncopyable-structs-and-enums.md),
[Rust NLL](https://rust-lang.github.io/rfcs/2094-nll.html))

### Surface posture to prototype

```text
fn parse(
  input: borrow Bytes,
  using arena: borrow mut Arena,
) -> Result<borrow(arena) Document, ParseError | OutOfMemory>

fn enqueue(message: take Message) -> Result<Unit, QueueFull>
```

`borrow`, `borrow mut`, and `take` are part of the declaration contract. Shared
read calls remain calm. The leading deeper-study candidate makes the two
semantic cliffs—consuming a named noncopyable binding and granting exclusive
mutation—visible at the call site, while temporary and constructor transfers
stay quiet:

```text
let document = parse(input:, arena:)?
enqueue(message: take message)?
```

Exact punning and modifier spelling are not settled. The ownership laboratory
must compare declaration-only, always-visible named transfer/exclusive access,
and ambiguity-only markers. It must reject any design where adding an unrelated
later use silently inserts a deep copy. See the [ownership and lifetime
decisive study](ownership-and-lifetime-decisive-study.md).

### V1 restrictions that buy clarity

- Partially moving an aggregate is rejected. Destructuring consumes the whole
  value, or a dedicated `take_field`-style operation replaces the field with a
  valid value.
- Escaping closures declare every capture as `copy`, `take`, or stable shared
  handle. They cannot borrow a local.
- Ordinary borrows cannot cross `await`. A future experiment may admit a borrow
  held in stable task-owned storage when cancellation and exclusivity are proven.
- Actor messages own their data or carry an explicit immutable shared handle;
  they never borrow sender-local state.
- Address-sensitive values use stable storage explicitly; pinning is not an
  accidental consequence of ordinary async code.

These restrictions are intentionally revisitable after the probe corpus. They
avoid implementing field-sensitive drop flags, self-referential async states,
and distant lifetime diagnostics before their value is established.

### Mutation and sharing

- `let` is immutable; `var` is uniquely owned local mutation.
- Mutable aliasing exists only through an exclusive borrow or an explicit
  synchronization/isolation type.
- Interior mutability is never smuggled through an ordinary shared reference.
- Transfer safety and shared-reference safety are separate compiler-derived
  properties, analogous to Rust's `Send` and `Sync`; unsafe declarations carry
  proof obligations and transitive review evidence. ([Rust `Send` and
  `Sync`](https://doc.rust-lang.org/nomicon/send-and-sync.html))
- Race-free programs receive a sequentially consistent explanation; atomics
  expose an explicit order and are tested against a written happens-before
  model. Go's DRF-SC contract is the simplicity target, while safe Lang rejects
  ordinary data races instead of defining a degraded racy mode. ([Go memory
  model](https://go.dev/ref/mem))

Low-level atomics admit `Relaxed`, `Acquire`, `Release`, `AcqRel`, and `SeqCst`;
the operation determines which orders are legal. The order is a required named
argument, not an inferred optimization hint. High-level locks, channels,
actors, and tasks establish their documented happens-before edges without
exposing atomic orders.

Volatile memory access is a distinct unsafe device/foreign operation. It
preserves required accesses under the declared address-space contract but is
not atomic and creates no inter-thread synchronization. ([Rust atomic
ordering](https://doc.rust-lang.org/std/sync/atomic/enum.Ordering.html),
[Rust volatile read](https://doc.rust-lang.org/core/ptr/fn.read_volatile.html))

## Effects, capabilities, and providers

### V1 effect meaning

An effect row is a set of authority requirements. A capability operation is a
normal, one-shot call to a statically known interface supplied by a provider.
It may return a value or `Result`, suspend if declared asynchronous, and emit a
semantic event. It does not expose or clone the caller's continuation.

```text
fn confirm(command: take ConfirmOrder)
  -> Result<Order, ConfirmError>
  with { Inventory, Payments, Clock }
```

- Public effect rows are explicit and normalized; private rows are inferred.
- Higher-order functions are effect-polymorphic over their callback's row.
- Capability identity, operation identity, caller, provider, result class, and
  causal parent are available to tooling.
- Provider resolution occurs at a composition boundary. It is not a global
  service locator and cannot depend on import order or runtime reflection.
- Candidate resolution is coherent: one binding or a compile error with the
  full candidate path.
- Lexical test overrides may only narrow or replace a declared provider and
  cannot escape their scope.
- A capability value carries a lifetime no longer than its provider and cannot
  be stored or transferred unless its contract permits it.
- Capabilities can be attenuated into smaller consumer-owned ports; they cannot
  be widened implicitly.

The `with` clause declares why `Clock.now()` is legal; the provider graph
supplies `Clock.system()` or `Clock.fixed(...)`. A call-site `!` remains a syntax
experiment and has no semantic role.

### Why not general handlers in v1

| Option | Power | Cost in this kernel | Disposition |
|---|---|---|---|
| Static capability calls | DI, authority, mocking, telemetry, async operations | Provider coherence and lifetime checking | adopt |
| One-shot resumable handlers | Generators, local control abstractions | Stack capture, resource/cancellation semantics | later experiment |
| Multi-shot handlers | Search, backtracking, probabilistic branching | Can duplicate/discard captured owned resources; needs control-flow linearity or copying | reject from v1 |
| General dynamic handler stack | Flexible local override | Handler-order meaning, hidden runtime lookup, optimizer barriers | reject from ordinary DI |

This does not prevent generators, parsers, transactions, retries, actors, or
workflows. They receive direct structured constructs or libraries whose resource
and failure behavior is narrower and easier to inspect.

## Failure, cancellation, and panic

### Expected failure

- Absence uses `Option<T>`.
- Recoverable domain, validation, resource, and external failures use
  `Result<T, E>`.
- An ignored `Result` is an error unless explicitly consumed by a named discard
  operation with rationale.
- `?` propagates only declared typed errors and preserves causal context.
- Public error variants are nominal and closed; translation between layers is
  explicit.
- Retryability, idempotency, and compensation are typed policy facts, not
  inferred from an error name.

### Cancellation

- Cancellation is sticky task control, not a business error and not undo.
- It is observed at explicit cancellation checks and operations whose contracts
  declare a cancellation point. `await` alone does not promise that an effect
  was cancelled or rolled back.
- Leaving a structured task scope cancels unfinished children and waits for
  their bounded cleanup before the scope completes.
- Fail-fast task groups request sibling cancellation on the first defect or
  unhandled failure, then report a stable aggregate keyed by child identity
  after children settle; scheduler race does not choose which cause survives.
- Cancellation cleanup runs under a shield with a declared time/resource
  budget. Exhausting the budget escalates to the owning supervision boundary
  and retains both the cause and unfinished-resource evidence.

Swift's structured-concurrency proposal makes child lifetime structural and
cancellation cooperative rather than implicit rollback. Its later cancellation
shield addresses cleanup that must run after cancellation. ([Swift structured
concurrency](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0304-structured-concurrency.md),
[Swift cancellation shields](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0504-task-cancellation-shields.md))

### Panic and isolation

- `panic` means a defect or broken invariant. Ordinary code cannot catch it.
- A panic terminates the smallest declared isolation boundary. Finite child-task
  panic propagates as parent-boundary panic after sibling cancellation; actor
  panic becomes a typed exit observed by its supervisor.
- A contained panic runs structural release for initialized values and runtime-
  owned resources before the boundary restarts or reports exit.
- A process-root fatal termination does not promise external cleanup. Critical
  correctness must use explicit `Result`, commit, transaction, or compensation.
- Panic payloads are typed defect records, not arbitrary thrown values.
- Panic never unwinds across a foreign ABI. A shim catches and translates it
  where supported or terminates the owning process boundary.
- An abort-only profile can host panic only at its process/root boundary; code
  requiring contained recovery does not compile for that profile.

Rust documents both unwind and abort strategies and the special hazards of
unwinding across FFI. Lang specifies the source boundary behavior and permits
different implementations only when they preserve it. ([Rust panic
semantics](https://doc.rust-lang.org/reference/panic.html))

## Resources and cleanup

### Split release from successful finish

Every owned resource has structural `release`; some also have an explicit,
fallible `finish` or `commit`.

```text
resource OutputFile {
  fn write(self: borrow mut, bytes: borrow Bytes)
    -> Result<Unit, WriteError>

  fn finish(self: take) -> Result<Unit, FlushError>
  release self: take
}
```

- `release` prevents leaks and makes the handle unusable. It is idempotent at
  the runtime boundary and cannot return an application value.
- Automatic release cannot run arbitrary user code, acquire ordinary locks,
  invoke unrelated capabilities, suspend, recurse without a bound, or panic.
- A provider/foreign release primitive may fail operationally; the failure is
  retained as a `ReleaseIssue` event and boundary-health signal. It does not
  overwrite the primary typed result.
- If successful close/flush/commit matters to correctness, the API requires an
  explicit consuming `finish()` whose error the caller handles. Automatic
  release still abandons the resource if `finish` fails or is skipped.
- Resources release in reverse completed-acquisition order on normal return,
  `Err` propagation, observed cancellation, and contained panic.
- Only fully initialized fields release. Partial construction either returns an
  owner of the initialized prefix to structural cleanup or never publishes the
  aggregate.

This avoids the body-versus-cleanup error lottery: there is only one primary
typed outcome, while every release issue remains causally attached. Rust's
custom `Drop` can run arbitrary code and can panic during unwinding, where a
second panic may abort; that is precisely the footgun the restricted release
protocol removes. ([Rust `Drop`](https://doc.rust-lang.org/core/ops/trait.Drop.html),
[Rust destructor order](https://doc.rust-lang.org/reference/destructors.html))

Domain compensation is not release. Releasing an inventory reservation,
reversing a payment, or publishing a saga event is an ordinary visible effect
with its own `Result`, idempotency policy, and retry evidence.

## Allocation and out-of-memory behavior

V1 distinguishes convenience allocation from recoverable governed allocation:

- Ordinary profile allocation is semantically visible as a cost and memory
  requirement. Exhaustion terminates the current configured isolation boundary
  with `OutOfMemory`; it is not catchable as an arbitrary exception.
- Reusable/native code that must recover accepts an allocator or arena
  capability and uses fallible operations returning `Result<_, OutOfMemory>`.
- Freestanding, embedded, and realtime profiles can reject any ungoverned
  allocation statically.
- An optimizer may remove or reuse ordinary allocation when identity is
  unobservable. It cannot remove or reorder an explicit governed allocation if
  that could change success, failure, accounting, or release order.
- Memory pressure, release-queue depth, arena reset cost, and managed-region
  work remain observable operational evidence, not language-level object
  identity.

```text
fn decode(bytes: borrow Bytes, using arena: borrow mut Arena)
  -> Result<Packet, DecodeError | OutOfMemory>
```

This keeps application code usable without lying to system libraries about
fallibility. Zig's allocator guidance and explicit `OutOfMemory` errors are the
primary precedent; the exact Lang surface remains a probe question. ([Zig
language reference](https://ziglang.org/documentation/master/))

Stack exhaustion is a defined `FatalRuntimeFailure`, never memory unsafety or a
catchable domain error. Ordinary profiles may grow, probe, or bound stacks;
realtime, embedded, and interrupt contexts require a static or measured stack
budget and may reject unbounded recursion. Tail-call optimization is never
required for correctness unless a future construct explicitly promises it.

## Numeric, comparison, and collection laws

- Integer runtime types have explicit widths and two's-complement
  representation. A machine-sized `USize` exists only for layout/index/ABI work,
  not as the default domain integer.
- Integer literals are exact and unbounded until checked against a type. The
  default unconstrained integer is `I64`; public APIs never infer it silently.
- Ordinary integer add/subtract/multiply/negate and narrowing conversions are
  checked in every build. Overflow panics as a defect; recoverable algorithms
  use named `checked_*`, `wrapping_*`, or `saturating_*` operations.
- Division by zero, invalid shifts, out-of-range indexing, and invalid enum/tag
  construction have defined defect panics in ordinary operations; explicitly
  checked forms return values such as `Option` or `Result`. Safe code never
  inherits C or LLVM undefined behavior.
- There are no implicit numeric promotions. Conversions state loss policy.
- `F32` and `F64` follow IEEE arithmetic within a declared strict or relaxed
  floating mode. Relaxed mode is explicit in the function/component contract.
- Floating values do not implement ordinary total `Eq`, `Hash`, or `Order`.
  Callers choose IEEE partial comparison, bit identity, total ordering, or an
  approximate domain policy. Rust's `total_cmp` illustrates why ordinary float
  equality and IEEE total order are different. ([Rust `f64`](https://doc.rust-lang.org/core/primitive.f64.html))
- Hash-map iteration order is unspecified and cannot feed canonical serialization,
  builds, tests, or signatures without sorting/canonicalization. Hash seeding is
  an explicit security/determinism policy.

Zig's distinct ordinary, wrapping, and saturating operators demonstrate the
value of keeping arithmetic policy visible, although Lang should prototype
named forms against punctuation. ([Zig integer
semantics](https://ziglang.org/documentation/master/#Integers))

## Layout, validity, unsafe, and FFI

### Safe memory laws

- Every read observes a fully initialized valid value of its type.
- Padding is never a typed value and is initialized before crossing a foreign,
  hashing, comparison, persistence, or cryptographic boundary.
- References are non-null, aligned, live, and provenance-carrying.
- Integer-to-pointer conversion cannot manufacture dereference authority.
- Layout is opaque unless a type opts into a declared `repr` contract.
- Safe enums cannot contain an invalid discriminant; safe booleans have only two
  representations.
- Bounds, alignment, provenance, alias, validity, and target-feature assumptions
  are checked or proven before lowering to an unchecked backend operation.

Pointer provenance is not merely a numeric address: C standards work and LLVM
optimization contracts both rely on origin/alias information that is easy for a
frontend to overpromise. ([WG14 pointer provenance](https://www.open-std.org/jtc1/sc22/wg14/www/docs/n2311.pdf),
[LLVM language reference](https://llvm.org/docs/LangRef.html))

### Unsafe is an obligation boundary

An unsafe operation names the exact obligations it does not check. An unsafe
component records:

- validity, initialization, bounds, alignment, provenance, alias, and lifetime;
- ownership transfer, release, panic/unwind, callback, reentrancy, and thread
  rules;
- atomic order and happens-before claims;
- ABI, layout, symbol version, target feature, and blocking behavior;
- sanitizer, interpreter, fuzz, conformance, and review evidence.

Unsafe does not suppress unrelated type, effect, ownership, or architecture
checks. Its obligations are queryable transitively from safe callers.

### Foreign declarations are quarantined claims

```text
foreign C library zlib {
  fn compress2(
    destination: out Buffer,
    source: borrow BytesView,
    level: CInt,
  ) -> CInt
  abi: platform_c
  retains: nothing
  unwind: forbidden
  blocking: false
  reentrant: false
}
```

- A foreign declaration is unsafe evidence, not a safe wrapper.
- Nullable pointers, lengths, sentinels, ownership, retention, and output
  initialization are explicit.
- `errno` or platform error state is captured before any other foreign call.
- C++/foreign exceptions and Lang panic do not cross a non-unwinding ABI.
- Asynchronous callback registration returns an owned registration resource;
  shutdown consumes it and waits for or invalidates in-flight callbacks before
  captured state can release.
- A dynamic-library handle owns every loaded symbol borrow. Unload is impossible
  while a symbol, callback registration, foreign object, or executing call tied
  to that library remains live.
- A blocking or CPU-long call is routed to a compatible executor or rejected.
- Generated bindings preserve uncertainty and require an audited safe adapter.

Rust's FFI documentation highlights callback lifetime and unwind hazards; Miri
also shows the value and the limits of an interpreter at a native boundary.
([Rust FFI](https://doc.rust-lang.org/nomicon/ffi.html), [Miri](https://github.com/rust-lang/miri))

## Observable behavior and telemetry

The semantic oracle records only events needed to compare meaning:

```text
SemanticEvent =
  | CapabilityRequested(call_id, capability, operation, provider, inputs_digest)
  | CapabilityCompleted(call_id, outcome_digest)
  | ResourceAcquired(resource_id, owner)
  | ResourceReleased(resource_id, reason, issue?)
  | TaskSpawned(task_id, parent_id)
  | CancellationRequested(task_id, cause)
  | CancellationObserved(task_id, point)
  | AtomicOperation(location_id, operation, order)
  | UnsafeBoundaryEntered(obligation_set)
  | PanicRaised(boundary_id, defect)
```

Wall-clock duration, addresses, allocation identities, scheduler interleavings,
and exporter detail are not ordinary semantic outputs. Deterministic test worlds
can add a scheduled-event trace. Production telemetry projects from the same
stable identities but may sample, aggregate, redact, or buffer without changing
program results.

Security/audit effects that the application depends on must be ordinary
capability operations, not best-effort telemetry.

## Backend-independent optimizer contract

The frontend lowers only facts it has proven. In particular:

- never infer LLVM `noalias`, `noundef`, range, alignment, unwind, or target
  attributes from convention;
- never make signed overflow, invalid shifts, uninitialized reads, or races
  backend undefined behavior in safe code;
- preserve check behavior across development and release builds;
- preserve resource-exit and effect sequencing even if pure values are removed;
- validate every IR stage and compare optimized native execution against the
  interpreter;
- use translation validation or reduced counterexamples for high-risk
  optimizations where practical.

LLVM explicitly distinguishes immediate undefined behavior, poison, and undef,
and notes that frontend attributes impose optimizer-visible obligations. That
complexity stays below Lang's portable IR boundary. ([LLVM UB
manual](https://llvm.org/docs/UndefinedBehavior.html))

## Runtime-profile monotonicity

Profiles may remove services or restrict programs; they cannot silently change
the meaning of accepted source.

| Profile | May omit/restrict | Must preserve |
|---|---|---|
| freestanding | heap, scheduler, containment, unwinding, telemetry exporter | value/evaluation/arithmetic/ownership/ABI laws |
| native-core | actors, managed regions, network providers | safe memory, resource, failure, FFI laws |
| application | distributed actors, service exporters | structured task and contained-panic laws it admits |
| service | nothing required by the standard service kit | the entire admitted kernel contract |
| realtime section | allocation, blocking, managed graphs, unbounded release, most effects | declared bounded execution and memory behavior |

If a program needs a missing service—contained panic cleanup, an allocator,
suspension, managed regions, or telemetry export—it fails profile checking. It
does not receive weaker hidden semantics.

## Interaction war game

| Interaction | Failure if designed independently | Kernel answer |
|---|---|---|
| ownership × effects | handler duplicates a continuation holding a file | no multi-shot continuations in v1 |
| ownership × async | borrow survives owner movement/cancellation | ordinary borrow cannot cross `await` |
| ownership × structural ports | provider satisfies shape but consumes differently | ownership modes participate in conformance |
| effects × DI | runtime service locator hides authority and lifetime | coherent static provider graph, scoped overrides |
| effects × optimizer | compiler reorders visible calls | source evaluation and capability order are semantic |
| errors × cleanup | close error erases body failure | primary outcome retained; release issue attached |
| panic × cleanup | user destructor panics during unwind | restricted non-panicking automatic release |
| cancellation × effects | cancellation described as rollback | request and observation are separate; effect contract states outcome |
| OOM × purity | allocation disappears from bounded system code | governed allocation is explicit and fallible |
| concurrency × sharing | immutable-looking wrapper hides mutable pointer | transfer/share derivation stops at unsafe claims |
| FFI × panic | unwind crosses incompatible runtime | forbidden by default; shim contains or terminates |
| floats × maps | NaN breaks equality/hash invariants | floats lack default total `Eq`/`Hash` |
| formatter × named calls | canonical reordering changes effect order | formatter preserves source argument order |
| telemetry × optimization | debugging events freeze every implementation detail | only causal semantic events constrain lowering |
| profiles × semantics | release build disables checks | profiles restrict programs, never weaken accepted rules |

## Rejected early

- General exception throwing/catching for expected failures.
- Arbitrary user code in implicit destructors or finalizers.
- Multi-shot continuations in the first ownership-capable kernel.
- Runtime-reflection DI or import-order provider selection.
- Implicit cloning of nontrivial values.
- Public lifetime, error, effect, or authority inference.
- Partial moves and borrowed locals across suspension in v1.
- Build-mode-dependent overflow, bounds, initialization, or cleanup semantics.
- Ambient nullable pointers or integer-created references in safe code.
- Foreign declarations treated as proof of safety.
- One universal runtime/collector required by every program.
- Telemetry treated as a substitute for correctness-significant audit effects.

## Remaining experimental seams

The contract deliberately leaves these as measured choices:

1. Exact surface words (`take` versus `consume`, visible effect-call marker,
   returned-borrow notation).
2. How much local borrow inference remains predictable on the probe corpus.
3. Whether a one-shot resumable handler earns admission after the direct
   capability model works.
4. Exact ordinary-heap OOM boundary and surface visibility by profile.
5. Precise RC/reuse versus unique/arena representation for immutable values.
6. Managed-region algorithm and safe-point strategy.
7. Strict floating reproducibility contract across targets and accelerators.
8. Whether checked arithmetic uses named operations or dedicated operators.
9. Cranelift, QBE, or another first native backend.
10. Which semantic events are sufficient for differential equivalence without
    constraining valid optimization.

Each seam maps to an executable case in the [semantic kernel probe
suite](semantic-kernel-probes.md). Until then, implementations must expose the
choice rather than smuggling it in as accidental behavior.

## Diminishing-return stop rule

Research on this kernel stops when every consequential rule has primary
precedent or is labeled synthesis, contradictory precedents are bounded, and
the next uncertainty is empirical. That point has been reached for design.

The next evidence must come from the typed core, interpreter, native lowering,
and adversarial probes—not another survey of language features.
