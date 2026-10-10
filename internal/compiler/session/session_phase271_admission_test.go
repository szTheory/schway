package session_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/szTheory/schway/internal/compiler/callgraph"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// requirePhase271ControlReached proves the origin-unpublishable control is
// actually reached by every layer before the admission tests assert refusal
// (D-27.1-12): checker clean, corevalidate valid, a linear U64->U64 entry, and
// exactly one core.origin_omitted problem from the published-origin peer.
func requirePhase271ControlReached(t *testing.T) []byte {
	t.Helper()
	source := mustPhase25Fixture(t, "shared_copy_accept.schway")
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("phase27.1 control not reached: checker diagnostics %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("phase27.1 control not reached: corevalidate refused: %+v", validated.Problems)
	}
	program := validated.Program()
	entry, err := callgraph.EntryFunction(program)
	if err != nil {
		t.Fatalf("phase27.1 control not reached: entry function: %v", err)
	}
	if entry.Linear == nil || entry.Match != nil || entry.Parameter.Type != "U64" || entry.ReturnType != "U64" {
		t.Fatalf("phase27.1 control not reached: entry is not a linear U64->U64 function: %+v", entry)
	}
	problems := originvalidate.ValidatePublished(program)
	if len(problems) != 1 || problems[0].Code != "core.origin_omitted" {
		t.Fatalf("phase27.1 control not reached: ValidatePublished = %+v, want exactly [core.origin_omitted]", problems)
	}
	return source
}

// phase271MissingClangRunner returns a runner whose clang does not exist, so
// any route that proceeds past admission surfaces native.tool_missing.
func phase271MissingClangRunner(t *testing.T) native.Runner {
	t.Helper()
	return native.Runner{ClangPath: filepath.Join(t.TempDir(), "missing-clang")}
}

func TestPhase271BuildApplicationRefusesOriginUnpublishable(t *testing.T) {
	source := requirePhase271ControlReached(t)
	outDir := t.TempDir()
	artifact := filepath.Join(outDir, "app")
	receipt, diagnostics, err := session.BuildApplication(context.Background(), source, artifact, phase271MissingClangRunner(t))
	if err != nil {
		_, isTool := err.(*native.ToolError)
		t.Fatalf("BuildApplication returned err=%v (*native.ToolError=%v); want diagnostic refusal with nil error", err, isTool)
	}
	if len(diagnostics) != 1 || diagnostics[0].Code != "core.origin_omitted" {
		t.Fatalf("BuildApplication diagnostics = %+v, want exactly [core.origin_omitted]", diagnostics)
	}
	if !reflect.DeepEqual(receipt, native.BuildReceipt{}) {
		t.Fatalf("BuildApplication receipt = %+v, want zero value", receipt)
	}
	entries, readErr := os.ReadDir(outDir)
	if readErr != nil || len(entries) != 0 {
		t.Fatalf("output directory not empty after refusal: entries=%v err=%v", entries, readErr)
	}
	for _, path := range []string{artifact, artifact + ".schway-build.json"} {
		if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
			t.Fatalf("%s exists after refusal (stat err=%v)", path, statErr)
		}
	}
	checkResult, checkErr := session.CheckCommandFile(testsupport.ProjectPath("testdata", "phase25", "shared_copy_accept.schway"))
	if checkErr != nil || len(checkResult.Diagnostics) == 0 {
		t.Fatalf("CheckCommandFile result=%+v err=%v", checkResult, checkErr)
	}
	if !reflect.DeepEqual(diagnostics[0], checkResult.Diagnostics[0]) {
		t.Fatalf("build diagnostic differs from check:\n build=%+v\n check=%+v", diagnostics[0], checkResult.Diagnostics[0])
	}
}

// phase271ReplayCases is one well-formed source case so case-set validation
// passes and admission is what refuses.
func phase271ReplayCases() session.ReplayCases {
	return session.ReplayCases{
		Schema: session.ReplayCasesSchema,
		Cases: []session.ReplayCase{{
			ID: "c1", Kind: session.ReplayCaseSource, Input: "7",
			Expected:        session.ReplayExpected{Kind: execution.OutcomeReturned, Value: "7"},
			ForeignOutcomes: []session.ReplayForeignOutcome{},
		}},
	}
}

func mustReadProjectFile(t *testing.T, parts ...string) []byte {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath(parts...))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func requireEmptyDir(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("output directory not empty: entries=%v err=%v", entries, err)
	}
}

func requireVerifyRefusal(t *testing.T, report session.ReplayReport, err error, want string) {
	t.Helper()
	if err == nil || err.Error() != want {
		t.Fatalf("VerifyApplicationCases err = %v, want %q", err, want)
	}
	if report.Verified || report.EmittedCDigest != "" || len(report.Cases) != 0 {
		t.Fatalf("refused verify report carries evidence: %+v", report)
	}
}

func TestApplicationRoutesApplySameAdmissionLayers(t *testing.T) {
	ctx := context.Background()

	checkerSource := mustPhase25Fixture(t, "exclusive_conflict_reject.schway")
	if checked := session.Check(checkerSource); len(checked.Diagnostics) == 0 || checked.Diagnostics[0].Code != "ownership.borrow_conflict" {
		t.Fatalf("phase27.1 parity source not at layer: checker fixture diagnostics %+v", checked.Diagnostics)
	}
	coreSource := mustReadProjectFile(t, "testdata", "phase07", "duplicate_function_name.schway")
	coreChecked := session.Check(coreSource)
	coreValidated := corevalidate.Validate(coreChecked.Program)
	if len(coreChecked.Diagnostics) != 0 || coreValidated.Valid || len(coreValidated.Problems) == 0 || coreValidated.Problems[0].Code != "core.duplicate_function_id" {
		t.Fatalf("phase27.1 parity source not at layer: corevalidate fixture diagnostics=%+v valid=%v problems=%+v", coreChecked.Diagnostics, coreValidated.Valid, coreValidated.Problems)
	}
	originSource := requirePhase271ControlReached(t)

	t.Run("checker/build", func(t *testing.T) {
		dir := t.TempDir()
		receipt, diagnostics, err := session.BuildApplication(ctx, checkerSource, filepath.Join(dir, "app"), phase271MissingClangRunner(t))
		if err != nil || len(diagnostics) == 0 || !reflect.DeepEqual(receipt, native.BuildReceipt{}) {
			t.Fatalf("receipt=%+v diagnostics=%+v err=%v", receipt, diagnostics, err)
		}
		requireEmptyDir(t, dir)
	})
	t.Run("checker/interp", func(t *testing.T) {
		result, diagnostics, err := session.RunInterpreterApplication(checkerSource, "7")
		if err != nil || len(diagnostics) == 0 || !reflect.DeepEqual(result, interpZeroResult()) {
			t.Fatalf("result=%+v diagnostics=%+v err=%v", result, diagnostics, err)
		}
	})
	t.Run("checker/verify", func(t *testing.T) {
		report, err := session.VerifyApplicationCases(ctx, checkerSource, phase271ReplayCases(), phase271MissingClangRunner(t))
		requireVerifyRefusal(t, report, err, "source is not accepted by the checker: ownership.borrow_conflict")
	})

	t.Run("corevalidate/build", func(t *testing.T) {
		dir := t.TempDir()
		receipt, diagnostics, err := session.BuildApplication(ctx, coreSource, filepath.Join(dir, "app"), phase271MissingClangRunner(t))
		if err == nil || err.Error() != "core validation failed: core.duplicate_function_id" || len(diagnostics) != 0 || !reflect.DeepEqual(receipt, native.BuildReceipt{}) {
			t.Fatalf("receipt=%+v diagnostics=%+v err=%v", receipt, diagnostics, err)
		}
		requireEmptyDir(t, dir)
	})
	t.Run("corevalidate/interp", func(t *testing.T) {
		result, diagnostics, err := session.RunInterpreterApplication(coreSource, "7")
		if err == nil || err.Error() != "core validation failed: core.duplicate_function_id" || len(diagnostics) != 0 || !reflect.DeepEqual(result, interpZeroResult()) {
			t.Fatalf("result=%+v diagnostics=%+v err=%v", result, diagnostics, err)
		}
	})
	t.Run("corevalidate/verify", func(t *testing.T) {
		report, err := session.VerifyApplicationCases(ctx, coreSource, phase271ReplayCases(), phase271MissingClangRunner(t))
		requireVerifyRefusal(t, report, err, "source core validation failed")
	})

	t.Run("origin/build", func(t *testing.T) {
		dir := t.TempDir()
		receipt, diagnostics, err := session.BuildApplication(ctx, originSource, filepath.Join(dir, "app"), phase271MissingClangRunner(t))
		if err != nil || len(diagnostics) != 1 || diagnostics[0].Code != "core.origin_omitted" || !reflect.DeepEqual(receipt, native.BuildReceipt{}) {
			t.Fatalf("receipt=%+v diagnostics=%+v err=%v", receipt, diagnostics, err)
		}
		requireEmptyDir(t, dir)
	})
	t.Run("origin/interp", func(t *testing.T) {
		result, diagnostics, err := session.RunInterpreterApplication(originSource, "7")
		if err == nil || err.Error() != "origin validation failed: core.origin_omitted" || len(diagnostics) != 0 || !reflect.DeepEqual(result, interpZeroResult()) {
			t.Fatalf("result=%+v diagnostics=%+v err=%v", result, diagnostics, err)
		}
	})
	t.Run("origin/verify", func(t *testing.T) {
		report, err := session.VerifyApplicationCases(ctx, originSource, phase271ReplayCases(), phase271MissingClangRunner(t))
		requireVerifyRefusal(t, report, err, "source origin validation failed: core.origin_omitted")
	})
}

func TestPhase271PositiveControlsReachRunner(t *testing.T) {
	controls := []struct {
		source   []string
		manifest []string
	}{
		{source: []string{"examples", "sum_to_n.schway"}},
		{source: []string{"examples", "fizzbuzz.schway"}},
		{source: []string{"examples", "phase24", "error.schway"}, manifest: []string{"examples", "phase23", "file_byte.bindings.json"}},
		{source: []string{"examples", "phase24", "transfer.schway"}, manifest: []string{"examples", "phase23", "file_byte.bindings.json"}},
		{source: []string{"examples", "phase26", "checked_add_overflow.schway"}},
		{source: []string{"examples", "phase22", "identity.schway"}, manifest: []string{"examples", "phase22", "identity.bindings.json"}},
		{source: []string{"examples", "phase23", "file_byte.schway"}, manifest: []string{"examples", "phase23", "file_byte.bindings.json"}},
	}
	for _, control := range controls {
		path := filepath.Join(control.source...)
		t.Run(path, func(t *testing.T) {
			var manifest []string
			if control.manifest != nil {
				manifest = []string{testsupport.ProjectPath(control.manifest...)}
			}
			_, diagnostics, err := session.BuildApplicationFile(context.Background(), testsupport.ProjectPath(control.source...), filepath.Join(t.TempDir(), "app"), phase271MissingClangRunner(t), manifest...)
			var toolError *native.ToolError
			if len(diagnostics) != 0 || !errors.As(err, &toolError) || toolError.Code != "native.tool_missing" {
				t.Fatalf("admission did not reach the runner: diagnostics=%+v err=%v", diagnostics, err)
			}
		})
	}
}

func TestPhase271VerifyFileRouteWritesOnlyFailReport(t *testing.T) {
	requirePhase271ControlReached(t)
	casesDir := t.TempDir()
	reportDir := t.TempDir()
	casesPath := filepath.Join(casesDir, "cases.json")
	encoded, err := json.Marshal(phase271ReplayCases())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(casesPath, encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	reportPath := filepath.Join(reportDir, "report.json")
	_, err = session.VerifyApplicationCasesFile(context.Background(), testsupport.ProjectPath("testdata", "phase25", "shared_copy_accept.schway"), casesPath, reportPath, phase271MissingClangRunner(t))
	if err == nil || err.Error() != "source origin validation failed: core.origin_omitted" {
		t.Fatalf("VerifyApplicationCasesFile err = %v, want source origin validation failed: core.origin_omitted", err)
	}
	entries, readErr := os.ReadDir(reportDir)
	if readErr != nil || len(entries) != 1 || entries[0].Name() != "report.json" {
		t.Fatalf("report directory entries = %v err=%v, want exactly report.json", entries, readErr)
	}
	caseEntries, readErr := os.ReadDir(casesDir)
	if readErr != nil || len(caseEntries) != 1 || caseEntries[0].Name() != "cases.json" {
		t.Fatalf("cases directory entries = %v err=%v, want exactly cases.json", caseEntries, readErr)
	}
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	var report session.ReplayReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("decode fail report: %v", err)
	}
	if report.Status != session.ReplayStatusFail || report.Verified || report.EmittedCDigest != "" || report.BuildID != "" || len(report.Cases) != 0 {
		t.Fatalf("fail report carries evidence: %+v", report)
	}
}

func interpZeroResult() interp.ApplicationResult { return interp.ApplicationResult{} }
