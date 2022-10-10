package parser

import (
	"fmt"
	"io"
	"strings"

	"github.com/siadat/well/scanner"
)

type CmdParser struct {
	src     io.Reader
	root    *Root
	scanner *scanner.CmdScanner
}

func NewParser() *CmdParser {
	return &CmdParser{}
}

func (p *CmdParser) Parse(src io.Reader) (*Root, error) {
	p.scanner = scanner.NewScanner(src)
	var nodes, err = p.parseContainerNodes(0, scanner.EOF)
	p.root = &Root{
		Items: nodes,
	}

	return p.root, err
}

func (p *CmdParser) parseContainerNodes(indent int, until scanner.CmdTokenType) ([]CmdNode, error) {
	var nodes []CmdNode
	for {
		var t, err = p.scanner.NextToken()
		if err != nil {
			return nil, err
		}
		if t.Typ == until {
			return nodes, nil
		}
		var n, err2 = p.tokenToNode(indent, until, t)
		if err2 != nil {
			return nil, err2
		}
		if n != nil {
			nodes = append(nodes, n)
		}
		if p.scanner.Eof() {
			return nodes, nil
		}
	}
}

func (p *CmdParser) tokenToNode(indent int, until scanner.CmdTokenType, t scanner.CmdToken) (CmdNode, error) {
	switch t.Typ {
	case scanner.EOF:
		return nil, nil
	case scanner.WORD:
		return Wrd{Lit: t.Lit}, nil
	case scanner.SPACE:
		return Whs{Lit: t.Lit}, nil
	case scanner.ARG:
		var segments = strings.SplitN(t.Lit, ":", 2)
		var name, flags string
		if len(segments) == 2 {
			name, flags = segments[0], segments[1]
		} else {
			name = t.Lit
		}
		return Var{Name: name, Opts: flags}, nil
	case scanner.SINGLE_QUOTE, scanner.DOUBLE_QUOTE,
		scanner.LDOUBLE_GUILLEMET, scanner.LSINGLE_GUILLEMET:

		var container = ContainerNode{
			Type: t.Typ,
		}
		var right = scanner.GetRight(t.Typ)
		var nodes, err = p.parseContainerNodes(indent+1, right)
		container.Items = nodes
		if err != nil {
			return nil, err // TODO: return nil?
		}
		return container, nil
	default:
		return nil, fmt.Errorf("unsupported token %s", t)
	}
}
