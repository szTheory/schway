package modload

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/ast"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

// fakeLoad runs the graph loader over an in-memory directory. The entry is
// parsed from entrySource; files maps "name.schway" to content. It returns the
// fake filesystem so a test can read its open counter.
func fakeLoad(t *testing.T, entrySource string, files map[string]string) (Set, []diagnostic.Diagnostic, *fakeFileSystem) {
	t.Helper()
	fsys := newFakeFileSystem()
	id := 10
	for name, content := range files {
		fsys.addFile(name, content, id)
		id++
	}
	parsed := syntax.Parse([]byte(entrySource))
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("entry fixture failed to parse: %+v", parsed.Diagnostics)
	}
	entry := parsedModule{program: parsed.Program, tree: parsed.Tree, source: Source{Bytes: []byte(entrySource)}}
	set, diags := loadGraph(entry, nil, newFakeResolver(fsys))
	return set, diags, fsys
}

func leafFiles(count int) (string, map[string]string) {
	names := make([]string, 0, count)
	files := map[string]string{}
	for index := 1; index <= count; index++ {
		name := fmt.Sprintf("m%d", index)
		names = append(names, name)
		files[name+".schway"] = libSource(name)
	}
	return entryUsing(names...), files
}

// chainFiles builds m1 -> m2 -> ... -> m<length>, with the entry importing m1.
func chainFiles(length int) (string, map[string]string) {
	files := map[string]string{}
	for index := 1; index <= length; index++ {
		name := fmt.Sprintf("m%d", index)
		if index < length {
			files[name+".schway"] = libSource(name, fmt.Sprintf("m%d", index+1))
		} else {
			files[name+".schway"] = libSource(name)
		}
	}
	return entryUsing("m1"), files
}

// padTo appends one line comment so source is exactly size bytes.
func padTo(t *testing.T, source string, size int) string {
	t.Helper()
	pad := size - len(source)
	if pad < 4 {
		t.Fatalf("source of %d bytes cannot be padded to %d", len(source), size)
	}
	return source + "//" + strings.Repeat("x", pad-3) + "\n"
}

func TestPhase30CapsExactNAndNPlusOne(t *testing.T) {
	t.Run("count", func(t *testing.T) {
		entry, files := leafFiles(15)
		if _, diags, fsys := fakeLoad(t, entry, files); len(diags) != 0 || fsys.opens != 15 {
			t.Fatalf("entry plus 15 leaves: diagnostics %v, opens %d; want none and 15", codesOf(diags), fsys.opens)
		}
		entry, files = leafFiles(16)
		_, diags, fsys := fakeLoad(t, entry, files)
		if len(diags) != 1 || diags[0].Code != "module.count_limit" {
			t.Fatalf("entry plus 16 leaves: %v, want module.count_limit", codesOf(diags))
		}
		if fsys.opens != 15 {
			t.Fatalf("the 17th module was opened: opens = %d, want 15", fsys.opens)
		}
		t.Run("countAfterOpen-control", func(t *testing.T) {
			countAfterOpenForTest = true
			defer func() { countAfterOpenForTest = false }()
			_, diags, fsys := fakeLoad(t, entry, files)
			if !hasCode(diags, "module.count_limit") || fsys.opens != 16 {
				t.Fatalf("diagnostics %v, opens %d; the mutant must open the 17th module", codesOf(diags), fsys.opens)
			}
		})
	})

	t.Run("depth", func(t *testing.T) {
		entry, files := chainFiles(8)
		if _, diags, _ := fakeLoad(t, entry, files); len(diags) != 0 {
			t.Fatalf("a chain of 8: %v, want none", codesOf(diags))
		}
		entry, files = chainFiles(9)
		_, diags, fsys := fakeLoad(t, entry, files)
		if len(diags) != 1 || diags[0].Code != "module.depth_limit" {
			t.Fatalf("a chain of 9: %v, want module.depth_limit", codesOf(diags))
		}
		if fsys.opens != 8 {
			t.Fatalf("the 9th module was opened: opens = %d, want 8", fsys.opens)
		}
	})

	t.Run("total-bytes", func(t *testing.T) {
		const mebibyte = 1 << 20
		build := func(withFifth bool) (string, map[string]string) {
			names := []string{"a", "b", "c"}
			if withFifth {
				names = append(names, "d")
			}
			files := map[string]string{}
			for _, name := range names[:3] {
				files[name+".schway"] = padTo(t, libSource(name), mebibyte)
			}
			if withFifth {
				files["d.schway"] = "x"
			}
			return padTo(t, entryUsing(names...), mebibyte), files
		}
		entry, files := build(false)
		total := len(entry)
		for _, content := range files {
			total += len(content)
		}
		if total != 4194304 {
			t.Fatalf("fixture totals %d bytes, want 4194304", total)
		}
		if _, diags, _ := fakeLoad(t, entry, files); len(diags) != 0 {
			t.Fatalf("exactly 4,194,304 bytes: %v, want none", codesOf(diags))
		}
		entry, files = build(true)
		_, diags, _ := fakeLoad(t, entry, files)
		if len(diags) != 1 || diags[0].Code != "module.total_bytes_limit" {
			t.Fatalf("4,194,305 bytes: %v, want module.total_bytes_limit", codesOf(diags))
		}
	})

	t.Run("merged-program", func(t *testing.T) {
		functions := func(prefix string, count int) string {
			var out strings.Builder
			for index := 0; index < count; index++ {
				fmt.Fprintf(&out, "fn %s%d(n: U64) -> U64 {\n  n\n}\n", prefix, index)
			}
			return out.String()
		}
		entry := func(fillers int, data int) string {
			var out strings.Builder
			out.WriteString(entryUsing("lib"))
			out.WriteString(functions("e", fillers))
			for index := 0; index < data; index++ {
				fmt.Fprintf(&out, "data D%d =\n  | One\n", index)
			}
			return out.String()
		}
		library := func(count int) map[string]string {
			return map[string]string{"lib.schway": libSource("lib") + functions("g", count-1)}
		}
		// 1 (main) + 511 entry fillers + 512 library functions = 1,024.
		if _, diags, _ := fakeLoad(t, entry(511, 0), library(512)); len(diags) != 0 {
			t.Fatalf("1,024 functions: %v, want none", codesOf(diags))
		}
		_, diags, _ := fakeLoad(t, entry(511, 0), library(513))
		limit, ok := findCode(diags, "module.program_limit")
		if !ok || limit.Unit != "lib" {
			t.Fatalf("1,025 functions: %+v, want module.program_limit in lib", diags)
		}
		// Declarations: 3,072 data + 1 main + 999 fillers + 24 library functions.
		if _, diags, _ := fakeLoad(t, entry(999, 3072), library(24)); len(diags) != 0 {
			t.Fatalf("4,096 declarations: %v, want none", codesOf(diags))
		}
		if _, diags, _ := fakeLoad(t, entry(999, 3072), library(25)); !hasCode(diags, "module.program_limit") {
			t.Fatalf("4,097 declarations: %v, want module.program_limit", codesOf(diags))
		}
	})

	t.Run("off-by-one-control", func(t *testing.T) {
		capOffByOneForTest = true
		defer func() { capOffByOneForTest = false }()
		entry, files := leafFiles(15)
		if _, diags, _ := fakeLoad(t, entry, files); !hasCode(diags, "module.count_limit") {
			t.Fatalf("exact-16 modules not refused under the mutant: %v", codesOf(diags))
		}
		entry, files = chainFiles(8)
		if _, diags, _ := fakeLoad(t, entry, files); !hasCode(diags, "module.depth_limit") {
			t.Fatalf("exact-8 depth not refused under the mutant: %v", codesOf(diags))
		}
		const mebibyte = 1 << 20
		files = map[string]string{}
		for _, name := range []string{"a", "b", "c"} {
			files[name+".schway"] = padTo(t, libSource(name), mebibyte)
		}
		if _, diags, _ := fakeLoad(t, padTo(t, entryUsing("a", "b", "c"), mebibyte), files); !hasCode(diags, "module.total_bytes_limit") {
			t.Fatalf("exact-4,194,304 bytes not refused under the mutant: %v", codesOf(diags))
		}
	})
}

func TestPhase30CapPrecedence(t *testing.T) {
	entry, files := chainFiles(9)
	// A back edge (a cycle) and a missing import alongside the over-deep chain.
	files["m2.schway"] = libSource("m2", "m3", "m1", "missing")
	_, diags, _ := fakeLoad(t, entry, files)
	if len(diags) != 1 || diags[0].Code != "module.depth_limit" {
		t.Fatalf("diagnostics = %v, want only module.depth_limit", codesOf(diags))
	}
	// Without the 9th module the same graph reports the cycle and the missing
	// module, which proves the precedence is what hid them.
	entry, files = chainFiles(8)
	files["m2.schway"] = libSource("m2", "m3", "m1", "missing")
	_, diags, _ = fakeLoad(t, entry, files)
	if !hasCode(diags, "module.import_cycle") || !hasCode(diags, "module.not_found") {
		t.Fatalf("control diagnostics = %v, want the cycle and not_found", codesOf(diags))
	}
}

func TestPhase30MergeCanonicalOrder(t *testing.T) {
	build := func(order []string) (Set, []diagnostic.Diagnostic) {
		files := map[string]string{"a.schway": libSource("a"), "b.schway": libSource("b")}
		return func() (Set, []diagnostic.Diagnostic) {
			set, diags, _ := fakeLoad(t, entryUsing(order...), files)
			return set, diags
		}()
	}
	// Same-length import text keeps every entry span identical, so only the
	// import declaration order can differ.
	forward, forwardDiags := build([]string{"a", "b"})
	reverse, reverseDiags := build([]string{"b", "a"})
	if len(forwardDiags) != 0 || len(reverseDiags) != 0 {
		t.Fatalf("diagnostics: %v / %v", codesOf(forwardDiags), codesOf(reverseDiags))
	}
	owners := func(program ast.Program) []string {
		var out []string
		for _, function := range program.Funcs {
			out = append(out, function.Module+"/"+function.Name)
		}
		return out
	}
	if want := []string{"/main", "a/f", "b/f"}; !reflect.DeepEqual(owners(forward.Program), want) || !reflect.DeepEqual(owners(reverse.Program), want) {
		t.Fatalf("funcs = %v / %v, want %v", owners(forward.Program), owners(reverse.Program), want)
	}
	strip := func(program ast.Program) ast.Program {
		program.Imports = nil
		program.Units = append([]ast.Unit(nil), program.Units...)
		program.Units[0].Imports = nil
		return program
	}
	if !reflect.DeepEqual(strip(forward.Program), strip(reverse.Program)) {
		t.Fatalf("reversing the import declarations changed the merged program")
	}

	t.Run("mergeInImportOrder-control", func(t *testing.T) {
		restore := SetMergeInImportOrderForTest(true)
		defer restore()
		mutantForward, _ := build([]string{"a", "b"})
		mutantReverse, _ := build([]string{"b", "a"})
		if reflect.DeepEqual(owners(mutantForward.Program), owners(mutantReverse.Program)) {
			t.Fatalf("the import-order mutant left the merged order unchanged: %v", owners(mutantReverse.Program))
		}
	})
}
