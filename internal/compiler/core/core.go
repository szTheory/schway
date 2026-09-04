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

// ForeignContract is the minimum obligation set this plan populates for one
// declared foreign C symbol a function calls: its name, allocator identity,
// the two mandatory admission policies (D-04-16), and the name of the
// nullary ADT its err edge carries (D-04-05). The fuller three-layer
// contract (target layout, capture/retention) lands in a later plan.
type ForeignContract struct {
	Symbol       string `json:"symbol"`
	Allocator    string `json:"allocator"`
	Unwind       string `json:"unwind,omitempty"`
	NonlocalExit string `json:"nonlocal_exit,omitempty"`
	Fails        string `json:"fails,omitempty"`
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
)

// AllOperationKinds returns every declared OperationKind, in declaration
// order. This is the single table every dispatch site (check, corevalidate,
// interp, cgen, pathoracle, originvalidate) is tested against (D-04-22): a
// constant added to the block above without also being added to this literal
// slice is exactly the defect this registry exists to catch --
// TestAllOperationKindsRegistered fails the moment the two counts diverge.
func AllOperationKinds() []OperationKind {
	return []OperationKind{OpCopy, OpMove, OpBorrowShared, OpBorrowExclusive, OpReturn, OpForeignCall, OpFail}
}

// TerminatorKinds returns exactly the operation kinds that end a block --
// always a subset of AllOperationKinds(). Phase 4 adds OpFail (a later plan
// adds a defect terminator); every pre-Phase-4 core artifact only ever ends
// a block with OpReturn.
func TerminatorKinds() []OperationKind {
	return []OperationKind{OpReturn, OpFail}
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
// on an edge between two blocks (never both). This phase's arm-body lowering
// does not yet populate LoanEndpoints — edge-specific liveness lands in a
// later plan — but the additive shape exists now so that plan does not need
// a schema change.
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
