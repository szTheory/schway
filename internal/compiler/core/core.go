package core

import "github.com/codename-lang/lang/internal/compiler/diagnostic"

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
const InterfaceSchema = "lang.interface/0"

// FunctionSignature is one function's body-stripped public surface: identity,
// parameter/return shape, its declared origin (if any), and its granted
// abilities. It deliberately has no Linear/Match field at all — not merely an
// omitted one — so a consumer decoding this type structurally cannot reach a
// body even by accident.
type FunctionSignature struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Parameter    Parameter     `json:"parameter"`
	ReturnType   string        `json:"return_type"`
	PublicOrigin *PublicOrigin `json:"public_origin,omitempty"`
	Abilities    []Ability     `json:"abilities"`
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
)

// AllOperationKinds returns every declared OperationKind, in declaration
// order. This is the single table every dispatch site (check, corevalidate,
// interp, cgen, pathoracle, originvalidate) is tested against (D-04-22): a
// constant added to the block above without also being added to this literal
// slice is exactly the defect this registry exists to catch --
// TestAllOperationKindsRegistered fails the moment the two counts diverge.
func AllOperationKinds() []OperationKind {
	return []OperationKind{OpCopy, OpMove, OpBorrowShared, OpBorrowExclusive, OpReturn, OpForeignCall, OpFail, OpRelease, OpDefect}
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
