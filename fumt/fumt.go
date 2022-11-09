package fumt

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/siadat/well/erroring"
	"github.com/siadat/well/syntax/ast"
	"github.com/siadat/well/syntax/parser"
	"github.com/siadat/well/syntax/scanner"
)

// This is an experimental formater. It does not support comments yet. It will
// proabbly be re-written using a different approach. Ideas are welcome. :)

type formater struct {
	indentLevel int
	parser      *parser.Parser
	debug       bool
}

func NewFormater() *formater {
	return &formater{}
}

func (ft *formater) SetDebug(v bool) {
	ft.debug = v
}

func (ft *formater) Format(src io.Reader, out io.Writer) error {
	ft.parser = parser.NewParser()
	ft.parser.SetDebug(ft.debug)
	ft.parser.SetIncludeComments(false) // TODO: include and preserve comments
	var node, parseErr = ft.parser.Parse(src)
	if parseErr != nil {
		return parseErr
	}

	var _, err = erroring.CallAndRecover[Error](func() struct{} {
		fmt.Fprintf(out, ft.format(node))
		return struct{}{}
	})
	return err
}

func (ft *formater) format(node ast.Node) string {
	switch node := node.(type) {
	case *ast.Root:
		var buf bytes.Buffer
		for _, decl := range node.Decls {
			fmt.Fprintf(&buf, "%s", ft.format(decl))
		}
		return buf.String()
	case *ast.BlockStmt:
		var buf bytes.Buffer
		ft.indentLevel += 1
		for _, stmt := range node.Statements {
			fmt.Fprintf(&buf, "%s", ft.format(stmt))
		}
		ft.indentLevel -= 1
		return ft.indent() + "{\n" +
			buf.String() +
			ft.indent() + "}\n"
	case *ast.FuncDecl:
		return ft.indent() + fmt.Sprintf("function %s%s %s\n", node.Name.Name, ft.format(node.Signature), ft.format(node.Body))
	case *ast.LetDecl:
		return ft.indent() + fmt.Sprintf("let %s = %s\n", node.Name.Name, ft.format(node.Rhs))
	case *ast.ExprStmt:
		return ft.indent() + fmt.Sprintf("%s\n", ft.format(node.X))
	case *ast.CallExpr:
		return fmt.Sprintf("%s%s", ft.format(node.Fun), ft.format(node.Arg))
	case *ast.Ident:
		return node.Name
	case *ast.String:
		return node.StringLit
	case *ast.Integer:
		return fmt.Sprintf("%d", node.Value)
	case *ast.Float:
		return fmt.Sprintf("%v", node.Value)
	case *ast.FuncSignature:
		var arguments = func() string {
			var args []string
			for _, arg := range node.Args {
				args = append(args, fmt.Sprintf("%s %s", arg.Name, arg.Type))
			}
			return strings.Join(args, ", ")
		}()

		var returns = func() string {
			var rets []string
			for _, ret := range node.RetTypes {
				rets = append(rets, ret)
			}
			if len(rets) == 0 {
				return ""
			} else if len(rets) == 1 {
				return rets[0]
			} else {
				return strings.Join(rets, ", ")
			}
		}()

		if returns == "" {
			return fmt.Sprintf("(%s)", arguments)
		} else {
			return fmt.Sprintf("(%s) %s", arguments, returns)
		}
	case *ast.ParenExpr:
		var perline = false
		if perline {
			var buf bytes.Buffer
			ft.indentLevel += 1
			for _, expr := range node.Exprs {
				fmt.Fprintf(&buf, "%s%s,\n", ft.indent(), ft.format(expr))
			}
			ft.indentLevel -= 1
			return "(\n" +
				buf.String() +
				ft.indent() + ")\n"
		} else {
			var exprs []string
			for _, expr := range node.Exprs {
				exprs = append(exprs, ft.format(expr))
			}
			return "(" + strings.Join(exprs, ", ") + ")"
		}
	default:
		return fmt.Sprintf("(TODO %T)", node)
	}
}

func (ft *formater) indent() string {
	return strings.Repeat("\t", ft.indentLevel)
}

type Error struct {
	err error
}

func (i Error) Error() string {
	return i.err.Error()
}

func (ft *formater) newError(pos scanner.Pos, f string, args ...any) error {
	var lines = ft.parser.MarkAt(pos, fmt.Sprintf(f, args...), false)
	return Error{fmt.Errorf("%s", strings.Join(lines, "\n"))}
}
