package producer

import (
	"fmt"
	"sort"

	"example.com/ai-lang/spike004/format"
)

// Build deliberately implements certificate production independently from the verifier.
func Build(artifact format.Artifact, mode string) (format.Certificate, error) {
	if mode != "recompute" && mode != "replay" {
		return format.Certificate{}, fmt.Errorf("unknown certificate mode %q", mode)
	}
	certificate := format.Certificate{
		Schema:         format.SchemaVersion,
		Mode:           mode,
		ArtifactDigest: format.ArtifactDigest(artifact),
	}
	for _, function := range artifact.Functions {
		certificate.Summaries = append(certificate.Summaries, format.SummaryClaim{
			Function: function.ID,
			Summary:  format.CanonicalSummary(function.Published),
		})
	}
	registry := map[string]format.TypeRule{}
	for _, rule := range artifact.Types {
		registry[rule.Name] = rule
	}
	for _, request := range artifact.AbilityRequests {
		abilities, err := abilities(registry, request, map[string]bool{})
		if err != nil {
			return format.Certificate{}, err
		}
		certificate.Abilities = append(certificate.Abilities, format.AbilityClaim{Type: request, Abilities: abilities})
	}
	for _, trace := range artifact.Traces {
		claim, err := execute(trace, mode == "replay")
		if err != nil {
			return format.Certificate{}, err
		}
		certificate.Traces = append(certificate.Traces, claim)
	}
	sort.Slice(certificate.Summaries, func(i, j int) bool { return certificate.Summaries[i].Function < certificate.Summaries[j].Function })
	sort.Slice(certificate.Abilities, func(i, j int) bool {
		return format.TypeKey(certificate.Abilities[i].Type) < format.TypeKey(certificate.Abilities[j].Type)
	})
	sort.Slice(certificate.Traces, func(i, j int) bool { return certificate.Traces[i].Trace < certificate.Traces[j].Trace })
	return certificate, nil
}

func abilities(registry map[string]format.TypeRule, expr format.TypeExpr, visiting map[string]bool) ([]string, error) {
	rule, ok := registry[expr.Name]
	if !ok {
		return nil, fmt.Errorf("producer: unknown type %s", expr.Name)
	}
	if len(expr.Args) != rule.Parameters {
		return nil, fmt.Errorf("producer: %s expects %d arguments", expr.Name, rule.Parameters)
	}
	key := format.TypeKey(expr)
	if visiting[key] {
		return nil, fmt.Errorf("producer: recursive ability rule %s", key)
	}
	visiting[key] = true
	defer delete(visiting, key)
	result := append([]string(nil), rule.Base...)
	for ability, indexes := range rule.Conditional {
		ok := true
		for _, index := range indexes {
			if index < 0 || index >= len(expr.Args) {
				return nil, fmt.Errorf("producer: invalid parameter %d for %s", index, expr.Name)
			}
			child, err := abilities(registry, expr.Args[index], visiting)
			if err != nil {
				return nil, err
			}
			if !contains(child, ability) {
				ok = false
			}
		}
		if ok {
			result = append(result, ability)
		}
	}
	return format.CanonicalAbilities(result), nil
}

type state struct {
	ownership map[string]string
	loans     map[string]format.LoanState
}

func execute(trace format.Trace, replay bool) (format.TraceClaim, error) {
	current := state{ownership: map[string]string{}, loans: map[string]format.LoanState{}}
	for _, place := range trace.Places {
		current.ownership[place] = "owned"
	}
	claim := format.TraceClaim{Trace: trace.ID}
	for _, event := range trace.Events {
		if err := transition(&current, event); err != nil {
			return format.TraceClaim{}, fmt.Errorf("producer: trace %s event %s: %w", trace.ID, event.ID, err)
		}
		if replay {
			claim.Snapshots = append(claim.Snapshots, snapshot(current, event.ID))
		}
	}
	claim.FinalPlaces = snapshot(current, "").Places
	return claim, nil
}

func transition(current *state, event format.Event) error {
	ownership, ok := current.ownership[event.Place]
	if !ok {
		return fmt.Errorf("unknown place %s", event.Place)
	}
	switch event.Op {
	case "borrow_shared", "borrow_exclusive":
		if ownership != "owned" {
			return fmt.Errorf("cannot borrow %s place", ownership)
		}
		for _, loan := range current.loans {
			if loan.Place == event.Place && (event.Op == "borrow_exclusive" || loan.Access == "exclusive") {
				return fmt.Errorf("borrow conflict")
			}
		}
		if event.Loan == "" || current.loans[event.Loan].ID != "" {
			return fmt.Errorf("invalid loan identity")
		}
		access := "shared"
		if event.Op == "borrow_exclusive" {
			access = "exclusive"
		}
		current.loans[event.Loan] = format.LoanState{ID: event.Loan, Place: event.Place, Access: access}
	case "end_loan":
		loan, ok := current.loans[event.Loan]
		if !ok || loan.Place != event.Place {
			return fmt.Errorf("unknown loan %s", event.Loan)
		}
		delete(current.loans, event.Loan)
	case "use":
		if ownership != "owned" {
			return fmt.Errorf("use of %s place", ownership)
		}
	case "move", "destroy":
		if ownership != "owned" {
			return fmt.Errorf("cannot %s %s place", event.Op, ownership)
		}
		for _, loan := range current.loans {
			if loan.Place == event.Place {
				return fmt.Errorf("active loan %s", loan.ID)
			}
		}
		if event.Op == "move" {
			current.ownership[event.Place] = "moved"
		} else {
			current.ownership[event.Place] = "destroyed"
		}
	default:
		return fmt.Errorf("unknown operation %s", event.Op)
	}
	return nil
}

func snapshot(current state, event string) format.Snapshot {
	result := format.Snapshot{EventID: event}
	for place, ownership := range current.ownership {
		result.Places = append(result.Places, format.PlaceState{Place: place, Ownership: ownership})
	}
	for _, loan := range current.loans {
		result.Loans = append(result.Loans, loan)
	}
	sort.Slice(result.Places, func(i, j int) bool { return result.Places[i].Place < result.Places[j].Place })
	sort.Slice(result.Loans, func(i, j int) bool { return result.Loans[i].ID < result.Loans[j].ID })
	return result
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
