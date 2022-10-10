package parser

import "github.com/siadat/well/scanner"

// Arg is an arg node
type Arg struct {
	Lit string
}

// Whs is a whitespace node
type Whs struct {
	Lit string
}

// Var is a var node
type Var struct {
	Name string
	Opts string
}

type ContainerNode struct {
	Type  scanner.CmdTokenType
	Items []CmdNode
}

type CmdNode interface {
	node()
}

func (Arg) node()           {}
func (ContainerNode) node() {}
func (Var) node()           {}
func (Whs) node()           {}
