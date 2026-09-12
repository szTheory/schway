# Language Maturity — Where Codename Lang Actually Stands

**Assessed:** 2026-09-08, at the start of M002.
**Re-assessed:** 2026-09-11, after Phase 10, at the Phase 11 planning gate.
**Purpose:** stop re-discovering the gap between how sophisticated the
verification stack sounds and how little the language can currently express.
Planning vocabulary ("semantic spine", "interprocedural equivalence") describes
the assurance machinery, not the language surface. Read this before believing a
roadmap makes the language usable.

**Second trap, same shape:** `wiki/example-tour.md` shows effect rows
(`with { Inventory, Payments }`), `?` propagation, `defer ... unless`, generics,
named arguments, and inline `spec`/`property` blocks. **None of those tokens are
in the lexer.** It is a design target and the README says so ("candidate designs,
not settled specifications"), but it reads like a language description. When
calibrating what exists, read `internal/compiler/syntax/token.go`, never the tour.

## The two axes

A language has two independent axes: **expressiveness** (what you can say) and
**assurance** (how much you can trust what it does). Most languages climb
expressiveness first and retrofit assurance later, or never — Rust took years,
C++ never did.

This project deliberately inverted that. **Assurance is unusually far along;
expressiveness is near zero.** Every progress question should be answered
against both axes or it will be misleading.

## Snapshot, verified 2026-09-08

### Assurance — real and working

Lossless parser + formatter · typed IR with per-function CFGs · affine
ownership and borrow checker · **two independent re-implementations** of the
admission rules (`corevalidate`, `originvalidate`) · deterministic interpreter
as semantic oracle · C17 codegen through Clang at `-O0`/`-O3`/`-flto` ·
ASan/UBSan lane · five-axis differential comparator · core-level HDD reducer ·
content-bound evidence manifests + verdict-free cache · structured diagnostics
as a versioned API · `lang explain` bounded cause DAG · `lang-repair`
automated repair through the JSON protocol · changed-risk lane selection ·
measured p50/p95/CoV feedback budget.

Independent re-derivation at trust crossings and differential testing against
an interpreter oracle are rare even in funded language projects. This is the
project's actual asset.

### Expressiveness — the honest inventory

Complete keyword set (`internal/compiler/syntax/token.go`):

```
module export fn data type let mut match try take borrow discard because defect foreign
```

What that list does **not** contain, and what therefore does not exist:

| Missing | Consequence |
|---|---|
| `if` / `else` | `match` is the only control flow |
| `while` / `for` / `loop` | **no iteration of any kind** |
| Arithmetic operators (`+ - * / %`) | **you cannot add two numbers** |
| Lang-to-Lang calls | **partially landed** — multi-function programs now *check* (Phases 07-10) but cannot *run* on either engine; see "The single-function guard inventory" below |
| Recursion | refused by design in M002 (cycle refusal) |
| Numeric types beyond `Byte` | no integers, no floats |
| Strings, arrays, collections | absent as value types |
| Generics, stdlib, non-C I/O | deferred, unscheduled |

**Concrete calibration: you cannot write FizzBuzz.** No loop, no modulo, no
print without calling out to C. Hello-world is only reachable through a
`foreign C` declaration.

Corpus at first assessment (2026-09-08): 58 `.lang` programs, 1,633 lines
total (~28 lines average, 193-line maximum), all single-function.

Corpus at re-assessment (2026-09-11): **89 programs, 3,096 lines** (~35 lines
average, 167-line maximum). No longer all single-function — the largest is a
13-function call-graph fixture (`testdata/phase07/deep_diamond_acyclic.lang`),
and multi-function fixtures now exist for Phases 07, 08, and 10. The shape is
still fixtures, not programs: every one exists to exercise one admission rule,
and the biggest file is mostly comment.

## The single-function guard inventory (verified 2026-09-11)

Phases 07-10 made `OpCall` real in `check`, `corevalidate`, `originvalidate`,
`pathoracle`, and (internally, via Go tests) `interp`. A two-function program
now passes `lang check` clean:

```bash
go run ./cmd/lang check testdata/phase07/call_basic.lang        # check pass
go run ./cmd/lang run --engine=interpreter testdata/…/call_basic.lang  # operational_failure
go run ./cmd/lang run --engine=native      testdata/…/call_basic.lang  # operational_failure
```

It cannot be executed by either engine. The refusal is **not** confined to the
two `cgen` entry points the roadmap names. A non-test scan finds
**32 `len(Functions) != 1` guards across 6 files in 3 packages** (55 including
tests):

| Package | Guards | Notable sites |
|---|---|---|
| `session` | 26 | `RunInterpreter`, `RunNative`, `interpreterInputs` (the CLI run path); `verifyBorrowedCorpus` (×9); Phase 5/6/7 verification lanes; `VerifyAliasFalseNoAlias` |
| `cgen` | 4 | `Emit`, `EmitNative`, `emitLinear`, `emitBranchOperations` |
| `reduce` | 2 | `Reduce` (hard error on a multi-function seed), `ProjectSource` (returns an unsupported-projection string) |

Re-verify: `awk '/^func /{f=$0;l=NR} /Functions\) != 1/{print FILENAME": "f}' $(find internal cmd -name '*.go' -not -name '*_test.go')`

Two consequences worth carrying into planning:

- **The reducer is single-function.** `reduce.Reduce` refuses a multi-function
  seed outright. Any phase that promises HDD reducer behaviour on
  multi-function programs must widen `reduce` as well as `cgen`.
- **The comparator and gate scaffolding are single-function.** The 26 `session`
  guards include the verification lanes that *are* the five-axis equivalence
  proof. Making multi-function programs runnable and making them *provable* are
  separate costs.

### Re-verify cheaply — do this rather than trusting this file

```bash
grep -oE '"[a-z_]+"' internal/compiler/syntax/token.go | sort -u   # keyword set
find . -name '*.lang' -not -path './.git/*' | wc -l                # corpus size
wc -l $(find . -name '*.lang' -not -path './.git/*') | tail -1     # corpus lines
grep -nE 'Plus|Minus|Star|Slash|Percent' internal/compiler/syntax/token.go  # arithmetic: empty == still absent
```

## Rough proportions

Judgments, not measurements — useful for calibration, not for reporting:

- As a **verification platform**: ~60-70% of the foundational machinery exists.
- As a **language you could write a program in**: ~5-10%.

## Distance to "usable", in dependency order

| # | Capability | Status |
|---|---|---|
| 1 | Lang-to-Lang calls | **M002, partially landed** — admitted and checked (07-10); *executable* is Phase 11 |
| 2 | Arithmetic + real numeric types | **not started, not scheduled** |
| 3 | Iteration (loops, or admitted bounded recursion) | not started; M002 actively refuses recursion |
| 4 | Strings, arrays, collections | not started |
| 5 | Modules / separate compilation | M003 lead candidate |
| 6 | Ability-bounded generics | deferred (ranked #1 of post-M002 features) |
| 7 | Standard library | deferred |
| 8 | I/O beyond raw foreign C calls (effect surface) | deferred, flagged premature |

Items 2-4 are what actually make the language writable, and **none of them are
on any roadmap yet.** That is the single most important thing this file
records. Expect several more milestones the size of M002.

## Strategic read, and the real risk

**The ordering is defensible.** Adding arithmetic to a sound ownership core is
comparatively easy; retrofitting ownership soundness onto an expressive
language is what historically takes years or never lands. The bet is that the
hard part is the part that got built first.

**The real risk is different, and it is not "we built the wrong thing."** An
assurance stack this elaborate has only ever been exercised against 28-line,
single-function, loop-free programs. Whether `loanLivenessFixpoint`, the
five-axis comparator, the HDD reducer, and the cache invalidation story hold up
once programs have real shape is genuinely unknown. **M002 is the first real
load test of machinery built in the absence of load** — which is exactly why
spike S-006 (cost-scaling probe) hard-gates Phase 08's planning.

Watch for the tell: analysis or corpora that were linear on one function
becoming superlinear across a call graph. Go 1.18's generics front-end
regression and Rust's Polonius performance wall are the named precedents.

## When this file goes stale

Rewrite the snapshot when any of these happen:

- ~~Lang-to-Lang calls land (M002 Phase 07) — the "single function" framing dies.~~
  **FIRED 2026-09-09; snapshot refreshed 2026-09-11.** Calls are admitted and
  checked but not executable, so the framing narrowed rather than died.
- Arithmetic or iteration is added — the proportions move materially.
- The corpus stops being dominated by <50-line single-function programs.
- Any milestone closes.

Related: [STANDING-VERDICTS.md](STANDING-VERDICTS.md) — decisions already
researched, so they are not re-litigated each milestone.
