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
