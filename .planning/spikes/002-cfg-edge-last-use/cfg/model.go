package cfg

import "ai-lang/ownership-kernel-workbench/ownership"

type Block struct {
	ID         string                `json:"id"`
	Operations []ownership.Operation `json:"operations,omitempty"`
	Successors []string              `json:"successors,omitempty"`
}

type Program struct {
	ID     string  `json:"id"`
	Entry  string  `json:"entry"`
	Blocks []Block `json:"blocks"`
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

type Diagnostic struct {
	Code    string `json:"code"`
	Program string `json:"program"`
	Block   string `json:"block,omitempty"`
	Message string `json:"message"`
}

type Endpoint struct {
	ID             string `json:"id"`
	Kind           string `json:"kind"`
	Loan           string `json:"loan"`
	Block          string `json:"block,omitempty"`
	AfterOperation int    `json:"after_operation"`
	From           string `json:"from,omitempty"`
	To             string `json:"to,omitempty"`
}

type Options struct {
	// Test-only fault injection. Omitting edge-specific endpoints recreates a
	// plausible implementation that performs only block-local last-use work.
	OmitEdgeEnds bool
}

type Analysis struct {
	Program              string              `json:"program"`
	Valid                bool                `json:"valid"`
	Diagnostic           *Diagnostic         `json:"diagnostic,omitempty"`
	LiveIn               map[string][]string `json:"live_in"`
	LiveOut              map[string][]string `json:"live_out"`
	Endpoints            []Endpoint          `json:"endpoints"`
	Blocks               int                 `json:"blocks"`
	Edges                int                 `json:"edges"`
	Loans                int                 `json:"loans"`
	TransferEvaluations  int                 `json:"transfer_evaluations"`
	WorklistReinsertions int                 `json:"worklist_reinsertions"`
}

type PathComparison struct {
	ID             string           `json:"id"`
	Blocks         []string         `json:"blocks"`
	Agreement      bool             `json:"agreement"`
	Analyzer       ownership.Result `json:"analyzer"`
	Oracle         ownership.Result `json:"oracle"`
	InsertedEnds   []string         `json:"inserted_ends"`
	OracleEnds     []string         `json:"oracle_ends"`
	TerminalGuards []string         `json:"terminal_guards,omitempty"`
}

type Comparison struct {
	Program       string           `json:"program"`
	Agreement     bool             `json:"agreement"`
	AnalyzerValid bool             `json:"analyzer_valid"`
	OracleValid   bool             `json:"oracle_valid"`
	Analysis      Analysis         `json:"analysis"`
	Paths         []PathComparison `json:"paths"`
}

type SearchResult struct {
	Programs     int         `json:"programs"`
	Mismatch     *Comparison `json:"mismatch,omitempty"`
	MinimalInput *Program    `json:"minimal_input,omitempty"`
}
