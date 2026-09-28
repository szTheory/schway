package corevalidate_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/codename-lang/lang/internal/compiler/ability"
	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// 07-11 Task 2 Test 3 (independence): this test file imports
// compiler/check as an ORDINARY dependency, purely to drive
// TestCallConsumePeerIndependentOfCheck's bilateral independent-disable
// proof (mirroring check_test.go's own
// TestCorevalidatePeerIndependentOfCheckCycleRefusal in reverse). This is
// NOT the independence property under test: that property is
// corevalidate.go's OWN production source never importing compiler/check
// (TestArbitraryMaskCannotEnterCoreValidation above, which reads
// corevalidate.go's file text directly, unaffected by what this _test.go
// file imports).

// mustDeriveTypeFactForTest builds a structurally-valid core.TypeFact for a
// primitive constructor (Byte or Buffer), deriving its real ability set
// through ability.Derive rather than hand-guessing it -- Buffer denies
// AbilityCopy (with a witness), Byte grants it, and a hardcoded ability
// list here would silently drift from the real derivation and trip
// core.ability_mismatch on an unrelated axis this file never intended to
// test. id is the caller-supplied core.TypeFact.ID.
func mustDeriveTypeFactForTest(t *testing.T, id, constructor string) core.TypeFact {
	t.Helper()
	shape := core.TypeRef{Constructor: constructor, Arguments: []core.TypeRef{}}
	derived, err := ability.Derive(shape)
	if err != nil {
		t.Fatalf("ability.Derive(%q): %v", constructor, err)
	}
	return core.TypeFact{ID: id, Shape: shape, Abilities: derived.Granted, NegativeWitnesses: derived.NegativeWitnesses}
}

func TestOwnershipMutationMatrix(t *testing.T) {
	valid := ownedProgram()
	if result := corevalidate.Validate(valid); !result.Valid {
		t.Fatalf("valid owned core rejected: %+v", result)
	}

	tests := []struct {
		name string
		code string
		edit func(*core.Program)
	}{
		{"forged copy operation", "core.ability.copy_denied", func(program *core.Program) {
			program.Functions[0].Linear.Operations[0].Kind = core.OpCopy
		}},
		{"forged positive ability", "core.ability_mismatch", func(program *core.Program) {
			program.Functions[0].Linear.Types[0].Abilities = append([]core.Ability{core.AbilityCopy}, program.Functions[0].Linear.Types[0].Abilities...)
		}},
		{"duplicate operation id", "core.duplicate_operation_id", func(program *core.Program) {
			program.Functions[0].Linear.Operations[1].ID = program.Functions[0].Linear.Operations[0].ID
		}},
		{"unknown place", "core.unknown_place", func(program *core.Program) {
			program.Functions[0].Linear.Operations[0].SourceID = "s1:test:fn:relay:place:absent"
		}},
		{"unknown type", "core.unknown_type", func(program *core.Program) {
			program.Functions[0].Linear.Places[0].TypeID = "s1:test:fn:relay:type:absent"
		}},
		{"unknown loan", "core.unknown_loan", func(program *core.Program) {
			borrow := borrowedProgram()
			*program = borrow
			program.Functions[0].Linear.Operations[0].LoanID = ""
		}},
		{"omitted move", "core.operation_order", func(program *core.Program) {
			operations := program.Functions[0].Linear.Operations
			program.Functions[0].Linear.Operations = append([]core.LinearOperation(nil), operations[1:]...)
		}},
		{"illegal loan transition", "core.move_while_borrowed", func(program *core.Program) {
			borrow := borrowedProgram()
			*program = borrow
			operations := program.Functions[0].Linear.Operations
			firstID, firstPoint := operations[1].ID, operations[1].PointID
			secondID, secondPoint := operations[2].ID, operations[2].PointID
			operations[1], operations[2] = operations[2], operations[1]
			operations[1].ID, operations[1].PointID = firstID, firstPoint
			operations[2].ID, operations[2].PointID = secondID, secondPoint
		}},
		{"inconsistent final claim", "core.final_claim_mismatch", func(program *core.Program) {
			program.Functions[0].Linear.Operations[1].SourceID = program.Functions[0].Parameter.ID
		}},
		{"self move target", "core.invalid_target", func(program *core.Program) {
			program.Functions[0].Linear.Operations[0].TargetID = program.Functions[0].Linear.Operations[0].SourceID
		}},
		{"target overwrite", "core.invalid_target", func(program *core.Program) {
			borrow := borrowedProgram()
			*program = borrow
			program.Functions[0].Linear.Operations[1].TargetID = program.Functions[0].Linear.Operations[0].TargetID
		}},
		{"skipped target ordinal", "core.invalid_target", func(program *core.Program) {
			borrow := borrowedProgram()
			*program = borrow
			program.Functions[0].Linear.Operations[0].TargetID = program.Functions[0].Linear.Places[2].ID
		}},
		{"reused borrow target", "core.invalid_target", func(program *core.Program) {
			borrow := borrowedProgram()
			*program = borrow
			program.Functions[0].Linear.Operations[1].Kind = core.OpBorrowShared
			program.Functions[0].Linear.Operations[1].LoanID = program.Functions[0].ID + ":loan:1"
			program.Functions[0].Linear.Operations[1].TargetID = program.Functions[0].Linear.Operations[0].TargetID
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutated := cloneProgram(t, valid)
			test.edit(&mutated)
			result := corevalidate.Validate(mutated)
			if result.Valid || len(result.Problems) == 0 || result.Problems[0].Code != test.code {
				t.Fatalf("mutation result=%+v, want first code %q", result, test.code)
			}
		})
	}
}

func phase23LocalOwnerProgram(t *testing.T) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "phase23", "file_byte.lang"))
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

func withoutPhase23Releases(program *core.Program) {
	function := &program.Functions[0]
	operations := function.Linear.Operations[:0]
	for _, operation := range function.Linear.Operations {
		if operation.Kind != core.OpRelease {
			operations = append(operations, operation)
		}
	}
	function.Linear.Operations = operations
}

func TestPhase23AcquisitionSeedsCoreOwnerObligation(t *testing.T) {
	program := phase23LocalOwnerProgram(t)
	if result := corevalidate.Validate(program); !result.Valid {
		t.Fatalf("valid acquire/use/release core rejected: %+v", result.Problems)
	}

	withoutPhase23Releases(&program)
	if result := corevalidate.Validate(program); result.Valid {
		t.Fatal("corevalidate accepted successful acquisition after every release was removed")
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
			program.Functions[0].Linear.Operations[index].Foreign.Symbol = "lang_file_byte_acquire"
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

func TestPhase23ResourceMutationCorevalidate(t *testing.T) {
	for _, mutation := range phase23OwnerMutations() {
		t.Run(mutation.name, func(t *testing.T) {
			program := phase23LocalOwnerProgram(t)
			if !mutation.edit(&program) {
				t.Fatal("mutation target was not reached")
			}
			if result := corevalidate.Validate(program); result.Valid {
				t.Fatal("corevalidate accepted the reached ownership mutation")
			}
		})
	}
}

func TestArbitraryMaskCannotEnterCoreValidation(t *testing.T) {
	encoded, err := json.Marshal(ownedProgram())
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"mask", "arbitrary", "bits"} {
		if strings.Contains(strings.ToLower(string(encoded)), forbidden) {
			t.Fatalf("serialized core contains arbitrary ability representation %q: %s", forbidden, encoded)
		}
	}
	source, err := os.ReadFile(testsupport.ProjectPath("internal", "compiler", "corevalidate", "corevalidate.go"))
	if err != nil {
		t.Fatal(err)
	}
	// Phase 09 (D-09-05): corevalidate_peer_liveness.go is a sibling
	// production file in this same package (D-09-04) -- scanned here too,
	// not assumed automatically covered, so a forbidden import landing
	// there is caught by every guard that would have caught it in
	// corevalidate.go.
	peerLivenessSource, err := os.ReadFile(testsupport.ProjectPath("internal", "compiler", "corevalidate", "corevalidate_peer_liveness.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"compiler/ast", "compiler/check", "compiler/interp", "compiler/cgen", "compiler/session", "compiler/ability"} {
		if strings.Contains(string(source), forbidden) {
			t.Fatalf("validator imports forbidden producer/engine package %q", forbidden)
		}
		if strings.Contains(string(peerLivenessSource), forbidden) {
			t.Fatalf("corevalidate_peer_liveness.go imports forbidden producer/engine package %q", forbidden)
		}
	}
}

func TestPhase17CorePeerDirectionalAbilities(t *testing.T) {
	program := phase17CoreProgram(t)
	result := corevalidate.Validate(program)
	if !result.Valid {
		t.Fatalf("canonical two-type core rejected: %+v", result.Problems)
	}
	for _, function := range program.Functions {
		signature, ok := result.PeerSignatures()[function.ID]
		if !ok || len(signature.Parameters) != 1 {
			t.Fatalf("%s peer signature missing parameter contract: %+v", function.Name, signature)
		}
		if signature.Parameters[0].Type != function.Parameter.Type || signature.Return.Type != function.ReturnType {
			t.Fatalf("%s peer contract lost directional types: %+v", function.Name, signature)
		}
		if function.Linear != nil && (!typeFactGrants(function.Linear.Types[0], core.AbilityDrop) || !signature.Return.Fresh) {
			t.Fatalf("%s did not retain parameter ability and independently derive return Fresh: %+v", function.Name, signature)
		}
	}
}

func typeFactGrants(fact core.TypeFact, wanted core.Ability) bool {
	for _, ability := range fact.Abilities {
		if ability == wanted {
			return true
		}
	}
	return false
}

func TestPhase17CorePeerDirectionalCallContract(t *testing.T) {
	program := phase17CoreProgram(t)
	if result := corevalidate.Validate(program); !result.Valid {
		t.Fatalf("canonical directional call rejected: %+v", result.Problems)
	}
	for _, function := range program.Functions {
		if function.Name != "main" || function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind != core.OpCall {
				continue
			}
			if operation.TypeID == function.ID+":type:0" || operation.CalleeID == "" {
				t.Fatalf("directional call did not retain distinct target/callee facts: %+v", operation)
			}
			return
		}
	}
	t.Fatal("canonical program has no call for directional peer validation")
}

func TestPhase17CorePeerReturnOnlyMutation(t *testing.T) {
	program := phase17CoreProgram(t)
	baseline := corevalidate.Validate(program)
	if !baseline.Valid {
		t.Fatalf("baseline rejected: %+v", baseline.Problems)
	}
	before := baseline.PeerSignatures()

	restore := corevalidate.SetPhase17ReturnLookupFaultForTest(true)
	t.Cleanup(restore)
	mutated := corevalidate.Validate(program)
	if !mutated.Valid {
		t.Fatalf("return-only fault should diverge the peer signature, not invalidate core: %+v", mutated.Problems)
	}
	after := mutated.PeerSignatures()
	for id, signature := range before {
		faulted, ok := after[id]
		if !ok {
			t.Fatalf("faulted peer signature missing %s", id)
		}
		if !reflect.DeepEqual(signature.Parameters, faulted.Parameters) {
			t.Fatalf("%s parameter contract changed under return-only fault: before=%+v after=%+v", id, signature.Parameters, faulted.Parameters)
		}
		if faulted.Return.Fresh {
			t.Fatalf("%s return-only fault left Fresh true: %+v", id, faulted.Return)
		}
	}
	restore()
	restore()
	restored := corevalidate.Validate(program)
	for id, signature := range restored.PeerSignatures() {
		var hasLinear bool
		for _, function := range program.Functions {
			if function.ID == id {
				hasLinear = function.Linear != nil
			}
		}
		if !hasLinear {
			continue
		}
		if !signature.Return.Fresh {
			t.Fatalf("%s return lookup did not restore", id)
		}
	}
}

func TestPhase17CorePeerIndependenceBoundary(t *testing.T) {
	directory := testsupport.ProjectPath("internal", "compiler", "corevalidate")
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("read corevalidate directory: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, imported := range file.Imports {
			value := strings.Trim(imported.Path.Value, "\"")
			if strings.Contains(value, "compiler/check") || strings.Contains(strings.ToLower(value), "returncontract") {
				t.Fatalf("%s imports forbidden producer/shared return dependency %q", entry.Name(), value)
			}
		}
		var callers []token.Pos
		goast.Inspect(file, func(node goast.Node) bool {
			call, ok := node.(*goast.CallExpr)
			if !ok {
				return true
			}
			if ident, ok := call.Fun.(*goast.Ident); ok && ident.Name == "SetPhase17ReturnLookupFaultForTest" {
				callers = append(callers, call.Pos())
			}
			return true
		})
		if len(callers) != 0 {
			t.Fatalf("%s has production caller(s) of return-only test control at %v", entry.Name(), callers)
		}
	}
}

func phase17CoreProgram(t *testing.T) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase17", "return_type_tracer.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("canonical source failed to check: %+v", checked.Diagnostics)
	}
	return checked.Program
}

func TestValidationOwnsTrustBoundaryCopy(t *testing.T) {
	program := ownedProgram()
	result := corevalidate.Validate(program)
	if !result.Valid {
		t.Fatalf("valid core rejected: %+v", result)
	}
	program.Functions[0].Linear.Types[0].Shape.Constructor = "Forged"
	program.Functions[0].Linear.Operations[0].Kind = core.OpCopy
	if got := result.Program(); !reflect.DeepEqual(got, ownedProgram()) {
		t.Fatalf("validated core aliases caller storage:\ngot=%+v", got)
	}
	first := result.Program()
	first.Functions[0].Linear.Places[0].Name = "mutated"
	if second := result.Program(); second.Functions[0].Linear.Places[0].Name == "mutated" {
		t.Fatal("Program returned mutable validator-owned storage")
	}
}

func TestUnvalidatedCoreCannotExecute(t *testing.T) {
	forged := ownedProgram()
	forged.Functions[0].Linear.Types[0].Abilities = append([]core.Ability{core.AbilityCopy}, forged.Functions[0].Linear.Types[0].Abilities...)
	if _, err := interp.Run(forged, "relay", "01020304"); err == nil || !strings.Contains(err.Error(), "core.ability_mismatch") {
		t.Fatalf("unvalidated forged core reached interpreter: %v", err)
	}
}

func TestCoreValidationWorkSeries(t *testing.T) {
	for _, facts := range []int{101, 1001, 10001} {
		program := scaleProgram(facts)
		result := corevalidate.Validate(program)
		if !result.Valid {
			t.Fatalf("facts=%d rejected: %+v", facts, result)
		}
		want := corevalidate.LinearWorkLimit(facts)
		if result.Checks != want {
			t.Fatalf("facts=%d checks=%d want exact linear formula %d", facts, result.Checks, want)
		}
	}
}

func TestOwnedClaimReorderRejected(t *testing.T) {
	program := ownedProgram()
	before, err := json.Marshal(program)
	if err != nil {
		t.Fatal(err)
	}
	operations := program.Functions[0].Linear.Operations
	operations[0], operations[1] = operations[1], operations[0]
	after, err := json.Marshal(program)
	if err != nil {
		t.Fatal(err)
	}
	if sha256.Sum256(before) == sha256.Sum256(after) {
		t.Fatal("semantic operation/final-claim reorder did not change identity")
	}
	result := corevalidate.Validate(program)
	if result.Valid || len(result.Problems) == 0 || result.Problems[0].Code != "core.operation_order" {
		t.Fatalf("reordered final claim survived: %+v", result)
	}
}

func TestCoordinatedSourceCoreEscapeIsNamed(t *testing.T) {
	falseClaim := ownedProgram()
	function := &falseClaim.Functions[0]
	function.Parameter.Type = "Byte"
	function.ReturnType = "Byte"
	fact := &function.Linear.Types[0]
	fact.Shape = core.TypeRef{Constructor: "Byte", Arguments: []core.TypeRef{}}
	fact.Abilities = []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape}
	fact.NegativeWitnesses = []core.AbilityWitness{}
	function.Linear.Operations[0].Kind = core.OpCopy

	result := corevalidate.Validate(falseClaim)
	if !result.Valid {
		t.Fatalf("internally consistent coordinated false claim was incorrectly reported as detected: %+v", result)
	}
	if corevalidate.KnownEscape != "escape:coordinated-source-core-lie" {
		t.Fatalf("unexpected proof-boundary name %q", corevalidate.KnownEscape)
	}
}

func ownedProgram() core.Program {
	functionID := "s1:test:fn:relay"
	typeID := functionID + ":type:0"
	parameterID := functionID + ":place:0"
	targetID := functionID + ":place:1"
	return core.Program{
		Schema: core.Schema1, Module: "test", ModuleID: "s1:test:module:test",
		Functions: []core.Function{{
			ID: functionID, Name: "relay", EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
			Parameter: core.Parameter{ID: parameterID, Name: "buffer", Type: "Buffer"}, ReturnType: "Buffer",
			Linear: &core.LinearBody{
				ID: functionID + ":linear",
				Types: []core.TypeFact{{
					ID: typeID, Shape: core.TypeRef{Constructor: "Buffer", Arguments: []core.TypeRef{}},
					Abilities:         []core.Ability{core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
					NegativeWitnesses: []core.AbilityWitness{{Ability: core.AbilityCopy, Path: []string{"Buffer"}}},
				}},
				Places: []core.Place{{ID: parameterID, Name: "buffer", TypeID: typeID}, {ID: targetID, Name: "delivered", TypeID: typeID}},
				Operations: []core.LinearOperation{
					{ID: functionID + ":op:0", PointID: functionID + ":point:linear:0", Kind: core.OpMove, SourceID: parameterID, TargetID: targetID, TypeID: typeID},
					{ID: functionID + ":op:1", PointID: functionID + ":point:linear:1", Kind: core.OpReturn, SourceID: targetID, TypeID: typeID},
				},
			},
		}},
	}
}

func borrowedProgram() core.Program {
	program := ownedProgram()
	function := &program.Functions[0]
	linear := function.Linear
	function.Parameter.Type = "Byte"
	function.ReturnType = "Byte"
	linear.Types[0] = core.TypeFact{
		ID: linear.Types[0].ID, Shape: core.TypeRef{Constructor: "Byte", Arguments: []core.TypeRef{}},
		Abilities:         []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
		NegativeWitnesses: []core.AbilityWitness{},
	}
	owner := function.Parameter.ID
	view := function.ID + ":place:1"
	observed := function.ID + ":place:2"
	delivered := function.ID + ":place:3"
	linear.Places = []core.Place{
		{ID: owner, Name: "buffer", TypeID: linear.Types[0].ID},
		{ID: view, Name: "view", TypeID: linear.Types[0].ID},
		{ID: observed, Name: "observed", TypeID: linear.Types[0].ID},
		{ID: delivered, Name: "delivered", TypeID: linear.Types[0].ID},
	}
	linear.Operations = []core.LinearOperation{
		{ID: function.ID + ":op:0", PointID: function.ID + ":point:linear:0", Kind: core.OpBorrowShared, SourceID: owner, TargetID: view, LoanID: function.ID + ":loan:0", TypeID: linear.Types[0].ID},
		{ID: function.ID + ":op:1", PointID: function.ID + ":point:linear:1", Kind: core.OpCopy, SourceID: view, TargetID: observed, TypeID: linear.Types[0].ID},
		{ID: function.ID + ":op:2", PointID: function.ID + ":point:linear:2", Kind: core.OpMove, SourceID: owner, TargetID: delivered, TypeID: linear.Types[0].ID},
		{ID: function.ID + ":op:3", PointID: function.ID + ":point:linear:3", Kind: core.OpReturn, SourceID: delivered, TypeID: linear.Types[0].ID},
	}
	return program
}

// reborrowProgram derives a second place from the loan target (a reborrow when
// derivation is OpBorrowShared, a copy of the loan when it is OpCopy) and then
// observes that derived place after the owner is moved. Loan liveness must be
// transitive for this to be refused: associating loan:0 only with its
// immediate target expires it at op:1 and admits the move at op:2.
func reborrowProgram(derivation core.OperationKind, observeAfterMove bool) core.Program {
	program := borrowedProgram()
	function := &program.Functions[0]
	linear := function.Linear
	owner := function.Parameter.ID
	view := function.ID + ":place:1"
	review := function.ID + ":place:2"
	delivered := function.ID + ":place:3"
	observed := function.ID + ":place:4"
	typeID := linear.Types[0].ID
	linear.Places = []core.Place{
		{ID: owner, Name: "buffer", TypeID: typeID},
		{ID: view, Name: "view", TypeID: typeID},
		{ID: review, Name: "review", TypeID: typeID},
		{ID: delivered, Name: "delivered", TypeID: typeID},
		{ID: observed, Name: "observed", TypeID: typeID},
	}
	derivedLoan := ""
	if derivation == core.OpBorrowShared {
		derivedLoan = function.ID + ":loan:1"
	}
	linear.Operations = []core.LinearOperation{
		{ID: function.ID + ":op:0", PointID: function.ID + ":point:linear:0", Kind: core.OpBorrowShared, SourceID: owner, TargetID: view, LoanID: function.ID + ":loan:0", TypeID: typeID},
		{ID: function.ID + ":op:1", PointID: function.ID + ":point:linear:1", Kind: derivation, SourceID: view, TargetID: review, LoanID: derivedLoan, TypeID: typeID},
		{ID: function.ID + ":op:2", PointID: function.ID + ":point:linear:2", Kind: core.OpMove, SourceID: owner, TargetID: delivered, TypeID: typeID},
	}
	if observeAfterMove {
		linear.Operations = append(linear.Operations,
			core.LinearOperation{ID: function.ID + ":op:3", PointID: function.ID + ":point:linear:3", Kind: core.OpCopy, SourceID: review, TargetID: observed, TypeID: typeID},
			core.LinearOperation{ID: function.ID + ":op:4", PointID: function.ID + ":point:linear:4", Kind: core.OpReturn, SourceID: delivered, TypeID: typeID},
		)
	} else {
		linear.Places = linear.Places[:4]
		linear.Operations = append(linear.Operations,
			core.LinearOperation{ID: function.ID + ":op:3", PointID: function.ID + ":point:linear:3", Kind: core.OpReturn, SourceID: delivered, TypeID: typeID},
		)
	}
	return program
}

func TestTransitiveLoanBlocksMove(t *testing.T) {
	for _, testCase := range []struct {
		name       string
		derivation core.OperationKind
	}{
		{name: "reborrow of a loan", derivation: core.OpBorrowShared},
		{name: "copy of a loan", derivation: core.OpCopy},
	} {
		blocked := corevalidate.Validate(reborrowProgram(testCase.derivation, true))
		if blocked.Valid || len(blocked.Problems) == 0 || blocked.Problems[0].Code != "core.move_while_borrowed" {
			t.Fatalf("%s: owner moved while a transitively derived loan was still observed: %+v", testCase.name, blocked)
		}
		unused := corevalidate.Validate(reborrowProgram(testCase.derivation, false))
		if !unused.Valid {
			t.Fatalf("%s: transitive loan liveness over-blocked a loan with no later use: %+v", testCase.name, unused)
		}
	}
}

func scaleProgram(facts int) core.Program {
	if facts < 1 {
		panic("scale facts must include the final claim")
	}
	functionID := "s1:scale:fn:copy"
	typeID := functionID + ":type:0"
	parameterID := functionID + ":place:0"
	linear := &core.LinearBody{
		ID: functionID + ":linear",
		Types: []core.TypeFact{{
			ID: typeID, Shape: core.TypeRef{Constructor: "Byte", Arguments: []core.TypeRef{}},
			Abilities:         []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
			NegativeWitnesses: []core.AbilityWitness{},
		}},
		Places:     make([]core.Place, 0, facts),
		Operations: make([]core.LinearOperation, 0, facts),
	}
	linear.Places = append(linear.Places, core.Place{ID: parameterID, Name: "source", TypeID: typeID})
	for index := 0; index < facts-1; index++ {
		targetID := fmt.Sprintf("%s:place:%d", functionID, index+1)
		linear.Places = append(linear.Places, core.Place{ID: targetID, Name: fmt.Sprintf("copy%d", index), TypeID: typeID})
		linear.Operations = append(linear.Operations, core.LinearOperation{
			ID: fmt.Sprintf("%s:op:%d", functionID, index), PointID: fmt.Sprintf("%s:point:linear:%d", functionID, index),
			Kind: core.OpCopy, SourceID: parameterID, TargetID: targetID, TypeID: typeID,
		})
	}
	last := facts - 1
	linear.Operations = append(linear.Operations, core.LinearOperation{
		ID: fmt.Sprintf("%s:op:%d", functionID, last), PointID: fmt.Sprintf("%s:point:linear:%d", functionID, last),
		Kind: core.OpReturn, SourceID: parameterID, TypeID: typeID,
	})
	return core.Program{
		Schema: core.Schema1, Module: "scale", ModuleID: "s1:scale:module:scale",
		Functions: []core.Function{{
			ID: functionID, Name: "copy", EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
			Parameter: core.Parameter{ID: parameterID, Name: "source", Type: "Byte"}, ReturnType: "Byte", Linear: linear,
		}},
	}
}

func cloneProgram(t *testing.T, program core.Program) core.Program {
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

func foreignAcquireProgram(t *testing.T) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "foreign_acquire_one.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture failed to check: %+v", checked.Diagnostics)
	}
	return checked.Program
}

// TestForeignRefusalsAreIndependentlyDerived proves corevalidate re-derives
// D-04-16's missing-policy refusal and D-04-02's call-target refusal purely
// from the core artifact, never from check's own AST-level facts: each
// mutation below is fed DIRECTLY to corevalidate.Validate, bypassing
// check.Program entirely, so the refusal cannot be riding on check's own
// gate.
func TestForeignRefusalsAreIndependentlyDerived(t *testing.T) {
	valid := foreignAcquireProgram(t)
	if result := corevalidate.Validate(valid); !result.Valid {
		t.Fatalf("valid foreign-call core rejected: %+v", result)
	}

	t.Run("missing unwind policy", func(t *testing.T) {
		mutated := cloneProgram(t, valid)
		mutated.Functions[0].ForeignContract.Unwind = ""
		result := corevalidate.Validate(mutated)
		if result.Valid || result.Problems[0].Code != "foreign.unwind_policy_undeclared" {
			t.Fatalf("expected foreign.unwind_policy_undeclared, got %+v", result)
		}
	})

	t.Run("missing nonlocal_exit policy", func(t *testing.T) {
		mutated := cloneProgram(t, valid)
		mutated.Functions[0].ForeignContract.NonlocalExit = ""
		result := corevalidate.Validate(mutated)
		if result.Valid || result.Problems[0].Code != "foreign.unwind_policy_undeclared" {
			t.Fatalf("expected foreign.unwind_policy_undeclared, got %+v", result)
		}
	})

	t.Run("call target resolves to a declared Lang function", func(t *testing.T) {
		mutated := cloneProgram(t, valid)
		// Introduce a second, otherwise-inert function whose name collides
		// with the foreign contract's symbol -- the same shape check.go's
		// own admission gate refuses at the AST level, re-derived here
		// purely from core.Function.Name identity.
		decoy := cloneProgram(t, valid).Functions[0]
		decoy.ID = decoy.ID + ":decoy"
		decoy.Name = mutated.Functions[0].ForeignContract.Symbol
		decoy.ForeignContract = nil
		decoy.Linear.ID = decoy.ID + ":linear"
		for index := range decoy.Linear.Places {
			decoy.Linear.Places[index].ID = decoy.ID + fmt.Sprintf(":place:%d", index)
		}
		mutated.Functions = append(mutated.Functions, decoy)
		result := corevalidate.Validate(mutated)
		if result.Valid || result.Problems[0].Code != "core.call_target_not_foreign" {
			t.Fatalf("expected core.call_target_not_foreign, got %+v", result)
		}
	})

	t.Run("missing D-04-12 obligation", func(t *testing.T) {
		mutated := cloneProgram(t, valid)
		mutated.Functions[0].ForeignContract.Capture = ""
		result := corevalidate.Validate(mutated)
		if result.Valid || result.Problems[0].Code != "foreign.obligation_undeclared" {
			t.Fatalf("expected foreign.obligation_undeclared, got %+v", result)
		}
	})
}

func resourceLifecycleProgram(t *testing.T, fixture string) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", fixture))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("%s: fixture failed to check: %+v", fixture, checked.Diagnostics)
	}
	return checked.Program
}

func releaseOperations(function core.Function, blockID string) []core.LinearOperation {
	byID := make(map[string]core.LinearOperation, len(function.Linear.Operations))
	for _, operation := range function.Linear.Operations {
		byID[operation.ID] = operation
	}
	var block core.Block
	for _, candidate := range function.Linear.Blocks {
		if candidate.ID == blockID {
			block = candidate
		}
	}
	var releases []core.LinearOperation
	for _, opID := range block.OperationIDs {
		if operation := byID[opID]; operation.Kind == core.OpRelease {
			releases = append(releases, operation)
		}
	}
	return releases
}

// TestForeignContractInternallyValidated proves corevalidate independently
// validates a declared core.RecordLayout's internal consistency -- offsets
// ascending, offsets plus sizes within the declared record size, no
// duplicate field name -- reading only the core artifact (Task 04-03-01).
func TestForeignContractInternallyValidated(t *testing.T) {
	valid := foreignAcquireProgram(t)
	if result := corevalidate.Validate(valid); !result.Valid {
		t.Fatalf("valid foreign-call core rejected: %+v", result)
	}

	t.Run("duplicate field name", func(t *testing.T) {
		mutated := cloneProgram(t, valid)
		layout := mutated.Functions[0].ForeignContract.Layout
		layout.Fields = append(layout.Fields, layout.Fields[0])
		result := corevalidate.Validate(mutated)
		if result.Valid || result.Problems[0].Code != "foreign.layout_invalid" {
			t.Fatalf("expected foreign.layout_invalid, got %+v", result)
		}
	})

	t.Run("offset not ascending", func(t *testing.T) {
		mutated := cloneProgram(t, valid)
		layout := mutated.Functions[0].ForeignContract.Layout
		layout.Fields = append(layout.Fields, core.LayoutField{Name: "extra", Size: 1, Alignment: 1, Offset: 0, CType: "unsigned char"})
		layout.Size = 2
		result := corevalidate.Validate(mutated)
		if result.Valid || result.Problems[0].Code != "foreign.layout_invalid" {
			t.Fatalf("expected foreign.layout_invalid, got %+v", result)
		}
	})

	t.Run("field exceeds declared record size", func(t *testing.T) {
		mutated := cloneProgram(t, valid)
		layout := mutated.Functions[0].ForeignContract.Layout
		layout.Fields[0].Size = layout.Size + 1
		result := corevalidate.Validate(mutated)
		if result.Valid || result.Problems[0].Code != "foreign.layout_invalid" {
			t.Fatalf("expected foreign.layout_invalid, got %+v", result)
		}
	})

	t.Run("layout omitted entirely", func(t *testing.T) {
		mutated := cloneProgram(t, valid)
		mutated.Functions[0].ForeignContract.Layout = nil
		result := corevalidate.Validate(mutated)
		if result.Valid || result.Problems[0].Code != "foreign.layout_invalid" {
			t.Fatalf("expected foreign.layout_invalid, got %+v", result)
		}
	})
}

// TestForeignSymbolNotIdentifierRefused is the 04-VERIFICATION.md gap-2
// falsifier (FFI-01/D-04-12): corevalidate must audit the SHAPE of
// core.ForeignContract.Symbol, not merely its presence, because cgen splices
// this exact string unsanitized into generated C at the extern declaration,
// the call-expression callee, foreignExternName, and the generated header's
// symbol comment. Every hostile subtest asserts the EXACT code
// foreign.symbol_not_identifier -- not merely "some refusal" -- so an
// unrelated earlier check firing cannot masquerade as this audit working.
func TestForeignSymbolNotIdentifierRefused(t *testing.T) {
	valid := foreignAcquireProgram(t)
	if result := corevalidate.Validate(valid); !result.Valid {
		t.Fatalf("valid foreign-call core rejected: %+v", result)
	}

	hostile := []struct {
		name   string
		symbol string
	}{
		{"semicolon and brace closing the extern and opening a new definition", "lang_res_open;}\nint injected(void){return 0;}//"},
		{"parenthesis-bearing fragment", "lang_res_open(int x)"},
		{"embedded newline", "lang_res_open\ninjected"},
		{"leading digit", "1lang_res_open"},
		{"embedded space", "lang_res_open injected"},
		{"comment terminator escaping the header comment", "lang_res_open*/int injected(void){return 0;}/*"},
		{"non-ASCII rune", "lang_res_open\u00e9"},
	}
	for _, hostileCase := range hostile {
		t.Run(hostileCase.name, func(t *testing.T) {
			mutated := cloneProgram(t, valid)
			mutated.Functions[0].ForeignContract.Symbol = hostileCase.symbol
			result := corevalidate.Validate(mutated)
			if result.Valid || result.Problems[0].Code != "foreign.symbol_not_identifier" {
				t.Fatalf("expected foreign.symbol_not_identifier, got %+v", result)
			}
		})
	}

	t.Run("identifier-shaped but unknown symbol is not refused by this check", func(t *testing.T) {
		mutated := cloneProgram(t, valid)
		mutated.Functions[0].ForeignContract.Symbol = "lang_res_open_renamed_but_well_formed"
		result := corevalidate.Validate(mutated)
		if result.Valid {
			return
		}
		if result.Problems[0].Code == "foreign.symbol_not_identifier" {
			t.Fatalf("an identifier-shaped unknown symbol must not be refused by foreign.symbol_not_identifier, got %+v", result)
		}
	})

	t.Run("end-to-end: cgen.Emit refuses before generating any C", func(t *testing.T) {
		mutated := cloneProgram(t, valid)
		mutated.Functions[0].ForeignContract.Symbol = "lang_res_open;}\nint injected(void){return 0;}//"
		generated, err := cgen.Emit(mutated)
		if err == nil {
			t.Fatalf("expected an error from cgen.Emit, got generated C:\n%s", generated)
		}
		if !strings.Contains(err.Error(), "foreign.symbol_not_identifier") {
			t.Fatalf("expected error to name foreign.symbol_not_identifier, got %v", err)
		}
		if generated != "" {
			t.Fatalf("expected empty generated string, got:\n%s", generated)
		}
	})
}

// TestForeignPolicyValueNotIdentifierRefused is 04-13's corevalidate
// falsifier (04-VERIFICATION.md gap 2b, FFI-01): a core.ForeignContract
// whose Allocator, Unwind or NonlocalExit is not a C identifier is refused
// with foreign.policy_value_not_identifier -- re-deriving, purely from the
// three flat string fields the artifact itself carries, an equivalent
// refusal to check.go's own AST-level admission gate (Task 1), so a
// corrupted core artifact that skipped check.go's gate is still caught.
func TestForeignPolicyValueNotIdentifierRefused(t *testing.T) {
	valid := foreignAcquireProgram(t)
	if result := corevalidate.Validate(valid); !result.Valid {
		t.Fatalf("valid foreign-call core rejected: %+v", result)
	}

	hostile := []struct {
		name  string
		field string
		value string
	}{
		{"Allocator: comment-terminator payload", "Allocator", "*/ int injected(void){return 1;} /*"},
		{"Unwind: comment-terminator payload", "Unwind", "*/ int injected(void){return 1;} /*"},
		{"NonlocalExit: comment-terminator payload", "NonlocalExit", "*/ int injected(void){return 1;} /*"},
		{"Allocator: semicolon and brace fragment", "Allocator", "libc;}malloc{"},
		{"Unwind: parenthesis-bearing fragment", "Unwind", "forbidden(void)"},
		{"NonlocalExit: embedded newline", "NonlocalExit", "forbidden\ninjected"},
		{"Allocator: embedded space", "Allocator", "libc malloc"},
		{"Unwind: leading digit", "Unwind", "1forbidden"},
		{"NonlocalExit: non-ASCII rune", "NonlocalExit", "forbiddené"},
	}
	for _, hostileCase := range hostile {
		t.Run(hostileCase.name, func(t *testing.T) {
			mutated := cloneProgram(t, valid)
			switch hostileCase.field {
			case "Allocator":
				mutated.Functions[0].ForeignContract.Allocator = hostileCase.value
			case "Unwind":
				mutated.Functions[0].ForeignContract.Unwind = hostileCase.value
			case "NonlocalExit":
				mutated.Functions[0].ForeignContract.NonlocalExit = hostileCase.value
			}
			result := corevalidate.Validate(mutated)
			if result.Valid || result.Problems[0].Code != "foreign.policy_value_not_identifier" {
				t.Fatalf("expected foreign.policy_value_not_identifier, got %+v", result)
			}
		})
	}

	t.Run("empty Allocator refuses (DD-04-13-01's stated consequence)", func(t *testing.T) {
		mutated := cloneProgram(t, valid)
		mutated.Functions[0].ForeignContract.Allocator = ""
		result := corevalidate.Validate(mutated)
		if result.Valid || result.Problems[0].Code != "foreign.policy_value_not_identifier" {
			t.Fatalf("expected foreign.policy_value_not_identifier, got %+v", result)
		}
	})

	t.Run("ordering: omitted Unwind still reports foreign.unwind_policy_undeclared", func(t *testing.T) {
		mutated := cloneProgram(t, valid)
		mutated.Functions[0].ForeignContract.Unwind = ""
		mutated.Functions[0].ForeignContract.Allocator = "*/ int injected(void){return 1;} /*"
		result := corevalidate.Validate(mutated)
		if result.Valid || result.Problems[0].Code != "foreign.unwind_policy_undeclared" {
			t.Fatalf("expected foreign.unwind_policy_undeclared to win, got %+v", result)
		}
	})

	t.Run("negative: identifier-shaped but unfamiliar allocator is not refused by this check", func(t *testing.T) {
		mutated := cloneProgram(t, valid)
		mutated.Functions[0].ForeignContract.Allocator = "a_different_allocator"
		result := corevalidate.Validate(mutated)
		if result.Valid {
			return
		}
		if result.Problems[0].Code == "foreign.policy_value_not_identifier" {
			t.Fatalf("an identifier-shaped unfamiliar allocator must not be refused by foreign.policy_value_not_identifier, got %+v", result)
		}
	})
}

// TestForeignContractCommentSafetyRefused is 04-13 Task 3's corevalidate
// falsifier (04-VERIFICATION.md gap 2b, FFI-01): every REMAINING
// core.ForeignContract string field cgen splices into generated C --
// Fails, InitializedState, Capture, Retention, Aliasing,
// Layout.ForeignTypeName, and each Layout.Fields[].Name / .CType -- is
// refused when it carries a comment terminator, control byte, non-ASCII
// byte, or (for the two name fields) is not identifier-shaped, or (for
// CType) is not a space-separated C type expression.
func TestForeignContractCommentSafetyRefused(t *testing.T) {
	valid := foreignAcquireProgram(t)
	if result := corevalidate.Validate(valid); !result.Valid {
		t.Fatalf("valid foreign-call core rejected: %+v", result)
	}

	commentFields := []string{"Fails", "InitializedState", "Capture", "Retention", "Aliasing", "Alias"}
	commentPayloads := []struct {
		name  string
		value string
	}{
		{"comment-terminator payload", "*/ int injected(void){return 1;} /*"},
		{"comment-opener payload", "/* injected"},
		{"embedded newline", "fully\ninjected"},
		{"embedded control byte", "fully\x01injected"},
		{"non-ASCII byte", "fully\xc3\xa9"},
	}
	setCommentField := func(mutated *core.Program, field, value string) {
		contract := mutated.Functions[0].ForeignContract
		switch field {
		case "Fails":
			contract.Fails = value
		case "InitializedState":
			contract.InitializedState = value
		case "Capture":
			contract.Capture = value
		case "Retention":
			contract.Retention = value
		case "Aliasing":
			contract.Aliasing = value
		case "Alias":
			contract.Alias = value
		}
	}
	for _, field := range commentFields {
		for _, payload := range commentPayloads {
			t.Run(field+": "+payload.name, func(t *testing.T) {
				mutated := cloneProgram(t, valid)
				setCommentField(&mutated, field, payload.value)
				result := corevalidate.Validate(mutated)
				if result.Valid || result.Problems[0].Code != "foreign.contract_field_not_c_safe" {
					t.Fatalf("expected foreign.contract_field_not_c_safe, got %+v", result)
				}
			})
		}
	}

	tokenFields := []string{"ForeignTypeName", "Fields[0].Name"}
	for _, field := range tokenFields {
		for _, payload := range commentPayloads {
			t.Run(field+": "+payload.name, func(t *testing.T) {
				mutated := cloneProgram(t, valid)
				layout := mutated.Functions[0].ForeignContract.Layout
				if field == "ForeignTypeName" {
					layout.ForeignTypeName = payload.value
				} else {
					layout.Fields[0].Name = payload.value
				}
				result := corevalidate.Validate(mutated)
				if result.Valid || result.Problems[0].Code != "foreign.contract_field_not_c_safe" {
					t.Fatalf("expected foreign.contract_field_not_c_safe, got %+v", result)
				}
			})
		}
		t.Run(field+": embedded space", func(t *testing.T) {
			mutated := cloneProgram(t, valid)
			layout := mutated.Functions[0].ForeignContract.Layout
			if field == "ForeignTypeName" {
				layout.ForeignTypeName = "lang foreign resource block"
			} else {
				layout.Fields[0].Name = "lang payload"
			}
			result := corevalidate.Validate(mutated)
			if result.Valid || result.Problems[0].Code != "foreign.contract_field_not_c_safe" {
				t.Fatalf("expected foreign.contract_field_not_c_safe, got %+v", result)
			}
		})
		t.Run(field+": leading digit", func(t *testing.T) {
			mutated := cloneProgram(t, valid)
			layout := mutated.Functions[0].ForeignContract.Layout
			if field == "ForeignTypeName" {
				layout.ForeignTypeName = "1lang_foreign_resource_block"
			} else {
				layout.Fields[0].Name = "1payload"
			}
			result := corevalidate.Validate(mutated)
			if result.Valid || result.Problems[0].Code != "foreign.contract_field_not_c_safe" {
				t.Fatalf("expected foreign.contract_field_not_c_safe, got %+v", result)
			}
		})
	}

	t.Run("CType: unsigned char is accepted", func(t *testing.T) {
		mutated := cloneProgram(t, valid)
		mutated.Functions[0].ForeignContract.Layout.Fields[0].CType = "unsigned char"
		if result := corevalidate.Validate(mutated); !result.Valid {
			t.Fatalf("expected unsigned char to be accepted, got %+v", result)
		}
	})
	for _, hostileCType := range []string{"unsigned  char", " unsigned char", "unsigned char*", "1char"} {
		t.Run("CType: "+hostileCType+" refuses", func(t *testing.T) {
			mutated := cloneProgram(t, valid)
			mutated.Functions[0].ForeignContract.Layout.Fields[0].CType = hostileCType
			result := corevalidate.Validate(mutated)
			if result.Valid || result.Problems[0].Code != "foreign.contract_field_not_c_safe" {
				t.Fatalf("expected foreign.contract_field_not_c_safe, got %+v", result)
			}
		})
	}

	t.Run("negative: empty Capture still refuses with foreign.obligation_undeclared", func(t *testing.T) {
		mutated := cloneProgram(t, valid)
		mutated.Functions[0].ForeignContract.Capture = ""
		result := corevalidate.Validate(mutated)
		if result.Valid || result.Problems[0].Code != "foreign.obligation_undeclared" {
			t.Fatalf("expected foreign.obligation_undeclared, got %+v", result)
		}
	})
	t.Run("negative: empty Layout.Fields[0].CType is accepted (cgen defaults it)", func(t *testing.T) {
		mutated := cloneProgram(t, valid)
		mutated.Functions[0].ForeignContract.Layout.Fields[0].CType = ""
		if result := corevalidate.Validate(mutated); !result.Valid {
			t.Fatalf("expected an empty CType to be accepted, got %+v", result)
		}
	})
}

// TestCorevalidateRefusesUnsafeAliasField is D-05-36's dedicated regression
// proof (closing the carried D-04-33 gap): a comment-terminating Alias value
// is refused with foreign.contract_field_not_c_safe, exactly like its
// Aliasing sibling. TestForeignContractCommentSafetyRefused above already
// exercises "Alias" through its own commentFields table; this test is the
// plan's named, standalone assertion of the same fact, so removing the new
// Alias clause from foreignContractFieldsCSafe fails THIS test directly.
func TestCorevalidateRefusesUnsafeAliasField(t *testing.T) {
	valid := foreignAcquireProgram(t)
	if result := corevalidate.Validate(valid); !result.Valid {
		t.Fatalf("valid foreign-call core rejected: %+v", result)
	}
	mutated := cloneProgram(t, valid)
	mutated.Functions[0].ForeignContract.Alias = "*/ int injected(void){return 1;} /*"
	result := corevalidate.Validate(mutated)
	if result.Valid || result.Problems[0].Code != "foreign.contract_field_not_c_safe" {
		t.Fatalf("expected foreign.contract_field_not_c_safe for a comment-terminating Alias value, got %+v", result)
	}
}

// TestAllocatorIdentityMismatchRejected proves T-04-14's allocator-identity
// requirement is refused independently by corevalidate: a release whose
// Allocator differs from its own acquisition's is rejected purely from the
// core artifact, never trusting check's own bookkeeping (Task 04-03-01).
func TestAllocatorIdentityMismatchRejected(t *testing.T) {
	valid := resourceLifecycleProgram(t, "acquire_three_success.lang")
	if result := corevalidate.Validate(valid); !result.Valid {
		t.Fatalf("valid resource-lifecycle core rejected: %+v", result)
	}
	mutated := cloneProgram(t, valid)
	function := &mutated.Functions[0]
	for index := range function.Linear.Operations {
		if function.Linear.Operations[index].Kind == core.OpRelease {
			function.Linear.Operations[index].Allocator = "a_different_allocator"
			break
		}
	}
	result := corevalidate.Validate(mutated)
	if result.Valid || result.Problems[0].Code != "foreign.release_allocator_mismatch" {
		t.Fatalf("expected foreign.release_allocator_mismatch, got %+v", result)
	}
}

// TestValidatorRederivesReleaseOrder proves corevalidate's independent
// backward-from-failure-edge rederivation (D-04-07/D-12a) agrees with the
// shipped artifact on all four Phase 4 plan-02 fixtures.
func TestValidatorRederivesReleaseOrder(t *testing.T) {
	for _, fixture := range []string{
		"acquire_three_success.lang", "acquire_three_fail_second.lang", "acquire_three_fail_third.lang", "discard_because.lang",
	} {
		program := resourceLifecycleProgram(t, fixture)
		if result := corevalidate.Validate(program); !result.Valid {
			t.Fatalf("%s: valid resource-lifecycle core rejected: %+v", fixture, result)
		}
	}
}

// TestReleaseOrderMutationMatrix proves each of moved, dropped, duplicated,
// and invented raises core.release_order_mismatch on an otherwise valid
// artifact (T-04-11).
func TestReleaseOrderMutationMatrix(t *testing.T) {
	valid := resourceLifecycleProgram(t, "acquire_three_success.lang")
	if result := corevalidate.Validate(valid); !result.Valid {
		t.Fatalf("valid resource-lifecycle core rejected: %+v", result)
	}
	successBlockID := valid.Functions[0].ID + ":block:success"

	t.Run("moved", func(t *testing.T) {
		mutated := cloneProgram(t, valid)
		function := &mutated.Functions[0]
		releases := releaseOperations(*function, successBlockID)
		if len(releases) < 2 {
			t.Fatal("expected at least two releases in the success block")
		}
		for index := range function.Linear.Operations {
			if function.Linear.Operations[index].ID == releases[0].ID {
				function.Linear.Operations[index].ReleasesOperationID = releases[1].ReleasesOperationID
			}
		}
		result := corevalidate.Validate(mutated)
		// Plan 10-05 Task 2 (D-10-34) added a SECOND independent knower --
		// peerCalleeFrameDrained, checked earlier in the same replay than
		// checkReleaseOrder -- that also observes this exact corruption
		// (the acquisition releases[0] used to discharge is now genuinely
		// unreleased): v.check's first-problem-wins convention means either
		// code can legitimately win the race. Any fail-closed rejection is
		// the evidence, never one pinned code -- the same reasoning
		// "dropped"/"duplicated" below already state for an identical
		// reason.
		if result.Valid || (result.Problems[0].Code != "core.release_order_mismatch" && result.Problems[0].Code != "core.callee_frame_not_drained") {
			t.Fatalf("expected core.release_order_mismatch or core.callee_frame_not_drained, got %+v", result)
		}
	})

	t.Run("dropped", func(t *testing.T) {
		mutated := cloneProgram(t, valid)
		function := &mutated.Functions[0]
		releases := releaseOperations(*function, successBlockID)
		if len(releases) == 0 {
			t.Fatal("expected at least one release in the success block")
		}
		var operations []core.LinearOperation
		for _, operation := range function.Linear.Operations {
			if operation.ID == releases[0].ID {
				continue
			}
			operations = append(operations, operation)
		}
		function.Linear.Operations = operations
		var blocks []core.Block
		for _, block := range function.Linear.Blocks {
			if block.ID == successBlockID {
				var opIDs []string
				for _, opID := range block.OperationIDs {
					if opID == releases[0].ID {
						continue
					}
					opIDs = append(opIDs, opID)
				}
				block.OperationIDs = opIDs
			}
			blocks = append(blocks, block)
		}
		function.Linear.Blocks = blocks
		result := corevalidate.Validate(mutated)
		// Dropping one operation from the flat Operations list also desyncs
		// the generic per-operation ordinal invariant (D-13's "operation N is
		// at position N"), so a fail-closed rejection is guaranteed even
		// though the specific code raised may be that generic ordinal check
		// rather than core.release_order_mismatch -- either is proof the
		// corruption is caught, never silently accepted.
		if result.Valid {
			t.Fatalf("expected a rejection, got valid: %+v", result)
		}
	})

	t.Run("duplicated", func(t *testing.T) {
		mutated := cloneProgram(t, valid)
		function := &mutated.Functions[0]
		releases := releaseOperations(*function, successBlockID)
		if len(releases) == 0 {
			t.Fatal("expected at least one release in the success block")
		}
		duplicate := releases[0]
		duplicate.ID = duplicate.ID + ":duplicate"
		var operations []core.LinearOperation
		var blocks []core.Block
		for _, operation := range function.Linear.Operations {
			operations = append(operations, operation)
			if operation.ID == releases[0].ID {
				operations = append(operations, duplicate)
			}
		}
		function.Linear.Operations = operations
		for _, block := range function.Linear.Blocks {
			if block.ID == successBlockID {
				var opIDs []string
				for _, opID := range block.OperationIDs {
					opIDs = append(opIDs, opID)
					if opID == releases[0].ID {
						opIDs = append(opIDs, duplicate.ID)
					}
				}
				block.OperationIDs = opIDs
			}
			blocks = append(blocks, block)
		}
		function.Linear.Blocks = blocks
		result := corevalidate.Validate(mutated)
		// Duplicating an operation also desyncs the generic per-operation
		// ordinal/point-identity invariants -- same reasoning as "dropped"
		// above: any fail-closed rejection is the evidence, not one exact
		// code.
		if result.Valid {
			t.Fatalf("expected a rejection, got valid: %+v", result)
		}
	})

	t.Run("invented", func(t *testing.T) {
		mutated := cloneProgram(t, valid)
		function := &mutated.Functions[0]
		releases := releaseOperations(*function, successBlockID)
		if len(releases) == 0 {
			t.Fatal("expected at least one release in the success block")
		}
		for index := range function.Linear.Operations {
			if function.Linear.Operations[index].ID == releases[0].ID {
				function.Linear.Operations[index].ReleasesOperationID = "s1:invented:op:absent"
			}
		}
		result := corevalidate.Validate(mutated)
		// See the identical "moved" subtest's own comment above: Plan
		// 10-05's peerCalleeFrameDrained independently observes this same
		// corruption too.
		if result.Valid || (result.Problems[0].Code != "core.release_order_mismatch" && result.Problems[0].Code != "core.callee_frame_not_drained") {
			t.Fatalf("expected core.release_order_mismatch or core.callee_frame_not_drained, got %+v", result)
		}
	})
}

// TestPeerCalleeFrameDrainedRefusesGenuinelyAbandonedAcquisition is Plan
// 10-05 Task 2's own dedicated end-to-end proof (D-10-33/D-10-34), driven
// off a REAL checked fixture (session.Check via resourceLifecycleProgram,
// never a hand-built core.Program): all three of
// acquire_three_success.lang's tracked acquisitions are repointed to a
// FABRICATED, non-existent ReleasesOperationID, making each one genuinely
// unreleased -- peerCalleeFrameDrained runs INSIDE derivePeerSignature,
// called at recordSummaryPeer's own entry point, strictly BEFORE
// checkReleaseOrder's own later rederivation in the same replay -- so
// core.CalleeFrameNotDrained is the exact FIRST problem, proving this
// invariant is genuinely wired into the validator's own pipeline, not
// merely a passing unit test of the bare predicate.
func TestPeerCalleeFrameDrainedRefusesGenuinelyAbandonedAcquisition(t *testing.T) {
	valid := resourceLifecycleProgram(t, "acquire_three_success.lang")
	mutated := cloneProgram(t, valid)
	function := &mutated.Functions[0]
	successBlockID := function.ID + ":block:success"
	releases := releaseOperations(*function, successBlockID)
	if len(releases) == 0 {
		t.Fatal("expected at least one release in the success block")
	}
	fabricated := map[string]bool{}
	for _, release := range releases {
		fabricated[release.ID] = true
	}
	for index := range function.Linear.Operations {
		if fabricated[function.Linear.Operations[index].ID] {
			function.Linear.Operations[index].ReleasesOperationID = "s1:invented:op:absent:" + function.Linear.Operations[index].ID
		}
	}
	result := corevalidate.Validate(mutated)
	if result.Valid {
		t.Fatalf("expected a rejection for a genuinely abandoned acquisition, got valid: %+v", result)
	}
	if result.Problems[0].Code != core.CalleeFrameNotDrained {
		t.Fatalf("expected %s as the FIRST problem, got %+v", core.CalleeFrameNotDrained, result.Problems)
	}
}

// TestPeerCalleeFrameDrainedAdmitsFullyDrainedFixtures proves the invariant
// does not move a single ALREADY-ADMITTED Phase 4 fixture to refused
// (this plan's own must_have): every combination of try/discard already
// covered by TestValidatorRederivesReleaseOrder still validates once this
// invariant is wired in.
func TestPeerCalleeFrameDrainedAdmitsFullyDrainedFixtures(t *testing.T) {
	for _, fixture := range []string{
		"acquire_three_success.lang", "acquire_three_fail_second.lang", "acquire_three_fail_third.lang",
		"discard_because.lang", "foreign_acquire_one.lang", "nonlocal_exit_probe.lang",
	} {
		program := resourceLifecycleProgram(t, fixture)
		if result := corevalidate.Validate(program); !result.Valid {
			t.Fatalf("%s: expected a fully drained fixture to validate, got %+v", fixture, result)
		}
	}
}

// TestReleaseOrderValidationWorkSeries proves the validator's release-order
// rederivation cost is linear in the number of blocks plus edges plus
// operations, demonstrated at three sizes (T-04-12): a discard-only fixture
// with no tracked acquisitions, a one-acquisition tracer, and a
// three-acquisition resource-lifecycle fixture.
func TestReleaseOrderValidationWorkSeries(t *testing.T) {
	sizes := []struct {
		fixture      string
		acquisitions int
	}{
		{"discard_because.lang", 0},
		{"foreign_acquire_one.lang", 1},
		{"acquire_three_success.lang", 3},
	}
	var series []int
	for _, size := range sizes {
		program := resourceLifecycleProgram(t, size.fixture)
		result := corevalidate.Validate(program)
		if !result.Valid {
			t.Fatalf("%s: valid core rejected: %+v", size.fixture, result)
		}
		series = append(series, result.Checks)
	}
	for index := 1; index < len(series); index++ {
		if series[index] <= series[index-1] {
			t.Fatalf("counted work series is not increasing: %v", series)
		}
	}
}

// TestMergeTerminalBlockDivergentReleaseSetsRefused is the 04-VERIFICATION.md
// gap-1 falsifier: it hand-constructs the exact shape checkReleaseOrder used
// to SKIP before commit 16fb0c9 -- a terminal block reachable by more than
// one incoming edge -- where the two chains genuinely disagree on what should
// be released. The added edge runs from the function's entry block (which
// completed only acquisition A) directly into the success block (whose fixed
// release list expects all three of A, B, and C released). Walking backward
// from the corrupted edge rederives a strictly shorter expected release list
// than the block's actual one, so corevalidate must refuse it with
// core.release_order_mismatch. This test is the mutation-kill falsifier for
// the per-incoming-edge rederivation added by 16fb0c9: reverting that hunk
// (restoring the pre-fix "len(incoming) != 1 { continue }" skip) turns this
// test red, since the corrupted block would then be silently skipped instead
// of refused -- see 04-08-SUMMARY.md for the recorded revert-and-fail output.
func TestMergeTerminalBlockDivergentReleaseSetsRefused(t *testing.T) {
	valid := resourceLifecycleProgram(t, "acquire_three_success.lang")
	if result := corevalidate.Validate(valid); !result.Valid {
		t.Fatalf("valid resource-lifecycle core rejected: %+v", result)
	}
	functionID := valid.Functions[0].ID
	entryBlockID := functionID + ":block:entry"
	successBlockID := functionID + ":block:success"

	mutated := cloneProgram(t, valid)
	function := &mutated.Functions[0]
	for index := range function.Linear.Blocks {
		if function.Linear.Blocks[index].ID == entryBlockID {
			function.Linear.Blocks[index].Successors = append(function.Linear.Blocks[index].Successors, successBlockID)
		}
	}
	function.Linear.Edges = append(function.Linear.Edges, core.Edge{
		ID: functionID + ":edge:corrupt:entry:success", FromBlockID: entryBlockID, ToBlockID: successBlockID, Pattern: "ok",
	})

	result := corevalidate.Validate(mutated)
	if result.Valid || result.Problems[0].Code != "core.release_order_mismatch" {
		t.Fatalf("expected core.release_order_mismatch, got %+v", result)
	}
}

// TestMergeTerminalBlockAgreeingChainsAccepted is the sibling negative-result
// row for the falsifier above: a SECOND incoming edge into the same success
// block, added from the SAME source (the last acquisition step) rather than
// an earlier one, rederives the identical expected release list the block
// already carries. Without this test, the previous test could be passing
// merely because ANY second incoming edge is refused -- exactly the "exactly
// one incoming edge" structural form Task 1 declined (it would also refuse
// discard_because.lang's legitimate merge).
func TestMergeTerminalBlockAgreeingChainsAccepted(t *testing.T) {
	valid := resourceLifecycleProgram(t, "acquire_three_success.lang")
	if result := corevalidate.Validate(valid); !result.Valid {
		t.Fatalf("valid resource-lifecycle core rejected: %+v", result)
	}
	functionID := valid.Functions[0].ID
	successBlockID := functionID + ":block:success"

	var agreeingSourceBlockID string
	for _, edge := range valid.Functions[0].Linear.Edges {
		if edge.ToBlockID == successBlockID && edge.Pattern == "ok" {
			agreeingSourceBlockID = edge.FromBlockID
		}
	}
	if agreeingSourceBlockID == "" {
		t.Fatal("expected an existing ok edge into the success block")
	}

	mutated := cloneProgram(t, valid)
	function := &mutated.Functions[0]
	function.Linear.Edges = append(function.Linear.Edges, core.Edge{
		ID: functionID + ":edge:corrupt:agreeing", FromBlockID: agreeingSourceBlockID, ToBlockID: successBlockID, Pattern: "ok",
	})

	result := corevalidate.Validate(mutated)
	if !result.Valid {
		t.Fatalf("expected two agreeing incoming edges to still validate, got %+v", result)
	}
}

// TestInteriorMergeDivergentHistoriesRefused is the 04-VERIFICATION.md gap-1
// falsifier one hop earlier than TestMergeTerminalBlockDivergentReleaseSetsRefused:
// it hand-corrupts an INTERIOR (non-terminal) block's incoming ok edges, not
// a terminal one. Before this plan's fix, okEdgeInto was a singular
// map[string]core.Edge populated by last-writer-wins assignment, so when two
// ok edges targeted the same interior block only the LAST one written was
// ever rederived -- the other declared history was silently discarded by map
// iteration order rather than examined. The two subtests differ only in
// where the corrupt edge is placed in linear.Edges, because the singular map
// keeps whichever edge is written LAST: "before" keeps the honest edge (so
// the unfixed validator ACCEPTS the corrupted program outright), while
// "after" keeps the corrupt edge (so the unfixed validator refuses, but with
// core.release_order_mismatch from the outer length comparison, never having
// examined the interior merge at all). Both orderings must refuse with the
// new core.release_order_merge_mismatch code once the fix lands, proving the
// interior comparison actually ran.
func TestInteriorMergeDivergentHistoriesRefused(t *testing.T) {
	valid := resourceLifecycleProgram(t, "acquire_three_success.lang")
	if result := corevalidate.Validate(valid); !result.Valid {
		t.Fatalf("valid resource-lifecycle core rejected: %+v", result)
	}
	functionID := valid.Functions[0].ID
	entryBlockID := functionID + ":block:entry"
	interiorBlockID := functionID + ":block:step:2"

	honestEdgeIndex := -1
	for index, edge := range valid.Functions[0].Linear.Edges {
		if edge.Pattern == "ok" && edge.ToBlockID == interiorBlockID {
			honestEdgeIndex = index
		}
	}
	if honestEdgeIndex == -1 {
		t.Fatalf("expected an existing ok edge into %s", interiorBlockID)
	}

	construct := func(t *testing.T, insertBeforeHonest bool) core.Program {
		t.Helper()
		mutated := cloneProgram(t, valid)
		function := &mutated.Functions[0]
		edgesBefore := len(function.Linear.Edges)
		corrupt := core.Edge{
			ID: functionID + ":edge:corrupt:entry:step2", FromBlockID: entryBlockID, ToBlockID: interiorBlockID, Pattern: "ok",
		}
		if insertBeforeHonest {
			edges := make([]core.Edge, 0, len(function.Linear.Edges)+1)
			edges = append(edges, function.Linear.Edges[:honestEdgeIndex]...)
			edges = append(edges, corrupt)
			edges = append(edges, function.Linear.Edges[honestEdgeIndex:]...)
			function.Linear.Edges = edges
		} else {
			function.Linear.Edges = append(function.Linear.Edges, corrupt)
		}
		for index := range function.Linear.Blocks {
			if function.Linear.Blocks[index].ID == entryBlockID {
				function.Linear.Blocks[index].Successors = append(function.Linear.Blocks[index].Successors, interiorBlockID)
			}
		}
		if got := len(function.Linear.Edges); got != edgesBefore+1 {
			t.Fatalf("expected edge count to grow by exactly 1, got %d -> %d", edgesBefore, got)
		}
		return mutated
	}

	t.Run("corrupt edge before the honest edge", func(t *testing.T) {
		mutated := construct(t, true)
		result := corevalidate.Validate(mutated)
		if result.Valid || result.Problems[0].Code != "core.release_order_merge_mismatch" {
			t.Fatalf("expected core.release_order_merge_mismatch, got %+v", result)
		}
	})

	t.Run("corrupt edge after the honest edge", func(t *testing.T) {
		mutated := construct(t, false)
		result := corevalidate.Validate(mutated)
		if result.Valid || result.Problems[0].Code != "core.release_order_merge_mismatch" {
			t.Fatalf("expected core.release_order_merge_mismatch, got %+v", result)
		}
	})
}

// TestInteriorMergeAgreeingHistoriesAccepted is the sibling negative-result
// row for the falsifier above: a SECOND ok edge into the same interior block
// from the SAME source the honest edge already comes from rederives an
// identical history, so the program must still validate. Without this test,
// the falsifier above could be passing merely because ANY second ok edge
// into an interior block is refused -- an over-refusal that would narrow the
// language exactly the way an "exactly one incoming edge" structural form
// already declined at the terminal-block loop.
func TestInteriorMergeAgreeingHistoriesAccepted(t *testing.T) {
	valid := resourceLifecycleProgram(t, "acquire_three_success.lang")
	if result := corevalidate.Validate(valid); !result.Valid {
		t.Fatalf("valid resource-lifecycle core rejected: %+v", result)
	}
	functionID := valid.Functions[0].ID
	interiorBlockID := functionID + ":block:step:2"

	var agreeingSourceBlockID string
	for _, edge := range valid.Functions[0].Linear.Edges {
		if edge.Pattern == "ok" && edge.ToBlockID == interiorBlockID {
			agreeingSourceBlockID = edge.FromBlockID
		}
	}
	if agreeingSourceBlockID == "" {
		t.Fatalf("expected an existing ok edge into %s", interiorBlockID)
	}

	mutated := cloneProgram(t, valid)
	function := &mutated.Functions[0]
	function.Linear.Edges = append(function.Linear.Edges, core.Edge{
		ID: functionID + ":edge:corrupt:agreeing:step2", FromBlockID: agreeingSourceBlockID, ToBlockID: interiorBlockID, Pattern: "ok",
	})

	result := corevalidate.Validate(mutated)
	if !result.Valid {
		t.Fatalf("expected two agreeing incoming edges into an interior block to still validate, got %+v", result)
	}
}

// TestTerminalBlockWithNoIncomingEdgeRefused removes every incoming edge into
// a non-entry terminal block. Before Task 3's structural peer check exists,
// checkReleaseOrder's own core.release_order_indeterminate refusal catches
// this; after Task 3 lands, blocksAndEdges' core.terminal_block_unreachable
// fires first (it runs earlier in Validate). The test accepts either code so
// it is green both before and after Task 3, rather than encoding which check
// fires first.
func TestTerminalBlockWithNoIncomingEdgeRefused(t *testing.T) {
	valid := resourceLifecycleProgram(t, "acquire_three_success.lang")
	if result := corevalidate.Validate(valid); !result.Valid {
		t.Fatalf("valid resource-lifecycle core rejected: %+v", result)
	}
	functionID := valid.Functions[0].ID
	successBlockID := functionID + ":block:success"

	mutated := cloneProgram(t, valid)
	function := &mutated.Functions[0]
	var edges []core.Edge
	for _, edge := range function.Linear.Edges {
		if edge.ToBlockID == successBlockID {
			continue
		}
		edges = append(edges, edge)
	}
	function.Linear.Edges = edges
	for index := range function.Linear.Blocks {
		var successors []string
		for _, successor := range function.Linear.Blocks[index].Successors {
			if successor == successBlockID {
				continue
			}
			successors = append(successors, successor)
		}
		function.Linear.Blocks[index].Successors = successors
	}

	result := corevalidate.Validate(mutated)
	if result.Valid {
		t.Fatalf("expected terminal block with no incoming edge to be refused, got valid: %+v", result)
	}
	code := result.Problems[0].Code
	if code != "core.release_order_indeterminate" && code != "core.terminal_block_unreachable" {
		t.Fatalf("expected core.release_order_indeterminate or core.terminal_block_unreachable, got %+v", result.Problems)
	}
}

// TestLegitimateDiscardMergeStillValidates is the regression guard for Task
// 1's option A: discard_because.lang's core program is itself a merge
// terminal block (entry's ok edge and err edge both target the same success
// block, since `discard`'s outcome is deliberately ignored on both paths) and
// must keep validating under the per-incoming-edge rederivation.
func TestLegitimateDiscardMergeStillValidates(t *testing.T) {
	program := resourceLifecycleProgram(t, "discard_because.lang")
	result := corevalidate.Validate(program)
	if !result.Valid {
		t.Fatalf("expected discard_because.lang to validate, got %+v", result)
	}
}

// TestTerminalBlockUnreachableRefused proves the Task 3 structural peer check
// in blocksAndEdges: a non-entry block terminated by core.OpReturn with zero
// incoming edges is refused with core.terminal_block_unreachable, independent
// of checkReleaseOrder (which never even runs, since blocksAndEdges executes
// first in Validate and returns false immediately).
func TestTerminalBlockUnreachableRefused(t *testing.T) {
	valid := resourceLifecycleProgram(t, "acquire_three_success.lang")
	if result := corevalidate.Validate(valid); !result.Valid {
		t.Fatalf("valid resource-lifecycle core rejected: %+v", result)
	}
	functionID := valid.Functions[0].ID
	successBlockID := functionID + ":block:success"

	mutated := cloneProgram(t, valid)
	function := &mutated.Functions[0]
	var edges []core.Edge
	for _, edge := range function.Linear.Edges {
		if edge.ToBlockID == successBlockID {
			continue
		}
		edges = append(edges, edge)
	}
	function.Linear.Edges = edges
	for index := range function.Linear.Blocks {
		var successors []string
		for _, successor := range function.Linear.Blocks[index].Successors {
			if successor == successBlockID {
				continue
			}
			successors = append(successors, successor)
		}
		function.Linear.Blocks[index].Successors = successors
	}

	result := corevalidate.Validate(mutated)
	if result.Valid || result.Problems[0].Code != "core.terminal_block_unreachable" {
		t.Fatalf("expected core.terminal_block_unreachable, got %+v", result)
	}
}

// TestCyclicOkEdgeChainRefusedNotHung is the 04-VERIFICATION.md gap-3 /
// 04-REVIEW.md CR-01 falsifier: checkReleaseOrder's rederive backward walk is
// the one graph walk in this file without a visited-set guard matching
// loanChainIndex.carriedLoans and blockReach. It hand-corrupts
// acquire_three_success.lang's declared ok-edge chain into a two-block cycle
// (step:1 <-> step:2) reaching the success block's incoming edge, so the
// backward walk started from that edge would loop forever without the guard.
// The test asserts BOTH observable properties the guard exists to provide:
// Validate RETURNS within a bounded wall-clock deadline (it does not hang),
// and the returned result carries the core.release_order_cyclic refusal
// rather than some other code or a silently truncated comparison.
func TestCyclicOkEdgeChainRefusedNotHung(t *testing.T) {
	valid := resourceLifecycleProgram(t, "acquire_three_success.lang")
	if result := corevalidate.Validate(valid); !result.Valid {
		t.Fatalf("valid resource-lifecycle core rejected: %+v", result)
	}
	functionID := valid.Functions[0].ID
	blockA := functionID + ":block:step:1"
	blockB := functionID + ":block:step:2"

	mutated := cloneProgram(t, valid)
	function := &mutated.Functions[0]
	edgesBefore := len(function.Linear.Edges)

	function.Linear.Edges = append(function.Linear.Edges,
		core.Edge{ID: functionID + ":edge:corrupt:cycle:b:a", FromBlockID: blockB, ToBlockID: blockA, Pattern: "ok"},
		core.Edge{ID: functionID + ":edge:corrupt:cycle:a:b", FromBlockID: blockA, ToBlockID: blockB, Pattern: "ok"},
	)
	for index := range function.Linear.Blocks {
		switch function.Linear.Blocks[index].ID {
		case blockA:
			function.Linear.Blocks[index].Successors = append(function.Linear.Blocks[index].Successors, blockB)
		case blockB:
			function.Linear.Blocks[index].Successors = append(function.Linear.Blocks[index].Successors, blockA)
		}
	}

	if got := len(function.Linear.Edges); got != edgesBefore+2 {
		t.Fatalf("expected edge count to grow by exactly 2, got %d -> %d", edgesBefore, got)
	}

	type outcome struct {
		result corevalidate.Result
	}
	done := make(chan outcome, 1)
	go func() {
		done <- outcome{result: corevalidate.Validate(mutated)}
	}()

	select {
	case got := <-done:
		if got.result.Valid || got.result.Problems[0].Code != "core.release_order_cyclic" {
			t.Fatalf("expected core.release_order_cyclic, got %+v", got.result)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("corevalidate.Validate hung on a cyclic ok-edge chain instead of refusing it")
	}
}

// TestAcyclicChainsStillValidateUnderCycleGuard is the accepting-path
// regression guard for the cycle guard above: every acyclic Phase 4 fixture,
// including discard_because.lang's legitimate two-incoming-edge merge, must
// keep validating with UNCHANGED counted work -- proving the guard costs
// nothing on the path it is not meant to refuse.
func TestAcyclicChainsStillValidateUnderCycleGuard(t *testing.T) {
	// 04-12 added one new accepting-path v.check per core.OpForeignCall (the
	// foreign.symbol_not_identifier audit); acquire_three_success.lang
	// declares three foreign calls, so the pinned count moved from 403 to
	// 403+3=406. 04-13 Task 2 adds a second new accepting-path v.check per
	// core.OpForeignCall (foreign.policy_value_not_identifier), moving the
	// pin again to 406+3=409. 04-13 Task 3 adds a THIRD and last new
	// accepting-path v.check per core.OpForeignCall
	// (foreign.contract_field_not_c_safe), moving the pin a final time to
	// 409+3=412 (measured from the built code, not assumed). This is a
	// mechanical, expected update, not a work-formula change: LinearWorkLimit
	// and the exact-formula TestCoreValidationWorkSeries are unaffected
	// because scaleProgram declares no OpForeignCall.
	//
	// Phase 07 (D-07-29) adds one new accepting-path v.check PER OPERATION
	// (the CalleeID kind-exclusivity check, run unconditionally regardless
	// of kind): acquire_three_success.lang's checked program carries 13
	// operations, moving the pin to 412+13=425.
	//
	// 07-07 adds one FLAT new accepting-path v.check per Validate call (not
	// per operation): the whole-program cycle peer (checkCallGraphAcyclic)
	// runs exactly once, moving the pin a final time to 425+1=426 (measured
	// from the built code, not assumed).
	//
	// Plan 10-05 Task 2 (D-10-33/D-10-34) adds one FLAT new accepting-path
	// v.check PER FUNCTION DECLARATION (peerCalleeFrameDrained, called once
	// inside derivePeerSignature): acquire_three_success.lang declares
	// exactly one function, moving the pin a final time to 426+1=427.
	//
	// Phase 17 preserves the single TypeFact representation for same-type
	// functions and adds one terminal return-contract check.
	const acquireThreeSuccessChecks = 428

	result := corevalidate.Validate(resourceLifecycleProgram(t, "acquire_three_success.lang"))
	if !result.Valid {
		t.Fatalf("expected acquire_three_success.lang to validate, got %+v", result)
	}
	if result.Checks != acquireThreeSuccessChecks {
		t.Fatalf("expected counted work with the terminal return contract check %d, got %d", acquireThreeSuccessChecks, result.Checks)
	}

	discardResult := corevalidate.Validate(resourceLifecycleProgram(t, "discard_because.lang"))
	if !discardResult.Valid {
		t.Fatalf("expected discard_because.lang to validate, got %+v", discardResult)
	}
}

// callerCalleeProgram builds a synthetic two-function core.Program (a caller
// with a single OpCall, immediately returned, plus a trivial callee) for
// Phase 07's CalleeID invariant tests (D-07-29). calleeID lets each test
// construct the exact CalleeID value under test -- empty, a real ID, or a
// bogus one -- without threading a full check.Program() build through it.
func callerCalleeProgram(calleeID string) core.Program {
	callerID := "s1:test:fn:caller"
	calleeFnID := "s1:test:fn:callee"
	callerTypeID := callerID + ":type:0"
	callerParamID := callerID + ":place:0"
	callerTargetID := callerID + ":place:1"
	calleeTypeID := calleeFnID + ":type:0"
	calleeParamID := calleeFnID + ":place:0"
	byteFact := core.TypeFact{
		ID: callerTypeID, Shape: core.TypeRef{Constructor: "Byte", Arguments: []core.TypeRef{}},
		Abilities:         []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
		NegativeWitnesses: []core.AbilityWitness{},
	}
	calleeFact := byteFact
	calleeFact.ID = calleeTypeID
	return core.Program{
		Schema: core.Schema1, Module: "test", ModuleID: "s1:test:module:test",
		Functions: []core.Function{
			{
				ID: callerID, Name: "caller", EntryPointID: callerID + ":point:entry", ReturnPointID: callerID + ":point:return",
				Parameter: core.Parameter{ID: callerParamID, Name: "value", Type: "Byte"}, ReturnType: "Byte",
				Linear: &core.LinearBody{
					ID:     callerID + ":linear",
					Types:  []core.TypeFact{byteFact},
					Places: []core.Place{{ID: callerParamID, Name: "value", TypeID: callerTypeID}, {ID: callerTargetID, Name: "result", TypeID: callerTypeID}},
					Operations: []core.LinearOperation{
						{ID: callerID + ":op:0", PointID: callerID + ":point:linear:0", Kind: core.OpCall, SourceID: callerParamID, TargetID: callerTargetID, TypeID: callerTypeID, CalleeID: calleeID},
						{ID: callerID + ":op:1", PointID: callerID + ":point:linear:1", Kind: core.OpReturn, SourceID: callerTargetID, TypeID: callerTypeID},
					},
				},
			},
			{
				ID: calleeFnID, Name: "callee", EntryPointID: calleeFnID + ":point:entry", ReturnPointID: calleeFnID + ":point:return",
				Parameter: core.Parameter{ID: calleeParamID, Name: "value", Type: "Byte"}, ReturnType: "Byte",
				Linear: &core.LinearBody{
					ID:         calleeFnID + ":linear",
					Types:      []core.TypeFact{calleeFact},
					Places:     []core.Place{{ID: calleeParamID, Name: "value", TypeID: calleeTypeID}},
					Operations: []core.LinearOperation{{ID: calleeFnID + ":op:0", PointID: calleeFnID + ":point:linear:0", Kind: core.OpReturn, SourceID: calleeParamID, TypeID: calleeTypeID}},
				},
			},
		},
	}
}

// callTypeContractProgram builds a synthetic two-function core.Program
// (07-09 Task 2), a materially different construction from
// callerCalleeProgram above: it parametrizes the CALLER's own type
// constructor and the CALLEE's declared Parameter.Type/ReturnType
// independently, so a test can construct exactly the mismatched-argument
// or mismatched-return shape it needs -- including a callee whose declared
// return type differs from its declared parameter type, a shape NO source
// program can produce (this language's sameType invariant forces exactly
// one type fact per function) -- reaching corevalidate's disjoint
// synthetic input space (D-07-19) the same way callerCalleeProgram already
// does for the CalleeID invariants. Every place in the CALLER (parameter,
// call target, and the final OpReturn's source) shares the caller's own
// single type fact, exactly as check's promoted derivation produces on
// every admitted program (07-09 Task 1 Test 6) -- there is no second
// caller-side type fact, because corevalidate's own PRE-EXISTING
// source.TypeID == operation.TypeID and target.TypeID == operation.TypeID
// laws (targetMatches) already force them equal. The return-type-mismatch
// shape is reached by diverging calleeReturnType from callerType, not by
// diverging the caller's own places from each other.
func callTypeContractProgram(t *testing.T, callerType, calleeParameterType, calleeReturnType string) core.Program {
	t.Helper()
	callerID := "s1:test:fn:caller"
	calleeFnID := "s1:test:fn:callee"
	callerTypeID := callerID + ":type:0"
	callerParamID := callerID + ":place:0"
	callerTargetID := callerID + ":place:1"
	calleeTypeID := calleeFnID + ":type:0"
	calleeParamID := calleeFnID + ":place:0"
	callerFact := mustDeriveTypeFactForTest(t, callerTypeID, callerType)
	calleeFact := mustDeriveTypeFactForTest(t, calleeTypeID, calleeParameterType)
	return core.Program{
		Schema: core.Schema1, Module: "test", ModuleID: "s1:test:module:test",
		Functions: []core.Function{
			{
				ID: callerID, Name: "caller", EntryPointID: callerID + ":point:entry", ReturnPointID: callerID + ":point:return",
				Parameter: core.Parameter{ID: callerParamID, Name: "value", Type: callerType}, ReturnType: callerType,
				Linear: &core.LinearBody{
					ID:     callerID + ":linear",
					Types:  []core.TypeFact{callerFact},
					Places: []core.Place{{ID: callerParamID, Name: "value", TypeID: callerTypeID}, {ID: callerTargetID, Name: "result", TypeID: callerTypeID}},
					Operations: []core.LinearOperation{
						{ID: callerID + ":op:0", PointID: callerID + ":point:linear:0", Kind: core.OpCall, SourceID: callerParamID, TargetID: callerTargetID, TypeID: callerTypeID, CalleeID: calleeFnID},
						{ID: callerID + ":op:1", PointID: callerID + ":point:linear:1", Kind: core.OpReturn, SourceID: callerTargetID, TypeID: callerTypeID},
					},
				},
			},
			{
				ID: calleeFnID, Name: "callee", EntryPointID: calleeFnID + ":point:entry", ReturnPointID: calleeFnID + ":point:return",
				Parameter: core.Parameter{ID: calleeParamID, Name: "value", Type: calleeParameterType}, ReturnType: calleeReturnType,
				Linear: &core.LinearBody{
					ID:         calleeFnID + ":linear",
					Types:      []core.TypeFact{calleeFact},
					Places:     []core.Place{{ID: calleeParamID, Name: "value", TypeID: calleeTypeID}},
					Operations: []core.LinearOperation{{ID: calleeFnID + ":op:0", PointID: calleeFnID + ":point:linear:0", Kind: core.OpReturn, SourceID: calleeParamID, TypeID: calleeTypeID}},
				},
			},
		},
	}
}

// TestPeerRefusesCallArgumentTypeMismatch is 07-09 Task 2 Test 1: a
// hand-built synthetic core.Program whose OpCall's source place resolves
// to a type fact with constructor Buffer while
// functionByID[CalleeID].Parameter.Type is Byte is refused by
// corevalidate.Validate with Problem.Code == core.CallArgumentTypeMismatch.
func TestPeerRefusesCallArgumentTypeMismatch(t *testing.T) {
	program := callTypeContractProgram(t, "Buffer", "Byte", "Buffer")
	result := corevalidate.Validate(program)
	if result.Valid {
		t.Fatalf("expected refusal for a mismatched argument type, got %+v", result)
	}
	if len(result.Problems) == 0 || result.Problems[0].Code != core.CallArgumentTypeMismatch {
		t.Fatalf("expected %s, got %+v", core.CallArgumentTypeMismatch, result.Problems)
	}
}

// TestPeerRefusesCallReturnTypeMismatch is 07-09 Task 2 Test 2: a
// hand-built synthetic core.Program whose OpCall's TARGET place resolves
// to a type fact whose constructor differs from
// functionByID[CalleeID].ReturnType is refused with
// Problem.Code == core.CallReturnTypeMismatch. This shape is unreachable
// from source (sameType forces one type fact per function) and is exactly
// what corevalidate's disjoint synthetic input space (D-07-19) exists to
// reach.
func TestPeerRefusesCallReturnTypeMismatch(t *testing.T) {
	program := callTypeContractProgram(t, "Byte", "Byte", "Buffer")
	result := corevalidate.Validate(program)
	if result.Valid {
		t.Fatalf("expected refusal for a mismatched target/return type, got %+v", result)
	}
	if len(result.Problems) == 0 || result.Problems[0].Code != core.CallReturnTypeMismatch {
		t.Fatalf("expected %s, got %+v", core.CallReturnTypeMismatch, result.Problems)
	}
}

// TestCallTypePeersIndependentOfCheck is 07-09 Task 2 Test 3
// (independence, T-07-09-03): this hand-built synthetic core.Program is
// NEVER produced by check.Program -- check never runs at all here, so
// this program's refusal cannot be a second observation of check's own
// verdict. corevalidate refuses it entirely on its own, from its own
// inputs, proving the peer is load-bearing on its own.
func TestCallTypePeersIndependentOfCheck(t *testing.T) {
	program := callTypeContractProgram(t, "Buffer", "Byte", "Buffer")
	result := corevalidate.Validate(program)
	if result.Valid {
		t.Fatalf("expected the peer to refuse independently of check, got %+v", result)
	}
	if len(result.Problems) == 0 || result.Problems[0].Code != core.CallArgumentTypeMismatch {
		t.Fatalf("expected %s, got %+v", core.CallArgumentTypeMismatch, result.Problems)
	}
}

// TestCallTypeContractFailsClosedOnAbsentFacts is 07-09 Task 2 Test 4
// (fail-closed, edge 2): an OpCall whose CalleeID resolves to a function
// with an EMPTY Parameter.Type, or whose source/target place resolves to
// no type fact at all, refuses. Absence never admits.
func TestCallTypeContractFailsClosedOnAbsentFacts(t *testing.T) {
	// The "empty declared parameter type" sub-case of this edge is proven
	// directly against the unexported checkCallTypeContract predicate in
	// corevalidate_call_type_internal_test.go (package corevalidate):
	// linearStructural's own pre-existing core.parameter_mismatch law
	// already refuses any structurally-valid core.Program whose
	// core.Function.Parameter.Type disagrees with its own type fact's
	// Shape.Constructor, so an empty ParameterType can never reach the
	// OpCall replay arm through corevalidate.Validate's full pipeline --
	// exactly the same unreachable-from-a-valid-program shape check.go's
	// own Test 3 (check_test.go) proves directly against
	// resolveCallBinding rather than through the full check.Program
	// pipeline.
	t.Run("unresolvable source type fact refuses", func(t *testing.T) {
		program := callTypeContractProgram(t, "Byte", "Byte", "Byte")
		program.Functions[0].Linear.Places[0].TypeID = "s1:test:fn:caller:type:missing"
		result := corevalidate.Validate(program)
		if result.Valid {
			t.Fatalf("expected refusal for an unresolvable source type fact, got %+v", result)
		}
	})
}

// TestCallTypeContractAdjacencyAndBoundary is 07-09 Task 2 Test 5 (edges
// 1/4): constructor equality admits; any inequality refuses, in both
// directions.
func TestCallTypeContractAdjacencyAndBoundary(t *testing.T) {
	cases := []struct {
		name       string
		callerType string
		calleeType string
		wantValid  bool
	}{
		{"Byte==Byte admits", "Byte", "Byte", true},
		{"Buffer==Buffer admits", "Buffer", "Buffer", true},
		{"Byte argument vs Buffer parameter refuses", "Byte", "Buffer", false},
		{"Buffer argument vs Byte parameter refuses", "Buffer", "Byte", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			program := callTypeContractProgram(t, tc.callerType, tc.calleeType, tc.callerType)
			result := corevalidate.Validate(program)
			if result.Valid != tc.wantValid {
				t.Fatalf("want valid=%v, got valid=%v (problems=%+v)", tc.wantValid, result.Valid, result.Problems)
			}
		})
	}
}

// TestCallTypeContractDoesNotShadowExistingCallLaws is 07-09 Task 2 Test 6
// (adjacency with existing laws): core.type_mismatch, core.invalid_target,
// core.callee_not_callable, and core.call_callee_unresolved each still
// fire for their own defect shapes; a program exhibiting one of those does
// not now report a type-contract code instead.
func TestCallTypeContractDoesNotShadowExistingCallLaws(t *testing.T) {
	t.Run("unresolved callee still reports core.call_callee_unresolved", func(t *testing.T) {
		program := callerCalleeProgram("s1:test:fn:does-not-exist")
		result := corevalidate.Validate(program)
		if result.Valid || len(result.Problems) == 0 || result.Problems[0].Code != core.CallCalleeUnresolved {
			t.Fatalf("expected %s, got %+v", core.CallCalleeUnresolved, result.Problems)
		}
	})
	t.Run("well-formed OpCall to a matching callee still validates", func(t *testing.T) {
		program := callerCalleeProgram("s1:test:fn:callee")
		result := corevalidate.Validate(program)
		if !result.Valid {
			t.Fatalf("expected a well-formed OpCall to validate cleanly, got %+v", result)
		}
	})
}

// callArgumentConsumeStraightLineProgram builds a synthetic two-function
// core.Program (07-11 Task 2, PVG-01/CR-01): the caller's straight-line
// body issues TWO core.OpCall operations from the SAME SourceID (the
// parameter place) to the SAME callee -- the double-consume shape
// 07-VERIFICATION.md PVG-01 / 07-REVIEW.md CR-01 names -- built entirely
// by the peer's own test, never once run through check.Program. The
// callee's declared ParameterType/ReturnType equal argumentType so
// checkCallTypeContract's pre-existing gates pass, isolating this test's
// own assertion to the consume rule.
func callArgumentConsumeStraightLineProgram(t *testing.T, argumentType string) core.Program {
	t.Helper()
	callerID := "s1:test:fn:caller"
	calleeFnID := "s1:test:fn:callee"
	callerTypeID := callerID + ":type:0"
	callerParamID := callerID + ":place:0"
	firstTargetID := callerID + ":place:1"
	secondTargetID := callerID + ":place:2"
	calleeTypeID := calleeFnID + ":type:0"
	calleeParamID := calleeFnID + ":place:0"
	callerFact := mustDeriveTypeFactForTest(t, callerTypeID, argumentType)
	calleeFact := mustDeriveTypeFactForTest(t, calleeTypeID, argumentType)
	return core.Program{
		Schema: core.Schema1, Module: "test", ModuleID: "s1:test:module:test",
		Functions: []core.Function{
			{
				ID: callerID, Name: "caller", EntryPointID: callerID + ":point:entry", ReturnPointID: callerID + ":point:return",
				Parameter: core.Parameter{ID: callerParamID, Name: "value", Type: argumentType}, ReturnType: argumentType,
				Linear: &core.LinearBody{
					ID:    callerID + ":linear",
					Types: []core.TypeFact{callerFact},
					Places: []core.Place{
						{ID: callerParamID, Name: "value", TypeID: callerTypeID},
						{ID: firstTargetID, Name: "first", TypeID: callerTypeID},
						{ID: secondTargetID, Name: "second", TypeID: callerTypeID},
					},
					Operations: []core.LinearOperation{
						{ID: callerID + ":op:0", PointID: callerID + ":point:linear:0", Kind: core.OpCall, SourceID: callerParamID, TargetID: firstTargetID, TypeID: callerTypeID, CalleeID: calleeFnID},
						{ID: callerID + ":op:1", PointID: callerID + ":point:linear:1", Kind: core.OpCall, SourceID: callerParamID, TargetID: secondTargetID, TypeID: callerTypeID, CalleeID: calleeFnID},
						{ID: callerID + ":op:2", PointID: callerID + ":point:linear:2", Kind: core.OpReturn, SourceID: secondTargetID, TypeID: callerTypeID},
					},
				},
			},
			{
				ID: calleeFnID, Name: "callee", EntryPointID: calleeFnID + ":point:entry", ReturnPointID: calleeFnID + ":point:return",
				Parameter: core.Parameter{ID: calleeParamID, Name: "value", Type: argumentType}, ReturnType: argumentType,
				Linear: &core.LinearBody{
					ID:         calleeFnID + ":linear",
					Types:      []core.TypeFact{calleeFact},
					Places:     []core.Place{{ID: calleeParamID, Name: "value", TypeID: calleeTypeID}},
					Operations: []core.LinearOperation{{ID: calleeFnID + ":op:0", PointID: calleeFnID + ":point:linear:0", Kind: core.OpReturn, SourceID: calleeParamID, TypeID: calleeTypeID}},
				},
			},
		},
	}
}

// callArgumentConsumeBranchProgram is callArgumentConsumeStraightLineProgram's
// branch-shaped counterpart (07-11 Task 2 Test 6): the identical
// double-consume, but the caller's three operations sit inside a single
// core.Block with no successors, routing corevalidate.Validate through
// replayBlocks (len(function.Linear.Blocks) > 0) rather than
// replayStraightLine -- proving the consume rule reaches BOTH replay
// arms through the one shared consumeCallArgument helper, not a
// duplicated copy.
func callArgumentConsumeBranchProgram(t *testing.T, argumentType string) core.Program {
	t.Helper()
	program := callArgumentConsumeStraightLineProgram(t, argumentType)
	caller := &program.Functions[0]
	entryID := caller.ID + ":block:entry"
	armID := caller.ID + ":block:arm"
	joinID := caller.ID + ":block:join"
	// Mirrors checkBranch's own real shape (check.go): an empty-ops entry
	// block dispatching to the arm block, the arm block carrying the
	// REAL operations (both core.OpCall's -- the double-consume this test
	// exists to reach -- terminated by the OpReturn), and an empty-ops
	// join block. Only a block with a non-empty OperationIDs list is
	// subject to checkReleaseOrder's core.release_order_indeterminate law
	// and the returnedBlocks/core.final_claim_mismatch completeness law,
	// so entry/join (both empty) are exempt by construction and only the
	// arm block's own termination and incoming edge matter -- exactly the
	// contract every real check.go producer (checkBranch,
	// checkForeignTracer, checkResourceLifecycle) already honors.
	caller.Linear.Blocks = []core.Block{
		{ID: entryID, PointID: caller.EntryPointID, OperationIDs: []string{}, Successors: []string{armID}},
		{ID: armID, PointID: caller.ID + ":point:arm", OperationIDs: []string{caller.Linear.Operations[0].ID, caller.Linear.Operations[1].ID, caller.Linear.Operations[2].ID}, Successors: []string{joinID}},
		{ID: joinID, PointID: caller.ID + ":point:return", OperationIDs: []string{}, Successors: []string{}},
	}
	caller.Linear.Edges = []core.Edge{
		{ID: caller.ID + ":edge:entry:arm", FromBlockID: entryID, ToBlockID: armID, Pattern: "arm"},
		{ID: caller.ID + ":edge:arm:join", FromBlockID: armID, ToBlockID: joinID, Pattern: "arm"},
	}
	return program
}

// TestPeerRefusesDoubleConsumedCallArgument is 07-11 Task 2 Test 1: a
// synthetic core.Program the peer builds itself -- two core.OpCall
// operations with the identical non-copyable (Buffer) SourceID -- is
// refused by corevalidate.Validate with core.place_uninitialized naming
// that source place. The program is hand-built, never one check
// produced.
func TestPeerRefusesDoubleConsumedCallArgument(t *testing.T) {
	program := callArgumentConsumeStraightLineProgram(t, "Buffer")
	result := corevalidate.Validate(program)
	if result.Valid {
		t.Fatalf("expected the double-consumed Buffer argument to be refused, got %+v", result)
	}
	if len(result.Problems) == 0 || result.Problems[0].Code != "core.place_uninitialized" {
		t.Fatalf("expected core.place_uninitialized, got %+v", result.Problems)
	}
	if result.Problems[0].Detail != "s1:test:fn:caller:place:0" {
		t.Fatalf("expected the problem to name the consumed source place, got %+v", result.Problems[0])
	}
}

// TestPeerAdmitsRepeatedCopyableCallArgument is 07-11 Task 2 Test 2: the
// identical synthetic shape with a COPYABLE (Byte) type is accepted --
// the peer does not over-refuse either.
func TestPeerAdmitsRepeatedCopyableCallArgument(t *testing.T) {
	program := callArgumentConsumeStraightLineProgram(t, "Byte")
	result := corevalidate.Validate(program)
	if !result.Valid {
		t.Fatalf("expected a copyable argument passed to two calls to be admitted, got %+v", result.Problems)
	}
}

// TestCallConsumePeerMatchesBranchReplayArm is 07-11 Task 2 Test 6: the
// identical double-consume, replayed through replayBlocks instead of
// replayStraightLine, is refused identically.
func TestCallConsumePeerMatchesBranchReplayArm(t *testing.T) {
	program := callArgumentConsumeBranchProgram(t, "Buffer")
	result := corevalidate.Validate(program)
	if result.Valid {
		t.Fatalf("expected the branch-replayed double-consumed Buffer argument to be refused, got %+v", result)
	}
	if len(result.Problems) == 0 || result.Problems[0].Code != "core.place_uninitialized" {
		t.Fatalf("expected core.place_uninitialized from the branch replay arm, got %+v", result.Problems)
	}

	admitted := callArgumentConsumeBranchProgram(t, "Byte")
	admittedResult := corevalidate.Validate(admitted)
	if !admittedResult.Valid {
		t.Fatalf("expected a copyable argument to still admit through the branch replay arm, got %+v", admittedResult.Problems)
	}
}

// TestCallConsumeFailsClosedOnUnresolvableType is 07-11 Task 2 Test 5
// (edge 2, fail-closed): a synthetic core.OpCall whose TypeID resolves to
// no type fact is refused -- never admitted as if copyable. An absent
// type fact is the refusing case.
func TestCallConsumeFailsClosedOnUnresolvableType(t *testing.T) {
	program := callArgumentConsumeStraightLineProgram(t, "Byte")
	program.Functions[0].Linear.Types = nil
	result := corevalidate.Validate(program)
	if result.Valid {
		t.Fatalf("expected refusal when the argument's type fact is unresolvable, got %+v", result)
	}
}

// mustCheckPhase07Fixture parses and checks a real testdata/phase07
// fixture through check.Program, mirroring readPhase07Fixture's own
// path resolution (check_test.go) from this package's directory.
func mustCheckPhase07Fixture(t *testing.T, name string) check.Result {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase07", name))
	if err != nil {
		t.Fatalf("read fixture %q: %v", name, err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("fixture %q failed to parse: %+v", name, parsed.Diagnostics)
	}
	return check.Program(parsed.Program)
}

// TestCallConsumePeerIndependentOfCheck is 07-11 Task 2 Test 3
// (independence, the coordinated-blindness assertion, mirroring
// check_test.go's TestCorevalidatePeerIndependentOfCheckCycleRefusal in
// reverse): with check's OWN consume seam disabled -- the producer's gate
// off -- corevalidate still refuses call_argument_used_twice.lang's
// REAL emitted core on its own. And with the peer's own seam disabled,
// check's own (default, never toggled) refusal still catches the same
// source. Each side refuses alone; neither is a second observation of
// the other's verdict.
func TestCallConsumePeerIndependentOfCheck(t *testing.T) {
	t.Run("peer refuses with check's own gate disabled", func(t *testing.T) {
		restore := check.SetCallArgumentConsumeSeamForTest(true)
		defer restore()

		result := mustCheckPhase07Fixture(t, "call_argument_used_twice.lang")
		if len(result.Diagnostics) != 0 {
			t.Fatalf("expected check's own consume gate to be disabled (program admitted by check), got %+v", result.Diagnostics)
		}

		validated := corevalidate.Validate(result.Program)
		if validated.Valid {
			t.Fatal("expected corevalidate's own independent peer to still refuse the double-consumed argument")
		}
		if len(validated.Problems) == 0 || validated.Problems[0].Code != "core.place_uninitialized" {
			t.Fatalf("expected core.place_uninitialized among corevalidate's problems, got %+v", validated.Problems)
		}
	})

	t.Run("check refuses with the peer's own gate disabled", func(t *testing.T) {
		restore := corevalidate.SetDisableCallArgumentConsumePeerForTest(true)
		defer restore()

		result := mustCheckPhase07Fixture(t, "call_argument_used_twice.lang")
		if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "ownership.use_after_move" {
			t.Fatalf("expected check's own refusal (unaffected by the peer's seam) to still fire, got %+v", result.Diagnostics)
		}
	})
}

// TestCallArgumentConsumeSeamSettersRestoreCleanly proves
// SetDisableCallArgumentConsumePeerForTest and
// SetForceCallArgumentConsumePeerForTest each return a restore closure
// that returns corevalidate.Validate's verdict to production behavior.
func TestCallArgumentConsumeSeamSettersRestoreCleanly(t *testing.T) {
	program := callArgumentConsumeStraightLineProgram(t, "Buffer")

	restoreDisable := corevalidate.SetDisableCallArgumentConsumePeerForTest(true)
	disabled := corevalidate.Validate(program)
	if !disabled.Valid {
		t.Fatalf("expected the disabled seam to wrongly admit the double-consume, got %+v", disabled.Problems)
	}
	restoreDisable()
	restored := corevalidate.Validate(program)
	if restored.Valid {
		t.Fatalf("expected production behavior restored after SetDisableCallArgumentConsumePeerForTest's restore, got %+v", restored)
	}

	copyableProgram := callArgumentConsumeStraightLineProgram(t, "Byte")
	restoreForce := corevalidate.SetForceCallArgumentConsumePeerForTest(true)
	forced := corevalidate.Validate(copyableProgram)
	if forced.Valid {
		t.Fatalf("expected the forced seam to wrongly refuse a copyable argument's second use, got %+v", forced)
	}
	restoreForce()
	restoredCopyable := corevalidate.Validate(copyableProgram)
	if !restoredCopyable.Valid {
		t.Fatalf("expected production behavior restored after SetForceCallArgumentConsumePeerForTest's restore, got %+v", restoredCopyable.Problems)
	}
}

// TestCallTypePeerMutationMatrix is 07-09 Task 3's peer-side completeness
// proof that BOTH SetDisableCallArgumentTypePeerForTest and
// SetDisableCallReturnTypePeerForTest are wired to a REAL, reachable
// production seam (as opposed to a seam nothing calls): each, engaged
// alone, changes corevalidate.Validate's verdict on the SAME
// argument-and-return-coupled synthetic defect program from refused to
// admitted-with-a-different-remaining-refusal or admitted outright,
// proving neither setter is dead code. The full per-control fail-closed
// proof (each predicate independently observed to make its OWN control
// go red, isolated from the other) is in
// corevalidate_call_type_internal_test.go's
// TestCheckCallTypeContractArgumentPeerSeamKilled and
// TestCheckCallTypeContractReturnPeerSeamKilled: this language's sameType
// invariant forces a self-consistent callee's ParameterType and
// ReturnType to be the SAME value, so on any WHOLE-PROGRAM synthetic
// built from a self-consistent callee, the argument-type and return-type
// predicates are mathematically the same boolean and cannot be observed
// to diverge through corevalidate.Validate's full pipeline -- exactly the
// coupling PHASE-07-DEBT.md's arity-N entry and this plan's language
// surface constraint already document for check's own mutation kill
// (check_test.go's TestCallArgumentTypeCheckMutationKilled uses the
// identical direct-API technique for the identical reason).
func TestCallTypePeerMutationMatrix(t *testing.T) {
	program := callTypeContractProgram(t, "Buffer", "Byte", "Byte")

	clean := corevalidate.Validate(program)
	if clean.Valid || len(clean.Problems) == 0 {
		t.Fatalf("expected a refusal before either mutation, got %+v", clean)
	}

	t.Run("argument_type_peer_disabled_changes_the_verdict", func(t *testing.T) {
		restore := corevalidate.SetDisableCallArgumentTypePeerForTest(true)
		defer restore()
		mutated := corevalidate.Validate(program)
		if mutated.Valid || len(mutated.Problems) == 0 || mutated.Problems[0].Code == clean.Problems[0].Code {
			t.Fatalf("expected the argument-type peer seam to change which law refuses (still-active return-type peer), got %+v", mutated)
		}
	})
	t.Run("return_type_peer_disabled_leaves_argument_refusal_intact", func(t *testing.T) {
		restore := corevalidate.SetDisableCallReturnTypePeerForTest(true)
		defer restore()
		mutated := corevalidate.Validate(program)
		if mutated.Valid || len(mutated.Problems) == 0 || mutated.Problems[0].Code != core.CallArgumentTypeMismatch {
			t.Fatalf("expected the still-active argument-type peer to keep refusing with %s, got %+v", core.CallArgumentTypeMismatch, mutated)
		}
	})
}

// TestCallConsumePeerMutationMatrix is 07-11 Task 3's peer-side
// completeness proof that both SetDisableCallArgumentConsumePeerForTest
// and SetForceCallArgumentConsumePeerForTest are wired to a REAL,
// reachable production seam, in BOTH directions (control:call.
// argument_consumed_when_noncopyable and control:call.
// copyable_argument_not_consumed): disabling the consume seam wrongly
// admits a double-consumed Buffer argument; forcing it wrongly refuses a
// repeated Byte argument.
func TestCallConsumePeerMutationMatrix(t *testing.T) {
	t.Run("disable_seam_wrongly_admits_noncopyable_double_consume", func(t *testing.T) {
		program := callArgumentConsumeStraightLineProgram(t, "Buffer")
		clean := corevalidate.Validate(program)
		if clean.Valid {
			t.Fatalf("expected a refusal before the mutation, got %+v", clean)
		}

		restore := corevalidate.SetDisableCallArgumentConsumePeerForTest(true)
		defer restore()
		mutated := corevalidate.Validate(program)
		if !mutated.Valid {
			t.Fatalf("expected the disabled consume seam to wrongly admit the double-consume, got %+v", mutated.Problems)
		}
	})
	t.Run("force_seam_wrongly_refuses_copyable_repeated_use", func(t *testing.T) {
		program := callArgumentConsumeStraightLineProgram(t, "Byte")
		clean := corevalidate.Validate(program)
		if !clean.Valid {
			t.Fatalf("expected admission before the mutation, got %+v", clean.Problems)
		}

		restore := corevalidate.SetForceCallArgumentConsumePeerForTest(true)
		defer restore()
		mutated := corevalidate.Validate(program)
		if mutated.Valid || len(mutated.Problems) == 0 || mutated.Problems[0].Code != "core.place_uninitialized" {
			t.Fatalf("expected the forced consume seam to wrongly refuse a copyable argument's second use, got %+v", mutated)
		}
	})
}

// TestOpCallEmptyCalleeIDIsRefused is Task 2 Test 1 (D-07-29): a synthetic
// core.Program with an OpCall whose CalleeID is empty is refused by
// corevalidate with a specific code, at whichever replay site the caller's
// straight-line shape reaches.
func TestOpCallEmptyCalleeIDIsRefused(t *testing.T) {
	program := callerCalleeProgram("")
	result := corevalidate.Validate(program)
	if result.Valid {
		t.Fatalf("expected refusal for an OpCall with an empty CalleeID, got %+v", result)
	}
	if len(result.Problems) == 0 || result.Problems[0].Code != "core.callee_id_missing" {
		t.Fatalf("expected core.callee_id_missing, got %+v", result.Problems)
	}
}

// TestNonOpCallWithCalleeIDIsRefused is Task 2 Test 2 (D-07-29): CalleeID is
// kind-exclusive, exactly as Allocator and ReleasesOperationID are -- an
// OpCopy carrying a non-empty CalleeID is refused.
func TestNonOpCallWithCalleeIDIsRefused(t *testing.T) {
	program := callerCalleeProgram("s1:test:fn:callee")
	program.Functions[0].Linear.Operations[1].CalleeID = "s1:test:fn:callee"
	result := corevalidate.Validate(program)
	if result.Valid {
		t.Fatalf("expected refusal for a non-OpCall operation carrying a CalleeID, got %+v", result)
	}
	if len(result.Problems) == 0 || result.Problems[0].Code != "core.callee_id_kind_exclusive" {
		t.Fatalf("expected core.callee_id_kind_exclusive, got %+v", result.Problems)
	}
}

// TestOpCallUnresolvedCalleeIsRefused is Task 2 Test 3 (D-07-45): a
// CalleeID naming no declared function is refused with its OWN typed
// identity, distinct from the (07-06) cycle code and from the two refusals
// above.
func TestOpCallUnresolvedCalleeIsRefused(t *testing.T) {
	program := callerCalleeProgram("s1:test:fn:does-not-exist")
	result := corevalidate.Validate(program)
	if result.Valid {
		t.Fatalf("expected refusal for a CalleeID naming no declared function, got %+v", result)
	}
	if len(result.Problems) == 0 || result.Problems[0].Code != core.CallCalleeUnresolved {
		t.Fatalf("expected %s, got %+v", core.CallCalleeUnresolved, result.Problems)
	}
	if core.CallCalleeUnresolved == "core.callee_id_missing" || core.CallCalleeUnresolved == "core.callee_id_kind_exclusive" || core.CallCalleeUnresolved == "core.call_graph_cycle" {
		t.Fatalf("core.CallCalleeUnresolved must be distinct from the other CalleeID refusal codes and the cycle code, got %q", core.CallCalleeUnresolved)
	}
}

// TestOpCallValidCalleeIDIsAccepted is the accepting-path counterpart: a
// well-formed OpCall whose CalleeID names a real declared function validates
// cleanly, proving the three refusals above are not vacuously always firing.
func TestOpCallValidCalleeIDIsAccepted(t *testing.T) {
	program := callerCalleeProgram("s1:test:fn:callee")
	result := corevalidate.Validate(program)
	if !result.Valid {
		t.Fatalf("expected a well-formed OpCall to validate cleanly, got %+v", result)
	}
}

// TestDisableCalleeResolutionCheckSeamSuppressesUnresolvedRefusal is Task 2
// Test 6's corevalidate half (QLT-08, D-07-41/D-07-42): with
// disableCalleeResolutionCheckForTest engaged, the SAME hand-built program
// TestOpCallUnresolvedCalleeIsRefused refuses now validates cleanly;
// restoring the seam (via defer) restores the refusal. Split from check's
// own seam test (check_test.go) because Go's export_test.go pattern only
// reaches a package's own external test package -- see 07-02-SUMMARY.md's
// identical cross-package split for the bilateral fault.
func TestDisableCalleeResolutionCheckSeamSuppressesUnresolvedRefusal(t *testing.T) {
	program := callerCalleeProgram("s1:test:fn:does-not-exist")

	restore := corevalidate.SetDisableCalleeResolutionCheckForTest(true)
	defer restore()
	if result := corevalidate.Validate(program); !result.Valid {
		t.Fatalf("expected the seam to suppress the unresolved-callee refusal, got %+v", result)
	}

	restore()
	if result := corevalidate.Validate(program); result.Valid {
		t.Fatalf("expected the refusal restored once the seam is disengaged, got %+v", result)
	} else if len(result.Problems) == 0 || result.Problems[0].Code != core.CallCalleeUnresolved {
		t.Fatalf("expected %s restored, got %+v", core.CallCalleeUnresolved, result.Problems)
	}
}
