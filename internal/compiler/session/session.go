package session

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/szTheory/schway/internal/compiler/callgraph"
	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/debugmap"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/evidence"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/pathoracle"
	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

type CheckResult struct {
	Tree        syntax.Tree
	Program     core.Program
	Diagnostics []diagnostic.Diagnostic
	Work        int
}

const MaxCheckWork = 100_000

type FormatResult struct {
	Source      []byte
	Canonical   []byte
	Diagnostics []diagnostic.Diagnostic
}

type NativeResult struct {
	CSource     string
	Interpreter []interp.Execution
	O0          native.Result
	O3          native.Result
}

type EngineMismatch struct {
	Optimization string
	Input        string
	Expected     string
	Actual       string
}

type NativeRunner interface {
	Run(context.Context, string, string, []string) (native.Result, error)
}

// OwnedBackendMutationRunner is a fail-closed verification seam around the
// production native runner. It corrupts exactly one generated owned-transfer
// value site before compiling the program, so a passing differential proves
// the execution document is causally derived from the C runtime state.
type OwnedBackendMutationRunner struct {
	runner        native.Runner
	mu            sync.Mutex
	optimizations []string
}

func NewOwnedBackendMutationRunner(runner native.Runner) *OwnedBackendMutationRunner {
	return &OwnedBackendMutationRunner{runner: runner}
}

// mutationMarker is the stable generated seam cgen emits at the single
// owned-transfer value site (D-02-07). Locating the mutation by this marker,
// rather than the previous exact source-derived line
// ("SCHWAY_BUFFER schway_value_delivered = ... op:0"), means renaming a fixture
// binding or reindenting the emitter cannot silently turn this control into
// an opaque operational failure — the seam survives both. The fail-closed
// exact-one requirement is unchanged: the marker must appear on exactly one
// line, or the control refuses to run rather than mutating an ambiguous or
// absent site.
const mutationMarker = "/* schway:mutation-site */"

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
	mutatedLine := lines[matched] + "\n  schway_value_delivered.bytes[0] ^= 0xffu; /* control: backend runtime causality */"
	lines[matched] = mutatedLine
	mutated := strings.Join(lines, "\n")
	r.mu.Lock()
	r.optimizations = append(r.optimizations, optimization)
	r.mu.Unlock()
	return r.runner.Run(ctx, mutated, optimization, inputs)
}

func (r *OwnedBackendMutationRunner) Optimizations() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.optimizations...)
}

// releaseMarker is the stable generated seam cgen emits at every OpRelease
// call site (D-04-07 Pitfall 2), distinct from mutationMarker above so the
// two mutation runners locate different seams and attack different
// artifacts: mutationMarker's runner corrupts a runtime VALUE the emitter
// wrote, while ReleaseOmissionMutationRunner below DELETES a whole generated
// line -- an entire release (its event AND its runtime ledger decrement,
// emitted on the same line so deleting one line removes both effects).
const releaseMarker = "/* schway:release-site */"

// legacyReleaseMarker names the seam in byte-frozen Phase 4/5 C artifacts.
// Those inputs remain part of hosted controls, so the mutation adapter accepts
// their historical label without changing their committed bytes.
const legacyReleaseMarker = "/* lang:release-site */"

// ReleaseOmissionMutationRunner is a fail-closed verification seam,
// structurally a sibling of OwnedBackendMutationRunner (its own mutex and
// optimization list, never shared) but attacking a DIFFERENT artifact: the
// emitter's own generated release call, not a mutation-site value. It
// requires the marker to appear on AT LEAST one line and deletes the FIRST
// one, refusing to run on a program with zero release sites (a
// control-invalid tool error, matching the exact-one-marker fail-closed
// shape) -- a program with no OpRelease at all is not a valid target for
// this control.
type ReleaseOmissionMutationRunner struct {
	runner        native.Runner
	mu            sync.Mutex
	optimizations []string
}

func NewReleaseOmissionMutationRunner(runner native.Runner) *ReleaseOmissionMutationRunner {
	return &ReleaseOmissionMutationRunner{runner: runner}
}

// Mutate performs ONLY the marker-scan-and-delete half of Run, with no
// execution -- the seam Phase 6's CleanupInjector calls directly (D-06-25:
// "reuse the existing release-omission mutation runner directly", not a
// reimplementation). Run below calls this same method, so there is exactly
// one release-marker scan in the whole tree, not two.
func (r *ReleaseOmissionMutationRunner) Mutate(cSource string) (string, error) {
	lines := strings.Split(cSource, "\n")
	matched := -1
	for index, line := range lines {
		if strings.Contains(line, releaseMarker) || strings.Contains(line, legacyReleaseMarker) {
			// The LAST marked line is always the success block's own final
			// release (this walker emits every err block inline, before the
			// success block it eventually falls through to), so it is the
			// one line every one of this plan's fixtures actually reaches
			// at runtime -- unlike an earlier err-block release, which a
			// program that never truly fails (this frozen TU's acquisition
			// only fails on real OOM) would never execute, making its
			// omission unobservable rather than a control-invalid target.
			matched = index
		}
	}
	if matched == -1 {
		return "", &native.ToolError{Code: "native.backend_control_invalid", Err: fmt.Errorf("release mutation marker count is 0, want at least 1")}
	}
	return strings.Join(append(append([]string(nil), lines[:matched]...), lines[matched+1:]...), "\n"), nil
}

func (r *ReleaseOmissionMutationRunner) Run(ctx context.Context, cSource, optimization string, inputs []string) (native.Result, error) {
	mutated, err := r.Mutate(cSource)
	if err != nil {
		return native.Result{}, err
	}
	r.mu.Lock()
	r.optimizations = append(r.optimizations, optimization)
	r.mu.Unlock()
	return r.runner.Run(ctx, mutated, optimization, inputs)
}

func (r *ReleaseOmissionMutationRunner) Optimizations() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.optimizations...)
}

// padInstallMarker/padEndMarker are duplicated, verbatim, from cgen.go's own
// constants of the same name (the same duplication pattern releaseMarker
// above already establishes for schway:release-site): they bracket the
// process-root nonlocal-exit landing pad's ENTIRE emitted span (D-04-17).
const padInstallMarker = "/* schway:nonlocal-pad-site */"
const padEndMarker = "/* schway:nonlocal-pad-end */"
const legacyPadInstallMarker = "/* lang:nonlocal-pad-site */"
const legacyPadEndMarker = "/* lang:nonlocal-pad-end */"

// ledgerPopulateMarker is duplicated, verbatim, from cgen.go's own constant
// of the same name: it marks the single generated line that flips one
// acquisition's ledger slot live (D-04-07/D-04-17).
const ledgerPopulateMarker = "/* schway:ledger-populate-site */"
const legacyLedgerPopulateMarker = "/* lang:ledger-populate-site */"

// NonlocalPadOmissionMutationRunner is control:foreign.nonlocal_exit_undetected's
// FIRST mutation-kill demonstration (D-04-21/D-10): it deletes the ENTIRE
// generated span from padInstallMarker through padEndMarker inclusive --
// not just the installation line, since a line-only deletion would leave an
// unmatched brace and fail to compile -- so the resulting program still
// declares schway_nonlocal_landing and still links against a foreign symbol
// that may longjmp into it, but the setjmp call that would have established
// the landing point never runs. A refusal to construct the mutation (the
// markers are absent) is itself a control-invalid tool error, matching the
// project's established fail-closed mutation-runner shape.
type NonlocalPadOmissionMutationRunner struct {
	runner        native.Runner
	mu            sync.Mutex
	optimizations []string
}

func NewNonlocalPadOmissionMutationRunner(runner native.Runner) *NonlocalPadOmissionMutationRunner {
	return &NonlocalPadOmissionMutationRunner{runner: runner}
}

func (r *NonlocalPadOmissionMutationRunner) Run(ctx context.Context, cSource, optimization string, inputs []string) (native.Result, error) {
	lines := strings.Split(cSource, "\n")
	start, end := -1, -1
	for index, line := range lines {
		if strings.Contains(line, padInstallMarker) || strings.Contains(line, legacyPadInstallMarker) {
			start = index
		}
		if strings.Contains(line, padEndMarker) || strings.Contains(line, legacyPadEndMarker) {
			end = index
		}
	}
	if start == -1 || end == -1 || end < start {
		return native.Result{}, &native.ToolError{Code: "native.backend_control_invalid", Err: fmt.Errorf("nonlocal pad marker span not found (start=%d end=%d)", start, end)}
	}
	mutated := strings.Join(append(append([]string(nil), lines[:start]...), lines[end+1:]...), "\n")
	r.mu.Lock()
	r.optimizations = append(r.optimizations, optimization)
	r.mu.Unlock()
	return r.runner.Run(ctx, mutated, optimization, inputs)
}

func (r *NonlocalPadOmissionMutationRunner) Optimizations() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.optimizations...)
}

// NonlocalLedgerOmissionMutationRunner is control:foreign.
// nonlocal_exit_undetected's SECOND, DIFFERENT mutation-kill demonstration
// (D-04-21/D-10): it deletes only the FIRST line bearing ledgerPopulateMarker
// -- the population site for the acquisition every fixture actually reaches
// at runtime, mirroring ReleaseOmissionMutationRunner's own "first
// reachable site" precedent -- so the pad's own leak count silently
// UNDERSTATES the true live set instead of the pad being skipped entirely
// (a different failure shape than NonlocalPadOmissionMutationRunner above).
type NonlocalLedgerOmissionMutationRunner struct {
	runner        native.Runner
	mu            sync.Mutex
	optimizations []string
}

func NewNonlocalLedgerOmissionMutationRunner(runner native.Runner) *NonlocalLedgerOmissionMutationRunner {
	return &NonlocalLedgerOmissionMutationRunner{runner: runner}
}

func (r *NonlocalLedgerOmissionMutationRunner) Run(ctx context.Context, cSource, optimization string, inputs []string) (native.Result, error) {
	lines := strings.Split(cSource, "\n")
	matched := -1
	for index, line := range lines {
		if strings.Contains(line, ledgerPopulateMarker) || strings.Contains(line, legacyLedgerPopulateMarker) {
			matched = index
			break
		}
	}
	if matched == -1 {
		return native.Result{}, &native.ToolError{Code: "native.backend_control_invalid", Err: fmt.Errorf("ledger populate marker count is 0, want at least 1")}
	}
	mutated := strings.Join(append(append([]string(nil), lines[:matched]...), lines[matched+1:]...), "\n")
	r.mu.Lock()
	r.optimizations = append(r.optimizations, optimization)
	r.mu.Unlock()
	return r.runner.Run(ctx, mutated, optimization, inputs)
}

func (r *NonlocalLedgerOmissionMutationRunner) Optimizations() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.optimizations...)
}

// countEventKind counts how many events of the given kind appear in events.
func countEventKind(events []execution.Event, kind string) int {
	count := 0
	for _, event := range events {
		if event.Kind == kind {
			count++
		}
	}
	return count
}

// TransposeReleaseOrder is the core-artifact mutation for
// control:resource.release_order_transposed: it exchanges the ReleasesOperationID
// of the first two OpRelease operations found in the given function's
// success block, attacking the CHECKER's materialized order (a different
// artifact than the emitter's output ReleaseOmissionMutationRunner attacks,
// and different again from the frozen foreign fixture a layout mutation
// attacks). Returns an error if the function has fewer than two releases in
// its success block -- a two-acquisition shape cannot even construct this
// mutation, which is exactly D-10's point.
func TransposeReleaseOrder(program core.Program) (core.Program, error) {
	encoded, err := json.Marshal(program)
	if err != nil {
		return core.Program{}, err
	}
	var mutated core.Program
	if err := json.Unmarshal(encoded, &mutated); err != nil {
		return core.Program{}, err
	}
	// Phase 11 (11-GUARD-LEDGER.md): KEPT. This mutation targets a
	// SPECIFIC single-function core artifact's success block by design --
	// widening it to N>1 would change what it attacks, not what it
	// accepts, so it stays scoped to the fixture it mutates.
	if len(mutated.Functions) != 1 {
		return core.Program{}, fmt.Errorf("release transposition expects one function")
	}
	function := &mutated.Functions[0]
	successBlockID := function.ID + ":block:success"
	var releaseIndices []int
	for _, block := range function.Linear.Blocks {
		if block.ID != successBlockID {
			continue
		}
		claimed := make(map[string]struct{}, len(block.OperationIDs))
		for _, opID := range block.OperationIDs {
			claimed[opID] = struct{}{}
		}
		for opIndex, operation := range function.Linear.Operations {
			if operation.Kind != core.OpRelease {
				continue
			}
			if _, inBlock := claimed[operation.ID]; inBlock {
				releaseIndices = append(releaseIndices, opIndex)
			}
		}
	}
	if len(releaseIndices) < 2 {
		return core.Program{}, fmt.Errorf("release transposition requires at least two releases in the success block, found %d", len(releaseIndices))
	}
	first, second := releaseIndices[0], releaseIndices[1]
	function.Linear.Operations[first].ReleasesOperationID, function.Linear.Operations[second].ReleasesOperationID =
		function.Linear.Operations[second].ReleasesOperationID, function.Linear.Operations[first].ReleasesOperationID
	function.Linear.Operations[first].SourceID, function.Linear.Operations[second].SourceID =
		function.Linear.Operations[second].SourceID, function.Linear.Operations[first].SourceID
	return mutated, nil
}

// Phase4EngineDisagreement is task 04-07-02's disagreement report (T-04-43):
// unlike EngineMismatch's coarse canonical-bytes comparison, it names the
// specific fixture, the specific engine pair, and the first differing
// event index or outcome field, so a Phase 4 three-engine disagreement is
// diagnosable without re-deriving the diff by hand.
type Phase4EngineDisagreement struct {
	Fixture    string
	EnginePair string
	Detail     string
}

func (d *Phase4EngineDisagreement) Error() string {
	return fmt.Sprintf("phase4 three-engine disagreement: fixture=%s pair=%s detail=%s", d.Fixture, d.EnginePair, d.Detail)
}

// firstExecutionDisagreement returns an empty string when left and right
// are identical (by the same canonical-bytes definition execution.Equal
// uses), or else names the first differing field: the outcome kind, the
// outcome value, the index and content of the first differing event (or a
// length mismatch), or the live-resource set.
func firstExecutionDisagreement(left, right execution.Execution) string {
	if execution.Equal(left, right) {
		return ""
	}
	if left.Outcome.Kind != right.Outcome.Kind {
		return fmt.Sprintf("outcome.kind: %q vs %q", left.Outcome.Kind, right.Outcome.Kind)
	}
	if left.Outcome.Value != right.Outcome.Value {
		return fmt.Sprintf("outcome.value: %q vs %q", left.Outcome.Value, right.Outcome.Value)
	}
	length := len(left.Events)
	if len(right.Events) > length {
		length = len(right.Events)
	}
	for index := 0; index < length; index++ {
		switch {
		case index >= len(left.Events):
			return fmt.Sprintf("events[%d]: missing vs %+v", index, right.Events[index])
		case index >= len(right.Events):
			return fmt.Sprintf("events[%d]: %+v vs missing", index, left.Events[index])
		case left.Events[index] != right.Events[index]:
			return fmt.Sprintf("events[%d]: %+v vs %+v", index, left.Events[index], right.Events[index])
		}
	}
	leftLive := append([]string(nil), left.LiveResources...)
	rightLive := append([]string(nil), right.LiveResources...)
	sort.Strings(leftLive)
	sort.Strings(rightLive)
	if !reflect.DeepEqual(leftLive, rightLive) {
		return fmt.Sprintf("live_resources: %+v vs %+v", leftLive, rightLive)
	}
	return "byte-identical JSON differs only in field ordering or an unrecognised field"
}

// phase4RunThreeEngines drives program's named function through the
// interpreter and both native optimization levels for a single input,
// auto-wiring the frozen foreign source for any declared ForeignContract
// symbol exactly as session.RunNative does, so a caller never has to
// remember which frozen TU a given symbol needs linked.
func phase4RunThreeEngines(ctx context.Context, fixture string, program core.Program, functionName, input string, runner native.Runner, expect native.TerminalOutcome) (interpreted, o0, o3 execution.Execution, err error) {
	interpreted, err = interp.Run(program, functionName, input)
	if err != nil {
		return execution.Execution{}, execution.Execution{}, execution.Execution{}, fmt.Errorf("interpreter run failed: %w", err)
	}
	cSource, cgenErr := Phase16ControlNativeC(program, "testdata/phase4/"+fixture)
	if cgenErr != nil {
		return execution.Execution{}, execution.Execution{}, execution.Execution{}, fmt.Errorf("cgen failed: %w", cgenErr)
	}
	nativeRunner := runner
	nativeRunner.Expect = expect
	for _, function := range program.Functions {
		if function.Name != functionName || function.ForeignContract == nil {
			continue
		}
		if sourcePaths := native.ForeignSourcePathsForSymbol(function.ForeignContract.Symbol); len(sourcePaths) != 0 {
			nativeRunner.ForeignSources = append(append([]string(nil), nativeRunner.ForeignSources...), sourcePaths...)
		}
	}
	o0Result, o0Err := nativeRunner.Run(ctx, cSource, "-O0", []string{input})
	if o0Err != nil || len(o0Result.Pairs) != 1 {
		return execution.Execution{}, execution.Execution{}, execution.Execution{}, fmt.Errorf("-O0 run failed: %w (pairs=%d)", o0Err, len(o0Result.Pairs))
	}
	o3Result, o3Err := nativeRunner.Run(ctx, cSource, "-O3", []string{input})
	if o3Err != nil || len(o3Result.Pairs) != 1 {
		return execution.Execution{}, execution.Execution{}, execution.Execution{}, fmt.Errorf("-O3 run failed: %w (pairs=%d)", o3Err, len(o3Result.Pairs))
	}
	interpreted, err = ProjectExecutionSchema2(program, interpreted)
	if err != nil {
		return execution.Execution{}, execution.Execution{}, execution.Execution{}, fmt.Errorf("project interpreter schema-2 evidence: %w", err)
	}
	o0, err = ProjectExecutionSchema2(program, o0Result.Pairs[0].Execution)
	if err != nil {
		return execution.Execution{}, execution.Execution{}, execution.Execution{}, fmt.Errorf("project -O0 schema-2 evidence: %w", err)
	}
	o3, err = ProjectExecutionSchema2(program, o3Result.Pairs[0].Execution)
	if err != nil {
		return execution.Execution{}, execution.Execution{}, execution.Execution{}, fmt.Errorf("project -O3 schema-2 evidence: %w", err)
	}
	return interpreted, o0, o3, nil
}

// Phase4CompareThreeEngines is the pure comparison half of task 04-07-02's
// differential: given three already-obtained execution documents, it
// compares all three pairwise -- interpreter-vs-O0, interpreter-vs-O3,
// O0-vs-O3 -- returning a *Phase4EngineDisagreement naming fixture, the
// specific engine pair, and the first differing field on any disagreement
// (T-04-43). Separated from Phase4ThreeEngineDifferential (which OBTAINS
// the three documents via real interp/native runs) so the comparison logic
// itself is directly testable against hand-constructed documents, without
// requiring a native toolchain invocation for every test case.
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

// ProjectExecutionSchema2 adapts the interpreter's retained historical
// execution document to the schema-2 boundary used by the whole-program C
// emitter.  It is deliberately a comparison-boundary projection: it does not
// participate in checking, emission, or execution.  Native program emission
// already writes schema-2 directly, while the interpreter keeps its older
// event identity for compatibility with pre-cut evidence.
func ProjectExecutionSchema2(program core.Program, document execution.Execution) (execution.Execution, error) {
	// Whole-program native emission already owns schema-2 causal identities.
	// Projecting it again would flatten every nested invocation to the entry
	// occurrence, turning valid call evidence into a peer refusal.
	if document.Schema == execution.Schema2 {
		return document, nil
	}
	entry, err := callgraph.EntryFunction(program)
	if err != nil {
		return execution.Execution{}, err
	}
	invocation, err := execution.FormatInvocation(entry.ID, nil)
	if err != nil {
		return execution.Execution{}, err
	}
	document.Schema = execution.Schema2
	defectReasons := make(map[string]string)
	for _, function := range program.Functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind == core.OpDefect && operation.Reason != "" {
				defectReasons[operation.ID+":event:defected"] = operation.Reason
			}
		}
	}
	// Only the legacy one-event switch document needs its historical match
	// return identity projected.  A match with arm operations (for example a
	// defect terminal) already has operation identities that correspond to
	// the program emitter and must not be collapsed into one synthetic event.
	bareMatch := entry.Match != nil && entry.Linear == nil && len(document.Events) == 1
	for index := range document.Events {
		document.Events[index].Schema = execution.Schema2
		document.Events[index].Invocation = invocation
		document.Events[index].Input = ""
		// Schema-2 deliberately omits legacy event decoration, except a defect
		// reason. That reason is a checked core-operation fact, not an
		// observation borrowed from native output.
		if document.Events[index].Kind != "function.defected" {
			document.Events[index].Output = ""
		} else {
			document.Events[index].Output = defectReasons[document.Events[index].ID]
		}
		if bareMatch {
			document.Events[index].ID = entry.ID + ":match:return"
			document.Events[index].SourcePlace = entry.Parameter.ID
			document.Events[index].TargetPlace = ""
			document.Events[index].TypeID = entry.Parameter.Type
		}
	}
	return document, nil
}

// Phase4ThreeEngineDifferential is task 04-07-02's own differential
// (ROADMAP SC4): it drives program's named function through the
// interpreter and both native optimization levels for input, then hands
// all three documents to Phase4CompareThreeEngines. A differential must
// never report only that a mismatch occurred.
func Phase4ThreeEngineDifferential(ctx context.Context, fixture string, program core.Program, functionName, input string, runner native.Runner, expect native.TerminalOutcome) (execution.Execution, error) {
	interpreted, o0, o3, err := phase4RunThreeEngines(ctx, fixture, program, functionName, input, runner, expect)
	if err != nil {
		return execution.Execution{}, fmt.Errorf("%s: %w", fixture, err)
	}
	if compareErr := Phase4CompareThreeEngines(fixture, interpreted, o0, o3); compareErr != nil {
		return execution.Execution{}, compareErr
	}
	return interpreted, nil
}

// Phase4CheckedProgram reads, checks, and independently validates a Phase 4
// corpus fixture, returning its single function's name alongside the
// validated core.Program -- the common prelude every Phase4ThreeEngineDifferential
// caller needs before it can drive the differential itself.
func Phase4CheckedProgram(corpus, fixture string) (core.Program, string, error) {
	source, err := readBoundedFile(filepath.Join(corpus, fixture), syntax.MaxSourceBytes)
	if err != nil {
		return core.Program{}, "", err
	}
	checked := Check(source)
	if len(checked.Diagnostics) != 0 {
		return core.Program{}, "", fmt.Errorf("%s: unexpected diagnostics: %+v", fixture, checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		return core.Program{}, "", fmt.Errorf("%s: core validation failed: %+v", fixture, validated.Problems)
	}
	program := validated.Program()
	// D-11-05: widened to permit N>1 functions -- callgraph.EntryFunction is
	// the single resolver of "which function IS this program", replacing
	// the old len(program.Functions) != 1 guard plus Functions[0] index.
	entry, err := callgraph.EntryFunction(program)
	if err != nil {
		return core.Program{}, "", fmt.Errorf("%s: %w", fixture, err)
	}
	return program, entry.Name, nil
}

// DefectHasNoReleaseAfter is control:defect.no_release_on_defect (D-04-18):
// zero resource.released events occur after a defect terminal record. It
// scans the WHOLE event list of a defect-outcome execution rather than
// special-casing any known-good shape, so a release emitted anywhere in a
// defect execution is caught regardless of position. A non-defect execution
// is vacuously true (nothing to check).
func DefectHasNoReleaseAfter(value execution.Execution) bool {
	if value.Outcome.Kind != "defect" {
		return true
	}
	for _, event := range value.Events {
		if event.Kind == "resource.released" {
			return false
		}
	}
	return true
}

// LayoutProbeContract is the Lang-side declared layout the layout mutation
// control (control:foreign.layout_mismatch, D-04-11/D-04-14) proves against
// a frozen private header fixture: two one-byte fields, `first` then
// `second`. testdata/phase4/foreign_layout_mismatch.golden.c deliberately
// transposes them, so compiling the generated conformance unit against that
// fixture must be refused under the project's existing -Werror flag set.
// This is a purpose-built probe record, independent of the production
// schway_res_open symbol's own (single-field) Layout, specifically so the
// mutation-kill demonstration exercises a genuine field TRANSPOSITION (which
// a one-field record cannot express) without touching the byte-frozen
// production fixture at all.
func LayoutProbeContract() *core.ForeignContract {
	return &core.ForeignContract{
		Symbol: "schway_layout_probe", Allocator: "libc_malloc", Unwind: "forbidden", NonlocalExit: "forbidden", Fails: "AcquireError",
		InitializedState: "fully", Capture: "none", Retention: "none", Aliasing: "none",
		Layout: &core.RecordLayout{
			Size: 2, Alignment: 1, ForeignTypeName: "schway_foreign_layout_probe_block",
			Fields: []core.LayoutField{
				{Name: "first", Size: 1, Alignment: 1, Offset: 0, CType: "unsigned char"},
				{Name: "second", Size: 1, Alignment: 1, Offset: 1, CType: "unsigned char"},
			},
		},
	}
}

// LayoutMutationRunner is control:foreign.layout_mismatch's mutation runner
// (D-04-11/Pitfall 2): it compiles a generated conformance unit against a
// FROZEN FIXTURE file path -- never a generated source -- and expects the
// compile to be refused. Per D-10 ("the two mutation directions attack
// different artifacts"), this attacks the frozen boundary fixture, a
// different artifact than ReleaseOmissionMutationRunner (the emitter's own
// output) or TransposeReleaseOrder (the checker's materialized order). It
// carries no field of a generated-source shape at all -- its only per-run
// input is FixturePath -- so "never opens a generated source" is a
// structural property of this type, not merely a runtime behavior.
type LayoutMutationRunner struct {
	Runner      native.Runner
	Contract    *core.ForeignContract
	FixturePath string
}

// Run assembles a minimal single-function core.Program carrying only
// r.Contract, generates its conformance unit against r.FixturePath, and
// compiles it as its own separate, bounded invocation. A nil error means the
// fixture at FixturePath conforms to r.Contract's declared layout; a
// *native.ToolError with code "native.conformance_failed" means it was
// refused at compile time.
func (r LayoutMutationRunner) Run(ctx context.Context) error {
	program := core.Program{
		Schema: core.Schema1, Module: "phase4.layout_probe", ModuleID: "phase4.layout_probe",
		Functions: []core.Function{{
			ID: "phase4.layout_probe:fn:probe", Name: "probe",
			EntryPointID: "phase4.layout_probe:fn:probe:point:entry", ReturnPointID: "phase4.layout_probe:fn:probe:point:return",
			ForeignContract: r.Contract,
		}},
	}
	source, err := cgen.EmitForeignConformance(program, r.FixturePath)
	if err != nil {
		return err
	}
	return r.Runner.CompileConformanceUnit(ctx, source)
}

// PayloadProbeDataType is control:payload.layout_mismatch's own purpose-built
// core.DataType (D-12-37), independent of testdata/phase12/payload_tracer.schway's
// production Outcome/Fault shape, specifically so the mutation-kill
// demonstration exercises a genuine field TRANSPOSITION with plain
// single-byte fields (mirroring LayoutProbeContract's own two-one-byte-field
// design) rather than pulling in SCHWAY_BUFFER's own real (padded, platform-
// dependent) struct layout, which check.PayloadRecordLayout's own
// payloadFieldShape declares as size 8/alignment 1 -- a deliberately
// informational approximation this control has no need to depend on.
func PayloadProbeDataType() core.DataType {
	dataType, err := core.NewDataType("phase12.payload_layout_probe:type:PayloadProbe", "PayloadProbe",
		[]string{"First", "Second"},
		[]core.AlternativeDetail{
			{Name: "First", PayloadType: "Byte"},
			{Name: "Second", PayloadType: "Byte"},
		}, diagnostic.Span{})
	if err != nil {
		panic(fmt.Sprintf("PayloadProbeDataType: %v", err))
	}
	return dataType
}

// PayloadLayoutMutationRunner is D-12-37's payload-side sibling of
// LayoutMutationRunner (control:foreign.layout_mismatch's precedent): it
// compiles a generated payload conformance unit against a FROZEN FIXTURE
// file path -- never a generated source -- and expects the compile to be
// refused. Per D-10 ("the two mutation directions attack different
// artifacts"), this attacks the frozen boundary fixture declaring the
// payload struct's alternative slots transposed or resized relative to what
// check.PayloadRecordLayout(DataType) derives -- a different artifact than
// LayoutMutationRunner's own foreign-contract-keyed fixture. It carries no
// field of a generated-source shape at all -- its only per-run inputs are
// the checker-derived DataType and FixturePath -- so "never opens a
// generated source" is a structural property of this type, not merely a
// runtime behavior.
//
// This control is NECESSARY for the C-side layout obligation and
// STRUCTURALLY INCAPABLE of catching the real bug a struct-shaped
// declaration cannot express: a _Static_assert polices only the struct
// DECLARATION's sizeof/_Alignof/offsetof, never WHICH FIELD a given match
// arm actually reads for a correct tag. D-12-38's decisive control --
// TestPayloadSlotSwapMutationKilled, a reverted production-hunk slot-swap
// mutation proving a genuine interpreter-vs-native value divergence -- is
// what covers that gap. Do not mistake this control for sufficient on its
// own.
type PayloadLayoutMutationRunner struct {
	Runner      native.Runner
	DataType    core.DataType
	FixturePath string
}

// Run assembles the payload conformance unit for r.DataType against
// r.FixturePath and compiles it as its own separate, bounded invocation. A
// nil error means the fixture at FixturePath conforms to
// check.PayloadRecordLayout(r.DataType)'s declared layout; a
// *native.ToolError with code "native.conformance_failed" means it was
// refused at compile time.
func (r PayloadLayoutMutationRunner) Run(ctx context.Context) error {
	source, err := cgen.EmitPayloadConformance(r.DataType, r.FixturePath)
	if err != nil {
		return err
	}
	return r.Runner.CompileConformanceUnit(ctx, source)
}

func (e *EngineMismatch) Error() string {
	return fmt.Sprintf("native %s mismatch for %s: expected %s, got %s", e.Optimization, e.Input, e.Expected, e.Actual)
}

func Check(source []byte) CheckResult {
	parsed := syntax.Parse(source)
	// Work is one unit per bounded lexer token plus the checker's explicit
	// ownership/type traversal count. It is deterministic and source-scaled.
	result := CheckResult{Tree: parsed.Tree, Diagnostics: append([]diagnostic.Diagnostic(nil), parsed.Diagnostics...), Work: len(parsed.Tree.Tokens)}
	if len(result.Diagnostics) > 0 {
		return result
	}
	checked := check.Program(parsed.Program)
	result.Work += checked.Work
	if result.Work > MaxCheckWork {
		result.Diagnostics = append(result.Diagnostics, diagnostic.Error("check.work_limit", diagnostic.Span{}, "checking exceeds the declared work limit"))
		return result
	}
	result.Program = checked.Program
	result.Diagnostics = append(result.Diagnostics, checked.Diagnostics...)
	return result
}

func Format(source []byte) FormatResult {
	parsed := syntax.Parse(source)
	return FormatResult{Source: append([]byte(nil), source...), Canonical: syntax.Format(parsed.Tree), Diagnostics: parsed.Diagnostics}
}

func FormatFile(path string) (FormatResult, error) {
	source, err := readBoundedFile(path, syntax.MaxSourceBytes)
	if err != nil {
		return FormatResult{}, err
	}
	return Format(source), nil
}

func FormatCommandFile(path string, checkOnly bool) (protocol.Result, error) {
	started := time.Now()
	formatted, err := FormatFile(path)
	if err != nil {
		return protocol.Result{}, err
	}
	result := protocol.New("format", protocol.StatusPass)
	if len(formatted.Diagnostics) > 0 {
		result.Status = protocol.StatusInvalid
		result.Diagnostics = formatted.Diagnostics
		return completeCommand(result, started, 1), nil
	}
	if checkOnly && !bytes.Equal(formatted.Source, formatted.Canonical) {
		result.Status = protocol.StatusInvalid
		result.Diagnostics = []diagnostic.Diagnostic{diagnostic.Error("format.non_canonical", diagnostic.Span{}, "source differs from canonical projection")}
		return completeCommand(result, started, 1), nil
	}
	if !checkOnly {
		result.Formatted = string(formatted.Canonical)
	}
	return completeCommand(result, started, 1), nil
}

func CheckFile(path string) (CheckResult, error) {
	source, err := readBoundedFile(path, syntax.MaxSourceBytes)
	if err != nil {
		return CheckResult{}, err
	}
	return Check(source), nil
}

// peerRefusalUnnamedCode is 07-10's fail-closed fallback code (D-10 union
// rule edge 2): corevalidate.Validate reporting !Valid with an EMPTY
// Problems slice is structurally impossible today, but absence of a named
// problem must never be read as absence of a refusal. Every command path
// that consults the peer falls back to this constant, never to
// protocol.StatusPass and never to an empty diagnostic code, when that
// impossible shape is ever produced.
const peerRefusalUnnamedCode = "core.peer_refusal_unnamed"

// peerRefusalDiagnostic formats corevalidate.Validate's own refusal into a
// single diagnostic.Diagnostic. It is called from three sites
// (CheckCommandFile, InterfaceExportCommandFile, InterfaceCoreCommandFile)
// that each report the SAME already-computed peer verdict -- this is
// shared FORMATTING of one derivation's own answer, never a reconciliation
// of two independent derivations (07-10 checkpoint prohibition): it reads
// nothing from check, contributes no predicate of its own, and callers
// still decide independently whether to call it at all.
func peerRefusalDiagnostic(validated corevalidate.Result, moduleID string) diagnostic.Diagnostic {
	code := peerRefusalUnnamedCode
	detail := fmt.Sprintf("corevalidate refused module %s", moduleID)
	if len(validated.Problems) > 0 {
		code = validated.Problems[0].Code
		detail = validated.Problems[0].Detail
	}
	return diagnostic.Error(code, diagnostic.Span{}, detail)
}

func originPeerRefusalDiagnostic(problem originvalidate.Problem, program core.Program, tree syntax.Tree) diagnostic.Diagnostic {
	fallback := func() diagnostic.Diagnostic {
		return diagnostic.Error(problem.Code, diagnostic.Span{}, problem.Detail)
	}
	if problem.FunctionID == "" || problem.ReturnOperationID == "" {
		return fallback()
	}
	var function *core.Function
	for index := range program.Functions {
		if program.Functions[index].ID == problem.FunctionID {
			function = &program.Functions[index]
			break
		}
	}
	if function == nil || function.Linear == nil {
		return fallback()
	}
	var returned *core.LinearOperation
	for index := range function.Linear.Operations {
		if function.Linear.Operations[index].ID == problem.ReturnOperationID && function.Linear.Operations[index].Kind == core.OpReturn {
			returned = &function.Linear.Operations[index]
			break
		}
	}
	if returned == nil {
		return fallback()
	}
	places := make(map[string]core.Place, len(function.Linear.Places))
	for _, place := range function.Linear.Places {
		places[place.ID] = place
	}
	resultPlace, exists := places[returned.SourceID]
	if !exists || resultPlace.Name == "" {
		return fallback()
	}
	start, end, exists := sourceFunctionTokenBounds(tree.Tokens, function.Name)
	if !exists {
		return fallback()
	}
	primary, exists := lastIdentifierSpan(tree.Tokens, start, end, resultPlace.Name)
	if !exists {
		return fallback()
	}
	definitions := make(map[string]core.LinearOperation, len(function.Linear.Operations))
	for _, operation := range function.Linear.Operations {
		if operation.TargetID != "" {
			definitions[operation.TargetID] = operation
		}
	}
	current := returned.SourceID
	visited := map[string]bool{}
	var borrow *core.LinearOperation
	for current != function.Parameter.ID && !visited[current] {
		visited[current] = true
		operation, ok := definitions[current]
		if !ok {
			break
		}
		if operation.Kind == core.OpBorrowShared || operation.Kind == core.OpBorrowExclusive {
			borrow = &operation
			break
		}
		if operation.Kind == core.OpCopy {
			break
		}
		current = operation.SourceID
	}
	if borrow == nil {
		return fallback()
	}
	borrowedPlace, hasBorrowedPlace := places[borrow.TargetID]
	borrowSource, hasBorrowSource := places[borrow.SourceID]
	if !hasBorrowedPlace || !hasBorrowSource || borrowedPlace.Name == "" || borrowSource.Name == "" {
		return fallback()
	}
	cause, exists := borrowExpressionSpan(tree.Tokens, start, end, borrowedPlace.Name, borrowSource.Name)
	if !exists {
		return fallback()
	}
	return diagnostic.Error(problem.Code, primary, problem.Detail, diagnostic.Cause{
		Kind:   "borrow_created_here",
		Detail: "borrow that created the escaped result",
		Span:   &cause,
	})
}

func sourceFunctionTokenBounds(tokens []syntax.Token, functionName string) (int, int, bool) {
	for index, token := range tokens {
		if token.Kind != syntax.TokenFn {
			continue
		}
		nameIndex := nextSignificantToken(tokens, index+1)
		if nameIndex < 0 || tokens[nameIndex].Kind != syntax.TokenIdentifier || tokens[nameIndex].Text != functionName {
			continue
		}
		parameterList := nextSignificantToken(tokens, nameIndex+1)
		if parameterList < 0 || tokens[parameterList].Kind != syntax.TokenLParen {
			continue
		}
		open := -1
		for cursor := nameIndex + 1; cursor < len(tokens); cursor++ {
			if tokens[cursor].Kind == syntax.TokenLBrace {
				open = cursor
				break
			}
		}
		if open < 0 {
			return 0, 0, false
		}
		depth := 1
		for cursor := open + 1; cursor < len(tokens); cursor++ {
			switch tokens[cursor].Kind {
			case syntax.TokenLBrace:
				depth++
			case syntax.TokenRBrace:
				depth--
				if depth == 0 {
					return open + 1, cursor, true
				}
			}
		}
		return 0, 0, false
	}
	return 0, 0, false
}

func nextSignificantToken(tokens []syntax.Token, start int) int {
	for index := start; index < len(tokens); index++ {
		if !tokens[index].Trivia() && tokens[index].Kind != syntax.TokenEOF {
			return index
		}
	}
	return -1
}

func lastIdentifierSpan(tokens []syntax.Token, start, end int, name string) (diagnostic.Span, bool) {
	var span diagnostic.Span
	found := false
	for index := start; index < end; index++ {
		if tokens[index].Kind == syntax.TokenIdentifier && tokens[index].Text == name {
			span = tokens[index].Span
			found = true
		}
	}
	return span, found
}

func borrowExpressionSpan(tokens []syntax.Token, start, end int, targetName, sourceName string) (diagnostic.Span, bool) {
	for index := start; index < end; index++ {
		if tokens[index].Kind != syntax.TokenLet {
			continue
		}
		target := nextSignificantToken(tokens, index+1)
		if target < 0 || tokens[target].Kind != syntax.TokenIdentifier || tokens[target].Text != targetName {
			continue
		}
		equal := nextSignificantToken(tokens, target+1)
		if equal < 0 {
			continue
		}
		borrow := nextSignificantToken(tokens, equal+1)
		if equal < 0 || tokens[equal].Kind != syntax.TokenEqual || borrow < 0 || tokens[borrow].Kind != syntax.TokenBorrow {
			continue
		}
		source := nextSignificantToken(tokens, borrow+1)
		if source >= 0 && tokens[source].Kind == syntax.TokenMut {
			source = nextSignificantToken(tokens, source+1)
		}
		if source >= 0 && source < end && tokens[source].Kind == syntax.TokenIdentifier && tokens[source].Text == sourceName {
			return diagnostic.Span{Start: tokens[borrow].Span.Start, End: tokens[source].Span.End}, true
		}
	}
	return diagnostic.Span{}, false
}

// checkCommandPeerSeam is 07-10 Task 1's unexported fault-injection seam
// for control:check.peer_consulted: when true, CheckCommandFile skips the
// corevalidate.Validate consult entirely, restoring the pre-07-10 CR-04
// defect (07-REVIEW.md) where a program the independent peer refused could
// still report status: pass, exit 0 from the gate users actually run.
// Reachable only from same-package tests via
// session_phase7_export_test.go's SetCheckCommandPeerSeam; no exported
// session symbol reaches it on a production path.
var checkCommandPeerSeam = false

// checkCommandPeerObservedForTest records which independent command peers
// CheckCommandFile actually consults. Production leaves it nil; same-package
// tests use it to pin the admission order without merging peer predicates.
var checkCommandPeerObservedForTest func(string)

func observeCheckCommandPeer(name string) {
	if checkCommandPeerObservedForTest != nil {
		checkCommandPeerObservedForTest(name)
	}
}

// interfacePeerRefusalSeam is 07-10 Task 2's unexported fault-injection
// seam for control:interface.peer_refusal_is_invalid: when true, both
// InterfaceExportCommandFile and InterfaceCoreCommandFile restore the
// pre-07-10 behaviour of returning a non-nil error that discards the
// peer's own code, instead of a protocol.StatusInvalid result carrying it
// (07-REVIEW.md CR-04's `interface` half). One seam kills the control on
// both paths -- they are one fact, asserted twice. Reachable only from
// same-package tests via session_phase7_export_test.go's
// SetInterfacePeerRefusalSeamForTest; no exported session symbol reaches
// it on a production path.
var interfacePeerRefusalSeam = false

// CheckCommandFile reports the REFUSING UNION of two independently derived
// admission layers, in fixed precedence order (07-10 checkpoint, SEM-04):
// (1) check's own diagnostics, (2) corevalidate.Validate's independent
// replay over the emitted core.Program alone, then (3)
// originvalidate.ValidatePublished on the peer-normalized program -- the
// exact order InterfaceExportCommandFile already uses. Neither layer's
// verdict suppresses, reconciles, or overrides the other; this function
// contributes no predicate of its own about ownership, types, or
// callability (07-REVIEW.md CR-04 / PVG-03).
func CheckCommandFile(path string) (protocol.Result, error) {
	started := time.Now()
	checked, err := CheckFile(path)
	if err != nil {
		return protocol.Result{}, err
	}
	result := protocol.New("check", protocol.StatusPass)
	if len(checked.Diagnostics) > 0 {
		result.Status = protocol.StatusInvalid
		result.Diagnostics = checked.Diagnostics
		return completeCommand(result, started, checked.Work), nil
	}
	if !checkCommandPeerSeam {
		observeCheckCommandPeer("corevalidate")
		validated := corevalidate.Validate(checked.Program)
		if !validated.Valid {
			result.Status = protocol.StatusInvalid
			result.Diagnostics = []diagnostic.Diagnostic{peerRefusalDiagnostic(validated, checked.Program.ModuleID)}
			return completeCommand(result, started, checked.Work), nil
		}
		checked.Program = validated.Program()
	}
	observeCheckCommandPeer("originvalidate")
	if problems := originvalidate.ValidatePublished(checked.Program); len(problems) > 0 {
		// D-04-27/WR-01: originvalidate.ValidatePublished no longer runs only
		// on the `interface export` path -- the foreign declaration surface
		// is a second place an alias fact can be silently absent (D-04-28),
		// so `schway check` must independently recompute every published
		// origin too, not merely trust what the checker declared.
		result.Status = protocol.StatusInvalid
		result.Diagnostics = []diagnostic.Diagnostic{originPeerRefusalDiagnostic(problems[0], checked.Program, checked.Tree)}
	} else {
		observeCheckCommandPeer("pathoracle")
		if err := pathoracle.ValidateLocalOwnerPaths(checked.Program); err != nil {
			result.Status = protocol.StatusInvalid
			result.Diagnostics = []diagnostic.Diagnostic{diagnostic.Error("core.pathoracle_refused", diagnostic.Span{}, err.Error())}
		} else {
			result.ModuleID = checked.Program.ModuleID
		}
	}
	return completeCommand(result, started, checked.Work), nil
}

func RunInterpreter(source []byte) ([]interp.Execution, []diagnostic.Diagnostic, error) {
	checked := Check(source)
	if len(checked.Diagnostics) > 0 {
		return nil, checked.Diagnostics, nil
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		return nil, nil, fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
	}
	checked.Program = validated.Program()
	// D-11-05: widened to permit N>1 functions -- callgraph.EntryFunction
	// replaces the old len(...Functions) != 1 guard, requiring a uniquely
	// resolvable entry rather than exactly one declared function.
	entry, err := callgraph.EntryFunction(checked.Program)
	if err != nil {
		return nil, nil, err
	}
	inputs, ok := interpreterInputs(checked.Program)
	if !ok {
		return nil, nil, os.ErrInvalid
	}
	executions := make([]interp.Execution, 0, len(inputs))
	for _, input := range inputs {
		execution, err := interp.Run(checked.Program, entry.Name, input)
		if err != nil {
			return nil, nil, err
		}
		executions = append(executions, execution)
	}
	return executions, nil, nil
}

// interpreterInputs derives the single-Byte/Buffer synthetic input this
// project's differential lanes drive every fixture with, reading the
// PARAMETER TYPE from the program's own resolved entry function
// (callgraph.EntryFunction, D-11-05) rather than indexing
// program.Functions[0] -- the TestSessionRunSitesDoNotIndexFunctionsZero
// invariant this phase adopts. The single-Byte/Buffer synthesis logic
// itself is unchanged; only which function it reads the parameter from
// changes. Kept as a same-signature helper (not widened to accept the
// caller's own already-resolved entry) so its many existing single-function
// callers across this package are unaffected -- EntryFunction resolves
// trivially to that sole function for every one of them.
func interpreterInputs(program core.Program) ([]string, bool) {
	function, err := callgraph.EntryFunction(program)
	if err != nil {
		return nil, false
	}
	if function.Linear != nil && function.Match == nil {
		switch function.Parameter.Type {
		case "Byte":
			return []string{"7"}, true
		case "Buffer":
			return []string{"01020304"}, true
		default:
			return nil, false
		}
	}
	if function.Match != nil {
		// Covers both the Phase 1 bare-arm match (Linear == nil) and the
		// Phase 3 branch-shaped match whose arms carry linear bodies
		// (Linear != nil) — both dispatch on every alternative of the
		// scrutinee's declared type. Phase 12 (D-12-05): a payload-carrying
		// data type's own alternative may itself reference ANOTHER declared
		// data type as its payload (e.g. Err(Fault)), so program.DataTypes
		// can legitimately hold more than one entry -- the scrutinee's own
		// type is resolved BY NAME against function.Parameter.Type, never
		// by assuming it is the program's only declared type.
		for _, dataType := range program.DataTypes {
			if dataType.Name == function.Parameter.Type {
				return append([]string(nil), dataType.Alternatives...), true
			}
		}
	}
	return nil, false
}

func RunInterpreterFile(path string) ([]interp.Execution, []diagnostic.Diagnostic, error) {
	source, err := readBoundedFile(path, syntax.MaxSourceBytes)
	if err != nil {
		return nil, nil, err
	}
	return RunInterpreter(source)
}

// publishedOriginProblemFile is D-04-27/WR-01's `schway check`/`schway run`
// wiring point: it independently recomputes every function's published
// origin from the typed core alone (originvalidate.ValidatePublished),
// exactly as `interface export` already does, returning the first problem
// as a diagnostic. This is deliberately a SEPARATE, narrow re-check at the
// CLI command-file layer rather than folded into RunInterpreter/RunNative
// themselves: those two functions are also the machinery every non-CLI test
// and verify-corpus lane in this repository drives directly, several of
// which intentionally exercise an undeclared-borrow-derived-return fixture
// at the run/interpret layer (e.g. TestExclusiveBorrowInterpreterNative) --
// only PUBLICATION (export, and now check/run's own command surface) was
// ever meant to gate on a declared origin (D-03-02/WR-01), not every
// internal call to the interpreter or native engine.
func publishedOriginProblemFile(path string) []diagnostic.Diagnostic {
	source, err := readBoundedFile(path, syntax.MaxSourceBytes)
	if err != nil {
		return nil
	}
	checked := Check(source)
	if len(checked.Diagnostics) > 0 {
		return nil
	}
	problems := originvalidate.ValidatePublished(checked.Program)
	if len(problems) == 0 {
		return nil
	}
	return []diagnostic.Diagnostic{diagnostic.Error(problems[0].Code, diagnostic.Span{}, problems[0].Detail)}
}

func RunInterpreterCommandFile(path string) (protocol.Result, error) {
	started := time.Now()
	executions, diagnostics, err := RunInterpreterFile(path)
	if err != nil {
		result := protocol.New("run", protocol.StatusOperational)
		result.Diagnostics = []diagnostic.Diagnostic{diagnostic.Error("tool.run_failed", diagnostic.Span{}, "interpreter operation failed")}
		return completeCommand(result, started, 1), nil
	}
	if len(diagnostics) == 0 {
		if originProblems := publishedOriginProblemFile(path); len(originProblems) > 0 {
			diagnostics = originProblems
		}
	}
	result := protocol.New("run", protocol.StatusPass)
	if len(diagnostics) > 0 {
		result.Status = protocol.StatusInvalid
		result.Diagnostics = diagnostics
	} else {
		result.Executions = executions
	}
	return completeCommand(result, started, len(executions)), nil
}

func runNative(ctx context.Context, source []byte, runner NativeRunner, fixture string) (NativeResult, []diagnostic.Diagnostic, error) {
	checked := Check(source)
	if len(checked.Diagnostics) > 0 {
		return NativeResult{}, checked.Diagnostics, nil
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		return NativeResult{}, nil, fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
	}
	checked.Program = validated.Program()
	// D-11-05: widened to permit N>1 functions -- callgraph.EntryFunction
	// replaces the old len(...Functions) != 1 guard.
	entry, err := callgraph.EntryFunction(checked.Program)
	if err != nil {
		return NativeResult{}, nil, err
	}
	inputs, ok := interpreterInputs(checked.Program)
	if !ok {
		return NativeResult{}, nil, os.ErrInvalid
	}
	interpreted := make([]interp.Execution, 0, len(inputs))
	for _, input := range inputs {
		execution, err := interp.Run(checked.Program, entry.Name, input)
		if err != nil {
			return NativeResult{}, nil, err
		}
		interpreted = append(interpreted, execution)
	}
	cSource, err := Phase16ControlNativeC(checked.Program, fixture)
	if err != nil {
		return NativeResult{}, nil, err
	}
	// A foreign-shaped function (D-04-10) needs the frozen foreign
	// translation unit linked in. This is additive wiring on a *native.Runner
	// value specifically -- a caller-supplied NativeRunner of any other
	// concrete type (e.g. a mutation runner) is passed through unmodified,
	// since none of those exercise a foreign-call program today.
	if contract := entry.ForeignContract; contract != nil {
		if concrete, ok := runner.(native.Runner); ok {
			// Resolve by the function's own symbol so the implementation and
			// any historical fixture adapter required by that symbol are linked.
			if sourcePaths := native.ForeignSourcePathsForSymbol(contract.Symbol); len(sourcePaths) != 0 {
				concrete.ForeignSources = append(append([]string(nil), concrete.ForeignSources...), sourcePaths...)
				runner = concrete
			}
		}
	}
	o0, err := runNativeInputs(ctx, runner, cSource, "-O0", inputs, interpreted)
	if err != nil {
		return NativeResult{}, nil, err
	}
	o3, err := runNativeInputs(ctx, runner, cSource, "-O3", inputs, interpreted)
	if err != nil {
		return NativeResult{}, nil, err
	}
	if len(o0.Pairs) != len(inputs) || len(o3.Pairs) != len(inputs) {
		return NativeResult{}, nil, &EngineMismatch{
			Optimization: "-O0/-O3", Expected: fmt.Sprintf("%d executions", len(inputs)),
			Actual: fmt.Sprintf("%d/%d executions", len(o0.Pairs), len(o3.Pairs)),
		}
	}
	for index, expected := range interpreted {
		// The sole post-cut public emitter serializes schema-2 documents. The
		// interpreter remains the independent semantic oracle, so project its
		// legacy representation at this comparison boundary and validate the
		// resulting /2 document with the independent peer before comparing it
		// to native output. Cut-M004 programs never reach this point: cgen
		// refuses them before native compilation.
		expectedDocument, projectionErr := ProjectExecutionSchema2(checked.Program, expected)
		if projectionErr != nil {
			return NativeResult{}, nil, fmt.Errorf("project schema-2 interpreter evidence: %w", projectionErr)
		}
		interpreted[index] = expectedDocument
		for _, result := range []*native.Result{&o0, &o3} {
			projected, projectionErr := ProjectExecutionSchema2(checked.Program, result.Pairs[index].Execution)
			if projectionErr != nil {
				return NativeResult{}, nil, fmt.Errorf("project %s schema-2 native evidence: %w", result.Optimization, projectionErr)
			}
			result.Pairs[index].Execution = projected
		}
		// The projection derives its invocation and defect facts from checked
		// program data, so native output can no longer repair the expected
		// document by carrying legacy schema decoration.
		if err := Phase5CompareProgramEngines("run-native", checked.Program, map[string]execution.Execution{"O0": o0.Pairs[index].Execution, "O3": o3.Pairs[index].Execution}); err != nil {
			return NativeResult{}, nil, err
		}
		for _, actual := range []native.Result{o0, o3} {
			if !execution.Equal(expectedDocument, actual.Pairs[index].Execution) {
				expectedBytes, _ := execution.CanonicalBytes(expectedDocument)
				actualBytes, _ := execution.CanonicalBytes(actual.Pairs[index].Execution)
				return NativeResult{}, nil, &EngineMismatch{Optimization: actual.Optimization, Input: inputs[index], Expected: string(expectedBytes), Actual: string(actualBytes)}
			}
		}
	}
	return NativeResult{CSource: cSource, Interpreter: interpreted, O0: o0, O3: o3}, nil, nil
}

func RunNative(ctx context.Context, source []byte, runner NativeRunner) (NativeResult, []diagnostic.Diagnostic, error) {
	return runNative(ctx, source, runner, "")
}

// BuildApplication checks and independently validates source, then emits and
// compiles a retained U64 application without executing its entry point.
func BuildApplication(ctx context.Context, source []byte, outputPath string, runner native.Runner, manifestPath ...string) (native.BuildReceipt, []diagnostic.Diagnostic, error) {
	checked := Check(source)
	if len(checked.Diagnostics) > 0 {
		return native.BuildReceipt{}, checked.Diagnostics, nil
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		if len(validated.Problems) == 0 {
			return native.BuildReceipt{}, nil, errors.New("core validation failed without a named problem")
		}
		return native.BuildReceipt{}, nil, fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
	}
	program := validated.Program()
	entry, err := callgraph.EntryFunction(program)
	if err != nil {
		return native.BuildReceipt{}, nil, err
	}
	localOwnerEntry := entry.Parameter.Type == "PathToken" && hasLocalForeignOperations(program)
	if entry.Match != nil || entry.Linear == nil || (entry.Parameter.Type != "U64" && !localOwnerEntry) || entry.ReturnType != "U64" {
		return native.BuildReceipt{}, nil, fmt.Errorf("application entry %q: only a linear U64-to-U64 or checked PathToken-to-U64 entry is supported", entry.ID)
	}
	if localOwnerEntry {
		if err := pathoracle.ValidateLocalOwnerPaths(program); err != nil {
			return native.BuildReceipt{}, nil, err
		}
		if len(manifestPath) != 1 {
			return native.BuildReceipt{}, nil, errors.New("local-owner application requires one explicit binding manifest")
		}
		bindings, err := native.ResolveBindings(manifestPath[0])
		if err != nil {
			return native.BuildReceipt{}, nil, err
		}
		if err := validateLocalOperationBindings(program, bindings.Manifest); err != nil {
			return native.BuildReceipt{}, nil, err
		}
	}
	cSource, err := cgen.EmitApplication(program)
	if err != nil {
		return native.BuildReceipt{}, nil, err
	}
	receipt, err := runner.BuildApplication(ctx, source, cSource, outputPath, manifestPath...)
	if err != nil {
		return native.BuildReceipt{}, nil, err
	}
	return receipt, nil, nil
}

func hasLocalForeignOperations(program core.Program) bool {
	for _, function := range program.Functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Foreign != nil {
				return true
			}
		}
	}
	return false
}

func validateLocalOperationBindings(program core.Program, manifest native.BindingManifest) error {
	expected := map[string]string{}
	for _, function := range program.Functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Foreign == nil {
				continue
			}
			if previous, exists := expected[operation.Foreign.Symbol]; exists && previous != operation.Foreign.ABIType {
				return fmt.Errorf("local binding %q has conflicting checked ABI types", operation.Foreign.Symbol)
			}
			expected[operation.Foreign.Symbol] = operation.Foreign.ABIType
		}
	}
	if len(expected) == 0 || len(expected) != len(manifest.Symbols) {
		return errors.New("local binding manifest symbols do not exactly cover checked foreign operations")
	}
	for _, symbol := range manifest.Symbols {
		abiType, exists := expected[symbol.Name]
		if !exists || abiType != symbol.FunctionType {
			return fmt.Errorf("local binding %q does not match its checked operation ABI", symbol.Name)
		}
	}
	return nil
}

func BuildApplicationFile(ctx context.Context, sourcePath, outputPath string, runner native.Runner, manifestPath ...string) (native.BuildReceipt, []diagnostic.Diagnostic, error) {
	source, err := readBoundedFile(sourcePath, syntax.MaxSourceBytes)
	if err != nil {
		return native.BuildReceipt{}, nil, err
	}
	return BuildApplication(ctx, source, outputPath, runner, manifestPath...)
}

// runNativeInputs is task 04-07-03's own bug fix, discovered by driving the
// shipped binary on out-of-corpus programs per D-04-21: `schway run
// --engine=native` previously ran every input through ONE shared,
// zero-value native.Runner.Expect (silently defaulting to ExpectValue), so
// ANY program whose real terminal outcome is a typed failure or a defect --
// whether a single-input resource/nonlocal-exit probe or one arm of a
// Match-shaped multi-arm function like defect_terminal.schway -- was
// rejected as an operational `native.run_signaled` failure. This affected
// EXISTING, already-committed Phase 4 corpus fixtures (nonlocal_exit_probe.schway),
// not only new out-of-corpus programs -- exactly the class of gap this
// closing plan exists to catch (Rule 1: a bug affecting a real, reachable
// shipped-binary command). Every input is now run SEPARATELY against a
// concrete native.Runner with its own Expect, derived from the SAME
// interpreter verdict this function already computed as its own oracle for
// that exact input -- never a new, independent guess. For a single-input,
// ordinary-returning program this derivation always resolves to
// ExpectValue, the prior default, so no currently-passing single-input
// differential changes behavior. Any NativeRunner that is not a concrete
// native.Runner (a mutation-runner wrapper) is untouched, calling through
// unchanged.
func runNativeInputs(ctx context.Context, runner NativeRunner, cSource, optimization string, inputs []string, interpreted []interp.Execution) (native.Result, error) {
	concrete, ok := runner.(native.Runner)
	if !ok {
		return runner.Run(ctx, cSource, optimization, inputs)
	}
	merged := native.Result{Optimization: optimization, Pairs: make([]native.Pair, 0, len(inputs))}
	for index, input := range inputs {
		perInput := concrete
		perInput.Expect = expectForOutcomeKind(interpreted[index].Outcome.Kind)
		result, err := perInput.Run(ctx, cSource, optimization, []string{input})
		if err != nil {
			return native.Result{}, err
		}
		merged.Pairs = append(merged.Pairs, result.Pairs...)
		merged.CompileTime += result.CompileTime
		merged.RunTime += result.RunTime
		merged.OutputBytes += result.OutputBytes
	}
	return merged, nil
}

// expectForOutcomeKind maps an interpreter-observed terminal outcome kind
// to the matching native.TerminalOutcome expectation.
func expectForOutcomeKind(kind string) native.TerminalOutcome {
	switch kind {
	case "typed_failure":
		return native.ExpectTypedFailure
	case execution.OutcomeDefect:
		return native.ExpectDefect
	default:
		return native.ExpectValue
	}
}

func RunNativeFile(ctx context.Context, path string, runner NativeRunner) (NativeResult, []diagnostic.Diagnostic, error) {
	source, err := readBoundedFile(path, syntax.MaxSourceBytes)
	if err != nil {
		return NativeResult{}, nil, err
	}
	relative, relErr := filepath.Rel(nat03ProjectRoot(), path)
	if relErr != nil {
		return NativeResult{}, nil, relErr
	}
	return runNative(ctx, source, runner, filepath.ToSlash(relative))
}

func RunNativeCommandFile(ctx context.Context, path string, runner NativeRunner) (protocol.Result, error) {
	started := time.Now()
	nativeResult, diagnostics, err := RunNativeFile(ctx, path, runner)
	if len(diagnostics) == 0 && err == nil {
		if originProblems := publishedOriginProblemFile(path); len(originProblems) > 0 {
			diagnostics = originProblems
		}
	}
	result := protocol.New("run", protocol.StatusPass)
	if len(diagnostics) > 0 {
		result.Status = protocol.StatusInvalid
		result.Diagnostics = diagnostics
		return completeCommand(result, started, 1), nil
	}
	if err != nil {
		result.Status = protocol.StatusOperational
		code := "native.tool_failure"
		var mismatch *EngineMismatch
		if errors.As(err, &mismatch) {
			result.Status = protocol.StatusMismatch
			code = "native.engine_mismatch"
		}
		var toolError *native.ToolError
		if errors.As(err, &toolError) {
			code = toolError.Code
		}
		result.Diagnostics = []diagnostic.Diagnostic{diagnostic.Error(code, diagnostic.Span{}, "native execution did not complete successfully")}
		return completeCommand(result, started, 1), nil
	}
	result.Executions = nativeResult.Interpreter
	return completeCommand(result, started, len(nativeResult.Interpreter)*3), nil
}

func EvidenceCommandFile(ctx context.Context, path string) (evidence.Product, protocol.Result, error) {
	started := time.Now()
	source, err := readBoundedFile(path, syntax.MaxSourceBytes)
	if err != nil {
		return evidence.Product{}, protocol.Result{}, err
	}
	facts, err := evidence.DefaultFacts(ctx, "clang")
	if err != nil {
		result := protocol.New("evidence", protocol.StatusOperational)
		result.Diagnostics = []diagnostic.Diagnostic{diagnostic.Error("evidence.tool_failure", diagnostic.Span{}, "unable to inspect native toolchain")}
		return evidence.Product{}, completeCommand(result, started, 1), nil
	}
	product, diagnostics, err := evidence.Build(source, facts)
	if err != nil {
		return evidence.Product{}, protocol.Result{}, err
	}
	result := protocol.New("evidence", protocol.StatusPass)
	if len(diagnostics) > 0 {
		result.Status = protocol.StatusInvalid
		result.Diagnostics = diagnostics
		return evidence.Product{}, completeCommand(result, started, 1), nil
	}
	result.Evidence = &protocol.EvidenceSummary{Schema: product.Manifest.Schema, ID: product.Manifest.ID, Digest: evidence.ContentDigest(product.ManifestBytes)}
	return product, completeCommand(result, started, 3), nil
}

func ValidateEvidenceCommandFile(ctx context.Context, manifestPath, sourcePath string) protocol.Result {
	manifestBytes, err := readBoundedFile(manifestPath, evidence.MaxManifestBytes)
	if err != nil {
		return commandProblem("evidence", protocol.StatusOperational, "tool.read_failed", "unable to read evidence manifest")
	}
	source, err := readBoundedFile(sourcePath, syntax.MaxSourceBytes)
	if err != nil {
		return commandProblem("evidence", protocol.StatusOperational, "tool.read_failed", "unable to read source")
	}
	manifest, err := evidence.DecodeStrict(manifestBytes)
	if err != nil {
		return commandProblem("evidence", protocol.StatusInvalid, evidence.ErrorCode(err), "evidence manifest is not valid")
	}
	facts, err := evidence.DefaultFacts(ctx, "clang")
	if err != nil {
		return commandProblem("evidence", protocol.StatusOperational, "evidence.tool_failure", "unable to inspect native toolchain")
	}
	if err := evidence.Validate(manifest, source, facts); err != nil {
		return commandProblem("evidence", protocol.StatusInvalid, evidence.ErrorCode(err), "evidence manifest does not match recomputed facts")
	}
	result := protocol.New("evidence", protocol.StatusPass)
	result.Evidence = &protocol.EvidenceSummary{Schema: manifest.Schema, ID: manifest.ID, Digest: evidence.ContentDigest(manifestBytes)}
	return result.Finalize()
}

// MaxInterfaceBytes bounds the two new Phase 3 (OWN-04) untrusted input
// surfaces this command family reads: a published interface summary and the
// raw core artifact bytes a `interface check` invocation binds it against.
// Sized like evidence.MaxManifestBytes — both are small, bounded JSON
// artifacts, not source.
const MaxInterfaceBytes = 1 << 21

// InterfaceExportCommandFile is the producer side of OWN-04's separate
// compilation demonstration: it checks sourcePath, independently recomputes
// every declared PublicOrigin fact from the typed core alone
// (originvalidate.ValidatePublished — T-03-02/T-03-16), and only on success
// writes a body-stripped, digest-bound core.Interface summary to outPath.
func InterfaceExportCommandFile(sourcePath, outPath string) (protocol.Result, error) {
	started := time.Now()
	source, err := readBoundedFile(sourcePath, syntax.MaxSourceBytes)
	if err != nil {
		return protocol.Result{}, err
	}
	checked := Check(source)
	result := protocol.New("interface", protocol.StatusPass)
	if len(checked.Diagnostics) > 0 {
		result.Status = protocol.StatusInvalid
		result.Diagnostics = checked.Diagnostics
		return completeCommand(result, started, checked.Work), nil
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		diag := peerRefusalDiagnostic(validated, checked.Program.ModuleID)
		if interfacePeerRefusalSeam {
			// control:interface.peer_refusal_is_invalid's own fault: restores
			// the pre-07-10 behaviour of discarding the peer's own code
			// behind a generic tool-failure error (07-REVIEW.md CR-04).
			return protocol.Result{}, fmt.Errorf("interface command: peer refusal seam active, would report %s", diag.Code)
		}
		result.Status = protocol.StatusInvalid
		result.Diagnostics = []diagnostic.Diagnostic{diag}
		return completeCommand(result, started, checked.Work), nil
	}
	checked.Program = validated.Program()
	if problems := originvalidate.ValidatePublished(checked.Program); len(problems) > 0 {
		result.Status = protocol.StatusInvalid
		result.Diagnostics = []diagnostic.Diagnostic{diagnostic.Error(problems[0].Code, diagnostic.Span{}, problems[0].Detail)}
		return completeCommand(result, started, checked.Work), nil
	}
	summary, err := originvalidate.BuildInterface(checked.Program)
	if err != nil {
		return protocol.Result{}, err
	}
	summaryBytes, err := json.Marshal(summary)
	if err != nil {
		return protocol.Result{}, err
	}
	if err := os.WriteFile(outPath, append(summaryBytes, '\n'), 0o600); err != nil {
		return protocol.Result{}, err
	}
	result.Interface = interfaceProjection(summary.Schema, summary.ModuleID, summary.CoreDigest, summary.Functions)
	result.ExpectedEscapes = originvalidate.ExpectedEscapes()
	return completeCommand(result, started, len(summary.Functions)), nil
}

// InterfaceCoreCommandFile writes the exact checked-and-validated core.Program
// bytes for sourcePath to outPath — the same bytes InterfaceExportCommandFile
// digests into a summary's CoreDigest. A real build pipeline already retains
// this artifact from `schway check`; this command exists so the two-invocation
// separate-compilation demonstration (export, then check) has a standalone
// way to obtain the core artifact a summary is bound to, without requiring
// `interface check` itself to reconstruct or re-derive it.
func InterfaceCoreCommandFile(sourcePath, outPath string) (protocol.Result, error) {
	started := time.Now()
	source, err := readBoundedFile(sourcePath, syntax.MaxSourceBytes)
	if err != nil {
		return protocol.Result{}, err
	}
	checked := Check(source)
	result := protocol.New("interface", protocol.StatusPass)
	if len(checked.Diagnostics) > 0 {
		result.Status = protocol.StatusInvalid
		result.Diagnostics = checked.Diagnostics
		return completeCommand(result, started, checked.Work), nil
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		diag := peerRefusalDiagnostic(validated, checked.Program.ModuleID)
		if interfacePeerRefusalSeam {
			// control:interface.peer_refusal_is_invalid's own fault: restores
			// the pre-07-10 behaviour of discarding the peer's own code
			// behind a generic tool-failure error (07-REVIEW.md CR-04).
			return protocol.Result{}, fmt.Errorf("interface command: peer refusal seam active, would report %s", diag.Code)
		}
		result.Status = protocol.StatusInvalid
		result.Diagnostics = []diagnostic.Diagnostic{diag}
		return completeCommand(result, started, checked.Work), nil
	}
	program := validated.Program()
	coreBytes, err := json.Marshal(program)
	if err != nil {
		return protocol.Result{}, err
	}
	// Written byte-for-byte with no appended newline: this is exactly the
	// digest preimage originvalidate.BuildInterface hashes into a summary's
	// CoreDigest, so a byte a consumer reads back here must match precisely.
	if err := os.WriteFile(outPath, coreBytes, 0o600); err != nil {
		return protocol.Result{}, err
	}
	result.ModuleID = program.ModuleID
	return completeCommand(result, started, len(program.Functions)), nil
}

// InterfaceCheckCommandFile is the consumer side of OWN-04's separate
// compilation demonstration: a genuinely separate CLI invocation that
// answers origin/access questions from summaryPath alone. coreBytesPath's
// content is hashed (via originvalidate.CheckSummary) and compared against
// the summary's recorded digest and is NEVER decoded as a core.Program — the
// consuming path never touches a body field, by construction, not by
// discipline.
func InterfaceCheckCommandFile(summaryPath, coreBytesPath string) (protocol.Result, error) {
	started := time.Now()
	summaryBytes, err := readBoundedFile(summaryPath, MaxInterfaceBytes)
	if err != nil {
		return protocol.Result{}, err
	}
	coreBytes, err := readBoundedFile(coreBytesPath, MaxInterfaceBytes)
	if err != nil {
		return protocol.Result{}, err
	}
	answers, checkErr := originvalidate.CheckSummary(summaryBytes, coreBytes)
	result := protocol.New("interface", protocol.StatusPass)
	if checkErr != nil {
		result.Status = protocol.StatusInvalid
		code := "origin.invalid_summary"
		var originError *originvalidate.Error
		if errors.As(checkErr, &originError) {
			code = originError.Code
		}
		result.Diagnostics = []diagnostic.Diagnostic{diagnostic.Error(code, diagnostic.Span{}, "interface summary rejected")}
		return completeCommand(result, started, 1), nil
	}
	var summary core.Interface
	if err := json.Unmarshal(summaryBytes, &summary); err != nil {
		return protocol.Result{}, err
	}
	functions := make([]protocol.InterfaceFunctionAnswer, 0, len(answers))
	for _, answer := range answers {
		functions = append(functions, protocol.InterfaceFunctionAnswer{ID: answer.ID, Name: answer.Name, Paths: answer.Paths, Access: answer.Access})
	}
	result.Interface = &protocol.InterfaceSummary{Schema: summary.Schema, ModuleID: summary.ModuleID, CoreDigest: summary.CoreDigest, Functions: functions}
	result.ExpectedEscapes = originvalidate.ExpectedEscapes()
	return completeCommand(result, started, len(answers)), nil
}

// DebugMapCommandFile is the CLI seam for the bounded debug-lineage
// experiment (D-01..D-04, Task 03-07-01): it checks and validates sourcePath
// exactly like `check`, then joins the honestly-checked source and core
// artifacts via debugmap.Build. When query is non-empty, the result contains
// only the single resolved entry for that operation ID (debugmap.Resolve),
// proving the honest-absence report through the shipped binary rather than
// only in-process; an empty query returns the full joined map.
func DebugMapCommandFile(path, query string) (protocol.Result, error) {
	started := time.Now()
	source, err := readBoundedFile(path, syntax.MaxSourceBytes)
	if err != nil {
		return protocol.Result{}, err
	}
	parsed := syntax.Parse(source)
	result := protocol.New("debug-map", protocol.StatusPass)
	if len(parsed.Diagnostics) > 0 {
		result.Status = protocol.StatusInvalid
		result.Diagnostics = parsed.Diagnostics
		return completeCommand(result, started, 1), nil
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) > 0 {
		result.Status = protocol.StatusInvalid
		result.Diagnostics = checked.Diagnostics
		return completeCommand(result, started, checked.Work), nil
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		return protocol.Result{}, fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
	}
	program := validated.Program()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	built, work, buildErr := debugmap.Build(ctx, parsed.Program, program)
	if buildErr != nil {
		code := "debugmap.build_failed"
		var debugErr *debugmap.Error
		if errors.As(buildErr, &debugErr) {
			code = debugErr.Code
		}
		result.Status = protocol.StatusOperational
		result.Diagnostics = []diagnostic.Diagnostic{diagnostic.Error(code, diagnostic.Span{}, "debug map construction failed")}
		return completeCommand(result, started, work), nil
	}

	entries := built.Entries
	if query != "" {
		entries = []debugmap.Entry{debugmap.Resolve(built, query)}
	}
	protocolEntries := make([]protocol.DebugMapEntry, 0, len(entries))
	for _, entry := range entries {
		protocolEntries = append(protocolEntries, protocol.DebugMapEntry{
			ID: entry.ID, CoreID: entry.CoreID, OperationID: entry.OperationID, PointID: entry.PointID,
			Kind: entry.Kind, Availability: string(entry.Availability),
		})
	}
	result.DebugMap = &protocol.DebugMapSummary{Schema: built.Schema, Entries: protocolEntries}
	return completeCommand(result, started, work), nil
}

// interfaceProjection builds the R-02 deliberately-lossy CLI projection of a
// lang.interface/1 summary (protocol.InterfaceSummary is the artifact of
// record's lossy peer, never a second copy of it — see
// protocol.InterfaceFunctionAnswer's doc comment). Function.Return replaces
// the pre-/1 optional PublicOrigin field: Mode == "owned" carries no
// origin/access to project, matching the old nil-PublicOrigin case exactly.
func interfaceProjection(schema, moduleID, coreDigest string, functions []core.FunctionSignature) *protocol.InterfaceSummary {
	summary := &protocol.InterfaceSummary{Schema: schema, ModuleID: moduleID, CoreDigest: coreDigest, Functions: make([]protocol.InterfaceFunctionAnswer, 0, len(functions))}
	for _, function := range functions {
		answer := protocol.InterfaceFunctionAnswer{ID: function.ID, Name: function.Name}
		if function.Return.Mode != "owned" {
			answer.Paths = function.Return.Paths
			answer.Access = function.Return.Mode
		}
		summary.Functions = append(summary.Functions, answer)
	}
	return summary
}

func readBoundedFile(path string, limit int) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(io.LimitReader(file, int64(limit)+1))
}

func commandProblem(command, status, code, message string) protocol.Result {
	result := protocol.New(command, status)
	result.Diagnostics = []diagnostic.Diagnostic{diagnostic.Error(code, diagnostic.Span{}, message)}
	return result.Finalize()
}

// TimingObservationEnabled reports whether SCHWAY_OBSERVE_TIMING is set to
// "1" -- the ONE place in the tree that reads this environment variable
// (TestTimingEnvironmentIsReadInExactlyOnePlace proves this by AST scan).
// completeCommand and StageRecorder.Breakdown both route through this
// single exported gate rather than each reading the variable a second way,
// so a golden/pinned-JSON test run with the switch unset can never
// intermittently fail from an unconditional time.Since (D-06-21).
func TimingObservationEnabled() bool {
	return os.Getenv("SCHWAY_OBSERVE_TIMING") == "1"
}

func completeCommand(result protocol.Result, started time.Time, work int) protocol.Result {
	// Bounded commands keep their default projection reproducible. The verify
	// orchestrator records wall-time observations explicitly; later telemetry
	// modes can opt into per-command timing without making ordinary output churn.
	if TimingObservationEnabled() {
		result.Metrics.ElapsedNS = time.Since(started).Nanoseconds()
	}
	result.Metrics.RecomputedWork = work
	return result.Finalize()
}

type VerifyOptions struct {
	ForceEngineMismatch bool
	ForceStaleManifest  bool
}

func VerifyCorpus(ctx context.Context, corpus string, runner native.Runner, options VerifyOptions) protocol.Result {
	if _, err := os.Stat(filepath.Join(corpus, "foreign_acquire_one.schway")); err == nil {
		return verifyForeignCorpus(ctx, corpus, runner)
	}
	if _, err := os.Stat(filepath.Join(corpus, "borrowed_view.schway")); err == nil {
		return verifyBorrowedCorpus(ctx, corpus, runner)
	}
	if _, err := os.Stat(filepath.Join(corpus, "owned_transfer.schway")); err == nil {
		return verifyOwnedCorpus(ctx, corpus, runner)
	}
	started := time.Now()
	result := protocol.New("verify", protocol.StatusPass)
	addLane := func(id, status string, controls []string, work, outputBytes int, laneStarted time.Time) {
		result.Lanes = append(result.Lanes, protocol.Lane{
			Schema: protocol.LaneSchema1, ID: id, Status: status,
			Controls: append([]string{}, controls...), RecomputedWork: work,
			ElapsedNS: time.Since(laneStarted).Nanoseconds(), PeakRSSStatus: protocol.PeakRSSUnavailable, OutputBytes: outputBytes,
		})
		result.Metrics.RecomputedWork += work
		result.Metrics.OutputBytes += outputBytes
	}
	fail := func(status, code, message string) protocol.Result {
		result.Status = status
		result.Diagnostics = append(result.Diagnostics, diagnostic.Error(code, diagnostic.Span{}, message))
		result.Metrics.ElapsedNS = time.Since(started).Nanoseconds()
		return result.Finalize()
	}

	validPath := filepath.Join(corpus, "toggle.schway")
	validSource, err := readBoundedFile(validPath, syntax.MaxSourceBytes)
	if err != nil {
		return fail(protocol.StatusOperational, "verify.fixture_missing", "required positive fixture is unavailable")
	}
	if len(validSource) > syntax.MaxSourceBytes {
		return fail(protocol.StatusInvalid, "verify.fixture_input_limit", "required positive fixture exceeds the source byte limit")
	}

	laneStarted := time.Now()
	checked := Check(validSource)
	firstExecutions, firstDiagnostics, firstErr := RunInterpreter(validSource)
	secondExecutions, secondDiagnostics, secondErr := RunInterpreter(validSource)
	deterministicPass := len(checked.Diagnostics) == 0 && len(firstDiagnostics) == 0 && len(secondDiagnostics) == 0 && firstErr == nil && secondErr == nil && reflect.DeepEqual(firstExecutions, secondExecutions)
	if !deterministicPass {
		addLane("lane:deterministic", "fail", nil, 3, 0, laneStarted)
		return fail(protocol.StatusMismatch, "verify.determinism_failed", "deterministic checking or interpretation disagreed")
	}
	addLane("lane:deterministic", "pass", nil, 3, len(validSource), laneStarted)

	laneStarted = time.Now()
	syntaxWork := 0
	syntaxBytes := 0
	for _, name := range []string{"toggle.schway", "comments.schway", "malformed.schway"} {
		source, readErr := readBoundedFile(filepath.Join(corpus, name), syntax.MaxSourceBytes)
		if readErr != nil {
			addLane("lane:syntax-properties", "fail", nil, syntaxWork, syntaxBytes, laneStarted)
			return fail(protocol.StatusOperational, "verify.fixture_missing", "required syntax fixture is unavailable")
		}
		if len(source) > syntax.MaxSourceBytes {
			return fail(protocol.StatusInvalid, "verify.fixture_input_limit", name)
		}
		parsed := syntax.Parse(source)
		syntaxWork++
		syntaxBytes += len(source)
		if !bytes.Equal(parsed.Tree.Bytes(), source) || len(parsed.Diagnostics) > 21 {
			addLane("lane:syntax-properties", "fail", nil, syntaxWork, syntaxBytes, laneStarted)
			return fail(protocol.StatusInvalid, "verify.syntax_property_failed", "lossless or bounded syntax property failed")
		}
		if len(parsed.Diagnostics) == 0 {
			canonical := syntax.Format(parsed.Tree)
			reparsed := syntax.Parse(canonical)
			syntaxWork++
			if len(reparsed.Diagnostics) != 0 || !bytes.Equal(canonical, syntax.Format(reparsed.Tree)) {
				addLane("lane:syntax-properties", "fail", nil, syntaxWork, syntaxBytes+len(canonical), laneStarted)
				return fail(protocol.StatusInvalid, "verify.syntax_property_failed", "canonical syntax fixed point failed")
			}
			syntaxBytes += len(canonical)
		}
	}
	addLane("lane:syntax-properties", "pass", nil, syntaxWork, syntaxBytes, laneStarted)

	laneStarted = time.Now()
	negativeSource, err := readBoundedFile(filepath.Join(corpus, "non_exhaustive.schway"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:negative-controls", "fail", nil, 0, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "required negative fixture is unavailable")
	}
	if len(negativeSource) > syntax.MaxSourceBytes {
		return fail(protocol.StatusInvalid, "verify.fixture_input_limit", "non_exhaustive.schway")
	}
	negative := Check(negativeSource)
	if !hasDiagnostic(negative.Diagnostics, "match.non_exhaustive") {
		addLane("lane:negative-controls", "fail", nil, 1, len(negativeSource), laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", "missing-arm control did not fail as expected")
	}
	addLane("lane:negative-controls", "pass", []string{"control:match.non_exhaustive"}, 1, len(negativeSource), laneStarted)

	laneStarted = time.Now()
	facts, err := evidence.DefaultFacts(ctx, runner.ClangPath)
	if err != nil {
		addLane("lane:evidence-bindings", "fail", nil, 0, 0, laneStarted)
		return fail(protocol.StatusOperational, "evidence.tool_failure", "unable to inspect native toolchain")
	}
	product, evidenceDiagnostics, err := evidence.Build(validSource, facts)
	if err != nil || len(evidenceDiagnostics) != 0 {
		addLane("lane:evidence-bindings", "fail", nil, 1, 0, laneStarted)
		return fail(protocol.StatusInvalid, "verify.evidence_build_failed", "unable to build current evidence")
	}
	base := product.Manifest
	if options.ForceStaleManifest {
		base.SourceDigest = "sha256:forced-stale-control"
	}
	if err := evidence.Validate(base, validSource, facts); err != nil {
		addLane("lane:evidence-bindings", "fail", nil, 2, len(product.ManifestBytes), laneStarted)
		return fail(protocol.StatusInvalid, evidence.ErrorCode(err), "current evidence failed validation")
	}
	stale := product.Manifest
	stale.SourceDigest = "sha256:deliberately-stale-control"
	if err := evidence.Validate(stale, validSource, facts); evidence.ErrorCode(err) != "evidence.source_mismatch" {
		addLane("lane:evidence-bindings", "fail", nil, 3, len(product.ManifestBytes), laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", "stale evidence control did not fail as expected")
	}
	addLane("lane:evidence-bindings", "pass", []string{"control:evidence.source_mismatch"}, 3, len(product.ManifestBytes), laneStarted)
	result.Evidence = &protocol.EvidenceSummary{Schema: product.Manifest.Schema, ID: product.Manifest.ID, Digest: evidence.ContentDigest(product.ManifestBytes)}

	laneStarted = time.Now()
	nativeResult, nativeDiagnostics, err := RunNative(ctx, validSource, runner)
	if len(nativeDiagnostics) != 0 {
		addLane("lane:native-differential", "fail", nil, 1, 0, laneStarted)
		return fail(protocol.StatusInvalid, "verify.native_source_invalid", "native fixture became invalid")
	}
	if err != nil {
		addLane("lane:native-differential", "fail", nil, 1, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.native_failed", "native differential lane did not complete")
	}
	if options.ForceEngineMismatch && len(nativeResult.O3.Pairs) > 0 {
		nativeResult.O3.Pairs[0].Execution.Outcome.Value += "-forced-mismatch"
	}
	for index, execution := range nativeResult.Interpreter {
		if !executionpkgEqual(execution, nativeResult.O0.Pairs[index].Execution) || !executionpkgEqual(execution, nativeResult.O3.Pairs[index].Execution) {
			addLane("lane:native-differential", "fail", []string{"control:interpreter-o0-o3"}, len(nativeResult.Interpreter)*3, nativeResult.O0.OutputBytes+nativeResult.O3.OutputBytes, laneStarted)
			return fail(protocol.StatusMismatch, "native.engine_mismatch", "interpreter, O0, and O3 outcomes disagree")
		}
	}
	addLane("lane:native-differential", "pass", []string{"control:interpreter-o0-o3"}, len(nativeResult.Interpreter)*3, nativeResult.O0.OutputBytes+nativeResult.O3.OutputBytes, laneStarted)

	requiredControls := []string{"control:match.non_exhaustive", "control:evidence.source_mismatch", "control:interpreter-o0-o3"}
	for _, required := range requiredControls {
		if !hasControl(result.Lanes, required) {
			return fail(protocol.StatusInvalid, "verify.control_missing", "a required verification control was not observed")
		}
	}
	for _, lane := range result.Lanes {
		if lane.RecomputedWork == 0 {
			return fail(protocol.StatusInvalid, "verify.zero_work", "a verification lane performed no work")
		}
	}
	result.Metrics.ElapsedNS = time.Since(started).Nanoseconds()
	return result.Finalize()
}

func verifyOwnedCorpus(ctx context.Context, corpus string, runner native.Runner) protocol.Result {
	started := time.Now()
	result := protocol.New("verify", protocol.StatusPass)
	result.ExpectedEscapes = []string{corevalidate.KnownEscape}
	addLane := func(id string, controls []string, work, outputBytes int, laneStarted time.Time) {
		result.Lanes = append(result.Lanes, protocol.Lane{Schema: protocol.LaneSchema1, ID: id, Status: "pass", Controls: append([]string(nil), controls...), RecomputedWork: work, ElapsedNS: time.Since(laneStarted).Nanoseconds(), PeakRSSStatus: protocol.PeakRSSUnavailable, OutputBytes: outputBytes})
		result.Metrics.RecomputedWork += work
		result.Metrics.OutputBytes += outputBytes
	}
	fail := func(status, code, message string) protocol.Result {
		result.Status = status
		result.Diagnostics = append(result.Diagnostics, diagnostic.Error(code, diagnostic.Span{}, message))
		result.Metrics.ElapsedNS = time.Since(started).Nanoseconds()
		return result.Finalize()
	}

	laneStarted := time.Now()
	ownershipControls := []struct{ file, code, control string }{
		{"use_after_move.schway", "ownership.use_after_move", "control:ownership.use_after_move"},
		{"move_while_borrowed.schway", "ownership.move_while_borrowed", "control:ownership.move_while_borrowed"},
		{"implicit_noncopy.schway", "ownership.transfer_requires_take", "control:ownership.transfer_requires_take"},
		{"reborrow_while_moved.schway", "ownership.move_while_borrowed", "control:ownership.move_while_reborrowed"},
	}
	ownershipControlNames := make([]string, 0, len(ownershipControls))
	for _, control := range ownershipControls {
		ownershipControlNames = append(ownershipControlNames, control.control)
	}
	ownershipBytes := 0
	for _, control := range ownershipControls {
		source, err := readBoundedFile(filepath.Join(corpus, control.file), syntax.MaxSourceBytes)
		if err != nil {
			return fail(protocol.StatusOperational, "verify.fixture_missing", control.file)
		}
		if len(source) > syntax.MaxSourceBytes {
			return fail(protocol.StatusInvalid, "verify.fixture_input_limit", control.file)
		}
		ownershipBytes += len(source)
		checked := Check(source)
		if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != control.code {
			return fail(protocol.StatusInvalid, "verify.control_missing", control.control)
		}
	}
	addLane("lane:owned-negative-controls", ownershipControlNames, len(ownershipControls), ownershipBytes, laneStarted)

	validSource, err := readBoundedFile(filepath.Join(corpus, "owned_transfer.schway"), syntax.MaxSourceBytes)
	if err != nil {
		return fail(protocol.StatusOperational, "verify.fixture_missing", "owned_transfer.schway")
	}
	if len(validSource) > syntax.MaxSourceBytes {
		return fail(protocol.StatusInvalid, "verify.fixture_input_limit", "owned_transfer.schway")
	}
	checked := Check(validSource)
	// Phase 11 (11-GUARD-LEDGER.md): KEPT. verifyOwnedCorpus is bound to
	// the fixed Phase 2 fixture owned_transfer.schway, genuinely
	// single-function by construction; it is unrelated to the
	// multi-function corpus this phase widens.
	if len(checked.Diagnostics) != 0 || len(checked.Program.Functions) != 1 {
		return fail(protocol.StatusInvalid, "verify.owned_invalid", "owned transfer fixture is invalid")
	}

	laneStarted = time.Now()
	forged := corevalidate.Validate(checked.Program).Program()
	forged.Functions[0].Linear.Types[0].Abilities = append([]core.Ability{core.AbilityCopy}, forged.Functions[0].Linear.Types[0].Abilities...)
	forgedResult := corevalidate.Validate(forged)
	duplicate := corevalidate.Validate(checked.Program).Program()
	duplicate.Functions[0].Linear.Operations[1].ID = duplicate.Functions[0].Linear.Operations[0].ID
	duplicateResult := corevalidate.Validate(duplicate)
	if forgedResult.Valid || len(forgedResult.Problems) == 0 || forgedResult.Problems[0].Code != "core.ability_mismatch" || duplicateResult.Valid || len(duplicateResult.Problems) == 0 || duplicateResult.Problems[0].Code != "core.duplicate_operation_id" {
		return fail(protocol.StatusInvalid, "verify.control_missing", "owned core mutation survived")
	}
	addLane("lane:owned-core-controls", []string{"control:ability.forged_copy", "control:core.duplicate_operation_id"}, forgedResult.Checks+duplicateResult.Checks, len(validSource), laneStarted)

	laneStarted = time.Now()
	nativeResult, diagnostics, err := RunNative(ctx, validSource, runner)
	if err != nil || len(diagnostics) != 0 || len(nativeResult.Interpreter) == 0 {
		return fail(protocol.StatusOperational, "verify.native_failed", "owned native differential did not complete")
	}
	addLane("lane:owned-native-differential", []string{"control:interpreter-o0-o3-owned"}, len(nativeResult.Interpreter)*3, nativeResult.O0.OutputBytes+nativeResult.O3.OutputBytes, laneStarted)

	laneStarted = time.Now()
	mutationRunner := NewOwnedBackendMutationRunner(runner)
	_, mutationDiagnostics, mutationErr := RunNative(ctx, validSource, mutationRunner)
	var mutationMismatch *EngineMismatch
	if len(mutationDiagnostics) != 0 {
		return fail(protocol.StatusInvalid, "verify.control_invalid", "backend causality control produced source diagnostics")
	}
	if !errors.As(mutationErr, &mutationMismatch) {
		var toolError *native.ToolError
		if errors.As(mutationErr, &toolError) {
			return fail(protocol.StatusOperational, toolError.Code, "backend causality control could not execute")
		}
		return fail(protocol.StatusMismatch, "verify.control_missing", "backend value mutation did not cause an engine mismatch")
	}
	if !reflect.DeepEqual(mutationRunner.Optimizations(), []string{"-O0", "-O3"}) {
		return fail(protocol.StatusOperational, "verify.control_incomplete", "backend causality control did not execute both optimization modes")
	}
	addLane("lane:owned-backend-causality", []string{"control:backend.runtime_causality"}, len(mutationRunner.Optimizations()), 0, laneStarted)

	laneStarted = time.Now()
	facts, err := evidence.DefaultFacts(ctx, runner.ClangPath)
	if err != nil {
		return fail(protocol.StatusOperational, "evidence.tool_failure", "unable to inspect native toolchain")
	}
	product, evidenceDiagnostics, err := evidence.Build(validSource, facts)
	if err != nil || len(evidenceDiagnostics) != 0 {
		return fail(protocol.StatusInvalid, "verify.evidence_build_failed", "unable to build owned evidence")
	}
	mutated := product.Manifest
	mutated.CoreDigest = "sha256:deliberately-stale-owned-core"
	if err := evidence.Validate(mutated, validSource, facts); evidence.ErrorCode(err) != "evidence.core_mismatch" {
		return fail(protocol.StatusInvalid, "verify.control_missing", "owned core evidence mutation survived")
	}
	addLane("lane:owned-evidence-bindings", []string{"control:evidence.core_mismatch"}, 2, len(product.ManifestBytes), laneStarted)
	result.Evidence = &protocol.EvidenceSummary{Schema: product.Manifest.Schema, ID: product.Manifest.ID, Digest: evidence.ContentDigest(product.ManifestBytes)}

	requiredControls := []string{
		"control:ownership.use_after_move",
		"control:ownership.move_while_borrowed",
		"control:ownership.transfer_requires_take",
		"control:ownership.move_while_reborrowed",
		"control:ability.forged_copy",
		"control:core.duplicate_operation_id",
		"control:interpreter-o0-o3-owned",
		"control:evidence.core_mismatch",
		"control:backend.runtime_causality",
	}
	for _, required := range requiredControls {
		if !hasControl(result.Lanes, required) {
			return fail(protocol.StatusInvalid, "verify.control_missing", required)
		}
	}
	for _, lane := range result.Lanes {
		if lane.RecomputedWork == 0 {
			return fail(protocol.StatusInvalid, "verify.zero_work", lane.ID)
		}
	}
	result.Metrics.ElapsedNS = time.Since(started).Nanoseconds()
	return result.Finalize()
}

// verifyBorrowedCorpus is the Phase 3 verify path (Task 03-07-02): it
// dispatches when borrowed_view.schway is present, before the Phase 2 probe is
// considered (VerifyCorpus checks this corpus's own dispatch file first).
// Every lane uses the Phase 1 addLane shape (an explicit status on every
// path, PATTERNS I-1), so a failing lane is still returned rather than
// silently dropped the way Phase 2's verifyOwnedCorpus would drop one.
//
// Phase 11 (11-GUARD-LEDGER.md): every single-function guard in this
// function is KEPT. Each is bound to a fixed, named Phase 3 fixture
// (borrowed_view.schway, public_view*.schway) that is genuinely
// single-function by construction -- these fixtures are unrelated to the
// multi-function corpus this phase widens, and narrowing what any of them
// check would be exactly the "weaker proof reported as a wider one"
// Task 1's own prohibition forbids.
func verifyBorrowedCorpus(ctx context.Context, corpus string, runner native.Runner) protocol.Result {
	started := time.Now()
	result := protocol.New("verify", protocol.StatusPass)
	result.ExpectedEscapes = append([]string{corevalidate.KnownEscape}, originvalidate.ExpectedEscapes()...)
	addLane := func(id, status string, controls []string, work, outputBytes int, laneStarted time.Time) {
		result.Lanes = append(result.Lanes, protocol.Lane{
			Schema: protocol.LaneSchema1, ID: id, Status: status,
			Controls: append([]string{}, controls...), RecomputedWork: work,
			ElapsedNS: time.Since(laneStarted).Nanoseconds(), PeakRSSStatus: protocol.PeakRSSUnavailable, OutputBytes: outputBytes,
		})
		result.Metrics.RecomputedWork += work
		result.Metrics.OutputBytes += outputBytes
	}
	fail := func(status, code, message string) protocol.Result {
		result.Status = status
		result.Diagnostics = append(result.Diagnostics, diagnostic.Error(code, diagnostic.Span{}, message))
		result.Metrics.ElapsedNS = time.Since(started).Nanoseconds()
		return result.Finalize()
	}

	// Lane 1: exclusive-conflict, exclusive-move negative controls.
	laneStarted := time.Now()
	exclusiveControls := []struct{ file, code, control string }{
		{"exclusive_exclusive_reject.schway", "ownership.borrow_conflict", "control:ownership.exclusive_conflict"},
		{"exclusive_move_reject.schway", "ownership.move_while_borrowed", "control:ownership.exclusive_move"},
	}
	exclusiveControlNames := make([]string, 0, len(exclusiveControls))
	for _, control := range exclusiveControls {
		exclusiveControlNames = append(exclusiveControlNames, control.control)
	}
	exclusiveBytes := 0
	for _, control := range exclusiveControls {
		source, err := readBoundedFile(filepath.Join(corpus, control.file), syntax.MaxSourceBytes)
		if err != nil {
			addLane("lane:borrowed-negative-controls", "fail", nil, exclusiveBytes+1, exclusiveBytes, laneStarted)
			return fail(protocol.StatusOperational, "verify.fixture_missing", control.file)
		}
		if len(source) > syntax.MaxSourceBytes {
			return fail(protocol.StatusInvalid, "verify.fixture_input_limit", control.file)
		}
		exclusiveBytes += len(source)
		checked := Check(source)
		if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != control.code {
			addLane("lane:borrowed-negative-controls", "fail", nil, exclusiveBytes, exclusiveBytes, laneStarted)
			return fail(protocol.StatusInvalid, "verify.control_missing", control.control)
		}
	}
	addLane("lane:borrowed-negative-controls", "pass", exclusiveControlNames, len(exclusiveControls), exclusiveBytes, laneStarted)

	// Lane 2 + 3: CFG-liveness controls, reusing the already-proven,
	// independently-tested lane constructors (03-04, 03-05) rather than
	// duplicating their mutation logic here.
	branchSource, err := readBoundedFile(filepath.Join(corpus, "borrowed_view.schway"), syntax.MaxSourceBytes)
	if err != nil {
		return fail(protocol.StatusOperational, "verify.fixture_missing", "borrowed_view.schway")
	}
	if len(branchSource) > syntax.MaxSourceBytes {
		return fail(protocol.StatusInvalid, "verify.fixture_input_limit", "borrowed_view.schway")
	}
	branchChecked := Check(branchSource)
	if len(branchChecked.Diagnostics) != 0 || len(branchChecked.Program.Functions) != 1 {
		return fail(protocol.StatusInvalid, "verify.borrowed_invalid", "borrowed view fixture is invalid")
	}
	branchValidated := corevalidate.Validate(branchChecked.Program)
	if !branchValidated.Valid {
		return fail(protocol.StatusInvalid, "verify.borrowed_invalid", "borrowed view fixture failed independent validation")
	}
	honestBranch := branchValidated.Program()

	endpointLane := BorrowedLoanEndpointControlLane(honestBranch)
	result.Lanes = append(result.Lanes, endpointLane)
	result.Metrics.RecomputedWork += endpointLane.RecomputedWork
	result.Metrics.OutputBytes += endpointLane.OutputBytes
	if endpointLane.Status != "pass" {
		return fail(protocol.StatusInvalid, "verify.control_missing", "control:core.loan_endpoint_mismatch")
	}

	oracleLane := PathOracleDisagreementLane(honestBranch)
	result.Lanes = append(result.Lanes, oracleLane)
	result.Metrics.RecomputedWork += oracleLane.RecomputedWork
	result.Metrics.OutputBytes += oracleLane.OutputBytes
	if oracleLane.Status != "pass" {
		return fail(protocol.StatusInvalid, "verify.control_missing", "control:cfg.path_oracle_disagreement")
	}

	// Lane 4: the three OWN-04 origin controls (understated, impossible,
	// stale-summary), mutating honestly-checked public_view* fixtures the
	// same way originvalidate's own falsifiers do (check.go's honest
	// producer can never construct any of the three dishonest shapes
	// itself, per D-09's mutation-kill precedent).
	laneStarted = time.Now()
	originWork := 0
	originBytes := 0

	understated, err := readBoundedFile(filepath.Join(corpus, "public_view_understated.schway"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork+1, originBytes, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "public_view_understated.schway")
	}
	originBytes += len(understated)
	understatedChecked := Check(understated)
	if len(understatedChecked.Diagnostics) != 0 || len(understatedChecked.Program.Functions) != 1 || understatedChecked.Program.Functions[0].PublicOrigin == nil {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork+1, originBytes, laneStarted)
		return fail(protocol.StatusInvalid, "verify.borrowed_invalid", "understated origin fixture is invalid")
	}
	understatedProgram := understatedChecked.Program
	understatedProgram.Functions[0].PublicOrigin = &core.PublicOrigin{Paths: []string{}, Access: understatedProgram.Functions[0].PublicOrigin.Access}
	originWork++
	if problems := originvalidate.ValidatePublished(understatedProgram); len(problems) != 1 || problems[0].Code != "core.origin_understated" {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork, originBytes, laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", "control:origin.understated_summary")
	}

	impossible, err := readBoundedFile(filepath.Join(corpus, "public_view_impossible.schway"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork+1, originBytes, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "public_view_impossible.schway")
	}
	originBytes += len(impossible)
	impossibleChecked := Check(impossible)
	if len(impossibleChecked.Diagnostics) != 0 || len(impossibleChecked.Program.Functions) != 1 || impossibleChecked.Program.Functions[0].PublicOrigin == nil {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork+1, originBytes, laneStarted)
		return fail(protocol.StatusInvalid, "verify.borrowed_invalid", "impossible access fixture is invalid")
	}
	impossibleProgram := impossibleChecked.Program
	impossibleProgram.Functions[0].PublicOrigin = &core.PublicOrigin{Paths: impossibleProgram.Functions[0].PublicOrigin.Paths, Access: "exclusive"}
	originWork++
	if problems := originvalidate.ValidatePublished(impossibleProgram); len(problems) != 1 || problems[0].Code != "core.origin_access_mismatch" {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork, originBytes, laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", "control:origin.impossible_summary")
	}

	honestView, err := readBoundedFile(filepath.Join(corpus, "public_view.schway"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork+1, originBytes, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "public_view.schway")
	}
	originBytes += len(honestView)
	honestViewChecked := Check(honestView)
	if len(honestViewChecked.Diagnostics) != 0 || len(honestViewChecked.Program.Functions) != 1 {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork+1, originBytes, laneStarted)
		return fail(protocol.StatusInvalid, "verify.borrowed_invalid", "public view fixture is invalid")
	}
	summary, err := originvalidate.BuildInterface(honestViewChecked.Program)
	if err != nil {
		return fail(protocol.StatusOperational, "verify.evidence_build_failed", "unable to build interface summary")
	}
	summaryBytes, err := json.Marshal(summary)
	if err != nil {
		return fail(protocol.StatusOperational, "verify.evidence_build_failed", "unable to marshal interface summary")
	}
	realCoreBytes, err := json.Marshal(honestViewChecked.Program)
	if err != nil {
		return fail(protocol.StatusOperational, "verify.evidence_build_failed", "unable to marshal core artifact")
	}
	staleCoreBytes := append(append([]byte(nil), realCoreBytes...), '/', '/', 's', 't', 'a', 'l', 'e')
	originWork++
	if _, checkErr := originvalidate.CheckSummary(summaryBytes, staleCoreBytes); checkErr == nil {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork, originBytes, laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", "control:origin.stale_summary")
	} else {
		var originError *originvalidate.Error
		if !errors.As(checkErr, &originError) || originError.Code != "origin.stale_summary" {
			addLane("lane:borrowed-origin-controls", "fail", nil, originWork, originBytes, laneStarted)
			return fail(protocol.StatusInvalid, "verify.control_missing", "control:origin.stale_summary")
		}
	}

	omitted, err := readBoundedFile(filepath.Join(corpus, "public_view_omitted.schway"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork+1, originBytes, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "public_view_omitted.schway")
	}
	originBytes += len(omitted)
	omittedChecked := Check(omitted)
	if len(omittedChecked.Diagnostics) != 0 || len(omittedChecked.Program.Functions) != 1 {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork+1, originBytes, laneStarted)
		return fail(protocol.StatusInvalid, "verify.borrowed_invalid", "omitted origin fixture is invalid")
	}
	originWork++
	if problems := originvalidate.ValidatePublished(omittedChecked.Program); len(problems) != 1 || problems[0].Code != "core.origin_omitted" {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork, originBytes, laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", "control:origin.omitted_summary")
	}

	// control:origin.mixed_access_chain needs no mutation injection, unlike
	// the understated and impossible controls above: check.go's honest
	// producer constructs the mismatching declaration itself (a mixed
	// shared/exclusive reborrow chain whose declared access the body cannot
	// support), which is precisely why the defect reached the shipped
	// binary (03-REVIEW.md CR-01, closed by 03-08).
	mixedAccess, err := readBoundedFile(filepath.Join(corpus, "public_view_mixed_access.schway"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork+1, originBytes, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "public_view_mixed_access.schway")
	}
	originBytes += len(mixedAccess)
	mixedAccessChecked := Check(mixedAccess)
	if len(mixedAccessChecked.Diagnostics) != 0 || len(mixedAccessChecked.Program.Functions) != 1 {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork+1, originBytes, laneStarted)
		return fail(protocol.StatusInvalid, "verify.borrowed_invalid", "mixed access chain fixture is invalid")
	}
	originWork++
	if problems := originvalidate.ValidatePublished(mixedAccessChecked.Program); len(problems) != 1 || problems[0].Code != "core.origin_access_mismatch" {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork, originBytes, laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", "control:origin.mixed_access_chain")
	}

	// control:origin.multi_arm_omitted (Task 03-10-03, closing
	// 03-VERIFICATION.md's SC3/SC4 multi-arm gap): like
	// control:origin.mixed_access_chain, no mutation injection is needed —
	// check.go's honest producer constructs this shape itself. A match-bodied
	// function whose first arm returns owned and whose second arm returns a
	// live shared borrow, with no declared origin (match functions cannot
	// declare one), must be refused publication with core.origin_omitted.
	multiArmOmitted, err := readBoundedFile(filepath.Join(corpus, "public_view_multi_arm_omitted.schway"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork+1, originBytes, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "public_view_multi_arm_omitted.schway")
	}
	originBytes += len(multiArmOmitted)
	multiArmOmittedChecked := Check(multiArmOmitted)
	if len(multiArmOmittedChecked.Diagnostics) != 0 || len(multiArmOmittedChecked.Program.Functions) != 1 {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork+1, originBytes, laneStarted)
		return fail(protocol.StatusInvalid, "verify.borrowed_invalid", "multi-arm omitted origin fixture is invalid")
	}
	originWork++
	if problems := originvalidate.ValidatePublished(multiArmOmittedChecked.Program); len(problems) != 1 || problems[0].Code != "core.origin_omitted" {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork, originBytes, laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", "control:origin.multi_arm_omitted")
	}

	// control:origin.multi_arm_access_conflict (Task 03-10-03): the
	// two-disagreeing-arms shape. The frontend has no spelling for a per-arm
	// origin declaration on a match-bodied function, so this control injects
	// a declared origin onto the checked function the same way the
	// understated and impossible controls above inject their mutations, to
	// exercise the DECLARED-and-conflicting path specifically; the
	// undeclared form of the same fixture is already covered by
	// TestMultiArmAccessConflictRejectedWhenDeclaredShared/... in
	// originvalidate_test.go (unit tests, not this gate).
	multiArmConflict, err := readBoundedFile(filepath.Join(corpus, "public_view_multi_arm_access_conflict.schway"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork+1, originBytes, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "public_view_multi_arm_access_conflict.schway")
	}
	originBytes += len(multiArmConflict)
	multiArmConflictChecked := Check(multiArmConflict)
	if len(multiArmConflictChecked.Diagnostics) != 0 || len(multiArmConflictChecked.Program.Functions) != 1 {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork+1, originBytes, laneStarted)
		return fail(protocol.StatusInvalid, "verify.borrowed_invalid", "multi-arm access conflict fixture is invalid")
	}
	multiArmConflictProgram := multiArmConflictChecked.Program
	multiArmConflictProgram.Functions[0].PublicOrigin = &core.PublicOrigin{
		Paths: []string{multiArmConflictProgram.Functions[0].Parameter.Name}, Access: "shared",
	}
	originWork++
	if problems := originvalidate.ValidatePublished(multiArmConflictProgram); len(problems) != 1 || problems[0].Code != "core.origin_access_mismatch" {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork, originBytes, laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", "control:origin.multi_arm_access_conflict")
	}

	addLane("lane:borrowed-origin-controls", "pass", []string{
		"control:origin.understated_summary", "control:origin.impossible_summary", "control:origin.stale_summary",
		"control:origin.omitted_summary", "control:origin.mixed_access_chain",
		"control:origin.multi_arm_omitted", "control:origin.multi_arm_access_conflict",
	}, originWork, originBytes, laneStarted)

	requiredControls := []string{
		"control:ownership.exclusive_conflict",
		"control:ownership.exclusive_move",
		"control:core.loan_endpoint_mismatch",
		"control:cfg.path_oracle_disagreement",
		"control:origin.understated_summary",
		"control:origin.impossible_summary",
		"control:origin.stale_summary",
		"control:origin.omitted_summary",
		"control:origin.mixed_access_chain",
		"control:origin.multi_arm_omitted",
		"control:origin.multi_arm_access_conflict",
	}
	for _, required := range requiredControls {
		if !hasControl(result.Lanes, required) {
			return fail(protocol.StatusInvalid, "verify.control_missing", required)
		}
	}
	for _, lane := range result.Lanes {
		if lane.RecomputedWork == 0 {
			return fail(protocol.StatusInvalid, "verify.zero_work", lane.ID)
		}
	}
	result.Metrics.ElapsedNS = time.Since(started).Nanoseconds()
	return result.Finalize()
}

func executionpkgEqual(left, right interp.Execution) bool { return execution.Equal(left, right) }

func VerifyCorpusFile(ctx context.Context, corpus string, runner native.Runner) protocol.Result {
	return VerifyCorpus(ctx, corpus, runner, VerifyOptions{})
}

// BorrowedLoanEndpointControlLane is Phase 3's required negative control for
// T-03-06/T-03-14: a validated branch program's own declared LoanEndpoints
// is mutated by dropping one entry (the strongest of the three variants
// TestLoanEndpointMutationMatrix proves independently — an omitted endpoint
// is caught by corevalidate.recomputeLoanEndpoints with no other structural
// invariant available to catch it first, unlike a moved or invented
// endpoint, which a referential-closure check could plausibly also reject)
// and re-validated; the control passes only when that mutation is rejected
// with exactly core.loan_endpoint_mismatch.
//
// This uses the Phase 1 VerifyCorpus addLane shape (an explicit status on
// every path) rather than Phase 2 verifyOwnedCorpus's shape (a hardcoded
// "pass" status that drops the lane entirely on any failure path —
// PATTERNS' inconsistency I-1): a failing control here is still returned as
// a lane with status "fail" and nonzero work, never silently absent.
//
// Of the five control-wiring points the repo requires (fixture/in-process
// mutation, control table row, lane with nonzero work, required-controls
// entry, CLI assertion through the shipped binary), this function and its
// tests cover the first three in-process. The Phase 3 corpus's
// required-controls entry and the shipped-CLI assertion are completed in
// 03-07, where the Phase 3 verify path and gate script are assembled.
func BorrowedLoanEndpointControlLane(honest core.Program) protocol.Lane {
	started := time.Now()
	fail := func(work int) protocol.Lane {
		return protocol.Lane{
			Schema: protocol.LaneSchema1, ID: "lane:borrowed-loan-endpoint-control", Status: "fail",
			Controls: nil, RecomputedWork: work, ElapsedNS: time.Since(started).Nanoseconds(), PeakRSSStatus: protocol.PeakRSSUnavailable,
		}
	}

	baseline := corevalidate.Validate(honest)
	if !baseline.Valid {
		return fail(baseline.Checks)
	}
	mutated := baseline.Program()
	if len(mutated.Functions) == 0 || mutated.Functions[0].Linear == nil || len(mutated.Functions[0].Linear.LoanEndpoints) == 0 {
		return fail(baseline.Checks) // nothing to drop -- the honest program carries no loan to control against
	}
	mutated.Functions[0].Linear.LoanEndpoints = mutated.Functions[0].Linear.LoanEndpoints[1:]
	result := corevalidate.Validate(mutated)
	if result.Valid || len(result.Problems) == 0 || result.Problems[0].Code != "core.loan_endpoint_mismatch" {
		return fail(baseline.Checks + result.Checks)
	}
	return protocol.Lane{
		Schema: protocol.LaneSchema1, ID: "lane:borrowed-loan-endpoint-control", Status: "pass",
		Controls: []string{"control:core.loan_endpoint_mismatch"}, RecomputedWork: baseline.Checks + result.Checks,
		ElapsedNS: time.Since(started).Nanoseconds(), PeakRSSStatus: protocol.PeakRSSUnavailable,
	}
}

// PathOracleDisagreementLane is Phase 3's third-mechanism verify lane
// (03-05, ROADMAP criterion 1). pathoracle.RecomputeEndpoints independently
// recomputes a checked branch function's loan endpoints straight from its
// declared Blocks/Edges/Operations -- it never reads the function's own
// recorded Linear.LoanEndpoints field at all, so it always reports the TRUE
// endpoint set regardless of what is stored there. This lane proves that
// independence is load-bearing: it first confirms the oracle agrees with an
// HONEST function's recorded endpoints (the ROADMAP criterion 1
// differential itself), then mutates the recorded set (dropping one entry,
// mirroring BorrowedLoanEndpointControlLane's own precedent) and confirms
// the oracle's freshly recomputed answer -- run again from the SAME
// Blocks/Edges/Operations, which the mutation never touched -- disagrees
// with the now-corrupted recorded set, naming the loan whose true endpoint
// the corruption altered. Uses the Phase 1 addLane shape (an explicit
// status on every path), matching BorrowedLoanEndpointControlLane and
// VerifyCorpus, never Phase 2's verifyOwnedCorpus shape that silently drops
// a failing lane.
func PathOracleDisagreementLane(honest core.Program) protocol.Lane {
	started := time.Now()
	fail := func(work int) protocol.Lane {
		return protocol.Lane{
			Schema: protocol.LaneSchema1, ID: "lane:path-oracle-disagreement", Status: "fail",
			Controls: nil, RecomputedWork: work, ElapsedNS: time.Since(started).Nanoseconds(), PeakRSSStatus: protocol.PeakRSSUnavailable,
		}
	}

	var branchFunction core.Function
	found := false
	for _, function := range honest.Functions {
		if function.Linear != nil && len(function.Linear.LoanEndpoints) > 0 {
			branchFunction = function
			found = true
			break
		}
	}
	if !found {
		return fail(0) // nothing to disagree about -- the honest program carries no branch loan
	}

	recomputed, work, err := pathoracle.RecomputeEndpoints(branchFunction, nil)
	if err != nil {
		return fail(work)
	}
	honestEndpoints := append([]core.LoanEndpoint(nil), branchFunction.Linear.LoanEndpoints...)
	sort.Slice(honestEndpoints, func(i, j int) bool { return honestEndpoints[i].ID < honestEndpoints[j].ID })
	if !reflect.DeepEqual(recomputed, honestEndpoints) {
		return fail(work) // the oracle must AGREE with an honest, uncorrupted function first (criterion 1)
	}

	corrupted := append([]core.LoanEndpoint(nil), branchFunction.Linear.LoanEndpoints[1:]...)
	mutatedFunction := branchFunction
	mutatedFunction.Linear = &core.LinearBody{
		ID: branchFunction.Linear.ID, Types: branchFunction.Linear.Types, Places: branchFunction.Linear.Places,
		Operations: branchFunction.Linear.Operations, Blocks: branchFunction.Linear.Blocks, Edges: branchFunction.Linear.Edges,
		LoanEndpoints: corrupted,
	}
	reconfirmed, reconfirmedWork, err := pathoracle.RecomputeEndpoints(mutatedFunction, nil)
	if err != nil {
		return fail(work + reconfirmedWork)
	}
	if reflect.DeepEqual(reconfirmed, corrupted) {
		return fail(work + reconfirmedWork) // dropping an endpoint failed to produce a disagreement
	}

	return protocol.Lane{
		Schema: protocol.LaneSchema1, ID: "lane:path-oracle-disagreement", Status: "pass",
		Controls: []string{"control:cfg.path_oracle_disagreement"}, RecomputedWork: work + reconfirmedWork,
		ElapsedNS: time.Since(started).Nanoseconds(), PeakRSSStatus: protocol.PeakRSSUnavailable,
	}
}

// Phase 4's three new expected escapes (04-CONTEXT.md <deferred> "Accepted
// residual limitations", re-recorded in 04-DEBT.md D-04-31), surfaced next to
// the two carried Phase 2/3 escapes (corevalidate.KnownEscape,
// originvalidate.KnownEscape) exactly as those two packages already surface
// their own escapes: named, declared, and asserted visible under expected
// escapes -- never claimed solved, and never permitted to appear as a
// detected control (task 04-07-01).
const (
	// EscapeCoordinatedForeignBoundaryLie is T-04-07/D-04-31 item 1: a
	// coordinated edit across the frozen foreign unit, the Lang declaration,
	// and the conformance unit's own expectations passes the gate on a wrong
	// boundary. No control in this repository closes a coordinated
	// multi-artifact lie, matching corevalidate.KnownEscape's own shape one
	// layer down.
	EscapeCoordinatedForeignBoundaryLie = "escape:coordinated-foreign-boundary-lie"
	// EscapeNonlocalExitBelowThePad is T-04-33/D-04-31 item 2, first half:
	// the process-root setjmp pad cannot see a longjmp to a foreign-owned
	// jmp_buf established below it (reachable only through foreign-invoked
	// callbacks, a shape this phase's language cannot construct).
	EscapeNonlocalExitBelowThePad = "escape:nonlocal-exit-below-the-pad"
	// EscapeForeignProcessExit is T-04-33/D-04-31 item 2, second half:
	// foreign exit()/_Exit()/raise() is not observable at all -- no pad seam
	// exists for either.
	EscapeForeignProcessExit = "escape:foreign-process-exit"
)

// Phase4RequiredControls is the complete Phase 4 required-control list
// (D-04-22, task 04-07-01): every control identifier this phase introduced
// across plans 01 through 06 -- the exhaustive dispatch control, both
// foreign admission refusals, both release mutations, the layout mutation,
// the zero-attribute control, the no-release-on-defect control, the unwind
// allowlist, the nonlocal-exit detection, the terminator walk, and the
// foreign-origin refusal. This is the single session-layer source of truth
// TestPhase4RequiredControlsMatchScript compares against scripts/verify-phase4.sh's
// own text -- a control added to one and forgotten in the other fails that
// test. It deliberately excludes control:defect.signal_adjudicated, which
// verifyForeignCorpus also enforces but which 04-VALIDATION.md's Required
// Negative Controls register does not name as one of the twelve.
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

// verifyForeignCorpus is Phase 4's foreign-call gate dispatch, selected by
// VerifyCorpus on the presence of foreign_acquire_one.schway, mirroring
// verifyBorrowedCorpus's own dispatch precedent. It asserts the two
// admission refusals D-04-16/D-04-02 require are visible to the gate as
// required negative controls with honest, nonzero recomputed work, using
// the explicit-status addLane shape (Phase 1's VerifyCorpus.addLane form,
// not the Phase 2 hardcoded-"pass" shape) so a lane's own failure still
// carries partial-work evidence.
//
// Phase 11 (11-GUARD-LEDGER.md): this function's own single-function
// guard is KEPT -- it is bound to the fixed Phase 4 fixture
// defect_terminal.schway, genuinely single-function by construction and
// unrelated to the multi-function corpus this phase widens.
func verifyForeignCorpus(ctx context.Context, corpus string, runner native.Runner) protocol.Result {
	started := time.Now()
	result := protocol.New("verify", protocol.StatusPass)
	result.ExpectedEscapes = []string{EscapeCoordinatedForeignBoundaryLie, EscapeNonlocalExitBelowThePad, EscapeForeignProcessExit}
	addLane := func(id, status string, controls []string, work, outputBytes int, laneStarted time.Time) {
		result.Lanes = append(result.Lanes, protocol.Lane{
			Schema: protocol.LaneSchema1, ID: id, Status: status,
			Controls: append([]string{}, controls...), RecomputedWork: work,
			ElapsedNS: time.Since(laneStarted).Nanoseconds(), PeakRSSStatus: protocol.PeakRSSUnavailable, OutputBytes: outputBytes,
		})
		result.Metrics.RecomputedWork += work
		result.Metrics.OutputBytes += outputBytes
	}
	fail := func(status, code, message string) protocol.Result {
		result.Status = status
		result.Diagnostics = append(result.Diagnostics, diagnostic.Error(code, diagnostic.Span{}, message))
		result.Metrics.ElapsedNS = time.Since(started).Nanoseconds()
		return result.Finalize()
	}

	negativeControls := []struct {
		file    string
		code    string
		control string
		lane    string
	}{
		{"foreign_unwind_undeclared.schway", "foreign.unwind_policy_undeclared", "control:foreign.unwind_policy_undeclared", "lane:foreign-unwind-policy-undeclared"},
		{"foreign_call_target_not_foreign.schway", "core.call_target_not_foreign", "control:foreign.call_target_not_foreign", "lane:foreign-call-target-not-foreign"},
	}
	for _, negative := range negativeControls {
		laneStarted := time.Now()
		source, err := readBoundedFile(filepath.Join(corpus, negative.file), syntax.MaxSourceBytes)
		if err != nil {
			addLane(negative.lane, "fail", nil, 1, 0, laneStarted)
			return fail(protocol.StatusOperational, "verify.fixture_missing", negative.file)
		}
		checked := Check(source)
		if !hasDiagnostic(checked.Diagnostics, negative.code) {
			addLane(negative.lane, "fail", nil, checked.Work+1, len(source), laneStarted)
			return fail(protocol.StatusInvalid, "verify.control_not_falsified", negative.control)
		}
		addLane(negative.lane, "pass", []string{negative.control}, checked.Work+1, len(source), laneStarted)
	}

	// Lane: the tracer fixture itself must still admit cleanly -- a
	// required-control corpus that only ever exercises refusals would not
	// prove the admission gate lets a genuinely well-formed program through.
	laneStarted := time.Now()
	positiveSource, err := readBoundedFile(filepath.Join(corpus, "foreign_acquire_one.schway"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:foreign-acquire-admitted", "fail", nil, 1, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "foreign_acquire_one.schway")
	}
	positiveChecked := Check(positiveSource)
	if len(positiveChecked.Diagnostics) != 0 {
		addLane("lane:foreign-acquire-admitted", "fail", nil, positiveChecked.Work+1, len(positiveSource), laneStarted)
		return fail(protocol.StatusInvalid, "verify.fixture_rejected", "foreign_acquire_one.schway")
	}
	validated := corevalidate.Validate(positiveChecked.Program)
	if !validated.Valid {
		addLane("lane:foreign-acquire-admitted", "fail", nil, positiveChecked.Work+validated.Checks, len(positiveSource), laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", "control:foreign.call_target_not_foreign")
	}
	addLane("lane:foreign-acquire-admitted", "pass", nil, positiveChecked.Work+validated.Checks, len(positiveSource), laneStarted)

	// Lane: control:resource.release_order_transposed (D-04-07/Pitfall 1).
	// Attacks the CHECKER's materialized order: exchange the
	// ReleasesOperationID of the first two OpRelease operations in the
	// three-acquisition fixture's success block, feed the mutated core
	// directly to corevalidate (never re-running check.go, so the checker
	// itself is not what is being tested here -- the independent
	// rederivation is), and require the mismatch.
	laneStarted = time.Now()
	releaseSource, err := readBoundedFile(filepath.Join(corpus, "acquire_three_success.schway"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:release-order-transposed", "fail", nil, 1, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "acquire_three_success.schway")
	}
	releaseChecked := Check(releaseSource)
	if len(releaseChecked.Diagnostics) != 0 {
		addLane("lane:release-order-transposed", "fail", nil, releaseChecked.Work+1, len(releaseSource), laneStarted)
		return fail(protocol.StatusInvalid, "verify.fixture_rejected", "acquire_three_success.schway")
	}
	transposed, err := TransposeReleaseOrder(releaseChecked.Program)
	if err != nil {
		addLane("lane:release-order-transposed", "fail", nil, releaseChecked.Work+1, len(releaseSource), laneStarted)
		return fail(protocol.StatusOperational, "verify.control_incomplete", "release transposition mutation could not be constructed")
	}
	transposedResult := corevalidate.Validate(transposed)
	if transposedResult.Valid || transposedResult.Problems[0].Code != "core.release_order_mismatch" {
		addLane("lane:release-order-transposed", "fail", nil, releaseChecked.Work+transposedResult.Checks, len(releaseSource), laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", "control:resource.release_order_transposed")
	}
	addLane("lane:release-order-transposed", "pass", []string{"control:resource.release_order_transposed"}, releaseChecked.Work+transposedResult.Checks, len(releaseSource), laneStarted)

	// Lane: control:resource.release_omitted (D-04-07/Pitfall 2). Attacks
	// the EMITTER's own generated C, a different artifact than the
	// transposition lane above: delete one generated line bearing
	// schway:release-site (event plus runtime ledger decrement, on the same
	// line) and require the mutated program's own native run to be
	// detectably wrong -- a live resource the emitter's own runtime ledger
	// never cleared, surfaced as an invalid execution document (nonzero
	// live_resources on a "returned" outcome, which validateExecution has
	// refused since before this phase existed, D-04-08/Pitfall 3).
	laneStarted = time.Now()
	releaseValidated := corevalidate.Validate(releaseChecked.Program)
	if !releaseValidated.Valid {
		addLane("lane:release-omitted", "fail", nil, releaseChecked.Work+releaseValidated.Checks, len(releaseSource), laneStarted)
		return fail(protocol.StatusInvalid, "verify.fixture_rejected", "acquire_three_success.schway")
	}
	// RunNative's special ForeignSources wiring only fires for a bare
	// native.Runner value; this lane's fixture is foreign-shaped, so the
	// frozen foreign TU must be linked in explicitly before wrapping.
	foreignRunner := runner
	foreignRunner.ForeignSources = append(append([]string(nil), foreignRunner.ForeignSources...), native.ForeignLegacyResourceSourcePaths()...)
	omissionRunner := NewReleaseOmissionMutationRunner(foreignRunner)
	_, _, omissionErr := runNative(ctx, releaseSource, omissionRunner, "testdata/phase4/acquire_three_success.schway")
	if omissionErr == nil {
		addLane("lane:release-omitted", "fail", nil, releaseChecked.Work+len(omissionRunner.Optimizations()), len(releaseSource), laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", "control:resource.release_omitted")
	}
	if Phase16M004RefusalFamily(omissionErr) == "foreign" {
		addLane("lane:release-omitted", "refused", nil, releaseChecked.Work+releaseValidated.Checks, len(releaseSource), laneStarted)
		return fail(protocol.StatusOperational, "verify.m004_refusal", "foreign")
	}
	var releaseToolError *native.ToolError
	omissionDetected := errors.As(omissionErr, &releaseToolError) && releaseToolError.Code == "native.invalid_execution"
	if !omissionDetected {
		addLane("lane:release-omitted", "fail", nil, releaseChecked.Work+len(omissionRunner.Optimizations()), len(releaseSource), laneStarted)
		return fail(protocol.StatusOperational, "verify.control_incomplete", "release omission mutation did not surface as an invalid execution document")
	}
	addLane("lane:release-omitted", "pass", []string{"control:resource.release_omitted"}, releaseChecked.Work+len(omissionRunner.Optimizations())+1, len(releaseSource), laneStarted)

	// Lane: control:foreign.layout_mismatch (D-04-11/Pitfall 2, task
	// 04-03-03). Attacks the FROZEN fixture (testdata/phase4/
	// foreign_layout_mismatch.golden.c), a different artifact than either
	// release-mutation lane above: the generated conformance unit compiled
	// against it must be refused at compile time under the existing
	// -Werror flag set.
	laneStarted = time.Now()
	layoutRunner := LayoutMutationRunner{
		Runner: runner, Contract: LayoutProbeContract(),
		FixturePath: filepath.Join(corpus, "foreign_layout_mismatch.golden.c"),
	}
	layoutErr := layoutRunner.Run(ctx)
	var layoutToolError *native.ToolError
	layoutRefused := errors.As(layoutErr, &layoutToolError) && layoutToolError.Code == "native.conformance_failed"
	if !layoutRefused {
		addLane("lane:foreign-layout-mismatch", "fail", nil, 1, 0, laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", "control:foreign.layout_mismatch")
	}
	addLane("lane:foreign-layout-mismatch", "pass", []string{"control:foreign.layout_mismatch"}, 1, 0, laneStarted)

	// Lane: control:foreign.no_unproven_attributes (D-04-13, task 04-03-03).
	// Scans every emitted C artifact for the tracer and release fixtures --
	// the compiled program, the generated header, and the generated
	// conformance unit (D-04-12's three named inspectable layers) -- plus
	// their lang.foreign/0 sidecar manifests, for a banned optimizer-visible
	// attribute token, and requires the manifest's emitted_attributes field
	// to be present and empty. EmitForeignHeader/EmitForeignConformance are
	// scanned here too (not just cgen.Emit's compiled-program output): a
	// banned token injected into the generated header -- the artifact a
	// human reviewer is most likely to actually read -- would otherwise go
	// completely undetected (WR-01).
	laneStarted = time.Now()
	tracerCSource, tracerErr := Phase16ControlNativeC(positiveChecked.Program, "testdata/phase4/foreign_acquire_one.schway")
	if tracerErr != nil {
		addLane("lane:foreign-no-unproven-attributes", "fail", nil, 1, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.control_incomplete", "unable to emit tracer C for attribute scan")
	}
	tracerManifest, tracerManifestErr := cgen.EmitForeignManifest(positiveChecked.Program)
	if tracerManifestErr != nil {
		addLane("lane:foreign-no-unproven-attributes", "fail", nil, 2, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.control_incomplete", "unable to emit tracer sidecar manifest for attribute scan")
	}
	tracerHeader, tracerHeaderErr := cgen.EmitForeignHeader(positiveChecked.Program)
	if tracerHeaderErr != nil {
		addLane("lane:foreign-no-unproven-attributes", "fail", nil, 3, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.control_incomplete", "unable to emit tracer header for attribute scan")
	}
	tracerConformance, tracerConformanceErr := cgen.EmitForeignConformance(positiveChecked.Program, native.ForeignResourcePrivateHeaderPath())
	if tracerConformanceErr != nil {
		addLane("lane:foreign-no-unproven-attributes", "fail", nil, 4, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.control_incomplete", "unable to emit tracer conformance unit for attribute scan")
	}
	releaseCSource, releaseCSourceErr := Phase16ControlNativeC(releaseChecked.Program, "testdata/phase4/acquire_three_success.schway")
	if releaseCSourceErr != nil {
		addLane("lane:foreign-no-unproven-attributes", "fail", nil, 5, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.control_incomplete", "unable to emit release C for attribute scan")
	}
	releaseManifest, releaseManifestErr := cgen.EmitForeignManifest(releaseChecked.Program)
	if releaseManifestErr != nil {
		addLane("lane:foreign-no-unproven-attributes", "fail", nil, 6, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.control_incomplete", "unable to emit release sidecar manifest for attribute scan")
	}
	releaseHeader, releaseHeaderErr := cgen.EmitForeignHeader(releaseChecked.Program)
	if releaseHeaderErr != nil {
		addLane("lane:foreign-no-unproven-attributes", "fail", nil, 7, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.control_incomplete", "unable to emit release header for attribute scan")
	}
	releaseConformance, releaseConformanceErr := cgen.EmitForeignConformance(releaseChecked.Program, native.ForeignResourcePrivateHeaderPath())
	if releaseConformanceErr != nil {
		addLane("lane:foreign-no-unproven-attributes", "fail", nil, 8, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.control_incomplete", "unable to emit release conformance unit for attribute scan")
	}
	if len(cgen.ScanForBannedAttributes(tracerCSource, tracerManifest, tracerHeader, tracerConformance, releaseCSource, releaseManifest, releaseHeader, releaseConformance)) != 0 {
		addLane("lane:foreign-no-unproven-attributes", "fail", nil, 8, 0, laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", "control:foreign.no_unproven_attributes")
	}
	if !strings.Contains(tracerManifest, `"emitted_attributes":[]`) {
		addLane("lane:foreign-no-unproven-attributes", "fail", nil, 8, 0, laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", "control:foreign.no_unproven_attributes")
	}
	addLane("lane:foreign-no-unproven-attributes", "pass", []string{"control:foreign.no_unproven_attributes"}, 8, len(tracerCSource)+len(tracerHeader)+len(tracerConformance)+len(releaseCSource)+len(releaseHeader)+len(releaseConformance), laneStarted)

	// Lane: control:foreign.unwind_forbidden (D-04-19, task 04-05-02).
	// Compiles the tracer program (linked against the frozen foreign TU
	// exactly as RunNative does) and runs the real `nm -u` allowlist control
	// against the linked binary itself -- a different artifact than the
	// attribute scan above (source text) or any execution document (runtime
	// behavior): the LINKED BINARY's own undefined-symbol table.
	laneStarted = time.Now()
	symbolRunner := runner
	symbolRunner.ForeignSources = append(append([]string(nil), symbolRunner.ForeignSources...), native.ForeignLegacyResourceSourcePaths()...)
	symbolBinaryPath, symbolCleanup, symbolCompileErr := symbolRunner.CompileOnly(ctx, tracerCSource, "-O0")
	if symbolCompileErr != nil {
		addLane("lane:foreign-unwind-forbidden", "fail", nil, 1, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.control_incomplete", "unable to compile tracer binary for the undefined-symbol allowlist control")
	}
	symbolStatus, symbolRejected, symbolToolErr := native.CheckUndefinedSymbolAllowlist(ctx, "", symbolBinaryPath, 0)
	symbolCleanup()
	if symbolToolErr != nil {
		addLane("lane:foreign-unwind-forbidden", "fail", nil, 2, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.control_incomplete", fmt.Sprintf("undefined-symbol allowlist control could not run: %v", symbolToolErr))
	}
	if symbolStatus != native.SymbolsPass {
		addLane("lane:foreign-unwind-forbidden", "fail", nil, 3, 0, laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", fmt.Sprintf("control:foreign.unwind_forbidden (rejected symbols: %+v)", symbolRejected))
	}
	addLane("lane:foreign-unwind-forbidden", "pass", []string{"control:foreign.unwind_forbidden"}, 3, 0, laneStarted)

	// Lane: control:defect.no_release_on_defect (D-04-18, task 04-04-03).
	// Runs the real interpreter over the shipped defect witness and requires
	// its own engine-produced event stream to carry zero resource.released
	// events for a defect outcome, then mutation-kills the control with a
	// hand-constructed execution document that DOES carry one -- an artifact
	// mutation in the same family as LayoutMutationRunner (attacking the
	// artifact this control inspects), never the interpreter's own source.
	laneStarted = time.Now()
	defectSource, err := readBoundedFile(filepath.Join(corpus, "defect_terminal.schway"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:defect-no-release", "fail", nil, 1, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "defect_terminal.schway")
	}
	defectChecked := Check(defectSource)
	if len(defectChecked.Diagnostics) != 0 {
		addLane("lane:defect-no-release", "fail", nil, defectChecked.Work+1, len(defectSource), laneStarted)
		return fail(protocol.StatusInvalid, "verify.fixture_rejected", "defect_terminal.schway")
	}
	if len(defectChecked.Program.Functions) != 1 || defectChecked.Program.Functions[0].Match == nil {
		addLane("lane:defect-no-release", "fail", nil, defectChecked.Work+1, len(defectSource), laneStarted)
		return fail(protocol.StatusOperational, "verify.control_incomplete", "defect fixture has an unsupported shape")
	}
	defectFunction := defectChecked.Program.Functions[0]
	work := defectChecked.Work
	defectVerdicts := 0
	var abortPattern, cleanPattern string
	for _, arm := range defectFunction.Match.Arms {
		armExecution, armErr := interp.Run(defectChecked.Program, defectFunction.Name, arm.Pattern)
		if armErr != nil {
			addLane("lane:defect-no-release", "fail", nil, work+1, len(defectSource), laneStarted)
			return fail(protocol.StatusOperational, "verify.control_incomplete", "interpreter run failed for defect fixture")
		}
		work++
		if armExecution.Outcome.Kind == "defect" {
			defectVerdicts++
			abortPattern = arm.Pattern
			if !DefectHasNoReleaseAfter(armExecution) {
				addLane("lane:defect-no-release", "fail", nil, work, len(defectSource), laneStarted)
				return fail(protocol.StatusInvalid, "verify.control_missing", "control:defect.no_release_on_defect")
			}
		} else {
			cleanPattern = arm.Pattern
		}
	}
	if defectVerdicts == 0 || abortPattern == "" || cleanPattern == "" {
		addLane("lane:defect-no-release", "fail", nil, work, len(defectSource), laneStarted)
		return fail(protocol.StatusOperational, "verify.control_incomplete", "defect fixture needs both a returning and a defecting arm")
	}
	mutatedDefect := execution.Execution{
		Schema: execution.Schema1, Outcome: execution.Outcome{Kind: "defect"},
		Events: []execution.Event{
			{Schema: execution.Schema1, ID: "mutated:event:0", Kind: "function.defected", FunctionID: defectFunction.ID, SourcePlace: "place:0", TypeID: "type:0", Output: "mutated"},
			{Schema: execution.Schema1, ID: "mutated:event:1", Kind: "resource.released", FunctionID: defectFunction.ID, SourcePlace: "place:0", TypeID: "type:0"},
		},
		LiveResources: []string{},
	}
	if DefectHasNoReleaseAfter(mutatedDefect) {
		addLane("lane:defect-no-release", "fail", nil, work+1, len(defectSource), laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_not_falsified", "control:defect.no_release_on_defect")
	}
	addLane("lane:defect-no-release", "pass", []string{"control:defect.no_release_on_defect"}, work+1, len(defectSource), laneStarted)

	// Lane: control:defect.signal_adjudicated (D-04-24, task 04-04-03).
	// Compiles and runs the shipped defect witness through the REAL
	// clang/exec toolchain under both terminal shapes its two arms produce:
	// the ordinary returning arm (Expect=ExpectValue, unaffected) and the
	// aborting arm (Expect=ExpectDefect), requiring the aborting process's
	// SIGABRT termination be adjudicated through
	// ProcessState.Sys().(syscall.WaitStatus) rather than a hardcoded exit
	// code, and its terminal record decode as a genuine defect outcome.
	laneStarted = time.Now()
	defectCSource, defectCSourceErr := cgen.EmitNative(defectChecked.Program)
	if defectCSourceErr != nil {
		addLane("lane:defect-signal-adjudicated", "fail", nil, 1, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.control_incomplete", "unable to emit defect fixture C")
	}
	cleanRunner := runner
	cleanRunner.Expect = native.ExpectValue
	if _, cleanErr := cleanRunner.Run(ctx, defectCSource, "-O0", []string{cleanPattern}); cleanErr != nil {
		addLane("lane:defect-signal-adjudicated", "fail", nil, 2, len(defectCSource), laneStarted)
		return fail(protocol.StatusOperational, "verify.control_incomplete", "native run of the returning arm failed")
	}
	abortRunner := runner
	abortRunner.Expect = native.ExpectDefect
	abortResult, abortErr := abortRunner.Run(ctx, defectCSource, "-O0", []string{abortPattern})
	if abortErr != nil {
		addLane("lane:defect-signal-adjudicated", "fail", nil, 3, len(defectCSource), laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", "control:defect.signal_adjudicated")
	}
	if len(abortResult.Pairs) != 1 || abortResult.Pairs[0].Execution.Outcome.Kind != "defect" {
		addLane("lane:defect-signal-adjudicated", "fail", nil, 3, len(defectCSource), laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", "control:defect.signal_adjudicated")
	}
	addLane("lane:defect-signal-adjudicated", "pass", []string{"control:defect.signal_adjudicated"}, 3, len(defectCSource), laneStarted)

	// Lane: control:foreign.nonlocal_exit_undetected (D-04-17/D-04-21, task
	// 04-05-03). First proves the POSITIVE case through the real generated
	// pipeline: the shipped nonlocal-exit probe, compiled and linked against
	// its own frozen foreign TU, reaches the process-root pad and reports
	// exactly its one live acquisition as leaked. Then mutation-kills the
	// control TWICE, attacking the emitter's own generated C (a different
	// artifact than any execution document) in the same family as
	// ReleaseOmissionMutationRunner: once by removing the pad installation
	// entirely, and once by dropping one ledger-population site so the leak
	// count understates the true live set -- the second failure is checked
	// SPECIFICALLY for a leak-count disagreement, not merely "some
	// difference," per D-09/D-04-21.
	laneStarted = time.Now()
	nonlocalSource, err := readBoundedFile(filepath.Join(corpus, "nonlocal_exit_probe.schway"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:nonlocal-exit-undetected", "fail", nil, 1, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "nonlocal_exit_probe.schway")
	}
	nonlocalChecked := Check(nonlocalSource)
	if len(nonlocalChecked.Diagnostics) != 0 {
		addLane("lane:nonlocal-exit-undetected", "fail", nil, nonlocalChecked.Work+1, len(nonlocalSource), laneStarted)
		return fail(protocol.StatusInvalid, "verify.fixture_rejected", "nonlocal_exit_probe.schway")
	}
	nonlocalCSource, nonlocalCSourceErr := Phase16ControlNativeC(nonlocalChecked.Program, "testdata/phase4/nonlocal_exit_probe.schway")
	if nonlocalCSourceErr != nil {
		addLane("lane:nonlocal-exit-undetected", "fail", nil, nonlocalChecked.Work+2, len(nonlocalSource), laneStarted)
		return fail(protocol.StatusOperational, "verify.control_incomplete", "unable to emit nonlocal-exit probe C")
	}
	nonlocalRunner := runner
	nonlocalRunner.Expect = native.ExpectDefect
	nonlocalRunner.ForeignSources = append(append([]string(nil), nonlocalRunner.ForeignSources...), native.ForeignLegacyNonlocalSourcePaths()...)
	goldenResult, goldenErr := nonlocalRunner.Run(ctx, nonlocalCSource, "-O0", []string{"7"})
	if goldenErr != nil || len(goldenResult.Pairs) != 1 {
		addLane("lane:nonlocal-exit-undetected", "fail", nil, nonlocalChecked.Work+3, len(nonlocalCSource), laneStarted)
		return fail(protocol.StatusOperational, "verify.control_incomplete", "golden nonlocal-exit probe run failed")
	}
	golden := goldenResult.Pairs[0].Execution
	goldenLeaks := countEventKind(golden.Events, "resource.leaked")
	if golden.Outcome.Kind != "defect" || goldenLeaks == 0 || len(golden.LiveResources) != goldenLeaks {
		addLane("lane:nonlocal-exit-undetected", "fail", nil, nonlocalChecked.Work+4, len(nonlocalCSource), laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", "control:foreign.nonlocal_exit_undetected")
	}
	padMutationRunner := NewNonlocalPadOmissionMutationRunner(nonlocalRunner)
	_, padMutationErr := padMutationRunner.Run(ctx, nonlocalCSource, "-O0", []string{"7"})
	if padMutationErr == nil {
		addLane("lane:nonlocal-exit-undetected", "fail", nil, nonlocalChecked.Work+5, len(nonlocalCSource), laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_not_falsified", "control:foreign.nonlocal_exit_undetected (pad omission not detected)")
	}
	ledgerMutationRunner := NewNonlocalLedgerOmissionMutationRunner(nonlocalRunner)
	mutatedResult, ledgerMutationErr := ledgerMutationRunner.Run(ctx, nonlocalCSource, "-O0", []string{"7"})
	leakCountDisagrees := false
	if ledgerMutationErr == nil && len(mutatedResult.Pairs) == 1 {
		mutatedLeaks := countEventKind(mutatedResult.Pairs[0].Execution.Events, "resource.leaked")
		leakCountDisagrees = mutatedLeaks != goldenLeaks || len(mutatedResult.Pairs[0].Execution.LiveResources) != len(golden.LiveResources)
	} else {
		// Losing the run entirely (rather than merely under-reporting) is
		// also a valid falsification -- the control's own required-behavior
		// is "an omitted population site must never be silently accepted."
		leakCountDisagrees = true
	}
	if !leakCountDisagrees {
		addLane("lane:nonlocal-exit-undetected", "fail", nil, nonlocalChecked.Work+6, len(nonlocalCSource), laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_not_falsified", "control:foreign.nonlocal_exit_undetected (ledger population omission did not surface as a leak-count disagreement)")
	}
	addLane("lane:nonlocal-exit-undetected", "pass", []string{"control:foreign.nonlocal_exit_undetected"}, nonlocalChecked.Work+6, len(nonlocalCSource), laneStarted)

	// Lane: control:terminator.walk_incomplete (D-04-29, task 04-06-01). The
	// single highest-risk item in the phase: asserts, per package, that the
	// set of core.OperationKind values originvalidate/pathoracle each
	// recognise as a terminator equals core.TerminatorKinds() exactly --
	// naming the missing member when it does not, rather than inferring
	// completeness from a differential that only compares paths the oracle
	// itself enumerated.
	laneStarted = time.Now()
	terminatorWork := 0
	terminatorMissing := ""
	for _, terminator := range core.TerminatorKinds() {
		terminatorWork++
		if !originvalidate.RecognizesTerminator(terminator) {
			terminatorMissing = fmt.Sprintf("originvalidate does not recognise terminator %q", terminator)
			break
		}
		terminatorWork++
		if !pathoracle.RecognizesTerminator(terminator) {
			terminatorMissing = fmt.Sprintf("pathoracle does not recognise terminator %q", terminator)
			break
		}
	}
	if terminatorMissing != "" {
		addLane("lane:terminator-walk-complete", "fail", nil, terminatorWork, 0, laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", "control:terminator.walk_incomplete ("+terminatorMissing+")")
	}
	addLane("lane:terminator-walk-complete", "pass", []string{"control:terminator.walk_incomplete"}, terminatorWork, 0, laneStarted)

	// Lane: control:origin.foreign_origin_omitted (D-04-28, task 04-06-02).
	// The shipped foreign_origin_omitted.schway fixture -- a foreign call
	// declared to borrow its argument, returned with no declared public
	// origin -- must be refused by originvalidate.ValidatePublished with
	// exactly that code, derived from the core artifact alone.
	laneStarted = time.Now()
	foreignOriginSource, err := readBoundedFile(filepath.Join(corpus, "foreign_origin_omitted.schway"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:foreign-origin-omitted", "fail", nil, 1, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "foreign_origin_omitted.schway")
	}
	foreignOriginChecked := Check(foreignOriginSource)
	if len(foreignOriginChecked.Diagnostics) != 0 {
		addLane("lane:foreign-origin-omitted", "fail", nil, foreignOriginChecked.Work+1, len(foreignOriginSource), laneStarted)
		return fail(protocol.StatusInvalid, "verify.fixture_rejected", "foreign_origin_omitted.schway")
	}
	foreignOriginProblems := originvalidate.ValidatePublished(foreignOriginChecked.Program)
	if len(foreignOriginProblems) != 1 || foreignOriginProblems[0].Code != "core.foreign_origin_omitted" {
		addLane("lane:foreign-origin-omitted", "fail", nil, foreignOriginChecked.Work+1, len(foreignOriginSource), laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", "control:origin.foreign_origin_omitted")
	}
	addLane("lane:foreign-origin-omitted", "pass", []string{"control:origin.foreign_origin_omitted"}, foreignOriginChecked.Work+1, len(foreignOriginSource), laneStarted)

	// Lane: control:kind.exhaustive_dispatch (D-04-22, task 04-07-01). This is
	// the session-layer, CLI-observable sibling of core_test.go's
	// TestAllOperationKindsHandledAtEverySite, which proves dispatch
	// completeness in-process across all four phases' corpora combined; this
	// lane proves the same claim is visible through `schway verify
	// testdata/phase4` itself, scoped to the four operation kinds THIS phase
	// introduced (OpForeignCall, OpRelease, OpFail, OpDefect) -- the kinds
	// Phase 1-3's own corpora cannot exercise, so this corpus is the only
	// place their dispatch completeness can be demonstrated through the
	// shipped binary. It drives every dispatch-fixture already read above
	// through corevalidate, pathoracle, originvalidate, interp, and cgen, and
	// requires all four kinds to be encountered by at least one operation.
	laneStarted = time.Now()
	dispatchFixtures := []string{"foreign_acquire_one.schway", "acquire_three_success.schway", "defect_terminal.schway", "nonlocal_exit_probe.schway"}
	encounteredKinds := make(map[core.OperationKind]bool)
	dispatchWork := 0
	for _, fixtureName := range dispatchFixtures {
		fixtureSource, fixtureErr := readBoundedFile(filepath.Join(corpus, fixtureName), syntax.MaxSourceBytes)
		if fixtureErr != nil {
			addLane("lane:kind-exhaustive-dispatch", "fail", nil, dispatchWork+1, 0, laneStarted)
			return fail(protocol.StatusOperational, "verify.fixture_missing", fixtureName)
		}
		fixtureChecked := Check(fixtureSource)
		if len(fixtureChecked.Diagnostics) != 0 {
			addLane("lane:kind-exhaustive-dispatch", "fail", nil, dispatchWork+1, len(fixtureSource), laneStarted)
			return fail(protocol.StatusInvalid, "verify.fixture_rejected", fixtureName)
		}
		dispatchWork += fixtureChecked.Work
		fixtureValidated := corevalidate.Validate(fixtureChecked.Program)
		if !fixtureValidated.Valid {
			addLane("lane:kind-exhaustive-dispatch", "fail", nil, dispatchWork+fixtureValidated.Checks, len(fixtureSource), laneStarted)
			return fail(protocol.StatusInvalid, "verify.fixture_rejected", fixtureName)
		}
		dispatchWork += fixtureValidated.Checks
		dispatchProgram := fixtureValidated.Program()
		dispatchCalleeContracts := originvalidate.BuildCalleeOriginFacts(dispatchProgram)
		for _, function := range dispatchProgram.Functions {
			if function.Linear != nil {
				for _, operation := range function.Linear.Operations {
					encounteredKinds[operation.Kind] = true
				}
				if function.Linear.ID != "" {
					if _, _, oracleErr := pathoracle.RecomputeEndpoints(function, nil); oracleErr != nil {
						addLane("lane:kind-exhaustive-dispatch", "fail", nil, dispatchWork+1, len(fixtureSource), laneStarted)
						return fail(protocol.StatusOperational, "verify.control_incomplete", "pathoracle dispatch error for "+fixtureName)
					}
				}
			}
			_ = originvalidate.RecomputeOriginPerReturn(function, dispatchCalleeContracts)
			dispatchWork++
			if function.Match != nil {
				for _, arm := range function.Match.Arms {
					if _, interpErr := interp.Run(dispatchProgram, function.Name, arm.Pattern); interpErr != nil {
						addLane("lane:kind-exhaustive-dispatch", "fail", nil, dispatchWork+1, len(fixtureSource), laneStarted)
						return fail(protocol.StatusOperational, "verify.control_incomplete", "interp dispatch error for "+fixtureName)
					}
					dispatchWork++
				}
			}
		}
		if len(dispatchProgram.Functions) == 1 {
			if _, cgenErr := Phase16ControlNativeC(dispatchProgram, "testdata/phase4/"+fixtureName); cgenErr != nil {
				addLane("lane:kind-exhaustive-dispatch", "fail", nil, dispatchWork+1, len(fixtureSource), laneStarted)
				return fail(protocol.StatusOperational, "verify.control_incomplete", "cgen dispatch error for "+fixtureName)
			}
			dispatchWork++
		}
	}
	for _, kind := range []core.OperationKind{core.OpForeignCall, core.OpRelease, core.OpFail, core.OpDefect} {
		if !encounteredKinds[kind] {
			addLane("lane:kind-exhaustive-dispatch", "fail", nil, dispatchWork+1, 0, laneStarted)
			return fail(protocol.StatusInvalid, "verify.control_missing", "control:kind.exhaustive_dispatch (kind "+string(kind)+" never encountered)")
		}
	}
	addLane("lane:kind-exhaustive-dispatch", "pass", []string{"control:kind.exhaustive_dispatch"}, dispatchWork+1, 0, laneStarted)

	requiredControls := append(append([]string(nil), Phase4RequiredControls()...), "control:defect.signal_adjudicated")
	for _, required := range requiredControls {
		if !hasControl(result.Lanes, required) {
			return fail(protocol.StatusInvalid, "verify.control_missing", required)
		}
	}
	for _, lane := range result.Lanes {
		if lane.RecomputedWork == 0 {
			return fail(protocol.StatusInvalid, "verify.zero_work", lane.ID)
		}
	}
	result.Metrics.ElapsedNS = time.Since(started).Nanoseconds()
	return result.Finalize()
}

func hasDiagnostic(diagnostics []diagnostic.Diagnostic, code string) bool {
	for _, problem := range diagnostics {
		if problem.Code == code {
			return true
		}
	}
	return false
}

func hasControl(lanes []protocol.Lane, expected string) bool {
	for _, lane := range lanes {
		for _, control := range lane.Controls {
			if control == expected {
				return true
			}
		}
	}
	return false
}
