package cfg

import (
	"sort"
	"strconv"

	"ai-lang/ownership-kernel-workbench/ownership"
)

type loanSet map[string]struct{}

func Analyze(program Program, options Options) Analysis {
	result := Analysis{
		Program: program.ID, Valid: true,
		LiveIn: map[string][]string{}, LiveOut: map[string][]string{},
		Endpoints: []Endpoint{},
	}
	blocks, order, predecessors, diagnostic := validate(program)
	if diagnostic != nil {
		result.Valid = false
		result.Diagnostic = diagnostic
		return result
	}
	result.Blocks = len(blocks)
	loanDefinitions := map[string]struct{}{}
	for _, id := range order {
		block := blocks[id]
		result.Edges += len(block.Successors)
		for _, operation := range block.Operations {
			if operation.Kind == ownership.BorrowShared || operation.Kind == ownership.BorrowExclusive {
				loanDefinitions[operation.Loan] = struct{}{}
			}
		}
	}
	result.Loans = len(loanDefinitions)

	liveIn := map[string]loanSet{}
	liveOut := map[string]loanSet{}
	queue := reversePostorderForBackward(program.Entry, blocks)
	inQueue := map[string]bool{}
	for _, id := range queue {
		inQueue[id] = true
	}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		inQueue[id] = false
		result.TransferEvaluations++
		out := loanSet{}
		for _, successor := range blocks[id].Successors {
			unionInto(out, liveIn[successor])
		}
		in := transferBackward(blocks[id].Operations, out)
		if !setsEqual(out, liveOut[id]) || !setsEqual(in, liveIn[id]) {
			liveOut[id] = out
			liveIn[id] = in
			for _, predecessor := range predecessors[id] {
				if !inQueue[predecessor] {
					queue = append(queue, predecessor)
					inQueue[predecessor] = true
					result.WorklistReinsertions++
				}
			}
		}
	}

	for _, id := range order {
		result.LiveIn[id] = sortedSet(liveIn[id])
		result.LiveOut[id] = sortedSet(liveOut[id])
	}
	result.Endpoints = deriveEndpoints(order, blocks, liveIn, liveOut, options)
	return result
}

func validate(program Program) (map[string]Block, []string, map[string][]string, *Diagnostic) {
	blocks := map[string]Block{}
	order := make([]string, 0, len(program.Blocks))
	loans := map[string]string{}
	for _, block := range program.Blocks {
		if block.ID == "" {
			return nil, nil, nil, &Diagnostic{Code: "cfg.block_id_required", Program: program.ID, Message: "every block requires a stable identity"}
		}
		if _, exists := blocks[block.ID]; exists {
			return nil, nil, nil, &Diagnostic{Code: "cfg.duplicate_block", Program: program.ID, Block: block.ID, Message: "block identity is duplicated"}
		}
		for _, operation := range block.Operations {
			if operation.Kind == ownership.EndLoan {
				return nil, nil, nil, &Diagnostic{Code: "cfg.explicit_end_out_of_scope", Program: program.ID, Block: block.ID, Message: "this spike isolates inferred loan endings"}
			}
			if operation.Kind == ownership.BorrowShared || operation.Kind == ownership.BorrowExclusive {
				if operation.Loan == "" {
					return nil, nil, nil, &Diagnostic{Code: "cfg.loan_id_required", Program: program.ID, Block: block.ID, Message: "borrow requires a stable loan identity"}
				}
				if first, exists := loans[operation.Loan]; exists {
					return nil, nil, nil, &Diagnostic{Code: "cfg.duplicate_loan", Program: program.ID, Block: block.ID, Message: "loan identity was already defined in block " + first}
				}
				loans[operation.Loan] = block.ID
			}
		}
		blocks[block.ID] = block
		order = append(order, block.ID)
	}
	if _, exists := blocks[program.Entry]; !exists {
		return nil, nil, nil, &Diagnostic{Code: "cfg.unknown_entry", Program: program.ID, Block: program.Entry, Message: "entry block does not exist"}
	}
	predecessors := map[string][]string{}
	for _, id := range order {
		for _, successor := range blocks[id].Successors {
			if _, exists := blocks[successor]; !exists {
				return nil, nil, nil, &Diagnostic{Code: "cfg.unknown_successor", Program: program.ID, Block: id, Message: "successor " + successor + " does not exist"}
			}
			predecessors[successor] = append(predecessors[successor], id)
		}
	}
	for id := range predecessors {
		sort.Strings(predecessors[id])
	}
	sort.Strings(order)
	return blocks, order, predecessors, nil
}

func transferBackward(operations []ownership.Operation, out loanSet) loanSet {
	live := cloneSet(out)
	for index := len(operations) - 1; index >= 0; index-- {
		operation := operations[index]
		switch operation.Kind {
		case ownership.ReadLoan, ownership.MutateLoan:
			live[operation.Loan] = struct{}{}
		case ownership.BorrowShared, ownership.BorrowExclusive:
			delete(live, operation.Loan)
		}
	}
	return live
}

func livenessAround(operations []ownership.Operation, out loanSet) []loanSet {
	after := make([]loanSet, len(operations))
	live := cloneSet(out)
	for index := len(operations) - 1; index >= 0; index-- {
		after[index] = cloneSet(live)
		operation := operations[index]
		switch operation.Kind {
		case ownership.ReadLoan, ownership.MutateLoan:
			live[operation.Loan] = struct{}{}
		case ownership.BorrowShared, ownership.BorrowExclusive:
			delete(live, operation.Loan)
		}
	}
	return after
}

func deriveEndpoints(order []string, blocks map[string]Block, liveIn, liveOut map[string]loanSet, options Options) []Endpoint {
	endpoints := []Endpoint{}
	for _, id := range order {
		block := blocks[id]
		after := livenessAround(block.Operations, liveOut[id])
		for index, operation := range block.Operations {
			loan := ""
			switch operation.Kind {
			case ownership.BorrowShared, ownership.BorrowExclusive, ownership.ReadLoan, ownership.MutateLoan:
				loan = operation.Loan
			}
			if loan != "" {
				if _, remainsLive := after[index][loan]; !remainsLive {
					endpoints = append(endpoints, Endpoint{
						ID:   "point:" + id + ":" + strconv.Itoa(index) + ":" + loan,
						Kind: "point", Loan: loan, Block: id, AfterOperation: index,
					})
				}
			}
		}
		if options.OmitEdgeEnds {
			continue
		}
		for _, successor := range block.Successors {
			for _, loan := range sortedSet(liveOut[id]) {
				if _, needed := liveIn[successor][loan]; needed {
					continue
				}
				endpoints = append(endpoints, Endpoint{
					ID:   "edge:" + id + ":" + successor + ":" + loan,
					Kind: "edge", Loan: loan, From: id, To: successor,
				})
			}
		}
	}
	sort.Slice(endpoints, func(i, j int) bool { return endpoints[i].ID < endpoints[j].ID })
	return endpoints
}

func reversePostorderForBackward(entry string, blocks map[string]Block) []string {
	visited := map[string]bool{}
	postorder := []string{}
	var visit func(string)
	visit = func(id string) {
		if visited[id] {
			return
		}
		visited[id] = true
		successors := append([]string{}, blocks[id].Successors...)
		sort.Strings(successors)
		for _, successor := range successors {
			visit(successor)
		}
		postorder = append(postorder, id)
	}
	visit(entry)
	unreachable := []string{}
	for id := range blocks {
		if !visited[id] {
			unreachable = append(unreachable, id)
		}
	}
	sort.Strings(unreachable)
	for _, id := range unreachable {
		visit(id)
	}
	return postorder
}

func cloneSet(source loanSet) loanSet {
	result := loanSet{}
	for value := range source {
		result[value] = struct{}{}
	}
	return result
}

func unionInto(target, source loanSet) {
	for value := range source {
		target[value] = struct{}{}
	}
}

func setsEqual(left, right loanSet) bool {
	if len(left) != len(right) {
		return false
	}
	for value := range left {
		if _, exists := right[value]; !exists {
			return false
		}
	}
	return true
}

func sortedSet(set loanSet) []string {
	values := make([]string, 0, len(set))
	for value := range set {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}
