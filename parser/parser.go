package parser

import (
	"fmt"
	"io"

	"github.com/siadat/well/scanner"
)

type CmdParser struct {
	src io.Reader

	root *ContainerNode

	tok scanner.CmdTokenType
	lit string
}

func NewParser() *CmdParser {
	return &CmdParser{}
}

func (p *CmdParser) reset() {
	// noop atm
}

func (p *CmdParser) Parse(src io.Reader) (*ContainerNode, error) {
	p.reset()
	p.next()

	var cnt, err = p.parseContainer()
	p.root = cnt

	return p.root, err
}

func (p *CmdParser) next() {
	// TODO
}

func (p *CmdParser) parseContainer() (*ContainerNode, error) {
	return nil, fmt.Errorf("TODO")
}
