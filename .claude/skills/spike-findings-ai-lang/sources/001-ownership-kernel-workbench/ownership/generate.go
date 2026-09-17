package ownership

type SearchResult struct {
	Programs     int         `json:"programs"`
	MaxDepth     int         `json:"max_depth"`
	Mismatch     *Comparison `json:"mismatch,omitempty"`
	MinimalInput *Program    `json:"minimal_input,omitempty"`
}

// FindMismatch enumerates programs breadth-first, so the first disagreement is
// minimal by operation count within this deliberately small alphabet.
func FindMismatch(maxDepth int, options CheckerOptions) SearchResult {
	result := SearchResult{MaxDepth: maxDepth}
	for _, program := range GeneratePrograms(maxDepth) {
		comparison := Compare(program, options)
		result.Programs++
		if !comparison.Agreement {
			result.Mismatch = &comparison
			result.MinimalInput = &program
			return result
		}
	}
	return result
}

// GeneratePrograms returns the bounded breadth-first corpus used by the
// differential runner. It is exported so additional independent passes can
// consume precisely the same ordered inputs without sharing transitions.
func GeneratePrograms(maxDepth int) []Program {
	alphabet := generatedAlphabet()
	programs := []Program{}
	for depth := 0; depth <= maxDepth; depth++ {
		indices := make([]int, depth)
		for {
			operations := []Operation{{Kind: Declare, Place: "p", Resource: true}}
			for _, index := range indices {
				operations = append(operations, alphabet[index])
			}
			programs = append(programs, Program{ID: "generated", Operations: operations})
			if depth == 0 || !increment(indices, len(alphabet)) {
				break
			}
		}
	}
	return programs
}

func generatedAlphabet() []Operation {
	return []Operation{
		{Kind: BorrowShared, Place: "p", Loan: "l"},
		{Kind: BorrowExclusive, Place: "p", Loan: "l"},
		{Kind: ReadPlace, Place: "p"},
		{Kind: MutatePlace, Place: "p"},
		{Kind: ReadLoan, Loan: "l"},
		{Kind: MutateLoan, Loan: "l"},
		{Kind: EndLoan, Loan: "l"},
		{Kind: Move, Place: "p", Target: "q"},
		{Kind: Release, Place: "p"},
	}
}

func increment(indices []int, base int) bool {
	for index := len(indices) - 1; index >= 0; index-- {
		indices[index]++
		if indices[index] < base {
			return true
		}
		indices[index] = 0
	}
	return false
}
