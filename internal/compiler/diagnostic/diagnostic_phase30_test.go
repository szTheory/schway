package diagnostic

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestPhase30UnitIdentity pins D-11: the same code and span in different units
// yields different IDs, the empty unit keeps the pre-phase ID, and the control
// shows why unit has to be in the identity.
func TestPhase30UnitIdentity(t *testing.T) {
	span := Span{Start: 3, End: 8}
	t.Run("diagnostic/0", func(t *testing.T) {
		base := Error("test.code", span, "m")
		a := base.InUnit("adler32")
		b := base.InUnit("lib.a")
		if base.ID == a.ID || base.ID == b.ID || a.ID == b.ID {
			t.Fatalf("ids not distinct: %s %s %s", base.ID, a.ID, b.ID)
		}
		if got := base.InUnit(""); got.ID != base.ID || got.Unit != "" {
			t.Fatalf("InUnit(\"\") changed the diagnostic: %+v", got)
		}
		if a.Unit != "adler32" {
			t.Fatalf("Unit = %q", a.Unit)
		}
	})
	t.Run("diagnostic/1", func(t *testing.T) {
		base := ErrorWithRepairs("test.code", span, "m", nil, Repair{Kind: "k2"}, Repair{Kind: "k1"})
		a := base.InUnit("adler32")
		b := base.InUnit("lib.a")
		if base.ID == a.ID || base.ID == b.ID || a.ID == b.ID {
			t.Fatalf("ids not distinct: %s %s %s", base.ID, a.ID, b.ID)
		}
		if got := base.InUnit(""); got.ID != base.ID {
			t.Fatalf("InUnit(\"\") changed the ID")
		}
		// InUnit must reproduce what a constructor with the same unit would hash.
		if want := errorWithRepairsID("test.code", span, nil, []string{"k1", "k2"}, "adler32"); a.ID != want {
			t.Fatalf("InUnit ID %s, want %s", a.ID, want)
		}
	})
	t.Run("control: unit left out of identity collides", func(t *testing.T) {
		unitOutOfIdentityForTest = true
		defer func() { unitOutOfIdentityForTest = false }()
		base := Error("test.code", span, "m")
		if a, b := base.InUnit("adler32"), base.InUnit("lib.a"); a.ID != b.ID {
			t.Fatalf("control expected a collision, got %s vs %s", a.ID, b.ID)
		}
	})
}

func TestPhase30UnitOmitempty(t *testing.T) {
	span := Span{Start: 10, End: 20}
	causes := []Cause{{Kind: "declared_here", Span: &Span{Start: 1, End: 2}}}
	d := Error("test.code", span, "a fixed message", causes...)
	encoded, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), `"unit"`) {
		t.Fatalf("empty unit leaked into JSON: %s", encoded)
	}
	if d.ID != "diagnostic:d34396fcc8ea6530a1f04dc3" {
		t.Fatalf("pinned ID moved: %s", d.ID)
	}
	withUnit, _ := json.Marshal(d.InUnit("adler32"))
	if !strings.Contains(string(withUnit), `"unit":"adler32"`) {
		t.Fatalf("unit missing: %s", withUnit)
	}
	if got := causes[0].EffectiveUnit("adler32"); got != "adler32" {
		t.Fatalf("cause did not inherit diagnostic unit: %q", got)
	}
	if got := (Cause{Unit: "lib"}).EffectiveUnit("adler32"); got != "lib" {
		t.Fatalf("cause own unit lost: %q", got)
	}
	if got := (Repair{}).EffectiveUnit("adler32"); got != "adler32" {
		t.Fatalf("repair did not inherit: %q", got)
	}
	if got := (Repair{Unit: "lib"}).EffectiveUnit(""); got != "lib" {
		t.Fatalf("repair own unit lost: %q", got)
	}
	if got := d.InUnit("adler32").String(); got != "test.code unit=adler32 [10:20]: a fixed message" {
		t.Fatalf("String() = %q", got)
	}
	if got := d.String(); got != "test.code [10:20]: a fixed message" {
		t.Fatalf("empty-unit String() changed: %q", got)
	}
}
