// Package main provides a multichecker that combines a custom analyzer to
// disallow the usage of os.Exit in the main package with other common Go
// static analysis checks. This multichecker helps enforce best practices
// in Go code by combining several useful checks into one tool.
//
// Launch Mechanics:
//
// To use this multichecker, follow these steps:
//
// 1. Build the multichecker executable:
//
//    go build -o multichecker
//
// 2. Run the multichecker on your Go code:
//
//    ./multichecker ./...
//
//    This command will analyze all Go files in the current directory and its
//    subdirectories.
//
// The multichecker combines several analyzers, including:
//
// - simple: To simplify single recieve channel operation.
// - stylecheck: Discourages the use of dot imports.
// - quickfix: Suggests applying De Morgan's law.
// - errcheck: Ensures that errors are checked.
// - bodyclose: Checks that HTTP response bodies are correctly closed.
// - staticcheck: A comprehensive set of static analysis checks.
// - Custom OsExitAnalyzer: Disallows the use of os.Exit in the main function of the main package.
package main

import (
	"slices"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	checks := []*analysis.Analyzer{
		simple.Analyzers[0].Analyzer,     // Use plain channel send or receive instead of single-case select
		stylecheck.Analyzers[1].Analyzer, // Dot imports are discouraged
		quickfix.Analyzers[0].Analyzer,   // Apply De Morgan’s law

		errcheck.Analyzer,  // Errcheck for unchecked errors
		bodyclose.Analyzer, // Checks whether resp.Body closed correctly

		OsExitAnalyzer,
	}

	staticchecks := make([]*analysis.Analyzer, len(staticcheck.Analyzers))
	for i, v := range staticcheck.Analyzers {
		staticchecks[i] = v.Analyzer
	}
	checks = slices.Concat(checks, staticchecks)

	multichecker.Main(
		checks...,
	)
}
