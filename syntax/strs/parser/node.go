package parser

import (
	"bytes"
	"fmt"

	"github.com/siadat/well/syntax/strs/scanner"
)

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
	String() string
}

func (Wrd) node()           {}
func (ContainerNode) node() {}
func (Root) node()          {}
func (Var) node()           {}
func (Whs) node()           {}

func (w Wrd) String() string {
	switch w.Lit {
	case "«", "»", "‹", "›", "$":
		return `\` + w.Lit
	default:
		return w.Lit
	}
}

func (c ContainerNode) String() string {
	var buf bytes.Buffer

	for _, item := range c.Items {
		fmt.Fprintf(&buf, item.String())
	}

	switch c.Type {
	case scanner.DOUBLE_QUOTE: // "
		return `"` + buf.String() + `"`
	case scanner.SINGLE_QUOTE: // "
		return `'` + buf.String() + `'`
	case scanner.LDOUBLE_GUILLEMET: // «
		return `«` + buf.String() + `»`
	case scanner.LSINGLE_GUILLEMET: // ‹
		return `‹` + buf.String() + `›`
	default:
		panic(fmt.Sprintf("unsupported container type %s", c.Type))
	}
}

func (r Root) String() string {
	var buf bytes.Buffer

	for _, item := range r.Items {
		fmt.Fprintf(&buf, item.String())
	}
	return `"` + buf.String() + `"`
}

func (v Var) String() string {
	if v.Opts == "" {
		return fmt.Sprintf("${%s}", v.Name)
	}
	return fmt.Sprintf("${%s:%%%s}", v.Name, v.Opts)
}

func (w Whs) String() string {
	return w.Lit
}
