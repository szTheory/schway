# Phase 35: Bitwise Operators, Literal Shifts, and Hex Literals - Research

**Researched:** 2026-10-13  
**Domain:** Go compiler frontend, U64 scalar IR, C17 code generation, deterministic compiler evidence  
**Confidence:** HIGH for source architecture and locked scope; MEDIUM for proposed external-language guidance

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### Scope and compatibility

- **D-01:** Use Go 1.24 standard-library/local implementation and the existing C17/Clang backend. No expression parser library, bigint package, new proof/replay framework or dependency. Keep legacy core/execution/diagnostic schema identifiers and current exit classes; new op kinds reuse existing fields rather than adding a mandatory wire field.
- **D-02:** Before compiler changes, commit an enumerated literal generated-C baseline for the current tracked source corpus through the relevant existing emitter routes. Record non-emitting cases explicitly. Keep all old goldens immutable and arm a reached byte-drift control. Programs without the new operators, grouping or hex extension retain their emitted C bytes; Adler's deliberately changed source is an explicit separate exception with unchanged behavioral answers. Pin previously admitted scalar/linear numeric policies and Bool identifier uses before conversion/parser changes.

#### Expression grammar and bounded grouping

- **D-03:** Use descending precedence `% > + > << >> > & > ^ > | > (<, ==)`, left association within each tier. Preserve the existing shared comparison tier and relative order of old operators. U64-only bitwise operators return U64; Bool operands remain refused. Evaluate ordinary operands in the existing left-to-right lowering order. Keep `&&`, `||`, `!` and subtraction refused; admitting `|` must not split `||` into admitted operations.
- **D-04:** Add transparent scalar parentheses for the exact roadmap witness. Parentheses preserve the inner expression kind rather than creating a runtime/core operation: `(3)` and nested groups around a literal are literal amounts; `(1 + 2)` and names remain non-literals even if constant propagation knows their value. Preserve lossless CST punctuation and source spans. Do not infer that grouping admits calls, tuples, Bool literal tails or other absent forms.
- **D-05:** Bound recursive scalar expression nesting at 64, with a distinct named expression-depth resource refusal and `parser.expression_depth` manifest entry. Cover grouping and recursively nested builtin indices; use iterative delimiter-aware recovery and preserve sibling diagnostics. Reuse a numeric limit value if appropriate, but do not label expression depth as type depth. Remove or bound the existing unsupported-`!` recursive path when touching the atom parser. This is an explicit parser resource-bound amendment, not a claim that source/token caps alone prevent stack exhaustion. Characterize old nested-index admission before implementation and record any resource-refusal delta separately; do not silently rebaseline old successful C witnesses.
- **D-06:** Preserve nested type closing tokens, comparisons, ownership keywords, declaration/match `|` separators, comments and strings. Planner may choose context-local adjacent-angle recognition or dedicated shift tokens with contextual type-close splitting; require contiguous source bytes for `<<`/`>>` and a nested-generic parse/format golden. Formatter output must reparse to the same admitted tree, be idempotent, keep numeric spelling, and place infix operators coherently without changing old formatting goldens.
- **D-07:** Under D-03 the written witness means `x ^ ((y << 3) & 0xFF)`. Pin `x=0x100`, `y=0x21` to literal decimal answer `264`; `(x ^ (y << 3)) & 0xFF` gives `8` and is a discriminating control. Document both rather than silently treating the original expression as a final mask. Use a no-input `main` so both public `schway run --engine=interpreter|native` routes execute it within the current CLI boundary.

#### U64 semantics, shift representation and refusals

- **D-08:** XOR/AND/OR are ordinary per-bit operations on exactly 64-bit unsigned values. `x << n` keeps the low 64 bits of `x * 2^n`; `x >> n` is logical right shift. There is no overflow defect, signed shift, masked count, target-dependent word width or dynamic amount. Counts 0, 31, 32 and 63 have independently authored expected vectors; include maximum/high-bit/all-zero values and differing operators/associations, not only engine agreement.
- **D-09:** Lower shifts to distinct immediate-only core kinds, recommended `OpShiftLeftU64` and `OpShiftRightU64`, using the existing `ConstU64` string for the canonical base-10 amount. `RightID` must be empty. Permit `ConstU64` only on `OpConst` and these shift kinds; require presence, canonical decimal spelling, range 0..63, U64 source/result, valid initialized source and fresh/valid target facts. No named amount place, constant-def lookup or producer-issued literal witness may substitute for this contract. Old core serialization/IDs stay unchanged when no new feature appears.
- **D-10:** Source checking independently requires a number leaf after transparent grouping; refuse names, computed constants, unsupported negative counts and representable 64+ counts with `surface.unsupported_shift_amount` and the one-line hint: `Shift amounts must be integer literals from 0 to 63; write a literal in that range.` Decimal and admitted hex amount spellings canonicalize only in core. Malformed numeric tokens and unrepresentable U64 values keep their existing named lexical/type failures, with an explicit code/span delta ledger if a targeted hex diagnostic is refined. Do not admit unary minus or subtraction to diagnose shifts.

- **D-11:** Corevalidate and originvalidate each independently enforce the immediate shape, U64 types and range, including missing/noncanonical/oversized amounts, extraneous `RightID`, wrong kind fields and malformed raw-core producer input. Their shift contract refusal uses the registered code/hint where diagnostics are available. Removing each refusal must be reached and caught in a table-driven control. Neither peer has source AST provenance: they prove valid IR immediate semantics, not that arbitrary raw core was generated from a particular literal token. Source syntax admission belongs to check; a new source-provenance protocol is out of scope.
- **D-12:** Generated shifts must contain a checked literal amount below 64. Use an unsigned non-promoted left type, e.g. `(uint64_t)((unsigned long long)lhs << 16)` and analogous right shift, so C integer promotion cannot introduce signed-shift UB. `uint64_t` is exact-width when available; unsigned long long is at least 64 bits, and the final cast yields the required low 64 bits even on a wider implementation. No unchecked fallback, compiler assumption or new optimizer attribute. A finite generated-C corpus scan must identify/count the actual new shift sites and fail on a dynamic or 64+ shift mutant; do not pass merely because no shift was found.

#### Hex spelling and existing numeric policies

- **D-13:** Hex is partially implemented already: lexer accepts `0x`/`0X`, mixed-case ASCII digits, and existing between-digit underscores; generic number AST/raw token and formatter preserve spelling. Complete uppercase `0X` conversion on the legacy linear path with the smallest local change. Keep leading zeros bounded by existing source/token limits and overflow by value, not digit count. No new hex token, suffix, alternate radix, separator placement, Unicode digit normalization or automatic uppercase/underscore rewrite.
- **D-14:** Preserve existing decimal/binary policies rather than normalizing them as collateral work. Legacy `parseU64Literal` treats `010` as decimal 10 and recognizes lowercase binary prefix; scalar base-0/fallback conversion treats `010` as 8 and accepts `0B10` as 2, while legacy rejects `0B10`. Record and test these route differences as deferred consistency debt. A shared converter is allowed only if it preserves explicit route policies and earns its complexity; the preferred patch is narrow `0X` support. The no-feature C contract forbids changing `010` now.
- **D-15:** Malformed numeric candidates retain `syntax.malformed_numeric_literal` and the complete raw candidate span (`0x`, invalid radix digits, trailing ASCII/Unicode suffix, bad separators). Use exact token/span ownership to avoid a second parser expectation for that same cause; retain other causes and limit diagnostics. Valid-but-overflowing hex keeps a named existing type/range code: legacy `check.literal_out_of_range`, scalar `type.invalid_u64` unless a targeted documented hex-only refinement is needed. Do not silently promise a uniform historical code. Pin code/count/span and spelling-preserving format/Tree.Bytes() for accepted and rejected inputs, including EOF, maximum U64, maximum+1 and long leading zeros.

#### Independent evidence and manifest

- **D-16:** Add independently implemented transfer cases in check, corevalidate and originvalidate, including their actual scalar fixed-point trigger/inventory paths. Keep each peer's type/shape/range admission independent of the producer and of a shared semantics evaluator. For exact-result transfer mutants, feed discriminating literal results into a downstream observable analysis obligation (e.g. statically zero remainder) or independently test that peer's transfer result; do not claim an isolated returned value was checked if the peer only analyzes safety. Each table row names the peer, op, seeded fault, reached branch and named failure; restore every mutable control.
- **D-17:** Keep pure operations event-free in interpreter, C evidence projection, causal peer and capacity inventories, with no new runtime reason or event. `executionpeer.Validate` explicitly checks observed causal structure and receives no entry argument; it cannot prove the numeric result of a pure operation. Independently test per-op causal admission/no-event behavior and reject fabricated step/defect events at those operations. Retain literal expected-answer and wrong-result controls in interpreter/native, at `-O0`, `-O3` and `-O3 -flto`, with a focused UBSan shift lane. New value replay or operand events would exceed this slice and defeat its event-free contract.
- **D-18:** **Evidence-backed acceptance clarification:** BIT-01/roadmap criterion 1 currently require a wrong-result mutant for every op at every peer, including the causal-only execution peer. Before writing plans, amend the live Phase 35 criterion and BIT-01 wording with a dated, source-grounded note: value-transfer mutants at check/core/origin and engine answer mutants; per-op event-free/forged-event controls at executionpeer. Preserve the original demand in the dated amendment/provenance, name this limitation in verification, and do not count causal admission as numeric validation. This recommendation is auto-followed under the user's explicit decision delegation; it preserves the goal of defined, independently tested answers without introducing runtime events. No other requirement or phase owner changes.
- **D-19:** Extend the existing deterministic `schway.surface/1` manifest with all five operators, grouping, precedence/association, literal-only count/range, low-bit left-shift rule, hex spelling and new expression-depth bound. Add the implemented shift refusal to its catalog. Keep ten top-level fields/source-free discovery and independent literal complete-source witness matrix plus finite source/admission/limit census; no producer-fed expected manifest or weakened discovery tests. Arm removal, false-admission, missing refusal/hint and weakened restriction controls, with restoration. Keep scalar expression calls absent until Phase 36.

#### Runnable consumer, lanes and phase cost

- **D-20:** Replace Adler's 16 `r = r + r` operations with the equivalent `b << 16`, retaining the existing final addition and module/cleanup contract. The maximum bounded `b` makes this equivalent to the old checked additions. Existing Phase 29/31 literal answer, checksum and `adler_low` output tests must pass unchanged, on both engines. Do not update expected answers to the new producer or require file input through the public interpreter CLI before Phase 37. Include a reached wrong-shift mutant in a focused additional test.
- **D-21:** Only focused named local selectors/narrow CLI checks with `GOCACHE=/tmp/ai-lang-verification-gocache`; no local project-wide suite, whole package runs, broad native matrix, push or hosted dispatch/binding. Plan the small optimization/literal-shift matrix as named feature tests, reusing existing runners. Existing recurring CI owns broad checks; Phase 38 owns hosted two-host final codedigest receipts. State local host/toolchain and incomplete host lanes honestly; human readability and cold/warm whole-loop latency remain unmeasured.
- **D-22:** Aim for about eight coherent plans or fewer. Baseline before compiler edits, then grammar/hex and core contracts, independently derived validators/execution, focused evidence/manifest, consumer/guide/closeout. Additional plans require a named runnable witness or safety obligation; do not split every role or op into a plan. SUMMARY coverage names real tests and the closed kind vocabulary; prioritize deterministic automation over human UAT.
- **D-23:** Retain dormant seeds z1g/trz/z1s with the scoped transition disposition below. Do not promote a database, telemetry framework or syntax study from generic forecasts. Update living documents only with newly inspected facts and preserve archived/differently dated claims; the three useful next capabilities remain Phase 35 computation, Phase 36 imported helpers and Phase 37 file CRC/second consumer/CLI parity.

### Planner's Discretion

- Exact local helper, op-name and test-file layout, consistent with the above contracts.
- Shift-token/type-close strategy and deterministic formatter spacing once pinned.
- Named expression-depth code/precise boundary counting, with 64-level bound and reachable malformed controls.
- Concrete finite baseline corpus, per-peer observability sinks and focused matrix arrangement; preserve historical receipts and goldens.

### Deferred Ideas (OUT OF SCOPE)

- Numeric route unification (`010`, `0B10`) requires a separately owned semantic compatibility decision; preserve the discrepancy as named debt now.
- Scalar result replay within executionpeer needs runtime inputs/CFG/value evidence and is not part of this event-free slice.
- Dynamic shifts, complement, rotations, numeric suffixes/new bases/types, hex output, arity-N, aggregates, writable buffers and additional backend remain outside Phase 35.
- Scalar helpers Phase 36; CRC consumer/second importer/interpreter file CLI/guide Phase 37; final two-host hosted binding Phase 38.
- Word counter/longest_line remain future M008, followed by a named bounded config parser/JSON consumer.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| BIT-01 | U64 `^`, `&` and `|` are admitted in scalar code. Interpreter and native agree with literal expected answers at -O0, -O3 and -O3 with LTO. Each peer (check, corevalidate, originvalidate, executionpeer) derives the ops independently, and a wrong-result mutant per op is caught. [VERIFIED: `.planning/REQUIREMENTS.md:35`] | Extend Pratt parse/lowering and all-operation inventory; separate value mutants (check/corevalidate/originvalidate) from causal event controls (executionpeer), following D-18. [VERIFIED: `internal/compiler/syntax/parser.go:749-795`; `internal/compiler/core/core.go:877-885`] [CITED: `35-CONTEXT.md D-16..D-18`] |
| BIT-02 | U64 `<<` and `>>` are admitted only when the shift amount is an integer literal from 0 to 63. Amounts 0, 31, 32 and 63 give pinned answers on both engines and at every optimization level. An amount of 64 or more, or a non-literal amount, is refused at check, core and origin with a named code. Generated C never contains an unguarded shift. `<<` drops high bits by definition, which the manifest and guide document. [VERIFIED: `.planning/REQUIREMENTS.md:36-39`] | Source literal-leaf check; immediate-only core field contract, independent validation triggers/transfers, unsigned C emission, event-free projection/capacity checks, reached rejection mutants and corpus shift scan. [CITED: `35-CONTEXT.md D-09..D-12, D-17, D-19`] |
| BIT-03 | Hex integer literals (`0x…`, case-insensitive digits) are admitted wherever decimal U64 literals are. Overflowing or malformed hex is refused with a named code, and the formatter keeps the source spelling. [VERIFIED: `.planning/REQUIREMENTS.md:40`] | Preserve lexer ownership and token text; narrowly add `0X` to legacy conversion, pin both existing conversion policies and complete malformed token/span/format cases. [VERIFIED: `internal/compiler/syntax/lexer.go:54-74`; `internal/compiler/check/check.go:6809-6818`; `internal/compiler/syntax/tree.go:5-19`] [CITED: `35-CONTEXT.md D-13..D-15`] |
| BIT-04 | `examples/checksum/adler32.schway` uses `b << 16` in place of 16 doublings. All of its pinned answers, and the checksum and `adler_low` outputs, stay byte-identical on both engines. [VERIFIED: `.planning/REQUIREMENTS.md:41`] | Use existing module/application runners and Phase 29/31 answer witnesses; add an actually reached wrong-shift control. [CITED: `35-CONTEXT.md D-20`] |
</phase_requirements>

## Summary

This phase is a coordinated syntax-to-evidence change, not a parser-only extension. The current Pratt parser has three scalar precedence levels, the AST has no group node, the lexer treats `<` and `>` as separate tokens, and the formatter spaces those tokens by context. Scalar operations also pass through independent admission/transfer triggers, interpreter/C dispatch, event projection, kind registries, and the Phase 34 surface manifest. Plan those seams together and keep pure integer operators event-free. [VERIFIED: `internal/compiler/syntax/parser.go:749-823`; `internal/compiler/ast/ast.go:222-227`; `internal/compiler/syntax/lexer.go:240-275`; `internal/compiler/syntax/format.go:55-58`; `internal/compiler/check/scalar.go:153-317`; `internal/compiler/core/core.go:877-885`; `internal/compiler/corevalidate/corevalidate.go:4722-4949`; `internal/compiler/originvalidate/originvalidate.go:1330-1449`; `internal/compiler/interp/interp.go:1713-1788`; `internal/compiler/cgen/cgen_program.go:173-278`; `internal/compiler/surfacemanifest/manifest.go:1-34`]

There is one acceptance inconsistency to resolve before plans: the current BIT-01 and roadmap criterion require a wrong-result mutant at `executionpeer`, while that peer only validates observed causal structure and receives no program entry/value result. D-18 already directs the evidence-backed amendment: value-transfer mutants at check/core/origin plus interpreter/native answer controls, and per-operation event-free/fabricated-event controls at the execution peer. The current public runnable witness also conflicts with the existing entry contracts: `run` requires a parameter, but binds only Byte, Buffer and Match input cases, while scalar checking accepts U64 or ByteView parameters. D-07’s no-input witness therefore needs an explicitly scoped entry decision or a replacement witness before claiming both public CLI routes are runnable. [VERIFIED: `internal/compiler/executionpeer/executionpeer.go:70-81`; `internal/compiler/session/session.go:1224-1255`; `internal/compiler/check/scalar.go:37-50`] [CITED: `35-CONTEXT.md D-07, D-18`]

**Primary recommendation:** Start with a reached, immutable old-source generated-C corpus baseline across existing emitters and pin numeric route differences; then implement transparent bounded grouping, context-safe shift parsing/formatting, and narrow uppercase-hex conversion. Add immediate-only shift IR fields and independently update every admission/transfer/event/schema consumer before adding focused interpreter/native vectors and Adler’s `b << 16` witness.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Operator tokens, precedence, grouping, recovery and formatting | API / Backend (compiler frontend) | — | The compiler frontend owns source interpretation and canonical formatting; retain token/CST spelling and spans. [VERIFIED: `internal/compiler/syntax/tree.go:5-19`; `internal/compiler/syntax/format.go:15-53`] |
| Literal and shift admission | API / Backend (checker) | Core validation | Source checker enforces literal provenance; independent IR validators enforce operation shape, types and immediate range. [CITED: `35-CONTEXT.md D-09..D-11`] |
| U64 computation | API / Backend (compiler IR/runtime) | Native C backend | Interpreter and C generation implement the same exact-width U64 meaning. [CITED: `35-CONTEXT.md D-08, D-12`] |
| Evidence and public surface | API / Backend | CLI/session | The manifest and session evidence expose accepted syntax, limits and engine answers without introducing runtime events. [VERIFIED: `internal/compiler/surfacemanifest/manifest.go:1-34`] [CITED: `35-CONTEXT.md D-17, D-19`] |
| Adler consumer | API / Backend | Interpreter and native backend | Existing source module supplies a real runnable consumer and preservation witness. [VERIFIED: `.planning/REQUIREMENTS.md:41`] [CITED: `35-CONTEXT.md D-20`] |

## Project Constraints (from AGENTS.md)

- Bootstrap and backend: `- **Bootstrap**: Go 1.24 standard library hosts Stage 0; readable C17 emitted to installed Clang is the development native path.` [VERIFIED: `AGENTS.md:35-36`]
- Dependencies: `- **Dependencies**: Prefer standard-library-only and shallow audited boundaries; every dependency must earn more than copy-local implementation.` [VERIFIED: `AGENTS.md:37`]
- Source and evidence: `- **Source authority**: Text remains Git-friendly and formatter-owned; parsing is lossless enough to preserve comments and produce stable recovery.` `- **Evidence**: Deterministic fixtures, negative controls, properties, mutation, differential execution, and sanitizers answer distinct questions and have explicit cost lanes.` [VERIFIED: `AGENTS.md:38-41`]
- Portability and security: `- **Portability**: macOS and Linux are the initial host priorities; wire contracts and target facts must not accidentally encode the current Apple arm64 host.` `- **Security**: Untrusted input, unsafe operations, FFI, secrets, and build authority stay explicit; no generic untaint or ambient authority escape.` [VERIFIED: `AGENTS.md:42-45`]
- Feedback and correctness: `- **Feedback latency**: Optimize for the whole generate → verify → run → observe → repair loop; report cold and warm distributions rather than one flattering number.` `- **Correctness**: Safe code has defined behavior; optimizer attributes and FFI guarantees must derive from checked facts.` [VERIFIED: `AGENTS.md:26-30`]
- Low-level/runtime scope: `- **Low-level capability**: Usable ownership, borrowing, deterministic cleanup, target layout, and C interop are milestone requirements, not post-v1 add-ons.` `- **Runtime posture**: No mandatory global tracing heap or service runtime; the first native slice uses stack/inline values and explicit resources.` [VERIFIED: `AGENTS.md:31-34`]
- Naming: Schway is settled as the public project/language name; this feature phase has no rename work. [VERIFIED: `AGENTS.md:21-22`]
- Existing patterns are authoritative because stack, conventions and architecture are not yet mapped: `Technology stack not yet documented.`, `Conventions not yet established.`, `Architecture not yet mapped. Follow existing patterns found in the codebase.` [VERIFIED: `AGENTS.md:54`; `61`; `68`]
- GSD workflow: start file-changing work through the applicable GSD command; phase work belongs to `$gsd-execute-phase`, while small fixes/docs use `$gsd-quick`. Planning has already started for this research and this agent writes only the requested research artifact. [VERIFIED: `AGENTS.md:80-90`]
- At workflow transitions, compare the living roadmap/maturity records with source witnesses and evidence, disposition only relevant seeds, preserve historical receipts, and surface three next capabilities. This transition review found Phase 35 computation, Phase 36 imported scalar helpers, and Phase 37 file CRC/second consumer/CLI parity as the ranked next capabilities; details appear below. [VERIFIED: `AGENTS.md:101-130`] [CITED: `35-CONTEXT.md`]
- Project skill discovery found no applicable project-local skill; `No project skills found.` [VERIFIED: `AGENTS.md:75`]

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go standard library | 1.24 (workspace `go.mod` toolchain) | Lexer/parser/checker, core validation, interpreter, session tests | Existing bootstrap/runtime; no new dependency under D-01. [VERIFIED: `go.mod:3`] [CITED: `35-CONTEXT.md D-01`] |
| C17 through installed Clang | C17; local Clang 21.0.0 | Generated native code and optimization/sanitizer evidence | Existing development backend and phase contract. Local tool version was observed; D-12 supplies the unsigned shift shape. [VERIFIED: `AGENTS.md:35-36`] [CITED: `35-CONTEXT.md D-12`] |

### Supporting

| Tool | Version | Purpose | When to Use |
|------|---------|---------|-------------|
| Go `strconv` | Go 1.24 standard library | Preserve existing U64 parsing behavior where the phase changes prefix handling | Keep conversion route policies explicit; do not centralize them unless compatibility remains exact. [CITED: https://go.dev/pkg/strconv/?m=old] [VERIFIED: `internal/compiler/check/check.go:6809-6818`] |
| Clang UndefinedBehaviorSanitizer | local Apple Clang 21.0.0 | Focused validation for generated shifts | Run the dedicated shift test/lane; UBSan shift checks are supplementary to semantic vectors and static literal-count scan. [CITED: https://clang.llvm.org/docs/UndefinedBehaviorSanitizer.html#available-checks] |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Existing Pratt parser plus local grouping/pair recognition | External/generated expression parser | Adds dependency and risks incompatible source recovery/precedence; outside D-01 and unnecessary for this bounded grammar. [CITED: 35-CONTEXT.md D-01, D-03..D-06] |
| Context-local adjacent-angle recognition | Dedicated `<<`/`>>` token plus contextual splitting of nested generic closers | Dedicated tokens can simplify operator parsing but increase lexer/parser/formatter type-close work; both options are permitted by D-06 and must preserve identical source/format controls. [CITED: 35-CONTEXT.md D-06] |
| Route-local numeric prefix conversion | Shared numeric converter | Can reduce duplication, but may erase established `010`/`0B10` route behavior; only use if both policies stay explicit and compatibility tests prove unchanged. [CITED: 35-CONTEXT.md D-13..D-15] |

No external package installation is needed; therefore a Package Legitimacy Audit is not applicable.

**Installation:** None.

**Version verification:** `go version go1.24.0 darwin/arm64`; `Apple clang version 21.0.0 (clang-2100.1.1.101)`. These are observed local versions; Linux/hosted toolchains are not verified here.

## Architecture Patterns

### System Architecture Diagram

```text
Schway source bytes
  -> lossless lexer (one malformed-numeric diagnostic owns a whole candidate)
  -> parser (adjacent-angle operator recognition, bounded scalar grouping/index recursion)
  -> formatter/tree round-trip and surface manifest (leaf consumers)
  -> checker (source literal-only shift rule; U64 expression lowering)
  -> core IR (shift count in kind-exclusive canonical ConstU64; no amount place)
  -> independent corevalidate and originvalidate admission/transfer
  -> interpreter -------------------------------> decimal/output observation
  -> C generator -> C17/Clang at O0/O3/LTO ----> native decimal/output observation
                     | event-free projection/capacity census
  -> session evidence: authored vectors, wrong-answer controls, peer controls,
     no-feature C baseline, corpus literal-shift scan, and manifest witnesses
```

### Recommended Project Structure

Keep changes in existing owner packages; no new package or dependency is warranted.

```text
internal/compiler/syntax/         # token recognition, bounded parser, lossless formatting
internal/compiler/check/          # source admission and producer lowering
internal/compiler/core/           # operation kind registration and immediate field contract
internal/compiler/corevalidate/   # independent shape/type/range and transfer
internal/compiler/originvalidate/ # independent origin admission/transfer triggers
internal/compiler/interp/         # U64 runtime semantics; event-free
internal/compiler/cgen/           # unsigned shift emission and event/capacity projection
internal/compiler/surface/        # named shift and expression-depth refusals
internal/compiler/surfacemanifest/ # installed source-free contract
internal/compiler/session/        # baseline, independent witnesses, native/interpreter controls
```

### Pattern 1: Context-local adjacent shift pairs over lossless angle tokens

**What:** Preserve `<` and `>` as current single-character tokens, and recognize `<<` / `>>` only when adjacent token spans have contiguous source offsets and parser is in scalar-expression context. This avoids globally changing nested generic closers or comparison tokens. Existing token declarations contain `TokenLAngle Kind = "<"` and `TokenRAngle Kind = ">"` [VERIFIED: `internal/compiler/syntax/token.go:43-51`]; lexer multi-character handling currently covers `==`, `->`, `=>`, while angle punctuation is otherwise split [VERIFIED: `internal/compiler/syntax/lexer.go:240-275`].

**When to use:** Recommended for this phase because type references recursively consume one closer, formatting has scalar/type-specific angle spacing, and the lexer is lossless. A dedicated shift token strategy is also allowed by D-06, but requires explicit contextual type-close splitting and the same nested-generic golden.

**Example:** Add an operator-pair peek/consume helper that checks both `Kind` and `left.Span.End == right.Span.Start`; use it in scalar Pratt parsing only. Treat a noncontiguous `x < <y` as two operators/recovery input, never as a shift. Update the formatter to emit each shift pair with one coherent infix spacing policy, and verify nested generic close formatting remains byte-stable. Do not consume `||` as two admitted bitwise pipes: its current refusal specifically recognizes contiguous pipe spans [VERIFIED: `internal/compiler/syntax/parser.go:798-823`].

### Pattern 2: Bounded transparent scalar groups

**What:** Parse `(` expression `)` as a transparent AST wrapper or by returning the same semantic expression with a span including the delimiters; retain parentheses in the lossless token tree. Grouping must preserve whether the inner node is a number leaf. The current expression AST fields are `Kind, Name, Value, Left, Right, Span` and contain no group kind [VERIFIED: `internal/compiler/ast/ast.go:222-227`].

**When to use:** For the roadmap witness and literal-only shifts. Parenthesized number leaves such as `(3)` remain literal; a group around `1 + 2` remains a computed expression. Do not use constant-folded values to satisfy literal provenance.

**Depth:** `parser` currently has `maxTypeDepth = 64` and `maxTypeNodes = 4096` [VERIFIED: `internal/compiler/syntax/parser.go:52-92`]. Add a separately named scalar-expression bound and manifest entry for groups, nested builtin indices, and any existing recursive unsupported unary path. Recover with the existing iterative delimiter-aware routine so a deep malformed expression does not prevent sibling diagnostics [VERIFIED: `internal/compiler/syntax/parser.go:1410-1459`]. Characterize old nested-index success/failure before treating new resource refusals as a baseline delta.

### Pattern 3: Source-specific number parsing stays explicit

**What:** Keep the legacy linear parser and scalar checker conversions separate unless a shared helper can explicitly preserve both route policies. The legacy `parseU64Literal` source currently says:

```go
digits := strings.ReplaceAll(text, "_", "")
base := 10
if strings.HasPrefix(digits, "0x") {
    base, digits = 16, digits[2:]
} else if strings.HasPrefix(digits, "0b") {
    base, digits = 2, digits[2:]
}
return strconv.ParseUint(digits, base, 64)
```

[VERIFIED: `internal/compiler/check/check.go:6809-6818`]

Scalar lowering instead attempts `strconv.ParseUint(text, 0, 64)` and then decimal fallback, while preserving its current AST-literal `0` remainder check [VERIFIED: `internal/compiler/check/scalar.go:608-668`]. Preserve the existing observed policy split: legacy `010` is decimal 10 and legacy only checks lowercase `0b`; scalar base-0 treats `010` as octal 8 and accepts `0B10` as 2. Pin both before edits; only the narrow uppercase `0X` legacy conversion is in scope under D-13/D-14. [CITED: 35-CONTEXT.md D-13..D-15; `strconv.ParseUint` base-zero behavior: https://go.dev/pkg/strconv/?m=old]

**Malformed ownership:** The lexer consumes a full numeric candidate and emits `syntax.malformed_numeric_literal` over its span; parser expectations should not duplicate that same root cause. Keep accepted raw token spelling through AST/tree formatter paths [VERIFIED: `internal/compiler/syntax/lexer.go:54-74`; `internal/compiler/syntax/tree.go:5-19`; `internal/compiler/syntax/format.go:261-268`].

### Pattern 4: Literal shift count uses an existing kind-exclusive IR field

**What:** Keep the shift amount out of `RightID` and out of a constant-def lookup. D-09 recommends two distinct immediate op kinds and the existing decimal string field. Current `OperationKind` definitions include `OpConst OperationKind = "const"`, `OpAddChecked OperationKind = "add_checked"`, `OpLessU64 OperationKind = "less_u64"`, `OpEqualU64 OperationKind = "equal_u64"`, and `OpRemainderU64 OperationKind = "remainder_u64"` [VERIFIED: `internal/compiler/core/core.go:706-731`]. The registered kind literal is enumerated in `AllOperationKinds()` [VERIFIED: `internal/compiler/core/core.go:877-885`].

The source comment on the existing field states: `// ConstU64 is Phase 19's kind-exclusive canonical U64 value, represented as base-10 digits (including "0"). It is populated only on OpConst; all other operation kinds leave it empty so legacy core JSON is stable.` [VERIFIED: `internal/compiler/core/core.go:969-972`]. Amend this exclusivity intentionally: only `OpConst` and the two immediate shift kinds may carry the field; retain `omitempty` so old core JSON and schema identifiers do not change when shifts are absent. Core kinds are a closed inventory, not just an enum: `TestAllOperationKindsRegistered` and `TestAllOperationKindsHandledAtEverySite` exercise declared operations across checker, validators and engines [VERIFIED: `internal/compiler/core/core_test.go:1055-1072`].

### Pattern 5: Update every independent admission, transfer trigger, and projection

Plan the following distinct code paths, rather than assuming one generic dispatcher covers all consumers:

| Owner | Current seam to extend | Required phase behavior |
|-------|-------------------------|-------------------------|
| `check` | `scalarBinaryOperator`, left-to-right lowering, literal-shift source check | Derive U64 result; require a number leaf after transparent groups for shift count; canonicalize only after successful source parsing. |
| `core` | `AllOperationKinds`, `Operation` fields | Register both kinds; specify `ConstU64`/`RightID` exclusivity and stable `omitempty` serialization. |
| `corevalidate` | operation shape/type switch; field exclusivity; block replay, straight-line transfer, scalar-CFG trigger/fixed point | Independently reject malformed immediate values/shape/types and compute U64 results in each relevant transfer route. Existing critical seams are `validateScalarFixedPoint` and `validateStraightLineScalarFacts` [VERIFIED: `internal/compiler/corevalidate/corevalidate.go:4722-4850`; `4892-4949`]. |
| `originvalidate` | independent `originTransferScalarOperation`, `originHasScalarCFG`, `RecomputeOriginPerReturn`, and explicit arithmetic admission trigger | Add independent shape/range/type checks and transfer cases; ensure origin recomputation and published problems actually reach them. `PublishProblemsFor` currently gates scalar analysis on the presence of a remainder operation, which must not be the only route that discovers new pure ops [VERIFIED: `internal/compiler/originvalidate/originvalidate.go:198-212`; `293-300`; `418-509`; `1330-1449`]. |
| `interp` | operation dispatch/value switch | Implement `^`, `&`, `|`, unsigned logical shifts; malformed IR fails closed; do not add defects/events/reasons. [VERIFIED: `internal/compiler/interp/interp.go:1713-1788`] |
| `pathoracle` | exhaustive operation ancestry switch; `OpConst` special fresh-root case | Register both new kinds without adding loan/permission semantics; shifts are U64 value operations and must stay loan-free. [VERIFIED: `internal/compiler/pathoracle/pathoracle.go:1044-1081`] |
| `cgen` | exported emitter admission gates; `needsU64` function signature inference; `hasScalarOperations` path selection; scalar operation dispatch; branch emitter; pure event projection and capacity census | `Emit`, `EmitNative` and `EmitApplication` validate at entry; ensure shift programs select scalar emission, signatures carry required U64 values, branch dispatch stays separate/fail-closed, and event projection remains empty. [VERIFIED: `internal/compiler/cgen/cgen.go:111-137`; `internal/compiler/cgen/cgen_program.go:173-278`; `867-899`; `2192-2196`; `2382-2385`; `2540-2814`] |
| `executionpeer` | causal operation kind admission and event classification | No numeric-result check; add each op to no-event admission and reject fabricated step/defect events. |
| `surfacemanifest` | versioned source-free installed facts | Extend operators, restrictions, limits and refusal catalog; keep the ten existing JSON fields and independent evidence census. |
| `surface` / CLI | registered named refusal plus check path; public run route | Add `surface.unsupported_shift_amount` and expression-depth refusal; resolve the no-input witness/entry mismatch before making the CLI runnable gain a task. |

For shifts encode the immediate in `ConstU64` as canonical base-10 `0..63`, require empty `RightID`, U64 initialized left input and fresh target. Corevalidate and originvalidate must each reject missing, empty/noncanonical, out-of-range, extra-field and wrong-type variants independently. Their proof is IR immediate validity; they cannot establish source-token literal provenance, which remains a checker obligation. [CITED: 35-CONTEXT.md D-09..D-11]

### Pattern 6: Keep event-free operation accounting consistent

The interpreter dispatch, cgen event projection, capacity calculation, operation schema, and execution peer have separate event-free inventories. Cgen `invocationEventCapacity` is a conservative bound based on operation count; keep its projected size conservative while confirming pure ops produce no actual events. `schema2ExecutionDocumentSize` has a per-op projection switch; add the new operations to its explicit event-free arm so default-unknown remains fail-closed [VERIFIED: `internal/compiler/cgen/cgen_program.go:173-278`]. Do not invent a runtime defect/reason; `session.go` has a separate defect-reason switch and shifts are pure [VERIFIED: `internal/compiler/session/session.go:527-545`] [CITED: `35-CONTEXT.md D-17`].

### Pattern 7: Independent complete-source manifest witnesses and baseline

The manifest is explicitly a leaf consumer and declares `const Schema = "schway.surface/1"`; its current ten JSON fields are `schema`, `coordinates`, `types`, `operators`, `statements`, `builtins`, `calls`, `restrictions`, `limits`, `refusals` [VERIFIED: `internal/compiler/surfacemanifest/manifest.go:1-34`]. Extend this existing shape; don't add a parallel discovery framework or feed producer results into expected witnesses. Existing Phase 34 manifest tests include named witness and source/limit consumer tests [VERIFIED: `internal/compiler/session/session_phase34_surface_test.go:452-534`; `session_phase34_limit_consumers_test.go:22-69`].

Before changing compiler behavior, follow the prior finite generated-C corpus pattern: it enumerates tracked `.schway` sources and records emitted/non-emitting outcomes, with byte drift and inertness controls. Existing `TestPhase34PreChangeCProgramGolden` and `TestPhase34PreChangeCGoldenIsNotInert` are real named anchors [VERIFIED: `internal/compiler/session/session_phase34_prechange_test.go:21-24`; `285-318`]. Audit all exported emitter routes: the existing corpus route covers `cgen.EmitApplication` and `cgen.EmitNative`; `cgen.Emit` is also used by evidence code and should either be included in the new baseline or explicitly justified as outside the phase's literal source corpus [VERIFIED: `internal/compiler/session/session_phase34_prechange_test.go:66-118`; `internal/compiler/evidence/evidence.go:211`]. Do not replace old goldens or treat Adler's intentional source rewrite as a no-feature case.

### Anti-Patterns to Avoid

- **Treating constant folding as literal provenance:** a named constant or `(1 + 2)` must not qualify for a shift count.
- **Global angle-token changes without contextual type tests:** can break nested generic closers or comparisons; pair recognition must require source adjacency.
- **One shared transfer function for independent peers:** it collapses the independence the phase is meant to test.
- **Testing only two engines agree:** both could share the same mistake; use independently authored expected values and reached wrong-answer controls.
- **Calling executionpeer numeric validation:** it has no entry point/value and validates observed causal structure only.
- **Weak or inert baseline:** enumerate non-emitting rows, run all selected emitter routes, and arm byte-drift controls before edits.
- **Normalizing numeric routes as cleanup:** would alter `010` or `0B10` behavior outside scope and may change old emitted bytes.
- **Relying on UBSan alone:** still require literal expected vectors, static shift-count scan and malformed immediate rejection controls.

### Suggested Plan Grouping

Keep the work to about eight coherent plans, with the following gates rather than one plan per operator or peer:

1. **Pre-change evidence gate:** resolve and record the D-18 amendment before plans; create immutable generated-C corpus with emitted/non-emitting rows, reached byte-drift control, existing `010`/`0B10` route policy pins and Bool identifier baseline.
2. **Frontend grammar:** add operator tokens/pairs and precedence, bounded transparent grouping, depth/resource refusal, lossless recovery, formatter idempotence and nested-generic controls.
3. **Numeric surface:** narrow uppercase-hex conversion, malformed-candidate ownership, overflow/span/spelling controls, and initial manifest surface/refusal/limit facts.
4. **Core/check producer:** register the two immediate kinds; define the `ConstU64` field exclusivity; add independent source literal admission and checker lowering with U64 transfer.
5. **Core validator:** implement raw immediate shape/type/range and transfer independently; exercise block replay, straight-line analysis, CFG detection and real fixed-point triggers with reached controls.
6. **Origin and value/event peers:** update independent origin admissions/transfers/recompute/publication triggers, pathoracle operation inventory, interpreter values, executionpeer event-free/fabricated-event controls, and cgen event/capacity/schema/function path inventories.
7. **Focused semantic evidence:** add independent authored literal vectors, wrong-answer controls, shift corpus scan, targeted O0/O3/LTO and UBSan checks, plus manifest witness/census controls. Gate on D-07 entry witness resolution.
8. **Runnable consumer and closeout:** replace Adler's 16 doublings with `b << 16`, preserve Phase 29/31 answers/outputs, run the reached wrong-shift control and write the phase summary/verification with local lane limits.

Plans 5–7 are justified by the independent validation, C undefined-behavior and event-free evidence obligations. Further splitting should name a specific runnable witness or safety obligation, per D-22.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Numeric parser/radix conversion | A fresh tokenizer or broad converter that silently unifies routes | Existing number token, `strconv.ParseUint`, and small route-local prefix handling | Lexer already owns candidate validation/spans; compatibility requires `010`/`0B10` route split remain pinned. [VERIFIED: `internal/compiler/syntax/lexer.go:54-74`; `internal/compiler/check/check.go:6809-6818`; `internal/compiler/check/scalar.go:608-668`] |
| Generic recursive expression parser framework | New parser dependency or generated grammar | Existing Pratt parser plus a local bounded group atom/pair helper | Existing parser already owns precedence, refusal recovery and source spans. [VERIFIED: `internal/compiler/syntax/parser.go:749-823`] |
| Generic shift safety framework | Runtime shift checks, masked counts, signed fixups, or optimizer assumptions | Checker literal admission plus immediate-only core facts and independent peer validation | Literal bounds eliminate dynamic-count ambiguity and enable safe unsigned C generation. [CITED: 35-CONTEXT.md D-08..D-12] |
| Runtime replay/event layer for pure integer ops | Operand/result event encoding or new defect reasons | Existing independent static transfer and engine answer controls; event-free projection | A new event violates the locked pure/event-free slice and does not repair executionpeer's missing value replay. [CITED: 35-CONTEXT.md D-16..D-18] |
| Manifest discovery or generated expected inventory | New manifest version or source-produced expected witnesses | Existing `schway.surface/1`, independent complete-source witnesses and finite source/admission/limit census | The schema is a leaf consumer; source-fed expected values would make discovery tests circular. [VERIFIED: `internal/compiler/surfacemanifest/manifest.go:1-34`] [CITED: `35-CONTEXT.md D-19`] |

**Key insight:** Phase correctness depends on independent contracts agreeing across source admission, core producers, each validator, both engines and evidence projections. Share syntax/IR primitives where appropriate, but preserve independent validation and evidence or a common omission can pass every layer.

## Common Pitfalls

### Pitfall 1: `|` admission accidentally accepts `||`

**What goes wrong:** The new bitwise OR parser consumes the first pipe of a logical OR refusal or reformats its two tokens as a valid bitwise operation.

**Why it happens:** The current unsupported-OR rule checks two contiguous pipe spans; operator recognition can run before that check. [VERIFIED: `internal/compiler/syntax/parser.go:798-823`]

**How to avoid:** Detect contiguous `||` before admitting `|`; tests must distinguish contiguous `||`, spaced `| |`, and a single `|`. Preserve the refusal code/hint for logical OR.

**Warning signs:** Any valid bitwise tree produced from source `||`, changed refusal ownership/count, or old formatter golden drift.

### Pitfall 2: Shift tokens steal nested type closers

**What goes wrong:** `<<`/`>>` tokenization changes how nested generic type references close, or formatter output no longer reparses.

**Why it happens:** The current type parser consumes one `>` at a time and scalar formatting spaces single angle tokens differently from type formatting. [VERIFIED: `internal/compiler/syntax/parser.go:941-969`; `internal/compiler/syntax/format.go:384-395`]

**How to avoid:** Require contiguous spans and recognize pairs contextually in scalar parsing; add nested-generic parse/format/reparse/idempotence controls, comparison cases and operator whitespace cases before plans count on the chosen token strategy.

**Warning signs:** Any changed generic AST/type close, lost comment/string text, or formatted source that parses into a different operator tree.

### Pitfall 3: Recursive parser path remains unbounded

**What goes wrong:** Adversarially deep parentheses or nested builtin indices exhaust Go stack before a normal diagnostic.

**Why it happens:** The current scalar atom parser recursively enters `!` and builtin index expressions; a source-length/token-size limit does not bound recursive call depth tightly enough. [VERIFIED: `internal/compiler/syntax/parser.go:843-905`] [CITED: `35-CONTEXT.md D-05`]

**How to avoid:** Apply one separately named expression-depth guard across groups and recursive indices, bound/remove the `!` recursion, and recover iteratively through delimiters while continuing to sibling diagnostics. Include depth 64 and 65 reached controls plus a nested-index regression.

**Warning signs:** Panic/stack overflow or skipped sibling diagnostics on malformed deep input; reusing a `type_depth` refusal for expression recursion.

### Pitfall 4: A constant value is mistaken for a source literal

**What goes wrong:** A named value or computed expression is accepted as the amount because propagation finds a constant.

**Why it happens:** Checker transfer knows values after syntax has already lost the source-form distinction.

**How to avoid:** Inspect AST leaf kind before lowering; transparent grouping can expose the same leaf but arithmetic and identifiers stay nonliteral regardless of known value. Emit the named surface refusal for nonliteral and representable 64+ values; let malformed/unrepresentable tokens retain their own diagnostic ownership. [CITED: 35-CONTEXT.md D-04, D-10, D-15]

**Warning signs:** A mutation replacing the immediate with a computed or named operand survives; two errors are emitted for one malformed numeric candidate.

### Pitfall 5: One analysis trigger remains unreachable for new operations

**What goes wrong:** New transfer code is present but not run because one peer only enters its fixed point for existing arithmetic/CFG triggers.

**Why it happens:** Admission, straight-line replay, block replay, scalar-CFG detection, origin-return recomputation and problem publication use separate inventories/triggers.

**How to avoid:** Include for each new op a table row naming actual trigger and transfer branch. Exercise scalar CFG and straight-line cases where those are distinct. For originvalidate explicitly test `originHasScalarCFG`, `RecomputeOriginPerReturn`, and its publication trigger; do not assume a new switch arm is enough. [VERIFIED: `internal/compiler/corevalidate/corevalidate.go:2379`; `2928-2955`; `3287-3314`; `4676-4685`; `4722-4850`; `4892-4949`; `internal/compiler/originvalidate/originvalidate.go:198-212`; `293-300`; `1330-1449`]

### Pitfall 6: Result mutation is not observable by the selected peer

**What goes wrong:** A wrong-result mutant is declared caught by a peer that only checks shape, range, safety or causal events.

**Why it happens:** Independent checkers do not all return an externally compared numeric answer; executionpeer's `Validate` receives only core program and execution document, and validates causality, not source entry/result. [VERIFIED: `internal/compiler/executionpeer/executionpeer.go:70-81`; `380-420`; `514-557`]

**How to avoid:** For check/core/origin transfer faults, use a downstream observable static obligation such as making a remainder divisor zero, or assert each peer's independently derived transfer result. For executionpeer use per-op event-free admission and fabricated step/defect-event controls. Put numeric expected and wrong-answer checks through interpreter/native. This is the D-18 acceptance correction, not evidence that causal admission proves values. [CITED: 35-CONTEXT.md D-16..D-18]

### Pitfall 7: C shift promotion or UB invalidates source-level semantics

**What goes wrong:** A C left operand is promoted to signed `int`, or a future unchecked count becomes width-sized, so generated C has undefined behavior or mismatches U64 wrap semantics.

**Why it happens:** C shift rules apply to integer-promoted operands; invalid shift counts and signed-left-shift conditions have undefined behavior under C17 §6.5.7. [CITED: https://www.open-std.org/jtc1/sc22/wg14/www/docs/n2176.pdf; https://open-std.org/JTC1/SC22/WG14/issues/c11c17/issue0463.html]

**How to avoid:** Validate each count to 0..63 before emission and cast left operand to unsigned long long before shifting, then cast to uint64_t for low-64-bit results, per D-12. Use focused UBSan plus a generated-C scan that counts actual shift sites and fails on dynamic or 64+ amounts. [CITED: 35-CONTEXT.md D-08, D-12; https://clang.llvm.org/docs/UndefinedBehaviorSanitizer.html#available-checks]

### Pitfall 8: “Every peer” baseline silently misses an emitter route

**What goes wrong:** A new corpus golden appears exhaustive but never calls an exported emitter used by another consumer.

**Why it happens:** Current Phase 34 frozen corpus uses `EmitApplication` and `EmitNative`; evidence code also uses `Emit`. [VERIFIED: `internal/compiler/session/session_phase34_prechange_test.go:66-118`; `internal/compiler/evidence/evidence.go:211`]

**How to avoid:** Enumerate selected routes explicitly, include non-emitting cases, assert baseline rows were reached, and arm a changed-byte control before code changes. Record the third emitter route or a concrete exclusion rationale.

## Code Examples

These are implementation shapes for planning, not code to copy verbatim. The source contract comes from D-09/D-12.

### Immediate-only U64 shift lowering

```go
// First establish source provenance: after transparent parentheses, amountExpr
// must still be the parser's number-leaf form. Do not constant-fold this test.
amount, err := strconv.ParseUint(rawLiteral, 0, 64)
if err != nil {
    return unsupportedOrExistingNumericFailure()
}
if amount >= 64 {
    return shiftAmountRefusal()
}
op.Kind = shiftKind
op.LeftID = valueID
op.RightID = ""
op.ConstU64 = strconv.FormatUint(amount, 10)
```

The field/op names and required literal range are grounded in D-09/D-10, not yet implemented in the repository. Do not blindly use base 0 for every existing numeric route: it would change legacy `010` behavior; gate its use to shift amount conversion and admitted decimal/hex token forms. [CITED: 35-CONTEXT.md D-09, D-10, D-14]

### C17 code-generation shape

```c
uint64_t result = (uint64_t)((unsigned long long)lhs << 16);
uint64_t logical_right = (uint64_t)((unsigned long long)lhs >> 63);
```

Emit only after validator has proved each immediate is 0..63. Keep the cast in generated text so the compiler does not depend on host `unsigned long long` being exactly 64 bits; final cast implements low-64-bit result. [CITED: 35-CONTEXT.md D-08, D-12; WG14 N2176 §6.5.7]

### Independent transfer and observable sink

For each peer, seed an input vector that distinguishes the selected bitwise/shift result, then either assert that peer's own transfer facts or route the result to an obligation whose verdict changes (for example, a remainder divisor that is zero only under the expected transfer). Keep admission mutations separate from transfer mutations. A static safety checker with no exposed result cannot be said to have caught a wrong returned number. [CITED: 35-CONTEXT.md D-16, D-18]

## State of the Art

| Existing approach | Phase 35 approach | When Changed | Impact |
|-------------------|-------------------|--------------|--------|
| One-token angle operators in the scalar Pratt parser | Context-local, contiguous `<<`/`>>` pair parsing while preserving one-angle type closes | Phase 35 | Adds shifts without changing nested generic delimiters or token spelling. [VERIFIED: `internal/compiler/syntax/token.go:43-51`; `internal/compiler/syntax/lexer.go:240-275`] [CITED: `35-CONTEXT.md D-06`] |
| `ConstU64` belongs only to `OpConst` | Same existing optional field also carries canonical shift immediate on two kinds | Phase 35 | Avoids a mandatory schema field while making malformed shift operands independently checkable. [VERIFIED: `internal/compiler/core/core.go:969-972`] [CITED: `35-CONTEXT.md D-09`] |
| No scalar grouping/expression-depth resource limit | Transparent bounded groups and a distinct expression-depth refusal/manifest limit | Phase 35 | Enables the locked witness while explicitly bounding recursion. [VERIFIED: `internal/compiler/syntax/parser.go:52-92`] [CITED: `35-CONTEXT.md D-04..D-05`] |

**Deprecated/outdated:** No external package/API is being adopted. Do not infer a source-language policy from Rust or Go precedence/literal grammars; those primary references are contrasts, while Schway's precedence and numeric behavior are locked by D-03/D-13/D-14.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | [ASSUMED] The local Clang toolchain can link programs with `-flto` and the installed UBSan runtime is available for the focused shift lane. Clang itself is installed, but those flags/runtimes were not exercised in this research. | Environment / Validation | Planning may need a bounded fallback or an environment checkpoint if the focused target cannot run locally. |
| A2 | [ASSUMED] Existing linearly parsed nested builtin indices admit depths greater than the new expression limit or otherwise change diagnostic/resource outcome when the 64 bound is applied. This was not executed; characterize it before implementation as D-05 requires. | Parser pattern / Open Questions | Could accidentally change existing accepted source behavior or claim a false compatibility delta. |
| A3 | [ASSUMED] `cgen.Emit` belongs in the no-feature baseline route set. It is a public emitter used by evidence code, but whether the frozen baseline needs to call it directly remains a planner choice. | Baseline pattern | Omitting a relevant emitter could leave a compatibility path unpinned. |
| A4 | [ASSUMED] D-07's public no-input `main` can be reconciled with current run-entry contracts without broadening an unrelated CLI/parameter boundary. No implementation route has been tested. | Open Questions | The phase's advertised public runnable witness may need a narrower contract or separate explicit scope. |

All other assumptions are either locked in `35-CONTEXT.md` or are source observations tied to the cited opened files. No new dependency/backend/runtime defect/event is recommended.

## Open Questions

1. **Resolve the D-07 runnable entry contradiction before plans allocate implementation tasks.**
   - What we know: parser function declarations require one parameter (`internal/compiler/syntax/parser.go:543-552`); current `interpreterInputs` supports Byte, Buffer and Match cases and returns no input for the default route (`internal/compiler/session/session.go:1224-1255`); scalar checking accepts U64 or ByteView parameters (`internal/compiler/check/scalar.go:37-50`).
   - What's unclear: how D-07's required no-input `main` source passes both current public `schway run --engine=interpreter` and native entry routes without adding entry behavior or changing function signature contract.
   - Recommendation: before planning the public witness task, have the parent resolve this within the delegated D-07 decision (e.g. amend the witness to the narrowest currently supported entry form if a verified form exists, or explicitly authorize a small scoped run-entry addition). Do not assert that current public run already works. This agent did not execute a CLI trial.

2. **D-18 live acceptance wording needs its required dated amendment.**
   - What we know: as read during research, `.planning/REQUIREMENTS.md` BIT-01 and `.planning/ROADMAP.md` criterion 1 call for wrong-result mutants at all peers; `executionpeer.Validate(program, document)` has no entry parameter and checks only observed causal structure. [VERIFIED: `internal/compiler/executionpeer/executionpeer.go:70-81`; `REQUIREMENTS.md:35`; `ROADMAP.md:195`]
   - What's unclear: the exact dated note/wording parent will add to both live artifacts before plan creation.
   - Recommendation: parent should apply D-18's explicit auto-followed correction before planning; retain provenance/original ask, and keep executionpeer's acceptance to per-op event-free plus forged-event controls. Research does not edit either artifact.

3. **Baseline emitter completeness.**
   - What we know: current frozen baseline test uses `EmitApplication` and `EmitNative`; `Emit` is used by evidence code.
   - What's unclear: whether the generated-C corpus should exercise direct `Emit` too, or document why application/native routes cover the phase's output contract.
   - Recommendation: name every emitter route selected and make baseline row coverage non-vacuous; preserve non-emitting source cases.

4. **Focused native matrix execution support.**
   - What we know: local host is macOS arm64 with Go 1.24 and Apple Clang 21; no `-O0`/`-O3`/LTO/UBSan test ran here.
   - What's unclear: whether native LTO and sanitizer runtime work in this environment.
   - Recommendation: include the focused matrix as an explicit named feature test, but treat unavailable local LTO/UBSan as an honest lane gap; do not widen to a broad native suite or hosted dispatch.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | Compiler/package tests | ✓ [VERIFIED: terminal version probe] | go1.24.0 darwin/arm64 | — |
| Clang | Generated native code and optimization tests | ✓ [VERIFIED: terminal version probe] | Apple clang 21.0.0 (clang-2100.1.1.101) | — |
| Clang LTO linker/runtime | Focused `-O3 -flto` feature test | ? (not tested) | — | Keep as a named lane; resolve only if focused probe fails. |
| Clang UBSan shift runtime | Focused shift sanitizer test | ? (not tested) | — | Use static literal scan and unsigned C review as separate evidence; do not claim sanitizer ran. |

No package registry installs or external services are needed. No fallback changes feature scope or introduces another backend.

## Validation Architecture

Nyquist validation is enabled in `.planning/config.json` (`"nyquist_validation": true`) [VERIFIED: `.planning/config.json:26`]. No compiler, interpreter, native, or feature tests were run during this research. Commands below are focused test selectors for planning, not test receipts. The `TestPhase35...` selectors are proposed tests and will not match until implemented.

[ASSUMED] Each `TestPhase35...` name below is a proposed test selector, not a currently defined Go test. Verify it appears in verbose output once added; a selector with no matching test can otherwise return success.

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go `testing` package (toolchain Go 1.24.0) |
| Config file | `go.mod` (module/test configuration; no separate framework config required) |
| Quick run command | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax -run '^(TestPhase19NumericToken|TestPhase19NumericMalformed|TestPhase19NumericLiteralRoundTrip|TestFormatLetBinaryExpressions|TestPhase35ScalarGrammar|TestPhase35NumericOwnership|TestPhase35ExpressionDepth|TestPhase35FormatterRoundTrip)$' -count=1 -v` |
| Full suite command | No phase-local full-suite command is authorized. Run the bounded package selectors in this section; no whole-package or project-wide command. |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| BIT-01 | Independent expected U64 bitwise results across interpreter and native optimization modes; peer-transfer mutants; causal event controls | unit + differential native integration + mutation | See proposed `TestPhase35FocusedValueAnswers`, `TestPhase35BitwiseOptimizationMatrix`, transfer-mutant commands below | New tests required; existing operation inventory tests exist |
| BIT-02 | Literal shift range/provenance, independent immediate admission/transfers, event-free evidence, unsigned generated code/static scan, sanitizer lane | unit + mutation + focused native integration | See proposed `TestPhase35ShiftAdmission`, `TestPhase35ShiftImmediateAdmission`, `TestPhase35PureOperationCausalControls`, `TestPhase35UnsignedShiftEmissionAndEventProjection`, `TestPhase35ShiftUBSan` below | New tests required; existing validators and event inventories exist |
| BIT-03 | Numeric prefix policy, full malformed candidate ownership, overflow, source spelling/format/Tree.Bytes preservation | unit + formatter golden | Existing Phase 19 syntax/check selectors and proposed `TestPhase35NumericRoutePolicy`, `TestPhase35NumericOwnership`, `TestPhase35FormatterRoundTrip` | Existing numeric tests; new hex edge coverage required |
| BIT-04 | Adler `b << 16` preserves existing vectors/checksum/adler_low outputs and catches wrong shift | cross-engine regression + mutation | Existing Phase29 interpreter/native fixture tests, `TestPhase31CaseMatrix`, and proposed `TestPhase35ExistingAdlerAnswers` | Existing named tests; new wrong-shift control required |

### Focused selectors and commands

Every command uses an anchored named selector, isolated cache and `-v -count=1`. Go `-run` can return success when no tests match, so inspect verbose output and require every expected test name to appear; a no-match exit code is a false pass. New Phase35 test names below are proposed names to create. Do not substitute `go test ./...`, a whole-package run, broad optimization/native matrix, push or hosted dispatch.

**1. Before changing compiler code, pin current corpus and compatibility: **

```bash
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase34PreChangeCProgramGolden|TestPhase34PreChangeCGoldenIsNotInert|TestPhase34BoolNameBaseline|TestPhase34IdentityBaseline)$' -count=1 -v
```

This confirms an existing baseline, not the new Phase35 baseline. Add a new immutable enumerated generated-C baseline and a reached byte-drift control before production edits as required by D-02.

**2. Syntax, formatter and numeric ownership:**

```bash
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax -run '^(TestPhase19NumericToken|TestPhase19NumericMalformed|TestPhase19NumericLiteralRoundTrip|TestFormatLetBinaryExpressions|TestPhase35ScalarGrammar|TestPhase35NumericOwnership|TestPhase35ExpressionDepth|TestPhase35FormatterRoundTrip)$' -count=1 -v
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/check -run '^(TestPhase19LiteralAdmission|TestPhase19LiteralRange|TestPhase19NumericRefusalFrontiers|TestPhase35NumericRoutePolicy|TestPhase35ShiftAdmission|TestPhase35ScalarTransferMutants)$' -count=1 -v
```

**3. Operation inventory and independent raw-core contracts/transfers:**

```bash
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/core -run '^(TestAllOperationKindsRegistered|TestAllOperationKindsHandledAtEverySite)$' -count=1 -v
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/corevalidate -run '^(TestPhase19OpConstIndependentAdmission|TestPhase19ForgedOpConstFactsRefused|TestPhase35ShiftImmediateAdmission|TestPhase35ScalarTransferMutants)$' -count=1 -v
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/originvalidate -run '^(TestPhase35ShiftImmediateAdmission|TestPhase35ScalarTransferMutants)$' -count=1 -v
```

**4. Pure runtime, causal projection, and C emission:**

```bash
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp -run '^TestPhase35BitwiseShiftVectorsEventFree$' -count=1 -v
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/executionpeer -run '^TestPhase35PureOperationCausalControls$' -count=1 -v
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -run '^TestPhase35UnsignedShiftEmissionAndEventProjection$' -count=1 -v
```

The executionpeer test is event/causal evidence only; it must not be counted as numeric transfer validation. C emission controls should inspect the actual shift sites and reject both a dynamic-count and a count of 64 mutant.

**5. Engine values, optimization and existing consumer preservation:**

```bash
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase29InterpreterFixtureVectors|TestPhase29NativeFixtureVectors|TestPhase31CaseMatrix|TestPhase34SurfaceLimitConsumers|TestPhase34SurfaceManifestWitnesses|TestPhase34SurfaceManifestCensus|TestPhase34SurfaceManifestControls|TestPhase35FocusedValueAnswers|TestPhase35BitwiseOptimizationMatrix|TestPhase35ShiftUBSan|TestPhase35ExistingAdlerAnswers|TestPhase35PreChangeEmitterCorpus)$' -count=1 -v
```

The Phase29/31 cases above are existing named consumer witnesses; the Phase35 selectors are new. Keep optimization feature cases bounded to authored bitwise/shift vectors at `-O0`, `-O3`, and `-O3 -flto`, and a focused UBSan shift lane. No general native matrix is implied. Preserve existing expected answers; the wrong-shift mutant must be reached by a focused Adler/answer test.

### Sampling Rate

- **Per task commit:** run the narrow package selector for each touched owner, with `GOCACHE=/tmp/ai-lang-verification-gocache` and `-run '^(...named tests...)$' -count=1 -v`.
- **Per wave merge:** run the corresponding focused commands above; verify all selector names were actually run.
- **Phase gate:** all focused contracts, engine vectors, named consumer regressions and mutation controls green before `$gsd-verify-work`; disclose that broad project/hosted checks were intentionally not run.

### Wave 0 Gaps

- [ ] Add and reach a new immutable generated-C corpus enumerator before compiler edits; include non-emitting cases and all selected emitter routes, plus a changed-byte control.
- [ ] Add grouped/nested/adjacent-angle syntax, formatter reparse/idempotence, numeric policy, depth 64/65 and malformed ownership tests.
- [ ] Add one immediate-contract suite per corevalidate and originvalidate plus transfer faults tied to real fixed-point triggers.
- [ ] Add event-free/fabricated-event controls per op, interpreter vectors, generated-C count scan, focused optimization and UBSan lanes.
- [ ] Extend independent surface witness/census controls and add new refusal/hint/limit witnesses.
- [ ] Add an engine-level wrong-answer control and Adler wrong-shift control; ensure public-run entry-contract issue is resolved before the no-input runnable witness plan.

## Security Domain

`security_enforcement` is enabled at ASVS level 1 [VERIFIED: `.planning/config.json:49-51`]. Phase 35 processes untrusted program text and emitted C, so validation, resource bounds and well-defined unsigned semantics apply. The current ASVS 5.0 taxonomy separates validation/logic and encoding/sanitization; use those compiler-relevant categories rather than mapping parser work to auth/session/authorization/cryptography controls. [CITED: https://cornucopia.owasp.org/taxonomy/asvs-5.0; https://owasp.org/projects/asvs]

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V1 Encoding and Sanitization (ASVS 5.0) | yes | Preserve tokens and source spans; diagnose malformed numeric candidates without confusing or truncating source. |
| V2 Validation and Business Logic (ASVS 5.0) | yes | Validate grammar, literal provenance/range, malformed IR immediates, and bounded recursive parser resources. |
| Authentication / Session / Access Control | no | No identity, session or authorization boundary is introduced by this phase. |
| Cryptography | no | Bitwise operators are general integer computation, not a cryptographic guarantee; no crypto API or secret handling is added. |

### Known Threat Patterns for a source compiler and C backend

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Deep recursive source nesting | Denial of Service | Distinct expression-depth limit, iterative delimiter-aware recovery, reached 64/65 and malformed nesting tests. |
| Malformed numeric candidate accepted after partial scan | Tampering | Lexer owns full candidate and span; exact code/count/span controls for malformed hex and suffix/separator cases. |
| Raw core producer forges shift shape/range | Tampering | Independent corevalidate and originvalidate immediate/type/range checks; reached mutation controls. |
| Signed/invalid C shift emitted | Tampering / Elevation of Privilege | U64-only source type, 0..63 immediate, unsigned C left operand, static output scan and focused UBSan. |
| Validator omission because a trigger was not updated | Tampering | Independently exercise actual operation inventory, scalar CFG/fixed-point and origin publication/recompute triggers. |

No cryptographic safety or secret-handling claim follows from this feature.

## Living Roadmap and Seed Transition Review

This is scoped source/document inspection, not newly executed compiler evidence. No feature test ran during Phase 35 discussion/research; prior Phase 33/34 checks and M006 receipts remain historical at their recorded revisions. Whole-loop cold/warm latency and human readability are unmeasured. [CITED: `35-CONTEXT.md` Runnable Witness and Transition Review]

| Capability | User-visible program | Current blocker | Smallest complete slice | Checker / guarantee changes | Evidence and debt | Owner / next action | Observation that changes priority |
|---|---|---|---|---|---|---|---|
| Bitwise computation and Adler shift | No-input arithmetic witness; checksum/`adler_low` examples | New operators, grouping, shift core contract and uppercase-hex conversion | Phase 35 U64 `^ & |`, literal shifts, hex and `b << 16` | Source literal rule; immediate range/shape; independent check/core/origin transfers; interpreter/native answer guarantees; event-free peer/cgen inventories | BIT-01..04; `010`/`0B10` route discrepancy; D-18 acceptance correction; old-source C baseline | Phase 35 planner/executor; resolve D-07 entry contract and D-18 wording before plans | A named consumer needs dynamic shift or different bound |
| Imported scalar helper | `crc_step`-style helper called in a loop and through importing module | Scalar calls remain absent | Phase 36 one-argument acyclic U64/Bool imported helper | Independent call/argument/result derivation and callee-defect evidence | CALL-01/02; no new call framework in Phase 35 | Phase 36 | A named consumer proves arity-N or another call boundary is needed |
| CRC file consumer and second importer | `crc32 FILE`, second consumer, interpreter/native parity and clean-checkout guide | Operator/helper prerequisites and interpreter file CLI route | Phase 37 CRC-32/ISO-HDLC source module and bounded file path | Independent CRC vectors, second-import single-copy control and outcome parity | APP-13/14, MOD-07, DX-24; hosted receipt owner is Phase 38 | Phase 37; final hosted bind in Phase 38 | A real file consumer needs a smaller API or different failure outcome |

| Dormant seed | Disposition | Source witness / named consumer | Trigger before promotion |
|---|---|---|---|
| z1g bounded symbol/caller query | Retain dormant | `session_phase6_query.go`; future `crc_step` navigation | A named reviewer measures a query accuracy/cost gap. |
| trz failure observability | Retain dormant | `native_app.go#readApplicationCapture`, `session_app_verify.go`; Adler/CRC failure | A concrete failure remains unexplained; bounded-allocation concern was closed in Phase 33. |
| z1s compare/source forms | Retain dormant | `syntax/format.go#Format`; bitwise expression review | A real reviewer/source-form contract dispute appears. |

No seed is promoted. Do not start a database, telemetry framework or generic syntax study from speculative forecasts. Preserve archived and differently dated living-document claims; update them only if this active workflow changes facts. [CITED: `35-CONTEXT.md` D-23 and Runnable Witness and Transition Review; `AGENTS.md:101-130`]

## Sources

### Primary (HIGH confidence: opened project source / official standards and documentation)

- `AGENTS.md` — project constraints, GSD workflow, living roadmap transition behavior.
- `35-CONTEXT.md` — locked D-01..D-23 decisions, scoped runnable witness, phase boundaries, dormant seed dispositions.
- `.planning/PRODUCT-ROADMAP.md`, `.planning/LANGUAGE-MATURITY.md`, `.planning/sources/DASHBIT-AI-ERA.md` — living priority, maturity and source appraisal checked for the transition review.
- `.planning/seeds/SEED-261008-z1g-bounded-symbol-and-caller-query.md`, `SEED-261008-trz-consumer-triggered-failure-observability.md`, `SEED-261008-z1s-compare-contract-and-source-forms.md` — relevant dormant seeds and scoped retain dispositions.
- `internal/compiler/syntax/{token.go,lexer.go,parser.go,tree.go,format.go}` — current tokenization, Pratt operators/refusals, recursive atoms, type closes, formatting and lossless bytes.
- `internal/compiler/check/{check.go,scalar.go}`, `internal/compiler/core/core.go` — legacy/scalar numeric routes, AST lowering, closed operation kinds and optional immediate field.
- `internal/compiler/corevalidate/corevalidate.go`, `internal/compiler/originvalidate/originvalidate.go`, `internal/compiler/executionpeer/executionpeer.go` — independent shape/transfer triggers and causal-only peer contract.
- `internal/compiler/interp/interp.go`, `internal/compiler/cgen/{cgen.go,cgen_program.go}`, `internal/compiler/session/session.go` — runtime, emitter and run-entry/event/capacity inventories.
- `internal/compiler/surface/refusals.go`, `internal/compiler/surfacemanifest/manifest.go`, `internal/compiler/session/session_phase34_surface_test.go`, `session_phase34_limit_consumers_test.go` — installed refusal and finite independent manifest census patterns.
- `internal/compiler/session/session_phase34_prechange_test.go`, `session_phase34_compat_test.go`, Phase 29/31 named tests — existing baseline and consumer evidence anchors.
- WG14 N2176, C17 draft, §6.5.7 — promoted operands, shift counts and unsigned shift rules: https://www.open-std.org/jtc1/sc22/wg14/www/docs/n2176.pdf
- Clang UndefinedBehaviorSanitizer checks — out-of-range shifts and `-fsanitize=shift`: https://clang.llvm.org/docs/UndefinedBehaviorSanitizer.html#available-checks
- OWASP ASVS 5.0 taxonomy and project page — current validation/sanitization category mapping: https://cornucopia.owasp.org/taxonomy/asvs-5.0 ; https://owasp.org/projects/asvs

### Secondary (MEDIUM confidence: official language documentation used as contrast)

- Go integer literal syntax and scanner: https://go.dev/ref/spec#Integer_literals ; https://go.dev/src/go/scanner/scanner.go
- Go `strconv.ParseUint` package documentation: https://go.dev/pkg/strconv/?m=old
- Rust expression precedence and integer token grammar: https://doc.rust-lang.org/reference/expressions.html#expression-precedence ; https://doc.rust-lang.org/reference/tokens.html#integer-literals
- WG14 issue 0463 discussion of C left-shift wording (consulted as context, not treated as a final normative amendment): https://open-std.org/JTC1/SC22/WG14/issues/c11c17/issue0463.html

### Tertiary (LOW confidence)

- None used to support implementation claims.

## Metadata

### Parent resolution of pre-plan dependencies — 2026-10-14

The research body above preserves the source inspection and decision text available when authored. The live context subsequently clarifies D-07: use a narrowly admitted scalar entry with a provably unused Byte formal, U64/Bool local computation and no Byte use/conversion; public conformance outcome value 264 is distinct from bare application stdout 264 newline. Existing one-parameter declaration, Byte fixture binding and U64 input-required behavior remain fixed. This bridge is planned new admission, not already passing behavior. Independent source/core/origin laws and public-route checks are mandatory; see current 35-CONTEXT.md D-07. This resolution supersedes the open question and assumption A4 without inventing a successful CLI trial.

D-18's live BIT-01 and Phase 35 criterion 1 now carry the dated original-demand/proof-authority amendment in REQUIREMENTS.md and ROADMAP.md. Causal per-op controls remain independent but are not numeric replay. The requirement table above quotes the prior wording as research provenance, not the current acceptance contract.

One selector correction: TestPhase31CaseMatrix lives in internal/compiler/native/native_phase31_matrix_test.go, not the session package used in the example command above. Final plans/validation must use the actual native package and real selectors.

**Confidence breakdown:**
- Standard stack: HIGH — existing Go/C paths and local versions are source/probe grounded; no external package is proposed.
- Architecture: HIGH — source inspections covered syntax, producer, operation registry, independent validators, both engines, event projections, entry route and manifest consumers.
- Pitfalls: HIGH for current code seams and acceptance boundary; MEDIUM for cross-platform optimizer/sanitizer availability, not exercised here.

**Research date:** 2026-10-13  
**Valid until:** 2026-11-12 (stable local architecture; recheck Clang/UBSan/LTO availability when implementing the native evidence lane)
