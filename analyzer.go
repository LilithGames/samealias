package samealias

import (
	"fmt"
	"go/ast"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var imports = map[string]string{}

var Analyzer = &analysis.Analyzer{
	Name: "samealias",
	Doc:  "check different aliases for same package",
	Run:  run,

	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.ImportSpec)(nil)}, func(n ast.Node) {
		visitImportSpecNode(n.(*ast.ImportSpec), pass)
	})
	return nil, nil
}

func visitImportSpecNode(node *ast.ImportSpec, pass *analysis.Pass) {
	if node.Name == nil {
		return
	}

	alias := ""
	if node.Name != nil {
		alias = node.Name.String()
	}

	if alias == "." {
		return // Dot aliases are generally used in tests, so ignore.
	}

	if strings.HasPrefix(alias, "_") {
		return // Used by go test and for auto-includes, not a conflict.
	}

	path, err := strconv.Unquote(node.Path.Value)
	if err != nil {
		pass.Reportf(node.Pos(), "import not quoted")
	}

	if alias != "" {
		val, ok := imports[path]
		if ok {
			if val != alias {
				message := fmt.Sprintf("package %q have different alias, %q, %q", path, alias, val)

				pass.Report(analysis.Diagnostic{
					Pos:     node.Pos(),
					End:     node.End(),
					Message: message,
					SuggestedFixes: []analysis.SuggestedFix{{
						Message: "Use same alias or do not use alias",
					}},
				})
			}
		} else {
			imports[path] = alias
		}
	}
}
