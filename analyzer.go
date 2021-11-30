package samealias

import (
	"bufio"
	"flag"
	"go/ast"
	"go/token"
	"os"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
)

type aliaspath struct {
	alias    string
	position token.Position
}

var imports = map[string]aliaspath{}

//nolint:gochecknoglobals
var flagSet flag.FlagSet

//nolint:gochecknoglobals
var (
	skipAutogens bool
)

//nolint:gochecknoinits
func init() {
	flagSet.BoolVar(&skipAutogens, "skipAutogens", false, "should the linter execute on autogen files as well")
}

func NewAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:  "samealias",
		Doc:   "check different aliases for same package",
		Run:   run,
		Flags: flagSet,
	}
}

func run(pass *analysis.Pass) (interface{}, error) {

	for _, file := range pass.Files {
		filename := pass.Fset.Position((file.Pos())).Filename
		if skipAutogens && isAutogenFile(filename) {
			continue
		}
		ast.Inspect(file, func(node ast.Node) bool {
			f, ok := node.(*ast.ImportSpec)
			if !ok {
				if node == nil {
					return true
				}
				return true
			}

			if f.Name == nil {
				return true
			}

			alias := ""
			if f.Name != nil {
				alias = f.Name.String()
			}

			if alias == "." {
				return true // Dot aliases are generally used in tests, so ignore.
			}
			if strings.HasPrefix(alias, "_") {
				return true // Used by go test and for auto-includes, not a conflict.
			}

			path, err := strconv.Unquote(f.Path.Value)
			if err != nil {
				pass.Reportf(f.Pos(), "import not quoted")
			}

			if alias != "" {
				val, ok := imports[path]
				if ok {
					if val.alias != alias {
						pass.Reportf(f.Pos(), "package %q have alias %q, conflict with %q in %q", path, alias, val.alias, val.position)
					}
				} else {
					imports[path] = aliaspath{alias: alias, position: pass.Fset.Position(f.Pos())}
				}
			}

			return true
		})
	}

	return nil, nil
}

// autogen files containe "do not edit" before package key word
func isAutogenFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines := strings.ToUpper(scanner.Text())
		if strings.Contains(lines, "PACKAGE") {
			return false
		}
		if strings.Contains(lines, "DO NOT EDIT") {
			return true
		}
	}

	return false
}
