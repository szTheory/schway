package originvalidate_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// originValidateForbiddenImports is D-10-16's closed forbidden set for
// BOTH guards below: `check` and `ast` (the pre-existing set) PLUS
// `corevalidate` -- Criterion 1 requires the OpCall walk to import
// "neither check NOR corevalidate," and before this plan neither guard's
// forbidden list could catch a corevalidate import at all. Deliberately
// NOT `/compiler/callgraph`: originvalidate's own callgraph import (line
// 18) is real and kept (D-10-19, PHASE-10-DEBT.md); the asymmetry with
// corevalidate's own guard (which does forbid callgraph) is recorded, not
// resolved, in that debt row.
//
// TestOriginValidatorImportsStayIndependent and
// TestOriginValidateImportsNeitherCheckNorAst deliberately stay two
// near-identical guards rather than one shared table (D-10-18): the
// existing redundancy across four-plus such guards in this codebase means
// blinding the mechanism requires multiple coordinated deletions, not one
// — that redundancy is worth more here than the maintainability win a
// single shared table would buy. Do not "clean this up" into one table.
var originValidateForbiddenImports = []string{"/compiler/check", "/compiler/ast", "/compiler/corevalidate"}

// directImportViolation is the go/parser ImportsOnly directory scan shared
// by both guards below: it reads originvalidate's own Go source files'
// import lists (never assumed from a doc comment) and returns the first
// forbidden import path found, or "" if none. Kept as a direct-import scan
// even though assertTransitiveImportsStayIndependent below also runs a
// transitive check over the SAME forbidden set -- the redundancy is
// deliberate (D-10-17): this scan catches a forbidden import added
// directly to this package; the transitive check catches one added
// indirectly, through a helper package that itself imports something
// forbidden.
func directImportViolation(t *testing.T, dir string, forbidden []string) string {
	t.Helper()
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
			for _, bad := range forbidden {
				if strings.HasSuffix(path, bad) {
					return entry.Name() + " imports " + path
				}
			}
		}
	}
	return ""
}

// transitiveImportViolation is D-10-17's own suffix-matching predicate,
// factored out so BOTH the real `go list -deps` scan below and its own
// negative control (TestTransitiveImportsGuardCanFail) exercise the
// IDENTICAL logic -- proving the negative control is testing the real
// predicate, not a second, drifting copy of it.
func transitiveImportViolation(deps []string, forbidden []string) string {
	for _, dep := range deps {
		for _, bad := range forbidden {
			if strings.HasSuffix(dep, bad) {
				return dep
			}
		}
	}
	return ""
}

// transitiveImportsViolation is D-10-17's hardening of the direct-import
// scan above: nobody adds an import of `check` to a re-deriver on purpose;
// they add a helper package that itself imports `check`. This shells out
// to `go list -deps`, which ships with the toolchain CI already requires
// (adds no dependency), and asserts no line of originvalidate's own
// transitive dependency closure has a forbidden suffix.
// maxGoListDepsOutputBytes bounds transitiveImportsViolation's captured
// `go list -deps` output. native_test.go's scanUnboundedSpawns lint (D-02-01)
// forbids the Output()/CombinedOutput() unbounded merged-buffer shape for
// every process this repo spawns; Plan 10-02 introduced this call site with
// that shape and Plan 10-03 logged it to deferred-items.md after fixing the
// identical pattern it had introduced in pathoracle_test.go. This is that
// deferred fix, applied at the Phase 10 Wave 2 post-merge gate. The bound and
// the exec.CommandContext deadline below are kept local to this file, mirroring
// pathoracle_test.go's own choice rather than reusing testsupport's unexported
// boundedWriter.
const maxGoListDepsOutputBytes = 1 << 20 // 1 MiB: far more than any real dependency-path listing

// goListDepsTimeout bounds the spawn itself, satisfying the lint's
// exec.CommandContext requirement.
const goListDepsTimeout = 2 * time.Minute

func phase23LocalOwnerProgram(t *testing.T) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "phase23", "file_byte.schway"))
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("Phase 23 source failed to parse: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("Phase 23 source failed checking: %+v", checked.Diagnostics)
	}
	return checked.Program
}

func removePhase23Releases(program *core.Program) {
	function := &program.Functions[0]
	operations := function.Linear.Operations[:0]
	for _, operation := range function.Linear.Operations {
		if operation.Kind != core.OpRelease {
			operations = append(operations, operation)
		}
	}
	function.Linear.Operations = operations
}

func TestPhase23OriginPeerTracksAcquiredOwnerThroughBorrowedUse(t *testing.T) {
	program := phase23LocalOwnerProgram(t)
	if problems := originvalidate.ValidatePublished(program); len(problems) != 0 {
		t.Fatalf("valid acquire/use/release core rejected: %+v", problems)
	}

	removePhase23Releases(&program)
	if problems := originvalidate.ValidatePublished(program); len(problems) == 0 {
		t.Fatal("originvalidate accepted successful acquisition after every release was removed")
	}
}

func phase23MutationIndex(program *core.Program, mode string) int {
	for index, operation := range program.Functions[0].Linear.Operations {
		if operation.Foreign != nil && operation.Foreign.Mode == mode {
			return index
		}
	}
	return -1
}

func renumberPhase23Operations(function *core.Function) {
	for index := range function.Linear.Operations {
		function.Linear.Operations[index].ID = fmt.Sprintf("%s:op:%d", function.ID, index)
		function.Linear.Operations[index].PointID = fmt.Sprintf("%s:point:linear:%d", function.ID, index)
	}
}

func phase23OwnerMutations() []struct {
	name string
	edit func(*core.Program) bool
} {
	return []struct {
		name string
		edit func(*core.Program) bool
	}{
		{"all releases deleted", func(program *core.Program) bool {
			release := phase23MutationIndex(program, "consume")
			if release < 0 {
				return false
			}
			operations := make([]core.LinearOperation, 0, len(program.Functions[0].Linear.Operations)-1)
			for _, operation := range program.Functions[0].Linear.Operations {
				if operation.Kind != core.OpRelease {
					operations = append(operations, operation)
				}
			}
			program.Functions[0].Linear.Operations = operations
			renumberPhase23Operations(&program.Functions[0])
			return true
		}},
		{"release before borrowed use", func(program *core.Program) bool {
			function := &program.Functions[0]
			borrow, release := phase23MutationIndex(program, "borrow"), phase23MutationIndex(program, "consume")
			if borrow < 0 || release < 0 {
				return false
			}
			function.Linear.Operations[borrow], function.Linear.Operations[release] = function.Linear.Operations[release], function.Linear.Operations[borrow]
			renumberPhase23Operations(function)
			return true
		}},
		{"duplicate release", func(program *core.Program) bool {
			function := &program.Functions[0]
			release := phase23MutationIndex(program, "consume")
			if release < 0 {
				return false
			}
			operations := function.Linear.Operations
			copyOfRelease := operations[release]
			mutated := make([]core.LinearOperation, 0, len(operations)+1)
			mutated = append(mutated, operations[:release+1]...)
			mutated = append(mutated, copyOfRelease)
			mutated = append(mutated, operations[release+1:]...)
			function.Linear.Operations = mutated
			renumberPhase23Operations(function)
			return true
		}},
		{"wrong owner operation identity", func(program *core.Program) bool {
			borrow, release := phase23MutationIndex(program, "borrow"), phase23MutationIndex(program, "consume")
			if borrow < 0 || release < 0 {
				return false
			}
			program.Functions[0].Linear.Operations[release].ReleasesOperationID = program.Functions[0].Linear.Operations[borrow].ID
			return true
		}},
		{"wrong resource place", func(program *core.Program) bool {
			borrow, release := phase23MutationIndex(program, "borrow"), phase23MutationIndex(program, "consume")
			if borrow < 0 || release < 0 {
				return false
			}
			program.Functions[0].Linear.Operations[release].SourceID = program.Functions[0].Linear.Operations[borrow].TargetID
			return true
		}},
		{"fabricated release identity", func(program *core.Program) bool {
			release := phase23MutationIndex(program, "consume")
			if release < 0 {
				return false
			}
			program.Functions[0].Linear.Operations[release].ReleasesOperationID = "fabricated-acquire"
			return true
		}},
		{"release omitted after typed use error", func(program *core.Program) bool {
			borrow := phase23MutationIndex(program, "borrow")
			if borrow < 0 || program.Functions[0].Linear.Operations[borrow].Foreign.Fails != "UseError" || phase23MutationIndex(program, "acquire") < 0 {
				return false
			}
			operations := make([]core.LinearOperation, 0, len(program.Functions[0].Linear.Operations)-1)
			for _, operation := range program.Functions[0].Linear.Operations {
				if operation.Kind != core.OpRelease {
					operations = append(operations, operation)
				}
			}
			program.Functions[0].Linear.Operations = operations
			renumberPhase23Operations(&program.Functions[0])
			return true
		}},
		{"wrong use symbol", func(program *core.Program) bool {
			index := phase23MutationIndex(program, "borrow")
			if index < 0 {
				return false
			}
			program.Functions[0].Linear.Operations[index].Foreign.Symbol = "schway_file_byte_acquire"
			return true
		}},
		{"wrong use ABI", func(program *core.Program) bool {
			index := phase23MutationIndex(program, "borrow")
			if index < 0 {
				return false
			}
			program.Functions[0].Linear.Operations[index].Foreign.ABIType = "wrong_file_byte_use_fn"
			return true
		}},
		{"missing use error status", func(program *core.Program) bool {
			index := phase23MutationIndex(program, "borrow")
			if index < 0 {
				return false
			}
			program.Functions[0].Linear.Operations[index].Foreign.Fails = ""
			return true
		}},
		{"mismatched release allocator", func(program *core.Program) bool {
			index := phase23MutationIndex(program, "consume")
			if index < 0 {
				return false
			}
			program.Functions[0].Linear.Operations[index].Foreign.Allocator = "other_allocator"
			program.Functions[0].Linear.Operations[index].Allocator = "other_allocator"
			return true
		}},
	}
}

func TestPhase23ResourceMutationOriginvalidate(t *testing.T) {
	for _, mutation := range phase23OwnerMutations() {
		t.Run(mutation.name, func(t *testing.T) {
			program := phase23LocalOwnerProgram(t)
			if !mutation.edit(&program) {
				t.Fatal("mutation target was not reached")
			}
			if problems := originvalidate.ValidatePublished(program); len(problems) == 0 {
				t.Fatal("originvalidate accepted the reached ownership mutation")
			}
		})
	}
}

// boundedGoListWriter caps the bytes captured from `go list -deps`,
// mirroring testsupport's own boundedWriter shape without depending on its
// unexported type.
type boundedGoListWriter struct {
	buffer     bytes.Buffer
	overflowed bool
}

func (w *boundedGoListWriter) Write(data []byte) (int, error) {
	if w.buffer.Len()+len(data) > maxGoListDepsOutputBytes {
		w.overflowed = true
		remaining := maxGoListDepsOutputBytes - w.buffer.Len()
		if remaining > 0 {
			w.buffer.Write(data[:remaining])
		}
		return len(data), nil
	}
	w.buffer.Write(data)
	return len(data), nil
}

func transitiveImportsViolation(t *testing.T, forbidden []string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), goListDepsTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "list", "-deps", "github.com/szTheory/schway/internal/compiler/originvalidate")
	cmd.Dir = testsupport.ProjectPath()
	stdout := &boundedGoListWriter{}
	cmd.Stdout = stdout
	if err := cmd.Run(); err != nil {
		t.Fatalf("go list -deps: %v", err)
	}
	if stdout.overflowed {
		t.Fatalf("go list -deps produced more than %d bytes; dependency listing is implausibly large", maxGoListDepsOutputBytes)
	}
	deps := strings.Split(strings.TrimSpace(stdout.buffer.String()), "\n")
	return transitiveImportViolation(deps, forbidden)
}

// TestOriginValidateImportsNeitherCheckNorAst is 03-06-01's structural
// falsifier for the plan's binding prohibition: originvalidate decides from
// the typed-core artifact alone and must never import the checker or the
// AST package. Reading the actual import lists (not trusting a doc comment)
// is the same technique the codebase already applies to enforce boundaries
// mechanically rather than by convention.
// TestOriginValidatorImportsStayIndependent is Task 04-06-02's extension of
// TestOriginValidateImportsNeitherCheckNorAst (T-04-37): the new
// checkForeignOriginOmitted path (D-04-28) must derive its refusal from the
// core artifact alone, exactly like every other check in this package --
// this re-runs the identical file-scan so the new function is covered by
// construction rather than by a second, drifting assertion.
//
// 10-02 Task 3 (D-10-16/D-10-17): both guards now ALSO forbid corevalidate,
// and both ALSO run the transitive go/list-deps scan below the direct scan
// -- the direct `go/parser` ImportsOnly scan is hardened, not replaced
// (both still run).
//
// Residual weakness (D-10-20, deliberately stated rather than papered
// over): this guard lives in the package it polices, so a single commit
// could add a forbidden import here and edit this very assertion in the
// same commit. Transitive checking does not fix that; neither would
// consolidating the two guards into one table.
func TestOriginValidatorImportsStayIndependent(t *testing.T) {
	dir := testsupport.ProjectPath("internal", "compiler", "originvalidate")
	if violation := directImportViolation(t, dir, originValidateForbiddenImports); violation != "" {
		t.Fatalf("%s, which originvalidate (including the foreign-origin path) must never depend on", violation)
	}
	if violation := transitiveImportsViolation(t, originValidateForbiddenImports); violation != "" {
		t.Fatalf("originvalidate transitively imports %s, which it must never depend on", violation)
	}
}

func TestOriginValidateImportsNeitherCheckNorAst(t *testing.T) {
	dir := testsupport.ProjectPath("internal", "compiler", "originvalidate")
	if violation := directImportViolation(t, dir, originValidateForbiddenImports); violation != "" {
		t.Fatalf("%s, which originvalidate must never depend on", violation)
	}
	if violation := transitiveImportsViolation(t, originValidateForbiddenImports); violation != "" {
		t.Fatalf("originvalidate transitively imports %s, which it must never depend on", violation)
	}
}

// TestTransitiveImportsGuardCanFail is D-10-17's negative control: the
// suffix-matching predicate both guards' transitive checks share
// (transitiveImportViolation) must actually report a violation over a
// synthetic dependency list containing a forbidden path -- proving the
// transitive guard can go red, not merely that it has never yet found
// anything.
func TestTransitiveImportsGuardCanFail(t *testing.T) {
	synthetic := []string{
		"github.com/szTheory/schway/internal/compiler/originvalidate",
		"github.com/szTheory/schway/internal/compiler/core",
		"github.com/szTheory/schway/internal/compiler/corevalidate",
	}
	if got := transitiveImportViolation(synthetic, originValidateForbiddenImports); got == "" {
		t.Fatal("expected the synthetic dependency list's forbidden corevalidate entry to be flagged")
	}
}

// TestPhase17OriginPeerDirectionalAbilities proves that the origin peer
// derives Drops from the parameter fact and Fresh from the separate return
// fact.  The two-type tracer deliberately grants both facts Drop; removing it
// only from the return fact is therefore the narrow mutation that catches a
// peer which accidentally reuses the parameter lookup for both contracts.
func TestPhase17OriginPeerDirectionalAbilities(t *testing.T) {
	program := phase17OriginProgram(t)
	baseline, err := originvalidate.BuildInterface(program)
	if err != nil {
		t.Fatalf("baseline BuildInterface: %v", err)
	}
	mutated := cloneOriginProgram(t, program)
	for functionIndex := range mutated.Functions {
		function := &mutated.Functions[functionIndex]
		if function.Linear == nil {
			continue
		}
		for factIndex := range function.Linear.Types {
			fact := &function.Linear.Types[factIndex]
			if fact.ID == function.ID+":type:1" {
				fact.Abilities = nil
			}
		}
	}

	summary, err := originvalidate.BuildInterface(mutated)
	if err != nil {
		t.Fatalf("BuildInterface: %v", err)
	}
	for _, signature := range summary.Functions {
		if len(signature.Parameters) != 1 {
			t.Fatalf("%s parameter contract missing: %+v", signature.Name, signature)
		}
		var baselineSignature core.FunctionSignature
		for _, candidate := range baseline.Functions {
			if candidate.ID == signature.ID {
				baselineSignature = candidate
				break
			}
		}
		if baselineSignature.Parameters[0] != signature.Parameters[0] {
			t.Fatalf("%s parameter contract changed under return-only ability mutation: before=%+v after=%+v", signature.Name, baselineSignature.Parameters[0], signature.Parameters[0])
		}
		if baselineSignature.Return.Fresh && signature.Return.Fresh {
			t.Fatalf("%s return Fresh reused parameter ability after return-only mutation: %+v", signature.Name, signature.Return)
		}
	}
}

func TestPhase17OriginPeerPublishedTwoTypeContract(t *testing.T) {
	program := phase17OriginProgram(t)
	summary, err := originvalidate.BuildInterface(program)
	if err != nil {
		t.Fatalf("BuildInterface: %v", err)
	}
	for _, function := range program.Functions {
		for _, signature := range summary.Functions {
			if signature.ID != function.ID {
				continue
			}
			if len(signature.Parameters) != 1 || signature.Parameters[0].Type != function.Parameter.Type || signature.Return.Type != function.ReturnType {
				t.Fatalf("%s collapsed its declared parameter/return contract: %+v", function.Name, signature)
			}
			break
		}
	}
}

func TestPhase17OriginPeerReturnOnlyMutation(t *testing.T) {
	program := phase17OriginProgram(t)
	baseline, err := originvalidate.BuildInterface(program)
	if err != nil {
		t.Fatalf("baseline BuildInterface: %v", err)
	}

	restore := originvalidate.SetPhase17ReturnLookupFaultForTest(true)
	t.Cleanup(restore)
	faulted, err := originvalidate.BuildInterface(program)
	if err != nil {
		t.Fatalf("faulted BuildInterface: %v", err)
	}
	for _, before := range baseline.Functions {
		for _, after := range faulted.Functions {
			if after.ID != before.ID {
				continue
			}
			if len(before.Parameters) != 1 || len(after.Parameters) != 1 || before.Parameters[0] != after.Parameters[0] {
				t.Fatalf("%s parameter contract changed under return-only fault: before=%+v after=%+v", before.Name, before.Parameters, after.Parameters)
			}
			if before.Return.Fresh && after.Return.Fresh {
				t.Fatalf("%s return-only fault left Fresh true: %+v", before.Name, after.Return)
			}
		}
	}
	restore()
	restore()
	restored, err := originvalidate.BuildInterface(program)
	if err != nil {
		t.Fatalf("restored BuildInterface: %v", err)
	}
	for _, signature := range restored.Functions {
		var hasLinear bool
		for _, function := range program.Functions {
			if function.ID == signature.ID {
				hasLinear = function.Linear != nil
			}
		}
		if !hasLinear {
			continue
		}
		if !signature.Return.Fresh {
			t.Fatalf("%s return lookup did not restore: %+v", signature.Name, signature.Return)
		}
	}
}

func TestPhase17OriginPeerIndependenceBoundary(t *testing.T) {
	directory := testsupport.ProjectPath("internal", "compiler", "originvalidate")
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", entry.Name(), err)
		}
		if violation := phase17OriginForbiddenImport(file); violation != "" {
			t.Fatalf("%s imports forbidden derivation dependency %q", entry.Name(), violation)
		}
		var callers []token.Pos
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "SetPhase17ReturnLookupFaultForTest" {
				callers = append(callers, call.Pos())
			}
			return true
		})
		if len(callers) != 0 {
			t.Fatalf("%s has production caller(s) of return-only test control at %v", entry.Name(), callers)
		}
	}

	seeded, err := parser.ParseFile(token.NewFileSet(), "seed.go", `package originvalidate
import _ "github.com/szTheory/schway/internal/compiler/corevalidate"
`, parser.ImportsOnly)
	if err != nil {
		t.Fatal(err)
	}
	if got := phase17OriginForbiddenImport(seeded); got == "" {
		t.Fatal("seeded prohibited derivation import was not detected")
	}
}

func phase17OriginForbiddenImport(file *ast.File) string {
	for _, imported := range file.Imports {
		value := strings.Trim(imported.Path.Value, `"`)
		if strings.Contains(value, "compiler/check") || strings.Contains(value, "compiler/corevalidate") || strings.Contains(strings.ToLower(value), "returncontract") {
			return value
		}
	}
	return ""
}

func phase17OriginProgram(t testing.TB) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase17", "return_type_tracer.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("canonical source failed to check: %+v", checked.Diagnostics)
	}
	return checked.Program
}

func cloneOriginProgram(t testing.TB, program core.Program) core.Program {
	t.Helper()
	encoded, err := json.Marshal(program)
	if err != nil {
		t.Fatal(err)
	}
	var cloned core.Program
	if err := json.Unmarshal(encoded, &cloned); err != nil {
		t.Fatal(err)
	}
	return cloned
}

func honestProgram(t testing.TB, fixture string) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase3", fixture))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	if checked.Program.Functions[0].PublicOrigin == nil {
		t.Fatalf("expected PublicOrigin on the honestly-checked function")
	}
	return checked.Program
}

// TestOriginUnderstatedRejected is 03-06-02's falsifier for the "omitted"
// defect: a declared origin set that omits a path the body actually derives
// from is rejected as core.origin_understated, detected purely by
// source-blind recomputation from the typed core (never by trusting the
// declaration). The dishonest declaration is injected by mutating an
// honestly-checked program's own fact — check.go's honest producer can never
// construct this shape itself, exactly as OV-02-01's mutation-kill precedent
// establishes for a different fact.
func TestOriginUnderstatedRejected(t *testing.T) {
	program := honestProgram(t, "public_view_understated.schway")
	program.Functions[0].PublicOrigin = &core.PublicOrigin{Paths: []string{}, Access: "shared"}
	problems := originvalidate.ValidatePublished(program)
	if len(problems) != 1 || problems[0].Code != "core.origin_understated" {
		t.Fatalf("expected exactly core.origin_understated, got %+v", problems)
	}
}

// TestOriginAccessMismatchRejected is 03-06-02's falsifier for the
// "impossible" defect: a declared access mode the body cannot produce is
// rejected as core.origin_access_mismatch by the same recomputation.
func TestOriginAccessMismatchRejected(t *testing.T) {
	program := honestProgram(t, "public_view_impossible.schway")
	program.Functions[0].PublicOrigin = &core.PublicOrigin{Paths: []string{"buffer"}, Access: "exclusive"}
	problems := originvalidate.ValidatePublished(program)
	if len(problems) != 1 || problems[0].Code != "core.origin_access_mismatch" {
		t.Fatalf("expected exactly core.origin_access_mismatch, got %+v", problems)
	}
}

// TestMixedAccessChainDerivesShared is 03-08-01's falsifier for CR-01: a
// reborrow chain whose closest-to-return hop is a shared reborrow of an
// exclusively-borrowed place must derive access "shared" — the hop nearest
// the returned place decides, not an earlier hop further up the chain. This
// is the honest recomputation (no mutation): check.go's own producer
// constructs the mismatching *declaration* on this fixture, but
// RecomputeOrigin's own body-derived answer must still be correct in
// isolation.
func TestMixedAccessChainDerivesShared(t *testing.T) {
	program := honestProgram(t, "public_view_mixed_access.schway")
	paths, access, ok := originvalidate.RecomputeOrigin(program.Functions[0], nil)
	if !ok {
		t.Fatalf("expected RecomputeOrigin to succeed")
	}
	if access != "shared" {
		t.Fatalf("expected access shared (closest-to-return hop), got %q", access)
	}
	if len(paths) != 1 || paths[0] != "buffer" {
		t.Fatalf("expected origin path [buffer], got %+v", paths)
	}
}

// TestMixedAccessChainRejectedAsAccessMismatch is 03-08-01's falsifier for
// CR-01's downstream effect: ValidatePublished must reject the fixture's
// declared borrow mut(buffer) (exclusive) against a body that only ever
// derives shared, naming both modes in the detail. Unlike
// TestOriginAccessMismatchRejected, no mutation is needed here — check.go's
// honest producer already constructs this exact mismatching declaration
// from the source's own `borrow mut(buffer)` return-type annotation.
//
// The reverse hop ordering (an exclusive reborrow of a shared loan) is
// exercised here too: it must be derived correctly (shared decides nothing
// once overridden by a closer exclusive hop) or rejected with a named code —
// never silently accepted as matching a declaration it does not support.
func TestMixedAccessChainRejectedAsAccessMismatch(t *testing.T) {
	program := honestProgram(t, "public_view_mixed_access.schway")
	problems := originvalidate.ValidatePublished(program)
	if len(problems) != 1 || problems[0].Code != "core.origin_access_mismatch" {
		t.Fatalf("expected exactly core.origin_access_mismatch, got %+v", problems)
	}
	if !strings.Contains(problems[0].Detail, `"exclusive"`) || !strings.Contains(problems[0].Detail, `"shared"`) {
		t.Fatalf("expected detail to name both declared and body-derived modes, got %q", problems[0].Detail)
	}

	// Reverse ordering: an exclusive reborrow of a shared loan. RecomputeOrigin
	// still derives the closest access mode, but publication independently
	// rejects this conflict because a shared parent cannot authorize an
	// exclusive child.
	reverseSource := `module owned.public_view_mixed_access_reverse

export {
  fn view
}

fn view(buffer: Buffer) -> borrow mut(buffer) Buffer {
  let shared = borrow buffer
  let exclusive = borrow mut shared
  exclusive
}
`
	checkedReverse := session.Check([]byte(reverseSource))
	if len(checkedReverse.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics on reverse-ordering fixture: %+v", checkedReverse.Diagnostics)
	}
	if paths, access, ok := originvalidate.RecomputeOrigin(checkedReverse.Program.Functions[0], nil); !ok || access != "exclusive" || len(paths) != 1 || paths[0] != "buffer" {
		t.Fatalf("expected reverse-ordering chain to derive exclusive, got paths=%+v access=%q ok=%v", paths, access, ok)
	}
	reverseProblems := originvalidate.ValidatePublished(checkedReverse.Program)
	if len(reverseProblems) != 1 || reverseProblems[0].Code != "core.borrow_conflict" {
		t.Fatalf("expected shared-to-exclusive reborrow to fail independently with core.borrow_conflict, got %+v", reverseProblems)
	}
}

// TestOmittedOriginRejected is 03-09-01's falsifier for D-03-02/GAP 2: a
// function with NO declared origin, whose body's only return is a
// borrow-derived place, must be refused publication with
// core.origin_omitted — the category ValidatePublished's old
// PublicOrigin == nil short-circuit made definitionally unreachable.
func TestOmittedOriginRejected(t *testing.T) {
	program := honestOmittedProgram(t, "public_view_omitted.schway")
	problems := originvalidate.ValidatePublished(program)
	if len(problems) != 1 || problems[0].Code != "core.origin_omitted" {
		t.Fatalf("expected exactly core.origin_omitted, got %+v", problems)
	}
	if !strings.Contains(problems[0].Detail, "buffer") || !strings.Contains(problems[0].Detail, `"exclusive"`) {
		t.Fatalf("expected detail to name the derived paths and access, got %q", problems[0].Detail)
	}
}

// TestOwnedReturnWithNoOriginStillPublishes is 03-09-01's falsifier for the
// non-over-firing half of the same behavior: a function with no declared
// origin whose return is NOT borrow-derived (RecomputeOrigin reports
// not-ok) must still publish with no problems, whether the function is
// straight-line (shared_shared_accept.schway, an owned take/return) or
// match-bodied (borrowed_view.schway, a branch function whose every arm
// returns an owned take).
func TestOwnedReturnWithNoOriginStillPublishes(t *testing.T) {
	for _, fixture := range []string{"shared_shared_accept.schway", "borrowed_view.schway"} {
		program := honestOmittedProgram(t, fixture)
		if problems := originvalidate.ValidatePublished(program); len(problems) != 0 {
			t.Fatalf("%s: expected no problems for an owned return with no declared origin, got %+v", fixture, problems)
		}
	}
}

// TestExclusiveBorrowCleanShapeChecksButCannotPublish is 03-09-01's
// falsifier for the fixture_disposition: session.Check on the exact
// exclusive_borrow_clean / relay source still returns zero diagnostics,
// while ValidatePublished on that same checked program now returns
// core.origin_omitted — the gate lives on the publication path only. 07-02
// D-07-44: the source now lives at testdata/phase07/clean_but_unpublishable.schway,
// the single extracted source of truth check_exclusive_test.go also reads,
// rather than a third embedded copy here.
func TestExclusiveBorrowCleanShapeChecksButCannotPublish(t *testing.T) {
	cleanSource, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase07", "clean_but_unpublishable.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(cleanSource)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("expected session.Check to still accept the exclusive_borrow_clean shape unchanged, got %+v", checked.Diagnostics)
	}
	problems := originvalidate.ValidatePublished(checked.Program)
	if len(problems) != 1 || problems[0].Code != "core.origin_omitted" {
		t.Fatalf("expected exactly core.origin_omitted on the publication path, got %+v", problems)
	}
}

// honestOmittedProgram is like honestProgram but does NOT assert a non-nil
// PublicOrigin — it is used for fixtures that deliberately declare none
// (public_view_omitted.schway) or whose shape never carries one
// (match-bodied borrowed_view.schway).
func honestOmittedProgram(t testing.TB, fixture string) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase3", fixture))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	return checked.Program
}

// TestPerReturnOriginsCoverEveryArm is Task 03-10-01's falsifier for the
// core defect: RecomputeOriginPerReturn must return one element per
// core.OpReturn, not just the first. A straight-line function collapses to
// exactly one element; the multi-arm fixture collapses to exactly two, one
// per arm, and only the second (the borrow-returning arm) is borrow-derived.
func TestPerReturnOriginsCoverEveryArm(t *testing.T) {
	straightLine := honestProgram(t, "public_view.schway")
	straightOrigins := originvalidate.RecomputeOriginPerReturn(straightLine.Functions[0], nil)
	if len(straightOrigins) != 1 {
		t.Fatalf("expected exactly 1 per-return origin for a straight-line function, got %+v", straightOrigins)
	}
	if !straightOrigins[0].Derived || straightOrigins[0].Access != "shared" {
		t.Fatalf("expected the straight-line origin to be shared-derived, got %+v", straightOrigins[0])
	}

	program := honestOmittedProgram(t, "public_view_multi_arm_omitted.schway")
	origins := originvalidate.RecomputeOriginPerReturn(program.Functions[0], nil)
	if len(origins) != 2 {
		t.Fatalf("expected exactly 2 per-return origins (one per arm), got %+v", origins)
	}
	derivedCount := 0
	for _, origin := range origins {
		if origin.Derived {
			derivedCount++
			if origin.Access != "shared" {
				t.Fatalf("expected the borrow-derived arm's access to be shared, got %+v", origin)
			}
		}
	}
	if derivedCount != 1 {
		t.Fatalf("expected exactly 1 borrow-derived return among the 2 arms, got %d in %+v", derivedCount, origins)
	}
}

func TestPhase19ConstantOriginStopsAtRoot(t *testing.T) {
	source := []byte("module phase19.constant_origin\nexport {\n  fn main\n}\nfn main(input: Byte) -> U64 {\n  let view = borrow input\n  let value = 42\n  value\n}\n")
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("constant origin witness failed admission: %+v", checked.Diagnostics)
	}
	function := checked.Program.Functions[0]
	operations := append([]core.LinearOperation(nil), function.Linear.Operations...)
	var constIndex int = -1
	var borrowedPlace string
	for index, operation := range operations {
		if operation.Kind == core.OpBorrowShared {
			borrowedPlace = operation.TargetID
		}
		if operation.Kind == core.OpConst {
			constIndex = index
		}
	}
	if constIndex < 0 || borrowedPlace == "" {
		t.Fatalf("witness lacks borrow and constant operations: %+v", operations)
	}
	// A forged source field must not let an origin walk cross the constant
	// root into a borrow. Core admission rejects this mutation separately.
	operations[constIndex].SourceID = borrowedPlace
	linear := *function.Linear
	linear.Operations = operations
	function.Linear = &linear
	origins := originvalidate.RecomputeOriginPerReturn(function, nil)
	if len(origins) != 1 || origins[0].Derived || origins[0].Access != "" {
		t.Fatalf("constant inherited a forged parameter origin: %+v", origins)
	}
}

// TestMultiArmOmittedOriginRejected is Task 03-10-01's falsifier for
// SC3/SC4: a match-bodied function whose first arm returns owned and second
// arm returns a live borrow, with no declared origin (match functions cannot
// declare one), must be refused publication with core.origin_omitted — not
// silently accepted because RecomputeOrigin only looked at the first arm.
func TestMultiArmOmittedOriginRejected(t *testing.T) {
	program := honestOmittedProgram(t, "public_view_multi_arm_omitted.schway")
	paths, access, ok := originvalidate.RecomputeOrigin(program.Functions[0], nil)
	if !ok {
		t.Fatalf("expected RecomputeOrigin to succeed on the multi-arm leak")
	}
	if access != "shared" {
		t.Fatalf("expected combined access shared, got %q", access)
	}
	if len(paths) != 1 || paths[0] != "flag" {
		t.Fatalf("expected origin path [flag], got %+v", paths)
	}
	problems := originvalidate.ValidatePublished(program)
	if len(problems) != 1 || problems[0].Code != "core.origin_omitted" {
		t.Fatalf("expected exactly core.origin_omitted, got %+v", problems)
	}
}

// TestOwnedArmsStillPublish is Task 03-10-01's non-over-firing falsifier:
// the widened, every-OpReturn walk must not newly refuse a match-bodied
// function whose every arm returns an owned value.
func TestOwnedArmsStillPublish(t *testing.T) {
	for _, fixture := range []string{"branch_view.schway", "borrowed_view.schway"} {
		program := honestOmittedProgram(t, fixture)
		if problems := originvalidate.ValidatePublished(program); len(problems) != 0 {
			t.Fatalf("%s: expected no problems for every-arm-owned, got %+v", fixture, problems)
		}
	}
}

// TestMultiArmAccessConflictDerivesNeitherArm is Task 03-10-02's falsifier
// for the combination law's case 3: two arms deriving different access
// modes must combine to the AccessConflicting sentinel, with paths unioned
// (not duplicated) — neither arm's own answer wins.
func TestMultiArmAccessConflictDerivesNeitherArm(t *testing.T) {
	program := honestOmittedProgram(t, "public_view_multi_arm_access_conflict.schway")
	origins := originvalidate.RecomputeOriginPerReturn(program.Functions[0], nil)
	if len(origins) != 2 {
		t.Fatalf("expected exactly 2 per-return origins, got %+v", origins)
	}
	seenShared, seenExclusive := false, false
	for _, origin := range origins {
		if !origin.Derived {
			t.Fatalf("expected both arms borrow-derived, got %+v", origin)
		}
		switch origin.Access {
		case "shared":
			seenShared = true
		case "exclusive":
			seenExclusive = true
		}
	}
	if !seenShared || !seenExclusive {
		t.Fatalf("expected one shared and one exclusive arm, got %+v", origins)
	}
	paths, access, ok := originvalidate.RecomputeOrigin(program.Functions[0], nil)
	if !ok || access != originvalidate.AccessConflicting {
		t.Fatalf("expected AccessConflicting, got access=%q ok=%v", access, ok)
	}
	if len(paths) != 1 || paths[0] != "flag" {
		t.Fatalf("expected exactly one unioned path [flag], got %+v", paths)
	}
}

// TestUnionPathsAreNotDuplicated pins the same non-duplication requirement
// directly, independent of the access-mode assertions above.
func TestUnionPathsAreNotDuplicated(t *testing.T) {
	program := honestOmittedProgram(t, "public_view_multi_arm_access_conflict.schway")
	paths, _, ok := originvalidate.RecomputeOrigin(program.Functions[0], nil)
	if !ok {
		t.Fatalf("expected RecomputeOrigin to succeed")
	}
	seen := make(map[string]int)
	for _, path := range paths {
		seen[path]++
	}
	for path, count := range seen {
		if count != 1 {
			t.Fatalf("path %q duplicated %d times in %+v", path, count, paths)
		}
	}
}

// TestMultiArmAccessConflictRejectedWhenDeclaredShared is Task 03-10-02's
// falsifier: an undeclared conflicting-arms function is refused with
// core.origin_omitted (the same code single-arm omission uses); a
// conflicting-arms function DECLARING "shared" is refused with
// core.origin_access_mismatch, whose Detail names the sentinel.
func TestMultiArmAccessConflictRejectedWhenDeclaredShared(t *testing.T) {
	undeclared := honestOmittedProgram(t, "public_view_multi_arm_access_conflict.schway")
	problems := originvalidate.ValidatePublished(undeclared)
	if len(problems) != 1 || problems[0].Code != "core.origin_omitted" {
		t.Fatalf("expected exactly core.origin_omitted for the undeclared conflicting-arms function, got %+v", problems)
	}

	declaredShared := honestOmittedProgram(t, "public_view_multi_arm_access_conflict.schway")
	declaredShared.Functions[0].PublicOrigin = &core.PublicOrigin{Paths: []string{"flag"}, Access: "shared"}
	problems = originvalidate.ValidatePublished(declaredShared)
	if len(problems) != 1 || problems[0].Code != "core.origin_access_mismatch" {
		t.Fatalf("expected exactly core.origin_access_mismatch, got %+v", problems)
	}
	if !strings.Contains(problems[0].Detail, originvalidate.AccessConflicting) {
		t.Fatalf("expected detail to name the conflicting sentinel, got %q", problems[0].Detail)
	}
}

// TestMultiArmAccessConflictRejectedWhenDeclaredExclusive mirrors the shared
// case with the other declarable access mode.
func TestMultiArmAccessConflictRejectedWhenDeclaredExclusive(t *testing.T) {
	declaredExclusive := honestOmittedProgram(t, "public_view_multi_arm_access_conflict.schway")
	declaredExclusive.Functions[0].PublicOrigin = &core.PublicOrigin{Paths: []string{"flag"}, Access: "exclusive"}
	problems := originvalidate.ValidatePublished(declaredExclusive)
	if len(problems) != 1 || problems[0].Code != "core.origin_access_mismatch" {
		t.Fatalf("expected exactly core.origin_access_mismatch, got %+v", problems)
	}
	if !strings.Contains(problems[0].Detail, originvalidate.AccessConflicting) {
		t.Fatalf("expected detail to name the conflicting sentinel, got %q", problems[0].Detail)
	}
}

// TestDeclaredConflictingAccessIsRefused is Task 03-10-02's falsifier for
// the declared-access domain check: a declared PublicOrigin.Access equal to
// the sentinel string itself is refused with core.origin_access_mismatch
// BEFORE any comparison with the recomputed answer — the sentinel is never
// a declarable mode, so a mutated summary cannot declare it and match a
// conflicting recomputation.
func TestDeclaredConflictingAccessIsRefused(t *testing.T) {
	// public_view.schway's recomputed access is "shared", not the sentinel, so
	// declaring the sentinel there is caught by the ordinary
	// declared-vs-recomputed mismatch comparison regardless of whether a
	// dedicated domain check exists. To actually falsify the domain check,
	// declare the sentinel on the ONE fixture whose own recomputed answer IS
	// the sentinel (public_view_multi_arm_access_conflict.schway) — without a
	// domain check running BEFORE the comparison, declared == recomputed and
	// this would incorrectly report no problems at all.
	conflicting := honestOmittedProgram(t, "public_view_multi_arm_access_conflict.schway")
	conflicting.Functions[0].PublicOrigin = &core.PublicOrigin{Paths: []string{"flag"}, Access: originvalidate.AccessConflicting}
	problems := originvalidate.ValidatePublished(conflicting)
	if len(problems) != 1 || problems[0].Code != "core.origin_access_mismatch" {
		t.Fatalf("expected exactly core.origin_access_mismatch for a declared sentinel access, got %+v", problems)
	}

	program := honestProgram(t, "public_view.schway")
	program.Functions[0].PublicOrigin = &core.PublicOrigin{Paths: []string{"buffer"}, Access: originvalidate.AccessConflicting}
	problems = originvalidate.ValidatePublished(program)
	if len(problems) != 1 || problems[0].Code != "core.origin_access_mismatch" {
		t.Fatalf("expected exactly core.origin_access_mismatch for a declared sentinel access, got %+v", problems)
	}
}

// TestPublishedOriginConsistentWithEveryReturn is Task 03-10-02's
// generalized every-return invariant (the tripwire against a future
// regression back to a single-return assumption): for every function in
// every testdata/phase3 fixture, RecomputeOrigin's combined triple must be
// exactly the conservative combination of RecomputeOriginPerReturn's
// elements.
func TestPublishedOriginConsistentWithEveryReturn(t *testing.T) {
	dir := testsupport.ProjectPath("testdata", "phase3")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	checkedAny := false
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".schway") {
			continue
		}
		source, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		checked := session.Check(source)
		if len(checked.Diagnostics) != 0 {
			continue
		}
		for _, function := range checked.Program.Functions {
			checkedAny = true
			perReturn := originvalidate.RecomputeOriginPerReturn(function, nil)
			var wantPaths []string
			pathSeen := make(map[string]bool)
			wantAccess := ""
			derivedCount := 0
			conflict := false
			for _, origin := range perReturn {
				if !origin.Derived {
					continue
				}
				derivedCount++
				if wantAccess == "" {
					wantAccess = origin.Access
				} else if origin.Access != wantAccess {
					conflict = true
				}
				for _, path := range origin.Paths {
					if !pathSeen[path] {
						pathSeen[path] = true
						wantPaths = append(wantPaths, path)
					}
				}
			}
			gotPaths, gotAccess, gotOK := originvalidate.RecomputeOrigin(function, nil)
			if derivedCount == 0 {
				if gotOK {
					t.Fatalf("%s/%s: expected not-ok for an owned function, got paths=%+v access=%q", entry.Name(), function.ID, gotPaths, gotAccess)
				}
				continue
			}
			if !gotOK {
				t.Fatalf("%s/%s: expected ok=true for a borrow-derived function, got not-ok", entry.Name(), function.ID)
			}
			expectedAccess := wantAccess
			if conflict {
				expectedAccess = originvalidate.AccessConflicting
			}
			if gotAccess != expectedAccess {
				t.Fatalf("%s/%s: expected combined access %q, got %q", entry.Name(), function.ID, expectedAccess, gotAccess)
			}
			if len(gotPaths) != len(wantPaths) {
				t.Fatalf("%s/%s: expected paths %+v, got %+v", entry.Name(), function.ID, wantPaths, gotPaths)
			}
			for index, path := range wantPaths {
				if gotPaths[index] != path {
					t.Fatalf("%s/%s: expected paths %+v, got %+v", entry.Name(), function.ID, wantPaths, gotPaths)
				}
			}
		}
	}
	if !checkedAny {
		t.Fatal("expected at least one function from testdata/phase3 to be checked")
	}
}

// TestStaleSummaryRejectedBeforeOtherChecks is 03-06-02's falsifier for
// T-03-03: a summary whose recorded digest does not match the core artifact
// it is checked against is rejected before any origin or access question is
// even asked — CheckSummary never unmarshals coreBytes into a struct that
// could carry a body, so the ordering is structural, not merely sequenced.
func TestStaleSummaryRejectedBeforeOtherChecks(t *testing.T) {
	program := honestProgram(t, "public_view.schway")
	summary, err := originvalidate.BuildInterface(program)
	if err != nil {
		t.Fatal(err)
	}
	// 07-01 Task 3 fills ClosureDigest for real; here it only needs to pass
	// DecodeInterface's shape check so CheckSummary's own staleness check
	// (the thing this test asserts) is what actually rejects the document.
	summary.Functions[0].ClosureDigest = validClosureDigestPlaceholder
	summaryBytes, err := json.Marshal(summary)
	if err != nil {
		t.Fatal(err)
	}
	staleCore := []byte(`{"not":"the real core artifact"}`)
	if _, err := originvalidate.CheckSummary(summaryBytes, staleCore); err == nil {
		t.Fatal("expected a stale-summary rejection")
	} else if code := errorCode(err); code != "origin.stale_summary" {
		t.Fatalf("expected origin.stale_summary, got %q (%v)", code, err)
	}

	// A summary whose declared origin would itself be dishonest must still be
	// rejected for staleness first, proving the digest check runs before any
	// origin-shaped decision — CheckSummary has no other check to reorder
	// against, which is itself the point: there IS no origin/access check
	// left to run once the digest fails.
	dishonest := summary
	dishonest.Functions[0].Return.Mode = "exclusive"
	dishonest.Functions[0].Return.Paths = []string{"nonexistent"}
	dishonestBytes, err := json.Marshal(dishonest)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := originvalidate.CheckSummary(dishonestBytes, staleCore); errorCode(err) != "origin.stale_summary" {
		t.Fatalf("expected origin.stale_summary ahead of any origin content, got %v", err)
	}
}

// TestOriginEscapeIsNamed is 03-06-02's falsifier naming the accepted
// TestCheckSummaryRoutesThroughDecodeInterface is 07-01 Task 2's Test 5
// (D-07-36): a /1 document missing Mode is refused when passed to
// CheckSummary, asserted through CheckSummary itself, not only through
// core.DecodeInterface directly -- this is the exact gap codex found
// ("CheckSummary unmarshals straight into core.Interface and checks only
// JSON validity and CoreDigest").
func TestCheckSummaryRoutesThroughDecodeInterface(t *testing.T) {
	missingModeDoc := []byte(`{
		"schema": "lang.interface/1",
		"module_id": "m",
		"core_digest": "` + validClosureDigestPlaceholder + `",
		"functions": [{
			"id": "f", "name": "f",
			"parameters": [{"id":"p","name":"p","type":"Byte","mode":"","drops":false}],
			"return": {"type":"Byte","mode":"owned","paths":[],"fresh":false},
			"abilities": [], "callable": false,
			"foreign": {"allocator":"","unwind":"","nonlocal_exit":""},
			"closure_digest": "` + validClosureDigestPlaceholder + `"
		}]
	}`)
	if _, err := originvalidate.CheckSummary(missingModeDoc, []byte(`{}`)); err == nil {
		t.Fatal("expected CheckSummary to refuse a /1 document with a missing Mode")
	} else if code := errorCode(err); code != "core.interface_missing_field" {
		t.Fatalf("expected core.interface_missing_field, got %q (%v)", code, err)
	}
}

// TestCheckSummaryRefusesV0Document is D-07-36/T-07-02's CheckSummary-level
// falsifier: a lang.interface/0 document is decodable but never admissible
// for a call, so CheckSummary must refuse it rather than silently answering
// origin questions from a frozen legacy shape it was never validated
// against.
func TestCheckSummaryRefusesV0Document(t *testing.T) {
	if _, err := originvalidate.CheckSummary([]byte(pinnedV0DocumentForCheckSummary), []byte(`{}`)); err == nil {
		t.Fatal("expected CheckSummary to refuse a lang.interface/0 document")
	} else if code := errorCode(err); code != "origin.summary_not_admissible" {
		t.Fatalf("expected origin.summary_not_admissible, got %q (%v)", code, err)
	}
}

const pinnedV0DocumentForCheckSummary = `{"schema":"lang.interface/0","module_id":"m1","core_digest":"` + validClosureDigestPlaceholder + `","functions":[{"id":"f1","name":"identity","parameter":{"id":"p1","name":"buffer","type":"Buffer"},"return_type":"Buffer","abilities":[]}]}`

// residual: a coordinated frontend-and-summary lie is declared as a named
// expected escape, never solved and never silently absent.
func TestOriginEscapeIsNamed(t *testing.T) {
	if originvalidate.KnownEscape == "" {
		t.Fatal("KnownEscape must be a non-empty named constant")
	}
	escapes := originvalidate.ExpectedEscapes()
	found := false
	for _, escape := range escapes {
		if escape == originvalidate.KnownEscape {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected KnownEscape among ExpectedEscapes, got %+v", escapes)
	}
}

// TestInterfaceSummaryOmitsBodies is 03-06-01's falsifier: the exported
// interface summary contains no linear or match body — not merely an omitted
// field, but a shape (core.FunctionSignature) that structurally has no such
// field to omit.
func TestInterfaceSummaryOmitsBodies(t *testing.T) {
	program := honestProgram(t, "public_view.schway")
	summary, err := originvalidate.BuildInterface(program)
	if err != nil {
		t.Fatal(err)
	}
	if len(summary.Functions) != 1 {
		t.Fatalf("expected one function signature, got %d", len(summary.Functions))
	}
	if summary.Functions[0].Return.Mode != "shared" {
		t.Fatalf("expected the origin fact to survive stripping: %+v", summary.Functions[0])
	}
	encoded, err := json.Marshal(summary)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{`"linear"`, `"match"`, `"operations"`, `"blocks"`} {
		if bytes.Contains(encoded, []byte(forbidden)) {
			t.Fatalf("interface summary leaked a body field %s: %s", forbidden, encoded)
		}
	}
}

// validClosureDigestPlaceholder is a syntactically valid sha256:+64-hex
// shape used only to satisfy DecodeInterface's shape check in tests that
// predate 07-01 Task 3 (which computes ClosureDigest for real) — it is not
// a real content digest of anything.
const validClosureDigestPlaceholder = "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func errorCode(err error) string {
	var typed *originvalidate.Error
	if errors.As(err, &typed) {
		return typed.Code
	}
	return ""
}

// phase4Program checks a testdata/phase4 fixture (rather than phase3's)
// through session.Check and returns the resulting core.Program, failing the
// test on any diagnostic.
func phase4Program(t testing.TB, fixture string) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", fixture))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	return checked.Program
}

// TestForeignBorrowDerivedReturnRecognised is D-04-28's positive falsifier:
// a foreign call declared to borrow its argument produces a return
// RecomputeOriginPerReturn recognises as Derived (access "shared"), and a
// matching declared PublicOrigin publishes cleanly through ValidatePublished
// -- proving the widening does not merely refuse the omitted case, it also
// correctly validates the declared one. check.go's checkFallibleLinear does
// not (this phase) wire a source-level `borrow(path)` return-type
// annotation onto a foreign-tracer function, so the declared side of this
// fixture is attached directly onto the checked core.Function -- exercising
// exactly the same originvalidate entry points a real declaration would.
func TestForeignBorrowDerivedReturnRecognised(t *testing.T) {
	program := phase4Program(t, "foreign_acquire_one.schway")
	if program.Functions[0].ForeignContract == nil {
		t.Fatal("expected foreign_acquire_one.schway to carry a ForeignContract")
	}
	program.Functions[0].ForeignContract.Alias = "borrow"

	origins := originvalidate.RecomputeOriginPerReturn(program.Functions[0], nil)
	found := false
	for _, origin := range origins {
		if origin.Derived {
			found = true
			if origin.Access != "shared" || len(origin.Paths) != 1 || origin.Paths[0] != "request" {
				t.Fatalf("expected a shared derivation from %q, got %+v", "request", origin)
			}
		}
	}
	if !found {
		t.Fatalf("expected at least one borrow-derived return, got %+v", origins)
	}

	program.Functions[0].PublicOrigin = &core.PublicOrigin{Paths: []string{"request"}, Access: "shared"}
	if problems := originvalidate.ValidatePublished(program); len(problems) != 0 {
		t.Fatalf("expected a correctly declared foreign-borrow origin to publish cleanly, got %+v", problems)
	}
}

// TestForeignOriginOmittedRejected is D-04-28's negative falsifier: the same
// shape with NO declared PublicOrigin is refused with
// core.foreign_origin_omitted, naming the offending function and the
// argument the origin derives from.
func TestForeignOriginOmittedRejected(t *testing.T) {
	program := phase4Program(t, "foreign_origin_omitted.schway")
	problems := originvalidate.ValidatePublished(program)
	if len(problems) != 1 || problems[0].Code != "core.foreign_origin_omitted" {
		t.Fatalf("expected exactly one core.foreign_origin_omitted problem, got %+v", problems)
	}
	if !strings.Contains(problems[0].Detail, "request") {
		t.Fatalf("expected the detail to name the argument the origin derives from: %+v", problems[0])
	}
}

// TestOriginWalksEveryTerminator is D-04-29's falsifier for originvalidate:
// before this widening, RecomputeOriginPerReturn's backward-walk collection
// loop recognised only core.OpReturn, so a function with a fail-only or
// defect-only path contributed nothing to the per-terminator origin
// picture. defect_terminal.schway's "Halt" arm exits ONLY through core.OpDefect
// (no OpReturn in that arm at all) and foreign_acquire_one.schway's err block
// exits ONLY through core.OpFail -- both are reachable inputs the old
// single-terminator condition would have missed entirely.
func TestOriginWalksEveryTerminator(t *testing.T) {
	defectProgram := phase4Program(t, "defect_terminal.schway")
	defectOrigins := originvalidate.RecomputeOriginPerReturn(defectProgram.Functions[0], nil)
	if len(defectOrigins) != 2 {
		t.Fatalf("expected one origin entry per terminator (return + defect), got %d: %+v", len(defectOrigins), defectOrigins)
	}

	foreignProgram := phase4Program(t, "foreign_acquire_one.schway")
	foreignOrigins := originvalidate.RecomputeOriginPerReturn(foreignProgram.Functions[0], nil)
	if len(foreignOrigins) != 2 {
		t.Fatalf("expected one origin entry per terminator (return + fail), got %d: %+v", len(foreignOrigins), foreignOrigins)
	}
}

// TestTerminatorSetReadFromRegistry asserts originvalidate.RecognizesTerminator
// agrees with core.TerminatorKinds() exactly -- every registered terminator
// is recognised, and a non-terminator kind (core.OpCopy) is not -- proving
// the package reads the registry rather than restating a private copy of it.
func TestTerminatorSetReadFromRegistry(t *testing.T) {
	for _, terminator := range core.TerminatorKinds() {
		if !originvalidate.RecognizesTerminator(terminator) {
			t.Fatalf("expected originvalidate to recognise registered terminator %q", terminator)
		}
	}
	if originvalidate.RecognizesTerminator(core.OpCopy) {
		t.Fatalf("expected originvalidate to NOT recognise core.OpCopy as a terminator")
	}
}

// TestTerminatorWalkMutationKilled is D-09's automated mutation-kill
// falsifier for D-04-29: narrowing the recognised terminator set (deleting
// core.OpFail, the exact mutation the throwaway-detached-worktree
// demonstration performs on the source) must make originvalidate lose the
// fail-only fixture's origin fact -- proving the widening actually bites,
// not merely that a differential stays green.
func TestTerminatorWalkMutationKilled(t *testing.T) {
	program := phase4Program(t, "foreign_acquire_one.schway")
	full := originvalidate.RecomputeOriginPerReturn(program.Functions[0], nil)

	original := originvalidate.TerminatorKindsOverride
	originvalidate.TerminatorKindsOverride = func() []core.OperationKind {
		return []core.OperationKind{core.OpReturn, core.OpDefect} // OpFail deleted
	}
	defer func() { originvalidate.TerminatorKindsOverride = original }()
	mutated := originvalidate.RecomputeOriginPerReturn(program.Functions[0], nil)

	if len(mutated) >= len(full) {
		t.Fatalf("mutation (deleting OpFail) had no observable effect: full=%d mutated=%d", len(full), len(mutated))
	}
}

// TestInterfaceV1FieldInvariantsAcrossCorpus is 07-01 Task 1's acceptance
// criterion: over the entire testdata/phase1..phase4 corpus, BuildInterface
// must produce, for every function, a Mode value in {owned, shared,
// exclusive} for every parameter and every return, and Return.Paths empty
// EXACTLY when Return.Mode == "owned" (D-07-09). Only fixtures that check
// cleanly (zero diagnostics) are exercised -- a rejected/malformed fixture
// never reaches BuildInterface in the real pipeline either.
func TestInterfaceV1FieldInvariantsAcrossCorpus(t *testing.T) {
	checkedAny := false
	for _, phase := range []string{"phase1", "phase2", "phase3", "phase4"} {
		dir := testsupport.ProjectPath("testdata", phase)
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".schway") {
				continue
			}
			source, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				t.Fatalf("read %s/%s: %v", phase, entry.Name(), err)
			}
			checked := session.Check(source)
			if len(checked.Diagnostics) > 0 {
				continue // rejected fixture: never reaches BuildInterface for real
			}
			summary, err := originvalidate.BuildInterface(checked.Program)
			if err != nil {
				t.Fatalf("%s/%s: BuildInterface: %v", phase, entry.Name(), err)
			}
			// 07-01 Task 3 (D-07-37): re-running BuildInterface over the same
			// checked program must produce byte-identical ClosureDigests.
			resummary, err := originvalidate.BuildInterface(checked.Program)
			if err != nil {
				t.Fatalf("%s/%s: BuildInterface (rerun): %v", phase, entry.Name(), err)
			}
			for _, function := range summary.Functions {
				if !strings.HasPrefix(function.ClosureDigest, "sha256:") || len(function.ClosureDigest) != len("sha256:")+64 {
					t.Fatalf("%s/%s: function %s ClosureDigest %q is not sha256:+64hex", phase, entry.Name(), function.ID, function.ClosureDigest)
				}
				for _, r := range function.ClosureDigest[len("sha256:"):] {
					if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
						t.Fatalf("%s/%s: function %s ClosureDigest %q has non-lowercase-hex byte", phase, entry.Name(), function.ID, function.ClosureDigest)
					}
				}
			}
			for index, function := range resummary.Functions {
				if function.ClosureDigest != summary.Functions[index].ClosureDigest {
					t.Fatalf("%s/%s: function %s ClosureDigest not identical across two BuildInterface runs: %q != %q", phase, entry.Name(), function.ID, summary.Functions[index].ClosureDigest, function.ClosureDigest)
				}
			}
			for _, function := range summary.Functions {
				checkedAny = true
				for _, parameter := range function.Parameters {
					if parameter.Mode != "owned" && parameter.Mode != "shared" && parameter.Mode != "exclusive" {
						t.Fatalf("%s/%s: function %s parameter %s: Mode %q outside {owned,shared,exclusive}", phase, entry.Name(), function.ID, parameter.Name, parameter.Mode)
					}
				}
				if function.Return.Mode != "owned" && function.Return.Mode != "shared" && function.Return.Mode != "exclusive" {
					t.Fatalf("%s/%s: function %s Return.Mode %q outside {owned,shared,exclusive}", phase, entry.Name(), function.ID, function.Return.Mode)
				}
				pathsEmpty := len(function.Return.Paths) == 0
				if (function.Return.Mode == "owned") != pathsEmpty {
					t.Fatalf("%s/%s: function %s Return.Paths=%v must be empty exactly when Mode==owned (Mode=%q)", phase, entry.Name(), function.ID, function.Return.Paths, function.Return.Mode)
				}
			}
		}
	}
	if !checkedAny {
		t.Fatal("expected at least one function across testdata/phase1..phase4 to be checked")
	}
}

// TestPublishProblemsForMatchesValidatePublishedAcrossCorpus is 07-02 Task
// 1's D-07-32 falsifier: ValidatePublished must return byte-identical
// Problem values before and after the refactor. Since ValidatePublished is
// now a thin loop over PublishProblemsFor, this asserts the equivalence
// directly rather than duplicating the old inlined logic in the test: for
// every checked fixture across testdata/phase1..phase4 and
// testdata/phase07, ValidatePublished(program) must equal the first
// non-empty PublishProblemsFor(function) result over program.Functions in
// order -- exactly the whole-program first-problem contract the doc comment
// claims is preserved.
func TestPublishProblemsForMatchesValidatePublishedAcrossCorpus(t *testing.T) {
	checkedAny := false
	for _, phase := range []string{"phase1", "phase2", "phase3", "phase4", "phase07"} {
		dir := testsupport.ProjectPath("testdata", phase)
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".schway") {
				continue
			}
			source, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				t.Fatalf("read %s/%s: %v", phase, entry.Name(), err)
			}
			checked := session.Check(source)
			if len(checked.Diagnostics) > 0 {
				continue
			}
			checkedAny = true
			calleeContracts := originvalidate.BuildCalleeOriginFacts(checked.Program)
			var expected []originvalidate.Problem
			for _, function := range checked.Program.Functions {
				if problems := originvalidate.PublishProblemsFor(function, calleeContracts); len(problems) > 0 {
					expected = problems
					break
				}
			}
			actual := originvalidate.ValidatePublished(checked.Program)
			if !problemsEqual(expected, actual) {
				t.Fatalf("%s/%s: ValidatePublished()=%+v does not match first non-empty PublishProblemsFor result %+v", phase, entry.Name(), actual, expected)
			}
		}
	}
	if !checkedAny {
		t.Fatal("expected at least one function across the corpus to be checked")
	}
}

func problemsEqual(a, b []originvalidate.Problem) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestCallableIsPublicationSafetyNotExportMembership is 07-02 Task 1's
// D-07-31 falsifier: Callable must be false for a function that checks
// clean but fails publication (the extracted clean_but_unpublishable.schway
// negative control, which trips core.origin_omitted -- not an unexported-
// callee shape, which would be the wrong negative control for this
// predicate), and true for every clean, publishable function in
// testdata/phase1..phase4.
func TestCallableIsPublicationSafetyNotExportMembership(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase07", "clean_but_unpublishable.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("expected clean_but_unpublishable.schway to check cleanly, got %+v", checked.Diagnostics)
	}
	problems := originvalidate.ValidatePublished(checked.Program)
	if len(problems) != 1 || problems[0].Code != "core.origin_omitted" {
		t.Fatalf("expected exactly core.origin_omitted, got %+v", problems)
	}
	summary, err := originvalidate.BuildInterface(checked.Program)
	if err != nil {
		t.Fatalf("BuildInterface: %v", err)
	}
	if len(summary.Functions) != 1 {
		t.Fatalf("expected exactly one function, got %d", len(summary.Functions))
	}
	if summary.Functions[0].Callable {
		t.Fatalf("expected Callable == false for the publication-unsafe fixture, got true")
	}

	checkedAny := false
	for _, phase := range []string{"phase1", "phase2", "phase3", "phase4"} {
		dir := testsupport.ProjectPath("testdata", phase)
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".schway") {
				continue
			}
			fixtureSource, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				t.Fatalf("read %s/%s: %v", phase, entry.Name(), err)
			}
			fixtureChecked := session.Check(fixtureSource)
			if len(fixtureChecked.Diagnostics) > 0 {
				continue
			}
			if len(originvalidate.ValidatePublished(fixtureChecked.Program)) > 0 {
				// This fixture is checked-clean but publication-unsafe by
				// design (e.g. public_view_omitted.schway) -- not part of
				// this assertion's "clean, publishable" set.
				continue
			}
			fixtureSummary, err := originvalidate.BuildInterface(fixtureChecked.Program)
			if err != nil {
				t.Fatalf("%s/%s: BuildInterface: %v", phase, entry.Name(), err)
			}
			for _, function := range fixtureSummary.Functions {
				checkedAny = true
				if !function.Callable {
					t.Fatalf("%s/%s: function %s expected Callable == true for a clean, publishable function, got false", phase, entry.Name(), function.ID)
				}
			}
		}
	}
	if !checkedAny {
		t.Fatal("expected at least one clean, publishable function across testdata/phase1..phase4")
	}
}

// TestBuildInterfaceNeverConsultsExportList is 07-02 Task 1's structural
// falsifier for D-07-31(c): the Callable derivation must never read an
// export list. It re-scans originvalidate.go's own source (the same
// technique TestOriginValidatorImportsStayIndependent uses for imports) and
// fails if the identifier "Exports" ever appears in this package.
func TestBuildInterfaceNeverConsultsExportList(t *testing.T) {
	path := testsupport.ProjectPath("internal", "compiler", "originvalidate", "originvalidate.go")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(source, []byte("Exports")) {
		t.Fatalf("originvalidate.go must never reference an export list; found the identifier \"Exports\"")
	}
}

// TestStage0SummaryMutationMatrix (originvalidate half) is 07-02 Task 3's
// D-07-24 gate for the two faults that flip THIS package's own
// fault-injection seams (typeFactExactIDMatchOverride,
// forceCallableAlwaysTrue): fault 1 (producer-only type-fact lookup) and
// fault 4 (producer-side Callable-always-true). Faults 2, 3, and 5 live in
// corevalidate_mutation_matrix_test.go (package corevalidate_test), which
// flips corevalidate's own seams instead -- neither package's unexported
// seam is reachable from the other's test package (Go's export_test.go
// pattern only reaches a package's OWN external test package), so the five
// faults are split across both packages' own TestStage0SummaryMutationMatrix
// by which seam each one needs, per D-07-42's unexported-only constraint.
func TestStage0SummaryMutationMatrix(t *testing.T) {
	t.Run("fault1_producer_only_type_fact_lookup", func(t *testing.T) {
		// D-07-23: forcing BuildInterface's exact-ID type-fact lookup to
		// never match drives it into the silent empty-abilities fallback.
		// The independent corevalidate peer, whose own lookup this seam
		// never touches, must still report the real abilities -- a
		// SPECIFIC abilities-field divergence, not a generic error.
		program := honestProgram(t, "public_view.schway")
		function := program.Functions[0]

		restore := originvalidate.SetTypeFactExactIDMatchOverrideForTest(func(factID, wantID string) bool { return false })
		defer restore()

		producerSummary, err := originvalidate.BuildInterface(program)
		if err != nil {
			t.Fatalf("BuildInterface: %v", err)
		}
		if len(producerSummary.Functions[0].Abilities) != 0 {
			t.Fatalf("expected the forced lookup failure to drive the producer into the empty-abilities fallback, got %+v", producerSummary.Functions[0].Abilities)
		}

		result := corevalidate.Validate(program)
		if !result.Valid {
			t.Fatalf("expected public_view.schway to corevalidate-validate, got problems: %+v", result.Problems)
		}
		peerSignature, ok := result.PeerSignatures()[function.ID]
		if !ok {
			t.Fatalf("function %s missing from peer signatures", function.ID)
		}
		if len(peerSignature.Abilities) == 0 {
			t.Fatal("expected the peer's independent lookup to be unaffected by the producer-only fault and still report non-empty abilities")
		}
		if len(peerSignature.Abilities) == len(producerSummary.Functions[0].Abilities) {
			t.Fatalf("expected a SPECIFIC abilities-field divergence (peer=%v, producer=%v), got equal lengths", peerSignature.Abilities, producerSummary.Functions[0].Abilities)
		}
	})

	t.Run("fault4_callable_producer", func(t *testing.T) {
		// Forcing ONLY the producer's Callable derivation to true must
		// still leave the peer refusing on clean_but_unpublishable.schway --
		// a real divergence, not a bilateral false agreement (see fault 3
		// in corevalidate_mutation_matrix_test.go).
		source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase07", "clean_but_unpublishable.schway"))
		if err != nil {
			t.Fatal(err)
		}
		checked := session.Check(source)
		if len(checked.Diagnostics) != 0 {
			t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
		}
		program := checked.Program
		function := program.Functions[0]

		restore := originvalidate.SetCallableForceOverrideForTest(true)
		defer restore()

		producerSummary, err := originvalidate.BuildInterface(program)
		if err != nil {
			t.Fatalf("BuildInterface: %v", err)
		}
		if !producerSummary.Functions[0].Callable {
			t.Fatal("expected the forced producer override to report Callable == true")
		}
		result := corevalidate.Validate(program)
		if !result.Valid {
			t.Fatalf("expected clean_but_unpublishable.schway to corevalidate-validate, got problems: %+v", result.Problems)
		}
		peerSignature, ok := result.PeerSignatures()[function.ID]
		if !ok {
			t.Fatalf("function %s missing from peer signatures", function.ID)
		}
		if peerSignature.Callable {
			t.Fatal("expected the UNFORCED peer to still independently refuse (Callable == false) -- diverging from the forced producer")
		}
	})
}

// TestOpCallOriginWalkGateIsLoadBearing is D-10-08's mutation-kill gate for
// Task 2's core.OpCall case: engaging the nil-default seam
// (disableOpCallOriginConsultForTest, exposed via
// SetDisableOpCallOriginConsultForTest) must regress twin_a_accept.schway to
// its pre-fix core.origin_omitted refusal -- proving the case decides
// something, not merely that it is syntactically present (D-10-08). Runs
// through the FULL session.CheckCommandFile pipeline (check ->
// corevalidate -> originvalidate's fixed precedence, session.go:719),
// never a direct package-level call: that precedence chain is exactly what
// masked D-09-51 for two phases.
func TestOpCallOriginWalkGateIsLoadBearing(t *testing.T) {
	fixture := testsupport.ProjectPath("testdata", "phase08", "twin_a_accept.schway")

	// Beat 1: the fixture now admits cleanly through the full CLI.
	result, err := session.CheckCommandFile(fixture)
	if err != nil {
		t.Fatalf("CheckCommandFile: %v", err)
	}
	if result.Status != protocol.StatusPass || len(result.Diagnostics) != 0 {
		t.Fatalf("expected twin_a_accept.schway to admit cleanly, got status=%q diagnostics=%+v", result.Status, result.Diagnostics)
	}

	// Beat 2: engage the seam, restored via defer.
	restore := originvalidate.SetDisableOpCallOriginConsultForTest(true)
	defer restore()

	// Beats 3-4: the fixture regresses to its pre-fix refusal, and the
	// EXACT diagnostic code is asserted, never merely that some refusal
	// occurred.
	regressed, err := session.CheckCommandFile(fixture)
	if err != nil {
		t.Fatalf("CheckCommandFile (seam engaged): %v", err)
	}
	if regressed.Status != protocol.StatusInvalid || len(regressed.Diagnostics) != 1 || regressed.Diagnostics[0].Code != "core.origin_omitted" {
		t.Fatalf("expected the seam to regress twin_a_accept.schway to exactly core.origin_omitted, got status=%q diagnostics=%+v", regressed.Status, regressed.Diagnostics)
	}
}

// TestOpCallOriginWalkPropagatesGenuineBorrowingCallee is Task 2's
// positive-direction falsifier: the fix NARROWS the walk, it does not
// disable it. A callee whose declared return contract genuinely IS a
// borrow of its own parameter still propagates that borrow-derived origin
// across the OpCall hop, all the way back to a caller that directly
// forwards its own parameter into the call -- reusing
// straightLineCallProgramForTest's (originvalidate_closure_chain_test.go)
// caller-calls-callee shape with calleeHasDeclaredOrigin=true.
func TestOpCallOriginWalkPropagatesGenuineBorrowingCallee(t *testing.T) {
	program := straightLineCallProgramForTest(true)
	caller := findByName(t, program, "caller")
	calleeContracts := originvalidate.BuildCalleeOriginFacts(program)

	paths, access, ok := originvalidate.RecomputeOrigin(caller, calleeContracts)
	if !ok {
		t.Fatal("expected caller's return to be recognised as borrow-derived once its callee genuinely declares a borrow-of-parameter return")
	}
	if access != "shared" {
		t.Fatalf("expected access shared (matching the callee's declared access), got %q", access)
	}
	if len(paths) != 1 || paths[0] != caller.Parameter.Name {
		t.Fatalf("expected origin path [%s] (caller's own parameter), got %+v", caller.Parameter.Name, paths)
	}
}

// bodyBlindnessViolations scans every function declaration in file named by
// targetFuncs and reports, for each parameter OTHER than the function under
// analysis's own leading `function core.Function` parameter, whether its
// rendered type text names core.Function or core.Program -- either of
// which would make a callee's BODY reachable from inside the walk (T-10-05).
// The leading `function core.Function` parameter is legitimately
// core.Function-shaped: it is the function the walk analyzes, never a
// callee. Shared between the real-file scan and the negative control below
// so both exercise the identical predicate.
func bodyBlindnessViolations(t *testing.T, fset *token.FileSet, file *ast.File, targetFuncs map[string]bool) []string {
	t.Helper()
	var violations []string
	checked := 0
	ast.Inspect(file, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || !targetFuncs[fn.Name.Name] || fn.Type.Params == nil {
			return true
		}
		checked++
		for _, field := range fn.Type.Params.List {
			var buf bytes.Buffer
			if err := printer.Fprint(&buf, fset, field.Type); err != nil {
				t.Fatalf("%s: render parameter type: %v", fn.Name.Name, err)
			}
			text := buf.String()
			if len(field.Names) == 1 && field.Names[0].Name == "function" && text == "core.Function" {
				// The function under analysis itself -- legitimate, not a
				// callee's body becoming reachable.
				continue
			}
			if strings.Contains(text, "core.Function") || strings.Contains(text, "core.Program") {
				violations = append(violations, fn.Name.Name+": parameter type "+text)
			}
		}
		return true
	})
	if checked == 0 {
		t.Fatal("bodyBlindnessViolations: no target function declarations found to scan")
	}
	return violations
}

// TestWalkReturnOriginSignatureStaysBodyBlind is D-10-04's mechanical,
// build-or-test-level falsifier for T-10-05: walkReturnOrigin,
// RecomputeOriginPerReturn, and RecomputeOrigin gained an added
// callee-contract parameter in this plan's Task 1, and this test asserts
// that parameter -- and every other parameter besides the function each of
// these already analyzes -- never names core.Function or core.Program.
// Nothing about the walk's OWN production behavior needs a callee's body;
// this test makes that a property of the TYPE SIGNATURE a future editor
// cannot silently violate, not a claim resting on a doc comment.
func TestWalkReturnOriginSignatureStaysBodyBlind(t *testing.T) {
	fset := token.NewFileSet()
	path := testsupport.ProjectPath("internal", "compiler", "originvalidate", "originvalidate.go")
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	targets := map[string]bool{"walkReturnOrigin": true, "RecomputeOriginPerReturn": true, "RecomputeOrigin": true}
	if violations := bodyBlindnessViolations(t, fset, file, targets); len(violations) != 0 {
		t.Fatalf("body-blindness violated: %+v", violations)
	}

	// Negative control (D-10-04's own bar: a mechanical assertion must be
	// observed to go red, never merely assumed to). A synthetic
	// walkReturnOrigin signature that leaks a core.Program parameter must
	// be flagged by the identical scan -- proving this is a real scanner,
	// not a vacuous one that always reports clean.
	syntheticSource := `package originvalidate

import "github.com/szTheory/schway/internal/compiler/core"

func walkReturnOrigin(function core.Function, calleeContracts map[string]calleeOriginFact, leaked core.Program) ReturnOrigin {
	return ReturnOrigin{}
}
`
	syntheticFset := token.NewFileSet()
	syntheticFile, err := parser.ParseFile(syntheticFset, "synthetic_leak.go", syntheticSource, 0)
	if err != nil {
		t.Fatal(err)
	}
	violations := bodyBlindnessViolations(t, syntheticFset, syntheticFile, map[string]bool{"walkReturnOrigin": true})
	if len(violations) == 0 {
		t.Fatal("expected the negative control's leaked core.Program parameter to be flagged -- the scan can never go red")
	}
}
