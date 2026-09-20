package native

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/syntax"
)

// projectRoot resolves the repository root the same way
// ForeignResourceSourcePath does, without depending on package session (an
// internal test file in package native must not import session, which
// itself imports native -- see #cycle note on CompileOnly's own doc
// comment for why this file duplicates rather than shares that helper).
func projectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

func compiledForeignAcquireOneBinary(t *testing.T) (string, func()) {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(projectRoot(), "testdata", "phase4", "foreign_acquire_one.lang"))
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("fixture failed to parse: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture failed to check: %+v", checked.Diagnostics)
	}
	if _, err := cgen.EmitNative(checked.Program); err == nil || !strings.Contains(err.Error(), "multi-function foreign-call bodies are not supported") {
		t.Fatalf("EmitNative must retain the named M004 refusal, got %v", err)
	}
	cSourceBytes, err := os.ReadFile(filepath.Join(projectRoot(), "testdata", "phase16", "historical", "foreign_acquire_one.c"))
	if err != nil {
		t.Fatal(err)
	}
	cSource := string(cSourceBytes)
	runner := DefaultRunner()
	runner.ForeignSources = []string{ForeignResourceSourcePath()}
	binaryPath, cleanup, err := runner.CompileOnly(context.Background(), cSource, "-O0")
	if err != nil {
		t.Fatal(err)
	}
	return binaryPath, cleanup
}

// TestUndefinedSymbolAllowlistRejectsNewSymbol proves D-04-19's allowlist is
// a POSITIVE list: a binary that links in an undefined symbol not on the
// allowlist (here, libc printf, which nothing in this project's generated C
// or frozen foreign TUs calls) is rejected.
func TestUndefinedSymbolAllowlistRejectsNewSymbol(t *testing.T) {
	directory := t.TempDir()
	sourcePath := filepath.Join(directory, "unlisted.c")
	source := "#include <stdio.h>\nint main(void) { printf(\"probe\\n\"); return 0; }\n"
	if err := os.WriteFile(sourcePath, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	runner := DefaultRunner()
	binaryPath, cleanup, err := runner.CompileOnly(context.Background(), source, "-O0")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	status, rejected, toolErr := CheckUndefinedSymbolAllowlist(context.Background(), "", binaryPath, 0)
	if toolErr != nil {
		t.Fatal(toolErr)
	}
	if status != SymbolsFail {
		t.Fatalf("expected SymbolsFail for an unlisted undefined symbol, got %s", status)
	}
	found := false
	for _, symbol := range rejected {
		if symbol == "printf" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected printf among rejected symbols, got %+v", rejected)
	}
}

// TestUndefinedSymbolAllowlistPassesTracer proves the control passes for
// the project's own real generated tracer program, compiled and linked
// against the frozen foreign TU exactly as the session-level gate does.
func TestUndefinedSymbolAllowlistPassesTracer(t *testing.T) {
	binaryPath, cleanup := compiledForeignAcquireOneBinary(t)
	defer cleanup()
	status, rejected, toolErr := CheckUndefinedSymbolAllowlist(context.Background(), "", binaryPath, 0)
	if toolErr != nil {
		t.Fatal(toolErr)
	}
	if status != SymbolsPass {
		t.Fatalf("expected SymbolsPass for the project's own tracer binary, got %s rejected=%+v", status, rejected)
	}
}

// TestUndefinedSymbolNormalizationHandlesBothFormats proves the
// leading-underscore normalization strips AT MOST one leading underscore --
// the Mach-O ("_malloc") vs ELF ("malloc") naming difference -- without
// mangling a symbol that has none.
func TestUndefinedSymbolNormalizationHandlesBothFormats(t *testing.T) {
	tests := []struct{ in, want string }{
		{"_malloc", "malloc"},
		{"malloc", "malloc"},
		{"_longjmp", "longjmp"},
		{"__someDoubleUnderscorePrefixed", "_someDoubleUnderscorePrefixed"},
	}
	for _, test := range tests {
		if got := normalizeSymbol(test.in); got != test.want {
			t.Fatalf("normalizeSymbol(%q) = %q, want %q", test.in, got, test.want)
		}
	}
}

// TestMissingSymbolToolReportsOperational proves D-04-19's tool-missing
// posture: pointing the tool path at a nonexistent binary must report
// SymbolsOperational -- NEVER SymbolsPass. "A tool that is absent is not
// evidence of absence."
func TestMissingSymbolToolReportsOperational(t *testing.T) {
	binaryPath, cleanup := compiledForeignAcquireOneBinary(t)
	defer cleanup()
	status, rejected, toolErr := CheckUndefinedSymbolAllowlist(context.Background(), filepath.Join(t.TempDir(), "no-such-nm-tool"), binaryPath, 0)
	if status == SymbolsPass {
		t.Fatalf("a missing tool must never report SymbolsPass")
	}
	if status != SymbolsOperational {
		t.Fatalf("expected SymbolsOperational for a missing tool, got %s rejected=%+v", status, rejected)
	}
	if toolErr == nil || toolErr.Code != "native.tool_missing" {
		t.Fatalf("expected native.tool_missing, got %+v", toolErr)
	}
}

// TestUnwindControlDoesNotInspectSections proves D-04-19's own rejected
// alternative: the control's pass/fail decision never depends on
// unwind-section presence, since a compact-unwind section can be present on
// a program that never unwinds on this platform (04-RESEARCH.md), which
// would make a section assertion pass or fail for the wrong reason. The
// project's own tracer binary genuinely carries Apple's mandatory compact
// unwind section (skips gracefully if otool is unavailable to confirm
// that), yet the control still passes -- proving its decision is symbol-set
// membership alone.
func TestUnwindControlDoesNotInspectSections(t *testing.T) {
	binaryPath, cleanup := compiledForeignAcquireOneBinary(t)
	defer cleanup()
	if _, err := exec.LookPath("otool"); err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var stdout bytes.Buffer
		command := exec.CommandContext(ctx, "otool", "-l", binaryPath)
		command.Stdout = &stdout
		if runErr := command.Run(); runErr == nil && !strings.Contains(stdout.String(), "__unwind_info") {
			t.Skip("env:otool -- host binary format has no inspectable unwind section; nothing to contrast against")
		}
	}
	status, rejected, toolErr := CheckUndefinedSymbolAllowlist(context.Background(), "", binaryPath, 0)
	if toolErr != nil {
		t.Fatal(toolErr)
	}
	if status != SymbolsPass {
		t.Fatalf("control must pass on symbol membership alone, regardless of unwind-section presence; got %s rejected=%+v", status, rejected)
	}
}
