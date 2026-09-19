package execution_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/execution"
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
