package cgen_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase27OutputExactLimit(t *testing.T) {
	fizzArtifact := buildPhase27OutputArtifact(t, "examples/fizzbuzz.schway", 4)
	assertPhase27NativeApp(t, fizzArtifact, "2", 0, []byte("1\n2\n"), "")

	sumArtifact := buildPhase27OutputArtifact(t, "examples/sum_to_n.schway", 3)
	assertPhase27NativeApp(t, sumArtifact, "10", 0, []byte("55\n"), "")
}

func TestPhase27OutputCrossingWrite(t *testing.T) {
	fizzArtifact := buildPhase27OutputArtifact(t, "examples/fizzbuzz.schway", 4)
	assertPhase27NativeApp(t, fizzArtifact, "3", 65, []byte("1\n2\n"), "schway: application output limit exceeded\n")

	sumArtifact := buildPhase27OutputArtifact(t, "examples/sum_to_n.schway", 2)
	assertPhase27NativeApp(t, sumArtifact, "10", 65, []byte("55"), "schway: application output limit exceeded\n")
}

func TestPhase27OutputProductionLimit(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "fizzbuzz.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("FizzBuzz check diagnostics: %+v", checked.Diagnostics)
	}
	generated, err := cgen.EmitApplication(checked.Program)
	if err != nil {
		t.Fatalf("EmitApplication: %v", err)
	}
	if !strings.Contains(generated, "#define SCHWAY_OUTPUT_LIMIT 65536u") || !strings.Contains(generated, "#define SCHWAY_EVIDENCE_OUTPUT_LIMIT 65536u") || execution.MaxApplicationOutputBytes != 65536 || execution.MaxApplicationEvidenceBytes != 65536 {
		t.Fatalf("production application and evidence limits are not independently pinned to 65536 bytes")
	}
}

func buildPhase27OutputArtifact(t *testing.T, sourceRelativePath string, outputLimit int) string {
	t.Helper()
	sourcePath := testsupport.ProjectPath(filepath.ToSlash(filepath.Clean(sourceRelativePath)))
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("%s check diagnostics: %+v", sourceRelativePath, checked.Diagnostics)
	}
	var generated string
	func() {
		restore := cgen.SetApplicationOutputLimitForTest(outputLimit)
		defer restore()
		generated, err = cgen.EmitApplication(checked.Program)
	}()
	if err != nil {
		t.Fatalf("EmitApplication(%s): %v", sourceRelativePath, err)
	}
	artifact := filepath.Join(t.TempDir(), filepath.Base(sourcePath)+".app")
	if _, err := native.BuildApplication(context.Background(), source, generated, artifact); err != nil {
		t.Fatalf("BuildApplication(%s): %v", sourceRelativePath, err)
	}
	return artifact
}

func assertPhase27NativeApp(t *testing.T, artifact, input string, wantCode int, wantStdout []byte, wantStderr string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	outcome, err := native.RunApplication(context.Background(), artifact, input, &stdout, &stderr)
	if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != wantCode || !bytes.Equal(stdout.Bytes(), wantStdout) || stderr.String() != wantStderr {
		t.Fatalf("input=%s outcome=%+v stdout=%q stderr=%q err=%v; want exit %d stdout=%q stderr=%q", input, outcome, stdout.Bytes(), stderr.String(), err, wantCode, wantStdout, wantStderr)
	}
}
