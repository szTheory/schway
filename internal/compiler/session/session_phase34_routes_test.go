package session_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/modload"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/session"
)

func TestPhase34SourceRoutePresentation(t *testing.T) {
	source := []byte("module draft.route\nexport { fn main }\nfn main(n: U64) -> U64 { n + }\n")
	path := filepath.Join(t.TempDir(), "route.schway")
	if err := os.WriteFile(path, source, 0o600); err != nil {
		t.Fatal(err)
	}

	reads := map[string]int{}
	restore := modload.SetReadObserverForTest(func(module string) { reads[module]++ })
	defer restore()

	checkOnce := func(route string) {
		t.Helper()
		if reads[""] != 1 || len(reads) != 1 {
			t.Fatalf("%s source reads = %v, want one entry snapshot", route, reads)
		}
		clear(reads)
	}
	assertCaptured := func(route string, result protocol.Result) {
		t.Helper()
		if result.Status != protocol.StatusInvalid || len(result.Diagnostics) == 0 || result.Presentation == nil {
			t.Fatalf("%s result = %+v, want invalid diagnostic with captured source", route, result)
		}
		captured, ok := result.Presentation.Sources[""]
		if !ok || captured.Label != path || !bytes.Equal(captured.Bytes, source) {
			t.Fatalf("%s presentation = %+v, want the one loaded source snapshot", route, result.Presentation)
		}
		human, err := protocol.Human(result)
		if err != nil {
			t.Fatal(err)
		}
		problem := result.Diagnostics[0]
		if problem.Primary.Line != 0 || problem.Primary.Col != 0 || !strings.Contains(human, path+":3:") || !strings.Contains(human, problem.ID) || !strings.Contains(human, problem.Code) || !strings.Contains(human, "n + }") || !strings.Contains(human, "^") {
			t.Fatalf("%s diagnostic lost its captured source attribution: %+v\n%s", route, problem, human)
		}
	}

	interpreter, err := session.RunInterpreterCommandFile(path)
	if err != nil {
		t.Fatal(err)
	}
	assertCaptured("interpreter", interpreter)
	checkOnce("interpreter")

	runner := &phase34RouteCountingRunner{}
	nativeResult, err := session.RunNativeCommandFile(context.Background(), path, runner)
	if err != nil {
		t.Fatal(err)
	}
	assertCaptured("native", nativeResult)
	if runner.calls != 0 {
		t.Fatalf("native runner received %d calls for rejected syntax, want zero", runner.calls)
	}
	checkOnce("native")

	// BuildApplicationFile is the retained public API. Its source-refusal
	// result remains unchanged while the CLI command gains presentation.
	_, diagnostics, err := session.BuildApplicationFile(context.Background(), path, filepath.Join(t.TempDir(), "app"), native.Runner{})
	if err != nil || len(diagnostics) == 0 || diagnostics[0].Code != interpreter.Diagnostics[0].Code {
		t.Fatalf("BuildApplicationFile refusal = diagnostics %v, err %v; want the same source refusal", diagnostics, err)
	}
	checkOnce("build")
}

type phase34RouteCountingRunner struct{ calls int }

func (runner *phase34RouteCountingRunner) Run(context.Context, string, string, []string) (native.Result, error) {
	runner.calls++
	return native.Result{}, nil
}

func TestPhase34DiagnosticRouteInventory(t *testing.T) {
	source := []byte("module draft.route\nexport { fn main }\nfn main(n: U64) -> U64 { n + }\n")
	makeSource := func() string {
		t.Helper()
		path := filepath.Join(t.TempDir(), "route.schway")
		if err := os.WriteFile(path, source, 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}

	formatPath := makeSource()
	formatted, err := session.FormatCommandFile(formatPath, false)
	if err != nil {
		t.Fatal(err)
	}
	phase34AssertLocatedRoute(t, "format", formatPath, source, formatted)

	phase34InstallFakeClang(t)
	evidencePath := makeSource()
	_, evidenceResult, err := session.EvidenceCommandFile(context.Background(), evidencePath)
	if err != nil {
		t.Fatal(err)
	}
	phase34AssertLocatedRoute(t, "evidence", evidencePath, source, evidenceResult)

	for _, route := range []struct {
		name string
		run  func(string) (protocol.Result, error)
	}{
		{name: "interface export", run: func(path string) (protocol.Result, error) {
			return session.InterfaceExportCommandFile(path, filepath.Join(t.TempDir(), "summary.json"))
		}},
		{name: "interface core", run: func(path string) (protocol.Result, error) {
			return session.InterfaceCoreCommandFile(path, filepath.Join(t.TempDir(), "core.json"))
		}},
	} {
		path := makeSource()
		result, err := route.run(path)
		if err != nil {
			t.Fatalf("%s: %v", route.name, err)
		}
		phase34AssertLocatedRoute(t, route.name, path, source, result)
	}

	debugPath := makeSource()
	debugResult, err := session.DebugMapCommandFile(debugPath, "")
	if err != nil {
		t.Fatal(err)
	}
	phase34AssertLocatedRoute(t, "debug-map", debugPath, source, debugResult)

	// Explain and query reuse their held single-file parse/check, while their
	// identity-bearing summary spans remain byte-only.
	identityPath := makeSource()
	checked := session.Check(source)
	if len(checked.Diagnostics) == 0 {
		t.Fatal("broken source produced no diagnostic ID for explain/query")
	}
	diagnosticID := checked.Diagnostics[0].ID
	explained, err := session.ExplainCommandFile(identityPath, diagnosticID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if explained.Presentation == nil || explained.Explain == nil || len(explained.Explain.Nodes) == 0 {
		t.Fatalf("explain result=%+v, want held presentation and cause summary", explained)
	}
	for _, node := range explained.Explain.Nodes {
		if node.Span != nil && (node.Span.Line != 0 || node.Span.Col != 0) {
			t.Fatalf("explain identity span gained source coordinates: %+v", node.Span)
		}
	}
	queried, err := session.QueryCommandFile(identityPath, diagnosticID, session.QueryOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if queried.Presentation == nil || queried.Query == nil || len(queried.Query.Facts) == 0 {
		t.Fatalf("query result=%+v, want held presentation and query facts", queried)
	}
	for _, fact := range queried.Query.Facts {
		if fact.Span != nil && (fact.Span.Line != 0 || fact.Span.Col != 0) {
			t.Fatalf("query identity span gained source coordinates: %+v", fact.Span)
		}
	}

	// Artifact/global failures keep zero spans and never borrow the unrelated
	// source path merely because a validation command also read source bytes.
	manifestPath := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(manifestPath, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	validation := session.ValidateEvidenceCommandFile(context.Background(), manifestPath, makeSource())
	phase34AssertUnlocatedRoute(t, "evidence validation", validation)

	summaryPath := filepath.Join(t.TempDir(), "summary.json")
	corePath := filepath.Join(t.TempDir(), "core.json")
	if err := os.WriteFile(summaryPath, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(corePath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	interfaceCheck, err := session.InterfaceCheckCommandFile(summaryPath, corePath)
	if err != nil {
		t.Fatal(err)
	}
	phase34AssertUnlocatedRoute(t, "interface check", interfaceCheck)

	corpusFailure := session.VerifyCorpusFile(context.Background(), filepath.Join(t.TempDir(), "missing-corpus"), native.Runner{})
	phase34AssertUnlocatedRoute(t, "verify corpus", corpusFailure)

	// Imported debug-map errors use the loader's captured module bytes and
	// relative label exactly once.
	entry := writeModuleFixture(t, map[string]string{
		"main.schway": "module app.main\n\nimport lib {\n  fn relay\n}\n\nexport {\n  fn main\n}\n\nfn main(n: U64) -> U64 {\n  let r = relay(n)\n  r\n}\n",
		"lib.schway":  "module lib\n\nexport {\n  fn relay\n}\n\nfn relay(n: U64) -> U64 {\n  let r = missing_name\n  r\n}\n",
	})
	reads := map[string]int{}
	restoreReads := modload.SetReadObserverForTest(func(module string) { reads[module]++ })
	imported, err := session.DebugMapCommandFile(entry, "")
	restoreReads()
	if err != nil || imported.Status != protocol.StatusInvalid || imported.Presentation == nil {
		t.Fatalf("imported debug-map result=%+v err=%v, want captured imported refusal", imported, err)
	}
	var problem diagnostic.Diagnostic
	found := false
	for _, candidate := range imported.Diagnostics {
		if candidate.Code == "name.unknown_place" {
			problem, found = candidate, true
			break
		}
	}
	if !found || problem.Unit != "lib" || reads[""] != 1 || reads["lib"] != 1 || len(reads) != 2 {
		t.Fatalf("imported debug-map diagnostics=%+v reads=%v, want one imported source read", imported.Diagnostics, reads)
	}
	if got := string(imported.Presentation.Sources["lib"].Bytes[problem.Primary.Start:problem.Primary.End]); got != "missing_name" || imported.Presentation.Sources["lib"].Label != "lib.schway" {
		t.Fatalf("imported debug-map snapshot=%+v, span selects %q", imported.Presentation.Sources["lib"], got)
	}
	human, err := protocol.Human(imported)
	if err != nil || !strings.Contains(human, "lib.schway:") || !strings.Contains(human, "missing_name") {
		t.Fatalf("imported debug-map source attribution missing: err=%v\n%s", err, human)
	}
}

func phase34AssertLocatedRoute(t *testing.T, route, path string, source []byte, result protocol.Result) {
	t.Helper()
	if result.Status != protocol.StatusInvalid || len(result.Diagnostics) == 0 || result.Presentation == nil {
		t.Fatalf("%s result=%+v, want source diagnostic with presentation", route, result)
	}
	captured, ok := result.Presentation.Sources[""]
	if !ok || captured.Label != path || !bytes.Equal(captured.Bytes, source) {
		t.Fatalf("%s presentation=%+v, want the held entry bytes", route, result.Presentation)
	}
	problem := result.Diagnostics[0]
	if problem.Primary.Line != 0 || problem.Primary.Col != 0 {
		t.Fatalf("%s mutated byte-authoritative diagnostic identity: %+v", route, problem.Primary)
	}
	jsonBytes, err := protocol.JSON(result)
	if err != nil {
		t.Fatal(err)
	}
	var wire struct {
		Diagnostics []struct {
			Primary diagnostic.Span `json:"primary_span"`
		} `json:"diagnostics"`
	}
	if err := json.Unmarshal(jsonBytes, &wire); err != nil || len(wire.Diagnostics) == 0 || wire.Diagnostics[0].Primary.Line != 3 || wire.Diagnostics[0].Primary.Col < 1 {
		t.Fatalf("%s JSON presentation=%s err=%v, want a source-derived line 3 location", route, jsonBytes, err)
	}
	human, err := protocol.Human(result)
	if err != nil || !strings.Contains(human, path+":3:") || !strings.Contains(human, problem.Code) || !strings.Contains(human, problem.ID) || !strings.Contains(human, "n + }") || !strings.Contains(human, "^") {
		t.Fatalf("%s human presentation err=%v\n%s", route, err, human)
	}
}

func phase34AssertUnlocatedRoute(t *testing.T, route string, result protocol.Result) {
	t.Helper()
	if len(result.Diagnostics) == 0 {
		t.Fatalf("%s result=%+v, want an explicit source-less diagnostic", route, result)
	}
	encoded, err := protocol.JSON(result)
	if err != nil {
		t.Fatal(err)
	}
	var wire struct {
		Diagnostics []struct {
			Primary struct {
				Line int `json:"line"`
				Col  int `json:"col"`
			} `json:"primary_span"`
		} `json:"diagnostics"`
	}
	if err := json.Unmarshal(encoded, &wire); err != nil || len(wire.Diagnostics) == 0 || wire.Diagnostics[0].Primary.Line != 0 || wire.Diagnostics[0].Primary.Col != 0 {
		t.Fatalf("%s fabricated a source location: %s err=%v", route, encoded, err)
	}
}

func phase34InstallFakeClang(t *testing.T) {
	t.Helper()
	directory := t.TempDir()
	clang := filepath.Join(directory, "clang")
	script := []byte("#!/bin/sh\ncase \"$1\" in\n  --version) printf 'clang fixture 1.0\\n' ;;\n  -dumpmachine) printf 'fixture-target\\n' ;;\n  *) exit 64 ;;\nesac\n")
	if err := os.WriteFile(clang, script, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", directory+string(os.PathListSeparator)+os.Getenv("PATH"))
}
