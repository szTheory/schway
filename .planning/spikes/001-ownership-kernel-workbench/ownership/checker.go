package ownership

import "sort"

type CheckerOptions struct {
	// Test-only fault injection proves the differential harness detects a
	// plausible checker bug instead of merely exercising happy paths.
	AllowMoveWhileShared bool
}

type checkerPlace struct {
	live      bool
	released  bool
	movedTo   string
	resource  bool
	order     int
	scope     int
	scopeID   string
	readLoans map[string]struct{}
	writeLoan string
}

type checkerLoan struct {
	owner string
	write bool
	live  bool
}

// Check statically interprets every instruction. It intentionally does not
// call the oracle transition implementation.
func Check(source Program, options CheckerOptions) Result {
	program := Normalize(source)
	result := Result{Program: source.ID, Valid: true, Events: []Event{}}
	places := map[string]*checkerPlace{}
	loans := map[string]*checkerLoan{}
	ordinal := 0
	scopes := []string{"root"}
	seenScopes := map[string]bool{"root": true}

	record := func(operation Operation, kind string) {
		result.Events = append(result.Events, Event{
			Sequence: len(result.Events), Kind: kind, Place: operation.Place,
			Target: operation.Target, Loan: operation.Loan, Scope: operation.Scope,
			SourceIndex: operation.SourceIndex, Synthetic: operation.Synthetic,
		})
	}
	reject := func(operation Operation, code, message string, causes ...Cause) Result {
		result.Valid = false
		result.Diagnostic = &Diagnostic{
			Code: code, Program: source.ID, SourceIndex: operation.SourceIndex,
			Operation: operation.Kind, Place: operation.Place, Loan: operation.Loan,
			Scope: operation.Scope,
			Message: message, Causes: causes, Repairs: RepairsFor(code),
		}
		return result
	}
	lookup := func(operation Operation) (*checkerPlace, *Result) {
		place, found := places[operation.Place]
		if !found {
			failure := reject(operation, "ownership.unknown_place", "place has not been declared")
			return nil, &failure
		}
		if !place.live {
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
			failure := reject(operation, code, message, causes...)
			return nil, &failure
		}
		return place, nil
	}
	closePlaces := func(minScope int, outcome string, sourceIndex int) *Result {
		type activePlace struct {
			name  string
			state *checkerPlace
		}
		active := []activePlace{}
		for name, state := range places {
			if state.live && state.scope >= minScope {
				active = append(active, activePlace{name, state})
			}
		}
		sort.Slice(active, func(i, j int) bool { return active[i].state.order > active[j].state.order })
		for _, item := range active {
			if len(item.state.readLoans) != 0 || item.state.writeLoan != "" {
				operation := Operation{Kind: Release, Place: item.name, SourceIndex: sourceIndex, Synthetic: true}
				failure := reject(operation, "ownership.release_while_borrowed", "lexical release conflicts with a live loan")
				return &failure
			}
			item.state.live = false
			item.state.released = true
			if item.state.resource {
				result.Events = append(result.Events, Event{
					Sequence: len(result.Events), Kind: "release_auto", Place: item.name,
					Scope: item.state.scopeID, Outcome: outcome,
					SourceIndex: sourceIndex, Synthetic: true,
				})
			}
		}
		return nil
	}
	finish := func(outcome string, sourceIndex int) *Result {
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
			if prior, found := places[operation.Place]; found && prior.live {
				return reject(operation, "ownership.duplicate_place", "cannot redeclare an initialized place")
			}
			ordinal++
			places[operation.Place] = &checkerPlace{
				live: true, resource: operation.Resource, order: ordinal,
				scope: len(scopes) - 1, scopeID: scopes[len(scopes)-1],
				readLoans: map[string]struct{}{},
			}
			record(operation, "declare")
		case BorrowShared, BorrowExclusive:
			place, failure := lookup(operation)
			if failure != nil {
				return *failure
			}
			if _, found := loans[operation.Loan]; found {
				return reject(operation, "ownership.duplicate_loan", "loan identity has already been used in this program")
			}
			if operation.Kind == BorrowShared {
				if place.writeLoan != "" {
					return reject(operation, "ownership.loan_conflict", "shared access conflicts with an exclusive loan", Cause{Relation: "exclusive_loan", Value: place.writeLoan})
				}
				place.readLoans[operation.Loan] = struct{}{}
				loans[operation.Loan] = &checkerLoan{owner: operation.Place, live: true}
			} else {
				if place.writeLoan != "" || len(place.readLoans) != 0 {
					return reject(operation, "ownership.loan_conflict", "exclusive access conflicts with a live loan")
				}
				place.writeLoan = operation.Loan
				loans[operation.Loan] = &checkerLoan{owner: operation.Place, write: true, live: true}
			}
			record(operation, string(operation.Kind))
		case ReadPlace:
			place, failure := lookup(operation)
			if failure != nil {
				return *failure
			}
			if place.writeLoan != "" {
				return reject(operation, "ownership.owner_access_conflict", "owner read conflicts with exclusive access", Cause{Relation: "exclusive_loan", Value: place.writeLoan})
			}
			record(operation, "read_place")
		case MutatePlace:
			place, failure := lookup(operation)
			if failure != nil {
				return *failure
			}
			if place.writeLoan != "" || len(place.readLoans) != 0 {
				return reject(operation, "ownership.owner_access_conflict", "owner mutation conflicts with a live loan")
			}
			record(operation, "mutate_place")
		case ReadLoan, MutateLoan:
			loan, found := loans[operation.Loan]
			if !found {
				return reject(operation, "ownership.unknown_loan", "loan has not been created")
			}
			if !loan.live {
				return reject(operation, "ownership.loan_ended", "loan has already ended")
			}
			if operation.Kind == MutateLoan && !loan.write {
				return reject(operation, "ownership.mutation_requires_exclusive", "mutation through a shared loan is forbidden")
			}
			operation.Place = loan.owner
			record(operation, string(operation.Kind))
		case EndLoan:
			loan, found := loans[operation.Loan]
			if !found {
				return reject(operation, "ownership.unknown_loan", "loan has not been created")
			}
			if !loan.live {
				return reject(operation, "ownership.loan_ended", "loan has already ended")
			}
			place := places[loan.owner]
			if loan.write {
				place.writeLoan = ""
			} else {
				delete(place.readLoans, operation.Loan)
			}
			loan.live = false
			operation.Place = loan.owner
			record(operation, "end_loan")
		case Move:
			place, failure := lookup(operation)
			if failure != nil {
				return *failure
			}
			blocked := place.writeLoan != "" || len(place.readLoans) != 0
			if options.AllowMoveWhileShared && place.writeLoan == "" {
				blocked = false
			}
			if blocked {
				return reject(operation, "ownership.move_while_borrowed", "cannot transfer ownership while a loan is live")
			}
			if target, found := places[operation.Target]; found && target.live {
				return reject(operation, "ownership.target_initialized", "move target is already initialized")
			}
			places[operation.Target] = &checkerPlace{
				live: true, resource: place.resource, order: place.order,
				scope: len(scopes) - 1, scopeID: scopes[len(scopes)-1],
				readLoans: map[string]struct{}{},
			}
			place.live = false
			place.movedTo = operation.Target
			record(operation, "move")
		case Release:
			place, failure := lookup(operation)
			if failure != nil {
				return *failure
			}
			if place.writeLoan != "" || len(place.readLoans) != 0 {
				return reject(operation, "ownership.release_while_borrowed", "cannot release a place while a loan is live")
			}
			place.live = false
			place.released = true
			record(operation, "release")
		case BeginScope:
			if operation.Scope == "" {
				return reject(operation, "ownership.scope_id_required", "lexical scope requires a stable identity")
			}
			if seenScopes[operation.Scope] {
				return reject(operation, "ownership.duplicate_scope", "scope identity has already been used in this program")
			}
			seenScopes[operation.Scope] = true
			scopes = append(scopes, operation.Scope)
			record(operation, "begin_scope")
		case EndScope:
			if len(scopes) == 1 {
				return reject(operation, "ownership.scope_underflow", "no nested lexical scope is active")
			}
			active := scopes[len(scopes)-1]
			if operation.Scope != active {
				return reject(operation, "ownership.scope_mismatch", "scope exit must match the innermost active scope", Cause{Relation: "active_scope", Value: active})
			}
			if failure := closePlaces(len(scopes)-1, "scope", operation.SourceIndex); failure != nil {
				return *failure
			}
			scopes = scopes[:len(scopes)-1]
			record(operation, "end_scope")
		case ExitError, ExitCancel, ExitPanic:
			outcome := map[Kind]string{ExitError: "error", ExitCancel: "cancel", ExitPanic: "panic"}[operation.Kind]
			if failure := finish(outcome, operation.SourceIndex); failure != nil {
				return *failure
			}
			return result
		default:
			return reject(operation, "ownership.unknown_operation", "operation is not part of the ownership kernel")
		}
	}
	last := len(source.Operations)
	if last > 0 {
		last--
	}
	if len(scopes) != 1 {
		return reject(Operation{Kind: EndScope, Scope: scopes[len(scopes)-1], SourceIndex: last}, "ownership.unclosed_scope", "nested lexical scope is not closed on normal completion")
	}
	if failure := finish("success", last); failure != nil {
		return *failure
	}
	return result
}
