# M003 Research: DX and the Agent Authoring Loop

**Researched:** 2026-09-17
**Mode:** Ecosystem + adversarial
**Tools actually run:** yes — `go build ./...` clean, binaries built to `/tmp/lang` and `/tmp/lang-repair` from `908bac3`, and every command output quoted below is real terminal output from this session, not reconstructed.
**Overall confidence:** HIGH on the local findings (I ran them), MEDIUM on the prior-art framing (secondary sources, cited).

---

## Verdict (the single highest-leverage DX move for M003)

**Ship Surface Honesty: one `lang capabilities` command that emits the closed, machine-readable language surface, plus one new diagnostic class — `surface.not_in_language` — emitted at the exact token that is off-surface, carrying the surface manifest's own vocabulary as `causes` and a `RequiresConfirmation` repair.** Everything else on the candidate list (more repair classes, LSP, REPL, richer `explain`, faster feedback) is downstream of a defect I reproduced empirically and which is *worse than "the error message is unhelpful"*: **the feedback channel is currently a constant function over the agent's edit.** Three structurally different hallucinated programs — `if/else`, `if` with the `else` deleted, and `if` with the braces deleted — produced **byte-identical diagnostics with identical diagnostic IDs and an identical result ID**. An agent in a repair loop edits, re-runs `lang check`, receives the same 314 bytes, and has no signal that anything changed. That is not a slow loop; it is a non-convergent loop, and it violates the core value's "without wasting iteration time" clause more directly than any other property of the toolchain. The fix is cheap — the surface is already a closed, enumerable set in the source (21 token kinds, 4 built-in type constructors, one statement grammar) — and it is the only DX item that gets *better*, not obsolete, as M003 widens the language: the manifest is a build artifact derived from the lexer and checker, so every `if`/`for`/`+` that M003 adds updates it for free, and every one it does not add keeps saying so by name.

---

## The agent loop today, traced end to end (with real output)

### Step 0 — What the agent is told before it writes anything

Nothing. `AGENTS.md` (96 lines) is entirely GSD workflow boilerplate and the PROJECT.md core value. It contains **zero tokens of Lang**. `README.md` (68 lines) is a link index to 40 wiki pages, one of which — `wiki/example-tour.md` — is design fiction that shows effect rows, `?` propagation, generics, and named arguments, **none of which are in the lexer** (`LANGUAGE-MATURITY.md:11-17`). An agent that reads the repo's own front door to learn the language is actively misled. This is the starting condition for every observation below.

### Step 1 — Happy path, a two-function program

```
$ /tmp/lang check testdata/phase07/call_basic.lang
result:8e32f3c4482509a832d79652 check pass module=s1:phase07.call_basic:module:phase07.call_basic
metrics elapsed_ns=0 peak_rss=unavailable output_bytes=177 recomputed_work=109
```

```
$ /tmp/lang --json check testdata/phase07/call_basic.lang
{"schema":"lang.command/1","command":"check","status":"pass","id":"result:8e32f3c4482509a832d79652","module_id":"s1:phase07.call_basic:module:phase07.call_basic","diagnostics":[],"executions":[],"lanes":[],"metrics":{"elapsed_ns":0,"peak_rss_status":"unavailable","output_bytes":307,"recomputed_work":109}}
```

```
$ /tmp/lang run --engine=interpreter testdata/phase07/call_basic.lang
result:ec6da087fb9d10ae048fd907 run pass
s1:phase07.call_basic:fn:identity:op:0:event:returned function.returned source_place=s1:phase07.call_basic:fn:identity:place:0 type_id=s1:phase07.call_basic:fn:identity:type:0
s1:phase07.call_basic:fn:main:op:1:event:returned function.returned source_place=s1:phase07.call_basic:fn:main:place:1 type_id=s1:phase07.call_basic:fn:main:type:0
metrics elapsed_ns=0 peak_rss=unavailable output_bytes=458 recomputed_work=1
```

**Good.** One-line envelope, stable `schema` field, bounded output, an addressable result ID, zero external deps, sub-10ms wall clock. The `--json` flag is a first-class mode, not an afterthought. Note that `LANGUAGE-MATURITY.md:88-92` still records multi-function programs as non-executable; **that is now stale** — Phase 11 landed and the interpreter runs them.

### Step 2 — The good defect: interprocedural loan liveness

```
$ /tmp/lang check testdata/phase13/derivation_interprocedural_loan_defect.lang
result:525b442b64c831d3abd67fc7 check invalid
diagnostic:064f1f93c4b82807b68f2bcd check.interprocedural_loan_liveness [1805:1811]: cannot transfer ownership while an interprocedurally-extended loan is still live
```

JSON (abridged to the diagnostic):

```json
{
  "schema": "lang.diagnostic/1",
  "id": "diagnostic:064f1f93c4b82807b68f2bcd",
  "code": "check.interprocedural_loan_liveness",
  "severity": "error",
  "primary_span": { "start": 1805, "end": 1811 },
  "message": "cannot transfer ownership while an interprocedurally-extended loan is still live",
  "causes": [
    { "kind": "borrow_created_here", "span": { "start": 1775, "end": 1781 } },
    { "kind": "loan_extended_by_call", "detail": "s1:phase13.derivation:fn:user", "span": { "start": 1827, "end": 1831 } },
    { "kind": "callee_return_contract", "detail": "s1:phase13.derivation:fn:user:parameters[0].mode=owned" }
  ],
  "repairs": [
    {
      "kind": "move_after_interprocedural_loan",
      "span": { "start": 1784, "end": 1841 },
      "replacement": "let result = user(borrowed)\n  let delivered = take buffer",
      "applicability": "MachineApplicable"
    }
  ]
}
```

**This is genuinely good, and better than most production compilers on the structured-output axis.** A three-step cause chain with per-cause spans and a cross-function `detail`, plus a complete, applicable edit with a rustc-shaped applicability tag. Very few compilers ship a *cause graph* in JSON at all; rustc ships `spans` and `children`, not a typed causal vocabulary.

### Step 3 — The repair driver closes the loop

```
$ /tmp/lang-repair --lang=/tmp/lang --source=/tmp/.../d.lang
status=repaired diagnosis=check.interprocedural_loan_liveness repair=move_after_interprocedural_loan subprocess_count=2
exit=0
```

Source after (real bytes):

```
fn escort(buffer: Buffer) -> Buffer {
  let borrowed = borrow buffer
  let result = user(borrowed)
  let delivered = take buffer
  result
}
```

```
$ /tmp/lang check .../d.lang
result:7d3dd9acb7fb7a929d054155 check pass module=s1:phase13.derivation:module:phase13.derivation
```

**The agent loop works, on the classes it covers.** Two subprocesses, deterministic, exit code 0, and the repaired source re-checks clean. This is a real, shipped, non-fictional agent tool loop. It deserves credit; most language projects at this maturity have nothing like it.

### Step 4 — `lang explain` on the same defect

```
$ /tmp/lang explain testdata/phase13/derivation_interprocedural_loan_defect.lang diagnostic:064f1f93c4b82807b68f2bcd
lang.explain/0 root=diagnostic:064f1f93c4b82807b68f2bcd nodes=4 edges=3
  diagnostic:064f1f93c4b82807b68f2bcd kind=check.interprocedural_loan_liveness availability=available
  diagnostic:064f1f93c4b82807b68f2bcd:cause:2 kind=callee_return_contract availability=not_captured
  diagnostic:064f1f93c4b82807b68f2bcd:cause:0 kind=borrow_created_here availability=available
  diagnostic:064f1f93c4b82807b68f2bcd:cause:1 kind=loan_extended_by_call availability=available
```

The JSON adds `function_id`, `function_name`, and `blame: true` per node — that is DX-05, and it is real. **But the human projection above is strictly less informative than the `--json check` output it explains.** It drops the `detail` strings, drops the message prose, drops the spans, and adds only an `availability` column. It renders identifiers, not a story. A user running `explain` to understand *why* gets a graph topology, not an explanation.

### Where the loop is good

| Property | Status | Evidence |
|---|---|---|
| Structured output as a first-class mode | Strong | `--json` on every subcommand; `schema` field on every record |
| Determinism | Strong | 3 identical `--json check` runs, `shasum` `0cc34d18…` each time |
| Bounded output | Strong | largest error output measured at 2,060 bytes; `output_bytes` budget row is 65,536 |
| Latency | Strong | `real 0.00` on the 13-function `deep_diamond_acyclic.lang` fixture, three runs |
| Addressable IDs | Strong | `diagnostic:`, `result:`, `s1:module:fn:place:` are all joinable via `query` |
| Repair applicability taxonomy | Strong | closed 3-value vocabulary, `DriverEligible` fails closed |
| Cause graphs on ownership defects | Strong | typed `kind`s with spans and cross-function `detail` |
| Zero external dependencies | Strong | `go build ./...` clean, no network |

### Where the loop is thin

| Gap | Evidence |
|---|---|
| **Spans are raw byte offsets only** | `[1805:1811]` on a file whose comment header is ~1,700 bytes. No line, no column, no source excerpt, no caret. A human cannot use this without a script; an agent must re-read and index the file itself. |
| **No `notes` / `help` / secondary labels** | `diagnostic.Diagnostic` (diagnostic.go:95-104) has exactly `Schema, ID, Code, Severity, Primary, Message, Causes, Repairs`. There is no field in which "did you mean" or "here is what is legal here" could travel. This is a type-level gap, not an unwritten-message gap. |
| **Repairs exist only on the ownership path** | `ErrorWithRepairs` has **19 call sites, all in `internal/compiler/check/check.go`**. Zero syntax diagnostics carry a repair, and all 51 `syntax.*` codes go through `Error()` on schema `/0` with no causes and no repairs. |
| **`lang-repair` returns nothing on the majority case** | On a hallucinated program: `{"status":"unrepairable","subprocess_count":1}` — the `diagnosis` field is *empty*. The repair tool tells the agent nothing at all about what it refused to repair. |
| **`explain` degenerates to a tautology on syntax errors** | `lang.explain/0 root=… nodes=1 edges=0` — one node, zero edges, restating the code. |
| **No capability surface** | `lang` has 11 subcommands; none of them answers "what can I write?". The closest thing is `query`, whose closed vocabulary (`diagnostic, debugmap, evidence, control, lane`) is about *artifacts*, not about the *language*. |
| **`metrics elapsed_ns=0` always** | Determinism was bought by zeroing the wall-clock field. The tool's own metrics channel therefore cannot report the very latency the project's top constraint names. Real timing must come from outside the protocol. |
| **Usage text is a 589-byte single line** | The `--help` path returns `tool.usage` as one unwrapped line listing 11 subcommands. Fine for an agent, hostile to a human. |

---

## The hallucination problem: what an agent gets when it writes `if`, and what it should get

This is the highest-leverage finding in the document, and it is worse than expected.

### `if` / `else`

```
$ /tmp/lang check if.lang
result:4256447a1d35446f4e8044b4 check invalid
diagnostic:e5d48d30900cfaf10602a2cc syntax.expected_rbrace [83:88]: expected `}`
diagnostic:f34ba9ea2b5f96f560264419 syntax.expected_declaration [89:90]: expected `data` or `fn` declaration
```

Byte 83:88 is `'value'`. Byte 89:90 is `'{'`. **The diagnostic does not point at `if` at all.** Root cause, confirmed by a second experiment:

```
$ cat id.lang   # function body is the single token `if`
$ /tmp/lang check id.lang
diagnostic:0ac6fdd6d2fd96b7ec802529 name.unknown [71:73]: linear result is unknown
```

`if` lexes as a plain `identifier`. The parser consumes it as the function's linear-result expression, then trips on the *next* token. Every hallucinated keyword therefore produces a diagnostic **one token late, blaming code the agent wrote correctly.** `return value` reproduces the identical shape:

```
$ /tmp/lang check ret.lang
diagnostic:cf5a5cd74276f96d72256425 syntax.expected_rbrace [77:82]: expected `}`
diagnostic:777311d9e6d4c792a4b0f46a syntax.expected_declaration [83:84]: expected `data` or `fn` declaration
```

### The spiral, reproduced

I simulated the two most plausible agent repairs of "expected `}`" — delete the `else` block (brace-balance theory), then delete the braces:

| Iteration | Program | Output |
|---|---|---|
| 1 | `if value { value } else { value }` | `result:4256447a1d35446f4e8044b4` · `diagnostic:e5d48d30900cfaf10602a2cc` `[83:88]` + `diagnostic:f34ba9ea2b5f96f560264419` `[89:90]` |
| 2 | `if value { value }` | `result:4256447a1d35446f4e8044b4` · **identical** |
| 3 | `if value` | `result:4256447a1d35446f4e8044b4` · **identical** |

**Three structurally distinct programs, byte-identical output, identical diagnostic IDs, identical result ID, identical spans.** The result ID is content-bound over the diagnostic set, and the diagnostic ID is content-bound over `{schema, code, span, causes}` — so two different programs that fail at the same byte offsets are *indistinguishable in the protocol*. That property is intentional and correct for caching. It is catastrophic for an agent loop, because the agent's only convergence signal is "did the output change," and here it provably cannot.

`lang-repair` gives no rescue:

```
$ /tmp/lang-repair --lang=/tmp/lang --source=if.lang --json
{"status":"unrepairable","subprocess_count":1}
```

Empty `diagnosis`. And `explain`:

```
$ /tmp/lang explain if.lang diagnostic:e5d48d30900cfaf10602a2cc
lang.explain/0 root=diagnostic:e5d48d30900cfaf10602a2cc nodes=1 edges=0
  diagnostic:e5d48d30900cfaf10602a2cc kind=syntax.expected_rbrace availability=available
```

### The other three hallucination classes

| Written | Real output | Verdict |
|---|---|---|
| `for i in 0..10 { total = total + i }` | **9 diagnostics, 2,060 bytes**, four of them `syntax.unexpected_byte: unexpected source character` at single-byte spans (the `.`s and the `+`), plus cascading `expected_binding_name` / `expected_equal` / `expected_binding_source` / `expected_rbrace` / `expected_declaration` | Worst case. No error mentions `for`, iteration, or arithmetic. Pure cascade; the agent must guess which of 9 is the real one. |
| `let sum = value + value` | `syntax.unexpected_byte [90:91]` then `syntax.expected_linear_result [90:91]` then two cascade errors | `+` is literally "unexpected source character". Does not say arithmetic operators do not exist. |
| `let greeting = "hello"` | `syntax.expected_binding_source [87:94]: expected \`identifier\`` | Single clean diagnostic (credit where due) but it says "expected identifier", not "string literals are not a value form in this language". |
| `fn main(value: Int) -> Int` | `type.unknown [47:52]: unknown type constructor "Int"` | **The best of the set** — names the offending token and the category. Still does not enumerate the four constructors that *do* exist (`Byte`, `Buffer`, `Box<_>`, `Pair<_,_>`, per `ability.go:120-168`) plus user `data` declarations. |
| `let r = helper(value)` (undeclared) | `core.call_callee_unresolved [82:88]: call target does not resolve to a declared function` | Good code, good message, no "declared functions are: main". |

**The pattern is consistent: the toolchain knows exactly what is legal at every one of these sites and never says so.** `type.unknown` has the constructor table in scope. `core.call_callee_unresolved` has the call graph in scope. `name.unknown` has the scope map in scope. `syntax.unexpected_byte` has the token table in scope. Not one of them spends a single byte of the 65,536-byte output budget on it.

### What it should get

For `if value { … }`:

```json
{
  "schema": "lang.diagnostic/1",
  "code": "surface.not_in_language",
  "severity": "error",
  "primary_span": { "start": 79, "end": 81 },
  "message": "`if` is not in this language's surface; it lexes as an ordinary identifier and no binding named `if` is in scope",
  "causes": [
    { "kind": "token_lexed_as", "detail": "identifier" },
    { "kind": "surface_version", "detail": "lang.surface/1@0e1f…" },
    { "kind": "nearest_construct", "detail": "match — the only branching form in this language" },
    { "kind": "cascade_suppressed", "detail": "4 downstream syntax diagnostics suppressed; re-run with --no-suppress" }
  ],
  "repairs": [
    { "kind": "use_match_instead_of_if", "span": { "start": 79, "end": 132 },
      "replacement": "match value {\n  … => value\n}",
      "applicability": "RequiresConfirmation" }
  ]
}
```

Four properties matter and all four are cheap:

1. **The span points at the off-surface token**, not one token later.
2. **The message names the token and the reason** — this is the Rust/Elm "name the thing, then say what to do" shape.
3. **Cascades are suppressed and counted.** The `for` loop's 9 diagnostics should be 1 + a count. Elm deliberately shows one error at a time for exactly this reason ([elm-lang.org](https://elm-lang.org/news/compiler-errors-for-humans)); Becker et al.'s readability work identifies message length and cascade noise as first-order factors ([Denny, Becker et al., CHI 2021](https://dl.acm.org/doi/fullHtml/10.1145/3411764.3445696)).
4. **The diagnostic ID now differs per hallucinated token**, because the code and span differ — which breaks the constant-function spiral at the protocol level, not by convention.

The detection rule is a whitelist, not a heuristic: maintain a closed set of *known off-surface tokens* (`if else while for loop return + - * / % && || ! break continue struct class impl trait async await import use pub const static enum interface function var def lambda yield throw catch finally`) and fire `surface.not_in_language` when the lexer produces an `identifier` whose text is in that set, or `unexpected_byte` on a character in the known-operator set. Cost: one map lookup in the lexer and one new diagnostic constructor. **There is no analysis involved.**

---

## Capability discovery: design for how an agent learns the language surface

An agent cannot grep for what a language *cannot* do. It needs a positive, closed, versioned statement. Anthropic's own tool-design guidance makes this the central point: tools are "contracts between deterministic systems and non-deterministic agents," and "helpful error messages that guide agents toward better queries help agents self-correct without wasting tool calls" ([Anthropic, *Writing effective tools for AI agents*](https://www.anthropic.com/engineering/writing-tools-for-agents)).

### Proposed: `lang capabilities [--json]`, schema `lang.surface/1`

Everything below is already a closed set in the source and can be **derived, not hand-written** — which is the whole point, because a hand-written manifest is `example-tour.md` again.

```json
{
  "schema": "lang.surface/1",
  "surface_digest": "0e1f4a…",
  "toolchain_commit": "908bac3",
  "keywords":        ["because","borrow","data","defect","discard","export","fn","foreign",
                      "let","match","module","mut","take","try","type"],
  "type_constructors": [
    {"name":"Byte","arity":0}, {"name":"Buffer","arity":0},
    {"name":"Box","arity":1},  {"name":"Pair","arity":2}
  ],
  "operators": [],
  "literals":  ["identifier"],
  "control_flow": ["match"],
  "not_in_language": [
    {"token":"if",    "category":"control_flow", "nearest":"match", "tracking":"none"},
    {"token":"for",   "category":"iteration",    "nearest":null,    "tracking":"LANGUAGE-MATURITY.md#distance-to-usable row 3"},
    {"token":"+",     "category":"arithmetic",   "nearest":null,    "tracking":"LANGUAGE-MATURITY.md#distance-to-usable row 2"},
    {"token":"string","category":"literal",      "nearest":null,    "tracking":"row 4"},
    {"token":"return","category":"control_flow", "nearest":"trailing linear-result expression"}
  ],
  "diagnostic_codes": ["check.interprocedural_loan_liveness", "… 228 total …"],
  "repair_kinds": [
    {"kind":"move_after_interprocedural_loan","applicability":"MachineApplicable"},
    {"kind":"wrap_call_in_try","applicability":"MachineApplicable"}
  ],
  "applicability_levels": ["MachineApplicable","RequiresConfirmation","Unspecified"],
  "commands": ["format","check","run","evidence","verify","interface","debug-map","explain","query","capabilities"],
  "examples": [
    {"name":"minimal","source":"module m\n\nexport {\n  fn main\n}\n\nfn main(value: Byte) -> Byte {\n  value\n}\n"},
    {"name":"call","source":"…"},
    {"name":"borrow_then_take","source":"…"},
    {"name":"match_result","source":"…"}
  ]
}
```

### Information-architecture rules for the manifest

- **It must be derived from the implementation and test-enforced against it.** The repo already has the enforcement pattern: `TestQueryMintsNoSixthVocabulary` asserts the dispatch table's key set equals `QueryVocabularies()`. Do the same: `TestSurfaceManifestMatchesLexer`, `TestSurfaceManifestMatchesTypeConstructors`, `TestEveryDiagnosticCodeIsInTheManifest`. A manifest that can drift is worse than none, because an agent will trust it.
- **`not_in_language` is the load-bearing half.** A positive list tells a model what to use; a negative list is what actually suppresses hallucination, because it addresses the specific tokens the model's prior pushes it toward. This is the section with no analogue in any existing language's tooling, and it is the section this project uniquely needs.
- **Include 4-6 complete, checking, runnable example programs inline.** Few-shot examples are the highest-density format for a model. They also serve the human reader. Derive them from `testdata/` and gate them in CI with `lang check` so they cannot rot.
- **One artifact, two readers.** Emit `--json` for the agent and a calm ~60-line plain-text projection for the human (`lang capabilities` with no flag). Same source, same digest.
- **Version and digest it.** `surface_digest` lets a diagnostic reference the exact surface it was produced against, lets the evidence manifest bind to it, and lets an agent cache it.
- **Wire it into `AGENTS.md`.** One generated block: "Run `lang capabilities` before writing Lang. Do not trust `wiki/example-tour.md`." Cheapest possible first-five-minutes fix, and it can ship in a single plan.

Prior art worth naming: Unison's codebase manager treats the codebase as "a structured object … not just a mutable bag of text files" and gives interactive discovery of what exists ([unison-lang.org](https://www.unison-lang.org/docs/the-big-idea/)). Lang cannot afford a structured editor, but it can afford the *query surface* that makes one useful. LSP's `initialize` capabilities handshake is the same idea at protocol level: the server states what it supports before the client asks for anything.

---

## Diagnostics quality vs the best in the field (gap table)

| Dimension | Best in field | Lang today | Gap |
|---|---|---|---|
| **Source excerpt + caret** | rustc, Clang, Elm all render the offending line with an underline; Clang's expressive diagnostics were an explicit differentiator vs GCC | Byte offsets only: `[1805:1811]` | **Large.** No line/col anywhere in `Span`. Trivially derivable from the lossless parser. |
| **Named token in the message** | Elm: "compilers should be assistance, not adversaries" ([elm-lang.org](https://elm-lang.org/news/compiler-errors-for-humans)) | `syntax.unexpected_byte: unexpected source character` | **Large.** The byte is known and not printed. |
| **"Did you mean" / nearest-match** | rustc and Clang both ship Levenshtein suggestions; Roc treats this as table stakes ([roc-lang.org](https://www.roc-lang.org/)) | None anywhere. `type.unknown "Int"` does not suggest `Byte` | **Large**, and the candidate sets are all in scope at emission. |
| **Enumerate what *is* legal here** | rustc's `expected one of …`; TypeScript's suggestion diagnostics | Never | **Large** and unusually cheap here, because every relevant set is closed. |
| **Cascade suppression** | Elm shows one error at a time; rustc suppresses derived errors after a parse failure | None — `for` produced 9 diagnostics from one mistake | **Large.** `syntax.too_many_errors` exists as a code, so a limiter mechanism exists; it is a ceiling, not a cause-aware suppressor. |
| **Machine-applicable structured fixes** | rustc's `Applicability` 4-value enum drives `cargo fix` ([rustc docs](https://doc.rust-lang.org/nightly/nightly-rustc/rustc_lint_defs/enum.Applicability.html)) | 3-value closed vocabulary, `DriverEligible` fails closed, 2 shipped classes | **Small.** Design is sound; coverage is narrow. |
| **Typed causal graph in JSON** | Nobody ships this | `causes[]` with `kind`/`detail`/`span`, cross-function, bounded | **Lang is ahead.** This is the project's genuine diagnostic asset. |
| **Stable error codes + `--explain`** | `rustc --explain E0382` gives a page of prose and examples | 228 codes exist; `lang explain` explains a *diagnostic instance*, not a *code* | **Medium.** `lang explain --code check.interprocedural_loan_liveness` is a missing, cheap command. |
| **Determinism / reproducibility** | Not a design goal for most compilers | Byte-identical across runs; content-bound IDs | **Lang is ahead**, with the agent-loop caveat above. |
| **Plain language / low jargon** | Becker et al. identify jargon, length, sentence structure, vocabulary as the measurable readability factors ([CHI 2021](https://dl.acm.org/doi/fullHtml/10.1145/3411764.3445696)) | "cannot transfer ownership while an interprocedurally-extended loan is still live" — accurate, and dense | **Medium.** Acceptable for the ownership core; a calm second sentence would help ("`buffer` is still on loan to `user` at this point"). |
| **One diagnostic type able to carry help** | rustc `Diagnostic` has `children`, `notes`, `suggestions` | `Diagnostic` has no `notes`/`help`/`labels` field | **Structural.** Adding one optional `notes []Note` field unblocks half this table at once, and — critically — can be done **without moving any diagnostic ID**, because identity is computed over `{Schema, Code, Span, Causes, RepairKinds}` only (diagnostic.go:109-144). The precedent is already set: `Span`, `Replacement`, and `Applicability` were deliberately excluded from identity for exactly this reason. |

The last row is the key engineering insight: **the schema was designed so that prose can be added without a schema bump or an ID churn.** The `/1` design already anticipated this. M003 can improve every message in the toolchain and break nothing.

---

## Auto-repair trust model (applicability levels, blast radius)

### What exists, and it is good

`diagnostic.go:40-96` defines a closed 3-value vocabulary — `MachineApplicable`, `RequiresConfirmation`, `Unspecified` — with three properties worth preserving verbatim:

- `NormalizeApplicability` maps empty → `Unspecified`, never → `MachineApplicable`. Fails closed.
- `DriverEligible` requires **all four** of `MachineApplicable` + non-empty `Kind` + `Span` + `Replacement`. A repair that claims machine-applicability without the material to apply it is refused.
- `ValidateRepair` refuses any value outside the vocabulary.

Compared against rustc's four-value enum — `MachineApplicable`, `MaybeIncorrect`, `HasPlaceholders`, `Unspecified` ([rustc_lint_defs::Applicability](https://doc.rust-lang.org/nightly/nightly-rustc/rustc_lint_defs/enum.Applicability.html)) — Lang collapses `MaybeIncorrect` and `HasPlaceholders` into `RequiresConfirmation`.

### The one gap M003 should close: `HasPlaceholders`

The moment `surface.not_in_language` ships, it will want to emit repairs like:

```
match value {
  … => value
}
```

That is a **placeholder** repair: not "maybe wrong," but "structurally incomplete by construction, and must never be spliced." Collapsing it into `RequiresConfirmation` loses a real distinction, and rustfix has an open, long-lived issue about exactly how to handle non-`MachineApplicable` suggestions safely ([rustfix#200](https://github.com/rust-lang/rustfix/issues/200); [cargo#13023](https://github.com/rust-lang/cargo/issues/13023)) — i.e. the industry has not solved the *lenient* direction, which is a good reason to stay strict and instead be *more precise*.

**Recommendation: add a fourth value, `HasPlaceholders`, with a hard invariant that a repair carrying it is never driver-eligible and its `Replacement` must contain at least one declared placeholder marker.** This is a widening of a closed vocabulary, which this repo has a precedent pattern for (`QLT02MetricVocabulary` widened twice, each time with a test proving the two chokepoints agree).

### Blast-radius rules worth writing down explicitly

The repo enforces most of these in code already; M003 should state them as invariants so they survive new repair classes:

1. **Single-span, single-file.** No repair may span more than one function today, and none may touch a file other than `--source`. `move_after_interprocedural_loan`'s span `[1784:1841]` is one contiguous statement pair inside one body — keep that property as a checked invariant, not a coincidence.
2. **Post-condition gate.** The driver must re-run `check` after applying and revert if the program does not reach `pass`. `subprocess_count=2` on the successful run is evidence this already happens; make it a named, tested invariant.
3. **Idempotence.** Applying a repair twice must be a no-op or a refusal. The D-13-10a finding — a repair whose replacement was byte-identical to the original — is precisely the failure mode a stated idempotence rule catches at design time rather than empirically in plan 13-05.
4. **No semantic-change repairs without confirmation.** `move_after_interprocedural_loan` is safe *only because* `take` is a pure compile-time ownership transfer and never a `Drop` (D-13-11). That argument is recorded in a fixture comment. It should be a required field: every `MachineApplicable` repair kind carries a `behavior_preservation_argument` string in the surface manifest, and a test asserts none is empty.
5. **Bounded iteration.** `lang-repair` should declare a maximum repair-apply count per invocation and report it, so an agent can distinguish "converged" from "gave up."
6. **Never auto-apply into `foreign C` blocks.** FFI is the one place a wrong splice is not merely a failed check but a memory-safety hazard. The security posture in PROJECT.md ("unsafe operations, FFI … stay explicit") argues this should be a structural refusal in `DriverEligible`, not a convention.
7. **`unrepairable` must say why.** The empty `diagnosis` field on the hallucination path is a trust bug as much as a DX bug: a tool that silently declines teaches an agent to ignore it.

---

## Human reviewability: what exists, what is missing

The core value names a human reviewer. Auditing AI-written Lang is a distinct job from writing it, and the toolchain is currently much better at the machine half.

### What exists

- **Canonical formatting.** `lang format --check` with proven losslessness and idempotence through generated programs and malformed AI edits (PROJECT.md Key Decisions). This is the single most valuable reviewability property a language can have, because it makes diffs semantic rather than stylistic. It is done and it is good.
- **Content-bound evidence manifests.** `evidence --validate MANIFEST FILE [--expand]`, with a cache "structurally incapable of holding a verdict."
- **`lang explain`'s cause DAG** with `function_id`/`function_name`/`blame` per node (DX-05).
- **`lang query`** joining five ID vocabularies — a real audit primitive.
- **`lang interface export/core/check`** — body-stripped, digest-bound signature summaries; a reviewer can diff the *interface* separately from the *implementation*.

### What is missing, ranked by how much a real reviewer would feel it

1. **A rendered diagnostic.** No line, no column, no source excerpt. A reviewer reading a CI log sees `[1805:1811]` and has to open an editor and compute a byte offset. This is the most basic reviewability gap in the toolchain and it is a few hours of work.
2. **No provenance channel.** Nothing records *that a repair was machine-applied* or *which one*. After `lang-repair` runs, the source file is indistinguishable from hand-written code. A reviewer auditing AI-written Lang most wants to know: which bytes did a human intend, which did a model generate, and which did the repair driver splice? A `.lang.provenance.json` sidecar — or better, an entry in the evidence manifest — recording `{repair_kind, applicability, span, pre_digest, post_digest, toolchain_commit}` per applied repair would be a genuinely novel and genuinely useful artifact, and it is directly in the spirit of the content-bound evidence layer that already exists.
3. **No review view.** There is no single command that says "here is what changed, here is what the checker now proves about it, here is which evidence lanes ran." `verify CORPUS` is close but is corpus-shaped, not change-shaped. A `lang review BASE..HEAD` producing a bounded, deterministic summary — functions touched, interface digests changed vs unchanged, lanes selected by `risk_lanes.json`, repairs applied — would be the reviewer's `git diff`.
4. **The human projections are machine-shaped.** `explain`'s human output is a list of node IDs. `debug-map` prints `s1:phase07.call_basic:fn:main:debug:1 kind=unknown`. `query`'s is the same. These are *the calm projections of machine cause graphs* the core value asks for, and right now they are neither calm nor projections — they are the machine form with the JSON braces removed. Plain-language rendering of the cause DAG ("`buffer` was borrowed at line 38; that loan is still live at line 40 because `user` takes its parameter by `owned`; so the `take` on line 39 cannot happen yet") is a well-defined, testable transformation from data the toolchain already has.
5. **No `--explain CODE`.** A reviewer encountering `check.interprocedural_loan_liveness` for the first time has no place to read what it means. 228 codes, zero documentation surface.

---

## DX metrics worth adopting (cheap, reusing the existing measure protocol)

The repo already has everything needed: `measure.Summary` (p50/p95/CoV), `measure.Verdicts()` = `{blocking, observed, not_ratified}`, `measure.Demote()`, `measure.GateEligibleMetrics()`, `session.QLT02MetricVocabulary()`, and `qlt02_budget_manifest.json` with `{machine_id, metric, gate_type, value_or_bound, unit, ratified_at, ratified_by_commit}` rows. **Every metric below is a new row in that manifest, not new infrastructure.** Note the existing discipline: only deterministic, machine-independent metrics may ever be `gate_type: hard`; wall clock and byte counts are permanently `observed`. Respect that — it is why the gates mean something.

| # | Metric | Definition | Gate type | Why it is cheap | What it catches |
|---|---|---|---|---|---|
| 1 | **`diagnostic_distinctness`** | Over a held-out corpus of *distinct* defective programs, the ratio of unique diagnostic-ID sets to programs. Target 1.0. | **hard** (deterministic, machine-independent) | Run `--json check` over a directory, hash the diagnostic ID set, count collisions. ~40 lines. | **The exact spiral I reproduced.** Today this metric is <1.0 and provably so. It is the single best one-number proxy for "can an agent tell that its edit did something." |
| 2 | **`surface_coverage`** | Fraction of the `not_in_language` token list that produces a `surface.not_in_language` diagnostic whose primary span covers the offending token. | **hard** | One fixture per token, generated from the manifest itself. | Regression when the lexer changes; also forces the manifest and the diagnostics to stay in sync. |
| 3 | **`cascade_ratio`** | Diagnostics emitted per *injected* defect, over the existing mutation/injector corpus (`session_phase13_injectors.go` already exists). p50 and p95. | **observed** | The injectors are built. Count diagnostics. | Today `for` scores 9. Elm's "one error at a time" principle expressed as a number. |
| 4 | **`repair_convergence_steps`** | `subprocess_count` from `lang-repair`, p50/p95, over the repairable corpus. | **hard** (deterministic) | Already printed by the tool. Just record it. | Silent regressions in repair quality; distinguishes "converged in 2" from "thrashed to the cap." |
| 5 | **`repair_coverage`** | Fraction of the defect corpus reaching `status=repaired`. | **hard** | Loop the corpus. | The honest successor to DX-07's "at least three classes" — a *rate*, not a count, so it cannot be gamed by adding a class that never fires (which is precisely how `use_matching_argument` failed). |
| 6 | **`diagnostic_bytes_per_defect`** | `output_bytes` / injected defects, p50/p95. | **observed** | Already in the metrics line. | Context-window cost of the loop. Measured today: 2,060 bytes for one `for` loop. |
| 7 | **`first_fix_latency`** | Wall clock for `check` → `lang-repair` → `check` on the repairable corpus, cold and warm, p50/p95/CoV. | **observed** (never hard — `elapsed_ns` is machine-dependent by existing rule) | `measure.Samples` already computes this. | The PROJECT.md constraint's literal text: "report cold and warm distributions rather than one flattering number." |
| 8 | **`hallucination_recovery_rate`** | Over a fixed set of N hallucinated programs, the fraction where a *scripted, non-LLM* repair policy driven only by the JSON output reaches `check pass` or `unrepairable` **with a non-empty `diagnosis`** within K steps. | **hard** if the policy is deterministic | A ~100-line script and a corpus of ~20 programs I have partly written in this session. | This is the end-to-end number the verdict is about. It is currently 0. |

Two adoption rules, both consistent with existing practice: add each row with `gate_type: observed` first, ratify to `hard` only after a measured baseline exists (the `peer_closure_recomputed_work_growth_exponent` precedent — declared in one plan, measured in another, ratified at a gate); and pair each hard gate with a negative control proving it is not inert (the `TestSeedEntryHazardIsReal` precedent).

---

## Role-lens disagreements

**DevRel vs the compiler-diagnostics specialist.** DevRel wants the first five minutes fixed: a real README quickstart, `lang capabilities`, working examples, and `AGENTS.md` telling an agent not to read `example-tour.md`. The diagnostics specialist wants line/column, source excerpts, and cascade suppression — deeper, slower, more valuable per-message. **Resolution:** they are not in conflict; the capability manifest is a one-plan item and the diagnostic work is a multi-plan item. Ship the manifest first because it changes the *distribution* of errors an agent hits, and the diagnostic work then pays off on a smaller, better-shaped error population.

**AI-tooling engineer vs the caching/determinism architect.** The agent-loop lens says the constant-function spiral is a top-severity bug. The determinism architect will correctly point out that content-bound diagnostic IDs are *load-bearing* for the evidence/cache layer and must not become nondeterministic. **Resolution — and this is the one real tension in the document:** the collision is not caused by content-binding, it is caused by the *identity content being too coarse* (`{schema, code, span, causes}` with byte-offset spans). Adding `surface.not_in_language` with a token-specific code and a token-specific span makes the three spiral iterations produce three different IDs **while remaining perfectly content-bound and deterministic.** Do not weaken determinism; enrich the content.

**Security lens vs product lens.** Product wants more auto-repair (it demos well, it is the thing that makes someone try this). Security wants the `MachineApplicable` set to stay small and the FFI refusal to be structural. **Resolution:** widen `RequiresConfirmation`/`HasPlaceholders` repairs freely — they are *suggestions with structure*, which is most of the product value for an agent that can evaluate them — and keep the driver-eligible set conservative. The agent does not need the tool to apply the fix; it needs the tool to *state* the fix precisely.

**Accessibility/plain-language vs the precision lens.** "cannot transfer ownership while an interprocedurally-extended loan is still live" is exact and dense. **Resolution:** do not soften the `message`; add a `notes` field and put the calm sentence there. Two audiences, two fields, one truth. The schema already supports adding this field without ID churn.

**Technical writer vs everyone.** The writer's objection is that a manifest which can drift from the implementation is a liability, not an asset — `example-tour.md` is already the proof. **Conceded fully, and it is the strongest constraint on the design:** the manifest must be *derived* and *test-enforced*, or it must not ship.

---

## Prior art and lessons (with sources)

- **Elm — "compilers should be assistance, not adversaries."** Elm's error-message work is the canonical reference point, and it demonstrably moved the field: it is credited with pushing Rust toward better diagnostics. Its concrete practices — show the source region, use plain language, show one error at a time, offer a hint — are the specific behaviors Lang lacks. ([elm-lang.org/news/compiler-errors-for-humans](https://elm-lang.org/news/compiler-errors-for-humans), [Changelog #218](https://changelog.com/podcast/218))
- **Rust — `Applicability` as a typed trust contract.** Four values: `MachineApplicable` ("definitely what the user intended, or maintains the exact meaning"), `HasPlaceholders` ("contains placeholders like `(...)` … will not result in valid Rust"), `MaybeIncorrect` ("might lead to new errors when applied"), `Unspecified`. `cargo fix`/`clippy --fix` apply only the first. Lang's 3-value version is faithful and fails closed; the missing value is `HasPlaceholders`, which the hallucination diagnostic will immediately need. ([rustc_lint_defs::Applicability](https://doc.rust-lang.org/nightly/nightly-rustc/rustc_lint_defs/enum.Applicability.html), [clippy discussion #9994](https://github.com/rust-lang/rust-clippy/discussions/9994))
- **rustfix's unsolved lenient direction.** There are long-running open issues on whether and how to auto-apply non-`MachineApplicable` suggestions. The lesson is defensive: the ecosystem with the most experience has *not* found a safe general rule, so Lang should invest in precision of classification rather than breadth of auto-application. ([rustfix#200](https://github.com/rust-lang/rustfix/issues/200), [cargo#13023](https://github.com/rust-lang/cargo/issues/13023))
- **Becker et al. on error-message research.** The systematic review ("Compiler Error Messages Considered Unhelpful") and the follow-on readability work identify **message length, jargon, sentence structure, and vocabulary** as the measurable factors, and — importantly for this project's metric design — catalogue five reasons enhancement studies produce conflicting results, including "researchers are measuring the wrong thing" and "the effects are hard to measure." That is a direct argument for metric #1 (`diagnostic_distinctness`), which measures a *structural* property rather than a subjective one. ([ITiCSE-WGR 2019](https://dl.acm.org/doi/10.1145/3344429.3372508), [CHI 2021](https://dl.acm.org/doi/fullHtml/10.1145/3411764.3445696))
- **Roc — error messages as a stated product pillar.** Roc explicitly aims to replicate Elm's "compiler is your friend, not your examiner" stance, and pairs it with nonblocking compilation (the program still runs where possible despite type errors). The nonblocking idea is *not* right for Lang — assurance is the asset — but "friendly errors are a headline feature, not polish" is exactly the framing the Key Decisions table already commits to. ([roc-lang.org](https://www.roc-lang.org/), [roc-lang.org/faq](https://www.roc-lang.org/faq))
- **Anthropic on tool design for agents.** "Tools represent a fundamentally new software paradigm: contracts between deterministic systems and non-deterministic agents … agents may hallucinate, misunderstand purposes, or call tools incorrectly." And specifically: "helpful error messages that guide agents toward better queries help agents self-correct without wasting tool calls." Lang's `check`/`repair` pair *is* an agent tool, and the guidance reads as a direct specification of what `surface.not_in_language` should do. ([anthropic.com/engineering/writing-tools-for-agents](https://www.anthropic.com/engineering/writing-tools-for-agents))
- **Unison — the codebase as a structured, queryable object.** "The Unison codebase is not just a mutable bag of text files, it's a structured object." UCM gives interactive discovery of what exists. Lang has the query substrate (`query`, `debug-map`, `interface`) but no *language-surface* query. `lang capabilities` is the missing vocabulary. ([unison-lang.org/docs/the-big-idea](https://www.unison-lang.org/docs/the-big-idea/), [codebase-editor-design.markdown](https://github.com/unisonweb/unison/blob/trunk/docs/codebase-editor-design.markdown))
- **LSP's capability handshake** is the protocol-level precedent for "state your surface before anyone asks." An LSP is explicitly Out of Scope in PROJECT.md, and should stay so — but the *initialize-capabilities* idea is free and applies to a CLI just as well.

---

## Adversarial pass: is DX work premature?

### The strongest case that it is

`STANDING-VERDICTS.md` contains this, already ratified: **"Don't build agent-loop/diagnostic support in parallel with semantic work. The defect taxonomy isn't knowable until the semantics exist; building blind produces the wrong cause-DAG edges."** That is the project's own process anti-pattern, and M002 confirmed it twice. DX-06's blame resolver is *built, tested, compile-time exhaustive, and structurally unreachable* — DX work done ahead of the semantics it needed. DX-07's `use_matching_argument` was *withdrawn empirically* because the single-type-per-function invariant makes the correct argument always the one already passed — DX work done ahead of the semantics it needed. **Two out of three Phase 13 criteria failed for the same reason, and that reason was a missing language feature.** D-13-02b's addendum records a *third* instance of the same root cause inside plan 13-06.

The deeper form of the argument: the language is ~5-10% of a writable language. No `if`, no loops, no arithmetic, no strings, no integers beyond `Byte`. You cannot write FizzBuzz. A hypothetical perfect diagnostic for `if` would say "`if` is not in this language" — which is *a true statement about a language nobody can use*, and improving how gracefully the toolchain refuses FizzBuzz does not move the project toward anyone writing FizzBuzz. Every hour on diagnostics is an hour not spent on the four PROJECT.md active items that all unblock each other: the type widening (which *automatically* closes DX-06 and DX-07), the emitter deletion (which is at its no-third-deferral limit under D-10-60), event identity (crossing a third milestone boundary), and control flow + arithmetic. And the type widening is *the* DX fix for Phase 13's partials, achieved entirely through semantic work. A DX phase in M003 risks building the next `resolveBlame`: correct, tested, and unreachable.

### Rebuttal, and where I concede

**I concede the general rule and most of the specific candidates.** More repair classes: premature — `use_matching_argument` is the direct evidence, and the defect taxonomy still is not knowable. LSP: explicitly Out of Scope, and correctly. REPL: there is nothing to evaluate interactively. Test/assertion construct: a language feature wearing DX clothes; it belongs in the semantics budget. Faster feedback: already sub-10ms with a ratified budget manifest; there is no problem. Richer `explain`: worth doing, but it is polish on a surface that is about to change, which is the second process anti-pattern ("don't Nyquist-validate a surface a milestone is about to change"). **On all of those, the adversary wins, and I am not recommending any of them.**

**The rebuttal is narrow and it is about category, not degree.** The standing verdict warns against building diagnostics *for semantics that do not exist yet*. Surface Honesty is the exact inverse: it is a diagnostic *about the boundary of what exists*, whose content is entirely derived from the current implementation and which is therefore **structurally incapable of being built blind.** `resolveBlame` was unreachable because it was written against a hypothesised future defect shape. `surface.not_in_language` fires on the very first hallucinated token, today, at 100% of the language's current boundary — I fired it by hand five times in this session. It cannot be premature in the way DX-06 was, because it has no dependency on any semantics that do not yet exist. It depends only on the lexer's token table, which is complete.

Three further points the adversary has to answer:

1. **The immaturity is the argument *for* it, not against.** A near-complete language rarely needs a "this is not in the language" diagnostic; hallucination is rare and self-evident. A 5%-complete language needs it desperately, because *nearly everything a model writes is off-surface.* This capability's value peaks now and declines monotonically as M003 adds features — which makes it the one DX item whose priority is strictly *decreasing over time.* Everything else on the candidate list can wait and lose nothing.
2. **It is the cheapest safety rail on M003's own semantic work.** M003 intends to add arithmetic and control flow. Each addition changes the surface. A derived, test-enforced manifest turns "what does the language support now" from a question requiring a `LANGUAGE-MATURITY.md` re-audit into a build artifact — which is a direct attack on the recurring failure that file exists to prevent. It pays for itself the first time someone asks "can we write a loop yet" and gets an answer from the compiler instead of from a wiki page.
3. **The constant-function spiral is a correctness defect in a shipped product API.** The Key Decisions table commits to treating "structured diagnostics and evidence as a versioned product API." A published API in which three different inputs return byte-identical responses including identical resource IDs is not a rough edge; it is a bug in the API contract. It would be worth fixing even if nobody ever wrote `if`.

**Net:** the adversary is right about ~80% of the candidate list and I am dropping those. The residual is one small, semantics-independent, decreasing-in-priority item that I would size at **one phase of 4-6 plans, sequenced after the type widening lands**, so the manifest is generated once against M003's widened surface rather than twice.

---

## Proposed M003 requirements this implies (draft REQ text, user-centric, testable)

**DX-08 — A machine-readable capability manifest.**
*As an AI agent about to write Lang, I can run `lang capabilities --json` and receive the complete, closed language surface — keywords, type constructors, operators, literal forms, control-flow forms, diagnostic codes, repair kinds, applicability levels, and an explicit `not_in_language` list — derived from the implementation and digest-versioned, so I do not have to guess what exists.*
Testable: `TestSurfaceManifestMatchesLexer` asserts the manifest's keyword set equals the lexer's token table; `TestSurfaceManifestMatchesTypeConstructors` asserts the constructor list equals `ability.go`'s dispatch arms; `TestEveryDiagnosticCodeIsInTheManifest` asserts set equality against a tree grep; every inline `examples[]` entry reaches `check pass`. Each test must fail in **both** directions (the `TestSeedEntryHazardIsReal` precedent).

**DX-09 — Off-surface tokens are named, not cascaded.**
*As an AI agent that wrote `if`, `for`, `+`, `return`, or a string literal, the first diagnostic I receive names the offending token, states that it is not in this language, points its primary span at that exact token, names the nearest real construct where one exists, and suppresses and counts downstream cascade diagnostics.*
Testable: one generated fixture per `not_in_language` entry; assert exactly one `surface.not_in_language` diagnostic whose `primary_span` covers the token's byte range; assert `cascade_ratio` ≤ 2 on every such fixture (measured today: up to 9).

**DX-10 — Distinct defects produce distinct feedback.**
*As an AI agent in a repair loop, editing my program in any way that changes its parse tree changes what `lang check` returns, so I can always tell whether my edit had an effect.*
Testable: `diagnostic_distinctness` = 1.0 over a held-out corpus of structurally distinct defective programs, using the existing structural-summary predicate from D-13-33 to certify the corpus members really are distinct. **This requirement currently fails**, with the three-iteration `if` spiral as the reproduction case — include it as the negative control.

**DX-11 — Diagnostics carry human coordinates and a calm second sentence.**
*As a human reviewing AI-written Lang in a CI log, every diagnostic shows line:column and the offending source line with the span underlined, and carries a plain-language `note` restating the machine message without jargon.*
Testable: assert every emitted diagnostic renders a `line:col` and a source excerpt; assert `notes` is non-empty for every code in a declared reviewability-critical subset; assert **zero diagnostic IDs changed** across the whole corpus by the addition (`Notes` must be excluded from identity, exactly as `Span`/`Replacement`/`Applicability` already are).

**DX-12 — `unrepairable` explains itself.**
*As an agent whose program `lang-repair` declined to fix, I receive the diagnosis code it saw, the applicability it found, and the reason it declined, never an empty field.*
Testable: assert `diagnosis` non-empty on every `status=unrepairable` result across the full defect corpus, including hallucinated programs. **Currently fails:** `{"status":"unrepairable","subprocess_count":1}`.

**DX-13 — Applied repairs leave a provenance record.**
*As a human auditing AI-written Lang, I can see which byte ranges were machine-repaired, by which repair kind, at which toolchain commit, bound to the pre- and post-edit source digests.*
Testable: applying a repair emits a provenance record; the record's digests bind to the actual bytes; `evidence --validate` refuses a mismatched record.

**Sequencing note.** DX-08/09/10/12 are one coherent phase and should be planned together, **after** the return-type-widening work, so the manifest is generated once against the widened surface. DX-11 and DX-13 are separable and lower priority; DX-11 in particular should not run concurrently with grammar changes, per the project's own "don't validate a surface a milestone is about to change" rule.

---

## Confidence + what would change my mind

**Confidence: HIGH** on the empirical findings. I built the toolchain from `908bac3`, ran `lang check`, `lang run`, `lang explain`, `lang query`, `lang debug-map`, `lang format`, and `lang-repair` against real fixtures and against nine hallucinated programs I wrote; every output block above is a verbatim paste. The spiral reproduction (three programs → identical IDs) is the strongest single piece of evidence and is reproducible in three commands.

**Confidence: MEDIUM** on the prior-art framing. The Elm, Roc, and Becker citations are secondary-source summaries; I could not fetch the Elm article body directly (the fetch returned no content), so the "one error at a time / show the source region" characterization rests on well-attested secondary description rather than my reading of the primary text. The rustc `Applicability` semantics are from the official rustc docs and are HIGH.

**Confidence: MEDIUM** on the sizing claim ("one phase, 4-6 plans"). The lexer and constructor tables are small and closed, which is why I believe it, but this repo's own history shows a new `OperationKind` costs 8-10 independent edits before it is real by the D-04-22 standard. A new *diagnostic code* should be far cheaper than a new operation kind, but I did not verify that by counting the edit sites for an existing code.

**What would change my mind:**

- **If the M003 charter turns out to add `if`, loops, and arithmetic all at once.** Then most of `not_in_language` empties in the same milestone, the hallucination surface shrinks dramatically, and the highest-leverage move becomes DX-11 (renderable diagnostics) instead, because the errors an agent hits would be real semantic errors rather than surface misses. This is the most likely thing to flip the verdict.
- **If someone runs a real agent loop and it does not spiral** — e.g. an LLM reliably recognises "expected `}`" as "this language has no `if`" from repo context alone. My spiral is a hand-simulated proxy, not a measured agent trajectory. A 20-trial measurement with a real model would settle it, and I would rank it as the single most valuable follow-up experiment. If recovery is already >80%, DX-09 drops to medium priority (DX-10 and DX-12 survive regardless, since both are API-contract defects).
- **If the diagnostic-ID collision turns out to be load-bearing for the cache** in a way I did not find. I read the identity computation and the design comments, and I believe enriching the *content* of identity preserves every determinism property — but I did not trace the cache-key derivation end to end.
- **If adding a `Notes` field turns out to move diagnostic IDs.** I read `Error` and `ErrorWithRepairs` and both compute identity over an explicit struct that a new field would not join. If some serializer elsewhere hashes the whole `Diagnostic`, DX-11's cost rises sharply and it should be deferred.
- **If `wiki/` already contains a derived grammar artifact I missed.** I read the README index and `LANGUAGE-MATURITY.md` but did not read all 40 wiki pages. If a machine-readable grammar already exists and is test-enforced, DX-08 shrinks to "expose it through the CLI," which is a one-plan item and strengthens the recommendation rather than weakening it.

---

## Sources

- [Elm — Compiler Errors for Humans](https://elm-lang.org/news/compiler-errors-for-humans) (MEDIUM — fetch returned no body; characterization from secondary sources)
- [Changelog #218 — Evan Czaplicki and Richard Feldman on Elm](https://changelog.com/podcast/218) (MEDIUM)
- [rustc_lint_defs::Applicability](https://doc.rust-lang.org/nightly/nightly-rustc/rustc_lint_defs/enum.Applicability.html) (HIGH — official rustc docs)
- [rust-clippy discussion #9994 — about Applicability](https://github.com/rust-lang/rust-clippy/discussions/9994) (MEDIUM)
- [rustfix#200 — cargo fix: improve handling of MaybeIncorrect suggestions](https://github.com/rust-lang/rustfix/issues/200) (MEDIUM)
- [cargo#13023 — automatically try applying non-MachineApplicable suggestions](https://github.com/rust-lang/cargo/issues/13023) (MEDIUM)
- [Becker et al. — Compiler Error Messages Considered Unhelpful (ITiCSE-WGR 2019)](https://dl.acm.org/doi/10.1145/3344429.3372508) (HIGH — peer-reviewed systematic review)
- [Denny, Becker et al. — On Designing Programming Error Messages for Novices: Readability and its Constituent Factors (CHI 2021)](https://dl.acm.org/doi/fullHtml/10.1145/3411764.3445696) (HIGH — peer-reviewed)
- [Barik et al. / Becker — On Novices' Interaction with Compiler Error Messages (ICER 2017)](https://dl.acm.org/doi/abs/10.1145/3105726.3106169) (HIGH)
- [Roc — language site](https://www.roc-lang.org/) and [FAQ](https://www.roc-lang.org/faq) (MEDIUM)
- [Anthropic — Writing effective tools for AI agents](https://www.anthropic.com/engineering/writing-tools-for-agents) (HIGH — primary vendor guidance)
- [Anthropic — Effective context engineering for AI agents](https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents) (HIGH)
- [Unison — The big idea](https://www.unison-lang.org/docs/the-big-idea/) and [codebase-editor-design](https://github.com/unisonweb/unison/blob/trunk/docs/codebase-editor-design.markdown) (MEDIUM)

**Repository sources (HIGH — read and executed this session, at `908bac3`):**
`internal/compiler/diagnostic/diagnostic.go` · `internal/compiler/check/check.go` · `internal/compiler/ability/ability.go:120-168` · `internal/compiler/syntax/token.go` · `internal/compiler/session/session_phase6_query.go` · `internal/compiler/session/session_phase6_budget.go` · `internal/compiler/session/qlt02_budget_manifest.json` · `internal/compiler/session/risk_lanes.json` · `internal/compiler/measure/statistics.go` · `internal/compiler/protocol/protocol.go` · `cmd/lang` · `cmd/lang-repair` · `testdata/phase07/call_basic.lang` · `testdata/phase13/derivation_interprocedural_loan_defect.lang` · `.planning/PROJECT.md` · `.planning/LANGUAGE-MATURITY.md` · `.planning/STANDING-VERDICTS.md` · `.planning/milestones/M002-MILESTONE-AUDIT.md` · `.planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/PHASE-13-DEBT.md` · `AGENTS.md` · `README.md`
