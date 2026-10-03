package interp

import (
	"bytes"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
)

func TestPhase27RepeatedEventIdentity(t *testing.T) {
	program := checkedPhase27OutputProgram(t, "fizzbuzz.schway")
	var copyID string
	for _, function := range program.Functions {
		if function.Name != "main" || function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind == core.OpCopy {
				copyID = operation.ID + ":event"
				break
			}
		}
	}
	if copyID == "" {
		t.Fatal("FizzBuzz source has no scalar snapshot copy")
	}
	for _, test := range []struct {
		input, output string
		occurrences   []uint64
	}{{"0", "", []uint64{}}, {"1", "1\n", []uint64{0}}, {"2", "1\n2\n", []uint64{0, 1}}} {
		result, err := RunApplication(program, "main", test.input)
		if err != nil || result.Execution.Outcome.Kind != "returned" || result.Execution.Outcome.Value != "" || !bytes.Equal(result.Output, []byte(test.output)) {
			t.Fatalf("input=%s execution=%+v output=%q err=%v; want Unit and separate application bytes %q", test.input, result.Execution, result.Output, err, test.output)
		}
		var got []uint64
		for _, event := range result.Execution.Events {
			if event.ID == copyID {
				got = append(got, event.Occurrence)
			}
			if event.Kind == "application.wrote" {
				t.Fatalf("application byte write leaked into execution evidence: %+v", event)
			}
		}
		if len(got) != len(test.occurrences) {
			t.Fatalf("input=%s copy occurrences=%v, want %v", test.input, got, test.occurrences)
		}
		for index := range got {
			if got[index] != test.occurrences[index] {
				t.Fatalf("input=%s copy occurrences=%v, want %v", test.input, got, test.occurrences)
			}
		}
	}
}
