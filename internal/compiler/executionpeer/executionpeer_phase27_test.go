package executionpeer_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/executionpeer"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase27PeerOccurrenceOrder(t *testing.T) {
	sourcePath := testsupport.ProjectPath("examples", "fizzbuzz.schway")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("FizzBuzz check diagnostics: %+v", checked.Diagnostics)
	}
	var copyID string
	for _, function := range checked.Program.Functions {
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
		t.Fatal("FizzBuzz source has no loop snapshot copy site")
	}
	artifact := filepath.Join(t.TempDir(), "fizzbuzz")
	if _, diagnostics, err := session.BuildApplicationFile(context.Background(), sourcePath, artifact, native.DefaultRunner()); err != nil || len(diagnostics) != 0 {
		t.Fatalf("BuildApplicationFile diagnostics=%+v err=%v", diagnostics, err)
	}
	var twoCopyDocument execution.Execution
	var twoCopyIndexes []int
	for _, test := range []struct {
		input, wantOutput string
		wantOccurrences   []uint64
	}{{"0", "", []uint64{}}, {"1", "1\n", []uint64{0}}, {"2", "1\n2\n", []uint64{0, 1}}} {
		t.Run("input-"+test.input, func(t *testing.T) {
			interpreted, diagnostics, err := session.RunInterpreterApplication(source, test.input)
			if err != nil || len(diagnostics) != 0 || interpreted.Execution.Outcome.Kind != execution.OutcomeReturned || string(interpreted.Output) != test.wantOutput {
				t.Fatalf("interpreter diagnostics=%+v outcome=%+v output=%q err=%v", diagnostics, interpreted.Execution.Outcome, interpreted.Output, err)
			}
			interpreterDocument, err := session.ProjectExecutionSchema2(checked.Program, interpreted.Execution)
			if err != nil {
				t.Fatalf("project interpreter execution: %v", err)
			}
			if err := executionpeer.Validate(checked.Program, interpreterDocument); err != nil {
				t.Fatalf("peer rejected interpreter input %s: %v", test.input, err)
			}
			if got := phase27CopyOccurrences(interpreterDocument, copyID); !equalUint64s(got, test.wantOccurrences) {
				t.Fatalf("interpreter input=%s copy occurrences=%v, want %v", test.input, got, test.wantOccurrences)
			}

			reportPath := filepath.Join(t.TempDir(), "capture.json")
			var stdout, stderr bytes.Buffer
			outcome, report, err := native.RunApplicationWithEvidence(context.Background(), artifact, test.input, reportPath, native.EvidenceEvents, &stdout, &stderr)
			if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 0 || report.CaptureStatus != native.EvidenceStatusComplete || report.Execution == nil || !bytes.Equal(stdout.Bytes(), []byte(test.wantOutput)) || stderr.Len() != 0 {
				t.Fatalf("native input=%s outcome=%+v report=%+v stdout=%q stderr=%q err=%v", test.input, outcome, report, stdout.Bytes(), stderr.String(), err)
			}
			if err := executionpeer.Validate(checked.Program, *report.Execution); err != nil {
				t.Fatalf("peer rejected native input %s: %v", test.input, err)
			}
			if got := phase27CopyOccurrences(*report.Execution, copyID); !equalUint64s(got, test.wantOccurrences) {
				t.Fatalf("native input=%s copy occurrences=%v, want %v", test.input, got, test.wantOccurrences)
			}
			if test.input == "2" {
				twoCopyDocument = *report.Execution
				for index, event := range twoCopyDocument.Events {
					if event.ID == copyID {
						twoCopyIndexes = append(twoCopyIndexes, index)
					}
				}
			}
		})
	}
	if len(twoCopyIndexes) != 2 {
		t.Fatalf("native input 2 copy event indexes=%v, want two", twoCopyIndexes)
	}
	controls := []struct {
		name   string
		mutate func(*execution.Execution)
	}{
		{name: "collision", mutate: func(document *execution.Execution) {
			document.Events[twoCopyIndexes[1]].Occurrence = document.Events[twoCopyIndexes[0]].Occurrence
		}},
		{name: "omitted-first-event", mutate: func(document *execution.Execution) {
			index := twoCopyIndexes[0]
			document.Events = append(document.Events[:index], document.Events[index+1:]...)
		}},
		{name: "reordered-events", mutate: func(document *execution.Execution) {
			document.Events[twoCopyIndexes[0]], document.Events[twoCopyIndexes[1]] = document.Events[twoCopyIndexes[1]], document.Events[twoCopyIndexes[0]]
		}},
	}
	for _, control := range controls {
		t.Run(control.name, func(t *testing.T) {
			forged := clonePhase27Execution(twoCopyDocument)
			control.mutate(&forged)
			if err := executionpeer.Validate(checked.Program, forged); err == nil {
				t.Fatalf("peer accepted %s repeated-copy control", control.name)
			}
		})
	}
}

func phase27CopyOccurrences(document execution.Execution, copyID string) []uint64 {
	var result []uint64
	for _, event := range document.Events {
		if event.ID == copyID {
			result = append(result, event.Occurrence)
		}
	}
	return result
}

func equalUint64s(left, right []uint64) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func clonePhase27Execution(document execution.Execution) execution.Execution {
	clone := document
	clone.Events = append([]execution.Event(nil), document.Events...)
	clone.LiveResources = append([]string(nil), document.LiveResources...)
	return clone
}
