package core

import (
	"encoding/json"

	"github.com/codename-lang/lang/internal/compiler/diagnostic"
)

const (
	Schema  = "lang.core/0"
	Schema1 = "lang.core/1"
)

type Program struct {
	Schema    string     `json:"schema"`
	Module    string     `json:"module"`
	ModuleID  string     `json:"module_id"`
	DataTypes []DataType `json:"data_types"`
	Functions []Function `json:"functions"`
}

type DataType struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Alternatives []string        `json:"alternatives"`
	Span         diagnostic.Span `json:"span"`
}

type Function struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	EntryPointID  string      `json:"entry_point_id"`
	ReturnPointID string      `json:"return_point_id"`
	Parameter     Parameter   `json:"parameter"`
	ReturnType    string      `json:"return_type"`
	Match         *Match      `json:"match,omitempty"`
	Linear        *LinearBody `json:"linear,omitempty"`
	// PublicOrigin is the Phase 3 OWN-04 fact (additive, omitempty): present
	// only when the function's return type carried a `borrow(path)` /
	// `borrow mut(path)` annotation. It is a sibling of Match/Linear, not a
	// field on TypeFact, because origin is a per-signature fact about which
	// parameter a result derives from, not a property of the returned type
	// itself (03-RESEARCH Q5).
	PublicOrigin *PublicOrigin `json:"public_origin,omitempty"`
	// ForeignContract is the Phase 4 D-04-12 fact (additive, omitempty):
	// present only when this function calls a declared `foreign C {}` symbol.
	// Every pre-Phase-4 function, and every Phase 4 function that calls no
	// foreign symbol, leaves this nil, so its serialized bytes are unchanged
	// (D-04-23).
	ForeignContract *ForeignContract `json:"foreign_contract,omitempty"`
	Span            diagnostic.Span  `json:"span"`
}

// ForeignContract is the complete obligation set FFI-01 names for one
// declared foreign C symbol a function calls: its name, allocator identity,
// the two mandatory admission policies (D-04-16), the name of the nullary
// ADT its err edge carries (D-04-05), and (from Phase 4 plan 03, D-04-12) the
// target record Layout plus the InitializedState/Capture/Retention/Aliasing
// obligation categories. Every field is either required (refused when
// omitted, never defaulted) or, for the four obligations this phase's
// language cannot yet exercise, a fixed structural fact the compiler itself
// derives -- see check.go's standardForeignObligations for why that is not
// an omission-tolerant default.
type ForeignContract struct {
	Symbol       string `json:"symbol"`
	Allocator    string `json:"allocator"`
	Unwind       string `json:"unwind,omitempty"`
	NonlocalExit string `json:"nonlocal_exit,omitempty"`
	Fails        string `json:"fails,omitempty"`
	// InitializedState, Capture, Retention, and Aliasing are Phase 4 plan 03's
	// remaining FFI-01 obligation categories (D-04-12). This phase's language
	// has no partial-field initialization, no closures, no threads, and no
	// calls into Lang, so all four are structurally fixed facts the compiler
	// derives identically for every declared symbol (check.go's
	// standardForeignObligations) rather than per-symbol declarations.
	// Capture/Retention/Aliasing are additionally named in the lang.foreign/0
	// sidecar manifest's unchecked_obligations list (D-04-12c), since nothing
	// in this phase exercises them.
	InitializedState string `json:"initialized_state,omitempty"`
	Capture          string `json:"capture,omitempty"`
	Retention        string `json:"retention,omitempty"`
	Aliasing         string `json:"aliasing,omitempty"`
	// Layout is the target record layout obligation (D-04-12/T-04-14): an
	// ordered field list with declared size, alignment, and offset, plus the
	// record's own declared size and alignment. This phase's foreign surface
	// has exactly one declared record shape (the {ok, value} two-field
	// by-value ABI result cgen's emitLinearForeign always generates), so
	// every foreign symbol's Layout is currently checker-derived -- a fixed
	// structural fact, not a per-symbol declaration (see check.go's
	// standardForeignLayout).
	Layout *RecordLayout `json:"layout,omitempty"`
	// Alias is Phase 4 plan 06's additive omitempty fact (D-04-28/FFI-01): a
	// declared foreign symbol's own aliasing obligation toward the return
	// value it produces on its ok edge. "" (the default) means the returned
	// value is fully owned -- no pre-Phase-4-plan-06 symbol declares this key,
	// so every existing fixture's serialized bytes are unchanged (D-04-23).
	// "borrow" means the ok-edge return is a shared-access alias of the
	// call's own argument; "retain" means an exclusive-access alias -- the
	// same two access modes core.PublicOrigin.Access already declares,
	// reused rather than inventing a third vocabulary. originvalidate reads
	// this field alone (never the checker) to recognise a foreign-call-
	// derived return as borrow-derived (D-04-28).
	Alias string `json:"alias,omitempty"`
}

// RecordLayout is one declared record's layout obligation: its own declared
// size and alignment, plus an ordered field list. LayoutField carries a
// field's name and its declared size, alignment, and offset within the
// record. Both are additive (Phase 4 plan 03), referenced only from
// ForeignContract.Layout.
type RecordLayout struct {
	Size      int           `json:"size"`
	Alignment int           `json:"alignment"`
	Fields    []LayoutField `json:"fields"`
	// ForeignTypeName is the C struct name Lang's declaration expects to find
	// on the foreign side (e.g. in the frozen translation unit's private
	// header): the generated conformance TU asserts sizeof/_Alignof/offsetof
	// pairs between this name and Lang's own generated declaration.
	ForeignTypeName string `json:"foreign_type_name"`
}

type LayoutField struct {
	Name      string `json:"name"`
	Size      int    `json:"size"`
	Alignment int    `json:"alignment"`
	Offset    int    `json:"offset"`
	// CType is the declared C type of this field (this phase's foreign
	// surface only ever declares single-byte fields, so this is always
	// "unsigned char" today; carried explicitly rather than hardcoded at the
	// emission site so the field is self-describing).
	CType string `json:"c_type"`
}

// PublicOrigin records the declared origin path(s) and access mode for a
// function whose return type is a borrowed view. Paths is one or more
// `parameter[.field]` strings; this phase's executable shapes (Byte, Buffer)
// carry no fields and every function has exactly one parameter, so the only
// legal path is the parameter's own name — the slice shape is kept plural so
// a future field-path or per-alternative extension is additive, not a schema
// change. Access is "shared" or "exclusive", carried independently of Paths.
type PublicOrigin struct {
	Paths  []string `json:"paths"`
	Access string   `json:"access"`
}

// Interface is the Phase 3 OWN-04 body-stripped separate-compilation
// artifact (03-RESEARCH Q6): a subset of Program containing only module
// identity, a digest binding it to the exact core.Program it was derived
// from, and per-function signatures — explicitly no Linear or Match body.
// A consuming process decodes only this shape and can never reconstruct a
// provider body from it.
type Interface struct {
	Schema     string              `json:"schema"`
	ModuleID   string              `json:"module_id"`
	CoreDigest string              `json:"core_digest"`
	Functions  []FunctionSignature `json:"functions"`
}

// InterfaceSchema versions the Interface artifact independently of the core
// schema it summarizes: adding a field here never moves a core.Program byte.
// Frozen (D-07-08): a lang.interface/0 document is decodable only by the
// pinned legacy struct InterfaceV0/FunctionSignatureV0, and is never
// admissible for a call.
const InterfaceSchema = "lang.interface/0"

// InterfaceSchema1 mints lang.interface/1 (D-07-08): the interprocedural
// call contract. Every required field on FunctionSignature/ParameterContract/
// ReturnContract below carries no omitempty and is refused when missing or
// empty by DecodeInterface (07-01 Task 2) -- absence is an error, not a
// defaulting opportunity (D-07-09). A /0 document stays decodable but is
// never admissible for a call; this bump moves no core.Program byte, which
// is exactly the property InterfaceSchema's independent versioning was
// created to buy.
const InterfaceSchema1 = "lang.interface/1"

// FunctionSignature is one function's body-stripped public surface under
// lang.interface/1 (D-07-08/D-07-09): identity, the callee's per-parameter
// ownership contract, its total return contract, its granted abilities,
// whether it may be called at all, its failure vocabulary, its worst-case
// foreign reach, and a canonical digest binding it to its own transitive
// closure. It deliberately has no Linear/Match field at all — not merely an
// omitted one — so a consumer decoding this type structurally cannot reach a
// body even by accident (the same argument style as /0).
//
// Every field is REQUIRED unless its doc comment says a legal empty value
// exists: DecodeInterface (07-01 Task 2) refuses a /1 document missing or
// emptying any required field. No field here may be populated by a caller
// admission path — SEM-05 forbids that path from reading a callee body, but
// does not forbid the PRODUCER (BuildInterface) from doing so; each field's
// R-01 authority names exactly where its value comes from.
type FunctionSignature struct {
	// ID and Name are copied verbatim from core.Function (R-01).
	ID   string `json:"id"`
	Name string `json:"name"`
	// Parameters is a slice per D-07-10 even though the checker admits
	// exactly arity 1 today (D-07-01): schema capacity is taken now, with
	// one producer, rather than later across five consumers. The arity
	// rule stays tight in the checker, where the cost lives.
	Parameters []ParameterContract `json:"parameters"`
	// Return is this function's total return contract (R-01: subsumes the
	// old optional PublicOrigin — the same two access modes, now total
	// rather than optional).
	Return ReturnContract `json:"return"`
	// Abilities is read from the existing function.ID+":type:0" type-fact
	// lookup (R-01, originvalidate.go BuildInterface) — unchanged from /0.
	Abilities []Ability `json:"abilities"`
	// Callable reports whether this function may appear as a call target
	// (D-04-03: callable subseteq publishable). This plan declares the
	// field and leaves it at its fail-closed zero value false; its
	// predicate is defined in 07-02 (D-07-31/D-07-32) — an unpopulated
	// field admits nothing.
	Callable bool `json:"callable"`
	// Fails reuses core.ForeignContract.Fails' vocabulary verbatim (R-01):
	// "" (the legal empty value) means infallible; no new failure
	// vocabulary is minted here.
	Fails string `json:"fails,omitempty"`
	// Foreign is this function's closure-derived worst-case foreign reach
	// (R-01), plain strings per the ForeignContract precedent (core.go:60-64,
	// planner discretion exercised: plain strings, not a new enumerated
	// type).
	Foreign ForeignReach `json:"foreign"`
	// ClosureDigest is D-07-37's canonical, non-self-referential digest
	// over this signature (with ClosureDigest itself zeroed) and its
	// callees' own ClosureDigests, sorted by ID — a real Merkle chain over
	// callee SUMMARY digests (D-07-12), never a callee body. 07-01 computed
	// only the zero-callee base case; 07-08 landed the chaining arm over
	// real callees, in callgraph.Order's reverse postorder, only after
	// cycle refusal exists (D-07-38: the chain terminates only on a DAG).
	// Changing a callee changes its caller's ClosureDigest — this is the
	// only thing that closes the stale-but-self-consistent hole a per-unit
	// hash leaves (an unchanged caller silently serving a stale verdict
	// when its callee changes).
	//
	// D-07-13: chaining WIDENS what staleness ClosureDigest detects; it
	// adds NO authenticity whatsoever. It remains an unkeyed content hash —
	// anyone who can write the summary can write a consistent chained
	// digest. It detects staleness, never forgery; the forgery answer is
	// independent re-derivation (07-02's summary peer, 07-08's own chained
	// re-derivation of it), never this digest.
	ClosureDigest string `json:"closure_digest"`
}

// ParameterContract is one parameter's ownership contract under
// lang.interface/1 (D-07-09). Every field is required.
type ParameterContract struct {
	// ID, Name, and Type are copied verbatim from core.Function.Parameter
	// (R-01).
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	// Mode is the declared parameter form (R-01): "owned", "shared", or
	// "exclusive" — no legal "". Today's grammar has exactly one parameter
	// form (by-value), so every function's Mode is "owned"; a future
	// borrow/borrow mut parameter form maps to shared/exclusive. Any other
	// value is unreachable and is refused by DecodeInterface.
	Mode string `json:"mode"`
	// Drops is producer-derived from the checked core (R-01): whether the
	// callee discharges the drop obligation on this owned parameter
	// (consumes/drops it internally) rather than moving it out to the
	// caller via an owned return. Producer-side body reading is legal here
	// — BuildInterface exists precisely to strip a body into a summary; no
	// caller admission path may derive this fact itself (SEM-05).
	Drops bool `json:"drops"`
}

// ReturnContract is a function's total return contract under
// lang.interface/1 (D-07-09) — it subsumes the old optional PublicOrigin
// (R-01): the same two access modes, now total rather than optional. Every
// field is required except where noted.
type ReturnContract struct {
	// Type is copied verbatim from core.Function.ReturnType (R-01).
	Type string `json:"type"`
	// Mode is "owned", "shared", or "exclusive" (R-01), derived from
	// core.Function.PublicOrigin: nil maps to {Mode: "owned", Paths: []};
	// Access "shared"/"exclusive" maps to {shared/exclusive, Paths}.
	Mode string `json:"mode"`
	// Paths names the parameter path(s) this return derives from. Legally
	// empty IFF Mode == "owned" (D-07-09); refused otherwise.
	Paths []string `json:"paths"`
	// Fresh is producer-derived (R-01): true iff Mode == "owned" and the
	// returned type carries the drop ability, i.e. the caller inherits a
	// new drop obligation. Producer-side body reading is legal; no caller
	// admission path may derive this fact itself (SEM-05).
	Fresh bool `json:"fresh"`
}

// ForeignReach is a function's closure-derived worst-case foreign reach
// (R-01): plain strings per the ForeignContract precedent (core.go:60-64,
// planner discretion exercised — plain strings, not a new enumerated type).
// "" is each field's legal empty value, meaning "none reachable".
type ForeignReach struct {
	Allocator    string `json:"allocator"`
	Unwind       string `json:"unwind"`
	NonlocalExit string `json:"nonlocal_exit"`
}

// ForeignReachConflict is 07-12's declared sentinel (CR-03/PVG-02): the
// value a worst-case join over a function's call closure produces for
// ForeignReach.Allocator (or, in principle, Unwind/NonlocalExit) when two
// callees declare DIFFERENT non-empty values and today's vocabulary
// provides no ordering between them -- an allocator name has no
// "more constraining than" relation the way "forbidden" does over
// "permitted". It means "assume the most constraining reach": a consumer
// must treat this exactly as if the reach were maximally hostile, never as
// an absent or default value. It exists so the join never resolves such a
// disagreement by an arbitrary pick (whichever callee happened to be
// iterated last) -- a last-writer-wins copy is precisely the defect class
// CR-03 found. See PHASE-07-DEBT.md's D-07-53 for the join's disclosed
// granularity limits.
const ForeignReachConflict = "conflict"

// InterfaceV0 and FunctionSignatureV0 pin the frozen lang.interface/0 shape
// (D-07-08): exactly the pre-/1-bump field set. A /0 document is decodable
// but never admissible for a call: FunctionSignatureV0 has no Callable,
// Parameters, Return, or ClosureDigest field at all — admission is
// structurally unreachable, not merely refused by convention, the same
// argument style already used for Interface/FunctionSignature above.
type InterfaceV0 struct {
	Schema     string                `json:"schema"`
	ModuleID   string                `json:"module_id"`
	CoreDigest string                `json:"core_digest"`
	Functions  []FunctionSignatureV0 `json:"functions"`
}

// FunctionSignatureV0 is the frozen pre-/1 FunctionSignature shape (D-07-08).
type FunctionSignatureV0 struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Parameter    Parameter     `json:"parameter"`
	ReturnType   string        `json:"return_type"`
	PublicOrigin *PublicOrigin `json:"public_origin,omitempty"`
	Abilities    []Ability     `json:"abilities"`
}

// DecodedInterface is DecodeInterface's dispatch result (D-07-36): exactly
// one of V0/V1 is populated, matching which schema the document declared.
// Admissible is true only for a validated /1 document — a /0 document is
// decodable (V0 populated) but Admissible is always false, and V0's own
// type has no Callable/Parameters/Return/ClosureDigest field to admit from
// even if a caller ignored Admissible (T-07-02).
type DecodedInterface struct {
	Schema     string
	V0         *InterfaceV0
	V1         *Interface
	Admissible bool
}

// DecodeError is DecodeInterface's typed refusal, matching
// originvalidate.Error's {Code}-only shape so callers can dispatch on Code
// the same way.
type DecodeError struct{ Code string }

func (e *DecodeError) Error() string { return e.Code }

// interfaceSchemaPeek reads only the schema string from an interface
// document (D-07-36's schema-peek dispatch), without decoding any other
// field, so a malformed or oversized body never has to be structurally
// interpreted before the dispatch decision is made.
type interfaceSchemaPeek struct {
	Schema string `json:"schema"`
}

// decoderRequireNonEmpty is D-07-42's unexported fault-injection seam for
// TestDecodeInterfaceRequiredModeMutationKilled (a same-package test, per
// D-07-42 — never an exported package-level var on a production path):
// production always requires a non-empty value; the test temporarily
// widens it to accept "" so the missing-Mode refusal is proven to actually
// bite, not merely appear to. Restoring it (via defer) restores the
// refusal. This mirrors pathoracle.go's injection shape (pathoracle.go:38-
// 51's bounded-const discipline): the accepted set is fail-closed and is
// never widened to make a test pass in production.
var decoderRequireNonEmpty = func(value string) bool { return value != "" }

// isValidDigest reports whether value matches the exact shape
// originvalidate.digest() emits: "sha256:" followed by exactly 64 lowercase
// hex characters. DecodeInterface never mints a second digest-shape check —
// this is a shape check only, not a re-derivation of the digest itself.
func isValidDigest(value string) bool {
	const prefix = "sha256:"
	if len(value) != len(prefix)+64 || value[:len(prefix)] != prefix {
		return false
	}
	for _, r := range value[len(prefix):] {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

// DecodeInterface implements D-07-36: it peeks only the schema string, then
// dispatches lang.interface/0 to the pinned InterfaceV0 path, lang.interface/1
// to strict validation, and any other schema value to a refusal. The
// accepted-schema set {InterfaceSchema, InterfaceSchema1} is fail-closed
// and is never widened to make a test pass.
func DecodeInterface(data []byte) (DecodedInterface, error) {
	var peek interfaceSchemaPeek
	if err := json.Unmarshal(data, &peek); err != nil {
		return DecodedInterface{}, &DecodeError{Code: "core.interface_invalid_json"}
	}
	switch peek.Schema {
	case InterfaceSchema:
		var v0 InterfaceV0
		if err := json.Unmarshal(data, &v0); err != nil {
			return DecodedInterface{}, &DecodeError{Code: "core.interface_invalid_json"}
		}
		return DecodedInterface{Schema: peek.Schema, V0: &v0, Admissible: false}, nil
	case InterfaceSchema1:
		return decodeInterfaceV1(data)
	default:
		return DecodedInterface{}, &DecodeError{Code: "core.interface_unknown_schema"}
	}
}

// decodeInterfaceV1 is DecodeInterface's strict lang.interface/1 validation
// path (D-07-36): every required field named on FunctionSignature/
// ParameterContract/ReturnContract's doc comments is checked present and
// non-empty, Mode is checked against its closed set, both digest fields are
// checked against isValidDigest's shape, and no two functions may share an
// ID. "Foreign" needs a raw-JSON presence check specifically because
// ForeignReach's own zero value ("" in every field) is a LEGAL value
// (meaning "no foreign reach"), indistinguishable from an absent key once
// typed-unmarshaled — every other required field's zero value is illegal,
// so a typed-value check alone is sufficient for it.
func decodeInterfaceV1(data []byte) (DecodedInterface, error) {
	var v1 Interface
	if err := json.Unmarshal(data, &v1); err != nil {
		return DecodedInterface{}, &DecodeError{Code: "core.interface_invalid_json"}
	}
	var raw struct {
		Functions []map[string]json.RawMessage `json:"functions"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return DecodedInterface{}, &DecodeError{Code: "core.interface_invalid_json"}
	}

	if !decoderRequireNonEmpty(v1.ModuleID) {
		return DecodedInterface{}, &DecodeError{Code: "core.interface_missing_field"}
	}
	if !isValidDigest(v1.CoreDigest) {
		return DecodedInterface{}, &DecodeError{Code: "core.interface_invalid_digest"}
	}

	seenIDs := make(map[string]bool, len(v1.Functions))
	for index, function := range v1.Functions {
		if seenIDs[function.ID] {
			return DecodedInterface{}, &DecodeError{Code: "core.interface_duplicate_function_id"}
		}
		seenIDs[function.ID] = true

		if !decoderRequireNonEmpty(function.ID) || !decoderRequireNonEmpty(function.Name) {
			return DecodedInterface{}, &DecodeError{Code: "core.interface_missing_field"}
		}
		if len(function.Parameters) == 0 {
			return DecodedInterface{}, &DecodeError{Code: "core.interface_missing_field"}
		}
		for _, parameter := range function.Parameters {
			if parameter.Mode == "" {
				if !decoderRequireNonEmpty(parameter.Mode) {
					return DecodedInterface{}, &DecodeError{Code: "core.interface_missing_field"}
				}
			} else if parameter.Mode != "owned" && parameter.Mode != "shared" && parameter.Mode != "exclusive" {
				return DecodedInterface{}, &DecodeError{Code: "core.interface_invalid_mode"}
			}
		}
		if function.Return.Mode == "" {
			if !decoderRequireNonEmpty(function.Return.Mode) {
				return DecodedInterface{}, &DecodeError{Code: "core.interface_missing_field"}
			}
		} else if function.Return.Mode != "owned" && function.Return.Mode != "shared" && function.Return.Mode != "exclusive" {
			return DecodedInterface{}, &DecodeError{Code: "core.interface_invalid_mode"}
		} else if (function.Return.Mode == "owned") != (len(function.Return.Paths) == 0) {
			return DecodedInterface{}, &DecodeError{Code: "core.interface_missing_field"}
		}
		if index < len(raw.Functions) {
			if _, present := raw.Functions[index]["foreign"]; !present {
				return DecodedInterface{}, &DecodeError{Code: "core.interface_missing_field"}
			}
		}
		if !decoderRequireNonEmpty(function.ClosureDigest) {
			return DecodedInterface{}, &DecodeError{Code: "core.interface_missing_field"}
		}
		if !isValidDigest(function.ClosureDigest) {
			return DecodedInterface{}, &DecodeError{Code: "core.interface_invalid_digest"}
		}
	}
	return DecodedInterface{Schema: InterfaceSchema1, V1: &v1, Admissible: true}, nil
}

type Parameter struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type TypeRef struct {
	Constructor string    `json:"constructor"`
	Arguments   []TypeRef `json:"arguments"`
}

type Ability string

const (
	AbilityCopy   Ability = "copy"
	AbilityDrop   Ability = "drop"
	AbilityShare  Ability = "share"
	AbilitySend   Ability = "send"
	AbilityEscape Ability = "escape"
)

type AbilityWitness struct {
	Ability Ability  `json:"ability"`
	Path    []string `json:"path"`
}

type TypeFact struct {
	ID                string           `json:"id"`
	Shape             TypeRef          `json:"shape"`
	Abilities         []Ability        `json:"abilities"`
	NegativeWitnesses []AbilityWitness `json:"negative_witnesses"`
}

type Place struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	TypeID string `json:"type_id"`
}

type OperationKind string

const (
	OpCopy            OperationKind = "copy"
	OpMove            OperationKind = "move"
	OpBorrowShared    OperationKind = "borrow_shared"
	OpBorrowExclusive OperationKind = "borrow_exclusive"
	OpReturn          OperationKind = "return"
	// OpForeignCall is Phase 4's only call surface (D-04-01): a fallible call
	// to a declared `foreign C {}` symbol. It produces two successor edges
	// (OkEdgeID/ErrEdgeID below) rather than a single target place.
	OpForeignCall OperationKind = "foreign_call"
	// OpFail is a terminator (see TerminatorKinds) that ends a block on the
	// err edge of an OpForeignCall, carrying a place of the declared failure
	// ADT (D-04-04/D-04-09): the only producer of a "typed_failure" terminal
	// outcome anywhere in the IR.
	OpFail OperationKind = "fail"
	// OpRelease is Phase 4 plan 02's resource-lifecycle operation (D-04-07):
	// a non-terminal transition discharging exactly one completed
	// OpForeignCall acquisition. It is never a terminator -- a block always
	// ends in OpReturn or OpFail, with zero or more OpRelease operations
	// immediately before that terminator.
	OpRelease OperationKind = "release"
	// OpDefect is a terminator (see TerminatorKinds) admissible only in a
	// match arm's terminal position (D-04-15): a real, reachable, abort-only
	// terminal outcome carrying the required non-empty reason string on its
	// own Reason field. It performs no release and reads no place value it
	// returns -- its SourceID exists only so the "every operation reads an
	// initialized place" invariant stays uniform across every OperationKind.
	OpDefect OperationKind = "defect"
	// OpCall is Phase 07's Lang-to-Lang call surface (D-07-29): unlike
	// OpForeignCall's two successor edges, an OpCall carries a single
	// TargetID (exactly like OpCopy) plus a CalleeID naming the resolved
	// callee's function ID. It is never a terminator (see TerminatorKinds)
	// -- a call is an ordinary binding, not a block-ending outcome.
	OpCall OperationKind = "call"
)

// CallCalleeUnresolved is Phase 07's typed identity for an OpCall whose
// CalleeID names no declared function (D-07-45): distinct from the
// call-graph cycle code (07-06), so check and corevalidate can each assert
// the same stable fact independently, and so 07-06's callgraph has a code
// to reuse as an inert string constant rather than inventing a second one.
// A CalleeID naming no declared function must never be silently dropped --
// a dropped edge is how a cycle escapes detection.
const CallCalleeUnresolved = "core.call_callee_unresolved"

// CalleeNotCallable is 07-05's checkpoint-ratified stable code (D-04-03,
// SEM-06) for an OpCall whose CalleeID resolves to a DECLARED function that
// is nonetheless not callable — Callable is publication safety
// (originvalidate.PublishProblemsFor, D-07-31/D-07-32), never export
// membership, and a call to a callee that fails that predicate is refused
// with diagnostic.Error and carries no repairs (D-07-31c: exporting the
// callee cannot fix an unsafe borrow-derived return, so no export_callee
// repair is ever offered for this code). Distinct from CallCalleeUnresolved
// (the callee names no declared function at all) and from the (07-06)
// call-graph cycle code — a fourth, materially different fact each time,
// not a fourth spelling of the same one.
const CalleeNotCallable = "core.callee_not_callable"

// CallGraphCycle is 07-06's checkpoint-ratified (D-07-15) stable code for a
// call graph whose OpCall edges form a cycle -- SEM-07's named refusal,
// singular: self-recursion, mutual recursion, and any longer cycle all
// carry this ONE code, distinguished by each cause's own cycle_length and
// membership, never by a second or third code. check's own
// internal/compiler/callgraph package derives this refusal from an
// iterative three-color DFS over its own completed in-memory program,
// before that program is ever returned (D-07-14); corevalidate (07-07)
// independently re-derives the same refusal with its own traversal and
// emits this identical inert string constant, with its own message and no
// spans. Distinct from CallCalleeUnresolved (the callee names no declared
// function at all) and from CalleeNotCallable (a declared but unpublishable
// callee) -- a third, materially different fact, not a third spelling of
// either.
const CallGraphCycle = "core.call_graph_cycle"

// CallArgumentTypeMismatch is 07-09's peer-side stable code for
// corevalidate's OWN, independently-derived refusal of an OpCall whose
// SOURCE place's own TypeFact.Shape.Constructor does not equal
// functionByID[operation.CalleeID].Parameter.Type -- the callee's own
// declared parameter contract, resolved through this program's own
// function set, never through check's callSignatureTable and never by
// asking check anything. Distinct from core.type_mismatch (an operation's
// own source.TypeID == operation.TypeID internal-consistency law, about
// ONE operation) -- this is a CROSS-FUNCTION contract: the caller's
// argument type versus a DIFFERENT function's declared parameter type.
// check's own, materially different derivation (an AST-derived
// name-keyed callee-contract table, consulted at emission time in
// resolveCallBinding) is check.call_argument_type_mismatch -- a
// deliberately different code string in a deliberately different
// namespace, so a diagnostics list can distinguish "check's gate fired"
// from "corevalidate's peer fired" rather than reading one fact twice
// under two names.
const CallArgumentTypeMismatch = "core.call_argument_type_mismatch"

// CallReturnTypeMismatch is 07-09's peer-side stable code for
// corevalidate's OWN, independently-derived refusal of an OpCall whose
// TARGET place's own TypeFact.Shape.Constructor does not equal
// functionByID[operation.CalleeID].ReturnType -- the callee's own
// declared return contract, resolved the same way as
// CallArgumentTypeMismatch above. Distinct from core.type_mismatch (a
// single operation's own internal consistency) and from
// type.return_mismatch (a function's own declared return type versus its
// own declared parameter type, checked once per function, never at a
// call site). check's own derivation of this same fact -- deriving the
// OpCall target's TypeID from the callee's declared return type at
// emission time, fail-closed when unresolvable -- is
// check.call_return_type_unrepresentable, again a distinct code in a
// distinct namespace for the same independence reason.
const CallReturnTypeMismatch = "core.call_return_type_mismatch"

// AllOperationKinds returns every declared OperationKind, in declaration
// order. This is the single table every dispatch site (check, corevalidate,
// interp, cgen, pathoracle, originvalidate) is tested against (D-04-22): a
// constant added to the block above without also being added to this literal
// slice is exactly the defect this registry exists to catch --
// TestAllOperationKindsRegistered fails the moment the two counts diverge.
func AllOperationKinds() []OperationKind {
	return []OperationKind{OpCopy, OpMove, OpBorrowShared, OpBorrowExclusive, OpReturn, OpForeignCall, OpFail, OpRelease, OpDefect, OpCall}
}

// TerminatorKinds returns exactly the operation kinds that end a block --
// always a subset of AllOperationKinds(). Phase 4 adds OpFail and OpDefect;
// every pre-Phase-4 core artifact only ever ends a block with OpReturn.
func TerminatorKinds() []OperationKind {
	return []OperationKind{OpReturn, OpFail, OpDefect}
}

type LinearOperation struct {
	ID       string        `json:"id"`
	PointID  string        `json:"point_id"`
	Kind     OperationKind `json:"kind"`
	SourceID string        `json:"source_id"`
	TargetID string        `json:"target_id,omitempty"`
	LoanID   string        `json:"loan_id,omitempty"`
	TypeID   string        `json:"type_id"`
	// OkEdgeID, ErrEdgeID, and ErrTargetID are Phase 4 additive omitempty
	// facts populated only on an OpForeignCall operation (D-04-04): the ok
	// edge continues at TargetID (an ordinary place, exactly like OpCopy's
	// target), while the err edge's own synthesized failure-ADT place is
	// ErrTargetID. Every pre-Phase-4 operation, and every operation kind
	// other than OpForeignCall, leaves all three empty.
	OkEdgeID    string `json:"ok_edge_id,omitempty"`
	ErrEdgeID   string `json:"err_edge_id,omitempty"`
	ErrTargetID string `json:"err_target_id,omitempty"`
	// ReleasesOperationID is Phase 4 plan 02's additive omitempty fact
	// (D-04-07): populated only on an OpRelease operation, it names the
	// OpForeignCall operation ID this release discharges, so a release is
	// always traceable to exactly one completed acquisition. Every
	// pre-plan-02 operation, and every operation kind other than OpRelease,
	// leaves this empty.
	ReleasesOperationID string `json:"releases_operation_id,omitempty"`
	// Allocator is Phase 4 plan 03's additive omitempty fact
	// (T-04-14/allocator-identity requirement): populated on an OpForeignCall
	// with that acquisition's own declared allocator identity, and copied
	// verbatim onto the OpRelease that discharges it. A release whose
	// Allocator differs from its own acquisition's is refused independently
	// by check (at emission time) and by corevalidate (by re-fetching the
	// acquisition and comparing, never trusting check's bookkeeping). Every
	// pre-plan-03 operation, and every operation kind other than
	// OpForeignCall/OpRelease, leaves this empty.
	Allocator string `json:"allocator,omitempty"`
	// Reason is Phase 4 plan 04's additive omitempty fact (D-04-15):
	// populated only on an OpDefect operation, it carries the required
	// non-empty reason string the source `defect "<reason>"` terminator
	// declared. Every pre-plan-04 operation, and every operation kind other
	// than OpDefect, leaves this empty.
	Reason string `json:"reason,omitempty"`
	// CalleeID is Phase 07's additive omitempty fact (D-07-29): populated
	// only on an OpCall operation, it holds the resolved callee's function
	// ID -- never its name. Every pre-Phase-07 operation, and every
	// operation kind other than OpCall, leaves this empty, so no
	// pre-Phase-07 core artifact moves a byte. Three refusals follow from
	// this invariant, derived independently at both the check and
	// corevalidate sites: an OpCall with an empty CalleeID; any non-OpCall
	// operation with a non-empty CalleeID; and a CalleeID naming no
	// declared function (its own typed identity, distinct from the cycle
	// code -- D-07-45).
	CalleeID string `json:"callee_id,omitempty"`
}

type LinearBody struct {
	ID         string            `json:"id"`
	Types      []TypeFact        `json:"types"`
	Places     []Place           `json:"places"`
	Operations []LinearOperation `json:"operations"`
	// Blocks, Edges, and LoanEndpoints are additive omitempty Phase 3 facts:
	// a plain straight-line linear body (Phase 2 and earlier) never
	// populates them, so its serialized bytes are unchanged (D-13). A body
	// that lowers from a branching match arm populates Blocks/Edges to
	// record its control-flow graph; LoanEndpoints records edge/point-
	// specific loan liveness once that analysis lands (03-03+) and stays
	// empty until then.
	Blocks        []Block        `json:"blocks,omitempty"`
	Edges         []Edge         `json:"edges,omitempty"`
	LoanEndpoints []LoanEndpoint `json:"loan_endpoints,omitempty"`
}

// Block is one basic block of a function's control-flow graph. IDs are
// function-local semantic ordinals (e.g. "{functionID}:block:entry",
// "{functionID}:block:arm:{n}", "{functionID}:block:join"), never source
// offsets, following the MatchArm.EdgeID precedent (core.go, Phase 1).
type Block struct {
	ID           string   `json:"id"`
	PointID      string   `json:"point_id"`
	OperationIDs []string `json:"operation_ids"`
	Successors   []string `json:"successors"`
}

// Edge is one directed control-flow edge between two blocks, keyed by the
// match arm pattern that selects it.
type Edge struct {
	ID          string `json:"id"`
	FromBlockID string `json:"from_block_id"`
	ToBlockID   string `json:"to_block_id"`
	Pattern     string `json:"pattern"`
}

// LoanEndpoint records where a loan ends, either at a point inside a block or
// on an edge between two blocks (never both). checkBranch's arm blocks
// populate this from loanLivenessFixpoint's backward worklist dataflow
// (03-03+). As of D-05-35(d) these endpoints are LOAD-BEARING, not
// decorative: the same fixpoint that materializes them (via
// computeLoanLastUses) is now the sole law deciding loan conflict/expiry
// admission in both analyzeStraightLine and analyzeArmBody -- the forward
// chain-inheritance scan (discoverLoanLastUses) that used to decide
// admission independently of this dataflow is retired, authorized by a
// recorded zero-divergence shadow run of the two laws over the full
// TestOwnershipSequenceExhaustive/TestBranchSequenceExhaustive enumeration.
// Straight-line functions still never populate this FIELD in their
// serialized core (see LinearBody.LoanEndpoints' own doc comment) -- only
// checkBranch's arm blocks serialize it -- but the fixpoint driving it now
// decides real admission everywhere, not only where it is observable.
type LoanEndpoint struct {
	ID               string `json:"id"`
	LoanID           string `json:"loan_id"`
	Kind             string `json:"kind"` // "point" or "edge"
	BlockID          string `json:"block_id,omitempty"`
	EdgeID           string `json:"edge_id,omitempty"`
	AfterOperationID string `json:"after_operation_id,omitempty"`
}

// HasClosedBody reports whether the function's body union is well-formed.
// Ordinarily this is a strict XOR of Match and Linear (Phase 1/2). Phase 3
// adds a third, deliberately narrow case: a match function whose arms carry
// linear bodies populates BOTH Match (arms/patterns) and Linear (the
// flattened operations plus Blocks/Edges the arms lower into) at once. That
// combination is closed only when every arm actually carries a block — see
// Match.HasBlocks and 03-PATTERNS inconsistency I-7.
func (f Function) HasClosedBody() bool {
	if f.Match != nil && f.Linear != nil {
		return f.Match.HasBlocks()
	}
	return (f.Match == nil) != (f.Linear == nil)
}

type Match struct {
	ID        string     `json:"id"`
	PointID   string     `json:"point_id"`
	Scrutinee string     `json:"scrutinee"`
	Arms      []MatchArm `json:"arms"`
}

// HasBlocks reports whether any arm carries a block (an arm-body lowering).
// A match with no arm bodies is unaffected: HasBlocks is false and the
// function's Linear field stays nil, keeping Phase 1 byte-identical.
func (m Match) HasBlocks() bool {
	for _, arm := range m.Arms {
		if arm.BlockID != "" {
			return true
		}
	}
	return false
}

type MatchArm struct {
	ID      string `json:"id"`
	EdgeID  string `json:"edge_id"`
	Pattern string `json:"pattern"`
	Value   string `json:"value"`
	// BlockID is the Phase 3 extension: present only when this arm's value
	// position held a full linear body, naming the block its operations
	// were lowered into. Exactly one of Value and BlockID is populated for
	// any given arm in a function that has arm bodies at all; Value alone
	// is populated for every arm of a Phase 1/2 bare-name-only match.
	BlockID string `json:"block_id,omitempty"`
}
