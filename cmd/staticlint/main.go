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
//	go build -o multichecker
//
// 2. Run the multichecker on your Go code:
//
//	./multichecker ./...
//
//	This command will analyze all Go files in the current directory and its
//	subdirectories.
//
// The multichecker combines several analyzers, including:
//
// - Standard analyzers: Includes analyzers from golang.org/x/tools/go/analysis/passes.
// - simple: To simplify single recieve channel operation.
// - stylecheck: Discourages the use of dot imports.
// - quickfix: Suggests applying De Morgan's law.
// - errcheck: Ensures that errors are checked.
// - bodyclose: Checks that HTTP response bodies are correctly closed.
// - staticcheck: A comprehensive set of static analysis checks from "SA" family.
// - Custom OsExitAnalyzer: Disallows the use of os.Exit in the main function of the main package.
package main

import (
	"slices"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	checks := []*analysis.Analyzer{
		asmdecl.Analyzer,          // Checks assembly files against Go declarations
		assign.Analyzer,           // Detects useless assignments
		atomic.Analyzer,           // Checks for misaligned 64-bit values in atomic operations
		bools.Analyzer,            // Detects common mistakes involving boolean operators
		buildtag.Analyzer,         // Ensures that build tags are used correctly
		cgocall.Analyzer,          // Detects unsafe cgo calls
		composite.Analyzer,        // Checks for unkeyed composite literals
		copylock.Analyzer,         // Detects locks that are copied
		deepequalerrors.Analyzer,  // Detects improper usage of reflect.DeepEqual with errors
		errorsas.Analyzer,         // Checks correct usage of errors.As
		httpresponse.Analyzer,     // Detects common mistakes with HTTP response handling
		ifaceassert.Analyzer,      // Detects improper type assertions
		inspect.Analyzer,          // Provides a framework for other analyzers
		lostcancel.Analyzer,       // Detects lost context cancelation
		nilfunc.Analyzer,          // Detects calling a nil function
		printf.Analyzer,           // Checks for incorrect format strings in printf-like functions
		shadow.Analyzer,           // Detects shadowed variables
		shift.Analyzer,            // Detects incorrect bit shift operations
		sortslice.Analyzer,        // Detects incorrect usage of sort.Slice
		stdmethods.Analyzer,       // Ensures standard method signatures are correct
		stringintconv.Analyzer,    // Detects string(int) conversions that may not be intended
		structtag.Analyzer,        // Ensures struct tags are well-formed
		testinggoroutine.Analyzer, // Detects goroutines in tests that might leak
		tests.Analyzer,            // Detects common mistakes in tests
		unmarshal.Analyzer,        // Detects issues when unmarshaling data
		unreachable.Analyzer,      // Detects unreachable code
		unsafeptr.Analyzer,        // Detects unsafe pointer usage
		unusedresult.Analyzer,     // Detects when a result from a function call is unused

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
