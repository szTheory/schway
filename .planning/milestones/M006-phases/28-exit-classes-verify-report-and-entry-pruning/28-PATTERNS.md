# Phase 28: Exit Classes, Verify Report, and Entry Pruning - Pattern Map

**Mapped:** 2026-10-10
**Files analyzed:** 22 new/modified
**Analogs found:** 22 / 22 (one partial: AST self-test negative control)

All analog paths below were confirmed git-tracked (`git ls-files`) or are plain source/testdata trees. No gitignored capability mirrors are referenced. Line numbers are against HEAD 32f8815; re-grep at plan time.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/compiler/exitcode/exitcode.go` (new) | utility (const table) | transform | `cmd/schway/main.go:25-28` const block + `native/native.go:32-37` contract consts | role-match |
| `internal/compiler/exitcode/exitcode_test.go` (new, pin + uniqueness) | test | batch | `cgen/cgen_phase273_test.go:87-97` (`TestPhase273CapacityContractLiteralsAgree`) | role-match |
| `internal/compiler/exitcode/scan_test.go` (new, AST scan + self-test) | test | batch (file walk) | `session/session_phase16_emitter_inventory_test.go` (`phase16PublicEmitterCalls`) | exact (AST style) |
| `internal/compiler/callgraph/prune.go` (new `PruneToReachable`) | utility | transform | `callgraph.go` `reachableClosureSize` (350-367) + `buildAdjacency` (381) | exact |
| `internal/compiler/callgraph/callgraph.go` (`main` precedence in `EntryFunction`) | utility | transform | itself (279-345) | exact |
| `internal/compiler/callgraph/callgraph_entry_test.go` (rewrite 130-158; add prune tests) | test | transform | itself | exact |
| `internal/compiler/session/session_entry.go` (new `resolveEntry`) | service | request-response | `session.go:1209-1215` call site + `callgraph.EntryAmbiguousError` mapping | role-match |
| `internal/compiler/session/session_admit.go` (new `admitProgram`) | service | request-response | `session.go` `CheckCommandFile` (1066-1106) + `peerRefusalDiagnostic` (839) | exact |
| `internal/compiler/session/session.go` (call-site collapse, `runNative` split) | service | request-response | itself (511, 602, 1122, 1209, 1235, 1305-1340, 1447) | exact |
| `internal/compiler/session/session_phase5_mismatch.go:393`, `session_app_verify.go:263` | service | request-response | same sites | exact |
| `internal/compiler/session/session_entry_inventory_test.go` (new meta-test `TestEntryFunctionCalledOnlyViaResolveEntry`) | test | batch | `session_phase16_emitter_inventory_test.go` | exact |
| `internal/compiler/session/session_phase28_manifest_test.go` + `testdata/phase28/entry-exit-class-manifest.json` (new env-guarded writer) | test + golden | batch | `session_phase273_identity_test.go` (`SCHWAY_PHASE273_WRITE_GOLDEN`, 128-160) | exact |
| `internal/compiler/protocol/protocol.go:681` `ExitCode` | utility | transform | itself | exact |
| `cmd/schway/main.go` (consts, app-run literals, extract `appRunExitCode`) | controller | request-response | itself (25-28, 232-262) | exact |
| `cmd/schway-repair/main.go:12-15` | controller | request-response | `cmd/schway/main.go:25-28` | exact |
| `internal/compiler/native/native.go:37` (alias) | config | n/a | itself | exact |
| `internal/compiler/cgen/cgen.go:735,774` and cgen exit-literal scan test | test | batch | `cgen_phase273_test.go:87-97` | role-match |
| `internal/compiler/session/session_app_verify.go` (digest, schema/2, ID cap, oversize error) | service | batch | itself (72-84, 296-360, 455-481, 552-561) | exact |
| `internal/compiler/session/session_phase273_appverify_test.go` (rewrite `Engines` asserts) | test | batch | itself (`phase273RequireUnsupportedCapacity`, 60-80) | exact |
| `AppVerifyReportBudget` / digest / oversize tests (new) | test | batch | `session_phase273_appverify_test.go` | role-match |
| exit-table docs test + README/wiki table | test + doc | transform | `session/self_describing_docs_test.go` (AST/doc derivation) | role-match |
| flip pins: `native_test.go:1522-1564`, `session_test.go:1818`, `session_phase16_production_paths_test.go:20-37`, `session_phase271_admission_test.go:169-195`, zero_call pins | test | request-response | the files themselves | exact |

## Pattern Assignments

### `internal/compiler/exitcode/exitcode.go` (utility, transform)

**Analog:** `cmd/schway/main.go:25-28` (named const block) and `native/native.go:32-37` (documented published contract).

```go
const (
	exitSuccess       = 0
	exitInvalidSource = 2
	exitOperational   = 3
	exitUsage         = 64
)
```
```go
// Event-capacity contract: generated execution/conformance C exits with
// EventCapacityExitStatus ...
const (
	CodeEventCapacityExhausted = "native.event_capacity_exhausted"
	EventCapacityExitStatus    = 69
	EventCapacityStderr        = "schway: event capacity exhausted\n"
)
```
Copy: one const block with per-name doc comments carrying the "never renumber" notes for 65/69. Package imports nothing. `func SignalExit(n int) int { return SignalBase + n }`. After the change `native.EventCapacityExitStatus = exitcode.EventCapacity` (keep the alias; `native.go:630` compares it).

**Consumers to rewire:** `protocol.ExitCode` (below), `cmd/schway/main.go` 25-28/242/245/251/254/257/260, `cmd/schway-repair/main.go` 12-15 (note repair keeps `Unrepaired = 1`).

### `internal/compiler/protocol/protocol.go` `ExitCode` (utility, transform)

**Analog:** itself, lines 681-697.
```go
func ExitCode(status string) int {
	switch status {
	case StatusPass:
		return 0
	case StatusInvalid:
		return 2
	case StatusOperational:
		return 3
	case StatusMismatch:
		return 4
	default:
		return 64
	}
}
```
Keep the `Status*` switch; return `exitcode.Success/Invalid/Operational/Mismatch/Usage`. `protocol` already imports cache/diagnostic/interp, so a leaf import is cycle-safe.

### `internal/compiler/exitcode/scan_test.go` (test, file walk + AST)

**Analog:** `session/session_phase16_emitter_inventory_test.go` lines 1-95. Reuse: `filepath.WalkDir` over a root from `testsupport.ProjectPath(...)`, `parser.ParseFile(fset, path, nil, 0)`, `ast.Inspect`, findings recorded as `relpath:name:line` via `fset.Position(node.Pos()).Line`, sorted for stable output, `t.Fatalf` on parse failure.

**Imports pattern** (inventory test lines 3-16): `encoding/json`, `go/ast`, `go/parser`, `go/token`, `os`, `path/filepath`, `sort`, `strconv`, `strings`, `testing`, plus `internal/compiler/testsupport`. Use `package exitcode_test` (external) so the leaf package stays import-free.

**Core walk pattern** (lines 33-93):
```go
err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
	if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_generated.go") {
		return err
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	...
	ast.Inspect(file, func(node ast.Node) bool { ... })
```
Factor the detector as a pure function over `(fset, *ast.File) []finding` so the self-test can call it on `parser.ParseFile(fset, "synthetic.go", src, 0)`. Filter out `_test.go` files, `cmd/schway-repair/repair.go`, and the exitcode package. Scope dirs per D-11/D-21: session, native, protocol, cgen, `cmd/schway`, `cmd/schway-repair/main.go`.

**Negative-control pattern:** there is no single existing "detector fires on synthetic source" test; nearest analogs are `session/evidence_grade_test.go:710-760` (synthetic rows, one per ladder rung) and `session/qlt01_test.go:144-174` (synthetic hazard rows "used only to prove the audit can go red"). Copy the table-driven shape: one synthetic source per D-11 position (os.Exit arg, return in exit-flow func, `exit*` const/var, `ExitCode`/`*ExitStatus` field, comparison/`128+x`, switch case) asserting >= 1 finding, plus one clean synthetic source asserting zero findings, plus a real-tree assertion that the walk scanned > 0 files (so it cannot pass vacuously).

**Known false-positive:** `native/sanitize.go:226-287` (`exitCode == 0`, `exitCode := -1`). D-21 requires `0` to be flagged in exit paths (use `exitcode.Success`) while `-1` sentinel is out of scope; decide per-file whether sanitize.go child status is an exit position. `native_app.go:424` (`ExitCode: 0`) must move to `exitcode.Success`.

### `internal/compiler/exitcode/exitcode_test.go` (pin + uniqueness)

**Analog:** `cgen/cgen_phase273_test.go:87-97` `TestPhase273CapacityContractLiteralsAgree` (asserts constants/literals agree). Pattern: assert `exitcode.AppInputRefused == 65`, `exitcode.EventCapacity == 69`, `native.EventCapacityExitStatus == exitcode.EventCapacity`, `SignalExit(2) == 130`, and that the values of Invalid/Operational/Mismatch/Usage/AppInputRefused/EventCapacity/AppTimeout/AppLaunchError/SignalBase are pairwise distinct (Success 0 and Unrepaired 1 are in different domains; document and handle the 3/Operational shared across compiler/repair).

### `internal/compiler/callgraph/prune.go` (utility, transform)

**Analog:** `callgraph.go` BFS `reachableClosureSize` (350-367) and `buildAdjacency` (381).
```go
func reachableClosureSize(adjacency map[string][]string, root string) int {
	visited := map[string]bool{root: true}
	queue := []string{root}
	count := 0
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, callee := range adjacency[current] {
			if visited[callee] { continue }
			visited[callee] = true
			count++
			queue = append(queue, callee)
		}
	}
	return count
}
```
Copy: iterative explicit-queue BFS (package discipline: never native recursion). `PruneToReachable(program core.Program, entryID string) (core.Program, error)`: call `buildAdjacency(program)` (propagates `unresolvedCalleeError`), BFS from entryID, then rebuild `Functions` by filtering the ORIGINAL `program.Functions` slice in order (declaration order preserved). Copy DataTypes/foreign decls/Module fields unchanged (assign `pruned := program` then replace `Functions`). Avoid the shape `len(x.Functions) != 1` (LANGUAGE-MATURITY census, `self_describing_docs_test.go`) and never index `Functions[0]` (`TestSessionRunSitesDoNotIndexFunctionsZero`, callgraph_entry_test.go:254). Return the original program value (or deep-equal copy) when everything is reachable, so `reflect.DeepEqual(pruned, p)` holds (nil-vs-empty slice caution: build with `make([]core.Function, 0, n)` only if the input slice is non-nil, or return `program` unchanged when count equals).

### `internal/compiler/callgraph/callgraph.go` `EntryFunction` (D-01)

**Analog:** itself, 279-345. Insert `main` precedence right after `byID` is built and before the `len(roots)` branch: if a function named exactly `main` exists, `resolve(thatID)`. Keep the `len(roots)==1` / zero-roots / closure-tiebreak (`bestCount != 1` -> `&entryAmbiguousError{roots: roots}`) logic and the `Error()` text byte-identical (two goldens pin it). `entryAmbiguousError.Roots()` already exists for the diagnostic message; the "add `fn main`" hint belongs only in the session diagnostic.

### `internal/compiler/callgraph/callgraph_entry_test.go` (rewrite 130-158)

**Analog:** the existing `TestEntryFunctionEmitsUnreachableFunctions` (checkedPhase11Fixture helper, `entry.Name != "main"`, scan `program.Functions` for `orphan`). Rewrite: assert `orphan` present in the checked program, absent from `PruneToReachable(program, entry.ID)`, entry unchanged on re-resolution of the pruned program, exactly one in-degree-zero root; add a dated 2026-10-10 note citing D-05. Add `TestPruneAllReachableDeepEqual` over the fully reachable fixtures and the unique-root property test (also required by D-19 re: `executionpeer.resolveEntry`, which keeps its own tiebreak copy at `executionpeer.go:272-318`).

### `internal/compiler/session/session_entry.go` (`resolveEntry`)

**Analog call-site pattern:** `session.go:1209-1215` (RunInterpreterApplication) and the `callgraph.EntryAmbiguousError(err)` mapping described in RESEARCH (callgraph.go:218-227: `Error()` is `callgraph.entry_ambiguous: ...`, `Code()` is `core.entry_ambiguous`).
```go
entry, err := callgraph.EntryFunction(program)
if err != nil {
	return interp.ApplicationResult{}, nil, err
}
```
New shape: `resolveEntry(program core.Program) (core.Function, core.Program, []diagnostic.Diagnostic, error)` (or returning a single diagnostic). On `callgraph.EntryAmbiguousError(err)` produce `diagnostic.Error(ambiguous.Code(), diagnostic.Span{}, msg)` where msg lists `Roots()` and says "add `fn main` to disambiguate". Other errors (e.g. unresolved callee) stay errors. Then call `callgraph.PruneToReachable`. Call-site map (only this helper may call `callgraph.EntryFunction`):

| Site | Disposition |
|---|---|
| session.go:511 `ProjectExecutionSchema2` | `entryOf` wrapper (the single allow-listed call) |
| 602 `Phase4CheckedProgram`, 1122 `RunInterpreter`, 1340 `runNative`, 1447 `BuildApplication` | `resolveEntry` |
| 1209 `RunInterpreterApplication` | `admitProgram` then `resolveEntry` |
| 1235 `interpreterInputs` | accept resolved entry; keep name, no `Functions[0]` |
| session_phase5_mismatch.go:393 | `entryOf` |
| session_app_verify.go:263 | hoist: resolve once before per-case loop |
| cgen/cgen_program.go:639 | cannot import session; allow-list in meta-test |

### `internal/compiler/session/session_admit.go` (`admitProgram`)

**Analog:** `CheckCommandFile`, session.go:1066-1106 (order and observer seams) plus `peerRefusalDiagnostic` (839-847) and `originPeerRefusalDiagnostic` (849).
```go
if !checkCommandPeerSeam {
	observeCheckCommandPeer("corevalidate")
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		result.Diagnostics = []diagnostic.Diagnostic{peerRefusalDiagnostic(validated, checked.Program.ModuleID)}
		...
	}
	checked.Program = validated.Program()
}
observeCheckCommandPeer("originvalidate")
if problems := originvalidate.ValidatePublished(checked.Program); len(problems) > 0 {
	... originPeerRefusalDiagnostic(problems[0], checked.Program, checked.Tree)
} else {
	observeCheckCommandPeer("pathoracle")
	if err := pathoracle.ValidateLocalOwnerPaths(checked.Program); err != nil {
		diagnostic.Error("core.pathoracle_refused", diagnostic.Span{}, err.Error())
```
Copy the order, observers and `checkCommandPeerSeam` gate verbatim into `admitProgram(checked CheckResult) (core.Program, []diagnostic.Diagnostic)` and make `CheckCommandFile` call it (so the existing seam tests stay green). `peerRefusalDiagnostic` already guards `len(validated.Problems) > 0` and falls back to `peerRefusalUnnamedCode`; do not copy the unguarded `Problems[0].Code` at session.go ~1116 and ~1335. Per D-18 order of operations is: Check -> admitProgram (full program) -> resolveEntry (prune) -> interp/cgen/app verify. Per D-19, admission wires into `check`, `build`, `RunInterpreterCommandFile`, `RunNativeCommandFile` ONLY. Do NOT put it in exported `RunInterpreter`/`RunNative`/`runNative` or `admitPublishedApplication`'s D-27.1-02 rule (session.go:1160-1169 comment). Split `runNative` into check+resolve+admit (command-file wrapper) and `runNativeChecked`. Drop the post-hoc `publishedOriginProblemFile` checks (session.go ~1289).

Replace in the old build/interpreter path (these are the W2 bugs):
```go
case applicationRefusedCore:
	...
	return interp.ApplicationResult{}, nil, fmt.Errorf("core validation failed: %s", adm.code)
```
with a returned diagnostic and nil error (session_phase271_admission_test.go:169, 176, 195 asserted the old text; flip to diagnostic asserts).

### `session/session_entry_inventory_test.go` (meta-test)

**Analog:** `phase16PublicEmitterCalls` + `TestPhase16PublicEmitterConsumerInventory` (inventory test lines 33-120): resolve import alias for `github.com/szTheory/schway/internal/compiler/callgraph`, reject dot-imports (`t.Fatalf("fail-closed: ... dot import in %s", path)`), match `*ast.SelectorExpr` calls with `Sel.Name == "EntryFunction"`, record `relpath:Name:line`. Allow-list: callgraph package, `_test.go`, the single `entryOf` call, and `cgen_program.go:639`. Prefer an in-test allow-list over a JSON registry (smaller than the phase16 registry).

### `session_phase28_manifest_test.go` + `testdata/phase28/entry-exit-class-manifest.json`

**Analog:** `session_phase273_identity_test.go` lines 1-60 (constants: schema, min-corpus floor `phase273MinAcceptedCorpus`, golden/entry structs) and 128-160 (env-guarded writer).
```go
if os.Getenv("SCHWAY_PHASE273_WRITE_GOLDEN") == "1" {
	revision := strings.TrimSpace(os.Getenv("SCHWAY_PHASE273_GOLDEN_REVISION"))
	if revision == "" {
		t.Fatal("SCHWAY_PHASE273_GOLDEN_REVISION must name the commit the golden is generated at")
	}
	golden := phase273CorpusGolden{Schema: phase273GoldenSchema, Revision: revision, Entries: observed}
	encoded, err := json.MarshalIndent(golden, "", "  ")
	...
	os.MkdirAll(filepath.Dir(goldenPath), 0o755)
	os.WriteFile(goldenPath, append(encoded, '\n'), 0o644)
	return
}
```
Use `SCHWAY_PHASE28_WRITE_GOLDEN` / `SCHWAY_PHASE28_GOLDEN_REVISION`, schema `phase28.entry-exit-class-manifest/1`, sorted entries by path, a min-count floor so the walk cannot go inert, and the allow-list-map idiom (`phase273ChangedConformancePaths`, `phase273NewlyAcceptedPaths`) for the 4 dead-function files (gate_corpus, unreachable, twin_alpha, twin_mirror) plus the zero_call flip. Per-entry fields: check class/code, interp-run class/code, admission class/code, entry source (`main`/`unique`/`tiebreak`/`ambiguous`), dead-function IDs. Emit C only through `session.Phase16ControlNativeC(checked.Program, "phase28-...")` (new tests must not call `cgen.Emit*`; the phase16 inventory tracks that). Commit the manifest BEFORE behavior changes (first plan). Existing goldens `testdata/phase273/conformance-corpus-hashes.json` and `testdata/phase272/accepted-corpus-hashes.json` remain the byte-identity safety net.

### `cmd/schway/main.go` app-run (controller)

**Analog:** itself, 236-262. Extract a pure `appRunExitCode(outcome native.RunOutcome, err error) int` so 124/125/128+n can be table-tested without a 30 s subprocess (RESEARCH). Existing branch structure to preserve:
```go
if errors.As(err, &toolError) && (toolError.Code == "native.input_too_long" || toolError.Code == "native.input_contains_nul") { ... return 65 }
... return 125
switch outcome.Kind {
case native.RunExited:   return outcome.ExitCode
case native.RunSignaled: return 128 + outcome.SignalNumber
case native.RunTimedOut: ... return 124
case native.RunLaunchError: ... return 125
default: ... return 125
```
Keep stderr writes (`schway app run: ...`) in the caller; the pure function returns only the int using `exitcode.AppInputRefused/AppLaunchError/SignalExit/AppTimeout`. Test: `cmd/schway/main_test.go` `TestAppRunExitCode` plus one real `app run /nonexistent -- 1` -> 125 CLI case.

### `cmd/schway-repair/main.go:12-15`

**Analog:** `cmd/schway/main.go:25-28`. Replace with `exitcode.Success`, `exitcode.Unrepaired`, `exitcode.Operational`, `exitcode.Usage`. Exclude `repair.go` from the scan.

### `cgen` exit-literal handling (cgen.go:735, 774; cgen_program.go many)

**Analog:** `cgen/cgen_phase273_test.go:87-97`. D-21/RESEARCH recommend a string-literal scan test rather than editing emitters (zero byte-identity risk): AST-walk cgen non-test sources for `*ast.BasicLit` strings, regex `exit\((\d+)\)` and `return (\d+);`, assert each N is in {0, `exitcode.AppInputRefused`, `exitcode.EventCapacity`}. Name `TestEmittedCExitLiteralsMatchTable`. Only if the planner chooses to edit cgen.go:774, the output string must stay byte-identical (both corpus goldens hash it; `phase273NamedCapacityRecorderLine` pins the exact text).

### `session/session_app_verify.go` (EVD-13)

**Analog:** itself.
- Struct (72-84): add `OutcomeDigest *ReplayOutcomeDigest \`json:"outcome_digest,omitempty"\`` beside `Engines map[string]execution.Execution \`json:"engines,omitempty"\``.
- Branches (329-355): the asymmetric branch and the both-level branch currently set `Engines = map[string]execution.Execution{"interpreter": interpreted[reportIndex]}`. Replace both with the digest and leave `Engines` nil.
```go
raw, _ := json.Marshal(interpreted[reportIndex])   // handle err; fail closed
report.Cases[reportIndex].OutcomeDigest = &ReplayOutcomeDigest{
	Engine: "interpreter", OutcomeKind: actual.Kind, OutcomeValue: actual.Value,
	EventCount: len(interpreted[reportIndex].Events),
	EvidenceSHA256: replayDigest(raw), EvidenceBytes: len(raw), ExhaustedLevel: level,
}
```
- Hash helper (455-458): `replayDigest(data []byte) string` (sha256 hex) — reuse.
- Diagnostic (479-481): `eventCapacityDiagnostic(level)` verbatim.
- Status logic unchanged: equal -> `ReplayStatusUnsupported`, unequal -> `ReplayStatusFail`; asymmetric -> `ReplayStatusFail`; `finalizeReplayReport` unchanged.
- Oversize (561): `fmt.Errorf("application verification report exceeds %d bytes", MaxReplayCasesBytes)` becomes a named `ErrApplicationVerificationReportOversize` (wrap with size+limit, `errors.Is`-testable); the check precedes `os.CreateTemp`, keep that so no partial file is written.
- Schema: `ApplicationVerifySchema` (line 28; also used at ~201) -> `"lang.app-verification/2"`. Also update `examples/phase22/README.md` ~100.
- `MaxReplayCaseIDBytes = 128`: add next to `MaxReplayCasesBytes`/`MaxReplayCaseCount`; reject in the `DecodeReplayCases` loop (which currently checks only empty/duplicate IDs) with a printable-ASCII check.
- Hoist `callgraph.EntryFunction` (line 263) out of the per-case loop into one `resolveEntry` before the loop; add check/admit/resolve so source refusals exit 2 (D-21), while capacity/tool outcomes stay exit 3 (`cmd/schway/main.go` ~296-304 keeps `exitcode.Operational` for `err != nil || !report.Verified`).

### `session_phase273_appverify_test.go` (rewrite)

**Analog:** itself, `phase273RequireUnsupportedCapacity` (60-80). Keep the status/prefix/`"at -O0"`/`"not a mismatch"` checks; replace the `len(caseReport.Engines) != 1` + `Engines["interpreter"]` block with:
```go
if len(caseReport.Engines) != 0 { t.Fatalf(...) }
d := caseReport.OutcomeDigest
if d == nil || d.Engine != "interpreter" || d.OutcomeKind != "returned" || d.OutcomeValue != "120" { t.Fatalf(...) }
```
(Do not delete the assertion; rewrite per D-15.) Also rewrite the one-level `fail` test at ~274-279 to digest asserts. New tests: `TestAppVerifyReportBudget` (16 x max ID/input/value must be <= `MaxReplayCasesBytes/2`), `TestAppVerifyOutcomeDigest` (re-marshal + re-hash), `TestAppVerifyReportOversize` (synthetic), `TestAppVerifySixteenExhausted` (single real-clang test), n = 14 pass / 15, 172, 173, 1000 unsupported. `phase273RequireNoSignalText` (marshalled report must not contain "signal/abort/run_signaled") remains valid.

### Pin flips (edit deliberately, dated note)

- `native/native_test.go:1522-1527` `cutAfterCheckRefusal` and its use at 1564: delete helper, use `refused(...)` with exit 2 `core.foreign_origin_omitted`. Keep 1507/1518 cut fixtures at exit 3 `native.tool_failure`.
- `session/session_test.go:1818-1822`: native `foreign_origin_omitted` -> `StatusInvalid` / `core.foreign_origin_omitted`.
- `session/session_phase16_production_paths_test.go:20-37`: swap `restrict_borrow.schway` for a check-clean cut fixture (e.g. `testdata/phase4/acquire_three_success.schway`; `restrict_borrow` is in `knownAdmissionDivergences`).
- `session_phase271_admission_test.go:169,176,195`: exact error text -> diagnostic asserts; `*/verify` rows (182, 201, 248) change under D-21.
- zero_call flips: `cgen_program_test.go:1239-1281 TestEmitProgramZeroCallEdges`, `session_phase11_differential_test.go:357-396`, `session_payload_replay_test.go:186 payloadCorpusExpectedSkips`, both corpus goldens (regenerate once, allow-listed).
- `session_admission_divergence_test.go`: add `testdata/phase12/payload_borrow_interaction.schway` -> `core.entry_ambiguous`; with D-18 (admit full program) the gate_corpus entry stays.
- `phase11CompareFourTiers` must keep returning the checked (unpruned) program (Pitfall 7).

### New cross-surface test `TestExitClassesUniformAcrossSurfaces`

**Analog:** `native/native_test.go` `TestShippedBinaryFourSubcommandCorpusMatrix` and `phase4CorpusMatrix` (helpers `refused(...)`, shipped binary, clang). Fixtures: `testdata/phase07/duplicate_function_name.schway` (corevalidate refusal, `core.duplicate_function_id`) and `testdata/phase4/foreign_origin_omitted.schway`, plus `testdata/phase12/payload_borrow_interaction.schway` for `core.entry_ambiguous`. Assert identical exit 2 and code on check / build / run --engine=interpreter / run --engine=native.

## Shared Patterns

### Independent validator layers (07-10 refusing union)
**Source:** `session.go:830-847` doc comment on `peerRefusalDiagnostic`.
**Apply to:** `admitProgram`. Share formatting only; never reconcile derivations; each layer called independently in fixed order.

### Fail-closed, diagnostics not errors for source refusals
**Source:** `CheckCommandFile` returns `protocol.Result` with `StatusInvalid` and `diagnostic.Error(code, diagnostic.Span{}, detail)`. Tool/IO failures stay errors -> `StatusOperational`.
**Apply to:** build, run (both engines), app verify refusals.

### Inventory/meta-test AST style
**Source:** `session_phase16_emitter_inventory_test.go` (alias resolution, dot-import rejection, sorted `path:name:line` output, `testsupport.ProjectPath`).
**Apply to:** EntryFunction-call-site test, exit-literal scan, cgen exit-string scan.

### Env-guarded golden writer
**Source:** `session_phase273_identity_test.go:128-160`.
**Apply to:** `testdata/phase28` manifest.

### Emit C through `Phase16ControlNativeC` in tests
**Source:** `session_phase273_identity_test.go:104`; reason in file header (phase16 emitter inventory tracks direct `cgen.Emit`/`EmitNative` in tests).

## No Analog Found

| File | Role | Data Flow | Reason |
|---|---|---|---|
| Detector self-test with negative control (inside `scan_test.go`) | test | batch | Only loosely analogous synthetic-row tests exist (`evidence_grade_test.go:710-760`, `qlt01_test.go:144-174`); no existing AST detector with a synthetic-source self-test. Use RESEARCH detector spec + the table-driven shape above. |
| `appRunExitCode` pure extraction | controller | transform | No existing pure exit-mapper in `cmd/`; derive from main.go 236-262. |
| Exit-table docs test (`TestExitCodeDocsTable`) | test | transform | Closest is `self_describing_docs_test.go` (doc/code derivation); no existing README table pin. |

## Corpus facts for the planner (from RESEARCH audit; re-verify in plan 1)

- 155 tracked `.schway`; 90 accepted by the checker. Tiebreak-without-`main` dependence among accepted-and-resolving programs: 0 (D-03 corpus audit satisfied; list in manifest).
- Byte-identity changes by design: `testdata/phase11/multi_function_gate_corpus.schway`, `multi_function_unreachable.schway`, `testdata/phase13/heldout_shared_callee_twin_alpha.schway`, `..._mirror.schway`; flip to runs: `multi_function_zero_call.schway`; stays ambiguous: `testdata/phase12/payload_borrow_interaction.schway` (check newly exit 2).
- Hard-coded shape matchers that may shift on pruned fixtures: cgen_program.go 1159/1214/1341/1374, interp.go 212-269, pathoracle.go:413. Run focused phase24/phase5 fixtures before and after.
- Confirm cgen C `main` mangling unchanged (cgen.go:171, 206).

## Metadata

**Analog search scope:** `internal/compiler/{session,callgraph,cgen,native,protocol,executionpeer}`, `cmd/schway`, `cmd/schway-repair`, `testdata/phase272`, `testdata/phase273`.
**Files scanned:** ~20 read/grepped in this pass plus the RESEARCH-verified set.
**Pattern extraction date:** 2026-10-10
