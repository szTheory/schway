---
spike: 003
idea: ownership-kernel
name: public-origins-generic-abilities
type: comparison
validates: "Given separately compiled APIs that return borrowed or generic values, when a body-blind consumer checks origin, access, and ability-sensitive uses, then compact public summaries agree with an independent body oracle and reject dishonest summaries"
verdict: VALIDATED
related: [001, 002]
tags: [ownership, origins, generics, abilities, separate-compilation, higher-ranked, differential]
---

# Spike 003: Public origins and generic abilities

## What this validates

Given separately compiled APIs that return borrowed or generic values, a public
interface can carry enough information for a consumer to check:

- owned, single-source, multi-source, field-projected, and tagged borrowed
  results;
- shared versus exclusive returned access;
- borrowed values nested in `Option`-like wrappers;
- borrowing iterators and origin-preserving type erasure;
- higher-ranked callback access whose fresh origin cannot escape one call;
- noncopyable generic arguments; and
- independent `copy`, `drop`, `share`, `send`, and `escape` abilities.

The consumer receives a serialized interface with implementation bodies
removed. A separately written oracle evaluates implementation facts. The
producer checker rejects under-declared origins, borrowed-as-owned results,
exclusive-as-shared results, and retained fresh callback origins before an
interface can be published.

This validates a public-summary direction, not a final surface grammar or a
formal lifetime calculus. Variance, subtyping, async suspension, dynamic trait
objects, native layout, and recursive origin-bearing types remain outside this
spike.

## Research

Mojo's lifetime checker uses symbolic origins derived from values, permits
origin unions, and normally infers origin parameters. It also documents an
important cost: wildcard origins can disable early destruction and obscure
exclusivity. That makes value-derived paths strong public facts, but makes an
unbounded “any origin” escape hatch a poor ordinary default.

Rust's elision rules cover the common single-input case but reject ambiguous
multi-input borrowed returns. Rust also needs higher-ranked bounds when a
callback must work for every fresh lifetime. The useful lesson is not the
apostrophe syntax; it is that ordinary source and universally quantified
scoped access are genuinely different cases.

Swift's recent noncopyable and nonescapable work shows two things. First,
copyability and escapability are independent properties. Second, adding them
late requires changes through generics, protocols, existentials, `Optional`,
`Result`, spans, and iterators. Swift's borrowing-iteration work also records a
real limitation: a view tied to iterator-owned generated storage cannot simply
be returned as though it depended on the sequence.

Move demonstrates why abilities should independently gate operations and why
generic containers derive them from their arguments rather than inherit one
coarse “resource” bit.

- [Mojo lifetimes, origins, and references](https://mojolang.org/docs/manual/values/lifetimes/)
- [Rust lifetime elision](https://doc.rust-lang.org/reference/lifetime-elision.html)
- [Rust higher-ranked lifetime bounds](https://doc.rust-lang.org/reference/trait-bounds.html#higher-ranked-trait-bounds)
- [Swift noncopyable generics](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0427-noncopyable-generics.md)
- [Swift nonescapable types](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0446-non-escapable.md)
- [Swift borrowing iteration](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0516-borrowing-sequence.md)
- [Move type abilities](https://move-language.github.io/move/abilities.html)

| Approach | Strength | Weakness | Disposition |
|---|---|---|---|
| value-parameter origin paths | direct APIs, stable domain names, field paths and unions, almost no binder algebra | unions conservatively retain every possible source; higher-ranked access still needs a scoped quantifier | selected public default |
| explicit region binders | established formal precedent and direct expression of outlives constraints | introduces compiler-chosen names and constraints into routine interfaces | retain as possible core/formal representation, not leading source surface |
| restricted callback combinators | fresh lifetime is structurally scoped and cannot be stored | changes composition and control flow; wrapping every returned view creates a second API style | retain as a targeted tool for scoped resources and callbacks |
| inferred body-only relationships | no annotation burden | separate consumers cannot check without bodies; implementation changes silently alter API legality | rejected |
| wildcard/erased origin | easy migration escape | loses locality and can retain too much or disable useful checking | unsafe boundary only |

## Chosen semantic shape

Ordinary public borrowed results name value paths:

```text
fn head(input: borrow Buffer) -> borrow(input) View[Byte]

fn choose(
  left: borrow Buffer,
  right: borrow Buffer,
) -> borrow(left | right) View[Byte]

fn choose_tagged(
  left: borrow Buffer,
  right: borrow Buffer,
) -> Left(borrow(left) View[Byte]) | Right(borrow(right) View[Byte])

fn left_field(pair: borrow PairBuffer) -> borrow(pair.left) View[Byte]
```

An exclusive result carries access separately from origin:

```text
fn head_mut(input: borrow mut Buffer) -> borrow mut(input) View[Byte]
```

A callback that receives fresh scoped access introduces a quantifier only at
that pressure point:

```text
fn with_view[R](
  input: borrow Buffer,
  use: for access fn(view: borrow(access) View[Byte]) -> R
    where R: escape,
) -> R
```

The spelling is experimental. The stable decision is the normalized contract:
origin paths and alternatives, shared/exclusive access, a fresh callback
origin, and independently derived abilities.

## Gates

1. A serialized interface contains no body facts, yet the consumer agrees with
   the body oracle on every declared and generated call.
2. Producer verification rejects every deliberately dishonest summary.
3. Tagged alternatives retain origin precision that an unrefined union loses.
4. Ability propagation is checked for every one of the 32 possible ability
   combinations through generic wrappers.
5. Reordering functions/origins and alpha-renaming caller bindings preserve
   results across 500 seeded trials.
6. Omitting one origin from a union produces a small stable counterexample.
7. A 1,000-origin adversarial summary remains linear to export and consume;
   measurements are observations, not product budgets.
8. The comparison reports source-shape cost without treating token count alone
   as a usability verdict.

## How to run

From this directory:

```sh
go test ./...
go test -race ./...
go test -cover ./...
go vet ./...
go run ./cmd/origin-lab -pretty=false
go run ./cmd/origin-lab -show-signatures
go run ./cmd/origin-lab -pretty=false -inject-origin-bug
```

The injected-fault command intentionally exits unsuccessfully after returning
the reduced mismatch.

## What to expect

The normal CLI emits a compact JSON decision report:

- 24 deterministic call fixtures pass;
- 4 dishonest producer summaries are rejected;
- 66 generated consumer/oracle cases agree;
- all three signature encodings are measured over the same 13 APIs;
- a 1,000-origin public summary is exported and consumed.

Over this deliberately small surface corpus, value-origin syntax used 244
lexical tokens and one explicit binder. Explicit regions used 272 tokens and
13 binders. Callback-only encoding used 361 tokens and changed 8 of 13 direct
APIs into `with_*` forms. These counts show representation cost, not scientific
human or AI superiority.

The injected fault removes `right` from the declared origin set for `choose`.
After 19 generated cases, the consumer permits mutation of `right` followed by
use of the returned view while the body oracle rejects it with
`ownership.borrow_conflict`.

## Observability

Healthy output contains only gate counts, encoding measurements, the
adversarial-scale observation, and elapsed time. Full signatures require an
explicit flag. Full call evidence appears automatically only for a mismatch.
All functions, calls, bindings, origins, return alternatives, abilities, and
diagnostics have stable machine-readable identities.

## Investigation trail

### Iteration 1 — body-blind interface and independent oracle

Created a JSON interface boundary that strips bodies before consumer checking.
The consumer resolves named origin paths against caller bindings; the separate
oracle reads implementation return facts. Producer verification ensures that
the body-to-interface handoff cannot understate lifetime or access authority.

### Iteration 2 — compare three public encodings

Rendered value-origin, explicit-region, and callback-only forms from the same
normalized functions. Explicit regions were less noisy than a caricature of
Rust syntax, but still introduced 12 more binders than value paths. Universal
callback conversion was the largest representation and changed API shape,
which confirms it should remain a scoped tool rather than the only way to lend
values.

### Iteration 3 — origin alternatives and generic abilities

An origin union correctly keeps both possible owners borrowed. A tagged return
can instead attach `left` and `right` origins to separate alternatives, allowing
the inactive owner to end after pattern refinement. Generic `Option`, `Result`,
`List`, `View`, iterator, and unique-box rules preserve each ability
independently; 32 exhaustive atomic ability combinations check both recursive
evaluators.

### Iteration 4 — mutability and erasure

Added shared/exclusive access to return summaries. The producer rejects an
exclusive result declared as shared. An erased view keeps its origin and lack
of `escape`; erasing a data representation is not permission to erase lifetime
or authority facts.

### Iteration 5 — shallow-copy harness defect

The first fault injection changed the public summary and unexpectedly changed
later signature metrics. The exporter had stripped bodies but had shallow-copied
return slices, so interface mutation shared backing storage with implementation
facts. Deep-copying the interface boundary fixed the defect, and a regression
test now proves fault injection cannot mutate the oracle. This is a concrete
warning for the production compiler: summaries need immutable, content-owned
representations, not merely different top-level structs.

### Iteration 6 — scale and representation properties

Added 500 deterministic randomized trials for function order, origin order,
and binding alpha-renaming. A 1,000-source union produced a roughly 79 KB JSON
interface and was consumed in low single-digit milliseconds in these host runs. The
case confirms linear behavior, but also shows that enormous flat unions need a
diagnostic/summary compression policy before they become realistic APIs.

## Results

**Verdict: VALIDATED for value-path public origin summaries, independently
derived generic abilities, and targeted higher-ranked callback scopes.**

The leading contract is:

- infer local origins;
- make exported borrowed-result relationships part of the versioned interface;
- name public origins with parameter/field paths and tagged alternatives;
- carry shared/exclusive access independently;
- use origin unions as safe conservative joins, not as a substitute for
  path-sensitive result types;
- introduce a fresh quantified origin only for APIs that truly lend access to
  caller code;
- derive abilities structurally and keep `copy`, `drop`, `share`, `send`, and
  `escape` independent;
- preserve origins and abilities through wrappers and type erasure; and
- verify generated summaries against bodies before publication.

This is enough to proceed to an independently encoded certificate checker. It
does not justify freezing the shown keywords or promising that all higher-order
or async lending can remain equally compact.

## Known limitations

- The workbench has no subtyping, variance, recursive types, associated types,
  or generic constraint solver.
- Field origins are represented as paths but disjoint-field conflict precision
  is not modeled.
- Higher-ranked evidence covers one callback pattern. Returning borrows from
  nested callbacks, lending iterators, and async suspension remain open.
- Type erasure is semantic metadata in the fixture, not a vtable, object layout,
  or native ABI.
- Ability rules are trusted type-definition inputs to this spike. The real
  compiler must derive positive abilities from fields or require an audited
  unsafe witness; library authors must not assert them freely.
- The 1,000-origin case demonstrates linear work and representation growth, not
  an acceptable public design. Large repeated origin sets may need interned
  DAGs or named summary nodes.
- Signature token counts are descriptive. AI generation/repair and human audit
  comparisons remain part of the later surface experiment.
