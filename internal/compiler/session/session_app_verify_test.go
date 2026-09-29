package session_test

import (
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
			data: `{"schema":"wrong","schema":"schway.replay-cases/1","cases":[{"id":"case-1","kind":"source","input":"1","expected":{"kind":"returned","value":"1"},"foreign_outcomes":[]}]}`,
		},
		{
			name: "nested expected value",
			data: `{"schema":"schway.replay-cases/1","cases":[{"id":"case-1","kind":"source","input":"1","expected":{"kind":"returned","value":"0","value":"1"},"foreign_outcomes":[]}]}`,
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
