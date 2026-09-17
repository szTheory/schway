package iplive

import (
	"errors"
	"fmt"
	"sort"
)

// Summary is the whole body-blind fact a caller is allowed to consult about a
// callee (OWN-06: "from callee signatures only -- never by re-walking callee
// bodies"). Two bits are enough to make the caller's move-while-borrowed
// decision summary-sensitive in both directions:
//
//   - UsesParam        -- the call itself reads through the argument loan, so
//     the loan is live AT the call site.
//   - ReturnsBorrowOfParam -- the returned value aliases the argument loan, so
//     the loan stays live for as long as the RESULT is,
//     which can be arbitrarily far past the call.
//
// Both bits are transitive: a relay that forwards its parameter inherits the
// bits of whatever it forwards to. That transitivity is the entire cost
// question -- it is what makes a summary depend on the whole reachable
// subgraph rather than on one body.
type Summary struct {
	UsesParam            bool `json:"uses_param"`
	ReturnsBorrowOfParam bool `json:"returns_borrow_of_param"`
}

// ErrBudget is returned when a provider exceeds its work budget. It is a
// measurement outcome, not a failure: the recompute mechanism is expected to
// hit it on shapes where its cost is super-polynomial.
var ErrBudget = errors.New("summary work budget exceeded")

// Provider is the caller-side interface to callee summaries. The two
// implementations differ ONLY in whether a derivation is remembered; the
// derivation law itself is shared, so the measured difference is attributable
// to memoization and nothing else.
type Provider interface {
	Summary(fnID string) (Summary, error)
	Work() int
	Name() string
	// Derivations reports how many times a full body-level derivation ran.
	Derivations() int
}

type providerBase struct {
	byID        map[string]*Function
	work        int
	derivations int
	budget      int
	// dropReturnsBorrow is the injected summary defect (fault seam 1): the
	// derivation forgets that a forwarded parameter comes back out through
	// the return, which under-approximates caller liveness.
	dropReturnsBorrow bool
}

func (b *providerBase) Work() int        { return b.work }
func (b *providerBase) Derivations() int { return b.derivations }

func (b *providerBase) charge(units int) error {
	b.work += units
	if b.budget > 0 && b.work > b.budget {
		return ErrBudget
	}
	return nil
}

// derive computes one function's summary from its own body, consulting the
// provider for every callee it reaches. It is a small local fixpoint because a
// call result can itself be forwarded into a later call.
func derive(base *providerBase, p Provider, fnID string) (Summary, error) {
	fn, ok := base.byID[fnID]
	if !ok {
		return Summary{}, fmt.Errorf("unresolved callee %q", fnID)
	}
	base.derivations++

	derived := map[string]bool{fn.ParamLoan: true}
	summary := Summary{}
	for {
		changed := false
		for _, block := range fn.Blocks {
			for _, op := range block.Ops {
				if err := base.charge(1); err != nil {
					return Summary{}, err
				}
				switch op.Kind {
				case OpUse:
					if derived[op.Loan] && !summary.UsesParam {
						summary.UsesParam, changed = true, true
					}
				case OpCall:
					if !derived[op.Loan] {
						continue
					}
					calleeSummary, err := p.Summary(op.Callee)
					if err != nil {
						return Summary{}, err
					}
					if calleeSummary.UsesParam && !summary.UsesParam {
						summary.UsesParam, changed = true, true
					}
					if calleeSummary.ReturnsBorrowOfParam && op.Result != "" && !derived[op.Result] {
						derived[op.Result], changed = true, true
					}
				}
			}
		}
		if !changed {
			break
		}
	}
	if fn.Returns != "" && derived[fn.Returns] && !base.dropReturnsBorrow {
		summary.ReturnsBorrowOfParam = true
	}
	return summary, nil
}

// MemoProvider derives each function's summary at most once and remembers it
// -- the "memoized summary cache designed in from the start" arm.
type MemoProvider struct {
	providerBase
	cache    map[string]Summary
	inFlight map[string]bool
	// staleKeyed is the injected cache defect (fault seam 2): the cache is
	// keyed by function ID alone, so an edit to a callee never invalidates a
	// caller's remembered summary.
	staleKeyed bool
}

// NewMemoProvider builds the memoized arm. budget <= 0 means unbounded.
func NewMemoProvider(p *Program, budget int) *MemoProvider {
	return &MemoProvider{
		providerBase: providerBase{byID: p.index(), budget: budget},
		cache:        map[string]Summary{},
		inFlight:     map[string]bool{},
	}
}

func (m *MemoProvider) Name() string { return "memoized" }

func (m *MemoProvider) Summary(fnID string) (Summary, error) {
	if err := m.charge(1); err != nil {
		return Summary{}, err
	}
	if cached, ok := m.cache[fnID]; ok {
		return cached, nil
	}
	if m.inFlight[fnID] {
		return Summary{}, fmt.Errorf("call_graph_cycle: %q reaches itself", fnID)
	}
	m.inFlight[fnID] = true
	summary, err := derive(&m.providerBase, m, fnID)
	delete(m.inFlight, fnID)
	if err != nil {
		return Summary{}, err
	}
	m.cache[fnID] = summary
	return summary, nil
}

// Invalidate models a callee edit. Without staleKeyed it evicts the edited
// function and every transitive caller of it (what a ClosureDigest chain buys);
// with staleKeyed it evicts only the edited function itself, leaving callers
// serving a stale-but-self-consistent verdict.
func (m *MemoProvider) Invalidate(p *Program, fnID string) int {
	if m.staleKeyed {
		delete(m.cache, fnID)
		return 1
	}
	callers := map[string][]string{}
	for i := range p.Functions {
		for _, callee := range p.Functions[i].callees() {
			callers[callee] = append(callers[callee], p.Functions[i].ID)
		}
	}
	evicted := map[string]bool{}
	queue := []string{fnID}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if evicted[id] {
			continue
		}
		evicted[id] = true
		delete(m.cache, id)
		queue = append(queue, callers[id]...)
	}
	return len(evicted)
}

// CachedIDs reports which summaries are currently remembered, sorted.
func (m *MemoProvider) CachedIDs() []string {
	ids := make([]string, 0, len(m.cache))
	for id := range m.cache {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// RecomputeProvider re-derives a callee's summary at every consultation -- the
// "extend loanLivenessFixpoint directly, no cache" arm. It is the arm the gate
// exists to price.
type RecomputeProvider struct {
	providerBase
	depth int
	// scratch, when enabled, remembers a derivation only for the duration of
	// one top-level consultation. It is the CHARITABLE unmemoized arm: no
	// summary survives across call sites, but a single caller does not
	// re-derive the same callee twice while answering one question. Its cost
	// is therefore the number of distinct call-graph PATHS rather than the
	// number of repeated consultations, which separates "exponential because
	// of sharing" from "exponential because of repetition".
	scratch    map[string]Summary
	useScratch bool
}

// EnableScratch turns the unmemoized arm into its charitable variant.
func (r *RecomputeProvider) EnableScratch() { r.useScratch = true }

// beginFunction is Analyze's hook: the scratch cache lives exactly as long as
// one caller's analysis, so nothing is remembered ACROSS callers.
func (r *RecomputeProvider) beginFunction() { r.scratch = map[string]Summary{} }

// NewRecomputeProvider builds the unmemoized arm. budget <= 0 means unbounded,
// which on a sharing-heavy call graph means "does not terminate in useful time".
func NewRecomputeProvider(p *Program, budget int) *RecomputeProvider {
	return &RecomputeProvider{providerBase: providerBase{byID: p.index(), budget: budget}}
}

func (r *RecomputeProvider) Name() string {
	if r.useScratch {
		return "recompute+scratch"
	}
	return "recompute"
}

func (r *RecomputeProvider) Summary(fnID string) (Summary, error) {
	if err := r.charge(1); err != nil {
		return Summary{}, err
	}
	if r.depth > len(r.byID)+1 {
		return Summary{}, fmt.Errorf("call_graph_cycle: %q reaches itself", fnID)
	}
	if r.useScratch {
		if r.scratch == nil {
			r.scratch = map[string]Summary{}
		}
		if cached, ok := r.scratch[fnID]; ok {
			return cached, nil
		}
	}
	r.depth++
	summary, err := derive(&r.providerBase, r, fnID)
	r.depth--
	if err != nil {
		return Summary{}, err
	}
	if r.useScratch {
		r.scratch[fnID] = summary
	}
	return summary, nil
}
