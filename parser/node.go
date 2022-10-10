package parser

import "github.com/siadat/well/scanner"

// Wrd is an arg node
type Wrd struct {
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

type Root struct {
	Items []CmdNode
}

type CmdNode interface {
	node()
}

func (Wrd) node()           {}
func (ContainerNode) node() {}
func (Root) node()          {}
func (Var) node()           {}
func (Whs) node()           {}
