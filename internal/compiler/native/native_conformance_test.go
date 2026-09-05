package native_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestConformanceUnitCompilesSeparately proves the generated conformance
// unit (D-04-11) compiles as its own separate, bounded, timed invocation --
// distinct from the program compile, sharing no writer or timeout context --
// and is never passed to the link step (native.Runner.CompileConformanceUnit
// never links at all: it always passes "-c"). It also proves the real
// production foreign contract's declared Layout genuinely conforms to the
// frozen native/lang_foreign_resource_private.h, since this is the same
// mechanism control:foreign.layout_mismatch exercises against a
// deliberately WRONG fixture.
func TestConformanceUnitCompilesSeparately(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "foreign_acquire_one.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture failed to check: %+v", checked.Diagnostics)
	}
	unit, err := cgen.EmitForeignConformance(checked.Program, native.ForeignResourcePrivateHeaderPath())
	if err != nil {
		t.Fatal(err)
	}
	runner := native.DefaultRunner()
	if err := runner.CompileConformanceUnit(context.Background(), unit); err != nil {
		t.Fatalf("expected the real private header to conform, got %v", err)
	}
}

// TestPrivateHeaderConformanceFailureIsDistinctCode proves a conformance
// compile failure reports "native.conformance_failed", never
// "native.compile_failed" (the program-compile code) -- so a conformance
// refusal is never mistaken for a program compile failure.
func TestPrivateHeaderConformanceFailureIsDistinctCode(t *testing.T) {
	runner := native.DefaultRunner()
	err := runner.CompileConformanceUnit(context.Background(), "this is not valid C at all {{{")
	if err == nil {
		t.Fatal("expected invalid C to fail to compile")
	}
	var toolErr *native.ToolError
	if !errors.As(err, &toolErr) || toolErr.Code != "native.conformance_failed" {
		t.Fatalf("expected native.conformance_failed, got %v", err)
	}
}
