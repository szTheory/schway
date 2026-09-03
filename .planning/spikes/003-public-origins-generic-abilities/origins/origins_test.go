package origins

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"
)

func TestFixtureCasesAreBodyBlindAndAgree(t *testing.T) {
	file := mustFixtures(t)
	public, diagnostics := ExportInterface(file, ProducerOptions{})
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected producer diagnostics: %+v", diagnostics)
	}
	public, err := RoundTripInterface(public)
	if err != nil {
		t.Fatal(err)
	}
	implementations, err := FunctionMap(file.Functions)
	if err != nil {
		t.Fatal(err)
	}
	for _, call := range file.Cases {
		call := call
		t.Run(call.ID, func(t *testing.T) {
			comparison := Compare(public, implementations, call)
			if !comparison.Agreement {
				data, _ := json.Marshal(comparison)
				t.Fatalf("consumer/oracle disagreement: %s", data)
			}
			if comparison.Consumer.Valid != call.ExpectValid {
				t.Fatalf("valid=%v, want %v: %+v", comparison.Consumer.Valid, call.ExpectValid, comparison.Consumer.Diagnostic)
			}
			if diagnosticCode(comparison.Consumer) != call.ExpectCode {
				t.Fatalf("code=%q, want %q", diagnosticCode(comparison.Consumer), call.ExpectCode)
			}
		})
	}
}

func TestDishonestProducerSummariesAreRejected(t *testing.T) {
	file := mustFixtures(t)
	want := map[string]string{
		"dishonest_union":    "producer.summary_omits_origin",
		"dishonest_owned":    "producer.ownership_mismatch",
		"dishonest_callback": "producer.scoped_callback_escape",
		"dishonest_access":   "producer.access_mismatch",
	}
	for _, function := range file.DishonestFunctions {
		diagnostic := VerifyProducer(publicFunction(function), function.Body)
		if diagnostic == nil || diagnostic.Code != want[function.ID] {
			t.Fatalf("%s diagnostic=%+v, want %s", function.ID, diagnostic, want[function.ID])
		}
	}
}

func TestInterfaceSerializationCannotLeakBodies(t *testing.T) {
	file := mustFixtures(t)
	public, _ := ExportInterface(file, ProducerOptions{})
	data, err := json.Marshal(public)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "\"body\"") {
		t.Fatalf("interface leaked body: %s", data)
	}
	roundTrip, err := RoundTripInterface(public)
	if err != nil {
		t.Fatal(err)
	}
	if !InterfacesEqual(public, roundTrip) {
		t.Fatal("interface changed across serialization")
	}
}

func TestAbilitiesPropagateIndependentlyThroughGenerics(t *testing.T) {
	file := mustFixtures(t)
	registry, err := NewRegistry(file.Types)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name     string
		typeExpr TypeExpr
		want     []Ability
	}{
		{"byte", TypeExpr{Name: "Byte"}, []Ability{Copy, Drop, Share, Send, Escape}},
		{"file", TypeExpr{Name: "File"}, []Ability{Drop, Send, Escape}},
		{"optional-file", TypeExpr{Name: "Option", Args: []TypeExpr{{Name: "File"}}}, []Ability{Drop, Send, Escape}},
		{"view-byte", TypeExpr{Name: "View", Args: []TypeExpr{{Name: "Byte"}}}, []Ability{Copy, Drop, Share, Send}},
		{"optional-view", TypeExpr{Name: "Option", Args: []TypeExpr{{Name: "View", Args: []TypeExpr{{Name: "Byte"}}}}}, []Ability{Copy, Drop, Share, Send}},
		{"guard", TypeExpr{Name: "Guard"}, []Ability{Drop}},
		{"unique-box-file", TypeExpr{Name: "UniqueBox", Args: []TypeExpr{{Name: "File"}}}, []Ability{Drop, Send, Escape}},
	}
	for _, test := range cases {
		got, err := registry.Abilities(test.typeExpr, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Fatalf("%s abilities=%v, want %v", test.name, got, test.want)
		}
	}
}

func TestAllAbilityCombinationsPropagateThroughGenericWrappers(t *testing.T) {
	for mask := 0; mask < 1<<len(AbilityOrder); mask++ {
		var base []Ability
		for index, ability := range AbilityOrder {
			if mask&(1<<index) != 0 {
				base = append(base, ability)
			}
		}
		specs := []TypeSpec{
			{Name: "Atom", Base: base},
			{Name: "Option", Parameters: 1, Conditional: map[Ability][]int{Copy: {0}, Drop: {0}, Share: {0}, Send: {0}, Escape: {0}}},
			{Name: "Result", Parameters: 2, Conditional: map[Ability][]int{Copy: {0, 1}, Drop: {0, 1}, Share: {0, 1}, Send: {0, 1}, Escape: {0, 1}}},
		}
		registry, err := NewRegistry(specs)
		if err != nil {
			t.Fatal(err)
		}
		option := TypeExpr{Name: "Option", Args: []TypeExpr{{Name: "Atom"}}}
		result := TypeExpr{Name: "Result", Args: []TypeExpr{{Name: "Atom"}, {Name: "Atom"}}}
		lookup := map[string]TypeSpec{"Atom": specs[0], "Option": specs[1], "Result": specs[2]}
		for _, ability := range AbilityOrder {
			want := containsAbility(base, ability)
			for _, expr := range []TypeExpr{option, result} {
				got, err := registry.Has(expr, ability, nil)
				if err != nil {
					t.Fatal(err)
				}
				if got != want || oracleHas(lookup, expr, ability, nil) != want {
					t.Fatalf("mask=%d type=%s ability=%s got=%v oracle=%v want=%v", mask, expr.Name, ability, got, oracleHas(lookup, expr, ability, nil), want)
				}
			}
		}
	}
}

func TestTaggedOriginAlternativesPreservePrecision(t *testing.T) {
	file := mustFixtures(t)
	public, _ := ExportInterface(file, ProducerOptions{})
	implementations, _ := FunctionMap(file.Functions)
	union := caseByID(t, file, "CALL-CHOOSE-001")
	tagged := caseByID(t, file, "CALL-TAGGED-001")
	if Compare(public, implementations, union).Consumer.Valid {
		t.Fatal("unrefined origin union should retain both possible sources")
	}
	if !Compare(public, implementations, tagged).Consumer.Valid {
		t.Fatal("tagged alternative should retain only its selected origin")
	}
}

func TestGeneratedConsumerAndBodyOracleAgree(t *testing.T) {
	file := mustFixtures(t)
	public, _ := ExportInterface(file, ProducerOptions{})
	implementations, _ := FunctionMap(file.Functions)
	cases := GenerateCases(file.Functions)
	if len(cases) < 50 {
		t.Fatalf("generated only %d cases", len(cases))
	}
	for index, call := range cases {
		comparison := Compare(public, implementations, call)
		if !comparison.Agreement {
			data, _ := json.Marshal(comparison)
			t.Fatalf("mismatch at %d: %s", index, data)
		}
	}
}

func TestInjectedOriginOmissionFindsSmallCounterexample(t *testing.T) {
	file := mustFixtures(t)
	before, _ := json.Marshal(file.Functions)
	search := FindInjectedMismatch(file)
	if search.Mismatch == nil || search.Input == nil {
		t.Fatal("injected origin omission produced no mismatch")
	}
	if search.Programs > 20 {
		t.Fatalf("counterexample required %d cases", search.Programs)
	}
	if search.Mismatch.Consumer.Valid == search.Mismatch.Oracle.Valid {
		t.Fatalf("fault was not semantically observable: %+v", search.Mismatch)
	}
	after, _ := json.Marshal(file.Functions)
	if string(before) != string(after) {
		t.Fatal("fault injection mutated implementation facts through shared backing storage")
	}
}

func TestEncodingComparisonKeepsValueOriginsCompact(t *testing.T) {
	file := mustFixtures(t)
	_, values := RenderAll(file.Functions, ValueOrigins)
	_, regions := RenderAll(file.Functions, ExplicitRegions)
	_, callbacks := RenderAll(file.Functions, CallbackOnly)
	if values.Characters >= regions.Characters || values.ExplicitBinders >= regions.ExplicitBinders {
		t.Fatalf("value origins did not reduce ceremony: values=%+v regions=%+v", values, regions)
	}
	if callbacks.AdaptedFunctions == 0 || callbacks.DirectFunctions >= values.DirectFunctions {
		t.Fatalf("callback-only control failed to expose API-shape adaptation: %+v", callbacks)
	}
}

func TestSeededRepresentationProperties(t *testing.T) {
	file := mustFixtures(t)
	public, _ := ExportInterface(file, ProducerOptions{})
	implementations, _ := FunctionMap(file.Functions)
	property := func(raw uint16, reverseFunctions, reverseOrigins, renameBindings bool) bool {
		candidate := cloneInterface(public)
		if reverseFunctions {
			reverseFunctionsInPlace(candidate.Functions)
		}
		if reverseOrigins {
			for functionIndex := range candidate.Functions {
				for returnIndex := range candidate.Functions[functionIndex].Returns {
					reverseStrings(candidate.Functions[functionIndex].Returns[returnIndex].Origins)
				}
			}
		}
		call := file.Cases[int(raw)%len(file.Cases)]
		if renameBindings {
			call = alphaRenameCall(call, "renamed_")
		}
		left := Compare(public, implementations, alphaRenameCall(file.Cases[int(raw)%len(file.Cases)], map[bool]string{true: "renamed_", false: ""}[renameBindings]))
		right := Compare(candidate, implementations, call)
		return left.Consumer.Valid == right.Consumer.Valid && diagnosticCode(left.Consumer) == diagnosticCode(right.Consumer)
	}
	config := &quick.Config{MaxCount: 500, Rand: rand.New(rand.NewSource(0x0A161A))}
	if err := quick.Check(property, config); err != nil {
		t.Fatal(err)
	}
}

func TestStableCanonicalInterfaceJSON(t *testing.T) {
	file := mustFixtures(t)
	first, _ := ExportInterface(file, ProducerOptions{})
	second, _ := ExportInterface(file, ProducerOptions{})
	left, _ := json.Marshal(first)
	right, _ := json.Marshal(second)
	if string(left) != string(right) {
		t.Fatalf("interface JSON changed:\n%s\n%s", left, right)
	}
}

func TestLargeOriginSummaryRemainsLinear(t *testing.T) {
	file := mustFixtures(t)
	large := LargeUnionFunction(1000)
	largeFile := FixtureFile{Schema: 1, Types: file.Types, Functions: []Function{large}}
	public, diagnostics := ExportInterface(largeFile, ProducerOptions{})
	if len(diagnostics) != 0 || len(public.Functions) != 1 {
		t.Fatalf("large summary failed: %+v", diagnostics)
	}
	call := generatedBase(large, "")
	call.Steps = []Step{{Kind: "use_result"}}
	if result := ConsumerCheck(public, call); !result.Valid {
		t.Fatalf("large consumer check failed: %+v", result.Diagnostic)
	}
	_, values := RenderAll([]Function{large}, ValueOrigins)
	_, regions := RenderAll([]Function{large}, ExplicitRegions)
	if values.OriginReferences != 1000 || regions.ExplicitBinders != 1001 {
		t.Fatalf("unexpected scale metrics: values=%+v regions=%+v", values, regions)
	}
}

func TestMalformedInterfacesProduceStableDiagnostics(t *testing.T) {
	file := mustFixtures(t)
	public, _ := ExportInterface(file, ProducerOptions{})
	call := caseByID(t, file, "CALL-HEAD-001")
	public.Functions[1].Returns[0].Origins = []string{"missing"}
	result := ConsumerCheck(public, call)
	if result.Valid || diagnosticCode(result) != "interface.unknown_origin" {
		t.Fatalf("unexpected result: %+v", result)
	}
	_, err := NewRegistry(append(file.Types, TypeSpec{Name: "Broken", Parameters: 1, Conditional: map[Ability][]int{Copy: {2}}}))
	if err == nil || !strings.Contains(err.Error(), "parameter 2") {
		t.Fatalf("expected invalid type rule diagnostic, got %v", err)
	}
}

func mustFixtures(t *testing.T) FixtureFile {
	t.Helper()
	file, err := LoadFixtures("../fixtures/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	return file
}

func caseByID(t *testing.T, file FixtureFile, id string) CallCase {
	t.Helper()
	for _, call := range file.Cases {
		if call.ID == id {
			return call
		}
	}
	t.Fatalf("case %s not found", id)
	return CallCase{}
}

func cloneInterface(input Interface) Interface {
	data, _ := json.Marshal(input)
	var result Interface
	_ = json.Unmarshal(data, &result)
	return result
}

func reverseFunctionsInPlace(values []Function) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}

func reverseStrings(values []string) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}

func alphaRenameCall(call CallCase, prefix string) CallCase {
	if prefix == "" {
		return call
	}
	arguments := make(map[string]string, len(call.Arguments))
	for name, binding := range call.Arguments {
		arguments[name] = binding
	}
	call.Arguments = arguments
	call.Bindings = append([]Binding(nil), call.Bindings...)
	call.Steps = append([]Step(nil), call.Steps...)
	for index := range call.Bindings {
		old := call.Bindings[index].ID
		call.Bindings[index].ID = prefix + old
		for name, binding := range call.Arguments {
			if binding == old {
				call.Arguments[name] = prefix + old
			}
		}
		for stepIndex := range call.Steps {
			if call.Steps[stepIndex].Binding == old {
				call.Steps[stepIndex].Binding = prefix + old
			}
		}
	}
	return call
}
