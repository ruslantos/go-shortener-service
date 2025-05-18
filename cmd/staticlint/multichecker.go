package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"honnef.co/go/tools/staticcheck"

	"github.com/ruslantos/go-shortener-service/cmd/staticlint/noosexitanalyzer"
)

// main запускает multichecker с выбранными анализаторами.
// multichecker позволяет запускать несколько статических анализаторов кода одновременно.
// В данном случае используются анализаторы из стандартной библиотеки Go и дополнительные из staticcheck.
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
		printf.Analyzer,           // Проверяет форматные строки в вызовах Printf и подобных функций.
		shadow.Analyzer,           // Проверяет наличие скрытия переменных.
		structtag.Analyzer,        // Проверяет корректность тегов структур.
		errorsas.Analyzer,         // Проверяет корректное использование функции errors.As.
		unmarshal.Analyzer,        // Проверяет корректное использование функций Unmarshal.
		noosexitanalyzer.Analyzer, // анализатор для проверки использования os.Exit
	}
	allChecks = append(allChecks, mychecks...)

	multichecker.Main(allChecks...)
}
