package origins

type Ability string

const (
	Copy   Ability = "copy"
	Drop   Ability = "drop"
	Share  Ability = "share"
	Send   Ability = "send"
	Escape Ability = "escape"
)

var AbilityOrder = []Ability{Copy, Drop, Share, Send, Escape}

type TypeExpr struct {
	Name string     `json:"name"`
	Args []TypeExpr `json:"args,omitempty"`
}

type TypeSpec struct {
	Name        string            `json:"name"`
	Parameters  int               `json:"parameters"`
	Base        []Ability         `json:"base,omitempty"`
	Conditional map[Ability][]int `json:"conditional,omitempty"`
}

type Parameter struct {
	Name string   `json:"name"`
	Mode string   `json:"mode"`
	Type TypeExpr `json:"type"`
}

type ReturnCase struct {
	Tag     string   `json:"tag,omitempty"`
	Kind    string   `json:"kind"`
	Access  string   `json:"access,omitempty"`
	Type    TypeExpr `json:"type"`
	Origins []string `json:"origins,omitempty"`
}

type CallbackContract struct {
	Parameter      string    `json:"parameter"`
	FreshOrigin    string    `json:"fresh_origin"`
	BorrowedFrom   []string  `json:"borrowed_from"`
	ResultType     TypeExpr  `json:"result_type"`
	ResultRequires []Ability `json:"result_requires,omitempty"`
}

type BodyFacts struct {
	Returns             []ReturnCase `json:"returns"`
	RetainedFreshOrigin string       `json:"retained_fresh_origin,omitempty"`
}

type Function struct {
	ID         string             `json:"id"`
	TypeParams []string           `json:"type_params,omitempty"`
	Params     []Parameter        `json:"params,omitempty"`
	Returns    []ReturnCase       `json:"returns"`
	Callbacks  []CallbackContract `json:"callbacks,omitempty"`
	Body       *BodyFacts         `json:"body,omitempty"`
}

type Binding struct {
	ID   string   `json:"id"`
	Type TypeExpr `json:"type"`
}

type CallbackArgument struct {
	Parameter     string   `json:"parameter"`
	ResultType    TypeExpr `json:"result_type"`
	ResultOrigins []string `json:"result_origins,omitempty"`
}

type Step struct {
	Kind    string `json:"kind"`
	Binding string `json:"binding,omitempty"`
}

type CallCase struct {
	ID          string              `json:"id"`
	Function    string              `json:"function"`
	Arguments   map[string]string   `json:"arguments,omitempty"`
	TypeArgs    map[string]TypeExpr `json:"type_args,omitempty"`
	ReturnTag   string              `json:"return_tag,omitempty"`
	Callback    *CallbackArgument   `json:"callback,omitempty"`
	Bindings    []Binding           `json:"bindings,omitempty"`
	Steps       []Step              `json:"steps,omitempty"`
	ExpectValid bool                `json:"expect_valid"`
	ExpectCode  string              `json:"expect_code,omitempty"`
}

type FixtureFile struct {
	Schema             int        `json:"schema"`
	Types              []TypeSpec `json:"types"`
	Functions          []Function `json:"functions"`
	DishonestFunctions []Function `json:"dishonest_functions"`
	Cases              []CallCase `json:"cases"`
}

type Diagnostic struct {
	Code       string   `json:"code"`
	Function   string   `json:"function,omitempty"`
	Case       string   `json:"case,omitempty"`
	Binding    string   `json:"binding,omitempty"`
	Origins    []string `json:"origins,omitempty"`
	Ability    Ability  `json:"ability,omitempty"`
	Message    string   `json:"message"`
	RepairKind string   `json:"repair_kind,omitempty"`
}

type CheckResult struct {
	Case       string      `json:"case"`
	Valid      bool        `json:"valid"`
	Diagnostic *Diagnostic `json:"diagnostic,omitempty"`
}

type Interface struct {
	Schema    int        `json:"schema"`
	Types     []TypeSpec `json:"types"`
	Functions []Function `json:"functions"`
}

type Comparison struct {
	Case      string      `json:"case"`
	Agreement bool        `json:"agreement"`
	Consumer  CheckResult `json:"consumer"`
	Oracle    CheckResult `json:"oracle"`
}

type EncodingMetric struct {
	Approach         string `json:"approach"`
	DirectFunctions  int    `json:"direct_functions"`
	AdaptedFunctions int    `json:"adapted_functions"`
	Characters       int    `json:"characters"`
	LexicalTokens    int    `json:"lexical_tokens"`
	ExplicitBinders  int    `json:"explicit_binders"`
	OriginReferences int    `json:"origin_references"`
}

type SearchResult struct {
	Programs int         `json:"programs"`
	Mismatch *Comparison `json:"mismatch,omitempty"`
	Input    *CallCase   `json:"minimal_input,omitempty"`
}
