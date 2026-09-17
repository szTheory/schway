package ownership

// Kind is a small kernel instruction. It is intentionally not proposed Lang
// syntax; it exists to make ownership laws executable.
type Kind string

const (
	Declare         Kind = "declare"
	BorrowShared    Kind = "borrow_shared"
	BorrowExclusive Kind = "borrow_exclusive"
	ReadPlace       Kind = "read_place"
	MutatePlace     Kind = "mutate_place"
	ReadLoan        Kind = "read_loan"
	MutateLoan      Kind = "mutate_loan"
	EndLoan         Kind = "end_loan"
	Move            Kind = "move"
	Release         Kind = "release"
	BeginScope      Kind = "begin_scope"
	EndScope        Kind = "end_scope"
	ExitError       Kind = "exit_error"
	ExitCancel      Kind = "exit_cancel"
	ExitPanic       Kind = "exit_panic"
)

type Operation struct {
	Kind        Kind   `json:"kind"`
	Place       string `json:"place,omitempty"`
	Target      string `json:"target,omitempty"`
	Loan        string `json:"loan,omitempty"`
	Scope       string `json:"scope,omitempty"`
	Resource    bool   `json:"resource,omitempty"`
	SourceIndex int    `json:"source_index,omitempty"`
	Synthetic   bool   `json:"synthetic,omitempty"`
}

type Program struct {
	ID         string      `json:"id"`
	Operations []Operation `json:"operations"`
}

type Event struct {
	Sequence    int    `json:"sequence"`
	Kind        string `json:"kind"`
	Place       string `json:"place,omitempty"`
	Target      string `json:"target,omitempty"`
	Loan        string `json:"loan,omitempty"`
	Scope       string `json:"scope,omitempty"`
	Mode        string `json:"mode,omitempty"`
	Outcome     string `json:"outcome,omitempty"`
	SourceIndex int    `json:"source_index"`
	Synthetic   bool   `json:"synthetic,omitempty"`
}

type Cause struct {
	Relation string `json:"relation"`
	Value    string `json:"value"`
}

type Diagnostic struct {
	Code        string   `json:"code"`
	Program     string   `json:"program"`
	SourceIndex int      `json:"source_index"`
	Operation   Kind     `json:"operation"`
	Place       string   `json:"place,omitempty"`
	Loan        string   `json:"loan,omitempty"`
	Scope       string   `json:"scope,omitempty"`
	Message     string   `json:"message"`
	Causes      []Cause  `json:"causes,omitempty"`
	Repairs     []Repair `json:"repairs,omitempty"`
}

type Repair struct {
	Kind   string `json:"kind"`
	Effect string `json:"effect"`
}

type Result struct {
	Program    string      `json:"program"`
	Valid      bool        `json:"valid"`
	Events     []Event     `json:"events"`
	Diagnostic *Diagnostic `json:"diagnostic,omitempty"`
	Operations int         `json:"operations"`
}

type Fixture struct {
	Name        string  `json:"name"`
	Program     Program `json:"program"`
	ExpectValid bool    `json:"expect_valid"`
	ExpectCode  string  `json:"expect_code,omitempty"`
}

type FixtureFile struct {
	Schema   int       `json:"schema"`
	Fixtures []Fixture `json:"fixtures"`
}

type Comparison struct {
	Program   string `json:"program"`
	Agreement bool   `json:"agreement"`
	Oracle    Result `json:"oracle"`
	Checker   Result `json:"checker"`
}

func RepairsFor(code string) []Repair {
	switch code {
	case "ownership.move_while_borrowed":
		return []Repair{
			{Kind: "move_after_last_use", Effect: "keeps zero-copy transfer and shortens the loan"},
			{Kind: "return_owned", Effect: "decouples the result at an explicit copy or allocation cost"},
			{Kind: "share_immutable", Effect: "changes unique ownership into explicit shared ownership"},
		}
	case "ownership.owner_access_conflict":
		return []Repair{
			{Kind: "end_loan_before_access", Effect: "shortens borrowed access"},
			{Kind: "perform_through_exclusive_loan", Effect: "keeps mutation under the existing unique authority"},
		}
	case "ownership.use_after_move":
		return []Repair{
			{Kind: "use_move_target", Effect: "follows the transferred owner"},
			{Kind: "change_transfer_to_borrow", Effect: "preserves caller ownership and narrows callee authority"},
		}
	case "ownership.release_while_borrowed":
		return []Repair{{Kind: "end_loan_before_release", Effect: "ensures no view survives resource release"}}
	case "ownership.mutation_requires_exclusive":
		return []Repair{{Kind: "request_exclusive_access", Effect: "prevents competing readers or writers during mutation"}}
	default:
		return nil
	}
}
