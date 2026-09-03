---
id: syntax-cognitive-ergonomics
title: Syntax and cognitive ergonomics
summary: Three source-interface candidates and a measured path toward calm, canonical code for expert human audit and precise agent manipulation.
type: design
status: candidate
confidence: mixed
created: 2026-09-02
updated: 2026-09-02
tags: [syntax, readability, formatting, ai, ergonomics]
related: [vision, feature-coherence, example-tour, compiler-feedback-latency, compute-efficiency-constitution, convergence-work-program, research-ledger, open-questions]
---

# Syntax and cognitive ergonomics

## Recommendation

Prototype three genuinely different source interfaces over one semantic model.
The leading candidate is a conventional expression-oriented surface with
braces, sparse punctuation, named arguments, and canonical formatting. Keep an
indentation-oriented projection as a serious challenger. Keep a uniform tree
projection as an agent/debug interchange, not the default human surface unless
testing overturns that recommendation.

Do not claim that camel case, snake case, braces, indentation, or Lisp syntax is
scientifically best in isolation. Readability studies are task-, population-,
font-, and familiarity-dependent; a systematic review found divergent results
for identifier style. Cognitive Dimensions supplies a better design method:
evaluate visibility, consistency, viscosity, hidden dependencies, error
proneness, and role-expressiveness across the notation and its tools together.
([systematic review](https://arxiv.org/abs/2208.12141),
[Cognitive Dimensions](https://www.cl.cam.ac.uk/~afb21/CognitiveDimensions/papers/Green1989.pdf))

## Visual contract

Candidate conventions:

- `snake_case` for values, functions, fields, effects, and file names;
- `PascalCase` for types, variants, components, actors, and capabilities;
- acronyms behave like words: `HttpRequest`, `JsonCodec`, `order_id`;
- lower-case dotted module paths: `shop.orders.checkout`;
- one canonical form for calls, records, variants, imports, pipes, and matches;
- formatter-owned line breaking, trailing commas, indentation, and blank lines;
- no manual column alignment, because renames create unrelated diff churn;
- blank lines may mark semantic phases in specs and procedures;
- named arguments for multi-role calls; exact-name punning allows `reservation:`;
- no user-defined precedence, operators, or sigils in the initial language;
- typed literals such as time, regex, SQL, or bytes may earn a fixed sigil/form
  only when a compile-time parser produces a typed, inspectable node;
- Unicode is welcome in strings and comments, while identifiers use a declared
  script/confusable policy and canonical normalization.

Unicode publishes mechanisms for detecting visually confusable identifiers,
including mixed- and whole-script cases. A compiler should diagnose these and
show code points; visual calm cannot come at the price of identifier spoofing.
([Unicode UTS #39](https://www.unicode.org/reports/tr39/))

## Surface A — conventional semantic prose

```text
fn place_order(command: PlaceOrder)
  -> Result<Order, PlaceOrderError>
  with { Inventory, Payments, Orders, Clock }
{
  let reservation = Inventory.reserve(items: command.lines)?

  let payment = Payments.charge(
    amount: command.total,
    idempotency_key: command.id,
  )?

  let order = Order(
    id: command.id,
    lines: command.lines,
    status: Confirmed(payment:, at: Clock.now()),
  )

  Orders.save(order:)?
  Ok(order)
}
```

This surface borrows familiar scanning landmarks from Rust, Swift, Gleam, and
Elixir without copying any one grammar. Delimiters make formatter recovery and
incremental parsing straightforward. Named roles and vertical argument layout
expose information without Java-like declarations everywhere.

Strengths: familiar to current models and engineers; robust error recovery;
clear diffs; explicit scope; maps naturally to common editors. Costs: closing
delimiter noise; `with` plus `?` plus typed signatures can become visually busy;
method/module capitalization needs one crisp rule.

## Surface B — indentation and keyword flow

```text
place_order command: PlaceOrder
  returns Result<Order, PlaceOrderError>
  uses Inventory, Payments, Orders, Clock

  reservation = Inventory.reserve items: command.lines ?

  payment = Payments.charge
    amount: command.total
    idempotency_key: command.id
  ?

  order = Order
    id: command.id
    lines: command.lines
    status: Confirmed payment:, at: Clock.now()

  Orders.save order: ?
  Ok order
```

This surface treats indentation as the dominant grouping signal and replaces
punctuation with words. It is closer to a restrained Elixir/ML/Python family
and can feel exceptionally breathable for pipelines and declarations.

Strengths: low punctuation; visual hierarchy and data flow dominate; compact.
Costs: multiline call boundaries and postfix `?` are less obvious; formatter
recovery from incomplete edits is harder; whitespace damage matters; adjacent
named arguments can resemble declarations. It risks becoming terse but
ambiguous unless the grammar is stricter than it first appears.

## Surface C — uniform semantic tree

```text
(fn place_order
  (inputs (command PlaceOrder))
  (returns (Result Order PlaceOrderError))
  (with Inventory Payments Orders Clock)
  (let reservation
    (? (Inventory.reserve (items (. command lines)))))
  (let payment
    (? (Payments.charge
      (amount (. command total))
      (idempotency_key (. command id)))))
  (let order
    (Order
      (id (. command id))
      (lines (. command lines))
      (status (Confirmed (payment payment) (at (Clock.now))))))
  (? (Orders.save (order order)))
  (Ok order))
```

This is a uniform, lossless source tree. Macros and transformations are easier
to define structurally; partial programs can remain data; an agent need not
infer precedence or syntactic category.

Strengths: tiny grammar; direct structural editing; uniform syntax; excellent
canonical interchange. Costs: human role words become visually submerged;
parenthesis density harms scanning; domain code reads more like representation
than prose; compact examples do not prove large-system readability. This
surface is most compelling as `program.tree` output or an optional projection.

## Comparison

| Criterion | Surface A | Surface B | Surface C |
|---|---:|---:|---:|
| Familiarity and model prior | high | medium-high | medium |
| Incremental parse recovery | high | medium | high |
| Human visual breathing | high | highest when well formatted | low-medium |
| Scope/call boundary visibility | high | medium | high |
| Direct structural editing | medium | medium | highest |
| Punctuation noise | medium | low | high/repetitive |
| Semantic role visibility | high | high | medium |
| Risk of dialects/precedence | low with restrictions | medium | low |
| Likely source-token count | medium | low | high |

The recommendation is A as the baseline, carrying B's careful vertical rhythm.
C should exist as a canonical machine projection even if humans never author
it. All three must lower to exactly the same typed/effect tree; otherwise this
is three languages rather than three interfaces.

## Specs should read like evidence, not a test framework

The preferred style is flat, independent, active-voice, and ordinary. A spec
has visible arrange, act, and assert paragraphs without nested `describe`,
shared-example magic, hook precedence, or a second metaprogrammed language.

```text
spec "declined payment releases the reservation" {
  let inventory = Inventory.memory(stock: fixtures.in_stock)
  let world = CheckoutWorld(
    inventory:,
    payments: Payments.decline(reason: CardDeclined),
    orders: Orders.memory(),
    clock: Clock.fixed(at: @2026-09-02T12:00:00Z),
  )

  let result = run place_order(command: fixtures.order) with world

  expect result == Err(PaymentFailed(reason: CardDeclined))
  expect inventory.events == [Reserved, Released]
}
```

The formatter preserves the semantic paragraph breaks. Helpers are ordinary
pure functions. Setup lifetime and effect overrides are explicit values. Table
and property specs are first-class data-driven forms, not macro nesting.

## Syntax slop the compiler can reject

- redundant parentheses, aliases, optional semicolons, and formatter choices;
- multiple spellings for the same import, call, lambda, or collection;
- shadowing in a scope unless explicitly requested;
- punning when the local and parameter types/roles do not match exactly;
- wildcard imports and hidden extension-method namespaces;
- user-defined truthiness, coercion, precedence, and implicit receiver lookup;
- confusable identifiers and mixed naming style in public APIs;
- meaningless one-letter public names, placeholder names, and stale generated
  comments, with a narrow waiver for mathematical/local idioms;
- test nesting and hooks that make execution order implicit.

Some of these are objective compiler errors; name quality is a lint/evidence
obligation because taste cannot be made sound.

## Evaluation protocol

Use the same 15–25 programs, semantic model, formatter, and diagnostic protocol
for every surface. Measure whole repair loops rather than isolated token count:

- first-pass parse/type/spec success across several models;
- tokens and wall time through successful repair;
- human time and accuracy answering effect, failure, and dependency questions;
- eye movement or at least task completion on realistic diffs, not snippets;
- diff churn after rename, parameter addition, formatting, and reordered fields;
- parser recovery and diagnostic quality on incomplete AI edits;
- stable AST mapping and comment preservation;
- subjective calm/joy after the objective tasks, not instead of them.

No syntax freezes before this experiment. The semantic design is allowed to
converge first. The full decision inventory, ownership-dependent syntax cases,
and exit criteria are in the
[convergence work program](convergence-work-program.md).
