// This file is plan 12-07's reverted-fix control (CR-01/IN-01/D-12-25),
// the way TestDuplicatePayloadTypeRefused is plan 12-06's. It proves the
// exactly-one-derivation invariant on two independent axes: the behavior
// of the shared derivation itself (TestAlternativeNameForPayloadTypeAmbiguityRefused,
// TestLookupAlternativeDetailMatchesCheckSemantics), and the structural
// fact that neither engine reimplements it locally
// (TestPayloadAlternativeResolutionHasExactlyOneDerivation, appended by
// plan 12-07 Task 2).
package core_test

import (
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// collidingPayloadDataType mirrors session.PayloadProbeDataType's own
// First(Byte)/Second(Byte) shape: two distinctly-named alternatives
// declaring the SAME payload type, built via core.NewDataType exactly the
// way a hand-built (never-parsed) core.Program could assemble it.
func collidingPayloadDataType(t *testing.T) core.DataType {
	t.Helper()
	dataType, err := core.NewDataType("test:type:Colliding", "Colliding",
		[]string{"First", "Second"},
		[]core.AlternativeDetail{
			{Name: "First", PayloadType: "Byte"},
			{Name: "Second", PayloadType: "Byte"},
		}, diagnostic.Span{})
	if err != nil {
		t.Fatalf("collidingPayloadDataType: %v", err)
	}
	return dataType
}

// unambiguousPayloadDataType is a data type whose alternatives declare
// distinct payload types, mirroring payload_tracer.schway's own Outcome
// shape (D-12-41).
func unambiguousPayloadDataType(t *testing.T) core.DataType {
	t.Helper()
	dataType, err := core.NewDataType("test:type:Outcome", "Outcome",
		[]string{"Ok", "Failed", "Nullary"},
		[]core.AlternativeDetail{
			{Name: "Ok", PayloadType: "Byte"},
			{Name: "Failed", PayloadType: "Buffer"},
		}, diagnostic.Span{})
	if err != nil {
		t.Fatalf("unambiguousPayloadDataType: %v", err)
	}
	return dataType
}

// TestAlternativeNameForPayloadTypeAmbiguityRefused covers
// core.AlternativeNameForPayloadType's five behaviors: unique match,
// ambiguous match (naming both colliding alternatives, never a first-match
// guess), no match, and the empty-payload-type query being an error rather
// than a nullary match.
func TestAlternativeNameForPayloadTypeAmbiguityRefused(t *testing.T) {
	t.Run("exactly one alternative declares the payload type", func(t *testing.T) {
		dataType := unambiguousPayloadDataType(t)
		name, err := core.AlternativeNameForPayloadType(dataType, "Byte")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if name != "Ok" {
			t.Fatalf("got %q, want %q", name, "Ok")
		}
	})

	t.Run("two alternatives collide: ambiguity refused, both names present, no first match returned", func(t *testing.T) {
		dataType := collidingPayloadDataType(t)
		name, err := core.AlternativeNameForPayloadType(dataType, "Byte")
		if err == nil {
			t.Fatalf("expected ambiguity error, got nil (silently resolved to %q)", name)
		}
		if name != "" {
			t.Fatalf("expected empty name alongside the error, got %q", name)
		}
		if !strings.Contains(err.Error(), "First") || !strings.Contains(err.Error(), "Second") {
			t.Fatalf("ambiguity error must name BOTH colliding alternatives, got: %v", err)
		}
		if !strings.Contains(err.Error(), "Colliding") {
			t.Fatalf("ambiguity error must name the data type, got: %v", err)
		}
	})

	t.Run("no alternative declares the payload type: distinguishable from ambiguity", func(t *testing.T) {
		dataType := unambiguousPayloadDataType(t)
		name, err := core.AlternativeNameForPayloadType(dataType, "SCHWAY_NONEXISTENT")
		if err == nil {
			t.Fatalf("expected no-match error, got nil (name %q)", name)
		}
		if strings.Contains(err.Error(), "ambiguous") {
			t.Fatalf("no-match error must be distinguishable from the ambiguity error, got: %v", err)
		}
	})

	t.Run("empty payload type query is always an error, never a nullary match", func(t *testing.T) {
		dataType := unambiguousPayloadDataType(t)
		name, err := core.AlternativeNameForPayloadType(dataType, "")
		if err == nil {
			t.Fatalf("expected error for empty payload type query, got nil (name %q)", name)
		}
	})
}

// TestLookupAlternativeDetailMatchesCheckSemantics asserts
// core.LookupAlternativeDetail's known-name and unknown-name results match
// what check.lookupAlternativeDetail returned before it became a
// delegation to this shared helper.
func TestLookupAlternativeDetailMatchesCheckSemantics(t *testing.T) {
	dataType := unambiguousPayloadDataType(t)

	t.Run("known name returns its declared detail", func(t *testing.T) {
		detail := core.LookupAlternativeDetail(dataType, "Ok")
		if detail.Name != "Ok" || detail.PayloadType != "Byte" {
			t.Fatalf("got %+v, want Name=Ok PayloadType=Byte", detail)
		}
	})

	t.Run("unknown name returns a zero-valued nullary detail carrying only that name", func(t *testing.T) {
		detail := core.LookupAlternativeDetail(dataType, "Nonexistent")
		if detail.Name != "Nonexistent" || detail.PayloadType != "" {
			t.Fatalf("got %+v, want Name=Nonexistent PayloadType=\"\"", detail)
		}
	})
}

// engineSourceHasNoLocalResolverDeclaration asserts, per file, that no
// line BEGINS WITH a func declaration of resolverName. The scan is
// anchored at the start of the line (after trimming leading whitespace) so
// a doc comment merely mentioning the identifier -- e.g. this very file's
// own comments naming "alternativeNameForPayloadType" -- can never trip it;
// only an actual `func alternativeNameForPayloadType(...)` declaration
// does. This mirrors core_convention_absence_test.go's own precedent of
// explaining why a check is a tripwire (line-anchored) rather than a
// heuristic (an unanchored substring scan, which prose can trip).
func engineSourceHasNoLocalResolverDeclaration(t *testing.T, relativePath, resolverName string) {
	t.Helper()
	path := testsupport.ProjectPath(relativePath)
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", relativePath, err)
	}
	prefix := "func " + resolverName
	for _, line := range strings.Split(string(source), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			t.Fatalf("%s: found a local declaration of %q -- the shared core.AlternativeNameForPayloadType derivation must not be reimplemented locally (CR-01/D-12-25)", relativePath, resolverName)
		}
	}
}

// engineSourceUsesSharedHelper asserts POSITIVELY that relativePath
// contains at least one reference to the exported core helper named --
// so this tripwire cannot be satisfied by simply deleting payload
// resolution altogether. Both directions (absence of the local copy AND
// presence of the shared one) are required for the test to mean anything.
func engineSourceUsesSharedHelper(t *testing.T, relativePath, helperReference string) {
	t.Helper()
	path := testsupport.ProjectPath(relativePath)
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", relativePath, err)
	}
	if !strings.Contains(string(source), helperReference) {
		t.Fatalf("%s: expected at least one reference to %q -- payload alternative resolution must read the shared core derivation, not be silently deleted", relativePath, helperReference)
	}
}

// TestPayloadAlternativeResolutionHasExactlyOneDerivation is plan 12-07's
// reverted-fix control, the way TestDuplicatePayloadTypeRefused is plan
// 12-06's: it goes red if either engine re-grows a local
// alternativeNameForPayloadType declaration, or if either engine stops
// reading the shared core.AlternativeNameForPayloadType.
//
// This test protects a defect class the project's own three-engine
// convergence tests (TestPayloadTracerThreeEngineAgreement,
// TestN1ConvergenceDifferential) CANNOT catch by construction: cgen and
// interp previously carried the SAME ambiguous derivation, independently
// reimplemented. Convergence testing only detects DISAGREEMENT between
// engines -- two engines silently agreeing on the same wrong guess at
// "which alternative does this belong to" produces perfect agreement, and
// a comparison-based control sees nothing wrong. A reviewer re-adding a
// local copy "for convenience" must turn THIS test red, in that
// reviewer's own diff, not rely on a convergence test that structurally
// cannot see the regression.
func TestPayloadAlternativeResolutionHasExactlyOneDerivation(t *testing.T) {
	const resolverName = "alternativeNameForPayloadType"

	t.Run("cgen: no local declaration", func(t *testing.T) {
		engineSourceHasNoLocalResolverDeclaration(t, "internal/compiler/cgen/cgen.go", resolverName)
	})
	t.Run("cgen: uses the shared core helper", func(t *testing.T) {
		engineSourceUsesSharedHelper(t, "internal/compiler/cgen/cgen.go", "core.AlternativeNameForPayloadType")
	})
	t.Run("interp: no local declaration", func(t *testing.T) {
		engineSourceHasNoLocalResolverDeclaration(t, "internal/compiler/interp/interp.go", resolverName)
	})
	t.Run("interp: uses the shared core helper", func(t *testing.T) {
		engineSourceUsesSharedHelper(t, "internal/compiler/interp/interp.go", "core.AlternativeNameForPayloadType")
	})
}
