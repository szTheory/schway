# Language Maturity — Where Codename Lang Actually Stands

**Assessed:** 2026-09-08, at the start of M002.
**Re-assessed:** 2026-09-11, after Phase 10, at the Phase 11 planning gate.
**Re-assessed:** 2026-09-17, Phase 14 (EVD-06 landed a machine check for this
file's own counts; the guard total corrected 32→22 and the corpus figures
refreshed — see "The single-function guard inventory" below).
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

Corpus at re-assessment (2026-09-11): 89 programs, 3,096 lines (~35 lines
average, 167-line maximum). No longer all single-function — the largest is a
13-function call-graph fixture (`testdata/phase07/deep_diamond_acyclic.lang`),
and multi-function fixtures now exist for Phases 07, 08, and 10. The shape is
still fixtures, not programs: every one exists to exercise one admission rule,
and the biggest file is mostly comment.

Corpus at re-assessment (2026-09-18, plan 14-07): **128 `.lang` programs, 4,311 lines total** (~34 lines average, 193-line maximum). Growth since 2026-09-17 is plan 14-07's own two new testdata/phase14/ fixtures (the witness registry's executed probes); the shape claim above (fixtures, not programs) still holds. This count is machine-checked by `TestLanguageMaturityCountsAreCurrent` in `internal/compiler/session/self_describing_docs_test.go`, independently of the "re-verify cheaply" block below.

## The single-function guard inventory (re-verified 2026-09-20, EVD-06 machine check)

Phases 07-10 made `OpCall` real in `check`, `corevalidate`, `originvalidate`,
`pathoracle`, and (internally, via Go tests) `interp`. A two-function program
now passes `lang check` clean:

```bash
go run ./cmd/lang check testdata/phase07/call_basic.lang        # check pass
go run ./cmd/lang run --engine=interpreter testdata/…/call_basic.lang  # operational_failure
go run ./cmd/lang run --engine=native      testdata/…/call_basic.lang  # operational_failure
```

It cannot be executed by either engine. The refusal is **not** confined to the
two `cgen` entry points the roadmap names. A non-test scan finds **19 `len(Functions) != 1` guards across 6 files in 1 packages** (45 including tests):

| Package | Guards | Notable sites |
|---|---|---|
| `session` | 19 | `RunInterpreter`, `RunNative` (the CLI run path); `verifyOwnedCorpus`, `verifyBorrowedCorpus` (×8), `TransposeReleaseOrder`; Phase 5/6 verification lanes; `admitPhase5Candidate` |

**This table is now machine-checked, not self-certified.**
`internal/compiler/session/self_describing_docs_test.go`'s
`TestLanguageMaturityCountsAreCurrent` re-derives every number above
independently, in Go, by walking the module tree with `go/parser` — never by
executing the `awk` line below. The `awk` line is a text match and cannot tell
a real guard from a comment that merely *mentions* the pattern (this tree has
several, e.g. `session.go`'s own "the old len(program.Functions) != 1 guard"
prose), so it overcounts; it originally reported 32 where the Go test's
comment-immune AST walk finds 22. The Go test is now this file's authority —
treat any drift between the table above and its output as this file being
stale, not the test being wrong.

Re-verify (approximate only — see the overcounting note above; kept as a cheap
human convenience, not an authority): `awk '/^func /{f=$0;l=NR} /Functions\) != 1/{print FILENAME": "f}' $(find internal cmd -name '*.go' -not -name '*_test.go')`

Two consequences worth carrying into planning, one of them corrected by this
re-verification:

- **The reducer is no longer single-function.** Phase 11 (D-11-29/D-11-30)
  widened `reduce.Reduce` to accept a multi-function `Seed`: two whole-program
  moves (drop-call-site, drop-orphan-function) run first, then the original
  per-function moves loop over every function in the seed. `reduce` therefore
  no longer appears in the guard table above — its remaining `len(Functions)`
  comparisons use `<= 1` / `== 1`, single-function-path shortcuts rather than
  `!= 1` refusals, so they fall outside this table's predicate. This corrects
  this file's prior claim that "`reduce.Reduce` refuses a multi-function seed
  outright," which was accurate at the 2026-09-11 re-assessment and has since
  been overtaken by Phase 11.
- **The comparator and gate scaffolding are still mostly single-function.**
  The 20 `session` guards include the verification lanes that *are* the
  five-axis equivalence proof. Making multi-function programs runnable and
  making them *provable* remain separate costs.

### Re-verify cheaply — do this rather than trusting this file

As of 2026-09-17 (EVD-06), the guard-count and corpus figures above are
machine-checked by `TestLanguageMaturityCountsAreCurrent`
(`go test ./internal/compiler/session/... -run TestLanguageMaturityCountsAreCurrent -count=1`)
independently of the commands below — the commands below stay for humans
who want a fast, approximate sanity check, but they are no longer this
file's authority.

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
- ~~This file's own guard-count and corpus figures silently drift from the
  tree.~~ **FIRED repeatedly (32 claimed vs. 22 actual, corpus stale in both
  directions) — closed 2026-09-17 (EVD-06): `TestLanguageMaturityCountsAreCurrent`
  now re-derives every stated count independently and fails the suite if this
  file and the tree disagree, so this bullet can no longer silently fire
  again.**

Related: [STANDING-VERDICTS.md](STANDING-VERDICTS.md) — decisions already
researched, so they are not re-litigated each milestone.
