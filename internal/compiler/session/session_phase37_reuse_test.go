package session_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/ast"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

var phase37ReuseOriginHashes = map[string]string{
	"examples/crc32/crc32.schway":             "8a8c7570409a6bf96258a35b5964dc899a2347871b564fb59e081beb011388c9",
	"examples/crc32/crc_low.schway":           "797f9cd2e733c3590a593518fe671a1fda553eb6ce22ea50d70f1a26d606dccc",
	"examples/crc32/lib/crc32.schway":         "304d47059ac5f5266776bb250794a3c64ae76ce9fe7b55a2a490facf1e35fb4f",
	"examples/crc32/adapter.c":                "b758836a1d6dcff2fc6b4f903c0046ba3269ec83a8f3329c03553a90a5fbb1ae",
	"examples/crc32/adapter.h":                "2c29a00ed8939668179fddf1c41cdd5db23e5a9d8574b5ec3e841dae729b612d",
	"examples/crc32/file_bytes.bindings.json": "c2f960ff2a92a5d9a60356711812ea60595a2f81cf6c0ff2c4b259ce6380103c",
}

type phase37ReuseFixture struct {
	root, crcRoot, manifest string
}

type phase37ReuseReceipt struct {
	entryModule string
	receipt     native.BuildReceipt
}

// TestPhase37CRC32SingleCopyAndIdentity keeps its census inside the concrete
// CRC closure, then proves a library-only edit changes both importers' module
// and native identities without changing their entry or adapter bytes.
func TestPhase37CRC32SingleCopyAndIdentity(t *testing.T) {
	phase37ReuseRequireTrackedOrigins(t)
	t.Cleanup(func() { phase37ReuseRequireTrackedOrigins(t) })
	root := testsupport.ProjectPath("examples", "crc32")
	phase37ReuseSourceCensus(t, root)
	phase37ReuseHistoricalStepAccounted(t)

	fixture := phase37ReuseClone(t, t.TempDir())
	unchangedPaths := []string{
		"examples/crc32/crc32.schway",
		"examples/crc32/crc_low.schway",
		"examples/crc32/adapter.c",
		"examples/crc32/adapter.h",
		"examples/crc32/file_bytes.bindings.json",
	}
	entryBytes := make(map[string][]byte, len(unchangedPaths))
	for _, relative := range unchangedPaths {
		data, err := os.ReadFile(filepath.Join(fixture.root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		entryBytes[relative] = data
	}

	entries := []string{"crc32", "crc_low"}
	cleanFirst := make(map[string]phase37ReuseReceipt, len(entries))
	for _, name := range entries {
		entryPath := filepath.Join(fixture.crcRoot, name+".schway")
		checked := phase37ReuseCheckAndImport(t, entryPath)
		entryModule := phase37ReuseEntryModule(t, entryPath)
		phase37ReuseRequireOneResolvedCall(t, checked.Program, entryModule)
		first := phase37ReuseBuild(t, fixture, entryPath, entryModule)
		second := phase37ReuseBuild(t, fixture, entryPath, entryModule)
		if first.receipt.InputID != second.receipt.InputID || first.receipt.Identity.InputID() != second.receipt.Identity.InputID() {
			t.Fatalf("no-edit source/closure identity for %s changed: %s/%s vs %s/%s", name, first.receipt.InputID, first.receipt.Identity.InputID(), second.receipt.InputID, second.receipt.Identity.InputID())
		}
		if first.receipt.ExecutableDigest == second.receipt.ExecutableDigest && first.receipt.BuildID != second.receipt.BuildID {
			t.Fatalf("no-edit BuildID changed despite equal executable digests for %s: %s vs %s", name, first.receipt.BuildID, second.receipt.BuildID)
		}
		if first.receipt.ExecutableDigest != second.receipt.ExecutableDigest {
			if first.receipt.Cacheable || second.receipt.Cacheable || first.receipt.DependencyClosure != "incomplete" || second.receipt.DependencyClosure != "incomplete" {
				t.Fatalf("no-edit executable variance for %s escaped the existing non-cacheable/incomplete receipt posture: first=%+v second=%+v", name, first.receipt, second.receipt)
			}
			t.Logf("no-edit native artifact bytes and BuildID varied for %s while InputID stayed stable: executable=%s/%s build=%s/%s; receipts remain non-cacheable with incomplete closure", name, first.receipt.ExecutableDigest, second.receipt.ExecutableDigest, first.receipt.BuildID, second.receipt.BuildID)
		}
		if !reflect.DeepEqual(first.receipt.Modules, second.receipt.Modules) {
			t.Fatalf("no-edit clean module digests for %s changed: %+v vs %+v", name, first.receipt.Modules, second.receipt.Modules)
		}
		cleanFirst[name] = first
	}

	libraryPath := filepath.Join(fixture.crcRoot, "lib", "crc32.schway")
	beforeLibrary, err := os.ReadFile(libraryPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(beforeLibrary), "crc = crc ^ 0xFFFFFFFF") != 1 {
		t.Fatal("library result-bit mutation anchor is not unique")
	}
	afterLibrary := strings.Replace(string(beforeLibrary), "crc = crc ^ 0xFFFFFFFF", "crc = crc ^ 0xFFFFFFFE", 1)
	if err := os.WriteFile(libraryPath, []byte(afterLibrary), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, relative := range unchangedPaths {
		got, err := os.ReadFile(filepath.Join(fixture.root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, entryBytes[relative]) {
			t.Fatalf("library-only edit changed immutable entry/adapter bytes %s", relative)
		}
	}

	for _, name := range entries {
		entryPath := filepath.Join(fixture.crcRoot, name+".schway")
		checked := phase37ReuseCheckAndImport(t, entryPath)
		phase37ReuseRequireOneResolvedCall(t, checked.Program, cleanFirst[name].entryModule)
		changed := phase37ReuseBuild(t, fixture, entryPath, cleanFirst[name].entryModule)
		phase37ReuseRequireLibraryInvalidation(t, cleanFirst[name], changed)
	}
}

func phase37ReuseSourceCensus(t *testing.T, crcRoot string) {
	t.Helper()
	var sources []string
	err := filepath.WalkDir(crcRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && filepath.Ext(path) == ".schway" {
			relative, err := filepath.Rel(testsupport.ProjectPath("."), path)
			if err != nil {
				return err
			}
			sources = append(sources, filepath.ToSlash(relative))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"examples/crc32/crc32.schway",
		"examples/crc32/crc_low.schway",
		"examples/crc32/lib/crc32.schway",
	}
	if !reflect.DeepEqual(sources, want) {
		t.Fatalf("scoped CRC source closure = %v, want exactly %v", sources, want)
	}

	libraryPath := filepath.Join(crcRoot, "lib", "crc32.schway")
	library := phase37ReuseParseFile(t, libraryPath)
	crc32Count, crcByteCount := 0, 0
	crc32Loops, crcByteLoops := 0, 0
	var crc32Function, crcByteFunction *ast.FuncDecl
	for _, function := range library.Funcs {
		switch function.Name {
		case "crc32":
			crc32Count++
			crc32Function = &function
			crc32Loops += phase37ReuseASTLoops(function.Body.Linear)
		case "crc_byte":
			crcByteCount++
			crcByteFunction = &function
			crcByteLoops += phase37ReuseASTLoops(function.Body.Linear)
		}
	}
	if library.Module != "lib.crc32" || crc32Count != 1 || crcByteCount != 1 || crc32Loops != 1 || crcByteLoops != 1 {
		t.Fatalf("shared library AST module=%q crc32=%d loops=%d crc_byte=%d loops=%d; want one byte loop in crc32 and one bit loop in crc_byte", library.Module, crc32Count, crc32Loops, crcByteCount, crcByteLoops)
	}
	if crc32Function == nil || crcByteFunction == nil || !phase37ReuseHasLoop(crc32Function.Body.Linear.Statements, "i", "n") || !phase37ReuseHasLoop(crcByteFunction.Body.Linear.Statements, "round", "8") {
		t.Fatal("shared library must keep the byte loop in crc32 and the exact eight-round loop in crc_byte")
	}
	if calls := phase37ReuseCallsInLoop(crc32Function.Body.Linear.Statements, "i", "n", "crc_byte"); calls != 1 {
		t.Fatalf("crc32 byte-loop helper calls=%d; want exactly one crc_byte call inside the byte loop", calls)
	}

	for _, name := range []string{"crc32", "crc_low"} {
		entryPath := filepath.Join(crcRoot, name+".schway")
		entry := phase37ReuseParseFile(t, entryPath)
		if len(entry.Imports) != 1 || entry.Imports[0].Module != "lib.crc32" || len(entry.Imports[0].Items) != 1 || entry.Imports[0].Items[0].Kind != "fn" || entry.Imports[0].Items[0].Name != "crc32" {
			t.Fatalf("%s AST imports=%+v; want exactly one import of lib.crc32.crc32", name, entry.Imports)
		}
		if phase37ReuseASTLoopsInProgram(entry) != 0 {
			t.Fatalf("%s entry contains a copied CRC loop", name)
		}
		for _, function := range entry.Funcs {
			if function.Name == "crc32" || function.Name == "crc_byte" {
				t.Fatalf("%s entry contains copied CRC function %s", name, function.Name)
			}
		}
	}
}

func phase37ReuseHistoricalStepAccounted(t *testing.T) {
	t.Helper()
	path := testsupport.ProjectPath("examples", "helper_calls", "crc_step.schway")
	program := phase37ReuseParseFile(t, path)
	stepCount, fullCRCCount := 0, 0
	for _, function := range program.Funcs {
		if function.Name == "crc_step" {
			stepCount++
		}
		if function.Name == "crc32" {
			fullCRCCount++
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if stepCount != 1 || fullCRCCount != 0 || !strings.Contains(string(data), "0xEDB88320") || strings.Contains(string(data), "byte_at(") || strings.Contains(string(data), "byte_len(") {
		t.Fatalf("historical helper_calls/crc_step.schway no longer has its scoped helper-only shape: functions=%+v", program.Funcs)
	}
}

func phase37ReuseParseFile(t *testing.T, path string) ast.Program {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(data)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("parse %s: diagnostics=%+v", path, parsed.Diagnostics)
	}
	return parsed.Program
}

func phase37ReuseASTLoops(body *ast.LinearBody) int {
	if body == nil {
		return 0
	}
	return phase37ReuseASTStatementLoops(body.Statements)
}

func phase37ReuseHasLoop(statements []ast.Statement, iterator, bound string) bool {
	for _, statement := range statements {
		if statement.Kind == "while" && statement.Expr != nil && statement.Expr.Kind == "less" &&
			statement.Expr.Left != nil && statement.Expr.Left.Kind == "name" && statement.Expr.Left.Name == iterator &&
			statement.Expr.Right != nil && statement.Expr.Right.Kind == "number" && statement.Expr.Right.Value == bound {
			return true
		}
		if phase37ReuseHasLoop(statement.Then, iterator, bound) || phase37ReuseHasLoop(statement.Else, iterator, bound) {
			return true
		}
	}
	return false
}

func phase37ReuseCallsInLoop(statements []ast.Statement, iterator, bound, callee string) int {
	for _, statement := range statements {
		if statement.Kind == "while" && statement.Expr != nil && statement.Expr.Kind == "less" &&
			statement.Expr.Left != nil && statement.Expr.Left.Kind == "name" && statement.Expr.Left.Name == iterator &&
			statement.Expr.Right != nil && statement.Expr.Right.Kind == "name" && statement.Expr.Right.Name == bound {
			return phase37ReuseCallsInStatements(statement.Then, callee)
		}
		if count := phase37ReuseCallsInLoop(statement.Then, iterator, bound, callee); count > 0 {
			return count
		}
		if count := phase37ReuseCallsInLoop(statement.Else, iterator, bound, callee); count > 0 {
			return count
		}
	}
	return 0
}

func phase37ReuseCallsInStatements(statements []ast.Statement, callee string) int {
	total := 0
	for _, statement := range statements {
		total += phase37ReuseCallsInExpr(statement.Expr, callee)
		total += phase37ReuseCallsInStatements(statement.Then, callee)
		total += phase37ReuseCallsInStatements(statement.Else, callee)
	}
	return total
}

func phase37ReuseCallsInExpr(expression *ast.Expr, callee string) int {
	if expression == nil {
		return 0
	}
	total := 0
	if expression.Kind == "call" && expression.Name == callee {
		total++
	}
	return total + phase37ReuseCallsInExpr(expression.Left, callee) + phase37ReuseCallsInExpr(expression.Right, callee)
}

func phase37ReuseASTLoopsInProgram(program ast.Program) int {
	total := 0
	for _, function := range program.Funcs {
		total += phase37ReuseASTLoops(function.Body.Linear)
	}
	return total
}

func phase37ReuseASTStatementLoops(statements []ast.Statement) int {
	total := 0
	for _, statement := range statements {
		if statement.Kind == "while" {
			total++
		}
		total += phase37ReuseASTStatementLoops(statement.Then)
		total += phase37ReuseASTStatementLoops(statement.Else)
	}
	return total
}

func phase37ReuseCheckAndImport(t *testing.T, entryPath string) session.CheckResult {
	t.Helper()
	checked, err := session.CheckFile(entryPath)
	if err != nil || len(checked.Diagnostics) != 0 {
		t.Fatalf("check %s: err=%v diagnostics=%+v", entryPath, err, checked.Diagnostics)
	}
	return checked
}

func phase37ReuseEntryModule(t *testing.T, entryPath string) string {
	t.Helper()
	program := phase37ReuseParseFile(t, entryPath)
	if program.Module == "" {
		t.Fatalf("entry %s has no module identity", entryPath)
	}
	return program.Module
}

func phase37ReuseRequireOneResolvedCall(t *testing.T, program core.Program, entryModule string) {
	t.Helper()
	if len(program.Imports) != 1 || program.Imports[0] != (core.ModuleImport{From: "", Module: "lib.crc32", Name: "crc32"}) {
		t.Fatalf("merged import rows for %s = %+v; want one entry import from lib.crc32.crc32 and no self-import", entryModule, program.Imports)
	}
	var implementation *core.Function
	for index := range program.Functions {
		function := &program.Functions[index]
		if function.Name == "crc32" {
			if implementation != nil {
				t.Fatalf("merged %s closure contains multiple crc32 functions", entryModule)
			}
			implementation = function
		}
	}
	if implementation == nil || implementation.Module != "lib.crc32" {
		t.Fatalf("merged %s crc32 implementation = %+v; want one function from lib.crc32", entryModule, implementation)
	}
	callCount := 0
	for _, function := range program.Functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind == core.OpCall && operation.CalleeID == implementation.ID {
				callCount++
			}
		}
	}
	if callCount != 1 {
		t.Fatalf("merged %s closure has %d resolved calls to imported crc32, want one", entryModule, callCount)
	}
}

func phase37ReuseClone(t *testing.T, root string) phase37ReuseFixture {
	t.Helper()
	files := []string{
		"examples/crc32/crc32.schway",
		"examples/crc32/crc_low.schway",
		"examples/crc32/lib/crc32.schway",
		"examples/crc32/adapter.c",
		"examples/crc32/adapter.h",
		"examples/crc32/file_bytes.bindings.json",
	}
	for _, relative := range files {
		data, err := os.ReadFile(testsupport.ProjectPath(filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		destination := filepath.Join(root, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(destination, data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	crcRoot := filepath.Join(root, "examples", "crc32")
	return phase37ReuseFixture{root: root, crcRoot: crcRoot, manifest: filepath.Join(crcRoot, "file_bytes.bindings.json")}
}

func phase37ReuseBuild(t *testing.T, fixture phase37ReuseFixture, entryPath, entryModule string) phase37ReuseReceipt {
	t.Helper()
	artifact := filepath.Join(fixture.root, ".phase37-build-"+filepath.Base(entryPath))
	receipt, diagnostics, err := session.BuildApplicationFile(context.Background(), entryPath, artifact, native.DefaultRunner(), fixture.manifest)
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("native build %s: err=%v diagnostics=%+v", entryPath, err, diagnostics)
	}
	if receipt.Identity == nil || receipt.BuildID == "" || receipt.BuildID != receipt.Identity.ID() || receipt.InputID == "" || receipt.InputID != receipt.Identity.InputID() {
		t.Fatalf("build receipt for %s lacks consistent identity: build=%q identity=%+v", entryPath, receipt.BuildID, receipt.Identity)
	}
	if receipt.ExecutableDigest == "" || receipt.Identity.ExecutableDigest != receipt.ExecutableDigest {
		t.Fatalf("build receipt for %s does not bind its executable digest consistently: receipt=%q identity=%q", entryPath, receipt.ExecutableDigest, receipt.Identity.ExecutableDigest)
	}
	if receipt.Cacheable || receipt.DependencyClosure != "incomplete" {
		t.Fatalf("build receipt for %s changed existing cacheability posture: cacheable=%t dependency_closure=%q", entryPath, receipt.Cacheable, receipt.DependencyClosure)
	}
	if !reflect.DeepEqual(receipt.Modules, receipt.Identity.Modules) {
		t.Fatalf("receipt module digests differ from identity modules for %s: receipt=%+v identity=%+v", entryPath, receipt.Modules, receipt.Identity.Modules)
	}
	if len(receipt.Modules) != 2 {
		t.Fatalf("module closure for %s has %d rows, want entry plus lib.crc32: %+v", entryPath, len(receipt.Modules), receipt.Modules)
	}
	if _, ok := phase37ReuseModule(t, receipt.Modules, "lib.crc32"); !ok {
		t.Fatalf("module closure for %s lacks lib.crc32: %+v", entryPath, receipt.Modules)
	}
	if _, ok := phase37ReuseModule(t, receipt.Modules, entryModule); !ok {
		t.Fatalf("module closure for %s lacks entry module %s: %+v", entryPath, entryModule, receipt.Modules)
	}
	return phase37ReuseReceipt{entryModule: entryModule, receipt: receipt}
}

func phase37ReuseRequireLibraryInvalidation(t *testing.T, clean, changed phase37ReuseReceipt) {
	t.Helper()
	if clean.entryModule != changed.entryModule {
		t.Fatalf("entry module changed across library edit: %q to %q", clean.entryModule, changed.entryModule)
	}
	cleanLibrary, cleanHasLibrary := phase37ReuseModule(t, clean.receipt.Modules, "lib.crc32")
	changedLibrary, changedHasLibrary := phase37ReuseModule(t, changed.receipt.Modules, "lib.crc32")
	cleanEntry, cleanHasEntry := phase37ReuseModule(t, clean.receipt.Modules, clean.entryModule)
	changedEntry, changedHasEntry := phase37ReuseModule(t, changed.receipt.Modules, changed.entryModule)
	if !cleanHasLibrary || !changedHasLibrary || !cleanHasEntry || !changedHasEntry {
		t.Fatalf("module closure rows missing: clean=%+v changed=%+v", clean.receipt.Modules, changed.receipt.Modules)
	}
	if cleanLibrary.Module != changedLibrary.Module || cleanLibrary.Path != changedLibrary.Path || cleanLibrary.Digest == changedLibrary.Digest {
		t.Fatalf("library digest did not change under stable module identity: clean=%+v changed=%+v", cleanLibrary, changedLibrary)
	}
	if cleanEntry != changedEntry {
		t.Fatalf("entry module digest changed under a library-only edit: clean=%+v changed=%+v", cleanEntry, changedEntry)
	}
	if clean.receipt.BuildID == changed.receipt.BuildID || clean.receipt.Identity.ID() == changed.receipt.Identity.ID() {
		t.Fatalf("library-only edit did not invalidate build identity: clean=%s/%s changed=%s/%s", clean.receipt.BuildID, clean.receipt.Identity.ID(), changed.receipt.BuildID, changed.receipt.Identity.ID())
	}
	if clean.receipt.InputID == changed.receipt.InputID || clean.receipt.Identity.InputID() == changed.receipt.Identity.InputID() {
		t.Fatalf("library-only edit did not invalidate source/closure identity: clean=%s/%s changed=%s/%s", clean.receipt.InputID, clean.receipt.Identity.InputID(), changed.receipt.InputID, changed.receipt.Identity.InputID())
	}
	if clean.receipt.ExecutableDigest == changed.receipt.ExecutableDigest {
		t.Fatalf("library-only result-bit edit did not change the built executable digest: %s", clean.receipt.ExecutableDigest)
	}
}

func phase37ReuseModule(t *testing.T, modules []native.ModuleDigest, name string) (native.ModuleDigest, bool) {
	t.Helper()
	var match native.ModuleDigest
	found := false
	for _, module := range modules {
		if module.Module == name {
			if found {
				t.Fatalf("module %s occurs more than once in closure rows %+v", name, modules)
			}
			match, found = module, true
		}
	}
	return match, found
}

func phase37ReuseRequireTrackedOrigins(t *testing.T) {
	t.Helper()
	for path, want := range phase37ReuseOriginHashes {
		data, err := os.ReadFile(testsupport.ProjectPath(filepath.FromSlash(path)))
		if err != nil {
			t.Fatalf("read tracked CRC origin %s: %v", path, err)
		}
		digest := sha256.Sum256(data)
		if got := hex.EncodeToString(digest[:]); got != want {
			t.Fatalf("tracked CRC origin %s sha256=%s, want pinned %s", path, got, want)
		}
	}
}
