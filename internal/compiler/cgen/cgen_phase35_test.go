package cgen_test

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const phase35TracerSource = `module phase35.native_shift

export {
  fn main
}

fn main(input: U64) -> U64 {
  let shifted = 1 << 63
  shifted
}
`

const phase35ScalarCensusSource = `module phase35.scalar_census

export {
  fn main
}

fn main(input: U64) -> U64 {
  let x = input ^ 0
  let y = x & 18446744073709551615
  let z = y | 0
  let left = z << 63
  let result = left >> 63
  result
}
`

type phase35ShiftControl struct {
	Mutation string `json:"mutation"`
	Failure  string `json:"failure"`
}

type phase35ShiftExpected struct {
	Name        string                `json:"name"`
	SourcePlace string                `json:"source_place"`
	TargetPlace string                `json:"target_place"`
	Kind        string                `json:"kind"`
	Operator    string                `json:"operator"`
	Count       string                `json:"count"`
	SourceType  string                `json:"source_type"`
	ResultType  string                `json:"result_type"`
	Controls    []phase35ShiftControl `json:"controls"`
}

type phase35ShiftManifest struct {
	Format             string                 `json:"format"`
	Module             string                 `json:"module"`
	Function           string                 `json:"function"`
	SourceType         string                 `json:"source_type"`
	ResultType         string                 `json:"result_type"`
	ExpectedShiftSites int                    `json:"expected_shift_sites"`
	Sites              []phase35ShiftExpected `json:"sites"`
	RemovedSite        struct {
		Name    string `json:"name"`
		Failure string `json:"failure"`
	} `json:"removed_site"`
}

type phase35ShiftSite struct {
	Target   string
	Source   string
	Operator string
	Count    string
	Line     string
}

func phase35Check(t *testing.T, source string) core.Program {
	t.Helper()
	checked := session.Check([]byte(source))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	return checked.Program
}

func phase35CloneProgram(t *testing.T, program core.Program) core.Program {
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

func TestPhase35NativeShiftTracer(t *testing.T) {
	program := phase35Check(t, phase35TracerSource)
	generated, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("EmitNative: %v", err)
	}
	function := program.Functions[0]
	var shift core.LinearOperation
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpShiftLeftU64 {
			shift = operation
			break
		}
	}
	if shift.ID == "" || shift.ConstU64 != "63" || shift.RightID != "" {
		t.Fatalf("checked shift = %+v, want left shift by canonical literal 63", shift)
	}
	placeIndex := func(id string) int {
		for index, place := range function.Linear.Places {
			if place.ID == id {
				return index
			}
		}
		return -1
	}
	sourceLocal := fmt.Sprintf("schway_scalar_%d", placeIndex(shift.SourceID))
	targetLocal := fmt.Sprintf("schway_scalar_%d", placeIndex(shift.TargetID))
	wantSite := fmt.Sprintf("%s = (uint64_t)((unsigned long long)%s << 63);", targetLocal, sourceLocal)
	if !strings.Contains(generated, wantSite) {
		t.Fatalf("generated C lacks unsigned exact-width shift site %q", wantSite)
	}
	for _, unsafe := range []string{"(long long)" + sourceLocal + " << 63", sourceLocal + " << UINT64_C(63)"} {
		if strings.Contains(generated, unsafe) {
			t.Fatalf("generated C contains unsafe shift fragment %q", unsafe)
		}
	}

	got, err := native.DefaultRunner().Run(context.Background(), generated, "-O0", []string{"0"})
	if err != nil || len(got.Pairs) != 1 {
		t.Fatalf("native run: pairs=%d err=%v", len(got.Pairs), err)
	}
	if value := got.Pairs[0].Execution.Outcome.Value; value != "9223372036854775808" {
		t.Fatalf("native result = %q, want 9223372036854775808", value)
	}
	events := got.Pairs[0].Execution.Events
	if len(events) != 2 || events[0].Kind != "value.copied" || events[1].Kind != "function.returned" {
		t.Fatalf("shift emitted unexpected events: %+v", events)
	}
}

func TestPhase35EmitterAdmissionGates(t *testing.T) {
	program := phase35Check(t, phase35TracerSource)
	for index := range program.Functions {
		if program.Functions[index].Linear == nil {
			continue
		}
		for operationIndex := range program.Functions[index].Linear.Operations {
			operation := &program.Functions[index].Linear.Operations[operationIndex]
			if operation.Kind == core.OpShiftLeftU64 {
				operation.ConstU64 = "64"
				break
			}
		}
	}
	for _, route := range []struct {
		name string
		emit func(core.Program) (string, error)
	}{
		{name: "Emit", emit: cgen.Emit},
		{name: "EmitNative", emit: cgen.EmitNative},
		{name: "EmitApplication", emit: cgen.EmitApplication},
	} {
		t.Run(route.name, func(t *testing.T) {
			generated, err := route.emit(phase35CloneProgram(t, program))
			if err == nil {
				t.Fatal("accepted forged shift count 64")
			}
			if generated != "" {
				t.Fatalf("malformed program emitted C (%d bytes)", len(generated))
			}
			if !strings.Contains(err.Error(), "core validation failed:") {
				t.Fatalf("refusal did not come from the independent raw-core gate: %v", err)
			}
		})
	}
}

func phase35ReadShiftManifest(t *testing.T) phase35ShiftManifest {
	t.Helper()
	encoded, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase35", "shift-sites.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest phase35ShiftManifest
	if err := json.Unmarshal(encoded, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Format != "schway.phase35.shift-sites/v1" || manifest.ExpectedShiftSites <= 0 || len(manifest.Sites) != manifest.ExpectedShiftSites {
		t.Fatalf("invalid positive shift-site manifest: %+v", manifest)
	}
	return manifest
}

var phase35ShiftCLine = regexp.MustCompile(`^\s*(schway_scalar_[0-9]+)\s*=\s*\(uint64_t\)\(\(unsigned long long\)(schway_scalar_[0-9]+)\s*(<<|>>)\s*([^)]*)\);\s*$`)

func phase35ScalarRegions(generated string) []string {
	var regions []string
	for offset := 0; offset < len(generated); {
		next := strings.Index(generated[offset:], "static uint64_t ")
		if next < 0 {
			break
		}
		start := offset + next
		lineEnd := strings.IndexByte(generated[start:], '\n')
		if lineEnd < 0 {
			break
		}
		signature := generated[start : start+lineEnd]
		if !strings.Contains(signature, "unsigned int invocation_index)") || !strings.HasSuffix(strings.TrimSpace(signature), "{") {
			offset = start + lineEnd + 1
			continue
		}
		close := strings.Index(generated[start+lineEnd:], "\n}\n")
		if close < 0 {
			return regions
		}
		end := start + lineEnd + close + 3
		regions = append(regions, generated[start:end])
		offset = end
	}
	return regions
}

func phase35ScanShiftSites(generated string, expectedCount int) ([]phase35ShiftSite, error) {
	var sites []phase35ShiftSite
	for _, region := range phase35ScalarRegions(generated) {
		for _, line := range strings.Split(region, "\n") {
			operatorIndex := strings.Index(line, "<<")
			operator := "<<"
			if operatorIndex < 0 {
				operatorIndex = strings.Index(line, ">>")
				operator = ">>"
			}
			if operatorIndex < 0 {
				continue
			}
			match := phase35ShiftCLine.FindStringSubmatch(line)
			if match == nil {
				if !strings.Contains(line[:operatorIndex], "(unsigned long long)") || !strings.Contains(line[:operatorIndex], "= (uint64_t)") {
					return nil, fmt.Errorf("shift.site.unsigned_operand")
				}
				return nil, fmt.Errorf("shift.site.shape")
			}
			if match[3] != operator {
				return nil, fmt.Errorf("shift.site.shape")
			}
			countText := strings.TrimSpace(match[4])
			count, err := strconv.ParseUint(countText, 10, 64)
			if err != nil || strconv.FormatUint(count, 10) != countText {
				return nil, fmt.Errorf("shift.site.dynamic_count")
			}
			if count > 63 {
				return nil, fmt.Errorf("shift.site.count_out_of_range")
			}
			sites = append(sites, phase35ShiftSite{Target: match[1], Source: match[2], Operator: match[3], Count: countText, Line: line})
		}
	}
	if len(sites) != expectedCount {
		return nil, fmt.Errorf("shift.site.count_mismatch")
	}
	return sites, nil
}

func phase35MutateShiftSite(source string, site phase35ShiftSite, mutation string) (string, error) {
	mutatedLine := site.Line
	switch mutation {
	case "dynamic_count":
		mutatedLine = strings.Replace(mutatedLine, site.Operator+" "+site.Count, site.Operator+" runtimeCount", 1)
	case "count_64":
		mutatedLine = strings.Replace(mutatedLine, site.Operator+" "+site.Count, site.Operator+" 64", 1)
	case "signed_operand":
		mutatedLine = strings.Replace(mutatedLine, "(unsigned long long)", "(long long)", 1)
	case "remove":
		mutated := strings.Replace(source, site.Line+"\n", "", 1)
		if mutated == source {
			return source, fmt.Errorf("control did not remove the selected shift line")
		}
		return mutated, nil
	default:
		return source, fmt.Errorf("unknown shift mutation %q", mutation)
	}
	mutated := strings.Replace(source, site.Line, mutatedLine, 1)
	if mutated == source {
		return source, fmt.Errorf("control %q did not reach its selected shift line", mutation)
	}
	return mutated, nil
}

func phase35PlaceByName(function core.Function, name string) (core.Place, bool) {
	if function.Linear == nil {
		return core.Place{}, false
	}
	for _, place := range function.Linear.Places {
		if place.Name == name {
			return place, true
		}
	}
	return core.Place{}, false
}

func phase35ShiftOperationForExpected(function core.Function, expected phase35ShiftExpected) (core.LinearOperation, bool) {
	source, ok := phase35PlaceByName(function, expected.SourcePlace)
	if !ok || function.Linear == nil {
		return core.LinearOperation{}, false
	}
	for index, operation := range function.Linear.Operations {
		if string(operation.Kind) != expected.Kind || operation.SourceID != source.ID || operation.ConstU64 != expected.Count || operation.RightID != "" {
			continue
		}
		for _, copy := range function.Linear.Operations[index+1:] {
			if copy.Kind != core.OpCopy || copy.SourceID != operation.TargetID {
				continue
			}
			if target, ok := phase35PlaceByName(function, expected.TargetPlace); ok && copy.TargetID == target.ID {
				return operation, true
			}
		}
	}
	return core.LinearOperation{}, false
}

func phase35PlaceIndex(function core.Function, id string) int {
	for index, place := range function.Linear.Places {
		if place.ID == id {
			return index
		}
	}
	return -1
}

func TestPhase35UnsignedShiftEmissionAndEventProjection(t *testing.T) {
	manifest := phase35ReadShiftManifest(t)
	program := phase35Check(t, phase35ScalarCensusSource)
	if len(program.Functions) != 1 {
		t.Fatalf("scalar census functions=%d, want one", len(program.Functions))
	}
	function := program.Functions[0]
	if function.Name != manifest.Function || function.Parameter.Type != manifest.SourceType || function.ReturnType != manifest.ResultType || function.Linear == nil {
		t.Fatalf("source/manifest function facts disagree: function=%+v manifest=%+v", function, manifest)
	}
	typeByID := make(map[string]string, len(function.Linear.Types))
	for _, fact := range function.Linear.Types {
		typeByID[fact.ID] = fact.Shape.Constructor
	}
	shiftOperations := make([]core.LinearOperation, 0)
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpShiftLeftU64 || operation.Kind == core.OpShiftRightU64 {
			shiftOperations = append(shiftOperations, operation)
		}
	}
	if len(shiftOperations) != manifest.ExpectedShiftSites {
		t.Fatalf("IR shift operations=%d, want independently listed %d", len(shiftOperations), manifest.ExpectedShiftSites)
	}
	for _, expected := range manifest.Sites {
		sourcePlace, sourceOK := phase35PlaceByName(function, expected.SourcePlace)
		_, targetOK := phase35PlaceByName(function, expected.TargetPlace)
		if !sourceOK || !targetOK {
			t.Fatalf("manifest site %q has unresolved source/target places", expected.Name)
		}
		operation, found := phase35ShiftOperationForExpected(function, expected)
		if !found {
			t.Fatalf("manifest site %q does not match checked IR: expected kind=%s source=%s target=%s count=%s; IR=%+v", expected.Name, expected.Kind, sourcePlace.ID, expected.TargetPlace, expected.Count, shiftOperations)
		}
		if typeByID[sourcePlace.TypeID] != expected.SourceType || typeByID[operation.TypeID] != expected.ResultType {
			t.Fatalf("manifest site %q types source=%s target=%s, want %s -> %s", expected.Name, typeByID[sourcePlace.TypeID], typeByID[operation.TypeID], expected.SourceType, expected.ResultType)
		}
	}

	generated, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("EmitNative: %v", err)
	}
	if _, err := cgen.EmitApplication(program); err != nil {
		t.Fatalf("valid U64-to-U64 application entry was refused: %v", err)
	}
	sites, err := phase35ScanShiftSites(generated, manifest.ExpectedShiftSites)
	if err != nil {
		t.Fatalf("generated shift-site scan: %v", err)
	}
	for _, expected := range manifest.Sites {
		operation, found := phase35ShiftOperationForExpected(function, expected)
		if !found {
			t.Fatalf("manifest site %q lost its checked operation", expected.Name)
		}
		want := phase35ShiftSite{
			Source:   fmt.Sprintf("schway_scalar_%d", phase35PlaceIndex(function, operation.SourceID)),
			Target:   fmt.Sprintf("schway_scalar_%d", phase35PlaceIndex(function, operation.TargetID)),
			Operator: expected.Operator,
			Count:    expected.Count,
		}
		matched := false
		for _, site := range sites {
			if site.Source == want.Source && site.Target == want.Target && site.Operator == want.Operator && site.Count == want.Count {
				matched = true
			}
		}
		if !matched {
			t.Fatalf("generated C lacks independently expected site %+v; observed %+v", want, sites)
		}
	}

	capacityPattern := regexp.MustCompile(`(?m)^#define SCHWAY_EVENT_CAPACITY ([0-9]+)u$`)
	capacityMatch := capacityPattern.FindStringSubmatch(generated)
	if capacityMatch == nil {
		t.Fatal("generated C lacks event-capacity definition")
	}
	capacity, err := strconv.Atoi(capacityMatch[1])
	if err != nil || capacity != len(function.Linear.Operations) {
		t.Fatalf("event capacity=%q, want conservative operation count %d", capacityMatch[1], len(function.Linear.Operations))
	}

	got, err := native.DefaultRunner().Run(context.Background(), generated, "-O0", []string{"1"})
	if err != nil || len(got.Pairs) != 1 {
		t.Fatalf("native run: pairs=%d err=%v", len(got.Pairs), err)
	}
	if value := got.Pairs[0].Execution.Outcome.Value; value != "1" {
		t.Fatalf("five-op native result = %q, want 1", value)
	}
	var wantKinds, wantIDs []string
	for _, operation := range function.Linear.Operations {
		switch operation.Kind {
		case core.OpCopy:
			wantKinds = append(wantKinds, "value.copied")
			wantIDs = append(wantIDs, operation.ID+":event")
		case core.OpReturn:
			wantKinds = append(wantKinds, "function.returned")
			wantIDs = append(wantIDs, operation.ID+":event:returned")
		}
	}
	events := got.Pairs[0].Execution.Events
	if len(events) != len(wantKinds) {
		t.Fatalf("pure operation emitted or removed an event: got=%+v wantKinds=%v", events, wantKinds)
	}
	for index, event := range events {
		if event.Kind != wantKinds[index] || event.ID != wantIDs[index] {
			t.Fatalf("event %d = (%s,%s), want (%s,%s)", index, event.Kind, event.ID, wantKinds[index], wantIDs[index])
		}
	}
	outputBound, err := cgen.ExecutionOutputSizeForTest(program)
	if err != nil {
		t.Fatalf("ExecutionOutputSizeForTest: %v", err)
	}
	if got.OutputBytes > outputBound {
		t.Fatalf("native output bytes=%d exceed schema-size bound=%d", got.OutputBytes, outputBound)
	}
	t.Run("value-vectors", phase35UnsignedShiftValueVectors)
	t.Run("byte-signature", phase35ByteFixtureSignaturePreserved)
	t.Run("branch-dispatch", phase35BranchEmitterCases)
}

func TestPhase35ShiftSiteControls(t *testing.T) {
	manifest := phase35ReadShiftManifest(t)
	program := phase35Check(t, phase35ScalarCensusSource)
	generated, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("EmitNative: %v", err)
	}
	baseline, err := phase35ScanShiftSites(generated, manifest.ExpectedShiftSites)
	if err != nil {
		t.Fatalf("baseline shift-site count: %v", err)
	}
	for _, expected := range manifest.Sites {
		operation, found := phase35ShiftOperationForExpected(program.Functions[0], expected)
		if !found {
			t.Fatalf("manifest site %q lost its checked operation", expected.Name)
		}
		sourceLocal := fmt.Sprintf("schway_scalar_%d", phase35PlaceIndex(program.Functions[0], operation.SourceID))
		targetLocal := fmt.Sprintf("schway_scalar_%d", phase35PlaceIndex(program.Functions[0], operation.TargetID))
		var selected *phase35ShiftSite
		for index := range baseline {
			if baseline[index].Source == sourceLocal && baseline[index].Target == targetLocal && baseline[index].Operator == expected.Operator {
				selected = &baseline[index]
				break
			}
		}
		if selected == nil {
			t.Fatalf("manifest site %q was not reached by the scanner", expected.Name)
		}
		for _, control := range expected.Controls {
			control := control
			t.Run(expected.Name+"/"+control.Mutation, func(t *testing.T) {
				mutated, err := phase35MutateShiftSite(generated, *selected, control.Mutation)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := phase35ScanShiftSites(mutated, manifest.ExpectedShiftSites); err == nil || err.Error() != control.Failure {
					t.Fatalf("mutant failure=%v, want exact %q", err, control.Failure)
				}
			})
		}
	}
	var removed *phase35ShiftSite
	for index := range baseline {
		if strings.Contains(baseline[index].Line, "schway_scalar_") {
			for _, expected := range manifest.Sites {
				if expected.Name == manifest.RemovedSite.Name && strings.Contains(baseline[index].Line, baseline[index].Operator+" "+expected.Count) {
					removed = &baseline[index]
					break
				}
			}
		}
	}
	if removed == nil {
		t.Fatalf("removed-site control %q did not resolve", manifest.RemovedSite.Name)
	}
	mutated, err := phase35MutateShiftSite(generated, *removed, "remove")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := phase35ScanShiftSites(mutated, manifest.ExpectedShiftSites); err == nil || err.Error() != manifest.RemovedSite.Failure {
		t.Fatalf("removed-site failure=%v, want exact %q", err, manifest.RemovedSite.Failure)
	}
	if _, err := phase35ScanShiftSites(generated, manifest.ExpectedShiftSites); err != nil {
		t.Fatalf("restored shift-site scan: %v", err)
	}
}

func phase35UnsignedShiftValueVectors(t *testing.T) {
	for _, vector := range []struct {
		direction string
		count     string
		input     string
		want      string
	}{
		{direction: "left", count: "0", input: "9223372036854775808", want: "9223372036854775808"},
		{direction: "left", count: "31", input: "1", want: "2147483648"},
		{direction: "left", count: "32", input: "1", want: "4294967296"},
		{direction: "left", count: "63", input: "1", want: "9223372036854775808"},
		{direction: "left", count: "1", input: "9223372036854775808", want: "0"},
		{direction: "right", count: "0", input: "9223372036854775808", want: "9223372036854775808"},
		{direction: "right", count: "31", input: "9223372036854775808", want: "4294967296"},
		{direction: "right", count: "32", input: "9223372036854775808", want: "2147483648"},
		{direction: "right", count: "63", input: "9223372036854775808", want: "1"},
	} {
		vector := vector
		t.Run(vector.direction+"/"+vector.count+"/"+vector.input, func(t *testing.T) {
			op := "<<"
			if vector.direction == "right" {
				op = ">>"
			}
			source := fmt.Sprintf("module phase35.vector_%s_%s\nexport { fn main }\nfn main(input: U64) -> U64 {\n  let shifted = input %s %s\n  shifted\n}\n", vector.direction, vector.count, op, vector.count)
			program := phase35Check(t, source)
			generated, err := cgen.EmitNative(program)
			if err != nil {
				t.Fatalf("EmitNative: %v", err)
			}
			got, err := native.DefaultRunner().Run(context.Background(), generated, "-O0", []string{vector.input})
			if err != nil || len(got.Pairs) != 1 {
				t.Fatalf("native shift-vector run: pairs=%d err=%v", len(got.Pairs), err)
			}
			if value := got.Pairs[0].Execution.Outcome.Value; value != vector.want {
				t.Fatalf("%s %s of %s = %q, want %s", vector.direction, vector.count, vector.input, value, vector.want)
			}
		})
	}
}

func phase35ByteFixtureSignaturePreserved(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "bitwise.schway"))
	if err != nil {
		t.Fatal(err)
	}
	program := phase35Check(t, string(source))
	if len(program.Functions) != 1 || program.Functions[0].Parameter.Type != "Byte" || program.Functions[0].ReturnType != "U64" {
		t.Fatalf("bitwise conformance entry lost its Byte-to-U64 signature: %+v", program.Functions)
	}
	generated, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("EmitNative Byte fixture: %v", err)
	}
	if !strings.Contains(generated, "static uint64_t ") || !strings.Contains(generated, "(unsigned char ") {
		t.Fatal("generated conformance C does not retain a Byte parameter and U64 result")
	}
	if _, err := cgen.EmitApplication(program); err == nil {
		t.Fatal("Byte fixture was accepted by the U64 application shell")
	}
	boolSource := `module phase35.bool_local_signature

export {
  fn main
}

fn main(input: U64) -> U64 {
  let shifted = input << 63
  let is_one = shifted == 1
  shifted
}
`
	boolProgram := phase35Check(t, boolSource)
	boolGenerated, err := cgen.EmitNative(boolProgram)
	if err != nil {
		t.Fatalf("EmitNative U64/Bool local signature: %v", err)
	}
	if !strings.Contains(boolGenerated, "int schway_scalar_") || !strings.Contains(boolGenerated, "uint64_t schway_scalar_") {
		t.Fatal("scalar C did not preserve both Bool and U64 local storage")
	}
}

func phase35BranchEmitterCases(t *testing.T) {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), testsupport.ProjectPath("internal", "compiler", "cgen", "cgen_program.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"OpBitXorU64": true, "OpBitAndU64": true, "OpBitOrU64": true,
		"OpShiftLeftU64": true, "OpShiftRightU64": true,
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name.Name != "emitProgramBranchOperation" {
			continue
		}
		ast.Inspect(function.Body, func(node ast.Node) bool {
			clause, ok := node.(*ast.CaseClause)
			if !ok {
				return true
			}
			for _, expression := range clause.List {
				selector, ok := expression.(*ast.SelectorExpr)
				if !ok || selector.Sel == nil {
					continue
				}
				if base, ok := selector.X.(*ast.Ident); ok && base.Name == "core" {
					delete(want, selector.Sel.Name)
				}
			}
			return true
		})
		if len(want) != 0 {
			t.Fatalf("branch operation switch omits Phase 35 kinds: %v", want)
		}
		return
	}
	t.Fatal("emitProgramBranchOperation dispatcher is absent")
}
