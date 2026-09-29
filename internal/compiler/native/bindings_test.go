package native

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const phase22BindingHeader = "#ifndef SUPPORT_H\n#define SUPPORT_H\n#include <stdint.h>\ntypedef uint64_t support_identity_fn(uint64_t);\nuint64_t support_identity(uint64_t);\n#endif\n"
const phase22BindingSource = "#include \"support.h\"\nuint64_t support_identity(uint64_t value) { return value; }\n"
const phase22BindingApp = "#include <stdio.h>\nint main(int argc, char **argv) { if (argc != 2) return 64; printf(\"%s\\n\", argv[1]); return 0; }\n"

func phase22Manifest() BindingManifest {
	return BindingManifest{Schema: BindingSchema, Sources: []string{"support.c"}, Headers: []string{"support.h"}, IncludeDirs: []string{"."},
		Symbols: []BindingSymbol{{Name: "support_identity", Header: "support.h", FunctionType: "support_identity_fn"}}, RuntimeDependencies: []string{"platform-c-runtime"}}
}

func writeBindingTestFile(t *testing.T, root, name, contents string) {
	t.Helper()
	full := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}

func writeBindingTestManifest(t *testing.T, root string, manifest BindingManifest) string {
	t.Helper()
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	writeBindingTestFile(t, root, "bindings.json", string(data))
	return filepath.Join(root, "bindings.json")
}

func phase22BindingsFixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	writeBindingTestFile(t, root, "support.h", phase22BindingHeader)
	writeBindingTestFile(t, root, "support.c", phase22BindingSource)
	return root, writeBindingTestManifest(t, root, phase22Manifest())
}

func TestPhase22BindingsBuild(t *testing.T) {
	root, manifest := phase22BindingsFixture(t)
	artifact := filepath.Join(root, "out", "app")
	receipt, err := BuildApplication(context.Background(), []byte("identity"), phase22BindingApp, artifact, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Bindings == nil || receipt.Identity == nil || len(receipt.Identity.Inventory) != 2 || len(receipt.Identity.IncludedHeaders) != 1 || receipt.Identity.IncludedHeaders[0] != "support.h" || !validSHA256(receipt.BuildID) || !validSHA256(receipt.InputID) {
		t.Fatalf("incomplete declared inventory: %+v", receipt)
	}
	if receipt.Cacheable || receipt.DependencyClosure != "incomplete" {
		t.Fatalf("unjustified closure: %+v", receipt)
	}
	var out, stderr bytes.Buffer
	result, err := RunApplication(context.Background(), artifact, "42", &out, &stderr)
	if err != nil || result.Kind != RunExited || result.ExitCode != 0 || out.String() != "42\n" || stderr.Len() != 0 {
		t.Fatalf("run=%+v error=%v stdout=%q stderr=%q", result, err, out.String(), stderr.String())
	}
}

func TestPhase22BindingsRejectInvalidInputs(t *testing.T) {
	cases := []struct {
		name, want string
		mutate     func(*testing.T, string, *BindingManifest)
	}{
		{"missing source", "missing.c", func(t *testing.T, r string, m *BindingManifest) { m.Sources[0] = "missing.c" }},
		{"missing header", "missing.h", func(t *testing.T, r string, m *BindingManifest) { m.Headers = append(m.Headers, "missing.h") }},
		{"absolute path", "relative path", func(t *testing.T, r string, m *BindingManifest) { m.Sources[0] = filepath.Join(r, "support.c") }},
		{"parent traversal", "traversal", func(t *testing.T, r string, m *BindingManifest) { m.Sources[0] = "sub/../support.c" }},
		{"include root escape", "include directory", func(t *testing.T, r string, m *BindingManifest) { m.IncludeDirs = []string{".."} }},
		{"symlink escape", "symlink escapes", func(t *testing.T, r string, m *BindingManifest) {
			other := t.TempDir()
			writeBindingTestFile(t, other, "outside.c", phase22BindingSource)
			if err := os.Symlink(filepath.Join(other, "outside.c"), filepath.Join(r, "link.c")); err != nil {
				t.Fatal(err)
			}
			m.Sources[0] = "link.c"
		}},
		{"directory symlink escape", "symlink escapes", func(t *testing.T, r string, m *BindingManifest) {
			if err := os.Symlink(t.TempDir(), filepath.Join(r, "include")); err != nil {
				t.Fatal(err)
			}
			m.IncludeDirs = []string{"include"}
		}},
		{"duplicate source", "duplicate declaration", func(t *testing.T, r string, m *BindingManifest) { m.Sources = append(m.Sources, "support.c") }},
		{"aliased source", "duplicate resolved path", func(t *testing.T, r string, m *BindingManifest) {
			if err := os.Symlink("support.c", filepath.Join(r, "alias.c")); err != nil {
				t.Fatal(err)
			}
			m.Sources = append(m.Sources, "alias.c")
		}},
		{"duplicate symbol", "duplicate symbol", func(t *testing.T, r string, m *BindingManifest) { m.Symbols = append(m.Symbols, m.Symbols[0]) }},
		{"runtime dependency", "platform-c-runtime", func(t *testing.T, r string, m *BindingManifest) { m.RuntimeDependencies = []string{"another-runtime"} }},
		{"symbol injection", "C identifiers", func(t *testing.T, r string, m *BindingManifest) { m.Symbols[0].Name = "support_identity;system" }},
		{"undeclared header", "extra.h", func(t *testing.T, r string, m *BindingManifest) {
			writeBindingTestFile(t, r, "extra.h", "#define EXTRA 1\n")
			writeBindingTestFile(t, r, "support.c", "#include \"extra.h\"\n"+phase22BindingSource)
		}},
		{"absolute included header", "out-of-root", func(t *testing.T, r string, m *BindingManifest) {
			outside := t.TempDir()
			writeBindingTestFile(t, outside, "outside.h", "#define EXTRA 1\n")
			writeBindingTestFile(t, r, "support.c", "#include "+cString(filepath.Join(outside, "outside.h"))+"\n"+phase22BindingSource)
		}},
		{"system-marked local header", "out-of-root", func(t *testing.T, r string, m *BindingManifest) {
			outside := t.TempDir()
			writeBindingTestFile(t, outside, "outside.h", "#pragma GCC system_header\n#define EXTRA 1\n")
			writeBindingTestFile(t, r, "support.c", "#include "+cString(filepath.Join(outside, "outside.h"))+"\n"+phase22BindingSource)
		}},
		{"incompatible type", "support_identity", func(t *testing.T, r string, m *BindingManifest) {
			writeBindingTestFile(t, r, "support.h", strings.Replace(phase22BindingHeader, "typedef uint64_t support_identity_fn(uint64_t);", "typedef double support_identity_fn(double);", 1))
		}},
		{"missing typedef", "missing_type", func(t *testing.T, r string, m *BindingManifest) { m.Symbols[0].FunctionType = "missing_type" }},
		{"wrong declared header", "support_identity", func(t *testing.T, r string, m *BindingManifest) {
			writeBindingTestFile(t, r, "wrong.h", "typedef unsigned long support_identity_fn(unsigned long);\n")
			m.Headers = append(m.Headers, "wrong.h")
			m.Symbols[0].Header = "wrong.h"
		}},
		{"object typedef", "must name a function type", func(t *testing.T, r string, m *BindingManifest) {
			writeBindingTestFile(t, r, "support.h", "typedef int support_identity_fn;\nextern int support_identity;\n")
			writeBindingTestFile(t, r, "support.c", "int support_identity = 1;\n")
		}},
		{"unresolved symbol", "missing_symbol", func(t *testing.T, r string, m *BindingManifest) {
			m.Symbols[0].Name = "missing_symbol"
			writeBindingTestFile(t, r, "support.h", phase22BindingHeader+"uint64_t missing_symbol(uint64_t);\n")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, _ := phase22BindingsFixture(t)
			m := phase22Manifest()
			tc.mutate(t, root, &m)
			manifest := writeBindingTestManifest(t, root, m)
			artifact := filepath.Join(root, "out", "app")
			_, err := BuildApplication(context.Background(), []byte("identity"), phase22BindingApp, artifact, manifest)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v, want %q", err, tc.want)
			}
			for _, p := range []string{artifact, applicationReceiptPath(artifact)} {
				if _, err := os.Stat(p); !os.IsNotExist(err) {
					t.Fatalf("failed build published %s: %v", p, err)
				}
			}
		})
	}
}

func TestPhase22BindingsClosedJSON(t *testing.T) {
	valid, _ := json.Marshal(phase22Manifest())
	for name, data := range map[string]string{
		"duplicate top-level": strings.Replace(string(valid), `"schema":`, `"schema":"schway.local-c/1","schema":`, 1),
		"duplicate nested":    strings.Replace(string(valid), `"name":`, `"name":"other","name":`, 1),
		"unknown":             strings.Replace(string(valid), `"schema":`, `"flags":["-w"],"schema":`, 1),
		"unknown nested":      strings.Replace(string(valid), `"name":`, `"extra":true,"name":`, 1),
		"case alias":          strings.Replace(string(valid), `"schema":`, `"Schema":`, 1),
		"nested case alias":   strings.Replace(string(valid), `"name":`, `"Name":`, 1),
		"trailing":            string(valid) + ` {}`,
		"oversized":           strings.Repeat(" ", maxBindingManifestBytes) + string(valid),
		"null":                "null",
	} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			writeBindingTestFile(t, root, "bindings.json", data)
			if _, err := LoadBindings(filepath.Join(root, "bindings.json")); err == nil {
				t.Fatal("accepted invalid manifest")
			}
		})
	}
}

func TestPhase22BindingsIgnoreAmbientIncludePath(t *testing.T) {
	root, manifest := phase22BindingsFixture(t)
	ambient := t.TempDir()
	writeBindingTestFile(t, ambient, "ambient.h", "#define VALUE 1\n")
	t.Setenv("CPATH", ambient)
	writeBindingTestFile(t, root, "support.c", "#include <ambient.h>\n"+phase22BindingSource)
	_, err := BuildApplication(context.Background(), []byte("identity"), phase22BindingApp, filepath.Join(root, "out", "app"), manifest)
	if err == nil || !strings.Contains(err.Error(), "ambient.h") {
		t.Fatalf("ambient include granted build authority: %v", err)
	}
}

func TestPhase22RelocatedBindingsCLI(t *testing.T) {
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	cli := filepath.Join(t.TempDir(), "schway")
	if out, stderr, err := phase22BindingCommand(repo, "go", "build", "-o", cli, "./cmd/schway"); err != nil {
		t.Fatalf("build CLI: %v\n%s\n%s", err, out, stderr)
	}
	var receipts []BuildReceipt
	for _, name := range []string{"original checkout", "relocated checkout"} {
		root := filepath.Join(t.TempDir(), name)
		for _, file := range []string{"identity.schway", "identity.bindings.json", "support.h", "support.c"} {
			data, err := os.ReadFile(filepath.Join(repo, "examples", "phase22", file))
			if err != nil {
				t.Fatal(err)
			}
			writeBindingTestFile(t, root, file, string(data))
		}
		artifact := filepath.Join(root, "out", "identity")
		out, stderr, err := phase22BindingCommand(root, cli, "--json", "build", filepath.Join(root, "identity.schway"), "--manifest", filepath.Join(root, "identity.bindings.json"), "--output", artifact)
		if err != nil {
			t.Fatalf("CLI manifest build: %v stderr=%s stdout=%s", err, stderr, out)
		}
		var receipt BuildReceipt
		if err := json.Unmarshal([]byte(out), &receipt); err != nil {
			t.Fatal(err)
		}
		receipts = append(receipts, receipt)
		files, err := os.ReadDir(filepath.Dir(artifact))
		if err != nil || len(files) != 2 {
			t.Fatalf("build intermediates remain: %v %v", files, err)
		}
		// Remove all original inputs before using the retained pair elsewhere.
		retained := filepath.Join(t.TempDir(), "identity")
		if err := os.Rename(artifact, retained); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(applicationReceiptPath(artifact), applicationReceiptPath(retained)); err != nil {
			t.Fatal(err)
		}
		if err := os.RemoveAll(root); err != nil {
			t.Fatal(err)
		}
		for _, input := range []string{"7", "42"} {
			out, stderr, err := phase22BindingCommand("", cli, "app", "run", retained, "--", input)
			if err != nil || out != input+"\n" || stderr != "" {
				t.Fatalf("retained CLI run: %v stdout=%q stderr=%q", err, out, stderr)
			}
		}
	}
	if receipts[0].InputID != receipts[1].InputID || !equalJSON(receipts[0].Identity.Inventory, receipts[1].Identity.Inventory) || receipts[0].Identity.ManifestDigest != receipts[1].Identity.ManifestDigest {
		t.Fatalf("relocation changed canonical input identity: %s != %s", receipts[0].InputID, receipts[1].InputID)
	}
	if receipts[0].ExecutableDigest == receipts[1].ExecutableDigest && receipts[0].BuildID != receipts[1].BuildID {
		t.Fatal("same inputs and binary bytes changed build ID")
	}
	if receipts[0].BuildID != receipts[1].BuildID {
		t.Log("binary bytes differ across builds; input IDs agree and both closures must remain incomplete")
	}
	for _, receipt := range receipts {
		if receipt.Cacheable || receipt.DependencyClosure != "incomplete" {
			t.Fatal("unknown host closure was accepted as complete")
		}
	}
}

func phase22BindingCommand(directory, name string, args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = directory
	var stdout, stderr boundedWriter
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if stdout.overflowed() || stderr.overflowed() {
		return "", "", fmt.Errorf("CLI test output exceeded bound")
	}
	return string(stdout.bytes()), string(stderr.bytes()), err
}

func TestPhase22BuildIdentityDeclaredInputMutations(t *testing.T) {
	// Two real functions and typedefs let a single manifest-field edit stay
	// linkable, proving identity invalidation independently of compiler refusal.
	header := phase22BindingHeader + "typedef uint64_t alternate_fn(uint64_t);\nuint64_t alternate_identity(uint64_t);\n"
	source := phase22BindingSource + "uint64_t alternate_identity(uint64_t value) { return value; }\n"
	var baseline BuildReceipt
	cases := []struct {
		name   string
		mutate func(*testing.T, string, *BindingManifest, *[]byte)
	}{
		{"baseline", func(t *testing.T, r string, m *BindingManifest, s *[]byte) {}},
		{"Lang source", func(t *testing.T, r string, m *BindingManifest, s *[]byte) {
			*s = append(*s, []byte("\n// source identity mutation\n")...)
		}},
		{"C source", func(t *testing.T, r string, m *BindingManifest, s *[]byte) {
			writeBindingTestFile(t, r, "support.c", source+"/* changed C bytes */\n")
		}},
		{"header", func(t *testing.T, r string, m *BindingManifest, s *[]byte) {
			writeBindingTestFile(t, r, "support.h", header+"/* changed header bytes */\n")
		}},
		{"symbol", func(t *testing.T, r string, m *BindingManifest, s *[]byte) { m.Symbols[0].Name = "alternate_identity" }},
		{"ABI typedef", func(t *testing.T, r string, m *BindingManifest, s *[]byte) {
			m.Symbols[0].FunctionType = "alternate_fn"
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			writeBindingTestFile(t, root, "support.h", header)
			writeBindingTestFile(t, root, "support.c", source)
			m := phase22Manifest()
			langSource := []byte("identity Lang source")
			tc.mutate(t, root, &m, &langSource)
			receipt, err := BuildApplication(context.Background(), langSource, phase22BindingApp, filepath.Join(root, "out", "app"), writeBindingTestManifest(t, root, m))
			if err != nil {
				t.Fatal(err)
			}
			if tc.name == "baseline" {
				baseline = receipt
				return
			}
			if receipt.InputID == baseline.InputID || receipt.BuildID == baseline.BuildID {
				t.Fatalf("%s mutation did not change input and artifact identity", tc.name)
			}
		})
	}
	// Non-user-configurable flags and host facts use the same serialized
	// production identity function. Runtime alternatives also fail resolution.
	for name, mutate := range map[string]func(*BuildIdentityInputs){
		"emitted C":     func(in *BuildIdentityInputs) { in.EmittedCDigest = digestBytes([]byte("other C")) },
		"compile flags": func(in *BuildIdentityInputs) { in.Flags[0] = "-std=c11" },
		"link argv": func(in *BuildIdentityInputs) {
			in.Commands[len(in.Commands)-1] = append(in.Commands[len(in.Commands)-1], "-static")
		},
		"compiler fingerprint": func(in *BuildIdentityInputs) { in.CompilerDigest = digestBytes([]byte("different compiler")) },
		"compiler version":     func(in *BuildIdentityInputs) { in.CompilerVersion += " changed" },
		"target":               func(in *BuildIdentityInputs) { in.Target = "different-target" },
		"host ABI":             func(in *BuildIdentityInputs) { in.HostABI = "different-host" },
		"runtime declaration":  func(in *BuildIdentityInputs) { in.RuntimeDependencies = []string{"different-runtime"} },
	} {
		t.Run(name, func(t *testing.T) {
			encoded, _ := json.Marshal(baseline.Identity)
			var in BuildIdentityInputs
			if err := json.Unmarshal(encoded, &in); err != nil {
				t.Fatal(err)
			}
			mutate(&in)
			if in.ID() == baseline.BuildID || in.InputID() == baseline.InputID {
				t.Fatalf("%s is omitted from identity", name)
			}
		})
	}
}

func TestPhase22BuildIdentityReceiptAndBinaryMismatch(t *testing.T) {
	root, manifest := phase22BindingsFixture(t)
	artifact := filepath.Join(root, "out", "app")
	receipt, err := BuildApplication(context.Background(), []byte("identity"), phase22BindingApp, artifact, manifest)
	if err != nil {
		t.Fatal(err)
	}
	originalBytes, err := os.ReadFile(artifact)
	if err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(root, "out", "other")
	if _, err := BuildApplication(context.Background(), []byte("other"), "int main(void) { return 17; }\n", other, manifest); err != nil {
		t.Fatal(err)
	}
	otherBytes, err := os.ReadFile(other)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(originalBytes, otherBytes) {
		t.Fatal("negative control binaries are identical")
	}
	launches := 0
	runner := DefaultRunner()
	runner.command = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		launches++
		return exec.CommandContext(ctx, name, args...)
	}
	for name, mutate := range map[string]func(*BuildReceipt){
		"swapped binary": func(r *BuildReceipt) {
			if err := os.WriteFile(artifact, otherBytes, 0o700); err != nil {
				t.Fatal(err)
			}
		},
		"stale build ID":   func(r *BuildReceipt) { r.BuildID = digestBytes([]byte("stale")) },
		"stale source":     func(r *BuildReceipt) { r.SourceDigest = digestBytes([]byte("other source")) },
		"changed flags":    func(r *BuildReceipt) { r.Flags = []string{"-O3"} },
		"missing manifest": func(r *BuildReceipt) { r.Bindings = nil },
		"missing identity": func(r *BuildReceipt) { r.Identity = nil },
		"changed ABI":      func(r *BuildReceipt) { r.Bindings.Symbols[0].FunctionType = "different_type" },
	} {
		t.Run(name, func(t *testing.T) {
			if err := os.WriteFile(artifact, originalBytes, 0o700); err != nil {
				t.Fatal(err)
			}
			encoded, _ := json.Marshal(receipt)
			var changed BuildReceipt
			if err := json.Unmarshal(encoded, &changed); err != nil {
				t.Fatal(err)
			}
			mutate(&changed)
			encoded, _ = json.Marshal(changed)
			if err := os.WriteFile(applicationReceiptPath(artifact), encoded, 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := runner.RunApplication(context.Background(), artifact, "7", &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
				t.Fatal("stale artifact/receipt accepted")
			}
			if launches != 0 {
				t.Fatalf("launched stale artifact %d times", launches)
			}
		})
	}
}

func TestPhase22BindingsBuildNeverLaunchesAndRunLaunchesOnce(t *testing.T) {
	root, manifest := phase22BindingsFixture(t)
	marker := filepath.Join(root, "launches.txt")
	artifact := filepath.Join(root, "out", "app")
	app := "#include <stdio.h>\nint main(int argc, char **argv) {\n" +
		"if (argc != 2) return 64;\nFILE *marker = fopen(" + cString(marker) + ", \"a\");\n" +
		"if (marker == NULL) return 90;\nfputs(\"launch\\n\", marker); fclose(marker);\nprintf(\"%s\\n\",argv[1]); return 0;\n}\n"
	launches := 0
	runner := DefaultRunner().EnableCommandRecording()
	runner.command = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		if name == artifact {
			launches++
		} else if filepath.Base(name) != "clang" {
			t.Errorf("unexpected build executable: %s", name)
		}
		return exec.CommandContext(ctx, name, args...)
	}
	if _, err := runner.BuildApplication(context.Background(), []byte("marker"), app, artifact, manifest); err != nil {
		t.Fatal(err)
	}
	if launches != 0 {
		t.Fatalf("build launched %d apps", launches)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("build executed application effects: %v", err)
	}
	if len(runner.LastCommandLines()) < 5 {
		t.Fatal("recorder did not witness C/probe compilation and linkage")
	}
	var out, stderr bytes.Buffer
	result, err := runner.RunApplication(context.Background(), artifact, "7", &out, &stderr)
	if err != nil || result.Kind != RunExited || result.ExitCode != 0 || out.String() != "7\n" || stderr.Len() != 0 || launches != 1 {
		t.Fatalf("run=%+v err=%v stdout=%q stderr=%q launches=%d", result, err, out.String(), stderr.String(), launches)
	}
	data, err := os.ReadFile(marker)
	if err != nil || string(data) != "launch\n" {
		t.Fatalf("marker=%q err=%v", data, err)
	}
}

func TestPhase22BindingsCanonicalManifestAndSnapshot(t *testing.T) {
	root, manifest := phase22BindingsFixture(t)
	before, err := ResolveBindings(manifest)
	if err != nil {
		t.Fatal(err)
	}
	compact, _ := json.Marshal(before.Manifest)
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, compact, "", "    "); err != nil {
		t.Fatal(err)
	}
	writeBindingTestFile(t, root, "bindings.json", pretty.String())
	after, err := ResolveBindings(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if before.ManifestDigest != after.ManifestDigest {
		t.Fatal("JSON whitespace changed canonical manifest identity")
	}
	runner := DefaultRunner()
	changed := false
	runner.command = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		if !changed {
			changed = true
			writeBindingTestFile(t, root, "support.c", "invalid C after snapshot\n")
		}
		return exec.CommandContext(ctx, name, args...)
	}
	receipt, err := runner.BuildApplication(context.Background(), []byte("identity"), phase22BindingApp, filepath.Join(root, "out", "app"), manifest)
	if err != nil {
		t.Fatalf("compiler did not use captured bytes: %v", err)
	}
	if !equalJSON(receipt.Identity.Inventory, before.Inventory) {
		t.Fatal("receipt does not bind captured bytes")
	}
}

func TestPhase22BindingsNestedHeadersAndSpacedPaths(t *testing.T) {
	root := t.TempDir()
	m := phase22Manifest()
	m.Sources = []string{"src folder/support.c"}
	m.Headers = []string{"include folder/support.h", "include folder/types.h"}
	m.IncludeDirs = []string{"include folder"}
	m.Symbols[0].Header = "include folder/support.h"
	writeBindingTestFile(t, root, m.Sources[0], phase22BindingSource)
	writeBindingTestFile(t, root, m.Headers[0], strings.Replace(phase22BindingHeader, "#include <stdint.h>", "#include \"types.h\"", 1))
	writeBindingTestFile(t, root, m.Headers[1], "#include <stdint.h>\n")
	receipt, err := BuildApplication(context.Background(), []byte("identity"), phase22BindingApp, filepath.Join(root, "out", "app"), writeBindingTestManifest(t, root, m))
	if err != nil {
		t.Fatal(err)
	}
	if !equalJSON(receipt.Identity.IncludedHeaders, []string{"include folder/support.h", "include folder/types.h"}) {
		t.Fatalf("transitive local inventory=%v", receipt.Identity.IncludedHeaders)
	}
}

func TestPhase22BuildIdentityFeedbackSamples(t *testing.T) {
	// Small bounded samples: first output path vs rebuild at that path. OS
	// caches are not flushed; these are process-cold, not cold-machine samples.
	var coldBuild, warmBuild, coldRun, warmRun []float64
	for i := 0; i < 3; i++ {
		root, manifest := phase22BindingsFixture(t)
		artifact := filepath.Join(root, "out", "app")
		for pass := 0; pass < 2; pass++ {
			start := time.Now()
			if _, err := BuildApplication(context.Background(), []byte("identity"), phase22BindingApp, artifact, manifest); err != nil {
				t.Fatal(err)
			}
			elapsed := float64(time.Since(start).Microseconds()) / 1000
			if pass == 0 {
				coldBuild = append(coldBuild, elapsed)
			} else {
				warmBuild = append(warmBuild, elapsed)
			}
		}
		for pass := 0; pass < 2; pass++ {
			var out, stderr bytes.Buffer
			start := time.Now()
			result, err := RunApplication(context.Background(), artifact, "42", &out, &stderr)
			elapsed := float64(time.Since(start).Microseconds()) / 1000
			if err != nil || result.Kind != RunExited || result.ExitCode != 0 || out.String() != "42\n" || stderr.Len() != 0 {
				t.Fatal("feedback sample failed")
			}
			if pass == 0 {
				coldRun = append(coldRun, elapsed)
			} else {
				warmRun = append(warmRun, elapsed)
			}
		}
	}
	t.Log(fmt.Sprintf("ms: first-build=%v rebuild=%v first-run=%v repeated-run=%v", coldBuild, warmBuild, coldRun, warmRun))
}
