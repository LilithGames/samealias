package samealias

import (
	"bufio"
	"fmt"
	"go/ast"
	"os"
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
	filename := pass.Fset.Position(pass.Files[0].Pos()).Filename

	res, err := isAutogenFile(filename)

	if err != nil {
		panic(err)
	} else if !res {
		inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		inspect.Preorder([]ast.Node{(*ast.ImportSpec)(nil)}, func(n ast.Node) {
			visitImportSpecNode(n.(*ast.ImportSpec), pass)
		})
		fmt.Println(filename, "not autogen file, need process")
	} else {
		fmt.Println(filename, "autogen file, no need process")
	}

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

func isAutogenFile(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines := strings.ToUpper(scanner.Text())
		if strings.Contains(lines, "PACKAGE") {
			return false, scanner.Err()
		}
		if strings.Contains(lines, "DO NOT EDIT") {
			return true, scanner.Err()
		}
	}
	return false, scanner.Err()
}
