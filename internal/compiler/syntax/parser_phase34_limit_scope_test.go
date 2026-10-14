package syntax

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/ast"
)

func TestPhase34LinearBindingLimitScope(t *testing.T) {
	if maxLinearBindings != 65536 {
		t.Fatalf("linear-binding cap = %d, want independently pinned 65536", maxLinearBindings)
	}
	// The token/source limits precede this AST guard. Exercise the guard itself
	// so its aggregate scope remains tested even when source cannot reach it.
	left := &ast.LinearBody{Bindings: make([]ast.Binding, 32768)}
	right := &ast.LinearBody{Bindings: make([]ast.Binding, 32768)}
	program := ast.Program{Funcs: []ast.FuncDecl{{Body: ast.Body{Linear: left}}, {Body: ast.Body{Linear: right}}}}
	if got := programLimit(program); got != "" {
		t.Fatalf("aggregate exact cap refused: %q", got)
	}
	right.Bindings = append(right.Bindings, ast.Binding{})
	if got := programLimit(program); got != "body facts" {
		t.Fatalf("aggregate cap+1 across individually smaller bodies = %q, want body facts", got)
	}
	for _, fn := range program.Funcs {
		if got := programLimit(ast.Program{Funcs: []ast.FuncDecl{fn}}); got != "" {
			t.Fatalf("individual body below cap refused: %q", got)
		}
	}
	right.Bindings = right.Bindings[:32768]
	if got := programLimit(program); got != "" {
		t.Fatalf("exact cap did not restore: %q", got)
	}
}
