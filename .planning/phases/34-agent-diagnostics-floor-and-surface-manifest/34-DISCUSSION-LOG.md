# Phase 34: Agent Diagnostics Floor and Surface Manifest - Discussion Log

> **Audit trail only.** Do not use as input to planning, research or execution agents. Decisions are captured in CONTEXT.md; this log preserves alternatives and delegated selections.

**Date:** 2026-10-13
**Phase:** 34-agent-diagnostics-floor-and-surface-manifest
**Areas discussed:** Source positions/compatible presentation; cause-oriented recovery; precise unsupported forms; truthful current manifest; independent evidence and transition boundaries.

## User direction and method

The user invoked `$gsd-discuss-phase 34`, requested breadth/depth fan-out through relevant roles with primary-source research, pros/cons, footguns and an adversarial pass, and said to synthesize coherent recommendations and “auto follow those.” This explicitly delegates choices and artifact creation within discussion; no question responses are fabricated. Their flat-dependency preference supplements the existing Go-standard-library constraint.

Three typed `gsd-advisor-researcher` agents researched disjoint areas, each with an initial comparison and adversarial follow-up. A final recovery follow-up examined Bool-like identifier names. Models/effort came from the GSD resolver. The root reconciled scope, source witnesses, recommendations and primary references. No advisor edited files or ran project suites. Discussion pre-hooks were empty; no prior phase-34 context/spec/checkpoint/plans existed, and no pending todo matched. No codebase maps/findings skills were found, so scouting used the relevant source paths. The latest three contexts consulted were 33, 32.1 and 32, with later dated amendments taking precedence over older debt/receipt choices.

The lens set was compiler/language architecture, agent authoring/repair, human audit, CLI/accessibility design, security, compatibility, QA/mutation, and CI/platform/feedback cost. Unrelated graphics/distributed-system/framework roles would not change these decisions and were not added for volume. Stop condition: a further lens or source must plausibly change a scoped contract or concrete test obligation. After the identifier counterexample, no remaining issue justified another broad round.

## Source coordinates and text presentation

| Option | Pros | Cons / footguns | Selected |
|--------|------|-----------------|----------|
| Canonical byte offsets plus 1-based scalar start line/col | Human-oriented UTF-8 location, old patch coordinates remain exact | Scalar count differs from display width; nested spans and identity need explicit projection | yes |
| UTF-8 byte columns | Direct offset mapping and Go convention | Looks shifted after multibyte characters; easy for clients to mistake for character units | |
| Two column fields | Both coordinate conventions available | Redundant with existing byte offsets; more consumer ambiguity and schema surface | |

**Delegated selection:** Scalar columns and unchanged byte authority (D-03..07). The code hashes complete spans, so raw field addition would silently change IDs. Loaded module source snapshots and relative paths must survive projection; a later file read could misattribute a caret to edited source. Synthetic `{0,0}` is not automatically an actual source position.

| Rendering option | Pros | Cons / footguns | Selected |
|------------------|------|-----------------|----------|
| Bounded escaped ASCII snippet with mapped caret | Deterministic cells, terminal-control safety, no width dependency | Non-ASCII appearance is escaped, though original-source scalar columns remain exact | yes |
| Raw Unicode and guessed rune width | Familiar appearance | Combining/wide/bidi text breaks caret alignment and can spoof output | |
| Full terminal Unicode-width/grapheme framework | Rich raw-text rendering | New dependency/table maintenance and terminal conventions without a current consumer | |

**Adversarial refinement:** Escape filenames/message inputs too; handle CRLF, invalid bytes, EOF/insertion and multi-line ranges; crop visibly, not silently. This is a deterministic contract, not evidence that human reviewers prefer escaped source. Width/tab/escape constants remain planner discretion. [rustc JSON](https://doc.rust-lang.org/rustc/json.html) provides the byte/scalar distinction; [Go positions](https://go.dev/src/go/token/position.go) supply a contrasting byte-column convention; [Clang diagnostics](https://clang.llvm.org/diagnostics.html) illustrate source/caret presentation.

## Recovery and unsupported constructs

| Option | Pros | Cons / footguns | Selected |
|--------|------|-----------------|----------|
| Local recognizers and delimiter-aware synchronization | Precise cause, sibling errors remain visible, reuses existing bounded parser | Must preserve caller-owned closers and distinguish malformed near misses | yes |
| Stop/suppress after first failure | Small visible-count reduction | Hides separate errors and leaves misleading first diagnostic | |
| General recovery AST/parser redesign | Rich malformed structure for editor consumers | Broad AST/formatter/semantic risk beyond this diagnostics floor | |

**Delegated selection:** Local recovery (D-09..12), not a count-only suppression rule. Preserve lossless bytes, progress/caps, skipped-region extents and the session error gate. [Go parser source](https://go.dev/src/go/parser/parser.go) supports narrow synchronization; [Clang recovery architecture](https://clang.llvm.org/docs/InternalsManual.html#handling-errors) explains the larger alternative, not a requirement to adopt it.

| Refusal choice | Pros | Cons / footguns | Selected |
|----------------|------|-----------------|----------|
| Seven distinct contextual codes and one-line hints | Agent can identify exact unsupported form; manifest can witness each refusal | Overbroad recognition can change legal identifier meaning or swallow syntax errors | yes |
| One unsupported-family code for every form | Fewer names | Loses actionable distinctions and one-to-one negative controls | |
| Spelling/global punctuation scan | Easy to enumerate | Hits comments/strings/arrows, confuses `byte_at` builtin with ordinary arity, reserves true/false accidentally | |

**Adversarial refinement:** `true`/`false` are identifiers, not keywords. The final follow-up found a plausible existing parameter/binding-name counterexample. D-14 requires a pre-edit admitted witness and sufficient binding/result context before classifying a literal tail. Ordinary call arity already has a checker refusal; declaration arity is the targeted parser case. Foreign declarations need the same precision; builtin `byte_at` keeps two inputs. Hints stay in message prose; no automatic repair authority follows.

The original scratch CRC/word-counter sources were not found. `.planning/research/M007/SUMMARY.md` records an approximate drafting observation, not repeatable literal inputs/counts. D-09 therefore requires labeled reconstructed drafts and a captured pre-change baseline; it does not launder the approximate history into current evidence.

## Manifest and independent tests

| Option | Pros | Cons / footguns | Selected |
|--------|------|-----------------|----------|
| Explicit current-facts manifest plus independent census/witnesses | Small local contract, contextual truth, omissions/false claims can fail in both directions | Curated finite evidence needs maintenance; no universal completeness proof | yes |
| Generate from parser/checker registry/reflection | Lists track declarations mechanically | Parsed syntax differs from checked/runnable surface; shared registry tests can be tautologies | |
| Capability/plugin framework with scheduled states | Rich negotiation for hypothetical external consumers | Adds lifecycle/dependency surface and mixes roadmap promises into installed compiler facts | |

**Delegated selection:** Standalone `schway.surface/1` JSON; current checked facts and precise limits/context, no timings or future states (D-17..21). [rustc print options](https://doc.rust-lang.org/rustc/command-line-arguments/print-options.html) are a narrow discovery-command analogue. The CLI route is source-free and must reject unrelated global flags before extraction can erase them. No JSON-mode-free human manifest is needed in this phase; a usage response is explicit and bounded.

**Adversarial refinement:** A blanket “calls are arity 1” claim is false for builtin `byte_at`; linear language calls do not establish scalar expression calls. AST types do not establish general records/String support. Tests need accepted/refused concrete contexts and source-census accounting, not two copies of the manifest list. Controls must reach missing admission, invented admission, missing refusal/hint and weakened constraints.

The advisor flagged an apparent future-shift registration conflict. Root resolved it against ROADMAP: the dynamic-shift code lands in Phase 35 and is then registered in the manifest introduced in 34. No requirement amendment or scheduled runtime claim is needed (D-16).

## Evidence, dependency and scope decisions

| Choice | Benefit | Tradeoff / control | Selected |
|--------|---------|--------------------|----------|
| Standard-library/local implementation | Shallow audit boundary and fewer abstractions | Must explicitly own coordinate/recovery/manifest cases; no unproven width claim | yes |
| New libraries/frameworks now | Could supply mature broad facilities | No named gap earning their dependency surface | |
| Pre-change C, code/count and identity goldens plus reached controls | Proves no semantic widening and useful rejection changes separately | Corpus must enumerate exclusions honestly; golden equality alone does not prove admission safety | yes |
| Broad local suites/host receipts during discussion | Wider observations | Contrary to STATE's local-suite rule and Phase 38's final hosted ownership | |

**Delegated selection:** D-22..26, focused tests during implementation and zero invented current passes. Preserve existing legacy IDs, unit attribution, repair-prose independence and route inventory. Approximately eight plans, with each extra plan justified by a runnable witness or named safety obligation rather than role count. SUMMARY coverage must name real tests with a valid kind.

## Transition review and discretion

Living PRODUCT-ROADMAP/LANGUAGE-MATURITY dated Phase 33 claims still match the inspected frontier. The three relevant seeds retain dormant status: named symbol/caller query lacks a measured gap, runtime diagnosis lacks an unexplained failure, and source-form comparison lacks a contested reviewer choice. The Phase 33 allocation repair resolves the historical appraisal's concrete concern; syntax feedback does not activate telemetry.

Current next capabilities remain Phase 34 diagnostics/manifest, Phase 35 defined bitwise/shift/hex computation, and Phase 36 scalar helpers for Phase 37's CRC consumer. CONTEXT's table names programs, blockers, complete slices, checker obligations, evidence owners and observations changing priority.

Root/planner discretion covers helper layout, tests/waves, deterministic rendering constants and hint wording within pinned safe contracts. No further UI/editor/runtime/framework expansion is selected. GSD's persistent `workflow.auto_advance: true` was read and canonical `check auto-mode` returned true; after committed context the skill routes into planning/execution. No flags or configuration preference were invented or enabled by this discussion.

## Deferred Ideas

Full recovery trees/IDE integration, raw Unicode display-width framework, LSP/query database, scheduled negotiation, automatic repairs and new semantic surface remain deferred. Phase 35/36/37/38 retain their named ownership. All three conditional source-appraisal experiments remain dormant; broader runtime observation and alternative canonical syntax are not phase-34 work.
