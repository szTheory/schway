# Phase 2: Owned Values and Abilities - Research

**Researched:** 2026-09-03
**Domain:** Affine ownership, independent type abilities, and source-to-C17 semantic preservation
**Confidence:** HIGH for the bounded implementation direction; MEDIUM for provisional surface spelling

## User Constraints

### Locked Phase Goal

> “The end-to-end compiler distinguishes cheap implicit copy from explicit transfer and derives independent value abilities coherently.” [VERIFIED: .planning/ROADMAP.md:45-55]

### Locked Requirements

> “OWN-01: Noncopyable values transfer exactly once, and use-after-move or move-during-loan programs are rejected with the transfer and conflict causes.” [VERIFIED: .planning/REQUIREMENTS.md:44-45]

> “OWN-02: Copy, drop, share, send, and escape abilities are derived independently, including through generic and aggregate types.” [VERIFIED: .planning/REQUIREMENTS.md:46-47]

### Locked Scope Boundaries

- Keep the implementation vertical: source, lossless syntax, checked core, interpreter, readable C17, native execution, protocol, and evidence must advance together. [VERIFIED: .planning/PROJECT.md:68-73]
- Keep Go 1.24 standard-library-only as Stage 0 and readable C17 through installed Clang as the reversible native path. [VERIFIED: AGENTS.md:32-35]
- Do not pull Phase 3 into this phase: CFG point/edge-specific last-use and public returned origins remain OWN-03/OWN-04. [VERIFIED: .planning/REQUIREMENTS.md:48-51]
- Preserve the explicit limitation that content-bound core validation does not prove a coordinated source-to-core translation claim. [VERIFIED: .planning/spikes/004-independent-certificate-checker/README.md:28-32]

## Summary

Phase 2 should add one **straight-line owned-value body** to the real compiler, not another workbench and not a general borrow checker. The minimum useful slice is: a function receives a compiler-defined noncopyable `Buffer`, `let delivered = take buffer` emits one explicit transfer, and the function returns `delivered`. A second fixture copies a cheap compiler-defined `Byte` with an unmarked binding. Two invalid source fixtures preserve the reduced controls: use after transfer and transfer while a borrow is still used later. [VERIFIED: .planning/spikes/001-ownership-kernel-workbench/README.md:75-96]

Abilities should be five independent booleans in typed core—`copy`, `drop`, `share`, `send`, and `escape`—with one structural derivation algorithm and one independently coded test oracle. `Box<T>` and `Pair<L, R>` are enough generic/aggregate pressure for this phase: an ability is granted exactly when every contained argument required by that constructor grants the same ability. The real compiler must own primitive roots; source authors must not assert positive primitive abilities yet. The earlier generic-ability spike exhaustively checked all 32 atomic masks and found that immutable, content-owned summaries are necessary because shallow copies let mutations corrupt the oracle. [VERIFIED: .planning/spikes/003-public-origins-generic-abilities/README.md:121-133] [VERIFIED: .planning/spikes/003-public-origins-generic-abilities/README.md:201-225]

The implementation should reuse Phase 1's deterministic command/evidence spine, but factor execution facts out of the interpreter so native output can be decoded into the same inert execution model and compared event-for-event. Phase 1 currently compares only native output values even though its interpreter exposes ordered events, so event-equivalent ownership is the most important cross-layer change. [VERIFIED: internal/compiler/session/session.go:171-207] [VERIFIED: internal/compiler/interp/interp.go:10-50]

**Primary recommendation:** implement a source-visible `take`, straight-line `let`, type applications, a memoized structural ability engine, linear place/loan checking, shared execution facts, strict native event decoding, and a compact independent core validator; defer CFG loans, public origins, resources, allocation, concurrency, and general generics.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|---|---|---|---|
| `let`, `take`, `borrow`, and `T<U>` parsing | Lossless syntax / AST | Formatter | Syntax remains reversible and canonical; it must not decide ownership legality. |
| Primitive and constructor ability rules | Type/ability subsystem | Core validator | One compiler-owned rule table derives facts; an independent consumer recomputes them at trust crossings. |
| Copy versus transfer selection | Semantic checker/lowering | Typed core | Legality is decided before execution and represented explicitly as `copy` or `move` operations. |
| Place initialization and active loans | Ownership checker | Diagnostic builder | A deterministic straight-line pass owns move state and produces stable causes/repairs. |
| Dynamic state oracle | Interpreter | Test generator | The interpreter executes already-checked operations and also detects deliberately corrupted core states. |
| Physical value representation and event emission | C17 lowering | Native runner | Generated C uses stack/inline values and emits bounded execution facts; the runner only compiles, runs, and decodes. |
| Human/JSON projection | Protocol | Session orchestration | Existing `lang.command/0` projection remains the user contract. [VERIFIED: internal/compiler/protocol/protocol.go:14-22] |
| Content binding | Evidence | Core validator | Evidence binds source/core/C/tool facts but never claims source-to-core proof. [VERIFIED: internal/compiler/evidence/evidence.go:97-132] |

## Phase Requirements

| ID | Description | Research Support |
|---|---|---|
| OWN-01 | Noncopyable values transfer exactly once; invalid post-move and borrowed moves fail causally. | Explicit `take`, linear operations, source spans on declaration/transfer/use, a straight-line loan-live pass, dynamic oracle, native event equivalence, and three negative controls. |
| OWN-02 | Five abilities derive independently through generic/aggregate shapes. | Five-bit ability set, per-ability witnesses, `Box<T>`/`Pair<L,R>` structural rules, exhaustive 32-mask and pairwise 1,024-combination checks, forged-fact mutation. |

## Project Constraints (from AGENTS.md)

- Optimize the whole generate → verify → run → observe → repair loop and report distributions rather than a flattering single number. [VERIFIED: AGENTS.md:23-25]
- Safe behavior and optimizer/FFI facts must come from checked facts. [VERIFIED: AGENTS.md:26-29]
- The first native slice has no mandatory tracing heap or service runtime. [VERIFIED: AGENTS.md:30-31]
- Use Go 1.24 standard library for Stage 0 and readable C17 through Clang. [VERIFIED: AGENTS.md:32-33]
- Prefer no dependency over an unearned dependency. [VERIFIED: AGENTS.md:34-35]
- Preserve formatter-owned, Git-friendly, lossless-enough source. [VERIFIED: AGENTS.md:36-37]
- Distinguish fixtures, negative controls, properties, mutations, differential execution, and sanitizers by question and cost lane. [VERIFIED: AGENTS.md:38-40]
- Initial host priorities are macOS and Linux; wire facts cannot encode Apple arm64 accidentally. [VERIFIED: AGENTS.md:41-42]
- Untrusted input, unsafe operations, FFI, secrets, and build authority stay explicit; there is no generic untaint/ambient-authority escape. [VERIFIED: AGENTS.md:43-44]
- Work is already inside the GSD phase workflow; implementation must continue through planned GSD execution. [VERIFIED: AGENTS.md:77-87]

## Standard Stack

### Core

| Technology | Version | Purpose | Why Standard |
|---|---:|---|---|
| Go | 1.24.0 | Stage 0 compiler, checker, interpreter, validator, protocol, tests | Installed and already the repository module target: `go 1.24`. [VERIFIED: go.mod:1-3] Environment probe returned `go version go1.24.0 darwin/arm64`. |
| Go standard library | Go 1.24 | `encoding/json`, `crypto/sha256`, `os/exec`, `testing`, `testing/quick`, fuzzing | Phase 1 is already dependency-free and its full tests pass in this session. [VERIFIED: .planning/REQUIREMENTS.md:12-13] |
| C | C17 | Reviewable development lowering | Existing runner invokes Clang with the exact flags `"-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic"`. [VERIFIED: internal/compiler/native/native.go:61-75] |
| Apple Clang | 21.0.0 on current host | `-O0`/`-O3` native differential | Environment probe returned `Apple clang version 21.0.0 (clang-2100.1.1.101)`; portability claims remain limited to this host until Linux runs. |

### Supporting

| Facility | Purpose | When to Use |
|---|---|---|
| `testing/quick` | Seeded bounded structural properties | Ordinary quick lane; Phase 1 already uses 1,000 deterministic generated cases. [VERIFIED: internal/compiler/syntax/syntax_test.go:171-219] |
| Native Go fuzzing | Coverage-guided malformed/core operation discovery | Time-bounded investigation/nightly lane; seeds still run during ordinary `go test`. [CITED: https://go.dev/doc/security/fuzz/] |
| SHA-256 | Content identity, not signer authority | Existing evidence and stable IDs only. [VERIFIED: internal/compiler/evidence/evidence.go:208-236] |

### Alternatives Considered

| Instead of | Could Use | Tradeoff | Decision |
|---|---|---|---|
| Explicit `take` on named noncopyable transfer | Implicit move | Less syntax, but an unrelated later use changes an apparently ordinary read into a hard semantic cliff and obscures cost in review. | Use explicit `take` in this slice. |
| Five independent abilities | One `Copy`/resource class | Smaller type surface, but conflates duplication, abandonment, sharing, transfer, and escape. Move's official ability rules show independent operation gates and structural propagation. [CITED: https://move-language.github.io/move/abilities.html] | Keep five independent facts. |
| Structural `Box`/`Pair` rules | General trait/constraint solver | More expressiveness, but introduces recursive requirements, coherence, solver termination, and monomorphization questions outside OWN-01/02. Swift SE-0427 records recursive requirement growth as a major complication. [CITED: https://github.com/swiftlang/swift-evolution/blob/main/proposals/0427-noncopyable-generics.md] | Use sealed constructors only. |
| Shared inert execution facts | Native results translated through interpreter APIs | Fewer files, but makes the native lane depend on the oracle implementation it is meant to compare. | Factor `internal/compiler/execution`. |
| Inline fixed buffer | Heap allocation and allocator ownership | More realistic allocation, but prematurely pulls OOM, partial initialization, cleanup, and FFI into Phase 2. | Use stack/inline bytes; Phase 4 owns resources. |

**Installation:** none. This phase must add no external packages.

## Prescriptive Semantic Design

### 1. Smallest vertical source tracer

Add only these grammar forms:

```text
type         := Identifier ("<" type ("," type)* ">")?
linear_body  := "{" binding* Identifier "}"
binding      := "let" Identifier "=" rhs
rhs          := Identifier | "take" Identifier | "borrow" Identifier
```

The parser can distinguish the legacy `match` body from a linear body by its first non-trivia token. `take`, `borrow`, and `let` must become real tokens so recovery and formatting own them. Type nesting is capped (recommended: 64) and a single type expression is capped by node count (recommended: 4,096); exceeding either emits a bounded syntax diagnostic rather than recursing until stack exhaustion. These numbers are proposed guardrails, not ratified language limits. [PROPOSED]

Use two canonical positive fixtures:

```text
module owned.transfer

export {
  fn relay
}

fn relay(buffer: Buffer) -> Buffer {
  let delivered = take buffer
  delivered
}
```

```text
module owned.copy

export {
  fn retain
}

fn retain(code: Byte) -> Byte {
  let duplicate = code
  code
}
```

The committed fixtures must already be canonical and the formatter must reproduce them exactly. `Byte` is compiler-defined with all five abilities. `Buffer` is compiler-defined with `drop`, `send`, and `escape`, but without `copy` or `share`. The runtime representative is a fixed inline byte array used only by the current bounded run harness; it is not a promise that all future buffers are inline or fixed-size. [PROPOSED]

Use three source controls:

```text
// ownership.use_after_move
fn invalid(buffer: Buffer) -> Buffer {
  let delivered = take buffer
  buffer
}
```

```text
// ownership.move_while_borrowed: `view` is read after the move point
fn invalid(buffer: Buffer) -> Buffer {
  let view = borrow buffer
  let delivered = take buffer
  let observed = view
  delivered
}
```

```text
// ownership.transfer_requires_take
fn invalid(buffer: Buffer) -> Buffer {
  let delivered = buffer
  delivered
}
```

The move-during-loan case needs a post-move use of `view`. The ownership spike already found that an unused loan ends immediately and therefore does **not** witness the conflict; its fault injection required the four operations `declare`, `borrow_shared`, `move`, `read_loan`. [VERIFIED: .planning/spikes/001-ownership-kernel-workbench/README.md:74-92]

### 2. AST and typed-core shape

Refactor both AST and core function bodies into an explicit closed union: `match` or `linear`. Go cannot enforce the union statically, so constructors plus validation must require exactly one populated variant. Do not retain a zero-value combination that both interpreter and C generator interpret independently. [RECOMMENDATION]

Add immutable serializable core facts:

```go
type Ability string

type AbilitySet struct {
    Copy   bool `json:"copy"`
    Drop   bool `json:"drop"`
    Share  bool `json:"share"`
    Send   bool `json:"send"`
    Escape bool `json:"escape"`
}

type TypeRef struct {
    Constructor string    `json:"constructor"`
    Arguments   []TypeRef `json:"arguments"`
}

type TypeFact struct {
    ID                string                   `json:"id"`
    Shape             TypeRef                  `json:"shape"`
    Abilities         AbilitySet               `json:"abilities"`
    NegativeWitnesses map[Ability][]string     `json:"negative_witnesses"`
}

type LinearOperation struct {
    ID       string `json:"id"`
    PointID  string `json:"point_id"`
    Kind     string `json:"kind"`
    SourceID string `json:"source_id,omitempty"`
    TargetID string `json:"target_id,omitempty"`
    LoanID   string `json:"loan_id,omitempty"`
    TypeID   string `json:"type_id"`
    Span     diagnostic.Span `json:"span"`
}
```

This is a proposed skeleton, not a frozen wire definition. Prefer sorted slices over maps in the final wire representation; the map above is illustrative only. Existing semantic IDs are readable `s1:` strings and current core already gives functions entry/return/match points and arm edges explicit IDs. [VERIFIED: internal/compiler/check/check.go:30-33] [VERIFIED: internal/compiler/check/check.go:81-87]

Use the operation kinds `copy`, `move`, `borrow_shared`, `read_loan`, and `return`. Preserve distinct declaration, source, transfer, loan, and use spans. Stable operation IDs derive from the function ID plus source-order ordinal; they must not derive from byte offset alone because canonical formatting and unrelated prefix edits should not change semantic identity. [PROPOSED]

### 3. Independent ability derivation

Implement `internal/compiler/ability` as a pure memoized function over `TypeRef` and a sealed primitive/constructor rule table:

```text
derive(Byte)   = {copy, drop, share, send, escape}
derive(Buffer) = {drop, send, escape}

derive(Box<T>).A      = derive(T).A
derive(Pair<L,R>).A   = derive(L).A AND derive(R).A
```

Evaluate each ability separately in the stable order `copy`, `drop`, `share`, `send`, `escape`. Return the smallest deterministic negative witness path such as `Pair.left -> Box.value -> Buffer`, not only `false`. Never infer one ability from another: `copy` does not imply `share`, `send` does not imply `escape`, and `drop` does not imply `copy`. [RECOMMENDATION]

Move officially gates copy/drop independently and requires corresponding abilities from all contained fields; generic instances conditionally retain each ability from type arguments. [CITED: https://move-language.github.io/move/abilities.html] Swift's staged noncopyable rollout shows that a pervasive implicit copy assumption blocks ordinary wrappers such as `Optional` and requires generic-system work later. [CITED: https://github.com/swiftlang/swift-evolution/blob/main/proposals/0427-noncopyable-generics.md]

Do not let source declare positive primitive abilities in Phase 2. Positive root facts belong to the compiler's audited table. Generic constructors only propagate or remove them. A later unsafe ability witness needs the transitive evidence policy, not a new annotation in this phase. The public-origin spike explicitly records that its ability rules were trusted inputs and says the real compiler must derive positives from fields or require an audited unsafe witness. [VERIFIED: .planning/spikes/003-public-origins-generic-abilities/README.md:259-271]

### 4. Straight-line ownership checker

Use two small passes over the linear body:

1. **Borrow-binding last-use scan:** record each borrow binding's final operation index. This is linear straight-line support only; no CFG, joins, loops, or edge-specific facts.
2. **Forward place-state pass:** maintain `initialized`/`moved` for owned places and active shared loans keyed by owner. End a loan immediately after its last indexed use, then process the next operation.

Rules:

- Bare RHS of a `copy` type emits `copy` and leaves the source initialized.
- Bare RHS of a noncopyable type emits `ownership.transfer_requires_take`; never silently reinterpret it as a move or clone.
- `take source` requires initialized source and no active loan, emits `move`, initializes target, and marks source moved.
- Reading a moved source emits `ownership.use_after_move` at the read, with causes pointing to the move and original declaration.
- Moving while a future-used loan remains active emits `ownership.move_while_borrowed` at `take`, with causes pointing to borrow creation and future use.
- Returning an initialized local consumes it into the terminal outcome without a second source-level `take`; the explicit internal return operation remains in core.
- At function exit, every still-initialized binding lacking `drop` is an error. `Buffer` has `drop`, so this phase does not yet model resource obligations.

Rust's reference provides a useful control model: a non-`Copy` move deinitializes a local until reinitialization, and only unborrowed variables are movable. [CITED: https://doc.rust-lang.org/reference/expressions.html#moved-and-copied-types] Its official diagnostics identify use-after-move and move-while-borrowed as distinct failures. [CITED: https://doc.rust-lang.org/error_codes/E0382.html] [CITED: https://doc.rust-lang.org/error_codes/E0505.html]

The pass must count inspected operations and type nodes. Verify linear work over 10, 100, 1,000, and 10,000 operations by counts, not a brittle wall-time threshold. Memoize `TypeRef` derivation by canonical structural key and reject recursive/user-defined generic shapes for now. [RECOMMENDATION]

### 5. Diagnostics and repairs

Extend ownership diagnostics with a sorted `repairs` array under `lang.diagnostic/1`; unchanged Phase 1 diagnostics remain `lang.diagnostic/0`. Current diagnostics have exactly the fields `"schema"`, `"id"`, `"code"`, `"severity"`, `"primary_span"`, `"message"`, and optional `"causes"`; IDs hash schema, code, span, and causes. [VERIFIED: internal/compiler/diagnostic/diagnostic.go:10-42]

Each ownership diagnostic should carry:

| Code | Primary | Required causes | Small repair choices |
|---|---|---|---|
| `ownership.use_after_move` | post-move use | `declared_here`, `moved_here`, stable place/type IDs | `use_transfer_target`; `move_use_before_transfer`; `explicit_clone` only if a declared clone exists |
| `ownership.move_while_borrowed` | `take` | `borrow_created_here`, `borrow_used_later`, owner/loan IDs | `move_after_last_borrow_use`; `remove_unused_borrow`; `change_consumer_to_borrow` only when its signature permits |
| `ownership.transfer_requires_take` | bare noncopyable RHS | declaration and missing `copy` witness path | `insert_take`; `borrow_instead` when result need not own; `explicit_clone` when available |
| `ability.missing` | operation requiring ability | requested ability and smallest negative field/type path | change operation; change wrapper; use explicit representation that lawfully grants ability |

Diagnostic IDs must include semantic causes but not prose or suggested display order. Human output leads with the binding and relationship; JSON preserves IDs, spans, type facts, and semantic repairs. Do not suggest “make the type Copy” for a unique buffer because that is often semantically invalid. [RECOMMENDATION]

### 6. Shared execution model and interpreter

Create `internal/compiler/execution` containing the inert `Outcome`, `Event`, and `Execution` wire types now owned by `interp`. The exact current execution schema is `"lang.execution/0"`; its event fields are `schema`, `id`, `kind`, `function_id`, `input`, and `output`. [VERIFIED: internal/compiler/interp/interp.go:10-31]

The Phase 2 event union should add explicit fields rather than overload input/output strings:

```text
value.copied       {source_place, target_place, type_id}
value.transferred  {source_place, target_place, type_id}
value.borrowed     {owner_place, loan_id, target_place, type_id}
value.read         {place_or_loan, type_id}
function.returned  {function_id, source_place, value_digest}
```

The interpreter replays typed operations with its own runtime place/loan state. Valid programs emit the event list; corrupted core that reads a moved place fails deterministically. It must not call the static checker's transition function. Comparison should use canonical semantic bytes containing outcome, events, and live resources, excluding timings, addresses, and native process details. [RECOMMENDATION]

### 7. Readable C17 and native equivalence

Use an inline C value for the tracer:

```c
typedef struct LANG_BUFFER {
  unsigned char bytes[4];
  size_t length;
} LANG_BUFFER;
```

Lower a checked move as an ordinary C value transfer plus compiler-local liveness bookkeeping; do not emit `restrict`, `noalias`, or pointer provenance claims. Phase 2 proves source authority and event order, not zero physical copy, heap ownership, or ABI stability. [RECOMMENDATION]

Generated C should print one bounded canonical JSON execution document. `native.Runner` must decode it strictly into `execution.Execution`, reject unknown/trailing JSON, enforce an output cap (recommended 64 KiB per invocation), and keep the existing timeout. The current runner already uses literal `exec.CommandContext`, a fresh temporary directory, five-second defaults, bounded compiler error output, and no shell. [VERIFIED: internal/compiler/native/native.go:40-95]

Compare interpreter, `-O0`, and `-O3` on:

- terminal outcome;
- ordered `value.copied`/`value.transferred`/`function.returned` events;
- exact stable event IDs and place/type IDs;
- empty live-resource set.

Do not compare C addresses, timing, or the physical number of machine copies. Native output must be a consumer of the core operation list, not a second inference of source intent. [RECOMMENDATION]

### 8. Core validation and evidence evolution

Add `internal/compiler/corevalidate`, independent of AST/checker packages. It validates:

- exactly one body variant;
- unique stable IDs and declared point order;
- every referenced place/type/loan exists;
- ability facts equal recomputation from type shapes and primitive rules;
- `copy` appears only for a copyable type;
- linear move/loan transitions are valid;
- final claims match replay.

Require validation before interpreter and C lowering and again when accepting content-bound evidence. Mutate one represented boundary at a time: forged `copy`, omitted move, duplicate operation ID, bad source place, wrong type ability, and event reorder. The earlier independent checker caught 12 mutation classes but intentionally missed a coordinated false typed-core statement. [VERIFIED: .planning/spikes/004-independent-certificate-checker/README.md:117-127]

Keep that limitation explicit in comments, diagnostics, research, and phase verification: a core validator establishes internal consistency and artifact integrity, not that the frontend translated source correctly. [VERIFIED: internal/compiler/evidence/evidence.go:97-101]

Versioning recommendation:

- keep `lang.command/0` because the envelope fields and exit taxonomy remain compatible;
- emit ownership diagnostics carrying repairs as `lang.diagnostic/1`, while unchanged Phase 1 diagnostics remain `lang.diagnostic/0`;
- emit `lang.core/1` and `lang.execution/1` for linear owned bodies;
- retain `lang.core/0` and `lang.execution/0` for the Phase 1 pure fixture during the milestone;
- let evidence record the concrete source/core/execution schema selected for each artifact; do not rewrite the Phase 1 golden merely because the compiler learned S2.

This per-artifact versioning avoids pretending an additive Go field is an unchanged wire contract while preserving Phase 1 reproducibility. [RECOMMENDATION]

## Architecture Patterns

### System Architecture Diagram

```text
source bytes
    |
    v
lossless lexer/parser -----> canonical formatter
    |
    v
AST (match | linear body)
    |
    v
type + ability derivation -----> stable negative witnesses
    |
    v
ownership lowering/checker ----> diagnostics + repairs
    |
    v
immutable typed core ----------> independent core validator
    |                                  |
    | pass                             | fail closed
    +----------------+-----------------+
                     |
          +----------+----------+
          |                     |
          v                     v
 deterministic interpreter   readable C17 emitter
          |                     |
          |                 Clang O0 / O3
          |                     |
          +------ semantic execution facts ------+
                              |
                         exact comparison
                              |
                    protocol + evidence manifest
```

### Recommended Project Structure

```text
internal/compiler/
├── ability/          # pure five-ability derivation and witness paths
├── ast/              # reversible source meaning; match/linear body union
├── check/            # typing, lowering, straight-line ownership legality
├── core/             # immutable serialized type/place/operation facts
├── corevalidate/     # AST/body-blind recomputation and transition checks
├── execution/        # inert outcome/event/live-resource wire model
├── interp/           # deterministic dynamic oracle over checked core
├── cgen/             # readable C17 lowering from explicit operations
├── native/           # shell-free compile/run plus strict output decoder
├── diagnostic/       # versioned causes and semantic repair actions
├── evidence/         # content/tool/policy binding
└── session/          # end-to-end orchestration and exact comparison
testdata/phase2/
├── owned_transfer.lang
├── implicit_copy.lang
├── use_after_move.lang
├── move_while_borrowed.lang
└── implicit_noncopy.lang
```

### Pattern 1: Derive, materialize, independently validate

The frontend derives abilities and operation kinds once, stores them in core, and consumers use those explicit facts. `corevalidate` independently recomputes the small laws at evidence/trust boundaries. This avoids making every backend a type checker while still catching corrupted facts. The independent certificate spike selected compact recomputation over full replay for normal trust crossings. [VERIFIED: .planning/spikes/004-independent-certificate-checker/README.md:59-83]

### Pattern 2: Semantic events are cross-engine facts

Interpreter and native executions produce the same inert fact type; neither translates the other's output. Existing Phase 1 tests already require deterministic interpreter bytes and O0/O3 native paths. [VERIFIED: internal/compiler/session/session_test.go:39-51] [VERIFIED: internal/compiler/session/session_test.go:85-102]

### Anti-Patterns to Avoid

- **Implicit move for noncopyable RHS:** defeats visible transfer and makes a later use alter meaning.
- **Size-based `copy`:** size is a cost lint, not semantic permission.
- **Ability implication lattice:** each fact must be tested independently; never encode `copy => share` or `send => escape`.
- **User-asserted positive primitive ability:** an unaudited declaration can make safe wrappers unsound.
- **Shared transition helper for checker, interpreter, and validator:** agreement becomes tautological.
- **General CFG in Phase 2:** straight-line future-use is sufficient for the move-during-loan control; Phase 3 owns joins/edges.
- **Heap-backed “realistic” buffer:** pulls allocator, OOM, cleanup, and provenance into the wrong phase.
- **Native stdout prose scraping:** use strict bounded execution facts.
- **Treating validation manifest as proof:** the coordinated frontend lie remains an expected escape.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---|---|---|---|
| JSON encoding/strict decoding | Custom serializer/parser | Go `encoding/json` with `DisallowUnknownFields` and trailing-value check | Existing evidence already uses this strict pattern. [VERIFIED: internal/compiler/evidence/evidence.go:135-155] |
| Content digests | Custom checksum | Go `crypto/sha256` | Existing evidence IDs and digests already standardize on SHA-256. [VERIFIED: internal/compiler/evidence/evidence.go:208-236] |
| Process timeout/quoting | Shell command construction | `exec.CommandContext` with literal argv | Existing native runner already avoids shell interpretation. [VERIFIED: internal/compiler/native/native.go:61-95] |
| Property framework dependency | New generator/shrinker library | `testing/quick`, explicit exhaustive generator, native fuzzing | Current scale is small and inspectable; Go fuzzing minimizes failures and retains them as seeds. [CITED: https://go.dev/doc/security/fuzz/] |
| General type solver | Traits, SAT/SMT, recursive constraint engine | Five independent structural folds over sealed type shapes | OWN-02 needs derivation, not open-world polymorphism. |

**Key insight:** the ownership/ability transition relation is the thing this project must implement itself; serialization, hashing, process control, and fuzz orchestration are not.

## Common Pitfalls

### Pitfall 1: An unused borrow does not witness move-during-loan

**What goes wrong:** the test borrows a value and immediately moves it, but correct last-use reasoning ends the unused loan before the move.

**How to avoid:** include a borrow-binding use after the attempted move. The spike's minimal fault case needed four operations for this reason. [VERIFIED: .planning/spikes/001-ownership-kernel-workbench/README.md:74-92]

### Pitfall 2: Result-only differential tests miss support defects

**What goes wrong:** checker and oracle both reject early, hiding disagreement in inferred loan endpoints or intermediate states.

**How to avoid:** compare normalized operations, loan last-use indices, ability facts, and event lists—not only accept/reject. The spike found an endpoint defect that result-only comparison had hidden. [VERIFIED: .planning/spikes/001-ownership-kernel-workbench/README.md:164-172]

### Pitfall 3: Shallow “immutable” summaries share backing storage

**What goes wrong:** a mutation control changes both interface and oracle inputs and therefore escapes.

**How to avoid:** deep-copy at trust boundaries and hash content-owned canonical bytes. This occurred in Spike 003. [VERIFIED: .planning/spikes/003-public-origins-generic-abilities/README.md:217-225]

### Pitfall 4: Ability derivation becomes a hidden solver

**What goes wrong:** recursive aliases, user rules, implications, and associated types make local checking unpredictable.

**How to avoid:** sealed acyclic `Byte`, `Buffer`, `Box`, and `Pair` shapes; memoized structural fold; explicit node/depth caps; no recursive generics in Phase 2.

### Pitfall 5: Native C physically copies a moved struct

**What goes wrong:** reviewers interpret a C assignment as violation of source move semantics.

**How to avoid:** document that move is an authority transition, not a stable-address or zero-machine-copy promise; compare semantic access/events and preserve optimization freedom. [VERIFIED: wiki/ownership-and-lifetime-decisive-study.md:654-663]

### Pitfall 6: Schema drift invalidates Phase 1 evidence accidentally

**What goes wrong:** globally changing one schema constant rewrites old golden artifacts even for unchanged S1 source.

**How to avoid:** select source/core/execution schema by admitted feature set and keep protocol envelope compatibility separate.

### Pitfall 7: Fuzzing enters the default unbounded lane

**What goes wrong:** CI and agent iteration become nondeterministic and expensive.

**How to avoid:** run seeds in `go test`; bound active fuzzing by time/iterations in an explicit investigation lane. Go's official guidance says fuzz targets must be fast/deterministic and active fuzzing otherwise runs until failure or cancellation. [CITED: https://go.dev/doc/security/fuzz/]

## Code Examples

These are implementation patterns for the planner, not frozen public APIs.

### Per-ability structural fold

```go
// Proposed Lang Stage 0 pattern. The independent test oracle must not call this.
func derive(shape core.TypeRef, ability core.Ability) Result {
    switch shape.Constructor {
    case "Byte", "Buffer":
        return primitiveRule(shape.Constructor, ability)
    case "Box", "Pair":
        for index, argument := range shape.Arguments {
            child := derive(argument, ability)
            if !child.Granted {
                return child.Prepend(fieldName(shape.Constructor, index))
            }
        }
        return Granted()
    default:
        return Rejected("unknown_type_constructor")
    }
}
```

The important property is five independent calls to `derive`, not one aggregate “resource kind.” The Move reference independently supports operation-gated, field-conditional ability checks. [CITED: https://move-language.github.io/move/abilities.html]

### Forward affine transition

```go
// Proposed checker pattern after straight-line loan last uses are known.
switch operation.Kind {
case core.OpCopy:
    requireInitialized(operation.SourceID)
    requireAbility(operation.TypeID, core.Copy)
    initialize(operation.TargetID)
case core.OpMove:
    requireInitialized(operation.SourceID)
    requireNoActiveLoan(operation.SourceID)
    move(operation.SourceID, operation.TargetID)
case core.OpReadLoan:
    requireActiveLoan(operation.LoanID)
}
endLoansWhoseLastUseIs(operation.PointID)
```

The checker, dynamic interpreter, and core validator must each implement their own small transition logic. The local workbench showed why shared normalization can hide support defects. [VERIFIED: .planning/spikes/001-ownership-kernel-workbench/README.md:164-172]

### Strict native execution decoding

```go
// Follow the existing evidence decoder pattern; additionally wrap the reader
// in an explicit output limit before Decode.
decoder := json.NewDecoder(io.LimitReader(stdout, maxExecutionBytes+1))
decoder.DisallowUnknownFields()
if err := decoder.Decode(&execution); err != nil {
    return executionError("native.invalid_execution")
}
if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
    return executionError("native.trailing_execution")
}
```

The existing evidence decoder already uses `DisallowUnknownFields` and rejects a second JSON value. [VERIFIED: internal/compiler/evidence/evidence.go:144-155]

## State of the Art

| Precedent | Applicable lesson | Lang Phase 2 boundary |
|---|---|---|
| Rust moved places | Non-`Copy` moves deinitialize a place; borrowed variables are not movable. [CITED: https://doc.rust-lang.org/reference/expressions.html#moved-and-copied-types] | Adopt the law, but require visible `take` for a named noncopyable transfer. |
| Move abilities | Operations are gated independently and generic/aggregate abilities are conditional on contents. [CITED: https://move-language.github.io/move/abilities.html] | Use five project-specific abilities; do not import Move global-storage semantics. |
| Swift noncopyable generics | Copyability assumptions spread through wrappers, protocols, existentials, and ABI. [CITED: https://github.com/swiftlang/swift-evolution/blob/main/proposals/0427-noncopyable-generics.md] | Preserve generic ability facts now; defer protocols, existentials, and ABI. |
| Existing Lang Spike 001 | Independently implemented state machines, exhaustive sequences, and injected faults found real harness defects. [VERIFIED: .planning/spikes/001-ownership-kernel-workbench/README.md:193-218] | Port only the reduced linear laws into the real spine. |
| Existing Lang Spike 004 | Compact validation catches typed-core corruption but not coordinated source-to-core lies. [VERIFIED: .planning/spikes/004-independent-certificate-checker/README.md:170-195] | Add compact core validation and retain the limitation verbatim. |

## Explicit Deferrals

| Deferred item | Preserved seam | Owning phase/trigger |
|---|---|---|
| CFG branches, loops, edge-specific last use | operation point IDs, loan IDs, straight-line last-use facts | Phase 3 / OWN-03 |
| Public borrowed results and field/alternative origins | `TypeRef`, loan access mode, function result metadata slot | Phase 3 / OWN-04 |
| Exclusive borrows and mutation | operation-kind namespace, place paths | Phase 3 unless needed by a reduced Phase 2 counterexample |
| Resource obligations and fallible release | `drop` remains distinct from resource finish; live-resources field retained | Phase 4 / RES-01 |
| Heap allocation, arenas, OOM | Buffer abstract type and backend representation boundary | Phase 4/native resource slice |
| FFI/layout/provenance | no unjustified C attributes; target facts remain in evidence | Phase 4–5 |
| Full generics, recursive types, traits, variance, subtyping | structural `TypeRef`, ability witnesses | Later milestone or two unrelated blocking workloads |
| Share/send operations, actors, threads | ability facts exist without operations | Later concurrency milestone |
| User-declared/unsafe ability witnesses | core validator rule seam and evidence owner | Later unsafe policy |
| Final `take`/generic syntax freeze | lossless CST and canonical formatter make spelling reversible | Later surface/diagnostic comparison |

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|---|---|---|
| A1 | [ASSUMED] A fixed inline four-byte `Buffer` is enough to expose authority transfer without allocation. | C17 lowering | If real representation pressure changes event/alias behavior, Phase 2 may need a small owned heap block and would pull cleanup forward. |
| A2 | [ASSUMED] `take`, `borrow`, `let`, and angle-bracket type applications will remain readable enough through this phase. | Source tracer | Syntax may change later; lossless CST/core operation separation keeps it reversible. |
| A3 | [ASSUMED] A 64-level/4,096-node type-expression cap is ample for the bounded corpus. | Parser safety | Too small rejects generated stress cases; too large risks stack/work spikes. Measure and adjust without freezing as a language limit. |
| A4 | [ASSUMED] Per-feature diagnostic/core/execution schema selection is less disruptive than a global milestone-wide schema bump. | Evidence evolution | If implementation complexity is disproportionate, require an explicit migration decision and regenerate affected Phase 1 golden evidence. |

## Open Questions (RESOLVED)

These decisions are selected for planning; each retains an executable rule for reopening it if evidence changes.

1. **Should the returned local require a second `take`?**
   - Selected: no; final expression is an explicit ownership sink in typed core.
   - Reopen if source-only audits confuse return transfer or if adding a later statement creates ambiguous behavior.
2. **Should native output be JSON or a simpler line protocol?**
   - Selected: strict single-document JSON because the repository already standardizes canonical JSON and strict decoding.
   - Reopen only if measured native output/escaping complexity dominates the small C emitter.
3. **Should the diagnostic schema bump globally or per feature?**
   - Selected: version diagnostics per feature; unchanged Phase 1 diagnostics remain `lang.diagnostic/0`, while ownership diagnostics carrying `repairs` use `lang.diagnostic/1`.
   - Reopen only if implementation complexity proves disproportionate, and then require an explicit migration decision with regenerated golden evidence.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|---|---|---:|---|---|
| Go | Stage 0 build/test | yes | 1.24.0 darwin/arm64 | none required |
| Clang | C17 O0/O3 lane | yes | Apple Clang 21.0.0 | Operational `native.tool_missing`; never convert to source invalidity. Existing behavior is tested. [VERIFIED: internal/compiler/session/session_test.go:53-63] |
| Network | ordinary build/test | not required | — | all default work remains offline |
| External package manager | phase implementation | not required | — | Go standard library only |

**Missing dependencies with no fallback:** none on the current host.

**Missing dependencies with fallback:** Linux execution is not present in this session; retain portability as unverified until a Linux runner executes the same evidence.

## Validation Architecture

### Test Framework

| Property | Value |
|---|---|
| Framework | Go 1.24 `testing`, `testing/quick`, native fuzz seeds, black-box `os/exec` |
| Config file | `go.mod` |
| Quick run command | `env GOCACHE=/tmp/ai-lang-phase2-cache go test ./...` |
| Full suite command | `sh scripts/verify-phase2.sh` |

Current infrastructure already contains deterministic/property, race, vet, CLI, native O0/O3, mutation, and strict evidence patterns. Phase 1's declared quick/full commands are `go test ./...` and `go test -race ./... && go vet ./... && go run ./cmd/lang verify testdata/phase1`. [VERIFIED: .planning/phases/01-canonical-pure-spine/01-VALIDATION.md:14-29]

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|---|---|---|---|---|
| OWN-01 | Valid `Buffer` transfer emits one move and identical interpreter/O0/O3 events | end-to-end differential | `go test ./internal/compiler/session -run TestOwnedTransferInterpreterNative -count=1` | ❌ Wave 0 |
| OWN-01 | Use after move reports declaration + move + use causes and repairs | compile-reject contract | `go test ./internal/compiler/session -run TestUseAfterMoveDiagnostic -count=1` | ❌ Wave 0 |
| OWN-01 | Future-used shared loan blocks move | compile-reject contract | `go test ./internal/compiler/session -run TestMoveWhileBorrowedDiagnostic -count=1` | ❌ Wave 0 |
| OWN-01 | Bare noncopyable RHS never becomes implicit move/clone | negative control | `go test ./internal/compiler/session -run TestTransferRequiresTake -count=1` | ❌ Wave 0 |
| OWN-01 | Linear checker and dynamic oracle agree on exhaustive short programs | model/differential | `go test ./internal/compiler/check -run TestOwnershipSequenceExhaustive -count=1` | ❌ Wave 0 |
| OWN-01 | Checker work is linear by counted operations | scale property | `go test ./internal/compiler/check -run TestOwnershipWorkSeries -count=1` | ❌ Wave 0 |
| OWN-02 | All 32 leaf masks remain independent through `Box` | exhaustive unit | `go test ./internal/compiler/ability -run TestBoxAllAbilityMasks -count=1` | ❌ Wave 0 |
| OWN-02 | All 1,024 mask pairs derive independently through `Pair` | exhaustive unit | `go test ./internal/compiler/ability -run TestPairAllAbilityMasks -count=1` | ❌ Wave 0 |
| OWN-02 | Nested missing ability reports the smallest stable path | contract/property | `go test ./internal/compiler/ability -run TestNegativeWitnessPath -count=1` | ❌ Wave 0 |
| OWN-02 | Forged `copy`/missing op/duplicate ID cannot reach execution | mutation | `go test ./internal/compiler/corevalidate -run TestOwnershipMutationMatrix -count=1` | ❌ Wave 0 |
| OWN-01/02 | Ownership syntax preserves bytes, format fixed point, and bounded recovery | syntax property/fuzz seeds | `go test ./internal/compiler/syntax -run 'TestOwnershipRoundTrip|TestOwnershipRecovery|FuzzParseFormat' -count=1` | ❌ Wave 0 |
| OWN-01/02 | Phase 1 remains green and Phase 2 controls are observed | phase gate | `sh scripts/verify-phase2.sh` | ❌ Wave 0 |

### Required Negative Controls

The phase verifier must fail closed unless all of these exact control IDs are observed: [PROPOSED]

```text
control:ownership.use_after_move
control:ownership.move_while_borrowed
control:ownership.transfer_requires_take
control:ability.forged_copy
control:core.duplicate_operation_id
control:interpreter-o0-o3-owned
control:evidence.core_mismatch
```

Add one expected escape named `escape:coordinated-source-core-lie`. It is documentation/evidence of a trust boundary, not a passing mutation detector. The source/body-blind checker cannot detect a false source claim when the typed core and certificate consistently repeat it. [VERIFIED: .planning/spikes/004-independent-certificate-checker/README.md:154-161]

### Property and Fuzz Lanes

- Ordinary `go test`: exhaustive short operation alphabet, all ability masks, deterministic generated type trees, and fuzz seeds.
- Investigation: `go test ./internal/compiler/check -fuzz=FuzzOwnershipLinear -fuzztime=30s`; never part of the default phase script.
- Persist every minimized failure under `testdata/fuzz/FuzzOwnershipLinear`; Go runs seed corpora during ordinary tests. [CITED: https://go.dev/doc/security/fuzz/]
- Fuzz targets must allocate all state per invocation, use no globals, perform no native compilation, and cap type depth/operations before evaluation. Go fuzz workers run targets in parallel/nondeterministic order. [CITED: https://go.dev/doc/security/fuzz/]

### Sampling Rate

- **Per task commit:** targeted test plus `env GOCACHE=/tmp/ai-lang-phase2-cache go test ./...`
- **Per wave merge:** `go test -race ./...` and `go vet ./...`
- **Phase gate:** `scripts/verify-phase2.sh` runs tests once, race once, vet once, Phase 1 corpus verification, then Phase 2 corpus verification.
- **Performance observation:** 20 warm process samples for format/check/interpreter/native/full verify, reported as p50/p95/min-max with work counts. Do not ratify a release budget before Phase 6.

### Wave 0 Gaps

- [ ] Add the five `testdata/phase2/*.lang` fixtures listed in Recommended Project Structure.
- [ ] Add `internal/compiler/ability/ability_test.go` before implementing derivation.
- [ ] Add `internal/compiler/corevalidate/corevalidate_test.go` with mutation skeletons.
- [ ] Add ownership syntax round-trip/recovery seeds.
- [ ] Add `scripts/verify-phase2.sh` without duplicating `go test` inside nested scripts.
- [ ] Add a strict native execution decoder test with unknown/trailing/oversized output controls.

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---|---|---|
| V2 Authentication | no | Compiler is a local CLI; no identity boundary in Phase 2. |
| V3 Session Management | no | No session/token state. |
| V4 Access Control | no | No remote or multi-user authority boundary. |
| V5 Input Validation | yes | Bounded lexer/parser recovery, type depth/node caps, strict core/native JSON decoding, identifier validation. |
| V6 Cryptography | limited | SHA-256 binds content only; do not claim signing, freshness, or authorization. [VERIFIED: .planning/spikes/004-independent-certificate-checker/README.md:197-208] |

### Known Threat Patterns for the Compiler Spine

| Pattern | STRIDE | Standard Mitigation |
|---|---|---|
| Deep/numerous generic type forms exhaust recursion or work | Denial of service | depth/node caps, memoized acyclic derivation, counted work series |
| Source identifier/text escapes into generated C/JSON | Tampering | generate only from checked identifiers, central C/JSON escaping, strict output cap |
| Forged positive ability authorizes copy/share | Elevation of privilege / Tampering | compiler-owned root table, independent core recomputation, mutation control |
| Corrupted core uses moved place | Tampering | fail-closed core validator plus independent interpreter state checks |
| Native child hangs or floods output | Denial of service | existing process timeout plus proposed bounded output decoder |
| Content digest treated as signer authority | Spoofing | explicitly label evidence as content binding only |
| Compiler shell injection | Tampering | keep literal `exec.CommandContext`; never interpolate a shell command. [VERIFIED: internal/compiler/native/native.go:61-95] |

## Sources

### Primary Local Evidence (HIGH confidence)

- `.planning/REQUIREMENTS.md:35-55` — exact Phase 2 and adjacent requirement boundaries.
- `.planning/ROADMAP.md:45-70` — goal, success criteria, and Phase 3 separation.
- `.planning/phases/01-canonical-pure-spine/01-VERIFICATION.md:8-52` — fresh Phase 1 gate and named limitations.
- `.planning/spikes/001-ownership-kernel-workbench/README.md:66-92,164-218` — reduced move/loan controls, intermediate-fact lesson, and validated bounded oracle.
- `.planning/spikes/003-public-origins-generic-abilities/README.md:121-133,201-225,235-274` — independent ability propagation, immutability defect, and limits.
- `.planning/spikes/004-independent-certificate-checker/README.md:14-32,85-100,117-127,154-215` — trust boundary, mutations, scale, and explicit escape.
- `internal/compiler/{syntax,check,core,interp,cgen,native,session,evidence,protocol}` — current real compiler spine inspected in this session.

### Primary External Sources (MEDIUM confidence via official documentation)

- https://doc.rust-lang.org/reference/expressions.html#moved-and-copied-types — move/copy/deinitialization rules.
- https://doc.rust-lang.org/reference/expressions/operator-expr.html#borrow-operators — shared/mutable borrow access restrictions.
- https://doc.rust-lang.org/error_codes/E0382.html — use-after-move diagnostic precedent.
- https://doc.rust-lang.org/error_codes/E0505.html — move-while-borrowed diagnostic precedent.
- https://move-language.github.io/move/abilities.html — independent operation gates and structural/generic ability conditions.
- https://github.com/swiftlang/swift-evolution/blob/main/proposals/0427-noncopyable-generics.md — generic noncopyability design and complexity lessons.
- https://go.dev/doc/security/fuzz/ — deterministic fuzz target, seed corpus, minimization, and regression behavior.

## Metadata

**Confidence breakdown:**

- Standard stack: HIGH — installed tools and repository source were inspected; no package inference.
- Architecture: HIGH for the bounded linear shape — it ports reduced results from five real compiler layers and prior executable ownership work.
- Surface spelling: MEDIUM — intentionally provisional and selected for implementation economy, not human/AI comparative evidence.
- Native representation: MEDIUM — inline buffer is a reversible tracer representation, not a performance or ABI claim.
- Pitfalls: HIGH — the key test/oracle/immutability failures occurred in local executable spikes.

**Research date:** 2026-09-03
**Valid until:** milestone M001 semantic contract changes or a reduced Phase 2 counterexample invalidates the linear model
