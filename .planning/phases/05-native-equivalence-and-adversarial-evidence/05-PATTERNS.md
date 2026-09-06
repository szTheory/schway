# Phase 5: Native Equivalence and Adversarial Evidence - Pattern Map

**Mapped:** 2026-09-05
**Files analyzed:** 13 new/modified surfaces derived from 05-CONTEXT.md/05-RESEARCH.md
**Analogs found:** 11 / 13 (2 have no analog and must lean on RESEARCH.md's verified toolchain commands instead)

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/compiler/cgen/cgen.go` (by-pointer lowering path, additive sibling of `emitLinear`) | codegen/emitter | transform | `internal/compiler/cgen/cgen.go`'s `emitLinearForeignOutputSupport` (D-04-20's additive streaming-event sibling of `emitLinearOutputSupport`) | exact — same "new sibling emitter, old ones untouched" shape |
| `internal/compiler/cgen/cgen.go` (`restrict` attribute emission + `EmittedAttributes` population) | codegen/emitter | transform | `internal/compiler/cgen/cgen.go`'s `EmitForeignManifest`/`foreignManifestDocument`/`BannedOptimizerAttributes`/`ScanForBannedAttributes` (D-04-13) | exact — same field, same scan machinery, this phase only stops emitting the empty array |
| `internal/compiler/check/check.go` (alias-fact proof: exclusive-borrow-for-call-duration derivation) | compiler frontend / checker | transform | `internal/compiler/check/check.go`'s `analyzeStraightLine`/`analyzeArmBody` + `discoverLoanLastUses` (ownership/loan derivation) | role-match — new fact derived from the same loan-tracking machinery, not a new dispatch site |
| `internal/compiler/check/check.go` (`loanLivenessFixpoint` domain broadening + `discoverLoanLastUses` shadow-mode retirement) | compiler frontend / checker | transform | `internal/compiler/check/check.go`'s own `loanLivenessFixpoint` (lines 537-655, already used for branch arms) and `testOnlyForceUniformLoanJoin` fault-injection-seam precedent | exact — same function widened, same file's own shadow/fault-injection precedent |
| `internal/compiler/corevalidate/corevalidate.go` (independent attribute-justification re-derivation) | independent validator | request-response (pure re-derivation, no shared helper) | `internal/compiler/corevalidate/corevalidate.go`'s existing `validCIdentifier`/`foreignContractFieldsCSafe` checks (lines 1666-1781) at call sites lines 316/349/386 | exact — same "independent re-derivation, never reads what cgen wrote" posture (D-12) |
| `internal/compiler/corevalidate/corevalidate.go` (`Alias` field audit — D-05-36) | independent validator | request-response | `internal/compiler/corevalidate/corevalidate.go:1765` `foreignContractFieldsCSafe` (add `Alias` alongside `Fails`/`InitializedState`/`Capture`/`Retention`) | exact — one-line addition to an existing predicate covering the exact same field class |
| `internal/compiler/native/native.go` (`-O3 -flto` build configuration, `lane:native-differential-lto`) | native build orchestration | batch/transform | `internal/compiler/native/native.go`'s `Runner.Run` compile invocation (lines 61-117, `-std=c17 -Wall -Wextra -Werror -pedantic <optimization> -c ...`) | exact — same invocation shape, additive flag |
| `internal/compiler/native/native.go` (separate sanitizer binary build + run, structurally isolated) | native build orchestration | batch/transform | `internal/compiler/native/native.go`'s `Runner.CompileConformanceUnit` (lines 216-237) — an existing "compile a distinct artifact, never linked into the differential" precedent | role-match — same "separate build config, separate artifact" shape, applied to a linked+executed binary instead of a conformance-only unit |
| `internal/compiler/session/session.go` (new mutation runners: alias-fact injection, second-allocator TU, hostile UAF variant) | mutation-kill test harness | event-driven (fail-closed marker-count guard, mutate, run, count work) | `internal/compiler/session/session.go`'s `OwnedBackendMutationRunner` (67-119), `ReleaseOmissionMutationRunner` (139-179), `NonlocalPadOmissionMutationRunner`/`NonlocalLedgerOmissionMutationRunner` (203-281), `LayoutMutationRunner` (554-566) | exact — five worked examples of the exact shape to copy |
| `internal/compiler/session/session.go` (`Phase5RequiredControls()`) | config/registry | batch (fail-closed set comparison) | `internal/compiler/session/session.go:2004-2019` `Phase4RequiredControls()` | exact |
| `internal/compiler/session/session_test.go` (`TestPhase5RequiredControlsMatchScript`, `TestPhase5VerifierScriptContract`) | test | batch (set-equality assertion) | `internal/compiler/session/session_test.go:2109-2201` `phase4VerifierScriptText`/`phase4OwnControlIdentifiers`/`TestPhase4VerifierScriptContract`/`TestPhase4RequiredControlsMatchScript` | exact |
| `internal/compiler/session/session.go` (exclusion-list fail-closed field-routing test target: `Outcome`/event/ledger field enumeration) | comparator / validation | transform | `internal/compiler/session/session.go`'s `Phase4CompareThreeEngines`/`firstExecutionDisagreement` (lines 366-464) — the existing pairwise comparator this phase's axis-widening extends | exact — same comparator, D-05-20/D-05-21 widen its axis set and add the fail-closed routing test |
| `scripts/verify-phase5.sh` | build/gate script | batch | `scripts/verify-phase4.sh` (full file, 94 lines) | exact — explicit peer-not-fork instruction in CONTEXT.md |
| New reducer package (`internal/compiler/reduce/` or similar — name at Claude's discretion) | new package, core-structure transform | transform (fixpoint over 5 fixed moves) | No close analog exists — closest structural precedent is `internal/compiler/session/session.go`'s `TransposeReleaseOrder` (a pure core-mutation function operating on `core.Program`, no shared helper with corevalidate) | no analog (new subsystem) — see "No Analog Found" |
| New `lang.mismatch/0` schema type + round-trip test | schema/model | transform | `internal/compiler/evidence/evidence.go`'s `Manifest` struct (lines 49-74, `lang.evidence/1`) plus its round-trip goldens in `evidence_test.go` (`readGolden`/`goldenProduct`, e.g. `TestEvidenceGoldenStable` pattern at lines 104-106, 339-350) | role-match — new top-level schema, but the "struct + JSON tag + golden round-trip test" shape is the exact one to copy |
| QLT-01 control registry (new file/table, format at Claude's discretion) + `TestQLT01RegistryComplete` | new registry + audit test | batch (completeness scan, fail-closed) | `internal/compiler/session/session.go`'s `Phase4RequiredControls()` + `TestPhase4RequiredControlsMatchScript`'s set-equality pattern; `internal/compiler/originvalidate`/`corevalidate`'s `KnownEscape = "escape:coordinated-source-core-lie"` constant (corevalidate.go:20) for the escape-citation half | role-match — no registry table exists yet, but both halves (fail-closed set check, `KnownEscape` reuse) have exact precedent |
| `escape:callback-invocation-unsubjected` / `escape:coordinated-source-to-core-false-claim` (new escape constants) | config/registry (string constants) | — | `internal/compiler/session/session.go`'s `EscapeCoordinatedForeignBoundaryLie`/`EscapeNonlocalExitBelowThePad`/`EscapeForeignProcessExit` constants (lines ~1978-1989) plus their `VerifyCorpus`/`verifyForeignCorpus` wiring into `result.ExpectedEscapes` and `TestExpectedEscapesAreVisibleNotSolved`'s "declared but never detected" assertion | exact |

## Pattern Assignments

### By-pointer lowering path + `restrict` emission (`internal/compiler/cgen/cgen.go`)

**Analog 1 — additive emitter path:** `emitLinearForeignOutputSupport` (cgen.go:958), doc comment at cgen.go:946-957.

**Core pattern to copy** (the "new sibling, old paths untouched" shape):
```go
// emitLinearForeignOutputSupport is D-04-20's selection site: every function
// emitLinearForeign handles carries a core.ForeignContract (checked at that
// function's own entry), so it always takes the STREAMING event path
// ... This is a NEW sibling of emitLinearOutputSupport, never called from
// emitLinear/emitBranch, so every existing committed generated-C golden
// (Phase 1/2/3, and the frozen foreign layout fixture) is untouched (D-04-23).
func emitLinearForeignOutputSupport(out *strings.Builder, function core.Function, typeName string) {
```
Apply this exact shape: the by-pointer lowering must be a new function selected only when a Phase 5 fixture's parameter is exclusively borrowed for the call's full duration, never wired into `emitLinear`/`emitBranch`'s existing dispatch. The **blocking precondition** it must fix first, verified verbatim this session:
```go
// internal/compiler/cgen/cgen.go:157 — the by-value emission D-05-02 forbids retrofitting restrict onto
fmt.Fprintf(&out, "static %s %s(%s %s) {\n", typeName, functionName, typeName, parameterName)
```

**Analog 2 — the attribute/manifest field this phase populates, not `evidence.Manifest`:** `foreignManifestDocument`/`EmitForeignManifest` (cgen.go:1314-1413), `BannedOptimizerAttributes`/`ScanForBannedAttributes` (cgen.go:1527-1559).

```go
// cgen.go:1332-1335 — the field Phase 5 populates (currently always []string{})
EmittedAttributes    []string           `json:"emitted_attributes"`
UncheckedObligations []string           `json:"unchecked_obligations"`
...
// cgen.go:1406 — EmitForeignManifest currently always writes the empty array
Layout: contract.Layout, EmittedAttributes: []string{}, UncheckedObligations: uncheckedForeignObligations(),
```
```go
// cgen.go:1532-1534 — restrict is CURRENTLY a banned token; D-05-01 makes it
// the one narrow exception, only on cgen-owned by-pointer parameters
var BannedOptimizerAttributes = []string{
	"restrict", "noalias", "nothrow", "__attribute__((malloc))", "nonnull", "returns_nonnull",
}
```
D-05-03's narrowing must thread through here: `restrict` stays banned on **foreign extern declarations** (this scan target is unchanged) but becomes legal in `EmittedAttributes` **only** when a matching `justified_by` entry exists and `corevalidate` independently re-derives it — do not just delete `"restrict"` from the slice; split the check into "declaration site" (still banned) vs "manifest entry" (now conditionally allowed with justification).

### `Alias` field discharge (D-05-36) — `internal/compiler/cgen/cgen.go` + `internal/compiler/corevalidate/corevalidate.go`

**Analog:** `unsafeForeignContractField` (cgen.go:873-912) and `foreignContractFieldsCSafe` (corevalidate.go:1765-1781) — both are one-predicate-per-field audits; add `Alias` as a new field check alongside `Aliasing`'s siblings (`Fails`, `InitializedState`, `Capture`, `Retention`).

```go
// cgen.go:889-897 — the exact sibling shape to extend with Alias
if !commentSafeForeignField(contract.Fails) {
    return "fails"
}
if !commentSafeForeignField(contract.InitializedState) {
    return "initialized_state"
}
...
if !commentSafeForeignField(contract.Aliasing) {
    return "aliasing"
}
```
```go
// corevalidate.go:1765-1768 — the exact sibling shape to extend with Alias
func foreignContractFieldsCSafe(contract *core.ForeignContract) bool {
	if !commentSafe(contract.Fails) || !commentSafe(contract.InitializedState) || !commentSafe(contract.Capture) || !commentSafe(contract.Retention) || !commentSafe(contract.Aliasing) {
		return false
	}
```
The regression test D-05-36 requires ("no `EmitForeign*` output ever contains an `Alias` value") should mirror `TestAttributeInjectionMakesControlFail`-style mutation-kill tests already present in `cgen`/`corevalidate` test files (grep `unsafeForeignContractField` callers for the exact test naming convention used for `Aliasing`).

### `discoverLoanLastUses` retirement — shadow-mode migration (`internal/compiler/check/check.go`)

**Analog:** `testOnlyForceUniformLoanJoin` (check.go:14-30) — this file's own precedent for a fault-injection/shadow seam that never gates production behavior until proven safe, plus `loanLivenessFixpoint` (check.go:537-655, already used for branch arms only) and `discoverLoanLastUses` (check.go:1945-2005, called from `analyzeArmBody` line 722 and `analyzeStraightLine` line 1703).

```go
// check.go:14-30 — the exact "seam that never gates behaviour until proven" doc-comment shape to copy for the shadow pass
var testOnlyForceUniformLoanJoin = false
```
D-05-35's five-step procedure (broaden scope → shadow-log both laws over `TestOwnershipSequenceExhaustive`/`TestBranchSequenceExhaustive` → require zero divergence → delete `discoverLoanLastUses` → promote `core.LoanEndpoint`) should be implemented as: (a) widen `loanLivenessFixpoint`'s `blocks []cfgBlockSpec` domain to also cover straight-line bodies (currently only `checkBranch` calls it, per check.go:363); (b) add a divergence-logging call inside `analyzeStraightLine`/`analyzeArmBody` that runs both the old (`discoverLoanLastUses`) and new (`loanLivenessFixpoint`) laws and records (never gates on) any disagreement; (c) only after a documented zero-divergence run over the full enumeration, delete `discoverLoanLastUses` and repoint the two call sites (check.go:722, check.go:1703).

### New mutation runners — alias-fact injection, second-allocator TU, hostile UAF variant (`internal/compiler/session/session.go`)

**Analog:** `OwnedBackendMutationRunner` (session.go:67-119) is the closest full worked example for "corrupt one marked site, run through `native.Runner`, count work":

```go
// session.go:90-113 — the exact marker-count-guard + mutate + run + record shape every new NAT-03 runner copies
const mutationMarker = "/* lang:mutation-site */"

func (r *OwnedBackendMutationRunner) Run(ctx context.Context, cSource, optimization string, inputs []string) (native.Result, error) {
	lines := strings.Split(cSource, "\n")
	matched := -1
	for index, line := range lines {
		if strings.Contains(line, mutationMarker) {
			if matched != -1 {
				return native.Result{}, &native.ToolError{Code: "native.backend_control_invalid", Err: fmt.Errorf("owned transfer mutation marker count is >1, want 1")}
			}
			matched = index
		}
	}
	if matched == -1 {
		return native.Result{}, &native.ToolError{Code: "native.backend_control_invalid", Err: fmt.Errorf("owned transfer mutation marker count is 0, want 1")}
	}
	mutatedLine := lines[matched] + "\n  lang_value_delivered.bytes[0] ^= 0xffu; /* control: backend runtime causality */"
	lines[matched] = mutatedLine
	mutated := strings.Join(lines, "\n")
	r.mu.Lock()
	r.optimizations = append(r.optimizations, optimization)
	r.mu.Unlock()
	return r.runner.Run(ctx, mutated, optimization, inputs)
}
```

For the **second allocator TU** (D-05-08) and **hostile UAF variant** (D-05-09), the closer analog is `LayoutMutationRunner` (session.go:554-566, attacks a frozen foreign fixture rather than emitted C) — read it directly alongside `LayoutProbeContract()` (session.go:530-553) since both new fixtures are, like layout mismatch, attacks on a **frozen foreign TU**, resolved by declared symbol via `native.ForeignSourcePathForSymbol` (confirmed reused, not re-plumbed, per RESEARCH.md's Integration Points).

**Test pattern to copy** (fail-closed assertion shape): `TestReleaseOmissionMutationIsMismatch` (session_test.go:504-524):
```go
func TestReleaseOmissionMutationIsMismatch(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase4", "acquire_three_success.lang")
	inner := native.DefaultRunner()
	inner.ForeignSources = []string{native.ForeignResourceSourcePath()}
	runner := session.NewReleaseOmissionMutationRunner(inner)
	_, _, err := session.RunNativeFile(context.Background(), path, runner)
	if err == nil {
		t.Fatal("expected the release omission mutation to be detected, got no error")
	}
	var toolError *native.ToolError
	if !errors.As(err, &toolError) || toolError.Code != "native.invalid_execution" {
		t.Fatalf("expected native.invalid_execution (a live resource on a returned outcome), got %v", err)
	}
	if len(runner.Optimizations()) == 0 {
		t.Fatal("expected the omission runner to have recorded at least one optimization attempt")
	}
}
```
Every new sanitizer-detected mutation (alias-fact false-`restrict`, allocator-mismatch, UAF) should follow this exact "build via mutation runner → run → assert specific error/diagnostic code → assert nonzero recorded work" shape, but assert on the ASan/UBSan diagnostic-substring-plus-exit-code tuple (D-05-14/D-05-16) instead of `native.invalid_execution` where the sanitizer lane is the detector.

### `Phase5RequiredControls()` + script-parity tests (`internal/compiler/session/session.go`, `session_test.go`, `scripts/verify-phase5.sh`)

**Analog:** `Phase4RequiredControls()` (session.go:2004-2019):
```go
func Phase4RequiredControls() []string {
	return []string{
		"control:kind.exhaustive_dispatch",
		"control:foreign.call_target_not_foreign",
		"control:foreign.unwind_policy_undeclared",
		"control:resource.release_omitted",
		"control:resource.release_order_transposed",
		"control:foreign.layout_mismatch",
		"control:foreign.no_unproven_attributes",
		"control:defect.no_release_on_defect",
		"control:foreign.unwind_forbidden",
		"control:foreign.nonlocal_exit_undetected",
		"control:terminator.walk_incomplete",
		"control:origin.foreign_origin_omitted",
	}
}
```
Copy verbatim for `Phase5RequiredControls()`, populated with the Phase 5 IDs named across D-05-01..D-05-30 (e.g. `control:alias.false_no_alias`, `control:native.sanitize.retained_pointer`, `control:native.sanitize.ubsan_no_recover`, `control:native.allocator_mismatch`, `control:native.use_after_free`, `control:interpreter-o0-o3-lto`, plus whatever exact spellings are chosen at Claude's discretion — D-05-18's enumeration-bound constant should also live near this function per RESEARCH's Open Question 2).

**Test analog:** `TestPhase4VerifierScriptContract`/`TestPhase4RequiredControlsMatchScript`/`phase4VerifierScriptText`/`phase4OwnControlIdentifiers` (session_test.go:2109-2201) — copy the anchor-string-plus-`for control in`-block extraction technique exactly, retargeted at `scripts/verify-phase5.sh` and the anchor comment `"Phase 5's own required-control set"`.

### `scripts/verify-phase5.sh`

**Analog:** `scripts/verify-phase4.sh` (full file, 94 lines) — read directly. Key structural elements to peer (never fork):
```sh
#!/bin/sh
set -eu

verify_tmp=$(mktemp -d "${TMPDIR:-/tmp}/codename-lang-phase4.XXXXXX")
trap 'rm -rf "$verify_tmp"' EXIT HUP INT TERM
export GOCACHE="$verify_tmp/go-cache"

sh scripts/assert-go-tests.sh --self-test ./internal/compiler/session TestTogglePipeline TestOwnedBackendMutationIsMismatch TestVerifyPhase2ControlsAndWork TestVerifyPhase3ControlsAndWork TestVerifyPhase4ControlsAndWork
go test ./...
go test -race ./...
go vet ./...
go build -o "$verify_tmp/lang" ./cmd/lang
"$verify_tmp/lang" --json verify testdata/phase1 >"$verify_tmp/phase1.json"
...
```
```sh
# Non-regression for Phase 1, Phase 2, and Phase 3 is proven by running
# THEIR OWN corpora with THIS phase's freshly built binary, never by
# invoking an older gate script (D-04-21/T-04-42). The prior phase's own
# gate script stays byte-frozen and is neither forked, extended, nor
# invoked here -- this script is a peer of it, not an extension.
```
```sh
# Phase 4's own required-control set (D-04-22/T-04-41). Every identifier
for control in \
	control:kind.exhaustive_dispatch \
	control:foreign.call_target_not_foreign \
	...
do
	grep -q "$control" "$verify_tmp/phase4.json" || { echo "..." >&2; exit 1; }
done
```
`verify-phase5.sh` must (a) never invoke `verify-phase1.sh`/`verify-phase2.sh`/`verify-phase3.sh`/`verify-phase4.sh`, (b) re-run testdata/phase1-4 with the freshly built binary for non-regression, (c) run testdata/phase5, (d) `for control in ...` its own D-05-17/D-05-18 required set with the anchor comment `TestPhase5RequiredControlsMatchScript` greps for, and (e) also pin `ASAN_OPTIONS`/`UBSAN_OPTIONS` per D-05-13 when invoking the sanitizer lane's own verify path.

### `-O3 -flto` build configuration and separate sanitizer binary (`internal/compiler/native/native.go`)

**Analog for LTO tier:** `Runner.Run`'s compile invocation (native.go:88, native.go:117):
```go
// native.go:88 — foreign TU compile invocation
foreignCommand := r.commandContext(foreignCtx, r.ClangPath, "-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", optimization, "-c", foreignSource, "-o", objectPath)
```
```go
// native.go:117 — link invocation
arguments := append([]string{"-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", optimization, sourcePath}, objectPaths...)
```
Add `-flto` to both the per-TU compile and the final link argument lists, gated on a new `Runner` field (e.g. `LTO bool`) so `lane:native-differential-lto` opts in without touching the existing `-O0`/`-O3` differential's own invocation. Verified working on this host (RESEARCH.md Pattern 4): `clang -std=c17 -O3 -flto -c a.c -o a.o` then `clang -O3 -flto a.o main.o -o out`, exit 0.

**Analog for the sanitizer binary's structural isolation:** `Runner.CompileConformanceUnit` (native.go:216-237) — the existing precedent for "compile a wholly separate artifact from the differential's own build, never fed into the comparator":
```go
func (r Runner) CompileConformanceUnit(parent context.Context, source string) error {
```
Model the sanitizer build as a new `Runner` method returning its own result type (never `native.Result`, per D-05-11's type-level exclusion requirement), invoking:
```
clang -std=c17 -O1 -g -fno-omit-frame-pointer -fsanitize=address,undefined -fno-sanitize-recover=all <source> -o <sanitizer_binary>
ASAN_OPTIONS=halt_on_error=1:abort_on_error=1:symbolize=0:detect_leaks=0:detect_odr_violation=0:alloc_dealloc_mismatch=1 \
UBSAN_OPTIONS=halt_on_error=1:print_stacktrace=0 \
<sanitizer_binary>
```
(both command lines verified working on this host this research session — see RESEARCH.md Pattern 3).

**Analog for host-tool-absence posture:** `control:foreign.unwind_forbidden`'s `nm -u` probe (native.go — search `unwind_forbidden`/`nm` in this file) returns `tool_missing`/operational, never a silent pass; the sanitizer-availability smoke probe (D-05-15) must return the same posture on a missing/broken sanitizer runtime.

### Comparator axis widening + fail-closed exclusion routing (`internal/compiler/session/session.go`)

**Analog:** `Phase4CompareThreeEngines`/`firstExecutionDisagreement` (session.go:405-464):
```go
func Phase4CompareThreeEngines(fixture string, interpreted, o0, o3 execution.Execution) error {
	for _, comparison := range []struct {
		pair  string
		left  execution.Execution
		right execution.Execution
	}{
		{"interpreter-vs-O0", interpreted, o0},
		{"interpreter-vs-O3", interpreted, o3},
		{"O0-vs-O3", o0, o3},
	} {
		if detail := firstExecutionDisagreement(comparison.left, comparison.right); detail != "" {
			return &Phase4EngineDisagreement{Fixture: fixture, EnginePair: comparison.pair, Detail: detail}
		}
	}
	return nil
}
```
D-05-20/D-05-21 widen this exact function's compared axes (adding exit status/signal, diagnostic-ID equivalence for reject-programs) and D-05-21's fail-closed field-routing test should walk `execution.Execution`'s (and the event/ledger types') own struct fields via reflection, asserting every field is routed to either a compared-axis branch or an explicit exclusion list — mirroring `TestPhase4RequiredControlsMatchScript`'s set-equality-over-two-sources shape, but over struct fields instead of control-ID strings.

### QLT-01 control registry + `TestQLT01RegistryComplete`

**Analog 1 (fail-closed completeness scan):** `TestPhase4RequiredControlsMatchScript` (session_test.go:2177-2201) — same "walk every row/entry, fail if incomplete" shape, retargeted at registry rows instead of a control-ID set.

**Analog 2 (the escape/citation vocabulary to reuse verbatim):** `originvalidate`/`corevalidate`'s `KnownEscape` constant:
```go
// corevalidate.go:17-20
// KnownEscape names the exact boundary this validator cannot prove. A producer
// ...
const KnownEscape = "escape:coordinated-source-core-lie"
```
and session.go's existing escape-declaration-plus-never-detected pattern:
```go
// session.go — EscapeCoordinatedForeignBoundaryLie / EscapeNonlocalExitBelowThePad / EscapeForeignProcessExit
// wired into result.ExpectedEscapes (verifyForeignCorpus, session.go:2032) and
// asserted "declared but never appears as a detected lane control" by
// TestExpectedEscapesAreVisibleNotSolved (session_test.go:2209-2230)
```
The registry's `waived` rows should cite spike IDs and superseded-decision IDs the same way `.planning/spikes/MANIFEST.md`'s own table (read in full this session) already tags each spike's verdict (`VALIDATED`/`PARTIAL`) and requirement coverage — spikes 004 (`independent-certificate-checker`, PARTIAL) and 005 (`native-ffi-provenance-cleanup`, PARTIAL) are the two whose control mechanisms this phase's registry must extract per D-05-28/D-05-31.

### `lang.mismatch/0` schema type + round-trip test

**Analog:** `evidence.Manifest` struct shape (evidence.go:49-74) plus its golden round-trip test pattern (`evidence_test.go`):
```go
// evidence.go:49-74 — the struct-plus-json-tag shape to copy for the new schema
type Manifest struct {
	Schema           string   `json:"schema"`
	ID               string   `json:"id"`
	IDAlgorithm      string   `json:"id_algorithm"`
	...
}
```
```go
// evidence_test.go:104-106 — the golden round-trip assertion shape to copy
product := goldenProduct(t)
if got, want := product.ManifestBytes, readGolden(t, "evidence.golden.json"); !bytes.Equal(got, want) {
	t.Fatalf("Phase 1 evidence golden changed:\ngot  %s\nwant %s", got, want)
}
```
Define the new `lang.mismatch/0` type with exactly D-05-26's field list (`schema`, `diverging_axis`, `engine_pair`, `diverging_operation_id`, `reduced_core`, `reduced_source`, `minimality`, `total_recomputed_work`, `reduction_attempts`, `event_window`, `causal_chain`, `causes`), reusing the existing diagnostic cause-graph shape verbatim for `causes` (grep `diagnostic.Diagnostic`'s cause-graph field in `internal/compiler/diagnostic` for the exact type to embed).

## Shared Patterns

### Fail-closed marker-count guard (mutation runners)
**Source:** `internal/compiler/session/session.go:92-113` (`OwnedBackendMutationRunner.Run`), plus its four siblings at lines 149-179, 213-238, 258-281, 566-581.
**Apply to:** every new NAT-03 mutation runner (alias-fact false-`restrict` injection, second-allocator-TU mismatch, hostile UAF variant) — refuse to run on `>1` or `0` matching markers, mutate exactly one located site, delegate to the wrapped `native.Runner`, record nonzero `RecomputedWork`.

### Verbatim-duplicated required-control set with equality test
**Source:** `internal/compiler/session/session.go:2004-2019` (`Phase4RequiredControls`) + `internal/compiler/session/session_test.go:2109-2201` (`phase4VerifierScriptText`/`phase4OwnControlIdentifiers`/`TestPhase4RequiredControlsMatchScript`) + `scripts/verify-phase4.sh` lines 40-60.
**Apply to:** `Phase5RequiredControls()`, `scripts/verify-phase5.sh`'s own `for control in ...` block, `TestPhase5RequiredControlsMatchScript`, and — per D-05-18/D-05-29 — the corpus enumeration-bound constant and the QLT-01 registry's control-ID cross-check, all via the same "duplicate then assert set-equality" discipline.

### Declared-but-never-detected expected escape
**Source:** `internal/compiler/session/session.go` `EscapeCoordinatedForeignBoundaryLie`/`EscapeNonlocalExitBelowThePad`/`EscapeForeignProcessExit` constants, `verifyForeignCorpus`'s `result.ExpectedEscapes = []string{...}` wiring (session.go:2032), and `TestExpectedEscapesAreVisibleNotSolved` (session_test.go:2209-2230).
**Apply to:** `escape:callback-invocation-unsubjected` (D-05-07) and `escape:coordinated-source-to-core-false-claim` (D-05-30) — both must be declared in `ExpectedEscapes` and asserted to never appear as a detected lane control, with D-05-30 additionally requiring a real adversarial program proven to pass the gate under that named escape.

### Independent re-derivation, no shared helper (D-12)
**Source:** `internal/compiler/corevalidate/corevalidate.go`'s `validCIdentifier`/`foreignContractFieldsCSafe` (lines 1666-1781), which independently re-implements checks `cgen.go`'s `unsafeForeignContractField` (cgen.go:873-912) also implements, with zero shared code between the two files.
**Apply to:** the attribute-justification re-derivation (D-05-04) — `corevalidate` must re-derive the borrow-fact justification from the core artifact's own loan facts and must never read `EmittedAttributes` as written by `cgen`.

### Host-tool-absence returns `tool_missing`/operational, never a silent pass
**Source:** `control:foreign.unwind_forbidden`'s `nm -u` probe in `internal/compiler/native/native.go` (D-04-19).
**Apply to:** D-05-15's sanitizer-runtime-availability smoke probe.

## No Analog Found

| File | Role | Data Flow | Reason |
|---|---|---|---|
| New reducer package (5 fixed moves, `MaxReductionAttempts = 64`, HDD-style) | new subsystem | transform (fixpoint) | No delta-debugging or core-structure minimization code exists anywhere in this repository today. The closest available precedent is `session.go`'s `TransposeReleaseOrder` (a pure `core.Program` mutation function with no shared helper), useful only for the "operate on typed core, not source text" discipline — not for the reduction-loop shape itself. Build from D-05-23's five-move spec and RESEARCH.md's Standard Stack section directly. |
| QLT-01 registry file/table itself (as opposed to its audit test) | new registry (format at Claude's discretion) | batch | No machine-readable registry of any kind exists in this repository; `.planning/spikes/MANIFEST.md` is the closest prior-art *document* (prose+table) but is not itself machine-readable Go/JSON/YAML consumed by a test. Extraction is a genuinely new artifact per D-05-28. |

## Metadata

**Analog search scope:** `internal/compiler/{cgen,check,corevalidate,session,native,evidence,corevalidate,originvalidate}`, `scripts/`, `.planning/spikes/`
**Files scanned:** `cgen.go` (1649 lines), `check.go` (2107 lines), `corevalidate.go` (1802 lines), `session.go` (2586 lines), `session_test.go` (2200+ lines read in ranges), `native.go` (555 lines), `evidence.go` (487 lines) + `evidence_test.go` (ranges), `scripts/verify-phase4.sh` (94 lines), `.planning/spikes/MANIFEST.md`
**Pattern extraction date:** 2026-09-05
