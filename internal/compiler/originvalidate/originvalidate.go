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

// RecomputeOrigin derives the origin path(s) and access mode a function's
// body actually returns, by tracing the OpReturn operation's SourceID
// backward through the function's flat Operations list (via a
// TargetID->SourceID chain built from core.LinearOperation facts alone)
// until it reaches the function's own parameter place or the chain breaks.
// This is a materially different mechanism from check.go's derivation, which
// builds the PublicOrigin fact forward from the declaration while lowering
// the body (D-12): this walks backward from the returned place using only
// core facts, and never consults function.PublicOrigin itself.
func RecomputeOrigin(function core.Function) (paths []string, access string, ok bool) {
	if function.Linear == nil || len(function.Linear.Operations) == 0 {
		return nil, "", false
	}
	operations := function.Linear.Operations
	sourceOf := make(map[string]core.LinearOperation, len(operations))
	var returnOp *core.LinearOperation
	for index := range operations {
		operation := operations[index]
		if operation.Kind == core.OpReturn {
			if returnOp == nil {
				returnOp = &operations[index]
			}
			continue
		}
		sourceOf[operation.TargetID] = operation
	}
	if returnOp == nil {
		return nil, "", false
	}
	current := returnOp.SourceID
	visited := make(map[string]bool)
	derivedAccess := ""
	for current != function.Parameter.ID {
		if visited[current] {
			return nil, "", false
		}
		visited[current] = true
		operation, exists := sourceOf[current]
		if !exists {
			return nil, "", false
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
		return nil, "", false
	}
	return []string{function.Parameter.Name}, derivedAccess, true
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
