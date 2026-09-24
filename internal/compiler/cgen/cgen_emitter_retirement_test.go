package cgen_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

func TestSupersededEmitterDefinitionsRemoved(t *testing.T) {
	packages, err := parser.ParseDir(token.NewFileSet(), testsupport.ProjectPath("internal", "compiler", "cgen"), func(info os.FileInfo) bool {
		return strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go")
	}, parser.AllErrors)
	if err != nil {
		t.Fatal(err)
	}
	legacy := map[string]bool{"emitMatch": true, "emitBranch": true, "emitLinear": true}
	for _, pkg := range packages {
		for _, file := range pkg.Files {
			for _, declaration := range file.Decls {
				function, ok := declaration.(*ast.FuncDecl)
				if ok && legacy[function.Name.Name] {
					t.Errorf("superseded emitter definition %s remains", function.Name.Name)
				}
			}
		}
	}
}
