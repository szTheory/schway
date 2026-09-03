package execution_test

import (
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
