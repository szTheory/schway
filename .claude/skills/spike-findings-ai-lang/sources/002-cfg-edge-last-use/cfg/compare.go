package cfg

import (
	"reflect"
	"sort"

	"ai-lang/ownership-kernel-workbench/ownership"
)

type boundedPath struct {
	blocks []string
}

func Compare(program Program, options Options, maxBlockVisits int) Comparison {
	analysis := Analyze(program, options)
	comparison := Comparison{
		Program: program.ID, Agreement: analysis.Valid,
		AnalyzerValid: analysis.Valid, OracleValid: analysis.Valid,
		Analysis: analysis, Paths: []PathComparison{},
	}
	if !analysis.Valid {
		return comparison
	}
	blocks := map[string]Block{}
	for _, block := range program.Blocks {
		blocks[block.ID] = block
	}
	paths := expandPaths(program.Entry, blocks, maxBlockVisits)
	if len(paths) == 0 {
		comparison.Agreement = false
		comparison.AnalyzerValid = false
		comparison.OracleValid = false
		comparison.Analysis.Valid = false
		comparison.Analysis.Diagnostic = &Diagnostic{Code: "cfg.no_bounded_exit_path", Program: program.ID, Message: "bounded path oracle found no terminating path"}
		return comparison
	}
	allAnalyzerValid := true
	allOracleValid := true
	for index, path := range paths {
		pathComparison := comparePath(program.ID, index, path, blocks, analysis.Endpoints)
		comparison.Paths = append(comparison.Paths, pathComparison)
		comparison.Agreement = comparison.Agreement && pathComparison.Agreement
		allAnalyzerValid = allAnalyzerValid && pathComparison.Analyzer.Valid
		allOracleValid = allOracleValid && pathComparison.Oracle.Valid
	}
	comparison.AnalyzerValid = allAnalyzerValid
	comparison.OracleValid = allOracleValid
	comparison.Agreement = comparison.Agreement && allAnalyzerValid == allOracleValid
	return comparison
}

func comparePath(programID string, pathIndex int, path boundedPath, blocks map[string]Block, endpoints []Endpoint) PathComparison {
	original := ownership.Program{ID: programID + "/oracle-path-" + itoa(pathIndex)}
	materialized := ownership.Program{ID: programID + "/analyzer-path-" + itoa(pathIndex)}
	borrowed := map[string]bool{}
	inserted := map[string]bool{}
	insertedIDs := []string{}

	for blockIndex, blockID := range path.blocks {
		block := blocks[blockID]
		for operationIndex, operation := range block.Operations {
			original.Operations = append(original.Operations, operation)
			materialized.Operations = append(materialized.Operations, operation)
			if operation.Kind == ownership.BorrowShared || operation.Kind == ownership.BorrowExclusive {
				borrowed[operation.Loan] = true
			}
			for _, endpoint := range endpoints {
				if endpoint.Kind == "point" && endpoint.Block == blockID && endpoint.AfterOperation == operationIndex {
					materialized.Operations = append(materialized.Operations, ownership.Operation{Kind: ownership.EndLoan, Loan: endpoint.Loan, Synthetic: true})
					inserted[endpoint.Loan] = true
					insertedIDs = append(insertedIDs, endpoint.ID)
				}
			}
		}
		if blockIndex+1 < len(path.blocks) {
			next := path.blocks[blockIndex+1]
			for _, endpoint := range endpoints {
				if endpoint.Kind == "edge" && endpoint.From == blockID && endpoint.To == next {
					materialized.Operations = append(materialized.Operations, ownership.Operation{Kind: ownership.EndLoan, Loan: endpoint.Loan, Synthetic: true})
					inserted[endpoint.Loan] = true
					insertedIDs = append(insertedIDs, endpoint.ID)
				}
			}
		}
	}

	// A terminal guard is intentionally late. It suppresses the imported
	// linear normalizer without rescuing a missing earlier CFG endpoint, so an
	// owner conflict still exposes the analyzer defect.
	guards := []string{}
	loanNames := make([]string, 0, len(borrowed))
	for loan := range borrowed {
		loanNames = append(loanNames, loan)
	}
	sort.Strings(loanNames)
	for _, loan := range loanNames {
		if !inserted[loan] {
			materialized.Operations = append(materialized.Operations, ownership.Operation{Kind: ownership.EndLoan, Loan: loan, Synthetic: true})
			guards = append(guards, loan)
		}
	}

	analyzer := ownership.Check(materialized, ownership.CheckerOptions{})
	oracle := ownership.Check(original, ownership.CheckerOptions{})
	agreement := analyzer.Valid == oracle.Valid
	if !analyzer.Valid && !oracle.Valid {
		agreement = agreement && analyzer.Diagnostic.Code == oracle.Diagnostic.Code
	}
	oracleEnds := oracleEndpointIDs(original)
	return PathComparison{
		ID: programID + "/path-" + itoa(pathIndex), Blocks: append([]string{}, path.blocks...),
		Agreement: agreement, Analyzer: analyzer, Oracle: oracle,
		InsertedEnds: insertedIDs, OracleEnds: oracleEnds, TerminalGuards: guards,
	}
}

func oracleEndpointIDs(program ownership.Program) []string {
	normalized := ownership.NormalizeOracle(program)
	ids := []string{}
	for _, operation := range normalized.Operations {
		if operation.Kind == ownership.EndLoan && operation.Synthetic {
			ids = append(ids, "linear:"+itoa(operation.SourceIndex)+":"+operation.Loan)
		}
	}
	return ids
}

func expandPaths(entry string, blocks map[string]Block, maxBlockVisits int) []boundedPath {
	if maxBlockVisits < 1 {
		maxBlockVisits = 1
	}
	paths := []boundedPath{}
	visits := map[string]int{}
	var walk func(string, []string)
	walk = func(id string, prefix []string) {
		if visits[id] >= maxBlockVisits {
			return
		}
		visits[id]++
		current := append(append([]string{}, prefix...), id)
		successors := append([]string{}, blocks[id].Successors...)
		sort.Strings(successors)
		if len(successors) == 0 {
			paths = append(paths, boundedPath{blocks: current})
		} else {
			for _, successor := range successors {
				walk(successor, current)
			}
		}
		visits[id]--
	}
	walk(entry, nil)
	return paths
}

func EndpointShape(analysis Analysis) []string {
	shape := make([]string, len(analysis.Endpoints))
	for index, endpoint := range analysis.Endpoints {
		shape[index] = endpoint.ID
	}
	sort.Strings(shape)
	return shape
}

func Equivalent(left, right Comparison) bool {
	return left.Agreement == right.Agreement &&
		left.AnalyzerValid == right.AnalyzerValid &&
		left.OracleValid == right.OracleValid &&
		reflect.DeepEqual(EndpointShape(left.Analysis), EndpointShape(right.Analysis))
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	result := ""
	for value > 0 {
		result = string(rune('0'+value%10)) + result
		value /= 10
	}
	return result
}
