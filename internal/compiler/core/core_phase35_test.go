package core_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
)

func TestPhase35CoreShape(t *testing.T) {
	wantKinds := []core.OperationKind{
		core.OpBitXorU64,
		core.OpBitAndU64,
		core.OpBitOrU64,
		core.OpShiftLeftU64,
		core.OpShiftRightU64,
	}
	registered := make(map[core.OperationKind]bool)
	for _, kind := range core.AllOperationKinds() {
		registered[kind] = true
	}
	for _, kind := range wantKinds {
		if !registered[kind] {
			t.Errorf("Phase 35 operation %q is absent from AllOperationKinds", kind)
		}
	}
	if got, want := len(core.AllOperationKinds()), 29; got != want {
		t.Fatalf("operation registry has %d entries, want %d", got, want)
	}
	if core.Schema != "lang.core/0" || core.Schema1 != "lang.core/1" {
		t.Fatalf("Phase 35 changed the existing core schema IDs: %q / %q", core.Schema, core.Schema1)
	}

	operationType := reflect.TypeOf(core.LinearOperation{})
	field, ok := operationType.FieldByName("ConstU64")
	if !ok || field.Type.Kind() != reflect.String || field.Tag.Get("json") != "const_u64,omitempty" {
		t.Fatalf("ConstU64 field = %+v, found=%v; immediate shifts must reuse its existing optional string", field, ok)
	}
	if got, want := operationType.NumField(), 25; got != want {
		t.Fatalf("Phase 35 changed LinearOperation field count to %d, want existing %d", got, want)
	}

	legacy := core.LinearOperation{
		ID:       "f:op:0",
		PointID:  "f:point:linear:0",
		Kind:     core.OpConst,
		TargetID: "f:place:1",
		TypeID:   "f:type:0",
		ConstU64: "42",
	}
	encoded, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"id":"f:op:0","point_id":"f:point:linear:0","kind":"const","source_id":"","target_id":"f:place:1","type_id":"f:type:0","const_u64":"42"}`
	if string(encoded) != want {
		t.Fatalf("legacy OpConst serialization changed: got %s, want %s", encoded, want)
	}
}
