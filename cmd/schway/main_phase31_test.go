package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// phase31RouteProbe derives a copied-directory probe program from checksum.schway:
// the module header is renamed and main's tail binding is replaced by tail.
func phase31RouteProbe(t *testing.T, dir, helpers, tail string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "checksum.schway"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	const header = "module examples.checksum.checksum\n"
	const body = "  let sum = adler32(view)\n  sum\n}\n"
	if strings.Count(source, header) != 1 || strings.Count(source, body) != 1 {
		t.Fatal("checksum.schway no longer has the expected header and main tail")
	}
	source = strings.Replace(source, header, "module examples.checksum.adler_low_probe\n", 1)
	source = strings.Replace(source, body, "  let sum = adler32(view)\n"+tail+"}\n\n"+helpers, 1)
	path := filepath.Join(dir, "adler_low_probe.schway")
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func phase31RouteInput(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "in.bin")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestPhase31ChecksumPublicRoute is the Phase 31 tracer (SC1, D-05, D-06,
// D-16, D-17): a real file reaches the checksum through build --manifest and
// app run on native.
func TestPhase31ChecksumPublicRoute(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("env:clang toolchain unavailable on this host")
	}
	source := testsupport.ProjectPath("examples", "checksum", "checksum.schway")
	manifest := testsupport.ProjectPath("examples", "checksum", "file_bytes.bindings.json")
	artifact := filepath.Join(t.TempDir(), "checksum")
	code, stdout, stderr := captureLangRun(t, []string{"build", source, "--manifest", manifest, "--output", artifact})
	if code != 0 || stdout != "built "+artifact+"\n" || stderr != "" {
		t.Fatalf("checksum build code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	t.Run("wikipedia", func(t *testing.T) {
		code, stdout, stderr := captureLangRun(t, []string{"app", "run", artifact, "--", phase31RouteInput(t, "Wikipedia")})
		if code != 0 || stdout != "300286872\n" || stderr != "" {
			t.Fatalf("wikipedia code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
	})
	t.Run("missing", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "phase31-secret-segment", "absent.bin")
		code, stdout, stderr := captureLangRun(t, []string{"app", "run", artifact, "--", missing})
		if code != 65 || stdout != "" || stderr != "FileError.FileUnavailable\n" || strings.Contains(stderr, "phase31-secret-segment") {
			t.Fatalf("missing code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
	})
	t.Run("trailing-call-tempdir", func(t *testing.T) {
		dir := phase30CopyExampleDir(t)
		probe := phase31RouteProbe(t, dir, "fn low16(sum: U64) -> U64 {\n  var r = sum\n  r = r % 65536\n  r\n}\n", "  let low = low16(sum)\n  low\n")
		probeArtifact := filepath.Join(t.TempDir(), "probe")
		code, stdout, stderr := captureLangRun(t, []string{"build", probe, "--manifest", filepath.Join(dir, "file_bytes.bindings.json"), "--output", probeArtifact})
		if code != 0 || stdout != "built "+probeArtifact+"\n" || stderr != "" {
			t.Fatalf("probe build code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
		code, stdout, stderr = captureLangRun(t, []string{"app", "run", probeArtifact, "--", phase31RouteInput(t, "Wikipedia")})
		if code != 0 || stdout != "920\n" || stderr != "" {
			t.Fatalf("probe run code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
	})
	t.Run("second-trailing-call-refused", func(t *testing.T) {
		dir := phase30CopyExampleDir(t)
		probe := phase31RouteProbe(t, dir, "fn low16(sum: U64) -> U64 {\n  var r = sum\n  r = r % 65536\n  r\n}\n", "  let low = low16(sum)\n  let again = low16(low)\n  again\n")
		code, stdout, stderr := captureLangRun(t, []string{"check", probe})
		if code != 2 || !strings.Contains(stdout+stderr, "check.file_bytes_owner_shape_unsupported") {
			t.Fatalf("second trailing call code=%d stdout=%q stderr=%q; want exit 2 and check.file_bytes_owner_shape_unsupported", code, stdout, stderr)
		}
	})
	t.Run("mutated-library", func(t *testing.T) {
		dir := phase30CopyExampleDir(t)
		libraryPath := filepath.Join(dir, "adler32.schway")
		library, err := os.ReadFile(libraryPath)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(library), "var a = 1\n") != 1 {
			t.Fatal("library no longer has exactly one `var a = 1` line")
		}
		if err := os.WriteFile(libraryPath, []byte(strings.Replace(string(library), "var a = 1\n", "var a = 2\n", 1)), 0o600); err != nil {
			t.Fatal(err)
		}
		mutatedArtifact := filepath.Join(t.TempDir(), "checksum-mutated")
		code, stdout, stderr := captureLangRun(t, []string{"build", filepath.Join(dir, "checksum.schway"), "--manifest", filepath.Join(dir, "file_bytes.bindings.json"), "--output", mutatedArtifact})
		if code != 0 || stdout != "built "+mutatedArtifact+"\n" || stderr != "" {
			t.Fatalf("mutated build code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
		code, stdout, stderr = captureLangRun(t, []string{"app", "run", mutatedArtifact, "--", phase31RouteInput(t, "Wikipedia")})
		if code != 0 || stderr != "" {
			t.Fatalf("mutated run code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
		if stdout == "300286872\n" {
			t.Fatal("mutated library still printed the pinned value; the tracer answer is not coming from the library")
		}
	})
	const unavailable = "FileError.FileUnavailable\n"
	runArg := func(t *testing.T, arg string) (int, string, string) {
		t.Helper()
		return captureLangRun(t, []string{"app", "run", artifact, "--", arg})
	}
	for _, row := range testsupport.ChecksumOracleRows() {
		if row.App != "checksum" || row.Outcome != "returned" {
			continue
		}
		switch row.Name {
		case "empty", "0x00", "Wikipedia", "65536x0xFF":
		default:
			continue
		}
		name := map[string]string{"empty": "oracle-empty", "0x00": "oracle-zero", "Wikipedia": "oracle-wikipedia", "65536x0xFF": "oracle-ff65536"}[row.Name]
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "in.bin")
			if err := os.WriteFile(path, row.Input, 0o600); err != nil {
				t.Fatal(err)
			}
			want := strconv.FormatUint(row.Want, 10) + "\n"
			code, stdout, stderr := runArg(t, path)
			if code != 0 || stdout != want || stderr != "" {
				t.Fatalf("%s code=%d stdout=%q stderr=%q, want 0 %q empty", name, code, stdout, stderr, want)
			}
		})
	}
	t.Run("oversize-65537", func(t *testing.T) {
		var data []byte
		for _, row := range testsupport.ChecksumOracleRows() {
			if row.App == "checksum" && row.Outcome == "FileTooLarge" {
				data = row.Input
			}
		}
		path := filepath.Join(t.TempDir(), testsupport.ChecksumSecretSegment, "big.bin")
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
		code, stdout, stderr := runArg(t, path)
		if code != 65 || stdout != "" || stderr != "FileError.FileTooLarge\n" {
			t.Fatalf("oversize code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
	})
	t.Run("leading-dash", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "-leading.bin"), []byte("Wikipedia"), 0o600); err != nil {
			t.Fatal(err)
		}
		t.Chdir(dir)
		code, stdout, stderr := runArg(t, "-leading.bin")
		if code != 0 || stdout != "300286872\n" || stderr != "" {
			t.Fatalf("leading-dash code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
	})
	t.Run("path-4096", func(t *testing.T) {
		code, stdout, stderr := runArg(t, strings.Repeat("a", 4096))
		if code != 65 || stdout != "" || stderr != unavailable {
			t.Fatalf("path-4096 code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
	})
	t.Run("path-4097", func(t *testing.T) {
		code, stdout, stderr := runArg(t, strings.Repeat("a", 4097))
		if code != 65 || stdout != "" || stderr != phase31OverLongStderr || strings.Contains(stderr, "aaaa") {
			t.Fatalf("path-4097 code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
	})
	t.Run("empty-argument", func(t *testing.T) {
		// Named silent-by-design exception: the generated C main returns 65
		// before any output (cgen_program.go:1064).
		code, stdout, stderr := runArg(t, "")
		if code != 65 || stdout != "" || stderr != "" {
			t.Fatalf("empty-argument code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
	})
	t.Run("path-free-stderr", func(t *testing.T) {
		dir := t.TempDir()
		secretDir := filepath.Join(dir, testsupport.ChecksumSecretSegment)
		if err := os.MkdirAll(secretDir, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(filepath.Join(secretDir, "dir.bin"), 0o700); err != nil {
			t.Fatal(err)
		}
		t.Chdir(dir)
		for _, arg := range []string{
			filepath.Join(secretDir, "absent.bin"),
			filepath.Join(secretDir, "dir.bin"),
			"-leading-absent.bin",
			strings.Repeat("a", 4097),
		} {
			code, stdout, stderr := runArg(t, arg)
			if code != 65 || stdout != "" || stderr == "" {
				t.Fatalf("arg %.20q code=%d stdout=%q stderr=%q", arg, code, stdout, stderr)
			}
			if strings.Contains(stderr, testsupport.ChecksumSecret) || strings.Contains(stderr, "-leading") || strings.Contains(stderr, "aaaa") {
				t.Fatalf("arg %.20q stderr names the argument: %q", arg, stderr)
			}
		}
	})
}

// phase31OverLongStderr is the fixed host refusal for an over-length argument
// (cmd/schway appRunExitCode); it never contains the argument.
const phase31OverLongStderr = "schway app run: native.input_too_long: argument is 4097 bytes; limit is 4096\n"
