# Phase 14: Evidence Instrument and Honest Scoping - Pattern Map

**Mapped:** 2026-09-17
**Files analyzed:** 17 (new + modified, per CONTEXT.md `<code_context>`/`<canonical_refs>` and RESEARCH.md's "Recommended Project Structure")
**Analogs found:** 17 / 17 — this is an instrument phase; the repo already ships the exact shape every new guard extends. There is no "no analog" bucket.

**Repo-wide guard-test convention (read this before any per-file section):**
- **Placement:** guards live as `_test.go` files colocated with the package they examine (`internal/compiler/session`, `internal/compiler/check`, `cmd/lang-repair`) — never a separate `tests/` tree.
- **Naming:** `TestXxx` in plain English-sentence form describing the invariant, e.g. `TestDebtRegistersAreWellFormed`, `TestNoNAT03RowRemainsPending`, `TestInjectorMarkerCountGuardIsNotInert`. Every guard that could silently do nothing pairs with a `TestXxxIsNotInert` / `TestXxxGuardIsNotInert` sibling in the same file.
- **CI invocation:** no separate CI config — `go test ./...` is the full-suite gate (confirmed green pre-Phase-14); per-phase VALIDATION.md rows name the exact `go test <pkg> -run <Name> -count=1` invocation. There is no `--self-test` CLI convention except `scripts/assert-go-tests.sh --self-test`, which is a shell (not Go) sentinel.
- **Shared helpers:** `testsupport.ProjectPath(...)` for repo-root-relative paths; `phaseArtifactGlob(parts ...string)` for globbing `.planning/phases/**` + `.planning/milestones/*-phases/**` together — every new `.planning/**`-scanning guard must call this, never `filepath.Glob` directly.
- **Closed-vocabulary tests:** a `map[string]bool` (or map to a reason string) declared as a package-level var beside the test that enforces it, e.g. `debtRegisterSeverities`, `debtRegisterLandingPhaseExemptions`.
- **Non-inertness proof shape:** build/copy a fixture, seed exactly one fault, assert the guard goes red; iterate one seeded fault per mechanizable kind if the guard mechanizes more than one thing (see `TestInjectorMarkerCountGuardIsNotInert` and D-14-28's three-fault design for `TestSuppressionWitnessGuardIsNotInert`).

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/compiler/session/verification_groundedness_test.go` (NEW) | test / lint-guard | batch (scan `.planning/**`, classify, pin) | `internal/compiler/session/session_test.go:2613-2715` (`TestDebtRegistersAreWellFormed` + `checkDebtRegister`) | exact |
| `internal/compiler/session/evidence_grade_test.go` (NEW) | test / derivation-cap | transform (row → ceiling) | `internal/compiler/session/session_test.go:2613-2715` (same table-parser shape) + `scripts/assert-go-tests.sh:4-42` (`selection()` — the `-list`-then-exact-match ladder D-14-03 reuses) | exact (two-analog composite) |
| `internal/compiler/session/session_test.go` (EXTENDED — `checkDebtRegister` grows `Grade`/`Witness`/closed `Landing phase`) | test / registry-guard | CRUD (well-formedness over a markdown table) | itself, current shape (`:2628-2715`) | exact (in-place extension) |
| `internal/compiler/session/session_phase5.go` (EDITED — delete `Phase5AssertMutationMovesAnAxis`) | service (mutation dispatcher) | event-driven (dispatch on `ControlID`) | `internal/compiler/session/session_phase5_alias.go:497-525` (`AssertMutationMovesAnAxis`, the surviving law) | exact — sibling implementation being collapsed into |
| `internal/compiler/session/session_phase5_alias.go` (EDITED — remove `Subjected: true` exclusion path is N/A here; remove stale `PENDING-05-08` marker/comment at `:523`) | service (mutation dispatcher) | event-driven | itself, current shape | exact (in-place edit) |
| `internal/compiler/session/session_phase5_alias_test.go` (EDITED — remove `t.Skip` at `:241`) | test | event-driven | itself, current shape (`:176-265`) | exact (in-place edit) |
| `internal/compiler/session/session_phase5_test.go` (EDITED — delete `TestNoNAT03RowRemainsPending` `:229-246`, replaced by the new suppression-witness guard) | test / suppression-guard | batch (AST scan across whole module) | `internal/compiler/session/session_phase6_injectors_test.go:823-863` (`TestInjectorMarkerCountGuardIsNotInert` — the temp-copy-and-seed-one-fault shape) | role-match (the *retired* guard is single-file `os.ReadFile`+line-match; the *replacement* must instead follow the injector non-inertness pattern) |
| `internal/compiler/session/qlt02_budget_manifest.json` (EDITED — new `observed` `elapsed_ns` wall-clock row) | config / data | batch (static JSON array) | itself, current shape (5 existing rows) | exact |
| `internal/compiler/syntax/parser.go` (EDITED — `default:` arm at `:169-171`, `recoverRegion` beside `recoverUntil` at `:699`, `p.problem` widened to variadic) | parser / core-compiler | transform (token stream → AST + diagnostics) | itself, current shape (`parseProgram`, `p.problem` at `:659-665`, `recoverUntil` at `:697-704`) | exact (in-place edit, ~14 lines, one call site) |
| `internal/compiler/check/diagnostic_distinctness_test.go` (NEW) | test / corpus-gate | batch (lex fixtures → kind-sequence → distinctness) | `internal/compiler/session/session_phase6_injectors_test.go:823-863` (non-inertness shape) + a lexer-only lint (no direct precedent; closest structural analog is any `_test.go` asserting over `syntax.Lex` output) | role-match |
| `cmd/lang-repair/repair.go` (EDITED — `Outcome` gains 3 `omitempty` fields, `selectRepair` returns a decision) | service / protocol boundary | request-response (subprocess JSON in/out) | itself, current shape (`Outcome` at `:112-116`, `selectRepair` at `:200-209`) | exact (in-place edit) |
| `cmd/lang-repair/repair_diagnosis_test.go` (NEW) | test / protocol-guard | batch (walk `testdata/*_capture.json`) | `cmd/lang-repair/antitheater_test.go:625-641` (`TestProseScrambleLeavesRepairBehaviourIdentical`, the stand-in-binary pattern) + `cmd/lang-repair/repair_test.go` (`TestUnrepairableDefectFailsTheGate`) | exact (two-analog composite) |
| `testdata/distinctness/*.lang` + `collision_control.json` (NEW corpus) | test fixture / frozen golden | batch (static, never regenerated) | `cmd/lang-repair/testdata/*_capture.json` (frozen capture corpus, same "never regenerate, only assert a property" discipline) | role-match |
| `.planning/UNREACHABLE-CLAIMS.md`, `.planning/EVIDENCE-RECONCILIATION.md` (NEW, GENERATED) | generated view / registry | batch (regenerate in memory, byte-compare) | `internal/compiler/session/qlt02_budget_manifest.json` + `TestQLT02BudgetManifestFileUnchangedDuringAudit` (`session_phase6_budget_test.go:555-563`) — the existing "file must not be hand-written/silently drift" discipline, closest byte-compare precedent in tree | role-match |
| `.planning/**/*-VALIDATION.md` (13 archived, EDITED — `Grade`/`Non-inertness` columns added, `Status` retired) | data / registry document | CRUD (row edits, append-only for repair) | `09-VALIDATION.md`'s own canonical 10-column header (`:56` table header + rows `:60-100`) | exact — this is the document whose shape is being migrated |
| `.planning/**/*-DEBT.md` (12 files, EDITED — `Grade`/`Witness` columns, closed `Landing phase`) | data / registry document | CRUD | `internal/compiler/session/session_test.go`'s `checkDebtRegister` + any existing `*-DEBT.md` (e.g. `PHASE-13-DEBT.md`, cited in CONTEXT.md `D-13-02b`/`D-13-10a`/`D-13-34` rows) | exact |
| `.planning/LANGUAGE-MATURITY.md` (EDITED — count corrected 32→26) | data / self-describing document | transform (doc's own embedded grep is its own re-verify) | itself, current shape (`:96-113`, its own `Re-verify:` line at `:105`) | exact |

## Pattern Assignments

### `internal/compiler/session/verification_groundedness_test.go` (test, batch scan)

**Analog:** `internal/compiler/session/session_test.go:2504-2626` (`phaseArtifactGlob`, `TestDebtRegistersAreWellFormed`)

**Imports pattern** — this package's existing test files import only stdlib (`os`, `strings`, `path/filepath`, `testing`, `fmt`) plus the local `testsupport` package; D-14-10's resolver additionally needs `go/parser`, `go/build`, `go/ast`, `go/token` (all stdlib, zero new deps).

**Path-resolution pattern** (`session_test.go:2504-2520`):
```go
func phaseArtifactGlob(parts ...string) ([]string, error) {
	live, err := filepath.Glob(testsupport.ProjectPath(append([]string{".planning", "phases"}, parts...)...))
	if err != nil {
		return nil, err
	}
	archived, err := filepath.Glob(testsupport.ProjectPath(append([]string{".planning", "milestones", "*-phases"}, parts...)...))
	if err != nil {
		return nil, err
	}
	return append(live, archived...), nil
}
```
Reuse this verbatim — do not write a new glob helper. Call as `phaseArtifactGlob("*", "*-VALIDATION.md")` for Tier-A file discovery (D-14-06), exactly as `TestDebtRegistersAreWellFormed` calls `phaseArtifactGlob("*", "*-DEBT.md")`.

**Core scan-and-subtest pattern** (`session_test.go:2613-2626`):
```go
func TestDebtRegistersAreWellFormed(t *testing.T) {
	registers, err := phaseArtifactGlob("*", "*-DEBT.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(registers) == 0 {
		t.Fatal("no *-DEBT.md register found under .planning/phases or .planning/milestones/*-phases")
	}
	for _, path := range registers {
		t.Run(filepath.Base(path), func(t *testing.T) {
			checkDebtRegister(t, path)
		})
	}
}
```
`TestVerificationGroundednessFrontierIsPinned`, `TestValidationRowGradesAreEarned`, and the corpus-floor assertions (D-14-20) all copy this "glob, assert non-empty (the floor), `t.Run` per file" shape.

**Table/line-scan pattern** (`session_test.go:2717-2764`, `debtRegisterTable`): per-line `strings.HasPrefix(line, "|")`, `strings.Split(strings.Trim(...,"|"), "|")`, header row becomes a `map[string]int`. D-14-15's line-scanner extends this exact loop with the `\|`-unescape step (`\|`→`|` then `\\`→`\`) and per-cell backtick-span extraction — do not write a new line-classifier from scratch.

**Error handling pattern:** every violation is `t.Fatalf`, not a returned error swallowed elsewhere — matches this repo's "no silent skip" convention. D-14-15's "unparseable verification command is a violation, not a pass" maps directly onto this `t.Fatalf`-on-anything-unexpected posture; there is no `continue`-past-a-bad-row anywhere in `checkDebtRegister`.

**Selection-resolution algorithm to reuse (D-14-03/D-14-10), `scripts/assert-go-tests.sh:4-42`:**
```sh
selection() {
	package=$1; shift
	listed=$(go test "$package" -list .)
	pattern=
	for requested do
		found=false
		while IFS= read -r discovered; do
			if [ "$discovered" = "$requested" ]; then found=true; break; fi
		done <<EOF
$listed
EOF
		if [ "$found" != true ]; then
			echo "assert-go-tests: target not discovered: $requested" >&2
			return 1
		fi
		...
	done
	go test "$package" -run "^($pattern)$" -count=1
}
```
And the `--self-test` sentinel (`:44-62`) that asserts a nonexistent test name (`TestCodenameLangSelectionGuardMustNotExist`) is correctly rejected — the direct model for `TestStaticTestIndexMatchesGoTestList`'s own accuracy control and for D-14-08's "sentinel proves selection, not document scanning" framing. Port the exact-match discipline (never substring/regex-loose matching a discovered name) into the Go static resolver; do not port the shell script itself — D-14-09 keeps script and lint as two non-overlapping artifacts.

---

### `internal/compiler/session/evidence_grade_test.go` (test, derivation ladder)

**Analog 1 — vocabulary + cap shape:** `debtRegisterSeverities` (`session_test.go:2592-2594`):
```go
var debtRegisterSeverities = map[string]bool{"blocker": true, "warning": true, "info": true}
```
This is the literal template for the closed grade vocabulary (`DEFINED`, `WIRED`, `REACHABLE`, `EXERCISED`, `MUTATION-KILLED`) and for `validationGradeBarExemptions` — copy `debtRegisterLandingPhaseExemptions`'s shape exactly:
```go
var debtRegisterLandingPhaseExemptions = map[string]string{
	"02-DEBT.md": "written 2026-09-04, before the landing-phase column existed; frozen prior art",
	"03-DEBT.md": "written 2026-09-04, before the landing-phase column existed; frozen prior art",
}
```
(file-scoped, dated reason string, package-level var beside the enforcing test) — this is the exact shape D-14-06's `validationGradeBarExemptions` must copy, keyed by filename not row.

**Analog 2 — the ladder's resolution primitive:** reuse `scripts/assert-go-tests.sh:8-45`'s exact-name-against-`-list`-output algorithm (see excerpt above), reimplemented as a Go static resolver per D-14-10 (walk `*_test.go` with `go/parser.ParseFile(..., parser.SkipObjectResolution)`, collect top-level `func (Test|Fuzz|Benchmark|Example)[A-Za-z0-9_]*`), with `TestStaticTestIndexMatchesGoTestList` as the one accuracy control that shells out to real `go test ./... -list .` and asserts set equality.

**Error handling pattern:** every derivation failure is a hard `t.Fatalf` naming the row and the exact reason (unresolved name, no run-record pass, etc.) — never a warning or skip, matching D-14-15's "no silent skips" rule extended to the grade cap.

---

### `internal/compiler/session/session_test.go` (EXTENDED — `checkDebtRegister`)

**Analog:** itself (`:2628-2715`).

**Core CRUD/well-formedness pattern to extend** (do not rewrite):
```go
required := []string{"ID", "Source", "Threat/Req", "Severity", "Item"}
if _, exempt := debtRegisterLandingPhaseExemptions[name]; !exempt {
	required = append(required, "Landing phase")
}
for _, column := range required {
	if _, present := columns[column]; !present {
		t.Fatalf("%s: Items table has no %q column (columns: %v)", name, column, columns)
	}
}
...
if severity := row[columns["Severity"]]; !debtRegisterSeverities[severity] {
	t.Fatalf("%s: row %s has severity %q outside the closed vocabulary (blocker, warning, info)", name, identifier, severity)
}
```
D-14-21's `Grade`/`Witness` columns and D-14-24's closed `Landing phase` form (`P<NN>` | `CLOSED(<sha>)` | `UNOWNED(<witness-id>)`) are additional required-columns + additional closed-vocabulary-membership checks appended to this exact loop — same file, same function, no new parser.

**Frontmatter count cross-check to reuse** (`:2644-2661`):
```go
if _, scanErr := fmt.Sscanf(strings.TrimSpace(line), "items: %d", &declared); scanErr != nil { ... }
...
if len(rows) != declared {
	t.Fatalf("%s: frontmatter declares items: %d but the Items table holds %d rows", name, declared, len(rows))
}
```
D-14-05's `graded_rows: N` frontmatter field on `*-VALIDATION.md` gets the identical `fmt.Sscanf`-based cross-check.

---

### `internal/compiler/session/session_phase5.go` / `session_phase5_alias.go` (EDITED — EVD-05 collapse)

**Analog:** the two coexisting implementations themselves.

**The wrapper being deleted** (`session_phase5.go:412-429`):
```go
func Phase5AssertMutationMovesAnAxis(ctx context.Context, mutation NAT03Mutation) error {
	if mutation.ControlID == ControlSanitizeAllocatorMismatch {
		return assertAllocatorMismatchMovesAxis(ctx, mutation)
	}
	return AssertMutationMovesAnAxis(ctx, mutation)
}
```

**The surviving law it delegates to, and its shape to extend** (`session_phase5_alias.go:497-525`):
```go
func AssertMutationMovesAnAxis(ctx context.Context, mutation NAT03Mutation) error {
	switch mutation.ControlID {
	case "control:foreign.layout_mismatch":
		return assertLayoutMismatchMovesAxis(ctx, mutation)
	case "control:resource.release_omitted":
		return assertReleaseOmissionMovesAxis(ctx, mutation)
	case "control:resource.release_order_transposed":
		return assertReleaseTranspositionMovesAxis(mutation)
	case "control:foreign.nonlocal_exit_undetected":
		return assertNonlocalExitMovesAxis(ctx, mutation)
	case ControlAliasFalseNoAlias:
		return assertAliasFalseNoAliasMovesAxis(ctx, mutation)
	default:
		return fmt.Errorf("AssertMutationMovesAnAxis: unsupported control %q (row not yet subjected — see PENDING-05-08)", mutation.ControlID)
	}
}
```
D-14-27 deletes `Phase5AssertMutationMovesAnAxis`, folds `ControlSanitizeAllocatorMismatch`'s `case` directly into this `switch`, and removes the `PENDING-05-08` text from the `default:` error string (this is also D-14-26's stale-marker-in-an-error-string target).

**The `NAT03Mutation` struct and the escape declaration it must carry** (`session_phase5_alias.go:431-436`, `:487-493`):
```go
type NAT03Mutation struct {
	ControlID     string
	CorpusProgram string
	ExpectedAxis  string
	Subjected     bool
	EscapeID      string
}
...
{
	ControlID:     "control:native.sanitize.retained_pointer",
	CorpusProgram: "testdata/phase5/retained_pointer.lang",
	ExpectedAxis:  AxisTerminalOutcome,
	Subjected:     false,
	EscapeID:      "escape:callback-invocation-unsubjected",
},
```
This row already carries the exact `escape:` shape D-14-23's witness grammar formalizes — D-14-27 keeps it `Subjected: false` with this `EscapeID`, which must now additionally resolve in the closed escape registry and carry its own `probe:`.

---

### `internal/compiler/session/session_phase5_test.go` (EDITED — replace `TestNoNAT03RowRemainsPending`)

**Analog (the retired guard, showing exactly what NOT to repeat):** `session_phase5_test.go:236-246`:
```go
func TestNoNAT03RowRemainsPending(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("internal", "compiler", "session", "session_phase5_alias.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(source), "\n") {
		if strings.TrimSpace(line) == "// PENDING-05-08" {
			t.Fatalf("session_phase5_alias.go still carries a PENDING-05-08 marker comment: %q", line)
		}
	}
}
```
This is D-14-22's instance-4 defect in the flesh: it reads exactly one file for exactly one trimmed-line match, and misses the identical string surviving as an error message (`session_phase5_alias.go:523`) and as prose (`session_phase5_test.go:207`, `session_phase5_alias_test.go:231`). **Delete this function entirely** (D-14-26) — do not extend it.

**Analog for its replacement's non-inertness shape:** `internal/compiler/session/session_phase6_injectors_test.go:823-840` (`TestInjectorMarkerCountGuardIsNotInert`):
```go
func TestInjectorMarkerCountGuardIsNotInert(t *testing.T) {
	source := bytes.ReplaceAll(phase6Fixture(t, "heldout_match_defect.lang"), []byte(matchTargetMarker), []byte(""))
	unguarded := matchInjectSkippingGuard(source)
	if !bytes.Equal(unguarded, source) {
		t.Fatalf("matchInjectSkippingGuard mutated a marker-absent source; it should have returned it unchanged")
	}
	uncheckedResult := Check(unguarded)
	if len(uncheckedResult.Diagnostics) != 0 {
		t.Fatalf("guard-disabled path's unmutated source did not check clean: %+v", uncheckedResult.Diagnostics)
	}
	_, err := MatchInjector{}.Inject(source)
	typed := injectorError(err)
	if typed == nil || typed.Code != InjectorTargetMissingCode {
		t.Fatalf("guarded MatchInjector.Inject did not refuse on the same input the guard-disabled path silently accepted: %v", err)
	}
}
```
`TestSuppressionWitnessGuardIsNotInert` (D-14-28) copies this "build/copy a fixture, remove/mutate exactly one thing, assert the guard now refuses where it previously accepted" shape three times (one per seeded-fault kind: uncited `t.Skip`, mismatched `callsite:` count, neutralized probe refusal / XPASS).

**Package-wide AST-scan shape to build (no direct in-repo precedent for scanning `_test.go` + `BasicLit` strings across every package — this is genuinely new mechanism):** use `go/parser` in `parser.ParseComments` mode (same stdlib the D-14-10 static resolver already uses), walking every `*ast.File` including `_test.go`, enumerating `*ast.CallExpr` (`Skip`/`Skipf`/`SkipNow` selectors), `//go:build` constraints, `*ast.Comment`, and `*ast.BasicLit` of kind `STRING` — reuse the same `go/ast` + `go/build.Context.MatchFile` plumbing as the static test index, do not write a second AST-walking helper.

---

### `internal/compiler/session/qlt02_budget_manifest.json` (EDITED — EVD-08's `observed` row)

**Analog:** itself, current shape (5 rows shown above) — a flat JSON array of objects, each with `machine_id`, `metric`, `gate_type` (`"hard"` | `"observed"`), `value_or_bound`, `unit`, `ratified_at`, `ratified_by_commit`. The `elapsed_ns` row already present (`gate_type: "observed"`, `value_or_bound: 5000000000`) is the exact template for the new wall-clock row — same shape, new `metric` name and a freshly measured `value_or_bound` (192.7s baseline, re-measured per RESEARCH.md A2, not copied from prose).

**Guard to extend, not replace** (`internal/compiler/session/session_phase6_budget_test.go:24-43`):
```go
func TestBudgetManifestLoads(t *testing.T) {
	rows, err := LoadQLT02BudgetManifest()
	...
	for _, row := range rows {
		if row.GateType != QLT02GateTypeHard && row.GateType != QLT02GateTypeObserved {
			t.Errorf("row %+v has out-of-vocabulary gate_type %q", row, row.GateType)
		}
		if row.RatifiedAt == "" || row.RatifiedAt == "TODO" { ... }
		if row.RatifiedByCommit == "" || row.RatifiedByCommit == "TODO" { ... }
	}
}
```
And the file-integrity guard (`session_phase6_budget_test.go:555-563`, `TestQLT02BudgetManifestFileUnchangedDuringAudit`) — the closest existing "assert a checked-in JSON file is not silently mutated during a test run" precedent in the tree, directly relevant to keeping the two D-14-13 generated `.planning/*.md` views honest.

---

### `internal/compiler/syntax/parser.go` (EDITED — DX-08's ~14-line fix)

**Analog:** itself, current shape.

**The exact call site to change** (`parser.go:161-173`, `parseProgram`):
```go
for p.peek().Kind != TokenEOF {
	switch p.peek().Kind {
	case TokenData:
		program.Data = append(program.Data, p.dataDecl())
	case TokenFn:
		program.Funcs = append(program.Funcs, p.funcDecl())
	case TokenForeign:
		program.Foreign = append(program.Foreign, p.foreignBlock())
	default:
		p.problem("syntax.expected_declaration", p.peek(), "expected `data` or `fn` declaration")
		p.recoverUntil(TokenData, TokenFn, TokenForeign, TokenEOF)
	}
}
```
Replace the `default:` arm to capture `unexpected := p.peek()`, call a new `recoverRegion(...)` (returning a `diagnostic.Span`), and pass `diagnostic.Cause{Kind: "skipped_region", Span: &skipped}` into a widened `p.problem`.

**`p.problem` to widen from fixed 3-arg to variadic causes** (`parser.go:659-665`):
```go
func (p *parser) problem(code string, token Token, message string) {
	if len(p.diagnostics) >= maxPrimaryDiagnostics {
		p.truncated = true
		return
	}
	p.diagnostics = append(p.diagnostics, diagnostic.Error(code, token.Span, message))
}
```
Widen to `func (p *parser) problem(code string, token Token, message string, causes ...diagnostic.Cause)` and thread `causes...` into `diagnostic.Error(code, token.Span, message, causes...)`. **This is the only signature change; all 40+ existing call sites compile unchanged** (variadic trailing param).

**`recoverUntil`, the sibling to place `recoverRegion` beside** (`parser.go:697-704+`):
```go
// recoverUntil is the only token-skipping recovery primitive. If it is not
// already at a caller-owned boundary, it must consume at least one token.
func (p *parser) recoverUntil(boundaries ...Kind) {
	if p.atAny(boundaries...) {
		return
	}
	start := p.position
	for !p.atAny(boundaries...) {
```
`recoverRegion` is `recoverUntil` that additionally records and returns the byte extent (`start` token's span start to the final consumed token's span end) it discarded — same loop structure, same boundary-check convention, no new control flow shape.

**Diagnostic identity — confirm untouched** (`internal/compiler/diagnostic/diagnostic.go:39-47`, `Error`):
```go
func Error(code string, span Span, message string, causes ...Cause) Diagnostic {
	identity := struct {
		Schema string
		Code   string
		Span   Span
		Causes []Cause
	}{Schema: Schema, Code: code, Span: span, Causes: causes}
	encoded, _ := json.Marshal(identity)
	sum := sha256.Sum256(encoded)
	return Diagnostic{Schema: Schema, ID: "diagnostic:" + hex.EncodeToString(sum[:12]), Code: code, Severity: "error", Primary: span, Message: message, Causes: causes}
}
```
`Error` already accepts variadic `causes` and already hashes `Causes` into identity — **zero changes needed to `diagnostic.go`**; confirms D-14-37's "no schema bump" ruling.

---

### `internal/compiler/check/diagnostic_distinctness_test.go` (NEW, DX-08's gate)

**Analog 1 (non-inertness shape):** `TestInjectorMarkerCountGuardIsNotInert` (excerpted above) — `TestDiagnosticDistinctnessGuardIsNotInert` follows the identical "run the metric over a fixed pre-mutated/frozen input and assert it does NOT report the healthy result" shape, except the fixture here is `testdata/distinctness/collision_control.json`, a **frozen, never-regenerated** capture (parallel to `cmd/lang-repair/testdata/*_capture.json`'s discipline below) rather than a live mutated copy.

**Analog 2 (frozen-capture discipline):** `cmd/lang-repair/testdata/*_capture.json` — checked-in JSON captures that are read and asserted against, never regenerated by the test itself. `collision_control.json` must be captured **before** the parser fix lands and never touched again — same "historical artifact, assert-a-property-not-a-golden-match" rule.

**Core predicate (from RESEARCH.md/CONTEXT.md D-14-40, no direct in-repo precedent, build fresh but keep it in this file, ~60 lines total across 3 assertions):** lex each fixture with `internal/compiler/syntax.Lex`, keep only non-trivia token kinds, compare kind-sequences pairwise for corpus distinctness; then compare `diagnostic.ID` sets and `result:` ID sets across the trio.

---

### `cmd/lang-repair/repair.go` (EDITED — DX-09's additive fields)

**Analog:** itself, current shape.

**The `Outcome` struct to extend additively** (`repair.go:112-116`):
```go
type Outcome struct {
	Status          string `json:"status"`
	DiagnosisCode   string `json:"diagnosis_code,omitempty"`
	RepairKind      string `json:"repair_kind,omitempty"`
	SubprocessCount int    `json:"subprocess_count"`
}
```
Add three `omitempty` siblings: `DiagnosisCodes []string`, `DeclineReason string`, `BestApplicability string` — no existing field's type changes (D-14-42's explicit constraint).

**The decline-vocabulary block to extend** (`repair.go:99-108`, the existing `Code*` const block pattern):
```go
const (
	CodeStdoutCapExceeded = "repair.stdout_cap_exceeded"
	CodeSubprocessFailed  = "repair.subprocess_failed"
	CodeDecodeFailed      = "repair.decode_failed"
	CodeSpanOutOfRange    = "repair.span_out_of_range"
	CodeSourceReadFailed  = "repair.source_read_failed"
	CodeSourceWriteFailed = "repair.source_write_failed"
)
```
Add `repair.none_offered`, `repair.none_eligible`, `repair.no_diagnostics` beside this block, same naming convention (`repair.<snake_case>`).

**`selectRepair`, the exact function to widen** (`repair.go:200-209`):
```go
func selectRepair(result checkResult) (jsonRepair, string, bool) {
	for _, d := range result.Diagnostics {
		for _, r := range d.Repairs {
			if driverEligible(r) {
				return r, d.Code, true
			}
		}
	}
	return jsonRepair{}, "", false
}
```
Widen the return to `(jsonRepair, decision, bool)` where `decision` carries the `decline_reason` classification on the false path — the discarded information (`""`) becomes the new signal. Document-order iteration is already deterministic (stated in the doc comment above this function) — reuse it as `diagnosis_code`'s definition (first diagnostic's code in document order).

**Boundary constraints to respect (verified untouched):**
- `driverEligible` (`repair.go:196-200`{approx}) stays re-declared locally, never imported from `internal/` (`cmd/lang-repair/import_boundary_test.go:1-30`'s `forbiddenImportSubstr = "/internal/"` scan).
- Every new field is read off already-decoded `jsonDiagnostic.Code` or a new local `const` — no new prose field is decoded, keeping `cmd/lang-repair/antitheater_test.go:625-641`'s `TestProseScrambleLeavesRepairBehaviourIdentical` clean.

---

### `cmd/lang-repair/repair_diagnosis_test.go` (NEW)

**Analog 1 (stand-in-binary / prose-scramble harness):** `cmd/lang-repair/antitheater_test.go:630-641`:
```go
func TestProseScrambleLeavesRepairBehaviourIdentical(t *testing.T) {
	standinBinary := buildStandin(t)
	t.Run("match", func(t *testing.T) {
		testSourceClassProseScramble(t, standinBinary, session.MatchInjector{}, "heldout_match_defect.lang", "match")
	})
	...
}
```
`TestUnrepairableDiagnosisGuardIsNotInert` reuses `buildStandin(t)` and the same stand-in-binary pattern to emit a synthetic `unrepairable` outcome with an empty `diagnosis_code` and assert the new guard fails.

**Analog 2 (the existing gate this test's positive path extends):** `cmd/lang-repair/repair_test.go`'s `TestUnrepairableDefectFailsTheGate` — same "assert `unrepairable` is never silently treated as a pass" posture, now additionally requiring a non-empty `diagnosis_code`/`decline_reason` on every such outcome across `cmd/lang-repair/testdata/*_capture.json`.

---

## Shared Patterns

### Closed vocabulary enforced by a Go test
**Source:** `debtRegisterSeverities` + `debtRegisterLandingPhaseExemptions` (`internal/compiler/session/session_test.go:2581-2594`)
**Apply to:** the grade vocabulary (EVD-02), witness-kind vocabulary (`probe:`/`callsite:`/`escape:`/`env:`, EVD-03/04), the reconciliation verdict vocabulary (`renamed`/`superseded`/`obsolete-by-design`/`under-scoped`, D-14-12), the closed `Landing phase` form (D-14-24), and `decline_reason` (DX-09).
```go
var debtRegisterSeverities = map[string]bool{"blocker": true, "warning": true, "info": true}
var debtRegisterLandingPhaseExemptions = map[string]string{
	"02-DEBT.md": "written 2026-09-04, before the landing-phase column existed; frozen prior art",
}
```

### `.planning/**` path resolution — never `filepath.Glob` directly
**Source:** `phaseArtifactGlob` (`internal/compiler/session/session_test.go:2504-2520`)
**Apply to:** every new `.planning/**`-scanning guard (groundedness lint, grade cap, debt-register extension).

### Markdown table parsing with frontmatter count cross-check
**Source:** `checkDebtRegister` + `debtRegisterTable` (`internal/compiler/session/session_test.go:2628-2764`)
**Apply to:** `*-VALIDATION.md` grading (EVD-02), `*-DEBT.md` witness columns (EVD-03/04), the two generated views' non-zero-row-count check (D-14-14c).

### Non-inertness proof: seed one fault, assert red
**Source:** `TestInjectorMarkerCountGuardIsNotInert` (`internal/compiler/session/session_phase6_injectors_test.go:823-840`)
**Apply to:** every new guard in this phase — `TestVerificationGroundednessIsNotInert`, `TestValidationGradeCapIsNotInert`, `TestSuppressionWitnessGuardIsNotInert` (three seeded faults), `TestDiagnosticDistinctnessGuardIsNotInert`, `TestUnrepairableDiagnosisGuardIsNotInert`.

### Exact-name test resolution (never substring/regex-loose)
**Source:** `scripts/assert-go-tests.sh:8-42` (`selection()`)
**Apply to:** D-14-03's grade-derivation ladder and D-14-10's static test-index resolver — both must resolve exact identifiers, never a partial match.

### No silent skips — unparseable input is a violation, not a pass
**Source:** every existing guard's `t.Fatalf`-on-anything-unexpected posture (no `continue`-past-malformed-row anywhere in `checkDebtRegister`); reinforced by `NormalizeApplicability` (never defaults toward driver-eligible) and `unrepairable` never being conflated with a pass (`cmd/lang-repair/repair_test.go`).
**Apply to:** D-14-15's line-scanner ("unparseable verification command" is a violation), D-14-23's witness grammar ("a missing symbol is a FAIL, not a zero"), DX-09's decline path.

### Frozen, never-regenerated golden artifacts
**Source:** `cmd/lang-repair/testdata/*_capture.json`
**Apply to:** `testdata/distinctness/collision_control.json` (D-14-39) and the two `.planning/*.md` generated views (D-14-13) — regenerate-in-memory-and-compare, never checked-in-and-hand-edited.

### Import/decode boundary discipline
**Source:** `cmd/lang-repair/import_boundary_test.go` (`forbiddenImportSubstr = "/internal/"`) + `cmd/lang-repair/antitheater_test.go` (`TestRepairDriverDecodesNoProseFields`)
**Apply to:** DX-09's `Outcome` extension — every new value must come from an already-decoded code or a new local const, never a newly-decoded prose field.

## No Analog Found

None — every file in this phase's scope has a strong (exact or role-match) analog already in the tree, consistent with RESEARCH.md's framing that this phase's entire risk is in derivation logic, not parsing/scanning substrate.

## Metadata

**Analog search scope:** `internal/compiler/session/*_test.go`, `internal/compiler/session/session_phase5*.go`, `internal/compiler/session/session_phase6_injectors*.go`, `internal/compiler/session/session_phase6_budget*.go`, `internal/compiler/syntax/parser.go`, `internal/compiler/syntax/lexer.go`, `internal/compiler/diagnostic/diagnostic.go`, `cmd/lang-repair/*.go`, `scripts/assert-go-tests.sh`, `.planning/milestones/M002-phases/09-.../09-VALIDATION.md`, `.planning/LANGUAGE-MATURITY.md`.
**Files scanned:** 17 target files against ~12 analog source files, all confirmed git-tracked via `git ls-files`.
**Pattern extraction date:** 2026-09-17
