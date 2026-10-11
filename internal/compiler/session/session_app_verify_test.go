package session_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/session"
)

func TestDecodeReplayCasesRejectsDuplicateJSONKeys(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			name: "top-level schema",
			data: `{"schema":"wrong","schema":"lang.replay-cases/1","cases":[{"id":"case-1","kind":"source","input":"1","expected":{"kind":"returned","value":"1"},"foreign_outcomes":[]}]}`,
		},
		{
			name: "nested expected value",
			data: `{"schema":"lang.replay-cases/1","cases":[{"id":"case-1","kind":"source","input":"1","expected":{"kind":"returned","value":"0","value":"1"},"foreign_outcomes":[]}]}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := session.DecodeReplayCases([]byte(test.data)); err == nil {
				t.Fatal("DecodeReplayCases accepted duplicate JSON keys")
			}
		})
	}
}

// TestDecodeReplayCasesCaseIDBound pins D-16: case IDs are printable ASCII and
// at most MaxReplayCaseIDBytes long.
func TestDecodeReplayCasesCaseIDBound(t *testing.T) {
	decode := func(id string) error {
		quoted, marshalErr := json.Marshal(id)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		data := `{"schema":"lang.replay-cases/1","cases":[{"id":` + string(quoted) + `,"kind":"source","input":"1","expected":{"kind":"returned","value":"1"},"foreign_outcomes":[]}]}`
		_, err := session.DecodeReplayCases([]byte(data))
		return err
	}
	if session.MaxReplayCaseIDBytes != 128 {
		t.Fatalf("MaxReplayCaseIDBytes=%d, want 128", session.MaxReplayCaseIDBytes)
	}
	if err := decode(strings.Repeat("a", 128)); err != nil {
		t.Fatalf("128-byte printable id refused: %v", err)
	}
	if err := decode(strings.Repeat("a", 129)); err == nil || !strings.Contains(err.Error(), "128") {
		t.Fatalf("129-byte id error=%v, want refusal naming the 128-byte bound", err)
	}
	for name, id := range map[string]string{
		"tab": "a\tb", "newline": "a\nb", "del": "a\x7fb", "high-byte": "a\u00e9b", "empty": "", "whitespace": "   ",
	} {
		if err := decode(id); err == nil {
			t.Fatalf("%s id %q accepted", name, id)
		}
	}
	if err := decode("space inside ~!"); err != nil {
		t.Fatalf("printable id with spaces refused: %v", err)
	}
}
