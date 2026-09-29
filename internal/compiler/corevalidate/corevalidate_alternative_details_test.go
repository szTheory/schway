package corevalidate_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
)

// programWithDataType returns a minimal, otherwise-valid core.Program (the
// same shape as ownedProgram) carrying one DataType built directly as a
// struct literal -- deliberately bypassing core.NewDataType -- so a program
// that "arrives already built" (decoded from JSON, produced by a future
// pass, or assembled in a test) can carry an AlternativeDetails list the
// constructor never had a chance to validate (D-12-09's second mechanism,
// Task 1).
func programWithDataType(dataType core.DataType) core.Program {
	program := ownedProgram()
	program.DataTypes = []core.DataType{dataType}
	return program
}

// TestAlternativeDetailsDesynchronizedNameAbsent is Task 1's corpus-wide
// D-12-09 invariant test: an AlternativeDetail naming an alternative absent
// from the owning DataType's Alternatives list is refused by corevalidate
// with a named code, regardless of how the program was constructed (never
// through core.NewDataType here).
func TestAlternativeDetailsDesynchronizedNameAbsent(t *testing.T) {
	dataType := core.DataType{
		ID: "s1:test:module:type:Outcome", Name: "Outcome",
		Alternatives: []string{"Ok", "Err"},
		AlternativeDetails: []core.AlternativeDetail{
			{Name: "Ok", PayloadType: "Buffer"},
			{Name: "Nope", PayloadType: "Buffer"}, // absent from Alternatives
		},
	}
	result := corevalidate.Validate(programWithDataType(dataType))
	if result.Valid {
		t.Fatalf("expected refusal for AlternativeDetails naming an absent alternative, got valid result")
	}
	if len(result.Problems) == 0 || result.Problems[0].Code != "core.alternative_details_desynchronized" {
		t.Fatalf("expected core.alternative_details_desynchronized, got %+v", result.Problems)
	}
}

// TestAlternativeDetailsDesynchronizedDuplicateName is Task 1's second row:
// an AlternativeDetails list naming the same alternative twice is refused by
// the same arm.
func TestAlternativeDetailsDesynchronizedDuplicateName(t *testing.T) {
	dataType := core.DataType{
		ID: "s1:test:module:type:Outcome", Name: "Outcome",
		Alternatives: []string{"Ok", "Err"},
		AlternativeDetails: []core.AlternativeDetail{
			{Name: "Ok", PayloadType: "Buffer"},
			{Name: "Ok", PayloadType: "Byte"}, // duplicate
		},
	}
	result := corevalidate.Validate(programWithDataType(dataType))
	if result.Valid {
		t.Fatalf("expected refusal for a duplicate AlternativeDetails name, got valid result")
	}
	if len(result.Problems) == 0 || result.Problems[0].Code != "core.alternative_details_desynchronized" {
		t.Fatalf("expected core.alternative_details_desynchronized, got %+v", result.Problems)
	}
}

// TestAlternativeDetailsSynchronizedAccepted proves the arm does not over-fire:
// a well-formed, name-set-synchronized AlternativeDetails list (as
// core.NewDataType itself would produce) is accepted.
func TestAlternativeDetailsSynchronizedAccepted(t *testing.T) {
	dataType := core.DataType{
		ID: "s1:test:module:type:Outcome", Name: "Outcome",
		Alternatives: []string{"Ok", "Err"},
		AlternativeDetails: []core.AlternativeDetail{
			{Name: "Ok", PayloadType: "Buffer"},
			{Name: "Err", PayloadType: "Byte"},
		},
	}
	result := corevalidate.Validate(programWithDataType(dataType))
	if !result.Valid {
		t.Fatalf("expected a synchronized AlternativeDetails list to validate, got %+v", result.Problems)
	}
}

// TestAlternativeDetailsEmptyStillAccepted proves every pre-Phase-12 program
// (no AlternativeDetails at all) is unaffected by the new arm.
func TestAlternativeDetailsEmptyStillAccepted(t *testing.T) {
	dataType := core.DataType{
		ID: "s1:test:module:type:Outcome", Name: "Outcome",
		Alternatives: []string{"Ok", "Err"},
	}
	result := corevalidate.Validate(programWithDataType(dataType))
	if !result.Valid {
		t.Fatalf("expected a DataType with no AlternativeDetails to validate, got %+v", result.Problems)
	}
}
