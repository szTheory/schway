package session_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

func checkedPhase16Fixture(t *testing.T, relative string) session.CheckResult {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath(relative))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("%s: check diagnostics: %+v", relative, checked.Diagnostics)
	}
	return checked
}

func TestPhase16ControlNativeCAdmittedUsesPublicEmitter(t *testing.T) {
	checked := checkedPhase16Fixture(t, "testdata/phase1/toggle.lang")
	want, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatalf("public emitter: %v", err)
	}
	got, err := session.Phase16ControlNativeC(checked.Program, "testdata/phase4/nonlocal_exit_probe.lang")
	if err != nil {
		t.Fatalf("session boundary: %v", err)
	}
	if !bytes.Equal([]byte(got), []byte(want)) {
		t.Fatal("session boundary did not return the public emitter result")
	}
}

func TestPhase16ControlNativeCPreservesM004Refusal(t *testing.T) {
	checked := checkedPhase16Fixture(t, "testdata/phase4/nonlocal_exit_probe.lang")
	_, want := cgen.EmitNative(checked.Program)
	got, err := session.Phase16ControlNativeC(checked.Program, "testdata/phase4/nonlocal_exit_probe.lang")
	if err == nil {
		t.Fatalf("session boundary returned historical C after public refusal: %q", got)
	}
	if got != "" {
		t.Fatalf("session boundary returned C with refusal: %q", got)
	}
	if err.Error() != want.Error() {
		t.Fatalf("refusal = %q, want public refusal %q", err, want)
	}
}
