package session_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const phase29FixtureManifestPath = "../../../examples/phase29/fixture_bytes.bindings.json"

// TestPhase29NativeFixtureVectors builds the real fixture program once with its
// binding manifest and runs the single artifact for every pinned case id (D-15,
// D-16): one artifact serves every vector, and an unknown case exits 65 with
// nothing printed.
func TestPhase29NativeFixtureVectors(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang unavailable on this host")
	}
	artifact := filepath.Join(t.TempDir(), "adler32_fixture")
	_, diagnostics, err := session.BuildApplicationFile(context.Background(), phase29FixturePath, artifact, native.DefaultRunner(), phase29FixtureManifestPath)
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("BuildApplicationFile diagnostics=%+v err=%v", diagnostics, err)
	}
	for caseID := uint64(0); caseID <= 9; caseID++ {
		want, ok := testsupport.Phase29FixtureExpected(caseID)
		if !ok {
			t.Fatalf("case %d has no pinned checksum", caseID)
		}
		var stdout, stderr bytes.Buffer
		outcome, err := native.RunApplication(context.Background(), artifact, strconv.FormatUint(caseID, 10), &stdout, &stderr)
		expected := strconv.FormatUint(want, 10) + "\n"
		if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 0 || stdout.String() != expected || stderr.Len() != 0 {
			t.Fatalf("case %d outcome=%+v stdout=%q stderr=%q err=%v, want %q exit 0", caseID, outcome, stdout.String(), stderr.String(), err, expected)
		}
	}
	var stdout, stderr bytes.Buffer
	outcome, err := native.RunApplication(context.Background(), artifact, "10", &stdout, &stderr)
	if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 65 || stdout.Len() != 0 {
		t.Fatalf("unknown case outcome=%+v stdout=%q stderr=%q err=%v, want exit 65 and empty stdout", outcome, stdout.String(), stderr.String(), err)
	}
}

// TestPhase29FixtureAdmission pins the manifest obligations of the fixture
// owner (T-29-21): BuildApplication refuses the fixture without a manifest,
// with a manifest that omits one checked symbol, and with an extra symbol,
// each with the local-owner error class.
func TestPhase29FixtureAdmission(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang unavailable on this host")
	}
	source, err := os.ReadFile(phase29FixturePath)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := os.ReadFile(phase29FixtureManifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(valid, &manifest); err != nil {
		t.Fatal(err)
	}
	symbols := manifest["symbols"].([]any)
	stage := func(t *testing.T, symbols []any) string {
		t.Helper()
		dir := t.TempDir()
		for _, name := range []string{"adapter.c", "adapter.h"} {
			data, err := os.ReadFile(filepath.Join(filepath.Dir(phase29FixtureManifestPath), name))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		forged := map[string]any{}
		for key, value := range manifest {
			forged[key] = value
		}
		forged["symbols"] = symbols
		data, err := json.Marshal(forged)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, "forged.bindings.json")
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	extra := map[string]any{"name": "schway_fixture_bytes_extra", "header": "adapter.h", "function_type": "schway_fixture_bytes_release_fn"}
	cases := []struct {
		name     string
		manifest func(t *testing.T) []string
		want     string
	}{
		{"no manifest", func(*testing.T) []string { return nil }, "local-owner application requires one explicit binding manifest"},
		{"missing view symbol", func(t *testing.T) []string {
			var kept []any
			for _, symbol := range symbols {
				if symbol.(map[string]any)["name"] != "schway_fixture_bytes_view" {
					kept = append(kept, symbol)
				}
			}
			return []string{stage(t, kept)}
		}, "local binding manifest symbols do not exactly cover checked foreign operations"},
		{"extra symbol", func(t *testing.T) []string {
			return []string{stage(t, append(append([]any{}, symbols...), extra))}
		}, "local binding manifest symbols do not exactly cover checked foreign operations"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, diagnostics, err := session.BuildApplication(context.Background(), source, filepath.Join(t.TempDir(), "fixture"), native.DefaultRunner(), tc.manifest(t)...)
			if len(diagnostics) != 0 || err == nil || err.Error() != tc.want {
				t.Fatalf("diagnostics=%+v err=%v, want error %q", diagnostics, err, tc.want)
			}
		})
	}
}
