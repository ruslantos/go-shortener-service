package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	checks := map[string]bool{
		"SA5000":   true,
		"SA6000":   true,
		"SA9004":   true,
		"fakejson": true,
	}
	var mychecks []*analysis.Analyzer
	for _, v := range staticcheck.Analyzers {
		if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	allChecks := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
	}
	allChecks = append(allChecks, mychecks...)

	multichecker.Main(allChecks...)
}
