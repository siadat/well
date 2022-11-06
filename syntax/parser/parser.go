package parser

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/siadat/well/erroring"
	"github.com/siadat/well/syntax/ast"
	"github.com/siadat/well/syntax/scanner"
	strs_parser "github.com/siadat/well/syntax/strs/parser"
	"github.com/siadat/well/syntax/token"
)

type Parser struct {
	src     io.Reader
	root    *ast.Root
	scanner *scanner.Scanner
	debug   bool
}

type ParseError struct {
	err error
}

func (e ParseError) Error() string {
	return e.err.Error()
}

func (p *Parser) proceed() scanner.Token {
	var t, err = p.scanner.NextToken()
	p.checkErr(err)
	return t
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) MarkAt(at scanner.Pos, msg string, showWhitespaces bool) []string {
	return p.scanner.MarkAt(at, msg, showWhitespaces)
}

func (p *Parser) init(src io.Reader) error {
	p.scanner = scanner.NewScanner(src)
	p.scanner.SetSkipWhitespace(true)
	p.scanner.SetSkipComment(true)
	p.scanner.SetDebug(p.debug)

	var _, err = p.scanner.NextToken()
	return err
}

func (p *Parser) SetDebug(debug bool) {
	p.debug = debug
	if p.scanner != nil {
		p.scanner.SetDebug(debug)
	}
}

func (p *Parser) Parse(src io.Reader) (retRoot *ast.Root, retErr error) {
	if err := p.init(src); err != nil {
		return nil, err
	}

	defer func() {
		var err = recover()
		switch err := err.(type) {
		case nil:
			return
		case ParseError:
			// if p.debug { debug.PrintStack() }
			var lines = p.MarkAt(p.scanner.CurrToken().Pos, err.Error(), false)
			retErr = fmt.Errorf("%s", strings.Join(lines, "\n"))
		default:
			fmt.Printf("unexpected error while parsing: %s\n", err)
			erroring.PrintTrace()
		}
	}()
	retRoot = &ast.Root{Decls: p.parseDecls()}
	return
}

func (p *Parser) ParseExpr(src io.Reader) (retExpr ast.Expr, retErr error) {
	if err := p.init(src); err != nil {
		return nil, err
	}

	defer func() {
		var err = recover()
		switch err := err.(type) {
		case nil:
			return
		case ParseError:
			// if p.debug { debug.PrintStack() }
			var lines = p.MarkAt(p.scanner.CurrToken().Pos, err.Error(), false)
			retErr = fmt.Errorf("%s", strings.Join(lines, "\n"))
		default:
			fmt.Printf("unexpected error while parsing: %s\n", err)
			erroring.PrintTrace()
		}
	}()
	retExpr = p.parseExpr(nil, token.LowestPrecedence)
	return
}

func (p *Parser) parseDecls() []ast.Decl {
	var nodes []ast.Decl
	for {
		var node = p.parseDecl()
		if node != nil {
			nodes = append(nodes, node)
		}
		if p.scanner.Eof() {
			return nodes
		}
	}
}

func (p *Parser) parseDecl() ast.Decl {
	var t = p.scanner.CurrToken()

	switch t.Lit {
	case "let":
		return p.parseLetDecl()
	case "function":
		return p.parseFuncDecl()
	}

	switch t.Typ {
	case token.NEWLINE:
		p.proceed()
		return nil
	case token.EOF:
		return nil
	default:
		panic(ParseError{fmt.Errorf("failed to parse a decl, unexpected %s", t)})
	}
}

func (p *Parser) parsePrimaryExpr() ast.Expr {
	switch t := p.scanner.CurrToken(); t.Typ {
	case token.STRING:
		p.proceed()

		v, err := strconv.Unquote(t.Lit)
		if err != nil {
			if p.debug {
				fmt.Printf("failed to unquote %q\n", t.Lit)
			}
			panic(ParseError{fmt.Errorf("failed to unquote string: %v", err)})
		}
		var raw = t.Lit[0] == '`'
		return &ast.String{
			Root:     MustParseStr(v, raw, p.debug),
			Position: t.Pos,
		}
	case token.IDENTIFIER:
		p.proceed()

		return &ast.Ident{
			Name:     t.Lit,
			Position: t.Pos,
		}
	case token.ADD, token.SUB:
		var op = t.Typ
		// signed expression, e.g. -1 or +value
		p.proceed()

		switch t := p.scanner.CurrToken(); t.Typ {
		case token.INTEGER,
			token.FLOAT,
			token.IDENTIFIER,
			token.LPAREN:
			return &ast.UnaryExpr{
				X:        p.parsePrimaryExpr(),
				Op:       op,
				Position: t.Pos,
			}
		default:
			panic(ParseError{fmt.Errorf("expected integer or float, got %s", t)})
		}
	case token.INTEGER:
		p.proceed()

		var d, err = strconv.ParseInt(t.Lit, 10, 64)
		p.checkErr(err)
		return &ast.Integer{
			Value:    int(d),
			Position: t.Pos,
		}
	case token.FLOAT:
		p.proceed()

		var f, err = strconv.ParseFloat(t.Lit, 64)
		p.checkErr(err)
		return &ast.Float{
			Value:    f,
			Position: t.Pos,
		}
	case token.LPAREN:
		return p.parseParenExpr()
	default:
		panic(ParseError{fmt.Errorf("failed to parse primary expression, got %s", t)})
	}
}

func (p *Parser) parseParenExpr() *ast.ParenExpr {
	var pos = p.scanner.CurrToken().Pos
	var exprs []ast.Expr = parseCsvInParens(p, func(p *Parser) ast.Expr {
		var expr = p.parseExpr(nil, token.LowestPrecedence)
		return expr
	})

	return &ast.ParenExpr{
		Exprs:    exprs,
		Position: pos,
	}
}

func (p *Parser) checkErr(err error) {
	if err != nil {
		panic(ParseError{err})
	}
}

func (p *Parser) parseExpr(lhs ast.Expr, minPrec token.Precedence) ast.Expr {
	if lhs == nil {
		lhs = p.parsePrimaryExpr()
	}

	for {
		var tk = p.scanner.CurrToken()
		var pos = tk.Pos
		var prec, isStillExpr = token.Precedences[tk.Typ]
		if !isStillExpr {
			return lhs
		}
		if tk.Typ == token.EOF {
			return lhs
		}
		if prec < minPrec {
			return lhs
		}

		switch tk.Typ {
		case token.LPAREN:
			var paren = p.parseParenExpr()
			lhs = &ast.CallExpr{
				Fun:      lhs,
				Arg:      paren,
				Position: lhs.Pos(),
			}
		default:
			// other kinds
			p.proceed()

			var rhs = p.parseExpr(nil, prec)
			lhs = &ast.BinaryExpr{
				X:        lhs,
				Y:        rhs,
				Op:       tk.Typ,
				Position: pos,
			}
		}
	}
}

func (p *Parser) skipOptionalNewlines() {
	for {
		var t = p.scanner.CurrToken()
		if t.Typ == token.NEWLINE {
			p.proceed()
		} else {
			return
		}
	}
}

func (p *Parser) expect(typ token.Token, lit string) {
	var t = p.scanner.CurrToken()
	if t.Typ == typ && t.Lit == lit {
		return
	}
	panic(ParseError{fmt.Errorf("expected %q, got %s", lit, t)})
}

func (p *Parser) expectType(typ token.Token) scanner.Token {
	var t = p.scanner.CurrToken()
	if t.Typ == typ {
		return t
	}
	panic(ParseError{fmt.Errorf("expected %s, got %s", typ, t)})
}

func (p *Parser) parseFuncDecl() ast.Decl {
	var pos = p.scanner.CurrToken().Pos
	p.expect(token.IDENTIFIER, "function")
	p.proceed()

	var identPos = p.scanner.CurrToken().Pos
	var name = p.expectType(token.IDENTIFIER)
	p.proceed()

	var signature = p.parseFuncSignature()

	p.expect(token.LBRACE, "{")
	p.proceed()

	p.expectType(token.NEWLINE)
	p.proceed()

	// function body
	var stmts []ast.Stmt
	for {
		var t = p.scanner.CurrToken()
		if t.Typ == token.RBRACE {
			break
		}
		if t.Typ == token.EOF {
			break
		}
		if t.Typ == token.NEWLINE {
			p.proceed()
			continue
		}
		var stmt = p.parseStmt()
		stmts = append(stmts, stmt)
	}

	p.expect(token.RBRACE, "}")
	p.proceed()

	return &ast.FuncDecl{
		Name:       &ast.Ident{Name: name.Lit, Position: identPos},
		Signature:  &signature,
		Statements: stmts,
		Position:   pos,
	}
}

func (p *Parser) parseStmt() ast.Stmt {
	var t = p.scanner.CurrToken()

	switch t.Lit {
	case "let":
		return p.parseLetDecl()
	}

	switch t.Typ {
	case token.IDENTIFIER:
		var pos = p.scanner.CurrToken().Pos
		return &ast.ExprStmt{
			X:        p.parseExpr(nil, token.LowestPrecedence),
			Position: pos,
		}
	default:
		panic(ParseError{fmt.Errorf("failed to parse a stmt, unexpected %s", t)})
	}
}

func (p *Parser) parseFuncSignatureArg() ast.FuncSignatureArg {
	var name = p.expectType(token.IDENTIFIER)
	p.proceed()

	var typ = p.expectType(token.IDENTIFIER)
	p.proceed()

	return ast.FuncSignatureArg{
		Name: name.Lit,
		Type: typ.Lit,
	}
}

func parseCsvInParens[T any](p *Parser, itemParseFunc func(p *Parser) T) []T {
	p.expect(token.LPAREN, "(")
	p.proceed()
	var items []T
For:
	for {
		var tk = p.scanner.CurrToken()
		switch tk.Typ {
		case token.RPAREN:
			break For
		case token.NEWLINE:
			p.skipOptionalNewlines()
		case token.COMMA:
			// This allows multiple commas as in `(1, 2,,,)`, I don't care atm,
			// because there probably will be a formatter that removes them and
			// converts it to `(1, 2)`
			p.proceed()
		default:
			items = append(items, itemParseFunc(p))
		}
	}
	p.expect(token.RPAREN, ")")
	p.proceed()
	return items
}

func (p *Parser) parseFuncSignature() ast.FuncSignature {
	var pos = p.scanner.CurrToken().Pos
	var args []ast.FuncSignatureArg = parseCsvInParens(p, func(p *Parser) ast.FuncSignatureArg {
		return p.parseFuncSignatureArg()
	})

	var retTypes []string
	var t = p.scanner.CurrToken()
	if t.Typ == token.IDENTIFIER {
		var typ = p.expectType(token.IDENTIFIER)
		p.proceed()
		retTypes = append(retTypes, typ.Lit)
	} else if t.Typ == token.LPAREN {
		retTypes = parseCsvInParens(p, func(p *Parser) string {
			var typ = p.expectType(token.IDENTIFIER)
			p.proceed()
			return typ.Lit
		})
	}

	return ast.FuncSignature{
		Args:     args,
		RetTypes: retTypes,
		Position: pos,
	}
}

func MustParseStr(s string, raw bool, debug bool) *strs_parser.Root {
	if raw {
		return &strs_parser.Root{
			Items: []strs_parser.CmdNode{strs_parser.Wrd{Lit: s}},
		}
	}
	var p = strs_parser.NewParser()
	var root, err = p.Parse(strings.NewReader(s))
	if err != nil {
		if debug {
			fmt.Printf("failed parsing %q\n", s)
		}
		panic(ParseError{fmt.Errorf("failed to parse str %q: %w", s, err)})
	}
	return root
}

func (p *Parser) parseLetDecl() ast.Decl {
	var pos = p.scanner.CurrToken().Pos
	p.expect(token.IDENTIFIER, "let")
	p.proceed()

	var namePos = p.scanner.CurrToken().Pos
	var name = p.expectType(token.IDENTIFIER)
	p.proceed()

	p.expect(token.ASSIGN, "=")
	p.proceed()

	var rhs = p.parseExpr(nil, token.LowestPrecedence)

	return &ast.LetDecl{
		Name: &ast.Ident{
			Name:     name.Lit,
			Position: namePos,
		},
		Rhs:      rhs,
		Position: pos,
	}
}
