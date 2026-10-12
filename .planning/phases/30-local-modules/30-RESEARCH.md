---
verification_role: proposal
---

# Phase 30: Local Modules - Research

**Researched:** 2026-10-12
**Domain:** Go 1.24 stdlib-only compiler (syntax / modload / check / core / corevalidate / callgraph / interp / cgen / diagnostic / protocol / session / CLI); C17 via Clang; no new dependencies
**Confidence:** HIGH on every repo fact below (read from source or probed with a built binary this session; probe programs were deleted afterwards and `git status` shows only the pre-existing `.planning/state.json` change). MEDIUM on the design recommendations flagged in Open Questions and tagged `[ASSUMED]` in the Assumptions Log.

<user_constraints>
## User Constraints (from CONTEXT.md)

D-01..D-17 are LOCKED (advisor research, auto-followed). Verbatim from `30-CONTEXT.md`:

### Locked Decisions

- **D-01** Import is a named import block; calls stay bare (`let sum = adler32(bytes)`); `ast.RHS.Callee` stays a bare string; no qualified `m.f(x)` form; no hybrid.
- **D-02** Grammar `program := "module" modulePath importDecl* exportBlock decl*`, `importDecl := "import" modulePath "{" ("fn" IDENT)+ "}"`; items whitespace/newline separated, no commas; reuse the `exports()` loop (`parser.go:266-286`); new `TokenImport`; empty list refused `syntax.expected_import_name`; `import` after `export` refused `syntax.import_after_export`; `type` items refused by check with a named code.
- **D-03** Formatter keeps source order, never sorts imports; one module per `import`, one item per line; reuse export-block context and `blankBeforeTopLevel`; idempotence and lossless lanes extended to `import`.
- **D-04** Collision refusals cover: import vs local `fn`; import vs import; imported `main` vs entry `main`; import of a reserved builtin (`byte_len`, `byte_at`, `write_*`). Collision diagnostic names both origins, no automatic repair. Deterministic repairs for exactly two refusals: duplicate import of the same module (delete the later one) and import after export (move it above `export`). `module.not_exported` gets a hint only. An unresolved bare call never synthesizes an import. No unused-import refusal.
- **D-05** Root = entry file's directory, opened with `os.OpenRoot`; `a.b.c` -> `<root>/a/b/c.schway`; no project marker, no upward search, no env root; resolver appends exact lowercase `.schway`.
- **D-06** Entry header is NOT checked against its path; imported headers must equal the path-derived dotted name (`module.path_mismatch`); an imported module whose name equals the entry header is `module.duplicate`.
- **D-07** Segments `[a-z][a-z0-9_]*`, <= 64 bytes, ASCII only, else `module.invalid_segment`; the check lives in modload, not the lexer.
- **D-08** Symlinks refused, including in-root ones: per-component `Root.Lstat`, any `ModeSymlink` -> `module.symlink`; `os.Root` is the second layer (escape -> `module.escape`).
- **D-09** Exact case verified on every platform via `ReadDir` of the parent through the root; miss -> `module.not_found`, or `module.case_mismatch` when an `EqualFold` match exists; tests must not depend on macOS-only behavior (on Linux create both spellings).
- **D-10** `Root.Lstat` + `IsRegular()` before `Open` (a FIFO open blocks > 2 s on macOS); after open `Stat` + `os.SameFile`; one limit+1 bounded read (`readBoundedFile` idiom, `session.go:1890`); a session-side `SourceSet` (modload) is the ONLY IO site; parse/check/hash/identity use in-memory bytes, never re-read; diagnostics carry root-relative names only.
- **D-11** `Unit` = module name, omitempty, empty = entry; added to `Diagnostic`, `Cause`, `Repair`, `debugmap.Entry`; Cause/Repair with empty Unit inherit the diagnostic's; `Span` stays file-local; `Unit` added `json:",omitempty"` to the anonymous identity structs of `diagnostic.Error` and `ErrorWithRepairs`; human projection prints a `unit:` prefix only when set. One-way door.
- **D-12** Imported C names `SCHWAY_M_<len>_<SEG>_..._<len>_<FN>` (lengths count source bytes before uppercasing); entry functions keep `cName(function.Name)` and current ordinal; imported functions allocated after entry functions, sorted by (module, function), ordinal = position in that order; entry names win a contested name; keep `^SCHWAY_[A-Z0-9_]*$`, no new `__`, no leading `_`. Planner must confirm cgen receives (module, bare name) separately.
- **D-13** Caps: per-module bytes 1,048,576 (existing `syntax.input_limit`, `Unit` names the file); distinct modules incl. entry 16 / 17 -> `module.count_limit`; total bytes 4,194,304 / +1 -> `module.total_bytes_limit`; import depth (entry = 0) 8 / 9 -> `module.depth_limit`; import declarations per module 16 / 17 -> `module.imports_limit`. Count and depth checked at enqueue (before open); imports per module at parse time; total after each bounded read; size never from `Stat`; cap refusals take precedence over cycle and missing-module diagnostics; visited set keyed on canonical module name; loader iterative or depth-bounded; merged program must still respect `maxFunctions = 1024` and `maxDeclarations = 4096`, the gate decides whether and under what code; N/N+1 fixtures generated in-test under `t.TempDir()`; armed mutants: `>` -> `>=` per cap, size from stat, count check after open, diamond counted twice.
- **D-14** No standalone spike; plan 30-00 is the Wave-0 tracer gate writing `30-00-GATE-VERDICT.md` (flags table with test names, `verdict: PROCEED | SPLIT`); flags (a) `os.Root` behavior incl. FIFO, (b) case check, (c) `Unit` consumer list + omitempty keeps goldens byte-identical, (d) injective naming with adversarial enumeration (`a_b.c` vs `a.b_c`, case variants, reversed-order build), (e) body-blind admission guard, (f) `maxFunctions` interaction with the merged program.
- **D-15** No-import byte-identity golden committed BEFORE any modules change: generated C for the accepted corpus plus diagnostic JSON and debugmap output for the refusal corpus; `TestPhase272AcceptedCorpusByteIdentity` and `TestPhase273ConformanceCorpusByteIdentity` stay green unchanged.
- **D-16** Plan ceiling 10 including 30-00. Split into 30 (front end) and 30.1 (admission, visibility, naming, C, runnable gain) only if: verdict SPLIT; plan count would exceed 10; `Unit` forces a non-omitempty/schema change to an existing golden; naming changes any no-import C bytes. Do not split only because the phase is large.
- **D-17** Zero human UAT; every success criterion maps to a named `go test` in existing CI; every plan names its exact verification command and an armed control; runnable gain checked against the pinned Phase 29 literals on both engines through the public CLI route (`build --manifest`, then `app run ART -- N`).

### Claude's Discretion

- Exact names of new `module.*` codes beyond those listed (convention `<family>.<snake_case>`; every MOD-05 refusal needs its own code): `module.import_cycle`, `module.not_found`, `module.duplicate`, `module.path_mismatch`, `module.import_collision`, `module.not_exported`.
- Internal package name and file layout for the loader (`modload` or a session file).
- How the Phase 29 library moves into a module file under `examples/phase30/` (header becomes the path-derived name, e.g. `adler32`); Phase 29 example files stay untouched or green unchanged.

### Deferred Ideas (OUT OF SCOPE)

- Qualified call form `m.f(x)` and import aliases.
- Exported types and data across modules (M006 refuses `type` imports).
- Entry-header/path check for programs without imports.
- Fixing the pre-existing `SCHWAY__X` output for source names with a leading `_` (`cName`).
- Sorted module hashing in build identity/evidence, `Program.Modules[]` in evidence, second consumer (Phase 31, MOD-06).
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| MOD-03 | Import another local file through an explicit import; header must match path; resolution never leaves the entry directory (no search path, `..`, absolute, symlink escape) | Loader design (Sections "Loader", "os.Root probe"), grammar and formatter inventory, escape table (`..`/absolute are unrepresentable in import syntax and refused as syntax errors; symlinks refused by per-component `Lstat`) |
| MOD-04 | Only exported names usable; private use refused with a stable code; independent validator re-derives visibility; cross-module calls admitted from signatures alone | check scope design (`buildCalleeContracts` per module), `core.Function.Module/Exported` + import facts, corevalidate law at `corevalidate.go:2900` and `:3299`, body-blind guard design |
| MOD-05 | Each module-graph error has its own named, source-attributed diagnostic; diamond loads/emits once; no-import programs byte-identical; diagnostics in imported files name the file | Graph/caps design, `Unit` plumbing inventory, D-15 golden design, naming (D-12) injectivity, diamond-once tests |
</phase_requirements>

## Summary

The CONTEXT settles the WHAT; the repo facts that most change the HOW are these. (1) There is exactly one parse-and-check seam, `session.Check(source []byte)` (`session.go:773`), and every file route reads bytes with `readBoundedFile` and hands them to it, so the loader plugs in as a path-taking sibling (`CheckSet`) that parses every module, merges, and calls the same `check.Program`; the byte-taking `Check` must fail closed on any `import` it cannot resolve because `evidence`, `explain`, `query`, `interface export/core`, the corpus-verify lanes and `debug-map` all call `syntax.Parse` + `check.Program` directly. The ARCHITECTURE note about a `publishedOriginProblemFile` re-read is stale: no such symbol exists in the tree; `admitProgram` already reuses `checked.Tree` (`session_admit.go:48-51`). (2) `check.Program` bakes ONE module name into every ID and into the bare-name `calleeContracts` table (`check.go:256`, `:1962-1972`), so cross-module support is a per-function module plus a per-module scope table, not a pre-merge string join. `core.Function` has `Name` (bare) separately, so cgen can already derive module-qualified names once `core.Function` gains `Module` (omitempty, entry module stays empty so no-import JSON is byte-identical). (3) Several consumers resolve a function by bare `Name`: `callgraph.EntryFunction` (`callgraph.go:330`), `interp.findFunction` (`interp.go:1979`), `debugmap` (`debugmap.go:167`); each needs a module-aware tie-break or a library `main` becomes a duplicate-entry refusal. (4) `corevalidate` is roughly quadratic in function count (measured: 1,024 functions 0.13 s, 4,096 functions 1.8 s, 16,384 functions 26.7 s on this host), so the merged program MUST be capped at the same 1,024 functions / 4,096 declarations as one file; `MaxCheckWork = 100_000` alone would still admit about 1,900 functions. (5) `Diagnostic` JSON/ID plumbing is safe with omitempty (pinned ID `diagnostic:d34396fcc8ea6530a1f04dc3` is an existing armed control) but the repair driver applies spans literally to the entry file and has no unit concept, so a cross-unit repair must be declined or it corrupts the wrong file.

The "Phase 30 composition point" comment already in `session_entry.go` says entry candidates come from the root module only; that is the contract to implement. No new dependency is needed: `os.OpenRoot`, `Root.Lstat/Open`, `os.SameFile`, `io.LimitReader`, `sort`, `syscall.Mkfifo` (unix test-only) are stdlib.

**Primary recommendation:** Build ten plans in six waves (table in "Suggested Plan Decomposition"): D-15 golden first inside 30-00, then syntax and `Unit` plumbing in parallel, then loader fs layer in parallel with check/core scope, then loader graph, peers and cgen in parallel, then session wiring, then the runnable gain with registers and closure. Keep all name/visibility semantics in `check` (one derivation) and corevalidate (independent second), keep `modload` structural (paths, files, caps, graph), and generate every adversarial fixture under `t.TempDir()` so only two new `.schway` files are committed.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Read module bytes (the only IO) | `modload` (session-side loader) | `session` (passes entry path, owns CLI-facing errors) | D-10: single IO site; `syntax`/`check`/`core`/peers stay pure |
| Parse one file, `import` grammar, imports-per-module cap | `syntax` | `modload` (stamps `Unit` on parse diagnostics) | Parser is file-local; cap is parse-time per D-13 |
| Graph: cycles, missing, duplicate, header/path, depth/count, diamond-once, canonical merge order | `modload` | `check` (never sees raw files) | Structural facts about files, not names |
| Import resolution, collisions, export visibility, scoped callee contracts | `check` | `corevalidate` (independent re-derivation) | One source-level derivation plus one independent core-level law |
| Visibility law on `OpCall` | `corevalidate` | `check` (producer) | MOD-04 independent validator |
| Entry selection, pruning | `callgraph` (module-aware `EntryFunction`) | `session_entry.go` (`resolveEntry`) | Library exports are never roots (comment in `session_entry.go`) |
| C names, diamond emitted once, no-import C identical | `cgen` | `callgraph.Order` (deterministic order by sorted ID) | D-12 |
| `Unit` on diagnostics, IDs, human text | `diagnostic`, `protocol` | `cmd/schway` (emit), `debugmap` | D-11 |
| Repair application | `cmd/schway-repair` | `diagnostic` (eligibility) | Driver edits one file; must decline cross-unit repairs |
| Build/run/verify routes | `session` file routes | `native` (receipt digests entry source only) | Whole-program composition, not separate compilation |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go standard library | 1.24.0 (go.mod `go 1.24`, local `go1.24.0 darwin/arm64`) | `os.OpenRoot`, `Root.Lstat`, `Root.Open`, `os.SameFile`, `io.LimitReader`, `sort`, `errors`, `io/fs` | Project rule: stdlib only; `os.Root` is the standard traversal-resistant API [VERIFIED: `go doc os.Root`, go1.24.0] |
| Clang (system) | Apple clang at `/usr/bin/clang` | Native lane for the runnable gain | Existing native path; CI installs/requires clang on both OSes [VERIFIED: `.github/workflows/ci.yml` "Require clang"] |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `syscall.Mkfifo` | stdlib (unix) | FIFO fixture for the non-blocking-refusal control | Test only, behind `//go:build unix` |
| `go/ast`, `go/parser` | stdlib | Inventory/AST scan guards (single IO site, source-read inventory) | Same technique as `session_entry_inventory_test.go` |
| `testing`, `t.TempDir()` | stdlib | Generated graph and cap fixtures | D-13: nothing large committed |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `os.Root` + per-component `Lstat` | `filepath.EvalSymlinks` + prefix check | Racy and platform dependent; D-08 already chose `os.Root` |
| New `modload` package | A `session_modules.go` file | A separate package lets an AST test assert it is the only package opening module files; recommended |

**Installation:** none (`go build ./...`).

**Version verification:** no external packages are added; nothing to look up on a registry.

## Package Legitimacy Audit

No external packages are installed in this phase. **Packages removed due to [SLOP] verdict:** none. **Packages flagged [SUS]:** none. `gsd-tools query package-legitimacy check` was not run because the candidate set is empty.

## Architecture Patterns

### System Architecture Diagram

```
CLI (cmd/schway: check | run --engine=* | build --manifest | app verify | debug-map)
        |  path
        v
session file route (CheckFile / runInterpreterCommand / runNativeCommand /
                    BuildApplicationFile / VerifyApplicationCasesFile / DebugMapCommandFile)
        |  entry path
        v
modload.Load(entryPath)  ==== the ONLY module IO site ====
   |  os.Open(entry)  (user-supplied authority, symlink-follow as today)
   |  os.OpenRoot(dir(entry))
   |  worklist (iterative): canonical module name -> enqueue? [count/depth caps BEFORE open]
   |      per component: Root.Lstat (symlink? regular? dir?) -> ReadDir exact-case -> Lstat IsRegular
   |      -> Open -> Stat+SameFile -> limit+1 bounded read -> total-bytes cap
   |      -> syntax.Parse(bytes)  [imports_limit at parse time]  -> stamp Unit on parse diagnostics
   |      -> header == path-derived name (else module.path_mismatch); name == entry header (module.duplicate)
   |      -> same file as entry (os.SameFile) => edge to entry node (cycle through entry)
   |  three-color iterative DFS => module.import_cycle (self / 2 / 3), diamond visited once
   |  merge: entry funcs (source order) + imported funcs sorted by (module), source order within
   |  merged-program cap (1,024 functions / 4,096 declarations)  => ast.Program{Units, FuncDecl.Module}
   v
check.Program(merged)   -- per-function module for IDs; per-module scope (own + imported items);
   |                       import resolution diagnostics (not_exported, collision, type item, builtin, main);
   |                       calleeContracts built from signatures only; Unit stamped per function
   v
core.Program (+ Function.Module / Function.Exported omitempty, Program import facts omitempty)
   v
admitProgram: corevalidate (visibility law at both replay sites) -> originvalidate -> pathoracle
   v
resolveEntry: callgraph.EntryFunction (entry-module candidates only) -> PruneToReachable
   v
interp (findFunction prefers entry module)  |  cgen (two-pass name allocation, D-12)
   v                                         v
Execution document                      C17 -> clang -> app run ART -- N
```

### Recommended Project Structure
```
internal/compiler/modload/        # NEW: fs layer (root, lstat, case, fifo, caps), graph, merge
  load.go  fs.go  graph.go  merge.go  *_test.go
internal/compiler/ast/ast.go      # + ImportDecl, ImportItem, Unit, FuncDecl.Module
internal/compiler/syntax/         # + TokenImport, import parsing, formatter, limits
internal/compiler/check/check.go  # per-module scope, Unit stamping, new module.* diagnostics
internal/compiler/core/core.go    # + Function.Module, Function.Exported, Program import facts
internal/compiler/corevalidate/   # + visibility law
internal/compiler/callgraph/      # EntryFunction module filter
internal/compiler/session/        # new session_modules.go (route wiring) + small edits
examples/phase30/                 # adler32.schway (library), adler32_cli.schway (entry), README.md
```

### Pattern 1: Loader plugs in beside `Check`, never inside it
**What:** `CheckSet(set)` = parse each module, merge, `check.Program`. `Check(source)` stays byte-taking and fails closed on unresolved imports (`check.Program` refuses `len(program.Imports) > 0 && program.Units == nil`). **When:** every file route. **Why:** the other direct `Parse`+`check.Program` callers (`evidence.go:181,198`; `session_phase6_explain.go:61`; `session_phase6_query.go:194,240`; `session_phase6_verify.go:275`; `session.go:1992,2001,1823`) would otherwise report a misleading unresolved-callee error or silently ignore imports.

### Pattern 2: Seeded unexported test seam per cap and per guard
Repo convention (D-07-42): package-level unexported `...ForTest` vars, same-package tests only, restored with `defer`. Each cap comparison goes through one helper `exceeds(n, cap)` with an off-by-one seam so `>` -> `>=` is a real armed mutant; loader filesystem access goes through a small interface so tests inject a fake whose `Stat` lies about size, whose open counter proves "never opened", and whose file identity changes between `Lstat` and `Open`.

### Anti-Patterns to Avoid
- **Joining module and name with `__` or plain `_`** (collides: `a_b.c` vs `a.b_c`; C17 7.1.3 reserves `__`). Use D-12 length prefixes.
- **Trusting `Stat().Size()` for the byte cap** (D-10, mutant in D-13).
- **`Root.Lstat("a/b/c.schway")` as the symlink test**: probe shows it follows an in-root intermediate symlink (`subl/n.schway` returned a regular file). Lstat each component.
- **Keying anything on bare function name across modules** (`debugmap`, `findFunction`, `EntryFunction`).
- **Committing adversarial `.schway` fixtures**: every committed file moves five registers (Pitfall 2).

## Where the loader plugs in (research item 1)

Source-reading entry points (all `readBoundedFile(path, syntax.MaxSourceBytes)` in non-test code) [VERIFIED: grep this session]:

| Route | Site | Action in Phase 30 |
|-------|------|--------------------|
| `check` | `CheckFile` `session.go:829` (used by `CheckCommandFile` `:1082`) | Route through loader |
| `run --engine=interpreter` | `RunInterpreterFile` `:1227`, `runInterpreterCommand` `:1238` | Route through loader (`RunInterpreter(source)` API stays single-source) |
| `run --engine=native` | `RunNativeFile` `:1544`, `runNativeCommand` `:1558` | Route through loader |
| `build [--manifest]` | `BuildApplicationFile` `:1482` -> `BuildApplication(source, ...)` | Route through loader; pass the ENTRY bytes as the `source` argument (receipt `SourceDigest = digestBytes(source)`, `native_app.go:213,252`); `EmittedCDigest` already covers library code so a library edit changes build identity; source-set hashing is Phase 31 (MOD-06) |
| `app verify` | `VerifyApplicationCasesFile` `session_app_verify.go:211` | Route through loader (same Check -> admitProgram -> resolveEntry steps) |
| `debug-map` | `DebugMapCommandFile` `:1819` | Route through loader (recommended, Open Question 6) |
| `format` | `FormatFile` `:798` | Unchanged: per-file, no graph needed |
| `evidence`, `evidence --validate` | `:1611`, `:1640`, `session_phase6_evidence.go:100` | Unchanged single-file; imports refused by name (`check.Program` fail-closed); module hashing is Phase 31 |
| `interface export/core`, `explain`, `query` | `:1674`, `:1729`, `session_phase6_explain.go:57`, `session_phase6_query.go:190,236` | Unchanged single-file; imports refused by name |
| Corpus verify lanes (`session.go:593,1961-3207`, `session_phase7.go:417`) | fixture reads | Unchanged |
| `app run ART -- N` | `native.RunApplication` | No source read (runs the artifact) |
| `cache` | `native.go:174` `cacheInputs(ctx, source string, ...)` | Keyed on emitted C, toolchain and foreign sources, not Lang source; no change |

`RunInterpreterApplication(source, input)` (`:1155`) has no non-test caller (`grep`); leave it single-source. `publishedOriginProblemFile` does not exist [VERIFIED: grep for the symbol across `internal` and `cmd` returned nothing]; `admitProgram` already takes `checked.Tree` (`session_admit.go:51`: `originPeerRefusalDiagnostic(problems[0], program, checked.Tree)`). That helper locates spans by scanning the ENTRY tree for `function.Name` (`session.go:898`), so for an imported function it would mislead; carry per-unit trees in `CheckResult` (`UnitTrees map[string]syntax.Tree`) or return the span-less fallback with `Unit` set.

**Single IO site test:** an AST test (pattern of `TestEntryFunctionCalledOnlyViaResolveEntry`, `session_entry_inventory_test.go`) asserting that, outside `modload` and `*_test.go`, no function in the pinned set above reads module bytes with `readBoundedFile` except the pinned single-file list; adding a new source-reading route fails the test until classified.

## Merging `ast.Program` (research item 2)

- `check.Program` takes one `ast.Program` and uses `program.Module` for every ID: `semanticID(program.Module, "fn", function.Name)` at `check.go:256`, `checkBranch(program.Module, ...)` `:293,371`, `checkLinear(program.Module, ...)` `:312`, parameter IDs `:463`. `semanticID` is `func semanticID(module, kind, name string) string { return "s1:" + module + ":" + kind + ":" + name }` (`check.go:6161`).
- `calleeContracts` is keyed by bare callee name: `contracts[function.Name] = calleeContract{` (`check.go:1962-1972`), and `functionNames[function.Name]` (`:252`). A merged flat table would let module A call module B's private function by bare name. **Fix:** `FuncDecl.Module string` (empty = entry) and `ast.Program.Units []ast.Unit` (name, imports, exports) set only by the merge; `check.Program` builds one scope per module (own functions + resolved import items) and passes the calling function's scope. Imported items enter a scope as `calleeContract{ID, ParameterType, ReturnType, ReturnsBorrowOfParam}` built from the declaration's signature fields only, so SEM-05 body-blindness holds by construction (the type has no body field).
- `semanticID` / function IDs: imported function ID = `s1:<defining module>:fn:<name>`; unique per module, so `core.duplicate_function_id` is unaffected. `callgraph.Order` roots are `sort.Strings` of function IDs (`callgraph.go:489-493`), so core order is independent of import order; canonicalize the merged `Funcs` anyway (entry functions in source order, then imported modules sorted by name, source order within a module).
- **Can `function.Name` reaching cgen carry (module, bare name)?** Yes: `core.Function` already stores the bare `Name` and a separate `ID` (`core.go:127-150`), and cgen allocates `functionNames[index] = globals.allocate(cName(function.Name), "function", index)` (`cgen_program.go:798`). Adding `Function.Module` (`json:"module,omitempty"`, entry functions leave it empty) gives cgen both parts without parsing the ID. **Do NOT** leave the new field untagged or `json:"-"`: `corevalidate.cloneProgram` is a JSON round trip (`corevalidate.go:5225-5235`), so an untagged field is silently dropped by `Validate(...).Program()`, which would turn every imported function into an entry function after admission.
- Constraints on imported modules in Phase 30 (recommended, Open Question 3): `data` and `foreign` declarations are refused in imported modules. `types` is a bare-name map (`check.go:132-149`) and foreign symbols are program-global (`check.go:156`); making them module-scoped is out of scope and D-02 already defers exported types.
- `core.mixed_body_versions` (`check.go:120-130`) stays program-wide: a match-bodied entry importing a linear library is refused, because `core.Program.Schema` cannot mix (comment at `check.go:128`).

## Unit plumbing: consumers of Diagnostic / Cause / Repair / debugmap.Entry (research item 3)

Non-test consumers [VERIFIED: grep of `diagnostic.Diagnostic|Cause|Repair|.Primary|.Causes|.Repairs` outside check/syntax]: `diagnostic/diagnostic.go` (definition, ID hashing), `protocol/protocol.go` (human line `:625`, `DebugMapEntry :179`, `Result.Diagnostics`), `debugmap/debugmap.go` (`Entry`, `astByName`), `cmd/schway/main.go:674` (`problemResult`), `cmd/schway-repair/repair.go` (`jsonRepair :57`, `jsonDiagnostic :64`, `driverEligible :272`, `selectRepair :341`, `applyRepair :416`), `reduce/mismatch.go:71-125` (embeds `[]diagnostic.Cause`), `evidence/evidence.go` (returns diagnostics), `session` (`session.go` 33 sites, `session_phase6_explain.go` 16, `session_phase6_query.go` 7, `session_admit.go`, `session_entry.go`, injectors), `pathoracle.go` (1), `check`, `syntax`.

ID hashing [VERIFIED `diagnostic.go:108-121,123-156`]: `Error` hashes `struct{Schema string; Code string; Span Span; Causes []Cause}`; `ErrorWithRepairs` adds `RepairKinds []string`. Adding `Unit string \`json:",omitempty"\`` to both anonymous structs (and `unit,omitempty` to `Cause`) leaves empty-unit bytes identical. Armed controls already in the tree: `TestDiagnosticZeroBytesUnchanged` pins `const wantID = "diagnostic:d34396fcc8ea6530a1f04dc3"` (`diagnostic_test.go:243`); `TestRepairFieldEnumerationIsExhaustive` pins `expected := []string{"Applicability", "Detail", "Kind", "Replacement", "Span"}` (`:176`) and forces the author to classify `Repair.Unit` as non-identity (like Span) and update the list.

Goldens that must stay byte-identical (run as the gate-(c) set): `testdata/phase16/validation-corpus-run-record.jsonl`, `testdata/distinctness/`, `cmd/schway-repair/testdata/*_diagnose_capture.json`, `TestPhase1DiagnosticGoldenUnchanged` (`session_test.go:100`), `diagnostic_test.go` tests above, `TestPhase272...`, `TestPhase273...`, `TestPhase28EntryExitClassManifest`, plus the new D-15 golden.

Where `Unit` is stamped:
1. **check** (one site): record `starts[i] = len(result.Diagnostics)` at the top of each iteration of the `for _, function := range program.Funcs` loop (`check.go:255`) and after the loop; after the loop stamp `result.Diagnostics[starts[i]:starts[i+1]]` with function i's module (the body has many `continue`s so wrapping per-append is error-prone). The reserved-builtin loop (`:104`) and the two program-level passes (`checkInterproceduralLoanLiveness`, per owning function; `checkCallGraphAcyclic`, whose members all lie in one module because the import graph is acyclic and calls follow imports) stamp by owning function.
2. **modload**: parse diagnostics of each imported file, graph diagnostics (primary span = the import declaration in the importing file; `Unit` = importer; cycle members carry `Cause.Unit` for hops in other files).
3. **session peers**: `peerRefusalDiagnostic` (`session.go:853`) has only `Code/Detail`; derive `Unit` from the module segment of the ID in `Detail` (`s1:<module>:...`; entry module -> empty) and, for `originvalidate.Problem`, from `FunctionID`. Best effort, documented.
4. **protocol human** (`protocol.go:625`): existing line is `"%s %s [%d:%d]: %s\n"`; recommended form when set: `"%s %s unit=%s [%d:%d]: %s\n"`, mirroring the existing `module=%s` header token; unchanged when empty. (CONTEXT says "a `unit:` prefix"; Open Question 7.)
5. **debugmap** (`debugmap.go:165-187`): key `astByName` by (module, name) using `ast.FuncDecl.Module` and `core.Function.Module`; set `Entry.Unit`. `protocol.DebugMapEntry` projects no span today (`session.go:1862-1865`); add `Unit` omitempty.
6. **schway-repair**: decode `unit` in `jsonDiagnostic` and `jsonRepair`; `selectRepair` treats a repair whose effective unit is non-empty as not driver-eligible (it falls under the existing `repair.none_eligible`), because `applyRepair` splices `[Span.Start, Span.End)` into the one source path it was given. Mutant: driver ignores `unit` -> patches the entry file at an imported span. The driver may not import `internal/compiler/*` (`import_boundary_test.go`) and may not decode prose fields (`TestRepairDriverDecodesNoProseFields`); `unit` is neither.
7. **explain/query/injectors**: stay single-file; a `Unit != ""` diagnostic never reaches them in Phase 30 because the single-source `Check` refuses imports first.

## Body-blind admission path (research item 4)

Two signature-only channels exist: `calleeContracts` (AST-derived, pre-body) and `callSignatureTable` (core-derived `core.FunctionSignature`, includes the `Callable` bit derived from the callee's CHECKED body by `originvalidate.BuildInterface`; `check.go:2020-2033`). Admission consults only these: `verifyCallableRefusal` reads `table.lookup(operation.CalleeID)` (`check.go:2077-2110`). Existing guard: `TestCallAdmissionBodyBlindControl` (`check_test.go:3918`) counts `calleeBodyReadObserved` at zero in production and non-zero with `verifyCallableRefusalBodyReadSeam` engaged. Phase 30 extension (all armed):

1. `TestPhase30CallAdmissionBodyBlind`: two-module sets whose library bodies differ (different valid U64 arithmetic) but whose signature table entry (`ID`, parameter, return, `Callable`) is identical; assert the importer's diagnostics, its `calleeContract` entries and its core function JSON are byte-identical and `calleeBodyReadObserved` fires zero times. Control: engage `verifyCallableRefusalBodyReadSeam` on the module fixture -> observer fires.
2. Converse (documents the boundary, SC2 wording): a body change that flips `Callable` (borrow-derived return without a declared origin) changes the SIGNATURE table entry and legitimately flips the caller to `core.callee_not_callable`; the guard compares equal-signature pairs only.
3. Structural: a `go/ast` scan asserting the scoped-contract builder (new function in `check.go`) references no selector `.Body`, `.Linear`, `.Match` (the type `calleeContract` has no such field).
4. Mutant `scopedContractBodyReadSeam` (reads the imported `FuncDecl.Body` to compute a contract) -> the equal-signature test goes red.

## corevalidate: re-deriving visibility independently (research item 5)

Data needed in `core` (all additive, omitempty so no-import bytes are unchanged): `Function.Module string` (defining module, empty for the entry module), `Function.Exported bool` (meaningful only when `Module != ""`; entry functions are never cross-module callees because the entry cannot be imported without a cycle), and `Program` import facts: `Imports []ModuleImport{From, Module, Name string}`. Law, added where `corevalidate` already resolves `CalleeID` at the two replay sites (`corevalidate.go:2900` and `:3299`, both `declaredFunctionIDs[operation.CalleeID]` followed by `peerCallable`), plus `checkCallGraphAcyclic`'s edge walk (`:539-600`) if a graph-level pass is preferred:

- same-module call: allowed;
- cross-module call: callee `Module != ""` AND `callee.Exported` (else `core.callee_not_exported`) AND an import fact `{From: caller.Module, Module: callee.Module, Name: callee.Name}` exists (else `core.callee_not_imported`); a call whose callee is in the entry module from a non-entry caller is refused.
- fail closed: absent facts refuse; a forged core that strips the import table refuses every cross-module call.

Honest boundary: a coordinated lie (matching `Exported` and import facts) escapes corevalidate by design, the same frontend trust boundary the spike findings document; Phase 31's module digests are the cross-check. Seeded faults per peer: `disableVisibilityPeerForTest` in corevalidate so a check-admitted forged program is shown refused by corevalidate alone; a check-side seam admitting a private import shows the independent refusal still fires.

## Resource caps and the merged program (research item 6)

`maxFunctions`/`maxDeclarations` are per file today: `maxDeclarations = 4096`, `maxFunctions = 1024` (`parser.go:15-16`) enforced in `programLimit` as `if len(program.Data)+len(program.Funcs) > maxDeclarations || len(program.Funcs) > maxFunctions {` (`parser.go:70`) once per `Parse`; breach replaces all diagnostics with `syntax.input_limit` at offset `len(source)` (`parser.go:59-62`). Nothing bounds the merged total except `MaxCheckWork = 100_000` (`session.go:41`) applied after `check.Program` (`:783`) to `len(parsed.Tree.Tokens)` plus checker work. Measured this session (darwin/arm64, trivial `Byte` call chains, probe deleted):

| Functions | `check.Program` work | `corevalidate.Validate` |
|-----------|---------------------|-------------------------|
| 1,024 (one file) | 20,471 (52,221 with tokens via `session.Check`) | 0.13 s |
| 2,048 | 40,942 | 0.44 s |
| 4,096 | 81,884 | 1.78 s |
| 16,384 (16 modules x 1,024, merged programmatically) | 327,536 | 26.7 s (`check.Program` 0.30 s) |

Quadratic-ish corevalidate growth means an uncapped merge is a 27 s admission cost that runs more than once per command (`admitProgram`, then `interp.Run` revalidates at `interp.go:612`). **Recommendation (gate flag f):** enforce the same totals on the merged program: functions <= 1,024 and functions + data <= 4,096, refused with a new named code `module.program_limit` (Unit = the module whose declarations first cross the limit in canonical order), exact-N accepted / N+1 refused fixtures generated in `TempDir`; additionally sum every module's `len(Tree.Tokens)` into `CheckResult.Work` so `check.work_limit` still bounds tokens. Direct `check.Program` callers (`DebugMapCommandFile`, `evidence`) have no `MaxCheckWork` guard today; the merged debug-map route must apply it. Timings are one host's; do not pin them, pin only the counts.

## Byte-identity: existing tests and the D-15 golden (research item 7)

Existing: `TestPhase272AcceptedCorpusByteIdentity` (`session_phase272_identity_test.go:128`; walks `examples/` and `testdata/`, `session.Check(source)`, hashes core JSON and `cgen.EmitApplication`) and `TestPhase273ConformanceCorpusByteIdentity` (`session_phase273_identity_test.go:146`; `Phase16ControlNativeC`). Both are frozen at older revisions with allow-lists (`phase272NewlyAcceptedPaths`, `phase272EntryPrecedencePaths`); neither covers diagnostics or debugmap, and neither pins HEAD.

**D-15 golden design (plan 30-00, first commit, test + golden only):** `testdata/phase30/prechange-corpus.json`, schema `phase30.prechange-corpus/1`, a `revision` field, generated with the existing env-var pattern (`SCHWAY_PHASE30_WRITE_GOLDEN=1`, revision supplied by the caller, no subprocess). Per `.schway` file under `examples/` and `testdata/` (sorted, same walk as 272/273): (1) SHA-256 of `session.CheckCommandFile(path)` result JSON (deterministic: `ElapsedNS` is set only when `SCHWAY_OBSERVE_TIMING=1`, `session.go:1916-1925`); this covers diagnostics JSON, IDs and the file route; (2) for accepted files, core JSON SHA, `cgen.EmitApplication` SHA or verbatim error, `Phase16ControlNativeC` SHA or error; (3) SHA of `DebugMapCommandFile(path, "")` result JSON; (4) SHA of `syntax.Format` output. Gate: `TestPhase30PreChangeGolden` must pass unchanged after every later plan; control `TestPhase30PreChangeGoldenIsNotInert` flips one recorded hash in a copy and expects failure. "Committed before the change" is made checkable by plan ordering (30-00 commits the golden first; every later plan `depends_on: [30-00]`) and the recorded `revision`.

New committed files appear in the walk, so the new golden and the 272/273 allow-lists each need an exemption path for `examples/phase30/` (below).

## Session meta-tests and registers that will move (research item 8)

| Register / test | Location | What moves |
|-----------------|----------|-----------|
| `TestPhase28EntryExitClassManifest` | `session_phase28_manifest_test.go`; `phase28MinCorpus = 155`; `phase28NewlyAddedEntries` | Every committed `.schway` under `examples/` or `testdata/` needs an entry. The walker invokes the FILE routes, so import-bearing files exercise the loader here too. |
| `TestPhase272...` / `TestPhase273...` | `phase272NewlyAcceptedPaths` / `phase273NewlyAcceptedPaths` | The library module alone checks clean: add `examples/phase30/adler32.schway`. The entry with imports is accepted only through the file route, so `session.Check(source)` refuses it and it never enters these walks. |
| `TestLanguageMaturityCountsAreCurrent` | `self_describing_docs_test.go`; sentence `**159 \`.schway\` programs, 5,196 lines total**` at `.planning/LANGUAGE-MATURITY.md:302` | Counts every `.schway` in the tree (skips only `.git`) and the `len(Functions) != 1` guard table; do not add such guards. |
| `TestPhase16PublicEmitterConsumerInventory` | `testdata/phase16/public-emitter-consumers.json`, rows keyed `file:Emit|EmitNative:line` (`evidence.go:Emit:211`, `session.go:EmitNative:3049`, `session_phase5.go:27`, ...) | Any edit that shifts those lines (every `session.go` edit above line 3049; any `evidence.go` edit above 211) requires re-pinning the row. Put loader wiring in a new `session_modules.go`, keep `session.go` edits minimal, re-pin once at the end of 30-08. Do not touch `evidence.go` in this phase. |
| `TestEntryFunctionCalledOnlyViaResolveEntry` | `session_entry_inventory_test.go` | `callgraph.EntryFunction` is called once in `session_entry.go` and once in `cgen_program.go`; module filtering must live INSIDE `EntryFunction`, no new call site. |
| `TestSessionRunSitesDoNotIndexFunctionsZero` | `callgraph_entry_test.go:436` | Do not index `Functions[0]`. |
| Groundedness / `TestPhase20ValidationLifecycle` | `verification_groundedness_test.go:1495,1031` | `30-VALIDATION.md` must not stay `status: draft`; `go test ... -run` rows in enforced documents (VALIDATION, VERIFICATION, SUMMARY, UAT) must resolve to real tests (static index skips `testdata` dirs). RESEARCH and PLAN are exempt only via `verification_role: proposal` and must not put a verdict token beside a command span. |
| `TestRepairFieldEnumerationIsExhaustive`, `TestDiagnosticZeroBytesUnchanged` | `diagnostic_test.go:176,243` | Update the field list with `Unit` classified non-identity; the pinned ID stays. |
| `TestGeneratedIdentifierNamespacesStayConfined`, `TestNativeIdentifiersRemainCollisionFree`, `cgen_names_test.go` | cgen | Must stay green; extend with the module-name universe. |
| `TestExitCodeDocsTable`, `TestCIWorkflowRunsCurrentAggregateGate` | session | Untouched unless a new exit code or a CI step is introduced (neither is). |
| `internal/compiler/core` tests | `core_test.go` | No new `OperationKind`; `TestAllOperationKinds*` unchanged. |

**Corollary:** generate every graph/cap/adversarial fixture under `t.TempDir()` (D-13 already says so); commit only `examples/phase30/adler32.schway`, `examples/phase30/adler32_cli.schway` (+ README). That keeps registers to: two `phase28NewlyAddedEntries` rows, one 272 and one 273 allow-list row, one LANGUAGE-MATURITY count/line update, plus the emitter-inventory re-pin.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Path traversal containment | prefix checks on joined paths, `EvalSymlinks` | `os.OpenRoot` + per-component `Root.Lstat` | Probe: `Root` rejects `..`, absolute and escaping symlinks; `Lstat` per component refuses in-root links |
| Bounded reads | `Stat().Size()` gating | `io.LimitReader(file, limit+1)` (`readBoundedFile`, `session.go:1890-1897`) | Size from `Stat` is the named mutant |
| Cycle detection | recursion with a visited bool | iterative three-color DFS copied from `callgraph.Order` | Gray/black distinction separates cycles from diamonds; recursion depth is attacker-controlled |
| Exact-case file lookup | `EqualFold` guess, `Lstat().Name()` | `ReadDir` of the parent, byte-equal compare | APFS echoes the requested name (probe) |
| Diagnostic ID with unit | ad hoc hashing | extend the two existing identity structs | One ID law, existing pinned control |
| Name mangling | custom separators | D-12 length-prefix scheme + the existing `cNames.allocate` suffix | Injective by construction; collision suffix already deterministic |
| Reading the same file twice | re-open for hashing | hash the in-memory bytes | D-10 |

## Common Pitfalls

### Pitfall 1: The byte-taking `Check` and its cousins
**What goes wrong:** an import-bearing file given to `evidence`, `explain`, `query`, `interface`, corpus verify parses fine and `check.Program` reports a bare unresolved-callee error (or, worse, ignores the import block). **Avoid:** `check.Program` refuses `Imports` without `Units` with one named code (recommended `module.imports_not_loaded`), pinned for each command by `TestPhase30SingleFileCommandsRefuseImports`.

### Pitfall 2: Every committed `.schway` moves five registers
See table above. **Avoid:** TempDir fixtures; register the two committed files in the same plan that adds them.

### Pitfall 3: `cloneProgram` silently drops untagged core fields
`corevalidate.Validate(p).Program()` JSON round-trips (`corevalidate.go:5225-5235`). A field without a `json` tag name is lost; `Module` and `Exported` vanish and every library function looks like an entry function. **Test:** `TestPhase30ValidatedProgramKeepsModuleFacts` (mutant: `json:"-"`).

### Pitfall 4: Name-keyed lookups
`callgraph.EntryFunction` collects `function.Name == "main"` across ALL functions (`callgraph.go:328-340`): a library with its own `main` becomes a duplicate-main refusal; unimported exported functions would be extra in-degree-zero roots (`:304-310`). `interp.findFunction` returns the first function with the name (`interp.go:1977-1984`), called with `entry.Name` from `session.go:1143,1306`. **Avoid:** restrict candidates to `Module == ""` inside `EntryFunction` and `findFunction`; `debugmap` key by (module, name).

### Pitfall 5: Unbounded merged program
Measured above: 26.7 s `corevalidate` at 16,384 functions. **Avoid:** merged cap (`module.program_limit`) plus token summing into `Work`.

### Pitfall 6: `os.Root` semantics (probed on go1.24.0 darwin/arm64 this session)
- `Root.Lstat("lib.schway")`, `"Lib.schway"`, `"LIB.SCHWAY"` all succeed on APFS (case-insensitive); `Lstat` does not verify case.
- `Root.Lstat("subl/n.schway")` where `subl -> sub` (in-root dir link) returns a regular file: only the last component is not followed. Lstat each component.
- `Root.Lstat("alias.schway")` on an in-root file link returns `Lrwxr-xr-x`; `Root.Open("subl/n.schway")` opens through the in-root dir link.
- `Root.Open` of an escaping symlink, `outdir/secret.schway` (dir link out), `../outside/secret.schway` and `/etc/hosts` all fail with `path escapes from parent`; that error satisfies none of `fs.ErrInvalid`, `fs.ErrPermission`, `fs.ErrNotExist` (it is a `*fs.PathError` with an unexported sentinel), so classify by pre-check elimination, never by error text. Missing files satisfy `fs.ErrNotExist`.
- `Root.Open` of a FIFO blocks (timed out at 2 s): require `Lstat(...).Mode().IsRegular()` first.
- There is no `Root.ReadDir` in 1.24: `root.Open(dir)` then `(*os.File).ReadDir(-1)` lists exact names.
- On a case-insensitive FS you cannot create two case variants: the both-spellings case is Linux-only (`t.Skip` after a creation probe), while "only `Lib.schway` exists, import `lib`" is `module.case_mismatch` on both OSes.
- Never format `*fs.PathError` text into a diagnostic (carries relative or absolute paths); map to codes and use module names.

### Pitfall 7: `..` and absolute paths are unrepresentable
`import` takes `modulePath` = identifiers joined by dots (`parser.go:258-264`); `..x`, `/etc/x`, `a/b` hit `syntax.expected_module_name` or `syntax.unexpected_byte` in the parser/lexer before any filesystem access. Test them as syntax refusals (SC1) AND test real filesystem escapes through symlinks (the only live vector). The lexer accepts Unicode letters (`lexer.go:137-141`), so `import café {` lexes; `module.invalid_segment` must fire in modload (D-07).

### Pitfall 8: Entry reached by its own path
`entry -> a -> entry` resolves to the entry file under its path-derived name; header differs, so naive loading reports `module.path_mismatch` (and re-reads the entry). **Avoid:** compare resolved files to the entry with `os.SameFile` before reading and treat the hit as an edge to the entry node so cycles through the entry are `module.import_cycle`. Self-import for the entry file is then also a cycle; a non-entry self-import (`a` imports `a`) is length 1.

### Pitfall 9: Exact-N for imports-per-module is unreachable end to end
16 valid import declarations need 16 distinct modules plus the entry = 17 > the 16-module cap; duplicates are refused as duplicates. **Test** the 16/17 boundary at the parser (`syntax.Parse` of 16 vs 17 import blocks, `module.imports_limit` only on 17) where the `>` -> `>=` mutant is killable; count (15 leaves + entry = 16 accepted, 16 leaves refused), depth (chain of 8 accepted, 9 refused), total bytes (4 x 1,048,576 accepted; a fifth file of 1 byte refused; pad with one `//` comment line so `MaxTokens` is not hit) are end-to-end.

### Pitfall 10: Repair shapes
`diagnostic.DriverEligible` requires `Replacement != ""` (`diagnostic.go` / `repair.go:273`), so "delete the later duplicate import" must replace the declaration with `"\n"`, not `""`; offer it only when the later item set is a subset of the earlier's (else deleting drops names). "Move import above export" is one span covering `[exportStart, importEnd)` with the rotated text. Both `MachineApplicable`; `Kind` participates in the ID, `Span`/`Replacement` do not.

### Pitfall 11: `__` in C names
The D-07 segment charset `[a-z][a-z0-9_]*` allows `a__b` and a trailing `_`; function names may start with `_`. Under D-12 `SCHWAY_M_2_A__3_FOO` contains `__` (C17 7.1.3). Resolve at the gate (Open Question 1); the adversarial enumeration test must include these.

### Pitfall 12: Peer diagnostics have no span or unit
`corevalidate.Problem` is `{Code, Detail}` (`corevalidate.go:84-87`) and `peerRefusalDiagnostic` uses `diagnostic.Span{}`. Attribute unit from the ID in `Detail`; accept span-less.

### Pitfall 13: The 29-style "Callable" body dependence
`Callable` is derived from the checked body (`originvalidate.PublishProblemsFor`); the SC2 guard test must hold `Callable` fixed (see body-blind section).

### Pitfall 14: Staleness claims
Build receipts bind `SourceDigest` of the entry bytes only; library edits change `EmittedCDigest` and therefore build identity (`native_app.go:211-214`), but `evidence` source digest is entry-only. State this as a non-claim until Phase 31.

## Code Examples

### Per-component resolution (sketch; the loader's core)
```go
// root, err := os.OpenRoot(entryDir)
// for i, seg := range segments {                 // segments validated [a-z][a-z0-9_]*, <=64 bytes
//     rel := path.Join(segments[:i+1]...) ; last := i == len(segments)-1
//     if last { rel += ".schway" }
//     info, err := root.Lstat(rel)
//     switch { case errors.Is(err, fs.ErrNotExist): return notFoundOrCase(root, parent, name)
//              case err != nil: return module.escape/io
//              case info.Mode()&fs.ModeSymlink != 0: return module.symlink
//              case !last && !info.IsDir(): return module.not_found
//              case last && !info.Mode().IsRegular(): return module.not_regular }
//     exactCase(root, parentDir, name)            // ReadDir via root.Open(parent): byte-equal or module.case_mismatch
// }
// f, _ := root.Open(rel); st, _ := f.Stat(); if !os.SameFile(lstatInfo, st) { module.unstable_file }
// data, _ := io.ReadAll(io.LimitReader(f, int64(syntax.MaxSourceBytes)+1))   // never from st.Size()
```
(Source: Go `os.Root` docs and the probe above; codes other than those in CONTEXT are discretion.)

### Identity struct with Unit (D-11)
```go
// diagnostic.go: Error()
identity := struct {
    Schema string; Code string; Span Span; Causes []Cause
    Unit string `json:",omitempty"`
}{Schema: Schema, Code: code, Span: span, Causes: causes, Unit: unit}
// Cause gains: Unit string `json:"unit,omitempty"`   (inherits Diagnostic.Unit when empty)
```
Verified existing shape: `diagnostic.go:109-114`.

### Stamping per function in `check.Program`
```go
starts := make([]int, len(program.Funcs)+1)
for i, function := range program.Funcs {
    starts[i] = len(result.Diagnostics)
    // ... existing body unchanged, with module := function.Module used for IDs ...
}
starts[len(program.Funcs)] = len(result.Diagnostics)
for i := range program.Funcs {
    for j := starts[i]; j < starts[i+1]; j++ { result.Diagnostics[j] = result.Diagnostics[j].InUnit(program.Funcs[i].Module) }
}
```
`InUnit` recomputes the ID with the Unit-bearing identity struct (empty unit returns the diagnostic unchanged, byte-identical).

### EntryFunction module filter
```go
// callgraph.go: candidates are entry-module functions only
if function.Name == "main" && function.Module == "" { mains = append(mains, function.ID) }
// roots: keep only ids whose function has Module == "" (in-degree still computed over the whole graph)
```
No-import programs have `Module == ""` everywhere: identical behavior.

### corevalidate visibility law (at the existing CalleeID site)
```go
if callee, ok := functionByID[operation.CalleeID]; ok && callee.Module != function.Module {
    if !v.check(callee.Module != "" && callee.Exported, "core.callee_not_exported", operation.CalleeID) { return false }
    if !v.check(v.programImports(function.Module, callee.Module, callee.Name), "core.callee_not_imported", operation.CalleeID) { return false }
}
```

### D-12 name function (pure, testable)
```go
// moduleCName("a.b", "f") == "SCHWAY_M_1_A_1_B_1_F"; moduleCName("adler32", "adler32") == "SCHWAY_M_7_ADLER32_7_ADLER32"
func moduleCName(module, fn string) string {
    var b strings.Builder
    b.WriteString("SCHWAY_M")
    for _, seg := range strings.Split(module, ".") { fmt.Fprintf(&b, "_%d_%s", len(seg), strings.ToUpper(seg)) }
    fmt.Fprintf(&b, "_%d_%s", len(fn), strings.ToUpper(fn)); return b.String()
}
// Allocation: pass 1 entry functions: globals.allocate(cName(f.Name), "function", index)  (unchanged, entry wins)
//             pass 2 imported functions sorted by (Module, Name): globals.allocate(moduleCName(...), "function", position)
```
Injectivity: each `(len, text)` pair is self-delimiting; the universe test must still enumerate `a_b.c` vs `a.b_c`, uppercase fn variants (`Foo` vs `foo` collapse under `ToUpper` and rely on `allocate`'s suffix), digit-leading names, `__`/trailing `_` segments, and entry functions literally named like a mangled name.

### Import grammar entry in the parser
```go
// parseProgram: after p.modulePath()
for p.peek().Kind == TokenImport { imports = append(imports, p.importDecl()) }   // cap: len(imports) > 16 -> module.imports_limit
program.Exports = p.exports()
// decl loop: case TokenImport => syntax.import_after_export (parse the block for recovery), repair = rotate
```
Items: `fn IDENT` or `type IDENT` (type kept in the AST so `check` refuses it by name); reuse the `looksLikeFunctionDecl()` break used by `exports()` (`parser.go:270-273`) for unclosed blocks.

## Runtime State Inventory

Not applicable: this is not a rename/refactor/migration phase. The one string that becomes semantic is the `import` identifier (now a keyword); a repo-wide search found no `.schway` file or Go test source using `import` as an identifier [VERIFIED: grep of `(let|var|fn|take|borrow) +import` and `.schway` content this session, no matches].

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `filepath` prefix checks / `EvalSymlinks` for containment | `os.Root` | Go 1.24 | Traversal-resistant opens; still follows in-root symlinks and cannot verify case |
| ARCHITECTURE C-name row `module__name` | D-12 length-prefixed `SCHWAY_M_...` | 2026-10-12 (CONTEXT) | No `__`, injective |
| STACK.md 64 KiB per-file / strip-suffix root | D-13 caps, D-05 entry-dir root | 2026-10-12 | Superseded |

**Deprecated/outdated:** the ARCHITECTURE reference to a `publishedOriginProblemFile` re-read (symbol absent in the current tree).

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Tighten imported-module segment/function-name grammar (no `__`, no trailing `_`, no leading `_`) to honor "no new `__`" under D-12 | Pitfall 11, Open Q1 | Either C names contain `__` or a locked decision (D-07 charset) is narrowed; needs user/gate confirmation |
| A2 | `module.duplicate` = imported name equals entry header; same module imported twice in one file = separate `module.duplicate_import` (repair: delete later) | Open Q2 | MOD-05 "duplicate module" fixture list could be read differently |
| A3 | Refuse `data` and `foreign` declarations inside imported modules (codes e.g. `module.imported_data_unsupported`, `module.imported_foreign_unsupported`) | Merging | Scope narrower than a reader of D-02 expects; alternative is module-scoped types (large) |
| A4 | Refuse importing any function named `main` (D-04 only names the collision with the entry's `main`) | Open Q4 | Over-refusal vs. D-04 literal |
| A5 | Merged-program cap `module.program_limit` at 1,024 functions / 4,096 declarations; constants exported from `syntax` | Caps | D-13 leaves the code to the gate; a different code is cosmetic, no cap is a 27 s admission |
| A6 | `debug-map` routed through the loader; `evidence`/`explain`/`query`/`interface` refuse imports with a named code (`module.imports_not_loaded`) | Where the loader plugs in | If the user expects those commands to work on module programs this phase, scope grows |
| A7 | Human projection `ID CODE unit=<module> [s:e]: message` | Unit plumbing | CONTEXT wording is "a `unit:` prefix"; cosmetic but published |
| A8 | Peer-refusal unit attribution from the module segment of the ID in `Problem.Detail` | Pitfall 12 | Imprecise attribution for peer refusals |
| A9 | corevalidate codes `core.callee_not_exported` / `core.callee_not_imported`; check code `module.not_exported` | corevalidate | Naming only |
| A10 | Cap directory listings (e.g. 65,536 entries examined) so a hostile huge directory cannot exhaust memory during the D-09 listing; refusal code discretionary | Security | Unbounded `ReadDir(-1)` on attacker-controlled directory |
| A11 | Timing numbers are host-specific (darwin/arm64, this session); CI Linux may differ | Caps | Only counts are pinned, never times |
| A12 | SEED-261008-z1g (bounded symbol/caller query): retain dormant at this transition; no measured reviewer gap, query stays single-file | Open Q | Planner must record the disposition in the active planning artifact |
| A13 | Cycle through the entry detected by `os.SameFile` against the entry file | Pitfall 8 | Entry given through an unusual path (symlink) could fail identity; entry is user authority |
| A14 | The PRODUCT-ROADMAP / LANGUAGE-MATURITY transition read required by AGENTS.md was only skimmed (grep for module claims), not executed as a full transition check | Project constraints | Planner should perform it |

## Open Questions

1. **`__` in C names (A1).** What we know: D-07 allows `a__b` and trailing `_`; fn names may lead with `_`; D-12 forbids new `__`. Unclear: whether narrowing the grammar is acceptable. Recommendation: gate flag (d) decides; refuse such imported segments (`module.invalid_segment`) and imported function names (new code) rather than escape.
2. **Which situations are `module.duplicate`?** CONTEXT assigns (a) imported == entry header. The same-module-twice case needs its own code (A2). Planner to confirm.
3. **Imported `data`/`foreign` (A3).** Recommend refusal with named codes; no CONTEXT text covers it.
4. **Importing `main` (A4).** Recommend always refuse.
5. **Merged cap (A5).** Recommend `module.program_limit`; gate measures and writes the verdict.
6. **Which commands accept module programs (A6).** Recommend check/run/build/app verify/debug-map; the rest refuse by name.
7. **Human `unit` rendering (A7).**
8. **Directory scan bound (A10).**
9. **Plan-count at ceiling.** Ten plans exactly; if 30-00 returns SPLIT or another deliverable lands, D-16 moves plans 05-09 to Phase 30.1.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | everything | yes | go1.24.0 darwin/arm64 (go.mod `go 1.24`) | none needed |
| clang | native lane, runnable gain | yes | `/usr/bin/clang` (Apple); CI requires clang on both OSes | tests `t.Skip("env:clang ...")` like `main_phase29_test.go:21-23` |
| FIFO / symlink creation | filesystem controls | yes on macOS and Linux (`syscall.Mkfifo`, `os.Symlink`) | stdlib | `//go:build unix`; skip elsewhere |
| Case-sensitive filesystem | both-spellings case test | no on this host (APFS, probe), yes on ubuntu CI | | creation probe then `t.Skip` |

**Missing dependencies with no fallback:** none. **With fallback:** case-sensitive FS locally (Linux CI covers it).

## Validation Architecture

> `workflow.nyquist_validation` is true in `.planning/config.json`. Zero human UAT (PROJECT.md § Verification Operating Preference). Every test name below is new unless marked "exists".

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing`, go1.24.0 |
| Config file | none (go.mod only) |
| Quick run command | `go test ./internal/compiler/modload ./internal/compiler/syntax ./internal/compiler/diagnostic ./internal/compiler/protocol ./internal/compiler/debugmap ./internal/compiler/check ./internal/compiler/core ./internal/compiler/corevalidate ./internal/compiler/callgraph ./internal/compiler/interp ./internal/compiler/cgen ./internal/compiler/session ./cmd/schway ./cmd/schway-repair -run 'Phase30' -count=1` |
| Session meta set | `go test ./internal/compiler/session -run 'Inventory\|Groundedness\|ValidationLifecycle\|TestPhase28EntryExitClassManifest\|TestPhase272AcceptedCorpusByteIdentity\|TestPhase273ConformanceCorpusByteIdentity\|TestLanguageMaturityCountsAreCurrent\|TestExitCodeDocsTable\|TestEntryFunctionCalledOnlyViaResolveEntry\|TestPhase16PublicEmitterConsumerInventory' -count=1` |
| Full suite command | hosted CI only (`go test ./...` and `go test -race -timeout=20m ./...` on ubuntu-latest and macos-latest); never run the full suite locally (project policy, memory note on focused runs) |

### Success Criterion / Decision -> Test Map

| ID | Behavior | Test (package) | Command | Armed control / mutant |
|----|----------|----------------|---------|------------------------|
| SC1 / D-17 | Adler-32 imported from `examples/phase30/adler32.schway` prints pinned answers 0..9 on native (CLI `build --manifest`, `app run -- N`) | `TestPhase30AdlerModuleAppCLI` (cmd/schway; clang gated) | `go test ./cmd/schway -run 'TestPhase30AdlerModuleAppCLI' -count=1` | library mutated to a wrong modulus changes the answers (hosted mutation lane is Phase 31; here: expected-value table is `testsupport.Phase29FixtureExpected`, a wrong table row must fail) |
| SC1 | Same answers on the interpreter via `CheckFile` + `interp.RunWithFixtureBytes` | `TestPhase30AdlerModuleInterpreter` (session) | `go test ./internal/compiler/session -run 'TestPhase30AdlerModuleInterpreter' -count=1` | inline-copy of library into entry must fail the single-copy test below |
| SC1 | Library exists once; no copy | `TestPhase30LibraryIsSingleCopy` (session) | `go test ./internal/compiler/session -run 'TestPhase30LibraryIsSingleCopy' -count=1` | mutant: copy `fn adler32(` into the entry file |
| SC1 | `..`, absolute, path-separator imports refused as syntax errors | `TestPhase30ImportPathSyntaxRefused` (syntax) | `go test ./internal/compiler/syntax -run 'TestPhase30ImportPathSyntaxRefused' -count=1` | accepting `a/b` mutant |
| SC1 / D-08 | Escaping symlink file, escaping dir symlink, in-root symlink -> `module.escape` / `module.symlink` | `TestPhase30SymlinkAndEscapeRefused` (modload) | `go test ./internal/compiler/modload -run 'TestPhase30SymlinkAndEscapeRefused' -count=1` | M-ESC1: skip per-component `Lstat`; M-ESC2: `os.Open(filepath.Join(...))` without `Root` |
| D-10 | FIFO refused without blocking (deadline 2 s) | `TestPhase30FIFORefusedWithoutBlocking` (modload, unix) | `go test ./internal/compiler/modload -run 'TestPhase30FIFORefusedWithoutBlocking' -count=1` | M-FIFO: open before `Lstat` -> the naive path blocks past the deadline |
| D-09 | Case: only `Lib.schway` + `import lib` -> `module.case_mismatch`; Linux both spellings loads exact | `TestPhase30CaseExactMatch` (modload) | `go test ./internal/compiler/modload -run 'TestPhase30CaseExactMatch' -count=1` | M-CASE: `EqualFold` accept |
| D-10 | TOCTOU: file identity changes between `Lstat` and `Open` -> refused; size never from `Stat` | `TestPhase30FakeFSToctouAndStatSize` (modload) | `go test ./internal/compiler/modload -run 'TestPhase30FakeFSToctouAndStatSize' -count=1` | M-STAT: size from `Stat`; M-SAME: drop `SameFile` |
| D-10 / SC4 | Each file read exactly once; single IO site | `TestPhase30LoaderReadsEachFileOnce` (modload), `TestPhase30SingleIOSite` (session AST scan) | `go test ./internal/compiler/modload ./internal/compiler/session -run 'TestPhase30LoaderReadsEachFileOnce\|TestPhase30SingleIOSite' -count=1` | M-REREAD: re-open for hashing -> counter 2 |
| SC2 | Private name import refused `module.not_exported` with `Unit` = importer, span = item | `TestPhase30PrivateNameRefused` (check) | `go test ./internal/compiler/check -run 'TestPhase30PrivateNameRefused' -count=1` | seam admits private import |
| SC2 | corevalidate refuses a forged cross-module call (private callee, missing import fact, entry-module callee) with check disabled | `TestPhase30CorevalidateVisibilityPeer` (corevalidate) | `go test ./internal/compiler/corevalidate -run 'TestPhase30CorevalidateVisibilityPeer' -count=1` | M-VIS: peer ignores `Exported`; `disableVisibilityPeerForTest` shows check alone is not the only refusal |
| SC2 | Validated program keeps `Module`/`Exported`/import facts | `TestPhase30ValidatedProgramKeepsModuleFacts` (corevalidate) | `go test ./internal/compiler/corevalidate -run 'TestPhase30ValidatedProgramKeepsModuleFacts' -count=1` | M-CLONE: `json:"-"` |
| SC2 | Body-blind: equal-signature body change leaves importer admission identical; observer zero | `TestPhase30CallAdmissionBodyBlind` (check) | `go test ./internal/compiler/check -run 'TestPhase30CallAdmissionBodyBlind' -count=1` | M-BODY: `scopedContractBodyReadSeam`; existing `verifyCallableRefusalBodyReadSeam` on module fixture |
| SC2 | Converse: Callable-flipping body change refuses the caller | `TestPhase30CallableFlipIsASignatureChange` (check) | same package, `-run 'TestPhase30CallableFlipIsASignatureChange'` | n/a (boundary doc) |
| SC3 | Self, 2-cycle, 3-cycle, cycle through entry each `module.import_cycle` attributed to importer, distinct fixtures; canonical rotation | `TestPhase30ImportCycles` (modload) | `go test ./internal/compiler/modload -run 'TestPhase30ImportCycles' -count=1` | M-CYC: visited-set instead of gray/black (diamond falsely refused or cycle missed) |
| SC3 | Missing -> `module.not_found`; header mismatch -> `module.path_mismatch`; imported name == entry header -> `module.duplicate`; invalid segment (uppercase, unicode, 65 bytes) | `TestPhase30GraphRefusals` (modload) | `go test ./internal/compiler/modload -run 'TestPhase30GraphRefusals' -count=1` | each case has a sibling accepted fixture |
| SC3 | Collisions: import vs local, import vs import, imported `main`, reserved builtin; message names both origins; no repair | `TestPhase30ImportCollisions` (check) | `go test ./internal/compiler/check -run 'TestPhase30ImportCollisions' -count=1` | seam disables the pair |
| SC3 | Diagnostic raised inside an imported file names that file; same offset in two files -> distinct IDs; entry diagnostic has empty unit | `TestPhase30ImportedFileDiagnosticNamesFile` (session) and `TestPhase30UnitIdentity` (diagnostic) | `go test ./internal/compiler/session ./internal/compiler/diagnostic -run 'TestPhase30ImportedFileDiagnosticNamesFile\|TestPhase30UnitIdentity' -count=1` | M-ATTR: skip stamping; M-ID: unit left out of the identity struct |
| D-11 | Existing IDs unchanged; field enumeration updated; Cause/Repair inherit | `TestPhase30UnitOmitempty` (diagnostic) plus existing `TestDiagnosticZeroBytesUnchanged`, `TestRepairFieldEnumerationIsExhaustive` | `go test ./internal/compiler/diagnostic -run 'TestPhase30UnitOmitempty\|TestDiagnosticZeroBytesUnchanged\|TestRepairFieldEnumerationIsExhaustive' -count=1` | M-OMIT: drop omitempty -> pinned ID changes |
| D-11 | Human `unit=` only when set; JSON `unit` | `TestPhase30UnitProjection` (protocol) | `go test ./internal/compiler/protocol -run 'TestPhase30UnitProjection' -count=1` | mutant prints always |
| D-11 | Repair driver declines cross-unit repair; no entry-file write | `TestPhase30RepairDriverDeclinesCrossUnit` (cmd/schway-repair) | `go test ./cmd/schway-repair -run 'TestPhase30RepairDriverDeclinesCrossUnit' -count=1` | M-UNIT: driver ignores `unit` -> entry bytes change |
| D-11 | debug-map entries carry `unit`, keyed by (module, name) | `TestPhase30DebugMapUnit` (debugmap/session) | `go test ./internal/compiler/debugmap ./internal/compiler/session -run 'TestPhase30DebugMapUnit' -count=1` | mutant keys by bare name |
| SC4 | Diamond: shared module read once, one core function, one C definition | `TestPhase30DiamondLoadsAndEmitsOnce` (session) | `go test ./internal/compiler/session -run 'TestPhase30DiamondLoadsAndEmitsOnce' -count=1` | M-DIA: diamond counted twice (count/dup) |
| SC4 / D-15 | No-import programs byte-identical across check JSON, core, C (both emitters), debugmap, format | `TestPhase30PreChangeGolden` (session) + existing 272/273 | `go test ./internal/compiler/session -run 'TestPhase30PreChangeGolden\|TestPhase272AcceptedCorpusByteIdentity\|TestPhase273ConformanceCorpusByteIdentity' -count=1` | `TestPhase30PreChangeGoldenIsNotInert` |
| SC4 / D-12 | C names injective over the adversarial universe; reversed import order -> identical C; entry names win; namespace confined | `TestPhase30CNamesInjective`, `TestPhase30ReversedImportOrderIdenticalC` (cgen/session) | `go test ./internal/compiler/cgen ./internal/compiler/session -run 'TestPhase30CNamesInjective\|TestPhase30ReversedImportOrderIdenticalC' -count=1` plus existing `TestGeneratedIdentifierNamespacesStayConfined` | M-JOIN: plain `_`/`__` join -> collision found by enumeration |
| D-13 | Caps N accepted / N+1 refused: count 16/17, depth 8/9, total bytes 4,194,304/+1, per-module `syntax.input_limit` naming the unit, imports 16/17 (parser level), merged program 1,024/1,025 functions | `TestPhase30CapsExactNAndNPlusOne` (modload), `TestPhase30ImportsLimitParse` (syntax) | `go test ./internal/compiler/modload ./internal/compiler/syntax -run 'TestPhase30CapsExactNAndNPlusOne\|TestPhase30ImportsLimitParse' -count=1` | `exceedsOffByOneForTest` (`>` -> `>=`) per cap; M-STAT; M-OPEN: count check after open (open counter shows 17th opened); M-DIA |
| D-02/D-03 | Grammar: empty list, import after export, `import` as identifier, `type` item, no commas; formatter golden, idempotent, lossless, source order kept | `TestPhase30ImportGrammar`, `TestPhase30FormatImport`, `TestPhase30GeneratedImportRoundTrips`, extend `FuzzParseFormat` seeds (syntax) | `go test ./internal/compiler/syntax -run 'Phase30' -count=1` | M-SORT: formatter sorts imports -> semantic-token-order test fails |
| D-04 | Repairs only for duplicate import and import-after-export; deletion uses `"\n"`; subset rule | `TestPhase30ImportRepairs` (syntax/check) | `go test ./internal/compiler/syntax ./internal/compiler/check -run 'TestPhase30ImportRepairs' -count=1` | repair applied to a non-subset duplicate must not be offered |
| Entry | Library `main` and unimported exports are never roots; `findFunction` prefers entry module | `TestPhase30EntryIgnoresLibraryMain` (callgraph), `TestPhase30FindFunctionPrefersEntryModule` (interp) | `go test ./internal/compiler/callgraph ./internal/compiler/interp -run 'TestPhase30EntryIgnoresLibraryMain\|TestPhase30FindFunctionPrefersEntryModule' -count=1` | filter removed -> `core.entry_ambiguous` |
| Fail-closed | Single-source commands refuse imports by name | `TestPhase30SingleFileCommandsRefuseImports` (session) | `go test ./internal/compiler/session -run 'TestPhase30SingleFileCommandsRefuseImports' -count=1` | mutant lets `check.Program` ignore `Imports` |
| Security | Diagnostics never contain the temp-dir absolute path or `PathError` text | `TestPhase30ErrorsNeverLeakAbsolutePaths` (modload/session) | `go test ./internal/compiler/modload ./internal/compiler/session -run 'TestPhase30ErrorsNeverLeakAbsolutePaths' -count=1` | mutant formats `err.Error()` |
| Gate | flags (a)-(f) with test-name evidence | `TestPhase30Gate*` (various) | `go test ./internal/compiler/modload ./internal/compiler/diagnostic ./internal/compiler/cgen ./internal/compiler/check ./internal/compiler/session -run 'TestPhase30Gate' -count=1` | each flag test carries its own control |
| Meta | registers, counts, inventory, groundedness | Session meta set above | see "Session meta set" | n/a |

### Sampling Rate
- **Per task commit:** focused `-run 'Phase30'` for the touched package; the session meta set whenever a `.schway` file, a `session.go`/`evidence.go` line, or a planning document with a `go test` row changed.
- **Per wave merge:** all `Phase30` rows for the wave's packages plus the three byte-identity tests.
- **Phase gate:** hosted full and race CI green on both OSes before `/gsd-verify-work` (record run ids; no remote access means an external-blocker note, not a gap plan).

### Wave 0 Gaps
- [ ] `testdata/phase30/prechange-corpus.json` + `TestPhase30PreChangeGolden` committed first, alone.
- [ ] Gate tests (a)-(f) and `30-00-GATE-VERDICT.md` in the Phase 29 shape (flags table, test names as evidence, `verdict: PROCEED | SPLIT`).
- [ ] `modload` filesystem interface seam (fake FS) before any graph code, so every cap/TOCTOU control is armable.
- [ ] `30-VALIDATION.md` finalized (not `draft`) before any hosted dispatch.

## Security Domain

### Applicable ASVS Categories (Level 1; `security_enforcement` true, block_on high)

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V1 Architecture / trust boundaries | yes | Loader is the only IO site; peers do not trust the loader; coordinated-lie boundary stated |
| V5 Input Validation | yes | Module segment grammar (ASCII, lowercase, 64 bytes), header equality, bounded reads, parser caps |
| V7 Error Handling and Logging | yes | Root-relative module names only; no `PathError` text or absolute paths in diagnostics |
| V12 Files and Resources | yes | `os.OpenRoot`, per-component `Lstat`, regular-file check before open, `SameFile` after open, size caps (per module, total, count, depth), no FIFO/device/symlink |
| V2/V3/V4/V6 | no | no auth, session, access-control or crypto surface |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Path traversal via import text | Tampering / Information disclosure | Unrepresentable in grammar; segment charset; `Root` |
| Symlink escape or alias (in-root link double-loads a file) | Tampering | Per-component `Lstat` refuses all symlinks; `Root` second layer |
| FIFO/device open hang | Denial of service | `IsRegular()` before `Open`, deadline test |
| TOCTOU file swap | Tampering | `SameFile(Lstat, Stat)` after open; single read |
| Case-fold / homoglyph module aliasing | Spoofing | Lowercase ASCII segments; exact-case `ReadDir` match |
| Zip-bomb style fan-out and deep chains | Denial of service | 16 modules, depth 8, 4 MiB total, 1 MiB each, merged 1,024 functions, iterative DFS |
| Quadratic validator cost via merged size | Denial of service | Merged-program cap (measured 26.7 s at 16,384 functions) |
| Huge directory listing | Denial of service | Bounded scan (A10) |
| Repair driver patching the wrong file | Tampering | Cross-unit repair declined; `unit` decoded structurally |
| Forged core claiming visibility | Spoofing | corevalidate law; documented coordinated-lie boundary |
| Test seams reachable in production | Elevation | Unexported `...ForTest` vars, same-package tests only (existing convention) |

## Project Constraints (from CLAUDE.md / AGENTS.md)

- Project instructions live in `AGENTS.md`; `CLAUDE.md` adds skill routing only (`Skill("spike-findings-ai-lang")` loaded). Findings applied: affine ownership and release order untouched (modules add no operation kinds); loan liveness remains a per-function CFG fact computed after merge; the signature table is the mechanism for body-blind admission (analogous to the memoized summary table); independently written implementations plus armed fault injection (one seeded fault per peer, a coordinated lie escapes by design and is stated); distinct evidence lanes.
- Go 1.24 stdlib only; text source formatter-owned; readable C17 to installed Clang; stable machine-readable diagnostics; macOS and Linux initial hosts (wire contracts must not encode the current host).
- GSD workflow enforcement: file changes go through a GSD command.
- Living-document rule: planning transition must read `PRODUCT-ROADMAP.md` / `LANGUAGE-MATURITY.md` and scan `.planning/seeds/`; this research only skimmed them (A14, A12). Modules stay "whole-program composition", never separate compilation or closing B1/DX-06.
- "Prefer a runnable gain in each feature phase" (the Adler-32 library import is it); no new backend; no invented completeness percentages.
- Focused tests only locally; full suites on hosted CI (28-CONTEXT and memory note).

## Suggested Plan Decomposition (advisory; ceiling 10, D-16)

| Plan | Wave | Scope | depends_on |
|------|------|-------|------------|
| 30-00 | 0 | FIRST commit: D-15 pre-change golden (+ non-inert control). Then gate tests (a) `os.Root`/FIFO, (b) case, (c) `Unit` consumer inventory + ID invariance, (d) naming enumeration, (e) body-blind tracer, (f) merged-cost measurement; write `30-00-GATE-VERDICT.md` | none |
| 30-01 | 1 | `syntax`/`ast`: `TokenImport`, `ImportDecl/ImportItem`, parser errors (`expected_import_name`, `import_after_export` + rotate repair, `module.imports_limit`), formatter, extended round-trip/fuzz lanes, exported `MaxFunctions/MaxDeclarations` | 30-00 |
| 30-02 | 1 | `Unit` plumbing: `diagnostic` (fields, ID, `InUnit`), `protocol` (human + debug-map entry), `debugmap.Entry`, `cmd/schway-repair` cross-unit decline, field-enumeration and pinned-ID tests | 30-00 |
| 30-03 | 2 | `modload` filesystem layer and caps: Root, per-component `Lstat`, exact-case listing, FIFO, `SameFile`, bounded read, total bytes, fake-FS seam, error-to-code mapping, path-leak test | 30-01, 30-02 |
| 30-05 | 2 | `check` + `core` scopes: `FuncDecl.Module`, `ast.Unit`, import resolution diagnostics (not_exported, collisions, type item, builtin, main, imported data/foreign), scoped body-blind contracts, per-function `Unit` stamping, `Function.Module/Exported` + import facts, `EntryFunction`/`findFunction` module filter, unresolved-imports fail-closed | 30-01, 30-02 |
| 30-04 | 3 | `modload` graph and merge: iterative DFS (self/2/3/via-entry), header/path, duplicate, missing, count/depth at enqueue, diamond-once, canonical merge order, merged-program cap, parse-diagnostic `Unit` | 30-03 |
| 30-06 | 3 | Peers: corevalidate visibility law at both replay sites, clone preservation, originvalidate/pathoracle/executionpeer module-agnostic proofs, seeded fault per peer | 30-05 |
| 30-07 | 3 | `cgen` D-12: `moduleCName`, two-pass allocation, injectivity universe, reversed-order identity, diamond once in C, no-import C identical | 30-05 |
| 30-08 | 4 | `session` wiring: `CheckSet`, routes (check, run both engines, build, app verify, debug-map), `UnitTrees`, peer-diagnostic unit, single-file command refusal, source-read inventory, single-IO-site AST test, `Work` summing, emitter-inventory re-pin | 30-04, 30-06, 30-07 |
| 30-09 | 5 | Runnable gain and closure: `examples/phase30/` library + entry + README, interpreter and CLI tests on both engines, register updates (manifest, 272/273, LANGUAGE-MATURITY), `30-VALIDATION.md`, STATE/ROADMAP/LANGUAGE-MATURITY amendments | 30-07, 30-08 |

Parallelism: W1 {01, 02}; W2 {03, 05}; W3 {04, 06, 07}; files are disjoint inside each wave (syntax/ast vs diagnostic/protocol/debugmap/repair; modload vs check/core/callgraph/interp; modload vs corevalidate vs cgen). D-16 SPLIT fallback: 30.1 = plans 05-09.

## Sources

### Primary (HIGH confidence; read or executed this session)
- Repo source: `internal/compiler/{syntax/parser.go,lexer.go,format.go; ast/ast.go; check/check.go,scalar.go; core/core.go; corevalidate/corevalidate.go; callgraph/callgraph.go; interp/interp.go; cgen/cgen.go,cgen_program.go; debugmap/debugmap.go; diagnostic/diagnostic.go,diagnostic_test.go; protocol/protocol.go; evidence/evidence.go; native/native_app.go; session/*.go}`, `cmd/schway/main.go`, `cmd/schway-repair/repair.go`, `.github/workflows/ci.yml`.
- Planning: `30-CONTEXT.md`, `.planning/research/ARCHITECTURE.md` §5, `29-RESEARCH.md`, `29-VALIDATION.md`, `29-00-GATE-VERDICT.md`, `29-SECURITY.md`, `REQUIREMENTS.md`, `ROADMAP.md` Phase 30, `PROJECT.md` verification preference, `PRODUCT-ROADMAP.md`, seed `SEED-261008-z1g`.
- Executed probes (deleted afterwards): `os.Root` behavior on go1.24.0 darwin/arm64; corpus export scan (145 parseable programs, 0 undeclared `fn` exports); merged-size timing of `check.Program`/`corevalidate.Validate`.
- `go doc os.Root`, `go doc os` (no exported escape sentinel).

### Secondary (MEDIUM confidence)
- CONTEXT-recorded empirical results for APFS FIFO blocking (reproduced by the probe).

### Tertiary (LOW confidence)
- None. No web sources were used; no external library is involved.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - stdlib only, `os.Root` behavior probed locally.
- Architecture: HIGH for integration points (file:line evidence), MEDIUM for the scope-table and core-fact designs (recommendations).
- Pitfalls: HIGH - each is backed by a read line or a probe result.

**Research date:** 2026-10-12
**Valid until:** 2026-11-11 (30 days; the tree is changing phase by phase, re-grep line anchors before editing)
