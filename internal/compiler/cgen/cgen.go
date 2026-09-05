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
	function := program.Functions[0]
	if function.Match != nil && function.Linear != nil {
		return emitBranch(program, function)
	}
	if function.Linear != nil {
		if len(function.Linear.Blocks) > 0 {
			return emitLinearForeign(program, function)
		}
		return emitLinear(function)
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
	function := program.Functions[0]
	if function.Match != nil && function.Linear != nil {
		// A branch-shaped function always emits the single-document
		// lang.execution/1 JSON protocol, exactly like emitLinear — there is
		// no separate "portable" text mode for it, matching emitLinear's own
		// unconditional JSON behavior (Emit and EmitNative call it the same
		// way).
		return emitBranch(program, function)
	}
	if function.Linear != nil {
		if len(function.Linear.Blocks) > 0 {
			return emitLinearForeign(program, function)
		}
		return emitLinear(function)
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
	// Phase 4 plan 02 resource ledger (D-04-07)
	"lang_resource_ids", "lang_resource_live", "lang_write_live_resources", "first",
	// Phase 4 plan 05 process-root nonlocal-exit landing pad (D-04-17)
	"lang_nonlocal_landing",
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
		case core.OpCopy, core.OpMove, core.OpBorrowShared, core.OpBorrowExclusive:
			target, exists := places[operation.TargetID]
			if !exists || declared[operation.TargetID] {
				return "", fmt.Errorf("operation %q has invalid target", operation.ID)
			}
			label := "copy"
			marker := ""
			if operation.Kind == core.OpMove {
				label = "authority transfer"
				// The backend causality control (D-02-07) locates its
				// mutation site by this stable generated marker instead of
				// an exact source-derived line, so renaming a fixture
				// binding or reindenting the emitter can no longer turn the
				// control into an opaque operational failure. See
				// session.OwnedBackendMutationRunner.
				marker = " /* lang:mutation-site */"
			} else if operation.Kind == core.OpBorrowShared {
				label = "shared borrow representation"
			} else if operation.Kind == core.OpBorrowExclusive {
				label = "exclusive borrow representation"
			}
			fmt.Fprintf(&out, "  %s %s = %s; /* %s: %s */%s\n", typeName, locals[target.ID], locals[source.ID], label, operation.ID, marker)
			fmt.Fprintf(&out, "  (void)%s;\n", locals[target.ID])
			eventKind := "value.copied"
			if operation.Kind == core.OpMove {
				eventKind = "value.transferred"
			} else if operation.Kind == core.OpBorrowShared {
				eventKind = "value.borrowed"
			} else if operation.Kind == core.OpBorrowExclusive {
				eventKind = "value.borrowed_exclusive"
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

// emitLinearForeign is the native lowering for a straight-line (Match-less)
// function whose linear body forks on a fallible foreign call (D-04-04):
// the shape checkFallibleLinear produces, distinguished from an ordinary
// Phase 1/2/3 linear body by len(function.Linear.Blocks) > 0. This plan's
// scope is narrow -- an entry block of exactly one OpForeignCall, an ok
// block ending in OpReturn, and an err block ending in OpFail -- so this
// emitter is intentionally not a general CFG-to-C translator; it errors
// rather than silently mishandling any richer shape a later plan may add.
//
// Per D-04-10, the generated C here never opens the frozen foreign
// translation unit's private header: it declares the extern symbol and its
// result-record shape itself, from the compiler's own contract. Only the
// conformance TU a later plan generates is permitted to include that
// header.
func emitLinearForeign(program core.Program, function core.Function) (string, error) {
	if function.ForeignContract == nil || function.ForeignContract.Symbol == "" {
		return "", fmt.Errorf("foreign-shaped function %q has no foreign contract", function.ID)
	}
	// Independent refusal (04-VERIFICATION.md gap 2, FFI-01/D-04-12):
	// Emit/EmitNative already call corevalidate.Validate first, but this
	// guard does not rely on that having happened -- see validForeignSymbol's
	// doc comment for why a second implementation is deliberate here.
	if !validForeignSymbol(function.ForeignContract.Symbol) {
		return "", fmt.Errorf("foreign symbol %q is not a C identifier", function.ForeignContract.Symbol)
	}
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
	blocksByID := make(map[string]core.Block, len(function.Linear.Blocks))
	for _, block := range function.Linear.Blocks {
		blocksByID[block.ID] = block
	}
	edgesByID := make(map[string]core.Edge, len(function.Linear.Edges))
	for _, edge := range function.Linear.Edges {
		edgesByID[edge.ID] = edge
	}
	operationsByID := make(map[string]core.LinearOperation, len(function.Linear.Operations))
	for _, operation := range function.Linear.Operations {
		operationsByID[operation.ID] = operation
	}

	names := newCNames(linearFixedNames...)
	locals := make(map[string]string, len(function.Linear.Places))
	for index, place := range function.Linear.Places {
		locals[place.ID] = names.allocate(cLocal(place.Name), "place", index)
	}
	symbolC := foreignExternName(function.ForeignContract.Symbol)
	resultType := symbolC + "_result"

	// ledger is Phase 4 plan 02's runtime live-resource accounting
	// (D-04-07): populated only when the function actually declares
	// OpRelease operations (the resource-lifecycle shape), so the 04-01
	// tracer shape -- no OpRelease anywhere -- keeps emitting the literal
	// "live_resources":[] it always has, byte-for-byte.
	ledger := newResourceLedger(function)

	var out strings.Builder
	out.WriteString("/* generated by Codename Lang; schema lang.c17/0 */\n")
	// setjmp.h and stdlib.h back the process-root landing pad (D-04-17):
	// setjmp/jmp_buf install and re-enter the single landing point, abort
	// terminates the pad itself (D-04-18: no release, no unwinding, no
	// containment -- see the pad body below).
	out.WriteString("#include <setjmp.h>\n#include <stddef.h>\n#include <stdio.h>\n#include <stdlib.h>\n#include <string.h>\n\n")
	// This typedef/extern pair is the compiler's own contract, formed from
	// the `foreign C {}` declaration alone (D-04-10) -- it never includes
	// the frozen foreign translation unit's private header.
	fmt.Fprintf(&out, "typedef struct %s {\n  unsigned char ok;\n  unsigned char value;\n} %s;\n\n", resultType, resultType)
	fmt.Fprintf(&out, "extern %s %s(unsigned char argument);\n\n", resultType, symbolC)
	// D-04-17: lang_nonlocal_landing is the ONE process-root landing point --
	// one per process, never per call or per borrow. It is declared here with
	// external linkage (no "static") so a foreign translation unit performing
	// a genuine nonlocal exit (native/lang_foreign_nonlocal.c) can longjmp
	// into it via its own `extern jmp_buf` declaration; nothing else in this
	// generated file, or any other emitter, ever declares this symbol, so
	// exactly one exists per linked program.
	out.WriteString("jmp_buf lang_nonlocal_landing;\n\n")
	// D-04-20: every function this emitter handles declares a foreign
	// acquisition (checked at this function's own entry above), so it always
	// takes the STREAMING event path -- the one selection site for D-04-20's
	// additive emitter, never the buffered emitEventSupport plain emitLinear
	// still uses byte-for-byte (D-04-23).
	emitLinearForeignOutputSupport(&out, function, typeName)
	ledger.emitDeclarations(&out)

	out.WriteString("int main(int argc, char **argv) {\n")
	out.WriteString("  if (argc != 2) return 64;\n")
	fmt.Fprintf(&out, "  if (strcmp(argv[1], %s) != 0) return 65;\n", strconv.Quote(input))
	fmt.Fprintf(&out, "  %s %s = %s;\n", typeName, locals[parameter.ID], initializer)
	// The events array opens FIRST, before any operation executes, so a
	// streamed lang_record_event call always lands inside a syntactically
	// valid (if not yet terminated) JSON array -- every terminal writer below
	// closes it with "]" plus the outcome/live_resources tail, never the
	// other order.
	out.WriteString("  if (!lang_write_literal(\"{\\\"schema\\\":\\\"lang.execution/1\\\",\\\"events\\\":[\")) return 74;\n")
	emitNonlocalPad(&out, function, ledger)

	currentBlockID := function.ID + ":block:entry"
	visited := make(map[string]bool, len(function.Linear.Blocks))
	step := 0
	for {
		if visited[currentBlockID] {
			return "", fmt.Errorf("foreign call chain revisits block %q", currentBlockID)
		}
		visited[currentBlockID] = true
		block, known := blocksByID[currentBlockID]
		if !known || len(block.OperationIDs) == 0 {
			return "", fmt.Errorf("block %q has an unsupported shape this phase", currentBlockID)
		}
		lastOp, known := operationsByID[block.OperationIDs[len(block.OperationIDs)-1]]
		if !known {
			return "", fmt.Errorf("block %q references an unknown operation", currentBlockID)
		}
		if lastOp.Kind != core.OpForeignCall {
			// Terminal block: zero or more OpRelease, then OpReturn. An
			// OpFail-terminated block is emitted inline at its own fork
			// point below and is never visited by this outer walk.
			if err := emitForeignReleasesAndReturn(&out, function, places, locals, operationsByID, block.OperationIDs, typeName, ledger); err != nil {
				return "", err
			}
			out.WriteString("  return 0;\n}\n")
			return out.String(), nil
		}
		if len(block.OperationIDs) != 1 {
			return "", fmt.Errorf("block %q has an unsupported shape this phase", currentBlockID)
		}
		okEdge, known := edgesByID[lastOp.OkEdgeID]
		if !known {
			return "", fmt.Errorf("foreign call %q references an unknown ok edge", lastOp.ID)
		}
		errEdge, known := edgesByID[lastOp.ErrEdgeID]
		if !known {
			return "", fmt.Errorf("foreign call %q references an unknown err edge", lastOp.ID)
		}
		resultLocal := names.allocate(cLocal(fmt.Sprintf("foreign_result_%d", step)), "foreign_result", step)
		step++
		fmt.Fprintf(&out, "  %s %s = %s(%s); /* %s */\n", resultType, resultLocal, symbolC, locals[lastOp.SourceID], lastOp.ID)
		fmt.Fprintf(&out, "  if (!lang_record_event(%s, %s, %s, %s, %s, %s)) return 74;\n",
			strconv.Quote("foreign.called"), strconv.Quote(lastOp.ID+":event"), strconv.Quote(function.ID),
			strconv.Quote(lastOp.SourceID), strconv.Quote(lastOp.TargetID), strconv.Quote(lastOp.TypeID))
		fmt.Fprintf(&out, "  if (!%s.ok) {\n", resultLocal)
		if errEdge.ToBlockID != okEdge.ToBlockID {
			errBlock, known := blocksByID[errEdge.ToBlockID]
			if !known {
				return "", fmt.Errorf("foreign call %q references an unknown err block", lastOp.ID)
			}
			if err := emitForeignReleasesAndFail(&out, program, function, operationsByID, errBlock.OperationIDs, "    ", ledger); err != nil {
				return "", err
			}
		}
		out.WriteString("  }\n")
		if ledger.tracks(lastOp.ID) {
			fmt.Fprintf(&out, "  %s\n", ledger.markLive(lastOp.ID))
		}
		fmt.Fprintf(&out, "  %s %s = %s.value;\n", typeName, locals[lastOp.TargetID], resultLocal)
		fmt.Fprintf(&out, "  (void)%s;\n", locals[lastOp.TargetID])
		currentBlockID = okEdge.ToBlockID
	}
}

// padInstallMarker/padEndMarker bracket the ENTIRE process-root landing pad
// block (D-04-17), not just its installation line: control:foreign.
// nonlocal_exit_undetected's first mutation-kill demonstration deletes every
// line from padInstallMarker through padEndMarker inclusive, removing the
// setjmp() call AND its whole pad body in one span so the remaining C stays
// syntactically valid (an unmatched brace from a line-only deletion would
// not compile). A program with the pad removed still declares
// lang_nonlocal_landing (now unused) and still links against a foreign
// symbol that may longjmp into it -- an uninitialized jmp_buf -- which is
// exactly the "detection now silently absent" state this control exists to
// catch.
const padInstallMarker = "/* lang:nonlocal-pad-site */"
const padEndMarker = "/* lang:nonlocal-pad-end */"

// nonlocalExitDefectReason is the exact "output" string the process-root
// pad's own function.defected event carries. It is duplicated verbatim (not
// imported) in interp.go's runLinearBlocks, which models the identical
// probe convention -- the two literal strings are kept textually identical
// by convention and comment, exactly as native/lang_foreign_nonlocal.c's
// "second call" rule is a shared, documented convention rather than a real
// cross-engine call.
const nonlocalExitDefectReason = "foreign nonlocal exit detected at process-root landing pad"

// emitNonlocalPad writes D-04-17's ONE process-root landing pad: a single
// setjmp() call installed immediately after the events array opens (so a
// foreign nonlocal exit reached before ANY operation below ever executes is
// still inside a syntactically valid, if not yet terminated, JSON events
// array), guarding a body that, per D-04-18, runs NO release and touches NO
// automatic storage changed since the setjmp call -- only the static-storage
// ledger the resourceLedger type already declares. Called exactly once per
// emitLinearForeign invocation, so "exactly one pad" is a structural
// property of this call site, not a runtime count.
//
// REACHABILITY (D-04-21/D-10 -- accepted residual limitations, T-04-33,
// named in uncheckedForeignObligations' sidecar list too, never claimed as
// covered): a landing point established BELOW this pad by a foreign-invoked callback
// that itself calls setjmp deeper in the call stack is invisible
// to this single process-root pad, as is a foreign call that terminates the
// process directly (exit()/_exit()) rather than performing a nonlocal exit
// back into Lang-controlled code. Neither is claimed detected anywhere in
// this project.
func emitNonlocalPad(out *strings.Builder, function core.Function, ledger *resourceLedger) {
	parameterTypeID := ""
	for _, place := range function.Linear.Places {
		if place.ID == function.Parameter.ID {
			parameterTypeID = place.TypeID
			break
		}
	}
	fmt.Fprintf(out, "  if (setjmp(lang_nonlocal_landing) != 0) { %s\n", padInstallMarker)
	// C17 7.13.2.1p3: every object of automatic storage duration that is
	// local to the function containing the setjmp invocation, that does not
	// have volatile-qualified type, and that is modified between the setjmp
	// invocation and a later longjmp call has an indeterminate value after
	// that longjmp. The pad below therefore reads ONLY the static-storage
	// resource ledger (never an automatic local) and calls NO release
	// function: releasing from an indeterminate handle would convert an
	// honest leak into a use-after-free (D-04-18). This law is stated here,
	// at the point it governs, not only in CONTEXT.md.
	fmt.Fprintf(out, "    if (!lang_record_event(%s, %s, %s, NULL, NULL, NULL)) abort();\n",
		strconv.Quote("foreign.nonlocal_exit"), strconv.Quote(function.ID+":event:nonlocal_exit"), strconv.Quote(function.ID))
	ledger.emitLeakEvents(out, function.ID)
	emitManualDefectEvent(out, function.ID+":event:nonlocal_defect", function.ID, function.Parameter.ID, parameterTypeID, nonlocalExitDefectReason)
	out.WriteString("    if (!lang_write_literal(\"],\\\"outcome\\\":{\\\"kind\\\":\\\"defect\\\",\\\"value\\\":\\\"\\\"},\\\"live_resources\\\":[\")) abort();\n")
	if len(ledger.ids) == 0 {
		out.WriteString("    if (!lang_write_literal(\"]}\\n\")) abort();\n")
	} else {
		out.WriteString("    if (!lang_write_live_resources()) abort();\n")
		out.WriteString("    if (!lang_write_literal(\"]}\\n\")) abort();\n")
	}
	// D-04-15/D-04-18: no catch, no containment, no unwinding, no cleanup --
	// abort-only at the process root, exactly like the generated
	// _Noreturn lang_defect function every ordinary defect path calls.
	out.WriteString("    abort();\n")
	fmt.Fprintf(out, "  } %s\n", padEndMarker)
}

// emitManualDefectEvent writes one function.defected event object directly
// with lang_write_json_string/lang_write_literal calls, exactly like
// emitBranchOperations' OpDefect case: the shared LANG_EVENT struct
// lang_record_event uses has no "output" field for the required reason
// string, and adding one there would move every other emitter's generated
// C. Assumes at least one event was already recorded via lang_record_event
// on this path (true here: emitNonlocalPad always records the
// foreign.nonlocal_exit event first), so the leading "," separator is
// unconditional rather than guarded by lang_event_count.
func emitManualDefectEvent(out *strings.Builder, id, functionID, sourcePlace, typeID, reason string) {
	out.WriteString("    if (!lang_write_bytes(\",\", 1u)) abort();\n")
	out.WriteString("    if (!lang_write_literal(\"{\\\"schema\\\":\\\"lang.execution/1\\\",\\\"id\\\":\")) abort();\n")
	fmt.Fprintf(out, "    if (!lang_write_json_string(%s)) abort();\n", strconv.Quote(id))
	out.WriteString("    if (!lang_write_literal(\",\\\"kind\\\":\\\"function.defected\\\",\\\"function_id\\\":\")) abort();\n")
	fmt.Fprintf(out, "    if (!lang_write_json_string(%s)) abort();\n", strconv.Quote(functionID))
	out.WriteString("    if (!lang_write_literal(\",\\\"source_place\\\":\")) abort();\n")
	fmt.Fprintf(out, "    if (!lang_write_json_string(%s)) abort();\n", strconv.Quote(sourcePlace))
	out.WriteString("    if (!lang_write_literal(\",\\\"type_id\\\":\")) abort();\n")
	fmt.Fprintf(out, "    if (!lang_write_json_string(%s)) abort();\n", strconv.Quote(typeID))
	out.WriteString("    if (!lang_write_literal(\",\\\"output\\\":\")) abort();\n")
	fmt.Fprintf(out, "    if (!lang_write_json_string(%s)) abort();\n", strconv.Quote(reason))
	out.WriteString("    if (!lang_write_literal(\"}\")) abort();\n")
}

// resourceLedger is the generated C's own runtime live-resource accounting
// (D-04-07/spike-005 ledger pattern): a compile-time-sized static array of
// flags, one per tracked acquisition (an OpForeignCall whose ok/err edges
// diverge AND is discharged by some OpRelease somewhere in the function),
// set live right after a successful acquisition and cleared by the matching
// release. It exists only so the release-omission mutation (deleting one
// generated release line) is OBSERVABLE at the native layer: without runtime
// tracking, a compile-time-only live_resources computation could never
// reflect a line a mutation deleted.
type resourceLedger struct {
	slot map[string]int
	ids  []string // parallel to slot: ids[slot[opID]] is opID's own place-based resource identifier
}

func newResourceLedger(function core.Function) *resourceLedger {
	ledger := &resourceLedger{slot: make(map[string]int)}
	placeByOp := make(map[string]string, len(function.Linear.Operations))
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpForeignCall {
			placeByOp[operation.ID] = operation.TargetID
		}
	}
	for _, operation := range function.Linear.Operations {
		if operation.Kind != core.OpRelease || operation.ReleasesOperationID == "" {
			continue
		}
		if _, already := ledger.slot[operation.ReleasesOperationID]; already {
			continue
		}
		ledger.slot[operation.ReleasesOperationID] = len(ledger.ids)
		ledger.ids = append(ledger.ids, placeByOp[operation.ReleasesOperationID])
	}
	return ledger
}

func (l *resourceLedger) tracks(opID string) bool {
	_, ok := l.slot[opID]
	return ok
}

func (l *resourceLedger) emitDeclarations(out *strings.Builder) {
	if len(l.ids) == 0 {
		return
	}
	fmt.Fprintf(out, "static const char *lang_resource_ids[%d] = {\n", len(l.ids))
	for _, id := range l.ids {
		fmt.Fprintf(out, "  %s,\n", strconv.Quote(id))
	}
	out.WriteString("};\n")
	fmt.Fprintf(out, "static int lang_resource_live[%d];\n\n", len(l.ids))
	out.WriteString("static int lang_write_live_resources(void) {\n")
	out.WriteString("  size_t index; int first = 1;\n")
	fmt.Fprintf(out, "  for (index = 0u; index < %du; index++) {\n", len(l.ids))
	out.WriteString("    if (!lang_resource_live[index]) continue;\n")
	out.WriteString("    if (!first && !lang_write_bytes(\",\", 1u)) return 0;\n")
	out.WriteString("    first = 0;\n")
	out.WriteString("    if (!lang_write_json_string(lang_resource_ids[index])) return 0;\n")
	out.WriteString("  }\n  return 1;\n}\n\n")
}

// ledgerPopulateMarker is the mutation-kill seam control:foreign.
// nonlocal_exit_undetected's second demonstration locates (D-04-21/D-10):
// deleting the FIRST line bearing this marker drops one acquisition's own
// live-tracking population, so the landing pad's own leak count understates
// the true live set -- a different mutation than deleting the pad
// installation itself (padInstallMarker/padEndMarker below), which instead
// makes the pad never run at all.
const ledgerPopulateMarker = "/* lang:ledger-populate-site */"

func (l *resourceLedger) markLive(opID string) string {
	return fmt.Sprintf("lang_resource_live[%d] = 1; %s", l.slot[opID], ledgerPopulateMarker)
}

// emitLeakEvents writes one lang_record_event("resource.leaked", ...) call
// per ledger-tracked acquisition, each conditioned on that acquisition's own
// live flag at the moment the process-root pad runs (D-04-17): unrolled in
// Go at emission time (not a C loop) because the tracked set is small and
// fixed per function, and because each event needs its own unique,
// human-legible ID -- exactly one call per slot, in ledger (first-acquired)
// order, so the reported leak sequence is deterministic rather than a
// function of runtime iteration.
func (l *resourceLedger) emitLeakEvents(out *strings.Builder, functionID string) {
	for index, id := range l.ids {
		fmt.Fprintf(out, "    if (lang_resource_live[%d] && !lang_record_event(%s, %s, %s, %s, NULL, NULL)) abort();\n",
			index, strconv.Quote("resource.leaked"), strconv.Quote(fmt.Sprintf("%s:event:leaked:%d", functionID, index)),
			strconv.Quote(functionID), strconv.Quote(id))
	}
}

// releaseStatement returns the release line for a discharged acquisition:
// the ordinary event-recording call plus, on the SAME generated line, the
// ledger decrement -- so the omission mutation (deleting one line bearing
// the marker below) removes both the event and the decrement together,
// leaving the resource observably live at termination. lang:release-site is
// a marker distinct from lang:mutation-site (D-02-07's owned-transfer
// control), because the two mutation runners must locate different seams.
func (l *resourceLedger) releaseStatement(function core.Function, operation core.LinearOperation) string {
	ledgerClear := ""
	if slot, ok := l.slot[operation.ReleasesOperationID]; ok {
		ledgerClear = fmt.Sprintf(" lang_resource_live[%d] = 0;", slot)
	}
	return fmt.Sprintf("if (!lang_record_event(%s, %s, %s, %s, NULL, %s)) return 74;%s /* lang:release-site */",
		strconv.Quote("resource.released"), strconv.Quote(operation.ID+":event"), strconv.Quote(function.ID),
		strconv.Quote(operation.SourceID), strconv.Quote(operation.TypeID), ledgerClear)
}

// emitForeignReleasesAndReturn writes a terminal (success) block: zero or
// more OpRelease operations, then the single OpReturn.
func emitForeignReleasesAndReturn(out *strings.Builder, function core.Function, places map[string]core.Place, locals map[string]string, operationsByID map[string]core.LinearOperation, operationIDs []string, typeName string, ledger *resourceLedger) error {
	for _, operationID := range operationIDs {
		operation, known := operationsByID[operationID]
		if !known {
			return fmt.Errorf("block references unknown operation %q", operationID)
		}
		switch operation.Kind {
		case core.OpRelease:
			fmt.Fprintf(out, "  %s\n", ledger.releaseStatement(function, operation))
		case core.OpReturn:
			source, known := places[operation.SourceID]
			if !known {
				return fmt.Errorf("operation %q has invalid source", operation.ID)
			}
			fmt.Fprintf(out, "  if (!lang_record_event(%s, %s, %s, %s, NULL, %s)) return 74; /* returned place: %s */\n",
				strconv.Quote("function.returned"), strconv.Quote(operation.ID+":event:returned"), strconv.Quote(function.ID),
				strconv.Quote(operation.SourceID), strconv.Quote(operation.TypeID), operation.ID)
			// D-04-20: the events array was opened at the top of main(); this
			// terminal writer closes it, then writes outcome/live_resources --
			// the terminal record is the LAST write on this path.
			out.WriteString("  if (!lang_write_literal(\"],\\\"outcome\\\":{\\\"kind\\\":\\\"returned\\\",\\\"value\\\":\\\"\")) return 74;\n")
			fmt.Fprintf(out, "  if (!lang_write_byte(%s)) return 74;\n", locals[source.ID])
			if len(ledger.ids) == 0 {
				out.WriteString("  if (!lang_write_literal(\"\\\"},\\\"live_resources\\\":[]}\\n\")) return 74;\n")
			} else {
				out.WriteString("  if (!lang_write_literal(\"\\\"},\\\"live_resources\\\":[\")) return 74;\n")
				out.WriteString("  if (!lang_write_live_resources()) return 74;\n")
				out.WriteString("  if (!lang_write_literal(\"]}\\n\")) return 74;\n")
			}
		default:
			return fmt.Errorf("block operation %q has unsupported kind %q this phase", operation.ID, operation.Kind)
		}
	}
	return nil
}

// emitForeignReleasesAndFail writes an err block: zero or more OpRelease
// operations, then the single OpFail terminator. errorLiteral is the
// compile-time-known value foreignFailureLiteral resolves (see its doc
// comment for why one literal is honest this phase).
func emitForeignReleasesAndFail(out *strings.Builder, program core.Program, function core.Function, operationsByID map[string]core.LinearOperation, operationIDs []string, indent string, ledger *resourceLedger) error {
	for _, operationID := range operationIDs {
		operation, known := operationsByID[operationID]
		if !known {
			return fmt.Errorf("err block references unknown operation %q", operationID)
		}
		switch operation.Kind {
		case core.OpRelease:
			fmt.Fprintf(out, "%s%s\n", indent, ledger.releaseStatement(function, operation))
		case core.OpFail:
			errorLiteral, err := foreignFailureLiteral(program, function)
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "%sif (!lang_record_event(%s, %s, %s, %s, NULL, %s)) return 74;\n",
				indent, strconv.Quote("function.failed"), strconv.Quote(operation.ID+":event:failed"), strconv.Quote(function.ID),
				strconv.Quote(operation.SourceID), strconv.Quote(operation.TypeID))
			// D-04-20: close the events array opened at the top of main(),
			// then write outcome/live_resources -- the terminal record is the
			// LAST write on this path too.
			fmt.Fprintf(out, "%sif (!lang_write_literal(\"],\\\"outcome\\\":{\\\"kind\\\":\\\"typed_failure\\\",\\\"value\\\":\")) return 74;\n", indent)
			fmt.Fprintf(out, "%sif (!lang_write_json_string(%s)) return 74;\n", indent, strconv.Quote(errorLiteral))
			if len(ledger.ids) == 0 {
				fmt.Fprintf(out, "%sif (!lang_write_literal(\"},\\\"live_resources\\\":[]}\\n\")) return 74;\n", indent)
			} else {
				fmt.Fprintf(out, "%sif (!lang_write_literal(\"},\\\"live_resources\\\":[\")) return 74;\n", indent)
				fmt.Fprintf(out, "%sif (!lang_write_live_resources()) return 74;\n", indent)
				fmt.Fprintf(out, "%sif (!lang_write_literal(\"]}\\n\")) return 74;\n", indent)
			}
			fmt.Fprintf(out, "%sreturn 0;\n", indent)
		default:
			return fmt.Errorf("err block operation %q has unsupported kind %q this phase", operation.ID, operation.Kind)
		}
	}
	return nil
}

// foreignFailureLiteral names the JSON string value an OpFail terminal
// record carries. Per D-04-05 the err edge's payload is a place of an
// ordinary declared nullary ADT, but this phase has no case-analysis syntax
// to pick a specific alternative at the failure site -- both engines
// deterministically agree on the ADT's FIRST declared alternative,
// documented here as a known narrowing (a real per-cause error value is
// future work), rather than silently picking an unreachable, engine-
// specific value.
func foreignFailureLiteral(program core.Program, function core.Function) (string, error) {
	for _, dataType := range program.DataTypes {
		if dataType.Name == function.ForeignContract.Fails {
			if len(dataType.Alternatives) == 0 {
				return "", fmt.Errorf("foreign failure type %q has no alternatives", dataType.Name)
			}
			return dataType.Alternatives[0], nil
		}
	}
	return "", fmt.Errorf("foreign failure type %q is not declared", function.ForeignContract.Fails)
}

// validForeignSymbol is a DELIBERATE second implementation of the exact
// predicate corevalidate.validCIdentifier applies (^[A-Za-z_][A-Za-z0-9_]*$,
// checked byte-by-byte). It is not an accidental duplicate: corevalidate's
// validCIdentifier is unexported, and cgen must be able to refuse a hostile
// Symbol without depending on corevalidate.Validate having run first -- the
// same independence posture corevalidate itself takes toward check.go
// throughout this phase. This guard protects the four sites that splice a
// Symbol raw into generated C: the extern declaration built around
// symbolC/resultType (below, ~line 352-373), the call-expression callee that
// reuses the same symbolC (~line 441), foreignExternName itself (below), and
// the generated header's `/* symbol: ... */` comment (EmitForeignHeader).
func validForeignSymbol(symbol string) bool {
	if symbol == "" {
		return false
	}
	first := symbol[0]
	if !(first >= 'A' && first <= 'Z' || first >= 'a' && first <= 'z' || first == '_') {
		return false
	}
	for index := 1; index < len(symbol); index++ {
		b := symbol[index]
		if !(b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z' || b >= '0' && b <= '9' || b == '_') {
			return false
		}
	}
	return true
}

// foreignExternName derives the linkage name a generated foreign call site
// declares `extern`. This deliberately does NOT go through cName/cLocal's
// ordinary-identifier allocator (cgen.go's namespace-closure invariant,
// TestGeneratedIdentifierNamespacesStayConfined) because it must match a
// REAL exported symbol in the byte-frozen foreign translation unit
// (native/lang_foreign_resource.c) verbatim, by convention, not by
// collision-avoidance allocation.
func foreignExternName(symbol string) string { return "_LANG_" + symbol }

func emitLinearOutputSupport(out *strings.Builder, function core.Function, typeName string) {
	emitEventSupport(out, len(function.Linear.Operations))
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

// emitEventSupport writes the LANG_EVENT macros, struct, bounded output
// writer, JSON-string escaper, and event recorder shared by every linear-
// shaped emitter (emitLinear and, from Phase 3, emitBranch). It deliberately
// excludes the scalar value writer (lang_write_buffer_hex / lang_write_byte)
// because a branch-shaped function never needs one: its returned value is a
// compile-time-known alternative name per case, written as a JSON string
// literal, not a dynamically-encoded scalar (see emitBranchOperations).
// emitLinearForeignOutputSupport is D-04-20's selection site: every function
// emitLinearForeign handles carries a core.ForeignContract (checked at that
// function's own entry), so it always takes the STREAMING event path
// (emitStreamingEventSupport) instead of the buffered one
// emitLinearOutputSupport/emitEventSupport give every other linear emitter.
// The buffered lang_write_events() writes nothing until the function's own
// terminal statement, so an aborting foreign-acquiring process would emit
// zero events -- exactly the gap that would make SC4's interpreter/native
// agreement unfalsifiable on the paths SC3 is about. This is a NEW sibling
// of emitLinearOutputSupport, never called from emitLinear/emitBranch, so
// every existing committed generated-C golden (Phase 1/2/3, and the frozen
// foreign layout fixture) is untouched (D-04-23).
func emitLinearForeignOutputSupport(out *strings.Builder, function core.Function, typeName string) {
	emitStreamingEventSupport(out)
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

// emitStreamingEventSupport is D-04-20's additive streaming event emitter:
// unlike emitEventSupport's lang_record_event (which appends to a fixed-size
// array replayed once by lang_write_events at the very end), this variant's
// lang_record_event writes the event's own JSON object to stdout the MOMENT
// it is called, so every event recorded before an abort()/nonlocal exit
// partway through a function has already reached the file descriptor. The
// caller writes the "events":[ opening literal FIRST, before any operation
// executes, and every terminal writer closes it with "]" plus the
// outcome/live_resources tail -- never the other order -- which is why this
// cannot simply replace emitEventSupport's lang_write_events in place.
// emitEventSupport itself is completely unmodified by this function's
// existence (D-04-23): every Phase 1/2/3 program, and every Phase 4 program
// with no foreign acquisition, still uses it byte-for-byte.
func emitStreamingEventSupport(out *strings.Builder) {
	out.WriteString("#define LANG_OUTPUT_LIMIT 65536u\n\n")
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
	out.WriteString("  if (lang_event_count != 0u && !lang_write_bytes(\",\", 1u)) return 0;\n")
	out.WriteString("  lang_event_count++;\n")
	out.WriteString("  if (!lang_write_literal(\"{\\\"schema\\\":\\\"lang.execution/1\\\",\\\"id\\\":\") || !lang_write_json_string(id)) return 0;\n")
	out.WriteString("  if (!lang_write_literal(\",\\\"kind\\\":\") || !lang_write_json_string(kind)) return 0;\n")
	out.WriteString("  if (!lang_write_literal(\",\\\"function_id\\\":\") || !lang_write_json_string(function_id)) return 0;\n")
	out.WriteString("  if (source_place != NULL && (!lang_write_literal(\",\\\"source_place\\\":\") || !lang_write_json_string(source_place))) return 0;\n")
	out.WriteString("  if (target_place != NULL && (!lang_write_literal(\",\\\"target_place\\\":\") || !lang_write_json_string(target_place))) return 0;\n")
	out.WriteString("  if (type_id != NULL && (!lang_write_literal(\",\\\"type_id\\\":\") || !lang_write_json_string(type_id))) return 0;\n")
	out.WriteString("  return lang_write_bytes(\"}\", 1u);\n}\n\n")
}

func emitEventSupport(out *strings.Builder, capacity int) {
	fmt.Fprintf(out, "#define LANG_OUTPUT_LIMIT 65536u\n#define LANG_EVENT_CAPACITY %du\n\n", capacity)
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
}

// functionHasDefect reports whether any operation in function's Linear body
// is an OpDefect -- the selection gate deciding whether emitBranch also
// emits the generated lang_defect support function (D-04-15). Every
// defect-free program's generated C is completely unaffected.
func functionHasDefect(function core.Function) bool {
	if function.Linear == nil {
		return false
	}
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpDefect {
			return true
		}
	}
	return false
}

// emitDefectSupport writes the generated _Noreturn lang_defect function
// (D-04-14/D-04-15): the one generated-code exemption from the zero-
// attribute control (D-04-13), because _Noreturn here is a property of a
// function cgen itself emits and every path of which provably ends in
// abort() -- not an unproven claim about a foreign callee. Called only from
// an emitter whose function actually contains an OpDefect operation, so
// every defect-free program's generated C is byte-for-byte unaffected
// (D-04-23).
func emitDefectSupport(out *strings.Builder) {
	out.WriteString("#include <stdlib.h>\n\n")
	out.WriteString("_Noreturn static void lang_defect(const char *reason) {\n")
	out.WriteString("  (void)reason; /* D-04-15: no catch, no containment, no unwinding, no cleanup -- abort-only at the process root */\n")
	out.WriteString("  abort();\n")
	out.WriteString("}\n\n")
}

// emitBranch is the native lowering for a match function whose arms carry
// linear bodies (D-12a's fourth OperationKind consumer, alongside check,
// corevalidate, and interp). The scrutinee is a nominal alternative type
// with no payload, so no operation in an arm's body can ever produce a
// DIFFERENT alternative than the one that selected it — copy/move/borrow all
// preserve the runtime value. The returned value for any given arm is
// therefore always exactly that arm's matched pattern, a compile-time-known
// literal, written directly as a JSON string rather than encoded dynamically
// (contrast emitLinear's Byte/Buffer scalar writer, which emitBranch has no
// analog of and does not need).
func emitBranch(program core.Program, function core.Function) (string, error) {
	if len(program.DataTypes) != 1 {
		return "", fmt.Errorf("branch C emitter expects one data type")
	}
	dataType := program.DataTypes[0]
	if len(dataType.Alternatives) == 0 {
		return "", fmt.Errorf("branch C emitter requires alternatives")
	}
	if function.Match == nil || function.Linear == nil {
		return "", fmt.Errorf("branch C emitter expects a match function carrying linear blocks")
	}

	places := make(map[string]core.Place, len(function.Linear.Places))
	for _, place := range function.Linear.Places {
		places[place.ID] = place
	}
	operationsByID := make(map[string]core.LinearOperation, len(function.Linear.Operations))
	for _, operation := range function.Linear.Operations {
		operationsByID[operation.ID] = operation
	}
	blocksByID := make(map[string]core.Block, len(function.Linear.Blocks))
	for _, block := range function.Linear.Blocks {
		blocksByID[block.ID] = block
	}

	names := newCNames(linearFixedNames...)
	typeName := names.allocate(cName(dataType.Name), "type", 0)
	alternativeNames := make([]string, len(dataType.Alternatives))
	alternativeBySource := make(map[string]string, len(dataType.Alternatives))
	for index, alternative := range dataType.Alternatives {
		alternativeNames[index] = names.allocate(typeName+"_"+cName(alternative), "alternative", index)
		alternativeBySource[alternative] = alternativeNames[index]
	}

	locals := make(map[string]string, len(function.Linear.Places))
	for index, place := range function.Linear.Places {
		locals[place.ID] = names.allocate(cLocal(place.Name), "place", index)
	}
	parameter, ok := places[function.Parameter.ID]
	if !ok {
		return "", fmt.Errorf("branch parameter place is absent")
	}

	var out strings.Builder
	out.WriteString("/* generated by Codename Lang; schema lang.c17/0 */\n")
	out.WriteString("/* Moves below are authority transitions; C value assignment makes no ABI or zero-copy claim. */\n")
	out.WriteString("#include <stddef.h>\n#include <stdio.h>\n#include <string.h>\n\n")
	fmt.Fprintf(&out, "typedef enum %s {\n", typeName)
	for index := range dataType.Alternatives {
		fmt.Fprintf(&out, "  %s = %d,\n", alternativeNames[index], index)
	}
	fmt.Fprintf(&out, "} %s;\n\n", typeName)
	emitEventSupport(&out, len(function.Linear.Operations))
	if functionHasDefect(function) {
		emitDefectSupport(&out)
	}

	out.WriteString("int main(int argc, char **argv) {\n")
	out.WriteString("  if (argc != 2) return 64;\n")
	fmt.Fprintf(&out, "  %s %s;\n", typeName, locals[parameter.ID])
	for index, alternative := range dataType.Alternatives {
		prefix := "if"
		if index > 0 {
			prefix = "else if"
		}
		fmt.Fprintf(&out, "  %s (strcmp(argv[1], %s) == 0) %s = %s;\n", prefix, strconv.Quote(alternative), locals[parameter.ID], alternativeNames[index])
	}
	out.WriteString("  else return 65;\n")
	fmt.Fprintf(&out, "  switch (%s) {\n", locals[parameter.ID])
	for _, arm := range function.Match.Arms {
		block, known := blocksByID[arm.BlockID]
		if !known {
			return "", fmt.Errorf("arm %q references unknown block %q", arm.ID, arm.BlockID)
		}
		fmt.Fprintf(&out, "    case %s: {\n", alternativeBySource[arm.Pattern])
		if err := emitBranchOperations(&out, function, places, locals, operationsByID, block.OperationIDs, typeName, arm.Pattern); err != nil {
			return "", err
		}
		out.WriteString("    }\n")
	}
	out.WriteString("  }\n  return 70; /* invalid safe-language tag: fail in main */\n}\n")
	return out.String(), nil
}

// emitBranchOperations writes one arm block's straight-line C, in core
// order, ending with the lang.execution/1 JSON document for that arm's
// return. returnLiteral is the compile-time-known alternative name this
// block always returns (see emitBranch's doc comment).
func emitBranchOperations(out *strings.Builder, function core.Function, places map[string]core.Place, locals map[string]string, operationsByID map[string]core.LinearOperation, operationIDs []string, typeName, returnLiteral string) error {
	for _, operationID := range operationIDs {
		operation, known := operationsByID[operationID]
		if !known {
			return fmt.Errorf("block references unknown operation %q", operationID)
		}
		source, sourceKnown := places[operation.SourceID]
		if !sourceKnown {
			return fmt.Errorf("operation %q has invalid source", operation.ID)
		}
		switch operation.Kind {
		case core.OpCopy, core.OpMove, core.OpBorrowShared, core.OpBorrowExclusive:
			target, exists := places[operation.TargetID]
			if !exists {
				return fmt.Errorf("operation %q has invalid target", operation.ID)
			}
			label := "copy"
			if operation.Kind == core.OpMove {
				label = "authority transfer"
			} else if operation.Kind == core.OpBorrowShared {
				label = "shared borrow representation"
			} else if operation.Kind == core.OpBorrowExclusive {
				label = "exclusive borrow representation"
			}
			fmt.Fprintf(out, "      %s %s = %s; /* %s: %s */\n", typeName, locals[target.ID], locals[source.ID], label, operation.ID)
			fmt.Fprintf(out, "      (void)%s;\n", locals[target.ID])
			eventKind := "value.copied"
			if operation.Kind == core.OpMove {
				eventKind = "value.transferred"
			} else if operation.Kind == core.OpBorrowShared {
				eventKind = "value.borrowed"
			} else if operation.Kind == core.OpBorrowExclusive {
				eventKind = "value.borrowed_exclusive"
			}
			fmt.Fprintf(out, "      if (!lang_record_event(%s, %s, %s, %s, %s, %s)) return 74;\n",
				strconv.Quote(eventKind), strconv.Quote(operation.ID+":event"), strconv.Quote(function.ID),
				strconv.Quote(operation.SourceID), strconv.Quote(operation.TargetID), strconv.Quote(operation.TypeID))
		case core.OpReturn:
			fmt.Fprintf(out, "      if (!lang_record_event(%s, %s, %s, %s, NULL, %s)) return 74; /* returned place: %s */\n",
				strconv.Quote("function.returned"), strconv.Quote(operation.ID+":event:returned"), strconv.Quote(function.ID),
				strconv.Quote(operation.SourceID), strconv.Quote(operation.TypeID), operation.ID)
			out.WriteString("      if (!lang_write_literal(\"{\\\"schema\\\":\\\"lang.execution/1\\\",\\\"outcome\\\":{\\\"kind\\\":\\\"returned\\\",\\\"value\\\":\")) return 74;\n")
			fmt.Fprintf(out, "      if (!lang_write_json_string(%s)) return 74;\n", strconv.Quote(returnLiteral))
			out.WriteString("      if (!lang_write_literal(\"},\\\"events\\\":[\")) return 74;\n")
			out.WriteString("      if (!lang_write_events()) return 74;\n")
			out.WriteString("      if (!lang_write_literal(\"],\\\"live_resources\\\":[]}\\n\")) return 74;\n")
			out.WriteString("      return 0;\n")
		case core.OpDefect:
			// D-04-15: lower to a call to the generated _Noreturn lang_defect
			// function, every path of which ends in abort(). The terminal
			// record (outcome kind "defect", no value, its own
			// function.defected event carrying the required reason string in
			// "output", empty live_resources -- no arm can carry a foreign
			// acquisition this phase) is written FIRST, so it is captured
			// even though the process then aborts; lang_defect's own
			// _Noreturn marker is the one and only exemption from the
			// zero-attribute control (D-04-13/D-04-14), because it is a
			// property of a function cgen itself emits, not an unproven
			// claim about a foreign callee.
			//
			// This event is assembled directly rather than through
			// lang_record_event/lang_write_events: the shared LANG_EVENT
			// struct those helpers use (emitEventSupport, frozen for every
			// other emitter, D-04-23) has no "output" field, and adding one
			// there would move every existing committed generated-C golden.
			out.WriteString("      if (!lang_write_literal(\"{\\\"schema\\\":\\\"lang.execution/1\\\",\\\"outcome\\\":{\\\"kind\\\":\\\"defect\\\",\\\"value\\\":\\\"\\\"},\\\"events\\\":[\")) return 74;\n")
			out.WriteString("      if (!lang_write_events()) return 74;\n")
			out.WriteString("      if (lang_event_count != 0u && !lang_write_bytes(\",\", 1u)) return 74;\n")
			out.WriteString("      if (!lang_write_literal(\"{\\\"schema\\\":\\\"lang.execution/1\\\",\\\"id\\\":\")) return 74;\n")
			fmt.Fprintf(out, "      if (!lang_write_json_string(%s)) return 74;\n", strconv.Quote(operation.ID+":event:defected"))
			out.WriteString("      if (!lang_write_literal(\",\\\"kind\\\":\\\"function.defected\\\",\\\"function_id\\\":\")) return 74;\n")
			fmt.Fprintf(out, "      if (!lang_write_json_string(%s)) return 74;\n", strconv.Quote(function.ID))
			out.WriteString("      if (!lang_write_literal(\",\\\"source_place\\\":\")) return 74;\n")
			fmt.Fprintf(out, "      if (!lang_write_json_string(%s)) return 74;\n", strconv.Quote(operation.SourceID))
			out.WriteString("      if (!lang_write_literal(\",\\\"type_id\\\":\")) return 74;\n")
			fmt.Fprintf(out, "      if (!lang_write_json_string(%s)) return 74;\n", strconv.Quote(operation.TypeID))
			out.WriteString("      if (!lang_write_literal(\",\\\"output\\\":\")) return 74;\n")
			fmt.Fprintf(out, "      if (!lang_write_json_string(%s)) return 74;\n", strconv.Quote(operation.Reason))
			out.WriteString("      if (!lang_write_literal(\"}\")) return 74;\n")
			out.WriteString("      if (!lang_write_literal(\"],\\\"live_resources\\\":[]}\\n\")) return 74;\n")
			fmt.Fprintf(out, "      lang_defect(%s);\n", strconv.Quote(operation.Reason))
			out.WriteString("      return 71; /* unreachable: lang_defect never returns */\n")
		case core.OpForeignCall, core.OpFail:
			// checkBranch never emits either kind inside a match arm body
			// this phase (no `try` support inside an arm) -- named here,
			// rather than falling into default:, so control:kind.exhaustive
			// _dispatch's six-site table finds a known-but-unsupported case
			// rather than an unknown one (D-04-22).
			return fmt.Errorf("operation %q: foreign calls inside a match arm body are not supported this phase", operation.ID)
		default:
			return fmt.Errorf("operation %q has unknown kind %q", operation.ID, operation.Kind)
		}
	}
	return nil
}

// ---------------------------------------------------------------------
// Phase 4 plan 03: the three inspectable layers derived from one
// authoritative core.ForeignContract (D-04-12), plus the zero-attribute
// control (D-04-13). All three emitters below are new top-level functions,
// siblings of Emit/EmitNative -- none is invoked from either, so a program's
// ordinary Emit/EmitNative output is completely unaffected by this section.
// ---------------------------------------------------------------------

// ForeignManifestSchema identifies the lang.foreign/0 sidecar manifest
// (D-04-12c): a separate schema/artifact from lang.core/*, digest-bound into
// evidence.Manifest.ForeignDigest but never merged into the core artifact
// itself.
const ForeignManifestSchema = "lang.foreign/0"

// foreignManifestDocument is the lang.foreign/0 sidecar's exact field
// layout: the complete core.ForeignContract plus two fields no
// ForeignContract itself carries. EmittedAttributes has no omitempty tag
// (D-04-13): it is deliberately a present, empty JSON array this phase, not
// an omission -- the distinction Phase 5 needs to assert on.
// UncheckedObligations names every obligation this phase declares but never
// exercises, so a quarantine reader never mistakes a declared fact for a
// proven one (D-04-12/D-10).
type foreignManifestDocument struct {
	Schema               string             `json:"schema"`
	Symbol               string             `json:"symbol"`
	Allocator            string             `json:"allocator"`
	Unwind               string             `json:"unwind"`
	NonlocalExit         string             `json:"nonlocal_exit"`
	Fails                string             `json:"fails"`
	InitializedState     string             `json:"initialized_state"`
	Capture              string             `json:"capture"`
	Retention            string             `json:"retention"`
	Aliasing             string             `json:"aliasing"`
	Layout               *core.RecordLayout `json:"layout"`
	EmittedAttributes    []string           `json:"emitted_attributes"`
	UncheckedObligations []string           `json:"unchecked_obligations"`
}

// uncheckedForeignObligations names every obligation this phase declares but
// never exercises (D-04-12c): the language has no closures, no threads, and
// no calls into Lang, so capture, retention, aliasing, and callback
// retention are declared facts, never proven ones. Kept as a function
// (rather than a package var) so a future phase narrowing this list has one
// call site to change, and so EmitForeignManifest and any test asserting on
// this list read the exact same values.
func uncheckedForeignObligations() []string {
	return []string{
		"capture", "retention", "aliasing", "callback_retention",
		// D-04-21/T-04-33: the process-root nonlocal-exit landing pad
		// (D-04-17) has two accepted residual blind spots, named here rather
		// than silently implied covered -- neither is observable by a single
		// process-root pad. See emitNonlocalPad's own doc comment and
		// testdata/phase4/nonlocal_exit_probe.lang's REACHABILITY comment.
		"nonlocal_exit_below_pad", "nonlocal_exit_process_termination",
	}
}

// singleForeignFunction returns the one function in program that declares a
// foreign contract, or an error if none does. Every Phase 4 foreign-shaped
// corpus fixture is single-function, matching Emit/EmitNative's own
// single-function expectation.
func singleForeignFunction(program core.Program) (core.Function, error) {
	for _, function := range program.Functions {
		if function.ForeignContract != nil {
			// Independent refusal (04-VERIFICATION.md gap 2, FFI-01/D-04-12):
			// EmitForeignManifest, EmitForeignHeader and EmitForeignConformance
			// all resolve their function through this call and NONE of them
			// calls corevalidate.Validate, so this is the one place that can
			// refuse a hostile Symbol before any of the three splices it.
			if !validForeignSymbol(function.ForeignContract.Symbol) {
				return core.Function{}, fmt.Errorf("foreign symbol %q is not a C identifier", function.ForeignContract.Symbol)
			}
			return function, nil
		}
	}
	return core.Function{}, fmt.Errorf("program declares no foreign contract")
}

// EmitForeignManifest serializes program's single foreign contract as a
// lang.foreign/0 sidecar manifest document (D-04-12c): the JSON is
// authoritative, and EmitForeignHeader's obligation comment block is
// generated FROM the same contract value, so the two can never drift
// (D-04-12).
func EmitForeignManifest(program core.Program) (string, error) {
	function, err := singleForeignFunction(program)
	if err != nil {
		return "", err
	}
	contract := function.ForeignContract
	document := foreignManifestDocument{
		Schema: ForeignManifestSchema, Symbol: contract.Symbol, Allocator: contract.Allocator,
		Unwind: contract.Unwind, NonlocalExit: contract.NonlocalExit, Fails: contract.Fails,
		InitializedState: contract.InitializedState, Capture: contract.Capture, Retention: contract.Retention, Aliasing: contract.Aliasing,
		Layout: contract.Layout, EmittedAttributes: []string{}, UncheckedObligations: uncheckedForeignObligations(),
	}
	encoded, err := json.Marshal(document)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

// foreignHeaderResultType is the generated header's own typedef name for the
// symbol's by-value ABI result -- allocated through the existing cNames
// deterministic-suffix-on-collision machinery (D-04-12a's "not a second ad
// hoc namer" requirement), not a hand-built string.
func foreignHeaderResultType(names *cNames, symbol string) string {
	return names.allocate(cName(symbol)+"_RESULT", "foreign_result_type", 0)
}

// EmitForeignHeader generates the `_LANG_`-namespaced header for program's
// single declared foreign symbol (D-04-12a): the extern declaration, a
// generated obligation comment block reproducing every contract field, and
// self-layout _Static_assert()s over Lang's own generated record. Every line
// of the comment block is written directly from the SAME core.ForeignContract
// value EmitForeignManifest serializes -- a test mutating one contract field
// and re-running both must see the corresponding comment line move,
// because there is exactly one source of truth for both (D-04-12: a
// hand-written comment beside a generated JSON is a second source of truth
// that will drift, which this shared-source construction forbids by
// design).
func EmitForeignHeader(program core.Program) (string, error) {
	function, err := singleForeignFunction(program)
	if err != nil {
		return "", err
	}
	contract := function.ForeignContract
	names := newCNames(linearFixedNames...)
	resultType := foreignHeaderResultType(names, contract.Symbol)
	symbolC := foreignExternName(contract.Symbol)
	guard := "LANG_FOREIGN_" + strings.ToUpper(cName(contract.Symbol)) + "_H"

	var out strings.Builder
	out.WriteString("/* generated by Codename Lang; schema lang.c17/0 (foreign header, D-04-12a) */\n")
	fmt.Fprintf(&out, "#ifndef %s\n#define %s\n\n", guard, guard)
	out.WriteString("#include <stddef.h>\n\n")
	out.WriteString("/* lang.foreign/0 obligations -- generated from the sidecar manifest;\n")
	out.WriteString(" * see EmitForeignManifest. Never hand-edit this block: a hand-written\n")
	out.WriteString(" * comment beside a generated JSON is a second source of truth that will\n")
	out.WriteString(" * drift, which is exactly what D-04-12 forbids. */\n")
	fmt.Fprintf(&out, "/* symbol: %s */\n", contract.Symbol)
	fmt.Fprintf(&out, "/* allocator: %s */\n", contract.Allocator)
	fmt.Fprintf(&out, "/* unwind: %s */\n", contract.Unwind)
	fmt.Fprintf(&out, "/* nonlocal_exit: %s */\n", contract.NonlocalExit)
	fmt.Fprintf(&out, "/* fails: %s */\n", contract.Fails)
	fmt.Fprintf(&out, "/* initialized_state: %s */\n", contract.InitializedState)
	fmt.Fprintf(&out, "/* capture: %s (unchecked_obligation) */\n", contract.Capture)
	fmt.Fprintf(&out, "/* retention: %s (unchecked_obligation) */\n", contract.Retention)
	fmt.Fprintf(&out, "/* aliasing: %s (unchecked_obligation) */\n", contract.Aliasing)
	if contract.Layout != nil {
		fmt.Fprintf(&out, "/* layout.foreign_type_name: %s */\n", contract.Layout.ForeignTypeName)
		fmt.Fprintf(&out, "/* layout.size: %d */\n", contract.Layout.Size)
		fmt.Fprintf(&out, "/* layout.alignment: %d */\n", contract.Layout.Alignment)
		for _, field := range contract.Layout.Fields {
			fmt.Fprintf(&out, "/* layout.field: %s size=%d alignment=%d offset=%d */\n", field.Name, field.Size, field.Alignment, field.Offset)
		}
	}
	out.WriteString("\n")
	fmt.Fprintf(&out, "typedef struct %s {\n  unsigned char ok;\n  unsigned char value;\n} %s;\n\n", resultType, resultType)
	fmt.Fprintf(&out, "extern %s %s(unsigned char argument);\n\n", resultType, symbolC)
	fmt.Fprintf(&out, "_Static_assert(sizeof(%s) == 2, \"%s must be a two-byte by-value ABI result\");\n", resultType, resultType)
	fmt.Fprintf(&out, "_Static_assert(_Alignof(%s) == 1, \"%s must have byte alignment\");\n", resultType, resultType)
	fmt.Fprintf(&out, "_Static_assert(offsetof(%s, ok) == 0, \"%s.ok must be the first field\");\n", resultType, resultType)
	fmt.Fprintf(&out, "_Static_assert(offsetof(%s, value) == 1, \"%s.value must follow ok\");\n", resultType, resultType)
	out.WriteString("\n#endif\n")
	return out.String(), nil
}

// EmitForeignConformance generates lang_foreign_conformance.c (D-04-11): the
// single, explicit, auditable translation unit where Lang's own generated
// declaration (EmitForeignHeader) and the foreign translation unit's private
// header at privateHeaderPath are permitted to meet. It defines no symbol --
// it exists only to be compiled, never linked -- and carries one
// _Static_assert triple (sizeof/_Alignof/offsetof) per declared record field,
// plus record-level size and alignment assertions, over
// contract.Layout.ForeignTypeName as declared in the included private
// header. Per D-04-10, no other compiler-produced artifact may include
// privateHeaderPath.
func EmitForeignConformance(program core.Program, privateHeaderPath string) (string, error) {
	function, err := singleForeignFunction(program)
	if err != nil {
		return "", err
	}
	contract := function.ForeignContract
	if contract.Layout == nil || contract.Layout.ForeignTypeName == "" {
		return "", fmt.Errorf("foreign contract %q has no layout to prove", contract.Symbol)
	}
	header, err := EmitForeignHeader(program)
	if err != nil {
		return "", err
	}
	layout := contract.Layout

	var out strings.Builder
	out.WriteString("/* generated by Codename Lang; schema lang.c17/0 (conformance TU, D-04-11).\n")
	out.WriteString(" * This is the single explicit, auditable place Lang's own declaration and\n")
	out.WriteString(" * the foreign translation unit's private header are permitted to meet. It\n")
	out.WriteString(" * defines no symbol and is compiled but never linked. */\n")
	out.WriteString(header)
	fmt.Fprintf(&out, "#include %q\n\n", privateHeaderPath)
	fmt.Fprintf(&out, "_Static_assert(sizeof(%s) == %d, \"%s size must match the declared layout\");\n", layout.ForeignTypeName, layout.Size, layout.ForeignTypeName)
	fmt.Fprintf(&out, "_Static_assert(_Alignof(%s) == %d, \"%s alignment must match the declared layout\");\n", layout.ForeignTypeName, layout.Alignment, layout.ForeignTypeName)
	for _, field := range layout.Fields {
		cType := field.CType
		if cType == "" {
			cType = "unsigned char"
		}
		fmt.Fprintf(&out, "_Static_assert(sizeof(%s) == %d, \"%s.%s size must match the declared layout\");\n", cType, field.Size, layout.ForeignTypeName, field.Name)
		fmt.Fprintf(&out, "_Static_assert(_Alignof(%s) == %d, \"%s.%s alignment must match the declared layout\");\n", cType, field.Alignment, layout.ForeignTypeName, field.Name)
		fmt.Fprintf(&out, "_Static_assert(offsetof(%s, %s) == %d, \"%s.%s offset must match the declared layout\");\n", layout.ForeignTypeName, field.Name, field.Offset, layout.ForeignTypeName, field.Name)
	}
	return out.String(), nil
}

// BannedOptimizerAttributes is D-04-13's single Go constant enumerating
// every optimizer-visible attribute token cgen must never emit this phase.
// Both the emitted-C/manifest scan below and TestAttributeInjectionMakesControlFail's
// own mutation-kill injection consume this SAME slice, so widening the set
// cannot silently bypass the scan.
var BannedOptimizerAttributes = []string{
	"restrict", "noalias", "nothrow", "__attribute__((malloc))", "nonnull", "returns_nonnull",
}

// NoreturnExemption is D-04-14's one named exemption from the zero-attribute
// control: `_Noreturn` on cgen's own generated defect function is a property
// of a function cgen itself emits -- every path of which provably ends in
// abort() -- not a claim about a foreign callee. It is not a member of
// BannedOptimizerAttributes and ScanForBannedAttributes never scans for it;
// named here so the exemption is explicit rather than an silent omission.
const NoreturnExemption = "_Noreturn"

// ScanForBannedAttributes scans every given emitted C artifact (and, by the
// caller passing EmitForeignManifest's own output, the sidecar manifest's
// emitted_attributes field) for any BannedOptimizerAttributes token,
// returning every match found across all sources. Nil/empty when no banned
// token appears anywhere -- the required-control state.
func ScanForBannedAttributes(sources ...string) []string {
	var found []string
	for _, source := range sources {
		for _, token := range BannedOptimizerAttributes {
			if strings.Contains(source, token) {
				found = append(found, token)
			}
		}
	}
	return found
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

// allocate returns preferred unchanged when it is not already taken, or a
// deterministic collision-suffixed candidate otherwise. INVARIANT (D-06 /
// D-02-05, enforced by TestNativeIdentifiersRemainCollisionFree): the suffix
// spelling must never introduce two adjacent underscores anywhere in the
// result. C17 section 7.1.3 reserves to the implementation every identifier
// that contains a double underscore, in any position, in the ordinary
// identifier namespace — not only identifiers that begin with one. A suffix
// built as "__LANG_" would therefore hand out reserved names on every actual
// collision, in generated code a conforming implementation is permitted to
// treat specially. "_LANG_" (a single leading underscore) has no double
// underscore and stays inside the confined "LANG_"/"lang_value_" namespaces
// documented above, so it is never reserved and never re-issues a reserved
// name.
func (n *cNames) allocate(preferred, category string, ordinal int) string {
	if _, exists := n.used[preferred]; !exists {
		n.used[preferred] = struct{}{}
		return preferred
	}
	base := preferred + "_LANG_" + strings.ToUpper(category) + "_" + strconv.Itoa(ordinal)
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
