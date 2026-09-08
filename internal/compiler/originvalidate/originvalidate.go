// Package originvalidate independently re-derives and verifies a function's
// declared public borrow origin (OWN-04) from the typed-core artifact alone.
// It intentionally does not know source or reuse checker code: it imports
// neither internal/compiler/check nor internal/compiler/ast, so a declared
// origin is never trusted, only recomputed from core.Program/core.Function
// facts — the same source-blind posture corevalidate already holds for
// ownership, applied here to the origin trust boundary (T-03-02/T-03-03/
// T-03-16).
package originvalidate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/codename-lang/lang/internal/compiler/core"
)

// KnownEscape names the exact boundary this package cannot prove: a producer
// whose frontend and published summary lie in a coordinated way remains
// outside source-blind, body-blind validation. It mirrors
// corevalidate.KnownEscape's shape, applied to the origin trust boundary
// instead of the source/core one (T-03-08, an accepted residual, never
// claimed as solved).
const KnownEscape = "escape:coordinated-frontend-summary-lie"

// TerminatorKindsOverride is a fault-injection seam for
// TestTerminatorWalkMutationKilled (D-09/D-04-29): production always walks
// the full core.TerminatorKinds() set; the test temporarily narrows it (the
// same "delete OpFail from the recognised set" mutation the throwaway-
// detached-worktree demonstration performs on the source directly) to prove
// that a walker recognising fewer terminators really does lose a fixture's
// origin fact. nil (the always-true production default) means "use
// core.TerminatorKinds() unmodified".
var TerminatorKindsOverride func() []core.OperationKind

func recognizedTerminatorKinds() []core.OperationKind {
	if TerminatorKindsOverride != nil {
		return TerminatorKindsOverride()
	}
	return core.TerminatorKinds()
}

// isTerminatorKind is originvalidate's own membership test against the
// terminator registry (D-04-29): a set-membership test against
// core.TerminatorKinds() rather than a literal restatement of it, so
// widening the registry widens this walker automatically. This is the
// single site (RecomputeOriginPerReturn's collection loop, immediately
// below) where the backward-walk collection discriminates a terminator
// operation from an ordinary one.
func isTerminatorKind(kind core.OperationKind) bool {
	for _, terminator := range recognizedTerminatorKinds() {
		if kind == terminator {
			return true
		}
	}
	return false
}

// RecognizesTerminator is session.go's control:terminator.walk_incomplete
// hook (D-04-29): it reports whether this package's own walker treats kind
// as a terminator, reading the exact same isTerminatorKind this file's
// production walk uses -- never a second, restated copy of the set.
func RecognizesTerminator(kind core.OperationKind) bool { return isTerminatorKind(kind) }

// ExpectedEscapes is the origin package's contribution to a verify result's
// expected-escapes list — surfaced next to corevalidate.KnownEscape, never
// reported as a detected control.
func ExpectedEscapes() []string { return []string{KnownEscape} }

// Problem is the origin package's independent-validation finding shape,
// matching corevalidate.Problem's {Code, Detail} contract.
type Problem struct {
	Code   string `json:"code"`
	Detail string `json:"detail,omitempty"`
}

// Error is CheckSummary's typed failure, matching evidence.ValidationError's
// {Code}-only shape so callers can dispatch on Code the same way.
type Error struct{ Code string }

func (e *Error) Error() string { return e.Code }

// ReturnOrigin is the origin derivation for exactly ONE core.OpReturn within
// a function's body: which paths and access mode that single return's
// SourceID traces back to the function's own parameter, if it traces at
// all. A straight-line function's Operations carries exactly one
// core.OpReturn; a match-bodied function carries one per arm (check.go's
// arm lowering appends every arm's Return into the same flat Operations
// slice, per corevalidate.replayBlocks' own "one return per block" note) —
// so a function's full origin picture is the SET of these, never any one of
// them alone.
type ReturnOrigin struct {
	// OperationID is the originating core.OpReturn's own ID.
	OperationID string
	// Paths names the parameter path(s) this return traces back to, when
	// Derived is true. Empty when Derived is false.
	Paths []string
	// Access is the derived access mode ("shared" or "exclusive") when
	// Derived is true. Empty when Derived is false.
	Access string
	// Derived reports whether the backward walk from this return reached
	// function.Parameter.ID through at least one borrow hop. false means an
	// owned return, or a chain that broke/cycled before reaching the
	// parameter.
	Derived bool
}

// AccessConflicting is the conservative-combination sentinel RecomputeOrigin
// reports when a function's arms derive different access modes for their
// respective returns: neither arm's answer is the true answer, so the
// sentinel names the disagreement itself rather than silently resolving to
// whichever arm happened to be walked. It is never a declarable
// core.PublicOrigin.Access value — ValidatePublished refuses any declared
// Access outside {"shared", "exclusive"} before ever comparing it against a
// recomputed answer, so a mutated summary cannot declare the sentinel and
// match a conflicting recomputation.
const AccessConflicting = "conflicting"

// RecomputeOriginPerReturn is the package's SOLE backward-walk site (Task
// 03-10-01's binding decision: exactly one such loop may exist in this
// file). It walks backward from EVERY core.OpReturn in
// function.Linear.Operations, in operation order, using one TargetID->
// SourceOperation map built once over the whole flat operations list. That
// single shared map is safe to reuse across arms because check.go's arm
// lowering pads one unreferenced place per arm's Return, keeping every
// operation's place ordinal globally distinct across the whole function
// (corevalidate.replayBlocks' own place-order invariant) — so no arm's walk
// can ever cross into a sibling arm's operations. Each walk independently
// preserves 03-08's first-seen-hop-wins rule: the hop nearest the returned
// place decides that return's derived access mode, per return.
func RecomputeOriginPerReturn(function core.Function) []ReturnOrigin {
	if function.Linear == nil || len(function.Linear.Operations) == 0 {
		return nil
	}
	operations := function.Linear.Operations
	sourceOf := make(map[string]core.LinearOperation, len(operations))
	var returnOps []*core.LinearOperation
	for index := range operations {
		operation := operations[index]
		// D-04-29: the discriminant is a membership test against the
		// terminator registry, not an equality test against core.OpReturn
		// alone -- Phase 4 introduces core.OpFail and core.OpDefect, and an
		// analysis that recognises only a return is exactly the defect class
		// 03-08/03-09/03-10 closed three times, now reproduced one layer up.
		// walkReturnOrigin below is unchanged: it is already terminator-
		// agnostic, walking backward from whichever operation this loop
		// collects.
		if isTerminatorKind(operation.Kind) {
			returnOps = append(returnOps, &operations[index])
			continue
		}
		sourceOf[operation.TargetID] = operation
	}
	if len(returnOps) == 0 {
		return nil
	}
	results := make([]ReturnOrigin, 0, len(returnOps))
	for _, returnOp := range returnOps {
		results = append(results, walkReturnOrigin(function, sourceOf, returnOp))
	}
	return results
}

// walkReturnOrigin performs exactly one backward walk, from one return
// operation, using the shared sourceOf map RecomputeOriginPerReturn built
// once for the whole function.
func walkReturnOrigin(function core.Function, sourceOf map[string]core.LinearOperation, returnOp *core.LinearOperation) ReturnOrigin {
	current := returnOp.SourceID
	visited := make(map[string]bool)
	derivedAccess := ""
	for current != function.Parameter.ID {
		if visited[current] {
			return ReturnOrigin{OperationID: returnOp.ID}
		}
		visited[current] = true
		operation, exists := sourceOf[current]
		if !exists {
			return ReturnOrigin{OperationID: returnOp.ID}
		}
		switch operation.Kind {
		case core.OpBorrowExclusive:
			if derivedAccess == "" {
				derivedAccess = "exclusive"
			}
		case core.OpBorrowShared:
			if derivedAccess == "" {
				derivedAccess = "shared"
			}
		case core.OpForeignCall:
			// D-04-28: a foreign declaration is itself a signature carrying
			// origin and access facts -- declaring a call foreign-only does
			// not escape origin reasoning, it moves the facts somewhere they
			// are asserted (function.ForeignContract.Alias) rather than
			// derived from a body the foreign symbol does not have. When the
			// contract declares this call borrows/retains its argument, the
			// hop counts exactly like an in-language borrow hop so a
			// correctly DECLARED PublicOrigin for such a function is
			// recognised as matching, not flagged as understated.
			if derivedAccess == "" && function.ForeignContract != nil {
				switch function.ForeignContract.Alias {
				case "borrow":
					derivedAccess = "shared"
				case "retain":
					derivedAccess = "exclusive"
				}
			}
		}
		current = operation.SourceID
	}
	if derivedAccess == "" {
		return ReturnOrigin{OperationID: returnOp.ID}
	}
	return ReturnOrigin{OperationID: returnOp.ID, Paths: []string{function.Parameter.Name}, Access: derivedAccess, Derived: true}
}

// RecomputeOrigin derives the origin path(s) and access mode a function's
// body actually returns, as the conservative combination of every
// RecomputeOriginPerReturn element (the "combination law", 03-10-PLAN.md):
//
//  1. No return is borrow-derived (every arm owned, or every chain broke) →
//     (nil, "", false). Byte-identical to today for an owned straight-line
//     function or an every-arm-owned match function.
//  2. Every borrow-derived return agrees on access mode → the ordered union
//     of their paths, that agreed access, ok=true. Byte-identical to today
//     for a one-return function.
//  3. Borrow-derived returns DISAGREE on access mode → the ordered union of
//     their paths, access=AccessConflicting, ok=true — neither arm's answer
//     wins.
//
// This function performs no backward walk itself; RecomputeOriginPerReturn
// is the only place that does.
func RecomputeOrigin(function core.Function) (paths []string, access string, ok bool) {
	perReturn := RecomputeOriginPerReturn(function)
	var derived []ReturnOrigin
	for _, origin := range perReturn {
		if origin.Derived {
			derived = append(derived, origin)
		}
	}
	if len(derived) == 0 {
		return nil, "", false
	}
	pathSeen := make(map[string]bool, len(derived))
	var orderedPaths []string
	combinedAccess := derived[0].Access
	conflict := false
	for _, origin := range derived {
		if origin.Access != combinedAccess {
			conflict = true
		}
		for _, path := range origin.Paths {
			if !pathSeen[path] {
				pathSeen[path] = true
				orderedPaths = append(orderedPaths, path)
			}
		}
	}
	if conflict {
		return orderedPaths, AccessConflicting, true
	}
	return orderedPaths, combinedAccess, true
}

// ValidatePublished recomputes every function's origin from its body and
// compares it against the declaration, following Spike 003's
// producer-verification gates. It returns the first problem only, matching
// corevalidate's first-problem-only accumulation, so the stable assertion
// target is always the first defect. Recomputation now runs unconditionally
// for every function, including one with no declared PublicOrigin at all
// (D-03-02/GAP 2, ROADMAP SC4's "omitted" category): when recomputation
// succeeds for such a function, its body actually returns a borrow-derived
// place that publication would otherwise expose indistinguishably from a
// fully-owned return, and that is refused with core.origin_omitted. When
// recomputation instead reports not-ok (an owned return, or a match-bodied
// function with no linear return chain), the function is left untouched
// exactly as before — the new check does not over-fire on an honest,
// declaration-free owned value.
// checkForeignOriginOmitted independently re-derives, for one function,
// whether its returned value traces back to a foreign call whose declared
// contract says it borrows or retains its argument (D-04-28). It performs
// its own backward walk over function.Linear.Operations rather than sharing
// RecomputeOriginPerReturn's (T-04-37: two derivations must not agree merely
// because they share a law) -- it reads only the core artifact, never the
// checker: the foreign contract's declared Alias obligation and the
// operation's own SourceID/TargetID places.
func checkForeignOriginOmitted(function core.Function) *Problem {
	if function.ForeignContract == nil {
		return nil
	}
	alias := function.ForeignContract.Alias
	if alias != "borrow" && alias != "retain" {
		return nil
	}
	if function.Linear == nil || function.PublicOrigin != nil {
		return nil
	}
	operations := function.Linear.Operations
	sourceOf := make(map[string]core.LinearOperation, len(operations))
	var foreignCallTarget string
	for index := range operations {
		operation := operations[index]
		if operation.Kind == core.OpForeignCall {
			foreignCallTarget = operation.TargetID
		}
		if isTerminatorKind(operation.Kind) {
			continue
		}
		sourceOf[operation.TargetID] = operation
	}
	if foreignCallTarget == "" {
		return nil
	}
	for index := range operations {
		operation := operations[index]
		if operation.Kind != core.OpReturn {
			continue
		}
		current := operation.SourceID
		visited := make(map[string]bool)
		for current != function.Parameter.ID {
			if current == foreignCallTarget {
				return &Problem{
					Code: "core.foreign_origin_omitted",
					Detail: fmt.Sprintf(
						"%s: return derives from a foreign call declared %q on argument %q, but no public origin is declared",
						function.ID, alias, function.Parameter.Name,
					),
				}
			}
			if visited[current] {
				break
			}
			visited[current] = true
			source, exists := sourceOf[current]
			if !exists {
				break
			}
			current = source.SourceID
		}
	}
	return nil
}

func ValidatePublished(program core.Program) []Problem {
	for _, function := range program.Functions {
		if problem := checkForeignOriginOmitted(function); problem != nil {
			return []Problem{*problem}
		}
		recomputedPaths, recomputedAccess, ok := RecomputeOrigin(function)
		if function.PublicOrigin == nil {
			if ok {
				return []Problem{{
					Code:   "core.origin_omitted",
					Detail: fmt.Sprintf("%s: no declared origin, but body derives origin %v with access %q", function.ID, recomputedPaths, recomputedAccess),
				}}
			}
			continue
		}
		// Declared-access domain check (Task 03-10-02): a declared Access
		// outside {"shared", "exclusive"} is refused BEFORE any comparison
		// with the recomputed answer, so a mutated summary cannot declare
		// AccessConflicting and have it match a genuinely conflicting
		// recomputation.
		if function.PublicOrigin.Access != "shared" && function.PublicOrigin.Access != "exclusive" {
			return []Problem{{
				Code:   "core.origin_access_mismatch",
				Detail: fmt.Sprintf("%s: declared access %q is not a declarable mode", function.ID, function.PublicOrigin.Access),
			}}
		}
		if !ok || !containsAll(function.PublicOrigin.Paths, recomputedPaths) {
			return []Problem{{
				Code:   "core.origin_understated",
				Detail: fmt.Sprintf("%s: declared origin %v does not cover the body-derived origin %v", function.ID, function.PublicOrigin.Paths, recomputedPaths),
			}}
		}
		if function.PublicOrigin.Access != recomputedAccess {
			return []Problem{{
				Code:   "core.origin_access_mismatch",
				Detail: fmt.Sprintf("%s: declared access %q, body derives %q", function.ID, function.PublicOrigin.Access, recomputedAccess),
			}}
		}
	}
	return nil
}

// containsAll reports whether every recomputed path is present in the
// declared set — a declared set that omits a body-derivable path is
// understated.
func containsAll(declared, recomputed []string) bool {
	present := make(map[string]bool, len(declared))
	for _, path := range declared {
		present[path] = true
	}
	for _, path := range recomputed {
		if !present[path] {
			return false
		}
	}
	return true
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// ClosureDigestDomainSeparator is D-07-37's canonical ClosureDigest
// preimage's fixed prefix, declared exactly once here and referenced
// everywhere else the preimage is built. It is not itself a digest of
// anything -- it exists solely so a future digest with a superficially
// similar preimage shape can never collide with this one.
const ClosureDigestDomainSeparator = "lang.closure_digest/1\x00"

// calleeDigestPair is one callee's (ID, ClosureDigest) pair, part of
// D-07-37's canonical preimage. Field order is fixed by declaration (never
// map iteration), so json.Marshal of a []calleeDigestPair is deterministic
// regardless of insertion order -- SORTING the slice by ID (below) is what
// makes the preimage itself order-independent; this struct's own encoding
// was never order-dependent to begin with.
type calleeDigestPair struct {
	ID            string `json:"id"`
	ClosureDigest string `json:"closure_digest"`
}

// closureDigestSortOverride is D-07-42's unexported fault-injection seam
// for a same-package mutation-kill test (originvalidate_internal_test.go):
// production always sorts callee pairs by ID; the test temporarily
// replaces this with the identity function to prove the out-of-order test
// actually depends on the sort, not merely appears to. nil (the
// always-true production default) means "sort by ID".
var closureDigestSortOverride func([]calleeDigestPair) []calleeDigestPair

func sortCalleeDigestPairs(pairs []calleeDigestPair) []calleeDigestPair {
	sorted := append([]calleeDigestPair(nil), pairs...)
	if closureDigestSortOverride != nil {
		return closureDigestSortOverride(sorted)
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ID < sorted[j].ID })
	return sorted
}

// closureDigestPreimageBytes builds D-07-37's canonical, non-self-
// referential preimage for signature's ClosureDigest: the domain separator,
// then signature with ClosureDigest itself ZEROED (so the digest never
// depends on its own prior value — non-self-reference, Test 1), then the
// callee (ID, ClosureDigest) pairs SORTED BY ID (Test 4) so the preimage is
// independent of the order callees happen to be supplied in.
//
// D-07-38: this plan calls this only with an empty/nil callees slice — the
// zero-callee base case, decided explicitly. The chaining arm over REAL
// callees lands in 07-08-PLAN.md, after cycle refusal exists: the chain
// terminates only on a DAG, so computing it before that gate would let a
// cyclic program reach a non-terminating digest computation before the
// gate that would refuse it. This function's signature already accepts
// callees so 07-08 supplies them without changing this preimage
// definition.
func closureDigestPreimageBytes(signature core.FunctionSignature, callees []calleeDigestPair) ([]byte, error) {
	signature.ClosureDigest = ""
	payload := struct {
		Signature core.FunctionSignature `json:"signature"`
		Callees   []calleeDigestPair     `json:"callees"`
	}{Signature: signature, Callees: sortCalleeDigestPairs(callees)}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	preimage := make([]byte, 0, len(ClosureDigestDomainSeparator)+len(body))
	preimage = append(preimage, []byte(ClosureDigestDomainSeparator)...)
	preimage = append(preimage, body...)
	return preimage, nil
}

// computeClosureDigest computes D-07-37's canonical ClosureDigest for one
// function signature and its (possibly empty) callee set, reusing the
// shared digest() helper — no second digest format is minted anywhere in
// this package.
func computeClosureDigest(signature core.FunctionSignature, callees []calleeDigestPair) (string, error) {
	preimage, err := closureDigestPreimageBytes(signature, callees)
	if err != nil {
		return "", err
	}
	return digest(preimage), nil
}

// BuildInterface strips every function body from program and binds the
// resulting summary to program's own content digest, reusing the same
// SHA-256 content-digest pattern already proven for evidence.CoreDigest
// (Spike 004's certificate binds the same way). Callers MUST run
// ValidatePublished against program first: BuildInterface packages what the
// producer already proved rather than re-deriving it.
//
// D-07-08: this emits Schema core.InterfaceSchema1 (lang.interface/1) —
// every function.ID+":type:0" abilities is unchanged from /0.
// Every /1 field is populated from its R-01 authority (see the field-level
// doc comments on core.FunctionSignature); Callable is left at its
// fail-closed zero value false (its predicate is 07-02's) and ClosureDigest
// is left empty (07-01 Task 3 fills it).
func BuildInterface(program core.Program) (core.Interface, error) {
	coreBytes, err := json.Marshal(program)
	if err != nil {
		return core.Interface{}, err
	}
	summary := core.Interface{
		Schema: core.InterfaceSchema1, ModuleID: program.ModuleID, CoreDigest: digest(coreBytes),
		Functions: make([]core.FunctionSignature, 0, len(program.Functions)),
	}
	for _, function := range program.Functions {
		abilities := []core.Ability{}
		hasDropAbility := false
		if function.Linear != nil {
			for _, fact := range function.Linear.Types {
				if fact.ID == function.ID+":type:0" {
					abilities = fact.Abilities
					break
				}
			}
		}
		for _, ability := range abilities {
			if ability == core.AbilityDrop {
				hasDropAbility = true
				break
			}
		}

		parameterContract := core.ParameterContract{
			ID: function.Parameter.ID, Name: function.Parameter.Name, Type: function.Parameter.Type,
			// D-07-01: today's grammar has exactly one parameter form
			// (by-value), so every parameter's Mode is "owned" (R-01).
			Mode:  "owned",
			Drops: hasDropAbility && !parameterEscapesOwned(function),
		}

		returnContract := core.ReturnContract{Type: function.ReturnType}
		switch {
		case function.PublicOrigin == nil:
			returnContract.Mode = "owned"
			returnContract.Paths = []string{}
		case function.PublicOrigin.Access == "shared":
			returnContract.Mode = "shared"
			returnContract.Paths = function.PublicOrigin.Paths
		default:
			returnContract.Mode = function.PublicOrigin.Access
			returnContract.Paths = function.PublicOrigin.Paths
		}
		returnContract.Fresh = returnContract.Mode == "owned" && hasDropAbility

		foreignReach := core.ForeignReach{}
		fails := ""
		if function.ForeignContract != nil {
			foreignReach = core.ForeignReach{
				Allocator: function.ForeignContract.Allocator, Unwind: function.ForeignContract.Unwind,
				NonlocalExit: function.ForeignContract.NonlocalExit,
			}
			fails = function.ForeignContract.Fails
		}

		signature := core.FunctionSignature{
			ID: function.ID, Name: function.Name,
			Parameters: []core.ParameterContract{parameterContract},
			Return:     returnContract,
			Abilities:  abilities,
			Callable:   false,
			Fails:      fails,
			Foreign:    foreignReach,
		}
		// D-07-38: only the zero-callee base case is computed in this plan
		// (nil callees) — the chaining arm over real callees lands in
		// 07-08-PLAN.md, after cycle refusal exists.
		closureDigest, err := computeClosureDigest(signature, nil)
		if err != nil {
			return core.Interface{}, err
		}
		signature.ClosureDigest = closureDigest
		summary.Functions = append(summary.Functions, signature)
	}
	return summary, nil
}

// parameterEscapesOwned reports whether function's return traces back to its
// own parameter as a directly moved/copied OWNED value, following only
// OpMove/OpCopy chains and never crossing an OpBorrowShared/
// OpBorrowExclusive/OpForeignCall operation. This is R-01's authority for
// ParameterContract.Drops: a borrow-derived return never transfers the
// parameter's ownership, so the parameter's drop obligation (if it has the
// Drop ability at all) still belongs to, and is discharged by, the callee.
// When the parameter itself IS the owned return value, ownership (and the
// obligation to drop it) moves to the caller instead -- the callee does not
// discharge it.
func parameterEscapesOwned(function core.Function) bool {
	if function.Linear == nil {
		return false
	}
	operations := function.Linear.Operations
	sourceOf := make(map[string]core.LinearOperation, len(operations))
	var returnOps []core.LinearOperation
	for _, operation := range operations {
		if operation.Kind == core.OpReturn {
			returnOps = append(returnOps, operation)
			continue
		}
		sourceOf[operation.TargetID] = operation
	}
	for _, returnOp := range returnOps {
		current := returnOp.SourceID
		visited := make(map[string]bool)
		for current != function.Parameter.ID {
			if visited[current] {
				break
			}
			visited[current] = true
			operation, exists := sourceOf[current]
			if !exists || (operation.Kind != core.OpMove && operation.Kind != core.OpCopy) {
				break
			}
			current = operation.SourceID
		}
		if current == function.Parameter.ID {
			return true
		}
	}
	return false
}

// FunctionAnswer is what a body-blind consumer can decide for one function:
// its declared origin and access mode, read directly from the summary.
type FunctionAnswer struct {
	ID     string
	Name   string
	Paths  []string
	Access string
}

// decodeErrorCode extracts a core.DecodeError's Code, so CheckSummary can
// re-surface DecodeInterface's own refusal code through originvalidate's
// {Code}-only Error shape rather than collapsing every structural refusal
// into one generic code.
func decodeErrorCode(err error) string {
	if decodeErr, ok := err.(*core.DecodeError); ok {
		return decodeErr.Code
	}
	return ""
}

// CheckSummary is the body-blind consuming path (OWN-04 success criterion
// 4's second, independent CLI invocation), rerouted through
// core.DecodeInterface per D-07-36: a document is now schema-peeked and
// strictly validated (presence, non-emptiness, Mode's closed set, digest
// shape, and function-ID uniqueness) before this function ever asks an
// origin or access question of it, and a lang.interface/0 document is
// refused outright (T-07-02: never admissible for a call).
//
// coreBytes is the RAW, UNPARSED bytes of the core artifact the summary
// claims to be bound to: this function hashes them and compares against the
// summary's recorded digest, and NEVER unmarshals coreBytes into any struct
// that could carry a Linear/Match body field — the digest check is the only
// use coreBytes is put to, and it runs before any origin or access question
// is answered (T-03-03: a stale summary is rejected before it is partially
// trusted).
func CheckSummary(summaryBytes, coreBytes []byte) ([]FunctionAnswer, error) {
	decoded, err := core.DecodeInterface(summaryBytes)
	if err != nil {
		if code := decodeErrorCode(err); code != "" {
			return nil, &Error{Code: code}
		}
		return nil, &Error{Code: "origin.invalid_summary"}
	}
	if !decoded.Admissible || decoded.V1 == nil {
		return nil, &Error{Code: "origin.summary_not_admissible"}
	}
	summary := *decoded.V1
	if summary.CoreDigest != digest(coreBytes) {
		return nil, &Error{Code: "origin.stale_summary"}
	}
	answers := make([]FunctionAnswer, 0, len(summary.Functions))
	for _, function := range summary.Functions {
		answer := FunctionAnswer{ID: function.ID, Name: function.Name}
		if function.Return.Mode != "owned" {
			answer.Paths = function.Return.Paths
			answer.Access = function.Return.Mode
		}
		answers = append(answers, answer)
	}
	return answers, nil
}
