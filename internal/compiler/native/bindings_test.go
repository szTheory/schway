package native

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
		"duplicate top-level": strings.Replace(string(valid), `"schema":`, `"schema":"lang.local-c/1","schema":`, 1),
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
