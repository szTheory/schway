package check

import (
	"fmt"
	"sort"
	"strings"

	"github.com/codename-lang/lang/internal/compiler/ability"
	"github.com/codename-lang/lang/internal/compiler/ast"
	"github.com/codename-lang/lang/internal/compiler/callgraph"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/originvalidate"
)

// testOnlyForceUniformLoanJoin is a fault-injection seam for
// TestUniformJoinPlacementFlipsBothVerdicts (03-03-02): when true, every
// arm-body loan's computed last use is forced to len(body.Bindings) -- as
// if edge-specific placement had been deleted and every loan ended
// uniformly at the join regardless of which edge actually needs it --
// instead of whatever computeLoanLastUses would otherwise compute. It is
// read only by analyzeArmBody, never by analyzeStraightLine, so the
// straight-line exhaustive differential (TestOwnershipSequenceExhaustive)
// is untouched by its existence. This is a test-only seam, not a
// production code path: a test that ships its own seam cannot be
// falsified by reverting a production hunk (the seam would simply never
// engage), so this specific claim is falsified by direct mutation instead
// -- flipping this variable and observing a verdict change IS the
// falsifying action, recorded verbatim in the owning plan's summary
// (D-09, mirroring 02-VALIDATION.md's recorded residual-weakness
// precedent for fault-injection seams).
var testOnlyForceUniformLoanJoin = false

type Result struct {
	Program     core.Program
	Diagnostics []diagnostic.Diagnostic
	Work        int
	// AliasFacts is D-05-01's admission-gating alias-lattice fact set
	// (deriveAliasFacts): exposed here, on the checker's existing result,
	// rather than as a new core.Program field, so cgen's restrict emission
	// and the evidence path can reach it without a lang.core/2 schema bump
	// (D-05-39). Empty for every Phase 1-4 program.
	AliasFacts []AliasFact
}

func Program(program ast.Program) Result {
	result := Result{Program: core.Program{Schema: core.Schema, Module: program.Module, ModuleID: semanticID(program.Module, "module", program.Module)}}
	// spanByOperationID is D-07-35's program-wide merge of every checked
	// function's own CallSpans: check's OWN emission-time bookkeeping,
	// consulted only when a core.call_graph_cycle diagnostic must project
	// an OpCall operation ID onto a source span. No Span is added to
	// core.LinearOperation.
	spanByOperationID := map[string]diagnostic.Span{}
	hasMatch, hasLinear := false, false
	for _, function := range program.Funcs {
		if function.Body.Linear != nil {
			hasLinear = true
		} else {
			hasMatch = true
		}
	}
	if hasMatch && hasLinear {
		span := diagnostic.Span{}
		if len(program.Funcs) > 0 {
			span = program.Funcs[0].Span
		}
		result.Diagnostics = append(result.Diagnostics, diagnostic.Error(
			"core.mixed_body_versions",
			span,
			"a module cannot mix match and linear function bodies until function-level core versioning is defined",
		))
		return result
	}
	types := make(map[string]core.DataType)
	for _, declaration := range program.Data {
		alternatives := make([]string, 0, len(declaration.Alternatives))
		for _, alternative := range declaration.Alternatives {
			alternatives = append(alternatives, alternative.Name)
		}
		dataType := core.DataType{ID: semanticID(program.Module, "type", declaration.Name), Name: declaration.Name, Alternatives: alternatives, Span: declaration.Span}
		types[declaration.Name] = dataType
		result.Program.DataTypes = append(result.Program.DataTypes, dataType)
	}

	// Phase 4: the foreign symbol table and the Lang function-name set both
	// exist purely so a fallible call's callee can be resolved against one or
	// the other (D-04-01/D-04-02) -- a callee resolving into functionNames
	// rather than foreignSymbols is core.call_target_not_foreign, never an
	// ordinary unknown-name error.
	foreignSymbols, foreignDiagnostics := collectForeignSymbols(program)
	if len(foreignDiagnostics) > 0 {
		result.Diagnostics = append(result.Diagnostics, foreignDiagnostics...)
		return result
	}
	functionNames := make(map[string]bool, len(program.Funcs))
	// calleeContracts is 07-09's pre-body callee-contract table
	// (superseding Phase 07's original functionIDs table): the same
	// name-keyed callee-resolution map functionIDs provided (a bare call's
	// callee name still resolves to the function's own semantic ID, D-07-01),
	// PLUS the callee's own declared parameter and return type constructors
	// -- built from the AST, so it genuinely precedes every body admission
	// (a declared type is a syntactic fact), unlike callSignatureTable
	// (below), whose Callable bit can only exist AFTER every body is
	// checked. Both tables are signature-only channels: neither carries a
	// *core.LinearBody or *core.Match, so SEM-05's body-blindness holds on
	// this path too.
	calleeContracts := buildCalleeContracts(program)
	for _, function := range program.Funcs {
		functionNames[function.Name] = true
	}

	for _, function := range program.Funcs {
		functionID := semanticID(program.Module, "fn", function.Name)
		if !function.Body.HasClosedVariant() {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error("core.invalid_body", function.Span, "function must have exactly one body variant"))
			continue
		}
		// OWN-04 scope fence: a match-bodied (S1) function — bare-arm or
		// arm-body/branch alike — must not return a borrowed view this phase.
		// This is deliberately checked before either match path below so the
		// rejection carries a named, span-bearing cause instead of falling
		// through to the generic type.return_mismatch the bare-match path
		// would otherwise produce for a return type it does not recognize.
		if function.Body.Linear == nil && function.ReturnOrigin != nil {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error(
				"ownership.match_borrowed_return_unsupported", function.ReturnOrigin.Span,
				"a match-bodied function cannot declare a borrowed return origin this phase",
			))
			continue
		}
		if function.Body.Linear != nil {
			var checked core.Function
			var diagnostics []diagnostic.Diagnostic
			var work int
			var callSpans map[string]diagnostic.Span
			if hasTryCall(function.Body.Linear) {
				checked, diagnostics, work = checkFallibleLinear(functionID, function, foreignSymbols, functionNames, types)
			} else {
				checked, diagnostics, work, callSpans = checkLinear(program.Module, functionID, function, calleeContracts, foreignSymbols)
			}
			result.Work += work
			result.Diagnostics = append(result.Diagnostics, diagnostics...)
			if len(diagnostics) == 0 {
				for opID, span := range callSpans {
					spanByOperationID[opID] = span
				}
				result.Program.Schema = core.Schema1
				result.Program.Functions = append(result.Program.Functions, checked)
				// D-05-01: alias facts are derivable only for the plain
				// straight-line shape checkLinear itself produces (Blocks
				// empty, no Match, no ForeignContract) -- checkFallibleLinear's
				// foreign-call shape (Blocks > 0) and any branch-shaped
				// function never reach this append site with those fields
				// clear, so this guard is never redundant, just explicit.
				if checked.Linear != nil && checked.Match == nil && checked.ForeignContract == nil && len(checked.Linear.Blocks) == 0 {
					endpoints, endpointWork := aliasFactEndpoints(checked.ID, checked.Linear.Operations)
					facts, factWork := deriveAliasFacts(checked, checked.Linear, endpoints)
					result.AliasFacts = append(result.AliasFacts, facts...)
					result.Work += endpointWork + factWork
				}
			}
			continue
		}
		matchID := semanticID(program.Module, "match", function.Name)
		dataType, ok := types[function.Parameter.Type.Constructor]
		if !ok {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error("type.unknown", function.Parameter.Span, "unknown parameter type"))
			continue
		}
		if !sameType(function.ReturnType, function.Parameter.Type) {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error("type.return_mismatch", function.Span, "S1 match result must have the parameter type"))
			continue
		}
		if function.Body.Scrutinee != function.Parameter.Name {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error("name.unknown_scrutinee", function.Body.Span, "match scrutinee is not the function parameter"))
			continue
		}
		for _, arm := range function.Body.Arms {
			if !arm.HasClosedVariant() {
				result.Diagnostics = append(result.Diagnostics, diagnostic.Error("core.invalid_body", arm.Span, "match arm must have exactly one value form"))
				continue
			}
		}
		hasArmBody := false
		for _, arm := range function.Body.Arms {
			if arm.Body != nil {
				hasArmBody = true
				break
			}
		}
		if hasArmBody {
			checked, diagnostics, work, callSpans := checkBranch(program.Module, functionID, matchID, function, dataType, sealedNames(types), calleeContracts, foreignSymbols)
			result.Work += work
			result.Diagnostics = append(result.Diagnostics, diagnostics...)
			if len(diagnostics) == 0 {
				result.Program.Schema = core.Schema1
				result.Program.Functions = append(result.Program.Functions, checked)
				for opID, span := range callSpans {
					spanByOperationID[opID] = span
				}
			}
			continue
		}
		seen := make(map[string]bool)
		arms := make([]core.MatchArm, 0, len(function.Body.Arms))
		for index, arm := range function.Body.Arms {
			if seen[arm.Pattern] {
				result.Diagnostics = append(result.Diagnostics, diagnostic.Error("match.subsumed", arm.Span, "alternative is already matched"))
				continue
			}
			if !contains(dataType.Alternatives, arm.Pattern) {
				result.Diagnostics = append(result.Diagnostics, diagnostic.Error("match.unreachable", arm.Span, "pattern is not an alternative of the scrutinee type"))
				continue
			}
			if !contains(dataType.Alternatives, arm.Value) {
				result.Diagnostics = append(result.Diagnostics, diagnostic.Error("type.invalid_variant", arm.Span, "match result is not an alternative of the return type"))
				continue
			}
			seen[arm.Pattern] = true
			arms = append(arms, core.MatchArm{
				ID: fmt.Sprintf("%s:arm:%d", functionID, index), EdgeID: fmt.Sprintf("%s:edge:%s", matchID, arm.Pattern),
				Pattern: arm.Pattern, Value: arm.Value,
			})
		}
		missing := make([]string, 0)
		for _, alternative := range dataType.Alternatives {
			if !seen[alternative] {
				missing = append(missing, alternative)
			}
		}
		if len(missing) > 0 {
			sort.Strings(missing)
			causes := make([]diagnostic.Cause, 0, len(missing))
			for _, name := range missing {
				causes = append(causes, diagnostic.Cause{Kind: "missing_alternative", Detail: name})
			}
			// A single missing alternative has one mechanical, source-level
			// fix -- IF the present arms already establish a safe value to
			// guess: insert a self-mapping arm ("alt => alt") immediately
			// after the last present arm, but ONLY when every present arm
			// already maps its own pattern to itself (allArmsSelfMap). alt
			// is always a member of dataType.Alternatives (the missing set
			// is drawn from exactly that list), so the appended arm is
			// always a legal match result, but "which value" is only a safe
			// guess, not a certainty, when the arms already in source
			// establish that identity-mapping convention -- otherwise (as
			// in testdata/phase1/non_exhaustive.lang's Off => On, which
			// does NOT self-map) there is no principled way to guess the
			// missing arm's target, and offering one anyway would be
			// exactly the fail-open shape this project refuses. This
			// condition is also what keeps lang.diagnostic/0's frozen Phase
			// 1 golden (TestPhase1DiagnosticGoldenUnchanged) byte-identical:
			// diagnostic.Error (schema /0) is used whenever no repair
			// applies, and ErrorWithRepairs (schema /1) always changes the
			// diagnostic's schema even with zero repairs attached, so the
			// two paths must never both reach the same call for the same
			// input (D-06-24/D-06-25's match injector target; D-06-31
			// "Phase 1 evidence bytes remain byte-identical"). More than one
			// missing alternative has no equally uncontroversial single
			// edit either, so it too stays classification-only.
			if len(missing) == 1 && len(function.Body.Arms) > 0 && allArmsSelfMap(function.Body.Arms) {
				insertion := function.Body.Arms[len(function.Body.Arms)-1].Span.End
				insertionSpan := diagnostic.Span{Start: insertion, End: insertion}
				alt := missing[0]
				result.Diagnostics = append(result.Diagnostics, diagnostic.ErrorWithRepairs(
					"match.non_exhaustive", function.Body.Span, "match does not cover every alternative", causes,
					diagnostic.Repair{
						Kind:          "add_missing_arm",
						Detail:        alt,
						Span:          &insertionSpan,
						Replacement:   fmt.Sprintf("\n    %s => %s", alt, alt),
						Applicability: diagnostic.ApplicabilityMachineApplicable,
					},
				))
			} else {
				result.Diagnostics = append(result.Diagnostics, diagnostic.Error(
					"match.non_exhaustive", function.Body.Span, "match does not cover every alternative", causes...,
				))
			}
		}
		result.Program.Functions = append(result.Program.Functions, core.Function{
			ID: functionID, Name: function.Name,
			EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
			Parameter:  core.Parameter{ID: semanticID(program.Module, "parameter", function.Name+"."+function.Parameter.Name), Name: function.Parameter.Name, Type: function.Parameter.Type.Constructor},
			ReturnType: function.ReturnType.Constructor,
			Match:      &core.Match{ID: matchID, PointID: functionID + ":point:match", Scrutinee: function.Body.Scrutinee, Arms: arms},
			Span:       function.Span,
		})
	}
	if len(result.Diagnostics) == 0 {
		// D-07-29/D-07-41: check's own independent re-derivation of the
		// three CalleeID invariants, from its own emission-time bookkeeping
		// (the function ID table this func just built) -- never by
		// consulting corevalidate. Run only once every function checked
		// clean, over the WHOLE assembled program, so a call that resolves
		// to a function declared later in source (declaration order is not
		// admission order) is still recognized.
		if invariant := verifyCallInvariants(result.Program.Functions); invariant != nil {
			result.Diagnostics = append(result.Diagnostics, *invariant)
		}
	}
	// signatureTable/signatureTableErr are hoisted out of the
	// verifyCallableRefusal admission arm below so the interprocedural pass
	// (D-08-12) can reuse the SAME build rather than invoking
	// buildCallSignatureTable a second time --
	// callSignatureTableBuildObserved (D-07-34's own ordering-
	// instrumentation seam) asserts exactly one build event per
	// check.Program call, which a second build would break
	// (TestCallSignatureTableBuiltBeforeCallableAdmissionRuns).
	var signatureTable callSignatureTable
	var signatureTableBuilt bool
	if len(result.Diagnostics) == 0 && !verifyCallInvariantsSeam {
		// D-07-42: sharing verifyCallInvariantsSeam here (rather than
		// gating only on len(result.Diagnostics)) keeps this admission arm
		// and verifyCallInvariants' resolves-to-a-declared-function
		// re-derivation failing TOGETHER under the same seeded mutation
		// (Task 2's TestVerifyCallInvariantsSeamRestoresBothRefusals): with
		// the seam engaged, resolveCallBinding's own fallback branch admits
		// a call whose CalleeID is a synthetic non-ID (the raw source
		// name, never a real function's ID), which this table would
		// correctly find absent and refuse -- a DIFFERENT, newer refusal
		// than the one that pre-existing test seam was written to
		// suppress. Suppressing this admission arm too, for the exact
		// span the seam already claims to simulate ("what a removed
		// resolves-to-a-declared-function check would let through"), keeps
		// that claim honest rather than letting this later plan's own
		// independent gate accidentally catch what the earlier seam was
		// built to demonstrate slipping through.
		//
		// D-07-34/SEM-06: the pre-body signature pass and the "call"
		// admission arm's Callable consult, run as their own final gate
		// over the WHOLE assembled program -- exactly where
		// verifyCallInvariants above already runs its own post-build
		// re-derivation. This is deliberate, not incidental: Callable
		// (originvalidate.PublishProblemsFor, D-07-31/D-07-32) is a
		// predicate over a function's CHECKED body (it recomputes return
		// origin from core.LinearOperation facts), so it cannot exist
		// before every function's own body has been individually admitted
		// -- there is no such thing as a "signature" that precedes a
		// function's own admission when the signature's Callable bit is
		// itself derived from that admission's output. What "before any
		// body admission" (D-07-34's own words) means here, and can only
		// coherently mean, is: before the ONE admission decision this
		// table exists to gate -- the "call" arm's Callable consult below
		// -- which is why that decision is deliberately NOT inlined into
		// resolveCallBinding (where CalleeID resolution already happens):
		// resolveCallBinding's per-function pass never reads Callable at
		// all, so nothing about admitting any individual function's body
		// ever depends on another function's publication status, and the
		// table this builds is used exactly once, immediately after it is
		// built, by nothing but the loop directly below.
		// 07-08, D-07-38: BuildInterface now runs callgraph.Order internally
		// before computing any ClosureDigest (the chain terminates only on
		// a DAG), so a cyclic program makes buildCallSignatureTable fail
		// here too -- reachable ONLY for a cycle, since verifyCallInvariants
		// above already required every CalleeID to resolve to a declared
		// function before this point is ever reached. Skip
		// verifyCallableRefusal entirely on that failure (leaving
		// result.Diagnostics empty) rather than admitting the empty-table
		// fallback's fail-closed-refuse-every-call behavior to run: that
		// would emit core.callee_not_callable for the first OpCall found,
		// masking the correct core.call_graph_cycle diagnostic the
		// dedicated gate below is about to construct with the right causes
		// and spans. Falling through with zero diagnostics is safe: the
		// gate below independently re-runs callgraph.Order and refuses the
		// SAME cycle properly.
		table, tableErr := buildCallSignatureTable(result.Program)
		if tableErr == nil {
			signatureTable, signatureTableBuilt = table, true
			if diag := verifyCallableRefusal(result.Program.Functions, table); diag != nil {
				result.Diagnostics = append(result.Diagnostics, *diag)
			}
		}
	}
	if len(result.Diagnostics) == 0 && !disableCallGraphCycleRefusalForTest {
		// D-07-14: run callgraph.Order over check's OWN completed
		// in-memory core.Program and refuse BEFORE returning it. A cyclic
		// core.Program therefore exists only as an ephemeral local inside
		// this function -- never returned, serialized, cached,
		// interpreted, or lowered. There is no AST-side graph.
		if diag := checkCallGraphAcyclic(result.Program, spanByOperationID); diag != nil {
			result.Diagnostics = append(result.Diagnostics, *diag)
			result.Program = core.Program{}
		}
	}
	if len(result.Diagnostics) == 0 {
		// D-08-12, D-08-27 interim rule: the interprocedural loan-liveness
		// pass runs STRICTLY here -- after callgraph.Order has already
		// proven the program acyclic (the gate immediately above) and after
		// every function's own intraprocedural admission has already passed
		// (nothing above this point has appended a diagnostic) -- so no
		// program is ever judged by both the intraprocedural and the
		// interprocedural liveness law for the same fact. Reuses Task 1's
		// signatureTable when the admission arm above already built one
		// (the common case); only rebuilds when it was skipped
		// (verifyCallInvariantsSeam, D-07-42's own test-only seam) --
		// callSignatureTableBuildObserved asserts exactly one build event
		// per check.Program call in production, so a second, unconditional
		// build here would break that invariant.
		if !signatureTableBuilt {
			if table, tableErr := buildCallSignatureTable(result.Program); tableErr == nil {
				signatureTable, signatureTableBuilt = table, true
			}
		}
		if signatureTableBuilt && !disableInterproceduralLoanLivenessForTest {
			summaries, summaryWork := buildInterproceduralSummaries(result.Program, signatureTable)
			result.Work += summaryWork
			result.Diagnostics = append(result.Diagnostics, checkInterproceduralLoanLiveness(result.Program, summaries, spanByOperationID)...)
		}
	}
	return result
}

// disableInterproceduralLoanLivenessForTest is Phase 09's own D-09-24/D-09-25
// fault-injection seam (QLT-08): when true, check's own post-assembly
// interprocedural loan-liveness pass (checkInterproceduralLoanLiveness) is
// skipped entirely -- check.Program admits a program it would otherwise
// refuse with check.interprocedural_loan_liveness -- so a same-package test
// can prove corevalidate's own, independently written peer
// (corevalidate_peer_liveness.go) still refuses the SAME endpoint fact
// wholly on its own, mirroring verifyCallableRefusalSeam's own established
// shape (check.go:1183) generalized to the liveness fact. Unexported,
// same-package-test-only, restored via defer in every test that engages it
// -- never an exported package-level mutable var on a production path
// (D-07-42).
var disableInterproceduralLoanLivenessForTest = false

// disableCallGraphCycleRefusalForTest is 07-07 Task 3 Test 1's own D-07-42
// independent-disable seam: when true, check's callgraph-based cycle
// refusal is skipped entirely, so a genuinely cyclic core.Program is
// returned as-is (never cleared) -- proving corevalidate's own,
// independently written cycle peer (07-07) still refuses it on its own,
// wholly without check's help. Unexported, same-package-test-only,
// restored via defer in every test that engages it -- never an exported
// package-level mutable var on a production path (D-07-42).
var disableCallGraphCycleRefusalForTest = false

// checkCallGraphAcyclic runs callgraph.Order over program and, on a
// discovered cycle, builds the core.call_graph_cycle diagnostic ratified
// at the 07-06 checkpoint (D-07-15): code core.call_graph_cycle, message
// "call graph contains a cycle; recursion is refused", causes ordered
// cycle_length (the TRUE member count, never the truncated one) followed
// by up to callgraph.MaxCycleCauses cycle_member causes, each carrying the
// span of that member's own outgoing edge (D-07-35: projected here from
// the OpCall operation ID callgraph's witness names, via spanByOperationID
// -- never a span callgraph or core.LinearOperation itself carries), and,
// on overflow, exactly one truncated cause naming
// protocol.TruncatedCallCycleBound. The diagnostic's own Primary is the
// span of the closing edge. No repairs (a cycle has no local, mechanical
// edit). Returns nil when the program is acyclic.
func checkCallGraphAcyclic(program core.Program, spanByOperationID map[string]diagnostic.Span) *diagnostic.Diagnostic {
	if _, err := callgraph.Order(program); err != nil {
		if cycle, ok := callgraph.CycleError(err); ok {
			members := cycle.Members()
			edgeOperationIDs := cycle.MemberEdgeOperationIDs()
			causes := []diagnostic.Cause{{Kind: "cycle_length", Detail: fmt.Sprintf("%d", len(members))}}
			emitted := len(members)
			if emitted > callgraph.MaxCycleCauses {
				emitted = callgraph.MaxCycleCauses
			}
			for i := 0; i < emitted; i++ {
				span := spanByOperationID[edgeOperationIDs[i]]
				causes = append(causes, diagnostic.Cause{Kind: "cycle_member", Detail: members[i], Span: &span})
			}
			if len(members) > callgraph.MaxCycleCauses {
				causes = append(causes, diagnostic.Cause{Kind: "truncated", Detail: callgraph.TruncatedCycleBound})
			}
			primary := spanByOperationID[cycle.ClosingOperationID()]
			diag := diagnostic.Error(core.CallGraphCycle, primary, "call graph contains a cycle; recursion is refused", causes...)
			return &diag
		}
		if _, ok := callgraph.UnresolvedCalleeError(err); ok && verifyCallInvariantsSeam {
			// verifyCallInvariantsSeam (D-07-41/D-07-42, Task 2 Test 6)
			// suppresses check's OWN resolves-to-a-declared-function
			// re-derivation everywhere it appears in this package,
			// including this independent callgraph-side re-derivation,
			// so the seam's single seeded mutation keeps every admission
			// arm it names failing TOGETHER rather than leaving this one
			// still refusing.
			return nil
		}
		// A forged or otherwise-inconsistent program surfaced an
		// unresolved-callee edge callgraph independently re-derived
		// (D-07-45) -- reuse its own stable code rather than inventing a
		// new one here.
		if coder, ok := err.(interface{ Code() string }); ok {
			diag := diagnostic.Error(coder.Code(), diagnostic.Span{}, err.Error())
			return &diag
		}
		diag := diagnostic.Error("core.call_graph_error", diagnostic.Span{}, err.Error())
		return &diag
	}
	return nil
}

// ---------------------------------------------------------------------
// Phase 08 Task 2: interprocedural loan liveness.
//
// The end-to-end tracer -- one interprocedural loan-liveness refusal
// travelling declared callee return contract -> in-memory summary bit ->
// forward canonicalization in derivePlaceLoans -> backward liveness -> a
// newly minted check.interprocedural_loan_liveness diagnostic -> the
// `lang check` CLI. Wires exactly ONE direction (a callee that returns a
// borrow of its own parameter, D-08-07); no transitivity, no bound -- those
// are later plans' expansion work.
// ---------------------------------------------------------------------

// interproceduralSummary is one function's per-callee interprocedural
// summary bit set (D-08-02/D-08-04/D-08-06): read-only once built, and
// derived EXCLUSIVELY from callSignatureTable's already-published
// core.FunctionSignature -- which structurally carries no Linear/Match
// field -- so consuming it can never reach a callee body (SEM-05, T-08-04).
type interproceduralSummary struct {
	// usesParam is 08-02's own derivation (D-08-01/D-08-02/D-08-03): true
	// iff the function's own checked body reads through its parameter (or a
	// place transitively derived from it) other than merely forwarding it
	// to its own terminating OpReturn, OR calls a callee whose OWN usesParam
	// is true (or is absent from summaries, the refusing direction) with a
	// parameter-derived argument. See deriveFunctionUsesParam.
	usesParam bool
	// returnsBorrowOfParam is true iff the callee's declared
	// core.FunctionSignature.Return.Mode is "shared" or "exclusive", false
	// iff "owned" (D-08-07 Task 2(a)) -- the ONE direction this plan wires.
	returnsBorrowOfParam bool
	// returnMode is the callee's own declared Return.Mode string, consulted
	// ONLY to build the ratified third cause's Detail
	// (<calleeID>:return.mode=<Mode>, D-08-20/23/24). Disclosing it is safe
	// (T-08-03): it is copied verbatim from the callee's DECLARED type, so
	// disclosure reveals nothing about the callee's control flow.
	returnMode string
	// parameterMode is the callee's own declared
	// core.FunctionSignature.Parameters[0].Mode string (08-02 Task 2),
	// consulted ONLY to build the backward-direction third cause's Detail
	// (<calleeID>:parameters[0].mode=<Mode>). Same disclosure-safety
	// argument as returnMode: copied verbatim from the callee's DECLARED
	// contract.
	parameterMode string
}

// interproceduralSummaryTable is a same-package, read-only view over every
// declared function's interproceduralSummary, mirroring callSignatureTable's
// own immutability-by-construction invariant (check.go:590-597): the only
// way to read it is lookup, there is no setter, and its zero value's lookup
// always misses -- a table that somehow failed to build treats every callee
// as "no interprocedural facts" (fail-closed toward the pre-Phase-08
// behaviour), never toward admitting a false direction.
type interproceduralSummaryTable struct {
	summaries map[string]interproceduralSummary
}

func (t interproceduralSummaryTable) lookup(calleeID string) (interproceduralSummary, bool) {
	summary, ok := t.summaries[calleeID]
	return summary, ok
}

// MaxInterproceduralLoanCauses bounds the interprocedural loan-liveness
// diagnostic's own cause list, mirroring callgraph.MaxCycleCauses. Unused by
// this plan's fixed three-cause template (D-08-23) -- Task 1's ratified
// shape never grows past three -- but declared now so 08-04's truncation
// work (the loan_liveness_bound refusal) has a named sibling constant to
// reference, per Task 1(d)/(d)'s instruction to declare both together.
const MaxInterproceduralLoanCauses = 32

// TruncatedInterproceduralLoanBound is the cause-list truncation marker
// constant ratified at this plan's Task 1 checkpoint, declared in package
// check (never protocol) alongside MaxInterproceduralLoanCauses, per the
// recorded check -> protocol -> interp -> [test] -> check import cycle.
const TruncatedInterproceduralLoanBound = "check.interprocedural_loan_bound"

// buildInterproceduralSummaries derives, for every declared function,
// exactly one interproceduralSummary entry -- read from table's already-
// published core.FunctionSignature for returnsBorrowOfParam/returnMode/
// parameterMode, and from the function's OWN checked body (once, via
// deriveFunctionUsesParam) for usesParam. program is, by the time
// check.Program calls this, already proven acyclic (checkCallGraphAcyclic
// runs strictly before this); a callgraph.Order failure here is therefore
// unreachable in production and is handled defensively -- "no
// interprocedural facts" for every function -- rather than by panicking.
//
// callgraph.Order's own doc comment names its result "reverse postorder",
// but empirically (TestOrderSortsAdjacencyByCalleeID, and this package's
// own TestSummaryDerivationIsOnePassPerFunction) that order lists a CALLER
// before every function it calls -- the opposite of what D-08-03's
// callee-before-caller transitivity requires: deriveFunctionUsesParam's
// OpCall case reads summaries.lookup(calleeID) and needs that entry
// ALREADY FINAL, which is only true if every callee has been derived
// before its caller. This loop therefore walks callgraph.Order's result
// BACKWARD (last-visited-first), which is the genuine callee-before-caller
// traversal Phase 08 needs -- iterating Order's raw output forward here
// would silently leave every transitive UsesParam bit false (or fail-safe
// true only for the immediately-unresolved case), an entry-order bug this
// package's own summary consumers would never surface as a build failure.
func buildInterproceduralSummaries(program core.Program, table callSignatureTable) (interproceduralSummaryTable, int) {
	if buildInterproceduralSummariesObserved != nil {
		buildInterproceduralSummariesObserved()
	}
	summaries := make(map[string]interproceduralSummary, len(program.Functions))
	result := interproceduralSummaryTable{summaries: summaries}
	order, err := callgraph.Order(program)
	if err != nil {
		return result, 0
	}
	functionByID := make(map[string]core.Function, len(program.Functions))
	for _, function := range program.Functions {
		functionByID[function.ID] = function
	}
	work := 0
	for i := len(order) - 1; i >= 0; i-- {
		functionID := order[i]
		work++
		var summary interproceduralSummary
		if signature, ok := table.lookup(functionID); ok {
			if interproceduralConsultObserved != nil {
				interproceduralConsultObserved(functionID, "return.mode")
			}
			summary.returnMode = signature.Return.Mode
			summary.returnsBorrowOfParam = signature.Return.Mode == "shared" || signature.Return.Mode == "exclusive"
			if len(signature.Parameters) > 0 {
				if interproceduralConsultObserved != nil {
					interproceduralConsultObserved(functionID, "parameters[0].mode")
				}
				summary.parameterMode = signature.Parameters[0].Mode
			}
		}
		if function, ok := functionByID[functionID]; ok {
			usesParam, usesWork := deriveFunctionUsesParam(function, result)
			summary.usesParam = usesParam
			work += usesWork
		}
		summaries[functionID] = summary
	}
	return result, work
}

// buildInterproceduralSummariesObserved is Task 3's D-08-12/D-08-37
// ordering-instrumentation seam, mirroring callSignatureTableBuildObserved:
// when non-nil, invoked once, synchronously, at the start of every
// buildInterproceduralSummaries call -- before it does anything else -- so
// a same-package test can prove a fresh memo is built on EVERY
// check.Program invocation (never persisted or reused across calls). nil in
// production: zero cost, zero allocation.
var buildInterproceduralSummariesObserved func()

// interproceduralConsultObserved is 08-03 Task 3's D-08-26 disclosure-proof
// instrumentation seam, mirroring callSignatureTableLookupObserved's own
// shape (check.go:1106-1114): when non-nil, invoked with the callee's own
// function ID and the exact signature-field name read, at every point
// buildInterproceduralSummaries consults a callee-signature field --
// currently exactly two call sites, both inside the block immediately
// below, reading "return.mode" and "parameters[0].mode" respectively. This
// is deliberately the SAME two-call-site set the ratified diagnostic
// template's own cause 3 Detail string draws from (D-08-20/D-08-23):
// nothing else in this derivation ever names a callee-signature field. nil
// in production: zero cost, zero allocation.
var interproceduralConsultObserved func(calleeID, field string)

// deriveFunctionUsesParam is 08-02 Task 1's own derivation (D-08-01/D-08-02):
// whether function's checked body uses its own declared parameter, computed
// exactly once here (never re-walked at a call site -- summaries is
// consulted, never recursed into) and transitive through relays via
// summaries, which by the time this runs already carries the FINAL entry
// for every callee (buildInterproceduralSummaries' own callee-before-caller
// traversal, D-08-03).
//
// Operations MUST be consumed in PROGRAM order for this to stay linear
// (D-08-11): a place is added to the parameter-derived set the first time
// an already-derived operation names it as a TARGET, so program order lets
// every downstream operation see its own antecedent already derived by the
// time the scan reaches it -- one full pass propagates the WHOLE chain, and
// a second pass then confirms no further insertion is possible (a fixpoint
// in exactly two passes). Reversed, an operation's antecedent has not yet
// been derived when the scan first reaches it, so the propagation can only
// advance by one additional place PER PASS -- O(n) passes of O(n) work
// each, quadratic in body length. Spike S-006 iteration 5 measured this
// directly: 4.0 work units per operation, flat from k=8 to k=512, in
// program order, against 12.4 -> 767.0 per operation for the identical
// chain listed in reverse -- a 192x penalty at k=512
// (TestSummaryDerivationRequiresProgramOrder pins both halves of this
// claim). This bound is inherited from today's emission order, not a
// property of the mechanism: any future pass that reorders operations
// before summary derivation hands the bound back.
//
// An operation whose SourceID is already parameter-derived counts as a USE
// of the parameter when it is anything other than an OpCall or the
// function's own terminating OpReturn (a plain read, or a read of
// something derived from the parameter), or when it IS an OpCall whose
// callee's own usesParam (via summaries.lookup) is true -- a callee absent
// from summaries counts as usesParam == true, the refusing direction, never
// the permitting one (an unresolved callee must never look "safer" than a
// resolved one).
//
// The OpReturn exemption below (D-10-27/D-10-29) covers a ZERO-HOP DIRECT
// return of the parameter-derived place ONLY -- the switch below matches on
// operation.Kind per operation, so it exempts only the terminating
// `case core.OpReturn:` arm itself, never an arbitrary identity chain that
// merely happens to end in one. An intermediate OpMove/OpCopy hop (e.g.
// `taken = take param; taken`) hits the `default:` arm on its OWN pass and
// sets usesParam = true before the scan ever reaches the OpReturn --
// PHASE-10-DEBT.md's D-10-27 re-executed the narrower reading this comment
// once claimed (excluding OpMove/OpCopy identity-forwards from the default
// arm) and found it falsified by eight currently-passing tests, including
// TestInterproceduralLivenessTwinPatternB itself. Only a literal
// `return param` with nothing between them is exempt; every real
// intermediate hop is a use. D-08-01's leaf-forwards-to-its-own-return case
// is this exact zero-hop shape, not a multi-hop one.
//
// Returns the derived bit and the total counted work: one unit per
// operation inspected per pass, plus one unit per genuine new set
// insertion -- honest, per-operation counting (D-05), never a flat
// per-function unit, so a reintroduced re-walk shows up here as work
// growing faster than operation count.
func deriveFunctionUsesParam(function core.Function, summaries interproceduralSummaryTable) (bool, int) {
	if deriveFunctionUsesParamObserved != nil {
		deriveFunctionUsesParamObserved(function.ID)
	}
	if function.Linear == nil {
		return false, 0
	}
	derived := make(map[string]bool, len(function.Linear.Operations)+1)
	if function.Parameter.ID != "" {
		derived[function.Parameter.ID] = true
	}
	usesParam := false
	work := 0
	for {
		changed := false
		for _, operation := range function.Linear.Operations {
			work++ // one unit per operation inspected, every pass
			if !derived[operation.SourceID] {
				continue
			}
			switch operation.Kind {
			case core.OpCall:
				if summary, ok := summaries.lookup(operation.CalleeID); !ok || summary.usesParam {
					usesParam = true
				}
			case core.OpReturn:
				// A plain forward of a parameter-derived place to the
				// function's own terminating return is not itself a use
				// (D-08-01's leaf-forwards-to-its-own-return case).
			default:
				usesParam = true
			}
			if operation.TargetID != "" && !derived[operation.TargetID] {
				derived[operation.TargetID] = true
				work++ // one unit per genuine new set insertion
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	return usesParam, work
}

// deriveFunctionUsesParamObserved is Task 1's own D-08-03 ordering
// instrumentation seam: when non-nil, invoked once, synchronously, at the
// start of every deriveFunctionUsesParam call, naming the function being
// derived -- so a same-package test can prove each function is derived
// EXACTLY ONCE per check.Program invocation regardless of how many callers
// it has (a shared leaf reached through a diamond-shaped call graph, for
// instance). nil in production: zero cost, zero allocation.
var deriveFunctionUsesParamObserved func(functionID string)

// cfgBlocksForFunction rebuilds the minimal []cfgBlockSpec shape
// loanLivenessFixpoint/materializeLoanEndpoints consume from an already-
// checked core.Function's own serialized Linear body -- reusing the exact
// shape checkBranch already builds for its own arm blocks, never
// re-deriving liveness from the AST. A straight-line function's
// Linear.Blocks stays empty by construction (LinearBody's own doc comment,
// core.go:715-721): such a function's whole Operations slice is treated as
// one synthetic, no-successor block, mirroring computeLoanLastUses' own
// single-block shape (check.go's "shadow:block:straight").
func cfgBlocksForFunction(function core.Function) []cfgBlockSpec {
	if function.Linear == nil {
		return nil
	}
	if len(function.Linear.Blocks) == 0 {
		return []cfgBlockSpec{{id: function.ID + ":block:straight", operations: function.Linear.Operations, successors: nil}}
	}
	operationsByID := make(map[string]core.LinearOperation, len(function.Linear.Operations))
	for _, operation := range function.Linear.Operations {
		operationsByID[operation.ID] = operation
	}
	blocks := make([]cfgBlockSpec, 0, len(function.Linear.Blocks))
	for _, block := range function.Linear.Blocks {
		operations := make([]core.LinearOperation, 0, len(block.OperationIDs))
		for _, id := range block.OperationIDs {
			if operation, ok := operationsByID[id]; ok {
				operations = append(operations, operation)
			}
		}
		blocks = append(blocks, cfgBlockSpec{id: block.ID, operations: operations, successors: block.Successors})
	}
	return blocks
}

// checkInterproceduralLoanLiveness is Task 2's new post-acyclicity pass
// (D-08-12). For every already-checked core.Function it re-derives loan
// liveness over that function's own serialized core.LinearOperation stream,
// this time summary-aware, and refuses an OpMove that transfers ownership
// of a loan's owner place while that loan is still live because a call the
// summary table proves extends it (a borrow-of-param return) has not yet
// had its own last use. It reads summaries and spanByOperationID only --
// never a callee's Linear or Match field (T-08-04). Emits at most one
// diagnostic per function, choosing the earliest offending OpMove by
// operation index (deterministic output).
//
// The liveness law is declared FINAL as of Phase 08's mandatory mid-phase
// gate (08-06-PLAN.md, PHASE-08-DEBT.md): criterion 1's nine-fixture
// adversarial corpus plus relay_escort_witness.lang, and criterion 3's
// measured <=1.2 growth exponent across five call-graph shapes (ratified in
// internal/compiler/session/qlt02_budget_manifest.json), were both
// adjudicated from code-level evidence before this scope was closed. Four
// facts make up that final scope, and a reviewer should be able to find all
// four here rather than reconstructing them from the phase's plan history:
//
//  1. The law consults EXACTLY two callee-signature bits and nothing else:
//     the callee's declared core.FunctionSignature.ReturnsBorrowOfParam
//     (`return.mode`) and its derived interproceduralSummary.UsesParam
//     (`parameters[0].mode`). No other field of the callee's signature, and
//     no field outside the signature, ever participates in the liveness
//     answer (TestInterproceduralDisclosedFieldSet asserts this as a closed
//     set, table-driven over the whole testdata/phase08 corpus).
//  2. Consumption at each call site (this function, and
//     interproceduralLoanConflict) reads ONLY the summary table and
//     core.FunctionSignature -- both of which are structurally body-free by
//     construction (OWN-06's recorded interpretation, D-08-05). The callee's
//     own core.Function.Linear or .Match is never read at a call site; a
//     caller can never "see" a callee's body through this law.
//  3. Production of those same two bits (buildInterproceduralSummaries)
//     walks each callee's OWN already-checked body exactly once,
//     callee-before-caller, strictly before any caller's admission consults
//     it -- a one-pass, two-tier production/consumption split, not a
//     re-walk triggered per call site.
//  4. This code lives in the check.* namespace, not core.*, and it STAYS
//     there. D-08-21 once scheduled it for promotion to core.* in Phase 09,
//     conditioned on both peers agreeing on-code. D-09-31 formally
//     SUPERSEDED that commitment: the trigger was never achievable, because
//     the two mechanisms cannot fail the same way (D-08-17) -- check runs a
//     backward worklist, corevalidate a forward reachability closure. The
//     corpus shows it directly: corevalidate refuses the relay-escort
//     witness via core.move_while_borrowed while check refuses the same
//     program via check.interprocedural_loan_liveness, and session's peer
//     gate treats that as two honest independent refusals. Promoting would
//     force check's three-role cause template onto a closure that has no
//     natural "call that extended the loan" -- reverse-engineering one
//     peer's shape into the other, the exact anti-pattern this phase
//     exists to prevent. Divergent codes per peer are already shipped
//     precedent (check.call_argument_type_mismatch /
//     core.CallArgumentTypeMismatch, D-07-46).
//
// The intraprocedural ownership.* loan-liveness law is deliberately NOT
// retired in this phase (D-08-27's interim rule, reconfirmed at the
// mid-phase gate): this pass runs LAST, as a separate walk over the
// completed core.Program after every function's intraprocedural admission
// has already passed, so no program is ever judged by both laws for the
// same fact. Retirement is Phase 09's charter, alongside corevalidate's
// independent peer. PHASE-08-DEBT.md is the register carrying this phase's
// declared deferrals (the accepted-program disclosure vehicle, the OWN-09
// document conflict, the two accepted-program peer-divergence fixtures, and
// Pattern B's real-fixture scope limit) -- consult it before assuming any
// open question here was silently dropped.
func checkInterproceduralLoanLiveness(program core.Program, summaries interproceduralSummaryTable, spanByOperationID map[string]diagnostic.Span) []diagnostic.Diagnostic {
	var diagnostics []diagnostic.Diagnostic
	for _, function := range program.Functions {
		if function.Linear == nil || len(function.Linear.Operations) == 0 {
			continue
		}
		blocks := cfgBlocksForFunction(function)
		fixpoint, diag := loanLivenessFixpoint(function.ID, blocks, summaries, function.Span)
		if diag != nil {
			// A CFG cycle here would already have been refused by this same
			// function's own intraprocedural admission, which has already
			// passed by the time this pass runs (D-08-12) -- unreachable in
			// production; skip rather than assume-and-panic.
			continue
		}
		edgeID := func(fromBlockID, toBlockID string) string { return fromBlockID + "->" + toBlockID }
		endpoints := materializeLoanEndpoints(function.ID, blocks, edgeID, fixpoint, summaries)

		operationIndexByID := make(map[string]int, len(function.Linear.Operations))
		for index, operation := range function.Linear.Operations {
			operationIndexByID[operation.ID] = index
		}
		lastUseIndexByLoan := make(map[string]int, len(endpoints))
		for _, endpoint := range endpoints {
			if endpoint.Kind != "point" {
				continue
			}
			index, ok := operationIndexByID[endpoint.AfterOperationID]
			if !ok {
				continue
			}
			if existing, seen := lastUseIndexByLoan[endpoint.LoanID]; !seen || index > existing {
				lastUseIndexByLoan[endpoint.LoanID] = index
			}
		}

		chain := derivePlaceLoans(function.Linear.Operations, summaries)

		// D-09-08/D-09-09/Task 2(d): testOnlyForceUniformLoanJoin re-attached
		// here, gated to branch-bodied (arm) functions ONLY -- mirroring the
		// seam's pre-restructure scope (analyzeArmBody, never
		// analyzeStraightLine). Force every loan's last use to its OWN arm
		// block's own last operation (the arm's own join point), exactly as
		// if edge-specific placement had been deleted and every loan ended
		// uniformly at the join regardless of which edge actually needs it.
		if testOnlyForceUniformLoanJoin && function.Match != nil {
			blockOfOperationID := make(map[string]string, len(function.Linear.Operations))
			blockLastIndex := make(map[string]int, len(blocks))
			for _, block := range blocks {
				maxIndex := -1
				for _, operation := range block.operations {
					blockOfOperationID[operation.ID] = block.id
					if index, ok := operationIndexByID[operation.ID]; ok && index > maxIndex {
						maxIndex = index
					}
				}
				blockLastIndex[block.id] = maxIndex
			}
			for loanID, borrow := range chain.borrowOperation {
				if blockID, ok := blockOfOperationID[borrow.ID]; ok {
					if last, ok2 := blockLastIndex[blockID]; ok2 && last >= 0 {
						lastUseIndexByLoan[loanID] = last
					}
				}
			}
		}

		// Iterate borrowed loans in sorted ID order, never in Go map order.
		// This language permits multiple concurrent SHARED borrows of the same
		// owner place, so two distinct loans can tie on the same offending
		// OpMove index; the strict `index < offendingIndex` tie-break below
		// then keeps whichever the range happened to yield first. Sorting here
		// is the same guarantee the deleted lowering-time gate already made
		// for the intraprocedural diagnostics -- so two runs never disagree
		// about which loan a diagnostic blames.
		borrowedLoanIDs := make([]string, 0, len(chain.borrowOperation))
		for loanID := range chain.borrowOperation {
			borrowedLoanIDs = append(borrowedLoanIDs, loanID)
		}
		sort.Strings(borrowedLoanIDs)

		offendingIndex := -1
		var offendingMove core.LinearOperation
		var offendingLoanID string
		for index, operation := range function.Linear.Operations {
			if operation.Kind != core.OpMove {
				continue
			}
			for _, loanID := range borrowedLoanIDs {
				borrow := chain.borrowOperation[loanID]
				if borrow.SourceID != operation.SourceID {
					continue
				}
				lastUse, ok := lastUseIndexByLoan[loanID]
				if !ok || index >= lastUse {
					continue
				}
				if offendingIndex == -1 || index < offendingIndex {
					offendingIndex = index
					offendingMove = operation
					offendingLoanID = loanID
				}
			}
		}

		// D-09-08/D-09-09/Task 2(a): the earliest new loan (OpBorrowShared/
		// OpBorrowExclusive) whose owner already carries an unexpired,
		// access-conflicting loan -- ownership.borrow_conflict's own
		// extension into this pass, using the SAME lastUseIndexByLoan the
		// move scan above already derived (D-09-07: never a second liveness
		// law). "Unexpired at index" mirrors the deleted lowering-time
		// activeLoans' own expiry threshold exactly: a loan born at an
		// earlier index survives through (and including) the index equal to
		// its own last use.
		borrowConflictIndex := -1
		var offendingBorrow, blockingBorrow core.LinearOperation
		for index, operation := range function.Linear.Operations {
			if operation.Kind != core.OpBorrowShared && operation.Kind != core.OpBorrowExclusive {
				continue
			}
			newLoanID := operation.LoanID
			if newLoanID == "" {
				continue
			}
			var conflicting []string
			for _, loanID := range borrowedLoanIDs {
				if loanID == newLoanID {
					continue
				}
				existing := chain.borrowOperation[loanID]
				if existing.SourceID != operation.SourceID {
					continue
				}
				birthIndex, ok := operationIndexByID[existing.ID]
				if !ok || birthIndex >= index {
					continue
				}
				lastUse, ok2 := lastUseIndexByLoan[loanID]
				if !ok2 || lastUse < index {
					continue
				}
				if operation.Kind == core.OpBorrowShared && existing.Kind == core.OpBorrowShared {
					continue
				}
				conflicting = append(conflicting, loanID)
			}
			if len(conflicting) == 0 {
				continue
			}
			sort.Strings(conflicting)
			if borrowConflictIndex == -1 || index < borrowConflictIndex {
				borrowConflictIndex = index
				offendingBorrow = operation
				blockingBorrow = chain.borrowOperation[conflicting[0]]
			}
		}

		// Earliest offending event wins overall, mirroring the single forward
		// admission walk the deleted lowering-time code used to perform
		// before D-09-08's authorized deletion moved this decision here
		// (D-09-13).
		if borrowConflictIndex != -1 && (offendingIndex == -1 || borrowConflictIndex < offendingIndex) {
			blockingLastUse := lastUseIndexByLoan[blockingBorrow.LoanID]
			diagnostics = append(diagnostics, borrowConflictDiagnosticPostAssembly(
				offendingBorrow, blockingBorrow, lastUseSpanAt(function, blockingLastUse, spanByOperationID), placesByID(function), spanByOperationID,
			))
			continue
		}
		if offendingIndex == -1 {
			continue
		}
		borrowOperation, hasBorrow := chain.borrowOperation[offendingLoanID]
		if !hasBorrow {
			continue
		}
		// Two, mutually exclusive, ways this conflict can be interprocedural
		// (08-02 Task 2 widens the original forward-only check): the FORWARD
		// direction (D-08-07), where derivePlaceLoans propagated the loan
		// across a call's return onto a place read later than the move; or
		// the BACKWARD direction (D-08-08), where the offending loan's own
		// last use (as this pass's own summary-aware fixpoint computed it)
		// IS the call itself, because the callee's usesParam gated
		// blockLoanLiveness's chain-ancestor walk open. Checking forward
		// first mirrors D-08-07's own priority (the mechanism this phase
		// landed first); a conflict can only ever be attributed to one
		// direction, since extendedByCall and a call-shaped last use are
		// populated by disjoint call-site facts.
		if callOperation, extended := chain.extendedByCall[offendingLoanID]; extended {
			summary, _ := summaries.lookup(callOperation.CalleeID)
			detail := fmt.Sprintf("%s:return.mode=%s", callOperation.CalleeID, summary.returnMode)
			diagnostics = append(diagnostics, interproceduralLoanLivenessDiagnostic(offendingMove, borrowOperation, callOperation, detail, spanByOperationID))
			continue
		}
		if lastUse, ok := lastUseIndexByLoan[offendingLoanID]; ok && lastUse >= 0 && lastUse < len(function.Linear.Operations) {
			if candidate := function.Linear.Operations[lastUse]; candidate.Kind == core.OpCall {
				summary, _ := summaries.lookup(candidate.CalleeID)
				detail := fmt.Sprintf("%s:parameters[0].mode=%s", candidate.CalleeID, summary.parameterMode)
				diagnostics = append(diagnostics, interproceduralLoanLivenessDiagnostic(offendingMove, borrowOperation, candidate, detail, spanByOperationID))
				continue
			}
		}
		// Neither direction fired: a plain intraprocedural conflict. Before
		// D-09-08's authorized deletion, this was already caught at lowering
		// time by analyzeStraightLine/analyzeArmBody's own activeLoans state
		// machine; this pass is now the SOLE decision point (D-09-09) and
		// raises the identical ownership.move_while_borrowed diagnostic.
		diagnostics = append(diagnostics, moveWhileBorrowedDiagnosticPostAssembly(
			offendingMove, borrowOperation, lastUseSpanAt(function, lastUseIndexByLoan[offendingLoanID], spanByOperationID), spanByOperationID,
		))
	}
	return diagnostics
}

// lastUseSpanAt projects a loan's last-use OPERATION INDEX back to a span,
// reconstructed from the program-wide span channel D-07-35/D-08-22 already
// built (spanByOperationID) -- never from AST, which this pass never reads.
// A last use at the function's own terminator (OpReturn/OpDefect) has no
// CallSpans entry (lowering never records one for the terminator), so it
// falls back to the function's own span, mirroring the deleted lowering-time
// derivation's identical "used at result position" fallback.
func lastUseSpanAt(function core.Function, lastUseIndex int, spanByOperationID map[string]diagnostic.Span) diagnostic.Span {
	if function.Linear == nil || lastUseIndex < 0 || lastUseIndex >= len(function.Linear.Operations) {
		return function.Span
	}
	operation := function.Linear.Operations[lastUseIndex]
	if operation.Kind == core.OpReturn || operation.Kind == core.OpDefect {
		return function.Span
	}
	if span, ok := spanByOperationID[operation.ID]; ok {
		return span
	}
	return function.Span
}

// placesByID indexes a function's own Linear.Places by ID, so a post-assembly
// diagnostic can recover a place's SOURCE-LEVEL surface name (core.Place.Name
// preserves it) without reading AST -- needed only for
// borrowConflictDiagnosticPostAssembly's narrow_to_shared_borrow repair text.
func placesByID(function core.Function) map[string]core.Place {
	if function.Linear == nil {
		return nil
	}
	byID := make(map[string]core.Place, len(function.Linear.Places))
	for _, place := range function.Linear.Places {
		byID[place.ID] = place
	}
	return byID
}

// moveWhileBorrowedDiagnosticPostAssembly builds ownership.move_while_borrowed
// at the post-assembly decision point, reproducing exactly the cause/repair
// shape analyzeStraightLine/analyzeArmBody used to build inline before
// D-09-08's authorized deletion moved this decision here.
func moveWhileBorrowedDiagnosticPostAssembly(move, borrow core.LinearOperation, lastUseSpan diagnostic.Span, spanByOperationID map[string]diagnostic.Span) diagnostic.Diagnostic {
	causes := []diagnostic.Cause{
		{Kind: "borrow_created_here", Span: spanPointer(spanByOperationID[borrow.ID])},
		{Kind: "borrow_used_later", Span: spanPointer(lastUseSpan)},
		{Kind: "loan", Detail: borrow.LoanID},
		{Kind: "owner", Detail: move.SourceID},
		{Kind: "type", Detail: move.TypeID},
	}
	primary := spanByOperationID[move.ID]
	return diagnostic.ErrorWithRepairs(
		"ownership.move_while_borrowed", primary, "cannot transfer ownership while a future-used shared loan is live", causes,
		diagnostic.Repair{Kind: "move_after_last_borrow_use"},
	)
}

// borrowConflictDiagnosticPostAssembly builds ownership.borrow_conflict at the
// post-assembly decision point, preserving the same cause shape and the same
// conditional narrow_to_shared_borrow repair (attached only when the NEW loan
// is exclusive) the deleted lowering-time gate used to build. The repair's
// Replacement text is reconstructed from core.Place.Name (a source-level fact
// core.Place already preserves), never from AST.
func borrowConflictDiagnosticPostAssembly(newBorrow, blockingBorrow core.LinearOperation, blockingLastUseSpan diagnostic.Span, places map[string]core.Place, spanByOperationID map[string]diagnostic.Span) diagnostic.Diagnostic {
	causes := []diagnostic.Cause{
		{Kind: "borrow_created_here", Span: spanPointer(spanByOperationID[blockingBorrow.ID])},
		{Kind: "borrow_used_later", Span: spanPointer(blockingLastUseSpan)},
		{Kind: "loan", Detail: blockingBorrow.LoanID},
		{Kind: "owner", Detail: newBorrow.SourceID},
		{Kind: "type", Detail: newBorrow.TypeID},
	}
	repairs := []diagnostic.Repair{{Kind: "create_loan_after_conflicting_loan_ends"}}
	if newBorrow.Kind == core.OpBorrowExclusive {
		// The repair's own Span must cover the WHOLE "let NAME = borrow mut
		// SOURCE" statement (Replacement rewrites all of it), never just the
		// RHS token spanByOperationID[newBorrow.ID] carries for the
		// diagnostic's own Primary/causes -- see the ":stmt"-suffixed entry
		// analyzeStraightLine/analyzeArmBody populate in the SAME channel.
		newBorrowSpan, ok := spanByOperationID[newBorrow.ID+":stmt"]
		if !ok {
			newBorrowSpan = spanByOperationID[newBorrow.ID]
		}
		targetName, sourceName := "", ""
		if places != nil {
			targetName = places[newBorrow.TargetID].Name
			sourceName = places[newBorrow.SourceID].Name
		}
		repairs = append(repairs, diagnostic.Repair{
			Kind: "narrow_to_shared_borrow", Span: &newBorrowSpan,
			Replacement:   "let " + targetName + " = borrow " + sourceName,
			Applicability: diagnostic.ApplicabilityMachineApplicable,
		})
	}
	return diagnostic.ErrorWithRepairs(
		"ownership.borrow_conflict", spanByOperationID[newBorrow.ID], "cannot create a loan while a conflicting loan is live", causes,
		repairs...,
	)
}

// interproceduralLoanLivenessDiagnostic builds the
// check.interprocedural_loan_liveness refusal ratified at this plan's Task 1
// checkpoint: exactly three causes in the fixed role order
// borrow_created_here / loan_extended_by_call / callee_return_contract
// (D-08-23), Primary is the offending OpMove's own span (D-08-22), and
// Repairs is nil (D-08-25) -- diagnostic.Error, never ErrorWithRepairs.
// causeDetail is cause 3's precomputed Detail string, chosen by the caller
// from a RECORDED fact (which direction fired), never re-derived here
// (08-02 Task 2): "<calleeID>:return.mode=<Mode>" for the forward direction
// (D-08-07), "<calleeID>:parameters[0].mode=<Mode>" for the backward
// direction (D-08-08).
func interproceduralLoanLivenessDiagnostic(move, borrow, call core.LinearOperation, causeDetail string, spanByOperationID map[string]diagnostic.Span) diagnostic.Diagnostic {
	borrowSpan := spanByOperationID[borrow.ID]
	callSpan := spanByOperationID[call.ID]
	causes := []diagnostic.Cause{
		{Kind: "borrow_created_here", Span: &borrowSpan},
		{Kind: "loan_extended_by_call", Span: &callSpan, Detail: call.CalleeID},
		{Kind: "callee_return_contract", Detail: causeDetail},
	}
	primary := spanByOperationID[move.ID]
	return diagnostic.Error(
		"check.interprocedural_loan_liveness", primary,
		"cannot transfer ownership while an interprocedurally-extended loan is still live", causes...,
	)
}

// checkCallArgumentTypeMismatch is 07-09's ratified check-side code for a
// call whose argument's type does not match the callee's declared
// parameter type, refused at emission time in resolveCallBinding. Its
// Primary span is the call site (binding.RHS.Span); its ordered Causes
// are callee / argument_type / declared_parameter_type. Constructed with
// diagnostic.Error -- no repairs: fixing it requires changing either the
// argument or the callee's own signature, and neither is a span-local
// replacement this checker can name (D-07-31c's standing reasoning).
// Distinct from core.CallArgumentTypeMismatch (corevalidate.go): the two
// codes name the SAME fact through two independent derivations and are
// deliberately never unified into one shared constant (see that const's
// own doc comment).
const checkCallArgumentTypeMismatch = "check.call_argument_type_mismatch"

// checkCallReturnTypeUnrepresentable is 07-09's ratified check-side code
// for the fail-closed half of deriving an OpCall's TargetID.TypeID from
// the callee's declared return type: refused when the callee's declared
// return type names no type fact available in the calling function, so no
// honest TargetID.TypeID can be derived. Its Primary span is the call
// site; its ordered Causes are callee / declared_return_type /
// available_type. diagnostic.Error, no repairs. Currently UNREACHABLE
// from any legal source program (sameType forces every function's return
// type to equal its parameter type) -- mutation-killed through
// callReturnTypeDerivationSeam, never through a .lang fixture; see
// PHASE-07-DEBT.md. Distinct from core.CallReturnTypeMismatch
// (corevalidate.go) for the same independence reason as
// checkCallArgumentTypeMismatch above.
const checkCallReturnTypeUnrepresentable = "check.call_return_type_unrepresentable"

// callArgumentTypeCheckSeam is 07-09 Task 3's D-07-41/D-07-42
// fault-injection seam for control:call.argument_type_matches_parameter:
// when true, resolveCallBinding's argument-type gate is skipped entirely,
// admitting a mismatched-argument call anyway. Unexported, false in
// production, set only from a same-package test that defers the restore
// immediately (see TestCallArgumentTypeCheckMutationKilled).
var callArgumentTypeCheckSeam = false

// callReturnTypeDerivationSeam is 07-09 Task 3's D-07-41/D-07-42
// fault-injection seam for control:call.target_type_from_callee_return:
// when true, the OpCall target's TypeID reverts to the caller's own
// argument place's TypeID -- the PRE-PLAN behavior, i.e. the exact defect
// 07-VERIFICATION.md's failed truth and 07-REVIEW.md's CR-01 named --
// instead of being derived from the callee's declared return contract.
// The seeded fault is literally the shipped bug, so the control's kill is
// a direct reproduction of the original finding. Unexported, false in
// production, set only from a same-package test that defers the restore
// immediately (see TestCallReturnTypeDerivationMutationKilled).
var callReturnTypeDerivationSeam = false

// callArgumentConsumeSeam is 07-11's D-07-41/D-07-42 fault-injection seam
// for control:call.argument_consumed_when_noncopyable: when true, the
// consume rule at the end of resolveCallBinding's declared-function arm is
// skipped entirely, so a call NEVER move-marks its argument regardless of
// its copy ability -- reproducing 07-VERIFICATION.md PVG-01 / 07-REVIEW.md
// CR-01's pre-plan hole, where the call boundary bypassed affine ownership
// entirely and the same non-copyable value could be passed to two separate
// calls and admitted by both check and corevalidate. Unexported, false in
// production, set only from a same-package test that defers the restore
// immediately (see TestCallArgumentConsumeMutationKilled).
var callArgumentConsumeSeam = false

// callArgumentConsumeAlwaysSeam is 07-11's D-07-41/D-07-42 fault-injection
// seam for control:call.copyable_argument_not_consumed: when true, a call
// argument is move-marked REGARDLESS of its copy ability -- the
// over-refusal fault for the non-refusing direction, so a COPYABLE
// argument's second use is wrongly refused with ownership.use_after_move.
// Over-refusal is a defect, not caution (07-11's own prohibition).
// Unexported, false in production, set only from a same-package test that
// defers the restore immediately (see
// TestCallArgumentConsumeOverRefusalMutationKilled).
var callArgumentConsumeAlwaysSeam = false

// SetCallArgumentConsumeSeamForTest is 07-11 Task 2 Test 3's cross-package
// independence seam. Go's build model excludes "_test.go" files from a
// normal package import, so a same-package-only unexported var (the shape
// callArgumentConsumeSeam itself uses for check's own same-package tests)
// cannot be reached by corevalidate's own independence test
// (TestCallConsumePeerIndependentOfCheck), which imports this package as
// an ordinary dependency to prove the peer still refuses
// call_argument_used_twice.lang's emitted core with the PRODUCER's own
// gate disabled. This mirrors corevalidate.go's own
// SetDisableCyclePeerForTest exactly (D-07-42): production-visible, but a
// documented test-only no-op unless a test explicitly calls it, always
// restored via the returned closure, and never called from any production
// code path in this repository.
func SetCallArgumentConsumeSeamForTest(disabled bool) (restore func()) {
	previous := callArgumentConsumeSeam
	callArgumentConsumeSeam = disabled
	return func() { callArgumentConsumeSeam = previous }
}

// calleeContract is 07-09's pre-body callee-contract entry (D-07-09,
// SEM-05): the callee's own declared parameter and return type
// constructor strings, keyed by function NAME in buildCalleeContracts
// below -- the same key functionIDs/functionNames already used, so callee
// resolution is unchanged. This is a SECOND, different table from
// callSignatureTable: callSignatureTable is core-derived and can only be
// built AFTER every function's body is checked (its Callable bit recomputes
// return origin from checked core.LinearOperation facts); calleeContract is
// AST-derived and therefore genuinely precedes every body admission -- a
// declared parameter/return type is a syntactic fact available before any
// body is walked. Like callSignatureTable, it carries no *core.LinearBody
// or *core.Match field, so SEM-05's body-blindness is preserved by
// construction on this path too. A missing entry's zero value has empty
// ParameterType/ReturnType strings, which the admission gate below treats
// as the REFUSING case (D-07-09): absence never admits.
type calleeContract struct {
	ID            string
	ParameterType string
	ReturnType    string
	// ReturnsBorrowOfParam is Task 3(b)'s widening (D-08-09): true iff the
	// callee's AST declares a borrowed return origin (function.ReturnOrigin
	// != nil) -- a genuinely PRE-BODY, syntactic fact, exactly like
	// ParameterType/ReturnType above. It is unused by the AST-shadow path
	// (computeLoanLastUses) this plan itself widens; it exists so a later
	// plan's fixture work and parity tests can name the declared fact on
	// both sides of the two admission paths without either path reading a
	// body.
	ReturnsBorrowOfParam bool
}

// buildCalleeContracts builds the pre-body callee-contract table from
// program's own AST, one entry per declared function, keyed by function
// name. See calleeContract's own doc comment for why this table is
// necessarily separate from callSignatureTable.
func buildCalleeContracts(program ast.Program) map[string]calleeContract {
	contracts := make(map[string]calleeContract, len(program.Funcs))
	for _, function := range program.Funcs {
		contracts[function.Name] = calleeContract{
			ID:                   semanticID(program.Module, "fn", function.Name),
			ParameterType:        function.Parameter.Type.Constructor,
			ReturnType:           function.ReturnType.Constructor,
			ReturnsBorrowOfParam: function.ReturnOrigin != nil,
		}
	}
	return contracts
}

// callSignatureTable is D-07-34's immutable signature table: a same-package,
// read-only view over each declared function's lang.interface/1
// FunctionSignature (ID, name, parameter contract, return contract, and the
// Callable bit -- D-07-31/D-07-32's publication-safety predicate, never
// export membership). It reuses core.FunctionSignature verbatim rather than
// minting a parallel type: FunctionSignature "deliberately has no
// Linear/Match field at all -- not merely an omitted one" (core.go's own
// doc comment on the type), so a caller admission path holding this table
// structurally cannot reach a callee body even by following a pointer
// (T-07-28). The zero value's lookup always misses -- Callable's Go zero
// value false extends to an absent table entry too (D-07-09) -- so a table
// that somehow failed to build refuses every call rather than admitting
// one, never the reverse. The only way to read it is lookup: there is no
// exported or unexported setter, so once buildCallSignatureTable returns,
// nothing in this package can mutate it (immutability by construction, not
// merely by convention).
type callSignatureTable struct {
	entries map[string]core.FunctionSignature
}

func (t callSignatureTable) lookup(calleeID string) (core.FunctionSignature, bool) {
	entry, ok := t.entries[calleeID]
	return entry, ok
}

// buildCallSignatureTable builds D-07-34's signature table from every
// function in program, using originvalidate.BuildInterface -- and, through
// it, originvalidate.PublishProblemsFor -- as Callable's SOLE authority
// (D-07-31/D-07-32, landed in 07-02). This is a deliberate reuse of the
// established single source of truth for what a "signature" is under
// lang.interface/1, not a second, competing derivation: the two genuinely
// independent derivations of the D-04-03 predicate this phase's threat
// register (T-07-31) requires are check's OWN admission-time consult of
// this table versus corevalidate's own, separately-implemented peer
// re-derivation (07-02's derivePeerSignature/peerCallable) -- never check
// versus originvalidate, which would just be the producer read twice.
// BuildInterface's failure modes are a json.Marshal error (which cannot
// occur for a program that reached this point with zero diagnostics -- no
// unsupported Go value ever enters core.Program) and, since 07-08
// (D-07-38), callgraph.Order refusing a cyclic graph before any
// ClosureDigest is computed. The returned error is never swallowed here:
// the call site skips verifyCallableRefusal entirely on error rather than
// admitting the empty-table fallback to run and mask the dedicated
// call-graph-cycle gate's own, more specific diagnostic (see that call
// site's comment).
func buildCallSignatureTable(program core.Program) (callSignatureTable, error) {
	if callSignatureTableBuildObserved != nil {
		callSignatureTableBuildObserved()
	}
	iface, err := originvalidate.BuildInterface(program)
	if err != nil {
		return callSignatureTable{}, err
	}
	entries := make(map[string]core.FunctionSignature, len(iface.Functions))
	for _, signature := range iface.Functions {
		entries[signature.ID] = signature
	}
	return callSignatureTable{entries: entries}, nil
}

// callSignatureTableBuildObserved is Task 1's D-07-34 ordering-instrumentation
// seam, the build-side counterpart to callSignatureTableLookupObserved: when
// non-nil, invoked once, synchronously, at the start of every
// buildCallSignatureTable call -- before it does anything else -- so a
// same-package test can record a single, real event order across a build and
// its lookups and assert the build strictly precedes every lookup. nil in
// production: zero cost, zero allocation.
var callSignatureTableBuildObserved func()

// verifyCallableRefusalSeam is D-07-41/D-07-42's fault-injection seam for
// Task 3's permitting-Callable control (QLT-08): when true, a callee whose
// table entry reports Callable == false (or whose entry is altogether
// absent) is treated as permitting the call anyway, instead of refusing it.
// Unexported, same-package-test-only, restored via defer in every test that
// engages it -- never an exported package-level mutable var on a
// production path (D-07-42).
var verifyCallableRefusalSeam = false

// callSignatureTableLookupObserved is Task 3's D-07-41 instrumentation
// seam: when non-nil, verifyCallableRefusal invokes it with every calleeID
// it looks up in table, so a same-package test can observe EXACTLY what
// this admission arm touched for a callee across a whole fixture and
// assert that set is exactly {signature table entry} -- never a body
// value. nil in production: zero cost, zero allocation, and the call site
// below is the ONLY place in this admission arm that ever names a callee's
// ID for a lookup.
var callSignatureTableLookupObserved func(calleeID string)

// verifyCallableRefusal is D-07-34/SEM-06's own "call" admission arm: for
// every core.OpCall operation in the whole assembled program, it looks the
// operation's own CalleeID up in table -- and reads NOTHING else about the
// callee -- refusing with core.CalleeNotCallable when the entry is absent
// (a synthetic gap; D-07-09's fail-closed default, since Callable's Go zero
// value is false) or its Callable bit is false. It never reads
// function.Linear or function.Match for the CALLEE (only for the CALLING
// function, to enumerate its own operations, which is the caller's own
// body -- never the callee's), so this arm structurally cannot reach a
// callee body even by mistake: table's entry type (core.FunctionSignature)
// carries neither field at all. The refusal carries no repairs
// (diagnostic.Error, never ErrorWithRepairs) per D-07-31c: exporting the
// callee cannot repair an unsafe borrow-derived return, so no
// export_callee repair is ever offered.
func verifyCallableRefusal(functions []core.Function, table callSignatureTable) *diagnostic.Diagnostic {
	// verifyCallableRefusalBodyReadSeam (Task 3, D-07-41/D-07-42) is the
	// body-blindness control's own kill: byID is built and consulted ONLY
	// when the seam is engaged (never in production), so a same-package
	// test can observe this arm reach directly into a callee's OWN
	// core.Function.Linear -- a body value -- instead of table, and prove
	// the body-blind control (TestCallAdmissionNeverReadsCalleeBody) goes
	// red exactly when this happens.
	var byID map[string]core.Function
	if verifyCallableRefusalBodyReadSeam {
		byID = make(map[string]core.Function, len(functions))
		for _, function := range functions {
			byID[function.ID] = function
		}
	}
	for _, function := range functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind != core.OpCall {
				continue
			}
			if callSignatureTableLookupObserved != nil {
				callSignatureTableLookupObserved(operation.CalleeID)
			}
			entry, ok := table.lookup(operation.CalleeID)
			callable := ok && entry.Callable
			if verifyCallableRefusalBodyReadSeam {
				// The fault itself: read the callee's OWN body directly
				// (Linear != nil) as a stand-in "callable" rule, instead of
				// consulting table at all. This is deliberately a
				// DIFFERENT determination than table's, so the divergence
				// is observable, not merely the act of touching Linear.
				if callee, exists := byID[operation.CalleeID]; exists {
					if calleeBodyReadObserved != nil {
						calleeBodyReadObserved(operation.CalleeID)
					}
					callable = callee.Linear != nil
				}
			}
			if verifyCallableRefusalSeam {
				callable = true
			}
			if !callable {
				diag := diagnostic.Error(
					core.CalleeNotCallable, function.Span, "call target is not callable",
					diagnostic.Cause{Kind: "callee", Detail: operation.CalleeID},
				)
				return &diag
			}
		}
	}
	return nil
}

// verifyCallableRefusalBodyReadSeam is Task 3's D-07-41/D-07-42
// fault-injection seam for the body-blindness control itself (QLT-08): when
// true, verifyCallableRefusal ALSO reads the callee's own core.Function
// directly (a body value, Linear) instead of consulting table alone --
// demonstrating exactly what a removed table-only discipline would let
// through. Unexported, same-package-test-only, restored via defer in every
// test that engages it -- never an exported package-level mutable var on a
// production path (D-07-42).
var verifyCallableRefusalBodyReadSeam = false

// calleeBodyReadObserved is Task 3's instrumentation seam: invoked ONLY by
// the fault path above (verifyCallableRefusalBodyReadSeam == true), never
// by the production admission arm, so a same-package test asserting zero
// invocations at the default (seam disengaged, across every
// testdata/phase07 fixture containing a call) proves the production
// admission arm structurally never reaches a callee body -- not merely
// that it "doesn't currently". nil in production: zero cost, zero
// allocation.
var calleeBodyReadObserved func(calleeID string)

// verifyCallInvariantsSeam is Task 2's D-07-41/D-07-42 fault-injection seam
// (QLT-08): when true, check's own independent re-derivation that a CalleeID
// resolves to a declared function is disabled. Unexported, same-package-
// test-only, restored via defer in every test that engages it -- never an
// exported package-level mutable var on a production path.
var verifyCallInvariantsSeam = false

// verifyCallInvariants independently re-derives D-07-29's three CalleeID
// invariants over the whole assembled program: an OpCall with an empty
// CalleeID; any non-OpCall operation with a non-empty CalleeID; and a
// CalleeID naming no declared function (D-07-45, gated by
// verifyCallInvariantsSeam for QLT-08's seeded-mutation control). This is
// check's OWN bookkeeping, derived from the function ID set it just built --
// it never consults corevalidate, mirroring corevalidate's own three checks
// derived independently on its own side (D-07-29's "derived independently at
// both sites" requirement).
func verifyCallInvariants(functions []core.Function) *diagnostic.Diagnostic {
	declared := make(map[string]bool, len(functions))
	for _, function := range functions {
		declared[function.ID] = true
	}
	for _, function := range functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind != core.OpCall {
				if operation.CalleeID != "" {
					diag := diagnostic.Error("core.callee_id_kind_exclusive", function.Span, "only an OpCall operation may carry a CalleeID")
					return &diag
				}
				continue
			}
			if operation.CalleeID == "" {
				diag := diagnostic.Error("core.callee_id_missing", function.Span, "an OpCall operation must carry a CalleeID")
				return &diag
			}
			if verifyCallInvariantsSeam {
				continue
			}
			if !declared[operation.CalleeID] {
				diag := diagnostic.Error(core.CallCalleeUnresolved, function.Span, "call target does not resolve to a declared function")
				return &diag
			}
		}
	}
	return nil
}

func sealedNames(types map[string]core.DataType) map[string]bool {
	names := make(map[string]bool, len(types))
	for name := range types {
		names[name] = true
	}
	return names
}

// maxBlocksPerFunction bounds T-03-01's CFG-shape denial-of-service surface
// at the lowering layer: entry block, one block per body arm, and the join
// block. Sized generously above anything this phase's fixtures need.
//
// Deliberately unreachable from real source today, exactly like
// ownership.borrow_requires_share (check_test.go's nonShareableTypeFact):
// the parser's own maxArmsPerMatch (64) caps arm count below this bound
// (2+64 = 66 < 128), so no real `.lang` program can ever trigger this arm
// of checkBranch through syntax.Parse. It exists as defense-in-depth for a
// future relaxation of the parser cap, and is exercised directly by a
// synthetic ast.Program in TestArmBodyLimits (check_branch_test.go), per
// D-10: an unreachable gate must be named as such, not assumed correct.
const maxBlocksPerFunction = 128

// checkBranch lowers a match function whose arms hold full linear bodies
// into a single core.Function that carries BOTH Match (arm identity and
// pattern/edge bookkeeping) and Linear (the flattened operations plus the
// Blocks/Edges the arms lower into). Per this phase's scope decision, a
// match that carries any arm body requires every arm to carry one — bare and
// body arms are never interleaved in the same function — so the native
// switch/case lowering never needs a bare-alternative fallback case inside a
// block-shaped function, and the interpreter/validator dispatch stays a
// simple "every arm has a BlockID" invariant rather than a per-arm union.
func checkBranch(module, functionID, matchID string, function ast.FuncDecl, dataType core.DataType, sealed map[string]bool, calleeContracts map[string]calleeContract, foreignSymbols map[string]foreignSymbolInfo) (core.Function, []diagnostic.Diagnostic, int, map[string]diagnostic.Span) {
	parameterType := coreType(function.Parameter.Type)
	derived, err := ability.DeriveSealed(parameterType, sealed)
	if err != nil {
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error("type.unknown", function.Parameter.Span, err.Error())}, typeNodeCount(parameterType), nil
	}
	typeID := functionID + ":type:0"
	parameterID := functionID + ":place:0"
	typeFact := core.TypeFact{ID: typeID, Shape: parameterType, Abilities: derived.Granted, NegativeWitnesses: derived.NegativeWitnesses}
	linear := &core.LinearBody{
		ID:         functionID + ":linear",
		Types:      []core.TypeFact{typeFact},
		Places:     []core.Place{{ID: parameterID, Name: function.Parameter.Name, TypeID: typeID}},
		Operations: []core.LinearOperation{},
	}

	work := typeNodeCount(parameterType)
	nextIndex := 0
	seen := make(map[string]bool)
	arms := make([]core.MatchArm, 0, len(function.Body.Arms))
	blocks := make([]core.Block, 0, len(function.Body.Arms)+2)
	edges := make([]core.Edge, 0, len(function.Body.Arms)*2)
	var diagnostics []diagnostic.Diagnostic
	entryBlockID := functionID + ":block:entry"
	joinBlockID := functionID + ":block:join"
	armBlockIDs := make([]string, 0, len(function.Body.Arms))
	armCFGBlocks := make([]cfgBlockSpec, 0, len(function.Body.Arms))
	armEdgeIDs := make(map[string]string, len(function.Body.Arms))
	callSpans := map[string]diagnostic.Span{}

	for index, arm := range function.Body.Arms {
		if seen[arm.Pattern] {
			diagnostics = append(diagnostics, diagnostic.Error("match.subsumed", arm.Span, "alternative is already matched"))
			continue
		}
		if !contains(dataType.Alternatives, arm.Pattern) {
			diagnostics = append(diagnostics, diagnostic.Error("match.unreachable", arm.Span, "pattern is not an alternative of the scrutinee type"))
			continue
		}
		if arm.Body == nil {
			diagnostics = append(diagnostics, diagnostic.Error(
				"core.mixed_arm_forms", arm.Span,
				"a match with any arm body requires every arm to carry a body this phase",
			))
			continue
		}
		seen[arm.Pattern] = true
		if len(blocks)+2 > maxBlocksPerFunction {
			diagnostics = append(diagnostics, diagnostic.Error("check.arm_body_limit", arm.Span, "function exceeds the declared block limit"))
			continue
		}

		armBlockID := fmt.Sprintf("%s:block:arm:%d", functionID, index)
		aliasOpID := fmt.Sprintf("%s:op:%d", functionID, nextIndex)
		aliasPointID := fmt.Sprintf("%s:point:linear:%d", functionID, nextIndex)
		aliasPlaceID := fmt.Sprintf("%s:place:%d", functionID, nextIndex+1)
		aliasOp := core.LinearOperation{
			ID: aliasOpID, PointID: aliasPointID, Kind: core.OpCopy, SourceID: parameterID, TargetID: aliasPlaceID, TypeID: typeID,
		}
		linear.Places = append(linear.Places, core.Place{ID: aliasPlaceID, Name: function.Parameter.Name, TypeID: typeID})
		linear.Operations = append(linear.Operations, aliasOp)
		armOperationIDs := []string{aliasOpID}
		armOps := []core.LinearOperation{aliasOp}
		nextIndex++
		work++

		support := analyzeArmBody(functionID, nextIndex, function.Parameter.Name, aliasPlaceID, arm.Body.Span, typeFact, arm.Body, calleeContracts, foreignSymbols)
		if support.Diagnostic != nil {
			diagnostics = append(diagnostics, *support.Diagnostic)
			continue
		}
		linear.Places = append(linear.Places, support.Places...)
		linear.Operations = append(linear.Operations, support.Operations...)
		for opID, span := range support.CallSpans {
			callSpans[opID] = span
		}
		for _, operation := range support.Operations {
			armOperationIDs = append(armOperationIDs, operation.ID)
		}
		armOps = append(armOps, support.Operations...)
		nextIndex += len(support.Operations)
		work += support.Work

		blocks = append(blocks, core.Block{
			ID: armBlockID, PointID: fmt.Sprintf("%s:point:arm:%d", functionID, index),
			OperationIDs: armOperationIDs, Successors: []string{joinBlockID},
		})
		armBlockIDs = append(armBlockIDs, armBlockID)
		armEdgeToJoinID := fmt.Sprintf("%s:edge:arm:%d:join", functionID, index)
		edges = append(edges,
			core.Edge{ID: fmt.Sprintf("%s:edge:entry:arm:%d", functionID, index), FromBlockID: entryBlockID, ToBlockID: armBlockID, Pattern: arm.Pattern},
			core.Edge{ID: armEdgeToJoinID, FromBlockID: armBlockID, ToBlockID: joinBlockID, Pattern: arm.Pattern},
		)
		armEdgeIDs[armBlockID+"->"+joinBlockID] = armEdgeToJoinID
		armCFGBlocks = append(armCFGBlocks, cfgBlockSpec{id: armBlockID, operations: armOps, successors: []string{joinBlockID}})
		arms = append(arms, core.MatchArm{
			ID: fmt.Sprintf("%s:arm:%d", functionID, index), EdgeID: fmt.Sprintf("%s:edge:%s", matchID, arm.Pattern),
			Pattern: arm.Pattern, BlockID: armBlockID,
		})
	}

	missing := make([]string, 0)
	for _, alternative := range dataType.Alternatives {
		if !seen[alternative] {
			missing = append(missing, alternative)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		causes := make([]diagnostic.Cause, 0, len(missing))
		for _, name := range missing {
			causes = append(causes, diagnostic.Cause{Kind: "missing_alternative", Detail: name})
		}
		diagnostics = append(diagnostics, diagnostic.Error("match.non_exhaustive", function.Body.Span, "match does not cover every alternative", causes...))
	}
	if len(diagnostics) > 0 {
		return core.Function{}, diagnostics, work, nil
	}

	blocks = append([]core.Block{{ID: entryBlockID, PointID: functionID + ":point:entry", OperationIDs: []string{}, Successors: armBlockIDs}}, blocks...)
	blocks = append(blocks, core.Block{ID: joinBlockID, PointID: functionID + ":point:return", OperationIDs: []string{}, Successors: []string{}})
	linear.Blocks = blocks
	linear.Edges = edges

	// Backward worklist loan-liveness dataflow (03-03, Q2): the join block
	// carries no operations, so this always converges to a per-arm-block
	// local answer under this phase's topology, but the fixpoint machinery
	// itself is general (see materializeLoanEndpoints's own multi-successor
	// test coverage in check_test.go).
	cfgBlocks := append(append([]cfgBlockSpec(nil), armCFGBlocks...), cfgBlockSpec{id: joinBlockID, successors: nil})
	fixpoint, diag := loanLivenessFixpoint(functionID, cfgBlocks, interproceduralSummaryTable{}, function.Body.Span)
	if diag != nil {
		diagnostics = append(diagnostics, *diag)
		return core.Function{}, diagnostics, work, nil
	}
	edgeIDLookup := func(fromBlockID, toBlockID string) string {
		if id, ok := armEdgeIDs[fromBlockID+"->"+toBlockID]; ok {
			return id
		}
		return fmt.Sprintf("%s:edge:%s:%s", functionID, fromBlockID, toBlockID)
	}
	linear.LoanEndpoints = materializeLoanEndpoints(functionID, cfgBlocks, edgeIDLookup, fixpoint, interproceduralSummaryTable{})
	work += fixpoint.work

	return core.Function{
		ID: functionID, Name: function.Name,
		EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
		Parameter:  core.Parameter{ID: parameterID, Name: function.Parameter.Name, Type: parameterType.Constructor},
		ReturnType: function.ReturnType.Constructor,
		Match:      &core.Match{ID: matchID, PointID: functionID + ":point:match", Scrutinee: function.Body.Scrutinee, Arms: arms},
		Linear:     linear,
		Span:       function.Span,
	}, nil, work, callSpans
}

// ---------------------------------------------------------------------
// 03-03: backward worklist loan liveness over the per-function CFG.
//
// This gives checkBranch's arm blocks their LoanEndpoint records (point vs
// edge) via a genuinely backward, block-local-transfer, worklist-to-a-
// fixpoint mechanism (Q2). Scope note (documented deviation, historical):
// the straight-line path's SERIALIZED core.LoanEndpoints stayed absent for
// Phase 1/2/3 -- 03-06's already-shipped TestPhase3FieldsAreOmittedWhenAbsent
// requires loan_endpoints stay absent from every Phase 1/2 program's
// serialized core, and every Phase 2 fixture is a straight-line body, so
// populating that FIELD there would have moved those already-verified
// bytes. That constraint is still honored (analyzeStraightLine never
// populates linear.LoanEndpoints), but D-05-35(d) has since made this same
// fixpoint (via computeLoanLastUses) the sole law deciding conflict/expiry
// in BOTH analyzeStraightLine and analyzeArmBody -- discoverLoanLastUses,
// the forward chain-inheritance scan that used to drive those decisions, is
// retired. This dataflow remains the sole producer of the observable
// core.LoanEndpoint records, still wired only into checkBranch's arm
// blocks.
// ---------------------------------------------------------------------

// cfgBlockSpec is the minimal per-block shape the backward worklist
// dataflow consumes: a stable ID, the block's own operations in program
// order, and its successor block IDs (edges out, unlabeled here -- the
// caller reattaches the real edge identity when materializing
// core.LoanEndpoint records).
type cfgBlockSpec struct {
	id         string
	operations []core.LinearOperation
	successors []string
}

// loanLivenessResult is the fixpoint's output: per-block live-in loan sets
// and the total counted work (one unit per transfer-function evaluation,
// one further unit per worklist reinsertion -- D-05/03-03-03).
type loanLivenessResult struct {
	liveIn map[string]map[string]bool
	work   int
}

// placeLoanChain is the O(1)-per-operation derivation this pass builds once
// per function, in a single forward pass over every block's operations
// concatenated in a stable order. A reborrow (`let review = borrow view`)
// is simultaneously a live view of its OWN new loan AND every loan its
// source place already carried -- the same transitive-liveness law
// discoverLoanLastUses' inheritance encodes, here expressed over place IDs
// instead of binding names. Rather than materializing that full ancestry as
// a list per place (an append-copy per operation that reintroduces exactly
// the Θ(N²) blowup this pass replaces, D-02-03/Q2(b)), it is kept as a
// linked chain: latestLoan[placeID] names only the FRESHEST loan a place is
// a view of, and parentLoan[loanID] names the loan immediately BEFORE it in
// the reborrow chain (absent for a loan with no reborrow ancestor). Walking
// the chain is deferred to blockLoanLiveness's own backward scan, which
// short-circuits the moment it reaches an already-recorded loan -- so the
// chain is walked in full at most ONCE per function (each loan visited and
// recorded exactly once across the whole block), bounding total per-block
// work to O(operations), not O(operations × chain depth).
type placeLoanChain struct {
	latestLoan map[string]string
	parentLoan map[string]string
	// borrowOperation and extendedByCall are Task 2's D-08-20..25 bookkeeping
	// for the interprocedural loan-liveness refusal: every other consumer of
	// placeLoanChain (materializeLoanEndpoints, computeLoanLastUses' shadow
	// path, blockLoanLiveness) ignores both fields, so their cost is confined
	// to checkInterproceduralLoanLiveness/interproceduralLoanLivenessDiagnostic.
	borrowOperation map[string]core.LinearOperation // loanID -> the OpBorrowShared/OpBorrowExclusive operation that created it (cause 1's span source)
	extendedByCall  map[string]core.LinearOperation // loanID -> the OpCall operation whose interprocedural summary propagated it past a call boundary (absent if the loan was never call-extended)
}

// derivePlaceLoans is D-08-07's FORWARD canonicalization pass: the
// interprocedural alias fact (an OpCall whose callee summary reports
// returnsBorrowOfParam) is established HERE, in this forward scan, never in
// blockLoanLiveness's backward transfer function (08-CONTEXT.md D-08-07;
// spike S-006 iteration 2 got this wrong and every accept-case test still
// passed). summaries' zero value (interproceduralSummaryTable{}) means "no
// interprocedural facts" -- every pre-Phase-08 call site passes it
// unchanged, reproducing today's behaviour exactly.
func derivePlaceLoans(operations []core.LinearOperation, summaries interproceduralSummaryTable) placeLoanChain {
	chain := placeLoanChain{
		latestLoan:      make(map[string]string, len(operations)),
		parentLoan:      make(map[string]string, len(operations)),
		borrowOperation: make(map[string]core.LinearOperation, len(operations)),
		extendedByCall:  make(map[string]core.LinearOperation, len(operations)),
	}
	for _, operation := range operations {
		inherited := chain.latestLoan[operation.SourceID]
		if operation.Kind == core.OpBorrowShared || operation.Kind == core.OpBorrowExclusive {
			chain.parentLoan[operation.LoanID] = inherited
			chain.borrowOperation[operation.LoanID] = operation
			if operation.TargetID != "" {
				chain.latestLoan[operation.TargetID] = operation.LoanID
			}
			continue
		}
		if operation.Kind == core.OpCall {
			if summary, ok := summaries.lookup(operation.CalleeID); ok && summary.returnsBorrowOfParam {
				chain.latestLoan[operation.TargetID] = chain.latestLoan[operation.SourceID]
				if inherited != "" {
					chain.extendedByCall[inherited] = operation
				}
				continue
			}
			// The callee is absent from the table or does not report
			// returnsBorrowOfParam (Mode == "owned"): give the call's result
			// a fresh identity by explicitly refusing to propagate. This
			// `continue` is load-bearing -- the generic inherit fallthrough
			// below already propagates `inherited` for an OpCall today, so
			// omitting this false-direction arm would leave that
			// over-approximation in place and the safe twin would never
			// admit (D-08-07 Task 2).
			delete(chain.latestLoan, operation.TargetID)
			continue
		}
		if inherited != "" && operation.TargetID != "" {
			chain.latestLoan[operation.TargetID] = inherited
		}
	}
	return chain
}

// loanBlockUse is one loan's last reference found within a single block by
// blockLoanLiveness's backward scan.
type loanBlockUse struct {
	loanID         string
	operationIndex int
	operationID    string
}

// blockLoanLiveness is the per-block transfer function: given a block's
// operations in program order, the function-global place->loan chain
// (derivePlaceLoans), and the set of loans already known live on exit
// (liveOut, the union of every successor's live-in set), it walks the
// operations BACKWARD. A reference to a place walks that place's loan chain
// from its freshest loan upward, marking each unrecorded ancestor live and
// recording the reference as its most-recent (first found, walking
// backward) use, stopping as soon as an already-recorded loan is reached
// (the amortized-linear short-circuit). Reaching a loan's own birth
// operation (OpBorrowShared/OpBorrowExclusive) removes it from the live
// set, since nothing earlier in program order can be "after" its creation.
// The returned live set is what must be live entering the block. The
// returned work count is one unit per operation inspected PLUS one unit per
// chain-ancestor step actually walked -- honest per-operation counting
// (D-05), not a flat per-block unit: a reintroduced unbounded chain walk
// (the Θ(N²) shape this pass replaces) would show up here as work growing
// faster than operation count, which TestReborrowChainWorkIsLinear asserts
// directly against.
//
// summaries is D-08-08 Task 2's own consuming clause: an OpCall enters the
// chain-ancestor walk below if and only if summaries.lookup(CalleeID)
// reports usesParam == true, OR the callee is absent from summaries (the
// refusing direction, never the permitting one) -- symmetric to how a
// plain reference is already treated unconditionally. Every OTHER
// operation kind keeps entering the walk exactly as before this task; this
// is a widening in one direction (a using callee now counts as a
// reference, where derivePlaceLoans' forward pass alone never established
// one) and a narrowing in the other (a non-using callee no longer
// over-approximates a use it never makes) -- both are required, since
// without the narrowing a callee that never touches its parameter would
// still extend the caller's loan and Pattern B's safe twin would never
// admit. summaries' zero value (interproceduralSummaryTable{}) always
// misses every lookup, so this gate is a no-op (unconditional entry, byte-
// identical to pre-Task-2 behaviour) for every pre-existing call site that
// still passes the zero-value table -- OWN-03's own intraprocedural
// admission law is untouched.
func blockLoanLiveness(operations []core.LinearOperation, chain placeLoanChain, liveOut map[string]bool, summaries interproceduralSummaryTable) ([]loanBlockUse, map[string]bool, int) {
	live := make(map[string]bool, len(liveOut))
	for loan := range liveOut {
		live[loan] = true
	}
	recorded := make(map[string]bool, len(live))
	var uses []loanBlockUse
	work := 0
	for index := len(operations) - 1; index >= 0; index-- {
		work++ // one unit per operation inspected
		operation := operations[index]
		entersWalk := true
		if operation.Kind == core.OpCall {
			summary, ok := summaries.lookup(operation.CalleeID)
			entersWalk = !ok || summary.usesParam
		}
		if entersWalk {
			for loan := chain.latestLoan[operation.SourceID]; loan != "" && !recorded[loan]; loan = chain.parentLoan[loan] {
				work++ // one unit per chain-ancestor step walked
				live[loan] = true
				recorded[loan] = true
				uses = append(uses, loanBlockUse{loanID: loan, operationIndex: index, operationID: operation.ID})
			}
		}
		if operation.Kind == core.OpBorrowShared || operation.Kind == core.OpBorrowExclusive {
			if !recorded[operation.LoanID] {
				recorded[operation.LoanID] = true
				uses = append(uses, loanBlockUse{loanID: operation.LoanID, operationIndex: index, operationID: operation.ID})
			}
			delete(live, operation.LoanID)
		}
	}
	return uses, live, work
}

func loanSetsEqual(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for loan := range a {
		if !b[loan] {
			return false
		}
	}
	return true
}

// cfgWalkFrame is the cycle pre-walk's own explicit-stack DFS frame
// (D-08-19a, porting callgraph.stackFrame's identical shape, D-07-18): the
// block's own ID and the index of the next successor still to visit.
// Pushed/popped entries replace the recursion frame a native-recursive walk
// would otherwise cost, so depth here is bounded by Go slice growth, never
// by the goroutine's own stack.
type cfgWalkFrame struct {
	id             string
	nextChildIndex int
}

// detectCFGCycle runs an iterative three-colour (white/gray/black)
// depth-first search over blocks' successors, rooted at every block in
// order (mirroring loanLivenessFixpoint's own former recursive walk, which
// also treated every block as a potential root) -- porting
// callgraph.Order's own explicit-stack structure (D-08-19a) rather than a
// self-calling closure: depth here is bounded by CFG block count within
// one function body, which is smaller than callgraph's roots, but "bounded
// by body size" is not the same guarantee a fixed native stack gives, and a
// long straight-line body is adversarially reachable (T-08-13). On a cycle
// it reports the SAME block identity the old recursive walk would have
// named in its error: the successor block found already on-stack (gray) at
// the moment of re-entry.
func detectCFGCycle(order []string, byID map[string]cfgBlockSpec) (blockID string, cyclic bool) {
	const (
		unvisited = 0
		inWork    = 1
		done      = 2
	)
	state := make(map[string]int, len(byID))
	for _, root := range order {
		if state[root] != unvisited {
			continue
		}
		stack := []cfgWalkFrame{{id: root}}
		state[root] = inWork
		for len(stack) > 0 {
			top := &stack[len(stack)-1]
			successors := byID[top.id].successors
			if top.nextChildIndex < len(successors) {
				child := successors[top.nextChildIndex]
				top.nextChildIndex++
				switch state[child] {
				case unvisited:
					state[child] = inWork
					stack = append(stack, cfgWalkFrame{id: child})
				case inWork:
					// A back edge to an on-stack (in-work) block is a real
					// cycle -- the exact re-entry point the old recursive
					// walk's own state[id]==inWork check would have named.
					return child, true
				case done:
					// A legitimately revisited already-finished block --
					// not a cycle.
				}
			} else {
				state[top.id] = done
				stack = stack[:len(stack)-1]
			}
		}
	}
	return "", false
}

// cfgBackEdgeDiagnostic builds the check.cfg_back_edge refusal (D-08-19b):
// Primary is span (the caller's own function-level span, since the CFG
// pre-walk operates over a whole function's block graph before any
// per-operation admission has anything narrower to point at), and the SOLE
// cause names the offending block's own ID -- never a bare formatted error
// string -- so an agent-side dispatch site can act on Causes[0].Detail
// directly (mirroring interproceduralLoanLivenessDiagnostic's own
// discipline). No repairs: a CFG cycle is a whole-topology defect no
// span-local edit resolves (matching core.call_graph_cycle's own
// non-repairable disposition for the analogous whole-program cycle,
// D-07-31c's reasoning restated at the intraprocedural CFG level).
func cfgBackEdgeDiagnostic(functionID, blockID string, span diagnostic.Span) diagnostic.Diagnostic {
	return diagnostic.Error(
		"check.cfg_back_edge", span,
		fmt.Sprintf("function %q's control-flow graph contains a cycle", functionID),
		diagnostic.Cause{Kind: "cycle_block", Detail: blockID},
	)
}

// loanLivenessBoundFactor is Task 2's derived iteration-bound multiplier
// (D-08-14). loanLivenessFixpoint's worklist loop is a monotone dataflow
// over a finite lattice of live-loan sets per block boundary, so it
// provably converges within lattice-height iterations -- the bound below is
// therefore an INTERNAL-CONSISTENCY ASSERTION, never a DoS defense: no
// `.lang` program can force divergence, only an implementation bug can (a
// non-monotone transfer function, a mutated lattice, or a lost `queued`
// flag reintroducing infinite reinsertion -- see PHASE-08-DEBT.md D-08-15).
// A flat magic constant was rejected: it is equally unprovokable from
// source (buying no additional safety over a derived bound), it carries no
// derivation trail an auditor can check, and it would eventually
// false-positive the day a legitimate program's block/loan count grows
// past whatever constant was picked. The bound is computed in ordinary Go
// `int` arithmetic as a product of small terms; at the sizes this language
// can express today (LANGUAGE-MATURITY.md: 58 programs, ~28 lines, one CFG
// block per match arm, arity 1) `blocks * loans` cannot approach
// math.MaxInt, so no overflow contract is specified and none is tested --
// this is a PLANNER ASSUMPTION (08-04-PLAN.md OWN-06/precision) that must be
// re-opened if a later milestone makes block or loan counts
// program-controlled (iteration, loops, collections) before this bound is
// trusted again.
const loanLivenessBoundFactor = 4

// loanLivenessBound derives the fixpoint's fail-closed iteration ceiling
// from the CFG's own shape (D-08-14): loanLivenessBoundFactor times the
// function's own block count times (its distinct loan count + 1). The "+1"
// is load-bearing, not decorative: a function with zero loans still costs
// one transfer-function evaluation per block (blockLoanLiveness inspects
// every operation regardless of whether any loan is involved), so a bound
// of exactly `factor * blocks * 0` would refuse every loan-free multi-block
// function on its very first non-trivial iteration -- the silent
// under-approximation criterion 2 forbids, inflicted on the MOST common
// case. With the "+1" floor this formula reduces, for the single-block
// reborrow-chain shape TestLivenessWorkScale/TestReborrowChainWorkIsLinear
// already pin at `4*n+4`, to exactly that same bound -- not a coincidence:
// both were independently derived from the same per-operation accounting.
// It scales with the input by construction (doubling blockCount doubles
// the bound, holding loans fixed), unlike a flat constant.
func loanLivenessBound(blockCount, distinctLoanCount int) int {
	return loanLivenessBoundFactor * blockCount * (distinctLoanCount + 1)
}

// loanLivenessBoundSeam is Task 2's own D-08-16/QLT-08 fault-injection seam:
// when true, loanLivenessFixpoint's worklist loop trips the named bound
// refusal on its very first iteration regardless of the actually-computed
// bound, proving the control has been SEEN TO FAIL, not merely to exist
// (TestLoanLivenessBoundMutationKilled). Unexported, false in production,
// set only from a same-package test that defers the restore immediately --
// the same shape callReturnTypeDerivationSeam above already uses. It must
// never become the exported pathoracle.TerminatorKindsOverride shape:
// D-07-42 forbids multiplying exported mutable globals across production
// check/corevalidate paths on -race and test-order grounds, and
// loanLivenessFixpoint's own []cfgBlockSpec parameter is itself unexported,
// so any test exercising this seam is same-package by construction.
var loanLivenessBoundSeam = false

// loanLivenessBoundExceededDiagnostic builds the
// check.loan_liveness_bound_exceeded refusal (Task 2, D-08-14/D-08-18):
// Primary is the function's OWN declaration span, since exceeding the
// bound is a whole-function property with no single offending operation to
// point at (unlike cfgBackEdgeDiagnostic above, which can at least name a
// block). Causes name the function itself and flag that the derived bound
// is block-count-scaled -- but NEVER the bound's own computed integer
// value, in either a Cause or the Message (D-08-18): retuning
// loanLivenessBoundFactor must never move this diagnostic's own published
// ID for two runs that both happen to exceed their (possibly different)
// bound -- TestLoanLivenessBoundValueIsNotInDiagnosticIdentity asserts this
// directly by varying block count and confirming identical diagnostic IDs.
// Repairs is nil: this is not one of lang-repair's five defect classes
// (matching Phase 07's non-repairable disposition for a whole-program
// call-graph cycle, the analogous whole-topology defect).
func loanLivenessBoundExceededDiagnostic(functionID string, span diagnostic.Span) diagnostic.Diagnostic {
	return diagnostic.Error(
		"check.loan_liveness_bound_exceeded", span,
		"loan-liveness fixpoint exceeded its derived, block-count-scaled iteration bound",
		diagnostic.Cause{Kind: "function", Detail: functionID},
		diagnostic.Cause{Kind: "block_count"},
	)
}

// loanLivenessFixpoint computes backward monotone dataflow over the finite
// lattice of live loan IDs per block boundary (Q2), iterated with a
// worklist to a fixpoint. blocks must be given in a stable order; every
// successor ID must resolve to a block in the same slice. A cycle (a block
// reachable from itself by following successors) is rejected fail-closed --
// OWN-03 is scoped to acyclic CFGs this phase (T-03-11) -- via the coded
// check.cfg_back_edge diagnostic (D-08-19b). The worklist loop itself is
// bounded by loanLivenessBound (D-08-14): reaching it returns the named
// check.loan_liveness_bound_exceeded refusal and a ZERO-VALUED
// loanLivenessResult, never a truncated live-in map -- span is the caller's
// own function-level span, attached to whichever of the two refusals fires.
func loanLivenessFixpoint(functionID string, blocks []cfgBlockSpec, summaries interproceduralSummaryTable, span diagnostic.Span) (loanLivenessResult, *diagnostic.Diagnostic) {
	byID := make(map[string]cfgBlockSpec, len(blocks))
	order := make([]string, 0, len(blocks))
	var allOps []core.LinearOperation
	for _, block := range blocks {
		byID[block.id] = block
		order = append(order, block.id)
		allOps = append(allOps, block.operations...)
	}
	placeLoan := derivePlaceLoans(allOps, summaries)

	if cycleBlockID, cyclic := detectCFGCycle(order, byID); cyclic {
		diag := cfgBackEdgeDiagnostic(functionID, cycleBlockID, span)
		return loanLivenessResult{}, &diag
	}

	predecessors := make(map[string][]string, len(blocks))
	for _, block := range blocks {
		for _, successor := range block.successors {
			predecessors[successor] = append(predecessors[successor], block.id)
		}
	}

	liveIn := make(map[string]map[string]bool, len(blocks))
	for _, id := range order {
		liveIn[id] = map[string]bool{}
	}

	queue := append([]string(nil), order...)
	queued := make(map[string]bool, len(blocks))
	for _, id := range order {
		queued[id] = true
	}

	distinctLoanCount := len(placeLoan.borrowOperation)
	bound := loanLivenessBound(len(blocks), distinctLoanCount)

	work := 0
	for len(queue) > 0 {
		if loanLivenessBoundSeam || work >= bound {
			diag := loanLivenessBoundExceededDiagnostic(functionID, span)
			return loanLivenessResult{}, &diag
		}
		id := queue[0]
		queue = queue[1:]
		queued[id] = false
		work++ // one transfer-function evaluation

		block := byID[id]
		liveOut := map[string]bool{}
		for _, successor := range block.successors {
			for loan := range liveIn[successor] {
				liveOut[loan] = true
			}
		}
		_, newLiveIn, transferWork := blockLoanLiveness(block.operations, placeLoan, liveOut, summaries)
		work += transferWork
		if !loanSetsEqual(liveIn[id], newLiveIn) {
			liveIn[id] = newLiveIn
			for _, predecessor := range predecessors[id] {
				if !queued[predecessor] {
					queue = append(queue, predecessor)
					queued[predecessor] = true
					work++ // one worklist reinsertion
				}
			}
		}
	}
	return loanLivenessResult{liveIn: liveIn, work: work}, nil
}

// materializeLoanEndpoints turns the fixpoint's converged live-in sets into
// core.LoanEndpoint records. Two, mutually exclusive, kinds are produced per
// loan per block:
//
//   - A POINT endpoint where the loan is referenced inside a block and does
//     NOT survive to that block's own live-out (liveOut, the union of every
//     successor's live-in) -- its last reference is genuinely inside this
//     block, so it ends at a point.
//   - An EDGE endpoint only at a genuine successor DIVERGENCE: a block with
//     more than one successor, where the loan is needed by at least one
//     successor (present in liveOut) but NOT by a specific other successor
//     (absent from that successor's own live-in). That is exactly the edge
//     the loan ends on -- the loan is still tracked along the successor(s)
//     that need it (and will receive its own point/edge endpoint further
//     downstream, wherever it is finally consumed), while the diverging
//     edge is where a program on THAT path may safely mutate/move the owner
//     (T-03-06's literal claim, and 03-03-02's fixture-pair falsifier).
//
// A single-successor block therefore never produces an edge endpoint for a
// loan flowing through it unremarked -- there is no divergence to record --
// matching checkBranch's current topology (every arm block has exactly one
// successor, the join), where genuine edge endpoints require a loan born
// before a real fan-out, which today's arm-body lowering does not yet
// produce from real source (D-10; proven possible in the general case by
// TestEdgeSpecificLiveOut).
func materializeLoanEndpoints(functionID string, blocks []cfgBlockSpec, edgeID func(fromBlockID, toBlockID string) string, result loanLivenessResult, summaries interproceduralSummaryTable) []core.LoanEndpoint {
	var placeLoanAll []core.LinearOperation
	for _, block := range blocks {
		placeLoanAll = append(placeLoanAll, block.operations...)
	}
	placeLoan := derivePlaceLoans(placeLoanAll, summaries)

	var endpoints []core.LoanEndpoint
	for _, block := range blocks {
		liveOut := map[string]bool{}
		for _, successor := range block.successors {
			for loan := range result.liveIn[successor] {
				liveOut[loan] = true
			}
		}

		if len(block.successors) > 1 {
			for loan := range liveOut {
				for _, successor := range block.successors {
					if result.liveIn[successor][loan] {
						continue
					}
					endpoints = append(endpoints, core.LoanEndpoint{
						ID:     edgeID(block.id, successor) + ":" + loan,
						LoanID: loan, Kind: "edge", EdgeID: edgeID(block.id, successor),
					})
				}
			}
		}

		uses, _, _ := blockLoanLiveness(block.operations, placeLoan, liveOut, summaries)
		for _, use := range uses {
			if liveOut[use.loanID] {
				continue // survives past this block; its endpoint lives elsewhere
			}
			endpoints = append(endpoints, core.LoanEndpoint{
				ID:     fmt.Sprintf("%s:point:%s:%d:%s", functionID, block.id, use.operationIndex, use.loanID),
				LoanID: use.loanID, Kind: "point", BlockID: block.id, AfterOperationID: use.operationID,
			})
		}
	}
	sort.Slice(endpoints, func(i, j int) bool { return endpoints[i].ID < endpoints[j].ID })
	return endpoints
}

// analyzeArmBody analyzes one match arm's linear body using the same
// straight-line ownership machinery as analyzeStraightLine, but numbers
// every place/operation/point/loan id starting at startIndex within the
// function's single flat Operations list rather than restarting at zero.
// This keeps every arm's operations, once concatenated in arm order,
// satisfying the exact same global ordinal invariant corevalidate already
// enforces for a non-branching linear body (place N is produced by
// operation N-1). The scrutinee is not re-declared here: parameterPlaceID
// names the place the caller already seeded (the implicit per-arm alias
// copy), so every arm's bindings are checked against an alias that only that
// arm can move, and one arm's move can never be observed as a false
// use-after-move by a sibling arm that never runs at the same time (mutually
// exclusive control flow — the two arms' places never collide because their
// IDs are distinct global ordinals). Deliberately a separate function from
// analyzeStraightLine, duplicating rather than reusing it, so the Phase 1/2
// straight-line path (checkLinear) carries zero risk from this addition.
//
// D-09-08/D-09-09: loan-liveness admission (conflict/expiry) is no longer
// decided here. lowering emits core.OpMove/OpBorrowShared/OpBorrowExclusive
// unconditionally and makes no timing decision; checkInterproceduralLoanLiveness's
// post-assembly pass, extended to the purely-intraprocedural case, is now the
// SOLE decision point for `ownership.move_while_borrowed` /
// `ownership.borrow_conflict` (D-09-07: one algorithm -- loanLivenessFixpoint/
// materializeLoanEndpoints -- two callers, never two laws). The summary-blind
// early call site (computeLoanLastUses, its shadow:place/shadow:loan
// scaffolding, and its own loanLivenessFixpoint("shadow", ...) call) that used
// to feed this function's own admission decision was DELETED under D-09-08's
// authorization at plan 09-08's mid-phase gate.
//
// LoanFinalUses/States below stay populated -- both are test-visible evidence
// surfaces with NO production reader (grep-verified), computed UP FRONT by
// loanFinalUseEvidence (this file) exactly as before, so they still answer
// for the FULL declared body even when a later per-binding fact fails and
// the walk below never reaches every binding -- an independent oracle
// populates the same fields the same way, over the same full body,
// regardless of which fact fails first.
func analyzeArmBody(functionID string, startIndex int, parameterName, parameterPlaceID string, parameterSpan diagnostic.Span, typeFact core.TypeFact, body *ast.LinearBody, calleeContracts map[string]calleeContract, foreignSymbols map[string]foreignSymbolInfo) ownershipSupport {
	result := ownershipSupport{
		Places: []core.Place{}, Operations: []core.LinearOperation{}, LoanFinalUses: []loanFinalUseFact{}, States: []ownershipStateFact{},
		Work: len(body.Bindings) + 1,
	}
	loanUses, discoveryWork := loanFinalUseEvidence(parameterName, body)
	result.Work += discoveryWork
	for index, binding := range body.Bindings {
		if binding.RHS.Kind == "borrow" || binding.RHS.Kind == "borrow_mut" {
			result.LoanFinalUses = append(result.LoanFinalUses, loanFinalUseFact{
				LoanID: fmt.Sprintf("%s:loan:%d", functionID, startIndex+index), Binding: binding.Name, OperationIndex: startIndex + loanUses[index].index,
			})
		}
	}
	places := map[string]*placeState{
		parameterName: {place: core.Place{ID: parameterPlaceID, Name: parameterName, TypeID: typeFact.ID}, declared: parameterSpan, initialized: true},
	}
	activeLoans := make(map[string]map[string]int) // ownerID -> loanID -> lastUse (evidence only, D-09-08/D-09-09)
	expiringLoans := make(map[int][]struct {
		ownerID string
		loanID  string
	})
	endLoans := func(index int) {
		for _, loan := range expiringLoans[index] {
			if loans := activeLoans[loan.ownerID]; loans != nil {
				delete(loans, loan.loanID)
				if len(loans) == 0 {
					delete(activeLoans, loan.ownerID)
				}
			}
		}
	}
	fail := func(problem diagnostic.Diagnostic) ownershipSupport {
		result.DiagnosticCode = problem.Code
		result.Diagnostic = &problem
		return result
	}
	useAfterMove := func(span diagnostic.Span, state *placeState) diagnostic.Diagnostic {
		causes := []diagnostic.Cause{
			{Kind: "declared_here", Span: spanPointer(state.declared)},
			{Kind: "moved_here", Span: state.movedAt},
			{Kind: "place", Detail: state.place.ID},
			{Kind: "transfer_target", Detail: state.moveTargetID},
			{Kind: "type", Detail: state.place.TypeID},
		}
		return diagnostic.ErrorWithRepairs(
			"ownership.use_after_move", span, "value was used after ownership transferred", causes,
			diagnostic.Repair{Kind: "use_transfer_target", Detail: state.moveTargetID},
			diagnostic.Repair{Kind: "move_use_before_transfer"},
		)
	}
	for index, binding := range body.Bindings {
		result.Work++
		global := startIndex + index
		if binding.RHS.Kind == "call" {
			op, target, diag := resolveCallBinding(functionID, global, binding, places, calleeContracts, typeFact, foreignSymbols)
			if diag != nil {
				return fail(*diag)
			}
			result.Places = append(result.Places, target)
			places[binding.Name] = &placeState{place: target, declared: binding.Span, initialized: true}
			result.Operations = append(result.Operations, op)
			if result.CallSpans == nil {
				result.CallSpans = map[string]diagnostic.Span{}
			}
			result.CallSpans[op.ID] = binding.RHS.Span
			endLoans(index)
			result.States = append(result.States, ownershipSnapshot(index, places, activeLoans))
			continue
		}
		source, ok := places[binding.RHS.Source]
		if !ok {
			return fail(diagnostic.Error("name.unknown", binding.RHS.Span, "binding source is unknown"))
		}
		if !source.initialized {
			return fail(useAfterMove(binding.RHS.Span, source))
		}
		target := core.Place{ID: fmt.Sprintf("%s:place:%d", functionID, global+1), Name: binding.Name, TypeID: source.place.TypeID}
		kind := core.OpCopy
		loanIDStr := ""
		var newLoanOwnerID string
		var newLoanLastUse int
		switch binding.RHS.Kind {
		case "take":
			// D-09-08/D-09-09: no activeLoans conflict check here -- lowering
			// emits core.OpMove unconditionally, and the post-assembly pass
			// decides ownership.move_while_borrowed (D-09-07).
			kind = core.OpMove
			source.initialized = false
			source.movedAt = spanPointer(binding.RHS.Span)
			source.moveTargetID = target.ID
		case "borrow":
			if !hasTypeAbility(typeFact, core.AbilityShare) {
				causes := []diagnostic.Cause{
					{Kind: "declared_here", Span: spanPointer(source.declared)},
					{Kind: "missing_ability", Detail: missingAbilityDetail(typeFact, core.AbilityShare)},
					{Kind: "place", Detail: source.place.ID},
					{Kind: "type", Detail: source.place.TypeID},
				}
				return fail(diagnostic.ErrorWithRepairs(
					"ownership.borrow_requires_share", binding.RHS.Span, "type does not grant the share ability required to borrow", causes,
					diagnostic.Repair{Kind: "use_take_instead"},
				))
			}
			// D-09-08/D-09-09: no conflictingLoan check here -- lowering
			// emits core.OpBorrowShared unconditionally, and the post-assembly
			// pass decides ownership.borrow_conflict (D-09-07). activeLoans
			// below is EVIDENCE bookkeeping only (feeds States.ActiveLoans).
			kind = core.OpBorrowShared
			loanIDStr = fmt.Sprintf("%s:loan:%d", functionID, global)
			newLoanOwnerID, newLoanLastUse = source.place.ID, loanUses[index].index
		case "borrow_mut":
			// An exclusive loan is gated on the same share-ability requirement
			// as a shared loan: the ability that permits observation without
			// duplicating ownership is exactly what an exclusive loan also
			// needs. No source-reachable Phase 1/2/3 type withholds share
			// (see nonShareableTypeFact, check_test.go), so this gate — like
			// the identical one immediately above for shared borrows — is
			// exercised only synthetically today (D-10).
			if !hasTypeAbility(typeFact, core.AbilityShare) {
				causes := []diagnostic.Cause{
					{Kind: "declared_here", Span: spanPointer(source.declared)},
					{Kind: "missing_ability", Detail: missingAbilityDetail(typeFact, core.AbilityShare)},
					{Kind: "place", Detail: source.place.ID},
					{Kind: "type", Detail: source.place.TypeID},
				}
				return fail(diagnostic.ErrorWithRepairs(
					"ownership.borrow_requires_share", binding.RHS.Span, "type does not grant the share ability required to borrow", causes,
					diagnostic.Repair{Kind: "use_take_instead"},
				))
			}
			kind = core.OpBorrowExclusive
			loanIDStr = fmt.Sprintf("%s:loan:%d", functionID, global)
			newLoanOwnerID, newLoanLastUse = source.place.ID, loanUses[index].index
		default:
			if !hasTypeAbility(typeFact, core.AbilityCopy) {
				causes := []diagnostic.Cause{
					{Kind: "declared_here", Span: spanPointer(source.declared)},
					{Kind: "missing_ability", Detail: missingAbilityDetail(typeFact, core.AbilityCopy)},
					{Kind: "place", Detail: source.place.ID},
					{Kind: "type", Detail: source.place.TypeID},
				}
				return fail(diagnostic.ErrorWithRepairs(
					"ownership.transfer_requires_take", binding.RHS.Span, "noncopyable binding requires explicit take", causes,
					diagnostic.Repair{Kind: "insert_take"},
				))
			}
		}
		if loanIDStr != "" {
			if activeLoans[newLoanOwnerID] == nil {
				activeLoans[newLoanOwnerID] = make(map[string]int)
			}
			activeLoans[newLoanOwnerID][loanIDStr] = newLoanLastUse
			expiringLoans[newLoanLastUse] = append(expiringLoans[newLoanLastUse], struct {
				ownerID string
				loanID  string
			}{ownerID: newLoanOwnerID, loanID: loanIDStr})
		}
		result.Places = append(result.Places, target)
		places[binding.Name] = &placeState{place: target, declared: binding.Span, initialized: true}
		operation := core.LinearOperation{
			ID: fmt.Sprintf("%s:op:%d", functionID, global), PointID: fmt.Sprintf("%s:point:linear:%d", functionID, global),
			Kind: kind, SourceID: source.place.ID, TargetID: target.ID, LoanID: loanIDStr, TypeID: source.place.TypeID,
		}
		result.Operations = append(result.Operations, operation)
		// D-08-22: widen the SAME program-wide span channel D-07-35 built for
		// call spans to also carry every take/borrow/borrow_mut operation's
		// own span, so Task 2's interprocedural loan-liveness diagnostic can
		// project its Primary (the offending take) and cause 1
		// (borrow_created_here) spans from spanByOperationID exactly as
		// checkCallGraphAcyclic already does for call spans -- never a
		// second span-carrying channel.
		if result.CallSpans == nil {
			result.CallSpans = map[string]diagnostic.Span{}
		}
		result.CallSpans[operation.ID] = binding.RHS.Span
		// D-09-08/D-09-09: also record the WHOLE binding statement's span
		// under a synthesized ":stmt" suffix key, in the SAME channel
		// (merged into spanByOperationID exactly like every other CallSpans
		// entry) -- needed ONLY by borrowConflictDiagnosticPostAssembly's
		// narrow_to_shared_borrow repair, whose Replacement text rewrites
		// the ENTIRE "let NAME = borrow mut SOURCE" statement, never just
		// the RHS token binding.RHS.Span already carries.
		result.CallSpans[operation.ID+":stmt"] = binding.Span
		endLoans(index)
		result.States = append(result.States, ownershipSnapshot(index, places, activeLoans))
	}
	result.Work++
	global := startIndex + len(body.Bindings)
	if body.DefectReason != "" {
		// D-04-15: a defect terminator needs no bound value at all -- it
		// carries the parameter's own alias place purely so every operation
		// still resolves a real, initialized SourceID (this package's
		// uniform "every operation reads an initialized place" invariant).
		// No release runs on this path (D-04-18); this arm shape can never
		// carry a foreign acquisition this phase, so there is nothing to
		// leak, but the shape stays consistent with a future acquiring arm.
		parameter := places[parameterName]
		result.Operations = append(result.Operations, core.LinearOperation{
			ID: fmt.Sprintf("%s:op:%d", functionID, global), PointID: fmt.Sprintf("%s:point:linear:%d", functionID, global),
			Kind: core.OpDefect, SourceID: parameter.place.ID, TypeID: parameter.place.TypeID, Reason: body.DefectReason,
		})
		result.Places = append(result.Places, core.Place{
			ID: fmt.Sprintf("%s:place:%d", functionID, global+1), Name: "_", TypeID: parameter.place.TypeID,
		})
		endLoans(len(body.Bindings))
		result.States = append(result.States, ownershipSnapshot(len(body.Bindings), places, activeLoans))
		return result
	}
	returned, ok := places[body.Result]
	if !ok {
		return fail(diagnostic.Error("name.unknown", body.Span, "linear result is unknown"))
	}
	if !returned.initialized {
		return fail(useAfterMove(body.Span, returned))
	}
	result.Operations = append(result.Operations, core.LinearOperation{
		ID: fmt.Sprintf("%s:op:%d", functionID, global), PointID: fmt.Sprintf("%s:point:linear:%d", functionID, global),
		Kind: core.OpReturn, SourceID: returned.place.ID, TypeID: returned.place.TypeID,
	})
	// A straight-line body's Return is always the last operation, so its
	// missing target never leaves a gap in the flat Places array (nothing
	// after it needs a place index). A branch's Return is NOT the last
	// operation in the whole function's flat list — the next arm's own
	// operations follow it — so corevalidate's place-order invariant
	// (place N is produced by operation N-1) would otherwise desync at
	// exactly this point. Padding with one unreferenced place per arm's
	// Return keeps every operation, including Return, consuming exactly
	// one place-index slot, restoring the same array-position coupling
	// analyzeStraightLine's callers already rely on.
	result.Places = append(result.Places, core.Place{
		ID: fmt.Sprintf("%s:place:%d", functionID, global+1), Name: "_", TypeID: returned.place.TypeID,
	})
	endLoans(len(body.Bindings))
	result.States = append(result.States, ownershipSnapshot(len(body.Bindings), places, activeLoans))
	return result
}

func checkLinear(module, functionID string, function ast.FuncDecl, calleeContracts map[string]calleeContract, foreignSymbols map[string]foreignSymbolInfo) (core.Function, []diagnostic.Diagnostic, int, map[string]diagnostic.Span) {
	parameterType := coreType(function.Parameter.Type)
	if !sameType(function.ReturnType, function.Parameter.Type) {
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error("type.return_mismatch", function.Span, "linear result must have the parameter type")}, typeNodeCount(parameterType), nil
	}
	// OWN-04: the borrowed-view return case relaxes nothing about type
	// identity above (the underlying type must still match the parameter
	// type) — it adds a new legal declaration alongside the unchanged
	// ordinary case, resolved on the syntactic discriminant
	// (function.ReturnOrigin != nil) rather than on the return type's shape.
	// A borrowed return with a path other than the function's own single
	// parameter is a causal, span-bearing rejection: this reduced language
	// has one parameter per function and no field-path-bearing executable
	// shape, so any other path can never be honest.
	if function.ReturnOrigin != nil && function.ReturnOrigin.Path != function.Parameter.Name {
		causes := []diagnostic.Cause{
			{Kind: "declared_path", Detail: function.ReturnOrigin.Path},
			{Kind: "only_legal_path", Detail: function.Parameter.Name},
		}
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error(
			"origin.unknown_path", function.ReturnOrigin.Span, "borrowed-view origin path must name the function's own parameter", causes...,
		)}, typeNodeCount(parameterType), nil
	}
	derived, err := ability.Derive(parameterType)
	if err != nil {
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error("type.unknown", function.Parameter.Span, err.Error())}, typeNodeCount(parameterType), nil
	}
	typeID := functionID + ":type:0"
	if !executableShape(parameterType) {
		// Ability derivation above already ran and produced facts for this
		// shape (derived.Granted/derived.NegativeWitnesses) — this gate
		// closes D-02-09 by refusing EXECUTION admission, not by
		// withholding ability derivation, so the shape stays available for
		// ability evidence (see TestAbilityFactsSurviveExecutionRejection).
		// Only Byte and Buffer have a native lowering this phase
		// (cgen.linearInput); every other shape (Box, Pair, or a nominal
		// data type used in a straight-line body) type-checks and derives
		// abilities cleanly but cannot be run by any engine, so it must be
		// refused here with a causal span rather than dying spanless
		// downstream on every engine (D-07).
		causes := []diagnostic.Cause{
			{Kind: "type", Detail: typeID},
			{Kind: "constructor", Detail: parameterType.Constructor},
		}
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.ErrorWithRepairs(
			"check.unexecutable_shape", function.Parameter.Span,
			"parameter shape has no native execution lowering this phase", causes,
			diagnostic.Repair{Kind: "use_executable_shape", Detail: "Byte or Buffer"},
		)}, typeNodeCount(parameterType), nil
	}
	parameterID := functionID + ":place:0"
	linear := &core.LinearBody{
		ID:         functionID + ":linear",
		Types:      []core.TypeFact{{ID: typeID, Shape: parameterType, Abilities: derived.Granted, NegativeWitnesses: derived.NegativeWitnesses}},
		Places:     []core.Place{{ID: parameterID, Name: function.Parameter.Name, TypeID: typeID}},
		Operations: []core.LinearOperation{},
	}
	support := analyzeStraightLine(functionID, function.Parameter.Name, function.Parameter.Span, linear.Types[0], function.Body.Linear, calleeContracts, foreignSymbols)
	if support.Diagnostic != nil {
		return core.Function{}, []diagnostic.Diagnostic{*support.Diagnostic}, support.Work, nil
	}
	linear.Places = support.Places
	linear.Operations = support.Operations
	var publicOrigin *core.PublicOrigin
	if function.ReturnOrigin != nil {
		publicOrigin = &core.PublicOrigin{Paths: []string{function.ReturnOrigin.Path}, Access: function.ReturnOrigin.Access}
	}
	return core.Function{
		ID: functionID, Name: function.Name, EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
		Parameter: core.Parameter{ID: parameterID, Name: function.Parameter.Name, Type: parameterType.Constructor}, ReturnType: function.ReturnType.Constructor,
		Linear: linear, PublicOrigin: publicOrigin, Span: function.Span,
	}, nil, support.Work + support.FixpointWork, support.CallSpans
}

// ---------------------------------------------------------------------
// Phase 4: fallible foreign call admission (D-04-01/D-04-02/D-04-04/D-04-05).
//
// This section is deliberately a separate, narrow path rather than a
// generalization of checkLinear/analyzeStraightLine: it handles exactly the
// one shape this plan's tracer proves -- a straight-line function whose sole
// binding is `try <foreign symbol>(<parameter>)`, immediately returned. A
// richer shape (ordinary bindings before or after the call, multiple calls)
// is out of scope this plan and is refused with a named diagnostic rather
// than silently mishandled. Keeping this fully separate from checkLinear
// means the Phase 1-3 straight-line path carries zero risk from this
// addition (D-04-23's byte-identity requirement), exactly as checkBranch
// stayed separate from analyzeStraightLine in Phase 3.
// ---------------------------------------------------------------------

// maxForeignBlocksPerProgram, maxForeignSymbolsPerBlock, and
// maxForeignParametersPerSymbol bound T-04-05's foreign declaration surface,
// derived from the existing declaration/function caps (syntax.go's
// maxDeclarations family) rather than an arbitrary round number: a program
// already cannot declare more than a few thousand top-level items, so a
// foreign surface bounded well below that is fail-closed, not merely
// advisory.
const (
	maxForeignBlocksPerProgram     = 64
	maxForeignSymbolsPerBlockCheck = 64
	maxForeignParametersPerSymbol  = 1
)

// foreignSymbolInfo is check.go's own resolved view of one declared foreign
// symbol: the policy keys collectForeignSymbols found, by name, so
// checkFallibleLinear (and, in a later task, the admission gate refusing a
// missing unwind/nonlocal_exit policy) can ask "was this key present"
// without re-scanning ast.ForeignPolicy each time.
type foreignSymbolInfo struct {
	Name         string
	Parameter    ast.Parameter
	ReturnType   ast.TypeRef
	Allocator    string
	HasAllocator bool
	Unwind       string
	HasUnwind    bool
	NonlocalExit string
	HasNonlocal  bool
	Fails        string
	HasFails     bool
	// Alias is Phase 4 plan 06's new optional policy key (D-04-28): "borrow"
	// or "retain" when the symbol's ok-edge return is declared to alias its
	// own argument, "" (the default) meaning fully owned. Unlike
	// Unwind/NonlocalExit this key has no admission gate requiring its
	// presence -- a symbol that genuinely returns an owned value simply
	// never declares it.
	Alias string
	Span  diagnostic.Span
}

// validForeignPolicyValue is check.go's OWN, deliberate third implementation
// of the C-identifier predicate `^[A-Za-z_][A-Za-z0-9_]*$` already carried by
// corevalidate.validCIdentifier and cgen.validForeignSymbol (04-13,
// 04-VERIFICATION.md gap 2b, FFI-01): this phase's standing independence
// posture requires every layer to be able to refuse a hostile value without
// depending on any other layer having run first, so this is not a shared
// helper. cgen.EmitForeignHeader splices every foreign policy value raw into
// a C comment at cgen.go:1332-1335 (allocator/unwind/nonlocal_exit) and the
// same unsanitized-splice pattern continues for the remaining contract
// fields (cgen.go:1336-1348); a value carrying a comment terminator (`*/`)
// escapes that comment and exposes live, uncommented top-level C source to
// the real C compiler session.go hands the generated unit to
// (native.Runner.CompileConformanceUnit). Iterates BYTES, not runes, so a
// multibyte rune is rejected by its individual bytes rather than accepted as
// one "character".
func validForeignPolicyValue(value string) bool {
	if len(value) == 0 {
		return false
	}
	first := value[0]
	if !(first >= 'A' && first <= 'Z' || first >= 'a' && first <= 'z' || first == '_') {
		return false
	}
	for index := 1; index < len(value); index++ {
		b := value[index]
		if !(b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z' || b >= '0' && b <= '9' || b == '_') {
			return false
		}
	}
	return true
}

// collectForeignSymbols builds the module-wide foreign symbol table from
// every declared `foreign C {}` block, applying T-04-05's caps fail-closed.
// It does not refuse a symbol for a missing unwind/nonlocal_exit policy --
// that admission gate is checkFallibleLinear's job (D-04-16), fired only for
// a symbol an actual call resolves to, so a declared-but-never-called
// under-specified symbol does not block an unrelated program. A hostile
// POLICY VALUE is different: it is never legitimate for any symbol, called
// or not, so refusing it at declaration time (04-13, below) is the earliest
// honest point and is not in tension with that deferral, which is about an
// ABSENT key, not a hostile one.
func collectForeignSymbols(program ast.Program) (map[string]foreignSymbolInfo, []diagnostic.Diagnostic) {
	if len(program.Foreign) > maxForeignBlocksPerProgram {
		return nil, []diagnostic.Diagnostic{diagnostic.Error("check.foreign_block_limit", program.Foreign[0].Span, "program exceeds the declared foreign block limit")}
	}
	symbols := make(map[string]foreignSymbolInfo)
	var policyDiagnostics []diagnostic.Diagnostic
	for _, block := range program.Foreign {
		if len(block.Symbols) > maxForeignSymbolsPerBlockCheck {
			return nil, []diagnostic.Diagnostic{diagnostic.Error("check.foreign_symbol_limit", block.Span, "foreign block exceeds the declared symbol limit")}
		}
		for _, symbol := range block.Symbols {
			if len(symbol.Policies) > maxForeignPoliciesPerSymbolCheck {
				return nil, []diagnostic.Diagnostic{diagnostic.Error("check.foreign_policy_limit", symbol.Span, "foreign symbol exceeds the declared policy limit")}
			}
			info := foreignSymbolInfo{Name: symbol.Name, Parameter: symbol.Parameter, ReturnType: symbol.ReturnType, Span: symbol.Span}
			for _, policy := range symbol.Policies {
				// 04-13 (04-VERIFICATION.md gap 2b, FFI-01): refuse a
				// hostile policy value BEFORE the key-specific switch below,
				// for EVERY key -- an unrecognised key's value is still
				// author-controlled text with no reason to admit a hostile
				// value for it. The message is a fixed literal and never
				// includes policy.Value: echoing an attacker-controlled
				// string containing newlines or quotes into diagnostic JSON
				// is the same class of defect this check exists to close;
				// policy.Span already locates the offending policy.
				if !validForeignPolicyValue(policy.Value) {
					policyDiagnostics = append(policyDiagnostics, diagnostic.Error(
						"check.foreign_policy_value_unsafe", policy.Span,
						"foreign policy value must be a C identifier",
					))
					continue
				}
				switch policy.Key {
				case "unwind":
					info.Unwind, info.HasUnwind = policy.Value, true
				case "nonlocal_exit":
					info.NonlocalExit, info.HasNonlocal = policy.Value, true
				case "allocator":
					info.Allocator, info.HasAllocator = policy.Value, true
				case "fails":
					info.Fails, info.HasFails = policy.Value, true
				case "alias":
					info.Alias = policy.Value
				}
			}
			symbols[symbol.Name] = info
		}
	}
	if len(policyDiagnostics) > 0 {
		return nil, policyDiagnostics
	}
	return symbols, nil
}

// maxForeignPoliciesPerSymbolCheck mirrors the parser's own
// maxForeignPoliciesPerSymbol cap (syntax package): declared again here,
// independently, rather than imported, so check.go's own admission surface
// is bounded even if a future caller constructs an ast.Program directly
// (bypassing the parser).
const maxForeignPoliciesPerSymbolCheck = 32

// hasTryCall reports whether a linear body's bindings contain a fallible
// foreign call consumer (D-04-06's `try` or `discard ... because` forms). A
// function with no such binding takes the entirely unchanged checkLinear
// path.
func hasTryCall(body *ast.LinearBody) bool {
	for _, binding := range body.Bindings {
		if binding.RHS.Kind == "try_call" || binding.RHS.Kind == "discard_call" {
			return true
		}
	}
	return false
}

// checkFallibleLinear dispatches between the two shapes this phase supports.
// The tracer shape (04-01) is a single `try` binding immediately returned.
// The resource-lifecycle shape (04-02, D-04-07) is a sequence of one or more
// try/discard foreign-call bindings whose result is the function's own
// parameter -- every acquired resource is released before return, so the
// parameter (never moved by a foreign call) is always what comes back.
func checkFallibleLinear(functionID string, function ast.FuncDecl, foreignSymbols map[string]foreignSymbolInfo, functionNames map[string]bool, dataTypes map[string]core.DataType) (core.Function, []diagnostic.Diagnostic, int) {
	parameterType := coreType(function.Parameter.Type)
	if !sameType(function.ReturnType, function.Parameter.Type) {
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error("type.return_mismatch", function.Span, "linear result must have the parameter type")}, typeNodeCount(parameterType)
	}
	derived, err := ability.Derive(parameterType)
	if err != nil {
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error("type.unknown", function.Parameter.Span, err.Error())}, typeNodeCount(parameterType)
	}
	typeID := functionID + ":type:0"
	work := typeNodeCount(parameterType) + 1
	if !executableShape(parameterType) {
		causes := []diagnostic.Cause{{Kind: "type", Detail: typeID}, {Kind: "constructor", Detail: parameterType.Constructor}}
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.ErrorWithRepairs(
			"check.unexecutable_shape", function.Parameter.Span,
			"parameter shape has no native execution lowering this phase", causes,
			diagnostic.Repair{Kind: "use_executable_shape", Detail: "Byte or Buffer"},
		)}, work
	}

	body := function.Body.Linear
	if len(body.Bindings) == 1 && body.Bindings[0].RHS.Kind == "try_call" && body.Result == body.Bindings[0].Name {
		return checkForeignTracer(functionID, function, parameterType, derived, typeID, work, body.Bindings[0], foreignSymbols, functionNames, dataTypes)
	}
	if len(body.Bindings) > 0 && body.Result == function.Parameter.Name && everyBindingIsFallible(body.Bindings) {
		return checkResourceLifecycle(functionID, function, parameterType, derived, typeID, work, foreignSymbols, functionNames, dataTypes)
	}
	return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error(
		"check.foreign_call_shape_unsupported", body.Span,
		"this phase supports only a single fallible foreign call immediately returned, or a sequence of try/discard foreign calls whose result is the function's own parameter",
	)}, work
}

// resolveCallBinding implements D-07-01's call admission predicate, shared
// verbatim by analyzeStraightLine and analyzeArmBody (Phase 07, D-07-40):
// arity is fixed at 1 this phase (D-07-07 defers arity-N to a later phase);
// the argument must resolve in the caller's own places scope map -- the
// function's own parameter or any prior let -- exactly like every other
// binding kind's Source lookup; and the callee must resolve to exactly one
// of three outcomes: a declared Lang function (admitted, CalleeID set to
// that function's own ID per D-07-29), a declared foreign symbol (refused
// with the pre-Phase-07 syntax.fallible_call_not_consumed code -- D-07-40
// relocates this refusal's enforcement layer from parse time to check time
// without changing its published code), or neither (refused with
// core.CallCalleeUnresolved, D-07-45 -- never silently dropped, since a
// dropped edge is how a cycle escapes detection).
func resolveCallBinding(functionID string, opOrdinal int, binding ast.Binding, places map[string]*placeState, calleeContracts map[string]calleeContract, typeFact core.TypeFact, foreignSymbols map[string]foreignSymbolInfo) (core.LinearOperation, core.Place, *diagnostic.Diagnostic) {
	if len(binding.RHS.Arguments) != 1 {
		causes := []diagnostic.Cause{
			{Kind: "declared_arity", Detail: fmt.Sprintf("%d", len(binding.RHS.Arguments))},
			{Kind: "supported_arity", Detail: "1"},
		}
		diag := diagnostic.Error("check.call_arity_unsupported", binding.RHS.Span, "this phase supports only arity-1 calls", causes...)
		return core.LinearOperation{}, core.Place{}, &diag
	}
	argumentName := binding.RHS.Arguments[0]
	argument, ok := places[argumentName]
	if !ok {
		diag := diagnostic.Error("name.unknown", binding.RHS.Span, "call argument is unknown")
		return core.LinearOperation{}, core.Place{}, &diag
	}
	if !argument.initialized {
		causes := []diagnostic.Cause{
			{Kind: "moved_here", Span: argument.movedAt},
			{Kind: "place", Detail: argument.place.ID},
			{Kind: "transfer_target", Detail: argument.moveTargetID},
		}
		diag := diagnostic.Error("ownership.use_after_move", binding.RHS.Span, "value was used after ownership transferred", causes...)
		return core.LinearOperation{}, core.Place{}, &diag
	}
	if contract, isFunction := calleeContracts[binding.RHS.Callee]; isFunction {
		// 07-09 D-07-09/SEM-05: the argument-type gate. Exact string
		// equality on the constructor -- never TypeID equality (a TypeID
		// is per-function and can never match across two functions, so
		// that comparison would refuse every legal call) and never a
		// coercion or subtyping relation (none exists). An absent
		// contract entry or an empty ParameterType refuses: absence is
		// the refusing case (D-07-09's fail-closed default), which is
		// exactly what the zero-value calleeContract's empty string
		// achieves against typeFact.Shape.Constructor (never empty for
		// an executable shape).
		if !callArgumentTypeCheckSeam && (typeFact.Shape.Constructor == "" || contract.ParameterType == "" || typeFact.Shape.Constructor != contract.ParameterType) {
			causes := []diagnostic.Cause{
				{Kind: "callee", Detail: contract.ID},
				{Kind: "argument_type", Detail: typeFact.Shape.Constructor},
				{Kind: "declared_parameter_type", Detail: contract.ParameterType},
			}
			diag := diagnostic.Error(checkCallArgumentTypeMismatch, binding.RHS.Span, "call argument type does not match the callee's declared parameter type", causes...)
			return core.LinearOperation{}, core.Place{}, &diag
		}
		// 07-09 T-07-09-02: the target place's type is derived from the
		// CALLEE'S OWN declared return contract, resolved against the
		// CALLER's own type facts -- never copied from the caller's
		// argument place. This function has exactly one type fact
		// (typeFact), per D-07-09's language-surface constraint (every
		// function's declared return type equals its declared parameter
		// type, sameType-enforced at the head of checkLinear/checkBranch/
		// checkFallibleLinear), so "resolved against the caller's own
		// type facts" means: the callee's declared return type must equal
		// this function's single type fact's own constructor. When it
		// does not, no honest TypeID can be derived and the call is
		// refused fail-closed -- never fabricated, never a fallback to
		// the argument's TypeID.
		var derivedTypeID string
		if callReturnTypeDerivationSeam {
			derivedTypeID = argument.place.TypeID
		} else {
			if contract.ReturnType == "" || contract.ReturnType != typeFact.Shape.Constructor {
				causes := []diagnostic.Cause{
					{Kind: "callee", Detail: contract.ID},
					{Kind: "declared_return_type", Detail: contract.ReturnType},
					{Kind: "available_type", Detail: typeFact.Shape.Constructor},
				}
				diag := diagnostic.Error(checkCallReturnTypeUnrepresentable, binding.RHS.Span, "callee's declared return type names no type fact available in the calling function", causes...)
				return core.LinearOperation{}, core.Place{}, &diag
			}
			derivedTypeID = typeFact.ID
		}
		target := core.Place{ID: fmt.Sprintf("%s:place:%d", functionID, opOrdinal+1), Name: binding.Name, TypeID: derivedTypeID}
		op := core.LinearOperation{
			ID: fmt.Sprintf("%s:op:%d", functionID, opOrdinal), PointID: fmt.Sprintf("%s:point:linear:%d", functionID, opOrdinal),
			Kind: core.OpCall, SourceID: argument.place.ID, TargetID: target.ID, TypeID: derivedTypeID, CalleeID: contract.ID,
		}
		// 07-11 (PVG-01/CR-01, T-07-11-01): a call transfers its argument.
		// A copyable argument (core.AbilityCopy) is COPIED -- the caller's
		// binding stays live, mirroring the binding switch's default arm.
		// A non-copyable argument is MOVED -- all four placeState fields
		// are set, mirroring the "take" arm verbatim, so the EXISTING
		// ownership.use_after_move gate at the top of this function fires
		// on any later use, including a second call. An argument whose
		// type fact carries no AbilityCopy witness at all (empty ability
		// list) is treated as non-copyable: fail-closed, never the
		// permitting default. This must stay the ONLY place this rule is
		// expressed -- see check.go's own doc comment on
		// resolveCallBinding and the two call arms' unchanged continues.
		if !callArgumentConsumeSeam && (callArgumentConsumeAlwaysSeam || !hasTypeAbility(typeFact, core.AbilityCopy)) {
			argument.initialized = false
			argument.movedAt = spanPointer(binding.RHS.Span)
			argument.moveTargetID = target.ID
			argument.moveTargetName = binding.Name
		}
		return op, target, nil
	}
	if _, isForeign := foreignSymbols[binding.RHS.Callee]; isForeign {
		diag := diagnostic.Error("syntax.fallible_call_not_consumed", binding.RHS.Span, "a fallible call must be the operand of `try`")
		return core.LinearOperation{}, core.Place{}, &diag
	}
	if verifyCallInvariantsSeam {
		// D-07-41/D-07-42 seam (QLT-08, Task 2 Test 6): skip the
		// resolves-to-a-declared-function predicate, admitting the call
		// anyway with a synthetic (non-ID) CalleeID -- demonstrates what a
		// removed check would let through. Shares verifyCallInvariants'
		// seam variable, both in package check, so this and the
		// post-build pass fail together under the same seeded mutation;
		// corevalidate carries the identical seam, independently, on its
		// own side (disableCalleeResolutionCheckForTest). 07-09: the
		// duplicated operation here derives its TypeID from typeFact.ID
		// -- the SAME source production now uses -- rather than
		// argument.place.TypeID, so this seam continues to differ from
		// production in exactly the ONE predicate it was written to
		// disable (callee resolution) and not incidentally in the type
		// derivation too. No calleeContract exists for this branch (the
		// callee name is, by construction, unresolved), so there is
		// nothing to derive a callee return type from here.
		target := core.Place{ID: fmt.Sprintf("%s:place:%d", functionID, opOrdinal+1), Name: binding.Name, TypeID: typeFact.ID}
		op := core.LinearOperation{
			ID: fmt.Sprintf("%s:op:%d", functionID, opOrdinal), PointID: fmt.Sprintf("%s:point:linear:%d", functionID, opOrdinal),
			Kind: core.OpCall, SourceID: argument.place.ID, TargetID: target.ID, TypeID: typeFact.ID, CalleeID: binding.RHS.Callee,
		}
		// Mirrors the declared-function arm's 07-11 consume rule above, so
		// this seam continues to differ from production in exactly the ONE
		// predicate it was written to disable (callee resolution) and not
		// incidentally in consumption too.
		if !callArgumentConsumeSeam && (callArgumentConsumeAlwaysSeam || !hasTypeAbility(typeFact, core.AbilityCopy)) {
			argument.initialized = false
			argument.movedAt = spanPointer(binding.RHS.Span)
			argument.moveTargetID = target.ID
			argument.moveTargetName = binding.Name
		}
		return op, target, nil
	}
	diag := diagnostic.Error(core.CallCalleeUnresolved, binding.RHS.Span, "call target does not resolve to a declared function")
	return core.LinearOperation{}, core.Place{}, &diag
}

func everyBindingIsFallible(bindings []ast.Binding) bool {
	for _, binding := range bindings {
		if binding.RHS.Kind != "try_call" && binding.RHS.Kind != "discard_call" {
			return false
		}
	}
	return true
}

// resolveForeignStep independently resolves one step's declared foreign
// symbol and its failure-ADT type fact, shared by both checkForeignTracer and
// checkResourceLifecycle so the two admission gates (D-04-02/D-04-16) and the
// argument-shape rule stay in exact lockstep between the tracer's single-call
// shape and the resource-lifecycle chain.
func resolveForeignStep(binding ast.Binding, functionParameterName string, foreignSymbols map[string]foreignSymbolInfo, functionNames map[string]bool, dataTypes map[string]core.DataType) (foreignSymbolInfo, core.TypeRef, ability.Result, *diagnostic.Diagnostic) {
	if len(binding.RHS.Arguments) != maxForeignParametersPerSymbol || binding.RHS.Arguments[0] != functionParameterName {
		problem := diagnostic.Error("name.unknown", binding.RHS.Span, "foreign call argument must be the function's own parameter")
		return foreignSymbolInfo{}, core.TypeRef{}, ability.Result{}, &problem
	}
	symbol, isForeign := foreignSymbols[binding.RHS.Callee]
	if !isForeign {
		if functionNames[binding.RHS.Callee] {
			causes := []diagnostic.Cause{{Kind: "callee", Detail: binding.RHS.Callee}}
			problem := diagnostic.ErrorWithRepairs(
				"core.call_target_not_foreign", binding.RHS.Span,
				"a fallible call's target must be a declared foreign symbol, not a Lang function", causes,
				diagnostic.Repair{Kind: "declare_foreign_symbol"},
			)
			return foreignSymbolInfo{}, core.TypeRef{}, ability.Result{}, &problem
		}
		problem := diagnostic.Error("name.unknown", binding.RHS.Span, "foreign call target is unknown")
		return foreignSymbolInfo{}, core.TypeRef{}, ability.Result{}, &problem
	}
	if problem := missingForeignPolicyDiagnostic(symbol); problem != nil {
		return foreignSymbolInfo{}, core.TypeRef{}, ability.Result{}, problem
	}
	errDataType, ok := dataTypes[symbol.Fails]
	if !symbol.HasFails || !ok || len(errDataType.Alternatives) == 0 {
		problem := diagnostic.Error("type.unknown", binding.Span, "foreign symbol's declared failure type is unknown or has no alternatives")
		return foreignSymbolInfo{}, core.TypeRef{}, ability.Result{}, &problem
	}
	errShape := core.TypeRef{Constructor: errDataType.Name}
	errDerived, derr := ability.DeriveSealed(errShape, map[string]bool{errDataType.Name: true})
	if derr != nil {
		problem := diagnostic.Error("type.unknown", binding.Span, derr.Error())
		return foreignSymbolInfo{}, core.TypeRef{}, ability.Result{}, &problem
	}
	return symbol, errShape, errDerived, nil
}

// checkForeignTracer lowers the 04-01 shape: a straight-line function whose
// sole binding is a fallible foreign call, immediately returned on the ok
// edge. It produces a core.Function whose Linear body carries three blocks
// (entry/ok/err) and two edges, exactly mirroring checkBranch's block/edge
// shape but keyed on a fallible call rather than a match arm.
func checkForeignTracer(functionID string, function ast.FuncDecl, parameterType core.TypeRef, derived ability.Result, typeID string, work int, tryBinding ast.Binding, foreignSymbols map[string]foreignSymbolInfo, functionNames map[string]bool, dataTypes map[string]core.DataType) (core.Function, []diagnostic.Diagnostic, int) {
	symbol, errShape, errDerived, problem := resolveForeignStep(tryBinding, function.Parameter.Name, foreignSymbols, functionNames, dataTypes)
	if problem != nil {
		return core.Function{}, []diagnostic.Diagnostic{*problem}, work
	}

	parameterID := functionID + ":place:0"
	okPlaceID := functionID + ":place:1"
	errPlaceID := functionID + ":place:2"
	errTypeID := functionID + ":type:1"
	entryBlockID := functionID + ":block:entry"
	okBlockID := functionID + ":block:ok"
	errBlockID := functionID + ":block:err"
	okEdgeID := functionID + ":edge:entry:ok"
	errEdgeID := functionID + ":edge:entry:err"
	callOpID := functionID + ":op:0"
	returnOpID := functionID + ":op:1"
	failOpID := functionID + ":op:2"

	linear := &core.LinearBody{
		ID: functionID + ":linear",
		Types: []core.TypeFact{
			{ID: typeID, Shape: parameterType, Abilities: derived.Granted, NegativeWitnesses: derived.NegativeWitnesses},
			{ID: errTypeID, Shape: errShape, Abilities: errDerived.Granted, NegativeWitnesses: errDerived.NegativeWitnesses},
		},
		Places: []core.Place{
			{ID: parameterID, Name: function.Parameter.Name, TypeID: typeID},
			{ID: okPlaceID, Name: tryBinding.Name, TypeID: typeID},
			{ID: errPlaceID, Name: "_err", TypeID: errTypeID},
		},
		Operations: []core.LinearOperation{
			{
				ID: callOpID, PointID: functionID + ":point:linear:0", Kind: core.OpForeignCall,
				SourceID: parameterID, TargetID: okPlaceID, TypeID: typeID,
				OkEdgeID: okEdgeID, ErrEdgeID: errEdgeID, ErrTargetID: errPlaceID,
			},
			{ID: returnOpID, PointID: functionID + ":point:linear:1", Kind: core.OpReturn, SourceID: okPlaceID, TypeID: typeID},
			{ID: failOpID, PointID: functionID + ":point:linear:2", Kind: core.OpFail, SourceID: errPlaceID, TypeID: errTypeID},
		},
		Blocks: []core.Block{
			{ID: entryBlockID, PointID: functionID + ":point:entry", OperationIDs: []string{callOpID}, Successors: []string{okBlockID, errBlockID}},
			{ID: okBlockID, PointID: functionID + ":point:ok", OperationIDs: []string{returnOpID}},
			{ID: errBlockID, PointID: functionID + ":point:err", OperationIDs: []string{failOpID}},
		},
		Edges: []core.Edge{
			{ID: okEdgeID, FromBlockID: entryBlockID, ToBlockID: okBlockID, Pattern: "ok"},
			{ID: errEdgeID, FromBlockID: entryBlockID, ToBlockID: errBlockID, Pattern: "err"},
		},
	}

	linear.Operations[0].Allocator = symbol.Allocator
	contract := buildForeignContract(symbol)

	return core.Function{
		ID: functionID, Name: function.Name,
		EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
		Parameter:       core.Parameter{ID: parameterID, Name: function.Parameter.Name, Type: parameterType.Constructor},
		ReturnType:      function.ReturnType.Constructor,
		Linear:          linear,
		ForeignContract: contract,
		Span:            function.Span,
	}, nil, work
}

// standardForeignLayout is this phase's fixed target-layout obligation
// (D-04-12/T-04-14): every foreign symbol this phase declares acquires
// exactly one malloc-backed record shaped like the frozen
// native/lang_foreign_resource_private.h's own lang_foreign_resource_block
// -- a single one-byte payload field -- so every declared foreign symbol
// shares an identical Layout. This is not an omission-tolerant default -- it
// never varies, and there is today nothing for a symbol to declare
// differently -- but a future phase with richer foreign record shapes must
// replace this with a real per-symbol declaration. The conformance unit
// (cgen.EmitForeignConformance) proves this declared belief against the real
// private header via paired sizeof/_Alignof/offsetof assertions.
func standardForeignLayout() *core.RecordLayout {
	return &core.RecordLayout{
		Size: 1, Alignment: 1, ForeignTypeName: "lang_foreign_resource_block",
		Fields: []core.LayoutField{
			{Name: "payload", Size: 1, Alignment: 1, Offset: 0, CType: "unsigned char"},
		},
	}
}

// standardForeignObligations returns this phase's fixed values for the
// InitializedState/Capture/Retention/Aliasing obligation categories
// (D-04-12): "fully" because this phase's language has no partial-field
// initialization shape at all; "none" for the other three because the
// language has no closures, no threads, and no calls into Lang for a
// foreign symbol to capture, retain, or alias anything through. These are
// declared facts, not proven ones -- cgen's sidecar manifest emitter records
// capture/retention/aliasing in unchecked_obligations precisely because
// nothing in this phase exercises them (D-04-12c).
func standardForeignObligations() (initializedState, capture, retention, aliasing string) {
	return "fully", "none", "none", "none"
}

// buildForeignContract assembles the complete core.ForeignContract for one
// resolved foreign symbol, sharing the standard obligations/layout every
// declared symbol carries this phase.
func buildForeignContract(symbol foreignSymbolInfo) *core.ForeignContract {
	initializedState, capture, retention, aliasing := standardForeignObligations()
	return &core.ForeignContract{
		Symbol: symbol.Name, Allocator: symbol.Allocator, Unwind: symbol.Unwind, NonlocalExit: symbol.NonlocalExit, Fails: symbol.Fails,
		InitializedState: initializedState, Capture: capture, Retention: retention, Aliasing: aliasing,
		Layout: standardForeignLayout(),
		Alias:  symbol.Alias,
	}
}

// resourceStep is check.go's own resolved bookkeeping for one step of a
// resource-lifecycle function: the block/place/type identity a step was
// assigned, plus the information the (D-04-07) reverse-order release
// materialization needs once the step completes.
type resourceStep struct {
	kind       string // "try_call" | "discard_call"
	symbol     foreignSymbolInfo
	callOpID   string
	okPlaceID  string
	errPlaceID string
	errTypeID  string
	blockID    string
	okEdgeID   string
	errEdgeID  string
}

// checkResourceLifecycle lowers a sequence of N try/discard foreign-call
// bindings (D-04-07/D-04-06) into N single-operation call blocks, one
// terminal err block per try_call step (releasing every completed try_call
// acquisition that precedes it, reverse of completion order), and one
// terminal success block (releasing every completed try_call acquisition, in
// full reverse order, then returning the function's own parameter -- never
// moved by any OpForeignCall, so it is always available to return once every
// acquired resource has been released).
//
// Order is decided by this file's own forward accumulation over the step
// sequence (mirroring the `expiringLoans[loan.lastUse] = append(...)` shape
// already established for loans in this file), never by map iteration: a
// release list derived from the emitter's own control-flow structure rather
// than from this materialized list is exactly the defect the three-
// acquisition fixture and the transposition mutation (Task 04-02-03) exist to
// catch.
func checkResourceLifecycle(functionID string, function ast.FuncDecl, parameterType core.TypeRef, derived ability.Result, typeID string, work int, foreignSymbols map[string]foreignSymbolInfo, functionNames map[string]bool, dataTypes map[string]core.DataType) (core.Function, []diagnostic.Diagnostic, int) {
	steps := function.Body.Linear.Bindings
	n := len(steps)
	parameterID := functionID + ":place:0"

	types := []core.TypeFact{{ID: typeID, Shape: parameterType, Abilities: derived.Granted, NegativeWitnesses: derived.NegativeWitnesses}}
	places := make([]core.Place, 1+n, 1+2*n)
	places[0] = core.Place{ID: parameterID, Name: function.Parameter.Name, TypeID: typeID}
	infos := make([]resourceStep, n)
	typeCounter := 1

	// corevalidate's targetMatches requires an OpForeignCall at flat
	// operation index i to produce place:(i+1) EXACTLY (the same
	// index-plus-one invariant every other transition operation obeys) --
	// so every step's ok place is assigned first, contiguously, in call
	// order (place:1..place:n), and every step's err place (never
	// index-checked -- only referenced by ErrTargetID) is appended
	// afterward, in step order, at place:(n+1)..place:(2n).
	for i, binding := range steps {
		symbol, errShape, errDerived, problem := resolveForeignStep(binding, function.Parameter.Name, foreignSymbols, functionNames, dataTypes)
		if problem != nil {
			return core.Function{}, []diagnostic.Diagnostic{*problem}, work
		}
		errTypeID := fmt.Sprintf("%s:type:%d", functionID, typeCounter)
		typeCounter++
		types = append(types, core.TypeFact{ID: errTypeID, Shape: errShape, Abilities: errDerived.Granted, NegativeWitnesses: errDerived.NegativeWitnesses})

		okPlaceID := fmt.Sprintf("%s:place:%d", functionID, i+1)
		errPlaceID := fmt.Sprintf("%s:place:%d", functionID, n+1+i)
		places[i+1] = core.Place{ID: okPlaceID, Name: binding.Name, TypeID: typeID}
		places = append(places, core.Place{ID: errPlaceID, Name: fmt.Sprintf("_err%d", i), TypeID: errTypeID})

		blockID := functionID + ":block:entry"
		if i > 0 {
			blockID = fmt.Sprintf("%s:block:step:%d", functionID, i)
		}
		infos[i] = resourceStep{
			kind: binding.RHS.Kind, symbol: symbol, callOpID: fmt.Sprintf("%s:op:%d", functionID, i),
			okPlaceID: okPlaceID, errPlaceID: errPlaceID, errTypeID: errTypeID, blockID: blockID,
		}
		work++
	}

	successBlockID := functionID + ":block:success"
	var blocks []core.Block
	var edges []core.Edge
	operations := make([]core.LinearOperation, n)
	for i := 0; i < n; i++ {
		okTarget := successBlockID
		if i+1 < n {
			okTarget = infos[i+1].blockID
		}
		errTarget := okTarget
		if infos[i].kind == "try_call" {
			errTarget = fmt.Sprintf("%s:block:err:%d", functionID, i)
		}
		infos[i].okEdgeID = fmt.Sprintf("%s:edge:step:%d:ok", functionID, i)
		infos[i].errEdgeID = fmt.Sprintf("%s:edge:step:%d:err", functionID, i)
		successors := []string{okTarget}
		if errTarget != okTarget {
			successors = append(successors, errTarget)
		}
		pointID := fmt.Sprintf("%s:point:step:%d", functionID, i)
		if i == 0 {
			// pathoracle/originvalidate require the entry block's PointID to
			// equal the function's own EntryPointID exactly (the same
			// convention checkForeignTracer and checkBranch already use).
			pointID = functionID + ":point:entry"
		}
		blocks = append(blocks, core.Block{
			ID: infos[i].blockID, PointID: pointID,
			OperationIDs: []string{infos[i].callOpID}, Successors: successors,
		})
		edges = append(edges,
			core.Edge{ID: infos[i].okEdgeID, FromBlockID: infos[i].blockID, ToBlockID: okTarget, Pattern: "ok"},
			core.Edge{ID: infos[i].errEdgeID, FromBlockID: infos[i].blockID, ToBlockID: errTarget, Pattern: "err"},
		)
		operations[i] = core.LinearOperation{
			ID: infos[i].callOpID, PointID: fmt.Sprintf("%s:point:linear:%d", functionID, i), Kind: core.OpForeignCall,
			SourceID: parameterID, TargetID: infos[i].okPlaceID, TypeID: typeID,
			OkEdgeID: infos[i].okEdgeID, ErrEdgeID: infos[i].errEdgeID, ErrTargetID: infos[i].errPlaceID,
			Allocator: infos[i].symbol.Allocator,
		}
	}

	// completed accumulates try_call steps in completion order as the
	// forward walk below reaches each one -- the single materialization
	// point D-04-07 requires. discard_call steps are never appended: their
	// acquired resource is not tracked for release this plan (a documented
	// narrowing -- see the plan's flagged_assumptions).
	var completed []resourceStep
	releaseOps := func(from []resourceStep) []core.LinearOperation {
		ops := make([]core.LinearOperation, 0, len(from))
		for j := len(from) - 1; j >= 0; j-- {
			acquired := from[j]
			opID := fmt.Sprintf("%s:op:%d", functionID, len(operations)+len(ops))
			ops = append(ops, core.LinearOperation{
				ID: opID, PointID: fmt.Sprintf("%s:point:linear:%d", functionID, len(operations)+len(ops)),
				Kind: core.OpRelease, SourceID: acquired.okPlaceID, TypeID: typeID, ReleasesOperationID: acquired.callOpID,
				// Allocator is copied verbatim from the acquisition this
				// release discharges (T-04-14): check.go is the sole
				// producer of both fields from the SAME acquired.symbol
				// value, so this can never disagree with itself here --
				// see releaseAllocatorMismatch below for check's own
				// defensive re-assertion of that invariant, and
				// corevalidate's independent re-derivation for a core
				// artifact this admission gate never produced.
				Allocator: acquired.symbol.Allocator,
			})
		}
		return ops
	}

	for i := 0; i < n; i++ {
		if infos[i].kind != "try_call" {
			continue
		}
		errOps := releaseOps(completed)
		failOpIndex := len(operations) + len(errOps)
		failOpID := fmt.Sprintf("%s:op:%d", functionID, failOpIndex)
		errOps = append(errOps, core.LinearOperation{
			ID: failOpID, PointID: fmt.Sprintf("%s:point:linear:%d", functionID, failOpIndex),
			Kind: core.OpFail, SourceID: infos[i].errPlaceID, TypeID: infos[i].errTypeID,
		})
		errOpIDs := make([]string, len(errOps))
		for idx, op := range errOps {
			errOpIDs[idx] = op.ID
		}
		blocks = append(blocks, core.Block{
			ID: fmt.Sprintf("%s:block:err:%d", functionID, i), PointID: fmt.Sprintf("%s:point:err:%d", functionID, i), OperationIDs: errOpIDs,
		})
		operations = append(operations, errOps...)
		completed = append(completed, infos[i])
	}

	successOps := releaseOps(completed)
	returnOpIndex := len(operations) + len(successOps)
	returnOpID := fmt.Sprintf("%s:op:%d", functionID, returnOpIndex)
	successOps = append(successOps, core.LinearOperation{
		ID: returnOpID, PointID: fmt.Sprintf("%s:point:linear:%d", functionID, returnOpIndex),
		Kind: core.OpReturn, SourceID: parameterID, TypeID: typeID,
	})
	successOpIDs := make([]string, len(successOps))
	for idx, op := range successOps {
		successOpIDs[idx] = op.ID
	}
	blocks = append(blocks, core.Block{ID: successBlockID, PointID: functionID + ":point:success", OperationIDs: successOpIDs})
	operations = append(operations, successOps...)

	if problem := releaseAllocatorMismatch(functionID, operations); problem != nil {
		return core.Function{}, []diagnostic.Diagnostic{*problem}, work
	}

	linear := &core.LinearBody{
		ID: functionID + ":linear", Types: types, Places: places, Operations: operations, Blocks: blocks, Edges: edges,
	}
	first := infos[0].symbol
	contract := buildForeignContract(first)

	return core.Function{
		ID: functionID, Name: function.Name,
		EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
		Parameter:       core.Parameter{ID: parameterID, Name: function.Parameter.Name, Type: parameterType.Constructor},
		ReturnType:      function.ReturnType.Constructor,
		Linear:          linear,
		ForeignContract: contract,
		Span:            function.Span,
	}, nil, work
}

// releaseAllocatorMismatch is check.go's own defensive re-assertion of
// T-04-14's allocator-identity requirement: for every OpRelease this
// admission gate just emitted, re-fetch the OpForeignCall it names
// (ReleasesOperationID) and require its Allocator field to match. Since
// check.go is the sole producer of both fields from the same acquired
// resourceStep, this can never fire against an honestly-constructed core
// artifact today -- it exists as defense-in-depth against a future bug in
// this emission path, alongside corevalidate's independent re-derivation
// against an artifact this gate never produced at all (a hand-mutated or
// corrupted core.Program).
func releaseAllocatorMismatch(functionID string, operations []core.LinearOperation) *diagnostic.Diagnostic {
	byID := make(map[string]core.LinearOperation, len(operations))
	for _, operation := range operations {
		byID[operation.ID] = operation
	}
	for _, operation := range operations {
		if operation.Kind != core.OpRelease {
			continue
		}
		acquisition, ok := byID[operation.ReleasesOperationID]
		if !ok || acquisition.Allocator == operation.Allocator {
			continue
		}
		causes := []diagnostic.Cause{
			{Kind: "release", Detail: operation.ID},
			{Kind: "acquisition", Detail: acquisition.ID},
		}
		problem := diagnostic.ErrorWithRepairs(
			"foreign.release_allocator_mismatch", diagnostic.Span{},
			"a release's declared allocator differs from its acquisition's", causes,
			diagnostic.Repair{Kind: "match_acquisition_allocator"},
		)
		return &problem
	}
	return nil
}

// missingForeignPolicyDiagnostic is Task 4's admission gate stub (D-04-16):
// a foreign declaration missing its unwind or nonlocal_exit policy is
// refused, independently, by both check and corevalidate, with no default
// value. It is implemented in Task 4; this plan's tracer fixture always
// supplies both policies, so this always returns nil today.
func missingForeignPolicyDiagnostic(symbol foreignSymbolInfo) *diagnostic.Diagnostic {
	missing := ""
	switch {
	case !symbol.HasUnwind:
		missing = "unwind"
	case !symbol.HasNonlocal:
		missing = "nonlocal_exit"
	default:
		return nil
	}
	causes := []diagnostic.Cause{
		{Kind: "foreign_symbol", Detail: symbol.Name},
		{Kind: "missing_policy", Detail: missing},
	}
	problem := diagnostic.ErrorWithRepairs(
		"foreign.unwind_policy_undeclared", symbol.Span,
		"a foreign symbol must declare both an unwind and a nonlocal_exit policy, with no default", causes,
		diagnostic.Repair{Kind: "declare_unwind_policy"},
	)
	return &problem
}

type loanFinalUseFact struct {
	LoanID         string
	Binding        string
	OperationIndex int
}

type ownershipStateFact struct {
	OperationIndex    int
	InitializedPlaces []string
	ActiveLoans       []string
}

type ownershipSupport struct {
	Places         []core.Place
	Operations     []core.LinearOperation
	LoanFinalUses  []loanFinalUseFact
	States         []ownershipStateFact
	Work           int
	DiagnosticCode string
	Diagnostic     *diagnostic.Diagnostic

	// FixpointWork is loanLivenessFixpoint's own counted cost for this
	// straight-line body (via computeLoanLastUses, D-05-35(d)'s sole
	// admission-deciding law), computed but deliberately NOT folded into Work
	// above. It is kept separate from Work: Work is the field
	// TestOwnershipSequenceExhaustive/TestOwnershipWorkSeries pin exactly
	// against an independent oracle (and a hand-derived formula) that predates
	// D-05-35 and has no way to reproduce loanLivenessFixpoint's own
	// internal, loan-chain-shape-dependent cost. checkLinear folds
	// FixpointWork into its own function-level RecomputedWork total (the same
	// place checkBranch already folds its own loanLivenessFixpoint call's
	// work), so the fixpoint's real cost is honestly counted, not hidden --
	// only kept out of the one field an independent oracle already pins
	// exactly (D-03-01/D-04-25's "does not silently add uncounted cost"
	// basis, applied without disturbing that pin).
	FixpointWork int

	// CallSpans is D-07-35's emission-time bookkeeping: for every "call"
	// binding admitted in this body, the resulting core.OpCall operation's
	// own ID mapped to the AST call-site span it came from
	// (binding.RHS.Span). No Span is added to core.LinearOperation itself
	// -- that would move bytes in every existing core artifact, against
	// D-07-08's whole point -- so this map is check's OWN in-memory
	// projection, built once per function body and merged program-wide in
	// Program(), read only when a core.call_graph_cycle diagnostic must
	// project an operation ID to a span.
	CallSpans map[string]diagnostic.Span
}

type placeState struct {
	place        core.Place
	declared     diagnostic.Span
	initialized  bool
	movedAt      *diagnostic.Span
	moveTargetID string
	// moveTargetName is the surface-level binding name the move landed on
	// (target.Name at the moment moveTargetID was set), kept alongside the
	// internal place ID purely so analyzeStraightLine's useAfterMove can
	// build a source-level MachineApplicable repair (D-06-24/D-06-25's move
	// injector): a diagnostic can address a place by ID, but a repair must
	// write source text, and only the surface name is source text. Set only
	// where moveTargetID is set; analyzeArmBody leaves it empty (its own
	// use_after_move repair stays classification-only, matching this file's
	// existing asymmetry for insert_take).
	moveTargetName string
}

// D-05-35(d): computeLoanLastUses (loanLivenessFixpoint's own last-use
// derivation) is now the sole law deciding conflict/expiry here --
// discoverLoanLastUses is retired, authorized by D-05-35(b)/(c)'s recorded
// zero-divergence shadow run over the full TestOwnershipSequenceExhaustive/
// TestBranchSequenceExhaustive enumeration.
//
// The discoveryWork term folded into result.Work below is deliberately kept
// as the historical flat len(body.Bindings)+1 accounting convention (not
// computeLoanLastUses' own fixpoint.work, which varies with loan-chain shape)
// -- TestOwnershipSequenceExhaustive/TestOwnershipWorkSeries pin this exact
// field against an independent oracle/hand-derived formula that predates
// D-05-35 and has no way to reproduce loanLivenessFixpoint's own internal
// cost shape. The fixpoint's REAL, honestly-varying cost is not hidden: it is
// counted separately via FixpointWork (see its own doc comment), folded into
// checkLinear's function-level RecomputedWork total.
func analyzeStraightLine(functionID, parameterName string, parameterSpan diagnostic.Span, typeFact core.TypeFact, body *ast.LinearBody, calleeContracts map[string]calleeContract, foreignSymbols map[string]foreignSymbolInfo) ownershipSupport {
	parameterID := functionID + ":place:0"
	result := ownershipSupport{
		Places:     []core.Place{{ID: parameterID, Name: parameterName, TypeID: typeFact.ID}},
		Operations: []core.LinearOperation{}, LoanFinalUses: []loanFinalUseFact{}, States: []ownershipStateFact{},
		Work: typeNodeCount(typeFact.Shape) + len(body.Bindings) + 1,
	}
	// D-09-08/D-09-09: no admission decision is made here anymore (see
	// analyzeArmBody's identical doc comment above). loanUses/FixpointWork
	// come from loanFinalUseEvidence -- the deleted computeLoanLastUses'
	// evidence-only replacement -- called UP FRONT exactly as before, so
	// LoanFinalUses stays populated for the FULL declared body even when a
	// later per-binding fact fails partway through. Work keeps the SAME
	// historical flat len(body.Bindings)+1 accounting convention
	// TestOwnershipSequenceExhaustive/TestOwnershipWorkSeries pin exactly
	// against an independent oracle; the fixpoint's real, honestly-varying
	// cost is counted separately via FixpointWork (see its own doc comment).
	loanUses, fixpointWork := loanFinalUseEvidence(parameterName, body)
	result.FixpointWork = fixpointWork
	result.Work += len(body.Bindings) + 1
	for index, binding := range body.Bindings {
		if binding.RHS.Kind == "borrow" || binding.RHS.Kind == "borrow_mut" {
			result.LoanFinalUses = append(result.LoanFinalUses, loanFinalUseFact{
				LoanID: fmt.Sprintf("%s:loan:%d", functionID, index), Binding: binding.Name, OperationIndex: loanUses[index].index,
			})
		}
	}
	places := map[string]*placeState{
		parameterName: {place: result.Places[0], declared: parameterSpan, initialized: true},
	}
	activeLoans := make(map[string]map[string]int) // ownerID -> loanID -> lastUse (evidence only, D-09-08/D-09-09)
	expiringLoans := make(map[int][]struct {
		ownerID string
		loanID  string
	})
	endLoans := func(index int) {
		for _, loan := range expiringLoans[index] {
			if loans := activeLoans[loan.ownerID]; loans != nil {
				delete(loans, loan.loanID)
				if len(loans) == 0 {
					delete(activeLoans, loan.ownerID)
				}
			}
		}
	}
	fail := func(problem diagnostic.Diagnostic) ownershipSupport {
		result.DiagnosticCode = problem.Code
		result.Diagnostic = &problem
		return result
	}
	useAfterMove := func(span diagnostic.Span, state *placeState) diagnostic.Diagnostic {
		causes := []diagnostic.Cause{
			{Kind: "declared_here", Span: spanPointer(state.declared)},
			{Kind: "moved_here", Span: state.movedAt},
			{Kind: "place", Detail: state.place.ID},
			{Kind: "transfer_target", Detail: state.moveTargetID},
			{Kind: "type", Detail: state.place.TypeID},
		}
		// repairSpan is a copy of span (the exact identifier text that named
		// the already-moved place), never a pointer aliasing the caller's own
		// span value. Replacement is the moved-to binding's surface name
		// (moveTargetName), the source-level counterpart of moveTargetID:
		// swapping the stale name for the live one is the mechanical fix a
		// use-after-move repair applies (D-06-24/D-06-25's move injector).
		// moveTargetName is only ever unset if moveTargetID is also unset,
		// in which case Replacement stays "" and DriverEligible is false by
		// construction — fail-closed, no explicit guard needed.
		repairSpan := span
		return diagnostic.ErrorWithRepairs(
			"ownership.use_after_move", span, "value was used after ownership transferred", causes,
			diagnostic.Repair{
				Kind:          "use_transfer_target",
				Detail:        state.moveTargetID,
				Span:          &repairSpan,
				Replacement:   state.moveTargetName,
				Applicability: diagnostic.ApplicabilityMachineApplicable,
			},
			diagnostic.Repair{Kind: "move_use_before_transfer"},
		)
	}
	for index, binding := range body.Bindings {
		result.Work++
		if binding.RHS.Kind == "call" {
			op, target, diag := resolveCallBinding(functionID, index, binding, places, calleeContracts, typeFact, foreignSymbols)
			if diag != nil {
				return fail(*diag)
			}
			result.Places = append(result.Places, target)
			places[binding.Name] = &placeState{place: target, declared: binding.Span, initialized: true}
			result.Operations = append(result.Operations, op)
			if result.CallSpans == nil {
				result.CallSpans = map[string]diagnostic.Span{}
			}
			result.CallSpans[op.ID] = binding.RHS.Span
			endLoans(index)
			result.States = append(result.States, ownershipSnapshot(index, places, activeLoans))
			continue
		}
		source, ok := places[binding.RHS.Source]
		if !ok {
			return fail(diagnostic.Error("name.unknown", binding.RHS.Span, "binding source is unknown"))
		}
		if !source.initialized {
			return fail(useAfterMove(binding.RHS.Span, source))
		}
		target := core.Place{ID: fmt.Sprintf("%s:place:%d", functionID, index+1), Name: binding.Name, TypeID: source.place.TypeID}
		kind := core.OpCopy
		loanIDStr := ""
		var newLoanOwnerID string
		var newLoanLastUse int
		switch binding.RHS.Kind {
		case "take":
			// D-09-08/D-09-09: no activeLoans conflict check here -- lowering
			// emits core.OpMove unconditionally, and the post-assembly pass
			// decides ownership.move_while_borrowed (D-09-07).
			kind = core.OpMove
			source.initialized = false
			source.movedAt = spanPointer(binding.RHS.Span)
			source.moveTargetID = target.ID
			source.moveTargetName = target.Name
		case "borrow":
			// A shared loan requires the share ability, exactly as the
			// implicit-copy path below requires copy. Without this gate the
			// checker emits core.OpBorrowShared that the source-blind
			// validator rejects as core.ability.share_denied, turning an
			// invalid program into a spanless operational failure.
			if !hasTypeAbility(typeFact, core.AbilityShare) {
				causes := []diagnostic.Cause{
					{Kind: "declared_here", Span: spanPointer(source.declared)},
					{Kind: "missing_ability", Detail: missingAbilityDetail(typeFact, core.AbilityShare)},
					{Kind: "place", Detail: source.place.ID},
					{Kind: "type", Detail: source.place.TypeID},
				}
				return fail(diagnostic.ErrorWithRepairs(
					"ownership.borrow_requires_share", binding.RHS.Span, "type does not grant the share ability required to borrow", causes,
					diagnostic.Repair{Kind: "use_take_instead"},
				))
			}
			// D-09-08/D-09-09: no conflictingLoan check here -- lowering
			// emits core.OpBorrowShared unconditionally, and the post-assembly
			// pass decides ownership.borrow_conflict (D-09-07). activeLoans
			// below is EVIDENCE bookkeeping only (feeds States.ActiveLoans).
			kind = core.OpBorrowShared
			loanIDStr = fmt.Sprintf("%s:loan:%d", functionID, index)
			newLoanOwnerID, newLoanLastUse = source.place.ID, loanUses[index].index
		case "borrow_mut":
			// An exclusive loan is gated on the same share-ability requirement
			// as a shared loan (see the comment on the identical gate above).
			if !hasTypeAbility(typeFact, core.AbilityShare) {
				causes := []diagnostic.Cause{
					{Kind: "declared_here", Span: spanPointer(source.declared)},
					{Kind: "missing_ability", Detail: missingAbilityDetail(typeFact, core.AbilityShare)},
					{Kind: "place", Detail: source.place.ID},
					{Kind: "type", Detail: source.place.TypeID},
				}
				return fail(diagnostic.ErrorWithRepairs(
					"ownership.borrow_requires_share", binding.RHS.Span, "type does not grant the share ability required to borrow", causes,
					diagnostic.Repair{Kind: "use_take_instead"},
				))
			}
			kind = core.OpBorrowExclusive
			loanIDStr = fmt.Sprintf("%s:loan:%d", functionID, index)
			newLoanOwnerID, newLoanLastUse = source.place.ID, loanUses[index].index
		default:
			if !hasTypeAbility(typeFact, core.AbilityCopy) {
				causes := []diagnostic.Cause{
					{Kind: "declared_here", Span: spanPointer(source.declared)},
					{Kind: "missing_ability", Detail: missingAbilityDetail(typeFact, core.AbilityCopy)},
					{Kind: "place", Detail: source.place.ID},
					{Kind: "type", Detail: source.place.TypeID},
				}
				takeSpan := binding.RHS.Span
				return fail(diagnostic.ErrorWithRepairs(
					"ownership.transfer_requires_take", binding.RHS.Span, "noncopyable binding requires explicit take", causes,
					diagnostic.Repair{
						Kind:          "insert_take",
						Span:          &takeSpan,
						Replacement:   "take " + binding.RHS.Source,
						Applicability: diagnostic.ApplicabilityMachineApplicable,
					},
				))
			}
		}
		if loanIDStr != "" {
			if activeLoans[newLoanOwnerID] == nil {
				activeLoans[newLoanOwnerID] = make(map[string]int)
			}
			activeLoans[newLoanOwnerID][loanIDStr] = newLoanLastUse
			expiringLoans[newLoanLastUse] = append(expiringLoans[newLoanLastUse], struct {
				ownerID string
				loanID  string
			}{ownerID: newLoanOwnerID, loanID: loanIDStr})
		}
		result.Places = append(result.Places, target)
		places[binding.Name] = &placeState{place: target, declared: binding.Span, initialized: true}
		operation := core.LinearOperation{
			ID: fmt.Sprintf("%s:op:%d", functionID, index), PointID: fmt.Sprintf("%s:point:linear:%d", functionID, index),
			Kind: kind, SourceID: source.place.ID, TargetID: target.ID, LoanID: loanIDStr, TypeID: source.place.TypeID,
		}
		result.Operations = append(result.Operations, operation)
		// D-08-22: see analyzeArmBody's identical span-widening comment.
		if result.CallSpans == nil {
			result.CallSpans = map[string]diagnostic.Span{}
		}
		result.CallSpans[operation.ID] = binding.RHS.Span
		// D-09-08/D-09-09: also record the WHOLE binding statement's span
		// under a synthesized ":stmt" suffix key, in the SAME channel
		// (merged into spanByOperationID exactly like every other CallSpans
		// entry) -- needed ONLY by borrowConflictDiagnosticPostAssembly's
		// narrow_to_shared_borrow repair, whose Replacement text rewrites
		// the ENTIRE "let NAME = borrow mut SOURCE" statement, never just
		// the RHS token binding.RHS.Span already carries.
		result.CallSpans[operation.ID+":stmt"] = binding.Span
		endLoans(index)
		result.States = append(result.States, ownershipSnapshot(index, places, activeLoans))
	}
	result.Work++
	returned, ok := places[body.Result]
	if !ok {
		return fail(diagnostic.Error("name.unknown", body.Span, "linear result is unknown"))
	}
	if !returned.initialized {
		return fail(useAfterMove(body.Span, returned))
	}
	ordinal := len(body.Bindings)
	result.Operations = append(result.Operations, core.LinearOperation{
		ID: fmt.Sprintf("%s:op:%d", functionID, ordinal), PointID: fmt.Sprintf("%s:point:linear:%d", functionID, ordinal),
		Kind: core.OpReturn, SourceID: returned.place.ID, TypeID: returned.place.TypeID,
	})
	endLoans(ordinal)
	result.States = append(result.States, ownershipSnapshot(ordinal, places, activeLoans))
	return result
}

type loanUse struct {
	index int
	span  diagnostic.Span
}

// loanFinalUseEvidence is D-09-08/D-09-09's evidence-only replacement for the
// deleted computeLoanLastUses. It derives, for every loan, its last
// transitively-derived use (binding index -> last use), via
// loanLivenessFixpoint's backward worklist dataflow -- the SAME shared
// machinery checkBranch's arm blocks and checkInterproceduralLoanLiveness's
// post-assembly pass already use (D-09-07: one algorithm, never a second
// law). It is called UNCONDITIONALLY, before any admission outcome is known,
// because LoanFinalUses/States are test-visible evidence fields
// (TestOwnershipSequenceExhaustive's byte-identical comparison) that an
// independent oracle populates from the FULL declared body regardless of
// which per-binding fact later fails -- so this function must be able to
// answer for the whole raw body too, not merely a truncated prefix an early
// `fail()` return happened to build real operations for.
//
// Unlike the deleted function, this one NEVER decides anything: no caller
// consults its output to accept or reject a binding (D-09-08's authorized
// reversal removed the last such consultation). It builds the identical
// synthetic, single-block "shadow:*" operation stream from body's bindings
// (kind/source only) that the deleted function built, then runs
// loanLivenessFixpoint/materializeLoanEndpoints over it to recover each
// loan's last-use binding index -- mechanically unchanged from before, only
// its role (evidence, not law) and its name have changed.
func loanFinalUseEvidence(parameterName string, body *ast.LinearBody) (map[int]loanUse, int) {
	const parameterPlaceID = "evidence:place:parameter"
	visible := map[string]string{parameterName: parameterPlaceID}
	operations := make([]core.LinearOperation, 0, len(body.Bindings)+1)
	loanIndexByLoanID := make(map[string]int, len(body.Bindings))
	for index, binding := range body.Bindings {
		var sourcePlaceID string
		if binding.RHS.Kind == "call" {
			// A call binding carries RHS.Arguments, not RHS.Source: resolve
			// from the call's own sole argument (arity 1) so the call becomes
			// a visible reference to that argument's loan chain in
			// blockLoanLiveness's backward walk.
			sourcePlaceID = fmt.Sprintf("evidence:place:unknown:%d", index)
			if len(binding.RHS.Arguments) == 1 {
				if placeID, ok := visible[binding.RHS.Arguments[0]]; ok {
					sourcePlaceID = placeID
				}
			}
		} else if placeID, ok := visible[binding.RHS.Source]; ok {
			sourcePlaceID = placeID
		} else {
			// An out-of-scope/unknown source is a real-admission rejection
			// (name.unknown): give it a place ID no other binding can ever
			// produce so the chain simply carries no inherited loan through
			// it.
			sourcePlaceID = fmt.Sprintf("evidence:place:unknown:%d", index)
		}
		targetPlaceID := fmt.Sprintf("evidence:place:%d", index)
		kind := core.OpCopy
		loanID := ""
		switch binding.RHS.Kind {
		case "take":
			kind = core.OpMove
		case "borrow":
			kind = core.OpBorrowShared
			loanID = fmt.Sprintf("evidence:loan:%d", index)
		case "borrow_mut":
			kind = core.OpBorrowExclusive
			loanID = fmt.Sprintf("evidence:loan:%d", index)
		case "call":
			kind = core.OpCall
		}
		operations = append(operations, core.LinearOperation{
			ID: fmt.Sprintf("evidence:op:%d", index), Kind: kind, SourceID: sourcePlaceID, TargetID: targetPlaceID, LoanID: loanID,
		})
		if loanID != "" {
			loanIndexByLoanID[loanID] = index
		}
		visible[binding.Name] = targetPlaceID
	}
	resultOperationIndex := len(body.Bindings)
	resultSourceID := parameterPlaceID
	if placeID, ok := visible[body.Result]; ok {
		resultSourceID = placeID
	}
	operations = append(operations, core.LinearOperation{
		ID: fmt.Sprintf("evidence:op:%d", resultOperationIndex), Kind: core.OpReturn, SourceID: resultSourceID,
	})

	blockID := "evidence:block:straight"
	block := cfgBlockSpec{id: blockID, operations: operations, successors: nil}
	uses := make(map[int]loanUse, len(loanIndexByLoanID))
	fixpoint, diag := loanLivenessFixpoint("evidence", []cfgBlockSpec{block}, interproceduralSummaryTable{}, diagnostic.Span{})
	if diag != nil {
		return uses, 0
	}
	edgeID := func(from, to string) string { return from + "->" + to }
	endpoints := materializeLoanEndpoints("evidence", []cfgBlockSpec{block}, edgeID, fixpoint, interproceduralSummaryTable{})

	operationIndexByID := make(map[string]int, len(operations))
	for opIndex, operation := range operations {
		operationIndexByID[operation.ID] = opIndex
	}
	for _, endpoint := range endpoints {
		loanIndex, ok := loanIndexByLoanID[endpoint.LoanID]
		if !ok {
			continue
		}
		lastOperationIndex := operationIndexByID[endpoint.AfterOperationID]
		bindingIndex := lastOperationIndex
		span := body.Span
		if bindingIndex < len(body.Bindings) {
			span = body.Bindings[bindingIndex].RHS.Span
		}
		uses[loanIndex] = loanUse{index: bindingIndex, span: span}
	}
	return uses, fixpoint.work
}

// AliasFact is D-05-01's optimizer-facing alias-lattice fact: an
// independently-checked promise that a function's sole parameter is covered,
// for the ENTIRE duration of that function's own execution (C17 section
// 6.7.3.1's "for the duration of that function's execution", not merely "at
// some point"), by a single loan chain rooted at the parameter itself. It
// exists to give NAT-03's optimizer-attribute controls an honest subject
// this phase (D-05-01): cgen's restrict emission and corevalidate's
// independent re-derivation both consume this shape, sharing no helper with
// each other or with this derivation (D-12 -- three independent
// derivations: check, cgen, corevalidate).
type AliasFact struct {
	Function  string
	Parameter string
	LoanID    string
	// Kind is AliasFactExclusiveBorrow or AliasFactUniqueOwner. Only
	// AliasFactExclusiveBorrow is derivable this phase: cgen's by-pointer
	// lowering (selectsByPointerLowering) never admits a moved-only owned
	// chain, so AliasFactUniqueOwner has no fixture and no lowering to
	// justify this plan ever emitting it -- the constant exists so the
	// lattice's second half is named, not silently absent, for a later plan
	// to derive.
	Kind string
}

const (
	AliasFactExclusiveBorrow = "exclusive_borrow"
	AliasFactUniqueOwner     = "unique_owner"
)

// aliasFactEndpoints computes the SAME backward-worklist loan-liveness
// fixpoint checkBranch's arm blocks already use (loanLivenessFixpoint plus
// materializeLoanEndpoints), over a straight-line function's own REAL,
// already-admitted operation stream (real per-function IDs, not the
// synthetic "shadow:" IDs computeLoanLastUses builds during admission
// itself). It is called only AFTER checkLinear has already produced the
// core.Function -- this is a POST-HOC re-derivation for deriveAliasFacts,
// never a second admission-deciding law: computeLoanLastUses remains the
// sole law deciding conflict/expiry (D-05-35(d)). A closed single-block CFG
// (no successors) never produces an edge endpoint, matching
// computeLoanLastUses' own single-block construction.
func aliasFactEndpoints(functionID string, operations []core.LinearOperation) ([]core.LoanEndpoint, int) {
	block := cfgBlockSpec{id: functionID + ":block:straight", operations: operations, successors: nil}
	fixpoint, diag := loanLivenessFixpoint(functionID, []cfgBlockSpec{block}, interproceduralSummaryTable{}, diagnostic.Span{})
	if diag != nil {
		return nil, 0
	}
	edgeID := func(from, to string) string { return from + "->" + to }
	endpoints := materializeLoanEndpoints(functionID, []cfgBlockSpec{block}, edgeID, fixpoint, interproceduralSummaryTable{})
	return endpoints, fixpoint.work
}

// deriveAliasFacts is D-05-01's admission-gating alias-fact derivation. It
// returns a fact ONLY when function's sole parameter is covered by an
// exclusive loan across every operation from the function's first use of the
// parameter to its own terminator -- deliberately the SAME structural
// condition cgen's selectsByPointerLowering checks (D-05-02), duplicated
// rather than shared (D-12), PLUS grounding the "for the whole call" claim
// in the liveness fixpoint's own materialized endpoints (never
// re-implementing last-use discovery itself): the exclusive loan's endpoint
// must be a "point" endpoint whose AfterOperationID is the function's own
// terminator, not merely somewhere earlier in the operation stream -- this
// is exactly what rejects a loan whose last use is before the terminator
// (a partially-covering loan never produces a fact). Counts its own work
// (one unit per operation inspected, matching the existing per-fact loops'
// accounting basis) as its second return value.
func deriveAliasFacts(function core.Function, linear *core.LinearBody, endpoints []core.LoanEndpoint) ([]AliasFact, int) {
	work := 0
	if linear == nil {
		return nil, work
	}
	// PublicOrigin != nil marks a declared borrow-return function (OWN-04's
	// public borrowed views, e.g. testdata/phase3/public_view_mixed_access.lang):
	// a distinct, already-shipped semantic category that can have the exact
	// same exclusive-borrow-then-reborrow-to-terminator operation shape as
	// this plan's own by-pointer fixture. Excluding it here mirrors
	// selectsByPointerLowering's own identical guard (cgen.go) -- the one
	// genuinely structural fact (not a name or file match) that tells the
	// two apart, required for TestAliasFactAgreesWithByPointerSelection to
	// hold on testdata/phase3/public_view_mixed_access.lang.
	if function.PublicOrigin != nil {
		return nil, work
	}
	operations := linear.Operations
	if len(operations) == 0 {
		return nil, work
	}
	first := operations[0]
	work++
	if first.Kind != core.OpBorrowExclusive || first.SourceID != function.Parameter.ID || first.TargetID == "" {
		return nil, work
	}
	current := first.TargetID
	terminatorIndex := -1
	for index := 1; index < len(operations); index++ {
		work++
		operation := operations[index]
		if operation.SourceID == function.Parameter.ID {
			return nil, work
		}
		if operation.SourceID != current {
			return nil, work
		}
		if operation.Kind == core.OpReturn {
			terminatorIndex = index
			break
		}
		if operation.TargetID == "" {
			return nil, work
		}
		current = operation.TargetID
	}
	if terminatorIndex != len(operations)-1 {
		return nil, work
	}
	terminator := operations[terminatorIndex]
	for _, endpoint := range endpoints {
		work++
		if endpoint.LoanID != first.LoanID {
			continue
		}
		if endpoint.Kind == "point" && endpoint.AfterOperationID == terminator.ID {
			return []AliasFact{{Function: function.ID, Parameter: function.Parameter.ID, LoanID: first.LoanID, Kind: AliasFactExclusiveBorrow}}, work
		}
		return nil, work
	}
	return nil, work
}

// D-09-08 (authorized at plan 09-08's mid-phase gate): the summary-blind
// early call site formerly named here -- built from raw AST straight into a
// single synthetic "shadow:*" block, fed a permanent zero-value
// interproceduralSummaryTable{}, and consulted by analyzeArmBody/
// analyzeStraightLine to DECIDE ownership.move_while_borrowed/
// ownership.borrow_conflict before any core.Function existed -- is DELETED.
// It was never a second LAW (D-09-07): it called the SAME
// loanLivenessFixpoint/materializeLoanEndpoints that survive below, fed
// synthetic operations instead of real ones. checkInterproceduralLoanLiveness's
// post-assembly pass, extended to the purely-intraprocedural case, is now the
// SOLE decision point. loanFinalUseEvidence (below) re-derives the two
// test-visible evidence fields this deleted function used to feed, kept as
// pure evidence rather than an admission-deciding consultation.

// ownershipSnapshot captures a straight-line/arm-body walk's InitializedPlaces
// and ActiveLoans at operation index. activeLoans is EVIDENCE bookkeeping
// only (D-09-08/D-09-09): it never gates an admission decision here -- the
// post-assembly pass decides ownership.move_while_borrowed/borrow_conflict
// independently, over its own last-use derivation.
func ownershipSnapshot(index int, places map[string]*placeState, activeLoans map[string]map[string]int) ownershipStateFact {
	initialized := make([]string, 0, len(places))
	for _, place := range places {
		if place.initialized {
			initialized = append(initialized, place.place.ID)
		}
	}
	sort.Strings(initialized)
	loans := make([]string, 0)
	for _, ownerLoans := range activeLoans {
		for loanID := range ownerLoans {
			loans = append(loans, loanID)
		}
	}
	sort.Strings(loans)
	return ownershipStateFact{OperationIndex: index, InitializedPlaces: initialized, ActiveLoans: loans}
}

func hasTypeAbility(fact core.TypeFact, wanted core.Ability) bool {
	for _, candidate := range fact.Abilities {
		if candidate == wanted {
			return true
		}
	}
	return false
}

// executableShape reports whether a linear (non-match) function's parameter
// type has a native execution lowering this phase. cgen.linearInput only
// maps Byte and Buffer; every other shape (Box, Pair, or a nominal data type
// used in a straight-line body) type-checks and derives abilities cleanly
// but has no engine that can run it (D-02-09/D-07). Match-based functions
// (checkBranch included) are unaffected: their parameter is always a
// declared nominal alternative type, always executable via emitMatch/
// emitBranch's enum-based lowering, so this gate is only consulted from
// checkLinear.
func executableShape(value core.TypeRef) bool {
	return value.Constructor == "Byte" || value.Constructor == "Buffer"
}

func typeNodeCount(value core.TypeRef) int {
	count := 1
	for _, argument := range value.Arguments {
		count += typeNodeCount(argument)
	}
	return count
}

func missingAbilityDetail(fact core.TypeFact, requested core.Ability) string {
	for _, witness := range fact.NegativeWitnesses {
		if witness.Ability == requested {
			return string(requested) + ":" + strings.Join(witness.Path, ".")
		}
	}
	return string(requested)
}

func spanPointer(span diagnostic.Span) *diagnostic.Span {
	copy := span
	return &copy
}

func coreType(value ast.TypeRef) core.TypeRef {
	result := core.TypeRef{Constructor: value.Constructor, Arguments: make([]core.TypeRef, 0, len(value.Arguments))}
	for _, argument := range value.Arguments {
		result.Arguments = append(result.Arguments, coreType(argument))
	}
	return result
}

func sameType(left, right ast.TypeRef) bool {
	return typeKey(coreType(left)) == typeKey(coreType(right))
}

func typeKey(value core.TypeRef) string {
	if len(value.Arguments) == 0 {
		return value.Constructor
	}
	parts := make([]string, 0, len(value.Arguments))
	for _, argument := range value.Arguments {
		parts = append(parts, typeKey(argument))
	}
	return value.Constructor + "<" + strings.Join(parts, ",") + ">"
}

func semanticID(module, kind, name string) string { return "s1:" + module + ":" + kind + ":" + name }

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

// allArmsSelfMap reports whether every arm of a bare-arm match maps its own
// pattern to itself (arm.Pattern == arm.Value). This is the sole gate on
// match.non_exhaustive's add_missing_arm repair (D-06-24/D-06-25): a missing
// arm's correct target value is only a safe mechanical guess when the arms
// already present establish that identity-mapping convention. An empty arm
// list vacuously self-maps but is never reached with this gate's meaning
// intact, since the caller also requires len(function.Body.Arms) > 0.
func allArmsSelfMap(arms []ast.MatchArm) bool {
	for _, arm := range arms {
		if arm.Pattern != arm.Value {
			return false
		}
	}
	return true
}
