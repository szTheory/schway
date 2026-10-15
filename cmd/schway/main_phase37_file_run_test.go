package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func phase37RunFixture() (string, string) {
	return testsupport.ProjectPath("examples", "crc32", "crc32.schway"),
		testsupport.ProjectPath("examples", "crc32", "file_bytes.bindings.json")
}

func phase37WriteInput(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("123456789"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func phase37RunArgs(engine, source, manifest, path string) []string {
	args := []string{"run", "--engine=" + engine, source}
	if manifest != "" {
		args = append(args, "--manifest", manifest)
	}
	return append(args, "--", path)
}

func TestPhase37PublicFileRunTracer(t *testing.T) {
	source, manifest := phase37RunFixture()
	inputs := t.TempDir()
	path := phase37WriteInput(t, inputs, "crc input.bin")

	t.Run("interpreter-with-clang-unavailable", func(t *testing.T) {
		bin := t.TempDir()
		sentinel := filepath.Join(t.TempDir(), "clang-was-invoked")
		clang := filepath.Join(bin, "clang")
		script := "#!/bin/sh\n: > " + sentinel + "\nexit 99\n"
		if err := os.WriteFile(clang, []byte(script), 0o700); err != nil {
			t.Fatal(err)
		}
		t.Setenv("PATH", bin)
		code, stdout, stderr := captureLangRun(t, phase37RunArgs("interpreter", source, "", path))
		if code != 0 || stdout != "3421780262\n" || stderr != "" {
			t.Fatalf("interpreter file run = (%d,%q,%q)", code, stdout, stderr)
		}
		if _, err := os.Stat(sentinel); !os.IsNotExist(err) {
			t.Fatalf("interpreter invoked Clang; sentinel stat err=%v", err)
		}
	})

	t.Run("native-private-one-shot", func(t *testing.T) {
		tempRoot := t.TempDir()
		t.Setenv("TMPDIR", tempRoot)
		code, stdout, stderr := captureLangRun(t, phase37RunArgs("native", source, manifest, path))
		if code != 0 || stdout != "3421780262\n" || stderr != "" {
			t.Fatalf("native file run = (%d,%q,%q)", code, stdout, stderr)
		}
		entries, err := os.ReadDir(tempRoot)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 0 {
			t.Fatalf("native private build artifacts leaked: %v", entries)
		}
	})
}

func TestPhase37FileRunGrammar(t *testing.T) {
	source, manifest := phase37RunFixture()
	usage := "usage: schway run --engine={interpreter|native} SOURCE [--manifest MANIFEST] -- PATH\n"
	for _, test := range []struct {
		name string
		args []string
	}{
		{name: "missing separator", args: []string{"run", "--engine=interpreter", source}},
		{name: "missing engine", args: []string{"run", source, "--", "input"}},
		{name: "duplicate engine", args: []string{"run", "--engine=interpreter", "--engine=native", source, "--manifest", manifest, "--", "input"}},
		{name: "unknown engine", args: []string{"run", "--engine=llvm", source, "--", "input"}},
		{name: "missing source", args: []string{"run", "--engine=interpreter", "--", "input"}},
		{name: "extra source", args: []string{"run", "--engine=interpreter", source, "other.schway", "--", "input"}},
		{name: "missing path", args: []string{"run", "--engine=interpreter", source, "--"}},
		{name: "extra path", args: []string{"run", "--engine=interpreter", source, "--", "input", "extra"}},
		{name: "native manifest required", args: []string{"run", "--engine=native", source, "--", "input"}},
		{name: "manifest empty", args: []string{"run", "--engine=native", source, "--manifest", "", "--", "input"}},
		{name: "duplicate manifest", args: []string{"run", "--engine=native", source, "--manifest", manifest, "--manifest", manifest, "--", "input"}},
		{name: "interpreter manifest forbidden", args: []string{"run", "--engine=interpreter", source, "--manifest", manifest, "--", "input"}},
		{name: "report forbidden", args: []string{"run", "--engine=native", source, "--manifest", manifest, "--report", "report", "--", "input"}},
		{name: "evidence forbidden", args: []string{"run", "--engine=native", source, "--manifest", manifest, "--evidence=events", "--", "input"}},
		{name: "global presentation forbidden", args: []string{"--json", "run", "--engine=interpreter", source, "--", "input"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			code, stdout, stderr := captureLangRun(t, test.args)
			if code != 64 || stdout != "" || stderr != usage {
				t.Fatalf("grammar result = (%d,%q,%q), want (64,empty,%q)", code, stdout, stderr, usage)
			}
		})
	}
}

func TestPhase37OpaqueFilePaths(t *testing.T) {
	source, manifest := phase37RunFixture()
	dir := t.TempDir()
	path := phase37WriteInput(t, dir, "--json ;$(literal) 雪.bin")
	for _, engine := range []string{"interpreter", "native"} {
		t.Run(engine, func(t *testing.T) {
			manifestArg := ""
			if engine == "native" {
				manifestArg = manifest
			}
			code, stdout, stderr := captureLangRun(t, phase37RunArgs(engine, source, manifestArg, path))
			if code != 0 || stdout != "3421780262\n" || stderr != "" {
				t.Fatalf("opaque path result = (%d,%q,%q)", code, stdout, stderr)
			}
		})
	}

	separatorPath := phase37WriteInput(t, dir, "--")
	code, stdout, stderr := captureLangRun(t, phase37RunArgs("interpreter", source, "", separatorPath))
	if code != 0 || stdout != "3421780262\n" || stderr != "" {
		t.Fatalf("literal separator path result = (%d,%q,%q)", code, stdout, stderr)
	}
}

func TestPhase37FileRunInputBoundary(t *testing.T) {
	source, manifest := phase37RunFixture()
	for _, engine := range []string{"interpreter", "native"} {
		manifestArg := ""
		if engine == "native" {
			manifestArg = manifest
		}
		t.Run(engine, func(t *testing.T) {
			for _, test := range []struct {
				name, path, wantStderr string
				wantCode               int
			}{
				{name: "empty", path: "", wantCode: 65, wantStderr: ""},
				{name: "4096 reaches file adapter", path: strings.Repeat("x", 4096), wantCode: 65, wantStderr: "FileError.FileUnavailable\n"},
				{name: "4097 refused privately", path: strings.Repeat("x", 4097), wantCode: 65, wantStderr: "schway run: input exceeds 4096 bytes\n"},
				{name: "NUL refused privately", path: "secret\x00path", wantCode: 65, wantStderr: "schway run: input contains NUL\n"},
			} {
				t.Run(test.name, func(t *testing.T) {
					code, stdout, stderr := captureLangRun(t, phase37RunArgs(engine, source, manifestArg, test.path))
					if code != test.wantCode || stdout != "" || stderr != test.wantStderr {
						t.Fatalf("boundary result = (%d,%q,%q), want (%d,empty,%q)", code, stdout, stderr, test.wantCode, test.wantStderr)
					}
				})
			}
		})
	}
}
