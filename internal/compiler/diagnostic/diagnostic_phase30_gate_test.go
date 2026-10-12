package diagnostic

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
)

func gateID(identity any) (string, []byte) {
	encoded, _ := json.Marshal(identity)
	sum := sha256.Sum256(encoded)
	return "diagnostic:" + hex.EncodeToString(sum[:12]), encoded
}

// TestPhase30GateUnitOmitemptyInvariance is flag (c) of the D-14 gate: a
// trailing `Unit string` tagged omitempty leaves every empty-unit identity, and
// therefore every pinned diagnostic ID, byte-identical (D-11).
func TestPhase30GateUnitOmitemptyInvariance(t *testing.T) {
	span := Span{Start: 10, End: 20}
	causes := []Cause{{Kind: "declared_here", Span: &Span{Start: 1, End: 2}}}

	type current struct {
		Schema string
		Code   string
		Span   Span
		Causes []Cause
	}
	type withOmitempty struct {
		Schema string
		Code   string
		Span   Span
		Causes []Cause
		Unit   string `json:",omitempty"`
	}
	type withoutOmitempty struct {
		Schema string
		Code   string
		Span   Span
		Causes []Cause
		Unit   string
	}
	idCurrent, bytesCurrent := gateID(current{Schema, "test.code", span, causes})
	idOmit, bytesOmit := gateID(withOmitempty{Schema, "test.code", span, causes, ""})
	_, bytesPlain := gateID(withoutOmitempty{Schema, "test.code", span, causes, ""})
	if string(bytesCurrent) != string(bytesOmit) {
		t.Fatalf("omitempty Unit changed identity bytes:\n%s\n%s", bytesCurrent, bytesOmit)
	}
	const wantID = "diagnostic:d34396fcc8ea6530a1f04dc3"
	if idCurrent != wantID || idOmit != wantID {
		t.Fatalf("ids %q / %q, want pinned %q", idCurrent, idOmit, wantID)
	}
	if got := Error("test.code", span, "a fixed message", causes...).ID; got != wantID {
		t.Fatalf("Error() ID %q, want %q", got, wantID)
	}
	if string(bytesPlain) == string(bytesCurrent) {
		t.Fatal("control: a Unit field without omitempty must change the identity bytes")
	}

	type currentRepairs struct {
		Schema      string
		Code        string
		Span        Span
		Causes      []Cause
		RepairKinds []string
	}
	type repairsOmit struct {
		Schema      string
		Code        string
		Span        Span
		Causes      []Cause
		RepairKinds []string
		Unit        string `json:",omitempty"`
	}
	type repairsPlain struct {
		Schema      string
		Code        string
		Span        Span
		Causes      []Cause
		RepairKinds []string
		Unit        string
	}
	kinds := []string{"add_borrow"}
	want := ErrorWithRepairs("test.code", span, "m", causes, Repair{Kind: "add_borrow"}).ID
	idA, bytesA := gateID(currentRepairs{Schema1, "test.code", span, causes, kinds})
	idB, bytesB := gateID(repairsOmit{Schema1, "test.code", span, causes, kinds, ""})
	_, bytesC := gateID(repairsPlain{Schema1, "test.code", span, causes, kinds, ""})
	if idA != want || idB != want || string(bytesA) != string(bytesB) {
		t.Fatalf("ErrorWithRepairs shape moved: %q %q want %q", idA, idB, want)
	}
	if string(bytesC) == string(bytesA) {
		t.Fatal("control: ErrorWithRepairs Unit without omitempty must change the bytes")
	}
}
