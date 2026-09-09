package iplive

import (
	"fmt"
	"sort"
)

// Conflict is one move-while-borrowed refusal, identified by the source
// operation that moves and the loan that was still live.
type Conflict struct {
	FunctionID string `json:"function"`
	OpID       string `json:"op"`
	LoanID     string `json:"loan"`
}

func (c Conflict) key() string { return c.FunctionID + "|" + c.OpID + "|" + c.LoanID }

// LocalLiveness is the per-operation live-after set restricted to loans BORN
// in the same function by a borrow operation -- the part of the answer that is
// comparable between a context-sensitive oracle and a context-insensitive
// analyzer (see oracle.go's projection note).
type LocalLiveness map[string][]string

// Analysis is one mechanism's whole-program answer plus its counted work.
type Analysis struct {
	Conflicts   []Conflict    `json:"conflicts"`
	Local       LocalLiveness `json:"-"`
	Work        int           `json:"work"`
	Derivations int           `json:"derivations"`
	Mechanism   string        `json:"mechanism"`
}

// Analyze runs a per-function backward worklist liveness for every declared
// function, consulting the provider once per distinct callee for its summary.
// Summaries are resolved in a pre-pass so a caller's fixpoint iterations never
// re-charge for them: the measured difference between the two providers is
// therefore the cost of DERIVING summaries and nothing else.
func Analyze(p *Program, provider Provider) (Analysis, error) {
	if _, err := p.ReversePostorder(); err != nil {
		return Analysis{}, err
	}
	result := Analysis{Local: LocalLiveness{}, Mechanism: provider.Name()}
	work := 0

	functions := append([]Function(nil), p.Functions...)
	sort.Slice(functions, func(i, j int) bool { return functions[i].ID < functions[j].ID })

	for i := range functions {
		fn := &functions[i]
		if scoped, ok := provider.(functionScoped); ok {
			scoped.beginFunction()
		}
		summaries := map[string]Summary{}
		for _, block := range fn.Blocks {
			for _, op := range block.Ops {
				if op.Kind != OpCall {
					continue
				}
				if _, ok := summaries[op.Callee]; ok {
					continue
				}
				summary, err := provider.Summary(op.Callee)
				if err != nil {
					return Analysis{}, err
				}
				summaries[op.Callee] = summary
			}
		}

		conflicts, local, fnWork, err := analyzeFunction(fn, summaries)
		if err != nil {
			return Analysis{}, err
		}
		work += fnWork
		result.Conflicts = append(result.Conflicts, conflicts...)
		for opID, loans := range local {
			result.Local[opID] = loans
		}
	}

	sort.Slice(result.Conflicts, func(i, j int) bool {
		return result.Conflicts[i].key() < result.Conflicts[j].key()
	})
	result.Work = work + provider.Work()
	result.Derivations = provider.Derivations()
	return result, nil
}

// functionScoped is implemented by a provider whose caching is scoped to one
// caller's analysis rather than to the whole program.
type functionScoped interface{ beginFunction() }

// loanPlaces maps every loan CREATED by a borrow in this body to the place it
// borrows. Call results are deliberately absent: canonicalize has already
// rewritten a borrow-derived result to the loan it aliases.
func loanPlaces(fn *Function) map[string]string {
	places := map[string]string{}
	for _, block := range fn.Blocks {
		for _, op := range block.Ops {
			if op.Kind == OpBorrow {
				places[op.Loan] = op.Place
			}
		}
	}
	return places
}

// canonicalize is the summary-consuming pre-pass. A call whose callee returns a
// borrow of its parameter binds its result to the SAME loan identity as the
// argument, so ordinary backward liveness then carries the loan past the call
// with no interprocedural special case left in the transfer function. A call
// whose callee returns an owned value binds its result to a fresh identity no
// borrow ever creates, so using that result keeps nothing alive.
//
// It walks blocks in slice order, which the corpus keeps topological: a call
// result is bound before any block that uses it (the same assumption the
// expansion oracle makes).
func canonicalize(fn *Function, summaries map[string]Summary) ([]Block, int) {
	alias := map[string]string{}
	canon := func(loan string) string {
		for hops := 0; hops < 64; hops++ {
			next, ok := alias[loan]
			if !ok || next == loan {
				return loan
			}
			loan = next
		}
		return loan
	}
	work := 0
	blocks := make([]Block, 0, len(fn.Blocks))
	for _, block := range fn.Blocks {
		rewrittenBlock := Block{ID: block.ID, Successors: block.Successors}
		for _, op := range block.Ops {
			work++
			rewritten := op
			rewritten.Loan = canon(op.Loan)
			if op.Kind == OpCall && op.Result != "" {
				if summaries[op.Callee].ReturnsBorrowOfParam {
					alias[op.Result] = rewritten.Loan
				} else {
					alias[op.Result] = op.ID + "#owned"
				}
				rewritten.Result = canon(op.Result)
			}
			rewrittenBlock.Ops = append(rewrittenBlock.Ops, rewritten)
		}
		blocks = append(blocks, rewrittenBlock)
	}
	return blocks, work
}

// transfer applies one block's operations backward to liveOut. After
// canonicalization the only interprocedural residue is UsesParam: whether the
// call itself reads through the argument loan.
func transfer(ops []Operation, summaries map[string]Summary, liveOut map[string]bool) (map[string]bool, int) {
	live := make(map[string]bool, len(liveOut))
	for loan := range liveOut {
		live[loan] = true
	}
	work := 0
	for i := len(ops) - 1; i >= 0; i-- {
		op := ops[i]
		work++
		switch op.Kind {
		case OpUse:
			live[op.Loan] = true
		case OpBorrow:
			delete(live, op.Loan) // the loan's region starts here
		case OpCall:
			if summaries[op.Callee].UsesParam {
				live[op.Loan] = true
			}
		case OpMove:
			// decided in the materialization pass, where the converged sets
			// are final -- deciding it here would emit one conflict per
			// worklist iteration
		}
	}
	return live, work
}

func sameSet(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

func analyzeFunction(fn *Function, summaries map[string]Summary) ([]Conflict, LocalLiveness, int, error) {
	blocks, work := canonicalize(fn, summaries)

	byID := make(map[string]*Block, len(blocks))
	order := make([]string, 0, len(blocks))
	for i := range blocks {
		byID[blocks[i].ID] = &blocks[i]
		order = append(order, blocks[i].ID)
	}
	predecessors := map[string][]string{}
	for i := range blocks {
		for _, successor := range blocks[i].Successors {
			if _, ok := byID[successor]; !ok {
				return nil, nil, 0, fmt.Errorf("%s: unresolved successor %q", fn.ID, successor)
			}
			predecessors[successor] = append(predecessors[successor], blocks[i].ID)
		}
	}

	liveIn := make(map[string]map[string]bool, len(blocks))
	for _, id := range order {
		liveIn[id] = map[string]bool{}
	}

	queue := append([]string(nil), order...)
	queued := map[string]bool{}
	for _, id := range order {
		queued[id] = true
	}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		queued[id] = false
		work++ // one transfer-function evaluation

		block := byID[id]
		liveOut := map[string]bool{}
		for _, successor := range block.Successors {
			for loan := range liveIn[successor] {
				liveOut[loan] = true
			}
		}
		newLiveIn, transferWork := transfer(block.Ops, summaries, liveOut)
		work += transferWork
		if !sameSet(liveIn[id], newLiveIn) {
			liveIn[id] = newLiveIn
			for _, predecessor := range predecessors[id] {
				if !queued[predecessor] {
					queue = append(queue, predecessor)
					queued[predecessor] = true
					work++ // one worklist reinsertion
				}
			}
		}
	}

	places := loanPlaces(fn)
	born := map[string]bool{}
	for loan := range places {
		born[loan] = true
	}

	var conflicts []Conflict
	local := LocalLiveness{}
	for i := range blocks {
		block := &blocks[i]
		live := map[string]bool{}
		for _, successor := range block.Successors {
			for loan := range liveIn[successor] {
				live[loan] = true
			}
		}
		for j := len(block.Ops) - 1; j >= 0; j-- {
			op := block.Ops[j]
			local[fn.ID+"|"+op.ID] = restrictSorted(live, born)
			if op.Kind == OpMove {
				for loan := range live {
					if places[loan] == op.Place {
						conflicts = append(conflicts, Conflict{FunctionID: fn.ID, OpID: op.ID, LoanID: loan})
					}
				}
			}
			live, _ = transfer(block.Ops[j:j+1], summaries, live)
		}
	}
	return conflicts, local, work, nil
}

func restrictSorted(live map[string]bool, keep map[string]bool) []string {
	out := make([]string, 0, len(live))
	for loan := range live {
		if keep[loan] {
			out = append(out, loan)
		}
	}
	sort.Strings(out)
	return out
}
