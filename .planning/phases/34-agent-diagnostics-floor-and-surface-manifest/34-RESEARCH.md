# Phase 34: Agent Diagnostics Floor and Surface Manifest - Research

**Researched:** 2026-10-13  
**Domain:** Go compiler diagnostics, parser recovery, CLI source presentation, checked-language surface documentation  
**Confidence:** HIGH for in-repo architecture and constraints; MEDIUM for recommended plan partition pending planner review.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### Scope, dependencies and compatibility
- **D-01:** Preserve every previously admitted source program, its generated C bytes, diagnostic codes outside the specifically targeted refusals, byte coordinates, semantic diagnostic IDs and command result IDs under location-only enrichment. Capture literal pre-change goldens before compiler edits. Reducing cascading diagnostics may change the diagnostic set and therefore a failed command's ID; do not pretend those intentional changes are presentation-only. Pin the permitted code/count delta separately from the unchanged corpus.
- **D-02:** Use Go 1.24 standard-library and small local code. No parser replacement, Unicode-width dependency, terminal framework, schema generator, plugin system, LSP server or persistent compiler database. Sharing presentation mechanics or a code/hint catalog is acceptable; sharing the producer's admission conclusion with independent validation is not. A dependency requires a concrete failure of the local alternative, with its narrower boundary and benefit recorded.

#### Source locations, identity and safe text
- **D-03:** `start` and `end` remain zero-based, half-open byte offsets and the sole edit coordinates. Added `line` and `col` denote the span's start, both 1-based; `col` counts Unicode scalar values in the original source, not UTF-8 bytes, UTF-16 units, graphemes or terminal cells. A tab counts as one scalar in the wire coordinate. Document this in the wire contract and manifest. Byte offsets already supply the alternative machine coordinate; do not add a second column field.
- **D-04:** Enrich every present, source-locatable diagnostic span: primary, cause and repair, using each span's effective module unit. Preserve nil spans. Handle UTF-8, CRLF, LF, empty source, last line without a newline, zero-width EOF and multi-line ranges. CRLF is one line break; both terminator bytes map to the preceding line's end position, and the next byte starts the next line. A lone CR is visible escaped content. Count each invalid UTF-8 byte as one position. An offset inside a valid rune maps to that rune's start for display, retaining the original byte offset. Reject invalid/inverted/out-of-source spans for presentation rather than silently clamping edit authority. Goldens must pin these conventions.
- **D-05:** Location fields and presentation context must never enter identity. `errorID` and `errorWithRepairsID` currently marshal complete spans and causes, so simply adding exported fields would be a regression. Explicitly preserve the old byte-only identity representation, including nested cause spans and unit identity, or use an equivalent wire-only projection. Keep frozen `lang.diagnostic/0`, `/1` and `lang.command/*` strings; this additive extension is not a coordinated rename/schema migration. Source-less legacy constructors must retain their old serialized form and pinned IDs. Test enrichment before and after `InUnit`, repeated enrichment, repair-kind identity and `Result.Finalize`.
- **D-06:** Compute locations and snippets from the exact bytes used in the current parse/check invocation. Reuse `syntax.Tree.Source`, loaded `UnitTrees` and the loader's source/relative-path inventory. No second read by pathname, directory scan, source lookup triggered by a diagnostic's untrusted text, or change to the loader's single-IO-site rule. Carry presentation provenance through the command result path without publishing all source bytes or including them in identity. This covers imported-file parser, checker and peer diagnostics as well as entry-file errors.
- **D-07:** Distinguish known zero-width source positions from the existing synthetic `Span{}` sentinel. Operational/usage/global diagnostics without a verified source position omit `line`/`col` and snippets; they must not acquire a fabricated `file:1:1`. A real start-of-file or empty-file syntax error is locatable. Preserve the diagnostic and its failure outcome when presentation data is unavailable. Use the CLI-supplied entry label and loader-confirmed relative imported paths; do not disclose a new canonical absolute path or change JSON's dotted `unit` into a filesystem path.
- **D-08:** Text source errors retain code and ID, add `file:line:col`, and show one source line plus a caret/range marker mapped through rendering. Use a deterministic, bounded ASCII rendering: ordinary ASCII remains readable, tabs expand at fixed stops, non-ASCII/invalid bytes and terminal control/bidi characters are visibly escaped. This trades raw Unicode appearance for exact alignment without a Unicode-width dependency. Escape hostile filenames and source-derived message text too; do not emit ANSI or read terminal width. Crop around the primary span with visible omission markers, keep at least one caret for insertion/EOF, and indicate a multi-line range without printing further source lines. Planner chooses and pins the tab stop, crop width and exact escape spelling. JSON gains coordinates, not snippets/source bodies; snippets remain presentation-only. Keep existing stream destinations and exit classes.

#### Cause-oriented recovery and reproducible drafts
- **D-09:** Before parser/lexer changes, commit a deterministic agent-draft corpus and literal baseline diagnostic codes, spans and counts. The original CRC-32 and word-counter scratch sources were not found: `.planning/research/M007/` contains only `SUMMARY.md`. Reconstruct representative drafts, explicitly label their provenance, and do not call them recovered originals or reuse the summary's approximate counts as measured baseline. Include isolated one-cause drafts, both reconstructed multi-cause programs, nested malformed delimiters, sibling declarations, EOF, comments/strings and near misses. After repair, one diagnosed recovery cause gets one primary diagnostic while independent causes remain visible. Record exact before/after counts per fixture.
- **D-10:** Extend the existing parser with local context-aware recognizers and delimiter-aware synchronization. A damaged expression/list/block resumes at a valid sibling or caller-owned closer; it does not consume that closer or skip the remainder of the file. Do not globally stop after the first error or suppress every later `expected }`/`expected declaration`. Preserve real missing-closer errors that are independent of the original cause. Keep recovery progress checks and existing source/token/diagnostic/type bounds; new delimiter scanning must be iterative/bounded and may not introduce uncontrolled recursion.
- **D-11:** Lexer/parser coordination suppresses a generic unexpected-byte diagnostic only when the parser claims the exact offending token/span for a dedicated refusal or its single recovery cause. Do not filter by broad message/code prefixes, punctuation anywhere in the file, or all lexer failures. Comments, quoted strings, `->`, admitted punctuation and ownership keywords retain their existing meaning. Malformed near misses continue to get ordinary syntax errors. Lossless `Tree.Bytes()` must still reproduce rejected source, and recovered/rejected ASTs never pass the session parse-error gate into checking or native emission.
- **D-12:** Preserve the historic skipped-region cause/identity distinctions, including the spiral trio's different discarded spans and distinct diagnostic/result IDs. Test two independent mistakes surviving recovery and one-cause cascades disappearing. Keep a named overflow/truncation diagnostic when the existing diagnostic cap is reached; “one per cause” is not permission to hide that limit. Arm controls that restore cascading, swallow the next sibling or drop a separate cause, and prove the corresponding corpus assertion fails.

#### Unsupported forms and hints
- **D-13:** Give each of the seven forms a dedicated stable code: `surface.unsupported_logical_and` (`&&`), `surface.unsupported_logical_or` (`||`), `surface.unsupported_logical_not` (`!`), `surface.unsupported_parameter_arity` (more than one declaration parameter), `surface.unsupported_if_without_else`, `surface.unsupported_bool_literal_tail`, and `surface.unsupported_subtraction`. Each has one concise hint that states an available alternative or the actual boundary. Hints must not promise syntax that checking refuses, silently change overflow/short-circuit behavior, or recommend an unchecked subtraction emulation. A case per form asserts precise primary span, code, one-line hint and absence of its previous misleading/cascading parse errors.
- **D-14:** Recognition is contextual. Excess parameters concerns Schway and foreign declarations; existing call-argument refusals retain `check.call_arity_unsupported`, and builtin `byte_at(view, index)` keeps its two-argument shape. Subtraction recognition must not capture arrows, strings/comments or unrelated malformed bytes. `true` and `false` remain identifiers, not new reserved keywords. Before implementing the Bool-tail refusal, pin previously admitted uses as parameter/binding names and reproduce the exact unbound tail shape. Classify a literal-shaped tail only with sufficient binding/result context to exclude a bound identifier; use the appropriate existing stage, not a spelling-only lexer/parser rule. This pre-edit witness is a required safety obligation, not an invitation to admit Bool literals or widen scalar result types.
- **D-15:** Keep hints in one-line diagnostic message text and expose the same literal hint in the manifest's refusal entry. No new automatic repair/replacement/applicability is inferred from a hint. Existing repair selection consumes structured byte edits, not prose, and cross-unit repair refusal stays intact. An optional dedicated diagnostic hint field is unnecessary for this phase. Tests independently pin the expected code and hint; sharing a catalog between emitters and the manifest does not supply the independent oracle.
- **D-16:** Do not add bitwise or hex unsupported hints in Phase 34. Do not claim a dynamic-shift refusal before its emitter exists. Phase 35 introduces BIT-02's refusal and extends the manifest introduced in Phase 34; REQUIREMENTS' reference to “the Phase 34 manifest” names that artifact, not early implementation. Phase 36 likewise changes the call rules when scalar helpers land. This resolves the advisor's apparent future-registration conflict without changing requirements ownership.

#### Current capability manifest and independent drift proof
- **D-17:** `schway surface --json` emits a standalone `schway.surface/1` document, not a command envelope with timing/work metrics. Use fixed field order, canonical array order and one final newline. No source argument, source/filesystem scan, Clang probe, runtime launch, network, wall-clock timestamp, host-dependent paths or current Git hash. Successful output is exit 0/stdout; missing JSON mode, duplicate JSON flag, extra source/argument and unrelated flags are usage failures under the existing exit table. Handle the route before unrelated global flags can be stripped and accidentally accepted. Also support the existing global `--json surface` placement. Add the route to usage and the route inventory as `no-entry-source`; do not expand the module-aware set.
- **D-18:** The manifest describes the installed compiler's currently checked surface: types, operators, statements, builtins, call rules, contextual restrictions, limits and the implemented `surface.unsupported_*` code/hint entries. No scheduled capability/status, roadmap promise, parse-only construct presented as admitted, general String/record claim or unconditional “all functions compose” claim. An admitted item identifies the contexts where it works. Existing one-parameter declarations and restricted linear calls are separate from scalar expression-call support, which is absent until Phase 36. Each builtin declares its own arity/input/result/context. Discovery does not replace `check` on a complete program or independent peer admission.
- **D-19:** Limits name their scope, unit and meaning: source/token/declaration and module-loader limits, deterministic analysis work where relevant, file/output/evidence caps by applicable route, and call/type restrictions. Distinguish semantic constraints, compiler resource refusal and runtime/tool route budgets. Expose current known facts; do not imply that every number is a language-semantic limit or derive target facts from the current Apple host. Use existing constants when emitting values; independently pinned tests catch value/scope/unit drift. Do not create a new limits framework or enumerate incidental private tuning constants without an authoring consumer.
- **D-20:** An explicit small typed manifest is preferred to reflection/parser-generated claims. Keep an independently authored admitted/refused witness matrix and a bounded source census of relevant parser/checker operator, statement, builtin, type and refusal branches. Compare census/witnesses to the manifest and the manifest to direct checker/command behavior. Every admitted entry needs a concrete accepted source witness in its stated context; every refusal needs its rejected source witness and expected code/hint. The census must fail on an unaccounted newly introduced surface branch/code, and the reverse direction must fail on a manifest-only claim. A manifest registry feeding both sides is not independence. State the finite corpus/census coverage limit; no proof over every possible program is claimed.
- **D-21:** Arm each drift direction: remove an admitted entry, add an unadmitted claim, remove a refusal/code or hint, and weaken a context/arity/limit restriction without changing its name. Pin `byte_at` arity 2 separately from ordinary calls; pin absent scalar expression calls; include all seven current refusal codes. Restoration of each control is required, and each must fail for its named reason rather than an unrelated fixture break. These tests extend when Phase 35/36 adds operators/refusals/calls. Static inspection alone is insufficient to substantiate a positive admission claim.

#### Evidence lanes and planning handoff
- **D-22:** Pin generated C for the existing admitted corpus before implementation, following the existing pre-change golden/armed-control pattern. The planner must enumerate the corpus and explicit non-emitting refusal cases, not silently drop a program because generation failed. After implementation, every previously emitting program matches byte-for-byte; a reached emission mutation proves the comparison is live. No new backend or semantic peer rule is warranted by a presentation phase. Maintain independently admitted safety/refusal tests beside the cosmetic golden.
- **D-23:** Cover presentation through public `check` and every existing diagnostic-bearing source route affected by the shared path, including imported errors and non-source operational errors. JSON compatibility tests decode enriched output with existing repair consumers and show that eligible repairs, exact byte ranges and replacement behavior are unchanged. Preserve `TestRepairDriverDecodesNoProseFields`, cross-unit decline and module-aware/single-file route classifications. Do not make evidence/interface/query newly module-aware just to obtain snippets. Avoid rerunning source checks merely to render an error.
- **D-24:** Shift acceptance into focused table-driven location, rendering, recovery, refusal, manifest and CLI tests with explicit negative controls. Preserve legacy identity, module attribution and repair tests. Use only focused local selectors with `GOCACHE=/tmp/ai-lang-verification-gocache`; STATE forbids local project suites. Existing CI owns broader checks; Phase 38 owns final hosted receipt binding. No full suite, native matrix or new hosted receipt was run during this discussion. SUMMARY coverage entries must name real tests and use the closed coverage-kind vocabulary. Aim for zero human UAT because these contracts have deterministic proxies; do not claim human readability or latency improvement without measurements.
- **D-25:** Retain the three dormant seeds with scoped dispositions below. No measured named caller-query gap, unexplained runtime failure or disputed canonical-source form was established here. Keep the source appraisal's pre-Phase-33 capture-allocation observation historical; Phase 33's bounded read and reached mutant are its resolution. Current top-three capability order remains diagnostics/manifest, bitwise computation, scalar helpers. No change to active roadmap/requirements ownership is requested.
- **D-26:** Plan approximately eight coherent slices at most unless a named runnable witness or safety obligation justifies growth. Source snapshot/identity preservation, contextual Bool-name preservation, finite recovery termination and bidirectional manifest controls are concrete safety obligations; do not split every role/construct into its own plan. Keep a runnable manifest and broken-draft `check` in this phase so it does not become a second purely enabling phase after 33. Planning may refine helper layout, test placement and rendering constants without reopening these contracts. The user delegated recommendations; the persistent GSD auto-advance preference governs the subsequent plan/execute transition.

### the agent's Discretion
- Local helper/package boundaries, test-file placement and dependency-aware plan/wave split.
- Exact text layout, ASCII escape spelling, tab stop and crop width, once documented and golden-pinned within D-08.
- Hint prose within D-13/D-15, with an independently tested available alternative and one-line requirement.
- Position-presence representation and presentation-context plumbing, while preserving old bytes/IDs, source snapshots, unit attribution and unavailable-location honesty.
- Narrow recognition/recovery implementation, with D-14's binding-preservation witness before choosing a stage.

### Deferred Ideas (OUT OF SCOPE)
- Full recovery AST/IDE-quality malformed trees, display-width Unicode library, LSP/editor adapter and persistent program database: no current consumer justification.
- Scheduled-capability negotiation, schema/plugin framework, broad parser rewrite and automatic fix suggestions: unnecessary to discover current checked surface.
- Bitwise/hex/shift semantics and dynamic-shift refusal: Phase 35; scalar helper calls: Phase 36; interpreter file-input CLI route and executable CRC guide: Phase 37; hosted final binding: Phase 38.
- Arity-N, records, general Bool literals, subtraction semantics, module-aware evidence/interface/query and foreign-adapter simplification: retain their existing later ownership.
- Runtime telemetry, source-form experiments and named caller queries: dormant seed conditions above.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| DX-20 | Text diagnostics show `file:line:col` plus a one-line caret snippet; JSON spans add 1-based line/column beside bytes without changing codes. | `syntax.Tree.Source`, session/module source snapshots, and protocol projections provide the bytes and output boundary; identity-safe enrichment is mandatory. |
| DX-21 | One diagnostic per actual syntax recovery point, with no one-cause cascade. | Existing parser recovery primitives are local and progress-checked; the deterministic reconstructed corpus and exact before/after baseline must precede parser edits. |
| DX-22 | Seven still-absent syntax/form cases have dedicated `surface.unsupported_*` diagnostics and accurate hints. | Recognition belongs where syntax and binding/result context exist; `true`/`false` must remain identifiers; declaration arity is distinct from call arity. |
| DX-23 | A standalone deterministic manifest describes the checked surface and detects drift in both directions. | Typed manifest plus separately authored witnesses and bounded branch census; direct checker/CLI behavior closes both directions. Phase 35 adds dynamic-shift refusal later. |
</phase_requirements>

## Summary

Treat this as four coupled deliverables around a byte-preserving compiler: source-aware diagnostic projection, cause-oriented parser recovery, contextual refusal messages, and a truthful source-free surface route. The strongest implementation seam is the existing `session` command path: `syntax.Tree.Source` and module loader records already retain exact input bytes, while `protocol.JSON` and `protocol.Human` currently receive diagnostics without presentation provenance. Carry a sidecar presentation context through that path; do not add location metadata to identity-bearing diagnostic spans. [CITED: .planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-CONTEXT.md §code_context; VERIFIED: internal/compiler/session/session.go:36-43,777-796; VERIFIED: internal/compiler/diagnostic/diagnostic.go:130-178; VERIFIED: internal/compiler/protocol/protocol.go:581-644]

Before any compiler edit, establish two independent preservation witnesses: capture literal baseline diagnostics for reconstructed agent drafts and pin byte-identical C for every previously emitting corpus program. The original CRC and word-counter scratch drafts are unavailable, so label reconstructed fixtures honestly. A fresh check proved the contextual Bool-name requirement: parameter `true` and binding `false` are accepted; an unbound `true` tail currently gives `name.unknown`. [CITED: .planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-CONTEXT.md D-09,D-14,D-22; VERIFIED: internal/compiler/syntax/parser.go:1-30; VERIFIED: internal/compiler/session/session_phase30_prechange_test.go:323-375; VERIFIED: internal/compiler/check/check.go:777-798]

Keep the deliverable runnable: public `schway check` should render a damaged draft from the current input snapshot, and `schway surface --json` should work without a source path. Build the manifest from checked/admitted facts and per-form witness programs, then compare it against both a finite parser/checker census and direct checking results. It reports current behavior only; Phase 35 owns bitwise/shift/hex and dynamic-shift refusal, Phase 36 owns scalar helper calls. [CITED: .planning/REQUIREMENTS.md:25-37,111-113; CITED: .planning/ROADMAP.md:143-160; VERIFIED: cmd/schway/main.go:35-125]

**Primary recommendation:** First plan a pre-edit evidence baseline and preservation gate, then implement presentation plumbing, bounded recovery/refusals, and the source-free manifest/CLI with independently authored controls. Keep the total near eight coherent plans, with a runnable `check` and `surface --json` in this phase. [CITED: 34-CONTEXT.md D-24,D-26]

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Span coordinates and text snippet rendering | API / Backend | CLI | Session/protocol constructs the diagnostic view; CLI supplies entry filename and chooses text versus JSON. |
| Parser recovery and unsupported syntax recognition | API / Backend | — | Lexer/parser/checker determine syntax and checked-context facts; recovered AST must remain gated from checking/emission. |
| Checked surface manifest | API / Backend | CLI | A typed compiler-owned model reports checked capabilities; CLI owns a source-free command route and usage classification. |
| Repair edit authority | API / Backend | CLI repair tool | Byte ranges and structured replacement remain authoritative; line/column and hint text are presentation only. |

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go standard library | 1.24.0 on this host | Compiler, Unicode decoding, JSON, CLI and deterministic testing | Project's Stage 0 constraint; module declares `go 1.24`. [VERIFIED: `go version` output; CITED: AGENTS.md Bootstrap; VERIFIED: go.mod:1-3] |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| No new package | — | Location mapping, ASCII escaping, typed manifest, and tests | Implement local bounded helpers per D-02; no dependency justified by current source. [CITED: 34-CONTEXT.md D-02,D-08] |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Small local byte/rune renderer | Unicode display-width package / terminal renderer | Rejected by locked scope: exact deterministic ASCII caret mapping avoids dependency, terminal width and bidi/control rendering ambiguity. [CITED: 34-CONTEXT.md D-02,D-08] |
| Explicit typed manifest and independent census | Reflection or parser-generated registry | Rejected: generated facts can only show parseability and risk sharing the producer's conclusion with the test oracle. [CITED: 34-CONTEXT.md D-18,D-20] |
| Local synchronization extensions | Parser rewrite or rich recovery AST | Existing parser already has bounded local recovery; full rewrite/AST is deferred. [VERIFIED: internal/compiler/syntax/parser.go:1166-1214; CITED: 34-CONTEXT.md Deferred Ideas] |

**Installation:** None. This phase introduces no external dependency. [CITED: 34-CONTEXT.md D-02]

## Architecture Patterns

### Source and diagnostic flow

```text
CLI check FILE ──read once──> session / module loader
                                  │ exact bytes + verified relative file label
                                  ▼
                         syntax.Parse / checker
                                  │ byte spans, causes, repairs, unit identity
                                  ▼
                 command result + presentation sidecar
                         ┌────────┴─────────┐
                         ▼                  ▼
               protocol.Human       protocol.JSON
               escaped line/caret   byte spans + 1-based scalar line/col
```

This is a conceptual data flow. The loader is the single module IO site, and check stops before checker admission if parser diagnostics exist. [VERIFIED: internal/compiler/modload/load.go:92-99,130-151; VERIFIED: internal/compiler/session/session.go:777-796; CITED: 34-CONTEXT.md D-06,D-11]

### Recommended Project Structure

Keep helpers next to their owning layer unless implementation shows a stable reusable boundary. Likely touch points are `internal/compiler/diagnostic/`, `internal/compiler/protocol/`, `internal/compiler/session/`, `internal/compiler/syntax/`, `internal/compiler/check/`, `internal/compiler/modload/`, and `cmd/schway/`; tests should follow the existing phase-specific `*_test.go` style. [VERIFIED: listed source and test paths above; CITED: 34-CONTEXT.md code_context]

### Pattern 1: Presentation metadata separated from identity

**What:** Keep semantic spans as byte-only edit authority and derive line/column/snippet from exact in-memory source bytes at projection time. Carry whether a span has a verified source origin separately from `Span{}` so real offset zero can render and synthetic global diagnostics cannot. [CITED: 34-CONTEXT.md D-03..D-07]

**When to use:** Every source-bearing diagnostic in `check`, imported module diagnostics, peer diagnostics, and source-bearing format routes; source-less usage/operational failures have no fabricated location. [CITED: 34-CONTEXT.md D-06,D-07,D-23]

**Example contract (not a proposed exported Go API):**

```text
span.start / span.end: unchanged zero-based half-open byte offsets
span.line / span.col: derived 1-based start coordinate, scalar-counted column
source snippet: human-only, one escaped/cropped line; never JSON source text
identity projection: old byte-only Span and Cause representation
```

Rust diagnostics provide a primary-source analogue where byte offsets and scalar coordinates coexist; Go's token positions demonstrate a byte-column alternative. Schway's coordinate contract is the locked D-03 convention, not either tool's compatibility promise. [CITED: https://doc.rust-lang.org/rustc/json.html; CITED: https://go.dev/src/go/token/position.go; CITED: 34-CONTEXT.md D-03]

### Pattern 2: Cause-owned, bounded recovery

**What:** Local parse recognizers report one primary diagnostic for the cause they claim, synchronize iteratively at valid siblings or caller-owned closers, preserve skipped-region identity, and prove progress or boundary arrival. Suppress lexer errors only for the exact claimed token span. [VERIFIED: internal/compiler/syntax/parser.go:1166-1214, 1020-1056; VERIFIED: internal/compiler/syntax/parser.go:121-147; VERIFIED: internal/compiler/syntax/lexer.go:168-168; CITED: 34-CONTEXT.md D-10..D-12]

**When to use:** Malformed expression/list/block and seven known absent forms. Do not globally stop parsing or broadly filter `syntax.unexpected_byte`. Existing code keeps a named overflow error at the diagnostic cap. [VERIFIED: internal/compiler/syntax/parser.go:8-18,121-147; CITED: 34-CONTEXT.md D-10..D-12]

**Verification shape:** Baseline code/span/count JSON captured first; post-change fixture table asserts one diagnostic per target recovery cause, independent second cause survival, sibling preservation, stable cause identity distinctions, and explicit truncation. Arm controls for cascade restoration, swallowed sibling and dropped independent cause. [CITED: 34-CONTEXT.md D-09,D-12,D-24]

### Pattern 3: Manifest as claim set, validated by independent witnesses

**What:** Define a small typed, explicitly authored manifest with canonical serialization. Maintain a distinct witness matrix and bounded census of parser/checker branches; compare in both directions, then check representative accepted/rejected programs using the normal command/checker. [CITED: 34-CONTEXT.md D-17..D-21]

**When to use:** `schway surface --json`, with no source or host probing. Route it before global argument stripping can hide unsupported flags, and add it to usage and route inventory as `no-entry-source`. [VERIFIED: cmd/schway/main.go:35-125,679-680; CITED: 34-CONTEXT.md D-17]

**Independence requirement:** Never let manifest entries be the sole generator of witnesses or census expectations. Every admitted statement includes the exact context and an accepted source witness; every refusal includes the rejected witness, code and literal hint. Static census is finite and must disclose its branch coverage boundary. [CITED: 34-CONTEXT.md D-18,D-20,D-21]

### Anti-Patterns to Avoid

- Adding exported line/column fields directly to identity-bearing `Span`: current diagnostic IDs hash spans and nested cause spans, and command result IDs consume diagnostic IDs. [VERIFIED: internal/compiler/diagnostic/diagnostic.go:130-178; VERIFIED: internal/compiler/protocol/protocol.go:548-579]
- Treating `Span{}` as “known start of file”; operational diagnostics intentionally use zero spans. [VERIFIED: internal/compiler/session/session.go:869-889; CITED: 34-CONTEXT.md D-07]
- Re-reading a source path to render errors, or resolving a path using diagnostic message text. Use the parsed tree/source inventory. [VERIFIED: internal/compiler/modload/load.go:130-151; CITED: 34-CONTEXT.md D-06]
- Reserving `true`/`false` lexically or classifying every matching tail as a Bool literal. Fresh `schway check` probe accepted parameter `true` and binding `false`; the unbound `true` tail currently reports only `name.unknown`. The contextual discriminator must preserve both outcomes. [VERIFIED: fresh focused probe, `GOCACHE=/tmp/ai-lang-verification-gocache go run ./cmd/schway check /tmp/schway-phase34-bool-name.schway --json` and corresponding bool-tail probe; CITED: 34-CONTEXT.md D-14]
- Sharing a single manifest table as both production output and expected test oracle, accepting merely parsed syntax as admitted, or claiming `byte_at` and ordinary calls share arity rules. [CITED: 34-CONTEXT.md D-18,D-20,D-21]
- Filtering lexer diagnostics by broad prefixes or suppressing every later “expected” error. Match only a parser-owned exact token/span and retain valid siblings. [VERIFIED: internal/compiler/syntax/parser.go:121-147; CITED: 34-CONTEXT.md D-10..D-12]

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Exact source bytes and module provenance | A second read/path scan | Existing `Tree.Source`, `CheckResult.UnitTrees`, and `modload.Set.Sources` with `RelPath` | The loader is the single IO site and already parsed those bytes. [VERIFIED: internal/compiler/syntax/tree.go:5-17; VERIFIED: internal/compiler/session/session.go:36-43; VERIFIED: internal/compiler/modload/load.go:92-99; VERIFIED: internal/compiler/modload/resolve.go:137-145] |
| Semantic repair edits | Hint parsing or line/column edits | Existing structured repair byte spans/replacement and repair-driver decoder | Repair authority already depends on exact byte edits; hints are not automatic fixes. [CITED: 34-CONTEXT.md D-03,D-15,D-23] |
| Checked capability claims | Grammar inventory as admission oracle | Direct check witnesses plus independent bounded branch census | Parsing does not establish checked admission. [VERIFIED: internal/compiler/session/session.go:777-796; CITED: 34-CONTEXT.md D-18,D-20] |
| Terminal-width layout | Full Unicode display-width handling | Bounded escaped ASCII snippet and pinned rendering constants | Locked scope excludes width dependency and terminal framework. [CITED: 34-CONTEXT.md D-02,D-08] |

**Key insight:** Line and snippet are projections of the invocation's captured source, while IDs and repairs remain defined over original byte spans. Keeping those channels separate lets the compiler improve human feedback without changing repair authority or semantic identity. [CITED: 34-CONTEXT.md D-03..D-08]

## Common Pitfalls

### Pitfall 1: Coordinate enrichment changes IDs
**What goes wrong:** Exported fields on spans serialize into the old identity structs, changing diagnostic IDs and therefore command result IDs. [VERIFIED: internal/compiler/diagnostic/diagnostic.go:130-178; VERIFIED: internal/compiler/protocol/protocol.go:548-579]
**How to avoid:** Preserve an explicit old-shape identity projection and test schema/ID goldens before/after enrichment, including nested cause and repair cases, unit stamping and finalization. [CITED: 34-CONTEXT.md D-05]

### Pitfall 2: Synthetic offset zero becomes a fake location
**What goes wrong:** A plain zero span is also used for source-less operational/check-limit diagnostics, while actual syntax can fail at byte zero or empty EOF. [VERIFIED: internal/compiler/session/session.go:869-889; VERIFIED: internal/compiler/syntax/parser.go:45-55; CITED: 34-CONTEXT.md D-07]
**How to avoid:** Carry verified source-origin/position-presence state; validate bounds before projection and omit location when provenance is unavailable. [CITED: 34-CONTEXT.md D-04,D-07]

### Pitfall 3: Recovery fixes the first error by swallowing another cause
**What goes wrong:** A recovery loop can consume a caller-owned closer or valid sibling; blanket suppression can hide a truly independent missing closer or second error. [CITED: 34-CONTEXT.md D-10..D-12]
**How to avoid:** Build and record baseline code/span/counts first, assert both independent-error survival and one-cause cascade removal, and arm controls against sibling swallowing and dropped causes. [CITED: 34-CONTEXT.md D-09,D-12]

### Pitfall 4: Refusal classifier steals valid identifiers or builtins
**What goes wrong:** `true`/`false` are legal identifiers and `byte_at` specifically takes two inputs even though ordinary call support is narrower. [CITED: 34-CONTEXT.md D-14,D-18]
**How to avoid:** Land accepted parameter/binding-name witnesses before changing recognition; classify tail form with binding and result context; test foreign and Schway declaration arity separately from call arity. Fresh name probe passed with CLI status `pass`. [VERIFIED: focused probe described above; CITED: 34-CONTEXT.md D-14]

### Pitfall 5: The manifest advertises what the parser accepts, not what checking admits
**What goes wrong:** Parse-only constructs and broad “function call” claims overstate what the checked/compiler pipeline supports. [VERIFIED: internal/compiler/session/session.go:777-796; CITED: 34-CONTEXT.md D-18]
**How to avoid:** Use positive accepted source witnesses and direct checker/CLI behavior; pin `byte_at` arity 2 and absent scalar expression calls as separate facts. [CITED: 34-CONTEXT.md D-18,D-20,D-21]

## Code Examples

No external package/API example is needed. The verified local extension points are:

```go
// Source tree retains the exact parse bytes.
type Tree struct {
    Source []byte
    Tokens []Token
}

// Cause and Repair span ownership resolves to their explicit unit or the
// owning diagnostic unit through EffectiveUnit.
```

The quoted fields/method behavior are present in [internal/compiler/syntax/tree.go:5-17] and [internal/compiler/diagnostic/diagnostic.go:181-193]. Do not copy this into diagnostic identity; derive presentation coordinates in a separate projection. [CITED: 34-CONTEXT.md D-05,D-06]

## Runtime State Inventory

Not applicable: this is not a rename, refactor, or migration phase. It adds an additive diagnostic presentation projection and a CLI route under the settled Schway identifiers. [CITED: 34-CONTEXT.md Phase Boundary]

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | Focused compiler/CLI verification | ✓ | go1.24.0 darwin/arm64 | None required |
| External packages/services | None | N/A | — | No new dependencies |

The host is Apple arm64; do not encode this host fact in the manifest. No Clang, native build or external service is needed for the phase's bounded local checks. [VERIFIED: `go version` output; CITED: 34-CONTEXT.md D-17,D-19,D-24]

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go `testing` package; Go 1.24.0 installed |
| Config file | `go.mod` |
| Quick run command | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/diagnostic -run '^TestPhase34DiagnosticIdentityProjection$' -count=1` |
| Full suite command | Not permitted locally by STATE/D-24; existing CI owns broad checks. |

Commands below are prospective test names to create; none is claimed to exist or pass. Run only focused selectors with the required cache. [CITED: .planning/STATE.md current verification policy; 34-CONTEXT.md D-24]

### Phase Requirements → Test Map

| Req ID | Behavior / claim | Test Type | Prospective focused command | File Exists? |
|--------|------------------|-----------|------------------------------|--------------|
| DX-20 | Byte/rune line map handles UTF-8, CRLF, LF, EOF, invalid UTF-8 and zero-width spans; text escapes/crops with aligned caret and JSON adds fields to primary/cause/repair spans. | Unit + CLI | `go test ./internal/compiler/protocol -run '^(TestPhase34DiagnosticLocations|TestPhase34HumanSnippetRendering)$' -count=1` and `go test ./cmd/schway -run '^TestPhase34CheckDiagnosticPresentation$' -count=1` | New tests (Wave 0 baseline first) |
| DX-20 | Existing IDs, command result IDs, repair decoding/ranges and module unit attribution remain fixed when enrichment is repeated and when `InUnit` is called. | Unit + integration | `go test ./internal/compiler/diagnostic -run '^TestPhase34DiagnosticIdentityProjection$' -count=1`; `go test ./internal/compiler/session -run '^TestPhase34ImportedDiagnosticPresentation$' -count=1` | New tests; retain existing repair/unit tests |
| DX-21 | Reconstructed drafts have literal before/after code/span/count baselines; one recovery cause reports once, independent cause and sibling survive. | Parser corpus + negative controls | `go test ./internal/compiler/syntax -run '^TestPhase34RecoveryCorpus$' -count=1` | New fixture corpus and test; baseline must be committed before parser modifications |
| DX-21 | Recovery terminates under existing caps, preserves caller-owned delimiters and lossless rejected bytes; arm cascade, sibling-swallow, and independent-cause-drop controls. | Parser mutation/control | `go test ./internal/compiler/syntax -run '^TestPhase34RecoveryControls$' -count=1` | New test; separate named overflow row |
| DX-22 | Each of seven forms gives exact span/code/one-line hint; near misses and valid `true`/`false` names keep current meaning; `byte_at` stays arity 2 and call-arity refusal stays distinct. | Parser/check table | `go test ./internal/compiler/syntax ./internal/compiler/check -run '^TestPhase34UnsupportedSurfaceRefusals$' -count=1` and `go test ./internal/compiler/check -run '^TestPhase34BoolSpellingsRemainIdentifiers$' -count=1` | New tests; fresh probes establish baseline only |
| DX-23 | Typed manifest deterministic bytes/newline/canonical order, no host/source values, accurate current checked claims and all refusal hints; direct source witnesses and independent finite census compare both directions. | Unit + behavioral census | `go test ./internal/compiler/check -run '^TestPhase34SurfaceManifestWitnesses$' -count=1` | New tests |
| DX-23 | CLI valid and invalid argument placements; source-free route, `--json surface`, missing/duplicate JSON, extra args, raw flags (including globally consumed `--depth`) all classified correctly; route inventory records `no-entry-source`. | CLI integration + route control | `go test ./cmd/schway -run '^(TestPhase34SurfaceRoute|TestPhase34SurfaceRouteInventory)$' -count=1` | New tests |
| DX-20..23 | Every pre-change emitting program produces byte-identical C, explicit refused cases remain represented, and an emission mutation reaches the golden comparator. | Compatibility gate + armed mutation | `go test ./internal/compiler/session -run '^(TestPhase34PreChangeCProgramGolden|TestPhase34PreChangeCGoldenIsNotInert)$' -count=1` | New snapshot/test; adapt Phase 30 pre-change precedent |

### Wave 0 Gaps and Evidence Ordering

- Capture `check` diagnostic code/span/count output from exact reconstructed fixtures before any parser or checker change. Record provenance as reconstructed; do not claim recovered originals. Include separate single-cause and multi-cause examples, nested delimiters, sibling declaration, EOF, comments/strings and near-misses. [CITED: 34-CONTEXT.md D-09]
- Enumerate existing admitted C-emitting source corpus and explicit refusal-only/non-emitting cases; commit literal C output before edits, with a reached mutation proving comparison. Do not silently exclude fixture build failures. [CITED: 34-CONTEXT.md D-22]
- Add baseline witnesses proving parameter `true`, binding `false`, and exact unbound `true` tail behavior before introducing `surface.unsupported_bool_literal_tail`. Newly executed local probes establish accepted names and the current unbound-tail `name.unknown` outcome; no code was edited. [VERIFIED: `go run ./cmd/schway check` probes reported above]
- No new dependencies or service setup are required. Existing CI owns broader Go checks; do not run a whole package/project suite, vet/build suite, native hosted receipt, or dispatch/push in this research phase. [CITED: 34-CONTEXT.md D-24]

### Cost Lanes

- Cheap recurring: focused tests for coordinate math, presentation, ID projection, route parsing, hints and manifest serialization.
- Medium focused: reconstructed corpus and finite source census with bidirectional negative controls; keep fixture count and census scope bounded and explicit.
- Existing CI only: full Go suites and broad cross-platform gates.
- Excluded here: whole-package/project tests, native hosted receipt or dispatch/push. Phase 38 owns hosted receipt binding. [CITED: 34-CONTEXT.md D-24,D-26]

## Security Domain

Security enforcement is enabled in `.planning/config.json`; this compiler phase handles untrusted source bytes and filenames and must preserve bounded work and output. [VERIFIED: .planning/config.json:49; CITED: 34-CONTEXT.md D-02,D-04,D-08,D-10]

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | No user/service authentication added. |
| V3 Session Management | No | No session or web state added. |
| V4 Access Control | Yes, route/repair boundaries | Keep route inventory and byte-edit authority unchanged; do not add a new module-aware route. |
| V5 Input Validation | Yes | Validate spans and offsets against exact source bounds; escape filenames, source snippets and source-derived message text; retain parser and render bounds. |
| V6 Cryptography | No | No cryptographic function or secret handling changes. |

### Known Threat Patterns

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Terminal control/bidi bytes or hostile filename in diagnostics | Spoofing / Elevation of presentation trust | Deterministic escaped ASCII, no ANSI, no terminal-width dependence. |
| Malformed spans or oversized/malformed source expanding work/output | Denial of Service / Tampering | Bounds-check spans; bounded crop and iterative recovery; retain source/token/diagnostic caps and named truncation. |
| Absolute path disclosure from imported diagnostics | Information Disclosure | Entry label supplied by CLI; imported labels from loader-confirmed relative provenance; do not infer a path from diagnostic content. |

ASVS is web-application oriented and supplies no specialized compiler diagnostic category; these controls derive from the phase's explicit untrusted-input and route constraints. [CITED: 34-CONTEXT.md D-04,D-06..D-08,D-10,D-17]

## State of the Art

| Analogue | Relevance | Boundary |
|----------|-----------|----------|
| rustc JSON diagnostics | Demonstrates structured source locations and separation of coordinates/context. | Does not define Schway's scalar-column or schema-compatibility contract. [CITED: https://doc.rust-lang.org/rustc/json.html] |
| Go parser | Local synchronization and progress checks are directly relevant patterns. | Do not replace Schway parser or copy parser architecture wholesale. [CITED: https://go.dev/src/go/parser/parser.go] |
| Clang diagnostics | Source location, snippet and caret are established compiler feedback. | Schway chooses bounded escaped ASCII and no terminal-width behavior. [CITED: https://clang.llvm.org/diagnostics.html] |
| rustc print options | Narrow source-free discovery command analogy. | Manifest is Schway-specific and reports only current checked claims. [CITED: https://doc.rust-lang.org/rustc/command-line-arguments/print-options.html] |

These primary-source analogues were consulted in the phase context on 2026-10-13; no new web research was needed because the implementation question is bounded to in-repo paths and decisions. [CITED: 34-CONTEXT.md canonical_refs]

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Exact helper/package placement and representation for source provenance is not yet selected. | Architecture | Could add unnecessary plumbing or distort API; planner should select the smallest sidecar that preserves unit and identity behavior. |
| A2 | Exact escaped spelling, crop width and tab stop remain open implementation choices. | Pattern 1 | Golden output could churn if selected without documenting and pinning. |
| A3 | The finite manifest census can be made exhaustive for the intentionally enumerated source branches without creating a language-wide reflection framework. | Pattern 3 | Omitted branch could escape drift detection; keep the census boundary explicit and controls reached. |
| A4 | Reconstruction can provide useful agent-style samples but cannot reproduce the unavailable original scratch corpus. | Validation | Baseline reductions are only claims about reconstructed fixtures, not historical CRC/word-counter drafts. |

## Open Questions — RESOLVED at planning (2026-10-13)

The original questions below are retained for provenance. These are selected planning answers and ordered execution obligations, not implemented behavior or passing test receipts.

1. **RESOLVED:** Choose a compact sidecar/projection representation and decide which command routes can retain the source snapshot without rerunning check or changing route classifications. Plan [34-02](34-02-PLAN.md) selects additive `Span.Line/Col` fields excluded from byte-only identity projections, an ignored `Diagnostic.SourcePosition` marker with `AtSource`, and ignored `Result.Presentation` carrying captured bytes/labels. Plan [34-03](34-03-PLAN.md) retains `CheckResult.Presentation`, loader source/relative-path inventory and imported early-error snapshots; Plan [34-08](34-08-PLAN.md) carries the same invocation through every affected diagnostic route and app-verify refusal without another read/check or expanding module authority (D-05..D-07,D-23).
2. **RESOLVED:** Choose and golden-pin text layout constants and escape spellings; no measurement justifies claims about human readability or feedback latency. Plan [34-02](34-02-PLAN.md), Task 3 selects tab stops of 4 and a crop bound of 120 rendered ASCII columns, doubled backslash, uppercase two-digit byte escapes and uppercase braced scalar escapes, visible crop markers, caret/range mapping and a first-line continuation indication. Its literal rendering goldens and documentation are execution obligations; readability/latency remain unmeasured (D-08,D-24).
3. **RESOLVED:** Enumerate the corpus and define exactly which source branches the finite surface census covers before assigning the manifest-test implementation plan. Plan [34-01](34-01-PLAN.md) enumerates all 162 tracked Schway inputs, recording both existing emitter outcomes and explicit non-emitting cases before production edits. Plan [34-07](34-07-PLAN.md) fixes the independent Go AST census boundary to `syntax/parser.go`, `syntax/lexer.go`, `check/check.go`, `check/scalar.go`, `ability/ability.go`, `check/check_modules_routes.go` and `surface/refusals.go`, with accepted complete-source witnesses, refused code/hint witnesses, explicit internal-branch dispositions and bidirectional controls. The seven-file census and finite input matrix do not prove arbitrary programs or branches outside that boundary (D-20..D-22).
4. **RESOLVED:** State the exact draft fixture provenance and capture its pre-edit baseline before parser changes. Plan [34-01](34-01-PLAN.md), Task 2 specifies twenty named fixtures labeled `reconstructed-2026-10-13`, including CRC-32 and word-counter drafts, literal code/span/ID/count observations and independent cause/sibling expectations. Its committed immutable baseline precedes parser/lexer/checker edits; Plans 34-04/05 write the separate after-state and Plan 34-08 checks exact authorized deltas. Missing original scratch drafts and approximate historical counts remain historical, not reconstructed-source measurements (D-09,D-12,D-14).

## Sources

### Primary (in-repository, current source inspected)
- `.planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-CONTEXT.md` — locked scope, compatibility, locations, recovery, refusal codes, manifest contract, seed dispositions, and test constraints.
- `.planning/REQUIREMENTS.md` and `.planning/ROADMAP.md` — DX-20..23, Phase 34 success criteria, Phase 35 dynamic-shift ownership.
- `internal/compiler/diagnostic/diagnostic.go`, `internal/compiler/protocol/protocol.go` — byte spans, nested causes/repairs, identity projection sites, finalization and output projection.
- `internal/compiler/syntax/parser.go`, `lexer.go`, `tree.go`; `internal/compiler/check/check.go`, `scalar.go` — parse gate, lossless source and local recovery/admission boundaries.
- `internal/compiler/modload/load.go`, `resolve.go`; `internal/compiler/session/session.go`; `cmd/schway/main.go` — exact source snapshots, relative path inventory, source-aware command path and CLI route.
- Existing test precedents: `parser_skipped_region_test.go`, `session_phase30_prechange_test.go`, `session_admission_divergence_test.go`, `main_phase32_routes_test.go`, and repair protocol tests.
- `AGENTS.md`, `.planning/STATE.md`, `.planning/PROJECT.md`, `.planning/PRODUCT-ROADMAP.md`, `.planning/LANGUAGE-MATURITY.md`, `.planning/sources/DASHBIT-AI-ERA.md`, and all three triggered dormant seed records — workflow, verification and transition context.
- Fresh focused behavior probes (2026-10-13): `/tmp/schway-phase34-bool-name.schway` accepted `true` parameter/`false` binding with status `pass`; `/tmp/schway-phase34-bool-tail.schway` produced only `name.unknown` on an unbound `true` tail. Command: `GOCACHE=/tmp/ai-lang-verification-gocache go run ./cmd/schway check <fixture> --json`.

### Primary documentation analogues (cited in phase context)
- Rust compiler JSON diagnostic format — https://doc.rust-lang.org/rustc/json.html
- Go token positions and parser recovery — https://go.dev/src/go/token/position.go and https://go.dev/src/go/parser/parser.go
- Clang diagnostics and error handling — https://clang.llvm.org/diagnostics.html and https://clang.llvm.org/docs/InternalsManual.html#handling-errors
- Rust compiler print options — https://doc.rust-lang.org/rustc/command-line-arguments/print-options.html

## Project Constraints (from AGENTS.md)

- Before file-changing work, begin with a GSD command; only use `$gsd-quick`, `$gsd-debug`, or `$gsd-execute-phase` for later non-research changes. This research invocation used `query init.phase-op 34` before writing. [VERIFIED: AGENTS.md GSD Workflow Enforcement]
- At planning transition, inspect living roadmap/maturity and only relevant seed triggers plus Dashbit appraisal; document scoped dispositions in the active planning artifact. All three dormant seeds remain dormant under the phase context's consumer thresholds. [CITED: AGENTS.md Living language roadmap; 34-CONTEXT.md D-25]
- Keep source witnesses, named checker guarantees, runnable gain, focused negative controls and the next three capabilities explicit; distinguish source inspection, fresh checks and historic receipts. [CITED: AGENTS.md Living language roadmap; .planning/PROJECT.md Verification Operating Preference]
- Prefer a runnable gain in a feature phase; an additional consecutive enabling-only phase triggers scope review. Do not invent percentages, require another backend, or introduce a general planning framework without a concrete consumer. [CITED: AGENTS.md Living language roadmap]

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — Go 1.24 is installed and the project explicitly mandates the Go standard library.
- Architecture: HIGH — source, identity, parser, module and CLI seams were read directly and match the locked context.
- Pitfalls: HIGH — identity hashing, parse gate, and recovery primitives are visible in source; contextual Bool-name witnesses were freshly executed.

**Research date:** 2026-10-13  
**Valid until:** 2026-11-12 (stable local architecture; recheck source if Phase 34 implementation or Phase 35 changes land first).
