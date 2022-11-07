package token

import (
	"fmt"
	"strconv"
)

type Token int

const (
	AnyTyp Token  = -1
	AnyLit string = ""
)

const (
	ILLEGAL Token = iota
	EOF
	WHITESPACE
	NEWLINE
	PIPE
	COMMENT

	literal_beg
	IDENTIFIER // main
	INTEGER    // 12345
	FLOAT      // 123.45
	STRING     // "abc"
	literal_end

	operator_beg
	LPAREN // (
	RPAREN // )
	LBRACK // [
	RBRACK // ]
	LBRACE // {
	RBRACE // }
	COLON  // :
	COMMA  // ,
	PERIOD // .
	ADD    // +
	SUB    // -
	NOT    // !
	MUL    // *
	QUO    // /
	REM    // %
	EQL    // ==
	REG    // ~~
	LSS    // <
	GTR    // >
	ASSIGN // =
	NEQ    // !=
	NREG   // !~
	LEQ    // <=
	GEQ    // >=
	operator_end

	keyword_beg
	FUNC
	RETURN
	LET
	keyword_end
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",

	EOF:        "EOF",
	WHITESPACE: "WHITESPACE",
	NEWLINE:    "NEWLINE",
	PIPE:       "PIPE",
	COMMENT:    "COMMENT",

	IDENTIFIER: "IDENTIFIER",
	INTEGER:    "INTEGER",
	FLOAT:      "FLOAT",
	STRING:     "STRING",

	ADD:    "ADD",
	SUB:    "SUB",
	MUL:    "MUL",
	QUO:    "QUO",
	REM:    "REM",
	LPAREN: "LPAREN",
	LBRACK: "LBRACK",
	LBRACE: "LBRACE",
	COMMA:  "COMMA",
	PERIOD: "PERIOD",
	RPAREN: "RPAREN",
	RBRACK: "RBRACK",
	RBRACE: "RBRACE",

	FUNC:   "func",
	RETURN: "return",
	LET:    "let", //+
}

type Precedence int

var LowestPrecedence Precedence = 0
var Precedences = map[Token]Precedence{
	ADD: 1,
	SUB: 1,

	REG:  1, // ~~
	NREG: 1, // !~

	MUL: 2,
	QUO: 2,

	LPAREN: 3,
}

func (tok Token) String() string {
	if tok == AnyTyp {
		return ":AnyTyp:"
	}
	var s = ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

type LiteralStringer string

func (lit LiteralStringer) String() string {
	if string(lit) == AnyLit {
		return ":AnyLit:"
	}
	return fmt.Sprintf("%q", string(lit))
}
