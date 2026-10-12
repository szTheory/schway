package session_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/callgraph"
	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/modload"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/pathoracle"
	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/session"
)

// Phase 30 session route tests (plan 30-08). Fixtures are generated under
// t.TempDir() only and never committed (D-27).

// writeModuleFixture writes files (slash-separated relative name to content)
// under a fresh TempDir and returns the path of main.schway.
func writeModuleFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(dir, "main.schway")
}

const (
	phase30TwiceEntry = "module app.main\n\nimport lib {\n  fn twice\n}\n\nexport {\n  fn main\n}\n\nfn main(n: U64) -> U64 {\n  let r = twice(n)\n  r\n}\n"
	phase30TwiceLib   = "module lib\n\nexport {\n  fn twice\n}\n\nfn twice(n: U64) -> U64 {\n  let d = n + n\n  d\n}\n"
)

func phase30TwiceFixture(t *testing.T) string {
	t.Helper()
	return writeModuleFixture(t, map[string]string{"main.schway": phase30TwiceEntry, "lib.schway": phase30TwiceLib})
}

func phase30ResultJSON(t *testing.T, result protocol.Result) string {
	t.Helper()
	encoded, err := protocol.JSON(result)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func phase30ResultHuman(t *testing.T, result protocol.Result) string {
	t.Helper()
	human, err := protocol.Human(result)
	if err != nil {
		t.Fatal(err)
	}
	return human
}

// phase30InterpretU64 runs a checked merged program under the interpreter the
// way the application lane does: resolve the entry, prune to what it reaches,
// then run it on one canonical U64 input.
func phase30InterpretU64(t *testing.T, program core.Program, input string) string {
	t.Helper()
	entry, err := callgraph.EntryFunction(program)
	if err != nil {
		t.Fatal(err)
	}
	pruned, err := callgraph.PruneToReachable(program, entry.ID)
	if err != nil {
		t.Fatal(err)
	}
	result, err := interp.RunApplication(pruned, entry.Name, input)
	if err != nil {
		t.Fatal(err)
	}
	return result.Execution.Outcome.Value
}

// TestPhase30CheckFileThroughLoader is the tracer: check and the interpreter
// route load an entry that imports a library, through modload and the merged
// program (D-22).
func TestPhase30CheckFileThroughLoader(t *testing.T) {
	entry := phase30TwiceFixture(t)

	result, err := session.CheckCommandFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != protocol.StatusPass {
		t.Fatalf("CheckCommandFile status = %s: %s", result.Status, phase30ResultJSON(t, result))
	}

	checked, err := session.CheckFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("CheckFile diagnostics: %+v", checked.Diagnostics)
	}
	if len(checked.UnitTrees) != 2 {
		t.Fatalf("UnitTrees = %d trees, want entry and lib", len(checked.UnitTrees))
	}
	if got := phase30InterpretU64(t, checked.Program, "21"); got != "42" {
		t.Fatalf("interpreter result = %q, want 42", got)
	}

	t.Run("interpreter-route", func(t *testing.T) {
		byteEntry := writeModuleFixture(t, map[string]string{
			"main.schway": "module app.main\n\nimport lib {\n  fn same\n}\n\nexport {\n  fn main\n}\n\nfn main(value: Byte) -> Byte {\n  let r = same(value)\n  r\n}\n",
			"lib.schway":  "module lib\n\nexport {\n  fn same\n}\n\nfn same(value: Byte) -> Byte {\n  value\n}\n",
		})
		executions, diagnostics, err := session.RunInterpreterFile(byteEntry)
		if err != nil || len(diagnostics) != 0 || len(executions) != 1 {
			t.Fatalf("RunInterpreterFile = %d executions, %+v, %v", len(executions), diagnostics, err)
		}
		if executions[0].Outcome.Value != "7" {
			t.Fatalf("outcome = %q, want 7", executions[0].Outcome.Value)
		}
		command, err := session.RunInterpreterCommandFile(byteEntry)
		if err != nil || command.Status != protocol.StatusPass || len(command.Executions) != 1 {
			t.Fatalf("RunInterpreterCommandFile = %s, %v", phase30ResultJSON(t, command), err)
		}
	})

	t.Run("no-import-bytes-identical", func(t *testing.T) {
		source, err := os.ReadFile(filepath.Join("..", "..", "..", "examples", "phase22", "identity.schway"))
		if err != nil {
			t.Fatal(err)
		}
		path := writeModuleFixture(t, map[string]string{"main.schway": string(source)})
		fromFile, err := session.CheckFile(path)
		if err != nil {
			t.Fatal(err)
		}
		fromBytes := session.Check(source)
		fileJSON, _ := json.Marshal(fromFile)
		byteJSON, _ := json.Marshal(fromBytes)
		if !bytes.Equal(fileJSON, byteJSON) {
			t.Fatalf("CheckFile differs from Check(bytes):\n%s\n%s", fileJSON, byteJSON)
		}
		if fromFile.UnitTrees != nil {
			t.Fatal("a no-import program must carry no UnitTrees")
		}
	})
}

func phase30RequireClang(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("env:clang toolchain unavailable on this host")
	}
}

// phase30ByteTwoModules is a Byte-entry two-module program: the interpreter and
// native run lanes drive only Byte/Buffer/data entries with synthetic inputs.
func phase30ByteTwoModules(t *testing.T) string {
	t.Helper()
	return writeModuleFixture(t, map[string]string{
		"main.schway": "module app.main\n\nimport lib {\n  fn same\n}\n\nexport {\n  fn main\n}\n\nfn main(value: Byte) -> Byte {\n  let r = same(value)\n  r\n}\n",
		"lib.schway":  "module lib\n\nexport {\n  fn same\n}\n\nfn same(value: Byte) -> Byte {\n  value\n}\n",
	})
}

func TestPhase30ModuleRoutes(t *testing.T) {
	entry := phase30TwiceFixture(t)

	t.Run("debug-map", func(t *testing.T) {
		result, err := session.DebugMapCommandFile(entry, "")
		if err != nil {
			t.Fatal(err)
		}
		if result.Status != protocol.StatusPass || result.DebugMap == nil {
			t.Fatalf("debug-map = %s", phase30ResultJSON(t, result))
		}
		units := map[string]int{}
		for _, mapEntry := range result.DebugMap.Entries {
			units[mapEntry.Unit]++
		}
		if units["lib"] == 0 || units[""] == 0 {
			t.Fatalf("debug-map units = %v, want entries for the entry module and for lib", units)
		}
		for _, mapEntry := range result.DebugMap.Entries {
			if strings.HasPrefix(mapEntry.CoreID, "s1:lib:") && mapEntry.Unit != "lib" {
				t.Fatalf("lib entry %s carries unit %q", mapEntry.ID, mapEntry.Unit)
			}
		}
	})

	t.Run("run-interpreter-command", func(t *testing.T) {
		result, err := session.RunInterpreterCommandFile(phase30ByteTwoModules(t))
		if err != nil || result.Status != protocol.StatusPass || len(result.Executions) != 1 {
			t.Fatalf("run = %s, %v", phase30ResultJSON(t, result), err)
		}
	})

	t.Run("run-native", func(t *testing.T) {
		phase30RequireClang(t)
		result, err := session.RunNativeCommandFile(context.Background(), phase30ByteTwoModules(t), native.DefaultRunner())
		if err != nil || result.Status != protocol.StatusPass {
			t.Fatalf("native run = %s, %v", phase30ResultJSON(t, result), err)
		}
	})

	t.Run("build-and-app-run", func(t *testing.T) {
		phase30RequireClang(t)
		output := filepath.Join(t.TempDir(), "app")
		receipt, diagnostics, err := session.BuildApplicationFile(context.Background(), entry, output, native.DefaultRunner())
		if err != nil || len(diagnostics) != 0 {
			t.Fatalf("build = %+v, %v", diagnostics, err)
		}
		entryBytes, readErr := os.ReadFile(entry)
		if readErr != nil {
			t.Fatal(readErr)
		}
		sum := sha256.Sum256(entryBytes)
		if !strings.Contains(receipt.SourceDigest, hex.EncodeToString(sum[:])) {
			t.Fatalf("receipt SourceDigest %q is not the entry digest %x (module hashing is Phase 31)", receipt.SourceDigest, sum)
		}
		var stdout, stderr bytes.Buffer
		if _, runErr := native.RunApplication(context.Background(), output, "21", &stdout, &stderr); runErr != nil {
			t.Fatalf("app run: %v\n%s", runErr, stderr.String())
		}
		if strings.TrimSpace(stdout.String()) != "42" {
			t.Fatalf("app output = %q, want 42", stdout.String())
		}
	})

	t.Run("app-verify", func(t *testing.T) {
		phase30RequireClang(t)
		dir := t.TempDir()
		cases := session.ReplayCases{Schema: session.ReplayCasesSchema, Cases: []session.ReplayCase{{
			ID: "double", Kind: session.ReplayCaseSource, Input: "21",
			Expected:        session.ReplayExpected{Kind: execution.OutcomeReturned, Value: "42"},
			ForeignOutcomes: []session.ReplayForeignOutcome{},
		}}}
		encoded, err := json.Marshal(cases)
		if err != nil {
			t.Fatal(err)
		}
		casesPath := filepath.Join(dir, "cases.json")
		if err := os.WriteFile(casesPath, encoded, 0o600); err != nil {
			t.Fatal(err)
		}
		report, err := session.VerifyApplicationCasesFile(context.Background(), entry, casesPath, filepath.Join(dir, "report.json"), native.DefaultRunner())
		if err != nil || !report.Verified {
			t.Fatalf("app verify verified=%v err=%v report=%+v", report.Verified, err, report)
		}
	})
}

func phase30LibBytes(t *testing.T, entry string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(filepath.Dir(entry), "lib.schway"))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func phase30FindCode(diagnostics []diagnostic.Diagnostic, code string) (diagnostic.Diagnostic, bool) {
	for _, d := range diagnostics {
		if d.Code == code {
			return d, true
		}
	}
	return diagnostic.Diagnostic{}, false
}

func TestPhase30ImportedFileDiagnosticNamesFile(t *testing.T) {
	const entrySource = "module app.main\n\nimport lib {\n  fn relay\n}\n\nexport {\n  fn main\n}\n\nfn main(n: U64) -> U64 {\n  let r = n\n  r\n}\n"
	typeError := writeModuleFixture(t, map[string]string{
		"main.schway": entrySource,
		"lib.schway":  "module lib\n\nexport {\n  fn relay\n}\n\nfn relay(value: U64) -> U64 {\n  let result = missing_name\n  result\n}\n",
	})

	t.Run("type-error-in-lib", func(t *testing.T) {
		result, err := session.CheckCommandFile(typeError)
		if err != nil {
			t.Fatal(err)
		}
		found, ok := phase30FindCode(result.Diagnostics, "name.unknown_place")
		if !ok || found.Unit != "lib" {
			t.Fatalf("diagnostics = %+v, want name.unknown_place in unit lib", result.Diagnostics)
		}
		lib := phase30LibBytes(t, typeError)
		if got := string(lib[found.Primary.Start:found.Primary.End]); got != "missing_name" {
			t.Fatalf("span selects %q of lib.schway, want missing_name", got)
		}
		if !strings.Contains(phase30ResultJSON(t, result), `"unit":"lib"`) {
			t.Fatal("JSON result lacks the lib unit")
		}
		if !strings.Contains(phase30ResultHuman(t, result), " unit=lib [") {
			t.Fatalf("human result lacks the unit: %s", phase30ResultHuman(t, result))
		}
		restore := check.SetSkipUnitStampingForTest(true)
		defer restore()
		skipped, err := session.CheckCommandFile(typeError)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(phase30ResultJSON(t, skipped), `"unit":"lib"`) {
			t.Fatal("control: with unit stamping skipped the lib diagnostic must lose its unit")
		}
	})

	t.Run("entry-error-has-no-unit", func(t *testing.T) {
		path := writeModuleFixture(t, map[string]string{
			"main.schway": strings.Replace(entrySource, "let r = n", "let r = missing_name", 1),
			"lib.schway":  "module lib\n\nexport {\n  fn relay\n}\n\nfn relay(value: U64) -> U64 {\n  let result = value\n  result\n}\n",
		})
		result, err := session.CheckCommandFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(result.Diagnostics) == 0 || strings.Contains(phase30ResultJSON(t, result), `"unit"`) {
			t.Fatalf("an entry-file error carries no unit: %s", phase30ResultJSON(t, result))
		}
	})

	t.Run("parse-error-in-lib", func(t *testing.T) {
		path := writeModuleFixture(t, map[string]string{
			"main.schway": entrySource,
			"lib.schway":  "module lib\n\nexport {\n  fn relay\n\nfn relay(value: U64) -> U64 {\n  let result = value\n  result\n}\n",
		})
		result, err := session.CheckCommandFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var parse *diagnostic.Diagnostic
		for index := range result.Diagnostics {
			if strings.HasPrefix(result.Diagnostics[index].Code, "syntax.") {
				parse = &result.Diagnostics[index]
				break
			}
		}
		if parse == nil || parse.Unit != "lib" {
			t.Fatalf("diagnostics = %+v, want a syntax.* diagnostic in unit lib", result.Diagnostics)
		}
		lib := phase30LibBytes(t, path)
		if parse.Primary.End > len(lib) || string(lib[parse.Primary.Start:parse.Primary.End]) != "fn" {
			t.Fatalf("parse span [%d:%d] does not index the offending token of lib.schway", parse.Primary.Start, parse.Primary.End)
		}
		if !strings.Contains(phase30ResultJSON(t, result), `"unit":"lib"`) {
			t.Fatal("JSON result lacks the lib unit")
		}
		restore := modload.SetSkipParseUnitStampForTest(true)
		defer restore()
		skipped, err := session.CheckCommandFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(phase30ResultJSON(t, skipped), `"unit":"lib"`) {
			t.Fatal("control: with the parse unit stamp skipped the diagnostic must lose its unit")
		}
	})
}

// phase30CloneProgram deep-copies a core program through its JSON form so a
// seeded fault never touches the checked original.
func phase30CloneProgram(t *testing.T, program core.Program) core.Program {
	t.Helper()
	encoded, err := json.Marshal(program)
	if err != nil {
		t.Fatal(err)
	}
	var clone core.Program
	if err := json.Unmarshal(encoded, &clone); err != nil {
		t.Fatal(err)
	}
	return clone
}

func phase30FunctionIndex(t *testing.T, program core.Program, id string) int {
	t.Helper()
	for index, function := range program.Functions {
		if function.ID == id {
			return index
		}
	}
	t.Fatalf("no core function %s", id)
	return -1
}

func TestPhase30PeerRefusalNamesUnit(t *testing.T) {
	t.Run("origin-refusal-from-source", func(t *testing.T) {
		path := writeModuleFixture(t, map[string]string{
			"main.schway": "module app.main\n\nimport lib {\n  fn relay\n}\n\nexport {\n  fn main\n}\n\nfn main(n: U64) -> U64 {\n  let r = n\n  r\n}\n",
			"lib.schway":  "module lib\n\nexport {\n  fn relay\n}\n\nfn relay(value: Buffer) -> Buffer {\n  let result = borrow value\n  result\n}\n",
		})
		result, err := session.CheckCommandFile(path)
		if err != nil {
			t.Fatal(err)
		}
		found, ok := phase30FindCode(result.Diagnostics, "core.origin_omitted")
		if !ok || found.Unit != "lib" {
			t.Fatalf("diagnostics = %+v, want core.origin_omitted in unit lib", result.Diagnostics)
		}
		lib := phase30LibBytes(t, path)
		if got := string(lib[found.Primary.Start:found.Primary.End]); got != "result" {
			t.Fatalf("primary span selects %q of lib.schway, want result", got)
		}
		if len(found.Causes) != 1 || found.Causes[0].Span == nil {
			t.Fatalf("causes = %+v", found.Causes)
		}
		cause := *found.Causes[0].Span
		if got := string(lib[cause.Start:cause.End]); got != "borrow value" {
			t.Fatalf("cause span selects %q of lib.schway, want borrow value", got)
		}
		if !strings.Contains(phase30ResultJSON(t, result), `"unit":"lib"`) {
			t.Fatal("JSON result lacks the lib unit")
		}
	})

	t.Run("corevalidate-refusal-names-unit", func(t *testing.T) {
		checked, err := session.CheckFile(phase30SameNameFixture(t))
		if err != nil || len(checked.Diagnostics) != 0 {
			t.Fatalf("fixture check: %+v %v", checked.Diagnostics, err)
		}
		forged := phase30CloneProgram(t, checked.Program)
		forged.Functions[phase30FunctionIndex(t, forged, "s1:lib:fn:helper")].ReturnType = "Bool"
		validated := corevalidate.Validate(forged)
		if validated.Valid {
			t.Fatal("forged signature was admitted")
		}
		refused := session.PeerRefusalDiagnosticForTest(validated, forged.ModuleID)
		if refused.Unit != "lib" {
			t.Fatalf("corevalidate refusal of a lib function has unit %q, want lib", refused.Unit)
		}
		entryForged := phase30CloneProgram(t, checked.Program)
		entryForged.Functions[phase30FunctionIndex(t, entryForged, "s1:app.main:fn:helper")].ReturnType = "Bool"
		entryRefused := session.PeerRefusalDiagnosticForTest(corevalidate.Validate(entryForged), entryForged.ModuleID)
		if entryRefused.Unit != "" {
			t.Fatalf("corevalidate refusal of an entry function has unit %q, want empty", entryRefused.Unit)
		}
	})

	t.Run("visibility-refusal-names-the-caller-unit", func(t *testing.T) {
		forgeUnexported := func(t *testing.T, fixture, calleeID string) (core.Program, corevalidate.Result) {
			checked, err := session.CheckFile(fixture)
			if err != nil || len(checked.Diagnostics) != 0 {
				t.Fatalf("fixture check: %+v %v", checked.Diagnostics, err)
			}
			forged := phase30CloneProgram(t, checked.Program)
			forged.Functions[phase30FunctionIndex(t, forged, calleeID)].Exported = false
			validated := corevalidate.Validate(forged)
			if validated.Valid || len(validated.Problems) == 0 || validated.Problems[0].Code != "core.callee_not_exported" {
				t.Fatalf("forged export was not refused as core.callee_not_exported: %+v", validated.Problems)
			}
			return forged, validated
		}
		// The entry calls the private-looking lib function: the caller is the
		// entry, so the diagnostic carries no unit.
		forged, validated := forgeUnexported(t, phase30SameNameFixture(t), "s1:lib:fn:run")
		if unit := session.PeerRefusalDiagnosticForTest(validated, forged.ModuleID).Unit; unit != "" {
			t.Fatalf("an entry caller was blamed on unit %q, want the entry (empty)", unit)
		}
		// A library caller into another library: the caller's unit, not the callee's.
		forged, validated = forgeUnexported(t, phase30DiamondFixture(t), "s1:c:fn:cf")
		if unit := session.PeerRefusalDiagnosticForTest(validated, forged.ModuleID).Unit; unit != "a" && unit != "b" {
			t.Fatalf("a library caller was blamed on unit %q, want its own module (a or b)", unit)
		}
		t.Run("detail-attribution-control", func(t *testing.T) {
			defer session.SetPeerDetailAttributionForTest(true)()
			forged, validated := forgeUnexported(t, phase30SameNameFixture(t), "s1:lib:fn:run")
			if unit := session.PeerRefusalDiagnosticForTest(validated, forged.ModuleID).Unit; unit != "lib" {
				t.Fatalf("the control must reproduce the callee blame, got unit %q", unit)
			}
		})
	})

	t.Run("unit-of-function-id", func(t *testing.T) {
		for _, row := range []struct{ id, entry, want string }{
			{"s1:lib:fn:helper", "app.main", "lib"},
			{"s1:lib.a:fn:helper:op:3", "app.main", "lib.a"},
			{"s1:app.main:fn:main", "app.main", ""},
			{"", "app.main", ""},
			{"not-an-id", "app.main", ""},
		} {
			if got := session.UnitOfFunctionIDForTest(row.id, row.entry); got != row.want {
				t.Errorf("unitOfFunctionID(%q, %q) = %q, want %q", row.id, row.entry, got, row.want)
			}
		}
	})
}

// phase30SameNameFixture is the D-12 assumption-delta program: the entry and
// lib each define a private `helper`. The entry's helper forwards to lib's
// exported run, whose own helper does a remainder by a nonzero literal, so the
// two helpers differ in body and in result.
func phase30SameNameFixture(t *testing.T) string {
	t.Helper()
	return writeModuleFixture(t, map[string]string{
		"main.schway": "module app.main\n\nimport lib {\n  fn run\n}\n\nexport {\n  fn main\n}\n\nfn helper(n: U64) -> U64 {\n  let r = run(n)\n  r\n}\n\nfn main(n: U64) -> U64 {\n  let r = helper(n)\n  r\n}\n",
		"lib.schway":  "module lib\n\nexport {\n  fn run\n}\n\nfn helper(n: U64) -> U64 {\n  let r = n % 7\n  r\n}\n\nfn run(n: U64) -> U64 {\n  let h = helper(n)\n  h\n}\n",
	})
}

const (
	phase30LibHelperID   = "s1:lib:fn:helper"
	phase30EntryHelperID = "s1:app.main:fn:helper"
)

// phase30DiamondFixture is entry -> a, b and a, b -> c. The entry's main calls
// a; b is called from an entry-side bridge so all four modules are reachable
// in the merged program.
func phase30DiamondFixture(t *testing.T) string {
	t.Helper()
	return writeModuleFixture(t, map[string]string{
		"main.schway": "module app.main\n\nimport a {\n  fn af\n}\n\nimport b {\n  fn bf\n}\n\nexport {\n  fn main\n}\n\nfn other(n: U64) -> U64 {\n  let r = bf(n)\n  r\n}\n\nfn main(n: U64) -> U64 {\n  let r = af(n)\n  r\n}\n",
		"a.schway":    "module a\n\nimport c {\n  fn cf\n}\n\nexport {\n  fn af\n}\n\nfn af(n: U64) -> U64 {\n  let r = cf(n)\n  r\n}\n",
		"b.schway":    "module b\n\nimport c {\n  fn cf\n}\n\nexport {\n  fn bf\n}\n\nfn bf(n: U64) -> U64 {\n  let r = cf(n)\n  r\n}\n",
		"c.schway":    "module c\n\nexport {\n  fn cf\n}\n\nfn cf(n: U64) -> U64 {\n  let d = n + 1\n  d\n}\n",
	})
}

func TestPhase30DiamondLoadsAndEmitsOnce(t *testing.T) {
	entry := phase30DiamondFixture(t)

	reads := func() map[string]int {
		counts := map[string]int{}
		restore := modload.SetReadObserverForTest(func(module string) { counts[module]++ })
		defer restore()
		if _, err := session.CheckFile(entry); err != nil {
			t.Fatal(err)
		}
		return counts
	}
	for module, count := range reads() {
		if count != 1 {
			t.Errorf("module %q read %d times, want once", module, count)
		}
	}

	result, err := session.CheckCommandFile(entry)
	if err != nil || result.Status != protocol.StatusPass {
		t.Fatalf("diamond check = %s, %v", phase30ResultJSON(t, result), err)
	}
	checked, err := session.CheckFile(entry)
	if err != nil || len(checked.Diagnostics) != 0 {
		t.Fatalf("diamond CheckFile: %+v %v", checked.Diagnostics, err)
	}
	if got := phase30InterpretU64(t, checked.Program, "20"); got != "21" {
		t.Fatalf("diamond interpreter result = %q, want 21", got)
	}
	emitted, err := cgen.EmitApplication(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	definition := regexp.MustCompile(`(?m)^[^\n;]*\bSCHWAY_M_1_C_2_CF\([^;\n]*\)\s*\{\s*$`)
	if count := len(definition.FindAllString(emitted, -1)); count != 1 {
		t.Fatalf("c.cf is defined %d times in the emitted C, want exactly 1", count)
	}

	t.Run("control-diamond-counted-twice", func(t *testing.T) {
		restore := modload.SetDiamondTwiceForTest(true)
		defer restore()
		counts := reads()
		if counts["c"] != 2 {
			t.Fatalf("control: with the diamond-twice mutant engaged c is read %d times, want 2", counts["c"])
		}
	})
}

// phase30OrderFixture builds an entry whose import declarations (and the item
// order inside them) follow the given order. Module names and item names have
// equal lengths, so every entry function keeps its span offsets.
func phase30OrderFixture(t *testing.T, reversed bool) string {
	t.Helper()
	importA := "import a {\n  fn fa\n  fn ga\n}\n\n"
	importB := "import b {\n  fn fb\n  fn gb\n}\n\n"
	if reversed {
		importA = "import a {\n  fn ga\n  fn fa\n}\n\n"
		importB = "import b {\n  fn gb\n  fn fb\n}\n\n"
	}
	imports := importA + importB
	if reversed {
		imports = importB + importA
	}
	return writeModuleFixture(t, map[string]string{
		"main.schway": "module app.main\n\n" + imports + "export {\n  fn main\n}\n\nfn main(n: U64) -> U64 {\n  let r = fa(n)\n  r\n}\n",
		"a.schway":    "module a\n\nexport {\n  fn fa\n  fn ga\n}\n\nfn fa(n: U64) -> U64 {\n  let d = n + 1\n  d\n}\n\nfn ga(n: U64) -> U64 {\n  let d = n + 2\n  d\n}\n",
		"b.schway":    "module b\n\nexport {\n  fn fb\n  fn gb\n}\n\nfn fb(n: U64) -> U64 {\n  let d = n + 3\n  d\n}\n\nfn gb(n: U64) -> U64 {\n  let d = n + 4\n  d\n}\n",
	})
}

func phase30CoreAndC(t *testing.T, path string) (string, string) {
	t.Helper()
	checked, err := session.CheckFile(path)
	if err != nil || len(checked.Diagnostics) != 0 {
		t.Fatalf("CheckFile: %+v %v", checked.Diagnostics, err)
	}
	encoded, err := json.Marshal(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	emitted, err := cgen.EmitApplication(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded), emitted
}

func TestPhase30ReversedImportOrderIdenticalC(t *testing.T) {
	forwardCore, forwardC := phase30CoreAndC(t, phase30OrderFixture(t, false))
	reversedCore, reversedC := phase30CoreAndC(t, phase30OrderFixture(t, true))
	if forwardCore != reversedCore {
		t.Fatalf("reversing the import order changed the core JSON:\n%s\n%s", forwardCore, reversedCore)
	}
	if forwardC != reversedC {
		t.Fatal("reversing the import order changed the emitted C")
	}

	t.Run("control-merge-in-import-order", func(t *testing.T) {
		restore := modload.SetMergeInImportOrderForTest(true)
		defer restore()
		mutantCore, mutantC := phase30CoreAndC(t, phase30OrderFixture(t, false))
		reversedMutantCore, reversedMutantC := phase30CoreAndC(t, phase30OrderFixture(t, true))
		if mutantCore == reversedMutantCore && mutantC == reversedMutantC {
			t.Fatal("control: merging in import order must make the two outputs differ")
		}
	})
}

func phase30FindProblem(problems []corevalidate.Problem, code string) (corevalidate.Problem, bool) {
	for _, problem := range problems {
		if problem.Code == code {
			return problem, true
		}
	}
	return corevalidate.Problem{}, false
}

// phase30RequireAttribution requires the evidence text to name the lib helper
// and not the entry helper.
func phase30RequireAttribution(t *testing.T, label, evidence string) {
	t.Helper()
	if !strings.Contains(evidence, phase30LibHelperID) {
		t.Fatalf("%s: evidence does not name %s:\n%s", label, phase30LibHelperID, evidence)
	}
	if strings.Contains(evidence, phase30EntryHelperID) {
		t.Fatalf("%s: evidence names the entry helper %s:\n%s", label, phase30EntryHelperID, evidence)
	}
}

// phase30IDSetDifference returns the symmetric difference of two ID lists,
// each element recorded as +id (only in faulted) or -id (only in faithful),
// and ~id when an ID moved position.
func phase30IDSetDifference(faithful, faulted []string) []string {
	faithfulIndex := map[string]int{}
	for index, id := range faithful {
		faithfulIndex[id] = index
	}
	faultedIndex := map[string]int{}
	for index, id := range faulted {
		faultedIndex[id] = index
	}
	var difference []string
	for id, index := range faultedIndex {
		if previous, ok := faithfulIndex[id]; !ok {
			difference = append(difference, "+"+id)
		} else if previous != index {
			difference = append(difference, "~"+id)
		}
	}
	for id := range faithfulIndex {
		if _, ok := faultedIndex[id]; !ok {
			difference = append(difference, "-"+id)
		}
	}
	sort.Strings(difference)
	return difference
}

func TestPhase30SameBareNameTwoModules(t *testing.T) {
	entry := phase30SameNameFixture(t)

	var observedPeers []string
	restore := session.SetCheckCommandPeerObserverForTest(func(name string) { observedPeers = append(observedPeers, name) })
	result, err := session.CheckCommandFile(entry)
	restore()
	if err != nil || result.Status != protocol.StatusPass {
		t.Fatalf("same-bare-name check = %s, %v", phase30ResultJSON(t, result), err)
	}
	if strings.Join(observedPeers, ",") != "corevalidate,originvalidate,pathoracle" {
		t.Fatalf("admission consulted peers %v, want corevalidate, originvalidate, pathoracle", observedPeers)
	}

	checked, err := session.CheckFile(entry)
	if err != nil || len(checked.Diagnostics) != 0 {
		t.Fatalf("CheckFile: %+v %v", checked.Diagnostics, err)
	}
	program := checked.Program
	libIndex := phase30FunctionIndex(t, program, phase30LibHelperID)
	entryIndex := phase30FunctionIndex(t, program, phase30EntryHelperID)
	if libIndex == entryIndex || program.Functions[libIndex].Module != "lib" || program.Functions[entryIndex].Module != "" {
		t.Fatalf("two helpers must be two core functions in two modules: %q %q", program.Functions[libIndex].Module, program.Functions[entryIndex].Module)
	}

	t.Run("interpreter-result", func(t *testing.T) {
		// The entry helper forwards to lib.run, whose helper takes n % 7: 20
		// gives 6 only when lib's own helper ran, and the trace must show both
		// helpers executed.
		entryFunction, err := callgraph.EntryFunction(program)
		if err != nil {
			t.Fatal(err)
		}
		pruned, err := callgraph.PruneToReachable(program, entryFunction.ID)
		if err != nil {
			t.Fatal(err)
		}
		applied, err := interp.RunApplication(pruned, entryFunction.Name, "20")
		if err != nil {
			t.Fatal(err)
		}
		if applied.Execution.Outcome.Value != "6" {
			t.Fatalf("interpreter result = %q, want 6", applied.Execution.Outcome.Value)
		}
		trace, _ := json.Marshal(applied.Execution.Events)
		if !strings.Contains(string(trace), phase30LibHelperID) || !strings.Contains(string(trace), phase30EntryHelperID) {
			t.Fatalf("trace does not show both helpers executing: %s", trace)
		}
	})

	t.Run("c-and-debug-map", func(t *testing.T) {
		emitted, err := cgen.EmitApplication(program)
		if err != nil {
			t.Fatal(err)
		}
		libDefinition := regexp.MustCompile(`(?m)^[^\n;]*\bSCHWAY_M_3_LIB_6_HELPER\([^;\n]*\)\s*\{\s*$`)
		if count := len(libDefinition.FindAllString(emitted, -1)); count != 1 {
			t.Fatalf("lib helper is defined %d times in the C, want 1", count)
		}
		helperDefinitions := regexp.MustCompile(`(?m)^[^\n;(]*\b(\w*(?i:helper)\w*)\([^;\n]*\)\s*\{\s*$`).FindAllStringSubmatch(emitted, -1)
		names := map[string]bool{}
		for _, match := range helperDefinitions {
			names[match[1]] = true
		}
		if len(helperDefinitions) != 2 || len(names) != 2 {
			t.Fatalf("C has %d helper definitions over %d names, want two distinct definitions: %v", len(helperDefinitions), len(names), names)
		}
		mapped, err := session.DebugMapCommandFile(entry, "")
		if err != nil || mapped.DebugMap == nil {
			t.Fatalf("debug-map: %v", err)
		}
		units := map[string]string{}
		for _, mapEntry := range mapped.DebugMap.Entries {
			if strings.HasSuffix(mapEntry.CoreID, ":fn:helper") {
				units[mapEntry.CoreID] = mapEntry.Unit
			}
		}
		if len(units) != 2 || units[phase30LibHelperID] != "lib" || units[phase30EntryHelperID] != "" {
			t.Fatalf("debug-map helper units = %v, want entry helper in unit \"\" and lib helper in unit lib", units)
		}
	})

	t.Run("census-corevalidate", func(t *testing.T) {
		faithful := corevalidate.Validate(phase30CloneProgram(t, program))
		if !faithful.Valid {
			t.Fatalf("faithful program refused: %+v", faithful.Problems)
		}
		forged := phase30CloneProgram(t, program)
		forged.Functions[phase30FunctionIndex(t, forged, phase30LibHelperID)].ReturnType = "Bool"
		refused := corevalidate.Validate(forged)
		problem, ok := phase30FindProblem(refused.Problems, "core.return_type_mismatch")
		if refused.Valid || !ok {
			t.Fatalf("forged signature was not refused as core.return_type_mismatch: %+v", refused.Problems)
		}
		if !strings.HasPrefix(problem.Detail, phase30LibHelperID+":") {
			t.Fatalf("detail %q does not begin with %s:", problem.Detail, phase30LibHelperID)
		}
		encoded, _ := json.Marshal(refused.Problems)
		phase30RequireAttribution(t, "corevalidate (containment)", string(encoded))
	})

	t.Run("census-originvalidate", func(t *testing.T) {
		admitted := corevalidate.Validate(phase30CloneProgram(t, program)).Program()
		if problems := originvalidate.ValidatePublished(admitted); len(problems) != 0 {
			t.Fatalf("faithful program refused: %+v", problems)
		}
		forged := phase30CloneProgram(t, program)
		function := &forged.Functions[phase30FunctionIndex(t, forged, phase30LibHelperID)]
		boolType := ""
		for _, fact := range function.Linear.Types {
			if fact.Shape.Constructor == "Bool" {
				boolType = fact.ID
			}
		}
		if boolType == "" {
			t.Fatal("lib helper has no Bool type fact")
		}
		changed := false
		for index := range function.Linear.Operations {
			if function.Linear.Operations[index].Kind == core.OpRemainderU64 {
				function.Linear.Operations[index].TypeID = boolType
				changed = true
			}
		}
		if !changed {
			t.Fatal("lib helper has no remainder operation")
		}
		problems := originvalidate.ValidatePublished(forged)
		if len(problems) == 0 || problems[0].Code != "core.scalar_operation_invalid" || !strings.HasPrefix(problems[0].Detail, phase30LibHelperID+": invalid remainder_u64 operation") {
			t.Fatalf("invalid remainder was not refused as expected: %+v", problems)
		}
		encoded, _ := json.Marshal(problems)
		phase30RequireAttribution(t, "originvalidate (containment)", string(encoded))
	})

	t.Run("census-pathoracle", func(t *testing.T) {
		if err := pathoracle.ValidateLocalOwnerPaths(phase30CloneProgram(t, program)); err != nil {
			t.Fatalf("faithful program refused: %v", err)
		}
		forged := phase30CloneProgram(t, program)
		function := &forged.Functions[phase30FunctionIndex(t, forged, phase30LibHelperID)]
		last := len(function.Linear.Operations) - 1
		function.Linear.Operations[last].Foreign = &core.ForeignOperationContract{
			Symbol: "forged_symbol", ABIType: "forged_abi", Mode: "acquire", ParameterType: "U64", ResultType: "U64", Unwind: "none", NonlocalExit: "none",
		}
		err := pathoracle.ValidateLocalOwnerPaths(forged)
		if err == nil || !strings.HasPrefix(err.Error(), "pathoracle.local_owner_") {
			t.Fatalf("local-owner path on lib helper was not refused: %v", err)
		}
		phase30RequireAttribution(t, "pathoracle (containment)", err.Error())
	})

	t.Run("census-callgraph", func(t *testing.T) {
		faithful, err := callgraph.Order(phase30CloneProgram(t, program))
		if err != nil {
			t.Fatal(err)
		}
		joined := strings.Join(faithful, ",")
		if !strings.Contains(joined, phase30LibHelperID) || !strings.Contains(joined, phase30EntryHelperID) {
			t.Fatalf("faithful order lacks a helper: %v", faithful)
		}
		forged := phase30CloneProgram(t, program)
		function := &forged.Functions[phase30FunctionIndex(t, forged, phase30LibHelperID)]
		function.Linear.Operations = append([]core.LinearOperation{{
			ID: phase30LibHelperID + ":op:99", PointID: phase30LibHelperID + ":point:linear:99", Kind: core.OpCall,
			SourceID: function.Parameter.ID, TargetID: phase30LibHelperID + ":place:3", TypeID: phase30LibHelperID + ":type:0",
			CalleeID: phase30LibHelperID,
		}}, function.Linear.Operations...)
		faulted, err := callgraph.Order(forged)
		// Both helper IDs are in the faithful output, so containment cannot
		// attribute: the evidence is the refusal's cycle witness, else the
		// faithful-vs-faulted ID difference.
		evidence := ""
		if err != nil {
			evidence = err.Error()
		} else {
			evidence = strings.Join(phase30IDSetDifference(faithful, faulted), ",")
		}
		if evidence == "" {
			t.Fatal("the self-call fault left the call-graph order unchanged")
		}
		phase30RequireAttribution(t, "callgraph (diff/witness)", evidence)
	})
}

func TestPhase30SingleFileCommandsRefuseImports(t *testing.T) {
	entry := phase30TwiceFixture(t)
	outDir := t.TempDir()
	refused := func(t *testing.T, result protocol.Result) bool {
		t.Helper()
		_, found := phase30FindCode(result.Diagnostics, "module.imports_not_loaded")
		return found
	}

	t.Run("evidence", func(t *testing.T) {
		_, result, err := session.EvidenceCommandFile(context.Background(), entry)
		if err != nil || !refused(t, result) {
			t.Fatalf("evidence = %s, %v", phase30ResultJSON(t, result), err)
		}
	})
	// explain and query re-derive the diagnostic from the single-file check,
	// which fails closed on an unloaded import; the diagnostic they find by
	// ID is that refusal.
	entryBytes, readErr := os.ReadFile(entry)
	if readErr != nil {
		t.Fatal(readErr)
	}
	fail, found := phase30FindCode(session.Check(entryBytes).Diagnostics, "module.imports_not_loaded")
	if !found {
		t.Fatal("single-file Check did not fail closed on the import")
	}
	t.Run("explain", func(t *testing.T) {
		result, err := session.ExplainCommandFile(entry, fail.ID, 1)
		if err != nil || result.Status != protocol.StatusPass || result.Explain == nil || result.Explain.RootID != fail.ID {
			t.Fatalf("explain = %s, %v", phase30ResultJSON(t, result), err)
		}
		if !strings.Contains(phase30ResultJSON(t, result), "module.imports_not_loaded") {
			t.Fatalf("explain does not name module.imports_not_loaded: %s", phase30ResultJSON(t, result))
		}
	})
	t.Run("query", func(t *testing.T) {
		result, err := session.QueryCommandFile(entry, fail.ID, session.QueryOptions{})
		if err != nil || result.Query == nil || len(result.Query.Facts) == 0 || result.Query.Facts[0].Kind != "module.imports_not_loaded" {
			t.Fatalf("query = %s, %v", phase30ResultJSON(t, result), err)
		}
	})
	t.Run("interface-export", func(t *testing.T) {
		result, err := session.InterfaceExportCommandFile(entry, filepath.Join(outDir, "summary.json"))
		if err != nil || !refused(t, result) {
			t.Fatalf("interface export = %s, %v", phase30ResultJSON(t, result), err)
		}
	})
	t.Run("interface-core", func(t *testing.T) {
		result, err := session.InterfaceCoreCommandFile(entry, filepath.Join(outDir, "core.json"))
		if err != nil || !refused(t, result) {
			t.Fatalf("interface core = %s, %v", phase30ResultJSON(t, result), err)
		}
	})

	t.Run("format-formats-imports", func(t *testing.T) {
		formatted, err := session.FormatFile(entry)
		if err != nil || len(formatted.Diagnostics) != 0 {
			t.Fatalf("FormatFile = %+v, %v", formatted.Diagnostics, err)
		}
		if !bytes.Equal(formatted.Canonical, formatted.Source) || !bytes.Contains(formatted.Canonical, []byte("import lib {\n  fn twice\n}\n")) {
			t.Fatalf("format did not keep the canonical import layout:\n%s", formatted.Canonical)
		}
		command, err := session.FormatCommandFile(entry, true)
		if err != nil || command.Status != protocol.StatusPass {
			t.Fatalf("format --check = %s, %v", phase30ResultJSON(t, command), err)
		}
	})

	t.Run("control-ignore-unloaded-imports", func(t *testing.T) {
		restore := check.SetIgnoreUnloadedImportsForTest(true)
		defer restore()
		_, result, err := session.EvidenceCommandFile(context.Background(), entry)
		if err != nil {
			t.Fatal(err)
		}
		if refused(t, result) {
			t.Fatal("control: with the guard ignored the evidence route must stop reporting module.imports_not_loaded")
		}
	})
}

// phase30WorkModule writes a module of count functions of ten scalar lets each
// (about 356 work units apiece); the exported function is `exported`.
func phase30WorkModule(module, exported, prefix string, count int, withMain bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "module %s\n\n", module)
	if withMain {
		b.WriteString("import lib {\n  fn g0\n}\n\n")
	}
	fmt.Fprintf(&b, "export {\n  fn %s\n}\n\n", exported)
	for i := 0; i < count; i++ {
		fmt.Fprintf(&b, "fn %s%d(n: U64) -> U64 {\n", prefix, i)
		for j := 0; j < 10; j++ {
			fmt.Fprintf(&b, "  let v%d = n + %d\n", j, j)
		}
		b.WriteString("  n\n}\n\n")
	}
	if withMain {
		b.WriteString("fn main(n: U64) -> U64 {\n  n\n}\n")
	}
	return b.String()
}

func TestPhase30ModuleWorkLimit(t *testing.T) {
	build := func(t *testing.T, entryFunctions, libFunctions int) string {
		t.Helper()
		return writeModuleFixture(t, map[string]string{
			"main.schway": phase30WorkModule("app.main", "main", "f", entryFunctions, true),
			"lib.schway":  phase30WorkModule("lib", "g0", "g", libFunctions, false),
		})
	}
	hasLimit := func(diagnostics []diagnostic.Diagnostic) bool {
		_, found := phase30FindCode(diagnostics, "check.work_limit")
		return found
	}

	over := build(t, 150, 150)
	// Each file alone is under the limit; the sum is not.
	for _, name := range []string{"main.schway", "lib.schway"} {
		source, err := os.ReadFile(filepath.Join(filepath.Dir(over), name))
		if err != nil {
			t.Fatal(err)
		}
		if got := session.Check(source).Work; got > session.MaxCheckWork {
			t.Fatalf("%s alone has work %d, want under %d", name, got, session.MaxCheckWork)
		}
	}
	checked, err := session.CheckFile(over)
	if err != nil || !hasLimit(checked.Diagnostics) {
		t.Fatalf("summed module work was not refused: %+v %v", checked.Diagnostics, err)
	}
	if checked.Work <= session.MaxCheckWork {
		t.Fatalf("summed work %d is not over the limit", checked.Work)
	}
	command, err := session.CheckCommandFile(over)
	if err != nil || !hasLimit(command.Diagnostics) {
		t.Fatalf("CheckCommandFile did not refuse with check.work_limit: %s %v", phase30ResultJSON(t, command), err)
	}
	mapped, err := session.DebugMapCommandFile(over, "")
	if err != nil || !hasLimit(mapped.Diagnostics) {
		t.Fatalf("DebugMapCommandFile did not refuse with check.work_limit: %s %v", phase30ResultJSON(t, mapped), err)
	}

	under := build(t, 100, 100)
	accepted, err := session.CheckCommandFile(under)
	if err != nil || accepted.Status != protocol.StatusPass {
		t.Fatalf("a module program under the limit was refused: %s %v", phase30ResultJSON(t, accepted), err)
	}
	if got, _ := session.CheckFile(under); got.Work <= 0 || got.Work > session.MaxCheckWork {
		t.Fatalf("under-limit work = %d", got.Work)
	}
}

func TestPhase30DebugMapNoImportUnguarded(t *testing.T) {
	source := phase30OverWorkNoImportSource()
	path := filepath.Join(t.TempDir(), "over_work.schway")
	if err := os.WriteFile(path, source, 0o644); err != nil {
		t.Fatal(err)
	}
	var golden string
	for _, entry := range phase30LoadGolden(t).Entries {
		if entry.Path == phase30GeneratedOverWorkKey {
			golden = entry.DebugMap
		}
	}
	if golden == "" {
		t.Fatalf("golden has no %s row", phase30GeneratedOverWorkKey)
	}
	result, err := session.DebugMapCommandFile(path, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, found := phase30FindCode(result.Diagnostics, "check.work_limit"); found {
		t.Fatal("the no-import debug-map path must carry no MaxCheckWork guard (D-15)")
	}
	if got := phase30JSONHash(t, result); got != golden {
		t.Fatalf("no-import debug-map hash %s differs from the golden %s", got, golden)
	}

	t.Run("control-guard-no-import", func(t *testing.T) {
		restore := session.SetDebugMapGuardNoImportForTest(true)
		defer restore()
		guarded, err := session.DebugMapCommandFile(path, "")
		if err != nil {
			t.Fatal(err)
		}
		if _, found := phase30FindCode(guarded.Diagnostics, "check.work_limit"); !found {
			t.Fatal("control: with the guard armed the no-import debug map must report check.work_limit")
		}
	})
}

// TestPhase30ErrorsNeverLeakAbsolutePaths is the route-level half of the
// host-path hygiene claim (T-30-31): no refused-module result names the
// TempDir path or its symlink-resolved form, and each message names modules.
func TestPhase30ErrorsNeverLeakAbsolutePaths(t *testing.T) {
	const importing = "module app.main\n\nimport lib {\n  fn twice\n}\n\nexport {\n  fn main\n}\n\nfn main(n: U64) -> U64 {\n  let r = twice(n)\n  r\n}\n"
	rows := []struct {
		name  string
		code  string
		names []string
		build func(t *testing.T) string
	}{
		{"not-found", "module.not_found", []string{"lib"}, func(t *testing.T) string {
			return writeModuleFixture(t, map[string]string{"main.schway": importing})
		}},
		{"path-mismatch", "module.path_mismatch", []string{"lib", "other"}, func(t *testing.T) string {
			return writeModuleFixture(t, map[string]string{
				"main.schway": importing,
				"lib.schway":  strings.Replace(phase30TwiceLib, "module lib", "module other", 1),
			})
		}},
		{"import-cycle", "module.import_cycle", []string{"a", "b"}, func(t *testing.T) string {
			return writeModuleFixture(t, map[string]string{
				"main.schway": "module app.main\n\nimport a {\n  fn fa\n}\n\nexport {\n  fn main\n}\n\nfn main(n: U64) -> U64 {\n  let r = fa(n)\n  r\n}\n",
				"a.schway":    "module a\n\nimport b {\n  fn fb\n}\n\nexport {\n  fn fa\n}\n\nfn fa(n: U64) -> U64 {\n  let r = fb(n)\n  r\n}\n",
				"b.schway":    "module b\n\nimport a {\n  fn fa\n}\n\nexport {\n  fn fb\n}\n\nfn fb(n: U64) -> U64 {\n  let r = fa(n)\n  r\n}\n",
			})
		}},
	}
	if os.PathSeparator == '/' {
		rows = append(rows, struct {
			name  string
			code  string
			names []string
			build func(t *testing.T) string
		}{"symlink", "module.symlink", []string{"lib"}, func(t *testing.T) string {
			entry := writeModuleFixture(t, map[string]string{"main.schway": importing, "real.schway": phase30TwiceLib})
			if err := os.Symlink(filepath.Join(filepath.Dir(entry), "real.schway"), filepath.Join(filepath.Dir(entry), "lib.schway")); err != nil {
				t.Fatalf("symlink creation failed on a unix host: %v", err)
			}
			return entry
		}})
	}

	hostPaths := func(t *testing.T, entry string) []string {
		dir := filepath.Dir(entry)
		paths := []string{dir}
		if resolved, err := filepath.EvalSymlinks(dir); err == nil && resolved != dir {
			paths = append(paths, resolved)
		}
		return paths
	}
	leaks := func(text string, paths []string) bool {
		for _, path := range paths {
			if strings.Contains(text, path) {
				return true
			}
		}
		return false
	}

	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			entry := row.build(t)
			result, err := session.CheckCommandFile(entry)
			if err != nil {
				t.Fatal(err)
			}
			found, ok := phase30FindCode(result.Diagnostics, row.code)
			if !ok {
				t.Fatalf("no %s diagnostic: %s", row.code, phase30ResultJSON(t, result))
			}
			jsonText, humanText := phase30ResultJSON(t, result), phase30ResultHuman(t, result)
			paths := hostPaths(t, entry)
			if leaks(jsonText, paths) || leaks(humanText, paths) {
				t.Fatalf("result leaks a host path %v:\n%s\n%s", paths, jsonText, humanText)
			}
			for _, name := range row.names {
				if !strings.Contains(found.Message, name) {
					t.Errorf("message %q does not name module %q", found.Message, name)
				}
			}
		})
	}

	t.Run("control-format-error-text", func(t *testing.T) {
		entry := rows[0].build(t)
		restore := modload.SetFormatErrorTextForTest(true)
		defer restore()
		result, err := session.CheckCommandFile(entry)
		if err != nil {
			t.Fatal(err)
		}
		if !leaks(phase30ResultJSON(t, result), hostPaths(t, entry)) {
			t.Fatalf("control: with the error-text leak mutant engaged the module.not_found result must contain the TempDir path: %s", phase30ResultJSON(t, result))
		}
	})
}
