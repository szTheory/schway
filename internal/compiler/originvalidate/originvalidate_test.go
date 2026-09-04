package originvalidate_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/originvalidate"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestOriginValidateImportsNeitherCheckNorAst is 03-06-01's structural
// falsifier for the plan's binding prohibition: originvalidate decides from
// the typed-core artifact alone and must never import the checker or the
// AST package. Reading the actual import lists (not trusting a doc comment)
// is the same technique the codebase already applies to enforce boundaries
// mechanically rather than by convention.
func TestOriginValidateImportsNeitherCheckNorAst(t *testing.T) {
	dir := testsupport.ProjectPath("internal", "compiler", "originvalidate")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	fileSet := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fileSet, filepath.Join(dir, entry.Name()), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imported := range file.Imports {
			path := strings.Trim(imported.Path.Value, `"`)
			if strings.HasSuffix(path, "/compiler/check") || strings.HasSuffix(path, "/compiler/ast") {
				t.Fatalf("%s imports %s, which originvalidate must never depend on", entry.Name(), path)
			}
		}
	}
}

func honestProgram(t testing.TB, fixture string) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase3", fixture))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	if checked.Program.Functions[0].PublicOrigin == nil {
		t.Fatalf("expected PublicOrigin on the honestly-checked function")
	}
	return checked.Program
}

// TestOriginUnderstatedRejected is 03-06-02's falsifier for the "omitted"
// defect: a declared origin set that omits a path the body actually derives
// from is rejected as core.origin_understated, detected purely by
// source-blind recomputation from the typed core (never by trusting the
// declaration). The dishonest declaration is injected by mutating an
// honestly-checked program's own fact — check.go's honest producer can never
// construct this shape itself, exactly as OV-02-01's mutation-kill precedent
// establishes for a different fact.
func TestOriginUnderstatedRejected(t *testing.T) {
	program := honestProgram(t, "public_view_understated.lang")
	program.Functions[0].PublicOrigin = &core.PublicOrigin{Paths: []string{}, Access: "shared"}
	problems := originvalidate.ValidatePublished(program)
	if len(problems) != 1 || problems[0].Code != "core.origin_understated" {
		t.Fatalf("expected exactly core.origin_understated, got %+v", problems)
	}
}

// TestOriginAccessMismatchRejected is 03-06-02's falsifier for the
// "impossible" defect: a declared access mode the body cannot produce is
// rejected as core.origin_access_mismatch by the same recomputation.
func TestOriginAccessMismatchRejected(t *testing.T) {
	program := honestProgram(t, "public_view_impossible.lang")
	program.Functions[0].PublicOrigin = &core.PublicOrigin{Paths: []string{"buffer"}, Access: "exclusive"}
	problems := originvalidate.ValidatePublished(program)
	if len(problems) != 1 || problems[0].Code != "core.origin_access_mismatch" {
		t.Fatalf("expected exactly core.origin_access_mismatch, got %+v", problems)
	}
}

// TestStaleSummaryRejectedBeforeOtherChecks is 03-06-02's falsifier for
// T-03-03: a summary whose recorded digest does not match the core artifact
// it is checked against is rejected before any origin or access question is
// even asked — CheckSummary never unmarshals coreBytes into a struct that
// could carry a body, so the ordering is structural, not merely sequenced.
func TestStaleSummaryRejectedBeforeOtherChecks(t *testing.T) {
	program := honestProgram(t, "public_view.lang")
	summary, err := originvalidate.BuildInterface(program)
	if err != nil {
		t.Fatal(err)
	}
	summaryBytes, err := json.Marshal(summary)
	if err != nil {
		t.Fatal(err)
	}
	staleCore := []byte(`{"not":"the real core artifact"}`)
	if _, err := originvalidate.CheckSummary(summaryBytes, staleCore); err == nil {
		t.Fatal("expected a stale-summary rejection")
	} else if code := errorCode(err); code != "origin.stale_summary" {
		t.Fatalf("expected origin.stale_summary, got %q (%v)", code, err)
	}

	// A summary whose declared origin would itself be dishonest must still be
	// rejected for staleness first, proving the digest check runs before any
	// origin-shaped decision — CheckSummary has no other check to reorder
	// against, which is itself the point: there IS no origin/access check
	// left to run once the digest fails.
	dishonest := summary
	dishonest.Functions[0].PublicOrigin = &core.PublicOrigin{Paths: []string{"nonexistent"}, Access: "exclusive"}
	dishonestBytes, err := json.Marshal(dishonest)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := originvalidate.CheckSummary(dishonestBytes, staleCore); errorCode(err) != "origin.stale_summary" {
		t.Fatalf("expected origin.stale_summary ahead of any origin content, got %v", err)
	}
}

// TestOriginEscapeIsNamed is 03-06-02's falsifier naming the accepted
// residual: a coordinated frontend-and-summary lie is declared as a named
// expected escape, never solved and never silently absent.
func TestOriginEscapeIsNamed(t *testing.T) {
	if originvalidate.KnownEscape == "" {
		t.Fatal("KnownEscape must be a non-empty named constant")
	}
	escapes := originvalidate.ExpectedEscapes()
	found := false
	for _, escape := range escapes {
		if escape == originvalidate.KnownEscape {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected KnownEscape among ExpectedEscapes, got %+v", escapes)
	}
}

// TestInterfaceSummaryOmitsBodies is 03-06-01's falsifier: the exported
// interface summary contains no linear or match body — not merely an omitted
// field, but a shape (core.FunctionSignature) that structurally has no such
// field to omit.
func TestInterfaceSummaryOmitsBodies(t *testing.T) {
	program := honestProgram(t, "public_view.lang")
	summary, err := originvalidate.BuildInterface(program)
	if err != nil {
		t.Fatal(err)
	}
	if len(summary.Functions) != 1 {
		t.Fatalf("expected one function signature, got %d", len(summary.Functions))
	}
	if summary.Functions[0].PublicOrigin == nil || summary.Functions[0].PublicOrigin.Access != "shared" {
		t.Fatalf("expected the origin fact to survive stripping: %+v", summary.Functions[0])
	}
	encoded, err := json.Marshal(summary)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{`"linear"`, `"match"`, `"operations"`, `"blocks"`} {
		if bytes.Contains(encoded, []byte(forbidden)) {
			t.Fatalf("interface summary leaked a body field %s: %s", forbidden, encoded)
		}
	}
}

func errorCode(err error) string {
	var typed *originvalidate.Error
	if errors.As(err, &typed) {
		return typed.Code
	}
	return ""
}
