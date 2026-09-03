package cfg

import (
	"fmt"

	"ai-lang/ownership-kernel-workbench/ownership"
)

func GenerateDiamondPrograms() []Program {
	armChoices := [][]ownership.Operation{
		{},
		{{Kind: ownership.ReadLoan, Loan: "view"}},
		{{Kind: ownership.ReadPlace, Place: "p"}},
		{{Kind: ownership.MutatePlace, Place: "p"}},
		{{Kind: ownership.Move, Place: "p", Target: "q"}},
	}
	suffixChoices := [][]ownership.Operation{
		{},
		{{Kind: ownership.ReadLoan, Loan: "view"}},
		{{Kind: ownership.MutatePlace, Place: "p"}},
		{{Kind: ownership.Move, Place: "p", Target: "q"}},
	}
	programs := []Program{}
	for leftIndex, left := range armChoices {
		for rightIndex, right := range armChoices {
			for suffixIndex, suffix := range suffixChoices {
				programs = append(programs, Program{
					ID:    fmt.Sprintf("GEN-DIAMOND-%d-%d-%d", leftIndex, rightIndex, suffixIndex),
					Entry: "entry",
					Blocks: []Block{
						{ID: "entry", Operations: []ownership.Operation{
							{Kind: ownership.Declare, Place: "p"},
							{Kind: ownership.BorrowShared, Place: "p", Loan: "view"},
						}, Successors: []string{"left", "right"}},
						{ID: "left", Operations: cloneOperations(left), Successors: []string{"join"}},
						{ID: "right", Operations: cloneOperations(right), Successors: []string{"join"}},
						{ID: "join", Operations: cloneOperations(suffix)},
					},
				})
			}
		}
	}
	return programs
}

func FindMismatch(options Options) SearchResult {
	result := SearchResult{}
	for _, program := range GenerateDiamondPrograms() {
		comparison := Compare(program, options, 3)
		result.Programs++
		if !comparison.Agreement {
			copy := program
			result.Mismatch = &comparison
			result.MinimalInput = &copy
			return result
		}
	}
	return result
}

func LargeDiamondChain(count int) Program {
	if count < 1 {
		count = 1
	}
	blocks := []Block{{
		ID: "b0000-entry",
		Operations: []ownership.Operation{
			{Kind: ownership.Declare, Place: "p"},
			{Kind: ownership.BorrowShared, Place: "p", Loan: "view"},
		},
		Successors: []string{"b0001-left", "b0001-right"},
	}}
	for index := 1; index <= count; index++ {
		left := fmt.Sprintf("b%04d-left", index)
		right := fmt.Sprintf("b%04d-right", index)
		join := fmt.Sprintf("b%04d-join", index)
		blocks = append(blocks,
			Block{ID: left, Operations: []ownership.Operation{{Kind: ownership.ReadPlace, Place: "p"}}, Successors: []string{join}},
			Block{ID: right, Successors: []string{join}},
		)
		successors := []string{}
		operations := []ownership.Operation{}
		if index == count {
			operations = append(operations,
				ownership.Operation{Kind: ownership.ReadLoan, Loan: "view"},
				ownership.Operation{Kind: ownership.MutatePlace, Place: "p"},
			)
		} else {
			successors = []string{fmt.Sprintf("b%04d-left", index+1), fmt.Sprintf("b%04d-right", index+1)}
		}
		blocks = append(blocks, Block{ID: join, Operations: operations, Successors: successors})
	}
	return Program{ID: fmt.Sprintf("LARGE-DIAMOND-%d", count), Entry: "b0000-entry", Blocks: blocks}
}

func cloneOperations(source []ownership.Operation) []ownership.Operation {
	return append([]ownership.Operation{}, source...)
}
