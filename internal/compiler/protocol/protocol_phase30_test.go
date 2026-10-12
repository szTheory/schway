package protocol_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/protocol"
)

func phase30UnitResult() protocol.Result {
	result := protocol.New("check", protocol.StatusInvalid)
	entry := diagnostic.Error("type.mismatch", diagnostic.Span{Start: 4, End: 9}, "entry problem")
	imported := diagnostic.Error("type.mismatch", diagnostic.Span{Start: 4, End: 9}, "imported problem").InUnit("adler32")
	result.Diagnostics = []diagnostic.Diagnostic{entry, imported}
	return result
}

// TestPhase30UnitProjection pins D-23: the human line gains a unit= token only
// when the diagnostic carries a unit, the empty-unit line is the exact
// pre-phase format, and JSON carries unit only when set.
func TestPhase30UnitProjection(t *testing.T) {
	result := phase30UnitResult()
	entry, imported := result.Diagnostics[0], result.Diagnostics[1]
	human, err := protocol.Human(result)
	if err != nil {
		t.Fatal(err)
	}
	wantEntry := entry.ID + " type.mismatch [4:9]: entry problem\n"
	if !strings.Contains(human, wantEntry) {
		t.Fatalf("empty-unit line is not the pre-change format %q in:\n%s", wantEntry, human)
	}
	wantImported := imported.ID + " type.mismatch unit=adler32 [4:9]: imported problem\n"
	if !strings.Contains(human, wantImported) || !strings.Contains(human, " unit=adler32 [") {
		t.Fatalf("unit-bearing line missing %q in:\n%s", wantImported, human)
	}
	if strings.Count(human, "unit=") != 1 {
		t.Fatalf("expected exactly one unit= token:\n%s", human)
	}
	if entry.ID == imported.ID {
		t.Fatalf("same code and span in two units share ID %s", entry.ID)
	}

	encoded, err := json.Marshal(result.Finalize())
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(string(encoded), `"unit":"adler32"`); got != 1 {
		t.Fatalf(`"unit":"adler32" appears %d times, want 1: %s`, got, encoded)
	}
	if got := strings.Count(string(encoded), `"unit"`); got != 1 {
		t.Fatalf(`"unit" key appears %d times, want 1 (empty unit must be omitted): %s`, got, encoded)
	}

	t.Run("control: always printing unit changes the empty-unit line", func(t *testing.T) {
		restore := protocol.SetAlwaysPrintUnitForTest(true)
		defer restore()
		armed, err := protocol.Human(phase30UnitResult())
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(armed, wantEntry) || strings.Count(armed, "unit=") != 2 {
			t.Fatalf("control did not observe the difference:\n%s", armed)
		}
	})

	t.Run("debug map entry unit is omitempty", func(t *testing.T) {
		plain, _ := json.Marshal(protocol.DebugMapEntry{ID: "e", Availability: "captured"})
		if strings.Contains(string(plain), "unit") {
			t.Fatalf("empty unit leaked into debug-map JSON: %s", plain)
		}
		withUnit, _ := json.Marshal(protocol.DebugMapEntry{ID: "e", Availability: "captured", Unit: "lib.a"})
		if !strings.Contains(string(withUnit), `"unit":"lib.a"`) {
			t.Fatalf("unit missing from debug-map JSON: %s", withUnit)
		}
	})
}
