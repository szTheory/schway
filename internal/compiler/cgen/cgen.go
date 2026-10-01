package cgen

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
)

const Schema = "lang.c17/0"

// opCallGroupedArmForTest is Task 3's D-07-41/D-07-42 fault-injection seam
// (QLT-08, plan 07-04's Test 4): when true, emitLinear's core.OpCall case
// is folded into the SAME grouped behaviour core.OpCopy/OpMove/
// OpBorrowShared/OpBorrowExclusive use, emitting copy-like C for a call and
// producing NO error at all -- exactly the stub-certification failure
// D-07-39 exists to prevent, since the emitted C would then look like a
// successful, ordinary value transfer with no callee ever invoked.
// Unexported, false by default, exercised only via the exported test-only
// wrapper EmitLinearForTest (export_test.go) from the external cgen_test
// package: never an exported package-level mutable var on a production
// path.
var opCallGroupedArmForTest = false

// payloadSlotSwapForTest is D-12-38 Task 2's fault-injection seam
// (session_payload_control_test.go's TestPayloadSlotSwapMutationKilled):
// when true, OpConstructPayload's case in emitBranchOperations writes the
// source payload into a DIFFERENT alternative's struct field than the one
// the (correct) tag names -- reads or writes the wrong alternative's
// payload slot for a correct tag, never touching which tag constant is
// written. Unexported, false by default, exercised only via the exported
// test-only wrapper SetPayloadSlotSwapForTest (export_test.go) from the
// external session_test package: never an exported package-level mutable
// var on a production path, following opCallGroupedArmForTest's own shape
// exactly.
var payloadSlotSwapForTest = false

// SetPayloadSlotSwapForTest installs/restores payloadSlotSwapForTest. Callers
// MUST defer the returned restore func immediately. This is a
// PRODUCTION-VISIBLE function (never an export_test.go symbol) because the
// D-12-38 mutation-kill test lives in a DIFFERENT package
// (internal/compiler/session's session_test), and Go's _test.go export
// trick (as SetOpCallGroupedArmForTest above uses) is visible only within
// cgen's OWN test binary -- following corevalidate.SetDisableCyclePeerForTest
// / interp.SetDisableEmptyTagSerializationForTest's own D-07-42/D-12-38
// cross-package precedent exactly (12-04-SUMMARY.md). A documented,
// clearly-named, test-only no-op; never called from any production code
// path in this repository.
func SetPayloadSlotSwapForTest(mutate bool) (restore func()) {
	previous := payloadSlotSwapForTest
	payloadSlotSwapForTest = mutate
	payloadSlotSwapInjectedWriteCount = 0
	return func() { payloadSlotSwapForTest = previous }
}

// payloadSlotSwapInjectedWriteCount counts how many times
// emitBranchOperations' OpConstructPayload case actually took
// wrongPayloadSlot's successful-target (ok == true) branch and emitted a
// wrong-slot write, WR-02's anti-vacuity counter. It is incremented ONLY
// inside that successful branch -- never in the else fallback (the
// documented no-op path wrongPayloadSlot's own doc comment names) and
// never on the non-mutated path -- and reset to zero by
// SetPayloadSlotSwapForTest when the seam is engaged, so
// TestPayloadSlotSwapMutationKilled can assert a wrong-slot write was
// ACTUALLY injected rather than trusting that engaging the flag alone
// proves anything. WR-02 identified that a fixture-shape regression (the
// data type losing its second payload-carrying alternative) could make
// wrongPayloadSlot silently fall back to the correct write on every call,
// letting the mutation-kill control pass while proving nothing. The
// restore closure returned by SetPayloadSlotSwapForTest deliberately does
// NOT reset this counter, so a caller can read it after `defer restore()`
// runs.
var payloadSlotSwapInjectedWriteCount int

// PayloadSlotSwapInjectedWriteCount reports how many wrong-slot writes
// D-12-38's fault-injection seam has actually emitted since the seam was
// last engaged via SetPayloadSlotSwapForTest(true). This is a
// PRODUCTION-VISIBLE function (never an export_test.go symbol), for the
// same cross-package reason SetPayloadSlotSwapForTest's own doc comment
// gives: the assertion reading it lives in
// internal/compiler/session's session_test package, an external consumer
// of cgen's normal (non-test) build, for which an export_test.go symbol
// does not exist at all -- _test.go exports are visible only within
// cgen's own test binary. A documented, clearly-named, test-only reader;
// never called from any production code path in this repository.
func PayloadSlotSwapInjectedWriteCount() int {
	return payloadSlotSwapInjectedWriteCount
}

// wrongPayloadSlot picks a DIFFERENT alternative's own struct field name
// and C type than altName's, for D-12-38's fault-injection seam above. It
// returns ok == false when dataType has no other payload-carrying
// alternative to misdirect into (the seam then falls back to the correct
// write, a documented no-op rather than a silent skip).
func wrongPayloadSlot(dataType core.DataType, payloadFieldBySource map[string]string, altName string) (field, cType string, ok bool) {
	for _, detail := range dataType.AlternativeDetails {
		if detail.Name == altName || detail.PayloadType == "" {
			continue
		}
		if wrongField, known := payloadFieldBySource[detail.Name]; known {
			return wrongField, payloadCTypeName(detail.PayloadType), true
		}
	}
	return "", "", false
}

func Emit(program core.Program) (string, error) {
	validated := corevalidate.Validate(program)
	if !validated.Valid {
		return "", fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
	}
	return emitProgram(validated.Program(), false)
}

// EmitNative shares Emit's one production lowering authority. The program
// emitter owns the execution document contract at the child-process boundary.
func EmitNative(program core.Program) (string, error) {
	validated := corevalidate.Validate(program)
	if !validated.Valid {
		return "", fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
	}
	return emitProgram(validated.Program(), true)
}

// EmitApplication lowers a checked U64-to-U64 entry through the same function
// bodies as the conformance emitter, with an ordinary decimal application
// boundary instead of an execution-document boundary.
func EmitApplication(program core.Program) (string, error) {
	validated := corevalidate.Validate(program)
	if !validated.Valid {
		return "", fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
	}
	return emitProgramWithShell(validated.Program(), programApplicationShell, false)
}

// The generated-C ordinary-identifier namespace is closed by two cooperating
// properties, not by the reserved lists alone. Both are load-bearing and both
// are enforced by tests (see cgen_names_test.go and names_internal_test.go):
//
//  1. PREFIX CONFINEMENT. Every identifier the allocator can ever hand out is
//     confined to one of exactly two namespaces: cName always returns
//     "SCHWAY_" + <[A-Z0-9_]* tail> and cLocal always returns
//     "schway_value_" + <[A-Za-z0-9_]* tail>. The derived preferred names built
//     on top of them (alternative names, the "<Type>_name" helper) keep the
//     "SCHWAY_" prefix, and cNames.allocate only ever appends to a preferred
//     name, so every allocated identifier stays inside those two namespaces.
//
//  2. HONEST RESERVATION. matchFixedNames and linearFixedNames are supersets of
//     every ordinary identifier the emitters write out themselves. A fixed
//     identifier that falls inside the "SCHWAY_"/"schway_value_" namespaces is
//     therefore also reserved, so the allocator can never re-issue it.
//
// Property 1 alone would make most of the reservations unreachable, and
// property 2 alone would be enough only if it were maintained perfectly. Keeping
// both — and testing both — means a future emitter change breaks a test rather
// than silently aliasing two distinct core identities onto one C identifier.
//
// The lists are deliberately supersets: a name is kept even when no current
// program-emission path writes it. Over-reservation is inert, because
// property 1 guarantees no preferred name can equal a lowercase or
// non-"SCHWAY_"-prefixed reserved entry; under-reservation is the dangerous
// direction, so the tests only forbid that one.

// matchFixedNames is every ordinary identifier emitMatch writes itself,
// excluding C keywords and the libc names it calls.
var matchFixedNames = []string{
	"main", "argc", "argv", "input", "output", "name", "value",
	"SCHWAY_EVENT", "SCHWAY_EVENT_CAPACITY", "SCHWAY_OUTPUT_LIMIT", "schway_events",
	"schway_event_count", "schway_output_count", "schway_write_bytes", "schway_write_literal",
	"schway_write_json_string", "schway_write_json_string_content", "schway_record_event", "schway_write_events",
	"schway_write_live_resources", "schway_invocations", "schway_entry_input", "schway_entry_output",
	"invocation", "invocation_index", "callee_function_id",
	"abort", "byte", "data", "encoded", "escape", "event", "function_id", "hex", "id",
	"index", "kind", "schway_entry_name", "length", "source_place", "target_place", "type_id",
}

// linearFixedNames is the reserved ordinary-identifier vocabulary for
// whole-program emission of linear-shaped functions. It covers emitted
// macros, typedefs, struct members, globals, helper functions, and locals.
var linearFixedNames = []string{
	// macros and typedefs
	"SCHWAY_BUFFER", "SCHWAY_EVENT", "SCHWAY_OUTPUT_LIMIT", "SCHWAY_EVENT_CAPACITY",
	// SCHWAY_BUFFER and SCHWAY_EVENT struct members
	"bytes", "length",
	"kind", "id", "function_id", "source_place", "target_place", "type_id",
	// file-scope globals
	"schway_events", "schway_event_count", "schway_output_count",
	// helper functions
	"schway_write_bytes", "schway_write_literal", "schway_write_json_string", "schway_write_json_string_content",
	"schway_record_event", "schway_write_events", "schway_write_buffer_hex", "schway_write_byte",
	"schway_write_u64", "schway_parse_u64_decimal",
	// Historical foreign resource-ledger output names remain reserved.
	"schway_resource_ids", "schway_resource_live", "schway_write_live_resources", "first",
	// Historical foreign nonlocal-exit output name.
	"schway_nonlocal_landing",
	// helper parameters and locals
	"data", "value", "hex", "byte", "escape", "encoded", "event", "index",
	// main
	"main", "argc", "argv", "input",
	"abort", "callee_function_id", "invocation", "invocation_index", "schway_entry_input",
	"schway_entry_output", "schway_invocations",
}

// selectsByPointerLowering is D-05-02's structural shape classifier. It
// returns true only when function's sole parameter is exclusively borrowed
// as literally the FIRST operation of a straight-line (Match-less,
// block-less) linear body, is never referenced again directly anywhere else
// in the operation stream, and every remaining operation forms one unbroken
// derivation chain rooted at that borrow's own target place and ending
// exactly at the function's OpReturn terminator. That is precisely "an
// exclusive loan covering every operation from the parameter's first use to
// the function terminator" (D-05-02) -- derived entirely from the core
// artifact's own LinearOperation.SourceID/TargetID/Kind facts, never a
// fixture name, function name, or allowlist. It is used to describe foreign
// manifest metadata and classify shapes that emitProgram refuses; it does
// not select a production lowering path. Branch-shaped and foreign-call
// bodies do not match this classifier.
//
// PublicOrigin == nil is also required: a function whose return type
// carries a declared `borrow(path)` annotation is already a distinct,
// previously-shipped semantic category (OWN-04's public borrowed views,
// e.g. testdata/phase3/public_view_mixed_access.schway) that can have the
// exact same exclusive-borrow-then-reborrow-to-terminator operation shape
// as the historical by-pointer fixture -- PublicOrigin is the one genuinely
// structural fact (not a name or file match) that tells the two apart, and
// is exactly what TestPhase5ByPointerLoweringIsAdditive asserts keeps every
// Phase 1-4 fixture on its own existing lowering path.
// SelectsByPointerLowering is the exported form of selectsByPointerLowering
// (D-05-02): promoted from a test-only accessor (export_test.go) to a real
// production export this plan, so check package's own tests can prove
// deriveAliasFacts (check.go) and this predicate decide the SAME condition
// on every corpus fixture (TestAliasFactAgreesWithByPointerSelection) --
// D-12's three independent derivations still share zero HELPERS with each
// other (this is read-only cross-package test verification, not a shared
// implementation). Emit/EmitNative route through emitProgram, which refuses
// by-pointer program bodies before serialization.
func SelectsByPointerLowering(function core.Function, linear *core.LinearBody) bool {
	return selectsByPointerLowering(function, linear)
}

func selectsByPointerLowering(function core.Function, linear *core.LinearBody) bool {
	if function.Match != nil || function.PublicOrigin != nil || linear == nil || len(linear.Blocks) > 0 {
		return false
	}
	operations := linear.Operations
	if len(operations) == 0 {
		return false
	}
	first := operations[0]
	if first.Kind != core.OpBorrowExclusive || first.SourceID != function.Parameter.ID || first.TargetID == "" {
		return false
	}
	for _, operation := range operations[1:] {
		if operation.SourceID == function.Parameter.ID {
			return false
		}
	}
	current := first.TargetID
	terminatorIndex := -1
	for index := 1; index < len(operations); index++ {
		operation := operations[index]
		if operation.SourceID != current {
			return false
		}
		if operation.Kind == core.OpReturn {
			terminatorIndex = index
			break
		}
		if operation.TargetID == "" {
			return false
		}
		current = operation.TargetID
	}
	return terminatorIndex == len(operations)-1
}

// AttributeSuppressionProfile is D-11-16/D-11-17's permanent suppression
// control for by-pointer attribute metadata emitted by the foreign manifest
// API. It does not admit a by-pointer program body.
type AttributeSuppressionProfile int

const (
	// AttributesJustified is the metadata API's zero value and retains the
	// established D-05-01..D-05-04 attribute claim, completely unaffected by
	// Phase 11's call-boundary zero-attribute claim
	// (D-11-09), which is scoped to Lang-to-Lang calls and never to this
	// pre-existing FFI by-pointer boundary. Every production caller that
	// never touches AttributeSuppressionProfile observes this value.
	AttributesJustified AttributeSuppressionProfile = iota
	// AttributesSuppressed withholds the manifest attribute even when the
	// structural by-pointer condition would otherwise justify it. It
	// exists ONLY as the mid-phase gate's own bisection tool (D-11-18):
	// proving cgen.ScanForBannedAttributes' conjunct is not vacuous by
	// re-running the same corpus through the same code path and watching
	// the scan's own verdict move. Declared PERMANENT, never scheduled for
	// removal: the attributes-off vs attributes-on vs
	// interpreter-inertness differential at -O3/-flto is a standing,
	// repeatable test of the qualifier's own semantic-inertness claim, and
	// this project's own bisection tool for "our promise, or Clang?" --
	// precisely the shape LLVM's own -opt-bisect-limit/OptPassGate ships
	// as a permanent transformation control, never a temporary flag
	// scheduled for deletion.
	AttributesSuppressed
)

// attributeSuppressionProfile is the package-level active profile.
// Production code paths never mutate this directly; only
// SetAttributeSuppressionProfileForTest does, and only from a test.
var attributeSuppressionProfile = AttributesJustified

// SetAttributeSuppressionProfileForTest installs profile as the active
// AttributeSuppressionProfile and returns a restore func; callers MUST
// defer it immediately. This is a real, production export (not
// export_test.go) because session's own mid-phase gate tests, in a
// different package, are this helper's only caller -- mirroring
// SelectsByPointerLowering's own cross-package promotion precedent
// (D-12).
func SetAttributeSuppressionProfileForTest(profile AttributeSuppressionProfile) (restore func()) {
	previous := attributeSuppressionProfile
	attributeSuppressionProfile = profile
	return func() { attributeSuppressionProfile = previous }
}

// selectsByPointerLoweringSharedOnly is D-05-05's adversarial-harness
// sibling of selectsByPointerLowering: the exact same structural shape
// (one borrow of the sole parameter as the function's first operation,
// never referenced directly again, every remaining operation forming one
// unbroken chain to the OpReturn terminator), except it requires the FIRST
// borrow to be OpBorrowShared, never OpBorrowExclusive. It exists to give
// control:alias.false_no_alias (session package) an honest, ENGINEERED
// subject: a shared loan alone never justifies `restrict` (unlike an
// exclusive loan, D-05-01), so a function selected here is lowered by
// pointer for structural symmetry with the exclusive case but carries no
// restrict qualifier of its own. emitProgram uses this classifier to refuse
// the shape; manifest metadata may still describe it.
// Mutually exclusive with selectsByPointerLowering by construction (the
// two predicates differ only in first.Kind, which can never be both
// OpBorrowExclusive and OpBorrowShared), so no function selectsByPointerLowering
// already selects is ever affected by this addition, and no existing
// Phase 1-4 (or restrict_borrow.schway) golden changes.
func selectsByPointerLoweringSharedOnly(function core.Function, linear *core.LinearBody) bool {
	if function.Match != nil || function.PublicOrigin != nil || linear == nil || len(linear.Blocks) > 0 {
		return false
	}
	operations := linear.Operations
	if len(operations) == 0 {
		return false
	}
	first := operations[0]
	if first.Kind != core.OpBorrowShared || first.SourceID != function.Parameter.ID || first.TargetID == "" {
		return false
	}
	for _, operation := range operations[1:] {
		if operation.SourceID == function.Parameter.ID {
			return false
		}
	}
	current := first.TargetID
	terminatorIndex := -1
	for index := 1; index < len(operations); index++ {
		operation := operations[index]
		if operation.SourceID != current {
			return false
		}
		if operation.Kind == core.OpReturn {
			terminatorIndex = index
			break
		}
		if operation.TargetID == "" {
			return false
		}
		current = operation.TargetID
	}
	return terminatorIndex == len(operations)-1
}

// pointerABIFact is the single checked shape consumed by the production
// declaration, call-site, body, and manifest writers. A pointer representation
// carries no alias, capture, alignment, or ownership promise by itself.
type pointerABIFact struct {
	PointerParameter bool
	ParameterCType   string
	ValueCType       string
	Access           string
	FunctionID       string
	ParameterID      string
}

func checkedSharedPointerABIFact(function core.Function) (pointerABIFact, bool) {
	if function.Parameter.Type != "U64" || function.ReturnType != "U64" || function.Linear == nil ||
		!selectsByPointerLoweringSharedOnly(function, function.Linear) || len(function.Linear.Operations) != 2 {
		return pointerABIFact{}, false
	}
	borrow, ret := function.Linear.Operations[0], function.Linear.Operations[1]
	if borrow.Kind != core.OpBorrowShared || borrow.SourceID != function.Parameter.ID || borrow.TargetID == "" || borrow.LoanID == "" ||
		ret.Kind != core.OpReturn || ret.SourceID != borrow.TargetID || ret.TypeID != borrow.TypeID {
		return pointerABIFact{}, false
	}
	parameterType := ""
	borrowType := ""
	for _, place := range function.Linear.Places {
		if place.ID == function.Parameter.ID {
			parameterType = place.TypeID
		}
		if place.ID == borrow.TargetID {
			borrowType = place.TypeID
		}
	}
	if parameterType == "" || parameterType != borrowType || parameterType != borrow.TypeID {
		return pointerABIFact{}, false
	}
	for _, fact := range function.Linear.Types {
		if fact.ID == parameterType && fact.Shape.Constructor == "U64" {
			return pointerABIFact{PointerParameter: true, ParameterCType: "const uint64_t *", ValueCType: "uint64_t", Access: "shared", FunctionID: function.ID, ParameterID: function.Parameter.ID}, true
		}
	}
	return pointerABIFact{}, false
}

// emittedAttributeForByPointerParameter returns D-05-01's restrict
// attribute metadata for a shape classified by selectsByPointerLowering.
// This manifest description does not admit the corresponding program body.
// The binding always exists in that case, since
// selectsByPointerLowering's own structural condition (an exclusive loan on
// the parameter, unbroken to the terminator) IS check's independently
// derived AliasFact condition (TestAliasFactAgreesWithByPointerSelection,
// check package) -- so this reads the loan ID directly off the operation
// stream cgen already owns (function.Linear.Operations[0].LoanID, the loan
// the first, exclusive-borrow operation creates) rather than importing
// check's AliasFact type (D-12: zero shared helpers).
func emittedAttributeForByPointerParameter(function core.Function) EmittedAttribute {
	loanID := ""
	if function.Linear != nil && len(function.Linear.Operations) > 0 {
		loanID = function.Linear.Operations[0].LoanID
	}
	return EmittedAttribute{Attr: "restrict", CoreNode: function.ID, Parameter: function.Parameter.ID, JustifiedBy: loanID}
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

// commentSafeForeignField is cgen's OWN, deliberate second implementation of
// corevalidate.commentSafe (04-13 Task 3, 04-VERIFICATION.md gap 2b,
// FFI-01): cgen must be able to refuse a hostile core.ForeignContract field
// without depending on corevalidate.Validate having run first, matching the
// independence posture validForeignSymbol already established for Symbol.
// An EMPTY string is comment-safe. Refuses "*/", "/*", any byte below 0x20,
// 0x7F, or any byte at or above 0x80 -- see corevalidate.commentSafe's doc
// comment for why "/*" is refused alongside "*/" (warnings-as-errors).
func commentSafeForeignField(value string) bool {
	if value == "" {
		return true
	}
	for index := 0; index < len(value); index++ {
		b := value[index]
		if b < 0x20 || b == 0x7F || b >= 0x80 {
			return false
		}
		if index+1 < len(value) {
			pair := value[index : index+2]
			if pair == "*/" || pair == "/*" {
				return false
			}
		}
	}
	return true
}

// validForeignCType is cgen's own second implementation of
// corevalidate.validCTypeExpression: Layout.Fields[].CType is spliced into a
// _Static_assert operand as a real C token sequence by EmitForeignConformance
// (below), not into a comment, so comment-safety alone is insufficient. An
// EMPTY string is valid (this file's own emission defaults it to
// "unsigned char"). Otherwise the value must be one or more C identifiers
// separated by exactly one single space.
func validForeignCType(value string) bool {
	if value == "" {
		return true
	}
	start := 0
	for index := 0; index <= len(value); index++ {
		if index == len(value) || value[index] == ' ' {
			if !validForeignSymbol(value[start:index]) {
				return false
			}
			start = index + 1
		}
	}
	return true
}

// unsafeForeignContractField is cgen's own independent audit of every
// core.ForeignContract string field EmitForeignHeader or
// EmitForeignConformance splices into generated C -- BOTH the three policy
// values (allocator, unwind, nonlocal_exit, each required to satisfy
// validForeignSymbol, matching corevalidate's identifier-shape audit) AND
// every remaining field corevalidate.foreignContractFieldsCSafe covers
// (cgen.go's own comment-block splice sites at what is now the block below,
// plus EmitForeignConformance's _Static_assert operands over
// Layout.ForeignTypeName/Fields[].Name/Fields[].CType). This is the SAME
// independence posture 04-12 established for Symbol and corevalidate takes
// toward check.go throughout this phase: cgen must be able to refuse the
// whole hostile contract without corevalidate.Validate having run, because
// EmitForeignManifest, EmitForeignHeader and EmitForeignConformance are
// exported entry points that never call it. Returns the empty string when
// every field is safe, otherwise the NAME of the first offending field --
// the name only, never the value, so no attacker-controlled byte reaches the
// caller.
func unsafeForeignContractField(contract *core.ForeignContract) string {
	if !validForeignSymbol(contract.Allocator) {
		return "allocator"
	}
	if !validForeignSymbol(contract.Unwind) {
		return "unwind"
	}
	if !validForeignSymbol(contract.NonlocalExit) {
		return "nonlocal_exit"
	}
	if !commentSafeForeignField(contract.Fails) {
		return "fails"
	}
	if !commentSafeForeignField(contract.InitializedState) {
		return "initialized_state"
	}
	if !commentSafeForeignField(contract.Capture) {
		return "capture"
	}
	if !commentSafeForeignField(contract.Retention) {
		return "retention"
	}
	if !commentSafeForeignField(contract.Aliasing) {
		return "aliasing"
	}
	// Alias is Phase 4 plan 06's additive aliasing-obligation field
	// (D-04-28); D-05-36 closes the carried D-04-33 audit gap by giving it
	// the exact same comment-safety check as its Aliasing sibling above.
	if !commentSafeForeignField(contract.Alias) {
		return "alias"
	}
	if contract.Layout != nil {
		if !validForeignSymbol(contract.Layout.ForeignTypeName) {
			return "layout.foreign_type_name"
		}
		for _, field := range contract.Layout.Fields {
			if !validForeignSymbol(field.Name) {
				return "layout.field.name"
			}
			if !validForeignCType(field.CType) {
				return "layout.field.ctype"
			}
		}
	}
	return ""
}

// foreignExternName derives the linkage name a generated foreign call site
// declares `extern`. This deliberately does NOT go through cName/cLocal's
// ordinary-identifier allocator (cgen.go's namespace-closure invariant,
// TestGeneratedIdentifierNamespacesStayConfined) because it must match a
// REAL exported symbol in the byte-frozen foreign translation unit
// (native/schway_foreign_resource.c) verbatim, by convention, not by
// collision-avoidance allocation.
func foreignExternName(symbol string) string { return "_SCHWAY_" + symbol }

func emitEventSupport(out *strings.Builder, capacity int) {
	fmt.Fprintf(out, "#define SCHWAY_OUTPUT_LIMIT 65536u\n#define SCHWAY_EVENT_CAPACITY %du\n\n", capacity)
	out.WriteString("typedef struct SCHWAY_EVENT {\n")
	out.WriteString("  const char *kind;\n  const char *id;\n  const char *function_id;\n")
	out.WriteString("  const char *source_place;\n  const char *target_place;\n  const char *type_id;\n} SCHWAY_EVENT;\n\n")
	out.WriteString("static SCHWAY_EVENT schway_events[SCHWAY_EVENT_CAPACITY];\n")
	out.WriteString("static size_t schway_event_count = 0u;\nstatic size_t schway_output_count = 0u;\n\n")
	out.WriteString("static int schway_write_bytes(const char *data, size_t length) {\n")
	out.WriteString("  if (length > SCHWAY_OUTPUT_LIMIT - schway_output_count) return 0;\n")
	out.WriteString("  if (length != 0u && fwrite(data, 1u, length, stdout) != length) return 0;\n")
	out.WriteString("  schway_output_count += length;\n  return 1;\n}\n\n")
	out.WriteString("static int schway_write_literal(const char *value) {\n  return schway_write_bytes(value, strlen(value));\n}\n\n")
	out.WriteString("static int schway_write_json_string(const char *value) {\n")
	out.WriteString("  static const char hex[] = \"0123456789abcdef\";\n")
	out.WriteString("  if (!schway_write_bytes(\"\\\"\", 1u)) return 0;\n")
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
	out.WriteString("    if (escape != NULL) { if (!schway_write_literal(escape)) return 0; }\n")
	out.WriteString("    else if (byte < 0x20u) {\n")
	out.WriteString("      char encoded[6] = {'\\\\', 'u', '0', '0', hex[byte >> 4u], hex[byte & 0x0fu]};\n")
	out.WriteString("      if (!schway_write_bytes(encoded, sizeof encoded)) return 0;\n")
	out.WriteString("    } else if (!schway_write_bytes(value, 1u)) return 0;\n")
	out.WriteString("  }\n  return schway_write_bytes(\"\\\"\", 1u);\n}\n\n")
	out.WriteString("static int schway_record_event(const char *kind, const char *id, const char *function_id, const char *source_place, const char *target_place, const char *type_id) {\n")
	out.WriteString("  if (schway_event_count >= SCHWAY_EVENT_CAPACITY) return 0;\n")
	out.WriteString("  schway_events[schway_event_count++] = (SCHWAY_EVENT){kind, id, function_id, source_place, target_place, type_id};\n")
	out.WriteString("  return 1;\n}\n\n")
	out.WriteString("static int schway_write_events(void) {\n")
	out.WriteString("  size_t index;\n  for (index = 0u; index < schway_event_count; index++) {\n")
	out.WriteString("    const SCHWAY_EVENT *event = &schway_events[index];\n")
	out.WriteString("    if (index != 0u && !schway_write_bytes(\",\", 1u)) return 0;\n")
	out.WriteString("    if (!schway_write_literal(\"{\\\"schema\\\":\\\"lang.execution/1\\\",\\\"id\\\":\") || !schway_write_json_string(event->id)) return 0;\n")
	out.WriteString("    if (!schway_write_literal(\",\\\"kind\\\":\") || !schway_write_json_string(event->kind)) return 0;\n")
	out.WriteString("    if (!schway_write_literal(\",\\\"function_id\\\":\") || !schway_write_json_string(event->function_id)) return 0;\n")
	out.WriteString("    if (event->source_place != NULL && (!schway_write_literal(\",\\\"source_place\\\":\") || !schway_write_json_string(event->source_place))) return 0;\n")
	out.WriteString("    if (event->target_place != NULL && (!schway_write_literal(\",\\\"target_place\\\":\") || !schway_write_json_string(event->target_place))) return 0;\n")
	out.WriteString("    if (event->type_id != NULL && (!schway_write_literal(\",\\\"type_id\\\":\") || !schway_write_json_string(event->type_id))) return 0;\n")
	out.WriteString("    if (!schway_write_bytes(\"}\", 1u)) return 0;\n  }\n  return 1;\n}\n\n")
}

// emitEventSupportSchema2 is the multi-function-only /2 writer. It stays a
// sibling of the frozen /0-/1 support above so legacy generated bytes and
// their writer ABI cannot move as a side effect of native call evidence.
func emitEventSupportSchema2(out *strings.Builder, capacity, outputLimit int, needsJSONContent, applicationEvidence bool, evidenceLimit int) {
	fmt.Fprintf(out, "#define SCHWAY_OUTPUT_LIMIT %du\n#define SCHWAY_EVENT_CAPACITY %du\n\n", outputLimit, capacity)
	if applicationEvidence {
		fmt.Fprintf(out, "#define SCHWAY_EVIDENCE_OUTPUT_LIMIT %du\n\n", evidenceLimit)
	}
	out.WriteString("typedef struct SCHWAY_EVENT {\n")
	out.WriteString("  const char *kind;\n  const char *id;\n  const char *function_id;\n")
	out.WriteString("  const char *source_place;\n  const char *target_place;\n  const char *type_id;\n  const char *invocation;\n  const char *callee_function_id;\n} SCHWAY_EVENT;\n\n")
	out.WriteString("static SCHWAY_EVENT schway_events[SCHWAY_EVENT_CAPACITY];\n")
	out.WriteString("static size_t schway_event_count = 0u;\nstatic size_t schway_output_count = 0u;\n\n")
	if applicationEvidence {
		out.WriteString("static FILE *schway_output_stream = NULL;\nstatic size_t schway_output_limit = SCHWAY_OUTPUT_LIMIT;\nstatic int schway_event_overflow = 0;\n\n")
		out.WriteString("static int schway_write_bytes(const char *data, size_t length) {\n  FILE *stream = schway_output_stream != NULL ? schway_output_stream : stdout;\n  if (schway_output_count > schway_output_limit || length > schway_output_limit - schway_output_count) return 0;\n  if (length != 0u && fwrite(data, 1u, length, stream) != length) return 0;\n  schway_output_count += length;\n  return 1;\n}\n\n")
	} else {
		out.WriteString("static int schway_write_bytes(const char *data, size_t length) {\n  if (length > SCHWAY_OUTPUT_LIMIT - schway_output_count) return 0;\n  if (length != 0u && fwrite(data, 1u, length, stdout) != length) return 0;\n  schway_output_count += length;\n  return 1;\n}\n\n")
	}
	out.WriteString("static int schway_write_literal(const char *value) {\n  return schway_write_bytes(value, strlen(value));\n}\n\n")
	if needsJSONContent {
		out.WriteString("static int schway_write_json_string_content(const char *value) {\n")
		out.WriteString("  static const char hex[] = \"0123456789abcdef\";\n  for (; *value != '\\0'; value++) {\n    unsigned char byte = (unsigned char)*value;\n    const char *escape = NULL;\n")
		out.WriteString("    if (byte == '\"') escape = \"\\\\\\\"\";\n    else if (byte == '\\\\') escape = \"\\\\\\\\\";\n    else if (byte == '\\b') escape = \"\\\\b\";\n    else if (byte == '\\f') escape = \"\\\\f\";\n    else if (byte == '\\n') escape = \"\\\\n\";\n    else if (byte == '\\r') escape = \"\\\\r\";\n    else if (byte == '\\t') escape = \"\\\\t\";\n")
		out.WriteString("    if (escape != NULL) { if (!schway_write_literal(escape)) return 0; }\n    else if (byte < 0x20u) {\n      char encoded[6] = {'\\\\', 'u', '0', '0', hex[byte >> 4u], hex[byte & 0x0fu]};\n      if (!schway_write_bytes(encoded, sizeof encoded)) return 0;\n    } else if (!schway_write_bytes(value, 1u)) return 0;\n  }\n  return 1;\n}\n\n")
		out.WriteString("static int schway_write_json_string(const char *value) {\n  return schway_write_bytes(\"\\\"\", 1u) && schway_write_json_string_content(value) && schway_write_bytes(\"\\\"\", 1u);\n}\n\n")
	} else {
		out.WriteString("static int schway_write_json_string(const char *value) {\n")
		out.WriteString("  static const char hex[] = \"0123456789abcdef\";\n  if (!schway_write_bytes(\"\\\"\", 1u)) return 0;\n  for (; *value != '\\0'; value++) {\n    unsigned char byte = (unsigned char)*value;\n    const char *escape = NULL;\n")
		out.WriteString("    if (byte == '\"') escape = \"\\\\\\\"\";\n    else if (byte == '\\\\') escape = \"\\\\\\\\\";\n    else if (byte == '\\b') escape = \"\\\\b\";\n    else if (byte == '\\f') escape = \"\\\\f\";\n    else if (byte == '\\n') escape = \"\\\\n\";\n    else if (byte == '\\r') escape = \"\\\\r\";\n    else if (byte == '\\t') escape = \"\\\\t\";\n")
		out.WriteString("    if (escape != NULL) { if (!schway_write_literal(escape)) return 0; }\n    else if (byte < 0x20u) {\n      char encoded[6] = {'\\\\', 'u', '0', '0', hex[byte >> 4u], hex[byte & 0x0fu]};\n      if (!schway_write_bytes(encoded, sizeof encoded)) return 0;\n    } else if (!schway_write_bytes(value, 1u)) return 0;\n  }\n  return schway_write_bytes(\"\\\"\", 1u);\n}\n\n")
	}
	if applicationEvidence {
		out.WriteString("static int schway_record_event(const char *kind, const char *id, const char *function_id, const char *source_place, const char *target_place, const char *type_id, const char *invocation, const char *callee_function_id) {\n  if (schway_event_count >= SCHWAY_EVENT_CAPACITY) { schway_event_overflow = 1; return 1; }\n  schway_events[schway_event_count++] = (SCHWAY_EVENT){kind, id, function_id, source_place, target_place, type_id, invocation, callee_function_id};\n  return 1;\n}\n\n")
	} else {
		out.WriteString("static int schway_record_event(const char *kind, const char *id, const char *function_id, const char *source_place, const char *target_place, const char *type_id, const char *invocation, const char *callee_function_id) {\n  if (schway_event_count >= SCHWAY_EVENT_CAPACITY) return 0;\n  schway_events[schway_event_count++] = (SCHWAY_EVENT){kind, id, function_id, source_place, target_place, type_id, invocation, callee_function_id};\n  return 1;\n}\n\n")
	}
	out.WriteString("static int schway_write_events(void) {\n  size_t index;\n  for (index = 0u; index < schway_event_count; index++) {\n    const SCHWAY_EVENT *event = &schway_events[index];\n    if (index != 0u && !schway_write_bytes(\",\", 1u)) return 0;\n")
	out.WriteString("    if (!schway_write_literal(\"{\\\"schema\\\":\\\"lang.execution/2\\\",\\\"id\\\":\") || !schway_write_json_string(event->id)) return 0;\n    if (!schway_write_literal(\",\\\"kind\\\":\") || !schway_write_json_string(event->kind)) return 0;\n    if (!schway_write_literal(\",\\\"function_id\\\":\") || !schway_write_json_string(event->function_id)) return 0;\n")
	out.WriteString("    if (event->source_place != NULL && (!schway_write_literal(\",\\\"source_place\\\":\") || !schway_write_json_string(event->source_place))) return 0;\n    if (event->target_place != NULL && (!schway_write_literal(\",\\\"target_place\\\":\") || !schway_write_json_string(event->target_place))) return 0;\n    if (event->type_id != NULL && (!schway_write_literal(\",\\\"type_id\\\":\") || !schway_write_json_string(event->type_id))) return 0;\n")
	out.WriteString("    if (!schway_write_literal(\",\\\"invocation\\\":\") || !schway_write_json_string(event->invocation)) return 0;\n    if (event->callee_function_id != NULL && (!schway_write_literal(\",\\\"callee_function_id\\\":\") || !schway_write_json_string(event->callee_function_id))) return 0;\n    if (!schway_write_bytes(\"}\", 1u)) return 0;\n  }\n  return 1;\n}\n\n")
}

// functionHasDefect reports whether any operation in function's Linear body
// is an OpDefect -- the selection gate deciding whether emitBranch also
// emits the generated schway_defect support function (D-04-15). Every
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

// emitDefectSupport writes the generated _Noreturn schway_defect function
// (D-04-14/D-04-15): the one generated-code exemption from the zero-
// attribute control (D-04-13), because _Noreturn here is a property of a
// function cgen itself emits and every path of which provably ends in
// abort() -- not an unproven claim about a foreign callee. Called only from
// an emitter whose function actually contains an OpDefect operation, so
// every defect-free program's generated C is byte-for-byte unaffected
// (D-04-23).
func emitDefectSupport(out *strings.Builder) {
	out.WriteString("#include <stdlib.h>\n\n")
	out.WriteString("_Noreturn static void schway_defect(const char *reason) {\n")
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
// payloadCannedInitializer returns the fixed, compile-time-known C
// initializer this tracer uses for a given payload type's canonical value
// (matching linearInput's own Buffer/Byte canned literals). This phase's
// single-CLI-argument protocol carries only the selected alternative's own
// name, never real per-invocation payload bytes (D-12-41: a documented
// functionality gap, not an architectural one) -- see 12-02-SUMMARY.md.
func payloadCannedInitializer(payloadType string) string {
	switch payloadType {
	case "Buffer":
		// A compound literal (C99 6.5.2.5), not a bare brace-init list: this
		// initializer is used on the right-hand side of a plain assignment
		// (`field = <initializer>;`), where a brace-enclosed list alone is
		// a syntax error outside a declaration.
		return "(SCHWAY_BUFFER){{1u, 2u, 3u, 4u}, 4u}"
	case "Byte":
		return "7u"
	default:
		// A nullary ADT payload type (e.g. Fault): represented as its own
		// one-byte tag, canonically zero (its sole alternative).
		return "0u"
	}
}

// payloadFieldInfoT is one alternative's payload struct field, projected
// from check.PayloadRecordLayout's own LayoutField (D-12-25: the C type
// name and field name both come from that one shared derived fact, never a
// second, cgen-local derivation).
type payloadFieldInfoT struct{ name, cType string }

// payloadFieldInfoHasCType reports whether any declared payload field uses
// the given C type name -- used to gate emitting SCHWAY_BUFFER's own typedef
// exactly once, and only when a Buffer payload is actually declared.
func payloadFieldInfoHasCType(fields map[string]payloadFieldInfoT, cType string) bool {
	for _, field := range fields {
		if field.cType == cType {
			return true
		}
	}
	return false
}

// payloadFieldNames projects payloadFieldInfo down to the plain string map
// emitBranchOperations reads (alternative name -> struct field name only
// -- the C type is not needed past the struct's own declaration above).
func payloadFieldNames(fields map[string]payloadFieldInfoT) map[string]string {
	result := make(map[string]string, len(fields))
	for alternative, field := range fields {
		result[alternative] = field.name
	}
	return result
}

// payloadCTypeName returns the C type name a destructured payload local is
// declared with -- the same names check.PayloadRecordLayout's own
// payloadFieldShape assigns per payload type (D-12-25).
func payloadCTypeName(payloadType string) string {
	switch payloadType {
	case "Buffer":
		return "SCHWAY_BUFFER"
	default:
		return "unsigned char"
	}
}

// emitBranchOperations writes one arm block's straight-line C, in core
// order, ending with the lang.execution/1 JSON document for that arm's
// return. returnLiteral is the compile-time-known alternative name this
// block always returns (see emitBranch's doc comment).
func emitBranchOperations(out *strings.Builder, function core.Function, dataType core.DataType, alternativeBySource, payloadFieldBySource map[string]string, places map[string]core.Place, locals map[string]string, operationsByID map[string]core.LinearOperation, operationIDs []string, typeName, returnLiteral string) error {
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
			fmt.Fprintf(out, "      if (!schway_record_event(%s, %s, %s, %s, %s, %s)) return 74;\n",
				strconv.Quote(eventKind), strconv.Quote(operation.ID+":event"), strconv.Quote(function.ID),
				strconv.Quote(operation.SourceID), strconv.Quote(operation.TargetID), strconv.Quote(operation.TypeID))
		case core.OpReturn:
			fmt.Fprintf(out, "      if (!schway_record_event(%s, %s, %s, %s, NULL, %s)) return 74; /* returned place: %s */\n",
				strconv.Quote("function.returned"), strconv.Quote(operation.ID+":event:returned"), strconv.Quote(function.ID),
				strconv.Quote(operation.SourceID), strconv.Quote(operation.TypeID), operation.ID)
			out.WriteString("      if (!schway_write_literal(\"{\\\"schema\\\":\\\"lang.execution/1\\\",\\\"outcome\\\":{\\\"kind\\\":\\\"returned\\\",\\\"value\\\":\")) return 74;\n")
			fmt.Fprintf(out, "      if (!schway_write_json_string(%s)) return 74;\n", strconv.Quote(returnLiteral))
			out.WriteString("      if (!schway_write_literal(\"},\\\"events\\\":[\")) return 74;\n")
			out.WriteString("      if (!schway_write_events()) return 74;\n")
			out.WriteString("      if (!schway_write_literal(\"],\\\"live_resources\\\":[]}\\n\")) return 74;\n")
			out.WriteString("      return 0;\n")
		case core.OpDefect:
			// D-04-15: lower to a call to the generated _Noreturn schway_defect
			// function, every path of which ends in abort(). The terminal
			// record (outcome kind "defect", no value, its own
			// function.defected event carrying the required reason string in
			// "output", empty live_resources -- no arm can carry a foreign
			// acquisition this phase) is written FIRST, so it is captured
			// even though the process then aborts; schway_defect's own
			// _Noreturn marker is the one and only exemption from the
			// zero-attribute control (D-04-13/D-04-14), because it is a
			// property of a function cgen itself emits, not an unproven
			// claim about a foreign callee.
			//
			// This event is assembled directly rather than through
			// schway_record_event/schway_write_events: the shared SCHWAY_EVENT
			// struct those helpers use (emitEventSupport, frozen for every
			// other emitter, D-04-23) has no "output" field, and adding one
			// there would move every existing committed generated-C golden.
			out.WriteString("      if (!schway_write_literal(\"{\\\"schema\\\":\\\"lang.execution/1\\\",\\\"outcome\\\":{\\\"kind\\\":\\\"defect\\\",\\\"value\\\":\\\"\\\"},\\\"events\\\":[\")) return 74;\n")
			out.WriteString("      if (!schway_write_events()) return 74;\n")
			out.WriteString("      if (schway_event_count != 0u && !schway_write_bytes(\",\", 1u)) return 74;\n")
			out.WriteString("      if (!schway_write_literal(\"{\\\"schema\\\":\\\"lang.execution/1\\\",\\\"id\\\":\")) return 74;\n")
			fmt.Fprintf(out, "      if (!schway_write_json_string(%s)) return 74;\n", strconv.Quote(operation.ID+":event:defected"))
			out.WriteString("      if (!schway_write_literal(\",\\\"kind\\\":\\\"function.defected\\\",\\\"function_id\\\":\")) return 74;\n")
			fmt.Fprintf(out, "      if (!schway_write_json_string(%s)) return 74;\n", strconv.Quote(function.ID))
			out.WriteString("      if (!schway_write_literal(\",\\\"source_place\\\":\")) return 74;\n")
			fmt.Fprintf(out, "      if (!schway_write_json_string(%s)) return 74;\n", strconv.Quote(operation.SourceID))
			out.WriteString("      if (!schway_write_literal(\",\\\"type_id\\\":\")) return 74;\n")
			fmt.Fprintf(out, "      if (!schway_write_json_string(%s)) return 74;\n", strconv.Quote(operation.TypeID))
			out.WriteString("      if (!schway_write_literal(\",\\\"output\\\":\")) return 74;\n")
			fmt.Fprintf(out, "      if (!schway_write_json_string(%s)) return 74;\n", strconv.Quote(operation.Reason))
			out.WriteString("      if (!schway_write_literal(\"}\")) return 74;\n")
			out.WriteString("      if (!schway_write_literal(\"],\\\"live_resources\\\":[]}\\n\")) return 74;\n")
			fmt.Fprintf(out, "      schway_defect(%s);\n", strconv.Quote(operation.Reason))
			out.WriteString("      return 71; /* unreachable: schway_defect never returns */\n")
		case core.OpForeignCall, core.OpFail:
			// checkBranch never emits either kind inside a match arm body
			// this phase (no `try` support inside an arm) -- named here,
			// rather than falling into default:, so control:kind.exhaustive
			// _dispatch's six-site table finds a known-but-unsupported case
			// rather than an unknown one (D-04-22).
			return fmt.Errorf("operation %q: foreign calls inside a match arm body are not supported this phase", operation.ID)
		case core.OpCall:
			// D-07-39/A-02/D-11-04: emitBranch (and this helper) is reached
			// only for a single-function branch-shaped program, which can
			// never legally contain an OpCall for the same reason
			// emitLinear's own OpCall arm documents -- provably
			// unreachable forward hygiene, wired through the same emitCall
			// helper Phase 11's multi-function assembler uses on its real
			// path (D-11-04: exactly one writer of a Lang-to-Lang call).
			target, exists := places[operation.TargetID]
			if !exists {
				return fmt.Errorf("operation %q has invalid target", operation.ID)
			}
			calleeName, calleeTypeName, ok := (*emitCallLookup)(nil).resolve(operation.CalleeID)
			if !ok {
				return fmt.Errorf("operation %q: Lang-to-Lang calls are not supported by native emission this phase (unresolved callee %q)", operation.ID, operation.CalleeID)
			}
			emitCall(out, calleeTypeName, locals[target.ID], calleeName, locals[source.ID], operation)
		case core.OpDestructurePayload:
			// D-12-05/D-12-14: reads the payload out of the SOURCE's own
			// struct field -- the field name comes from payloadFieldBySource
			// (D-12-25's shared check.PayloadRecordLayout, never a
			// cgen-local re-derivation), resolved from the operation's own
			// PayloadType fact back to an alternative name via the ONE
			// shared, ambiguity-detecting derivation core.
			// AlternativeNameForPayloadType (IN-01/CR-01, D-12-25).
			target, exists := places[operation.PayloadTargetID]
			if !exists {
				return fmt.Errorf("operation %q has invalid target", operation.ID)
			}
			altName, altErr := core.AlternativeNameForPayloadType(dataType, operation.PayloadType)
			if altErr != nil {
				return fmt.Errorf("operation %q: %w", operation.ID, altErr)
			}
			field, knownField := payloadFieldBySource[altName]
			if !knownField {
				return fmt.Errorf("operation %q: no struct field for alternative %q", operation.ID, altName)
			}
			cType := payloadCTypeName(operation.PayloadType)
			fmt.Fprintf(out, "      %s %s = %s.%s; /* payload destructure: %s */\n", cType, locals[target.ID], locals[source.ID], field, operation.ID)
			fmt.Fprintf(out, "      (void)%s;\n", locals[target.ID])
			fmt.Fprintf(out, "      if (!schway_record_event(%s, %s, %s, %s, %s, %s)) return 74;\n",
				strconv.Quote("value.payload_destructured"), strconv.Quote(operation.ID+":event"), strconv.Quote(function.ID),
				strconv.Quote(operation.SourceID), strconv.Quote(operation.PayloadTargetID), strconv.Quote(operation.TypeID))
		case core.OpConstructPayload:
			// D-12-05/D-12-14: builds a NEW tagged struct value, setting the
			// tag to the constructed alternative's own `#define`d constant
			// and writing the source payload into that alternative's own
			// struct field -- both facts read from the SAME shared inputs
			// as the destructure case above, never re-derived independently.
			target, exists := places[operation.TargetID]
			if !exists {
				return fmt.Errorf("operation %q has invalid target", operation.ID)
			}
			altName, altErr := core.AlternativeNameForPayloadType(dataType, operation.PayloadType)
			if altErr != nil {
				return fmt.Errorf("operation %q: %w", operation.ID, altErr)
			}
			field, knownField := payloadFieldBySource[altName]
			if !knownField {
				return fmt.Errorf("operation %q: no struct field for alternative %q", operation.ID, altName)
			}
			tagConstant, knownTag := alternativeBySource[altName]
			if !knownTag {
				return fmt.Errorf("operation %q: no tag constant for alternative %q", operation.ID, altName)
			}
			fmt.Fprintf(out, "      %s %s; /* payload construct: %s */\n", typeName, locals[target.ID], operation.ID)
			fmt.Fprintf(out, "      %s.tag = %s;\n", locals[target.ID], tagConstant)
			if payloadSlotSwapForTest {
				// D-12-38 Task 2's fault-injection seam: write the source
				// payload into a DIFFERENT alternative's struct field than
				// the one the (correct) tag names -- reads or writes the
				// WRONG alternative's payload slot for a correct tag. The
				// assignment stays type-safe (truncating/widening through a
				// scalar or a compound literal) rather than an out-of-bounds
				// memcpy, so any observed engine disagreement is attributable
				// to the wrong SLOT alone, never to separately-introduced
				// undefined behavior.
				if wrongField, wrongCType, ok := wrongPayloadSlot(dataType, payloadFieldBySource, altName); ok {
					payloadSlotSwapInjectedWriteCount++
					srcCType := payloadCTypeName(operation.PayloadType)
					switch {
					case wrongCType == srcCType:
						fmt.Fprintf(out, "      %s.%s = %s; /* D-12-38 mutation: wrong-slot write */\n", locals[target.ID], wrongField, locals[source.ID])
					case wrongCType == "SCHWAY_BUFFER":
						fmt.Fprintf(out, "      %s.%s = (SCHWAY_BUFFER){{%s}, 1u}; /* D-12-38 mutation: wrong-slot write, widened */\n", locals[target.ID], wrongField, locals[source.ID])
					default:
						fmt.Fprintf(out, "      %s.%s = %s.bytes[0]; /* D-12-38 mutation: wrong-slot write, truncated */\n", locals[target.ID], wrongField, locals[source.ID])
					}
				} else {
					fmt.Fprintf(out, "      %s.%s = %s;\n", locals[target.ID], field, locals[source.ID])
				}
			} else {
				fmt.Fprintf(out, "      %s.%s = %s;\n", locals[target.ID], field, locals[source.ID])
			}
			fmt.Fprintf(out, "      (void)%s;\n", locals[target.ID])
			fmt.Fprintf(out, "      if (!schway_record_event(%s, %s, %s, %s, %s, %s)) return 74;\n",
				strconv.Quote("value.payload_constructed"), strconv.Quote(operation.ID+":event"), strconv.Quote(function.ID),
				strconv.Quote(operation.SourceID), strconv.Quote(operation.TargetID), strconv.Quote(operation.TypeID))
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
// (D-04-12c): a separate schema/artifact from schway.core/*, digest-bound into
// evidence.Manifest.ForeignDigest but never merged into the core artifact
// itself.
const ForeignManifestSchema = "lang.foreign/0"

// foreignManifestDocument is the lang.foreign/0 sidecar's exact field
// layout: the complete core.ForeignContract plus two fields no
// ForeignContract itself carries. EmittedAttributes has no omitempty tag
// (D-04-13): it is deliberately a present, empty JSON array this phase when
// nothing was emitted, not an omission. As of D-05-01/D-05-04 it can also be
// a present, POPULATED array: an empty []EmittedAttribute and an empty
// []string both serialize to the same `[]`, so every Phase 1-4 sidecar's
// serialized bytes stay unchanged (TestForeignManifestBytesUnchangedForPriorPhases
// asserts this directly, not merely assumes it).
// UncheckedObligations names every obligation this phase declares but never
// exercises, so a quarantine reader never mistakes a declared fact for a
// proven one (D-04-12/D-10).
type foreignManifestDocument struct {
	Schema                   string                    `json:"schema"`
	Symbol                   string                    `json:"symbol"`
	Allocator                string                    `json:"allocator"`
	Unwind                   string                    `json:"unwind"`
	NonlocalExit             string                    `json:"nonlocal_exit"`
	Fails                    string                    `json:"fails"`
	InitializedState         string                    `json:"initialized_state"`
	Capture                  string                    `json:"capture"`
	Retention                string                    `json:"retention"`
	Aliasing                 string                    `json:"aliasing"`
	Layout                   *core.RecordLayout        `json:"layout"`
	EmittedAttributes        []EmittedAttribute        `json:"emitted_attributes"`
	EmittedPointerParameters []EmittedPointerParameter `json:"emitted_pointer_parameters,omitempty"`
	UncheckedObligations     []string                  `json:"unchecked_obligations"`
}

// EmittedAttribute is one lang.foreign/0 sidecar emitted_attributes entry
// (D-05-04): the optimizer-visible attribute cgen emitted, which core node
// and parameter it was emitted on, and the borrow-fact identity that
// justifies it. corevalidate independently re-derives JustifiedBy from
// core's own loan facts and refuses any entry whose claim does not match
// (ValidateEmittedAttributes, corevalidate.go) -- it declares its own local
// AttributeClaim decoding of this exact JSON shape rather than importing
// this type, so the two packages share no helper (D-12).
type EmittedAttribute struct {
	Attr        string `json:"attr"`
	CoreNode    string `json:"core_node"`
	Parameter   string `json:"parameter"`
	JustifiedBy string `json:"justified_by"`
}

type EmittedPointerParameter struct {
	CoreNode  string `json:"core_node"`
	Parameter string `json:"parameter"`
	CType     string `json:"c_type"`
	Access    string `json:"access"`
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
		// testdata/phase4/nonlocal_exit_probe.schway's REACHABILITY comment.
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
			// 04-13 (04-VERIFICATION.md gap 2b, FFI-01): the SAME
			// independent-refusal reasoning as the Symbol guard above, now
			// extended to every OTHER core.ForeignContract string field
			// EmitForeignHeader's and EmitForeignConformance's splice sites
			// reach. The field NAME unsafeForeignContractField returns is a
			// fixed vocabulary of literals, so no attacker-controlled byte
			// reaches the error string at all -- a stronger guarantee than
			// %q escaping, and the right one here because the payload's
			// whole point is a comment terminator that %q does not
			// neutralize.
			if offendingField := unsafeForeignContractField(function.ForeignContract); offendingField != "" {
				return core.Function{}, fmt.Errorf("foreign contract field %s is not safe to splice into generated C", offendingField)
			}
			return function, nil
		}
	}
	return core.Function{}, fmt.Errorf("program declares no foreign contract")
}

// singleManifestFunction returns the one function in program that
// EmitForeignManifest has a lang.foreign/0 sidecar to say something about:
// either a declared foreign contract (singleForeignFunction's existing
// Phase 4 scope, delegated to verbatim so its validation/error behavior is
// completely unchanged for every foreign-shaped program) or, when no
// function declares a foreign contract at all, the sole function selecting
// D-05-02's by-pointer lowering. EmitForeignHeader and EmitForeignConformance
// stay scoped to singleForeignFunction directly -- a by-pointer function has
// no separate header or conformance unit to emit, so generalizing THEIR
// selector would be reaching for a case that does not exist.
func singleManifestFunction(program core.Program) (core.Function, error) {
	for _, function := range program.Functions {
		if function.ForeignContract != nil {
			return singleForeignFunction(program)
		}
	}
	for _, function := range program.Functions {
		if function.Linear != nil && selectsByPointerLowering(function, function.Linear) {
			return function, nil
		}
		if _, ok := checkedSharedPointerABIFact(function); ok {
			return function, nil
		}
	}
	return core.Function{}, fmt.Errorf("program declares no foreign contract")
}

// EmitForeignManifest serializes program's single foreign contract (or, as
// of D-05-01, its single by-pointer-lowered function) as a lang.foreign/0
// sidecar manifest document (D-04-12c): the JSON is authoritative, and
// EmitForeignHeader's obligation comment block is generated FROM the same
// contract value, so the two can never drift (D-04-12).
func EmitForeignManifest(program core.Program) (string, error) {
	function, err := singleManifestFunction(program)
	if err != nil {
		return "", err
	}
	document := foreignManifestDocument{
		Schema: ForeignManifestSchema, EmittedAttributes: []EmittedAttribute{}, UncheckedObligations: uncheckedForeignObligations(),
	}
	if contract := function.ForeignContract; contract != nil {
		document.Symbol, document.Allocator = contract.Symbol, contract.Allocator
		document.Unwind, document.NonlocalExit, document.Fails = contract.Unwind, contract.NonlocalExit, contract.Fails
		document.InitializedState, document.Capture, document.Retention, document.Aliasing = contract.InitializedState, contract.Capture, contract.Retention, contract.Aliasing
		document.Layout = contract.Layout
	} else {
		if fact, ok := checkedSharedPointerABIFact(function); ok {
			document.EmittedPointerParameters = append(document.EmittedPointerParameters, EmittedPointerParameter{
				CoreNode: fact.FunctionID, Parameter: fact.ParameterID, CType: fact.ParameterCType, Access: fact.Access,
			})
		} else {
			document.EmittedAttributes = append(document.EmittedAttributes, emittedAttributeForByPointerParameter(function))
		}
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

// EmitForeignHeader generates the `_SCHWAY_`-namespaced header for program's
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
	guard := "SCHWAY_FOREIGN_" + strings.ToUpper(cName(contract.Symbol)) + "_H"

	var out strings.Builder
	out.WriteString("/* generated by Schway; schema lang.c17/0 (foreign header, D-04-12a) */\n")
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

// EmitForeignConformance generates schway_foreign_conformance.c (D-04-11): the
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
	out.WriteString("/* generated by Schway; schema lang.c17/0 (conformance TU, D-04-11).\n")
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

// EmitPayloadConformance is D-12-37's payload-side sibling of
// EmitForeignConformance: given a payload-carrying core.DataType, it emits a
// conformance translation unit asserting (via _Static_assert/offsetof) that
// a privately-declared C struct at privateHeaderPath matches
// check.PayloadRecordLayout(dataType) -- the ONE shared checker-derived
// layout fact (D-12-25), never a second independent derivation. It defines
// no symbol and is compiled but never linked, following
// EmitForeignConformance's own shape exactly. This control is NECESSARY for
// the C-side layout obligation and STRUCTURALLY INCAPABLE of catching a
// wrong-slot read for a correct tag: a _Static_assert polices only the
// struct DECLARATION's sizeof/_Alignof/offsetof, never which field a given
// match arm actually reads (D-12-38 is the control that covers that gap).
func EmitPayloadConformance(dataType core.DataType, privateHeaderPath string) (string, error) {
	layout := check.PayloadRecordLayout(dataType)
	if layout.ForeignTypeName == "" {
		return "", fmt.Errorf("payload record layout for %q has no foreign type name to prove", dataType.Name)
	}

	var out strings.Builder
	out.WriteString("/* generated by Schway; schema lang.c17/0 (payload conformance TU, D-12-37).\n")
	out.WriteString(" * This is the single explicit, auditable place the checker-derived payload\n")
	out.WriteString(" * struct layout (check.PayloadRecordLayout) and a private struct declaration\n")
	out.WriteString(" * are permitted to meet. It defines no symbol and is compiled but never\n")
	out.WriteString(" * linked. */\n")
	if payloadRecordHasCType(layout, "SCHWAY_BUFFER") {
		out.WriteString("typedef struct SCHWAY_BUFFER {\n  unsigned char bytes[4];\n  size_t length;\n} SCHWAY_BUFFER;\n\n")
	}
	out.WriteString("#include <stddef.h>\n")
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

// payloadRecordHasCType reports whether any field of layout uses the given C
// type name -- used to gate emitting SCHWAY_BUFFER's own typedef exactly once,
// and only when a Buffer payload field is actually declared, mirroring
// payloadFieldInfoHasCType's own emitBranch-side gate.
func payloadRecordHasCType(layout *core.RecordLayout, cType string) bool {
	for _, field := range layout.Fields {
		if field.CType == cType {
			return true
		}
	}
	return false
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

// JustifiableAttributes is D-05-03's second and only other exemption from
// the zero-attribute control (NoreturnExemption above is the first). Unlike
// NoreturnExemption -- a property of a function cgen itself emits, requiring
// no justification at all -- a JustifiableAttributes token is legal ONLY
// inside a lang.foreign/0 manifest's emitted_attributes entry, ONLY when
// that entry carries a non-empty JustifiedBy binding corevalidate
// independently re-derives (ScanForUnjustifiedAttributes below,
// ValidateEmittedAttributes in corevalidate.go), and NEVER inside a foreign
// extern declaration, where ScanForBannedAttributes' existing
// declaration-site ban stays exactly as strict as Phase 4
// (BannedOptimizerAttributes is unchanged, still containing "restrict").
// The exemption's exact boundary: cgen-owned by-pointer parameters only,
// never a foreign extern declaration cgen does not own.
var JustifiableAttributes = []string{"restrict"}

// ScanForUnjustifiedAttributes returns every entry in attributes whose Attr
// is not a member of JustifiableAttributes, or whose JustifiedBy is empty.
// A non-empty result is a hard build failure (D-05-03b), never a warning --
// an attribute with no proven backing fact must never ship.
func ScanForUnjustifiedAttributes(attributes []EmittedAttribute) []string {
	justifiable := make(map[string]bool, len(JustifiableAttributes))
	for _, token := range JustifiableAttributes {
		justifiable[token] = true
	}
	var found []string
	for _, attribute := range attributes {
		if !justifiable[attribute.Attr] || attribute.JustifiedBy == "" {
			found = append(found, attribute.Attr)
		}
	}
	return found
}

func linearInput(function core.Function) (input, initializer, typeName string, err error) {
	switch function.Parameter.Type {
	case "Buffer":
		return "01020304", "{{1u, 2u, 3u, 4u}, 4u}", "SCHWAY_BUFFER", nil
	case "Byte":
		return "7", "7u", "unsigned char", nil
	case "U64":
		return "0", "UINT64_C(0)", "uint64_t", nil
	case "PathToken":
		return "", "NULL", "const char *", nil
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
// built as "__SCHWAY_" would therefore hand out reserved names on every actual
// collision, in generated code a conforming implementation is permitted to
// treat specially. "_SCHWAY_" (a single leading underscore) has no double
// underscore and stays inside the confined "SCHWAY_"/"schway_value_" namespaces
// documented above, so it is never reserved and never re-issues a reserved
// name.
func (n *cNames) allocate(preferred, category string, ordinal int) string {
	if _, exists := n.used[preferred]; !exists {
		n.used[preferred] = struct{}{}
		return preferred
	}
	base := preferred + "_SCHWAY_" + strings.ToUpper(category) + "_" + strconv.Itoa(ordinal)
	for attempt, candidate := 0, base; ; attempt, candidate = attempt+1, base+"_"+strconv.Itoa(attempt) {
		if _, exists := n.used[candidate]; !exists {
			n.used[candidate] = struct{}{}
			return candidate
		}
	}
}

// cName maps a source type, alternative, function, or parameter name into the
// uppercase "SCHWAY_" namespace. INVARIANT (enforced by
// TestGeneratedIdentifierNamespacesStayConfined): the result always matches
// ^SCHWAY_[A-Z0-9_]*$. Do not relax the uppercasing or the prefix — the closure
// argument for the generated-C namespace depends on every allocated identifier
// living in the "SCHWAY_" or "schway_value_" namespace and nowhere else.
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
	return "SCHWAY_" + out.String()
}

// cLocal maps a source place name into the "schway_value_" namespace. INVARIANT
// (enforced by TestGeneratedIdentifierNamespacesStayConfined): the result
// always matches ^schway_value_[A-Za-z0-9_]*$. No fixed identifier emitted by any
// emitter may be placed in this namespace unless it is also reserved.
func cLocal(name string) string {
	var out strings.Builder
	out.WriteString("schway_value_")
	for _, r := range name {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' {
			out.WriteRune(r)
		} else {
			out.WriteByte('_')
		}
	}
	return out.String()
}
