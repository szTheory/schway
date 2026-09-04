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

	"github.com/codename-lang/lang/internal/compiler/core"
)

// KnownEscape names the exact boundary this package cannot prove: a producer
// whose frontend and published summary lie in a coordinated way remains
// outside source-blind, body-blind validation. It mirrors
// corevalidate.KnownEscape's shape, applied to the origin trust boundary
// instead of the source/core one (T-03-08, an accepted residual, never
// claimed as solved).
const KnownEscape = "escape:coordinated-frontend-summary-lie"

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
		if operation.Kind == core.OpReturn {
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
func ValidatePublished(program core.Program) []Problem {
	for _, function := range program.Functions {
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

// BuildInterface strips every function body from program and binds the
// resulting summary to program's own content digest, reusing the same
// SHA-256 content-digest pattern already proven for evidence.CoreDigest
// (Spike 004's certificate binds the same way). Callers MUST run
// ValidatePublished against program first: BuildInterface packages what the
// producer already proved rather than re-deriving it.
func BuildInterface(program core.Program) (core.Interface, error) {
	coreBytes, err := json.Marshal(program)
	if err != nil {
		return core.Interface{}, err
	}
	summary := core.Interface{
		Schema: core.InterfaceSchema, ModuleID: program.ModuleID, CoreDigest: digest(coreBytes),
		Functions: make([]core.FunctionSignature, 0, len(program.Functions)),
	}
	for _, function := range program.Functions {
		abilities := []core.Ability{}
		if function.Linear != nil {
			for _, fact := range function.Linear.Types {
				if fact.ID == function.ID+":type:0" {
					abilities = fact.Abilities
					break
				}
			}
		}
		summary.Functions = append(summary.Functions, core.FunctionSignature{
			ID: function.ID, Name: function.Name, Parameter: function.Parameter, ReturnType: function.ReturnType,
			PublicOrigin: function.PublicOrigin, Abilities: abilities,
		})
	}
	return summary, nil
}

// FunctionAnswer is what a body-blind consumer can decide for one function:
// its declared origin and access mode, read directly from the summary.
type FunctionAnswer struct {
	ID     string
	Name   string
	Paths  []string
	Access string
}

// CheckSummary is the body-blind consuming path (OWN-04 success criterion
// 4's second, independent CLI invocation). coreBytes is the RAW, UNPARSED
// bytes of the core artifact the summary claims to be bound to: this
// function hashes them and compares against the summary's recorded digest,
// and NEVER unmarshals coreBytes into any struct that could carry a
// Linear/Match body field — the digest check is the only use coreBytes is
// put to, and it runs before any origin or access question is answered
// (T-03-03: a stale summary is rejected before it is partially trusted).
func CheckSummary(summaryBytes, coreBytes []byte) ([]FunctionAnswer, error) {
	var summary core.Interface
	if err := json.Unmarshal(summaryBytes, &summary); err != nil {
		return nil, &Error{Code: "origin.invalid_summary"}
	}
	if summary.CoreDigest != digest(coreBytes) {
		return nil, &Error{Code: "origin.stale_summary"}
	}
	answers := make([]FunctionAnswer, 0, len(summary.Functions))
	for _, function := range summary.Functions {
		answer := FunctionAnswer{ID: function.ID, Name: function.Name}
		if function.PublicOrigin != nil {
			answer.Paths = function.PublicOrigin.Paths
			answer.Access = function.PublicOrigin.Access
		}
		answers = append(answers, answer)
	}
	return answers, nil
}
