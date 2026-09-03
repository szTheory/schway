---
id: ownership-lifetime-decisive-study
title: Ownership and lifetime decisive study
summary: A whole-design comparison of ownership models, a recommended v0 contract, realistic edge cases, stakeholder critiques, and the experiment that must falsify it.
type: design
status: candidate
confidence: medium
created: 2026-09-03
updated: 2026-09-03
tags: [ownership, borrowing, lifetimes, resources, memory, experiment]
related: [semantic-kernel-contract, semantic-kernel-probes, ownership-experiment-readiness-audit, native-low-level-profile, memory-reclamation-policy, concurrency-memory, resources-locks-caching, effects-and-capabilities, compiler-feedback-latency, convergence-work-program, language-ecosystem-lifecycle-backcast, research-ledger, open-questions]
---

# Ownership and lifetime decisive study

## Reader and outcome

This study is for the compiler/runtime implementer, library author, systems
programmer, AI-tooling author, and reviewer deciding whether Lang's ownership
model is ready to shape the first implementation.

After reading it, they should be able to:

1. distinguish the independent problems hidden behind the word “ownership”;
2. implement the leading semantic model without treating the surface examples
   as a frozen grammar;
3. predict the result of representative happy paths and edge cases;
4. run the ownership laboratory and know what evidence would reject the model;
5. see which decisions are kernel laws, storage policies, or later extensions.

## Recommendation in one page

Use **affine owned values with temporary exclusive/shared access, explicit
resource obligations, and owner-confined graph regions**.

The human model is:

```text
I own this value.
I may lend read access to many users or exclusive access to one user.
I may transfer ownership, after which I cannot use the old binding.
I must resolve explicit obligations and release resources predictably.
I must opt into sharing, cycles, stable addresses, or raw foreign memory.
```

The compiler model is more precise:

```text
places + initialization states + loans + origin dependencies
  + type abilities + resource obligations + isolation domains
  + pointer provenance + storage policies
```

The key is not to expose every compiler term in ordinary source.

### Leading v0 contract

- Values are owned and immutable by default; `var` grants unique local
  mutation without creating a shared identity.
- Types separately describe whether values may be duplicated, abandoned,
  transferred between isolation domains, shared concurrently, escape a scope,
  or require a stable address. These are independent properties, not one
  `Copy`/“resource” Boolean.
- Private functions may infer parameter access. Public functions and
  structural ports publish `borrow`, exclusive-borrow, and `take` modes.
- A shared borrow permits reads. An exclusive borrow permits mutation and
  excludes every other access for its live extent.
- Local borrow extents end at last use through intraprocedural control-flow
  analysis. Public returned views explicitly identify their source origin;
  public lifetime relationships are never guessed from implementation bodies.
- Ordinary owned results are the default. Borrowed results are nonescapable
  views tied to one or more declared inputs.
- No implicit deep clone. Cheap bitwise/value duplication is type-declared;
  logical cloning and explicit immutable sharing remain visible operations.
- Consuming a named noncopyable binding and granting exclusive access at a
  call site are the leading candidates for visible markers. Shared reads and
  temporary/constructor transfers stay quiet. Exact spelling remains a syntax
  experiment.
- Pattern matching borrows by default. A consuming match is explicit and must
  consume the whole value. V0 has no residual partially moved aggregates.
- Escaping closures explicitly capture by copy, transfer, or stable shared
  handle. They cannot retain a local borrow.
- Ordinary borrows cannot cross `await`, actor send, task detach, storage,
  callback registration, or a foreign retention boundary. Specialized
  task-owned stable views may be reconsidered after measurement.
- Actor/task transfer requires unique ownership or an explicit immutable
  shared form. Borrowed sender-local state never becomes a message.
- Implicit destruction performs only compiler-known, non-failing structural
  work. Files, sockets, locks, mappings, registrations, and foreign handles
  are lexical resources. Commit, flush, durable sync, and other meaningful
  fallible completion remain explicit consuming `Result` operations.
- Unique storage, arenas, precise reference counting/reuse, static storage,
  foreign storage, and owner-confined tracing regions are representations under
  this contract—not separate general-purpose source languages.
- Safe pointers/views carry bounds, provenance, permissions, initialization,
  and origin. Integer-address reconstruction, uninitialized typed values,
  alias promises, and unchecked access live behind audited unsafe components.
- The ownership checker is local and incremental. Package interfaces carry
  ownership summaries. No whole-program lifetime inference is required for an
  edit or for separate compilation.

This model is a synthesis of Rust/Oxide-style affine ownership, Swift/Hylo
value-access ergonomics, Move's separation of copy and discard permissions,
Cyclone/Verona region ideas, Pony's isolation lessons, and Koka/Perceus as a
storage optimization candidate. It deliberately does not inherit any one
language's complete worldview.

## Ownership is not one decision

Five independent axes must compose:

| Axis | Question | Kernel responsibility | Policy/implementation freedom |
|---|---|---|---|
| Value authority | Who may read, mutate, duplicate, transfer, or abandon a value? | access and ability laws | representation and calling convention |
| Reference validity | What storage may a view reach, and how long is it usable? | origin, escape, and exclusivity laws | static proof versus approved dynamic check |
| Storage/reclamation | Where does data live and when is memory reusable? | observable release and OOM contract | stack, inline, arena, unique heap, RC/reuse, managed region |
| Resource protocol | Who must close, finish, unregister, unlock, or roll back? | obligation state and cleanup order | provider implementation and cleanup scheduling |
| Concurrency isolation | What may cross tasks, actors, threads, interrupts, and devices? | transfer/share/isolation laws | scheduler, mailbox, pool, and machine implementation |

This decomposition prevents two recurring category errors:

- selecting reference counting does not decide aliasing or mutation safety;
- selecting ownership does not decide whether every immutable value is stored
  uniquely, inline, in an arena, or behind a shared representation.

## Design requirements

The model must satisfy all of these together:

1. memory and data-race safety in ordinary code;
2. usable native, embedded, realtime, FFI, and data-oriented programming in
   the first implementation;
3. a calm application surface that does not require lifetime algebra for
   routine domain and service code;
4. deterministic, inspectable resource cleanup and bounded latency evidence;
5. fast local/incremental checking and stable package summaries;
6. explicit semantic cost—no refactor-triggered hidden deep copies;
7. structured concurrency and actor transfer without sender-local borrows;
8. first-class effects/capabilities without duplicating owned resources through
   unrestricted continuations;
9. precise machine-readable diagnostics and explanations for AI repair;
10. narrow unsafe and foreign boundaries with transitive proof obligations;
11. a migration path for richer borrowed views and managed graphs without
    replacing the core IR;
12. one mental model across profiles even when the physical memory strategy
    changes.

## Five genuinely different models

These alternatives expose different concepts; they are not spelling variants.

### Model A — explicit affine references

Every nontrivial API exposes owned, shared-borrowed, and exclusive-borrowed
references; lifetime/origin relationships appear wherever necessary.

```text
fn slice<'data>(data: &'data Bytes, range: Range) -> &'data Bytes
fn append(target: &'a mut Buffer, source: &'b Bytes)
fn send(message: Message)
```

**What it hides:** almost nothing about aliasing. Storage and allocation can
still vary.

**Strongest at:** low-level expressiveness, separate compilation, optimizer
facts, and making unsafe preconditions precise.

**Weakest at:** visual noise, higher-order APIs, graph-shaped application code,
and diagnostics when compiler regions—not domain intent—become the explanation.

**Lesson:** Rust's non-lexical lifetimes show that flow-sensitive local extents
remove large amounts of lexical ceremony. Polonius also illustrates the
algorithmic danger: borrow analysis has distinct initialization, liveness, and
loan phases, and naive formulations can be too slow. ([Rust NLL](https://rust-lang.github.io/rfcs/2094-nll.html),
[Polonius rules](https://rust-lang.github.io/polonius/rules.html),
[Polonius status](https://rust-lang.github.io/polonius/current_status.html))

**Disposition:** use this precision in typed IR and unsafe/public boundaries,
not as the universal human vocabulary.

### Model B — mutable value semantics and projections

Programs speak about independent values. Temporary projections grant read or
modify access but cannot become arbitrary stored references.

```text
subscript longer(left: borrow Text, right: borrow Text)
  -> project Text

modify longer(left:, right:) using emphasize
```

**What it hides:** explicit lifetime parameters and much reference identity.
The projection is an access path, not a freely storable pointer.

**Strongest at:** calm high-level code, in-place mutation, local reasoning, and
preventing reference-shaped APIs from spreading.

**Weakest at:** FFI pointers, intrusive structures, callbacks, iterators that
outlive a call, self-references, and APIs that genuinely store borrowed views.

Hylo demonstrates this family: all ordinary types have value semantics and
subscripts project temporary access without lifetime annotations. Swift's
ownership evolution demonstrates the pressure that returns: noncopyable
generics, nonescapable values, lifetime-dependent spans, yielding accessors,
and borrowing iteration each need explicit language and library support.
([Hylo introduction](https://hylo-lang.org/introduction/),
[Swift nonescapable types](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0446-non-escapable.md),
[Swift borrowing iteration](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0516-borrowing-sequence.md))

**Disposition:** make this the ordinary source mental model, but preserve
origins and nonescapability in IR from day one so low-level views are not a
retrofit.

### Model C — linear capabilities everywhere

Values or function arrows state exactly-once use. I/O, allocation, authority,
and resource state are threaded as linear values.

```text
fn read(file: File, buffer: Buffer) -> (File, Buffer, Result<Count, ReadError>)
fn close(file: File) -> Unit
```

**What it hides:** almost no resource flow. The type system directly prohibits
duplication and, under strict linearity, abandonment.

**Strongest at:** scarce assets, protocol state, authority, verified resource
accounting, and explicit effects.

**Weakest at:** branching, error recovery, ordinary collections, closures,
application ergonomics, and APIs that must repeatedly return updated tokens.

Austral implements capabilities as linear types. Move usefully separates
`copy` and `drop` abilities, showing that duplication and abandonment are
different questions. Linear Haskell attaches multiplicity to arrows rather
than splitting every data type into linear/nonlinear twins. ([Austral
specification](https://austral-lang.org/spec/spec.html), [Move
abilities](https://move-language.github.io/move/abilities.html), [Linear
Haskell](https://arxiv.org/abs/1710.09756))

**Disposition:** reject universal strict linearity. Adopt affine ownership for
ordinary noncopyable values and explicit must-resolve obligations only where
abandonment would be a semantic bug.

### Model D — region and isolation ownership

An owner controls a whole object graph. Internal aliasing is flexible; crossing
the region boundary is restricted to moves, immutable snapshots/shares, or
checked handles.

```text
within region request_memory {
  let document = parse(bytes:, allocate_in: request_memory)
  render(document:)
}

actor.send(take graph_region)
```

**What it hides:** per-object lifetime tracking within a region and, if
managed, cyclic reclamation details.

**Strongest at:** parsers, compiler IR, request data, actor heaps, mutable
graphs, bulk release, ownership attribution, and compartmentalization.

**Weakest at:** independently escaping objects, long-lived cross-region edges,
fine-grained reclamation, and region-retention spikes.

Cyclone combined lexical regions, unique pointers, and reference-counted
objects rather than forcing one mechanism. Verona investigates linear regions
and compartmentalization; Pony uses reference capabilities and recovery to
move or freeze isolated graphs, but Pony's own learning material acknowledges
that reference capabilities are its most difficult concept. ([Cyclone safe
memory management](https://www.cs.umd.edu/users/mwh/papers/hicks03safe.html),
[Project Verona](https://www.microsoft.com/en-us/research/project/project-verona/),
[Pony reference capabilities](https://www.ponylang.io/learn/reference-capabilities/))

**Disposition:** adopt explicit arenas immediately and preserve an
owner-confined managed-region seam. Do not expose a large reference-capability
matrix in ordinary types.

### Model E — runtime-checked references with compiler optimization

References carry runtime generation, bounds, reference counts, or dynamic
exclusivity metadata. Static analysis removes checks when it can.

```text
let view = object.checked_view()
view.read() // traps if generation or dynamic loan no longer matches
```

**What it hides:** some static lifetime restrictions and ownership annotations.

**Strongest at:** dynamic graphs, incremental migration, and cases where a
static checker would conservatively reject a safe execution.

**Weakest at:** hard realtime, pointer size/cache behavior, proof portability,
failure timing, and making all invalid programs fail before deployment.

Swift uses runtime exclusivity checks for cases static analysis does not prove.
Vale proposes generational references as a different memory-safety tradeoff,
but its performance and mature systems coverage are not sufficient evidence
for Lang to make this the universal base. ([Swift exclusivity](https://www.swift.org/blog/swift-5-exclusivity/),
[Vale generational references](https://vale.dev/vision/safety-generational-references))

**Disposition:** use dynamic checking in the interpreter, sanitizer, foreign
guards, or deliberately dynamic containers—not as the default native pointer
or a substitute for the static ownership contract.

## Why the hybrid is coherent

The hybrid assigns each model the job it does best:

```text
ordinary source          mutable value semantics and access modes
typed ownership IR       affine places, loans, origins, obligations
native resources         unique ownership and lexical release
phase-shaped memory      arenas/regions
immutable multi-owner    explicit share + measured RC/reuse candidate
cyclic identity graph    opt-in owner-confined managed region
concurrency              unique transfer or immutable share
foreign/raw memory       bounded provenance-carrying views + audited unsafe
development evidence     dynamic interpreter/sanitizer checks
```

It remains one model because every lane obeys the same authority law: mutation
requires unique access, borrowed access cannot outlive its origin, transfer
ends the sender's authority, and resource obligations have one owner.

## Proposed semantic vocabulary

Names below are semantic roles, not frozen keywords.

### Type abilities

| Property | Meaning | Default posture |
|---|---|---|
| duplicable | creating another independently usable value is legal and semantically defined | derived only for declared types |
| discardable | a value may leave scope without an explicit application operation | most ordinary values; not unresolved obligations |
| transferable | unique ownership may cross an isolation/foreign boundary | derived structurally with negative witnesses |
| shareable | concurrent immutable aliases are legal | explicit for immutable/thread-safe representations |
| escapable | the value may outlive the current lexical/access scope | ordinary owned values yes; borrowed views no |
| address-stable | the storage will not move until a named owner/resource ends | explicit stable-storage owner only |

These properties compose structurally through fields and generic parameters,
with explicit unsafe witnesses only for implementation-defined interiors.
Move's independent `copy` and `drop` abilities are the key precedent for not
collapsing them. Swift's noncopyable generics show why generic containers and
protocols must model noncopyability early, rather than assuming every type can
be copied. ([Move abilities](https://move-language.github.io/move/abilities.html),
[Swift noncopyable generics](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0427-noncopyable-generics.md))

### Parameter and result modes

```text
borrow T          shared temporary access; no retention or mutation
borrow mut T      exclusive temporary access; no competing access
take T            callee receives ownership
T                 owned result by default
borrow(source) T  nonescapable result whose validity depends on source
```

Public signatures and structural ports expose modes. Private declarations may
infer them, but the formatter/editor/agent projection can render the inferred
contract. Ownership modes participate in function type compatibility and API
fingerprints. Swift explicitly notes that changing `borrowing`/`consuming`
conventions affects ABI; Lang should treat such changes as compatibility
events even before committing to a stable native ABI. ([Swift parameter
ownership](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0377-parameter-ownership-modifiers.md))

### Origins, not user-managed lifetime algebra

The IR associates every borrowed view with an origin set and permitted access:

```text
origin(view) = {input}
access(view) = shared
escape(view) <= scope(input)
```

Most local origins are anonymous and inferred. Public APIs name source
relationships with values/parameters, not compiler-generated region variables.

```text
fn header(frame: borrow Frame) -> borrow(frame) HeaderView

fn choose(
  left: borrow Buffer,
  right: borrow Buffer,
  selector: Side,
) -> borrow(left | right) BytesView
```

The second result is usable only while both potential sources remain valid
unless the type system can preserve a path-sensitive result variant. An API
that returns `LeftView | RightView` may retain more precision than merging
origins. Oxide's view of lifetimes as approximated reference provenance is the
formal starting point; Swift's lifetime-dependent `Span` work is the practical
API pressure test. ([Oxide](https://arxiv.org/abs/1903.00982), [Swift Span
properties](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0456-stdlib-span-properties.md))

### Checker architecture

Per function, use a small sequence of monotone analyses over typed control
flow:

1. definite initialization and move state for every place/path;
2. origin construction and projection;
3. liveness of borrowed values;
4. loan propagation, conflict, and escape checking;
5. obligation completion and release-path checking;
6. isolation transfer/share checking;
7. generation of an ownership summary and evidence.

Run a cheap conservative screen first. Functions it proves safe skip the more
precise analysis; potential conflicts receive the precise analysis before any
user-facing rejection. The fast lane must not silently accept something the
release lane rejects under the same declared assurance level.

Package checking consumes signatures/summaries, not dependency bodies.
Changing a private body without changing its ownership/effect/type summary
must not invalidate downstream packages. Whole-program escape analysis may
improve stack placement or remove RC operations, but never changes whether the
source is legal.

Polonius explicitly separates initialization, liveness, and loan analyses and
has experimented with precision/performance grades. Cyclone's region work
emphasized intraprocedural analysis and separate compilation. These support the
architecture, not a claim that Lang's latency target is already met.
([Polonius](https://rust-lang.github.io/polonius/rules.html), [Cyclone
regions](https://www.cs.cornell.edu/projects/cyclone/papers/cyclone-regions.pdf))

## Realistic examples and expected behavior

The syntax is illustrative. Each example is a semantic fixture first.

### Pure domain code stays quiet

```text
record Money {
  cents: I64,
  currency: Currency,
}

fn total(lines: borrow List<LineItem>) -> Money {
  lines.fold(initial: Money.zero(currency: USD), using: add_line)
}
```

`lines` is read-only temporary access. `Money` can be passed in registers or
inlined regardless of the source-level value semantics. No heap strategy is
implied.

### Unique mutation is visible but local

```text
fn normalize(bytes: borrow mut Buffer) -> Unit {
  for index in bytes.indices() {
    bytes[index] = ascii_lower(bytes[index])
  }
}

var payload = Buffer.from(bytes: input)
normalize(payload:) // candidate surface renders exclusive access here
send(payload:)      // candidate surface renders transfer here
```

After transfer, `payload` is uninitialized. There is no implicit clone because
a later line attempts another use.

### Borrowed parse result from an arena

```text
fn parse(
  input: borrow Bytes,
  using arena: borrow mut Arena,
) -> Result<borrow(arena) Document, ParseError | OutOfMemory>
```

Happy path: the returned `Document` keeps an origin-dependent loan on the arena;
reset or mutation that could invalidate its storage is rejected while the view
is live. Error path: initialized arena objects are structurally released or
reclaimed by the arena contract. A document cannot be returned from the
arena's scope, sent to an actor, or stored in a longer-lived object without an
owning copy/share/snapshot.

### Borrowing and consuming pattern matches

```text
data Connection =
  | Open { socket: Socket, peer: Address }
  | Closed { reason: CloseReason }

match connection { // borrows
  Open(socket:, peer:) => inspect(socket:, peer:),
  Closed(reason:) => report(reason:),
}

match take connection { // consumes the entire value
  Open(socket:, peer:) => hand_off(socket:, peer:),
  Closed(reason:) => archive(reason:),
}
```

V0 does not allow moving `socket` while leaving a residual `connection` whose
destruction requires hidden field-state flags. Swift's later addition of
borrowing versus consuming matching for noncopyable variants is evidence that
this must be designed with pattern matching, not bolted on afterward. ([Swift
borrowing/consuming switch](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0432-noncopyable-switch.md))

### Escaping closure capture

```text
use registration = Events.on_message(
  capture: { take decoder, share metrics },
  fn(message:) {
    decoder.decode(message:)
    metrics.increment(name: "decoded")
  },
)?
```

The registration owns the closure and therefore the transferred decoder and
shared metrics handle. Unregistering destroys the closure before releasing the
foreign/event source. Capturing `borrow decoder` is rejected because the
callback escapes.

### Structured async resource

```text
use connection = Network.connect(address:, deadline:) ?

within Task.group(failure: cancel_siblings) {
  let reader = Task.start(capture: { borrow connection }) {
    connection.read_frame(deadline: inherit)
  }
  reader.await()?
}
```

The task is lexical and cannot outlive `connection`, so a specialized
structured borrow may be expressible. A general future stored elsewhere may
not retain the borrow. The conservative v0 rule may initially require moving a
session/lease owned by the group instead; this example is intentionally a
decisive pressure case rather than assumed-valid syntax.

### Actor transfer and explicit sharing

```text
let frame = Camera.capture()?
renderer.send(frame: take)?

let palette = Shared.freeze(Palette.load(path:)?)
for worker in workers {
  worker.send(palette: share palette)?
}
```

The large frame is transferred without copy. The immutable palette has
multiple explicit owners; its representation and retain/release costs appear
in cost evidence. A borrowed slice of `frame` cannot be sent.

### Fallible completion remains explicit

```text
use file = Files.create(path:) ?
file.write(bytes:) ?
file.sync() ?
file.finish() ?
```

Scope exit attempts fallback release on success, error, cancellation, or
panic. `sync` and `finish` are explicit because durability or protocol
completion failure matters. If the body and fallback cleanup both fail, the
primary failure remains primary and the cleanup issue is attached.

### FFI retention and callback lifetime

```text
foreign C fn register_callback(
  context: stable borrow CallbackState,
  callback: CCallback,
) -> Result<Registration, CError>
retains context until Registration.release
unwind forbidden
thread callback_executor
```

The returned `Registration` owns the retention obligation. `CallbackState`
must live in stable storage, cannot move while registered, and must meet the
declared thread/reentrancy contract. A generated header supplies only a
candidate contract; sanitizer, ABI, and hostile callback evidence remain
required.

### Partially initialized native storage

```text
unsafe module packet_builder
  proves { initialization, bounds, release_on_error }
{
  fn build(into: borrow mut Uninit<Packet>, source: borrow Bytes)
    -> Result<Init<Packet>, DecodeError>
}
```

Uninitialized bytes are not a `Packet`. The transition to `Init<Packet>` is
one-way and requires every field invariant. Error paths release only fields
proved initialized. Rust's `MaybeUninit` documentation shows how validity and
partial initialization interact with layout and destruction. ([Rust
`MaybeUninit`](https://doc.rust-lang.org/core/mem/union.MaybeUninit.html))

### Cyclic graph

```text
within managed region ui owned by window_actor {
  let root = UiGraph.build(spec:)
  root.bind_bidirectional_controls()
}
```

Cycles exist only inside the named owner. An outside value receives an actor
address, immutable snapshot, or checked weak/remote handle—not a direct mutable
reference. The experiment compares this against indexed arenas and weak edges;
managed tracing is admitted only if it wins on real graph workloads.

### Raw pointer provenance

```text
unsafe fn device_slice(
  base: DeviceAddress,
  count: USize,
) -> VolatileSlice<U32>
proves {
  address_space: mmio,
  bounds: count * size_of<U32>,
  alignment: align_of<U32>,
  provenance: platform_device_map,
}
```

A numeric address does not automatically regain permission to access memory.
Rust's strict-provenance APIs and CHERI's capability model both make bounds,
permissions, and origin materially distinct from the numeric address.
([Rust pointer provenance](https://doc.rust-lang.org/core/ptr/index.html),
[CHERI specification](https://github.com/CTSRD-CHERI/cheri-specification/blob/main/chap-intro.tex))

## Edge-case law matrix

| Case | Leading v0 result | Why |
|---|---|---|
| assign a cheap duplicable scalar | duplicate value | no identity or release observable |
| assign a noncopyable buffer | transfer, old place uninitialized | avoids hidden clone/RC decision |
| pass owned buffer to `borrow` parameter | temporary shared loan | callee cannot retain or mutate |
| pass same value as shared and exclusive in one call | reject unless evaluation completes shared use before exclusive activation under a narrowly specified rule | avoids ambiguous action at a distance |
| borrow two provably disjoint fields | allow | place paths prove non-overlap |
| borrow two dynamic indices | reject or use checked `split` API | general disjointness needs runtime proof |
| mutation after last shared use | allow | nonlexical local loan end |
| branch moves on one path and the value is dead after the join | allow | no later access or implicit release needs an initialized value |
| branch moves on one path and the value is later used or released | reject or use an explicit state variant | ownership state must agree on every path where the place remains live |
| loop conditionally moves a carried value | require state variant/`Option`-like model | fixed-point ownership must be explicit |
| return view from one input | explicit source origin | separate compilation remains sound |
| return view from either of two inputs | union/origin-set or result variant | caller must preserve every possible source |
| store borrowed view in owned heap object | reject unless object is nonescapable and origin-parameterized | owner could outlive source |
| borrow across ordinary `await` | reject in first model | suspension, cancellation, and movement complicate exclusivity |
| lexical child borrows parent resource | candidate specialized allowance | child cannot outlive group; must prove cleanup order |
| detached task captures local | copy, transfer, or shared stable handle only | no parent lifetime relation remains |
| actor message contains borrowed field | reject transitively | sender storage cannot back receiver message |
| self-referential object | construct in stable storage through safe builder | moving after self-reference would invalidate it |
| address-stable value is moved | ownership of stable handle may move; referent does not | separate handle movement from storage movement |
| partial move from value with release logic | reject | prevents hidden, path-sensitive destructor state |
| consuming whole destructure | allow when every field/result is accounted for | aggregate no longer exists |
| shared immutable cycle | reject strong cycle or require managed owner | plain RC would leak |
| last RC release has deep cascade | correct but budgeted/deferred outside critical regions | determinism alone does not guarantee latency |
| arena reset with live view | reject or delay reset through owned guard | view validity is tied to arena generation/scope |
| arena reset has huge release work | measure and warn/restrict in critical region | bulk release can produce a tail spike |
| resource ignored at scope exit | compiler performs guaranteed release attempt | lexical ownership is the fallback |
| must-resolve value ignored | reject with legal terminal operations | abandonment violates protocol, not merely memory |
| `finish` fails then fallback close fails | preserve both, finish primary | no last-error-wins loss |
| panic during partial initialization | release initialized fields only | definite-init state controls cleanup |
| panic crosses `extern C` non-unwind boundary | shim contains or terminates declared boundary | foreign ABI cannot observe Lang unwind |
| raw pointer converted to integer and back | unsafe exposed-provenance operation or reject | address is insufficient authority |
| ordinary shared view mutates through hidden cell | impossible without explicit cell/synchronization type | optimizer and reviewer may rely on immutability |
| mutex guard crosses `await` | reject by default | avoids deadlock and scheduler retention |
| effect handler resumes twice while holding resource | impossible in v1 static capabilities | duplicate continuation would duplicate/lose affine state |
| generic collection instantiated with noncopyable element | works when operations state required abilities | no implicit `Copy` assumption in generic core |
| dynamic port object owns noncopyable state | existential carries destroy/move metadata; cloning unavailable unless declared | type erasure cannot erase ownership law |
| foreign code retains a non-retained buffer | rejected adapter contract or explicit copy/stable registration | lifetime crosses call boundary |
| OOM in reusable/native allocation | typed failure | caller controls budget/policy |
| service default allocation OOM | configured smallest isolation termination may be allowed | profile policy does not rewrite library semantics |

## Interaction map

### Ownership × effects

Capabilities themselves have owner/provider lifetimes. A provider may lend an
operation authority, transfer a scoped token, or return a resource. General
multi-shot handlers remain outside v1 because a continuation can capture an
owned value and then be duplicated or discarded. Research on control-flow
linearity shows this combination can be modeled, but it is not free complexity.
([Soundly Handling Linearity](https://www.research.ed.ac.uk/files/407801113/Soundly_Handling_TANG_DOA07112023_VOR_CC_BY.pdf))

### Ownership × failure and cancellation

Expected errors use ordinary control flow and unwind lexical resource scopes.
Panic/defect containment still releases language-owned resources according to
the selected panic strategy. Cancellation is sticky task control; cleanup is
bounded and shielded narrowly. None of these imply rollback of an external
effect.

### Ownership × structural typing

A structural port match includes parameter ownership, returned-origin
relationships, abilities, errors, and effects. A function that consumes a
value does not satisfy a port that promises only to borrow it. A returned
owning value does not automatically satisfy a borrowed-result contract if the
cost/identity semantics differ.

### Ownership × trust flow

Ownership proves who may access storage; it does not prove that data is valid,
authorized, confidential, or safe for a sink. A uniquely owned malicious
command is still malicious. Conversely, linear capabilities are useful because
they join authority with nonduplication, but endorsement and declassification
remain separate operations.

### Ownership × AI tooling

The compiler should emit a compact ownership explanation graph:

```json
{
  "code": "ownership.escape",
  "value": "header",
  "mode": "borrowed_shared",
  "origin": ["frame"],
  "origin_ends_at": "scope:decode_request",
  "escape": "return",
  "repairs": [
    {"kind": "return_owned", "cost": "copy 48 bytes"},
    {"kind": "extend_owner", "effect": "changes API lifetime"},
    {"kind": "return_origin_variant", "effect": "preserves zero-copy"}
  ]
}
```

The human projection leads with the domain value and violated relationship;
the machine projection includes place/origin IDs, control-flow witnesses,
estimated copy/share costs, and transactional semantic edits. Ownership
visualization research and Rust's diagnostic structure both support showing
the time/path relationship instead of only reporting the final use site.
([Grounded conceptual model](https://arxiv.org/abs/2309.04134), [rustc
diagnostics](https://rustc-dev-guide.rust-lang.org/diagnostics.html))

## Stakeholder gauntlet

| Lens | What this specialist will challenge | Required answer/evidence |
|---|---|---|
| language/type theorist | Is the calculus sound with subtyping, generics, closures, effects, and recursion? | small formal core, progress/preservation or equivalent, executable correspondence |
| borrow-checker engineer | Does inference terminate locally and incrementally? | query graph, loop fixed points, adversarial complexity series, no downstream body reads |
| optimizer/backend engineer | Which alias, provenance, alignment, and initialization facts are legal to emit? | written IR semantics and translation validation; never infer stronger backend attributes |
| systems/embedded engineer | Can memory, layout, allocation, OOM, MMIO, interrupts, and cleanup be controlled? | no-runtime native probes, allocator/arena APIs, explicit unsafe proof obligations |
| realtime engineer | Are allocation and release costs bounded at the critical point? | critical-region effect/cost gate and p99.9/max stall evidence; no hidden RC cascade |
| application/library author | Can routine APIs avoid lifetime notation and work with noncopyable generics? | public access modes, ordinary calm calls, borrowed collection/iterator corpus |
| concurrency/runtime engineer | Can values move/share safely without a global heap or global lock? | structured-task/actor transfer, isolation-region proofs, RC contention and mailbox evidence |
| FFI/platform engineer | Can C ownership, retention, callbacks, unwind, layout, and dynamic unload be represented? | generated candidate bindings plus explicit adapter contracts and hostile ABI tests |
| security engineer | Does unsafe remain small, reviewable, and incapable of laundering false guarantees? | transitive unsafe ledger, source/sink trust separate from ownership, sanitizer/formal evidence |
| SRE/performance engineer | Can allocation, copying, sharing, retention, and release tails be attributed? | per-owner metrics, resource inventory, budget breach events, load/fault/OOM profiles |
| AI-eval/tooling engineer | Can an agent repair violations without receiving the entire compiler state? | stable codes, minimal cause graph, semantic repair actions, hidden multi-model tasks |
| human reviewer | Can ownership/cost be read in a diff without an IDE? | public modes, transfer/exclusive markers, canonical formatting, timed audit |
| ecosystem/evolution owner | Which ownership changes break source, ABI, or serialized/foreign contracts? | separate fingerprints, edition migrations, ecosystem impact builds |
| verification engineer | Can the interpreter, checker, and native backend disagree observably? | semantic events, reducers, generated adversaries, independent certificate checks |

No single lens can veto every ergonomic choice, but any soundness, native
expressiveness, incremental-checking, or runtime-latency failure is a hard veto.

## The ownership laboratory

The previous twelve workloads were correct but too coarse. The decisive study
uses six suites, each with positive, compile-reject, runtime-adversarial, and
repair tasks.

### Suite 1 — local values and control flow

1. scalar and small-record duplication;
2. large buffer transfer and explicit clone;
3. last-use borrow shortening;
4. shared/exclusive conflict;
5. disjoint fields and dynamic indices;
6. branch move/join;
7. loop-carried optional ownership;
8. borrow and consuming matches;
9. whole destructure and prohibited partial residual;
10. nested method/call evaluation requiring two-phase access.

### Suite 2 — views, collections, and generics

11. slice returned from one source;
12. view selected from two sources;
13. arena-backed AST and reset;
14. iterator invalidation by mutation;
15. borrowing iteration over noncopyable elements;
16. yielding/projection accessor;
17. generic `Option`, `Result`, list, map, and iterator with noncopyable types;
18. higher-order `map`/`fold` with borrowed and consuming callbacks;
19. trait/port object with noncopyable erased state;
20. recursive persistent tree with explicit immutable sharing.

### Suite 3 — resources, failures, and suspension

21. file/socket/lock reverse cleanup order;
22. acquisition failure halfway through a resource stack;
23. body error plus cleanup error;
24. explicit commit/finish then fallback release;
25. panic during partial initialization;
26. cancellation during acquisition, body, and cleanup;
27. lexical child task borrowing a resource;
28. rejected detached borrowed capture;
29. zero-copy async parser pressure case;
30. lock guard across suspension rejection.

### Suite 4 — concurrency and graphs

31. large unique actor message transfer;
32. explicit immutable share across many actors;
33. nested non-transferable field witness;
34. Swift-style disconnected-region transfer case;
35. actor-local cyclic graph;
36. indexed-arena graph alternative;
37. cross-owner weak/remote handle;
38. atomic/shared-memory escape;
39. RC contention and false sharing;
40. million-node last-reference release and actor termination.

### Suite 5 — native, FFI, and representation

41. C borrowed input buffer;
42. C-owned returned allocation with matching release;
43. foreign call retaining input;
44. callback registration/unregistration;
45. callback after library unload adversary;
46. panic/unwind at C and C++ boundaries;
47. pinned/self-referential future;
48. intrusive list and DMA buffer;
49. initialized/uninitialized arrays with failing element construction;
50. pointer arithmetic, provenance, integer round-trip, alignment, and bounds;
51. unions/tag validity and `repr(C)` layout;
52. custom allocator, arena exhaustion, and OOM policy.

### Suite 6 — whole-program ergonomics and evolution

53. streaming JSON parser/cleaner;
54. HTTP request body and connection pool lease;
55. compiler IR transform with arena and graph phases;
56. emulator or renderer frame loop;
57. embedded driver with MMIO/interrupt boundary;
58. plugin/dynamic library handle and reload;
59. public API changes from borrow to take and owned to borrowed result;
60. mixed-version packages using ownership summaries;
61. AI repair from minimal structured diagnostic;
62. timed human ownership/cost audit from source-only diff;
63. warm private-body edit and public-summary edit;
64. generated adversarial programs reduced after interpreter/native mismatch.

Every workload must include its happy path, expected failure variants, boundary
conditions, and at least one plausible but unsafe/overly expensive design.

## Candidate implementations to compare

Run the same corpus through four checker/surface configurations and three
storage strategies. Do not build twelve full compilers; share the typed core
and vary the policy/annotation requirements.

### Checker/surface configurations

1. **Explicit affine:** explicit access/transfer at declarations and call sites;
   public origin relations.
2. **Boundary-explicit value access:** private inference, public modes,
   consuming/exclusive call-site markers, public origins. This is the leader.
3. **Inference-heavy:** modes and local/public origins inferred as far as
   possible, with overlays rather than source annotations.
4. **Projection-constrained:** no generally storable borrows in safe ordinary
   code; closure/yield APIs for temporary access; explicit native views only.

### Storage strategies

1. unique/arena baseline;
2. precise RC/reuse for eligible immutable cycle-free values;
3. one owner-confined managed region for cyclic graphs.

Ordinary ARC, a global tracing heap, and generational checked references are
controls, not leading production candidates. They answer whether the hybrid's
complexity actually buys something.

## Measurement protocol

### Hard vetoes

Reject a candidate if any of these survives reduction:

- a safe program can use invalid memory, double-release, or create a data race;
- interpreter and native optimized execution disagree on ownership/resource
  behavior outside declared nondeterminism;
- a public API's legality depends on reading an implementation body;
- adding an unrelated later use silently introduces a deep copy or changes
  resource identity;
- resource correctness depends on tracing finalization;
- a required native workload has no safe expression except pervasive unsafe;
- the ordinary model requires a global collector, scheduler, or runtime;
- the checker has an avoidable adversarial complexity cliff that prevents the
  promised edit loop;
- release mode removes safety semantics instead of proving checks unnecessary.

### Scored outcomes

| Dimension | Weight | Evidence |
|---|---:|---|
| soundness and semantic locality | 20 | formal core, negative probes, backend differential |
| human auditability | 15 | time, accuracy, gaze/navigation proxy, source-only explanation |
| AI generation and repair | 15 | first-pass correctness, repair rounds, tokens, tool calls, hidden tests |
| check/compile economics | 15 | p50/p95/p99 edit latency, work/query counts, peak memory, invalidation |
| runtime predictability | 15 | copies, retains, allocations, RSS, p99.9/max stalls, critical-path work |
| native/FFI expressiveness | 10 | safe coverage, unsafe lines/obligations, ABI/sanitizer results |
| application/library ergonomics | 5 | annotation density, helper/wrapper count, generic reuse |
| evolution and interoperability | 5 | fingerprint stability, migration size, mixed-version result |

Record distributions and effect sizes, not a single composite score. The
weights prioritize correctness and feedback while preventing a syntactically
pleasant design from hiding runtime or FFI failure.

### Diagnostic metrics

- distance from the primary error to ownership origin/transfer/conflict;
- number of concepts and compiler-internal names required to explain it;
- whether the diagnostic distinguishes an unsafe program from checker
  conservatism;
- whether each repair changes semantics/cost and says how;
- stability of diagnostic code/cause graph after formatting and unrelated edits;
- reduced-probe size and time to first useful failure.

Research on ownership usability emphasizes the difference between a truly
unsafe program and a safe program rejected by an incomplete analysis. Lang's
diagnostic protocol must make that distinction explicit rather than presenting
all rejections as moral failures by the author. ([The Usability of
Ownership](https://arxiv.org/abs/2011.06171))

## Compiler implementation sequence

### Stage 1 — executable dynamic model

- Implement owned/moved/released states and shared/exclusive loans in the
  reference interpreter.
- Give every allocation, place, origin, resource, and obligation a stable test
  identity.
- Execute deliberately invalid kernel programs under the interpreter to prove
  that each static rejection prevents a concrete bad state.

### Stage 2 — minimal static checker

- Definite initialization and whole-value moves.
- Local shared/exclusive borrows with last-use liveness.
- Public parameter modes and single-source returned origins.
- Whole-consuming pattern match; no partial residual.
- Escaping capture modes.
- Lexical resource release and explicit obligation completion.

### Stage 3 — native lowering

- Unique buffers, slices, arenas, drop order, OOM, and one C boundary.
- Validate optimizer alias/provenance facts against the written core.
- Differentially execute every blocking probe in interpreter and native modes.

### Stage 4 — hard boundaries

- Multi-source views, noncopyable generics, trait/port erasure.
- Structured-task lending experiment.
- Stable address, callbacks, uninitialized storage, dynamic libraries.
- Actor transfer/share derivation.

### Stage 5 — storage contest

- Precise RC/reuse on eligible immutable values.
- Indexed arena versus one managed cyclic region.
- Release-tail, contention, peak-memory, and graph-ergonomics comparison.

### Stage 6 — source and diagnostic selection

- Render the surviving semantics through the ownership surfaces.
- Run human/AI audit, repair, diff, and token studies.
- Only then freeze call-site markers and final keywords.

## Anti-patterns to reject early

- **Hidden clone on conflict:** copying because later code still uses a moved
  value makes performance and identity depend on an unrelated edit.
- **One magic reference type:** bounds, origin, mutability, retention,
  concurrency, and storage cannot safely remain undocumented conventions.
- **Lifetime inference across public bodies:** it breaks separate compilation,
  makes diagnostics nonlocal, and turns refactors into API changes.
- **Every object chooses its collector:** arbitrary mixtures fracture generic
  APIs and make latency/cycles impossible to explain.
- **RAII as arbitrary user code:** I/O, locks, allocation, async work, or
  failure in implicit destruction creates hidden effects and double-failure
  problems.
- **A large capability qualifier lattice in ordinary code:** Pony-like power is
  valuable at concurrency boundaries; spreading it everywhere defeats the calm
  source goal.
- **Reference-first domain APIs:** returning/storing references by habit creates
  identity and lifetime coupling where values or IDs would be clearer.
- **Unsafe as checker suppression:** unsafe grants specific primitive authority
  but does not disable types, effects, initialization, or architecture checks.
- **Foreign signature equals foreign proof:** C headers do not state retention,
  aliasing, thread, unwind, initialization, or shutdown truth.
- **No-GC equals bounded latency:** RC cascades, arena reset, allocator faults,
  and destructors can all spike.
- **Borrow checker as optimizer:** legality cannot depend on whole-program escape
  optimization; optimization consumes proven ownership facts.
- **Teaching only rules, not repairs:** an agent or human needs the value's
  permission timeline and costed legal transformations.

## Formal and unsafe assurance

The portable ownership core should be specified independently of source and
backend. Oxide is a useful starting point for a source-oriented affine borrow
calculus; it is not a drop-in language spec. The model must add resource
obligations, effects, cancellation, managed-owner boundaries, FFI, and the
project's defined panic/OOM semantics.

Unsafe libraries extend the safe language and therefore require semantic
obligations, not only fenced syntax. RustBelt's central lesson is that safe
interfaces backed by unsafe implementations need a proof story for why they
are sound extensions. Lang's pragmatic ladder is:

1. machine-readable invariants and trust boundary;
2. safe wrapper tests and negative compile probes;
3. interpreter/native differential execution;
4. fuzzing, sanitizers, model checking, and ABI tests;
5. optional machine-checked proof for high-value primitives;
6. transitive evidence and reviewer/owner identity in release artifacts.

([RustBelt](https://people.mpi-sws.org/~dreyer/papers/rustbelt/paper.pdf),
[Rust unsafe-code guidelines](https://rust-lang.github.io/unsafe-code-guidelines/glossary.html))

## What is settled, provisional, and open

### Settled project boundary

- usable native ownership/resources are required in the first implementation;
- no mandatory global tracing heap;
- no arbitrary effects in implicit destruction;
- explicit arenas and foreign/unsafe boundaries;
- managed cyclic graphs, if admitted, are owner-confined;
- correctness does not change between development and release modes.

### Strong leading recommendation

- value/access source model over universal explicit reference syntax;
- affine noncopyable values, with duplication and abandonment separated;
- local flow-sensitive loans and explicit public ownership summaries;
- nonescapable borrowed views with source origins;
- visible consuming/exclusive semantic cliffs at call sites;
- borrowing match by default and whole-value consuming match;
- no ordinary borrow across suspension initially;
- structured transfer/share rather than ordinary shared mutable memory.

### Must be measured before stabilization

- exact call-site transfer syntax and whether named punned forms remain calm;
- how much private parameter-mode inference is predictable;
- field/index sensitivity and the safe `split`/projection API;
- structured child-task borrowing and zero-copy async pressure;
- noncopyable collection/iterator/higher-order interface ergonomics;
- precise RC/reuse eligibility and release scheduling;
- the managed-region representation and whether it is needed for v1;
- dynamic exclusivity fallback scope;
- exact pointer provenance and unsafe alias contract;
- checker algorithm, worst-case behavior, and incremental fingerprints.

## Stop rule

Do not perform another general ownership survey before implementation. Reopen
the semantic recommendation only if the laboratory finds one of these:

1. two independent realistic workloads require the same missing ownership
   primitive;
2. a required native/FFI case cannot be expressed safely without pervasive
   unsafe or copies;
3. public summaries cannot support separate compilation;
4. compiler or runtime budgets fail because of the semantic model rather than
   an implementation choice;
5. interpreter/native/formal models disagree;
6. independent authors repeatedly misunderstand the same irreducible concept;
7. a serious alternative wins the declared whole-loop rubric.

Otherwise, new concerns become probes, official toolkit designs, storage
policies, or later proposals. The next useful knowledge comes from implementing
Stages 1–3 and running Suites 1–3 plus the first FFI cases.

The final cross-cutting seam and unknown-horizon check is recorded in the
[ownership experiment readiness audit](ownership-experiment-readiness-audit.md).
It adds combinatorial overlays and conservative preservation requirements; it
does not introduce another ownership family.

## Primary evidence and limits

- [Rust NLL](https://rust-lang.github.io/rfcs/2094-nll.html), [two-phase
  borrows](https://rust-lang.github.io/rfcs/2025-nested-method-calls.html), and
  [Polonius](https://rust-lang.github.io/polonius/rules.html) establish concrete
  flow-sensitive borrow-checking designs; they do not establish Lang's
  ergonomics or latency.
- [Oxide](https://arxiv.org/abs/1903.00982) gives a formal account of a
  Rust-like ownership core; Lang adds mechanisms and must prove its own model.
- Swift's [parameter ownership](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0377-parameter-ownership-modifiers.md),
  [noncopyable types](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0390-noncopyable-structs-and-enums.md),
  [nonescapable types](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0446-non-escapable.md),
  and [region isolation](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0414-region-based-isolation.md)
  show the practical cross-feature surface; they also show how much library and
  generic design ownership touches.
- [Hylo](https://hylo-lang.org/introduction/) demonstrates a reference-light
  mutable-value/projection design; it remains a developing language and does
  not settle every FFI or ecosystem case.
- [Move abilities](https://move-language.github.io/move/abilities.html) and the
  [Move resource paper](https://arxiv.org/abs/2004.05106) establish a useful
  separation between duplication, abandonment, and storage for their domain;
  general-purpose resource cleanup has additional requirements.
- [Austral](https://austral-lang.org/spec/spec.html), [Pony](https://www.ponylang.io/learn/reference-capabilities/),
  [Cyclone](https://www.cs.umd.edu/users/mwh/papers/hicks03safe.html), and
  [Verona](https://www.microsoft.com/en-us/research/project/project-verona/)
  establish credible linear, capability, and region families. Their existence
  is not evidence that Lang should expose all of their concepts.
- [Perceus](https://doi.org/10.1145/3453483.3454032) establishes precise RC and
  reuse for a cycle-free functional core; concurrency, cycles, FFI, and Lang's
  latency distribution remain experiments.
- Rust's [destructor rules](https://doc.rust-lang.org/reference/destructors.html),
  [`MaybeUninit`](https://doc.rust-lang.org/core/mem/union.MaybeUninit.html),
  [pointer provenance](https://doc.rust-lang.org/core/ptr/index.html), and
  [interior mutability](https://doc.rust-lang.org/reference/interior-mutability.html)
  show why cleanup, validity, aliasing, and provenance require independent
  semantic answers.
- [CHERI](https://github.com/CTSRD-CHERI/cheri-specification/blob/main/chap-intro.tex)
  demonstrates that bounds, permissions, integrity, and provenance can receive
  hardware support; ordinary targets still need a software semantic model.

Sources were rechecked on 2026-09-03. Comparative rankings and the hybrid
recommendation are project synthesis, not claims made by those sources.
