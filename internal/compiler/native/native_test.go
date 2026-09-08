package native

import (
	"bytes"
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
	"syscall"
	"testing"
	"time"

	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// phase4OutOfCorpusCase is one hand-written, out-of-corpus program (task
// 04-07-03, D-04-21) and the subcommands to exercise against it through the
// SHIPPED binary. None of these sources is read from testdata/ -- each is
// written fresh to t.TempDir() -- so TestShippedBinaryExercisesEveryPhase4Behavior
// can assert none of them lives under a corpus directory.
type phase4OutOfCorpusCase struct {
	behavior string
	source   string
	steps    []phase4OutOfCorpusStep
}

// phase4OutOfCorpusStep is one subcommand invocation and its expected,
// VERBATIM-recorded outcome: an exit code, and (when non-empty) a
// diagnostic code the JSON output must contain.
type phase4OutOfCorpusStep struct {
	args           []string // "{path}" is substituted with the source file's path
	wantExit       int
	wantDiagnostic string
}

func phase4OutOfCorpusCases() []phase4OutOfCorpusCase {
	return []phase4OutOfCorpusCase{
		{
			behavior: "Fallible foreign call through try",
			source: `module outofcorpus.tracer_probe

export {
  fn main
}

foreign C {

  fn lang_res_open(request: Byte) -> Byte {
    unwind: forbidden
    nonlocal_exit: forbidden
    allocator: "libc_malloc"
    fails: AcquireError
  }
}

data AcquireError =
  | OpenFailed

fn main(request: Byte) -> Byte {
  let handle = try lang_res_open(request)
  handle
}
`,
			steps: []phase4OutOfCorpusStep{
				{args: []string{"format", "--check", "{path}"}, wantExit: 0},
				{args: []string{"check", "{path}"}, wantExit: 0},
				{args: []string{"run", "--engine=interpreter", "{path}"}, wantExit: 0},
				{args: []string{"run", "--engine=native", "{path}"}, wantExit: 0},
			},
		},
		{
			behavior: "Three-stage acquisition with reverse-order release",
			source: `module outofcorpus.triple_acquire

export {
  fn main
}

foreign C {

  fn lang_res_open(request: Byte) -> Byte {
    unwind: forbidden
    nonlocal_exit: forbidden
    allocator: "libc_malloc"
    fails: AcquireError
  }
}

data AcquireError =
  | OpenFailed

fn main(request: Byte) -> Byte {
  let first = try lang_res_open(request)
  let second = try lang_res_open(request)
  let third = try lang_res_open(request)
  request
}
`,
			steps: []phase4OutOfCorpusStep{
				{args: []string{"format", "--check", "{path}"}, wantExit: 0},
				{args: []string{"check", "{path}"}, wantExit: 0},
				{args: []string{"run", "--engine=interpreter", "{path}"}, wantExit: 0},
				{args: []string{"run", "--engine=native", "{path}"}, wantExit: 0},
			},
		},
		{
			behavior: `discard ... because consumer`,
			source: `module outofcorpus.discard_probe

export {
  fn main
}

foreign C {

  fn lang_res_open(request: Byte) -> Byte {
    unwind: forbidden
    nonlocal_exit: forbidden
    allocator: "libc_malloc"
    fails: AcquireError
  }
}

data AcquireError =
  | OpenFailed

fn main(request: Byte) -> Byte {
  discard lang_res_open(request) because "out-of-corpus probe never inspects this acquisition"
  request
}
`,
			steps: []phase4OutOfCorpusStep{
				{args: []string{"format", "--check", "{path}"}, wantExit: 0},
				{args: []string{"check", "{path}"}, wantExit: 0},
				{args: []string{"run", "--engine=interpreter", "{path}"}, wantExit: 0},
				{args: []string{"run", "--engine=native", "{path}"}, wantExit: 0},
			},
		},
		{
			behavior: "Defect terminator",
			source: `module outofcorpus.defect_probe

export {
  type Signal
  fn triage
}

data Signal =
  | Go
  | Halt

fn triage(flag: Signal) -> Signal {
  match flag {
    Go => {
      let held = take flag
      held
    }
    Halt => {
      defect "out-of-corpus halt"
    }
  }
}
`,
			steps: []phase4OutOfCorpusStep{
				{args: []string{"format", "--check", "{path}"}, wantExit: 0},
				{args: []string{"check", "{path}"}, wantExit: 0},
				{args: []string{"run", "--engine=interpreter", "{path}"}, wantExit: 0},
				{args: []string{"run", "--engine=native", "{path}"}, wantExit: 0},
			},
		},
		{
			behavior: "Nonlocal-exit probe",
			source: `module outofcorpus.nonlocal_probe

export {
  fn main
}

foreign C {

  fn lang_nonlocal_probe(request: Byte) -> Byte {
    unwind: forbidden
    nonlocal_exit: possible
    allocator: "libc_malloc"
    fails: ProbeError
  }
}

data ProbeError =
  | ProbeFailed

fn main(request: Byte) -> Byte {
  let handle = try lang_nonlocal_probe(request)
  let trigger = try lang_nonlocal_probe(request)
  request
}
`,
			steps: []phase4OutOfCorpusStep{
				{args: []string{"format", "--check", "{path}"}, wantExit: 0},
				{args: []string{"check", "{path}"}, wantExit: 0},
				{args: []string{"run", "--engine=interpreter", "{path}"}, wantExit: 0},
				{args: []string{"run", "--engine=native", "{path}"}, wantExit: 0},
			},
		},
		{
			behavior: "Foreign-origin refusal",
			source: `module outofcorpus.origin_probe

export {
  fn main
}

foreign C {

  fn lang_res_open(request: Byte) -> Byte {
    unwind: forbidden
    nonlocal_exit: forbidden
    allocator: "libc_malloc"
    fails: AcquireError
    alias: "borrow"
  }
}

data AcquireError =
  | OpenFailed

fn main(request: Byte) -> Byte {
  let handle = try lang_res_open(request)
  handle
}
`,
			steps: []phase4OutOfCorpusStep{
				{args: []string{"--json", "check", "{path}"}, wantExit: 2, wantDiagnostic: "core.foreign_origin_omitted"},
			},
		},
		{
			behavior: "Policy-less foreign declaration refusal",
			source: `module outofcorpus.unwind_undeclared_probe

export {
  fn main
}

foreign C {

  fn lang_res_open(request: Byte) -> Byte {
    allocator: "libc_malloc"
    fails: AcquireError
  }
}

data AcquireError =
  | OpenFailed

fn main(request: Byte) -> Byte {
  let handle = try lang_res_open(request)
  handle
}
`,
			steps: []phase4OutOfCorpusStep{
				{args: []string{"--json", "check", "{path}"}, wantExit: 2, wantDiagnostic: "foreign.unwind_policy_undeclared"},
			},
		},
		{
			behavior: "Lang-targeted call refusal",
			source: `module outofcorpus.call_target_probe

export {
  fn main
  fn helper
}

foreign C {

  fn lang_res_open(request: Byte) -> Byte {
    unwind: forbidden
    nonlocal_exit: forbidden
    allocator: "libc_malloc"
    fails: AcquireError
  }
}

data AcquireError =
  | OpenFailed

fn helper(value: Byte) -> Byte {
  value
}

fn main(request: Byte) -> Byte {
  let handle = try helper(request)
  handle
}
`,
			steps: []phase4OutOfCorpusStep{
				{args: []string{"--json", "check", "{path}"}, wantExit: 2, wantDiagnostic: "core.call_target_not_foreign"},
			},
		},
	}
}

// TestShippedBinaryExercisesEveryPhase4Behavior is task 04-07-03's own
// shipped-binary register closure (D-04-21): every Phase 4 behavior named
// in the plan's behavior list is driven against a freshly built ./cmd/lang
// on a hand-written program that is in NO corpus, with subcommands, exit
// codes, and diagnostic codes recorded so the exercise is repeatable
// rather than a one-time manual act.
func TestShippedBinaryExercisesEveryPhase4Behavior(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	for _, testCase := range phase4OutOfCorpusCases() {
		t.Run(testCase.behavior, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "probe.lang")
			if err := os.WriteFile(path, []byte(testCase.source), 0o600); err != nil {
				t.Fatal(err)
			}
			if strings.Contains(path, "testdata") {
				t.Fatalf("out-of-corpus source must never live under a corpus directory, got %s", path)
			}
			for _, step := range testCase.steps {
				arguments := make([]string, len(step.args))
				for index, argument := range step.args {
					if argument == "{path}" {
						argument = path
					}
					arguments[index] = argument
				}
				result := testsupport.RunCLI(t, binary, nil, arguments...)
				t.Logf("subcommand=%v exit=%d stdout=%s", arguments, result.Exit, result.Stdout)
				if result.Exit != step.wantExit {
					t.Fatalf("subcommand %v: exit=%d, want %d (stdout=%s stderr=%s)", arguments, result.Exit, step.wantExit, result.Stdout, result.Stderr)
				}
				if step.wantDiagnostic != "" && !strings.Contains(string(result.Stdout), `"code":"`+step.wantDiagnostic+`"`) {
					t.Fatalf("subcommand %v: expected diagnostic code %s, got stdout=%s", arguments, step.wantDiagnostic, result.Stdout)
				}
			}
		})
	}
}

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
		// D-04-20: an input that ends before a value finishes decoding (both
		// of these) is now the distinct native.terminal_record_absent hard
		// failure, never the generic native.invalid_execution a COMPLETE but
		// malformed document still reports -- see
		// TestTerminalRecordAbsenceIsHardFailure/TestTruncationAndAbsenceReportDistinctCodes.
		{name: "malformed", data: []byte(`{"schema":`), code: "native.terminal_record_absent"},
		{name: "truncated", data: append([]byte{}, valid[:len(valid)-1]...), code: "native.terminal_record_absent"},
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
	case "abort-mid-stream":
		// D-04-20's abort falsifier: write a syntactically-open (never
		// closed) JSON document containing one streamed event, flush it,
		// then die by SIGABRT -- exactly the shape a real lang_defect()
		// call leaves behind if fired mid-stream. Proves the parent's own
		// stdout capture retains every byte written before the signal.
		_, _ = os.Stdout.WriteString(`{"schema":"lang.execution/1","events":[{"schema":"lang.execution/1","id":"op:0:event","kind":"value.copied","function_id":"fn:probe","source_place":"place:0","target_place":"place:1","type_id":"type:0"}`)
		_ = os.Stdout.Sync()
		_ = syscall.Kill(os.Getpid(), syscall.SIGABRT)
		time.Sleep(5 * time.Second) // should never be reached
	case "abort-clean":
		// A complete, valid defect terminal record, written in full before
		// the process aborts -- the honest D-04-15 shape, as opposed to
		// abort-mid-stream's deliberately partial one.
		valid, _ := execution.CanonicalBytes(execution.Execution{
			Schema: execution.Schema1, Outcome: execution.Outcome{Kind: "defect", Value: ""},
			Events: []execution.Event{{
				Schema: execution.Schema1, ID: "op:0:event:defected", Kind: "function.defected",
				FunctionID: "fn:probe", SourcePlace: "place:0", TypeID: "type:0", Output: "reason",
			}},
			LiveResources: []string{},
		})
		_, _ = os.Stdout.Write(valid)
		_ = os.Stdout.Sync()
		_ = syscall.Kill(os.Getpid(), syscall.SIGABRT)
		time.Sleep(5 * time.Second) // should never be reached
	case "exit-nonzero":
		os.Exit(3)
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

// TestDefectExpectationRejectsReturnedDocument is ExpectDefect's discriminant
// sibling to TestValidateExecutionTypedFailureDiscriminates: a returned
// document must never be accepted as a defect terminal outcome, and vice
// versa.
func TestDefectExpectationRejectsReturnedDocument(t *testing.T) {
	defect := execution.Execution{
		Schema:  execution.Schema1,
		Outcome: execution.Outcome{Kind: "defect", Value: ""},
		Events: []execution.Event{
			{Schema: execution.Schema1, ID: "op0:event:defected", Kind: "function.defected", FunctionID: "f1", SourcePlace: "p0", TypeID: "t0", Output: "halt requested"},
		},
		LiveResources: []string{},
	}
	if err := validateExecution(defect, ExpectDefect); err != nil {
		t.Fatalf("valid defect document rejected under ExpectDefect: %v", err)
	}
	returned := execution.Execution{
		Schema:  execution.Schema1,
		Outcome: execution.Outcome{Kind: "returned", Value: "7"},
		Events: []execution.Event{
			{Schema: execution.Schema1, ID: "op0:event:returned", Kind: "function.returned", FunctionID: "f1", SourcePlace: "p0", TypeID: "t0"},
		},
		LiveResources: []string{},
	}
	if err := validateExecution(returned, ExpectDefect); err == nil {
		t.Fatalf("returned document accepted under ExpectDefect")
	}
	if err := validateExecution(defect, ExpectValue); err == nil {
		t.Fatalf("defect document accepted under ExpectValue")
	}
}

// TestTerminalRecordAbsenceIsHardFailure is D-04-20's own falsifier: stdout
// that ends before a terminal record finishes decoding is a HARD FAILURE
// with its own distinct code, never an ordinary "invalid_execution" parse
// error and never tolerated as a passing run.
func TestTerminalRecordAbsenceIsHardFailure(t *testing.T) {
	_, err := decodeExecution([]byte(`{"schema":"lang.execution/1","events":[`), ExpectValue)
	var toolError *ToolError
	if !errors.As(err, &toolError) || toolError.Code != "native.terminal_record_absent" {
		t.Fatalf("code=%v want=native.terminal_record_absent err=%v", toolError, err)
	}
}

// TestTruncationAndAbsenceReportDistinctCodes proves the two failure shapes
// are distinguishable in the harness: output cut off by the existing byte
// cap reports native.run_stdout_truncated, while output that ended because
// the process died before a terminal record was written reports the
// distinct native.terminal_record_absent -- never the same code.
func TestTruncationAndAbsenceReportDistinctCodes(t *testing.T) {
	_, truncErr := decodeExecution(bytes.Repeat([]byte("x"), MaxStreamBytes+2), ExpectValue)
	var truncToolError *ToolError
	if !errors.As(truncErr, &truncToolError) || truncToolError.Code != "native.run_stdout_truncated" {
		t.Fatalf("truncation code=%v want=native.run_stdout_truncated err=%v", truncToolError, truncErr)
	}
	_, absentErr := decodeExecution([]byte(`{"schema":"lang.execution/1"`), ExpectValue)
	var absentToolError *ToolError
	if !errors.As(absentErr, &absentToolError) || absentToolError.Code != "native.terminal_record_absent" {
		t.Fatalf("absence code=%v want=native.terminal_record_absent err=%v", absentToolError, absentErr)
	}
	if truncToolError.Code == absentToolError.Code {
		t.Fatalf("truncation and absence must report distinct codes, both got %q", truncToolError.Code)
	}
}

// TestAbortingProgramStreamsEventsBeforeDying proves the run-stdout capture
// itself (native.go's own boundedWriter, reused unchanged for an aborting
// child per D-04-24) retains every byte written before a child dies by
// signal -- the underlying guarantee D-04-20's streaming emitter design
// depends on: writes reaching the pipe before SIGABRT are never lost by the
// parent's own capture layer, regardless of what the generated C itself
// chooses to write. The captured (necessarily incomplete, since the process
// died mid-write) document surfaces as native.terminal_record_absent, with
// the streamed event bytes visible in the error for inspection.
func TestAbortingProgramStreamsEventsBeforeDying(t *testing.T) {
	calls := 0
	runner := Runner{Expect: ExpectDefect, command: func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
		calls++
		stage := "compile"
		if calls > 1 {
			stage = "run"
		}
		command := exec.CommandContext(ctx, os.Args[0], "-test.run=TestNativeHelperProcess", "--", "abort-mid-stream", stage)
		// GOTRACEBACK=crash: without it, Go's own runtime signal handler
		// intercepts SIGABRT and calls os.Exit(2) instead of letting the raw
		// signal terminate the process, which would make
		// ProcessState.Sys().(syscall.WaitStatus).Signaled() false -- exactly
		// the wrong shape for this falsifier. This only affects the
		// synthetic Go test helper process; the REAL generated C's abort()
		// call has no such interception.
		command.Env = append(os.Environ(), "GO_WANT_NATIVE_HELPER=1", "GOTRACEBACK=crash")
		return command
	}}
	_, err := runner.Run(context.Background(), "int main(void) { return 0; }", "-O0", []string{"input"})
	var toolError *ToolError
	if !errors.As(err, &toolError) {
		t.Fatalf("expected a ToolError, got %v", err)
	}
	if toolError.Code != "native.terminal_record_absent" {
		t.Fatalf("code=%s want=native.terminal_record_absent (partial stream captured, but no terminal record) err=%v", toolError.Code, err)
	}
	if !strings.Contains(toolError.Error(), "value.copied") {
		t.Fatalf("streamed event bytes were lost when the child aborted: %v", err)
	}
}

// TestAbortSignalAdjudicatedByWaitStatus proves D-04-24's adjudication: a
// process that terminates via SIGABRT while the caller expects a defect
// terminal outcome is the expected abort-only shape, decided ONLY through
// ProcessState.Sys().(syscall.WaitStatus) -- its complete terminal record is
// then decoded successfully as a genuine defect execution.
func TestAbortSignalAdjudicatedByWaitStatus(t *testing.T) {
	calls := 0
	runner := Runner{Expect: ExpectDefect, command: func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
		calls++
		stage := "compile"
		if calls > 1 {
			stage = "run"
		}
		command := exec.CommandContext(ctx, os.Args[0], "-test.run=TestNativeHelperProcess", "--", "abort-clean", stage)
		// GOTRACEBACK=crash: see TestAbortingProgramStreamsEventsBeforeDying's
		// identical comment -- without it Go's own runtime intercepts SIGABRT
		// with os.Exit(2) instead of a real signalled termination.
		command.Env = append(os.Environ(), "GO_WANT_NATIVE_HELPER=1", "GOTRACEBACK=crash")
		return command
	}}
	result, err := runner.Run(context.Background(), "int main(void) { return 0; }", "-O0", []string{"input"})
	if err != nil {
		t.Fatalf("expected the SIGABRT termination to be adjudicated as the expected defect shape, got %v", err)
	}
	if len(result.Pairs) != 1 || result.Pairs[0].Execution.Outcome.Kind != "defect" {
		t.Fatalf("expected a decoded defect execution, got %+v", result)
	}
}

// TestNonzeroExitIsDistinctFromSignal proves the three exit shapes (clean,
// nonzero without a signal, and signalled) report three distinct codes: an
// ordinary nonzero os.Exit is native.run_failed, while a SIGABRT the caller
// did NOT expect (Expect stays ExpectValue) is the distinct
// native.run_signaled -- never conflated with an ordinary run failure.
func TestNonzeroExitIsDistinctFromSignal(t *testing.T) {
	run := func(mode string) *ToolError {
		calls := 0
		runner := Runner{command: func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
			calls++
			stage := "compile"
			if calls > 1 {
				stage = "run"
			}
			command := exec.CommandContext(ctx, os.Args[0], "-test.run=TestNativeHelperProcess", "--", mode, stage)
			// GOTRACEBACK=crash is inert for exit-nonzero and required for
			// abort-clean to actually terminate by signal -- see
			// TestAbortingProgramStreamsEventsBeforeDying's comment.
			command.Env = append(os.Environ(), "GO_WANT_NATIVE_HELPER=1", "GOTRACEBACK=crash")
			return command
		}}
		_, err := runner.Run(context.Background(), "int main(void) { return 0; }", "-O0", []string{"input"})
		var toolError *ToolError
		if !errors.As(err, &toolError) {
			t.Fatalf("%s: expected a ToolError, got %v", mode, err)
		}
		return toolError
	}
	nonzero := run("exit-nonzero")
	if nonzero.Code != "native.run_failed" {
		t.Fatalf("nonzero exit code=%s want=native.run_failed", nonzero.Code)
	}
	// abort-clean SIGABRTs, but this runner's Expect stays the zero value
	// (ExpectValue), so the signal is unexpected here -- distinct from both
	// nonzero and (in TestAbortSignalAdjudicatedByWaitStatus) the expected
	// case.
	signaled := run("abort-clean")
	if signaled.Code != "native.run_signaled" {
		t.Fatalf("unexpected signal code=%s want=native.run_signaled", signaled.Code)
	}
	if nonzero.Code == signaled.Code {
		t.Fatalf("nonzero exit and signalled termination must report distinct codes, both got %q", nonzero.Code)
	}
}

// phase4CorpusMatrix is the recorded, exhaustive four-subcommand outcome of
// the SHIPPED ./cmd/lang binary against every Phase 4 corpus fixture. It is
// the in-repo, re-runnable replacement for the hand-driven shipped-binary
// tables that 04-01-SUMMARY.md and 04-02-SUMMARY.md recorded verbatim as a
// one-time manual act (D-04-21): the same claim, asserted by CI on every
// commit rather than re-typed by a human at UAT time.
//
// Every accepting fixture must be clean through all four subcommands, and
// every refusing fixture must be refused by `check`, `run
// --engine=interpreter`, and `run --engine=native` with the SAME diagnostic
// code -- an entry point that accepts what a peer refuses is the exact
// three-engine divergence Phase 4's differential lane exists to catch, and
// two such divergences (both in `run --engine=native`) were real bugs found
// and fixed during 04-07.
func phase4CorpusMatrix() []phase4OutOfCorpusCase {
	// clean is a fixture the whole toolchain accepts: exit 0 from all four
	// subcommands.
	clean := func(fixture string) phase4OutOfCorpusCase {
		return phase4OutOfCorpusCase{
			behavior: fixture,
			steps: []phase4OutOfCorpusStep{
				{args: []string{"--json", "format", "--check", "{path}"}, wantExit: 0},
				{args: []string{"--json", "check", "{path}"}, wantExit: 0},
				{args: []string{"--json", "run", "--engine=interpreter", "{path}"}, wantExit: 0},
				{args: []string{"--json", "run", "--engine=native", "{path}"}, wantExit: 0},
			},
		}
	}
	// refused is a negative-control fixture: canonically formatted (unless
	// formatExit says otherwise), then refused identically by check and both
	// engines with one diagnostic code.
	refused := func(fixture string, formatExit int, formatDiagnostic, diagnostic string) phase4OutOfCorpusCase {
		return phase4OutOfCorpusCase{
			behavior: fixture,
			steps: []phase4OutOfCorpusStep{
				{args: []string{"--json", "format", "--check", "{path}"}, wantExit: formatExit, wantDiagnostic: formatDiagnostic},
				{args: []string{"--json", "check", "{path}"}, wantExit: 2, wantDiagnostic: diagnostic},
				{args: []string{"--json", "run", "--engine=interpreter", "{path}"}, wantExit: 2, wantDiagnostic: diagnostic},
				{args: []string{"--json", "run", "--engine=native", "{path}"}, wantExit: 2, wantDiagnostic: diagnostic},
			},
		}
	}
	return []phase4OutOfCorpusCase{
		clean("acquire_three_fail_second.lang"),
		clean("acquire_three_fail_third.lang"),
		clean("acquire_three_success.lang"),
		clean("defect_terminal.lang"),
		clean("discard_because.lang"),
		clean("foreign_acquire_one.lang"),
		clean("nonlocal_exit_probe.lang"),
		// D-07-40 (deliberate edit): this fixture's refusal moved from a
		// syntax-level refusal to a check-level one (D-07-01) -- the parser
		// now accepts a bare call's shape unconditionally, and check refuses
		// a foreign callee with the SAME published code
		// (syntax.fallible_call_not_consumed), not a new one. `format
		// --check` no longer short-circuits on a parser diagnostic for this
		// fixture, so it now runs the real formatter, which finds this
		// pre-Phase-07 fixture's leading-comment-to-module spacing
		// non-canonical (format.non_canonical) -- a pre-existing formatting
		// fact this plan surfaces but does not alter (the fixture's bytes
		// are unchanged; only what layer refuses the SEMANTIC shape moved).
		refused("fallible_call_unconsumed.lang", 2, "format.non_canonical", "syntax.fallible_call_not_consumed"),
		refused("foreign_call_target_not_foreign.lang", 0, "", "core.call_target_not_foreign"),
		refused("foreign_origin_omitted.lang", 0, "", "core.foreign_origin_omitted"),
		refused("foreign_policy_value_injection.lang", 0, "", "check.foreign_policy_value_unsafe"),
		refused("foreign_unwind_undeclared.lang", 0, "", "foreign.unwind_policy_undeclared"),
	}
}

// TestShippedBinaryFourSubcommandCorpusMatrix drives the shipped binary
// through `format --check`, `check`, `run --engine=interpreter`, and `run
// --engine=native` on every Phase 4 corpus fixture and compares against the
// recorded matrix. The matrix must be exhaustive: a fixture added to
// testdata/phase4 without an entry here fails the test rather than being
// silently unexercised.
func TestShippedBinaryFourSubcommandCorpusMatrix(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	corpus := testsupport.ProjectPath("testdata", "phase4")
	entries, err := filepath.Glob(filepath.Join(corpus, "*.lang"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("testdata/phase4 holds no .lang fixture")
	}
	matrix := phase4CorpusMatrix()
	expected := make(map[string]phase4OutOfCorpusCase, len(matrix))
	for _, testCase := range matrix {
		if _, duplicate := expected[testCase.behavior]; duplicate {
			t.Fatalf("phase4CorpusMatrix lists %s twice", testCase.behavior)
		}
		expected[testCase.behavior] = testCase
	}
	present := make(map[string]bool, len(entries))
	for _, path := range entries {
		present[filepath.Base(path)] = true
	}
	for fixture := range expected {
		if !present[fixture] {
			t.Fatalf("phase4CorpusMatrix names %s, which is not in testdata/phase4", fixture)
		}
	}
	for _, path := range entries {
		fixture := filepath.Base(path)
		testCase, listed := expected[fixture]
		if !listed {
			t.Fatalf("testdata/phase4/%s has no phase4CorpusMatrix entry; add its recorded four-subcommand outcome", fixture)
		}
		t.Run(fixture, func(t *testing.T) {
			runShippedBinarySteps(t, binary, path, testCase.steps)
		})
	}
}

// TestFallibleCallUnconsumedRefusedAtCheckNotFormat is Task 3's native-level
// pin (D-07-40, 07-03-PLAN.md): drives the shipped binary through the
// recorded four-subcommand matrix for fallible_call_unconsumed.lang
// specifically -- `check`, `run --engine=interpreter`, and
// `run --engine=native` all refuse it with the SAME preserved diagnostic
// code phase4CorpusMatrix declares, and `format --check` now reports
// format.non_canonical rather than the syntax code (the enforcement layer
// moved to check; the published code did not).
func TestFallibleCallUnconsumedRefusedAtCheckNotFormat(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	path := testsupport.ProjectPath("testdata", "phase4", "fallible_call_unconsumed.lang")
	var testCase phase4OutOfCorpusCase
	for _, candidate := range phase4CorpusMatrix() {
		if candidate.behavior == "fallible_call_unconsumed.lang" {
			testCase = candidate
		}
	}
	if len(testCase.steps) == 0 {
		t.Fatal("phase4CorpusMatrix has no entry for fallible_call_unconsumed.lang")
	}
	runShippedBinarySteps(t, binary, path, testCase.steps)
}

// runShippedBinarySteps drives one source file through a recorded sequence of
// shipped-binary subcommands, substituting "{path}" and asserting each step's
// exit code and (when named) diagnostic code.
func runShippedBinarySteps(t *testing.T, binary, path string, steps []phase4OutOfCorpusStep) {
	t.Helper()
	for _, step := range steps {
		arguments := make([]string, len(step.args))
		for index, argument := range step.args {
			if argument == "{path}" {
				argument = path
			}
			arguments[index] = argument
		}
		result := testsupport.RunCLI(t, binary, nil, arguments...)
		t.Logf("subcommand=%v exit=%d", arguments, result.Exit)
		if result.Exit != step.wantExit {
			t.Fatalf("subcommand %v: exit=%d, want %d (stdout=%s stderr=%s)", arguments, result.Exit, step.wantExit, result.Stdout, result.Stderr)
		}
		if step.wantDiagnostic != "" && !strings.Contains(string(result.Stdout), `"code":"`+step.wantDiagnostic+`"`) {
			t.Fatalf("subcommand %v: expected diagnostic code %s, got stdout=%s", arguments, step.wantDiagnostic, result.Stdout)
		}
	}
}

// TestOutOfCorpusSourcesAreGenuinelyNovel closes the automatable half of
// 04-07's D3 human-judgment checkpoint ("the out-of-corpus fixtures' genuine
// novelty relative to the corpus"). Every hand-written program driven by
// TestShippedBinaryExercisesEveryPhase4Behavior must differ from every file
// checked into testdata/ -- a "hand-written, out-of-corpus" program that had
// drifted into a byte-copy of a corpus fixture would make the shipped-binary
// register a restatement of the corpus rather than an independent exercise
// of it.
//
// Byte-inequality is the mechanical half and is what this test asserts. The
// remaining question -- whether a program is *interestingly* novel rather
// than merely non-identical -- stays a review judgment and is deliberately
// not claimed here.
func TestOutOfCorpusSourcesAreGenuinelyNovel(t *testing.T) {
	corpusRoot := testsupport.ProjectPath("testdata")
	corpus := make(map[string]string)
	err := filepath.Walk(corpusRoot, func(path string, info fs.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() || !strings.HasSuffix(path, ".lang") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		corpus[strings.TrimSpace(string(data))] = path
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(corpus) == 0 {
		t.Fatal("testdata holds no .lang fixture to compare against")
	}
	for _, testCase := range phase4OutOfCorpusCases() {
		if match, identical := corpus[strings.TrimSpace(testCase.source)]; identical {
			t.Fatalf("out-of-corpus program %q is byte-identical to corpus fixture %s", testCase.behavior, match)
		}
	}
}
