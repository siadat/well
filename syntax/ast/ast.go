package ast

import (
	"github.com/siadat/well/syntax/scanner"
	"github.com/siadat/well/syntax/token"
)

type Root struct {
	Decls []Decl
}

type FuncSignature struct {
	ArgNames []string
	ArgTypes []string
	RetTypes []string
}

type LetDecl struct {
	Name string
	Rhs  Expr
}

type FuncDecl struct {
	Name       string
	Signature  FuncSignature
	Statements []Stmt
}

type ExprStmt struct {
	X Expr
}

type BinaryExpr struct {
	X  Expr
	Y  Expr
	Op token.Token
}

type UnaryExpr struct {
	X  Expr
	Op token.Token
}

type ParenExpr struct {
	X Expr
}

type CallExpr struct {
	Fun Expr
	Arg Expr
}

type AssignExpr struct {
	Name string
	Expr Expr
}

type ExprList struct {
	Items []Expr
}

type File struct {
	// TODO
}

type Ident struct {
	Name string
	Pos  scanner.Pos
}

type String struct {
	Value string
}

type Integer struct {
	Value int
}

type Float struct {
	Value float64
}

type Node interface {
	node()
}

type Expr interface {
	node()
	expr()
}

type Stmt interface {
	node()
}

type Decl interface {
	Stmt
	decl()
}

func (Root) node()          {}
func (LetDecl) node()       {}
func (FuncDecl) node()      {}
func (ExprList) node()      {}
func (FuncSignature) node() {}
func (ExprStmt) node()      {}

func (LetDecl) decl()  {}
func (FuncDecl) decl() {}

func (LetDecl) stmt()  {}
func (ExprStmt) stmt() {}

func (Ident) node()      {}
func (Integer) node()    {}
func (String) node()     {}
func (Float) node()      {}
func (BinaryExpr) node() {}
func (UnaryExpr) node()  {}
func (ParenExpr) node()  {}
func (AssignExpr) node() {}
func (File) node()       {}
func (CallExpr) node()   {}

func (Ident) expr()      {}
func (Integer) expr()    {}
func (String) expr()     {}
func (Float) expr()      {}
func (BinaryExpr) expr() {}
func (UnaryExpr) expr()  {}
func (ParenExpr) expr()  {}
func (AssignExpr) expr() {}
func (File) expr()       {}
func (CallExpr) expr()   {}
