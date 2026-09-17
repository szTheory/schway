package ownership

import "sort"

// Normalize assigns source indices and inserts an EndLoan after the final use
// of a loan when the source did not provide an explicit end. This is the first
// deliberately narrow non-lexical-lifetime experiment: linear programs only.
func Normalize(program Program) Program {
	ops := make([]Operation, len(program.Operations))
	copy(ops, program.Operations)

	explicitEndAt := map[string][]int{}
	borrowAt := map[string]int{}
	lastUse := map[string]int{}
	for index := range ops {
		ops[index].SourceIndex = index
		operation := ops[index]
		switch operation.Kind {
		case BorrowShared, BorrowExclusive:
			if _, exists := borrowAt[operation.Loan]; !exists {
				borrowAt[operation.Loan] = index
				lastUse[operation.Loan] = index
			}
		case ReadLoan, MutateLoan:
			lastUse[operation.Loan] = index
		case EndLoan:
			explicitEndAt[operation.Loan] = append(explicitEndAt[operation.Loan], index)
		}
	}

	endsAt := map[int][]string{}
	for loan, start := range borrowAt {
		hasLaterExplicitEnd := false
		for _, end := range explicitEndAt[loan] {
			if end > start {
				hasLaterExplicitEnd = true
				break
			}
		}
		if hasLaterExplicitEnd {
			continue
		}
		end := lastUse[loan]
		if end < start {
			end = start
		}
		endsAt[end] = append(endsAt[end], loan)
	}

	normalized := Program{ID: program.ID}
	for index, operation := range ops {
		normalized.Operations = append(normalized.Operations, operation)
		loans := endsAt[index]
		sort.Strings(loans)
		for _, loan := range loans {
			normalized.Operations = append(normalized.Operations, Operation{
				Kind:        EndLoan,
				Loan:        loan,
				SourceIndex: index,
				Synthetic:   true,
			})
		}
	}
	return normalized
}
