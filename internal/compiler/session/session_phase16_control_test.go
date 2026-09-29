package session_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
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
	checked := checkedPhase16Fixture(t, "testdata/phase1/toggle.schway")
	want, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatalf("public emitter: %v", err)
	}
	got, err := session.Phase16ControlNativeC(checked.Program, "testdata/phase4/nonlocal_exit_probe.schway")
	if err != nil {
		t.Fatalf("session boundary: %v", err)
	}
	if !bytes.Equal([]byte(got), []byte(want)) {
		t.Fatal("session boundary did not return the public emitter result")
	}
}

func TestPhase16ControlNativeCPreservesM004Refusal(t *testing.T) {
	checked := checkedPhase16Fixture(t, "testdata/phase4/nonlocal_exit_probe.schway")
	_, want := cgen.EmitNative(checked.Program)
	got, err := session.Phase16ControlNativeC(checked.Program, "testdata/phase4/nonlocal_exit_probe.schway")
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

func TestPhase16ProductionSourcesCannotLoadFrozenC(t *testing.T) {
	files, err := filepath.Glob(testsupport.ProjectPath("internal/compiler/session/*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		source, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if refs := phase16FrozenCSelectors(string(source)); len(refs) > 0 {
			t.Fatalf("production source %s can select frozen C through %q", file, refs)
		}
	}
}

func phase16FrozenCSelectors(source string) []string {
	var found []string
	for _, selector := range []string{"testdata/phase16/historical", "file-frozen-evidence.json", "generated-frozen-evidence.json"} {
		if strings.Contains(source, selector) {
			found = append(found, selector)
		}
	}
	return found
}
