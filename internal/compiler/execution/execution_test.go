package execution_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/execution"
)

func TestExecutionCanonicalBytesExcludePhysicalObservations(t *testing.T) {
	value := execution.Execution{
		Schema:  execution.Schema1,
		Outcome: execution.Outcome{Kind: "returned", Value: "01020304"},
		Events: []execution.Event{{
			Schema: execution.Schema1, ID: "event:0", Kind: "value.transferred",
			FunctionID: "fn:relay", SourcePlace: "place:0", TargetPlace: "place:1", TypeID: "type:0",
		}},
		LiveResources: []string{},
	}
	encoded, err := execution.CanonicalBytes(value)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"optimization", "elapsed", "address", "pid", "stderr", "stdout"} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("physical observation %q entered semantic bytes: %s", forbidden, encoded)
		}
	}
}

func TestPhase26EventOccurrenceEncoding(t *testing.T) {
	var value execution.Execution
	input := `{"schema":"lang.execution/2","outcome":{"kind":"returned","value":"7"},"events":[{"schema":"lang.execution/2","id":"copy:event","kind":"value.copied","function_id":"fn:main","invocation":"fn:main","occurrence":1}],"live_resources":[]}`
	if err := json.Unmarshal([]byte(input), &value); err != nil {
		t.Fatal(err)
	}
	encoded, err := execution.CanonicalBytes(value)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(encoded, []byte(`"occurrence":1`)) {
		t.Fatalf("canonical execution dropped nonzero occurrence: %s", encoded)
	}

	legacy := execution.Execution{Schema: execution.Schema2, Outcome: execution.Outcome{Kind: "returned", Value: "7"}, Events: []execution.Event{{Schema: execution.Schema2, ID: "copy:event", Kind: "value.copied", FunctionID: "fn:main", Invocation: "fn:main"}}, LiveResources: []string{}}
	legacyBytes, err := execution.CanonicalBytes(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(legacyBytes, []byte(`"occurrence"`)) {
		t.Fatalf("zero occurrence changed one-shot canonical bytes: %s", legacyBytes)
	}
}

func TestExecutionLegacyBytesFrozen(t *testing.T) {
	tests := []struct {
		name  string
		value execution.Execution
		want  string
	}{
		{
			name:  "schema zero",
			value: execution.Execution{Schema: execution.Schema0, Outcome: execution.Outcome{Kind: "returned", Value: "7"}, Events: []execution.Event{{Schema: execution.Schema0, ID: "op:return:event", Kind: "function.returned", FunctionID: "fn:main", Input: "7", Output: "7"}}, LiveResources: []string{}},
			want:  `{"schema":"lang.execution/0","outcome":{"kind":"returned","value":"7"},"events":[{"schema":"lang.execution/0","id":"op:return:event","kind":"function.returned","function_id":"fn:main","input":"7","output":"7"}],"live_resources":[]}`,
		},
		{
			name:  "schema one",
			value: execution.Execution{Schema: execution.Schema1, Outcome: execution.Outcome{Kind: "returned", Value: "7"}, Events: []execution.Event{{Schema: execution.Schema1, ID: "op:return:event", Kind: "function.returned", FunctionID: "fn:main", SourcePlace: "place:result", TypeID: "type:Byte"}}, LiveResources: []string{}},
			want:  `{"schema":"lang.execution/1","outcome":{"kind":"returned","value":"7"},"events":[{"schema":"lang.execution/1","id":"op:return:event","kind":"function.returned","function_id":"fn:main","source_place":"place:result","type_id":"type:Byte"}],"live_resources":[]}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := execution.CanonicalBytes(test.value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, []byte(test.want)) {
				t.Fatalf("legacy canonical bytes moved\n got: %s\nwant: %s", got, test.want)
			}
		})
	}
}
