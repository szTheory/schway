package interp_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// phase29MutantSource is the adler32 body with its per-byte read bound through
// an in-loop `let step`. It is var-first, so the in-loop `let` parses through
// the scalar statement list and needs no let-first widening.
const phase29MutantSource = `module phase29.events.mutant

export {
  fn adler32
}

fn adler32(bytes: ByteView) -> U64 {
  var a = 1
  var b = 0
  var i = 0
  var n = byte_len(bytes)
  while i < n {
    let step = byte_at(bytes, i)
    a = a + step
    a = a % 65521
    b = b + a
    b = b % 65521
    i = i + 1
  }
  var r = b
  r = r + r
  r = r + a
  r
}
`

func phase29EventCount(t *testing.T, program core.Program, input []byte) (int, execution.Outcome) {
	t.Helper()
	run, err := interp.RunByteViewForTest(program, "adler32", input)
	if err != nil {
		t.Fatalf("RunByteViewForTest(%d bytes): %v", len(input), err)
	}
	return len(run.Events), run.Outcome
}

// TestPhase29NoLoopEventsInterpreter: the admitted adler32 records the same
// number of events for 4,096 and 5,552-byte inputs (the loop is event-free),
// while the let-in-loop mutant records events that grow with input length, so
// the equality is not vacuous.
func TestPhase29NoLoopEventsInterpreter(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase29", "adler32_library.schway"))
	if err != nil {
		t.Fatal(err)
	}
	library := session.Check(source)
	if len(library.Diagnostics) != 0 {
		t.Fatalf("library diagnostics: %+v", library.Diagnostics)
	}
	short, long := bytes.Repeat([]byte{0x00}, 4096), bytes.Repeat([]byte{0xFF}, 5552)

	shortEvents, shortOutcome := phase29EventCount(t, library.Program, short)
	longEvents, longOutcome := phase29EventCount(t, library.Program, long)
	if shortOutcome.Kind != execution.OutcomeReturned || longOutcome.Kind != execution.OutcomeReturned {
		t.Fatalf("outcomes = %+v / %+v, want returned", shortOutcome, longOutcome)
	}
	if shortEvents != longEvents {
		t.Fatalf("event count depends on input length: %d (4096 bytes) vs %d (5552 bytes)", shortEvents, longEvents)
	}

	// Armed control: the let-in-loop mutant must check clean and then scale.
	mutant := session.Check([]byte(phase29MutantSource))
	if len(mutant.Diagnostics) != 0 {
		t.Fatalf("mutant diagnostics: %+v", mutant.Diagnostics)
	}
	mutantShort, _ := phase29MutantRun(t, mutant.Program, short)
	mutantLong, _ := phase29MutantRun(t, mutant.Program, long)
	if mutantLong <= mutantShort {
		t.Fatalf("let step mutant events did not grow with input length: %d vs %d; the event-free assertion is inert", mutantShort, mutantLong)
	}
}

// phase29MutantRun reports the mutant's event count; an event-capacity failure
// is also an observed failure of the mutant and is reported as a large count.
func phase29MutantRun(t *testing.T, program core.Program, input []byte) (int, bool) {
	t.Helper()
	run, err := interp.RunByteViewForTest(program, "adler32", input)
	if err != nil {
		t.Fatalf("mutant run: %v", err)
	}
	if run.Outcome.Kind != execution.OutcomeReturned {
		t.Logf("mutant outcome: %+v", run.Outcome)
	}
	return len(run.Events), run.Outcome.Kind == execution.OutcomeReturned
}

// TestPhase29ByteLoopOrderAndEmpty pins the BYTE-03 edges: the loop runs zero
// iterations on an empty view (the result is the initial 1), and bytes are
// read in ascending index order (swapping two bytes changes the checksum).
func TestPhase29ByteLoopOrderAndEmpty(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase29", "adler32_library.schway"))
	if err != nil {
		t.Fatal(err)
	}
	library := session.Check(source)
	if len(library.Diagnostics) != 0 {
		t.Fatalf("library diagnostics: %+v", library.Diagnostics)
	}
	value := func(input []byte) string {
		_, outcome := phase29EventCount(t, library.Program, input)
		if outcome.Kind != execution.OutcomeReturned {
			t.Fatalf("outcome = %+v", outcome)
		}
		return outcome.Value
	}
	if got := value(nil); got != "1" {
		t.Fatalf("adler32(empty) = %s, want 1", got)
	}
	// a=1+1=2,b=2 then a=4,b=6 -> 6<<16|4; swapped: a=3,b=3 then a=4,b=7.
	if got := value([]byte{1, 2}); got != "393220" {
		t.Fatalf("adler32(1,2) = %s, want 393220", got)
	}
	if got := value([]byte{2, 1}); got != "458756" {
		t.Fatalf("adler32(2,1) = %s, want 458756", got)
	}
}
