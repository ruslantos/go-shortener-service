// Package noosexitanalyzer
// Вместо os.Exit рекомендуется:
//   - Использовать return с кодом возврата
//   - Применять log.Fatal для фатальных ошибок
//   - Возвращать ошибки из вложенных функций
//
// Пример ошибочного кода:
//
//	func main() {
//	    os.Exit(1) // вызовет ошибку анализатора
//	}
//
// Пример корректного кода:
//
//	func main() {
//	    if err := run(); err != nil {
//	        log.Fatal(err)
//	    }
//	}
package noosexitanalyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "noosexitanalyzer prohibits direct calls to os.Exit in main function of main package"

// Analyzer настраивает и возвращает анализатор для проверки вызовов os.Exit.
//
// Анализатор проверяет только функции main в пакете main. При обнаружении
// прямого вызова os.Exit в этих функциях выдает предупреждение.
//
// Требует наличия inspect.Analyzer в зависимостях.
var Analyzer = &analysis.Analyzer{
	Name:     "noosexit",
	Doc:      doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// run реализует основную логику анализатора.
//
// Функция проверяет, что анализируемый пакет является main, затем находит
// функцию main и проверяет её тело на наличие вызовов os.Exit.
// При обнаружении таких вызовов генерирует диагностическое сообщение.
func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		funcDecl := node.(*ast.FuncDecl)

		if funcDecl.Name.Name != "main" {
			return
		}

		ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			ident, ok := selExpr.X.(*ast.Ident)
			if !ok {
				return true
			}

			if ident.Name == "os" && selExpr.Sel.Name == "Exit" {
				pass.Reportf(callExpr.Pos(), "direct call to os.Exit in main function of main package is forbidden")
			}

			return true
		})
	})

	return nil, nil
}
