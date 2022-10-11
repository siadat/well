package scanner

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

type CmdScanner struct {
	src          []rune
	currToken    CmdToken
	currRune     rune
	position     int
	readPosition int
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

func (s *CmdScanner) isIdentifier() bool {
	var ch = s.currRune
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
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
		'{', '}',
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
	position := s.position

	if want, curr := '%', s.currRune; curr != want {
		s.readRune() // skip
		return ""
	}
	s.readRune() // skip

	switch s.currRune {
	case '-', 'f', 's', 'q', 'Q': // ok
		s.readRune()
	default:
		s.readRune()
		return ""
	}

	return string(s.src[position:s.position])
}

func (s *CmdScanner) readIdentifier() string {
	position := s.position
	for s.isIdentifier() {
		s.readRune()
	}
	return string(s.src[position:s.position])
}

func (s *CmdScanner) readWhitespace() CmdToken {
	position := s.position
	for s.isWhitespace() {
		s.readRune()
	}
	return CmdToken{
		Typ: SPACE,
		Lit: string(s.src[position:s.position]),
	}
}

func (s *CmdScanner) readWord() CmdToken {
	position := s.position
	for s.isWord() {
		s.readRune()
	}
	return CmdToken{
		Typ: WORD,
		Lit: string(s.src[position:s.position]),
	}
}

func (s *CmdScanner) PrintCursor(layout string, args ...interface{}) {
	lines := strings.Split(string(s.src), "\n")
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
	fmt.Print(b.String())
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
	// s.PrintCursor("deferred")
	return t, err
}

func (s *CmdScanner) nextToken() (CmdToken, error) {
	var tok = CmdToken{
		Typ: ILLEGAL_TOKEN,
	}
	switch s.currRune {
	case '\\': // escape character
		if s.position >= len(s.src)-1 {
			return CmdToken{
				Typ: ILLEGAL_TOKEN,
				Lit: "\\",
			}, fmt.Errorf("backslash at the end of the source")
		}
		s.readRune() // get the next one
		tok = CmdToken{WORD, fmt.Sprintf("%c", s.currRune)}
		s.readRune()
		return tok, nil
	case '\'':
		tok = CmdToken{SINGLE_QUOTE, fmt.Sprintf("%c", s.currRune)}
		s.readRune()
		return tok, nil
	case '"':
		tok = CmdToken{DOUBLE_QUOTE, fmt.Sprintf("%c", s.currRune)}
		s.readRune()
		return tok, nil
	case '$':
		// NOTE: we return here, because we don't want to readRune()
		//       because readRune is already called inside readVariable
		tok = s.readVariable()
		return tok, nil
	case '}':
	case '«':
		tok = CmdToken{LDOUBLE_GUILLEMET, fmt.Sprintf("%c", s.currRune)}
		s.readRune()
		return tok, nil
	case '»':
		tok = CmdToken{RDOUBLE_GUILLEMET, fmt.Sprintf("%c", s.currRune)}
		s.readRune()
		return tok, nil
	case '‹':
		tok = CmdToken{LSINGLE_GUILLEMET, fmt.Sprintf("%c", s.currRune)}
		s.readRune()
		return tok, nil
	case '›':
		tok = CmdToken{RSINGLE_GUILLEMET, fmt.Sprintf("%c", s.currRune)}
		s.readRune()
		return tok, nil
	case ' ', '\t', '\n', '\r':
		// NOTE: we return here, because we don't want to readRune()
		//       because readRune is already called inside readWord
		tok = s.readWhitespace()
		return tok, nil
	case 0:
		tok = CmdToken{EOF, ""}
	default:
		// NOTE: we return here, because we don't want to readRune()
		//       because readRune is already called inside readWord
		tok = s.readWord()
		return tok, nil
	}
	s.readRune()

	return tok, nil
}
