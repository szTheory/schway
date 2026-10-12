package cgen

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
)

// phase30Placement moves the source function `source` into `module` under the
// bare name `name`, the way the module loader's merge does for an imported
// function.
type phase30Placement struct{ source, module, name string }

// phase30MergedProgram checks a single-module source whose entry module is
// "probe.m", then rewrites the placed functions into imported modules: IDs
// follow the s1:<module>:fn:<name> shape, core.Function.Module/Name/Exported
// are set, and the Imports facts cover every cross-module call edge. It
// stands in for modload+check output so the cgen layer is tested alone.
func phase30MergedProgram(t *testing.T, source string, placements []phase30Placement) core.Program {
	t.Helper()
	program := phase29CheckedProgram(t, source)
	raw, err := json.Marshal(program)
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	for _, placement := range placements {
		from := "s1:" + program.Module + ":fn:" + placement.source
		to := "s1:" + placement.module + ":fn:" + placement.name
		text = strings.ReplaceAll(text, from+`"`, to+`"`)
		text = strings.ReplaceAll(text, from+":", to+":")
	}
	var merged core.Program
	if err := json.Unmarshal([]byte(text), &merged); err != nil {
		t.Fatal(err)
	}
	moduleOf := map[string]string{}
	byID := map[string]int{}
	for _, placement := range placements {
		id := "s1:" + placement.module + ":fn:" + placement.name
		index := -1
		for i, function := range merged.Functions {
			if function.ID == id {
				index = i
			}
		}
		if index < 0 {
			t.Fatalf("placed function %q not found after ID rewrite", id)
		}
		merged.Functions[index].Module = placement.module
		merged.Functions[index].Name = placement.name
		merged.Functions[index].Exported = true
		moduleOf[id] = placement.module
	}
	for i, function := range merged.Functions {
		byID[function.ID] = i
	}
	seen := map[core.ModuleImport]bool{}
	for _, function := range merged.Functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind != core.OpCall {
				continue
			}
			callee := merged.Functions[byID[operation.CalleeID]]
			if callee.Module == "" || callee.Module == function.Module {
				continue
			}
			fact := core.ModuleImport{From: function.Module, Module: callee.Module, Name: callee.Name}
			if !seen[fact] {
				seen[fact] = true
				merged.Imports = append(merged.Imports, fact)
			}
		}
	}
	sort.Slice(merged.Imports, func(i, j int) bool {
		a, b := merged.Imports[i], merged.Imports[j]
		if a.From != b.From {
			return a.From < b.From
		}
		if a.Module != b.Module {
			return a.Module < b.Module
		}
		return a.Name < b.Name
	})
	return merged
}

const phase30AdlerSource = "module probe.m\n\nexport {\n  fn main\n}\n\nfn adler32(n: U64) -> U64 {\n  var s = n\n  s = s + 1\n  s\n}\n\nfn main(n: U64) -> U64 {\n  let r = adler32(n)\n  r\n}\n"

// TestPhase30ModuleFunctionCName is the D-12 tracer: an imported function is
// defined and called under its length-prefixed module name, and the entry
// function keeps the name it has in the single-module emission.
func TestPhase30ModuleFunctionCName(t *testing.T) {
	merged := phase30MergedProgram(t, phase30AdlerSource, []phase30Placement{{"adler32", "adler32", "adler32"}})
	single := phase29CheckedProgram(t, phase30AdlerSource)

	mergedC, err := EmitApplication(merged)
	if err != nil {
		t.Fatalf("emit merged: %v", err)
	}
	singleC, err := emitProgramWithShell(single, programApplicationShell, false)
	if err != nil {
		t.Fatalf("emit single: %v", err)
	}
	const imported = "SCHWAY_M_7_ADLER32_7_ADLER32"
	definition := regexp.MustCompile(`(?m)^static uint64_t ` + imported + `\([^)]*\) \{$`)
	if got := len(definition.FindAllString(mergedC, -1)); got != 1 {
		t.Fatalf("definition lines for %s = %d, want 1:\n%s", imported, got, mergedC)
	}
	if !regexp.MustCompile(`= ` + imported + `\(`).MatchString(mergedC) {
		t.Fatalf("main does not call %s:\n%s", imported, mergedC)
	}
	if strings.Contains(mergedC, "SCHWAY_ADLER32") {
		t.Fatalf("imported function leaked its bare entry-style name:\n%s", mergedC)
	}
	if !strings.Contains(singleC, "SCHWAY_MAIN(") || !strings.Contains(mergedC, "SCHWAY_MAIN(") {
		t.Fatal("entry main did not keep its cName in both emissions")
	}
	if strings.Count(mergedC, "SCHWAY_MAIN(") != strings.Count(singleC, "SCHWAY_MAIN(") {
		t.Fatal("entry main identifier use changed between single-module and merged emission")
	}

	t.Run("clang", func(t *testing.T) {
		clang, err := exec.LookPath("clang")
		if err != nil {
			t.Skip("env:clang")
		}
		dir := t.TempDir()
		source := filepath.Join(dir, "merged.c")
		if err := os.WriteFile(source, []byte(mergedC), 0o600); err != nil {
			t.Fatal(err)
		}
		command := exec.Command(clang, "-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", "-c", source, "-o", filepath.Join(dir, "merged.o"))
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("clang rejected merged C: %v\n%s", err, output)
		}
	})
}

// phase30Universe is the adversarial placement set of the Task 2 rows: modules
// a_b, a and adler32 against names that collide when joined naively or when
// uppercased, plus an entry function literally named like a mangled name.
func phase30Universe() (source string, placements []phase30Placement) {
	modules := []string{"a_b", "a", "adler32"}
	names := []string{"c", "b_c", "f", "Foo", "foo"}
	var chain []string
	chain = append(chain, "m_1_a_1_f")
	for _, module := range modules {
		for _, name := range names {
			placeholder := fmt.Sprintf("g%d", len(placements))
			placements = append(placements, phase30Placement{placeholder, module, name})
			chain = append(chain, placeholder)
		}
	}
	var out strings.Builder
	out.WriteString("module probe.m\n\nexport {\n  fn main\n}\n\n")
	for index, name := range chain {
		if index == len(chain)-1 {
			fmt.Fprintf(&out, "fn %s(n: U64) -> U64 {\n  n\n}\n\n", name)
			continue
		}
		fmt.Fprintf(&out, "fn %s(n: U64) -> U64 {\n  let r = %s(n)\n  r\n}\n\n", name, chain[index+1])
	}
	fmt.Fprintf(&out, "fn main(n: U64) -> U64 {\n  let r = %s(n)\n  r\n}\n", chain[0])
	return out.String(), placements
}

var phase30Definition = regexp.MustCompile(`(?m)^static [A-Za-z_0-9 ]+? (SCHWAY_[A-Z0-9_]*)\([^\n]*\) \{$`)

func phase30DefinedNames(c string) []string {
	var names []string
	for _, match := range phase30Definition.FindAllStringSubmatch(c, -1) {
		names = append(names, match[1])
	}
	return names
}

// TestPhase30CNamesInjective emits the adversarial universe and requires every
// defined identifier to be distinct and confined. The entry function whose
// name equals a mangled name keeps it (entry wins); the imported twin takes the
// allocate suffix. importedFirstForTest observes the contest being decided by
// pass order.
func TestPhase30CNamesInjective(t *testing.T) {
	source, placements := phase30Universe()
	program := phase30MergedProgram(t, source, placements)
	shape := regexp.MustCompile(`^SCHWAY_[A-Z0-9_]*$`)

	text, err := emitProgramWithShell(program, programApplicationShell, false)
	if err != nil {
		t.Fatalf("emit: %v", err)
	}
	defined := phase30DefinedNames(text)
	if len(defined) < len(placements)+2 {
		t.Fatalf("found %d defined functions, want at least %d:\n%s", len(defined), len(placements)+2, text)
	}
	seen := map[string]bool{}
	for _, name := range defined {
		if seen[name] {
			t.Fatalf("identifier %q defined twice", name)
		}
		seen[name] = true
		if !shape.MatchString(name) || strings.Contains(name, "__") {
			t.Fatalf("identifier %q leaves the confined namespace", name)
		}
	}
	if !strings.Contains(text, "SCHWAY_M_1_A_1_F(") {
		t.Fatal("entry function m_1_a_1_f lost its own cName")
	}

	names := allocateFunctionNames(program.Functions)
	byKey := map[string]string{}
	for index, function := range program.Functions {
		byKey[function.Module+"|"+function.Name] = names[index]
	}
	if got := byKey["|m_1_a_1_f"]; got != "SCHWAY_M_1_A_1_F" {
		t.Fatalf("entry m_1_a_1_f = %q, want SCHWAY_M_1_A_1_F", got)
	}
	if got := byKey["a|f"]; got == "SCHWAY_M_1_A_1_F" || !strings.HasPrefix(got, "SCHWAY_M_1_A_1_F_SCHWAY_FUNCTION_") {
		t.Fatalf("imported a.f = %q, want the allocate suffix after the entry contest", got)
	}
	if byKey["a|Foo"] == byKey["a|foo"] {
		t.Fatalf("case variants a.Foo and a.foo share %q", byKey["a|Foo"])
	}

	t.Run("control-imported-first", func(t *testing.T) {
		previous := importedFirstForTest
		importedFirstForTest = true
		defer func() { importedFirstForTest = previous }()
		flipped := allocateFunctionNames(program.Functions)
		flippedByKey := map[string]string{}
		for index, function := range program.Functions {
			flippedByKey[function.Module+"|"+function.Name] = flipped[index]
		}
		if flippedByKey["|m_1_a_1_f"] == "SCHWAY_M_1_A_1_F" {
			t.Fatal("control did not move the entry name: imported-first order is not observed")
		}
	})
}

// TestPhase30EmitOrderIndependent requires the same C for any order of the
// functions slice and the imports facts. Layer note: callgraph.Order fixes the
// emitted function order before allocation, so whole-program emission is
// order independent by that layer plus the canonical (Module, Name) sort in
// allocateFunctionNames; the control exercises the allocation layer directly,
// where skipCanonicalImportSortForTest makes input order visible.
func TestPhase30EmitOrderIndependent(t *testing.T) {
	source, placements := phase30Universe()
	program := phase30MergedProgram(t, source, placements)
	canonical, err := emitProgramWithShell(program, programApplicationShell, false)
	if err != nil {
		t.Fatalf("emit: %v", err)
	}

	reversed := program
	reversed.Functions = append([]core.Function(nil), program.Functions...)
	for i, j := 0, len(reversed.Functions)-1; i < j; i, j = i+1, j-1 {
		reversed.Functions[i], reversed.Functions[j] = reversed.Functions[j], reversed.Functions[i]
	}
	reversed.Imports = append([]core.ModuleImport(nil), program.Imports...)
	for i, j := 0, len(reversed.Imports)-1; i < j; i, j = i+1, j-1 {
		reversed.Imports[i], reversed.Imports[j] = reversed.Imports[j], reversed.Imports[i]
	}
	// corevalidate wants canonically sorted Imports; this test feeds the
	// emitter directly, so only the function order is permuted for it.
	reversedC, err := emitProgramWithShell(reversed, programApplicationShell, false)
	if err != nil {
		t.Fatalf("emit reversed: %v", err)
	}
	if reversedC != canonical {
		t.Fatal("reversing the functions and imports changed the generated C")
	}

	// Allocation layer: a module whose names contend for one allocate suffix.
	contending := []core.Function{
		{ID: "s1:a:fn:foo", Name: "foo", Module: "a"},
		{ID: "s1:a:fn:Foo", Name: "Foo", Module: "a"},
		{ID: "s1:b:fn:f", Name: "f", Module: "b"},
	}
	flipped := []core.Function{contending[2], contending[1], contending[0]}
	assignment := func(functions []core.Function) map[string]string {
		out := map[string]string{}
		for index, name := range allocateFunctionNames(functions) {
			out[functions[index].ID] = name
		}
		return out
	}
	forward, backward := assignment(contending), assignment(flipped)
	for id, name := range forward {
		if backward[id] != name {
			t.Fatalf("canonical allocation depends on input order: %s -> %q vs %q", id, name, backward[id])
		}
	}

	t.Run("control-skip-canonical-sort", func(t *testing.T) {
		previous := skipCanonicalImportSortForTest
		skipCanonicalImportSortForTest = true
		defer func() { skipCanonicalImportSortForTest = previous }()
		forwardRaw, backwardRaw := assignment(contending), assignment(flipped)
		differs := false
		for id, name := range forwardRaw {
			if backwardRaw[id] != name {
				differs = true
			}
		}
		if !differs {
			t.Fatal("skipCanonicalImportSortForTest did not make allocation order dependent: the control is inert")
		}
	})
}

// TestPhase30EmitDefinesEachFunctionOnce requires that a function reached
// through two import paths (the merged diamond, present once in the core
// program) is defined exactly once.
func TestPhase30EmitDefinesEachFunctionOnce(t *testing.T) {
	program := phase30MergedProgram(t, diamondSource(), []phase30Placement{
		{"shared", "c", "shared"}, {"left", "a", "left"}, {"right", "b", "right"},
	})
	text, err := emitProgramWithShell(program, programApplicationShell, false)
	if err != nil {
		t.Fatalf("emit: %v", err)
	}
	for _, name := range []string{"SCHWAY_M_1_C_6_SHARED", "SCHWAY_M_1_A_4_LEFT", "SCHWAY_M_1_B_5_RIGHT"} {
		definition := regexp.MustCompile(`(?m)^static uint64_t ` + name + `\([^)]*\) \{$`)
		if got := len(definition.FindAllString(text, -1)); got != 1 {
			t.Fatalf("%s defined %d times, want 1:\n%s", name, got, text)
		}
	}
	seen := map[string]bool{}
	for _, name := range phase30DefinedNames(text) {
		if seen[name] {
			t.Fatalf("%s defined twice", name)
		}
		seen[name] = true
	}
}

func diamondSource() string {
	return "module probe.m\n\nexport {\n  fn main\n}\n\nfn shared(n: U64) -> U64 {\n  n\n}\n\nfn left(n: U64) -> U64 {\n  let r = shared(n)\n  r\n}\n\nfn right(n: U64) -> U64 {\n  let r = shared(n)\n  r\n}\n\nfn main(n: U64) -> U64 {\n  let l = left(n)\n  let r = right(l)\n  r\n}\n"
}
