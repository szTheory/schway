package session_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/session"
)

func TestPhase36HelperAnswersAndDefects(t *testing.T) {
	const source = `module phase36.red_crc_step
export { fn main }

fn crc_step(state: U64) -> U64 {
  let shifted = state >> 1
  shifted
}

fn main(state: U64) -> U64 {
  let result = crc_step(state)
  result
}
`
	got, diagnostics, err := session.RunInterpreterApplication([]byte(source), "1")
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("interpreter application diagnostics=%+v err=%v", diagnostics, err)
	}
	if got.Execution.Outcome.Value != "3988292384" {
		t.Fatalf("crc_step(1) = %s, want reflected polynomial result 3988292384", got.Execution.Outcome.Value)
	}
}
