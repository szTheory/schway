package session

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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
}

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

func (e *EngineMismatch) Error() string {
	return fmt.Sprintf("native %s mismatch for %s: expected %s, got %s", e.Optimization, e.Input, e.Expected, e.Actual)
}

func Check(source []byte) CheckResult {
	parsed := syntax.Parse(source)
	result := CheckResult{Tree: parsed.Tree, Diagnostics: append([]diagnostic.Diagnostic(nil), parsed.Diagnostics...)}
	if len(result.Diagnostics) > 0 {
		return result
	}
	checked := check.Program(parsed.Program)
	result.Program = checked.Program
	result.Diagnostics = append(result.Diagnostics, checked.Diagnostics...)
	return result
}

func Format(source []byte) FormatResult {
	parsed := syntax.Parse(source)
	return FormatResult{Source: append([]byte(nil), source...), Canonical: syntax.Format(parsed.Tree), Diagnostics: parsed.Diagnostics}
}

func FormatFile(path string) (FormatResult, error) {
	source, err := os.ReadFile(path)
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
	source, err := os.ReadFile(path)
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
	return completeCommand(result, started, 1), nil
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
	if function.Match != nil && function.Linear == nil && len(program.DataTypes) == 1 {
		return append([]string(nil), program.DataTypes[0].Alternatives...), true
	}
	return nil, false
}

func RunInterpreterFile(path string) ([]interp.Execution, []diagnostic.Diagnostic, error) {
	source, err := os.ReadFile(path)
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
	source, err := os.ReadFile(path)
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
	source, err := os.ReadFile(path)
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
	result.Evidence = &protocol.EvidenceSummary{Schema: evidence.Schema, ID: product.Manifest.ID, Digest: evidence.ContentDigest(product.ManifestBytes)}
	return product, completeCommand(result, started, 3), nil
}

func ValidateEvidenceCommandFile(ctx context.Context, manifestPath, sourcePath string) protocol.Result {
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return commandProblem("evidence", protocol.StatusOperational, "tool.read_failed", "unable to read evidence manifest")
	}
	source, err := os.ReadFile(sourcePath)
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
	result.Evidence = &protocol.EvidenceSummary{Schema: evidence.Schema, ID: manifest.ID, Digest: evidence.ContentDigest(manifestBytes)}
	return result.Finalize()
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
	_ = started
	result.Metrics.RecomputedWork = work
	return result.Finalize()
}

type VerifyOptions struct {
	ForceEngineMismatch bool
	ForceStaleManifest  bool
}

func VerifyCorpus(ctx context.Context, corpus string, runner native.Runner, options VerifyOptions) protocol.Result {
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
	validSource, err := os.ReadFile(validPath)
	if err != nil {
		return fail(protocol.StatusOperational, "verify.fixture_missing", "required positive fixture is unavailable")
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
		source, readErr := os.ReadFile(filepath.Join(corpus, name))
		if readErr != nil {
			addLane("lane:syntax-properties", "fail", nil, syntaxWork, syntaxBytes, laneStarted)
			return fail(protocol.StatusOperational, "verify.fixture_missing", "required syntax fixture is unavailable")
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
	negativeSource, err := os.ReadFile(filepath.Join(corpus, "non_exhaustive.lang"))
	if err != nil {
		addLane("lane:negative-controls", "fail", nil, 0, 0, laneStarted)
		return fail(protocol.StatusOperational, "verify.fixture_missing", "required negative fixture is unavailable")
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
	result.Evidence = &protocol.EvidenceSummary{Schema: evidence.Schema, ID: product.Manifest.ID, Digest: evidence.ContentDigest(product.ManifestBytes)}

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
