package exitcheck

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer - это статический анализатор, который проверяет прямые вызовы os.Exit
// в функции main пакета main.
//
// Этот анализатор помогает обеспечить лучшие практики обработки ошибок,
// запрещая использование os.Exit в функциях main, что может
// усложнить тестирование и корректное завершение работы.
//
// Анализатор сообщает о проблемах, когда находит:
//   - Прямые вызовы os.Exit() в любой функции в пакете main
//   - Вызовы log.Fatal(), которые внутренне вызывают os.Exit()
//   - Вызовы log.Fatalf(), которые внутренне вызывают os.Exit()
//
// Примеры нарушений:
//
//	func main() {
//	    if err != nil {
//	        os.Exit(1) // Это будет сообщено
//	    }
//	    log.Fatal("error") // Это тоже будет сообщено
//	}
//
// Рекомендуемые альтернативы:
//
//	func main() {
//	    if err != nil {
//	        log.Printf("error: %v", err)
//	        return // или используйте правильный механизм обработки ошибок
//	    }
//	}
var Analyzer = &analysis.Analyzer{
	Name:     "exitcheck",
	Doc:      "check for os.Exit calls in main package",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// run - это основная функция анализатора, которая выполняет фактический анализ.
// Она проверяет AST каждого файла в пакете и сообщает о любых вызовах os.Exit
// или log.Fatal, найденных в пакете main.
//
// Параметры:
//   - pass: Проход анализа, содержащий информацию о текущем пакете
//
// Функция:
//   - Использует инспектор для обхода AST
//   - Проверяет, является ли текущий пакет "main"
//   - Ищет вызовы функций os.Exit, log.Fatal и log.Fatalf
//   - Сообщает диагностические сообщения для любых найденных нарушений
func run(pass *analysis.Pass) (interface{}, error) {
	// Проверяем только пакет main
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Определяем типы узлов, которые мы хотим проверить
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	// Обходим AST
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return
		}

		// Проверяем, является ли это вызовом функции
		fun, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}

		// Получаем имя пакета и имя функции
		ident, ok := fun.X.(*ast.Ident)
		if !ok {
			return
		}

		pkgName := ident.Name
		funcName := fun.Sel.Name

		// Проверяем вызовы os.Exit
		if pkgName == "os" && funcName == "Exit" {
			pass.Reportf(call.Pos(), "direct call to os.Exit in main package")
			return
		}

		// Проверяем вызовы log.Fatal и log.Fatalf
		if pkgName == "log" && (funcName == "Fatal" || funcName == "Fatalf") {
			pass.Reportf(call.Pos(), "call to log.%s in main package (calls os.Exit)", funcName)
			return
		}
	})

	return nil, nil
}
