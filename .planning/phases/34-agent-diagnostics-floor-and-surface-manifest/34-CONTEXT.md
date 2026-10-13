# Phase 34: Agent Diagnostics Floor and Surface Manifest - Context

**Gathered:** 2026-10-13
**Status:** Ready for planning
**Decision authority:** The user selected all relevant areas, requested role-based fan-out and adversarial review to diminishing returns, and explicitly delegated following the coherent recommendations. No interactive answers or approvals are invented. Three advisor agents completed initial and adversarial passes; the recovery advisor also checked the identifier-preservation edge.

<domain>
## Phase Boundary

Deliver DX-20, DX-21, DX-22 and DX-23 within ROADMAP Phase 34: locatable text/JSON diagnostics, one diagnostic per actual syntax recovery cause, precise refusals for seven currently absent forms, and a deterministic current-surface manifest. The runnable gains are `schway check` on a broken agent draft and `schway surface --json` without a source file.

This phase changes presentation and rejection quality, not admitted program semantics, core IR, runtime behavior or native lowering. Phase 35 owns bitwise operators, shifts, hex and the dynamic-shift refusal; Phase 36 owns scalar helper calls. Phase 38 owns final hosted receipts. The existing exit table, module-aware route boundary and repair authority remain in force.

</domain>

<decisions>
## Implementation Decisions

### Scope, dependencies and compatibility

- **D-01:** Preserve every previously admitted source program, its generated C bytes, diagnostic codes outside the specifically targeted refusals, byte coordinates, semantic diagnostic IDs and command result IDs under location-only enrichment. Capture literal pre-change goldens before compiler edits. Reducing cascading diagnostics may change the diagnostic set and therefore a failed command's ID; do not pretend those intentional changes are presentation-only. Pin the permitted code/count delta separately from the unchanged corpus.
- **D-02:** Use Go 1.24 standard-library and small local code. No parser replacement, Unicode-width dependency, terminal framework, schema generator, plugin system, LSP server or persistent compiler database. Sharing presentation mechanics or a code/hint catalog is acceptable; sharing the producer's admission conclusion with independent validation is not. A dependency requires a concrete failure of the local alternative, with its narrower boundary and benefit recorded.

### Source locations, identity and safe text

- **D-03:** `start` and `end` remain zero-based, half-open byte offsets and the sole edit coordinates. Added `line` and `col` denote the span's start, both 1-based; `col` counts Unicode scalar values in the original source, not UTF-8 bytes, UTF-16 units, graphemes or terminal cells. A tab counts as one scalar in the wire coordinate. Document this in the wire contract and manifest. Byte offsets already supply the alternative machine coordinate; do not add a second column field.
- **D-04:** Enrich every present, source-locatable diagnostic span: primary, cause and repair, using each span's effective module unit. Preserve nil spans. Handle UTF-8, CRLF, LF, empty source, last line without a newline, zero-width EOF and multi-line ranges. CRLF is one line break; both terminator bytes map to the preceding line's end position, and the next byte starts the next line. A lone CR is visible escaped content. Count each invalid UTF-8 byte as one position. An offset inside a valid rune maps to that rune's start for display, retaining the original byte offset. Reject invalid/inverted/out-of-source spans for presentation rather than silently clamping edit authority. Goldens must pin these conventions.
- **D-05:** Location fields and presentation context must never enter identity. `errorID` and `errorWithRepairsID` currently marshal complete spans and causes, so simply adding exported fields would be a regression. Explicitly preserve the old byte-only identity representation, including nested cause spans and unit identity, or use an equivalent wire-only projection. Keep frozen `lang.diagnostic/0`, `/1` and `lang.command/*` strings; this additive extension is not a coordinated rename/schema migration. Source-less legacy constructors must retain their old serialized form and pinned IDs. Test enrichment before and after `InUnit`, repeated enrichment, repair-kind identity and `Result.Finalize`.
- **D-06:** Compute locations and snippets from the exact bytes used in the current parse/check invocation. Reuse `syntax.Tree.Source`, loaded `UnitTrees` and the loader's source/relative-path inventory. No second read by pathname, directory scan, source lookup triggered by a diagnostic's untrusted text, or change to the loader's single-IO-site rule. Carry presentation provenance through the command result path without publishing all source bytes or including them in identity. This covers imported-file parser, checker and peer diagnostics as well as entry-file errors.
- **D-07:** Distinguish known zero-width source positions from the existing synthetic `Span{}` sentinel. Operational/usage/global diagnostics without a verified source position omit `line`/`col` and snippets; they must not acquire a fabricated `file:1:1`. A real start-of-file or empty-file syntax error is locatable. Preserve the diagnostic and its failure outcome when presentation data is unavailable. Use the CLI-supplied entry label and loader-confirmed relative imported paths; do not disclose a new canonical absolute path or change JSON's dotted `unit` into a filesystem path.
- **D-08:** Text source errors retain code and ID, add `file:line:col`, and show one source line plus a caret/range marker mapped through rendering. Use a deterministic, bounded ASCII rendering: ordinary ASCII remains readable, tabs expand at fixed stops, non-ASCII/invalid bytes and terminal control/bidi characters are visibly escaped. This trades raw Unicode appearance for exact alignment without a Unicode-width dependency. Escape hostile filenames and source-derived message text too; do not emit ANSI or read terminal width. Crop around the primary span with visible omission markers, keep at least one caret for insertion/EOF, and indicate a multi-line range without printing further source lines. Planner chooses and pins the tab stop, crop width and exact escape spelling. JSON gains coordinates, not snippets/source bodies; snippets remain presentation-only. Keep existing stream destinations and exit classes.

### Cause-oriented recovery and reproducible drafts

- **D-09:** Before parser/lexer changes, commit a deterministic agent-draft corpus and literal baseline diagnostic codes, spans and counts. The original CRC-32 and word-counter scratch sources were not found: `.planning/research/M007/` contains only `SUMMARY.md`. Reconstruct representative drafts, explicitly label their provenance, and do not call them recovered originals or reuse the summary's approximate counts as measured baseline. Include isolated one-cause drafts, both reconstructed multi-cause programs, nested malformed delimiters, sibling declarations, EOF, comments/strings and near misses. After repair, one diagnosed recovery cause gets one primary diagnostic while independent causes remain visible. Record exact before/after counts per fixture.
- **D-10:** Extend the existing parser with local context-aware recognizers and delimiter-aware synchronization. A damaged expression/list/block resumes at a valid sibling or caller-owned closer; it does not consume that closer or skip the remainder of the file. Do not globally stop after the first error or suppress every later `expected }`/`expected declaration`. Preserve real missing-closer errors that are independent of the original cause. Keep recovery progress checks and existing source/token/diagnostic/type bounds; new delimiter scanning must be iterative/bounded and may not introduce uncontrolled recursion.
- **D-11:** Lexer/parser coordination suppresses a generic unexpected-byte diagnostic only when the parser claims the exact offending token/span for a dedicated refusal or its single recovery cause. Do not filter by broad message/code prefixes, punctuation anywhere in the file, or all lexer failures. Comments, quoted strings, `->`, admitted punctuation and ownership keywords retain their existing meaning. Malformed near misses continue to get ordinary syntax errors. Lossless `Tree.Bytes()` must still reproduce rejected source, and recovered/rejected ASTs never pass the session parse-error gate into checking or native emission.
- **D-12:** Preserve the historic skipped-region cause/identity distinctions, including the spiral trio's different discarded spans and distinct diagnostic/result IDs. Test two independent mistakes surviving recovery and one-cause cascades disappearing. Keep a named overflow/truncation diagnostic when the existing diagnostic cap is reached; “one per cause” is not permission to hide that limit. Arm controls that restore cascading, swallow the next sibling or drop a separate cause, and prove the corresponding corpus assertion fails.

### Unsupported forms and hints

- **D-13:** Give each of the seven forms a dedicated stable code: `surface.unsupported_logical_and` (`&&`), `surface.unsupported_logical_or` (`||`), `surface.unsupported_logical_not` (`!`), `surface.unsupported_parameter_arity` (more than one declaration parameter), `surface.unsupported_if_without_else`, `surface.unsupported_bool_literal_tail`, and `surface.unsupported_subtraction`. Each has one concise hint that states an available alternative or the actual boundary. Hints must not promise syntax that checking refuses, silently change overflow/short-circuit behavior, or recommend an unchecked subtraction emulation. A case per form asserts precise primary span, code, one-line hint and absence of its previous misleading/cascading parse errors.
- **D-14:** Recognition is contextual. Excess parameters concerns Schway and foreign declarations; existing call-argument refusals retain `check.call_arity_unsupported`, and builtin `byte_at(view, index)` keeps its two-argument shape. Subtraction recognition must not capture arrows, strings/comments or unrelated malformed bytes. `true` and `false` remain identifiers, not new reserved keywords. Before implementing the Bool-tail refusal, pin previously admitted uses as parameter/binding names and reproduce the exact unbound tail shape. Classify a literal-shaped tail only with sufficient binding/result context to exclude a bound identifier; use the appropriate existing stage, not a spelling-only lexer/parser rule. This pre-edit witness is a required safety obligation, not an invitation to admit Bool literals or widen scalar result types.
- **D-15:** Keep hints in one-line diagnostic message text and expose the same literal hint in the manifest's refusal entry. No new automatic repair/replacement/applicability is inferred from a hint. Existing repair selection consumes structured byte edits, not prose, and cross-unit repair refusal stays intact. An optional dedicated diagnostic hint field is unnecessary for this phase. Tests independently pin the expected code and hint; sharing a catalog between emitters and the manifest does not supply the independent oracle.
- **D-16:** Do not add bitwise or hex unsupported hints in Phase 34. Do not claim a dynamic-shift refusal before its emitter exists. Phase 35 introduces BIT-02's refusal and extends the manifest introduced in Phase 34; REQUIREMENTS' reference to “the Phase 34 manifest” names that artifact, not early implementation. Phase 36 likewise changes the call rules when scalar helpers land. This resolves the advisor's apparent future-registration conflict without changing requirements ownership.

### Current capability manifest and independent drift proof

- **D-17:** `schway surface --json` emits a standalone `schway.surface/1` document, not a command envelope with timing/work metrics. Use fixed field order, canonical array order and one final newline. No source argument, source/filesystem scan, Clang probe, runtime launch, network, wall-clock timestamp, host-dependent paths or current Git hash. Successful output is exit 0/stdout; missing JSON mode, duplicate JSON flag, extra source/argument and unrelated flags are usage failures under the existing exit table. Handle the route before unrelated global flags can be stripped and accidentally accepted. Also support the existing global `--json surface` placement. Add the route to usage and the route inventory as `no-entry-source`; do not expand the module-aware set.
- **D-18:** The manifest describes the installed compiler's currently checked surface: types, operators, statements, builtins, call rules, contextual restrictions, limits and the implemented `surface.unsupported_*` code/hint entries. No scheduled capability/status, roadmap promise, parse-only construct presented as admitted, general String/record claim or unconditional “all functions compose” claim. An admitted item identifies the contexts where it works. Existing one-parameter declarations and restricted linear calls are separate from scalar expression-call support, which is absent until Phase 36. Each builtin declares its own arity/input/result/context. Discovery does not replace `check` on a complete program or independent peer admission.
- **D-19:** Limits name their scope, unit and meaning: source/token/declaration and module-loader limits, deterministic analysis work where relevant, file/output/evidence caps by applicable route, and call/type restrictions. Distinguish semantic constraints, compiler resource refusal and runtime/tool route budgets. Expose current known facts; do not imply that every number is a language-semantic limit or derive target facts from the current Apple host. Use existing constants when emitting values; independently pinned tests catch value/scope/unit drift. Do not create a new limits framework or enumerate incidental private tuning constants without an authoring consumer.
- **D-20:** An explicit small typed manifest is preferred to reflection/parser-generated claims. Keep an independently authored admitted/refused witness matrix and a bounded source census of relevant parser/checker operator, statement, builtin, type and refusal branches. Compare census/witnesses to the manifest and the manifest to direct checker/command behavior. Every admitted entry needs a concrete accepted source witness in its stated context; every refusal needs its rejected source witness and expected code/hint. The census must fail on an unaccounted newly introduced surface branch/code, and the reverse direction must fail on a manifest-only claim. A manifest registry feeding both sides is not independence. State the finite corpus/census coverage limit; no proof over every possible program is claimed.
- **D-21:** Arm each drift direction: remove an admitted entry, add an unadmitted claim, remove a refusal/code or hint, and weaken a context/arity/limit restriction without changing its name. Pin `byte_at` arity 2 separately from ordinary calls; pin absent scalar expression calls; include all seven current refusal codes. Restoration of each control is required, and each must fail for its named reason rather than an unrelated fixture break. These tests extend when Phase 35/36 adds operators/refusals/calls. Static inspection alone is insufficient to substantiate a positive admission claim.

### Evidence lanes and planning handoff

- **D-22:** Pin generated C for the existing admitted corpus before implementation, following the existing pre-change golden/armed-control pattern. The planner must enumerate the corpus and explicit non-emitting refusal cases, not silently drop a program because generation failed. After implementation, every previously emitting program matches byte-for-byte; a reached emission mutation proves the comparison is live. No new backend or semantic peer rule is warranted by a presentation phase. Maintain independently admitted safety/refusal tests beside the cosmetic golden.
- **D-23:** Cover presentation through public `check` and every existing diagnostic-bearing source route affected by the shared path, including imported errors and non-source operational errors. JSON compatibility tests decode enriched output with existing repair consumers and show that eligible repairs, exact byte ranges and replacement behavior are unchanged. Preserve `TestRepairDriverDecodesNoProseFields`, cross-unit decline and module-aware/single-file route classifications. Do not make evidence/interface/query newly module-aware just to obtain snippets. Avoid rerunning source checks merely to render an error.
- **D-24:** Shift acceptance into focused table-driven location, rendering, recovery, refusal, manifest and CLI tests with explicit negative controls. Preserve legacy identity, module attribution and repair tests. Use only focused local selectors with `GOCACHE=/tmp/ai-lang-verification-gocache`; STATE forbids local project suites. Existing CI owns broader checks; Phase 38 owns final hosted receipt binding. No full suite, native matrix or new hosted receipt was run during this discussion. SUMMARY coverage entries must name real tests and use the closed coverage-kind vocabulary. Aim for zero human UAT because these contracts have deterministic proxies; do not claim human readability or latency improvement without measurements.
- **D-25:** Retain the three dormant seeds with scoped dispositions below. No measured named caller-query gap, unexplained runtime failure or disputed canonical-source form was established here. Keep the source appraisal's pre-Phase-33 capture-allocation observation historical; Phase 33's bounded read and reached mutant are its resolution. Current top-three capability order remains diagnostics/manifest, bitwise computation, scalar helpers. No change to active roadmap/requirements ownership is requested.
- **D-26:** Plan approximately eight coherent slices at most unless a named runnable witness or safety obligation justifies growth. Source snapshot/identity preservation, contextual Bool-name preservation, finite recovery termination and bidirectional manifest controls are concrete safety obligations; do not split every role/construct into its own plan. Keep a runnable manifest and broken-draft `check` in this phase so it does not become a second purely enabling phase after 33. Planning may refine helper layout, test placement and rendering constants without reopening these contracts. The user delegated recommendations; the persistent GSD auto-advance preference governs the subsequent plan/execute transition.

### Planner's Discretion

- Local helper/package boundaries, test-file placement and dependency-aware plan/wave split.
- Exact text layout, ASCII escape spelling, tab stop and crop width, once documented and golden-pinned within D-08.
- Hint prose within D-13/D-15, with an independently tested available alternative and one-line requirement.
- Position-presence representation and presentation-context plumbing, while preserving old bytes/IDs, source snapshots, unit attribution and unavailable-location honesty.
- Narrow recognition/recovery implementation, with D-14's binding-preservation witness before choosing a stage.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Scope, prior decisions and operational boundaries

- `.planning/ROADMAP.md` §Phase 34 and Phases 35–38 ordering.
- `.planning/REQUIREMENTS.md` DX-20..DX-23 and dynamic-shift ownership note.
- `.planning/PROJECT.md` current M007, verification preference and dependency constraints.
- `.planning/STATE.md` current position, focused-only local checks and cache path.
- `AGENTS.md` workflow enforcement and living-roadmap transition checks.
- `.planning/research/M007/SUMMARY.md` scratch-authoring account; original drafts are unavailable.
- `.planning/phases/33-carried-debt-bounded-capture-read-and-generic-freeze-gate/33-CONTEXT.md` receipt/debt and independent evidence boundaries.
- `.planning/phases/33-carried-debt-bounded-capture-read-and-generic-freeze-gate/33-TRANSITION-CHECK.md` carried evidence and current capabilities.
- `.planning/milestones/M006-phases/32.1-address-tech-debt-cli-exit-classes-and-module-aware-routes/32.1-CONTEXT.md` exit classes, module route classifications and structured repair boundary; Phase 33 supersedes its receipt/predicate debt.
- `.planning/milestones/M006-phases/32-clean-checkout-guide-and-final-receipts/32-CONTEXT.md` pre-change golden/public route precedent; receipts remain historical.

### Source, identity, recovery and independent evidence

- `internal/compiler/diagnostic/diagnostic.go` and `internal/compiler/diagnostic/diagnostic_test.go` spans, nested causes, repairs and frozen identity.
- `internal/compiler/diagnostic/diagnostic_phase30_test.go` unit identity and omitted entry unit.
- `internal/compiler/protocol/protocol.go`, `internal/compiler/protocol/protocol_test.go`, `internal/compiler/protocol/protocol_phase30_test.go` projections, finalization and schema compatibility.
- `internal/compiler/syntax/parser.go`, `internal/compiler/syntax/lexer.go`, `internal/compiler/syntax/token.go`, `internal/compiler/syntax/tree.go` bounded parsing and exact source bytes.
- `internal/compiler/syntax/parser_skipped_region_test.go` and `testdata/distinctness/` historic skipped-region/spiral identity witnesses.
- `internal/compiler/check/check.go` and `internal/compiler/check/scalar.go` checked admission versus parsed syntax and call context.
- `internal/compiler/modload/load.go`, `internal/compiler/modload/resolve.go` source snapshots, relative path provenance and loader bounds.
- `internal/compiler/session/session.go`, `internal/compiler/session/session_phase30_modules_test.go`, `internal/compiler/session/session_phase30_inventory_test.go` source command plumbing and single IO site.
- `internal/compiler/session/session_phase30_prechange_test.go` frozen C comparison and reached control precedent.
- `internal/compiler/session/session_admission_divergence_test.go` admission census and explicitly accounted divergences.
- `cmd/schway/main.go`, `cmd/schway/main_phase32_routes_test.go` CLI dispatch, global flags, usage and public route inventory.
- `cmd/schway-repair/repair.go`, `cmd/schway-repair/repair_test.go`, `cmd/schway-repair/repair_phase30_test.go` byte-only patch authority, prose independence and cross-unit decline.

### Living roadmap and conditional seeds

- `.planning/PRODUCT-ROADMAP.md` and `.planning/LANGUAGE-MATURITY.md` current dated Phase 33 amendments; earlier sections are historical.
- `.planning/sources/DASHBIT-AI-ERA.md` source-bound appraisal, distinct evidence labels and narrow consumer triggers.
- `.planning/seeds/SEED-261008-z1g-bounded-symbol-and-caller-query.md`.
- `.planning/seeds/SEED-261008-trz-consumer-triggered-failure-observability.md`.
- `.planning/seeds/SEED-261008-z1s-compare-contract-and-source-forms.md`.

### Primary-source analogues consulted (2026-10-13)

- [rustc JSON diagnostics](https://doc.rust-lang.org/rustc/json.html): byte offsets coexist with scalar-value coordinates and optional location/context. This supports the coordinate distinction, not Schway conformance.
- [Go position source](https://go.dev/src/go/token/position.go): byte columns are a viable contrasting convention; Schway deliberately documents scalar columns.
- [Go parser source](https://go.dev/src/go/parser/parser.go): local synchronization and progress control are relevant patterns, not a proposal to copy the complete parser.
- [Clang expressive diagnostics](https://clang.llvm.org/diagnostics.html): source location, snippet and caret presentation.
- [Clang recovery/error architecture](https://clang.llvm.org/docs/InternalsManual.html#handling-errors): a rich recovery AST is a larger alternative, deferred here.
- [rustc print options](https://doc.rust-lang.org/rustc/command-line-arguments/print-options.html): source-free compiler discovery command as a narrow analogue.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- `syntax.Tree.Source` retains exact parsed bytes; `Tree.Bytes()` reconstructs lossless tokens/trivia, including rejected text.
- `CheckResult.UnitTrees` and `modload.Set.Sources` retain module-local bytes; `modload.Source.RelPath` is loader-confirmed relative provenance. Current command projections discard this context, so retaining it is the main integration work.
- `Cause.EffectiveUnit` / `Repair.EffectiveUnit` already define span-unit inheritance.
- Parser `recoverUntil`, `recoverRegion` and `assertProgressOrBoundary` offer local bounded recovery. The code/hint recognizers must preserve delimiter ownership and skipped extents.
- Existing literal golden, seeded mutation and closed route/coverage inventory tests provide standard-library evidence patterns.

### Established Patterns

- `diagnostic.errorID` and `errorWithRepairsID` serialize full spans/causes. `Result.Finalize` consumes diagnostic IDs rather than message/location prose. Enriching a span without isolating identity is unsafe.
- `session.Check` returns on parse diagnostics before checker admission. Parsed nodes alone are not admitted language facts.
- `true`/`false` are identifier spellings. Scalar Bool state exists, but scalar result/context restrictions must be witnessed rather than inferred from the AST.
- Ordinary declarations have one parameter; ordinary language-call admission and scalar-call admission differ; builtin `byte_at` has two inputs.
- Current module-aware set is check/run/build/app verify/debug-map. A source-free manifest does not justify changing it.

### Integration Points

- Source-aware command construction in session, then `protocol.JSON` / `Human` and CLI `emit`; avoid a second check/read for rendering.
- Dedicated refusal recognition in syntax/check where context is available, with one-line hints and lossless recovery.
- A source-free `surface` dispatch before unrelated flag extraction, plus usage and the public route audit.
- Independent surface census/witness tests and pre-change C/identity goldens before implementation.

</code_context>

<specifics>
## Specific Ideas and Transition Check

The user's preference is “another copy and paste is better than another dep,” with a dependency allowed when a concrete need earns it. The role lenses applied were agent author/repair consumer, human auditor, compiler/language architect, CLI/accessibility design, security/untrusted input, compatibility, QA/mutation, and CI/platform/feedback cost. Graphics, distributed-service and unrelated framework lenses have no concrete consumer in this phase.

**Evidence distinction:** This discussion inspected current source at `a18a956ced12817e48c30605d411cdeb9547bfa6` and consulted primary documentation. It did not execute compiler behavior checks or measure latency/readability. Phase 33's passed checks and archived M006 hosted runs are historical receipts at their recorded revisions. All new phase-34 test obligations above remain implementation work.

### Scoped seed dispositions

| Seed | Current source witness / named consumer | Disposition | Observation that would change it |
|------|----------------------------------------|-------------|---------------------------------|
| SEED-261008-z1g | `session_phase6_query.go`, existing bounded query and ordinary search; Phase 36 CRC helper review is a possible later consumer | retain dormant | A measured exact-answer or material whole-loop cost gap for one named helper/callsite task; a surface manifest is not that query experiment |
| SEED-261008-trz | Phase 33 `readApplicationCapture` and `TestPhase33CaptureReadBounded`; checksum failure capture/replay already has a diagnosis path | retain dormant | A named runtime failure unexplained after capture/observers/debug map/replay; syntax drafting errors do not trigger runtime telemetry |
| SEED-261008-z1s | `check/scalar.go`, canonical source and current CRC drafts; this phase preserves grammar admission | retain dormant | An actual reviewer/consumer contests a concrete contract/inference/source form with a measurable guarantee or review question |

The Dashbit appraisal's allocation concern is addressed by Phase 33, as its transition check records. Do not revise the historical source review into a new execution claim. No living capability fact changed in this discussion, so the current dated living-document amendments remain accurate; this context records the new planning decisions and seed audit.

### Three most useful next capabilities

| User-visible program / capability | Current blocker | Smallest complete slice | Checker / guarantee changes | Evidence/debt references | Owner / next action | Priority-changing observation |
|-----------------------------------|-----------------|-------------------------|-----------------------------|--------------------------|---------------------|-------------------------------|
| Broken CRC/word-counter draft `schway check`; `schway surface --json` | Offset-only diagnostics, cascading recovery, absent truthful manifest | Complete Phase 34's source-snapshot locations, seven refusals, bounded recovery and source-free manifest | Preserve byte repair authority/identity and admitted semantics; validate current facts independently | DX-20..23; skipped-region/unit/repair goldens; reconstructed drafts need a measured baseline | Phase 34 planner/executors; capture baseline first | A required rendering/recovery design changes admitted programs or loses a cause: narrow/redesign before proceeding |
| A scalar expression `x ^ (y << 3) & 0xFF`; Adler-32 with `b << 16` | Bitwise/shift/hex absence | Phase 35 defined U64 operators, literal amounts below 64 and independent pinned vectors; CRC consumer follows in 37 | Six affected consumers derive operations independently; named dynamic/range refusals prevent C shift UB; update manifest | BIT-01..04; existing Adler-32 answers and no-op C goldens; EVD-17 in 38 | Phase 35 after 34 | CRC needs a non-literal shift or other semantics: revisit the selected algorithm before widening |
| Imported `crc_step` helper used in a scalar loop | Scalar-to-scalar expression calls absent | Phase 36 arity-1 helpers with returned scalar and callee-defect propagation; integrate file CRC in 37 | Independent argument/result/defect derivation, no recursion and unchanged ownership boundary; update manifest | CALL-01/02; APP-13/14 and MOD-07; existing cycle and defect refusals | Phase 36, then 37 integration | A second named consumer requires arity-N/records/owned calls: re-scope through milestone ownership, not this diagnostics phase |

The first slice remains runnable; another enabling-only phase would require scope review. No language-completeness percentage, new backend or general planning framework follows from this review.

</specifics>

<deferred>
## Deferred Ideas

- Full recovery AST/IDE-quality malformed trees, display-width Unicode library, LSP/editor adapter and persistent program database: no current consumer justification.
- Scheduled-capability negotiation, schema/plugin framework, broad parser rewrite and automatic fix suggestions: unnecessary to discover current checked surface.
- Bitwise/hex/shift semantics and dynamic-shift refusal: Phase 35; scalar helper calls: Phase 36; interpreter file-input CLI route and executable CRC guide: Phase 37; hosted final binding: Phase 38.
- Arity-N, records, general Bool literals, subtraction semantics, module-aware evidence/interface/query and foreign-adapter simplification: retain their existing later ownership.
- Runtime telemetry, source-form experiments and named caller queries: dormant seed conditions above.

</deferred>

---

*Phase: 34-agent-diagnostics-floor-and-surface-manifest*
*Context gathered: 2026-10-13*
