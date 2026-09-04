package native

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/codename-lang/lang/internal/compiler/execution"
)

func TestNativeStreamsIndependentlyBounded(t *testing.T) {
	tests := []struct {
		name string
		mode string
		code string
	}{
		{name: "compile stdout", mode: "compile-stdout-flood", code: "native.compile_stdout_truncated"},
		{name: "compile stderr", mode: "compile-stderr-flood", code: "native.compile_stderr_truncated"},
		{name: "run stdout", mode: "run-stdout-flood", code: "native.run_stdout_truncated"},
		{name: "run stderr", mode: "run-stderr-flood", code: "native.run_stderr_truncated"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			runner := Runner{command: func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
				calls++
				stage := "compile"
				if calls > 1 {
					stage = "run"
				}
				command := exec.CommandContext(ctx, os.Args[0], "-test.run=TestNativeHelperProcess", "--", test.mode, stage)
				command.Env = append(os.Environ(), "GO_WANT_NATIVE_HELPER=1")
				return command
			}}
			_, err := runner.Run(context.Background(), "int main(void) { return 0; }", "-O0", []string{"input"})
			var toolError *ToolError
			if !errors.As(err, &toolError) || toolError.Code != test.code {
				t.Fatalf("code=%v want=%s err=%v", toolError, test.code, err)
			}
		})
	}
}

// TestNativeTimeoutHasFalsifier closes D-02-04: the compile and run deadlines
// each have a committed hang-mode negative control. A helper process that
// blocks past the runner's own (deliberately short, for test speed) timeout
// must be observed to fire native.timeout at both the compile and run
// stages -- never native.compile_failed or native.run_failed, which would
// misreport a hang as a compile/run defect rather than a deadline.
func TestNativeTimeoutHasFalsifier(t *testing.T) {
	tests := []struct {
		name  string
		mode  string
		stage string
	}{
		{name: "compile hang", mode: "compile-hang", stage: "compile"},
		{name: "run hang", mode: "run-hang", stage: "run"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			runner := Runner{Timeout: 50 * time.Millisecond, command: func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
				calls++
				stage := "compile"
				if calls > 1 {
					stage = "run"
				}
				command := exec.CommandContext(ctx, os.Args[0], "-test.run=TestNativeHelperProcess", "--", test.mode, stage)
				command.Env = append(os.Environ(), "GO_WANT_NATIVE_HELPER=1")
				return command
			}}
			_, err := runner.Run(context.Background(), "int main(void) { return 0; }", "-O0", []string{"input"})
			var toolError *ToolError
			if !errors.As(err, &toolError) || toolError.Code != "native.timeout" {
				t.Fatalf("expected native.timeout at the %s stage, got %v", test.stage, err)
			}
		})
	}
}

// unboundedSpawnAllowlist is intentionally empty. Every Go file in this module
// outside .planning/spikes/ must spawn processes through exec.CommandContext
// (never a bare, aliased, or context.Background()-rooted constructor) with
// independently bounded stdout and stderr streams, and never through
// CombinedOutput/Output or an unbounded pipe read. Any entry added here must
// carry an explicit written justification; an unjustified entry is a review
// failure, not a waiver.
var unboundedSpawnAllowlist = map[string]string{}

// spawnViolation is one AST-resolved instance of a forbidden spawn shape
// (D-02-01's four documented evasions of the former two-needle substring
// scan): needle is a stable short identifier (kept for allowlist/log parity
// with the scanner's prior textual shape), reason is the human-readable
// justification for why the shape is forbidden.
type spawnViolation struct{ needle, reason string }

// scanUnboundedSpawns resolves, via go/ast rather than substring matching,
// every spawn constructor call and stream-capture call in file and reports
// the four evasions D-02-01 named against the prior substring scan:
//  1. exec.Command( — a bare constructor call with no context deadline at
//     all — and exec.CommandContext(context.Background(), ...) — a
//     constructor call that DOES carry a context parameter but roots it at
//     context.Background(), which carries no deadline either. Both are
//     "a context-free constructor" in effect.
//  2. .CombinedOutput() / .Output() — "a merged-output helper" that reads
//     the child's output into one unbounded buffer.
//  3. .StdoutPipe() / .StderrPipe() — "a pipe read with no limit": this
//     project's own bounded pattern is an independently size-capped
//     io.Writer assigned to Cmd.Stdout/Stderr, never a pipe read.
//  4. Assigning a bare exec.Command / exec.CommandContext selector value to
//     an identifier without calling it at the assignment site — "an
//     aliased constructor" that defeats call-site auditing of 1-3 above,
//     and any later call through that alias.
//
// This resolves the actual spawn constructor and stream-writer shape rather
// than matching a literal substring, so renaming a local variable, adding
// whitespace, or wrapping the call in a helper of a different name cannot
// evade it the way the substring scan's four documented defeats could.
func scanUnboundedSpawns(file *ast.File) []spawnViolation {
	execAlias := importAlias(file, "os/exec", "exec")
	if execAlias == "" {
		return nil // file does not import os/exec at all -- cannot spawn a process this way
	}
	contextAlias := importAlias(file, "context", "context")

	isExecSelector := func(expr ast.Expr, method string) bool {
		sel, ok := expr.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		ident, ok := sel.X.(*ast.Ident)
		return ok && ident.Name == execAlias && sel.Sel.Name == method
	}
	isContextBackground := func(expr ast.Expr) bool {
		call, ok := expr.(*ast.CallExpr)
		if !ok {
			return false
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		ident, ok := sel.X.(*ast.Ident)
		return ok && contextAlias != "" && ident.Name == contextAlias && sel.Sel.Name == "Background"
	}

	// Pass 1: aliased constructors -- a bare exec.Command/exec.CommandContext
	// selector value assigned to an identifier instead of called directly.
	aliased := map[string]bool{}
	ast.Inspect(file, func(node ast.Node) bool {
		assign, ok := node.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for index, rhs := range assign.Rhs {
			if index >= len(assign.Lhs) {
				continue
			}
			if isExecSelector(rhs, "Command") || isExecSelector(rhs, "CommandContext") {
				if ident, ok := assign.Lhs[index].(*ast.Ident); ok {
					aliased[ident.Name] = true
				}
			}
		}
		return true
	})

	var violations []spawnViolation
	if len(aliased) > 0 {
		violations = append(violations, spawnViolation{
			needle: "aliased-spawn-constructor",
			reason: "aliases exec.Command/exec.CommandContext to an indirect identifier instead of calling it directly, defeating call-site auditing",
		})
	}

	// Pass 2: direct constructor and stream-capture calls, plus calls
	// through an alias resolved in pass 1.
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fun := call.Fun.(type) {
		case *ast.SelectorExpr:
			if ident, ok := fun.X.(*ast.Ident); ok && ident.Name == execAlias {
				switch fun.Sel.Name {
				case "Command":
					violations = append(violations, spawnViolation{needle: "exec.Command(", reason: "spawns a process without a context deadline (use exec.CommandContext)"})
				case "CommandContext":
					if len(call.Args) > 0 && isContextBackground(call.Args[0]) {
						violations = append(violations, spawnViolation{needle: "CommandContext(context.Background())", reason: "spawns using context.Background(), which carries no deadline"})
					}
				}
				return true
			}
			switch fun.Sel.Name {
			case "CombinedOutput", "Output":
				violations = append(violations, spawnViolation{needle: fun.Sel.Name + "()", reason: "merges/collects output into one unbounded buffer instead of an independently bounded writer"})
			case "StdoutPipe", "StderrPipe":
				violations = append(violations, spawnViolation{needle: fun.Sel.Name + "()", reason: "reads a process stream through an unbounded pipe read instead of an independently bounded writer"})
			}
		case *ast.Ident:
			if aliased[fun.Name] {
				violations = append(violations, spawnViolation{needle: "aliased-spawn-constructor-call", reason: "calls an aliased exec.Command/exec.CommandContext identifier"})
			}
		}
		return true
	})
	return violations
}

// importAlias returns the local identifier a file binds importPath to: the
// explicit rename if one is present, defaultName if the import is present
// unaliased, or "" if the file does not import importPath at all (or
// blank-imports it, `_`, which can never be referenced as a selector base).
func importAlias(file *ast.File, importPath, defaultName string) string {
	for _, imported := range file.Imports {
		path, err := strconv.Unquote(imported.Path.Value)
		if err != nil || path != importPath {
			continue
		}
		if imported.Name == nil {
			return defaultName
		}
		if imported.Name.Name == "_" {
			return ""
		}
		return imported.Name.Name
	}
	return ""
}

// TestSourceNeverSpawnsUnboundedProcesses is the repo-wide successor to the
// former native.go-only single-file grep, and (D-02-01) the AST-resolving
// successor to the two-needle substring scan it was later widened into: it
// parses every *.go file in the module and reports every AST-resolved
// forbidden spawn shape scanUnboundedSpawns names, rather than matching a
// literal substring a rename or reformat could evade.
// Only .planning/spikes/ is excluded, because throwaway spike labs are not
// part of the compiler or its test support.
func TestSourceNeverSpawnsUnboundedProcesses(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller path unavailable")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	spikes := filepath.Join(root, ".planning", "spikes") + string(filepath.Separator)
	fileSet := token.NewFileSet()
	scanned := 0
	walkErr := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return fs.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasPrefix(path, spikes) {
			return nil
		}
		relative, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		relative = filepath.ToSlash(relative)
		parsed, parseErr := parser.ParseFile(fileSet, path, nil, 0)
		if parseErr != nil {
			return parseErr
		}
		scanned++
		for _, violation := range scanUnboundedSpawns(parsed) {
			if reason, allowed := unboundedSpawnAllowlist[relative]; allowed {
				t.Logf("allowlisted %s (%s): %s", relative, violation.needle, reason)
				continue
			}
			t.Errorf("%s uses %s, which %s", relative, violation.needle, violation.reason)
		}
		return nil
	})
	if walkErr != nil {
		t.Fatal(walkErr)
	}
	if scanned == 0 {
		t.Fatal("scanner found no Go sources; the guard would pass vacuously")
	}
}

// TestSpawnGuardCatchesKnownEvasions plants each of D-02-01's four
// documented evasions of the former substring scan in a throwaway source
// file (never written to disk inside the module -- parsed directly from an
// in-memory string via go/parser) and confirms scanUnboundedSpawns catches
// every one of them, plus confirms the scanner does not fire on an honestly
// bounded spawn shape matching this project's own convention.
func TestSpawnGuardCatchesKnownEvasions(t *testing.T) {
	tests := []struct {
		name   string
		source string
		needle string
	}{
		{
			name: "context-free constructor with plain buffers",
			source: `package evasion
import ("bytes"; "context"; "os/exec")
func run() {
	cmd := exec.CommandContext(context.Background(), "echo")
	var out bytes.Buffer
	cmd.Stdout = &out
	_ = cmd.Run()
}
`,
			needle: "CommandContext(context.Background())",
		},
		{
			name: "merged-output helper",
			source: `package evasion
import ("context"; "os/exec")
func run(ctx context.Context) {
	cmd := exec.CommandContext(ctx, "echo")
	_, _ = cmd.CombinedOutput()
}
`,
			needle: "CombinedOutput()",
		},
		{
			name: "pipe read with no limit",
			source: `package evasion
import ("context"; "io"; "os/exec")
func run(ctx context.Context) {
	cmd := exec.CommandContext(ctx, "echo")
	stdout, _ := cmd.StdoutPipe()
	_ = cmd.Start()
	_, _ = io.ReadAll(stdout)
}
`,
			needle: "StdoutPipe()",
		},
		{
			name: "aliased constructor",
			source: `package evasion
import "os/exec"
func run() {
	spawn := exec.CommandContext
	_ = spawn
}
`,
			needle: "aliased-spawn-constructor",
		},
	}
	fileSet := token.NewFileSet()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := parser.ParseFile(fileSet, "evasion.go", test.source, 0)
			if err != nil {
				t.Fatal(err)
			}
			violations := scanUnboundedSpawns(parsed)
			found := false
			for _, violation := range violations {
				if violation.needle == test.needle {
					found = true
				}
			}
			if !found {
				t.Fatalf("evasion %q was not caught: violations=%+v", test.name, violations)
			}
		})
	}

	t.Run("bare exec.Command is still caught", func(t *testing.T) {
		parsed, err := parser.ParseFile(fileSet, "evasion.go", `package evasion
import "os/exec"
func run() { _ = exec.Command("echo") }
`, 0)
		if err != nil {
			t.Fatal(err)
		}
		violations := scanUnboundedSpawns(parsed)
		if len(violations) != 1 || violations[0].needle != "exec.Command(" {
			t.Fatalf("expected exactly one exec.Command( violation, got %+v", violations)
		}
	})

	t.Run("an honestly bounded spawn does not fire", func(t *testing.T) {
		parsed, err := parser.ParseFile(fileSet, "honest.go", `package honest
import ("context"; "os/exec"; "time")
func run() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "echo")
	var stdout, stderr boundedWriter
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	_ = cmd.Run()
}
type boundedWriter struct{}
func (boundedWriter) Write(data []byte) (int, error) { return len(data), nil }
`, 0)
		if err != nil {
			t.Fatal(err)
		}
		if violations := scanUnboundedSpawns(parsed); len(violations) != 0 {
			t.Fatalf("honestly bounded spawn shape flagged: %+v", violations)
		}
	})
}

func TestExecutionDecoderRejectsMalformedOutput(t *testing.T) {
	valid, err := execution.CanonicalBytes(execution.Execution{
		Schema:  execution.Schema1,
		Outcome: execution.Outcome{Kind: "returned", Value: "01020304"},
		Events:  []execution.Event{}, LiveResources: []string{},
	})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		data []byte
		code string
	}{
		{name: "missing document fields", data: []byte(`{}`), code: "native.invalid_execution"},
		{name: "unknown field", data: []byte(`{"schema":"lang.execution/1","outcome":{"kind":"returned","value":"x"},"events":[],"live_resources":[],"unknown":true}`), code: "native.invalid_execution"},
		{name: "duplicate document", data: append(append([]byte{}, valid...), valid...), code: "native.trailing_execution"},
		{name: "trailing value", data: append(append([]byte{}, valid...), []byte(` true`)...), code: "native.trailing_execution"},
		{name: "malformed", data: []byte(`{"schema":`), code: "native.invalid_execution"},
		{name: "truncated", data: append([]byte{}, valid[:len(valid)-1]...), code: "native.invalid_execution"},
		{name: "oversized", data: append(append([]byte{}, valid...), []byte(strings.Repeat(" ", MaxStreamBytes))...), code: "native.run_stdout_truncated"},
		{name: "duplicate key", data: []byte(`{"schema":"lang.execution/1","schema":"lang.execution/1","outcome":{"kind":"returned","value":"x"},"events":[],"live_resources":[]}`), code: "native.invalid_execution"},
		{name: "unknown schema", data: []byte(`{"schema":"lang.execution/9","outcome":{"kind":"returned","value":"x"},"events":[],"live_resources":[]}`), code: "native.invalid_execution"},
		{name: "unknown outcome", data: []byte(`{"schema":"lang.execution/1","outcome":{"kind":"mystery","value":"x"},"events":[],"live_resources":[]}`), code: "native.invalid_execution"},
		{name: "empty outcome", data: []byte(`{"schema":"lang.execution/1","outcome":{"kind":"returned","value":""},"events":[],"live_resources":[]}`), code: "native.invalid_execution"},
		{name: "unknown event", data: []byte(`{"schema":"lang.execution/1","outcome":{"kind":"returned","value":"x"},"events":[{"schema":"lang.execution/1","id":"e","kind":"mystery","function_id":"f"}],"live_resources":[]}`), code: "native.invalid_execution"},
		{name: "missing transition fields", data: []byte(`{"schema":"lang.execution/1","outcome":{"kind":"returned","value":"x"},"events":[{"schema":"lang.execution/1","id":"e","kind":"value.copied","function_id":"f"}],"live_resources":[]}`), code: "native.invalid_execution"},
		{name: "return before transition", data: []byte(`{"schema":"lang.execution/1","outcome":{"kind":"returned","value":"x"},"events":[{"schema":"lang.execution/1","id":"r","kind":"function.returned","function_id":"f","source_place":"p","type_id":"t"},{"schema":"lang.execution/1","id":"e","kind":"value.copied","function_id":"f","source_place":"p","target_place":"q","type_id":"t"}],"live_resources":[]}`), code: "native.invalid_execution"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, decodeErr := decodeExecution(test.data, ExpectValue)
			var toolError *ToolError
			if !errors.As(decodeErr, &toolError) || toolError.Code != test.code {
				t.Fatalf("code=%v want=%s err=%v", toolError, test.code, decodeErr)
			}
		})
	}
}

func TestNativeToolFailureIsOperational(t *testing.T) {
	_, err := Runner{ClangPath: filepath.Join(t.TempDir(), "missing-clang")}.Run(
		context.Background(), "int main(void) { return 0; }", "-O0", []string{"input"},
	)
	var toolError *ToolError
	if !errors.As(err, &toolError) || toolError.Code != "native.tool_missing" {
		t.Fatalf("expected native.tool_missing, got %T %v", err, err)
	}
}

func TestNativeHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_NATIVE_HELPER") != "1" {
		return
	}
	mode, stage := os.Args[len(os.Args)-2], os.Args[len(os.Args)-1]
	flood := strings.Repeat("x", MaxStreamBytes+2)
	if stage == "compile" {
		switch mode {
		case "compile-stdout-flood":
			_, _ = os.Stdout.WriteString(flood)
		case "compile-stderr-flood":
			_, _ = os.Stderr.WriteString(flood)
		case "compile-hang":
			// D-02-04's compile-deadline falsifier: block well past the
			// runner's own deliberately short test Timeout. The parent's
			// context deadline kills this process; it never exits on its own.
			time.Sleep(10 * time.Second)
		}
		return
	}
	valid, _ := execution.CanonicalBytes(execution.Execution{
		Schema:  execution.Schema0,
		Outcome: execution.Outcome{Kind: "returned", Value: "input"},
		Events:  []execution.Event{}, LiveResources: []string{},
	})
	switch mode {
	case "run-stdout-flood":
		_, _ = os.Stdout.WriteString(flood)
	case "run-stderr-flood":
		_, _ = os.Stdout.Write(valid)
		_, _ = os.Stderr.WriteString(flood)
	case "run-hang":
		// D-02-04's run-deadline falsifier: block well past the runner's own
		// deliberately short test Timeout, mirroring compile-hang above.
		time.Sleep(10 * time.Second)
	default:
		_, _ = os.Stdout.Write(valid)
	}
}

// TestValidateExecutionStillRejectsOldGrounds is Pitfall 3's regression test
// (04-RESEARCH.md): widening validateExecution to accept an expected-outcome
// axis must not loosen the pre-Phase-4 "value" contract. Every one of the
// original hard-reject conditions must still reject under ExpectValue.
func TestValidateExecutionStillRejectsOldGrounds(t *testing.T) {
	valid := func() execution.Execution {
		return execution.Execution{
			Schema:  execution.Schema1,
			Outcome: execution.Outcome{Kind: "returned", Value: "7"},
			Events: []execution.Event{
				{Schema: execution.Schema1, ID: "op0:event:returned", Kind: "function.returned", FunctionID: "f1", SourcePlace: "p0", TypeID: "t0"},
			},
			LiveResources: []string{},
		}
	}
	if err := validateExecution(valid(), ExpectValue); err != nil {
		t.Fatalf("valid returned document rejected: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(execution.Execution) execution.Execution
	}{
		{"wrong outcome kind", func(v execution.Execution) execution.Execution { v.Outcome.Kind = "typed_failure"; return v }},
		{"empty outcome value", func(v execution.Execution) execution.Execution { v.Outcome.Value = ""; return v }},
		{"nil events", func(v execution.Execution) execution.Execution { v.Events = nil; return v }},
		{"empty events", func(v execution.Execution) execution.Execution { v.Events = []execution.Event{}; return v }},
		{"nil live resources", func(v execution.Execution) execution.Execution { v.LiveResources = nil; return v }},
		{"nonempty live resources", func(v execution.Execution) execution.Execution { v.LiveResources = []string{"r1"}; return v }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutated := test.mutate(valid())
			if err := validateExecution(mutated, ExpectValue); err == nil {
				t.Fatalf("expected rejection under ExpectValue, got nil error")
			}
		})
	}
}

// TestValidateExecutionTypedFailureDiscriminates proves the new axis
// discriminates in both directions rather than merely widening: a
// typed_failure-expecting call accepts a typed_failure document and rejects
// a returned one, and vice versa.
func TestValidateExecutionTypedFailureDiscriminates(t *testing.T) {
	typedFailure := execution.Execution{
		Schema:  execution.Schema1,
		Outcome: execution.Outcome{Kind: "typed_failure", Value: "OpenFailed"},
		Events: []execution.Event{
			{Schema: execution.Schema1, ID: "op0:event:failed", Kind: "function.failed", FunctionID: "f1", SourcePlace: "p2", TypeID: "t1"},
		},
		LiveResources: []string{},
	}
	if err := validateExecution(typedFailure, ExpectTypedFailure); err != nil {
		t.Fatalf("valid typed_failure document rejected under ExpectTypedFailure: %v", err)
	}
	if err := validateExecution(typedFailure, ExpectValue); err == nil {
		t.Fatalf("typed_failure document accepted under ExpectValue")
	}

	returned := execution.Execution{
		Schema:  execution.Schema1,
		Outcome: execution.Outcome{Kind: "returned", Value: "7"},
		Events: []execution.Event{
			{Schema: execution.Schema1, ID: "op0:event:returned", Kind: "function.returned", FunctionID: "f1", SourcePlace: "p0", TypeID: "t0"},
		},
		LiveResources: []string{},
	}
	if err := validateExecution(returned, ExpectTypedFailure); err == nil {
		t.Fatalf("returned document accepted under ExpectTypedFailure")
	}
}
