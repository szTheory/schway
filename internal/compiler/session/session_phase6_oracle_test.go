// session_phase6_oracle_test.go proves D-06-26's split oracle actually
// exists and runs: the differential-behaviour fallback lives HERE, in
// internal/compiler/session, in-process, re-running the shipped
// reduce/mismatch machinery's own -O0-vs--O3 comparison over a file the
// repair driver has ALREADY produced and exited from -- never inside
// cmd/lang-repair, which may not import internal/compiler/reduce
// (D-06-28). This file is package session (an internal test, like
// session_phase6_injectors_test.go), so it can reach the unexported
// interpreterInputs/phase5DefaultRunner helpers the same way
// session_phase6_mismatch.go's own production wiring does.
package session

import (
	"bytes"
	"context"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// buildLangRepairBinaryTimeout mirrors testsupport.BuildCLITimeout: cmd/
// lang-repair is package main and can never be imported (Go forbids
// importing package main), so exercising it from any test anywhere in
// the tree -- including this one -- means building and spawning it as a
// real subprocess, exactly like testsupport.BuildCLI does for ./cmd/lang.
const buildLangRepairBinaryTimeout = 5 * time.Minute

// buildLangRepairBinary is a small local copy of testsupport.BuildCLIErr's
// shape, parametrized on package path since testsupport.BuildCLI hardcodes
// ./cmd/lang.
func buildLangRepairBinary(t testing.TB) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "lang-repair")
	ctx, cancel := context.WithTimeout(context.Background(), buildLangRepairBinaryTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", binary, "./cmd/lang-repair")
	cmd.Dir = testsupport.ProjectPath()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("building cmd/lang-repair: %v\n%s%s", err, stdout.String(), stderr.String())
	}
	return binary
}

// assertRepairDriverNeverReferencesReducer is a direct, live go/ast
// falsifier for D-06-26/D-06-28's split: the differential-behaviour
// fallback oracle must never live inside cmd/lang-repair. It is narrower
// than (and independent of) import_boundary_test.go's own broad
// "/internal/" scan -- this one names the specific package the fallback
// oracle depends on, so a future refactor of the boundary lint cannot
// accidentally stop covering this specific forbidden dependency.
func assertRepairDriverNeverReferencesReducer(t testing.TB) {
	t.Helper()
	dir := testsupport.ProjectPath("cmd", "lang-repair")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	fileSet := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fileSet, filepath.Join(dir, entry.Name()), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imported := range file.Imports {
			path := strings.Trim(imported.Path.Value, `"`)
			if strings.Contains(path, "/compiler/reduce") {
				t.Fatalf("%s imports %s -- the differential-behaviour fallback oracle must never live inside cmd/lang-repair (D-06-26, D-06-28)", entry.Name(), path)
			}
		}
	}
}

// TestDifferentialFallbackOracleRunsInProcess drives the REAL
// cmd/lang-repair BINARY as an external subprocess (never an import --
// package main cannot be imported) to repair a real injected move
// defect, then, entirely independently and only AFTER that process has
// exited, asks the fallback oracle's own differential machinery (an
// interpreter run at -O0 compared against a native-compiled run at -O3,
// via the same Phase5CompareEngines this package's own mismatch-reduce
// wiring already uses) whether the repaired program's own two engines
// agree. "No disagreement" is itself a reached verdict -- a differential
// comparison that finds no divergence has still run and answered the
// question, exactly as a byte-identity oracle that finds two byte
// strings equal has still reached a verdict.
func TestDifferentialFallbackOracleRunsInProcess(t *testing.T) {
	assertRepairDriverNeverReferencesReducer(t)

	langBinary := testsupport.BuildCLI(t)
	repairBinary := buildLangRepairBinary(t)

	original, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase6", "heldout_move_defect.lang"))
	if err != nil {
		t.Fatal(err)
	}
	mutated, err := MoveInjector{}.Inject(original)
	if err != nil {
		t.Fatalf("injecting move defect: %v", err)
	}

	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "heldout_move_defect.lang")
	if err := os.WriteFile(sourcePath, mutated, 0o644); err != nil {
		t.Fatal(err)
	}

	// The driver runs, repairs, and EXITS. Everything below happens after
	// that process is gone -- the fallback oracle never shares an address
	// space, a subprocess, or a code path with the driver.
	result := testsupport.RunCLI(t, repairBinary, nil, "--lang="+langBinary, "--source="+sourcePath, "--json")
	if result.Exit != 0 {
		t.Fatalf("lang-repair failed: exit=%d stderr=%s stdout=%s", result.Exit, result.Stderr, result.Stdout)
	}

	repaired, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}

	checked := Check(repaired)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("repaired program failed to check clean: %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("repaired program rejected by corevalidate: %+v", validated.Problems)
	}
	program := validated.Program()
	cSource, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatal(err)
	}
	inputs, ok := interpreterInputs(program)
	if !ok || len(inputs) == 0 {
		t.Fatal("no derivable native input for the repaired program")
	}

	runner := phase5DefaultRunner()
	o0, err := runner.Run(context.Background(), cSource, "-O0", []string{inputs[0]})
	if err != nil {
		t.Fatalf("-O0 run: %v", err)
	}
	o3, err := runner.Run(context.Background(), cSource, "-O3", []string{inputs[0]})
	if err != nil {
		t.Fatalf("-O3 run: %v", err)
	}
	if len(o0.Pairs) != 1 || len(o3.Pairs) != 1 {
		t.Fatalf("expected exactly one execution per optimization level, got O0=%d O3=%d", len(o0.Pairs), len(o3.Pairs))
	}

	engines := map[string]execution.Execution{"O0": o0.Pairs[0].Execution, "O3": o3.Pairs[0].Execution}
	verdict := Phase5CompareEngines("differential-fallback-oracle", engines)
	if verdict != nil {
		t.Fatalf("differential-behaviour fallback oracle found a disagreement on a correctly repaired program: %v", verdict)
	}
}
