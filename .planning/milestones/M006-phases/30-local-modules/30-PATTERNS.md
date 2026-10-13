# Phase 30: Local Modules - Pattern Map

**Mapped:** 2026-10-12
**Files analyzed:** 31 new/modified (groups below)
**Analogs found:** 27 exact or role-match / 31. The 4 with no direct analog are listed under "No Analog Found".

All analog paths were checked as git-tracked source (`git ls-files` non-empty for the core ones: session_entry.go, format.go, callgraph.go, check.go, testdata/phase272). None is a gitignored mirror. Line numbers are as of HEAD `194cbf9`. RESEARCH.md says to re-grep anchors before editing.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/compiler/modload/fs.go` (Root, per-component Lstat, exact case, bounded read) | service (sole IO site) | file-I/O | `session.go:1890` `readBoundedFile` (idiom only), `session_entry.go` (single site pattern) | partial (no Root user exists) |
| `internal/compiler/modload/graph.go` (iterative DFS, caps, diamond once) | service | batch/transform | `callgraph/callgraph.go:472-560` `Order` | role-match |
| `internal/compiler/modload/merge.go` | utility | transform | `callgraph/callgraph.go` (canonical sorted order) | partial |
| `modload/*_test.go` (fake-FS seam, caps N/N+1, FIFO, symlink) | test | file-I/O | `syntax/syntax_test.go:32-41`, `callgraph.go:93-112` `...ForTest` seams | role-match |
| `internal/compiler/ast/ast.go` (+ImportDecl, Unit, FuncDecl.Module) | model | transform | existing `ast.Export` and `ast.Program` | exact |
| `internal/compiler/syntax/{token,lexer}.go` (+TokenImport) | config | transform | `token.go:13` `TokenExport`, `lexer.go:12` | exact |
| `internal/compiler/syntax/parser.go` (importDecl, caps, errors) | parser | transform | `parser.go:258-286` `modulePath`/`exports` | exact |
| `internal/compiler/syntax/format.go` (import layout) | formatter | transform | `format.go:105-126,240-256` export context | exact |
| `syntax/*_test.go` (grammar, format, fuzz, generated limit) | test | transform | `syntax_phase29_test.go`, `syntax_test.go:32-41` | exact |
| `internal/compiler/diagnostic/diagnostic.go` (Unit, InUnit, identity structs) | model | transform | `diagnostic.go:100-156` | exact |
| `diagnostic/diagnostic_test.go` (field enumeration, ID control) | test | transform | `diagnostic_test.go:175,232-250` | exact |
| `internal/compiler/protocol/protocol.go` (unit in JSON and human line) | model | request-response | `protocol.go:618-632` | exact |
| `internal/compiler/debugmap/debugmap.go` (Entry.Unit, key by module+name) | service | transform | `debugmap.go:160-190` | exact |
| `cmd/schway-repair/repair.go` (decode unit, decline cross-unit) | utility | request-response | `repair.go:50-75` jsonRepair/jsonDiagnostic | exact |
| `internal/compiler/check/check.go` (per-module scope, Unit stamping, module.* codes) | service | transform | `check.go` `calleeContracts` and `semanticID` (anchors in RESEARCH) | exact (in-place edit) |
| `internal/compiler/core/core.go` (+Function.Module/Exported, import facts) | model | transform | `core.Function` fields (`core.go:127-150`) | exact |
| `internal/compiler/corevalidate/corevalidate.go` (visibility law, clone preservation) | service | transform | sites `:2900`, `:3299`, `cloneProgram :5225` | exact (in-place) |
| `internal/compiler/callgraph/callgraph.go` (`EntryFunction` module filter) | service | transform | `callgraph.go:322-345` | exact (in-place) |
| `internal/compiler/interp/interp.go` (`findFunction`) | service | transform | `interp.go:1977-1984` | exact (in-place) |
| `internal/compiler/cgen/{cgen,cgen_program}.go` (`moduleCName`, two-pass allocate) | service | transform | `cgen.go:1560-1600`, `cgen_program.go:798` | exact |
| `internal/compiler/session/session_modules.go` (CheckSet, routes) | service | request-response | `session.go:829` `CheckFile`, `session_entry.go` `resolveEntry` | role-match |
| `session/session_phase30_*_test.go` (single-IO-site AST, diamond, unit, refusals) | test | request-response | `session_entry_inventory_test.go` | exact |
| `testdata/phase30/prechange-corpus.json` | config (golden) | batch | `testdata/phase272/accepted-corpus-hashes.json` | exact |
| `session/session_phase30_prechange_test.go` (golden + non-inert control) | test | batch | `session_phase272_identity_test.go:128-215` | exact |
| `session/session_phase30_gate_test.go` + `30-00-GATE-VERDICT.md` | test + doc | mixed | `session_phase29_gate_test.go`, `29-00-GATE-VERDICT.md` | exact |
| `examples/phase30/adler32.schway` (library module) | example | transform | `examples/phase29/adler32_fixture.schway` | exact |
| `examples/phase30/adler32_cli.schway` (entry with import) | example | request-response | `examples/phase29/adler32_fixture.schway` (foreign block, `main`) | exact |
| `examples/phase30/README.md` | doc | n/a | `examples/phase29/README.md` | exact |
| `cmd/schway/main_phase30_test.go` | test | request-response | `cmd/schway/main_phase29_test.go:1-66` | exact |
| Register edits: `session_phase28_manifest_test.go`, `session_phase272/273_identity_test.go`, `.planning/LANGUAGE-MATURITY.md`, `public-emitter-consumers.json` | config | n/a | prior-phase rows (see Shared Patterns) | exact |
| `cmd/schway/main.go` (human unit rendering, routing) | controller | request-response | `main.go:674` `problemResult` | role-match |
| `30-VALIDATION.md` | doc | n/a | `29-VALIDATION.md` | exact |

## Pattern Assignments

### `internal/compiler/syntax/parser.go` import grammar (parser, transform)

**Analog:** `internal/compiler/syntax/parser.go:258-286`

**Module path + item loop to copy** (lines 258-286):
```go
func (p *parser) modulePath() string {
	parts := []string{p.identifier("syntax.expected_module_name").Text}
	for p.accept(TokenDot) {
		parts = append(parts, p.identifier("syntax.expected_module_name").Text)
	}
	return strings.Join(parts, ".")
}

func (p *parser) exports() []ast.Export {
	p.expect(TokenExport, "syntax.expected_export")
	p.expect(TokenLBrace, "syntax.expected_lbrace")
	var exports []ast.Export
	for !p.atAny(TokenRBrace, TokenData, TokenEOF) {
		if p.peek().Kind == TokenFn && p.looksLikeFunctionDecl() {
			break
		}
		kind := p.peek()
		if kind.Kind != TokenType && kind.Kind != TokenFn {
			p.problem("syntax.expected_export_kind", kind, "expected `type` or `fn` in export block")
			p.advance()
			continue
		}
		p.advance()
		name := p.identifier("syntax.expected_export_name")
		exports = append(exports, ast.Export{Kind: kind.Text, Name: name.Text, Span: spanFrom(kind, name)})
	}
	p.expect(TokenRBrace, "syntax.expected_rbrace")
	return exports
}
```
`importDecl` = `expect(TokenImport)`, `modulePath()`, `expect(TokenLBrace)`, then the same item loop. Keep `type` items in the AST (check refuses them by name, D-02). An empty list emits `syntax.expected_import_name`.

**Limit pattern** (lines 10-17, 59-62, 70):
```go
maxDeclarations   = 4096
maxFunctions      = 1024
...
if limit := programLimit(program); limit != "" {
	p.diagnostics = []diagnostic.Diagnostic{inputLimit(len(source))}
```
Add the imports-per-module cap here, as a parse-time `module.imports_limit` that goes through one `exceeds(n, cap)` helper (armed `>` to `>=` mutant). Export `MaxFunctions` and `MaxDeclarations` for `module.program_limit` (D-21).

**Token registration** (`token.go:13`, `lexer.go:12`):
```go
TokenExport     Kind = "export"
"export":  TokenExport,
```
Add `TokenImport` and `"import"` in the same two places.

---

### `internal/compiler/syntax/format.go` (formatter, transform)

**Analog:** the same file, the export branch.

**Top-level declaration branch** (lines 105-126):
```go
case TokenModule:
	f.blankBeforeTopLevel()
	f.write("module ")
case TokenExport:
	f.blankBeforeTopLevel()
	f.write("export ")
```
Add `case TokenImport:` with `blankBeforeTopLevel()`, `write("import ")` and `f.header = "import"`.

**Brace context** (lines 240-256):
```go
case TokenLBrace:
	context := "block"
	if f.previous == TokenExport {
		context = "export"
	} else if f.header != "" {
		context = f.header
	}
```
For the item-per-line layout, make `import {` resolve to context `"export"` (or an `"import"` context handled alongside it). The `TokenFn` branch at line ~113 checks `f.context() == "export"` to call `ensureLine()`, so extend that check. Do NOT sort. Semantic tokens never move (`format.go:8`). Armed mutant: the formatter sorts imports.

---

### `internal/compiler/syntax/*_test.go` (tests)

**Analog:** `syntax/syntax_phase29_test.go:1-50` for table-driven parse tests, `syntax/syntax_test.go:32-41` for generated N/N+1.

```go
func phase29Source(body string) []byte {
	return []byte("module phase29.surface\n\nexport {\n  fn probe\n}\n\n...")
}
...
for _, test := range []struct{ name, body, code string }{ ... } {
	t.Run(test.name, func(t *testing.T) {
		parsed := syntax.Parse(phase29Source(test.body))
		if len(parsed.Diagnostics) == 0 || parsed.Diagnostics[0].Code != test.code { t.Fatalf(...) }
```
```go
func TestTokenBudgetStopsWithOneStableDiagnostic(t *testing.T) {
	result := syntax.Parse([]byte(strings.Repeat("x ", syntax.MaxTokens+1)))
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "syntax.input_limit" {
```
Use `strings.Repeat` to generate 16 vs 17 `import` blocks for `TestPhase30ImportsLimitParse` (Pitfall 9). Tests are in `package syntax_test`.

---

### `internal/compiler/diagnostic/diagnostic.go` (Unit plumbing)

**Analog:** `diagnostic.go:100-156`

**Identity struct to extend** (lines 108-114):
```go
identity := struct {
	Schema string
	Code   string
	Span   Span
	Causes []Cause
}{Schema: Schema, Code: code, Span: span, Causes: causes}
```
Add `Unit string \`json:",omitempty"\`` to this struct and to the `ErrorWithRepairs` struct (lines 143-149). Add `Unit string \`json:"unit,omitempty"\`` to `Diagnostic` (struct at lines ~100-107), `Cause` and `Repair`. `InUnit(unit)` returns the diagnostic unchanged for an empty unit and otherwise recomputes the ID. Repair kinds only go into the ID (comment at lines 128-136), so classify `Repair.Unit` as non-identity.

**Existing armed controls to update, not weaken:**
- `diagnostic_test.go:175` `TestRepairFieldEnumerationIsExhaustive` has `expected := []string{"Applicability", "Detail", "Kind", "Replacement", "Span"}`. Add `"Unit"`.
- `diagnostic_test.go:243` pins `const wantID = "diagnostic:d34396fcc8ea6530a1f04dc3"`. It must not change.

---

### `internal/compiler/protocol/protocol.go` (human projection)

**Analog:** `protocol.go:618-632`
```go
fmt.Fprintf(&output, "%s %s %s", result.ID, result.Command, result.Status)
if result.ModuleID != "" {
	fmt.Fprintf(&output, " module=%s", result.ModuleID)
}
...
for _, problem := range result.Diagnostics {
	fmt.Fprintf(&output, "%s %s [%d:%d]: %s\n", problem.ID, problem.Code, problem.Primary.Start, problem.Primary.End, problem.Message)
}
```
D-23 form: `ID CODE unit=<module> [s:e]: message`, with `unit=` only when set. Copy the `if result.ModuleID != ""` conditional shape. Armed mutant: always print.

---

### `internal/compiler/debugmap/debugmap.go`

**Analog:** `debugmap.go:165-187`
```go
astByName := make(map[string]ast.FuncDecl, len(astProgram.Funcs))
for _, function := range astProgram.Funcs {
	astByName[function.Name] = function
}
...
for _, function := range checkedProgram.Functions {
	astFunction, ok := astByName[function.Name]
```
Change the key to a `struct{module, name string}` using `ast.FuncDecl.Module` and `core.Function.Module`, and set `Entry.Unit` (omitempty). Armed mutant: key by bare name.

---

### `cmd/schway-repair/repair.go` (repair driver)

**Analog:** `repair.go:50-75`
```go
type jsonRepair struct {
	Kind          string    `json:"kind"`
	Span          *jsonSpan `json:"span,omitempty"`
	Replacement   string    `json:"replacement,omitempty"`
	Applicability string    `json:"applicability,omitempty"`
}
type jsonDiagnostic struct {
	Code    string       `json:"code"`
	Repairs []jsonRepair `json:"repairs,omitempty"`
}
```
Add `Unit string \`json:"unit,omitempty"\`` to both. `selectRepair` (`:341`) and `driverEligible` (`:272`) must treat an effective unit that is not the entry as ineligible (existing `repair.none_eligible`). Constraints: no `internal/compiler/*` import (`import_boundary_test.go`) and no prose fields (`TestRepairDriverDecodesNoProseFields`).

---

### `internal/compiler/modload/graph.go` (service, DFS)

**Analog:** `callgraph/callgraph.go:472-560` `Order` (iterative three-color DFS).

```go
colorOf := make(map[string]color, len(adjacency))
onStackIndex := make(map[string]int, len(adjacency))
...
stack := []stackFrame{{id: root}}
colorOf[root] = gray
...
switch colorOf[child] {
case white: colorOf[child] = gray; stack = append(stack, stackFrame{id: child}); onStackIndex[child] = len(stack) - 1
case gray:  recordCandidate(onStackIndex[child])   // real cycle: members = stack[startIndex:]
case black: // diamond: legitimately revisited shared node, not a cycle
```
Copy the shape, with these differences:
- Nodes are canonical module names.
- The graph is built lazily, so cap checks (count, depth) run at enqueue, before open (D-13).
- The cycle witness is the stack slice. Report it canonically rotated (lexicographically smallest rotation, the D-07-43 idea) in the `module.import_cycle` message.
- The entry node is identified with `os.SameFile` (Pitfall 8, D-24).

**Test-seam convention** (`callgraph.go:93-112`): package-level unexported `grayVsVisitedMutationForTest`, `skipSelfEdgeForTest` and similar. Mirror this for `exceedsOffByOneForTest`, a size-from-Stat mutant, a count-after-open mutant, and `diamondTwiceForTest`. Same-package tests only, restored with `defer`.

---

### `internal/compiler/modload/fs.go` (file-I/O)

**Analog (idiom only):** `session/session.go:1890-1897`
```go
func readBoundedFile(path string, limit int) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil { return nil, err }
	defer file.Close()
	return io.ReadAll(io.LimitReader(file, int64(limit)+1))
}
```
Use the `io.LimitReader(f, limit+1)` read, but NOT `os.Open` directly. Sequence per D-08..D-10, with the sketch at RESEARCH "Per-component resolution":
1. `os.OpenRoot(entryDir)`.
2. For each path component, `Root.Lstat`, refusing `ModeSymlink` (`module.symlink`).
3. Exact case via `root.Open(parent)` then `(*os.File).ReadDir(-1)`, bounded to 65,536 entries (D-24).
4. `Lstat().Mode().IsRegular()` before `Open`.
5. `Open`, then `Stat` and `os.SameFile`.
6. Bounded read. Size never comes from Stat.

Hide the filesystem behind a small interface so the fake-FS tests can lie about size, count opens, and swap file identity. No analog exists for `os.OpenRoot` (grep found none in the repo). Never format `*fs.PathError` text into diagnostics (Pitfall 6).

---

### `internal/compiler/cgen` naming (service)

**Analog:** `cgen.go:1560-1600`, `cgen_program.go:798`
```go
func (n *cNames) allocate(preferred, category string, ordinal int) string { ... base := preferred + "_SCHWAY_" + strings.ToUpper(category) + "_" + strconv.Itoa(ordinal) ... }
func cName(name string) string { ... return "SCHWAY_" + out.String() }   // invariant ^SCHWAY_[A-Z0-9_]*$
...
for index, function := range functions {
	functionNames[index] = globals.allocate(cName(function.Name), "function", index)
```
Pass 1 stays this loop for entry functions (`Module == ""`), unchanged, so no-import C is identical. Pass 2 allocates imported functions sorted by (Module, Name) with `moduleCName` (RESEARCH "D-12 name function" gives the exact body). Tests extend `cgen_names_test.go`, `TestGeneratedIdentifierNamespacesStayConfined` and `TestNativeIdentifiersRemainCollisionFree`.

---

### `callgraph.EntryFunction`, `interp.findFunction` (in-place)

**Analog:** `callgraph.go:329-336` and `interp.go:1977-1984`
```go
for _, function := range program.Functions {
	if function.Name == "main" { mains = append(mains, function.ID) }
}
```
```go
func findFunction(program core.Program, name string) (core.Function, bool) {
	for _, function := range program.Functions {
		if function.Name == name { return function, true }
```
Add `&& function.Module == ""`. Roots are likewise filtered to entry-module functions. Keep the single call site per `session_entry_inventory_test.go` (filtering lives inside `EntryFunction`).

---

### `internal/compiler/session/session_modules.go` (route wiring)

**Analog:** `session.go:829-835`
```go
func CheckFile(path string) (CheckResult, error) {
	source, err := readBoundedFile(path, syntax.MaxSourceBytes)
	if err != nil { return CheckResult{}, err }
	return Check(source), nil
}
```
`CheckSet`/the loader-backed file routes are siblings of this, in a new file (keeps `session.go` line numbers stable for the emitter-inventory rows). Downstream flow is unchanged: `check.Program`, `admitProgram`, then `resolveEntry` (`session_entry.go:~30`, which has the "Phase 30 composition point" comment to update). Keep the byte-taking `Check` failing closed on `Imports` (`module.imports_not_loaded`).

---

### `session/session_phase30_*_test.go` (single-IO-site and inventory)

**Analog:** `session_entry_inventory_test.go:1-140`

AST scan with alias resolution, fail-closed on dot imports, a walk of `internal` and `cmd` that skips `_test.go`, a sorted `want` slice, and a `negative-control` subtest:
```go
calls = append(calls, entryFunctionCallsIn(t, filepath.ToSlash(relative), path, nil)...)
...
want := []string{"internal/compiler/cgen/cgen_program.go", "internal/compiler/session/session_entry.go"}
t.Run("negative-control", func(t *testing.T) {
	synthetic := "package x\nimport cg \"" + callgraphImportPath + "\"\nfunc f() { _, _ = cg.EntryFunction(nil) }\n"
```
Copy this for `TestPhase30SingleIOSite` (opens of module bytes outside `modload`, `OpenRoot` and `readBoundedFile` call sites) and the scoped-contract builder referencing no `.Body`/`.Linear`/`.Match` selectors.

---

### `testdata/phase30/prechange-corpus.json` + golden test

**Analog:** `session_phase272_identity_test.go:30-215`, golden `testdata/phase272/accepted-corpus-hashes.json`.

**Golden types and path** (lines 62-86):
```go
type phase272CorpusGolden struct { Schema string `json:"schema"`; Revision string `json:"revision"`; Entries []phase272CorpusEntry `json:"entries"` }
const phase272GoldenSchema = "phase272.accepted-corpus-hashes/1"
func phase272GoldenPath() string { return testsupport.ProjectPath("testdata", "phase272", "accepted-corpus-hashes.json") }
```
**Walk** (lines 88-126): `filepath.Walk` over `examples` and `testdata`, `.schway` only, skip `> syntax.MaxSourceBytes`, SHA-256 of `json.Marshal(...)`, entries sorted by path.

**Write mode** (lines 130-157):
```go
if os.Getenv("SCHWAY_PHASE272_WRITE_GOLDEN") == "1" {
	revision := strings.TrimSpace(os.Getenv("SCHWAY_PHASE272_GOLDEN_REVISION"))
	if revision == "" { t.Fatal("... must name the commit the golden is generated at") }
	...json.MarshalIndent(golden, "", "  ") ... os.WriteFile(goldenPath, append(encoded, '\n'), 0o644)
```
**Compare + allow-list** (lines 158-214): `golden.Schema` check, `len(golden.Entries) < MinCorpus` guard, per-path hash compare, plus "newly accepted absent from golden and not in allow-list" error.

Phase 30 changes:
- Env names are `SCHWAY_PHASE30_WRITE_GOLDEN` and `SCHWAY_PHASE30_GOLDEN_REVISION`. Schema is `phase30.prechange-corpus/1`.
- Each entry records per D-15: `CheckCommandFile(path)` JSON SHA, core JSON SHA, `EmitApplication` SHA or error, `Phase16ControlNativeC` SHA or error, `DebugMapCommandFile(path, "")` JSON SHA, `syntax.Format` SHA.
- It needs an exemption for `examples/phase30/` paths (same pattern as the `phase272NewlyAcceptedPaths` allow-list).
- Add `TestPhase30PreChangeGoldenIsNotInert`, which flips one recorded hash in a copy and expects failure.
- Do NOT call emitters other than `cgen.EmitApplication` and `Phase16ControlNativeC` from new test files without re-pinning the emitter inventory (the file header comment at `session_phase272_identity_test.go:20-22` says so).

---

### `session_phase30_gate_test.go` + `30-00-GATE-VERDICT.md`

**Analog:** `session_phase29_gate_test.go:1-110` (imports of `session`, `corevalidate`, `interp`, `testsupport`; inline-string sources so no `.schway` file joins the corpus; table-driven `{name, source, code}`), and `29-00-GATE-VERDICT.md`.

**Verdict doc shape:**
- Heading `# Phase 29 Wave-0 Gate Verdict`.
- A `## Flags` table with columns `Flag | Pass | Evidence` (test names as evidence).
- Free-text sections.
- `## Decision` ending with a bare line `verdict: PROCEED`.

Phase 30 uses flags (a)-(f) from D-14. Gate test names are `TestPhase30Gate*`, each with its own armed control. Do not put a verdict token beside a `go test` command span in proposal-role documents (groundedness lint).

---

### `examples/phase30/adler32.schway`, `adler32_cli.schway`, `README.md`

**Analog:** `examples/phase29/adler32_fixture.schway` (git-tracked, together with `README.md`, `adapter.c/h`, `fixture_bytes.bindings.json`).

Library (`adler32.schway`) copies the `fn adler32(bytes: ByteView) -> U64 { ... }` body. Its header is `module adler32` (path-derived, D-06) and it declares `export { fn adler32 }`. It has no `data` or `foreign` (D-20).

Entry (`adler32_cli.schway`) copies the header/export/`foreign C` block/`data FixtureError`/`main` from the phase 29 file. It adds the D-01 import block between `module` and `export`:
```
module examples.phase30.adler32_cli

import adler32 {
  fn adler32
}

export {
  fn main
}
```
The entry keeps the `foreign` block because the foreign owner stays in the entry (D-20). Reuse `examples/phase29/fixture_bytes.bindings.json` and the adapter files via the manifest path. Do not copy them, and leave the phase 29 files untouched. The library must exist as exactly one source file (`TestPhase30LibraryIsSingleCopy`).

---

### `cmd/schway/main_phase30_test.go`

**Analog:** `cmd/schway/main_phase29_test.go:20-66`
```go
if _, err := exec.LookPath("clang"); err != nil { t.Skip("env:clang toolchain unavailable on this host") }
source := testsupport.ProjectPath("examples", "phase29", "adler32_fixture.schway")
manifest := testsupport.ProjectPath("examples", "phase29", "fixture_bytes.bindings.json")
artifact := filepath.Join(t.TempDir(), "adler32")
code, stdout, stderr := captureLangRun(t, []string{"build", source, "--manifest", manifest, "--output", artifact})
if code != 0 || stdout != "built "+artifact+"\n" || stderr != "" { ... }
for caseID := uint64(0); caseID <= 9; caseID++ {
	want, ok := testsupport.Phase29FixtureExpected(caseID)
	...captureLangRun(t, []string{"app", "run", artifact, "--", input})
```
Point `source` at `examples/phase30/adler32_cli.schway`. Keep the 10 pinned values via `testsupport.Phase29FixtureExpected`. Note the existing `run-<engine>-unchanged` subtests assert exit 3 for a foreign-owner program under `schway run`; re-derive the expectation for the phase 30 entry rather than copying blindly. The interpreter half lives in session (`CheckFile` + `interp.RunWithFixtureBytes`, see `session_phase29_interp_fixture_test.go`).

---

## Shared Patterns

### Armed seam convention (every cap, guard and mutant)
**Source:** `callgraph/callgraph.go:93-112` (`grayVsVisitedMutationForTest`, `skipSelfEdgeForTest`), `check_test.go:3918-3945` (`verifyCallableRefusalBodyReadSeam`, `calleeBodyReadObserved`)
**Apply to:** modload, check, corevalidate, diagnostic, repair tests.
```go
defer func() { verifyCallableRefusalBodyReadSeam = false; calleeBodyReadObserved = nil }()
var bodyReads int
calleeBodyReadObserved = func(string) { bodyReads++ }
result := Program(mustParseProgram(t, source))
```
Production default is zero observed reads. The control engages the seam and the observer must fire. Phase 30's `TestPhase30CallAdmissionBodyBlind` extends this on a two-module fixture.

### Generated N / N+1 fixtures
**Source:** `syntax/syntax_test.go:32-41` (`strings.Repeat`, one stable `syntax.input_limit`). **Apply to:** all D-13 cap tests. Build under `t.TempDir()`. Nothing large is committed (D-27).

### Env-gated golden writer
**Source:** `session_phase272_identity_test.go:130-157`. **Apply to:** the pre-change golden only.

### Byte-identity guards that must stay green unchanged
`TestPhase272AcceptedCorpusByteIdentity` (`:128`), `TestPhase273ConformanceCorpusByteIdentity` (`session_phase273_identity_test.go:146`), `TestDiagnosticZeroBytesUnchanged`, `TestPhase1DiagnosticGoldenUnchanged`, `TestPhase28EntryExitClassManifest`.

### Register updates when a `.schway` file is committed
- `session_phase272_identity_test.go:33-48` `phase272NewlyAcceptedPaths`: add `"examples/phase30/adler32.schway": true,` with a dated comment (the library alone checks clean). The entry with imports is refused by the byte-taking `Check` and never joins the walk.
- `session_phase273_identity_test.go`: the equivalent `phase273NewlyAcceptedPaths` row.
- `session_phase28_manifest_test.go:232` `phase28NewlyAddedEntries`: add rows for the two new files. `phase28MinCorpus = 155` (`:~29`) is a floor, so bump only if the plan chooses to.
- `.planning/LANGUAGE-MATURITY.md:302` count sentence (`TestLanguageMaturityCountsAreCurrent`).
- `testdata/phase16/public-emitter-consumers.json`: re-pin once at the end of 30-08 if any `session.go` line above 3049 moved.

### Error and diagnostic naming
Codes use `<family>.<snake_case>`. All new refusals go through `diagnostic.Error(code, span, message, causes...)` (`diagnostic.go:107`) or `ErrorWithRepairs` when a repair exists. Only two refusals carry repairs (D-04): `module.duplicate_import` (replacement `"\n"`, because `DriverEligible` needs a non-empty replacement) and `syntax.import_after_export`.

### Path hygiene
Diagnostics and records carry root-relative module names only. Never put `err.Error()` of a `*fs.PathError` or an absolute path in a diagnostic (`TestPhase30ErrorsNeverLeakAbsolutePaths`).

## No Analog Found

| File | Role | Data Flow | Reason |
|---|---|---|---|
| `modload/fs.go` Root / per-component Lstat / exact-case listing | service | file-I/O | No `os.OpenRoot`, `Mkfifo` or `os.Symlink` use exists in the tree (grep returned nothing). Use RESEARCH "Per-component resolution" and the Pitfall 6 probe table. Only the `LimitReader` read idiom has an analog. |
| FIFO / symlink / case-variant tests (`//go:build unix`) | test | file-I/O | No prior filesystem-adversarial tests. Use `syscall.Mkfifo`, `os.Symlink`, a 2 s deadline, and a creation probe with `t.Skip` for the Linux-only both-spellings case. |
| Fake-FS interface (Stat lies about size, open counter, identity swap) | test seam | file-I/O | No existing FS abstraction. Follow only the `...ForTest` convention above. |
| `corevalidate` visibility law (`core.callee_not_exported`, `core.callee_not_imported`) | service | transform | No module concept exists. Add at the two existing `declaredFunctionIDs[operation.CalleeID]` replay sites (`corevalidate.go:2900`, `:3299`), per the RESEARCH snippet. Fields must be JSON-tagged because `cloneProgram` (`:5225-5235`) round-trips through JSON. |

## Metadata

**Analog search scope:** `internal/compiler/{syntax,diagnostic,protocol,debugmap,callgraph,cgen,interp,session,testsupport}`, `cmd/schway`, `cmd/schway-repair`, `examples/phase29`, `testdata/phase272`, `.planning/phases/29-*`.
**Files scanned:** about 25 read (targeted ranges), plus greps across `internal` and `cmd`.
**Pattern extraction date:** 2026-10-12
