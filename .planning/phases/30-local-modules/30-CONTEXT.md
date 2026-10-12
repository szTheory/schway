# Phase 30: Local Modules - Context

**Gathered:** 2026-10-12
**Status:** Ready for planning

<domain>
## Phase Boundary

A Schway program can import another local source file as a module, using an
explicit `import` declaration and the callee module's `export` list. Resolution
is closed under the entry file's directory. There is no search path, and `..`,
absolute paths and symlinks are all refused. Only exported names are usable,
and corevalidate re-derives visibility independently. Cross-module calls are
admitted from signatures alone. The following are each refused with their own
named diagnostic, attributed to the source that caused them:
- self-import, length-2 and length-3 cycles;
- a missing module;
- a duplicate module;
- a header/path mismatch;
- a name collision.

A diamond import loads and emits the shared module once. Every program without
imports keeps byte-identical generated C, diagnostics and evidence. The
runnable gain: the Phase 29 Adler-32 program runs with the library in its own
module file, imported rather than copied, and prints the same pinned answers on
both engines.

Not in this phase:
- the second consumer, the dual-flip library mutation, and sorted module
  hashing in build identity and evidence for MOD-06 (Phase 31);
- the file checksum app (Phase 31);
- the guide (Phase 32).

Modules are whole-program composition. Never describe them as separate
compilation or as closing B1/DX-06.

</domain>

<decisions>
## Implementation Decisions

All decisions below come from advisor research, with one researcher per gray
area at the `minimal_decisive` tier. They were **auto-followed** per the user's
standing preference to fan out research and then follow its recommendations.
Conflicts between researchers were resolved in favor of
`.planning/research/ARCHITECTURE.md` §5.1 where it already decided the point.

M006 research Decision 7 (`.planning/research/SUMMARY.md`) remains settled and
is not repeated here:
- file = module;
- explicit imports, with no glob, alias or re-export;
- collisions are refused;
- the export list is the public surface, and corevalidate re-derives it;
- cycles are refused by name;
- diamonds load once;
- whole-program composition into one `ast.Program`;
- cross-module admission is body-blind, through the signature table;
- injective C names;
- additive, omitempty evidence fields.

### Import surface and formatter
- **D-01:** An import is a named import block, and calls stay bare:
  ```
  module examples.phase30.adler32_cli

  import adler32 {
    fn adler32
  }

  export {
    fn main
  }
  ```
  - The call stays `let sum = adler32(bytes)`, so `ast.RHS.Callee` remains a
    bare string and the six string-keyed consumers keep their shape.
  - There is no qualified `m.f(x)` call form in M006. It would need
    dotted-callee changes in three call parsers, and with no aliases it buys
    nothing.
  - The hybrid of the two forms is rejected.

  — **Reversibility:** costly — the surface gets pinned by fixtures,
  formatter goldens and repairs. A qualified form could still be added later
  as an alternative spelling without breaking this one.
- **D-02:** Grammar:
  ```
  program    := "module" modulePath importDecl* exportBlock decl*
  importDecl := "import" modulePath "{" ("fn" IDENT)+ "}"
  ```
  - Items are separated by whitespace or newlines, with no commas. This copies
    the `exports()` loop at `parser.go:266-286`, which the parser may reuse.
  - `import` becomes a keyword through a new `TokenImport`.
  - An empty list is refused with `syntax.expected_import_name`.
  - An `import` placed after `export` is refused with
    `syntax.import_after_export` rather than falling into generic declaration
    recovery.
  - `type` items are refused by check with a named code. Exported
    types and data across modules are out of M006 scope.
- **D-03:** The formatter keeps source order and does not sort imports, because
  semantic tokens never move (`format.go:8`).
  - One module goes in each `import` declaration, with one item per line. The
    layout reuses the export-block context and `blankBeforeTopLevel`.
  - The idempotence and lossless lanes are extended to cover `import`.
  - Identity is not affected by source order, since module hashing is sorted
    (Phase 31).
- **D-04:** Collision refusals cover every pair of sources for a name:
  - import vs local `fn`;
  - import vs import;
  - imported `main` vs the entry's `main`;
  - an import of a reserved builtin name (`byte_len`, `byte_at`, `write_*`).

  A collision diagnostic names both origins. It has no automatic repair,
  because renaming is a semantic choice and there is no alias to offer.

  Deterministic repairs exist for exactly two refusals:
  - **duplicate import of the same module:** delete the later one.
  - **import after export:** move the import above `export`.

  `module.not_exported` gets a hint only, because the fix is an edit to another
  file. An unresolved bare call never synthesizes an import. There is no
  unused-import refusal.

### Name ↔ path resolution
- **D-05:** The root is the entry file's directory, opened with
  `os.OpenRoot`. Module `a.b.c` resolves to `<root>/a/b/c.schway`.
  - This supersedes the "strip the entry's module suffix" rule in
    `STACK.md` Decision 5 (ARCHITECTURE §5.1 already chose the entry directory).
  - There is no project marker, no upward search and no environment root.
  - The resolver appends the `.schway` extension, which is exact and
    lowercase. Users never write an extension or a separator.
- **D-06:** The entry module's own header is **not** checked against its path.
  - Every existing program keeps its root-relative or phase-named header, such
    as `examples.phase29.adler32_fixture`, and no new check runs on programs
    without imports.
  - Imported modules' headers must equal the dotted name derived from their
    path, or they are refused with `module.path_mismatch`.
  - An imported module whose name equals the entry's header is refused with
    `module.duplicate`.

  — **Reversibility:** costly — tightening the entry check later would add a
  new refusal to existing programs.
- **D-07:** Segment charset for module paths in `import` and in imported
  headers:
  - each segment is `[a-z][a-z0-9_]*`;
  - each segment is at most 64 bytes;
  - ASCII only.

  Anything else is refused with `module.invalid_segment`. The lexer accepts
  Unicode letters (`lexer.go:137-141`), so the module-path check lives in
  modload, not the lexer. Lowercase-only removes case and homoglyph collisions
  up front.
- **D-08:** Symlinks are refused, including symlinks that stay inside the root.
  - Each path component is checked with `Root.Lstat`. Any `ModeSymlink` is
    refused with `module.symlink`.
  - `os.Root` remains a second layer against escape. On escape it reports
    `path escapes from parent`, which maps to `module.escape`.
  - Reason: `os.Root` silently follows in-root links. That aliases two names to
    one file and breaks duplicate detection and identity.
  - Verified empirically on Go 1.24.0 darwin/arm64:
    - an in-root symlink opens;
    - escapes (an escaping symlink, an absolute target, `..`, or `/etc/hosts`)
      fail.
- **D-09:** Exact case is verified on every platform.
  - Run `ReadDir` on the parent through the root, and require an entry whose
    name is byte-equal to the expected file name.
  - On APFS, `os.Root` opens case variants and `Lstat().Name()` echoes the
    requested name, so neither can verify case.
  - A miss is refused with `module.not_found`, or with `module.case_mismatch`
    when an `EqualFold` match exists.
  - Tests must not depend on macOS-only behavior. On Linux, create both
    spellings to exercise the check.
- **D-10:** File kind, TOCTOU and the single read:
  1. Use `Root.Lstat` and require `Mode().IsRegular()` **before** `Open`. On
     macOS, `Open` on a FIFO blocks for more than 2 s. Do not copy
     `readBindingFile`, which opens and then stats.
  2. After the open, `Stat` and require `os.SameFile` with the Lstat result.
  3. Read once with a limit+1 bounded read, reusing the `readBoundedFile` idiom
     at `session.go:1890`.
  4. A session-side `SourceSet` (modload) is the **only** IO site. Parse,
     check, hashing and identity all use those in-memory bytes, and the loader
     never re-reads a file.
  5. Diagnostics and records carry root-relative names only.

### Diagnostic attribution and C names
- **D-11:** `Unit` holds the **module name** and is omitempty. An empty `Unit`
  means the entry module.
  - It is added to `Diagnostic`, `Cause`, `Repair` and `debugmap.Entry`.
  - A `Cause` or `Repair` with an empty `Unit` inherits the diagnostic's
    `Unit`. It is set explicitly only for a span in another file, such as an
    export-list hint.
  - `Span` stays file-local. There is no `Span.File` field and no
    virtual-concatenated offsets.
  - `Unit` is added as `json:",omitempty"` to the anonymous identity structs in
    `diagnostic.Error` and `ErrorWithRepairs`. Existing IDs are then unchanged,
    and same-offset diagnostics in different files get distinct IDs.
  - The module-to-path mapping comes from the loader's module table, which
    evidence surfaces as `Program.Modules[]` in Phase 31.
  - The human projection shows the unit only when `Unit` is set, as in-line
    `unit=<module>` (form fixed by D-23).

  — **Reversibility:** one-way — `unit` becomes part of the published
  diagnostic protocol that agents consume.
- **D-12:** Imported functions use length-prefixed C names inside the
  `SCHWAY_` namespace:
  - The form is `SCHWAY_M_<len>_<SEG>_..._<len>_<FN>`, where each length counts
    source bytes before uppercasing. Example: `adler32.adler32` becomes
    `SCHWAY_M_7_ADLER32_7_ADLER32`.
  - Entry-module functions keep today's `cName(function.Name)` and their
    current ordinal, so single-module C stays byte-identical.
  - Imported functions are allocated after the entry functions, in canonical
    order sorted by (module, function). The ordinal is their position in that
    order, so a reversed import order produces identical C and a diamond is
    emitted once.
  - Entry names win a contested name. A colliding imported name gets the
    existing deterministic `allocate` suffix.
  - The `^SCHWAY_[A-Z0-9_]*$` invariant is kept, along with no new `__` and no
    leading `_`.
  - Rejected alternatives:
    - `_S…`, Itanium or v0-style names (C17 7.1.3 reserves them);
    - `module__name` (reserved `__`);
    - numeric module indexes (unstable).
  - The planner must confirm that cgen receives (module, bare name) separately,
    not a pre-joined string.

### Resource caps
- **D-13:** Caps are taken from ARCHITECTURE §5.1, which supersedes the
  STACK.md 64 KiB per-file suggestion:

  | Cap | Accepted N | Refused N+1 | Code |
  |---|---|---|---|
  | Per-module source bytes | 1,048,576 (`syntax.MaxSourceBytes`) | 1,048,577 | existing `syntax.input_limit`, with `Unit` naming the file. Do not add a second size number. |
  | Distinct modules, including the entry (a diamond counts once) | 16 | 17 | `module.count_limit` |
  | Total source bytes read | 4,194,304 | 4,194,305 | `module.total_bytes_limit` |
  | Import depth (entry = 0, longest chain) | 8 | 9 | `module.depth_limit` |
  | Import declarations per module (duplicates counted) | 16 | 17 | `module.imports_limit` |

  Where the checks run:
  - **Count and depth** are checked when a module is enqueued, before it is
    opened.
  - **Imports per module** is checked at parse time, before any import is
    resolved.
  - **Total bytes** is checked after each bounded read.

  Rules:
  - Size never comes from `Stat`.
  - Cap refusals take precedence over cycle and missing-module diagnostics.
  - The visited set is keyed on the canonical module name.
  - The loader is iterative or explicitly depth-bounded.
  - The merged program must still respect `maxFunctions = 1024` and
    `maxDeclarations = 4096` (`parser.go:15-16,70`, today per file). The gate
    decides whether the merged total is checked and under what code.

  Exact-N and N+1 fixtures are generated in-test under `t.TempDir()`, following
  the `TestTokenBudgetStopsWithOneStableDiagnostic` pattern. Nothing large is
  committed. The armed mutants:
  - `>` changed to `>=` for each cap;
  - a size check that uses stat;
  - a count check after the open;
  - a diamond counted twice.

### Gate, plan shape, verification
- **D-14:** There is no standalone `/gsd-spike`, following the Phase 29 D-10
  precedent. Plan 30-00 is a Wave-0 tracer gate that writes
  `30-00-GATE-VERDICT.md`, in the same shape as the 29 verdict: a flags table
  with test names as evidence, then `verdict: PROCEED | SPLIT`. Every later plan
  depends on it. The gate settles each of the following as an automated test
  with an armed control:
  - (a) `os.Root` behavior per D-08/D-10, covering escapes, in-root symlinks
    and a FIFO that is refused without blocking;
  - (b) the case check per D-09;
  - (c) the full list of `Unit` consumers, and confirmation that omitempty
    keeps every existing golden byte-identical;
  - (d) injective naming per D-12, with an adversarial collision enumeration:
    `a_b.c` vs `a.b_c`, case variants, and a reversed-order build;
  - (e) a body-blind admission guard;
  - (f) the `maxFunctions` interaction with the merged program.
- **D-15:** The no-import byte-identity golden is committed **before** any
  modules change (success criterion 4). It covers generated C for the
  accepted corpus, plus diagnostic JSON and debugmap output for the existing
  refusal corpus. Existing byte-identity tests
  (`TestPhase272AcceptedCorpusByteIdentity`,
  `TestPhase273ConformanceCorpusByteIdentity`) stay green unchanged.
- **D-16:** The plan ceiling is 10, including 30-00. Split the work into
  Phase 30, the front end (syntax, loader, caps, graph refusals, `Unit`), and
  Phase 30.1 (admission, visibility re-derivation, naming and C emission,
  runnable gain) if any of the following holds:
  - the gate verdict is SPLIT;
  - the plan count would exceed 10;
  - `Unit` forces a non-omitempty or schema change to an existing golden;
  - the naming changes any no-import C bytes.

  Do not split only because the phase is large.
- **D-17:** Zero human UAT, per PROJECT.md § Verification Operating
  Preference:
  - every success criterion maps to a named `go test` in the existing CI;
  - every plan names its exact verification command and an armed control;
  - the runnable gain is checked against the pinned Phase 29 literals on both
    engines through the public CLI route, the same route Phase 29 used: build
    with `--manifest`, then `app run ART -- N`.

### Post-research decisions (2026-10-12, resolving 30-RESEARCH.md Open Questions 1-9 / A1-A14)
These were auto-followed per the standing preference, and each takes the
researcher's recommendation.
- **D-18 (amends D-07; resolves A1):** Imported module segments use
  `[a-z][a-z0-9]*(_[a-z0-9]+)*`: no `__`, no trailing `_`, at most 64 bytes.
  An exported function that is imported must have a name with no leading `_`,
  no `__` and no trailing `_`. Anything else is refused at the import site with
  `module.invalid_segment` or `module.invalid_import_name`. This keeps D-12's
  "no new `__`" without an escape scheme. Entry-module names are unaffected.
- **D-19 (resolves A2):** Duplicate refusals use two codes:
  - `module.duplicate`: an imported module name equals the entry header.
  - `module.duplicate_import`: the same module is imported twice in one file.
    Its repair deletes the later import, with a `"\n"` replacement because
    `DriverEligible` requires a non-empty `Replacement`.
    *Amended 2026-10-12 (plan check):* the deletion repair is offered only
    when every item in the later declaration also appears in the earlier
    one, so deleting it cannot drop an import. Otherwise the refusal stands
    with no repair.
- **D-20 (resolves A3, A4):** Two more refusals, both by name:
  - Imported modules may not declare `data` or `foreign` blocks:
    `module.imported_data_unsupported` and
    `module.imported_foreign_unsupported`. The foreign owner stays in the entry
    module, and the library stays pure scalar.
  - Importing any function named `main` is always refused.
- **D-21 (resolves A5):** The merged program is capped at the existing
  per-file limits, 1,024 functions and 4,096 declarations, exported from
  `syntax`. Over the cap is refused with `module.program_limit`.
  Exact-N/N+1 controls apply. The measured reason is that corevalidate cost is
  roughly quadratic: 26.7 s at 16,384 functions.
- **D-22 (resolves A6):** The commands that run through the loader are `check`,
  `run` (both engines), `build`, `app verify` and `debug-map`. `format`,
  `evidence`, `explain`, `query` and `interface` stay single-file. They refuse a
  source containing `import` with `module.imports_not_loaded`. `check.Program`
  fails closed on unresolved `Imports`.
  *Amended 2026-10-12 during planning:* `format` is the exception. Per D-03 it
  formats `import` declarations as plain syntax without loading modules, so it
  never refuses them. Only `evidence`, `explain`, `query` and `interface`
  refuse with `module.imports_not_loaded`. This is pinned by
  `TestPhase30SingleFileCommandsRefuseImports` (30-08).
- **D-23 (amends D-11 wording; resolves A7, A8):** The human projection is
  `ID CODE unit=<module> [s:e]: message`, with `unit=` present only when set.
  It is not a `unit:` prefix. Peer-refusal attribution derives the unit from
  the module segment of the ID in `Problem.Detail`. `cmd/schway-repair`
  declines any repair whose effective unit is not the entry, and never writes
  the entry file for it.
- **D-24 (resolves A10, A13):** Two loader rules:
  - The D-09 directory listing examines at most 65,536 entries; beyond that
    it is refused with `module.directory_limit`.
  - A cycle through the entry is detected with `os.SameFile` against the
    entry file and reported as `module.import_cycle`, never as
    `path_mismatch`.
- **D-25:** `..`, absolute paths and separators cannot be written in `import`
  grammar, so they are pinned as syntax refusals. Symlinks are the real escape
  vector and are pinned in modload (refines SC1 evidence).
  `module.imports_limit` exact-N is tested at parser level. End to end, 16
  imports would exceed the 16-module cap.
- **D-26 (resolves A12, A14):** SEED-261008-z1g (bounded symbol/caller query)
  stays dormant. Plan 30-09 performs the full AGENTS.md transition read
  (PRODUCT-ROADMAP, LANGUAGE-MATURITY, seeds) and records it.
- **D-27:** Committed fixtures are limited to `examples/phase30/` (library,
  entry, README). Every adversarial fixture is generated under `t.TempDir()`,
  so the five corpus registers (phase28 manifest, 272/273 allow-lists, maturity
  count, emitter inventory) move only for those files.

- **D-28 (resolves A9, A11; plan check, 2026-10-12):**
  - The corevalidate codes are `core.callee_not_exported` and
    `core.callee_not_imported`. The check code is `module.not_exported`.
  - Only counts are pinned. Timing figures (A11, host-specific) are evidence
    for D-21 and are never asserted in tests.
- **Pre-planning transition read (AGENTS.md, 2026-10-12):** this read covered
  `.planning/PRODUCT-ROADMAP.md` (M006 recommendation 2; lines 167-229, 303,
  329-346), `.planning/LANGUAGE-MATURITY.md` (lines 166, 189, 270, 343-360) and
  the three dormant seeds. The current source witness is that the parser has no
  `import` token, so local modules are absent, as both documents state.
  Dispositions:
  - Both documents keep "whole-program composition; not B1/DX-06", which is
    consistent with this phase's boundary.
  - **SEED-261008-z1g** (caller query): retain dormant. No named reviewer gap
    exists yet. Re-check after 30-09, when cross-file callers exist.
  - **SEED-261008-z1s** (contract and source forms): retain dormant. The import
    surface (D-01) was chosen by advisor research, and no named reviewer
    contests it.
  - **SEED-261008-trz** (failure observability): retain dormant. No named
    failure is undiagnosable. D-11 `Unit` attribution improves cross-file
    diagnosis.
  - Plan 30-09 repeats the read after completion (D-26).

- **D-29 (final plan check, 2026-10-12):** A library module may define a
  private function named `main`. It is an ordinary non-entry function: it is
  never an entry root, and no refusal rule applies to defining it. D-20 refuses
  only *importing* `main`. Tests assert the invariant this fixes: entry-unit
  (`Unit == ""`) diagnostics and the entry result are identical with and
  without the library `main`. No library-`main` refusal code is invented.
- **D-27 amendment (final plan check, 2026-10-12):** two committed files fall
  outside `examples/phase30/`. Both are accepted:
  - `testdata/phase30/prechange-corpus.json`, the D-15 golden. D-15 requires
    it, and it is evidence, not a fixture.
  - The `testdata/phase16/public-emitter-consumers.json` line re-pin forced by
    `session.go` edits. It moves a line number only, with no corpus content
    change.
- **D-15 scope note (final plan check, 2026-10-12):** "evidence" in the phase
  boundary is pinned through its inputs. The golden covers check JSON, core
  JSON, both C emitters, debug-map and format. The `evidence`, `explain` and
  `query` routes stay single-file (D-22) and are unchanged code paths for
  programs without imports. Their byte identity follows from unchanged core
  and C and is already pinned by the existing 272/273 identity tests.

### Claude's Discretion
- The exact names of the new `module.*` refusal codes beyond those listed. They
  must follow the `<family>.<snake_case>` convention, and every MOD-05 refusal
  must have its own code:
  - `module.import_cycle` (the message lists the cycle, and self, length-2 and
    length-3 cycles are distinct fixtures);
  - `module.not_found`;
  - `module.duplicate`;
  - `module.path_mismatch`;
  - `module.import_collision`;
  - `module.not_exported`.
- The internal package name and file layout for the loader (`modload` or a
  session file).
- How the Phase 29 library moves into a module file under `examples/phase30/`.
  Its header becomes the path-derived name, such as `adler32`. Leave the
  Phase 29 example files untouched as historical evidence, or keep them green
  unchanged.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Scope and requirements
- `.planning/ROADMAP.md` § Phase 30 — goal, success criteria 1-4, research flag
  and plan-count checkpoint
- `.planning/REQUIREMENTS.md` — MOD-03, MOD-04, MOD-05 (MOD-06 is Phase 31);
  Future Requirements and Out of Scope for modules
- `.planning/PROJECT.md` § Verification Operating Preference — zero human UAT,
  armed controls

### Settled module design
- `.planning/research/SUMMARY.md` Decision 7 (Modules), Decision 10, and the
  open items "Module caps", "`os.Root` symlink and macOS case-insensitivity"
  and "not B1/DX-06"
- `.planning/research/ARCHITECTURE.md` §5.1 — per-sub-decision
  recommendations: resolution site, naming, visibility, caps, whole-program
  flow, admission guard, staleness, attribution. D-12 overrides its C-symbol
  row (`module__name`).
- `.planning/research/STACK.md` Decision 5 — local modules. Its root rule and
  its 64 KiB cap are superseded by D-05 and D-13.
- `.planning/research/PITFALLS.md` — items 8 (import escape), 9 (graph
  semantics) and 10 (C collisions and ordering); normalization and case
  collisions (~line 171); root-relative paths in records (~line 177)

### Precedent
- `.planning/phases/29-byte-views-and-the-adler-32-library/29-CONTEXT.md` —
  D-02/D-04 (reserved builtin names), D-10/D-11 (gate-plan precedent),
  D-15/D-16 (public runnable route)
- `.planning/phases/29-byte-views-and-the-adler-32-library/29-00-GATE-VERDICT.md`
  — the verdict file shape to copy
- `examples/phase29/` — the Adler-32 fixture program, manifest and README,
  whose answers the runnable gain must reproduce

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `syntax.modulePath()` (`internal/compiler/syntax/parser.go:258-264`) parses
  dotted module paths. Reuse it for `import`.
- `exports()` item loop (`parser.go:266-286`) is the template for the import
  item list.
- `readBoundedFile` (`internal/compiler/session/session.go:1890`) is the
  limit+1 bounded read idiom.
- `syntax.MaxSourceBytes`, `MaxTokens`, `maxDeclarations` and `maxFunctions`
  (`parser.go:13-17`), with the `syntax.input_limit` refusal at `parser.go:88`.
- `cNames.allocate` / `cName` (`internal/compiler/cgen/cgen.go:1560,1580`) and
  function naming at `cgen_program.go:798`. The invariant and closed-namespace
  argument are at `cgen.go:140-160`. Reserved-`__` guard at `cgen.go:1547-1558`.
- `semanticID` (`internal/compiler/check/check.go:6161`): function IDs
  `s1:<module>:fn:<name>` now carry the real defining module.
- `TestTokenBudgetStopsWithOneStableDiagnostic`
  (`internal/compiler/syntax/syntax_test.go:32-41`) is the pattern for
  generated N/N+1 fixtures.

### Established Patterns
- Six consumers read one `ast.Program`: check, corevalidate, originvalidate,
  interp, cgen and pathoracle/callgraph. Whole-program merge keeps their shape.
  Every new op or field must reach all of them, with a seeded fault per peer.
- Diagnostic IDs hash `{Schema, Code, Span, Causes}`. Repairs hash kinds only
  (`diagnostic.go:~110,~133`).
- `Span` is file-local (`diagnostic.go:16-19`). Protocol spans are at
  `protocol.go:241,265,315`, and debugmap spans at `debugmap.go:111-119`.
- Byte-identity corpus tests guard generated C. Formatter idempotence and
  lossless lanes guard syntax.

### Integration Points
- New session-side loader (`SourceSet`): the only IO site, feeding parse →
  merge → `check` → `callgraph.Prune` → peers.
- `ast.Program` gains `Imports`. The merged program carries module identity
  per function.
- Callee signature table (`callSignatureTable`/`calleeContracts` in check) is
  the body-blind admission path. The guard test proves that a callee body is
  never read.
- corevalidate re-derives visibility from `Function.Exported` and refuses
  cross-module `OpCall` to an unexported callee.
- `cmd/schway` human diagnostic output gains an in-line `unit=<module>` field
  when the unit is set (D-23 amended the earlier "`unit:` prefix" wording).

</code_context>

<specifics>
## Specific Ideas

- The runnable gain uses the same 10 pinned Phase 29 case values (Adler-32 of
  cases 0..9) with the library imported. The library must exist as exactly one
  source file; a copy is a failure.
- The adversarial fixture list from research:
  - imported `main` vs entry `main`;
  - an import equal to a local `fn`;
  - an import of `byte_len`;
  - the same module imported twice;
  - self-import;
  - `import` used as an identifier;
  - `import m {}`;
  - `import` after `export`;
  - `..`, absolute and symlink escapes;
  - an in-root symlink;
  - a FIFO;
  - a case variant;
  - a non-ASCII segment;
  - a reversed import order;
  - a diamond;
  - cap N and N+1.

</specifics>

<deferred>
## Deferred Ideas

- A qualified call form (`m.f(x)`) and import aliases. Revisit only alongside
  aliasing, which M006 Future Requirements lists.
- Exported types and data across modules. M006 refuses `type` imports.
- An entry-header/path check for programs without imports, which would be a new
  refusal on existing programs.
- Fixing the pre-existing `SCHWAY__X` reserved-identifier output for source
  names with a leading `_` (`cName`). It is out of scope because it would
  change existing C bytes.
- Sorted module hashing in build identity and evidence, `Program.Modules[]`
  in evidence, and the second consumer: Phase 31 (MOD-06).

</deferred>

---

*Phase: 30-local-modules*
*Context gathered: 2026-10-12*
