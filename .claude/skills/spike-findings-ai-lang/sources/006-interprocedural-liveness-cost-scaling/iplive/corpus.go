package iplive

import "fmt"

// The corpus is built from four body templates. Every template is a real
// discriminator: its summary bits change some caller's admission decision, so
// a wrong summary cannot hide inside an inert body.
//
//	leafUse   -- reads its parameter, returns owned.   {UsesParam:true,  Returns:false}
//	leafPass  -- touches nothing, hands the parameter back. {false, true}
//	relay     -- forwards its parameter to its callees and hands the first
//	             result back; inherits its callees' bits.
//	probe     -- borrows a local owner, calls, then moves the owner. The move
//	             is legal exactly when the callee's summary says the loan is
//	             already dead.
func leafUse(id string) Function {
	return Function{
		ID: id, ParamLoan: "p",
		Blocks: []Block{{ID: "entry", Ops: []Operation{{ID: id + ":0", Kind: OpUse, Loan: "p"}}}},
	}
}

func leafPass(id string) Function {
	return Function{
		ID: id, ParamLoan: "p",
		Blocks:  []Block{{ID: "entry", Ops: []Operation{}}},
		Returns: "p",
	}
}

// relay forwards the parameter into every callee and returns the first call's
// result. probeKind selects which caller-side conflict pattern is appended:
//
//	"after"  -- borrow, call, MOVE, use(result): conflicts iff the callee
//	            returns a borrow of its parameter.
//	"before" -- borrow, MOVE, call: conflicts iff the callee uses its parameter.
//	""       -- no probe.
func relay(id string, probeKind string, callees ...string) Function {
	fn := Function{ID: id, ParamLoan: "p"}
	ops := []Operation{}
	for i, callee := range callees {
		result := fmt.Sprintf("r%d", i)
		ops = append(ops, Operation{
			ID: fmt.Sprintf("%s:call%d", id, i), Kind: OpCall,
			Loan: "p", Callee: callee, Result: result,
		})
	}
	if len(callees) > 0 {
		fn.Returns = "r0"
	}
	switch probeKind {
	case "after":
		ops = append(ops,
			Operation{ID: id + ":borrow", Kind: OpBorrow, Loan: "l", Place: "own"},
			Operation{ID: id + ":probe", Kind: OpCall, Loan: "l", Callee: callees[0], Result: "pr"},
			Operation{ID: id + ":move", Kind: OpMove, Place: "own"},
			Operation{ID: id + ":useres", Kind: OpUse, Loan: "pr"},
		)
	case "before":
		ops = append(ops,
			Operation{ID: id + ":borrow", Kind: OpBorrow, Loan: "l", Place: "own"},
			Operation{ID: id + ":move", Kind: OpMove, Place: "own"},
			Operation{ID: id + ":probe", Kind: OpCall, Loan: "l", Callee: callees[0], Result: "pr"},
		)
	}
	fn.Blocks = []Block{{ID: "entry", Ops: ops}}
	return fn
}

// branchRelay is the one branchy template: the loan is borrowed before a
// fan-out, moved on ONE arm, and the call result is used after the join. The
// move is legal exactly when the callee does NOT hand the parameter back --
// an edge-specific endpoint decided by an interprocedural fact.
func branchRelay(id string, callee string) Function {
	return Function{
		ID: id, ParamLoan: "p", Returns: "r0",
		Blocks: []Block{
			{ID: "entry", Ops: []Operation{
				{ID: id + ":call0", Kind: OpCall, Loan: "p", Callee: callee, Result: "r0"},
				{ID: id + ":borrow", Kind: OpBorrow, Loan: "l", Place: "own"},
				{ID: id + ":probe", Kind: OpCall, Loan: "l", Callee: callee, Result: "pr"},
			}, Successors: []string{"then", "else"}},
			{ID: "then", Ops: []Operation{
				{ID: id + ":then:move", Kind: OpMove, Place: "own"},
			}, Successors: []string{"join"}},
			{ID: "else", Ops: []Operation{}, Successors: []string{"join"}},
			{ID: "join", Ops: []Operation{
				{ID: id + ":join:use", Kind: OpUse, Loan: "pr"},
			}},
		},
	}
}

// Shape names the generated call-graph topologies.
type Shape string

const (
	ShapeChain   Shape = "chain"   // f0 -> f1 -> ... -> leaf; no sharing
	ShapeTree    Shape = "tree"    // binary tree; no sharing
	ShapeDiamond Shape = "diamond" // layered, every node shares two callees
	ShapeDense   Shape = "dense"   // layered, fan-out 3, heavy sharing
	ShapeParser  Shape = "parser"  // seeded mixed DAG, depth ~8, fan-out 2-4
	ShapeForward Shape = "forward" // ONE caller forwarding a result through k calls
)

// Generate builds a program of the given shape with approximately n functions.
func Generate(shape Shape, n int) (*Program, error) {
	switch shape {
	case ShapeChain:
		return chain(n), nil
	case ShapeTree:
		return tree(n), nil
	case ShapeDiamond:
		return layered(n, 2), nil
	case ShapeDense:
		return layered(n, 3), nil
	case ShapeParser:
		return parser(n), nil
	case ShapeForward:
		return forward(n), nil
	}
	return nil, fmt.Errorf("unknown shape %q", shape)
}

func probeFor(index int) string {
	switch index % 3 {
	case 0:
		return "after"
	case 1:
		return "before"
	}
	return ""
}

func chain(n int) *Program {
	p := &Program{Roots: []string{"f0"}}
	for i := 0; i < n-1; i++ {
		p.Functions = append(p.Functions, relay(fmt.Sprintf("f%d", i), probeFor(i), fmt.Sprintf("f%d", i+1)))
	}
	p.Functions = append(p.Functions, leafPass(fmt.Sprintf("f%d", n-1)))
	return p
}

func tree(n int) *Program {
	p := &Program{Roots: []string{"f0"}}
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("f%d", i)
		left, right := 2*i+1, 2*i+2
		switch {
		case left >= n:
			if i%2 == 0 {
				p.Functions = append(p.Functions, leafPass(id))
			} else {
				p.Functions = append(p.Functions, leafUse(id))
			}
		case right >= n:
			p.Functions = append(p.Functions, relay(id, probeFor(i), fmt.Sprintf("f%d", left)))
		default:
			p.Functions = append(p.Functions, relay(id, probeFor(i), fmt.Sprintf("f%d", left), fmt.Sprintf("f%d", right)))
		}
	}
	return p
}

// layered builds the sharing-heavy shapes: width-`width` layers where every
// node calls `width` nodes of the next layer. The number of distinct
// call-graph PATHS is width^layers while the number of functions is only
// width*layers -- the exact gap a memoized cache collapses and an unmemoized
// recomputation pays in full.
func layered(n, width int) *Program {
	layers := n / width
	if layers < 2 {
		layers = 2
	}
	p := &Program{Roots: []string{"f0"}}
	id := func(layer, index int) string { return fmt.Sprintf("l%dn%d", layer, index) }

	root := relay("f0", "", id(0, 0))
	p.Functions = append(p.Functions, root)
	counter := 0
	for layer := 0; layer < layers; layer++ {
		for index := 0; index < width; index++ {
			counter++
			if layer == layers-1 {
				if index%2 == 0 {
					p.Functions = append(p.Functions, leafPass(id(layer, index)))
				} else {
					p.Functions = append(p.Functions, leafUse(id(layer, index)))
				}
				continue
			}
			callees := make([]string, 0, width)
			for offset := 0; offset < width; offset++ {
				callees = append(callees, id(layer+1, (index+offset)%width))
			}
			p.Functions = append(p.Functions, relay(id(layer, index), probeFor(counter), callees...))
		}
	}
	return p
}

// parser is the "realistic fan-out" shape EFF-02 asks about: a seeded DAG with
// depth ~8 and fan-out 2-4, a tenth of it leaves, and one branchy body per
// layer so the CFG side is not trivially straight-line.
func parser(n int) *Program {
	p := &Program{Roots: []string{"f0"}}
	seed := uint64(0x5EED_5006)
	next := func(mod int) int {
		seed = seed*6364136223846793005 + 1442695040888963407
		return int((seed >> 33) % uint64(mod))
	}
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("f%d", i)
		remaining := n - i - 1
		if remaining == 0 || (i > n/3 && next(10) == 0) {
			if i%2 == 0 {
				p.Functions = append(p.Functions, leafPass(id))
			} else {
				p.Functions = append(p.Functions, leafUse(id))
			}
			continue
		}
		window := remaining
		if window > 6 {
			window = 6
		}
		fanout := 2 + next(3)
		seen := map[string]bool{}
		var callees []string
		for k := 0; k < fanout; k++ {
			target := fmt.Sprintf("f%d", i+1+next(window))
			if seen[target] {
				continue
			}
			seen[target] = true
			callees = append(callees, target)
		}
		if i%7 == 3 {
			p.Functions = append(p.Functions, branchRelay(id, callees[0]))
			continue
		}
		p.Functions = append(p.Functions, relay(id, probeFor(i), callees...))
	}
	return p
}

// forward is the shape that stresses the DERIVATION rather than the call
// graph: one caller whose body threads a single borrow through k successive
// calls, each consuming the previous call's result. The call graph is a star
// of depth 1 -- every cost it shows belongs to deriving one summary from one
// body.
func forward(n int) *Program {
	callees := n - 1
	if callees < 1 {
		callees = 1
	}
	p := &Program{Roots: []string{"f0"}}
	caller := Function{ID: "f0", ParamLoan: "p", Returns: fmt.Sprintf("r%d", callees-1)}
	ops := []Operation{}
	previous := "p"
	for i := 0; i < callees; i++ {
		callee := fmt.Sprintf("g%d", i)
		result := fmt.Sprintf("r%d", i)
		ops = append(ops, Operation{
			ID: fmt.Sprintf("f0:call%d", i), Kind: OpCall,
			Loan: previous, Callee: callee, Result: result,
		})
		previous = result
		p.Functions = append(p.Functions, leafPass(callee))
	}
	// Keep the forwarded loan observable: borrow, forward it once more, move,
	// then use the result.
	ops = append(ops,
		Operation{ID: "f0:borrow", Kind: OpBorrow, Loan: "l", Place: "own"},
		Operation{ID: "f0:probe", Kind: OpCall, Loan: "l", Callee: "g0", Result: "pr"},
		Operation{ID: "f0:move", Kind: OpMove, Place: "own"},
		Operation{ID: "f0:useres", Kind: OpUse, Loan: "pr"},
	)
	caller.Blocks = []Block{{ID: "entry", Ops: ops}}
	p.Functions = append(p.Functions, caller)
	return p
}

// ReversedForward lists the same forwarding chain in REVERSE program order --
// each call appears before the call that produces its argument. Real
// core.LinearOperation sequences never look like this; it exists to price what
// the program-order assumption is worth.
func ReversedForward(k int) *Program {
	p := &Program{Roots: []string{"main"}}
	caller := Function{ID: "f0", ParamLoan: "p", Returns: fmt.Sprintf("r%d", k-1)}
	var ops []Operation
	for i := k - 1; i >= 0; i-- {
		previous := "p"
		if i > 0 {
			previous = fmt.Sprintf("r%d", i-1)
		}
		ops = append(ops, Operation{
			ID: fmt.Sprintf("f0:call%d", i), Kind: OpCall,
			Loan: previous, Callee: fmt.Sprintf("g%d", i), Result: fmt.Sprintf("r%d", i),
		})
		p.Functions = append(p.Functions, leafPass(fmt.Sprintf("g%d", i)))
	}
	caller.Blocks = []Block{{ID: "entry", Ops: ops}}
	p.Functions = append(p.Functions, caller)
	// f0 must itself be a callee, or nothing ever derives its summary.
	p.Functions = append(p.Functions, relay("main", "after", "f0"))
	return p
}

// Fixtures are the hand-written programs whose expected refusals are stated by
// hand rather than by either mechanism.
func Fixtures() map[string]*Program {
	return map[string]*Program{
		// The returned borrow keeps the loan alive past the move.
		"returns_borrow_conflicts": {
			Roots:     []string{"caller"},
			Functions: []Function{relay("caller", "after", "callee"), leafPass("callee")},
		},
		// An owned return does not: the same caller body is legal.
		"owned_return_legal": {
			Roots:     []string{"caller"},
			Functions: []Function{relay("caller", "after", "callee"), leafUse("callee")},
		},
		// The callee reads through the loan, so a move BEFORE the call is illegal.
		"uses_param_conflicts": {
			Roots:     []string{"caller"},
			Functions: []Function{relay("caller", "before", "callee"), leafUse("callee")},
		},
		// A callee that touches nothing makes that same move legal.
		"unused_param_legal": {
			Roots:     []string{"caller"},
			Functions: []Function{relay("caller", "before", "callee"), leafPass("callee")},
		},
		// Transitivity: the relayed bit has to survive three hops.
		"transitive_relay_conflicts": {
			Roots: []string{"caller"},
			Functions: []Function{
				relay("caller", "after", "mid1"),
				relay("mid1", "", "mid2"),
				relay("mid2", "", "leaf"),
				leafPass("leaf"),
			},
		},
		// The move is on one arm only, and the returned borrow is used after
		// the join: an edge-specific endpoint decided interprocedurally.
		"branch_arm_conflicts": {
			Roots:     []string{"caller"},
			Functions: []Function{branchRelay("caller", "callee"), leafPass("callee")},
		},
	}
}
