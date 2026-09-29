package corevalidate_test

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// ---------------------------------------------------------------------
// Phase 09 Plan 06, Task 3 (D-09-30): the peer's OWN independently-written
// corpus-wide closed-consulted-field-set test -- the positive half of
// D-08-26 (check_disclosure_peer_test.go, package check, carries the
// cross-peer identity half). Written from scratch, deliberately NOT a copy
// of check_test.go's TestInterproceduralDisclosedFieldSet body: two
// independently-written assertions of the same closed set is the point,
// not one shared assertion read twice.
//
// D-09-29 records why the ACCEPTED-program disclosure never gets a runtime
// artifact (the consulted set is a compile-time constant, so a runtime
// record would reprint a constant on every run at real qlt02-gated cost
// with zero incremental per-program information) -- this test itself
// stays the permanent test-of-record for the peer's own half of that
// constant, exactly mirroring check_test.go's own role for its half.
// ---------------------------------------------------------------------

// peerAllowedConsultedFields is the ONLY closed set corevalidate's
// interprocedural derivation may ever consult -- exactly check's own
// closed set (check_test.go's allowedFields), because both peers are
// independent re-derivations of the SAME SEM-05 body-blind question and
// must therefore stay within the SAME declared-contract surface, never a
// body fact and never a third field.
var peerAllowedConsultedFields = map[string]bool{"return.mode": true, "parameters[0].mode": true}

// assertPeerConsultedFieldsClosed is this test's own shared assertion
// helper (not shared with check's file -- this package cannot reach that
// file's unexported helpers regardless, and would not even if it could).
func assertPeerConsultedFieldsClosed(t *testing.T, label string, fields []string) {
	t.Helper()
	if len(fields) == 0 {
		t.Fatalf("%s: expected at least one consulted field, got none", label)
	}
	for _, field := range fields {
		if !peerAllowedConsultedFields[field] {
			t.Fatalf("%s: consulted disallowed field %q -- must be a subset of {return.mode, parameters[0].mode}", label, field)
		}
	}
}

// TestPeerDisclosedFieldSet is Task 3(b)'s own machine-assertion of the
// peer's half of D-08-26's criterion 4: over the real testdata/phase08
// `.schway` corpus (plus testdata/phase07/relay_escort_witness.schway, the
// same fixture set check_test.go's own sibling test sweeps) AND the
// synthetic call-graph shapes, corevalidate.Result.PeerConsultedFields()
// is a non-empty subset of exactly {"return.mode", "parameters[0].mode"}
// -- and, over BOTH corpora together, the union is exactly that full set
// (never a proper subset that would leave a hole a mutation could hide
// in).
func TestPeerDisclosedFieldSet(t *testing.T) {
	unionFields := map[string]bool{}

	t.Run("lang corpus", func(t *testing.T) {
		paths, err := filepath.Glob(testsupport.ProjectPath("testdata", "phase08", "*.schway"))
		if err != nil {
			t.Fatalf("glob testdata/phase08: %v", err)
		}
		paths = append(paths, testsupport.ProjectPath("testdata", "phase07", "relay_escort_witness.schway"))
		sort.Strings(paths)
		if len(paths) == 0 {
			t.Fatal("expected at least one fixture")
		}

		for _, path := range paths {
			path := path
			t.Run(filepath.Base(path), func(t *testing.T) {
				source, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("read %s: %v", path, err)
				}
				parsed := syntax.Parse(source)
				if len(parsed.Diagnostics) != 0 {
					t.Fatalf("fixture %s failed to parse: %+v", path, parsed.Diagnostics)
				}
				checkResult := check.Program(parsed.Program)
				validated := corevalidate.Validate(checkResult.Program)
				fields := validated.PeerConsultedFields()
				assertPeerConsultedFieldsClosed(t, filepath.Base(path), fields)
				for _, field := range fields {
					unionFields[field] = true
				}
			})
		}
	})

	t.Run("synthetic call-graph shapes", func(t *testing.T) {
		for _, shape := range testsupport.CallGraphCorpusShapes() {
			shape := shape
			t.Run(shape, func(t *testing.T) {
				program, err := testsupport.GenerateCallGraphCorpus(shape, 32)
				if err != nil {
					t.Fatalf("testsupport.GenerateCallGraphCorpus(%q, 32): %v", shape, err)
				}
				program.Schema = core.Schema1
				program.Module = "peerdisclosure"
				program.ModuleID = "s1:peerdisclosure:module:peerdisclosure"
				program = completePeerCostProgramForCorevalidate(program)
				validated := corevalidate.Validate(program)
				if !validated.Valid {
					t.Fatalf("shape=%s: corevalidate.Validate reported invalid, problems=%v", shape, validated.Problems)
				}
				fields := validated.PeerConsultedFields()
				assertPeerConsultedFieldsClosed(t, shape, fields)
				for _, field := range fields {
					unionFields[field] = true
				}
			})
		}
	})

	var union []string
	for field := range unionFields {
		union = append(union, field)
	}
	sort.Strings(union)
	if len(union) != len(peerAllowedConsultedFields) {
		t.Fatalf("union of consulted fields across both corpora = %v, want exactly %d fields (return.mode, parameters[0].mode)", union, len(peerAllowedConsultedFields))
	}
	for field := range peerAllowedConsultedFields {
		if !unionFields[field] {
			t.Fatalf("field %q was never consulted by either corpus -- the closed-set claim would be vacuously narrow", field)
		}
	}
}
