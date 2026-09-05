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

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/debugmap"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/evidence"
	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/originvalidate"
	"github.com/codename-lang/lang/internal/compiler/pathoracle"
	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/syntax"
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
// ("LANG_BUFFER lang_value_delivered = ... op:0"), means renaming a fixture
// binding or reindenting the emitter cannot silently turn this control into
// an opaque operational failure — the seam survives both. The fail-closed
// exact-one requirement is unchanged: the marker must appear on exactly one
// line, or the control refuses to run rather than mutating an ambiguous or
// absent site.
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
const releaseMarker = "/* lang:release-site */"

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

func (r *ReleaseOmissionMutationRunner) Run(ctx context.Context, cSource, optimization string, inputs []string) (native.Result, error) {
	lines := strings.Split(cSource, "\n")
	matched := -1
	for index, line := range lines {
		if strings.Contains(line, releaseMarker) {
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
		return native.Result{}, &native.ToolError{Code: "native.backend_control_invalid", Err: fmt.Errorf("release mutation marker count is 0, want at least 1")}
	}
	mutated := strings.Join(append(append([]string(nil), lines[:matched]...), lines[matched+1:]...), "\n")
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
// lang_res_open symbol's own (single-field) Layout, specifically so the
// mutation-kill demonstration exercises a genuine field TRANSPOSITION (which
// a one-field record cannot express) without touching the byte-frozen
// production fixture at all.
func LayoutProbeContract() *core.ForeignContract {
	return &core.ForeignContract{
		Symbol: "lang_layout_probe", Allocator: "libc_malloc", Unwind: "forbidden", NonlocalExit: "forbidden", Fails: "AcquireError",
		InitializedState: "fully", Capture: "none", Retention: "none", Aliasing: "none",
		Layout: &core.RecordLayout{
			Size: 2, Alignment: 1, ForeignTypeName: "lang_foreign_layout_probe_block",
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
	} else {
		result.ModuleID = checked.Program.ModuleID
	}
	return completeCommand(result, started, checked.Work), nil
}

func RunInterpreter(source []byte) ([]interp.Execution, []diagnostic.Diagnostic, error) {
	checked := Check(source)
	if len(checked.Diagnostics) > 0 {
		return nil, checked.Diagnostics, nil
	}
	if len(checked.Program.Functions) != 1 {
		return nil, nil, os.ErrInvalid
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		return nil, nil, fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
	}
	checked.Program = validated.Program()
	inputs, ok := interpreterInputs(checked.Program)
	if !ok {
		return nil, nil, os.ErrInvalid
	}
	executions := make([]interp.Execution, 0, len(inputs))
	for _, input := range inputs {
		execution, err := interp.Run(checked.Program, checked.Program.Functions[0].Name, input)
		if err != nil {
			return nil, nil, err
		}
		executions = append(executions, execution)
	}
	return executions, nil, nil
}

func interpreterInputs(program core.Program) ([]string, bool) {
	if len(program.Functions) != 1 {
		return nil, false
	}
	function := program.Functions[0]
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
	if function.Match != nil && len(program.DataTypes) == 1 {
		// Covers both the Phase 1 bare-arm match (Linear == nil) and the
		// Phase 3 branch-shaped match whose arms carry linear bodies
		// (Linear != nil) — both dispatch on every alternative of the
		// scrutinee's declared type.
		return append([]string(nil), program.DataTypes[0].Alternatives...), true
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

func RunInterpreterCommandFile(path string) (protocol.Result, error) {
	started := time.Now()
	executions, diagnostics, err := RunInterpreterFile(path)
	if err != nil {
		result := protocol.New("run", protocol.StatusOperational)
		result.Diagnostics = []diagnostic.Diagnostic{diagnostic.Error("tool.run_failed", diagnostic.Span{}, "interpreter operation failed")}
		return completeCommand(result, started, 1), nil
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

func RunNative(ctx context.Context, source []byte, runner NativeRunner) (NativeResult, []diagnostic.Diagnostic, error) {
	checked := Check(source)
	if len(checked.Diagnostics) > 0 {
		return NativeResult{}, checked.Diagnostics, nil
	}
	if len(checked.Program.Functions) != 1 {
		return NativeResult{}, nil, os.ErrInvalid
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		return NativeResult{}, nil, fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
	}
	checked.Program = validated.Program()
	inputs, ok := interpreterInputs(checked.Program)
	if !ok {
		return NativeResult{}, nil, os.ErrInvalid
	}
	interpreted := make([]interp.Execution, 0, len(inputs))
	for _, input := range inputs {
		execution, err := interp.Run(checked.Program, checked.Program.Functions[0].Name, input)
		if err != nil {
			return NativeResult{}, nil, err
		}
		interpreted = append(interpreted, execution)
	}
	cSource, err := cgen.EmitNative(checked.Program)
	if err != nil {
		return NativeResult{}, nil, err
	}
	// A foreign-shaped function (D-04-10) needs the frozen foreign
	// translation unit linked in. This is additive wiring on a *native.Runner
	// value specifically -- a caller-supplied NativeRunner of any other
	// concrete type (e.g. a mutation runner) is passed through unmodified,
	// since none of those exercise a foreign-call program today.
	if checked.Program.Functions[0].ForeignContract != nil {
		if concrete, ok := runner.(native.Runner); ok {
			concrete.ForeignSources = append(append([]string(nil), concrete.ForeignSources...), native.ForeignResourceSourcePath())
			runner = concrete
		}
	}
	o0, err := runner.Run(ctx, cSource, "-O0", inputs)
	if err != nil {
		return NativeResult{}, nil, err
	}
	o3, err := runner.Run(ctx, cSource, "-O3", inputs)
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
		for _, actual := range []native.Result{o0, o3} {
			if !execution.Equal(expected, actual.Pairs[index].Execution) {
				expectedBytes, _ := execution.CanonicalBytes(expected)
				actualBytes, _ := execution.CanonicalBytes(actual.Pairs[index].Execution)
				return NativeResult{}, nil, &EngineMismatch{Optimization: actual.Optimization, Input: inputs[index], Expected: string(expectedBytes), Actual: string(actualBytes)}
			}
		}
	}
	return NativeResult{CSource: cSource, Interpreter: interpreted, O0: o0, O3: o3}, nil, nil
}

func RunNativeFile(ctx context.Context, path string, runner NativeRunner) (NativeResult, []diagnostic.Diagnostic, error) {
	source, err := readBoundedFile(path, syntax.MaxSourceBytes)
	if err != nil {
		return NativeResult{}, nil, err
	}
	return RunNative(ctx, source, runner)
}

func RunNativeCommandFile(ctx context.Context, path string, runner NativeRunner) (protocol.Result, error) {
	started := time.Now()
	nativeResult, diagnostics, err := RunNativeFile(ctx, path, runner)
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
		return protocol.Result{}, fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
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
// this artifact from `lang check`; this command exists so the two-invocation
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
		return protocol.Result{}, fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
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

func interfaceProjection(schema, moduleID, coreDigest string, functions []core.FunctionSignature) *protocol.InterfaceSummary {
	summary := &protocol.InterfaceSummary{Schema: schema, ModuleID: moduleID, CoreDigest: coreDigest, Functions: make([]protocol.InterfaceFunctionAnswer, 0, len(functions))}
	for _, function := range functions {
		answer := protocol.InterfaceFunctionAnswer{ID: function.ID, Name: function.Name}
		if function.PublicOrigin != nil {
			answer.Paths = function.PublicOrigin.Paths
			answer.Access = function.PublicOrigin.Access
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

func completeCommand(result protocol.Result, started time.Time, work int) protocol.Result {
	// Bounded commands keep their default projection reproducible. The verify
	// orchestrator records wall-time observations explicitly; later telemetry
	// modes can opt into per-command timing without making ordinary output churn.
	if os.Getenv("LANG_OBSERVE_TIMING") == "1" {
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
	if _, err := os.Stat(filepath.Join(corpus, "foreign_acquire_one.lang")); err == nil {
		return verifyForeignCorpus(ctx, corpus, runner)
	}
	if _, err := os.Stat(filepath.Join(corpus, "borrowed_view.lang")); err == nil {
		return verifyBorrowedCorpus(ctx, corpus, runner)
	}
	if _, err := os.Stat(filepath.Join(corpus, "owned_transfer.lang")); err == nil {
		return verifyOwnedCorpus(ctx, corpus, runner)
	}
	started := time.Now()
	result := protocol.New("verify", protocol.StatusPass)
	addLane := func(id, status string, controls []string, work, outputBytes int, laneStarted time.Time) {
		result.Lanes = append(result.Lanes, protocol.Lane{
			Schema: "lang.verify-lane/0", ID: id, Status: status,
			Controls: append([]string{}, controls...), RecomputedWork: work,
			ElapsedNS: time.Since(laneStarted).Nanoseconds(), PeakRSSStatus: "unavailable", OutputBytes: outputBytes,
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

	validPath := filepath.Join(corpus, "toggle.lang")
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
	for _, name := range []string{"toggle.lang", "comments.lang", "malformed.lang"} {
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
	negativeSource, err := readBoundedFile(filepath.Join(corpus, "non_exhaustive.lang"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:negative-controls", "fail", nil, 0, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "required negative fixture is unavailable")
	}
	if len(negativeSource) > syntax.MaxSourceBytes {
		return fail(protocol.StatusInvalid, "verify.fixture_input_limit", "non_exhaustive.lang")
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
		result.Lanes = append(result.Lanes, protocol.Lane{Schema: "lang.verify-lane/0", ID: id, Status: "pass", Controls: append([]string(nil), controls...), RecomputedWork: work, ElapsedNS: time.Since(laneStarted).Nanoseconds(), PeakRSSStatus: "unavailable", OutputBytes: outputBytes})
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
		{"use_after_move.lang", "ownership.use_after_move", "control:ownership.use_after_move"},
		{"move_while_borrowed.lang", "ownership.move_while_borrowed", "control:ownership.move_while_borrowed"},
		{"implicit_noncopy.lang", "ownership.transfer_requires_take", "control:ownership.transfer_requires_take"},
		{"reborrow_while_moved.lang", "ownership.move_while_borrowed", "control:ownership.move_while_reborrowed"},
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

	validSource, err := readBoundedFile(filepath.Join(corpus, "owned_transfer.lang"), syntax.MaxSourceBytes)
	if err != nil {
		return fail(protocol.StatusOperational, "verify.fixture_missing", "owned_transfer.lang")
	}
	if len(validSource) > syntax.MaxSourceBytes {
		return fail(protocol.StatusInvalid, "verify.fixture_input_limit", "owned_transfer.lang")
	}
	checked := Check(validSource)
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
// dispatches when borrowed_view.lang is present, before the Phase 2 probe is
// considered (VerifyCorpus checks this corpus's own dispatch file first).
// Every lane uses the Phase 1 addLane shape (an explicit status on every
// path, PATTERNS I-1), so a failing lane is still returned rather than
// silently dropped the way Phase 2's verifyOwnedCorpus would drop one.
func verifyBorrowedCorpus(ctx context.Context, corpus string, runner native.Runner) protocol.Result {
	started := time.Now()
	result := protocol.New("verify", protocol.StatusPass)
	result.ExpectedEscapes = append([]string{corevalidate.KnownEscape}, originvalidate.ExpectedEscapes()...)
	addLane := func(id, status string, controls []string, work, outputBytes int, laneStarted time.Time) {
		result.Lanes = append(result.Lanes, protocol.Lane{
			Schema: "lang.verify-lane/0", ID: id, Status: status,
			Controls: append([]string{}, controls...), RecomputedWork: work,
			ElapsedNS: time.Since(laneStarted).Nanoseconds(), PeakRSSStatus: "unavailable", OutputBytes: outputBytes,
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
		{"exclusive_exclusive_reject.lang", "ownership.borrow_conflict", "control:ownership.exclusive_conflict"},
		{"exclusive_move_reject.lang", "ownership.move_while_borrowed", "control:ownership.exclusive_move"},
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
	branchSource, err := readBoundedFile(filepath.Join(corpus, "borrowed_view.lang"), syntax.MaxSourceBytes)
	if err != nil {
		return fail(protocol.StatusOperational, "verify.fixture_missing", "borrowed_view.lang")
	}
	if len(branchSource) > syntax.MaxSourceBytes {
		return fail(protocol.StatusInvalid, "verify.fixture_input_limit", "borrowed_view.lang")
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

	understated, err := readBoundedFile(filepath.Join(corpus, "public_view_understated.lang"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork+1, originBytes, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "public_view_understated.lang")
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

	impossible, err := readBoundedFile(filepath.Join(corpus, "public_view_impossible.lang"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork+1, originBytes, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "public_view_impossible.lang")
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

	honestView, err := readBoundedFile(filepath.Join(corpus, "public_view.lang"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork+1, originBytes, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "public_view.lang")
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

	omitted, err := readBoundedFile(filepath.Join(corpus, "public_view_omitted.lang"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork+1, originBytes, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "public_view_omitted.lang")
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
	mixedAccess, err := readBoundedFile(filepath.Join(corpus, "public_view_mixed_access.lang"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork+1, originBytes, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "public_view_mixed_access.lang")
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
	multiArmOmitted, err := readBoundedFile(filepath.Join(corpus, "public_view_multi_arm_omitted.lang"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork+1, originBytes, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "public_view_multi_arm_omitted.lang")
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
	multiArmConflict, err := readBoundedFile(filepath.Join(corpus, "public_view_multi_arm_access_conflict.lang"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:borrowed-origin-controls", "fail", nil, originWork+1, originBytes, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "public_view_multi_arm_access_conflict.lang")
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
			Schema: "lang.verify-lane/0", ID: "lane:borrowed-loan-endpoint-control", Status: "fail",
			Controls: nil, RecomputedWork: work, ElapsedNS: time.Since(started).Nanoseconds(), PeakRSSStatus: "unavailable",
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
		Schema: "lang.verify-lane/0", ID: "lane:borrowed-loan-endpoint-control", Status: "pass",
		Controls: []string{"control:core.loan_endpoint_mismatch"}, RecomputedWork: baseline.Checks + result.Checks,
		ElapsedNS: time.Since(started).Nanoseconds(), PeakRSSStatus: "unavailable",
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
			Schema: "lang.verify-lane/0", ID: "lane:path-oracle-disagreement", Status: "fail",
			Controls: nil, RecomputedWork: work, ElapsedNS: time.Since(started).Nanoseconds(), PeakRSSStatus: "unavailable",
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

	recomputed, work, err := pathoracle.RecomputeEndpoints(branchFunction)
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
	reconfirmed, reconfirmedWork, err := pathoracle.RecomputeEndpoints(mutatedFunction)
	if err != nil {
		return fail(work + reconfirmedWork)
	}
	if reflect.DeepEqual(reconfirmed, corrupted) {
		return fail(work + reconfirmedWork) // dropping an endpoint failed to produce a disagreement
	}

	return protocol.Lane{
		Schema: "lang.verify-lane/0", ID: "lane:path-oracle-disagreement", Status: "pass",
		Controls: []string{"control:cfg.path_oracle_disagreement"}, RecomputedWork: work + reconfirmedWork,
		ElapsedNS: time.Since(started).Nanoseconds(), PeakRSSStatus: "unavailable",
	}
}

// verifyForeignCorpus is Phase 4's foreign-call gate dispatch, selected by
// VerifyCorpus on the presence of foreign_acquire_one.lang, mirroring
// verifyBorrowedCorpus's own dispatch precedent. It asserts the two
// admission refusals D-04-16/D-04-02 require are visible to the gate as
// required negative controls with honest, nonzero recomputed work, using
// the explicit-status addLane shape (Phase 1's VerifyCorpus.addLane form,
// not the Phase 2 hardcoded-"pass" shape) so a lane's own failure still
// carries partial-work evidence.
func verifyForeignCorpus(ctx context.Context, corpus string, runner native.Runner) protocol.Result {
	started := time.Now()
	result := protocol.New("verify", protocol.StatusPass)
	addLane := func(id, status string, controls []string, work, outputBytes int, laneStarted time.Time) {
		result.Lanes = append(result.Lanes, protocol.Lane{
			Schema: "lang.verify-lane/0", ID: id, Status: status,
			Controls: append([]string{}, controls...), RecomputedWork: work,
			ElapsedNS: time.Since(laneStarted).Nanoseconds(), PeakRSSStatus: "unavailable", OutputBytes: outputBytes,
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
		{"foreign_unwind_undeclared.lang", "foreign.unwind_policy_undeclared", "control:foreign.unwind_policy_undeclared", "lane:foreign-unwind-policy-undeclared"},
		{"foreign_call_target_not_foreign.lang", "core.call_target_not_foreign", "control:foreign.call_target_not_foreign", "lane:foreign-call-target-not-foreign"},
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
	positiveSource, err := readBoundedFile(filepath.Join(corpus, "foreign_acquire_one.lang"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:foreign-acquire-admitted", "fail", nil, 1, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "foreign_acquire_one.lang")
	}
	positiveChecked := Check(positiveSource)
	if len(positiveChecked.Diagnostics) != 0 {
		addLane("lane:foreign-acquire-admitted", "fail", nil, positiveChecked.Work+1, len(positiveSource), laneStarted)
		return fail(protocol.StatusInvalid, "verify.fixture_rejected", "foreign_acquire_one.lang")
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
	releaseSource, err := readBoundedFile(filepath.Join(corpus, "acquire_three_success.lang"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:release-order-transposed", "fail", nil, 1, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "acquire_three_success.lang")
	}
	releaseChecked := Check(releaseSource)
	if len(releaseChecked.Diagnostics) != 0 {
		addLane("lane:release-order-transposed", "fail", nil, releaseChecked.Work+1, len(releaseSource), laneStarted)
		return fail(protocol.StatusInvalid, "verify.fixture_rejected", "acquire_three_success.lang")
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
	// lang:release-site (event plus runtime ledger decrement, on the same
	// line) and require the mutated program's own native run to be
	// detectably wrong -- a live resource the emitter's own runtime ledger
	// never cleared, surfaced as an invalid execution document (nonzero
	// live_resources on a "returned" outcome, which validateExecution has
	// refused since before this phase existed, D-04-08/Pitfall 3).
	laneStarted = time.Now()
	releaseValidated := corevalidate.Validate(releaseChecked.Program)
	if !releaseValidated.Valid {
		addLane("lane:release-omitted", "fail", nil, releaseChecked.Work+releaseValidated.Checks, len(releaseSource), laneStarted)
		return fail(protocol.StatusInvalid, "verify.fixture_rejected", "acquire_three_success.lang")
	}
	// RunNative's special ForeignSources wiring only fires for a bare
	// native.Runner value; this lane's fixture is foreign-shaped, so the
	// frozen foreign TU must be linked in explicitly before wrapping.
	foreignRunner := runner
	foreignRunner.ForeignSources = append(append([]string(nil), foreignRunner.ForeignSources...), native.ForeignResourceSourcePath())
	omissionRunner := NewReleaseOmissionMutationRunner(foreignRunner)
	_, _, omissionErr := RunNative(ctx, releaseSource, omissionRunner)
	if omissionErr == nil {
		addLane("lane:release-omitted", "fail", nil, releaseChecked.Work+len(omissionRunner.Optimizations()), len(releaseSource), laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", "control:resource.release_omitted")
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
	// Scans every emitted C artifact for the tracer and release fixtures,
	// plus their lang.foreign/0 sidecar manifests, for a banned
	// optimizer-visible attribute token, and requires the manifest's
	// emitted_attributes field to be present and empty.
	laneStarted = time.Now()
	tracerCSource, tracerErr := cgen.Emit(positiveChecked.Program)
	if tracerErr != nil {
		addLane("lane:foreign-no-unproven-attributes", "fail", nil, 1, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.control_incomplete", "unable to emit tracer C for attribute scan")
	}
	tracerManifest, tracerManifestErr := cgen.EmitForeignManifest(positiveChecked.Program)
	if tracerManifestErr != nil {
		addLane("lane:foreign-no-unproven-attributes", "fail", nil, 2, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.control_incomplete", "unable to emit tracer sidecar manifest for attribute scan")
	}
	releaseCSource, releaseCSourceErr := cgen.Emit(releaseChecked.Program)
	if releaseCSourceErr != nil {
		addLane("lane:foreign-no-unproven-attributes", "fail", nil, 3, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.control_incomplete", "unable to emit release C for attribute scan")
	}
	releaseManifest, releaseManifestErr := cgen.EmitForeignManifest(releaseChecked.Program)
	if releaseManifestErr != nil {
		addLane("lane:foreign-no-unproven-attributes", "fail", nil, 4, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.control_incomplete", "unable to emit release sidecar manifest for attribute scan")
	}
	if len(cgen.ScanForBannedAttributes(tracerCSource, tracerManifest, releaseCSource, releaseManifest)) != 0 {
		addLane("lane:foreign-no-unproven-attributes", "fail", nil, 4, 0, laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", "control:foreign.no_unproven_attributes")
	}
	if !strings.Contains(tracerManifest, `"emitted_attributes":[]`) {
		addLane("lane:foreign-no-unproven-attributes", "fail", nil, 4, 0, laneStarted)
		return fail(protocol.StatusInvalid, "verify.control_missing", "control:foreign.no_unproven_attributes")
	}
	addLane("lane:foreign-no-unproven-attributes", "pass", []string{"control:foreign.no_unproven_attributes"}, 4, len(tracerCSource)+len(releaseCSource), laneStarted)

	// Lane: control:defect.no_release_on_defect (D-04-18, task 04-04-03).
	// Runs the real interpreter over the shipped defect witness and requires
	// its own engine-produced event stream to carry zero resource.released
	// events for a defect outcome, then mutation-kills the control with a
	// hand-constructed execution document that DOES carry one -- an artifact
	// mutation in the same family as LayoutMutationRunner (attacking the
	// artifact this control inspects), never the interpreter's own source.
	laneStarted = time.Now()
	defectSource, err := readBoundedFile(filepath.Join(corpus, "defect_terminal.lang"), syntax.MaxSourceBytes)
	if err != nil {
		addLane("lane:defect-no-release", "fail", nil, 1, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "defect_terminal.lang")
	}
	defectChecked := Check(defectSource)
	if len(defectChecked.Diagnostics) != 0 {
		addLane("lane:defect-no-release", "fail", nil, defectChecked.Work+1, len(defectSource), laneStarted)
		return fail(protocol.StatusInvalid, "verify.fixture_rejected", "defect_terminal.lang")
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

	requiredControls := []string{
		"control:foreign.unwind_policy_undeclared",
		"control:foreign.call_target_not_foreign",
		"control:resource.release_order_transposed",
		"control:resource.release_omitted",
		"control:foreign.layout_mismatch",
		"control:foreign.no_unproven_attributes",
		"control:defect.no_release_on_defect",
		"control:defect.signal_adjudicated",
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
