package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func phase37CloneProgram(t *testing.T, program core.Program) core.Program {
	t.Helper()
	data, err := json.Marshal(program)
	if err != nil {
		t.Fatal(err)
	}
	var clone core.Program
	if err := json.Unmarshal(data, &clone); err != nil {
		t.Fatal(err)
	}
	return clone
}

func phase37MainFunction(program *core.Program) *core.Function {
	for index := range program.Functions {
		if program.Functions[index].Name == "main" && program.Functions[index].Module == "" {
			return &program.Functions[index]
		}
	}
	return nil
}

func phase37WriteFile(t *testing.T, root, relative string, data []byte) string {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func phase37CRCSourceTree(t *testing.T, root string) (string, []byte) {
	t.Helper()
	entryBytes, err := os.ReadFile(testsupport.ProjectPath("examples", "crc32", "crc32.schway"))
	if err != nil {
		t.Fatal(err)
	}
	libraryBytes, err := os.ReadFile(testsupport.ProjectPath("examples", "crc32", "lib", "crc32.schway"))
	if err != nil {
		t.Fatal(err)
	}
	entryPath := phase37WriteFile(t, root, "main.schway", entryBytes)
	phase37WriteFile(t, root, "lib/crc32.schway", libraryBytes)
	return entryPath, entryBytes
}

func TestPhase37FileApplicationAuthority(t *testing.T) {
	sourcePath := testsupport.ProjectPath("examples", "crc32", "crc32.schway")
	checked, _, err := checkSourcePath(sourcePath)
	if err != nil || len(checked.Diagnostics) != 0 {
		t.Fatalf("checked CRC fixture: diagnostics=%+v err=%v", checked.Diagnostics, err)
	}
	main := phase37MainFunction(&checked.Program)
	if main == nil || main.Linear == nil || len(main.Linear.Operations) < 5 {
		t.Fatalf("CRC fixture did not expose the expected raw FileBytes entry: %+v", main)
	}
	if _, _, diagnostics, err := resolveEntry(checked.Program); err != nil || len(diagnostics) != 0 {
		t.Fatalf("resolve CRC entry: diagnostics=%+v err=%v", diagnostics, err)
	}

	t.Run("raw-core-cannot-forge-file-application-authority", func(t *testing.T) {
		mutations := []struct {
			name   string
			mutate func(*core.Program)
		}{
			{name: "wrong foreign policy", mutate: func(program *core.Program) {
				phase37MainFunction(program).Linear.Operations[0].Foreign.Unwind = "possible"
			}},
			{name: "wrong ABI", mutate: func(program *core.Program) {
				phase37MainFunction(program).Linear.Operations[0].Foreign.ABIType = "forged_file_reader"
			}},
			{name: "wrong failure alternative order", mutate: func(program *core.Program) {
				for index := range program.DataTypes {
					if program.DataTypes[index].Name == "FileError" {
						program.DataTypes[index].Alternatives[0], program.DataTypes[index].Alternatives[1] = program.DataTypes[index].Alternatives[1], program.DataTypes[index].Alternatives[0]
					}
				}
			}},
			{name: "wrong release pairing", mutate: func(program *core.Program) {
				phase37MainFunction(program).Linear.Operations[3].ReleasesOperationID = "forged-acquire"
			}},
			{name: "extra acquire", mutate: func(program *core.Program) {
				function := phase37MainFunction(program)
				function.Linear.Operations = append(function.Linear.Operations, function.Linear.Operations[0])
			}},
			{name: "extra view", mutate: func(program *core.Program) {
				function := phase37MainFunction(program)
				function.Linear.Operations = append(function.Linear.Operations, function.Linear.Operations[1])
			}},
			{name: "forged entry shape", mutate: func(program *core.Program) {
				phase37MainFunction(program).Parameter.Type = "U64"
			}},
			{name: "borrowed view escape", mutate: func(program *core.Program) {
				function := phase37MainFunction(program)
				function.Linear.Operations[len(function.Linear.Operations)-1].SourceID = function.Linear.Operations[1].TargetID
			}},
		}
		for _, test := range mutations {
			t.Run(test.name, func(t *testing.T) {
				program := phase37CloneProgram(t, checked.Program)
				test.mutate(&program)
				result, err := interp.RunFileApplication(program, "main", filepath.Join(t.TempDir(), "must-not-be-read"))
				if err == nil {
					t.Fatalf("raw core reached the file reader: outcome=%+v output=%q", result.Execution.Outcome, result.Output)
				}
			})
		}
	})

	t.Run("whole-merged-program-admitted-before-pruning", func(t *testing.T) {
		root := t.TempDir()
		entryPath, entryBytes := phase37CRCSourceTree(t, root)
		entry := strings.Replace(string(entryBytes), "\nexport {\n", "\nimport lib.bad {\n  fn bad\n}\n\nexport {\n", 1)
		if entry == string(entryBytes) {
			t.Fatal("could not add the unused authority neighbor import")
		}
		if err := os.WriteFile(entryPath, []byte(entry), 0o600); err != nil {
			t.Fatal(err)
		}
		badNeighbor := []byte(`module lib.bad

export {
  fn bad
}

fn bad(buffer: Buffer) -> Buffer {
  let view = borrow mut buffer
  let reviewed = borrow view
  view
}
`)
		phase37WriteFile(t, root, "lib/bad.schway", badNeighbor)

		loaded, _, err := checkSourcePath(entryPath)
		if err != nil || len(loaded.Diagnostics) != 0 {
			t.Fatalf("checker rejected the unused neighbor before independent admission: diagnostics=%+v err=%v", loaded.Diagnostics, err)
		}
		if _, found := findFunctionForPhase37(loaded.Program, "bad"); !found {
			t.Fatal("merged checked program omitted the imported bad function")
		}
		_, pruned, entryDiagnostics, err := resolveEntry(loaded.Program)
		if err != nil || len(entryDiagnostics) != 0 {
			t.Fatalf("resolve merged entry: diagnostics=%+v err=%v", entryDiagnostics, err)
		}
		if _, found := findFunctionForPhase37(pruned, "bad"); found {
			t.Fatal("fixture control failed: invalid imported function remains reachable after pruning")
		}
		originalRunner := runAdmittedFileApplication
		readerEntries := 0
		runAdmittedFileApplication = func(program core.Program, functionName, input string) (interp.ApplicationResult, error) {
			readerEntries++
			return interp.ApplicationResult{
				Execution: interp.Execution{Outcome: interp.Outcome{Kind: "returned", Value: "3421780262"}},
				Output:    []byte("3421780262\n"),
			}, nil
		}
		t.Cleanup(func() { runAdmittedFileApplication = originalRunner })
		if _, diagnostics, err := RunInterpreterFileApplication(sourcePath, "reader-positive-control"); err != nil || len(diagnostics) != 0 {
			t.Fatalf("valid admitted reader control: diagnostics=%+v err=%v", diagnostics, err)
		}
		if readerEntries != 1 {
			t.Fatalf("reader sentinel control reached %d times, want one", readerEntries)
		}
		readerEntries = 0
		_, admissionDiagnostics := admitProgram(loaded)
		if len(admissionDiagnostics) != 1 || admissionDiagnostics[0].Code != "core.origin_omitted" {
			t.Fatalf("whole-program admission diagnostics=%+v, want core.origin_omitted", admissionDiagnostics)
		}
		_, diagnostics, err := RunInterpreterFileApplication(entryPath, filepath.Join(root, "must-not-be-read"))
		if err != nil || len(diagnostics) != 1 || diagnostics[0].Code != "core.origin_omitted" {
			t.Fatalf("file application did not refuse unused neighbor before execution: diagnostics=%+v err=%v", diagnostics, err)
		}
		if readerEntries != 0 {
			t.Fatalf("invalid unused imported function reached the fixed file reader %d times; want zero", readerEntries)
		}
	})

	t.Run("module-root-traversal-and-symlink-refused", func(t *testing.T) {
		outside := t.TempDir()
		outsideSource := phase37WriteFile(t, outside, "crc32.schway", []byte("module lib.crc32\n"))
		entryBytes, err := os.ReadFile(testsupport.ProjectPath("examples", "crc32", "crc32.schway"))
		if err != nil {
			t.Fatal(err)
		}
		for _, kind := range []string{"directory", "file"} {
			t.Run(kind, func(t *testing.T) {
				root := t.TempDir()
				entryPath := phase37WriteFile(t, root, "main.schway", entryBytes)
				link := filepath.Join(root, "lib")
				target := outside
				if kind == "file" {
					if err := os.MkdirAll(link, 0o755); err != nil {
						t.Fatal(err)
					}
					link = filepath.Join(link, "crc32.schway")
					target = outsideSource
				}
				if err := os.Symlink(target, link); err != nil {
					t.Fatalf("create module %s symlink on supported host: %v", kind, err)
				}
				_, diagnostics, err := RunInterpreterFileApplication(entryPath, filepath.Join(root, "must-not-be-read"))
				if err != nil || len(diagnostics) == 0 || diagnostics[0].Code != "module.symlink" {
					t.Fatalf("module %s symlink refusal: diagnostics=%+v err=%v", kind, diagnostics, err)
				}
			})
		}
	})
}

func findFunctionForPhase37(program core.Program, name string) (core.Function, bool) {
	for _, function := range program.Functions {
		if function.Name == name {
			return function, true
		}
	}
	return core.Function{}, false
}
