//go:build phase21_lto_evidence

package session_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase21EmittedMultiFunctionLTOComparison(t *testing.T) {
	const fixture = "testdata/phase14/multi_function_match_refusal.schway"
	clangPath, err := exec.LookPath("clang")
	if err != nil {
		t.Fatalf("find cschway: %v", err)
	}
	versionContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var versionOutput boundedVersionOutput
	versionCommand := exec.CommandContext(versionContext, clangPath, "--version")
	versionCommand.Stdout = &versionOutput
	versionCommand.Stderr = &versionOutput
	err = versionCommand.Run()
	if err != nil {
		t.Fatalf("clang --version: %v: %s", err, versionOutput.String())
	}
	version := strings.SplitN(strings.TrimSpace(versionOutput.String()), "\n", 2)[0]
	program, entryName := phase16CheckedFixture(t, fixture)
	if len(program.Functions) < 2 {
		t.Fatalf("fixture must be multi-function, got %d functions", len(program.Functions))
	}
	emittedC, err := cgen.EmitProgramNativeForTest(program)
	if err != nil {
		t.Fatalf("direct emitProgram: %v", err)
	}
	if !strings.Contains(emittedC, "lang.execution/2") {
		t.Fatal("direct emitProgram did not emit schema-2 C")
	}
	fixtureBytes, err := os.ReadFile(testsupport.ProjectPath(fixture))
	if err != nil {
		t.Fatalf("read fixture for digest: %v", err)
	}
	engines := phase11RunFourTiersWithSupplier(t, context.Background(), program, entryName, "Off", func(core.Program) (string, error) {
		return emittedC, nil
	}, "direct emitProgram(..., true)")
	engines["interpreter"] = phase16ProjectInterpreterSchema2(t, program, engines["interpreter"])
	if err := session.Phase5CompareEngines(fixture, engines); err != nil {
		t.Fatalf("direct emitProgram four-tier disagreement: %v", err)
	}
	if len(engines) != 4 {
		t.Fatalf("fixture %s produced %d execution lanes, want interpreter, -O0, -O3, and -O3 -flto", fixture, len(engines))
	}
	digest := sha256.Sum256([]byte(emittedC))
	fmt.Printf("Phase 21 emitted multi-function semantic comparison: fixture=%s fixture_sha256=%x emitted_c_sha256=%x host=%s/%s cschway_path=%s cschway_version=%q common_flags=[-std=c17,-Wall,-Wextra,-Werror,-pedantic] lanes=[interpreter,-O0,-O3,-O3 -flto] comparator=all-pairs-semantic-equality\n",
		fixture, sha256.Sum256(fixtureBytes), digest, runtime.GOOS, runtime.GOARCH, clangPath, version)
}

type boundedVersionOutput struct{ bytes.Buffer }

func (b *boundedVersionOutput) Write(p []byte) (int, error) {
	const maxVersionOutput = 4096
	if b.Len()+len(p) > maxVersionOutput {
		return 0, fmt.Errorf("clang version output exceeded %d bytes", maxVersionOutput)
	}
	return b.Buffer.Write(p)
}
