package check_test

import (
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

// TestPhase321ModuleAwareRoutes pins D-12: the module-aware route set is named
// once, returned fresh, and is the source of the module.imports_not_loaded
// refusal text. It mirrors phase30ModuleRouteFunctions/phase30BoundedWant in the
// session package by comment only.
func TestPhase321ModuleAwareRoutes(t *testing.T) {
	want := []string{"check", "run", "build", "app verify", "debug-map"}

	t.Run("set", func(t *testing.T) {
		got := check.ModuleAwareRoutes()
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Fatalf("ModuleAwareRoutes() = %v, want %v", got, want)
		}
	})
	t.Run("fresh-slice", func(t *testing.T) {
		first := check.ModuleAwareRoutes()
		if len(first) == 0 {
			t.Fatal("ModuleAwareRoutes() is empty")
		}
		first[0] = "mutated"
		if second := check.ModuleAwareRoutes(); len(second) == 0 || second[0] == "mutated" {
			t.Fatalf("mutating the returned slice changed a later call: %v", second)
		}
	})
	t.Run("refusal-message", func(t *testing.T) {
		source := "module m\n\nimport lib {\n  fn twice\n}\n\nexport {\n  fn main\n}\n\nfn main(n: U64) -> U64 {\n  n\n}\n"
		parsed := syntax.Parse([]byte(source))
		result := check.Program(parsed.Program)
		message := ""
		for _, d := range result.Diagnostics {
			if d.Code == "module.imports_not_loaded" {
				message = d.Message
			}
		}
		if message == "" {
			t.Fatalf("no module.imports_not_loaded diagnostic: %+v", result.Diagnostics)
		}
		for _, route := range want {
			if !strings.Contains(message, route) {
				t.Errorf("message %q does not name route %q", message, route)
			}
		}
		if !strings.Contains(message, "(check, run, build, app verify, debug-map)") {
			t.Errorf("message %q does not list the set in order", message)
		}
		if !strings.Contains(message, "reads a single file by design") {
			t.Errorf("message %q does not say the refusing routes read a single file by design", message)
		}
	})
}
