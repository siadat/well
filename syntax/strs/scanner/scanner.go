package scanner

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// The most interesting thing about the scanner is that it includes whitespaces
// as well. Whether whitespaces are ignore or not is a decision made by the
// parser, not the scanner.

type CmdScanner struct {
	src          []rune
	currToken    CmdToken
	currRune     rune
	position     int
	readPosition int

	debug bool
}

type CmdToken struct {
	Typ CmdTokenType
	Lit string
}

func (t CmdToken) String() string {
	return fmt.Sprintf("%s(%q)", t.Typ, t.Lit)
}

func NewScanner(src io.Reader) *CmdScanner {
	var s, err = ioutil.ReadAll(src)
	if err != nil {
		panic(err)
	}
	var scanner = &CmdScanner{
		src: []rune(string(s)),
	}
	scanner.readRune()
	return scanner
}

func (s *CmdScanner) SetDebug(v bool) {
	s.debug = v
}

func (s *CmdScanner) Eof() bool {
	return s.currToken.Typ == EOF
}

func (s *CmdScanner) CurrToken() CmdToken {
	return s.currToken
}

func (s *CmdScanner) reset() {
	s.position = 0
	s.currRune = 0
}

func (s *CmdScanner) readRune() {
	if s.readPosition >= len(s.src) {
		s.currRune = 0
	} else {
		s.currRune = s.src[s.readPosition]
	}
	s.position = s.readPosition
	s.readPosition += 1
}

func (s *CmdScanner) isIdentifierPartFirst() bool {
	var ch = s.currRune
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func (s *CmdScanner) isIdentifierMiddle() bool {
	var ch = s.currRune
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || (ch >= '0' && ch <= '9')
}

func (s *CmdScanner) isWhitespace() bool {
	switch s.currRune {
	case ' ', '\t', '\n', '\r':
		return true
	}
	return false
}

func (s *CmdScanner) isWord() bool {
	switch s.currRune {
	case '\\',
		'\'', '"',
		'$',
		'«', '»',
		'‹', '›',
		' ', '\t', '\n', '\r',
		0:
		return false
	}
	return true
}

func (s *CmdScanner) readVariable() CmdToken {
	if curr := s.currRune; curr != '$' {
		s.readRune() // skip
		return CmdToken{
			Typ: ILLEGAL_TOKEN,
			Lit: fmt.Sprintf("expected $ got %c", curr),
		}
	}
	s.readRune()

	if curr := s.currRune; curr != '{' {
		s.readRune() // skip
		return CmdToken{
			Typ: ILLEGAL_TOKEN,
			Lit: fmt.Sprintf("expected { got %c", curr),
		}
	}
	s.readRune()

	if !s.isIdentifierPartFirst() {
		return CmdToken{
			Typ: ILLEGAL_TOKEN,
			Lit: fmt.Sprintf("expected identifier, got %c", s.currRune),
		}
	}
	var name = s.readIdentifier()

	if curr := s.currRune; curr == '}' {
		s.readRune()
		return CmdToken{
			Typ: ARG,
			Lit: name,
		}
	}

	if curr := s.currRune; curr != ':' {
		s.readRune() // skip
		return CmdToken{
			Typ: ILLEGAL_TOKEN,
			Lit: fmt.Sprintf("expected : got %c", curr),
		}
	}
	s.readRune()

	var flags = s.readVariableFlags()

	if curr := s.currRune; curr != '}' {
		s.readRune() // skip
		return CmdToken{
			Typ: ILLEGAL_TOKEN,
			Lit: fmt.Sprintf("expected } got %c", curr),
		}
	}
	s.readRune()

	return CmdToken{
		Typ: ARG,
		Lit: fmt.Sprintf("%s:%s", name, flags),
	}
}

func (s *CmdScanner) readVariableFlags() string {
	var position = s.position

	if want, curr := '%', s.currRune; curr != want {
		s.readRune() // skip
		return ""
	}
	s.readRune() // skip

	for s.currRune != '}' && s.currToken.Typ != EOF {
		s.readRune()
	}

	return string(s.src[position:s.position])
}

func (s *CmdScanner) readIdentifier() string {
	var position = s.position
	s.readRune()
	for s.isIdentifierMiddle() {
		s.readRune()
	}
	return string(s.src[position:s.position])
}

func (s *CmdScanner) readWhitespace() CmdToken {
	var position = s.position
	for s.isWhitespace() {
		s.readRune()
	}
	return CmdToken{
		Typ: SPACE,
		Lit: string(s.src[position:s.position]),
	}
}

func (s *CmdScanner) readWord() CmdToken {
	var position = s.position
	for s.isWord() {
		s.readRune()
	}
	return CmdToken{
		Typ: WORD,
		Lit: string(s.src[position:s.position]),
	}
}

func (s *CmdScanner) PrintCursor(layout string, args ...interface{}) {
	var lines = strings.Split(string(s.src), "\n")
	var b strings.Builder
	var line, column = s.getCurrPosition()

	var ch string
	if s.currRune == 0 {
		ch = "0"
	} else {
		ch = fmt.Sprintf("%q", s.currRune)
	}

	var prefix = fmt.Sprintf(layout, args...)
	b.WriteString(fmt.Sprintf("%s  %s\n", prefix, lines[line]))
	b.WriteString(fmt.Sprintf("%s  %s▲ [%d]=%s token=%s\n", prefix, strings.Repeat(" ", column), s.position, ch, s.currToken))
	fmt.Fprintf(os.Stderr, b.String())
	// fmt.Println("[debug] ", prefix, string(s.src), len(s.src), fmt.Sprintf("[%4d]=%c", s.position, s.currRune))
}

func (s *CmdScanner) getCurrPosition() (int, int) {
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

func (s *CmdScanner) NextToken() (CmdToken, error) {
	var t, err = s.nextToken()
	s.currToken = t
	if s.debug {
		s.PrintCursor("[debug]")
	}
	return t, err
}

func (s *CmdScanner) nextToken() (CmdToken, error) {
	// NOTE: We don't need to call readRune() in the cases where a s.read*() is called.
	//       because readRune() is already called inside those functions.
	switch s.currRune {
	case '\\': // escape character
		// TODO: consider entering literal $ with $$ and « with «« etc...
		// TODO: Not sure if this is the correct logic

		var rn = s.currRune
		s.readRune() // get the next one

		switch s.currRune {
		case '«', '»', '‹', '›', '$':
			var tok = CmdToken{WORD, fmt.Sprintf("%c", s.currRune)}
			s.readRune()
			return tok, nil
		default:
			var tok = CmdToken{WORD, fmt.Sprintf("%c", rn)}
			return tok, nil
		}
	case '\'':
		var tok = CmdToken{SINGLE_QUOTE, fmt.Sprintf("%c", s.currRune)}
		s.readRune()
		return tok, nil
	case '"':
		var tok = CmdToken{DOUBLE_QUOTE, fmt.Sprintf("%c", s.currRune)}
		s.readRune()
		return tok, nil
	case '$':
		var tok = s.readVariable()
		return tok, nil
	case '«':
		var tok = CmdToken{LDOUBLE_GUILLEMET, fmt.Sprintf("%c", s.currRune)}
		s.readRune()
		return tok, nil
	case '»':
		var tok = CmdToken{RDOUBLE_GUILLEMET, fmt.Sprintf("%c", s.currRune)}
		s.readRune()
		return tok, nil
	case '‹':
		var tok = CmdToken{LSINGLE_GUILLEMET, fmt.Sprintf("%c", s.currRune)}
		s.readRune()
		return tok, nil
	case '›':
		var tok = CmdToken{RSINGLE_GUILLEMET, fmt.Sprintf("%c", s.currRune)}
		s.readRune()
		return tok, nil
	case ' ', '\t', '\n', '\r':
		var tok = s.readWhitespace()
		return tok, nil
	case 0:
		var tok = CmdToken{EOF, ""}
		s.readRune()
		return tok, nil
	default:
		var tok = s.readWord()
		return tok, nil
	}
}
