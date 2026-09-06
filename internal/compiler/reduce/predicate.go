package reduce

// Signature is the fingerprint of one interpreter-mismatch observation that
// Interesting compares a seed against a reduction candidate on. It is
// intentionally decoupled from internal/compiler/session's own
// Phase5EngineDisagreement type: reduce must stay a leaf package (no
// dependency on session), since plan 05-12 wires session -> reduce, and a
// reduce -> session dependency in the other direction would create an
// import cycle. Plan 05-12's own wiring file is responsible for mapping a
// session.Phase5EngineDisagreement into a Signature (D-05-24) -- checked
// against the actual import graph before that plan writes it, per this
// plan's own instruction to keep reduce dependency-free.
type Signature struct {
	// Axis is the diverging comparator axis (e.g. one of
	// session.AxisTerminalOutcome / AxisEventOrder / AxisResourceLedger /
	// AxisExitStatusSignal / AxisDiagnosticID) carried as a plain string so
	// this package never imports session's constants.
	Axis string
	// OperationID is the first-diverging operation's identity.
	OperationID string
	// CausalRole is the fallback identity D-05-24 explicitly permits when an
	// operation's POSITION shifts under reduction but its causal role in the
	// program does not (e.g. "the acquisition whose release is observed
	// last"). Only one of OperationID/CausalRole needs to match for the
	// predicate to accept a shifted-position candidate.
	CausalRole string
	// EnginePair names which two engines disagreed (e.g.
	// "interpreter-vs-O0").
	EnginePair string
	// ForeignCallSequence is the ordered list of foreign symbol
	// invocations on the causal path -- rejecting drift here is the local
	// analogue of C-Reduce's UB-drift failure: the foreign boundary is the
	// one place this language's UB genuinely lives, so a reduction that
	// perturbs the call order, the symbol, or an argument shape is
	// producing a DIFFERENT program, not a smaller one (D-05-24).
	ForeignCallSequence []string
}

// Predicate is declared in reduce.go alongside Reduce, which it drives.

// Interesting implements D-05-24's conjunction exactly: a reduction
// candidate is interesting only when it preserves
//
//   - the same diverging Axis, AND
//   - the same first-diverging OperationID -- OR, if the operation's
//     position shifted under reduction, the same CausalRole (the ONE
//     permitted relaxation D-05-24 names), AND
//   - the same disagreeing EnginePair, AND
//   - no drift in the foreign-boundary call sequence on the causal path
//     (order, symbol, or argument shape).
//
// A candidate that still fails, but on a different axis, at a different
// operation, or between a different engine pair, is REJECTED as
// uninteresting -- this is what stops the reducer from wandering off to
// report a different bug than the one it started with.
func Interesting(seed Signature, candidate Signature) bool {
	if seed.Axis == "" || seed.Axis != candidate.Axis {
		return false
	}
	if seed.EnginePair == "" || seed.EnginePair != candidate.EnginePair {
		return false
	}
	operationMatches := seed.OperationID != "" && seed.OperationID == candidate.OperationID
	causalRoleMatches := seed.CausalRole != "" && seed.CausalRole == candidate.CausalRole
	if !operationMatches && !causalRoleMatches {
		return false
	}
	if foreignCallSequenceDrifted(seed.ForeignCallSequence, candidate.ForeignCallSequence) {
		return false
	}
	return true
}

// foreignCallSequenceDrifted reports whether candidate's foreign call
// sequence differs from seed's in length, order, or content -- any of
// which independently means a reordered call, a changed symbol, or (since
// each entry is expected to encode symbol+argument-shape together) a
// changed argument shape.
func foreignCallSequenceDrifted(seed, candidate []string) bool {
	if len(seed) != len(candidate) {
		return true
	}
	for i := range seed {
		if seed[i] != candidate[i] {
			return true
		}
	}
	return false
}
