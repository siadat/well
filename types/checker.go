package types

import (
	"fmt"
	"io"
	"runtime/debug"
	"strings"

	"github.com/siadat/well/syntax/ast"
	"github.com/siadat/well/syntax/parser"
	"github.com/siadat/well/syntax/scanner"
)

func NewChecker() typeChecker {
	return typeChecker{
		types: make(map[ast.Expr]Type),
	}
}

type typeChecker struct {
	types  map[ast.Expr]Type
	parser *parser.Parser
	debug  bool
}

func (tc *typeChecker) SetDebug(v bool) {
	tc.debug = v
}

func (tc typeChecker) Check(src io.Reader) (types map[ast.Expr]Type, retErr error) {
	tc.parser = parser.NewParser()
	tc.parser.SetDebug(tc.debug)
	var node, err = tc.parser.Parse(src)
	if err != nil {
		return nil, err
	}

	defer func() {
		var err = recover()
		switch err := err.(type) {
		case nil:
			return
		case Error:
			// if interp.Debug { debug.PrintStack() }
			retErr = err
		default:
			fmt.Printf("unexpected error: %s\n", err)
			debug.PrintStack()
		}
	}()

	tc.check(node)
	types = tc.types
	return
}

func (tc typeChecker) check(node ast.Node) {
	switch node := node.(type) {
	case *ast.Root:
		for _, decl := range node.Decls {
			tc.check(decl)
		}
	case ast.ParenExpr:
		for _, expr := range node.Exprs {
			tc.check(expr)
		}
	case ast.ExprStmt:
		tc.check(node.X)
	case ast.CallExpr:
		tc.types[node.Fun] = WellType{"Function"}
		tc.check(node.Arg)
	case ast.ReturnStmt:
		tc.check(node.Expr)
	case ast.Ident:
		// NoOp
	case ast.FuncSignature:
		// TODO
	case ast.FuncDecl:
		tc.types[node.Name] = WellType{"Function"}
		tc.check(node.Signature)
		for _, stmt := range node.Statements {
			tc.check(stmt)
		}
	case ast.Integer:
		tc.types[node] = WellType{"Integer"}
	case ast.Float:
		tc.types[node] = WellType{"Float"}
	case ast.String:
		tc.types[node] = WellType{"String"}
	case ast.LetDecl:
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
