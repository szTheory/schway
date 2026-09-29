package interp_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/session"
)

func TestPhase19OpConstInterpreter(t *testing.T) {
	tests := []struct {
		literal string
		want    string
	}{
		{literal: "42", want: "42"},
		{literal: "0x2a", want: "42"},
		{literal: "0b10_1010", want: "42"},
		{literal: "0", want: "0"},
		{literal: "18446744073709551615", want: "18446744073709551615"},
	}
	for _, tt := range tests {
		t.Run(tt.literal, func(t *testing.T) {
			source := []byte("module phase19.interp\nexport {\n  fn main\n}\nfn main(input: Byte) -> U64 {\n  let value = " + tt.literal + "\n  value\n}\n")
			checked := session.Check(source)
			if len(checked.Diagnostics) != 0 {
				t.Fatalf("check diagnostics: %+v", checked.Diagnostics)
			}
			got, err := interp.Run(checked.Program, "main", "7")
			if err != nil {
				t.Fatal(err)
			}
			if got.Outcome.Value != tt.want {
				t.Fatalf("result = %q, want %q", got.Outcome.Value, tt.want)
			}
		})
	}
}
