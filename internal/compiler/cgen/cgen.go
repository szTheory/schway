package cgen

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/execution"
)

const Schema = "lang.c17/0"

func Emit(program core.Program) (string, error) {
	validated := corevalidate.Validate(program)
	if !validated.Valid {
		return "", fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
	}
	program = validated.Program()
	if len(program.Functions) != 1 {
		return "", fmt.Errorf("C emitter expects one function")
	}
	if program.Functions[0].Linear != nil {
		return emitLinear(program.Functions[0])
	}
	return emitMatch(program, false)
}

// EmitNative keeps the Phase 1 evidence bytes frozen while selecting the
// single-document execution protocol required at the child-process boundary.
func EmitNative(program core.Program) (string, error) {
	validated := corevalidate.Validate(program)
	if !validated.Valid {
		return "", fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
	}
	program = validated.Program()
	if len(program.Functions) != 1 {
		return "", fmt.Errorf("C emitter expects one function")
	}
	if program.Functions[0].Linear != nil {
		return emitLinear(program.Functions[0])
	}
	return emitMatch(program, true)
}

// The generated-C ordinary-identifier namespace is closed by two cooperating
// properties, not by the reserved lists alone. Both are load-bearing and both
// are enforced by tests (see cgen_names_test.go and names_internal_test.go):
//
//  1. PREFIX CONFINEMENT. Every identifier the allocator can ever hand out is
//     confined to one of exactly two namespaces: cName always returns
//     "LANG_" + <[A-Z0-9_]* tail> and cLocal always returns
//     "lang_value_" + <[A-Za-z0-9_]* tail>. The derived preferred names built
//     on top of them (alternative names, the "<Type>_name" helper) keep the
//     "LANG_" prefix, and cNames.allocate only ever appends to a preferred
//     name, so every allocated identifier stays inside those two namespaces.
//
//  2. HONEST RESERVATION. matchFixedNames and linearFixedNames are supersets of
//     every ordinary identifier the emitters write out themselves. A fixed
//     identifier that falls inside the "LANG_"/"lang_value_" namespaces is
//     therefore also reserved, so the allocator can never re-issue it.
//
// Property 1 alone would make most of the reservations unreachable, and
// property 2 alone would be enough only if it were maintained perfectly. Keeping
// both — and testing both — means a future emitter change breaks a test rather
// than silently aliasing two distinct core identities onto one C identifier.
//
// The lists are deliberately supersets: a name is kept even when no current
// emitter path writes it (for example "input" in the linear emitter, or
// LANG_BUFFER in the non-Buffer lowering). Over-reservation is inert, because
// property 1 guarantees no preferred name can equal a lowercase or
// non-"LANG_"-prefixed reserved entry; under-reservation is the dangerous
// direction, so the tests only forbid that one.

// matchFixedNames is every ordinary identifier emitMatch writes itself,
// excluding C keywords and the libc names it calls.
var matchFixedNames = []string{
	"main", "argc", "argv", "input", "output", "name", "value",
}

// linearFixedNames is every ordinary identifier emitLinear and
// emitLinearOutputSupport write themselves, excluding C keywords and the libc
// names they call. It covers the macros, typedefs, struct members, globals,
// helper functions, and every helper parameter and local.
var linearFixedNames = []string{
	// macros and typedefs
	"LANG_BUFFER", "LANG_EVENT", "LANG_OUTPUT_LIMIT", "LANG_EVENT_CAPACITY",
	// LANG_BUFFER and LANG_EVENT struct members
	"bytes", "length",
	"kind", "id", "function_id", "source_place", "target_place", "type_id",
	// file-scope globals
	"lang_events", "lang_event_count", "lang_output_count",
	// helper functions
	"lang_write_bytes", "lang_write_literal", "lang_write_json_string",
	"lang_record_event", "lang_write_events", "lang_write_buffer_hex", "lang_write_byte",
	// helper parameters and locals
	"data", "value", "hex", "byte", "escape", "encoded", "event", "index",
	// main
	"main", "argc", "argv", "input",
}

func emitMatch(program core.Program, executionJSON bool) (string, error) {
	if len(program.DataTypes) != 1 {
		return "", fmt.Errorf("match C emitter expects one data type")
	}
	dataType := program.DataTypes[0]
	function := program.Functions[0]
	if len(dataType.Alternatives) == 0 {
		return "", fmt.Errorf("S1 C emitter requires alternatives")
	}

	names := newCNames(matchFixedNames...)
	typeName := names.allocate(cName(dataType.Name), "type", 0)
	alternativeNames := make([]string, len(dataType.Alternatives))
	alternativeBySource := make(map[string]string, len(dataType.Alternatives))
	for index, alternative := range dataType.Alternatives {
		alternativeNames[index] = names.allocate(typeName+"_"+cName(alternative), "alternative", index)
		alternativeBySource[alternative] = alternativeNames[index]
	}
	functionName := names.allocate(cName(function.Name), "function", 0)
	parameterName := names.allocate(cName(function.Parameter.Name), "parameter", 0)
	helperName := names.allocate(typeName+"_name", "type_name", 0)

	var out strings.Builder
	out.WriteString("/* generated by Codename Lang; schema lang.c17/0 */\n")
	out.WriteString("#include <stdio.h>\n#include <string.h>\n\n")
	fmt.Fprintf(&out, "typedef enum %s {\n", typeName)
	for index := range dataType.Alternatives {
		fmt.Fprintf(&out, "  %s = %d,\n", alternativeNames[index], index)
	}
	fmt.Fprintf(&out, "} %s;\n\n", typeName)
	fmt.Fprintf(&out, "static %s %s(%s %s) {\n", typeName, functionName, typeName, parameterName)
	fmt.Fprintf(&out, "  switch (%s) {\n", parameterName)
	for _, arm := range function.Match.Arms {
		fmt.Fprintf(&out, "    case %s: return %s;\n", alternativeBySource[arm.Pattern], alternativeBySource[arm.Value])
	}
	out.WriteString("  }\n  return (int)255; /* invalid safe-language tag: fail in main */\n}\n\n")
	fmt.Fprintf(&out, "static const char *%s(%s value) {\n  switch (value) {\n", helperName, typeName)
	for _, alternative := range dataType.Alternatives {
		fmt.Fprintf(&out, "    case %s: return %s;\n", alternativeBySource[alternative], strconv.Quote(alternative))
	}
	out.WriteString("  }\n  return NULL;\n}\n\n")
	out.WriteString("int main(int argc, char **argv) {\n  if (argc != 2) return 64;\n")
	fmt.Fprintf(&out, "  %s input;\n", typeName)
	for index, alternative := range dataType.Alternatives {
		prefix := "if"
		if index > 0 {
			prefix = "else if"
		}
		fmt.Fprintf(&out, "  %s (strcmp(argv[1], %s) == 0) input = %s;\n", prefix, strconv.Quote(alternative), alternativeNames[index])
	}
	out.WriteString("  else return 65;\n")
	fmt.Fprintf(&out, "  %s output = %s(input);\n", typeName, functionName)
	fmt.Fprintf(&out, "  const char *name = %s(output);\n", helperName)
	if !executionJSON {
		out.WriteString("  if (name == NULL) return 70;\n  printf(\"%s\\n\", name);\n  return 0;\n}\n")
		return out.String(), nil
	}
	out.WriteString("  if (name == NULL) return 70;\n")
	for index, arm := range function.Match.Arms {
		prefix := "if"
		if index > 0 {
			prefix = "else if"
		}
		document := execution.Execution{
			Schema:  execution.Schema0,
			Outcome: execution.Outcome{Kind: "returned", Value: arm.Value},
			Events: []execution.Event{{
				Schema: execution.Schema0, ID: arm.ID + ":event:returned", Kind: "function.returned",
				FunctionID: function.ID, Input: arm.Pattern, Output: arm.Value,
			}},
			LiveResources: []string{},
		}
		encoded, err := json.Marshal(document)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&out, "  %s (strcmp(argv[1], %s) == 0) puts(%s);\n", prefix, strconv.Quote(arm.Pattern), strconv.Quote(string(encoded)))
	}
	out.WriteString("  else return 71;\n  return 0;\n}\n")
	return out.String(), nil
}

func emitLinear(function core.Function) (string, error) {
	input, initializer, typeName, err := linearInput(function)
	if err != nil {
		return "", err
	}
	places := make(map[string]core.Place, len(function.Linear.Places))
	for _, place := range function.Linear.Places {
		places[place.ID] = place
	}
	parameter, ok := places[function.Parameter.ID]
	if !ok {
		return "", fmt.Errorf("linear parameter place is absent")
	}
	names := newCNames(linearFixedNames...)
	placeIDs := make([]string, len(function.Linear.Places))
	for index, place := range function.Linear.Places {
		placeIDs[index] = place.ID
	}
	locals := make(map[string]string, len(placeIDs))
	for index, id := range placeIDs {
		locals[id] = names.allocate(cLocal(function.Linear.Places[index].Name), "place", index)
	}

	var out strings.Builder
	out.WriteString("/* generated by Codename Lang; schema lang.c17/0 */\n")
	out.WriteString("/* Moves below are authority transitions; C value assignment makes no ABI or zero-copy claim. */\n")
	out.WriteString("#include <stddef.h>\n#include <stdio.h>\n#include <string.h>\n\n")
	if function.Parameter.Type == "Buffer" {
		out.WriteString("typedef struct LANG_BUFFER {\n  unsigned char bytes[4];\n  size_t length;\n} LANG_BUFFER;\n\n")
	}
	emitLinearOutputSupport(&out, function, typeName)
	out.WriteString("int main(int argc, char **argv) {\n")
	out.WriteString("  if (argc != 2) return 64;\n")
	fmt.Fprintf(&out, "  if (strcmp(argv[1], %s) != 0) return 65;\n", strconv.Quote(input))
	fmt.Fprintf(&out, "  %s %s = %s;\n", typeName, locals[parameter.ID], initializer)
	declared := map[string]bool{parameter.ID: true}
	for _, operation := range function.Linear.Operations {
		source := places[operation.SourceID]
		switch operation.Kind {
		case core.OpCopy, core.OpMove, core.OpBorrowShared:
			target, exists := places[operation.TargetID]
			if !exists || declared[operation.TargetID] {
				return "", fmt.Errorf("operation %q has invalid target", operation.ID)
			}
			label := "copy"
			if operation.Kind == core.OpMove {
				label = "authority transfer"
			} else if operation.Kind == core.OpBorrowShared {
				label = "shared borrow representation"
			}
			fmt.Fprintf(&out, "  %s %s = %s; /* %s: %s */\n", typeName, locals[target.ID], locals[source.ID], label, operation.ID)
			fmt.Fprintf(&out, "  (void)%s;\n", locals[target.ID])
			eventKind := "value.copied"
			if operation.Kind == core.OpMove {
				eventKind = "value.transferred"
			} else if operation.Kind == core.OpBorrowShared {
				eventKind = "value.borrowed"
			}
			fmt.Fprintf(&out, "  if (!lang_record_event(%s, %s, %s, %s, %s, %s)) return 74;\n",
				strconv.Quote(eventKind), strconv.Quote(operation.ID+":event"), strconv.Quote(function.ID),
				strconv.Quote(operation.SourceID), strconv.Quote(operation.TargetID), strconv.Quote(operation.TypeID))
			declared[operation.TargetID] = true
		case core.OpReturn:
			fmt.Fprintf(&out, "  if (!lang_record_event(%s, %s, %s, %s, NULL, %s)) return 74; /* returned place: %s */\n",
				strconv.Quote("function.returned"), strconv.Quote(operation.ID+":event:returned"), strconv.Quote(function.ID),
				strconv.Quote(operation.SourceID), strconv.Quote(operation.TypeID), operation.ID)
			out.WriteString("  if (!lang_write_literal(\"{\\\"schema\\\":\\\"lang.execution/1\\\",\\\"outcome\\\":{\\\"kind\\\":\\\"returned\\\",\\\"value\\\":\\\"\")) return 74;\n")
			if function.Parameter.Type == "Buffer" {
				fmt.Fprintf(&out, "  if (!lang_write_buffer_hex(&%s)) return 74;\n", locals[source.ID])
			} else {
				fmt.Fprintf(&out, "  if (!lang_write_byte(%s)) return 74;\n", locals[source.ID])
			}
			out.WriteString("  if (!lang_write_literal(\"\\\"},\\\"events\\\":[\")) return 74;\n")
			out.WriteString("  if (!lang_write_events()) return 74;\n")
			out.WriteString("  if (!lang_write_literal(\"],\\\"live_resources\\\":[]}\\n\")) return 74;\n")
		default:
			return "", fmt.Errorf("operation %q has unknown kind %q", operation.ID, operation.Kind)
		}
	}
	out.WriteString("  return 0;\n}\n")
	return out.String(), nil
}

func emitLinearOutputSupport(out *strings.Builder, function core.Function, typeName string) {
	fmt.Fprintf(out, "#define LANG_OUTPUT_LIMIT 65536u\n#define LANG_EVENT_CAPACITY %du\n\n", len(function.Linear.Operations))
	out.WriteString("typedef struct LANG_EVENT {\n")
	out.WriteString("  const char *kind;\n  const char *id;\n  const char *function_id;\n")
	out.WriteString("  const char *source_place;\n  const char *target_place;\n  const char *type_id;\n} LANG_EVENT;\n\n")
	out.WriteString("static LANG_EVENT lang_events[LANG_EVENT_CAPACITY];\n")
	out.WriteString("static size_t lang_event_count = 0u;\nstatic size_t lang_output_count = 0u;\n\n")
	out.WriteString("static int lang_write_bytes(const char *data, size_t length) {\n")
	out.WriteString("  if (length > LANG_OUTPUT_LIMIT - lang_output_count) return 0;\n")
	out.WriteString("  if (length != 0u && fwrite(data, 1u, length, stdout) != length) return 0;\n")
	out.WriteString("  lang_output_count += length;\n  return 1;\n}\n\n")
	out.WriteString("static int lang_write_literal(const char *value) {\n  return lang_write_bytes(value, strlen(value));\n}\n\n")
	out.WriteString("static int lang_write_json_string(const char *value) {\n")
	out.WriteString("  static const char hex[] = \"0123456789abcdef\";\n")
	out.WriteString("  if (!lang_write_bytes(\"\\\"\", 1u)) return 0;\n")
	out.WriteString("  for (; *value != '\\0'; value++) {\n")
	out.WriteString("    unsigned char byte = (unsigned char)*value;\n")
	out.WriteString("    const char *escape = NULL;\n")
	out.WriteString("    if (byte == '\"') escape = \"\\\\\\\"\";\n")
	out.WriteString("    else if (byte == '\\\\') escape = \"\\\\\\\\\";\n")
	out.WriteString("    else if (byte == '\\b') escape = \"\\\\b\";\n")
	out.WriteString("    else if (byte == '\\f') escape = \"\\\\f\";\n")
	out.WriteString("    else if (byte == '\\n') escape = \"\\\\n\";\n")
	out.WriteString("    else if (byte == '\\r') escape = \"\\\\r\";\n")
	out.WriteString("    else if (byte == '\\t') escape = \"\\\\t\";\n")
	out.WriteString("    if (escape != NULL) { if (!lang_write_literal(escape)) return 0; }\n")
	out.WriteString("    else if (byte < 0x20u) {\n")
	out.WriteString("      char encoded[6] = {'\\\\', 'u', '0', '0', hex[byte >> 4u], hex[byte & 0x0fu]};\n")
	out.WriteString("      if (!lang_write_bytes(encoded, sizeof encoded)) return 0;\n")
	out.WriteString("    } else if (!lang_write_bytes(value, 1u)) return 0;\n")
	out.WriteString("  }\n  return lang_write_bytes(\"\\\"\", 1u);\n}\n\n")
	out.WriteString("static int lang_record_event(const char *kind, const char *id, const char *function_id, const char *source_place, const char *target_place, const char *type_id) {\n")
	out.WriteString("  if (lang_event_count >= LANG_EVENT_CAPACITY) return 0;\n")
	out.WriteString("  lang_events[lang_event_count++] = (LANG_EVENT){kind, id, function_id, source_place, target_place, type_id};\n")
	out.WriteString("  return 1;\n}\n\n")
	out.WriteString("static int lang_write_events(void) {\n")
	out.WriteString("  size_t index;\n  for (index = 0u; index < lang_event_count; index++) {\n")
	out.WriteString("    const LANG_EVENT *event = &lang_events[index];\n")
	out.WriteString("    if (index != 0u && !lang_write_bytes(\",\", 1u)) return 0;\n")
	out.WriteString("    if (!lang_write_literal(\"{\\\"schema\\\":\\\"lang.execution/1\\\",\\\"id\\\":\") || !lang_write_json_string(event->id)) return 0;\n")
	out.WriteString("    if (!lang_write_literal(\",\\\"kind\\\":\") || !lang_write_json_string(event->kind)) return 0;\n")
	out.WriteString("    if (!lang_write_literal(\",\\\"function_id\\\":\") || !lang_write_json_string(event->function_id)) return 0;\n")
	out.WriteString("    if (event->source_place != NULL && (!lang_write_literal(\",\\\"source_place\\\":\") || !lang_write_json_string(event->source_place))) return 0;\n")
	out.WriteString("    if (event->target_place != NULL && (!lang_write_literal(\",\\\"target_place\\\":\") || !lang_write_json_string(event->target_place))) return 0;\n")
	out.WriteString("    if (event->type_id != NULL && (!lang_write_literal(\",\\\"type_id\\\":\") || !lang_write_json_string(event->type_id))) return 0;\n")
	out.WriteString("    if (!lang_write_bytes(\"}\", 1u)) return 0;\n  }\n  return 1;\n}\n\n")
	if function.Parameter.Type == "Buffer" {
		fmt.Fprintf(out, "static int lang_write_buffer_hex(const %s *value) {\n", typeName)
		out.WriteString("  static const char hex[] = \"0123456789abcdef\";\n  size_t index;\n")
		out.WriteString("  if (value->length > sizeof value->bytes) return 0;\n")
		out.WriteString("  for (index = 0u; index < value->length; index++) {\n")
		out.WriteString("    char encoded[2] = {hex[value->bytes[index] >> 4u], hex[value->bytes[index] & 0x0fu]};\n")
		out.WriteString("    if (!lang_write_bytes(encoded, sizeof encoded)) return 0;\n  }\n  return 1;\n}\n\n")
	} else {
		out.WriteString("static int lang_write_byte(unsigned char value) {\n")
		out.WriteString("  char encoded[3];\n  int length = snprintf(encoded, sizeof encoded, \"%u\", (unsigned int)value);\n")
		out.WriteString("  return length > 0 && (size_t)length < sizeof encoded && lang_write_bytes(encoded, (size_t)length);\n}\n\n")
	}
}

func linearInput(function core.Function) (input, initializer, typeName string, err error) {
	switch function.Parameter.Type {
	case "Buffer":
		return "01020304", "{{1u, 2u, 3u, 4u}, 4u}", "LANG_BUFFER", nil
	case "Byte":
		return "7", "7u", "unsigned char", nil
	default:
		return "", "", "", fmt.Errorf("unsupported linear C type %q", function.Parameter.Type)
	}
}

type cNames struct{ used map[string]struct{} }

// newCNames builds the single global identifier allocator for one emitted
// translation unit. The reserved argument must be a superset of every fixed
// ordinary identifier the caller emits (see matchFixedNames /
// linearFixedNames); combined with the prefix confinement of cName and cLocal
// this is what makes the generated-C ordinary-identifier namespace closed.
func newCNames(reserved ...string) *cNames {
	result := &cNames{used: make(map[string]struct{}, len(reserved))}
	for _, name := range reserved {
		result.used[name] = struct{}{}
	}
	return result
}

func (n *cNames) allocate(preferred, category string, ordinal int) string {
	if _, exists := n.used[preferred]; !exists {
		n.used[preferred] = struct{}{}
		return preferred
	}
	base := preferred + "__LANG_" + strings.ToUpper(category) + "_" + strconv.Itoa(ordinal)
	for attempt, candidate := 0, base; ; attempt, candidate = attempt+1, base+"_"+strconv.Itoa(attempt) {
		if _, exists := n.used[candidate]; !exists {
			n.used[candidate] = struct{}{}
			return candidate
		}
	}
}

// cName maps a source type, alternative, function, or parameter name into the
// uppercase "LANG_" namespace. INVARIANT (enforced by
// TestGeneratedIdentifierNamespacesStayConfined): the result always matches
// ^LANG_[A-Z0-9_]*$. Do not relax the uppercasing or the prefix — the closure
// argument for the generated-C namespace depends on every allocated identifier
// living in the "LANG_" or "lang_value_" namespace and nowhere else.
func cName(name string) string {
	var out strings.Builder
	for _, r := range name {
		if r >= 'a' && r <= 'z' {
			out.WriteRune(r - ('a' - 'A'))
		} else if r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' {
			out.WriteRune(r)
		} else {
			out.WriteByte('_')
		}
	}
	return "LANG_" + out.String()
}

// cLocal maps a source place name into the "lang_value_" namespace. INVARIANT
// (enforced by TestGeneratedIdentifierNamespacesStayConfined): the result
// always matches ^lang_value_[A-Za-z0-9_]*$. No fixed identifier emitted by any
// emitter may be placed in this namespace unless it is also reserved.
func cLocal(name string) string {
	var out strings.Builder
	out.WriteString("lang_value_")
	for _, r := range name {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' {
			out.WriteRune(r)
		} else {
			out.WriteByte('_')
		}
	}
	return out.String()
}
