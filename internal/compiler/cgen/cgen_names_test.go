package cgen_test

import (
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/session"
)

// This file enforces the two properties that make the generated-C ordinary
// identifier namespace closed (documented above matchFixedNames in cgen.go):
//
//	1. PREFIX CONFINEMENT — every allocated identifier lives in the "LANG_" or
//	   "lang_value_" namespace and nowhere else.
//	2. HONEST RESERVATION — every fixed identifier the emitters write themselves
//	   is in the emitter's reserved set.
//
// The fixed-identifier set is DERIVED FROM THE GENERATED OUTPUT rather than
// hand-listed, so it cannot drift away from what the emitters actually emit.
// Two structurally identical programs whose source names are pairwise disjoint
// are emitted; identifiers present in BOTH outputs are exactly the ones the
// emitter contributes itself, and identifiers present in only one are exactly
// the ones the allocator derived from source.

// cExternalNames are the C keywords, standard types, and libc identifiers the
// emitters reference but do not own. They are the only identifiers allowed to
// appear in generated output without being reserved. Adding a new libc call to
// an emitter must be a deliberate edit here, not a silent one.
var cExternalNames = map[string]bool{
	"break": true, "case": true, "char": true, "const": true, "default": true,
	"else": true, "enum": true, "for": true, "if": true, "int": true,
	"return": true, "sizeof": true, "static": true, "struct": true,
	"switch": true, "typedef": true, "unsigned": true, "void": true,
	"while": true,

	"NULL": true, "size_t": true, "stdout": true,
	"fwrite": true, "printf": true, "puts": true,
	"snprintf": true, "strcmp": true, "strlen": true,
}

// identifierPattern matches a C ordinary identifier.
var identifierPattern = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*`)

// cIdentifiers returns every ordinary identifier in generated C, with comments,
// string literals, character literals, and #include header names removed so
// that literal payload (JSON keys, source-derived alternative spellings) is
// never mistaken for an identifier.
func cIdentifiers(t *testing.T, generated string) map[string]bool {
	t.Helper()
	var stripped strings.Builder
	runes := []rune(generated)
	for index := 0; index < len(runes); {
		switch {
		case runes[index] == '/' && index+1 < len(runes) && runes[index+1] == '*':
			end := strings.Index(string(runes[index:]), "*/")
			if end < 0 {
				t.Fatalf("generated C has an unterminated block comment")
			}
			index += len([]rune(string(runes[index:])[:end+2]))
		case runes[index] == '"' || runes[index] == '\'':
			quote := runes[index]
			index++
			for index < len(runes) && runes[index] != quote {
				if runes[index] == '\\' {
					index++
				}
				index++
			}
			if index >= len(runes) {
				t.Fatalf("generated C has an unterminated literal")
			}
			index++
			stripped.WriteRune(' ')
		case runes[index] == '<' && strings.HasSuffix(strings.TrimRight(stripped.String(), " \t"), "#include"):
			for index < len(runes) && runes[index] != '\n' {
				index++
			}
		default:
			stripped.WriteRune(runes[index])
			index++
		}
	}
	text := stripped.String()
	found := make(map[string]bool)
	for _, span := range identifierPattern.FindAllStringIndex(text, -1) {
		// Skip the tail of a numeric literal: C spells suffixes and hex digits
		// inline (1u, 65536u, 0x0fu), and the identifier pattern would report
		// "u"/"x0fu" as identifiers. No real C identifier can follow a digit.
		if span[0] > 0 && text[span[0]-1] >= '0' && text[span[0]-1] <= '9' {
			continue
		}
		found[text[span[0]:span[1]]] = true
	}
	delete(found, "include")
	delete(found, "define")
	return found
}

func emitAll(t *testing.T, source string) []string {
	t.Helper()
	checked := session.Check([]byte(source))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture did not check: %+v", checked.Diagnostics)
	}
	portable, err := cgen.Emit(checked.Program)
	if err != nil {
		t.Fatalf("Emit: %v", err)
	}
	native, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatalf("EmitNative: %v", err)
	}
	return []string{portable, native}
}

func identifiersOf(t *testing.T, source string) map[string]bool {
	t.Helper()
	union := make(map[string]bool)
	for _, generated := range emitAll(t, source) {
		for identifier := range cIdentifiers(t, generated) {
			union[identifier] = true
		}
	}
	return union
}

func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// TestGeneratedIdentifierNamespacesStayConfined is the load-bearing enforcement
// of the closure argument. It fails if an emitter starts writing a fixed
// identifier that is neither reserved nor a known external C name (for example
// a new helper local called lang_value_scratch), and it fails if any
// source-derived identifier escapes the two confined namespaces.
func TestGeneratedIdentifierNamespacesStayConfined(t *testing.T) {
	tests := []struct {
		name     string
		reserved []string
		left     string
		right    string
		// leftOnly are identifiers that must be classified source-derived, a
		// sanity check that the differential actually separates the two sets.
		leftOnly []string
	}{
		{
			name:     "match",
			reserved: cgen.MatchFixedNames,
			left:     "module a.one\n\nexport {\n  type Switch\n  fn toggle\n}\n\ndata Switch =\n  | Off\n  | On\n\nfn toggle(state: Switch) -> Switch {\n  match state {\n    Off => On\n    On => Off\n  }\n}\n",
			right:    "module z.two\n\nexport {\n  type Fixture\n  fn convert\n}\n\ndata Fixture =\n  | Alpha\n  | Beta\n\nfn convert(item: Fixture) -> Fixture {\n  match item {\n    Alpha => Beta\n    Beta => Alpha\n  }\n}\n",
			leftOnly: []string{"LANG_SWITCH", "LANG_SWITCH_LANG_OFF", "LANG_TOGGLE", "LANG_STATE", "LANG_SWITCH_name"},
		},
		{
			name:     "linear Buffer",
			reserved: cgen.LinearFixedNames,
			left:     "module owned.transfer\n\nexport {\n  fn relay\n}\n\nfn relay(buffer: Buffer) -> Buffer {\n  let delivered = take buffer\n  delivered\n}\n",
			right:    "module owned.shipment\n\nexport {\n  fn dispatch\n}\n\nfn dispatch(cargo: Buffer) -> Buffer {\n  let handed = take cargo\n  handed\n}\n",
			leftOnly: []string{"lang_value_buffer", "lang_value_delivered"},
		},
		{
			name:     "linear Byte",
			reserved: cgen.LinearFixedNames,
			left:     "module owned.copy\n\nexport {\n  fn retain\n}\n\nfn retain(code: Byte) -> Byte {\n  let kept = code\n  kept\n}\n",
			right:    "module owned.hold\n\nexport {\n  fn preserve\n}\n\nfn preserve(digit: Byte) -> Byte {\n  let saved = digit\n  saved\n}\n",
			leftOnly: []string{"lang_value_code", "lang_value_kept"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reserved := make(map[string]bool, len(test.reserved))
			for _, name := range test.reserved {
				reserved[name] = true
			}

			left := identifiersOf(t, test.left)
			right := identifiersOf(t, test.right)

			fixed := make(map[string]bool)
			for identifier := range left {
				if right[identifier] {
					fixed[identifier] = true
				}
			}
			if len(fixed) == 0 {
				t.Fatalf("differential produced no fixed identifiers; the two fixtures are not comparable")
			}
			for _, expected := range test.leftOnly {
				if !left[expected] {
					t.Fatalf("fixture does not emit expected source-derived identifier %q", expected)
				}
				if fixed[expected] {
					t.Fatalf("source-derived identifier %q was classified as fixed; fixtures are not name-disjoint", expected)
				}
			}

			// HONEST RESERVATION: every identifier the emitter writes itself is
			// reserved, or is an explicitly acknowledged external C name.
			var unreserved []string
			for identifier := range fixed {
				if reserved[identifier] || cExternalNames[identifier] {
					continue
				}
				unreserved = append(unreserved, identifier)
			}
			if len(unreserved) != 0 {
				sort.Strings(unreserved)
				t.Errorf("emitter writes fixed identifiers that are neither reserved nor known external C names: %v\n"+
					"reserve them in cgen.go (or, if they are libc, add them to cExternalNames)", unreserved)
			}

			// PREFIX CONFINEMENT: every identifier the allocator derived from
			// source lives in exactly one of the two confined namespaces.
			var escaped []string
			for identifier := range left {
				if fixed[identifier] || cExternalNames[identifier] {
					continue
				}
				if strings.HasPrefix(identifier, "LANG_") || strings.HasPrefix(identifier, "lang_value_") {
					continue
				}
				escaped = append(escaped, identifier)
			}
			if len(escaped) != 0 {
				sort.Strings(escaped)
				t.Errorf("source-derived identifiers escaped the LANG_/lang_value_ namespaces: %v", escaped)
			}

			t.Logf("fixed identifiers derived from generated output: %v", sortedKeys(fixed))
		})
	}
}

// TestReservedSetsCoverTheirOwnNamespace restates the closure argument as a
// direct check on the reserved lists: any reserved entry inside the confined
// namespaces is only safe because it is reserved, and any entry outside them is
// only safe because prefix confinement puts it out of the allocator's reach.
// A reserved entry must never be producible by cName or cLocal by accident.
func TestReservedSetsCoverTheirOwnNamespace(t *testing.T) {
	for _, group := range [][]string{cgen.MatchFixedNames, cgen.LinearFixedNames} {
		for _, name := range group {
			if strings.HasPrefix(name, "lang_value_") {
				t.Errorf("fixed identifier %q sits in the cLocal namespace; rename it or the closure argument becomes reservation-only", name)
			}
		}
	}
}

// TestIdentifierPrefixInvariance is the direct falsifier for property 1. It
// fails if cName stops uppercasing or loses its prefix, or if cLocal loses its
// prefix, or if the collision suffix moves an identifier out of its namespace.
func TestIdentifierPrefixInvariance(t *testing.T) {
	namePattern := regexp.MustCompile(`^LANG_[A-Z0-9_]*$`)
	localPattern := regexp.MustCompile(`^lang_value_[A-Za-z0-9_]*$`)

	sources := []string{
		"", "a", "A", "_", "0", "z9", "Switch", "switch", "main", "value", "hex",
		"lang_value_x", "lang_events", "LANG_EVENT", "lang_write_bytes",
		"α", "β", "日本語", "a b", "a-b", "a.b", "x__LANG_PLACE_1",
		"index", "data", "length", "encoded", "type_id", "function_id",
	}
	for _, source := range sources {
		if got := cgen.CName(source); !namePattern.MatchString(got) {
			t.Errorf("cName(%q) = %q, which breaks ^LANG_[A-Z0-9_]*$", source, got)
		}
		if got := cgen.CLocal(source); !localPattern.MatchString(got) {
			t.Errorf("cLocal(%q) = %q, which breaks ^lang_value_[A-Za-z0-9_]*$", source, got)
		}
	}

	// The collision suffix must keep an identifier inside its namespace, and
	// must never hand back a reserved name.
	reserved := append(append([]string{}, cgen.MatchFixedNames...), cgen.LinearFixedNames...)
	for _, source := range sources {
		for _, preferred := range []string{cgen.CName(source), cgen.CLocal(source)} {
			for _, category := range []string{"type", "alternative", "function", "parameter", "type_name", "place"} {
				got := cgen.AllocateFrom(append(reserved, preferred), preferred, category, 0)
				if got == preferred {
					t.Errorf("allocate re-issued reserved name %q", preferred)
				}
				if !strings.HasPrefix(got, preferred) {
					t.Errorf("allocate(%q) = %q left its namespace", preferred, got)
				}
				for _, name := range reserved {
					if got == name {
						t.Errorf("allocate(%q) = %q collided with reserved name", preferred, got)
					}
				}
			}
		}
	}
}
