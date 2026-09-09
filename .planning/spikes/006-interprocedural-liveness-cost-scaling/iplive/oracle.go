package iplive

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// The oracle is a whole-program EXPANSION checker, written to share nothing
// with the mechanism under test: it has no notion of a summary at all. Every
// call is inlined at its own call site (context-sensitively), producing one
// monolithic CFG, and liveness is then computed by a naive round-robin
// iteration over every block until nothing changes -- no worklist, no
// predecessor bookkeeping, no summaries.
//
// Projection note. The expansion is context-SENSITIVE and the mechanism is
// context-INSENSITIVE, so the two are only comparable where a
// meet-over-all-contexts union is the right ground truth:
//
//   - Conflicts project exactly. A move-while-borrowed refusal names an
//     operation and a loan that both live in one function's own namespace.
//   - Local liveness projects exactly, restricted to loans BORN in the same
//     function. A parameter loan deliberately is not compared: in a context
//     its liveness is decided by the caller's later uses, which a body-blind
//     analyzer cannot and must not see.
//
// ErrOracleBudget is a bounded-oracle outcome, not a defect: expansion is
// exponential in a sharing-heavy call graph, which is exactly why the real
// checker may not use one.
var ErrOracleBudget = errors.New("oracle expansion budget exceeded")

// OracleResult is the expansion oracle's projected answer.
type OracleResult struct {
	Conflicts []Conflict
	Local     LocalLiveness
	// ExpandedOps is the size of the monolithic program the oracle had to
	// build -- the number this spike reports whenever it wants to show what
	// context sensitivity costs.
	ExpandedOps int
	Covered     map[string]bool
}

type expander struct {
	byID       map[string]*Function
	blocks     []Block
	unsubst    map[string]map[string]string
	instanceOf map[string]string
	ops        int
	budget     int
	depth      int
}

// Oracle expands every root and returns the projected ground truth.
func Oracle(p *Program, budget int) (OracleResult, error) {
	if _, err := p.ReversePostorder(); err != nil {
		return OracleResult{}, err
	}
	exp := &expander{
		byID:       p.index(),
		unsubst:    map[string]map[string]string{},
		instanceOf: map[string]string{},
		budget:     budget,
	}
	roots := append([]string(nil), p.Roots...)
	sort.Strings(roots)
	for _, root := range roots {
		fn, ok := exp.byID[root]
		if !ok {
			return OracleResult{}, fmt.Errorf("unresolved root %q", root)
		}
		// The root's parameter loan has no caller, so it is expanded under
		// its own name: nothing outside can keep it live.
		if _, _, _, err := exp.expand(fn.ID, "root:"+fn.ID+"~"+fn.ParamLoan, "root:"+fn.ID); err != nil {
			return OracleResult{}, err
		}
	}
	return exp.solve(p)
}

// expand inlines one instance of fnID whose parameter loan is argLoan, under
// context ctx. It returns the instance's entry block, its exit blocks, and the
// expanded loan its return value aliases ("" when the return is owned).
func (e *expander) expand(fnID, argLoan, ctx string) (string, []string, string, error) {
	fn, ok := e.byID[fnID]
	if !ok {
		return "", nil, "", fmt.Errorf("unresolved callee %q", fnID)
	}
	e.depth++
	defer func() { e.depth-- }()
	if e.depth > len(e.byID)+1 {
		return "", nil, "", fmt.Errorf("call_graph_cycle at %q", fnID)
	}
	e.instanceOf[ctx] = fnID

	subst := map[string]string{fn.ParamLoan: argLoan}
	name := func(local string) string {
		if local == "" {
			return ""
		}
		if mapped, ok := subst[local]; ok {
			return mapped
		}
		return ctx + "~" + local
	}
	blockEntry := func(blockID string) string { return ctx + "#" + blockID + "#0" }

	var exits []string
	for _, block := range fn.Blocks {
		segment := 0
		segmentID := func() string { return fmt.Sprintf("%s#%s#%d", ctx, block.ID, segment) }
		current := Block{ID: segmentID()}
		for _, op := range block.Ops {
			e.ops++
			if e.budget > 0 && e.ops > e.budget {
				return "", nil, "", ErrOracleBudget
			}
			if op.Kind != OpCall {
				current.Ops = append(current.Ops, Operation{
					ID:    ctx + "|" + op.ID,
					Kind:  op.Kind,
					Loan:  name(op.Loan),
					Place: name(op.Place),
				})
				continue
			}
			calleeCtx := ctx + "/" + op.ID
			calleeEntry, calleeExits, calleeReturn, err := e.expand(op.Callee, name(op.Loan), calleeCtx)
			if err != nil {
				return "", nil, "", err
			}
			segment++
			nextID := segmentID()
			current.Successors = []string{calleeEntry}
			e.blocks = append(e.blocks, current)
			for _, exit := range calleeExits {
				e.attachSuccessor(exit, nextID)
			}
			if op.Result != "" {
				if calleeReturn != "" {
					subst[op.Result] = calleeReturn
				} else {
					// An owned return aliases nothing: give it a fresh name no
					// borrow ever creates, so uses of it keep no loan alive.
					subst[op.Result] = calleeCtx + "~owned"
				}
			}
			current = Block{ID: nextID}
		}
		for _, successor := range block.Successors {
			current.Successors = append(current.Successors, blockEntry(successor))
		}
		if len(block.Successors) == 0 {
			exits = append(exits, current.ID)
		}
		e.blocks = append(e.blocks, current)
	}

	inverse := map[string]string{}
	for local, expanded := range subst {
		inverse[expanded] = local
	}
	for _, block := range fn.Blocks {
		for _, op := range block.Ops {
			for _, local := range []string{op.Loan, op.Place, op.Result} {
				if local == "" {
					continue
				}
				if _, ok := subst[local]; !ok {
					inverse[ctx+"~"+local] = local
				}
			}
		}
	}
	e.unsubst[ctx] = inverse

	returned := ""
	if fn.Returns != "" {
		if mapped, ok := subst[fn.Returns]; ok {
			returned = mapped
		} else {
			returned = ctx + "~" + fn.Returns
		}
	}
	return blockEntry(fn.Blocks[0].ID), exits, returned, nil
}

func (e *expander) attachSuccessor(blockID, successor string) {
	for i := range e.blocks {
		if e.blocks[i].ID == blockID {
			e.blocks[i].Successors = append(e.blocks[i].Successors, successor)
			return
		}
	}
}

// solve runs naive round-robin liveness over the expanded CFG and projects the
// answer back onto source identities.
func (e *expander) solve(p *Program) (OracleResult, error) {
	byID := make(map[string]*Block, len(e.blocks))
	for i := range e.blocks {
		byID[e.blocks[i].ID] = &e.blocks[i]
	}
	places := map[string]string{}
	for i := range e.blocks {
		for _, op := range e.blocks[i].Ops {
			if op.Kind == OpBorrow {
				places[op.Loan] = op.Place
			}
		}
	}

	liveIn := make(map[string]map[string]bool, len(e.blocks))
	for i := range e.blocks {
		liveIn[e.blocks[i].ID] = map[string]bool{}
	}
	for round := 0; ; round++ {
		if round > len(e.blocks)+2 {
			return OracleResult{}, errors.New("oracle liveness did not converge")
		}
		changed := false
		for i := range e.blocks {
			block := &e.blocks[i]
			liveOut := map[string]bool{}
			for _, successor := range block.Successors {
				if _, ok := liveIn[successor]; !ok {
					return OracleResult{}, fmt.Errorf("oracle: unresolved successor %q", successor)
				}
				for loan := range liveIn[successor] {
					liveOut[loan] = true
				}
			}
			next, _ := transfer(block.Ops, nil, liveOut)
			if !sameSet(liveIn[block.ID], next) {
				liveIn[block.ID] = next
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	sourcePlaces := map[string]map[string]string{}
	for i := range p.Functions {
		sourcePlaces[p.Functions[i].ID] = loanPlaces(&p.Functions[i])
	}

	result := OracleResult{Local: LocalLiveness{}, ExpandedOps: e.ops, Covered: map[string]bool{}}
	seen := map[string]bool{}
	union := map[string]map[string]bool{}
	for i := range e.blocks {
		block := &e.blocks[i]
		liveOut := map[string]bool{}
		for _, successor := range block.Successors {
			for loan := range liveIn[successor] {
				liveOut[loan] = true
			}
		}
		live := liveOut
		for j := len(block.Ops) - 1; j >= 0; j-- {
			op := block.Ops[j]
			ctx, sourceOp := splitOpID(op.ID)
			fnID := e.instanceOf[ctx]
			result.Covered[fnID] = true
			inverse := e.unsubst[ctx]
			born := sourcePlaces[fnID]

			key := fnID + "|" + sourceOp
			if union[key] == nil {
				union[key] = map[string]bool{}
			}
			for loan := range live {
				local, ok := inverse[loan]
				if !ok {
					continue
				}
				if _, isBorn := born[local]; !isBorn {
					continue
				}
				union[key][local] = true
			}
			if op.Kind == OpMove {
				for loan := range live {
					if places[loan] != op.Place {
						continue
					}
					local, ok := inverse[loan]
					if !ok {
						continue
					}
					conflict := Conflict{FunctionID: fnID, OpID: sourceOp, LoanID: local}
					if !seen[conflict.key()] {
						seen[conflict.key()] = true
						result.Conflicts = append(result.Conflicts, conflict)
					}
				}
			}
			live, _ = transfer(block.Ops[j:j+1], nil, live)
		}
	}
	for key, loans := range union {
		sorted := make([]string, 0, len(loans))
		for loan := range loans {
			sorted = append(sorted, loan)
		}
		sort.Strings(sorted)
		result.Local[key] = sorted
	}
	sort.Slice(result.Conflicts, func(i, j int) bool {
		return result.Conflicts[i].key() < result.Conflicts[j].key()
	})
	return result, nil
}

func splitOpID(expanded string) (string, string) {
	index := strings.LastIndex(expanded, "|")
	if index < 0 {
		return "", expanded
	}
	return expanded[:index], expanded[index+1:]
}
