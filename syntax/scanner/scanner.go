package scanner

import (
	"fmt"
	"io"
	"strings"

	"github.com/siadat/well/syntax/token"
)

type Scanner struct {
	src []rune

	currToken    Token
	currRune     rune
	position     int
	readPosition int

	nextRune       rune
	skipWhitespace bool
	skipComment    bool

	debug bool
}

type Pos int

type Token struct {
	Typ token.Token
	Lit string
	Pos Pos
}

func (t Token) String() string {
	return fmt.Sprintf("%s(%s) at %d", t.Typ, token.LiteralStringer(t.Lit), t.Pos)
}

func (t Token) ShortString() string {
	if t.Lit == token.AnyLit {
		return fmt.Sprintf("%s at %d", t.Typ, t.Pos)
	}
	return fmt.Sprintf("%s(%s) at %d", t.Typ, token.LiteralStringer(t.Lit), t.Pos)
}

func NewScanner(src io.Reader) *Scanner {
	var s, err = io.ReadAll(src)
	if err != nil {
		panic(err)
	}
	var scanner = &Scanner{
		src: []rune(string(s)),
	}
	scanner.readRune()
	return scanner
}

const EndOfInput = 0

func (p *Scanner) SetDebug(debug bool) {
	p.debug = debug
}

func (s *Scanner) SetSkipWhitespace(v bool) {
	s.skipWhitespace = v
}

func (s *Scanner) SetSkipComment(v bool) {
	s.skipComment = v
}

func (s *Scanner) readRune() {
	if s.readPosition >= len(s.src) {
		s.currRune = EndOfInput
	} else {
		s.currRune = s.src[s.readPosition]
	}
	s.position = s.readPosition
	s.readPosition += 1

	// nextRune
	if s.readPosition >= len(s.src) {
		s.nextRune = EndOfInput
	} else {
		s.nextRune = s.src[s.readPosition]
	}
}

func (s *Scanner) NextToken() (Token, error) {
	var t, err = s.nextToken()
	s.currToken = t
	if err != nil {
		s.readRune() // skip
	}
	return t, err
}

func (s *Scanner) isEndOfExpression() bool {
	var ch = s.currRune
	if ch == ' ' || ch == ',' || ch == ']' || ch == '\n' || ch == EndOfInput {
		return true
	}
	// ".."
	if ch == '.' && s.nextRune == '.' {
		return true
	}
	return false
}

func (s *Scanner) Eof() bool {
	return s.currToken.Typ == token.EOF
}

func (s *Scanner) CurrPosition() int {
	return s.position
}

func (s *Scanner) CurrToken() Token {
	return s.currToken
}

func (s *Scanner) nextToken() (Token, error) {
	if s.debug {
		s.PrintCursor("debug")
	}
	var start = s.position

	switch s.currRune {
	case '\n', '\r':
		var tok = Token{token.NEWLINE, fmt.Sprintf("%c", s.currRune), Pos(start)}
		s.readRune()
		return tok, nil
	case '|':
		var tok = Token{token.PIPE, fmt.Sprintf("%c", s.currRune), Pos(start)}
		s.readRune()
		return tok, nil
	case '[':
		var tok = Token{token.LBRACK, fmt.Sprintf("%c", s.currRune), Pos(start)}
		s.readRune()
		return tok, nil
	case ']':
		var tok = Token{token.RBRACK, fmt.Sprintf("%c", s.currRune), Pos(start)}
		s.readRune()
		return tok, nil
	case '(':
		var tok = Token{token.LPAREN, fmt.Sprintf("%c", s.currRune), Pos(start)}
		s.readRune()
		return tok, nil
	case ')':
		var tok = Token{token.RPAREN, fmt.Sprintf("%c", s.currRune), Pos(start)}
		s.readRune()
		return tok, nil
	case ',':
		var tok = Token{token.COMMA, fmt.Sprintf("%c", s.currRune), Pos(start)}
		s.readRune()
		return tok, nil
	case '%':
		var tok = Token{token.REM, fmt.Sprintf("%c", s.currRune), Pos(start)}
		s.readRune()
		return tok, nil
	case '*':
		var tok = Token{token.MUL, fmt.Sprintf("%c", s.currRune), Pos(start)}
		s.readRune()
		return tok, nil
	case '/':
		var tok = Token{token.QUO, fmt.Sprintf("%c", s.currRune), Pos(start)}
		s.readRune()
		return tok, nil
	case '.':
		var tok = Token{token.PERIOD, fmt.Sprintf("%c", s.currRune), Pos(start)}
		s.readRune()
		return tok, nil
	case ':':
		var tok = Token{token.COLON, fmt.Sprintf("%c", s.currRune), Pos(start)}
		s.readRune()
		return tok, nil
	case '{':
		var tok = Token{token.LBRACE, fmt.Sprintf("%c", s.currRune), Pos(start)}
		s.readRune()
		return tok, nil
	case '}':
		var tok = Token{token.RBRACE, fmt.Sprintf("%c", s.currRune), Pos(start)}
		s.readRune()
		return tok, nil
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return s.readNumber()
	case '-':
		// either a sign (e.g. '-1', or a sub e.g. '1 - -2')
		var tok = Token{token.SUB, fmt.Sprintf("%c", s.currRune), Pos(start)}
		s.readRune()
		return tok, nil
	case '+':
		// either a sign (e.g. '+1', or a sum e.g. '1 + +2')
		var tok = Token{token.ADD, fmt.Sprintf("%c", s.currRune), Pos(start)}
		s.readRune()
		return tok, nil
	case '=':
		// this can be '=' or '=='
		if s.nextRune == '=' {
			var tok = Token{token.EQL, fmt.Sprintf("%c%c", s.currRune, s.nextRune), Pos(start)}
			s.readRune()
			s.readRune()
			return tok, nil
		} else {
			var tok = Token{token.ASSIGN, fmt.Sprintf("%c", s.currRune), Pos(start)}
			s.readRune()
			return tok, nil
		}
	case '>':
		// this can be '>' or '>='
		if s.nextRune == '=' {
			var tok = Token{token.GEQ, fmt.Sprintf("%c%c", s.currRune, s.nextRune), Pos(start)}
			s.readRune()
			s.readRune()
			return tok, nil
		} else {
			var tok = Token{token.GTR, fmt.Sprintf("%c", s.currRune), Pos(start)}
			s.readRune()
			return tok, nil
		}
	case '<':
		// this can be '<' or '<='
		if s.nextRune == '=' {
			var tok = Token{token.LEQ, fmt.Sprintf("%c%c", s.currRune, s.nextRune), Pos(start)}
			s.readRune()
			s.readRune()
			return tok, nil
		} else {
			var tok = Token{token.LSS, fmt.Sprintf("%c", s.currRune), Pos(start)}
			s.readRune()
			return tok, nil
		}
	case '!':
		// this can be '!' or '!='
		if s.nextRune == '=' {
			var tok = Token{token.NEQ, fmt.Sprintf("%c%c", s.currRune, s.nextRune), Pos(start)}
			s.readRune()
			s.readRune()
			return tok, nil
		} else {
			var tok = Token{token.NOT, fmt.Sprintf("%c", s.currRune), Pos(start)}
			s.readRune()
			return tok, nil
		}
	case '#':
		if s.skipComment {
			var ret, err = s.readComment()
			if err != nil {
				return ret, err
			}
			return s.nextToken()
		} else {
			return s.readComment()
		}
	case '"':
		return s.readString(s.currRune)
	case ' ', '\t':
		if s.skipWhitespace {
			var ret, err = s.readWhitespace()
			if err != nil {
				return ret, err
			}
			return s.nextToken()
		} else {
			return s.readWhitespace()
		}
	case EndOfInput:
		var tok = Token{token.EOF, "", Pos(start)}
		s.readRune()
		return tok, nil
	default:
		return s.readIdentifier()
	}
}

func (s *Scanner) readIdentifier() (Token, error) {
	if s.isIdentifierPartFirst() {
		return s.readIdentifierUnquoted()
	} else {
		return Token{
			token.ILLEGAL,
			fmt.Sprintf("%c", s.currRune),
			Pos(s.position),
		}, fmt.Errorf("unexpected identifier character %c", s.currRune)
	}
}

func (s *Scanner) readString(ender rune) (Token, error) {
	var position = s.position
	s.readRune() // skip opener, i.e. "

	for {
		switch s.currRune {
		case '\\':
			s.readRune() // skip \
			s.readRune() // skip the char after \
		case ender:
			s.readRune()
			return Token{
				token.STRING,
				string(s.src[position:s.position]),
				Pos(position),
			}, nil
		default:
			s.readRune()
		}
	}
}

func (s *Scanner) readComment() (Token, error) {
	var position = s.position
For:
	for {
		switch s.currRune {
		case '\n', EndOfInput:
			break For
		default:
			s.readRune()
		}
	}
	return Token{
		token.COMMENT,
		string(s.src[position:s.position]),
		Pos(position),
	}, nil
}

func (s *Scanner) readWhitespace() (Token, error) {
	var position = s.position
	for s.isWhitespace() {
		s.readRune()
	}
	return Token{
		token.WHITESPACE,
		string(s.src[position:s.position]),
		Pos(position),
	}, nil
}

func (s *Scanner) isWhitespace() bool {
	var ch = s.currRune
	return ' ' == ch || '\t' == ch
}

func (s *Scanner) readIdentifierUnquoted() (Token, error) {
	var position = s.position
	s.readRune()
	for s.isIdentifierMiddle() {
		s.readRune()
	}
	var lit = string(s.src[position:s.position])
	return Token{token.IDENTIFIER, lit, Pos(position)}, nil
}

func (s *Scanner) isIdentifierPartFirst() bool {
	var ch = s.currRune
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch == '$'
}

func (s *Scanner) isIdentifierMiddle() bool {
	var ch = s.currRune
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || (ch >= '0' && ch <= '9')
}

func isNumerical(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func (s *Scanner) readNumber() (Token, error) {
	var position = s.position

	if s.currRune == '-' || s.currRune == '+' {
		s.readRune() // skip sign
	}

	var isFloat = false
	for {
		switch s.currRune {
		case '.':
			isFloat = true
			s.readRune()
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			s.readRune()
		default:
			if isFloat {
				return Token{
					token.FLOAT,
					string(s.src[position:s.position]),
					Pos(position),
				}, nil
			} else {
				return Token{
					token.INTEGER,
					string(s.src[position:s.position]),
					Pos(position),
				}, nil
			}
		}
	}
}

func (s *Scanner) PrintCursor(layout string, args ...interface{}) {
	var lines = strings.Split(string(s.src), "\n")
	var b strings.Builder
	var line, column = s.getCurrPosition()

	var ch string
	if s.currRune == EndOfInput {
		ch = "EndOfInput"
	} else {
		ch = fmt.Sprintf("%q", s.currRune)
	}

	var prefix = fmt.Sprintf(layout, args...)
	var linestr = strings.ReplaceAll(lines[line], "\t", "␣")
	if line == len(lines)-1 {
		linestr = linestr + "·"
	} else {
		linestr = linestr + "⏎"
	}
	b.WriteString(fmt.Sprintf("%s %s\n", prefix, linestr))
	// b.WriteString(fmt.Sprintf("%s %s│\n", prefix, strings.Repeat(" ", column)))
	b.WriteString(fmt.Sprintf("%s %s╰─src[%d]=%s currToken=%s\n", prefix, strings.Repeat(" ", column), s.position, ch, s.currToken))
	// b.WriteString(fmt.Sprintf("%s %s  currToken=%s\n", prefix, strings.Repeat(" ", column), s.currToken))
	if s.position == 0 {
		var header = fmt.Sprintf("Parsing %q", string(s.src))
		if len(header) > 80 {
			header = header[:80] + "..."
		}
		fmt.Printf("%s %s\n", strings.Repeat(" ", len(prefix)), strings.Repeat("=", len(header)))
		fmt.Printf("%s %s\n", strings.Repeat(" ", len(prefix)), header)
		fmt.Printf("%s %s\n", strings.Repeat(" ", len(prefix)), strings.Repeat("=", len(header)))
	}
	fmt.Print(b.String())
}

func (s *Scanner) getCurrPosition() (int, int) {
	var line = 0
	var column = 0
	for i := 0; i < s.position && i < len(s.src); i++ {
		if s.src[i] == '\n' {
			line += 1
			column = 0
		} else {
			column += 1
		}
	}
	return line, column
}
