package corevalidate_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

func phase19ConstantProgram(t *testing.T) core.Program {
	t.Helper()
	parsed := syntax.Parse([]byte("module phase19.peer\nexport {\n  fn main\n}\nfn main(input: Byte) -> U64 {\n  let value = 42\n  value\n}\n"))
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("parse literal witness: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check literal witness: %+v", checked.Diagnostics)
	}
	return checked.Program
}

func TestPhase19OpConstIndependentAdmission(t *testing.T) {
	valid := phase19ConstantProgram(t)
	result := corevalidate.Validate(valid)
	if !result.Valid {
		t.Fatalf("checker-produced OpConst rejected: %+v", result.Problems)
	}
	operation := valid.Functions[0].Linear.Operations[0]
	if operation.Kind != core.OpConst || operation.ConstU64 != "42" || operation.SourceID != "" {
		t.Fatalf("unexpected constant fact: %+v", operation)
	}
}

func TestPhase19ForgedOpConstFactsRefused(t *testing.T) {
	tests := []struct {
		name string
		edit func(*core.LinearOperation)
	}{
		{"empty value", func(op *core.LinearOperation) { op.ConstU64 = "" }},
		{"leading zero", func(op *core.LinearOperation) { op.ConstU64 = "042" }},
		{"sign", func(op *core.LinearOperation) { op.ConstU64 = "+42" }},
		{"nondigit", func(op *core.LinearOperation) { op.ConstU64 = "4x" }},
		{"overflow", func(op *core.LinearOperation) { op.ConstU64 = "18446744073709551616" }},
		{"forged source", func(op *core.LinearOperation) { op.SourceID = "invented" }},
		{"loan field", func(op *core.LinearOperation) { op.LoanID = "invented-loan" }},
		{"callee field", func(op *core.LinearOperation) { op.CalleeID = "invented-callee" }},
		{"payload field", func(op *core.LinearOperation) { op.PayloadType = "invented-payload" }},
		{"wrong type", func(op *core.LinearOperation) { op.TypeID = "missing-type" }},
		{"missing target", func(op *core.LinearOperation) { op.TargetID = "" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutated := cloneProgram(t, phase19ConstantProgram(t))
			test.edit(&mutated.Functions[0].Linear.Operations[0])
			if got := corevalidate.Validate(mutated); got.Valid {
				t.Fatalf("forged OpConst was admitted: %+v", got)
			}
		})
	}
}
