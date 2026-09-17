// Package iplive is the S-006 workbench: a throwaway model of summary-based
// interprocedural loan liveness, built only to answer one cost question.
//
// It deliberately mirrors the shapes the real checker already has
// (internal/compiler/check's backward worklist over a per-function CFG, and
// internal/compiler/callgraph's acyclic order) without importing them: the
// gate is about how cost scales with call-graph size and shape, not about
// reproducing the checker's exact diagnostics.
package iplive

import (
	"fmt"
	"sort"
)

// OpKind is the workbench's minimal linear-operation alphabet. It is a
// deliberate subset of core.LinearOperation: borrow/use/move give the
// move-while-borrowed law something to decide, and call is the only
// interprocedural edge.
type OpKind string

const (
	OpBorrow OpKind = "borrow"
	OpUse    OpKind = "use"
	OpMove   OpKind = "move"
	OpCall   OpKind = "call"
)

// Operation is one linear operation with a stable identity.
type Operation struct {
	ID     string `json:"id"`
	Kind   OpKind `json:"kind"`
	Loan   string `json:"loan,omitempty"`   // borrow: created; use: used; call: argument loan
	Place  string `json:"place,omitempty"`  // borrow/move: the owner place
	Callee string `json:"callee,omitempty"` // call: callee function ID
	Result string `json:"result,omitempty"` // call: loan name bound to the call's result
}

// Block is one CFG block: stable ID, its own operations in program order, and
// its successor block IDs.
type Block struct {
	ID         string      `json:"id"`
	Ops        []Operation `json:"ops"`
	Successors []string    `json:"successors,omitempty"`
}

// Function is one function body. Blocks[0] is the entry block, and the slice
// order is topological (a block never precedes its own predecessors) so the
// oracle's single-pass expansion can resolve call results by walking it once.
type Function struct {
	ID        string  `json:"id"`
	ParamLoan string  `json:"param_loan"`
	Blocks    []Block `json:"blocks"`
	// Returns names the loan this function hands back, or "" when the return
	// value is owned and unrelated to any loan.
	Returns string `json:"returns,omitempty"`
}

// Program is a whole compilation unit plus the roots the oracle expands from.
type Program struct {
	Functions []Function `json:"functions"`
	Roots     []string   `json:"roots"`
}

func (p *Program) index() map[string]*Function {
	m := make(map[string]*Function, len(p.Functions))
	for i := range p.Functions {
		m[p.Functions[i].ID] = &p.Functions[i]
	}
	return m
}

func (f *Function) callees() []string {
	var out []string
	for _, block := range f.Blocks {
		for _, op := range block.Ops {
			if op.Kind == OpCall {
				out = append(out, op.Callee)
			}
		}
	}
	return out
}

// OpCount is the whole-program operation count -- the size axis an
// intraprocedural-only analyzer would be linear in.
func (p *Program) OpCount() int {
	total := 0
	for i := range p.Functions {
		for _, block := range p.Functions[i].Blocks {
			total += len(block.Ops)
		}
	}
	return total
}

// CallEdgeCount counts call-graph edges with multiplicity (one per call site).
func (p *Program) CallEdgeCount() int {
	total := 0
	for i := range p.Functions {
		total += len(p.Functions[i].callees())
	}
	return total
}

// ReversePostorder returns a callee-before-caller order over the call graph,
// refusing a cycle fail-closed the way callgraph.Order does. The traversal is
// an explicit-stack white/gray/black DFS -- no native recursion, so a deep
// chain cannot exhaust the Go stack.
func (p *Program) ReversePostorder() ([]string, error) {
	byID := p.index()
	const (
		white = 0
		gray  = 1
		black = 2
	)
	state := make(map[string]int, len(p.Functions))
	var order []string

	type frame struct {
		id   string
		next int
	}
	roots := make([]string, 0, len(p.Functions))
	for i := range p.Functions {
		roots = append(roots, p.Functions[i].ID)
	}
	sort.Strings(roots)

	for _, root := range roots {
		if state[root] != white {
			continue
		}
		stack := []frame{{id: root}}
		state[root] = gray
		for len(stack) > 0 {
			top := &stack[len(stack)-1]
			fn, ok := byID[top.id]
			if !ok {
				return nil, fmt.Errorf("unresolved callee %q", top.id)
			}
			callees := fn.callees()
			if top.next < len(callees) {
				callee := callees[top.next]
				top.next++
				switch state[callee] {
				case gray:
					return nil, fmt.Errorf("call_graph_cycle: %q reaches itself", callee)
				case black:
					continue
				}
				if _, ok := byID[callee]; !ok {
					return nil, fmt.Errorf("unresolved callee %q", callee)
				}
				state[callee] = gray
				stack = append(stack, frame{id: callee})
				continue
			}
			state[top.id] = black
			order = append(order, top.id)
			stack = stack[:len(stack)-1]
		}
	}
	return order, nil
}
