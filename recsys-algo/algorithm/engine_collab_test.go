package algorithm

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestCollaborativeScoresSingleAssignment(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "candidate_sources.go", nil, 0)
	if err != nil {
		t.Fatalf("parse engine.go: %v", err)
	}

	var fn *ast.FuncDecl
	for _, decl := range file.Decls {
		if declFn, ok := decl.(*ast.FuncDecl); ok && declFn.Name.Name == "getCollaborativeCandidates" {
			fn = declFn
			break
		}
	}
	if fn == nil {
		t.Fatal("getCollaborativeCandidates not found")
	}

	assignments := 0
	ast.Inspect(fn, func(node ast.Node) bool {
		assign, ok := node.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for _, lhs := range assign.Lhs {
			index, ok := lhs.(*ast.IndexExpr)
			if !ok {
				continue
			}
			if ident, ok := index.X.(*ast.Ident); ok && ident.Name == "scores" {
				assignments++
			}
		}
		return true
	})

	if assignments != 1 {
		t.Fatalf("expected 1 scores[...] assignment in getCollaborativeCandidates, got %d", assignments)
	}
}
