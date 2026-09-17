package iplive

import "sort"

// Agreement is one mechanism's comparison against the expansion oracle.
type Agreement struct {
	Mechanism      string   `json:"mechanism"`
	ConflictDiffs  []string `json:"conflict_diffs,omitempty"`
	LivenessDiffs  []string `json:"liveness_diffs,omitempty"`
	OracleConflict int      `json:"oracle_conflicts"`
	ExpandedOps    int      `json:"oracle_expanded_ops"`
	Agreed         bool     `json:"agreed"`
}

// CheckAgainstOracle analyzes p with the given provider and compares the
// answer to the context-sensitive expansion oracle, restricted to the
// functions the oracle actually reached from the program's roots.
func CheckAgainstOracle(p *Program, provider Provider, oracleBudget int) (Agreement, error) {
	analysis, err := Analyze(p, provider)
	if err != nil {
		return Agreement{Mechanism: provider.Name()}, err
	}
	oracle, err := Oracle(p, oracleBudget)
	if err != nil {
		return Agreement{Mechanism: provider.Name()}, err
	}

	var covered []Conflict
	for _, conflict := range analysis.Conflicts {
		if oracle.Covered[conflict.FunctionID] {
			covered = append(covered, conflict)
		}
	}
	local := LocalLiveness{}
	for key, loans := range analysis.Local {
		local[key] = loans
	}

	agreement := Agreement{
		Mechanism:      provider.Name(),
		ConflictDiffs:  DiffConflicts(covered, oracle.Conflicts),
		LivenessDiffs:  DiffLocal(local, oracle.Local, oracle.Covered),
		OracleConflict: len(oracle.Conflicts),
		ExpandedOps:    oracle.ExpandedOps,
	}
	sort.Strings(agreement.ConflictDiffs)
	agreement.Agreed = len(agreement.ConflictDiffs) == 0 && len(agreement.LivenessDiffs) == 0
	return agreement, nil
}
