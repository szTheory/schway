package cgen

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/codename-lang/lang/internal/compiler/callgraph"
	"github.com/codename-lang/lang/internal/compiler/core"
)

// callBoundaryAttributeSetForTest is
// TestEmittedAttributeSetCommentIsDerivedNotLiteral's own fault-injection
// seam (D-04-12): nil (the default, and the ONLY value any production
// caller ever observes) means "use emitCallBoundaryAttributeSet's real,
// always-empty derivation"; a non-nil slice substitutes a counterfactual
// non-empty set so a test can prove the rendered comment tracks the
// derivation itself rather than restating a literal beside it. Exported
// (via export_test.go's SetCallBoundaryAttributeSetForTest) only for the
// external cgen_test package; never set on a production path.
var callBoundaryAttributeSetForTest []string

// emitCallBoundaryAttributeSet is D-11-09's own named derivation point:
// under this package's architecture, emitCall (this file) never mints an
// optimizer-visible attribute at any Lang-to-Lang call boundary (D-11-04),
// so absent the test seam above this always returns an empty, non-nil
// slice. Every caller -- emitCallBoundaryAttributeComment below, and any
// future gate that asks the same question -- reads THIS one derivation,
// so a later change to emitCall's own emission automatically changes what
// every caller reports (D-04-12: no hand-written comment beside a
// generated value that can drift from it).
func emitCallBoundaryAttributeSet(functions []core.Function) []string {
	if callBoundaryAttributeSetForTest != nil {
		return append([]string{}, callBoundaryAttributeSetForTest...)
	}
	_ = functions
	return []string{}
}

// emitCallBoundaryAttributeComment renders
// emitCallBoundaryAttributeSet's own value (sorted, so the statement is
// byte-stable regardless of the program's own function declaration
// order) as a generated C comment naming the set's size, whether it is
// empty, and D-11-09 by identifier -- NAT-05's explicit, visible, empty
// call-boundary attribute set.
func emitCallBoundaryAttributeComment(functions []core.Function) string {
	set := append([]string{}, emitCallBoundaryAttributeSet(functions)...)
	sort.Strings(set)
	if len(set) == 0 {
		return "/* call-boundary attribute set: EMPTY (0 entries). Per D-11-09, cgen.emitCall never mints an optimizer-visible alias or capture attribute at any Lang-to-Lang call boundary this phase -- NAT-05 is satisfied by this explicit, generated empty set, not by silence. */\n"
	}
	return fmt.Sprintf("/* call-boundary attribute set: %d entries: %s (D-11-09). */\n", len(set), strings.Join(set, ", "))
}

// emitCallLookup carries the whole-program name/type tables emitCall needs
// to resolve an OpCall's CalleeID into its already-allocated generated C
// function name and C return-type name (D-11-04/D-11-08). A nil
// *emitCallLookup always fails to resolve -- this is deliberate, not an
// oversight: emitLinear's and emitBranchOperations' own OpCall arms (both
// single-function paths, cgen.go) call resolve on a nil lookup because a
// one-function program can never legally contain an OpCall (self-recursion
// is refused as a call-graph cycle before either arm could run), so those
// two call sites always take the "unresolved" branch. emitProgram (this
// file) is the one caller that ever supplies a real, populated lookup.
type emitCallLookup struct {
	indexByFunctionID map[string]int
	functionNames     []string
	returnTypeNames   []string
}

func (l *emitCallLookup) resolve(calleeID string) (functionName, returnTypeName string, ok bool) {
	if l == nil {
		return "", "", false
	}
	index, known := l.indexByFunctionID[calleeID]
	if !known {
		return "", "", false
	}
	return l.functionNames[index], l.returnTypeNames[index], true
}

// emitCall is D-11-04's single writer of a Lang-to-Lang call anywhere in
// cgen: it emits ONE C call expression assigning the resolved callee's
// result to the operation's own TargetID local, using the same inline
// provenance comment convention (`/* label: operationID */`) every
// surrounding straight-line arm already uses. It never mints an
// optimizer-visible attribute on either side of the call -- `restrict` is a
// property of a definition and its own prototype (D-05-01..D-05-04), never
// of a call site, and D-11-09 declares Phase 11 emits zero call-boundary
// alias attributes at all.
func emitCall(out *strings.Builder, calleeTypeName, targetLocal, calleeName, argumentLocal string, operation core.LinearOperation) {
	fmt.Fprintf(out, "  %s %s = %s(%s); /* call: %s */\n", calleeTypeName, targetLocal, calleeName, argumentLocal, operation.ID)
	fmt.Fprintf(out, "  (void)%s;\n", targetLocal)
}

// emitProgram is Phase 11's whole-program C17 translation-unit assembler
// (D-11-01/D-11-03): the additive path Emit and EmitNative dispatch to
// whenever a validated program declares more than one function. Per
// D-11-03 it owns every whole-TU concern itself -- includes, typedefs,
// every support block, the resource ledger, every function's prototype,
// every function's definition, and exactly one `main` invoking the
// program's resolved entry function (callgraph.EntryFunction) -- while
// each function's own body writes statements and events only.
//
// Order of emission: the #include block; the LANG_BUFFER typedef when any
// emitted function's parameter type is Buffer; emitEventSupport (called,
// never forked) sized to the WHOLE program's total operation count, since
// every function shares one events buffer; emitDefectSupport when any
// function contains an OpDefect; every function prototype, in
// callgraph.Order's order; every function definition, in the same order;
// then `main`.
//
// This phase's corpus is entirely straight-line (D-11-24's tracer scope):
// no core.Match, no core.Block-based control flow, and no
// core.ForeignContract on any function this assembler emits. A function
// carrying any of those shapes is refused here with a named error rather
// than silently mis-emitted -- nothing in this phase's testdata/phase11
// corpus exercises a multi-function branch or foreign body; that remains a
// documented scope limit for a later phase, not a live gap this one
// papers over.
//
// executionJSON is accepted for signature symmetry with Emit/EmitNative's
// existing emitMatch(program, executionJSON) call convention and for a
// future multi-function branch body; it is unused today because this
// phase's assembler always writes the lang.execution/1 JSON document (the
// single behavior emitLinear's own single-function path already has).
func emitProgram(program core.Program, executionJSON bool) (string, error) {
	_ = executionJSON

	order, err := callgraph.Order(program)
	if err != nil {
		return "", err
	}
	entry, err := callgraph.EntryFunction(program)
	if err != nil {
		return "", err
	}

	byID := make(map[string]core.Function, len(program.Functions))
	for _, function := range program.Functions {
		byID[function.ID] = function
	}

	functions := make([]core.Function, 0, len(order))
	for _, id := range order {
		function, ok := byID[id]
		if !ok {
			return "", fmt.Errorf("emitProgram: call-graph order names unknown function %q", id)
		}
		if function.Match != nil {
			return "", fmt.Errorf("function %q: multi-function branch bodies are not supported by native emission this phase", function.ID)
		}
		if function.Linear == nil {
			return "", fmt.Errorf("function %q: has no linear body", function.ID)
		}
		if len(function.Linear.Blocks) > 0 {
			return "", fmt.Errorf("function %q: multi-function foreign-call bodies are not supported by native emission this phase", function.ID)
		}
		if function.ForeignContract != nil {
			return "", fmt.Errorf("function %q: multi-function foreign contracts are not supported by native emission this phase", function.ID)
		}
		functions = append(functions, function)
	}

	// Two-tier name allocation (D-11-08/T-11-06): ONE global cNames
	// allocates every function's own C name, iterating in
	// callgraph.Order's order so allocation is deterministic. Each
	// function's own return/parameter C type name is derived the same way
	// linearInput already derives it for the single-function path.
	globals := newCNames(linearFixedNames...)
	functionNames := make([]string, len(functions))
	typeNames := make([]string, len(functions))
	for index, function := range functions {
		functionNames[index] = globals.allocate(cName(function.Name), "function", index)
		_, _, typeName, err := linearInput(function)
		if err != nil {
			return "", fmt.Errorf("function %q: %w", function.ID, err)
		}
		typeNames[index] = typeName
	}
	// globalNames snapshots every name the global allocator has handed out
	// (function names plus the reserved list) so each function's own FRESH
	// per-function cNames (below) is seeded with the reserved list PLUS
	// every globally allocated name -- this is what keeps the second
	// function's places named lang_value_x rather than lang_value_x_2
	// (D-11-08): C block scope already makes locals independent; the seed
	// is what makes shadowing impossible.
	globalNames := make([]string, 0, len(globals.used))
	for name := range globals.used {
		globalNames = append(globalNames, name)
	}

	indexByFunctionID := make(map[string]int, len(functions))
	for index, function := range functions {
		indexByFunctionID[function.ID] = index
	}
	lookup := &emitCallLookup{indexByFunctionID: indexByFunctionID, functionNames: functionNames, returnTypeNames: typeNames}

	needsBuffer := false
	needsByte := false
	needsDefect := false
	totalOperations := 0
	for _, function := range functions {
		switch function.Parameter.Type {
		case "Buffer":
			needsBuffer = true
		case "Byte":
			needsByte = true
		}
		if functionHasDefect(function) {
			needsDefect = true
		}
		totalOperations += len(function.Linear.Operations)
	}

	var out strings.Builder
	out.WriteString("/* generated by Codename Lang; schema lang.c17/0 */\n")
	out.WriteString("/* Moves below are authority transitions; C value assignment makes no ABI or zero-copy claim. */\n")
	out.WriteString(emitCallBoundaryAttributeComment(functions))
	out.WriteString("#include <stddef.h>\n#include <stdio.h>\n#include <stdlib.h>\n#include <string.h>\n\n")
	if needsBuffer {
		out.WriteString("typedef struct LANG_BUFFER {\n  unsigned char bytes[4];\n  size_t length;\n} LANG_BUFFER;\n\n")
	}
	emitEventSupport(&out, totalOperations)
	if needsDefect {
		emitDefectSupport(&out)
	}
	if needsBuffer {
		emitProgramBufferWriter(&out, "LANG_BUFFER")
	}
	if needsByte {
		emitProgramByteWriter(&out)
	}

	// Prototypes before definitions (D-11-03): what makes a forward-
	// referenced callee legal C17.
	for index := range functions {
		fmt.Fprintf(&out, "static %s %s(%s);\n", typeNames[index], functionNames[index], typeNames[index])
	}
	out.WriteString("\n")

	for index, function := range functions {
		if err := emitProgramFunction(&out, function, typeNames[index], functionNames[index], globalNames, lookup); err != nil {
			return "", err
		}
	}

	entryIndex, ok := indexByFunctionID[entry.ID]
	if !ok {
		return "", fmt.Errorf("emitProgram: resolved entry %q is absent from the emitted function set", entry.ID)
	}
	input, initializer, entryTypeName, err := linearInput(entry)
	if err != nil {
		return "", err
	}
	out.WriteString("int main(int argc, char **argv) {\n")
	// A declared-but-never-called function (D-11-05's unreachable-function
	// contract: it is still emitted as a real C definition, dead code that
	// does not change entry resolution) would otherwise be flagged
	// unused-function under this project's -Werror build. Taking each
	// function's own address here is a harmless, side-effect-free reference
	// -- never a call -- that keeps every emitted definition provably
	// reachable from a linker's perspective without changing which
	// function `main` actually invokes.
	for index := range functions {
		fmt.Fprintf(&out, "  (void)%s;\n", functionNames[index])
	}
	// lang_write_buffer_hex/lang_write_byte (emitProgramBufferWriter/
	// emitProgramByteWriter above) are emitted whenever ANY function in the
	// program declares that parameter type -- not only when the RESOLVED
	// ENTRY function does. A declared-but-never-called Buffer/Byte
	// function (D-11-05's own unreachable-function contract) whose type
	// differs from the entry's own type would otherwise leave the
	// corresponding writer genuinely unreferenced, failing this project's
	// -Werror -Wunused-function build (Rule 1: a real, reachable bug this
	// gate corpus's own touch function, declared Buffer-typed and never
	// called, first exposed). The same harmless, side-effect-free
	// address-taking reference used for user functions above applies here
	// too.
	if needsBuffer {
		out.WriteString("  (void)lang_write_buffer_hex;\n")
	}
	if needsByte {
		out.WriteString("  (void)lang_write_byte;\n")
	}
	out.WriteString("  if (argc != 2) return 64;\n")
	fmt.Fprintf(&out, "  if (strcmp(argv[1], %s) != 0) return 65;\n", strconv.Quote(input))
	fmt.Fprintf(&out, "  %s lang_entry_input = %s;\n", entryTypeName, initializer)
	fmt.Fprintf(&out, "  %s lang_entry_output = %s(lang_entry_input);\n", entryTypeName, functionNames[entryIndex])
	out.WriteString("  if (!lang_write_literal(\"{\\\"schema\\\":\\\"lang.execution/1\\\",\\\"outcome\\\":{\\\"kind\\\":\\\"returned\\\",\\\"value\\\":\\\"\")) return 74;\n")
	if entryTypeName == "LANG_BUFFER" {
		out.WriteString("  if (!lang_write_buffer_hex(&lang_entry_output)) return 74;\n")
	} else {
		out.WriteString("  if (!lang_write_byte(lang_entry_output)) return 74;\n")
	}
	out.WriteString("  if (!lang_write_literal(\"\\\"},\\\"events\\\":[\")) return 74;\n")
	out.WriteString("  if (!lang_write_events()) return 74;\n")
	out.WriteString("  if (!lang_write_literal(\"],\\\"live_resources\\\":[]}\\n\")) return 74;\n")
	out.WriteString("  return 0;\n}\n")

	return out.String(), nil
}

// emitProgramBufferWriter/emitProgramByteWriter are the multi-function
// assembler's own copies of emitLinearOutputSupport's scalar-value writer
// selection (cgen.go): a Buffer-parameterized program needs
// lang_write_buffer_hex, a Byte-parameterized one needs lang_write_byte.
// Kept here (not shared with emitLinearOutputSupport) because that
// function also unconditionally calls emitEventSupport, which emitProgram
// has already called once for the whole program (D-11-01: calling the
// single-function helper a second time would duplicate every LANG_EVENT
// declaration it contains).
func emitProgramBufferWriter(out *strings.Builder, typeName string) {
	fmt.Fprintf(out, "static int lang_write_buffer_hex(const %s *value) {\n", typeName)
	out.WriteString("  static const char hex[] = \"0123456789abcdef\";\n  size_t index;\n")
	out.WriteString("  if (value->length > sizeof value->bytes) return 0;\n")
	out.WriteString("  for (index = 0u; index < value->length; index++) {\n")
	out.WriteString("    char encoded[2] = {hex[value->bytes[index] >> 4u], hex[value->bytes[index] & 0x0fu]};\n")
	out.WriteString("    if (!lang_write_bytes(encoded, sizeof encoded)) return 0;\n  }\n  return 1;\n}\n\n")
}

func emitProgramByteWriter(out *strings.Builder) {
	out.WriteString("static int lang_write_byte(unsigned char value) {\n")
	out.WriteString("  char encoded[3];\n  int length = snprintf(encoded, sizeof encoded, \"%u\", (unsigned int)value);\n")
	out.WriteString("  return length > 0 && (size_t)length < sizeof encoded && lang_write_bytes(encoded, (size_t)length);\n}\n\n")
}

// emitProgramFunction writes ONE function's own C definition: its own
// fresh, per-function cNames (D-11-08: seeded with the reserved list plus
// every globally allocated name) locals allocator, its straight-line
// operations, and a bare C `return` -- never the whole-TU
// lang.execution/1 JSON tail emitLinear's own single-function path writes,
// since that tail belongs to `main` alone in the multi-function assembler
// (a non-entry callee's own return value is consumed by its caller's own
// OpCall, never written to stdout directly; the entry function's own
// return value is written to stdout exactly once, by `main`, after it
// returns).
func emitProgramFunction(out *strings.Builder, function core.Function, typeName, functionName string, globalNames []string, lookup *emitCallLookup) error {
	places := make(map[string]core.Place, len(function.Linear.Places))
	for _, place := range function.Linear.Places {
		places[place.ID] = place
	}
	parameter, ok := places[function.Parameter.ID]
	if !ok {
		return fmt.Errorf("function %q: linear parameter place is absent", function.ID)
	}

	reserved := make([]string, 0, len(linearFixedNames)+len(globalNames))
	reserved = append(reserved, linearFixedNames...)
	reserved = append(reserved, globalNames...)
	names := newCNames(reserved...)

	locals := make(map[string]string, len(function.Linear.Places))
	locals[parameter.ID] = names.allocate(cLocal(parameter.Name), "place", 0)
	for index, place := range function.Linear.Places {
		if place.ID == parameter.ID {
			continue
		}
		locals[place.ID] = names.allocate(cLocal(place.Name), "place", index)
	}

	fmt.Fprintf(out, "static %s %s(%s %s) {\n", typeName, functionName, typeName, locals[parameter.ID])
	declared := map[string]bool{parameter.ID: true}
	returned := false
	for _, operation := range function.Linear.Operations {
		source, sourceKnown := places[operation.SourceID]
		if !sourceKnown {
			return fmt.Errorf("operation %q has invalid source", operation.ID)
		}
		switch operation.Kind {
		case core.OpCopy, core.OpMove, core.OpBorrowShared, core.OpBorrowExclusive:
			target, exists := places[operation.TargetID]
			if !exists || declared[operation.TargetID] {
				return fmt.Errorf("operation %q has invalid target", operation.ID)
			}
			label := "copy"
			marker := ""
			if operation.Kind == core.OpMove {
				label = "authority transfer"
				marker = " /* lang:mutation-site */"
			} else if operation.Kind == core.OpBorrowShared {
				label = "shared borrow representation"
			} else if operation.Kind == core.OpBorrowExclusive {
				label = "exclusive borrow representation"
			}
			fmt.Fprintf(out, "  %s %s = %s; /* %s: %s */%s\n", typeName, locals[target.ID], locals[source.ID], label, operation.ID, marker)
			fmt.Fprintf(out, "  (void)%s;\n", locals[target.ID])
			eventKind := "value.copied"
			if operation.Kind == core.OpMove {
				eventKind = "value.transferred"
			} else if operation.Kind == core.OpBorrowShared {
				eventKind = "value.borrowed"
			} else if operation.Kind == core.OpBorrowExclusive {
				eventKind = "value.borrowed_exclusive"
			}
			fmt.Fprintf(out, "  if (!lang_record_event(%s, %s, %s, %s, %s, %s)) abort();\n",
				strconv.Quote(eventKind), strconv.Quote(operation.ID+":event"), strconv.Quote(function.ID),
				strconv.Quote(operation.SourceID), strconv.Quote(operation.TargetID), strconv.Quote(operation.TypeID))
			declared[operation.TargetID] = true
		case core.OpCall:
			target, exists := places[operation.TargetID]
			if !exists || declared[operation.TargetID] {
				return fmt.Errorf("operation %q has invalid target", operation.ID)
			}
			calleeName, calleeTypeName, ok := lookup.resolve(operation.CalleeID)
			if !ok {
				return fmt.Errorf("operation %q: call to unresolved callee %q", operation.ID, operation.CalleeID)
			}
			emitCall(out, calleeTypeName, locals[target.ID], calleeName, locals[source.ID], operation)
			declared[operation.TargetID] = true
		case core.OpReturn:
			// Unlike emitLinear's single-function OpReturn arm, this never
			// writes the lang.execution/1 JSON tail: that belongs to
			// `main` alone (emitProgram), written exactly once after the
			// resolved entry function returns. A callee's own event is
			// still recorded here, attributed to ITS OWN function ID
			// (interp.terminalOutcome's identical rule, D-10-32), so the
			// caller's own events and the callee's own function.returned
			// event both land in the shared events buffer in execution
			// order regardless of call depth.
			fmt.Fprintf(out, "  if (!lang_record_event(%s, %s, %s, %s, NULL, %s)) abort(); /* returned place: %s */\n",
				strconv.Quote("function.returned"), strconv.Quote(operation.ID+":event:returned"), strconv.Quote(function.ID),
				strconv.Quote(operation.SourceID), strconv.Quote(operation.TypeID), operation.ID)
			fmt.Fprintf(out, "  return %s;\n", locals[source.ID])
			returned = true
		default:
			return fmt.Errorf("operation %q has unknown kind %q", operation.ID, operation.Kind)
		}
	}
	if !returned {
		return fmt.Errorf("function %q: linear body has no return operation", function.ID)
	}
	out.WriteString("}\n\n")
	return nil
}
