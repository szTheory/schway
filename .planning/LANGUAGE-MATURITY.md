# Language Maturity — Where Codename Lang Actually Stands

**Assessed:** 2026-09-08, at the start of M002.
**Purpose:** stop re-discovering the gap between how sophisticated the
verification stack sounds and how little the language can currently express.
Planning vocabulary ("semantic spine", "interprocedural equivalence") describes
the assurance machinery, not the language surface. Read this before believing a
roadmap makes the language usable.

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
| Lang-to-Lang calls | every program is one function; M002 adds this |
| Recursion | refused by design in M002 (cycle refusal) |
| Numeric types beyond `Byte` | no integers, no floats |
| Strings, arrays, collections | absent as value types |
| Generics, stdlib, non-C I/O | deferred, unscheduled |

**Concrete calibration: you cannot write FizzBuzz.** No loop, no modulo, no
print without calling out to C. Hello-world is only reachable through a
`foreign C` declaration.

Corpus at assessment time: 58 `.lang` programs, 1,633 lines total (~28 lines
average, 193-line maximum), all single-function.

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
| 1 | Lang-to-Lang calls | **M002, in progress** (Phases 07-13) |
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

- Lang-to-Lang calls land (M002 Phase 07) — the "single function" framing dies.
- Arithmetic or iteration is added — the proportions move materially.
- The corpus stops being dominated by <50-line single-function programs.
- Any milestone closes.

Related: [STANDING-VERDICTS.md](STANDING-VERDICTS.md) — decisions already
researched, so they are not re-litigated each milestone.
