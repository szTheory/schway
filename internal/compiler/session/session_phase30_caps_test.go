package session_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/session"
)

// phase30CapFixtures builds one refused-by-cap program per D-13 cap code. The
// non-refused boundary fixture is the armed control: it must NOT name any cap
// code, so each assertion below can fail.
func phase30CapFixture(t *testing.T, kind string) string {
	t.Helper()
	const header = "module app.main\n\n"
	const tail = "export {\n  fn main\n}\n\nfn main(n: U64) -> U64 {\n  n\n}\n"
	lib := func(name string, imports ...string) string {
		var out strings.Builder
		out.WriteString("module " + name + "\n\n")
		for _, module := range imports {
			out.WriteString("import " + module + " {\n  fn f\n}\n\n")
		}
		out.WriteString("export {\n  fn f\n}\n\nfn f(n: U64) -> U64 {\n  n\n}\n")
		return out.String()
	}
	pad := func(source string, size int) string {
		return source + "//" + strings.Repeat("x", size-len(source)-3) + "\n"
	}
	files := map[string]string{}
	var entry strings.Builder
	entry.WriteString(header)
	switch kind {
	case "count":
		for index := 1; index <= 16; index++ {
			name := fmt.Sprintf("m%d", index)
			entry.WriteString("import " + name + " {\n  fn f\n}\n\n")
			files[name+".schway"] = lib(name)
		}
	case "depth":
		entry.WriteString("import m1 {\n  fn f\n}\n\n")
		for index := 1; index <= 9; index++ {
			name := fmt.Sprintf("m%d", index)
			if index < 9 {
				files[name+".schway"] = lib(name, fmt.Sprintf("m%d", index+1))
			} else {
				files[name+".schway"] = lib(name)
			}
		}
	case "bytes":
		const mebibyte = 1 << 20
		for _, name := range []string{"a", "b", "c", "d"} {
			entry.WriteString("import " + name + " {\n  fn f\n}\n\n")
			files[name+".schway"] = pad(lib(name), mebibyte)
		}
		files["d.schway"] = lib("d")
		entry.WriteString(tail)
		files["main.schway"] = pad(entry.String(), mebibyte)
		return writeModuleFixture(t, files)
	case "control":
		entry.WriteString("import m1 {\n  fn f\n}\n\n")
		files["m1.schway"] = lib("m1")
	}
	entry.WriteString(tail)
	files["main.schway"] = entry.String()
	return writeModuleFixture(t, files)
}

var phase30CapCodes = map[string]string{
	"count": "module.count_limit",
	"depth": "module.depth_limit",
	"bytes": "module.total_bytes_limit",
}

// TestPhase30CapDiagnosticsReachCheckAndDebugMap pins CR-01: the loader's cap
// refusals surface on the session routes, not module.imports_not_loaded.
func TestPhase30CapDiagnosticsReachCheckAndDebugMap(t *testing.T) {
	for kind, code := range phase30CapCodes {
		t.Run(kind, func(t *testing.T) {
			entry := phase30CapFixture(t, kind)
			routes := map[string]func() (protocol.Result, error){
				"check":     func() (protocol.Result, error) { return session.CheckCommandFile(entry) },
				"debug-map": func() (protocol.Result, error) { return session.DebugMapCommandFile(entry, "") },
			}
			for name, route := range routes {
				result, err := route()
				if err != nil {
					t.Fatalf("%s: %v", name, err)
				}
				text := phase30ResultJSON(t, result)
				if result.Status != protocol.StatusInvalid || !strings.Contains(text, code) {
					t.Fatalf("%s: status %s, want invalid naming %s: %s", name, result.Status, code, text)
				}
				if strings.Contains(text, "module.imports_not_loaded") {
					t.Fatalf("%s: cap refusal fell through to module.imports_not_loaded: %s", name, text)
				}
			}
		})
	}

	t.Run("control-within-caps-names-no-cap-code", func(t *testing.T) {
		result, err := session.CheckCommandFile(phase30CapFixture(t, "control"))
		if err != nil {
			t.Fatal(err)
		}
		text := phase30ResultJSON(t, result)
		for _, code := range phase30CapCodes {
			if strings.Contains(text, code) {
				t.Fatalf("a within-cap program names %s: %s", code, text)
			}
		}
	})
}
