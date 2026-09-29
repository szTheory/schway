package measure

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
)

// fakeCommand returns a commandFactory that runs a fixed shell script
// (mirroring evidence's and cache's own fixture-command test seam) instead
// of spawning the real clang / sysctl binaries, so these tests are
// deterministic and hermetic.
func fakeCommand(t *testing.T, dir, script string) commandFactory {
	t.Helper()
	path := filepath.Join(dir, "fixture-probe.sh")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, path)
	}
}

func fixedFacts() MachineFacts {
	return MachineFacts{
		OS: "darwin", Arch: "arm64", CPUModel: "Apple M2",
		LogicalCores: 8, GoVersion: "go1.24.0", ClangVersion: "Apple clang version 15.0.0",
	}
}

func TestMachineFactFieldsAreClosed(t *testing.T) {
	got := FactFieldNames()
	want := []string{"os", "arch", "cpu_model", "logical_cores", "go_version", "cschway_version"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FactFieldNames() = %v, want %v", got, want)
	}

	// Reflect over MachineFacts's own json tags and assert they set-equal
	// FactFieldNames() exactly -- adding a seventh field without updating
	// the list fails this test.
	facts := reflect.TypeOf(MachineFacts{})
	tags := make([]string, 0, facts.NumField())
	for i := 0; i < facts.NumField(); i++ {
		tag := facts.Field(i).Tag.Get("json")
		tags = append(tags, strings.Split(tag, ",")[0])
	}
	sortedGot := append([]string(nil), tags...)
	sortedWant := append([]string(nil), want...)
	sort.Strings(sortedGot)
	sort.Strings(sortedWant)
	if !reflect.DeepEqual(sortedGot, sortedWant) {
		t.Fatalf("MachineFacts json tags = %v, want set-equal to %v", tags, want)
	}
}

func TestMachineIDIsDerivedFromDeclaredFacts(t *testing.T) {
	base := fixedFacts()
	id1 := MachineID(base)
	id2 := MachineID(base)
	if id1 != id2 {
		t.Fatalf("MachineID not stable across calls: %q vs %q", id1, id2)
	}
	if !strings.HasPrefix(id1, "machine:") {
		t.Fatalf("MachineID() = %q, want machine: prefix", id1)
	}
	hexPart := strings.TrimPrefix(id1, "machine:")
	if len(hexPart) != 12 {
		t.Fatalf("MachineID() hex part = %q (len %d), want 12 hex characters", hexPart, len(hexPart))
	}

	// Table-driven: changing any one fact changes the ID.
	cases := []struct {
		name   string
		mutate func(*MachineFacts)
	}{
		{"os", func(f *MachineFacts) { f.OS = "linux" }},
		{"arch", func(f *MachineFacts) { f.Arch = "amd64" }},
		{"cpu_model", func(f *MachineFacts) { f.CPUModel = "Intel i9" }},
		{"logical_cores", func(f *MachineFacts) { f.LogicalCores = 16 }},
		{"go_version", func(f *MachineFacts) { f.GoVersion = "go1.25.0" }},
		{"cschway_version", func(f *MachineFacts) { f.ClangVersion = "clang version 16.0.0" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mutated := fixedFacts()
			tc.mutate(&mutated)
			mutatedID := MachineID(mutated)
			if mutatedID == id1 {
				t.Fatalf("MachineID unchanged after mutating %s: %q", tc.name, mutatedID)
			}
		})
	}
}

// TestMachineIDExcludesHostFingerprints is T-06-MEASURE-01's structural and
// runtime backstop: a go/ast scan of machine.go asserting the source never
// references a host-fingerprint API, plus a runtime assertion that the
// live host's hostname/username/home-directory never appear as substrings
// of a probed MachineFacts's marshalled JSON.
func TestMachineIDExcludesHostFingerprints(t *testing.T) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "machine.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	forbidden := map[string]bool{
		"Hostname":    true, // os.Hostname
		"Interfaces":  true, // net.Interfaces
		"UserHomeDir": true, // os.UserHomeDir
		"Current":     true, // user.Current
	}
	ast.Inspect(file, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if forbidden[sel.Sel.Name] {
			t.Fatalf("machine.go references forbidden host-fingerprint selector %q", sel.Sel.Name)
		}
		return true
	})

	// Runtime assertion: a real MachineFacts encoding never contains this
	// host's hostname, username, or home directory as a substring.
	hostname, hostErr := os.Hostname()
	homeDir, homeErr := os.UserHomeDir()
	var username string
	if u, uerr := user.Current(); uerr == nil {
		username = u.Username
	}

	dir := t.TempDir()
	command := fakeCommand(t, dir, "#!/bin/sh\necho Apple clang version 15.0.0\n")
	facts, err := probeMachine(context.Background(), command)
	if err != nil {
		t.Fatal(err)
	}
	encoded := facts.OS + facts.Arch + facts.CPUModel + facts.GoVersion + facts.ClangVersion

	if hostErr == nil && hostname != "" && strings.Contains(encoded, hostname) {
		t.Fatalf("MachineFacts encoding leaks hostname %q", hostname)
	}
	if homeErr == nil && homeDir != "" && strings.Contains(encoded, homeDir) {
		t.Fatalf("MachineFacts encoding leaks home directory %q", homeDir)
	}
	if username != "" && strings.Contains(encoded, username) {
		t.Fatalf("MachineFacts encoding leaks username %q", username)
	}
}

func TestMachineProbeIsBounded(t *testing.T) {
	t.Run("timeout", func(t *testing.T) {
		dir := t.TempDir()
		command := fakeCommand(t, dir, "#!/bin/sh\nsleep 6\n")
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		_, err := probeMachine(ctx, command)
		var typed *Error
		if !errors.As(err, &typed) || typed.Code != "measure.probe_timeout" {
			t.Fatalf("probeMachine() error = %v, want measure.probe_timeout", err)
		}
	})

	t.Run("truncation", func(t *testing.T) {
		dir := t.TempDir()
		// Emit far more than MaxProbeBytes so the bounded writer reports
		// truncation rather than silently accepting unbounded output.
		script := "#!/bin/sh\nyes A | head -c 200000\n"
		command := fakeCommand(t, dir, script)
		_, err := probeMachine(context.Background(), command)
		var typed *Error
		if !errors.As(err, &typed) || typed.Code != "measure.probe_stdout_truncated" {
			t.Fatalf("probeMachine() error = %v, want measure.probe_stdout_truncated", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		dir := t.TempDir()
		command := fakeCommand(t, dir, "#!/bin/sh\necho Apple clang version 15.0.0\n")
		facts, err := probeMachine(context.Background(), command)
		if err != nil {
			t.Fatal(err)
		}
		if facts.OS != runtime.GOOS {
			t.Fatalf("facts.OS = %q, want %q", facts.OS, runtime.GOOS)
		}
		if facts.ClangVersion == "" {
			t.Fatal("facts.ClangVersion is empty, want probed value")
		}
	})
}

// TestMeasureImportsStayIndependent asserts measure imports nothing from
// internal/compiler/protocol, internal/compiler/session, or
// internal/compiler/diagnostic, matching the corevalidate/pathoracle/
// originvalidate import-boundary test shape already shipped in this tree.
func TestMeasureImportsStayIndependent(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	forbidden := []string{"/compiler/protocol", "/compiler/session", "/compiler/diagnostic"}
	fileSet := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		file, err := parser.ParseFile(fileSet, entry.Name(), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imported := range file.Imports {
			path := strings.Trim(imported.Path.Value, `"`)
			for _, forbid := range forbidden {
				if strings.HasSuffix(path, forbid) {
					t.Fatalf("%s imports %s, which measure must never depend on", entry.Name(), path)
				}
			}
		}
	}
}

func TestMachineProbeContextDeadline(t *testing.T) {
	// Sanity: a parent context with its own short deadline still surfaces
	// the typed timeout error rather than hanging.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	dir := t.TempDir()
	command := fakeCommand(t, dir, "#!/bin/sh\nsleep 2\n")
	_, err := probeMachine(ctx, command)
	if err == nil {
		t.Fatal("probeMachine() with an already-tight parent deadline returned nil error")
	}
}
