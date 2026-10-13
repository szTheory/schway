# Phase 34: Agent Diagnostics Floor and Surface Manifest - Pattern Map

**Mapped:** 2026-10-13  
**Files analyzed:** 8 implementation files (tests and reconstructed fixtures follow package-local conventions)  
**Analogs found:** 8 / 8

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/compiler/diagnostic/diagnostic.go` | model / utility | transform | same file: identity constructors and `InUnit` | exact |
| `internal/compiler/protocol/protocol.go` | model / utility | transform | same file: `JSON`, `Human`, `Result.Finalize` | exact |
| `internal/compiler/session/session.go` | service | request-response | `Check`, `CheckFile`, `CheckCommandFile` | exact |
| `internal/compiler/syntax/parser.go` | parser | transform | `recoverUntil`, `recoverRegion`, `assertProgressOrBoundary` | exact |
| `internal/compiler/syntax/lexer.go` | parser / utility | transform | token diagnostics consumed by `Parse` | role-match |
| `internal/compiler/check/check.go` | service | transform | existing staged checking and refusal constructors | exact |
| `cmd/schway/main.go` | route / controller | request-response | `run`, `runCheck`, `usageResult` | exact |
| `internal/compiler/modload/load.go` | service | file-I/O | bounded loader and source inventory | exact |

All analog paths above are git-tracked. Research names new tests in those package directories; test additions should follow existing `*_test.go` conventions, but exact new test filenames remain planner discretion.

## Pattern Assignments

### `internal/compiler/diagnostic/diagnostic.go` (model / utility, transform)

**Analog:** same file, `errorID`, `errorWithRepairsID`, `Error`, `InUnit`.

**Identity pattern** (lines 130-178):

```go
func errorID(code string, span Span, causes []Cause, unit string) string {
	identity := struct {
		Schema string
		Code   string
		Span   Span
		Causes []Cause
		Unit   string `json:",omitempty"`
	}{Schema: Schema, Code: code, Span: span, Causes: causes, Unit: identityUnit(unit)}
	encoded, _ := json.Marshal(identity)
	sum := sha256.Sum256(encoded)
	return "diagnostic:" + hex.EncodeToString(sum[:12])
}
```

Keep presentation provenance and derived coordinates out of this hash shape. Nested cause spans and unit attribution currently participate; `InUnit` recomputes IDs through these constructors.

**Repair pattern** (lines 30-43, 181-193): repairs carry byte spans and units, with edit details intentionally outside identity; resolve each repair/cause location through `EffectiveUnit`.

### `internal/compiler/protocol/protocol.go` (model / utility, transform)

**Analog:** same file, `Result.Finalize`, `JSON`, and `Human`.

**JSON projection pattern** (lines 548-579, 581-594): `JSON` finalizes the result, marshals a typed result, appends exactly one newline, and retries output-byte accounting up to four times. Add diagnostic location fields to the presentation/wire projection without changing result identity inputs.

**Human projection pattern** (lines 598-644):

```go
func Human(result Result) (string, error) {
	result = result.Finalize()
	for attempts := 0; attempts < 4; attempts++ {
		output := human(result)
		if result.Metrics.OutputBytes == len(output) {
			return output, nil
		}
		result.Metrics.OutputBytes = len(output)
	}
	return "", fmt.Errorf("command output size did not converge")
}
```

Preserve this convergence/error behavior and existing stream/exit routing. Snippets belong only in human rendering; JSON should expose coordinates without source text.

### `internal/compiler/session/session.go` (service, request-response)

**Analog:** `Check`, `CheckFile`, and `CheckCommandFile` (lines 777-842, 1103 onward); `CheckResult.UnitTrees` is declared at lines 36-43.

**Parse gate pattern** (lines 777-796):

```go
parsed := syntax.Parse(source)
result := CheckResult{Tree: parsed.Tree, Diagnostics: append([]diagnostic.Diagnostic(nil), parsed.Diagnostics...), Work: len(parsed.Tree.Tokens)}
if len(result.Diagnostics) > 0 {
	return result
}
checked := check.Program(parsed.Program)
```

Carry the exact parse/check source snapshot and loader-confirmed file labels through command presentation. Retain the early parse-error gate so rejected/recovered ASTs never enter checker or native emission. Do not re-read paths to render errors.

### `internal/compiler/syntax/parser.go` (parser, transform)

**Analog:** same file, `recoverUntil`, `recoverRegion`, `assertProgressOrBoundary` (lines 1168-1214), plus `report` (lines 231-239).

**Bounded recovery pattern** (lines 1168-1208):

```go
func (p *parser) recoverUntil(boundaries ...Kind) {
	if p.atAny(boundaries...) {
		return
	}
	start := p.position
	for !p.atAny(boundaries...) {
		p.advance()
	}
	if p.position <= start {
		panic("parser recovery made no progress")
	}
}
```

Build local recognizers on caller-owned boundaries and preserve the discarded span from `recoverRegion`. Keep progress assertions, sibling/closer ownership, named overflow diagnostics and exact-token lexer coordination; do not add broad diagnostic filtering.

### `internal/compiler/syntax/lexer.go` (parser / utility, transform)

**Analog:** tokenization and unexpected-byte reporting, consumed by `syntax.Parse` (`parser.go:66`, token normalization near lines 126-147).

Keep lexer changes narrow: a parser refusal may claim only the exact offending token/span. Preserve trivia/string/comment handling and lossless `Tree.Bytes()` behavior. The parser is the closest coordination boundary; avoid global punctuation or message-prefix suppression.

### `internal/compiler/check/check.go` (service, transform)

**Analog:** existing refusal constructors and staged checker entry points; e.g. `checkCallGraphAcyclic` and `verifyCallableRefusal` (around lines 734 and 2550), plus `stampUnit` (2181).

**Refusal pattern:** construct `diagnostic.Error` at the source AST span with a stable code and concise message, and stamp module ownership where diagnostics cross module checking. Keep contextual conditions in the checker where binding/result facts exist; specifically, `true` and `false` remain ordinary identifiers, ordinary calls keep their existing arity refusal, and `byte_at` retains its own arity.

The manifest may share a small code/hint catalog, but its accepted/refused witnesses and branch census must remain independently authored from that catalog.

### `cmd/schway/main.go` (route / controller, request-response)

**Analog:** `run` (lines 35-125), `runCheck` (332-339), and `usageResult` (679-681).

**Route pattern** (lines 35-63):

```go
func run(args []string) int {
	// Route the application command before global flag extraction...
	if len(args) > 0 && args[0] == "app" {
		return runApplication(args)
	}
	args, jsonMode, ok := extractJSON(args)
	if !ok {
		return emit(usageResult(), true, false)
	}
```

Recognize `surface` before global extraction can consume unrelated flags, while explicitly preserving the supported `--json surface` placement. Keep invalid argument shapes on the existing usage-result/exit table. Add route help and classify it as `no-entry-source`; do not add it to module-aware routes.

### `internal/compiler/modload/load.go` (service, file-I/O)

**Analog:** existing bounded one-site module reads and source inventory (`load.go:92-99,130-151`; relative path records in `resolve.go:137-145`).

Use the loaded bytes and loader-confirmed relative path inventory as presentation provenance for imported-file diagnostics. This is the authority for file labels; never resolve paths from diagnostic prose or perform a second read.

## Shared Patterns

### Preserve byte authority and identity

**Sources:** `diagnostic/diagnostic.go:130-178`; `syntax/tree.go:5-17`; `protocol/protocol.go:548-594`. Keep byte offsets as the sole edit coordinates. Compute line/column and human snippets from captured source separately; omit locations when verified source provenance is absent. Nil spans remain absent, and invalid/inverted/out-of-bounds spans do not get clamped into apparent edit authority.

### Keep checking independent from claims

**Sources:** `session/session.go:777-796`; `check/check.go` refusal/check paths. The surface document describes checked current behavior, not parser acceptance. Validate each admission/refusal with independent source witnesses and a bounded census, including reverse-direction drift controls.

### Maintain deterministic bounds

**Sources:** `parser.go:121-147,1168-1214`; `protocol.go:581-644`; `modload/load.go:92-151`. Preserve source/token/diagnostic limits, bounded output and deterministic iteration/serialization. Recovery must leave a valid caller boundary untouched and demonstrate progress.

## No Analog Found

None for the listed implementation files. The typed surface-manifest schema itself is new; reuse explicit Go struct serialization and existing CLI/result conventions, while keeping admission witnesses independent.

## Metadata

**Analog search scope:** `internal/compiler/{diagnostic,protocol,session,syntax,check,modload}/`, `cmd/schway/`  
**Files scanned:** 8 tracked source analogs  
**Pattern extraction date:** 2026-10-13
