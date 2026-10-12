package cgen

import (
	"regexp"
	"sort"
	"strings"
	"testing"
)

type gateNamePair struct{ module, function string }

func gateNameUniverse() []gateNamePair {
	segments := []string{"a", "b", "a_b", "b_c", "ab", "x1", "adler32", "m"}
	functions := []string{"f", "c", "b_c", "foo", "Foo", "FOO", "x_1", "adler32"}
	var modules []string
	var extend func(prefix []string, depth int)
	extend = func(prefix []string, depth int) {
		if len(prefix) > 0 {
			modules = append(modules, strings.Join(prefix, "."))
		}
		if depth == 0 {
			return
		}
		for _, segment := range segments {
			extend(append(append([]string(nil), prefix...), segment), depth-1)
		}
	}
	extend(nil, 3)
	var pairs []gateNamePair
	for _, module := range modules {
		for _, function := range functions {
			pairs = append(pairs, gateNamePair{module, function})
		}
	}
	return pairs
}

func gateCanonical(pairs []gateNamePair) []gateNamePair {
	out := append([]gateNamePair(nil), pairs...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].module != out[j].module {
			return out[i].module < out[j].module
		}
		return out[i].function < out[j].function
	})
	return out
}

// gateAllocate names pairs the way plan 30-07 will: canonical order, one
// allocate call per pair, entry names reserved first.
func gateAllocate(entryNames []string, pairs []gateNamePair) map[gateNamePair]string {
	names := newCNames()
	for index, entry := range entryNames {
		names.allocate(cName(entry), "function", index)
	}
	result := map[gateNamePair]string{}
	for index, pair := range gateCanonical(pairs) {
		result[pair] = names.allocate(moduleCName(pair.module, pair.function), "function", index)
	}
	return result
}

// gateCollapsedGroups returns name -> pairs for every name shared by 2+ pairs
// and reports whether any shared group is NOT a pure function-case collapse.
func gateCollapsedGroups(pairs []gateNamePair) (map[string][]gateNamePair, bool) {
	byName := map[string][]gateNamePair{}
	for _, pair := range pairs {
		name := moduleCName(pair.module, pair.function)
		byName[name] = append(byName[name], pair)
	}
	collapsed := map[string][]gateNamePair{}
	foreign := false
	for name, group := range byName {
		if len(group) < 2 {
			continue
		}
		collapsed[name] = group
		for _, other := range group[1:] {
			if other.module != group[0].module || !strings.EqualFold(other.function, group[0].function) {
				foreign = true
			}
		}
	}
	return collapsed, foreign
}

// TestPhase30GateCNamesInjective is flag (d) of the D-14 gate.
func TestPhase30GateCNamesInjective(t *testing.T) {
	if got := moduleCName("adler32", "adler32"); got != "SCHWAY_M_7_ADLER32_7_ADLER32" {
		t.Fatalf("moduleCName(adler32, adler32) = %q", got)
	}
	if got := moduleCName("a.b", "f"); got != "SCHWAY_M_1_A_1_B_1_F" {
		t.Fatalf("moduleCName(a.b, f) = %q", got)
	}

	pairs := gateNameUniverse()
	shape := regexp.MustCompile(`^SCHWAY_[A-Z0-9_]*$`)
	for _, pair := range pairs {
		name := moduleCName(pair.module, pair.function)
		if strings.Contains(name, "__") {
			t.Fatalf("%v -> %q contains a double underscore", pair, name)
		}
		if !shape.MatchString(name) {
			t.Fatalf("%v -> %q leaves the SCHWAY_ namespace", pair, name)
		}
	}

	collapsed, foreign := gateCollapsedGroups(pairs)
	if foreign {
		t.Fatalf("a name is shared by pairs that are not pure function-case variants: %v", collapsed)
	}
	if len(collapsed) == 0 {
		t.Fatal("expected the foo/Foo/FOO case-collapse set to be exercised")
	}
	t.Logf("universe %d pairs, %d case-collapsed names", len(pairs), len(collapsed))

	var collapseSet []gateNamePair
	for _, group := range collapsed {
		collapseSet = append(collapseSet, group...)
	}
	assigned := gateAllocate(nil, collapseSet)
	seen := map[string]gateNamePair{}
	for pair, name := range assigned {
		if prior, dup := seen[name]; dup {
			t.Fatalf("allocate gave %q to both %v and %v", name, prior, pair)
		}
		seen[name] = pair
	}

	ascending := gateCanonical(pairs)
	descending := append([]gateNamePair(nil), ascending...)
	for i, j := 0, len(descending)-1; i < j; i, j = i+1, j-1 {
		descending[i], descending[j] = descending[j], descending[i]
	}
	first, second := gateAllocate(nil, ascending), gateAllocate(nil, descending)
	for pair, name := range first {
		if second[pair] != name {
			t.Fatalf("allocation depends on input order: %v got %q then %q", pair, name, second[pair])
		}
	}

	t.Run("entry name wins", func(t *testing.T) {
		const entry = "m_7_adler32_7_adler32"
		if got := cName(entry); got != moduleCName("adler32", "adler32") {
			t.Fatalf("cName(%q) = %q, want the module-shaped name", entry, got)
		}
		result := gateAllocate([]string{entry}, []gateNamePair{{"adler32", "adler32"}})
		module := result[gateNamePair{"adler32", "adler32"}]
		if module == cName(entry) || !strings.HasPrefix(module, cName(entry)+"_SCHWAY_") {
			t.Fatalf("module function got %q; entry must keep %q and the module name must be suffixed", module, cName(entry))
		}
	})

	t.Run("control plain join collides", func(t *testing.T) {
		plainJoinModuleNameForTest = true
		defer func() { plainJoinModuleNameForTest = false }()
		if _, foreignCollision := gateCollapsedGroups(pairs); !foreignCollision {
			t.Fatal("the plain-join seam must produce a collision such as a_b.c against a.b_c")
		}
		if moduleCName("a_b.c", "f") != moduleCName("a.b_c", "f") {
			t.Fatal("expected a_b.c and a.b_c to collide under the plain join")
		}
	})
}
