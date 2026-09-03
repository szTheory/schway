package ownership

import "sort"

// NormalizeOracle derives linear loan endpoints independently from Normalize.
// Keeping a second implementation makes checker/oracle agreement meaningful
// for last-use insertion instead of routing both through one algorithm.
func NormalizeOracle(program Program) Program {
	operations := make([]Operation, len(program.Operations))
	for index, operation := range program.Operations {
		operation.SourceIndex = index
		operations[index] = operation
	}

	endsAt := map[int][]string{}
	seenLoans := map[string]bool{}
	for start, operation := range operations {
		if operation.Kind != BorrowShared && operation.Kind != BorrowExclusive {
			continue
		}
		if seenLoans[operation.Loan] {
			continue
		}
		seenLoans[operation.Loan] = true

		lastUse := start
		hasExplicitEnd := false
		for cursor := start + 1; cursor < len(operations); cursor++ {
			candidate := operations[cursor]
			if candidate.Loan != operation.Loan {
				continue
			}
			switch candidate.Kind {
			case ReadLoan, MutateLoan:
				lastUse = cursor
			case EndLoan:
				hasExplicitEnd = true
			}
			if hasExplicitEnd {
				break
			}
		}
		if !hasExplicitEnd {
			endsAt[lastUse] = append(endsAt[lastUse], operation.Loan)
		}
	}

	normalized := Program{ID: program.ID}
	for index, operation := range operations {
		normalized.Operations = append(normalized.Operations, operation)
		loans := endsAt[index]
		sort.Strings(loans)
		for _, loan := range loans {
			normalized.Operations = append(normalized.Operations, Operation{
				Kind: EndLoan, Loan: loan, SourceIndex: index, Synthetic: true,
			})
		}
	}
	return normalized
}
