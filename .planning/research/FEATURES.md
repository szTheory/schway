# Feature Landscape: M005 Practical Computation

**Domain:** General-purpose systems language; bounded scalar computation and ordinary application output
**Researched:** 2026-10-02
**Confidence:** MEDIUM (milestone scope and current gaps come from project requirements and living direction; semantic comparisons use official language references; exact M005 input/output caps are a requirements-author decision)

## Scope and Evidence Labels

M005 is user-approved to make ordinary Schway source run bounded `sum_to_n`, then FizzBuzz with exact output. The current app route already accepts bounded U64 input. The milestone context says U64 constants exist but arithmetic, comparisons/Bool, scalar loops, remainder, and bounded text output do not; the checker refuses CFG cycles. The feature recommendations below are research inferences, not already approved semantic decisions. Keep loop-carried resource and loan state refused unless a narrow exception is independently supported.

Concrete acceptance witnesses should pin a `.schway` source, public build/run command, input bound, exact stdout bytes, stderr, and exit outcome. For example, `sum_to_n(10)` returns/prints `55`; `FizzBuzz(15)` emits lines `1`, `2`, `Fizz`, ..., `FizzBuzz`, ending in one newline. The user-facing maximum input must be explicit; it is not selected by existing project docs.

## Table Stakes

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Typed U64 addition, subtraction, comparison, equality, and Bool conditions | A loop counter and accumulator cannot be expressed without them | High | Same checked meaning across source checker, independent core/origin validators, interpreter, path oracle, and C emitter. Avoid C signed-overflow assumptions. |
| A scalar predicate loop with ordinary continuation | Users need a direct way to express `while i <= n { ... }` | High | A single understandable loop form plus conditional branching and assignment suffices. Require CFG fixed-point analysis of scalar state. Initially reject resource/loan changes crossing a back edge. |
| Defined arithmetic boundary behavior | `sum_to_n` can exceed U64 even when each input parses | High | Recommend checked overflow with explicit runtime failure and stable process outcome, plus a documented app input cap for the normal supported range. Do not let optimization flags change behavior. |
| Remainder with specified integer semantics and zero-divisor behavior | FizzBuzz requires divisibility checks (`i % 3 == 0`) | Medium | For U64 remainder define `x = q*y + r`, `0 <= r < y`; compile-time zero divisor is rejected, dynamic zero produces a defined arithmetic failure. |
| Minimal fixed text output and decimal U64 formatting | FizzBuzz and the sum example must be observable as ordinary programs | Medium | Fixed literals (`Fizz`, `Buzz`, `FizzBuzz`, newline), decimal formatting, bounded writes; preserve stdout/stderr and application status on the retained native route. No general string runtime needed. |
| Exact end-to-end independent answers | Parser/checker acceptance does not prove computation runs correctly | High | Pin expected output independently of both engines; compare interpreter and native execution; use controls that break transfer, loop fixed point, arithmetic boundary, and output truncation assumptions. |

## Differentiators

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| One small semantic core shared by `sum_to_n` and FizzBuzz | A runnable gain demonstrates that the new operations compose in source, not just as isolated fixtures | High | Prefer one checked operator/loop/output path instead of fixture-specific lowering. `sum_to_n` is the first smaller end-to-end slice; FizzBuzz proves comparison, remainder, branches, and output together. |
| Static input bound plus runtime checked arithmetic | Keeps normal app use predictable and still makes unexpected arithmetic overflow defined | Medium | The input cap is a product/runtime contract; checked operations remain required because callers or future source paths can exceed assumptions. A proof-only cap is not a substitute for runtime semantics. |
| Small output surface with explicit byte bound | Gives deterministic resource use and portable output without introducing `String`, Unicode, or a global runtime | Medium | Derive max output bytes from the input bound; reject before partial output or define a stable partial-write failure rule. Avoid allocating the full output if line-at-a-time writing suffices. |
| Loops whose simple state is independently re-derived | Supports future computation without prematurely widening ownership guarantees | High | Fixed-point analysis and reached mutation controls are valuable Schway-specific evidence differentiators. State the refusal boundary for loans/resources in diagnostics. |

## Semantic Choices and Recommendation

| Decision | Recommendation | Tradeoffs and precedent |
|----------|----------------|-------------------------|
| U64 addition/accumulator overflow | Checked failure, same in interpreter and native, plus an explicit supported input cap | Wrapping is portable if deliberately specified (Go defines unsigned operations modulo `2^n`) but silently returns a wrong sum. Saturation is deterministic but silently changes arithmetic too. Static bounds alone are brittle when call limits or code change. Swift's default traps and Rust's checked/wrapping/saturating APIs show that policy can be explicit; avoid Rust build-mode-dependent overflow behavior. Sources: [Go integer overflow](https://go.dev/ref/spec#Integer_overflow), [Swift integer bounds and overflow](https://docs.swift.org/swift-book/LanguageGuide/TheBasics.html), [Rust operator overflow](https://doc.rust-lang.org/reference/expressions/operator-expr.html), [Rust `u64` checked operations](https://doc.rust-lang.org/std/primitive.u64.html). |
| Division/remainder by zero | Reject constant zero statically; a dynamic zero has a defined checked arithmetic error and never reaches C `/` or `%` | Go specifies runtime panic on integer zero divisor; C-family behavior is not a portable contract. For unsigned operands specify quotient/remainder identities and ensure emitter guards dynamic zero. Source: [Go integer operators](https://go.dev/ref/spec#Integer_operators). |
| Loop form and control flow | Admit a condition-controlled `while`-style loop and the conditional branch form required to update/exit; defer `for`, iterator protocols, labels, `break`/`continue`, and loop-valued expressions | Go and Rust both document predicate loops; a `while` loop maps directly to a CFG header/condition/body/back edge and is easiest to review. Keep initial loop form narrow, but do not hardwire analysis to one syntax node. Sources: [Go for statements](https://go.dev/ref/spec#For_statements), [Rust predicate loops](https://doc.rust-lang.org/reference/expressions/loop-expr.html). |
| Loop state | Scalar U64/Bool locals only in the first fixed-point slice; resource/loan carry over a back edge remains refused | Prevents M004 cleanup/borrow state from being accidentally accepted by CFG convergence. Add resource loop states only after acquisition identity and cleanup order are independently tested. |
| Output format and bounds | Fixed UTF-8/ASCII byte literals, unsigned decimal formatting, `\n` line endings, output cap derived from accepted `n`; no locale dependence | For a manageable FizzBuzz target, a small chosen cap (e.g. `n <= 1,000`) bounds worst-case numeric output to roughly 21,000 bytes. This is an example recommendation, not a committed limit. Choose before requirements are finalized. Preserve ordinary stream semantics already established by the app route. Sources: [Go `fmt` output destinations and errors](https://pkg.go.dev/fmt), [Go `io.Writer`](https://pkg.go.dev/io#Writer). |
| Same milestone or defer FizzBuzz | Keep both in M005, staged behind the first runnable `sum_to_n` gain | This scope is explicitly user-approved. FizzBuzz is the smallest natural consumer that composes remainder, comparison, branching, loops, literals, decimal output, and exact text. Do not make it block shipping the earlier sum slice; each phase should deliver a runnable program. |

## Anti-Features

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| Implicit wrap, saturate, or target/compiler-dependent arithmetic | Can produce plausible but incorrect results; build flags/toolchain could change outcomes | Checked defined failure; use explicit wrapping/saturating operations later only when a concrete bit-level consumer needs them |
| Unchecked C arithmetic as the language definition | C behavior and optimizer assumptions are not a suitable semantic authority, especially around overflow and zero division | Guard or lower checked U64 operations so the C path implements the Schway contract exactly |
| General-purpose `for` iterators, `break`, `continue`, labels, and nested-loop guarantees in the initial slice | Adds control-flow/liveness surface beyond the two witnesses | One predicate loop and simple if/else first; admit more forms when a runnable consumer requires them |
| Resource acquisition, release, or loan state across loop back edges | Can invalidate cleanup order, dynamic identity, and borrow guarantees | Refuse clearly with source attribution while scalar fixed-point analysis lands |
| Full string/Unicode/formatting runtime or generic collections | FizzBuzz needs only a few fixed strings and decimal output; these abstractions add ownership, encoding, and allocation decisions | Fixed byte literals, decimal U64 formatter, bounded writes; defer general text to a byte/string consumer |
| Unbounded FizzBuzz output or whole-output allocation | Caller can request huge work/output and defeat bounded resource posture | Fix max input and output bytes; stream each complete line or fail under an explicit output rule |
| Treating agreement between interpreter and C as the only oracle | Shared bug can make both engines agree | Independent expected answers and negative controls for loop transfer, zero division, overflow, and output boundary |

## Feature Dependencies

```text
U64 comparisons + Bool + conditional continuation → condition-controlled scalar loop
U64 addition + checked overflow semantics          → sum_to_n accumulator
remainder + conditional branch                     → FizzBuzz classification
fixed literals + decimal U64 + bounded write         → exact FizzBuzz stdout
CFG fixed-point scalar facts                        → accepted scalar loop back edge
independent answers + reached mutation controls       → credible native/interpreter claim
```

## MVP Recommendation

1. **Runnable `sum_to_n` slice:** add defined checked U64 addition/comparison, Bool branching, scalar `while`, and decimal output. Pick and document the maximum accepted input (or reject larger inputs before execution); verify `0 -> 0`, a representative value such as `10 -> 55`, max accepted input, first rejected input, and checked overflow behavior. Establish one public native app route and exact stdout/exit contract.
2. **Scalar loop analysis boundary:** solve and independently validate scalar facts to a CFG fixed point; add loop-carried-state mutation controls. Keep any resource or loan transfer across back edges structurally refused.
3. **FizzBuzz runnable composition:** add U64 remainder, fixed literals, the output cap, and a source program that emits exact lines for `1..=15`; cover 0/1/15 and the maximum accepted input, and use independent expected bytes for interpreter and native.

Defer generic `for` and iterator support, nested/labelled control transfer, signed arithmetic, multiplication, division, general strings/Unicode, dynamic allocation for output, arbitrary IO, and resource/loan loop carries. These are not required by the approved examples.

## Stakeholder Fan-Out: Main Risks and Review Questions

| Lens | Review concern | Concrete question for M005 contract |
|------|----------------|-------------------------------------|
| Product | A successful compile without an ordinary runnable example does not meet the milestone | Which public command runs the source, and what exact output does `sum_to_n(10)` produce? |
| Compiler/runtime | CFG cycles need convergence, and all six semantic consumers must agree | Which facts are fixed-point facts, and which states stay forbidden over a back edge? |
| Security | User-controlled U64 and output can amplify work or produce huge streams | What input and byte limits reject oversized work before observable partial output? |
| Portability | Native C, host integer types, locale, and line endings can leak into semantics | Are U64 operations unsigned and width-fixed, and are output bytes locale-independent with `\n`? |
| Evidence | Differential agreement may mask a shared implementation defect | What separate expected-answer fixture and changed-assumption control fail if loop transfer or overflow behavior is wrong? |
| Usability/AI authoring | Narrow syntax should be obvious and errors actionable | Can the common loop be written in a few predictable lines, with unsupported resource carry reported at its source? |
| Human audit | Output and control flow need to be inspectable without hidden coercions | Can the reviewer see loop condition, update, cap, and output bound directly in source/contract? |
| Language/ecosystem | Surface compatibility is not the goal; coherent semantics are | Does each admitted construct earn its semantic and evidence cost through one of these two runnable programs? |

## Sources

- Go Language Specification — integer overflow and integer operators: https://go.dev/ref/spec#Integer_overflow and https://go.dev/ref/spec#Integer_operators (official language specification; accessed 2026-10-02).
- Go Language Specification — for statements: https://go.dev/ref/spec#For_statements (official language specification; accessed 2026-10-02).
- Rust Reference — operator overflow and loop expressions: https://doc.rust-lang.org/reference/expressions/operator-expr.html and https://doc.rust-lang.org/reference/expressions/loop-expr.html (official language reference; accessed 2026-10-02).
- Rust standard library — `u64` checked/wrapping/saturating arithmetic APIs: https://doc.rust-lang.org/std/primitive.u64.html (official standard library docs; accessed 2026-10-02).
- Swift Book — integer bounds and overflow behavior: https://docs.swift.org/swift-book/LanguageGuide/TheBasics.html (official language book; accessed 2026-10-02).
- Go standard library — output destinations and writer errors: https://pkg.go.dev/fmt and https://pkg.go.dev/io#Writer (official standard library docs; accessed 2026-10-02).
- Project-internal approved scope and implementation frontier: `.planning/PROJECT.md`, `.planning/PRODUCT-ROADMAP.md`, `.planning/LANGUAGE-MATURITY.md`, `.planning/REQUIREMENTS.md` (2026-10-02 planning context; requirements file still describes M004 active set and its Future Requirements section contains M005 candidates).
