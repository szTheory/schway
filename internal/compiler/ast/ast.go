package ast

import "github.com/szTheory/schway/internal/compiler/diagnostic"

type Program struct {
	Module  string
	Exports []Export
	Data    []DataDecl
	Funcs   []FuncDecl
	// Foreign is the Phase 4 `foreign C { }` declaration surface (D-04-01):
	// zero or more blocks, each declaring one or more foreign symbols a
	// linear body may call fallibly through `try`.
	Foreign []ForeignBlock
}

// ForeignBlock is one `foreign C { }` declaration. Language is the literal
// text following `foreign` (this phase only admits "C").
type ForeignBlock struct {
	Language string
	Symbols  []ForeignSymbol
	Span     diagnostic.Span
}

// ForeignSymbol is one declared foreign C function this phase's linear
// bodies may call through `try`. Policies carries every `key: value` line the
// declaration wrote, in source order, so check.go — not the parser — decides
// which keys are missing (D-04-16: absence is refusal, with no default).
type ForeignSymbol struct {
	Name       string
	Parameter  Parameter
	ReturnType TypeRef
	Policies   []ForeignPolicy
	Span       diagnostic.Span
}

// ForeignPolicy is one `key: value` line inside a foreign symbol's
// declaration body (e.g. `unwind: forbidden`, `allocator: "libc_malloc"`).
// IsString records whether Value was written as a quoted string literal
// (allocator identity) or a bare identifier (policy names).
type ForeignPolicy struct {
	Key      string
	Value    string
	IsString bool
	Span     diagnostic.Span
}

type Export struct {
	Kind string
	Name string
	Span diagnostic.Span
}

type DataDecl struct {
	Name         string
	Alternatives []Alternative
	Span         diagnostic.Span
}

type Alternative struct {
	Name string
	// PayloadType is Phase 12's additive field (D-12-11/D-12-13): the bare
	// type name written inside `( ... )` following the alternative's own
	// name (e.g. `Ok(Buffer)`), or "" for a nullary alternative -- mirrors
	// DataType.PayloadType's bare-string convention.
	PayloadType string
	Span        diagnostic.Span
}

type FuncDecl struct {
	Name         string
	Parameter    Parameter
	ReturnOrigin *BorrowOrigin
	ReturnType   TypeRef
	Body         Body
	Span         diagnostic.Span
}

// BorrowOrigin is the Phase 3 return-type annotation `borrow(path)` /
// `borrow mut(path)` preceding an ordinary return TypeRef. Its presence is
// the syntactic discriminant for the borrowed-view return case (OWN-04):
// exactly one path is supported this phase (the function's own parameter
// name — the language has a single parameter and no field-path-bearing
// executable shape), and Access names the declared access mode.
type BorrowOrigin struct {
	Path   string
	Access string // "shared" | "exclusive"
	Span   diagnostic.Span
}

type Parameter struct {
	Name string
	Type TypeRef
	Span diagnostic.Span
}

type TypeRef struct {
	Constructor string
	Arguments   []TypeRef
}

type Body struct {
	MatchExpr
	Linear *LinearBody
}

func (b Body) HasMatch() bool { return b.MatchExpr.Scrutinee != "" }

func (b Body) HasClosedVariant() bool { return b.HasMatch() != (b.Linear != nil) }

type MatchExpr struct {
	Scrutinee string
	Arms      []MatchArm
	Span      diagnostic.Span
}

type MatchArm struct {
	Pattern string
	Value   string
	// Body is the Phase 3 extension: an arm's value position may hold a full
	// linear body instead of a bare alternative name. Exactly one of Value
	// and Body is populated — see HasClosedVariant, the arm-level analog of
	// Body.HasClosedVariant (03-PATTERNS inconsistency I-7).
	Body *LinearBody
	// Binder is Phase 12's additive field (D-12-13): the identifier written
	// inside `( ... )` in this arm's PATTERN position (a destructuring
	// binder, e.g. `Ok(v) =>`) -- "" when the pattern position carries no
	// parenthesized identifier. Binder is orthogonal to the Value/Body
	// exclusivity above: an arm may carry a binder alongside either form.
	Binder string
	// ConstructBinder is Phase 12 Plan 03's additive field (D-12-15's third
	// refusal): the identifier written inside `( ... )` in this arm's bare
	// VALUE position (a construction argument, e.g. `=> Ok(w)`) -- ""
	// when the value position carries no parenthesized identifier, or when
	// the arm has no Value at all (a Body arm). Kept SEPARATE from Binder
	// rather than collapsed at parse time (the earlier Plan 02 shape
	// silently let a non-empty value-side name override an empty or
	// DIFFERING pattern-side one) precisely so check.go can compare the two
	// names: D-12-14 provides exactly one shared payload place per arm this
	// phase, so an arm naming two DIFFERENT places (Binder != "" &&
	// ConstructBinder != "" && Binder != ConstructBinder) is a genuine
	// arity mismatch, not a stylistic choice -- check.payload_arity_mismatch.
	ConstructBinder string
	Span            diagnostic.Span
}

// HasClosedVariant reports whether exactly one of the arm's two value forms
// (a bare alternative name, or a full linear body) is populated.
func (a MatchArm) HasClosedVariant() bool { return (a.Value != "") != (a.Body != nil) }

type LinearBody struct {
	Bindings []Binding
	Result   string
	// TerminalMatch is Phase 18's sole extension to a linear function body:
	// after its ordered prefix, the body may end in the existing match form.
	// It is nil for ordinary straight-line and match-arm bodies.
	TerminalMatch *MatchExpr
	// DefectReason is populated instead of Result when this body's terminal
	// position is `defect "<reason>"` (D-04-15): a real, reachable, abort-only
	// terminal outcome. Exactly one of Result and DefectReason is ever
	// populated for any given body.
	DefectReason string
	Span         diagnostic.Span
}

type Binding struct {
	Name string
	RHS  RHS
	Span diagnostic.Span
}

type RHS struct {
	Kind   string
	Source string
	Span   diagnostic.Span
	// Numeric literals keep their exact source spelling in Source and their
	// token bounds in Span until checker admission assigns numeric meaning.
	// Callee and Arguments are populated when Kind == "try_call",
	// Kind == "discard_call" (D-04-06): a fallible foreign call, admissible
	// only as the operand of `try` or `discard ... because`; or
	// Kind == "call" (Phase 07, D-07-01): a bare call whose callee identity
	// (Lang function, foreign symbol, or unresolved) is a check-time fact,
	// not a parse-time one -- the parser accepts the shape unconditionally
	// and admission relocates to check. Callee names the callee; Arguments
	// is its argument place names in source order. "call" never sets
	// Rationale.
	Callee    string
	Arguments []string
	// ArgumentSpans preserves exact token bounds for source edits that replace
	// one argument without touching the callee or surrounding punctuation.
	ArgumentSpans []diagnostic.Span
	// Rationale is populated only when Kind == "discard_call": the required
	// non-empty string literal explaining why the call's failure is not
	// actionable here (D-04-06). Empty for every other Kind.
	Rationale string
}
