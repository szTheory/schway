// This file proves OWN-05's CORE-LAYER claim (D-09-34, D-09-35), the
// separate, weaker-trust-boundary counterpart to
// internal/compiler/syntax/syntax_convention_override_test.go's
// SOURCE-layer claim. `corevalidate`'s whole role is validating a
// core.Program it did not produce -- a hostile or corrupted producer never
// crosses the parser at all, so "not expressible in source" says nothing
// about what a directly-hand-built (or maliciously hand-edited) core
// artifact could claim. This file proves the core-layer claim on its own
// terms: core.LinearOperation has no field a per-call-site convention
// override could occupy, and the one field a hostile producer COULD use to
// claim a non-"owned" parameter convention (ParameterContract.Mode /
// ReturnContract.Mode) is already validated against its closed three-value
// set at DecodeInterface time.
//
// D-09-35: no new seam is minted here. A seam-backed refusal for a field
// the wire format does not have would be dead weight invented to fill a
// shape that does not exist -- the opposite of D-07-47
// (check.call_return_type_unrepresentable) and D-08-15 (the loan-liveness
// bound), where a real, reachable-looking code path existed and had to be
// proven closed. Here there is no such path: the fields below are the
// complete, exhaustive set, and the one field with wire-level claim
// surface (Mode) is confirmed fail-closed by
// TestParameterModeDecodeIsClosedSet using the EXISTING decode control at
// core.go's DecodeInterface, never a newly-added one.
package core_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
)

// linearOperationExpectedFields is core.LinearOperation's complete,
// explicitly-listed field-name set (core/core.go:657-707, read in full at
// planning time). Listing it explicitly -- rather than only scanning for
// forbidden substrings -- is what makes this a tripwire rather than a
// heuristic: any future additive field, whatever its name, changes this
// set and must be justified in THIS test's own diff, not merely pass
// through a substring filter that happens not to match it.
var linearOperationExpectedFields = []string{
	"ID", "PointID", "Kind", "SourceID", "TargetID", "LoanID", "TypeID",
	"OkEdgeID", "ErrEdgeID", "ErrTargetID", "ReleasesOperationID",
	"Allocator", "Reason", "CalleeID",
}

// TestConventionOverrideNotExpressibleInCore is D-09-34/D-09-35's
// core-layer structural-absence proof: core.LinearOperation's complete
// field set equals the explicitly-listed expected set above, no field name
// contains "mode", "convention", "ownership", or "transfer"
// (case-insensitive), and an artifact carrying an unknown extra key on a
// linear operation cannot survive a JSON round-trip into a populated
// override -- there is no field for it to land in.
func TestConventionOverrideNotExpressibleInCore(t *testing.T) {
	operationType := reflect.TypeOf(core.LinearOperation{})

	got := make([]string, 0, operationType.NumField())
	for i := 0; i < operationType.NumField(); i++ {
		got = append(got, operationType.Field(i).Name)
	}
	if len(got) != len(linearOperationExpectedFields) {
		t.Fatalf("core.LinearOperation field count changed: got %v (%d fields), want %v (%d fields) -- a field was added or removed; update this test's expected set deliberately", got, len(got), linearOperationExpectedFields, len(linearOperationExpectedFields))
	}
	for i, name := range linearOperationExpectedFields {
		if got[i] != name {
			t.Fatalf("core.LinearOperation field %d = %q, want %q -- the field set or its order changed", i, got[i], name)
		}
	}

	forbidden := []string{"mode", "convention", "ownership", "transfer"}
	for _, name := range linearOperationExpectedFields {
		lowered := strings.ToLower(name)
		for _, term := range forbidden {
			if strings.Contains(lowered, term) {
				t.Fatalf("core.LinearOperation.%s's name contains %q -- a per-call-site convention override slot now exists; D-09-34's core-layer structural-absence claim is broken", name, term)
			}
		}
	}

	// An unknown extra key on a linear operation must not survive decode
	// into any populated field -- there is no slot for it to land in.
	raw := `{
		"id": "s1:probe:fn:x:op:0", "point_id": "s1:probe:fn:x:point:linear:0",
		"kind": "move", "source_id": "s1:probe:fn:x:place:0",
		"target_id": "s1:probe:fn:x:place:1", "type_id": "s1:probe:fn:x:type:0",
		"convention_override": "exclusive"
	}`
	var decoded core.LinearOperation
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}
	roundTripped, err := json.Marshal(decoded)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}
	if strings.Contains(string(roundTripped), "convention_override") || strings.Contains(string(roundTripped), "exclusive") {
		t.Fatalf("an unknown convention_override key survived the JSON round-trip: %s -- core.LinearOperation now has a slot for it", roundTripped)
	}
}

// parameterModeHostileTable and parameterModeLegalTable are
// TestParameterModeDecodeIsClosedSet's shared fixture halves: five hostile
// mode values that must each be refused with core.interface_invalid_mode,
// and the three legal values that must each be admitted -- proving the
// control by BOTH directions, not merely by over-refusal.
var parameterModeHostileTable = []string{
	"",             // empty-after-present: handled separately below (missing_field, not invalid_mode)
	"Owned",        // capitalized
	"borrowed",     // plausible-looking but not the declared vocabulary
	"caller-owned", // an override-shaped string, exactly the claim D-09-35 worries about
	" owned",       // whitespace-padded
}

var parameterModeLegalTable = []string{"owned", "shared", "exclusive"}

// buildInterfaceV1Document returns a minimal, otherwise-valid
// lang.interface/1 document as a mutable map, so each subtest can set
// exactly one parameter Mode value and observe DecodeInterface's response
// to it in isolation. Mirrors this package's own
// validInterfaceV1Document/deepCopyJSON precedent (core_test.go) without
// depending on that unexported helper directly (different file, same
// package, but keeping this file's own fixture self-contained avoids a
// cross-file coupling on a helper's exact shape).
func buildInterfaceV1Document(mode string) map[string]any {
	return map[string]any{
		"schema":      "lang.interface/1",
		"module_id":   "m1",
		"core_digest": validHexDigestForConventionAbsence,
		"functions": []any{
			map[string]any{
				"id":   "f1",
				"name": "identity",
				"parameters": []any{
					map[string]any{"id": "p1", "name": "buffer", "type": "Buffer", "mode": mode, "drops": false},
				},
				"return":         map[string]any{"type": "Buffer", "mode": "owned", "paths": []any{}, "fresh": false},
				"abilities":      []any{},
				"callable":       false,
				"foreign":        map[string]any{"allocator": "", "unwind": "", "nonlocal_exit": ""},
				"closure_digest": validHexDigestForConventionAbsence,
			},
		},
	}
}

func buildReturnModeInterfaceV1Document(mode string, paths []any) map[string]any {
	document := buildInterfaceV1Document("owned")
	function := document["functions"].([]any)[0].(map[string]any)
	function["return"] = map[string]any{"type": "Buffer", "mode": mode, "paths": paths, "fresh": false}
	return document
}

// validHexDigestForConventionAbsence mirrors core_test.go's own
// validHexDigest shape (a syntactically valid sha256:+64-lowercase-hex
// digest), declared separately in this file to avoid depending on that
// other file's unexported constant directly.
const validHexDigestForConventionAbsence = "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcd" + "ef"

// TestParameterModeDecodeIsClosedSet confirms and extends coverage of the
// EXISTING closed-set decode control at core.go's decodeInterfaceV1 (the
// `core.interface_invalid_mode` checks against Parameters[i].Mode and
// Return.Mode): the one core-level field a hostile or corrupted producer
// could use to claim a non-"owned" convention is already validated against
// its closed three-value set at decode time, so no new seam is needed
// (D-09-35). Both directions are driven: five hostile values must each be
// refused, and all three legal values must each be admitted, so the
// control is not proven merely by over-refusal.
func TestParameterModeDecodeIsClosedSet(t *testing.T) {
	t.Run("Parameters[0].Mode hostile values refused", func(t *testing.T) {
		for _, mode := range parameterModeHostileTable {
			t.Run("mode="+mode, func(t *testing.T) {
				document := buildInterfaceV1Document(mode)
				data, err := json.Marshal(document)
				if err != nil {
					t.Fatalf("marshal: %v", err)
				}
				_, err = core.DecodeInterface(data)
				if err == nil {
					t.Fatalf("expected refusal for Parameters[0].Mode = %q, got a clean decode", mode)
				}
				var decodeErr *core.DecodeError
				if de, ok := err.(*core.DecodeError); ok {
					decodeErr = de
				}
				if decodeErr == nil {
					t.Fatalf("expected a *core.DecodeError, got %v (%T)", err, err)
				}
				if mode == "" {
					if decodeErr.Code != "core.interface_missing_field" {
						t.Fatalf("expected core.interface_missing_field for an empty Mode, got %q", decodeErr.Code)
					}
					return
				}
				if decodeErr.Code != "core.interface_invalid_mode" {
					t.Fatalf("expected core.interface_invalid_mode for Mode = %q, got %q", mode, decodeErr.Code)
				}
			})
		}
	})

	t.Run("Parameters[0].Mode legal values admitted", func(t *testing.T) {
		for _, mode := range parameterModeLegalTable {
			t.Run("mode="+mode, func(t *testing.T) {
				document := buildInterfaceV1Document(mode)
				data, err := json.Marshal(document)
				if err != nil {
					t.Fatalf("marshal: %v", err)
				}
				decoded, err := core.DecodeInterface(data)
				if err != nil {
					t.Fatalf("expected Parameters[0].Mode = %q to decode cleanly, got %v", mode, err)
				}
				if decoded.V1 == nil || !decoded.Admissible {
					t.Fatalf("expected Parameters[0].Mode = %q to be Admissible: %+v", mode, decoded)
				}
			})
		}
	})

	t.Run("Return.Mode hostile values refused", func(t *testing.T) {
		for _, mode := range []string{"Owned", "borrowed", "caller-owned", " owned"} {
			t.Run("mode="+mode, func(t *testing.T) {
				document := buildReturnModeInterfaceV1Document(mode, []any{"buffer"})
				data, err := json.Marshal(document)
				if err != nil {
					t.Fatalf("marshal: %v", err)
				}
				_, err = core.DecodeInterface(data)
				var decodeErr *core.DecodeError
				if de, ok := err.(*core.DecodeError); ok {
					decodeErr = de
				}
				if decodeErr == nil || decodeErr.Code != "core.interface_invalid_mode" {
					t.Fatalf("expected core.interface_invalid_mode for Return.Mode = %q, got %v", mode, err)
				}
			})
		}
	})

	t.Run("Return.Mode legal values admitted", func(t *testing.T) {
		for _, mode := range parameterModeLegalTable {
			t.Run("mode="+mode, func(t *testing.T) {
				paths := []any{}
				if mode != "owned" {
					paths = []any{"buffer"}
				}
				document := buildReturnModeInterfaceV1Document(mode, paths)
				data, err := json.Marshal(document)
				if err != nil {
					t.Fatalf("marshal: %v", err)
				}
				decoded, err := core.DecodeInterface(data)
				if err != nil {
					t.Fatalf("expected Return.Mode = %q to decode cleanly, got %v", mode, err)
				}
				if decoded.V1 == nil || !decoded.Admissible {
					t.Fatalf("expected Return.Mode = %q to be Admissible: %+v", mode, decoded)
				}
			})
		}
	})
}
