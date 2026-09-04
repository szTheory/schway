package session

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/evidence"
	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/native"
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

func executionpkgEqual(left, right interp.Execution) bool { return execution.Equal(left, right) }

func VerifyCorpusFile(ctx context.Context, corpus string, runner native.Runner) protocol.Result {
	return VerifyCorpus(ctx, corpus, runner, VerifyOptions{})
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
