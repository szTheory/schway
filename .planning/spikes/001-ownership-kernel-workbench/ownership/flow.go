package ownership

import "sort"

type FlowStep struct {
	Operation *Operation  `json:"operation,omitempty"`
	Branch    *FlowBranch `json:"branch,omitempty"`
	Loop      *FlowLoop   `json:"loop,omitempty"`
}

type FlowBranch struct {
	ID   string     `json:"id"`
	Then []FlowStep `json:"then"`
	Else []FlowStep `json:"else"`
}

type FlowLoop struct {
	ID   string     `json:"id"`
	Body []FlowStep `json:"body"`
}

type FlowProgram struct {
	ID    string     `json:"id"`
	Steps []FlowStep `json:"steps"`
}

type FlowOptions struct {
	// Test-only fault injection. Treating a maybe-live place as definitely live
	// is an archetypal unsound join implementation.
	AllowMaybeUse bool
}

type FlowResult struct {
	Program    string            `json:"program"`
	Valid      bool              `json:"valid"`
	Diagnostic *Diagnostic       `json:"diagnostic,omitempty"`
	States     map[string]string `json:"states"`
	Loans      map[string]string `json:"loans,omitempty"`
	Operations int               `json:"operations"`
}

type FlowComparison struct {
	Program   string       `json:"program"`
	Agreement bool         `json:"agreement"`
	Analyzer  FlowResult   `json:"analyzer"`
	Paths     []Comparison `json:"paths"`
}

type FlowFixture struct {
	Name        string      `json:"name"`
	Program     FlowProgram `json:"program"`
	ExpectValid bool        `json:"expect_valid"`
	ExpectCode  string      `json:"expect_code,omitempty"`
}

type FlowFixtureFile struct {
	Schema   int           `json:"schema"`
	Fixtures []FlowFixture `json:"fixtures"`
}

const (
	flowLive uint8 = 1 << iota
	flowMoved
	flowReleased
)

type flowPlace struct {
	states   uint8
	resource bool
}

const (
	flowLoanActive uint8 = 1 << iota
	flowLoanEnded
	flowLoanAbsent
)

type flowLoan struct {
	states    uint8
	owner     string
	write     bool
	ambiguous bool
}

type flowAnalysis struct {
	program    string
	places     map[string]flowPlace
	loans      map[string]flowLoan
	diagnostic *Diagnostic
	operations int
	options    FlowOptions
}

func AnalyzeFlow(program FlowProgram, options FlowOptions) FlowResult {
	analysis := &flowAnalysis{
		program: program.ID, places: map[string]flowPlace{},
		loans: map[string]flowLoan{}, options: options,
	}
	analysis.sequence(program.Steps)
	states := map[string]string{}
	names := make([]string, 0, len(analysis.places))
	for name := range analysis.places {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		states[name] = renderFlowState(analysis.places[name].states)
	}
	loans := map[string]string{}
	loanNames := make([]string, 0, len(analysis.loans))
	for name := range analysis.loans {
		loanNames = append(loanNames, name)
	}
	sort.Strings(loanNames)
	for _, name := range loanNames {
		loans[name] = renderFlowLoanState(analysis.loans[name])
	}
	return FlowResult{
		Program: program.ID, Valid: analysis.diagnostic == nil,
		Diagnostic: analysis.diagnostic, States: states, Loans: loans,
		Operations: analysis.operations,
	}
}

func (analysis *flowAnalysis) sequence(steps []FlowStep) {
	for _, step := range steps {
		if analysis.diagnostic != nil {
			return
		}
		switch {
		case countFlowForms(step) != 1:
			operation := Operation{}
			if step.Operation != nil {
				operation = *step.Operation
			}
			analysis.reject(operation, "ownership.invalid_flow_step", "flow step must contain exactly one operation, branch, or loop")
		case step.Operation != nil:
			analysis.apply(*step.Operation)
		case step.Branch != nil:
			analysis.branch(*step.Branch)
		case step.Loop != nil:
			analysis.loop(*step.Loop)
		}
	}
}

func countFlowForms(step FlowStep) int {
	count := 0
	if step.Operation != nil {
		count++
	}
	if step.Branch != nil {
		count++
	}
	if step.Loop != nil {
		count++
	}
	return count
}

func (analysis *flowAnalysis) branch(branch FlowBranch) {
	thenAnalysis := analysis.clone()
	thenAnalysis.sequence(branch.Then)
	if thenAnalysis.diagnostic != nil {
		analysis.diagnostic = thenAnalysis.diagnostic
		analysis.operations += thenAnalysis.operations - analysis.operations
		return
	}
	elseAnalysis := analysis.clone()
	elseAnalysis.sequence(branch.Else)
	if elseAnalysis.diagnostic != nil {
		analysis.diagnostic = elseAnalysis.diagnostic
		analysis.operations += elseAnalysis.operations - analysis.operations
		return
	}
	analysis.operations = max(thenAnalysis.operations, elseAnalysis.operations)
	analysis.places = joinFlowPlaces(thenAnalysis.places, elseAnalysis.places)
	analysis.loans = joinFlowLoans(thenAnalysis.loans, elseAnalysis.loans)
	if loan, found := firstAmbiguousFlowLoan(analysis.loans); found {
		analysis.reject(Operation{Kind: ReadLoan, Loan: loan}, "ownership.loan_join_mismatch", "incoming paths disagree about the loan owner or mode")
	}
}

func (analysis *flowAnalysis) loop(loop FlowLoop) {
	entry := cloneFlowPlaces(analysis.places)
	entryLoans := cloneFlowLoans(analysis.loans)
	current := cloneFlowPlaces(entry)
	currentLoans := cloneFlowLoans(entryLoans)
	const maxIterations = 8
	for iteration := 0; iteration < maxIterations; iteration++ {
		bodyAnalysis := &flowAnalysis{
			program: analysis.program, places: cloneFlowPlaces(current),
			loans: cloneFlowLoans(currentLoans), operations: analysis.operations,
			options: analysis.options,
		}
		bodyAnalysis.sequence(loop.Body)
		if bodyAnalysis.diagnostic != nil {
			analysis.diagnostic = bodyAnalysis.diagnostic
			analysis.operations = bodyAnalysis.operations
			return
		}
		joined := joinFlowPlaces(entry, bodyAnalysis.places)
		joinedLoans := joinFlowLoans(entryLoans, bodyAnalysis.loans)
		analysis.operations = bodyAnalysis.operations
		if loan, found := firstAmbiguousFlowLoan(joinedLoans); found {
			analysis.loans = joinedLoans
			analysis.reject(Operation{Kind: ReadLoan, Loan: loan}, "ownership.loan_join_mismatch", "loop paths disagree about the loan owner or mode")
			return
		}
		if equalFlowPlaces(current, joined) && equalFlowLoans(currentLoans, joinedLoans) {
			analysis.places = joined
			analysis.loans = joinedLoans
			return
		}
		current = joined
		currentLoans = joinedLoans
	}
	analysis.reject(Operation{}, "ownership.flow_nonconvergent", "loop ownership analysis did not reach a fixed point")
}

func joinFlowPlaces(left, right map[string]flowPlace) map[string]flowPlace {
	joined := cloneFlowPlaces(left)
	for name, other := range right {
		place := joined[name]
		place.states |= other.states
		place.resource = place.resource || other.resource
		joined[name] = place
	}
	for name, place := range joined {
		_, inLeft := left[name]
		_, inRight := right[name]
		if inLeft != inRight {
			place.states |= flowReleased
			joined[name] = place
		}
	}
	return joined
}

func cloneFlowPlaces(source map[string]flowPlace) map[string]flowPlace {
	places := make(map[string]flowPlace, len(source))
	for name, place := range source {
		places[name] = place
	}
	return places
}

func joinFlowLoans(left, right map[string]flowLoan) map[string]flowLoan {
	joined := cloneFlowLoans(left)
	for name, other := range right {
		loan, found := joined[name]
		if !found {
			other.states |= flowLoanAbsent
			joined[name] = other
			continue
		}
		loan.states |= other.states
		loan.ambiguous = loan.ambiguous || other.ambiguous || loan.owner != other.owner || loan.write != other.write
		joined[name] = loan
	}
	for name, loan := range joined {
		if _, inRight := right[name]; !inRight {
			loan.states |= flowLoanAbsent
			joined[name] = loan
		}
	}
	return joined
}

func cloneFlowLoans(source map[string]flowLoan) map[string]flowLoan {
	loans := make(map[string]flowLoan, len(source))
	for name, loan := range source {
		loans[name] = loan
	}
	return loans
}

func firstAmbiguousFlowLoan(loans map[string]flowLoan) (string, bool) {
	names := make([]string, 0, len(loans))
	for name := range loans {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if loans[name].ambiguous {
			return name, true
		}
	}
	return "", false
}

func equalFlowPlaces(left, right map[string]flowPlace) bool {
	if len(left) != len(right) {
		return false
	}
	for name, place := range left {
		if right[name] != place {
			return false
		}
	}
	return true
}

func equalFlowLoans(left, right map[string]flowLoan) bool {
	if len(left) != len(right) {
		return false
	}
	for name, loan := range left {
		if right[name] != loan {
			return false
		}
	}
	return true
}

func (analysis *flowAnalysis) clone() *flowAnalysis {
	return &flowAnalysis{
		program: analysis.program, places: cloneFlowPlaces(analysis.places), operations: analysis.operations,
		loans: cloneFlowLoans(analysis.loans), options: analysis.options,
	}
}

func (analysis *flowAnalysis) apply(operation Operation) {
	analysis.operations++
	switch operation.Kind {
	case Declare:
		if place, found := analysis.places[operation.Place]; found && place.states&flowLive != 0 {
			analysis.reject(operation, "ownership.duplicate_place", "cannot redeclare a place that may already be initialized")
			return
		}
		analysis.places[operation.Place] = flowPlace{states: flowLive, resource: operation.Resource}
	case ReadPlace:
		if !analysis.requireLive(operation) {
			return
		}
		if loan, found := analysis.conflictingLoan(operation.Place, true); found {
			analysis.reject(operation, "ownership.owner_access_conflict", "owner read conflicts with exclusive access", Cause{Relation: "loan", Value: loan})
			return
		}
	case MutatePlace:
		if !analysis.requireLive(operation) {
			return
		}
		if loan, found := analysis.conflictingLoan(operation.Place, false); found {
			analysis.reject(operation, "ownership.owner_access_conflict", "owner mutation conflicts with a live loan", Cause{Relation: "loan", Value: loan})
			return
		}
	case BorrowShared, BorrowExclusive:
		if !analysis.requireLive(operation) {
			return
		}
		if _, found := analysis.loans[operation.Loan]; found {
			analysis.reject(operation, "ownership.duplicate_loan", "loan identity has already been used on an incoming path")
			return
		}
		write := operation.Kind == BorrowExclusive
		if loan, found := analysis.conflictingLoan(operation.Place, !write); found {
			analysis.reject(operation, "ownership.loan_conflict", "requested access conflicts with a live loan", Cause{Relation: "loan", Value: loan})
			return
		}
		analysis.loans[operation.Loan] = flowLoan{states: flowLoanActive, owner: operation.Place, write: write}
	case ReadLoan, MutateLoan:
		loan, ok := analysis.requireActiveLoan(operation)
		if !ok {
			return
		}
		if operation.Kind == MutateLoan && !loan.write {
			analysis.reject(operation, "ownership.mutation_requires_exclusive", "mutation through a shared loan is forbidden")
			return
		}
	case EndLoan:
		loan, ok := analysis.requireActiveLoan(operation)
		if !ok {
			return
		}
		loan.states = flowLoanEnded
		analysis.loans[operation.Loan] = loan
	case Move:
		if !analysis.requireLive(operation) {
			return
		}
		if loan, found := analysis.conflictingLoan(operation.Place, false); found {
			analysis.reject(operation, "ownership.move_while_borrowed", "cannot transfer ownership while a loan may be live", Cause{Relation: "loan", Value: loan})
			return
		}
		if target, found := analysis.places[operation.Target]; found && target.states&flowLive != 0 {
			analysis.reject(operation, "ownership.target_maybe_initialized", "move target may already be initialized")
			return
		}
		source := analysis.places[operation.Place]
		analysis.places[operation.Place] = flowPlace{states: flowMoved, resource: source.resource}
		analysis.places[operation.Target] = flowPlace{states: flowLive, resource: source.resource}
	case Release:
		if !analysis.requireLive(operation) {
			return
		}
		if loan, found := analysis.conflictingLoan(operation.Place, false); found {
			analysis.reject(operation, "ownership.release_while_borrowed", "cannot release a place while a loan may be live", Cause{Relation: "loan", Value: loan})
			return
		}
		place := analysis.places[operation.Place]
		place.states = flowReleased
		analysis.places[operation.Place] = place
	default:
		analysis.reject(operation, "ownership.unsupported_flow_operation", "operation is not yet supported by the branch-join analyzer")
	}
}

func (analysis *flowAnalysis) conflictingLoan(owner string, exclusiveOnly bool) (string, bool) {
	names := make([]string, 0, len(analysis.loans))
	for name := range analysis.loans {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		loan := analysis.loans[name]
		if loan.owner == owner && loan.states&flowLoanActive != 0 && (!exclusiveOnly || loan.write) {
			return name, true
		}
	}
	return "", false
}

func (analysis *flowAnalysis) requireActiveLoan(operation Operation) (flowLoan, bool) {
	loan, found := analysis.loans[operation.Loan]
	if !found || loan.states == flowLoanAbsent {
		analysis.reject(operation, "ownership.unknown_loan", "loan has not been created on any incoming path")
		return flowLoan{}, false
	}
	if loan.ambiguous {
		analysis.reject(operation, "ownership.loan_join_mismatch", "incoming paths disagree about the loan owner or mode")
		return flowLoan{}, false
	}
	if loan.states == flowLoanActive {
		return loan, true
	}
	if loan.states == flowLoanEnded {
		analysis.reject(operation, "ownership.loan_ended", "loan has already ended")
		return flowLoan{}, false
	}
	analysis.reject(operation, "ownership.loan_maybe_inactive", "loan is not active on every incoming path", Cause{Relation: "incoming_states", Value: renderFlowLoanState(loan)})
	return flowLoan{}, false
}

func (analysis *flowAnalysis) requireLive(operation Operation) bool {
	place, found := analysis.places[operation.Place]
	if !found {
		analysis.reject(operation, "ownership.unknown_place", "place has not been declared on any incoming path")
		return false
	}
	if place.states == flowLive || (analysis.options.AllowMaybeUse && place.states&flowLive != 0) {
		return true
	}
	code := "ownership.maybe_uninitialized"
	message := "place is not initialized on every incoming path"
	if place.states == flowMoved {
		code = "ownership.use_after_move"
		message = "ownership was transferred from this place"
	} else if place.states == flowReleased {
		code = "ownership.use_after_release"
		message = "place was already released"
	}
	analysis.reject(operation, code, message, Cause{Relation: "incoming_states", Value: renderFlowState(place.states)})
	return false
}

func (analysis *flowAnalysis) reject(operation Operation, code, message string, causes ...Cause) {
	analysis.diagnostic = &Diagnostic{
		Code: code, Program: analysis.program, SourceIndex: operation.SourceIndex,
		Operation: operation.Kind, Place: operation.Place, Loan: operation.Loan,
		Scope: operation.Scope,
		Message: message, Causes: causes, Repairs: RepairsFor(code),
	}
}

func renderFlowState(states uint8) string {
	parts := []string{}
	if states&flowLive != 0 {
		parts = append(parts, "live")
	}
	if states&flowMoved != 0 {
		parts = append(parts, "moved")
	}
	if states&flowReleased != 0 {
		parts = append(parts, "released_or_absent")
	}
	if len(parts) == 0 {
		return "absent"
	}
	result := parts[0]
	for _, part := range parts[1:] {
		result += "|" + part
	}
	return result
}

func renderFlowLoanState(loan flowLoan) string {
	parts := []string{}
	if loan.states&flowLoanActive != 0 {
		parts = append(parts, "active")
	}
	if loan.states&flowLoanEnded != 0 {
		parts = append(parts, "ended")
	}
	if loan.states&flowLoanAbsent != 0 {
		parts = append(parts, "absent")
	}
	if loan.ambiguous {
		parts = append(parts, "ambiguous")
	}
	if len(parts) == 0 {
		return "absent"
	}
	result := parts[0]
	for _, part := range parts[1:] {
		result += "|" + part
	}
	return result
}

func ExpandFlow(program FlowProgram) []Program {
	paths := expandSteps(program.Steps, 3)
	programs := make([]Program, 0, len(paths))
	for index, operations := range paths {
		programs = append(programs, Program{ID: program.ID + "/path-" + itoa(index), Operations: operations})
	}
	return programs
}

func expandSteps(steps []FlowStep, loopLimit int) [][]Operation {
	paths := [][]Operation{{}}
	for _, step := range steps {
		var fragments [][]Operation
		if countFlowForms(step) != 1 {
			fragments = [][]Operation{{{Kind: "invalid_flow_step"}}}
		} else if step.Operation != nil {
			fragments = [][]Operation{{*step.Operation}}
		} else if step.Branch != nil {
			fragments = append(expandSteps(step.Branch.Then, loopLimit), expandSteps(step.Branch.Else, loopLimit)...)
		} else if step.Loop != nil {
			bodyPaths := expandSteps(step.Loop.Body, loopLimit)
			fragments = [][]Operation{{}}
			for count := 1; count <= loopLimit; count++ {
				fragments = append(fragments, repeatFlowPaths(bodyPaths, count)...)
			}
		}
		combined := make([][]Operation, 0, len(paths)*len(fragments))
		for _, prefix := range paths {
			for _, fragment := range fragments {
				path := append([]Operation{}, prefix...)
				path = append(path, fragment...)
				combined = append(combined, path)
			}
		}
		paths = combined
	}
	return paths
}

func repeatFlowPaths(bodyPaths [][]Operation, count int) [][]Operation {
	result := [][]Operation{{}}
	for iteration := 0; iteration < count; iteration++ {
		combined := make([][]Operation, 0, len(result)*len(bodyPaths))
		for _, prefix := range result {
			for _, body := range bodyPaths {
				path := append([]Operation{}, prefix...)
				path = append(path, body...)
				combined = append(combined, path)
			}
		}
		result = combined
	}
	return result
}

func CompareFlow(program FlowProgram, options FlowOptions) FlowComparison {
	analyzer := AnalyzeFlow(program, options)
	paths := ExpandFlow(program)
	comparisons := make([]Comparison, 0, len(paths))
	allPathsValid := true
	pathsAgree := true
	for _, path := range paths {
		comparison := Compare(path, CheckerOptions{})
		comparisons = append(comparisons, comparison)
		if !comparison.Agreement {
			pathsAgree = false
		}
		if !comparison.Checker.Valid {
			allPathsValid = false
		}
	}
	return FlowComparison{
		Program: program.ID, Agreement: pathsAgree && analyzer.Valid == allPathsValid,
		Analyzer: analyzer, Paths: comparisons,
	}
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
