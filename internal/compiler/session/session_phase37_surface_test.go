package session_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/surfacemanifest"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const (
	phase37InterpreterFileRun   = "run --engine=interpreter SOURCE -- PATH"
	phase37NativeFileRun        = "run --engine=native SOURCE --manifest MANIFEST -- PATH"
	phase37SurfaceWitnessSchema = "phase37.surface-witnesses/1"
)

func phase37ReadSurfaceWitnessAdditions(t *testing.T) phase34SurfaceWitnessDocument {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase37", "surface-witnesses.json"))
	if err != nil {
		t.Fatalf("read additive Phase 37 surface witnesses: %v", err)
	}
	var document phase34SurfaceWitnessDocument
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("decode additive Phase 37 surface witnesses: %v", err)
	}
	if document.Schema != phase37SurfaceWitnessSchema || document.Boundary == "" || len(document.Witnesses) == 0 || len(document.Limits) == 0 {
		t.Fatalf("invalid Phase 37 surface witness header: schema=%q boundary=%q witnesses=%d limits=%d", document.Schema, document.Boundary, len(document.Witnesses), len(document.Limits))
	}
	return document
}

func phase37SurfaceLimit(manifest surfacemanifest.Manifest, id string) (surfacemanifest.Limit, bool) {
	for _, item := range manifest.Limits {
		if item.ID == id {
			return item, true
		}
	}
	return surfacemanifest.Limit{}, false
}

func phase37SurfaceType(manifest surfacemanifest.Manifest, id string) (surfacemanifest.TypeEntry, bool) {
	for _, item := range manifest.Types {
		if item.ID == id {
			return item, true
		}
	}
	return surfacemanifest.TypeEntry{}, false
}

func phase37SurfaceEntry(manifest surfacemanifest.Manifest, id string) (surfacemanifest.Entry, bool) {
	for _, item := range manifest.Restrictions {
		if item.ID == id {
			return item, true
		}
	}
	return surfacemanifest.Entry{}, false
}

func phase37SurfaceWireProblems(encoded []byte) []string {
	var wire map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &wire); err != nil {
		return []string{fmt.Sprintf("surface JSON is invalid: %v", err)}
	}
	want := []string{"builtins", "calls", "coordinates", "limits", "operators", "refusals", "restrictions", "schema", "statements", "types"}
	got := make([]string, 0, len(wire))
	for key := range wire {
		got = append(got, key)
	}
	sort.Strings(got)
	if !reflectPhase37Strings(got, want) {
		return []string{fmt.Sprintf("surface wire fields = %v, want exactly %v", got, want)}
	}
	return nil
}

func reflectPhase37Strings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func phase37PathRoutesProblems(routes []string) []string {
	want := []string{phase37InterpreterFileRun, phase37NativeFileRun}
	var problems []string
	for _, route := range want {
		if !phase34ContainsProblem(routes, route) {
			problems = append(problems, "missing_path_consumer "+route)
		}
	}
	for _, route := range routes {
		if !phase34ContainsProblem(want, route) {
			problems = append(problems, "unsupported_path_consumer "+route)
		}
	}
	sort.Strings(problems)
	return problems
}

func phase37FileBytesProblems(entry surfacemanifest.TypeEntry) []string {
	if entry.ID != "file_bytes" || len(entry.Constraints) != 1 || !strings.Contains(entry.Constraints[0], "checked binding configuration") || !strings.Contains(entry.Constraints[0], "no universal language-wide byte cap") {
		return []string{"FileBytes facts overstate or omit route-specific cap ownership"}
	}
	return nil
}

func phase37FileRouteRestrictionProblems(entry surfacemanifest.Entry) []string {
	wantRoutes := []string{phase37InterpreterFileRun, phase37NativeFileRun}
	if entry.ID != "source_file_run_input" || entry.Syntax != "run --engine={interpreter|native} SOURCE -- PATH" || !reflectPhase37Strings(entry.Contexts, wantRoutes) || len(entry.Constraints) != 1 || !strings.Contains(entry.Constraints[0], "opaque") || !strings.Contains(entry.Constraints[0], "empty") || !strings.Contains(entry.Constraints[0], "NUL") || !strings.Contains(entry.Constraints[0], "4096 bytes") {
		return []string{"source-file run PathToken grammar or input restrictions changed"}
	}
	return nil
}

func phase37CRCInputConsumers(t *testing.T) []string {
	t.Helper()
	root := testsupport.ProjectPath("examples", "crc32")
	var consumers []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".schway" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(data)
		if strings.Contains(text, "schway_file_bytes_acquire(path: PathToken)") && strings.Contains(text, "schway_file_bytes_view(owner)") && strings.Contains(text, "crc32(view)") {
			relative, err := filepath.Rel(testsupport.ProjectPath(), path)
			if err != nil {
				return err
			}
			consumers = append(consumers, filepath.ToSlash(relative))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(consumers)
	return consumers
}

func phase37CRCConsumerProblems(consumers []string) []string {
	want := []string{"examples/crc32/crc32.schway", "examples/crc32/crc_low.schway"}
	var problems []string
	for _, path := range want {
		if !phase34ContainsProblem(consumers, path) {
			problems = append(problems, "missing_crc_file_consumer "+path)
		}
	}
	for _, path := range consumers {
		if !phase34ContainsProblem(want, path) {
			problems = append(problems, "unsupported_crc_file_consumer "+path)
		}
	}
	sort.Strings(problems)
	return problems
}

func TestPhase37FileRouteSurfaceFacts(t *testing.T) {
	manifest := surfacemanifest.Current()
	encoded, err := surfacemanifest.JSON(manifest)
	if err != nil {
		t.Fatal(err)
	}
	again, err := surfacemanifest.JSON(surfacemanifest.Current())
	if err != nil || !bytes.Equal(encoded, again) {
		t.Fatalf("surface serialization is not deterministic: err=%v", err)
	}
	if problems := phase37SurfaceWireProblems(encoded); len(problems) != 0 {
		t.Errorf("public surface wire contract: %v", problems)
	}
	var wireSchema string
	var wire map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &wire); err != nil || json.Unmarshal(wire["schema"], &wireSchema) != nil || wireSchema != "schway.surface/1" {
		t.Errorf("actual public surface schema = %q, decode error=%v", wireSchema, err)
	}
	if got := interp.FileBytesLimit; got != 65536 {
		t.Errorf("fixed interpreter FileBytes adapter cap = %d, want independently pinned 65536", got)
	}
	if item, ok := phase37SurfaceLimit(manifest, "application.argument_bytes"); !ok || item.Value != 4096 || !reflectPhase37Strings(item.Routes, []string{"app run"}) {
		t.Errorf("existing app run argument limit changed: %+v found=%t", item, ok)
	}
	if item, ok := phase37SurfaceLimit(manifest, "file_run.path_bytes"); !ok {
		t.Error("surface omits the source-file run PathToken argument limit")
	} else {
		problems := phase37PathRoutesProblems(item.Routes)
		if item.Category != "source-file run input" || item.Scope != "one opaque PathToken supplied after the file-run separator" || item.Unit != "bytes" || item.Value != 4096 || len(item.Meaning) == 0 || len(problems) != 0 {
			t.Errorf("source-file run PathToken limit facts = %+v; route findings=%v", item, problems)
		}
	}
	if item, ok := phase37SurfaceLimit(manifest, "interpreter.filebytes_bytes"); !ok {
		t.Error("surface omits the fixed interpreter FileBytes adapter limit")
	} else if item.Category != "interpreter file input" || item.Scope != "one regular file read by the fixed interpreter FileBytes adapter" || item.Unit != "bytes" || item.Value != 65536 || !reflectPhase37Strings(item.Routes, []string{phase37InterpreterFileRun}) || !strings.Contains(item.Meaning, "native binding-selected caps remain independent") || strings.Contains(item.Meaning, "universal FileBytes cap") {
		t.Errorf("interpreter FileBytes limit facts = %+v", item)
	}
	fileBytes, ok := phase37SurfaceType(manifest, "file_bytes")
	if !ok {
		t.Error("surface omits FileBytes type facts")
	} else if problems := phase37FileBytesProblems(fileBytes); len(problems) != 0 {
		t.Errorf("FileBytes facts overstate or omit route-specific cap ownership: %v; entry=%+v", problems, fileBytes)
	}
	pathRule, ok := phase37SurfaceEntry(manifest, "source_file_run_input")
	if !ok {
		t.Error("surface omits the source-file run PathToken restriction")
	} else if problems := phase37FileRouteRestrictionProblems(pathRule); len(problems) != 0 {
		t.Errorf("PathToken input boundary facts are incomplete: %v; entry=%+v", problems, pathRule)
	}
	consumers := phase37CRCInputConsumers(t)
	if problems := phase37CRCConsumerProblems(consumers); len(problems) != 0 {
		t.Errorf("independently enumerated CRC file consumers: %v (got %v)", problems, consumers)
	}
	document := phase34ReadSurfaceWitnesses(t)
	observed := phase34ObserveSurfaceWitnesses(t, document.Witnesses)
	if problems := phase34FullSurfaceProblems(manifest, document, observed); len(problems) != 0 {
		t.Errorf("independent surface witnesses differ from public facts: %v", problems)
	}

	t.Run("negative controls", func(t *testing.T) {
		validRoutes := []string{phase37InterpreterFileRun, phase37NativeFileRun}
		if problems := phase37PathRoutesProblems(validRoutes); len(problems) != 0 {
			t.Fatalf("valid path-route census findings: %v", problems)
		}
		missingRoute := []string{phase37InterpreterFileRun}
		if problems := phase37PathRoutesProblems(missingRoute); len(problems) == 0 {
			t.Fatal("missing path consumer passed the route census")
		}
		if problems := phase37CRCConsumerProblems([]string{"examples/crc32/crc32.schway"}); len(problems) == 0 {
			t.Fatal("missing CRC entry passed the independent consumer census")
		}
		var wire map[string]json.RawMessage
		if err := json.Unmarshal(encoded, &wire); err != nil {
			t.Fatal(err)
		}
		wire["unplanned"] = json.RawMessage(`{}`)
		mutated, err := json.Marshal(wire)
		if err != nil {
			t.Fatal(err)
		}
		if problems := phase37SurfaceWireProblems(mutated); len(problems) == 0 {
			t.Fatal("extra surface field passed the ten-field wire control")
		}
		universal := *fileBytesTypeCopy(fileBytes)
		universal.Constraints = []string{"All FileBytes values have one universal 65536-byte language-wide limit."}
		if problems := phase37FileBytesProblems(universal); len(problems) == 0 {
			t.Fatal("universal FileBytes cap passed the route-scope control")
		}
		mutatedRoutes := []string{phase37InterpreterFileRun}
		if problems := phase37PathRoutesProblems(mutatedRoutes); len(problems) == 0 {
			t.Fatal("missing native path consumer passed the route-scope control")
		}
		mutatedRule := pathRule
		mutatedRule.Constraints = []string{"All PathToken inputs use a universal cap."}
		if problems := phase37FileRouteRestrictionProblems(mutatedRule); len(problems) == 0 {
			t.Fatal("wrong PathToken restriction passed the route-scope control")
		}
	})
}

func fileBytesTypeCopy(entry surfacemanifest.TypeEntry) *surfacemanifest.TypeEntry {
	copy := entry
	copy.Constraints = append([]string(nil), entry.Constraints...)
	return &copy
}
