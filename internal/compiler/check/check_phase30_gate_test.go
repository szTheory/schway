package check

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

var gateForbiddenContractNames = map[string]bool{"Body": true, "Linear": true, "Match": true, "Funcs": true}

// gateContractViolations returns every forbidden name used as a field name or
// as an identifier or selector in a field type of the calleeContract struct
// declared in src.
func gateContractViolations(t *testing.T, filename string, src any) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), filename, src, 0)
	if err != nil {
		t.Fatal(err)
	}
	var found []string
	ast.Inspect(file, func(node ast.Node) bool {
		spec, ok := node.(*ast.TypeSpec)
		if !ok || spec.Name.Name != "calleeContract" {
			return true
		}
		structType, ok := spec.Type.(*ast.StructType)
		if !ok {
			t.Fatalf("calleeContract is not a struct")
		}
		for _, field := range structType.Fields.List {
			for _, name := range field.Names {
				if gateForbiddenContractNames[name.Name] {
					found = append(found, name.Name)
				}
			}
			ast.Inspect(field.Type, func(inner ast.Node) bool {
				switch typed := inner.(type) {
				case *ast.Ident:
					if gateForbiddenContractNames[typed.Name] {
						found = append(found, typed.Name)
					}
				case *ast.SelectorExpr:
					if gateForbiddenContractNames[typed.Sel.Name] {
						found = append(found, typed.Sel.Name)
					}
				}
				return true
			})
		}
		return false
	})
	return found
}

// TestPhase30GateBodyBlindContractShape is flag (e) of the D-14 gate.
// Observed calleeContract fields: ID, ParameterType, ReturnType,
// ReturnsBorrowOfParam. None names a body, linear body, match or function list.
func TestPhase30GateBodyBlindContractShape(t *testing.T) {
	if got := gateContractViolations(t, "check.go", nil); len(got) != 0 {
		t.Fatalf("calleeContract carries body-bearing names %v", got)
	}

	t.Run("negative-control", func(t *testing.T) {
		synthetic := "package check\ntype calleeContract struct { ID string; Body ast.Body }\n"
		got := gateContractViolations(t, "synthetic.go", synthetic)
		if len(got) == 0 || got[0] != "Body" {
			t.Fatalf("the scanner must flag Body, got %v", got)
		}
	})

	t.Run("existing guard is tied to the gate", func(t *testing.T) {
		file, err := parser.ParseFile(token.NewFileSet(), "check_test.go", nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			if fn, ok := declaration.(*ast.FuncDecl); ok && fn.Recv == nil && fn.Name.Name == "TestCallAdmissionBodyBlindControl" {
				return
			}
		}
		t.Fatal("TestCallAdmissionBodyBlindControl is missing from check_test.go")
	})
}
