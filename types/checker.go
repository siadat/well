package types

import (
	"fmt"
	"io"
	"strings"

	"github.com/siadat/well/erroring"
	"github.com/siadat/well/fumt"
	"github.com/siadat/well/syntax/ast"
	"github.com/siadat/well/syntax/parser"
	"github.com/siadat/well/syntax/scanner"
)

func NewChecker() typeChecker {
	return typeChecker{
		types:         make(map[ast.Expr]Type),
		files:         make(map[string]scanner.Pos),
		commands:      make(map[string]scanner.Pos),
		externalDecls: make(map[string]struct{}),
	}
}

type typeChecker struct {
	types  map[ast.Expr]Type
	parser *parser.Parser
	debug  bool

	files    map[string]scanner.Pos
	commands map[string]scanner.Pos

	externalDecls map[string]struct{}
}

func (tc *typeChecker) UnresolvedDependencies() []string {
	const lenLimit = 110
	var rets []string
	for name, pos := range tc.files {
		var line, col = tc.parser.GetLineColAt(pos)
		if len(name) > lenLimit {
			name = name[:lenLimit] + "..."
		}
		rets = append(rets, fmt.Sprintf("%d:%d \t%s", line+1, col+1, name))
	}
	for name, pos := range tc.commands {
		if _, ok := tc.externalDecls[name]; ok {
			continue
		}
		var line, col = tc.parser.GetLineColAt(pos)
		if len(name) > lenLimit {
			name = name[:lenLimit] + "..."
		}
		rets = append(rets, fmt.Sprintf("%d:%d \t%s", line+1, col+1, name))
	}
	return rets
}

func (tc *typeChecker) SetDebug(v bool) {
	tc.debug = v
}

func (tc *typeChecker) Check(src io.Reader) (map[ast.Expr]Type, error) {
	tc.parser = parser.NewParser()
	tc.parser.SetDebug(tc.debug)
	var node, parseErr = tc.parser.Parse(src)
	if parseErr != nil {
		return nil, parseErr
	}

	return erroring.CallAndRecover[Error](func() map[ast.Expr]Type {
		tc.check(node)
		return tc.types
	})
}

func (tc *typeChecker) check(node ast.Node) {
	switch node := node.(type) {
	case *ast.Root:
		for _, decl := range node.Decls {
			tc.check(decl)
		}
	case *ast.ParenExpr:
		for _, expr := range node.Exprs {
			tc.check(expr)
		}
	case *ast.ExprStmt:
		tc.check(node.X)
	case *ast.CallExpr:
		tc.types[node] = WellType{"Function"}
		tc.types[node.Fun] = WellType{"Function"}
		tc.check(node.Arg)

		switch fun := node.Fun.(type) {
		case *ast.Ident:
			switch fun.Name {
			case "pipe", "pipe_capture":
				for _, expr := range node.Arg.Exprs {
					switch expr := expr.(type) {
					case *ast.CallExpr:
						if expr, ok := expr.Fun.(*ast.Ident); ok {
							var formater = fumt.NewFormater()
							var command = formater.FormatNode(expr)
							tc.commands[command] = node.Pos()
						} else {
							panic(tc.newError(node.Pos(), "args to pipe must be simple call expressions"))
						}
					default:
						panic(tc.newError(node.Pos(), "args to pipe must be call expressions"))
					}
				}
			}
		}
	case *ast.ReturnStmt:
		tc.check(node.Expr)
	case *ast.Ident:
		// NoOp
	case *ast.IfStmt:
		// NoOp
	case *ast.FuncSignature:
		// TODO
	case *ast.BinaryExpr:
		// TODO
	case *ast.FuncDecl:
		tc.types[node.Name] = WellType{"Function"}

		if node.IsExternal {
			tc.externalDecls[node.Name.Name] = struct{}{}
		}

		tc.check(node.Signature)
		for _, stmt := range node.Body.Statements {
			tc.check(stmt)
		}
	case *ast.Integer:
		tc.types[node] = WellType{"Integer"}
	case *ast.Float:
		tc.types[node] = WellType{"Float"}
	case *ast.String:
		tc.types[node] = WellType{"String"}
	case *ast.LetDecl:
		tc.check(node.Rhs)
		tc.types[node.Name] = tc.types[node.Rhs]
	default:
		panic(tc.newError(node.Pos(), "unsupported node type %T", node))
	}
}

type Error struct {
	err error
}

func (i Error) Error() string {
	return i.err.Error()
}

func (tc *typeChecker) newError(pos scanner.Pos, f string, args ...any) error {
	var lines = tc.parser.MarkAt(pos, fmt.Sprintf(f, args...), false)
	return Error{fmt.Errorf("%s", strings.Join(lines, "\n"))}
}
