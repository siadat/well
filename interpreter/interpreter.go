package interpreter

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/siadat/well/erroring"
	"github.com/siadat/well/syntax/ast"
	"github.com/siadat/well/syntax/parser"
	"github.com/siadat/well/syntax/scanner"
	"github.com/siadat/well/syntax/strs/expander"
	"github.com/siadat/well/syntax/token"
)

var NoPos scanner.Pos = -1

type Interpreter struct {
	Stdout  io.Writer
	Stderr  io.Writer
	Verbose bool
	Debug   bool

	parser *parser.Parser

	currEvalNode ast.Node
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

func (interp *Interpreter) evalParsed(node ast.Node, env Environment) (Object, error) {
	return erroring.CallAndRecover[InterpError](func() Object {
		interp.eval(node, env)
		return nil
	})
}

func (interp *Interpreter) builtins() map[string]*Builtin {
	// TODO: allow mocking external commands for test
	var builtinsSlice = []*Builtin{
		{
			"_exec", func(pipedArg Object, posArgs []Object, kvArgs map[string]Object) (Object, error) {
				if len(posArgs) != 1 {
					return nil, fmt.Errorf("_exec expects 1 args, got %d", len(posArgs))
				}

				var stdin io.Reader
				if pipedArg != nil {
					stdin = pipedArg.(*PipeStream).ReadCloser
				}
				var cmdArgs = posArgs[0].(*String).AsArgs
				var cmd = exec.CommandContext(context.TODO(), cmdArgs[0], cmdArgs[1:]...)

				// var pr, pw, err = os.Pipe()
				cmd.Stdin = stdin
				var stdout, err = cmd.StdoutPipe()
				if err != nil {
					return nil, err
				}

				var runErr = cmd.Start()
				return &PipeStream{ReadCloser: stdout}, runErr
			},
		},
		{
			"print_stream", func(pipedArg Object, posArgs []Object, kvArgs map[string]Object) (Object, error) {
				if len(posArgs) != 1 {
					return nil, fmt.Errorf("print_stream expects 1 args, got %d", len(posArgs))
				}
				var arg = posArgs[0]
				if arg == nil {
					return nil, fmt.Errorf("argument is %v", arg)
				}
				var stream = arg.(*PipeStream)
				for {
					var buf = make([]byte, 128)
					var n, err = stream.ReadCloser.Read(buf)
					if err == io.EOF {
						break
					}
					if err != nil {
						return nil, err
					}
					fmt.Fprint(interp.Stdout, string(buf[:n]))
				}
				return nil, nil
			},
		},
		{
			"println", func(pipedArg Object, posArgs []Object, kvArgs map[string]Object) (Object, error) {
				for i, arg := range posArgs {
					if arg == nil {
						return nil, fmt.Errorf("argument %d value is %v", i+1, arg)
					}
					fmt.Fprint(interp.Stdout, arg.GoValue())
					if i != len(posArgs)-1 {
						fmt.Fprint(interp.Stdout, " ")
					}
				}
				fmt.Fprint(interp.Stdout, "\n")
				return nil, nil
			},
		},
		{
			"print", func(pipedArg Object, posArgs []Object, kvArgs map[string]Object) (Object, error) {
				for i, arg := range posArgs {
					fmt.Fprint(interp.Stdout, arg.GoValue())
					if i != len(posArgs)-1 {
						fmt.Fprint(interp.Stdout, " ")
					}
				}
				// fmt.Fprint(interp.Stdout, "\n")
				return nil, nil
			},
		},
		{
			"exit", func(pipedValue Object, posArgs []Object, kvArgs map[string]Object) (Object, error) {
				if len(posArgs) != 2 {
					return nil, fmt.Errorf("read expects 2 args, got %d", len(posArgs))
				}
				var code = posArgs[0].(*Integer).Value
				var msg = posArgs[1].(*String).AsSingle
				fmt.Fprintf(interp.Stderr, "%s\n", msg)
				os.Exit(code)
				return nil, nil
			},
		},
		{
			"read", func(pipedValue Object, posArgs []Object, kvArgs map[string]Object) (Object, error) {
				if len(posArgs) != 0 {
					return nil, fmt.Errorf("read expects 0 args, got %d", len(posArgs))
				}
				var scanner = bufio.NewScanner(os.Stdin)
				scanner.Scan()
				if err := scanner.Err(); err != nil {
					return nil, err
				}
				return &String{AsSingle: scanner.Text()}, nil
			},
		},
		{
			"read_regex", func(pipedValue Object, posArgs []Object, kvArgs map[string]Object) (Object, error) {
				if len(posArgs) != 1 {
					return nil, fmt.Errorf("read expects 1 arg, got %d", len(posArgs))
				}
				var arg = posArgs[0].(*String).AsSingle
				var re = regexp.MustCompile(arg)

				var scanner = bufio.NewScanner(os.Stdin)

				for {
					scanner.Scan()
					if err := scanner.Err(); err != nil {
						return nil, err
					}
					if re.MatchString(scanner.Text()) {
						break
					}
					fmt.Fprintf(interp.Stderr, "Entered text did not match %#q, please try again:\n", re)
				}

				return &String{AsSingle: scanner.Text()}, nil
			},
		},
		{
			"read_int", func(pipedValue Object, posArgs []Object, kvArgs map[string]Object) (Object, error) {
				if len(posArgs) > 1 {
					return nil, fmt.Errorf("read expects 0 or 1 arg, got %d", len(posArgs))
				}

				var defaultValue int
				var hasDefatulValue bool
				if len(posArgs) == 1 {
					hasDefatulValue = true
					defaultValue = posArgs[0].(*Integer).Value
				}

				var scanner = bufio.NewScanner(os.Stdin)

				for {
					scanner.Scan()
					if err := scanner.Err(); err != nil {
						return nil, err
					}

					if scanner.Text() == "" && hasDefatulValue {
						return &Integer{Value: defaultValue}, nil
					}

					var d, err = strconv.ParseInt(scanner.Text(), 10, 64)
					if err != nil {
						fmt.Fprintf(interp.Stderr, "Invalid int (%s), please try again:\n", err)
						continue
					}
					return &Integer{Value: int(d)}, nil
				}
			},
		},
		{
			"date", func(pipedValue Object, posArgs []Object, kvArgs map[string]Object) (Object, error) {
				return &String{AsSingle: fmt.Sprintf("%v", time.Now())}, nil
			},
		},
	}
	var m = make(map[string]*Builtin, len(builtinsSlice))
	for _, b := range builtinsSlice {
		m[b.Name] = b
	}
	m["echo"] = m["print"] // TODO: maybe deprecate echo?
	return m
}

func (interp *Interpreter) mustSet(env Environment, name string, obj Object) {
	if err := env.Set(name, obj); err != nil {
		panic(interp.newError(interp.currEvalNode.Pos(), "%s", err))
	}
}

func (interp *Interpreter) eval(node ast.Node, env Environment) Object {
	interp.currEvalNode = node
	if interp.Debug {
		fmt.Printf("[eval] %T %+v\n", node, node)
	}
	switch node := node.(type) {
	case *ast.Root:
		for _, decl := range node.Decls {
			interp.eval(decl, env)
		}
		return interp.eval(
			&ast.CallExpr{
				Fun: &ast.Ident{Name: "main", Position: NoPos},
				Arg: &ast.ParenExpr{Exprs: nil},
				PipedArg: &ast.ParenExpr{Exprs: []ast.Expr{
					&ast.Ident{Name: "MainStdin", Position: NoPos},
				}},
			},
			env,
		)
	case *ast.ParenExpr:
		var objs []Object
		for _, expr := range node.Exprs {
			objs = append(objs, interp.eval(expr, env))
		}
		return &Paren{Objects: objs}
	case *ast.ExprStmt:
		return interp.eval(node.X, env)
	case *ast.CallExpr:
		var funcDef = interp.eval(node.Fun, env)
		if interp.Verbose {
			switch f := node.Fun.(type) {
			case *ast.Ident:
				var line, col = interp.parser.GetLineColAt(f.Pos())
				fmt.Fprintf(os.Stderr, "+ called %v(...) at %d:%d\n", f.Name, line+1, col+1)
			default:
				panic(interp.newError(node.Pos(), "unsupported call expressiong of type %T", f))
			}
		}
		// TODO: trace function calls
		switch funcDef := funcDef.(type) {
		case *Builtin:
			var positionals []Object
			var pipedObjects []Object
			var keywords map[string]Object

			for _, arg := range node.Arg.Exprs {
				var obj = interp.eval(arg, env)
				positionals = append(positionals, obj)
			}
			for _, arg := range node.PipedArg.Exprs {
				var obj = interp.eval(arg, env)
				pipedObjects = append(pipedObjects, obj)
			}

			var pipedObject Object
			if len(pipedObjects) > 0 {
				pipedObject = pipedObjects[0]
			}

			var userResult, userErr = funcDef.Func(pipedObject, positionals, keywords)
			if userErr != nil {
				panic(interp.newError(node.Pos(), "%s", userErr))
			}
			return userResult
		case *Function:
			var callLen = len(node.Arg.Exprs)
			var declLen = len(funcDef.Signature.Args)
			if callLen != declLen {
				panic(interp.newError(node.Arg.Pos(), "%s takes %d args, call is sending %d arg", funcDef, declLen, callLen))
			}

			want := len(funcDef.Signature.PipedArgs)
			got := len(node.PipedArg.Exprs)
			if want != got {
				panic(interp.newError(node.Arg.Pos(), "%s takes %d piped args, call is sending %v", funcDef, want, got))
			}

			var positionalArgNames []string
			for _, param := range funcDef.Signature.Args {
				positionalArgNames = append(positionalArgNames, param.Name)
			}

			var pipedArgNames []string
			for _, param := range funcDef.Signature.PipedArgs {
				pipedArgNames = append(pipedArgNames, param.Name)
			}

			var newEnv = env.Global().NewScope()

			var pipedIdx = 0
			for _, arg := range node.PipedArg.Exprs {
				var obj = interp.eval(arg, env) // here we should use the old env
				var name = pipedArgNames[0]
				interp.mustSet(newEnv, name, obj)
				pipedIdx += 1
			}

			var positionalIdx = 0
			for _, arg := range node.Arg.Exprs {
				var obj = interp.eval(arg, env) // here we should use the old env
				var name = positionalArgNames[positionalIdx]
				interp.mustSet(newEnv, name, obj)
				positionalIdx += 1
			}

			// ast.BlockStmt
			// return interp.eval(funcDef.Body, newEnv)

			var result = interp.eval(funcDef.Body, newEnv)
			switch result := result.(type) {
			case nil:
				return nil
			case *ReturnStmt:
				return result.Expr
			default:
				panic(interp.newError(node.Pos(), "unexpected return type %T", result))
			}
		default:
			panic(interp.newError(node.Pos(), "unsupported function type %T", funcDef))
		}
	case *ast.ReturnStmt:
		return &ReturnStmt{Expr: interp.eval(node.Expr, env)}
	case *ast.Ident:
		val, err := env.Get(node.Name)
		if err != nil {
			// try builtins:
			if b, ok := interp.builtins()[node.Name]; ok {
				return b
			}
			panic(interp.newError(node.Pos(), "%q is missing: %v", node.Name, err))
		}
		return val
	case *ast.FuncDecl:
		interp.mustSet(
			env,
			node.Name.Name,
			&Function{
				Name:      node.Name.Name,
				Signature: node.Signature,
				Body:      node.Body,
			},
		)
		// Old note: We return nil, because function declaration in this
		// language are not expressions atm. If in the futre you want to
		// support anonymous function, return the Function here.
		return nil
	case *ast.Integer:
		return &Integer{Value: node.Value}
	case *ast.Float:
		return &Float{Value: node.Value}
	case *ast.BinaryExpr:
		switch node.Op {
		case token.REG:
			var x = interp.eval(node.X, env)
			var y = interp.eval(node.Y, env)
			// TODO: check regexp compilation statically on a best effort basis
			var re = regexp.MustCompile(y.(*String).AsSingle)
			return &Boolean{Value: re.MatchString(x.(*String).AsSingle)}
		case token.NREG:
			var x = interp.eval(node.X, env)
			var y = interp.eval(node.Y, env)
			// TODO: check regexp compilation statically on a best effort basis
			var re = regexp.MustCompile(y.(*String).AsSingle)
			return &Boolean{Value: re.MatchString(x.(*String).AsSingle)}
		case token.EQL:
			var x = interp.eval(node.X, env)
			var y = interp.eval(node.Y, env)
			return &Boolean{Value: x.GoValue() == y.GoValue()}
		default:
			panic(interp.newError(node.Pos(), "unsupported binary operator %q", node.Op))
		}
	case *ast.BlockStmt:
		for _, stmt := range node.Statements {
			var result = interp.eval(stmt, env)
			switch result := result.(type) {
			case *ReturnStmt:
				// TODO: statically check unreachable code
				return result // result.Expr
				// return result.Expr
			}
		}
		return nil
	case *ast.IfStmt:
		var cond = interp.eval(node.Cond, env)
		if cond.(*Boolean).Value == true {
			return interp.eval(node.Body, env)
		} else if node.Else != nil {
			return interp.eval(node.Else, env)
		}
		return nil
	case *ast.LetDecl:
		interp.mustSet(env, node.Name.Name, interp.eval(node.Rhs, env))
		return nil
	case *ast.String:
		var envFunc = func(name string) interface{} {
			val, err := env.Get(name)
			if err != nil {
				panic(interp.newError(node.Pos(), "%q is missing: %v", name, err))
			}
			return val
		}

		var rendered, err = expander.EncodeToString(node.Root, envFunc)
		if err != nil {
			panic(interp.newError(node.Pos(), "failed to render string: %v", err))
		}
		if interp.Verbose {
			// fmt.Fprintf(interp.Stderr, "+ string %q\n", rendered)
		}

		var words, encodeErr = expander.EncodeToCmdArgs(node.Root, envFunc)
		if encodeErr != nil {
			panic(interp.newError(node.Pos(), "failed to create args: %s", encodeErr))
		}
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
	if pos == NoPos {
		return InterpError{fmt.Errorf(f, args...)}
	}
	var lines = interp.parser.MarkAt(pos, fmt.Sprintf(f, args...), false)
	return InterpError{fmt.Errorf("%s", strings.Join(lines, "\n"))}
}
