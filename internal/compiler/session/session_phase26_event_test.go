package session_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/executionpeer"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase26SourceRepeatedCopyEvidence(t *testing.T) {
	sourcePath := testsupport.ProjectPath("examples", "sum_to_n.schway")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("source-authored scalar-copy check diagnostics: %+v", checked.Diagnostics)
	}
	var copyID, counterID string
	for _, function := range checked.Program.Functions {
		if function.Name != "main" || function.Linear == nil {
			continue
		}
		for _, place := range function.Linear.Places {
			if place.Name == "i" {
				counterID = place.ID
			}
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind == core.OpCopy {
				copyID = operation.ID
				if operation.SourceID != counterID {
					t.Fatalf("source-authored snapshot copies %q, want loop counter %q", operation.SourceID, counterID)
				}
			}
		}
	}
	if copyID == "" || counterID == "" {
		t.Fatal("ordinary sum_to_n source did not lower let snapshot = i to a scalar OpCopy")
	}
	artifact := filepath.Join(t.TempDir(), "sum_to_n")
	_, diagnostics, err := session.BuildApplicationFile(context.Background(), sourcePath, artifact, native.DefaultRunner())
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("BuildApplicationFile diagnostics=%+v err=%v", diagnostics, err)
	}
	for _, test := range []struct {
		input, want string
		count       int
	}{{"3", "6", 3}, {"0", "0", 0}} {
		t.Run("input-"+test.input, func(t *testing.T) {
			interpreted, err := interp.Run(checked.Program, "main", test.input)
			if err != nil || interpreted.Outcome.Kind != "returned" || interpreted.Outcome.Value != test.want {
				t.Fatalf("interpreter outcome=%+v err=%v", interpreted.Outcome, err)
			}
			if err := executionpeer.Validate(checked.Program, interpreted); err != nil {
				t.Fatalf("interpreter peer verification failed: %v", err)
			}
			var interpreterCopies []uint64
			for _, event := range interpreted.Events {
				if event.ID == copyID+":event" {
					interpreterCopies = append(interpreterCopies, event.Occurrence)
				}
			}
			assertPhase26Occurrences(t, "interpreter", interpreterCopies, test.count)

			reportPath := filepath.Join(t.TempDir(), "capture.json")
			var stdout, stderr bytes.Buffer
			outcome, report, err := native.RunApplicationWithEvidence(context.Background(), artifact, test.input, reportPath, native.EvidenceEvents, &stdout, &stderr)
			if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 0 || stdout.String() != test.want+"\n" || stderr.Len() != 0 {
				t.Fatalf("native outcome=%+v stdout=%q stderr=%q err=%v", outcome, stdout.String(), stderr.String(), err)
			}
			if report.CaptureStatus != native.EvidenceStatusComplete || report.Execution == nil {
				t.Fatalf("native event capture status=%q execution=%v", report.CaptureStatus, report.Execution)
			}
			if err := executionpeer.Validate(checked.Program, *report.Execution); err != nil {
				t.Fatalf("native peer verification failed: %v", err)
			}
			var nativeCopies []uint64
			for _, event := range report.Execution.Events {
				if event.ID == copyID+":event" {
					nativeCopies = append(nativeCopies, event.Occurrence)
				}
			}
			assertPhase26Occurrences(t, "native app", nativeCopies, test.count)
		})
	}
}

func assertPhase26Occurrences(t *testing.T, engine string, got []uint64, count int) {
	t.Helper()
	if len(got) != count {
		t.Fatalf("%s source copy events=%v, want %d", engine, got, count)
	}
	for index, occurrence := range got {
		if occurrence != uint64(index) {
			t.Fatalf("%s copy event %d occurrence=%d, want %d", engine, index, occurrence, index)
		}
	}
}
