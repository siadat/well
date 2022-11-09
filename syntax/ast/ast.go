package ast

import (
	"github.com/siadat/well/syntax/scanner"
	strs_parser "github.com/siadat/well/syntax/strs/parser"
	"github.com/siadat/well/syntax/token"
)

type Root struct {
	Decls []Decl
}

type FuncSignatureArg struct {
	Name string
	Type string
}

type FuncSignature struct {
	Args     []FuncSignatureArg
	RetTypes []string

	Position scanner.Pos
}

type LetDecl struct {
	Name *Ident
	Rhs  Expr

	Position scanner.Pos
}

type FuncDecl struct {
	Name      *Ident
	Signature *FuncSignature
	Body      *BlockStmt

	Position scanner.Pos
}

type ExprStmt struct {
	X Expr

	Position scanner.Pos
}

type ReturnStmt struct {
	Expr Expr

	Position scanner.Pos
}

type IfStmt struct {
	Cond Expr
	Body *BlockStmt
	Else Stmt

	Position scanner.Pos
}

type BlockStmt struct {
	Statements []Stmt
	Position   scanner.Pos
}

type BinaryExpr struct {
	X  Expr
	Y  Expr
	Op token.Token

	Position scanner.Pos
}

type UnaryExpr struct {
	X  Expr
	Op token.Token

	Position scanner.Pos
}

type ParenExpr struct {
	Exprs []Expr

	Position scanner.Pos
}

type CallExpr struct {
	Fun Expr
	Arg *ParenExpr

	Position scanner.Pos
}

type AssignExpr struct {
	Name string
	Expr Expr

	Position scanner.Pos
}

type File struct {
}

type Ident struct {
	Name string

	Position scanner.Pos
}

type String struct {
	Root      *strs_parser.Root
	StringLit string

	Position scanner.Pos
}

type Integer struct {
	Value int

	Position scanner.Pos
}

type Float struct {
	Value float64

	Position scanner.Pos
}

type Node interface {
	node()
	Pos() scanner.Pos
}

type Expr interface {
	Node
	expr()
}

type Stmt interface {
	Node
	// stmt() // TODO: add this?
}

type Decl interface {
	Stmt
	decl()
}

func (*Root) node()          {}
func (*LetDecl) node()       {}
func (*FuncDecl) node()      {}
func (*FuncSignature) node() {}
func (*ExprStmt) node()      {}
func (*ReturnStmt) node()    {}
func (*IfStmt) node()        {}
func (*BlockStmt) node()     {}
func (*Ident) node()         {}
func (*Integer) node()       {}
func (*String) node()        {}
func (*Float) node()         {}
func (*BinaryExpr) node()    {}
func (*UnaryExpr) node()     {}
func (*ParenExpr) node()     {}
func (*AssignExpr) node()    {}
func (*File) node()          {}
func (*CallExpr) node()      {}

func (e *Root) Pos() scanner.Pos          { return -1 }
func (e *LetDecl) Pos() scanner.Pos       { return e.Position }
func (e *FuncDecl) Pos() scanner.Pos      { return e.Position }
func (e *FuncSignature) Pos() scanner.Pos { return e.Position }
func (e *ExprStmt) Pos() scanner.Pos      { return e.Position }
func (e *ReturnStmt) Pos() scanner.Pos    { return e.Position }
func (e *IfStmt) Pos() scanner.Pos        { return e.Position }
func (e *BlockStmt) Pos() scanner.Pos     { return e.Position }
func (e *Ident) Pos() scanner.Pos         { return e.Position }
func (e *Integer) Pos() scanner.Pos       { return e.Position }
func (e *String) Pos() scanner.Pos        { return e.Position }
func (e *Float) Pos() scanner.Pos         { return e.Position }
func (e *BinaryExpr) Pos() scanner.Pos    { return e.Position }
func (e *UnaryExpr) Pos() scanner.Pos     { return e.Position }
func (e *ParenExpr) Pos() scanner.Pos     { return e.Position }
func (e *AssignExpr) Pos() scanner.Pos    { return e.Position }
func (e *File) Pos() scanner.Pos          { return -1 }
func (e *CallExpr) Pos() scanner.Pos      { return e.Position }

func (*Ident) expr()      {}
func (*Integer) expr()    {}
func (*String) expr()     {}
func (*Float) expr()      {}
func (*BinaryExpr) expr() {}
func (*UnaryExpr) expr()  {}
func (*ParenExpr) expr()  {}
func (*AssignExpr) expr() {}
func (*File) expr()       {}
func (*CallExpr) expr()   {}

func (*LetDecl) decl()  {}
func (*FuncDecl) decl() {}

func (*LetDecl) stmt()    {}
func (*ExprStmt) stmt()   {}
func (*ReturnStmt) stmt() {}
func (*IfStmt) stmt()     {}
func (*BlockStmt) stmt()  {}
