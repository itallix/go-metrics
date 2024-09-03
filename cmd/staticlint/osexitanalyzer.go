package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// OsExitAnalyzer is analyzer that forbids the usage of os.Exit in the main function.
var OsExitAnalyzer = &analysis.Analyzer{
	Name: "noexit",
	Doc:  "checks for usage of os.Exit in the main function of the main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspectMain := func(n ast.Node) bool {
		// Look for the function calls
		if callExpr, ok := n.(*ast.CallExpr); ok {
			// Ensure it's a selector expression (e.g., os.Exit)
			if fun, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				if pkgIdent, ok := fun.X.(*ast.Ident); ok {
					// Check if it's calling os.Exit
					if pkgIdent.Name == "os" && fun.Sel.Name == "Exit" {
						pass.Reportf(callExpr.Pos(), "usage of os.Exit is not allowed in the main function of the main package")
					}
				}
			}
		}
		return true
	}

	// Only analyze the main package
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	// Traverse the AST of each file in the package
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			// Check if the node is a function declaration
			if funcDecl, ok := n.(*ast.FuncDecl); ok {
				// Check if the function name is "main"
				if funcDecl.Name.Name == "main" {
					// Traverse the body of the "main" function
					ast.Inspect(funcDecl.Body, inspectMain)
				}
			}
			return true
		})
	}
	return nil, nil
}
