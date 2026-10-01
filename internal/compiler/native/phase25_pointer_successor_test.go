package native_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase25SharedNativeShape(t *testing.T) {
	clang, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal("Phase 25 native pointer witness requires installed clang")
	}
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase25", "shared_copy_accept.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("shared-copy fixture failed to check: %+v", checked.Diagnostics)
	}
	generated, err := cgen.EmitProgramNativeForTest(checked.Program)
	if err != nil {
		t.Fatalf("shared pointer helper was refused: %v", err)
	}
	path := filepath.Join(t.TempDir(), "shared_copy.c")
	if err := os.WriteFile(path, []byte(generated), 0o600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(clang, "-std=c17", "-Werror", "-fsyntax-only", path).CombinedOutput(); err != nil {
		t.Fatalf("clang rejected emitted shared-pointer C: %v\n%s", err, output)
	}
}
