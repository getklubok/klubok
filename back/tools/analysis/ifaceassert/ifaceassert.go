package ifaceassert

import (
	"go/ast"
	"go/token"
	"reflect"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer проверяет, что для каждого интерфейса в пакете есть compile-time assertion вида:
//
//	var _ InterfaceName = (*SomeType)(nil)
//
// Это помогает обнаружить несоответствие реализации интерфейсу прямо в пакете, где интерфейс объявлен.
func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:       "ifaceassert",
		Doc:        "require var _ Interface = (*T)(nil) assertion for every interface",
		Run:        run,
		ResultType: reflect.TypeOf(struct{}{}),
	}
}

// run – основная логика проверки.
func run(pass *analysis.Pass) (any, error) {
	interfaces := collectInterfaces(pass)
	asserts := collectAsserts(pass)
	for name, pos := range interfaces {
		if !asserts[name] {
			pass.Reportf(pos, "interface %s must have an assertion: var _ %s = (*T)(nil)", name, name)
		}
	}

	return struct{}{}, nil
}

// collectInterfaces все имена интерфейсов, объявленных в пакете, и их позиции.
func collectInterfaces(pass *analysis.Pass) map[string]token.Pos { //nolint:gocognit // TODO: refactor
	interfaces := make(map[string]token.Pos)

	for _, file := range pass.Files {
		if isGeneratedFile(file) {
			continue
		}
		for _, decl := range file.Decls {
			gen, declOK := decl.(*ast.GenDecl)
			if !declOK || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				ts, specOK := spec.(*ast.TypeSpec)
				if !specOK {
					continue
				}
				if _, isIface := ts.Type.(*ast.InterfaceType); isIface {
					interfaces[ts.Name.Name] = ts.Pos()
				}
			}
		}
	}

	return interfaces
}

// collectAsserts собирает все assert-объявления вида var _ Interface = (...)(nil).
func collectAsserts(pass *analysis.Pass) map[string]bool { //nolint:gocognit // TODO: refactor
	asserts := make(map[string]bool)

	for _, file := range pass.Files {
		if isGeneratedFile(file) {
			continue
		}
		for _, decl := range file.Decls {
			gen, declOK := decl.(*ast.GenDecl)
			if !declOK || gen.Tok != token.VAR {
				continue
			}
			for _, spec := range gen.Specs {
				vs, specOK := spec.(*ast.ValueSpec)
				if !specOK || len(vs.Names) != 1 || vs.Names[0].Name != "_" {
					continue
				}
				// Тип переменной должен быть *ast.Ident с именем интерфейса
				if ident, typeOK := vs.Type.(*ast.Ident); typeOK {
					// Проверяем, что значение – это вызов с nil (упрощённо)
					if hasNilCall(vs.Values) {
						asserts[ident.Name] = true
					}
				}
			}
		}
	}

	return asserts
}

// isGeneratedFile возвращает true, если файл помечен как сгенерированный
// стандартным маркером "// Code generated ... DO NOT EDIT.".
func isGeneratedFile(file *ast.File) bool {
	for _, group := range file.Comments {
		for _, comment := range group.List {
			if strings.Contains(comment.Text, "Code generated") {
				return true
			}
		}
	}
	return false
}

// hasNilCall проверяет, что среди значений есть вызов вида (...)(nil) или (*T)(nil) или T(nil),
// достаточно общего CallExpr с nil.
func hasNilCall(exprs []ast.Expr) bool {
	for _, e := range exprs {
		if call, eOK := e.(*ast.CallExpr); eOK {
			// Ищем nil в аргументах
			for _, arg := range call.Args {
				if ident, argOK := arg.(*ast.Ident); argOK && ident.Name == "nil" {
					return true
				}
			}
		}
	}
	return false
}
