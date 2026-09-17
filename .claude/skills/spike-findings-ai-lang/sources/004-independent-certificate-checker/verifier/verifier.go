package verifier

import (
	"crypto/subtle"
	"fmt"
	"reflect"
	"sort"

	"example.com/ai-lang/spike004/format"
)

type Result struct {
	Valid       bool                `json:"valid"`
	Diagnostics []format.Diagnostic `json:"diagnostics,omitempty"`
	Checks      int                 `json:"checks"`
}

// Verify is body- and source-blind. It validates the declared typed-core facts
// and certificate without importing or calling the producer implementation.
func Verify(artifact format.Artifact, certificate format.Certificate) Result {
	checker := checker{artifact: artifact, certificate: certificate}
	checker.run()
	return Result{Valid: len(checker.diagnostics) == 0, Diagnostics: checker.diagnostics, Checks: checker.checks}
}

type checker struct {
	artifact    format.Artifact
	certificate format.Certificate
	diagnostics []format.Diagnostic
	checks      int
}

func (c *checker) run() {
	if c.artifact.Schema != format.SchemaVersion || c.certificate.Schema != format.SchemaVersion {
		c.add("format.schema", "artifact", "unsupported schema", "")
		return
	}
	if c.certificate.Mode != "recompute" && c.certificate.Mode != "replay" {
		c.add("certificate.mode", "certificate", "mode must be recompute or replay", "")
	}
	c.checkUniqueIDs()
	wantDigest := format.ArtifactDigest(c.artifact)
	c.checks++
	if subtle.ConstantTimeCompare([]byte(wantDigest), []byte(c.certificate.ArtifactDigest)) != 1 {
		c.add("certificate.subject_mismatch", c.artifact.Module, "certificate is stale or belongs to another artifact", "")
		return
	}
	c.checkSummaries()
	c.checkAbilities()
	c.checkTraces()
}

func (c *checker) checkUniqueIDs() {
	seen := map[string]string{}
	add := func(kind, id string) {
		c.checks++
		if id == "" {
			c.add("artifact.empty_id", kind, "identity must not be empty", "")
			return
		}
		if previous, exists := seen[kind+":"+id]; exists {
			c.add("artifact.duplicate_id", id, "duplicate "+kind+"; first seen as "+previous, "")
		} else {
			seen[kind+":"+id] = id
		}
	}
	for _, rule := range c.artifact.Types {
		add("type", rule.Name)
	}
	for _, function := range c.artifact.Functions {
		add("function", function.ID)
	}
	for _, trace := range c.artifact.Traces {
		add("trace", trace.ID)
		local := map[string]bool{}
		for _, place := range trace.Places {
			if local[place] {
				c.add("artifact.duplicate_id", trace.ID+"/"+place, "duplicate place", "")
			}
			local[place] = true
		}
		for _, event := range trace.Events {
			add("event", event.ID)
		}
	}
}

func (c *checker) checkSummaries() {
	claims := map[string]format.Summary{}
	for _, claim := range c.certificate.Summaries {
		if _, exists := claims[claim.Function]; exists {
			c.add("certificate.duplicate_claim", claim.Function, "duplicate summary claim", "")
		}
		claims[claim.Function] = format.CanonicalSummary(claim.Summary)
	}
	for _, function := range c.artifact.Functions {
		c.checks++
		fact := format.CanonicalSummary(function.ReturnFact)
		published := format.CanonicalSummary(function.Published)
		if !reflect.DeepEqual(fact, published) {
			c.add("artifact.summary_mismatch", function.ID, "published ownership summary does not equal typed-core return fact", "")
		}
		claim, ok := claims[function.ID]
		if !ok {
			c.add("certificate.missing_claim", function.ID, "summary claim missing", "")
			continue
		}
		if !reflect.DeepEqual(claim, published) {
			c.add("certificate.summary_mismatch", function.ID, "certificate summary does not equal published summary", "")
		}
		delete(claims, function.ID)
	}
	for id := range claims {
		c.add("certificate.unknown_claim", id, "summary claim has no function", "")
	}
}

func (c *checker) checkAbilities() {
	rules := map[string]format.TypeRule{}
	for _, rule := range c.artifact.Types {
		rules[rule.Name] = rule
		for ability, indexes := range rule.Conditional {
			if !knownAbility(ability) {
				c.add("artifact.unknown_ability", rule.Name, ability, "")
			}
			for _, index := range indexes {
				if index < 0 || index >= rule.Parameters {
					c.add("artifact.invalid_ability_rule", rule.Name, fmt.Sprintf("parameter %d outside arity %d", index, rule.Parameters), "")
				}
			}
		}
	}
	claims := map[string][]string{}
	for _, claim := range c.certificate.Abilities {
		key := format.TypeKey(claim.Type)
		claims[key] = format.CanonicalAbilities(claim.Abilities)
	}
	for _, request := range c.artifact.AbilityRequests {
		key := format.TypeKey(request)
		got, err := c.deriveAbilities(rules, request, map[string]bool{})
		c.checks++
		if err != nil {
			c.add("artifact.ability_derivation", key, err.Error(), "")
			continue
		}
		claim, ok := claims[key]
		if !ok {
			c.add("certificate.missing_claim", key, "ability claim missing", "")
			continue
		}
		if !reflect.DeepEqual(got, claim) {
			c.add("certificate.ability_mismatch", key, fmt.Sprintf("derived %v, certificate says %v", got, claim), "")
		}
		delete(claims, key)
	}
	for key := range claims {
		c.add("certificate.unknown_claim", key, "ability claim was not requested", "")
	}
}

func (c *checker) deriveAbilities(rules map[string]format.TypeRule, expr format.TypeExpr, visiting map[string]bool) ([]string, error) {
	rule, ok := rules[expr.Name]
	if !ok {
		return nil, fmt.Errorf("unknown type %s", expr.Name)
	}
	if len(expr.Args) != rule.Parameters {
		return nil, fmt.Errorf("%s expects %d type arguments, received %d", expr.Name, rule.Parameters, len(expr.Args))
	}
	key := format.TypeKey(expr)
	if visiting[key] {
		return nil, fmt.Errorf("recursive ability derivation at %s", key)
	}
	visiting[key] = true
	defer delete(visiting, key)
	set := map[string]bool{}
	for _, ability := range rule.Base {
		if !knownAbility(ability) {
			return nil, fmt.Errorf("unknown base ability %s", ability)
		}
		set[ability] = true
	}
	for ability, indexes := range rule.Conditional {
		if !knownAbility(ability) {
			continue
		}
		allowed := true
		for _, index := range indexes {
			if index < 0 || index >= len(expr.Args) {
				allowed = false
				continue
			}
			child, err := c.deriveAbilities(rules, expr.Args[index], visiting)
			if err != nil {
				return nil, err
			}
			if !member(child, ability) {
				allowed = false
			}
		}
		if allowed {
			set[ability] = true
		}
	}
	var result []string
	for _, ability := range format.AbilityOrder {
		if set[ability] {
			result = append(result, ability)
		}
	}
	return result, nil
}

type machine struct {
	places map[string]string
	loans  map[string]format.LoanState
}

func (c *checker) checkTraces() {
	claims := map[string]format.TraceClaim{}
	for _, claim := range c.certificate.Traces {
		if _, exists := claims[claim.Trace]; exists {
			c.add("certificate.duplicate_claim", claim.Trace, "duplicate trace claim", "")
		}
		claims[claim.Trace] = claim
	}
	for _, trace := range c.artifact.Traces {
		claim, ok := claims[trace.ID]
		if !ok {
			c.add("certificate.missing_claim", trace.ID, "trace claim missing", "")
			continue
		}
		current := machine{places: map[string]string{}, loans: map[string]format.LoanState{}}
		for _, place := range trace.Places {
			current.places[place] = "owned"
		}
		for index, event := range trace.Events {
			c.apply(&current, trace.ID, event)
			if c.certificate.Mode == "replay" {
				if index >= len(claim.Snapshots) {
					c.add("certificate.missing_snapshot", trace.ID, "snapshot missing", event.ID)
					continue
				}
				want := capture(current, event.ID)
				if !reflect.DeepEqual(want, normalizeSnapshot(claim.Snapshots[index])) {
					c.add("certificate.snapshot_mismatch", trace.ID, "snapshot does not match independently replayed state", event.ID)
				}
			}
		}
		if c.certificate.Mode == "replay" && len(claim.Snapshots) != len(trace.Events) {
			c.add("certificate.snapshot_count", trace.ID, fmt.Sprintf("received %d snapshots for %d events", len(claim.Snapshots), len(trace.Events)), "")
		}
		final := capture(current, "").Places
		if !reflect.DeepEqual(final, normalizePlaces(claim.FinalPlaces)) {
			c.add("certificate.final_state_mismatch", trace.ID, "final place state differs", "")
		}
		delete(claims, trace.ID)
	}
	for id := range claims {
		c.add("certificate.unknown_claim", id, "trace claim has no trace", "")
	}
}

func (c *checker) apply(current *machine, traceID string, event format.Event) {
	c.checks++
	ownership, exists := current.places[event.Place]
	if !exists {
		c.add("ownership.unknown_place", traceID, event.Place, event.ID)
		return
	}
	switch event.Op {
	case "borrow_shared", "borrow_exclusive":
		if ownership != "owned" {
			c.add("ownership.borrow_unavailable", event.Place, ownership, event.ID)
			return
		}
		if event.Loan == "" {
			c.add("ownership.loan_id", event.Place, "loan identity is empty", event.ID)
			return
		}
		if _, duplicate := current.loans[event.Loan]; duplicate {
			c.add("ownership.loan_id", event.Place, "loan identity already exists", event.ID)
			return
		}
		access := "shared"
		if event.Op == "borrow_exclusive" {
			access = "exclusive"
		}
		for _, loan := range current.loans {
			if loan.Place == event.Place && (access == "exclusive" || loan.Access == "exclusive") {
				c.add("ownership.borrow_conflict", event.Place, access+" conflicts with "+loan.Access, event.ID)
				return
			}
		}
		current.loans[event.Loan] = format.LoanState{ID: event.Loan, Place: event.Place, Access: access}
	case "end_loan":
		loan, ok := current.loans[event.Loan]
		if !ok || loan.Place != event.Place {
			c.add("ownership.unknown_loan", event.Place, event.Loan, event.ID)
			return
		}
		delete(current.loans, event.Loan)
	case "use":
		if ownership != "owned" {
			c.add("ownership.use_after_"+ownership, event.Place, "place is not owned", event.ID)
		}
	case "move", "destroy":
		if ownership != "owned" {
			c.add("ownership.double_consume", event.Place, ownership, event.ID)
			return
		}
		for _, loan := range current.loans {
			if loan.Place == event.Place {
				c.add("ownership.consume_during_loan", event.Place, loan.ID, event.ID)
				return
			}
		}
		if event.Op == "move" {
			current.places[event.Place] = "moved"
		} else {
			current.places[event.Place] = "destroyed"
		}
	default:
		c.add("ownership.unknown_operation", traceID, event.Op, event.ID)
	}
}

func capture(current machine, event string) format.Snapshot {
	snapshot := format.Snapshot{EventID: event}
	for place, ownership := range current.places {
		snapshot.Places = append(snapshot.Places, format.PlaceState{Place: place, Ownership: ownership})
	}
	for _, loan := range current.loans {
		snapshot.Loans = append(snapshot.Loans, loan)
	}
	return normalizeSnapshot(snapshot)
}

func normalizeSnapshot(snapshot format.Snapshot) format.Snapshot {
	snapshot.Places = normalizePlaces(snapshot.Places)
	sort.Slice(snapshot.Loans, func(i, j int) bool { return snapshot.Loans[i].ID < snapshot.Loans[j].ID })
	return snapshot
}

func normalizePlaces(places []format.PlaceState) []format.PlaceState {
	result := append([]format.PlaceState(nil), places...)
	sort.Slice(result, func(i, j int) bool { return result[i].Place < result[j].Place })
	return result
}

func knownAbility(ability string) bool { return member(format.AbilityOrder, ability) }

func member(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func (c *checker) add(code, subject, detail, event string) {
	c.diagnostics = append(c.diagnostics, format.Diagnostic{Code: code, Subject: subject, Event: event, Detail: detail})
}
