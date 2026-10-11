package session_test

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// TestPhase29InterpreterFixtureVectors runs the committed fixture program in
// the interpreter for every pinned case id (D-16). Bytes come only from the
// independent Go generator (D-07); the interpreter never reads the C adapter.
func TestPhase29InterpreterFixtureVectors(t *testing.T) {
	source, err := os.ReadFile(phase29FixturePath)
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture diagnostics: %+v", checked.Diagnostics)
	}
	// The tracer: case 3 ("Wikipedia") first, then all ten ids.
	tracer, err := interp.RunWithFixtureBytes(checked.Program, "main", "3", testsupport.Phase29FixtureBytes)
	if err != nil {
		t.Fatal(err)
	}
	if tracer.Outcome.Kind != execution.OutcomeReturned || tracer.Outcome.Value != "300286872" || len(tracer.LiveResources) != 0 {
		t.Fatalf("case 3 = %+v live=%v, want returned 300286872 with no live resources", tracer.Outcome, tracer.LiveResources)
	}
	for _, caseID := range testsupport.Phase29FixtureCaseIDs() {
		want, ok := testsupport.Phase29FixtureExpected(caseID)
		if !ok {
			t.Fatalf("case %d has no pinned checksum", caseID)
		}
		run, err := interp.RunWithFixtureBytes(checked.Program, "main", strconv.FormatUint(caseID, 10), testsupport.Phase29FixtureBytes)
		if err != nil {
			t.Fatalf("case %d: %v", caseID, err)
		}
		if run.Outcome.Kind != execution.OutcomeReturned || run.Outcome.Value != strconv.FormatUint(want, 10) || len(run.LiveResources) != 0 {
			t.Fatalf("case %d = %+v live=%v, want returned %d with no live resources", caseID, run.Outcome, run.LiveResources, want)
		}
	}
}

// phase29InlineFixture wraps a replacement adler32 callee in the committed
// fixture's foreign declarations and main, so the interpreter's terminal rules
// can be probed without committing another .schway file.
func phase29InlineFixture(t *testing.T, callee string) []byte {
	t.Helper()
	source, err := os.ReadFile(phase29FixturePath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	start := strings.Index(text, "fn adler32(")
	end := strings.Index(text, "fn main(")
	if start < 0 || end < start {
		t.Fatal("fixture source no longer has the expected adler32/main layout")
	}
	return []byte(text[:start] + callee + "\n" + text[end:])
}

// TestPhase29InterpreterFixtureUnknownCase pins the acquire failure path: an
// unknown case id publishes no owner and leaves no live resources.
func TestPhase29InterpreterFixtureUnknownCase(t *testing.T) {
	source, err := os.ReadFile(phase29FixturePath)
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture diagnostics: %+v", checked.Diagnostics)
	}
	run, err := interp.RunWithFixtureBytes(checked.Program, "main", "10", testsupport.Phase29FixtureBytes)
	if err != nil {
		t.Fatal(err)
	}
	if run.Outcome.Kind != execution.OutcomeTypedFailure || run.Outcome.Value != "FixtureMissing" {
		t.Fatalf("case 10 outcome = %+v, want typed_failure FixtureMissing", run.Outcome)
	}
	if len(run.LiveResources) != 0 {
		t.Fatalf("case 10 live resources = %v, want none", run.LiveResources)
	}
	for _, event := range run.Events {
		if event.Kind == "resource.released" || event.Kind == "resource.leaked" {
			t.Fatalf("case 10 recorded owner lifecycle event %+v", event)
		}
	}
}

// TestPhase29InterpreterDefectWithLiveOwner pins D-19 and D-20: only the
// byte-index defect may terminate with the fixture owner live. A `defect`
// statement and an overflow defect, each raised with the owner live, stay
// rejected, so the exception cannot widen silently.
func TestPhase29InterpreterDefectWithLiveOwner(t *testing.T) {
	run := func(t *testing.T, callee, input string) (execution.Outcome, []execution.Event, []string, error) {
		t.Helper()
		checked := session.Check(phase29InlineFixture(t, callee))
		if len(checked.Diagnostics) != 0 {
			t.Fatalf("inline fixture diagnostics: %+v", checked.Diagnostics)
		}
		result, err := interp.RunWithFixtureBytes(checked.Program, "main", input, testsupport.Phase29FixtureBytes)
		return result.Outcome, result.Events, result.LiveResources, err
	}

	t.Run("byte index defect is abort-like with the owner live", func(t *testing.T) {
		outcome, events, live, err := run(t, `fn adler32(bytes: ByteView) -> U64 {
  var n = byte_len(bytes)
  var x = byte_at(bytes, n)
  x
}
`, "2")
		if err != nil {
			t.Fatalf("byte index defect with live owner rejected: %v", err)
		}
		if outcome.Kind != execution.OutcomeDefect || outcome.Value != "" {
			t.Fatalf("outcome = %+v, want empty defect", outcome)
		}
		if len(events) == 0 {
			t.Fatal("defect recorded no event")
		}
		final := events[len(events)-1]
		if final.Kind != "function.defected" || final.Output != "byte index out of bounds" {
			t.Fatalf("final event = %+v, want function.defected byte index out of bounds", final)
		}
		if len(live) != 1 {
			t.Fatalf("live resources = %v, want exactly the fixture owner", live)
		}
	})

	t.Run("defect statement with the owner live stays rejected", func(t *testing.T) {
		_, _, _, err := run(t, `fn adler32(bytes: ByteView) -> U64 {
  var n = byte_len(bytes)
  if n < 1000 {
    defect "boom"
  } else {
    n = n + 0
  }
  n
}
`, "2")
		if err == nil || !strings.Contains(err.Error(), "live local resources") {
			t.Fatalf("err = %v, want the live-resources rejection", err)
		}
	})

	t.Run("overflow defect with the owner live stays rejected", func(t *testing.T) {
		_, _, _, err := run(t, `fn adler32(bytes: ByteView) -> U64 {
  var m = 18446744073709551615
  var n = byte_len(bytes)
  m = m + n
  m
}
`, "2")
		if err == nil || !strings.Contains(err.Error(), "live local resources") {
			t.Fatalf("err = %v, want the live-resources rejection", err)
		}
	})
}
