package interpreter

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"time"

	"github.com/siadat/well/syntax/ast"
	"github.com/siadat/well/syntax/parser"
	"github.com/siadat/well/syntax/scanner"
	"github.com/siadat/well/syntax/strs/expander"
)

type Interpreter struct {
	Stdout  io.Writer
	Stderr  io.Writer
	Verbose bool
	Debug   bool

	parser *parser.Parser
}

func NewInterpreter(stdout, stderr io.Writer) *Interpreter {
	return &Interpreter{
		Stdout: stdout,
		Stderr: stderr,
	}
}

func (interp *Interpreter) SetVerbose(v bool) {
	interp.Verbose = v
}

func (interp *Interpreter) SetDebug(v bool) {
	interp.Debug = v
}

func (interp *Interpreter) Eval(src io.Reader, env Environment) (Object, error) {
	interp.parser = parser.NewParser()
	interp.parser.SetDebug(interp.Debug)
	var node, err = interp.parser.Parse(src)
	if err != nil {
		return nil, err
	}
	return interp.evalParsed(node, env)
}

func (interp *Interpreter) evalParsed(node ast.Node, env Environment) (obj Object, retErr error) {
	defer func() {
		var err = recover()
		switch err := err.(type) {
		case nil:
			return
		case InterpError:
			// if interp.Debug { debug.PrintStack() }
			retErr = err
		default:
			fmt.Printf("unexpected error: %s\n", err)
			debug.PrintStack()
		}
	}()
	interp.eval(node, env)
	return
}

func (interp *Interpreter) builtins() map[string]*Builtin {
	var builtinsSlice = []*Builtin{
		{
			"external", func(posArgs []Object, kvArgs map[string]Object) error {
				var arg = posArgs[0].(*String)
				if interp.Debug {
					fmt.Fprintf(os.Stderr, "call external command %#v\n", arg)
				}
				var cmd = exec.Command(arg.AsArgs[0], arg.AsArgs[1:]...)
				cmd.Stdout = interp.Stdout
				cmd.Stderr = interp.Stdout
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("external command failed: %v", err)
				}
				return nil
			},
		},
		{
			"echo", func(posArgs []Object, kvArgs map[string]Object) error {
				for i, arg := range posArgs {
					fmt.Fprint(interp.Stdout, arg.GoValue())
					if i != len(posArgs)-1 {
						fmt.Fprint(interp.Stdout, " ")
					}
				}
				fmt.Fprint(interp.Stdout, "\n")
				return nil
			},
		},
		{
			"date", func(posArgs []Object, kvArgs map[string]Object) error {
				fmt.Fprintf(interp.Stdout, "%v\n", time.Now())
				return nil
			},
		},
	}
	var m = make(map[string]*Builtin, len(builtinsSlice))
	for _, b := range builtinsSlice {
		m[b.Name] = b
	}
	return m
}

func (interp *Interpreter) getMainFuncDecl(file *ast.Root) *ast.FuncDecl {
	for _, decl := range file.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if decl.Name.Name == "main" {
				return decl
			}
		}
	}
	panic(interp.newError(-1, "missing main func declaration"))
}

func (interp *Interpreter) eval(node ast.Node, env Environment) Object {
	if interp.Debug {
		fmt.Printf("DEBUG eval %T %+v\n", node, node)
	}
	switch node := node.(type) {
	case *ast.Root:
		for _, decl := range node.Decls {
			interp.eval(decl, env)
		}
		return interp.eval(ast.CallExpr{
			Fun: ast.Ident{Name: "main", Position: -1},
			Arg: ast.ParenExpr{
				Exprs: nil,
			},
		}, env)
	case ast.ParenExpr:
		var objs []Object
		for _, expr := range node.Exprs {
			objs = append(objs, interp.eval(expr, env))
		}
		return &Paren{Objects: objs}
	case ast.ExprStmt:
		return interp.eval(node.X, env)
	case ast.CallExpr:
		funcDef := interp.eval(node.Fun, env)
		switch funcDef := funcDef.(type) {
		case *Builtin:
			var positionals []Object
			var keywords map[string]Object

			for _, arg := range node.Arg.Exprs {
				var obj = interp.eval(arg, env)
				positionals = append(positionals, obj)
			}

			var userErr = funcDef.Func(positionals, keywords)
			if userErr != nil {
				panic(interp.newError(node.Pos(), "%s", userErr))
			}
			return nil
		case *Function:
			var positionalArgNames []string
			for _, param := range funcDef.Signature.ArgNames {
				positionalArgNames = append(positionalArgNames, param)
			}

			positionalIdx := 0
			newEnv := env.NewScope()
			for _, arg := range node.Arg.Exprs {
				obj := interp.eval(arg, env)

				name := positionalArgNames[positionalIdx]
				newEnv.MustSet(name, obj) // interp.eval(name, env))
				positionalIdx += 1
			}

			// ast.BlockStmt
			// return interp.eval(funcDef.Body, newEnv)

			for _, stmt := range funcDef.Body {
				result := interp.eval(stmt, env)
				switch result := result.(type) {
				case *ReturnStmt:
					return result.Expr
				}
			}
			return nil
		default:
			panic(interp.newError(node.Pos(), "unsupported function type %T", funcDef))
		}
	case ast.ReturnStmt:
		return &ReturnStmt{Expr: interp.eval(node.Expr, env)}
	case ast.Ident:
		val, err := env.Get(node.Name)
		if err != nil {
			// try builtins:
			if b, ok := interp.builtins()[node.Name]; ok {
				return b
			}
			panic(interp.newError(node.Pos(), "ident %q not found: %v", node.Name, err))
		}
		return val
	case ast.FuncDecl:
		env.MustSet(node.Name.Name, &Function{
			Name:      node.Name.Name,
			Signature: node.Signature,
			Body:      node.Statements,
			Env:       env,
		})
		// Old note: We return nil, because function declaration in this
		// language are not expressions atm. If in the futre you want to
		// support anonymous function, return the Function here.
		return nil
	case ast.Integer:
		return &Integer{Value: node.Value}
	case ast.Float:
		return &Float{Value: node.Value}
	case ast.LetDecl:
		env.MustSet(node.Name.Name, interp.eval(node.Rhs, env))
		return nil
	case ast.String:
		var envFunc = func(name string) interface{} {
			val, err := env.Get(name)
			if err != nil {
				panic(interp.newError(node.Pos(), "ident %q not found: %v", name, err))
			}
			return val
		}

		var rendered, err = expander.EncodeToString(node.Root, envFunc)
		if err != nil {
			panic(interp.newError(node.Pos(), "failed to render string: %v", err))
		}
		if interp.Verbose {
			fmt.Fprintf(interp.Stderr, "+%s\n", rendered)
		}

		var words = expander.EncodeToCmdArgs(node.Root, envFunc)
		return &String{
			AsSingle: rendered,
			AsArgs:   words,
		}
	default:
		panic(interp.newError(node.Pos(), "unsupported node type %T", node))
	}
}

type InterpError struct {
	err error
}

func (i InterpError) Error() string {
	return i.err.Error()
}

func (interp *Interpreter) newError(pos scanner.Pos, f string, args ...any) error {
	var lines = interp.parser.MarkAt(pos, fmt.Sprintf(f, args...), false)
	return InterpError{fmt.Errorf("%s", strings.Join(lines, "\n"))}
}
