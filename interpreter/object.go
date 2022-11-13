package interpreter

import (
	"fmt"
	"io"

	"github.com/siadat/well/syntax/ast"
)

type Object interface {
	GoValue() interface{}
	String() string
	isObject()
}

type Paren struct {
	Objects []Object
}

type PipeStream struct {
	ReadCloser io.ReadCloser
}

// WriteCloser io.WriteCloser

type Integer struct {
	Value int
}

type Float struct {
	Value float64
}

type String struct {
	// TODO: refactor this, so ugly
	AsSingle string
	AsArgs   []string
}

type Boolean struct {
	Value bool
}

type ExtDecl struct {
	Name string
	Path string
}

type Function struct {
	Name      string
	Signature *ast.FuncSignature
	Body      *ast.BlockStmt
	// Env       Environment
}

type ReturnStmt struct {
	Expr Object
}

type Builtin struct {
	Name string
	Func func(Object, []Object, map[string]Object) (Object, error)
}

var NoValue = struct{}{}

func (i *Paren) String() string      { return fmt.Sprintf("%#v", i.Objects) }
func (i *PipeStream) String() string { return fmt.Sprintf("%#v", i.ReadCloser) }
func (i *Integer) String() string    { return fmt.Sprintf("%d", i.Value) }
func (i *Float) String() string      { return fmt.Sprintf("%f", i.Value) }
func (i *String) String() string     { return fmt.Sprintf("%s", i.AsSingle) }
func (i *Boolean) String() string    { return fmt.Sprintf("%v", i.Value) }
func (i *ExtDecl) String() string    { return fmt.Sprintf("external %s %v", i.Name, i.Path) }
func (i *Function) String() string   { return fmt.Sprintf("function %s", i.Name) }
func (i *ReturnStmt) String() string { return fmt.Sprintf("retrun %s", i.Expr.String()) }
func (i *Builtin) String() string    { return fmt.Sprintf("builtin %s", i.Name) }

func (i *Paren) GoValue() interface{}      { return i.Objects }
func (i *PipeStream) GoValue() interface{} { return nil /* internal? */ }
func (i *Integer) GoValue() interface{}    { return i.Value }
func (i *Float) GoValue() interface{}      { return i.Value }
func (i *String) GoValue() interface{}     { return i.AsSingle }
func (i *Boolean) GoValue() interface{}    { return i.Value }
func (i *ExtDecl) GoValue() interface{}    { return NoValue }
func (i *Function) GoValue() interface{}   { return NoValue }
func (i *ReturnStmt) GoValue() interface{} { return NoValue }
func (i *Builtin) GoValue() interface{}    { return NoValue }

func (i *Paren) isObject()      {}
func (i *PipeStream) isObject() {}
func (i *Integer) isObject()    {}
func (i *Float) isObject()      {}
func (i *String) isObject()     {}
func (i *Boolean) isObject()    {}
func (i *ExtDecl) isObject()    {}
func (i *Function) isObject()   {}
func (i *ReturnStmt) isObject() {}
func (i *Builtin) isObject()    {}
