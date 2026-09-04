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
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	EntryPointID  string          `json:"entry_point_id"`
	ReturnPointID string          `json:"return_point_id"`
	Parameter     Parameter       `json:"parameter"`
	ReturnType    string          `json:"return_type"`
	Match         *Match          `json:"match,omitempty"`
	Linear        *LinearBody     `json:"linear,omitempty"`
	Span          diagnostic.Span `json:"span"`
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
)

type LinearOperation struct {
	ID       string        `json:"id"`
	PointID  string        `json:"point_id"`
	Kind     OperationKind `json:"kind"`
	SourceID string        `json:"source_id"`
	TargetID string        `json:"target_id,omitempty"`
	LoanID   string        `json:"loan_id,omitempty"`
	TypeID   string        `json:"type_id"`
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
