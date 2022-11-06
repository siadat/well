package interpreter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/siadat/well/erroring"
	"github.com/siadat/well/piper"
	"github.com/siadat/well/syntax/ast"
	"github.com/siadat/well/syntax/parser"
	"github.com/siadat/well/syntax/scanner"
	"github.com/siadat/well/syntax/strs/expander"
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
	var builtinsSlice = []*Builtin{
		{
			"external", func(posArgs []Object, kvArgs map[string]Object) (Object, error) {
				var args []string
				for _, a := range posArgs {
					var arg = a.(*String)
					// TODO: this is already parsed, but
					// piper.External is also parsing it.
					// Change piper.External to accept an
					// already parsed str.
					args = append(args, arg.AsSingle)
				}
				piper.External(args...).Read(interp.Stdout, interp.Stderr)
				return nil, nil
			},
		},
		{
			"external_json", func(posArgs []Object, kvArgs map[string]Object) (Object, error) {
				var arg = posArgs[0].(*String)
				if interp.Debug {
					fmt.Fprintf(interp.Stderr, "call external command %#v\n", arg)
				}
				var stdout bytes.Buffer
				var stderr bytes.Buffer
				var cmd = exec.Command(arg.AsArgs[0], arg.AsArgs[1:]...)
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr

				var err = cmd.Run()
				var retBuf bytes.Buffer
				var enc = json.NewEncoder(&retBuf)
				var encodeErr = enc.Encode(map[string]string{
					"stdout": stdout.String(),
					"stderr": stderr.String(),
				})
				if encodeErr != nil {
					return nil, fmt.Errorf("encoding json failed: %s", encodeErr)
				}
				if err != nil {
					return nil, fmt.Errorf("external command failed: %v, output:\n%s", err, strings.TrimSpace(retBuf.String()))
				}
				return &String{AsSingle: strings.TrimSpace(retBuf.String())}, nil
			},
		},
		{
			"external_capture", func(posArgs []Object, kvArgs map[string]Object) (Object, error) {
				var args []string
				for _, a := range posArgs {
					var arg = a.(*String)
					// TODO: this is already parsed, but
					// piper.External is also parsing it.
					// Change piper.External to accept an
					// already parsed str.
					args = append(args, arg.AsSingle)
				}
				var stdout bytes.Buffer
				piper.External(args...).Read(&stdout, interp.Stderr)
				return &String{AsSingle: stdout.String()}, nil
			},
		},
		{
			"println", func(posArgs []Object, kvArgs map[string]Object) (Object, error) {
				for i, arg := range posArgs {
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
			"print", func(posArgs []Object, kvArgs map[string]Object) (Object, error) {
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
			"read", func(posArgs []Object, kvArgs map[string]Object) (Object, error) {
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
			"read_regex", func(posArgs []Object, kvArgs map[string]Object) (Object, error) {
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
			"read_int", func(posArgs []Object, kvArgs map[string]Object) (Object, error) {
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
			"date", func(posArgs []Object, kvArgs map[string]Object) (Object, error) {
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
		fmt.Printf("DEBUG eval %T %+v\n", node, node)
	}
	switch node := node.(type) {
	case *ast.Root:
		for _, decl := range node.Decls {
			interp.eval(decl, env)
		}
		return interp.eval(&ast.CallExpr{
			Fun: &ast.Ident{Name: "main", Position: NoPos},
			Arg: &ast.ParenExpr{Exprs: nil},
		}, env)
	case *ast.ParenExpr:
		var objs []Object
		for _, expr := range node.Exprs {
			objs = append(objs, interp.eval(expr, env))
		}
		return &Paren{Objects: objs}
	case *ast.ExprStmt:
		return interp.eval(node.X, env)
	case *ast.CallExpr:
		funcDef := interp.eval(node.Fun, env)
		switch funcDef := funcDef.(type) {
		case *Builtin:
			var positionals []Object
			var keywords map[string]Object

			for _, arg := range node.Arg.Exprs {
				var obj = interp.eval(arg, env)
				positionals = append(positionals, obj)
			}

			var userResult, userErr = funcDef.Func(positionals, keywords)
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

			var positionalArgNames []string
			for _, param := range funcDef.Signature.Args {
				positionalArgNames = append(positionalArgNames, param.Name)
			}

			var positionalIdx = 0
			var newEnv = env.NewScope()
			for _, arg := range node.Arg.Exprs {
				var obj = interp.eval(arg, env)
				var name = positionalArgNames[positionalIdx]
				interp.mustSet(env, name, obj)
				positionalIdx += 1
			}

			// ast.BlockStmt
			// return interp.eval(funcDef.Body, newEnv)

			for _, stmt := range funcDef.Body {
				result := interp.eval(stmt, newEnv)
				switch result := result.(type) {
				case *ReturnStmt:
					return result.Expr
				}
			}
			return nil
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
		interp.mustSet(env, node.Name.Name, &Function{
			Name:      node.Name.Name,
			Signature: node.Signature,
			Body:      node.Statements,
			Env:       env,
		})
		// Old note: We return nil, because function declaration in this
		// language are not expressions atm. If in the futre you want to
		// support anonymous function, return the Function here.
		return nil
	case *ast.Integer:
		return &Integer{Value: node.Value}
	case *ast.Float:
		return &Float{Value: node.Value}
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
			fmt.Fprintf(interp.Stderr, "+%s\n", rendered)
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
