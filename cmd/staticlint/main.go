package main

import (
	"go/ast"
	"go/types"
	"strings"

	"github.com/sonatard/noctx"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/ast/astutil"

	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"

	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

var forbiddenCallsOutsideMainAnalyzer = &analysis.Analyzer{
	Name: "forbiddenCallsOutsideMain",
	Doc:  "проверка на os.Exit() и fatal / panic вне функции main",
	Run:  runForbiddenCallsOutsideMain,
}

func runForbiddenCallsOutsideMain(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}

		var funcStack []bool

		astutil.Apply(file, func(cursor *astutil.Cursor) bool {
			switch n := cursor.Node().(type) {
			case *ast.FuncDecl:
				isRealMain := pass.Pkg.Name() == "main" &&
					n.Recv == nil &&
					n.Name.Name == "main"

				funcStack = append(funcStack, isRealMain)

			case *ast.FuncLit:
				funcStack = append(funcStack, false)

			case *ast.CallExpr:
				if isInsideRealMain(funcStack) {
					return true
				}

				if name := forbiddenCallName(pass, n); name != "" {
					pass.Reportf(n.Pos(), "запрещенный вызов %s вне функции main", name)
				}
			}

			return true
		}, func(cursor *astutil.Cursor) bool {
			switch cursor.Node().(type) {
			case *ast.FuncDecl, *ast.FuncLit:
				funcStack = funcStack[:len(funcStack)-1]
			}

			return true
		})
	}

	return nil, nil
}

func isInsideRealMain(funcStack []bool) bool {
	return len(funcStack) > 0 && funcStack[len(funcStack)-1]
}

func forbiddenCallName(pass *analysis.Pass, call *ast.CallExpr) string {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		obj := pass.TypesInfo.Uses[fun]

		builtin, ok := obj.(*types.Builtin)
		if ok && builtin.Name() == "panic" {
			return "panic"
		}

	case *ast.SelectorExpr:
		obj, ok := pass.TypesInfo.Uses[fun.Sel].(*types.Func)
		if !ok || obj.Pkg() == nil {
			return ""
		}

		switch obj.Pkg().Path() {
		case "os":
			if obj.Name() == "Exit" {
				return "os.Exit"
			}

		case "log":
			if isLogFatal(obj.Name()) {
				return "log." + obj.Name()
			}
		}
	}

	return ""
}

func isLogFatal(name string) bool {
	switch name {
	case "Fatal", "Fatalf", "Fatalln":
		return true
	default:
		return false
	}
}

func main() {
	analyzers := []*analysis.Analyzer{
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		defers.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		framepointer.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,

		forbiddenCallsOutsideMainAnalyzer,

		bodyclose.Analyzer,
		noctx.Analyzer,
	}

	for _, a := range staticcheck.Analyzers {
		if strings.HasPrefix(a.Analyzer.Name, "SA") {
			analyzers = append(analyzers, a.Analyzer)
		}
	}

	for _, a := range simple.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}
	for _, a := range stylecheck.Analyzers {
		if a.Analyzer.Name == "ST1005" {
			analyzers = append(analyzers, a.Analyzer)
			break
		}
	}

	multichecker.Main(analyzers...)
}
