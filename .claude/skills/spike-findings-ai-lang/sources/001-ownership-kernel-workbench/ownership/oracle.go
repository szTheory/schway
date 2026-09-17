package ownership

import "sort"

type oraclePlace struct {
	initialized bool
	released    bool
	movedTo     string
	resource    bool
	order       int
	scope       int
	scopeID     string
	shared      map[string]bool
	exclusive   string
}

type oracleLoan struct {
	place  string
	mode   string
	active bool
}

// RunOracle executes the normalized kernel as a dynamic ownership monitor.
// Its implementation is intentionally separate from the static checker.
func RunOracle(source Program) Result {
	program := NormalizeOracle(source)
	result := Result{Program: source.ID, Valid: true, Events: []Event{}}
	places := map[string]*oraclePlace{}
	loans := map[string]*oracleLoan{}
	order := 0
	scopes := []string{"root"}
	seenScopes := map[string]bool{"root": true}

	emit := func(operation Operation, kind string) {
		result.Events = append(result.Events, Event{
			Sequence: len(result.Events), Kind: kind, Place: operation.Place,
			Target: operation.Target, Loan: operation.Loan, Scope: operation.Scope,
			SourceIndex: operation.SourceIndex, Synthetic: operation.Synthetic,
		})
	}
	fail := func(operation Operation, code, message string, causes ...Cause) Result {
		result.Valid = false
		result.Diagnostic = &Diagnostic{
			Code: code, Program: source.ID, SourceIndex: operation.SourceIndex,
			Operation: operation.Kind, Place: operation.Place, Loan: operation.Loan,
			Scope: operation.Scope,
			Message: message, Causes: causes, Repairs: RepairsFor(code),
		}
		return result
	}
	ready := func(operation Operation) (*oraclePlace, *Result) {
		place, exists := places[operation.Place]
		if !exists {
			failure := fail(operation, "ownership.unknown_place", "place has not been declared")
			return nil, &failure
		}
		if !place.initialized {
			code := "ownership.uninitialized"
			message := "place is not initialized"
			causes := []Cause{}
			if place.movedTo != "" {
				code = "ownership.use_after_move"
				message = "ownership was transferred from this place"
				causes = append(causes, Cause{Relation: "moved_to", Value: place.movedTo})
			} else if place.released {
				code = "ownership.use_after_release"
				message = "place was already released"
			}
			failure := fail(operation, code, message, causes...)
			return nil, &failure
		}
		return place, nil
	}
	closePlaces := func(minScope int, outcome string, sourceIndex int) *Result {
		active := make([]struct {
			name  string
			place *oraclePlace
		}, 0)
		for name, place := range places {
			if place.initialized && place.scope >= minScope {
				active = append(active, struct {
					name  string
					place *oraclePlace
				}{name, place})
			}
		}
		sort.Slice(active, func(i, j int) bool { return active[i].place.order > active[j].place.order })
		for _, item := range active {
			if len(item.place.shared) > 0 || item.place.exclusive != "" {
				operation := Operation{Kind: Release, Place: item.name, SourceIndex: sourceIndex, Synthetic: true}
				failure := fail(operation, "ownership.release_while_borrowed", "lexical release conflicts with a live loan")
				return &failure
			}
			item.place.initialized = false
			item.place.released = true
			if item.place.resource {
				result.Events = append(result.Events, Event{
					Sequence: len(result.Events), Kind: "release_auto", Place: item.name,
					Scope: item.place.scopeID, Outcome: outcome,
					SourceIndex: sourceIndex, Synthetic: true,
				})
			}
		}
		return nil
	}
	cleanup := func(outcome string, sourceIndex int) *Result {
		if failure := closePlaces(0, outcome, sourceIndex); failure != nil {
			return failure
		}
		result.Events = append(result.Events, Event{
			Sequence: len(result.Events), Kind: "exit", Outcome: outcome,
			SourceIndex: sourceIndex, Synthetic: true,
		})
		return nil
	}

	for _, operation := range program.Operations {
		result.Operations++
		switch operation.Kind {
		case Declare:
			if place, exists := places[operation.Place]; exists && place.initialized {
				return fail(operation, "ownership.duplicate_place", "cannot redeclare an initialized place")
			}
			order++
			places[operation.Place] = &oraclePlace{
				initialized: true, resource: operation.Resource, order: order,
				scope: len(scopes) - 1, scopeID: scopes[len(scopes)-1],
				shared: map[string]bool{},
			}
			emit(operation, "declare")
		case BorrowShared, BorrowExclusive:
			place, failure := ready(operation)
			if failure != nil {
				return *failure
			}
			if _, exists := loans[operation.Loan]; exists {
				return fail(operation, "ownership.duplicate_loan", "loan identity has already been used in this program")
			}
			if operation.Kind == BorrowShared {
				if place.exclusive != "" {
					return fail(operation, "ownership.loan_conflict", "shared access conflicts with an exclusive loan", Cause{Relation: "exclusive_loan", Value: place.exclusive})
				}
				place.shared[operation.Loan] = true
				loans[operation.Loan] = &oracleLoan{place: operation.Place, mode: "shared", active: true}
			} else {
				if place.exclusive != "" || len(place.shared) > 0 {
					return fail(operation, "ownership.loan_conflict", "exclusive access conflicts with a live loan")
				}
				place.exclusive = operation.Loan
				loans[operation.Loan] = &oracleLoan{place: operation.Place, mode: "exclusive", active: true}
			}
			emit(operation, string(operation.Kind))
		case ReadPlace:
			place, failure := ready(operation)
			if failure != nil {
				return *failure
			}
			if place.exclusive != "" {
				return fail(operation, "ownership.owner_access_conflict", "owner read conflicts with exclusive access", Cause{Relation: "exclusive_loan", Value: place.exclusive})
			}
			emit(operation, "read_place")
		case MutatePlace:
			place, failure := ready(operation)
			if failure != nil {
				return *failure
			}
			if place.exclusive != "" || len(place.shared) > 0 {
				return fail(operation, "ownership.owner_access_conflict", "owner mutation conflicts with a live loan")
			}
			emit(operation, "mutate_place")
		case ReadLoan, MutateLoan:
			loan, exists := loans[operation.Loan]
			if !exists {
				return fail(operation, "ownership.unknown_loan", "loan has not been created")
			}
			if !loan.active {
				return fail(operation, "ownership.loan_ended", "loan has already ended")
			}
			if operation.Kind == MutateLoan && loan.mode != "exclusive" {
				return fail(operation, "ownership.mutation_requires_exclusive", "mutation through a shared loan is forbidden")
			}
			operation.Place = loan.place
			emit(operation, string(operation.Kind))
		case EndLoan:
			loan, exists := loans[operation.Loan]
			if !exists {
				return fail(operation, "ownership.unknown_loan", "loan has not been created")
			}
			if !loan.active {
				return fail(operation, "ownership.loan_ended", "loan has already ended")
			}
			place := places[loan.place]
			if loan.mode == "shared" {
				delete(place.shared, operation.Loan)
			} else {
				place.exclusive = ""
			}
			loan.active = false
			operation.Place = loan.place
			emit(operation, "end_loan")
		case Move:
			place, failure := ready(operation)
			if failure != nil {
				return *failure
			}
			if place.exclusive != "" || len(place.shared) > 0 {
				return fail(operation, "ownership.move_while_borrowed", "cannot transfer ownership while a loan is live")
			}
			if target, exists := places[operation.Target]; exists && target.initialized {
				return fail(operation, "ownership.target_initialized", "move target is already initialized")
			}
			places[operation.Target] = &oraclePlace{
				initialized: true, resource: place.resource, order: place.order,
				scope: len(scopes) - 1, scopeID: scopes[len(scopes)-1],
				shared: map[string]bool{},
			}
			place.initialized = false
			place.movedTo = operation.Target
			emit(operation, "move")
		case Release:
			place, failure := ready(operation)
			if failure != nil {
				return *failure
			}
			if place.exclusive != "" || len(place.shared) > 0 {
				return fail(operation, "ownership.release_while_borrowed", "cannot release a place while a loan is live")
			}
			place.initialized = false
			place.released = true
			emit(operation, "release")
		case BeginScope:
			if operation.Scope == "" {
				return fail(operation, "ownership.scope_id_required", "lexical scope requires a stable identity")
			}
			if seenScopes[operation.Scope] {
				return fail(operation, "ownership.duplicate_scope", "scope identity has already been used in this program")
			}
			seenScopes[operation.Scope] = true
			scopes = append(scopes, operation.Scope)
			emit(operation, "begin_scope")
		case EndScope:
			if len(scopes) == 1 {
				return fail(operation, "ownership.scope_underflow", "no nested lexical scope is active")
			}
			active := scopes[len(scopes)-1]
			if operation.Scope != active {
				return fail(operation, "ownership.scope_mismatch", "scope exit must match the innermost active scope", Cause{Relation: "active_scope", Value: active})
			}
			if failure := closePlaces(len(scopes)-1, "scope", operation.SourceIndex); failure != nil {
				return *failure
			}
			scopes = scopes[:len(scopes)-1]
			emit(operation, "end_scope")
		case ExitError, ExitCancel, ExitPanic:
			outcome := map[Kind]string{ExitError: "error", ExitCancel: "cancel", ExitPanic: "panic"}[operation.Kind]
			if failure := cleanup(outcome, operation.SourceIndex); failure != nil {
				return *failure
			}
			return result
		default:
			return fail(operation, "ownership.unknown_operation", "operation is not part of the ownership kernel")
		}
	}
	last := len(source.Operations)
	if last > 0 {
		last--
	}
	if len(scopes) != 1 {
		return fail(Operation{Kind: EndScope, Scope: scopes[len(scopes)-1], SourceIndex: last}, "ownership.unclosed_scope", "nested lexical scope is not closed on normal completion")
	}
	if failure := cleanup("success", last); failure != nil {
		return *failure
	}
	return result
}
